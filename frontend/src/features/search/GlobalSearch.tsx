import { useEffect, useRef, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '@/api/client';
import { useTranslation } from '@/i18n/i18n';

interface PlayerResult {
  user_id: string;
  username: string;
  alliance_tag?: string | null;
  points: number;
}
interface AllianceResult {
  tag: string;
  name: string;
  members: number;
  points: number;
}
interface PlanetResult {
  planet_id: string;
  name: string;
  galaxy: number;
  system: number;
  position: number;
  owner: string;
}
interface SearchResponse {
  players: PlayerResult[];
  alliances: AllianceResult[];
  planets: PlanetResult[];
}

export function GlobalSearch({ open, onClose, onNavigate }: {
  open: boolean;
  onClose: () => void;
  onNavigate?: (target: { kind: 'player' | 'alliance' | 'planet'; data: PlayerResult | AllianceResult | PlanetResult }) => void;
}) {
  const { t } = useTranslation('search');
  const [q, setQ] = useState('');
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (open && inputRef.current) {
      inputRef.current.focus();
      setQ('');
    }
  }, [open]);

  useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [open, onClose]);

  const results = useQuery({
    queryKey: ['search', q],
    queryFn: () => api.get<SearchResponse>(`/api/search?q=${encodeURIComponent(q)}`),
    enabled: open && q.trim().length >= 2,
    staleTime: 10000,
  });

  if (!open) return null;

  const data = results.data;
  const hasResults = data && (data.players.length || data.alliances.length || data.planets.length);

  return (
    <div
      onClick={onClose}
      style={{
        position: 'fixed', inset: 0, zIndex: 1000,
        background: 'rgba(0,0,0,0.6)',
        display: 'flex', alignItems: 'flex-start', justifyContent: 'center',
        paddingTop: '10vh',
      }}
    >
      <div
        onClick={(e) => e.stopPropagation()}
        style={{
          width: '90%', maxWidth: 640,
          background: 'var(--ox-bg-panel)',
          border: '1px solid var(--ox-border)',
          borderRadius: 8,
          overflow: 'hidden',
          boxShadow: '0 10px 50px rgba(0,0,0,0.5)',
        }}
      >
        <div style={{ padding: 14, borderBottom: '1px solid var(--ox-border)' }}>
          <input
            ref={inputRef}
            type="text"
            placeholder={t('placeholder')}
            value={q}
            onChange={(e) => setQ(e.target.value)}
            style={{
              width: '100%', padding: '10px 12px',
              background: 'transparent',
              border: '1px solid var(--ox-border)', borderRadius: 4,
              color: 'var(--ox-fg)', fontSize: 15,
            }}
          />
          <div style={{ marginTop: 6, fontSize: 10, color: 'var(--ox-fg-muted)', fontFamily: 'var(--ox-mono)' }}>
            {t('hint')}
          </div>
        </div>

        <div style={{ maxHeight: '60vh', overflowY: 'auto' }}>
          {q.trim().length < 2 && (
            <div style={{ padding: 24, textAlign: 'center', color: 'var(--ox-fg-muted)' }}>
              {t('promptEmpty')}
            </div>
          )}

          {q.trim().length >= 2 && results.isLoading && (
            <div style={{ padding: 24, textAlign: 'center', color: 'var(--ox-fg-muted)' }}>{t('searching')}</div>
          )}

          {q.trim().length >= 2 && !results.isLoading && !hasResults && (
            <div style={{ padding: 24, textAlign: 'center', color: 'var(--ox-fg-muted)' }}>
              {t('notFound')}
            </div>
          )}

          {data && data.players.length > 0 && (
            <Section title={t('sectionPlayers')} icon="👤">
              {data.players.map((p) => (
                <button key={p.user_id} type="button" className="ox-search-item"
                  onClick={() => { onNavigate?.({ kind: 'player', data: p }); onClose(); }}
                  style={rowStyle}
                >
                  <span style={{ flex: 1 }}>
                    <span style={{ fontWeight: 600 }}>{p.username}</span>
                    {p.alliance_tag && <span style={{ marginLeft: 8, fontFamily: 'var(--ox-mono)', color: 'var(--ox-accent)', fontSize: 14 }}>[{p.alliance_tag}]</span>}
                  </span>
                  <span style={{ fontFamily: 'var(--ox-mono)', fontSize: 14, color: 'var(--ox-fg-muted)' }}>
                    {Math.round(p.points).toLocaleString('ru-RU')}
                  </span>
                </button>
              ))}
            </Section>
          )}

          {data && data.alliances.length > 0 && (
            <Section title={t('sectionAlliances')} icon="🤝">
              {data.alliances.map((a) => (
                <button key={a.tag} type="button" className="ox-search-item"
                  onClick={() => { onNavigate?.({ kind: 'alliance', data: a }); onClose(); }}
                  style={rowStyle}
                >
                  <span style={{ flex: 1 }}>
                    <span style={{ fontFamily: 'var(--ox-mono)', color: 'var(--ox-accent)' }}>[{a.tag}]</span>
                    <span style={{ marginLeft: 8 }}>{a.name}</span>
                  </span>
                  <span style={{ fontFamily: 'var(--ox-mono)', fontSize: 14, color: 'var(--ox-fg-muted)' }}>
                    {t('membersCount', { n: String(a.members) })} · {Math.round(a.points).toLocaleString('ru-RU')}
                  </span>
                </button>
              ))}
            </Section>
          )}

          {data && data.planets.length > 0 && (
            <Section title={t('sectionPlanets')} icon="🪐">
              {data.planets.map((p) => (
                <button key={p.planet_id} type="button" className="ox-search-item"
                  onClick={() => { onNavigate?.({ kind: 'planet', data: p }); onClose(); }}
                  style={rowStyle}
                >
                  <span style={{ flex: 1 }}>
                    <span style={{ fontWeight: 600 }}>{p.name}</span>
                    <span style={{ marginLeft: 8, fontFamily: 'var(--ox-mono)', color: 'var(--ox-fg-muted)', fontSize: 14 }}>
                      [{p.galaxy}:{p.system}:{p.position}]
                    </span>
                  </span>
                  <span style={{ fontSize: 14, color: 'var(--ox-fg-muted)' }}>{p.owner || '—'}</span>
                </button>
              ))}
            </Section>
          )}
        </div>
      </div>
    </div>
  );
}

const rowStyle: React.CSSProperties = {
  display: 'flex', alignItems: 'center', width: '100%',
  padding: '10px 14px',
  background: 'transparent', border: 'none', borderBottom: '1px solid var(--ox-border)',
  color: 'var(--ox-fg)', textAlign: 'left', cursor: 'pointer',
};

function Section({ title, icon, children }: { title: string; icon: string; children: React.ReactNode }) {
  return (
    <div>
      <div style={{
        padding: '8px 14px',
        fontSize: 13, fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.08em',
        color: 'var(--ox-fg-muted)', background: 'rgba(255,255,255,0.02)',
      }}>
        {icon} {title}
      </div>
      {children}
    </div>
  );
}
