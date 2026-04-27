# Аудит юридического соответствия

**Дата последнего аудита:** 2026-04-27
**Промпт:** `docs/prompts/legal-compliance-audit.md`
**Связанные документы:** [release-roadmap.md](../release-roadmap.md),
`docs/plans/{39,40,41,42,43,44,45,46,47,48,49}-*.md`,
[license-audit.md](license-audit.md).

## Покрытие

Проверены все 5 текущих подпроектов в `projects/`:

| Подпроект | Стек | Манифест |
|---|---|---|
| identity | Go 1.23 | `backend/go.mod` |
| billing | Go 1.23 | `backend/go.mod` |
| game-nova | Go 1.23 + TS (npm) | `backend/go.mod`, `frontend/package.json` |
| game-origin | PHP 8.3 (clean-room) | `composer.json` (без внешних зависимостей) |
| portal | Go 1.23 + TS (npm) | `backend/go.mod`, `frontend/package.json` |

> Примечание: на момент аудита (27.04.2026) подпроект назывался
> `auth`, переименован в `identity` по плану 51 в тот же день. В
> таблицах ниже используется новое имя.

Других стеков (Rust/Python/Java) нет. Manifest-обнаружение нашло те же
проекты, что и `ls projects/` — синхронизация полная.

## Чек-лист соответствия

| Требование | identity | billing | game-nova | game-origin | portal | Статус |
|---|---|---|---|---|---|---|
| Лицензии Go-зависимостей (CI license-check) | ✅ | ✅ | ✅ | — | ✅ | ОК |
| Лицензии npm production-зависимостей | — | — | ✅ | — | ✅ | ОК |
| Лицензии Composer-зависимостей | — | — | — | ⚠️ N/A | — | См. gap 1.1 |
| Отсутствие GPL/AGPL/LGPL в исходниках | ✅ | ✅ | ✅ | ✅ | ✅ | ОК (план 50 Ф.1) |
| Атрибуция AI в коммитах (новый trailer) | — | — | — | — | — | См. gap 3.1 |
| Чекбокс согласия в форме регистрации | ✅ | — | ✅ | ⚠️ | ✅ | См. gap 4.1 |
| Миграция user_consents (или эквивалент) | ✅ | — | — | — | — | ОК (identity централизован) |
| Endpoint `DELETE /users/me` | ✅ | — | — | — | — | ОК (identity централизован) |
| Возрастная маркировка «12+» | — | — | ✅ | ✅ | ✅ | ОК (план 50 Ф.3) |
| UGC blacklist при создании контента | ✅ | — | ✅ | ✅ | — | ОК (план 50 Ф.4) |
| Кнопка «Пожаловаться» в UI | — | — | ✅ | ✅ | — | План 50 Ф.5 (2026-04-28) |
| Чекбокс акцепта оферты | ✅ | — | ✅ | ⚠️ | ✅ | См. gap 4.1 (то же место) |
| Ссылки на оферту/правила/refund/privacy в UI | — | — | — | ✅ | ✅ | ОК (план 50 Ф.6) |
| Webhook-безопасность платёжного шлюза | — | ✅ | — | — | — | ОК (план 42) |

Легенда: ✅ — реализовано; — — поверхность отсутствует, не требуется;
⚠️ — gap, см. ниже.

## Найденные gap'ы

### Gap 1.1 — PHP/Composer аудит лицензий

`projects/game-origin/composer.json` не содержит внешних зависимостей —
все require это PHP-расширения. CI-job `license-check` PHP не сканирует.

Это безопасное состояние, но защита **не закреплена в CI**. Если в
будущем кто-то добавит внешнюю Composer-зависимость с GPL — CI этого
не поймает.

**Что делать:** расширить план 40 новой фазой — добавить в CI
проверку, что `composer.json` не содержит секцию `require` со
значениями кроме `php` и `ext-*`. Альтернатива — настроить
`composer licenses` со whitelist'ом, если зависимости появятся.

### Gap 2.1 — LGPL-зависимость TinyMCE в game-origin ✅ ЗАКРЫТО (2026-04-27, план 50 Ф.1)

В `projects/game-origin/public/js/tiny_mce/` лежал **TinyMCE под
LGPL 2.1**. LGPL — copyleft, несовместим с PolyForm Noncommercial.
TinyMCE упоминался в чате game-origin как WYSIWYG-редактор сообщений
(`src/templates/standard/chat.tpl`).

