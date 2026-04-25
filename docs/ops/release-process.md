# Релиз-процесс

**Дата**: 2026-04-26
**Контекст**: ответ на вопрос «как будет происходить релиз при trunk-based
разработке, если коммитим в `main`».

## TL;DR

- `main` всегда релизо-готов; коммитим туда (через PR из коротких feature-веток
  или напрямую — как сейчас).
- Релиз = git-тег по SemVer (`v0.X.Y`) на коммите `main`.
- CD-workflow по тегу собирает образы, пушит в registry; деплой на VPS — `docker
  compose pull && up -d`.
- Незрелые фичи прячутся за feature flags (план 31 Ф.2 ✅), поэтому на `main`
  можно мержить даже наполовину готовое.
- Release-ветки **не нужны**, пока в проде живёт одна версия (одиночная игра,
  один VPS).

## Что уже есть

| Компонент | Состояние |
|---|---|
| CI ([.github/workflows/ci.yml](../../.github/workflows/ci.yml)) | ✅ lint+test+build+e2e+security на push в `main` и PR |
| Prod-стек ([deploy/docker-compose.prod.yml](../../deploy/docker-compose.prod.yml)) | ✅ собирается локально на хосте через `--build` |
| Health/ready endpoints | ✅ план 31 Ф.1 |
| Feature flags | ✅ план 31 Ф.2 |
| Graceful shutdown (SIGTERM + 30s) | ✅ [backend/cmd/server/main.go:482](../../backend/cmd/server/main.go#L482) |
| Git-теги | ❌ пока нет ни одного |
| CD workflow (release.yml) | ❌ нет |
| Image registry (GHCR) | ❌ не используется, prod билдит локально |

## Схема релиза

### 1. Версионирование тегами (SemVer)

```bash
# main зелёный, хочется выкатить:
git tag -a v0.3.0 -m "release: scheduler + chat fan-out"
git push origin v0.3.0
```

- `v0.X.0` — feature-релиз (новый план/итерация).
- `v0.X.Y` — патч/хотфикс.
- `v1.0.0` — публичный запуск (см. [release-roadmap.md](../release-roadmap.md)).

### 2. Release workflow на тег

Новый `.github/workflows/release.yml`, триггер `on: push: tags: ['v*']`:

1. Билдит `backend` и `frontend` образы.
2. Пушит в **GHCR** (`ghcr.io/<owner>/oxsar-nova-{backend,frontend}`)
   с тегами `:v0.3.0` **и** `:latest`.
3. Опционально: SSH на VPS → `docker compose pull && docker compose up -d`.
   На старте — пропускаем, дёргаем вручную на VPS.

### 3. Деплой на VPS

Поправить [deploy/docker-compose.prod.yml](../../deploy/docker-compose.prod.yml):
вместо `build:` использовать pinned image:

```yaml
services:
  backend:
    image: ghcr.io/<owner>/oxsar-nova-backend:${VERSION:-latest}
  frontend:
    image: ghcr.io/<owner>/oxsar-nova-frontend:${VERSION:-latest}
```

Выкатка на VPS:

```bash
echo "VERSION=v0.3.0" > .env
docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.prod.yml pull
docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.prod.yml up -d
```

Откат: меняем `VERSION=v0.2.5` в `.env` и тот же `up -d`.

### 4. Хотфикс

- Баг в `v0.3.0` → коммит в `main` (быстрый PR) → тег `v0.3.1` → CD катит.
- Если `main` уже ушёл далеко вперёд и катить целиком нельзя — **только
  тогда** делается `release/v0.3` от тега `v0.3.0`, туда черри-пик фикса,
  тег `v0.3.1` от ветки. Для одиночной игры — редкий случай.

### 5. Откат БД-миграций

Goose поддерживает `down`, но миграции у нас **forward-only** в проде
(см. [oxsar-spec.txt](../oxsar-spec.txt) §17). Если релиз сломал данные —
восстановление из бэкапа ([deploy/backup.sh](../../deploy/backup.sh)),
а не goose down. Поэтому: breaking-change миграции выкатываются отдельным
релизом без кода, который их использует (expand-then-contract).

## Что нужно сделать (минимальный set-up)

1. **`.github/workflows/release.yml`** — собирает и пушит образы по тегу
   в GHCR (~80 строк YAML).
2. **Правка [deploy/docker-compose.prod.yml](../../deploy/docker-compose.prod.yml)** —
   `image:` вместо `build:`, версия из `${VERSION}`.
3. **Первый тег** `v0.1.0` на текущем `main` для отсчёта (опционально — после
   того, как закроются оставшиеся фазы плана 31/32).

Auto-deploy через SSH из workflow добавляется позже, когда настроены
deploy-ключи на VPS.

## Почему не release-ветки

Trunk-based + теги работает для одиночной игры, потому что:
- В проде живёт **одна** версия — нет необходимости поддерживать `v0.2.x`
  параллельно с `v0.3.x`.
- Хотфикс быстрее накатить на свежий `main` (там меньше отставание от прода),
  чем держать долгоживущую release-ветку.
- Меньше merge-конфликтов, меньше когнитивной нагрузки.

Release-ветки оправданы для:
- Поддержки нескольких major-версий в проде (SaaS с self-hosted клиентами).
- Регулируемых релизов (медицина, банки) с длинным циклом QA на конкретной версии.

Ни то, ни другое к oxsar-nova не относится.

## Ссылки

- [release-roadmap.md](../release-roadmap.md) — приоритеты до v1.0.
- [docs/plans/31-zero-downtime-deploy.md](../plans/31-zero-downtime-deploy.md) —
  health/ready, feature flags, blue-green (отложен).
- [vps-sizing.md](vps-sizing.md) — железо под разные DAU.
- [scaling.md](scaling.md) — горизонтальное масштабирование.
