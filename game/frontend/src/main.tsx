import React from 'react';
import ReactDOM from 'react-dom/client';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { App } from './App';
import { I18nProvider, type Lang } from '@/i18n/i18n';
import { reportWebVitals, logWebVitals } from './lib/web-vitals';
import './styles/app.css';

class RootErrorBoundary extends React.Component<
  { children: React.ReactNode },
  { error: Error | null; componentStack: string | null }
> {
  constructor(props: { children: React.ReactNode }) {
    super(props);
    this.state = { error: null, componentStack: null };
  }
  static getDerivedStateFromError(error: Error) {
    return { error };
  }
  override componentDidCatch(_error: Error, info: React.ErrorInfo) {
    const stack = info.componentStack ?? '';
    this.setState({ componentStack: stack });
    // Вывод в заголовок вкладки для быстрой идентификации
    const match = stack.match(/at (\w+Screen|\w+Component|\w+)/);
    if (match) document.title = `💥 ${match[1]}`;
    console.error('[ErrorBoundary] component stack:\n' + stack);
  }
  handleReset = () => {
    localStorage.clear();
    sessionStorage.clear();
    window.location.reload();
  };
  override render() {
    if (this.state.error) {
      // Extract first meaningful component name from stack
      const stack = this.state.componentStack ?? '';
      const match = stack.match(/^\s*at (\w+)/m);
      const where = match ? match[1] : null;
      return (
        <div style={{
          display: 'flex', flexDirection: 'column', alignItems: 'center',
          justifyContent: 'center', height: '100vh', gap: 16,
          fontFamily: 'monospace', color: '#ccc', background: '#0a0c10',
        }}>
          <div style={{ fontSize: 40 }}>⚠️</div>
          <div style={{ fontSize: 16, color: '#f66' }}>App error / Ошибка приложения</div>
          {where && (
            <div style={{ fontSize: 15, color: '#fa0', fontWeight: 700 }}>
              Component / Компонент: {where}
            </div>
          )}
          <div style={{ fontSize: 14, color: '#aaa', maxWidth: 500, textAlign: 'center' }}>
            {this.state.error.message}
          </div>
          <pre style={{
            fontSize: 10, color: '#555', maxWidth: 600, maxHeight: 160,
            overflow: 'auto', textAlign: 'left', whiteSpace: 'pre-wrap',
            background: '#111', padding: '8px 12px', borderRadius: 6,
          }}>
            {stack.trim().split('\n').slice(0, 12).join('\n')}
          </pre>
          <button
            type="button"
            onClick={this.handleReset}
            style={{
              marginTop: 8, padding: '8px 20px', background: '#63d9ff22',
              border: '1px solid #63d9ff55', borderRadius: 6,
              color: '#63d9ff', cursor: 'pointer', fontSize: 16,
            }}
          >
            Reset & Reload
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
