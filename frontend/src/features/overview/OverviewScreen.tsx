import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useEffect, useState } from 'react';
import { api } from '@/api/client';
import { buildingName, imageOfId, nameOf, planetImageOf, planetImageSize } from '@/api/catalog';
import type { Planet, QueueItem, ShipyardQueueItem } from '@/api/types';
import { Countdown } from '@/ui/Countdown';
import { ProgressBar } from '@/ui/ProgressBar';
import { ResourceTicker } from '@/ui/ResourceTicker';
import { useToast } from '@/ui/Toast';

interface FleetRow {
  id: string;
  dst_galaxy: number;
  dst_system: number;
  dst_position: number;
  dst_is_moon: boolean;
  mission: number;
  state: string;
  depart_at: string;
  arrive_at: string;
  return_at?: string | null;
  ships?: Record<string, number>;
  carry?: { metal: number; silicon: number; hydrogen: number };
}

interface MyRank {
  rank: number;
  type: string;
  points: number;
  e_points?: number;
}

const MISSION_LABELS: Record<number, string> = {
  7: 'Транспорт', 8: 'Колонизация', 9: 'Переработка',
  10: 'Атака', 11: 'Шпионаж', 15: 'Экспедиция',
};

const MISSION_ICONS: Record<number, string> = {
  7: '📦', 8: '🌍', 9: '♻️', 10: '⚔️', 11: '🔭', 15: '🚀',
};

function formatCoords(p: Planet) {
  return `[${p.galaxy}:${p.system}:${p.position}]`;
}


export function OverviewScreen() {
  const planets = useQuery({
    queryKey: ['planets'],
    queryFn: () => api.get<{ planets: Planet[] }>('/api/planets'),
    refetchInterval: 30000,
  });

  const unread = useQuery({
    queryKey: ['messages-unread'],
    queryFn: () => api.get<{ count: number }>('/api/messages/unread-count'),
    refetchInterval: 60000,
  });

  const me = useQuery({
    queryKey: ['highscore-me'],
    queryFn: () => api.get<MyRank>('/api/highscore/me'),
    refetchInterval: 60000,
  });

  const fleets = useQuery({
    queryKey: ['fleets'],
    queryFn: () => api.get<{ fleets: FleetRow[] }>('/api/fleet'),
    refetchInterval: 10000,
  });

  const list = planets.data?.planets ?? [];
  const [selectedPlanetId, setSelectedPlanetId] = useState<string | null>(null);
  const activeFleets = (fleets.data?.fleets ?? []).filter((f) => {
    if (f.state === 'done') return false;
    const finishAt = f.state === 'returning' && f.return_at ? f.return_at : f.arrive_at;
    return new Date(finishAt).getTime() > Date.now();
  });
  const unreadCount = unread.data?.count ?? 0;
  const selectedPlanet = list.find((p) => p.id === selectedPlanetId) ?? list[0];

  if (planets.isLoading) {
    return (
      <div style={{ padding: 24 }}>
        <div className="ox-skeleton" style={{ height: 80, marginBottom: 12 }} />
        <div className="ox-skeleton" style={{ height: 80, marginBottom: 12 }} />
        <div className="ox-skeleton" style={{ height: 200 }} />
      </div>
    );
  }

  if (!list.length) {
    return (
      <div style={{ padding: 24 }}>
        <div className="ox-alert ox-alert-warning">Нет планет. Попробуйте перезагрузить страницу.</div>
      </div>
    );
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20, padding: '16px 0' }}>

      {/* Уведомление о сообщениях */}
      {unreadCount > 0 && (
        <div style={{
          padding: '10px 16px',
          background: 'rgba(99,217,255,0.08)',
          border: '1px solid rgba(99,217,255,0.3)',
          borderRadius: 8,
          fontSize: 13,
          color: 'var(--ox-accent)',
          display: 'flex',
          alignItems: 'center',
          gap: 8,
        }}>
          ✉️ У вас <b>{unreadCount}</b> {unreadCount === 1 ? 'непрочитанное сообщение' : 'непрочитанных сообщений'}
        </div>
      )}

      {/* Статистика игрока */}
      {me.data && (
        <div className="ox-panel" style={{ padding: '12px 20px', display: 'flex', gap: 32, flexWrap: 'wrap', alignItems: 'center' }}>
          <StatItem label="Очки" value={Math.floor(me.data.points).toLocaleString('ru-RU')} />
          <StatItem label="Место в рейтинге" value={`#${me.data.rank}`} accent />
          {(me.data.e_points ?? 0) > 0 && (
            <StatItem label="Боевой опыт" value={Math.floor(me.data.e_points!).toLocaleString('ru-RU')} />
          )}
        </div>
      )}

      {/* Флоты в движении */}
      {activeFleets.length > 0 && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          <div style={{ fontSize: 11, fontWeight: 700, letterSpacing: '0.08em', textTransform: 'uppercase', color: 'var(--ox-fg-dim)', paddingLeft: 2 }}>
            Флоты в пути
          </div>
          {activeFleets.map((f) => (
            <FleetEventRow key={f.id} fleet={f} />
          ))}
        </div>
      )}

      {/* Карусель планет */}
      {list.length > 1 && (
        <div style={{ display: 'flex', gap: 10, overflowX: 'auto', paddingBottom: 4 }}>
          {list.map((p) => {
            const active = p.id === selectedPlanet?.id;
            return (
              <button
                key={p.id}
                type="button"
                className="planet-card"
                onClick={() => setSelectedPlanetId(p.id)}
                style={{
                  display: 'flex', flexDirection: 'column', alignItems: 'center',
                  minWidth: 80, padding: '8px 6px', borderRadius: 8, cursor: 'pointer',
                  border: `1px solid ${active ? 'var(--ox-accent)' : 'var(--ox-border)'}`,
                  background: active ? 'rgba(99,217,255,0.08)' : 'var(--ox-bg-card)',
                  flexShrink: 0, height: 110,
                }}
              >
                <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                  <img
                    src={p.is_moon ? '/images/planets/mond.jpg' : planetImageOf(p.position, p.id, p.planet_type)}
                    alt=""
                    style={{ width: planetImageSize(p.diameter), height: planetImageSize(p.diameter), borderRadius: 5, objectFit: 'cover' }}
                  />
                </div>
                <span style={{ fontSize: 11, fontWeight: 600, color: active ? 'var(--ox-accent)' : 'var(--ox-fg)', textAlign: 'center', maxWidth: 76, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', marginTop: 4 }}>
                  {p.name}
                </span>
                <span style={{ fontSize: 10, color: 'var(--ox-fg-muted)', fontFamily: 'var(--ox-mono)' }}>
                  [{p.galaxy}:{p.system}:{p.position}]
                </span>
              </button>
            );
          })}
        </div>
      )}

      {/* Карточка выбранной планеты */}
      {selectedPlanet && <PlanetOverviewCard key={selectedPlanet.id} planet={selectedPlanet} />}
    </div>
  );
}

