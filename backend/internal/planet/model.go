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
	ID                 string
	UserID             string
	IsMoon             bool
	Name               string
	Galaxy             int
	System             int
	Position           int
	Diameter           int
	UsedFields         int
	TempMin            int
	TempMax            int
	Metal              float64
	Silicon            float64
	Hydrogen           float64
	LastResUpdate      time.Time
	SolarSatelliteProd int
	BuildFactor        float64
	ResearchFactor     float64
	ProduceFactor      float64
	EnergyFactor       float64
	StorageFactor      float64
}

// Resources — срез ресурсов (без множителей), удобен для API.
type Resources struct {
	Metal    float64 `json:"metal"`
	Silicon  float64 `json:"silicon"`
	Hydrogen float64 `json:"hydrogen"`
}
