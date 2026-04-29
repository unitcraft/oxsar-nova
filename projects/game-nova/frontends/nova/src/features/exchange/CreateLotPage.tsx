// Создание лота биржи (план 76 Ф.4). Форма + live-валидация +
// POST /api/exchange/lots с Idempotency-Key (R9).

import { useMemo, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api, genIdempotencyKey, type ApiError } from '@/api/client';
import { exchangeApi, type ExchangeLot } from '@/api/exchange';
import { ARTEFACTS, nameOf } from '@/api/catalog';
import type { Artefact } from '@/api/types';
import { useToast } from '@/ui/Toast';
import { useTranslation } from '@/i18n/i18n';
import {
  EXPIRES_OPTIONS,
  MAX_QUANTITY_PER_LOT,
  errorMessageKey,
  validateCreateLot,
} from './filters';

interface Props {
  onBack: () => void;
  onCreated: (lot: ExchangeLot) => void;
}

export function CreateLotPage({ onBack, onCreated }: Props) {
  const { t } = useTranslation('exchange');
  const { t: ti } = useTranslation('info');
  const qc = useQueryClient();
  const toast = useToast();

  const [artifactUnitId, setArtifactUnitId] = useState<number | null>(null);
  const [quantityStr, setQuantityStr] = useState('1');
  const [priceStr, setPriceStr] = useState('100');
  const [expiresInHours, setExpiresInHours] = useState<number>(24);
  // Idempotency-Key стабилен на всю сессию формы (не пересчитывается при
  // ретраях — это и нужно для дедупа на бэке). Регенерируем после успеха
  // (на случай повторной отправки той же формы).
  const [idemKey, setIdemKey] = useState<string>(() => genIdempotencyKey());

  const list = useQuery({
    queryKey: ['artefacts'],
    queryFn: () => api.get<{ artefacts: Artefact[] | null }>('/api/artefacts'),
    staleTime: 5_000,
  });

  // Доступные артефакты — только state='held' (можно листинговать).
  // Группируем по unit_id, чтобы определить максимум qty.
  const heldByUnit = useMemo(() => {
    const out = new Map<number, number>();
    for (const a of list.data?.artefacts ?? []) {
      if (a.state !== 'held') continue;
      out.set(a.unit_id, (out.get(a.unit_id) ?? 0) + 1);
    }
    return out;
  }, [list.data]);

  const available = artifactUnitId !== null ? (heldByUnit.get(artifactUnitId) ?? 0) : 0;
  const quantity = parseInt(quantityStr) || 0;
  const priceOxsarit = parseInt(priceStr) || 0;

  const validationKey = validateCreateLot({
    artifactUnitId,
    quantity,
    available,
    priceOxsarit,
    expiresInHours,
  });

  const create = useMutation({
    mutationFn: () => exchangeApi.createLot(
      {
        artifact_unit_id: artifactUnitId!,
        quantity,
        price_oxsarit: priceOxsarit,
        expires_in_hours: expiresInHours,
      },
      { idempotencyKey: idemKey },
    ),
    onSuccess: (resp) => {
      void qc.invalidateQueries({ queryKey: ['exchange'] });
      void qc.invalidateQueries({ queryKey: ['artefacts'] });
      toast.show('success', t('toastListedTitle'), t('lotCreated', { quantity: String(quantity) }));
      setIdemKey(genIdempotencyKey());
      onCreated(resp.lot);
    },
    onError: (err) => {
      const e = err as ApiError;
      const code = errorMessageKey(e.code);
      const message = code === 'generic' ? (e.message || t('errors', 'generic')) : t('errors', code);
      toast.show('danger', t('toastCreateErrTitle'), message);
    },
  });

  const onSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (validationKey !== null) return;
    create.mutate();
  };

  const heldOptions = ARTEFACTS.filter((a) => (heldByUnit.get(a.id) ?? 0) > 0);

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
        <button type="button" className="btn-ghost btn-sm" onClick={onBack}>← {t('backToList')}</button>
        <h2 style={{ margin: 0, fontSize: 18, fontFamily: 'var(--ox-font)', fontWeight: 700 }}>
          {t('createTitle')}
        </h2>
      </div>

      {heldOptions.length === 0 ? (
        <div className="ox-panel" style={{ padding: 24, textAlign: 'center', color: 'var(--ox-fg-dim)' }}>
          {t('noHeldArtefacts')}
        </div>
      ) : (
        <form onSubmit={onSubmit} className="ox-panel" style={{ padding: 16, display: 'flex', flexDirection: 'column', gap: 14 }}>
          <Field label={t('formArtefact')}>
            <select
              required
              value={artifactUnitId ?? ''}
              onChange={(e) => {
                setArtifactUnitId(e.target.value === '' ? null : Number(e.target.value));
                setQuantityStr('1');
              }}
            >
              <option value="">{t('formPickArtefact')}</option>
              {heldOptions.map((a) => (
                <option key={a.id} value={a.id}>
                  {nameOf(a.id, ti)} ({heldByUnit.get(a.id) ?? 0})
                </option>
              ))}
            </select>
          </Field>

          <Field
            label={t('formQuantity')}
            {...(artifactUnitId !== null
              ? { hint: t('formQuantityHint', { available: String(available), max: String(MAX_QUANTITY_PER_LOT) }) }
              : {})}
          >
            <input
              type="number"
              inputMode="numeric"
              min={1}
              max={Math.min(available, MAX_QUANTITY_PER_LOT) || 1}
              value={quantityStr}
              onChange={(e) => setQuantityStr(e.target.value)}
              disabled={artifactUnitId === null}
              style={{ width: 140 }}
            />
          </Field>

          <Field label={t('formPrice')} hint={t('formPriceHint')}>
            <input
              type="number"
              inputMode="numeric"
              min={1}
              value={priceStr}
              onChange={(e) => setPriceStr(e.target.value)}
              style={{ width: 180 }}
            />
            <span style={{ marginLeft: 8, color: 'var(--ox-fg-dim)', fontSize: 13 }}>
              {t('oxsaritShort')}
            </span>
          </Field>

          <Field label={t('formExpiresIn')}>
            <select
              value={expiresInHours}
              onChange={(e) => setExpiresInHours(Number(e.target.value))}
            >
              {EXPIRES_OPTIONS.map((o) => (
                <option key={o.hours} value={o.hours}>{t(o.tKey as never)}</option>
              ))}
            </select>
          </Field>

          {validationKey !== null && (
            <div role="alert" style={{ color: 'var(--ox-danger)', fontSize: 14 }}>
              {t('validation', validationKey)}
            </div>
          )}

          <div style={{ display: 'flex', gap: 12 }}>
            <button
              type="submit"
              className="btn btn-success btn-sm"
              disabled={validationKey !== null || create.isPending}
            >
              {create.isPending ? t('creating') : t('submitCreate')}
            </button>
            <button type="button" className="btn-ghost btn-sm" onClick={onBack}>
              {t('cancelForm')}
            </button>
          </div>
        </form>
      )}
    </div>
  );
}

function Field({ label, hint, children }: { label: string; hint?: string; children: React.ReactNode }) {
  return (
    <label style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
      <span style={{ fontSize: 13, color: 'var(--ox-fg-dim)' }}>{label}</span>
      <div style={{ display: 'flex', alignItems: 'center' }}>{children}</div>
      {hint && <span style={{ fontSize: 12, color: 'var(--ox-fg-muted)' }}>{hint}</span>}
    </label>
  );
}
