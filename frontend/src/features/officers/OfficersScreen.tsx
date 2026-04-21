import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';
import { useTranslation } from '@/i18n/i18n';

// OfficersScreen — 4 временных officer'а с эффектами на факторы.
// Активация за credit, авто-ревет через event kind=62.

interface Entry {
  key: string;
  title: string;
  description: string;
  duration_days: number;
  cost_credit: number;
  activated_at?: string | null;
  expires_at?: string | null;
}

export function OfficersScreen() {
  const { t, tf } = useTranslation();
  const qc = useQueryClient();

  const officers = useQuery({
    queryKey: ['officers'],
    queryFn: () => api.get<{ officers: Entry[] | null }>('/api/officers'),
    refetchInterval: 15000,
  });
  const credit = useQuery({
    queryKey: ['artefact-market', 'credit'],
    queryFn: () => api.get<{ credit: number }>('/api/artefact-market/credit'),
    refetchInterval: 15000,
  });

  const activate = useMutation({
    mutationFn: (key: string) => api.post<Entry>(`/api/officers/${key}/activate`),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['officers'] });
      void qc.invalidateQueries({ queryKey: ['artefact-market'] });
      void qc.invalidateQueries({ queryKey: ['planets'] });
    },
  });

  const list = officers.data?.officers ?? [];
  const creditVal = credit.data?.credit ?? 0;

  return (
    <section>
      <h2>{tf('global', 'MENU_OFFICERS', 'Офицеры')}</h2>
      <p>
        <b>{tf('Main', 'CREDIT', 'Credit')}:</b> {creditVal}
      </p>

      {list.length === 0 ? (
        <p>{tf('Main', 'OFFICERS_EMPTY', 'Нет доступных офицеров.')}</p>
      ) : (
        <table className="ox-table">
          <thead>
            <tr>
              <th>{tf('Main', 'OFFICER', 'Офицер')}</th>
              <th>{tf('Main', 'OFF_DESC', 'Эффект')}</th>
              <th>{tf('Main', 'OFF_DURATION', 'Срок')}</th>
              <th>{tf('Main', 'OFF_COST', 'Цена')}</th>
              <th>{tf('Main', 'ACTION', 'Действие')}</th>
            </tr>
          </thead>
          <tbody>
            {list.map((e) => {
              const active = !!e.expires_at;
              return (
                <tr key={e.key} style={{ opacity: active ? 1 : 0.85 }}>
                  <td><b>{e.title}</b></td>
                  <td>{e.description}</td>
                  <td>{e.duration_days} {tf('Main', 'OFF_DAYS', 'дн.')}</td>
                  <td className="num">{e.cost_credit}</td>
                  <td>
                    {active ? (
                      <span>
                        {tf('Main', 'OFF_UNTIL', 'до')}{' '}
                        {new Date(e.expires_at!).toLocaleString('ru-RU')}
                      </span>
                    ) : (
                      <button
                        type="button"
                        disabled={activate.isPending || creditVal < e.cost_credit}
                        onClick={() => activate.mutate(e.key)}
                      >
                        {tf('Main', 'OFF_ACTIVATE', 'Активировать')}
                      </button>
                    )}
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      )}

      {activate.isError && (
        <div className="ox-error">
          {activate.error instanceof Error ? activate.error.message : t('global', 'ERROR')}
        </div>
      )}
    </section>
  );
}
