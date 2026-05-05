// Файл реализует AutopilotService — rule-based автопилот (план 06.1).
//
// Жизненный цикл одного запроса:
//
//  1. Игрок жмёт «Запустить Автопилот» в виджете советника.
//  2. Фронт → POST /api/ai-advisor/autopilot/advise {strategy}.
//  3. Handler → AutopilotService.Enqueue:
//     - проверяет дневной лимит и баланс кредитов;
//     - в одной транзакции списывает AutopilotCostCredits, создаёт
//       запись в ai_advisor_log (status='pending') и event
//       KindAutopilotAdvise (fire_at = now);
//     - возвращает {job_id} = id записи лога.
//  4. Воркер забирает event'у в processOne(), вызывает
//     AutopilotWorkerHandler.Handle → AutopilotService.Compute:
//     - buildSnapshot (чтение из БД через сервисы);
//     - scoreCandidates (pure функция);
//     - UPDATE ai_advisor_log SET action_params=JSONB, status='ready'.
//     - Compute идемпотентен: повторный вызов с тем же event.id
//       проверяет current status — если уже ready/executed/expired,
//       возвращает nil без перерасчёта.
//  5. Фронт периодически (или по WS-пушу в Ф.6) делает
//     GET /result/{job_id} → AutopilotService.Result.
//  6. Игрок выбирает карточку → POST /execute {rec_id}.
//  7. AutopilotService.Execute:
//     - валидирует, что лог status='ready' и rec_id есть в action_params;
//     - вызывает executeRecommendation (building.Enqueue / ...);
//     - UPDATE ai_advisor_log SET status='executed', event_id=...
//       (event_id здесь хранит id queue-item / fleet, для аудита);
//     - возвращает {event_id}.
package aiadvisor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/internal/automsg"
	"oxsar/game-nova/internal/building"
	"oxsar/game-nova/internal/config"
	"oxsar/game-nova/internal/event"
	"oxsar/game-nova/internal/fleet"
	"oxsar/game-nova/internal/planet"
	"oxsar/game-nova/internal/profession"
	"oxsar/game-nova/internal/repo"
	"oxsar/game-nova/internal/research"
	"oxsar/game-nova/internal/score"
	"oxsar/game-nova/pkg/ids"
)

var (
	// ErrInvalidStrategy — неизвестное значение стратегии.
	ErrInvalidStrategy = errors.New("autopilot: invalid strategy")
	// ErrJobNotFound — записи с таким job_id нет (или принадлежит другому юзеру).
	ErrJobNotFound = errors.New("autopilot: job not found")
	// ErrJobNotReady — топ-3 ещё не рассчитан, повторите запрос.
	ErrJobNotReady = errors.New("autopilot: job not ready")
	// ErrRecommendationNotFound — указанный rec_id не из этого job'а.
	ErrRecommendationNotFound = errors.New("autopilot: recommendation not found")
)

// AutopilotService — основной сервис автопилота.
//
// Зависимости разделены по ролям:
//   - DB / Catalog — для прямых SQL и доступа к каталогу юнитов;
//   - PlanetSvc / ScoreSvc — чтение состояния (snapshot);
//   - BuildingSvc / ResearchSvc — выполнение рекомендаций и time/cost
//     калькуляции при scoring;
//   - PointsK — коэффициенты очков, нужны scoring-у для оценки выгоды.
//
// Hub (chat.Hub) опционален — если nil, WS-пуш не делается, фронт
// получит результат через polling. Это допустимо в Ф.1.
// GameTuning — тюнинг-параметры экономики, которые нужны scoring-у
// (помимо коэффициентов очков). Извлечены из config.GameConfig чтобы
// AutopilotService не зависел от всей game-конфигурации.
type GameTuning struct {
	PointsK          config.PointsCoefficients
	GameSpeed        float64
	ResearchSpeed    float64
	ProtectionPeriod int // секунды защиты новичка от атак
}

type AutopilotService struct {
	db             repo.Exec
	cfg            config.AIAdvisorConfig
	tuning         GameTuning
	catalog        *config.Catalog
	planetSvc      *planet.Service
	scoreSvc       *score.Service
	buildSvc       *building.Service
	researchSvc    *research.Service
	fleetSvc       *fleet.TransportService
	professionSvc  *profession.Service
	automsg        *automsg.Service
}

// WithAutoMsg подключает сервис системных сообщений. Если задан,
// Compute дополнительно сохраняет рекомендации в inbox пользователя
// (folder=20 «Системные») — полезно когда игрок offline.
func (s *AutopilotService) WithAutoMsg(am *automsg.Service) *AutopilotService {
	s.automsg = am
	return s
}

