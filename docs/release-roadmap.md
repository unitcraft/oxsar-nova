# Roadmap до запуска в прод

**Дата составления**: 2026-04-24
**Источники**: [status.md](status.md), [balance/audit.md](balance/audit.md),
[plans/](plans/), [simplifications.md](simplifications.md).

Сквозной приоритетный список того, что осталось закрыть перед релизом.
Живой документ — обновлять при закрытии пунктов.

---

## 🔴 Блокеры — без этого запускать нельзя

| # | Что | Почему блокер | Объём |
|---|---|---|---|
| ~~[18 Фаза 1](plans/18-unit-rebalance.md)~~ | ~~Порт недостающих rapidfire из legacy (38 записей)~~ | ✅ **done 2026-04-24** — BA-004 patched, BA-001/002 частично patched (ждёт симуляции). | — |
| ~~[21 B](plans/21-gameplay-hardening.md)~~ | ~~Экспедиции: min_fleet + reward cap~~ | ✅ **done 2026-04-24** — BA-003 patched (лимит 50k fleet + reward cap ×3 + lost scaling + полное уничтожение при lost из плана 17 B1). | — |
| [21 C](plans/21-gameplay-hardening.md) | Щиты: golden-тест high-tech | BA-005 — turtle-стратегия с Small Shield + Shield Tech 10+ может оказаться too strong. Сначала симуляция. | S |
| ~~[20 Ф.1](plans/20-legacy-port.md)~~ | ~~Vacation mode~~ | ✅ **done 2026-04-24** — min 48h, blocking events, target/sender shields, prod=0, /api/me fields. Auto-disable через 30d — остаток (не блокер). | — |
| ~~[20 Ф.2](plans/20-legacy-port.md)~~ | ~~Fleet slots через computer_tech~~ | ✅ **done 2026-04-24** — `1 + floor(computer_tech/6)`, UI-индикатор, 409 при overflow. | — |
| ~~[17 A1](plans/17-gameplay-improvements.md)~~ | ~~Антибашинг~~ | ✅ **done 2026-04-24** — max 4 атаки (ATTACK_SINGLE/ACS) от attacker→defender за 5h, ErrBashingLimit → 409. | — |
| ops | Резервное копирование БД + мониторинг | В repo нет. Без бэкапа первый сбой = смерть проекта. | M |

**Вместе:** ~1–1.5 недели работы, из которых 70% — планы 21+18.

---

## 🟡 Важные — желательно до массового анонса

| # | Что | Почему | Объём |
|---|---|---|---|
| ~~[20 Ф.3](plans/20-legacy-port.md)~~ | ~~Миссия POSITION~~ | ✅ **done 2026-04-25** — mission=6, ALLY/NAP target validation, PositionArriveHandler. | — |
| ~~[20 Ф.4](plans/20-legacy-port.md)~~ | ~~Сенсорная Фаланга~~ | ✅ **done 2026-04-25** — GET /api/phalanx, формула radius из legacy, 5000H за скан. | — |
| ~~[19](plans/19-game-wiki.md)~~ | ~~Вики игры~~ | ✅ **MVP done 2026-04-25** — backend API + генератор из configs. Frontend UI (`/wiki`) — отложен. | — |
| ~~[14 Ф.2.4–2.6](plans/14-admin-expansion.md)~~ | ~~Force-recall fleet, soft-delete user~~ | ✅ **done 2026-04-25** — admin endpoints для флота/планет/user-delete с auditable. | — |
| ~~[14 Ф.8.2](plans/14-admin-expansion.md)~~ | ~~Rate-limit админ-действий~~ | ✅ **done 2026-04-25** — 100 write/hour на админа, 429 + Retry-After. | — |
| ~~payment~~ | ~~Второй шлюз (ЮKassa или Enot.io)~~ | ✅ **done 2026-04-25** — Enot.io gateway + 5 unit-тестов. PAYMENT_PROVIDER=enot переключает. | — |

---

## 🟢 Nice-to-have — после первой волны игроков

- [24 ai-players](plans/24-ai-players.md) — боты имитируют активную галактику. Очень полезно при малом DAU («живой галактика»), но требует 2–3 недели только на MVP. Если запуск с 50+ реальными игроками — можно отложить.
- ~~[20 Ф.5](plans/20-legacy-port.md)~~ — ✅ **done 2026-04-25**: Stargate Jump (POST /api/stargate), cooldown по формуле legacy.
- [20 Ф.6](plans/20-legacy-port.md) — Moon destruction: аналогично, эндгейм.
- ~~[20 Ф.7](plans/20-legacy-port.md)~~ — ✅ **done 2026-04-25** ([ADR-0005](adr/0005-astrophysics.md)): Astrophysics с MAX(computer+1, astro/2+1) — не ломает существующих.
- ~~[20 Ф.8](plans/20-legacy-port.md)~~ — ✅ **done 2026-04-25**: IGR network — sum top (1+igr) labs, ускоряет research при множественных лабораториях.
- [22 Ф.2.2](plans/22-configs-cleanup.md) — orphan-юниты (decorative), не влияют на баланс.
- [15 Этап 4](plans/15-alien-holding-thursday.md) — UI для alien/holding: механика работает, UI позже.
- [14 Ф.4–Ф.7](plans/14-admin-expansion.md) — админ-дашборды / runbooks: полезны, но не блокер.
- [17 B–H](plans/17-gameplay-improvements.md) — daily quests, рейтинги, события: ретеншн-механики, итеративно.

---

## Дорожная карта

### Фаза 1 (неделя 1) — баланс и анти-эксплойт

✅ 18 Фаза 1 (rapidfire-порт, 2026-04-24) → ✅ симуляция ([результаты](balance/simulation-2026-04-24.md)) → ✅ 21 A ([ADR-0004](adr/0004-lancer-cost.md), Lancer cost) → 21 B (экспедиции) → 20 Ф.2 (fleet slots) → 20 Ф.1 (vacation) → 17 A1 (антибашинг). 18 Фаза 2 **не нужна** — DS endgame снят симуляцией.

### Фаза 2 (неделя 2) — PvP-инфраструктура и онбординг

20 Ф.3 (POSITION) → 20 Ф.4 (Phalanx) → 19 (wiki MVP)

### Фаза 3 (неделя 3) — операционная готовность

Бэкапы + мониторинг → 14 Ф.2.4–2.6 + 14 Ф.8.2 → второй платёжный шлюз → нагрузочный тест на VPS по [ops/vps-sizing.md](ops/vps-sizing.md)

### Запуск

### Фаза 4 (после запуска)

20 Ф.7 / Ф.8 / Ф.5 / Ф.6 → 24 боты → 17 B–H → 14 Ф.4+

---

## Предусловия для старта Фазы 1

1. **ADR на балансные отклонения от legacy**. План 18 Фаза 1 — чистый порт legacy
   и ADR **не требует** (CLAUDE.md: «берутся один-в-один из oxsar2 и oxsar2-java»).
   ADR нужен только для 18 Фазы 2 (отклонения от legacy, например shield DS 50k→30k)
   и 21 Блок A (Lancer cost) — если после симуляции они окажутся нужны.
2. **Решить по плану 24 (боты)** — если «живая галактика» на старте,
   Фазу 2 надо заменить на MVP ботов (2–3 недели), и запуск сдвигается
   ещё на 2 недели.
3. **LICENSE / CLA / privacy policy** — в README упомянуто, но для
   прод-запуска нужна публичная политика данных.
