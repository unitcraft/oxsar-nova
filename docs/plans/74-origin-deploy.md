# План 74: origin deploy + DNS + config

**Дата**: 2026-04-28
**Статус**: Скелет (детали допишет агент-реализатор при старте)
**Зависимости**: блокируется планами 72 (origin-фронт готов) +
73 (CI зелёный); ADR-0010 (имя поддомена — открытый вопрос).
**Связанные документы**:
- [62-origin-on-nova-feasibility.md](62-origin-on-nova-feasibility.md)
- [docs/research/origin-vs-nova/roadmap-report.md](../research/origin-vs-nova/roadmap-report.md) —
  R1-R5 + раздел плана 74
- [ADR-0010](../adr/0010-universe-domain-naming.md) — открытый
  вопрос про префикс поддомена.
- [36-portal-multiverse.md](36-portal-multiverse.md) — registry
  вселенных.

---

## Цель

Поднять origin-вселенную как **третью** рядом с uni01/uni02 в
prod. Финальный план серии ремастера — после него игроки могут
начать играть в legacy-режим на новом стеке.

---

## Что делаем

### DNS / поддомен

- Зарегистрировать поддомен по решению ADR-0010
  (`origin.oxsar-nova.ru` / `classic.oxsar-nova.ru` / другое).
- Указать на CDN / nginx с deploy origin-фронта.

### Deploy

- Свой Vite-bundle origin-frontend в CDN (CI-job).
- Backend — общий game-nova (новых процессов не нужно, scaling
  работает через план 32).
- nginx / Caddy конфиг для нового поддомена.

### Конфигурация

- CORS: добавить новый origin в `ALLOWED_ORIGINS` portal-backend
  (план 56), admin-bff (план 53), game-nova (для handoff).
- Регистрация в `universes` registry (план 36): запись с
  `code = 'origin'` (или другое из ADR-0010), `display_name`
  на русском («Origin», «Классика» — по решению).
- Override-файл `configs/balance/origin.yaml` уже в репо
  (план 64) — `code='origin'` автоматически активирует override
  через `LoadFor("origin")`.
- Identity / billing — переиспользуют общие сервисы, конфигурация
  вселенной только в game-nova и portal-frontend.

### Smoke

- Регистрация нового пользователя через portal → вход во
  все три вселенные → выбор origin.
- Постройка / атака / чат / альянс — работают.
- Screenshot-diff (план 73) на проде проходит.

### Release notes

- Объявление в portal-новостях (план 36) про новую вселенную.
- Документация в `docs/release-roadmap.md`: ремастер запущен.

---

## Что НЕ делаем

- Не миграцируем игроков из game-origin-php в новую origin —
  это fresh start (см. план 62, 0 игроков сейчас).
- Не выключаем game-origin-php сразу — он работает на :8092 как
  fallback / референс ещё месяц-два, пока не подтвердится
  стабильность.
- Не вводим cross-universe механики (торговля между universes,
  единый рейтинг) — это будущие отдельные планы.

## Этапы (детали — при старте)

- Ф.1. ADR-0010 закрыть (выбор имени) — это блокер.
- Ф.2. DNS / поддомен.
- Ф.3. CDN / nginx конфигурация.
- Ф.4. CORS обновления (portal-backend, admin-bff, game-nova).
- Ф.5. universes registry-запись.
- Ф.6. Smoke в проде.
- Ф.7. Release notes + документация.
- Ф.8. Финализация.

## Конвенции (R1-R5)

- Имя поддомена — по ADR-0010.
- `universes.code` — snake_case (`origin`, `origin_classic` или
  что выберет ADR-0010), не camelCase. Имя файла override =
  `configs/balance/<code>.yaml`.
- `display_name` — русский («Origin», «Классика», «Андромеда» —
  по решению).

## Объём

1 неделя. Конфигурация + smoke.

## References

- ADR-0010 — выбор имени.
- План 36 — registry вселенных.
- План 32 — multi-instance scaling.
- План 50 — game-origin-php юр-готов (на :8092 fallback).
