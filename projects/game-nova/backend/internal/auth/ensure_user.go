package auth

import (
	"context"
	"log/slog"
	"net/http"

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

// EnsureUserMiddleware гарантирует, что юзер из RSA-claims существует
// в game-nova users table. План 36 Ф.12: lazy-create в middleware с ON CONFLICT
// (защита от гонки одновременных запросов).
//
// Тяжёлая инициализация (стартовая планета, welcome-сообщение) делается
// АСИНХРОННО — fire-and-forget, чтобы не задерживать первый ответ.
// Если планета не назначилась, будет повторная попытка на следующем запросе
// (пока planets-таблица пуста для этого юзера).
//
// Должен стоять ПОСЛЕ RSAMiddleware. Без RSA-claims в context — пропускает
// запрос (legacy HS256 не зеркалит, юзер уже в БД).
func EnsureUserMiddleware(pool *pgxpool.Pool, starter StarterAssigner, automsg AutomsgSender) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := RSAClaims(r.Context())
			if !ok {
				next.ServeHTTP(w, r)
				return
			}
			ctx := r.Context()
			tag, err := pool.Exec(ctx, `
				INSERT INTO users (id, username, email, password_hash)
				VALUES ($1, $2, $3, NULL)
				ON CONFLICT (id) DO NOTHING
			`, claims.Subject, claims.Username, claims.Email)
			if err != nil {
				// Не валим запрос: middleware невидим для эндпоинта.
				// Если /api/me потом упадёт — клиент увидит 500, мы увидим лог.
				slog.WarnContext(ctx, "ensure-user insert failed",
					slog.String("user_id", claims.Subject),
					slog.String("err", err.Error()))
				next.ServeHTTP(w, r)
				return
			}
			if tag.RowsAffected() == 1 {
				// Юзер только что создан — назначить стартовую планету и
				// отправить welcome. Асинхронно, чтобы не задерживать ответ.
				go bootstrapNewUser(claims.Subject, claims.Username, starter, automsg)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func bootstrapNewUser(userID, username string, starter StarterAssigner, automsg AutomsgSender) {
	ctx := context.Background()
	if starter != nil {
		if _, err := starter.Assign(ctx, userID); err != nil {
			slog.WarnContext(ctx, "starter planet assign failed",
				slog.String("user_id", userID),
				slog.String("err", err.Error()))
			return
		}
	}
	if automsg != nil {
		_ = automsg.Send(ctx, nil, userID, "welcome", map[string]string{"username": username})
		_ = automsg.Send(ctx, nil, userID, "starterGuide", nil)
	}
}
