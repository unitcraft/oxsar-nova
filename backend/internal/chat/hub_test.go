package chat

import (
	"context"
	"testing"
	"time"
)

// При rdb=nil Hub работает как in-memory (single-instance fallback).
// Эти тесты НЕ требуют Redis и проверяют локальный broadcast-путь.

func TestNewHub_NilRedisLocalBroadcast(t *testing.T) {
	h := NewHub()

	c1 := &client{channel: "global", send: make(chan Message, 1)}
	c2 := &client{channel: "ally:42", send: make(chan Message, 1)}
	h.register(c1)
	h.register(c2)
	defer h.unregister(c1)
	defer h.unregister(c2)

	h.Broadcast(context.Background(), Message{Channel: "global", Body: "hi"})

	select {
	case got := <-c1.send:
		if got.Body != "hi" {
			t.Errorf("c1 got %q", got.Body)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("c1 didn't receive message")
	}

	select {
	case got := <-c2.send:
		t.Errorf("c2 must not receive cross-channel msg: %v", got)
	case <-time.After(50 * time.Millisecond):
		// OK: c2 на другом канале
	}
}

func TestHub_SlowClientNotBlocking(t *testing.T) {
	h := NewHub()
	// send-buffer = 1; заполним и проверим что Broadcast не зависнет.
	c := &client{channel: "global", send: make(chan Message, 1)}
	h.register(c)
	defer h.unregister(c)

	h.Broadcast(context.Background(), Message{Channel: "global", Body: "first"})
	// Второй вызов: канал полон, но Broadcast должен вернуться мгновенно.
	done := make(chan struct{})
	go func() {
		h.Broadcast(context.Background(), Message{Channel: "global", Body: "second"})
		close(done)
	}()
	select {
	case <-done:
		// OK
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Broadcast blocked on slow client")
	}
}

func TestNewHubWithRedis_NilDegradesToLocal(t *testing.T) {
	// При rdb=nil NewHubWithRedis ведёт себя как NewHub — Broadcast
	// не зовёт Publish (был бы nil-pointer panic).
	h := NewHubWithRedis(context.Background(), nil, nil)
	defer h.Close()

	c := &client{channel: "global", send: make(chan Message, 1)}
	h.register(c)
	defer h.unregister(c)

	// Не должно паниковать
	h.Broadcast(context.Background(), Message{Channel: "global", Body: "ok"})

	select {
	case got := <-c.send:
		if got.Body != "ok" {
			t.Errorf("got %q", got.Body)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("no message received from local fallback")
	}
}
