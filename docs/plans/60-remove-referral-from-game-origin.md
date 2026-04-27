# План 60: Удаление legacy реферальной системы из game-origin

**Дата**: 2026-04-27
**Статус**: Активный
**Зависимости**: нет блокирующих. Не зависит от плана 58 (валюта) или
плана 59 (новая реферальная программа). План 59 будет реализовывать
реферальную программу с нуля на современном стеке (portal-backend +
admin-bff + admin-frontend), legacy-реализация в game-origin не
переиспользуется.
**Связанные документы**:
- [59-referral-program.md](59-referral-program.md) — новая реферальная
  программа (после плана 58, делается с нуля).
- [50-game-origin-legal-fix.md](50-game-origin-legal-fix.md) — общая
  работа по legacy game-origin.

---

## Цель

Полностью удалить legacy реферальную систему из `projects/game-origin/`.

**Обоснование:**
- Новая реферальная программа (план 59) реализуется с нуля на portal-backend
  с современной архитектурой (антифрод, RBAC, admin-bff, оксариты как
  награды). Legacy-механика game-origin её не использует.
- Хранить мёртвый код увеличивает технический долг и поверхность ошибок:
  legacy `Referral.class.php` может неожиданно сработать или дать ссылки
  на несуществующие endpoint'ы.
- Legacy-реферальная система привязана к старой валютной модели
  («кредиты» в game-origin БД), которая после плана 58 теряет смысл.
- Удаление ничего не ломает — реферальная страница в game-origin не
  является обязательной для игрового цикла.

---

## Что удаляем

### 1. PHP-код

- `projects/game-origin/src/game/page/Referral.class.php` — основной
  класс страницы рефералов.
- Упоминания `Referral` в:
  - `projects/game-origin/src/game/Menu.class.php` — пункт меню;
  - `projects/game-origin/src/game/xml/Menu.xml` — XML-описание меню;
  - `projects/game-origin/src/game/page/Page.class.php` — роутинг;
  - `projects/game-origin/src/game/EventHandler.class.php` — если есть
    обработчики реферальных событий;
  - `projects/game-origin/src/game/cronjob/RemoveInactiveUser.php` —
    если есть логика очистки реферальных связей;
  - `projects/game-origin/src/game/AccountCreator.class.php` — если
    при регистрации проверяется реферальный код;
  - `projects/game-origin/src/game/Functions.inc.php` — общие
    функции (если есть `referralBonus()` и т.п.);
  - `projects/game-origin/src/core/legacy_payment/payment.inc.php` —
    если payment'ы давали реферальный bonus.

### 2. Шаблоны

- `projects/game-origin/src/templates/standard/referral.tpl` — основной
  шаблон страницы.
- Упоминания реферала в:
  - `projects/game-origin/src/templates/standard/before_content.tpl`;
  - `projects/game-origin/src/templates/standard/main.tpl`.
- Кэши: `projects/game-origin/src/cache/templates/standard/referral.cache.php`
  и любые другие связанные кэши (Smarty компилирует автоматически —
  при удалении исходного `.tpl` кэш инвалидируется).

### 3. БД

Найти таблицы/колонки реферальной системы (вероятные имена):
- `na_referral`, `na_referrals`, `referrals`;
- `na_user.referrer_id`, `na_user.referrer_userid`, `na_user.referral_*`;
- `na_user.invited_by`.

Для каждой найденной структуры:
- Если используется только реферальной системой и больше нигде —
  миграция `DROP TABLE` / `ALTER TABLE DROP COLUMN`.
- Если используется в других местах (аналитика, бухгалтерия) —
  оставить, удалить только PHP-обращения.

Точный список — определяется в Ф.1 (поиск) ниже.

### 4. Конфиги и i18n

- Реферальные строки в `na_phrases` (legacy таблица фраз) — не удаляем
  физически, но помечаем как deprecated. Альтернативно — `DELETE FROM
  na_phrases WHERE phrase_key LIKE 'referral_%'` если уверены что нигде
  больше не используется.
