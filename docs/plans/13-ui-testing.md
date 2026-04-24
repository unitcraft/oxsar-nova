# План 13: Тестирование UI

Цель: убедиться, что frontend покрывает всю функциональность игры, не падает
на граничных случаях и вызывает все публичные API. Платежи — через
симуляционный шлюз (не трогаем живую Робокассу).

**Статус: Ф.0–Ф.5 первичный проход завершён** (2026-04-24).
См. [docs/ui/test-matrix.md](../ui/test-matrix.md) — актуальная матрица покрытия.

Scope: end-to-end (E2E) прохождение всех 35 экранов + проверка покрытия API.
Вне scope: unit-тесты компонентов (их отсутствие — отдельный tech debt, см.
docs/simplifications.md), визуальная регрессия, нагрузочное тестирование.

---

## Ф.0 Инфраструктура (приоритет: блокер)

### Ф.0.1 Симулятор платежного шлюза `mock` (бэкенд)

Нужен, чтобы проходить весь флоу покупки кредитов без живых денег.
Реализация в `backend/internal/payment/mock.go`:

- `MockGateway implements Gateway`:
  - `BuildPayURL` возвращает URL внутреннего эндпоинта-симулятора:
    `{base}/api/payment/mock/pay?order={orderID}&result=success` (также
    `result=fail` — для проверки негативного сценария).
  - `VerifyWebhook` — без проверки подписи, просто читает `order_id` +
    `amount_kop` из тела/query.
  - `SuccessResponse` пишет `OK{orderID}` plaintext.
- `PAYMENT_PROVIDER=mock` в `config.go` → `service.go` выбирает `MockGateway`.
- Эндпоинт `GET /api/payment/mock/pay` (регистрируется только если provider=mock):
  - При `result=success` → вызывает `svc.ConfirmPayment(orderID, "mock-"+orderID)`,
    редиректит на `PAYMENT_RETURN_URL + ?payment=success`.
  - При `result=fail` → редиректит на `?payment=fail`, заказ остаётся `pending`.
- Признак mock-режима в `/api/payment/packages` → поле `test_mode: true`, чтобы
  UI показывал баннер «Тестовый режим — реальных списаний нет».

Проверка:
- [ ] `PAYMENT_PROVIDER=mock` в `.env.test` поднимает бэкенд без ошибок
- [ ] `go test ./backend/internal/payment/...` — покрытие mock-сценария
- [ ] `/api/payment/order` возвращает URL вида `/api/payment/mock/pay?…`
- [ ] Обращение к mock-URL с `result=success` зачисляет кредиты идемпотентно

### Ф.0.2 CLI-утилита сид-данных для тестов

Файл `backend/cmd/tools/testseed/main.go`. Наполняет БД детерминированным
состоянием для E2E:
- 5 игроков: `admin` (superadmin), `alice` (начинающий), `bob` (прокачанный),
  `eve` (жертва для атаки), `charlie` (союзник по альянсу).
- Alice: стартовая планета, `buildings[metal_mine]=1`.
- Bob: планета с прокачанными зданиями (mine 20, lab 10, shipyard 10),
  исследования до гипердвигателя 2, флот (100 лёгких истребителей), 5000 кр.
- Eve: слабая планета (для теста атаки), рядом с bob по координатам.
- Alliance `[UT]` с charlie (leader) + bob (member) — для тестов альянса.
- Сообщение в inbox alice (welcome).
- Открытый лот артефакта на арт-рынке.
- Готовый лот на товарном рынке (metal→silicon).

Флаг `--reset` сначала `TRUNCATE` всех таблиц (кроме миграций), потом сидит.
Вызов: `make test-seed`.

Проверка:
- [ ] `make test-seed` идемпотентен (повторный запуск даёт тот же state)
- [ ] Все 5 логинов работают через `/api/auth/login`

### Ф.0.3a Docker-стек для CI

`deploy/docker-compose.e2e.yml` поднимает полный стек одной командой:
pg (tmpfs) + redis + migrate + testseed (--reset) + backend (PAYMENT_PROVIDER=mock)
+ worker + frontend + playwright. Healthcheck'и на `/healthz` и vite.
Playwright — последний контейнер с `--exit-code-from playwright` определяет
код выхода всей команды.

`deploy/Dockerfile.testseed` — одноcontainer one-shot, собирает бинарь из
`backend/cmd/tools/testseed`.

