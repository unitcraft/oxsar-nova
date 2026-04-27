# Formula DSL legacy game-origin

**Дата сборки**: 2026-04-28
**Контекст**: артефакт плана 62 — описание DSL формул баланса в
game-origin. Используется для проектирования параметризации nova
под legacy-режим (план 63+).

---

## Источники истины

В game-origin балансовые данные хранятся **в трёх разных местах**:

1. **`na_construction` (БД, varbinary-строки-формулы)** — основной
   источник. Колонки `prod_*`, `cons_*`, `charge_*`, `basic_*`
   содержат либо число (`basic_*`), либо PHP-выражение в виде
   строки (`prod_*`, `cons_*`, `charge_*`).
2. **`config/consts.php` (PHP define)** — константы с ID юнитов и
   зданий (`UNIT_METALMINE=1`, `UNIT_HYDROGEN_LAB=3` и т.д.),
   глобальные параметры (`METAL_BASIC_PROD=20`).
3. **`config/params.php` (PHP)** — конфиги приложения, начальные
   ресурсы, мультипликаторы.

Парсер строк-формул:
- **`Functions.inc.php:41`** — `parseChargeFormula()` (для
  стоимостей по уровню — `charge_*`).
- **`Planet.class.php:592`** — `parseSpecialFormula()` (для
  производства/потребления — `prod_*`, `cons_*` с контекстом
  планеты).

---

## Таблица `na_construction` — DDL

```sql
CREATE TABLE `na_construction` (
  `buildingid` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `race` tinyint(3) unsigned NOT NULL DEFAULT '1',
  `mode` tinyint(3) unsigned NOT NULL,
  `name` varbinary(255) NOT NULL,
  `test` tinyint(1) unsigned NOT NULL DEFAULT '0',
  `front` tinyint(3) unsigned NOT NULL DEFAULT '10',
  `ballistics` tinyint(3) unsigned NOT NULL DEFAULT '0',
  `masking` tinyint(3) unsigned NOT NULL DEFAULT '0',
  `basic_metal` double(15,0) unsigned NOT NULL,
  `basic_silicon` double(15,0) unsigned NOT NULL,
  `basic_hydrogen` double(15,0) unsigned NOT NULL,
  `basic_energy` double(15,0) unsigned NOT NULL,
  `basic_credit` int(10) unsigned NOT NULL DEFAULT '0',
  `basic_points` int(10) unsigned NOT NULL DEFAULT '0',
  `prod_metal` varbinary(255) NOT NULL,
  `prod_silicon` varbinary(255) NOT NULL,
  `prod_hydrogen` varbinary(255) NOT NULL,
  `prod_energy` varbinary(255) NOT NULL,
  `cons_metal` varbinary(255) NOT NULL,
  `cons_silicon` varbinary(255) NOT NULL,
  `cons_hydrogen` varbinary(255) NOT NULL,
  `cons_energy` varbinary(255) NOT NULL,
  `charge_metal` varbinary(255) NOT NULL,
  `charge_silicon` varbinary(255) NOT NULL,
  `charge_hydrogen` varbinary(255) NOT NULL,
  `charge_energy` varbinary(255) NOT NULL,
  `charge_credit` varbinary(255) NOT NULL DEFAULT '',
  `charge_points` varbinary(255) NOT NULL,
  `special` varbinary(255) NOT NULL,
  `demolish` float NOT NULL,
  `display_order` int(10) unsigned NOT NULL,
  PRIMARY KEY (`buildingid`),
  KEY `mode` (`mode`,`test`),
  KEY `display_order` (`display_order`,`buildingid`)
) ENGINE=InnoDB AUTO_INCREMENT=366 DEFAULT CHARSET=binary;
```

### Ключевые колонки

| Колонка | Тип | Назначение |
|---|---|---|
| `mode` | TINYINT | 1=здания, 2=исследования, 3=флот, 4=оборона |
| `basic_*` | DOUBLE | базовые числа на уровень 1 (как есть) |
| `prod_*` | VARBINARY(255) | формула производства за час (DSL-строка) |
| `cons_*` | VARBINARY(255) | формула потребления за час (DSL-строка) |
| `charge_*` | VARBINARY(255) | формула стоимости постройки уровня N (DSL-строка) |
| `special` | VARBINARY(255) | специальная формула (например, ёмкость хранилищ) |
| `demolish` | FLOAT | коэффициент возврата ресурсов при сносе |
| `front` / `ballistics` / `masking` | TINYINT | боевые параметры (для mode=3,4) |

**Уникальность**: одна таблица содержит **здания, исследования,
флот И оборону** одновременно — отделяются полем `mode`. В nova
эти сущности разделены по разным конфигам (`buildings.yml`, `units.yml`,
`defense.yml`).

---

## Парсер формул

### `parseChargeFormula()` — базовый (без контекста планеты)

