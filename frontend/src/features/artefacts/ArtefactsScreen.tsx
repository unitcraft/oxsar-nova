import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';
import { ARTEFACTS, nameOf, imageOf } from '@/api/catalog';
import type { Artefact } from '@/api/types';
import { useToast } from '@/ui/Toast';
import { Countdown } from '@/ui/Countdown';
import { ScreenSkeleton } from '@/ui/Skeleton';

const STATE_LABEL: Record<string, string> = {
  held: '📦 В инвентаре',
  active: '✅ Активен',
  delayed: '⏳ Активируется',
  listed: '🏷 На продаже',
  expired: '💀 Истёк',
  consumed: '⚡ Использован',
};

export function ArtefactsScreen() {
  const qc = useQueryClient();
  const toast = useToast();
  const [sellingID, setSellingID] = useState<string | null>(null);
  const [priceInput, setPriceInput] = useState('100');

  const list = useQuery({
    queryKey: ['artefacts'],
    queryFn: () => api.get<{ artefacts: Artefact[] | null }>('/api/artefacts'),
    refetchInterval: 5000,
  });

  const activate = useMutation({
    mutationFn: (id: string) => api.post<Artefact>(`/api/artefacts/${id}/activate`),
    onMutate: async (id) => {
      await qc.cancelQueries({ queryKey: ['artefacts'] });
      const prev = qc.getQueryData<{ artefacts: Artefact[] | null }>(['artefacts']);
      qc.setQueryData<{ artefacts: Artefact[] | null }>(['artefacts'], (old) => ({
        artefacts: old?.artefacts?.map((a) => a.id === id ? { ...a, state: 'delayed' as const } : a) ?? null,
      }));
      return { prev };
    },
    onSuccess: (a) => {
      void qc.invalidateQueries({ queryKey: ['artefacts'] });
      void qc.invalidateQueries({ queryKey: ['planets'] });
      toast.show('success', 'Артефакт', `${nameOf(a.unit_id)} активирован`);
    },
    onError: (err, _id, ctx) => {
      if (ctx?.prev) qc.setQueryData(['artefacts'], ctx.prev);
      toast.show('danger', 'Ошибка активации', err instanceof Error ? err.message : '');
    },
  });
  const deactivate = useMutation({
    mutationFn: (id: string) => api.post<void>(`/api/artefacts/${id}/deactivate`),
    onMutate: async (id) => {
      await qc.cancelQueries({ queryKey: ['artefacts'] });
      const prev = qc.getQueryData<{ artefacts: Artefact[] | null }>(['artefacts']);
      qc.setQueryData<{ artefacts: Artefact[] | null }>(['artefacts'], (old) => ({
        artefacts: old?.artefacts?.map((a) => a.id === id ? { ...a, state: 'held' as const } : a) ?? null,
      }));
      return { prev };
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['artefacts'] });
      void qc.invalidateQueries({ queryKey: ['planets'] });
      toast.show('info', 'Артефакт деактивирован');
    },
    onError: (err, _id, ctx) => {
      if (ctx?.prev) qc.setQueryData(['artefacts'], ctx.prev);
      toast.show('danger', 'Ошибка', err instanceof Error ? err.message : '');
    },
  });
  const sell = useMutation({
    mutationFn: (p: { id: string; price: number }) =>
      api.post<void>(`/api/artefacts/${p.id}/sell`, { price: p.price }),
    onMutate: async ({ id }) => {
      await qc.cancelQueries({ queryKey: ['artefacts'] });
      const prev = qc.getQueryData<{ artefacts: Artefact[] | null }>(['artefacts']);
      qc.setQueryData<{ artefacts: Artefact[] | null }>(['artefacts'], (old) => ({
        artefacts: old?.artefacts?.map((a) => a.id === id ? { ...a, state: 'listed' as const } : a) ?? null,
      }));
      return { prev };
    },
    onSuccess: () => {
      setSellingID(null);
      void qc.invalidateQueries({ queryKey: ['artefacts'] });
      void qc.invalidateQueries({ queryKey: ['artefact-market'] });
      toast.show('success', 'Артефакт выставлен на продажу');
    },
    onError: (err, _p, ctx) => {
      if (ctx?.prev) qc.setQueryData(['artefacts'], ctx.prev);
      toast.show('danger', 'Ошибка продажи', err instanceof Error ? err.message : '');
    },
  });

  const items = list.data?.artefacts ?? [];
  const active = items.filter((a) => a.state === 'active');
  const held = items.filter((a) => a.state === 'held' || a.state === 'delayed');
  const other = items.filter((a) => !['active', 'held', 'delayed'].includes(a.state));

  function openSellForm(id: string) { setSellingID(id); setPriceInput('100'); }
  function confirmSell() {
    if (!sellingID) return;
    const price = Number(priceInput);
    if (price > 0) sell.mutate({ id: sellingID, price });
  }

  if (list.isLoading) {
    return <ScreenSkeleton />;
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <h2 style={{ margin: 0, fontSize: 18, fontFamily: 'var(--ox-font)', fontWeight: 700 }}>
        ✨ Артефакты
      </h2>

      {items.length === 0 && (
        <div className="ox-panel" style={{ padding: 24, textAlign: 'center', color: 'var(--ox-fg-dim)' }}>
          Инвентарь пуст. Артефакты появляются как награда за бой/экспедицию или покупаются в Рынке артефактов.
        </div>
      )}

      {active.length > 0 && (
        <ArtefactGroup title="Активные" items={active} sellingID={sellingID} priceInput={priceInput}
          setPriceInput={setPriceInput} openSellForm={openSellForm} confirmSell={confirmSell}
          onActivate={(id) => activate.mutate(id)} onDeactivate={(id) => deactivate.mutate(id)}
          pending={activate.isPending || deactivate.isPending || sell.isPending}
        />
      )}
      {held.length > 0 && (
        <ArtefactGroup title="В инвентаре" items={held} sellingID={sellingID} priceInput={priceInput}
          setPriceInput={setPriceInput} openSellForm={openSellForm} confirmSell={confirmSell}
          onActivate={(id) => activate.mutate(id)} onDeactivate={(id) => deactivate.mutate(id)}
          pending={activate.isPending || deactivate.isPending || sell.isPending}
        />
      )}
      {other.length > 0 && (
        <ArtefactGroup title="Прочие" items={other} sellingID={sellingID} priceInput={priceInput}
          setPriceInput={setPriceInput} openSellForm={openSellForm} confirmSell={confirmSell}
          onActivate={(id) => activate.mutate(id)} onDeactivate={(id) => deactivate.mutate(id)}
          pending={activate.isPending || deactivate.isPending || sell.isPending}
        />
      )}
    </div>
  );
}

