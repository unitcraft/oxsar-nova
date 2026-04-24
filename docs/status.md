# Матрица готовности модулей

Легенда:
- ✅ — реализовано в соответствии с ТЗ.
- 🟡 — каркас/частично, есть TODO.
- ⬜ — только заявлено в ТЗ, не начато.

Milestones из §16 [oxsar-spec.txt](oxsar-spec.txt).

| Модуль                      | Статус | M  | Комментарий                                              |
|-----------------------------|:------:|:--:|----------------------------------------------------------|
| Монорепо, CI, Makefile      | ✅     | M0 | GitHub Actions: lint, test, build.                       |
| Docker-compose (pg + redis) | ✅     | M0 | `make dev-up`.                                           |
| goose-миграции (v0)         | ✅     | M0 | Базовые таблицы (users, planets, events, …).             |
| Go-сервер (chi, slog, ctx)  | ✅     | M0 | Health, graceful shutdown, middleware.                   |
| Auth (register/login/JWT)   | ✅     | M0 | argon2id + JWT + старт-планета при регистрации.          |
| Config loader (ENV + YAML)  | ✅     | M0 | Справочники + research + requirements + rapidfire.       |
| Economy tick (ресурсы)      | ✅     | M1 | DSL-путь (legacy-формулы из construction.yml, закоммичен). Fallback на приближения удалён. |
| Starter planet              | ✅     | M1 | +buildings[metal_mine,silicon_lab,solar_plant]=1 чтобы энергия не уходила в минус с первого тика. |
| Requirements checker        | ✅     | M2 | Один пакет, используется в building/research/shipyard.   |
| Building queue              | ✅     | M1 | Один слот + requirements-проверка + отмена с refund.     |
| Research queue              | ✅     | M2 | Одно исследование на игрока, ресурсы с планеты.          |
| Shipyard queue              | ✅     | M2 | Корабли и оборона, per-unit time × count.                |
| Defense                     | ✅     | M2 | В рамках shipyard — общая очередь.                       |
| Galaxy                      | ✅     | M3 | Repository + GET /api/galaxy/{g}/{s} + GalaxyScreen.      |
| Fleet / missions            | ✅     | M5 | TRANSPORT + ATTACK + RECYCLING + SPY + COLONIZE (kind=7,8,9,10,11). |
| Battle engine               | ✅     | M4.4b | rapidfire ✅, RECYCLING ✅, debris ✅, moon creation ✅, ACS multi-side (KindAttackAlliance=12) ✅. |
| Expedition                  | ✅     | M5 | 5 исходов (resources/artefact/pirates/loss/nothing), PvE-бой, отчёт + сообщение. |
| ExpedPlanetCreator          | ✅     | M5 | extra_planet outcome (5%): создаёт планету на случайном свободном слоте, проверяет лимит. |
| Rockets / stargate          | ✅     | M5 | Interplanetary rockets (kind=16): Launch → Impact. Stargate — позже. |
| Artefact                    | ✅     | M5.0 | Apply/Revert/Resync + delay activation (kind=63). Deactivate bugfix. |
| Artefact Market (credit)    | ✅     | M5.1 | List/Sell/Buy/Cancel + UI. users.credit как валюта. |
| Alien AI                    | ✅     | M5.2 | spawn раз в 6ч, 3 тира, KindAlienAttack=35, лут 30%, GRAB_CREDIT (0.08-0.1%), GIFT_CREDIT (5-10%, max 500), artefact drop 20%, координаты полёта ✅. Нет: HALT state machine (L). |
| Repair Factory              | ✅     | M4.4c | DISASSEMBLE + REPAIR end-to-end (API + worker + UI). damaged-юниты из боя чинятся целой пачкой. Defense repair ✅ (migration 0032, damaged_count+shell_percent в defense). |
| AutoMsg                     | ✅     | M4.2 | WELCOME/STARTER_GUIDE при регистрации + INACTIVITY_REMINDER (ежедневный воркер, last_seen_at). |
| Alliance / chat / message   | ✅     | M6 | messages: inbox + mark-read + compose + delete ✅. Alliance MVP ✅ (create/join/leave/disband, relations NAP/WAR/ALLY с mutual acknowledge, кастомные ранги). Chat WebSocket ✅. |
| Market (exchange)           | ✅     | M6 | Быстрый обменник (M↔Si↔H по курсам) ✅ + ордерная книга (CreateLot/ListLots/CancelLot/AcceptLot, migration 0022) ✅. |
| Officers                    | ✅     | M7 | 4 officer (ADMIRAL/GEOLOGIST/ENGINEER/MERCHANT), Activate→Expire через event kind=62, factor-поля. Group exclusivity: ADMIRAL+ENGINEER взаимоисключают (group_key 'build', migration 0033). |
| Achievements                | ✅     | M7 | MVP: 5 достижений (FIRST_METAL/SILICON/ARTEFACT/WIN/COLONY), lazy-check. |
| Tutorial                    | ✅     | M7 | 6 шагов (mine→solar→lab→computer_tech→ship→expedition), +10 кредитов за шаг, lazy-check. |
| Simulator UI                | ✅     | M7.1 | BattleSimScreen с реальными боевыми характеристиками из каталога + таблица потерь + multi-run (num_sim 2–20). |
| Admin panel                 | ✅     | M8 | GET /admin/stats + users list/ban/unban/credit/role. AutoMsg CMS (GET/PUT /admin/automsgs). AdminOnly middleware. |
| Payment integrations        | ⬜     | M9 | v2: WebMoney/Robokassa/A1/2Pay/VK/OK/MailRu.             |
| Event-loop worker           | ✅     | M3 | +Transport (kind=7 arrive) +Return (kind=20).            |
| Frontend каркас             | ✅     | M0 | Vite + TS strict + TanStack Query + Zustand.             |
| Score / Highscore           | ✅     | M5+ | RecalcUser/RecalcAll (формулы PointRenewer), GET /api/highscore, ScoreScreen. |
| Frontend экраны             | ✅     | M3 | 19 вкладок: overview, buildings, research, shipyard, repair, galaxy, fleet, market, rockets, artefacts, art-market, officers, tutorial, achievements, score, messages, alliance, chat, sim. |
| Порт дизайна oxsar2         | 🟡     | —  | CSS-темы перенесены как placeholder, ассеты копируются. Полный UI-редизайн — в [docs/ui/design-spec.md](ui/design-spec.md). |
| Порт .tpl-шаблонов          | ⬜     | —  | 134 .tpl (~12600 строк) — эталон разметки. Заменены React-экранами (19 шт.), прямой порт не планируется. |
| import-datasheets CLI       | ✅     | M0.1 | construction/ships/requirements/artefacts → YAML.        |
| import-phrases CLI          | ✅     | M0.2 | Запущен: configs/i18n/{ru,en}.yml, 23 группы, 1463 ключа. |
| pkg/formula (DSL eval)      | ✅     | M0.3 | parser+eval, 12+ legacy-формул проходят тесты.           |
| pkg/sqldump                 | ✅     | M0.1 | Общий парсер phpMyAdmin-дампов для обоих CLI.            |
| i18n runtime (backend)      | ✅     | M0.2 | Bundle + Tr + fallback + /api/i18n/{lang}.               |
| i18n runtime (frontend)     | ✅     | M0.2 | I18nProvider + useTranslation (t/tf); все экраны переведены с inline-дефолтами. |
