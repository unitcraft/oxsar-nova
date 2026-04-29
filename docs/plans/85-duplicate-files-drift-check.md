# План 85: CI-проверка drift'а DUPLICATE-файлов между Go-модулями

**Дата**: 2026-04-29
**Статус**: ✅ ЗАКРЫТО (2026-04-29)
**Зависимости**: нет (точечная инфра-задача).

> **Итог**: реализован `cmd/tools/check-duplicates/` — Go-утилита,
> которая walk'ает `projects/*/backend/`, ищет `.go`-файлы с маркером
> `// DUPLICATE: этот файл скопирован`, парсит шапку (список
> путей-копий + блок до пустой строки после `// Причина дубля:`),
> хеширует нормализованное тело (с заменой per-module
> `oxsar/<module>/` → placeholder, чтобы import-prefix не считался
> drift'ом) и сверяет хеши внутри группы. При расхождении печатает
> unified-diff (LCS-based, через stdlib) и exit 1. Флаг `-fix`
> тиражирует тело эталона (первый путь в шапке) с заменой
> import-prefix обратно на per-module. Make-таргеты:
> `make check-duplicates` (включён в `make lint`) и
> `make sync-duplicates`. CI-step добавлен в backend-job
> `.github/workflows/ci.yml`.
>
> При первом запуске инструмент нашёл 4 реальных drift'а:
> - `pkg/jwtrs/jwtrs.go` (identity получил доп. метод `AccessTTL()`,
>   которого не было в game-nova/portal/billing) — починено
>   тиражированием метода через `-fix` после ручной правки эталона.
> - `pkg/metrics/metrics.go` — game-nova монолитно расширен
>   game-specific Register*-вызовами (Balance/Alliance/Notepad/
>   Billing/AlienBuyout/Exchange), не совместимыми с identity/
>   portal/billing. Решение: снять DUPLICATE-маркер с game-nova-копии,
>   оставив 3-местную drift-группу (identity/portal/billing).
>   Записано в `docs/simplifications.md` P85.2.
>
> Smoke (Ф.4) выполнен: clean → exit 0; синтетический drift в
> `pkg/ids/ids.go` → exit 1 с читаемым diff'ом; revert → exit 0.
> Tests на сам инструмент — нет (P85.1, осознанно).
**Связанные документы**:
- [docs/plans/38-billing-service.md](38-billing-service.md) §«Структура» — фиксирует 9 DUPLICATE-точек billing-service.
- Файлы с маркером `DUPLICATE` (grep по `DUPLICATE: этот файл скопирован`) — ~50+ копий между `game-nova`, `identity`, `portal`, `billing`.

---

## Контекст

В монорепо oxsar-nova каждый бекенд-сервис — отдельный Go-модуль
(`projects/{game-nova,identity,portal,billing}/backend/go.mod`).
Инфраструктурные пакеты (HTTP middleware, JWT/JWKS, pgx pool,
metrics/trace setup, UUID, transaction helper) **намеренно
скопированы** в каждый модуль с маркером `DUPLICATE` и явным
списком путей-копий в шапке файла.

Это сознательное решение (см. план 38 §«Структура»): альтернативы
(shared-модуль через `replace`, мономодуль, `go.work`) ломают
Docker build-context, или связывают релиз-циклы сервисов, или
требуют переделки инфраструктуры. Дубль из ~9 файлов × 4 модуля
дешевле, пока его поддерживают синхронным.

**Проблема**: синхронность поддерживается **руками**. При правке
одной копии забыть остальные — лёгкая ошибка. Drift между
копиями уже сейчас не виден без ручного diff.

---

## Цель

Добавить автоматическую проверку, что все группы DUPLICATE-файлов
содержат идентичный код (минус первые N строк с маркером и
per-module путями). Гонять в CI на каждом PR. При drift'е — фейл
с понятным diff'ом.

Дополнительно — `make sync-duplicates` для одностороннего
тиражирования эталона по копиям (использовать аккуратно, не на
каждый чих).

---

## Не цель

- **Не** переходим на shared-модуль (вариант обсуждён, отвергнут —
  ROI отрицательный).
- **Не** меняем структуру модулей, build-pipeline, Dockerfiles.
- **Не** трогаем `internal/limits/handler.go` где маркер
  `DUPLICATE-pattern` — это копия идиомы, а не файла; проверка её
  не касается.

---

## Что делаем

### 1. Скрипт `scripts/check-duplicates.sh`

Логика:
1. Найти все Go-файлы с маркером `// DUPLICATE: этот файл скопирован`.
2. Извлечь из шапки каждого файла **список путей-копий** (строки
   после маркера, начинающиеся с `//   - `).
3. Сгруппировать файлы по этому списку (все копии одной группы
   должны указывать на одинаковый набор путей — это первая
   проверка консистентности).
4. Для каждой группы:
   - Прочитать содержимое каждой копии.
   - Срезать «шапку» — всё до первой непустой строки **после**
     закрывающей строки маркера (последняя строка маркера —
     `// Причина дубля:` или просто пустая строка после списка
     путей; договоримся: шапка заканчивается первой пустой
     строкой после строки `// Причина дубля:`).
   - Посчитать SHA256 от оставшегося содержимого.
   - Все хеши в группе должны совпасть.
5. При расхождении — вывести `diff -u` между эталонной копией
   (по convention — копия из модуля, который указан **первым**
   в списке путей) и каждой расходящейся, exit 1.

Скрипт пишется как Go-программа в `cmd/tools/check-duplicates/`
(а не bash) — единая платформа, проще тестировать, кросс-платформа
для Windows-разработчика. Положить в **game-nova/backend/cmd/tools/**
(там уже есть `battle-sim`, `import-datasheets`, `import-legacy-user`).

### 2. Make-таргет `make check-duplicates`

В корневом Makefile:

```makefile
check-duplicates:
	go run ./projects/game-nova/backend/cmd/tools/check-duplicates
.PHONY: check-duplicates
```

Включить в существующий `make lint`:

```makefile
lint: ... check-duplicates
```

### 3. CI-job

В `.github/workflows/` (или эквивалент) — добавить шаг
`make check-duplicates` в существующий lint-job. Время работы
скрипта — секунды, отдельный workflow не нужен.

### 4. Make-таргет `make sync-duplicates`

Опциональный, **не обязательный**. Berёт первую копию из каждой
группы как эталон, перезаписывает остальные с сохранением их
шапки (только маркер + список путей; `Причина дубля` тоже остаётся
из эталона). Использовать вручную после сознательной правки
эталона. Не запускать в CI.

Реализация в той же Go-программе с флагом `-fix`.

---

## Список ожидаемых групп DUPLICATE

По состоянию на `00e5431aea`:

| Группа | Копии | Где |
|---|---|---|
| `pkg/ids/ids.go` | 4 | game-nova, identity, portal, billing |
| `pkg/ids/ids_test.go` | 4 | те же |
| `pkg/jwtrs/jwtrs.go` | 4 | те же |
| `pkg/jwtrs/jwtrs_test.go` | 4 | те же |
| `pkg/metrics/metrics.go` | 4 | те же |
| `pkg/trace/trace.go` | 4 | те же |
| `internal/httpx/router.go` | 4 | те же |
| `internal/httpx/response.go` | 4 | те же |
| `internal/httpx/response_test.go` | 4 | те же |
| `internal/httpx/logger.go` | 4 | те же |
| `internal/httpx/recover.go` | 4 | те же |
| `internal/httpx/trace.go` | 4 | те же |
| `internal/storage/postgres.go` | 4 | те же |
| `internal/repo/tx.go` | 3+ | game-nova, identity, portal, billing |
| `internal/auth/password.go` | 2 | game-nova, identity |
| `internal/auth/password_test.go` | 2 | те же |
| `internal/auth/jwksloader.go` | 3 | game-nova, portal, billing (identity — origin) |
| `internal/universe/registry.go` | 2 | game-nova, portal |

Точный список генерируется самим скриптом из grep'а — таблица
выше для документации, не для hardcode.

---

## Acceptance criteria

- ✅ `go run ./.../check-duplicates` на чистом репо — exit 0,
  выводит «N групп, M файлов, всё синхронно».
- ✅ Если правишь одну копию (например, добавляешь пробел в
  `game-nova/.../ids.go`) и не правишь остальные — `make
  check-duplicates` падает с unified-diff.
- ✅ `make sync-duplicates` после правки эталона — все копии
  становятся идентичны, `make check-duplicates` снова проходит.
- ✅ CI прогон: новый PR с правкой одной копии → red CI.

---

## Edge-cases

1. **Per-module отличия в маркере**: каждая копия содержит свой
   собственный путь первой строкой (`// projects/game-nova/...`).
   Решение: при сравнении срезаем **всю шапку до первой пустой
   строки после `// Причина дубля:`**.
2. **Файлы без маркера, но идентичные**: скрипт находит группы
   **по маркеру**, не по имени. Файл без маркера — не проверяется,
   это feature, не bug.
3. **Test-файлы с разными `package X_test`**: package name —
   часть содержимого, проверяется как обычный код. У нас все
   копии используют один и тот же package name, проблемы не должно
   быть — но если возникнет, добавим в шапку строку
   `// package: foo` и срежем при сравнении.
4. **Imports с разным module-path**: НЕ должно быть. Все
   DUPLICATE-файлы — это либо stdlib-only (`pkg/ids/ids.go`
   импортит `github.com/google/uuid`), либо self-contained.
   Если в DUPLICATE-файле появляется `import "oxsar/billing/..."` —
   это бажный дубль, его и должна поймать проверка.

---

## Trade-offs / simplifications

- **Не** автогенерируем GitHub Action — добавляем в существующий
  lint-flow проекта. Если flow'а нет — план 85 не блокирует план,
  скрипт ходит локально через `make`.
- **Эталон выбирается convention'ом** (первый путь в маркере), а
  не флагом `// AUTHORITATIVE`. Если нужно поменять — сначала
  правится список путей во **всех** копиях (так что эталон
  всегда первый), потом запускается `sync-duplicates`. Чуть
  громоздко, но избегает ещё одного типа маркера.
- **Не пишем тесты на `check-duplicates`** до первой реальной
  ловли drift'а. Скрипт простой, тесты на него — overengineering.
  Если поломается — починим по факту.

---

## План работ

1. **Ф.1** — Go-программа в `cmd/tools/check-duplicates/`:
   - Walk по `projects/*/backend/`, поиск файлов с маркером.
   - Парсинг шапки, группировка, сравнение.
   - Diff-output при расхождении.
   - Флаг `-fix` для тиражирования.
2. **Ф.2** — Makefile-таргет, включение в `make lint`.
3. **Ф.3** — CI-шаг в существующем lint workflow.
4. **Ф.4** — Smoke: руками внести пробел в одну копию `pkg/ids/`,
   запустить `make check-duplicates`, убедиться что фейлит с
   читаемым diff'ом. Откатить.
5. **Ф.5** — Записать в `docs/simplifications.md` если что-то
   упростили (например, отказ от тестов на сам скрипт).
6. **Ф.6** — Дневник в `docs/project-creation.txt`, commit.

---

## Оценка

Полдня работы. ~150-200 строк Go в `check-duplicates/main.go`,
плюс 2-3 строки в Makefile, плюс 3-4 строки в CI-конфиге.

---

## Ссылки

- [docs/plans/38-billing-service.md](38-billing-service.md) — где впервые
  зафиксирован DUPLICATE-паттерн.
- Маркер в коде:
  [projects/game-nova/backend/pkg/ids/ids.go:5-12](../../projects/game-nova/backend/pkg/ids/ids.go#L5-L12).
