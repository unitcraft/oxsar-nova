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
import { useCurrentPlanetStore } from '@/stores/currentPlanet';
import type { GalaxyCell } from '@/api/types';

// План 72.1.32: legacy `Galaxy.class.php` строки 230-240 — маркер
// активности. (*) если <15 минут, (N min) если <1ч, '' иначе. Скрыт
// если canMonitorActivity=false или это сам игрок.
function activityLabel(lastSeen: string | null | undefined): string {
  if (!lastSeen) return '';
  const ms = Date.now() - new Date(lastSeen).getTime();
  if (ms < 0) return '';
  if (ms < 15 * 60_000) return '(*)';
  if (ms < 60 * 60_000) return `(${Math.floor(ms / 60_000)} min)`;
  return '';
}

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
  // План 72.1.24: пробрасываем current planet id для legacy-cost 10H.
  const currentPlanetId = useCurrentPlanetStore((s) => s.planetId);
  const [galaxy, setGalaxy] = useState(() =>
    clamp(Number(gParam ?? 1), GALAXY_MIN, GALAXY_MAX),
  );
  const [system, setSystem] = useState(() =>
    clamp(Number(sParam ?? 1), SYSTEM_MIN, SYSTEM_MAX),
  );

  const q = useQuery({
    queryKey: QK.galaxy(galaxy, system),
    queryFn: () =>
      fetchSystem(galaxy, system, currentPlanetId ?? undefined),
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
          {q.data?.cells.map((cell: GalaxyCell) => {
            // План 72.1.32: метка активности видна только если viewer
            // имеет star_surveillance и target в радиусе.
            const showActivity =
              q.data?.can_monitor_activity === true &&
              cell.owner_id != null;
            const actLabel = showActivity
              ? activityLabel(cell.owner_last_seen)
              : '';
            // In-missile range: показываем «🚀» action для атаки ракетой.
            const canRocket =
              q.data?.in_missile_range === true &&
              cell.has_planet &&
              cell.owner_id != null;
            return (
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
                <td>
                  {cell.owner_username ?? ''}
                  {cell.owner_username && (cell.owner_rank != null || cell.owner_e_points != null) && (
                    <>
                      <br />
                      <small>
                        {cell.owner_rank != null && <>#{cell.owner_rank}</>}
                        {cell.owner_e_points != null && (
                          <>
                            {cell.owner_rank != null && ' · '}
                            E:{formatNumber(cell.owner_e_points)}
                          </>
                        )}
                        {actLabel && <> · <span className="true">{actLabel}</span></>}
                        {cell.owner_vacation && <> · 💤</>}
                        {cell.owner_banned && <> · 🚫</>}
                        {cell.is_friend && <> · ⭐</>}
                      </small>
                    </>
                  )}
                </td>
                <td>
                  {cell.alliance_tag ? (
                    <>
                      [{cell.alliance_tag}]
                      {cell.owner_alliance_rank != null && (
                        <>
                          <br />
                          <small>#{cell.owner_alliance_rank}</small>
                        </>
                      )}
                      {cell.relation && (
                        <>
                          <br />
                          <small className={cell.relation === 'war' ? 'false' : 'true'}>
                            {cell.relation}
                          </small>
                        </>
                      )}
                    </>
                  ) : (
                    '—'
                  )}
                </td>
                <td>
                  {cell.debris_metal > 0 || cell.debris_silicon > 0
                    ? `${formatNumber(cell.debris_metal)} / ${formatNumber(cell.debris_silicon)}`
                    : ''}
                </td>
                <td>
                  {cell.has_planet && (
                    <>
                      <button
                        type="button"
                        className="button"
                        onClick={() =>
                          navigate(`/mission?g=${galaxy}&s=${system}&p=${cell.position}`)
                        }
                      >
                        {t('galaxy', 'attackTitle')}
                      </button>
                      {canRocket && (
                        <>
                          {' '}
                          <button
                            type="button"
                            className="button"
                            title={t('galaxy', 'rocketAttack') || 'Rocket attack'}
                            onClick={() =>
                              cell.planet_id != null
                                ? navigate(`/rocket-attack/${cell.planet_id}`)
                                : navigate(
                                    `/rocket-attack?g=${galaxy}&s=${system}&p=${cell.position}`,
                                  )
                            }
                          >
                            🚀
                          </button>
                        </>
                      )}
                    </>
                  )}
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}
