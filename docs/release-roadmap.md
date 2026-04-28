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
| ~~[21 C](plans/21-gameplay-hardening.md)~~ | ~~Щиты: golden-тест high-tech~~ | ✅ **done 2026-04-25** — BA-005 подтверждён и закрыт: дефект портирования Java applyShots исправлен, ignoreAttack=0 для планетарных щитов (id 49/50), golden-тест проходит. | — |
| ~~[20 Ф.1](plans/20-legacy-port.md)~~ | ~~Vacation mode~~ | ✅ **done 2026-04-24** — min 48h, blocking events, target/sender shields, prod=0, /api/me fields. Auto-disable через 30d — остаток (не блокер). | — |
| ~~[20 Ф.2](plans/20-legacy-port.md)~~ | ~~Fleet slots через computer_tech~~ | ✅ **done 2026-04-24** — `1 + floor(computer_tech/6)`, UI-индикатор, 409 при overflow. | — |
| ~~[17 A1](plans/17-gameplay-improvements.md)~~ | ~~Антибашинг~~ | ✅ **done 2026-04-24** — max 4 атаки (ATTACK_SINGLE/ACS) от attacker→defender за 5h, ErrBashingLimit → 409. | — |
| ~~ops~~ | ~~Резервное копирование БД + мониторинг~~ | ✅ **done 2026-04-25** — deploy/backup.sh (pg_dump+rotate+S3), docker-compose.monitoring.yml (Prometheus+Grafana+pg-exporter), /metrics на сервере и воркере, runbook. | — |

**Вместе:** ~1–1.5 недели работы, из которых 70% — планы 21+18.

---

## 🟡 Важные — желательно до массового анонса

| # | Что | Почему | Объём |
|---|---|---|---|
| ~~[20 Ф.3](plans/20-legacy-port.md)~~ | ~~Миссия POSITION~~ | ✅ **done 2026-04-25** — mission=6, ALLY/NAP target validation, PositionArriveHandler. | — |
| ~~[20 Ф.4](plans/20-legacy-port.md)~~ | ~~Сенсорная Фаланга~~ | ✅ **done 2026-04-25** — GET /api/phalanx, формула radius из legacy, 5000H за скан. | — |
| ~~[19](plans/19-game-wiki.md)~~ | ~~Вики игры~~ | ✅ **Полный MVP done 2026-04-25** — backend API + генератор + frontend UI с собственным md-renderer (без react-markdown зависимости). | — |
| ~~[14 Ф.2.4–2.6](plans/14-admin-expansion.md)~~ | ~~Force-recall fleet, soft-delete user~~ | ✅ **done 2026-04-25** — admin endpoints для флота/планет/user-delete с auditable. | — |
| ~~[14 Ф.8.2](plans/14-admin-expansion.md)~~ | ~~Rate-limit админ-действий~~ | ✅ **done 2026-04-25** — 100 write/hour на админа, 429 + Retry-After. | — |
| ~~payment~~ | ~~Второй шлюз (ЮKassa или Enot.io)~~ | ✅ **done 2026-04-25** — Enot.io gateway + 5 unit-тестов. PAYMENT_PROVIDER=enot переключает. | — |

---

## 🟢 Nice-to-have — после первой волны игроков

