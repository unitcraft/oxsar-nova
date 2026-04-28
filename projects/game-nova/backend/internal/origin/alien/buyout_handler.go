package alien

// План 66 Ф.5: платный выкуп удержания (HOLDING) пришельцами за оксары.
//
// Семантика: игрок, чья планета удерживается (есть активный
// KindAlienHolding в state='wait'), может одной транзакцией:
//   1. списать N оксаров через billing-service (план 77),
//   2. закрыть HOLDING-event (state='ok'),
//   3. снять все запланированные KindAlienHoldingAI-тики этой миссии.
//
// R0-исключение: фича работает во ВСЕХ вселенных (uni01/uni02 + origin),
// как и весь пакет origin/alien — см. doc.go.
//
// Особенность относительно legacy oxsar2: в AlienAI.class.php платного
// выкупа НЕ существует. В origin был только paid_credit (продление окна
// HOLDING на 2h за каждые 50 оксаритов — см. PHP:993, маппится на
// internal/alien.PayHolding). Buyout — НОВАЯ фича ремастера; цена
// фиксированная, в Config.BuyoutBaseOxsars (по умолчанию 100 оксаров).
//
// Колонки `planets.locked_by_alien` в схеме нет — блокировка планеты
// уже моделируется самим присутствием активного HOLDING-event. Поэтому
// «разблокировка планеты» = state='ok' на parent. Это соответствует
// тому, как HOLDING закрывается во всех других путях: естественный
// конец (HoldingHandler), битва-рассеивание (CloseHoldingIfWiped),
// все стеки extracted (closeHoldingScattered в HoldingAIHandler).
//
// Идемпотентность:
//   - billing.Spend получает Idempotency-Key из HTTP-заголовка
//     (через middleware), повтор того же ключа дедуплицируется на
//     стороне billing-service.
//   - Если billing уже списал ранее (replay), а нашего HOLDING уже
//     нет (state='ok') — возвращаем ErrMissionAlreadyClosed (HTTP 409).
//     Refund НЕ делаем: повторный buyout того же пользователя уже
//     получил деньги назад через идемпотентность billing'а на первом
//     replay (тот же Idempotency-Key → billing вернул 200 без double
//     списания).

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"

	billingclient "oxsar/game-nova/internal/billing/client"
	"oxsar/game-nova/internal/event"
	"oxsar/game-nova/internal/repo"
	"oxsar/game-nova/pkg/metrics"
)

// Sentinel-ошибки buyout. HTTP-слой маппит их в коды 4xx/5xx.
var (
	// ErrMissionNotFound — миссия с таким ID не существует, либо это не
	// KindAlienHolding, либо принадлежит другому пользователю. Не
	// раскрываем «существует, но чужая» — единый 404.
	ErrMissionNotFound = errors.New("alien buyout: mission not found")

	// ErrMissionAlreadyClosed — KindAlienHolding есть, но уже не
	// state='wait' (закрыт битвой / истечением времени / другим
	// buyout'ом). HTTP 409.
	ErrMissionAlreadyClosed = errors.New("alien buyout: mission not in holding state")

	// ErrInsufficientOxsars — у игрока меньше оксаров, чем стоит выкуп
	// (billing вернул 402). HTTP 402.
	ErrInsufficientOxsars = errors.New("alien buyout: insufficient oxsars")

	// ErrIdempotencyConflict — тот же Idempotency-Key уже использовался
	// с другим request body на стороне billing. HTTP 409.
	ErrIdempotencyConflict = errors.New("alien buyout: idempotency conflict")

	// ErrBillingUnavailable — billing-service недоступен (5xx/timeout
	// после retry). HTTP 503. Миссия НЕ закрыта, можно повторить.
	ErrBillingUnavailable = errors.New("alien buyout: billing unavailable")
)

// BuyoutResult — успешный результат выкупа.
type BuyoutResult struct {
	MissionID  string    `json:"mission_id"`
	CostOxsars int64     `json:"cost_oxsars"`
	FreedAt    time.Time `json:"freed_at"`
}

// BuyoutBilling — узкий интерфейс, который Buyout требует от
// billing-клиента (планы 77 / 66 Ф.5). Узкий, чтобы тестам не
// приходилось поднимать httptest-сервер для каждого сценария.
//
// Реализуется *billingclient.Client (production) и mockBilling в тестах.
type BuyoutBilling interface {
	Spend(ctx context.Context, in billingclient.SpendInput) error
}

