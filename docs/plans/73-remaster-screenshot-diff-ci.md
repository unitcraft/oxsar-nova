# План 73 (ремастер): Screenshot-diff CI (Playwright + visual regression)

**Дата**: 2026-04-28
**Статус**: Скелет (детали допишет агент-реализатор при старте)
**Зависимости**: блокируется планом 72 (хотя бы первые экраны
готовы для эталонов).
**Связанные документы**:
- [62-origin-on-nova-feasibility.md](62-origin-on-nova-feasibility.md)
- [docs/research/origin-vs-nova/origin-ui-replication.md](../research/origin-vs-nova/origin-ui-replication.md)
- [docs/research/origin-vs-nova/roadmap-report.md](../research/origin-vs-nova/roadmap-report.md) —
  R1-R5 + раздел плана 73
- [docs/legacy/game-origin-access.md](../legacy/game-origin-access.md) —
  запущенный legacy-PHP origin на :8092 (источник эталонов)

---

## Цель

Автоматизированное сравнение нового origin-фронта (план 72) со
скриншотами эталонного game-origin-php. Гарантирует, что
pixel-perfect не «уплывает» при правках UI.

---

## Что делаем

### Снятие эталонов

- Скрипт `tests/e2e/origin-baseline/take-screenshots.sh`
  (Playwright headed):
  - Поднять `game-origin-php` на :8092 (через docker-compose из
    плана 50 / docs/legacy/game-origin-access.md).
  - Залогиниться dev-логином.
  - Пройти **все 50 prod-экранов** S-NNN.
  - Сохранить PNG в `tests/e2e/origin-baseline/screenshots/`.
- Эталоны коммитятся в репо (это «золото» для diff-теста).
- При **намеренном** изменении legacy-UI (если когда-то
  захочется) — отдельный скрипт `update-baselines.sh`
  + явная фиксация в коммит-сообщении.

### Тесты на новый origin-фронт

- Playwright suite в `tests/e2e/origin-frontend/`.
- Для каждого S-NNN экрана:
  - Поднять новый origin-фронт + nova-backend.
  - Залогиниться + перейти на экран.
  - Снять скриншот.
  - Сравнить с baseline через `pixelmatch` (threshold 0.5%).
  - Если diff > threshold — fail с указанием файла-эталона.

### CI-job

- `.github/workflows/screenshot-diff.yml`:
  - Поднимает оба стека (game-origin-php + новый origin-frontend +
    nova-backend) в matrix.
  - Запускает Playwright suite.
  - При diff > threshold — fail PR.
  - Артефакт: HTML-отчёт с diff-картинками.

### Регламент

- При намеренном изменении legacy-UI (новые фишки) — обновление
  baselines делается отдельным коммитом + явное обоснование в
  release notes.
- При технических правках (refactor) — diff должен быть 0%.

---

## Что НЕ делаем

- Не сравниваем nova-frontend (uni01/uni02) — у него свой визуал,
  не legacy.
- Не запускаем visual diff на каждый PR в backend — только на
  origin-frontend changes.
- Не делаем pixel-perfect с допуском 0% (бессмысленно из-за
  font-rendering вариаций) — реалистичный 0.5%.

## Этапы (детали — при старте)

- Ф.1. Скрипт снятия эталонов с game-origin-php.
- Ф.2. Snapshot первого набора (5-7 экранов, проверить процесс).
- Ф.3. Playwright suite на новый origin-фронт.
- Ф.4. CI-job + matrix.
- Ф.5. Регламент обновления baselines.
- Ф.6. Финализация.

## Конвенции (R1-R5)

- Имена baseline-файлов: `s-001-main.png`, `s-002-constructions.png`
  (snake_case, ID + название экрана).
- Threshold 0.5% — задокументировать в скрипте.
- Игнор-зоны через Playwright `mask:` опцию:
  - Текущее время в шапке.
  - **Баннеры и рекламные блоки** в legacy-PHP — намеренно
    отсутствуют в новом фронте (см. roadmap-report «Часть V»),
    их области в эталонах маскируются.
  - Экраны, выведенные из первой итерации (Achievements,
    Tutorial) — **не снимаются** в baseline-набор; в новом
    фронте их нет.
- При снятии эталонов — **dev-инстанс legacy-PHP** должен быть
  настроен **без рекламных блоков** (если их можно отключить
  через config), чтобы маски не понадобились. Если отключить
  нельзя — маскировать.

## Объём

2 недели. Скрипт + Playwright suite + CI.

## References

- origin-ui-replication.md — список 50 экранов для baseline.
- docs/legacy/game-origin-access.md — как поднять game-origin-php.
- Существующая Playwright-инфраструктура nova
  (`projects/game-nova/frontend/e2e/`) — паттерны.
