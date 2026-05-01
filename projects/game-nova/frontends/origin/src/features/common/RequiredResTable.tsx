// RequiredResTable — pixel-perfect клон legacy required_res_table.tpl.
// Показывает требуемые ресурсы для постройки/исследования + время.
//
// План 72.1.45 §4: расширен на energy/credit/points (legacy строки 24-44
// required_res_table.tpl + legacy юниты 322-330 ниже могут требовать energy
// в качестве upkeep, credit для VIP-ускорения, points для special-флоу).

import { formatNumber, formatDuration } from '@/lib/format';
import { useTranslation } from '@/i18n/i18n';

export interface RequiredResProps {
  metal: number;
  silicon: number;
  hydrogen: number;
  energy?: number;
  credit?: number;
  points?: number;
  /** Текущий запас ресурсов на планете (для подсветки нехватки) */
  available?: {
    metal: number;
    silicon: number;
    hydrogen: number;
    energy?: number;
    credit?: number;
    points?: number;
  };
  /** Время постройки в секундах */
  seconds: number;
}

export function RequiredResTable({
  metal,
  silicon,
  hydrogen,
  energy = 0,
  credit = 0,
  points = 0,
  available,
  seconds,
}: RequiredResProps) {
  const { t } = useTranslation();
  function diff(req: number, have: number | undefined): number {
    if (have === undefined) return 0;
    return Math.max(0, req - have);
  }

  const metalLack = diff(metal, available?.metal);
  const siliconLack = diff(silicon, available?.silicon);
  const hydrogenLack = diff(hydrogen, available?.hydrogen);
  const energyLack = diff(energy, available?.energy);
  const creditLack = diff(credit, available?.credit);
  const pointsLack = diff(points, available?.points);

  return (
    <table
      className="table_no_background"
      cellSpacing={0}
      cellPadding={0}
      title={t('info', 'requires') || 'Требуется'}
    >
      <tbody>
        {metal > 0 && (
          <tr>
            <td>{t('overview', 'metal') || 'Металл'}</td>
            <td className={metalLack > 0 ? 'notavailable' : 'true'}>
              {formatNumber(metal)}
            </td>
            <td>{metalLack > 0 && <>({formatNumber(metalLack)})</>}</td>
          </tr>
        )}
        {silicon > 0 && (
          <tr>
            <td>{t('overview', 'silicon') || 'Кремний'}</td>
            <td className={siliconLack > 0 ? 'notavailable' : 'true'}>
              {formatNumber(silicon)}
            </td>
            <td>{siliconLack > 0 && <>({formatNumber(siliconLack)})</>}</td>
          </tr>
        )}
        {hydrogen > 0 && (
          <tr>
            <td>{t('overview', 'hydrogen') || 'Водород'}</td>
            <td className={hydrogenLack > 0 ? 'notavailable' : 'true'}>
              {formatNumber(hydrogen)}
            </td>
            <td>{hydrogenLack > 0 && <>({formatNumber(hydrogenLack)})</>}</td>
          </tr>
        )}
        {energy > 0 && (
          <tr>
            <td>{t('overview', 'energy') || 'Энергия'}</td>
            <td className={energyLack > 0 ? 'notavailable' : 'true'}>
              {formatNumber(energy)}
            </td>
            <td>{energyLack > 0 && <>({formatNumber(energyLack)})</>}</td>
          </tr>
        )}
        {credit > 0 && (
          <tr>
            <td>{t('overview', 'credits') || 'Кредиты'}</td>
            <td className={creditLack > 0 ? 'notavailable' : 'true'}>
              {formatNumber(credit)}
            </td>
            <td>{creditLack > 0 && <>({formatNumber(creditLack)})</>}</td>
          </tr>
        )}
        {points > 0 && (
          <tr>
            <td>{t('overview', 'points') || 'Очки'}</td>
            <td className={pointsLack > 0 ? 'notavailable' : 'true'}>
              {formatNumber(points)}
            </td>
            <td>{pointsLack > 0 && <>({formatNumber(pointsLack)})</>}</td>
          </tr>
        )}
        <tr>
          <td>{t('info', 'requireTime') || 'Время'}</td>
          <td colSpan={2}>{formatDuration(seconds)}</td>
        </tr>
      </tbody>
    </table>
  );
}
