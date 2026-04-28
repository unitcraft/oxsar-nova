// S-007 Empire — обзор всех планет игрока (план 72 Ф.2 Spring 1).
//
// Pixel-perfect зеркало legacy `src/templates/standard/empire.tpl`:
//   ntable с шапкой [№, Планета, Диаметр, Поля, Температура, УМИ, Ресурсы].
//
// Spring 1 покрывает только верхнюю таблицу планет. Нижние блоки
// `construction/shipyard/defense/moon/research` (по столбцу на каждую
// планету) требуют агрегированный endpoint `GET /api/empire/buildings`
// которого в openapi.yaml нет — записано в docs/simplifications.md.
//
// Поля Diameter / Fields / Temperature / reseachVirtLab отсутствуют в
// текущей `Planet` schema (см. openapi.yaml). Рендерим их как «—»
// до расширения schema (тоже фигурирует в simplifications.md).

import { useNavigate } from 'react-router-dom';
import { useResolvedPlanet } from '@/features/common/useResolvedPlanet';
import { useCurrentPlanetStore } from '@/stores/currentPlanet';
import { formatNumber, formatCoords } from '@/lib/format';
import { useTranslation } from '@/i18n/i18n';

export function EmpireScreen() {
  const { planets, isLoading } = useResolvedPlanet();
  const setCurrent = useCurrentPlanetStore((s) => s.set);
  const navigate = useNavigate();
  const { t } = useTranslation();

  if (isLoading) {
    return <div className="idiv">…</div>;
  }
  if (planets.length === 0) {
    return <div className="idiv">{t('overview', 'noPlanets')}</div>;
  }

  return (
    <table className="ntable">
      <thead>
        <tr>
          <th>№</th>
          <th>{t('empire', 'groupPlanet')}</th>
          <th>{t('empire', 'rowDiameter')}</th>
          <th>{t('empire', 'rowFields')}</th>
          <th>{t('empire', 'rowTemp')}</th>
          <th>УМИ</th>
          <th colSpan={2}>{t('empire', 'groupResources')}</th>
        </tr>
      </thead>
      <tbody>
        {planets.map((p, idx) => {
          const selected = p.is_moon ? false : true;
          return (
            <tr key={p.id}>
              <td>{idx + 1}.</td>
              <td>
                <button
                  type="button"
                  className="link-button"
                  onClick={() => {
                    setCurrent(p.id);
                    navigate('/');
                  }}
                  aria-label={`${p.name} ${formatCoords(p.galaxy, p.system, p.position)}`}
                >
                  {selected ? '* ' : ''}
                  {p.name}
                  <br />
                  {formatCoords(p.galaxy, p.system, p.position)}
                </button>
              </td>
              <td>—</td>
              <td>—</td>
              <td>—</td>
              <td>—</td>
              <td>
                М <br />К <br />В
              </td>
              <td align="right">
                {formatNumber(p.metal)}
                <br />
                {formatNumber(p.silicon)}
                <br />
                {formatNumber(p.hydrogen)}
              </td>
            </tr>
          );
        })}
      </tbody>
    </table>
  );
}
