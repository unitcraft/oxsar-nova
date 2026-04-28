import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '@/api/client';
import { BUILDINGS, SHIPS, DEFENSE, buildingName, nameOf } from '@/api/catalog';
import { useTranslation } from '@/i18n/i18n';

interface EmpirePlanet {
  id: string;
  name: string;
  galaxy: number;
  system: number;
  position: number;
  is_moon: boolean;
  diameter: number;
  used_fields: number;
  temp_min: number;
  temp_max: number;
  metal: number;
  silicon: number;
  hydrogen: number;
  buildings: Record<string, number>;
  ships: Record<string, number>;
  defense: Record<string, number>;
}

const BUILDING_GROUPS: Array<{ tkey: string; ids: number[] }> = [
  { tkey: 'groupMining',    ids: [1, 2, 3] },
  { tkey: 'groupEnergy',    ids: [4, 5, 6] },
  { tkey: 'groupStorages',  ids: [7, 8, 9] },
  { tkey: 'groupProduction', ids: [10, 11, 12, 14, 15] },
  { tkey: 'groupSpecial',   ids: [21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31] },
];

const ALL_BUILDING_IDS = BUILDINGS.map((b) => b.id);
const GROUPED_IDS = new Set(BUILDING_GROUPS.flatMap((g) => g.ids));
const OTHER_BUILDING_IDS = ALL_BUILDING_IDS.filter((id) => !GROUPED_IDS.has(id));

const ALL_SHIP_IDS = SHIPS.map((s) => s.id);
const ALL_DEFENSE_IDS = DEFENSE.map((d) => d.id);

function fmtRes(v: number): string {
  if (v >= 1_000_000) return `${(v / 1_000_000).toFixed(1)}M`;
  if (v >= 1_000) return `${Math.round(v / 1_000)}k`;
  return String(Math.round(v));
}

function LevelCell({ level, maxLevel = 40 }: { level: number; maxLevel?: number }) {
  if (level === 0) return <td style={{ textAlign: 'center', color: 'var(--ox-fg-muted)', fontSize: 13 }}>—</td>;
  const isMax = level >= maxLevel;
  return (
    <td style={{
      textAlign: 'center',
      fontFamily: 'var(--ox-mono)',
      fontSize: 14,
      fontWeight: 600,
      color: isMax ? 'var(--ox-success, #22c55e)' : 'var(--ox-fg)',
      background: isMax ? 'rgba(34,197,94,0.07)' : undefined,
    }}>
      {level}{isMax && <span style={{ fontSize: 9, marginLeft: 2 }}>MAX</span>}
    </td>
  );
}

function CountCell({ count }: { count: number }) {
  if (count === 0) return <td style={{ textAlign: 'center', color: 'var(--ox-fg-muted)', fontSize: 13 }}>—</td>;
  return (
    <td style={{ textAlign: 'center', fontFamily: 'var(--ox-mono)', fontSize: 14 }}>
      {fmtRes(count)}
    </td>
  );
}

function SectionHeader({
  label, colSpan, collapsed, onToggle,
}: {
  label: string; colSpan: number; collapsed: boolean; onToggle: () => void;
}) {
  return (
    <tr
      onClick={onToggle}
      style={{ cursor: 'pointer', userSelect: 'none', background: 'var(--ox-bg-card)' }}
    >
      <td colSpan={colSpan} style={{
        padding: '6px 10px',
        fontSize: 13, fontWeight: 700, letterSpacing: '0.08em',
        textTransform: 'uppercase', color: 'var(--ox-fg-dim)',
        borderTop: '1px solid var(--ox-border)',
      }}>
        {collapsed ? '▶' : '▼'} {label}
      </td>
    </tr>
  );
}

