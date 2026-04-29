// Package miner — система Шахтёр-уровня (план 72.1 ч.17).
//
// Не зависит от профессии «miner»: of_points начисляется ВСЕМ игрокам
// при добыче ресурсов (legacy `Functions.inc.php:1633` —
// `updateMinerPoints` вызывается из event-handler'а после завершения
// production-tick).
//
// Уровень `of_level` повышается автоматически по мере накопления
// `of_points`. Порог следующего уровня — `NeedPoints(level)`.
// При level-up игрок получает кредит-награду из `RewardCredits[level]`.
//
// Эффект on-game пока только в UI (отображение «Шахтёр N» на
// MainScreen). В legacy-PHP в production tick of_level **не** влияет
// на множители добычи.
package miner

import "math"

// NeedPoints — порог `of_points` для следующего level-up.
// Формула из legacy `Functions.inc.php:1640,1664`:
//
//	level<1 → 100, иначе round(pow(1.5, level-1) * 200)
//
// Прогрессия для первых уровней:
//
//	level 0 → 100, 1 → 200, 2 → 300, 3 → 450, 4 → 675,
//	level 5 → 1013, 6 → 1519, 7 → 2278, ...
func NeedPoints(level int) int64 {
	if level < 1 {
		return 100
	}
	return int64(math.Round(math.Pow(1.5, float64(level-1)) * 200))
}

// RewardCredits — кредит-награда при достижении соответствующего
// уровня. Источник: legacy `consts.php:905` (POINTS_PER_MINNING_LEVEL).
// После уровня 13 награда стабильно равна последнему элементу (300).
var RewardCredits = map[int]int64{
	1:  10,
	2:  25,
	3:  50,
	4:  75,
	5:  100,
	6:  125,
	7:  150,
	8:  175,
	9:  200,
	10: 225,
	11: 250,
	12: 275,
	13: 300,
}

// rewardForLevel — награда при достижении `level`. Для level>13 =
// последняя ячейка (legacy `Functions.inc.php:1655` — `end($array)`).
func rewardForLevel(level int) int64 {
	if v, ok := RewardCredits[level]; ok {
		return v
	}
	if level > 13 {
		return 300
	}
	return 0
}

// LevelUpResult — результат накопления of_points.
//
// Вызов: `LevelUp(currentLevel, currentPoints, addedPoints)` —
// возвращает финальный уровень и финальные points (вычитая пороги
// при каждом level-up), плюс суммарную credit-награду.
//
// Особенность legacy-формулы (`Functions.inc.php:1648-1665`): pointer'а
// ОБА не reset'ятся после level-up. На каждом уровне вычитается
// `NeedPoints(level)` из `points`, и проверка повторяется. Loop
// гарантирует что после возврата `points < NeedPoints(NewLevel)`.
type LevelUpResult struct {
	NewLevel       int
	NewPoints      int64
	CreditsAwarded int64
	LevelUps       int // сколько раз сработал level-up (для логов)
}

// LevelUp пересчитывает уровень и points после добавления `add` points.
//
// Безопасно вызывать с add=0 (тогда LevelUps=0, NewLevel=current,
// NewPoints=current). Возвращает идентичный результат по любому
// порядку добавления points (sum-проверка в тестах).
func LevelUp(currentLevel int, currentPoints, add int64) LevelUpResult {
	if currentLevel < 0 {
		currentLevel = 0
	}
	if currentPoints < 0 {
		currentPoints = 0
	}
	if add < 0 {
		add = 0
	}
	pts := currentPoints + add
	level := currentLevel
	credits := int64(0)
	ups := 0

	// Защита от бесконечного цикла: NeedPoints растёт экспоненциально,
	// 100 итераций покрывает любые разумные значения (level 100 → ~10^17).
	for i := 0; i < 100; i++ {
		need := NeedPoints(level)
		if pts < need {
			break
		}
		pts -= need
		level++
		credits += rewardForLevel(level)
		ups++
	}

	return LevelUpResult{
		NewLevel:       level,
		NewPoints:      pts,
		CreditsAwarded: credits,
		LevelUps:       ups,
	}
}
