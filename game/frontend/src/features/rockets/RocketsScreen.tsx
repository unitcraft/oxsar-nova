import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';
import type { Planet } from '@/api/types';
import { useToast } from '@/ui/Toast';
import { useTranslation } from '@/i18n/i18n';

const MISSILE_SILO_ID = 13;
const SILO_CAP_PER_LEVEL = 10;

interface LaunchResult {
  impact_id: string;
  count: number;
  launch_at: string;
  impact_at: string;
}

export function RocketsScreen({ planet }: { planet: Planet }) {
  const { t } = useTranslation('rockets');
  const qc = useQueryClient();
  const toast = useToast();

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
  const have = stock.data?.count ?? 0;
  const maxLaunch = siloMax > 0 ? Math.min(have, siloMax) : have;

  const [g, setG] = useState(planet.galaxy);
  const [s, setS] = useState(planet.system);
  const [pos, setPos] = useState(planet.position);
  const [isMoon, setIsMoon] = useState(false);
  const [count, setCount] = useState(1);
  const [last, setLast] = useState<LaunchResult | null>(null);

  const rocketWord = count === 1 ? t('rocketOne') : count >= 2 && count <= 4 ? t('rocketFew') : t('rocketMany');

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
      toast.show('info', t('toastTitle'), t('toastBody', { count: String(res.count), g: String(g), s: String(s), pos: String(pos) }));
    },
    onError: (err) => {
      toast.show('danger', t('toastError'), err instanceof Error ? err.message : t('toastErrBody'));
    },
  });

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <h2 style={{ margin: 0, fontSize: 18, fontFamily: 'var(--ox-font)', fontWeight: 700 }}>
        {t('title', { planetName: planet.name })}
      </h2>

      <div className="ox-panel" style={{ padding: 16, display: 'flex', alignItems: 'center', gap: 16 }}>
        <img src="/images/units/interplanetary_rocket.gif" alt="" width={48} height={48} style={{ imageRendering: 'pixelated', flexShrink: 0 }} />
        <div>
          <div style={{ fontSize: 20, fontWeight: 700, fontFamily: 'var(--ox-mono)' }}>{have}</div>
          <div style={{ fontSize: 14, color: 'var(--ox-fg-dim)' }}>{t('stockLabel')}</div>
          {siloLevel > 0 ? (
            <div style={{ fontSize: 14, color: 'var(--ox-fg-muted)', marginTop: 2 }}>
              {t('siloInfo', { level: String(siloLevel), max: String(siloMax) })}
            </div>
          ) : (
            <div style={{ fontSize: 14, color: 'var(--ox-warning)', marginTop: 2 }}>
              {t('siloMissing')}
            </div>
          )}
        </div>
      </div>

      <div className="ox-panel" style={{ padding: 20 }}>
        <div style={{ fontSize: 15, fontWeight: 700, letterSpacing: '0.06em', textTransform: 'uppercase', color: 'var(--ox-fg-muted)', marginBottom: 16 }}>
          {t('launchTitle')}
        </div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
          <div>
            <label style={{ fontSize: 14, color: 'var(--ox-fg-dim)', display: 'block', marginBottom: 4 }}>{t('coordsLabel')}</label>
            <div style={{ display: 'flex', gap: 8, alignItems: 'center', flexWrap: 'wrap' }}>
              <span style={{ fontSize: 14, color: 'var(--ox-fg-muted)' }}>G</span>
              <input type="number" min={1} max={16} value={g} onChange={(e) => setG(Number(e.target.value))} style={{ width: 56 }} />
              <span style={{ fontSize: 14, color: 'var(--ox-fg-muted)' }}>S</span>
              <input type="number" min={1} max={999} value={s} onChange={(e) => setS(Number(e.target.value))} style={{ width: 70 }} />
              <span style={{ fontSize: 14, color: 'var(--ox-fg-muted)' }}>P</span>
              <input type="number" min={1} max={15} value={pos} onChange={(e) => setPos(Number(e.target.value))} style={{ width: 56 }} />
              <label style={{ display: 'flex', alignItems: 'center', gap: 4, fontSize: 14 }}>
                <input type="checkbox" checked={isMoon} onChange={(e) => setIsMoon(e.target.checked)} />
                {t('moonLabel')}
              </label>
            </div>
          </div>

          <div>
            <label style={{ fontSize: 14, color: 'var(--ox-fg-dim)', display: 'block', marginBottom: 4 }}>{t('countLabel', { max: String(maxLaunch) })}</label>
            <input
              type="number" min={1} max={maxLaunch} value={count}
              onChange={(e) => setCount(Math.max(1, Math.min(maxLaunch, Number(e.target.value))))}
              style={{ width: 100 }}
            />
          </div>

          <div style={{ fontSize: 14, color: 'var(--ox-fg-muted)' }}>
            {t('hintText')}
          </div>

          <div>
            <button
              type="button"
              className="btn btn-danger"
              disabled={launch.isPending || have < 1 || count < 1 || count > maxLaunch}
              onClick={() => launch.mutate()}
            >
              {launch.isPending ? '…' : t('launchBtn', { count: String(count), rocketWord })}
            </button>
          </div>

          {last && (
            <div className="ox-alert" style={{ marginTop: 4 }}>
              {t('lastLaunch', { count: String(last.count), at: new Date(last.impact_at).toLocaleString('ru-RU') })}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
