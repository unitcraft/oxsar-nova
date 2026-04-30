// S-046 Simulator — боевой симулятор (план 72.1 ч.20.7).
// Pixel-perfect клон legacy simulator.tpl + ADR-0002 порт
// oxsar2-java/Assault.java (rendering на frontend).
//
// Структура (legacy):
// 1. Технологии атакующего/защитника (tech-уровни 4 шт.).
// 2. Таблица юнитов: ship-name / attacker quantity / defender quantity.
// 3. Кнопка «Симулировать».
// 4. Результат: победитель + потери + раунды.
//
// Используется backend POST /api/simulator/run (порт battle.Calculate).

import { useState } from 'react';
import { useMutation } from '@tanstack/react-query';
import {
  runSimulation,
  type SimInput,
  type SimReport,
  type SimSide,
  type SimUnit,
} from '@/api/simulator';
import { catalogByGroup } from '@/features/common/catalog';
import { useTranslation } from '@/i18n/i18n';
import { formatNumber } from '@/lib/format';

interface UnitStats {
  attack: number;
  shield: number;
  shell: number;
}

// Минимальные дефолтные статы для тестового симулятора.
// В legacy эти значения берутся из БД construction (с tech-модификаторами).
// Здесь — упрощённо для UI; backend Calculate сам применит tech.
const UNIT_STATS: Record<number, UnitStats> = {
  // ships
  29: { attack: 5, shield: 10, shell: 4000 },        // small_transporter
  30: { attack: 5, shield: 25, shell: 12000 },       // large_transporter
  31: { attack: 50, shield: 10, shell: 4000 },       // light_fighter
  32: { attack: 150, shield: 25, shell: 10000 },     // strong_fighter
  33: { attack: 400, shield: 50, shell: 27000 },     // cruiser
  34: { attack: 1000, shield: 200, shell: 60000 },   // battle_ship
  35: { attack: 700, shield: 400, shell: 70000 },    // frigate
  36: { attack: 50, shield: 100, shell: 30000 },     // colony_ship
  37: { attack: 1, shield: 10, shell: 16000 },       // recycler
  38: { attack: 0, shield: 0.01, shell: 1000 },      // espionage_sensor
  39: { attack: 1, shield: 1, shell: 2000 },         // solar_satellite
  40: { attack: 1000, shield: 500, shell: 75000 },   // bomber
  41: { attack: 2000, shield: 500, shell: 110000 },  // star_destroyer
  42: { attack: 200000, shield: 50000, shell: 9000000 }, // death_star
  // defense
  43: { attack: 80, shield: 20, shell: 2000 },       // rocket_launcher
  44: { attack: 100, shield: 25, shell: 2000 },      // light_laser
  45: { attack: 250, shield: 100, shell: 8000 },     // strong_laser
  46: { attack: 150, shield: 500, shell: 8000 },     // ion_gun
  47: { attack: 1100, shield: 200, shell: 35000 },   // gauss_gun
  48: { attack: 3000, shield: 300, shell: 100000 },  // plasma_gun
  49: { attack: 1, shield: 2000, shell: 20000 },     // small_shield
  50: { attack: 1, shield: 10000, shell: 100000 },   // large_shield
};

interface TechSet {
  gun: number;
  shield: number;
  shell: number;
  laser: number;
  ion: number;
  plasma: number;
}

const ZERO_TECH: TechSet = { gun: 0, shield: 0, shell: 0, laser: 0, ion: 0, plasma: 0 };

