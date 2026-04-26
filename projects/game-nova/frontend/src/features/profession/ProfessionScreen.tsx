import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';
import { useToast } from '@/ui/Toast';
import { useTranslation } from '@/i18n/i18n';

interface ProfessionDTO {
  key: string;
  label: string;
  bonus?: Record<string, number>;
  malus?: Record<string, number>;
}

interface CurrentInfo {
  profession: string;
  label: string;
  next_change_allowed?: string | null;
}

const BONUS_KEY: Record<string, string> = {
  metalmine: 'bonusProductionMetal',
  silicon_lab: 'bonusProductionSilicon',
  solar_plant: 'bonusProductionHydrogen',
  shipyard: 'bonusBuildSpeed',
  gun: 'bonusShipAttack',
  shield_weapon: 'bonusShipShield',
  shell_weapon: 'bonusShipHull',
  ballistics: 'bonusFleetSpeed',
  masking: 'bonusEspionage',
  defense_factory: 'bonusBuildSpeed',
  rocket_station: 'bonusBuildSpeed',
  computer_tech: 'bonusResearchSpeed',
  gravi: 'bonusResearchSpeed',
  combustion_drive: 'bonusFleetSpeed',
  impulse_drive: 'bonusFleetSpeed',
  hyperspace_drive: 'bonusFleetSpeed',
};

const PROFESSION_ICONS: Record<string, string> = {
  miner: '⛏️',
  attacker: '⚔️',
  defender: '🛡️',
  tank: '🔫',
  none: '⚪',
};

const PROFESSION_ORDER = ['miner', 'attacker', 'defender', 'tank'];

function fmtDelta(v: number): string {
  return v > 0 ? `+${v}` : String(v);
}

function timeUntil(iso: string, avail: string, unitDay: string, unitHour: string, unitMin: string): string {
  const ms = new Date(iso).getTime() - Date.now();
  if (ms <= 0) return avail;
  const d = Math.floor(ms / 86400000);
  const h = Math.floor((ms % 86400000) / 3600000);
  if (d > 0) return `${d}${unitDay} ${h}${unitHour}`;
  const m = Math.floor((ms % 3600000) / 60000);
  return `${h}${unitHour} ${m}${unitMin}`;
}

