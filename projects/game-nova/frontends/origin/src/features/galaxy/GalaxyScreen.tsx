// S-005 Galaxy — карта галактики (план 72 Ф.2 Spring 1).
//
// Pixel-perfect зеркало legacy `galaxy.tpl`:
//   - Верхняя строка: ввод galaxy/system + кнопки «‹‹» / «››».
//   - Таблица .galaxy-browser с 15 позициями системы:
//     # | Планета | Игрок | Альянс | Обломки | Действия
//   - На клик по строке — навигация в /mission/<srcPlanetId> с цели.
//
// Endpoint: GET /api/galaxy/{g}/{s} — возвращает 15 GalaxyCell.

import { useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { fetchSystem } from '@/api/galaxy';
import { QK } from '@/api/query-keys';
import { useTranslation } from '@/i18n/i18n';
import { formatNumber } from '@/lib/format';

const GALAXY_MIN = 1;
const GALAXY_MAX = 16;
const SYSTEM_MIN = 1;
const SYSTEM_MAX = 999;

function clamp(value: number, min: number, max: number): number {
  if (Number.isNaN(value)) return min;
  return Math.max(min, Math.min(max, value));
}

export function GalaxyScreen() {
  const { galaxy: gParam, system: sParam } = useParams();
  const navigate = useNavigate();
  const { t } = useTranslation();
  const [galaxy, setGalaxy] = useState(() =>
    clamp(Number(gParam ?? 1), GALAXY_MIN, GALAXY_MAX),
  );
  const [system, setSystem] = useState(() =>
    clamp(Number(sParam ?? 1), SYSTEM_MIN, SYSTEM_MAX),
  );

  const q = useQuery({
    queryKey: QK.galaxy(galaxy, system),
    queryFn: () => fetchSystem(galaxy, system),
  });

  function go(g: number, s: number) {
    const cg = clamp(g, GALAXY_MIN, GALAXY_MAX);
    const cs = clamp(s, SYSTEM_MIN, SYSTEM_MAX);
    setGalaxy(cg);
    setSystem(cs);
    navigate(`/galaxy/${cg}/${cs}`);
  }

  return (
    <div className="idiv">
      <table className="ntable galaxy-browser">
        <thead>
          <tr>
            <th colSpan={3}>{t('galaxy', 'galaxy')}</th>
            <th colSpan={3}>{t('galaxy', 'system')}</th>
          </tr>
          <tr>
            <td>
              <button
                type="button"
                className="button"
                aria-label={`${t('galaxy', 'galaxy')} prev`}
                onClick={() => go(galaxy - 1, system)}
              >
                «
              </button>
            </td>
            <td>
              <input
                type="number"
                className="center"
                min={GALAXY_MIN}
                max={GALAXY_MAX}
                value={galaxy}
                onChange={(e) =>
                  setGalaxy(clamp(Number(e.target.value), GALAXY_MIN, GALAXY_MAX))
                }
                onBlur={() => go(galaxy, system)}
                aria-label={t('galaxy', 'galaxy')}
              />
            </td>
            <td>
              <button
                type="button"
                className="button"
                aria-label={`${t('galaxy', 'galaxy')} next`}
                onClick={() => go(galaxy + 1, system)}
              >
                »
              </button>
            </td>
            <td>
              <button
                type="button"
                className="button"
                aria-label={`${t('galaxy', 'system')} prev`}
                onClick={() => go(galaxy, system - 1)}
              >
                «
              </button>
            </td>
            <td>
              <input
                type="number"
                className="center"
                min={SYSTEM_MIN}
                max={SYSTEM_MAX}
                value={system}
                onChange={(e) =>
                  setSystem(clamp(Number(e.target.value), SYSTEM_MIN, SYSTEM_MAX))
                }
                onBlur={() => go(galaxy, system)}
                aria-label={t('galaxy', 'system')}
              />
            </td>
            <td>
              <button
                type="button"
                className="button"
                aria-label={`${t('galaxy', 'system')} next`}
                onClick={() => go(galaxy, system + 1)}
              >
                »
              </button>
            </td>
          </tr>
        </thead>
      </table>

      <table className="ntable galaxy-browser">
        <thead>
          <tr>
            <th>{t('galaxy', 'colPos')}</th>
            <th>{t('galaxy', 'colPlanet')}</th>
            <th>{t('galaxy', 'colPlayer')}</th>
            <th>{t('galaxy', 'colAlliance')}</th>
            <th>{t('galaxy', 'colDebris')}</th>
            <th>{t('galaxy', 'actions')}</th>
          </tr>
        </thead>
        <tbody>
          {q.isLoading && (
            <tr>
              <td colSpan={6}>…</td>
            </tr>
          )}
          {q.error && (
            <tr>
              <td colSpan={6}>{(q.error as Error).message}</td>
            </tr>
          )}
          {q.data?.cells.map((cell) => (
            <tr key={cell.position}>
              <td>{cell.position}</td>
              <td>
                {cell.has_planet ? (cell.planet_name ?? '—') : ''}
                {cell.has_moon && (
                  <>
                    <br />
                    🌙 {cell.moon_name ?? t('galaxy', 'moonFallback')}
                  </>
                )}
              </td>
              <td>{cell.owner_username ?? ''}</td>
              <td>—</td>
              <td>
                {cell.debris_metal > 0 || cell.debris_silicon > 0
                  ? `${formatNumber(cell.debris_metal)} / ${formatNumber(cell.debris_silicon)}`
                  : ''}
              </td>
              <td>
                {cell.has_planet && (
                  <button
                    type="button"
                    className="button"
                    onClick={() =>
                      navigate(`/mission?g=${galaxy}&s=${system}&p=${cell.position}`)
                    }
                  >
                    {t('galaxy', 'attackTitle')}
                  </button>
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
