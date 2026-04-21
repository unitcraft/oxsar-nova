import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';
import { nameOf } from '@/api/catalog';
import { useTranslation } from '@/i18n/i18n';
import type { Artefact } from '@/api/types';

export function ArtefactsScreen() {
  const { t, tf } = useTranslation();
  const qc = useQueryClient();
  const list = useQuery({
    queryKey: ['artefacts'],
    queryFn: () => api.get<{ artefacts: Artefact[] | null }>('/api/artefacts'),
    refetchInterval: 5000,
  });

  const activate = useMutation({
    mutationFn: (id: string) => api.post<Artefact>(`/api/artefacts/${id}/activate`),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['artefacts'] });
      void qc.invalidateQueries({ queryKey: ['planets'] });
    },
  });
  const deactivate = useMutation({
    mutationFn: (id: string) => api.delete<void>(`/api/artefacts/${id}/deactivate`),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['artefacts'] });
      void qc.invalidateQueries({ queryKey: ['planets'] });
    },
  });

  if (list.isLoading) return <p>{tf('Main', 'ARTEFACTS_LOADING', 'Загрузка артефактов…')}</p>;
  if (list.error)
    return (
      <p className="ox-error">
        {t('global', 'ERROR')}: {list.error instanceof Error ? list.error.message : ''}
      </p>
    );

  const items = list.data?.artefacts ?? [];
  if (items.length === 0) {
    return (
      <section>
        <h2>{t('global', 'MENU_ARTEFACTS')}</h2>
        <p>
          {tf(
            'Main',
            'ARTEFACTS_EMPTY',
            'Инвентарь пуст. Артефакты появляются как награда за бой/экспедицию, покупаются в Artefact Market за credit (M5.1) или выдаются админом.',
          )}
        </p>
      </section>
    );
  }

  const actionLabel = (a: Artefact): string =>
    a.state === 'active'
      ? tf('Main', 'ARTEFACT_DEACTIVATE', 'Деактивировать')
      : a.state === 'held'
        ? tf('Main', 'ARTEFACT_ACTIVATE', 'Активировать')
        : '—';

  return (
    <section>
      <h2>{t('global', 'MENU_ARTEFACTS')}</h2>
      <table className="ox-table">
        <thead>
          <tr>
            <th>{tf('Main', 'ARTEFACT', 'Артефакт')}</th>
            <th>{tf('Main', 'STATE', 'Состояние')}</th>
            <th>{tf('Main', 'EXPIRES_AT', 'Истекает')}</th>
            <th>{tf('Main', 'ACTION', 'Действие')}</th>
          </tr>
        </thead>
        <tbody>
          {items.map((a) => (
            <tr key={a.id}>
              <td>{nameOf(a.unit_id)}</td>
              <td>{a.state}</td>
              <td>{a.expire_at ? new Date(a.expire_at).toLocaleString('ru-RU') : '—'}</td>
              <td>
                {a.state === 'held' && (
                  <button
                    type="button"
                    disabled={activate.isPending}
                    onClick={() => activate.mutate(a.id)}
                  >
                    {actionLabel(a)}
                  </button>
                )}
                {a.state === 'active' && (
                  <button
                    type="button"
                    disabled={deactivate.isPending}
                    onClick={() => deactivate.mutate(a.id)}
                  >
                    {actionLabel(a)}
                  </button>
                )}
                {a.state !== 'held' && a.state !== 'active' && <span>—</span>}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
      {activate.isError && (
        <div className="ox-error">
          {activate.error instanceof Error
            ? activate.error.message
            : tf('Main', 'ARTEFACT_ACTIVATE_ERROR', 'ошибка активации')}
        </div>
      )}
      {deactivate.isError && (
        <div className="ox-error">
          {deactivate.error instanceof Error
            ? deactivate.error.message
            : tf('Main', 'ARTEFACT_DEACTIVATE_ERROR', 'ошибка деактивации')}
        </div>
      )}
    </section>
  );
}