`deploy/Dockerfile.playwright` — на базе `mcr.microsoft.com/playwright`
(с chromium внутри), версия образа должна соответствовать версии
`@playwright/test` в `frontend/package.json`.

Запуск:
- `make test-e2e-docker` — полный прогон, билд образов + прогон тестов
- `make test-e2e-docker-down` — очистка

В CI job `e2e` в `.github/workflows/ci.yml` поднимает этот compose,
аттачит playwright-report/test-results/docker-logs при падении.

Проверка:
- [x] `docker compose -f deploy/docker-compose.e2e.yml config` валиден
- [ ] Первый прогон `make test-e2e-docker` зелёный (требуется локальный Docker)

### Ф.0.3 Playwright как E2E-раннер

Почему Playwright: TanStack Query + lazy chunks ломают наивные unit-подходы,
живой браузер закрывает вопрос «UI не падает». Cypress тоже подошёл бы —
выбор в пользу Playwright из-за встроенного trace viewer и параллелизма.

- `frontend/e2e/` — каталог со спеками.
- `frontend/playwright.config.ts` — `baseURL=http://localhost:5173`,
  `webServer` поднимает `make backend-run` + `vite`. Фикстура
  `resetAndSeed` вызывает `make test-seed` перед suite.
- Хелпер `login(page, 'alice')` кладёт JWT в `localStorage` напрямую
  (чтобы не прогонять форму логина в каждом тесте).
- `make test-e2e` = `pnpm playwright test`.

Проверка:
- [ ] `make test-e2e` поднимает dev-стек и прогоняет smoke-спек

### Ф.0.4 Smoke-проверка «ни один экран не падает»

Один спек `smoke.spec.ts`: логинится как `bob`, открывает поочерёдно все 35
вкладок (через `#hash` навигацию), на каждой ждёт отсутствия текста
`Error boundary` / `Ошибка` в DOM и отсутствия `console.error`. На каждой
ошибке — attach скриншот+trace.

Это — baseline. Если в CI падает smoke — остальные спеки не запускаются.

Проверка:
- [ ] Spec проходит для `alice` (новичок) и `bob` (прокачанный игрок)

### Ф.0.5 Аудит покрытия API → UI

Утилита `frontend/scripts/api-coverage.ts`:
1. Парсит `api/openapi.yaml` → список `{method, path}` (≈70 эндпоинтов).
2. Рекурсивно grep'ит `frontend/src/` на `api.get('/api/…')`, `api.post(…)`,
   вытаскивает используемые эндпоинты.
3. Печатает диф: эндпоинты, которые есть в OpenAPI, но не вызываются из UI.
4. Whitelist в `api-coverage.whitelist.txt` для намеренных исключений (webhook,
   admin-only, refresh внутри client).

`make api-coverage` — запуск. В CI — упасть если diff не пуст и не в whitelist.

Проверка:
- [ ] Первый прогон даёт список «висячих» API — разобрать: либо добавить UI,
      либо занести в whitelist с обоснованием

---

## Ф.1 Критический путь (приоритет: HIGH)

Без этих сценариев игра не играется. Каждый сценарий — отдельный `.spec.ts`
в `frontend/e2e/critical/`.

### Ф.1.1 Auth
- [ ] Register: заполнить форму → 201 → авто-логин → виден Overview
- [ ] Login existing → Overview
- [ ] Wrong password → видна ошибка, остаётся на LoginScreen
- [ ] Logout → возврат на LoginScreen
- [ ] Refresh: протухший access_token → автоматический rotate через /refresh,
      запрос не фейлится для пользователя

### Ф.1.2 Buildings (alice, новый игрок)
- [ ] Виден список зданий с уровнями и стоимостью
- [ ] Клик «построить metal_mine» → появляется в очереди с таймером
- [ ] Cancel в очереди → refund ресурсов, очередь пуста
- [ ] Недостаток ресурсов → кнопка disabled + tooltip

### Ф.1.3 Research (bob)
- [ ] Старт исследования → очередь на экране + в Overview
- [ ] Заблокированное по requirements исследование показывает, чего не хватает

### Ф.1.4 Shipyard (bob)
- [ ] Построить 10 лёгких истребителей → очередь × count × per-unit time
- [ ] Построить оборону (ракетная установка ×5)
- [ ] Requirement заблокирован — показана причина

