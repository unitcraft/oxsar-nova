// S-RA Rocket Attack — ракетная атака (план 72.1 ч.20.5 + 72.1.35).
// Pixel-perfect клон legacy `rocket_attack.tpl`. Backend:
//   GET  /api/planets/{id}/rockets         — stock IPM.
//   POST /api/planets/{id}/rockets/launch  — запуск атаки.
//
// Legacy `RocketAttack.class.php`:
//  - quantity max = stock IPM (UNIT_INTERPLANETARY_ROCKET=52).
//  - primary_target — селект defense-юнитов (ID 43-49) +
//    «all» (target_unit_id=0).
//  - При успехе redirect на /Main; на фронте — toast + invalidate
//    rockets stock.

import { useNavigate, useParams, useSearchParams } from 'react-router-dom';
import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { fetchRocketStock, launchRocket } from '@/api/rocket';
import { fetchPlanet } from '@/api/planets';
import { useTranslation } from '@/i18n/i18n';
import { useResolvedPlanet } from '@/features/common/useResolvedPlanet';
import { catalogByGroup } from '@/features/common/catalog';
import type { ApiError } from '@/api/client';

export function RocketAttackScreen() {
  const { t } = useTranslation();
  const { planet, planetId: srcPlanetId } = useResolvedPlanet();
  const params = useParams<{ id?: string }>();
  const [search] = useSearchParams();
  const navigate = useNavigate();
  const qc = useQueryClient();

  // План 72.1.35: target — id планеты в URL `/rocket-attack/:id`,
  // либо координаты в query `?g=&s=&p=&moon=`.
  const targetPlanetId = params.id ?? '';
  const isMoon = search.get('moon') === '1' || search.get('moon') === 'true';

  // Координаты цели — приоритет: explicit query > подгрузка планеты по id.
  const queryG = search.get('g');
  const queryS = search.get('s');
  const queryP = search.get('p');

  // Если задан target id — подтянем координаты из /api/planets/{id}.
  const targetQ = useQuery({
    queryKey: ['target-planet', targetPlanetId],
    queryFn: () =>
      targetPlanetId ? fetchPlanet(targetPlanetId) : Promise.reject(new Error('no target')),
    enabled: !!targetPlanetId && !queryG,
  });

  const targetGalaxy = queryG
    ? Number(queryG)
    : targetQ.data?.galaxy ?? 0;
  const targetSystem = queryS
    ? Number(queryS)
    : targetQ.data?.system ?? 0;
  const targetPos = queryP
    ? Number(queryP)
    : targetQ.data?.position ?? 0;
  const targetCoords = `[${targetGalaxy}:${targetSystem}:${targetPos}${isMoon ? ' L' : ''}]`;

  // Stock IPM на исходной планете.
  const stockQ = useQuery({
    queryKey: ['rocket-stock', srcPlanetId],
    queryFn: () =>
      srcPlanetId ? fetchRocketStock(srcPlanetId) : Promise.reject(new Error('no src')),
    enabled: !!srcPlanetId,
  });
  const maxRockets = stockQ.data?.count ?? 0;

  const [quantity, setQuantity] = useState('0');
  const [target, setTarget] = useState('all');
  const [errMsg, setErrMsg] = useState<string | null>(null);

  const defenseUnits = catalogByGroup('defense');

  const launch = useMutation({
    mutationFn: () => {
      if (!srcPlanetId) return Promise.reject(new Error('no planet'));
      const cnt = Math.max(0, Math.min(maxRockets, Number(quantity) || 0));
      if (cnt <= 0) return Promise.reject(new Error('count must be > 0'));
      return launchRocket(srcPlanetId, {
        dst: {
          galaxy: targetGalaxy,
          system: targetSystem,
          position: targetPos,
          is_moon: isMoon,
        },
        count: cnt,
        target_unit_id: target === 'all' ? 0 : Number(target),
      });
    },
    onSuccess: () => {
      setErrMsg(null);
      if (srcPlanetId) {
        void qc.invalidateQueries({ queryKey: ['rocket-stock', srcPlanetId] });
        void qc.invalidateQueries({ queryKey: ['planets'] });
      }
      // Legacy редиректит на Main — у нас на главную.
      navigate('/');
    },
    onError: (e) => setErrMsg((e as ApiError).message),
  });

  if (!planet) return <div className="idiv">…</div>;

  return (
    <form
      method="post"
      action="#"
      onSubmit={(e) => {
        e.preventDefault();
        launch.mutate();
      }}
    >
      <table className="ntable">
        <tbody>
          <tr>
            <th colSpan={2}>
              {t('galaxy', 'rocketAttack') || 'Ракетная атака'}: {targetCoords}
            </th>
          </tr>
          <tr>
            <td>
              {t('galaxy', 'quantity') || 'Количество'} ({t('galaxy', 'max') || 'max'} {maxRockets})
            </td>
            <td>
              <input
                type="number"
                name="quantity"
                min={0}
                max={maxRockets}
                value={quantity}
                onChange={(e) => setQuantity(e.target.value)}
              />
            </td>
          </tr>
          <tr>
            <td>{t('galaxy', 'primaryTarget') || 'Основная задача'}</td>
            <td>
              <select
                name="target"
                value={target}
                onChange={(e) => setTarget(e.target.value)}
              >
                <option value="all">{t('galaxy', 'all') || 'Все'}</option>
                {defenseUnits
                  .filter(
                    // Legacy строки 135-138: исключаем interceptor (51) и
                    // interplanetary (52) ракеты — они не цель IPM.
                    (e) => e.id !== 51 && e.id !== 52,
                  )
                  .map((entry) => {
                    const [g, k] = entry.i18n.split('.') as [string, string];
                    return (
                      <option key={entry.id} value={String(entry.id)}>
                        {t(g, k)}
                      </option>
                    );
                  })}
              </select>
            </td>
          </tr>
          <tr>
            <td colSpan={2} className="center">
              <input
                type="submit"
                name="start"
                value={launch.isPending ? '…' : (t('mission', 'attacker') || 'Атаковать')}
                className="button"
                disabled={launch.isPending || maxRockets === 0}
              />
              {errMsg && (
                <div>
                  <span className="false">{errMsg}</span>
                </div>
              )}
            </td>
          </tr>
        </tbody>
      </table>
    </form>
  );
}
