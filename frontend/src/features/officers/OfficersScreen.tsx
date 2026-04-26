import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';
import { useToast } from '@/ui/Toast';
import { Countdown } from '@/ui/Countdown';
import { useTranslation } from '@/i18n/i18n';

interface Entry {
  key: string;
  title: string;
  description: string;
  duration_days: number;
  cost_credit: number;
  effect?: Record<string, number> | null;
  activated_at?: string | null;
  expires_at?: string | null;
}

function fmtEffect(effect: Record<string, number> | null | undefined, effectLabel: (k: string) => string): string | null {
  if (!effect) return null;
  return Object.entries(effect)
    .filter(([, v]) => v !== 1)
    .map(([k, v]) => {
      const label = effectLabel(k);
      const pct = Math.round((v - 1) * 100);
      return `${label} ${pct > 0 ? '+' : ''}${pct}%`;
    })
    .join(', ') || null;
}

export function OfficersScreen() {
  const { t } = useTranslation('officersUi');
  const qc = useQueryClient();
  const toast = useToast();
  const [autoRenewKeys, setAutoRenewKeys] = useState<Set<string>>(new Set());

  const effectLabel = (k: string): string => {
    if (k === 'produce_factor') return t('effectProduce');
    if (k === 'build_factor') return t('effectBuild');
    if (k === 'research_factor') return t('effectResearch');
    if (k === 'energy_factor') return t('effectEnergy');
    if (k === 'storage_factor') return t('effectStorage');
    return k;
  };

  const officers = useQuery({
    queryKey: ['officers'],
    queryFn: () => api.get<{ officers: Entry[] | null }>('/api/officers'),
    refetchInterval: 15000,
  });
  const credit = useQuery({
    queryKey: ['artefact-market', 'credit'],
    queryFn: () => api.get<{ credit: number }>('/api/artefact-market/credit'),
    refetchInterval: 15000,
  });

  const activate = useMutation({
    mutationFn: ({ key, autoRenew }: { key: string; autoRenew: boolean }) =>
      api.post<Entry>(`/api/officers/${key}/activate`, { auto_renew: autoRenew }),
    onMutate: async ({ key }) => {
      await qc.cancelQueries({ queryKey: ['officers'] });
      await qc.cancelQueries({ queryKey: ['artefact-market', 'credit'] });
      const prevOfficers = qc.getQueryData<{ officers: Entry[] | null }>(['officers']);
      const prevCredit = qc.getQueryData<{ credit: number }>(['artefact-market', 'credit']);
      const officer = prevOfficers?.officers?.find((e) => e.key === key);
      if (officer) {
        const expiresAt = new Date(Date.now() + officer.duration_days * 86400000).toISOString();
        qc.setQueryData<{ officers: Entry[] | null }>(['officers'], (old) => ({
          officers: old?.officers?.map((e) => e.key === key ? { ...e, activated_at: new Date().toISOString(), expires_at: expiresAt } : e) ?? null,
        }));
        qc.setQueryData<{ credit: number }>(['artefact-market', 'credit'], (old) => ({
          credit: (old?.credit ?? 0) - officer.cost_credit,
        }));
      }
      return { prevOfficers, prevCredit };
    },
    onSuccess: (e) => {
      void qc.invalidateQueries({ queryKey: ['officers'] });
      void qc.invalidateQueries({ queryKey: ['artefact-market'] });
      void qc.invalidateQueries({ queryKey: ['planets'] });
      toast.show('success', t('toastTitle'), t('toastActivated', { name: e.title, days: String(e.duration_days) }));
    },
    onError: (err, _vars, ctx) => {
      if (ctx?.prevOfficers) qc.setQueryData(['officers'], ctx.prevOfficers);
      if (ctx?.prevCredit) qc.setQueryData(['artefact-market', 'credit'], ctx.prevCredit);
      toast.show('danger', t('toastError'), err instanceof Error ? err.message : '');
    },
  });

  const list = officers.data?.officers ?? [];
  const creditVal = credit.data?.credit ?? 0;

  function toggleAutoRenew(key: string, checked: boolean) {
    const next = new Set(autoRenewKeys);
    if (checked) next.add(key); else next.delete(key);
    setAutoRenewKeys(next);
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: 8 }}>
        <h2 style={{ margin: 0, fontSize: 18, fontFamily: 'var(--ox-font)', fontWeight: 700 }}>
          {t('title')}
        </h2>
        <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
          <span style={{ fontSize: 14, color: 'var(--ox-fg-dim)' }}>{t('balance')}</span>
          <span style={{ fontFamily: 'var(--ox-mono)', fontWeight: 700, color: 'var(--ox-accent)', fontSize: 15 }}>
            {creditVal} cr
          </span>
        </div>
      </div>

      {list.length === 0 ? (
        <div style={{ color: 'var(--ox-fg-dim)', fontSize: 16 }}>{t('empty')}</div>
      ) : (
        <div className="ox-cards-grid">
          {list.map((e) => {
            const active = !!e.expires_at;
            const canAfford = creditVal >= e.cost_credit;
            const autoRenew = autoRenewKeys.has(e.key);
            return (
              <div
                key={e.key}
                className="ox-unit-card"
                style={active ? { borderColor: 'var(--ox-success)', boxShadow: '0 0 0 1px var(--ox-success)' } : undefined}
              >
                <div className="ox-unit-card-img" style={{ fontSize: 36, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                  {active ? '⭐' : '👤'}
                </div>
                <div className="ox-unit-card-body">
                  <div className="ox-unit-card-name">{e.title}</div>
                  <div style={{ fontSize: 14, color: 'var(--ox-fg-dim)', marginBottom: 4 }}>{e.description}</div>
                  <div style={{ fontSize: 13, color: 'var(--ox-fg-muted)' }}>
                    {t('daysAndCost', { days: String(e.duration_days), cost: String(e.cost_credit) })}
                  </div>
                  {fmtEffect(e.effect, effectLabel) && (
                    <div style={{ fontSize: 13, color: 'var(--ox-success, #22c55e)', marginTop: 3 }}>
                      ✦ {fmtEffect(e.effect, effectLabel)}
                    </div>
                  )}
                  {active && e.expires_at && (
                    <div style={{ fontSize: 14, color: 'var(--ox-success)', marginTop: 4 }}>
                      {t('expires')} <Countdown finishAt={e.expires_at} />
                    </div>
                  )}
                </div>
                <div className="ox-unit-card-footer" style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
                  {active ? (
                    <div style={{ fontSize: 14, color: 'var(--ox-success)', textAlign: 'center', fontWeight: 600 }}>
                      {t('active')}
                    </div>
                  ) : (
                    <>
                      <label style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: 14, cursor: 'pointer' }}>
                        <input
                          type="checkbox"
                          checked={autoRenew}
                          onChange={(ev) => toggleAutoRenew(e.key, ev.target.checked)}
                        />
                        {t('autoRenew')}
                      </label>
                      <button
                        type="button"
                        className={`btn btn-sm${!canAfford ? ' btn-ghost' : ''}`}
                        style={{ width: '100%' }}
                        disabled={activate.isPending || !canAfford}
                        title={!canAfford ? t('notEnoughTitle') : undefined}
                        onClick={() => activate.mutate({ key: e.key, autoRenew })}
                      >
                        {canAfford ? t('activateBtn', { cost: String(e.cost_credit) }) : t('notEnoughCr')}
                      </button>
                    </>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
