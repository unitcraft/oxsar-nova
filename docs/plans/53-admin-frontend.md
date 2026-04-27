# План 53: Admin Frontend — отдельная профессиональная админ-консоль

**Дата**: 2026-04-27
**Статус**: Активный
**Зависимости**: **План 51** (rename) и **План 52** (RBAC) должны быть выполнены.
**Связанные документы**: [51-rename-auth-to-identity.md](51-rename-auth-to-identity.md),
[52-rbac-unification.md](52-rbac-unification.md), [54-billing-limits.md](54-billing-limits.md).

---

## Зачем

После плана 52 у нас identity-service знает про роли и permissions,
но **нет UI для администрирования**:
- Управление ролями юзеров (grant/revoke).
- Просмотр audit log.
- Управление биллингом (план 54 даст endpoints).
- Управление UGC/модерацией (план 48 — позже).
- Game-операции (planet ops, fleet recall) — уже есть в game-nova/admin/,
  но раскиданы.

Цель — **единая admin-консоль на отдельном домене** для всех
admin-задач, с production-grade безопасностью (2FA, IP-allowlist,
strict CSP) и **профессиональным техническим дизайном** (стиль
Stripe Dashboard / Linear / Vercel / Grafana — utilitarian, без
геймификации).

## Архитектура

### Стек

- **Frontend**: React 18 + TypeScript 5 + Vite (как везде в монорепо).
- **UI primitives**: **shadcn/ui** (production-grade, accessible,
  copy-paste components, не npm-зависимость → полный контроль) +
  **Radix UI** primitives под капотом.
- **Стайлинг**: **Tailwind CSS 3** (utility-first, минимальный CSS bundle).
- **Иконки**: **Lucide** (line-style, тонкие, professional).
- **State management**: **Zustand** (как в монорепо) для UI state +
  **TanStack Query v5** для server state.
- **Формы**: **react-hook-form** + **zod** валидация (стандарт-де-факто).
- **Таблицы**: **TanStack Table v8** (sorting, filtering, pagination,
  row-selection — всё что нужно для data-dense admin).
- **Графики**: **Recharts** (минималистичные, кастомизируемые).
- **Routing**: **React Router v6**.
- **Auth**: OAuth2 Authorization Code + PKCE (S256), JWT в memory
  (не в localStorage), refresh через httpOnly cookie.
- **i18n**: **только русский** (ru) на старте — локализация админки
  не нужна (один админ-аудитория).

### Размещение

- **Каталог**: `projects/admin-frontend/` (новый проект в монорепо).
- **Backend**: НЕТ собственного бэкенда. Frontend дёргает напрямую:
  - `identity-service` (`/api/admin/*`) — управление ролями.
  - `billing-service` (`/api/admin/billing/*`) — биллинг (план 54).
  - `game-nova` (`/api/admin/*`) — game-операции (мигрируется в плане 52).
  - `moderation-service` (когда будет, план 48).
- **Деплой**: отдельный `Dockerfile.admin` + nginx serve статики +
  CDN опционально.
- **Домен**: `admin.oxsar-nova.ru` (HTTPS only, HSTS preload).

### Безопасность (production-grade)

1. **Network-level**:
   - **IP-allowlist** на nginx/CDN (whitelist из ENV `ADMIN_ALLOW_IPS`,
     CIDR-формат). Запрос с IP вне списка → 403 на nginx-уровне до
     frontend.
   - **HSTS preload** (включает HTTPS-only forever, страховка от
     downgrade).
   - **Mutual TLS (mTLS)** опционально (отложено на отдельную
     суб-задачу) — клиентский сертификат, выдаваемый каждому админу.

2. **Auth-flow**: OAuth2 Authorization Code + PKCE (S256):
   - Юзер заходит на `admin.oxsar-nova.ru` → redirect на
     `identity.oxsar-nova.ru/oauth/authorize?...&code_challenge=...`.
   - identity показывает login-page (логин + 2FA).
   - Успешный auth → redirect обратно с `?code=...`.
   - Frontend обменивает `code` на JWT через POST
     `/oauth/token` с `code_verifier`.
   - Access token в memory (Zustand store, не localStorage).
   - Refresh token в **httpOnly Secure SameSite=Strict cookie**.
   - Auto-refresh за 1 минуту до истечения access токена.

