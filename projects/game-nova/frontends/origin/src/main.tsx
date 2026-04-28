// Entry-point origin-фронта (план 72 Ф.1).
//
// Связывает QueryClientProvider + I18nProvider + StrictMode.
// Семантика идентична nova-main.tsx (тот же стек, те же
// настройки query). Отличия:
//   - Свой localStorage-namespace для языка ('oxsar-origin-lang'),
//     чтобы origin/nova могли иметь разный язык одновременно.
//   - Меньше boilerplate (нет error-boundary с reset — добавится
//     в Ф.7-Ф.9 при стабилизации, на Ф.1 не критично).

import React from 'react';
import ReactDOM from 'react-dom/client';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { App } from './App';
import { I18nProvider, type Lang } from '@/i18n/i18n';
import './styles/app.css';

const LANG_STORAGE_KEY = 'oxsar-origin-lang';

function loadLang(): Lang {
  const raw = localStorage.getItem(LANG_STORAGE_KEY);
  return raw === 'en' ? 'en' : 'ru';
}

const qc = new QueryClient({
  defaultOptions: {
    queries: { staleTime: 10_000, retry: 1, refetchOnWindowFocus: false },
  },
});

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
