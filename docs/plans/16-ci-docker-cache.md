# План 16: Ускорение CI через docker buildx cache

**Цель:** сократить время `e2e` job в GitHub Actions с ~10-12 минут до
~2-4 минут за счёт кеширования Docker-слоёв между запусками.

**Зачем сейчас:** пока PR'ов мало — терпимо. Как только поток PR
вырастет (≥5/день), каждая минута × число запусков = часы ожидания и
потраченных runner-минут. GitHub Actions даёт 2000 бесплатных минут/мес
на private-repo — при 10 PR × 12 мин только e2e — 120 мин, треть лимита.

**Когда делать:** триггер — начали заливать 3+ PR в день ИЛИ CI стал
блокировать рабочий процесс (ждать 15 минут зелёного перед merge —
уже боль).

Сейчас `simplifications.md` → раздел «E2E в Docker без кеша docker
buildx» фиксирует это как известное упрощение, приоритет M.

---

## Что происходит сегодня

Job `e2e` в `.github/workflows/ci.yml`:

```yaml
- name: Build & run E2E stack (detached)
  run: docker compose -f deploy/docker-compose.e2e.yml up -d --build
```

Каждый прогон:
1. Pull базовых образов (golang, node, playwright, alpine, postgres)
2. `go mod download` для backend/worker/testseed (одно и то же 3 раза)
3. `go build` для 3 бинарей
4. `npm install` в frontend (~500 пакетов)
5. `npm install + npx playwright install chromium` в playwright-контейнере

**Ни один слой не кешируется между runs** — Docker-демон в runner'е
одноразовый. Итого ~6-8 минут чистого билда до старта самих тестов.

---

## Что делаем

### Ф.1 Включить buildx + GitHub Actions cache (HIGH)

**Основной фикс:** Docker Buildx поддерживает `--cache-from type=gha` /
`--cache-to type=gha`, который использует GitHub-native artifact cache
(до 10 ГБ на repo, LRU). На повторных запусках неизменённые слои
берутся из кеша мгновенно.

`.github/workflows/ci.yml` — job `e2e`:

```yaml
e2e:
  needs: [backend, frontend]
  runs-on: ubuntu-latest
  timeout-minutes: 20
  steps:
    - uses: actions/checkout@v4

    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v3

    - name: Build images with GHA cache
      uses: docker/bake-action@v5
      with:
        files: |
          ./deploy/docker-compose.e2e.yml
        set: |
          *.cache-from=type=gha
          *.cache-to=type=gha,mode=max
        load: true

    - name: Up E2E stack
      run: docker compose -f deploy/docker-compose.e2e.yml up -d

    - name: Wait for Playwright
      run: |
        docker compose -f deploy/docker-compose.e2e.yml logs -f --no-log-prefix playwright &
        LOGS=$!
        STATUS=$(docker wait deploy-playwright-1 || echo 1)
        kill $LOGS 2>/dev/null || true
        exit $STATUS
    # ... остальное без изменений (artifact upload, teardown)
```

**Ключевые моменты:**
- `docker/bake-action` билдит сразу все сервисы из compose с общим
  cache namespace'ом.
- `mode=max` — кешируем все промежуточные слои, не только финальный.
  Даст наибольшую экономию, но cache разрастается быстрее (до 10 ГБ
  лимит).
- `load: true` — после билда образы доступны в локальном daemon'е
  под теми же именами, compose их подхватит как «уже собранные».

**Проверка:**
- [ ] Первый прогон после merge плана — без ускорения (cache пустой),
      ~10-12 минут
- [ ] Второй прогон (commit не меняющий Dockerfile/deps) — должен
      быть ~2-4 минуты
- [ ] В лог Actions печатает строки `CACHED` рядом с слоями

---

### Ф.2 Оптимизировать Dockerfile'ы под кеш (HIGH)

Даже с buildx cache, если слой `COPY backend ./backend` нарушается при
любом изменении одного .go файла, все последующие слои (go build)
пересобираются.

#### Ф.2.1 Backend Dockerfile

