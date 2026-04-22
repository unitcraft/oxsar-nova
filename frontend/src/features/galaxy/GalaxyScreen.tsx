import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '@/api/client';
import { planetImageOf } from '@/api/catalog';
import type { Planet } from '@/api/types';

interface CellView {
  position: number;
  has_planet: boolean;
  planet_name?: string;
  planet_id?: string | null;
  planet_type?: string | null;
  has_moon: boolean;
  moon_name?: string;
  owner_username?: string;
  owner_id?: string;
  owner_rank?: number;
  debris_metal: number;
  debris_silicon: number;
}

interface SystemView {
  galaxy: number;
  system: number;
  cells: CellView[];
}

function formatNum(v: number): string {
  if (v >= 1_000_000) return (v / 1_000_000).toFixed(1) + 'M';
  if (v >= 1_000) return (v / 1_000).toFixed(0) + 'K';
  return Math.floor(v).toLocaleString('ru-RU');
}

function clamp(v: number, lo: number, hi: number): number {
  if (Number.isNaN(v)) return lo;
  return Math.max(lo, Math.min(hi, v));
}

// Кнопки миссий для строки галактики. Открывает FleetScreen с предзаполненными координатами.
// Для MVP: переход через state не реализован, поэтому кнопки используют navigator.clipboard.
// TODO: пробросить setTab + координаты через контекст или колбэк.
function MissionButtons({ cell, onMission }: {
  cell: CellView;
  onMission: (mission: number, position: number, isMoon: boolean) => void;
}) {
  if (!cell.has_planet) return null;
  return (
    <div style={{ display: 'flex', gap: 4, flexWrap: 'wrap' }}>
      <button
        type="button"
        className="btn-ghost btn-sm"
        style={{ fontSize: 11, padding: '2px 7px' }}
        title="Шпионаж"
        onClick={() => onMission(11, cell.position, false)}
      >🔭</button>
      <button
        type="button"
        className="btn-ghost btn-sm"
        style={{ fontSize: 11, padding: '2px 7px' }}
        title="Атака"
        onClick={() => onMission(10, cell.position, false)}
      >⚔️</button>
      <button
        type="button"
        className="btn-ghost btn-sm"
        style={{ fontSize: 11, padding: '2px 7px' }}
        title="Транспорт"
        onClick={() => onMission(7, cell.position, false)}
      >📦</button>
      {(cell.debris_metal > 0 || cell.debris_silicon > 0) && (
        <button
          type="button"
          className="btn-ghost btn-sm"
          style={{ fontSize: 11, padding: '2px 7px' }}
          title="Переработка обломков"
          onClick={() => onMission(9, cell.position, false)}
        >♻️</button>
      )}
      {cell.has_moon && (
        <button
          type="button"
          className="btn-ghost btn-sm"
          style={{ fontSize: 11, padding: '2px 7px' }}
          title="Шпионаж на луну"
          onClick={() => onMission(11, cell.position, true)}
        >🌑🔭</button>
      )}
    </div>
  );
}

