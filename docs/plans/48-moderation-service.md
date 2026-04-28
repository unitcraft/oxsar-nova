# План 48: Moderation Service — общий микросервис UGC-модерации

**Дата**: 2026-04-27
**Статус**: Черновик (запускать при наступлении триггера, см. ниже)
**Зависимости**: план 46 закрыт (есть рабочий blacklist + жалобы); план 43
закрыт (game-origin живой и стабильный).
**Связанные документы**:
[46-age-rating-ugc.md](46-age-rating-ugc.md),
[../ops/ugc-moderation.md](../ops/ugc-moderation.md).

---

## Зачем отдельный сервис

После плана 46 у нас есть UGC-фильтр в **двух местах**:

- `projects/game-nova/backend/internal/moderation/` — Go, читает
  `configs/moderation/blacklist.yaml`.
- `projects/identity/backend/internal/moderation/` — Go, копия того же кода
  (разные модули → не получилось общего пакета).
- В `projects/game-origin-php/` (PHP) сейчас **нет фильтра вообще** — план 48
  Шаг 0 закроет это локальной PHP-обёрткой над тем же YAML, без сети.

Когда захочется заменить простой blacklist на что-то посильнее (Yandex
Cloud Content Filter, локальная ML-модель, Two Hat) — придётся
менять в трёх местах. Микросервис снимает эту проблему за счёт сетевого
hop'а.

**Это не делается заранее** — пока хватает blacklist'а, +1 сервис в
инфраструктуре (deploy, monitoring, security) обойдётся дороже выгоды.

## Триггеры запуска плана

Делаем, когда сработает **любое** из:

1. **Реальные обходы**: в админке `/api/admin/reports` появляются
   подтверждённые жалобы на l33t-speak / транслит / намеренные
   опечатки, которые blacklist пропускает. Порог — 5 подтверждённых
   случаев за неделю.
2. **Штраф или предупреждение РКН** на конкретный пропущенный контент.
3. **Рост ручной модерации** настолько, что 24-часовой SLA из
   [ugc-moderation.md](../ops/ugc-moderation.md) перестаёт держаться
   (>10 жалоб в день стабильно).
4. **Третий потребитель** появляется (например, `oxsar/forum-service`
   или комментарии к артефактам) — три потребителя одной логики уже
   оправдывают сервис.

До триггера — план в черновике, к нему не возвращаемся.

---

## Архитектура

### Контракт

REST, не WS — обращения на каждое сообщение чата терпят 5–20ms
сетевой hop. Если латентность станет проблемой — кэш в потребителях.

```
POST /moderate
Body: {
  "text": "...",
  "context": "username|alliance_name|chat_message|planet_desc|...",
  "user_id": "..."          // опционально, для rate-limit
}
Response 200: {
  "decision": "allow" | "block" | "review",
  "reason":   "profanity|drugs|extremism|...",
  "match":    "найденный корень или фрагмент",
  "score":    0.0–1.0       // confidence (релевантно для ML, для
                            //   blacklist всегда 1.0 при block)
}
```

`review` — отдельный статус «не уверен, отдай модератору в очередь
жалоб». В реализации шага 1 не используется (только allow/block);
появляется на шаге 2 при подключении ML.

### Дополнительные эндпоинты

```
POST /moderate/batch        # массив текстов за один запрос — для
                              истории чата при загрузке.
GET  /healthz                # readiness/liveness.
GET  /metrics                # Prometheus.
GET  /config/version         # версия загруженного blacklist'а — чтобы
                              клиенты могли инвалидировать кэш.
```

### Модули

`projects/moderation/backend/` — четвёртый Go-модуль (по аналогии
с `auth`, `billing`, `game-nova`).

```
cmd/server/         # main, env-config (MOD_ADDR, BLACKLIST_PATH, REDIS_URL)
internal/
  modsvc/           # бизнес-логика: backend (blacklist v1, ML v2, …)
  storage/          # Postgres (audit-log решений, опционально)
  httpx/            # стандартный response/error wrapper
pkg/
  metrics/          # prometheus
```

### Backend strategy

Шаг 1 (на запуске сервиса) — единственный backend `BlacklistBackend`,
читающий тот же `configs/moderation/blacklist.yaml`. По сути — перенос
текущей логики из `internal/moderation` в отдельный сервис.

Шаг 2 (когда нужно сильнее) — добавляется `YandexCloudBackend`,
`LocalMLBackend`. Цепочка backend'ов: blacklist → cloud → ml; на
первом `block` останавливаемся.

### Audit log

Таблица `moderation_decisions` (опционально на шаге 1, обязательно
на шаге 2):

```
id, ts, user_id (nullable), context, text_hash, text_preview (200ch),
decision, reason, backend, score
```

Текст полностью **не храним** — только sha256-хэш + первые 200
символов. Это ПДн (уровень 4 по 152-ФЗ — не специальные категории, но
лучше не накапливать) и предмет потенциальных судебных запросов.

---

## Этапы

