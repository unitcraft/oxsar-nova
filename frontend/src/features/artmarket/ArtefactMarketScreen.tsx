import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';
import { nameOf } from '@/api/catalog';
import { useToast } from '@/ui/Toast';

interface Offer {
  id: string;
  artefact_id: string;
  seller_user_id: string;
  seller_name?: string;
  unit_id: number;
  price_credit: number;
  listed_at: string;
}

export function ArtefactMarketScreen() {
  const qc = useQueryClient();
  const toast = useToast();
  const [filter, setFilter] = useState<'all' | 'mine'>('all');

  const me = useQuery({
    queryKey: ['me'],
    queryFn: () => api.get<{ user_id: string; username: string }>('/api/me'),
    staleTime: Infinity,
  });
  const offers = useQuery({
    queryKey: ['artefact-market', 'offers'],
    queryFn: () => api.get<{ offers: Offer[] | null }>('/api/artefact-market/offers'),
    refetchInterval: 10000,
  });
  const credit = useQuery({
    queryKey: ['artefact-market', 'credit'],
    queryFn: () => api.get<{ credit: number }>('/api/artefact-market/credit'),
    refetchInterval: 10000,
  });

  const buy = useMutation({
    mutationFn: (offerID: string) => api.post<void>(`/api/artefact-market/offers/${offerID}/buy`),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['artefact-market'] });
      void qc.invalidateQueries({ queryKey: ['artefacts'] });
      toast.show('success', 'Куплено', 'Артефакт добавлен в инвентарь');
    },
    onError: (err) => { toast.show('danger', 'Ошибка покупки', err instanceof Error ? err.message : ''); },
  });
  const cancel = useMutation({
    mutationFn: (offerID: string) => api.delete<void>(`/api/artefact-market/offers/${offerID}`),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['artefact-market'] });
      void qc.invalidateQueries({ queryKey: ['artefacts'] });
      toast.show('info', 'Оффер отменён');
    },
  });

  const all = offers.data?.offers ?? [];
  const myUserID = me.data?.user_id;
  const creditVal = credit.data?.credit ?? 0;
  const shown = filter === 'all' ? all : all.filter((o) => o.seller_user_id === myUserID);

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: 8 }}>
        <h2 style={{ margin: 0, fontSize: 18, fontFamily: 'var(--ox-font)', fontWeight: 700 }}>
          🏷 Рынок артефактов
        </h2>
        <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
          <span style={{ fontSize: 12, color: 'var(--ox-fg-dim)' }}>Баланс:</span>
          <span style={{ fontFamily: 'var(--ox-mono)', fontWeight: 700, color: 'var(--ox-accent)', fontSize: 15 }}>
            {creditVal} cr
          </span>
        </div>
      </div>

      <div className="ox-tabs">
        <button type="button" aria-pressed={filter === 'all'} onClick={() => setFilter('all')}>
          📋 Все офферы ({all.length})
        </button>
        <button type="button" aria-pressed={filter === 'mine'} onClick={() => setFilter('mine')}>
          👤 Мои ({all.filter((o) => o.seller_user_id === myUserID).length})
        </button>
      </div>

      {shown.length === 0 ? (
        <div style={{ color: 'var(--ox-fg-dim)', fontSize: 14, padding: '8px 0' }}>
          Нет офферов.
        </div>
      ) : (
        <div className="ox-panel" style={{ overflow: 'hidden' }}>
          <div className="ox-table-responsive">
            <table className="ox-table" style={{ margin: 0 }}>
              <thead>
                <tr>
                  <th>Артефакт</th>
                  <th>Продавец</th>
                  <th>Цена</th>
                  <th>Дата</th>
                  <th />
                </tr>
              </thead>
              <tbody>
                {shown.map((o) => {
                  const isMine = o.seller_user_id === myUserID;
                  const canAfford = creditVal >= o.price_credit;
                  const listedDate = new Date(o.listed_at).toLocaleDateString('ru-RU', { day: '2-digit', month: '2-digit' });
                  return (
                    <tr key={o.id}>
                      <td data-label="Артефакт">{nameOf(o.unit_id)}</td>
                      <td data-label="Продавец">{o.seller_name ?? '—'}</td>
                      <td data-label="Цена" className="num" style={{ fontFamily: 'var(--ox-mono)', color: 'var(--ox-accent)' }}>
                        {o.price_credit} cr
                      </td>
                      <td data-label="Дата" style={{ fontSize: 11, color: 'var(--ox-fg-muted)', fontFamily: 'var(--ox-mono)' }}>
                        {listedDate}
                      </td>
                      <td>
                        {isMine ? (
                          <button type="button" className="btn-ghost btn-sm" disabled={cancel.isPending} onClick={() => cancel.mutate(o.id)}>
                            Отменить
                          </button>
                        ) : (
                          <button
                            type="button"
                            className={`btn btn-sm${!canAfford ? ' btn-ghost' : ' btn-success'}`}
                            disabled={buy.isPending || !canAfford}
                            title={!canAfford ? 'Недостаточно кредитов' : undefined}
                            onClick={() => buy.mutate(o.id)}
                          >
                            {canAfford ? 'Купить' : 'Мало cr'}
                          </button>
                        )}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  );
}
