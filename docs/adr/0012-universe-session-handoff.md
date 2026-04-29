# ADR-0012: Universe session handoff (cross-domain SSO)

**Дата**: 2026-04-29
**Статус**: Принято
**План**: 72.2 ([72.2-universe-session-handoff.md](../plans/72.2-universe-session-handoff.md))

## Контекст

Архитектура oxsar-nova мультидоменная:
- `oxsar-nova.ru` — публичный портал.
- `uni01.oxsar-nova.ru`, `uni02.oxsar-nova.ru`, ... — отдельные
  поддомены под игровые вселенные (Nova, Speed, Classic).
- `admin.oxsar-nova.ru` — админ-консоль (отдельный flow, не влияет
  на этот ADR).

Юзер логинится на портале (identity-service выдаёт RSA-JWT, токены
хранятся в `localStorage` портала). Когда юзер кликает «Играть» в
карточке вселенной, он должен попасть на game-домен **уже залогиненным**.

`localStorage` per-origin: токены портала (`localhost:5174`) НЕ доступны
коду на game-домене (`localhost:5176`). Следовательно, нужен механизм
cross-domain передачи сессии.

## Решённое

Реализован **single-use handoff-code** flow:

1. Portal-frontend вызывает `POST /api/universes/{id}/session` →
   portal-backend проксирует в identity `POST /auth/universe-token`.
2. Identity-service генерирует UUIDv7 handoff-код, кладёт в Redis
   (`handoff:<token>` → `userID`, TTL 30s, single-use через GETDEL).
3. Portal возвращает `redirect_url` = `<game-uri>/auth/handoff?code=<X>`.
4. Browser делает `window.location.assign(redirect_url)`.
5. Game-фронт `HandoffPage` обменивает код через
   `POST /auth/token/exchange` → fresh access+refresh tokens.
6. Tokens сохраняются в game-localStorage; редирект на `/`.

## Почему так

### Альтернативы рассмотрены

**A. Cookies на родительском домене (`*.oxsar-nova.ru`).** Отвергнуто:
- Localstorage у нас плоский, не cookie. Менять storage = большая
  миграция.
- В dev (localhost:NNNN) cookies на «родительском» домене не работают
  без отдельной настройки `/etc/hosts`. Дев-experience страдает.
- Cookies везде → CSRF concerns в каждом game-API. Сейчас Bearer-only,
  CSRF не нужен.

**B. URL fragment (`#token=...`).** Отвергнуто:
- Fragment не передаётся в Referer, но логируется браузером в history.
- Один JS-error в game-фронте — токен утёк в Sentry/log агрегатор.
- Один-time код безопаснее: даже если `?code=` попадёт в Referer
  (он не попадёт, мы сразу делаем `history.replaceState`), он
  одноразовый и протух через 30s.

**C. OAuth implicit flow.** Отвергнуто:
- Семантически = то же что мы сделали, но с тяжёлой OAuth-машинерией
  (consent screen, scopes, state-параметр). У нас single-tenant SSO,
  consent не нужен.

**D. Session-cookie передаётся через `postMessage` (iframe-bridge).**
Отвергнуто:
- Сложная схема с invisible iframe.
- Хрупкая (third-party cookie блокировки в Safari/Firefox/Brave).

### Выбрана handoff-code, потому что

- ✅ Простая реализация (≈30 строк handler-кода в identity).
- ✅ Безопасно: код одноразовый, TTL 30s, привязан к user_id,
  обмен через server-to-server (никаких клиентских secrets).
- ✅ Нет cookies → нет CSRF-surface на game-API.
- ✅ Работает идентично в dev (HTTP localhost) и prod (HTTPS).
- ✅ Минимальный URL-leak risk: код в URL живёт один request → редирект,
  history заменяется на `/`.

## Безопасность

- **Single-use**: Redis GETDEL атомарно. Code не может быть переиспользован.
- **TTL 30s**: окно перехвата (если бы кто-то подсмотрел код) минимально.
- **User binding**: код хранит `userID`, identity при exchange выдаёт
  токен ровно этому юзеру.
- **Universe binding** (план 72.2): identity дополнительно проверяет
  что universe_id из запроса соответствует выбранной (anti-replay).
- **Rate-limit**: 25/мин/IP на оба endpoint'а
  (`/auth/universe-token`, `/auth/token/exchange`) — защита от
  bruteforce и DoS.
- **HTTPS-only в prod**: portal-backend строит `redirect_url` как
  `https://<subdomain>.<base>` если `dev_url` пусто (по умолчанию
  в проде). Перехват по сети — невозможен.
- **Audit log**: каждое issue + exchange логируется в identity (audit
  таблица — вынесено в TODO 72.2 §9).

## LoginScreen на game-доменах удалён

После плана 72.2 на `uni*.oxsar-nova.ru` **нет формы пароля вообще**.
Это:
- Снижает фишинг-риск (атакующему нет смысла подделывать game-домен —
  юзер всё равно логинится только на портале).
- Упрощает UX (одна точка входа = меньше путаницы).
- Стимулирует юзера запоминать только один URL (`oxsar-nova.ru`).

Юзер пришёл на game-домен без токена → автоматический редирект на портал.

## Refresh-flow

Параллельно реализован refresh-flow в game-фронтах:
- На 401 от game-API → пытаемся `/auth/refresh` с refresh-токеном.
- Успех → retry оригинальный запрос.
- Fail → logout (стираем токены) → AuthGate увидит null и редиректит
  на портал.
- Race-protection: глобальный `refreshInFlight` промис гарантирует
  один refresh при N параллельных 401.

## Out of scope (для будущих ADR)

- Audit-таблица для handoff (issue/exchange events) — миграция нужна
  для compliance, отложено до запроса юристов.
- Device-binding (handoff-код привязать к UA/IP) — для mobile-юзеров
  и NAT может ломать. Решение зависит от threat-model.
- Multi-tab handoff (юзер в двух вкладках портала — какая выиграет).
  Сейчас просто оба запроса возвращают валидные коды, browser выберет
  тот в который кликнули. Race не страшен.
