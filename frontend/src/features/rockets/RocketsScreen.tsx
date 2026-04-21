import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';
import { useTranslation } from '@/i18n/i18n';
import type { Planet } from '@/api/types';

const MISSILE_SILO_ID = 13;
const SILO_CAP_PER_LEVEL = 10;

// RocketsScreen — запуск межпланетарных ракет (kind=16). Ракеты
// летят без возврата, бьют только по defense цели.

interface LaunchResult {
  impact_id: string;
  count: number;
  launch_at: string;
  impact_at: string;
}

export function RocketsScreen({ planet }: { planet: Planet }) {
  const { t, tf } = useTranslation();
  const qc = useQueryClient();

  const stock = useQuery({
    queryKey: ['rockets-stock', planet.id],
    queryFn: () => api.get<{ count: number }>(`/api/planets/${planet.id}/rockets`),
    refetchInterval: 3000,
  });

  const buildingsQ = useQuery({
    queryKey: ['buildings-levels', planet.id],
    queryFn: () => api.get<{ levels: Record<string, number> }>(`/api/planets/${planet.id}/buildings`),
    staleTime: 30000,
  });

  const siloLevel = buildingsQ.data?.levels[MISSILE_SILO_ID] ?? 0;
  const siloMax = siloLevel * SILO_CAP_PER_LEVEL;

  const [g, setG] = useState(planet.galaxy);
  const [s, setS] = useState(planet.system);
  const [pos, setPos] = useState(planet.position);
  const [isMoon, setIsMoon] = useState(false);
  const [count, setCount] = useState(1);
  const [last, setLast] = useState<LaunchResult | null>(null);

  const launch = useMutation({
    mutationFn: () =>
      api.post<LaunchResult>(`/api/planets/${planet.id}/rockets/launch`, {
        dst: { galaxy: g, system: s, position: pos, is_moon: isMoon },
        count,
      }),
    onSuccess: (res) => {
      setLast(res);
      void qc.invalidateQueries({ queryKey: ['rockets-stock', planet.id] });
      void qc.invalidateQueries({ queryKey: ['shipyard-inventory', planet.id] });
    },
  });

  const have = stock.data?.count ?? 0;
  const maxLaunch = siloMax > 0 ? Math.min(have, siloMax) : have;

  return (
    <section>
      <h2>{tf('global', 'MENU_ROCKETS', 'Ракеты')} — {planet.name}</h2>
      <p>
        {tf(
          'Main',
          'ROCKETS_HINT',
          'Межпланетарные ракеты летят без возврата, уничтожают оборону цели. 1 ракета = 12000 урона по пулу defense × shell.',
        )}
      </p>

      <p>
        <b>{tf('Main', 'ROCKETS_STOCK', 'В наличии')}:</b> {have}
        {siloLevel > 0 && (
          <span style={{ color: 'var(--ox-muted, #888)', marginLeft: 8 }}>
            ({tf('Main', 'ROCKETS_SILO_LVL', 'шахта ур.')} {siloLevel},{' '}
            {tf('Main', 'ROCKETS_MAX', 'макс.')} {siloMax})
          </span>
        )}
        {siloLevel === 0 && (
          <span style={{ color: 'orange', marginLeft: 8 }}>
            {tf('Main', 'ROCKETS_NO_SILO', '— постройте Ракетную шахту для лимита')}
          </span>
        )}
      </p>

      <h3>{t('Main', 'POSITION')}</h3>
      <div style={{ display: 'flex', gap: 8, marginBottom: 12, alignItems: 'center' }}>
        G&nbsp;<input type="number" min={1} max={16} value={g} onChange={(e) => setG(Number(e.target.value))} style={{ width: 60 }} />
        S&nbsp;<input type="number" min={1} max={999} value={s} onChange={(e) => setS(Number(e.target.value))} style={{ width: 80 }} />
        P&nbsp;<input type="number" min={1} max={15} value={pos} onChange={(e) => setPos(Number(e.target.value))} style={{ width: 60 }} />
        <label>
          <input type="checkbox" checked={isMoon} onChange={(e) => setIsMoon(e.target.checked)} />
          &nbsp;moon
        </label>
      </div>

      <label>
        {tf('Main', 'MARKET_AMOUNT', 'Количество')}:
        <input
          type="number"
          min={1}
          max={have}
          value={count}
          onChange={(e) => setCount(Math.max(1, Math.min(maxLaunch, Number(e.target.value))))}
          style={{ width: 100, marginLeft: 8 }}
        />
      </label>

      <div style={{ marginTop: 12 }}>
        <button
          type="button"
          disabled={launch.isPending || have < 1 || count < 1 || count > maxLaunch}
          onClick={() => launch.mutate()}
        >
          {launch.isPending ? '…' : tf('Main', 'ROCKETS_LAUNCH', 'Запустить')}
        </button>
      </div>

      {launch.isError && (
        <div className="ox-error">
          {launch.error instanceof Error ? launch.error.message : t('global', 'ERROR')}
        </div>
      )}

      {last && (
        <div style={{ marginTop: 12 }}>
          <b>{tf('Main', 'ROCKETS_LAST', 'Последний пуск')}:</b>{' '}
          {last.count} {tf('Main', 'ROCKETS_WORD', 'ракет')}, {tf('Main', 'UNTIL', 'до')}{' '}
          {new Date(last.impact_at).toLocaleString('ru-RU')}
        </div>
      )}
    </section>
  );
}