function StatItem({ label, value, accent }: { label: string; value: string; accent?: boolean }) {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
      <div style={{ fontSize: 10, color: 'var(--ox-fg-muted)', textTransform: 'uppercase', letterSpacing: '0.06em' }}>{label}</div>
      <div style={{ fontSize: 16, fontWeight: 700, fontFamily: 'var(--ox-mono)', color: accent ? 'var(--ox-accent)' : undefined }}>
        {value}
      </div>
    </div>
  );
}

function FleetEventRow({ fleet: f }: { fleet: FleetRow }) {
  const qc = useQueryClient();
  const toast = useToast();

  const recall = useMutation({
    mutationFn: (id: string) => api.post<unknown>(`/api/fleet/${id}/recall`),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['fleets'] });
      toast.show('success', 'Флот отозван', '');
    },
    onError: () => toast.show('danger', 'Ошибка', 'Не удалось отозвать флот'),
  });

  const isOutbound = f.state === 'outbound';
  const isReturning = f.state === 'returning';
  const finishAt = isReturning && f.return_at ? f.return_at : f.arrive_at;

  const total = new Date(finishAt).getTime() - new Date(f.depart_at).getTime();
  const elapsed = Date.now() - new Date(f.depart_at).getTime();
  const pct = total > 0 ? Math.min(100, (elapsed / total) * 100) : 100;

  const icon = MISSION_ICONS[f.mission] ?? '🚀';
  const label = MISSION_LABELS[f.mission] ?? `#${f.mission}`;
  const target = `[${f.dst_galaxy}:${f.dst_system}:${f.dst_position}]${f.dst_is_moon ? ' 🌑' : ''}`;
  const stateLabel = isReturning ? '← Возврат' : '→ В пути';

  return (
    <div className="ox-panel" style={{ padding: '10px 16px' }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 6 }}>
        <span style={{ fontSize: 18, flexShrink: 0 }}>{icon}</span>
        <div style={{ flex: 1, minWidth: 0 }}>
          <div style={{ fontSize: 13, fontWeight: 600 }}>
            {label}
            <span style={{ marginLeft: 8, fontSize: 12, color: 'var(--ox-fg-muted)', fontFamily: 'var(--ox-mono)' }}>
              {target}
            </span>
          </div>
          <div style={{ fontSize: 11, color: isReturning ? 'var(--ox-success)' : 'var(--ox-accent)' }}>
            {stateLabel}
          </div>
        </div>
        <span style={{ fontSize: 13, fontFamily: 'var(--ox-mono)', color: 'var(--ox-fg-dim)', flexShrink: 0 }}>
          <Countdown finishAt={finishAt} />
        </span>
        {isOutbound && (
          <button
            type="button"
            className="btn"
            onClick={() => recall.mutate(f.id)}
            disabled={recall.isPending}
            style={{ fontSize: 11, padding: '3px 10px', flexShrink: 0 }}
          >
            Отозвать
          </button>
        )}
      </div>
      <ProgressBar pct={pct} variant={isReturning ? 'success' : 'default'} height={3} />
      {f.ships && Object.keys(f.ships).length > 0 && (
        <div style={{ marginTop: 6, display: 'flex', flexWrap: 'wrap', gap: '2px 10px' }}>
          {Object.entries(f.ships).map(([unitId, count]) => {
            const id = Number(unitId);
            const img = imageOfId(id);
            return (
              <span key={unitId} style={{ display: 'inline-flex', alignItems: 'center', gap: 3, fontSize: 11, color: 'var(--ox-fg-dim)', fontFamily: 'var(--ox-mono)' }}>
                {img && <img src={img} alt="" width={16} height={16} style={{ imageRendering: 'pixelated', opacity: 0.85 }} />}
                {nameOf(id)} ×{count}
              </span>
            );
          })}
          {f.carry && (f.carry.metal > 0 || f.carry.silicon > 0 || f.carry.hydrogen > 0) && (
            <span style={{ fontSize: 11, color: 'var(--ox-fg-muted)', marginLeft: 4 }}>
              [{f.carry.metal > 0 && `⛏${f.carry.metal.toLocaleString('ru-RU')}`}{f.carry.silicon > 0 && ` 🔷${f.carry.silicon.toLocaleString('ru-RU')}`}{f.carry.hydrogen > 0 && ` 💧${f.carry.hydrogen.toLocaleString('ru-RU')}`}]
            </span>
          )}
        </div>
      )}
    </div>
  );
}

