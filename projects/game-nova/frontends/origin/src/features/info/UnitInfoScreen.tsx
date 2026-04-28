// S-019 UnitInfo (план 72 Ф.4 Spring 3).
//
// Pixel-perfect зеркало legacy `templates/standard/unitinfo.tpl`:
// заголовок + описание + таблица боевых характеристик / груза /
// скорости / стоимости + блок rapidfire против других юнитов.
//
// Endpoint:
//   GET /api/units/catalog/{type}  → UnitCatalogEntry
// kind = ship | defense | research. Для research показываем
// preview-таблицу cost по уровням; для ship/defense — статы.

import { useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { fetchUnitCatalog } from '@/api/catalog';
import { QK } from '@/api/query-keys';
import { useTranslation } from '@/i18n/i18n';
import { formatNumber } from '@/lib/format';

export function UnitInfoScreen() {
  const params = useParams<{ type?: string }>();
  const type = params.type ?? '';
  const { t } = useTranslation();

  const q = useQuery({
    queryKey: QK.unitCatalog(type),
    queryFn: () => fetchUnitCatalog(type),
    enabled: type.length > 0,
    staleTime: 60 * 60 * 1000,
  });

  if (q.isLoading) return <div className="idiv">…</div>;
  if (q.isError || !q.data) {
    return (
      <table className="ntable">
        <tbody>
          <tr>
            <td className="center">
              <i>{t('alliance', 'nothing')}</i>
            </td>
          </tr>
        </tbody>
      </table>
    );
  }

  const entry = q.data;
  const nameKey = entry.key.replace(/_([a-z])/g, (_, c: string) => c.toUpperCase());
  const name = t('info', nameKey);
  const fullDescKey = `${nameKey}FullDesc`;
  const descKey = `${nameKey}Desc`;
  const fullDesc = t('info', fullDescKey);
  const desc = t('info', descKey);
  const hasFull = fullDesc !== `[info.${fullDescKey}]`;
  const hasDesc = desc !== `[info.${descKey}]`;

  return (
    <>
      <table className="ntable">
        <thead>
          <tr>
            <th colSpan={3}>{name}</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td colSpan={3}>
              {hasFull ? <span>{fullDesc}</span> : hasDesc ? <span>{desc}</span> : <i>—</i>}
            </td>
          </tr>
          {(entry.attack !== null && entry.attack !== undefined) && (
            <>
              <tr>
                <th colSpan={3}>{t('battlestats', 'title')}</th>
              </tr>
              <tr>
                <td>{t('battlestats', 'colResult') || 'Атака'}</td>
                <td colSpan={2}>{formatNumber(entry.attack)}</td>
              </tr>
              <tr>
                <td>{t('battlestats', 'colLoot') || 'Щит'}</td>
                <td colSpan={2}>{formatNumber(entry.shield ?? 0)}</td>
              </tr>
              <tr>
                <td>{t('battlestats', 'colDate') || 'Броня'}</td>
                <td colSpan={2}>{formatNumber(entry.shell ?? 0)}</td>
              </tr>
              {entry.speed !== null && entry.speed !== undefined && (
                <tr>
                  <td>{t('records', 'colRecord') || 'Скорость'}</td>
                  <td colSpan={2}>{formatNumber(entry.speed)}</td>
                </tr>
              )}
              {entry.cargo !== null && entry.cargo !== undefined && (
                <tr>
                  <td>{t('records', 'colHolder') || 'Груз'}</td>
                  <td colSpan={2}>{formatNumber(entry.cargo)}</td>
                </tr>
              )}
              {entry.fuel !== null && entry.fuel !== undefined && (
                <tr>
                  <td>{t('repair', 'fuelPer') || 'Топливо'}</td>
                  <td colSpan={2}>{formatNumber(entry.fuel)}</td>
                </tr>
              )}
            </>
          )}
          <tr>
            <th colSpan={3}>{t('records', 'colMine') || 'Стоимость'}</th>
          </tr>
          <tr>
            <td>{t('score', 'colMetal')}</td>
            <td>{t('score', 'colSilicon')}</td>
            <td>{t('score', 'colHydrogen')}</td>
          </tr>
          <tr>
            <td>{formatNumber(entry.cost.metal)}</td>
            <td>{formatNumber(entry.cost.silicon)}</td>
            <td>{formatNumber(entry.cost.hydrogen)}</td>
          </tr>
        </tbody>
      </table>

      {entry.kind === 'research' && entry.preview && entry.preview.length > 0 && (
        <table className="ntable">
          <thead>
            <tr>
              <th>{t('techtree', 'levelAbbr')}</th>
              <th>{t('score', 'colMetal')}</th>
              <th>{t('score', 'colSilicon')}</th>
              <th>{t('score', 'colHydrogen')}</th>
            </tr>
          </thead>
          <tbody>
            {entry.preview.map((row) => (
              <tr key={row.level}>
                <td className="center">{row.level}</td>
                <td>{formatNumber(row.cost.metal)}</td>
                <td>{formatNumber(row.cost.silicon)}</td>
                <td>{formatNumber(row.cost.hydrogen)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}

      {entry.rapidfire && entry.rapidfire.length > 0 && (
        <table className="ntable">
          <thead>
            <tr>
              <th colSpan={2}>Rapidfire</th>
            </tr>
          </thead>
          <tbody>
            {entry.rapidfire.map((rf) => (
              <tr key={rf.target_id}>
                <td>#{rf.target_id}</td>
                <td className="center">{rf.multiplier}×</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </>
  );
}
