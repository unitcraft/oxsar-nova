# Промпт: выполнить план 72 Ф.5 — Spring 4 (14 экранов: communication / settings / utilities)

**Дата создания**: 2026-04-28
**План**: [docs/plans/72-remaster-origin-frontend-pixel-perfect.md](../../plans/72-remaster-origin-frontend-pixel-perfect.md)
**Зависимости**: ✅ Ф.1 Bootstrap (54fabbdf46), ✅ Ф.2 Spring 1
(47d1f0ef65), ✅ Ф.3 Spring 2 (48ef07cf19+590a68b428), ✅ Ф.4
Spring 3 (150a832200), ✅ план 67 (alliance — для chat-ally),
✅ план 68 (биржа — но используется в Spring 5).
**Объём**: ~3000-4500 строк TS + CSS + i18n + возможно расширение
openapi.yaml (документирование существующих, но не задокументированных
endpoints), 2-3 коммита.

---

```
Задача: реализовать Spring 4 плана 72 — pixel-perfect клон 14
экранов origin-фронта группы communication/settings/utilities.

КОНТЕКСТ:

Ф.1-Ф.4 закрыты, в origin-фронте уже работают ~29-30 экранов
(Bootstrap layout + Spring 1 + Spring 2 + Spring 3). Backend
расширен в Ф.4 пакетом internal/catalog/. Сейчас Spring 4.

Spring 4 экраны (по docs/research/origin-vs-nova/origin-ui-
replication.md и плану 72 §«Реализация всех 50 prod-экранов»):

Communication (4):
- S-034 Friends (список друзей, добавление/удаление, поиск)
- S-035 MSG (личные сообщения inbox/sent/draft/trash)
- S-036 Chat (общий чат — вселенная или мир)
- S-037 ChatAlly (альянс-чат)

Notes/Search (2):
- S-038 Notepad (личные заметки игрока, уже есть endpoint
  /api/notepad из плана 69)
- S-039 Search (поиск игроков/альянсов)

Premium/Player (2):
- S-040 Officer (наём офицеров за credit/oxsariты)
- S-041 Profession (выбор профессии игрока — раз в жизни/N дней,
  даёт пассивные бонусы)

Settings (1):
- S-042 Settings (язык/часовой пояс/уведомления/email-предпочтения/
  смена пароля/удаление аккаунта)

Static pages (3):
- S-043 UserAgreement (страница пользовательского соглашения)
- S-044 Changelog (история обновлений игры)
- S-045 Support (форма обращения в поддержку — план 56 reports)

Tools (2):
- S-046 Widgets (виджеты на главной — отображаемая статистика,
  быстрые действия; в legacy — сборная страница)
- S-047 AdvTechCalculator (advanced tech calculator —
  калькулятор стоимости/времени исследований по уровням; чисто
  client-side утилита)

ИТОГО 14 экранов. План говорит «12 экранов» — фактически 14
после раскрытия. Tutorial (S-048) ИСКЛЮЧЁН из первой итерации
(см. план 72 «Что НЕ делаем»).

ВАЖНО ПРО BACKEND В SPRING 4:

Большинство endpoints УЖЕ существует:
- /api/friends (List/Add/Remove) — main.go строки 443-445.
- /api/messages, /api/messages/unread-count, /api/messages/{id},
  /api/messages/{id}/read — openapi 1078-1143.
- /api/chat/{kind}/{history,send,ws,unread} — main.go 578-581.
- /api/notepad — план 69.
- /api/search — main.go 435.
- /api/officers, /api/officers/{key}/activate — openapi 1285-1302.
- profession-handler — internal/profession/.
- settings-handler — internal/settings/.

ОДНАКО: многие из них замонтированы напрямую в main.go, **БЕЗ
описания в openapi.yaml**. Это нормально-работающее, но
недокументированное состояние.

ПОДХОД ДЛЯ SPRING 4:

(А) Добавь описания в openapi.yaml для эндпоинтов которыми
    пользуешься на фронте (полнота R2 = OpenAPI первым). Это
    R15-обязательно, не trade-off — endpoints существуют, но
    без openapi-документации фронт строится «по угадайке».
    Это закрывает технический долг плана 36/15/etc.

(Б) Если endpoint реально отсутствует (chat-history endpoints
    могут не иметь pagination, friends может не отдавать
    profile-info friend'а) — расширь backend по правилу R15
    (см. план 72 §«Backend-расширения по требованию»). Это
    полная реализация, не mock.

(В) S-046 Widgets — в legacy это «сборка на главной из других
    данных» (производство, очередь стройки, последние сообщения,
    статистика империи). В нашем S-001 Main (Spring 1) это
    **уже частично реализовано**. Spring 4 Widgets либо
    дублирует Main, либо вообще не существует как отдельный
    экран в новой архитектуре. Уточни в legacy templates/widgets.tpl
    — если это **на самом деле** тот же Main что мы сделали в
    Spring 1, помечай S-046 в simplifications.md как «закрыт
    через S-001 Main, Spring 4 не реализует» (✅ trade-off,
    дубликат). Если другое — реализуй.

(Г) S-047 AdvTechCalculator — клиентская утилита. Формулы
    стоимости/времени уже есть в catalog (Spring 3). Просто
    UI-форма с input level → output cost для исследования.
    Backend не нужен.

ВАЖНО ПРО CHAT:

В legacy chat использовался BBCode (`[b]bold[/b]`, `[url]...[/url]`).
План 72 явно говорит **«BBCode выкидывается, заменяется TipTap»**
(зависит от плана 57 mail-service).

Однако план 57 (mail) — справочный документ, **не выполняется**.
TipTap в origin-фронте отложен на Ф.8 плана 72.

В Spring 4: chat реализуется без TipTap, **plain text**. BBCode
сообщения из legacy-БД (если они там есть) рендерятся как plain
text с `[b]...[/b]` оставленными как литералы (не парсятся,
не render'ятся как HTML — это безопаснее и приемлемо для
переходного периода до Ф.8). Это **trade-off Ф.5 → Ф.8**,
записать в simplifications.md как P72.S4.X.

ПЕРЕД НАЧАЛОМ:

ПЕРВЫМ ДЕЙСТВИЕМ (до любого чтения плана):

1) git status --short. cat docs/active-sessions.md.

2) ОБЯЗАТЕЛЬНО добавь свою строку в раздел «Активные сессии»:
   | <N> | План 72 Ф.5 Spring 4 (14 экранов origin: communication+settings+utilities) | projects/game-nova/frontends/origin/, projects/game-nova/api/openapi.yaml (документирование existing endpoints), docs/plans/72-..., docs/active-sessions.md, docs/project-creation.txt, docs/simplifications.md | <дата-время> | feat(origin/frontend): Ф.5 Spring 4 — communication+settings+utilities |

3) Параллельных сессий не должно быть — это большая сессия в
   одиночку. Если кто-то ещё активен — спроси пользователя.

ТОЛЬКО ПОСЛЕ — переходи к чтению:

4) Прочитай ПОЛНОСТЬЮ:
   - docs/plans/72-remaster-origin-frontend-pixel-perfect.md
   - docs/research/origin-vs-nova/origin-ui-replication.md секции
     S-034..S-047 (детальное описание)
   - docs/research/origin-vs-nova/roadmap-report.md «Часть I.5» R0-R15

5) Прочитай выборочно:
   - projects/game-nova/api/openapi.yaml — секции messages,
     officers (есть в openapi); проверь что НЕТ для friends,
     chat, search, profession, settings, support — это нужно
     добавить.
   - projects/game-nova/backend/cmd/server/main.go — найди
     replicate routes для friends/chat/search/profession/settings
     (handler.Method вызовы).
   - projects/game-nova/backend/internal/{friends,chat,search,
     officer,profession,settings,notepad,message}/handler.go —
     понять реальные protocols (DTO, query params).
   - projects/game-nova/frontends/origin/src/main.tsx (роутер
     для добавления новых routes).
   - projects/game-nova/frontends/origin/src/api/ — API-инфра.
   - 1-2 экрана из Spring 1-3 как эталоны UX-стиля
     (например, MainScreen, AllianceOverviewScreen).
   - projects/game-legacy-php/templates/ — *.tpl для каждого
     S-NNN экрана (friends.tpl, msg.tpl, chat.tpl, chatally.tpl,
     notepad.tpl, search.tpl, officer.tpl, profession.tpl,
     settings.tpl, useragreement.tpl, changelog.tpl, support.tpl,
     widgets.tpl, advtechcalculator.tpl).

ЧТО НУЖНО СДЕЛАТЬ:

### Архитектурно

1. **Routes** в роутере origin-фронта:
   - /friends → FriendsScreen
   - /msg/:folder? → MessagesScreen (folder=inbox/sent/draft/trash)
   - /chat → ChatScreen (общий)
   - /chat/ally → ChatAllyScreen
   - /notepad → NotepadScreen
   - /search → SearchScreen
   - /officer → OfficerScreen
   - /profession → ProfessionScreen
   - /settings → SettingsScreen
   - /user-agreement → UserAgreementScreen
   - /changelog → ChangelogScreen
   - /support → SupportScreen
   - /widgets → WidgetsScreen (если не закрыт через Main)
   - /tools/tech-calc → AdvTechCalculatorScreen

2. **API-модули** в src/api/:
   - friends.ts, messages.ts (расширить если есть), chat.ts,
     notepad.ts (есть с Spring 1?), search.ts, officer.ts,
     profession.ts, settings.ts, support.ts.

3. **OpenAPI расширение** — для каждого endpoint'а который
   используется во frontend и не задокументирован в openapi.yaml:
   - Добавить path-секцию с request/response/parameters.
   - DTO в components/schemas.
   - НЕ переписывать backend-handler — он работает; только
     документация.

   Это R15-обязательная часть, не trade-off. Каждый
   незадокументированный endpoint — пробел в R2 «OpenAPI первым»,
   и закрытие этого пробела часть Spring 4.

4. **i18n (R12)** — grep по configs/i18n/{ru,en}.yml на
   friends|message|chat|notepad|search|officer|profession|
   settings|support. Цель ≥95% переиспользования. План 67 уже
   добавил chat/friends/messages-ключи — переиспользуй.

5. **Pixel-perfect** — HTML+CSS-классы зеркалят legacy *.tpl.

6. **Тесты** — vitest + RTL, 1-2 теста на каждый экран. Spring 4
   = 14 экранов = 16-25 тестов минимум.

7. **Idempotency-Key (R9)** на mutation'ах:
   - friends.add (POST), friends.remove (DELETE).
   - messages.send (POST).
   - chat.send (POST) — обязательно, без него можно дважды
     отправить одно сообщение.
   - officer.activate (POST).
   - profession.choose (POST).
   - settings.update (POST/PUT).
   - support.submit (POST).

### Экраны (детали — изучи по легаси-tpl)

#### S-034 Friends (FriendsScreen)

- Список друзей (TanStack Query → GET /api/friends).
- Каждая карточка: username, last-seen, planet (опционально).
- Действия: «Удалить» (DELETE /api/friends/{userId} с
  confirm-modal), «Написать сообщение» (navigate /msg/new).
- Кнопка «Добавить» → modal с поиском по username/email →
  POST /api/friends/{userId}.

#### S-035 MSG (MessagesScreen)

- Tabs: Inbox / Sent / Draft / Trash. URL /msg/inbox по
  умолчанию.
- Список сообщений (TanStack Query useInfiniteQuery
  /api/messages?folder=...&cursor=...).
- Каждая строка: from, subject, preview, date, read/unread.
- Клик на сообщение → /msg/{id} → MessageDetailScreen.
- Кнопка «Новое» → MessageComposeScreen — input to (username),
  subject, body (plain text — TipTap только в Ф.8).
- Idempotency-Key на send.

#### S-036 Chat (ChatScreen) и S-037 ChatAlly (ChatAllyScreen)

- WebSocket к /api/chat/{kind}/ws (kind=universe для S-036,
  ally для S-037).
- Buffer messages в Zustand chat-store.
- TanStack Query → GET /api/chat/{kind}/history для load-on-mount.
- Каждое сообщение: from, body (plain text BBCode literal —
  P72.S4.X simplification), timestamp.
- Input + кнопка Send (POST /api/chat/{kind}/send с
  Idempotency-Key).
- Auto-scroll к новому.

#### S-038 Notepad (NotepadScreen)

- Existing /api/notepad endpoint (план 69).
- TanStack Query → GET /api/notepad → string content.
- Textarea с auto-save (debounce 1s) → POST /api/notepad с
  body content.
- Лимит 50000 символов (план 69).

#### S-039 Search (SearchScreen)

- Tabs: Игроки / Альянсы / Планеты (если применимо).
- Input + результаты в TanStack Query → GET /api/search?q=&type=.
- Клик на результат → переход к профилю/альянсу/координатам.

#### S-040 Officer (OfficerScreen)

- Список доступных officer-typов (TanStack Query →
  GET /api/officers).
- Каждый officer: имя, описание, эффект, цена в credit/oxsariты,
  длительность, активный/неактивный.
- Кнопка «Нанять» → POST /api/officers/{key}/activate с
  Idempotency-Key.
- Возможные ошибки: 402 insufficient credit/oxsariты, 409
  already active.

#### S-041 Profession (ProfessionScreen)

- Список профессий (фермер / пират / шахтёр / etc.).
- Текущая профессия игрока подсвечена.
- Кнопка «Выбрать» → POST /api/profession (если cooldown
  позволяет, иначе disabled с tooltip «Доступно через X дней»).
- Idempotency-Key.

#### S-042 Settings (SettingsScreen)

- Form с разделами:
  - Язык (select ru/en).
  - Часовой пояс (select).
  - Email-уведомления (checkbox-список по типам).
  - Vacation mode (toggle, опционально).
  - Смена пароля (input old + new + confirm, отдельная кнопка).
  - Удаление аккаунта (отдельная секция с confirm-flow,
    использует /api/auth/account/delete или подобный из плана 51).
- POST /api/settings с Idempotency-Key.

#### S-043 UserAgreement (UserAgreementScreen)

- Статическая страница из i18n (или /api/legal/user-agreement
  endpoint если есть в backend).
- Markdown rendering или HTML (без TipTap).

#### S-044 Changelog (ChangelogScreen)

- Список релизов с версиями и описаниями.
- Источник: либо статический MD-файл в frontend, либо
  /api/changelog endpoint. Уточни в legacy.

#### S-045 Support (SupportScreen)

- Form: subject, body, category (game-bug/payment/account/other).
- POST /api/reports (план 56) или /api/support — уточни
  существующий endpoint.
- Idempotency-Key.

#### S-046 Widgets (WidgetsScreen)

- УТОЧНИ В LEGACY: templates/widgets.tpl.
- Если это **дубликат Main** (Spring 1 S-001) — помечай в
  simplifications.md как «закрыт через S-001», routes для
  /widgets либо вообще нет, либо redirect на /.
- Если есть отдельная семантика — реализуй.

#### S-047 AdvTechCalculator (AdvTechCalculatorScreen)

- Чисто client-side утилита.
- Form: select research → input target_level → output cost
  (метал/кристалл/водород) + время.
- Формулы — из catalog Spring 3 (GET /api/research/tree уже
  отдаёт level_formulas).
- Опционально: comparison table нескольких research's.
- Backend не нужен.

### Финализация Ф.5

- Шапка плана 72: Ф.5 ✅, Spring 4 закрыт.
- НЕ закрываешь весь план 72 — впереди Ф.6 (Spring 5),
  Ф.7 (i18n рус), Ф.8 (TipTap чат), Ф.9 (финал).
- Запись итерации в docs/project-creation.txt («72 Ф.5 — Spring 4»).
- В коммите указать i18n переиспользовано/новых.

══════════════════════════════════════════════════════════════════
ОБЯЗАТЕЛЬНЫЕ ПРАВИЛА R0-R15 — см. блок rules-r0-r15.md
СКОПИРУЙ СЮДА полностью раздел из docs/prompts/_blocks/rules-r0-r15.md

Особо важно:
- R0: backend nova не меняем по существу (только документация
  openapi).
- R2: OpenAPI первым — добавь описания всех использованных
  endpoints. Это закрывает тех-долг.
- R5: pixel-perfect для origin-фронта.
- R9: Idempotency-Key на send/add/remove/activate/choose/
  update/submit.
- R12: i18n grep сначала, цель 95%.
- R15: BBCode → plain text — это TRADE-OFF (P72.S4.X), TipTap в
  Ф.8. Documented в simplifications.md, не пропуск.
══════════════════════════════════════════════════════════════════

══════════════════════════════════════════════════════════════════
GIT-ИЗОЛЯЦИЯ ОТ ПАРАЛЛЕЛЬНЫХ СЕССИЙ — см. блок git-isolation.md
СКОПИРУЙ СЮДА полностью раздел из docs/prompts/_blocks/git-isolation.md

Свои пути для CC_AGENT_PATHS:
- projects/game-nova/frontends/origin/

  (вся папка целиком — твой territory)

- projects/game-nova/api/openapi.yaml (расширение документацией
  для chat/friends/search/profession/settings/support)
- docs/plans/72-remaster-origin-frontend-pixel-perfect.md
- docs/active-sessions.md
- docs/project-creation.txt
- docs/simplifications.md (P72.S4.* записи)

ВНИМАНИЕ: НЕ трогай:
- projects/game-nova/frontends/nova/ — там работает nova-фронт.
- projects/game-nova/backend/ — backend закрыт, документация в openapi.
- projects/game-legacy-php/ — только читаешь как источник истины.
══════════════════════════════════════════════════════════════════

КОММИТЫ:

2-3 коммита (изоляция blame для крупной сессии):

1) feat(origin/frontend): Ф.5 Spring 4 ч.1 — communication
   (friends + msg + chat + ally-chat) (план 72)
2) feat(origin/frontend): Ф.5 Spring 4 ч.2 — settings + premium
   (notepad + search + officer + profession + settings) (план 72)
3) feat(origin/frontend): Ф.5 Spring 4 ч.3 — static + utilities
   (user-agreement + changelog + support + widgets + tech-calc) +
   финализация (план 72)

ИЛИ если объём окажется средним (~3000 строк) — 2 коммита:
1) ч.1 communication + settings (~7-8 экранов)
2) ч.2 static + utilities + финализация (~6-7 экранов)

Trailer: Generated-with: Claude Code
ВСЕГДА с двойным тире:
git commit -m "..." -- $CC_AGENT_PATHS

ЧЕГО НЕ ДЕЛАТЬ:

- НЕ менять nova-фронт.
- НЕ менять backend handlers (только openapi-документацию).
- НЕ реализовывать TipTap для chat — это Ф.8.
- НЕ реализовывать Tutorial — исключён из первой итерации.
- НЕ переносить рекламу/баннеры из legacy.
- НЕ закрывать весь план 72 — Ф.5 это Spring 4.
- НЕ забывай Idempotency-Key.
- НЕ забывай про -- в git commit.

УСПЕШНЫЙ ИСХОД:

- 13-14 экранов работают (router → каждый рендерится).
- typecheck + build + tests зелёные.
- Все экраны pixel-perfect клоны legacy.
- i18n: 95%+ переиспользования.
- openapi.yaml расширена для chat/friends/search/profession/
  settings/support (документирование existing endpoints).
- Шапка плана 72: Ф.5 ✅ (Spring 4).
- BBCode → plain text trade-off записан в simplifications.md
  (P72.S4.X), ссылка на Ф.8.
- Запись в docs/project-creation.txt.
- Удалена строка из docs/active-sessions.md.

Стартуй. Это длинная сессия 6-10 часов — промежуточные апдейты
текстом по 5-10 слов («communication готов», «settings готов»,
«ч.1 коммит», и т.д.). Моя проверка — после финального коммита.
```
