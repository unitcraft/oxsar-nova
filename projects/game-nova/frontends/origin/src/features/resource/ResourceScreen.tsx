// S-R01 Resource — экран производства ресурсов (план 72.1 ч.19).
// Pixel-perfect клон legacy resource.tpl.

import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { fetchResourceReport, updateResourceFactors } from '@/api/resource';
import { QK } from '@/api/query-keys';
import { useResolvedPlanet } from '@/features/common/useResolvedPlanet';
import { useTranslation } from '@/i18n/i18n';
import { formatNumber } from '@/lib/format';
import type { ResourceBuilding } from '@/api/types';

function fmt(val: number): string {
  if (val === 0) return '0';
  return formatNumber(Math.round(val));
}

function signClass(val: number): string {
  if (val > 0) return 'true';
  if (val < 0) return 'false';
  return '';
}

interface FactorInputProps {
  building: ResourceBuilding;
  value: number;
  onChange: (unitId: number, val: number) => void;
}

function FactorInput({ building, value, onChange }: FactorInputProps) {
  if (!building.allow_factor) return <>&nbsp;</>;
  return (
    <>
      <input
        type="text"
        name={String(building.unit_id)}
        id={`factor_${building.unit_id}`}
        value={value}
        maxLength={3}
        size={3}
        onChange={(e) => {
          const v = Math.min(100, Math.max(0, Number(e.target.value) || 0));
          onChange(building.unit_id, v);
        }}
      />
      %{' '}
      <select
        onChange={(e) => {
          const v = Number(e.target.value);
          if (!isNaN(v)) onChange(building.unit_id, v);
        }}
        defaultValue="none"
      >
        <option value="none" className="center">-</option>
        <option value="0">0%</option>
        <option value="25">25%</option>
        <option value="50">50%</option>
        <option value="75">75%</option>
        <option value="100">100%</option>
      </select>
    </>
  );
}

