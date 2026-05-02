// Package fleet реализует перемещение флотов.
//
// Каждая миссия — отдельная реализация интерфейса Mission. Базовый
// контракт: валидация -> стоимость -> создание события прибытия.
// Фактический расчёт боя, добычи, ресайкла и т.п. делегируется на
// handler-ы event-loop (domain/event), чтобы fleet-пакет не зависел
// от battle, recycling и других доменов.
package fleet

import (
	"time"
)

// Dispatch — запрос на отправку флота от UI.
type Dispatch struct {
	UserID       string
	SrcPlanetID  string
	DstGalaxy    int
	DstSystem    int
	DstPosition  int
	DstIsMoon    bool
	Mission      int
	Ships        map[int]int64 // unit_id -> count
	SpeedPercent int
	CarryMetal   int64
	CarrySilicon int64
	CarryHydro   int64
	HoldSeconds  int
}

// Fleet — флот в полёте (проекция строки БД).
type Fleet struct {
	ID           string     `json:"id"`
	OwnerUserID  string     `json:"owner_user_id"`
	SrcPlanetID  string     `json:"src_planet_id"`
	DstGalaxy    int        `json:"dst_galaxy"`
	DstSystem    int        `json:"dst_system"`
	DstPosition  int        `json:"dst_position"`
	DstIsMoon    bool       `json:"dst_is_moon"`
	Mission      int        `json:"mission"`
	State        string     `json:"state"`
	DepartAt     time.Time  `json:"depart_at"`
	ArriveAt     time.Time  `json:"arrive_at"`
	ReturnAt     *time.Time `json:"return_at"`
	HoldSeconds  int        `json:"hold_seconds"`
	Carry        Resources  `json:"carry"`
	SpeedPercent int        `json:"speed_percent"`
	Ships        map[int]int64 `json:"ships"`
	// План 72.1.48: ACS-formation. Для ACS-флотов содержит UUID
	// acs_groups записи; иначе пусто.
	ACSGroupID   *string    `json:"acs_group_id,omitempty"`
	// План 72.1.48 (доделка): rate-limit на load/unload + резерв H.
	ControlTimes     int   `json:"control_times,omitempty"`
	MaxControlTimes  int   `json:"max_control_times,omitempty"`
	BackConsumption  int64 `json:"back_consumption,omitempty"`
}

type Resources struct {
	Metal    int64 `json:"metal"`
	Silicon  int64 `json:"silicon"`
	Hydrogen int64 `json:"hydrogen"`
}
