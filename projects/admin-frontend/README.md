# oxsar-nova admin frontend

Отдельная админ-консоль для оператора системы (план 53).

- React 18 + TypeScript strict + Vite + Tailwind + shadcn/ui (copy-paste primitives).
- OAuth2 PKCE (S256) auth через identity-service.
- Permission-based UI guards (план 52, RBAC).
- Domain: `admin.oxsar-nova.ru` (HTTPS only, HSTS preload, IP-allowlist).

## Dev

```bash
npm install
npm run dev      # http://localhost:5174
npm run build
npm run typecheck
npm run lint
npm run test
```

См. также:
- [docs/plans/53-admin-frontend.md](../../docs/plans/53-admin-frontend.md) — план фаз.
- [docs/architecture/rbac.md](../../docs/architecture/rbac.md) — RBAC-модель (план 52).
