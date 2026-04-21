package rocket

// defStack — стек единицы обороны с кол-вом и прочностью корпуса.
type defStack struct {
	UnitID int
	Count  int64
	Shell  int
}

// damageResult — итог ракетного удара для одного стека.
type damageResult struct {
	UnitID int
	Lost   int64
}

// applyRocketDamage вычисляет потери обороны при ракетном ударе.
//
// Параметры:
//   - rocketCount: кол-во ракет
//   - stacks: оборона цели (без ABM — они обрабатываются отдельно)
//   - targetUnitID: приоритетная цель (0 = равномерно)
//
// Возвращает список потерь по стекам.
func applyRocketDamage(rocketCount int64, stacks []defStack, targetUnitID int) []damageResult {
	if len(stacks) == 0 || rocketCount == 0 {
		return nil
	}

	totalDamage := rocketCount * int64(missileDamage)
	remainingDamage := totalDamage
	var losses []damageResult
	var otherStacks []defStack

	if targetUnitID > 0 {
		for _, d := range stacks {
			if d.UnitID != targetUnitID {
				otherStacks = append(otherStacks, d)
				continue
			}
			killed := remainingDamage / int64(d.Shell)
			if killed > d.Count {
				killed = d.Count
			}
			if killed > 0 {
				remainingDamage -= killed * int64(d.Shell)
				losses = append(losses, damageResult{UnitID: d.UnitID, Lost: killed})
			}
		}
	} else {
		otherStacks = stacks
	}

	var totalPool int64
	for _, d := range otherStacks {
		totalPool += d.Count * int64(d.Shell)
	}
	if totalPool > 0 && remainingDamage > 0 {
		for _, d := range otherStacks {
			share := float64(d.Count*int64(d.Shell)) / float64(totalPool)
			dmg := int64(float64(remainingDamage) * share)
			killed := dmg / int64(d.Shell)
			if killed > d.Count {
				killed = d.Count
			}
			if killed <= 0 {
				continue
			}
			losses = append(losses, damageResult{UnitID: d.UnitID, Lost: killed})
		}
	}
	return losses
}
