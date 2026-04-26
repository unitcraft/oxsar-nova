// Package artefact — хранение и применение эффектов артефактов.
//
// Ключевая идея (§5.10.1 ТЗ): эффекты МАТЕРИАЛИЗУЮТСЯ в полях
// users / planets. Это повторяет архитектуру oxsar2 (см.
// game/Artefact.class.php::activateArtefact / deactivateArtefact).
//
// Пакет делает три вещи:
//   1. Apply(user, artefact)  — активация: UPDATE полей-факторов.
//   2. Revert(user, artefact) — деактивация: зеркальная операция.
//   3. Resync(user)           — сброс всех полей в дефолт +
//                                переприменение активных (cron).
//
// БЕЗ IO на уровне effects.go — это чистая логика «что и как менять».
// Оркестрация с БД — в service.go.
package artefact

import (
	"errors"
	"fmt"

	"github.com/oxsar/nova/backend/internal/config"
)

// Направление применения эффекта.
type direction int

const (
	dirApply  direction = 1
	dirRevert direction = -1
)

// FactorChange — одна операция над полем-фактором.
//
// Scope определяет, чьё поле менять:
//   scope_user          — users.<Field>
//   scope_planet        — planets.<Field> WHERE id = PlanetID
//   scope_all_planets   — planets.<Field> WHERE user_id = UserID
//
// Op = "set" -> записать NewValue (используется для MERCHANTS_MARK).
// Op = "add" -> прибавить Delta (для всех остальных факторов).
type FactorChange struct {
	Scope    string
	Field    string // exchange_rate | research_factor | build_factor | produce_factor | energy_factor | storage_factor
	Op       string // set | add
	Delta    float64
	NewValue float64 // только для Op=set
}

// ErrNonStackable — попытка активировать не-stackable артефакт, когда
// такой уже активен. Соответствует getActiveCount() > 0 в oxsar2.
var ErrNonStackable = errors.New("artefact: non-stackable already active")

// ErrUnsupported — артефакт описан, но его тип эффекта не
// поддерживается effects-слоем (например, one_shot).
// Его реализация — в M5.1.
var ErrUnsupported = errors.New("artefact: effect type not supported yet")

// BattleModifier — итоговые множители для боевых параметров атакующей стороны.
// Применяются in-memory во время боя, не материализуются в БД.
type BattleModifier struct {
	AttackMul float64 // множитель атаки (1.0 = без изменений)
	ShieldMul float64 // множитель щита
	ShellMul  float64 // множитель брони
}

// ComputeBattleModifier суммирует эффекты нескольких активных battle_bonus
// артефактов в один итоговый множитель через перемножение.
func ComputeBattleModifier(specs []config.ArtefactSpec) BattleModifier {
	m := BattleModifier{AttackMul: 1, ShieldMul: 1, ShellMul: 1}
	for _, s := range specs {
		if s.Effect.Type != "battle_bonus" {
			continue
		}
		if s.Effect.BattleAttack != 0 {
			m.AttackMul *= s.Effect.BattleAttack
		}
		if s.Effect.BattleShield != 0 {
			m.ShieldMul *= s.Effect.BattleShield
		}
		if s.Effect.BattleShell != 0 {
			m.ShellMul *= s.Effect.BattleShell
		}
	}
	return m
}

// computeChanges возвращает одну операцию, которую нужно применить к БД
// для данного артефакта и направления. nil — если эффект не
// материализуется в factor-полях (battle-бонус: применяется in-memory,
// одноразовое действие: реализуется отдельно).
func computeChanges(spec config.ArtefactSpec, dir direction) (*FactorChange, error) {
	e := spec.Effect
	switch e.Type {
	case "factor_user":
		return factorChange("user", e, dir)
	case "factor_planet":
		return factorChange("planet", e, dir)
	case "factor_all_planets":
		return factorChange("all_planets", e, dir)
	case "battle_bonus":
		return nil, nil
	case "one_shot":
		return nil, ErrUnsupported
	default:
		return nil, fmt.Errorf("artefact: unknown effect type %q", e.Type)
	}
}

func factorChange(scope string, e config.ArtefactEffect, dir direction) (*FactorChange, error) {
	if !allowedField(e.Field) {
		return nil, fmt.Errorf("artefact: unsupported field %q", e.Field)
	}
	switch e.Op {
	case "set":
		v := e.ActiveValue
		if dir == dirRevert {
			v = e.InactiveValue
		}
		return &FactorChange{Scope: scope, Field: e.Field, Op: "set", NewValue: v}, nil
	case "add":
		delta := e.Value
		if dir == dirRevert {
			delta = -delta
		}
		return &FactorChange{Scope: scope, Field: e.Field, Op: "add", Delta: delta}, nil
	default:
		return nil, fmt.Errorf("artefact: unknown op %q", e.Op)
	}
}

// allowedField — белый список полей, чтобы не получить SQL-инъекцию
// через кривой YAML и не словить «UPDATE users SET drop_all_tables=…».
// Имена хардкожены.
func allowedField(f string) bool {
	switch f {
	case "exchange_rate", "research_factor",
		"build_factor", "produce_factor", "energy_factor", "storage_factor":
		return true
	}
	return false
}