export function ResourceScreen() {
  const { planetId, planet, isLoading: planetLoading } = useResolvedPlanet();
  const { t } = useTranslation();
  const qc = useQueryClient();
  const [localFactors, setLocalFactors] = useState<Record<number, number>>({});

  const reportQ = useQuery({
    queryKey: planetId ? QK.resourceReport(planetId) : ['noop-rr'],
    queryFn: () => (planetId ? fetchResourceReport(planetId) : Promise.reject()),
    enabled: planetId !== null,
  });

  const save = useMutation({
    mutationFn: () => {
      const payload: Record<string, number> = {};
      const report = reportQ.data;
      if (!report) return Promise.reject();
      for (const b of report.buildings) {
        payload[String(b.unit_id)] = localFactors[b.unit_id] ?? b.factor;
      }
      return updateResourceFactors(planetId!, payload);
    },
    onSuccess: () => {
      setLocalFactors({});
      if (planetId) {
        void qc.invalidateQueries({ queryKey: QK.resourceReport(planetId) });
        void qc.invalidateQueries({ queryKey: QK.planet(planetId) });
      }
    },
  });

  if (planetLoading || reportQ.isLoading) return <div className="idiv">…</div>;
  if (!planetId || !reportQ.data) return <div className="idiv">—</div>;

  const report = reportQ.data;
  const planetName = planet?.name ?? report.planet_name;

  const totalEnergy =
    (planet?.energy_remaining ?? 0) - (planet?.energy_prod ?? 0);

  function factorVal(b: ResourceBuilding): number {
    return localFactors[b.unit_id] ?? b.factor;
  }

  function buildingName(b: ResourceBuilding): string {
    // snake_case → camelCase: metal_mine→metalmine, silicon_lab→siliconLab
    const camel = b.name.replace(/_([a-z])/g, (_, c: string) => c.toUpperCase());
    const val = t('info', camel);
    if (!val.startsWith('[')) return val;
    // fallback: убрать _ без смены регистра (metal_mine→metalmine)
    const flat = b.name.replace(/_/g, '');
    const val2 = t('info', flat);
    if (!val2.startsWith('[')) return val2;
    return b.name;
  }

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        save.mutate();
      }}
    >
      <table className="ntable">
        <tbody>
          <tr>
            <th colSpan={6}>{t('resource', 'resourceProductionForPlanet')} {planetName}</th>
          </tr>
          <tr>
            <td></td>
            <td align="right"><b>Металл</b></td>
            <td align="right"><b>Кремний</b></td>
            <td align="right"><b>Водород</b></td>
            <td align="right"><b>Энергия</b></td>
            <td></td>
          </tr>

          {/* Базовое производство */}
          <tr>
            <td><b>{t('resource', 'natural')}</b></td>
            <td align="right">
              <span className={signClass(report.basic_metal)}>{fmt(report.basic_metal)}</span>
            </td>
            <td align="right">
              <span className={signClass(report.basic_silicon)}>{fmt(report.basic_silicon)}</span>
            </td>
            <td align="right">0</td>
            <td align="right">0</td>
            <td></td>
          </tr>

          {/* Здания */}
          {report.buildings
            .filter((b) => b.level > 0)
            .map((b) => (
              <tr key={b.unit_id}>
                <td><b>{buildingName(b)} ({b.level})</b></td>
                <td align="right">
                  {b.prod_metal > 0
                    ? <span className="true">{fmt(b.prod_metal)}</span>
                    : b.prod_metal < 0
                    ? <span className="false">{fmt(Math.abs(b.prod_metal))}</span>
                    : '0'}
                </td>
                <td align="right">
                  {b.prod_silicon > 0
                    ? <span className="true">{fmt(b.prod_silicon)}</span>
                    : b.prod_silicon < 0
                    ? <span className="false">{fmt(Math.abs(b.prod_silicon))}</span>
                    : '0'}
                </td>
                <td align="right">
                  {b.prod_hydrogen > 0
                    ? <span className="true">{fmt(b.prod_hydrogen)}</span>
                    : b.prod_hydrogen < 0
                    ? <span className="false">{fmt(Math.abs(b.prod_hydrogen))}</span>
                    : '0'}
                </td>
                <td align="right">
                  {b.cons_energy > 0
                    ? <span className="false">{fmt(b.cons_energy)}</span>
                    : b.cons_energy < 0
                    ? <span className="true">{fmt(Math.abs(b.cons_energy))}</span>
                    : '0'}
                </td>
                <td>
                  <FactorInput
                    building={b}
                    value={factorVal(b)}
                    onChange={(id, val) =>
                      setLocalFactors((prev) => ({ ...prev, [id]: val }))
                    }
                  />
                </td>
              </tr>
            ))}

          {/* Склад */}
          <tr>
            <td className="strongBorderTop"><b>{t('resource', 'storage')}</b></td>
            <td align="right" className="strongBorderTop">
              <span className="true">{fmt(report.storage_metal)}</span>
            </td>
            <td align="right" className="strongBorderTop">
              <span className="true">{fmt(report.storage_silicon)}</span>
            </td>
            <td align="right" className="strongBorderTop">
              <span className="true">{fmt(report.storage_hydrogen)}</span>
            </td>
            <td align="right" className="strongBorderTop">-</td>
            <td className="strongBorderTop">
              <input
                type="button"
                className="button"
                value={t('resource', 'shutDown') ?? 'Выключить'}
                onClick={() => {
                  const factors: Record<number, number> = {};
                  for (const b of report.buildings) {
                    if (b.allow_factor) factors[b.unit_id] = 0;
                  }
                  setLocalFactors(factors);
                }}
              />
            </td>
          </tr>

          {/* Итого в час */}
          <tr>
            <td><b>{t('resource', 'hourlyProduction')}</b></td>
            <td align="right">
              <span className={signClass(report.metal_per_hour)}>{fmt(report.metal_per_hour)}</span>
            </td>
            <td align="right">
              <span className={signClass(report.silicon_per_hour)}>{fmt(report.silicon_per_hour)}</span>
            </td>
            <td align="right">
              <span className={signClass(report.hydrogen_per_hour)}>{fmt(report.hydrogen_per_hour)}</span>
            </td>
            <td align="right">
              <span className={totalEnergy <= 0 ? 'false' : 'true'}>{fmt(totalEnergy)}</span>
            </td>
            <td>
              <input
                type="button"
                className="button"
                value={t('resource', 'startUp') ?? 'Включить'}
                onClick={() => {
                  const factors: Record<number, number> = {};
                  for (const b of report.buildings) {
                    if (b.allow_factor) factors[b.unit_id] = 100;
                  }
                  setLocalFactors(factors);
                }}
              />
            </td>
          </tr>

          {/* В сутки */}
          <tr>
            <td className="strongBorderTop"><b>{t('resource', 'dailyProduction')}</b></td>
            <td align="right" className="strongBorderTop">
              <span className={signClass(report.metal_per_day)}>{fmt(report.metal_per_day)}</span>
            </td>
            <td align="right" className="strongBorderTop">
              <span className={signClass(report.silicon_per_day)}>{fmt(report.silicon_per_day)}</span>
            </td>
            <td align="right" className="strongBorderTop">
              <span className={signClass(report.hydrogen_per_day)}>{fmt(report.hydrogen_per_day)}</span>
            </td>
            <td align="right" className="strongBorderTop">-</td>
            <td className="strongBorderTop">
              <input
                type="submit"
                name="update"
                value="Применить"
                className="button"
                disabled={save.isPending}
              />
            </td>
          </tr>

          {/* В неделю */}
          <tr>
            <td><b>{t('resource', 'weeklyProduction')}</b></td>
            <td align="right">
              <span className={signClass(report.metal_per_week)}>{fmt(report.metal_per_week)}</span>
            </td>
            <td align="right">
              <span className={signClass(report.silicon_per_week)}>{fmt(report.silicon_per_week)}</span>
            </td>
            <td align="right">
              <span className={signClass(report.hydrogen_per_week)}>{fmt(report.hydrogen_per_week)}</span>
            </td>
            <td align="right">-</td>
            <td>&nbsp;</td>
          </tr>
        </tbody>
      </table>
    </form>
  );
}
