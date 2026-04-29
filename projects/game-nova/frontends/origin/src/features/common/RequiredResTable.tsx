// RequiredResTable — pixel-perfect клон legacy required_res_table.tpl.
// Показывает требуемые ресурсы для постройки/исследования + время.

import { formatNumber, formatDuration } from '@/lib/format';

export interface RequiredResProps {
  metal: number;
  silicon: number;
  hydrogen: number;
  /** Текущий запас ресурсов на планете (для подсветки нехватки) */
  available?: { metal: number; silicon: number; hydrogen: number };
  /** Время постройки в секундах */
  seconds: number;
}

export function RequiredResTable({ metal, silicon, hydrogen, available, seconds }: RequiredResProps) {
  function diff(req: number, have: number | undefined): number {
    if (have === undefined) return 0;
    return Math.max(0, req - have);
  }

  const metalLack    = diff(metal, available?.metal);
  const siliconLack  = diff(silicon, available?.silicon);
  const hydrogenLack = diff(hydrogen, available?.hydrogen);

  return (
    <table className="table_no_background" cellSpacing={0} cellPadding={0} title="Требуется">
      <tbody>
        {metal > 0 && (
          <tr>
            <td>Металл</td>
            <td className={metalLack > 0 ? 'notavailable' : 'true'}>
              {formatNumber(metal)}
            </td>
            <td>{metalLack > 0 && <>({formatNumber(metalLack)})</>}</td>
          </tr>
        )}
        {silicon > 0 && (
          <tr>
            <td>Кремний</td>
            <td className={siliconLack > 0 ? 'notavailable' : 'true'}>
              {formatNumber(silicon)}
            </td>
            <td>{siliconLack > 0 && <>({formatNumber(siliconLack)})</>}</td>
          </tr>
        )}
        {hydrogen > 0 && (
          <tr>
            <td>Водород</td>
            <td className={hydrogenLack > 0 ? 'notavailable' : 'true'}>
              {formatNumber(hydrogen)}
            </td>
            <td>{hydrogenLack > 0 && <>({formatNumber(hydrogenLack)})</>}</td>
          </tr>
        )}
        <tr>
          <td>Время</td>
          <td colSpan={2}>{formatDuration(seconds)}</td>
        </tr>
      </tbody>
    </table>
  );
}
