// S-009 / S-019 Alliance list + search (план 72 Ф.3 Spring 2 ч.1).
//
// Pixel-perfect зеркало legacy `templates/standard/allysearch.tpl` +
// `ally_search_result.tpl` (search input + результаты в ntable).
//
// Расширения относительно legacy-PHP:
//   - Полнотекстовый поиск + фильтры (план 67 Ф.4, U-012):
//     q, is_open, min_members/max_members. UI добавляет один input для
//     query + чекбокс «только открытые», остальные фильтры скрыты до
//     явной нужды (R5: визуально остаёмся близко к legacy).
//   - Клик по строке → /alliance/{id} (профиль альянса).

import { useEffect, useMemo, useState } from 'react';
import { Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { buildSearchQuery, fetchAllianceList } from '@/api/alliance';
import { QK } from '@/api/query-keys';
import { useTranslation } from '@/i18n/i18n';

export function AllianceListScreen() {
  const { t } = useTranslation();
  const [query, setQuery] = useState('');
  const [onlyOpen, setOnlyOpen] = useState(false);
  const [debouncedQuery, setDebouncedQuery] = useState('');

  // Debounce 300ms — не дёргать backend на каждый keystroke.
  useEffect(() => {
    const tid = window.setTimeout(() => setDebouncedQuery(query), 300);
    return () => window.clearTimeout(tid);
  }, [query]);

  const qs = useMemo(
    () =>
      buildSearchQuery({
        q: debouncedQuery || undefined,
        is_open: onlyOpen ? true : undefined,
        limit: 50,
      }),
    [debouncedQuery, onlyOpen],
  );

  const list = useQuery({
    queryKey: QK.alliancesSearch(qs),
    queryFn: () => fetchAllianceList(qs),
    staleTime: 15_000,
  });

  const alliances = list.data?.alliances ?? [];

  return (
    <>
      <table className="ntable">
        <thead>
          <tr>
            <th>{t('alliance', 'allianceSearch')}</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>
              <input
                type="text"
                value={query}
                maxLength={64}
                onChange={(e) => setQuery(e.target.value)}
                placeholder={t('alliance', 'name')}
              />{' '}
              <label>
                <input
                  type="checkbox"
                  checked={onlyOpen}
                  onChange={(e) => setOnlyOpen(e.target.checked)}
                />{' '}
                {t('alliance', 'enableApplications')}
              </label>
            </td>
          </tr>
        </tbody>
      </table>

      {alliances.length > 0 ? (
        <table className="ntable">
          <thead>
            <tr>
              <th colSpan={4}>{t('alliance', 'searchResults')}</th>
            </tr>
            <tr>
              <th>{t('alliance', 'tag')}</th>
              <th>{t('alliance', 'name')}</th>
              <th>{t('alliance', 'members')}</th>
              <th>{t('alliance', 'join')}</th>
            </tr>
          </thead>
          <tbody>
            {alliances.map((al) => (
              <tr key={al.id}>
                <td>
                  <Link to={`/alliance/${al.id}`}>[{al.tag}]</Link>
                </td>
                <td>{al.name}</td>
                <td className="center">{al.member_count}</td>
                <td className="center">
                  {al.is_open
                    ? t('alliance', 'labelOpen')
                    : t('alliance', 'labelClosed')}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      ) : (
        <div className="idiv">
          {list.isLoading ? '…' : t('alliance', 'noMatchesFound') || '—'}
        </div>
      )}
    </>
  );
}
