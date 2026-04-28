// S-003 Research — исследования (план 72 Ф.2 Spring 1).
//
// Pixel-perfect зеркало legacy `research.tpl`:
//   1) Верхняя ntable очередь, если активная.
//   2) Нижняя ntable список технологий из CATALOG (group=research).
//
// Endpoints:
//   GET  /api/research                — очередь + уровни (агрегировано)
//   POST /api/planets/{id}/research   — поставить исследование

import { useParams } from 'react-router-dom';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { fetchResearch, startResearch } from '@/api/research';
import { QK } from '@/api/query-keys';
import { useResolvedPlanet } from '@/features/common/useResolvedPlanet';
import { catalogByGroup } from '@/features/common/catalog';
import { useTranslation } from '@/i18n/i18n';
import { secondsUntil, formatDuration } from '@/lib/format';

export function ResearchScreen() {
  const { planetId: urlId } = useParams();
  const { planetId } = useResolvedPlanet(urlId);
  const { t } = useTranslation();
  const qc = useQueryClient();

  const overviewQ = useQuery({
    queryKey: QK.research(),
    queryFn: fetchResearch,
  });

  const start = useMutation({
    mutationFn: (unitId: number) => startResearch(planetId!, unitId),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: QK.research() });
      if (planetId) {
        void qc.invalidateQueries({ queryKey: QK.planet(planetId) });
      }
    },
  });

  if (!planetId) {
    return <div className="idiv">{t('overview', 'noPlanets')}</div>;
  }

  const queue = overviewQ.data?.queue ?? [];
  const levels = overviewQ.data?.levels ?? {};
  const techs = catalogByGroup('research');

  return (
    <>
      {queue.length > 0 && (
        <table className="ntable">
          <thead>
            <tr>
              <th colSpan={4}>{t('info', 'eventModeResearch')}</th>
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
              </tr>
            ))}
          </tbody>
        </table>
      )}

      <table className="ntable">
        <thead>
          <tr>
            <th colSpan={3}>{t('research', 'title')}</th>
          </tr>
        </thead>
        <tbody>
          {techs.map((entry) => {
            const [group, key] = entry.i18n.split('.') as [string, string];
            const lvl = levels[String(entry.id)] ?? 0;
            return (
              <tr key={entry.id}>
                <td width="1px">#{entry.id}</td>
                <td>
                  {t(group, key)}{' '}
                  <span className="false">
                    ({t('research', 'level', { n: lvl })})
                  </span>
                </td>
                <td width="100px">
                  <button
                    type="button"
                    className="button"
                    onClick={() => start.mutate(entry.id)}
                    disabled={start.isPending || queue.length > 0}
                    aria-label={t('research', 'study')}
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
