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
  // compact: убрать header «Результат — Победитель · Раундов» —
  // используется на публичном /battle-report/:id, где показываем
  // «только сам бой» (план 72.1 ч.20.11.11).
  compact?: boolean;
  // startedAt: ISO timestamp начала боя. В compact-режиме рендерится
  // как фраза «Флоты соперников встрелись в <datetime> часов:» —
  // legacy assaultReport.assaultTime (план 72.1 ч.20.11.12).
  startedAt?: string;
  // reportId: UUID отчёта. Используется для friend-link «Покажи это
  // сражение друзьям» в финальном SUMMARY (Java assaultReport.friendLink).
  reportId?: string;
}

// formatBattleTime — формат «Thu, 30 Apr 2026, 21:31:57» через
// Intl.DateTimeFormat. Локаль зависит от current i18n lang (ru/en) —
// не хардкодим месяцы/дни недели.
function formatBattleTime(iso: string, lang: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  return new Intl.DateTimeFormat(lang === 'ru' ? 'ru-RU' : 'en-GB', {
    weekday: 'short',
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
  }).format(d);
}

export function BattleReportView({
  report,
  title,
  compact = false,
  startedAt,
  reportId,
}: Props) {
  const { t, lang } = useTranslation();
  const winnerLabel =
    report.winner === 'attackers'
      ? t('mission', 'attackerWins')
      : report.winner === 'defenders'
        ? t('mission', 'defenderWins')
        : t('mission', 'draw');
  const winnerClass =
    report.winner === 'attackers'
      ? 'true'
      : report.winner === 'defenders'
        ? 'false'
        : '';
  const headerTitle = title ?? t('mission', 'simulationResult');
  return (
    <>
      {!compact && (
        <table className="ntable" style={{ marginTop: 16 }}>
          <thead>
            <tr>
              <th>
                {headerTitle} —{' '}
                <span className={winnerClass}>{winnerLabel}</span>
                {' · '}
                {t('mission', 'roundCount')}: {report.rounds}
              </th>
            </tr>
          </thead>
        </table>
      )}

      {/* «Флоты соперников встрелись в <datetime> часов:» —
          legacy assaultReport.assaultTime (план 72.1 ч.20.11.12). */}
      {startedAt && (
        <p style={{ textAlign: 'center', marginTop: 8 }}>
          <b>
            {t('assaultReport', 'assaultTime', { time: formatBattleTime(startedAt, lang) })}
          </b>
        </p>
      )}

      {(report.rounds_trace ?? []).map((rt) => (
        <RoundView
          key={rt.index}
          index={rt.index}
          attacker={rt.attacker_side}
          defender={rt.defender_side}
        />
      ))}

      <SummaryBlock
        report={report}
        reportId={reportId}
        winnerLabel={winnerLabel}
        winnerClass={winnerClass}
      />
    </>
  );
}

