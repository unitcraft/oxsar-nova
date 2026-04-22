import { createContext, useCallback, useContext, useState } from 'react';

type ToastKind = 'info' | 'success' | 'warning' | 'danger';

interface ToastItem {
  id: number;
  kind: ToastKind;
  title: string;
  message?: string;
  persistent?: boolean;
}

interface ToastCtx {
  show: (kind: ToastKind, title: string, message?: string, persistent?: boolean) => void;
}

const ToastContext = createContext<ToastCtx>({ show: () => {} });

const ICONS: Record<ToastKind, string> = {
  info: 'ℹ️', success: '✅', warning: '⚠️', danger: '🚨',
};

let _id = 0;

export function ToastProvider({ children }: { children: React.ReactNode }) {
  const [toasts, setToasts] = useState<ToastItem[]>([]);

  const show = useCallback((kind: ToastKind, title: string, message?: string, persistent = false) => {
    const id = ++_id;
    setToasts((prev) => [...prev, { id, kind, title, message, persistent }]);
    if (!persistent) {
      setTimeout(() => setToasts((prev) => prev.filter((t) => t.id !== id)), 5000);
    }
  }, []);

  const dismiss = (id: number) => setToasts((prev) => prev.filter((t) => t.id !== id));

  return (
    <ToastContext.Provider value={{ show }}>
      {children}
      <div className="ox-toast-container">
        {toasts.map((t) => (
          <div key={t.id} className={`ox-toast ${t.kind}`} onClick={() => dismiss(t.id)}>
            <div className="ox-toast-icon">{ICONS[t.kind]}</div>
            <div className="ox-toast-body">
              <div className="ox-toast-title">{t.title}</div>
              {t.message && <div className="ox-toast-msg">{t.message}</div>}
            </div>
          </div>
        ))}
      </div>
    </ToastContext.Provider>
  );
}

export function useToast() {
  return useContext(ToastContext);
}
