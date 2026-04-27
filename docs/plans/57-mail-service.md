# План 57: Mail-service — единая почтовая система oxsar-nova

**Дата**: 2026-04-27
**Статус**: Черновик-эпик (запускать **после** публичного запуска
проекта; сейчас — справочный документ для будущего планирования).
**Зависимости**: план 51 (identity), план 52 (RBAC), план 53
(admin-frontend + admin-bff), план 56 (reports → portal как
прецедент централизованного сервиса).
**Связанные документы**: [38-billing-service.md](38-billing-service.md)
(паттерн отдельного микросервиса), [56-reports-to-portal.md](56-reports-to-portal.md)
(паттерн централизации платформенных функций).

---

## Цель

Создать **единую почтовую систему** для всех вселенных проекта
oxsar-nova: личные сообщения между игроками + богатые системные
сообщения от платформы (боевые отчёты, шпионские данные,
бан-уведомления, биллинговые квитанции, событийные подарки).

### Ключевые архитектурные решения

1. **Отдельный микросервис `mail-service`** — не в portal, не в
   game-nova. Чистый Go-модуль `projects/mail/` с собственной БД.
   Аргументация — см. раздел «Архитектурные решения» ниже.

2. **Frontend на TipTap** (MIT) — rich text editor с custom-nodes
   для богатого контента. Единый клиент-компонент, используется во
   всех React-фронтендах (game-nova, portal, после переписывания —
   game-origin-go).

3. **Единый inbox** для пользователя на всех вселенных. Игрок,
   играющий в uni01 и uni02, видит общий inbox; кросс-вселенские
   личные сообщения — естественны.

4. **Системные сообщения как rich-карточки** — не plain text.
   Боевой отчёт = интерактивная карточка с превью кораблей,
   ссылками на планеты, кнопками «Атаковать». Шпион-отчёт —
   карта/таблица с шкалой обнаружения. Бан-уведомление — цветная
   карточка со ссылкой на правила.

5. **Уведомления в шапке** игровых клиентов — иконка inbox с
   счётчиком непрочитанных. Inline-preview последних 5 писем без
   полного редактора, полный inbox — на portal/mail или модалкой.

---

## Архитектурные решения

### Почему отдельный микросервис, а не в portal

| Критерий | Mail в portal | Отдельный mail-service |
|---|---|---|
| Архитектурная согласованность | portal становится «жирным» | следует тренду (identity, billing, mail) |
| Изоляция отказов | сбой почты валит portal | портал и почта независимы |
| Масштабируемость | общий ресурс | отдельные ресурсы для write-heavy потока |
| Owner/ответственность | размывается | чёткий: «всё про почту → mail-service» |
| Эксплуатация | проще на старте | +1 сервис в compose, но управляемо |
| Соответствие проекту | приемлемо | **лучше** — четвёртый Go-модуль рядом с identity/billing/portal |

Решение: **отдельный mail-service**. Цена — один лишний Docker-сервис;
взамен — изоляция отказов, чёткий owner, независимое масштабирование.

### Почему TipTap

| Кандидат | Лицензия | Подходит? |
|---|---|---|
| TinyMCE 6+ | GPL/коммерческая | ❌ GPL несовместим с PolyForm |
| CKEditor 5 | GPL/коммерческая | ❌ |
| **TipTap** | MIT | ✅ |
| Lexical (Meta) | MIT | ✅, но меньше готовых компонентов |
| Quill | BSD-3-Clause | ✅, но меньше расширяем |
| Trix | MIT | ✅, но плохо расширяем под custom-nodes |
| Editor.js | Apache-2.0 | ❌ — block-style, не подходит под почту |
| Slate.js | MIT | ✅, но слишком низкоуровневый (3+ дня на MVP) |

Решение: **TipTap**. Аргументы:
- MIT;
- архитектура headless + extensions идеальна для custom-nodes
  (battle-report-card, player-mention, planet-link и т.п.);
- активное развитие, документация, экосистема плагинов;
- используется в Notion-clones, Linear, Atlassian — устойчивый выбор.

### Почему единый inbox, а не «inbox per universe»

При переписывании game-origin на Go+React (в будущем) у игрока
будет один identity-аккаунт на все вселенные. Дробить inbox по
вселенным — фрагментирует UX (как Gmail с разными inbox'ами для
разных адресов). Единый inbox с тегом «вселенная» в каждом письме —
стандарт индустрии.

