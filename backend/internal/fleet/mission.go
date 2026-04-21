// Package fleet реализует перемещение флотов.
//
// Каждая миссия — отдельная реализация интерфейса Mission. Базовый
// контракт: валидация -> стоимость -> создание события прибытия.
// Фактический расчёт боя, добычи, ресайкла и т.п. делегируется на
// handler-ы event-loop (domain/event), чтобы fleet-пакет не зависел
// от battle, recycling и других доменов.
package fleet

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

// Mission — поведение конкретной миссии флота (транспорт, атака, шпион…).
// Реализации живут в fleet/missions_*.go.
type Mission interface {
	Kind() int // соответствует event.Kind / EVENT_* из consts.php
	Validate(ctx context.Context, in Dispatch) error
	Cost(in Dispatch) (metal, silicon, hydrogen int64)
	Duration(in Dispatch) time.Duration
	OnArrive(ctx context.Context, tx pgx.Tx, f Fleet) error
	OnReturn(ctx context.Context, tx pgx.Tx, f Fleet) error
}

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
	ID           string
	OwnerUserID  string
	SrcPlanetID  string
	DstGalaxy    int
	DstSystem    int
	DstPosition  int
	DstIsMoon    bool
	Mission      int
	State        string
	DepartAt     time.Time
	ArriveAt     time.Time
	ReturnAt     *time.Time
	HoldSeconds  int
	Carry        Resources
	SpeedPercent int
	Ships        map[int]int64
}

type Resources struct {
	Metal    int64
	Silicon  int64
	Hydrogen int64
}
