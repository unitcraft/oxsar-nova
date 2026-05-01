// S-R01 Resource — экран производства ресурсов (план 72.1 ч.19).
// Pixel-perfect клон legacy resource.tpl.

import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { fetchResourceReport, updateResourceFactors } from '@/api/resource';
import { fetchMe } from '@/api/me';
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
  disabled?: boolean;
}

function FactorInput({ building, value, onChange, disabled }: FactorInputProps) {
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
        disabled={disabled}
        onChange={(e) => {
          const v = Math.min(100, Math.max(0, Number(e.target.value) || 0));
          onChange(building.unit_id, v);
        }}
      />
      %{' '}
      <select
        disabled={disabled}
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

  // План 72.1.26: legacy `Resource.class.php` строка 33 блокирует
  // POST update при umode (`if(!NS::getUser()->get("umode"))`).
  // umode определяется наличием vacation_since в /api/me.
  const meQ = useQuery({ queryKey: QK.me(), queryFn: fetchMe });
  const umode = meQ.data?.vacation_since != null;

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
    // План 72.1.26 ч.B: virt/halting/solar — отдельные имена из resource.*.
    if (b.kind === 'fleet') return t('resource', 'fleetConsumption') || 'Флот';
    if (b.kind === 'stock_fleet')
      return t('resource', 'fleetStockConsumption') || 'Лоты биржи';
    if (b.kind === 'defense') return t('resource', 'defenseConsumption') || 'Оборона';
    if (b.kind === 'halting') {
      const tpl =
        t('resource', 'haltingConsumption', {
          planet: b.halting_from_coord ?? '',
          fleet: '',
        }) || `Удержание с ${b.halting_from_coord ?? '?'}`;
      return tpl;
    }

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

  // План 72.1.26 ч.B: virt/halting/solar — это «не здания», у них нет
  // фильтра allow_factor. Показываем все строки с непустым уровнем/счётчиком.
  function shouldShowRow(b: ResourceBuilding): boolean {
    if (b.level <= 0 && (b.kind === 'building' || b.kind == null)) return false;
    if (b.kind === 'building' || b.kind == null) {
      return b.level > 0 && b.allow_factor;
    }
    // solar / fleet / stock_fleet / defense / halting → всегда показываем
    // если есть потребление или производство (level=count>0).
    return b.level > 0;
  }

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        if (umode) return; // План 72.1.26: legacy блокировка POST в umode.
        save.mutate();
      }}
    >
      {umode && (
        <div className="false" style={{ padding: '0.5em', textAlign: 'center' }}>
          {t('resource', 'umodeWarning') ||
            'Изменение факторов производства недоступно в режиме отпуска.'}
        </div>
      )}
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

          {/* Здания + virt/halting/solar (план 72.1.26 ч.B). */}
          {report.buildings.filter(shouldShowRow).map((b) => {
            // Net-значения по ресурсу = prod − cons (legacy показывает
            // потребление красным, производство — зелёным).
            const netMetal = (b.prod_metal ?? 0) - (b.cons_metal ?? 0);
            const netSilicon = (b.prod_silicon ?? 0) - (b.cons_silicon ?? 0);
            const netHydrogen = (b.prod_hydrogen ?? 0) - (b.cons_hydrogen ?? 0);
            return (
              <tr key={b.unit_id} title={b.helptip ?? undefined}>
                <td>
                  <b>
                    {buildingName(b)}
                    {b.kind === 'building' || b.kind == null ? ` (${b.level})` : ''}
                  </b>
                </td>
                <td align="right">
                  {netMetal > 0 ? (
                    <span className="true">{fmt(netMetal)}</span>
                  ) : netMetal < 0 ? (
                    <span className="false">{fmt(Math.abs(netMetal))}</span>
                  ) : (
                    '0'
                  )}
                </td>
                <td align="right">
                  {netSilicon > 0 ? (
                    <span className="true">{fmt(netSilicon)}</span>
                  ) : netSilicon < 0 ? (
                    <span className="false">{fmt(Math.abs(netSilicon))}</span>
                  ) : (
                    '0'
                  )}
                </td>
                <td align="right">
                  {netHydrogen > 0 ? (
                    <span className="true">{fmt(netHydrogen)}</span>
                  ) : netHydrogen < 0 ? (
                    <span className="false">{fmt(Math.abs(netHydrogen))}</span>
                  ) : (
                    '0'
                  )}
                </td>
                <td align="right">
                  {b.cons_energy > 0 ? (
                    <span className="false">{fmt(b.cons_energy)}</span>
                  ) : b.cons_energy < 0 ? (
                    <span className="true">{fmt(Math.abs(b.cons_energy))}</span>
                  ) : (
                    '0'
                  )}
                </td>
                <td>
                  <FactorInput
                    building={b}
                    value={factorVal(b)}
                    disabled={umode}
                    onChange={(id, val) =>
                      setLocalFactors((prev) => ({ ...prev, [id]: val }))
                    }
                  />
                </td>
              </tr>
            );
          })}

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
                disabled={umode}
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
                disabled={umode}
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
                disabled={save.isPending || umode}
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

          {/* В месяц (план 72.1.26 — legacy resource.tpl: *720). */}
          <tr>
            <td><b>{t('resource', 'monthlyProduction')}</b></td>
            <td align="right">
              <span className={signClass(report.metal_per_month)}>{fmt(report.metal_per_month)}</span>
            </td>
            <td align="right">
              <span className={signClass(report.silicon_per_month)}>{fmt(report.silicon_per_month)}</span>
            </td>
            <td align="right">
              <span className={signClass(report.hydrogen_per_month)}>{fmt(report.hydrogen_per_month)}</span>
            </td>
            <td align="right">-</td>
            <td>&nbsp;</td>
          </tr>
        </tbody>
      </table>
    </form>
  );
}
