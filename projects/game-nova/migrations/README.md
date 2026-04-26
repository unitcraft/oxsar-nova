# Миграции БД

Используется [goose](https://github.com/pressly/goose). Правила:

- Миграции **append-only** — нумерация возрастает, старые файлы не
  переписываются.
- Down-миграции обязательны для первых 30 дней после релиза.
- Одна миграция = одно логическое изменение. Не смешивать «таблица
  + индекс на чужую таблицу».
- Названия: `NNNN_<short_slug>.sql`.

Применить:

```bash
GOOSE_DRIVER=postgres GOOSE_DBSTRING="postgres://..." goose up
```

или `make migrate-up`.
