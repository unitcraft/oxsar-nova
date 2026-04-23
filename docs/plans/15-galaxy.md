# План 15: Экран галактики — доработка GalaxyScreen

## Статус: В РАЗРАБОТКЕ (задачи 1–4 выполнены; задачи 5–8 — P2/P3)

## Контекст

Legacy страница `/game.php/Galaxy` — просмотр звёздной системы: планеты, луны,
обломки, игроки, альянсы, статусы активности, кнопки миссий и атак.

**Legacy источник:** `d:\Sources\oxsar2\www\templates\standard\galaxy.tpl` + `d:\Sources\oxsar2\www\game\page\Galaxy.class.php`

**Legacy UI:** http://localhost:8080/game.php/Galaxy (логин test / пароль quoYaMe1wHo4xaci)

**Nova экран:** `frontend/src/features/galaxy/GalaxyScreen.tsx`

**Nova backend:** `backend/internal/galaxy/`

---

## Сравнение: legacy vs nova

### 1. Что есть в nova (реализовано)

| Функция | Статус |
|---------|--------|
| Таблица 15 позиций системы | ✅ |
| Навигация: стрелки и input для галактики/системы | ✅ |
| Данные планеты: имя, тип, координаты | ✅ |
| Луна: наличие, имя | ✅ |
| Обломки: металл/кремний числами | ✅ |
| Владелец: username, rank | ✅ |
| Кнопки миссий (шпионаж, атака, транспорт, переработка, шпионаж луны) | ✅ |
| Авторефреш каждые 10 сек | ✅ |
| Выделение собственных планет (🏠, синий фон) | ✅ |
| Skeleton loading при загрузке | ✅ |

### 2. Что есть в legacy, но отсутствует в nova

| Функция | Приоритет | Комментарий |
|---------|-----------|-------------|
| **Тег и ранг альянса** | ✅ | Колонка «Альянс» с тегом |
| **Активность игрока/планеты** | ✅ | `(*)` / `(N min)` / `(N h)` под именем |
| **Статусы игрока** | ✅ | i/I/b/v — иконки с abbr title |
| **Легенда статусов (footer)** | ✅ | tfoot под таблицей |
| **Tooltip луны** | ✅ | name | diameter км | temp°C при наведении |
| **Tooltip обломков** | ✅ | Металл/Кремний при наведении |
| **Star Surveillance — мониторинг флота** | P2 | Кнопка 👁 если есть здание на луне |
| **Ракетная атака** | P2 | Кнопка 🚀 если цель в радиусе ракет |
| **Расход водорода за просмотр** | P3 | −10 H за просмотр чужой системы |
| **Экспедиция (16-я позиция)** | P3 | Ссылка за пределы системы |

### 3. Данные есть в backend, но не передаются на frontend

| Поле | Где хранится | Статус |
|------|-------------|--------|
| `user.last_activity` | таблица `users` | ❌ нет в CellView |
| `user.umode` (отпуск) | таблица `users` | ❌ нет в CellView |
| `user.banned` | таблица `ban_u` или `users` | ❌ нет в CellView |
| Альянс: тег, ранг, описание | таблица `alliances` / `alliance_members` | ❌ нет в CellView |
| `moon.diameter` | таблица `planets` (is_moon=true) | ❌ нет в CellView |
| `moon.temp_min/temp_max` | таблица `planets` | ❌ нет в CellView |

---

## Структура legacy (что передаётся в шаблон)

### Скалярные переменные:
- `galaxy`, `system` — текущие координаты
- `canMonitorActivity` — есть ли Star Surveillance здание
- `missileRange` — радиус ракетной атаки (вычисляется из уровня MissileBase)

### Массив строк (по позициям 1..15):

**Планета:**
- `systempos` — позиция (1..15)
- `planetname` — имя (ссылка на Mission)
- `picture` — иконка
- `metal`, `silicon` — обломки с tooltip

**Луна:**
- `moonid`, `moonname`, `moonpic`
- `moonsize`, `moontemp` — размер и температура
- `moonactivity` — активность
- `moonrocket` — ссылка RocketAttack (если в диапазоне)