3. **2FA / WebAuthn (passkeys)**:
   - **WebAuthn primary** (FaceID, TouchID, Windows Hello, YubiKey).
   - **TOTP backup** (Google Authenticator, для случаев когда нет
     WebAuthn устройства).
   - Регистрация: при первом логине или через профиль.
   - Recovery codes: 10 одноразовых кодов в момент enrollment.
   - Делается в identity-service, frontend только UI.

4. **Session management**:
   - **Один активный JWT на админа**. Новый логин на другом устройстве
     инвалидирует старый (через `users.session_jti` поле в БД).
   - **Forced logout**: superadmin может invalidate сессии любого
     юзера через identity API.
   - **Idle timeout**: 30 минут без активности → auto-logout (refresh
     не пробрасывается).

5. **Security headers** (strict):
   - `Content-Security-Policy: default-src 'self'; script-src 'self'
     'nonce-RANDOM'; style-src 'self' 'unsafe-inline'; img-src 'self'
     data:; connect-src 'self' https://identity.oxsar-nova.ru
     https://billing.oxsar-nova.ru ...; frame-ancestors 'none';
     base-uri 'self'; form-action 'self'`.
   - `Strict-Transport-Security: max-age=63072000; includeSubDomains; preload`.
   - `X-Frame-Options: DENY`.
   - `X-Content-Type-Options: nosniff`.
   - `Referrer-Policy: strict-origin-when-cross-origin`.
   - `Permissions-Policy: camera=(), microphone=(), geolocation=()`.

6. **Audit logging** на каждое admin-действие (на стороне target
   сервиса, не frontend; план 52 описал).

## Дизайн-спецификация (Design Spec)

### Эталоны и анти-эталоны

**Берём референсы у**:
- **Stripe Dashboard** — data-dense таблицы, минимализм.
- **Linear** — компактность, чистая типографика, dark mode.
- **Vercel Dashboard** — простые формы, accessible.
- **Grafana** — графики и метрики.
- **Sentry** — фильтры, поиск, deep-linking.
- **GitHub Settings** — формы с inline-валидацией.

**НЕ берём**:
- Material Design (Google Material) — слишком игрушечный, гигантские
  отступы.
- Bootstrap 5 — слишком обобщённый, шаблонный.
- AntDesign — корпоративный китайский стиль, перегруженный.

### Принципы

1. **Utilitarian**: каждый пиксель работает. Никаких декоративных
   элементов, тематических иллюстраций, эмодзи в UI.
2. **Data-density**: компактные таблицы (40px row height), inline
   actions, sortable headers, фильтры в шапке. Не «карточки с
   огромными отступами».
3. **Monospace для technical data**: ID, timestamps, JSON, IP-адреса.
4. **Никаких анимаций «эффектности»**: только функциональные
   (loading-spinner, fade-in модального окна).
5. **Ограниченная палитра**:
   - Primary: gray-50 → gray-950 (нейтрал).
   - Accent: один синий (blue-600) для interactive elements.
   - Semantic: red (destructive), green (success), amber (warning),
     blue (info). Mapping через CSS variables.
6. **Dark mode first** (light как опция). Админы работают в тёмных
   IDE.

### Типографика

- **Sans**: **Inter** (system fallback: -apple-system, system-ui).
  Inter из Google Fonts, локально hosted.
- **Mono**: **JetBrains Mono** (для ID, timestamps, кода).
- Размеры: 12px (mono в таблицах), 13px (default UI), 14px (body),
  16px (headings).
- Line-height: 1.4 (compact), 1.6 (reading text).

### Layout

```
┌────────────────────────────────────────────────────────────┐
│  ┌─Topbar─────────────────────────────────────┬─User menu─┐│
│  │ logo  │  search                            │  J. Doe ▾ ││
│  └───────┴────────────────────────────────────┴───────────┘│
├──────────────┬─────────────────────────────────────────────┤
│              │                                              │
│  Sidebar     │  Content area                                │
│              │                                              │
│  Dashboard   │  ┌─Page header─────────────────────────────┐ │
│  Users       │  │  H1   description   [primary action]    │ │
│  Roles       │  └─────────────────────────────────────────┘ │
│  Billing     │                                              │
│  Game ops    │  ┌─Filter bar──────────────────────────────┐ │
│  Audit log   │  │  filters | search | bulk-actions        │ │
│              │  └─────────────────────────────────────────┘ │
│  ──────      │                                              │
│              │  ┌─Data table──────────────────────────────┐ │
│  Settings    │  │  sortable headers | pagination          │ │
│              │  └─────────────────────────────────────────┘ │
└──────────────┴──────────────────────────────────────────────┘
```