**Что сделано:** при изучении chat.tpl выяснилось, что весь
TinyMCE-код находился в `{if[0 && isAdmin()]}`-блоках — выключенных
с момента импорта legacy. Реальный chat-input — обычный
`<input type="text">`. Удалено: 430 файлов библиотеки
(`projects/game-origin/public/js/tiny_mce/` ≈ 430 МБ) +
выключенные TinyMCE-блоки в шаблоне. Backend-strip не нужен:
сообщения в БД — plain text + bbcode (свой парсер в `Chat.class.php`).
Smoke: чат рендерится HTTP 200, 0 совпадений `tiny_mce` в HTML,
`/js/tiny_mce/tiny_mce.js` → 404.

### Gap 3.1 — атрибуция AI в коммитах

После разделительного коммита `1bae5ee3fd` (план 41) в репо появилось
~87 коммитов с trailer'ом `Co-Authored-By: Claude` и только 5 с
`Generated-with: Claude Code`. Соседние агенты не подхватили правило
attribution через проектный `.claude/settings.json`.

Юридически это не катастрофа — `docs/origin-rights.md` §6 объясняет,
что trailer не означает соавторства, — но факт повторного появления
говорит, что настройка не дошла до всех рабочих сессий.

**Что делать:** проверить, существует ли проектный `.claude/settings.json`
с блоком `attribution`. Если нет — создать (`.claude/` в .gitignore,
поэтому файл локальный per-developer; проектный работает только если
не gitignored или используется через alternative-механизм). Если
gitignore блокирует — повторно объяснить процедуру через
`docs/ops/claude-code-attribution.md` или применить альтернативу.

### Gap 4.1 — регистрация в game-origin без согласия ✅ ЗАКРЫТО (2026-04-28, план 50 Ф.2)

Уточнено пользователем: прямая регистрация в game-origin закрыта,
вход возможен только через handoff из portal (как и в game-nova).
Согласия на обработку ПДн и акцепт оферты собираются на portal
(планы 44, 47) и попадают в `user_consents` identity-сервиса до
момента handoff'а в game-origin.

Проверено grep'ом по `projects/game-origin/src/` и `public/`:
внешних вызовов `AccountCreator::registerUser` / `new AccountCreator(...)`
**нет** (единственное упоминание — комментарий в
`core/AjaxRequestHelper.abstract_class.php`).

В шапку `src/game/AccountCreator.class.php` добавлен
предупреждающий блок-комментарий: класс остаётся как helper для
lazy-create при handoff'е, добавлять публичные точки входа без
отдельного плана с чекбоксами + integration user_consents API
запрещено.

Чекбоксы согласия в game-origin не требуются — их роль выполняют
формы portal/AuthPage и game-nova/LoginScreen.

### Gap 5.1 — нет «12+» в game-origin ✅ ЗАКРЫТО (2026-04-27, план 50 Ф.3)

В portal и game-nova frontend компонент `AgeRating` есть. В
game-origin (PHP-templates) — нет. Если пользователь играет в
game-origin прямо — он не видит маркировку «12+».

**Что сделано:** в `src/templates/standard/layout.tpl` добавлен блок
`<div class="oxsar-footer">` с `.age-rating` (12+) перед закрытием
`</body>`; CSS-стили в `public/css/layout.css`.

### Gap 6.1 — UGC-модерация в game-origin покрывает только никнейм ✅ ЗАКРЫТО (2026-04-27, план 50 Ф.4)

Класс `Moderation` в game-origin применяется только в
`AccountCreator.class.php` при регистрации (проверка ника). Не
проверяются: чат, названия альянсов, описания планет, темы личных
сообщений.

**Что сделано:** в `Moderation.util.class.php` добавлен метод `mask()`
по образцу `MaskForbidden` из game-nova; применён в 7 UGC-точках
(12 вызовов): чат глобальный/альянсовый (`Chat`, `ChatAlly`), ЛС
тема+тело (`MSG`), название и тег альянса при создании
(`Alliance::foundAlliance` через `isForbidden()`-блокировку),
описания альянса (extern/intern/applicationtext), заявка на вступление,
глобальная рассылка альянса (тема + тело). Описания планеты в legacy
не существует как редактируемое поле — поверхность не найдена.
Notepad — приватный, не UGC. Verifications: 5 файлов `php -l` чисто +
unit-тест mask() 11/11 PASS.

### Gap 6.2 — нет «Пожаловаться» в game-origin ✅ ЗАКРЫТО (2026-04-28, план 50 Ф.5)

Кнопка «🚩 Пожаловаться» добавлена в game-origin. Поверхности:
- никнейм игрока в `playerstats` (через `UserList::formatRow` →
  поле `report` в loop);
- общий чат (`chat.tpl` — рядом с шапкой) и чат альянса
  (`chatally.tpl`);
- страница своего альянса (`allypage_own.tpl` — рядом с тегом).

Шаблон `_report_button.tpl` (CSS+HTML модалки+JS-обёртка `oxReport`)
подключён один раз в `layout.tpl`. PHP-helper `getReportButton($type, $id)`
в `Functions.inc.php` рендерит `<button>` с inline-onclick.

