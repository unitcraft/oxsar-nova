package planet

// План 65 Ф.6 (D-032+U-009): премиум-телепорт планеты на новые
// координаты, оплата оксарами через billing-service.
//
// Концепция:
//   - HTTP-handler принимает POST /api/planets/{id}/teleport с
//     обязательным заголовком Idempotency-Key (RFC). Idempotency-кеш
//     даёт middleware, поверх роута (см. cmd/server/main.go).
//   - Валидация: ownership планеты, координаты в диапазонах
//     planets.coords_range (galaxy 1..16, system 1..999, position 1..15),
//     cooldown по users.last_planet_teleport_at, целевой slot не занят.
//   - Списание стоимости через billing-client.Spend (план 77).
//     При ErrInsufficientOxsar → 402, ErrIdempotencyConflict → 409,
//     ErrBillingUnavailable → 503.
//   - INSERT events с Kind=KindTeleportPlanet и fire_at = now + duration.
//     Сам телепорт (UPDATE planets.galaxy/system/position +
//     users.last_planet_teleport_at) выполняет event-handler в worker'е,
//     см. internal/event/teleport_handler.go.
//
// Артефактный гейтинг (ARTEFACT_PLANET_TELEPORTER в legacy origin)
// в nova не реализован — только оплата (см. docs/simplifications.md).
//
// Per-universe (R10): nova на уровне БД single-universe (universes-membership
// в identity-сервисе, см. план 36). Cross-universe телепорт физически
// невозможен — все planets в одной БД и одной вселенной по deployment'у.

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/internal/auth"
	billingclient "oxsar/game-nova/internal/billing/client"
	"oxsar/game-nova/internal/event"
	"oxsar/game-nova/internal/httpx"
	"oxsar/game-nova/internal/repo"
	"oxsar/game-nova/pkg/metrics"
)

// pgxTx — alias на pgx.Tx, чтобы код handler'а не зависел от полного
// пакетного пути в каждой строке.
type pgxTx = pgx.Tx

// pgxErrNoRows зеркалит pgx.ErrNoRows для удобства использования с
// errors.Is в pre-check.
var pgxErrNoRows = pgx.ErrNoRows

// Локальные sentinel-ошибки для прохождения pre-check логики через единый
// switch errors.Is. Не экспортируются.
var (
	errForbiddenLocal             = errors.New("teleport: not planet owner")
	errSamePositionLocal          = errors.New("teleport: target equals current")
	errCooldownLocal              = errors.New("teleport: cooldown active")
	errOccupiedLocal              = errors.New("teleport: target slot occupied")
	errBillingNotConfiguredLocal  = errors.New("teleport: billing client not configured")
)

// Координатные лимиты (зеркалят CHECK planets.coords_range из миграции
// 0002_planets_galaxy.sql). Хардкод оправдан тем, что эти числа жёстко
// прибиты к схеме БД — изменить их нельзя без миграции, поэтому
// дублирование в env/config'е добавило бы только риск рассинхрона.
const (
	coordGalaxyMin   = 1
	coordGalaxyMax   = 16
	coordSystemMin   = 1
	coordSystemMax   = 999
	coordPositionMin = 1
	coordPositionMax = 15
)

// errBillingUnavailable — локальный sentinel для 503-ответа. httpx.ErrLocked
// использовать нельзя (это 423 для frozen wallet), а отдельный sentinel в
// httpx/response.go добавлять не хочется ради одного места — DUPLICATE-файл,
// синхронизировать его между четырьмя сервисами.
var errBillingUnavailable = &httpx.Error{
	Status:  http.StatusServiceUnavailable,
	Code:    "billing_unavailable",
	Message: "billing service unavailable",
}

// TeleportConfig — параметры из game-nova config.GameConfig, передаются в
// конструктор. Обособлены в отдельную структуру, чтобы не таскать весь
// GameConfig сквозь tests.
type TeleportConfig struct {
	CostOxsars      int64
	CooldownHours   int
	DurationMinutes int
}

