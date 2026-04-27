# План 61: admin-bff — миграция httputil.ReverseProxy.Director → Rewrite

**Дата**: 2026-04-28
**Статус**: Активный
**Зависимости**: план 53 (admin-bff) — выполнен. Этот план — точечный
рефактор уже работающего кода.
**Связанные документы**:
- [53-admin-frontend.md](53-admin-frontend.md) — основной план admin-bff,
  где этот рефактор должен был учитываться, но был сделан старым API.

---

## Цель

Заменить устаревший `httputil.ReverseProxy.Director` на современный
`Rewrite` в `projects/admin-bff/internal/proxy/proxy.go`. Это:

1. **Снимает deprecation warning** в Go 1.20+ (мы на 1.23 по
   `CLAUDE.md`).
2. **Закрывает small security gap** — `Director` не делает auto-strip
   `X-Forwarded-*` headers из входящего запроса, что позволяет
   header-smuggling от клиента (передача фейкового
   `X-Forwarded-For: <вражеский-IP>` поверх admin-bff).

---

## Контекст

### Текущий код (deprecated API)

`projects/admin-bff/internal/proxy/proxy.go:36-45`:

```go
rp := httputil.NewSingleHostReverseProxy(u)
defaultDirector := rp.Director
rp.Director = func(req *http.Request) {
    defaultDirector(req)
    req.Header.Del("Cookie")
    req.Header.Del("X-CSRF-Token")
    req.Header.Set("X-Forwarded-Proto", "https")
}
```

Проблемы:

1. `Director` помечен `Deprecated: Use Rewrite instead.` начиная
   с Go 1.20. IDE подсвечивает зачёркнутым.
2. `defaultDirector(req)` — зависимость от внутренней реализации
   `NewSingleHostReverseProxy`.
3. **Нет auto-strip** `X-Forwarded-For`, `Forwarded`, `Authorization`
   из входящего запроса. Клиент может прислать
   `X-Forwarded-For: 8.8.8.8` — наш код этого не убирает, upstream
   увидит подделанный IP в логах/audit.
4. Auth-инжекция сделана отдельно в `Handler()` через
   `r.Header.Set("Authorization", ...)` — но `Director` не очищает
   `Authorization` от клиента, и пользователь теоретически мог бы
   подменить его до того, как `Handler` его перезапишет (на
   практике мы перезаписываем правильно, но это слабая
   защита-по-умолчанию).

### Что предлагает Go 1.20+

`Rewrite` принимает `*httputil.ProxyRequest` с двумя полями:
- `In` — оригинальный входящий запрос (read-only по соглашению);
- `Out` — запрос, который пойдёт upstream (тут модифицируем).

`SetURL(target)` устанавливает upstream-URL.
`SetXForwarded()` — **безопасно** устанавливает `X-Forwarded-Proto`,
`X-Forwarded-Host`, `X-Forwarded-For` на основе **только** входящего
запроса (`In.RemoteAddr`, `In.URL.Scheme`, `In.Host`). Старые
`X-Forwarded-*` из клиентского запроса **не пробрасываются**.

Также `Rewrite` (по документации) **по умолчанию** не пробрасывает
`Authorization` и `Forwarded` из `In` в `Out`.

---

## Что меняем

### 1. proxy.go — переход на Rewrite

```go
rp := &httputil.ReverseProxy{
    Rewrite: func(r *httputil.ProxyRequest) {
        r.SetURL(u)              // устанавливает Scheme/Host/Path для Out
        r.SetXForwarded()        // безопасный X-Forwarded-*
        // Чистим заголовки, которые backend не должен видеть
        r.Out.Header.Del("Cookie")
        r.Out.Header.Del("X-CSRF-Token")
        // Authorization устанавливается в Handler(), здесь —
        // явная гарантия что клиентский Authorization не
        // утекает (Rewrite по умолчанию не пробрасывает,
        // но явный Del не вредит).
        r.Out.Header.Del("Authorization")
    },
    ErrorHandler: func(w http.ResponseWriter, req *http.Request, err error) {
        slog.ErrorContext(req.Context(), "upstream error",
            slog.String("upstream", name),
            slog.String("err", err.Error()))
        http.Error(w, "upstream unavailable", http.StatusBadGateway)
    },
}
```

`url.Parse(target)` оставляем, тип `*Upstream` не меняется.

### 2. Handler() — порядок инжекции Authorization

После Rewrite Authorization сначала удаляется (Rewrite + явный
`Del`), потом в `Handler()` ставится из server-side session:

```go
func (u *Upstream) Handler() http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        sess, ok := handler.SessionFromContext(r.Context())
        if !ok {
            http.Error(w, "no session", http.StatusUnauthorized)
            return
        }
        r.Header.Set("Authorization", "Bearer "+sess.AccessToken)
        u.proxy.ServeHTTP(w, r)
    })
}
```

**Важно:** `r` здесь — это `In` для Rewrite. Rewrite получит
этот запрос с уже установленным правильным `Authorization` —
но в `Out` он не пробросится автоматически. Поэтому нужно
**явно** скопировать Authorization в `Out` внутри Rewrite:

