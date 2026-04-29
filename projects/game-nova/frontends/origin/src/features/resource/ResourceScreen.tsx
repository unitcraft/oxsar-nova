// S-R01 Resource — экран производства ресурсов (план 72.1).
//
// Pixel-perfect клон legacy Resource.class.php:
//   - Таблица зданий: название, уровень, металл/кремний/водород (ч/д/н),
//     потребление энергии, коэффициент производства (factor %).
//   - Строки «Базовое производство», «Склад», «Итого».
//   - Итоговая сводка (в/сут, в/нед) под таблицей.
//   - Ввод factor для зданий с allow_factor=true + кнопка «Сохранить».
//
// Endpoints:
//   GET  /api/planets/{id}/resource-report
//   POST /api/planets/{id}/resource-update

import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { fetchResourceReport, updateResourceFactors } from '@/api/resource';
import { QK } from '@/api/query-keys';
import { useResolvedPlanet } from '@/features/common/useResolvedPlanet';
import { formatNumber } from '@/lib/format';
import type { ResourceBuilding } from '@/api/types';

function fmtProd(val: number): string {
  if (val === 0) return '0';
  if (val < 0) return formatNumber(Math.round(val));
  return formatNumber(Math.round(val));
}

interface FactorCellProps {
  building: ResourceBuilding;
  localFactor: number;
  onChange: (unitId: number, val: number) => void;
}

function FactorCell({ building, localFactor, onChange }: FactorCellProps) {
  if (!building.allow_factor) {
    return <td className="center">{building.factor}%</td>;
  }
  return (
    <td className="center">
      <input
        type="number"
        min={0}
        max={100}
        value={localFactor}
        onChange={(e) => {
          const v = Math.min(100, Math.max(0, Number(e.target.value) || 0));
          onChange(building.unit_id, v);
        }}
        style={{ width: 50, textAlign: 'center' }}
      />
      %
    </td>
  );
}

export function ResourceScreen() {
  const { planetId, isLoading: planetLoading } = useResolvedPlanet();
  const qc = useQueryClient();

  const reportQ = useQuery({
    queryKey: planetId ? QK.resourceReport(planetId) : ['noop-rr'],
    queryFn: () => (planetId ? fetchResourceReport(planetId) : Promise.reject()),
    enabled: planetId !== null,
  });

  const [factors, setFactors] = useState<Record<number, number>>({});

  const report = reportQ.data;

  // Инициализируем локальный state факторов из данных сервера при первой загрузке
  const localFactors: Record<number, number> = {};
  if (report) {
    for (const b of report.buildings) {
      localFactors[b.unit_id] = factors[b.unit_id] ?? b.factor;
    }
  }

  const save = useMutation({
    mutationFn: () => {
      const payload: Record<string, number> = {};
      for (const [id, val] of Object.entries(localFactors)) {
        payload[id] = val;
      }
      return updateResourceFactors(planetId!, payload);
    },
    onSuccess: () => {
      if (planetId) {
        setFactors({});
        void qc.invalidateQueries({ queryKey: QK.resourceReport(planetId) });
        void qc.invalidateQueries({ queryKey: QK.planet(planetId) });
      }
    },
  });

  if (planetLoading || reportQ.isLoading) {
    return <div className="idiv">…</div>;
  }
  if (!planetId || !report) {
    return <div className="idiv">—</div>;
  }

  const hasFactorBuildings = report.buildings.some((b) => b.allow_factor);

  return (
    <>
      <table className="ntable" style={{ width: '100%' }}>
        <thead>
          <tr>
            <th>Здание</th>
            <th>Ур.</th>
            <th>Металл/ч</th>
            <th>Кремний/ч</th>
            <th>Водород/ч</th>
            <th>Энергия</th>
            <th>Фактор</th>
          </tr>
        </thead>
        <tbody>
          {/* Здания */}
          {report.buildings.map((b) => (
            <tr key={b.unit_id}>
              <td>{b.name}</td>
              <td className="center">{b.level}</td>
              <td className="center">{fmtProd(b.prod_metal)}</td>
              <td className="center">{fmtProd(b.prod_silicon)}</td>
              <td className="center">{fmtProd(b.prod_hydrogen)}</td>
              <td className="center">
                {b.cons_energy !== 0 ? fmtProd(-Math.abs(b.cons_energy)) : '0'}
              </td>
              <FactorCell
                building={b}
                localFactor={localFactors[b.unit_id] ?? b.factor}
                onChange={(id, val) =>
                  setFactors((prev) => ({ ...prev, [id]: val }))
                }
              />
            </tr>
          ))}

          {/* Базовое производство */}
          <tr>
            <td colSpan={2}><b>Базовое производство</b></td>
            <td className="center">{fmtProd(report.basic_metal)}</td>
            <td className="center">{fmtProd(report.basic_silicon)}</td>
            <td className="center">{fmtProd(report.basic_hydrogen)}</td>
            <td className="center">—</td>
            <td className="center">—</td>
          </tr>

          {/* Склад */}
          <tr>
            <td colSpan={2}><b>Склад</b></td>
            <td className="center">{formatNumber(Math.round(report.storage_metal))}</td>
            <td className="center">{formatNumber(Math.round(report.storage_silicon))}</td>
            <td className="center">{formatNumber(Math.round(report.storage_hydrogen))}</td>
            <td className="center">—</td>
            <td className="center">—</td>
          </tr>

          {/* Итого */}
          <tr>
            <td colSpan={2}><b>Итого в час</b></td>
            <td className="center"><b>{fmtProd(report.metal_per_hour)}</b></td>
            <td className="center"><b>{fmtProd(report.silicon_per_hour)}</b></td>
            <td className="center"><b>{fmtProd(report.hydrogen_per_hour)}</b></td>
            <td className="center">—</td>
            <td className="center">—</td>
          </tr>
        </tbody>
      </table>

      {/* Сводка: в день и в неделю */}
      <table className="ntable" style={{ width: '100%', marginTop: 8 }}>
        <thead>
          <tr>
            <th>Период</th>
            <th>Металл</th>
            <th>Кремний</th>
            <th>Водород</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>В сутки</td>
            <td className="center">{fmtProd(report.metal_per_day)}</td>
            <td className="center">{fmtProd(report.silicon_per_day)}</td>
            <td className="center">{fmtProd(report.hydrogen_per_day)}</td>
          </tr>
          <tr>
            <td>В неделю</td>
            <td className="center">{fmtProd(report.metal_per_week)}</td>
            <td className="center">{fmtProd(report.silicon_per_week)}</td>
            <td className="center">{fmtProd(report.hydrogen_per_week)}</td>
          </tr>
        </tbody>
      </table>

      {hasFactorBuildings && (
        <div style={{ marginTop: 8, textAlign: 'center' }}>
          <button
            type="button"
            className="button"
            onClick={() => save.mutate()}
            disabled={save.isPending}
          >
            {save.isPending ? 'Сохранение…' : 'Сохранить производство'}
          </button>
          {save.isSuccess && (
            <span style={{ marginLeft: 8, color: 'green' }}>Сохранено</span>
          )}
          {save.isError && (
            <span style={{ marginLeft: 8, color: 'red' }}>Ошибка сохранения</span>
          )}
        </div>
      )}
    </>
  );
}
