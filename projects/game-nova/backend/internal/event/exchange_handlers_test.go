package event

import (
	"encoding/json"
	"testing"
)

// Smoke-тесты для exchange-handler payload'ов. Полный integration-тест с
// реальной БД (TEST_DATABASE_URL) — в exchange/repo_pgx_integration_test.go
// (не реализован — auto-skip без БД, см. план 68 Ф.7).
//
// Здесь проверяем что payload-структуры корректно (де)сериализуются и
// что handler правильно сообщает об отсутствии обязательных полей.

func TestExchangeExpirePayload_RoundTrip(t *testing.T) {
	src := ExchangeExpirePayload{LotID: "lot-123"}
	raw, err := json.Marshal(src)
	if err != nil {
		t.Fatal(err)
	}
	var got ExchangeExpirePayload
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if got != src {
		t.Errorf("round-trip mismatch: %+v vs %+v", got, src)
	}
}

func TestExchangeBanPayload_RoundTrip(t *testing.T) {
	src := ExchangeBanPayload{
		SellerUserID: "user-X",
		Reason:       "fraud_detected",
	}
	raw, err := json.Marshal(src)
	if err != nil {
		t.Fatal(err)
	}
	var got ExchangeBanPayload
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if got != src {
		t.Errorf("round-trip mismatch: %+v vs %+v", got, src)
	}
}

// Проверяет что handler возвращает явную ошибку при пустом lot_id.
// Важно: без этого guard'а будет обращение к БД с empty UUID и непонятный
// pgx-error в логах.
func TestHandleExchangeExpire_RejectsEmptyLotID(t *testing.T) {
	e := Event{
		ID:      "evt-1",
		Kind:    KindExchangeExpire,
		Payload: json.RawMessage(`{}`),
	}
	err := HandleExchangeExpire(t.Context(), nil, e)
	if err == nil {
		t.Fatal("expected error for empty lot_id, got nil")
	}
}

func TestHandleExchangeExpire_InvalidJSON(t *testing.T) {
	e := Event{
		ID:      "evt-1",
		Kind:    KindExchangeExpire,
		Payload: json.RawMessage(`not json`),
	}
	err := HandleExchangeExpire(t.Context(), nil, e)
	if err == nil {
		t.Fatal("expected error for invalid json, got nil")
	}
}

func TestHandleExchangeBan_RejectsEmptySellerID(t *testing.T) {
	e := Event{
		ID:      "evt-1",
		Kind:    KindExchangeBan,
		Payload: json.RawMessage(`{"reason":"x"}`),
	}
	err := HandleExchangeBan(t.Context(), nil, e)
	if err == nil {
		t.Fatal("expected error for empty seller_user_id, got nil")
	}
}

func TestHandleExchangeBan_InvalidJSON(t *testing.T) {
	e := Event{
		ID:      "evt-1",
		Kind:    KindExchangeBan,
		Payload: json.RawMessage(`not json`),
	}
	err := HandleExchangeBan(t.Context(), nil, e)
	if err == nil {
		t.Fatal("expected error for invalid json, got nil")
	}
}
