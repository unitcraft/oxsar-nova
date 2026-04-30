// Общий компонент просмотра боевого отчёта (план 72.1 ч.20.8).
// Используется Simulator и /battle-reports/:id. Pixel-perfect клон
// legacy oxsar2-java/Assault.java HTML rendering.
//
// Принимает структурированный battle.Report (frontend SimReport).
// Отображает: победитель + раунды + потери М/К/В + поле обломков +
// шанс луны.

import { useTranslation } from '@/i18n/i18n';
import { formatNumber } from '@/lib/format';
import type { SimReport } from '@/api/simulator';

interface Props {
  report: SimReport;
  /** Заголовок таблицы (по умолчанию «Результат симуляции»). */
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
  const headerTitle =
    title ?? (t('mission', 'simulationResult') ?? 'Результат симуляции');
  return (
    <table className="ntable" style={{ marginTop: 16 }}>
      <thead>
        <tr>
          <th colSpan={4}>
            {headerTitle} —{' '}
            <span className={winnerClass}>{winnerLabel}</span>
            {' · '}
            {t('mission', 'roundCount') ?? 'Раундов'}: {report.rounds}
          </th>
        </tr>
      </thead>
      <tbody>
        {(report.rounds_trace ?? []).map((rt) => (
          <tr key={rt.index}>
            <td>{t('mission', 'round') ?? 'Раунд'} {rt.index + 1}</td>
            <td colSpan={3}>
              {t('mission', 'attackersAlive') ?? 'Атакующие'}:{' '}
              <b>{formatNumber(rt.attackers_alive)}</b>{' · '}
              {t('mission', 'defendersAlive') ?? 'Защитники'}:{' '}
              <b>{formatNumber(rt.defenders_alive)}</b>
            </td>
          </tr>
        ))}

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
  );
}