**Игрок:**
- `username`, `userid`
- `rank`, `cur_points`, `e_points`
- `activity` — строка: `(*)`, `(5 min)`, `(1 h)`, пусто
- `user_status_long` — строка статусов: `i`, `I`, `b`, `n`, `v`, `s`

**Альянс:**
- `alliance` — тег (ссылка на страницу)
- `allydesc` — описание tooltip
- `alliance_rank` — ранг в альянсе
- `allypage`, `homepage`, `memberlist` — ссылки

**Действия:**
- `sendesp` — шпионаж планеты
- `sendmoonspy` — шпионаж луны
- `message` — отправить сообщение
- `buddyrequest` — запрос дружбы
- `rocketattack` — ракетная атака (если в диапазоне)
- `monitorfleet` — мониторинг (если есть Star Surveillance)

---

## План реализации

### Задача 1 — Расширить CellView в backend API [P1]

**Описание:** Текущий `CellView` не содержит альянс, активность и статусы.
Нужно расширить SQL-запрос и структуру ответа.

**Backend** — `backend/internal/galaxy/repository.go`:

```go
type CellView struct {
    // существующие поля
    Position      int     `json:"position"`
    HasPlanet     bool    `json:"has_planet"`
    PlanetName    *string `json:"planet_name,omitempty"`
    PlanetID      *string `json:"planet_id,omitempty"`
    PlanetType    *string `json:"planet_type,omitempty"`
    HasMoon       bool    `json:"has_moon"`
    MoonName      *string `json:"moon_name,omitempty"`
    OwnerUsername *string `json:"owner_username,omitempty"`
    OwnerID       *string `json:"owner_id,omitempty"`
    OwnerRank     *int    `json:"owner_rank,omitempty"`
    DebrisMetal   int64   `json:"debris_metal"`
    DebrisSilicon int64   `json:"debris_silicon"`

    // новые поля
    OwnerLastActive *time.Time `json:"owner_last_active,omitempty"`
    OwnerVacation   bool       `json:"owner_vacation,omitempty"`
    OwnerBanned     bool       `json:"owner_banned,omitempty"`
    OwnerNewbie     bool       `json:"owner_newbie,omitempty"`
    OwnerStrong     bool       `json:"owner_strong,omitempty"`  // dm_points >> points
    AllianceTag     *string    `json:"alliance_tag,omitempty"`
    AllianceRank    *int       `json:"alliance_rank,omitempty"`
    MoonDiameter    *int       `json:"moon_diameter,omitempty"`
    MoonTempMin     *int       `json:"moon_temp_min,omitempty"`
    MoonTempMax     *int       `json:"moon_temp_max,omitempty"`
}
```

SQL расширение — добавить в JOIN:
- `LEFT JOIN alliances a ON u.alliance_id = a.id` — тег, ранг
- `LEFT JOIN bans bn ON bn.user_id = u.id AND bn.active = true` — забан
- Поля `u.last_login` → `owner_last_active`
- Поля `u.vacation_mode` → `owner_vacation`
- Поля луны: `moon.diameter`, `moon.temp_min`, `moon.temp_max`

**OpenAPI** — обновить схему `GalaxyCell`:

```yaml
owner_last_active: { type: string, format: date-time, nullable: true }
owner_vacation:    { type: boolean }
owner_banned:      { type: boolean }
owner_newbie:      { type: boolean }
alliance_tag:      { type: string, nullable: true }
alliance_rank:     { type: integer, nullable: true }
moon_diameter:     { type: integer, nullable: true }
moon_temp_min:     { type: integer, nullable: true }
moon_temp_max:     { type: integer, nullable: true }
```

**Файлы:**
- `backend/internal/galaxy/repository.go`
- `api/openapi.yaml`

---

### Задача 2 — Альянс и активность в таблице [P1]

**Описание:** Добавить колонку альянса и строку активности под именем игрока.

**Frontend** — `GalaxyScreen.tsx`:

```tsx
// Колонка альянса — после колонки игрока
<td className="ox-mono">
  {cell.alliance_tag
    ? <span className="alliance-tag">[{cell.alliance_tag}]</span>
    : <span className="ox-muted">—</span>}
  {cell.alliance_rank && (
    <div className="ox-sub">#{cell.alliance_rank}</div>
  )}
</td>

// Активность — под именем игрока в той же ячейке
<div className="ox-sub ox-muted">
  {formatActivity(cell.owner_last_active)}
</div>
```

