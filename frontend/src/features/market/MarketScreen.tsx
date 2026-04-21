import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';
import { useTranslation } from '@/i18n/i18n';
import type { Planet } from '@/api/types';

interface Lot {
  id: string;
  seller_id: string;
  seller_name: string;
  sell_resource: string;
  sell_amount: number;
  buy_resource: string;
  buy_amount: number;
  state: string;
  created_at: string;
}

// MarketScreen — обменник metal ↔ silicon ↔ hydrogen + ордерная книга (лоты).

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

function LotsPanel({ planet }: { planet: Planet }) {
  const qc = useQueryClient();
  const [sellRes, setSellRes] = useState<Res>('metal');
  const [sellAmt, setSellAmt] = useState(1000);
  const [buyRes, setBuyRes] = useState<Res>('silicon');
  const [buyAmt, setBuyAmt] = useState(500);

  const lots = useQuery({
    queryKey: ['market-lots'],
    queryFn: () => api.get<Lot[]>('/api/market/lots'),
    refetchInterval: 15_000,
  });

  const create = useMutation({
    mutationFn: () =>
      api.post('/api/market/lots', {
        planet_id: planet.id,
        sell_resource: sellRes,
        sell_amount: sellAmt,
        buy_resource: buyRes,
        buy_amount: buyAmt,
      }),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['market-lots'] }),
  });

  const cancel = useMutation({
    mutationFn: (id: string) => api.delete(`/api/market/lots/${id}`),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['market-lots'] }),
  });

  const accept = useMutation({
    mutationFn: (id: string) =>
      api.post(`/api/market/lots/${id}/accept`, { planet_id: planet.id }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['market-lots'] });
      void qc.invalidateQueries({ queryKey: ['planets'] });
    },
  });

  const userId = planet.user_id;

  return (
    <section style={{ marginTop: 24 }}>
      <h3>Ордерная книга</h3>
      <div style={{ display: 'flex', gap: 8, alignItems: 'center', flexWrap: 'wrap', marginBottom: 8 }}>
        <span>Продать:</span>
        <input type="number" min={1} value={sellAmt} onChange={(e) => setSellAmt(Math.max(1, Number(e.target.value)))} style={{ width: 100 }} />
        <select value={sellRes} onChange={(e) => setSellRes(e.target.value as Res)}>
          <option value="metal">Металл</option>
          <option value="silicon">Кремний</option>
          <option value="hydrogen">Водород</option>
        </select>
        <span>за:</span>
        <input type="number" min={1} value={buyAmt} onChange={(e) => setBuyAmt(Math.max(1, Number(e.target.value)))} style={{ width: 100 }} />
        <select value={buyRes} onChange={(e) => setBuyRes(e.target.value as Res)}>
          <option value="metal">Металл</option>
          <option value="silicon">Кремний</option>
          <option value="hydrogen">Водород</option>
        </select>
        <button onClick={() => create.mutate()} disabled={create.isPending || sellRes === buyRes}>
          {create.isPending ? '…' : 'Выставить лот'}
        </button>
      </div>
      {create.isError && <div style={{ color: 'red' }}>{String(create.error)}</div>}

      <table style={{ width: '100%', borderCollapse: 'collapse', marginTop: 8 }}>
        <thead>
          <tr>
            <th>Продавец</th><th>Продаёт</th><th>За</th><th></th>
          </tr>
        </thead>
        <tbody>
          {(lots.data ?? []).map((l) => (
            <tr key={l.id} style={{ borderBottom: '1px solid #333' }}>
              <td>{l.seller_name || l.seller_id.slice(0, 8)}</td>
              <td>{l.sell_amount} {l.sell_resource}</td>
              <td>{l.buy_amount} {l.buy_resource}</td>
              <td>
                {l.seller_id === userId ? (
                  <button onClick={() => cancel.mutate(l.id)} disabled={cancel.isPending}>Отмена</button>
                ) : (
                  <button onClick={() => accept.mutate(l.id)} disabled={accept.isPending}>Купить</button>
                )}
              </td>
            </tr>
          ))}
          {(lots.data ?? []).length === 0 && (
            <tr><td colSpan={4} style={{ textAlign: 'center', color: '#888' }}>Нет открытых лотов</td></tr>
          )}
        </tbody>
      </table>
    </section>
  );
}

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
          ({tf('Main', 'RATE', 'курс')} {last.rate.toFixed(4)})
        </div>
      )}
      <LotsPanel planet={planet} />
    </section>
  );
}
