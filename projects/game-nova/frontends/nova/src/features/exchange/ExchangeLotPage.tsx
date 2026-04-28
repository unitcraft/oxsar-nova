// Детали лота биржи + действия покупки/отзыва (план 76 Ф.3).

import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api, genIdempotencyKey, type ApiError } from '@/api/client';
import { exchangeApi } from '@/api/exchange';
import { nameOf } from '@/api/catalog';
import { Confirm } from '@/ui/Confirm';
import { useToast } from '@/ui/Toast';
import { ScreenSkeleton } from '@/ui/Skeleton';
import { useTranslation } from '@/i18n/i18n';
import { errorMessageKey } from './filters';

interface Props {
  lotId: string;
  onBack: () => void;
}

export function ExchangeLotPage({ lotId, onBack }: Props) {
  const { t } = useTranslation('exchange');
  const { t: ti } = useTranslation('info');
  const { t: te } = useTranslation('exchange');
  const qc = useQueryClient();
  const toast = useToast();
  const [confirmKind, setConfirmKind] = useState<null | 'buy' | 'cancel'>(null);

  const me = useQuery({
    queryKey: ['me'],
    queryFn: () => api.get<{ user_id: string; credit: number }>('/api/me'),
    staleTime: 60_000,
  });

  const lotQ = useQuery({
    queryKey: ['exchange', 'lot', lotId],
    queryFn: () => exchangeApi.getLot(lotId),
    refetchInterval: 15_000,
  });

  const buy = useMutation({
    mutationFn: () => exchangeApi.buyLot(lotId, { idempotencyKey: genIdempotencyKey() }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['exchange'] });
      void qc.invalidateQueries({ queryKey: ['me'] });
      void qc.invalidateQueries({ queryKey: ['artefacts'] });
      toast.show('success', t('toastBoughtTitle'), t('lotBought'));
      onBack();
    },
    onError: (err) => {
      const e = err as ApiError;
      // 409 lot_not_active → инвалидируем и возвращаемся в список,
      // чтобы пользователь увидел актуальное состояние.
      if (e.status === 409) {
        void qc.invalidateQueries({ queryKey: ['exchange'] });
        toast.show('warning', t('toastBuyErrTitle'), t(`errors.${errorMessageKey(e.code)}` as never));
        onBack();
        return;
      }
      const code = errorMessageKey(e.code);
      const message = code === 'generic' ? (e.message || t('errors.generic')) : te(`errors.${code}` as never);
      toast.show('danger', t('toastBuyErrTitle'), message);
    },
  });

  const cancel = useMutation({
    mutationFn: () => exchangeApi.cancelLot(lotId),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['exchange'] });
      void qc.invalidateQueries({ queryKey: ['artefacts'] });
      toast.show('info', t('toastCancelledTitle'), t('lotCancelled'));
      onBack();
    },
    onError: (err) => {
      const e = err as ApiError;
      const code = errorMessageKey(e.code);
      const message = code === 'generic' ? (e.message || t('errors.generic')) : te(`errors.${code}` as never);
      toast.show('danger', t('toastCancelErrTitle'), message);
    },
  });

  if (lotQ.isLoading) return <ScreenSkeleton />;
  if (lotQ.isError || !lotQ.data) {
    return (
      <div className="ox-panel" style={{ padding: 24, textAlign: 'center' }}>
        <div style={{ marginBottom: 12, color: 'var(--ox-fg-dim)' }}>{t('loadError')}</div>
        <button type="button" className="btn-ghost btn-sm" onClick={onBack}>{t('backToList')}</button>
      </div>
    );
  }

  const { lot, items } = lotQ.data;
  const isMine = me.data?.user_id === lot.seller_user_id;
  const myCredit = me.data?.credit ?? 0;
  const canAfford = myCredit >= lot.price_oxsarit;
  const isActive = lot.status === 'active';

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
        <button type="button" className="btn-ghost btn-sm" onClick={onBack}>← {t('backToList')}</button>
        <h2 style={{ margin: 0, fontSize: 18, fontFamily: 'var(--ox-font)', fontWeight: 700 }}>
          {t('detailTitle')}
        </h2>
      </div>

      <div className="ox-panel" style={{ padding: 16, display: 'flex', flexDirection: 'column', gap: 12 }}>
        <Row label={t('detailArtefact')} value={nameOf(lot.artifact_unit_id, ti)} />
        <Row label={t('detailQuantity')} value={String(lot.quantity)} />
        <Row
          label={t('detailPrice')}
          value={`${lot.price_oxsarit} ${t('oxsaritShort')}`}
          accent
        />
        <Row
          label={t('detailUnitPrice')}
          value={`${lot.unit_price_oxsarit ?? Math.floor(lot.price_oxsarit / Math.max(lot.quantity, 1))} ${t('oxsaritShort')}`}
        />
        <Row label={t('detailSeller')} value={lot.seller_username ?? '—'} />
        <Row label={t('detailStatus')} value={t(`status${capitalize(lot.status)}` as never)} />
        <Row label={t('detailCreatedAt')} value={formatDateTime(lot.created_at)} />
        <Row label={t('detailExpiresAt')} value={formatDateTime(lot.expires_at)} />
        {lot.sold_at && <Row label={t('detailSoldAt')} value={formatDateTime(lot.sold_at)} />}
        <Row label={t('detailItemCount')} value={String(items.length)} />
      </div>

      {isActive && (
        <div style={{ display: 'flex', gap: 12, alignItems: 'center' }}>
          {isMine ? (
            <button
              type="button"
              className="btn-danger btn-sm"
              disabled={cancel.isPending}
              onClick={() => setConfirmKind('cancel')}
            >
              {t('cancelLotBtn')}
            </button>
          ) : (
            <>
              <button
                type="button"
                className={`btn btn-sm${!canAfford ? ' btn-ghost' : ' btn-success'}`}
                disabled={!canAfford || buy.isPending}
                title={!canAfford ? t('errors.insufficientOxsarits') : undefined}
                onClick={() => setConfirmKind('buy')}
              >
                {canAfford ? t('buyLotBtn') : t('buyLotCantAfford')}
              </button>
              <span style={{ fontSize: 13, color: 'var(--ox-fg-dim)' }}>
                {t('detailYourBalance')}: {myCredit} {t('oxsaritShort')}
              </span>
            </>
          )}
        </div>
      )}

      {confirmKind === 'buy' && (
        <Confirm
          title={t('confirmBuyTitle')}
          message={t('confirmBuyMessage', {
            qty: String(lot.quantity),
            name: nameOf(lot.artifact_unit_id, ti),
            price: String(lot.price_oxsarit),
            unit: t('oxsaritShort'),
          })}
          confirmLabel={t('buyLotBtn')}
          onConfirm={() => { setConfirmKind(null); buy.mutate(); }}
          onCancel={() => setConfirmKind(null)}
        />
      )}
      {confirmKind === 'cancel' && (
        <Confirm
          danger
          title={t('confirmCancelTitle')}
          message={t('confirmCancelMessage')}
          confirmLabel={t('cancelLotBtn')}
          onConfirm={() => { setConfirmKind(null); cancel.mutate(); }}
          onCancel={() => setConfirmKind(null)}
        />
      )}
    </div>
  );
}

function Row({ label, value, accent }: { label: string; value: string; accent?: boolean }) {
  return (
    <div style={{ display: 'flex', alignItems: 'baseline', gap: 12 }}>
      <span style={{ minWidth: 180, color: 'var(--ox-fg-dim)', fontSize: 14 }}>{label}</span>
      <span style={{
        fontFamily: 'var(--ox-mono)',
        fontWeight: accent ? 700 : 500,
        color: accent ? 'var(--ox-accent)' : 'var(--ox-fg)',
      }}>
        {value}
      </span>
    </div>
  );
}

function capitalize(s: string): string {
  return s ? s[0]!.toUpperCase() + s.slice(1) : '';
}

function formatDateTime(iso: string): string {
  const d = new Date(iso);
  if (!Number.isFinite(d.getTime())) return '—';
  return d.toLocaleString('ru-RU', {
    year: 'numeric', month: '2-digit', day: '2-digit',
    hour: '2-digit', minute: '2-digit',
  });
}
