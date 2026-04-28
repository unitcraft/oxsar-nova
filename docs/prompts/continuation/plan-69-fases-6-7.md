# Continuation: план 69 — Ф.6-Ф.7 (notes endpoint verify, финал)

**Применение**: вставить блок ниже в новую сессию Claude Code.
**Объём**: ~1-2 дня, ~50-150 строк правок.

---

```
Задача: завершить план 69 (ремастер) — Ф.6 (notes endpoint
verify) + Ф.7 (финализация).

КОНТЕКСТ:

План 69 уже сильно продвинут:
- Ф.0 дельта-аудит (коммит 80c7ef08e0) — нашёл существующие
  миграции, сократил план с 9 до 5 ALTER.
- Ф.1 миграция 0072_users_remaster_fields.sql (коммит 32d24a1f2b) —
  5 полей: max_points, protected_until_at, is_observer,
  last_planet_teleport_at, last_global_chat_read_at,
  last_ally_chat_read_at.
- Ф.3-Ф.5 handlers + защитная логика protected_until_at +
  cooldowns (коммит b47eb98f81).

Эта сессия: Ф.6 verify notes endpoint + Ф.7 финализация.

ПЕРЕД НАЧАЛОМ:

1) git status --short.

2) Прочитай:
   - docs/plans/69-remaster-domain-fields-extension.md (своё ТЗ +
     Ф.0 отчёт)
   - КОММИТЫ: 80c7ef08e0, 32d24a1f2b, b47eb98f81 — твоя работа.
   - Миграция 0050_notepad.sql — таблица user_notepad
     (user_id PK, content text, updated_at).

ЧТО НУЖНО СДЕЛАТЬ:

Ф.6. Notes endpoint verify:
- Найти существующий endpoint для user_notepad. Грепни:
  grep -rln "user_notepad\|notepad\|notes" projects/game-nova/backend
  + projects/game-nova/api/openapi.yaml.
- Если endpoint GET/PUT /api/users/me/notes УЖЕ ЕСТЬ — verify
  что работает (smoke-тест), отметить в плане как ✅.
- Если endpoint НЕТ — реализовать:
  · GET /api/users/me/notes (читает user_notepad).
  · PUT /api/users/me/notes (обновляет content + updated_at).
  · CHECK на размер ≤ 16KB (если ещё нет в миграции — добавить
    миграцией 00NN_notepad_size_check.sql, следующий свободный
    номер после 0078).
  · OpenAPI первым (R2).
  · Тесты на чтение/запись/превышение лимита.

Ф.7. Финализация:
- Шапка плана 69 → ✅ Завершён <дата>.
- docs/project-creation.txt — итерация 69.
- В divergence-log.md пометить ✅:
  · D-001 (max_points)
  · D-003 (account_deletion_scheduled_at — частично, основная
    реализация в 0051)
  · D-004 (protected_until_at)
  · D-005 (is_observer)
  · D-008 (profession_changed_at — закрыто 0046, не план 69)
  · D-016 (last_planet_teleport_at)
  · D-019 (отказ — обоснование в плане)
  · D-020 (last_global/ally_chat_read_at)
  · W1 (notes — закрыто 0050 user_notepad)
- В divergence-log.md ОТКАЗЫ:
  · D-007 (ui_theme/ui_pack — отказ, YAGNI)
  · D-021 (race — отказ, мёртвое поле)

ПРАВИЛА (R0-R15):
- R0: не правь modern-числа.
- R2: OpenAPI первым.
- R6: REST (PUT для обновления).
- R8: Prometheus метрики.
- R9: Idempotency для PUT (опционально, но желательно).
- R12: i18n — grep nova-bundle.
- R15: без упрощений.

GIT-ИЗОЛЯЦИЯ:
- Свои пути: internal/users/ или internal/notepad/, openapi.yaml,
  возможно миграция 00NN_notepad_size_check.sql, тесты,
  docs/plans/69-..., docs/project-creation.txt,
  docs/research/origin-vs-nova/divergence-log.md.

КОММИТ:

1 коммит: feat(users): notes endpoint verify + финализация (план 69 Ф.6+Ф.7).
ИЛИ если endpoint уже есть и нечего делать — просто
docs(plan-69): финализация (план 69 Ф.7).

ЧЕГО НЕ ДЕЛАТЬ:
- НЕ дублировать существующий endpoint, если он уже есть.
- НЕ менять modern-числа.
- НЕ забыть про R12 (i18n).

УСПЕШНЫЙ ИСХОД:
- Notes endpoint работает (через GET/PUT).
- Все D-NNN/W1 плана 69 закрыты ✅.
- D-007, D-021 в отказах с обоснованием.
- План 69 ✅.

Стартуй.
```