**Файл**: `projects/game-origin/src/game/Functions.inc.php:41`

```php
function parseChargeFormula($formula, $basic, $level)
{
    $formula = Str::replace("{level}", $level, $formula);
    $formula = Str::replace("{basic}", $basic, $formula);
    $formula = trim($formula);
    if(!$formula)
    {
        $formula = $basic;
    }
    $result = 0;
    eval("\$result = ".$formula.";");
    return max(0, round($result));
}
```

**Логика**:
1. Подставляет `{level}` → числовое значение.
2. Подставляет `{basic}` → значение из колонки `basic_*`.
3. Если формула пуста — возвращает `$basic` (на уровне 1).
4. **Вычисляет через PHP `eval()`** — DSL = подмножество PHP.
5. `round` + `max(0, ...)`.

**⚠️ Безопасность**: использование `eval()` — anti-pattern. Формулы
поступают из БД, и любая SQL-инъекция в `na_construction` приводит
к RCE. В nova-режиме нужен изолированный парсер (recursive descent
или ExpressionLanguage-стиль).

### `parseSpecialFormula()` — расширенный (с контекстом планеты)

**Файл**: `projects/game-origin/src/game/Planet.class.php:592`

```php
protected function parseSpecialFormula($formula, $level)
{
    $formula = Str::replace("{level}", $level, $formula);
    $formula = Str::replace("{temp}", $this->data["temperature"], $formula);
    $_self = $this;
    $formula = preg_replace_callback("#\{tech\=([0-9]+)\}#i", function($m) use($_self){
        return $_self->getResearch($m[1]);
    }, $formula);
    $formula = preg_replace_callback("#\{building\=([0-9]+)\}#i", function($m) use($_self){
        return $_self->getBuilding($m[1]);
    }, $formula);
    $result = 0;
    eval("\$result = ".$formula.";");
    return round($result);
}
```

**Дополнительные замены**:
- `{temp}` → температура планеты (`$this->data["temperature"]`).
- `{tech=NN}` → уровень технологии NN (`getResearch(NN)`).
- `{building=NN}` → уровень здания NN на этой планете (`getBuilding(NN)`).

**Используется** для `prod_*` и `cons_*` (контекст планеты нужен:
температура для водорода, уровень технологии для производства).

---

## Синтаксис DSL

### Переменные

| Переменная | Доступна в | Значение |
|---|---|---|
| `{level}` | оба парсера | текущий уровень здания/юнита |
| `{basic}` | `parseChargeFormula` | значение из `basic_*` |
| `{temp}` | `parseSpecialFormula` | температура планеты (-200..+200) |
| `{tech=NN}` | `parseSpecialFormula` | уровень технологии с buildingid=NN |
| `{building=NN}` | `parseSpecialFormula` | уровень здания с buildingid=NN на текущей планете |

### Операторы

- Арифметика: `+`, `-`, `*`, `/`, `**` (возведение в степень)
- Скобки: `()` для группировки
- Унарный минус: `-X`

### Функции

Любые PHP-функции, доступные через `eval()`:
- `pow(base, exp)` — основная для экспоненциальных шкал
- `floor(x)`, `ceil(x)`, `round(x)`, `abs(x)`
- `min(a, b, ...)`, `max(a, b, ...)`

---

## Примеры формул из БД

### Здания (mode=1)

| Здание | basic_metal | charge_metal | prod_metal |
|---|---|---|---|
| Metal Mine (id=1) | 60 | `floor({basic} * pow(1.5, ({level} - 1)))` | `floor(30 * {level} * pow(1.1+{tech=23}*0.0006, {level}))` |
| Silicon Lab (id=2) | 48 | `floor({basic} * pow(1.6, ({level} - 1)))` | `floor(20 * {level} * pow(1.1+{tech=24}*0.0007, {level}))` |
| Hydrogen Lab (id=3) | 225 | `floor({basic} * pow(1.5, ({level} - 1)))` | `floor(10 * {level} * pow(1.1+{tech=25}*0.0008, {level}) * (-0.002 * {temp} + 1.28))` |
| Solar Plant (id=4) | 75 | (стандарт) | (отсутствует — generates `charge_energy`) |
| Hydrogen Plant (id=5) | 900 | `floor(50 * {level} * pow(1.1+{tech=18}*0.0005, {level}))` | (как basic — простая шкала) |
| Metal Storage (id=9) | 1000 | `20 * pow(1.5, {level})` | (через `special`) |

### Особые случаи

- **Hydrogen Lab**: формула содержит `{temp}` — холодные планеты
  производят больше водорода (-0.002 * temp + 1.28).
- **Hydrogen Lab cons_metal**: `floor(20 * {level} * pow(1.1-{tech=18}*0.0005, {level}))`
  — потребление **уменьшается** с уровнем энергетической технологии.
- **Solar Plant**: производит энергию, не имеет `charge_metal` со
  стандартной шкалой.

