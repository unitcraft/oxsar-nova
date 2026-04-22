import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';
import type { Planet } from '@/api/types';
import { useToast } from '@/ui/Toast';

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

const RES_LABEL: Record<Res, string> = { metal: '🟠 Металл', silicon: '💎 Кремний', hydrogen: '💧 Водород' };

export function MarketScreen({ planet }: { planet: Planet }) {
  const qc = useQueryClient();
  const toast = useToast();
  const [tab, setTab] = useState<'exchange' | 'lots'>('exchange');

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
    mutationFn: () => api.post<ExchangeResult>(`/api/planets/${planet.id}/market/exchange`, { from, to, amount }),
    onSuccess: (res) => {
      setLast(res);
      void qc.invalidateQueries({ queryKey: ['planets'] });
      toast.show('success', 'Обмен', `${res.from_amount} ${res.from} → ${res.to_amount} ${res.to}`);
    },
    onError: (err) => { toast.show('danger', 'Ошибка обмена', err instanceof Error ? err.message : ''); },
  });

  const preview = rates.data
    ? Math.floor((amount * rates.data[from]) / rates.data[to] / rates.data.user_rate)
    : null;

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: 8 }}>
        <h2 style={{ margin: 0, fontSize: 18, fontFamily: 'var(--ox-font)', fontWeight: 700 }}>
          Рынок — {planet.name}
        </h2>
        {rates.data && (
          <span style={{ fontSize: 13, color: 'var(--ox-fg-dim)' }}>
            Ваш курс: <span style={{ fontFamily: 'var(--ox-mono)', color: 'var(--ox-accent)' }}>{rates.data.user_rate.toFixed(2)}</span>
          </span>
        )}
      </div>

      <div className="ox-tabs">
        <button type="button" aria-pressed={tab === 'exchange'} onClick={() => setTab('exchange')}>
          ⇄ Обмен
        </button>
        <button type="button" aria-pressed={tab === 'lots'} onClick={() => setTab('lots')}>
          📋 Ордерная книга
        </button>
      </div>

      {tab === 'exchange' && (
        <div className="ox-panel" style={{ padding: 20 }}>
          <div style={{ fontSize: 12, color: 'var(--ox-fg-dim)', marginBottom: 16 }}>
            Курс M:Si:H = 1:2:4. Чем выше ваш personal rate, тем меньше получаете при обмене.
          </div>

          <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
            <div style={{ display: 'flex', gap: 12, flexWrap: 'wrap', alignItems: 'flex-end' }}>
              <div>
                <label style={{ fontSize: 12, color: 'var(--ox-fg-dim)', display: 'block', marginBottom: 4 }}>Отдать</label>
                <select value={from} onChange={(e) => setFrom(e.target.value as Res)}>
                  {(Object.entries(RES_LABEL) as [Res, string][]).map(([k, v]) => (
                    <option key={k} value={k}>{v}</option>
                  ))}
                </select>
              </div>
              <div>
                <label style={{ fontSize: 12, color: 'var(--ox-fg-dim)', display: 'block', marginBottom: 4 }}>Получить</label>
                <select value={to} onChange={(e) => setTo(e.target.value as Res)}>
                  {(Object.entries(RES_LABEL) as [Res, string][]).map(([k, v]) => (
                    <option key={k} value={k}>{v}</option>
                  ))}
                </select>
              </div>
              <div>
                <label style={{ fontSize: 12, color: 'var(--ox-fg-dim)', display: 'block', marginBottom: 4 }}>Количество</label>
                <input
                  type="number" min={1} value={amount}
                  onChange={(e) => setAmount(Math.max(1, Number(e.target.value)))}
                  style={{ width: 120 }}
                />
              </div>
            </div>

            {preview !== null && from !== to && (
              <div style={{ fontSize: 14 }}>
                Вы получите:{' '}
                <span style={{ fontWeight: 700, color: 'var(--ox-accent)', fontFamily: 'var(--ox-mono)' }}>
                  {preview.toLocaleString('ru-RU')}
                </span>{' '}
                {RES_LABEL[to]}
              </div>
            )}

            <div>
              <button
                type="button"
                className="btn"
                disabled={exchange.isPending || from === to || amount <= 0}
                onClick={() => exchange.mutate()}
              >
                {exchange.isPending ? '…' : 'Обменять'}
              </button>
            </div>

            {last && (
              <div className="ox-alert" style={{ fontSize: 13 }}>
                Последний обмен: {last.from_amount} {last.from} → {last.to_amount} {last.to} (курс {last.rate.toFixed(4)})
              </div>
            )}
          </div>
        </div>
      )}

      {tab === 'lots' && <LotsPanel planet={planet} userId={planet.user_id} />}
    </div>
  );
}

