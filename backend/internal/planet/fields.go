// Package planet — расчёт лимита полей планеты/луны (план 23).
//
// Формула перенесена из legacy Planet.class.php:getMaxFields (строки 682-708).
// Константа PLANET_FIELD_ADDITION=10 из na_config (na_options в нашей БД).
// TEMP_MOON_SIZE_MAX=2500 из consts.php:602.

package planet

import "math"

// FieldConsts — настройки для MaxFields. Вынесены параметрами, чтобы
// можно было тестировать без global state и при желании поменять
// через config.
type FieldConsts struct {
	// PlanetFieldAddition — статическая надбавка к max_fields планеты
	// (legacy PLANET_FIELD_ADDITION из na_config, по умолчанию 10).
	PlanetFieldAddition int
	// TempMoonSizeMax — максимальный диаметр «маленькой» луны, для
	// которой применяется ×2-бонус к полям (legacy TEMP_MOON_SIZE_MAX=2500).
	TempMoonSizeMax int
	// TerraFormerFieldsPerLevel — сколько полей даёт один уровень
	// terra_former (legacy = 5).
	TerraFormerFieldsPerLevel int
	// MoonLabFieldsPerLevel — сколько полей даёт один уровень
	// moon_lab на луне (legacy = 5).
	MoonLabFieldsPerLevel int
}

// DefaultFieldConsts — дефолтные значения из legacy. Менять только
// если явно пересматриваем механику.
var DefaultFieldConsts = FieldConsts{
	PlanetFieldAddition:       10,
	TempMoonSizeMax:           2500,
	TerraFormerFieldsPerLevel: 5,
	MoonLabFieldsPerLevel:     5,
}

// MaxFields возвращает максимальное количество полей для застройки
// на планете/луне. Формула legacy:
//
//	base = round((diameter / 1000)^2)
//
// Planet:
//
//	max = base + terra_former × 5 + PLANET_FIELD_ADDITION
//
// Moon:
//
//	if diameter <= 2500: base *= 2
//	max = base + moon_lab × 5
//
// buildings — карта unit_id → level (может быть nil, тогда бонусы = 0).
func MaxFields(p *Planet, buildings map[int]int, c FieldConsts) int {
	base := round(math.Pow(float64(p.Diameter)/1000.0, 2))
	if p.IsMoon {
		if p.Diameter <= c.TempMoonSizeMax {
			base *= 2
		}
		moonLab := buildings[350] // UNIT_MOON_LAB
		return base + moonLab*c.MoonLabFieldsPerLevel
	}
	terraFormer := buildings[58] // UNIT_TERRA_FORMER
	return base + terraFormer*c.TerraFormerFieldsPerLevel + c.PlanetFieldAddition
}

// round — PHP-стиль округления (banker-round отличается от Go math.Round
// только на .5, но в formula мы не попадаем на .5 при integer diameter).
func round(x float64) int {
	return int(math.Round(x))
}
