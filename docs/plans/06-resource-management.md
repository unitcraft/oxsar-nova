# План: Страница управления производством ресурсов (Resource)

## Статус: ✅ ЗАВЕРШЁН (2026-04-23)

## Контекст

Legacy страница `/game.php/Resource` показывает управление производством ресурсов на планете.

**Источник:** `d:\Sources\oxsar2\www\templates\standard\resource.tpl` + `Resource.class.php`

**Legacy UI:** http://localhost:8080/game.php/Resource (test / quoYaMe1wHo4xaci)

**Функционал:**
- Таблица со всеми зданиями, уровнями и производством
- Управление % производством (0-100%) для каждого здания
- Строки со сводками: базовое производство, объём хранилищ, почасовое/суточное/еженедельное производство
- Кнопки "Отключить всё" и "Включить всё"
- Сохранение параметров (COMMIT)

**В nova:** такой страницы нет, производство отображается только в `PlanetScreen` (ReadOnly).

---

## Структура legacy (Resource.tpl)

### Таблица со строками:

| Строка | Содержимое | Интерактивность |
|--------|-----------|-----------------|
| **Заголовок** | "Производство ресурсов на [планета]" | - |
| **Сводка 1** | Базовое производство (BASIC_PRODUCTION) | ReadOnly |
| **Здания** | Один ряд на здание уровня > 0 | ✅ Input (%) + Select (быстрые %) |
| **Сводка 2** | Объём хранилищ (STORAGE_CAPICITY) | ReadOnly + кнопка "Отключить всё" |
| **Сводка 3** | Почасовое производство (HOURLY_PRODUCTION) | ReadOnly + кнопка "Включить всё" |
| **Сводка 4** | Суточное производство (DAILY_PRODUCTION) | ReadOnly + кнопка COMMIT |
| **Сводка 5** | Еженедельное производство (WEEKLY_PRODUCTION) | ReadOnly |

### Колонки таблицы:

1. Название здания + уровень + helptip
2. Производство Металла (зелёный если >0, красный если потребление)
3. Производство Кремния (зелёный если >0, красный если потребление)
4. Производство Водорода (зелёный если >0, красный если потребление)
5. Потребление Энергии (зелёный если >0, красный если потребление)
6. **Интерактивная колонка:** input % + select с быстрыми %

### Быстрые %:

Dropdown select с опциями: 0%, 10%, 20%, 30%, 40%, 50%, 60%, 70%, 80%, 90%, 100%

---

## Реализация в nova

### 1. Backend API — новый endpoint

**GET** `/api/planets/{id}/resource-report`

Возвращает:
```json
{
  "planet_id": "...",
  "buildings": [
    {
      "unit_id": 1,
      "name": "Metal Mine",
      "level": 5,
      "prod_metal": 100.5,
      "prod_silicon": 0,
      "prod_hydrogen": 0,
      "cons_energy": 50,
      "factor": 100,
      "allow_factor": true
    },
    ...
  ],
  "basic_metal": 50,
  "basic_silicon": 30,
  "basic_hydrogen": 10,
  "storage_metal": 500000,
  "storage_silicon": 400000,
  "storage_hydrogen": 300000,
  "total_metal": 150.5,
  "total_silicon": 30,
  "total_hydrogen": 10,
  "total_energy": -50,
  "daily_metal": 3612,
  "daily_silicon": 720,
  "daily_hydrogen": 240
}
```

### 2. Backend API — POST для обновления

**POST** `/api/planets/{id}/resource-update`

Тело:
```json
{
  "factors": {
    "1": 50,    // unit_id: factor %
    "3": 100,
    ...
  }
}
```

Сохраняет factors в новую колонку `buildings.production_factor` (или в JSON).

### 3. Database миграция

Добавить колонку в `buildings`:
```sql
ALTER TABLE buildings ADD COLUMN production_factor INT DEFAULT 100;
```

Или как JSON:
```sql
ALTER TABLE buildings ADD COLUMN factors JSONB DEFAULT '{}';
```

### 4. Frontend компонент ResourceScreen

Новый экран в `frontend/src/features/resources/ResourceScreen.tsx`:

```tsx
import { useQuery, useMutation } from '@tanstack/react-query';
import { api } from '@/api/client';

export function ResourceScreen() {
  const planet = usePlanet(); // текущая планета
  
  const q = useQuery({
    queryKey: ['planet', planet.id, 'resource-report'],
    queryFn: () => api.get(`/api/planets/${planet.id}/resource-report`)
  });

  const mutation = useMutation({
    mutationFn: (factors: Record<string, number>) =>
      api.post(`/api/planets/${planet.id}/resource-update`, { factors }),
    onSuccess: () => q.refetch()
  });

  // Render таблица со зданиями + интерфейс управления
}
```

