import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';
import { RESEARCH, nameOf } from '@/api/catalog';
import { useTranslation } from '@/i18n/i18n';
import type { Planet, QueueItem, ResearchState } from '@/api/types';

export function ResearchScreen({ planet }: { planet: Planet }) {
  const { t, tf } = useTranslation();
  const qc = useQueryClient();
  const state = useQuery({
    queryKey: ['research'],
    queryFn: () => api.get<ResearchState>('/api/research'),
    refetchInterval: 2000,
  });

  const enqueue = useMutation({
    mutationFn: (unitId: number) =>
      api.post<QueueItem>(`/api/planets/${planet.id}/research`, { unit_id: unitId }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['research'] });
      void qc.invalidateQueries({ queryKey: ['planets'] });
    },
  });

  const levels = state.data?.levels ?? {};
  const activeResearch = state.data?.queue[0];

  return (
    <section>
      <h2>{t('global', 'MENU_RESEARCH')}</h2>
      {activeResearch ? (
        <p>
          {tf('Main', 'RESEARCH_IN_PROGRESS', 'Идёт:')}{' '}
          <b>{nameOf(activeResearch.unit_id)}</b> → {tf('Main', 'LEVEL_SHORT', 'ур.')}{' '}
          {activeResearch.target_level}, {tf('Main', 'UNTIL', 'до')}{' '}
          {new Date(activeResearch.end_at).toLocaleTimeString('ru-RU')}
        </p>
      ) : (
        <p>{tf('Main', 'RESEARCH_IDLE', 'Свободно — можно запустить новое исследование.')}</p>
      )}

      <table className="ox-table">
        <thead>
          <tr>
            <th>{tf('Main', 'TECHNOLOGY', 'Технология')}</th>
            <th>{tf('Main', 'LEVEL', 'Уровень')}</th>
            <th>{tf('Main', 'ACTION', 'Действие')}</th>
          </tr>
        </thead>
        <tbody>
          {RESEARCH.map((r) => (
            <tr key={r.id}>
              <td>{r.name}</td>
              <td className="num">{levels[r.id.toString()] ?? 0}</td>
              <td>
                <button
                  type="button"
                  disabled={enqueue.isPending || !!activeResearch}
                  onClick={() => enqueue.mutate(r.id)}
                >
                  {tf('Main', 'RESEARCH_BUTTON', 'Исследовать')}
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>

      {enqueue.isError && (
        <div className="ox-error">
          {enqueue.error instanceof Error ? enqueue.error.message : t('global', 'ERROR')}
        </div>
      )}
    </section>
  );
}