Кросс-вселенский диалог между альянсами — естественно работает.
Системные письма от platform'ы (биллинг, бан, обновления) — приходят
в тот же inbox без дублирования.

---

## Модель данных

### Таблица `mailboxes` (одна на пользователя)

```sql
CREATE TABLE mailboxes (
    user_id            UUID PRIMARY KEY,  -- из identity
    quota_bytes        BIGINT NOT NULL DEFAULT 104857600, -- 100 MB
    used_bytes         BIGINT NOT NULL DEFAULT 0,
    unread_count       INT NOT NULL DEFAULT 0,
    last_message_at    TIMESTAMPTZ,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### Таблица `folders`

```sql
CREATE TABLE folders (
    id          BIGSERIAL PRIMARY KEY,
    user_id     UUID NOT NULL REFERENCES mailboxes(user_id),
    name        TEXT NOT NULL,
    type        TEXT NOT NULL,  -- inbox/sent/drafts/trash/spam/custom
    sort_order  INT NOT NULL DEFAULT 0,
    UNIQUE (user_id, name)
);
```

Системные папки (inbox/sent/drafts/trash/spam) создаются при
первом обращении пользователя.

### Таблица `messages`

```sql
CREATE TABLE messages (
    id              UUID PRIMARY KEY,
    thread_id       UUID,                    -- группировка по диалогу
    folder_id       BIGINT REFERENCES folders(id),
    user_id         UUID NOT NULL,           -- владелец копии письма
    sender_id       UUID,                    -- NULL для системных
    sender_kind     TEXT NOT NULL,           -- 'user' / 'system' / 'admin'
    sender_label    TEXT,                    -- для системных: 'Боевой ИИ', 'Биллинг', и т.д.
    universe_id     TEXT,                    -- 'uni01', 'uni02', 'origin' — если письмо привязано к вселенной
    subject         TEXT NOT NULL,
    body_json       JSONB NOT NULL,          -- TipTap ProseMirror document
    body_html       TEXT,                    -- pre-rendered для быстрого превью
    is_read         BOOLEAN NOT NULL DEFAULT FALSE,
    is_starred      BOOLEAN NOT NULL DEFAULT FALSE,
    has_attachments BOOLEAN NOT NULL DEFAULT FALSE,
    sent_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at      TIMESTAMPTZ              -- для авто-удаления старых писем
);

