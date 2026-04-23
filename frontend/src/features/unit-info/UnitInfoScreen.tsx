import { useQuery } from '@tanstack/react-query';
import { api } from '@/api/client';
import { BUILDINGS, MOON_BUILDINGS, RESEARCH, SHIPS, DEFENSE, costForLevel, imageOf, formatNum, fmtReqs, nameOf } from '@/api/catalog';

interface Props {
  kind: 'building' | 'research' | 'ship' | 'defense';
  unitId: number;
  currentLevel: number;
  planetId?: string;
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

export function UnitInfoScreen({ kind, unitId, currentLevel, planetId }: Props) {
  const buildingsQ = useQuery({
    queryKey: ['buildings', planetId],
    queryFn: () => api.get<{ build_seconds: Record<string, number> }>(`/api/planets/${planetId}/buildings`),
    enabled: kind === 'building' && planetId != null,
    staleTime: 30000,
  });

  if (kind === 'ship' || kind === 'defense') {
    return <CombatUnitInfo kind={kind} unitId={unitId} />;
  }

  const entry = kind === 'building'
    ? [...BUILDINGS, ...MOON_BUILDINGS].find((x) => x.id === unitId)
    : RESEARCH.find((x) => x.id === unitId);

  if (!entry) return null;

  const isBuilding = kind === 'building';
  const prodRate = isBuilding ? PRODUCTION_RATES[entry.key] : undefined;
  const requires = 'requires' in entry ? entry.requires : undefined;
  const realBuildSeconds = buildingsQ.data?.build_seconds;

  const startLevel = Math.max(1, currentLevel - 2);
  const endLevel = startLevel + LEVELS_RANGE - 1;
  const rows = Array.from({ length: endLevel - startLevel + 1 }, (_, i) => startLevel + i);

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      {/* Заголовок */}
      <div style={{ display: 'flex', gap: 14, alignItems: 'flex-start' }}>
        <img
          src={imageOf(entry.key)} alt={entry.name} width={128} height={128}
          style={{ imageRendering: 'pixelated', borderRadius: 8, background: 'rgba(0,0,0,0.3)', padding: 4, flexShrink: 0 }}
        />
        <div>
          <h2 style={{ margin: 0, fontSize: 20, fontFamily: 'var(--ox-font)', fontWeight: 700 }}>{entry.name}</h2>
          {entry.fullDesc && (
            <div style={{ fontSize: 13, color: 'var(--ox-fg-muted)', fontStyle: 'italic', marginTop: 4 }}>{entry.fullDesc}</div>
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
              const staticSecs = isBuilding
                ? buildTimeSecs(entry.key, entry.costFactor, lvl)
                : researchTimeSecs(entry.key, entry.costFactor, lvl);
              const secs = (isBuilding && lvl === currentLevel + 1 && realBuildSeconds?.[String(unitId)] != null)
                ? realBuildSeconds[String(unitId)]!
                : staticSecs;
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
        {isBuilding
          ? (realBuildSeconds ? 'Время следующего уровня с учётом фабрики роботов и нано-фабрики.' : 'Время указано без учёта фабрики роботов и нано-фабрики.')
          : 'Время указано без учёта уровня исследовательской лаборатории.'}
      </div>

    </div>
  );
}

// Базовое время постройки: (metal+silicon)/5000 * 2 * 3600 сек (при shipyard=1, без нанофабрики)
function combatBuildTimeSecs(metal: number, silicon: number): number {
  return Math.round(((metal + silicon) / 5000) * 2 * 3600);
}

function CombatUnitInfo({ kind, unitId }: { kind: 'ship' | 'defense'; unitId: number }) {
  const unitCatalog = kind === 'ship' ? SHIPS : DEFENSE;
  const allUnits = [...SHIPS, ...DEFENSE];
  const entry = unitCatalog.find((x) => x.id === unitId);
  if (!entry) return null;

  const c = entry.cost;
  const isShip = kind === 'ship';
  const structure = c ? c.metal + c.silicon : null;
  const buildTime = c ? combatBuildTimeSecs(c.metal, c.silicon) : null;

  // Есть ли различие между атакующим и обороняющимся режимами
  const hasDualMode = isShip && (
    (entry.attacker_front != null && entry.attacker_front !== entry.front) ||
    (entry.attacker_ballistics != null && entry.attacker_ballistics !== entry.ballistics) ||
    (entry.attacker_masking != null && entry.attacker_masking !== entry.masking)
  );

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      {/* Заголовок */}
      <div style={{ display: 'flex', gap: 14, alignItems: 'flex-start' }}>
        <img
          src={imageOf(entry.key)} alt={entry.name} width={128} height={128}
          style={{ imageRendering: 'pixelated', borderRadius: 8, background: 'rgba(0,0,0,0.3)', padding: 4, flexShrink: 0 }}
        />
        <div>
          <h2 style={{ margin: 0, fontSize: 20, fontFamily: 'var(--ox-font)', fontWeight: 700 }}>{entry.name}</h2>
          {entry.description && (
            <div style={{ fontSize: 13, color: 'var(--ox-fg-muted)', fontStyle: 'italic', marginTop: 4 }}>{entry.description}</div>
          )}
        </div>
      </div>

      {/* Требования */}
      {entry.requires && entry.requires.length > 0 && (
        <div className="ox-panel" style={{ padding: '10px 14px', fontSize: 13, color: 'var(--ox-fg-muted)' }}>
          🔒 Требуется: {fmtReqs(entry.requires)}
        </div>
      )}

      {/* Боевые характеристики */}
      <div className="ox-panel" style={{ padding: 0, overflowX: 'auto' }}>
        <div style={{ padding: '10px 14px 6px', fontSize: 11, fontWeight: 700, letterSpacing: '0.08em', textTransform: 'uppercase', color: 'var(--ox-fg-muted)' }}>
          Боевые характеристики
        </div>
        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
          <thead>
            <tr style={{ borderBottom: '1px solid var(--ox-border)', color: 'var(--ox-fg-muted)', fontSize: 12 }}>
              <th style={cellLeft}>Параметр</th>
              {hasDualMode ? <th style={cell}>В атаке</th> : null}
              <th style={cell}>{hasDualMode ? 'В обороне' : 'Значение'}</th>
            </tr>
          </thead>
          <tbody>
            {hasDualMode ? (
              <>
                <DualRow label="⚔ Атака"       atk={entry.attack.toLocaleString('ru-RU')}       def={entry.attack.toLocaleString('ru-RU')} />
                <DualRow label="🛡 Щит"         atk={entry.shield.toLocaleString('ru-RU')}       def={entry.shield.toLocaleString('ru-RU')} />
                <DualRow label="❤ Броня"        atk={entry.shell.toLocaleString('ru-RU')}        def={entry.shell.toLocaleString('ru-RU')} />
                {entry.front != null && <DualRow label="🎯 Приоритет цели" atk={String(entry.attacker_front ?? entry.front)} def={String(entry.front)} />}
                {entry.ballistics != null && <DualRow label="🎲 Баллистика" atk={String(entry.attacker_ballistics ?? entry.ballistics)} def={String(entry.ballistics)} />}
                {entry.masking != null && <DualRow label="👻 Маскировка" atk={String(entry.attacker_masking ?? entry.masking)} def={String(entry.masking)} />}
              </>
            ) : (
              <>
                <StatRow label="⚔ Атака"  value={entry.attack.toLocaleString('ru-RU')} />
                <StatRow label="🛡 Щит"    value={entry.shield.toLocaleString('ru-RU')} />
                <StatRow label="❤ Броня"   value={entry.shell.toLocaleString('ru-RU')} />
                {entry.front != null && <StatRow label="🎯 Приоритет цели" value={String(entry.front)} />}
                {entry.ballistics != null && entry.ballistics > 0 && <StatRow label="🎲 Баллистика" value={String(entry.ballistics)} />}
                {entry.masking != null && entry.masking > 0 && <StatRow label="👻 Маскировка" value={String(entry.masking)} />}
              </>
            )}
          </tbody>
        </table>
      </div>

      {/* Другие характеристики */}
      <div className="ox-panel" style={{ padding: 0, overflowX: 'auto' }}>
        <div style={{ padding: '10px 14px 6px', fontSize: 11, fontWeight: 700, letterSpacing: '0.08em', textTransform: 'uppercase', color: 'var(--ox-fg-muted)' }}>
          Другие характеристики
        </div>
        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
          <tbody>
            {entry.cargo != null && entry.cargo > 0 && (
              <StatRow label="📦 Грузоподъёмность" value={entry.cargo.toLocaleString('ru-RU')} />
            )}
            {entry.speed != null && entry.speed > 0 && (
              <StatRow label="🚀 Скорость" value={entry.speed.toLocaleString('ru-RU')} />
            )}
            {entry.fuel != null && entry.fuel > 0 && (
              <StatRow label="⛽ Расход топлива" value={`${entry.fuel}/ед.`} />
            )}
            {structure != null && structure > 0 && (
              <StatRow label="🔩 Конструкция" value={structure.toLocaleString('ru-RU')} />
            )}
            {c && c.metal > 0 && (
              <StatRow label="🟠 Металл" value={c.metal.toLocaleString('ru-RU')} />
            )}
            {c && c.silicon > 0 && (
              <StatRow label="💎 Кремний" value={c.silicon.toLocaleString('ru-RU')} />
            )}
            {c && c.hydrogen > 0 && (
              <StatRow label="💧 Водород" value={c.hydrogen.toLocaleString('ru-RU')} />
            )}
            {buildTime != null && buildTime > 0 && (
              <StatRow label="⏱ Время постройки (базовое)" value={fmtSecs(buildTime)} />
            )}
          </tbody>
        </table>
      </div>

      {/* Быстрый огонь */}
      {entry.rapidfire && Object.keys(entry.rapidfire).length > 0 && (
        <div className="ox-panel" style={{ padding: 0, overflowX: 'auto' }}>
          <div style={{ padding: '10px 14px 6px', fontSize: 11, fontWeight: 700, letterSpacing: '0.08em', textTransform: 'uppercase', color: 'var(--ox-fg-muted)' }}>
            Быстрый огонь (rapidfire)
          </div>
          <table style={{ width: '100%', borderCollapse: 'collapse' }}>
            <thead>
              <tr style={{ borderBottom: '1px solid var(--ox-border)', color: 'var(--ox-fg-muted)', fontSize: 12 }}>
                <th style={cellLeft}>Цель</th>
                <th style={cell}>Выстрелов за раунд</th>
              </tr>
            </thead>
            <tbody>
              {Object.entries(entry.rapidfire).map(([targetId, shots]) => (
                <tr key={targetId} style={{ borderBottom: '1px solid rgba(255,255,255,0.04)' }}>
                  <td style={{ ...cellLeft, color: 'var(--ox-fg)' }}>
                    {nameOf(Number(targetId))}
                  </td>
                  <td style={{ ...cell, color: 'var(--ox-accent)', fontWeight: 600 }}>{shots}×</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Кто стреляет быстро по этому юниту */}
      {(() => {
        const shooters = allUnits.filter((u) => u.rapidfire && u.rapidfire[entry.id]);
        if (shooters.length === 0) return null;
        return (
          <div className="ox-panel" style={{ padding: 0, overflowX: 'auto' }}>
            <div style={{ padding: '10px 14px 6px', fontSize: 11, fontWeight: 700, letterSpacing: '0.08em', textTransform: 'uppercase', color: 'var(--ox-fg-muted)' }}>
              Уязвим к быстрому огню
            </div>
            <table style={{ width: '100%', borderCollapse: 'collapse' }}>
              <thead>
                <tr style={{ borderBottom: '1px solid var(--ox-border)', color: 'var(--ox-fg-muted)', fontSize: 12 }}>
                  <th style={cellLeft}>Атакующий</th>
                  <th style={cell}>Выстрелов за раунд</th>
                </tr>
              </thead>
              <tbody>
                {shooters.map((shooter) => (
                  <tr key={shooter.id} style={{ borderBottom: '1px solid rgba(255,255,255,0.04)' }}>
                    <td style={{ ...cellLeft, color: 'var(--ox-fg)' }}>{shooter.name}</td>
                    <td style={{ ...cell, color: 'var(--ox-danger)', fontWeight: 600 }}>{shooter.rapidfire![entry.id]}×</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        );
      })()}
    </div>
  );
}

function StatRow({ label, value }: { label: string; value: string }) {
  return (
    <tr style={{ borderBottom: '1px solid rgba(255,255,255,0.04)' }}>
      <td style={{ ...cellLeft, color: 'var(--ox-fg-muted)', fontSize: 13 }}>{label}</td>
      <td style={{ ...cell, color: 'var(--ox-fg)', fontSize: 13 }}>{value}</td>
    </tr>
  );
}

function DualRow({ label, atk, def }: { label: string; atk: string; def: string }) {
  return (
    <tr style={{ borderBottom: '1px solid rgba(255,255,255,0.04)' }}>
      <td style={{ ...cellLeft, color: 'var(--ox-fg-muted)', fontSize: 13 }}>{label}</td>
      <td style={{ ...cell, color: 'var(--ox-fg)', fontSize: 13 }}>{atk}</td>
      <td style={{ ...cell, color: 'var(--ox-fg)', fontSize: 13 }}>{def}</td>
    </tr>
  );
}
