import { useQuery } from '@tanstack/react-query';
import { api } from '../../api/client';
import { useTranslation } from '@/i18n/i18n';

type GalaxyEvent = {
  id: number;
  kind: string;
  started_at: string;
  ends_at: string;
  params: Record<string, unknown>;
};

const KIND_ICON: Record<string, string> = {
  meteor_storm: '☄️',
  solar_flare:  '🌞',
  trade_forum:  '📈',
  star_nebula:  '🌌',
};

function formatRemaining(endsAt: string, unitHour: string, unitMin: string): string {
  const ms = new Date(endsAt).getTime() - Date.now();
  if (ms <= 0) return '—';
  const h = Math.floor(ms / 3_600_000);
  const m = Math.floor((ms % 3_600_000) / 60_000);
  if (h > 0) return `${h}${unitHour} ${m}${unitMin}`;
  return `${m}${unitMin}`;
}

/**
 * Глобальный баннер активного галактического события (план 17 F).
 * Возвращает null если событий нет (204 No Content от backend).
 */
export function GalaxyEventBanner() {
  const { t, tf } = useTranslation('galaxyEvent');
  const unitHour = t('global', 'timeUnitHour');
  const unitMin  = t('global', 'timeUnitMin');
  const q = useQuery({
    queryKey: ['galaxy-event'],
    queryFn: () => api.get<GalaxyEvent | undefined>('/api/galaxy-event'),
    refetchInterval: 60_000,
  });

  const e = q.data;
  if (!e) return null;
  const icon = KIND_ICON[e.kind] ?? '✨';
  const title = tf('galaxyEvent', `kind.${e.kind}.title`, t('defaultTitle'));
  const descr = tf('galaxyEvent', `kind.${e.kind}.descr`, t('defaultDescr'));
  return (
    <div
      className="ox-panel"
      style={{
        padding: '8px 12px',
        margin: '8px 16px',
        display: 'flex',
        alignItems: 'center',
        gap: 10,
        background: 'linear-gradient(90deg, var(--ox-bg-deep) 0%, var(--ox-bg) 100%)',
        borderLeft: '3px solid var(--ox-accent, #4a90e2)',
      }}
    >
      <span style={{ fontSize: 24 }}>{icon}</span>
      <div style={{ display: 'flex', flexDirection: 'column', flex: 1 }}>
        <strong>{title}</strong>
        <span style={{ fontSize: 13, color: 'var(--ox-fg-muted)' }}>{descr}</span>
      </div>
      <span style={{ fontSize: 13, color: 'var(--ox-fg-muted)' }}>
        {t('untilEnd')} <strong>{formatRemaining(e.ends_at, unitHour, unitMin)}</strong>
      </span>
    </div>
  );
}
