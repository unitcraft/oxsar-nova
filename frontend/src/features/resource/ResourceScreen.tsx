import { useEffect, useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { catalog } from '@/api/catalog';
import type { ResourceBuilding } from '@/api/types';
import { useToast } from '@/ui/Toast';
import { ResourceScreenSkeleton } from '@/ui/Skeleton';

interface FactorFormData {
  [unitId: string]: number;
}

export function ResourceScreen({
  planetId,
  onBack,
}: {
  planetId: string;
  onBack?: () => void;
}) {
  const queryClient = useQueryClient();
  const toast = useToast();
  const [factors, setFactors] = useState<FactorFormData>({});

  const { data: report, isLoading } = useQuery({
    queryKey: ['resource-report', planetId],
    queryFn: () => catalog.getResourceReport(planetId),
  });

  useEffect(() => {
    if (report?.buildings) {
      const initialFactors: FactorFormData = {};
      report.buildings.forEach((b) => {
        initialFactors[b.unit_id.toString()] = b.factor;
      });
      setFactors(initialFactors);
    }
  }, [report]);

  const updateMutation = useMutation({
    mutationFn: () => {
      const numericFactors: { [unitId: string]: number } = {};
      Object.entries(factors).forEach(([unitId, factor]) => {
        numericFactors[unitId] = Math.max(0, Math.min(100, factor));
      });
      return catalog.updateResourceFactors(planetId, {
        factors: numericFactors,
      });
    },
    onMutate: async () => {
      // Отменить все pending запросы для этого ключа
      await queryClient.cancelQueries({
        queryKey: ['resource-report', planetId],
      });

      // Сохранить старые данные для rollback'а при ошибке
      const previousReport = queryClient.getQueryData<any>([
        'resource-report',
        planetId,
      ]);

      // Оптимистично обновить локальный кэш
      if (previousReport) {
        const updatedReport = {
          ...previousReport,
          buildings: previousReport.buildings.map((b: any) => ({
            ...b,
            factor: factors[b.unit_id.toString()] ?? b.factor,
          })),
        };
        queryClient.setQueryData(['resource-report', planetId], updatedReport);
      }

      return { previousReport };
    },
    onError: (_err, _variables, context) => {
      // Откатить на старые данные при ошибке
      if (context?.previousReport) {
        queryClient.setQueryData(
          ['resource-report', planetId],
          context.previousReport
        );
      }
    },
    onSuccess: () => {
      // Переопрос чтобы убедиться что данные синхронизированы
      queryClient.invalidateQueries({
        queryKey: ['resource-report', planetId],
      });
      toast.show('success', 'Факторы сохранены');
    },
    onError: () => {
      toast.show('danger', 'Ошибка сохранения');
    },
  });

  const handleFactorChange = (unitId: string, value: number) => {
    setFactors((prev) => ({
      ...prev,
      [unitId]: Math.max(0, Math.min(100, value)),
    }));
  };

  const handleQuickSet = (unitId: string, value: number) => {
    handleFactorChange(unitId, value);
  };

  if (isLoading) {
    return <ResourceScreenSkeleton />;
  }

  if (!report) {
    return <div className="alert alert-error">Failed to load resource report</div>;
  }

  const productionBuildings = report.buildings.filter((b) => b.allow_factor);

  return (
    <div className="space-y-6 pb-20">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold">Управление ресурсами</h2>
        {onBack && (
          <button className="btn btn-sm btn-ghost" onClick={onBack}>
            ← Назад
          </button>
        )}
      </div>

      <div className="grid grid-cols-3 gap-4">
        <div className="card bg-base-200">
          <div className="card-body p-4">
            <div className="text-sm opacity-70">Хранилище Металл</div>
            <div className="text-xl font-bold">{Math.floor(report.metal_total)}</div>
            <div className="text-xs opacity-50">
              +{report.metal_per_hour.toFixed(2)}/ч
            </div>
          </div>
        </div>
        <div className="card bg-base-200">
          <div className="card-body p-4">
            <div className="text-sm opacity-70">Хранилище Кремний</div>
            <div className="text-xl font-bold">{Math.floor(report.silicon_total)}</div>
            <div className="text-xs opacity-50">
              +{report.silicon_per_hour.toFixed(2)}/ч
            </div>
          </div>
        </div>
        <div className="card bg-base-200">
          <div className="card-body p-4">
            <div className="text-sm opacity-70">Хранилище Водород</div>
            <div className="text-xl font-bold">{Math.floor(report.hydrogen_total)}</div>
            <div className="text-xs opacity-50">
              +{report.hydrogen_per_hour.toFixed(2)}/ч
            </div>
          </div>
        </div>
      </div>

      <div className="card bg-base-100 shadow">
        <div className="card-body">
          <h3 className="card-title text-lg">Производство (в час)</h3>
          <div className="grid grid-cols-3 gap-4 mt-4">
            <div>
              <div className="text-sm opacity-70">Металл</div>
              <div className="text-lg font-semibold text-success">
                +{report.metal_per_hour.toFixed(2)}
              </div>
            </div>
            <div>
              <div className="text-sm opacity-70">Кремний</div>
              <div className="text-lg font-semibold text-success">
                +{report.silicon_per_hour.toFixed(2)}
              </div>
            </div>
            <div>
              <div className="text-sm opacity-70">Водород</div>
              <div className="text-lg font-semibold text-success">
                +{report.hydrogen_per_hour.toFixed(2)}
              </div>
            </div>
          </div>
        </div>
      </div>

      {productionBuildings.length > 0 && (
        <div className="card bg-base-100 shadow">
          <div className="card-body">
            <h3 className="card-title text-lg">Корректировка производства</h3>
            <div className="space-y-4 mt-4">
              {productionBuildings.map((building) => (
                <BuildingFactorControl
                  key={building.unit_id}
                  building={building}
                  factor={factors[building.unit_id.toString()] ?? building.factor}
                  onChange={(value) =>
                    handleFactorChange(building.unit_id.toString(), value)
                  }
                  onQuickSet={(value) =>
                    handleQuickSet(building.unit_id.toString(), value)
                  }
                />
              ))}
            </div>

            <div className="mt-6 flex gap-2">
              <button
                className="btn btn-primary flex-1"
                disabled={updateMutation.isPending}
                onClick={() => updateMutation.mutate()}
              >
                {updateMutation.isPending ? (
                  <span className="loading loading-spinner loading-sm"></span>
                ) : (
                  'Сохранить'
                )}
              </button>
              <button
                className="btn btn-ghost flex-1"
                onClick={() => {
                  if (report.buildings) {
                    const reset: FactorFormData = {};
                    report.buildings.forEach((b) => {
                      reset[b.unit_id.toString()] = b.factor;
                    });
                    setFactors(reset);
                  }
                }}
              >
                Отменить
              </button>
            </div>
          </div>
        </div>
      )}

      {report.buildings.length > 0 && (
        <div className="card bg-base-100 shadow">
          <div className="card-body">
            <h3 className="card-title text-lg">Полный отчет</h3>
            <div className="overflow-x-auto mt-4">
              <table className="table table-sm">
                <thead>
                  <tr>
                    <th>Здание</th>
                    <th>Уровень</th>
                    <th>Металл/ч</th>
                    <th>Кремний/ч</th>
                    <th>Водород/ч</th>
                    <th>Энергия/ч</th>
                  </tr>
                </thead>
                <tbody>
                  {report.buildings.map((b) => (
                    <tr key={b.unit_id}>
                      <td className="font-semibold">{b.name}</td>
                      <td>{b.level}</td>
                      <td className="text-success">
                        +{(b.prod_metal * (factors[b.unit_id.toString()] ?? b.factor)) / 100).toFixed(2)}
                      </td>
                      <td className="text-success">
                        +{(b.prod_silicon * (factors[b.unit_id.toString()] ?? b.factor)) / 100).toFixed(2)}
                      </td>
                      <td className="text-success">
                        +{(b.prod_hydrogen * (factors[b.unit_id.toString()] ?? b.factor)) / 100).toFixed(2)}
                      </td>
                      <td className="text-error">
                        -{b.cons_energy.toFixed(2)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

function BuildingFactorControl({
  building,
  factor,
  onChange,
  onQuickSet,
}: {
  building: ResourceBuilding;
  factor: number;
  onChange: (value: number) => void;
  onQuickSet: (value: number) => void;
}) {
  return (
    <div className="border-b pb-4 last:border-b-0">
      <div className="flex items-center justify-between mb-2">
        <div>
          <div className="font-semibold">{building.name}</div>
          <div className="text-sm opacity-70">Уровень {building.level}</div>
        </div>
        <div className="text-right">
          <div className="text-lg font-bold">{factor}%</div>
          <div className="text-xs opacity-70">
            {building.prod_metal > 0 && (
              <>
                +{((building.prod_metal * factor) / 100).toFixed(1)} M{' '}
              </>
            )}
            {building.prod_silicon > 0 && (
              <>
                +{((building.prod_silicon * factor) / 100).toFixed(1)} S{' '}
              </>
            )}
            {building.prod_hydrogen > 0 && (
              <>
                +{((building.prod_hydrogen * factor) / 100).toFixed(1)} H
              </>
            )}
          </div>
        </div>
      </div>

      <div className="flex gap-2 mb-2">
        <input
          type="range"
          min="0"
          max="100"
          step="1"
          value={factor}
          onChange={(e) => onChange(parseInt(e.target.value, 10))}
          className="range range-sm flex-1"
        />
      </div>

      <div className="flex gap-1 text-xs">
        {[0, 25, 50, 75, 100].map((val) => (
          <button
            key={val}
            className={`px-2 py-1 rounded font-semibold transition-colors ${
              factor === val
                ? 'bg-primary text-primary-content'
                : 'bg-base-300 hover:bg-base-400'
            }`}
            onClick={() => onQuickSet(val)}
          >
            {val}%
          </button>
        ))}
      </div>
    </div>
  );
}
