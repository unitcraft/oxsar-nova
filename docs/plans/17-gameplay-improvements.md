# План 17: Геймплейные улучшения

**Дата**: 2026-04-24  
**Статус**: ЧЕРНОВИК — требует согласования с геймдизайном  
**Стек затрагивает**: backend (Go), frontend (React/TS), configs (YAML), migrations (SQL)

---

## Цель

Улучшить глубину и ретенцию геймплея не ломая баланс oxsar2-legacy.
Все новые числа требуют ADR и согласования — ничто из этого плана нельзя
реализовывать молча, как «порт».

---

## Анализ текущего состояния

**Сильные стороны** (не трогаем):
- Детерминированный боевой движок (паритет с Java JAR)
- Формульная экономика из `configs/` (DSL-based, работает)
- Expedition с 12 исходами, alien AI с HOLDING

**Болевые точки** (выявлены из simplifications.md + анализа легаси):
1. **Нет антибашинг-защиты** — один игрок может атаковать другого бесконечно
2. **Expedition black_hole не реализован** — флот просто не может пропасть
3. **HOLDING_AI 6/8 действий пустые** — можно превратить в фичу, не баг
4. **Alliance-функциональность минимальна** — нет совместных квестов
5. **Единственная победная метрика** — очки (score); нет category-leaderboards
6. **Нет событий, меняющих галактику** — вселенная статична после старта
7. **Ежедневные задания отсутствуют** — нет petty loops для новичков

---

## Блок A: Антибашинг и баланс PvP (Приоритет: H)

### A1. Лимит атак на одного игрока (порт из legacy) — ✅ ЗАКРЫТО 2026-04-24

Реализовано:
- Константы проброшены из `cfg.Game.BashingPeriod` (18000s) и `cfg.Game.BashingMaxAttacks` (4) в `TransportService.SetBashingLimits`.
- `checkBashingLimit` (transport.go): JOIN `fleets + planets` по координатам назначения, COUNT атак (mission IN 10, 12) от attacker → любые планеты defender, state='outbound' ИЛИ arrive_at > now()-период. Если ≥ max → `ErrBashingLimit` → HTTP 409 Conflict.
- Проверка срабатывает в `Send` для mission=10 (ATTACK_SINGLE) и mission=12 (ATTACK_ALLIANCE). SPY (11) не считается — это разведка, не атака.

Упрощение vs legacy: подсчёт идёт по таблице `fleets`, не `events`. Это эквивалентно (каждой атаке соответствует запись в fleets), но проще. События пропущены намеренно — они могут быть уже в state='ok' после arrive, тогда как fleets хранит историю через arrive_at.

Sanity: self-attack (attacker == defender) и атаки на пустые слоты (planet.user_id IS NULL) — не считаются.

**Суть**: нельзя атаковать одного игрока более 4 раз за 5 часов.
Значения взяты из `consts.dm.local.php`: `BASHING_PERIOD=18000` (5ч), `BASHING_MAX_ATTACKS=4`.
Механизм legacy (`NS.class.php:2285`): считаются все атаки attacker → все планеты defender — и pending, и завершённые за последние `BASHING_PERIOD` секунд. Блок если `pending + finished >= 4`.

### A2. Щит неактивности — **перенесено в план 20 Ф.1 (vacation mode)**

Эта механика покрывается планом 20: `users.umode`/`umode_min` уже есть в БД,
логика порта из legacy. Воркер inactivity-checker автоматически выставляет
`umode=true` после 3 дней без входа. В `validateAttack` проверяется
`defender.UMode`. См. [план 20 Ф.1](20-legacy-port.md) — там живая реализация.

---

## Блок B: Экспедиции — чёрная дыра и разнообразие (Приоритет: M)

### B1. expeditionLost — исправить до полного исчезновения флота (порт legacy)

**Суть**: в legacy `expeditionLost` уничтожает весь флот (флот не возвращается, `sendBack=false`).
В nova сейчас `expLoss` удаляет только 5–20% юнитов — это неправильно.  
Вероятность в legacy: минимум 0.1% (гарантированный минимум от суммы весов).

