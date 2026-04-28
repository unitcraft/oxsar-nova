# Готовый промпт: сверка экранов game-origin с legacy

Назначение — быстро запустить сверку UI после крупных правок (переписывание
`src/core/`, замена шаблонизатора, миграции БД). Скопируй секцию ниже целиком
в чат с Claude Code и запусти.

Связано: `projects/game-legacy-php/tools/compare-with-legacy.sh`,
memory-reference `reference_game_origin_routing.md` (только `?go=Page`
работает), `feedback_audit_agent_verify.md` (отчёты Explore-агентов
надо верифицировать руками).

---

## Промпт

> Сравни экраны game-origin с legacy oxsar2.
>
> Запусти `bash projects/game-legacy-php/tools/compare-with-legacy.sh` в фоне
> (`run_in_background: true`, timeout 300000ms). Скрипт логинится в legacy
> (localhost:8080, user `test` / `quoYaMe1wHo4xaci`) и в нашу dev-вселенную
> (localhost:8092 через `dev-login.php`), curl'ит 41 страницу через
> `?go=Page`, нормализует HTML (убирает таймстампы/числа/sid) и пишет
> diff в `projects/game-legacy-php/tools/compare-output/diff/<page>.diff`
> плюс отчёт `report.md`.
>
> Жди завершения через `until grep -q "Report:" <output-file>; do sleep 6; done`
> (тоже в фоне, по уведомлению — не sleep'ом в основном потоке, иначе блок).
>
> После завершения:
>
> 1. Сравни новый `projects/game-legacy-php/tools/compare-output/report.md`
>    с baseline `/tmp/report_baseline.md` через
>    `python /tmp/compare_diff.py /tmp/report_baseline.md
>    projects/game-legacy-php/tools/compare-output/report.md`.
>    Если baseline или скрипт отсутствует — пересоздай скрипт (парсит
>    markdown-таблицу `| Page | size | leg_size | diff | status |`,
>    считает delta diff/size, печатает REGRESSIONS / IMPROVEMENTS /
>    UNCHANGED).
> 2. Прочитай `tail -30` от output-файла compare-скрипта для краткой
>    сводки (этого хватит чтобы увидеть «Итог»).
> 3. Открой diff-файлы тех страниц, где регрессия
>    (`projects/game-legacy-php/tools/compare-output/diff/<page>.diff`).
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
>   `cp projects/game-legacy-php/tools/compare-output/report.md /tmp/report_baseline.md`.
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
  docker compose -f projects/game-legacy-php/docker/docker-compose.yml ps
  ```
- [ ] В `na_user` есть юзер `test` (legacy) и `userid=1` (наш) с
  применённым snapshot-фикстурой
  (`tools/apply-test-user-fixture.sh`).
- [ ] memcached активен (иначе isFirstRun-защиты отключены, что
  отражается в diff'е через дополнительные `TOO_MANY_REQUESTS`).

## ~~Если compare сообщает о массовом «-4330 байт на каждой странице»~~ — ЗАКРЫТО (план 50 Ф.0, 2026-04-27)

Раньше эта секция гласила, что разница ~4.3KB на каждой странице — это
не регрессия (у legacy-юзера CSS user-style включён, у нашего нет).
**Это было неверно**: причина — регрессия плана 43, `getUserStyles()`
искал CSS через `APP_ROOT_DIR` (= `src/`) при том что после
реструктуризации `css/` живёт в `public/`. План 50 Ф.0 заменил пути
на `GAME_ORIGIN_DIR."public/"`. Теперь user-style CSS подгружается
идентично у нас и в legacy.

## Известные стабильные «🔴 major» diff'ы — не регрессии

Эти страницы стабильно показывают major diff (>700 lines) в каждом
прогоне. Это **не баги нашего кода**, не требуют исправления:

- **BuildingInfo / UnitInfo / ArtefactInfo** — страницы по дизайну
  требуют `?id=N` в URL. Compare-скрипт зовёт их без id, обе
  вселенные кидают `GenericException("Unkown building/unit/...")`.
  После плана 37.5f наш error-page стилизован как у legacy
  (`<title>GenericException</title>` + красный заголовок), но
  **БЕЗ stack trace и source-excerpt** (security-leak). Diff с
  legacy ~30-50 lines (legacy показывает stack), это правильно.
  Чтобы свести diff к минимуму — нужно расширить compare-скрипт
  поддержкой sub-параметров и звать `?go=BuildingInfo&id=1`.
  Не сделано. См. план 37.5f.

- **Changelog** — у legacy таблицы `na_changelog` нет, контент
  тянулся из внешнего `netassault.ru` (legacy update-server). У нас
  в `Changelog::index()` стоит TODO (план 37.5d.7) — пустой массив
  `release`. Diff ~811 lines = весь legacy-changelog. Чтобы закрыть
  — нужна собственная таблица `na_changelog` со списком наших
  патчей (отдельная задача, см. план 37.5f).

## Косметические разницы (НЕ регрессии)

- `©Dominator` в `<title>` legacy — у нас этого скина нет.
- `galaxy_distance_mult = 1` (наша) vs `20000` (legacy) — настройка
  вселенной, не код.
- Разные ресурсы планет (`7.000.000` металла у нашего test vs
  `147.650.000` у legacy) — состояние БД, не код.
- Разные текущие планеты (`Hello World` vs `Фокус`) —
  `cur-planet` маркер от того, какая планета выбрана.
- Появление блока `<div class="oxsar-footer">` (12+ + legal-links) —
  целевое изменение от плана 50 Ф.3+Ф.6, добавляет ~12 lines
  diff на каждой странице.
