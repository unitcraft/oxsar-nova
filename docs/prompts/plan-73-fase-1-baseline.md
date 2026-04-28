# Промпт: выполнить план 73 Ф.1+Ф.2 (snapshot-baseline для Spring 1+2)

**Дата создания**: 2026-04-28
**План**: [docs/plans/73-remaster-screenshot-diff-ci.md](../plans/73-remaster-screenshot-diff-ci.md)
**Зависимости**: ✅ план 72 Ф.1 (Bootstrap) + Ф.2 Spring 1 (7 экранов)
+ Ф.3 Spring 2 (~22 экрана). Сейчас в origin-фронте есть **22-29
экранов** (Spring 1 + Spring 2). 73 Ф.1 берёт их как baseline.
Ф.3-Ф.5 (Playwright suite на новый фронт + CI-интеграция) —
отдельные сессии.
**Объём**: ~300-500 строк bash + Playwright + ~20 PNG screenshots,
1 коммит.

---

```
Задача: выполнить план 73 Ф.1+Ф.2 — снять эталонные скриншоты
22-29 экранов origin-фронта (Spring 1 + Spring 2) с running
legacy-php-стека для будущего pixel-diff в CI.

КОНТЕКСТ:

План 72 Ф.1+Ф.2+Ф.3 закрыты. На текущий момент в новом origin-фронте
реализованы примерно эти экраны:
- Spring 1 (план 72 Ф.2, коммит 47d1f0ef65): Main, Constructions,
  Research, Shipyard, Galaxy, Mission, Empire (7 шт.)
- Spring 2 ч.1 (план 72 Ф.3 коммит 48ef07cf19): 12 alliance экранов
- Spring 2 ч.2 (план 72 Ф.3 коммит 590a68b428): Resource, Market,
  Repair, Battlestats, FleetOperations (5 шт.)

Итого ~22-24 prod-экрана в новом origin-фронте. План 73 Ф.1 снимает
эталонные screenshots **из legacy-php** (не из нового origin-фронта)
для этих экранов — это «золотой» снимок, с которым потом будет
сравниваться новый.

Spring 3-5 (планы 72 Ф.4+Ф.5+Ф.6) ещё не сделаны — для них baseline
не снимаем, отложим в отдельную сессию.

ЦЕЛЬ Ф.1+Ф.2:
- Скрипт `tests/e2e/origin-baseline/take-screenshots.sh` (Playwright
  headed) который поднимает legacy-php, логинится и снимает 22-24
  экрана в `tests/e2e/origin-baseline/screenshots/`.
- ОДИН ИЗ этих 22-24 экранов снят и закоммичен (Ф.2 — проверить
  процесс, smoke).
- Регламент в README что и как обновляется (`update-baselines.sh`,
  CI триггер при намеренном изменении).

Ф.3 (Playwright suite на новый origin-фронт + diff через pixelmatch),
Ф.4 (CI-job), Ф.5 (регламент release notes) — отдельные сессии после
Spring 3+4+5 будут готовы.

ВАЖНО: legacy-php (projects/game-legacy-php/) запускается через
свой docker-compose (mysql:5.7 + php-fpm + nginx + memcached).
Это НЕ часть основного nova-стека — отдельный набор контейнеров,
прослушивает обычно :8092.

См. docs/legacy/game-legacy-access.md (или docs/legacy/game-origin-access.md
если переименование плана 78 не дочистило этот файл — проверь по
факту имени) для тестового логина / dev-аккаунта.

ПЕРЕД НАЧАЛОМ:

ПЕРВЫМ ДЕЙСТВИЕМ (до любого чтения плана):

1) git status --short. cat docs/active-sessions.md.

2) ОБЯЗАТЕЛЬНО добавь свою строку в раздел «Активные сессии»:
   | <N> | План 73 Ф.1+Ф.2 baseline screenshots | tests/e2e/origin-baseline/, projects/game-legacy-php/docker/ (только запуск, не правка), docs/plans/73-..., scripts/legacy-screenshot/ | <дата-время> | feat(e2e): Ф.1+Ф.2 baseline screenshots для origin-фронта (план 73) |

3) Параллельные сессии (могут быть): 80 (auth-cleanup, deploy/),
   72 Ф.4 (frontends/origin/). Они НЕ пересекаются с твоими файлами:
   - 80 трогает deploy/ — ты не трогаешь deploy/.
   - 72 Ф.4 трогает frontends/origin/ — ты не трогаешь.

   Если в active-sessions.md есть 73 Ф.3 или другая 73-сессия —
   пересечение неизбежно, СТОП, спроси пользователя.

ТОЛЬКО ПОСЛЕ — переходи к чтению:

4) Прочитай ПОЛНОСТЬЮ:
   - docs/plans/73-remaster-screenshot-diff-ci.md (твоё ТЗ)
   - docs/research/origin-vs-nova/origin-ui-replication.md секции
     S-001..S-024 (для понимания какие экраны и какие URL)
   - docs/legacy/game-legacy-access.md (тестовый логин)

5) Прочитай выборочно:
   - projects/game-legacy-php/docker/docker-compose.yml (как поднять)
   - projects/game-legacy-php/templates/main.tpl + alliance.tpl
     (один из них для понимания структуры — какой URL ведёт на
     какой экран, ?go=Page&action=...)
   - tests/e2e/ — есть ли уже Playwright-инфра в проекте? (если есть —
     адаптируйся; если нет — создай новую папку tests/e2e/origin-
     baseline/ с минимальной обвязкой)

ЧТО НУЖНО СДЕЛАТЬ:

### Ф.1. Скрипт снятия эталонов

- Создать `tests/e2e/origin-baseline/`:
  - `package.json` или поделить с существующей Playwright-инфрой
    (если есть в `projects/game-nova/frontends/nova/e2e/` — там
    уже Playwright настроен, можно расширить или сделать standalone
    setup; выбери проще).
  - `take-screenshots.sh` — bash-скрипт:
    1. Поднимает legacy-php стек:
       `cd projects/game-legacy-php/docker && docker compose up -d`
    2. Ждёт healthy (curl -f http://localhost:8092/?go=Login или
       что есть в legacy для healthcheck; max 30s).
    3. Запускает Playwright spec:
       `npx playwright test --headed --project=chromium baseline.spec.ts`
    4. На finish — `docker compose stop` (или down если нужно
       чисто, но stop быстрее для повторного запуска).
  - `baseline.spec.ts` — Playwright-сценарий:
    - beforeAll: login через ?go=Login form (юзер `test` /
      пароль из game-legacy-access.md).
    - Для каждого S-NNN экрана: page.goto, page.waitForLoadState,
      page.screenshot({path: `screenshots/<S-NNN>.png`,
      fullPage: true}).
    - Список экранов в `screens.ts`:
      ```
      export const SCREENS = [
        {id: 'S-001', name: 'main', url: '/?go=Main'},
        {id: 'S-002', name: 'constructions', url: '/?go=Constructions'},
        ... // 22-24 экрана для Spring 1+2
      ];
      ```
      URL'ы — реальные легасные ?go=Page&action=... паттерны.
      Если экран = одна страница с табами — снимай каждое состояние
      отдельно (S-008-overview, S-008-members и т.п.).

  - `update-baselines.sh` — то же что take-screenshots, но с явным
    выводом «вы переписываете эталоны, продолжить? (y/N)».

  - `README.md` — регламент: когда снимать, как обновлять, что
    коммитить, как читать diff-отчёт (Ф.3 в будущем).

### Ф.2. Snapshot первого набора (5-7 экранов smoke)

- Запусти take-screenshots.sh локально (или попроси пользователя
  если у него Docker под рукой).
- Снимок 5-7 экранов из 22-24 (для smoke процесса):
  S-001 Main, S-002 Constructions, S-003 Research, S-008 Alliance
  Overview, S-020 Resource, S-021 Market, S-022 Repair.
- Закоммить эти 5-7 PNG в `tests/e2e/origin-baseline/screenshots/`.
  Размер каждого PNG ~50-200 KB → суммарно ~500KB-1.5MB.
- В коммит-сообщении: «Ф.2 baseline 5-7 экранов snapshot, остальные
  17 — в Ф.2.5 после ручного запуска пользователем».

ЕСЛИ DOCKER НЕДОСТУПЕН в твоей среде:
- Сценарий take-screenshots.sh + baseline.spec.ts всё равно создай.
- В commit-message: «Ф.2 snapshot не выполнен (Docker недоступен в
  агент-среде). Пользователь запускает локально:
  bash tests/e2e/origin-baseline/take-screenshots.sh».
- README пометить «снимки baseline отсутствуют — выполнить локально
  при первом запуске».

### Smoke

- bash tests/e2e/origin-baseline/take-screenshots.sh — должен
  отработать без ошибок (если Docker есть).
- ls tests/e2e/origin-baseline/screenshots/ — должны быть PNG.
- Открыть один PNG в просмотрщике, убедиться что это реальный
  скриншот legacy-страницы (не пустой, не error-page).

### Финализация

- Шапка плана 73: Ф.1 ✅, Ф.2 🟡 (smoke сделан / отложен под
  пользователя; Ф.2.5 «снять остальные 17 экранов» — отдельная
  быстрая сессия после).
- Запись итерации в docs/project-creation.txt.

══════════════════════════════════════════════════════════════════
ОБЯЗАТЕЛЬНЫЕ ПРАВИЛА R0-R15 — см. блок rules-r0-r15.md
СКОПИРУЙ СЮДА полностью раздел из docs/prompts/_blocks/rules-r0-r15.md

Особо важно:
- R0: nova-баланс/код не меняем — только tests/.
- R5: pixel-perfect actually — это и есть screenshot-diff подход.
══════════════════════════════════════════════════════════════════

══════════════════════════════════════════════════════════════════
GIT-ИЗОЛЯЦИЯ ОТ ПАРАЛЛЕЛЬНЫХ СЕССИЙ — см. блок git-isolation.md
СКОПИРУЙ СЮДА полностью раздел из docs/prompts/_blocks/git-isolation.md

Свои пути для CC_AGENT_PATHS:
- tests/e2e/origin-baseline/

  (вся папка целиком — твой territory; параллельные планы 80 и
   72 Ф.4 в неё не лезут)

- docs/plans/73-remaster-screenshot-diff-ci.md
- docs/active-sessions.md
- docs/project-creation.txt (запись итерации)

ВНИМАНИЕ: НЕ трогай:
- projects/game-legacy-php/ — только запуск через docker compose,
  файлы НЕ правишь.
- projects/game-nova/frontends/origin/ (план 72 Ф.4 параллельно).
- deploy/ (план 80 параллельно).
- .github/workflows/ — CI-job это Ф.4, не сейчас.
══════════════════════════════════════════════════════════════════

КОММИТЫ:

Один коммит:

feat(e2e): Ф.1+Ф.2 baseline screenshots для origin-фронта (план 73)

Trailer: Generated-with: Claude Code
ВСЕГДА с двойным тире:
git commit -m "..." -- $CC_AGENT_PATHS

ЧЕГО НЕ ДЕЛАТЬ:

- НЕ делай Ф.3 (Playwright на новый origin-фронт) — отдельная сессия
  после Spring 3+4+5.
- НЕ делай Ф.4 (CI-интеграция) — отдельная сессия (требует чтобы
  Ф.3 был готов).
- НЕ снимай baseline для Spring 3+4+5 экранов которых ещё нет.
- НЕ редактируй legacy-PHP код.
- НЕ забывай про -- в git commit.

УСПЕШНЫЙ ИСХОД:

- tests/e2e/origin-baseline/ создан с README + take-screenshots.sh +
  baseline.spec.ts + screens.ts.
- 5-7 baseline PNG закоммичены (если Docker доступен), либо
  baseline создаётся при первом локальном запуске пользователем.
- Шапка плана 73: Ф.1 ✅, Ф.2 🟡 (snapshot сделан или отложен).
- Запись в docs/project-creation.txt.
- Удалена строка из docs/active-sessions.md.

Стартуй.
```
