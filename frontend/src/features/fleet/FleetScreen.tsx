import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';
import { SHIPS, nameOf } from '@/api/catalog';
import { useTranslation } from '@/i18n/i18n';
import type { Planet } from '@/api/types';

// Простой Fleet UI — пока только TRANSPORT (mission=7).
// Остальные миссии (ATTACK/SPY/COLONIZE) подключатся в M4/M5.
// Layout reference: oxsar2/www/templates/standard/missions.tpl +
// missions2/3.tpl (пошаговый wizard). В v1 делаем однострочно:
// один экран — один submit.

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
}

export function FleetScreen({ planet }: { planet: Planet }) {
  const { t, tf } = useTranslation();
  const qc = useQueryClient();
  const fleets = useQuery({
    queryKey: ['fleets'],
    queryFn: () => api.get<{ fleets: FleetRow[] | null }>('/api/fleet'),
    refetchInterval: 3000,
  });

  const [g, setG] = useState(planet.galaxy);
  const [s, setS] = useState(planet.system);
  const [pos, setPos] = useState(planet.position);
  const [isMoon, setIsMoon] = useState(false);
  const [speed, setSpeed] = useState(100);
  const [metal, setMetal] = useState(0);
  const [silicon, setSilicon] = useState(0);
  const [hydrogen, setHydrogen] = useState(0);
  const [ships, setShips] = useState<Record<number, number>>({});
  // mission: 7=TRANSPORT, 9=RECYCLING, 10=ATTACK. Carry-поля имеют
  // смысл только для TRANSPORT, для остальных миссий клиентский
  // форм-стейт игнорируется (send.mutate обнуляет их ниже).
  const [mission, setMission] = useState(7);
  const [colonyName, setColonyName] = useState('');

  const send = useMutation({
    mutationFn: () => {
      // Carry имеет смысл только для TRANSPORT (mission=7).
      // Для ATTACK/RECYCLING насильно обнуляем, чтобы не возить
      // «туда» ресурсы случайно (легко забыть сбросить поля).
      // Carry имеет смысл для TRANSPORT (mission=7) и COLONIZE
      // (mission=8 — стартовые ресурсы новой планеты).
      const carryAllowed = mission === 7 || mission === 8;
      const carryM = carryAllowed ? metal : 0;
      const carryS = carryAllowed ? silicon : 0;
      const carryH = carryAllowed ? hydrogen : 0;
      return api.post<unknown>('/api/fleet', {
        src_planet_id: planet.id,
        dst: { galaxy: g, system: s, position: pos, is_moon: isMoon },
        ships: Object.fromEntries(
          Object.entries(ships).filter(([, n]) => Number(n) > 0),
        ),
        carry_metal: carryM,
        carry_silicon: carryS,
        carry_hydrogen: carryH,
        speed_percent: speed,
        mission,
        colony_name: mission === 8 ? colonyName : undefined,
      });
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['fleets'] });
      void qc.invalidateQueries({ queryKey: ['planets'] });
      setShips({});
      setMetal(0);
      setSilicon(0);
      setHydrogen(0);
    },
  });

  const recall = useMutation({
    mutationFn: (id: string) => api.post<unknown>(`/api/fleet/${id}/recall`),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['fleets'] });
    },
  });

  const list = fleets.data?.fleets ?? [];

  return (
    <section>
      <h2>{t('global', 'MENU_FLEET')}</h2>

      <h3>{tf('Main', 'MISSION', 'Миссия')}</h3>
      <div style={{ marginBottom: 12 }}>
        <select value={mission} onChange={(e) => setMission(Number(e.target.value))}>
          <option value={7}>{tf('Main', 'MISSION_TRANSPORT', '7 — Транспорт')}</option>
          <option value={10}>{tf('Main', 'MISSION_ATTACK', '10 — Атака')}</option>
          <option value={9}>{tf('Main', 'MISSION_RECYCLING', '9 — Переработка')}</option>
          <option value={11}>{tf('Main', 'MISSION_SPY', '11 — Шпионаж')}</option>
          <option value={8}>{tf('Main', 'MISSION_COLONIZE', '8 — Колонизация')}</option>
          <option value={15}>{tf('Main', 'MISSION_EXPEDITION', '15 — Экспедиция')}</option>
        </select>
      </div>

      <h3>{t('Main', 'POSITION')}</h3>
      <div style={{ display: 'flex', gap: 8, marginBottom: 12, alignItems: 'center' }}>
        G&nbsp;<input type="number" min={1} max={16} value={g} onChange={(e) => setG(Number(e.target.value))} style={{ width: 60 }} />
        S&nbsp;<input type="number" min={1} max={999} value={s} onChange={(e) => setS(Number(e.target.value))} style={{ width: 80 }} />
        P&nbsp;<input type="number" min={1} max={15} value={pos} onChange={(e) => setPos(Number(e.target.value))} style={{ width: 60 }} />
        <label>
          <input type="checkbox" checked={isMoon} onChange={(e) => setIsMoon(e.target.checked)} />
          &nbsp;{tf('Main', 'MOON', 'Луна')}
        </label>
      </div>

      <h3>{t('global', 'MENU_SHIPYARD')}</h3>
      <table className="ox-table">
        <thead>
          <tr>
            <th>{t('global', 'MENU_SHIPYARD')}</th>
            <th>{t('Main', 'POSITION')}</th>
          </tr>
        </thead>
        <tbody>
          {SHIPS.map((ship) => (
            <tr key={ship.id}>
              <td>{nameOf(ship.id)}</td>
              <td>
                <input
                  type="number"
                  min={0}
                  value={ships[ship.id] ?? 0}
                  onChange={(e) =>
                    setShips({ ...ships, [ship.id]: Math.max(0, Number(e.target.value)) })
                  }
                  style={{ width: 100 }}
                />
              </td>
            </tr>
          ))}
        </tbody>
      </table>

      {mission === 8 && (
        <div style={{ marginBottom: 12 }}>
          <label>
            {tf('Main', 'COLONY_NAME', 'Название колонии')}:{' '}
            <input
              type="text"
              value={colonyName}
              onChange={(e) => setColonyName(e.target.value)}
              placeholder={tf('Main', 'COLONY_NAME_PLACEHOLDER', 'Colony')}
              maxLength={40}
              style={{ width: 200 }}
            />
          </label>
        </div>
      )}

      {(mission === 7 || mission === 8) && (
        <>
          <h3>
            {t('global', 'METAL')} / {t('global', 'SILICON')} / {t('global', 'HYDROGEN')}
          </h3>
          <div style={{ display: 'flex', gap: 8, marginBottom: 12 }}>
            <input type="number" min={0} value={metal} onChange={(e) => setMetal(Number(e.target.value))} placeholder={t('global', 'METAL')} />
            <input type="number" min={0} value={silicon} onChange={(e) => setSilicon(Number(e.target.value))} placeholder={t('global', 'SILICON')} />
            <input type="number" min={0} value={hydrogen} onChange={(e) => setHydrogen(Number(e.target.value))} placeholder={t('global', 'HYDROGEN')} />
          </div>
        </>
      )}

      <div style={{ marginBottom: 12 }}>
        {tf('Main', 'SPEED_PERCENT', 'Скорость %')}&nbsp;
        <input type="range" min={10} max={100} step={10} value={speed} onChange={(e) => setSpeed(Number(e.target.value))} />
        &nbsp;{speed}%
      </div>

      <button type="button" disabled={send.isPending} onClick={() => send.mutate()}>
        {send.isPending ? '…' : tf('Main', 'FLEET_MESSAGE_OWN', 'Отправить')}
      </button>
      {send.isError && (
        <div className="ox-error">
          {send.error instanceof Error ? send.error.message : t('global', 'ERROR')}
        </div>
      )}

      <h3>{t('global', 'MENU_FLEET')}</h3>
      {list.length === 0 ? (
        <p>—</p>
      ) : (
        <table className="ox-table">
          <thead>
            <tr>
              <th>#</th>
              <th>{t('Main', 'POSITION')}</th>
              <th>{tf('Main', 'FLEET_STATE', 'Статус')}</th>
              <th>{tf('Main', 'FLEET_ARRIVE', 'Прилёт')}</th>
              <th>{tf('Main', 'FLEET_RETURN', 'Возврат')}</th>
              <th />
            </tr>
          </thead>
          <tbody>
            {list.map((f) => (
              <tr key={f.id}>
                <td>{f.id.slice(0, 8)}</td>
                <td>
                  [{f.dst_galaxy}:{f.dst_system}:{f.dst_position}
                  {f.dst_is_moon ? ' 🌑' : ''}]
                </td>
                <td>{f.state}</td>
                <td>{new Date(f.arrive_at).toLocaleTimeString('ru-RU')}</td>
                <td>{f.return_at ? new Date(f.return_at).toLocaleTimeString('ru-RU') : '—'}</td>
                <td>
                  {f.state === 'outbound' && (
                    <button
                      type="button"
                      disabled={recall.isPending}
                      onClick={() => recall.mutate(f.id)}
                    >
                      {tf('Main', 'FLEET_RETURN_BACK', 'Отозвать')}
                    </button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
      {recall.isError && (
        <div className="ox-error">
          {recall.error instanceof Error ? recall.error.message : t('global', 'ERROR')}
        </div>
      )}
    </section>
  );
}