### Шаг 0 — PHP-обёртка для game-origin (БЫСТРЫЙ ВАРИАНТ)

**Делается прямо сейчас**, до запуска основного плана. Закрывает
«один источник истины» без микросервиса.

- `projects/game-origin-php/src/core/moderation/Blacklist.class.php` —
  читает `projects/game-nova/configs/moderation/blacklist.yaml` (путь
  через конфиг), реализует `isForbidden(string $input): bool`.
- Нормализация — точно такая же, как в Go-версии (lowercase, удаление
  всего, что не буква). Тесты — паритет с Go-тестами.
- Использование: при регистрации (`User::create`), смене ника,
  создании альянса, отправке сообщения в legacy-чат.

После шага 0 у нас три копии **логики** (Go × 2 + PHP), но один YAML.
Это допустимо: в микросервисном будущем код выкинется, останется
маленький REST-клиент.

### Шаг 1 — Скелет сервиса + миграция потребителей

Запускается при срабатывании триггера.

- Go-модуль `projects/moderation/backend/` (cmd/server, internal/modsvc).
- Контракт `POST /moderate` + `/healthz` + `/metrics`.
- Один backend — `BlacklistBackend` (тот же YAML-источник).
- Тесты: golden-тесты совместимости с текущей `IsForbidden`
  (тот же набор inputs — те же ответы).
- docker-compose dev-stack: добавить service.
- nginx config: внутренний CIDR-only, не наружу.

### Шаг 2 — Миграция game-nova/auth/game-origin на сетевой клиент

- В `auth/internal/moderation` и `game-nova/internal/moderation` —
  переключаем `IsForbidden` на HTTP-клиента к moderation-service.
  Старая локальная логика остаётся как fallback (если сервис
  недоступен — логируем warning + используем локальный blacklist).
- В `game-origin` — Guzzle-клиент на тот же endpoint, та же
  fallback-семантика (локальная PHP-копия).
- Тесты: integration через docker-compose, проверка fallback'а
  (остановили moderation-service → потребители продолжают работать
  на локальном blacklist).

### Шаг 3 — Сторонний сервис / ML

- Подключение `YandexCloudBackend` (Yandex Cloud Content Filter,
  REST-API). Конфиг — endpoint + API-key через env. Только для
  контента в РФ-юрисдикции (из памяти: hetzner/cloudflare непригодны).
- Цепочка backend'ов в `modsvc.Service`: сначала blacklist (быстро),
  потом cloud (если blacklist=allow), на `block` останавливаемся.
- Audit log включается обязательно — нужно понимать, что именно
  cloud режет, чтобы правильно подкручивать пороги.
- Опциональный `LocalMLBackend` — fine-tune'нный rubert-tiny на
  ru-toxic-corpus. Делается, только если cloud-сервис ломает
  юнит-экономику.

### Шаг 4 — Расширение контекста

`/moderate` принимает `context`, чтобы фильтр был контекстным:

- `username` — строгий, blacklist + reserved + длина.
- `chat_message` — мягкий, разрешает ругательство если оно не
  направлено на другого игрока (опционально).
- `alliance_name` — строгий.

На шаге 1 контекст игнорируется (один blacklist на всё). На шаге 3
backend может использовать его для разных порогов / разных моделей.

---

## Что НЕ делаем

- Не пишем собственный ML с нуля — fine-tune готовых моделей
  (rubert-tiny, deberta-v3-russian-toxicity) дешевле и точнее.
- Не делаем стриминг/WebSocket для chat — REST на каждое сообщение
  достаточно, latency 5–20ms незаметна на фоне сетевого WS-hop'а.
- Не пытаемся ловить контекстные намёки и сарказм — это уже не
  модерация, а тяжёлый NLP. Для оскорблений по существу хватает
  blacklist + cloud.
- Не делаем «прогрев» blacklist'а через crowdsourcing на старте —
  ручной список + жалобы пользователей дают тот же результат с
  меньшим оверхедом.

---

## Риски

- **Latency**: на каждое сообщение в чате — сетевой hop. Mitigation:
  кэш в потребителях по text_hash, TTL 1 час. Для blacklist-backend'а
  ответ детерминирован, кэш безопасен.
- **Single point of failure**: упадёт moderation-service — упадут
  регистрация и чат. Mitigation: fallback на локальный blacklist в
  каждом потребителе (см. шаг 2).
- **Юрисдикция Cloud-провайдера**: Yandex Cloud — РФ, ОК. Cloudflare,
  Perspective API (Google) непригодны (правило проекта: VPS/CDN/
  managed только из РФ).
- **Накопление ПДн в audit-log**: храним хэш + 200 символов превью.
  Полный текст не храним. Превью обнуляется через 6 месяцев (TTL по
  cron, как chat_messages).

---

## Итог

Шаг 0 — пишем сейчас, ~50 строк PHP + общий YAML, без микросервиса.

Шаги 1–4 — лежат черновиком до триггера. Когда триггер случится —
оценка работы 1–2 недели на шаги 1–2, шаг 3 — отдельным заходом
после стабилизации.
