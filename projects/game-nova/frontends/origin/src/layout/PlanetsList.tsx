// Список планет (правая колонка) origin-фронта.
//
// Pixel-perfect клон layout.tpl #planets + legacy style.css.
// <li class="goto"> кликается целиком (как jQuery .goto в legacy main.js).

import { useNavigate } from 'react-router-dom';
import { useResolvedPlanet } from '@/features/common/useResolvedPlanet';
import { useCurrentPlanetStore } from '@/stores/currentPlanet';
import { planetImageSmallUrl, moonImageSmallUrl } from '@/lib/planet-image';

export function PlanetsList() {
  const { planets, planetId } = useResolvedPlanet();
  const setCurrent = useCurrentPlanetStore((s) => s.set);
  const navigate = useNavigate();

  return (
    <div id="planets">
      <ul>
        {planets.map((p) => {
          const isMoon = p.is_moon === true;
          const isCurrent = p.id === planetId;
          const cls = isMoon
            ? isCurrent
              ? 'goto cur-moon'
              : 'goto moon-select'
            : isCurrent
              ? 'goto cur-planet'
              : 'goto';
          const imgSrc = isMoon
            ? moonImageSmallUrl()
            : planetImageSmallUrl(p.planet_type ?? null, p.id);
          const imgSize = isMoon ? 20 : 60;
          return (
            <li
              key={p.id}
              className={cls}
              onClick={() => {
                setCurrent(p.id);
                navigate('/');
              }}
            >
              <img src={imgSrc} alt={p.name} width={imgSize} height={imgSize} />
              <br />
              {p.name}
            </li>
          );
        })}
      </ul>
    </div>
  );
}
