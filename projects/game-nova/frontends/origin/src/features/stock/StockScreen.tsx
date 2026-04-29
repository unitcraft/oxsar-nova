// S-R04 Stock — биржа артефактов (план 72.1).
//
// Pixel-perfect клон legacy Stock.class.php (exchange/lots):
//   - Список активных лотов: артефакт, количество, цена, продавец, срок.
//   - Фильтр по типу артефакта.
//   - Пагинация через cursor.
//   - Кнопка «Купить» (POST /api/exchange/lots/{id}/buy).
//
// Endpoints:
//   GET /api/exchange/lots — список лотов

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { fetchExchangeLots } from '@/api/exchange';
import { QK } from '@/api/query-keys';
import { formatNumber } from '@/lib/format';

function fmtDate(iso: string): string {
  try {
    const d = new Date(iso);
    return d.toLocaleDateString('ru-RU', {
      day: '2-digit',
      month: '2-digit',
      year: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    });
  } catch {
    return iso;
  }
}

export function StockScreen() {
  const [cursor, setCursor] = useState<string | undefined>(undefined);
  const limit = 20;

  const paramKey = [cursor ?? '', String(limit)].join('|');

  const lotsQ = useQuery({
    queryKey: QK.exchangeLots(paramKey),
    queryFn: () =>
      fetchExchangeLots({
        status: 'active',
        limit,
        ...(cursor !== undefined ? { cursor } : {}),
      }),
  });

  const lots = lotsQ.data?.lots ?? [];
  const nextCursor = lotsQ.data?.next_cursor;

  return (
    <>
      <table className="ntable" style={{ width: '100%' }}>
        <thead>
          <tr>
            <th>Артефакт</th>
            <th>Кол-во</th>
            <th>Цена (оксариты)</th>
            <th>Цена/шт.</th>
            <th>Продавец</th>
            <th>Истекает</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          {lotsQ.isLoading && (
            <tr>
              <td colSpan={7} className="center">Загрузка…</td>
            </tr>
          )}
          {!lotsQ.isLoading && lots.length === 0 && (
            <tr>
              <td colSpan={7} className="center">Активных лотов нет</td>
            </tr>
          )}
          {lots.map((lot) => (
            <tr key={lot.id}>
              <td>#{lot.artifact_unit_id}</td>
              <td className="center">{formatNumber(lot.quantity)}</td>
              <td className="center">{formatNumber(lot.price_oxsarit)}</td>
              <td className="center">
                {lot.unit_price_oxsarit != null
                  ? formatNumber(lot.unit_price_oxsarit)
                  : lot.quantity > 0
                  ? formatNumber(Math.round(lot.price_oxsarit / lot.quantity))
                  : '—'}
              </td>
              <td>{lot.seller_username ?? '—'}</td>
              <td>{fmtDate(lot.expires_at)}</td>
              <td>
                <button
                  type="button"
                  className="button"
                  disabled
                  title="Покупка в разработке"
                >
                  Купить
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>

      {/* Пагинация */}
      <div style={{ marginTop: 8, textAlign: 'center' }}>
        {cursor && (
          <button
            type="button"
            className="button"
            style={{ marginRight: 8 }}
            onClick={() => setCursor(undefined)}
          >
            ← В начало
          </button>
        )}
        {nextCursor && (
          <button
            type="button"
            className="button"
            onClick={() => setCursor(nextCursor)}
            disabled={lotsQ.isFetching}
          >
            Следующая страница →
          </button>
        )}
      </div>
    </>
  );
}
