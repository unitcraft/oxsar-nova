import { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '@/api/client';
import { nameOf } from '@/api/catalog';
import { useTranslation } from '@/i18n/i18n';

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

const CATEGORY_ICON: Record<Category, string> = {
  building: '🏗',
  research: '🔬',
  ship:     '🛸',
  defense:  '🛡',
  score:    '🏆',
};
const CATEGORY_KEY: Record<Category, string> = {
  building: 'catBuildings',
  research: 'catResearch',
  ship:     'catFleet',
  defense:  'catBattle',
  score:    'catTotal',
};

export function RecordsScreen() {
  const { t } = useTranslation('recordsUi');
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
      .sort((a, b) => (CATEGORY_KEY[a.category] ?? a.category).localeCompare(CATEGORY_KEY[b.category] ?? b.category));
  }, [q.data, cat, search]);

  return (
    <div style={{ maxWidth: 1000, margin: '0 auto', padding: 24, display: 'flex', flexDirection: 'column', gap: 16 }}>
      <h2 style={{ margin: 0, fontSize: 20, fontWeight: 700 }}>🏅 {t('title')}</h2>
      <p style={{ margin: 0, fontSize: 14, color: 'var(--ox-fg-dim)' }}>
        {t('colHolder')}
      </p>

      <div className="ox-tabs">
        <button type="button" aria-pressed={cat === 'all'} onClick={() => setCat('all')}>{t('filterAll')}</button>
        {(Object.keys(CATEGORY_ICON) as Category[]).map((k) => (
          <button key={k} type="button" aria-pressed={cat === k} onClick={() => setCat(k)}>{CATEGORY_ICON[k]} {t(CATEGORY_KEY[k]!)}</button>
        ))}
      </div>

      <input
        type="text"
        placeholder={`🔍 ${t('filterPlaceholder')}`}
        value={search}
        onChange={(e) => setSearch(e.target.value)}
        style={{ padding: '6px 10px', maxWidth: 300 }}
      />

      <div className="ox-panel" style={{ overflow: 'hidden' }}>
        {q.isLoading && <div style={{ padding: 20 }}><div className="ox-skeleton" style={{ height: 200 }} /></div>}
        {!q.isLoading && filtered.length === 0 && (
          <div style={{ padding: 24, textAlign: 'center', color: 'var(--ox-fg-muted)' }}>
            {t('empty')}
          </div>
        )}
        {filtered.length > 0 && (
          <div className="ox-table-responsive">
            <table className="ox-table" style={{ margin: 0 }}>
              <thead>
                <tr>
                  <th style={{ width: 80 }}>{t('colCategory')}</th>
                  <th>{t('colRank')}</th>
                  <th>{t('colHolder')}</th>
                  <th className="num">{t('colRecord')}</th>
                  <th className="num">{t('colMine')}</th>
                </tr>
              </thead>
              <tbody>
                {filtered.map((r) => {
                  const name = r.unit_id ? nameOf(r.unit_id) : r.key;
                  const catKey = CATEGORY_KEY[r.category];
                  const ratio = r.value > 0 ? r.my_value / r.value : 0;
                  return (
                    <tr key={`${r.category}-${r.key}`}>
                      <td style={{ fontSize: 14, color: 'var(--ox-fg-muted)' }}>{CATEGORY_ICON[r.category]} {catKey ? t(catKey) : r.category}</td>
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
