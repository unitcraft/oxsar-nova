import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { api } from '@/api/client';
import { nameOf } from '@/api/catalog';
import { useToast } from '@/ui/Toast';
import { useTranslation } from '@/i18n/i18n';

type FleetStack = { unit_id: number; quantity: number };

type Holding = {
  event_id: string;
  planet_id: string;
  planet_name: string;
  galaxy: number;
  system: number;
  position: number;
  tier: number;
  start_time: string;
  ends_at: string;
  paid_credit: number;
  paid_times: number;
  max_ends_at: string;
  alien_fleet: FleetStack[];
};

function formatRemaining(endsAt: string, expires: string, unitDay: string, unitHour: string, unitMin: string): string {
  const ms = new Date(endsAt).getTime() - Date.now();
  if (ms <= 0) return expires;
  const h = Math.floor(ms / 3_600_000);
  const m = Math.floor((ms % 3_600_000) / 60_000);
  if (h > 24) return `${Math.floor(h / 24)}${unitDay} ${h % 24}${unitHour}`;
  if (h > 0) return `${h}${unitHour} ${m}${unitMin}`;
  return `${m}${unitMin}`;
}

/**
 * AlienHoldingPanel — список планет, захваченных пришельцами (план 15 этап 4).
 * Виджет под Overview / отдельный экран. Показывает таймер + alien-флот +
 * кнопку оплаты продления.
 */
export function AlienHoldingPanel() {
  const { t } = useTranslation('alien');
  const { t: ti } = useTranslation('info');
  const qc = useQueryClient();
  const toast = useToast();
  const [amount, setAmount] = useState<Record<string, number>>({});

  const q = useQuery({
    queryKey: ['alien-holdings'],
    queryFn: () => api.get<{ holdings: Holding[] }>('/api/alien/holdings/me'),
    refetchInterval: 60_000,
  });

  const pay = useMutation({
    mutationFn: ({ eventID, amt }: { eventID: string; amt: number }) =>
      api.post<{ extended_seconds: number; new_ends_at: string; capped: boolean }>(
        `/api/alien/holding/${eventID}/pay`,
        { amount: amt }
      ),
    onSuccess: (res, vars) => {
      void qc.invalidateQueries({ queryKey: ['alien-holdings'] });
      void qc.invalidateQueries({ queryKey: ['me'] });
      const minutes = Math.round(res.extended_seconds / 60);
      toast.show('success', t('payBtn'), `+${t('payMinutes', { n: String(minutes) })}`);
      setAmount({ ...amount, [vars.eventID]: 0 });
    },
    onError: (err) => {
      toast.show('danger', t('payBtn'), err instanceof Error ? err.message : '');
    },
  });

  const list = q.data?.holdings ?? [];
  if (list.length === 0) return null;

  return (
    <div className="ox-panel" style={{ padding: 12, margin: '8px 16px' }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8 }}>
        <span style={{ fontSize: 20 }}>👽</span>
        <strong>{t('title')}</strong>
      </div>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
        {list.map((h) => (
          <div key={h.event_id} style={{ borderTop: '1px solid var(--ox-border)', paddingTop: 8 }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline' }}>
              <div>
                <strong>{h.planet_name}</strong>
                <span style={{ marginLeft: 8, color: 'var(--ox-fg-muted)' }}>
                  [{h.galaxy}:{h.system}:{h.position}] · {t('tierLabel')} {h.tier}
                </span>
              </div>
              <div style={{ fontSize: 13, color: 'var(--ox-fg-muted)' }}>
                {t('departsLabel')} <strong>{formatRemaining(h.ends_at, t('expiresLabel'), t('global', 'timeUnitDay'), t('global', 'timeUnitHour'), t('global', 'timeUnitMin'))}</strong>
              </div>
            </div>

            {h.alien_fleet.length > 0 && (
              <div style={{ marginTop: 4, fontSize: 13 }}>
                {t('fleetLabel')} {h.alien_fleet.map((s) => `${s.quantity}× ${nameOf(s.unit_id, ti)}`).join(', ')}
              </div>
            )}

            {h.paid_times > 0 && (
              <div style={{ fontSize: 12, color: 'var(--ox-fg-muted)' }}>
                {t('paidLabel')} {h.paid_credit} (×{h.paid_times})
              </div>
            )}

            <div style={{ display: 'flex', gap: 8, marginTop: 8, alignItems: 'center' }}>
              <input
                type="number"
                min={50}
                step={50}
                placeholder="50+"
                value={amount[h.event_id] ?? ''}
                onChange={(e) => setAmount({ ...amount, [h.event_id]: Number(e.target.value) })}
                style={{ width: 100, padding: '4px 8px' }}
              />
              <button
                className="ox-btn ox-btn-primary"
                disabled={(amount[h.event_id] ?? 0) <= 0 || pay.isPending}
                onClick={() => pay.mutate({ eventID: h.event_id, amt: amount[h.event_id] ?? 0 })}
              >
                {t('payBtn')}
              </button>
              <span style={{ fontSize: 12, color: 'var(--ox-fg-muted)' }}>
                {t('payHint')}
              </span>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
