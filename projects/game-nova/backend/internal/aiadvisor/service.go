// Package aiadvisor реализует AI-советника для игроков.
//
// Два режима: Claude API (за кредиты) или Ollama (бесплатно).
// Rate limit: MaxPerDay вопросов в сутки на игрока.
// Кредиты списываются ПОСЛЕ получения ответа (ошибка API → не списывается).
package aiadvisor

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/internal/config"
	"oxsar/game-nova/internal/repo"
	"oxsar/game-nova/pkg/ids"
)

var (
	ErrNotEnoughCredit  = errors.New("aiadvisor: not enough credits")
	ErrRateLimitReached = errors.New("aiadvisor: daily limit reached")
	ErrUnknownModel     = errors.New("aiadvisor: unknown model")
	ErrNoBackend        = errors.New("aiadvisor: no LLM backend configured")
)

type Service struct {
	db  repo.Exec
	llm LLMClient
	cfg config.AIAdvisorConfig
}

// NewService создаёт сервис AI-советника.
// llm выбирается автоматически: если cfg.OllamaURL != "" — Ollama, иначе Anthropic.
// Если ни один не настроен — сервис возвращает ErrNoBackend при каждом Ask.
func NewService(db repo.Exec, cfg config.AIAdvisorConfig) *Service {
	var llm LLMClient
	if cfg.OllamaURL != "" {
		llm = NewOllamaClient(cfg.OllamaURL)
	} else if cfg.APIKey != "" {
		llm = NewAnthropicClient(cfg.APIKey, cfg.ProxyURL)
	}
	return &Service{db: db, llm: llm, cfg: cfg}
}

// AskResult — ответ сервиса.
type AskResult struct {
	Answer     string  `json:"answer"`
	CreditsUsed float64 `json:"credits_used"`
}

// EstimateResult — стоимость без реального запроса.
type EstimateResult struct {
	Model       string  `json:"model"`
	CreditCost  int     `json:"credit_cost"`
	DailyUsed   int     `json:"daily_used"`
	DailyLimit  int     `json:"daily_limit"`
	HasBalance  bool    `json:"has_balance"`
}

// Ask задаёт вопрос AI-советнику.
// Для Ollama model игнорируется (используется cfg.OllamaModel).
// Кредиты списываются только при успешном ответе.
func (s *Service) Ask(ctx context.Context, userID, model, question string) (AskResult, error) {
	if s.llm == nil {
		return AskResult{}, ErrNoBackend
	}

	// Ollama — бесплатно, модель фиксирована.
	isOllama := s.cfg.OllamaURL != ""
	creditCost := 0
	actualModel := model
	if isOllama {
		actualModel = s.cfg.OllamaModel
	} else {
		cost, ok := KnownModels[model]
		if !ok {
			return AskResult{}, ErrUnknownModel
		}
		creditCost = cost
	}

	// Проверка rate limit и баланса в транзакции.
	var balance float64
	var dailyCount int
	err := s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `
			SELECT COALESCE(credit, 0) FROM users WHERE id = $1
		`, userID).Scan(&balance); err != nil {
			return fmt.Errorf("read balance: %w", err)
		}
		if err := tx.QueryRow(ctx, `
			SELECT COUNT(*) FROM ai_advisor_log
			WHERE user_id = $1 AND created_at > now() - interval '24 hours'
		`, userID).Scan(&dailyCount); err != nil {
			return fmt.Errorf("read daily count: %w", err)
		}
		if dailyCount >= s.cfg.MaxPerDay {
			return ErrRateLimitReached
		}
		if !isOllama && balance < float64(creditCost) {
			return ErrNotEnoughCredit
		}
		return nil
	})
	if err != nil {
		return AskResult{}, err
	}

	// Строим системный контекст (статическая часть — знания об игре).
	systemPrompt := buildStaticGameKnowledge()
	playerCtx, _ := s.buildPlayerContext(ctx, userID)
	if playerCtx != "" {
		systemPrompt += "\n\n" + playerCtx
	}

	// Вызов LLM.
	answer, err := s.llm.Ask(ctx, actualModel, systemPrompt, question, s.cfg.MaxTokens)
	if err != nil {
		return AskResult{}, fmt.Errorf("aiadvisor: llm: %w", err)
	}

	// Списываем кредиты и логируем (оба действия в одной транзакции).
	err = s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		if !isOllama && creditCost > 0 {
			if _, err := tx.Exec(ctx, `
				UPDATE users SET credit = credit - $1 WHERE id = $2
			`, creditCost, userID); err != nil {
				return fmt.Errorf("deduct credits: %w", err)
			}
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO ai_advisor_log (id, user_id, model, question, answer, credits, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, ids.New(), userID, actualModel, question, answer, creditCost, time.Now().UTC()); err != nil {
			return fmt.Errorf("insert log: %w", err)
		}
		return nil
	})
	if err != nil {
		return AskResult{Answer: answer, CreditsUsed: 0}, nil // ответ есть, но списание не прошло — вернём ответ
	}

	return AskResult{Answer: answer, CreditsUsed: float64(creditCost)}, nil
}

