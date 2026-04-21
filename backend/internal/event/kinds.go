// Package event — единый event-loop (§6 ТЗ).
//
// Все действия с задержкой кладутся в таблицу events с fire_at. Воркер
// каждые EVENT_BATCH_PROCESS_TIME секунд выбирает события с fire_at
// <= now() и state=wait, блокирует FOR UPDATE SKIP LOCKED и вызывает
// соответствующий handler.
//
// Обработчики ОБЯЗАНЫ быть идемпотентными: повторный запуск с тем же
// event_id не должен дублировать эффект (§6.1, §18.3 ТЗ).
package event

// Kind — типы событий (из §5.6 ТЗ, константы EVENT_* в consts.php).
type Kind int

const (
	KindBuildConstruction          Kind = 1
	KindDemolishConstruction       Kind = 2
	KindResearch                   Kind = 3
	KindBuildFleet                 Kind = 4
	KindBuildDefense               Kind = 5
	KindPosition                   Kind = 6
	KindTransport                  Kind = 7
	KindColonize                   Kind = 8
	KindRecycling                  Kind = 9
	KindAttackSingle               Kind = 10
	KindSpy                        Kind = 11
	KindAttackAlliance             Kind = 12
	KindHalt                       Kind = 13
	KindMoonDestruction            Kind = 14
	KindExpedition                 Kind = 15
	KindRocketAttack               Kind = 16
	KindHolding                    Kind = 17
	KindReturn                     Kind = 20
	KindDeliveryUnits              Kind = 21
	KindDeliveryResources          Kind = 22
	KindStargateTransport          Kind = 28
	KindRepair                     Kind = 50
	KindDisassemble                Kind = 51
	KindArtefactExpire             Kind = 60
	KindArtefactDisappear          Kind = 61
	KindOfficerExpire              Kind = 62
	KindArtefactDelay              Kind = 63
	KindBatchProcessIntervalSecond      = 10
)

// State — состояние события (соответствует SQL-enum event_state).
type State string

const (
	StateWait  State = "wait"
	StateStart State = "start"
	StateOK    State = "ok"
	StateError State = "error"
)