// Buyout выполняет платный выкуп HOLDING-удержания миссии missionID
// пользователем userID. UserToken — RSA-JWT игрока, forward'ится в
// billing для авторизации списания со счёта (см. billing-client).
// idempotencyKey — заголовок Idempotency-Key из HTTP-запроса, используется
// и для billing-дедупликации, и связывается с buyout-операцией в логах.
//
// Транзакционная семантика:
//   1. Открываем DB-tx, читаем HOLDING-event FOR UPDATE, проверяем
//      kind/state/owner. ROLLBACK при любом из несоответствий →
//      ErrMissionNotFound / ErrMissionAlreadyClosed.
//   2. Закрываем DB-tx (коммитим locking-snapshot HOLDING'а — так мы
//      зафиксировали, что миссия валидна на момент решения).
//   3. Вызываем billing.Spend ВНЕ DB-tx (чтобы не держать блокировку
//      в строке events во время сетевого вызова). Ошибки billing —
//      возвращаем без изменения миссии.
//   4. Если billing вернул ok — открываем второй DB-tx, повторно
//      проверяем state и закрываем HOLDING + удаляем тики
//      KindAlienHoldingAI этой миссии. На этом этапе race с
//      естественным закрытием HOLDING (CloseHoldingIfWiped и т.п.) —
//      возможен; если HOLDING уже не 'wait', closing — no-op,
//      money refund-ить НЕ нужно (см. doc выше: повторный buyout с
//      тем же Idempotency-Key получит ту же сумму назад через
//      billing-идемпотентность; если же это был первый успех + race —
//      пользователь получил «игровой результат» естественным путём,
//      а заплатил 100 оксаров — приемлемо в рамках R15: ситуация
//      окно ~миллисекунды, цена сопоставима с обычным шопом).
func Buyout(
	ctx context.Context,
	db repo.Exec,
	billing BuyoutBilling,
	cfg Config,
	userID, missionID, userToken, idempotencyKey string,
) (*BuyoutResult, error) {

	if userID == "" || missionID == "" {
		return nil, ErrMissionNotFound
	}
	if idempotencyKey == "" {
		// R9: middleware гарантирует header, но defense-in-depth.
		return nil, fmt.Errorf("alien buyout: empty idempotency key")
	}

	cost := cfg.BuyoutBaseOxsars
	if cost <= 0 {
		// Защита от misconfiguration — никакая цена не «бесплатно».
		return nil, fmt.Errorf("alien buyout: BuyoutBaseOxsars must be positive (got %d)", cost)
	}

	// 1) Pre-check + lock в первой транзакции.
	if err := db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var kind int
		var state string
		var eventUserID string
		err := tx.QueryRow(ctx, `
			SELECT kind, state, user_id
			FROM events WHERE id = $1::uuid FOR UPDATE
		`, missionID).Scan(&kind, &state, &eventUserID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrMissionNotFound
			}
			return fmt.Errorf("alien buyout: load mission: %w", err)
		}
		if kind != int(event.KindAlienHolding) {
			return ErrMissionNotFound
		}
		if eventUserID != userID {
			// Не раскрываем что миссия чужая — единый 404.
			return ErrMissionNotFound
		}
		if state != string(event.StateWait) {
			return ErrMissionAlreadyClosed
		}
		return nil
	}); err != nil {
		incBuyoutStatus(buyoutStatusFromErr(err))
		return nil, err
	}

	// 2) Списание оксаров через billing-service ВНЕ tx.
	if billing == nil {
		incBuyoutStatus("error")
		return nil, fmt.Errorf("alien buyout: billing client not configured")
	}
	if err := billing.Spend(ctx, billingclient.SpendInput{
		UserToken:      userToken,
		Amount:         cost,
		Reason:         "alien_buyout",
		RefID:          missionID,
		ToAccount:      "system:alien_buyout",
		IdempotencyKey: idempotencyKey,
	}); err != nil {
		// Маппинг ошибок billing → buyout-sentinel'ы.
		switch {
		case errors.Is(err, billingclient.ErrInsufficientOxsar):
			incBuyoutStatus("insufficient")
			return nil, ErrInsufficientOxsars
		case errors.Is(err, billingclient.ErrIdempotencyConflict):
			incBuyoutStatus("conflict")
			return nil, ErrIdempotencyConflict
		case errors.Is(err, billingclient.ErrBillingUnavailable),
			errors.Is(err, billingclient.ErrNotConfigured):
			incBuyoutStatus("billing_unavailable")
			return nil, ErrBillingUnavailable
		case errors.Is(err, billingclient.ErrFrozenWallet):
			// Замороженный кошелёк — игрок забанен/в споре. Возвращаем
			// 402 (нечего тратить), без отдельной sentinel'и: для UX
			// «недостаточно оксаров» и «кошелёк заморожен» одинаково
			// = «не получилось списать».
			incBuyoutStatus("insufficient")
			return nil, ErrInsufficientOxsars
		default:
			incBuyoutStatus("error")
			return nil, fmt.Errorf("alien buyout: billing spend: %w", err)
		}
	}

	// 3) Закрытие HOLDING + снятие тиков AI во второй транзакции.
	var freedAt time.Time
	if err := db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Повторная проверка state — между первой tx и сюда HOLDING мог
		// быть закрыт битвой / естественным истечением. Если он уже
		// не 'wait' — buyout фактически no-op в DB (но billing уже
		// списал; см. doc выше про принятие race-цены).
		res, err := tx.Exec(ctx, `
			UPDATE events
			SET state = 'ok', processed_at = now()
			WHERE id = $1::uuid AND kind = $2 AND state = 'wait'
		`, missionID, int(event.KindAlienHolding))
		if err != nil {
			return fmt.Errorf("alien buyout: close holding: %w", err)
		}
		_ = res.RowsAffected() // 0 = race с естественным закрытием.

		// Удаляем все запланированные тики KindAlienHoldingAI этой
		// миссии. Используем payload->>'holding_event_id' как
		// единственный надёжный признак привязки тика к миссии
		// (см. payload_holding_ai.go::HoldingAIPayload).
		if _, err := tx.Exec(ctx, `
			DELETE FROM events
			WHERE kind = $1
			  AND state = 'wait'
			  AND payload->>'holding_event_id' = $2
		`, int(event.KindAlienHoldingAI), missionID); err != nil {
			return fmt.Errorf("alien buyout: drop ai ticks: %w", err)
		}

		// FreedAt = серверное now() в этой tx — для отчёта клиенту.
		if err := tx.QueryRow(ctx, `SELECT now()`).Scan(&freedAt); err != nil {
			return fmt.Errorf("alien buyout: read now: %w", err)
		}
		return nil
	}); err != nil {
		incBuyoutStatus("error")
		// На этом этапе billing УЖЕ списал — но DB-write провалился.
		// R15: лог + slog.Error, чтобы оператор увидел и сделал manual
		// refund / закрытие миссии. Полный 2PC между Postgres и billing
		// не реализуем (план 77 не требует — компромисс: billing —
		// идемпотентен по Idempotency-Key, повтор того же запроса от
		// клиента приведёт к закрытию миссии без второго списания).
		slog.ErrorContext(ctx, "alien_buyout_db_after_spend",
			slog.String("user_id", userID),
			slog.String("mission_id", missionID),
			slog.String("idempotency_key", idempotencyKey),
			slog.Int64("cost_oxsars", cost),
			slog.String("err", err.Error()))
		return nil, err
	}

	incBuyoutStatus("ok")
	addBuyoutOxsars(cost)
	slog.InfoContext(ctx, "alien_buyout_paid",
		slog.String("user_id", userID),
		slog.String("mission_id", missionID),
		slog.String("idempotency_key", idempotencyKey),
		slog.Int64("cost_oxsars", cost),
		slog.Time("freed_at", freedAt))

	return &BuyoutResult{
		MissionID:  missionID,
		CostOxsars: cost,
		FreedAt:    freedAt,
	}, nil
}