// SummaryBlock — финальный блок отчёта, pixel-perfect клон Java SUMMARY
// (Assault.java:1066-1170, план 72.1 ч.20.11.13). Структура:
//   - Заголовок «Итоговый результат».
//   - Финальные таблицы юнитов по сторонам (последний RoundSide).
//   - Текст победителя или «Сражение закончилось в ничью.».
//   - При победе атакующего — «Он получает: …» с трофеями.
//   - Потери атакующего/обороняющегося (text, не таблица).
//   - «На орбите находится: …» (если есть debris).
//   - Опыт обеих сторон.
//   - Friend-link на /battle-report/{id}.
function SummaryBlock({
  report,
  reportId,
  winnerLabel,
  winnerClass,
}: {
  report: SimReport;
  reportId: string | undefined;
  winnerLabel: string;
  winnerClass: string;
}) {
  const { t } = useTranslation();
  const last = (report.rounds_trace ?? [])[
    (report.rounds_trace ?? []).length - 1
  ];
  const atkLoss = (report.attackers ?? []).reduce(
    (acc, s) => ({
      m: acc.m + s.lost_metal,
      s: acc.s + s.lost_silicon,
      h: acc.h + s.lost_hydrogen,
    }),
    { m: 0, s: 0, h: 0 },
  );
  const defLoss = (report.defenders ?? []).reduce(
    (acc, s) => ({
      m: acc.m + s.lost_metal,
      s: acc.s + s.lost_silicon,
      h: acc.h + s.lost_hydrogen,
    }),
    { m: 0, s: 0, h: 0 },
  );
  const atkTotalLoss = atkLoss.m + atkLoss.s + atkLoss.h;
  const defTotalLoss = defLoss.m + defLoss.s + defLoss.h;
  const friendLink =
    typeof window !== 'undefined' && reportId
      ? `${window.location.origin}/battle-report/${reportId}`
      : '';

  return (
    <div style={{ marginTop: 16, textAlign: 'center' }}>
      <p>
        <b>{t('assaultReport', 'summary')}</b>
      </p>

      {last && <SideBlock side={last.attacker_side} kind="attacker" />}
      {last && <SideBlock side={last.defender_side} kind="defender" />}

      <p style={{ marginTop: 12 }}>
        <b className={winnerClass}>
          {report.winner === 'attackers'
            ? t('assaultReport', 'attackerWon')
            : report.winner === 'defenders'
              ? t('assaultReport', 'defenderWon')
              : winnerLabel}
        </b>
      </p>

      {report.winner === 'attackers'
        && ((report.haul_metal ?? 0) + (report.haul_silicon ?? 0) + (report.haul_hydrogen ?? 0) > 0) && (
        <p>
          {t('assaultReport', 'attackerHaul')}
          <br />
          {formatNumber(report.haul_metal ?? 0)} {t('global', 'metal')},{' '}
          {formatNumber(report.haul_silicon ?? 0)} {t('global', 'silicon')}{' '}
          {t('assaultReport', 'and')}{' '}
          {formatNumber(report.haul_hydrogen ?? 0)} {t('global', 'hydrogen')}.
        </p>
      )}

      <p>
        {t('assaultReport', 'attackerLostRes4', {
          total: formatNumber(atkTotalLoss),
          metal: formatNumber(atkLoss.m),
          silicon: formatNumber(atkLoss.s),
          hydrogen: formatNumber(atkLoss.h),
        })}
        <br />
        {t('assaultReport', 'defenderLostRes4', {
          total: formatNumber(defTotalLoss),
          metal: formatNumber(defLoss.m),
          silicon: formatNumber(defLoss.s),
          hydrogen: formatNumber(defLoss.h),
        })}
      </p>

      {((report.debris_metal ?? 0) > 0 || (report.debris_silicon ?? 0) > 0) && (
        <p>
          {t('assaultReport', 'debrisMetalAndSilicon', {
            metal: formatNumber(report.debris_metal ?? 0),
            silicon: formatNumber(report.debris_silicon ?? 0),
          })}
        </p>
      )}

      <p>
        {t('assaultReport', 'attackerExperience', {
          count: formatNumber(report.attacker_exp ?? 0),
        })}
        <br />
        {t('assaultReport', 'defenderExperience', {
          count: formatNumber(report.defender_exp ?? 0),
        })}
      </p>

      {(report.moon_chance ?? 0) > 0 && (
        <p>
          {t('assaultReport', 'moonChance', {
            percent: String(Math.floor((report.moon_chance ?? 0) * 100)),
          })}
          {report.moon_created && (
            <>
              <br />
              <b className="true">{t('assaultReport', 'moon')}</b>
            </>
          )}
        </p>
      )}

      {friendLink && (
        <p>
          {t('assaultReport', 'friendLink')}
          <br />
          <a href={friendLink} className="false2">{friendLink}</a>
        </p>
      )}
    </div>
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
        <b>{t('mission', 'round')}: {index + 1}</b>
      </p>

      <SideBlock side={attacker} kind="attacker" />
      <SideBlock side={defender} kind="defender" />

      {/* Fight таблица */}
      <p style={{ textAlign: 'center' }}>
        <b>{t('assaultReport', 'fight')}</b>
      </p>
      <table className="atable">
        <thead>
          <tr>
            <th>&nbsp;</th>
            <th><FightIcon name="shots_number" title={t('assaultReport', 'fightShotsNumber')} /></th>
            <th><FightIcon name="shots_power" title={t('assaultReport', 'fightShotsPower')} /></th>
            <th><FightIcon name="shots_miss" title={t('assaultReport', 'fightShotsMiss')} /></th>
            <th><FightIcon name="shield_absorb" title={t('assaultReport', 'fightShieldAbsorb')} /></th>
            <th><FightIcon name="shell_destroyed" title={t('assaultReport', 'fightShellDestroyed')} /></th>
            <th><FightIcon name="units_destroyed" title={t('assaultReport', 'fightUnitsDestroyed')} /></th>
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
        {t('assaultReport', 'attackerShots', { count: formatNumber(attacker.shots) })}{' '}
        {t('assaultReport', 'attackerPower', { power: formatNumber(Math.round(attacker.power)) })}{' '}
        {t('assaultReport', 'defenderShield', { shield: formatNumber(Math.round(attacker.shield_absorbed)) })}
        <br />
        {t('assaultReport', 'defenderShots', { count: formatNumber(defender.shots) })}{' '}
        {t('assaultReport', 'defenderPower', { power: formatNumber(Math.round(defender.power)) })}{' '}
        {t('assaultReport', 'attackerShield', { shield: formatNumber(Math.round(defender.shield_absorbed)) })}
      </p>
    </div>
  );
}

