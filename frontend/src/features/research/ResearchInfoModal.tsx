import { useEffect } from 'react';
import { RESEARCH, costForLevel, imageOf, formatNum } from '@/api/catalog';
import type { ResearchEntry } from '@/api/catalog';

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

function researchTimeSecs(r: ResearchEntry, level: number): number {
  const timeBase: Record<string, number> = {
    spyware: 240, computer_tech: 120, gun_tech: 300, shield_tech: 300,
    shell_tech: 240, energy_tech: 180, hyperspace_tech: 600,
    combustion_engine: 180, impulse_engine: 300, hyperspace_engine: 600,
    laser_tech: 240, ion_tech: 480, plasma_tech: 600,
    expo_tech: 480, ballistics_tech: 360, masking_tech: 360,
  };
  const base = timeBase[r.key] ?? 180;
  return Math.round(base * r.costFactor ** (level - 1));
}

const LEVELS_RANGE = 10;

export function ResearchInfoModal({ unitId, currentLevel, onClose }: Props) {
  const r = RESEARCH.find((x) => x.id === unitId);

  useEffect(() => {
    function onKey(e: KeyboardEvent) { if (e.key === 'Escape') onClose(); }
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [onClose]);

  if (!r) return null;

  const startLevel = Math.max(1, currentLevel - 2);
  const endLevel = startLevel + LEVELS_RANGE - 1;
  const rows = Array.from({ length: endLevel - startLevel + 1 }, (_, i) => startLevel + i);

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
        maxWidth: 560,
        width: '100%',
        maxHeight: '90vh',
        overflowY: 'auto',
        display: 'flex',
        flexDirection: 'column',
        gap: 16,
      }}>
        <div style={{ display: 'flex', gap: 12, alignItems: 'flex-start' }}>
          <img
            src={imageOf(r.key)} alt={r.name} width={56} height={56}
            style={{ imageRendering: 'pixelated', borderRadius: 6, background: 'rgba(0,0,0,0.3)', padding: 4, flexShrink: 0 }}
          />
          <div style={{ flex: 1, minWidth: 0 }}>
            <div style={{ fontSize: 16, fontWeight: 700, fontFamily: 'var(--ox-font)', marginBottom: 4 }}>{r.name}</div>
            <div style={{ fontSize: 12, color: 'var(--ox-fg-muted)', fontStyle: 'italic' }}>{r.benefit}</div>
          </div>
          <button type="button" className="btn-ghost btn-sm" onClick={onClose} style={{ fontSize: 16, padding: '2px 8px', flexShrink: 0 }}>✕</button>
        </div>

        <div style={{ overflowX: 'auto' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 12, fontFamily: 'var(--ox-mono)' }}>
            <thead>
              <tr style={{ borderBottom: '1px solid var(--ox-border)', color: 'var(--ox-fg-muted)', textAlign: 'right' }}>
                <th style={{ textAlign: 'left', padding: '4px 8px', fontWeight: 600 }}>Ур.</th>
                {r.costBase.metal > 0    && <th style={{ padding: '4px 8px' }}>🟠</th>}
                {r.costBase.silicon > 0  && <th style={{ padding: '4px 8px' }}>💎</th>}
                {r.costBase.hydrogen > 0 && <th style={{ padding: '4px 8px' }}>💧</th>}
                <th style={{ padding: '4px 8px' }}>⏱</th>
              </tr>
            </thead>
            <tbody>
              {rows.map((lvl) => {
                const cost = costForLevel(r.costBase, r.costFactor, lvl);
                const secs = researchTimeSecs(r, lvl);
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
                    {r.costBase.metal > 0    && <td style={{ padding: '4px 8px', textAlign: 'right' }}>{formatNum(cost.metal)}</td>}
                    {r.costBase.silicon > 0  && <td style={{ padding: '4px 8px', textAlign: 'right' }}>{formatNum(cost.silicon)}</td>}
                    {r.costBase.hydrogen > 0 && <td style={{ padding: '4px 8px', textAlign: 'right' }}>{formatNum(cost.hydrogen)}</td>}
                    <td style={{ padding: '4px 8px', textAlign: 'right' }}>{fmtSecs(secs)}</td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>

        <div style={{ fontSize: 11, color: 'var(--ox-fg-muted)' }}>
          Время указано без учёта уровня исследовательской лаборатории.
        </div>
      </div>
    </div>
  );
}
