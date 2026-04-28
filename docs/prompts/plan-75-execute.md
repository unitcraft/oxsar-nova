# Промпт: выполнить план 75 (переименование game-origin → game-origin-php)

**Дата создания**: 2026-04-28
**План**: [docs/plans/75-rename-game-origin-to-php.md](../plans/75-rename-game-origin-to-php.md)
**Применение**: вставить блок ниже в новую сессию Claude Code в
рабочей директории `d:\Sources\oxsar-nova`. Агент прочитает план 75
самостоятельно и выполнит переименование + массовый find-replace.
**Объём**: 1 большой коммит, ~50-100 файлов, ~1-2 часа.

**Желательно запускать без параллельных сессий** (массовое
переименование чувствительно к чужим конфликтам).

---

```
Задача: выполнить план 75 — переименование projects/game-origin/ →
projects/game-origin-php/ для освобождения пути под новый
React-фронт ремастера.

ВАЖНОЕ:
- Это рефакторинг путей. Затрагивает много файлов (docs, deploy,
  scripts, возможно код). Нужна аккуратность.
- Желательно без параллельных сессий — массовый find-replace
  чувствителен к чужим изменениям.

ПЕРЕД НАЧАЛОМ:

1) git status --short — должно быть ЧИСТО. Если есть чужие
   изменения от параллельных сессий, спроси пользователя
   приостановить их или подожди.

2) Прочитай ПОЛНОСТЬЮ:
   - docs/plans/75-rename-game-origin-to-php.md (твоё ТЗ —
     соблюдай каждую фазу Ф.1-Ф.7)
   - CLAUDE.md (правила: коммиты, языки, hooks)

3) Прочитай выборочно:
   - docs/research/origin-vs-nova/roadmap-report.md (контекст
     ремастера — ради чего освобождаем путь)
   - docs/plans/62-origin-on-nova-feasibility.md (источник
     многих file:line ссылок)

ЧТО НУЖНО СДЕЛАТЬ (по фазам плана 75):

Ф.1. Move директории

  git mv projects/game-origin projects/game-origin-php

После этого git status показывает массовый renamed (со 100% match —
git распознаёт переименование автоматически без правок содержимого).
Если git показывает add+delete вместо rename — что-то не так,
проверь.

Ф.2. Find-replace в документации

Полный grep + замена в docs/. Особое внимание:
- docs/research/origin-vs-nova/*.md — здесь сотни ссылок
  (артефакты плана 62), это самая большая часть работы.
- НЕ ТРОГАТЬ исторические записи в docs/project-creation.txt —
  они фиксируют состояние мира на момент написания, замена path
  исказит историю. Только новая запись «итерация 75» в Ф.7.

Поиск:
  grep -rn "projects/game-origin/" docs/ | grep -v "project-creation.txt"

Замена (адаптируй под Windows-окружение):
  - На Linux/macOS: sed -i 's|projects/game-origin/|projects/game-origin-php/|g'
  - На Windows: использовать Edit-tool с replace_all=true
    или PowerShell

КРИТИЧНО: использовать **точный паттерн с слэшем** —
"projects/game-origin/" (с финальным слэшем), чтобы НЕ зацепить
"projects/game-origin-php/" уже после первого этапа. Без слэша
зацепит и сломает.

Verify: grep -rn "projects/game-origin/" docs/ — должно
возвращать только:
- исторические записи в docs/project-creation.txt
- комментарии «зарезервировано под новый фронт» (если есть)

Ф.3. Find-replace в deploy / scripts

То же самое для:
- deploy/docker-compose*.yml — пути volumes, build context
- deploy/*.conf, deploy/Caddyfile (если есть)
- scripts/*.sh
- Makefile (если упоминает)
- корневые Dockerfile.* (если есть)

После Ф.3:
- docker-compose -f deploy/docker-compose.multiverse.yml config
  должен парсить YAML без ошибок (можно проверить только если
  Docker доступен).

Ф.4. Find-replace в коде (минимум)

Большинство кода обращается к origin через HTTP/CORS, не через
ФС. Но проверить:

  grep -rn "game-origin" projects/ --include="*.go" \
    --include="*.ts" --include="*.tsx" --include="*.php"

Особое внимание:
- projects/game-origin-php/config/consts.php — если есть
  абсолютные пути define('GAME_ORIGIN_DIR', ...)
- projects/game-origin-php/bootstrap.php или подобные
- projects/game-nova/backend/... — CORS allowed origins,
  handoff URLs

Большинство PHP-include внутри origin — относительные через
APP_ROOT_DIR. Должно быть ОК без правок.

Ф.5. CLAUDE.md

Обновить раздел «Структура»:

```
- projects/game-nova/backend  — Go entry points + домены
- projects/game-nova/frontend — React фронт современной вселенной
- projects/game-origin-php/   — legacy PHP реализация origin
                                (clean-room rewrite, на удаление
                                после готовности нового фронта)