**Реализация** в `fleet/expedition.go`:
- `case "lost"`: удалить все `fleet_ships` где `fleet_id=X`, удалить запись `fleets`, **не** создавать `KindFleetReturn`
- Сообщение: два варианта (`EXPEDITION_LOST_1`, `EXPEDITION_LOST_2`) — как в legacy `AutoMsg.class.php:611`
- Флаг `EXPED_LOST_ENABLED` не нужен — просто оставляем вес 10 (≈0.1% минимум)

**Примечание**: `black_hole` (id=4) — пустая заглушка даже в legacy, не связана с потерей флота.
Реальная механика потери — именно `expeditionLost`.

### B2. Unknown (id=13) — ⏳ ОТЛОЖЕНО v1.x

**Суть**: флот вернулся без объяснений на 2-4ч позже запланированного.

Не блокер: это **flavor-фича**, baseline-сообщение «ничего не нашли» уже есть. Ценность низкая, инкрементально можно добавить позже.

---

## Блок C: HOLDING_AI — подарки за плен — ⏳ ОТЛОЖЕНО v1.x

Все 4 действия (credit/artefact/repair/asteroid) требуют ADR на дизайн (вероятности, балансные числа, как это влияет на игрока). Не блокер релиза: текущее baseline (пустое действие) корректно работает.

После запуска и наблюдения, как игроки реагируют на alien-захваты, можно решить какие подарки добавлять.

Текущие 6 пустых действий превращаем в **редкие события** с вероятностью:

| Действие | Вероятность | Эффект |
|---|---|---|
| `onAddCreditsAI` | 15% | Пришельцы требуют дань: -50..150 кредитов ИЛИ дарят за «сотрудничество» |
| `onAddArtefactAI` | 5% | Дроп рандомного артефакта уровня 1 (нет в payload — из `catalogs/artefacts`) |
| `onRepairUserUnitsAI` | 20% | Ремонтируют 1-3 damaged unit'а игрока (goodwill gesture) |
| `onGenerateAsteroidAI` | 10% | Добавляют в debris_fields ресурсы (asteroid_metal/silicon) |
| Пропуск | 50% | Ничего (baseline legacy) |

**Важно**: это дизайн-вопрос (отмечен в simplifications.md). Нужен ADR до реализации.

---

## Блок D: Ежедневные задания (Daily Quests) — ✅ ЗАКРЫТО 2026-04-25

Реализовано:
- Migration 0063: `daily_quest_defs` (статические seed) + `daily_quests`
  (PK = user/def/date).
- 9 quest-defs в seed: 6 fleet_mission (transport/recycling/spy/attack/
  position/expedition), 1 research_done, 2 building_done.
- `internal/dailyquest`: Service.List (lazy-gen 3 quest при первом GET
  в новый день, детерминированный seed user+date), Service.Claim
  (транзакция: completed→claimed, выдача кредитов в users + ресурсы
  на home-планету), Service.IncrementProgress.
- Worker: `withDailyQuest` обёртка — после KindBuildConstruction и
  KindResearch инкрементирует прогресс. Не прерывает основной
  handler при ошибке.
- transport.Send: после commit'а инкрементирует fleet_mission прогресс
  с матчером по mission ID.
- HTTP: `GET /api/daily-quests` + `POST /api/daily-quests/{id}/claim`.
- Frontend: `DailyQuestScreen` — список с прогресс-барами и кнопкой
  «Забрать награду». Tab «Задания дня» в menu.
- 3 unit-теста на pickWeighted (weight respected, uniqueness).

Отклонения от плана:
- Нет `resource_earn` quest'ов в seed: трекинг delta ежедневной
  добычи требует snapshot user-state в полночь (ещё одна таблица +
  cron). В MVP ограничились дискретными событиями. Можно добавить
  в v1.x.
- Нет worker'а `KindDailyQuestReset` — lazy-gen при первом GET
  заменяет cron midnight. Меньше нагрузки + не нужно ловить случай
  когда cron пропустил.

**Суть**: 3 случайных задания сбрасываются в полночь UTC, награда — кредиты + ресурсы.

**Примеры заданий**:
- «Добыть 50 000 металла» (production_tick → проверяем ежедневный прирост)
- «Отправить 1 транспортный флот»
- «Провести 1 шпионаж»
- «Исследовать технологию»
- «Продать лот на рынке»