function LotsPanel({ planet, userId }: { planet: Planet; userId: string }) {
  const qc = useQueryClient();
  const toast = useToast();
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
    mutationFn: () => api.post('/api/market/lots', { planet_id: planet.id, sell_resource: sellRes, sell_amount: sellAmt, buy_resource: buyRes, buy_amount: buyAmt }),
    onSuccess: () => { void qc.invalidateQueries({ queryKey: ['market-lots'] }); toast.show('success', 'Лот выставлен'); },
    onError: (err) => { toast.show('danger', 'Ошибка', err instanceof Error ? err.message : ''); },
  });

  const cancel = useMutation({
    mutationFn: (id: string) => api.delete(`/api/market/lots/${id}`),
    onSuccess: () => { void qc.invalidateQueries({ queryKey: ['market-lots'] }); toast.show('info', 'Лот отменён'); },
  });

  const accept = useMutation({
    mutationFn: (id: string) => api.post(`/api/market/lots/${id}/accept`, { planet_id: planet.id }),
    onSuccess: () => { void qc.invalidateQueries({ queryKey: ['market-lots'] }); void qc.invalidateQueries({ queryKey: ['planets'] }); toast.show('success', 'Сделка выполнена'); },
    onError: (err) => { toast.show('danger', 'Ошибка', err instanceof Error ? err.message : ''); },
  });

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      {/* Create lot */}
      <div className="ox-panel" style={{ padding: 16 }}>
        <div style={{ fontSize: 12, fontWeight: 700, letterSpacing: '0.06em', textTransform: 'uppercase', color: 'var(--ox-fg-muted)', marginBottom: 12 }}>
          Выставить лот
        </div>
        <div style={{ display: 'flex', gap: 10, alignItems: 'flex-end', flexWrap: 'wrap' }}>
          <div>
            <label style={{ fontSize: 12, color: 'var(--ox-fg-dim)', display: 'block', marginBottom: 4 }}>Продать</label>
            <div style={{ display: 'flex', gap: 6 }}>
              <input type="number" min={1} value={sellAmt} onChange={(e) => setSellAmt(Math.max(1, Number(e.target.value)))} style={{ width: 100 }} />
              <select value={sellRes} onChange={(e) => setSellRes(e.target.value as Res)}>
                {(Object.entries(RES_LABEL) as [Res, string][]).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
              </select>
            </div>
          </div>
          <div style={{ fontSize: 16, color: 'var(--ox-fg-muted)', paddingBottom: 6 }}>→</div>
          <div>
            <label style={{ fontSize: 12, color: 'var(--ox-fg-dim)', display: 'block', marginBottom: 4 }}>Получить</label>
            <div style={{ display: 'flex', gap: 6 }}>
              <input type="number" min={1} value={buyAmt} onChange={(e) => setBuyAmt(Math.max(1, Number(e.target.value)))} style={{ width: 100 }} />
              <select value={buyRes} onChange={(e) => setBuyRes(e.target.value as Res)}>
                {(Object.entries(RES_LABEL) as [Res, string][]).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
              </select>
            </div>
          </div>
          <button type="button" className="btn btn-sm" onClick={() => create.mutate()} disabled={create.isPending || sellRes === buyRes}>
            {create.isPending ? '…' : 'Выставить'}
          </button>
        </div>
      </div>

      {/* Lots table */}
      <div className="ox-panel" style={{ overflow: 'hidden' }}>
        <div className="ox-table-responsive">
          <table className="ox-table" style={{ margin: 0 }}>
            <thead>
              <tr>
                <th>Продавец</th>
                <th>Продаёт</th>
                <th>Хочет</th>
                <th />
              </tr>
            </thead>
            <tbody>
              {(lots.data ?? []).map((l) => (
                <tr key={l.id}>
                  <td data-label="Продавец">{l.seller_name || l.seller_id.slice(0, 8)}</td>
                  <td data-label="Продаёт" style={{ fontFamily: 'var(--ox-mono)' }}>{l.sell_amount.toLocaleString('ru-RU')} {l.sell_resource}</td>
                  <td data-label="Хочет" style={{ fontFamily: 'var(--ox-mono)' }}>{l.buy_amount.toLocaleString('ru-RU')} {l.buy_resource}</td>
                  <td>
                    {l.seller_id === userId ? (
                      <button type="button" className="btn-ghost btn-sm" onClick={() => cancel.mutate(l.id)} disabled={cancel.isPending}>Отмена</button>
                    ) : (
                      <button type="button" className="btn btn-sm btn-success" onClick={() => accept.mutate(l.id)} disabled={accept.isPending}>Купить</button>
                    )}
                  </td>
                </tr>
              ))}
              {(lots.data ?? []).length === 0 && (
                <tr><td colSpan={4} style={{ textAlign: 'center', color: 'var(--ox-fg-muted)', padding: '20px 0' }}>Нет открытых лотов</td></tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