- [24 ai-players](plans/24-ai-players.md) — боты имитируют активную галактику. Очень полезно при малом DAU («живой галактика»), но требует 2–3 недели только на MVP. Если запуск с 50+ реальными игроками — можно отложить.
- ~~[20 Ф.5](plans/20-legacy-port.md)~~ — ✅ **done 2026-04-25**: Stargate Jump (POST /api/stargate), cooldown по формуле legacy.
- ~~[20 Ф.6](plans/20-legacy-port.md)~~ — ✅ **done 2026-04-25**: Moon Destruction (kind=25/27) с OGame-формулой rip-roll, single + ACS.
- ~~[20 Ф.7](plans/20-legacy-port.md)~~ — ✅ **done 2026-04-25** ([ADR-0005](adr/0005-astrophysics.md)): Astrophysics с MAX(computer+1, astro/2+1) — не ломает существующих.
- ~~[20 Ф.8](plans/20-legacy-port.md)~~ — ✅ **done 2026-04-25**: IGR network — sum top (1+igr) labs, ускоряет research при множественных лабораториях.
- ~~[22 Ф.2.2](plans/22-configs-cleanup.md)~~ — ✅ **решено 2026-04-25** ([ADR-0006](adr/0006-orphan-units-deferred.md)): orphan-юниты НЕ реализуем в v1, knownOrphans = roadmap для v1.x.
- ~~[15 Этап 4](plans/15-alien-holding-thursday.md)~~ — ✅ **MVP done 2026-04-25**: панель захваченных планет в Overview, GET /api/alien/holdings/me + платёж через готовый endpoint.
- [14 Ф.4–Ф.7](plans/14-admin-expansion.md) — админ-дашборды / runbooks: полезны, но не блокер.
- ~~[17 D](plans/17-gameplay-improvements.md)~~ — ✅ **done 2026-04-25**: Daily Quests с lazy-gen, 9 типов заданий, claim API, frontend tab.
- ~~17 E~~ — ✅ уже реализовано (категориальные рейтинги в ScoreScreen существуют с M5+).
- ~~17 F~~ — ✅ **MVP done 2026-04-25**: meteor_storm с +30% metal через admin-API, баннер на Overview.
- ~~17 G1~~ — ✅ **done 2026-04-25**: forecast ресурсов в Overview.
- ~~17 G2/G3/G4~~ — отложены v1.x (UX мелочи).
- ~~17 H~~ — отложены v1.x (точность боя, не блокер).
- ~~17 B2, C~~ — отложены v1.x (B1 закрыт, B2/C — flavor).

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

## Пост-запуск v2: Портал мира Oxsar

Реализуется после стабилизации v1. Подробный план — [36-portal-multiverse.md](plans/36-portal-multiverse.md).

| Фаза | Что | Оценка |
|---|---|---|
| Ф.1 | Выделение Auth Service (RSA-256, отдельный бинарник) | 3–4 дня |
| Ф.2 | Universe Registry — конфиг и список вселенных | 1–2 дня |
| Ф.3 | Portal Frontend + Backend (новости, список вселенных) | 3–4 дня |
| Ф.4 | Лента предложений с голосованием кредитами | 4–5 дней |
| Ф.5 | OAuth — вход через ВКонтакте, Mail.ru, Яндекс | 3–4 дня |
| Ф.6 | Global credits, перенос платёжных шлюзов в Auth Service | 2–3 дня |

**Итого**: ~16–22 рабочих дня. Ф.1–Ф.6 можно начать параллельно с первой волной игроков v1.

**Предусловия для Ф.5** (OAuth): публичный домен + регистрация приложений в ВКонтакте / Mail.ru / Яндекс.

---

## Пост-запуск v3: Ремастер origin на nova-backend

Реализуется после стабилизации v2. Подробное обоснование и
декомпозиция — [docs/research/origin-vs-nova/roadmap-report.md](research/origin-vs-nova/roadmap-report.md).

Стратегия зафиксирована планом 62: pixel-perfect клон UI origin на
React/TipTap, backend = game-nova с override-схемой для вселенной
**origin** (ADR-0010 Accepted: `origin.oxsar-nova.ru`,
`universes.code = 'origin'`). 55 экранов origin → S-NNN, 46
расхождений → D-NNN, 15+22 UI-функций и UX-микрологики → U-NNN/X-NNN.

15 правил R0-R15 в roadmap-report «Часть I.5» — обязательны для
всей серии (R0 геймплей nova заморожен, R15 без упрощений как
для прода, R6 REST API с нуля и т.д.).

| План | Что | Оценка | Зависит от |
|---|---|---|---|
| 64 | origin.yaml override + per-universe balance loading | 2 нед | — |
| 65 | Расширение event-loop (7 событий: TELEPORT_PLANET, DESTROY_BUILDING ×2, DELIVERY_ARTEFACTS, EXCHANGE ×2, ALLIANCE_ATTACK_ADDITIONAL) | 3-4 нед | 64 |
| 66 | AlienAI полный паритет во всех вселенных + спец-юниты (R0-исключение) | 3 нед | 64 |
| 67 | Расширение alliance: 3 описания, передача лидерства, гранулярные ранги, лог, расширенные дипстатусы (5 enum), buddy-list | 2-3 нед | — |
| 68 | Биржа артефактов (player-to-player, с premium) | 3-4 нед | — |
| 69 | Расширение domain-полей в users (без race и ui_theme — отказ; +notes для notepad) | 2 нед | — |
| 70 | Achievements расширение | ⏸ **Отложен** до пост-запуска | — |
| 71 | UX-микрологика origin → nova-frontend (X-NNN) | 2-3 нед | — |
| **72** | **Origin-фронт — pixel-perfect клон (50 экранов; без Achievements/Tutorial/баннеров)** | **12-16 нед** | **64-69, 71 + 57** |
| 73 | Screenshot-diff CI (Playwright + visual regression) | 2 нед | 72 |
| 74 | origin deploy + DNS + config | 1 нед | 72, 73 |
| 76 | nova-frontend UI для биржи (uni01/uni02) | 1-2 нед | 68 |