### Ф.1.5 Galaxy (bob)
- [ ] Переход на координаты [1:1:1] → виден список планет
- [ ] Клик по своей планете → «Это ваша планета» (без кнопки атаки)
- [ ] Клик по `eve` → доступны Atk/Spy/Transport

### Ф.1.6 Fleet (bob → eve)
- [ ] Transport: выбрать ресурсы + корабли → отправить → видна миссия в списке
- [ ] Recall активной миссии → миссия меняется на возвращение
- [ ] Attack (kind=1): миссия прилетает, появляется battle-report в messages
- [ ] Spy: отправить 1 шпиона → через N сек — espionage-report в messages

### Ф.1.7 Messages
- [ ] Inbox показывает welcome + battle-report + espionage-report
- [ ] Открыть battle-report → видны потери обеих сторон, debris
- [ ] Unread badge в шапке уменьшается после прочтения
- [ ] Удалить сообщение → оно исчезает из списка
- [ ] Compose: написать alice → alice видит в inbox

### Ф.1.8 Overview
- [ ] Ресурсы (metal/silicon/hydrogen) тикают ResourceTicker'ом
- [ ] Смена планеты в PlanetSwitcher — контент экрана перерисовывается
- [ ] Предупреждение о недостатке энергии показывается, если `energy_remaining<0`

---

## Ф.2 Основная функциональность (приоритет: MEDIUM)

### Ф.2.1 Repair
- [ ] Damaged-юниты после боя видны в списке
- [ ] Disassemble → пачка разобрана, ресурсы вернулись
- [ ] Repair → очередь ремонта, по завершении — юниты целые

### Ф.2.2 Market (обменник + ордеры)
- [ ] Quick exchange metal→silicon по курсу
- [ ] Create lot → виден в публичном списке
- [ ] Accept чужой лот → ресурсы перетекли

### Ф.2.3 Rockets
- [ ] Построить межпланетную ракету (через shipyard)
- [ ] Launch с указанием цели → impact event → уведомление в messages

### Ф.2.4 Artefacts
- [ ] Список артефактов игрока (сидится через testseed)
- [ ] Activate → эффект применён (видно бонус в планете)
- [ ] Deactivate → эффект снят

### Ф.2.5 Artefact Market
- [ ] Открытый лот виден в списке
- [ ] Buy за кредиты → артефакт у покупателя, кредиты списаны
- [ ] Sell: продать свой → лот появился в моих офферах
- [ ] Cancel своего лота → артефакт возвращён

### Ф.2.6 Officers
- [ ] Активировать ADMIRAL за кредиты → таймер, бонус применён
- [ ] Попытка активировать ENGINEER при активном ADMIRAL → ошибка (group_key)

### Ф.2.7 Expeditions
- [ ] Миссия kind=15 на пустую позицию 16 → expedition-report в messages
- [ ] Report показывает исход (resources/artefact/pirates/loss/nothing)

### Ф.2.8 Alliance
- [ ] Create alliance → становится leader
- [ ] Invite → принять на другом аккаунте → member
- [ ] Declare WAR на другой альянс с mutual acknowledge
- [ ] Leave → список членов сократился
- [ ] Disband (leader) → альянс исчез

### Ф.2.9 Chat
- [ ] WS подключается, history подгружается
- [ ] Отправить сообщение → видно в соседнем браузере (второй browser context)
- [ ] Потеря соединения → visible reconnect

### Ф.2.10 Score / Highscore
- [ ] Таблица показывает bob выше alice
- [ ] Клик по строке с координатами → переход в Galaxy на те же координаты
- [ ] Search по нику находит игрока

### Ф.2.11 Achievements
- [ ] Начальные ачивки видны как locked/unlocked (по состоянию сида)
- [ ] Построить первый metal_mine → lazy-check выдаёт achievement в messages

### Ф.2.12 Tutorial / Profession
- [ ] Шаги туториала видны, прогресс-бар заполняется
- [ ] Завершение шага → +10 кредитов (видно в шапке)

### Ф.2.13 Battle Sim
- [ ] Собрать состав атакующих/защитников → запустить → таблица потерь
- [ ] Multi-run (num_sim=5) → показан разброс исходов

### Ф.2.14 Credits (Payments, через mock!)
- [ ] Видны 5 пакетов, виден баннер «Тестовый режим»
- [ ] Клик «Купить starter» → popup → подтвердить success → redirect
      обратно → toast «Оплата прошла» → баланс вырос на 1000
