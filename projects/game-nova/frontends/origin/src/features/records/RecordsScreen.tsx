// S-024 Records (план 72 Ф.4 Spring 3).
//
// Pixel-perfect зеркало legacy `templates/standard/records.tpl`:
// рекорды по категориям (стройки / исследования / флот / оборона /
// score) с держателем рекорда, значением и личным результатом игрока.
//
// Endpoint:
//   GET /api/records  → { records: RecordEntry[] }

import { Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { fetchRecords } from '@/api/catalog';
import { QK } from '@/api/query-keys';
import type { RecordEntry } from '@/api/types';
import { useTranslation } from '@/i18n/i18n';
import { formatNumber } from '@/lib/format';

interface SectionDef {
  category: RecordEntry['category'];
  titleKey: string;
  routePrefix: '/building' | '/unit' | null;
}

const SECTIONS: SectionDef[] = [
  { category: 'score', titleKey: 'colMine', routePrefix: null },
  { category: 'building', titleKey: 'kindBuildings', routePrefix: '/building' },
  { category: 'research', titleKey: 'kindResearch', routePrefix: '/unit' },
  { category: 'ship', titleKey: 'kindShips', routePrefix: '/unit' },
  { category: 'defense', titleKey: 'kindDefense', routePrefix: '/unit' },
];

export function RecordsScreen() {
  const { t } = useTranslation();
  const q = useQuery({
    queryKey: QK.records(),
    queryFn: fetchRecords,
    staleTime: 60_000,
  });

  if (q.isLoading) return <div className="idiv">…</div>;

  const records = q.data?.records ?? [];

  return (
    <table className="ntable">
      <thead>
        <tr>
          <th colSpan={2}>{t('records', 'title')}</th>
          <th>{t('score', 'colPlayer')}</th>
          <th className="center">{t('records', 'colRecord')}</th>
          <th className="center">{t('records', 'colMine')}</th>
        </tr>
      </thead>
      <tbody>
        {records.length === 0 && (
          <tr>
            <td colSpan={5} className="center">
              {t('records', 'empty')}
            </td>
          </tr>
        )}
        {SECTIONS.map((section) => {
          const items = records.filter((r) => r.category === section.category);
          if (items.length === 0) return null;
          return (
            <SectionRows
              key={section.category}
              titleKey={section.titleKey}
              items={items}
              routePrefix={section.routePrefix}
            />
          );
        })}
      </tbody>
    </table>
  );
}

interface SectionRowsProps {
  titleKey: string;
  items: RecordEntry[];
  routePrefix: '/building' | '/unit' | null;
}

function SectionRows({ titleKey, items, routePrefix }: SectionRowsProps) {
  const { t } = useTranslation();
  return (
    <>
      <tr>
        <th colSpan={5}>{t(titleKey === 'colMine' ? 'records' : 'techtree', titleKey === 'colMine' ? 'catTotal' : titleKey)}</th>
      </tr>
      {items.map((r) => {
        const i18nKey = r.key.replace(/_([a-z])/g, (_, c: string) =>
          c.toUpperCase(),
        );
        const name = t('info', i18nKey);
        const display = name !== `[info.${i18nKey}]` ? name : r.key;
        return (
          <tr key={`${r.category}-${r.key}`}>
            <td style={{ width: '1%' }}>
              {r.unit_id ? `#${r.unit_id}` : '—'}
            </td>
            <td>
              {routePrefix && r.unit_id ? (
                <Link to={`${routePrefix}/${r.unit_id}`}>{display}</Link>
              ) : (
                display
              )}
            </td>
            <td>{r.holder_name || '—'}</td>
            <td className="center">{formatNumber(r.value)}</td>
            <td className="center">{formatNumber(r.my_value)}</td>
          </tr>
        );
      })}
    </>
  );
}
