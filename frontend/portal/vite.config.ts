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
    proxy: {
      '/api': process.env.VITE_PORTAL_API ?? 'http://localhost:8090',
    },
  },
  build: {
    target: 'es2022',
    sourcemap: command === 'serve',
    outDir: 'dist',
  },
}));
