# План 75: переименование `projects/game-origin/` → `projects/game-origin-php/`

**Дата**: 2026-04-28
**Статус**: Активный
**Зависимости**: нет блокирующих. Желательно выполнить **до** запуска
первого плана из ремастер-серии (64-74), чтобы не обновлять file:line
во всех артефактах исследования плана 62.
**Связанные документы**:
- [62-origin-on-nova-feasibility.md](62-origin-on-nova-feasibility.md) —
  исследование, которое произвело 9 артефактов с file:line по текущему
  пути.
- [docs/research/origin-vs-nova/roadmap-report.md](../research/origin-vs-nova/roadmap-report.md)
  — серия планов 64-74, которые будут писаться с file:line.

---

## Цель

Освободить путь `projects/game-origin/` под **новый React-фронт**
(будущая целевая реализация origin-вселенной), а текущую
PHP-реализацию переименовать в `projects/game-origin-php/`.

После этого:
- `projects/game-origin/` — целевая папка для ремастера. Внутри
  появится `frontend/` (новый React-клон по pixel-perfect стратегии
  плана 62).
- `projects/game-origin-php/` — временная legacy-реализация. Удалится
  после готовности нового фронта (один `git rm -r`).

Симметрично с `projects/game-nova/{backend,frontend}` — единый
паттерн именования игровых вселенных.

---

## Контекст

### Почему сейчас

После плана 62 у нас 9 артефактов исследования (`docs/research/origin-vs-nova/`),
содержащих сотни file:line ссылок на `projects/game-origin/...`.
Грядут планы 64-74 (ремастер) — они будут писаться с привязкой к этим
file:line. Каждый день откладывания добавляет ссылок.

Окно для переименования сейчас идеальное:
- Активной разработки origin нет (план 50 закрыт).
- Артефакты плана 62 свежие (один батч find-replace закроет всё).
- Ремастер ещё не начат — будущие планы сразу пишутся с правильным
  путём.

### Альтернатива «оставить как есть»

Если не переименовать сейчас:
- Через несколько месяцев придётся делать **два** переименования
  одновременно: `game-origin → game-origin-php` + новый `game-origin/`
  под фронт. Удвоенная сложность, удвоенный риск.
- Или жить с асимметрией: `game-origin/` это PHP, а новый React живёт
  где-то в `game-origin-frontend/` или подобном — путаница.

---

## Что меняем

### 1. Сам move

```bash
git mv projects/game-origin projects/game-origin-php
```

Это базовая операция. Дальше — find-replace всех ссылок.

### 2. Find-replace по проекту

Полный grep по `projects/game-origin` (без префикса `-php`) и замена
на `projects/game-origin-php`:

```bash
grep -rl "projects/game-origin" \
  --include="*.md" --include="*.txt" --include="*.yml" \
  --include="*.yaml" --include="*.json" --include="*.go" \
  --include="*.ts" --include="*.tsx" --include="*.sh" \
  --include="*.php" --include="Makefile" --include="Dockerfile*" \
  | grep -v node_modules
```

**Категории затронутых файлов:**

#### 2.1. Документация

- `CLAUDE.md` — упоминания путей (структура проекта).
- `docs/plans/*.md` — планы 37, 41, 43, 50, 60, 62, 63 (и более ранние,
  если упоминают origin).
- `docs/research/origin-vs-nova/*.md` — все 9 артефактов плана 62
  (большинство file:line здесь). Это **самая большая часть работы**.
- `docs/legacy/game-origin-access.md` — про запущенный origin.
- `docs/ops/legal-compliance-audit.md` — gap'ы game-origin.
- `docs/ops/license-audit.md` — Composer-аудит origin.
- `docs/ops/ugc-moderation.md` — упоминания origin.
- `docs/release-roadmap.md` — план ремастера.
- `docs/project-creation.txt` — исторический дневник (ВНИМАНИЕ:
  исторические записи **не трогать**, оставить старый путь как
  факт момента; только добавить новую запись итерации 75).