- [ ] Fail-сценарий: `?result=fail` → toast ошибки → баланс не изменился
- [ ] Повторный вызов webhook с тем же order_id → кредиты не задвоились
- [ ] История покупок показывает paid-запись

---

## Ф.3 Второстепенные экраны (приоритет: LOW)

По одному happy-path спеку на экран — проверить, что открывается и показывает
данные из API.

- [ ] Ф.3.1 Empire — сводка по всем планетам
- [ ] Ф.3.2 Techtree — граф рисуется, клик по узлу открывает UnitInfo
- [ ] Ф.3.3 Battlestats — история боёв фильтруется по датам
- [ ] Ф.3.4 Records — топ рекордов
- [ ] Ф.3.5 Notepad — сохраняет текст, выживает reload
- [ ] Ф.3.6 Referral — реф-ссылка копируется, список приглашённых
- [ ] Ф.3.7 Friends — добавить/удалить друга
- [ ] Ф.3.8 Settings — язык переключается (ru↔en), сохраняется
- [ ] Ф.3.9 PlanetOptions — переименовать планету, сделать home
- [ ] Ф.3.10 Resource — детальный отчёт по ресурсам планеты
- [ ] Ф.3.11 UnitInfo — открывается для building/research/ship/defense
- [ ] Ф.3.12 GlobalSearch (Ctrl+K) — поиск по игрокам/альянсам/координатам

### Ф.3.13 Admin (под admin-логином)
- [ ] Stats: числа игроков/планет
- [ ] Users list + ban/unban `eve` → eve не может логиниться
- [ ] Credit grant → у игрока +N кредитов
- [ ] Role change user→admin
- [ ] AutoMsg CMS: отредактировать WELCOME → newuser при регистрации видит новый текст

---

## Ф.4 Граничные случаи и устойчивость (приоритет: MEDIUM)

### Ф.4.1 Сетевые ошибки
- [ ] Offline mode: TanStack Query показывает stale + баннер
- [ ] 500 от сервера: toast ошибки, кнопка retry
- [ ] 401 mid-session: redirect на login

### Ф.4.2 Пустые состояния
- [ ] Новичок без сообщений → «Нет сообщений»
- [ ] Нет артефактов → дружелюдный empty state, не `undefined`
- [ ] Нет миссий → пустая таблица флота

### Ф.4.3 Большие числа
- [ ] Баланс ресурсов > 1 млрд — форматирование `1.2B`, не переполнение
- [ ] Очередь из 1000 кораблей — UI не лагает

### Ф.4.4 i18n
- [ ] Переключение ru↔en без перезагрузки, все экраны перерисовываются
- [ ] Ни один экран не показывает ключ вида `MENU_FOO` (непереведённое)

### Ф.4.5 Мобильный viewport
- [ ] Playwright project `mobile` (375×667): BottomNav виден, MoreSheet
      открывается, все критические экраны не обрезаются

### Ф.4.6 Вёрстка: видимость и отсутствие наложений

Автоматическая проверка макета на каждом экране — элементы не вылезают за
контейнер, не перекрывают друг друга, не обрезаются.

Хелпер `expectNoLayoutIssues(page)` в `frontend/e2e/helpers/layout.ts`:
1. **Отсутствие горизонтального скролла на body**:
   `document.documentElement.scrollWidth <= clientWidth + 1` (1px tolerance).
2. **Видимость ключевых landmark'ов**: `.ox-header`, `.ox-sidebar` (desktop) или
   `.ox-bottom-nav` (mobile), `.ox-content`, `.ox-footer` — все с
   `getBoundingClientRect().width > 0 && height > 0` и не `display:none`.
3. **Нет overflow у интерактивных элементов**: для каждого `button, a, input,
   select, .btn, .ox-nav-btn` проверить, что `scrollWidth <= clientWidth`
   (текст не обрезан `...` там, где его не должно быть) и элемент полностью
   внутри вьюпорта (`rect.right <= window.innerWidth`, `rect.bottom` — для
   above-the-fold).
4. **Нет пересечений между соседними кликабельными элементами** (самое важное
   против «наездов»): взять все `button, a[href], [role="button"]`, видимые в
   вьюпорте; для каждой пары проверить, что их `getBoundingClientRect` не
   пересекаются. Whitelist для намеренных оверлеев (`.ox-modal-overlay`,
   `.ox-planet-dropdown`, `.badge` внутри родителя).
