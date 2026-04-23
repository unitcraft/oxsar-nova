# План 11: Доведение UI-порта до полного паритета с легаси

Закрывает упрощения, зафиксированные в `docs/simplifications.md` секция
**UI Porting (H-план, 2026-04-23)**. План рассчитан на последовательное
закрытие, но задачи независимы и могут браться отдельно.

Допущения:
- backend уже имеет скелеты (settings / empire / notepad / search /
  techtree / battlestats / referral) — в этом плане в основном
  расширения, не новые домены.
- frontend API-клиент и UI-примитивы (`ox-panel`, `ox-input`, toast)
  уже используются во всех новых экранах.

---

## Оценка упрощений

Каждое упрощение оценено по трём осям: **оправдано** (было ли решение
сознательным трейдоффом, а не срезанием угла), **стоимость возврата**
(сколько работы, чтобы закрыть), **ценность** (что даёт игроку).

| № | Упрощение | Оправдано | Стоимость | Ценность |
|---|-----------|-----------|-----------|----------|
| 1 | H.1.7 Messages — папки без producer'ов | Частично | M | H |
| 2 | H.1.5 Galaxy — цвета альянс-отношений | Да | M | M |
| 3 | H.1.6 Score — координаты гл. планеты | Да | S | L |
| 4 | H.2.10 Search — переход из результатов | Нет | S | M |
| 5 | H.2.2 Techtree — без графа | Да | M | L |
| 6 | H.1.9 Market — торговля флотом | Да | L | M |
| 7 | H.3 Settings — delete account / dnd | Частично | L | M |
| 8 | H.2.6 Records | Да | S | L |
| 9 | H.2.7 ResTransferStats | Да | S | L |
| 10 | H.2.9 Friends | Частично | M | M |
| 11 | H.2.11 Payment UI (пресеты) | Да | S | L |

Комментарии:

1. **Messages-папки без producer'ов** — UI-заглушка без продюсеров
   показывает пустые вкладки. Оправдано частично: пустая вкладка лучше,
   чем перестройка FOLDERS потом, но пользователь видит «сломанное».
   Надо закрыть, пока свежо.
2. **Alliance relations** — оправдано, backend-слой отсутствует, требует
   миграцию + CRUD. Не экстренно, но нужно для боёвки/дипломатии.
3. **Координаты в рейтинге** — JOIN с planets нагружает запрос; выдача
   координат через поиск/галактику покрывает use-case.
4. **Search без контекста цели** — не оправдано: добавить пропсы
   `initialTarget` / `initialQuery` — тривиально, пропускать стыдно.
5. **Techtree без SVG** — сетка карточек с ✓/✗ требованиями
   информативнее графа; добавление react-flow (~30 KB) не даёт
   пропорционального UX-выигрыша.
6. **Market — торговля флотом** — оправдано: это отдельная ордерная
   книга (sell_ships JSON), требует безопасного холдинга юнитов.
7. **Settings — delete account** — удаление каскадом через alliance /
   planets / fleets / messages нетривиально. DnD сортировки — чисто
   миграция + UI, можно сделать отдельно.
8. **Records (рекорды)** — один SQL `MAX() OVER ... LIMIT 1 per kind`.
   Низкая ценность: дублирует топы в score.
9. **ResTransferStats** — legacy показывал топ получателей ресурсов.
   Для nova без колонки `resource_transfer_log` не агрегируется —
   нужно добавить лог (из fleet.transport onSuccess).
10. **Friends** — легко добавить CRUD, но без онлайн-статуса и тесной
    интеграции (напр., подсветка в galaxy) ценность низкая.
11. **Payment пресеты** — план G (07-payments.md) уже реализует выбор
    пакетов. Дополнение (show-what-it-buys) — визуальное, не
    блокирующее.

**Итог:** главный блокер — (4) Search и (1) Messages producers (P-M,
низкая стоимость, высокая ценность). (2) alliance_relations — P-M
стратегически. Остальное — P-L, закрывать по мере расширения features.

---

## Порядок реализации

### Шаг 1 — H.2.10 Search: навигация с контекстом (P-M, 2 файла)

**Задача:** клик по результату поиска открывает нужный экран уже
отфильтрованный / со скроллом.

**Backend:** без изменений.

**Frontend:**
- [ ] `GalaxyScreen.tsx`: prop `initialCoords?: { g: number; s: number }`.
  В `useState(homePlanet.galaxy)` заменить на
  `useState(initialCoords?.g ?? homePlanet.galaxy)`.
- [ ] `ScoreScreen.tsx`: prop `initialQuery?: string`, пробрасывать
  в `PlayersTab`, добавить input-фильтр по username.
- [ ] `App.tsx::AuthenticatedApp`: state
  `galaxyInitialCoords` и `scoreInitialQuery`, заполнять из
  `GlobalSearch onNavigate`, передавать в `<GalaxyScreen>` /
  `<ScoreScreen>`.

