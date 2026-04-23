import { useEffect } from 'react';
import { BUILDINGS, MOON_BUILDINGS, costForLevel, imageOf, formatNum } from '@/api/catalog';
import type { BuildingEntry } from '@/api/catalog';

interface Props {
  unitId: number;
  currentLevel: number;
  onClose: () => void;
}

function fmtSecs(secs: number): string {
  if (secs < 60) return `${secs}с`;
  const m = Math.floor(secs / 60) % 60;
  const h = Math.floor(secs / 3600) % 24;
  const d = Math.floor(secs / 86400);
  if (d > 0) return `${d}д ${h}ч ${m}м`;
  if (h > 0) return `${h}ч ${m}м`;
  return `${m}м`;
}

function buildTimeSecs(b: BuildingEntry, level: number): number {
  const timeBase: Record<string, number> = {
    metal_mine: 60, silicon_lab: 75, hydrogen_lab: 90, solar_plant: 60,
    hydrogen_plant: 90, robotic_factory: 180, nano_factory: 600, shipyard: 180,
    metal_storage: 120, silicon_storage: 120, hydrogen_storage: 120,
    research_lab: 180, missile_silo: 360, repair_factory: 30,
    moon_base: 300, star_surveillance: 600, star_gate: 1800, moon_robotic_factory: 240,
  };
  const base = timeBase[b.key] ?? 120;
  return Math.round(base * b.costFactor ** (level - 1));
}

// Производство в час по уровню (формулы из legacy economy/production.go)
const PRODUCTION_RATES: Record<string, { base: number; label: string }> = {
  metal_mine:      { base: 30,   label: '🟠/ч' },
  silicon_lab:     { base: 20,   label: '💎/ч' },
  hydrogen_lab:    { base: 10,   label: '💧/ч' },
  solar_plant:     { base: 20,   label: '⚡/ч' },
  hydrogen_plant:  { base: 22.5, label: '⚡/ч' },
};

function productionAtLevel(key: string, level: number): number | null {
  const r = PRODUCTION_RATES[key];
  if (!r) return null;
  return Math.floor(r.base * level * 1.1 ** level);
}

const LEVELS_RANGE = 10;

