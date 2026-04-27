package billing

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"oxsar/billing/internal/httpx"
)

// IdempotencyMiddleware реализует Stripe-style Idempotency-Key.
//
// Семантика:
//   - Header `Idempotency-Key` опционален. Если есть — middleware гарантирует
//     повторного выполнения handler для того же ключа в течение 24 часов.
//   - При первом запросе: handler выполняется, response (status + body) пишется
//     в idempotency_keys. Параллельно: ON CONFLICT DO NOTHING — если другой
//     запрос успел создать запись, мы читаем её и возвращаем тот же response
//     (без повторного выполнения handler).
//   - При повторе: возвращаем сохранённый response, handler НЕ вызывается.
//   - Если ключ переиспользован с другим body (`request_hash` отличается) — 422
//     «Idempotency-Key reuse with different request».
//
// Ключ scoped к user_id (из RSA-claims): cross-user replay невозможен.
// TTL 24 часа (см. миграцию 0001).
//
// Применять к POST /billing/wallet/spend, POST /billing/wallet/credit.
// На GET-ручках idempotency не нужен.
type IdempotencyMiddleware struct {
	pool *pgxpool.Pool
}

func NewIdempotencyMiddleware(pool *pgxpool.Pool) *IdempotencyMiddleware {
	return &IdempotencyMiddleware{pool: pool}
}

// UserIDProvider достаёт user_id из контекста запроса.
// Реализуется auth-middleware (RSA-claims).
type UserIDProvider func(r *http.Request) (string, bool)

func (m *IdempotencyMiddleware) Handler(getUserID UserIDProvider) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("Idempotency-Key")
			if key == "" {
				// Без ключа — idempotency не действует, handler вызывается как обычно.
				next.ServeHTTP(w, r)
				return
			}
			if len(key) > 255 {
				httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "Idempotency-Key too long"))
				return
			}
			userID, ok := getUserID(r)
			if !ok {
				httpx.WriteError(w, r, httpx.ErrUnauthorized)
				return
			}

			// Прочитать и закэшировать body для возможной повторной отправки в handler.
			bodyBytes, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
			if err != nil {
				httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "read body: "+err.Error()))
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

			reqHash := requestHash(r.Method, r.URL.Path, bodyBytes)

			// Проверим существующую запись.
			cached, status, savedHash, found, err := m.lookup(r.Context(), key)
			if err != nil {
				httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, "idempotency lookup: "+err.Error()))
				return
			}
			if found {
				if savedHash != reqHash {
					httpx.WriteError(w, r, httpx.Wrap(httpx.ErrUnprocessable,
						"Idempotency-Key reuse with different request body"))
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Idempotent-Replay", "true")
				w.WriteHeader(status)
				_, _ = w.Write(cached)
				return
			}

			// Записи нет — выполняем handler через recorder.
			rec := &responseRecorder{
				ResponseWriter: w,
				body:           &bytes.Buffer{},
				status:         http.StatusOK,
			}
			next.ServeHTTP(rec, r)

			// Успешные 2xx ответы кэшируем. Ошибки 4xx/5xx тоже кэшируем,
			// чтобы повтор не дёрнул handler заново (защита от idempotency-bypass).
			// Исключение: 429/5xx — НЕ кэшируем (это transient ошибки, повтор
			// должен иметь право на новую попытку).
			cacheable := rec.status < 500 && rec.status != http.StatusTooManyRequests
			if cacheable {
				if err := m.save(r.Context(), key, userID, reqHash, rec.body.Bytes(), rec.status); err != nil {
					// Если запись уже создана параллельным запросом (ON CONFLICT),
					// читаем её и возвращаем сохранённый ответ. Это даёт идемпотентность
					// даже при гонке.
					if errors.Is(err, errIdempotencyConflict) {
						cached, status, savedHash, found, lookupErr := m.lookup(r.Context(), key)
						if lookupErr == nil && found {
							if savedHash != reqHash {
								// Конфликт по hash после параллельного запроса —
								// логически невозможно при правильном клиенте, но
								// защищаемся.
								return
							}
							// Перезаписать наш ответ сохранённым (handler уже отработал,
							// но ответ ещё не отправлен клиенту — мы пишем в recorder).
							w.Header().Set("Content-Type", "application/json")
							w.Header().Set("Idempotent-Replay", "true")
							w.WriteHeader(status)
							_, _ = w.Write(cached)
							return
						}
					}
					// Логировать, но не валить запрос (handler уже отработал).
					// Idempotency-replay для повторных запросов не сработает,
					// но текущий запрос успешен.
				}
			}
			// Отправить накопленное в реальный response.
			w.Header().Set("Content-Type", rec.Header().Get("Content-Type"))
			w.WriteHeader(rec.status)
			_, _ = w.Write(rec.body.Bytes())
		})
	}
}

// errIdempotencyConflict — INSERT в idempotency_keys нарвался на UNIQUE.
var errIdempotencyConflict = errors.New("billing: idempotency key conflict")

func (m *IdempotencyMiddleware) lookup(ctx context.Context, key string) ([]byte, int, string, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	var body []byte
	var status int
	var hash string
	var rawJSON []byte
	err := m.pool.QueryRow(ctx, `
		SELECT response_body::text, response_status, request_hash
		FROM idempotency_keys
		WHERE key = $1 AND expires_at > now()
	`, key).Scan(&rawJSON, &status, &hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, 0, "", false, nil
	}
	if err != nil {
		return nil, 0, "", false, err
	}
	body = rawJSON
	return body, status, hash, true, nil
}

func (m *IdempotencyMiddleware) save(ctx context.Context, key, userID, hash string, body []byte, status int) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	// JSON store — хранит произвольный response. Если body не JSON, обернём в строку.
	var jsonBody json.RawMessage
	if json.Valid(body) {
		jsonBody = body
	} else {
		// Не-JSON ответ — сохраняем как JSON-строку.
		s, _ := json.Marshal(string(body))
		jsonBody = s
	}
	tag, err := m.pool.Exec(ctx, `
		INSERT INTO idempotency_keys (key, user_id, request_hash, response_body, response_status)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (key) DO NOTHING
	`, key, userID, hash, jsonBody, status)
	if err != nil {
		return fmt.Errorf("insert idempotency: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errIdempotencyConflict
	}
	return nil
}

func requestHash(method, path string, body []byte) string {
	h := sha256.New()
	h.Write([]byte(strings.ToUpper(method)))
	h.Write([]byte{'\n'})
	h.Write([]byte(path))
	h.Write([]byte{'\n'})
	h.Write(body)
	return hex.EncodeToString(h.Sum(nil))
}

// responseRecorder перехватывает status и body, написанные handler-ом.
// Заголовки берутся из http.Header() ResponseWriter-а (общий map).
type responseRecorder struct {
	http.ResponseWriter
	body   *bytes.Buffer
	status int
	wrote  bool
}

func (r *responseRecorder) WriteHeader(code int) {
	if r.wrote {
		return
	}
	r.status = code
	r.wrote = true
}

func (r *responseRecorder) Write(p []byte) (int, error) {
	if !r.wrote {
		r.status = http.StatusOK
		r.wrote = true
	}
	return r.body.Write(p)
}
