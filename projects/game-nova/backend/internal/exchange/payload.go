package exchange

// Типизированные payload'ы для exchange_history.payload (R13).
//
// Каждое значение event_kind имеет собственную структуру. payload пишется
// при INSERT и парсится при SELECT через json.Marshal/Unmarshal в репозитории.
// Хранение через JSONB позволяет добавлять опциональные поля без миграций
// схемы, а типизированный Go-струкут гарантирует, что код не пишет ad-hoc
// map[string]any (это требование R13).

// HistoryPayloadCreated — payload для event_kind='created'.
//
// При выставлении лота сохраняем ключевые параметры. Это избыточно
// относительно exchange_lots (можно было бы JOIN'ить), но даёт audit
// «как было на момент создания» если позже статус/quantity лота изменились.
type HistoryPayloadCreated struct {
	ArtifactUnitID int   `json:"artifact_unit_id"`
	Quantity       int   `json:"quantity"`
	PriceOxsarit   int64 `json:"price_oxsarit"`
	ExpiresInHours int   `json:"expires_in_hours"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
}

// HistoryPayloadBought — payload для event_kind='bought'.
//
// Сохраняет финансовые параметры на момент покупки (для антифрод-аналитики
// и AVG-расчёта в pricing.go).
type HistoryPayloadBought struct {
	BuyerUserID  string `json:"buyer_user_id"`
	SellerUserID string `json:"seller_user_id"`
	Quantity     int    `json:"quantity"`
	PriceOxsarit int64  `json:"price_oxsarit"`
}

// HistoryPayloadCancelled — payload для event_kind='cancelled'.
//
// Reason — символическое имя причины: 'seller_cancel' (вручную через
// DELETE handler) или 'banned' (отозван при бан-сценарии — но тогда
// event_kind='banned', а не 'cancelled').
type HistoryPayloadCancelled struct {
	Reason string `json:"reason"`
}

// HistoryPayloadExpired — payload для event_kind='expired'.
//
// Системное событие, без actor_user_id. EventID — id события из таблицы
// events, которое триггернуло истечение (для трассировки event-loop).
type HistoryPayloadExpired struct {
	EventID string `json:"event_id"`
}

// HistoryPayloadBanned — payload для event_kind='banned'.
//
// Вызывается KindExchangeBan handler'ом, когда модератор/admin банит
// seller'а — все его активные лоты отзываются. SellerUserID и Reason
// сохраняются для audit-trail.
type HistoryPayloadBanned struct {
	SellerUserID string `json:"seller_user_id"`
	Reason       string `json:"reason"`
	EventID      string `json:"event_id,omitempty"`
}
