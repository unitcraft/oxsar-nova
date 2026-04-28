# Origin Frontend

Pixel-perfect клон визуала legacy-PHP origin (тема standard) на
современном стеке React/TS, работающий на nova-API.

**Статус**: Ф.1 Bootstrap (план
[72-remaster-origin-frontend-pixel-perfect.md](../../../../docs/plans/72-remaster-origin-frontend-pixel-perfect.md)).
Экраны Spring 1-5 — отдельными итерациями.

## Зачем отдельный bundle

Origin — это **отдельная игровая вселенная** на едином nova-движке.
Визуал у origin собственный (legacy-look), не nova-look. Поэтому это
**отдельный Vite-проект**, не часть `frontends/nova/`.

API общий — nova/origin делят backend (`projects/game-nova/backend`)
и тот же `openapi.yaml`.

## Стек

- React 18 + TypeScript (strict, `noUncheckedIndexedAccess`,
  `exactOptionalPropertyTypes`).
- TanStack Query (server-state) + Zustand (UI-state).
- Vite 5 + esbuild.
- I18n: переиспользует nova-bundle (`/api/i18n/{lang}`) — см. R12
  плана 72.

## Запуск

```bash
cd projects/game-nova/frontends/origin
npm install
npm run gen:api          # генерирует src/api/schema.d.ts
npm run dev              # http://localhost:5174
```

Backend (nova) ожидается на `http://localhost:8080` — по умолчанию
проксируется через Vite (см. `vite.config.ts`). Identity-service —
на `:9000`, billing — на `:9100`.

## Скрипты

- `npm run dev` — Vite dev-сервер на порту 5174 (5173 = nova).
- `npm run build` — production build.
- `npm run typecheck` — `tsc -b`.
- `npm run lint` — ESLint.
- `npm run gen:api` — регенерация типов из `../../api/openapi.yaml`.

## Структура

```
src/
  api/          HTTP-клиент + сгенерированные OpenAPI-типы (gitignored)
  i18n/         I18nProvider + useTranslation (переиспользуют nova-bundle)
  layout/       3-frame обёртка (TopHeader / LeftMenu / Planets / Footer)
  stores/       Zustand-store'ы (auth)
  styles/       Тема: theme.css + base.css + layout.css → app.css
  App.tsx       Root-компонент
  main.tsx      Entry-point
public/
  assets/origin/images/   Минимальный набор UI-фонов (см. LICENSING.md)
  favicon.ico
```

## Pixel-perfect (R5)

CSS-переменные в `src/styles/theme.css` извлечены 1:1 из
[`projects/game-legacy-php/public/css/style.css`](../../game-legacy-php/public/css/style.css).
Цвета/отступы/шрифты не меняем без явного ADR — план 73 проверяет
screenshot-diff ≤ 0.5%.

## I18n

Origin переиспользует nova-bundle (`projects/game-nova/configs/i18n/`).
**Перед добавлением любой новой строки** — `grep` по nova-bundle
(R12). Если ключ есть — берём его, не дублируем. Если нет — заводим
по nova-конвенции (snake_case, namespace).

В коммитах указывать соотношение **переиспользовано/новых ключей**
(метрика качества плана 72).

## Что НЕ делаем в первой итерации

- Адаптив (mobile/tablet) — после старта.
- Тёмная тема — после старта.
- **Achievements** — план 70 отложен.
- **Tutorial** — onboarding идёт через portal/identity.
- **Реклама/баннеры** legacy-PHP — не переносятся.
- **Реферальный экран** — кнопка ведёт на portal в новой вкладке.

## Намеренные расхождения с legacy

- BBCode чата → TipTap (план 57 / план 72 Ф.8).
- Шрифты `AGHELV/FUTUR` (legacy proprietary) → системный fallback
  Georgia → подробности в `public/assets/origin/images/LICENSING.md`.