Сейчас (см. [backend/Dockerfile](../../backend/Dockerfile)):

```dockerfile
COPY backend/go.mod backend/go.sum ./backend/
RUN cd backend && go mod download   # OK — deps-слой

COPY backend ./backend               # ломается при любой правке .go
RUN cd backend && go build ...       # пересобирается всегда
```

**Проблема:** `go mod download` кешируется нормально, но `go build`
— нет. Каждый PR будет собирать 3 бинаря с нуля (~1-2 мин).

**Фикс:** оставить как есть — кеш GHA всё равно поможет на ветках
без изменений backend-кода (только фронт или docs).

Дополнительно проверить, что **`.dockerignore`** включает `*.exe`,
`backend/server.exe`, `backend/worker.exe`, `backend/testseed.exe` —
иначе билд-контекст раздувается (у нас в status.exe файлы видны).

#### Ф.2.2 Frontend Dockerfile

```dockerfile
COPY frontend/package.json ./
COPY frontend/package-lock.json* ./
RUN npm install                      # OK — deps-слой

COPY frontend ./                     # ломается на каждом изменении tsx
```

**Уже оптимально.** `npm install` кешируется при неизменном lock.

#### Ф.2.3 Playwright Dockerfile

```dockerfile
COPY frontend/package.json frontend/package-lock.json* ./frontend/
RUN cd frontend && npm install && npx playwright install chromium  # ~5 мин
COPY frontend ./frontend
COPY api ./api
```

**Оптимально.** Самый дорогой шаг (install chromium ~5 минут) — в
слое до copy исходников, кешируется при неизменных deps.

**Проверка:**
- [ ] `.dockerignore` исключает `*.exe`, `node_modules`, `test-results/`
- [ ] Второй прогон с правкой `*.tsx` не перетягивает `npm install`

---

### Ф.3 Разбить e2e job на этапы для видимости (MEDIUM)

Сейчас всё в одном большом step'е. Если билд упал — не видно на каком
именно сервисе. Разделим:

```yaml
- name: Build backend image
  run: docker buildx bake -f deploy/docker-compose.e2e.yml backend worker testseed --set '*.cache-from=type=gha' --set '*.cache-to=type=gha,mode=max' --load

- name: Build frontend image
  run: docker buildx bake -f deploy/docker-compose.e2e.yml frontend --set '*.cache-from=type=gha' --set '*.cache-to=type=gha,mode=max' --load

- name: Build playwright image
  run: docker buildx bake -f deploy/docker-compose.e2e.yml playwright --set '*.cache-from=type=gha' --set '*.cache-to=type=gha,mode=max' --load
```

**Польза:** в UI GitHub Actions сразу видно, какой этап сколько занял
и что упало. Плюс можно распараллелить через `matrix` (но только если
будет явно больно).

**Проверка:**
- [ ] В Actions-логе три отдельных шага build-* с таймингами

---

### Ф.4 Прогрев кеша на main (MEDIUM)

**Проблема:** PR использует кеш от последнего main, но если main меняет
dep'ы, следующий PR платит цену rebuild.

**Фикс:** отдельный job `cache-warm` на push в main:

```yaml
cache-warm:
  if: github.ref == 'refs/heads/main'
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - uses: docker/setup-buildx-action@v3
    - name: Build & push all images to GHA cache
      uses: docker/bake-action@v5
      with:
        files: ./deploy/docker-compose.e2e.yml
        set: |
          *.cache-from=type=gha
          *.cache-to=type=gha,mode=max
        # load: false — не нужен локально, только в cache
```

PR после merge'а в main уже подтянет свежий кеш.

**Проверка:**
- [ ] После merge в main, cache-warm отработал ≤ 12 минут
- [ ] Следующий PR-e2e использует новый кеш

---

### Ф.5 Локально — volume-кеш для go-mod и npm (LOW)

**Не связано с CI, но тоже экономит время разработчика.** Сейчас при
`make test-e2e-docker` каждый rebuild backend перекачивает go-deps.