// TeleportHandler — HTTP-handler POST /api/planets/{id}/teleport.
//
// Зависимости:
//   - db: транзактор для атомарной проверки + INSERT events.
//   - billing: client для Spend (план 77).
//   - cfg: TeleportConfig (cost, cooldown, duration).
type TeleportHandler struct {
	db      repo.Exec
	billing *billingclient.Client
	cfg     TeleportConfig
}

// NewTeleportHandler создаёт handler.
//
// Параметр billing может быть nil (при BILLING_URL="" клиент возвращает
// ErrNotConfigured на каждом вызове — тогда handler отдаст 503). Это
// нормальное поведение для dev-окружений без billing-сервиса.
func NewTeleportHandler(db repo.Exec, billing *billingclient.Client, cfg TeleportConfig) *TeleportHandler {
	metrics.RegisterTeleport()
	return &TeleportHandler{db: db, billing: billing, cfg: cfg}
}

// teleportRequest — body POST /api/planets/{id}/teleport.
type teleportRequest struct {
	TargetGalaxy   int `json:"target_galaxy"`
	TargetSystem   int `json:"target_system"`
	TargetPosition int `json:"target_position"`
}

// teleportResponse — body 200 OK.
type teleportResponse struct {
	EventID    string    `json:"event_id"`
	FireAt     time.Time `json:"fire_at"`
	CostOxsars int64     `json:"cost_oxsars"`
}

