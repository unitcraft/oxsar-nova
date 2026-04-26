import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';
import { useTranslation } from '@/i18n/i18n';
import { SHIPS, nameOf, imageOf, imageOfId } from '@/api/catalog';
import type { Planet } from '@/api/types';
import { Countdown } from '@/ui/Countdown';
import { useToast } from '@/ui/Toast';
import { ScreenSkeleton } from '@/ui/Skeleton';

const GAME_SPEED = 0.75;

function galaxyDistance(
  { galaxy: g1, system: s1, position: p1 }: { galaxy: number; system: number; position: number },
  { galaxy: g2, system: s2, position: p2 }: { galaxy: number; system: number; position: number },
): number {
  if (g1 !== g2) return 20000 * Math.abs(g1 - g2);
  if (s1 !== s2) return 2700 + 95 * Math.abs(s1 - s2);
  if (p1 !== p2) return 1000 + 5 * Math.abs(p1 - p2);
  return 5;
}

function flightSecs(dist: number, minSpeed: number, speedPct: number): number {
  if (minSpeed <= 0) return 60;
  const raw = 10 + (3500 / speedPct) * Math.sqrt((10 * dist) / minSpeed);
  return Math.max(1, raw / GAME_SPEED);
}

function fmtDuration(secs: number, uS: string, uM: string, uH: string, uD: string): string {
  if (secs < 60) return `${Math.ceil(secs)}${uS}`;
  const m = Math.floor(secs / 60) % 60;
  const h = Math.floor(secs / 3600) % 24;
  const d = Math.floor(secs / 86400);
  if (d > 0) return `${d}${uD} ${h}${uH} ${m}${uM}`;
  if (h > 0) return `${h}${uH} ${m}${uM}`;
  return `${m}${uM}`;
}

interface FleetRow {
  id: string;
  src_planet_id: string;
  dst_galaxy: number;
  dst_system: number;
  dst_position: number;
  dst_is_moon: boolean;
  mission: number;
  state: string;
  depart_at: string;
  arrive_at: string;
  return_at?: string | null;
  carry: { metal: number; silicon: number; hydrogen: number };
  ships?: Record<string, number>;
}

const MISSION_IDS = [6, 7, 8, 9, 10, 11, 15] as const;
const MISSION_TKEYS: Record<number, string> = {
  6: 'missionRebase', 7: 'missionTransport', 8: 'missionColonize',
  9: 'missionRecycle', 10: 'missionAttack', 11: 'missionSpy', 15: 'missionExpedition',
};
const STATE_TKEYS: Record<string, string> = {
  outbound: 'stateOutbound', returning: 'stateReturning', arrived: 'stateArrived',
};

interface InitialDst { g: number; s: number; pos: number; isMoon: boolean; mission: number }

