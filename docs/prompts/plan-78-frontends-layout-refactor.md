# Промпт: выполнить план 78 (раскладка фронтов + game-legacy-php)

**Дата создания**: 2026-04-28
**План**: [docs/plans/78-frontends-layout-refactor.md](../plans/78-frontends-layout-refactor.md)
**Зависимости**: блокирует план 72 (новый origin-фронт).
**Объём**: ~1-2 часа агента, ~440 правок путей, 2 коммита.

---

```
Задача: выполнить план 78 — раскладка фронтов внутри game-nova/
и переименование legacy-PHP.

КОНТЕКСТ:

Сейчас в репо асимметрия и путаница:

1. projects/game-nova/frontend/ — единственный фронт (nova-стиль).
   План 72 хотел положить новый origin-фронт в
   projects/game-origin/frontend/ — соседний проект. Это плохо:
   один backend (game-nova) обслуживает оба фронта, имена врут.

   Решение: оба фронта живут под game-nova/frontends/
   (мн. ч.) — nova/ для uni01/uni02 и origin/ для вселенной origin.

2. projects/game-origin-php/ — clean-room rewrite legacy oxsar2 на
   PHP. Имя путает: звучит как «origin-фронт на PHP», хотя это
   legacy. После создания нового origin-фронта (план 72) путаница
   усилится.

   Решение: переименовать в projects/game-legacy-php/.
   Терминологически (memory feedback_legacy_origin_terminology):
   legacy = d:\Sources\oxsar2 + clean-room rewrite в репо;
   origin = новая вселенная на game-nova-backend.

ВАЖНО ПРО ТЕРМИНОЛОГИЮ:
- nova / origin / legacy — имена в путях, идентификаторах, кодовых
  ссылках. Строго.
- modern — допустимо как ПРИЛАГАТЕЛЬНОЕ-эпоха в текстах
  («modern-числа», «modern-эпоха») для противопоставления legacy.
  В путях/именах файлов НЕ использовать.

ПЕРЕД НАЧАЛОМ:

1) git status --short. cat docs/active-sessions.md. План 78 нужно
   запускать в одиночку — он трогает Makefile/CI/CLAUDE.md/Dockerfile,
   которые задевают любую другую сессию.

2) Прочитай ПОЛНОСТЬЮ:
   - docs/plans/78-frontends-layout-refactor.md (твоё ТЗ)
   - docs/research/origin-vs-nova/roadmap-report.md «Часть I.5»
     R0-R15 (особенно R7 — backward compat не требуется до плана 74)
   - docs/plans/75-rename-game-origin-to-php.md (образец safe-rename)

3) Прочитай выборочно:
   - Makefile (frontend-* таргеты)
   - .github/workflows/ci.yml (frontend-job paths)
   - deploy/docker-compose.yml + deploy/docker-compose.e2e.yml
     (где упоминается frontend и game-origin-php)
   - projects/game-nova/frontend/Dockerfile
   - CLAUDE.md секция «Структура»

4) Добавь свою строку в docs/active-sessions.md:
   | <N> | План 78 frontends + legacy-PHP rename | projects/, Makefile, .github/workflows/, CLAUDE.md, deploy/, docs/ | <дата-время> | refactor(repo): план 78 |

   (Slot — следующее натуральное число без ведущих нулей; смотри
    активную таблицу в active-sessions.md, бери N+1 от последней
    использованной строки.)

ЧТО НУЖНО СДЕЛАТЬ:

### Ф.1. Переносы папок

git mv projects/game-nova/frontend projects/game-nova/frontends/nova
git mv projects/game-origin-php projects/game-legacy-php

ПРОВЕРЬ что git распознал ОБА mv как rename, не как delete+add:
git status --short — должны быть строки 'R  <old> -> <new>'.
Если delete+add — разберись почему (обычно слишком много правок в
файлах внутри; mv делать ДО любых правок содержимого).

### Ф.2. Сборка / деплой / код (замена 1: frontend → frontends/nova)

ВАЖНО: каждое упоминание `projects/game-nova/frontend` → 
`projects/game-nova/frontends/nova`. НЕ затрагивай portal/frontend
или admin-frontend — это другие проекты.

Файлы для обновления:
- Makefile (таргеты frontend-run, frontend-build, frontend-test)
- .github/workflows/ci.yml (paths/working-directory)
- deploy/docker-compose.yml (context для frontend-сервиса)
- deploy/docker-compose.e2e.yml (context + volumes)
- deploy/Dockerfile.playwright (COPY/WORKDIR)
- projects/game-nova/frontends/nova/Dockerfile (внутри проверить
  относительные пути — переехал вместе с папкой)
- .gitignore (записи node_modules/dist)
- CLAUDE.md (секция «Структура»)
- projects/game-nova/frontends/nova/src/components/feedback/feedback.ts
  (упоминание пути в комментарии — поиск точечный)

### Ф.3. Сборка / код (замена 2: game-origin-php → game-legacy-php)

ВАЖНО: 355 упоминаний в 67 файлах. Большинство — документация и
промпты, find-replace. Но ЕСТЬ Go-код и shell-скрипты с путями —
их обновить отдельно с проверкой компиляции.

**Go-код** (упоминают путь в комментариях/тестовых данных):
- projects/game-nova/backend/internal/event/handlers.go
- projects/game-nova/backend/internal/origin/alien/doc.go
- projects/game-nova/backend/internal/origin/alien/golden_test.go
- projects/game-nova/backend/internal/origin/economy/golden_test.go
- projects/game-nova/backend/cmd/tools/import-legacy-balance/main.go

После замены: go build ./... + go vet ./... в каждом модуле
(game-nova/backend, identity/backend, billing/backend,
admin-bff, portal/backend) — должно собираться зелёным.

**Shell/PHP скрипты внутри переехавшей папки**:
- projects/game-legacy-php/tools/*.sh — проверить пути (относительные
  ОК, абсолютные с GIT_ROOT-зависимостью — переписать).
- projects/game-legacy-php/tools/dump-alien-ai.php
- projects/game-legacy-php/tools/dump-balance-formulas.php
  (include-пути, обычно относительные — должны работать)
- projects/game-legacy-php/migrations/fixtures/README.md
- projects/game-legacy-php/src/core/util/Moderation.util.class.php

**Docker/CI**:
- Makefile — таргеты для legacy-PHP (если есть)
- deploy/docker-compose*.yml — context legacy-PHP сервиса (если есть)
- .github/workflows/ci.yml — paths legacy-job'ов
- .gitignore — node_modules/vendor для legacy-PHP

**Тонкость**: ПРОВЕРЬ docs/legacy/game-origin-access.md — переименовать
файл в game-legacy-access.md И обновить все ссылки (~3+).

### Ф.4. Документация (поиск-замена)

Замена 1: `projects/game-nova/frontend` → `projects/game-nova/frontends/nova`
- CLAUDE.md
- docs/plans/*.md (~12 файлов)
- docs/prompts/**/*.md (~10 файлов)
- docs/research/origin-vs-nova/*.md (~2 файла)
- docs/adr/*.md (1 файл)
- docs/ops/*.md (1 файл)

Замена 2: `game-origin-php` → `game-legacy-php`
- CLAUDE.md
- docs/plans/*.md (~30 файлов)
- docs/prompts/**/*.md (~15 файлов)
- docs/research/origin-vs-nova/*.md (~6 файлов)
- docs/adr/0010-universe-domain-naming.md
- docs/legacy/game-legacy-access.md (после переименования файла)
- docs/ops/legal-compliance-audit.md
- docs/simplifications.md (только текущие записи — НЕ исторические)
- docs/ai-debug-examples/*.md (2 файла)
- docs/ui/dev-log.md

**НЕ ТРОГАЕМ исторические записи**:
- docs/project-creation.txt — путь зафиксирован как факт момента
  написания (правило плана 55).
- В docs/simplifications.md — закрытые записи прошлых планов.
  Только активные/будущие правь.

### Ф.5. Smoke

- make frontend-run стартует Vite на 5173 из новой папки.
- make frontend-build собирает прод-бандл.
- make backend-test зелёный (golden_test.go читают legacy-PHP-tools
  по новым путям).
- make e2e — Playwright проходит базовый smoke (если в env есть
  Docker; если нет — пропусти и зафиксируй в коммите).
- docker compose -f deploy/docker-compose.yml up uni01-frontend собирает
  и стартует.
- Если есть docker-compose сервис legacy-PHP — стартует с context'ом
  projects/game-legacy-php.

### Ф.6. Финализация

- Шапка плана 78 ✅ всех фаз.
- Запись итерации в docs/project-creation.txt («78 — раскладка
  фронтов + переименование legacy-PHP»).
- Обновить docs/plans/72-remaster-origin-frontend-pixel-perfect.md
  и docs/prompts/plan-72-remaster-origin-frontend-pixel-perfect.md:
  путь projects/game-origin/frontend/ → projects/game-nova/frontends/origin/.
- Обновить docs/plans/76-remaster-nova-frontend-exchange-ui.md и
  его промпт.
- Обновить CLAUDE.md секцию «Структура».

### Ф.7. Memory (вне репо, опционально)

Если пишешь от лица главной сессии (не агент-делегат) — обновить
3 memory-записи:
- feedback_legacy_origin_terminology.md
- reference_legacy_docker.md
- project_origin_vs_nova.md

Замена `game-origin-php` → `game-legacy-php`.

══════════════════════════════════════════════════════════════════
ОБЯЗАТЕЛЬНЫЕ ПРАВИЛА R0-R15 — см. блок rules-r0-r15.md
СКОПИРУЙ СЮДА полностью раздел из docs/prompts/_blocks/rules-r0-r15.md

Этот план почти весь — refactor путей, не gameplay/feature. Из R-правил
особо важно:
- R0: НЕ менять modern-числа, формулы, баланс. Не должны измениться.
- R7: backward compat технических интерфейсов не требуется до плана 74,
  можно ломать пути в Go-импортах если возникнут (но переименование
  директорий не должно ломать импорты — игнорируй).
- R15: без MVP-сокращений. Переноси аккуратно, не оставляй grep-хвостов.
══════════════════════════════════════════════════════════════════

══════════════════════════════════════════════════════════════════
GIT-ИЗОЛЯЦИЯ ОТ ПАРАЛЛЕЛЬНЫХ СЕССИЙ — см. блок git-isolation.md
СКОПИРУЙ СЮДА полностью раздел из docs/prompts/_blocks/git-isolation.md

Этот план — особый случай: он трогает ОЧЕНЬ много общих файлов
(Makefile, CLAUDE.md, .gitignore, CI, docker-compose, ~50 файлов
docs/plans/, ~15 файлов docs/prompts/). Параллельная сессия с этим
планом ПРАКТИЧЕСКИ ГАРАНТИРУЕТ конфликт.

Прочитай docs/active-sessions.md ПЕРЕД стартом — если там есть ЛЮБАЯ
другая активная сессия, СПРОСИ пользователя «безопасно ли запускать
план 78 сейчас». Не стартуй автономно при наличии параллели.

CC_AGENT_PATHS для этого плана сложно ограничить (refactor широкий).
Установи как:

export CC_AGENT_PATHS="projects/ Makefile CLAUDE.md .gitignore .github/ deploy/ docs/ scripts/"

Это очень широко, но точнее задать невозможно — план рефакторит
всю верхушку репо. Перед каждым git commit ОБЯЗАТЕЛЬНО:
git status --short — посмотри что в индексе
git diff --cached --name-only — финальная проверка
Если видишь файлы с расширением .go которые НЕ в твоём списке
(handlers.go, doc.go, golden_test.go, main.go в import-legacy-balance,
feedback.ts) — это не твоё, git reset HEAD -- <file>.
══════════════════════════════════════════════════════════════════

КОММИТЫ:

Два коммита для blame-изоляции:

1) refactor(game-nova): frontend → frontends/nova (план 78 Ф.1.1)
   - git mv frontend → frontends/nova
   - все правки замены 1 (Makefile, CI, deploy, docs)

2) refactor(legacy): game-origin-php → game-legacy-php (план 78 Ф.1.2)
   - git mv game-origin-php → game-legacy-php
   - все правки замены 2 (Go-код, shell, PHP, docs, CI)
   - переименование docs/legacy/game-origin-access.md

Trailer в обоих: Generated-with: Claude Code

ВСЕГДА с двойным тире:
git commit -m "..." -- $CC_AGENT_PATHS

ЧЕГО НЕ ДЕЛАТЬ:

- НЕ менять содержимое файлов внутри переехавших папок (только
  пути в коммерческих ссылках на них).
- НЕ создавать frontends/origin/ — это работа плана 72.
- НЕ переименовывать `nova` в коде/типах/i18n — только раскладка папок.
- НЕ реорганизовывать portal/, admin-frontend/, identity/, billing/.
- НЕ вычищать прилагательное `modern` из текстов (это эпоха, не имя).
- НЕ трогать исторические записи в project-creation.txt /
  simplifications.md.
- НЕ забывай про -- в git commit (5-й прецедент в memory будет твой).
- НЕ запускайся параллельно с другими сессиями — сначала спроси.

УСПЕШНЫЙ ИСХОД:

- projects/game-nova/frontend/ больше не существует.
- projects/game-nova/frontends/nova/ работает (Vite/build/test зелёные).
- projects/game-origin-php/ больше не существует.
- projects/game-legacy-php/ работает (PHP-CLI tools зелёные, golden
  тесты проходят).
- 0 упоминаний `game-nova/frontend` в репо (кроме исторических
  записей в project-creation.txt).
- 0 упоминаний `game-origin-php` в репо (кроме исторических записей).
- go build / go test зелёные на всех модулях.
- Makefile-таргеты frontend-* работают.
- CI green (или подтверждение «локально не запускал»).
- План 72 промпт обновлён (путь origin-фронта).
- План 76 промпт обновлён.
- Шапка плана 78 ✅.
- Запись в docs/project-creation.txt — итерация 78.
- Удалена строка из docs/active-sessions.md.

Стартуй ТОЛЬКО если в active-sessions.md нет других активных слотов.
```