CREATE INDEX ON messages (user_id, folder_id, sent_at DESC);
CREATE INDEX ON messages (thread_id);
```

### Таблица `attachments`

```sql
CREATE TABLE attachments (
    id              UUID PRIMARY KEY,
    message_id      UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    filename        TEXT NOT NULL,
    content_type    TEXT NOT NULL,
    size_bytes      BIGINT NOT NULL,
    storage_key     TEXT NOT NULL,           -- S3-ключ (Selectel Object Storage)
    uploaded_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### Таблица `system_templates`

```sql
CREATE TABLE system_templates (
    id              TEXT PRIMARY KEY,        -- 'battle.report', 'billing.receipt'
    title_template  TEXT NOT NULL,           -- 'Бой на планете {planet}'
    body_template   JSONB NOT NULL,          -- TipTap-документ с placeholders
    category        TEXT NOT NULL,           -- 'combat', 'billing', 'moderation', 'event', 'announcement'
    severity        TEXT NOT NULL DEFAULT 'info',  -- info/warning/danger
    locale          TEXT NOT NULL DEFAULT 'ru',
    UNIQUE (id, locale)
);
```

Шаблоны редактируются в admin-frontend (план 53), модерация
системных писем — через RBAC `mail:templates:edit`.

---

## Custom TipTap-nodes

Сверх стандартного rich-text (bold/italic/lists/links) добавляются:

| Node | Что делает | Пример использования |
|---|---|---|
| `<player-mention data-user-id="...">` | ссылка на профиль игрока + tooltip | «@AdmiralVeles напиши когда сможешь» |
| `<planet-link data-coords="1:42:7">` | ссылка на координаты планеты + миниатюра | «встретимся на 1:42:7» |
| `<fleet-tag data-fleet-id="...">` | компактная карточка флота с составом | «отправил на тебя [этот флот]» |
| `<battle-report-card data-battle-id="...">` | rich-карточка боевого отчёта | системное: результат боя |
| `<spy-report-card data-spy-id="...">` | карточка шпионажа (юниты, ресурсы, шкала обнаружения) | системное |
| `<expedition-summary-card data-event-id="...">` | результат экспедиции | системное |
| `<resource-gift data-amount="...">` | анимированный подарок ресурсов с кнопкой «Получить» | системное |
| `<system-notification data-severity="warning">` | цветная полоса с заголовком/иконкой | бан-уведомление, апдейты правил |
| `<adr-embed data-adr-id="...">` | превью ADR с before/after | при изменении баланса |
| `<vote-card data-proposal-id="...">` | embed голосования из portal | приглашение проголосовать |

Custom-nodes реализуются как TipTap-extensions, рендерятся
React-компонентами. Frontend клиент — единый пакет
`projects/mail/frontend-client/` (npm-package или shared-module),
импортируется во все React-фронтенды.

---

## API-контракт (черновик)

### Игроку

```
GET    /api/mail/folders                 список папок
GET    /api/mail/messages                inbox с пагинацией, фильтрами
GET    /api/mail/messages/{id}           одно письмо
POST   /api/mail/messages                отправить письмо (body_json TipTap)
POST   /api/mail/messages/{id}/read      пометить прочитанным
POST   /api/mail/messages/{id}/star      звезда
POST   /api/mail/messages/{id}/move      переместить в папку
DELETE /api/mail/messages/{id}           удалить (в trash, потом hard delete)
POST   /api/mail/attachments             загрузить вложение (multipart, → S3)
GET    /api/mail/attachments/{id}        скачать
GET    /api/mail/inbox/summary           счётчик непрочитанных + 5 последних (для уведомлений в шапке)
```

### Системные письма (вызываются другими сервисами)

```
POST   /api/mail/system                  отправить письмо от системы
                                         {template_id, recipient_user_id, vars}
                                         требует internal-token, не пользовательский JWT
```

Вызовы:
- **game-nova battle-engine** → `POST /api/mail/system` с `template_id=battle.report`,
  `vars={battle_id, attacker, defender, ...}`.
- **billing** → `template_id=billing.receipt`, `vars={amount, package_id, ...}`.
- **moderation** → `template_id=moderation.ban_notice`, `vars={reason, until, ...}`.

### Админу (через admin-bff)

```
GET    /api/admin/mail/messages          поиск по всем юзерам
GET    /api/admin/mail/messages/{id}     просмотр (проверка жалоб от плана 56)
POST   /api/admin/mail/messages/{id}/quarantine   карантин (спам/нарушение)
GET    /api/admin/mail/templates         список системных шаблонов
PUT    /api/admin/mail/templates/{id}    редактирование шаблона
```

RBAC: `mail:read:any`, `mail:moderate`, `mail:templates:edit`.

---

## Этапы (укрупнённо, без детализации)

### Ф.1. Скаффолдинг mail-service

- `projects/mail/backend/` — go.mod, скелет main.go, Dockerfile;
- `projects/mail/migrations/` — миграции 0001 (схема БД);
- `projects/mail/api/openapi.yaml` — контракт;
- Регистрация в `deploy/docker-compose.yml`;
- CI: build + lint + license-check (план 40).

### Ф.2. Базовый CRUD писем

- POST/GET/DELETE messages.
- Folders (inbox/sent/drafts/trash).
- read/star/move actions.
- Аутентификация через identity-JWT.
- Без TipTap, body — plain text для тестирования API.

### Ф.3. Frontend-клиент (TipTap MVP)

- `projects/mail/frontend-client/` — shared npm-package
  (или просто папка с экспортами).
- TipTap с базовым toolbar (Bold/Italic/Underline/Lists/Link/Quote/Undo).
- React-компоненты `<MailInbox>`, `<MailComposer>`, `<MailMessage>`.
- Стилизация под общий дизайн oxsar-nova.

### Ф.4. Custom TipTap-nodes (этап 1 — базовые)

- `<player-mention>` (с identity-резолвом username).
- `<planet-link>` (с резолвом владельца через game-nova).
- `<fleet-tag>` (с резолвом состава флота).
- Toolbar-кнопки для вставки.

### Ф.5. Системные письма

- Endpoint `POST /api/mail/system`.
- Internal-token для вызовов из других сервисов.
- Шаблонная система (`system_templates` + variable substitution).
- Custom-nodes этапа 2:
  - `<battle-report-card>`,
  - `<spy-report-card>`,
  - `<expedition-summary-card>`,
  - `<resource-gift>`,
  - `<system-notification>`.

### Ф.6. Интеграция с источниками системных писем

- game-nova battle-engine → шлёт боевые отчёты.
- billing → шлёт квитанции.
- moderation → шлёт бан-уведомления.
- portal → шлёт обновления правил, объявления.

### Ф.7. Уведомления в шапке клиентов

- В game-nova frontend — иконка inbox + счётчик +
  inline-preview 5 последних писем.
- Аналогично в portal.
- Polling или WebSocket (на старте — polling каждые 30 секунд).

### Ф.8. Вложения

- S3-storage (Selectel Object Storage по memory `project_audience_ru`).
- Multipart upload через `/api/mail/attachments`.
- Quota-проверка по `mailboxes.used_bytes`.
- Auto-cleanup attachments при удалении письма.

### Ф.9. Админка через admin-bff

- В admin-frontend — страницы поиска/просмотра почты, редактирования шаблонов.
- Proxy-эндпоинты в admin-bff (план 53 паттерн).
- RBAC: `mail:read:any`, `mail:moderate`, `mail:templates:edit`.

### Ф.10. Финализация

- Юр-аудит (промпт `legal-compliance-audit.md`):
  152-ФЗ — переписка между пользователями относится к ПДн,
  должна быть отражена в Privacy Policy. 149-ФЗ — модерация спама/
  противоправного контента в почте. Хранение логов 6 месяцев — для
  жалоб (интеграция с планом 56).
- Документация (`docs/ops/mail-service.md`).
- Миграция данных (если есть legacy переписки — отдельная задача).
- Запись в `docs/project-creation.txt`.

**Итоговый объём**: 2–3 недели работы агента, 8–15 коммитов.

---

## Что отложено (не в scope этого плана)

- **Real-time чат-как-почта** (мгновенные сообщения вне почты) —
  отдельный план, отдельный сервис.
- **Голосовые сообщения** — требует отдельной аудио-инфраструктуры.
- **Шифрование сообщений end-to-end** — для игрового контента
  обычно избыточно.
- **Anti-spam ML/heuristics** — на старте достаточно blacklist'а
  из плана 46 + жалоб через план 56. ML — позже.
- **Email-уведомления о новых письмах в почте** (мета-почта) —
  через `symfony/mailer` или эквивалент, отдельная задача после
  настройки SMTP-провайдера.
- **Почтовые группы / рассылки** — для системных объявлений; на
  старте достаточно «отправка всем активным» через batch-генерацию.

---

## Юридические аспекты (для будущей юр-проверки)

При реализации обязательно учесть:

1. **152-ФЗ**: переписка пользователей — это персональные данные
   (содержимое сообщения может содержать ПДн). Privacy Policy
   должна включать упоминание почты как обрабатываемой категории
   (поправка к `docs/legal/privacy-policy.md` после реализации).
   Право на удаление аккаунта (план 44) — должно деперсонализировать
   и почту: `sender_id` затереть, `body_json` оставить (получатель
   имеет право на свою копию письма).
2. **149-ФЗ**: модерация спама/противоправного контента + хранение
   данных по запросу регуляторов 6 месяцев. Жалобы на письма
   приходят через план 56 (reports → portal). После резолюции —
   возможен карантин/удаление.
3. **436-ФЗ**: возрастная маркировка отображается в footer'е
   почтового интерфейса (общий компонент AgeRating из плана 46).
4. **Оферта**: использование почты подпадает под общие правила
   игры (план 47, `docs/legal/game-rules.md`) — запрет спама,
   оскорблений, противоправного контента.

---

## Триггеры для запуска плана

Этот план не запускается «по таймеру». Триггеры:

- **Публичный запуск проекта произошёл, проект стабилен** — почта
  становится одной из приоритетных фич follow-up'а.
- **Игроки активно жалуются** на отсутствие нормальной почты в
  legacy-чате (текущий Smarty-based чат — не почта в смысле
  inbox).
- **Game-origin переписан на Go+React** — это разблокирует
  единый клиент-компонент для всех вселенных.
- **Появился запрос на богатые системные сообщения** (например,
  игроки хотят боевой отчёт-карточку, а не таблицу).

До этих триггеров — план остаётся черновиком-эпиком как ориентир
для будущих архитектурных решений.

---

## Итог

Эпик-план на 2–3 недели работы (10 фаз). Архитектурное решение —
отдельный микросервис `mail-service` на Go + единый TipTap-клиент
во всех React-фронтендах + богатые custom-nodes для системных писем.
Запускается после публичного запуска и переписывания game-origin
на Go+React. До тех пор — справочный документ.
