import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { fileURLToPath, URL } from 'node:url';

export default defineConfig(({ command }) => ({
  plugins: [react()],
  resolve: {
    alias: { '@': fileURLToPath(new URL('./src', import.meta.url)) },
  },
  server: {
    port: 5174,
    host: true,
    allowedHosts: true,
    // /api/* — portal-backend (новости, feedback, /api/universes).
    // /auth/* — identity-service (логин, регистрация, профиль, кредиты).
    // План 36 Ф.11. В проде nginx разводит по доменам, в dev — vite proxy.
    proxy: {
      '/api': process.env.VITE_PORTAL_API ?? 'http://localhost:8090',
      '/auth': {
        target: process.env.VITE_IDENTITY_TARGET ?? 'http://localhost:9000',
        changeOrigin: true,
        // /auth/handoff — frontend-route (план 36 Ф.8). См. game-nova vite.config.ts.
        bypass: (req) => {
          if (req.url?.startsWith('/auth/handoff')) {
            return '/index.html';
          }
          return null;
        },
      },
      '/.well-known/jwks.json':
        process.env.VITE_IDENTITY_TARGET ?? 'http://localhost:9000',
      // План 38 Ф.7: billing-service для кошельков и платежей.
      '/billing': process.env.VITE_BILLING_TARGET ?? 'http://localhost:9100',
    },
  },
  build: {
    target: 'es2022',
    sourcemap: command === 'serve',
    outDir: 'dist',
  },
}));
