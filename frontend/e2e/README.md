# E2E-тесты UI

Playwright-спеки для oxsar-nova (план [docs/plans/13-ui-testing.md](../../docs/plans/13-ui-testing.md)).

## Подготовка

1. Поднять инфру и засидить БД:

   ```bash
   make dev-up
   make migrate-up
   PAYMENT_PROVIDER=mock PAYMENT_MOCK_BASE_URL=http://localhost:8080 PAYMENT_RETURN_URL=http://localhost:5173/ make backend-run   # в отдельном терминале
   make worker-run        # в отдельном терминале
   make frontend-run      # в отдельном терминале
   make test-seed         # в отдельном терминале — сидит 5 игроков
   ```

2. Поставить зависимости Playwright (браузеры):

   ```bash
   cd frontend && npx playwright install chromium
   ```

## Запуск

- Все спеки: `make test-e2e` или `cd frontend && npm run test:e2e`
- UI-мод: `cd frontend && npm run test:e2e:ui`
- Один спек: `cd frontend && npx playwright test e2e/smoke.spec.ts`

## Структура

- `fixtures/auth.ts` — `loginAs(page, 'bob')`, фиксированные UUID пользователей из
  [backend/cmd/tools/testseed/main.go](../../backend/cmd/tools/testseed/main.go)
- `helpers/layout.ts` — `expectNoLayoutIssues(page, screenName)` — Ф.4.6,
  проверяет отсутствие горизонтального скролла, пересечений кликабельных
  элементов, нулевых текстовых узлов
- `smoke.spec.ts` — baseline, открывает все 30+ вкладок под bob и alice
- `critical/` — Ф.1 сценарии (auth/buildings/research/…)
- `domains/` — Ф.2 (repair/market/…)
- `edges/` — Ф.4 (offline, i18n, mobile, layout-edge)

## Тестовые пользователи

Пароль у всех: `test-password-123`.

| Логин | Роль | Назначение |
|-------|-------|-------------|
| admin | superadmin | админ-сценарии |
| alice | player | новичок, пустые состояния |
| bob | player | прокачан, богатый state |
| eve | player | жертва атаки (флот bob'а) |
| charlie | player | лидер альянса [UT], союзник bob'а |
