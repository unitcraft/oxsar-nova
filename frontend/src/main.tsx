import React from 'react';
import ReactDOM from 'react-dom/client';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { App } from './App';
import { I18nProvider, type Lang } from './i18n/i18n';
import './styles/app.css';

const qc = new QueryClient({
  defaultOptions: {
    queries: { staleTime: 10_000, retry: 1, refetchOnWindowFocus: false },
  },
});

// Язык пользователя: сохраняется в localStorage для стабильности
// между сессиями. Валидируем значение, чтобы подменённый ключ не
// привёл к непредсказуемому поведению.
function loadLang(): Lang {
  const raw = localStorage.getItem('oxsar-lang');
  return raw === 'en' ? 'en' : 'ru';
}

const root = document.getElementById('root');
if (!root) throw new Error('root element not found');

ReactDOM.createRoot(root).render(
  <React.StrictMode>
    <QueryClientProvider client={qc}>
      <I18nProvider lang={loadLang()}>
        <App />
      </I18nProvider>
    </QueryClientProvider>
  </React.StrictMode>,
);