**Итого**: 6-9 месяцев (минимальный путь без отложенного 70).

**Предусловия для плана 72** (origin-фронт): план 57 (mail-service /
TipTap), аудит лицензий шрифтов и иконок origin (план 72 Ф.1),
план 75 закрыт (`projects/game-origin/` свободна — ✅).

**Что реализуем cross-universe** (получают и uni01/uni02):
- 65 — TELEPORT_PLANET, DESTROY_BUILDING (общие game-механики).
- 66 — AlienAI + спец-юниты (R0-исключение по решению пользователя).
- 67 — alliance-расширения (диплома́тия, ранги, buddy-list).
- 68 — биржа артефактов (через план 76 для UI nova).
- 71 — UX-микрологика.

### Breaking changes для modern-вселенных перед запуском origin

Перед запуском origin (план 74) — release notes для игроков
uni01/uni02 о следующих изменениях, появляющихся в их вселенных:

- **AlienAI становится активной**: по четвергам в 19:00
  инопланетные флоты могут прилетать на планеты игроков
  (грабёж оксаритов, удержание планеты, подарки). Раньше
  AlienAI в nova была в упрощённом виде (план 15 этапы 1-2).
- **Спец-юниты доступны для постройки**: после посещения
  AlienAI-флотом своей планеты игрок открывает доступ к
  Lancer / Shadow / Transplantator / Collector / Planet Shield /
  Armored Terran (как в legacy-PHP origin).
- **Биржа артефактов**: новый экран в навигации, можно покупать
  и продавать артефакты за оксариты с premium-маркером.
- **Расширенная alliance-дипломатия**: 5 статусов вместо 3
  (`hostile_neutral` и `nap` добавлены).
- **Buddy-list**: новый экран друзей.
- **TELEPORT_PLANET**: премиум-фича через оксары.

Эти изменения — **намеренные upgrade'ы игрового опыта modern**,
не R0-нарушения (зафиксированы в roadmap-report «Часть I.5» как
явные исключения R0 по решению пользователя 2026-04-28).

---

## Предусловия для старта Фазы 1

1. **ADR на балансные отклонения от legacy**. План 18 Фаза 1 — чистый порт legacy
   и ADR **не требует** (CLAUDE.md: «берутся один-в-один из oxsar2 и oxsar2-java»).
   ADR нужен только для 18 Фазы 2 (отклонения от legacy, например shield DS 50k→30k)
   и 21 Блок A (Lancer cost) — если после симуляции они окажутся нужны.
2. **Решить по плану 24 (боты)** — если «живая галактика» на старте,
   Фазу 2 надо заменить на MVP ботов (2–3 недели), и запуск сдвигается
   ещё на 2 недели.
3. **Юридическая обвязка** — отдельный блок планов 39–47 (см. ниже
   раздел «Юридическая обвязка к публичному запуску»).

---

## Юридическая обвязка к публичному запуску

Сквозная цепочка планов 39–47, оформляющая правовой статус проекта,
лицензирование, обработку персональных данных, возрастную маркировку,
оферту и платежи. **До прохождения этого блока публичный запуск
запрещён** — без него регистрация пользователей и приём платежей не
имеют легальной основы по РФ-праву.

### Текущий статус (обновлён 2026-04-27)

**Юридическая обвязка:**

