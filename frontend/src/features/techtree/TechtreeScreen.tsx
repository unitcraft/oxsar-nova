import { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '@/api/client';
import { nameOf } from '@/api/catalog';
import { TechtreeGraph } from './TechtreeGraph';

type NodeKind = 'building' | 'research' | 'ship' | 'defense';

interface Requirement {
  kind: string;
  key: string;
  level: number;
  have: number;
  met: boolean;
}
interface TechNode {
  key: string;
  kind: NodeKind;
  id: number;
  current_level: number;
  unlocked: boolean;
  requirements: Requirement[];
}
interface TechtreeData {
  nodes: TechNode[];
}

const KIND_META: Record<NodeKind, { label: string; icon: string }> = {
  building: { label: 'Постройки',     icon: '🏗' },
  research: { label: 'Исследования',  icon: '🔬' },
  ship:     { label: 'Флот',          icon: '🛸' },
  defense:  { label: 'Оборона',       icon: '🛡' },
};

type Filter = 'all' | 'unlocked' | 'locked';

export function TechtreeScreen() {
  const q = useQuery({
    queryKey: ['techtree'],
    queryFn: () => api.get<TechtreeData>('/api/techtree'),
  });

  const [kind, setKind] = useState<NodeKind>('research');
  const [filter, setFilter] = useState<Filter>('all');
  const [search, setSearch] = useState('');
  const [view, setView] = useState<'cards' | 'graph'>('cards');

  const filtered = useMemo(() => {
    const all = q.data?.nodes ?? [];
    return all
      .filter((n) => n.kind === kind)
      .filter((n) => {
        if (filter === 'unlocked') return n.unlocked;
        if (filter === 'locked') return !n.unlocked;
        return true;
      })
      .filter((n) => {
        if (!search.trim()) return true;
        return n.key.toLowerCase().includes(search.toLowerCase())
          || nameOf(n.id).toLowerCase().includes(search.toLowerCase());
      })
      .sort((a, b) => {
        if (a.unlocked !== b.unlocked) return a.unlocked ? -1 : 1;
        return a.key.localeCompare(b.key);
      });
  }, [q.data, kind, filter, search]);

  return (
    <div style={{ maxWidth: 1000, margin: '0 auto', padding: 24, display: 'flex', flexDirection: 'column', gap: 16 }}>
      <h2 style={{ margin: 0, fontSize: 20, fontWeight: 700 }}>🌳 Дерево технологий</h2>

      <div className="ox-tabs">
        {(Object.entries(KIND_META) as [NodeKind, typeof KIND_META[NodeKind]][]).map(([k, m]) => (
          <button key={k} type="button" aria-pressed={kind === k} onClick={() => setKind(k)}>
            {m.icon} {m.label}
          </button>
        ))}
      </div>

      <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap', alignItems: 'center' }}>
        <input
          type="text"
          placeholder="Поиск по названию…"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          style={{ flex: 1, minWidth: 200, padding: '6px 10px' }}
        />
        <div style={{ display: 'flex', gap: 4 }}>
          <button type="button" className={filter === 'all' ? 'btn btn-sm' : 'btn-ghost btn-sm'} onClick={() => setFilter('all')}>Все</button>
          <button type="button" className={filter === 'unlocked' ? 'btn btn-sm' : 'btn-ghost btn-sm'} onClick={() => setFilter('unlocked')}>✓ Доступно</button>
          <button type="button" className={filter === 'locked' ? 'btn btn-sm' : 'btn-ghost btn-sm'} onClick={() => setFilter('locked')}>🔒 Закрыто</button>
        </div>
        <div style={{ display: 'flex', gap: 4 }}>
          <button type="button" className={view === 'cards' ? 'btn btn-sm' : 'btn-ghost btn-sm'} onClick={() => setView('cards')} title="Карточки">🗂 Карточки</button>
          <button type="button" className={view === 'graph' ? 'btn btn-sm' : 'btn-ghost btn-sm'} onClick={() => setView('graph')} title="Граф">🌐 Граф</button>
        </div>
      </div>

      {q.isLoading && <div className="ox-skeleton" style={{ height: 400, borderRadius: 8 }} />}

      {!q.isLoading && filtered.length === 0 && (
        <div style={{ padding: 24, textAlign: 'center', color: 'var(--ox-fg-muted)' }}>
          Ничего не найдено
        </div>
      )}

      {view === 'cards' ? (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))', gap: 10 }}>
          {filtered.map((n) => <TechCard key={`${n.kind}-${n.key}`} node={n} />)}
        </div>
      ) : (
        <TechtreeGraph nodes={q.data?.nodes ?? []} kind={kind} />
      )}
    </div>
  );
}

function TechCard({ node }: { node: TechNode }) {
  const name = nameOf(node.id) || node.key;
  const showLevel = node.kind === 'building' || node.kind === 'research';

  return (
    <div
      className="ox-panel"
      style={{
        padding: 14,
        opacity: node.unlocked ? 1 : 0.75,
        borderLeft: `3px solid ${node.unlocked ? 'var(--ox-success)' : 'var(--ox-fg-muted)'}`,
      }}
    >
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 6 }}>
        <div style={{ fontWeight: 600 }}>{name}</div>
        {showLevel && node.current_level > 0 && (
          <span style={{ fontFamily: 'var(--ox-mono)', fontSize: 14, color: 'var(--ox-accent)' }}>
            Ур. {node.current_level}
          </span>
        )}
        {!showLevel && node.unlocked && (
          <span style={{ fontSize: 13, color: 'var(--ox-success)' }}>✓ доступно</span>
        )}
      </div>

      {node.requirements.length === 0 ? (
        <div style={{ fontSize: 13, color: 'var(--ox-fg-muted)', fontStyle: 'italic' }}>
          Без предусловий
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 3 }}>
          {node.requirements.map((req, i) => (
            <div
              key={`${req.kind}-${req.key}-${i}`}
              style={{
                fontSize: 13,
                color: req.met ? 'var(--ox-fg-dim)' : 'var(--ox-danger)',
                fontFamily: 'var(--ox-mono)',
                display: 'flex',
                gap: 4,
              }}
            >
              <span>{req.met ? '✓' : '✗'}</span>
              <span style={{ flex: 1 }}>
                {req.kind === 'building' ? '🏗' : '🔬'} {req.key}
              </span>
              <span>
                {req.have}/{req.level}
              </span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