function useNow(intervalMs = 1000) {
  const [now, setNow] = useState(() => Date.now());
  useEffect(() => {
    const t = setInterval(() => setNow(Date.now()), intervalMs);
    return () => clearInterval(t);
  }, [intervalMs]);
  return now;
}

function PlanetOverviewCard({ planet }: { planet: Planet & { diameter?: number; used_fields?: number; temp_min?: number; temp_max?: number } }) {
  const now = useNow();
  const qc = useQueryClient();
  const invalidateQueues = () => {
    void qc.invalidateQueries({ queryKey: ['buildings-queue', planet.id] });
    void qc.invalidateQueries({ queryKey: ['shipyard-queue', planet.id] });
  };
  const bQueue = useQuery({
    queryKey: ['buildings-queue', planet.id],
    queryFn: () => api.get<{ queue: QueueItem[] }>(`/api/planets/${planet.id}/buildings/queue`),
    refetchInterval: 5000,
  });
  const sQueue = useQuery({
    queryKey: ['shipyard-queue', planet.id],
    queryFn: () => api.get<{ queue: ShipyardQueueItem[] }>(`/api/planets/${planet.id}/shipyard/queue`),
    refetchInterval: 5000,
  });

  const bItems = (bQueue.data?.queue ?? []).filter((i) => new Date(i.end_at).getTime() > now);
  const sItems = (sQueue.data?.queue ?? []).filter((i) => new Date(i.end_at).getTime() > now);
  const hasActivity = bItems.length > 0 || sItems.length > 0;

  const diameter = planet.diameter;
  const usedFields = planet.used_fields ?? 0;
  const maxF = diameter ? Math.floor(diameter / 1000) * 5 : null;
  const tempMin = planet.temp_min;
  const tempMax = planet.temp_max;

  return (
    <div className="ox-panel" style={{ overflow: 'hidden' }}>
      {/* Заголовок планеты */}
      <div style={{
        display: 'flex', alignItems: 'center', gap: 12,
        padding: '14px 20px', borderBottom: '1px solid var(--ox-border)',
        background: 'linear-gradient(135deg, rgba(99,217,255,0.04) 0%, transparent 60%)',
      }}>
        <img
          src={planet.is_moon ? '/images/planets/mond.jpg' : planetImageOf(planet.position, planet.id, planet.planet_type)}
          alt=""
          style={{ width: planetImageSize(planet.diameter), height: planetImageSize(planet.diameter), borderRadius: 6, objectFit: 'cover', flexShrink: 0 }}
        />
        <div style={{ flex: 1, minWidth: 0 }}>
          <div style={{ fontWeight: 700, fontSize: 16, fontFamily: 'var(--ox-font)' }}>
            {planet.name}
            {planet.is_moon && <span style={{ marginLeft: 6, fontSize: 11, color: 'var(--ox-fg-dim)', fontWeight: 400 }}>луна</span>}
          </div>
          <div style={{ fontSize: 12, color: 'var(--ox-fg-muted)', fontFamily: 'var(--ox-mono)' }}>
            {formatCoords(planet)}
          </div>
        </div>
        {hasActivity && (
          <span className="ox-badge ox-badge-accent" style={{ fontSize: 10 }}>активность</span>
        )}
      </div>

      {/* Характеристики планеты */}
      {(diameter || tempMin !== undefined) && (
        <div style={{
          display: 'flex', gap: 0, flexWrap: 'wrap',
          borderBottom: '1px solid var(--ox-border)',
          background: 'rgba(255,255,255,0.02)',
        }}>
          {diameter && (
            <PlanetStat icon="📐" label="Диаметр" value={`${diameter.toLocaleString('ru-RU')} км`} />
          )}
          {maxF !== null && (
            <PlanetStat icon="🔲" label="Поля" value={`${usedFields} / ${maxF}`} />
          )}
          {tempMin !== undefined && tempMax !== undefined && (
            <PlanetStat icon="🌡️" label="Температура" value={`${tempMin}°C … ${tempMax}°C`} />
          )}
        </div>
      )}

      {/* Ресурсы */}
      <div style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(auto-fit, minmax(130px, 1fr))',
        gap: 1,
        background: 'var(--ox-border)',
        borderBottom: hasActivity ? '1px solid var(--ox-border)' : undefined,
      }}>
        <ResourceCell icon="⛏" label="Металл" value={planet.metal} ratePerSec={0} />
        <ResourceCell icon="🔷" label="Кремний" value={planet.silicon} ratePerSec={0} />
        <ResourceCell icon="💧" label="Водород" value={planet.hydrogen} ratePerSec={0} />
      </div>

      {/* Очереди */}
      {hasActivity && (
        <div style={{ padding: '12px 20px', display: 'flex', flexDirection: 'column', gap: 8 }}>
          {bItems.map((item) => (
            <ActiveQueueItem
              key={item.id}
              icon="🏗"
              label={`${buildingName(item.unit_id)} → ур. ${item.target_level}`}
              startAt={item.start_at}
              endAt={item.end_at}
              onDone={invalidateQueues}
            />
          ))}
          {sItems.map((item) => (
            <ActiveQueueItem
              key={item.id}
              icon="🚀"
              label={`${nameOf(item.unit_id)} × ${item.count}`}
              startAt={item.start_at}
              endAt={item.end_at}
              onDone={invalidateQueues}
            />
          ))}
        </div>
      )}
    </div>
  );
}

