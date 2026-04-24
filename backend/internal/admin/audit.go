// Package admin — audit-лог деструктивных операций.
//
// AuditMiddleware навешивается на non-GET роуты внутри /api/admin/*.
// После успешного (2xx) ответа middleware кладёт запись в таблицу
// admin_audit_log: admin_id, действие, target (если есть), тело
// запроса (без потенциальных секретов).
//
// Почему после ответа: если handler упал — операция не выполнилась,
// аудит-запись была бы ложной. Проверяем status code.
//
// Читать аудит: GET /api/admin/audit (см. ListAudit в этом файле).

package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/oxsar/nova/backend/internal/auth"
	"github.com/oxsar/nova/backend/internal/httpx"
	"github.com/oxsar/nova/backend/internal/repo"
)

// максимальный размер payload, который сохраняем в БД (байт).
// Всё больше усечём до этого предела — реальные админские запросы
// умещаются в 1 КБ, но defense-in-depth против случайного огромного JSON.
const maxAuditPayloadBytes = 4096

// actionFromRoute выводит action-ключ из метода + chi route-pattern.
// Пример: POST /users/{id}/ban → "users.ban", PUT /automsgs/{key} → "automsgs.update".
// Логика проста: последний сегмент пути, если он не параметр, иначе "update".
func actionFromRoute(method, pattern string) string {
	// Уберём ведущий /api/admin, если middleware навешан глубже.
	pattern = strings.TrimPrefix(pattern, "/api/admin")
	pattern = strings.TrimPrefix(pattern, "/admin")
	pattern = strings.TrimPrefix(pattern, "/")

	parts := strings.Split(pattern, "/")
	// Найдём первый не-параметр (сущность: users/automsgs/events/...)
	var entity, verb string
	for _, p := range parts {
		if p == "" {
			continue
		}
		if strings.HasPrefix(p, "{") {
			continue
		}
		if entity == "" {
			entity = p
		} else {
			verb = p
		}
	}

	if verb == "" {
		// POST /users → users.create, DELETE /users/{id} → users.delete, PUT → users.update
		switch method {
		case http.MethodPost:
			verb = "create"
		case http.MethodPut, http.MethodPatch:
			verb = "update"
		case http.MethodDelete:
			verb = "delete"
		default:
			verb = strings.ToLower(method)
		}
	}
	if entity == "" {
		return strings.ToLower(method)
	}
	return entity + "." + verb
}

// clientIP — извлекаем IP клиента с учётом X-Forwarded-For (за nginx).
func clientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		// первый IP из списка — настоящий клиент
		if i := strings.Index(fwd, ","); i >= 0 {
			return strings.TrimSpace(fwd[:i])
		}
		return strings.TrimSpace(fwd)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// AuditMiddleware пишет запись в admin_audit_log для не-GET запросов
// с 2xx-ответом.
//
// Использование:
//
//	ar.Route("/admin", func(ar chi.Router) {
//	    ar.Use(admin.AdminOnly(db))
//	    ar.Use(admin.AuditMiddleware(db))  // ← после AdminOnly
//	    …
//	})
func AuditMiddleware(db repo.Exec) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet || r.Method == http.MethodHead {
				next.ServeHTTP(w, r)
				return
			}

			// Считаем тело: оно может понадобиться handler'у дальше,
			// потому читаем в буфер и подставляем r.Body заново.
			var payload []byte
			if r.Body != nil && r.ContentLength != 0 {
				buf, err := io.ReadAll(io.LimitReader(r.Body, maxAuditPayloadBytes+1))
				if err == nil {
					payload = buf
					r.Body = io.NopCloser(bytes.NewReader(buf))
				}
			}

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			status := ww.Status()
			if status == 0 {
				status = http.StatusOK
			}
			if status >= 400 {
				// Не пишем аудит для неуспешных — операция не применилась.
				return
			}

			uid, _ := auth.UserID(r.Context())
			if uid == "" {
				// Без admin_id запись бесполезна (и нарушит FK).
				return
			}

			pattern := chi.RouteContext(r.Context()).RoutePattern()
			action := actionFromRoute(r.Method, pattern)
			targetKind, targetID := targetFromChi(r, pattern)

			payloadJSON := sanitizePayload(payload)

			go writeAudit(db, auditEntry{
				AdminID:    uid,
				Action:     action,
				TargetKind: targetKind,
				TargetID:   targetID,
				Payload:    payloadJSON,
				Status:     status,
				IP:         clientIP(r),
				UserAgent:  r.UserAgent(),
			})
		})
	}
}

type auditEntry struct {
	AdminID    string
	Action     string
	TargetKind string
	TargetID   string
	Payload    json.RawMessage
	Status     int
	IP         string
	UserAgent  string
}

