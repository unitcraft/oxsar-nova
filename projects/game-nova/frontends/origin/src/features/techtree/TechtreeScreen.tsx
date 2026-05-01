// S-021 Techtree (план 72 Ф.4 Spring 3).
//
// Pixel-perfect зеркало legacy `templates/standard/techtree.tpl` —
// табличное представление дерева технологий: секции
// строения / исследования / верфь / оборона / лунные постройки,
// для каждого юнита — текущий уровень игрока + список requirements
// с пометкой met/unmet.
//
// Endpoint:
//   GET /api/techtree[?planet_id=...]  → { nodes: TechtreeNode[] }

import { Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { fetchTechtree } from '@/api/catalog';
import { QK } from '@/api/query-keys';
import type { TechtreeNode, TechtreeRequirement } from '@/api/types';
import { useTranslation } from '@/i18n/i18n';
import { findCatalog } from '@/features/common/catalog';

interface SectionDef {
  kind: TechtreeNode['kind'];
  titleKey: string;
  routePrefix: '/building' | '/unit';
  // План 72.1.22: для секции 'building' разделяем на наземные/лунные.
  filter?: (n: TechtreeNode) => boolean;
}

const SECTIONS: SectionDef[] = [
  {
    kind: 'building',
    titleKey: 'kindBuildings',
    routePrefix: '/building',
    filter: (n) => !n.moon_only,
  },
  {
    kind: 'building',
    titleKey: 'kindMoonBuildings',
    routePrefix: '/building',
    filter: (n) => !!n.moon_only,
  },
  { kind: 'research', titleKey: 'kindResearch', routePrefix: '/unit' },
  { kind: 'ship', titleKey: 'kindShips', routePrefix: '/unit' },
  { kind: 'defense', titleKey: 'kindDefense', routePrefix: '/unit' },
];

export function TechtreeScreen() {
  const { t } = useTranslation();
  const q = useQuery({
    queryKey: QK.techtree(),
    queryFn: () => fetchTechtree(),
    staleTime: 60_000,
  });

  if (q.isLoading) return <div className="idiv">…</div>;

  const nodes = q.data?.nodes ?? [];

  return (
    <table className="ntable">
      <thead>
        <tr>
          <th colSpan={3}>{t('techtree', 'title')}</th>
        </tr>
      </thead>
      <tbody>
        {SECTIONS.map((section, sectionIdx) => {
          const items = nodes.filter(
            (n) => n.kind === section.kind && (!section.filter || section.filter(n)),
          );
          if (items.length === 0) return null;
          return (
            <SectionRows
              key={`${section.kind}-${sectionIdx}`}
              titleKey={section.titleKey}
              items={items}
              routePrefix={section.routePrefix}
            />
          );
        })}
      </tbody>
    </table>
  );
}

interface SectionRowsProps {
  titleKey: string;
  items: TechtreeNode[];
  routePrefix: '/building' | '/unit';
}

function SectionRows({ titleKey, items, routePrefix }: SectionRowsProps) {
  const { t } = useTranslation();
  // Стабильная сортировка: разблокированные сверху, дальше — по id.
  const sorted = [...items].sort((a, b) => {
    if (a.unlocked !== b.unlocked) return a.unlocked ? -1 : 1;
    return a.id - b.id;
  });
  return (
    <>
      <tr>
        <th colSpan={2}>{t('techtree', titleKey)}</th>
        <th>{t('alliance', 'rank') || 'Уровень / Требования'}</th>
      </tr>
      {sorted.map((n) => {
        const i18nKey = n.key.replace(/_([a-z])/g, (_, c: string) =>
          c.toUpperCase(),
        );
        const name = t('info', i18nKey);
        const display = name !== `[info.${i18nKey}]` ? name : n.key;
        return (
          <tr key={`${n.kind}-${n.id}`}>
            <td style={{ width: '1%' }}>
              {findCatalog(n.id) ? (
                <img
                  src={`/assets/origin/images/units/${findCatalog(n.id)!.icon}.gif`}
                  alt={display}
                  width={32}
                  height={32}
                  onError={(e) => {
                    (e.target as HTMLImageElement).style.display = 'none';
                  }}
                />
              ) : (
                `#${n.id}`
              )}
            </td>
            <td>
              <Link to={`${routePrefix}/${n.id}`}>{display}</Link>
            </td>
            <td>
              {n.kind === 'building' || n.kind === 'research' ? (
                <span className={n.current_level > 0 ? 'true' : ''}>
                  {t('techtree', 'levelAbbr')} {n.current_level}
                </span>
              ) : n.unlocked ? (
                <span className="true">{t('techtree', 'available')}</span>
              ) : (
                <span className="false">{t('techtree', 'locked')}</span>
              )}
              {n.requirements.length > 0 && (
                <RequirementsList requirements={n.requirements} />
              )}
            </td>
          </tr>
        );
      })}
    </>
  );
}

function RequirementsList({
  requirements,
}: {
  requirements: TechtreeRequirement[];
}) {
  const { t } = useTranslation();
  return (
    <div style={{ fontSize: 'smaller', marginTop: 4 }}>
      {requirements.map((r, i) => {
        const i18nKey = r.key.replace(/_([a-z])/g, (_, c: string) =>
          c.toUpperCase(),
        );
        const name = t('info', i18nKey);
        const display = name !== `[info.${i18nKey}]` ? name : r.key;
        return (
          <div key={`${r.kind}-${r.key}-${i}`}>
            <span className={r.met ? 'true' : 'false'}>
              {display} {t('techtree', 'levelAbbr')} {r.level}
              {' ('}
              {r.have}/{r.level}
              {')'}
            </span>
          </div>
        );
      })}
    </div>
  );
}
