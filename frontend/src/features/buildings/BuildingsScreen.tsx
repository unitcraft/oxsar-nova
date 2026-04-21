import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';
import { BUILDINGS, nameOf } from '@/api/catalog';
import { useTranslation } from '@/i18n/i18n';
import type { Planet, QueueItem } from '@/api/types';

export function BuildingsScreen({ planet }: { planet: Planet }) {
  const { t, tf } = useTranslation();
  const qc = useQueryClient();
  const queue = useQuery({
    queryKey: ['buildings-queue', planet.id],
    queryFn: () => api.get<{ queue: QueueItem[] }>(`/api/planets/${planet.id}/buildings/queue`),
    refetchInterval: 2000,
  });

  const enqueue = useMutation({
    mutationFn: (unitId: number) =>
      api.post<QueueItem>(`/api/planets/${planet.id}/buildings`, { unit_id: unitId }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['buildings-queue', planet.id] });
      void qc.invalidateQueries({ queryKey: ['planets'] });
    },
  });

  return (
    <section>
      <h2>
        {t('global', 'MENU_CONSTRUCTIONS')} — {planet.name}
      </h2>
      <table className="ox-table">
        <thead>
          <tr>
            <th>{t('global', 'MENU_CONSTRUCTIONS')}</th>
            <th>{tf('Main', 'ACTION', 'Действие')}</th>
          </tr>
        </thead>
        <tbody>
          {BUILDINGS.map((b) => (
            <tr key={b.id}>
              <td>{b.name}</td>
              <td>
                <button
                  type="button"
                  disabled={enqueue.isPending}
                  onClick={() => enqueue.mutate(b.id)}
                >
                  {tf('Main', 'BUILD_BUTTON', 'Построить')}
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

      <h3>{tf('Main', 'QUEUE_HEADER', 'Очередь')}</h3>
      {queue.data && queue.data.queue.length > 0 ? (
        <ul>
          {queue.data.queue.map((q) => (
            <li key={q.id}>
              {nameOf(q.unit_id)} → {tf('Main', 'LEVEL_SHORT', 'ур.')} {q.target_level},{' '}
              {tf('Main', 'UNTIL', 'до')} {new Date(q.end_at).toLocaleTimeString('ru-RU')}
            </li>
          ))}
        </ul>
      ) : (
        <p>{tf('Main', 'QUEUE_EMPTY', 'Очередь пуста.')}</p>
      )}
    </section>
  );
}
