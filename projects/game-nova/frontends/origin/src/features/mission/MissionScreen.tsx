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
import { Link, useNavigate, useParams, useSearchParams } from 'react-router-dom';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { ApiError } from '@/api/client';
import { dispatchFleet, stargateJump } from '@/api/fleet';
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
  // План 72.1.47: HOLDING (legacy mode=17).
  { code: 17, key: 'missionHolding' },
];

export function MissionScreen() {
  const { planetId: urlId } = useParams();
  const [search] = useSearchParams();
  const navigate = useNavigate();
  const { planetId, planet, planets } = useResolvedPlanet(urlId);
  const { t } = useTranslation();
  const qc = useQueryClient();
  const [stargateErr, setStargateErr] = useState<string | null>(null);
  const [stargateDst, setStargateDst] = useState('');
  const [stargateCounts, setStargateCounts] = useState<Record<string, string>>({});

  const [galaxy, setGalaxy] = useState(() => Number(search.get('g') ?? 1));
  const [system, setSystem] = useState(() => Number(search.get('s') ?? 1));
  const [position, setPosition] = useState(() => Number(search.get('p') ?? 1));
  const [mission, setMission] = useState<MissionCode>(7);
  const [speed, setSpeed] = useState(100);
  const [counts, setCounts] = useState<Record<string, string>>({});
  // План 72.1.47: resource-carry (legacy missions.tpl fields fleetMetal/Silicon/Hydrogen).
  const [carryMetal, setCarryMetal] = useState('0');
  const [carrySilicon, setCarrySilicon] = useState('0');
  const [carryHydrogen, setCarryHydrogen] = useState('0');
  const [colonyName, setColonyName] = useState('');
  const [acsGroupId, setAcsGroupId] = useState('');
  // План 72.1.47: HOLDING duration (legacy holdingtime, 0..99 часов).
  const [holdingHours, setHoldingHours] = useState('0');

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

  // План 72.1.47: StarGate jump (legacy `Mission.class.php::starGateJump`).
  // Доступно только если src — луна с jump_gate>=1, dst — другая луна с
  // jump_gate>=1. Backend проверяет всё (cooldown / banned units / position).
  const stargate = useMutation({
    mutationFn: (input: { src: string; dst: string; ships: Record<string, number> }) =>
      stargateJump({ src_planet_id: input.src, dst_planet_id: input.dst, ships: input.ships }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: QK.shipyardInventory(planetId!) });
      void qc.invalidateQueries({ queryKey: QK.planets() });
      setStargateCounts({});
      setStargateErr(null);
    },
    onError: (e) => setStargateErr((e as ApiError).message),
  });

  const ships = useMemo(() => catalogByGroup('ship'), []);

  if (!planetId) {
    return <div className="idiv">{t('overview', 'noPlanets')}</div>;
  }

  const inv: { ships: Record<string, number>; defense: Record<string, number> } =
    invQ.data ?? { ships: {}, defense: {} };

  // План 72.1.47: список лун пользователя для StarGate-целей.
  const otherMoons = useMemo(
    () => planets.filter((p) => p.is_moon && p.id !== planetId),
    [planets, planetId],
  );

  function handleStargate() {
    if (!planetId || !stargateDst) return;
    const selected: Record<string, number> = {};
    for (const [k, v] of Object.entries(stargateCounts)) {
      const n = Math.max(0, Math.floor(Number(v) || 0));
      if (n > 0) selected[k] = n;
    }
    if (Object.keys(selected).length === 0) return;
    stargate.mutate({ src: planetId, dst: stargateDst, ships: selected });
  }

  function handleSend() {
    const selected: Record<string, number> = {};
    for (const [k, v] of Object.entries(counts)) {
      const n = Math.max(0, Math.floor(Number(v) || 0));
      if (n > 0) selected[k] = n;
    }
    if (!planetId || Object.keys(selected).length === 0) return;
    const sp = Math.max(10, Math.min(100, speed));
    const cm = Math.max(0, Math.floor(Number(carryMetal) || 0));
    const cs = Math.max(0, Math.floor(Number(carrySilicon) || 0));
    const ch = Math.max(0, Math.floor(Number(carryHydrogen) || 0));
    const input: FleetDispatchInput = {
      src_planet_id: planetId,
      dst: { galaxy, system, position },
      ships: selected,
      speed_percent: sp,
      mission,
      ...(cm > 0 ? { carry_metal: cm } : {}),
      ...(cs > 0 ? { carry_silicon: cs } : {}),
      ...(ch > 0 ? { carry_hydrogen: ch } : {}),
      // План 72.1.47: ACS-группа только для mission=12; colony — только для 8;
      // holding_hours — только для 17.
      ...(mission === 12 && acsGroupId ? { acs_group_id: acsGroupId } : {}),
      ...(mission === 8 && colonyName ? { colony_name: colonyName } : {}),
      ...(mission === 17
        ? { holding_hours: Math.max(0, Math.min(99, Math.floor(Number(holdingHours) || 0))) }
        : {}),
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
          {/* План 72.1.47: resource-carry (legacy missions.tpl fleetMetal/Silicon/Hydrogen). */}
          <tr>
            <td>{t('overview', 'metal') || 'Металл'}</td>
            <td>
              <input
                type="number"
                min={0}
                value={carryMetal}
                onChange={(e) => setCarryMetal(e.target.value)}
                aria-label={t('overview', 'metal') || 'Металл'}
              />
            </td>
          </tr>
          <tr>
            <td>{t('overview', 'silicon') || 'Кремний'}</td>
            <td>
              <input
                type="number"
                min={0}
                value={carrySilicon}
                onChange={(e) => setCarrySilicon(e.target.value)}
                aria-label={t('overview', 'silicon') || 'Кремний'}
              />
            </td>
          </tr>
          <tr>
            <td>{t('overview', 'hydrogen') || 'Водород'}</td>
            <td>
              <input
                type="number"
                min={0}
                value={carryHydrogen}
                onChange={(e) => setCarryHydrogen(e.target.value)}
                aria-label={t('overview', 'hydrogen') || 'Водород'}
              />
            </td>
          </tr>
          {/* Mission=8 colonize → colony_name; mission=12 ACS → acs_group_id. */}
          {mission === 8 && (
            <tr>
              <td>{t('mission', 'colonyName') || 'Имя колонии'}</td>
              <td>
                <input
                  type="text"
                  value={colonyName}
                  maxLength={32}
                  onChange={(e) => setColonyName(e.target.value)}
                  placeholder="Colony"
                  aria-label="colony_name"
                />
              </td>
            </tr>
          )}
          {mission === 12 && (
            <tr>
              <td>{t('mission', 'acsGroupId') || 'ACS group ID'}</td>
              <td>
                <input
                  type="text"
                  value={acsGroupId}
                  onChange={(e) => setAcsGroupId(e.target.value)}
                  placeholder={t('mission', 'acsGroupHint') || 'пусто = создать новую'}
                  aria-label="acs_group_id"
                />
              </td>
            </tr>
          )}
          {/* План 72.1.47: HOLDING — длительность удержания на цели. */}
          {mission === 17 && (
            <tr>
              <td>{t('mission', 'holdingHours') || 'Удерживать (ч)'}</td>
              <td>
                <input
                  type="number"
                  min={0}
                  max={99}
                  value={holdingHours}
                  onChange={(e) => setHoldingHours(e.target.value)}
                  aria-label="holding_hours"
                />
              </td>
            </tr>
          )}
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

      {/* План 72.1.47: ссылка на отзыв флотов (legacy retreat action был
          здесь же, у нас вынесен в /fleet-operations). */}
      <div className="idiv">
        <Link to="/fleet-operations">
          ← {t('mission', 'recallFleetsLink') || 'Активные флоты / отозвать'}
        </Link>
      </div>

      {/* План 72.1.47: StarGate jump (legacy `Mission.class.php::starGateJump`).
          Доступно только если src — луна и есть другие луны игрока. */}
      {planet?.is_moon && otherMoons.length > 0 && (
        <table className="ntable">
          <thead>
            <tr>
              <th colSpan={2}>{t('mission', 'stargateTitle') || '🌀 StarGate jump'}</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td>{t('fleet', 'destination') || 'Цель'}</td>
              <td>
                <select
                  value={stargateDst}
                  onChange={(e) => setStargateDst(e.target.value)}
                  aria-label="stargate_dst"
                >
                  <option value="">—</option>
                  {otherMoons.map((m) => (
                    <option key={m.id} value={m.id}>
                      {m.name} [{m.galaxy}:{m.system}:{m.position} L]
                    </option>
                  ))}
                </select>
              </td>
            </tr>
            {ships.map((entry) => {
              const stock = inv.ships[String(entry.id)] ?? 0;
              if (stock === 0) return null;
              const [grp, key] = entry.i18n.split('.') as [string, string];
              return (
                <tr key={`sg-${entry.id}`}>
                  <td>{t(grp, key)} ({stock})</td>
                  <td>
                    <input
                      type="number"
                      min={0}
                      max={stock}
                      value={stargateCounts[String(entry.id)] ?? ''}
                      onChange={(e) =>
                        setStargateCounts((c) => ({
                          ...c,
                          [String(entry.id)]: e.target.value,
                        }))
                      }
                      aria-label={`stargate ${t(grp, key)}`}
                    />
                  </td>
                </tr>
              );
            })}
            <tr>
              <td colSpan={2} align="center">
                <button
                  type="button"
                  className="button"
                  onClick={handleStargate}
                  disabled={stargate.isPending || !stargateDst}
                >
                  {t('mission', 'stargateBtn') || 'Прыжок'}
                </button>
                {stargateErr && (
                  <div>
                    <span className="false">{stargateErr}</span>
                  </div>
                )}
              </td>
            </tr>
          </tbody>
        </table>
      )}
    </>
  );
}
