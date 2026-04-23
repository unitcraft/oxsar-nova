package payment

// CreditPackage описывает один вариант пополнения кредитов.
type CreditPackage struct {
	Key          string
	Label        string
	Credits      int
	BonusCredits int
	PriceKop     int // в копейках: 4900 = 49.00 руб
}

func (p CreditPackage) TotalCredits() int  { return p.Credits + p.BonusCredits }
func (p CreditPackage) PriceRub() float64  { return float64(p.PriceKop) / 100 }

// Packages — доступные пакеты кредитов, в порядке возрастания цены.
var Packages = []CreditPackage{
	{Key: "trial",   Label: "Пробный",      Credits: 400,   BonusCredits: 0,    PriceKop: 4900},
	{Key: "starter", Label: "Стартовый",    Credits: 1000,  BonusCredits: 0,    PriceKop: 10000},
	{Key: "medium",  Label: "Средний",      Credits: 3000,  BonusCredits: 200,  PriceKop: 25000},
	{Key: "big",     Label: "Большой",      Credits: 7000,  BonusCredits: 500,  PriceKop: 50000},
	{Key: "max",     Label: "Максимальный", Credits: 15000, BonusCredits: 2000, PriceKop: 100000},
}

// PackageByKey ищет пакет по ключу.
func PackageByKey(key string) (CreditPackage, bool) {
	for _, p := range Packages {
		if p.Key == key {
			return p, true
		}
	}
	return CreditPackage{}, false
}