// NewAutopilotService — конструктор. Любая зависимость может быть nil
// (тесты передают только нужное), но для прод-окружения server/main.go
// и worker/main.go обязаны передать все.
func NewAutopilotService(
	db repo.Exec,
	cfg config.AIAdvisorConfig,
	tuning GameTuning,
	catalog *config.Catalog,
	planetSvc *planet.Service,
	scoreSvc *score.Service,
	buildSvc *building.Service,
	researchSvc *research.Service,
	fleetSvc *fleet.TransportService,
	professionSvc *profession.Service,
) *AutopilotService {
	return &AutopilotService{
		db:             db,
		cfg:            cfg,
		tuning:         tuning,
		catalog:        catalog,
		planetSvc:      planetSvc,
		scoreSvc:       scoreSvc,
		buildSvc:       buildSvc,
		researchSvc:    researchSvc,
		fleetSvc:       fleetSvc,
		professionSvc:  professionSvc,
	}
}

// Enqueue ставит задание автопилота в очередь воркера.
//
// Списание кредитов, создание лога и event'а — в одной транзакции;
// при ошибке состояние не меняется.
//
// Дневной лимит — общий с LLM-советником (cfg.MaxPerDay): и LLM-вопросы,
// и autopilot-запросы считаются по таблице ai_advisor_log за последние
// 24 часа. Это предотвращает ситуацию, когда игрок обходит лимит LLM
// через автопилот (или наоборот).
func (s *AutopilotService) Enqueue(ctx context.Context, userID string, strategy Strategy) (EnqueueResult, error) {
	if !strategy.Validate() {
		return EnqueueResult{}, ErrInvalidStrategy
	}

	cost := s.cfg.AutopilotCostCredits
	logID := ids.New()
	now := time.Now().UTC()

	err := s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Проверка баланса и rate limit.
		var balance float64
		var dailyCount int
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
		if balance < cost {
			return ErrNotEnoughCredit
		}

		// Списание кредитов.
		if cost > 0 {
			if _, err := tx.Exec(ctx, `
				UPDATE users SET credit = credit - $1 WHERE id = $2
			`, cost, userID); err != nil {
				return fmt.Errorf("deduct credits: %w", err)
			}
		}

		// Запись в ai_advisor_log (status='pending').
		// model='autopilot' для совместимости с уже существующей колонкой
		// (она NOT NULL); реальная стратегия — в strategy.
		if _, err := tx.Exec(ctx, `
			INSERT INTO ai_advisor_log (
				id, user_id, model, question, answer,
				tokens_used, credits, created_at,
				source, strategy, status
			) VALUES ($1, $2, 'autopilot', '', '', 0, $3, $4, $5, $6, $7)
		`, logID, userID, cost, now, AutopilotSource, string(strategy), AutopilotStatusPending); err != nil {
			return fmt.Errorf("insert log: %w", err)
		}

		// Создание event'а KindAutopilotAdvise. Воркер заберёт его сразу
		// (fire_at = now) и вызовет AutopilotService.Compute.
		payload := AdviseInput{
			UserID:   userID,
			Strategy: strategy,
			LogID:    logID,
		}
		uid := userID
		if _, err := event.Insert(ctx, tx, event.InsertOpts{
			UserID:  &uid,
			Kind:    event.KindAutopilotAdvise,
			FireAt:  now,
			Payload: payload,
		}); err != nil {
			return fmt.Errorf("insert event: %w", err)
		}
		return nil
	})
	if err != nil {
		return EnqueueResult{}, err
	}

	return EnqueueResult{JobID: logID}, nil
}

