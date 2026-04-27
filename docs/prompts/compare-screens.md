# Готовый промпт: сверка экранов game-origin с legacy

Назначение — быстро запустить сверку UI после крупных правок (переписывание
`src/core/`, замена шаблонизатора, миграции БД). Скопируй секцию ниже целиком
в чат с Claude Code и запусти.

Связано: `projects/game-origin/tools/compare-with-legacy.sh`,
memory-reference `reference_game_origin_routing.md` (только `?go=Page`
работает), `feedback_audit_agent_verify.md` (отчёты Explore-агентов
надо верифицировать руками).

---

## Промпт

> Сравни экраны game-origin с legacy oxsar2.
>
> Запусти `bash projects/game-origin/tools/compare-with-legacy.sh` в фоне
> (`run_in_background: true`, timeout 300000ms). Скрипт логинится в legacy
> (localhost:8080, user `test` / `quoYaMe1wHo4xaci`) и в нашу dev-вселенную
> (localhost:8092 через `dev-login.php`), curl'ит 41 страницу через
> `?go=Page`, нормализует HTML (убирает таймстампы/числа/sid) и пишет
> diff в `projects/game-origin/tools/compare-output/diff/<page>.diff`
> плюс отчёт `report.md`.
>
> Жди завершения через `until grep -q "Report:" <output-file>; do sleep 6; done`
> (тоже в фоне, по уведомлению — не sleep'ом в основном потоке, иначе блок).
>
> После завершения:
>
> 1. Сравни новый `projects/game-origin/tools/compare-output/report.md`
>    с baseline `/tmp/report_baseline.md` через
>    `python /tmp/compare_diff.py /tmp/report_baseline.md
>    projects/game-origin/tools/compare-output/report.md`.
>    Если baseline или скрипт отсутствует — пересоздай скрипт (парсит
>    markdown-таблицу `| Page | size | leg_size | diff | status |`,
>    считает delta diff/size, печатает REGRESSIONS / IMPROVEMENTS /
>    UNCHANGED).
> 2. Прочитай `tail -30` от output-файла compare-скрипта для краткой
>    сводки (этого хватит чтобы увидеть «Итог»).
> 3. Открой diff-файлы тех страниц, где регрессия
>    (`projects/game-origin/tools/compare-output/diff/<page>.diff`).
>    Для каждой реальной регрессии (HTML-структура отличается) — найди
>    причину в коде. Косметические разницы (контент БД, отсутствующие
>    user-style CSS, `©Dominator` в title, `galaxy_distance_mult` value)
>    — НЕ регрессии моего кода.
> 4. Проверь HTTP-коды и размеры:
>    ```bash
>    for p in Main Empire Stock Galaxy Constructions Notepad Alliance MSG \
>             Preferences ExchangeOpts Records Shipyard Research Mission; do
>      c=$(curl -s -b /tmp/oxsar.cookie -o /dev/null -w "%{http_code}" \
>              "http://localhost:8092/game.php?go=$p")
>      s=$(curl -s -b /tmp/oxsar.cookie \
>              "http://localhost:8092/game.php?go=$p" | wc -c)
>      echo "$p: $c size=$s"
>    done
>    ```
>    Все ответы должны быть 200 с **разными** размерами. Одинаковые
>    ~15KB = маска Main page, значит routing не сработал
>    (см. memory `reference_game_origin_routing.md`).
>
> Если baseline/cookie отсутствуют:
> - Логин-cookie:
>   `curl -sc /tmp/oxsar.cookie http://localhost:8092/dev-login.php?userid=1 -o /dev/null`.
> - Если `/tmp/report_baseline.md` потерян (новая машина) — текущий
>   compare-report и есть новый baseline; сохрани его:
>   `cp projects/game-origin/tools/compare-output/report.md /tmp/report_baseline.md`.
>
> Дай краткий итог: число регрессий, число улучшений, ошибок 4xx/5xx,
> замечания по конкретным страницам. Если все 39+/41 без регрессий и 0
> ошибок — подтверждение что предыдущая работа не сломала UI.

---

## Чек-лист предусловий

Перед запуском убедись:

- [ ] Docker-стеки legacy (8080) и game-origin (8092) подняты:
  ```bash
  docker compose -f d:/Sources/oxsar2/docker-compose.yml ps
  docker compose -f projects/game-origin/docker/docker-compose.yml ps
  ```
- [ ] В `na_user` есть юзер `test` (legacy) и `userid=1` (наш) с
  применённым snapshot-фикстурой
  (`tools/apply-test-user-fixture.sh`).
- [ ] memcached активен (иначе isFirstRun-защиты отключены, что
  отражается в diff'е через дополнительные `TOO_MANY_REQUESTS`).

## Если compare сообщает о массовом «-4330 байт на каждой странице»

Это **не регрессия**: legacy-юзер `test` имеет включённые user-style
CSS (`/css/us_bg/a-bg-117.css`, `/css/us_table/table_td2_bg_80.css`),
а наш test-юзер их не имеет. Размер блока ~4.3KB одинаков везде потому
что вставки идут в `<head>` layout-шаблона.