export function ProfessionScreen() {
  const { t } = useTranslation('profession');
  const qc = useQueryClient();
  const toast = useToast();
  const unitDay  = t('global', 'timeUnitDay');
  const unitHour = t('global', 'timeUnitHour');
  const unitMin  = t('global', 'timeUnitMin');

  const list = useQuery({
    queryKey: ['professions'],
    queryFn: () => api.get<{ professions: ProfessionDTO[] }>('/api/professions'),
    staleTime: 300000,
  });

  const current = useQuery({
    queryKey: ['professions', 'me'],
    queryFn: () => api.get<CurrentInfo>('/api/professions/me'),
    refetchInterval: 30000,
  });

  const change = useMutation({
    mutationFn: (key: string) => api.post('/api/professions/me', { profession: key }),
    onSuccess: (_data, key) => {
      void qc.invalidateQueries({ queryKey: ['professions', 'me'] });
      void qc.invalidateQueries({ queryKey: ['me'] });
      const prof = list.data?.professions.find((p) => p.key === key);
      toast.show('success', t('title'), prof?.label ?? '');
    },
    onError: (err: unknown) => {
      const msg = (err as { message?: string })?.message ?? '';
      toast.show('danger', t('title'), msg);
    },
  });

  const professions = [...(list.data?.professions ?? [])].sort(
    (a, b) => PROFESSION_ORDER.indexOf(a.key) - PROFESSION_ORDER.indexOf(b.key)
  );

  const currentKey = current.data?.profession ?? 'none';
  const nextChange = current.data?.next_change_allowed;
  const canChangeFree = !nextChange || new Date(nextChange).getTime() <= Date.now();

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20, padding: '16px 0' }}>

      <div style={{ display: 'flex', alignItems: 'center', gap: 12, flexWrap: 'wrap' }}>
        <h2 style={{ margin: 0, fontSize: 18, fontWeight: 700 }}>{t('title')}</h2>
        {current.data && (
          <span style={{
            padding: '3px 10px', borderRadius: 20,
            background: currentKey === 'none' ? 'var(--ox-bg-card)' : 'rgba(56,189,248,0.12)',
            border: '1px solid var(--ox-border)',
            fontSize: 15, color: 'var(--ox-fg-dim)',
          }}>
            {PROFESSION_ICONS[currentKey] ?? '⚪'} {t('currentLabel')}: <b style={{ color: 'var(--ox-fg)' }}>{current.data.label || t('available')}</b>
          </span>
        )}
      </div>

      {/* Информация о смене */}
      <div className="ox-panel" style={{ padding: '12px 16px', fontSize: 15, color: 'var(--ox-fg-dim)', display: 'flex', flexWrap: 'wrap', gap: 16, alignItems: 'center' }}>
        {canChangeFree ? (
          <span style={{ color: 'var(--ox-success, #22c55e)', fontWeight: 600 }}>
            ✅ {t('available')}
          </span>
        ) : nextChange ? (
          <span style={{ color: 'var(--ox-warn, #f59e0b)' }}>
            🕐 <b>{timeUntil(nextChange, t('available'), unitDay, unitHour, unitMin)}</b>
          </span>
        ) : null}
      </div>

      {/* Карточки профессий */}
      {list.isLoading ? (
        <div style={{ display: 'flex', gap: 12 }}>
          {[1, 2, 3, 4].map((i) => (
            <div key={i} className="ox-skeleton" style={{ height: 220, flex: 1, borderRadius: 10 }} />
          ))}
        </div>
      ) : (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(220px, 1fr))', gap: 14 }}>
          {professions.map((p) => {
            const isActive = p.key === currentKey;
            const isChanging = change.isPending && change.variables === p.key;
            return (
              <div
                key={p.key}
                className="ox-unit-card"
                style={{
                  display: 'flex', flexDirection: 'column', gap: 0,
                  borderColor: isActive ? 'var(--ox-success, #22c55e)' : undefined,
                  boxShadow: isActive ? '0 0 0 1px var(--ox-success, #22c55e)' : undefined,
                  opacity: isChanging ? 0.7 : 1,
                  transition: 'box-shadow 0.2s, border-color 0.2s, opacity 0.2s',
                }}
              >
                <div className="ox-unit-card-img" style={{ fontSize: 36, textAlign: 'center', paddingTop: 8 }}>
                  {PROFESSION_ICONS[p.key] ?? '⚪'}
                </div>
                <div className="ox-unit-card-body" style={{ flex: 1 }}>
                  <div className="ox-unit-card-name" style={{ fontSize: 15, fontWeight: 700, marginBottom: 8 }}>
                    {p.label}
                    {isActive && <span style={{ fontSize: 13, marginLeft: 6, color: 'var(--ox-success, #22c55e)', fontWeight: 400 }}>● {t('active')}</span>}
                  </div>

                  {/* Бонусы */}
                  {p.bonus && Object.keys(p.bonus).length > 0 && (
                    <div style={{ marginBottom: 6 }}>
                      {Object.entries(p.bonus).map(([k, v]) => (
                        <div key={k} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', fontSize: 14, padding: '2px 0' }}>
                          <span style={{ color: 'var(--ox-fg-dim)' }}>{BONUS_KEY[k] ? t(BONUS_KEY[k]!) : k}</span>
                          <span style={{ color: 'var(--ox-success, #22c55e)', fontWeight: 700, fontFamily: 'var(--ox-mono)', minWidth: 28, textAlign: 'right' }}>
                            {fmtDelta(v)}
                          </span>
                        </div>
                      ))}
                    </div>
                  )}

                  {/* Штрафы */}
                  {p.malus && Object.keys(p.malus).length > 0 && (
                    <div>
                      {Object.entries(p.malus).map(([k, v]) => (
                        <div key={k} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', fontSize: 14, padding: '2px 0' }}>
                          <span style={{ color: 'var(--ox-fg-muted)' }}>{BONUS_KEY[k] ? t(BONUS_KEY[k]!) : k}</span>
                          <span style={{ color: 'var(--ox-danger)', fontWeight: 700, fontFamily: 'var(--ox-mono)', minWidth: 28, textAlign: 'right' }}>
                            {fmtDelta(v)}
                          </span>
                        </div>
                      ))}
                    </div>
                  )}
                </div>

                <div className="ox-unit-card-footer" style={{ paddingTop: 10 }}>
                  {isActive ? (
                    <div style={{ textAlign: 'center', fontSize: 14, color: 'var(--ox-success, #22c55e)', fontWeight: 600 }}>
                      ✅ {t('currentLabel')}
                    </div>
                  ) : (
                    <button
                      type="button"
                      className="btn btn-sm"
                      style={{ width: '100%' }}
                      disabled={change.isPending}
                      onClick={() => {
                        if (confirm(t('confirmChoose', { name: p.label, days: '14' }))) {
                          change.mutate(p.key);
                        }
                      }}
                    >
                      {isChanging ? '…' : t('chooseBtn')}
                    </button>
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
