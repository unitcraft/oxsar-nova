# План 14: Экран ремонта кораблей — доработка RepairScreen

## Статус: В РАЗРАБОТКЕ (задача 1 выполнена; задачи 2–4 — P2)

## Контекст

Legacy страница `/game.php/Repair` — ремонт повреждённых кораблей и разбор здоровых
юнитов на ресурсы. Включает очередь ремонтного ангара, список юнитов с расчётом
стоимости, progress-bar занятости хранилища.

**Legacy источник:** `d:\Sources\oxsar2\www\templates\standard\repair.tpl` + `d:\Sources\oxsar2\www\ext\page\ExtRepair.class.php`

**Legacy UI:** http://localhost:8080/game.php/Repair (логин test / пароль quoYaMe1wHo4xaci)

**Nova экран:** `frontend/src/features/repair/RepairScreen.tsx`

**Nova backend:** `backend/internal/repair/`

---

## Сравнение: legacy vs nova

### 1. Что есть в nova (реализовано)

| Функция | Статус |
|---------|--------|
| Две очереди (repair + disassemble) с refetchInterval | ✅ |
| Список повреждённых юнитов (GET /repair/damaged) | ✅ |
| Табы «Ремонт» / «Разбор» | ✅ |
| Карточки юнитов: иконка, имя, количество, повреждения | ✅ |
| ProgressBar shell_percent (danger/warning) | ✅ |
| Кнопка «Починить все» → POST /repair/repair | ✅ |
| Разбор: input количества + расчёт рефунда (×0.7) | ✅ |
| Toast-уведомления об успехе/ошибке | ✅ |
| Query invalidation на мутацию | ✅ |
| Формулы legacy: scalePerUnit, ceil10, multiplyCost | ✅ |

### 2. Что есть в legacy, но отсутствует в nova

| Функция | Приоритет | Комментарий |
|---------|-----------|-------------|
| **Отмена задания в очереди (cancel)** | ✅ | DELETE /repair/queue/{queueId} + кнопка ✕ |
| **Хранилище ремонта (repair_storage)** | P2 | Progress-bar занятости в шапке, в nova не отображается |
| **Требования на карточке юнита** | P2 | Бэк проверяет, фронт не показывает причину блокировки |
| **Скорость и топливо для кораблей** | P2 | Данные есть в configs/ships.yml, не показываются |
| **Фильтр «Доступные / Все»** | P3 | Скрывать юниты без выполненных requirements |
| **Описание юнита** | P3 | Legacy показывает description на карточке |
| **VIP-ускорение (за кредиты)** | P4 | Намеренно пропускаем |
| **Пакеты текстур (image packages)** | P4 | Не актуально |

### 3. Данные есть в каталоге, но не показываются в UI

| Поле | Где хранится | Статус |
|------|-------------|--------|
| `speed` | configs/ships.yml | ❌ не показывается |
| `fuel` | configs/ships.yml | ❌ не показывается |
| `shell` (макс. прочность) | configs/ships.yml | ❌ показывается только % |

---

## Формулы legacy (зафиксировать в коде)

### Разбор (disassemble)

```
required = ceil(base_cost * 0.2 / 10) * 10   // списывается при постановке в очередь
return   = ceil(base_cost * 0.9 / 10) * 10   // возвращается при завершении
earn     = return - required ≈ base_cost * 0.7
duration = buildTime(base * 0.1, base * 0.1) // время на 1 юнит
```

### Ремонт (repair)

```
struct_scale         = 0.1 * (100 - shell_percent) / 100
required_{m,s,h}     = ceil(base_cost * struct_scale / 10) * 10
duration             = buildTime(base * 0.1, base * 0.1) // время на 1 юнит
```

---

## Структура legacy (что передаётся в шаблон)

```
repair_storage:
  storage  — макс. мест в ремонтном хранилище
  used     — занято
  free     — свободно

events[]:           — очередь заданий
  eventid, name, quantity, cancel_link, vip_link
  event_pb_value    — прогресс для JS progressbar()

units[]:            — корабли и оборонительные
  unit_id, name, icon
  dock_capacity     — макс. вместимость в хранилище
  max_dock_units    — с учётом свободного места
  required_{m,s,h}  — стоимость ремонта/разбора
  earn_{m,s,h}      — прибыль при разборе
  prod_time         — время на 1 юнит
  can_build         — bool (requirements выполнены)
  required_constructions — текст если не выполнены
```

---

## План реализации

### Задача 1 — Отмена задания в очереди [P1]

**Backend:**

