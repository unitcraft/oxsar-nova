# Матрица UI-тестов

Сгенерирована вручную на базе плана [13-ui-testing.md](plans/13-ui-testing.md)
и содержимого [frontend/e2e/](../frontend/e2e/). Обновлять при добавлении
новых spec-файлов.

**Текущий статус (2026-04-24): 110/110 passed, 0 failed** — полный прогон
`make test-e2e-docker` зелёный (~7 минут). См. commit `<following commit>`.

Легенда:
- ✅ — спек написан, автопроверка в CI
- 🟡 — поверхностная smoke-проверка, глубокий сценарий отложен
- ⬜ — нет спека

## Ф.1 Критический путь

| Сценарий | Файл | Статус |
|----------|------|--------|
| Auth: register | critical/auth.spec.ts | ⬜ (register form есть, spec отложен — hash/ratelimit) |
| Auth: login + wrong password | critical/auth.spec.ts | ✅ |
| Auth: logout | critical/auth.spec.ts | ✅ |
| Auth: refresh rotation | smoke.spec.ts (косвенно, через loginAs) | 🟡 |
| Buildings: список + уровни | critical/planet-core.spec.ts | ✅ |
| Buildings: старт/cancel очереди | — | ⬜ |
| Research: список | critical/planet-core.spec.ts | ✅ |
| Shipyard: список | critical/planet-core.spec.ts | ✅ |
| Galaxy: координаты, метки | critical/space.spec.ts | ✅ |
| Fleet: открытие + empty state | critical/space.spec.ts | ✅ |
| Fleet: полный Transport/Attack/Spy | — | ⬜ (нужны точные селекторы форм) |
| Messages: inbox + welcome | critical/messages.spec.ts | ✅ |
| Messages: unread badge | critical/messages.spec.ts | 🟡 |
| Messages: compose / delete | — | ⬜ |
| Overview: ресурсы + планета | critical/planet-core.spec.ts | ✅ |

## Ф.2 Основная функциональность

| Сценарий | Файл | Статус |
|----------|------|--------|
| Repair | domains/screens.spec.ts | ✅ smoke |
| Market (exchange + lots) | domains/screens.spec.ts | ✅ smoke |
| Rockets | domains/screens.spec.ts | ✅ smoke |
| Artefacts | domains/screens.spec.ts | ✅ smoke |
| Art-market | domains/screens.spec.ts | ✅ smoke |
| Officers | domains/screens.spec.ts | ✅ smoke |
| Alliance (member view) | domains/screens.spec.ts | ✅ smoke |
| Alliance: create/invite/leave | — | ⬜ |
| Chat | domains/screens.spec.ts | ✅ smoke |
| Score | domains/screens.spec.ts | ✅ smoke |
| Achievements | domains/screens.spec.ts | ✅ smoke |
| Tutorial / Profession | domains/screens.spec.ts | ✅ smoke |
| Battle Sim | domains/screens.spec.ts | ✅ smoke |
| **Credits (mock payment)** | domains/credits.spec.ts | ✅ **полный** (happy-path + fail) |

## Ф.3 Второстепенные

| Экран | Файл | Статус |
|-------|------|--------|
| Empire | domains/secondary.spec.ts | ✅ smoke |
| Techtree | domains/secondary.spec.ts | ✅ smoke |
| Battlestats | domains/secondary.spec.ts | ✅ smoke |
| Records | domains/secondary.spec.ts | ✅ smoke |
| Notepad | domains/secondary.spec.ts | ✅ smoke |
| Referral | domains/secondary.spec.ts | ✅ smoke |
| Friends | domains/secondary.spec.ts | ✅ smoke |
| Settings | domains/secondary.spec.ts | ✅ smoke |
| Planet-options | domains/secondary.spec.ts | ✅ smoke |
| Resource | domains/secondary.spec.ts | ✅ smoke |
| Global Search (Ctrl+K) | domains/secondary.spec.ts | ✅ |
| Admin (superadmin) | domains/secondary.spec.ts | ✅ |
| Admin hidden for player | domains/secondary.spec.ts | ✅ |

## Ф.4 Граничные случаи

| Сценарий | Файл | Статус |
|----------|------|--------|
| 500 from server | edges/resilience.spec.ts | ✅ |
| 401 mid-session → logout | edges/resilience.spec.ts | ✅ |
| Empty states (alice) | edges/resilience.spec.ts | ✅ |
| i18n: нет голых ключей | edges/i18n-mobile.spec.ts | ✅ |
| i18n: переключение ru↔en | — | ⬜ |
| Mobile: bottom nav + more sheet | edges/i18n-mobile.spec.ts | ✅ |
| Layout integrity (Ф.4.6) | helpers/layout.ts → smoke.spec.ts | ✅ на всех 30 вкладках |

## Как дополнять

1. Глубокие сценарии (полный флоу Attack, alliance invite) — отдельным
   spec-файлом, с подготовкой БД через `testseed` или API-вызовами
   из фикстуры.
2. Новые экраны — добавлять в `SECONDARY_TABS` в `domains/secondary.spec.ts`,
   специфику — в отдельный spec.
3. Любое падение smoke → либо фикс UI, либо карантин (`.skip` с тикетом).

## Оффлайн-режим

В плане стоит Ф.4.1 offline — не реализован в первом проходе: TanStack
Query показывает stale-data, но тест требует отключения network через
CDP (`context.setOffline(true)`) и проверки, что Vite HMR websocket
не ломает тест. Оставлено как follow-up.
