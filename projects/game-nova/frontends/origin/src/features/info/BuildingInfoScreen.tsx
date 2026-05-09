// S-018 BuildingInfo — pixel-perfect клон legacy `constructions.tpl`
// в режиме `info_id > 0` (вызывается через `game.php/ConstructionInfo/<id>`).
//
// Структура одной таблицы:
//   <th colspan=3>Уровень N (+M)  <название>
//   <td image><td FullDesc + required_res_table><td upgrade-link>
// + (если level>0 и demolish>0) блок:
//   <th colspan=3>DEMOLISH name level
//   <td colspan=3>REQUIRES <m> <s> <h> <br/> PRODUCTION_TIME <t>
//   <td colspan=3>Снести-link  (если ресурсов хватает)
// + (опц) pack-блок.

import { useParams, Link } from 'react-router-dom';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { fetchBuildingCatalog } from '@/api/catalog';
import {
  fetchBuildingsOverview,
  fetchBuildingQueue,
  demolishBuilding,
  packBuilding,
  enqueueBuilding,
  cancelBuildingTask,
} from '@/api/buildings';
import { QK } from '@/api/query-keys';
import { useTranslation } from '@/i18n/i18n';
import { findCatalog } from '@/features/common/catalog';
import { useResolvedPlanet } from '@/features/common/useResolvedPlanet';
import { ConfirmDialog, useConfirm } from '@/features/common/ConfirmDialog';
import { RequiredResTable } from '@/features/common/RequiredResTable';
import { ConstructionProgress } from '@/features/common/ConstructionProgress';
import { useAutoInvalidateOnTaskEnd } from '@/features/common/useAutoInvalidateOnTaskEnd';
import type { ApiError } from '@/api/client';

