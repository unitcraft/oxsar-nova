// Vite-конфиг origin-фронта (план 72 Ф.1).
//
// Origin-фронт — это **отдельный bundle**, визуально pixel-perfect
// клон legacy-PHP (тема standard) на современном стеке. Работает
// на nova-API без backend-адаптеров (R6 плана 72).
//
// Сетевая модель такая же, как у nova-фронта (см. соседний
// vite.config.ts):
//   - /api → backend nova на 8080
//   - /auth → identity-service на 9000
//   - /billing → billing-service на 9100
//
// Порт dev-сервера — 5174 (5173 занят nova-фронтом, чтобы можно было
// держать обе тулзы одновременно открытыми во время разработки).

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
    watch: { usePolling: true, interval: 500 },
    proxy: {
      '/api': process.env.VITE_PROXY_TARGET ?? 'http://localhost:8080',
      '/healthz': process.env.VITE_PROXY_TARGET ?? 'http://localhost:8080',
      '/auth': {
        target: process.env.VITE_IDENTITY_TARGET ?? 'http://localhost:9000',
        changeOrigin: true,
        bypass: (req) => {
          if (req.url?.startsWith('/auth/handoff')) {
            return '/index.html';
          }
          return null;
        },
      },
      '/.well-known/jwks.json':
        process.env.VITE_IDENTITY_TARGET ?? 'http://localhost:9000',
      '/billing': process.env.VITE_BILLING_TARGET ?? 'http://localhost:9100',
    },
    warmup: {
      clientFiles: ['./src/main.tsx', './src/App.tsx'],
    },
  },
  optimizeDeps: {
    include: [
      'react',
      'react-dom',
      'react-dom/client',
      '@tanstack/react-query',
      'zustand',
      'zod',
    ],
  },
  build: {
    target: 'es2022',
    sourcemap: command === 'serve',
    rollupOptions: {
      output: {
        manualChunks: {
          'vendor-react': ['react', 'react-dom', 'react-dom/client'],
          'vendor-query': ['@tanstack/react-query'],
          'vendor-state': ['zustand', 'zod'],
        },
      },
    },
  },
}));