export function SimulatorScreen() {
  const { t } = useTranslation();
  const ships = catalogByGroup('ship');
  const defense = catalogByGroup('defense');

  const [aTech, setATech] = useState<TechSet>({ ...ZERO_TECH });
  const [dTech, setDTech] = useState<TechSet>({ ...ZERO_TECH });
  const [aQuantities, setAQuantities] = useState<Record<number, string>>({});
  const [dQuantities, setDQuantities] = useState<Record<number, string>>({});

  const sim = useMutation<SimReport, Error, SimInput>({
    mutationFn: runSimulation,
  });

  function buildUnits(quantities: Record<number, string>): SimUnit[] {
    const out: SimUnit[] = [];
    for (const idStr of Object.keys(quantities)) {
      const id = Number(idStr);
      const qty = Math.max(0, Math.floor(Number(quantities[id]) || 0));
      if (qty <= 0) continue;
      const stats = UNIT_STATS[id];
      if (!stats) continue;
      out.push({
        unit_id: id,
        quantity: qty,
        attack: stats.attack,
        shield: stats.shield,
        shell: stats.shell,
      });
    }
    return out;
  }

  function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    const attackerUnits = buildUnits(aQuantities);
    const defenderUnits = buildUnits(dQuantities);
    if (attackerUnits.length === 0 || defenderUnits.length === 0) return;

    const attackers: SimSide[] = [
      { user_id: 'sim-attacker', tech: aTech, units: attackerUnits },
    ];
    const defenders: SimSide[] = [
      { user_id: 'sim-defender', tech: dTech, units: defenderUnits },
    ];
    sim.mutate({ attackers, defenders, rounds: 6 });
  }

  function entryName(id: number, group: 'ship' | 'defense'): string {
    const cat = catalogByGroup(group).find((c) => c.id === id);
    if (!cat) return `#${id}`;
    const [g, k] = cat.i18n.split('.') as [string, string];
    return t(g, k);
  }

  return (
    <form onSubmit={onSubmit}>
      {/* Tech levels */}
      <table className="ntable">
        <thead>
          <tr>
            <th colSpan={7}>{t('mission', 'simulator') ?? 'Симулятор боя'}</th>
          </tr>
          <tr>
            <th>&nbsp;</th>
            <th>{t('info', 'gunTech') ?? 'Оружейная'}</th>
            <th>{t('info', 'shieldTech') ?? 'Щитовая'}</th>
            <th>{t('info', 'shellTech') ?? 'Броневая'}</th>
            <th>{t('info', 'laserTech') ?? 'Лазерная'}</th>
            <th>{t('info', 'ionTech') ?? 'Ионная'}</th>
            <th>{t('info', 'plasmaTech') ?? 'Плазменная'}</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>{t('mission', 'attacker') ?? 'Атакующий'}</td>
            {(['gun', 'shield', 'shell', 'laser', 'ion', 'plasma'] as const).map((k) => (
              <td key={`a-${k}`}>
                <input
                  type="text"
                  size={2}
                  maxLength={2}
                  value={aTech[k]}
                  onChange={(e) =>
                    setATech({ ...aTech, [k]: Math.max(0, Number(e.target.value) || 0) })
                  }
                />
              </td>
            ))}
          </tr>
          <tr>
            <td>{t('mission', 'defender') ?? 'Защитник'}</td>
            {(['gun', 'shield', 'shell', 'laser', 'ion', 'plasma'] as const).map((k) => (
              <td key={`d-${k}`}>
                <input
                  type="text"
                  size={2}
                  maxLength={2}
                  value={dTech[k]}
                  onChange={(e) =>
                    setDTech({ ...dTech, [k]: Math.max(0, Number(e.target.value) || 0) })
                  }
                />
              </td>
            ))}
          </tr>
        </tbody>
      </table>

      {/* Units table */}
      <table className="ntable">
        <thead>
          <tr>
            <th>{t('mission', 'shipName') ?? 'Юнит'}</th>
            <th>{t('mission', 'attacker') ?? 'Атакующий'}</th>
            <th>{t('mission', 'defender') ?? 'Защитник'}</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <th colSpan={3} className="center">{t('shipyard', 'tabShips') ?? 'Флот'}</th>
          </tr>
          {ships.map((entry) => (
            <tr key={`ship-${entry.id}`}>
              <td>{entryName(entry.id, 'ship')}</td>
              <td className="center">
                <input
                  type="text"
                  size={8}
                  value={aQuantities[entry.id] ?? ''}
                  onChange={(e) =>
                    setAQuantities({ ...aQuantities, [entry.id]: e.target.value })
                  }
                />
              </td>
              <td className="center">
                <input
                  type="text"
                  size={8}
                  value={dQuantities[entry.id] ?? ''}
                  onChange={(e) =>
                    setDQuantities({ ...dQuantities, [entry.id]: e.target.value })
                  }
                />
              </td>
            </tr>
          ))}

          <tr>
            <th colSpan={3} className="center">{t('shipyard', 'tabDefense') ?? 'Оборона'}</th>
          </tr>
          {defense.map((entry) => (
            <tr key={`def-${entry.id}`}>
              <td>{entryName(entry.id, 'defense')}</td>
              <td className="center">—</td>
              <td className="center">
                <input
                  type="text"
                  size={8}
                  value={dQuantities[entry.id] ?? ''}
                  onChange={(e) =>
                    setDQuantities({ ...dQuantities, [entry.id]: e.target.value })
                  }
                />
              </td>
            </tr>
          ))}
        </tbody>
        <tfoot>
          <tr>
            <td colSpan={3} className="center">
              <input
                type="submit"
                className="button"
                value={t('mission', 'simulate') ?? 'Симулировать'}
                disabled={sim.isPending}
              />
            </td>
          </tr>
        </tfoot>
      </table>

      {/* Report */}
      {sim.data && <SimReportView report={sim.data} />}
    </form>
  );
}

