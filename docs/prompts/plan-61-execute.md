# Промпт: выполнить план 61 (admin-bff httputil.ReverseProxy.Director → Rewrite)

**Дата создания**: 2026-04-28
**План**: [docs/plans/61-admin-bff-rewrite-migration.md](../plans/61-admin-bff-rewrite-migration.md)
**Применение**: вставить блок ниже в новую сессию Claude Code в
рабочей директории `d:\Sources\oxsar-nova`. Агент прочитает план 61
самостоятельно и выполнит точечный рефактор + тесты.
**Объём**: 1 коммит, ~50-80 строк изменений, 30-60 минут работы.

---

```
Задача: выполнить план 61 — миграция admin-bff с deprecated
httputil.ReverseProxy.Director на современный Rewrite API.

ВАЖНОЕ:
- Это ТОЧЕЧНЫЙ рефактор уже работающего кода. Не переписывай
  больше, чем требует план.
- Не блокирует ничего, можно работать независимо. Параллельно
  идёт большой агент по плану 62 — будь аккуратен с git
  (поимённый add).

ПЕРЕД НАЧАЛОМ:

1) git status --short — если есть чужие изменения от параллельных
   сессий, спроси пользователя. Бери только свои файлы (admin-bff).

2) Прочитай ПОЛНОСТЬЮ:
   - docs/plans/61-admin-bff-rewrite-migration.md (твоё ТЗ —
     вся реализация уже расписана пошагово, включая код-сниппеты
     и тестовые сценарии)
   - CLAUDE.md (правила: коммиты, языки, hooks, контекст ctx)

3) Прочитай выборочно по мере необходимости:
   - projects/admin-bff/internal/proxy/proxy.go (текущая реализация
     с Director — то, что нужно мигрировать)
   - projects/admin-bff/internal/proxy/proxy_test.go (если уже
     существует — расширяешь; если нет — создаёшь минимальный)
   - projects/admin-bff/internal/handler/ (для понимания, как
     SessionFromContext работает)
   - docs/plans/53-admin-frontend.md (контекст: admin-bff живёт
     внутри 53, ты делаешь точечный фикс уже закрытого плана)

ЧТО НУЖНО СДЕЛАТЬ (по фазам плана 61):

Ф.1. Рефактор proxy.go
  - Заменить Director-блок (строки 36-45) на Rewrite-блок согласно
    шаблону из плана §"Что меняем", п.1+п.2.
  - Использовать httputil.ProxyRequest API (Go 1.20+, у нас 1.23).
  - SetURL(u) + SetXForwarded() — безопасные defaults.
  - Чистить r.Out.Header: Cookie, X-CSRF-Token, Authorization
    (последний копируется явно из In, см. план §2).
  - ErrorHandler вынести в отдельное поле структуры.
  - Тип Upstream и публичный API НЕ менять (NewUpstream / Handler /
    MatchPrefix остаются как есть).

Ф.2. Тесты
  - Если proxy_test.go нет — создать минимальный с httptest.NewServer
    как upstream-mock.
  - Кейсы (см. план §3):
    1) Клиент шлёт X-Forwarded-For: evil → upstream получает свой
       (на основе RemoteAddr), не клиентский.
    2) Клиент шлёт Authorization: Bearer evil → upstream получает
       токен из server-side session, не клиентский.
    3) Клиент шлёт Cookie: session=... → upstream получает без Cookie.
    4) Клиент шлёт X-CSRF-Token: ... → upstream получает без него.
    5) Базовый smoke: GET с правильной session → upstream получает
       Authorization + правильные заголовки.
  - go test ./internal/proxy/... — все зелёные.
  - go vet ./... в admin-bff — чисто (deprecation warning исчез).

Ф.3. Smoke-тест в Docker (опционально, если dev-stack доступен)
  - docker-compose up admin-bff (если работает локально).
  - Логин через admin-frontend, проверить что админские страницы
    работают (Users, Roles, Audit, Game-ops, Billing).
  - Проверить через DevTools / curl: X-Forwarded-For в upstream =
    IP клиента, не подделка.
  - Если dev-stack не запускается — пропустить эту фазу,
    отметить в коммит-сообщении.

Ф.4. Финализация
  - Обновить шапку плана 61 — статус "Завершён <дата>".
  - Запись в docs/project-creation.txt — итерация 61.
  - Коммит: refactor(admin-bff): миграция Director → Rewrite (план 61).

ВАЖНОЕ: GIT-ИЗОЛЯЦИЯ ОТ ПАРАЛЛЕЛЬНЫХ СЕССИЙ:

Параллельно крупный агент работает по плану 62 в
docs/research/origin-vs-nova/. Чтобы не пересечься:

- git status --short ПЕРЕД каждым git add
- git add ТОЛЬКО конкретные пути:
  · projects/admin-bff/internal/proxy/proxy.go
  · projects/admin-bff/internal/proxy/proxy_test.go (если новый)
  · docs/plans/61-admin-bff-rewrite-migration.md
  · docs/project-creation.txt
- git status --short ПЕРЕД git commit (убедись что не захватил
  чужое из docs/research/...)
- Если случайно staged чужой файл: git reset HEAD -- путь
- НИКОГДА не делай git add . или git add -A

КОММИТ:
- conventional: refactor(admin-bff): миграция Director → Rewrite (план 61)
- В теле сообщения упомянуть:
  · что снимает deprecation warning Go 1.20+
  · что закрывает header-smuggling gap (X-Forwarded-For подделка)
  · ссылка на план 61
- trailer: Generated-with: Claude Code (НЕ Co-Authored-By —
  git hook уберёт автоматически, см. CLAUDE.md Onboarding)

ЧЕГО НЕ ДЕЛАТЬ:

- Не менять публичный API Upstream / NewUpstream / Handler /
  MatchPrefix — только внутреннюю реализацию.
- Не вводить новые заголовки в Out (X-Real-IP и т.п.) — это
  отдельная задача если понадобится.
- Не трогать admin-bff routes — только proxy-инфраструктуру.
- Не делать "пока я тут, заодно подправлю" — рефакторинг по плану
  61 и всё.
- Не писать "возможно", "вероятно" — каждое решение должно быть
  обосновано планом 61 или Go-документацией.

ОЦЕНКА ОБЪЁМА:

30-60 минут работы. Если идёт сильно дольше — что-то пошло не так,
вернись и перечитай план §"Что меняем".

ИЗВЕСТНЫЕ РИСКИ (см. план 61, секция "Известные риски"):

- Rewrite не пробрасывает какой-то заголовок ожидаемо → тесты Ф.2 + smoke Ф.3.
- Authorization-инжекция сломалась → явная копия In.Authorization →
  Out.Authorization в Rewrite + тест.
- Backend-сервис ожидает X-Forwarded-Proto: https всегда → после
  SetXForwarded() при необходимости перезаписать
  r.Out.Header.Set("X-Forwarded-Proto", "https"). Решить по
  результатам тестов, не превентивно.

УСПЕШНЫЙ ИСХОД:

- proxy.go использует Rewrite, не Director.
- go vet чисто, deprecation warning исчез.
- Тесты в proxy_test.go проходят все 5 кейсов.
- Header-smuggling от клиента невозможен (X-Forwarded-* подделать
  нельзя — SetXForwarded() их пересоздаёт безопасно).
- 1 коммит, шапка плана 61 помечена ✅.

Стартуй.
```
