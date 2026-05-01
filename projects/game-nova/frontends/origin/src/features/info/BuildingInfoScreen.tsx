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
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { fetchBuildingCatalog } from '@/api/catalog';
import {
  fetchBuildingsOverview,
  demolishBuilding,
  packBuilding,
} from '@/api/buildings';
import { QK } from '@/api/query-keys';
import { useTranslation } from '@/i18n/i18n';
import { formatNumber, formatDuration } from '@/lib/format';
import { findCatalog } from '@/features/common/catalog';
import { useResolvedPlanet } from '@/features/common/useResolvedPlanet';
import type { ApiError } from '@/api/client';
import { useState } from 'react';

export function BuildingInfoScreen() {
  const params = useParams<{ type?: string }>();
  const type = params.type ?? '';
  const { t } = useTranslation();
  const qc = useQueryClient();
  const { planetId } = useResolvedPlanet();
  const [demolishErr, setDemolishErr] = useState<string | null>(null);

  const q = useQuery({
    queryKey: QK.buildingCatalog(type),
    queryFn: () => fetchBuildingCatalog(type),
    enabled: type.length > 0,
    staleTime: 60 * 60 * 1000,
  });

  // План 72.1.33: текущий уровень здания нужен для demolish-button.
  const overviewQ = useQuery({
    queryKey: planetId ? QK.buildingsOverview(planetId) : ['noop-bo'],
    queryFn: () =>
      planetId ? fetchBuildingsOverview(planetId) : Promise.reject(),
    enabled: planetId !== null,
  });

  const demolishMut = useMutation({
    mutationFn: () => {
      if (!planetId || !q.data) return Promise.reject(new Error('no planet'));
      return demolishBuilding(planetId, q.data.id);
    },
    onSuccess: () => {
      setDemolishErr(null);
      if (planetId) {
        void qc.invalidateQueries({ queryKey: QK.buildingsOverview(planetId) });
        void qc.invalidateQueries({ queryKey: QK.planet(planetId) });
      }
    },
    onError: (e) => setDemolishErr((e as ApiError).message),
  });

  // План 72.1.33 ч.2: pack building через packing-артефакт.
  const [packErr, setPackErr] = useState<string | null>(null);
  const packBuildingMut = useMutation({
    mutationFn: () => {
      if (!planetId || !q.data) return Promise.reject(new Error('no planet'));
      return packBuilding(planetId, q.data.id);
    },
    onSuccess: () => {
      setPackErr(null);
      if (planetId) {
        void qc.invalidateQueries({ queryKey: QK.buildingsOverview(planetId) });
        void qc.invalidateQueries({ queryKey: QK.planet(planetId) });
        void qc.invalidateQueries({ queryKey: QK.artefacts() });
      }
    },
    onError: (e) => setPackErr((e as ApiError).message),
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
  // План 72.1.33: текущий уровень здания на planetId.
  const curLevel = overviewQ.data?.levels[String(entry.id)] ?? 0;
  // Demolish factor: catalog не отдаёт его в catalog endpoint, поэтому
  // показываем кнопку всегда если curLevel>0; backend проверит `spec.demolish > 0`
  // и вернёт ErrUnknownUnit если фактор не задан.
  const canDemolish = curLevel > 0;
  const fullDescKey = `${nameKey}FullDesc`;
  const descKey = `${nameKey}Desc`;
  const fullDesc = t('info', fullDescKey);
  const desc = t('info', descKey);
  const hasFull = fullDesc !== `[info.${fullDescKey}]`;
  const hasDesc = desc !== `[info.${descKey}]`;
  // План 72.1.23: legacy `BuildingInfo::index` ставит `building_image`
  // через Image::getImage(getUnitImage(name)). Origin не показывал.
  const catalog = findCatalog(entry.id);

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
              {catalog && (
                <img
                  src={`/assets/origin/images/units/${catalog.icon}.gif`}
                  alt={name}
                  width={120}
                  height={120}
                  style={{ float: 'left', marginRight: 8 }}
                  onError={(e) => {
                    (e.target as HTMLImageElement).style.display = 'none';
                  }}
                />
              )}
              {hasFull ? <span>{fullDesc}</span> : hasDesc ? <span>{desc}</span> : <i>—</i>}
            </td>
          </tr>
        </tbody>
      </table>

      {/* План 72.1.33: demolish secton (legacy BuildingInfo::DEMOLISH_NOW). */}
      {canDemolish && planetId && (
        <table className="ntable">
          <thead>
            <tr>
              <th>{t('buildinginfo', 'demolish') || 'Снос здания'}</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td className="center">
                <p>
                  {t('buildinginfo', 'currentLevel') || 'Текущий уровень'}:{' '}
                  <b>{curLevel}</b>{' '}
                  →{' '}
                  <b>{curLevel - 1}</b>
                </p>
                <button
                  type="button"
                  className="button"
                  disabled={demolishMut.isPending}
                  onClick={() => {
                    if (
                      window.confirm(
                        (t('buildinginfo', 'demolishConfirm') as string) ||
                          `Снести ${name} до уровня ${curLevel - 1}?`,
                      )
                    ) {
                      demolishMut.mutate();
                    }
                  }}
                >
                  {demolishMut.isPending
                    ? '…'
                    : t('buildinginfo', 'demolishNow') || 'Снести сейчас'}
                </button>
                {demolishErr && (
                  <div>
                    <span className="false">{demolishErr}</span>
                  </div>
                )}

                {/* План 72.1.33 ч.2: pack-building через packing-артефакт. */}
                <p style={{ marginTop: '0.8em' }}>
                  <button
                    type="button"
                    className="button"
                    disabled={packBuildingMut.isPending}
                    onClick={() => {
                      if (
                        window.confirm(
                          (t('buildinginfo', 'packConfirm') as string) ||
                            `Упаковать ${name} в артефакт (нужен packing-артефакт на этой планете)?`,
                        )
                      ) {
                        packBuildingMut.mutate();
                      }
                    }}
                  >
                    {packBuildingMut.isPending
                      ? '…'
                      : t('buildinginfo', 'packBuilding') || '📦 Упаковать здание'}
                  </button>
                  {packErr && (
                    <div>
                      <span className="false">{packErr}</span>
                    </div>
                  )}
                </p>
              </td>
            </tr>
          </tbody>
        </table>
      )}

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
