// Top header origin-фронта.
//
// Pixel-perfect клон legacy `topHeader` (layout.tpl + NS.class.php):
//   - <table class="top_header_res"> (НЕ ul)
//   - Колонка 1: имя планеты + координаты + "Склад:"
//   - Колонки 2-5: иконка + label + значение (class="false" при переполнении) + cap
//   - Колонка 6: кредиты + "Пополнить"
//
// Цветовая логика (NS.class.php):
//   metal >= metal_cap  → class "false" (красный)
//   иначе               → class "" (без цвета)

import { Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { useAuthStore } from '@/stores/auth';
import { useResolvedPlanet } from '@/features/common/useResolvedPlanet';
import { fetchMe } from '@/api/me';
import { QK } from '@/api/query-keys';
import { formatNumber, formatCoords } from '@/lib/format';
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

function atCap(value: number, cap: number): boolean {
  return cap > 0 && value >= cap;
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

  const metalVal = planet ? Math.floor(planet.metal) : 0;
  const siliconVal = planet ? Math.floor(planet.silicon) : 0;
  const hydrogenVal = planet ? Math.floor(planet.hydrogen) : 0;

  const metalFull = planet ? atCap(metalVal, planet.metal_cap) : false;
  const siliconFull = planet ? atCap(siliconVal, planet.silicon_cap) : false;
  const hydrogenFull = planet ? atCap(hydrogenVal, planet.hydrogen_cap) : false;

  return (
    <div id="topHeader" style={{ textAlign: 'center' }}>
      <table width="auto" cellPadding={0} cellSpacing={0} className="top_header_res">
        <tbody>
          <tr>
            {/* Колонка 1: планета + координаты + "Склад:" */}
            <td className="header-planet-name">
              {planet ? (
                <>
                  <b><Link to="/planet-options">{planet.name}</Link></b>{' '}
                  <Link to={`/galaxy/${planet.galaxy}/${planet.system}`}>
                    [{formatCoords(planet.galaxy, planet.system, planet.position)}]
                  </Link>
                  <br />
                  <br />
                  {t('overview', 'storageLabel')}
                </>
              ) : '—'}
            </td>

            {/* Металл */}
            <td className="header-resource">
              <img src="/assets/origin/images/met.gif" title={t('overview', 'resMetalLabel')} alt="" />
              <br />
              <span className="ressource">{t('overview', 'resMetalLabel')}</span>
              <br />
              <span id="header_layout_metal" className={metalFull ? 'false' : ''}>
                {planet ? formatNumber(metalVal) : '—'}
              </span>
              <br />
              <span className={metalFull ? 'false' : ''}>
                {planet ? fmtCap(planet.metal_cap) : ''}
              </span>
            </td>

            {/* Кремний */}
            <td className="header-resource">
              <img src="/assets/origin/images/silicon.gif" title={t('overview', 'resSiliconLabel')} alt="" />
              <br />
              <span className="ressource">{t('overview', 'resSiliconLabel')}</span>
              <br />
              <span id="header_layout_silicon" className={siliconFull ? 'false' : ''}>
                {planet ? formatNumber(siliconVal) : '—'}
              </span>
              <br />
              <span className={siliconFull ? 'false' : ''}>
                {planet ? fmtCap(planet.silicon_cap) : ''}
              </span>
            </td>

            {/* Водород */}
            <td className="header-resource">
              <img src="/assets/origin/images/hydrogen.gif" title={t('overview', 'resHydrogenLabel')} alt="" />
              <br />
              <span className="ressource">{t('overview', 'resHydrogenLabel')}</span>
              <br />
              <span id="header_layout_hydrogen" className={hydrogenFull ? 'false' : ''}>
                {planet ? formatNumber(hydrogenVal) : '—'}
              </span>
              <br />
              <span className={hydrogenFull ? 'false' : ''}>
                {planet ? fmtCap(planet.hydrogen_cap) : ''}
              </span>
            </td>

            {/* Энергия */}
            <td className="header-resource">
              <img src="/assets/origin/images/energy.gif" title={t('global', 'energy')} alt="" />
              <br />
              <span className="ressource">{t('global', 'energy')}</span>
              <br />
              <span id="header_layout_energy" className="">
                {planet
                  ? `${formatNumber(Math.floor(energyRemaining))} (${formatNumber(Math.floor(energyProd))})`
                  : '—'}
              </span>
              <br />
              {planet ? formatNumber(Math.floor(energyProd)) : ''}
            </td>

            {/* Кредиты */}
            <td className="header-resource">
              <img src="/assets/origin/images/credit.gif" alt="" />
              <br />
              <span className="ressource">{t('global', 'credits')}</span>
              <br />
              <span id="header_layout_credit">
                {formatNumber(Math.floor(credit))}
              </span>
              <br />
              <Link to="/payment">{t('global', 'creditPay')}</Link>
            </td>

            {/* Username + logout */}
            <td className="header-resource" style={{ whiteSpace: 'nowrap' }}>
              <span>{meQ.data?.username ?? ''}</span>
              <br />
              <button type="button" className="button" onClick={logout} style={{ marginTop: 2 }}>
                {t('global', 'btnLogout')}
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  );
}
