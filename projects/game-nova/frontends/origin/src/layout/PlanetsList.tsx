// Список планет (правая колонка) origin-фронта.
//
// План 72 Ф.1 — каркас без данных.
// План 72 Ф.2 Spring 1 — подключено к /api/planets через
// useResolvedPlanet; клик меняет активную планету в Zustand-store
// и навигирует на главный экран. Pixel-perfect зеркало legacy
// `#planets` со списком имён планет.

import { useNavigate } from 'react-router-dom';
import { useResolvedPlanet } from '@/features/common/useResolvedPlanet';
import { useCurrentPlanetStore } from '@/stores/currentPlanet';
import { formatCoords } from '@/lib/format';

export function PlanetsList() {
  const { planets, planetId } = useResolvedPlanet();
  const setCurrent = useCurrentPlanetStore((s) => s.set);
  const navigate = useNavigate();

  return (
    <ul id="planets">
      {planets.map((p) => {
        const cls = p.is_moon ? 'cur-moon' : 'cur-planet';
        const active = p.id === planetId ? ' active' : '';
        return (
          <li key={p.id} className={`${cls}${active}`}>
            <button
              type="button"
              className="link-button"
              aria-label={p.name}
              onClick={() => {
                setCurrent(p.id);
                navigate('/');
              }}
            >
              {p.name}
              <br />
              <small>{formatCoords(p.galaxy, p.system, p.position)}</small>
            </button>
          </li>
        );
      })}
    </ul>
  );
}
