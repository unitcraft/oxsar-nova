// Полный pixel-perfect rendering боевого отчёта
// (план 72.1 ч.20.11.4 — порт oxsar2-java/Assault.java).
//
// Структура:
// 1. Заголовок: «Результат симуляции / Боевой отчёт» + winner + раундов.
// 2. Для каждого раунда:
//    - <p><b>Раунд: N</b></p>
//    - Атакующий: tech-power панель (gun/shield/armoring/ballistics/masking)
//      + per-unit таблица (Type/Quantity/Guns/Shields/Shells/Front/...)
//    - Защитник: то же самое
//    - «Бой» — Fight-таблица 6 колонок:
//      shots / power / miss / shield_absorb / shell_destroyed / units_destroyed
// 3. Сводные потери.
// 4. Поле обломков, шанс луны.

import { useTranslation } from '@/i18n/i18n';
import { formatNumber } from '@/lib/format';
import type { SimReport, SimRoundSide, SimRoundUnit } from '@/api/simulator';
import { findCatalog } from '@/features/common/catalog';

interface Props {
  report: SimReport;
  title?: string;
}

export function BattleReportView({ report, title }: Props) {
  const { t } = useTranslation();
  const winnerLabel =
    report.winner === 'attackers'
      ? t('mission', 'attackerWins') ?? 'Атакующий победил'
      : report.winner === 'defenders'
        ? t('mission', 'defenderWins') ?? 'Защитник победил'
        : t('mission', 'draw') ?? 'Ничья';
  const winnerClass =
    report.winner === 'attackers'
      ? 'true'
      : report.winner === 'defenders'
        ? 'false'
        : '';
  const headerTitle = title ?? (t('mission', 'simulationResult') ?? 'Результат симуляции');
  return (
    <>
      <table className="ntable" style={{ marginTop: 16 }}>
        <thead>
          <tr>
            <th>
              {headerTitle} —{' '}
              <span className={winnerClass}>{winnerLabel}</span>
              {' · '}
              {t('mission', 'roundCount') ?? 'Раундов'}: {report.rounds}
            </th>
          </tr>
        </thead>
      </table>

      {(report.rounds_trace ?? []).map((rt) => (
        <RoundView
          key={rt.index}
          index={rt.index}
          attacker={rt.attacker_side}
          defender={rt.defender_side}
        />
      ))}

      {/* Сводные потери */}
      <table className="ntable" style={{ marginTop: 16 }}>
        <tbody>
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
    </>
  );
}

interface RoundProps {
  index: number;
  attacker: SimRoundSide;
  defender: SimRoundSide;
}

function RoundView({ index, attacker, defender }: RoundProps) {
  const { t } = useTranslation();
  return (
    <div style={{ marginTop: 16 }}>
      <p>
        <b>{t('mission', 'round') ?? 'Раунд'}: {index + 1}</b>
      </p>

      <SideBlock side={attacker} kind="attacker" />
      <SideBlock side={defender} kind="defender" />

      {/* Fight таблица */}
      <p>
        <b>{t('assaultReport', 'fight') ?? 'Бой'}</b>
      </p>
      <table className="atable">
        <thead>
          <tr>
            <th>&nbsp;</th>
            <th>{t('assaultReport', 'shotsNumber') ?? 'Выстрелов'}</th>
            <th>{t('assaultReport', 'shotsPower') ?? 'Мощность'}</th>
            <th>{t('assaultReport', 'shotsMiss') ?? 'Промахи'}</th>
            <th>{t('assaultReport', 'shieldAbsorb') ?? 'Поглощено щитами'}</th>
            <th>{t('assaultReport', 'shellDestroyedCol') ?? 'Уничтожено брони'}</th>
            <th>{t('assaultReport', 'unitsDestroyed') ?? 'Уничтожено юнитов'}</th>
          </tr>
        </thead>
        <tbody>
          <FightRow side={attacker} oppShield={defender.shield_absorbed} oppShell={defender.shell_destroyed} kind="attacker" />
          <FightRow side={defender} oppShield={attacker.shield_absorbed} oppShell={attacker.shell_destroyed} kind="defender" />
        </tbody>
      </table>
    </div>
  );
}

function SideBlock({ side, kind }: { side: SimRoundSide; kind: 'attacker' | 'defender' }) {
  const { t } = useTranslation();
  const label =
    kind === 'attacker'
      ? (t('mission', 'attacker') ?? 'Атакующий')
      : (t('mission', 'defender') ?? 'Защитник');
  return (
    <>
      <div>
        <b>{label}</b>
      </div>
      <div style={{ marginBottom: 8 }}>
        <small>
          {t('assaultReport', 'gunPower') ?? 'Оружие'}: {Math.round(side.gun_power_pct)}%
          {' · '}
          {t('assaultReport', 'shieldPower') ?? 'Щиты'}: {Math.round(side.shield_power_pct)}%
          {' · '}
          {t('assaultReport', 'armoring') ?? 'Броня'}: {Math.round(side.armoring_pct)}%
          {' · '}
          {t('assaultReport', 'ballisticsPower') ?? 'Баллистика'}: {side.ballistics_lvl}
          {' · '}
          {t('assaultReport', 'maskingPower') ?? 'Маскировка'}: {side.masking_lvl}
        </small>
      </div>
      <UnitTable units={side.units} />
    </>
  );
}