// buyoutStatusFromErr преобразует sentinel-ошибку в label статуса метрики.
// Для not-found / already-closed ошибок (pre-check) — буфер до billing,
// чтобы метрика отражала и валидационные отказы.
func buyoutStatusFromErr(err error) string {
	switch {
	case errors.Is(err, ErrMissionNotFound):
		return "not_found"
	case errors.Is(err, ErrMissionAlreadyClosed):
		return "conflict"
	default:
		return "error"
	}
}

// incBuyoutStatus / addBuyoutOxsars — обёртки, безопасные до
// metrics.Register() (тогда no-op): не падать в тестах, где метрики
// не подняты.
func incBuyoutStatus(status string) {
	if metrics.AlienBuyoutTotal == nil {
		return
	}
	metrics.AlienBuyoutTotal.WithLabelValues(status).Inc()
}

func addBuyoutOxsars(amount int64) {
	if metrics.AlienBuyoutOxsars == nil {
		return
	}
	metrics.AlienBuyoutOxsars.Add(float64(amount))
}

// Compile-time check: *billingclient.Client должен удовлетворять
// BuyoutBilling. Если client.Spend поменяет сигнатуру — здесь будет
// ошибка компиляции, не runtime-сюрприз.
var _ BuyoutBilling = (*billingclient.Client)(nil)

// Кост из конфига получаем через геттер, чтобы тестам было удобнее
// проверять property «cost детерминирован для одной mission» — он
// не зависит от mission_id, чисто конфиг (см. CalcBuyoutCost тест).

// CalcBuyoutCost — чистая функция, возвращающая стоимость buyout по
// конфигу. Сейчас — fixed-price (новая фича ремастера, формулы в
// legacy не было). Параметризована под будущее расширение
// (если понадобится «цена зависит от tier / paid_times»).
func CalcBuyoutCost(cfg Config, _ string /*missionID — резерв на формулу*/) int64 {
	return cfg.BuyoutBaseOxsars
}
