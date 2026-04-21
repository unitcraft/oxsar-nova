import { useQuery } from '@tanstack/react-query';
import { api } from '@/api/client';
import { useTranslation } from '@/i18n/i18n';

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
  const { tf } = useTranslation();
  const q = useQuery({
    queryKey: ['tutorial'],
    queryFn: () => api.get<TutorialStatus>('/api/tutorial'),
    refetchInterval: 10000,
  });

  if (q.isLoading) return <p>…</p>;
  if (q.error || !q.data) return <p className="ox-error">error</p>;

  const { steps, complete } = q.data;
  const doneCount = steps.filter((s) => s.done).length;

  return (
    <section>
      <h2>{tf('global', 'MENU_TUTORIAL', 'Туториал')}</h2>

      {complete ? (
        <p style={{ color: 'var(--ox-ok, green)', fontWeight: 600 }}>
          {tf('Main', 'TUTORIAL_COMPLETE', '🎉 Туториал завершён! Вы получили все начальные награды.')}
        </p>
      ) : (
        <p>
          {tf('Main', 'TUTORIAL_PROGRESS', 'Выполнено')}:{' '}
          <b>{doneCount} / {steps.length}</b>
          {' · '}
          {tf('Main', 'TUTORIAL_REWARD_HINT', 'Каждый шаг даёт +10 кредитов.')}
        </p>
      )}

      <ol style={{ paddingLeft: 20 }}>
        {steps.map((step) => (
          <li
            key={step.index}
            style={{
              marginBottom: 12,
              opacity: step.done ? 0.6 : 1,
            }}
          >
            <span style={{ fontWeight: step.done ? 400 : 600 }}>
              {step.done ? '✓ ' : '○ '}
              {step.title}
            </span>
            <br />
            <small style={{ color: 'var(--ox-muted, #888)' }}>{step.description}</small>
          </li>
        ))}
      </ol>
    </section>
  );
}
