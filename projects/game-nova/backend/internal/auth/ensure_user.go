package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// StarterAssigner назначает стартовую планету новому пользователю.
// Описывает только сигнатуру planet.Starter — чтобы избежать import cycle.
type StarterAssigner interface {
	Assign(ctx context.Context, userID string) (string, error)
}

// AutomsgSender отправляет авто-сообщение пользователю (welcome, starter-guide).
// Описывает только сигнатуру automsg.Service.Send.
type AutomsgSender interface {
	Send(ctx context.Context, tx pgx.Tx, userID, key string, vars map[string]string) error
}

// EnsureUserConfig — параметры lazy-create middleware.
type EnsureUserConfig struct {
	Pool       *pgxpool.Pool
	Starter    StarterAssigner
	Automsg    AutomsgSender
	UniverseID string // план 36 Critical-5: регистрируется в universe_memberships
	// AuthServiceURL — base URL auth-service для внутренних вызовов.
	// Пустая строка — пропускаем регистрацию membership (для тестов).
	AuthServiceURL string
}

// EnsureUserMiddleware гарантирует, что юзер из RSA-claims существует
// в game-nova users table. План 36 Ф.12: lazy-create в middleware с ON CONFLICT
// (защита от гонки одновременных запросов).
//
// Тяжёлая инициализация (стартовая планета, welcome-сообщение, регистрация
// universe_membership в auth-service) делается АСИНХРОННО — fire-and-forget,
// чтобы не задерживать первый ответ.
//
// Должен стоять ПОСЛЕ RSAMiddleware.
func EnsureUserMiddleware(cfg EnsureUserConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := RSAClaims(r.Context())
			if !ok {
				next.ServeHTTP(w, r)
				return
			}
			ctx := r.Context()
			// План 36 Nice-10: email больше не в JWT-claims (PII).
			// В game-db email NULLABLE — для отображения email берётся через
			// /auth/me в auth-service по требованию (admin-views и т.п.).
			tag, err := cfg.Pool.Exec(ctx, `
				INSERT INTO users (id, username, email, password_hash)
				VALUES ($1, $2, NULL, NULL)
				ON CONFLICT (id) DO NOTHING
			`, claims.Subject, claims.Username)
			if err != nil {
				// Не валим запрос: middleware невидим для эндпоинта.
				slog.WarnContext(ctx, "ensure-user insert failed",
					slog.String("user_id", claims.Subject),
					slog.String("err", err.Error()))
				next.ServeHTTP(w, r)
				return
			}
			if tag.RowsAffected() == 1 {
				// Юзер только что создан — асинхронно бутстрапим.
				go bootstrapNewUser(claims.Subject, claims.Username, cfg)
			} else {
				// План 36 Critical-8: юзер уже есть, но bootstrap мог упасть
				// раньше. Если cur_planet_id IS NULL — пробуем повторно
				// назначить стартовую планету (не блокирующе).
				go retryBootstrapIfNeeded(claims.Subject, claims.Username, cfg)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// retryBootstrapIfNeeded дозапускает bootstrap, если он не завершился ранее.
// Сейчас проверяет только наличие стартовой планеты. План 36 Critical-8.
//
// Запрос на cur_planet_id выполняется на каждом запросе — это лишняя
// SELECT-нагрузка, но дешёвая (PK lookup). Можно оптимизировать через
// in-memory cache «уже-проверенных» юзеров, если профайлинг покажет hot path.
func retryBootstrapIfNeeded(userID, username string, cfg EnsureUserConfig) {
	if cfg.Pool == nil || cfg.Starter == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var hasPlanet bool
	err := cfg.Pool.QueryRow(ctx,
		`SELECT cur_planet_id IS NOT NULL FROM users WHERE id = $1`, userID,
	).Scan(&hasPlanet)
	if err != nil || hasPlanet {
		return
	}
	slog.InfoContext(ctx, "retrying starter planet assign",
		slog.String("user_id", userID))
	if _, err := cfg.Starter.Assign(ctx, userID); err != nil {
		slog.WarnContext(ctx, "retry starter planet assign failed",
			slog.String("user_id", userID),
			slog.String("err", err.Error()))
		return
	}
	// Welcome повторно не шлём — он уже мог быть отправлен (или не критичен).
}

func bootstrapNewUser(userID, username string, cfg EnsureUserConfig) {
	ctx := context.Background()
	if cfg.Starter != nil {
		if _, err := cfg.Starter.Assign(ctx, userID); err != nil {
			slog.WarnContext(ctx, "starter planet assign failed",
				slog.String("user_id", userID),
				slog.String("err", err.Error()))
			// Продолжаем bootstrap — welcome и membership всё ещё имеет смысл.
		}
	}
	if cfg.Automsg != nil {
		_ = cfg.Automsg.Send(ctx, nil, userID, "welcome", map[string]string{"username": username})
		_ = cfg.Automsg.Send(ctx, nil, userID, "starterGuide", nil)
	}
	if cfg.AuthServiceURL != "" && cfg.UniverseID != "" {
		registerUniverseMembership(ctx, cfg.AuthServiceURL, userID, cfg.UniverseID)
	}
}

// registerUniverseMembership шлёт POST /auth/universes/register в auth-service —
// чтобы при следующей выдаче JWT этот universe попал в active_universes claim.
// План 36 Critical-5.
func registerUniverseMembership(ctx context.Context, authServiceURL, userID, universeID string) {
	body, _ := json.Marshal(map[string]string{
		"user_id":     userID,
		"universe_id": universeID,
	})
	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost,
		authServiceURL+"/auth/universes/register", bytes.NewReader(body))
	if err != nil {
		slog.WarnContext(ctx, "build universes/register request failed", slog.String("err", err.Error()))
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.WarnContext(ctx, "universes/register request failed",
			slog.String("user_id", userID),
			slog.String("universe_id", universeID),
			slog.String("err", err.Error()))
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		slog.WarnContext(ctx, "universes/register non-2xx",
			slog.String("user_id", userID),
			slog.String("universe_id", universeID),
			slog.Int("status", resp.StatusCode))
	}
}
