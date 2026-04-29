# Docker-стеки: когда какой запускать

У нас три compose-файла, каждый под свой сценарий. Шпаргалка, чтобы не
путаться.

## TL;DR

| Задача | Команда |
|--------|---------|
| Разработка, HMR, ручная проверка | `make dev-up` |
| Прогон E2E локально | `make test-e2e-docker` |
| E2E в CI | автоматически, job `e2e` в `.github/workflows/ci.yml` |
| Посмотреть UI с mock-оплатой и свежим seed | `make ui-preview` |
| Всё потушить | `make dev-down` / `make test-e2e-docker-down` / `make ui-preview-down` |

---

## `deploy/docker-compose.yml` — dev-стек

**Когда использовать:** обычная разработка, отладка, ручной прогон фич.

**Что даёт:**
- Постоянный postgres (volume `pg-data`) — БД переживает рестарты
- Frontend с bind-mount `../frontend:/app` → HMR работает, правишь tsx
  на хосте → Vite сразу перезагружает браузер
- Порты: UI **5173**, API **8081**, pg 5433, redis 6380
- Backend в обычном режиме (без mock-платежей)

**Запуск:**

```bash
make dev-up
# или
docker compose -f deploy/docker-compose.yml up --build
```

**Минусы:**
- Медленная пересборка backend (нет volume) — меняешь .go → `docker compose restart backend`
- Платежи через реальный шлюз (или отключены, если нет PAYMENT_PROVIDER)
- БД не сбрасывается автоматически — чтобы «с нуля», делать `make dev-down` с `-v`:
  `docker compose -f deploy/docker-compose.yml down -v`

---

## `deploy/docker-compose.e2e.yml` — E2E / воспроизводимый прогон

**Когда использовать:**
- Прогон Playwright локально перед PR (`make test-e2e-docker`)
- Нужна чистая БД (tmpfs — ничего не персистит между запусками)
- Проверка «как в CI»

**Что даёт:**
- `PAYMENT_PROVIDER=mock` + `PAYMENT_MOCK_BASE_URL=http://uni01-backend:8080` →
  платежи без денег, редирект внутри docker-сети
- Postgres на **tmpfs** (быстрее холодного старта, полный сброс между запусками)
- One-shot `testseed --reset` → 5 детерминированных игроков
  (admin/alice/bob/eve/charlie, пароль `DevPass123`)
- Healthcheck'и на всех сервисах, retry-логика для backend
- Playwright-контейнер по умолчанию сразу запускает тесты

**Порты наружу НЕ проброшены** — всё общается внутри docker-сети по
service-именам (`http://uni01-backend:8080`, `http://uni01-frontend:5173`).

**Запуск полного E2E-прогона:**

```bash
make test-e2e-docker
```

Это эквивалент: `docker compose up --build` (detached) + `docker wait
playwright` + забрать его exit-code + `docker compose down`.

**Остановить:**

```bash
make test-e2e-docker-down
```

---

## `deploy/docker-compose.e2e.ports.yml` — override для ручного осмотра

**Когда использовать:** хочешь своими глазами посмотреть UI на
детерминированных seed-данных и/или проверить mock-флоу платежей.

**Что даёт:** поверх e2e-стека пробрасывает порты:
- frontend: `5173:5173`
- backend: `8081:8080`

**Запуск:**

```bash
make ui-preview
```

(Под капотом — `docker compose -f …e2e.yml -f …e2e.ports.yml up -d
postgres redis migrate backend worker testseed frontend`. Playwright
не стартует — он бы сразу прогнал тесты и вышел.)

После запуска открой <http://localhost:5173>, логин любого тестового
игрока (пароль `DevPass123`):
- **bob** — superadmin с прокачанной планетой и флотом
- **alice** — новичок, пустые состояния
- **admin** — superadmin со средней планетой
- **charlie** — лидер альянса `[UT]`
- **eve** — жертва рядом с bob (для тестов атаки)

**Остановить:**

```bash
make ui-preview-down
```

---

## Что выбрать в типичных ситуациях

**«Я пишу новый frontend-компонент»** → `make dev-up`. HMR быстрее, БД
переживает перезапуски.

**«Я закончил фичу, хочу убедиться что ничего не упало»** → `make
test-e2e-docker`. Прогонит 110 E2E-тестов в чистом окружении.

**«Мне нужен UI с конкретными данными (прокачанный игрок, альянс, лот)»**
→ e2e-стек + ports override. testseed даёт стабильный state
(см. [backend/cmd/tools/testseed/main.go](../../backend/cmd/tools/testseed/main.go)).

**«Хочу проверить оплату кредитов»** → только e2e-стек (там mock
gateway). В dev-стеке платежи требуют реальной Робокассы либо не работают.

**«БД сломалась, хочу пересидеть»:**
- Dev-стек: `make dev-down && make dev-up && make test-seed`
- E2E-стек: `make test-e2e-docker-down && make test-e2e-docker` (tmpfs
  сбрасывает сам)

---

## Общие грабли

- **`docker compose` путается между проектами.** Оба compose-файла
  используют одинаковые имена контейнеров (`deploy-postgres-1` и т.д.) —
  одновременно два стека запустить не получится. Туши один перед
  стартом другого.
- **Порт 5173 / 5433 / 6380 занят.** Скорее всего забыл потушить
  прошлый запуск. `docker ps | grep deploy-` → `docker compose … down`.
- **Изменения в `vite.config.ts` / `docker-compose.*.yml`.** Нужен
  полный `up --build` или `docker compose build frontend`. Bind-mount
  перекрывает только `/app`, но не саму entrypoint-конфигурацию.
- **Vite в e2e-стеке без bind-mount.** Это сознательно: CI-прогон
  должен использовать «запечённый» образ. Если хочешь HMR в e2e — см.
  [deploy/docker-compose.yml](../../deploy/docker-compose.yml).

---

## Связанное

- [deploy/docker-compose.yml](../../deploy/docker-compose.yml) — dev-стек
- [deploy/docker-compose.e2e.yml](../../deploy/docker-compose.e2e.yml) — E2E
- [deploy/docker-compose.e2e.ports.yml](../../deploy/docker-compose.e2e.ports.yml) — override
- [Makefile](../../Makefile) — `dev-up`, `test-e2e-docker`, `test-seed`
- [docs/plans/13-ui-testing.md](../plans/13-ui-testing.md) — план E2E-тестирования
- [frontend/e2e/README.md](../../frontend/e2e/README.md) — как писать новые спеки
