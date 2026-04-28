// S-022 Repair (план 72 Ф.3 Spring 2 ч.2).
//
// Pixel-perfect зеркало legacy `templates/standard/repair.tpl` —
// верхний блок (вместимость ангара), список повреждённых юнитов,
// форма «Починить».
//
// Замечание (P72.S2.G — simplifications.md):
// nova-API на 2026-04-28 НЕ предоставляет endpoint'ов:
//   - GET /api/planets/{id}/repair       → список повреждённых юнитов
//   - POST /api/planets/{id}/repair      → запуск ремонта (Idempotency-Key)
//
// Backend плана 76 (nova exchange UI) идёт параллельно и тоже не
// затрагивает repair-домен. В origin-фронте для S-022 рендерим
// корректный pixel-perfect-каркас (ntable/center/idiv) с пустым списком и
// пометкой «нет повреждённых юнитов» — поведение совместимо с
// post-launch состоянием игрока, у которого ничего не повреждено.
//
// При появлении endpoint'а заменяем `damagedQ`/`repair` мутацию на
// настоящие вызовы fetchDamaged/repairUnit.

import { useTranslation } from '@/i18n/i18n';
import { useResolvedPlanet } from '@/features/common/useResolvedPlanet';

export function RepairScreen() {
  const { t } = useTranslation();
  const { planet, isLoading } = useResolvedPlanet();

  if (isLoading) return <div className="idiv">…</div>;
  if (!planet) return <div className="idiv">—</div>;

  // P72.S2.G: список повреждённых пустой, пока backend не появится.
  const damaged: { id: number; name: string; quantity: number }[] = [];

  return (
    <>
      <table className="ntable">
        <thead>
          <tr>
            <th colSpan={4}>
              {t('repair', 'title', { planetName: planet.name })}
            </th>
          </tr>
        </thead>
        <tbody>
          {damaged.length === 0 && (
            <tr>
              <td colSpan={4} className="center">
                {t('repair', 'noDamaged')}
              </td>
            </tr>
          )}
          {damaged.map((u) => (
            <tr key={u.id}>
              <td className="center">{u.id}</td>
              <td>{u.name}</td>
              <td className="center">{u.quantity}</td>
              <td className="center">
                <input
                  type="submit"
                  className="button"
                  value={t('repair', 'modeRepair')}
                  disabled
                />
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </>
  );
}