- **Topbar**: 48px height, logo слева, command-palette search в центре
  (Cmd+K), user menu справа.
- **Sidebar**: 240px width, collapse до 56px (только иконки) для
  power-users. Group headers, active highlight.
- **Page header**: H1 (16px sm-bold), сero опциональное description
  (13px text-muted), правое-выровненный primary action button.
- **Filter bar**: над таблицами, inline. Filters сохраняются в URL
  query params (deep-linking).
- **Data table**: TanStack Table, 40px rows, hover-highlight,
  sortable columns, fixed header при scroll, pagination внизу.
- **Modals**: shadcn/ui Dialog. Минимум modal'ов — предпочитаем
  inline-edit и slide-over panels.

### Компонентная библиотека (shadcn/ui)

Берём только нужное:
- Button, Input, Textarea, Select, Combobox, Switch, Checkbox.
- Table (поверх TanStack Table).
- Dialog, Sheet (slide-over), Tooltip, Popover, DropdownMenu.
- Toast (notifications).
- Card (для дашборда).
- Tabs, Accordion (если нужны).
- Badge (для status indicators).
- Skeleton (loading states).

Каждый компонент копируется в `src/components/ui/` (не npm-package),
полная свобода кастомизации.

### Страницы

```
/login                              # OAuth redirect-target после auth flow
/                                    # Dashboard
/users                               # Список юзеров
/users/{id}                          # Детали юзера + текущие роли
/users/{id}/edit                     # Edit profile, ban/unban
/users/{id}/roles                    # Grant/revoke roles
/roles                               # Список всех ролей
/roles/{id}                          # Детали роли + permissions
/billing                             # Дашборд биллинга (план 54)
/billing/payments                    # Список транзакций
/billing/refunds                     # Возвраты
/billing/limits                      # Лимит самозанятого
/game-ops/events                     # Dead events (game-nova)
/game-ops/planets                    # Operations над планетами
/game-ops/fleets                     # Force recall
/moderation/reports                  # UGC reports (план 48)
/moderation/blacklist                # Blacklist phrases
/audit                               # Audit log с фильтрами
/audit/{id}                          # Детали записи
/settings                            # Personal settings (2FA, password)
```

### Dashboard (главный экран)

Карточки-метрики (4 в ряд):
- **Active users** (за последние 24h) — число + sparkline.
- **Revenue today** — RUB + delta vs вчера.
- **Pending reports** (UGC moderation) — число + link to /moderation/reports.
- **Dead events** (game-nova) — число + link to /game-ops/events.

Под карточками:
- **Recent admin actions** (last 10 audit-events) — table-row.
- **System health** — status indicators (identity, billing, game-nova,
  game-origin) с last-ping.

Минимум графики на главной — детали в специализированных разделах.

### i18n

- **Только русский** на старте.
- Структура `src/i18n/ru.ts` с typed keys (`as const` + helper).
- В будущем легко добавить en через ту же структуру.

## Этапы

### Ф.1. Скаффолдинг

- `projects/admin-frontend/` — Vite + React 18 + TS strict.
- `package.json`: react, react-dom, react-router-dom v6, zustand,
  @tanstack/react-query v5, @tanstack/react-table v8,
  react-hook-form, zod, lucide-react, tailwindcss, recharts, msw
  (для dev mocks).
- shadcn/ui init: `npx shadcn-ui@latest init` (или manual copy
  компонентов из shadcn/ui registry).
- Tailwind config с кастомной палитрой (gray scale + accent).
- ESLint + Prettier (как в game-nova/portal).
- Vitest + Playwright для тестов.
- Dockerfile (multi-stage: builder → nginx serve static).
- `Makefile`: `admin-run`, `admin-build`, `admin-test`.