// Teleport — http.HandlerFunc.
//
// Семантика:
//
//	1. JWT валиден → uid из контекста; иначе 401.
//	2. Idempotency-Key обязателен (R9); иначе 400.
//	3. Decode body, проверить координаты в допустимых диапазонах.
//	4. Tx: SELECT planet (id, user_id, galaxy/system/position, is_moon)
//	   FOR UPDATE → ownership, не та же позиция (no-op).
//	5. SELECT users.last_planet_teleport_at — cooldown.
//	6. SELECT planets WHERE galaxy=? AND system=? AND position=? AND is_moon=is_moon
//	   AND destroyed_at IS NULL → должно быть пусто (occupied → 409).
//	7. Tx2: billing.Spend(...). Списание оксаров. Sentinel-ошибки →
//	   соответствующие HTTP-коды.
//	8. Tx3 (после успешного Spend): event.Insert(KindTeleportPlanet,
//	   fire_at=now+duration, payload).
//	9. 200 OK с {event_id, fire_at, cost_oxsars}.
//
// Транзакции разнесены: проверка → spend → INSERT event. Между шагом 6 и 8
// возможен race (другой игрок зайдёт на slot), который ловится в
// event-handler'е через UNIQUE-constraint planets и Refund — см.
// event/teleport_handler.go.
func (h *TeleportHandler) Teleport(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		if metrics.PlanetTeleportDuration != nil {
			metrics.PlanetTeleportDuration.Observe(time.Since(start).Seconds())
		}
	}()

	uid, ok := auth.UserID(r.Context())
	if !ok || uid == "" {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	planetID := chi.URLParam(r, "id")
	if planetID == "" {
		incTeleportStatus("error")
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "planet id required"))
		return
	}

	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		incTeleportStatus("error")
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "Idempotency-Key header required"))
		return
	}

	var req teleportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		incTeleportStatus("invalid_coords")
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json body"))
		return
	}
	if !validCoord(req.TargetGalaxy, coordGalaxyMin, coordGalaxyMax) ||
		!validCoord(req.TargetSystem, coordSystemMin, coordSystemMax) ||
		!validCoord(req.TargetPosition, coordPositionMin, coordPositionMax) {
		incTeleportStatus("invalid_coords")
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "coordinates out of range"))
		return
	}

	ctx := r.Context()
	cooldown := time.Duration(h.cfg.CooldownHours) * time.Hour
	now := time.Now().UTC()

	// 1) Pre-flight checks в одной транзакции: ownership, не та же позиция,
	// cooldown, occupied slot. Эти проверки read-only — без записи.
	type preCheck struct {
		isMoon bool
	}
	var pre preCheck

	preErr := h.db.InTx(ctx, func(ctx context.Context, tx pgxTx) error {
		var (
			ownerID                       string
			curG, curS, curP              int
			isMoon                        bool
			lastTeleportAt                *time.Time
		)
		err := tx.QueryRow(ctx, `
			SELECT user_id, galaxy, system, position, is_moon
			FROM planets
			WHERE id = $1 AND destroyed_at IS NULL
			FOR UPDATE
		`, planetID).Scan(&ownerID, &curG, &curS, &curP, &isMoon)
		if err != nil {
			return err
		}
		if ownerID != uid {
			return errForbiddenLocal
		}
		if curG == req.TargetGalaxy && curS == req.TargetSystem && curP == req.TargetPosition {
			return errSamePositionLocal
		}

		// Cooldown.
		if err := tx.QueryRow(ctx,
			`SELECT last_planet_teleport_at FROM users WHERE id = $1`, uid,
		).Scan(&lastTeleportAt); err != nil {
			return err
		}
		if lastTeleportAt != nil && now.Sub(*lastTeleportAt) < cooldown {
			return errCooldownLocal
		}

		// Occupied slot. Уникальный constraint в schemе — (galaxy, system,
		// position, is_moon). Проверяем явно, чтобы вернуть 409 до Spend'а
		// (дешевле, чем ловить refund при INSERT-конфликте). Гонка между
		// этим SELECT и INSERT event'а возможна — её ловит event-handler.
		var existsID string
		err = tx.QueryRow(ctx, `
			SELECT id FROM planets
			WHERE galaxy = $1 AND system = $2 AND position = $3 AND is_moon = $4
			  AND destroyed_at IS NULL
			LIMIT 1
		`, req.TargetGalaxy, req.TargetSystem, req.TargetPosition, isMoon).Scan(&existsID)
		if err == nil {
			return errOccupiedLocal
		}
		if !errors.Is(err, pgxErrNoRows) {
			return err
		}
		pre.isMoon = isMoon
		return nil
	})
	if preErr != nil {
		switch {
		case errors.Is(preErr, pgxErrNoRows):
			incTeleportStatus("error")
			httpx.WriteError(w, r, httpx.ErrNotFound)
		case errors.Is(preErr, errForbiddenLocal):
			incTeleportStatus("error")
			httpx.WriteError(w, r, httpx.ErrForbidden)
		case errors.Is(preErr, errSamePositionLocal):
			incTeleportStatus("invalid_coords")
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "target equals current position"))
		case errors.Is(preErr, errCooldownLocal):
			incTeleportStatus("cooldown")
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "teleport cooldown still active"))
		case errors.Is(preErr, errOccupiedLocal):
			incTeleportStatus("occupied")
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, "target slot occupied"))
		default:
			slog.ErrorContext(ctx, "teleport_precheck_failed",
				slog.String("user_id", uid),
				slog.String("planet_id", planetID),
				slog.String("err", preErr.Error()))
			incTeleportStatus("error")
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, "precheck failed"))
		}
		return
	}

	// 2) Spend оксаров через billing.
	userToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	spendErr := errBillingNotConfiguredLocal
	if h.billing != nil {
		spendErr = h.billing.Spend(ctx, billingclient.SpendInput{
			UserToken:      userToken,
			Amount:         h.cfg.CostOxsars,
			Reason:         "planet_teleport",
			RefID:          planetID,
			ToAccount:      "system:teleport",
			IdempotencyKey: idemKey,
		})
	}
	if spendErr != nil {
		switch {
		case errors.Is(spendErr, billingclient.ErrInsufficientOxsar):
			incTeleportStatus("insufficient")
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrPaymentRequired, "insufficient oxsars"))
		case errors.Is(spendErr, billingclient.ErrIdempotencyConflict):
			incTeleportStatus("conflict")
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, "Idempotency-Key reuse with different body"))
		case errors.Is(spendErr, billingclient.ErrBillingUnavailable),
			errors.Is(spendErr, billingclient.ErrNotConfigured),
			errors.Is(spendErr, errBillingNotConfiguredLocal):
			incTeleportStatus("billing_unavailable")
			httpx.WriteError(w, r, errBillingUnavailable)
		default:
			slog.ErrorContext(ctx, "teleport_spend_failed",
				slog.String("user_id", uid),
				slog.String("planet_id", planetID),
				slog.String("err", spendErr.Error()))
			incTeleportStatus("error")
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, "billing spend failed"))
		}
		return
	}

	// 3) INSERT event (KindTeleportPlanet). После Spend'а: если INSERT
	// упадёт — нужно сделать Refund (best-effort). Tx нужен только под
	// саму вставку event'а; идемпотентность (повтор того же ключа уже
	// списанного Spend'а) гарантирует middleware Idempotency-Key.
	fireAt := now.Add(time.Duration(h.cfg.DurationMinutes) * time.Minute)
	payload := event.TeleportPlanetPayload{
		TargetGalaxy:   req.TargetGalaxy,
		TargetSystem:   req.TargetSystem,
		TargetPosition: req.TargetPosition,
		CostOxsars:     h.cfg.CostOxsars,
		IdempotencyKey: idemKey,
	}

	var eventID string
	insertErr := h.db.InTx(ctx, func(ctx context.Context, tx pgxTx) error {
		uidLocal, planetIDLocal := uid, planetID
		var err error
		eventID, err = event.Insert(ctx, tx, event.InsertOpts{
			UserID:   &uidLocal,
			PlanetID: &planetIDLocal,
			Kind:     event.KindTeleportPlanet,
			FireAt:   fireAt,
			Payload:  payload,
		})
		return err
	})
	if insertErr != nil {
		slog.ErrorContext(ctx, "teleport_event_insert_failed_refunding",
			slog.String("user_id", uid),
			slog.String("planet_id", planetID),
			slog.String("err", insertErr.Error()))
		// Best-effort refund. Если billing не отвечает — не блокируем
		// клиента, оставляем висящую транзакцию для админ-recon.
		if h.billing != nil {
			_ = h.billing.Refund(ctx, billingclient.SpendInput{
				UserToken:      userToken,
				Amount:         h.cfg.CostOxsars,
				Reason:         "planet_teleport_event_insert_failed",
				RefID:          planetID,
				ToAccount:      "system:teleport",
				IdempotencyKey: idemKey + ":refund",
			})
		}
		incTeleportStatus("error")
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, "schedule teleport failed"))
		return
	}

	slog.InfoContext(ctx, "teleport_scheduled",
		slog.String("user_id", uid),
		slog.String("planet_id", planetID),
		slog.String("event_id", eventID),
		slog.Int("target_galaxy", req.TargetGalaxy),
		slog.Int("target_system", req.TargetSystem),
		slog.Int("target_position", req.TargetPosition),
		slog.Int64("cost_oxsars", h.cfg.CostOxsars),
		slog.Time("fire_at", fireAt))
	incTeleportStatus("ok")
	httpx.WriteJSON(w, r, http.StatusOK, teleportResponse{
		EventID:    eventID,
		FireAt:     fireAt,
		CostOxsars: h.cfg.CostOxsars,
	})
}

func validCoord(v, min, max int) bool {
	return v >= min && v <= max
}

func incTeleportStatus(status string) {
	if metrics.PlanetTeleportTotal == nil {
		return
	}
	metrics.PlanetTeleportTotal.WithLabelValues(status).Inc()
}
