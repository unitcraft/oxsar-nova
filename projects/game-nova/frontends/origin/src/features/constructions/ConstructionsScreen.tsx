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
import { useAutoInvalidateOnTaskEnd } from '@/features/common/useAutoInvalidateOnTaskEnd';
import { catalogByGroup } from '@/features/common/catalog';
import { RequiredResTable } from '@/features/common/RequiredResTable';
import { ConstructionProgress } from '@/features/common/ConstructionProgress';
import { ConfirmDialog, useConfirm } from '@/features/common/ConfirmDialog';
import { fetchSettings } from '@/api/settings';
import { useTranslation } from '@/i18n/i18n';

export function ConstructionsScreen() {
  const { planetId, planet } = useResolvedPlanet();
  const { t } = useTranslation();
  const qc = useQueryClient();
  // План 72.1.53 ч.B: in-game confirm-dialog.
  const { confirm, dialogProps } = useConfirm();
  // План 72.1.55.E (effects): show_all_constructions preference
  // (legacy `Preferences::updateUserData::show_all_constructions`).
  // Если false — скрываем здания с unmet requirements.
  const settingsQ = useQuery({
    queryKey: QK.settings(),
    queryFn: fetchSettings,
    staleTime: 60_000,
  });
  const showAll = settingsQ.data?.show_all_constructions ?? true;

  const queueQ = useQuery({
    queryKey: planetId ? QK.buildingQueue(planetId) : ['noop-bq'],
    queryFn: () => (planetId ? fetchBuildingQueue(planetId) : Promise.resolve([])),
    enabled: planetId !== null,
    refetchInterval: 10_000,
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

  // План 72.1.59: авто-инвалидация очереди по окончании ближайшей
  // задачи (общий хук). Без него экран остаётся со старым snapshot'ом
  // когда прогресс-бар доезжает до 100%.
  const queueData = queueQ.data ?? [];
  useAutoInvalidateOnTaskEnd(queueData, () => {
    if (!planetId) return;
    void qc.invalidateQueries({ queryKey: QK.buildingQueue(planetId) });
    void qc.invalidateQueries({ queryKey: QK.buildingsOverview(planetId) });
    void qc.invalidateQueries({ queryKey: QK.planet(planetId) });
    void qc.invalidateQueries({ queryKey: QK.planets() });
  });

  if (!planetId) {
    return <div className="idiv">{t('overview', 'noPlanets')}</div>;
  }

  const queue = queueData;
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
                  <td>
                    {name}&nbsp;{task.target_level}
                  </td>
                  {/* План 72.1.45 §4: VIP-progress bar (legacy
                      constructions.tpl L.24-50 + ExtRepair L.244-249). */}
                  <td width="130px">
                    <ConstructionProgress
                      startAt={task.start_at}
                      endAt={task.end_at}
                    />
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
                      value={t('buildings', 'vipBtn')}
                      title={t('buildings', 'vipHint')}
                      onClick={async () => {
                        if (await confirm({
                          title: t('buildings', 'vipBtn'),
                          message: t('buildings', 'vipConfirm'),
                        })) {
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
            <th colSpan={3}>{t('buildings', 'constructions')}</th>
          </tr>

          {buildings
            // План 72.1.55.E (effects): фильтр по show_all_constructions.
            // Если выкл и здание не построено + есть unmet requirements —
            // скрываем (legacy 1:1).
            .filter((entry) => {
              if (showAll) return true;
              const lvl = levels[String(entry.id)] ?? 0;
              if (lvl > 0) return true; // уже построено — показываем
              const u = unmet[String(entry.id)] ?? [];
              return u.length === 0;
            })
            .map((entry) => {
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
                      {t('buildings', 'level', { n: level })}
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
                        {t('buildings', 'requirements')}:
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
                      {t('buildings', 'buildingAtWork')}
                    </span>
                  ) : (
                    <>
                      <button
                        type="button"
                        className={`btn-link ${enough ? 'true' : 'false'}`}
                        onClick={() => enqueue.mutate(entry.id)}
                        disabled={enqueue.isPending || !enough}
                      >
                        ⚒ {t('buildings', 'upgradeToLevel')}<br />
                        {level + 1}
                      </button>
                      {/* План 72.1.40: legacy demolish action в списке.
                          План 72.1.59: красная button-danger + 💥 (взрыв)
                          подчёркивает деструктивность действия. */}
                      {level > 0 && (
                        <div style={{ marginTop: 4 }}>
                          <button
                            type="button"
                            className="button-danger"
                            disabled={demolish.isPending}
                            title={t('buildinginfo', 'demolish')}
                            onClick={async () => {
                              if (await confirm({
                                title: t('buildinginfo', 'demolish'),
                                message: t('buildinginfo', 'demolishConfirm', {
                                  building: t(group, key),
                                  level: level - 1,
                                }),
                                destructive: true,
                              })) {
                                demolish.mutate(entry.id);
                              }
                            }}
                          >
                            💥 {t('buildinginfo', 'demolishNow')}
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
      <ConfirmDialog {...dialogProps} />
    </>
  );
}