### Ф.2. Auth flow (OAuth2 PKCE)

- `src/lib/auth/`:
  - `pkce.ts` — генерация code_verifier (43-128 символов), code_challenge (SHA-256 base64url).
  - `flow.ts` — initiate redirect, exchange code, refresh, logout.
  - `store.ts` — Zustand store для current user + JWT claims.
- Routes: `/login` (initiate flow), `/callback` (process redirect),
  `/logout`.
- Auto-refresh: интерсептор в TanStack Query, 1 минута до expiration.
- 401 response → redirect на /login с saved location.
- Тесты: msw для identity-mock + Playwright happy/error flows.

### Ф.3. Layout + Routing

- `src/components/layout/`:
  - `RootLayout.tsx` — Topbar + Sidebar + Outlet.
  - `Topbar.tsx` — logo, search (placeholder для Cmd+K в Ф.10),
    UserMenu.
  - `Sidebar.tsx` — навигация, collapse state в Zustand.
  - `PageHeader.tsx` — общий header для контента.
- React Router v6 с nested routes, lazy loading для страниц.
- Permission-based route guards: каждый route декорирован списком
  required permissions; если у юзера их нет — 403.

### Ф.4. Users + Roles management

- `/users`: TanStack Table со списком, фильтры (role, status),
  search-by-username/email, пагинация. Use TanStack Query
  with infinite scroll или page-based.
- `/users/{id}`: детали + табы (Profile, Roles, Sessions, Audit).
- Grant/Revoke role: shadcn Dialog с form (role select +
  expires_at + reason textarea).
- `/roles`: компактная таблица со списком ролей.
- `/roles/{id}`: детали + checkboxes для permissions toggle (если
  superadmin).

### Ф.5. Audit log

- `/audit`: фильтры (actor, target, action, date-range),
  full-text search по reason.
- Каждая запись: actor, target, action, role/permission,
  timestamp, IP, user-agent.
- Кликабельный actor → /users/{id}, target → /users/{id}.
- Export to CSV.

### Ф.6. Game-ops миграция

- Перенести существующие game-nova/admin страницы в admin-frontend:
  - `/game-ops/events` — dead events (resurrect, retry, cancel).
  - `/game-ops/planets` — rename, transfer, delete.
  - `/game-ops/fleets` — force recall.
- Удалить `projects/game-nova/frontend/src/features/admin/` (UI
  переехал; backend endpoints остаются).
- Заменить ссылки в game-nova UI на «Открыть admin-консоль» (если
  юзер имеет admin-permissions).

### Ф.7. Moderation (заглушки)

- `/moderation/reports` — заглушка с placeholder-table (план 48
  откроется отдельно).
- `/moderation/blacklist` — view-only список (для редактирования
  пока через прямой git PR в configs/moderation/blacklist.yaml).

### Ф.8. Settings + 2FA

- `/settings/profile` — change name/email/password.
- `/settings/security`:
  - WebAuthn enrollment (button «Add passkey» → browser
    credentials API).
  - TOTP enrollment (QR-code + backup codes display).
  - Recovery codes regeneration.
  - Active sessions с возможностью force logout остальных.
- Backend endpoints в identity-service (план 52 base + this plan
  extends).

### Ф.9. Security hardening

- nginx config с CSP, HSTS, X-Frame-Options и прочими headers.
- IP-allowlist в `deploy/nginx/admin.conf`:
  ```
  geo $admin_allowed { default 0; include /etc/nginx/admin-ips.conf; }
  if ($admin_allowed = 0) { return 403; }
  ```
- ADMIN_ALLOW_IPS в `deploy/.env.production`.
- Smoke: запрос с IP не из списка → 403; с IP из списка → 200.

### Ф.10. UX polish

- **Cmd+K command palette** (cmdk library) — quick navigation, search
  по юзерам и actions.
- **Keyboard shortcuts** (J/K для навигации в таблицах, ? для help).
- **Empty states** для пустых таблиц.
- **Skeleton loaders** во всех async-компонентах.
- **Toast notifications** для успехов/ошибок (shadcn Toaster).
- **Error boundary** на root level.

### Ф.11. CI/CD + deploy

- `.github/workflows/admin-frontend.yml` — typecheck, lint, test,
  build.
