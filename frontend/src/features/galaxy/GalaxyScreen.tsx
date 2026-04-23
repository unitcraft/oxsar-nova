import { useState, useCallback, useEffect } from 'react';
import { useQuery, useMutation } from '@tanstack/react-query';
import { api } from '@/api/client';
import { planetImageOf, formatNum, SHIPS } from '@/api/catalog';
import type { Planet } from '@/api/types';
import { useToast } from '@/ui/Toast';

interface CellView {
  position: number;
  has_planet: boolean;
  planet_name?: string;
  planet_id?: string | null;
  planet_type?: string | null;
  has_moon: boolean;
  moon_name?: string;
  moon_diameter?: number;
  moon_temp_min?: number;
  moon_temp_max?: number;
  owner_username?: string;
  owner_id?: string;
  owner_rank?: number;
  owner_last_seen?: string | null;
  owner_vacation?: boolean;
  owner_banned?: boolean;
  alliance_tag?: string | null;
  relation?: 'nap' | 'war' | 'ally' | null;
  is_friend?: boolean;
  debris_metal: number;
  debris_silicon: number;
}

interface SystemView {
  galaxy: number;
  system: number;
  cells: CellView[];
}

const GAME_SPEED = 0.75;

// Расстояние между двумя точками галактики
function galaxyDistance(
  src: { galaxy: number; system: number; position: number },
  dst: { galaxy: number; system: number; position: number },
): number {
  if (src.galaxy !== dst.galaxy) return 20000 * Math.abs(src.galaxy - dst.galaxy);
  if (src.system !== dst.system) return 2700 + 95 * Math.abs(src.system - dst.system);
  if (src.position !== dst.position) return 1000 + 5 * Math.abs(src.position - dst.position);
  return 5;
}

// Время полёта в секундах
function flightSecs(dist: number, minSpeed: number, speedPct: number): number {
  if (minSpeed <= 0) return 60;
  const raw = 10 + (3500 / speedPct) * Math.sqrt((10 * dist) / minSpeed);
  return Math.max(1, raw / GAME_SPEED);
}

function fmtDuration(secs: number): string {
  if (secs < 60) return `${Math.ceil(secs)}с`;
  const m = Math.floor(secs / 60) % 60;
  const h = Math.floor(secs / 3600) % 24;
  const d = Math.floor(secs / 86400);
  if (d > 0) return `${d}д ${h}ч ${m}м`;
  if (h > 0) return `${h}ч ${m}м`;
  return `${m}м`;
}


function clamp(v: number, lo: number, hi: number): number {
  if (Number.isNaN(v)) return lo;
  return Math.max(lo, Math.min(hi, v));
}

function formatActivity(lastSeen?: string | null): string {
  if (!lastSeen) return '';
  const mins = Math.floor((Date.now() - new Date(lastSeen).getTime()) / 60000);
  if (mins < 15) return '(*)';
  if (mins < 60) return `(${mins} min)`;
  const hrs = Math.floor(mins / 60);
  if (hrs < 24) return `(${hrs} h)`;
  return '';
}

function isInactiveDays(lastSeen: string | null | undefined, days: number): boolean {
  if (!lastSeen) return false;
  return Date.now() - new Date(lastSeen).getTime() >= days * 86400000;
}

function PlayerStatuses({ cell }: { cell: CellView }) {
  if (!cell.owner_id) return null;
  const parts: React.ReactNode[] = [];
  if (cell.owner_banned) {
    parts.push(<abbr key="b" title="Забанен" style={{ color: 'var(--ox-danger)', cursor: 'help' }}>b</abbr>);
  } else if (cell.owner_vacation) {
    parts.push(<abbr key="v" title="Режим отпуска" style={{ color: 'var(--ox-accent)', cursor: 'help' }}>v</abbr>);
  }
  if (isInactiveDays(cell.owner_last_seen, 21)) {
    parts.push(<abbr key="I" title="Очень неактивный (21+ дн)" style={{ cursor: 'help' }}>I</abbr>);
  } else if (isInactiveDays(cell.owner_last_seen, 7)) {
    parts.push(<abbr key="i" title="Неактивный (7+ дн)" style={{ cursor: 'help' }}>i</abbr>);
  }
  if (parts.length === 0) return null;
  return (
    <span style={{ fontSize: 10, fontFamily: 'var(--ox-mono)', color: 'var(--ox-fg-muted)', marginLeft: 4, letterSpacing: '0.05em' }}>
      ({parts})
    </span>
  );
}

