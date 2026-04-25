import { useQuery } from '@tanstack/react-query';
import { api } from '../../api/client';

type GalaxyEvent = {
  id: number;
  kind: string;
  started_at: string;
  ends_at: string;
  params: Record<string, unknown>;
};

const KIND_META: Record<string, { icon: string; title: string; descr: string }> = {
  meteor_storm: { icon: '☄️', title: 'Метеоритный шторм',  descr: '+30% к добыче металла' },
  solar_flare:  { icon: '🌞', title: 'Солнечная вспышка',  descr: '−20% энергии' },
  trade_forum:  { icon: '📈', title: 'Торговый форум',     descr: 'Льготные курсы рынка' },
  star_nebula:  { icon: '🌌', title: 'Звёздная туманность', descr: '+15% к мощи экспедиций' },
};

function formatRemaining(endsAt: string): string {
  const ms = new Date(endsAt).getTime() - Date.now();
  if (ms <= 0) return '—';
  const h = Math.floor(ms / 3_600_000);
  const m = Math.floor((ms % 3_600_000) / 60_000);
  if (h > 0) return `${h}ч ${m}м`;
  return `${m}м`;
}

/**
 * Глобальный баннер активного галактического события (план 17 F).
 * Возвращает null если событий нет (204 No Content от backend).
 */
export function GalaxyEventBanner() {
  const q = useQuery({
    queryKey: ['galaxy-event'],
    queryFn: () => api.get<GalaxyEvent | undefined>('/api/galaxy-event'),
    refetchInterval: 60_000,
  });

  const e = q.data;
  if (!e) return null;
  const meta = KIND_META[e.kind] ?? { icon: '✨', title: e.kind, descr: 'Действует особый эффект' };
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
      <span style={{ fontSize: 24 }}>{meta.icon}</span>
      <div style={{ display: 'flex', flexDirection: 'column', flex: 1 }}>
        <strong>{meta.title}</strong>
        <span style={{ fontSize: 13, color: 'var(--ox-fg-muted)' }}>{meta.descr}</span>
      </div>
      <span style={{ fontSize: 13, color: 'var(--ox-fg-muted)' }}>
        До конца: <strong>{formatRemaining(e.ends_at)}</strong>
      </span>
    </div>
  );
}
