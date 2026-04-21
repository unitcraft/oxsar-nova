# Матрица готовности модулей

Легенда:
- ✅ — реализовано в соответствии с ТЗ.
- 🟡 — каркас/частично, есть TODO.
- ⬜ — только заявлено в ТЗ, не начато.

Milestones из §16 [oxsar-spec.txt](../oxsar-spec.txt).

| Модуль                      | Статус | M  | Комментарий                                              |
|-----------------------------|:------:|:--:|----------------------------------------------------------|
| Монорепо, CI, Makefile      | ✅     | M0 | GitHub Actions: lint, test, build.                       |
| Docker-compose (pg + redis) | ✅     | M0 | `make dev-up`.                                           |
| goose-миграции (v0)         | ✅     | M0 | Базовые таблицы (users, planets, events, …).             |
| Go-сервер (chi, slog, ctx)  | ✅     | M0 | Health, graceful shutdown, middleware.                   |
| Auth (register/login/JWT)   | ✅     | M0 | argon2id + JWT + старт-планета при регистрации.          |
| Config loader (ENV + YAML)  | ✅     | M0 | Справочники + research + requirements + rapidfire.       |
| Economy tick (ресурсы)      | ✅     | M1 | DSL-путь (legacy-формулы из construction.yml) + fallback на приближения. |
| Starter planet              | ✅     | M1 | +buildings[metal_mine,silicon_lab,solar_plant]=1 чтобы энергия не уходила в минус с первого тика. |
| Requirements checker        | ✅     | M2 | Один пакет, используется в building/research/shipyard.   |
| Building queue              | ✅     | M1 | Один слот + requirements-проверка + отмена с refund.     |
| Research queue              | ✅     | M2 | Одно исследование на игрока, ресурсы с планеты.          |
| Shipyard queue              | ✅     | M2 | Корабли и оборона, per-unit time × count.                |
| Defense                     | ✅     | M2 | В рамках shipyard — общая очередь.                       |
| Galaxy                      | ✅     | M3 | Repository + GET /api/galaxy/{g}/{s} + GalaxyScreen.      |
| Fleet / missions            | ✅     | M5 | TRANSPORT + ATTACK + RECYCLING + SPY + COLONIZE (kind=7,8,9,10,11). |
| Battle engine               | 🟡     | M4.4b | +UI отчёта боя (MessagesScreen → BattleReportView). Нет: compose/delete messages, rapidfire из каталога, RECYCLING. |
| Expedition                  | 🟡     | M5 | 5 исходов (resources / artefact / pirates / loss / nothing) по seed от fleet_id. PvE-бой с 5 light_fighter. |
| ExpedPlanetCreator          | ⬜     | M5 | ex ext/ExpedPlanetCreator.class.php.                     |
| Rockets / stargate          | 🟡     | M5 | Interplanetary rockets (kind=16): Launch → Impact. Stargate — позже. |
| Artefact                    | 🟡     | M5.0 | Apply/Revert/Resync для factor-эффектов (6 артефактов). |
| Artefact Market (credit)    | 🟡     | M5.1 | List/Sell/Buy/Cancel + UI. users.credit как валюта. |
| Alien AI                    | ⬜     | M5.2 | ex game/AlienAI.class.php 1127 LOC.                    |
| Repair Factory              | ✅     | M4.4c | DISASSEMBLE + REPAIR end-to-end (API + worker + UI). damaged-юниты из боя чинятся целой пачкой. |
| AutoMsg                     | ⬜     | M4.2 | ex game/AutoMsg.class.php 1228 LOC.                    |
| Alliance / chat / message   | 🟡     | M6 | messages: inbox + mark-read + battle-report view (M4.4b). Compose/folders/alliance — M6. |
| Market (exchange)           | 🟡     | M6 | MVP: M↔Si↔H по фиксированным курсам (1:2:4) × users.exchange_rate. Order-book / офферы — позже. |
| Officers                    | ⬜     | M7 | Планируется.                                             |
| Achievements                | ⬜     | M7 | ex game/Achievements.class.php 1044 LOC.                 |
| Tutorial                    | ⬜     | M7 | ex ext/page/ExtTutorial + game-классы.                   |
| Simulator UI                | ⬜     | M7.1 | ex ext/page/ExtSimulator 749 LOC.                      |
| Admin panel                 | ⬜     | M8 | Планируется.                                             |
| Payment integrations        | ⬜     | M9 | v2: WebMoney/Robokassa/A1/2Pay/VK/OK/MailRu.             |
| Event-loop worker           | ✅     | M3 | +Transport (kind=7 arrive) +Return (kind=20).            |
| Frontend каркас             | ✅     | M0 | Vite + TS strict + TanStack Query + Zustand.             |
| Frontend экраны             | ✅     | M3 | +Rockets — теперь 13 экранов.                            |
| Порт дизайна oxsar2         | 🟡     | —  | CSS-темы перенесены как placeholder, ассеты копируются.  |
| Порт .tpl-шаблонов          | ⬜     | —  | 134 .tpl (~12600 строк) — эталон разметки UI, §8 ТЗ.     |
| import-datasheets CLI       | ✅     | M0.1 | construction/ships/requirements/artefacts → YAML.        |
| import-phrases CLI          | ✅     | M0.2 | Запущен: configs/i18n/{ru,en}.yml, 23 группы, 1463 ключа. |
| pkg/formula (DSL eval)      | ✅     | M0.3 | parser+eval, 12+ legacy-формул проходят тесты.           |
| pkg/sqldump                 | ✅     | M0.1 | Общий парсер phpMyAdmin-дампов для обоих CLI.            |
| i18n runtime (backend)      | ✅     | M0.2 | Bundle + Tr + fallback + /api/i18n/{lang}.               |
| i18n runtime (frontend)     | ✅     | M0.2 | I18nProvider + useTranslation (t/tf); все экраны переведены с inline-дефолтами. |
