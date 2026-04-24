import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';
import { useToast } from '@/ui/Toast';

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

const BONUS_LABELS: Record<string, string> = {
  metalmine: 'Рудник металла',
  silicon_lab: 'Рудник кремния',
  solar_plant: 'Солнечная электростанция',
  shipyard: 'Верфь',
  gun: 'Оружейная технология',
  shield_weapon: 'Щитовая технология',
  shell_weapon: 'Броневая технология',
  ballistics: 'Баллистика',
  masking: 'Маскировка',
  defense_factory: 'Оборонный завод',
  rocket_station: 'Ракетная шахта',
  computer_tech: 'Компьютерная технология',
  gravi: 'Гравитационная технология',
  combustion_drive: 'Реактивный двигатель',
  impulse_drive: 'Импульсный двигатель',
  hyperspace_drive: 'Гиперпространственный двигатель',
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

function timeUntil(iso: string): string {
  const ms = new Date(iso).getTime() - Date.now();
  if (ms <= 0) return 'доступно';
  const d = Math.floor(ms / 86400000);
  const h = Math.floor((ms % 86400000) / 3600000);
  if (d > 0) return `${d}д ${h}ч`;
  const m = Math.floor((ms % 3600000) / 60000);
  return `${h}ч ${m}м`;
}

export function ProfessionScreen() {
  const qc = useQueryClient();
  const toast = useToast();

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
      toast.show('success', 'Профессия изменена', prof ? `Теперь вы ${prof.label}` : '');
    },
    onError: (err: unknown) => {
      const msg = (err as { message?: string })?.message ?? 'Ошибка';
      if (msg.includes('cooldown') || msg.includes('too soon')) {
        toast.show('danger', 'Слишком рано', 'Смена профессии доступна раз в 14 дней');
      } else if (msg.includes('credit')) {
        toast.show('danger', 'Недостаточно кредитов', 'Смена профессии стоит 1000 кредитов');
      } else {
        toast.show('danger', 'Ошибка', msg);
      }
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
        <h2 style={{ margin: 0, fontSize: 18, fontWeight: 700 }}>Профессия</h2>
        {current.data && (
          <span style={{
            padding: '3px 10px', borderRadius: 20,
            background: currentKey === 'none' ? 'var(--ox-bg-card)' : 'rgba(56,189,248,0.12)',
            border: '1px solid var(--ox-border)',
            fontSize: 15, color: 'var(--ox-fg-dim)',
          }}>
            {PROFESSION_ICONS[currentKey] ?? '⚪'} Сейчас: <b style={{ color: 'var(--ox-fg)' }}>{current.data.label || 'Нет профессии'}</b>
          </span>
        )}
      </div>

      {/* Информация о смене */}
      <div className="ox-panel" style={{ padding: '12px 16px', fontSize: 15, color: 'var(--ox-fg-dim)', display: 'flex', flexWrap: 'wrap', gap: 16, alignItems: 'center' }}>
        <span>
          💳 Стоимость смены: <b style={{ color: 'var(--ox-accent)' }}>1 000 кредитов</b>
        </span>
        <span>
          ⏱ Кулдаун: <b>14 дней</b>
        </span>
        {canChangeFree ? (
          <span style={{ color: 'var(--ox-success, #22c55e)', fontWeight: 600 }}>
            ✅ Смена сейчас бесплатна
          </span>
        ) : nextChange ? (
          <span style={{ color: 'var(--ox-warn, #f59e0b)' }}>
            🕐 Бесплатная смена через: <b>{timeUntil(nextChange)}</b>
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
                    {isActive && <span style={{ fontSize: 13, marginLeft: 6, color: 'var(--ox-success, #22c55e)', fontWeight: 400 }}>● активна</span>}
                  </div>

                  {/* Бонусы */}
                  {p.bonus && Object.keys(p.bonus).length > 0 && (
                    <div style={{ marginBottom: 6 }}>
                      {Object.entries(p.bonus).map(([k, v]) => (
                        <div key={k} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', fontSize: 14, padding: '2px 0' }}>
                          <span style={{ color: 'var(--ox-fg-dim)' }}>{BONUS_LABELS[k] ?? k}</span>
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
                          <span style={{ color: 'var(--ox-fg-muted)' }}>{BONUS_LABELS[k] ?? k}</span>
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
                      ✅ Текущая профессия
                    </div>
                  ) : (
                    <button
                      type="button"
                      className="btn btn-sm"
                      style={{ width: '100%' }}
                      disabled={change.isPending}
                      onClick={() => {
                        const label = canChangeFree ? '' : ' (1 000 кредитов)';
                        if (confirm(`Сменить профессию на "${p.label}"?${label}`)) {
                          change.mutate(p.key);
                        }
                      }}
                    >
                      {isChanging ? 'Смена…' : `Выбрать${canChangeFree ? '' : ' (1000 💳)'}`}
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