// ---- Rocket launch inline panel ----
function RocketPanel({
  g, s, pos,
  srcPlanets,
  onClose,
}: {
  g: number;
  s: number;
  pos: number;
  srcPlanets: Planet[];
  onClose: () => void;
}) {
  const toast = useToast();
  const [srcPlanetId, setSrcPlanetId] = useState(srcPlanets[0]?.id ?? '');
  const [count, setCount] = useState(1);

  const srcPlanet = srcPlanets.find((p) => p.id === srcPlanetId);

  // Запрос количества ракет на выбранной планете
  const rockets = useQuery({
    queryKey: ['rockets', srcPlanetId],
    queryFn: () => api.get<{ count: number }>(`/api/planets/${srcPlanetId}/rockets`),
    enabled: !!srcPlanetId,
  });

  const launch = useMutation({
    mutationFn: () =>
      api.post<unknown>(`/api/planets/${srcPlanetId}/rockets/launch`, {
        dst: { galaxy: g, system: s, position: pos, is_moon: false },
        count,
        target_unit_id: 0,
      }),
    onSuccess: () => {
      toast.show('Ракеты запущены', 'success');
      onClose();
    },
    onError: (e: Error) => toast.show(e.message, 'error'),
  });

  const maxRockets = rockets.data?.count ?? 0;

  // Расстояние и время полёта ракеты (скорость ракеты ≈ 1, но прилетает мгновенно)
  const dist = srcPlanet ? galaxyDistance(
    { galaxy: srcPlanet.galaxy, system: srcPlanet.system, position: srcPlanet.position },
    { galaxy: g, system: s, position: pos },
  ) : 0;
  const flightTime = dist > 0 ? fmtDuration(flightSecs(dist, 1000, 100)) : '—';

  return (
    <div style={{ marginTop: 6, padding: '8px 10px', background: 'var(--ox-bg-panel)', border: '1px solid var(--ox-border)', borderRadius: 6, fontSize: 12 }}>
      <div style={{ fontWeight: 600, marginBottom: 6 }}>🚀 Ракетный удар [{g}:{s}:{pos}]</div>

      {srcPlanets.length > 1 && (
        <div style={{ marginBottom: 6 }}>
          <label style={{ fontSize: 11, color: 'var(--ox-fg-muted)' }}>Источник</label>
          <select value={srcPlanetId} onChange={(e) => setSrcPlanetId(e.target.value)} style={{ display: 'block', width: '100%', marginTop: 2 }}>
            {srcPlanets.map((p) => (
              <option key={p.id} value={p.id}>{p.name} [{p.galaxy}:{p.system}:{p.position}]</option>
            ))}
          </select>
        </div>
      )}

      <div style={{ display: 'flex', gap: 8, alignItems: 'center', marginBottom: 6 }}>
        <div>
          <label style={{ fontSize: 11, color: 'var(--ox-fg-muted)' }}>Количество</label>
          <input
            type="number" min={1} max={maxRockets || 1} value={count}
            onChange={(e) => setCount(clamp(Number(e.target.value), 1, maxRockets || 1))}
            style={{ display: 'block', width: 72, marginTop: 2 }}
          />
        </div>
        <div style={{ fontSize: 11, color: 'var(--ox-fg-muted)' }}>
          <div>Доступно: {rockets.isLoading ? '...' : maxRockets}</div>
          <div>Время: {flightTime}</div>
        </div>
      </div>

      <div style={{ display: 'flex', gap: 6 }}>
        <button
          type="button"
          className="btn-primary btn-sm"
          disabled={launch.isPending || count < 1 || count > maxRockets}
          onClick={() => launch.mutate()}
        >Запустить</button>
        <button type="button" className="btn-ghost btn-sm" onClick={onClose}>Отмена</button>
      </div>
    </div>
  );
}

