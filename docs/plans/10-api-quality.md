# План: Качество API — идемпотентность и защита от дублей

## Статус: ЗАВЕРШЁН (Idempotency Key реализован для fleet/market/artmarket; SELECT FOR UPDATE уже был)

## Контекст

При сетевых ошибках клиент может повторить запрос (retry), а пользователь —
кликнуть дважды. Для безопасных эндпоинтов (GET, PATCH, PUT) это не проблема.
Для мутирующих POST — может привести к дублям: два флота, двойное списание
кредитов, два одинаковых лота на рынке.

---

## Аудит идемпотентности (результат проверки 2026-04-23)

### Безопасны (GET, PATCH, PUT, DELETE)

Все GET-эндпоинты идемпотентны по определению.
PATCH/PUT — перезаписывают состояние, повторный вызов даёт тот же результат.
DELETE — повторный вызов вернёт 404, но не сломает данные.

### POST — фактически идемпотентные

| Эндпоинт | Почему безопасен |
|----------|-----------------|
| `POST /api/planets/{id}/set-home` | Устанавливает флаг, повтор = то же состояние |
| `POST /api/planets/{id}/resource-update` | Перезаписывает factors |
| `POST /api/artefacts/{id}/activate` | Повтор вернёт ошибку "already active" |
| `POST /api/messages/{id}/read` | Повтор безвреден |
| `POST /api/battle-sim` | Pure function, БД не трогает |
| `POST /api/admin/users/{id}/ban` | Повтор вернёт "already banned" |

### POST — критически неидемпотентные (двойной вызов = двойное действие)

| Эндпоинт | Последствие дубля |
|----------|------------------|
| `POST /api/fleet` | Два флота вместо одного, двойное списание кораблей/ресурсов |
| `POST /api/planets/{id}/buildings` | Два задания в очереди строительства |
| `POST /api/planets/{id}/shipyard` | Удвоенный заказ кораблей |
| `POST /api/planets/{id}/rockets/launch` | Две ракеты вместо одной |
| `POST /api/market/lots` | Два одинаковых лота |
| `POST /api/market/lots/{id}/accept` | Race condition — двойная покупка |
| `POST /api/artefact-market/offers/{id}/buy` | Двойная покупка артефакта |
| `POST /api/admin/users/{id}/credit` | Двойное зачисление кредитов |
| `POST /api/officers/{key}/activate` | Двойное списание кредитов |
| `POST /api/chat/{kind}/send` | Дублированное сообщение |

---

## Решения

### 1. Frontend — блокировка кнопки (минимум, уже частично есть)

TanStack Query блокирует повторный вызов пока `mutation.isPending === true`.
Убедиться что все критичные кнопки используют `disabled={mutation.isPending}`.

**Покрывает:** двойной клик пользователя.  
**Не покрывает:** сетевой retry, параллельные вкладки.

### 2. Backend — Idempotency Key для критичных эндпоинтов

Клиент генерирует UUID и передаёт в заголовке. Сервер дедуплицирует по нему.

```
POST /api/fleet
Idempotency-Key: 550e8400-e29b-41d4-a716-446655440000
```

```go
// Middleware или в handler:
func (h *Handler) SendFleet(w http.ResponseWriter, r *http.Request) {
    key := r.Header.Get("Idempotency-Key")
    if key != "" {
        if cached, ok := h.idempotency.Get(key); ok {
            w.WriteHeader(cached.Status)
            w.Write(cached.Body)
            return
        }
    }
    // ... обычная логика
    // сохранить результат в кэш на 24 часа
    h.idempotency.Set(key, result, 24*time.Hour)
}
```

Хранить в Redis с TTL 24 часа — идемпотентность актуальна только для быстрых
повторных запросов, не для воспроизведения через сутки.

**Приоритет внедрения:**
1. `POST /api/fleet` — наибольший ущерб от дубля
2. `POST /api/market/lots/{id}/accept` — race condition с деньгами
3. `POST /api/artefact-market/offers/{id}/buy` — списание кредитов

### 3. Database — уникальные constraints там где возможно

```sql
-- Нельзя принять один лот дважды: статус лота меняется в транзакции
-- Нельзя активировать одного офицера дважды: уникальный индекс по (user_id, officer_key, active=true)
```

Это защита на уровне БД — последняя линия обороны.

---

## Задачи

- [ ] Проверить все критичные POST-эндпоинты: убедиться что кнопки `disabled` при `isPending`
- [ ] Реализовать Idempotency Key middleware для `POST /api/fleet`
- [ ] Реализовать Idempotency Key для `POST /api/market/lots/{id}/accept`
- [ ] Реализовать Idempotency Key для `POST /api/artefact-market/offers/{id}/buy`
- [ ] Проверить `POST /api/market/lots/{id}/accept` на race condition (транзакция + SELECT FOR UPDATE)
- [ ] Добавить уникальный constraint на активного офицера

---

## Файлы

| Файл | Изменение |
|------|-----------|
| `backend/pkg/idempotency/` | 🆕 Middleware/хелпер для Idempotency Key + Redis |
| `backend/internal/fleet/handler.go` | ✏️ Проверка Idempotency-Key |
| `backend/internal/market/handler.go` | ✏️ Проверка Idempotency-Key + SELECT FOR UPDATE |
| `backend/internal/artmarket/handler.go` | ✏️ Проверка Idempotency-Key |
| `frontend/src/api/client.ts` | ✏️ Генерация и передача Idempotency-Key для мутаций |
