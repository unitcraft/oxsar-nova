// Top header origin-фронта.
//
// План 72.1 ч.16: pixel-perfect зеркало legacy `topHeader` (см.
// projects/game-legacy-php/src/templates/standard/before_content.tpl).
//
// Layout (5 колонок ressource + universe-name + username + lang + logout):
//   1. Металл  | 7.000.000   Склад: 7.000k
//   2. Кремний | 4.500.000           4.500k
//   3. Водород | 2.000.000           2.000k
//   4. Энергия | 0 (0)
//   5. Кредиты | 39.866,05   Пополнить
//
// Данные: первые 3 ресурса + энергия — из активной планеты
// (`useResolvedPlanet`). Кредиты — из `/api/me`.

import { useQuery } from '@tanstack/react-query';
import { useAuthStore } from '@/stores/auth';
import { useResolvedPlanet } from '@/features/common/useResolvedPlanet';
import { fetchMe } from '@/api/me';
import { QK } from '@/api/query-keys';
import { formatNumber } from '@/lib/format';
import { useTranslation } from '@/i18n/i18n';

function fmtCap(cap: number): string {
  if (cap >= 1000) {
    const k = Math.floor(cap / 1000);
    const tail = cap % 1000;
    if (tail === 0) return `${k}k`;
    return `${k}.${String(tail).padStart(3, '0')}k`;
  }
  return String(cap);
}

export function TopHeader() {
  const logout = useAuthStore((s) => s.logout);
  const { planet } = useResolvedPlanet();
  const { t } = useTranslation();
  const meQ = useQuery({
    queryKey: QK.me(),
    queryFn: fetchMe,
    staleTime: 30_000,
  });

  const credit = meQ.data?.credit ?? 0;
  const energyRemaining = planet?.energy_remaining ?? 0;
  const energyProd = planet?.energy_prod ?? 0;
  const storageLabel = t('overview', 'storageLabel');

  return (
    <div id="topHeader">
      <ul>
        <li className="ressource">
          <span>{t('overview', 'resMetalLabel')}</span>
          <br />
          {planet ? formatNumber(Math.floor(planet.metal)) : '—'}
          {planet && (
            <>
              <br />
              <small>
                {storageLabel} {fmtCap(planet.metal_cap)}
              </small>
            </>
          )}
        </li>
        <li className="ressource">
          <span>{t('overview', 'resSiliconLabel')}</span>
          <br />
          {planet ? formatNumber(Math.floor(planet.silicon)) : '—'}
          {planet && (
            <>
              <br />
              <small>{fmtCap(planet.silicon_cap)}</small>
            </>
          )}
        </li>
        <li className="ressource">
          <span>{t('overview', 'resHydrogenLabel')}</span>
          <br />
          {planet ? formatNumber(Math.floor(planet.hydrogen)) : '—'}
          {planet && (
            <>
              <br />
              <small>{fmtCap(planet.hydrogen_cap)}</small>
            </>
          )}
        </li>
        <li className="ressource">
          <span>{t('global', 'energy')}</span>
          <br />
          {planet
            ? `${formatNumber(Math.floor(energyRemaining))} (${formatNumber(Math.floor(energyProd))})`
            : '—'}
        </li>
        <li className="ressource">
          <span>{t('global', 'credits')}</span>
          <br />
          {formatNumber(Math.floor(credit))}
          <br />
          <small>{t('global', 'creditPay')}</small>
        </li>
        <li className="universe-name">
          <span>Oxsar Classic</span>
        </li>
        <li>
          <span>{meQ.data?.username ?? ''}</span>
        </li>
        <li>
          <select defaultValue="ru" disabled>
            <option value="ru">RU</option>
          </select>
        </li>
        <li>
          <button type="button" className="button" onClick={logout}>
            {t('global', 'btnLogout')}
          </button>
        </li>
      </ul>
    </div>
  );
}