// ---- Mission buttons ----
function MissionButtons({
  cell, g, s,
  srcPlanet,
  srcPlanets,
  onMission,
}: {
  cell: CellView;
  g: number;
  s: number;
  srcPlanet: Planet;
  srcPlanets: Planet[];
  onMission: (mission: number, position: number, isMoon: boolean) => void;
}) {
  const [showRockets, setShowRockets] = useState(false);

  if (!cell.has_planet) return null;

  // Рассчитать время полёта/расход водорода для подсказок (берём самый медленный корабль)
  const dist = galaxyDistance(
    { galaxy: srcPlanet.galaxy, system: srcPlanet.system, position: srcPlanet.position },
    { galaxy: g, system: s, position: cell.position },
  );
  const minSpeed = Math.min(...SHIPS.filter((s) => s.fuel !== undefined).map((s) => s.speed ?? Infinity).filter(isFinite));
  const flightTime = minSpeed > 0 ? fmtDuration(flightSecs(dist, minSpeed, 100)) : '—';
  const fuelHint = `Расстояние: ${dist}\nВремя (мин. скорость): ${flightTime}`;

  return (
    <div>
      <div style={{ display: 'flex', gap: 4, flexWrap: 'wrap' }}>
        <button
          type="button"
          className="btn-ghost btn-sm"
          style={{ fontSize: 11, padding: '2px 7px' }}
          title={`Шпионаж\n${fuelHint}`}
          onClick={() => onMission(11, cell.position, false)}
        >🔭</button>
        <button
          type="button"
          className="btn-ghost btn-sm"
          style={{ fontSize: 11, padding: '2px 7px' }}
          title={`Атака\n${fuelHint}`}
          onClick={() => onMission(10, cell.position, false)}
        >⚔️</button>
        <button
          type="button"
          className="btn-ghost btn-sm"
          style={{ fontSize: 11, padding: '2px 7px' }}
          title={`Транспорт\n${fuelHint}`}
          onClick={() => onMission(7, cell.position, false)}
        >📦</button>
        {(cell.debris_metal > 0 || cell.debris_silicon > 0) && (
          <button
            type="button"
            className="btn-ghost btn-sm"
            style={{ fontSize: 11, padding: '2px 7px' }}
            title={`Переработка обломков\n${fuelHint}`}
            onClick={() => onMission(9, cell.position, false)}
          >♻️</button>
        )}
        {cell.has_moon && (
          <button
            type="button"
            className="btn-ghost btn-sm"
            style={{ fontSize: 11, padding: '2px 7px' }}
            title={`Шпионаж на луну\n${fuelHint}`}
            onClick={() => onMission(11, cell.position, true)}
          >🌑🔭</button>
        )}
        <button
          type="button"
          className="btn-ghost btn-sm"
          style={{ fontSize: 11, padding: '2px 7px', color: showRockets ? 'var(--ox-danger)' : undefined }}
          title="Ракетный удар"
          onClick={() => setShowRockets((v) => !v)}
        >🚀</button>
      </div>
      {showRockets && (
        <RocketPanel
          g={g} s={s} pos={cell.position}
          srcPlanets={srcPlanets}
          onClose={() => setShowRockets(false)}
        />
      )}
    </div>
  );
}

// ---- Star Surveillance — закладки систем в localStorage ----
function useSurveillance() {
  const key = 'galaxy_watch';
  const load = (): string[] => {
    try { return JSON.parse(localStorage.getItem(key) ?? '[]') as string[]; }
    catch { return []; }
  };
  const save = (list: string[]) => localStorage.setItem(key, JSON.stringify(list));

  const isWatching = useCallback((g: number, s: number): boolean => load().includes(`${g}:${s}`), []);
  const toggle = useCallback((g: number, s: number) => {
    const coord = `${g}:${s}`;
    const list = load();
    const next = list.includes(coord) ? list.filter((x) => x !== coord) : [...list, coord];
    save(next);
    return next.includes(coord);
  }, []);
  const watched = useCallback((): Array<{ g: number; s: number }> =>
    load().map((x) => { const [a, b] = x.split(':'); return { g: Number(a), s: Number(b) }; }),
  []);

  return { isWatching, toggle, watched };
}

