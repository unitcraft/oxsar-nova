# Промпт: выполнить план 73 (screenshot-diff CI)

**Дата создания**: 2026-04-28
**План**: [docs/plans/73-remaster-screenshot-diff-ci.md](../plans/73-remaster-screenshot-diff-ci.md)
**Зависимости**: блокируется планом 72 (хотя бы первые экраны).
**Объём**: 2 нед.

---

```
Задача: выполнить план 73 (ремастер) — screenshot-diff CI
(Playwright + visual regression) для проверки pixel-perfect
паритета origin-фронта с legacy-PHP origin.

ВАЖНОЕ:
- Зависит от плана 72 (хотя бы первые экраны Spring 1).
- legacy-PHP origin доступен на :8092 (см. docs/legacy/
  game-origin-access.md).

ПЕРЕД НАЧАЛОМ:

1) git status --short.

2) Прочитай:
   - docs/plans/73-remaster-screenshot-diff-ci.md
   - docs/research/origin-vs-nova/origin-ui-replication.md
   - docs/research/origin-vs-nova/roadmap-report.md «Часть I.5»
     (R0-R15)
   - docs/legacy/game-origin-access.md — как поднять legacy-PHP
     origin на :8092.

3) Выборочно:
   - projects/game-nova/frontend/e2e/ — паттерны Playwright nova.

ЧТО НУЖНО СДЕЛАТЬ:

1. Скрипт снятия эталонов tests/e2e/origin-baseline/take-screenshots.sh:
   - Поднять legacy-PHP origin на :8092 (через docker-compose).
   - **Важно: dev-инстанс legacy-PHP без рекламных блоков**
     (если можно отключить через config). Если нет — маски в
     Playwright.
   - Залогиниться dev-логином.
   - Пройти все 50 prod-экранов S-NNN.
   - Сохранить PNG в tests/e2e/origin-baseline/screenshots/.
   - НЕ снимать Achievements (S-Achievements) и Tutorial (S-Tutorial) —
     они исключены из плана 72.
   - Эталоны коммитятся в репо.

2. Playwright suite tests/e2e/origin-frontend/:
   - Для каждого S-NNN экрана:
     · Поднять новый origin-фронт + nova-backend.
     · Залогиниться + перейти на экран.
     · Снять скриншот.
     · Сравнить с baseline через pixelmatch (threshold 0.5%).
     · Если diff > threshold — fail с указанием файла-эталона.

3. CI-job .github/workflows/screenshot-diff.yml:
   - Поднимает оба стека (legacy-PHP origin :8092 + новый
     origin-frontend + nova-backend) в matrix.
   - Запускает Playwright suite.
   - При diff > threshold — fail PR.
   - Артефакт: HTML-отчёт с diff-картинками.

4. Игнор-зоны (Playwright mask:):
   - Текущее время в шапке.
   - Баннеры и рекламные блоки в legacy-PHP — намеренно отсутствуют
     в новом фронте, маскируются.
   - Динамические числа (счётчики ресурсов в шапке — фиксировать
     через mock-данные).

5. Регламент обновления baselines:
   - При намеренном изменении legacy-UI (если когда-то захочется) —
     отдельный скрипт update-baselines.sh + явное обоснование в
     release notes.
   - При технических правках (refactor) — diff должен быть 0%.

ОБЯЗАТЕЛЬНЫЕ ПРАВИЛА:

R0: не трогаем geymplay nova/origin при настройке CI.
R5: pixel-perfect — это R5, threshold 0.5% обоснован в плане 73.
R15: без упрощений — все 50 экранов покрыты, не «10 для проверки».

GIT-ИЗОЛЯЦИЯ:
- Свои пути: tests/e2e/origin-baseline/, tests/e2e/origin-frontend/,
  .github/workflows/screenshot-diff.yml, docs/plans/73-...

КОММИТЫ:

2 коммита:
1. feat(ci): screenshot-diff baselines (план 73 Ф.1+Ф.2).
2. feat(ci): Playwright suite + workflow + регламент (план 73 Ф.3-Ф.6).

ЧЕГО НЕ ДЕЛАТЬ:
- НЕ снимать Achievements/Tutorial (они отложены).
- НЕ запускать на каждый PR в backend — только на origin-frontend
  changes.
- НЕ делать threshold 0% (бессмысленно из-за font-rendering).

УСПЕШНЫЙ ИСХОД:
- 50 эталонов в tests/e2e/origin-baseline/screenshots/.
- Playwright suite зелёный для готовых экранов плана 72.
- CI-job настроен.
- Регламент в плане 73 финализирован.

Стартуй.
```
