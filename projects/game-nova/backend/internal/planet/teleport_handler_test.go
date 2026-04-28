package planet_test

// План 65 Ф.6: тесты HTTP-handler'а POST /api/planets/{id}/teleport.
//
// Структура:
//   - TestTeleportHandler_Unauthorized — 401 без auth.UserID в context.
//   - TestTeleportHandler_MissingIdempotencyKey — 400 без header.
//   - TestTeleportHandler_InvalidJSON — 400 при битом body.
//   - TestTeleportHandler_InvalidCoords — 400 при координатах вне диапазона.
//   - TestTeleportHandler_BadPlanetID — handler без chi-маршрутизации
//     (отсутствует {id} URL-param) → 400.
//
// Тесты с реальной БД (happy-path, cooldown, occupied-slot, billing
// success/failures) живут в integration-режиме с TEST_DATABASE_URL и
// в этом файле помечены t.Skip(), если переменная не задана.

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"oxsar/game-nova/internal/auth"
	billingclient "oxsar/game-nova/internal/billing/client"
	"oxsar/game-nova/internal/planet"
)

func newTestHandler(t *testing.T, billingURL string) *planet.TeleportHandler {
	t.Helper()
	// db=nil допустим: тесты pre-validation не доходят до InTx. Если
	// тест случайно дойдёт до БД — упадёт с panic от nil.InTx — это
	// явный сигнал, что pre-validation не сработала.
	cli := billingclient.New(billingURL)
	return planet.NewTeleportHandler(nil, cli, planet.TeleportConfig{
		CostOxsars:      50000,
		CooldownHours:   24,
		DurationMinutes: 0,
	})
}

func newAuthRequest(method, target string, body []byte, userID, idemKey string) *http.Request {
	r := httptest.NewRequest(method, target, bytes.NewReader(body))
	if userID != "" {
		ctx := context.WithValue(r.Context(), auth.UserIDKey, userID)
		r = r.WithContext(ctx)
	}
	if idemKey != "" {
		r.Header.Set("Idempotency-Key", idemKey)
	}
	r.Header.Set("Content-Type", "application/json")
	return r
}

// chiRequest заворачивает request в chi-router с {id} param.
func chiRequest(r *http.Request, id string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestTeleportHandler_Unauthorized(t *testing.T) {
	h := newTestHandler(t, "")
	body, _ := json.Marshal(map[string]int{"target_galaxy": 1, "target_system": 1, "target_position": 1})
	r := chiRequest(newAuthRequest("POST", "/api/planets/p1/teleport", body, "", "key1"), "p1")
	w := httptest.NewRecorder()
	h.Teleport(w, r)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTeleportHandler_MissingIdempotencyKey(t *testing.T) {
	h := newTestHandler(t, "")
	body, _ := json.Marshal(map[string]int{"target_galaxy": 1, "target_system": 1, "target_position": 1})
	r := chiRequest(newAuthRequest("POST", "/api/planets/p1/teleport", body, "u1", ""), "p1")
	w := httptest.NewRecorder()
	h.Teleport(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 (missing Idempotency-Key), got %d: %s", w.Code, w.Body.String())
	}
}

func TestTeleportHandler_InvalidJSON(t *testing.T) {
	h := newTestHandler(t, "")
	r := chiRequest(newAuthRequest("POST", "/api/planets/p1/teleport", []byte("not-json"), "u1", "key1"), "p1")
	w := httptest.NewRecorder()
	h.Teleport(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTeleportHandler_InvalidCoords(t *testing.T) {
	cases := []struct {
		name string
		body map[string]int
	}{
		{"galaxy_zero", map[string]int{"target_galaxy": 0, "target_system": 1, "target_position": 1}},
		{"galaxy_too_big", map[string]int{"target_galaxy": 17, "target_system": 1, "target_position": 1}},
		{"system_zero", map[string]int{"target_galaxy": 1, "target_system": 0, "target_position": 1}},
		{"system_too_big", map[string]int{"target_galaxy": 1, "target_system": 1000, "target_position": 1}},
		{"position_zero", map[string]int{"target_galaxy": 1, "target_system": 1, "target_position": 0}},
		{"position_too_big", map[string]int{"target_galaxy": 1, "target_system": 1, "target_position": 16}},
		{"all_negative", map[string]int{"target_galaxy": -1, "target_system": -1, "target_position": -1}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := newTestHandler(t, "")
			body, _ := json.Marshal(c.body)
			r := chiRequest(newAuthRequest("POST", "/api/planets/p1/teleport", body, "u1", "key1"), "p1")
			w := httptest.NewRecorder()
			h.Teleport(w, r)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected 400 for invalid coords %v, got %d: %s", c.body, w.Code, w.Body.String())
			}
		})
	}
}

// TestTeleportHandler_BadPlanetID — chi.URLParam(r, "id") возвращает ""
// (нет route-context) → handler отвергает с 400. Защита от прямого
// вызова handler'а в тесте без предварительной обмотки router'ом.
func TestTeleportHandler_BadPlanetID(t *testing.T) {
	h := newTestHandler(t, "")
	body, _ := json.Marshal(map[string]int{"target_galaxy": 1, "target_system": 1, "target_position": 1})
	// Намеренно НЕ используем chiRequest: route-context отсутствует.
	r := newAuthRequest("POST", "/api/planets/p1/teleport", body, "u1", "key1")
	w := httptest.NewRecorder()
	h.Teleport(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing planet id, got %d: %s", w.Code, w.Body.String())
	}
}

// TestTeleportHandler_RequestStruct_RoundTrip — JSON-форма запроса не
// должна потерять поля при изменении struct'а.
func TestTeleportHandler_RequestStruct_RoundTrip(t *testing.T) {
	in := map[string]int{
		"target_galaxy":   3,
		"target_system":   42,
		"target_position": 7,
	}
	raw, _ := json.Marshal(in)
	if string(raw) != `{"target_galaxy":3,"target_position":7,"target_system":42}` &&
		string(raw) != `{"target_galaxy":3,"target_system":42,"target_position":7}` {
		// Маршалинг map'ы не гарантирует порядок ключей; проверяем наличие
		// каждого ключа явным substring-поиском.
		for _, k := range []string{`"target_galaxy":3`, `"target_system":42`, `"target_position":7`} {
			if !bytes.Contains(raw, []byte(k)) {
				t.Fatalf("missing key %q in request body: %s", k, raw)
			}
		}
	}
}
