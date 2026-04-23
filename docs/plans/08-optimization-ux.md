# План: Оптимизация производительности и улучшение UX

Дата: 2026-04-23  
Приоритет: M (после базовой функциональности)

## Цель

Улучшить производительность приложения, оптимизировать сетевые запросы, 
улучшить UX через feedback и микроинтерфейсы.

---

## 1. Frontend Оптимизация

### 1.1 Оптимистичные обновления (Optimistic Updates)

**Задача**: При отправке запроса на сервер, сразу обновить UI локально,
затем синхронизировать.

**Где**: 
- Переименование планеты (PATCH /planets/{id})
- Управление факторами производства (POST /planets/{id}/resource-update)
- Создание лота на рынке (POST /market/lots)
- Активация артефакта (POST /artefacts/{id}/activate)

**Реализация**:
- TanStack Query `useMutation` → добавить `onMutate` для optimistic update
- На ошибку — rollback через `onError`
- Пример: при PATCH планеты `setQueryData([...], updatedPlanet)` перед запросом

### 1.2 Lazy Loading и код-сплиттинг

**Текущее состояние**: Все экраны уже lazy-loaded через `React.lazy()`.

**Проверить**: Размеры бандлов (npm run build → проверить dist/).

### 1.3 Кеширование на клиенте (Request Deduplication)

**Задача**: Не отправлять повторные запросы за N секунд.

**Где**:
- `/api/planets/{id}/buildings/queue` — refetchInterval: 30s (текущий интервал)
- `/api/planets/{id}/resource-report` — staleTime: 60s перед обновлением

**Текущее**: useQuery с refetchInterval уже установлены, проверить оптимальные значения.

---

## 2. Backend Оптимизация

### 2.1 Индексы БД

**Проверить существующие индексы**:
```sql
\d+ buildings    -- planet_id, unit_id
\d+ planets      -- user_id, destroyed_at
\d+ events       -- user_id, created_at, kind
```

**Добавить если нет**:
- `buildings(planet_id, unit_id)` - для быстрого поиска одного здания
- `planets(user_id, destroyed_at)` - для ListByUser (уже есть?)
- `planets(galaxy, system, position, is_moon)` - для galaxy-queries

### 2.2 Батч-операции

**Задача**: Уменьшить кол-во round-trips в БД.

**Где**:
- Получение зданий + уровней исследований в одном запросе (вместо отдельных)
- Обновление нескольких факторов в одном UPSERT

**Текущее состояние**: UpdateResourceFactors делает N запросов (по одному на здание).
**Оптимизация**: Использовать CASE/WHEN в одном UPDATE:
```sql
UPDATE buildings SET production_factor = CASE 
  WHEN unit_id = 1 THEN 50
  WHEN unit_id = 2 THEN 75
  ELSE production_factor
END
WHERE planet_id = $1 AND unit_id = ANY($2::int[]);
```

### 2.3 Query Profiling

**Задача**: Найти медленные запросы.

**Инструмент**: PostgreSQL `log_min_duration_statement = 100` (логирует > 100ms).

**Проверить**:
- `/api/planets` для пользователя с 50+ планетами (ListByUser)
- `/api/highscore` (сортировка, лимит 100)
- `/api/battle-reports/{id}` (LEFT JOIN x5?)

---

## 3. UX улучшения

### 3.1 Прогресс-индикаторы

**Добавить**:
- Skeleton loaders вместо пустого экрана при загрузке (уже частично есть)
- Плавная загрузка данных пошагово (например, сначала шапка, потом таблица)

### 3.2 Уведомления (Toasts)

**Проверить**: Toast компонент уже есть в `ui/Toast.tsx`.

**Где использовать**:
- Успех: «Планета переименована»
- Ошибка: «Не удалось сохранить изменения»
- Info: «Загрузка отменена»

### 3.3 Валидация на фронте (optimistic)

**Добавить валидацию перед отправкой**:
- Имя планеты: 1–50 символов, не пусто после trim
- Цена лота: > 0
- Фактор производства: 0–100

**Инструмент**: `zod` или встроенная валидация.

### 3.4 Keyboard shortcuts

