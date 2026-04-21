# testdata/battle

Golden-файлы для бойцового движка.

- `cases/*.json` — входные сценарии (флоты, технологии, seed).
- `expected/*.json` — ожидаемые Report-ы, верифицированные против
  `d:\Sources\oxsar2-java\assault\dist\oxsar2-java.jar` (см. §14.4 ТЗ).

Добавление фикстуры:

```bash
# 1. положить cases/<name>.json
# 2. прогнать референсный jar:
java -jar <path>/oxsar2-java.jar < cases/<name>.json > expected/<name>.json
# 3. зафиксировать обе стороны в PR
```

TODO: автогенератор фикстур + дифф-отчёт при миссинге на этапе M4.