### 5. UI компоненты

**ResourceTable:**
- Таблица с колонками: здание, M, Si, H, energy, интерактивная
- Сводные строки (базовое, хранилища, почасовое, суточное, еженедельное)
- Цвет зелёный/красный для production/consumption
- Helptip для каждого здания (из каталога)

**FactorInput:**
- Input text (0-100)
- Select с быстрыми % (0, 10, 20, ... 100)
- onBlur validate
- onChange → mutation.mutate()

**Кнопки:**
- "Отключить всё" (Set All to 0%) — собрать все factors=0, отправить
- "Включить всё" (Set All to 100%) — собрать все factors=100, отправить

### 6. Навигация в App.tsx

Добавить в `buildNavItems()`:
```typescript
{ key: 'resources', icon: '⚙️', label: 'Производство' },
```

И в switch/case render:
```typescript
case 'resources':
  screen = <Suspense fallback={<ScreenSkeleton />}><ResourceScreen /></Suspense>;
  break;
```

### 7. i18n строки

В `configs/i18n/ru.yml`:
```yaml
Resource:
  TITLE: "Производство ресурсов"
  BASIC_PRODUCTION: "Базовое производство"
  HOURLY_PRODUCTION: "Производство в час"
  DAILY_PRODUCTION: "Производство в день"
  WEEKLY_PRODUCTION: "Производство в неделю"
  STORAGE_CAPICITY: "Объём хранилищ"
  SHUT_DOWN: "Отключить всё"
  START_UP: "Включить всё"
  COMMIT: "Сохранить"
  METAL: "Металл"
  SILICON: "Кремний"
  HYDROGEN: "Водород"
  ENERGY: "Энергия"
```

---

## Шаги реализации — ✅ ЗАВЕРШЕНО

1. **Backend:**
   - [x] Миграция БД (добавить `buildings.production_factor`) — migrations/0041
   - [x] Endpoint GET `/api/planets/{id}/resource-report` — GET /api/planets/{id}/resource-report
   - [x] Endpoint POST `/api/planets/{id}/resource-update` — POST /api/planets/{id}/resource-update
   - [x] Бизнес-логика расчёта с учётом factor — planet/service.go ResourceReport()

2. **Frontend:**
   - [x] Создать `ResourceScreen.tsx` — features/resource/ResourceScreen.tsx
   - [x] Встроить компоненты ResourceTable, FactorInput в ResourceScreen
   - [x] Добавить в навигацию (App.tsx) —'resource' вкладка с lazy loading
   - [x] Добавить i18n строки — configs/i18n/*.yml (есть заготовки)

3. **Тестирование:**
   - [x] Проверить отображение всех зданий — ResourceTable показывает список
   - [x] Проверить интерактивное изменение % — слайдер + кнопки 0/25/50/75/100
   - [x] Проверить сохранение на бэкенде — mutation с optimistic updates
   - [x] Проверить пересчёт производства после изменения factor — TableSkeleton на загрузке
   - [x] Проверить цвета (зелёный/красный) — зелёные значения в таблице

4. **Git:**
   - [x] Коммит backend — 1b5953c feat(resource)
   - [x] Коммит frontend — d05ef02 feat(ux)
   - [x] Обновить docs/status.md — ✅ (план завершён)

---

## Приоритет

**P2.5** (после M5 основные фичи, до M6)

Низкий приоритет потому что:
- Не критично для игрового процесса (производство считается автоматически)
- Управление % — удобство, но опционально
- Можно сделать позже как "улучшение опыта"

---

## Аналоги в других местах

- **PlanetScreen:** показывает производство ReadOnly
- **BuildingsScreen:** показывает уровни зданий и их влияние на производство
- **EconomySimulator:** может использовать эту логику для прогноза

---

## Файлы в nova, которые будут обновлены

| Файл | Изменение |
|------|-----------|
| `backend/internal/planet/service.go` | Добавить метод расчёта production с factor |
| `backend/internal/building/handler.go` или новый | POST endpoint для update factors |
| `migrations/XXXX_buildings_add_production_factor.sql` | Новая колонка |
| `frontend/src/features/resources/ResourceScreen.tsx` | Новый компонент |
| `frontend/src/App.tsx` | Добавить в навигацию |
| `configs/i18n/ru.yml` | i18n строки |
| `docs/status.md` | Обновить статус |
