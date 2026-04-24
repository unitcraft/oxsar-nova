# План 23: Лимит полей при постройке зданий

**Цель:** запретить постройку зданий при переполнении `used_fields >=
max_fields(diameter)`. Ввести функцию `max_fields` и проверку в
`building.Service.StartBuild`. Это prerequisite для `terra_former`
(см. план 22 Ф.2.2 — он добавляет +5 полей/уровень, но проверять
пока нечего).

**Scope:** backend-валидация + UI-индикатор. Не трогаем баланс
существующих зданий (все сейчас влезают в стандартный diameter).

**Обнаружено в процессе плана 22**: `UsedFields` в модели планеты
растёт при постройке, показывается в UI, но **не проверяется** — то
есть прямо сейчас можно строить бесконечно.

---

## Legacy reference

`/d/Sources/oxsar2/www/game/Planet.class.php:702-705`:
```php
if ($builds[UNIT_TERRA_FORMER]) {
    $max += $builds[UNIT_TERRA_FORMER] * 5;
}
```
То есть `max_fields = f(diameter) + terra_former_level * 5 [+ moon_lab * 5
для лун]`. `f(diameter)` — обычно `floor(diameter² / 1_000_000)` или
что-то близкое (свериться по `na_construction`-формуле в legacy).

---

## Фазы

### Ф.1 Базовая функция max_fields (HIGH)

- [ ] Добавить в `backend/internal/planet/service.go` (или `model.go`):
  ```go
  // MaxFields возвращает максимум полей для планеты по формуле legacy:
  //   base = floor((diameter/1000)^2)
  //   + terra_former level × 5 (если есть здание)
  //   + moon_lab level × 5 (только для лун)
  func MaxFields(p *Planet, buildings map[int]int) int
  ```
- [ ] Тест `service_test.go`: стартовая Homeworld (18800) → 354 поля;
      +terra_former level 3 → 369.

### Ф.2 Проверка при постройке (HIGH)

- [ ] В `building.Service.StartBuild` (или в queue-validation): если
      `used_fields + 1 > max_fields` — вернуть `ErrFieldsExhausted`.
- [ ] В handler: 400 Bad Request с понятным сообщением.
- [ ] E2E-тест: забить планету до лимита, попытаться построить +1 —
      ожидать 400.

### Ф.3 UI-индикатор (MEDIUM)

- [ ] В `BuildingsScreen`: показывать `X / Y полей` рядом с диаметром.
- [ ] При приближении к лимиту (>90%) — жёлтый badge.
- [ ] При превышении — красный + подсказка «недоступно для постройки».

### Ф.4 Terra Former (LOW, зависит от Ф.1)

После Ф.1 разблокирует план 22 Ф.2.2 кандидата terra_former:
- [ ] В `configs/buildings.yml` добавить terra_former с cost/factor
      из construction.yml (id=58).
- [ ] `MaxFields()` учитывает `terra_former_level × 5`.
- [ ] UI-карточка в BuildingsScreen.

---

## Порядок

1. Ф.1 — 1 час (включая тесты)
2. Ф.2 — 30 мин (простое дополнение к StartBuild)
3. Ф.3 — 30-60 мин (UI)
4. Ф.4 — 20 мин после Ф.1-3

**Итого: ~3-4 часа.**

---

## Связанное

- План 22 (configs-cleanup) — Ф.2.2 terra_former ждёт этого плана.
- План 20 (legacy-port) — планетарные формулы, если будем портировать
  размер планеты по позиции.
