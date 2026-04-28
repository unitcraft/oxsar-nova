# Блок: ОБЯЗАТЕЛЬНЫЕ ПРАВИЛА R0-R15 (включи в промпт via reference)

**Источник:** `docs/research/origin-vs-nova/roadmap-report.md`
«Часть I.5».

**Применение в промпте:** скопируй секцию ниже целиком в раздел
«ОБЯЗАТЕЛЬНЫЕ ПРАВИЛА» continuation/initial-промпта. Добавь
плана-специфические уточнения после.

---

```
ОБЯЗАТЕЛЬНЫЕ ПРАВИЛА (R0-R15, см. roadmap-report.md «Часть I.5»):

R0: ГЕЙМПЛЕЙ NOVA ЗАМОРОЖЕН — не правь modern-числа/механики/
формулы. Если число отличается между origin и nova — параметризуй
в configs/balance/<universe>.yaml, не меняй nova.

R1: snake_case в БД и YAML, snake_case JSON, английский, полные
слова. Валюта — `oxsar`/`oxsarit` без постфиксов (ADR-0009).

R2: OpenAPI первым — сначала схема в openapi.yaml, потом
сгенерированный TS-клиент, потом backend.

R3: log/slog с полями user_id, planet_id, event_id, trace_id для
каждого нового handler/service.

R4: golden + property-based тесты для battle/economy/event.
Покрытие изменённых строк ≥ 85%.

R5: pixel-perfect ТОЛЬКО в плане 72 (origin-фронт). В других
планах — nova-стиль.

R6: REST API с нуля — ресурсы во множ. числе, HTTP-методы по
семантике, JSON, snake_case полей. НЕ калькой с origin
?go=Page&action=...

R7: backward compat технических интерфейсов — N/A до публичного
запуска (плана 74). API-схемы / БД-миграции / payload event'ов
можно ломать. R7 НЕ распространяется на геймплей (это R0).

R8: Prometheus counter+histogram для каждого нового
handler/endpoint/service-метода. ОБЯЗАТЕЛЬНО, не trade-off.

R9: Idempotency-Key для всех мутирующих API (POST/PUT/DELETE).
ОБЯЗАТЕЛЬНО.

R10: per-universe изоляция данных — universe_id в новых таблицах
с per-universe семантикой + WHERE universe_id во всех SELECT/UPDATE.

R11: rate-limiting на публичных endpoint'ах (особенно UGC).

R12: i18n с самого начала. Все user-facing строки через
i18n.Tr(ctx, "key"). ОБЯЗАТЕЛЬНО grep по
projects/game-nova/configs/i18n/ru.yml и en.yml перед созданием
новых ключей. Цель — максимальное переиспользование.
В коммите указывать соотношение «переиспользовано/новых».

R13: typed payload через Go-struct + json schema (не сырой JSONB).

R14: migration policy для legacy-PHP-данных через миграцию
(`up.sql` или импорт-скрипт), не runtime-код.

R15: БЕЗ УПРОЩЕНИЙ КАК ДЛЯ ПРОДА — тесты, обработка ошибок,
метрики, безопасность, производительность со старта.
Никаких MVP-сокращений / TODO позже / "упростим до запуска".

R15 УТОЧНЕНО — что СЧИТАЕТСЯ упрощением vs ПРОПУСК:

🚫 ПРОПУСК (НЕ trade-off, считается багом — добавить и не отмечать
в simplifications.md):
- R8 Prometheus метрики для нового handler — 5-10 строк, всегда.
  Аргумент «низкочастотный endpoint» НЕ принимается.
- R9 Idempotency-Key для мутирующего endpoint'а — middleware
  существует, подключение 1-2 строки.
- R12 i18n — хардкод строки в Go-handler не оправдан, Tr() = тот
  же объём кода.
- R10 universe_id в новой per-universe таблице — 1 строка DDL.
- R3 slog с trace_id — стандартный паттерн.
- DB-level CHECK на размер VARCHAR/TEXT при наличии handler-лимита —
  defense-in-depth, тривиально.
- NOT NULL на колонке которая по семантике не nullable.
- FK constraint на ссылку.
- Index на FK или часто-фильтруемой колонке.

✅ TRADE-OFF (можно записать в simplifications.md с обоснованием):
- Архитектурное отклонение от плана с причиной + планом возврата.
- Не реализованная фича из плана, явно отложенная до Ф.X.
- Реализация требует существенно больше работы чем доступно
  (1+ часов агента) — отложить с явным «Приоритет M/L».

✅ НЕ упрощение (тривиальные адаптации к реальности):
- Имя поля отличается от плана, но сделано лучше архитектурно.
- Лимит другой если обоснован (handler уже работает с этим
  значением).
- Endpoint существовал — задокументирован задним числом.
```
