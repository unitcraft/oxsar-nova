// Package formula — безопасный typed-DSL для парсинга и вычисления
// балансовых формул oxsar2 (см. §5.2.1 ТЗ).
//
// Зачем: в legacy na_construction.prod_metal и т.п. хранятся строки
// вроде
//     floor(30 * {level} * pow(1.1+{tech=23}*0.0006, {level}))
// и в PHP они вычислялись через eval(). В nova мы хотим
//   1) сохранить формулы как данные (можно тюнить без редеплоя Go);
//   2) не тянуть eval и goja — они overkill и attack surface;
//   3) дать геймдизу возможность безопасно править configs/*.yml.
//
// Пакет предоставляет:
//   formula.Parse(src) (*Expr, error)   — разбор в AST
//   expr.Eval(Context) (float64, error) — вычисление
//
// Поддерживаемый синтаксис:
//   числа:      42, 3.14, 0.0006, -1
//   переменные: {level} {basic} {temp} {tech=N}
//   операторы:  + - * /  и  унарный -
//   скобки:     ( ... )
//   функции:    floor, ceil, round, pow, min, max, sqrt, abs
//
// Нет: присваиваний, переменных-пользователей, циклов, сравнений,
// строк, if. Сознательно — чтобы DSL оставался тривиальным и
// безопасным.
package formula

// Context — значения переменных для одного вычисления.
//
// Level      — уровень здания/исследования/корабля.
// Basic      — значение соответствующего basic_* поля из
//              na_construction (базовая стоимость уровня 1).
// Temperature — средняя температура планеты (hydrogen_lab).
// Tech       — карта buildingid -> уровень из research2user. Ключи
//              — числа вроде 18 (energy_tech), 23 (laser_tech).
//              Отсутствующий ключ = 0.
type Context struct {
	Level       int
	Basic       int64
	Temperature int
	Tech        map[int]int
}

// techLevel возвращает уровень технологии, 0 если не задан.
func (c Context) techLevel(id int) int {
	if c.Tech == nil {
		return 0
	}
	return c.Tech[id]
}