export function GalaxyScreen({ homePlanet, userId, onFleetMission, planets, initialCoords }: {
  homePlanet: Planet;
  userId: string;
  planets?: Planet[];
  onFleetMission?: (g: number, s: number, pos: number, isMoon: boolean, mission: number) => void;
  initialCoords?: { galaxy: number; system: number } | null;
}) {
  const [g, setG] = useState(initialCoords?.galaxy ?? homePlanet.galaxy);
  const [s, setS] = useState(initialCoords?.system ?? homePlanet.system);

  useEffect(() => {
    if (initialCoords) {
      setG(initialCoords.galaxy);
      setS(initialCoords.system);
    }
  }, [initialCoords?.galaxy, initialCoords?.system]);
  const [watchLabel, setWatchLabel] = useState<string | null>(null);
  const surveillance = useSurveillance();

  const sys = useQuery({
    queryKey: ['galaxy', g, s],
    queryFn: () => api.get<SystemView>(`/api/galaxy/${g}/${s}`),
    refetchInterval: 10_000,
  });

  function handleMission(mission: number, pos: number, isMoon: boolean) {
    onFleetMission?.(g, s, pos, isMoon, mission);
  }

  function handleWatch() {
    const watching = surveillance.toggle(g, s);
    setWatchLabel(watching ? 'Система добавлена в наблюдение' : 'Система удалена из наблюдения');
    setTimeout(() => setWatchLabel(null), 2000);
  }

  const isWatching = surveillance.isWatching(g, s);
  const watched = surveillance.watched();

  // Текущая планета игрока как источник для расчётов (первая в списке planets или homePlanet)
  const srcPlanets = planets ?? [homePlanet];
  const srcPlanet = srcPlanets.find((p) => !p.is_moon) ?? homePlanet;

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

          {/* Star Surveillance toggle */}
          <button
            type="button"
            className="btn-ghost btn-sm"
            title={isWatching ? 'Убрать из наблюдения' : 'Добавить в наблюдение'}
            style={{ fontSize: 13, color: isWatching ? 'var(--ox-accent)' : 'var(--ox-fg-muted)' }}
            onClick={handleWatch}
          >{isWatching ? '👁‍🗨' : '👁'}</button>
        </div>
      </div>

      {/* Surveillance feedback */}
      {watchLabel && (
        <div style={{ fontSize: 12, color: 'var(--ox-accent)', fontFamily: 'var(--ox-mono)' }}>
          {watchLabel}
        </div>
      )}

      {/* Watched systems */}
      {watched.length > 0 && (
        <div style={{ display: 'flex', gap: 6, flexWrap: 'wrap', alignItems: 'center' }}>
          <span style={{ fontSize: 11, color: 'var(--ox-fg-muted)' }}>Наблюдение:</span>
          {watched.map(({ g: wg, s: ws }) => (
            <button
              key={`${wg}:${ws}`}
              type="button"
              className="btn-ghost btn-sm"
              style={{ fontSize: 11, fontFamily: 'var(--ox-mono)', padding: '2px 7px', color: wg === g && ws === s ? 'var(--ox-accent)' : undefined }}
              onClick={() => { setG(wg); setS(ws); }}
            >[{wg}:{ws}]</button>
          ))}
        </div>
      )}

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
                  <th>Альянс</th>
                  <th>Обломки</th>
                  <th style={{ width: 160 }}>Миссии</th>
                </tr>
              </thead>
              <tbody>
                {(sys.data.cells ?? []).map((c) => {
                  const isOwn = !!c.owner_id && c.owner_id === userId;
                  const moonTitle = c.has_moon
                    ? [
                        c.moon_name ?? 'Луна',
                        c.moon_diameter ? `${c.moon_diameter} км` : '',
                        c.moon_temp_min != null && c.moon_temp_max != null
                          ? `${c.moon_temp_min}..${c.moon_temp_max}°C`
                          : '',
                      ].filter(Boolean).join(' | ')
                    : '';
                  const debrisTitle = (c.debris_metal > 0 || c.debris_silicon > 0)
                    ? `Обломки\nМеталл: ${c.debris_metal.toLocaleString('ru-RU')}\nКремний: ${c.debris_silicon.toLocaleString('ru-RU')}`
                    : '';
                  const activity = formatActivity(c.owner_last_seen);
                  return (
                    <tr
                      key={c.position}
                      style={
                        c.position === 16
                          ? { background: 'rgba(139,92,246,0.07)' }
                          : !c.has_planet
                          ? { opacity: 0.4 }
                          : isOwn
                          ? { background: 'rgba(99,217,255,0.07)' }
                          : c.relation === 'ally'
                          ? { background: 'rgba(34,197,94,0.08)' }
                          : c.relation === 'war'
                          ? { background: 'rgba(239,68,68,0.10)' }
                          : c.relation === 'nap'
                          ? { background: 'rgba(245,158,11,0.08)' }
                          : undefined
                      }
                    >
                      <td data-label="#" className="num" style={{ fontFamily: 'var(--ox-mono)', color: c.position === 16 ? 'var(--ox-accent)' : 'var(--ox-fg-muted)' }}>
                        {c.position}
                      </td>
                      <td data-label="Планета">
                        {c.position === 16 ? (
                          <span style={{ color: 'var(--ox-accent)', fontStyle: 'italic', fontSize: 13 }}>🌌 Бесконечные дали</span>
                        ) : c.has_planet ? (
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
                              <span title={moonTitle} style={{ fontSize: 13, cursor: moonTitle ? 'help' : undefined }}>🌑</span>
                            )}
                          </span>
                        ) : (
                          <span style={{ color: 'var(--ox-fg-muted)' }}>—</span>
                        )}
                      </td>
                      <td data-label="Игрок">
                        {c.owner_username ? (
                          <span>
                            {c.is_friend && <span style={{ marginRight: 4, color: 'var(--ox-accent)' }} title="В друзьях">⭐</span>}
                            <span style={{ fontWeight: 600 }}>{c.owner_username}</span>
                            <PlayerStatuses cell={c} />
                            {c.owner_rank !== undefined && c.owner_rank !== null && (
                              <span style={{ marginLeft: 6, fontSize: 11, color: 'var(--ox-fg-muted)', fontFamily: 'var(--ox-mono)' }}>
                                #{c.owner_rank}
                              </span>
                            )}
                            {activity && (
                              <span style={{ marginLeft: 6, fontSize: 10, color: 'var(--ox-fg-muted)', fontFamily: 'var(--ox-mono)' }}>
                                {activity}
                              </span>
                            )}
                          </span>
                        ) : c.position !== 16 ? (
                          <button
                            type="button"
                            className="btn-ghost btn-sm"
                            style={{ fontSize: 11, padding: '2px 7px', color: 'var(--ox-fg-muted)' }}
                            title="Отправить экспедицию"
                            onClick={() => onFleetMission?.(g, s, c.position, false, 15)}
                          >🌌 Экспедиция</button>
                        ) : null}
                      </td>
                      <td data-label="Альянс">
                        {c.alliance_tag
                          ? <span style={{ fontFamily: 'var(--ox-mono)', fontSize: 12, color: 'var(--ox-accent)' }}>[{c.alliance_tag}]</span>
                          : <span style={{ color: 'var(--ox-fg-muted)' }}>—</span>
                        }
                      </td>
                      <td data-label="Обломки" className="num">
                        {c.debris_metal > 0 || c.debris_silicon > 0 ? (
                          <span
                            title={debrisTitle}
                            style={{ color: 'var(--ox-warning)', fontFamily: 'var(--ox-mono)', fontSize: 12, cursor: 'help' }}
                          >
                            🟠{formatNum(c.debris_metal)} / 💎{formatNum(c.debris_silicon)}
                          </span>
                        ) : '—'}
                      </td>
                      <td data-label="Миссии">
                        {c.position === 16 && (
                          <button
                            type="button"
                            className="btn-ghost btn-sm"
                            style={{ fontSize: 11, padding: '2px 7px', color: 'var(--ox-accent)' }}
                            title="Отправить экспедицию в Бесконечные дали"
                            onClick={() => onFleetMission?.(g, s, 16, false, 15)}
                          >🌌 Экспедиция</button>
                        )}
                        {c.position !== 16 && !isOwn && c.has_planet && (
                          <MissionButtons
                            cell={c}
                            g={g}
                            s={s}
                            srcPlanet={srcPlanet}
                            srcPlanets={srcPlanets.filter((p) => !p.is_moon)}
                            onMission={handleMission}
                          />
                        )}
                        {c.position !== 16 && isOwn && c.has_planet && (
                          <button
                            type="button"
                            className="btn-ghost btn-sm"
                            style={{ fontSize: 11, padding: '2px 7px', color: 'var(--ox-fg-muted)' }}
                            title="Отправить экспедицию с этой планеты"
                            onClick={() => onFleetMission?.(g, s, c.position, false, 15)}
                          >🌌</button>
                        )}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
              <tfoot>
                <tr>
                  <td colSpan={6} style={{ fontSize: 10, color: 'var(--ox-fg-muted)', padding: '8px 12px', fontFamily: 'var(--ox-mono)', borderTop: '1px solid var(--ox-border)' }}>
                    <b>(*)</b> только что&nbsp;&nbsp;
                    <b>i</b> неактивный (7+ дн)&nbsp;&nbsp;
                    <b>I</b> очень неактивный (21+ дн)&nbsp;&nbsp;
                    <b style={{ color: 'var(--ox-danger)' }}>b</b> забанен&nbsp;&nbsp;
                    <b style={{ color: 'var(--ox-accent)' }}>v</b> отпуск&nbsp;&nbsp;
                    <b>🚀</b> ракетный удар&nbsp;&nbsp;
                    <b>🌌</b> экспедиция&nbsp;&nbsp;
                    <b>👁</b> наблюдение
                    <br />
                    <span style={{ display: 'inline-block', width: 10, height: 10, background: 'rgba(34,197,94,0.8)', marginRight: 4, verticalAlign: 'middle' }} />союзник&nbsp;&nbsp;
                    <span style={{ display: 'inline-block', width: 10, height: 10, background: 'rgba(245,158,11,0.8)', marginRight: 4, verticalAlign: 'middle' }} />НЕН&nbsp;&nbsp;
                    <span style={{ display: 'inline-block', width: 10, height: 10, background: 'rgba(239,68,68,0.8)', marginRight: 4, verticalAlign: 'middle' }} />война
                  </td>
                </tr>
              </tfoot>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
