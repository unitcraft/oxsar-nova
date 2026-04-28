package main

import (
	"context"
	"database/sql"
	"fmt"
)

// Construction — одна запись из na_construction (origin balance).
// Соответствует структуре docker-mysql-1.
//
// Имена ResourceFormulas — DSL-строки (varbinary), могут быть пустыми.
type Construction struct {
	BuildingID   int
	Name         string // CAPSCASE, e.g. METALMINE, HYDROGEN_LAB
	Race         int
	Mode         int    // 1=building, 2=research, 3=ship, 4=defense, ...
	Test         int
	Front        int
	Ballistics   int
	Masking      int

	BasicMetal    float64
	BasicSilicon  float64
	BasicHydrogen float64
	BasicEnergy   float64
	BasicCredit   int64
	BasicPoints   int64

	ProdMetal    string
	ProdSilicon  string
	ProdHydrogen string
	ProdEnergy   string

	ConsMetal    string
	ConsSilicon  string
	ConsHydrogen string
	ConsEnergy   string

	ChargeMetal    string
	ChargeSilicon  string
	ChargeHydrogen string
	ChargeEnergy   string
	ChargeCredit   string
	ChargePoints   string

	Special      string
	Demolish     float64
	DisplayOrder int
}

// loadConstructions читает na_construction.
func loadConstructions(ctx context.Context, db *sql.DB) ([]Construction, error) {
	const q = `SELECT
		buildingid, name, race, mode, test, front, ballistics, masking,
		basic_metal, basic_silicon, basic_hydrogen, basic_energy, basic_credit, basic_points,
		prod_metal, prod_silicon, prod_hydrogen, prod_energy,
		cons_metal, cons_silicon, cons_hydrogen, cons_energy,
		charge_metal, charge_silicon, charge_hydrogen, charge_energy, charge_credit, charge_points,
		special, demolish, display_order
	FROM na_construction ORDER BY mode, display_order, buildingid`

	rows, err := db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Construction
	for rows.Next() {
		var c Construction
		if err := rows.Scan(
			&c.BuildingID, &c.Name, &c.Race, &c.Mode, &c.Test, &c.Front, &c.Ballistics, &c.Masking,
			&c.BasicMetal, &c.BasicSilicon, &c.BasicHydrogen, &c.BasicEnergy, &c.BasicCredit, &c.BasicPoints,
			&c.ProdMetal, &c.ProdSilicon, &c.ProdHydrogen, &c.ProdEnergy,
			&c.ConsMetal, &c.ConsSilicon, &c.ConsHydrogen, &c.ConsEnergy,
			&c.ChargeMetal, &c.ChargeSilicon, &c.ChargeHydrogen, &c.ChargeEnergy, &c.ChargeCredit, &c.ChargePoints,
			&c.Special, &c.Demolish, &c.DisplayOrder,
		); err != nil {
			return nil, fmt.Errorf("scan na_construction: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// ShipDatasheet — одна запись из na_ship_datasheet (бой-параметры).
type ShipDatasheet struct {
	UnitID            int
	Capicity          int64 // sic — origin schema typo
	Speed             int
	Consume           int
	Attack            int
	Shield            int
	Front             int
	Ballistics        int
	Masking           int
	AttackerAttack    int
	AttackerShield    int
	AttackerFront     int
	AttackerBallistics int
	AttackerMasking   int
}

func loadShipDatasheet(ctx context.Context, db *sql.DB) ([]ShipDatasheet, error) {
	const q = `SELECT
		unitid, capicity, speed, consume, attack, shield, front, ballistics, masking,
		attacker_attack, attacker_shield, attacker_front, attacker_ballistics, attacker_masking
	FROM na_ship_datasheet ORDER BY unitid`
	rows, err := db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ShipDatasheet
	for rows.Next() {
		var s ShipDatasheet
		if err := rows.Scan(
			&s.UnitID, &s.Capicity, &s.Speed, &s.Consume, &s.Attack, &s.Shield, &s.Front, &s.Ballistics, &s.Masking,
			&s.AttackerAttack, &s.AttackerShield, &s.AttackerFront, &s.AttackerBallistics, &s.AttackerMasking,
		); err != nil {
			return nil, fmt.Errorf("scan na_ship_datasheet: %w", err)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// RapidfireEntry — одна запись из na_rapidfire.
type RapidfireEntry struct {
	UnitID int
	Target int
	Value  int
}

// loadRapidfire читает na_rapidfire. Если в основном DSN таблица
// пустая (как у docker-mysql-1 — origin использует ext-перекрытия),
// caller должен попробовать fallback DSN на oxsar2-mysql-1 через
// loadRapidfireFromDSN.
func loadRapidfire(ctx context.Context, db *sql.DB) ([]RapidfireEntry, error) {
	const q = `SELECT unitid, target, value FROM na_rapidfire ORDER BY unitid, target`
	rows, err := db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []RapidfireEntry
	for rows.Next() {
		var r RapidfireEntry
		if err := rows.Scan(&r.UnitID, &r.Target, &r.Value); err != nil {
			return nil, fmt.Errorf("scan na_rapidfire: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
