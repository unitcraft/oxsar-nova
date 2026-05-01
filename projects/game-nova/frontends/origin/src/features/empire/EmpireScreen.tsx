// S-007 Empire — обзор всех планет игрока (план 72.1.37).
//
// Pixel-perfect зеркало legacy `empire.tpl`: верхняя таблица
// планет + 5 вкладок (constructions/shipyard/defense/moon/research)
// с агрегатами по планетам в столбцах.

import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { fetchEmpire, type EmpirePlanet } from '@/api/empire';
import { useCurrentPlanetStore } from '@/stores/currentPlanet';
import { formatNumber, formatCoords } from '@/lib/format';
import { useTranslation } from '@/i18n/i18n';
import { catalogByGroup } from '@/features/common/catalog';

type Tab = 'overview' | 'buildings' | 'ships' | 'defense' | 'research';

export function EmpireScreen() {
  const setCurrent = useCurrentPlanetStore((s) => s.set);
  const navigate = useNavigate();
  const { t } = useTranslation();
  const [tab, setTab] = useState<Tab>('overview');

  const q = useQuery({
    queryKey: ['empire'],
    queryFn: fetchEmpire,
    staleTime: 30_000,
  });

  if (q.isLoading) return <div className="idiv">…</div>;
  if (!q.data || q.data.planets.length === 0) {
    return <div className="idiv">{t('overview', 'noPlanets')}</div>;
  }
  const { planets, research } = q.data;
  // Луны в hidden — отдельная вкладка `moon` (legacy).
  const planetsOnly = planets.filter((p) => !p.is_moon);
  const moons = planets.filter((p) => p.is_moon);

  return (
    <>
      {/* Tab switcher */}
      <table className="ntable">
        <tbody>
          <tr>
            {(['overview', 'buildings', 'ships', 'defense', 'research'] as Tab[]).map(
              (key) => (
                <td
                  key={key}
                  className="center"
                  style={{
                    background: tab === key ? '#444' : undefined,
                    cursor: 'pointer',
                  }}
                  onClick={() => setTab(key)}
                >
                  {t('empire', `tab_${key}`) || key}
                </td>
              ),
            )}
          </tr>
        </tbody>
      </table>

      {tab === 'overview' && (
        <OverviewTable planets={planets} setCurrent={setCurrent} navigate={navigate} t={t} />
      )}
      {tab === 'buildings' && (
        <UnitTable
          planets={planetsOnly}
          group="building"
          getCount={(p, id) => p.buildings[String(id)] ?? 0}
          t={t}
        />
      )}
      {tab === 'ships' && (
        <UnitTable
          planets={planets}
          group="ship"
          getCount={(p, id) => p.ships[String(id)] ?? 0}
          t={t}
        />
      )}
      {tab === 'defense' && (
        <UnitTable
          planets={planetsOnly}
          group="defense"
          getCount={(p, id) => p.defense[String(id)] ?? 0}
          t={t}
        />
      )}
      {tab === 'research' && <ResearchBlock research={research} t={t} />}

      {moons.length > 0 && tab === 'overview' && (
        <table className="ntable">
          <thead>
            <tr>
              <th colSpan={4}>{t('empire', 'tab_moon') || 'Луны'}</th>
            </tr>
          </thead>
          <tbody>
            {moons.map((m, idx) => (
              <tr key={m.id}>
                <td>{idx + 1}.</td>
                <td>
                  {m.name} <br />
                  {formatCoords(m.galaxy, m.system, m.position) + ' L'}
                </td>
                <td align="right">{formatNumber(m.diameter)} км</td>
                <td align="right">
                  {m.used_fields}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </>
  );
}

type TFunc = (group: string, key: string, vars?: Record<string, string>) => string;

function OverviewTable({
  planets,
  setCurrent,
  navigate,
  t,
}: {
  planets: EmpirePlanet[];
  setCurrent: (id: string) => void;
  navigate: (path: string) => void;
  t: TFunc;
}) {
  return (
    <table className="ntable">
      <thead>
        <tr>
          <th>№</th>
          <th>{t('empire', 'groupPlanet')}</th>
          <th>{t('empire', 'rowDiameter')}</th>
          <th>{t('empire', 'rowFields')}</th>
          <th>{t('empire', 'rowTemp')}</th>
          {/* План 72.1.45: УМИ (research_virt_lab). */}
          <th>{t('empire', 'umi') || 'УМИ'}</th>
          <th colSpan={2}>{t('empire', 'groupResources')}</th>
        </tr>
      </thead>
      <tbody>
        {planets.map((p, idx) => (
          <tr key={p.id}>
            <td>{idx + 1}.</td>
            <td>
              <button
                type="button"
                className="link-button"
                onClick={() => {
                  setCurrent(p.id);
                  navigate('/');
                }}
                aria-label={`${p.name} ${formatCoords(p.galaxy, p.system, p.position)}`}
              >
                {p.name}
                {p.is_moon && ' 🌙'}
                <br />
                {formatCoords(p.galaxy, p.system, p.position) + (p.is_moon ? ' L' : '')}
              </button>
            </td>
            <td align="right">{formatNumber(p.diameter)} км</td>
            <td align="right">{p.used_fields}</td>
            <td align="right">
              {p.temp_min}…{p.temp_max} °C
            </td>
            <td align="right">{formatNumber(p.umi)}</td>
            <td>
              М <br />К <br />В
            </td>
            <td align="right">
              {formatNumber(p.metal)}
              <br />
              {formatNumber(p.silicon)}
              <br />
              {formatNumber(p.hydrogen)}
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}

function UnitTable({
  planets,
  group,
  getCount,
  t,
}: {
  planets: EmpirePlanet[];
  group: 'building' | 'ship' | 'defense';
  getCount: (p: EmpirePlanet, id: number) => number;
  t: TFunc;
}) {
  const units = catalogByGroup(group);
  return (
    <table className="ntable">
      <thead>
        <tr>
          <th>{t('empire', `tab_${group}`) || group}</th>
          {planets.map((p) => (
            <th key={p.id}>
              {p.name}
              <br />
              <small>{formatCoords(p.galaxy, p.system, p.position) + (p.is_moon ? ' L' : '')}</small>
            </th>
          ))}
        </tr>
      </thead>
      <tbody>
        {units.map((u) => {
          const [g, k] = u.i18n.split('.') as [string, string];
          const counts = planets.map((p) => getCount(p, u.id));
          // Скрываем строки где у всех планет 0.
          if (counts.every((c) => c === 0)) return null;
          return (
            <tr key={u.id}>
              <td>{t(g, k)}</td>
              {counts.map((c, i) => (
                <td key={planets[i]!.id} align="right">
                  {c > 0 ? formatNumber(c) : ''}
                </td>
              ))}
            </tr>
          );
        })}
      </tbody>
    </table>
  );
}

function ResearchBlock({
  research,
  t,
}: {
  research: Record<string, number>;
  t: TFunc;
}) {
  const techs = catalogByGroup('research');
  return (
    <table className="ntable">
      <thead>
        <tr>
          <th>{t('empire', 'tab_research') || 'Исследования'}</th>
          <th>{t('techtree', 'levelAbbr') || 'Ур.'}</th>
        </tr>
      </thead>
      <tbody>
        {techs.map((tech) => {
          const lvl = research[String(tech.id)] ?? 0;
          if (lvl === 0) return null;
          const [g, k] = tech.i18n.split('.') as [string, string];
          return (
            <tr key={tech.id}>
              <td>{t(g, k)}</td>
              <td align="right">{lvl}</td>
            </tr>
          );
        })}
      </tbody>
    </table>
  );
}