export function GalaxyScreen({ homePlanet, userId, onFleetMission }: {
  homePlanet: Planet;
  userId: string;
  onFleetMission?: (g: number, s: number, pos: number, isMoon: boolean, mission: number) => void;
}) {
  const [g, setG] = useState(homePlanet.galaxy);
  const [s, setS] = useState(homePlanet.system);

  const sys = useQuery({
    queryKey: ['galaxy', g, s],
    queryFn: () => api.get<SystemView>(`/api/galaxy/${g}/${s}`),
    refetchInterval: 10_000,
  });

  function handleMission(mission: number, pos: number, isMoon: boolean) {
    onFleetMission?.(g, s, pos, isMoon, mission);
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      {/* Nav bar */}
      <div style={{ display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap' }}>
        <h2 style={{ margin: 0, fontSize: 18, fontFamily: 'var(--ox-font)', fontWeight: 700, flex: 1 }}>
          Галактика&nbsp;
          <span style={{ fontFamily: 'var(--ox-mono)', color: 'var(--ox-accent)' }}>
            [{g}:{s}]
          </span>
        </h2>

        <div style={{ display: 'flex', gap: 6, alignItems: 'center', flexWrap: 'wrap' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
            <button type="button" className="btn-ghost btn-sm btn-icon" onClick={() => setG((v) => Math.max(1, v - 1))}>←</button>
            <span style={{ fontSize: 12, color: 'var(--ox-fg-muted)' }}>G</span>
            <input
              type="number" min={1} max={16} value={g}
              onChange={(e) => setG(clamp(Number(e.target.value), 1, 16))}
              style={{ width: 52, textAlign: 'center' }}
            />
            <button type="button" className="btn-ghost btn-sm btn-icon" onClick={() => setG((v) => Math.min(16, v + 1))}>→</button>
          </div>

          <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
            <button type="button" className="btn-ghost btn-sm btn-icon" onClick={() => setS((v) => Math.max(1, v - 1))}>←</button>
            <span style={{ fontSize: 12, color: 'var(--ox-fg-muted)' }}>S</span>
            <input
              type="number" min={1} max={999} value={s}
              onChange={(e) => setS(clamp(Number(e.target.value), 1, 999))}
              style={{ width: 68, textAlign: 'center' }}
            />
            <button type="button" className="btn-ghost btn-sm btn-icon" onClick={() => setS((v) => Math.min(999, v + 1))}>→</button>
          </div>
        </div>
      </div>

      {/* Galaxy table */}
      <div className="ox-panel" style={{ overflow: 'hidden' }}>
        {sys.isLoading && (
          <div style={{ padding: 20 }}>
            {Array.from({ length: 6 }).map((_, i) => (
              <div key={i} className="ox-skeleton" style={{ height: 36, marginBottom: 6 }} />
            ))}
          </div>
        )}

        {sys.error && (
          <div style={{ padding: 20 }}>
            <div className="ox-alert ox-alert-danger">
              Ошибка: {sys.error instanceof Error ? sys.error.message : 'неизвестная ошибка'}
            </div>
          </div>
        )}

        {sys.data && (
          <div className="ox-table-responsive">
            <table className="ox-table" style={{ margin: 0 }}>
              <thead>
                <tr>
                  <th style={{ width: 36 }}>#</th>
                  <th>Планета</th>
                  <th>Игрок</th>
                  <th>Обломки</th>
                  <th style={{ width: 120 }}>Миссии</th>
                </tr>
              </thead>
              <tbody>
                {(sys.data.cells ?? []).map((c) => {
                  const isOwn = !!c.owner_id && c.owner_id === userId;
                  return (
                    <tr
                      key={c.position}
                      style={
                        !c.has_planet
                          ? { opacity: 0.4 }
                          : isOwn
                          ? { background: 'rgba(99,217,255,0.07)' }
                          : undefined
                      }
                    >
                      <td data-label="#" className="num" style={{ fontFamily: 'var(--ox-mono)', color: 'var(--ox-fg-muted)' }}>
                        {c.position}
                      </td>
                      <td data-label="Планета">
                        {c.has_planet ? (
                          <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}>
                            {c.planet_id && (
                              <img
                                src={planetImageOf(c.position, c.planet_id, c.planet_type ?? undefined)}
                                alt=""
                                style={{ width: 24, height: 24, borderRadius: 3, objectFit: 'cover', flexShrink: 0 }}
                              />
                            )}
                            {isOwn && <span style={{ fontSize: 11 }}>🏠</span>}
                            <span style={{ fontWeight: isOwn ? 700 : 400 }}>{c.planet_name}</span>
                            {c.has_moon && (
                              <span title={c.moon_name ?? 'Луна'} style={{ fontSize: 13 }}>🌑</span>
                            )}
                          </span>
                        ) : (
                          <span style={{ color: 'var(--ox-fg-muted)' }}>—</span>
                        )}
                      </td>
                      <td data-label="Игрок">
                        {c.owner_username ? (
                          <span>
                            <span style={{ fontWeight: 600 }}>{c.owner_username}</span>
                            {c.owner_rank !== undefined && c.owner_rank !== null && (
                              <span style={{ marginLeft: 6, fontSize: 11, color: 'var(--ox-fg-muted)', fontFamily: 'var(--ox-mono)' }}>
                                #{c.owner_rank}
                              </span>
                            )}
                          </span>
                        ) : (
                          <span style={{ color: 'var(--ox-fg-muted)' }}>—</span>
                        )}
                      </td>
                      <td data-label="Обломки" className="num">
                        {c.debris_metal > 0 || c.debris_silicon > 0 ? (
                          <span style={{ color: 'var(--ox-warning)', fontFamily: 'var(--ox-mono)', fontSize: 12 }}>
                            ⛏{formatNum(c.debris_metal)} / 🔷{formatNum(c.debris_silicon)}
                          </span>
                        ) : '—'}
                      </td>
                      <td data-label="Миссии">
                        {!isOwn && (
                          <MissionButtons cell={c} onMission={handleMission} />
                        )}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
