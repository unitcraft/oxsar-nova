package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"

	"github.com/go-chi/chi/v5"

	"github.com/oxsar/nova/backend/internal/auth"
	"github.com/oxsar/nova/backend/internal/httpx"
	"github.com/oxsar/nova/backend/internal/repo"
	"github.com/oxsar/nova/backend/pkg/ids"
)

const (
	maxBodyLen   = 500
	historyLimit = 50
	writeBuf     = 32
)

// Handler — HTTP + WebSocket handler для чата.
type Handler struct {
	hub *Hub
	db  repo.Exec
}

func NewHandler(hub *Hub, db repo.Exec) *Handler {
	return &Handler{hub: hub, db: db}
}

// channelFor разрешает channel по параметру роута:
//   - "global"   → channel "global"
//   - "alliance" → channel "ally:<alliance_id>" (из users.alliance_id)
func (h *Handler) channelFor(ctx context.Context, kind, userID string) (string, error) {
	switch kind {
	case "global":
		return "global", nil
	case "alliance":
		var allianceID *string
		if err := h.db.Pool().QueryRow(ctx,
			`SELECT alliance_id FROM users WHERE id=$1`, userID).Scan(&allianceID); err != nil {
			return "", fmt.Errorf("chat: read alliance_id: %w", err)
		}
		if allianceID == nil {
			return "", fmt.Errorf("chat: not in alliance")
		}
		return "ally:" + *allianceID, nil
	default:
		return "", fmt.Errorf("chat: unknown channel kind %q", kind)
	}
}

// History GET /api/chat/{kind}/history
// Возвращает последние N сообщений канала (JSON-array, без WebSocket).
func (h *Handler) History(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	kind := chi.URLParam(r, "kind")
	channel, err := h.channelFor(r.Context(), kind, uid)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
		return
	}

	rows, err := h.db.Pool().Query(r.Context(), `
		SELECT cm.id, cm.author_id, COALESCE(u.username, ''), cm.body, cm.created_at, cm.edited_at
		FROM chat_messages cm
		LEFT JOIN users u ON u.id = cm.author_id
		WHERE cm.channel = $1 AND cm.deleted_at IS NULL
		ORDER BY cm.created_at DESC
		LIMIT $2
	`, channel, historyLimit)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	defer rows.Close()

	var msgs []Message
	for rows.Next() {
		var m Message
		var createdAt time.Time
		var editedAt *time.Time
		if err := rows.Scan(&m.ID, &m.AuthorID, &m.AuthorName, &m.Body, &createdAt, &editedAt); err != nil {
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
			return
		}
		m.Channel = channel
		m.CreatedAt = createdAt.UTC().Format(time.RFC3339)
		m.Kind = "msg"
		if editedAt != nil {
			s := editedAt.UTC().Format(time.RFC3339)
			m.EditedAt = &s
		}
		msgs = append(msgs, m)
	}
	if err := rows.Err(); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	// Реверс: отдаём от старых к новым.
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	httpx.WriteJSON(w, r, http.StatusOK, msgs)
}

// Connect GET /api/chat/{kind}/ws (WebSocket upgrade)
func (h *Handler) Connect(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	kind := chi.URLParam(r, "kind")
	channel, err := h.channelFor(r.Context(), kind, uid)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
		return
	}

	// Читаем username для broadcast.
	var username string
	_ = h.db.Pool().QueryRow(r.Context(), `SELECT username FROM users WHERE id=$1`, uid).Scan(&username)

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // CORS проверяется вышестоящим middleware
	})
	if err != nil {
		return
	}
	defer conn.CloseNow()

	c := &client{channel: channel, send: make(chan Message, writeBuf)}
	h.hub.register(c)
	defer h.hub.unregister(c)

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Горутина: пишем входящие broadcast-сообщения клиенту + ping каждые 30s.
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := conn.Ping(ctx); err != nil {
					cancel()
					return
				}
			case msg, open := <-c.send:
				if !open {
					return
				}
				if err := wsjson.Write(ctx, conn, msg); err != nil {
					cancel()
					return
				}
			}
		}
	}()

	// Читаем от клиента, персистируем, бродкастим.
	type inbound struct {
		Body string `json:"body"`
	}
	for {
		var in inbound
		if err := wsjson.Read(ctx, conn, &in); err != nil {
			break
		}
		body := strings.TrimSpace(in.Body)
		if body == "" || len([]rune(body)) > maxBodyLen {
			continue
		}

		msgID := ids.New()
		now := time.Now().UTC()
		if _, err := h.db.Pool().Exec(ctx, `
			INSERT INTO chat_messages (id, channel, author_id, body, created_at)
			VALUES ($1, $2, $3, $4, $5)
		`, msgID, channel, uid, body, now); err != nil {
			break
		}

		msg := Message{
			ID:         msgID,
			Channel:    channel,
			AuthorID:   uid,
			AuthorName: username,
			Body:       body,
			CreatedAt:  now.Format(time.RFC3339),
			Kind:       "msg",
		}
		h.hub.Broadcast(ctx, msg)
	}

	conn.Close(websocket.StatusNormalClosure, "")
}