- projects/game-origin/       — зарезервировано под новый
                                React-фронт origin (ремастер,
                                планы 64-74)
- projects/portal/...
```

Ф.6. Verify

- git diff --stat показывает в основном rename (100% match) +
  правки file:line в docs.
- grep -rn "projects/game-origin/" . (БЕЗ -php суффикса) —
  должно остаться только:
  · исторические записи в docs/project-creation.txt (ожидаемо)
  · комментарии «зарезервировано под новый фронт» (ожидаемо)
- Если возможно — make build / docker-compose config / etc.

Ф.7. Финализация

1. Обновить шапку плана 75 — статус «✅ Завершён <дата>».
2. Запись в docs/project-creation.txt — итерация 75 (хронология,
   что переименовано и зачем).
3. Обновить docs/research/origin-vs-nova/roadmap-report.md —
   отметить, что путь projects/game-origin/ теперь свободен под
   новый фронт.
4. Финальный коммит:
   refactor(repo): projects/game-origin → projects/game-origin-php (план 75)

КОММИТ:

Один большой коммит. Структура сообщения:

```
refactor(repo): projects/game-origin → projects/game-origin-php (план 75)

Освобождает путь projects/game-origin/ под целевую папку нового
React-фронта (ремастер, планы 64-74). Текущая PHP-реализация
переименована в projects/game-origin-php/ — будет удалена после
готовности нового фронта.

Затронуто:
- git mv projects/game-origin → projects/game-origin-php
- docs/research/origin-vs-nova/* — сотни file:line ссылок
- docs/plans/* (планы 37, 41, 43, 50, 60, 62, 63, 75 — где
  упоминается путь)
- docs/legacy/game-origin-access.md, docs/ops/*, docs/release-roadmap.md
- deploy/docker-compose.multiverse.yml + другие deploy
- CLAUDE.md — раздел «Структура»
- (если нужно) projects/*/...

Не трогали: docs/project-creation.txt (исторические записи),
git-историю (никаких filter-branch).

Generated-with: Claude Code
```

ВАЖНОЕ: GIT-БЕЗОПАСНОСТЬ:

- git mv для самого переименования (не rm + add).
- В коммит-сообщении НЕ забыть Generated-with: Claude Code
  (НЕ Co-Authored-By — git hook уберёт автоматически).
- Если параллельные сессии всё-таки начнутся — поимённый
  git add + git status --short перед commit.

ЧЕГО НЕ ДЕЛАТЬ:

- НЕ трогать исторические записи в docs/project-creation.txt
  (они фиксируют состояние мира).
- НЕ создавать projects/game-origin/ в этом плане — это для
  будущих планов ремастера.
- НЕ менять содержимое migrations / SQL / .tpl шаблонов —
  только пути в конфигах.
- НЕ трогать d:\Sources\oxsar2\ (внешний референс).
- НЕ делать filter-branch / filter-repo — только git mv.
- НЕ заменять "game-origin" без слэша — зацепит "game-origin-php"
  и сломает свою же работу.

ОЦЕНКА ОБЪЁМА:

1-2 часа работы. Если идёт сильно дольше — проверь, не пытаешься
ли вручную править то, что должен был сделать sed/Edit replace_all.
Если сильно быстрее — проверь полноту grep'а: пропущенный файл =
сломанная ссылка.

ИЗВЕСТНЫЕ РИСКИ (см. план 75, секция «Известные риски»):

- sed/replace зацепил больше, чем надо → точный паттерн
  "projects/game-origin/" (с слэшем).
- git не распознал rename → проверить через git status, должен
  быть R100.
- docker-compose сломался → docker-compose config в Ф.6.
- PHP-include внутри origin сломались → большинство относительные,
  но проверить config/consts.php в Ф.4.

УСПЕШНЫЙ ИСХОД:

- projects/game-origin-php/ существует, projects/game-origin/ нет.
- Все file:line ссылки в docs обновлены.
- docker-compose config / make build парсятся.
- 1 большой коммит, шапка плана 75 ✅.
- Готовность к запуску планов ремастера 64-74 без правки путей.

Стартуй.
```