function UnitTable({ units }: { units: SimRoundUnit[] }) {
  const { t } = useTranslation();
  if (!units || units.length === 0) {
    return null;
  }
  // Формируем колонки, как в Java printParticipant: одна строка-заголовок
  // с типами + строки quantity/guns/shields/shells/front (+ ballistics/masking
  // если хоть у одного юнита уровень > 0).
  const showBallistics = units.some((u) => (u.ballistics_level ?? 0) > 0);
  const showMasking = units.some((u) => (u.masking_level ?? 0) > 0);
  return (
    <table className="atable" style={{ marginBottom: 8 }}>
      <thead>
        <tr>
          <th>&nbsp;</th>
          {units.map((u) => (
            <th key={`name-${u.unit_id}`}>{unitName(u, t)}</th>
          ))}
        </tr>
      </thead>
      <tbody>
        <tr>
          <th>{t('assaultReport', 'quantity') ?? 'Количество'}</th>
          {units.map((u) => (
            <td key={`q-${u.unit_id}`} style={{ whiteSpace: 'nowrap' }}>
              {formatNumber(u.start_turn_quantity)}
              {u.start_turn_damaged > 0 && (
                <>
                  <br />
                  <span className="rep_quantity_damage">
                    {u.start_turn_damaged !== u.start_turn_quantity && (
                      <>{formatNumber(u.start_turn_damaged)} - </>
                    )}
                    {u.damaged_shell_percent}%
                  </span>
                </>
              )}
            </td>
          ))}
        </tr>
        <tr>
          <th>{t('assaultReport', 'guns') ?? 'Атака'}</th>
          {units.map((u) => (
            <td key={`g-${u.unit_id}`}>{formatNumber(Math.round(u.attack))}</td>
          ))}
        </tr>
        <tr>
          <th>{t('assaultReport', 'shields') ?? 'Щит'}</th>
          {units.map((u) => (
            <td key={`s-${u.unit_id}`}>{formatNumber(Math.round(u.shield))}</td>
          ))}
        </tr>
        <tr>
          <th>{t('assaultReport', 'shells') ?? 'Броня'}</th>
          {units.map((u) => (
            <td key={`sh-${u.unit_id}`}>{formatNumber(Math.round(u.shell))}</td>
          ))}
        </tr>
        <tr>
          <th>{t('assaultReport', 'front') ?? 'Фронт'}</th>
          {units.map((u) => (
            <td key={`f-${u.unit_id}`}>{u.front}</td>
          ))}
        </tr>
        {showBallistics && (
          <tr>
            <th>{t('assaultReport', 'ballisticsPower') ?? 'Баллистика'}</th>
            {units.map((u) => (
              <td key={`b-${u.unit_id}`}>{u.ballistics_level ?? 0}</td>
            ))}
          </tr>
        )}
        {showMasking && (
          <tr>
            <th>{t('assaultReport', 'maskingPower') ?? 'Маскировка'}</th>
            {units.map((u) => (
              <td key={`m-${u.unit_id}`}>{u.masking_level ?? 0}</td>
            ))}
          </tr>
        )}
        <tr>
          <th>%</th>
          {units.map((u) => (
            <td key={`a-${u.unit_id}`}>
              <div className="rep_destroyed_back_div" style={{ width: 50, height: 6, background: '#444' }}>
                <div
                  className="rep_alive_over_div"
                  style={{ width: `${u.alive_percent}%`, height: '100%', background: '#4a4' }}
                />
              </div>
            </td>
          ))}
        </tr>
      </tbody>
    </table>
  );
}

function unitName(u: SimRoundUnit, t: (g: string, k: string) => string): string {
  if (u.name) return u.name;
  const cat = findCatalog(u.unit_id);
  if (!cat) return `#${u.unit_id}`;
  const [g, k] = cat.i18n.split('.') as [string, string];
  return t(g, k);
}

interface FightRowProps {
  side: SimRoundSide;
  oppShield: number;
  oppShell: number;
  kind: 'attacker' | 'defender';
}

// Java: миссы = power - shieldAbsorb - shellDestroyed.
// FightRow для атакующего: shieldAbsorb/shellDestroyed — defender's
// (т.е. сколько защитник поглотил/потерял = side.shield_absorbed,
// side.shell_destroyed — это уже те значения).
function FightRow({ side, kind }: FightRowProps) {
  const { t } = useTranslation();
  const label =
    kind === 'attacker'
      ? (t('assaultReport', 'fightAttacker') ?? 'Атакующий')
      : (t('assaultReport', 'fightDefender') ?? 'Защитник');
  const miss = Math.max(0, side.power - side.shield_absorbed - side.shell_destroyed);
  return (
    <tr>
      <th>{label}</th>
      <td>{formatNumber(side.shots)}</td>
      <td>{formatNumber(Math.round(side.power))}</td>
      <td><PowerCell value={miss} total={side.power} /></td>
      <td><PowerCell value={side.shield_absorbed} total={side.power} /></td>
      <td><PowerCell value={side.shell_destroyed} total={side.power} /></td>
      <td>{formatNumber(side.units_destroyed)}</td>
    </tr>
  );
}

// PowerCell — Java formatPower: число + надстрочный процент от total.
function PowerCell({ value, total }: { value: number; total: number }) {
  const v = value < 0 ? 0 : value;
  const pct = total > 0 ? Math.round((v * 100) / total) : 0;
  return (
    <>
      {formatNumber(Math.round(v))}
      {total > 0 && (
        <sup>
          <span className="rep_quantity_damage_low"> {pct}%</span>
        </sup>
      )}
    </>
  );
}
