# Промпт: выполнить план 64 (origin.yaml override + per-universe balance loading)

**Дата создания**: 2026-04-28
**План**: [docs/plans/64-remaster-origin-yaml-override.md](../plans/64-remaster-origin-yaml-override.md)
**Применение**: вставить блок ниже в новую сессию Claude Code в
рабочей директории `d:\Sources\oxsar-nova`.
**Объём**: 3 коммита, ~800-1500 строк Go + ~3000-5000 строк YAML
автогенерации, ~2 недели агента.

**Это стартовый план серии ремастера 64-74.** Без него блокируются
65, 66, 69.

---

```
Задача: выполнить план 64 (ремастер) — origin.yaml override +
per-universe balance loading. Стартовый план серии ремастера origin
на nova-backend.

ВАЖНОЕ:
- Это исследовательский + код-задача, КОД НА Go ПИШЕМ.
- Стратегия серии 64-74 уже зафиксирована — НЕ переоценивать,
  выполнять по плану.
- Параллельно могут идти агенты по другим планам серии (67, 68, 71).
  Они не пересекаются по файлам.

ПЕРЕД НАЧАЛОМ:

1) git status --short — если есть чужие изменения от параллельных
   сессий, бери только свои файлы.

2) Прочитай ПОЛНОСТЬЮ:
   - docs/plans/64-remaster-origin-yaml-override.md (твоё ТЗ)
   - docs/research/origin-vs-nova/roadmap-report.md «Часть I.5» —
     ВСЕ R0-R15 правила (особенно R0 геймплей nova заморожен,
     R1 имена, R10 per-universe изоляция, R12 i18n переиспользование,
     R15 без упрощений как для прода)
   - CLAUDE.md (правила проекта)

3) Прочитай выборочно по мере необходимости:
   - docs/research/origin-vs-nova/divergence-log.md D-022, D-026..D-030
     (расхождения, которые этот план закрывает)
   - docs/research/origin-vs-nova/formula-dsl.md (DSL legacy-формул)
   - docs/legacy/game-origin-access.md (доступ к live origin на :8092
     для импорта)
   - projects/game-origin-php/migrations/001_schema.sql,
     002_data.sql (источник данных na_construction)
   - projects/game-origin-php/src/game/Functions.inc.php:41
     (parseChargeFormula) и projects/game-origin-php/src/game/
     Planet.class.php:592 (parseSpecialFormula) — DSL парсеры
   - projects/game-nova/configs/buildings.yml, units.yml — формат
     для override
   - projects/game-nova/backend/internal/economy/, building/,
     research/, shipyard/ — потребители balance

ЧТО НУЖНО СДЕЛАТЬ (по фазам плана 64):

Ф.1. Скаффолд loader без БД-миграций
- Создать internal/balance/loader.go с типами Bundle, Building, Unit и
  функцией LoadDefaults() (читает существующие configs/buildings.yml,
  units.yml, rapidfire.yml — modern profile).
- Никаких миграций БД (override-схема, не balance_profile column).
- Тест: LoadDefaults() возвращает ту же конфигурацию, что раньше
  читалась прямым доступом к YAML. Все existing nova-тесты остаются
  зелёными — это критерий приёма Ф.1.

Ф.2. Импорт-скрипт cmd/tools/import-legacy-balance/main.go
- Парсер DSL для статических формул: pow(base, exp), floor, round,
  min, max + арифметика + переменная level.
- Connect к docker-mysql-1 через database/sql + go-sql-driver/mysql.
- SELECT из na_construction, na_rapidfire, na_ship_datasheet,
  na_options. consts.php-define-константы — копировать руками в
  YAML globals: (PHP-define не парсится).
- Генерация:
  · configs/balance/origin.yaml — override чисел origin (basic_*,
    charge_table[], RF, и т.д.).
  · ДОПОЛНЕНИЕ к дефолтным configs/units.yml / ships.yml /
    rapidfire.yml — алиен-юниты (alien_unit_1..5) + спец-юниты
    (lancer_ship, shadow_ship, ship_transplantator, ship_collector,
    small_planet_shield, large_planet_shield, armored_terran).
    Это R0-исключение по решению пользователя 2026-04-28: AlienAI и
    спец-юниты работают во всех вселенных, не только origin.
- Verify: спот-проверка нескольких значений вручную против live origin.

Ф.3. Override-loader для origin
- Расширить internal/balance/loader.go функцией
  LoadFor(universeCode string):
  · Если есть configs/balance/<universeCode>.yaml → deep merge
    поверх дефолта.
  · Иначе → возвращает чистый дефолт.
- In-memory кеш per-universe.
- Тест: LoadFor("uni01") = дефолт; LoadFor("origin") = bundle с
  origin-числами (например, bundle.Buildings["metal_mine"].Basic.Metal == 60).

Ф.4. Динамические формулы — internal/origin/economy/
- Создать модуль с production-функциями (приоритет — каждая на
  каждый тик):
  · MetalMineProduction(bundle, level, energy_factor) → float64
  · SiliconMineProduction
  · HydrogenLabProduction(bundle, level, planet_temp, hydro_tech) —
    закрывает D-029 (температурный модификатор)
  · SolarPlantProduction
  · EnergyConsumption (для energy-потребителей)
- Имя пакета — `internal/origin/economy/` (НЕ `internal/legacy/economy/` —
  legacy = только d:\Sources\oxsar2 и game-origin-php; origin = новая
  целевая вселенная).
- Golden-тесты с эталонами из live-origin (Ф.4.1).

Ф.4.1. Сбор golden-эталонов
- Создать tools/dump-planet-prod.php — PHP-CLI скрипт, дампит
  Planet::updateProduction() для разных temp/tech.
- SELECT планет (5-10 разных temp) из na_planet → JSON в
  internal/origin/economy/testdata/golden_planet_*.json.
- Go-golden-тест читает JSON и сверяет. Допуск:
  · точное совпадение для целых
  · абсолютная погрешность ≤ 1 для дробных (PHP eval vs Go math.Round)
  · допуск явно прокомментирован в тесте, не превышать.

Ф.5. Использование bundle в существующих сервисах
- internal/economy/, building/, research/, shipyard/ — переключить
  чтение на bundle через LoadFor(universeCode).
- Все existing game-nova тесты должны оставаться зелёными
  (modern не сломали).
- Создать первый сценарий origin: тест-вселенная с code='origin' и
  файлом configs/balance/origin.yaml — её числа отличаются.

Ф.6. E2E + Smoke
- Поднять dev-стенд game-nova с двумя вселенными:
  uni01 (modern, без override) и origin (с configs/balance/origin.yaml).
- Поставить здание в очередь в каждой → стоимость отличается
  (Metal Mine lvl 1 в origin: 60/15/0/0; в uni01: текущий nova).
- Тик производства даёт origin-числа в origin-вселенной.

Ф.7. Финализация
- Обновить шапку плана 64 → ✅ Завершён <дата>.
- Запись в docs/project-creation.txt — итерация 64.
- В docs/research/origin-vs-nova/divergence-log.md — пометить
  D-022, D-026, D-027, D-028, D-029, D-030 как ✅ ЗАКРЫТО (план 64).

ОБЯЗАТЕЛЬНЫЕ ПРАВИЛА (R0-R15, см. roadmap-report «Часть I.5»):

R0: геймплей nova ЗАМОРОЖЕН — никаких правок чисел modern (uni01/uni02).
R1: snake_case в БД и YAML, snake_case JSON, английский, полные слова.
R3: log/slog с полями user_id/planet_id/event_id/trace_id.
R4: golden + property-based тесты для economy. Покрытие изменённого ≥ 85%.
R8: Prometheus counter+histogram для всех новых service-методов.
R12: i18n — нет хардкода строк; перед созданием новой строки grep
  по projects/game-nova/configs/i18n/, переиспользовать существующие
  ключи. В коммите указать: сколько ключей переиспользовано / новых.
R15: без упрощений как для прода — тесты, обработка ошибок, метрики
  со старта; НЕ «MVP-сокращений» / «TODO: позже».

Полный список конвенций — в плане 64 секция «КОНВЕНЦИИ ИМЕНОВАНИЯ»
и в roadmap-report.md.

GIT-ИЗОЛЯЦИЯ:
- git status --short ПЕРЕД каждым git add
- git add ТОЛЬКО свои пути (configs/balance/, internal/balance/,
  internal/origin/, cmd/tools/import-legacy-balance/, configs/units.yml +
  ships.yml + rapidfire.yml для алиен-юнитов, существующие
  internal/economy/, building/, research/, shipyard/, tools/
  dump-planet-prod.php, docs/plans/64-..., docs/project-creation.txt,
  docs/research/origin-vs-nova/divergence-log.md)
- git status --short ПЕРЕД commit (не захвати чужое)
- НИКОГДА git add . / git add -A

КОММИТЫ:

3 коммита (рекомендация):
1. feat(balance): per-universe balance loader (план 64 Ф.1+Ф.3)
   — loader scaffold + override-схема, существующие nova-тесты
   зелёные.
2. feat(balance): импорт origin → configs/balance/origin.yaml
   + алиен/спец-юниты в default (план 64 Ф.2)
   — CLI-импортёр + сгенерированный YAML.
3. feat(origin/economy): динамические формулы + golden-тесты +
   интеграция (план 64 Ф.4-Ф.6).

Conventional commits + trailer Generated-with: Claude Code (НЕ
Co-Authored-By — git hook уберёт).

ЧЕГО НЕ ДЕЛАТЬ:
- НЕ менять existing YAML configs/buildings.yml, units.yml,
  rapidfire.yml для modern (R0). Только ДОБАВЛЯТЬ алиен/спец-юниты.
- НЕ реализовывать полный DSL-evaluator на Go в рантайме. Только
  в импорт-скрипте.
- НЕ вводить universes.balance_profile column (используем override-
  файлы по universes.code).
- НЕ дублировать ключи i18n — grep сначала (R12).

ОЦЕНКА ОБЪЁМА:

~2 недели работы агента. Если идёт сильно дольше — что-то пошло
не так, перечитай план 64. Если сильно быстрее — проверь R15
(тесты, метрики, обработка ошибок).

УСПЕШНЫЙ ИСХОД:

- LoadDefaults() / LoadFor("uni01") = текущий nova-баланс.
- LoadFor("origin") = origin-числа (Metal Mine 60/15/0/0 и т.д.).
- Алиен-юниты + спец-юниты в дефолте, видны во всех вселенных.
- Все existing game-nova тесты зелёные (R0).
- Шапка плана 64 ✅, D-022/026/027/028/029/030 закрыты.
- 3 коммита, conventional + trailer.
- Разблокированы планы 65, 66, 69.

Стартуй.
```
