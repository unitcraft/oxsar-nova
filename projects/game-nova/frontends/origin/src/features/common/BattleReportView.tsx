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
      <p style={{ textAlign: 'center' }}>
        <b>{t('mission', 'round') ?? 'Раунд'}: {index + 1}</b>
      </p>

      <SideBlock side={attacker} kind="attacker" />
      <SideBlock side={defender} kind="defender" />

      {/* Fight таблица */}
      <p style={{ textAlign: 'center' }}>
        <b>{t('assaultReport', 'fight') ?? 'Бой'}</b>
      </p>
      <table className="atable">
        <thead>
          <tr>
            <th>&nbsp;</th>
            <th title={t('assaultReport', 'fightShotsNumber') ?? 'Делает выстрелов'}>
              {t('assaultReport', 'shotsNumber') ?? 'Выстрелов'}
            </th>
            <th title={t('assaultReport', 'fightShotsPower') ?? 'Мощность огня'}>
              {t('assaultReport', 'shotsPower') ?? 'Мощность'}
            </th>
            <th title={t('assaultReport', 'fightShotsMiss') ?? 'Промахи, попадания в уничтоженные.'}>
              {t('assaultReport', 'shotsMiss') ?? 'Промахи'}
            </th>
            <th title={t('assaultReport', 'fightShieldAbsorb') ?? 'Щиты противника поглотили'}>
              {t('assaultReport', 'shieldAbsorb') ?? 'Поглощено щитами'}
            </th>
            <th title={t('assaultReport', 'fightShellDestroyed') ?? 'Разрушено брони противника'}>
              {t('assaultReport', 'shellDestroyedCol') ?? 'Уничтожено брони'}
            </th>
            <th title={t('assaultReport', 'fightUnitsDestroyed') ?? 'Уничтожено юнитов противника'}>
              {t('assaultReport', 'unitsDestroyed') ?? 'Уничтожено юнитов'}
            </th>
          </tr>
        </thead>
        <tbody>
          <FightRow side={attacker} oppShield={defender.shield_absorbed} oppShell={defender.shell_destroyed} kind="attacker" />
          <FightRow side={defender} oppShield={attacker.shield_absorbed} oppShell={attacker.shell_destroyed} kind="defender" />
        </tbody>
      </table>

      {/* Параграфы про выстрелы и щиты — Java printShootStat (план 72.1
          ч.20.11.9). i18n шаблоны attackerShots/defenderShield уже
          содержат «Атакующий…» / «Щиты обороняющегося…», поэтому
          лейблы стороны не дублируем. */}
      <p style={{ textAlign: 'center', marginTop: 8 }}>
        {(t('assaultReport', 'attackerShots') ?? 'Атакующий делает {{count}} выстрелов.').replace(
          '{{count}}',
          formatNumber(attacker.shots),
        )}{' '}
        {(t('assaultReport', 'attackerPower') ?? 'общей мощностью {{power}}.').replace(
          '{{power}}',
          formatNumber(Math.round(attacker.power)),
        )}{' '}
        {(t('assaultReport', 'defenderShield') ?? 'Щиты обороняющегося поглощают {{shield}}.').replace(
          '{{shield}}',
          formatNumber(Math.round(attacker.shield_absorbed)),
        )}
        <br />
        {(t('assaultReport', 'defenderShots') ?? 'Обороняющийся делает {{count}} выстрелов.').replace(
          '{{count}}',
          formatNumber(defender.shots),
        )}{' '}
        {(t('assaultReport', 'defenderPower') ?? 'общей мощностью {{power}}.').replace(
          '{{power}}',
          formatNumber(Math.round(defender.power)),
        )}{' '}
        {(t('assaultReport', 'attackerShield') ?? 'Щиты атакующего поглощают {{shield}}.').replace(
          '{{shield}}',
          formatNumber(Math.round(defender.shield_absorbed)),
        )}
      </p>
    </div>
  );
}

// formatPosition — «[g:s:p]» с возможным «(луна)».
function formatPosition(side: SimRoundSide, t: (g: string, k: string) => string | null): string {
  const g = side.galaxy ?? 0;
  const s = side.system ?? 0;
  const p = side.position ?? 0;
  const moon = side.is_moon ? ` (${t('global', 'moon') ?? 'луна'})` : '';
  return `[${g}:${s}:${p}]${moon}`;
}