export function BuildingInfoScreen() {
  const params = useParams<{ type?: string }>();
  const type = params.type ?? '';
  const { t } = useTranslation();
  const qc = useQueryClient();
  const { planetId, planet } = useResolvedPlanet();
  const { confirm, dialogProps } = useConfirm();

  const catalogQ = useQuery({
    queryKey: QK.buildingCatalog(type),
    queryFn: () => fetchBuildingCatalog(type),
    enabled: type.length > 0,
    staleTime: 60 * 60 * 1000,
  });

  const overviewQ = useQuery({
    queryKey: planetId ? QK.buildingsOverview(planetId) : ['noop-bo'],
    queryFn: () =>
      planetId ? fetchBuildingsOverview(planetId) : Promise.reject(),
    enabled: planetId !== null,
  });

  const queueQ = useQuery({
    queryKey: planetId ? QK.buildingQueue(planetId) : ['noop-bq'],
    queryFn: () => (planetId ? fetchBuildingQueue(planetId) : Promise.resolve([])),
    enabled: planetId !== null,
    refetchInterval: 10_000,
  });

  const enqueue = useMutation({
    mutationFn: (unitId: number) => enqueueBuilding(planetId!, unitId),
    onSuccess: () => {
      window.scrollTo({ top: 0, behavior: 'smooth' });
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
        void qc.invalidateQueries({ queryKey: QK.planet(planetId) });
      }
    },
  });

  const [demolishErr, setDemolishErr] = useState<string | null>(null);
  const demolishMut = useMutation({
    mutationFn: () => {
      if (!planetId || !catalogQ.data) return Promise.reject(new Error('no planet'));
      return demolishBuilding(planetId, catalogQ.data.id);
    },
    onSuccess: () => {
      setDemolishErr(null);
      window.scrollTo({ top: 0, behavior: 'smooth' });
      if (planetId) {
        void qc.invalidateQueries({ queryKey: QK.buildingQueue(planetId) });
        void qc.invalidateQueries({ queryKey: QK.buildingsOverview(planetId) });
        void qc.invalidateQueries({ queryKey: QK.planet(planetId) });
      }
    },
    onError: (e) => {
      setDemolishErr((e as ApiError).message);
      window.scrollTo({ top: 0, behavior: 'smooth' });
    },
  });

  const [packErr, setPackErr] = useState<string | null>(null);
  const packMut = useMutation({
    mutationFn: () => {
      if (!planetId || !catalogQ.data) return Promise.reject(new Error('no planet'));
      return packBuilding(planetId, catalogQ.data.id);
    },
    onSuccess: () => {
      setPackErr(null);
      if (planetId) {
        void qc.invalidateQueries({ queryKey: QK.buildingsOverview(planetId) });
        void qc.invalidateQueries({ queryKey: QK.planet(planetId) });
        void qc.invalidateQueries({ queryKey: QK.artefacts() });
      }
    },
    onError: (e) => setPackErr((e as ApiError).message),
  });

  const queue = queueQ.data ?? [];
  useAutoInvalidateOnTaskEnd(queue, () => {
    if (!planetId) return;
    void qc.invalidateQueries({ queryKey: QK.buildingQueue(planetId) });
    void qc.invalidateQueries({ queryKey: QK.buildingsOverview(planetId) });
    void qc.invalidateQueries({ queryKey: QK.planet(planetId) });
  });

  if (catalogQ.isLoading) return <div className="idiv">…</div>;
  if (catalogQ.isError || !catalogQ.data) {
    return (
      <table className="ntable">
        <tbody>
          <tr>
            <td className="center">
              <i>{t('alliance', 'nothing')}</i>
            </td>
          </tr>
        </tbody>
      </table>
    );
  }

  const entry = catalogQ.data;
  const nameKey = entry.key.replace(/_([a-z])/g, (_, c: string) => c.toUpperCase());
  const name = t('info', nameKey);
  const fullDescKey = `${nameKey}FullDesc`;
  const descKey = `${nameKey}Desc`;
  const fullDesc = t('info', fullDescKey);
  const desc = t('info', descKey);
  const hasFull = !fullDesc.startsWith('[');
  const hasDesc = !desc.startsWith('[');
  const description = hasFull ? fullDesc : (hasDesc ? desc : '');
  const catalog = findCatalog(entry.id);
  const icon = catalog?.icon ?? '';

  const levels = overviewQ.data?.levels ?? {};
  const seconds = overviewQ.data?.build_seconds ?? {};
  const costs = overviewQ.data?.build_costs ?? {};
  const level = levels[String(entry.id)] ?? 0;
  // added_level — сколько уровней этого здания в очереди стройки/сноса.
  // Положительное при upgrade, отрицательное при demolish.
  const added = (queueQ.data ?? [])
    .filter((q) => q.unit_id === entry.id)
    .reduce((acc, q) => acc + (q.target_level > level ? 1 : -1), 0);
  const nextLevel = level + 1;
  const cost = costs[String(entry.id)] ?? { metal: 0, silicon: 0, hydrogen: 0 };
  const secs = seconds[String(entry.id)] ?? 0;

  const available = planet
    ? { metal: Math.floor(planet.metal), silicon: Math.floor(planet.silicon), hydrogen: Math.floor(planet.hydrogen) }
    : { metal: 0, silicon: 0, hydrogen: 0 };
  const enoughBuild =
    available.metal >= cost.metal &&
    available.silicon >= cost.silicon &&
    available.hydrogen >= cost.hydrogen;

  const isInQueue = queue.some((t) => t.unit_id === entry.id);
  const queueBusy = queue.length > 0;

  // Demolish: legacy `Construction.class.php:407-464` — показывается
  // только когда $true_level > 0 (level - addedBuilding > 0). Стоимость
  // сноса = (1/factor) × cost@(true_level), backend проверит точнее;
  // у нас в overview нет demolish_cost — используем эвристику cost@curLevel
  // как приближение для UI-баннера; backend всё равно валидирует.
  const trueLevel = level - Math.max(0, added);
  const canDemolish = trueLevel > 0;
  const demolishCost = costs[String(entry.id)] ?? { metal: 0, silicon: 0, hydrogen: 0 };
  const enoughDemolish =
    available.metal >= demolishCost.metal &&
    available.silicon >= demolishCost.silicon &&
    available.hydrogen >= demolishCost.hydrogen;

  return (
    <>
      {demolishErr && (
        <div className="false" style={{ padding: 4, marginBottom: 4, textAlign: 'center' }}>
          {demolishErr}
        </div>
      )}
      {packErr && (
        <div className="false" style={{ padding: 4, marginBottom: 4, textAlign: 'center' }}>
          {packErr}
        </div>
      )}

      {queue.length > 0 && (
        <table className="ntable">
          <tbody>
            <tr>
              <th colSpan={4}>{t('buildings', 'outstandingMissions')}</th>
            </tr>
            {queue.map((task, idx) => {
              const cat = findCatalog(task.unit_id);
              const [g, k] = cat ? (cat.i18n.split('.') as [string, string]) : ['info', ''];
              const tname = cat ? t(g, k) : `#${task.unit_id}`;
              return (
                <tr key={task.id}>
                  <td width="1px">{idx + 1}.</td>
                  <td>{tname}&nbsp;{task.target_level}</td>
                  <td width="130px">
                    <ConstructionProgress startAt={task.start_at} endAt={task.end_at} />
                  </td>
                  <td width="100px" align="center">
                    <button
                      type="button"
                      className="button"
                      disabled={cancel.isPending}
                      onClick={async () => {
                        if (await confirm({
                          title: t('buildings', 'cancelTask'),
                          message: t('buildings', 'cancelConfirm'),
                          destructive: true,
                        })) {
                          cancel.mutate(task.id);
                        }
                      }}
                    >
                      ✕
                    </button>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      )}

      <table className="ntable">
        <tbody>
          {/* Шапка: <th colspan=3> Уровень N (+M)  <название> */}
          <tr>
            <th colSpan={3}>
              {(level > 0 || added !== 0) && (
                <span style={{ float: 'right' }}>
                  {t('buildings', 'level', { n: level })}
                  {added > 0 && <span className="true"> (+{added})</span>}
                  {added < 0 && <span className="false"> ({added})</span>}
                </span>
              )}
              {name}
            </th>
          </tr>

          <tr>
            <td width="1px" style={{ verticalAlign: 'top' }}>
              {icon && (
                <Link to={`/building/${entry.id}`}>
                  <img
                    src={`/assets/origin/images/units/${icon}.gif`}
                    alt={name}
                    onError={(e) => {
                      (e.target as HTMLImageElement).style.display = 'none';
                    }}
                  />
                </Link>
              )}
            </td>
            <td style={{ verticalAlign: 'top', textAlign: 'left' }}>
              {description && (
                <div style={{ fontSize: 'smaller', clear: 'both' }}>{description}</div>
              )}
              <div style={{ marginTop: 6 }}>
                <RequiredResTable
                  metal={cost.metal}
                  silicon={cost.silicon}
                  hydrogen={cost.hydrogen}
                  available={available}
                  seconds={secs}
                />
              </div>
            </td>
            <td width="100px" align="center" style={{ verticalAlign: 'middle' }}>
              {queueBusy && isInQueue ? (
                <span className="false">{t('buildings', 'buildingAtWork')}</span>
              ) : (
                <button
                  type="button"
                  className={`btn-link ${enoughBuild ? 'true' : 'false'}`}
                  onClick={() => enqueue.mutate(entry.id)}
                  disabled={enqueue.isPending || !enoughBuild}
                >
                  {t('buildings', 'upgradeToLevel')} {nextLevel}
                </button>
              )}
            </td>
          </tr>

          {/* Demolish-блок (legacy `constructions.tpl:119-131`). */}
          {canDemolish && (
            <>
              <tr>
                <th colSpan={3} align="center">
                  {t('buildinginfo', 'demolish')} {name} {trueLevel}
                </th>
              </tr>
              <tr>
                <td colSpan={3} align="center">
                  {t('info', 'requires')}{' '}
                  {demolishCost.metal > 0 && <>{t('overview', 'metal')}: {demolishCost.metal} </>}
                  {demolishCost.silicon > 0 && <>{t('overview', 'silicon')}: {demolishCost.silicon} </>}
                  {demolishCost.hydrogen > 0 && <>{t('overview', 'hydrogen')}: {demolishCost.hydrogen} </>}
                  <br />
                  {t('info', 'requireTime')}{' '}
                  {Math.floor(secs / 2)}s
                </td>
              </tr>
              {!isInQueue && enoughDemolish && (
                <tr>
                  <td colSpan={3} align="center">
                    <button
                      type="button"
                      className="btn-link false"
                      disabled={demolishMut.isPending}
                      onClick={async () => {
                        if (await confirm({
                          title: t('buildinginfo', 'demolish'),
                          message: t('buildinginfo', 'demolishConfirm', {
                            building: name,
                            level: String(trueLevel - 1),
                          }),
                          destructive: true,
                        })) {
                          demolishMut.mutate();
                        }
                      }}
                    >
                      {t('buildinginfo', 'demolishNow')}
                    </button>
                  </td>
                </tr>
              )}
            </>
          )}

          {/* Pack-building блок (план 72.1.33 ч.2). */}
          {canDemolish && (
            <tr>
              <td colSpan={3} align="center">
                <button
                  type="button"
                  className="btn-link"
                  disabled={packMut.isPending}
                  onClick={async () => {
                    if (await confirm({
                      title: t('buildinginfo', 'packBuilding'),
                      message: t('buildinginfo', 'packConfirm'),
                    })) {
                      packMut.mutate();
                    }
                  }}
                >
                  📦 {t('buildinginfo', 'packBuilding')} {name} {trueLevel}
                </button>
              </td>
            </tr>
          )}
        </tbody>
      </table>
      <ConfirmDialog {...dialogProps} />
    </>
  );
}