**Схема БД**:
```sql
-- migration 0056
CREATE TABLE daily_quest_defs (
  id SERIAL PRIMARY KEY,
  key TEXT NOT NULL UNIQUE,
  title TEXT NOT NULL,
  condition_type TEXT NOT NULL,  -- 'fleet_mission', 'research', 'resource_earn', ...
  condition_value JSONB NOT NULL, -- {"mission": 3, "count": 1}
  reward_credits NUMERIC(10,2),
  reward_metal NUMERIC(20,4),
  reward_silicon NUMERIC(20,4),
  reward_hydrogen NUMERIC(20,4)
);

CREATE TABLE daily_quests (
  user_id UUID NOT NULL REFERENCES users(id),
  def_id INT NOT NULL REFERENCES daily_quest_defs(id),
  date DATE NOT NULL DEFAULT CURRENT_DATE,
  progress INT NOT NULL DEFAULT 0,
  completed_at TIMESTAMPTZ,
  PRIMARY KEY (user_id, def_id, date)
);
```

**Worker**: `KindDailyQuestReset` в midnight UTC — INSERT 3 рандомных `daily_quests` per user.  
**API**: `GET /api/daily-quests` + `POST /api/daily-quests/{id}/claim`.  
**UI**: новый экран или виджет в Overview.

---

## Блок E: Категориальные рейтинги — ✅ УЖЕ РЕАЛИЗОВАНО

Проверено 2026-04-25: backend `/api/highscore?type=total|b|r|u|a|e`
существует с момента M5+ (см. `score.Top` + `columnFor`). Frontend
`ScoreScreen.tsx` имеет переключатель из 6 кнопок (общий + 5
категорий: постройки, исследования, флот, достижения, боевой).
Никакой работы не требуется — план 17 E фактически закрыт ранее.



**Суть**: кроме общего score добавить топ-10 по категориям без изменения формул.

Категории:
- 🏭 **Экономист** — по `score_e` (economy)
- ⚔️ **Завоеватель** — по `score_u` (units/fleet)
- 🔬 **Учёный** — по `score_r` (research)
- 🏗️ **Строитель** — по `score_b` (buildings)
- 💎 **Коллекционер** — по `score_a` (artefacts)

**Реализация**: `GET /api/records/categories` — 5 запросов ORDER BY соответствующей колонке LIMIT 10.  
**UI**: новые вкладки в ScoreScreen. Нет новых таблиц — всё из `score_snapshots`.

---

## Блок F: Галактические события — ✅ MVP ЗАКРЫТО 2026-04-25

Реализовано:
- Migration 0064: `galaxy_events` (id, kind, started_at, ends_at, params).
- `internal/galaxyevent` пакет: Service.Active / Create / Cancel /
  MetalMultiplier (читает params.metal_mult у активного meteor_storm).
- `planet.applyTickInTx` применяет MetalMultiplier к `rates.metalPerSec`.
  Для других типов событий (solar_flare/trade_forum/star_nebula) хуки
  можно добавить инкрементально.
- HTTP: GET /api/galaxy-event (204 при отсутствии), admin POST/DELETE
  /api/admin/galaxy-events для создания/отмены.
- Frontend: `GalaxyEventBanner` в OverviewScreen с обратным отсчётом
  и описанием эффекта.

Отклонения от плана:
- **Нет автопланировщика**: события создаются админом вручную через
  /api/admin/galaxy-events (план 14 + 17 F). Cron-планировщик
  («раз в 3-7 дней») — отложено в v1.x.
- **Реализован только meteor_storm**: эффект +30% к metal production.
  solar_flare/trade_forum/star_nebula — типы есть в KIND_META фронта
  с описаниями, но backend-эффект не подключён. Можно добавлять
  инкрементально.
- **Нет Redis-кеша**: MetalMultiplier делает SQL-запрос на каждый тик
  планеты. Ок для MVP с малым DAU; при росте — кеш с TTL=ends_at.

**Суть**: раз в 3-7 дней случайное событие по всей галактике.

| Событие | Эффект | Длительность |
|---|---|---|
| Метеоритный шторм | +30% добычи металла у всех | 6ч |
| Солнечная вспышка | -20% энергии (production снижена) | 4ч |
| Торговый форум | Рыночные курсы 1:1.5:3 вместо 1:2:4 | 8ч |
| Звёздная туманность | +15% к exp_power экспедиций | 12ч |