| План | Что | Статус |
|---|---|---|
| [39](plans/39-license-ru-jurisdiction.md) | Governing law (РФ) | ✅ выполнен |
| [40](plans/40-license-audit.md) | Аудит лицензий + CI license-check | ✅ выполнен |
| [41](plans/41-origin-rights.md) | Origin-rights, AI как инструмент | ✅ выполнен |
| [42](plans/42-yookassa.md) | YooKassa | ✅ выполнен |
| [43](plans/43-game-origin-composer.md) | Recipe → Composer | ✅ выполнен |
| [44](plans/44-personal-data-152fz.md) | 152-ФЗ — ПДн | ✅ выполнен (Ф.4 — внерепно) |
| [45](plans/45-trademark.md) | Товарный знак «oxsar» | 📝 ручные шаги пользователя |
| [46](plans/46-age-rating-ugc.md) | 436-ФЗ + UGC | ✅ выполнен |
| [47](plans/47-offer-tos.md) | Оферта + правила + refund | ✅ выполнен |
| [49](plans/49-doc-hygiene-pii.md) | Гигиена ПДн в документации | ✅ выполнен |
| [50](plans/50-game-origin-legal-fix.md) | game-origin gaps | 🟡 5/7 фаз; открыты Ф.2, Ф.5 |

**Архитектура и инфраструктура:**

| План | Что | Статус |
|---|---|---|
| [51](plans/51-rename-auth-to-identity.md) | auth → identity rename | ✅ выполнен |
| [52](plans/52-rbac-unification.md) | RBAC unification | ✅ выполнен |
| [53](plans/53-admin-frontend.md) | Admin-frontend + admin-bff | 🚧 в работе (Ф.1-6, 9, 11, 12 закрыты) |
| [54](plans/54-billing-limits-reports.md) | Billing-лимиты + admin-отчёты | ⏸️ приостановлен (для согласования с 58) |
| [55](plans/55-doc-sync-after-identity-rename.md) | Sync документации после rename | ✅ выполнен |
| [56](plans/56-reports-to-portal.md) | Reports → portal | 🟡 backend закрыт; открыты Ф.4, Ф.6, Ф.7, Ф.8 |

**Эпики и крупные планы (не запускаются сейчас):**

| План | Что | Статус |
|---|---|---|
| [48](plans/48-moderation-service.md) | Moderation-service | 📝 черновик (запуск по триггеру) |
| [57](plans/57-mail-service.md) | Mail-service на TipTap | 📝 эпик (после публичного запуска) |
| [58](plans/58-currency-rebranding.md) | Кредиты → Оксары + Оксариты | 📝 не запускался |
| [59](plans/59-referral-program.md) | Реферальная программа | 📝 не запускался (после 58) |
| [60](plans/60-remove-referral-from-game-origin.md) | Удалить legacy реферальную систему из game-origin | 🚧 в работе |
| [61](plans/61-admin-bff-rewrite-migration.md) | admin-bff: Director → Rewrite (deprecated API + security) | 📝 не срочно |

**Ручные шаги пользователя (внерепно):**

| Шаг | Когда |
|---|---|
| 45 Ф.1 — поиск по реестру Роспатента | срочно (10 мин) |
| 45 Ф.2 — подача заявки на товарный знак | до публичного анонса |
| 44 Ф.4 — уведомление РКН | до открытия публичной регистрации |
| YooKassa: реальный shopId + «Мой налог» | за неделю до запуска |

### Оптимальный порядок выполнения (актуальный)

Финальная схема с учётом всех планов 39-59 + ADR-0009 (валюта Оксары/Оксариты):

```
ФАЗА A. ЗАВЕРШЕНИЕ ОТКРЫТЫХ ПЛАНОВ          ВНЕРЕПНО (параллельно)

[1] 56 Ф.4-Ф.8 — frontend reports          [P1] 45 Ф.1 — поиск ФИПС (10 мин)
    + перенос админки в admin-frontend          ↓
        ↓                                  [P2] 45 Ф.2 — подача заявки
[2] 50 Ф.2 — заглушка прямой регистрации        в Роспатент (1 день)
    в game-origin (задание готово)              срок защиты — с даты подачи
        ↓
[3] 50 Ф.5 — кнопка «Пожаловаться»          [P3] 44 Ф.4 — РКН (30 мин)
    в game-origin (после 56)                    можно сейчас, если 44 готов
        ↓
[4] 53 — оставшиеся фазы admin-frontend
    (соседний агент продолжает)

ФАЗА B. ВАЛЮТНАЯ МОДЕЛЬ + АДМИНИСТРАТИВНЫЕ
═══════════════════════════════════════════

[5] 58 — кредиты → Оксары + Оксариты        После 58 — refresh плана 53/54:
    (5-7 коммитов, 8-14 часов)              - 53: rename admin /credit
        ↓                                     /api endpoints на /oxsar
[6] 54 — billing-limits + admin-reports     - 54: завершить с правильной
    (продолжается после паузы — теперь        терминологией
    с правильной терминологией)
        ↓
[7] 59 — реферальная программа              [P4] YooKassa shopId + Мой налог
    (после 58 — оксариты должны быть!)          (за неделю до запуска)
        ↓
ФАЗА C. ПРЕДЗАПУСКОВАЯ ПРОВЕРКА
═══════════════════════════════════════════

[8] Юр-аудит ещё раз через
    docs/prompts/legal-compliance-audit.md
        ↓
[9] Маркетинг-материалы для ребрендинга
    (банер «Кредиты переименованы в Оксары»,
    новость, FAQ)
        ↓
🚀 ПУБЛИЧНЫЙ ЗАПУСК

ФАЗА D. ПОСТЗАПУСК (без приоритета)
═══════════════════════════════════════════

[10] 57 — mail-service (эпик 2-3 недели)
[11] 48 — moderation-service (по триггеру)
[12] 45 Ф.3-Ф.4 — TM-номер в README
     (когда Роспатент выдаст свидетельство,
     6-18 мес после P2)
```