- Любые YAML/INI конфиги с настройками реферальной программы
  (`referral_bonus_amount`, `referral_credit_percent`) — удалить.

### 5. Тестовые данные

- `projects/game-origin/tools/compare-output/` — там есть `Referral.html`
  снапшоты. Они в gitignore? Если нет — удалить как часть очистки.
- В sample-данных (`apply-test-user-fixture.sh`, `snapshot-legacy-user.sh`)
  — если есть реферальные fixtures, убрать.

### 6. Документация

- `docs/legacy/game-reference.md` — если упоминает реферальную систему,
  пометить как «удалено в плане 60».
- В оферте (`docs/legal/offer.md`) — оставить упоминание реферальной
  программы (план 47 §X / план 59 после реализации). Это про **новую**
  программу, не legacy. Не трогаем.

---

## Чего НЕ делаем

- Не удаляем модель данных целиком если она имеет смысл вне реферальной
  системы (например, `users.invited_by` может быть полезным аналитическим
  полем — оставим решать в Ф.1).
- Не удаляем historical-записи в `docs/project-creation.txt` или
  `docs/ui/dev-log.md` — это исторический дневник.
- Не трогаем план 59 (новая реферальная программа) — он остаётся
  активным.
- Не трогаем шаблоны/код, не связанные с реферальной системой
  (только потому что в той же папке).
- Не реализуем «миграцию реферальных связей в новую систему» — эта
  legacy данные не переиспользуются (другая семантика, другая антифрод,
  другая валюта).

---

## Этапы

### Ф.1. Полный поиск всех реферальных артефактов

Систематический grep по всему проекту:

```bash
# Код PHP
grep -rln "реферал\|referral\|referer\|referrer\|invite_by\|invited_by" \
  projects/game-origin/src/ 2>/dev/null | grep -v cache/

# Шаблоны
grep -rln "referral\|реферал" projects/game-origin/src/templates/

# БД миграции (если есть)
grep -rln "referral\|referrer\|invited_by" \
  projects/game-origin/migrations/ 2>/dev/null

# Конфиги
grep -rln "referral\|реферал" projects/game-origin/config/ 2>/dev/null
```

Документировать результат как **inventory** в коммит-сообщении.

Также проверить **БД-схему** на dev-сервере:

```bash
docker exec game-origin-postgres psql -U postgres -d game_origin -c "\dt" \
  | grep -i ref
docker exec game-origin-postgres psql -U postgres -d game_origin -c "\d users" \
  | grep -i ref
```

### Ф.2. Удаление PHP-кода

- `git rm projects/game-origin/src/game/page/Referral.class.php`.
- В `Menu.class.php` / `Menu.xml` — удалить пункт меню.
- В `Page.class.php` — удалить роутинг (case `'Referral'` → удалить).
- В `EventHandler.class.php`, `cronjob/RemoveInactiveUser.php` — удалить
  реферальные обработчики (если есть).
- В `AccountCreator.class.php` — удалить логику применения реферального
  кода при регистрации (как и в Ф.2 плана 50 — у нас прямая регистрация
  закрыта, но если есть код применения реф-бонуса — убрать).
- В `Functions.inc.php` / `payment.inc.php` — удалить функции типа
  `applyReferralBonus()`, `getReferrer()`, `creditReferralBonus()`.

### Ф.3. Удаление шаблонов

- `git rm projects/game-origin/src/templates/standard/referral.tpl`.
- В `before_content.tpl`, `main.tpl` — удалить блоки/ссылки на
  «Рефералы».
- Кэши пересоберутся автоматически при следующем запуске.

### Ф.4. Удаление БД-структур

По результатам Ф.1 создать миграцию `projects/game-origin/migrations/0NNN_drop_referral.sql`:

```sql
-- Если есть отдельная таблица:
DROP TABLE IF EXISTS na_referral;

-- Если есть колонки в users:
ALTER TABLE na_user
  DROP COLUMN IF EXISTS referrer_id,
  DROP COLUMN IF EXISTS referral_bonus_paid,
  DROP COLUMN IF EXISTS invited_by;
```

Если какие-то колонки **нужны для аналитики** (например, чтобы знать
исторически, кто кого приглашал) — оставить, добавить комментарий
«deprecated, не используется логикой, для аналитики».

Прогнать миграцию в dev-БД, проверить что приложение работает.

### Ф.5. Удаление i18n и конфигов

- Если есть отдельный конфиг — удалить.
- Если в `na_phrases` есть реферальные строки — `DELETE FROM
  na_phrases WHERE phrase_key LIKE 'referral_%' OR phrase_key LIKE
  'invitation_%'` (через ту же миграцию из Ф.4).

### Ф.6. Smoke-тест

После удаления — запустить game-origin и пройти основные сценарии:
- Главная страница (Main).
- Профиль пользователя.
- Магазин / payment.
- Меню навигации (нет битых ссылок).
- Регистрация нового пользователя (если применимо).
- Через `bash projects/game-origin/tools/compare-with-legacy.sh`
  если работает — отличия от legacy ожидаемы (Referral исчез).

Никаких 500-х, никаких пустых страниц, никаких «class not found».

### Ф.7. Финализация

- `git status --short` → коммитим только своими файлами поимённо.
- Запись в `docs/project-creation.txt` — итерация 60.
- Коммит: `chore(game-origin): удалить legacy реферальную систему (план 60)`.

В сообщении коммита перечислить **все удалённые сущности** (файлы,
таблицы/колонки, шаблоны) — это inventory из Ф.1.

---

## Тестирование

- Smoke-тест по Ф.6.
- `php -l` (lint) на оставшихся PHP-файлах — не должно быть syntax
  errors из-за удалённых include'ов.
- `compare-with-legacy.sh` (если работает) — diff показывает только
  ожидаемые изменения (Referral.html отсутствует, но это OK).

---

## Объём

1 коммит, ~10–20 файлов изменено/удалено + миграция БД. Объём —
~30–60 минут работы агента.

---

## Когда запускать

Можно делать **прямо сейчас** — план не блокируется ничем.

Параллельно с другими работами в game-origin (например, план 50 Ф.5
«кнопка Пожаловаться») допустимо, но нужна синхронизация: оба плана
трогают `Menu.class.php` / `Menu.xml`. Лучше делать **последовательно**
с другими game-origin-задачами.

---

## Известные риски

| Риск | Митигация |
|---|---|
| Удалили колонку, которая используется неявно | Ф.1 — полный grep, smoke-тест в Ф.6. Если что-то найдётся — точечно восстановить. |
| Сломались pages с битыми ссылками на Referral в меню | Ф.6 smoke включает проверку всех меню-пунктов. |
| Legacy-данные «кто кого пригласил» теряются | Если **нужно** их сохранить для аналитики — сделать `pg_dump --table=na_referral` в Ф.4 перед `DROP`, положить в `oxsar-nova-private/legacy-archive/`. По умолчанию — данные не сохраняются (новая программа их не переиспользует). |

---

## Что после плана 60

После выполнения плана 60 в game-origin:
- Нет страницы `?go=Referral`.
- Нет реферального кода в форме регистрации (а её и нет, прямая
  регистрация закрыта по плану 50 Ф.2).
- Нет реферальных бонусов при пополнении.

Когда будет реализован план 59 (новая реферальная программа на
portal-backend), игроки во всех вселенных (включая game-origin)
получат реферальную страницу через **portal-frontend** (URL
`https://oxsar-nova.ru/profile/referrals`). Если в game-origin
нужно дать ссылку на эту страницу — добавить пункт меню
«Реферальная программа» с переходом на portal-URL (в новой вкладке).
Это можно сделать как **отдельный мини-шаг** после плана 59,
не в плане 60.
