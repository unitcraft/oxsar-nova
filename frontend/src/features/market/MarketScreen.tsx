import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';
import { SHIPS, nameOf } from '@/api/catalog';
import type { Planet } from '@/api/types';
import { useToast } from '@/ui/Toast';
import { ScreenSkeleton } from '@/ui/Skeleton';

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
  const [tab, setTab] = useState<'exchange' | 'lots' | 'credit'>('exchange');

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
    onMutate: async () => {
      await qc.cancelQueries({ queryKey: ['planets'] });
      const previous = qc.getQueryData(['planets']);
      if (previous) {
        qc.setQueryData(['planets'], (old: any) => ({
          ...old,
          planets: old.planets.map((p: any) =>
            p.id === planet.id ? {
              ...p,
              [from]: (p[from] ?? 0) - amount,
              [to]: (p[to] ?? 0) + (preview ?? 0),
            } : p
          ),
        }));
      }
      return { previous };
    },
    onSuccess: (res) => {
      setLast(res);
      void qc.invalidateQueries({ queryKey: ['planets'] });
      toast.show('success', 'Обмен', `${res.from_amount} ${res.from} → ${res.to_amount} ${res.to}`);
    },
    onError: (_err, _variables, context) => {
      if (context?.previous) {
        qc.setQueryData(['planets'], context.previous);
      }
      toast.show('danger', 'Ошибка обмена', _err instanceof Error ? _err.message : '');
    },
  });

  const preview = rates.data
    ? Math.floor((amount * rates.data[from]) / rates.data[to] / rates.data.user_rate)
    : null;

  if (rates.isLoading) {
    return <ScreenSkeleton />;
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: 8 }}>
        <h2 style={{ margin: 0, fontSize: 18, fontFamily: 'var(--ox-font)', fontWeight: 700 }}>
          Рынок — {planet.name}
        </h2>
        {rates.data && (
          <span style={{ fontSize: 15, color: 'var(--ox-fg-dim)' }}>
            Ваш курс: <span style={{ fontFamily: 'var(--ox-mono)', color: 'var(--ox-accent)' }}>{rates.data.user_rate.toFixed(2)}</span>
          </span>
        )}
      </div>

      <div className="ox-tabs">
        <button type="button" aria-pressed={tab === 'exchange'} onClick={() => setTab('exchange')}>
          ⇄ Обмен
        </button>
        <button type="button" aria-pressed={tab === 'credit'} onClick={() => setTab('credit')}>
          💳 Кредиты
        </button>
        <button type="button" aria-pressed={tab === 'lots'} onClick={() => setTab('lots')}>
          📋 Ордерная книга
        </button>
      </div>

      {rates.data && tab === 'exchange' && (
        <div style={{ fontSize: 14, color: 'var(--ox-fg-muted)', fontFamily: 'var(--ox-mono)' }}>
          🌐 Глобальный курс: 1 металл = {(rates.data.metal / rates.data.silicon).toFixed(2)} кремния = {(rates.data.metal / rates.data.hydrogen).toFixed(2)} водорода
        </div>
      )}

      {tab === 'exchange' && (
        <div className="ox-panel" style={{ padding: 20 }}>
          <div style={{ fontSize: 14, color: 'var(--ox-fg-dim)', marginBottom: 16 }}>
            Курс M:Si:H = 1:2:4. Чем выше ваш personal rate, тем меньше получаете при обмене.
          </div>

          <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
            <div style={{ display: 'flex', gap: 12, flexWrap: 'wrap', alignItems: 'flex-end' }}>
              <div>
                <label style={{ fontSize: 14, color: 'var(--ox-fg-dim)', display: 'block', marginBottom: 4 }}>Отдать</label>
                <select value={from} onChange={(e) => setFrom(e.target.value as Res)}>
                  {(Object.entries(RES_LABEL) as [Res, string][]).map(([k, v]) => (
                    <option key={k} value={k}>{v}</option>
                  ))}
                </select>
              </div>
              <div>
                <label style={{ fontSize: 14, color: 'var(--ox-fg-dim)', display: 'block', marginBottom: 4 }}>Получить</label>
                <select value={to} onChange={(e) => setTo(e.target.value as Res)}>
                  {(Object.entries(RES_LABEL) as [Res, string][]).map(([k, v]) => (
                    <option key={k} value={k}>{v}</option>
                  ))}
                </select>
              </div>
              <div>
                <label style={{ fontSize: 14, color: 'var(--ox-fg-dim)', display: 'block', marginBottom: 4 }}>Количество</label>
                <input
                  type="number" min={1} value={amount}
                  onChange={(e) => setAmount(Math.max(1, Number(e.target.value)))}
                  style={{ width: 120 }}
                />
              </div>
            </div>

            {preview !== null && from !== to && (
              <div style={{ fontSize: 16 }}>
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
              <div className="ox-alert" style={{ fontSize: 15 }}>
                Последний обмен: {last.from_amount} {last.from} → {last.to_amount} {last.to} (курс {last.rate.toFixed(4)})
              </div>
            )}
          </div>
        </div>
      )}

      {tab === 'credit' && <CreditPanel planet={planet} userRate={rates.data?.user_rate ?? 1.2} />}

      {tab === 'lots' && <LotsPanel planet={planet} userId={planet.user_id} />}
    </div>
  );
}

