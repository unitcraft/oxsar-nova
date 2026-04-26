// Package chat — WebSocket-чат (Global / Alliance / PM).
//
// Hub — broadcast-сервер. Каждый подключённый клиент регистрирует себя
// на один channel (строку); Hub рассылает входящие сообщения всем
// клиентам этого channel.
//
// Multi-instance (план 32 Ф.5):
// При наличии Redis Hub публикует все сообщения в pub/sub-канал
// "chat:<channel>"; отдельная горутина (runSubscriber) читает оттуда
// и рассылает локальным клиентам. Так сообщения от backend-1 видны
// клиентам, подключённым к backend-2.
//
// Если Redis недоступен (rdb=nil или PSubscribe упал) — Hub продолжает
// работать через local-broadcast: degradation до single-instance,
// без полного отказа чата.
package chat

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
)

// Message — сообщение в чате (отдано клиенту как JSON).
// Kind: "msg" (новое), "edit" (изменено), "delete" (удалено).
type Message struct {
	ID         string  `json:"id"`
	Channel    string  `json:"channel"`
	AuthorID   string  `json:"author_id"`
	AuthorName string  `json:"author_name"`
	Body       string  `json:"body"`
	CreatedAt  string  `json:"created_at"`
	EditedAt   *string `json:"edited_at,omitempty"`
	Kind       string  `json:"kind"` // "msg" | "edit" | "delete"
}

type client struct {
	channel string
	send    chan Message
}

// Hub управляет подключёнными клиентами и рассылкой сообщений.
//
// Если rdb != nil — публикация идёт через Redis pub/sub, а локальная
// рассылка происходит из subscriber-горутины. Иначе — прямой
// local-broadcast (single-instance fallback).
type Hub struct {
	mu      sync.RWMutex
	clients map[*client]struct{}

	rdb            *redis.Client
	pubSubRunning  atomic.Bool // true если subscriber успешно стартовал
	subscriberDone chan struct{}
	pattern        string // "chat:*" — для теста можно подменить
	log            *slog.Logger
}

// NewHub возвращает Hub без Redis (single-instance режим).
// Старая сигнатура — handler-ы и тесты не сломаются.
func NewHub() *Hub {
	return &Hub{
		clients: make(map[*client]struct{}),
		pattern: "chat:*",
		log:     slog.Default(),
	}
}

// NewHubWithRedis создаёт Hub с Redis pub/sub fan-out'ом.
//
// При rdb=nil ведёт себя как NewHub(). Subscriber-горутина запускается
// фоном и читает chat:* до отмены ctx. Используйте Close() для явного
// shutdown'а.
func NewHubWithRedis(ctx context.Context, rdb *redis.Client, log *slog.Logger) *Hub {
	if log == nil {
		log = slog.Default()
	}
	h := &Hub{
		clients:        make(map[*client]struct{}),
		rdb:            rdb,
		subscriberDone: make(chan struct{}),
		pattern:        "chat:*",
		log:            log,
	}
	if rdb != nil {
		go h.runSubscriber(ctx)
	} else {
		close(h.subscriberDone)
	}
	return h
}

// Close дожидается остановки subscriber-горутины. Hub не использует
// сетевые ресурсы вне Redis — закрывать самих клиентов не нужно
// (обрыв WS обрабатывается в handler.go).
func (h *Hub) Close() {
	if h.subscriberDone != nil {
		<-h.subscriberDone
	}
}

func (h *Hub) register(c *client) {
	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()
}

func (h *Hub) unregister(c *client) {
	h.mu.Lock()
	delete(h.clients, c)
	h.mu.Unlock()
	close(c.send)
}

// Broadcast рассылает сообщение всем клиентам канала на всех инстансах.
//
// Если rdb доступен и subscriber работает — идёт через Redis. Это
// гарантирует, что клиенты на других backend-инстансах получат
// сообщение через свой subscriber. На текущем инстансе сообщение
// тоже придёт (Redis psubscribe доставляет publisher'у).
//
// Если rdb=nil или Publish упал (Redis недоступен) — fallback на
// прямой local-broadcast. Игроки на других инстансах не получат, но
// этот инстанс продолжит работать.
func (h *Hub) Broadcast(ctx context.Context, msg Message) {
	if h.rdb != nil && h.pubSubRunning.Load() {
		data, err := json.Marshal(msg)
		if err == nil {
			if err := h.rdb.Publish(ctx, "chat:"+msg.Channel, data).Err(); err == nil {
				return
			}
			h.log.WarnContext(ctx, "chat: redis publish failed, falling back to local",
				slog.String("err", err.Error()),
				slog.String("channel", msg.Channel))
		}
	}
	h.broadcastLocal(msg)
}

// broadcastLocal рассылает сообщение клиентам этого инстанса.
// Используется и из publisher'а (fallback), и из subscriber'а
// (нормальный путь pub/sub).
func (h *Hub) broadcastLocal(msg Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients {
		if c.channel != msg.Channel {
			continue
		}
		select {
		case c.send <- msg:
		default:
			// медленный клиент — пропускаем, не блокируем broadcast
		}
	}
}

// runSubscriber подписан на chat:* в Redis, читает сообщения и
// рассылает их локальным клиентам через broadcastLocal. При обрыве
// connection делает retry с backoff'ом до отмены ctx.
func (h *Hub) runSubscriber(ctx context.Context) {
	defer close(h.subscriberDone)
	backoff := time.Second
	const maxBackoff = 30 * time.Second

	for ctx.Err() == nil {
		ps := h.rdb.PSubscribe(ctx, h.pattern)
		// Receive (а не Channel) даёт явную синхронизацию: ждём,
		// пока сервер подтвердит подписку, и только потом включаем
		// Publish-путь. Иначе несколько первых сообщений могут
		// уйти в пустоту.
		if _, err := ps.Receive(ctx); err != nil {
			_ = ps.Close()
			if ctx.Err() != nil {
				return
			}
			h.log.WarnContext(ctx, "chat: redis psubscribe failed, retrying",
				slog.String("err", err.Error()),
				slog.Duration("backoff", backoff))
			h.pubSubRunning.Store(false)
			select {
			case <-ctx.Done():
				return
			case <-time.After(backoff):
			}
			if backoff < maxBackoff {
				backoff *= 2
			}
			continue
		}

		h.pubSubRunning.Store(true)
		backoff = time.Second
		ch := ps.Channel()
		h.log.InfoContext(ctx, "chat: redis subscriber active", slog.String("pattern", h.pattern))

		for m := range ch {
			channel := strings.TrimPrefix(m.Channel, "chat:")
			var msg Message
			if err := json.Unmarshal([]byte(m.Payload), &msg); err != nil {
				h.log.WarnContext(ctx, "chat: bad payload",
					slog.String("err", err.Error()),
					slog.String("channel", m.Channel))
				continue
			}
			// Если Channel в Message не совпадает с patterned key —
			// верим payload'у (он надёжнее, key — для роутинга).
			if msg.Channel == "" {
				msg.Channel = channel
			}
			h.broadcastLocal(msg)
		}

		// ch закрылся — обычно это означает обрыв; цикл повторит
		// PSubscribe.
		h.pubSubRunning.Store(false)
		_ = ps.Close()
		if ctx.Err() != nil {
			return
		}
		h.log.WarnContext(ctx, "chat: redis subscriber channel closed, reconnecting")
	}
}
