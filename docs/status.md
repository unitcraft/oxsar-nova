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
| Battle engine               | 🟡     | M4.4b | rapidfire ✅, RECYCLING ✅, debris ✅. Нет: moon creation (kind=14), ACS multi-side. |
| Expedition                  | ✅     | M5 | 5 исходов (resources/artefact/pirates/loss/nothing), PvE-бой, отчёт + сообщение. |
| ExpedPlanetCreator          | ✅     | M5 | extra_planet outcome (5%): создаёт планету на случайном свободном слоте, проверяет лимит. |
| Rockets / stargate          | ✅     | M5 | Interplanetary rockets (kind=16): Launch → Impact. Stargate — позже. |
| Artefact                    | ✅     | M5.0 | Apply/Revert/Resync + delay activation (kind=63). Deactivate bugfix. |
| Artefact Market (credit)    | ✅     | M5.1 | List/Sell/Buy/Cancel + UI. users.credit как валюта. |
| Alien AI                    | ⬜     | M5.2 | ex game/AlienAI.class.php 1127 LOC.                    |
| Repair Factory              | ✅     | M4.4c | DISASSEMBLE + REPAIR end-to-end (API + worker + UI). damaged-юниты из боя чинятся целой пачкой. |
| AutoMsg                     | 🟡     | M4.2 | WELCOME/STARTER_GUIDE при регистрации, идемпотентно, {{var}}-шаблоны. Scheduled сообщения — позже. |
| Alliance / chat / message   | 🟡     | M6 | messages: inbox + mark-read + compose + delete ✅. Alliance MVP ✅ (create/join/leave/disband). Chat (WebSocket) — M6+. |
| Market (exchange)           | 🟡     | M6 | MVP: M↔Si↔H по фиксированным курсам (1:2:4) × users.exchange_rate. Order-book / офферы — позже. |
| Officers                    | ✅     | M7 | 4 officer (ADMIRAL/GEOLOGIST/ENGINEER/MERCHANT), Activate→Expire через event kind=62, factor-поля. |
| Achievements                | 🟡     | M7 | MVP: 5 достижений (FIRST_METAL/SILICON/ARTEFACT/WIN/COLONY), lazy-check при GET. Идемпотентно. |
| Tutorial                    | ⬜     | M7 | ex ext/page/ExtTutorial + game-классы.                   |
| Simulator UI                | ✅     | M7.1 | BattleSimScreen с реальными боевыми характеристиками из каталога + таблица потерь. |
| Admin panel                 | ⬜     | M8 | Планируется.                                             |
| Payment integrations        | ⬜     | M9 | v2: WebMoney/Robokassa/A1/2Pay/VK/OK/MailRu.             |
| Event-loop worker           | ✅     | M3 | +Transport (kind=7 arrive) +Return (kind=20).            |
| Frontend каркас             | ✅     | M0 | Vite + TS strict + TanStack Query + Zustand.             |
| Score / Highscore           | ✅     | M5+ | RecalcUser/RecalcAll (формулы PointRenewer), GET /api/highscore, ScoreScreen. |
| Frontend экраны             | ✅     | M3 | +Score — теперь 14 экранов.                              |
| Порт дизайна oxsar2         | 🟡     | —  | CSS-темы перенесены как placeholder, ассеты копируются.  |
| Порт .tpl-шаблонов          | ⬜     | —  | 134 .tpl (~12600 строк) — эталон разметки UI, §8 ТЗ.     |
| import-datasheets CLI       | ✅     | M0.1 | construction/ships/requirements/artefacts → YAML.        |
| import-phrases CLI          | ✅     | M0.2 | Запущен: configs/i18n/{ru,en}.yml, 23 группы, 1463 ключа. |
| pkg/formula (DSL eval)      | ✅     | M0.3 | parser+eval, 12+ legacy-формул проходят тесты.           |
| pkg/sqldump                 | ✅     | M0.1 | Общий парсер phpMyAdmin-дампов для обоих CLI.            |
| i18n runtime (backend)      | ✅     | M0.2 | Bundle + Tr + fallback + /api/i18n/{lang}.               |
| i18n runtime (frontend)     | ✅     | M0.2 | I18nProvider + useTranslation (t/tf); все экраны переведены с inline-дефолтами. |