- `docs/simplifications.md` — если упоминает origin.
- Прочие docs/ — по grep.

#### 2.2. Deploy / scripts

- `deploy/docker-compose*.yml` — пути volumes, build context,
  Dockerfile.
- `deploy/*.conf`, `deploy/Caddyfile` и подобное.
- `scripts/*.sh` — bash-скрипты.
- `Makefile` — таргеты origin.
- Любые корневые `Dockerfile.*`.

#### 2.3. Код

- `projects/admin-bff/...` — если есть upstream-конфиг с путём origin
  (вряд ли — admin-bff проксирует на HTTP, не на ФС).
- `projects/portal/backend/...` — если есть пути к origin.
- `projects/identity/backend/...` — то же.
- `projects/game-nova/backend/...` — упоминания (CORS handoff к origin
  и т.п.).
- `projects/game-origin-php/src/...` — внутренние include'ы PHP
  (большинство относительные через `APP_ROOT_DIR`, но проверить).
- Конфиги PHP внутри origin: `config/consts.php`,
  `config/bd_connect_info.php`, `bootstrap.php` — если содержат
  абсолютные пути.

### 3. CLAUDE.md — обновить структуру

В разделе «Структура» поправить пример с `projects/game-nova/...`:
добавить парную пару `projects/game-origin-php/...` (PHP) +
зарезервированную `projects/game-origin/` (целевая).

---

## Чего НЕ делаем

- **Не трогаем исторические записи в `docs/project-creation.txt`** —
  старые итерации фиксируют состояние мира на момент написания.
  Замена path в них исказит историю. Только новая запись
  «итерация 75 — переименование».
- **Не трогаем git-историю** (никаких filter-branch / filter-repo).
  `git mv` достаточно.
- **Не создаём новую `projects/game-origin/`** в этом плане — она
  появится при создании первого React-кода (план N из ремастер-серии).
  Здесь только освобождаем путь.
- **Не меняем содержимое migrations / SQL** — путь к файлам внутри
  origin-php не критичен, миграции продолжат работать.