// Compute — обработчик event'а KindAutopilotAdvise (вызывается воркером
// внутри транзакции).
//
// Идемпотентность: если лог уже не в pending — return nil без работы.
// Это безопасно при retry (worker может повторно дёрнуть тот же event.id).
func (s *AutopilotService) Compute(ctx context.Context, tx pgx.Tx, e event.Event) error {
	var input AdviseInput
	if err := json.Unmarshal(e.Payload, &input); err != nil {
		return fmt.Errorf("autopilot: compute: parse payload: %w", err)
	}
	if input.LogID == "" || input.UserID == "" {
		return fmt.Errorf("autopilot: compute: empty log_id or user_id in payload")
	}

	// Идемпотентность: читаем текущий status под FOR UPDATE,
	// чтобы исключить race между retry-ями.
	var currentStatus string
	if err := tx.QueryRow(ctx, `
		SELECT status FROM ai_advisor_log
		WHERE id = $1 AND user_id = $2 AND source = $3
		FOR UPDATE
	`, input.LogID, input.UserID, AutopilotSource).Scan(&currentStatus); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Лог удалён (например админом или TTL'ом — пока не реализован).
			return nil
		}
		return fmt.Errorf("autopilot: compute: read status: %w", err)
	}
	if currentStatus != AutopilotStatusPending {
		return nil // уже обработан
	}

	// Чтение snapshot и scoring выполняются вне транзакции воркера,
	// потому что это длительные операции (дёргают много сервисов).
	// Транзакция держится только на UPDATE результата.
	snap, err := buildSnapshot(ctx, SnapshotDeps{
		DB:               s.db,
		PlanetSvc:        s.planetSvc,
		ScoreSvc:         s.scoreSvc,
		BuildingSvc:      s.buildSvc,
		ResearchSvc:      s.researchSvc,
		ProtectionPeriod: s.tuning.ProtectionPeriod,
	}, input.UserID)
	if err != nil {
		return fmt.Errorf("autopilot: compute: snapshot: %w", err)
	}

	// Время постройки следующего уровня для каждого здания — учитывает
	// robotic_factory / nano_factory на планете и game speed. Без этой
	// карты scoring не может ранжировать кандидатов по delta/время.
	buildSecondsByPlanet := make(map[string]map[int]int, len(snap.Planets))
	if s.buildSvc != nil {
		for _, ps := range snap.Planets {
			seconds, err := s.buildSvc.BuildSecondsMap(ctx, ps.ID, ps.Buildings)
			if err != nil {
				return fmt.Errorf("autopilot: compute: build seconds for %s: %w", ps.ID, err)
			}
			buildSecondsByPlanet[ps.ID] = seconds
		}
	}

	recs := scoreCandidates(snap, input.Strategy, scoringInputs{
		Catalog:              s.catalog,
		PointsK:              s.tuning.PointsK,
		BuildSecondsByPlanet: buildSecondsByPlanet,
		GameSpeed:            s.tuning.GameSpeed,
		ResearchSpeed:        s.tuning.ResearchSpeed,
	})
	for i := range recs {
		if recs[i].ID == "" {
			recs[i].ID = ids.New()
		}
	}

	result := AutopilotResult{
		JobID:    input.LogID,
		Status:   AutopilotStatusReady,
		Strategy: input.Strategy,
		Recs:     recs,
	}
	payload, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("autopilot: compute: marshal recs: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		UPDATE ai_advisor_log
		SET action_params = $1, status = $2
		WHERE id = $3 AND user_id = $4
	`, payload, AutopilotStatusReady, input.LogID, input.UserID); err != nil {
		return fmt.Errorf("autopilot: compute: update log: %w", err)
	}

	// Inbox-message: дублируем рекомендации в почту пользователя,
	// чтобы игрок-офлайн увидел их при следующем входе. Кнопка «Выполнить»
	// в письме делается фронтом по deep-link на advisor с pre-loaded job.
	// Ошибка inbox не критична — ai_advisor_log уже обновлён.
	if s.automsg != nil && len(recs) > 0 {
		title := "Автопилот: новые рекомендации"
		body := buildInboxBody(input.Strategy, recs, input.LogID)
		// folder=20 — системные сообщения (legacy consts.php
		// MESSAGE_FOLDER_FROM_AI = пока не выделен, используем 20).
		_ = s.automsg.SendDirect(ctx, tx, input.UserID, 20, title, body)
	}

	// WS-пуш — Ф.6. До тех пор фронт получает результат через polling.

	return nil
}

// buildInboxBody — текст inbox-сообщения с топ-3 рекомендациями.
// Plain text, без HTML — формат legacy messages.body.
func buildInboxBody(strategy Strategy, recs []Recommendation, jobID string) string {
	body := "Стратегия: " + string(strategy) + "\n\n"
	for i, rec := range recs {
		body += fmt.Sprintf("%d. %s\n   %s\n\n", i+1, rec.Description, rec.Benefit)
	}
	body += "Откройте экран «Советник», чтобы выполнить выбранное действие.\n"
	body += "Job: " + jobID
	return body
}

// HistoryItem — элемент истории запросов автопилота.
type HistoryItem struct {
	JobID     string  `json:"job_id"`
	Strategy  string  `json:"strategy"`
	Status    string  `json:"status"`
	CreatedAt string  `json:"created_at"` // ISO-8601
	Credits   float64 `json:"credits"`
	// Если status='executed' — описание выполненной рекомендации.
	ExecutedDescription string `json:"executed_description,omitempty"`
}

// History возвращает последние N запросов автопилота для пользователя.
// N жёстко 20 — этого достаточно для UI-вкладки «История».
// Сортировка: created_at DESC.
func (s *AutopilotService) History(ctx context.Context, userID string) ([]HistoryItem, error) {
	rows, err := s.db.Pool().Query(ctx, `
		SELECT id, COALESCE(strategy, ''), status, created_at, credits, action_params
		FROM ai_advisor_log
		WHERE user_id = $1 AND source = $2
		ORDER BY created_at DESC
		LIMIT 20
	`, userID, AutopilotSource)
	if err != nil {
		return nil, fmt.Errorf("autopilot: history: %w", err)
	}
	defer rows.Close()

	var out []HistoryItem
	for rows.Next() {
		var (
			h         HistoryItem
			createdAt time.Time
			ap        []byte
		)
		if err := rows.Scan(&h.JobID, &h.Strategy, &h.Status, &createdAt, &h.Credits, &ap); err != nil {
			return nil, err
		}
		h.CreatedAt = createdAt.Format(time.RFC3339)
		// Если есть action_params — попытаемся достать описание выполненной рекомендации
		// (для status='executed' она помечена через UPDATE в Execute? — нет, мы только
		// меняем status. Извлекаем первую recs[0].description как proxy.)
		if len(ap) > 0 && h.Status == AutopilotStatusExecuted {
			var stored AutopilotResult
			if err := json.Unmarshal(ap, &stored); err == nil && len(stored.Recs) > 0 {
				h.ExecutedDescription = stored.Recs[0].Description
			}
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

// Result — polling-эндпоинт. Возвращает текущее состояние job'а.
// Если status='pending', recs пуст; фронт должен повторить запрос
// через 1–2 секунды.
func (s *AutopilotService) Result(ctx context.Context, userID, jobID string) (AutopilotResult, error) {
	var status string
	var actionParams []byte
	var strategy *string
	err := s.db.Pool().QueryRow(ctx, `
		SELECT status, action_params, strategy
		FROM ai_advisor_log
		WHERE id = $1 AND user_id = $2 AND source = $3
	`, jobID, userID, AutopilotSource).Scan(&status, &actionParams, &strategy)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AutopilotResult{}, ErrJobNotFound
		}
		return AutopilotResult{}, fmt.Errorf("autopilot: result: %w", err)
	}

	res := AutopilotResult{JobID: jobID, Status: status}
	if strategy != nil {
		res.Strategy = Strategy(*strategy)
	}
	if len(actionParams) > 0 {
		// action_params хранит уже сериализованный AutopilotResult (с recs).
		// Достаём recs из него; status/strategy/job_id берём из колонок,
		// чтобы не гонять между ними.
		var stored AutopilotResult
		if err := json.Unmarshal(actionParams, &stored); err == nil {
			res.Recs = stored.Recs
		}
	}
	return res, nil
}

// Execute подтверждает и выполняет одну из топ-3 рекомендаций.
//
// Валидации:
//   - запись принадлежит user'у и status='ready';
//   - rec_id присутствует в action_params.recs.
//
// Если выполнение успешно — UPDATE status='executed', event_id=...
// (event_id хранится для аудита; UI использует его, чтобы открыть
// детали созданного действия).
func (s *AutopilotService) Execute(ctx context.Context, userID, jobID, recID string) (ExecuteResult, error) {
	// Сначала вне транзакции читаем текущее состояние и находим rec.
	res, err := s.Result(ctx, userID, jobID)
	if err != nil {
		return ExecuteResult{}, err
	}
	if res.Status != AutopilotStatusReady {
		return ExecuteResult{}, ErrJobNotReady
	}
	var rec *Recommendation
	for i := range res.Recs {
		if res.Recs[i].ID == recID {
			rec = &res.Recs[i]
			break
		}
	}
	if rec == nil {
		return ExecuteResult{}, ErrRecommendationNotFound
	}

	// executeRecommendation сам открывает свои транзакции (через
	// building.Enqueue / research.Enqueue). Они идемпотентны на уровне
	// своих сервисов.
	eventID, err := executeRecommendation(ctx, executorDeps{
		Building:   s.buildSvc,
		Research:   s.researchSvc,
		Fleet:      s.fleetSvc,
		Profession: s.professionSvc,
	}, userID, *rec)
	if err != nil {
		return ExecuteResult{}, err
	}

	// Обновление статуса лога — в отдельной транзакции (к этому моменту
	// игровое событие уже создано). При ошибке UPDATE — игрок увидит
	// статус ready, но при повторном Execute ErrJobNotReady не сработает,
	// так как rec уже выполнен; это вернёт ошибку из Enqueue (например
	// ErrQueueBusy). Это допустимо — система не теряет действие, лишь
	// расходится статус лога с реальностью.
	if _, err := s.db.Pool().Exec(ctx, `
		UPDATE ai_advisor_log
		SET status = $1
		WHERE id = $2 AND user_id = $3
	`, AutopilotStatusExecuted, jobID, userID); err != nil {
		// Лог не критичен для доменной операции — игнорируем (но возвращаем eventID).
		_ = err
	}

	return ExecuteResult{EventID: eventID}, nil
}
