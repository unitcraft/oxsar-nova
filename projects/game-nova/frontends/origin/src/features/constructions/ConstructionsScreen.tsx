// S-002 Constructions — строительство зданий (план 72 Ф.2 Spring 1 → финализация).
// Pixel-perfect клон legacy constructions.tpl.

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  cancelBuildingTask,
  enqueueBuilding,
  fetchBuildingQueue,
} from '@/api/buildings';
import { fetchResourceReport } from '@/api/resource';
import { QK } from '@/api/query-keys';
import { useResolvedPlanet } from '@/features/common/useResolvedPlanet';
import { catalogByGroup } from '@/features/common/catalog';
import { useTranslation } from '@/i18n/i18n';
import { secondsUntil, formatDuration } from '@/lib/format';

export function ConstructionsScreen() {
  const { planetId } = useResolvedPlanet();
  const { t } = useTranslation();
  const qc = useQueryClient();

  const queueQ = useQuery({
    queryKey: planetId ? QK.buildingQueue(planetId) : ['noop-bq'],
    queryFn: () => (planetId ? fetchBuildingQueue(planetId) : Promise.resolve([])),
    enabled: planetId !== null,
  });

  const reportQ = useQuery({
    queryKey: planetId ? QK.resourceReport(planetId) : ['noop-rr-c'],
    queryFn: () => (planetId ? fetchResourceReport(planetId) : Promise.reject()),
    enabled: planetId !== null,
  });

  const enqueue = useMutation({
    mutationFn: (unitId: number) => enqueueBuilding(planetId!, unitId),
    onSuccess: () => {
      if (planetId) {
        void qc.invalidateQueries({ queryKey: QK.buildingQueue(planetId) });
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
      }
    },
  });

  if (!planetId) {
    return <div className="idiv">{t('overview', 'noPlanets')}</div>;
  }

  const queue = queueQ.data ?? [];
  const buildings = catalogByGroup('building');

  // Уровни из resource-report: unit_id → level
  const levelMap: Record<number, number> = {};
  for (const b of reportQ.data?.buildings ?? []) {
    levelMap[b.unit_id] = b.level;
  }

  return (
    <>
      {queue.length > 0 && (
        <table className="ntable">
          <tbody>
            <tr>
              <th colSpan={4}>{t('buildings', 'outstandingMissions')}</th>
            </tr>
            {queue.map((task, idx) => {
              const cat = buildings.find((b) => b.id === task.unit_id);
              const [g, k] = cat ? (cat.i18n.split('.') as [string, string]) : ['info', ''];
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
            const level = levelMap[entry.id] ?? 0;
            const descKey = `${key}Desc`;
            const desc = t(group, descKey);
            const hasDesc = !desc.startsWith('[');
            return (
              <tr key={entry.id}>
                <td width="1px" style={{ verticalAlign: 'top' }}>
                  <img
                    src={`/assets/origin/images/units/${entry.id}.gif`}
                    alt=""
                    onError={(e) => { (e.target as HTMLImageElement).style.display = 'none'; }}
                  />
                </td>
                <td style={{ verticalAlign: 'top' }}>
                  <div style={{ width: '100%' }}>
                    {level > 0 && (
                      <span style={{ float: 'right' }}>
                        {t('buildings', 'levelAbbr')} {level}
                      </span>
                    )}
                    {t(group, key)}
                  </div>
                  {hasDesc && (
                    <div style={{ clear: 'both', fontSize: 'smaller' }}>{desc}</div>
                  )}
                </td>
                <td width="100px" align="center" style={{ verticalAlign: 'top' }}>
                  <input
                    type="button"
                    className="button"
                    value={t('buildings', 'build')}
                    onClick={() => enqueue.mutate(entry.id)}
                    disabled={enqueue.isPending}
                  />
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </>
  );
}