- **Не трогаем `d:\Sources\oxsar2\`** (оригинальный legacy с ext/) —
  внешний референс, не часть нашего репо.

---

## Этапы

### Ф.1. Move директории

```bash
git mv projects/game-origin projects/game-origin-php
```

После этого `git status` показывает массовый renamed (со 100% match —
git распознаёт переименование без правок содержимого).

### Ф.2. Find-replace в документации

Полный grep + замена в `docs/`. Особое внимание:
- `docs/research/origin-vs-nova/*.md` — здесь сотни ссылок.
- Не трогать исторические записи в `docs/project-creation.txt`.

Команда (sed, аккуратно):

```bash
grep -rl "projects/game-origin" docs/ \
  | grep -v "project-creation.txt" \
  | xargs sed -i 's|projects/game-origin/|projects/game-origin-php/|g'
```

(адаптировать под Windows-окружение и BSD/GNU sed разницу)

После — verify: `grep -rn "projects/game-origin/" docs/` должно
возвращать только то, что ожидаемо (исторические записи в
project-creation.txt, и опечатки если есть).

### Ф.3. Find-replace в deploy / scripts

То же для `deploy/`, `scripts/`, `Makefile`, корневые `Dockerfile.*`.

После — `docker-compose config` (или аналог) должен парсить YAML
без ошибок.

### Ф.4. Find-replace в коде (минимум)

Большинство кода обращается к origin через HTTP, не через ФС.
Но проверить:
- `grep -rn "game-origin" projects/ --include="*.go" --include="*.ts" --include="*.tsx"`
- Конфиги PHP внутри `projects/game-origin-php/` — если есть
  абсолютные пути типа `define('GAME_ORIGIN_DIR', '/projects/game-origin/...')`.

### Ф.5. CLAUDE.md

Обновить раздел «Структура»:

```
- projects/game-nova/backend  — Go entry points + домены
- projects/game-nova/frontend — React фронт современной вселенной
- projects/game-origin-php/   — legacy PHP реализация origin (clean-room
                                rewrite, на удаление после готовности
                                нового фронта)
- projects/game-origin/       — зарезервировано под новый React-фронт
                                origin (план N из ремастер-серии 64-74)
- ...
```

### Ф.6. Verify

- `git diff --stat` показывает в основном rename (100% match) +
  правки file:line в docs.
- `grep -rn "projects/game-origin/" .` (без `-php` суффикса) —
  должно остаться только:
  · исторические записи в `docs/project-creation.txt` (ожидаемо)
  · комментарии «зарезервировано под новый фронт» (ожидаемо)
- `make backend-run` (game-nova) запускается без ошибок.
- `docker-compose config` парсится.
- Если запускается origin-PHP-стек на :8092 — он должен подняться
  с новым путём.

### Ф.7. Финализация

1. Обновить шапку плана 75 — статус «✅ Завершён <дата>».
2. Запись в `docs/project-creation.txt` — итерация 75 (хронология).
3. Обновить `docs/research/origin-vs-nova/roadmap-report.md` — отметить
   что путь `game-origin/` теперь свободен под новый фронт.
4. Коммит: `refactor(repo): projects/game-origin → projects/game-origin-php (план 75)`.

---

## Тестирование

- `git status` после Ф.1 показывает чистый rename, не два независимых
  add+delete. Если git не распознал переименование — что-то не так.
- После Ф.6:
  - `make test` (если работает в текущем окружении).
  - `docker-compose config` без ошибок.
  - Smoke origin на :8092 (если поднимаешь стек).

---

## Объём

- Один большой коммит, ~50-100 файлов с правкой пути (в основном
  docs).
- Сама операция `git mv` — одна команда.
- Время выполнения: **~1-2 часа агента** в активном темпе.

---

## Когда запускать

**Прямо сейчас**, до запуска первого плана из ремастер-серии (64-74).

Желательно — отдельной сессией без параллельных, чтобы не было
конфликтов в git. Если параллельные есть — координация через
поимённый `git add` (см. CLAUDE.md, правило `feedback_parallel_session_check.md`).

---

## Известные риски

| Риск | Митигация |
|---|---|
| sed съел больше, чем надо (например, `game-origin-something` где это другое имя) | Точный паттерн `projects/game-origin/` (с слэшем после) — не зацепит `game-origin-php/`, `game-origin-frontend/`, и т.п. |
| Сломался docker-compose / Dockerfile | Ф.3 + verify: `docker-compose config` парсит YAML. |
| Сломались PHP-include внутри origin | Большинство относительные через `APP_ROOT_DIR`. Если что-то абсолютное — Ф.4 поправит. |
| Сломались артефакты плана 62 (file:line) | Ф.2 — массовая замена в `docs/research/origin-vs-nova/*.md`. После — verify через grep. |
| git не распознал переименование, делает add+delete | `git mv` всегда даёт rename. Если diff показывает иначе — проверить, не было ли промежуточных правок. |
| Параллельные сессии трогают game-origin одновременно | Перед стартом — `git status --short` + спросить пользователя. |

---

## Что после плана 75

- `projects/game-origin/` свободна.
- Ремастер-серия 64-74 запускается с уже правильными путями.
- Когда новый React-фронт будет готов — он создаётся в
  `projects/game-origin/frontend/` без дополнительных переименований.
- После полной готовности и перевода игроков — `git rm -r
  projects/game-origin-php/` одной операцией.

---

## References

- План 62 — источник file:line, которые будут массово обновлены.
- План 43 — clean-room rewrite PHP, после которого образовалась
  текущая `game-origin/`.
- CLAUDE.md, секция «Структура» — будет обновлена в Ф.5.