**Добавить быстрые клавиши**:
- `Esc` — закрыть модальное окно
- `Tab` — переключение между экранами
- `Ctrl+S` — сохранить (если есть unsaved changes)

---

## 4. Мониторинг и логирование

### 4.1 Frontend error tracking

**Инструмент**: ErrorBoundary + Sentry / LogRocket (опционально).

**Где реализовать**:
- Error boundary в App.tsx
- Отлов необработанных ошибок promise (unhandledrejection)

### 4.2 Performance metrics

**Добавить**: Web Vitals (LCP, FID, CLS).

**Инструмент**: `web-vitals` npm package.

### 4.3 Backend logs

**Текущее**: `log/slog` с JSON-форматом.

**Улучшение**: Добавить трассировку (trace-id) для корреляции запросов.

---

## 5. Security audit

### 5.1 CSRF protection

**Проверить**: CORS headers, SameSite cookie policy.

### 5.2 Rate limiting

**Текущее**: `/api/auth/*` — 20 req/min per IP.

**Добавить** (опционально):
- API endpoints — более мягкий лимит (100 req/min per user)
- Защита от перебора параметров (fuzzing)

### 5.3 Input validation

**Проверить**: Все POST/PATCH/PUT endpoints валидируют input.

---

## 6. Documentation

### 6.1 API documentation (OpenAPI)

**Текущее**: `api/openapi.yaml` содержит основные endpoints.

**Добавить**:
- Примеры request/response для каждого endpoint
- Описание rate limits в документации
- Changelog версий API

### 6.2 Frontend component library

**Задача**: Документировать переиспользуемые компоненты (Button, Card, Modal и т.д.).

**Где**: `frontend/src/ui/` — добавить .stories.tsx (Storybook) или README.

---

## Статус реализации

| Пункт | Статус | Комментарий |
|-------|--------|-----------|
| Optimistic Updates | 🟡 | Реализовано для MarketScreen, НЕ реализовано для ResourceScreen |
| Lazy Loading | ✅ | Все экраны lazy-loaded |
| Кеширование | ✅ | useQuery с staleTime/refetchInterval |
| Индексы БД | 🟡 | Проверить наличие |
| Батч-операции | ✅ | UpdateResourceFactors оптимизирован |
| Query Profiling | ⬜ | Не начинали |
| Skeleton loaders | ✅ | Улучшенные скелетоны для всех экранов |
| Toasts | ✅ | Toast provider уже есть |
| Валидация на фронте | 🟡 | Есть, можно расширить |
| Keyboard shortcuts | ⬜ | Не реализовано |
| Error tracking | 🟡 | Есть ErrorBoundary, можно добавить Sentry |
| Metrics | ✅ | Web Vitals реализованы |
| CSRF protection | ✅ | CORS с AllowCredentials настроены |
| Rate limiting | ✅ | Redis-based limiter (20 req/min auth) |
| Input validation | ✅ | Реализовано на backend |
| OpenAPI docs | ✅ | Основные endpoints задокументированы |

---

## Приоритизация и выполнение

**High (сразу) — ✅ ЗАВЕРШЕНО**:
- ✅ 2.2 Батч-операции (UpdateResourceFactors с CASE/WHEN)
- ✅ 4.2 Performance metrics (Web Vitals с Performance API)

**Medium (этот спринт) — ✅ ЗАВЕРШЕНО**:
- ✅ 1.1 Optimistic Updates (ResourceScreen + MarketScreen)
- ✅ 3.1 Skeleton loaders (улучшенные типизированные скелетоны)
- ✅ 5.1 Security audit (CORS + rate limiting подтверждены)

**Low (P3) — ЧАСТИЧНО**:
- ✅ 3.4 Keyboard shortcuts (Alt+H/B/R/M, Esc, Ctrl+S)
- ⬜ 4.1 Error tracking (Sentry integration) — не делаем, нет аккаунта
- ⬜ 6.2 Storybook (UI компоненты) — не делаем, нецелесообразно без дизайн-системы
- ✅ 1.1 Optimistic Updates для ArtefactsScreen/OfficersScreen (2026-04-23)