**DoD:** Ctrl+K → выбрать планету → открывается Galaxy на её
координатах; выбрать игрока → Score с отфильтрованной строкой и
автоскроллом.

---

### Шаг 2 — H.1.7 Messages: AutoMsg-producer'ы (P-M, 4 endpoint-интеграции)

**Задача:** писать сообщения в новые папки из существующих событий.

**Backend** (`backend/internal/automsg/service.go`):
- [ ] `SendPhalanxScan(ctx, userID, scanData)` → folder=11. Вызывать
  из `internal/fleet/phalanx.go` после успешного скана.
- [ ] `SendAllianceEvent(ctx, userID, kind, payload)` → folder=6.
  Вызывать из `internal/alliance/service.go::JoinRequest`,
  `ApproveJoin`, `Leave`, `RankChange`.
- [ ] `SendArtefactNotice(ctx, userID, artefactKey, kind)` → folder=7.
  Вызывать из `internal/artefact/expire.go` при истечении и
  `internal/artefact/service.go::Activate`.
- [ ] `SendCreditTransaction(ctx, userID, amount, reason)` → folder=8.
  Вызывать из `internal/payment/webhook.go` и
  `internal/referral/service.go::ProcessPurchase`.

**Тесты:** property-тест — каждый домен при изменении состояния
пишет ровно одно сообщение в ожидаемую папку.

**Frontend:** уже готов (папки видимы).

**DoD:** пройти сценарии «скан фалангой / join альянс / активировать
артефакт / купить кредиты» → в соответствующей папке появляется
запись.

---

### Шаг 3 — H.1.5 Galaxy: отношения альянсов (P-M, миграция + CRUD + UI)

**Задача:** цветовое выделение строк галактики по отношениям.

**Backend:**
- [ ] Миграция `0051_alliance_relations.sql`:
  ```sql
  CREATE TABLE alliance_relations (
    from_alliance_id uuid NOT NULL REFERENCES alliances(id) ON DELETE CASCADE,
    to_alliance_id   uuid NOT NULL REFERENCES alliances(id) ON DELETE CASCADE,
    kind             text NOT NULL CHECK (kind IN ('ally','enemy','nap','trade')),
    created_at       timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (from_alliance_id, to_alliance_id)
  );
  ```
- [ ] `internal/alliance/relations.go`: `SetRelation`, `DropRelation`,
  `ListForAlliance`. RBAC: только leader/vice-leader.
- [ ] `internal/galaxy/repository.go::SystemView` — `LEFT JOIN
  alliance_relations` на `(me.alliance_id, target.alliance_id)`,
  поле `relation *string` в CellView.
- [ ] Endpoints: `GET /api/alliance/{id}/relations`,
  `POST /api/alliance/{id}/relations`, `DELETE`.

**Frontend:**
- [ ] `GalaxyScreen.tsx::CellView` + `relation` в строках, CSS-классы
  `.galaxy-row-ally` (зелёный фон), `.galaxy-row-enemy` (красный),
  `.galaxy-row-nap` (синий), `.galaxy-row-trade` (жёлтый).
- [ ] `AllianceScreen.tsx`: новая вкладка «Дипломатия» со списком
  и CRUD.
- [ ] Легенда в tfoot GalaxyScreen.

**DoD:** лидер альянса устанавливает отношение; цвет появляется у
всех членов альянса в галактике.

---

### Шаг 4 — H.1.9 Market: ордерная книга флота (P-M, миграция + CRUD)

**Задача:** лоты с продажей кораблей (пакет ship_id → count) за
ресурсы или кредиты.

**Backend:**
- [ ] Миграция `0052_fleet_lots.sql`:
  ```sql
  ALTER TABLE market_lots
    ADD COLUMN kind text NOT NULL DEFAULT 'resource'
      CHECK (kind IN ('resource','fleet')),
    ADD COLUMN sell_fleet jsonb;
  CREATE INDEX ix_market_lots_kind ON market_lots(kind, state);
  ```
  `sell_fleet` формат: `{"202": 50, "204": 10}` (unit_id → count).
- [ ] `market.CreateFleetLot(ctx, userID, planetID, fleet, buyRes,
  buyAmount)` — атомарно: списать ships с планеты, положить в
  `market_lots` с `kind='fleet'`, состояние `open`.
- [ ] `market.AcceptFleetLot(ctx, userID, planetID, lotID)` —
  атомарно: списать buy-ресурс у покупателя, зачислить продавцу,
  зачислить ships покупателю (в ships таблицу его планеты),
  state → `closed`.
- [ ] `CancelLot` для kind='fleet' — вернуть ships продавцу.
- [ ] `ListLots` принимает `kind` query-параметр.