// Estimate возвращает стоимость запроса без его выполнения.
func (s *Service) Estimate(ctx context.Context, userID, model string) (EstimateResult, error) {
	isOllama := s.cfg.OllamaURL != ""
	creditCost := 0
	if !isOllama {
		cost, ok := KnownModels[model]
		if !ok {
			return EstimateResult{}, ErrUnknownModel
		}
		creditCost = cost
	}

	var balance float64
	var dailyCount int
	if err := s.db.Pool().QueryRow(ctx, `SELECT COALESCE(credit,0) FROM users WHERE id=$1`, userID).Scan(&balance); err != nil {
		return EstimateResult{}, err
	}
	if err := s.db.Pool().QueryRow(ctx, `
		SELECT COUNT(*) FROM ai_advisor_log
		WHERE user_id=$1 AND created_at > now() - interval '24 hours'
	`, userID).Scan(&dailyCount); err != nil {
		return EstimateResult{}, err
	}

	return EstimateResult{
		Model:      model,
		CreditCost: creditCost,
		DailyUsed:  dailyCount,
		DailyLimit: s.cfg.MaxPerDay,
		HasBalance: balance >= float64(creditCost),
	}, nil
}

// buildPlayerContext собирает контекст игрока (планеты, ресурсы).
// Не критично — при ошибке возвращает пустую строку.
func (s *Service) buildPlayerContext(ctx context.Context, userID string) (string, error) {
	rows, err := s.db.Pool().Query(ctx, `
		SELECT name, galaxy, system, position, metal, silicon, hydrogen
		FROM planets WHERE user_id=$1 ORDER BY created_at LIMIT 5
	`, userID)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	result := "Планеты игрока:\n"
	for rows.Next() {
		var name string
		var g, sys, pos int
		var metal, silicon, hydrogen float64
		if err := rows.Scan(&name, &g, &sys, &pos, &metal, &silicon, &hydrogen); err != nil {
			continue
		}
		result += fmt.Sprintf("- %s [%d:%d:%d] M=%.0f Si=%.0f H=%.0f\n",
			name, g, sys, pos, metal, silicon, hydrogen)
	}
	return result, rows.Err()
}

// buildStaticGameKnowledge возвращает статический системный промпт об игре.
func buildStaticGameKnowledge() string {
	return `Ты — AI-советник в браузерной космической стратегии Oxsar Nova (порт OGame/Oxsar2).

Правила игры:
- Добывай ресурсы (металл, кремний, водород) через шахты и лаборатории.
- Строй корабли и оборону для защиты и атаки.
- Исследования улучшают производство и боевые характеристики.
- Боевые технологии: Weapons (+10%/уровень атаки), Shield (+10%/уровень щита), Armor (+10%/уровень брони).
- Флот: лёгкий истребитель → крейсер → линкор → звёздный разрушитель и другие.
- Экспедиции дают ресурсы, артефакты, кредиты.
- Артефакты активируются и дают бонусы к производству или бою.
- Профессии: miner (производство), attacker (атака), defender (оборона), tank (броня).
- Кредиты: получают за победы, достижения, ежедневный бонус. Используются для смены профессии, офицеров.

Отвечай кратко и по делу на вопросы игрока. Используй его контекст (планеты) если он предоставлен.`
}