Добавить в `deploy/docker-compose.e2e.yml`:

```yaml
backend:
  build:
    context: ..
    dockerfile: backend/Dockerfile
    cache_from:
      - type=local,src=/tmp/oxsar-cache/backend
    cache_to:
      - type=local,dest=/tmp/oxsar-cache/backend,mode=max
```

И в Makefile target `test-e2e-docker` перед `up --build`:

```make
.PHONY: test-e2e-docker
test-e2e-docker:
	mkdir -p /tmp/oxsar-cache
	DOCKER_BUILDKIT=1 docker compose -f deploy/docker-compose.e2e.yml up -d --build
```

**Польза:** повторный прогон локально ускоряется с ~4 до ~1 минуты.

**Приоритет:** L — пока каждый локальный прогон терпим.

---

## Метрики успеха

До (baseline, текущий CI без оптимизаций):
- E2E job cold: **~10-12 мин** (build ~6-8 мин + тесты ~3-4 мин)
- E2E job без изменений deps: **~10-12 мин** (кеша нет, всё пересобирается)

После Ф.1+Ф.2:
- E2E job cold (первый прогон плана): ~10-12 мин (как было)
- E2E job с hit-кешем (PR без изменений deps): **~3-4 мин** (6-8× быстрее билд)
- E2E job с частичным hit (PR меняет только код, не deps): **~4-5 мин**

После Ф.4 (cache-warm на main):
- Среднее время PR-e2e: стабильные **~3-4 мин** даже после merge в main

---

## Порядок реализации

1. **Ф.2.1** (`.dockerignore` проверка) — 5 минут, можно сделать первым
2. **Ф.1** (buildx + GHA cache) — основная работа, 30-60 минут
3. **Первый прогон** — проверить что всё собирается, замерить baseline
4. **Второй прогон (trivial commit)** — замерить hit-rate
5. **Ф.3** (разбить на шаги) — косметика, 15 минут
6. **Ф.4** (cache-warm на main) — 15 минут, после стабилизации Ф.1
7. **Ф.5** (локальный volume-cache) — по потребности, опционально

---

## Риски

- **Cache corruption.** Иногда GHA cache ломается (partial upload).
  Фикс: ручная очистка через GitHub API или через pattern `cache-bust`
  в env переменной job'а. Документировать в runbook.
- **Cache расходует лимит 10 ГБ.** `mode=max` агрессивный. Мониторить
  через `gh actions cache list` и чистить старые.
- **buildx требует Docker 23+.** GitHub runners уже имеют, локально —
  надо проверить. Если старее — `docker buildx install` или обновление.
- **PR из forks не имеет доступа к cache.** GitHub Actions security
  model: fork'и пишут в свой cache namespace. Внешние PR будут
  медленнее. Для solo-проекта не важно, для open-source — ок.

---

## Что НЕ делаем

- **Перенос билдов в dedicated CI (e.g. BuildKite, CircleCI).** Это
  стрельба из пушки по воробьям. GHA cache + buildx закроет 90% боли.
- **Arm64 multi-arch билды.** Никому не нужно пока.
- **Registry-кеш (push/pull от GHCR).** Сложнее GHA cache, выгоды без
  self-hosted runner'ов нет.
- **Предсобранные base-images в GHCR.** Если бы билд base-образа
  занимал 15+ минут — было бы оправдано. Сейчас нет.

---

## Связанное

- [simplifications.md](../simplifications.md) — раздел «E2E в Docker
  без кеша docker buildx» — эта запись закроется после Ф.1
- [.github/workflows/ci.yml](../../.github/workflows/ci.yml) — текущий
  CI, job `e2e` будет переписан
- [docs/ops/runbooks/docker-stacks.md](../ops/runbooks/docker-stacks.md) —
  шпаргалка по стекам, если появятся команды для работы с кешем —
  добавить туда секцию
- [docs/plans/13-ui-testing.md](13-ui-testing.md) — план E2E-тестов,
  этот план ускоряет их CI-часть
