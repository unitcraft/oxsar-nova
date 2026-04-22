import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';
import { useToast } from '@/ui/Toast';
import { Countdown } from '@/ui/Countdown';

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

const EFFECT_LABELS: Record<string, string> = {
  produce_factor:  'Производство',
  build_factor:    'Строительство',
  research_factor: 'Исследования',
  energy_factor:   'Энергия',
  storage_factor:  'Склад',
};

function fmtEffect(effect: Record<string, number> | null | undefined): string | null {
  if (!effect) return null;
  return Object.entries(effect)
    .filter(([, v]) => v !== 1)
    .map(([k, v]) => {
      const label = EFFECT_LABELS[k] ?? k;
      const pct = Math.round((v - 1) * 100);
      return `${label} ${pct > 0 ? '+' : ''}${pct}%`;
    })
    .join(', ') || null;
}

export function OfficersScreen() {
  const qc = useQueryClient();
  const toast = useToast();
  const [autoRenewKeys, setAutoRenewKeys] = useState<Set<string>>(new Set());

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
    onSuccess: (e) => {
      void qc.invalidateQueries({ queryKey: ['officers'] });
      void qc.invalidateQueries({ queryKey: ['artefact-market'] });
      void qc.invalidateQueries({ queryKey: ['planets'] });
      toast.show('success', 'Офицер', `${e.title} активирован на ${e.duration_days} дн.`);
    },
    onError: (err) => { toast.show('danger', 'Ошибка', err instanceof Error ? err.message : ''); },
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
          🎖 Офицеры
        </h2>
        <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
          <span style={{ fontSize: 12, color: 'var(--ox-fg-dim)' }}>Баланс:</span>
          <span style={{ fontFamily: 'var(--ox-mono)', fontWeight: 700, color: 'var(--ox-accent)', fontSize: 15 }}>
            {creditVal} cr
          </span>
        </div>
      </div>

      {list.length === 0 ? (
        <div style={{ color: 'var(--ox-fg-dim)', fontSize: 14 }}>Нет доступных офицеров.</div>
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
                  <div style={{ fontSize: 12, color: 'var(--ox-fg-dim)', marginBottom: 4 }}>{e.description}</div>
                  <div style={{ fontSize: 11, color: 'var(--ox-fg-muted)' }}>
                    {e.duration_days} дн. · {e.cost_credit} cr
                  </div>
                  {fmtEffect(e.effect) && (
                    <div style={{ fontSize: 11, color: 'var(--ox-success, #22c55e)', marginTop: 3 }}>
                      ✦ {fmtEffect(e.effect)}
                    </div>
                  )}
                  {active && e.expires_at && (
                    <div style={{ fontSize: 12, color: 'var(--ox-success)', marginTop: 4 }}>
                      Истекает: <Countdown finishAt={e.expires_at} />
                    </div>
                  )}
                </div>
                <div className="ox-unit-card-footer" style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
                  {active ? (
                    <div style={{ fontSize: 12, color: 'var(--ox-success)', textAlign: 'center', fontWeight: 600 }}>
                      ✅ Активен
                    </div>
                  ) : (
                    <>
                      <label style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: 12, cursor: 'pointer' }}>
                        <input
                          type="checkbox"
                          checked={autoRenew}
                          onChange={(ev) => toggleAutoRenew(e.key, ev.target.checked)}
                        />
                        Авто-продление
                      </label>
                      <button
                        type="button"
                        className={`btn btn-sm${!canAfford ? ' btn-ghost' : ''}`}
                        style={{ width: '100%' }}
                        disabled={activate.isPending || !canAfford}
                        title={!canAfford ? 'Недостаточно кредитов' : undefined}
                        onClick={() => activate.mutate({ key: e.key, autoRenew })}
                      >
                        {canAfford ? `Активировать (${e.cost_credit} cr)` : 'Мало cr'}
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
