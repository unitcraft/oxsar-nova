// S-P2P P2P Exchange — статистика лотов брокера (план 72.1 §20.12).
//
// Pixel-perfect клон legacy `templates/standard/exchange.tpl`:
//  - Header: фильтры date_min, date_max + сортировка.
//  - Summary: total / sold / turnover / profit.
//  - Таблица лотов: дата / lot / amount / price / status / profit.
//  - Pagination: ◀ N/M ▶ (USER_PER_PAGE=25).

import { useState } from 'react';
import { Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { fetchBrokerStats } from '@/api/p2pExchange';
import type { BrokerStatsFilters } from '@/api/p2pExchange';
import { useTranslation } from '@/i18n/i18n';
import { formatNumber } from '@/lib/format';

export function P2PExchangeScreen() {
  const { t } = useTranslation();
  const today = new Date();
  const monthAgo = new Date(today);
  monthAgo.setDate(today.getDate() - 30);
  const fmtDate = (d: Date) => d.toISOString().slice(0, 10);

  const [dateMin, setDateMin] = useState(fmtDate(monthAgo));
  const [dateMax, setDateMax] = useState(fmtDate(today));
  const [sortField, setSortField] =
    useState<NonNullable<BrokerStatsFilters['sort_field']>>('date');
  const [sortOrder, setSortOrder] =
    useState<NonNullable<BrokerStatsFilters['sort_order']>>('desc');
  const [page, setPage] = useState(1);

  const q = useQuery({
    queryKey: ['broker-stats', dateMin, dateMax, sortField, sortOrder, page],
    queryFn: () =>
      fetchBrokerStats({
        date_min: dateMin,
        date_max: dateMax,
        sort_field: sortField,
        sort_order: sortOrder,
        page,
      }),
  });

  // Resolve unit_id → name через i18n info.* (артефакты нумерованы
  // в одном пространстве с ships/defense). Если ключа нет, показываем #ID.
  const unitName = (unitID: number): string => {
    return `#${unitID}`;
  };

  const data = q.data;
  const rows = data?.rows ?? [];
  const summary = data?.summary;
  const pages = data?.pages ?? 1;
  const fee = data?.fee ?? 5;

  return (
    <>
      <div className="idiv" style={{ textAlign: 'right' }}>
        {/* План 72.1.45 §9: ссылка на admin-страницу настроек брокера. */}
        <Link to="/p2p-exchange/opts">
          ⚙ {t('exchangeOpts', 'title')}
        </Link>
      </div>
      <table className="ntable">
        <thead>
          <tr>
            <th colSpan={4}>
              {data?.title || t('p2pExchange', 'title')}
            </th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>{t('statistics', 'dateFirst')}:</td>
            <td>
              <input
                type="date"
                value={dateMin}
                onChange={(e) => {
                  setDateMin(e.target.value);
                  setPage(1);
                }}
              />
            </td>
            <td>{t('statistics', 'dateLast')}:</td>
            <td>
              <input
                type="date"
                value={dateMax}
                onChange={(e) => {
                  setDateMax(e.target.value);
                  setPage(1);
                }}
              />
            </td>
          </tr>
          <tr>
            <td>{t('p2pExchange', 'sortField')}:</td>
            <td>
              <select
                value={sortField}
                onChange={(e) =>
                  setSortField(e.target.value as typeof sortField)
                }
              >
                <option value="date">
                  {t('p2pExchange', 'colDate')}
                </option>
                <option value="lot">
                  {t('p2pExchange', 'colLot')}
                </option>
                <option value="lot_amount">
                  {t('p2pExchange', 'colAmount')}
                </option>
                <option value="lot_price">
                  {t('p2pExchange', 'colPrice')}
                </option>
                <option value="lot_profit">
                  {t('p2pExchange', 'colProfit')}
                </option>
              </select>
            </td>
            <td colSpan={2}>
              <select
                value={sortOrder}
                onChange={(e) =>
                  setSortOrder(e.target.value as typeof sortOrder)
                }
              >
                <option value="desc">
                  {t('statistics', 'bsSortDesc')}
                </option>
                <option value="asc">
                  {t('statistics', 'bsSortAsc')}
                </option>
              </select>
            </td>
          </tr>
        </tbody>
      </table>

      {summary && (
        <table className="ntable">
          <thead>
            <tr>
              <th colSpan={4}>
                {t('p2pExchange', 'summary')}
              </th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td>{t('p2pExchange', 'totalLots')}:</td>
              <td>
                <b>{formatNumber(summary.total)}</b>
              </td>
              <td>{t('p2pExchange', 'soldLots')}:</td>
              <td>
                <b className="true">{formatNumber(summary.sold)}</b>
              </td>
            </tr>
            <tr>
              <td>{t('p2pExchange', 'turnover')}:</td>
              <td>
                <b>{formatNumber(summary.turnover)}</b>
              </td>
              <td>
                {t('p2pExchange', 'profit')} ({fee}%):
              </td>
              <td>
                <b className="true">{formatNumber(Math.round(summary.profit))}</b>
              </td>
            </tr>
          </tbody>
        </table>
      )}

      <table className="ntable">
        <thead>
          <tr>
            <th>{t('p2pExchange', 'colDate')}</th>
            <th>{t('p2pExchange', 'colLot')}</th>
            <th>{t('p2pExchange', 'colAmount')}</th>
            <th>{t('p2pExchange', 'colPrice')}</th>
            <th>{t('p2pExchange', 'colStatus')}</th>
            <th>{t('p2pExchange', 'colProfit')}</th>
          </tr>
        </thead>
        <tbody>
          {q.isLoading && (
            <tr>
              <td colSpan={6} className="center">
                …
              </td>
            </tr>
          )}
          {!q.isLoading && rows.length === 0 && (
            <tr>
              <td colSpan={6} className="center">
                <i>{t('alliance', 'nothing')}</i>
              </td>
            </tr>
          )}
          {rows.map((row) => (
            <tr key={row.lot_id}>
              <td className="center">
                {new Date(row.sold_at).toLocaleString('ru-RU')}
              </td>
              <td>{unitName(row.unit_id)}</td>
              <td className="center">{formatNumber(row.quantity)}</td>
              <td className="center">{formatNumber(row.price)}</td>
              <td className="center">
                <span
                  className={
                    row.status === 'sold'
                      ? 'true'
                      : row.status === 'cancelled' || row.status === 'expired'
                        ? 'false'
                        : ''
                  }
                >
                  {t('p2pExchange', `status_${row.status}`)}
                </span>
              </td>
              <td className="center">
                {row.status === 'sold' ? (
                  <span className="true">{formatNumber(Math.round(row.profit))}</span>
                ) : (
                  '0'
                )}
              </td>
            </tr>
          ))}
        </tbody>
        {pages > 1 && (
          <tfoot>
            <tr>
              <td colSpan={6} className="center">
                <button
                  type="button"
                  className="button"
                  disabled={page <= 1 || q.isFetching}
                  onClick={() => setPage((p) => Math.max(1, p - 1))}
                >
                  ◀
                </button>
                {' '}
                {page} / {pages}
                {' '}
                <button
                  type="button"
                  className="button"
                  disabled={page >= pages || q.isFetching}
                  onClick={() => setPage((p) => Math.min(pages, p + 1))}
                >
                  ▶
                </button>
              </td>
            </tr>
          </tfoot>
        )}
      </table>
    </>
  );
}
