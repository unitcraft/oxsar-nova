// Vite конфиг frontend'а oxsar-nova. Оптимизирован под частые
// пересборки при портировании legacy-экранов.
//
// Ключевые моменты:
//   - optimizeDeps.include: Vite пре-бандлит эти пакеты один раз при
//     старте и кеширует в node_modules/.vite. Без явного списка
//     Vite «открывает» зависимости лениво — каждый новый импорт в
//     коде вызывает holdup в dev-сервере на первый запрос экрана.
//   - server.warmup: прогреваем граф ключевых экранов до того, как
//     пользователь их откроет. На i7/SSD экономит ~300–800 мс
//     первого рендера каждого экрана.
//   - manualChunks в prod: vendor-разбивка. Изменения в домене
//     (feature) не меняют vendor-чанк, браузер берёт его из кеша.
//   - build.target: es2022 (у нас tsconfig ES2022), не тратим время
//     на транспиляцию в более ранние версии.
//
// Что НЕ делаем:
//   - sourcemap=false в dev — нужен для отладки.
//   - minify в dev — Vite и так не минифицирует; оставляем esbuild
//     по умолчанию для prod (быстрее terser в 20 раз).

import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { fileURLToPath, URL } from 'node:url';

export default defineConfig(({ command }) => ({
  plugins: [react()],
  resolve: {
    alias: { '@': fileURLToPath(new URL('./src', import.meta.url)) },
  },
  server: {
    port: 5173,
    // Разрешаем запросы с любого host — в docker network фронт дёргают
    // по service-имени (http://frontend:5173), а Vite 5 по умолчанию
    // отбивает с 403 всё, что не localhost.
    host: true,
    allowedHosts: true,
    watch: { usePolling: true, interval: 500 },
    // В docker'е backend живёт по имени сервиса (backend:8080), а
    // локально — на localhost:8080. Читаем из env, чтобы один и тот
    // же код работал в обоих режимах.
    proxy: {
      '/api': process.env.VITE_PROXY_TARGET ?? 'http://localhost:8080',
      '/healthz': process.env.VITE_PROXY_TARGET ?? 'http://localhost:8080',
    },
    // warmup — Vite начнёт трансформировать эти модули сразу после
    // старта сервера, не ждать первого запроса. Первые попадания на
    // overview/buildings/research стали «мгновенными».
    warmup: {
      clientFiles: [
        './src/main.tsx',
        './src/App.tsx',
        './src/i18n/i18n.tsx',
        './src/features/auth/LoginScreen.tsx',
        './src/features/overview/OverviewScreen.tsx',
      ],
    },
  },
  optimizeDeps: {
    // Список pre-bundled пакетов. Если добавляем новую крупную
    // dependency — дописать сюда, иначе первый импорт вызовет
    // принудительный rebundle (видно в консоли: «new dependencies
    // optimized»).
    include: [
      'react',
      'react-dom',
      'react-dom/client',
      '@tanstack/react-query',
      '@tanstack/react-router',
      'zustand',
      'zod',
    ],
  },
  build: {
    target: 'es2022',
    sourcemap: command === 'serve',
    rollupOptions: {
      output: {
        // Ручная нарезка vendor-чанков. Цель — сохранить их хеш
        // стабильным между итерациями доработки. UI меняется —
        // vendor не меняется — браузер не перекачивает 300kb
        // react+query+router при каждом релизе dev-билда.
        manualChunks: {
          'vendor-react': ['react', 'react-dom', 'react-dom/client'],
          'vendor-query': ['@tanstack/react-query', '@tanstack/react-router'],
          'vendor-state': ['zustand', 'zod'],
        },
      },
    },
  },
}));