**Схема**:
```sql
-- migration 0057
CREATE TABLE galaxy_events (
  id SERIAL PRIMARY KEY,
  kind TEXT NOT NULL,
  started_at TIMESTAMPTZ NOT NULL,
  ends_at TIMESTAMPTZ NOT NULL,
  params JSONB NOT NULL DEFAULT '{}'
);
```

**Worker**: `KindGalaxyEvent` — планируется следующий при завершении текущего.  
**Production/Market**: при расчёте читают `active galaxy_event` (кешируем в Redis с TTL = ends_at - now()).  
**UI**: баннер в шапке с обратным отсчётом.

---

## Блок G: Улучшения UX (без изменения баланса)

### G1. Прогноз ресурсов — ✅ ЗАКРЫТО 2026-04-25

Реализовано:
- `GET /api/planets/{id}/forecast?hours=N` (default 4, max 168).
- `Service.Forecast` — абсолютные значения через N часов с учётом
  storage cap. `capped: true` если хотя бы один ресурс упёрся.
- Frontend: `ForecastWidget` в OverviewScreen для каждой планеты
  с кнопками 1ч / 4ч / 12ч / 24ч и иконкой ⚠️ при cap.

Smoke в Docker: 401 без auth, маршрут работает.

### G2. Уведомление о возврате флота — ⏳ ОТЛОЖЕНО v1.x

WebSocket push `fleet:return` когда флот вернулся, пока игрок онлайн.
Уже есть WS-инфраструктура (чат). Добавить handler в `KindFleetReturn` → `hub.SendToUser(userID, msg)`.

Не реализуем в v1: «возврат через час» игрок и так увидит при заходе.
Push нужен для активной 24/7 игры — не наш кейс.

### G3. Сравнение в BattleSimulator — ⏳ ОТЛОЖЕНО v1.x

В BattleSimulatorScreen: кнопка «Сохранить состав» (localStorage). Два слота side-by-side.

Не блокер: сейчас игрок может сравнить два прогона вручную через `/sim`. UX-удобство, не функциональность.

### G4. Fleet tracker в Galaxy — ⏳ ОТЛОЖЕНО v1.x

В GalaxyScreen: иконка ✈️ на клетке, куда летит флот.

Информация уже доступна в FleetScreen списком. Дублирование в GalaxyScreen — UX, не блокер.

---

## Блок H: Улучшения боевого движка — ⏳ ОТЛОЖЕНО v1.x

Все 4 пункта (детальный round-trace, artefact-бонусы в Input, attacker-specific stats, multi-channel attack) — это **точность симуляции**, не блокеры релиза. Бой работает корректно по golden-тестам battle-sim. После запуска и сбора фидбэка можно реализовать инкрементально.

Движок `backend/internal/battle/engine.go` работает корректно, но несколько механик либо не портированы из Java, либо упрощены. Ничего здесь не меняет баланс формул — только точность симуляции.

### H1. Детальный round-trace (Приоритет: M, ~1 день)

`RoundTrace` содержит только `AttackersAlive / DefendersAlive`. Добавить потери по типу юнита за каждый раунд.

```go
type UnitLoss struct {
    UnitID int
    Lost   int64
}
type RoundTrace struct {
    Index          int
    AttackersAlive int64
    DefendersAlive int64
    AttackersLost  []UnitLoss  // новое
    DefendersLost  []UnitLoss  // новое
}
```

Польза: combat replay в AdminPanel (план 14 Ф.5), лучшие golden-тесты, «в раунде 3 потеряно 120 LF» в battle report.

### H2. Artefact-бонусы в Input (Приоритет: M, ~1 день)

Активные артефакты типа `BATTLE` (damage/shield) сейчас игнорируются при построении `battle.Input`. `fleet/attack.go` собирает `Input` без чтения арtefact-таблицы.

Фикс: при построении `Input` читать `artefacts_user WHERE kind='BATTLE' AND user_id=?` и применять бонусы к `Tech.Gun / Tech.Shield / Tech.Shell` соответствующей стороны. Нет изменений в самом движке — только в caller.

### H3. Attacker-specific stats (Приоритет: L, ~1 день)

