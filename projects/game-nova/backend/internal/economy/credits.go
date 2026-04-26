package economy

// Балансовые константы кредитной экономики (план 11).
// 1000 кредитов = 100 рублей (1 кредит = 0.1 руб).
const (
	CreditDailyLogin    int64 = 10  // ежедневный бонус за вход
	CreditBattleWinMin  int64 = 5   // минимум за победу в бою
	CreditBattleWinMax  int64 = 50  // максимум за победу в бою
	CreditAchievement   int64 = 50  // за разблокировку достижения
	CreditFirstColony   int64 = 100 // за первую колонизацию
	CreditFirstExped    int64 = 50  // за первую экспедицию
)

// BattleWinCredits возвращает количество кредитов за победу в бою
// пропорционально мощи уничтоженного флота противника.
// defPower — сумма очков боевой силы всех уничтоженных защитников.
func BattleWinCredits(defPower float64) int64 {
	if defPower <= 0 {
		return 0
	}
	// Линейная шкала: 100k → 5 кр, 1M → 50 кр.
	credits := int64(defPower / 20000)
	if credits < CreditBattleWinMin {
		credits = CreditBattleWinMin
	}
	if credits > CreditBattleWinMax {
		credits = CreditBattleWinMax
	}
	return credits
}
