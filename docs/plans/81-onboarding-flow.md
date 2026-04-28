# План 81: Onboarding-flow для новых игроков

**Дата**: 2026-04-28
**Статус**: Черновик. Запуск **после плана 74** (публичный запуск) —
сначала нужны реальные новички для измерения drop-off, потом
прицельный onboarding по данным.
**Зависимости**: блокируется планом 74. Желательно после планов 66
(AlienAI), 68 (биржа артефактов), 70 (achievements) — иначе tutorial
будет ссылаться на фичи в работе.
**Связанные документы**:
- [docs/plans/05-ui-features.md](05-ui-features.md) — UI-стек nova-фронта.
- [docs/plans/24-ai-players.md](24-ai-players.md) — боты, могут быть «соседями» новичка для первого боя/контакта.
- [docs/ops/payment-integration.md](../ops/payment-integration.md) — welcome-rewards в оксарах через billing.

---

## Зачем

В F2P-играх без onboarding теряется 40-70% игроков на первом экране
(industry data: GameAnalytics, deltaDNA). Цель плана — провести нового
игрока от регистрации до первого осознанного действия за ~5-15 минут,
закрепить базовые механики, мотивировать продолжить.

**Не цель:**
- Не запускать до плана 74 — без реальных новичков нет данных для
  балансировки tutorial.
- Не делать pixel-perfect tutorial из legacy oxsar2 — у legacy его
  нет в современном виде.
- Не вводить аналитику-провайдеров (GameAnalytics, Mixpanel) на этом
  этапе — Prometheus достаточно для drop-off-метрик.

---

## Применимость по вселенным

- **nova-фронт** (uni01/uni02): полный onboarding по этому плану.
- **origin-фронт** (план 72): отдельный onboarding (origin-стиль) —
  отложить, после nova-варианта проще портировать; либо вообще не
  делать (origin-фронт целит в опытных legacy-игроков, которым
  tutorial не нужен).

---

## Состав

### Ф.1. Backend: onboarding-state + welcome-rewards

- Миграция: `users.onboarding_state JSONB` (default: пустой объект).
  Поля: `current_step`, `completed_steps[]`, `dismissed_hints[]`,
  `welcome_pack_claimed` (boolean), `started_at`, `finished_at`.
- `internal/onboarding/` — пакет:
  - `service.go` — методы `GetState`, `AdvanceStep`, `DismissHint`,
    `ClaimWelcomePack`, `MarkFinished`. R10 (universe-aware).
  - `handler.go` — REST: `GET /api/onboarding/state`,
    `POST /api/onboarding/advance`, `POST /api/onboarding/dismiss-hint`,
    `POST /api/onboarding/welcome-pack`.
  - `metrics.go` — Prometheus counter `oxsar_onboarding_step_total{step,status}`,
    histogram `oxsar_onboarding_duration_seconds{step}`.
- Hook в `auth.Register`: при создании юзера — заинициализировать
  onboarding_state (current_step=0).
- Welcome-pack: при первом GET state клиентом, если
  `welcome_pack_claimed=false` — показать модалку. По клику
  `ClaimWelcomePack` начисляет (через event/economy):
  - 2× стартовых ресурсов на первый день.
  - 1 бесплатный базовый офицер на 24h.
  - 1-2 оксара (показать что валюта существует).
- R8 (Prometheus), R9 (Idempotency-Key на claim), R12 (i18n тексты
  welcome-pack), R3 (slog).

### Ф.2. Tutorial-квестовая цепочка (контент + UI)

