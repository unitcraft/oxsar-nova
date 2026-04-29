// S-003 Research — исследования (план 72.1 ч.20).
// Pixel-perfect клон legacy research.tpl + required_res_table.tpl.

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { fetchResearch, startResearch } from '@/api/research';
import { QK } from '@/api/query-keys';
import { useResolvedPlanet } from '@/features/common/useResolvedPlanet';
import { catalogByGroup } from '@/features/common/catalog';
import { RequiredResTable } from '@/features/common/RequiredResTable';
import { useTranslation } from '@/i18n/i18n';
import { secondsUntil, formatDuration } from '@/lib/format';

export function ResearchScreen() {
  const { planetId, planet } = useResolvedPlanet();
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
  const seconds = overviewQ.data?.research_seconds ?? {};
  const costs = overviewQ.data?.research_costs ?? {};
  const techs = catalogByGroup('research');

  const available = planet
    ? { metal: Math.floor(planet.metal), silicon: Math.floor(planet.silicon), hydrogen: Math.floor(planet.hydrogen) }
    : { metal: 0, silicon: 0, hydrogen: 0 };

  function canBuild(unitId: number): boolean {
    const c = costs[String(unitId)];
    if (!c) return false;
    return available.metal >= c.metal && available.silicon >= c.silicon && available.hydrogen >= c.hydrogen;
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
            const secs = seconds[String(entry.id)] ?? 0;
            const cost = costs[String(entry.id)] ?? { metal: 0, silicon: 0, hydrogen: 0 };
            const descKey = `${key}Desc`;
            const desc = t(group, descKey);
            const hasDesc = !desc.startsWith('[');
            const enough = canBuild(entry.id);
            return (
              <tr key={entry.id}>
                <td width="1px" style={{ verticalAlign: 'top' }}>
                  <img
                    src={`/assets/origin/images/units/${entry.icon}.gif`}
                    alt={t(group, key)}
                    onError={(e) => { (e.target as HTMLImageElement).style.display = 'none'; }}
                  />
                </td>
                <td style={{ verticalAlign: 'top' }}>
                  <div style={{ width: '100%' }}>
                    <span style={{ float: 'right' }}>
                      Уровень {lvl}
                    </span>
                    {t(group, key)}
                  </div>
                  {hasDesc && (
                    <div style={{ clear: 'both', fontSize: 'smaller' }}>{desc}</div>
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
                  {queue.length > 0 ? (
                    <span className="false">
                      {t('buildings', 'buildingAtWork') ?? 'Занято'}
                    </span>
                  ) : (
                    <button
                      type="button"
                      className={`btn-link ${enough ? 'true' : 'false'}`}
                      onClick={() => start.mutate(entry.id)}
                      disabled={start.isPending || !enough}
                    >
                      {t('buildings', 'researchOfLevel') ?? 'Исследовать'}<br />
                      {t('buildings', 'levelAbbr') === 'Ур.' ? 'уровень' : 'уровень'} {lvl + 1}
                    </button>
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
