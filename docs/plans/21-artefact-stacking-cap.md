# План 21: Лимит стеккинга артефактов

## Проблема

`Catalyst` (id=301) и `Power Generator` (id=302) имеют `stackable: true` в
`configs/artefacts.yml`. Нет верхнего лимита — игрок может активировать
неограниченное количество одинаковых артефактов одновременно:

- 10× Catalyst = +100% к производству всех ресурсов
- 5× Power Generator = +75% к энергии

Это потенциальный эксплойт через систематические экспедиции + artmarket.

## Решение

Добавить поле `max_stacks` в `ArtefactSpec`. При попытке активировать
артефакт сверх лимита — возвращать ошибку.

## Шаги реализации

### Шаг 1 — Добавить max_stacks в конфиг и структуру

В `configs/artefacts.yml` добавить `max_stacks` для стеккируемых артефактов:

```yaml
catalyst:
  max_stacks: 3   # +30% производство максимум

power_generator:
  max_stacks: 2   # +30% энергия максимум

# Остальные stackable — без лимита (или добавить по мере необходимости)
```

В `backend/internal/config/catalog.go`, `ArtefactSpec`:
```go
MaxStacks int `yaml:"max_stacks,omitempty"`
```

### Шаг 2 — Проверка при активации

В `backend/internal/artefact/service.go`, в методе активации артефакта
(поиск: `Activate` или место где артефакт переходит в `active`):

```go
if spec.MaxStacks > 0 {
    var activeCount int
    err := tx.QueryRow(ctx,
        `SELECT count(*) FROM user_artefacts
         WHERE user_id=$1 AND artefact_id=$2 AND state='active'`,
        userID, artefactID,
    ).Scan(&activeCount)
    if err != nil {
        return fmt.Errorf("count active: %w", err)
    }
    if activeCount >= spec.MaxStacks {
        return ErrMaxStacksReached
    }
}
```

### Шаг 3 — Новая ошибка

```go
var ErrMaxStacksReached = errors.New("artefact: max stacks already active")
```

Вернуть `400 Bad Request` из HTTP-хендлера.

### Шаг 4 — Тесты

- Тест: активация третьего Catalyst проходит (max_stacks=3)
- Тест: активация четвёртого Catalyst возвращает `ErrMaxStacksReached`
- Тест: артефакт без `max_stacks` (=0) — лимит не применяется

## Что НЕ меняется

- Уже активные артефакты сверх лимита (если накоплены до введения ограничения) —
  не деактивируются, лимит применяется только к **новым** активациям.
- Нестеккируемые артефакты (`stackable: false`) — без изменений.

## Проверка готовности

- [ ] `ArtefactSpec.MaxStacks` в catalog.go
- [ ] `catalyst.max_stacks: 3`, `power_generator.max_stacks: 2` в artefacts.yml
- [ ] `ErrMaxStacksReached` объявлена и возвращается при превышении
- [ ] HTTP-хендлер возвращает 400 при `ErrMaxStacksReached`
- [ ] Тест: 4-й Catalyst → ошибка
- [ ] `make test` зелёный