```go
// backend/internal/repair/service.go
func (s *Service) Cancel(ctx context.Context, userID, queueID string) error {
    // 1. Читать repair_queue FOR UPDATE
    // 2. Проверить ownership (user_id == userID)
    // 3. Если mode=disassemble → вернуть ресурсы + юниты на планету
    // 4. Если mode=repair → вернуть только ресурсы
    // 5. Удалить событие из events (по payload queue_id)
    // 6. Удалить запись из repair_queue
    // 7. Если задание было активным → перепланировать следующее в очереди
}
```

```go
// backend/internal/repair/handler.go
// DELETE /api/planets/{planetId}/repair/queue/{queueId}
func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request)
```

**OpenAPI** — добавить endpoint:
```yaml
/api/planets/{id}/repair/queue/{queueId}:
  delete:
    summary: Cancel a repair/disassemble queue item
    responses:
      204: No Content
```

**Frontend:**

```tsx
// RepairScreen.tsx — добавить кнопку ✕ в строку очереди
const cancelMutation = useMutation({
  mutationFn: (queueId: string) =>
    api.delete(`/api/planets/${planet.id}/repair/queue/${queueId}`),
  onSuccess: () => {
    void qc.invalidateQueries({ queryKey: ['repair-queue', planet.id] });
    void qc.invalidateQueries({ queryKey: ['planets'] });
    toast.show('success', 'Задание отменено');
  },
});
```

**Требования к UI:**
- Современный вид — кнопка отмены маленькая, иконка ✕, появляется при hover
- Не должно быть confirm-диалога — отмена быстрая, ресурсы возвращаются

---

### Задача 2 — Хранилище ремонта в шапке [P2]

**Backend** — добавить поля в GET /repair/queue ответ:

```go
type QueueResponse struct {
    Queue   []RepairQueueItem `json:"queue"`
    Storage RepairStorage     `json:"storage"`
}

type RepairStorage struct {
    Total int `json:"total"` // уровень repair_factory × capacity_per_level
    Used  int `json:"used"`  // сумма count по активным заданиям в очереди
    Free  int `json:"free"`  // total - used
}
```

**Frontend** — шапка над очередью:

```
Ремонтный ангар  [████████░░░░░░░] 8 / 20 мест
```

- Progress-bar с процентом занятости
- Цвет: зелёный < 50%, жёлтый 50-80%, красный > 80%
- Современный вид в стиле `ox-*` CSS-переменных

---

### Задача 3 — Скорость и топливо на карточке корабля [P2]

Данные уже есть в `configs/ships.yml` и каталоге. Добавить в карточку юнита при режиме «Разбор»:

```
[Крейсер]  В наличии: 5
           🚀 10 000   ⛽ 200/ед.
           Металл +1 400  Кремний +490
```

- Показывать только для кораблей (is_defense = false)
- Для оборонительных сооружений скрывать (у них нет speed/fuel)

---

### Задача 4 — Требования на карточке (🔒) [P2]

Если `can_repair = false` (требования не выполнены):

```
[Линкор]  В наличии: 2  Повреждено: 1
          🔒 Требуется: Верфь ур. 7
```

**Backend** — добавить в GET /repair/damaged ответ поле `requirements_met bool` и `requirements_text string`.

**Frontend** — заменить кнопку «Починить» на задизейбленную с тултипом о причине.

---

## Намеренные упрощения vs legacy (не реализуем)

| Функция | Причина |
|---------|---------|
| VIP-ускорение за кредиты | Система кредитов не приоритетна (план 11 пропущен) |
| Пакеты текстур (image packages) | Устаревшая механика, не нужна |
| Отдельная страница /Disassemble | В nova объединено в одном экране с табами — лучше UX |

---

## Приоритеты

| # | Задача | Приоритет | Усилие |
|---|--------|-----------|--------|
| 1 | Отмена задания в очереди | ✅ | M (backend + frontend) |
| 2 | Хранилище ремонта в шапке | P2 | S (backend поле + frontend ui) |
| 3 | Скорость и топливо на карточке | P2 | S (только frontend) |
| 4 | Требования на карточке | P2 | M (backend поле + frontend ui) |

**Требования к UI (все задачи):**
- Современный, минималистичный дизайн в стиле `ox-*`
- Touch-friendly: кнопки достаточно крупные
- Визуальный фидбек: disabled + spinner пока идёт запрос

---

## Файлы, которые будут изменены

| Файл | Изменение |
|------|-----------|
| `backend/internal/repair/service.go` | ✏️ Добавить Cancel(), поля Storage и requirements |
| `backend/internal/repair/handler.go` | ✏️ DELETE /repair/queue/{id}, расширить GET /queue |
| `api/openapi.yaml` | ✏️ Добавить DELETE endpoint, обновить схемы |
| `frontend/src/features/repair/RepairScreen.tsx` | ✏️ Кнопка отмены, storage шапка, требования |
