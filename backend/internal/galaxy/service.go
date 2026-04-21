// Package galaxy отвечает за координаты, выборку системы и свободных
// позиций.
//
// Статус: M3 (пока каркас). Генерация галактики и поиск свободных
// клеток будут подключены с переносом CronCommand-логики из oxsar2.
package galaxy

import "errors"

// Coords — позиция в галактике.
type Coords struct {
	Galaxy   int  `json:"galaxy"`
	System   int  `json:"system"`
	Position int  `json:"position"`
	IsMoon   bool `json:"is_moon"`
}

// Validate проверяет координаты по ограничениям ТЗ (galaxy 1..16,
// system 1..999, position 1..15).
func (c Coords) Validate() error {
	if c.Galaxy < 1 || c.Galaxy > 16 {
		return errors.New("galaxy out of range 1..16")
	}
	if c.System < 1 || c.System > 999 {
		return errors.New("system out of range 1..999")
	}
	if c.Position < 1 || c.Position > 15 {
		return errors.New("position out of range 1..15")
	}
	return nil
}

// Distance возвращает полётную «цену» между двумя координатами в
// OGame-значениях. TODO (M3): формула OGame (межгалактика/интерсистема/
// интерпланета).
func Distance(a, b Coords) int {
	switch {
	case a.Galaxy != b.Galaxy:
		return 20000 * absInt(a.Galaxy-b.Galaxy)
	case a.System != b.System:
		return 2700 + 95*absInt(a.System-b.System)
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