export function FleetScreen({ planet, initialDst }: { planet: Planet; initialDst?: InitialDst }) {
  const { t } = useTranslation('fleetUi');
  const { t: tg } = useTranslation('global');
  const uS = tg('timeUnitSec');
  const uM = tg('timeUnitMin');
  const uH = tg('timeUnitHour');
  const uD = tg('timeUnitDay');
  const qc = useQueryClient();
  const toast = useToast();

  const fleets = useQuery({
    queryKey: ['fleets'],
    queryFn: () => api.get<{ fleets: FleetRow[] | null; slots_used: number; slots_max: number }>('/api/fleet'),
    refetchInterval: 3000,
  });

  const [g, setG] = useState(initialDst?.g ?? planet.galaxy);
  const [s, setS] = useState(initialDst?.s ?? planet.system);
  const [pos, setPos] = useState(initialDst?.pos ?? planet.position);
  const [isMoon, setIsMoon] = useState(initialDst?.isMoon ?? false);
  const [speed, setSpeed] = useState(100);
  const [metal, setMetal] = useState(0);
  const [silicon, setSilicon] = useState(0);
  const [hydrogen, setHydrogen] = useState(0);
  const [ships, setShips] = useState<Record<number, number>>({});
  const [mission, setMission] = useState(initialDst?.mission ?? 7);
  const [colonyName, setColonyName] = useState('');

  const send = useMutation({
    mutationFn: () => {
      const carryAllowed = mission === 6 || mission === 7 || mission === 8;
      return api.post<unknown>('/api/fleet', {
        src_planet_id: planet.id,
        dst: { galaxy: g, system: s, position: pos, is_moon: isMoon },
        ships: Object.fromEntries(Object.entries(ships).filter(([, n]) => Number(n) > 0)),
        carry_metal: carryAllowed ? metal : 0,
        carry_silicon: carryAllowed ? silicon : 0,
        carry_hydrogen: carryAllowed ? hydrogen : 0,
        speed_percent: speed,
        mission,
        colony_name: mission === 8 ? colonyName : undefined,
      });
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['fleets'] });
      void qc.invalidateQueries({ queryKey: ['planets'] });
      setShips({});
      setMetal(0); setSilicon(0); setHydrogen(0);
      const missionKey = MISSION_TKEYS[mission];
      const missionLabel = missionKey ? t(missionKey) : t('missionFallback');
      toast.show('success', t('sent'), t('sentBody', { mission: missionLabel, coords: `${g}:${s}:${pos}` }));
    },
    onError: (err) => {
      toast.show('danger', tg('error'), err instanceof Error ? err.message : t('sendErr'));
    },
  });

  const recall = useMutation({
    mutationFn: (id: string) => api.post<unknown>(`/api/fleet/${id}/recall`),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['fleets'] });
      toast.show('info', t('recalled'));
    },
    onError: (err) => {
      toast.show('danger', tg('error'), err instanceof Error ? err.message : t('recallErr'));
    },
  });

  const list = fleets.data?.fleets ?? [];
  const totalShips = Object.values(ships).reduce((a, b) => a + b, 0);
  const totalCargo = SHIPS.reduce((sum, ship) => sum + (ship.cargo ?? 0) * (ships[ship.id] ?? 0), 0);

  const fleetPreview = (() => {
    if (totalShips === 0) return null;
    const selectedShips = SHIPS.filter((s) => (ships[s.id] ?? 0) > 0);
    const minSpeed = Math.min(...selectedShips.map((s) => s.speed ?? Infinity));
    if (!isFinite(minSpeed) || minSpeed <= 0) return null;
    const dist = galaxyDistance(
      { galaxy: planet.galaxy, system: planet.system, position: planet.position },
      { galaxy: g, system: s, position: pos },
    );
    const secs = flightSecs(dist, minSpeed, speed);
    const totalFuel = selectedShips.reduce((sum, ship) => {
      const count = ships[ship.id] ?? 0;
      const f = ship.fuel ?? 0;
      return sum + Math.round(f * dist / 35000 * (speed / 100 + 1) ** 2) * count;
    }, 0);
    return { secs, totalFuel };
  })();

  if (fleets.isLoading) {
    return <ScreenSkeleton />;
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <h2 style={{ margin: 0, fontSize: 18, fontFamily: 'var(--ox-font)', fontWeight: 700 }}>
        {t('title', { planetName: planet.name })}
      </h2>

      {/* Fleet slots indicator (план 20 Ф.2) */}
      {typeof fleets.data?.slots_max === 'number' && (
        <div className="ox-panel" style={{ padding: '8px 16px', fontSize: 13, color: 'var(--ox-fg-muted)' }}>
          {t('slots')} <strong style={{ color: (fleets.data.slots_used ?? 0) >= fleets.data.slots_max ? 'var(--ox-danger, #ff6b6b)' : 'var(--ox-fg)' }}>
            {fleets.data.slots_used ?? 0} / {fleets.data.slots_max}
          </strong>
          {' '}<span style={{ opacity: 0.7 }}>{t('slotsHint')}</span>
        </div>
      )}

      {/* Active fleets */}
      {list.length > 0 && (
        <div className="ox-panel" style={{ overflow: 'hidden' }}>
          <div style={{ padding: '10px 16px 8px', fontSize: 13, fontWeight: 700, letterSpacing: '0.08em', textTransform: 'uppercase', color: 'var(--ox-fg-muted)', borderBottom: '1px solid var(--ox-border)' }}>
            {t('activeFleets', { count: String(list.length) })}
          </div>
          <div style={{ overflowX: 'auto' }}>
            <table className="ox-table" style={{ margin: 0 }}>
              <thead>
                <tr>
                  <th>{t('colMission')}</th>
                  <th>{t('colDestination')}</th>
                  <th>{t('colComposition')}</th>
                  <th>{t('colStatus')}</th>
                  <th>{t('colArrival')}</th>
                  <th />
                </tr>
              </thead>
              <tbody>
                {list.map((f) => (
                  <tr key={f.id}>
                    <td>{(() => { const k = MISSION_TKEYS[f.mission]; return k ? t(k) : `#${f.mission}`; })()}</td>
                    <td style={{ fontFamily: 'var(--ox-mono)', fontSize: 14 }}>
                      [{f.dst_galaxy}:{f.dst_system}:{f.dst_position}{f.dst_is_moon ? '🌑' : ''}]
                    </td>
                    <td>
                      {f.ships && Object.keys(f.ships).length > 0 && (
                        <div style={{ display: 'flex', flexWrap: 'wrap', gap: '2px 8px' }}>
                          {Object.entries(f.ships).map(([unitId, count]) => {
                            const id = Number(unitId);
                            const img = imageOfId(id);
                            return (
                              <span key={unitId} style={{ display: 'inline-flex', alignItems: 'center', gap: 3, fontSize: 13, color: 'var(--ox-fg-dim)', fontFamily: 'var(--ox-mono)' }}>
                                {img && <img src={img} alt="" width={14} height={14} style={{ imageRendering: 'pixelated', opacity: 0.85 }} />}
                                {nameOf(id)} ×{count}
                              </span>
                            );
                          })}
                        </div>
                      )}
                    </td>
                    <td>
                      <span className={`ox-badge${f.state === 'outbound' ? ' ox-badge-accent' : ''}`}>
                        {(() => { const k = STATE_TKEYS[f.state]; return k ? t(k) : f.state; })()}
                      </span>
                    </td>
                    <td style={{ fontFamily: 'var(--ox-mono)', fontSize: 14 }}>
                      <div><Countdown finishAt={f.state === 'outbound' ? f.arrive_at : (f.return_at ?? f.arrive_at)} /></div>
                    </td>
                    <td>
                      {f.state === 'outbound' && (
                        <button
                          type="button"
                          className="btn-ghost btn-sm"
                          disabled={recall.isPending}
                          onClick={() => recall.mutate(f.id)}
                        >
                          {t('recall')}
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Send form */}
      <div className="ox-panel" style={{ padding: 20 }}>
        <div style={{ fontSize: 15, fontWeight: 700, letterSpacing: '0.06em', textTransform: 'uppercase', color: 'var(--ox-fg-muted)', marginBottom: 16 }}>
          {t('newMission')}
        </div>

        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(260px, 1fr))', gap: 20 }}>
          {/* Mission & destination */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
            <div>
              <label style={{ fontSize: 14, color: 'var(--ox-fg-dim)', display: 'block', marginBottom: 4 }}>{t('missionLabel')}</label>
              <select value={mission} onChange={(e) => setMission(Number(e.target.value))} style={{ width: '100%' }}>
                {MISSION_IDS.map((id) => {
                  const key = MISSION_TKEYS[id] ?? 'missionFallback';
                  return <option key={id} value={id}>{t(key)}</option>;
                })}
              </select>
            </div>

            <div>
              <label style={{ fontSize: 14, color: 'var(--ox-fg-dim)', display: 'block', marginBottom: 4 }}>{t('destination')}</label>
              <div style={{ display: 'flex', gap: 6, alignItems: 'center', flexWrap: 'wrap' }}>
                <span style={{ fontSize: 14, color: 'var(--ox-fg-muted)' }}>G</span>
                <input type="number" min={1} max={16} value={g} onChange={(e) => setG(Number(e.target.value))} style={{ width: 56 }} />
                <span style={{ fontSize: 14, color: 'var(--ox-fg-muted)' }}>S</span>
                <input type="number" min={1} max={999} value={s} onChange={(e) => setS(Number(e.target.value))} style={{ width: 70 }} />
                <span style={{ fontSize: 14, color: 'var(--ox-fg-muted)' }}>P</span>
                <input type="number" min={1} max={15} value={pos} onChange={(e) => setPos(Number(e.target.value))} style={{ width: 56 }} />
                <label style={{ display: 'flex', alignItems: 'center', gap: 4, fontSize: 14 }}>
                  <input type="checkbox" checked={isMoon} onChange={(e) => setIsMoon(e.target.checked)} />
                  🌑
                </label>
              </div>
            </div>

            {mission === 8 && (
              <div>
                <label style={{ fontSize: 14, color: 'var(--ox-fg-dim)', display: 'block', marginBottom: 4 }}>{t('colonyName')}</label>
                <input type="text" value={colonyName} onChange={(e) => setColonyName(e.target.value)} placeholder="Colony" maxLength={40} style={{ width: '100%' }} />
              </div>
            )}

            {(mission === 6 || mission === 7 || mission === 8) && (
              <div>
                <label style={{ fontSize: 14, color: 'var(--ox-fg-dim)', display: 'block', marginBottom: 4 }}>
                  {t('cargo')}
                  {totalCargo > 0 && (
                    <span style={{ marginLeft: 8, fontFamily: 'var(--ox-mono)', color: 'var(--ox-fg-muted)' }}>
                      📦 {t('cargoMax')} {totalCargo.toLocaleString('ru-RU')}
                    </span>
                  )}
                </label>
                <div style={{ display: 'flex', gap: 6, flexWrap: 'wrap' }}>
                  <input type="number" min={0} value={metal} onChange={(e) => setMetal(Number(e.target.value))} placeholder={t('metalPh')} style={{ flex: 1, minWidth: 80 }} />
                  <input type="number" min={0} value={silicon} onChange={(e) => setSilicon(Number(e.target.value))} placeholder={t('siliconPh')} style={{ flex: 1, minWidth: 80 }} />
                  <input type="number" min={0} value={hydrogen} onChange={(e) => setHydrogen(Number(e.target.value))} placeholder={t('hydrogenPh')} style={{ flex: 1, minWidth: 80 }} />
                </div>
              </div>
            )}

            <div>
              <label style={{ fontSize: 14, color: 'var(--ox-fg-dim)', display: 'block', marginBottom: 4 }}>
                {t('speedPct', { pct: String(speed) })}
              </label>
              <input type="range" min={10} max={100} step={10} value={speed} onChange={(e) => setSpeed(Number(e.target.value))} style={{ width: '100%' }} />
            </div>
          </div>

          {/* Ships selection */}
          <div>
            <div style={{ fontSize: 14, color: 'var(--ox-fg-dim)', marginBottom: 8 }}>
              {t('ships')} {totalShips > 0 && <span style={{ color: 'var(--ox-accent)', fontWeight: 700 }}>({t('selectedCount', { count: String(totalShips) })})</span>}
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
              {SHIPS.map((ship) => (
                <div key={ship.id} style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                  <img src={imageOf(ship.key)} alt="" width={28} height={28} style={{ imageRendering: 'pixelated', flexShrink: 0 }} />
                  <span style={{ flex: 1, fontSize: 15 }}>{nameOf(ship.id)}</span>
                  <input
                    type="number"
                    min={0}
                    value={ships[ship.id] ?? 0}
                    onChange={(e) => setShips({ ...ships, [ship.id]: Math.max(0, Number(e.target.value)) })}
                    style={{ width: 80 }}
                  />
                </div>
              ))}
            </div>
          </div>
        </div>

        {fleetPreview && (
          <div style={{ marginTop: 16, padding: '10px 14px', background: 'var(--ox-surface)', borderRadius: 6, border: '1px solid var(--ox-border)', display: 'flex', gap: 20, flexWrap: 'wrap', fontSize: 14, fontFamily: 'var(--ox-mono)', color: 'var(--ox-fg-dim)' }}>
            <span>⏱ {fmtDuration(fleetPreview.secs, uS, uM, uH, uD)}</span>
            <span>↩ {fmtDuration(fleetPreview.secs * 2, uS, uM, uH, uD)}</span>
            {fleetPreview.totalFuel > 0 && <span>💧 {fleetPreview.totalFuel.toLocaleString('ru-RU')} ({t('fuelOneWay')})</span>}
          </div>
        )}

        <div style={{ marginTop: 20, display: 'flex', gap: 10, justifyContent: 'flex-end' }}>
          <button
            type="button"
            className="btn"
            disabled={send.isPending || totalShips === 0}
            onClick={() => send.mutate()}
          >
            {send.isPending ? '…' : `🚀 ${t('sendButton')}`}
          </button>
        </div>
      </div>
    </div>
  );
}
