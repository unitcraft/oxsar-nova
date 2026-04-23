# План B: Экспедиции — полный порт с legacy

---

## История (завершено)

### ✅ Базовая механика экспедиций (итерация ~30)
- `ExpeditionHandler` в `fleet/expedition.go`
- 6 исходов с фиксированными вероятностями: resources (30%), artefact (5%),
  extra_planet (5%), pirates (20%), loss (15%), nothing (25%)
- Детерминированный RNG по seed от fleet_id
- Запись в `expedition_reports` + сообщение игроку

### ✅ Полный порт legacy (план 22, итерация 22)
Реализовано в `fleet/expedition.go`:

**Новые исходы (было 6, стало 12):**
- `resource` — ресурсы по формуле res_k (metal + silicon + hydrogen)
- `asteroid` — ресурсы без водорода
- `artefact` — вес зависит от exp_power (hours ≥ 4)
- `extra_planet` — временная планета (expires_at = 12–24ч)
- `xSkirmish` — пираты, масштаб по exp_power и силе флота
- `battlefield` — бой с повреждённым флотом, выжившие переходят к игроку
- `credit` — начисление кредитов с учётом buy_credit за 3 дня
- `delay` — сдвиг fire_at события возврата вправо на 10–30%
- `fast` — сдвиг fire_at влево на 10–60%
- `ship` — корабли (LF или recycler) добавляются в fleet_ships
- `loss` — потеря 5–20% кораблей
- `nothing` — пустой исход

**Формулы:**
- `exp_power = expo_tech + hours×2 + spy_tech/10 × pow(spy_probes, 0.4)`
- `visited_scale` — штраф повторных посещений (pow(visits, -0.2...-0.7))
- `res_k = random(500k,1M) × base × jitter × 2` (2% шанс ×100)
- Weighted random выбор исхода (вместо фиксированных %)
- Jitter ±5% на каждый вес + 0.01% шанс ×10 или обнуления

**Новые миграции:**
- `0043_expedition_visits.sql` — таблица `expedition_visits(user_id, galaxy, system, visits)`
- `0044_planets_expires_at.sql` — колонка `planets.expires_at TIMESTAMPTZ`

**Payload:**
- `transportPayload` дополнен `return_event_id` и `flight_seconds`
- `transport.Send` генерирует `returnEventID` заранее и передаёт в оба события

**Тесты** (`expedition_test.go`):
- `TestCalcExpPower` — формула exp_power
- `TestCalcVisitedScale` — 20 посещений < 0.3
- `TestCalcExpWeights_ZeroPower` — ship/battlefield/artefact = 0 при hours=0
- `TestCalcExpWeights_HighPower` — artefact/credit ненулевые при hours=4
- `TestWeightedChoice_Deterministic` — детерминированность
- `TestCalcResK_Basic` — порядок величины res_k
- `TestCalcPirateCount_Bounds` — min 3, max 500

---

## Открытые задачи (упрощения, принятые сознательно)

### B.1 Воркер: удаление временных планет (приоритет: MEDIUM)

**Проблема:** `planets.expires_at` выставляется, но воркер не удаляет истёкшие планеты.

**Решение:** Добавить cron-handler или отдельный KindExpirePlanet event:
```sql
DELETE FROM planets WHERE expires_at IS NOT NULL AND expires_at < now()
```
Можно добавить как периодическую задачу в `cmd/worker/main.go` (раз в час).

**Проверка готовности:**
- [ ] Воркер удаляет планеты с `expires_at < now()`
- [ ] Тест или интеграционный тест

---

### B.2 credit_purchases таблица (приоритет: LOW → реализовать вместе с платёжной системой)

**Проблема:** `expCredit` делает запрос к `credit_purchases`, которой не существует.
Сейчас `buyCredit = 0` (ошибка игнорируется), формула работает без этой компоненты.

**Решение:** Создать таблицу `credit_purchases(user_id, amount, created_at)` при
реализации платёжной системы (план F). Подключить к `expCredit` автоматически.

---

### B.3 black_hole и unknown (приоритет: NONE)

`black_hole` (id=4) не реализован даже в legacy. `unknown` (id=13) зарезервирован.
Не реализовывать до появления требований.
