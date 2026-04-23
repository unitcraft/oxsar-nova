# План D: Артефакты

---

## История (завершено)

### ✅ Базовая система артефактов (итерация ~25)
- `artefacts_user` таблица, состояния: held → active → expired
- `ArtefactSpec`: effect type, stackable, lifetime_seconds
- `computeChanges`: применение эффектов к полям планеты (factor_metal, factor_silicon и т.д.)
- `KindArtefactExpire` event — автоматическое истечение
- UI: ArtefactsScreen, список held/active с кнопками активации

### ✅ Рынок артефактов (итерация ~35)
- `artefact_market` таблица, CreateOffer / BuyOffer / CancelOffer
- UI: ArtmarketScreen с фильтром «мои/чужие», inline price input
- Idempotency key для BuyOffer

### ✅ Артефакты из экспедиций (итерация ~30)
- `expArtefact` в `fleet/expedition.go`: случайный артефакт → state=held
- Вес artefact в weighted-random зависит от exp_power (план 22)

---

## Открытые задачи

### D.1 Лимит стеккинга (план 21, приоритет: HIGH)

**Проблема:** `Catalyst` (id=301) и `Power Generator` (id=302) имеют `stackable: true`
без верхнего лимита. 10× Catalyst = +100% производства — потенциальный эксплойт
через экспедиции + artmarket.

**Шаг 1** — `configs/artefacts.yml`:
```yaml
catalyst:
  max_stacks: 3   # +30% производство максимум

power_generator:
  max_stacks: 2   # +30% энергия максимум
```

**Шаг 2** — `backend/internal/config/catalog.go`, `ArtefactSpec`:
```go
MaxStacks int `yaml:"max_stacks,omitempty"`
```

**Шаг 3** — `backend/internal/artefact/service.go`, при активации:
```go
if spec.MaxStacks > 0 {
    var activeCount int
    tx.QueryRow(ctx,
        `SELECT count(*) FROM artefacts_user
         WHERE user_id=$1 AND unit_id=$2 AND state='active'`,
        userID, artefactID).Scan(&activeCount)
    if activeCount >= spec.MaxStacks {
        return ErrMaxStacksReached
    }
}
```

**Шаг 4** — `var ErrMaxStacksReached = errors.New("artefact: max stacks already active")`
HTTP-handler возвращает 400 при этой ошибке.

**Шаг 5** — Тесты:
- Активация в пределах лимита проходит
- Превышение → `ErrMaxStacksReached`
- `max_stacks=0` → лимит не применяется

**Примечание:** Уже активные артефакты сверх лимита (накоплены ранее) не деактивируются —
лимит только на новые активации.

**Проверка готовности:**
- [ ] `ArtefactSpec.MaxStacks` в catalog.go
- [ ] `catalyst.max_stacks: 3`, `power_generator.max_stacks: 2` в artefacts.yml
- [ ] `ErrMaxStacksReached` объявлена и возвращается
- [ ] HTTP-handler возвращает 400
- [ ] Тесты
- [ ] `make test` зелёный
