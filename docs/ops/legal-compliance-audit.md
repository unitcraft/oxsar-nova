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
| auth | Go 1.23 | `backend/go.mod` |
| billing | Go 1.23 | `backend/go.mod` |
| game-nova | Go 1.23 + TS (npm) | `backend/go.mod`, `frontend/package.json` |
| game-origin | PHP 8.3 (clean-room) | `composer.json` (без внешних зависимостей) |
| portal | Go 1.23 + TS (npm) | `backend/go.mod`, `frontend/package.json` |

Других стеков (Rust/Python/Java) нет. Manifest-обнаружение нашло те же
проекты, что и `ls projects/` — синхронизация полная.

## Чек-лист соответствия

| Требование | auth | billing | game-nova | game-origin | portal | Статус |
|---|---|---|---|---|---|---|
| Лицензии Go-зависимостей (CI license-check) | ✅ | ✅ | ✅ | — | ✅ | ОК |
| Лицензии npm production-зависимостей | — | — | ✅ | — | ✅ | ОК |
| Лицензии Composer-зависимостей | — | — | — | ⚠️ N/A | — | См. gap 1.1 |
| Отсутствие GPL/AGPL/LGPL в исходниках | ✅ | ✅ | ✅ | ⚠️ | ✅ | См. gap 2.1 |
| Атрибуция AI в коммитах (новый trailer) | — | — | — | — | — | См. gap 3.1 |
| Чекбокс согласия в форме регистрации | ✅ | — | ✅ | ⚠️ | ✅ | См. gap 4.1 |
| Миграция user_consents (или эквивалент) | ✅ | — | — | — | — | ОК (auth централизован) |
| Endpoint `DELETE /users/me` | ✅ | — | — | — | — | ОК (auth централизован) |
| Возрастная маркировка «12+» | — | — | ✅ | ⚠️ | ✅ | См. gap 5.1 |
| UGC blacklist при создании контента | ✅ | — | ✅ | ⚠️ | — | См. gap 6.1 |
| Кнопка «Пожаловаться» в UI | — | — | ✅ | ⚠️ | — | См. gap 6.2 |
| Чекбокс акцепта оферты | ✅ | — | ✅ | ⚠️ | ✅ | См. gap 4.1 (то же место) |
| Ссылки на оферту/правила/refund/privacy в UI | — | — | — | ⚠️ | ✅ | См. gap 7.1 |
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

### Gap 2.1 — LGPL-зависимость TinyMCE в game-origin

В `projects/game-origin/public/js/tiny_mce/` лежит **TinyMCE под
LGPL 2.1**. LGPL — copyleft, несовместим с PolyForm Noncommercial.
TinyMCE используется в чате game-origin как WYSIWYG-редактор
сообщений (см. `src/templates/standard/chat.tpl`).

Объём — несколько сотен мегабайт legacy-вёрстки 2010-х годов.

**Что делать:** обновить TinyMCE до версии 6+ (под MIT с 2014 г.) или
заменить на другой редактор (Quill, CodeMirror — оба MIT). Создать
отдельный план под эту замену; до его выполнения — game-origin не
готова к публичному запуску с PolyForm.

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

### Gap 4.1 — регистрация в game-origin без согласия

В `projects/game-origin/src/game/AccountCreator.class.php` нет
упоминаний согласия на ПДн или акцепта оферты. Если пользователь
может зарегистрироваться напрямую в game-origin (без захода через
portal с handoff page) — у этой регистрации нет согласия по 152-ФЗ
и оферте.

**Требует уточнения:** в плане 36 lazy-create user добавлен в
**game-nova middleware**, но не в game-origin. Нужно проверить, можно
ли в принципе зарегистрироваться напрямую в game-origin или вход
работает только через handoff page из portal с уже подписанными
согласиями.

**Что делать:** уточнить у автора. Если регистрация в game-origin
возможна напрямую — добавить чекбоксы согласия в форму или закрыть
прямую регистрацию (только через portal). Если регистрация только
через handoff — задокументировать это в комментариях
`AccountCreator.class.php` и в плане.

### Gap 5.1 — нет «12+» в game-origin

В portal и game-nova frontend компонент `AgeRating` есть. В
game-origin (PHP-templates) — нет. Если пользователь играет в
game-origin прямо — он не видит маркировку «12+».

**Что делать:** добавить блок «12+» в `src/templates/standard/layout.tpl`
(footer) game-origin. Простая HTML-вставка, ~5 строк.

### Gap 6.1 — UGC-модерация в game-origin покрывает только никнейм

Класс `Moderation` в game-origin применяется только в
`AccountCreator.class.php` при регистрации (проверка ника). Не
проверяются: чат, названия альянсов, описания планет, темы личных
сообщений.

**Что делать:** расширить применение `Moderation::isForbidden()` на
все UGC-точки в game-origin. Объём — точечные правки в
chat-handler'е, alliance-create, message-send и т.п.

### Gap 6.2 — нет «Пожаловаться» в game-origin

В game-origin не реализована кнопка «Пожаловаться» (есть только
«assault report» — это про боевые отчёты, не про жалобы на
пользователей). Жалобы из game-origin не попадают в централизованный
admin-flow плана 46.

**Что делать:** добавить компонент жалобы в legacy-templates
game-origin, отправляющий POST на тот же endpoint `/api/reports`,
что и game-nova. Альянс/чат/никнейм должны иметь кнопку. Объём —
~50–100 строк в Smarty-шаблонах + JS-обёртка.

### Gap 7.1 — нет ссылок на юр-документы в game-origin footer

В footer'е game-origin нет ссылок на оферту, правила, refund,
privacy. Если пользователь играет напрямую — он не видит юр-документы.

**Что делать:** добавить блок ссылок в `layout.tpl` game-origin,
указывающий на portal-URL'ы (`/offer`, `/game-rules`, `/refund`,
`/privacy`).

### Gap 10.1 — copyright UnitPoint в шаблонах game-origin

В 10+ шаблонах game-origin стоит шапка
`Copyright (c) 2009-2010 UnitPoint <support@unitpoint.ru>` —
копирайт ликвидированной компании. По стратегии B3 (план 41) проект
позиционируется как новое самостоятельное произведение, и копирайт
должен быть автора-правообладателя.

Дополнительно email `support@unitpoint.ru` — публичная утечка
реквизита.

**Что делать:** заменить на `Copyright (c) 2026 oxsar-nova authors`
в этих шаблонах, удалить email. Простая поисковая замена, ~10 файлов.

## Сводка

**Подпроектов с полной обвязкой:** auth, portal, game-nova, billing
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
