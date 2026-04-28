# Промпт: выполнить план 79 (рефакторинг раскладки deploy/)

**Дата создания**: 2026-04-28
**План**: [docs/plans/79-deploy-structure-refactor.md](../plans/79-deploy-structure-refactor.md)
**Зависимости**: должен идти ПОСЛЕ плана 78 (frontends + legacy-PHP).
**Объём**: ~3-4 часа агента, ~180 правок путей + ~80 строк base.yml,
3-4 коммита.

---

```
Задача: выполнить план 79 — рефакторинг раскладки deploy/ к
современным конвенциям (compose/base + override'ы, инфра/сервис
Dockerfile-разделение, configs/, examples/, scripts/).

КОНТЕКСТ:

Сейчас deploy/ — плоская папка с ~25 файлами:
- 9 docker-compose-файлов (postgres продублирован в 4, redis в 2)
- 8 Dockerfile'ов (mix сервисных и инфра-only)
- 4 конфига (nginx*2, prometheus, grafana/)
- 2 примера (.env, admin-ips)
- 1 скрипт (backup.sh)

Проблемы (отсюда план):
- postgres:16-alpine продублирован — правки версии в 4 местах
- сервисные Dockerfile'ы (admin-bff, admin-frontend, frontend-prod)
  оторваны от своего кода
- mix Dockerfile/compose/configs/scripts на одном уровне без правил
- backup.sh в deploy/ — это скрипт, ему место в scripts/

Этот план запускается ПОСЛЕ плана 78. К моменту старта 79:
- projects/game-nova/frontends/nova/ уже существует (план 78)
- projects/game-legacy-php/ уже существует (план 78)
- Все upstream правки в Makefile/CI/docs уже применены под план 78

Это ЕДИНСТВЕННАЯ активная сессия — refactor затрагивает Makefile,
CI, runbook'и, ВСЕ docker-compose. Параллельная сессия = конфликт.

ПЕРЕД НАЧАЛОМ:

1) git status --short. cat docs/active-sessions.md.
   Если есть ДРУГАЯ активная сессия — СТОП, спроси пользователя.

2) Прочитай ПОЛНОСТЬЮ:
   - docs/plans/79-deploy-structure-refactor.md (твоё ТЗ)
   - docs/research/origin-vs-nova/roadmap-report.md «Часть I.5» R0-R15
   - docs/ops/runbooks/docker-stacks.md (главный потребитель — 9 ссылок)
   - docs/ops/release-process.md (6 ссылок)
   - docs/ops/runbooks/backup-and-monitoring.md (8 ссылок)

3) Прочитай выборочно:
   - deploy/docker-compose.yml + deploy/docker-compose.multiverse.yml
     (16KB + 11KB — самые большие, нужно понять что выносить в base.yml)
   - deploy/docker-compose.e2e.yml + e2e.ports.yml
   - deploy/docker-compose.admin.yml + monitoring.yml + scaling.yml +
     yookassa-mock.yml + prod.yml (мельче)
   - Makefile (10 ссылок на deploy/...)
   - .github/workflows/ci.yml + admin-console.yml (CI потребитель)

4) Добавь свою строку в docs/active-sessions.md:
   | <N> | План 79 deploy/ refactor | deploy/, projects/<service>/Dockerfile, Makefile, .github/, scripts/, docs/ | <дата-время> | refactor(deploy): план 79 |

   (Slot — следующее натуральное число без ведущих нулей; смотри
    активную таблицу в active-sessions.md, бери N+1 от последней
    использованной строки.)

ЧТО НУЖНО СДЕЛАТЬ:

### Ф.1. Сборка base.yml

Цель: вынести общие сервисы в один файл, чтобы убрать дублирование.

1. Создай deploy/compose/base.yml с:
   - postgres:16-alpine (с volumes, healthcheck, ENV — самая полная версия
     из существующих docker-compose.yml)
   - redis:7-alpine
   - networks (default + internal если есть)
   - volumes (postgres-data, redis-data, etc.)

2. Все версии и параметры сверь между 4 копиями postgres в:
   - docker-compose.yml
   - docker-compose.e2e.yml
   - docker-compose.monitoring.yml (если есть)
   - docker-compose.multiverse.yml
   - docker-compose.yookassa-mock.yml
   
   Если параметры расходятся (разные ENV, разные порты, разные volumes) —
   ВЫБЕРИ полную версию из docker-compose.yml для base.yml, а остальные
   override-нюансы вынеси в соответствующий override-файл (e.g.
   `e2e.yml` может override port если он другой).

3. Если у postgres в e2e.yml — другая база/seed, оставь это в e2e.yml
   как override через postgres → environment override; БАЗОВЫЙ образ
   остаётся в base.yml.

### Ф.2. Перенос compose-файлов

git mv deploy/docker-compose.yml             deploy/compose/dev.yml
git mv deploy/docker-compose.prod.yml        deploy/compose/prod.yml
git mv deploy/docker-compose.e2e.yml         deploy/compose/e2e.yml
git mv deploy/docker-compose.e2e.ports.yml   deploy/compose/e2e.ports.yml
git mv deploy/docker-compose.multiverse.yml  deploy/compose/multiverse.yml
git mv deploy/docker-compose.scaling.yml     deploy/compose/scaling.yml
git mv deploy/docker-compose.admin.yml       deploy/compose/admin.yml
git mv deploy/docker-compose.monitoring.yml  deploy/compose/monitoring.yml
git mv deploy/docker-compose.yookassa-mock.yml deploy/compose/yookassa-mock.yml

ВНУТРИ каждого override:
- Убрать определения postgres/redis (теперь в base.yml).
- Убрать дублирующиеся networks/volumes (если они в base.yml).
- ОБНОВИТЬ relative-пути в `build:` (context) и `volumes:` (host paths)
  на новую глубину — теперь файл лежит на 1 уровень глубже.
  Было: `context: ../projects/game-nova/backend`
  Стало: `context: ../../projects/game-nova/backend`
- ПРОВЕРЬ работоспособность через `docker compose -f deploy/compose/base.yml
  -f deploy/compose/<file>.yml config` — должен парситься без ошибок.

### Ф.3. Перенос Dockerfile'ов

**Сервисные → рядом с кодом:**

git mv deploy/Dockerfile.admin-bff      projects/admin-bff/Dockerfile
git mv deploy/Dockerfile.admin-frontend projects/admin-frontend/Dockerfile
git mv deploy/Dockerfile.frontend-prod  projects/game-nova/frontends/nova/Dockerfile.prod

ВНУТРИ каждого Dockerfile проверь относительные пути COPY:
- Было: `COPY projects/admin-bff/. /app/` (из deploy/)
- Стало: `COPY . /app/` (теперь Dockerfile внутри сервиса)

И обнови соответствующие docker-compose-файлы:
- Было: `dockerfile: ../deploy/Dockerfile.admin-bff` + `context: ..`
- Стало: `dockerfile: Dockerfile` + `context: ../../projects/admin-bff`

**Инфра → в deploy/docker/** (с переименованием в стиле .Dockerfile):

git mv deploy/Dockerfile.migrate     deploy/docker/migrate.Dockerfile
git mv deploy/Dockerfile.playwright  deploy/docker/playwright.Dockerfile
git mv deploy/Dockerfile.testseed    deploy/docker/testseed.Dockerfile
git mv deploy/Dockerfile.prometheus  deploy/docker/prometheus.Dockerfile
git mv deploy/Dockerfile.grafana     deploy/docker/grafana.Dockerfile

ВНУТРИ каждого инфра-Dockerfile проверь относительные пути COPY
(если есть — теперь относительно deploy/docker/).

И обнови compose-файлы: `dockerfile: ../docker/migrate.Dockerfile`
(если context = `..`).

### Ф.4. Перенос конфигов и примеров

git mv deploy/nginx.admin.conf            deploy/configs/nginx.admin.conf
git mv deploy/nginx.frontend.conf         deploy/configs/nginx.frontend.conf
git mv deploy/prometheus.yml              deploy/configs/prometheus.yml
git mv deploy/grafana                     deploy/configs/grafana
git mv deploy/.env.multiverse.example     deploy/examples/.env.multiverse.example
git mv deploy/admin-ips.conf.example      deploy/examples/admin-ips.conf.example

Обнови volume-mounts в compose-файлах:
- Было: `- ../prometheus.yml:/etc/prometheus/prometheus.yml:ro`
- Стало: `- ../configs/prometheus.yml:/etc/prometheus/prometheus.yml:ro`

И nginx.frontend.conf — он используется в frontend-prod-Dockerfile
(переехал в projects/game-nova/frontends/nova/Dockerfile.prod):
- Возможно потребуется COPY ../../deploy/configs/nginx.frontend.conf
  с обновлённым относительным путём.

### Ф.5. Перенос скриптов

git mv deploy/backup.sh scripts/backup.sh

Обнови ссылки:
- В docs/ops/runbooks/backup-and-monitoring.md (8 ссылок).
- Если backup.sh упоминается в cron-файле/.github/workflows/ —
  тоже обнови.
- Если backup.sh использует относительные пути ВНУТРИ — он лежал в
  deploy/, теперь в scripts/, проверь что не упирается в cd
  ../<somewhere>.

### Ф.6. Обновление потребителей

**Makefile (10 ссылок):**

Замени все `docker compose -f deploy/docker-compose.X.yml` на
комбинации, и заведи таргеты-обёртки:

compose-up:           docker compose -f deploy/compose/base.yml -f deploy/compose/dev.yml up -d
compose-down:         docker compose -f deploy/compose/base.yml -f deploy/compose/dev.yml down
compose-logs:         docker compose -f deploy/compose/base.yml -f deploy/compose/dev.yml logs -f
compose-up-monitoring: ... + -f deploy/compose/monitoring.yml
compose-up-e2e:       ... + -f deploy/compose/e2e.yml -f deploy/compose/e2e.ports.yml
compose-up-prod:      docker compose -f deploy/compose/base.yml -f deploy/compose/prod.yml up -d
compose-up-multiverse: ... + -f deploy/compose/multiverse.yml
compose-up-admin:     ... + -f deploy/compose/admin.yml
compose-up-yookassa:  ... + -f deploy/compose/yookassa-mock.yml

ОСТАВЬ обратную совместимость существующих make-таргетов (dev-up,
backend-run, frontend-run и пр.) — они должны вызывать новые таргеты
внутри. То есть фасад тот же, начинка переехала.

**CI (.github/workflows/ci.yml + admin-console.yml — 14 ссылок):**

Замени все `-f deploy/docker-compose.X.yml` на комбинации
`-f deploy/compose/base.yml -f deploy/compose/<X>.yml`.

**Документация (~140 ссылок):**

Используй find-replace по этим файлам:
- docs/ops/runbooks/docker-stacks.md (КРИТИЧНЫЙ — 9 ссылок,
  оперативный runbook).
- docs/ops/runbooks/backup-and-monitoring.md (8 ссылок).
- docs/ops/release-process.md (6 ссылок).
- docs/ops/admin-access.md (7 ссылок).
- docs/plans/*.md (~30 файлов, ~80 ссылок).
- docs/prompts/*.md (~5 файлов).

**НЕ ТРОГАЕМ исторические записи**:
- docs/project-creation.txt — путь зафиксирован как факт момента.
- docs/simplifications.md — закрытые записи.

### Ф.7. Smoke (КРИТИЧНО — без этого refactor может сломать prod)

ОБЯЗАТЕЛЬНО:
1. `docker compose -f deploy/compose/base.yml -f deploy/compose/dev.yml config`
   — должен распарситься без ошибок. Все services видны.
2. `docker compose -f deploy/compose/base.yml -f deploy/compose/dev.yml build`
   — все сервисные образы собираются.
3. `make compose-up` — стартует postgres+redis+backend+frontend.
   Проверь docker compose ps — все health=healthy через ~30 сек.
4. Если есть docker, попробуй:
   - `make compose-up-monitoring` — добавляются prometheus+grafana,
     prometheus targets живые.
   - `make compose-up-e2e` — Playwright-стек встаёт.
   - `make compose-up-multiverse` — uni01+uni02 встают.
   - `make compose-up-admin` — admin-bff отвечает.
5. Backup-скрипт smoke: `bash scripts/backup.sh --dry-run` если
   поддерживает; иначе ручной первый шаг.

Если Docker недоступен в окружении — пропусти и зафиксируй в
коммите «smoke с Docker не выполнен — пользователь должен проверить».

### Ф.8. Финализация

- Шапка плана 79 ✅ всех фаз.
- Запись итерации в docs/project-creation.txt («79 — раскладка deploy»).
- Обновить CLAUDE.md секцию «Запуск» под новые таргеты.
- Обновить docs/ops/runbooks/docker-stacks.md под новую раскладку
  (это главный оперативный потребитель).

══════════════════════════════════════════════════════════════════
ОБЯЗАТЕЛЬНЫЕ ПРАВИЛА R0-R15 — см. блок rules-r0-r15.md
СКОПИРУЙ СЮДА полностью раздел из docs/prompts/_blocks/rules-r0-r15.md

Этот план — refactor инфры, не gameplay/feature. Из R-правил особо
важно:
- R0: НЕ менять modern-числа, формулы, баланс. Не должны измениться.
- R7: backward compat технических интерфейсов не требуется до плана 74,
  можно менять docker-compose-комбинации (это и есть план).
- R15: без MVP-сокращений. Перенеси аккуратно, проверь smoke.
══════════════════════════════════════════════════════════════════

══════════════════════════════════════════════════════════════════
GIT-ИЗОЛЯЦИЯ ОТ ПАРАЛЛЕЛЬНЫХ СЕССИЙ — см. блок git-isolation.md
СКОПИРУЙ СЮДА полностью раздел из docs/prompts/_blocks/git-isolation.md

Этот план — особый случай: трогает ВСЁ что связано с deploy/CI/runbook.
Параллельная сессия с этим планом ПРАКТИЧЕСКИ ГАРАНТИРУЕТ конфликт.

Прочитай docs/active-sessions.md ПЕРЕД стартом — если там есть ЛЮБАЯ
другая активная сессия, СПРОСИ пользователя «безопасно ли запускать
план 79 сейчас». Не стартуй автономно при наличии параллели.

CC_AGENT_PATHS:

export CC_AGENT_PATHS="deploy/ projects/admin-bff/Dockerfile projects/admin-frontend/Dockerfile projects/game-nova/frontends/nova/Dockerfile.prod scripts/backup.sh Makefile CLAUDE.md .github/ docs/"

Перед каждым git commit ОБЯЗАТЕЛЬНО:
git status --short
git diff --cached --name-only
══════════════════════════════════════════════════════════════════

КОММИТЫ:

3-4 коммита для blame-изоляции:

1) refactor(deploy): extract base.yml + compose в compose/ (Ф.1+Ф.2)
   - new deploy/compose/base.yml
   - 9 git mv compose/*.yml + правки внутри (убран postgres/redis,
     обновлены relative-пути)

2) refactor(deploy): сервисные Dockerfile'ы → projects/, инфра → docker/ (Ф.3)
   - 3 git mv сервисных Dockerfile'ов
   - 5 git mv инфра-Dockerfile'ов в deploy/docker/
   - правки compose-файлов под новые dockerfile-пути

3) refactor(deploy): configs/, examples/, scripts/backup.sh (Ф.4+Ф.5)
   - 6 git mv (конфиги, примеры, backup.sh)
   - правки volume-mounts в compose

4) refactor(deploy): обновить потребителей (Ф.6+Ф.8)
   - Makefile (новые compose-* таргеты + back-compat)
   - .github/workflows/*.yml
   - docs/ops/runbooks/, docs/ops/release-process.md, docs/plans/,
     docs/prompts/
   - CLAUDE.md
   - финальная запись в project-creation.txt

Trailer во всех: Generated-with: Claude Code
ВСЕГДА с двойным тире:
git commit -m "..." -- $CC_AGENT_PATHS

ЧЕГО НЕ ДЕЛАТЬ:

- НЕ переписывать содержимое compose-сервисов (только перетасовка +
  base extract).
- НЕ вводить k8s/helm/terraform/ansible — отложено до пост-запуска.
- НЕ переименовывать существующие сервисы внутри compose.
- НЕ удалять обратную совместимость Makefile-таргетов (dev-up
  должен продолжать работать).
- НЕ запускайся параллельно — refactor затрагивает общие пути.
- НЕ трогай исторические записи в project-creation.txt /
  simplifications.md.
- НЕ забывай про -- в git commit.

УСПЕШНЫЙ ИСХОД:

- deploy/compose/base.yml существует, содержит postgres+redis+
  networks+volumes.
- deploy/compose/{dev,prod,e2e,e2e.ports,multiverse,scaling,admin,
  monitoring,yookassa-mock}.yml — override-файлы без дублирования
  postgres/redis.
- deploy/docker/ содержит 5 инфра-Dockerfile'ов в стиле <name>.Dockerfile.
- deploy/configs/ содержит nginx.admin.conf, nginx.frontend.conf,
  prometheus.yml, grafana/.
- deploy/examples/ содержит .env.multiverse.example, admin-ips.conf.example.
- scripts/backup.sh существует.
- projects/admin-bff/Dockerfile, projects/admin-frontend/Dockerfile,
  projects/game-nova/frontends/nova/Dockerfile.prod существуют.
- deploy/Dockerfile.* и deploy/docker-compose.* (старые) НЕ существуют.
- Makefile compose-* таргеты работают.
- CI green (или ручная пометка).
- 0 упоминаний `deploy/docker-compose.X.yml` (плоской схемы) в
  активных файлах (только в исторических записях).
- Шапка плана 79 ✅.
- Запись в docs/project-creation.txt — итерация 79.
- Удалена строка из docs/active-sessions.md.

Стартуй ТОЛЬКО если в active-sessions.md нет других активных слотов.
```
