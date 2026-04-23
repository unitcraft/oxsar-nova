import { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '@/api/client';
import { nameOf } from '@/api/catalog';

type Category = 'building' | 'research' | 'ship' | 'defense' | 'score';

interface ServerRecord {
  category: Category;
  key: string;
  unit_id?: number;
  holder_id: string;
  holder_name: string;
  value: number;
  my_value: number;
}

const CATEGORY_META: Record<Category, { label: string; icon: string }> = {
  building: { label: 'Постройки',    icon: '🏗' },
  research: { label: 'Исследования', icon: '🔬' },
  ship:     { label: 'Флот',         icon: '🛸' },
  defense:  { label: 'Оборона',      icon: '🛡' },
  score:    { label: 'Очки',         icon: '🏆' },
};

export function RecordsScreen() {
  const [cat, setCat] = useState<Category | 'all'>('all');
  const [search, setSearch] = useState('');

  const q = useQuery({
    queryKey: ['records'],
    queryFn: () => api.get<{ records: ServerRecord[] }>('/api/records'),
    staleTime: 60000,
  });

  const filtered = useMemo(() => {
    const all = q.data?.records ?? [];
    return all
      .filter((r) => cat === 'all' || r.category === cat)
      .filter((r) => {
        if (!search.trim()) return true;
        const name = r.unit_id ? nameOf(r.unit_id) : r.key;
        return name.toLowerCase().includes(search.toLowerCase()) ||
          r.holder_name.toLowerCase().includes(search.toLowerCase());
      })
      .sort((a, b) => (CATEGORY_META[a.category]?.label ?? a.category).localeCompare(CATEGORY_META[b.category]?.label ?? b.category));
  }, [q.data, cat, search]);

  return (
    <div style={{ maxWidth: 1000, margin: '0 auto', padding: 24, display: 'flex', flexDirection: 'column', gap: 16 }}>
      <h2 style={{ margin: 0, fontSize: 20, fontWeight: 700 }}>🏅 Рекорды сервера</h2>
      <p style={{ margin: 0, fontSize: 12, color: 'var(--ox-fg-dim)' }}>
        Топ-1 игрок по каждой категории + ваш текущий показатель.
      </p>

      <div className="ox-tabs">
        <button type="button" aria-pressed={cat === 'all'} onClick={() => setCat('all')}>📊 Все</button>
        {(Object.entries(CATEGORY_META) as [Category, typeof CATEGORY_META[Category]][]).map(([k, m]) => (
          <button key={k} type="button" aria-pressed={cat === k} onClick={() => setCat(k)}>{m.icon} {m.label}</button>
        ))}
      </div>

      <input
        type="text"
        placeholder="🔍 Поиск…"
        value={search}
        onChange={(e) => setSearch(e.target.value)}
        style={{ padding: '6px 10px', maxWidth: 300 }}
      />

      <div className="ox-panel" style={{ overflow: 'hidden' }}>
        {q.isLoading && <div style={{ padding: 20 }}><div className="ox-skeleton" style={{ height: 200 }} /></div>}
        {!q.isLoading && filtered.length === 0 && (
          <div style={{ padding: 24, textAlign: 'center', color: 'var(--ox-fg-muted)' }}>
            Рекордов пока нет
          </div>
        )}
        {filtered.length > 0 && (
          <div className="ox-table-responsive">
            <table className="ox-table" style={{ margin: 0 }}>
              <thead>
                <tr>
                  <th style={{ width: 80 }}>Категория</th>
                  <th>Позиция</th>
                  <th>Держатель</th>
                  <th className="num">Рекорд</th>
                  <th className="num">Мой результат</th>
                </tr>
              </thead>
              <tbody>
                {filtered.map((r) => {
                  const name = r.unit_id ? nameOf(r.unit_id) : r.key;
                  const meta = CATEGORY_META[r.category];
                  const ratio = r.value > 0 ? r.my_value / r.value : 0;
                  return (
                    <tr key={`${r.category}-${r.key}`}>
                      <td style={{ fontSize: 12, color: 'var(--ox-fg-muted)' }}>{meta?.icon} {meta?.label}</td>
                      <td style={{ fontWeight: 500 }}>{name}</td>
                      <td style={{ fontWeight: 600 }}>🏆 {r.holder_name}</td>
                      <td className="num" style={{ color: 'var(--ox-accent)', fontWeight: 700 }}>
                        {Math.round(r.value).toLocaleString('ru-RU')}
                      </td>
                      <td className="num" style={{ color: ratio >= 1 ? 'var(--ox-success)' : 'var(--ox-fg-dim)' }}>
                        {Math.round(r.my_value).toLocaleString('ru-RU')}
                        {r.value > 0 && r.my_value > 0 && (
                          <span style={{ fontSize: 10, marginLeft: 4, color: 'var(--ox-fg-muted)' }}>
                            ({(ratio * 100).toFixed(0)}%)
                          </span>
                        )}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
