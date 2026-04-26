import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';
import { nameOf } from '@/api/catalog';
import { useToast } from '@/ui/Toast';
import { useTranslation } from '@/i18n/i18n';

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
  const { t } = useTranslation('artMarketUi');
  const { t: ti } = useTranslation('info');
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
      toast.show('success', t('toastBoughtTitle'), t('toastBoughtBody'));
    },
    onError: (err) => { toast.show('danger', t('toastBuyErrTitle'), err instanceof Error ? err.message : ''); },
  });
  const cancel = useMutation({
    mutationFn: (offerID: string) => api.delete<void>(`/api/artefact-market/offers/${offerID}`),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['artefact-market'] });
      void qc.invalidateQueries({ queryKey: ['artefacts'] });
      toast.show('info', t('toastCancelledTitle'));
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
          {t('title')}
        </h2>
        <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
          <span style={{ fontSize: 14, color: 'var(--ox-fg-dim)' }}>{t('balanceLabel')}</span>
          <span style={{ fontFamily: 'var(--ox-mono)', fontWeight: 700, color: 'var(--ox-accent)', fontSize: 15 }}>
            {creditVal} cr
          </span>
        </div>
      </div>

      <div className="ox-tabs">
        <button type="button" aria-pressed={filter === 'all'} onClick={() => setFilter('all')}>
          {t('tabAll')} ({all.length})
        </button>
        <button type="button" aria-pressed={filter === 'mine'} onClick={() => setFilter('mine')}>
          {t('tabMine')} ({all.filter((o) => o.seller_user_id === myUserID).length})
        </button>
      </div>

      {shown.length === 0 ? (
        <div style={{ color: 'var(--ox-fg-dim)', fontSize: 16, padding: '8px 0' }}>
          {t('empty')}
        </div>
      ) : (
        <div className="ox-panel" style={{ overflow: 'hidden' }}>
          <div className="ox-table-responsive">
            <table className="ox-table" style={{ margin: 0 }}>
              <thead>
                <tr>
                  <th>{t('colArtefact')}</th>
                  <th>{t('colSeller')}</th>
                  <th>{t('colPrice')}</th>
                  <th>{t('colDate')}</th>
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
                      <td data-label={t('colArtefact')}>
                        <div>{nameOf(o.unit_id, ti)}</div>
                      </td>
                      <td data-label={t('colSeller')}>{o.seller_name ?? '—'}</td>
                      <td data-label={t('colPrice')} className="num" style={{ fontFamily: 'var(--ox-mono)', color: 'var(--ox-accent)' }}>
                        {o.price_credit} cr
                      </td>
                      <td data-label={t('colDate')} style={{ fontSize: 13, color: 'var(--ox-fg-muted)', fontFamily: 'var(--ox-mono)' }}>
                        {listedDate}
                      </td>
                      <td>
                        {isMine ? (
                          <button type="button" className="btn-ghost btn-sm" disabled={cancel.isPending} onClick={() => cancel.mutate(o.id)}>
                            {t('cancelBtn')}
                          </button>
                        ) : (
                          <button
                            type="button"
                            className={`btn btn-sm${!canAfford ? ' btn-ghost' : ' btn-success'}`}
                            disabled={buy.isPending || !canAfford}
                            title={!canAfford ? t('tooltipNoCredits') : undefined}
                            onClick={() => buy.mutate(o.id)}
                          >
                            {canAfford ? t('buyBtn') : t('buyBtnCantAfford')}
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
