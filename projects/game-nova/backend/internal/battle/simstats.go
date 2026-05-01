package battle

import "time"

// MultiRun — N-кратный прогон Calculate с инкрементом seed.
//
// План 72.1 ч.20.11.7: для UI симулятора (legacy /game.php?go=Simulator)
// нужна сводка по num_sim итераций — победы атак/обороны/ничьи в %,
// средние раунды/потери/обломки/опыт/время. Возвращает SimStats и
// последний Report (его ID сохраняется в battle_reports и передаётся
// фронту как ссылка «Отчёт о сражении»).
//
// Если n ≤ 1, выполняется один прогон, SimStats считается с
// num_sim = 1 и долями 0/100% по фактическому исходу.
func MultiRun(in Input, n int) (SimStats, Report, error) {
	if n < 1 {
		n = 1
	}
	t0 := time.Now()
	var (
		attWins, defWins, draws      int
		totalRounds                  int
		totalMoonChance              float64
		atkLostM, atkLostS, atkLostH int64
		defLostM, defLostS, defLostH int64
		atkLostPts, defLostPts       float64 // План 72.1.34: средние lost_points.
		atkLostUnits, defLostUnits   int64
		debrisM, debrisS             int64
		// План 72.1.3 / BA-012: суммируем РЕАЛЬНЫЙ опыт из rep.AttackerExp/
		// DefenderExp (atan-based formula в computeExperience). До фикса
		// здесь была proxy-сумма потерь ресурсов противника (defLostM+S+H),
		// что давало в SimStats.AttackerExp значения порядка 10^6 вместо
		// реальных 5-10 очков опыта — UI симулятора показывал бессмыслицу.
		atkExp, defExp int64
		last           Report
	)
	// План 72.1.3 / BA-021: при seed=0 первая итерация (i=0) попадает на
	// guard `rng.New(0) → golden_ratio`, тогда как i=1,2,...,n получают
	// нормальный xorshift. Из-за этого первая симуляция при seed=0 имела
	// другой RNG-character, статистика смещалась. Если seed0 пуст —
	// смещаем индексирование на 1, чтобы Seed никогда не был 0.
	seed0 := in.Seed
	seedOffset := uint64(0)
	if seed0 == 0 {
		seedOffset = 1
	}
	for i := 0; i < n; i++ {
		in.Seed = seed0 + uint64(i) + seedOffset
		rep, err := Calculate(in)
		if err != nil {
			return SimStats{}, Report{}, err
		}
		last = rep
		totalRounds += rep.Rounds
		totalMoonChance += rep.MoonChance
		debrisM += rep.DebrisMetal
		debrisS += rep.DebrisSilicon
		switch rep.Winner {
		case "attackers":
			attWins++
		case "defenders":
			defWins++
		default:
			draws++
		}
		for _, s := range rep.Attackers {
			atkLostM += s.LostMetal
			atkLostS += s.LostSilicon
			atkLostH += s.LostHydrogen
			atkLostPts += s.LostPoints
			atkLostUnits += s.LostUnits
		}
		for _, s := range rep.Defenders {
			defLostM += s.LostMetal
			defLostS += s.LostSilicon
			defLostH += s.LostHydrogen
			defLostPts += s.LostPoints
			defLostUnits += s.LostUnits
		}
		atkExp += int64(rep.AttackerExp)
		defExp += int64(rep.DefenderExp)
	}
	elapsed := time.Since(t0).Seconds()
	fn := float64(n)
	stats := SimStats{
		NumSim:               n,
		AttackerWinPct:       float64(attWins) * 100 / fn,
		DefenderWinPct:       float64(defWins) * 100 / fn,
		DrawPct:              float64(draws) * 100 / fn,
		AvgRounds:            float64(totalRounds) / fn,
		AvgMoonChance:        totalMoonChance / fn,
		AttackerLostMetal:    float64(atkLostM) / fn,
		AttackerLostSilicon:  float64(atkLostS) / fn,
		AttackerLostHydrogen: float64(atkLostH) / fn,
		DefenderLostMetal:    float64(defLostM) / fn,
		DefenderLostSilicon:  float64(defLostS) / fn,
		DefenderLostHydrogen: float64(defLostH) / fn,
		AttackerLostPoints:   atkLostPts / fn,
		DefenderLostPoints:   defLostPts / fn,
		AttackerLostUnits:    float64(atkLostUnits) / fn,
		DefenderLostUnits:    float64(defLostUnits) / fn,
		DebrisMetal:          float64(debrisM) / fn,
		DebrisSilicon:        float64(debrisS) / fn,
		AttackerExp:          float64(atkExp) / fn,
		DefenderExp:          float64(defExp) / fn,
		GenTimeAll:           elapsed,
		GenTime:              elapsed / fn,
	}
	return stats, last, nil
}
