// S-002 Constructions — строительство зданий (план 72 Ф.2 Spring 1).
//
// Pixel-perfect зеркало legacy `constructions.tpl`:
//   1) Верхняя ntable «OUTSTANDING_MISSIONS» — текущая очередь стройки.
//   2) Нижняя ntable «CONSTRUCTIONS» — список доступных зданий с
//      кнопкой «Построить».
//
// Endpoints:
//   GET    /api/planets/{id}/buildings/queue
//   POST   /api/planets/{id}/buildings
//   DELETE /api/planets/{id}/buildings/queue/{taskId}
//
// На текущий момент откуда брать «уровни/стоимости/время» — не
// раскрыто публичным endpoint'ом (см. simplifications.md). Spring 1
// рендерит названия из CATALOG, без уровней — план 72 Ф.X добавит
// агрегированный GET /api/planets/{id}/buildings (список с уровнями).

import { useParams } from 'react-router-dom';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  cancelBuildingTask,
  enqueueBuilding,
  fetchBuildingQueue,
} from '@/api/buildings';
import { QK } from '@/api/query-keys';
import { useResolvedPlanet } from '@/features/common/useResolvedPlanet';
import { catalogByGroup } from '@/features/common/catalog';
import { useTranslation } from '@/i18n/i18n';
import { secondsUntil, formatDuration } from '@/lib/format';

export function ConstructionsScreen() {
  const { planetId: urlId } = useParams();
  const { planetId } = useResolvedPlanet(urlId);
  const { t } = useTranslation();
  const qc = useQueryClient();

  const queueQ = useQuery({
    queryKey: planetId ? QK.buildingQueue(planetId) : ['noop'],
    queryFn: () => (planetId ? fetchBuildingQueue(planetId) : Promise.resolve([])),
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

  return (
    <>
      {queue.length > 0 && (
        <table className="ntable">
          <thead>
            <tr>
              <th colSpan={4}>{t('info', 'eventModeConstruction')}</th>
            </tr>
          </thead>
          <tbody>
            {queue.map((task, idx) => (
              <tr key={task.id}>
                <td width="1px">{idx + 1}.</td>
                <td>
                  #{task.unit_id} → {task.target_level}
                </td>
                <td>{formatDuration(secondsUntil(task.end_at))}</td>
                <td>
                  <button
                    type="button"
                    className="button"
                    onClick={() => cancel.mutate(task.id)}
                    disabled={cancel.isPending}
                  >
                    {t('info', 'abort')}
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}

      <table className="ntable">
        <thead>
          <tr>
            <th colSpan={3}>{t('empire', 'groupOtherBuildings')}</th>
          </tr>
        </thead>
        <tbody>
          {buildings.map((entry) => {
            const [group, key] = entry.i18n.split('.') as [string, string];
            return (
              <tr key={entry.id}>
                <td width="1px">#{entry.id}</td>
                <td>{t(group, key)}</td>
                <td width="100px">
                  <button
                    type="button"
                    className="button"
                    onClick={() => enqueue.mutate(entry.id)}
                    disabled={enqueue.isPending}
                  >
                    {t('research', 'study')}
                  </button>
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </>
  );
}