```ts
// Утилита форматирования активности (как в legacy)
function formatActivity(lastActive?: string | null): string {
  if (!lastActive) return '';
  const mins = Math.floor((Date.now() - new Date(lastActive).getTime()) / 60000);
  if (mins < 15) return '(*)';           // только что
  if (mins < 60) return `(${mins} min)`;
  const hrs = Math.floor(mins / 60);
  if (hrs < 24) return `(${hrs} h)`;
  return '';                              // давно — не показываем
}
```

**Файлы:**
- `frontend/src/features/galaxy/GalaxyScreen.tsx`
- `frontend/src/api/types.ts` — обновить `GalaxyCell`

---

### Задача 3 — Статусы игрока и легенда [P1]

**Описание:** Отображать иконки статусов рядом с именем игрока и добавить
легенду в подвал таблицы.

**Frontend:**

```tsx
// Статус-иконки рядом с именем (маленький текст)
<span className="player-status">
  {cell.owner_vacation && <abbr title="Режим отпуска">v</abbr>}
  {cell.owner_banned   && <abbr title="Забанен">b</abbr>}
  {cell.owner_newbie   && <abbr title="Новичок (защита)">n</abbr>}
  {cell.owner_strong   && <abbr title="Сильный игрок">s</abbr>}
  {isInactive(cell.owner_last_active, 30) && <abbr title="Очень неактивный">I</abbr>}
  {isInactive(cell.owner_last_active, 15) && <abbr title="Неактивный">i</abbr>}
</span>

// Легенда в footer таблицы
<tfoot>
  <tr>
    <td colSpan={7} className="galaxy-legend">
      <b>i</b> неактивный (15+ дн) &nbsp;
      <b>I</b> очень неактивный (30+ дн) &nbsp;
      <b className="banned">b</b> забанен &nbsp;
      <b className="strong">s</b> сильный игрок &nbsp;
      <b className="newbie">n</b> новичок &nbsp;
      <b className="vacation">v</b> отпуск
    </td>
  </tr>
</tfoot>
```

**Файлы:**
- `frontend/src/features/galaxy/GalaxyScreen.tsx`

---

### Задача 4 — Tooltip луны и обломков [P2]

**Описание:** При наведении на луну показывать размер и температуру.
При наведении на обломки — детальный breakdown.

**Frontend:**

```tsx
// Tooltip луны
<span title={cell.moon_diameter
  ? `${cell.moon_name} | ${cell.moon_diameter} км | ${cell.moon_temp_min}..${cell.moon_temp_max}°C`
  : cell.moon_name ?? ''}>
  🌑
</span>

// Tooltip обломков
<span title={`Обломки\nМеталл: ${cell.debris_metal.toLocaleString()}\nКремний: ${cell.debris_silicon.toLocaleString()}`}>
  {cell.debris_metal > 0 && <span className="debris-m">🟠 {fmt(cell.debris_metal)}</span>}
  {cell.debris_silicon > 0 && <span className="debris-s">💎 {fmt(cell.debris_silicon)}</span>}
</span>
```

**Файлы:**
- `frontend/src/features/galaxy/GalaxyScreen.tsx`

---

### Задача 5 — Star Surveillance (мониторинг флота) [P2]

**Описание:** Кнопка мониторинга флота на чужой планете, доступна если у игрока
есть здание Star Surveillance (`unit_id` = STAR_SURVEILLANCE) на луне текущей планеты.

**Backend** — добавить в `GET /api/galaxy/{g}/{s}` поле `can_monitor bool` в ответ
(вычисляется по наличию здания у запрашивающего пользователя).

**Frontend:**

```tsx
{sys?.can_monitor && !isOwn && cell.has_planet && (
  <button onClick={() => openMonitor(cell.planet_id)} title="Мониторинг флота">
    👁
  </button>
)}
```

**Файлы:**
- `backend/internal/galaxy/handler.go` — передать userID в repo
- `backend/internal/galaxy/repository.go` — проверить здание
- `frontend/src/features/galaxy/GalaxyScreen.tsx`