### Зависимости между планами

```
Ф.5 plan-50 ──depends-on──→ план 56
plan-58 ─────depends-on──→ план 38, 42 (✅) + плановое окно деплоя
plan-54 ─────depends-on──→ план 51, 52 (✅), план 53, **план 58** (терминология)
plan-59 ─────depends-on──→ план 58 (оксариты должны существовать)
plan-57 ─────depends-on──→ публичный запуск + переписывание game-origin на Go+React
plan-48 ─────depends-on──→ триггер из plan-46 (рост спама/жалоб)
```

### Пошаговая последовательность

#### Прямо сейчас можно делать (параллельно):

1. **56 Ф.4-Ф.8** — frontend и финализация reports → portal.
   Backend уже в portal, осталось переключить ReportButton.tsx
   и перенести AdminReportsTab в admin-frontend через admin-bff.
2. **50 Ф.2** — заглушка прямой регистрации в `AccountCreator.class.php`
   (задание готово, ~30 минут).
3. **60** — удалить legacy реферальную систему из game-origin
   (~30-60 минут, не зависит ни от чего; полностью переделывается
   в плане 59 на новом стеке).
4. **45 Ф.1+Ф.2** — поиск по реестру Роспатента + подача заявки
   (внерепные шаги пользователя).
5. **44 Ф.4** — уведомление РКН (внерепный шаг пользователя,
   30 минут на Госуслугах).

#### После того как 56 закрыт:

5. **50 Ф.5** — кнопка «Пожаловаться» в game-origin (Smarty + JS +
   POST на portal-mail-API из плана 56).

#### После того как все открытые планы 50, 56 закрыты:

6. **58** — ребрендинг валюты. Это maintenance window:
   миграция БД (rename `users.credit → users.oxsarit`, создание
   `wallets.oxsar`), backend (charge API + smart-pay), frontend
   (BalanceBadge с двумя валютами), i18n-корректировка во всех
   проектах. **Maintenance window 5-15 минут или blue-green
   деплой по плану 31.**

#### После 58:

7. **53 refresh** — переименовать `/api/admin/users/{id}/credit` →
   `/oxsar` в admin-frontend и admin-bff (если 53 ещё не закрыт
   соседом полностью).
8. **54** — продолжить billing-limits с правильной терминологией.
   План был приостановлен; после 58 — возобновить.
9. **59** — реферальная программа. Оксариты как награды (план 58
   их создал).

#### Перед публичным запуском:

10. **YooKassa shopId** — за неделю до запуска регистрация магазина,
    привязка к учётной системе оператора, реальные ENV-переменные.
11. **Юр-аудит ещё раз** — прогон промпта
    `docs/prompts/legal-compliance-audit.md`. Подтвердить, что после
    58/59 нет регрессий в обвязке.
12. **Маркетинг-материалы** — банер про ребрендинг, новость, FAQ.

#### Запуск и после:

13. 🚀 **публичный запуск**.
14. **57** — mail-service (эпик, 2-3 недели после запуска).
15. **48** — moderation-service (только по триггеру).
16. **45 Ф.3-Ф.4** — вписать TM-номер в README, когда Роспатент
    выдаст свидетельство (6-18 мес после P2).