function SimReportView({ report }: { report: SimReport }) {
  const { t } = useTranslation();
  const winnerLabel =
    report.winner === 'attackers'
      ? t('mission', 'attackerWins') ?? 'Атакующий победил'
      : report.winner === 'defenders'
        ? t('mission', 'defenderWins') ?? 'Защитник победил'
        : t('mission', 'draw') ?? 'Ничья';
  return (
    <table className="ntable" style={{ marginTop: 16 }}>
      <thead>
        <tr>
          <th colSpan={4}>
            {t('mission', 'simulationResult') ?? 'Результат симуляции'} —{' '}
            <span className={report.winner === 'attackers' ? 'true' : 'false'}>
              {winnerLabel}
            </span>
          </th>
        </tr>
        <tr>
          <th>{t('mission', 'roundCount') ?? 'Раундов'}: {report.rounds}</th>
          <th colSpan={3}>&nbsp;</th>
        </tr>
      </thead>
      <tbody>
        {/* Round-by-round */}
        {(report.rounds_trace ?? []).map((rt) => (
          <tr key={rt.index}>
            <td>{t('mission', 'round') ?? 'Раунд'} {rt.index + 1}</td>
            <td colSpan={3}>
              {t('mission', 'attackersAlive') ?? 'Атакующие'}:{' '}
              <b>{formatNumber(rt.attackers_alive)}</b>{' '}·{' '}
              {t('mission', 'defendersAlive') ?? 'Защитники'}:{' '}
              <b>{formatNumber(rt.defenders_alive)}</b>
            </td>
          </tr>
        ))}

        {/* Losses */}
        <tr>
          <th colSpan={4}>{t('mission', 'losses') ?? 'Потери'}</th>
        </tr>
        {(report.attackers ?? []).map((s) => (
          <tr key={`a-loss-${s.user_id}`}>
            <td>{t('mission', 'attacker') ?? 'Атакующий'}</td>
            <td>М: {formatNumber(s.lost_metal)}</td>
            <td>К: {formatNumber(s.lost_silicon)}</td>
            <td>В: {formatNumber(s.lost_hydrogen)}</td>
          </tr>
        ))}
        {(report.defenders ?? []).map((s) => (
          <tr key={`d-loss-${s.user_id}`}>
            <td>{t('mission', 'defender') ?? 'Защитник'}</td>
            <td>М: {formatNumber(s.lost_metal)}</td>
            <td>К: {formatNumber(s.lost_silicon)}</td>
            <td>В: {formatNumber(s.lost_hydrogen)}</td>
          </tr>
        ))}

        {/* Debris */}
        {(report.debris_metal ?? 0) > 0 && (
          <>
            <tr>
              <th colSpan={4}>{t('mission', 'debris') ?? 'Поле обломков'}</th>
            </tr>
            <tr>
              <td colSpan={2}>М: {formatNumber(report.debris_metal ?? 0)}</td>
              <td colSpan={2}>К: {formatNumber(report.debris_silicon ?? 0)}</td>
            </tr>
          </>
        )}

        {/* Moon chance */}
        {(report.moon_chance ?? 0) > 0 && (
          <tr>
            <td colSpan={4}>
              {t('mission', 'moonChance') ?? 'Шанс луны'}:{' '}
              <span className={report.moon_created ? 'true' : ''}>
                {Math.floor((report.moon_chance ?? 0) * 100)}%
              </span>
              {report.moon_created && ' ✓'}
            </td>
          </tr>
        )}
      </tbody>
    </table>
  );
}
