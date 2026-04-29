// Package galaxy отвечает за координаты, выборку системы и свободных
// позиций.
//
// План 72.1 часть 12: лимиты галактика/система — параметры вселенной
// (`Config.Game.NumGalaxies` / `NumSystems`), читаются из
// `configs/universes.yaml`. Раньше были hardcoded 1..16 / 1..999 — это
// был баг (несоответствие YAML и реального runtime). Position 1..15
// остаётся фиксированным (общий лимит игрового мира OGame-classic).
package galaxy

import "fmt"

// Coords — позиция в галактике.
type Coords struct {
	Galaxy   int  `json:"galaxy"`
	System   int  `json:"system"`
	Position int  `json:"position"`
	IsMoon   bool `json:"is_moon"`
}

// PositionMax — максимальная позиция в системе (OGame-classic, не зависит
// от вселенной). Если когда-либо потребуется per-universe — вынести в
// configs/universes.yaml аналогично NumGalaxies/NumSystems.
const PositionMax = 15

// Validate проверяет координаты по лимитам вселенной (numGalaxies,
// numSystems) и фиксированному PositionMax.
func (c Coords) Validate(numGalaxies, numSystems int) error {
	if numGalaxies <= 0 || numSystems <= 0 {
		return fmt.Errorf("galaxy.Validate: invalid limits numGalaxies=%d numSystems=%d", numGalaxies, numSystems)
	}
	if c.Galaxy < 1 || c.Galaxy > numGalaxies {
		return fmt.Errorf("galaxy out of range 1..%d", numGalaxies)
	}
	if c.System < 1 || c.System > numSystems {
		return fmt.Errorf("system out of range 1..%d", numSystems)
	}
	if c.Position < 1 || c.Position > PositionMax {
		return fmt.Errorf("position out of range 1..%d", PositionMax)
	}
	return nil
}

// Distance возвращает полётную «цену» между двумя координатами
// (OGame-classic с кольцевой топологией систем — план 72.1 часть 12).
//
// Кольцевая топология: системы образуют замкнутое кольцо, поэтому система 1
// и система numSystems — соседи. Расстояние между системами:
// `min(|s1-s2|, numSystems - |s1-s2|)`. Раньше брался `|s1-s2|` (линейный),
// что давало некорректное время полёта между крайними системами.
func Distance(a, b Coords, numSystems int) int {
	switch {
	case a.Galaxy != b.Galaxy:
		return 20000 * absInt(a.Galaxy-b.Galaxy)
	case a.System != b.System:
		diff := absInt(a.System - b.System)
		if numSystems > 0 && diff > numSystems/2 {
			diff = numSystems - diff
		}
		return 2700 + 95*diff
	case a.Position != b.Position:
		return 1000 + 5*absInt(a.Position-b.Position)
	default:
		return 5
	}
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
