// S-R04 Stock — биржа артефактов (план 72.1 ч.19).
// Pixel-perfect клон legacy stock.tpl.

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
    <table className="ntable">
      <colgroup>
        <col width="1px" />
        <col />
        <col width="10%" />
        <col width="10%" />
        <col width="10%" />
        <col width="10%" />
      </colgroup>
      <thead>
        <tr>
          <th style={{ textAlign: 'center' }} colSpan={6}>
            <div style={{ float: 'right' }}>
              {cursor && (
                <button
                  type="button"
                  className="button"
                  style={{ marginRight: 4 }}
                  onClick={() => setCursor(undefined)}
                >
                  ◀ В начало
                </button>
              )}
              {nextCursor && (
                <button
                  type="button"
                  className="button"
                  onClick={() => setCursor(nextCursor)}
                  disabled={lotsQ.isFetching}
                >
                  Вперёд ▶
                </button>
              )}
            </div>
            Биржа
          </th>
        </tr>
        <tr className="center">
          <th colSpan={2}>Лот</th>
          <th>Количество</th>
          <th>Цена</th>
          <th style={{ textAlign: 'right' }}>Продавец</th>
          <th>&nbsp;</th>
        </tr>
      </thead>

      <tfoot>
        <tr>
          <th style={{ textAlign: 'center' }} colSpan={6}>
            <div style={{ float: 'right' }}>
              {cursor && (
                <button
                  type="button"
                  className="button"
                  style={{ marginRight: 4 }}
                  onClick={() => setCursor(undefined)}
                >
                  ◀ В начало
                </button>
              )}
              {nextCursor && (
                <button
                  type="button"
                  className="button"
                  onClick={() => setCursor(nextCursor)}
                  disabled={lotsQ.isFetching}
                >
                  Вперёд ▶
                </button>
              )}
            </div>
          </th>
        </tr>
      </tfoot>

      <tbody>
        {lotsQ.isLoading && (
          <tr>
            <td colSpan={6} className="center">Загрузка…</td>
          </tr>
        )}
        {!lotsQ.isLoading && lots.length === 0 && (
          <tr>
            <td colSpan={6} className="center">Активных лотов нет</td>
          </tr>
        )}
        {lots.map((lot, i) => (
          <tr key={lot.id}>
            <td align="center">{i + 1}.</td>
            <td>
              Арт. #{lot.artifact_unit_id}
            </td>
            <td align="right">
              {formatNumber(lot.quantity)}
              {lot.quantity > 1 && (
                <>
                  <br />
                  <span style={{ fontSize: 'smaller' }}>мин: 1</span>
                </>
              )}
            </td>
            <td align="right">
              {formatNumber(lot.price_oxsarit)}
              {lot.quantity > 1 && (
                <>
                  <br />
                  {formatNumber(
                    lot.unit_price_oxsarit ??
                      Math.round(lot.price_oxsarit / lot.quantity),
                  )}
                  /шт.
                </>
              )}
            </td>
            <td align="right">
              {lot.seller_username ?? '—'}
              <br />
              истекает: {fmtDate(lot.expires_at)}
            </td>
            <td align="center" valign="middle">
              <input
                type="button"
                className="button"
                value="Купить"
                disabled
                title="Покупка недоступна"
              />
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}