function LotsPanel({ planet, userId }: { planet: Planet; userId: string }) {
  const qc = useQueryClient();
  const toast = useToast();
  const [subTab, setSubTab] = useState<'resource' | 'fleet'>('resource');
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

  if (subTab === 'fleet') {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        <div className="ox-tabs">
          <button type="button" aria-pressed={subTab === 'resource'} onClick={() => setSubTab('resource')}>💱 Ресурсы</button>
          <button type="button" aria-pressed={subTab === 'fleet'} onClick={() => setSubTab('fleet')}>🛸 Флот</button>
        </div>
        <FleetLotsPanel planet={planet} userId={userId} />
      </div>
    );
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      <div className="ox-tabs">
        <button type="button" aria-pressed={subTab === 'resource'} onClick={() => setSubTab('resource')}>💱 Ресурсы</button>
        <button type="button" aria-pressed={subTab === 'fleet'} onClick={() => setSubTab('fleet')}>🛸 Флот</button>
      </div>
      {/* Create lot */}
      <div className="ox-panel" style={{ padding: 16 }}>
        <div style={{ fontSize: 14, fontWeight: 700, letterSpacing: '0.06em', textTransform: 'uppercase', color: 'var(--ox-fg-muted)', marginBottom: 12 }}>
          Выставить лот
        </div>
        <div style={{ display: 'flex', gap: 10, alignItems: 'flex-end', flexWrap: 'wrap' }}>
          <div>
            <label style={{ fontSize: 14, color: 'var(--ox-fg-dim)', display: 'block', marginBottom: 4 }}>Продать</label>
            <div style={{ display: 'flex', gap: 6 }}>
              <input type="number" min={1} value={sellAmt} onChange={(e) => setSellAmt(Math.max(1, Number(e.target.value)))} style={{ width: 100 }} />
              <select value={sellRes} onChange={(e) => setSellRes(e.target.value as Res)}>
                {(Object.entries(RES_LABEL) as [Res, string][]).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
              </select>
            </div>
          </div>
          <div style={{ fontSize: 16, color: 'var(--ox-fg-muted)', paddingBottom: 6 }}>→</div>
          <div>
            <label style={{ fontSize: 14, color: 'var(--ox-fg-dim)', display: 'block', marginBottom: 4 }}>Получить</label>
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

interface CreditExchangeResult {
  direction: 'to_credit' | 'from_credit';
  resource: string;
  resource_delta: number;
  credit_delta: number;
}

function CreditPanel({ planet, userRate }: { planet: Planet; userRate: number }) {
  const qc = useQueryClient();
  const toast = useToast();
  const [direction, setDirection] = useState<'to_credit' | 'from_credit'>('to_credit');
  const [resource, setResource] = useState<Res>('metal');
  const [amount, setAmount] = useState(1000);

  const RES_COST: Record<Res, number> = { metal: 1, silicon: 2, hydrogen: 4 };
  const CREDIT_RATE_PER_UNIT = 100;

  const preview = direction === 'to_credit'
    ? amount * RES_COST[resource] / CREDIT_RATE_PER_UNIT / userRate
    : Math.floor(amount * CREDIT_RATE_PER_UNIT / RES_COST[resource] / userRate);

  const exchange = useMutation({
    mutationFn: () => api.post<CreditExchangeResult>(`/api/planets/${planet.id}/market/credit`, {
      direction, resource, amount,
    }),
    onSuccess: (res) => {
      void qc.invalidateQueries({ queryKey: ['planets'] });
      void qc.invalidateQueries({ queryKey: ['me'] });
      const msg = res.direction === 'to_credit'
        ? `${Math.abs(res.resource_delta)} ${res.resource} → ${res.credit_delta.toFixed(2)} кред.`
        : `${amount} кред. → ${res.resource_delta} ${res.resource}`;
      toast.show('success', 'Обмен', msg);
    },
    onError: (e) => toast.show('danger', 'Ошибка', e instanceof Error ? e.message : ''),
  });

  return (
    <div className="ox-panel" style={{ padding: 20, display: 'flex', flexDirection: 'column', gap: 14 }}>
      <div style={{ fontSize: 14, color: 'var(--ox-fg-dim)' }}>
        Обмен ресурсов на кредиты и обратно. 1 кредит = {CREDIT_RATE_PER_UNIT} единиц металла.
        Курс учитывает ваш personal rate ({userRate.toFixed(2)}).
      </div>

      <div style={{ display: 'flex', gap: 12, flexWrap: 'wrap', alignItems: 'flex-end' }}>
        <div>
          <label style={{ fontSize: 14, color: 'var(--ox-fg-dim)', display: 'block', marginBottom: 4 }}>Направление</label>
          <select value={direction} onChange={(e) => setDirection(e.target.value as 'to_credit' | 'from_credit')}>
            <option value="to_credit">Ресурс → Кредиты</option>
            <option value="from_credit">Кредиты → Ресурс</option>
          </select>
        </div>
        <div>
          <label style={{ fontSize: 14, color: 'var(--ox-fg-dim)', display: 'block', marginBottom: 4 }}>Ресурс</label>
          <select value={resource} onChange={(e) => setResource(e.target.value as Res)}>
            {(Object.entries(RES_LABEL) as [Res, string][]).map(([k, v]) => (
              <option key={k} value={k}>{v}</option>
            ))}
          </select>
        </div>
        <div>
          <label style={{ fontSize: 14, color: 'var(--ox-fg-dim)', display: 'block', marginBottom: 4 }}>
            {direction === 'to_credit' ? 'Количество ресурса' : 'Количество кредитов'}
          </label>
          <input
            type="number" min={1} value={amount}
            onChange={(e) => setAmount(Math.max(1, Number(e.target.value)))}
            style={{ width: 140 }}
          />
        </div>
      </div>

      <div style={{ fontSize: 16 }}>
        {direction === 'to_credit' ? (
          <>Вы получите: <span style={{ fontFamily: 'var(--ox-mono)', color: 'var(--ox-accent)', fontWeight: 700 }}>{preview.toFixed(2)}</span> кредитов</>
        ) : (
          <>Вы получите: <span style={{ fontFamily: 'var(--ox-mono)', color: 'var(--ox-accent)', fontWeight: 700 }}>{(preview as number).toLocaleString('ru-RU')}</span> {RES_LABEL[resource]}</>
        )}
      </div>

      <div>
        <button
          type="button"
          className="btn"
          disabled={exchange.isPending || amount <= 0}
          onClick={() => exchange.mutate()}
        >
          {exchange.isPending ? '…' : 'Обменять'}
        </button>
      </div>
    </div>
  );
}

interface FleetLot {
  id: string;
  seller_id: string;
  seller_name: string;
  planet_id: string;
  sell_fleet: Record<string, number>;
  buy_resource: string;
  buy_amount: number;
  state: string;
  created_at: string;
}

function FleetLotsPanel({ planet, userId }: { planet: Planet; userId: string }) {
  const qc = useQueryClient();
  const toast = useToast();

  const lots = useQuery({
    queryKey: ['market-fleet-lots'],
    queryFn: () => api.get<{ lots: FleetLot[] }>('/api/market/fleet-lots'),
    refetchInterval: 15000,
  });

  const [buildFleet, setBuildFleet] = useState<Record<number, number>>({});
  const [buyRes, setBuyRes] = useState<Res>('metal');
  const [buyAmt, setBuyAmt] = useState(10000);

  const createLot = useMutation({
    mutationFn: () => api.post<FleetLot>(`/api/planets/${planet.id}/market/fleet-lots`, {
      fleet: buildFleet, buy_resource: buyRes, buy_amount: buyAmt,
    }),
    onSuccess: () => {
      setBuildFleet({});
      void qc.invalidateQueries({ queryKey: ['market-fleet-lots'] });
      void qc.invalidateQueries({ queryKey: ['planets'] });
      toast.show('success', 'Лот флота выставлен');
    },
    onError: (e) => toast.show('danger', 'Ошибка', e instanceof Error ? e.message : ''),
  });

  const cancelLot = useMutation({
    mutationFn: (id: string) => api.delete(`/api/market/fleet-lots/${id}`),
    onSuccess: () => { void qc.invalidateQueries({ queryKey: ['market-fleet-lots'] }); toast.show('info', 'Лот отменён'); },
  });

  const acceptLot = useMutation({
    mutationFn: (id: string) => api.post(`/api/market/fleet-lots/${id}/accept`, { planet_id: planet.id }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['market-fleet-lots'] });
      void qc.invalidateQueries({ queryKey: ['planets'] });
      toast.show('success', 'Сделка выполнена');
    },
    onError: (e) => toast.show('danger', 'Ошибка', e instanceof Error ? e.message : ''),
  });

  const totalShips = Object.values(buildFleet).reduce((s, v) => s + (v || 0), 0);

  return (
    <>
      {/* Создание лота */}
      <div className="ox-panel" style={{ padding: 16, display: 'flex', flexDirection: 'column', gap: 12 }}>
        <div style={{ fontSize: 14, fontWeight: 700, letterSpacing: '0.06em', textTransform: 'uppercase', color: 'var(--ox-fg-muted)' }}>
          Выставить флот на продажу
        </div>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(160px, 1fr))', gap: 8 }}>
          {SHIPS.map((s) => (
            <div key={s.id} style={{ display: 'flex', gap: 6, alignItems: 'center' }}>
              <span style={{ flex: 1, fontSize: 14 }}>🛸 {s.name}</span>
              <input
                type="number"
                min={0}
                value={buildFleet[s.id] ?? 0}
                onChange={(e) => {
                  const v = Math.max(0, Number(e.target.value));
                  setBuildFleet((m) => ({ ...m, [s.id]: v }));
                }}
                style={{ width: 70 }}
              />
            </div>
          ))}
        </div>
        <div style={{ display: 'flex', gap: 10, alignItems: 'flex-end', flexWrap: 'wrap' }}>
          <div>
            <label style={{ fontSize: 14, color: 'var(--ox-fg-dim)', display: 'block', marginBottom: 4 }}>Хочу получить</label>
            <div style={{ display: 'flex', gap: 6 }}>
              <input type="number" min={1} value={buyAmt} onChange={(e) => setBuyAmt(Math.max(1, Number(e.target.value)))} style={{ width: 120 }} />
              <select value={buyRes} onChange={(e) => setBuyRes(e.target.value as Res)}>
                {(Object.entries(RES_LABEL) as [Res, string][]).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
              </select>
            </div>
          </div>
          <button
            type="button"
            className="btn btn-sm"
            disabled={createLot.isPending || totalShips === 0 || buyAmt <= 0}
            onClick={() => {
              // Отфильтровать 0-значения.
              const nz: Record<number, number> = {};
              Object.entries(buildFleet).forEach(([k, v]) => {
                if (v > 0) nz[Number(k)] = v;
              });
              setBuildFleet(nz);
              createLot.mutate();
            }}
          >
            {createLot.isPending ? '…' : `Выставить (${totalShips} кораблей)`}
          </button>
        </div>
      </div>

      {/* Лоты */}
      <div className="ox-panel" style={{ overflow: 'hidden' }}>
        <div className="ox-table-responsive">
          <table className="ox-table" style={{ margin: 0 }}>
            <thead>
              <tr>
                <th>Продавец</th>
                <th>Состав</th>
                <th>Цена</th>
                <th />
              </tr>
            </thead>
            <tbody>
              {(lots.data?.lots ?? []).map((l) => {
                const isOwn = l.seller_id === userId;
                return (
                  <tr key={l.id}>
                    <td style={{ fontWeight: 600 }}>{isOwn ? '(вы)' : l.seller_name}</td>
                    <td style={{ fontSize: 14 }}>
                      {Object.entries(l.sell_fleet).map(([idStr, cnt]) => (
                        <div key={idStr}>🛸 {nameOf(Number(idStr))} × {cnt}</div>
                      ))}
                    </td>
                    <td className="num" style={{ fontFamily: 'var(--ox-mono)' }}>
                      {Math.round(l.buy_amount).toLocaleString('ru-RU')} {RES_LABEL[l.buy_resource as Res] ?? l.buy_resource}
                    </td>
                    <td>
                      {isOwn ? (
                        <button type="button" className="btn-ghost btn-sm" onClick={() => cancelLot.mutate(l.id)}>Отменить</button>
                      ) : (
                        <button type="button" className="btn btn-sm" disabled={acceptLot.isPending} onClick={() => acceptLot.mutate(l.id)}>Купить</button>
                      )}
                    </td>
                  </tr>
                );
              })}
              {(lots.data?.lots ?? []).length === 0 && (
                <tr>
                  <td colSpan={4} style={{ textAlign: 'center', padding: 20, color: 'var(--ox-fg-muted)' }}>
                    Нет открытых лотов флота
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </>
  );
}