POST идёт на portal-backend (`PORTAL_BASE_URL` из `consts.php`,
default `https://oxsar-nova.ru` в prod, `http://localhost:8090` в dev) —
тот же централизованный endpoint `/api/reports`, что у game-nova
(plan 56). JWT берётся серверно из cookie `oxsar-jwt` (HttpOnly +
SameSite=Strict не отдаст её на cross-origin) и embed'ится в JS как
строка для `Authorization: Bearer`-header'а.

Origin `https://origin.oxsar-nova.ru` добавлен в
`ALLOWED_ORIGINS` portal-backend (`deploy/docker-compose.multiverse.yml`).

### Gap 7.1 — нет ссылок на юр-документы в game-origin footer ✅ ЗАКРЫТО (2026-04-27, план 50 Ф.6)

В footer'е game-origin нет ссылок на оферту, правила, refund,
privacy. Если пользователь играет напрямую — он не видит юр-документы.

**Что сделано:** в `src/templates/standard/layout.tpl` добавлен блок
`.legal-links` с absolute-URL на `https://oxsar-nova.ru/{offer,game-rules,refund,privacy}`,
все ссылки `target="_blank" rel="noopener"`.

### Gap 10.1 — copyright UnitPoint в шаблонах game-origin ✅ ЗАКРЫТО (2026-04-27, план 50 Ф.7)

В 107 шаблонах game-origin (фактически, не 10+, как изначально
оценивалось) стояла шапка
`Copyright (c) 2009-2010 UnitPoint <support@unitpoint.ru>` —
копирайт ликвидированной компании. По стратегии B3 (план 41) проект
позиционируется как новое самостоятельное произведение, и копирайт
должен быть автора-правообладателя.

Дополнительно email `support@unitpoint.ru` — публичная утечка
реквизита.

**Что сделано:** во всех 107 `.tpl` шапка заменена на
`Oxsar https://oxsar-nova.ru` + `Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.`.
Email `support@unitpoint.ru` в `support.tpl` (регламент техподдержки)
заменён на `gibesapiselfbab@hotmail.com`. Текст самого регламента
поддержки (упоминания «ООО Юнит Поинт»/«Оператор» в `support.tpl`)
**не переписан** — это полноценный юр-документ ликвидированной
компании, требует отдельного плана (новый регламент техподдержки или
удаление страницы).

## Сводка

**Подпроектов с полной обвязкой:** identity, portal, game-nova, billing
(все требования закрыты).

**Подпроект с пробелами:** game-origin — сосредоточение всех
найденных gap'ов (LGPL TinyMCE, отсутствие 152/436/149-ФЗ-обвязки в
legacy-вёрстке, copyright UnitPoint).

**Атрибуционная проблема:** соседние агенты продолжают писать
`Co-Authored-By: Claude` (gap 3.1).

## Что обновлено / создано в результате аудита

См. изменения в:
- (если будут созданы) `docs/plans/50-game-origin-legal-fix.md` —
  сводный план закрытия gap'ов 2.1, 4.1, 5.1, 6.1, 6.2, 7.1, 10.1
  в game-origin.
- (если будет расширен) `docs/plans/40-license-audit.md` —
  добавление PHP/Composer-проверки (gap 1.1).
- (если будет расширен) `docs/plans/41-origin-rights.md` —
  напоминание про trailer (gap 3.1).

## История аудитов

| Дата | Подпроектов | Gap'ов найдено | Скорректировано планов | Создано планов |
|---|---|---|---|---|
| 2026-04-27 | 5 | 8 | TBD | TBD (по решению автора) |
| 2026-04-27 (повторно, после плана 50 Ф.3+Ф.6+Ф.7) | 5 | 5 (закрыты 5.1, 7.1, 10.1; остаются 1.1, 2.1, 3.1, 4.1, 6.1, 6.2 — 6 шт., но 1.1+3.1 — общесистемные, не game-origin) | план 50 | — |
| 2026-04-27 (повторно, после плана 50 Ф.4) | 5 | 4 (закрыт 6.1; остаются 1.1, 2.1, 3.1, 4.1, 6.2) | план 50 | — |
| 2026-04-27 (повторно, после плана 50 Ф.1) | 5 | 3 (закрыт 2.1; остаются 1.1, 3.1, 4.1, 6.2 — но 1.1, 3.1 общесистемные) | план 50 | — |
| 2026-04-28 (повторно, после плана 50 Ф.5) | 5 | 2 (закрыт 6.2; остаются 1.1, 3.1, 4.1 — но 1.1, 3.1 общесистемные, 4.1 ждёт решения о direct-регистрации в game-origin) | план 50 | — |
