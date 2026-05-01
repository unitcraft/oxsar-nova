package exchange

import "errors"

// Sentinel-ошибки биржи (план 68 Ф.3). HTTP-handler классифицирует их
// и возвращает соответствующие коды (см. handler.go и openapi.yaml).
var (
	// ErrLotNotFound — запрашиваемый лот не существует. → 404.
	ErrLotNotFound = errors.New("exchange: lot not found")

	// ErrNotASeller — операция (cancel) пытается отозвать чужой лот. → 403.
	ErrNotASeller = errors.New("exchange: not the seller")

	// ErrLotNotActive — лот уже sold/cancelled/expired. → 409.
	ErrLotNotActive = errors.New("exchange: lot is not active")

	// ErrCannotBuyOwnLot — попытка купить собственный лот. → 403.
	ErrCannotBuyOwnLot = errors.New("exchange: cannot buy own lot")

	// ErrInsufficientArtefacts — у seller'а меньше N held-артефактов
	// нужного unit_id, чтобы выставить лот. → 422.
	ErrInsufficientArtefacts = errors.New("exchange: insufficient held artefacts to escrow")

	// ErrInsufficientOxsarits — у buyer'а меньше price_oxsarit на счету. → 402.
	ErrInsufficientOxsarits = errors.New("exchange: insufficient oxsarits")

	// ErrPriceCapExceeded — цена за единицу выше rolling-30d AVG × multiplier
	// (антифрод-защита от money-laundering через биржу). → 422.
	ErrPriceCapExceeded = errors.New("exchange: price cap exceeded")

	// ErrPermitRequired — нет permit «Знак торговца». В MVP всегда false
	// (см. AlwaysAllowPermit), оставлено в API для будущего gating'а. → 422.
	ErrPermitRequired = errors.New("exchange: merchant permit required")

	// ErrMaxActiveLots — превышен лимит активных лотов на игрока. → 422.
	ErrMaxActiveLots = errors.New("exchange: max active lots reached")

	// ErrMaxQuantity — quantity > max_quantity_per_lot (cap=100 в MVP). → 422.
	ErrMaxQuantity = errors.New("exchange: quantity exceeds per-lot maximum")

	// ErrInvalidExpiry — expires_in_hours вне допустимого диапазона. → 400.
	ErrInvalidExpiry = errors.New("exchange: invalid expires_in_hours")

	// ErrInvalidQuantity — quantity <= 0. → 400.
	ErrInvalidQuantity = errors.New("exchange: quantity must be positive")

	// ErrInvalidPrice — price_oxsarit <= 0. → 400.
	ErrInvalidPrice = errors.New("exchange: price must be positive")

	// ErrUserHasNoPlanet — у buyer'а нет ни одной планеты, артефакты
	// некуда поселить. Теоретически невозможно, защита от corner-case. → 409.
	ErrUserHasNoPlanet = errors.New("exchange: buyer has no planet")

	// План 72.1.27: Premium + Ban.

	// ErrPremiumLimit — превышен EXCH_PREMIUM_LIST_MAX_SIZE=5 одновременных
	// featured-лотов. → 429.
	ErrPremiumLimit = errors.New("exchange: premium list is full")

	// ErrInsufficientCreditPremium — у юзера меньше credit на premium-cost. → 402.
	ErrInsufficientCreditPremium = errors.New("exchange: not enough credit for premium")

	// ErrAdminRequired — действие (Ban) доступно только админам. → 403.
	ErrAdminRequired = errors.New("exchange: admin role required")

	// ErrLotAlreadyFeatured — лот уже featured (внутри окна 2ч). → 409.
	ErrLotAlreadyFeatured = errors.New("exchange: lot already featured within window")

	// ErrLotBanned — попытка работать с забаненным лотом. → 410.
	ErrLotBanned = errors.New("exchange: lot is banned")
)
