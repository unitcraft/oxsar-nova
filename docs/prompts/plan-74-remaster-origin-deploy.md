# Промпт: выполнить план 74 (origin deploy)

**Дата создания**: 2026-04-28
**План**: [docs/plans/74-remaster-origin-deploy.md](../plans/74-remaster-origin-deploy.md)
**Зависимости**: блокируется планами 72 (origin-фронт готов) + 73
(CI зелёный); ADR-0010 закрыт ✅ (`origin.oxsar-nova.ru`).
**Объём**: 1 нед.

---

```
Задача: выполнить план 74 (ремастер) — поднять origin-вселенную
как третью рядом с uni01/uni02 в production. Финальный план серии
ремастера.

ПЕРЕД НАЧАЛОМ:

1) git status --short.

2) Прочитай:
   - docs/plans/74-remaster-origin-deploy.md
   - docs/research/origin-vs-nova/roadmap-report.md «Часть I.5»
     (R0-R15)
   - docs/adr/0010-universe-domain-naming.md (Accepted:
     `origin.oxsar-nova.ru`).
   - docs/release-roadmap.md «Пост-запуск v3» (включая «Breaking
     changes для modern-вселенных перед запуском origin»).

3) Выборочно:
   - План 36 (multiverse / universes registry).
   - План 32 (multi-instance scaling).
   - План 50 (legacy-PHP origin на :8092 — fallback).

ЧТО НУЖНО СДЕЛАТЬ:

1. DNS / поддомен:
   - `origin.oxsar-nova.ru` (по ADR-0010).
   - Указать на CDN / nginx с deploy origin-фронта.

2. Deploy origin-фронта:
   - Свой Vite-bundle в CDN (CI-job).
   - Backend = общий game-nova (новых процессов не нужно, scaling
     через план 32).
   - nginx / Caddy конфиг для нового поддомена.

3. CORS / ALLOWED_ORIGINS:
   - Добавить `https://origin.oxsar-nova.ru` в:
     · portal-backend (план 56).
     · admin-bff (план 53).
     · game-nova (для handoff).

4. Регистрация в universes registry:
   - INSERT в `universes`: `code='origin'`,
     `display_name='Origin'` (или другой русский вариант),
     `created_at=NOW()`.
   - Override активируется автоматически наличием
     `configs/balance/origin.yaml` (план 64).

5. Smoke в проде:
   - Регистрация нового пользователя через portal → вход во
     все три вселенные (uni01, uni02, origin) → выбор origin.
   - Постройка / атака / чат / альянс / биржа — работают.
   - Screenshot-diff (план 73) на проде проходит.

6. Release notes:
   - Объявление в portal-новостях (план 36) про новую вселенную.
   - **Breaking changes для modern**: см. release-roadmap.md
     раздел «Breaking changes для modern-вселенных перед запуском
     origin» — release notes для uni01/uni02 о появляющихся
     изменениях (AlienAI, спец-юниты, биржа, B1 дипстатусы,
     buddy-list, TELEPORT_PLANET).

7. Документация:
   - docs/release-roadmap.md — ремастер запущен ✅.
   - docs/project-creation.txt — итерация 74.

ОБЯЗАТЕЛЬНЫЕ ПРАВИЛА:

R0: НЕ трогать geymplay при деплое.
R10: убедиться что universe_id корректно работает (новая вселенная
не утекает в данные uni01/uni02).
R11: rate-limit правила для нового origin.
R15: smoke-тесты обязательны перед объявлением запуска.

GIT-ИЗОЛЯЦИЯ:
- Свои пути: deploy/ (nginx-config, docker-compose),
  .github/workflows/, infrastructure-конфиги, миграция SQL для
  registry, docs/plans/74-..., docs/release-roadmap.md.

КОММИТЫ:

2 коммита:
1. feat(deploy): origin universe registry + CORS + nginx (план 74).
2. docs(release): анонс origin + breaking changes для modern (план 74 финализация).

ЧЕГО НЕ ДЕЛАТЬ:
- НЕ миграцируем игроков из legacy-PHP-origin :8092 в новую origin.
  Fresh start (0 игроков, см. план 62).
- НЕ выключаем legacy-PHP origin на :8092 сразу — он работает как
  fallback / референс ещё месяц-два.
- НЕ вводим cross-universe механики (торговля между universes,
  единый рейтинг) — будущие отдельные планы.

УСПЕШНЫЙ ИСХОД:
- `origin.oxsar-nova.ru` доступен.
- universes registry содержит code='origin'.
- Smoke в проде успешен.
- Release notes объявлены.
- Серия ремастера 64-74 закрыта.
- Можно запускать план 70 (achievements реактивация).

Стартуй.
```
