import React from 'react';
import ReactDOM from 'react-dom/client';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { App } from './App';
import { I18nProvider, type Lang } from './i18n/i18n';
import { reportWebVitals, logWebVitals } from './lib/web-vitals';
import './styles/app.css';

class RootErrorBoundary extends React.Component<
  { children: React.ReactNode },
  { error: Error | null }
> {
  constructor(props: { children: React.ReactNode }) {
    super(props);
    this.state = { error: null };
  }
  static getDerivedStateFromError(error: Error) {
    return { error };
  }
  handleReset = () => {
    localStorage.clear();
    sessionStorage.clear();
    window.location.reload();
  };
  override render() {
    if (this.state.error) {
      return (
        <div style={{
          display: 'flex', flexDirection: 'column', alignItems: 'center',
          justifyContent: 'center', height: '100vh', gap: 16,
          fontFamily: 'monospace', color: '#ccc', background: '#0a0c10',
        }}>
          <div style={{ fontSize: 40 }}>⚠️</div>
          <div style={{ fontSize: 16, color: '#f66' }}>Ошибка приложения</div>
          <div style={{ fontSize: 12, color: '#666', maxWidth: 400, textAlign: 'center' }}>
            {this.state.error.message}
          </div>
          <button
            type="button"
            onClick={this.handleReset}
            style={{
              marginTop: 8, padding: '8px 20px', background: '#63d9ff22',
              border: '1px solid #63d9ff55', borderRadius: 6,
              color: '#63d9ff', cursor: 'pointer', fontSize: 14,
            }}
          >
            Сбросить кэш и перезагрузить
          </button>
        </div>
      );
    }
    return this.props.children;
  }
}

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
    <RootErrorBoundary>
    <QueryClientProvider client={qc}>
      <I18nProvider lang={loadLang()}>
        <App />
      </I18nProvider>
    </QueryClientProvider>
    </RootErrorBoundary>
  </React.StrictMode>,
);

// Отслеживание Web Vitals метрик (только в development)
declare const __DEV__: boolean;
if (typeof __DEV__ !== 'undefined' && __DEV__) {
  reportWebVitals(logWebVitals);
}
