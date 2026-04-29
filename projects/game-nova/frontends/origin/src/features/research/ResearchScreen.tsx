// S-003 Research — исследования (план 72.1 финализация).
// Pixel-perfect клон legacy research.tpl.

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { fetchResearch, startResearch } from '@/api/research';
import { QK } from '@/api/query-keys';
import { useResolvedPlanet } from '@/features/common/useResolvedPlanet';
import { catalogByGroup } from '@/features/common/catalog';
import { useTranslation } from '@/i18n/i18n';
import { secondsUntil, formatDuration } from '@/lib/format';

export function ResearchScreen() {
  const { planetId } = useResolvedPlanet();
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
      if (planetId) void qc.invalidateQueries({ queryKey: QK.planet(planetId) });
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
          <tbody>
            <tr>
              <th colSpan={4}>{t('buildings', 'outstandingMissions')}</th>
            </tr>
            {queue.map((task, idx) => {
              const cat = techs.find((c) => c.id === task.unit_id);
              const [g, k] = cat ? (cat.i18n.split('.') as [string, string]) : ['info', ''];
              const name = cat ? t(g, k) : `#${task.unit_id}`;
              return (
                <tr key={task.id}>
                  <td width="1px">{idx + 1}.</td>
                  <td colSpan={2}>
                    {name}&nbsp;{task.target_level}
                  </td>
                  <td width="100px">{formatDuration(secondsUntil(task.end_at))}</td>
                </tr>
              );
            })}
          </tbody>
        </table>
      )}

      <table className="ntable">
        <tbody>
          <tr>
            <th colSpan={3}>{t('buildings', 'research') ?? 'Исследования'}</th>
          </tr>

          {techs.map((entry) => {
            const [group, key] = entry.i18n.split('.') as [string, string];
            const lvl = levels[String(entry.id)] ?? 0;
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
                    {lvl > 0 && (
                      <span style={{ float: 'right' }}>
                        {t('buildings', 'levelAbbr')} {lvl}
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
                    value={t('research', 'study') ?? 'Изучить'}
                    onClick={() => start.mutate(entry.id)}
                    disabled={start.isPending || queue.length > 0}
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
