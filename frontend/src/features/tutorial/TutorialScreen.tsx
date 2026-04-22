import { useQuery } from '@tanstack/react-query';
import { api } from '@/api/client';
import { ProgressBar } from '@/ui/ProgressBar';

interface TutorialStep {
  index: number;
  title: string;
  description: string;
  done: boolean;
}

interface TutorialStatus {
  steps: TutorialStep[];
  state: number;
  complete: boolean;
}

const STEP_RESOURCES: [number, number, number][] = [
  [500,  200,  0],
  [300,  300,  100],
  [500,  500,  200],
  [1000, 500,  300],
  [2000, 1000, 500],
  [5000, 3000, 1000],
];

export function TutorialScreen() {
  const q = useQuery({
    queryKey: ['tutorial'],
    queryFn: () => api.get<TutorialStatus>('/api/tutorial'),
    refetchInterval: 10000,
  });

  if (q.isLoading) {
    return <div style={{ padding: 24 }}><div className="ox-skeleton" style={{ height: 200 }} /></div>;
  }
  if (q.error || !q.data) {
    return <div className="ox-alert ox-alert-danger">Ошибка загрузки туториала</div>;
  }

  const { steps, complete } = q.data;
  const doneCount = steps.filter((s) => s.done).length;
  const pct = steps.length > 0 ? (doneCount / steps.length) * 100 : 0;

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: 8 }}>
        <h2 style={{ margin: 0, fontSize: 18, fontFamily: 'var(--ox-font)', fontWeight: 700 }}>
          📖 Туториал
        </h2>
        <span style={{ fontSize: 13, color: 'var(--ox-fg-dim)', fontFamily: 'var(--ox-mono)' }}>
          {doneCount}/{steps.length}
        </span>
      </div>

      {complete ? (
        <div className="ox-panel" style={{ padding: 20, textAlign: 'center' }}>
          <div style={{ fontSize: 32, marginBottom: 8 }}>🎉</div>
          <div style={{ fontWeight: 700, fontSize: 15, color: 'var(--ox-success)' }}>
            Туториал завершён! Вы получили все начальные награды.
          </div>
        </div>
      ) : (
        <div className="ox-panel" style={{ padding: '12px 16px' }}>
          <ProgressBar pct={pct} variant="success" height={6} showLabel />
        </div>
      )}

      <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
        {steps.map((step) => (
          <div
            key={step.index}
            className="ox-panel"
            style={{
              padding: '12px 16px',
              display: 'flex', alignItems: 'flex-start', gap: 12,
              opacity: step.done ? 0.65 : 1,
            }}
          >
            <div style={{ fontSize: 22, flexShrink: 0, marginTop: 1 }}>
              {step.done ? '✅' : '○'}
            </div>
            <div style={{ flex: 1 }}>
              <div style={{ fontWeight: step.done ? 400 : 600, fontSize: 14, marginBottom: 2 }}>
                {step.title}
              </div>
              <div style={{ fontSize: 12, color: 'var(--ox-fg-dim)', marginBottom: 4 }}>
                {step.description}
              </div>
              {!step.done && (() => {
                const res = STEP_RESOURCES[step.index - 1];
                if (!res) return null;
                const [m, si, h] = res;
                return (
                  <div style={{ fontSize: 11, fontFamily: 'var(--ox-mono)', color: 'var(--ox-success)', display: 'flex', gap: 8, flexWrap: 'wrap' }}>
                    <span>💳 +10 cr</span>
                    {m > 0  && <span>⛏ +{m.toLocaleString('ru-RU')}</span>}
                    {si > 0 && <span>🔷 +{si.toLocaleString('ru-RU')}</span>}
                    {h > 0  && <span>💧 +{h.toLocaleString('ru-RU')}</span>}
                  </div>
                );
              })()}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
