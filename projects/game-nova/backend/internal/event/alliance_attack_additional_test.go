package event_test

// План 65 Ф.5: тесты HandleAllianceAttackAdditional (no-op handler).
//
// Концепция: handler — намеренный no-op для совместимости с legacy
// origin (см. handler-doc в handlers.go). Тестируем:
//   - вызов с tx==nil безопасен (handler не делает SQL);
//   - возвращает nil без побочных эффектов;
//   - валиден для любого payload (включая пустой и невалидный JSON —
//     handler не парсит payload, так как не использует его).

import (
	"context"
	"encoding/json"
	"testing"

	"oxsar/game-nova/internal/event"
)

// TestAllianceAttackAdditional_NoOp — handler возвращает nil без
// обращения к tx (поэтому tx==nil безопасно).
func TestAllianceAttackAdditional_NoOp(t *testing.T) {
	cases := []struct {
		name    string
		payload json.RawMessage
	}{
		{"nil_payload", nil},
		{"empty_payload", json.RawMessage(``)},
		{"empty_object", json.RawMessage(`{}`)},
		{"foreign_payload", json.RawMessage(`{"foo":"bar","n":42}`)},
		{"malformed_json", json.RawMessage(`{not valid json`)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := event.Event{
				ID:      "test-evt",
				Kind:    event.KindAllianceAttackAdditional,
				Payload: tc.payload,
			}
			err := event.HandleAllianceAttackAdditional(context.Background(), nil, e)
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}
		})
	}
}