10 шагов в `configs/onboarding/tutorial.yaml`:
1. Построй первое здание (металл-шахта).
2. Запусти ускорение / дождись завершения.
3. Произведи первый ресурс (клик по ticker'у).
4. Построй солнечную станцию (объяснить энергию).
5. Изучи первое исследование.
6. Построй верфь.
7. Построй первый корабль.
8. Запусти первую экспедицию.
9. Найди другого игрока на карте.
10. Прочитай первое сообщение в inbox.

Каждый шаг — id, i18n-ключи `onboarding.step.<id>.{title,description}`,
podsvetka-target (CSS-селектор UI-элемента для подсветки), reward.

UI:
- `frontends/nova/src/features/onboarding/`:
  - `OnboardingProvider.tsx` — context, query state из API.
  - `WelcomeTour.tsx` — серия 2-3 вступительных экранов после регистрации.
  - `TutorialQuestPanel.tsx` — постоянный виджет в углу UI «Текущий квест».
  - `Spotlight.tsx` — overlay-подсветка целевого UI-элемента.
  - `ContextualHint.tsx` — тултип при первом появлении экрана.
  - `WelcomePackModal.tsx` — модалка с подарками.
- Hook `useOnboardingHint(hintId)` в каждом главном экране.

### Ф.3. Contextual hints (первые 2-3 часа игры)

20-30 hints, отображаются один раз при первом появлении экрана:
- BuildingsScreen: «Здесь ты строишь здания на планете».
- ResearchScreen: «Исследования открывают новые технологии и юниты».
- ShipyardScreen: «Тут производятся корабли и оборона».
- FleetScreen: «Управление флотом — отправка экспедиций, атак».
- SpaceScreen: «Карта галактики. Координаты — galaxy:system:position».
- AllianceScreen: «Альянсы — кооперация с другими игроками».
- AchievementsScreen: «Достижения дают бонусы и хвастаться есть чем».
- и т.д.

Каждый hint — i18n-ключ `onboarding.hint.<id>`, dismiss-state в БД.

### Ф.4. Метрики и dashboards

- Grafana-панель `Onboarding funnel`: drop-off по шагам tutorial.
- Метрика TTV (Time To Value): время от регистрации до первого боя
  / первой покупки оксаров.
- A/B-инфраструктура (без провайдера, только feature-флаг в
  `users.onboarding_state.variant`) — заложить, использовать
  после первых 500 игроков.

### Ф.5. Тесты

- e2e (Playwright): полный flow регистрация → tutorial шаг 1-10 →
  finished_at заполнен.
- e2e: skip-кнопка обходит tutorial, welcome-pack не теряется.
- Property-based: AdvanceStep идемпотентен (повторный вызов не
  меняет current_step).
- Backend: integration-tests на API.

### Ф.6. Финализация

- Шапка плана 81 ✅.
- Запись итерации в project-creation.txt.
- Обновление CLAUDE.md (новый домен onboarding).
- Документация в docs/ops/runbooks/onboarding-tuning.md: как читать
  drop-off, как править контент tutorial.

---

## Что НЕ делаем

- **Не вводим внешние analytics-провайдеры** — Prometheus + Grafana
  достаточно. Если в будущем потребуется retention-funnel-tooling
  (Amplitude, GameAnalytics) — отдельный план.
- **Не делаем skippable tutorial обязательным** — кнопка skip всегда
  доступна (опытные игроки legacy придут).
- **Не привязываем tutorial к origin-вселенной** — она для legacy-
  ветеранов, им не нужен.
- **Не делаем мобильную адаптивность** — отдельный план.

---

## Объём

~2-3 недели одной сессией.
- Backend: ~600-800 строк Go (миграция, service, handler, metrics, hooks).
- Frontend: ~1000-1500 строк (6-7 компонентов + интеграция в 8+ экранов).
- Контент: ~50 i18n-ключей (10 шагов × {title,description,reward} +
  20-30 hints).
- Тесты: ~400-600 строк (Go integration + Playwright e2e).
- Docs: ~100 строк.

5-7 коммитов.

---

## Триггеры запуска

- ✅ План 74 (публичный запуск) закрыт.
- ✅ Первые 50-100 реальных регистраций случились.
- ✅ Видно реальный drop-off на первом экране (Prometheus покажет).
- ✅ Геймплей стабилизирован (66/68/70 закрыты, baseline-фичи готовы).

До этих триггеров — план остаётся черновиком, не запускается.

---

## Связанные black-box фичи

Фичи которые **могут** быть встроены в tutorial, но запланированы
отдельно:

- **Welcome-pack оксаров** — через billing-client, см. план 77.
- **AI-боты как первые соседи** — план 24, чтобы новичок гарантированно
  кого-то находил на карте на шаге 9.
- **Achievements: «Завершил tutorial»** — план 70, добавить
  специальное достижение `tutorial_completed`.
- **System-mail с приветствием** — после плана 57 (mail-service),
  отправлять welcome-message в inbox для шага 10.

Эти связи фиксируются в плане 81, реализация остаётся в их планах.
