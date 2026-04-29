// Список планет (правая колонка) origin-фронта.
//
// План 72 Ф.1 — каркас без данных.
// План 72 Ф.2 Spring 1 — подключено к /api/planets через
// useResolvedPlanet; клик меняет активную планету в Zustand-store
// и навигирует на главный экран. Pixel-perfect зеркало legacy
// `#planets` — иконки 60×60 для планет, 20×20 для лун.

import { useNavigate } from 'react-router-dom';
import { useResolvedPlanet } from '@/features/common/useResolvedPlanet';
import { useCurrentPlanetStore } from '@/stores/currentPlanet';
import { planetImageSmallUrl, moonImageSmallUrl } from '@/lib/planet-image';

export function PlanetsList() {
  const { planets, planetId } = useResolvedPlanet();
  const setCurrent = useCurrentPlanetStore((s) => s.set);
  const navigate = useNavigate();

  return (
    <ul id="planets">
      {planets.map((p) => {
        const isMoon = p.is_moon === true;
        const isCurrent = p.id === planetId;
        const cls = isMoon
          ? isCurrent
            ? 'cur-moon'
            : 'moon-select'
          : isCurrent
            ? 'cur-planet'
            : '';
        const imgSrc = isMoon
          ? moonImageSmallUrl()
          : planetImageSmallUrl(p.planet_type ?? null, p.id);
        const imgSize = isMoon ? 20 : 60;
        return (
          <li key={p.id} className={cls || undefined}>
            <button
              type="button"
              className="link-button"
              aria-label={p.name}
              onClick={() => {
                setCurrent(p.id);
                navigate('/');
              }}
            >
              <img src={imgSrc} alt={p.name} width={imgSize} height={imgSize} />
              <br />
              {p.name}
            </button>
          </li>
        );
      })}
    </ul>
  );
}