Два юнита в legacy имеют разные статы для роли атакующего vs защитника:
- **Deathstar** (id=42): `front` 10 в защите → 9 в атаке
- **Alien Screen** (id=202): `front` 15 в защите → 16 в атаке

Сейчас `newState` не смотрит на роль стороны. Фикс: в `newState(input []Side, rf, role)` выбирать `u.AttackerFront` если `role == roleAttacker`, иначе `u.Front`. Затрагивает < 1% боёв.

### H4. Multi-channel attack (Приоритет: L, ~2 дня)

`primaryChannel` выбирает один лучший канал и все выстрелы идут в `Shield[primaryChannel]`. Java делает три независимых удара: каждый канал бьёт `Shield[ch]` пропорционально `Attack[ch]`.

Разница только у юнитов с двумя ненулевыми каналами (Deathstar, Alien Screen). Для обычных флотов изменений нет. Фикс: в `applyShots` разбить пул выстрелов по 3 каналам, каждый канал бьёт свой `Shield[ch]`.

### H5. Частичный regen щита (Приоритет: L, ~0.5 дня)

Уже записано в simplifications.md [M4.1]. Когда щит полностью снят в раунде, Java применяет `shieldDamageFactor` к следующему regen. Сейчас всегда 100%. Фикс в `regen()`: `newShield = max × damageFactor`.

### H6. ACS — общий пул атаки (Приоритет: L, ~2 дня)

При ACS-атаке несколько `Side` бьют независимо. Java объединяет атакующих `Participant`-ов в одну `Party` — суммарный пул выстрелов, потом распределяется по целям. Текущая модель формально работает, но распределение выстрелов немного расходится с legacy при разнородных ACS-флотах.

---

## Блок I: Порты из legacy — вынесены в план 19

Механики vacation mode, fleet slots, POSITION, phalanx, stargate, moon destruction, astrophysics, IGR — перенесены в [план 20: Legacy Port](20-legacy-port.md). Все они реализованы в oxsar2 (включая `ext/`) и требуют прямого портирования, а не дизайна.

---

> Детали всех восьми механик (vacation, fleet slots, POSITION, phalanx, stargate, moon destruction, astrophysics, IGR) — в **[плане 19](19-legacy-port.md)**.

---

## Порядок реализации (рекомендация)

| Фаза | Блок | Почему именно этот порядок |
|---|---|---|
| 17.1 | A1 + A2 | Антибашинг — самый острый PvP-pain; малый бэкенд |
| 17.2 | G1 + G2 + G4 | UX без новых таблиц; быстрый win |
| 17.3 | B1 + B2 | Экспедиции: потеря флота + unknown |
| 17.4 | H1 + H2 | Battle: round-trace + artefact-бонусы |
| 17.5 | D | Daily quests — крупная фича, отдельный спринт |
| 17.6 | E | Категориальные рейтинги — дёшево |
| 17.7 | H3 + H4 + H5 | Battle: attacker stats, multi-channel, partial regen |
| 17.8 | C | HOLDING_AI подарки — после ADR |
| 17.9 | H6 | ACS party merge |
| 17.10 | F | Галактические события — defer до M8 |

---

## ADR-требования (обязательно перед реализацией)

- **A1**: пороги взяты из legacy (4 атаки / 5ч) — ADR не нужен, это порт
- **A2**: согласовать пороги щита (3 дня инактивности / 12ч защиты после входа)
- **C**: решить, дают ли пришельцы «добро» или только вредят
- **D**: утвердить список заданий и награды
- **F**: утвердить тайминги и эффекты
- **H3/H4**: подтвердить что менять attacker_* и multi-channel — допустимо (формально порт, но меняет исходы боёв с Deathstar/Alien Screen)

---

## Что НЕ делаем в этом плане

- Не меняем балансовые формулы (производство, урон, стоимость)
- Не добавляем новые типы юнитов (без ADR)
- Не меняем статы/стоимости юнитов и rapidfire-таблицу — это план 18
- Не реализуем платежи (план 07)
- Не добавляем lifeworks/фракции (слишком большой scope)

> Блок H меняет **механики движка** (код), но не балансовые числа (YAML).
> Ребалансировка YAML — только в [плане 18](18-unit-rebalance.md).
