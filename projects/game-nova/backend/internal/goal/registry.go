package goal

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5"
)

// SnapshotEvaluator — функция, которая по текущему состоянию БД
// возвращает прогресс цели для пользователя.
//
// Возвращаемый прогресс ограничивается target в Engine.updateProgress;
// функция может возвращать «сырое» значение (например, текущий уровень
// здания), движок сам обрежет до target и решит, completed ли.
type SnapshotEvaluator func(ctx context.Context, tx pgx.Tx, userID string, cond ConditionSpec) (int, error)

// CounterMatcher — функция, которая решает, инкрементировать ли
// counter-цель при event'е.
//
// Реализации (см. conditions/event_count.go) могут проверять
// event_kind, фильтр на payload и т.п.
type CounterMatcher func(eventKind int, payload []byte, cond ConditionSpec) bool

var (
	registryMu       sync.RWMutex
	snapshotRegistry = map[string]SnapshotEvaluator{}
	counterRegistry  = map[string]CounterMatcher{}
)

// RegisterSnapshot — добавить snapshot-условие в registry. Вызывается
// в init() файлов condition-типов (conditions/building_level.go и т.п.).
//
// Повторная регистрация одного типа — паника (программная ошибка).
func RegisterSnapshot(condType string, fn SnapshotEvaluator) {
	registryMu.Lock()
	defer registryMu.Unlock()
	if _, ok := snapshotRegistry[condType]; ok {
		panic("goal: snapshot condition already registered: " + condType)
	}
	snapshotRegistry[condType] = fn
}

// RegisterCounter — добавить counter-условие в registry.
func RegisterCounter(condType string, fn CounterMatcher) {
	registryMu.Lock()
	defer registryMu.Unlock()
	if _, ok := counterRegistry[condType]; ok {
		panic("goal: counter condition already registered: " + condType)
	}
	counterRegistry[condType] = fn
}

// snapshotByType — внутренний lookup, используется Engine.
func snapshotByType(condType string) (SnapshotEvaluator, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	fn, ok := snapshotRegistry[condType]
	return fn, ok
}

// counterByType — внутренний lookup.
func counterByType(condType string) (CounterMatcher, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	fn, ok := counterRegistry[condType]
	return fn, ok
}

// IsSnapshot — true если condType зарегистрирован как snapshot.
// Используется в тестах и Engine.
func IsSnapshot(condType string) bool {
	_, ok := snapshotByType(condType)
	return ok
}

// IsCounter — true если condType зарегистрирован как counter.
func IsCounter(condType string) bool {
	_, ok := counterByType(condType)
	return ok
}