export function EmpireScreen() {
  const { t } = useTranslation('empire');
  const { t: ti } = useTranslation('info');
  const { data, isLoading } = useQuery({
    queryKey: ['empire'],
    queryFn: () => api.get<{ planets: EmpirePlanet[] }>('/api/empire'),
    refetchInterval: 60000,
  });

  const [collapsedGroups, setCollapsedGroups] = useState<Set<string>>(new Set());

  function toggleGroup(key: string) {
    setCollapsedGroups((prev) => {
      const next = new Set(prev);
      if (next.has(key)) next.delete(key);
      else next.add(key);
      return next;
    });
  }

  if (isLoading) {
    return (
      <div style={{ padding: 24 }}>
        <div className="ox-skeleton" style={{ height: 400, borderRadius: 8 }} />
      </div>
    );
  }

  const planets = data?.planets ?? [];
  if (planets.length === 0) {
    return <div style={{ padding: 24, color: 'var(--ox-fg-dim)' }}>{t('noPlanets')}</div>;
  }

  const colCount = planets.length + 1;

  const totalMetal = planets.reduce((s, p) => s + p.metal, 0);
  const totalSilicon = planets.reduce((s, p) => s + p.silicon, 0);
  const totalHydrogen = planets.reduce((s, p) => s + p.hydrogen, 0);

  const usedBuildingIds = new Set<number>();
  for (const p of planets) {
    for (const id of Object.keys(p.buildings)) usedBuildingIds.add(Number(id));
  }
  const usedShipIds = new Set<number>();
  for (const p of planets) {
    for (const id of Object.keys(p.ships)) usedShipIds.add(Number(id));
  }
  const usedDefenseIds = new Set<number>();
  for (const p of planets) {
    for (const id of Object.keys(p.defense)) usedDefenseIds.add(Number(id));
  }

  const thStyle: React.CSSProperties = {
    padding: '8px 10px', fontFamily: 'var(--ox-mono)', fontSize: 13,
    fontWeight: 700, color: 'var(--ox-fg-dim)', whiteSpace: 'nowrap',
    background: 'var(--ox-bg-card)', position: 'sticky', top: 0, zIndex: 1,
    borderBottom: '1px solid var(--ox-border)',
  };
  const rowLabelStyle: React.CSSProperties = {
    padding: '4px 10px', fontSize: 14, color: 'var(--ox-fg-dim)',
    whiteSpace: 'nowrap', position: 'sticky', left: 0, zIndex: 1,
    background: 'var(--ox-bg)', borderRight: '1px solid var(--ox-border)',
    minWidth: 160,
  };

  function renderBuildingRows(ids: number[]) {
    return ids
      .filter((id) => usedBuildingIds.has(id))
      .map((id) => (
        <tr key={id}>
          <td style={rowLabelStyle}>{buildingName(id, ti)}</td>
          {planets.map((p) => (
            <LevelCell key={p.id} level={p.buildings[String(id)] ?? 0} />
          ))}
        </tr>
      ));
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16, padding: '16px 0' }}>
      <h2 style={{ margin: 0, fontSize: 18, fontWeight: 700 }}>{t('title')}</h2>

      <div className="ox-panel" style={{ padding: '10px 16px', display: 'flex', gap: 24, flexWrap: 'wrap', fontSize: 15 }}>
        <span style={{ color: 'var(--ox-fg-dim)' }}>{t('totalResources')}</span>
        <span>🟠 <b style={{ fontFamily: 'var(--ox-mono)' }}>{fmtRes(totalMetal)}</b></span>
        <span>💎 <b style={{ fontFamily: 'var(--ox-mono)' }}>{fmtRes(totalSilicon)}</b></span>
        <span>💧 <b style={{ fontFamily: 'var(--ox-mono)' }}>{fmtRes(totalHydrogen)}</b></span>
      </div>

      <div style={{ overflowX: 'auto', borderRadius: 8, border: '1px solid var(--ox-border)' }}>
        <table style={{ borderCollapse: 'collapse', width: '100%', minWidth: planets.length * 90 + 160 }}>
          <thead>
            <tr>
              <th style={{ ...thStyle, position: 'sticky', left: 0, zIndex: 2 }}>{t('colParam')}</th>
              {planets.map((p) => (
                <th key={p.id} style={{ ...thStyle, textAlign: 'center', maxWidth: 90 }}>
                  <div>{p.is_moon ? '🌑' : '🪐'} {p.name}</div>
                  <div style={{ fontWeight: 400, fontSize: 10, color: 'var(--ox-fg-muted)' }}>
                    [{p.galaxy}:{p.system}:{p.position}]
                  </div>
                </th>
              ))}
            </tr>
          </thead>

          <tbody>
            <SectionHeader label={t('groupPlanet')} colSpan={colCount} collapsed={collapsedGroups.has('planet')} onToggle={() => toggleGroup('planet')} />
            {!collapsedGroups.has('planet') && <>
              <tr>
                <td style={rowLabelStyle}>{t('rowDiameter')}</td>
                {planets.map((p) => (
                  <td key={p.id} style={{ textAlign: 'center', fontSize: 13, fontFamily: 'var(--ox-mono)' }}>
                    {p.diameter.toLocaleString('ru-RU')}
                  </td>
                ))}
              </tr>
              <tr>
                <td style={rowLabelStyle}>{t('rowFields')}</td>
                {planets.map((p) => (
                  <td key={p.id} style={{ textAlign: 'center', fontSize: 13, fontFamily: 'var(--ox-mono)' }}>
                    {p.used_fields}
                  </td>
                ))}
              </tr>
              <tr>
                <td style={rowLabelStyle}>{t('rowTemp')}</td>
                {planets.map((p) => (
                  <td key={p.id} style={{ textAlign: 'center', fontSize: 13 }}>
                    {p.temp_min}…{p.temp_max}°C
                  </td>
                ))}
              </tr>
            </>}

            <SectionHeader label={t('groupResources')} colSpan={colCount} collapsed={collapsedGroups.has('res')} onToggle={() => toggleGroup('res')} />
            {!collapsedGroups.has('res') && <>
              {([t('rowMetal'), t('rowSilicon'), t('rowHydrogen')] as const).map((label, i) => (
                <tr key={label}>
                  <td style={rowLabelStyle}>{label}</td>
                  {planets.map((p) => (
                    <td key={p.id} style={{ textAlign: 'center', fontSize: 13, fontFamily: 'var(--ox-mono)' }}>
                      {fmtRes([p.metal, p.silicon, p.hydrogen][i]!)}
                    </td>
                  ))}
                </tr>
              ))}
            </>}

            {BUILDING_GROUPS.map((group) => {
              const visibleIds = group.ids.filter((id) => usedBuildingIds.has(id));
              if (visibleIds.length === 0) return null;
              const key = `bg-${group.tkey}`;
              return (
                <>
                  <SectionHeader key={`hdr-${key}`} label={t(group.tkey)} colSpan={colCount} collapsed={collapsedGroups.has(key)} onToggle={() => toggleGroup(key)} />
                  {!collapsedGroups.has(key) && renderBuildingRows(visibleIds)}
                </>
              );
            })}

            {(() => {
              const visibleOther = OTHER_BUILDING_IDS.filter((id) => usedBuildingIds.has(id));
              if (visibleOther.length === 0) return null;
              return (
                <>
                  <SectionHeader label={t('groupOtherBuildings')} colSpan={colCount} collapsed={collapsedGroups.has('other-b')} onToggle={() => toggleGroup('other-b')} />
                  {!collapsedGroups.has('other-b') && renderBuildingRows(visibleOther)}
                </>
              );
            })()}

            {usedShipIds.size > 0 && <>
              <SectionHeader label={t('groupFleet')} colSpan={colCount} collapsed={collapsedGroups.has('ships')} onToggle={() => toggleGroup('ships')} />
              {!collapsedGroups.has('ships') && ALL_SHIP_IDS.filter((id) => usedShipIds.has(id)).map((id) => (
                <tr key={id}>
                  <td style={rowLabelStyle}>{nameOf(id, ti)}</td>
                  {planets.map((p) => <CountCell key={p.id} count={p.ships[String(id)] ?? 0} />)}
                </tr>
              ))}
            </>}

            {usedDefenseIds.size > 0 && <>
              <SectionHeader label={t('groupDefense')} colSpan={colCount} collapsed={collapsedGroups.has('defense')} onToggle={() => toggleGroup('defense')} />
              {!collapsedGroups.has('defense') && ALL_DEFENSE_IDS.filter((id) => usedDefenseIds.has(id)).map((id) => (
                <tr key={id}>
                  <td style={rowLabelStyle}>{nameOf(id, ti)}</td>
                  {planets.map((p) => <CountCell key={p.id} count={p.defense[String(id)] ?? 0} />)}
                </tr>
              ))}
            </>}
          </tbody>
        </table>
      </div>
    </div>
  );
}