**Frontend:**
- [ ] `MarketScreen.tsx::LotsPanel`: вкладки «Ресурсы» / «Флот».
  Форма создания лота флота — селектор кораблей (SHIPS фильтр по
  inventory планеты), input количества, buy_resource, buy_amount.
- [ ] Таблица лотов флота: колонка «Состав» (иконки + числа).

**Тесты:** holding — юниты должны быть в «морозилке», атомарность
принятия лота (write-lock на lot row).

**DoD:** создать лот с 100 крейсерами за 1M металла, купить другим
игроком; инвентарь обновляется у обоих.

---

### Шаг 5 — H.2.9 Friends (P-M, CRUD + presence + galaxy-подсветка)

**Backend:**
- [ ] Миграция `0053_friends.sql`:
  ```sql
  CREATE TABLE friends (
    user_id     uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    friend_id   uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at  timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, friend_id),
    CHECK (user_id != friend_id)
  );
  ```
- [ ] `internal/friends/handler.go`: List, Add, Remove.
- [ ] Расширить `galaxy/repository.go::SystemView` — флаг
  `is_friend` в CellView (EXISTS через friends).

**Frontend:**
- [ ] `features/friends/FriendsScreen.tsx` — таблица: ник, очки,
  last_seen (онлайн-статус), кнопка удалить.
- [ ] Galaxy: иконка ⭐ рядом с ником друга; кнопка «Добавить в друзья»
  в MissionButtons.

**DoD:** добавить игрока в друзья; в галактике подсветка; в
FriendsScreen видно онлайн-статус (~last_seen <5min = «онлайн»).

---

### Шаг 6 — H.3 Settings: опасная зона (P-L, cascade)

**Задача:** DELETE /api/me + UI с подтверждением через ввод ника.

**Backend:**
- [ ] `internal/settings/delete.go::DeleteAccount(ctx, userID, confirmUsername)`:
  1. Verify username совпадает.
  2. Soft-delete: `UPDATE users SET deleted_at=now(), username='[deleted]'+id`.
  3. Cascade: пометить planets.destroyed_at=now() (ресурсы возвращаются
     как обломки? нет — просто удалить), отменить pending fleets
     (возврат не нужен), покинуть alliance, обнулить market_lots.
  4. Transaction.
- [ ] Endpoint `DELETE /api/me` (требует `confirm_username` в body).

**Frontend:**
- [ ] `SettingsScreen.tsx`: новая секция «Опасная зона» (красная
  граница), кнопка «Удалить аккаунт» → modal с input для ника,
  disabled пока не совпадает.
- [ ] После успеха: `logout()` + toast.

**DoD:** удаление необратимо для конкретного юзера; другие игроки не
видят его в рейтингах/галактике/сообщениях.

---

### Шаг 7 — H.3 Settings: drag&drop сортировка планет (P-L)

**Backend:**
- [ ] Миграция `0054_planets_sort_order.sql`:
  ```sql
  ALTER TABLE planets ADD COLUMN sort_order integer NOT NULL DEFAULT 0;
  CREATE INDEX ix_planets_user_sort ON planets(user_id, sort_order);
  ```
- [ ] `PATCH /api/planets/order` — body `{"planet_ids": ["uuid1","uuid2",...]}`,
  SQL `UPDATE planets SET sort_order = pos WHERE id = $1 AND user_id = $2`.
- [ ] `planetH.List` и galaxy — `ORDER BY sort_order, created_at`.

**Frontend:**
- [ ] Добавить в SettingsScreen секцию «Порядок планет» с
  `react-dnd` или нативным `draggable` (без библиотеки).
- [ ] `PlanetSwitcher` и header respect sort_order.

**DoD:** перетащить планету в списке — порядок сохраняется и
отражается во всех местах (header switcher, empire, galaxy jump).

---

### Шаг 8 — H.1.6 Score: колонка координат (P-L, 1 JOIN)

**Backend:**
- [ ] `internal/score/service.go::Top` — добавить JOIN:
  ```sql
  LEFT JOIN LATERAL (
    SELECT galaxy, system, position FROM planets
    WHERE user_id = u.id AND destroyed_at IS NULL
    ORDER BY created_at ASC LIMIT 1
  ) hp ON true
  ```
  Поля `home_galaxy, home_system, home_position` в `Entry`.

**Frontend:**
- [ ] `ScoreScreen::PlayersTab` — колонка «Координаты» (кликабельно,
  открывает GalaxyScreen на этих координатах — см. шаг 1).

**DoD:** в рейтинге виден клик по координатам → переход в галактику.

---

### Шаг 9 — H.2.2 Techtree: SVG-граф зависимостей (P-L, opt-in view)

**Задача:** альтернативный «Граф» вид рядом с текущим «Карточки».

