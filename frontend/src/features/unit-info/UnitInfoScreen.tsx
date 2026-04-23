import { BUILDINGS, MOON_BUILDINGS, RESEARCH, costForLevel, imageOf, formatNum, fmtReqs } from '@/api/catalog';

interface Props {
  kind: 'building' | 'research';
  unitId: number;
  currentLevel: number;
  onBack: () => void;
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

function buildTimeSecs(key: string, costFactor: number, level: number): number {
  const timeBase: Record<string, number> = {
    metal_mine: 60, silicon_lab: 75, hydrogen_lab: 90, solar_plant: 60,
    hydrogen_plant: 90, robotic_factory: 180, nano_factory: 600, shipyard: 180,
    metal_storage: 120, silicon_storage: 120, hydrogen_storage: 120,
    research_lab: 180, missile_silo: 360, repair_factory: 30,
    moon_base: 300, star_surveillance: 600, star_gate: 1800, moon_robotic_factory: 240,
  };
  return Math.round((timeBase[key] ?? 120) * costFactor ** (level - 1));
}

function researchTimeSecs(key: string, costFactor: number, level: number): number {
  const timeBase: Record<string, number> = {
    spyware: 240, computer_tech: 120, gun_tech: 300, shield_tech: 300,
    shell_tech: 240, energy_tech: 180, hyperspace_tech: 600,
    combustion_engine: 180, impulse_engine: 300, hyperspace_engine: 600,
    laser_tech: 240, ion_tech: 480, plasma_tech: 600,
    expo_tech: 480, ballistics_tech: 360, masking_tech: 360,
  };
  return Math.round((timeBase[key] ?? 180) * costFactor ** (level - 1));
}

const PRODUCTION_RATES: Record<string, { base: number; label: string }> = {
  metal_mine:     { base: 30,   label: '🟠/ч' },
  silicon_lab:    { base: 20,   label: '💎/ч' },
  hydrogen_lab:   { base: 10,   label: '💧/ч' },
  solar_plant:    { base: 20,   label: '⚡/ч' },
  hydrogen_plant: { base: 22.5, label: '⚡/ч' },
};

function productionAtLevel(key: string, level: number): number | null {
  const r = PRODUCTION_RATES[key];
  if (!r) return null;
  return Math.floor(r.base * level * 1.1 ** level);
}

const LEVELS_RANGE = 10;

const cell: React.CSSProperties = { padding: '6px 12px', textAlign: 'right', fontFamily: 'var(--ox-mono)', fontSize: 13 };
const cellLeft: React.CSSProperties = { ...cell, textAlign: 'left' };

export function UnitInfoScreen({ kind, unitId, currentLevel, onBack }: Props) {
  const entry = kind === 'building'
    ? [...BUILDINGS, ...MOON_BUILDINGS].find((x) => x.id === unitId)
    : RESEARCH.find((x) => x.id === unitId);

  if (!entry) return null;

  const isBuilding = kind === 'building';
  const prodRate = isBuilding ? PRODUCTION_RATES[entry.key] : undefined;
  const requires = 'requires' in entry ? entry.requires : undefined;

  const startLevel = Math.max(1, currentLevel - 2);
  const endLevel = startLevel + LEVELS_RANGE - 1;
  const rows = Array.from({ length: endLevel - startLevel + 1 }, (_, i) => startLevel + i);

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      {/* Кнопка назад */}
      <div>
        <button type="button" className="btn-ghost btn-sm" onClick={onBack} style={{ fontSize: 13 }}>
          ← Назад
        </button>
      </div>

      {/* Заголовок */}
      <div style={{ display: 'flex', gap: 14, alignItems: 'flex-start' }}>
        <img
          src={imageOf(entry.key)} alt={entry.name} width={64} height={64}
          style={{ imageRendering: 'pixelated', borderRadius: 8, background: 'rgba(0,0,0,0.3)', padding: 4, flexShrink: 0 }}
        />
        <div>
          <h2 style={{ margin: 0, fontSize: 20, fontFamily: 'var(--ox-font)', fontWeight: 700 }}>{entry.name}</h2>
          {'description' in entry && entry.description && (
            <div style={{ fontSize: 13, color: 'var(--ox-fg-muted)', fontStyle: 'italic', marginTop: 4 }}>{entry.description}</div>
          )}
          {'benefit' in entry && (
            <div style={{ fontSize: 13, color: 'var(--ox-fg-muted)', fontStyle: 'italic', marginTop: 4 }}>{entry.benefit}</div>
          )}
          {currentLevel > 0 && (
            <div style={{ fontSize: 13, color: 'var(--ox-accent)', marginTop: 4, fontFamily: 'var(--ox-mono)' }}>Уровень {currentLevel}</div>
          )}
        </div>
      </div>

      {/* Пререквизиты */}
      {requires && requires.length > 0 && (
        <div className="ox-panel" style={{ padding: '10px 14px', fontSize: 13, color: 'var(--ox-fg-muted)' }}>
          🔒 Требуется: {fmtReqs(requires)}
        </div>
      )}

      {/* Таблица уровней */}
      <div className="ox-panel" style={{ padding: 0, overflowX: 'auto' }}>
        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
          <thead>
            <tr style={{ borderBottom: '1px solid var(--ox-border)', color: 'var(--ox-fg-muted)', fontSize: 12 }}>
              <th style={cellLeft}>Ур.</th>
              {entry.costBase.metal > 0    && <th style={cell}>🟠</th>}
              {entry.costBase.silicon > 0  && <th style={cell}>💎</th>}
              {entry.costBase.hydrogen > 0 && <th style={cell}>💧</th>}
              <th style={cell}>⏱</th>
              {prodRate && <th style={cell}>{prodRate.label}</th>}
            </tr>
          </thead>
          <tbody>
            {rows.map((lvl) => {
              const cost = costForLevel(entry.costBase, entry.costFactor, lvl);
              const secs = isBuilding
                ? buildTimeSecs(entry.key, entry.costFactor, lvl)
                : researchTimeSecs(entry.key, entry.costFactor, lvl);
              const prod = productionAtLevel(entry.key, lvl);
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
                  <td style={{ ...cellLeft, fontWeight: isCurrent || isNext ? 700 : 400 }}>
                    {lvl}{isCurrent ? ' ←' : isNext ? ' →' : ''}
                  </td>
                  {entry.costBase.metal > 0    && <td style={cell}>{formatNum(cost.metal)}</td>}
                  {entry.costBase.silicon > 0  && <td style={cell}>{formatNum(cost.silicon)}</td>}
                  {entry.costBase.hydrogen > 0 && <td style={cell}>{formatNum(cost.hydrogen)}</td>}
                  <td style={cell}>{fmtSecs(secs)}</td>
                  {prod !== null && <td style={cell}>{formatNum(prod)}</td>}
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>

      <div style={{ fontSize: 11, color: 'var(--ox-fg-muted)' }}>
        {isBuilding ? 'Время указано без учёта фабрики роботов и нано-фабрики.' : 'Время указано без учёта уровня исследовательской лаборатории.'}
      </div>

      {/* Полное описание */}
      {entry.fullDesc && (
        <div className="ox-panel" style={{ padding: '14px 16px' }}>
          <div style={{ fontSize: 12, fontWeight: 600, color: 'var(--ox-fg-muted)', marginBottom: 8, textTransform: 'uppercase', letterSpacing: '0.06em' }}>Описание</div>
          <div style={{ fontSize: 13, color: 'var(--ox-fg-dim)', lineHeight: 1.7 }}>{entry.fullDesc}</div>
        </div>
      )}
    </div>
  );
}