function SideBlock({ side, kind }: { side: SimRoundSide; kind: 'attacker' | 'defender' }) {
  const { t } = useTranslation();
  const label =
    kind === 'attacker'
      ? (t('mission', 'attacker') ?? 'Атакующий')
      : (t('mission', 'defender') ?? 'Защитник');
  const position = formatPosition(side, t);
  return (
    <>
      {/* Заголовок: «Атакующий <username> [g:s:p] (луна)» — Java
          формирует так перед tech-power каждого участника раунда. */}
      <p style={{ textAlign: 'center', marginBottom: 0 }}>
        <b>{label}</b>
        {side.username ? ` ${side.username}` : ''}
        {' '}
        {position}
      </p>
      <p style={{ textAlign: 'center', marginTop: 0, marginBottom: 8 }}>
        <small>
          {t('assaultReport', 'gunPower') ?? 'Оружие'}: {Math.round(side.gun_power_pct)}%
          {'   '}
          {t('assaultReport', 'shieldPower') ?? 'Щиты'}: {Math.round(side.shield_power_pct)}%
          {'   '}
          {t('assaultReport', 'armoring') ?? 'Броня'}: {Math.round(side.armoring_pct)}%
          {'   '}
          {t('assaultReport', 'ballisticsPower') ?? 'Баллистика'}: {side.ballistics_lvl}
          {'   '}
          {t('assaultReport', 'maskingPower') ?? 'Маскировка'}: {side.masking_lvl}
        </small>
      </p>
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
          {units.map((u) => {
            // Класс damaged-строки: shellPercent <= 70% — оранжевый
            // (rep_quantity_damage), > 70 — бледный (rep_quantity_damage_low).
            // Java printParticipant строки 1587-1588.
            const damagedCls =
              u.damaged_shell_percent <= 70
                ? 'rep_quantity_damage'
                : 'rep_quantity_damage_low';
            return (
              <td key={`q-${u.unit_id}`} style={{ whiteSpace: 'nowrap' }}>
                {formatNumber(u.start_turn_quantity)}
                {/* diff: «( -18 )» при потерях прошлого раунда */}
                {(u.start_turn_quantity_diff ?? 0) < 0 && (
                  <>
                    {' '}
                    <span className="rep_quantity_diff">
                      ( {formatNumber(u.start_turn_quantity_diff)} )
                    </span>
                  </>
                )}
                {u.start_turn_damaged > 0 && (
                  <>
                    <br />
                    <span className={damagedCls}>
                      {u.start_turn_damaged !== u.start_turn_quantity && (
                        <>{formatNumber(u.start_turn_damaged)} - </>
                      )}
                      {u.damaged_shell_percent}%
                    </span>
                  </>
                )}
              </td>
            );
          })}
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
          {units.map((u) => {
            // Полоска alive%: зелёный = живые здоровые, красный
            // справа = убитые с начала боя. damaged-юниты включены в
            // зелёное (их «недо-shell» отдельной шкалой не показываем).
            const alive = Math.max(0, Math.min(100, u.alive_percent));
            return (
              <td key={`a-${u.unit_id}`}>
                <div
                  className="rep_destroyed_back_div"
                  style={{ width: 50, height: 6, display: 'flex', background: 'transparent' }}
                >
                  <div
                    className="rep_alive_over_div"
                    style={{ width: `${alive}%`, height: '100%', background: '#5cd0c8' }}
                  />
                  <div
                    className="rep_destroyed_over_div"
                    style={{ width: `${100 - alive}%`, height: '100%', background: '#d54' }}
                  />
                </div>
              </td>
            );
          })}
        </tr>
      </tbody>
    </table>
  );
}

function unitName(u: SimRoundUnit, t: (g: string, k: string) => string): string {
  // Catalog — источник истины для имён юнитов (i18n ключ привязан к
  // unit_id). Backend `name` — это raw key (типа «small_transporter»),
  // не годится для отображения. Fallback на name только если catalog
  // не нашёл.
  const cat = findCatalog(u.unit_id);
  if (cat) {
    const [g, k] = cat.i18n.split('.') as [string, string];
    const tr = t(g, k);
    if (tr) return tr;
  }
  if (u.name) return u.name;
  return `#${u.unit_id}`;
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
