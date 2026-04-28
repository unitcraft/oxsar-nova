// S-016 Alliance audit-log (план 72 Ф.3 Spring 2 ч.1).
//
// Журнал событий альянса (план 67 Ф.2, U-013) — нет прямого аналога в
// legacy *.tpl, реализуем как pixel-perfect-консистентную ntable
// (R5: «pixel-perfect = визуал legacy экрана» в новом блоке трактуется
// как «использовать legacy CSS-классы и табличный layout»).
//
// Endpoint: GET /api/alliances/{id}/audit?action=&actor_id=&limit=&offset=

import { useMemo, useState } from 'react';
import { Link, Navigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { buildAuditQuery, fetchAuditLog } from '@/api/alliance';
import { QK } from '@/api/query-keys';
import { useTranslation } from '@/i18n/i18n';
import { useMyAlliance } from './common';

const PAGE_SIZE = 50;

export function AllianceAuditLogScreen() {
  const { t } = useTranslation();
  const my = useMyAlliance();
  const [actionFilter, setActionFilter] = useState('');
  const [page, setPage] = useState(0);

  const allianceID = my.data?.alliance.id ?? '';
  const qs = useMemo(
    () =>
      buildAuditQuery({
        action: actionFilter || undefined,
        limit: PAGE_SIZE,
        offset: page * PAGE_SIZE,
      }),
    [actionFilter, page],
  );

  const audit = useQuery({
    queryKey: QK.allianceAudit(allianceID, qs),
    queryFn: () => fetchAuditLog(allianceID, qs),
    enabled: !!allianceID,
    staleTime: 10_000,
  });

  if (my.isLoading) return <div className="idiv">…</div>;
  if (!my.data) return <Navigate to="/alliance" replace />;

  const al = my.data.alliance;
  const entries = audit.data?.entries ?? [];

  return (
    <>
      <div className="idiv">
        <Link to="/alliance/me">← {al.tag}</Link>
      </div>

      <table className="ntable">
        <thead>
          <tr>
            <th colSpan={4}>{t('alliance', 'audit.title')}</th>
          </tr>
          <tr>
            <td colSpan={4}>
              <label htmlFor="action">
                {t('alliance', 'audit.filterAction')}
              </label>{' '}
              <input
                type="text"
                id="action"
                value={actionFilter}
                onChange={(e) => {
                  setActionFilter(e.target.value);
                  setPage(0);
                }}
              />
            </td>
          </tr>
          <tr>
            <th>{t('alliance', 'audit.colWhen')}</th>
            <th>{t('alliance', 'audit.colActor')}</th>
            <th>{t('alliance', 'audit.colAction')}</th>
            <th>{t('alliance', 'audit.colTarget')}</th>
          </tr>
        </thead>
        <tbody>
          {entries.length === 0 && (
            <tr>
              <td colSpan={4} className="center">
                {t('alliance', 'audit.empty')}
              </td>
            </tr>
          )}
          {entries.map((e) => (
            <tr key={e.id}>
              <td>{new Date(e.created_at).toLocaleString('ru-RU')}</td>
              <td>{e.actor_name || t('alliance', 'audit.actorSystem')}</td>
              <td>
                {t(
                  'alliance',
                  `audit.action.${e.action}`,
                ).startsWith('[')
                  ? e.action
                  : t('alliance', `audit.action.${e.action}`)}
              </td>
              <td>{e.target_name ?? '—'}</td>
            </tr>
          ))}
        </tbody>
        {entries.length > 0 && (
          <tfoot>
            <tr>
              <td colSpan={4} className="center">
                <input
                  type="button"
                  className="button"
                  value={`← ${t('alliance', 'audit.prevPage')}`}
                  disabled={page === 0}
                  onClick={() => setPage((p) => Math.max(0, p - 1))}
                />{' '}
                <input
                  type="button"
                  className="button"
                  value={`${t('alliance', 'audit.nextPage')} →`}
                  disabled={entries.length < PAGE_SIZE}
                  onClick={() => setPage((p) => p + 1)}
                />
              </td>
            </tr>
          </tfoot>
        )}
      </table>
    </>
  );
}
