// S-R04 Stock — биржа артефактов (план 72.1 ч.19 + 72.1.8).
// Pixel-perfect клон legacy stock.tpl.
//
// План 72.1.8 ч.A: подключены mutations Buy и Cancel — раньше кнопка
// «Купить» была disabled с пометкой «Покупка недоступна», recall не
// существовал. Backend (internal/exchange/) был готов давно.
// План 72.1.8 ч.B: добавлена inline-форма «Выставить лот»
// (CreateLot) — раньше юзер не мог продавать артефакты вообще.

import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  buyLot,
  cancelLot,
  createLot,
  fetchExchangeLots,
} from '@/api/exchange';
import { fetchArtefacts } from '@/api/artefacts';
import { QK } from '@/api/query-keys';
import type { ApiError } from '@/api/client';
import { useAuthStore } from '@/stores/auth';
import { formatNumber } from '@/lib/format';
import { useTranslation } from '@/i18n/i18n';

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
  const { t } = useTranslation();
  const qc = useQueryClient();
  const myUserId = useAuthStore((s) => s.userId);
  const [cursor, setCursor] = useState<string | undefined>(undefined);
  const [actionMsg, setActionMsg] = useState<string | null>(null);
  const [actionErr, setActionErr] = useState<string | null>(null);
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

  const buyMut = useMutation({
    mutationFn: (lotID: string) => buyLot(lotID),
    onSuccess: () => {
      setActionMsg(t('stock', 'buySuccess'));
      setActionErr(null);
      void qc.invalidateQueries({ queryKey: QK.exchangeLots(paramKey) });
      void qc.invalidateQueries({ queryKey: QK.me() });
      setTimeout(() => setActionMsg(null), 3000);
    },
    onError: (e) => {
      setActionErr((e as ApiError).message);
      setActionMsg(null);
    },
  });

  const cancelMut = useMutation({
    mutationFn: (lotID: string) => cancelLot(lotID),
    onSuccess: () => {
      setActionMsg(t('stock', 'cancelSuccess'));
      setActionErr(null);
      void qc.invalidateQueries({ queryKey: QK.exchangeLots(paramKey) });
      setTimeout(() => setActionMsg(null), 3000);
    },
    onError: (e) => {
      setActionErr((e as ApiError).message);
      setActionMsg(null);
    },
  });

  // План 72.1.8 ч.B: создание лота. Список юзерских артефактов
  // фетчится только когда форма открыта (lazy через enabled).
  const [createOpen, setCreateOpen] = useState(false);
  const [cArt, setCArt] = useState<number | ''>('');
  const [cQty, setCQty] = useState('1');
  const [cPrice, setCPrice] = useState('');
  const [cExp, setCExp] = useState('72'); // дефолт 72 ч (3 суток).

  const myArtsQ = useQuery({
    queryKey: QK.artefacts(),
    queryFn: fetchArtefacts,
    enabled: createOpen,
  });

  const createMut = useMutation({
    mutationFn: createLot,
    onSuccess: () => {
      setActionMsg(t('stock', 'createSuccess'));
      setActionErr(null);
      setCreateOpen(false);
      setCArt('');
      setCQty('1');
      setCPrice('');
      setCExp('72');
      void qc.invalidateQueries({ queryKey: QK.exchangeLots(paramKey) });
      void qc.invalidateQueries({ queryKey: QK.artefacts() });
      setTimeout(() => setActionMsg(null), 3000);
    },
    onError: (e) => {
      setActionErr((e as ApiError).message);
      setActionMsg(null);
    },
  });

  function onCreateSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (cArt === '') return;
    const qty = Math.max(1, Math.floor(Number(cQty) || 0));
    const price = Math.max(1, Math.floor(Number(cPrice) || 0));
    const exp = Math.max(1, Math.floor(Number(cExp) || 0));
    createMut.mutate({
      artifact_unit_id: cArt,
      quantity: qty,
      price_oxsarit: price,
      expires_in_hours: exp,
    });
  }

  const lots = lotsQ.data?.lots ?? [];
  const nextCursor = lotsQ.data?.next_cursor;

  // Список уникальных unit_id артефактов в инвентаре. На продажу
  // допустимы только не-активные и не-просроченные — state=held.
  // Active/delayed/expired/consumed — escrow в exchange отвергает.
  const myArtefactsByUnit: { unit_id: number; count: number }[] = (() => {
    const m = new Map<number, number>();
    for (const a of myArtsQ.data?.artefacts ?? []) {
      if (a.state !== 'held') continue;
      m.set(a.unit_id, (m.get(a.unit_id) ?? 0) + 1);
    }
    return Array.from(m.entries()).map(([unit_id, count]) => ({ unit_id, count }));
  })();

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
              <button
                type="button"
                className="button"
                style={{ marginRight: 4 }}
                onClick={() => setCreateOpen((v) => !v)}
              >
                {createOpen ? t('stock', 'createCancel') : t('stock', 'createBtn')}
              </button>
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
        {createOpen && (
          <tr>
            <td colSpan={6}>
              <form
                onSubmit={onCreateSubmit}
                style={{ display: 'flex', gap: 8, alignItems: 'center', flexWrap: 'wrap', padding: 4 }}
              >
                <label>
                  {t('stock', 'createArtLabel')}:{' '}
                  <select
                    value={cArt}
                    onChange={(e) => setCArt(e.target.value === '' ? '' : Number(e.target.value))}
                    required
                  >
                    <option value="">— {t('stock', 'createArtPlaceholder')} —</option>
                    {myArtefactsByUnit.map((a) => (
                      <option key={a.unit_id} value={a.unit_id}>
                        #{a.unit_id} ({a.count})
                      </option>
                    ))}
                  </select>
                </label>
                <label>
                  {t('stock', 'createQtyLabel')}:{' '}
                  <input
                    type="number"
                    min={1}
                    value={cQty}
                    onChange={(e) => setCQty(e.target.value)}
                    style={{ width: 80 }}
                    required
                  />
                </label>
                <label>
                  {t('stock', 'createPriceLabel')}:{' '}
                  <input
                    type="number"
                    min={1}
                    value={cPrice}
                    onChange={(e) => setCPrice(e.target.value)}
                    style={{ width: 100 }}
                    required
                  />
                </label>
                <label>
                  {t('stock', 'createExpiryLabel')}:{' '}
                  <input
                    type="number"
                    min={1}
                    value={cExp}
                    onChange={(e) => setCExp(e.target.value)}
                    style={{ width: 60 }}
                    required
                  />
                </label>
                <button
                  type="submit"
                  className="button"
                  disabled={createMut.isPending || cArt === ''}
                >
                  {t('stock', 'createSubmit')}
                </button>
              </form>
            </td>
          </tr>
        )}
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
              {lot.seller_user_id === myUserId ? (
                <input
                  type="button"
                  className="button"
                  value={t('stock', 'cancelBtn')}
                  disabled={cancelMut.isPending}
                  onClick={() => cancelMut.mutate(lot.id)}
                />
              ) : (
                <input
                  type="button"
                  className="button"
                  value={t('stock', 'buyBtn')}
                  disabled={buyMut.isPending}
                  onClick={() => buyMut.mutate(lot.id)}
                />
              )}
            </td>
          </tr>
        ))}
        {(actionMsg || actionErr) && (
          <tr>
            <td colSpan={6} className="center">
              {actionMsg && <span className="true">{actionMsg}</span>}
              {actionErr && <span className="false">{actionErr}</span>}
            </td>
          </tr>
        )}
      </tbody>
    </table>
  );
}
