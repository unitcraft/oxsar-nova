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
- `asteroid` — ресурсы без водорода (и шанс +30 если hours=0)
- `artefact` — вес зависит от exp_power (hours ≥ 4, cap = xSkirmish/2)
- `extra_planet` — временная планета (expires_at = 12–24ч из legacy EXPED_PLANET_LIFETIME_MIN/MAX)
- `xSkirmish` — пираты, масштаб по exp_power и силе флота (только hours ≥ 3)
- `battlefield` — бой с повреждённым флотом, выжившие переходят к игроку (hours ≥ 1)
- `credit` — начисление кредитов с учётом buy_credit за 3 дня (hours ≥ 4)
- `delay` — сдвиг fire_at события возврата вправо на 10–30% (hours ≥ 1)
- `fast` — сдвиг fire_at влево на 10–60%
- `ship` — корабли (LF или recycler) добавляются в fleet_ships (только hours ≥ 2)
- `loss` — потеря 5–20% кораблей
- `nothing` — пустой исход

**Формулы (из legacy Expedition.class.php):**
```
exp_power = expo_tech + hours×2 + spy_tech/10 × pow(spy_probes, 0.4)
visited_scale: 3–4 раза → visits^(−0.2), 5–9 → visits^(−0.3),
               10–19 → visits^(−0.5), 20+ → visits^(−0.7)
               × ((hours+1)/6)^1.5
res_k = random(500k,1M) × res_scale × 2   (2% шанс ×100 джекпот)
resourceDiscovery  = ceil(100 × 1.22^exp_power) × jitter
xSkirmish штраф: >10000 кораблей → ×0.5, >100 Deathstar → ×0.1
```

**Новые миграции:**
- `0043_expedition_visits.sql` — таблица `expedition_visits(user_id, galaxy, system, visits)`
- `0044_planets_expires_at.sql` — колонка `planets.expires_at TIMESTAMPTZ`

**Payload:**
- `transportPayload` дополнен `return_event_id` и `flight_seconds`
- `transport.Send` генерирует `returnEventID` заранее и передаёт в оба события

**Тесты** (`expedition_test.go`):
- `TestCalcExpPower` — формула exp_power
- `TestCalcVisitedScale` — 20 посещений даёт visited_scale < 0.3
- `TestCalcExpWeights_ZeroPower` — ship/battlefield/artefact = 0 при hours=0
- `TestCalcExpWeights_HighPower` — artefact/credit ненулевые при hours=4
- `TestWeightedChoice_Deterministic` — детерминированность
- `TestCalcResK_Basic` — порядок величины res_k
- `TestCalcPirateCount_Bounds` — min 3, max 500

---

## Открытые задачи (упрощения, принятые сознательно)

### ✅ B.1 Воркер: удаление временных планет

- `cmd/worker/main.go`: cron раз в час — `DELETE FROM planets WHERE expires_at IS NOT NULL AND expires_at < now()`

---

### B.2 credit_purchases таблица (приоритет: LOW → реализовать вместе с платёжной системой)

**Проблема:** `expCredit` делает запрос к `credit_purchases`, которой не существует.
Сейчас `buyCredit = 0` (ошибка игнорируется), формула работает без этой компоненты.

**Решение:** Создать таблицу `credit_purchases(id, user_id, amount, price_rub, created_at)`
при реализации платёжной системы (план F.1). Подключить к `expCredit` автоматически.

---

### B.3 black_hole и unknown (приоритет: NONE)

`black_hole` (id=4) не реализован даже в legacy. `unknown` (id=13) зарезервирован.
Не реализовывать до появления требований.