---

### Задача 6 — Ракетная атака [P2]

**Описание:** Кнопка ракеты доступна если:
- Цель находится в пределах радиуса ракет игрока
- У игрока есть `MissileBase` с нужным уровнем

**Backend** — добавить в ответ `missile_range int` (0 = нет ракетного оружия).

**Frontend:**

```tsx
// Радиус вычисляется как уровень MissileBase × 5 систем
{sys?.missile_range > 0 && !isOwn && cell.has_planet &&
  Math.abs(cell_system - currentSystem) <= sys.missile_range && (
  <button onClick={() => sendRocket(cell.planet_id)} title="Ракетная атака">
    🚀
  </button>
)}
```

**Файлы:**
- `backend/internal/galaxy/repository.go` — вычислить missile_range
- `frontend/src/features/galaxy/GalaxyScreen.tsx`

---

### Задача 7 — Расход водорода за просмотр чужой системы [P3]

**Описание:** При загрузке чужой системы списывать −10 водорода с текущей планеты
(как в legacy). Не списывать при просмотре собственной системы.

**Backend** — добавить endpoint или сделать в рамках `GET /api/galaxy/{g}/{s}`:

```go
// В handler.go:
if g != homePlanet.Galaxy || s != homePlanet.System {
    if err := planetSvc.ConsumeHydrogen(ctx, userID, 10); err != nil {
        // не блокировать просмотр при нехватке — просто логируем
        slog.Warn("galaxy: hydrogen consume failed", "err", err)
    }
}
```

**Файлы:**
- `backend/internal/galaxy/handler.go`
- `backend/internal/planet/service.go` — метод ConsumeHydrogen

---

### Задача 8 — Экспедиция (16-я позиция) [P3]

**Описание:** Ссылка за пределы системы под таблицей, если экспедиции включены.

**Frontend:**

```tsx
{EXPEDITION_ENABLED && (
  <tr className="expedition-row">
    <td colSpan={7}>
      <a onClick={() => onFleetMission?.(g, s, 16, false, MISSION_EXPEDITION)}>
        🚀 Экспедиция — за пределы системы
      </a>
    </td>
  </tr>
)}
```

**Файлы:**
- `frontend/src/features/galaxy/GalaxyScreen.tsx`

---

## Приоритеты

| # | Задача | Приоритет | Усилие |
|---|--------|-----------|--------|
| 1 | Расширить CellView в backend (активность, альянс, статусы, луна) | ✅ | M |
| 2 | Альянс и активность в таблице | ✅ | S |
| 3 | Статусы игрока и легенда в footer | ✅ | S |
| 4 | Tooltip луны и обломков | ✅ | S |
| 5 | Star Surveillance (мониторинг флота) | P2 | M |
| 6 | Ракетная атака | P2 | M |
| 7 | Расход водорода за просмотр | P3 | S |
| 8 | Экспедиция (16-я позиция) | P3 | S |

---

## Намеренные упрощения vs legacy

| Функция | Причина |
|---------|---------|
| Ссылки на страницу альянса | Раздел альянсов не реализован — откладываем |
| Запрос дружбы (buddy request) | Система друзей не реализована |
| Боевой симулятор в галактике | Отдельный экран, не в галактике |
| Цветовые теги (друг/враг/конфедерация) | Требует системы отношений между игроками |
| AJAX мониторинг планеты (iframe) | Заменить на отдельный экран или модал |

---

## Файлы, которые будут изменены

| Файл | Изменение |
|------|-----------|
| `backend/internal/galaxy/repository.go` | ✏️ Расширить CellView, JOIN с альянсом/банами/луной |
| `backend/internal/galaxy/handler.go` | ✏️ Передать userID, добавить can_monitor/missile_range |
| `backend/internal/planet/service.go` | ✏️ Метод ConsumeHydrogen (задача 7) |
| `api/openapi.yaml` | ✏️ Расширить схему GalaxyCell и SystemView |
| `frontend/src/api/types.ts` | ✏️ Обновить интерфейс GalaxyCell |
| `frontend/src/features/galaxy/GalaxyScreen.tsx` | ✏️ Новые колонки, tooltip'ы, кнопки, легенда |
