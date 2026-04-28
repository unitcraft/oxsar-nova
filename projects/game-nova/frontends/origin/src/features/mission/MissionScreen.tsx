// S-006 Mission — отправка миссии (план 72 Ф.2 Spring 1).
//
// Pixel-perfect зеркало legacy `missions.tpl`:
//   1) Таблица «Корабли» с input количества по каждому ship-id.
//   2) Поля: галактика/система/позиция назначения, скорость (10..100%),
//      тип миссии (атака=10, шпионаж=11, экспедиция=15, транспорт=7,
//      колонизация=8, переработка=9, перебазирование=6, ACS=12).
//   3) Кнопка «Отправить флот» → POST /api/fleet с FleetDispatch.
//
// Endpoint: POST /api/fleet (mission-код в payload, см. types.ts MissionCode).
//
// Параметры из URL: ?g=1&s=42&p=7 (приходят с Galaxy экрана).

import { useState, useMemo } from 'react';
import { useNavigate, useParams, useSearchParams } from 'react-router-dom';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { dispatchFleet } from '@/api/fleet';
import { fetchShipyardInventory } from '@/api/shipyard';
import { QK } from '@/api/query-keys';
import { useResolvedPlanet } from '@/features/common/useResolvedPlanet';
import { catalogByGroup } from '@/features/common/catalog';
import type { FleetDispatchInput, MissionCode } from '@/api/types';
import { useTranslation } from '@/i18n/i18n';

const MISSION_CODES: { code: MissionCode; key: string }[] = [
  { code: 7, key: 'missionTransport' },
  { code: 10, key: 'missionAttack' },
  { code: 11, key: 'missionSpy' },
  { code: 8, key: 'missionColonize' },
  { code: 9, key: 'missionRecycle' },
  { code: 15, key: 'missionExpedition' },
  { code: 6, key: 'missionRebase' },
  { code: 12, key: 'missionAttack' },
];

export function MissionScreen() {
  const { planetId: urlId } = useParams();
  const [search] = useSearchParams();
  const navigate = useNavigate();
  const { planetId } = useResolvedPlanet(urlId);
  const { t } = useTranslation();
  const qc = useQueryClient();

  const [galaxy, setGalaxy] = useState(() => Number(search.get('g') ?? 1));
  const [system, setSystem] = useState(() => Number(search.get('s') ?? 1));
  const [position, setPosition] = useState(() => Number(search.get('p') ?? 1));
  const [mission, setMission] = useState<MissionCode>(7);
  const [speed, setSpeed] = useState(100);
  const [counts, setCounts] = useState<Record<string, string>>({});

  const invQ = useQuery({
    queryKey: planetId ? QK.shipyardInventory(planetId) : ['noop-mission'],
    queryFn: () =>
      planetId
        ? fetchShipyardInventory(planetId)
        : Promise.resolve({ ships: {}, defense: {} }),
    enabled: planetId !== null,
  });

  const dispatch = useMutation({
    mutationFn: (input: FleetDispatchInput) => dispatchFleet(input),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: QK.fleet() });
      if (planetId) {
        void qc.invalidateQueries({ queryKey: QK.shipyardInventory(planetId) });
        void qc.invalidateQueries({ queryKey: QK.planet(planetId) });
      }
      navigate('/');
    },
  });

  const ships = useMemo(() => catalogByGroup('ship'), []);

  if (!planetId) {
    return <div className="idiv">{t('overview', 'noPlanets')}</div>;
  }

  const inv: { ships: Record<string, number>; defense: Record<string, number> } =
    invQ.data ?? { ships: {}, defense: {} };

  function handleSend() {
    const selected: Record<string, number> = {};
    for (const [k, v] of Object.entries(counts)) {
      const n = Math.max(0, Math.floor(Number(v) || 0));
      if (n > 0) selected[k] = n;
    }
    if (!planetId || Object.keys(selected).length === 0) return;
    const sp = Math.max(10, Math.min(100, speed));
    const input: FleetDispatchInput = {
      src_planet_id: planetId,
      dst: { galaxy, system, position },
      ships: selected,
      speed_percent: sp,
      mission,
    };
    dispatch.mutate(input);
  }

  return (
    <>
      <table className="ntable">
        <thead>
          <tr>
            <th colSpan={4}>{t('mission', 'fleet')}</th>
          </tr>
          <tr>
            <th colSpan={2}>{t('main', 'unitId')}</th>
            <th>{t('shipyard', 'inStock', { count: '' })}</th>
            <th>{t('galaxy', 'quantity')}</th>
          </tr>
        </thead>
        <tbody>
          {ships.map((entry) => {
            const [grp, key] = entry.i18n.split('.') as [string, string];
            const stock = inv.ships[String(entry.id)] ?? 0;
            return (
              <tr key={entry.id}>
                <td width="1px">#{entry.id}</td>
                <td>{t(grp, key)}</td>
                <td>{stock}</td>
                <td>
                  <input
                    type="number"
                    className="center"
                    min={0}
                    max={stock}
                    value={counts[String(entry.id)] ?? ''}
                    onChange={(e) =>
                      setCounts((c) => ({ ...c, [String(entry.id)]: e.target.value }))
                    }
                    aria-label={`${t(grp, key)} ${t('galaxy', 'quantity')}`}
                  />
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>

      <table className="ntable">
        <thead>
          <tr>
            <th colSpan={2}>{t('fleet', 'destination')}</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>{t('galaxy', 'galaxy')}</td>
            <td>
              <input
                type="number"
                value={galaxy}
                onChange={(e) => setGalaxy(Number(e.target.value))}
                aria-label={t('galaxy', 'galaxy')}
              />
            </td>
          </tr>
          <tr>
            <td>{t('galaxy', 'system')}</td>
            <td>
              <input
                type="number"
                value={system}
                onChange={(e) => setSystem(Number(e.target.value))}
                aria-label={t('galaxy', 'system')}
              />
            </td>
          </tr>
          <tr>
            <td>{t('main', 'position')}</td>
            <td>
              <input
                type="number"
                value={position}
                onChange={(e) => setPosition(Number(e.target.value))}
                aria-label={t('main', 'position')}
              />
            </td>
          </tr>
          <tr>
            <td>{t('fleet', 'missionLabel')}</td>
            <td>
              <select
                value={mission}
                onChange={(e) => setMission(Number(e.target.value) as MissionCode)}
                aria-label={t('fleet', 'missionLabel')}
              >
                {MISSION_CODES.map((m) => (
                  <option key={m.code} value={m.code}>
                    {t('fleet', m.key)}
                  </option>
                ))}
              </select>
            </td>
          </tr>
          <tr>
            <td>{t('fleet', 'speedPct', { pct: speed })}</td>
            <td>
              <input
                type="number"
                min={10}
                max={100}
                step={10}
                value={speed}
                onChange={(e) => setSpeed(Number(e.target.value))}
                aria-label="speed"
              />
            </td>
          </tr>
          <tr>
            <td colSpan={2} align="center">
              <button
                type="button"
                className="button"
                onClick={handleSend}
                disabled={dispatch.isPending}
              >
                {t('fleet', 'sendButton')}
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </>
  );
}