function ArtefactGroup({
  title, items, sellingID, priceInput, setPriceInput,
  openSellForm, confirmSell, onActivate, onDeactivate, pending,
}: {
  title: string;
  items: Artefact[];
  sellingID: string | null;
  priceInput: string;
  setPriceInput: (v: string) => void;
  openSellForm: (id: string) => void;
  confirmSell: () => void;
  onActivate: (id: string) => void;
  onDeactivate: (id: string) => void;
  pending: boolean;
}) {
  return (
    <div>
      <div style={{ fontSize: 13, fontWeight: 700, letterSpacing: '0.08em', textTransform: 'uppercase', color: 'var(--ox-fg-muted)', marginBottom: 8 }}>
        {title}
      </div>
      <div className="ox-cards-grid">
        {items.map((a) => (
          <div
            key={a.id}
            className="ox-unit-card"
            style={a.state === 'active' ? { borderColor: 'var(--ox-success)', boxShadow: '0 0 0 1px var(--ox-success)' } : undefined}
          >
            <div className="ox-unit-card-img">
              {(() => {
                const meta = ARTEFACTS.find((x) => x.id === a.unit_id);
                return meta
                  ? <img src={imageOf(meta.key)} alt={meta.name} width={64} height={64} style={{ imageRendering: 'pixelated' }} />
                  : <span style={{ fontSize: 36 }}>✨</span>;
              })()}
            </div>
            <div className="ox-unit-card-body">
              <div className="ox-unit-card-name">{nameOf(a.unit_id)}</div>
              <div style={{ fontSize: 14, color: 'var(--ox-fg-dim)', marginBottom: 2 }}>
                {STATE_LABEL[a.state] ?? a.state}
              </div>
              {(() => {
                const meta = ARTEFACTS.find((x) => x.id === a.unit_id);
                return meta && (
                  <div style={{ fontSize: 13, color: 'var(--ox-success)', fontStyle: 'italic', marginBottom: 2 }}>
                    {meta.benefit}
                  </div>
                );
              })()}
              {a.expire_at && a.state === 'active' && (
                <div style={{ fontSize: 14, color: 'var(--ox-fg-muted)' }}>
                  Истекает: <Countdown finishAt={a.expire_at} />
                </div>
              )}
            </div>
            <div className="ox-unit-card-footer" style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
              {a.state === 'held' && sellingID !== a.id && (
                <div style={{ display: 'flex', gap: 6 }}>
                  <button type="button" className="btn btn-sm" style={{ flex: 1 }} disabled={pending} onClick={() => onActivate(a.id)}>
                    Активировать
                  </button>
                  <button type="button" className="btn-ghost btn-sm" disabled={pending} onClick={() => openSellForm(a.id)}>
                    Продать
                  </button>
                </div>
              )}
              {a.state === 'held' && sellingID === a.id && (
                <div style={{ display: 'flex', gap: 4, alignItems: 'center' }}>
                  <input
                    type="number" min={1} value={priceInput}
                    onChange={(e) => setPriceInput(e.target.value)}
                    onKeyDown={(e) => { if (e.key === 'Enter') confirmSell(); if (e.key === 'Escape') openSellForm(''); }}
                    style={{ width: 70, flexShrink: 0 }}
                    autoFocus
                  />
                  <button type="button" className="btn btn-sm" disabled={pending || Number(priceInput) <= 0} onClick={confirmSell}>OK</button>
                  <button type="button" className="btn-ghost btn-sm" onClick={() => openSellForm('')}>✕</button>
                </div>
              )}
              {a.state === 'active' && (
                <button type="button" className="btn-ghost btn-sm" style={{ width: '100%' }} disabled={pending} onClick={() => onDeactivate(a.id)}>
                  Деактивировать
                </button>
              )}
              {a.state === 'expired' && (
                <div style={{ fontSize: 14, color: 'var(--ox-fg-muted)', textAlign: 'center' }}>💀 Истёк</div>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
