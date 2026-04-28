// Список лотов биржи (план 76 Ф.2). Cursor-pagination через
// useInfiniteQuery, debounced filters, карточный layout в nova-стиле.

import { useEffect, useMemo, useRef, useState } from 'react';
import { useInfiniteQuery, useQuery } from '@tanstack/react-query';
import { api } from '@/api/client';
import { exchangeApi, type ExchangeLot, type ListLotsResponse } from '@/api/exchange';
import { ARTEFACTS, nameOf } from '@/api/catalog';
import { ScreenSkeleton } from '@/ui/Skeleton';
import { useTranslation } from '@/i18n/i18n';
import {
  EMPTY_FILTERS,
  buildQueryParams,
  validatePriceRange,
  type ExchangeFilters,
  type StatusFilter,
} from './filters';

interface Props {
  onOpenLot: (id: string) => void;
  onCreate: () => void;
}

const PAGE_SIZE = 50;
const DEBOUNCE_MS = 300;

export function ExchangeListPage({ onOpenLot, onCreate }: Props) {
  const { t } = useTranslation('exchange');
  const { t: ti } = useTranslation('info');
  const { t: tg } = useTranslation('global');

  const [draft, setDraft] = useState<ExchangeFilters>(EMPTY_FILTERS);
  const [active, setActive] = useState<ExchangeFilters>(EMPTY_FILTERS);

  // Debounce: коммитим draft в active через DEBOUNCE_MS после последней
  // правки. TanStack Query сам перезапускает запрос при смене ключа.
  useEffect(() => {
    const id = window.setTimeout(() => setActive(draft), DEBOUNCE_MS);
    return () => window.clearTimeout(id);
  }, [draft]);

  const priceErr = validatePriceRange(active);

  const me = useQuery({
    queryKey: ['me'],
    queryFn: () => api.get<{ user_id: string; credit: number }>('/api/me'),
    staleTime: 60_000,
  });

  const lots = useInfiniteQuery({
    queryKey: ['exchange', 'lots', serialize(active)],
    queryFn: ({ pageParam }) =>
      exchangeApi.listLots(buildQueryParams(active, pageParam, PAGE_SIZE).toString()),
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (last: ListLotsResponse) => last.next_cursor ?? undefined,
    staleTime: 10_000,
    enabled: priceErr === null,
  });

  const allLots: ExchangeLot[] = useMemo(
    () => (lots.data?.pages ?? []).flatMap((p) => p.lots),
    [lots.data],
  );

  // Бесконечный скролл: IntersectionObserver на sentinel в конце списка.
  const sentinelRef = useRef<HTMLDivElement>(null);
  useEffect(() => {
    const el = sentinelRef.current;
    if (!el) return;
    const obs = new IntersectionObserver(
      (entries) => {
        const e = entries[0];
        if (e?.isIntersecting && lots.hasNextPage && !lots.isFetchingNextPage) {
          void lots.fetchNextPage();
        }
      },
      { rootMargin: '120px' },
    );
    obs.observe(el);
    return () => obs.disconnect();
  }, [lots]);

  const myUserId = me.data?.user_id;
  const myCredit = me.data?.credit ?? 0;

  if (lots.isLoading) return <ScreenSkeleton />;

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: 8 }}>
        <h2 style={{ margin: 0, fontSize: 18, fontFamily: 'var(--ox-font)', fontWeight: 700 }}>
          {t('listTitle')}
        </h2>
        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <span style={{ fontSize: 14, color: 'var(--ox-fg-dim)' }}>{t('balanceLabel')}</span>
          <span style={{ fontFamily: 'var(--ox-mono)', fontWeight: 700, color: 'var(--ox-accent)', fontSize: 15 }}>
            {myCredit}&nbsp;{t('oxsaritShort')}
          </span>
          <button type="button" className="btn btn-sm" onClick={onCreate}>
            {t('createBtn')}
          </button>
        </div>
      </div>

      <FilterBar
        value={draft}
        onChange={setDraft}
        artefactName={(id) => nameOf(id, ti)}
        labelArtefact={t('filterArtefact')}
        labelStatus={t('filterStatus')}
        labelMin={t('filterMinPrice')}
        labelMax={t('filterMaxPrice')}
        labelSeller={t('filterSellerId')}
        labelAll={tg('all')}
        statusLabels={{
          active: t('statusActive'),
          sold: t('statusSold'),
          cancelled: t('statusCancelled'),
          expired: t('statusExpired'),
          all: tg('all'),
        }}
      />

      {priceErr !== null && (
        <div role="alert" style={{ color: 'var(--ox-danger)', fontSize: 14 }}>
          {t(priceErr === 'priceRangeInvalid' ? 'priceRangeInvalid' : 'priceNegative')}
        </div>
      )}

      {!lots.isLoading && allLots.length === 0 && (
        <div className="ox-panel" style={{ padding: 24, textAlign: 'center', color: 'var(--ox-fg-dim)' }}>
          <div style={{ marginBottom: 12 }}>{t('emptyList')}</div>
          <button type="button" className="btn btn-sm" onClick={onCreate}>
            {t('createFirstBtn')}
          </button>
        </div>
      )}

      {allLots.length > 0 && (
        <div className="ox-panel" style={{ overflow: 'hidden' }}>
          <div className="ox-table-responsive">
            <table className="ox-table" style={{ margin: 0 }}>
              <thead>
                <tr>
                  <th>{t('colArtefact')}</th>
                  <th>{t('colQty')}</th>
                  <th>{t('colPrice')}</th>
                  <th>{t('colUnitPrice')}</th>
                  <th>{t('colSeller')}</th>
                  <th>{t('colExpiresAt')}</th>
                  <th />
                </tr>
              </thead>
              <tbody>
                {allLots.map((l) => (
                  <LotRow
                    key={l.id}
                    lot={l}
                    myUserId={myUserId}
                    artName={nameOf(l.artifact_unit_id, ti)}
                    onOpen={() => onOpenLot(l.id)}
                    labels={{
                      open: t('openBtn'),
                      mine: t('mineBadge'),
                      oxsaritShort: t('oxsaritShort'),
                    }}
                  />
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      <div ref={sentinelRef} style={{ height: 1 }} />
      {lots.isFetchingNextPage && (
        <div style={{ textAlign: 'center', color: 'var(--ox-fg-dim)', padding: 8 }}>
          {t('loadingMore')}
        </div>
      )}
    </div>
  );
}

interface LotRowProps {
  lot: ExchangeLot;
  myUserId: string | undefined;
  artName: string;
  onOpen: () => void;
  labels: { open: string; mine: string; oxsaritShort: string };
}

function LotRow({ lot, myUserId, artName, onOpen, labels }: LotRowProps) {
  const isMine = lot.seller_user_id === myUserId;
  const expires = formatRelative(lot.expires_at);
  const unit = lot.unit_price_oxsarit ?? Math.floor(lot.price_oxsarit / Math.max(lot.quantity, 1));
  return (
    <tr>
      <td>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <span>{artName}</span>
          {isMine && (
            <span className="badge" style={{ fontSize: 11 }}>{labels.mine}</span>
          )}
        </div>
      </td>
      <td className="num" style={{ fontFamily: 'var(--ox-mono)' }}>{lot.quantity}</td>
      <td className="num" style={{ fontFamily: 'var(--ox-mono)', color: 'var(--ox-accent)' }}>
        {lot.price_oxsarit}&nbsp;{labels.oxsaritShort}
      </td>
      <td className="num" style={{ fontFamily: 'var(--ox-mono)', color: 'var(--ox-fg-dim)' }}>
        {unit}&nbsp;{labels.oxsaritShort}
      </td>
      <td>{lot.seller_username ?? '—'}</td>
      <td style={{ fontSize: 13, color: 'var(--ox-fg-muted)', fontFamily: 'var(--ox-mono)' }}>{expires}</td>
      <td>
        <button type="button" className="btn-ghost btn-sm" onClick={onOpen}>{labels.open}</button>
      </td>
    </tr>
  );
}

interface FilterBarProps {
  value: ExchangeFilters;
  onChange: (next: ExchangeFilters) => void;
  artefactName: (id: number) => string;
  labelArtefact: string;
  labelStatus: string;
  labelMin: string;
  labelMax: string;
  labelSeller: string;
  labelAll: string;
  statusLabels: Record<StatusFilter, string>;
}

function FilterBar({
  value, onChange, artefactName,
  labelArtefact, labelStatus, labelMin, labelMax, labelSeller, labelAll,
  statusLabels,
}: FilterBarProps) {
  return (
    <div className="ox-panel" style={{ padding: 12, display: 'flex', flexWrap: 'wrap', gap: 12, alignItems: 'flex-end' }}>
      <Field label={labelArtefact}>
        <select
          value={value.artifactUnitId ?? ''}
          onChange={(e) => onChange({
            ...value,
            artifactUnitId: e.target.value === '' ? null : Number(e.target.value),
          })}
        >
          <option value="">{labelAll}</option>
          {ARTEFACTS.map((a) => (
            <option key={a.id} value={a.id}>{artefactName(a.id)}</option>
          ))}
        </select>
      </Field>
      <Field label={labelMin}>
        <input
          type="number"
          inputMode="numeric"
          min={0}
          value={value.minPrice}
          onChange={(e) => onChange({ ...value, minPrice: e.target.value })}
          style={{ width: 110 }}
        />
      </Field>
      <Field label={labelMax}>
        <input
          type="number"
          inputMode="numeric"
          min={0}
          value={value.maxPrice}
          onChange={(e) => onChange({ ...value, maxPrice: e.target.value })}
          style={{ width: 110 }}
        />
      </Field>
      <Field label={labelStatus}>
        <select
          value={value.status}
          onChange={(e) => onChange({ ...value, status: e.target.value as StatusFilter })}
        >
          <option value="active">{statusLabels.active}</option>
          <option value="sold">{statusLabels.sold}</option>
          <option value="cancelled">{statusLabels.cancelled}</option>
          <option value="expired">{statusLabels.expired}</option>
          <option value="all">{statusLabels.all}</option>
        </select>
      </Field>
      <Field label={labelSeller}>
        <input
          type="text"
          value={value.sellerId}
          onChange={(e) => onChange({ ...value, sellerId: e.target.value })}
          placeholder="UUID"
          style={{ width: 200 }}
        />
      </Field>
    </div>
  );
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <label style={{ display: 'flex', flexDirection: 'column', gap: 4, fontSize: 13 }}>
      <span style={{ color: 'var(--ox-fg-dim)' }}>{label}</span>
      {children}
    </label>
  );
}

function serialize(f: ExchangeFilters): string {
  // Стабильная сериализация для query-key. Дефолтное состояние всегда
  // даёт одинаковый ключ независимо от порядка манипуляций.
  return JSON.stringify({
    a: f.artifactUnitId,
    n: f.minPrice.trim(),
    x: f.maxPrice.trim(),
    s: f.sellerId.trim(),
    t: f.status,
  });
}

function formatRelative(iso: string): string {
  const ts = new Date(iso).getTime();
  if (!Number.isFinite(ts)) return '—';
  const deltaSec = Math.floor((ts - Date.now()) / 1000);
  if (deltaSec <= 0) return '—';
  const days = Math.floor(deltaSec / 86_400);
  if (days >= 1) return `${days}d`;
  const hours = Math.floor(deltaSec / 3600);
  if (hours >= 1) return `${hours}h`;
  const minutes = Math.floor(deltaSec / 60);
  if (minutes >= 1) return `${minutes}m`;
  return `${deltaSec}s`;
}