function PlanetStat({ icon, label, value }: { icon: string; label: string; value: string }) {
  return (
    <div style={{
      padding: '7px 16px',
      display: 'flex', alignItems: 'center', gap: 6,
      fontSize: 12, color: 'var(--ox-fg-dim)',
      borderRight: '1px solid var(--ox-border)',
    }}>
      <span>{icon}</span>
      <span style={{ color: 'var(--ox-fg-dim)', marginRight: 2 }}>{label}:</span>
      <span style={{ fontFamily: 'var(--ox-mono)', color: 'var(--ox-fg)' }}>{value}</span>
    </div>
  );
}

function ResourceCell({
  icon, label, value, ratePerSec,
}: {
  icon: string;
  label: string;
  value: number;
  ratePerSec: number;
}) {
  return (
    <div style={{
      background: 'var(--ox-bg-panel)',
      padding: '10px 14px',
      display: 'flex', alignItems: 'center', gap: 8,
    }}>
      <span style={{ fontSize: 16 }}>{icon}</span>
      <div style={{ flex: 1, minWidth: 0 }}>
        <div style={{ fontSize: 10, color: 'var(--ox-fg-muted)', textTransform: 'uppercase', letterSpacing: '0.06em' }}>
          {label}
        </div>
        <ResourceTicker value={value} ratePerSec={ratePerSec} />
      </div>
    </div>
  );
}

function ActiveQueueItem({
  icon, label, startAt, endAt, onDone,
}: {
  icon: string;
  label: string;
  startAt: string;
  endAt: string;
  onDone?: () => void;
}) {
  const total = new Date(endAt).getTime() - new Date(startAt).getTime();
  const elapsed = Date.now() - new Date(startAt).getTime();
  const pct = total > 0 ? Math.min(100, (elapsed / total) * 100) : 100;

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 8, fontSize: 13 }}>
        <span>{icon}</span>
        <span style={{ flex: 1, fontWeight: 600 }}>{label}</span>
        <Countdown finishAt={endAt} {...(onDone ? { onDone } : {})} />
      </div>
      <ProgressBar pct={pct} variant="default" height={4} />
    </div>
  );
}
