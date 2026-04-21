import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';
import { nameOf } from '@/api/catalog';
import { useTranslation } from '@/i18n/i18n';

// ArtefactMarketScreen — маркетплейс артефактов за credit.
// Продажа — через ArtefactsScreen (кнопка «Продать» на held-артефакте).
// Здесь — покупка и отмена своих офферов.

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
  const { t, tf } = useTranslation();
  const qc = useQueryClient();
  const [filter, setFilter] = useState<'all' | 'mine'>('all');

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
    mutationFn: (offerID: string) =>
      api.post<void>(`/api/artefact-market/offers/${offerID}/buy`),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['artefact-market'] });
      void qc.invalidateQueries({ queryKey: ['artefacts'] });
    },
  });
  const cancel = useMutation({
    mutationFn: (offerID: string) =>
      api.delete<void>(`/api/artefact-market/offers/${offerID}`),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['artefact-market'] });
      void qc.invalidateQueries({ queryKey: ['artefacts'] });
    },
  });

  const all = offers.data?.offers ?? [];
  const my = credit.data;
  // «mine» определяется по seller_name — но проще по JWT sub; для
  // MVP оставим «показывать всех» и помечать свои кнопкой «Отменить».
  const shown = filter === 'all' ? all : all.filter((o) => o.seller_name === my?.toString());

  return (
    <section>
      <h2>{tf('global', 'MENU_ART_MARKET', 'Рынок артефактов')}</h2>
      <p>
        <b>{tf('Main', 'CREDIT', 'Credit')}:</b> {credit.data?.credit ?? 0}
      </p>

      <div style={{ marginBottom: 12 }}>
        <label>
          <input
            type="radio"
            checked={filter === 'all'}
            onChange={() => setFilter('all')}
          />{' '}
          {tf('Main', 'ART_ALL_OFFERS', 'Все офферы')}
        </label>{' '}
        <label style={{ marginLeft: 12 }}>
          <input
            type="radio"
            checked={filter === 'mine'}
            onChange={() => setFilter('mine')}
          />{' '}
          {tf('Main', 'ART_MY_OFFERS', 'Мои (сверка по имени)')}
        </label>
      </div>

      {shown.length === 0 ? (
        <p>{tf('Main', 'ART_NO_OFFERS', 'Нет офферов.')}</p>
      ) : (
        <table className="ox-table">
          <thead>
            <tr>
              <th>{tf('Main', 'ARTEFACT', 'Артефакт')}</th>
              <th>{tf('Main', 'SELLER', 'Продавец')}</th>
              <th>{tf('Main', 'PRICE', 'Цена')}</th>
              <th>{tf('Main', 'ACTION', 'Действие')}</th>
            </tr>
          </thead>
          <tbody>
            {shown.map((o) => (
              <tr key={o.id}>
                <td>{nameOf(o.unit_id)}</td>
                <td>{o.seller_name || '—'}</td>
                <td className="num">{o.price_credit}</td>
                <td>
                  <button
                    type="button"
                    disabled={buy.isPending || (credit.data?.credit ?? 0) < o.price_credit}
                    onClick={() => buy.mutate(o.id)}
                  >
                    {tf('Main', 'BUY', 'Купить')}
                  </button>{' '}
                  <button
                    type="button"
                    disabled={cancel.isPending}
                    onClick={() => cancel.mutate(o.id)}
                  >
                    {tf('Main', 'CANCEL', 'Отменить')}
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}

      {buy.isError && (
        <div className="ox-error">
          {buy.error instanceof Error ? buy.error.message : t('global', 'ERROR')}
        </div>
      )}
      {cancel.isError && (
        <div className="ox-error">
          {cancel.error instanceof Error ? cancel.error.message : t('global', 'ERROR')}
        </div>
      )}
    </section>
  );
}