- Dockerfile multi-stage: builder (npm install + vite build) →
  nginx-alpine (serve dist/).
- `deploy/docker-compose.production.yml`: service `admin-frontend`,
  bind to `admin.oxsar-nova.ru`.
- `deploy/nginx/admin.conf`: TLS, security headers, IP-allowlist,
  proxy to static files.
- Smoke в production: открытие dashboard, логин, базовый flow.

### Ф.12. Документация

- `docs/architecture/admin-frontend.md` — архитектура, стек,
  flow auth.
- `docs/ops/admin-access.md` — как получить доступ:
  1. Bootstrap superadmin через identity-cli.
  2. Add IP в `ADMIN_ALLOW_IPS`.
  3. Setup WebAuthn / TOTP.
  4. Login flow.
- `docs/design/admin-design-spec.md` — design system documentation
  (палитра, типографика, компоненты, do's/don'ts).
- `docs/project-creation.txt` — итерация 53.

## Тестирование

### Unit / Component (Vitest)

- Каждый shadcn-component обернут в `*.test.tsx`.
- Hooks (useAuth, usePermission) — отдельные тесты.
- Утилиты (PKCE, JWT parsing) — coverage 100%.

### Integration (Vitest + msw)

- Auth flow с моком identity (msw handlers).
- TanStack Query кейсы: pagination, filtering, mutations.
- Permission-based UI: разный JWT → разные кнопки видны.

### E2E (Playwright)

- Login flow (real identity test instance).
- Grant role to user → check audit log.
- Revoke role → permission removed.
- Failed 2FA → error message.
- IP outside allowlist → 403.

### Visual regression

- Storybook + Chromatic (опционально, отдельной задачей).

## Производительность

- Bundle size budget: < 300KB gzipped main chunk.
- Code-splitting по routes (React Router lazy).
- Tailwind PurgeCSS убирает неиспользуемые классы.
- Tree-shaking lucide-react (только нужные иконки).
- React Query staleTime/cacheTime настроены для admin-данных
  (большой staleTime — данные редко меняются, кроме audit log).

## Доступность (a11y)

- shadcn/ui построен на Radix UI → keyboard navigation, ARIA-labels,
  focus management из коробки.
- Контраст: gray-text на фоне ≥ 4.5:1 (WCAG AA).
- Все интерактивные элементы достижимы по Tab.
- Скринридеры: проверять через axe-core в Vitest.

## Риски

1. **shadcn-компоненты vs npm-deps**: copy-paste в репо означает что
   обновления не приходят автоматически. Митигация: в каждом
   компоненте комментарий с ссылкой на source + version, ежемесячный
   audit.
2. **Bundle size**: TanStack Table + Recharts могут разрастись.
   Митигация: monitoring через `npm run build --report`, lazy-load
   Recharts только на dashboard.
3. **2FA UX**: WebAuthn сложен для старых браузеров. Митигация:
   TOTP fallback. Recovery codes хранятся юзером безопасно.
4. **Параллельная сессия с соседним агентом**: при разработке
   admin-frontend кто-то может править game-nova/admin. Делать в
   feature-branch.

## Out of scope

- mTLS клиентских сертификатов (отложено).
- Real-time updates через WebSocket (полить через React Query refetch
  достаточно).
- Mobile-friendly admin (desktop-only пока).
- Многотенантность (один tenant сейчас).

## Альтернативы (отвергнуты)

- **Расширить game-nova/frontend admin-таб**: смешивает игровой UI
  и админ. Невозможна изоляция домена/IP-allowlist.
- **Готовый admin-фреймворк (react-admin, refine)**: hard to customize,
  большой bundle, прокладка над AntDesign — не наш стиль.
- **Storybook как admin-UI**: только для разработки, не для прода.
- **Server-side rendered admin (Next.js)**: для admin не нужен SEO,
  Vite SPA проще.

## Итог

Отдельный `projects/admin-frontend/` — production-grade admin-консоль
с профессиональным utilitarian-дизайном (стиль Stripe/Linear), OAuth2
PKCE flow, 2FA через WebAuthn+TOTP, IP-allowlist, strict CSP. Готова
для подключения биллинг-секции (план 54), модерации (план 48) и
будущих доменов.
