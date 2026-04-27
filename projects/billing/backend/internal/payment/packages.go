package payment

import "errors"

var ErrPackageNotFound = errors.New("payment: package not found")

// Packages — каталог пакетов кредитов. В MVP захардкожено;
// при надобности переехать в БД (`packages` table) — план 38 след. итерация.
//
// Конвертация: 1 RUB = 100 копеек = 100 OXC.
// Бонусные кредиты — за объём.
var Packages = []Package{
	{ID: "pack_500", Title: "500 кредитов", AmountKop: 50000, Credits: 50000},
	{ID: "pack_2000", Title: "2000 кредитов + 200 бонус", AmountKop: 200000, Credits: 200000, Bonus: 20000, IsBest: true},
	{ID: "pack_5000", Title: "5000 кредитов + 1000 бонус", AmountKop: 500000, Credits: 500000, Bonus: 100000},
	{ID: "pack_10000", Title: "10000 кредитов + 3000 бонус", AmountKop: 1000000, Credits: 1000000, Bonus: 300000},
}

// FindPackage возвращает пакет по id (ошибка если не найден).
func FindPackage(id string) (Package, error) {
	for _, p := range Packages {
		if p.ID == id {
			return p, nil
		}
	}
	return Package{}, ErrPackageNotFound
}

// TotalCredits сумма Credits + Bonus (что юзер реально получит).
func (p Package) TotalCredits() int64 {
	return p.Credits + p.Bonus
}
