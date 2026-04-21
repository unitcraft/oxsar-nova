import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';
import { useTranslation } from '@/i18n/i18n';
import type { Planet } from '@/api/types';

// MarketScreen — простой обменник metal ↔ silicon ↔ hydrogen.
// Стоимости фиксированы (M=1, Si=2, H=4), per-user exchange_rate
// применяется сервером: чем выше rate, тем меньше получает юзер.

interface Rates {
  metal: number;
  silicon: number;
  hydrogen: number;
  user_rate: number;
}

interface ExchangeResult {
  from: string;
  to: string;
  from_amount: number;
  to_amount: number;
  rate: number;
}

type Res = 'metal' | 'silicon' | 'hydrogen';

export function MarketScreen({ planet }: { planet: Planet }) {
  const { t, tf } = useTranslation();
  const qc = useQueryClient();

  const rates = useQuery({
    queryKey: ['market', 'rates'],
    queryFn: () => api.get<Rates>('/api/market/rates'),
    staleTime: 60_000,
  });

  const [from, setFrom] = useState<Res>('metal');
  const [to, setTo] = useState<Res>('silicon');
  const [amount, setAmount] = useState(1000);
  const [last, setLast] = useState<ExchangeResult | null>(null);

  const exchange = useMutation({
    mutationFn: () =>
      api.post<ExchangeResult>(`/api/planets/${planet.id}/market/exchange`, {
        from,
        to,
        amount,
      }),
    onSuccess: (res) => {
      setLast(res);
      void qc.invalidateQueries({ queryKey: ['planets'] });
    },
  });

  // Предварительная оценка: стоимость * user_rate применяется в backend.
  const preview = rates.data
    ? Math.floor((amount * rates.data[from]) / rates.data[to] / rates.data.user_rate)
    : null;

  return (
    <section>
      <h2>{tf('global', 'MENU_MARKET', 'Рынок')} — {planet.name}</h2>
      <p>
        {tf(
          'Main',
          'MARKET_HINT',
          'Курс M:Si:H = 1:2:4. Чем выше ваш personal rate, тем меньше получаете.',
        )}
      </p>

      {rates.data && (
        <p>
          <b>{tf('Main', 'MARKET_YOUR_RATE', 'Ваш курс')}:</b>{' '}
          {rates.data.user_rate.toFixed(2)}
        </p>
      )}

      <div style={{ display: 'flex', gap: 12, alignItems: 'center', marginBottom: 12 }}>
        <label>
          {tf('Main', 'MARKET_FROM', 'Отдать')}:
          <select value={from} onChange={(e) => setFrom(e.target.value as Res)} style={{ marginLeft: 8 }}>
            <option value="metal">{t('global', 'METAL')}</option>
            <option value="silicon">{t('global', 'SILICON')}</option>
            <option value="hydrogen">{t('global', 'HYDROGEN')}</option>
          </select>
        </label>
        <label>
          {tf('Main', 'MARKET_TO', 'Получить')}:
          <select value={to} onChange={(e) => setTo(e.target.value as Res)} style={{ marginLeft: 8 }}>
            <option value="metal">{t('global', 'METAL')}</option>
            <option value="silicon">{t('global', 'SILICON')}</option>
            <option value="hydrogen">{t('global', 'HYDROGEN')}</option>
          </select>
        </label>
        <label>
          {tf('Main', 'MARKET_AMOUNT', 'Количество')}:
          <input
            type="number"
            min={1}
            value={amount}
            onChange={(e) => setAmount(Math.max(1, Number(e.target.value)))}
            style={{ width: 120, marginLeft: 8 }}
          />
        </label>
      </div>

      {preview !== null && from !== to && (
        <p>
          {tf('Main', 'MARKET_YOU_GET', 'Вы получите')}: <b>{preview}</b>
        </p>
      )}

      <button
        type="button"
        disabled={exchange.isPending || from === to || amount <= 0}
        onClick={() => exchange.mutate()}
      >
        {exchange.isPending ? '…' : tf('Main', 'MARKET_EXCHANGE', 'Обменять')}
      </button>

      {exchange.isError && (
        <div className="ox-error">
          {exchange.error instanceof Error ? exchange.error.message : t('global', 'ERROR')}
        </div>
      )}

      {last && (
        <div style={{ marginTop: 12 }}>
          <b>{tf('Main', 'MARKET_LAST', 'Последний обмен')}:</b>{' '}
          {last.from_amount} {last.from} → {last.to_amount} {last.to}{' '}
          (rate {last.rate.toFixed(4)})
        </div>
      )}
    </section>
  );
}