```go
Rewrite: func(r *httputil.ProxyRequest) {
    r.SetURL(u)
    r.SetXForwarded()
    r.Out.Header.Del("Cookie")
    r.Out.Header.Del("X-CSRF-Token")
    // Копируем Authorization из In (его установил Handler) в Out.
    if auth := r.In.Header.Get("Authorization"); auth != "" {
        r.Out.Header.Set("Authorization", auth)
    }
},
```

Это правильный паттерн: server-side контролирует, какие заголовки
идут upstream; никаких неявных пробросов.

### 3. Тесты

Если в `proxy_test.go` уже есть тесты (проверить) — добавить
случаи:

- Клиент шлёт `X-Forwarded-For: evil` — upstream получает свой
  правильный (на основе `RemoteAddr`), не клиентский.
- Клиент шлёт `Authorization: Bearer evil` — upstream получает
  токен из server-side session (если установлен `Handler`'ом),
  не клиентский.
- Клиент шлёт `Cookie: session=...` — upstream получает запрос
  без Cookie.
- Клиент шлёт `X-CSRF-Token: ...` — upstream получает без него.
- Базовый smoke: GET с правильной session → upstream получает
  Authorization + правильные заголовки.

Если тестов нет — добавить минимальный `proxy_test.go` с
`httptest.NewServer` для upstream-mock.

### 4. Документация (минимально)

Если в `docs/ops/` или README admin-bff есть упоминание
архитектуры proxy — обновить (но скорее всего не требуется,
это внутренний детали реализации).

---

## Чего НЕ делаем

- **Не меняем публичный API** `Upstream` / `NewUpstream` /
  `Handler` / `MatchPrefix` — только внутреннюю реализацию.
- **Не вводим новые заголовки** в Out (например, `X-Real-IP`) —
  это отдельная задача если понадобится.
- **Не трогаем admin-bff routes** — только proxy-инфраструктуру.

---

## Этапы

### Ф.1. Рефактор proxy.go

- Заменить `Director`-блок на `Rewrite`-блок (см. выше).
- Убедиться, что `httputil.ProxyRequest` доступен (Go 1.20+, у нас
  1.23 — ✅).
- `go build ./...` в `projects/admin-bff/` — компилируется.
- `go vet ./...` — чисто (deprecation warning должен исчезнуть).

### Ф.2. Тесты

- Добавить (или расширить) `proxy_test.go` с кейсами из §3 выше.
- Запустить `go test ./internal/proxy/...` — все зелёные.

### Ф.3. Smoke-тест в Docker

- Запустить admin-bff локально через docker-compose (если есть
  dev-stack из плана 53).
- Залогиниться через admin-frontend.
- Проверить через DevTools / curl, что:
  - админские запросы идут (видны в логах upstream-сервисов);
  - `X-Forwarded-For` в upstream — это IP клиента, а не подделка.

### Ф.4. Финализация

- Обновить шапку плана 61 — отметить как ✅ с датой.
- Запись в `docs/project-creation.txt`: итерация 61.
- Коммит: `refactor(admin-bff): миграция Director → Rewrite (план 61)`.

---

## Тестирование

Помимо unit-тестов из Ф.2:

- Smoke в Docker по Ф.3.
- Регрессия: все админские страницы admin-frontend продолжают
  работать (Users, Roles, Audit, Game-ops, Billing).

---

## Объём

1 коммит, ~50–80 строк изменений (рефакторинг proxy.go + 1-2
unit-теста).

Время выполнения: **~30–60 минут агента**.

---

## Когда запускать

Не блокирует ничего, можно делать в любой свободный момент.

Не имеет смысла **до** закрытия плана 53 (admin-bff в активной
разработке) — но 53 уже почти закрыт, можно делать сейчас или
позже параллельно с любыми другими задачами.

---

## Известные риски

| Риск | Митигация |
|---|---|
| `Rewrite` API ведёт себя иначе чем ожидалось (например, не пробрасывает какой-то заголовок) | Тщательные тесты в Ф.2; smoke в Ф.3. |
| Сломали Authorization-инжекцию (proxy не получает токен) | Явная копия `In.Authorization → Out.Authorization` в Rewrite + тест |
| Какой-то заголовок неожиданно отсутствует в `Out` | По умолчанию Rewrite копирует все non-deprecated headers из In; deprecated (Forwarded, Authorization) — нет. Если что-то нужно — копируем явно. |
| Backend-сервис ожидает `X-Forwarded-Proto: https` всегда (даже когда admin-bff за HTTP) | После `SetXForwarded()` можно перезаписать: `r.Out.Header.Set("X-Forwarded-Proto", "https")`. Решить по результатам тестов. |

---

## Что после плана 61

- Современный, безопасный proxy-слой в admin-bff.
- Шаблон для будущих proxy-сервисов (если появится `portal-bff`,
  `mobile-bff` и т.п.) — сразу с правильной реализацией.
- Можно запустить статический анализатор (golangci-lint
  `staticcheck`) в CI — он перестанет ругаться на этот файл.

---

## References

- Go docs: [`net/http/httputil#ReverseProxy`](https://pkg.go.dev/net/http/httputil#ReverseProxy)
- Release notes Go 1.20:
  https://go.dev/doc/go1.20#netroutehttputilpkgnethttphttputil
- Security: header smuggling через `X-Forwarded-For`:
  OWASP Cheat Sheet on Reverse Proxy.
