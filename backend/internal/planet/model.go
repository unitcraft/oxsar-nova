// Package planet отвечает за состояние планет: координаты, ресурсы,
// множители, тик экономики.
//
// Пакет обязан быть единственным, кто модифицирует метал/кристалл/
// дейтерий планеты. Любое списание и начисление проходит через
// Service.ApplyDelta, который также пишет res_log (§18.10 ТЗ).
package planet

import "time"

// Planet — доменная модель. Ресурсы хранятся как float64 для дешёвой
// интерполяции во время тика, но при коммите в БД округляются согласно
// правилам §18.9 ТЗ (добыча — floor в пользу игрока, списание — ceil).
type Planet struct {
	ID                 string    `json:"id"`
	UserID             string    `json:"user_id"`
	IsMoon             bool      `json:"is_moon"`
	Name               string    `json:"name"`
	Galaxy             int       `json:"galaxy"`
	System             int       `json:"system"`
	Position           int       `json:"position"`
	Diameter           int       `json:"diameter"`
	UsedFields         int       `json:"used_fields"`
	TempMin            int       `json:"temp_min"`
	TempMax            int       `json:"temp_max"`
	Metal              float64   `json:"metal"`
	Silicon            float64   `json:"silicon"`
	Hydrogen           float64   `json:"hydrogen"`
	LastResUpdate      time.Time `json:"last_res_update"`
	SolarSatelliteProd int       `json:"solar_satellite_prod"`
	BuildFactor        float64   `json:"build_factor"`
	ResearchFactor     float64   `json:"research_factor"`
	ProduceFactor      float64   `json:"produce_factor"`
	EnergyFactor       float64   `json:"energy_factor"`
	StorageFactor      float64   `json:"storage_factor"`
}

// Resources — срез ресурсов (без множителей), удобен для API.
type Resources struct {
	Metal    float64 `json:"metal"`
	Silicon  float64 `json:"silicon"`
	Hydrogen float64 `json:"hydrogen"`
}
