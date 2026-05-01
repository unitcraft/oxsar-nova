package fleet

import "oxsar/game-nova/internal/event"

// Семантические предикаты над типом миссии.
//
// Тип миссии хранится в `fleets.mission` как int (= event.Kind того же
// значения). Эти helper'ы избавляют от россыпи `m == 10 || m == 12`
// по транспорту и упрощают понимание кода.
//
// Если добавляется новый kind — нужно решить, относится ли он к одной
// из категорий ниже, и обновить switch-case.

// isValidMission — миссия известна сервису transport.Send.
func isValidMission(m int) bool {
	switch event.Kind(m) {
	case event.KindPosition,
		event.KindTransport,
		event.KindColonize,
		event.KindRecycling,
		event.KindAttackSingle,
		event.KindSpy,
		event.KindAttackAlliance,
		event.KindExpedition,
		event.KindAttackDestroyMoon,
		event.KindAttackAllianceDestroyMoon,
		event.KindHolding:
		return true
	}
	return false
}

// isAggressiveMission — атака/разведка/уничтожение луны.
// Используется для vacation-shield: target в отпуске → отказ.
// SPY (11) включён, потому что разведка — тоже агрессия.
func isAggressiveMission(m int) bool {
	switch event.Kind(m) {
	case event.KindAttackSingle,
		event.KindSpy,
		event.KindAttackAlliance,
		event.KindAttackDestroyMoon,
		event.KindAttackAllianceDestroyMoon:
		return true
	}
	return false
}

// isAttackMission — атака без разведки. Используется для антибашинга
// (SPY не считается атакой).
func isAttackMission(m int) bool {
	switch event.Kind(m) {
	case event.KindAttackSingle,
		event.KindAttackAlliance,
		event.KindAttackDestroyMoon,
		event.KindAttackAllianceDestroyMoon:
		return true
	}
	return false
}

// isMoonDestroyMission — kind=25 или 27.
func isMoonDestroyMission(m int) bool {
	k := event.Kind(m)
	return k == event.KindAttackDestroyMoon || k == event.KindAttackAllianceDestroyMoon
}

// isFleetSlotMission — миссия занимает слот флота (план 20 Ф.2).
// Не считаются: EXPEDITION (отдельные слоты через astro_tech),
// DELIVERY_UNITS (artefact delivery, kind=21).
//
// Note: legacy перечисляет ещё DELIVERY_RESOURCES=22 — у нас не
// используется, но если появится, добавить сюда.
func isFleetSlotMission(m int) bool {
	k := event.Kind(m)
	if k == event.KindExpedition || k == event.KindDeliveryUnits {
		return false
	}
	return true
}

// requiresExistingTarget — миссия летит на существующую планету/луну.
// COLONIZE создаёт планету сама, EXPEDITION летит в неисследованную зону.
func requiresExistingTarget(m int) bool {
	k := event.Kind(m)
	return k != event.KindColonize && k != event.KindExpedition
}

// allowsCarryResources — миссия может везти груз.
// POSITION, TRANSPORT, COLONIZE, HOLDING — да; остальные — нет.
func allowsCarryResources(m int) bool {
	switch event.Kind(m) {
	case event.KindPosition, event.KindTransport, event.KindColonize, event.KindHolding:
		return true
	}
	return false
}
