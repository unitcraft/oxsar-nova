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
          <div style={{ fontSize: 12, color: 'var(--ox-fg-dim)', marginBottom: 6 }}>
            Каждый шаг даёт +10 кредитов.
          </div>
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
            <div>
              <div style={{ fontWeight: step.done ? 400 : 600, fontSize: 14, marginBottom: 2 }}>
                {step.title}
              </div>
              <div style={{ fontSize: 12, color: 'var(--ox-fg-dim)' }}>
                {step.description}
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