func writeAudit(db repo.Exec, e auditEntry) {
	// Отдельный контекст — основной r.Context уже может быть отменён
	// к моменту записи (клиент закрыл соединение). 2 секунды на INSERT
	// — щедро.
	ctx, cancel := contextWithTimeout(2 * time.Second)
	defer cancel()

	var ip any
	if e.IP != "" && e.IP != "<nil>" {
		ip = e.IP
	}

	_, err := db.Pool().Exec(ctx, `
		INSERT INTO admin_audit_log
			(admin_id, action, target_kind, target_id, payload, status, ip, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, e.AdminID, e.Action, e.TargetKind, e.TargetID, e.Payload, e.Status, ip, e.UserAgent)
	if err != nil {
		slog.Warn("admin audit: insert failed",
			slog.String("admin_id", e.AdminID),
			slog.String("action", e.Action),
			slog.String("err", err.Error()))
	}
}

// contextWithTimeout вынесено чтобы не тянуть "context" в общий импорт.
func contextWithTimeout(d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), d)
}

// targetFromChi пытается определить kind+id цели операции из первого
// path-параметра. Для /users/{id}/ban → ("user", <id>).
// Для /automsgs/{key} → ("automsg", <key>).
func targetFromChi(r *http.Request, pattern string) (kind, id string) {
	rctx := chi.RouteContext(r.Context())
	if rctx == nil {
		return "", ""
	}

	// Сущность из паттерна — первый непараметрический сегмент.
	pattern = strings.TrimPrefix(pattern, "/api/admin")
	pattern = strings.TrimPrefix(pattern, "/admin")
	parts := strings.Split(strings.TrimPrefix(pattern, "/"), "/")
	for _, p := range parts {
		if p == "" || strings.HasPrefix(p, "{") {
			continue
		}
		kind = singularize(p)
		break
	}

	// Возьмём первый непустой URL-param (id, key, …).
	for i, name := range rctx.URLParams.Keys {
		if i >= len(rctx.URLParams.Values) {
			break
		}
		v := rctx.URLParams.Values[i]
		if v != "" && name != "*" {
			id = v
			break
		}
	}
	return kind, id
}

func singularize(s string) string {
	// users → user, automsgs → automsg, events → event; messages → message.
	if strings.HasSuffix(s, "sses") { // addresses-подобные — не встречаются, но запасной
		return s
	}
	if strings.HasSuffix(s, "s") {
		return strings.TrimSuffix(s, "s")
	}
	return s
}

// sanitizePayload — убираем потенциально-секретные поля (password, token)
// из JSON-тела до записи в журнал. Если тело не JSON или пустое — пишем {}.
// Если тело длиннее порога — обрезаем и ставим маркер.
func sanitizePayload(raw []byte) json.RawMessage {
	if len(raw) == 0 {
		return json.RawMessage(`{}`)
	}
	if len(raw) > maxAuditPayloadBytes {
		// Превышение лимита — не парсим, записываем маркер.
		return json.RawMessage(`{"_truncated": true}`)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		// не JSON — сохраняем как base64? Нет, достаточно маркера.
		return json.RawMessage(`{"_nonjson": true}`)
	}
	for k := range m {
		low := strings.ToLower(k)
		if strings.Contains(low, "password") || strings.Contains(low, "token") || strings.Contains(low, "secret") {
			m[k] = "***"
		}
	}
	b, err := json.Marshal(m)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return b
}

// AuditEntry — DTO для ListAudit.
type AuditEntry struct {
	ID         string          `json:"id"`
	AdminID    string          `json:"admin_id"`
	AdminName  string          `json:"admin_name"`
	Action     string          `json:"action"`
	TargetKind string          `json:"target_kind"`
	TargetID   string          `json:"target_id"`
	Payload    json.RawMessage `json:"payload"`
	Status     int             `json:"status"`
	IP         string          `json:"ip,omitempty"`
	UserAgent  string          `json:"user_agent,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}

// ListAudit GET /api/admin/audit?from=&to=&admin_id=&action=&target_id=&limit=&offset=
func (h *Handler) ListAudit(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	q := `
		SELECT a.id, a.admin_id, COALESCE(u.username, '') AS admin_name,
		       a.action, a.target_kind, a.target_id, a.payload,
		       a.status, COALESCE(a.ip::text, '') AS ip, a.user_agent, a.created_at
		FROM admin_audit_log a
		LEFT JOIN users u ON u.id = a.admin_id
		WHERE 1=1`
	args := []any{}

	addFilter := func(sql string, v any) {
		args = append(args, v)
		q += " AND " + strings.Replace(sql, "?", "$"+strconv.Itoa(len(args)), 1)
	}

	if v := r.URL.Query().Get("admin_id"); v != "" {
		addFilter("a.admin_id = ?", v)
	}
	if v := r.URL.Query().Get("action"); v != "" {
		addFilter("a.action = ?", v)
	}
	if v := r.URL.Query().Get("target_id"); v != "" {
		addFilter("a.target_id = ?", v)
	}
	if v := r.URL.Query().Get("from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			addFilter("a.created_at >= ?", t)
		}
	}
	if v := r.URL.Query().Get("to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			addFilter("a.created_at <= ?", t)
		}
	}

	q += " ORDER BY a.created_at DESC LIMIT $" + strconv.Itoa(len(args)+1) +
		" OFFSET $" + strconv.Itoa(len(args)+2)
	args = append(args, limit, offset)

	rows, err := h.db.Pool().Query(r.Context(), q, args...)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	defer rows.Close()

	entries := []AuditEntry{}
	for rows.Next() {
		var e AuditEntry
		var payloadBytes []byte
		if err := rows.Scan(&e.ID, &e.AdminID, &e.AdminName, &e.Action,
			&e.TargetKind, &e.TargetID, &payloadBytes, &e.Status, &e.IP,
			&e.UserAgent, &e.CreatedAt); err != nil {
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
			return
		}
		if len(payloadBytes) > 0 {
			e.Payload = payloadBytes
		} else {
			e.Payload = json.RawMessage("{}")
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}

	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{
		"entries": entries,
		"limit":   limit,
		"offset":  offset,
	})
}