export function BuildingInfoModal({ unitId, currentLevel, onClose }: Props) {
  const b = [...BUILDINGS, ...MOON_BUILDINGS].find((x) => x.id === unitId);

  useEffect(() => {
    function onKey(e: KeyboardEvent) { if (e.key === 'Escape') onClose(); }
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [onClose]);

  if (!b) return null;

  const startLevel = Math.max(1, currentLevel - 2);
  const endLevel = startLevel + LEVELS_RANGE - 1;
  const rows = Array.from({ length: endLevel - startLevel + 1 }, (_, i) => startLevel + i);
  const prodRate = PRODUCTION_RATES[b.key];

  return (
    <div
      style={{
        position: 'fixed', inset: 0, zIndex: 1000,
        background: 'rgba(0,0,0,0.6)', backdropFilter: 'blur(4px)',
        display: 'flex', alignItems: 'center', justifyContent: 'center',
        padding: 16,
      }}
      onClick={(e) => { if (e.target === e.currentTarget) onClose(); }}
    >
      <div style={{
        background: 'var(--ox-bg-panel)',
        border: '1px solid var(--ox-border)',
        borderRadius: 'var(--ox-r-lg)',
        padding: 20,
        maxWidth: 600,
        width: '100%',
        maxHeight: '90vh',
        overflowY: 'auto',
        display: 'flex',
        flexDirection: 'column',
        gap: 16,
      }}>
        {/* Header */}
        <div style={{ display: 'flex', gap: 12, alignItems: 'flex-start' }}>
          <img
            src={imageOf(b.key)} alt={b.name} width={56} height={56}
            style={{ imageRendering: 'pixelated', borderRadius: 6, background: 'rgba(0,0,0,0.3)', padding: 4, flexShrink: 0 }}
          />
          <div style={{ flex: 1, minWidth: 0 }}>
            <div style={{ fontSize: 16, fontWeight: 700, fontFamily: 'var(--ox-font)', marginBottom: 4 }}>{b.name}</div>
            {b.description && (
              <div style={{ fontSize: 12, color: 'var(--ox-fg-muted)', fontStyle: 'italic' }}>{b.description}</div>
            )}
          </div>
          <button
            type="button" className="btn-ghost btn-sm"
            onClick={onClose}
            style={{ fontSize: 16, padding: '2px 8px', flexShrink: 0 }}
          >✕</button>
        </div>

        {/* Table */}
        <div style={{ overflowX: 'auto' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 12, fontFamily: 'var(--ox-mono)' }}>
            <thead>
              <tr style={{ borderBottom: '1px solid var(--ox-border)', color: 'var(--ox-fg-muted)', textAlign: 'right' }}>
                <th style={{ textAlign: 'left', padding: '4px 8px', fontWeight: 600 }}>Ур.</th>
                {b.costBase.metal > 0    && <th style={{ padding: '4px 8px' }}>🟠</th>}
                {b.costBase.silicon > 0  && <th style={{ padding: '4px 8px' }}>💎</th>}
                {b.costBase.hydrogen > 0 && <th style={{ padding: '4px 8px' }}>💧</th>}
                <th style={{ padding: '4px 8px' }}>⏱</th>
                {prodRate && <th style={{ padding: '4px 8px' }}>{prodRate.label}</th>}
              </tr>
            </thead>
            <tbody>
              {rows.map((lvl) => {
                const cost = costForLevel(b.costBase, b.costFactor, lvl);
                const secs = buildTimeSecs(b, lvl);
                const prod = productionAtLevel(b.key, lvl);
                const isCurrent = lvl === currentLevel;
                const isNext = lvl === currentLevel + 1;
                return (
                  <tr
                    key={lvl}
                    style={{
                      borderBottom: '1px solid rgba(255,255,255,0.04)',
                      background: isCurrent ? 'rgba(56,189,248,0.08)' : isNext ? 'rgba(56,189,248,0.04)' : undefined,
                      color: isCurrent ? 'var(--ox-accent)' : isNext ? 'var(--ox-fg)' : 'var(--ox-fg-dim)',
                    }}
                  >
                    <td style={{ padding: '4px 8px', fontWeight: isCurrent || isNext ? 700 : 400 }}>
                      {lvl}{isCurrent ? ' ←' : isNext ? ' →' : ''}
                    </td>
                    {b.costBase.metal > 0    && <td style={{ padding: '4px 8px', textAlign: 'right' }}>{formatNum(cost.metal)}</td>}
                    {b.costBase.silicon > 0  && <td style={{ padding: '4px 8px', textAlign: 'right' }}>{formatNum(cost.silicon)}</td>}
                    {b.costBase.hydrogen > 0 && <td style={{ padding: '4px 8px', textAlign: 'right' }}>{formatNum(cost.hydrogen)}</td>}
                    <td style={{ padding: '4px 8px', textAlign: 'right' }}>{fmtSecs(secs)}</td>
                    {prod !== null && <td style={{ padding: '4px 8px', textAlign: 'right' }}>{formatNum(prod)}</td>}
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>

        <div style={{ fontSize: 11, color: 'var(--ox-fg-muted)' }}>
          Время указано без учёта фабрики роботов и нано-фабрики.
        </div>

        {/* Полное описание */}
        {b.fullDesc && (
          <details style={{ fontSize: 12, color: 'var(--ox-fg-dim)' }}>
            <summary style={{ cursor: 'pointer', color: 'var(--ox-fg-muted)', userSelect: 'none', marginBottom: 6 }}>Подробнее</summary>
            <div style={{ lineHeight: 1.6, paddingTop: 4 }}>{b.fullDesc}</div>
          </details>
        )}
      </div>
    </div>
  );
}
