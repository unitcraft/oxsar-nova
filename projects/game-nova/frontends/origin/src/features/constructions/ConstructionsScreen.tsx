// S-002 Constructions — строительство зданий (план 72.1 ч.20).
// Pixel-perfect клон legacy constructions.tpl + required_res_table.tpl.

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  cancelBuildingTask,
  demolishBuilding,
  enqueueBuilding,
  fetchBuildingQueue,
  fetchBuildingsOverview,
  startBuildingVIP,
  type BuildingsOverview,
} from '@/api/buildings';
import { QK } from '@/api/query-keys';
import { useResolvedPlanet } from '@/features/common/useResolvedPlanet';
import { catalogByGroup } from '@/features/common/catalog';
import { RequiredResTable } from '@/features/common/RequiredResTable';
import { useTranslation } from '@/i18n/i18n';

export function ConstructionsScreen() {
  const { planetId, planet } = useResolvedPlanet();
  const { t } = useTranslation();
  const qc = useQueryClient();

  const queueQ = useQuery({
    queryKey: planetId ? QK.buildingQueue(planetId) : ['noop-bq'],
    queryFn: () => (planetId ? fetchBuildingQueue(planetId) : Promise.resolve([])),
    enabled: planetId !== null,
  });

  const overviewQ = useQuery<BuildingsOverview>({
    queryKey: planetId ? QK.buildingsOverview(planetId) : ['noop-bo'],
    queryFn: () =>
      planetId
        ? fetchBuildingsOverview(planetId)
        : Promise.resolve<BuildingsOverview>({
            levels: {},
            build_seconds: {},
            build_costs: {},
            requirements_unmet: {},
          }),
    enabled: planetId !== null,
  });

  const enqueue = useMutation({
    mutationFn: (unitId: number) => enqueueBuilding(planetId!, unitId),
    onSuccess: () => {
      if (planetId) {
        void qc.invalidateQueries({ queryKey: QK.buildingQueue(planetId) });
        void qc.invalidateQueries({ queryKey: QK.buildingsOverview(planetId) });
        void qc.invalidateQueries({ queryKey: QK.planet(planetId) });
        void qc.invalidateQueries({ queryKey: QK.planets() });
      }
    },
  });

  const cancel = useMutation({
    mutationFn: (taskId: string) => cancelBuildingTask(planetId!, taskId),
    onSuccess: () => {
      if (planetId) {
        void qc.invalidateQueries({ queryKey: QK.buildingQueue(planetId) });
        void qc.invalidateQueries({ queryKey: QK.buildingsOverview(planetId) });
      }
    },
  });

  // План 72.1.40: legacy `Constructions` имеет ссылку «снести здание»
  // (DemolishConstruction). Раньше demolish был только в /building/:type
  // (BuildingInfoScreen), но в legacy он доступен и в списке зданий.
  const demolish = useMutation({
    mutationFn: (unitId: number) => demolishBuilding(planetId!, unitId),
    onSuccess: () => {
      if (planetId) {
        void qc.invalidateQueries({ queryKey: QK.buildingQueue(planetId) });
        void qc.invalidateQueries({ queryKey: QK.buildingsOverview(planetId) });
        void qc.invalidateQueries({ queryKey: QK.planet(planetId) });
      }
    },
  });

  // План 72.1.44: VIP-instant старт стройки за credits.
  const vip = useMutation({
    mutationFn: (taskId: string) => startBuildingVIP(planetId!, taskId),
    onSuccess: () => {
      if (planetId) {
        void qc.invalidateQueries({ queryKey: QK.buildingQueue(planetId) });
        void qc.invalidateQueries({ queryKey: QK.me() });
      }
    },
  });

  if (!planetId) {
    return <div className="idiv">{t('overview', 'noPlanets')}</div>;
  }

  const queue = queueQ.data ?? [];
  const isMoon = planet?.is_moon === true;

  const allBuildings = catalogByGroup('building');
  const buildings = allBuildings.filter((b) =>
    isMoon ? b.moonOnly === true || b.id < 54 || b.id >= 100 : b.moonOnly !== true,
  );

  const levels = overviewQ.data?.levels ?? {};
  const buildSecs = overviewQ.data?.build_seconds ?? {};
  const buildCosts = overviewQ.data?.build_costs ?? {};
  const unmet = overviewQ.data?.requirements_unmet ?? {};

  const available = planet
    ? {
        metal: Math.floor(planet.metal),
        silicon: Math.floor(planet.silicon),
        hydrogen: Math.floor(planet.hydrogen),
      }
    : { metal: 0, silicon: 0, hydrogen: 0 };

  function canBuild(unitId: number): boolean {
    const c = buildCosts[String(unitId)];
    if (!c) return false;
    if (
      available.metal < c.metal ||
      available.silicon < c.silicon ||
      available.hydrogen < c.hydrogen
    ) {
      return false;
    }
    return true;
  }

  return (
    <>
      {queue.length > 0 && (
        <table className="ntable">
          <tbody>
            <tr>
              <th colSpan={5}>{t('buildings', 'outstandingMissions')}</th>
            </tr>
            {queue.map((task, idx) => {
              const cat = allBuildings.find((b) => b.id === task.unit_id);
              const [g, k] = cat
                ? (cat.i18n.split('.') as [string, string])
                : ['info', ''];
              const name = cat ? t(g, k) : `#${task.unit_id}`;
              return (
                <tr key={task.id}>
                  <td width="1px">{idx + 1}.</td>
                  <td colSpan={2}>
                    {name}&nbsp;{task.target_level}
                  </td>
                  <td width="100px">
                    <input
                      type="button"
                      className="button"
                      value={t('info', 'abort')}
                      onClick={() => cancel.mutate(task.id)}
                      disabled={cancel.isPending}
                    />
                  </td>
                  {/* План 72.1.44: VIP-instant старт за credits. */}
                  <td width="80px">
                    <input
                      type="button"
                      className="button"
                      value={t('buildings', 'vipBtn') || '⚡ VIP'}
                      title={t('buildings', 'vipHint') || 'Мгновенный старт за кредиты'}
                      onClick={() => {
                        if (
                          window.confirm(
                            (t('buildings', 'vipConfirm') as string) ||
                              'Мгновенный старт стройки за кредиты?',
                          )
                        ) {
                          vip.mutate(task.id);
                        }
                      }}
                      disabled={vip.isPending}
                    />
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      )}

      <table className="ntable">
        <tbody>
          <tr>
            <th colSpan={3}>{t('buildings', 'constructions') ?? 'Постройки'}</th>
          </tr>

          {buildings.map((entry) => {
            const [group, key] = entry.i18n.split('.') as [string, string];
            const level = levels[String(entry.id)] ?? 0;
            const secs = buildSecs[String(entry.id)] ?? 0;
            const cost = buildCosts[String(entry.id)] ?? {
              metal: 0,
              silicon: 0,
              hydrogen: 0,
            };
            const requirementsUnmet = unmet[String(entry.id)] ?? [];
            const hasRequirements = requirementsUnmet.length > 0;
            const descKey = `${key}Desc`;
            const desc = t(group, descKey);
            const hasDesc = !desc.startsWith('[');
            const enough = canBuild(entry.id);
            const queueBusy = queue.length > 0;

            return (
              <tr key={entry.id}>
                <td width="1px" style={{ verticalAlign: 'top' }}>
                  <img
                    src={`/assets/origin/images/units/${entry.icon}.gif`}
                    alt={t(group, key)}
                    onError={(e) => {
                      (e.target as HTMLImageElement).style.display = 'none';
                    }}
                  />
                </td>
                <td style={{ verticalAlign: 'top' }}>
                  <div style={{ width: '100%' }}>
                    <span style={{ float: 'right' }}>
                      Уровень {level}
                    </span>
                    {t(group, key)}
                  </div>
                  {hasDesc && (
                    <div style={{ clear: 'both', fontSize: 'smaller' }}>
                      {desc}
                    </div>
                  )}
                  {hasRequirements ? (
                    <div style={{ marginTop: 6 }}>
                      <span className="normal">
                        {t('buildings', 'requirements') ?? 'Требования'}:
                      </span>
                      <br />
                      {requirementsUnmet.map((u) => {
                        const reqCat = allBuildings.find((b) => b.id === u.unit_id);
                        const [rg, rk] = reqCat
                          ? (reqCat.i18n.split('.') as [string, string])
                          : ['info', ''];
                        const reqName = reqCat ? t(rg, rk) : `#${u.unit_id}`;
                        return (
                          <span
                            key={u.unit_id}
                            className="false"
                            style={{ display: 'inline-block', marginRight: 8 }}
                          >
                            {reqName} {u.required_level}
                          </span>
                        );
                      })}
                    </div>
                  ) : (
                    <div style={{ marginTop: 6 }}>
                      <RequiredResTable
                        metal={cost.metal}
                        silicon={cost.silicon}
                        hydrogen={cost.hydrogen}
                        available={available}
                        seconds={secs}
                      />
                    </div>
                  )}
                </td>
                <td
                  width="100px"
                  align="center"
                  style={{ verticalAlign: 'top' }}
                >
                  {hasRequirements ? (
                    <span className="false">—</span>
                  ) : queueBusy ? (
                    <span className="false">
                      {t('buildings', 'buildingAtWork') ?? 'Занято'}
                    </span>
                  ) : (
                    <>
                      <button
                        type="button"
                        className={`btn-link ${enough ? 'true' : 'false'}`}
                        onClick={() => enqueue.mutate(entry.id)}
                        disabled={enqueue.isPending || !enough}
                      >
                        {t('buildings', 'upgradeToLevel') ?? 'Построить'}<br />
                        уровень {level + 1}
                      </button>
                      {/* План 72.1.40: legacy demolish action в списке. */}
                      {level > 0 && (
                        <div style={{ marginTop: 4 }}>
                          <button
                            type="button"
                            className="button"
                            disabled={demolish.isPending}
                            title={t('buildinginfo', 'demolish') || 'Снос здания'}
                            onClick={() => {
                              if (
                                window.confirm(
                                  (t('buildinginfo', 'demolishConfirm') as string) ||
                                    'Снести здание на 1 уровень?',
                                )
                              ) {
                                demolish.mutate(entry.id);
                              }
                            }}
                          >
                            ⚒ {t('buildinginfo', 'demolishNow') || 'Снести'}
                          </button>
                        </div>
                      )}
                    </>
                  )}
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </>
  );
}
