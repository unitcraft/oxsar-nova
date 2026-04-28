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
	KindDeliveryArtefacts          Kind = 23 // план 65 Ф.2 (D-035): доставка артефактов флотом — груз — записи artefacts_user, не ресурсы
	KindAttackDestroyMoon          Kind = 25 // план 20 Ф.6: атака с целью уничтожить луну
	KindAttackDestroyBuilding      Kind = 26 // план 65 Ф.3 (D-037): атака с целью разрушить здание
	KindAttackAllianceDestroyMoon  Kind = 27 // план 20 Ф.6: ACS-вариант
	KindAttackAllianceDestroyBuilding Kind = 29 // план 65 Ф.4 (D-037): ACS-вариант разрушения здания
	KindAllianceAttackAdditional   Kind = 30 // план 65 Ф.5: служебный referrer для ACS (no-op в nova, см. handler-doc)
	KindTeleportPlanet             Kind = 31 // план 65 Ф.6 (D-032+U-009): телепорт планеты на новые координаты, премиум через оксары
	KindStargateTransport          Kind = 28
	KindStargateJump               Kind = 32 // план 20 Ф.5: мгновенный прыжок флота между лунами с jump_gate
	KindAlienFlyUnknown            Kind = 33 // миссия пришельцев без явной цели (план 15, этап 3)
	KindAlienHolding               Kind = 34 // пришельцы удерживают планету игрока
	KindAlienAttack                Kind = 35 // инопланетяне атакуют планету игрока
	KindAlienHalt                  Kind = 36 // переходное состояние 12–24ч перед HOLDING
	KindAlienGrabCredit            Kind = 37 // отдельный сценарий — кража кредитов (план 15, этап 3)
	KindAlienHoldingAI             Kind = 80 // AI-тик внутри HOLDING (выгрузка ресурсов и др.)
	KindAlienChangeMissionAI       Kind = 81 // смена миссии в полёте (план 15, этап 3)
	KindRepair                     Kind = 50
	KindDisassemble                Kind = 51
	KindArtefactExpire             Kind = 60
	KindArtefactDisappear          Kind = 61
	KindOfficerExpire              Kind = 62
	KindArtefactDelay              Kind = 63
	KindRaidWarning                Kind = 64 // уведомление защитнику за 10 мин до атаки
	KindExpirePlanet               Kind = 65 // soft-удаление временной планеты при expires_at
	KindScoreRecalcAll             Kind = 70 // ежедневный batch-пересчёт очков всех игроков
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