**Frontend (только):**
- [ ] `npm install reactflow` (~70 KB gzip).
- [ ] `features/techtree/TechtreeGraph.tsx`: узлы из `nodes`, рёбра
  по requirements. Layout через `dagre` (или ручной по kind-columns).
- [ ] Кнопка-тумблер «Карточки / Граф» в `TechtreeScreen`.
- [ ] Узлы цветом: зелёный=unlocked, серый=locked, accent=current>0.

**DoD:** граф рендерится, клик по узлу скроллит к карточке; при
большом числе узлов (~100) производительность приемлемая (60fps пан/зум).

---

### Шаг 10 — H.2.6 Records (рекорды) (P-L, 1 endpoint + экран)

**Backend:**
- [ ] `internal/records/handler.go::List`:
  ```sql
  SELECT category, username, value FROM (
    SELECT 'max_metal_mine' AS category, u.username, MAX(b.level) AS value
    FROM buildings b JOIN planets p ON ... JOIN users u ON ...
    WHERE b.unit_id = 1 GROUP BY u.username ORDER BY value DESC LIMIT 1
  ) t UNION ALL ...
  ```
  Или генерировать один запрос на категорию и склеивать в сервисе.
- [ ] Endpoint `GET /api/records`.

**Frontend:**
- [ ] `features/records/RecordsScreen.tsx` — таблица «Категория /
  Держатель / Значение / Мой результат».

**DoD:** страница показывает топ-1 по каждой категории + собственный
показатель игрока.

---

### Шаг 11 — H.2.7 ResTransferStats (P-L, лог + агрегация)

**Backend:**
- [ ] Миграция `0055_resource_transfers.sql`:
  ```sql
  CREATE TABLE resource_transfers (
    id           bigserial PRIMARY KEY,
    from_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
    to_user_id   uuid REFERENCES users(id) ON DELETE SET NULL,
    metal        numeric(20,0) NOT NULL DEFAULT 0,
    silicon      numeric(20,0) NOT NULL DEFAULT 0,
    hydrogen     numeric(20,0) NOT NULL DEFAULT 0,
    at           timestamptz NOT NULL DEFAULT now()
  );
  CREATE INDEX ix_rt_from ON resource_transfers(from_user_id, at DESC);
  CREATE INDEX ix_rt_to   ON resource_transfers(to_user_id,   at DESC);
  ```
- [ ] `internal/fleet/transport.go::OnArrive` — вставлять запись
  при mission=7 (Transport).
- [ ] `internal/score/handler.go::ResTransferStats`:
  `GET /api/stats/resource-transfers?direction=sent|received&period=week|month|all`
  → топ-20 (суммарное по total_value = m + 2*si + 4*h).

**Frontend:**
- [ ] В `ScoreScreen` новая вкладка «Торговля» с 2 под-вкладками
  «Отправители» / «Получатели».

**DoD:** трансферы логируются, топ обновляется в реальном времени.

---

### Шаг 12 — H.2.11 Payment UI: пресеты с описанием (P-L)

**Frontend только:**
- [ ] `features/payment/CreditsScreen.tsx`: к карточкам пакетов
  добавить блок «Что можно купить» (ссылки на Officers, Artefacts
  с суммами).
- [ ] Кнопка «Своя сумма» → input (min=100, max=10000).
- [ ] Подсказка с рекомендациями (500 кр. = 1 мес. офицера, 1000 кр.
  = полный набор офицеров).

**DoD:** экран показывает ценность каждого пакета без ухода с него.

---

## Сводка

**Объём:**
- 5 миграций (0051..0055).
- 4 новых backend-домена: alliance_relations, friends, records, resource_transfers (+ extend market, settings, score).
- ~12 новых frontend-файлов/экранов.
- 4 новых producer'а в `automsg`.

**Затраты (человеко-часы, грубо):**
- Шаг 1 Search — 2ч.
- Шаг 2 Messages producers — 4ч.
- Шаг 3 Alliance relations — 8ч.
- Шаг 4 Fleet market — 10ч.
- Шаг 5 Friends — 6ч.
- Шаг 6 Delete account — 4ч.
- Шаг 7 Planet sort — 3ч.
- Шаг 8 Score coords — 1ч.
- Шаг 9 Techtree graph — 4ч.
- Шаг 10 Records — 3ч.
- Шаг 11 ResTransfers — 4ч.
- Шаг 12 Payment presets — 2ч.
- **Итого:** ~51ч.

**Рекомендация:** делать шаги 1-3 ближайшими итерациями (высокая
ценность, малые/средние затраты); 4-5-9 вместе как «социальный»
пакет; 6-7-8-10-11-12 — как заполнители между крупными планами.

После каждого шага: запись в simplifications.md (секция «Закрытые»),
итерация в project-creation.txt, коммит.
