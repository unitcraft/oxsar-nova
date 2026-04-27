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
    // /auth/* — auth-service (логин, регистрация, профиль, кредиты).
    // План 36 Ф.11. В проде nginx разводит по доменам, в dev — vite proxy.
    proxy: {
      '/api': process.env.VITE_PORTAL_API ?? 'http://localhost:8090',
      '/auth': process.env.VITE_AUTH_TARGET ?? 'http://localhost:9000',
      '/.well-known/jwks.json':
        process.env.VITE_AUTH_TARGET ?? 'http://localhost:9000',
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