5. **Нет «нулевых» текстовых узлов**: ни один элемент с непустым `textContent`
   не должен иметь `width:0` или `height:0` (признак сломанного flex/grid).
6. **Z-order sanity**: модалки (`.ox-modal`) при открытии перекрывают весь
   контент (bounding box шире/выше содержимого под ними), шапка не
   перекрывается контентом при скролле (`position:sticky` сохраняется).

Прогоняется на каждом экране из smoke-спека (Ф.0.4), в двух viewport'ах
(desktop 1440×900 + mobile 375×667) и в двух состояниях наполнения:
- `alice` — пустые состояния (мало данных, empty states)
- `bob` — полные состояния (длинные списки, большие числа, очереди на 10+
  позиций)

Дополнительно для динамических элементов:
- [ ] Открытые дропдауны (`PlanetSwitcher`, `GlobalSearch`): проверить после
      открытия, что dropdown полностью в вьюпорте и не перекрывается шапкой
- [ ] Модалки (MoreSheet, подтверждения): overlay закрывает весь экран,
      контент модалки центрирован и не выходит за край
- [ ] Toasts: стек из 3+ toast'ов не наезжает на шапку и друг на друга
- [ ] Таблицы с длинными никами/координатами — ячейки не разъезжаются,
      нет горизонтального скролла внутри таблицы (только при явном `overflow-x:auto`)
- [ ] Countdown-таймеры (`<Countdown />`) — ширина стабильна при смене секунд
      (моноширинный шрифт), не дёргают соседние элементы

Проверка:
- [ ] `expectNoLayoutIssues` падает с понятным сообщением («button X at (a,b,c,d)
      overlaps button Y at (e,f,g,h) on #research screen, desktop viewport»)
- [ ] Интегрировано в smoke-spec — каждая из 35 вкладок проверяется
- [ ] Screenshot на падении аттачится в trace

---

## Ф.5 Отчёт о покрытии и оформление

- [x] CI job `e2e` в GitHub Actions: поднимает полный стек через
      `docker compose -f deploy/docker-compose.e2e.yml up`, аттачит
      playwright-report + test-results + docker-logs при падении
- [x] CI job `api-coverage`: `npm run api:coverage` на каждом PR
- [ ] `make api-coverage` зелёный (все API или покрыты UI, или в whitelist)
- [ ] Матрица `docs/ui/test-matrix.md`: экран × сценарий × статус — обновляется
      по итогам Ф.1–Ф.4
- [ ] Итерация в `docs/project-creation.txt`
- [ ] Любые обнаруженные баги — либо фикс, либо тикет-запись в
      `docs/simplifications.md` с приоритетом

---

## Порядок реализации

1. **Ф.0** целиком — фундамент, без него ничего не запустить
2. **Ф.1** параллельно (smoke уже ловит 80% падений)
3. **Ф.2** — домен за доменом, порядок: repair → market → fleet-доп → artefacts →
   art-market → officers → expedition → alliance → chat → score → achievements →
   tutorial → sim → **credits (mock)**
4. **Ф.3** — один прогон по всем в конце
5. **Ф.4** добавляется параллельно с Ф.2 (один sub-spec на категорию)
6. **Ф.5** — финальная сборка отчёта

**Оценка:** Ф.0 — 1 итерация, Ф.1 — 1–2, Ф.2 — 2–3, Ф.3 — 1, Ф.4 — 1, Ф.5 — 0.5.
Итого ~7 итераций.

---

## Риски и открытые вопросы

- WebSocket-тесты (Chat) флакают → стабилизировать через явный `waitFor(msg)`,
  не `sleep`
- Боевой движок детерминирован через seed'нутый RNG — для E2E нужно
  использовать фиксированный seed в testseed, иначе golden-отчёты уплывают
- Playwright webServer стартует долго (~15 с) — не гонять в watch, только в CI
  и перед PR

## Что НЕ делаем в этом плане

- Unit-тесты компонентов (отдельный tech debt)
- Visual regression (Percy/Chromatic) — позже, если появится бюджет
- Нагрузочное (k6) — отдельный план
- Реальные платежи Робокассы в staging — только вручную перед prod-релизом