### Простые vs сложные формулы

- **Простая**: `floor({basic} * pow(1.5, ({level} - 1)))` — типовая
  для большинства зданий-стандартов (×1.5 каждый уровень).
- **С технологией**: `floor(30 * {level} * pow(1.1+{tech=23}*0.0006, {level}))`
  — производство растёт с уровнем технологии оружия (id=23).
- **С температурой**: только водородные лаборатории.

---

## Сопутствующие таблицы

### `na_attack_formation`

```sql
CREATE TABLE `na_attack_formation` (
  `eventid` int(10) unsigned NOT NULL,
  `name` varbinary(128) NOT NULL,
  `time` int(10) unsigned NOT NULL,
  PRIMARY KEY (`eventid`)
) ENGINE=InnoDB DEFAULT CHARSET=binary;
```

Минимальная структура. Хранит именованные боевые формации
(альянсовые атаки). Без формул — параметры берутся из
`na_construction`.

### `na_artefact_datasheet`

```sql
CREATE TABLE `na_artefact_datasheet` (
  `typeid` int(11) unsigned NOT NULL DEFAULT '0',
  `buyable` tinyint(1) unsigned NOT NULL DEFAULT '1',
  `auto_active` tinyint(1) unsigned NOT NULL DEFAULT '0',
  `movable` tinyint(1) unsigned NOT NULL DEFAULT '1',
  `unique` tinyint(1) unsigned NOT NULL DEFAULT '0',
  `usable` tinyint(1) unsigned NOT NULL DEFAULT '0',
  `trophy_chance` tinyint(1) unsigned NOT NULL DEFAULT '1',
  `delay` int(11) unsigned NOT NULL DEFAULT '0',
  `use_times` int(11) unsigned NOT NULL DEFAULT '0',
  `use_duration` int(11) unsigned NOT NULL DEFAULT '0',
  `lifetime` int(11) unsigned NOT NULL DEFAULT '0',
  `effect_type` tinyint(1) unsigned NOT NULL DEFAULT '0',
  `max_active` int(10) unsigned NOT NULL DEFAULT '0',
  `quota` float NOT NULL DEFAULT '1',
  PRIMARY KEY (`typeid`),
  CONSTRAINT `na_artefact_datasheet_ibfk_1` FOREIGN KEY (`typeid`)
    REFERENCES `na_construction` (`buildingid`)
    ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
```

Параметры артефактов (флаги поведения, лимиты использования). FK
на `na_construction.buildingid`. Формулы стоимости — там же.

### `na_engine`

```sql
CREATE TABLE `na_engine` (
  `engineid` int(4) unsigned NOT NULL,
  `factor` int(3) unsigned NOT NULL,
  PRIMARY KEY (`engineid`)
) ENGINE=InnoDB DEFAULT CHARSET=binary;
```

Множители скорости двигателей. Без формул — простой `factor`
(целое число).

### `na_options`

Не используется для формул. Конфигурация runtime-параметров —
в `config/consts.php` и `config/params.php`.

---

## Архитектурное расхождение (для журнала D-NNN)

В nova **формулы хранятся в Go-коде** (`projects/game-nova/backend/
internal/economy/`), а числа — в YAML (`configs/units.yml` и т.д.).
В origin **формулы хранятся в БД** как DSL-строки и интерпретируются
PHP-`eval()`.

Для legacy-режима в nova нужно решить (см. план 62 § Категория 2):
- **B1**: извлечь все формулы origin в `configs/balance/legacy.yaml`
  как **числа** (вычислив каждый уровень заранее) и переключаться
  по `universe.balance_profile`.
- **B2**: добавить в nova поддержку DSL-формул в БД (новые таблицы
  `legacy_construction`, парсер). **Anti-pattern** — две системы
  хранения баланса.
- **B3**: гибрид — простые числа в YAML, формулы только там, где
  нужна динамика (температура, уровень технологии).

**Предпочтительно B1+B3**: предвычислить charge_* (стоимость по
уровням) в YAML, оставить динамику prod_* (с температурой/технологией)
как Go-функции в `internal/legacy/` под флагом
`universe.balance_profile = legacy`.

Подробное предложение — в `divergence-log.md` (D-001 — DSL формул).

---

## References

- `projects/game-origin/src/game/Functions.inc.php:41` — `parseChargeFormula`
- `projects/game-origin/src/game/Planet.class.php:592` — `parseSpecialFormula`
- `projects/game-origin/migrations/001_schema.sql:727-762` — DDL
  `na_construction`
- `projects/game-origin/migrations/002_data.sql` — seed (фактический
  баланс)
- [docs/balance/analysis.md](../../balance/analysis.md) — формулы nova
- [docs/legacy/game-reference.md](../../legacy/game-reference.md) §
  «Параметры юнитов из БД легаси» — таблица `na_ship_datasheet`,
  `na_rapidfire`
