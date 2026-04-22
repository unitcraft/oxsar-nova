# План объединения систем: Tutorial → Achievement

## Контекст

В кодовой базе две отдельные системы:

1. **Achievement** (`backend/internal/achievement/`) — пассивные достижения
   - CheckAll() — проверяет условия и открывает достижения (FIRST_METAL, FIRST_WIN, SCORE_1000 и т.д.)
   - Таблицы: `achievement_defs` + `achievements_user`
   - UI: `AchievementsScreen.tsx` (уже готов, показывает достижения + прогресс)

2. **Tutorial** (`backend/internal/tutorial/`) — активные квест-цепочки (§5.20 ТЗ)
   - Многошаговые задачи с триггерами и наградами
   - Построй шахту → солнечное дерево → лабораторию → исследуй → корабль → миссия
   - Отключается в DEATHMATCH-режиме

**Проблема:** дублирование. Нужна одна **унифицированная система Достижений** с категориями.

## Новая архитектура

**Единая система `achievement`:**
- Расширить `achievement_defs` добавить колонку `category` (enum: `passive` | `starter`)
- Пассивные достижения (FIRST_METAL, BATTLE_10) — открываются CheckAll()
- Starter достижения (цепочка задач) — новый CheckAllStarter()
- UI одна вкладка "🏆 Достижения" с фильтрацией по категориям

---

## Шаги реализации

### 1. Миграция БД
Добавить в новую миграцию:
```sql
ALTER TABLE achievement_defs ADD COLUMN category TEXT DEFAULT 'passive';
-- Обновить старые ачивки:
UPDATE achievement_defs SET category='passive' WHERE key LIKE 'FIRST_%' OR key LIKE 'BATTLE_%' OR key LIKE 'FLEET_%' OR key LIKE 'SCORE_%';
-- Starter ачивки будут заполнены в шаге 2
```

### 2. `backend/internal/achievement/service.go` — добавить CheckAllStarter()
Новая функция с триггерами Starter-ачивок:
```go
// CheckAllStarter открывает стартовые ачивки (цепочка квестов).
func (s *Service) CheckAllStarter(ctx context.Context, userID string) error {
    type check struct {
        key string
        sql string
    }
    checks := []check{
        {"STARTER_BUILD_METALMINE", `
            SELECT EXISTS (
                SELECT 1 FROM buildings b
                JOIN planets p ON p.id = b.planet_id
                WHERE p.user_id = $1 AND b.unit_id = 1 AND b.level >= 1
            )`},
        {"STARTER_BUILD_SOLARPLANT", `
            SELECT EXISTS (
                SELECT 1 FROM buildings b
                JOIN planets p ON p.id = b.planet_id
                WHERE p.user_id = $1 AND b.unit_id = 3 AND b.level >= 1
            )`},
        // ... остальные триггеры из старого Tutorial
    }
    // Аналогично CheckAll(), вызвать UnlockIfNew для каждого триггера
}
```

Вызовы в Handler.List():
```go
h.svc.CheckAll(r.Context(), uid)        // пассивные
h.svc.CheckAllStarter(r.Context(), uid) // стартовые
```

### 3. Добавить Starter-ачивки в `achievement_defs`
Наполнить seed-миграцию:
```yaml
- key: STARTER_BUILD_METALMINE
  title: "Первая шахта"
  description: "Постройте шахту на своей планете"
  points: 10
  category: starter
  order: 1

- key: STARTER_BUILD_SOLARPLANT
  title: "Солнечная энергия"
  description: "Постройте солнечный растений на своей планете"
  points: 10
  category: starter
  order: 2

# ... и т.д. 6-7 шагов цепочки
```

### 4. Убрать пакет `backend/internal/tutorial/`
- Удалить `backend/internal/tutorial/` (service.go, handler.go, types.go)
- Удалить из `cmd/worker/main.go` инъекцию `tutorial.Service`
- Удалить из `cmd/server/main.go` роут `/api/tutorial` (если есть)

### 5. Обновить фронтенд
- `frontend/src/features/tutorial/TutorialScreen.tsx` → удалить
- `AchievementsScreen.tsx` — уже готов, добавить фильтр по категориям (опционально)
- UI показывает все ачивки в одном списке (пассивные + стартовые)

### 6. Документация
- `oxsar-spec.txt` — обновить §5.20 заголовок на «Достижения (полная система)»
- `docs/status.md` — объединить Tutorial ✅ + Achievement ✅ → Achievements ✅
- `project-creation.txt` — зафиксировать объединение

---

## Зависимости

- Миграция БД — добавить `category` колонку, populate Starter-ачивки
- achievement_defs должна быть seed-миграция, чтобы добавлять новые ачивки
- API `/api/achievements` уже существует, работает с новой структурой

## Что НЕ входит в этот план

- UI фильтрацию по категориям (добавится позже, если нужна)
- Сезонные / специальные ачивки (M8+)
- Analytics / leaderboards для ачивок
