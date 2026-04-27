# План 40: Аудит лицензий зависимостей (GPL/AGPL)

**Дата**: 2026-04-26
**Статус**: Активный
**Зависимости**: нет.
**Связанные документы**: [LICENSE](../../LICENSE) (PolyForm Noncommercial 1.0.0),
[COMMERCIAL-LICENSE.md](../../COMMERCIAL-LICENSE.md), `.github/workflows/ci.yml`.

---

## Цель

Гарантировать совместимость лицензий зависимостей с PolyForm Noncommercial.
Лицензии семейства GPL и AGPL **несовместимы** с PolyForm Noncommercial:
GPL требует распространения производных под GPL, AGPL дополнительно
считает "использование на сервере = распространение". Любая такая
зависимость превращает весь oxsar-nova в продукт под GPL/AGPL, что
отменяет защитную функцию PolyForm.

Цель плана:
1. Однократно проверить текущий граф зависимостей (4 Go-модуля + 2
   frontend-проекта на npm) на наличие GPL/AGPL.
2. Закрепить проверку в CI, чтобы никто не добавил несовместимую
   зависимость в будущем.
3. Документировать результат и список разрешённых семейств лицензий.

Совместимые с PolyForm семейства: MIT, Apache-2.0, BSD-2/3-Clause, ISC,
MPL-2.0, Unlicense, CC0, Zlib. Все остальные — рассматривать индивидуально
при первом появлении.

---

## Что меняем

### 1. Однократный аудит

Прогон сканеров на текущем состоянии:

- Go-модули (4 шт): `go-licenses report ./...` или эквивалент в каждом из
  `projects/{game-nova,portal,auth,billing}/backend/`.
- Frontend (2 шт): `npx license-checker --production --json` в каждом из
  `projects/{game-nova,portal}/frontend/`.

Результат — сводная таблица в `docs/ops/license-audit.md`: пакет, версия,
лицензия, вердикт (OK / нужно проверить / **GPL/AGPL — заменить**).

### 2. CI-проверка

В `.github/workflows/ci.yml` — новый job `license-check`:

- Для каждого Go-модуля: `go-licenses check ./... --disallowed_types=forbidden,restricted`
  (forbidden = GPL/AGPL, restricted = LGPL и подобные copyleft с условиями).
- Для frontend: `license-checker --production --onlyAllow "MIT;Apache-2.0;BSD-2-Clause;BSD-3-Clause;ISC;MPL-2.0;Unlicense;CC0-1.0;Zlib"`
  (whitelist по семействам).

При появлении несовместимой лицензии CI падает, PR не сливается. Job
встраивается параллельно с `lint` и `test`, не блокирует их.

### 3. Документация

`docs/ops/license-audit.md` (новый):
- список разрешённых семейств лицензий;
- ссылка на job в CI;
- инструкция: что делать, если CI обнаружил GPL/AGPL (заменить пакет,
  либо найти dual-licensed альтернативу).

В `CONTRIBUTING.md` — короткий абзац-напоминание: при добавлении
зависимостей следить за лицензией; CI проверит автоматически.

---

## Этапы

### Ф.1. Однократный аудит — Go

- Установить `go-licenses` (`go install github.com/google/go-licenses@latest`).
- Прогнать `go-licenses report ./...` в каждом из 4 backend-модулей;
  собрать csv.
- Прогнать `go-licenses check ./... --disallowed_types=forbidden,restricted`
  и убедиться, что exit 0. Если нет — точечно разбирать каждый случай.

### Ф.2. Однократный аудит — Node

- В каждом из 2 frontend-проектов: `npx license-checker --production --summary`
  для обзора, `npx license-checker --production --json > licenses.json`
  для машиночитаемого вывода.
- Проверить, что нет GPL/AGPL/LGPL в production-зависимостях.

### Ф.3. Сводный документ

- Создать `docs/ops/license-audit.md` со сводной таблицей, вердиктом и
  датой последнего аудита.
- Если ручных вердиктов нет (всё чисто) — короткий документ из 1–2
  абзацев + список разрешённых лицензий.

### Ф.4. CI-job

- Добавить `license-check` в `.github/workflows/ci.yml`:
  - matrix-стратегия по 4 Go-модулям;
  - отдельные шаги для 2 frontend-проектов;
  - `go-licenses` ставится через `go install` в шаге;
  - `license-checker` — через `npx` без global install.
- Добавить кэширование Go-модулей и npm чтобы не замедлять CI.

### Ф.5. CONTRIBUTING + финализация

- Короткий абзац в `CONTRIBUTING.md` про лицензии зависимостей.
- Обновить `docs/project-creation.txt`: итерация 40.
- Коммит: `chore(ci): аудит лицензий зависимостей и CI-проверка`.

---

## Тестирование

- Локальный прогон `go-licenses check` в каждом backend-модуле — exit 0.
- Локальный прогон `license-checker` в каждом frontend — без ошибок.
- В CI — pull request с искусственным добавлением GPL-пакета (например,
  `readline` через cgo-биндинг) проваливает job. Проверить и откатить
  тестовое изменение.
- В CI — обычный PR проходит license-check за разумное время (≤2 мин).

---

## Итог

Один новый CI-job + один документ + короткая правка CONTRIBUTING. 1 коммит.
Закрывает риск незаметного добавления GPL/AGPL-зависимости, который
аннулирует защиту PolyForm Noncommercial.
