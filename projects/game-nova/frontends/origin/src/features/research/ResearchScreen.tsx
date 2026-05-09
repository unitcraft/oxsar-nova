// S-003 Research — исследования (план 72.1 ч.20).
// Pixel-perfect клон legacy research.tpl + required_res_table.tpl.

import { Link } from 'react-router-dom';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  fetchResearch,
  startResearch,
  cancelResearch,
  startResearchVIP,
} from '@/api/research';
import { QK } from '@/api/query-keys';
import { useResolvedPlanet } from '@/features/common/useResolvedPlanet';
import { useAutoInvalidateOnTaskEnd } from '@/features/common/useAutoInvalidateOnTaskEnd';
import { catalogByGroup } from '@/features/common/catalog';
import { RequiredResTable } from '@/features/common/RequiredResTable';
import { ConfirmDialog, useConfirm } from '@/features/common/ConfirmDialog';
import { VipButton } from '@/features/common/VipButton';
import { ConstructionProgress } from '@/features/common/ConstructionProgress';
import { fetchSettings, updateSettings } from '@/api/settings';
import { useTranslation } from '@/i18n/i18n';
import { formatDuration } from '@/lib/format';

export function ResearchScreen() {
  const { planetId, planet } = useResolvedPlanet();
  const { t } = useTranslation();
  const qc = useQueryClient();
  // План 72.1.53 ч.B: in-game confirm-dialog.
  const { confirm, dialogProps } = useConfirm();

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

  // План 72.1.39: cancel research-задачи.
  const cancelMut = useMutation({
    mutationFn: (queueId: string) => cancelResearch(queueId),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: QK.research() });
      if (planetId) void qc.invalidateQueries({ queryKey: QK.planet(planetId) });
    },
  });

  // План 72.1.44: VIP-instant старт research за credits.
  const vipMut = useMutation({
    mutationFn: (queueId: string) => startResearchVIP(queueId),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: QK.research() });
      void qc.invalidateQueries({ queryKey: QK.me() });
    },
  });

  // Pack-research реализуется только на info-странице (UnitInfoScreen)
  // — 1:1 с legacy `constructions.tpl`, где `{var}ext_pack_research{/var}`
  // показывается только при `info_id > 0`.

  // План 72.1.59: авто-инвалидация очереди по окончании ближайшей
  // задачи (общий хук). Без него экран остаётся со старым snapshot'ом
  // когда прогресс-бар доезжает до 100%. QK.planets() инвалидируем
  // тоже — research может изменить production/storage (через
  // research-эффекты), и тик ресурсов в TopHeader должен сменить темп.
  const queue = overviewQ.data?.queue ?? [];
  useAutoInvalidateOnTaskEnd(queue, () => {
    void qc.invalidateQueries({ queryKey: QK.research() });
    void qc.invalidateQueries({ queryKey: QK.planets() });
    if (planetId) void qc.invalidateQueries({ queryKey: QK.planet(planetId) });
  });

  if (!planetId) {
    return <div className="idiv">{t('overview', 'noPlanets')}</div>;
  }

  const levels = overviewQ.data?.levels ?? {};
  const addedLevels = overviewQ.data?.added_levels ?? {};
  const seconds = overviewQ.data?.research_seconds ?? {};
  const costs = overviewQ.data?.research_costs ?? {};
  const allTechs = catalogByGroup('research');
  const apiOrder = overviewQ.data?.order;
  const techs = apiOrder
    ? apiOrder.flatMap((id) => { const t = allTechs.find((x) => x.id === id); return t ? [t] : []; })
    : allTechs;
  // План 72.1.55.E (effects): show_all_research preference.
  const settingsQ = useQuery({
    queryKey: QK.settings(),
    queryFn: fetchSettings,
    staleTime: 60_000,
  });
  const showAll = settingsQ.data?.show_all_research ?? true;
  const toggleShowAll = useMutation({
    mutationFn: (next: boolean) => updateSettings({ show_all_research: next }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: QK.settings() });
    },
  });
  // Filter: если выкл и tech 0-level + cost == 0 (= unmet/max),
  // скрываем. Backend research/handler не возвращает unmet явно;
  // cost=0 — proxy для «недоступно».
  const visibleTechs = techs.filter((entry) => {
    if (showAll) return true;
    const lvl = levels[String(entry.id)] ?? 0;
    if (lvl > 0) return true;
    const c = costs[String(entry.id)];
    if (!c) return false;
    return c.metal > 0 || c.silicon > 0 || c.hydrogen > 0;
  });

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
              <th colSpan={5}>{t('buildings', 'outstandingMissions')}</th>
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
                  <td width="130px">
                    <ConstructionProgress startAt={task.start_at} endAt={task.end_at} />
                  </td>
                  {/* План 72.1.39: cancel-кнопка (legacy Research::abort). */}
                  <td width="60px" align="center">
                    <button
                      type="button"
                      className="button"
                      disabled={cancelMut.isPending}
                      title={t('buildings', 'cancelTask')}
                      onClick={async () => {
                        if (await confirm({
                          title: t('buildings', 'cancelTask'),
                          message: t('buildings', 'cancelConfirm'),
                          destructive: true,
                        })) {
                          cancelMut.mutate(task.id);
                        }
                      }}
                    >
                      ✕
                    </button>
                  </td>
                  {/* План 72.1.44: VIP-instant старт за credits. */}
                  <td width="80px" align="center">
                    <VipButton
                      taskId={task.id}
                      endAt={task.end_at}
                      onVip={(id) => vipMut.mutate(id)}
                      isPending={vipMut.isPending}
                      label="⚡"
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
            <th colSpan={3}>{t('buildings', 'research')}</th>
          </tr>
          <tr>
            <th colSpan={2} style={{ textAlign: 'right' }}>
              <label htmlFor="show_all_research_cb">
                <strong>{t('global', 'showUnavailable')}</strong>
              </label>{' '}
              <input
                type="checkbox"
                id="show_all_research_cb"
                checked={showAll}
                disabled={toggleShowAll.isPending}
                onChange={(e) => toggleShowAll.mutate(e.target.checked)}
              />
            </th>
            <th>&nbsp;</th>
          </tr>

          {visibleTechs.map((entry) => {
            const [group, key] = entry.i18n.split('.') as [string, string];
            const lvl = levels[String(entry.id)] ?? 0;
            const added = addedLevels[String(entry.id)] ?? 0;
            const secs = seconds[String(entry.id)] ?? 0;
            const cost = costs[String(entry.id)] ?? { metal: 0, silicon: 0, hydrogen: 0 };
            const descKey = `${key}Desc`;
            const desc = t(group, descKey);
            const hasDesc = !desc.startsWith('[');
            const enough = canBuild(entry.id);
            return (
              <tr key={entry.id}>
                <td width="1px" style={{ verticalAlign: 'top' }}>
                  <Link to={`/unit/${entry.id}`}>
                    <img
                      src={`/assets/origin/images/units/${entry.icon}.gif`}
                      alt={t(group, key)}
                      onError={(e) => { (e.target as HTMLImageElement).style.display = 'none'; }}
                    />
                  </Link>
                </td>
                <td style={{ verticalAlign: 'top', textAlign: 'left' }}>
                  <div style={{ width: '100%' }}>
                    <span style={{ float: 'right' }}>
                      {t('research', 'level', { n: String(lvl) })}
                      {added !== 0 && (
                        <span className={added > 0 ? 'true' : 'false'}>
                          {' '}({added > 0 ? '+' : ''}{added})
                        </span>
                      )}
                    </span>
                    <strong>
                      <Link to={`/unit/${entry.id}`}>{t(group, key)}</Link>
                    </strong>
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
                      {t('buildings', 'buildingAtWork')}
                    </span>
                  ) : (
                    <button
                      type="button"
                      className={`btn-link ${enough ? 'true' : 'false'}`}
                      onClick={() => start.mutate(entry.id)}
                      disabled={start.isPending || !enough}
                    >
                      {t('buildings', 'researchOfLevel')} {lvl + 1}
                    </button>
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
