// Package balance — per-universe балансовые бандлы (план 64).
//
// Bundle — корневая структура: оборачивает существующий config.Catalog
// (modern-баланс — uni01, uni02 и любые будущие modern-вселенные) и
// добавляет per-universe override-слой (для вселенной origin).
//
// Архитектурный приём — override-файлы. Дефолтные значения для всех
// вселенных живут в configs/buildings.yml, units.yml, rapidfire.yml,
// research.yml, ships.yml, defense.yml (modern-баланс). Для конкретной
// вселенной X можно положить configs/balance/<universe.id>.yaml —
// перекроет нужные значения. Если файла нет — Bundle == чистый дефолт.
//
// Идентификация вселенной идёт по существующему Universe.ID
// (поле в configs/universes.yaml; в плане 64 называлось "code"
// в origin-терминологии). Никаких новых полей в БД не вводится.
//
// LoadDefaults() возвращает modern-bundle (без override). LoadFor(id)
// возвращает bundle для конкретной вселенной с применённым override
// (если файл есть) или дефолт (если нет).
//
// Для R0-совместимости modern-вселенные читают только дефолтный
// слой — их балансовые числа не пересматриваются. Расхождения с origin
// решаются через configs/balance/origin.yaml, не правкой modern-YAML.
//
// Bundle также экспортирует Globals — глобальные коэффициенты
// (basic_prod_metal/silicon/hydrogen, hydrogen_temp_coefficient/
// _intercept). Они используются динамическими формулами производства
// (internal/origin/economy/*) для расчёта на каждый тик.
package balance