// Send POST /api/chat/{kind}/send (REST fallback для клиентов без WS)
func (h *Handler) Send(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	kind := chi.URLParam(r, "kind")
	channel, err := h.channelFor(r.Context(), kind, uid)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
		return
	}

	var req struct {
		Body string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	body := strings.TrimSpace(req.Body)
	if body == "" || len([]rune(body)) > maxBodyLen {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "body empty or too long"))
		return
	}

	var username string
	_ = h.db.Pool().QueryRow(r.Context(), `SELECT username FROM users WHERE id=$1`, uid).Scan(&username)

	msgID := ids.New()
	now := time.Now().UTC()
	if _, err := h.db.Pool().Exec(r.Context(), `
		INSERT INTO chat_messages (id, channel, author_id, body, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, msgID, channel, uid, body, now); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}

	msg := Message{
		ID:         msgID,
		Channel:    channel,
		AuthorID:   uid,
		AuthorName: username,
		Body:       body,
		CreatedAt:  now.Format(time.RFC3339),
		Kind:       "msg",
	}
	h.hub.Broadcast(r.Context(), msg)
	httpx.WriteJSON(w, r, http.StatusCreated, msg)
}

const editWindow = 5 * time.Minute

// EditMessage PATCH /api/chat/messages/{id}
// Только автор, только в течение editWindow после создания.
func (h *Handler) EditMessage(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	msgID := chi.URLParam(r, "id")

	var req struct {
		Body string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	body := strings.TrimSpace(req.Body)
	if body == "" || len([]rune(body)) > maxBodyLen {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "body empty or too long"))
		return
	}

	now := time.Now().UTC()
	var channel, authorID string
	var createdAt time.Time
	err := h.db.Pool().QueryRow(r.Context(),
		`SELECT channel, author_id, created_at FROM chat_messages WHERE id=$1 AND deleted_at IS NULL`,
		msgID,
	).Scan(&channel, &authorID, &createdAt)
	if err != nil {
		httpx.WriteError(w, r, httpx.ErrNotFound)
		return
	}
	if authorID != uid {
		httpx.WriteError(w, r, httpx.ErrForbidden)
		return
	}
	if now.Sub(createdAt) > editWindow {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrForbidden, "edit window expired"))
		return
	}

	if _, err := h.db.Pool().Exec(r.Context(),
		`UPDATE chat_messages SET body=$1, edited_at=$2 WHERE id=$3`,
		body, now, msgID,
	); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}

	editedAt := now.Format(time.RFC3339)
	msg := Message{
		ID:        msgID,
		Channel:   channel,
		AuthorID:  uid,
		Body:      body,
		EditedAt:  &editedAt,
		Kind:      "edit",
	}
	h.hub.Broadcast(r.Context(), msg)
	httpx.WriteJSON(w, r, http.StatusOK, msg)
}

// DeleteMessage DELETE /api/chat/messages/{id}
// Только автор, только в течение editWindow после создания.
func (h *Handler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	msgID := chi.URLParam(r, "id")

	now := time.Now().UTC()
	var channel, authorID string
	var createdAt time.Time
	err := h.db.Pool().QueryRow(r.Context(),
		`SELECT channel, author_id, created_at FROM chat_messages WHERE id=$1 AND deleted_at IS NULL`,
		msgID,
	).Scan(&channel, &authorID, &createdAt)
	if err != nil {
		httpx.WriteError(w, r, httpx.ErrNotFound)
		return
	}
	if authorID != uid {
		httpx.WriteError(w, r, httpx.ErrForbidden)
		return
	}
	if now.Sub(createdAt) > editWindow {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrForbidden, "delete window expired"))
		return
	}

	if _, err := h.db.Pool().Exec(r.Context(),
		`UPDATE chat_messages SET deleted_at=$1 WHERE id=$2`,
		now, msgID,
	); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}

	msg := Message{ID: msgID, Channel: channel, AuthorID: uid, Kind: "delete"}
	h.hub.Broadcast(r.Context(), msg)
	w.WriteHeader(http.StatusNoContent)
}