// formatPosition — «[g:s:p]» с возможным «(луна)».
function formatPosition(side: SimRoundSide, t: TFunc): string {
  const g = side.galaxy ?? 0;
  const s = side.system ?? 0;
  const p = side.position ?? 0;
  const moon = side.is_moon ? ` (${t('global', 'moon')})` : '';
  return `[${g}:${s}:${p}]${moon}`;
}

function SideBlock({ side, kind }: { side: SimRoundSide; kind: 'attacker' | 'defender' }) {
  const { t } = useTranslation();
  const label = kind === 'attacker' ? t('mission', 'attacker') : t('mission', 'defender');
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
          {t('assaultReport', 'gunPower')}: {Math.round(side.gun_power_pct)}%
          {'   '}
          {t('assaultReport', 'shieldPower')}: {Math.round(side.shield_power_pct)}%
          {'   '}
          {t('assaultReport', 'armoring')}: {Math.round(side.armoring_pct)}%
          {'   '}
          {t('assaultReport', 'ballisticsPower')}: {side.ballistics_lvl}
          {'   '}
          {t('assaultReport', 'maskingPower')}: {side.masking_lvl}
        </small>
      </p>
      <UnitTable units={side.units} />
    </>
  );
}

// TFunc — сигнатура t (group, key, vars?).
type TFunc = (g: string, k: string, vars?: Record<string, string | number>) => string;

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
          <th>{t('assaultReport', 'quantity')}</th>
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
          <th>{t('assaultReport', 'guns')}</th>
          {units.map((u) => (
            <td key={`g-${u.unit_id}`}>{formatNumber(Math.round(u.attack))}</td>
          ))}
        </tr>
        <tr>
          <th>{t('assaultReport', 'shields')}</th>
          {units.map((u) => (
            <td key={`s-${u.unit_id}`}>{formatNumber(Math.round(u.shield))}</td>
          ))}
        </tr>
        <tr>
          <th>{t('assaultReport', 'shells')}</th>
          {units.map((u) => (
            <td key={`sh-${u.unit_id}`}>{formatNumber(Math.round(u.shell))}</td>
          ))}
        </tr>
        <tr>
          <th>{t('assaultReport', 'front')}</th>
          {units.map((u) => (
            <td key={`f-${u.unit_id}`}>{u.front}</td>
          ))}
        </tr>
        {showBallistics && (
          <tr>
            <th>{t('assaultReport', 'ballisticsPower')}</th>
            {units.map((u) => (
              <td key={`b-${u.unit_id}`}>{u.ballistics_level ?? 0}</td>
            ))}
          </tr>
        )}
        {showMasking && (
          <tr>
            <th>{t('assaultReport', 'maskingPower')}</th>
            {units.map((u) => (
              <td key={`m-${u.unit_id}`}>{u.masking_level ?? 0}</td>
            ))}
          </tr>
        )}
        <tr>
          <th>%</th>
          {units.map((u) => {
            // Полоска alive% (legacy: красный фон-«потери», зелёный
            // overlay-«живые» поверх). 1:1 с oxsar2 Battlestats CSS:
            // .rep_destroyed_back_div = background rgb(252,50,50);
            // .rep_alive_over_div = width: alive%; зелёный overlay.
            const alive = Math.max(0, Math.min(100, u.alive_percent));
            return (
              <td key={`a-${u.unit_id}`}>
                <div
                  className="rep_destroyed_back_div"
                  style={{ height: 5, background: 'rgb(252, 50, 50)' }}
                >
                  <div
                    className="rep_alive_over_div"
                    style={{
                      width: `${alive}%`,
                      height: '100%',
                      background: 'rgb(92, 208, 200)',
                    }}
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

function unitName(u: SimRoundUnit, t: TFunc): string {
  // Catalog — источник истины для имён юнитов (i18n ключ привязан к
  // unit_id). Backend `name` — это raw key (типа «small_transporter»),
  // не годится для отображения. Fallback на name только если catalog
  // не нашёл.
  const cat = findCatalog(u.unit_id);
  if (cat) {
    const [g, k] = cat.i18n.split('.') as [string, string];
    return t(g, k);
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
      ? t('assaultReport', 'fightAttacker')
      : t('assaultReport', 'fightDefender');
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

// FightIcon — иконка-заголовок Fight-таблицы (план 72.1 ч.20.11.10).
// Java: <th>{img}FIGHT_<NAME>{/img}</th>. Файлы взяты из legacy
// www/images/fight_*.gif и положены в public/assets/origin/images/.
type FightIconName =
  | 'shots_number'
  | 'shots_power'
  | 'shots_miss'
  | 'shield_absorb'
  | 'shell_destroyed'
  | 'units_destroyed';

function FightIcon({ name, title }: { name: FightIconName; title: string }) {
  return (
    <img
      src={`/assets/origin/images/fight_${name}.gif`}
      alt={title}
      title={title}
      style={{ verticalAlign: 'middle' }}
    />
  );
}
