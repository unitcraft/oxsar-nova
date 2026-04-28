// S-018 BuildingInfo (план 72 Ф.4 Spring 3).
//
// Pixel-perfect зеркало legacy `templates/standard/buildinginfo.tpl` —
// статическая страница описания здания. Берёт данные из реального
// catalog endpoint:
//
//   GET /api/buildings/catalog/{type}  → BuildingCatalogEntry
//
// Имя/описание выводятся через i18n.info.{key} (в bundle уже есть
// большинство значений, см. R12 переиспользование).
//
// Preview-таблица: уровень, стоимость, время постройки, производство/
// энергия (если применимо).

import { useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { fetchBuildingCatalog } from '@/api/catalog';
import { QK } from '@/api/query-keys';
import { useTranslation } from '@/i18n/i18n';
import { formatNumber, formatDuration } from '@/lib/format';

export function BuildingInfoScreen() {
  const params = useParams<{ type?: string }>();
  const type = params.type ?? '';
  const { t } = useTranslation();

  const q = useQuery({
    queryKey: QK.buildingCatalog(type),
    queryFn: () => fetchBuildingCatalog(type),
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
            <th>{name}</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>
              {hasFull ? <span>{fullDesc}</span> : hasDesc ? <span>{desc}</span> : <i>—</i>}
            </td>
          </tr>
        </tbody>
      </table>

      <table className="ntable">
        <thead>
          <tr>
            <th>{t('techtree', 'levelAbbr')}</th>
            <th>{t('techtree', 'kindBuildings')}</th>
            <th>{t('records', 'colRecord')}</th>
            <th>{t('battlestats', 'colDate')}</th>
          </tr>
        </thead>
        <tbody>
          {entry.preview.map((row) => (
            <tr key={row.level}>
              <td className="center">{row.level}</td>
              <td>
                {t('score', 'colMetal')}: {formatNumber(row.cost.metal)}
                {row.cost.silicon > 0 && (
                  <>
                    {' · '}
                    {t('score', 'colSilicon')}: {formatNumber(row.cost.silicon)}
                  </>
                )}
                {row.cost.hydrogen > 0 && (
                  <>
                    {' · '}
                    {t('score', 'colHydrogen')}: {formatNumber(row.cost.hydrogen)}
                  </>
                )}
              </td>
              <td className="center">
                {row.production_per_hour && row.production_per_hour > 0
                  ? formatNumber(Math.floor(row.production_per_hour))
                  : row.energy_output && row.energy_output > 0
                    ? `+${formatNumber(Math.floor(row.energy_output))}`
                    : row.energy_demand && row.energy_demand > 0
                      ? `−${formatNumber(Math.floor(row.energy_demand))}`
                      : '—'}
              </td>
              <td className="center">{formatDuration(row.build_seconds)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </>
  );
}
