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

import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { useResolvedPlanet } from '@/features/common/useResolvedPlanet';
import { fetchMe } from '@/api/me';
import { QK } from '@/api/query-keys';
import { formatNumber, formatCoords } from '@/lib/format';
import { useTranslation } from '@/i18n/i18n';

// TickValue — отрисовка ресурса с короткой мягкой подсветкой при
// изменении значения. Реализация: `key={value}` пересоздаёт <span>,
// CSS-анимация .res-tick-flash проигрывается заново на каждом тике.
// extraClass — опциональный модификатор (например, 'false' при cap).
function TickValue({ value, extraClass }: { value: string; extraClass?: string | undefined }) {
  return (
    <span
      key={value}
      className={`res-tick-flash${extraClass ? ' ' + extraClass : ''}`}
    >
      {value}
    </span>
  );
}

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

  // План 72.1.59: тик ресурсов 1:1 с legacy `Page.class.php:73-93`.
  // Legacy использует jquery.iterator.js (step=production/3600 в сек):
  //   - тик активен если storage > current ИЛИ production < 0
  //   - потолок при росте — storage (cap)
  //   - при потреблении — нижняя граница 0
  // Здесь делаем то же на клиенте: linear extrapolation по
  // `last_res_update` + per_sec, ограниченная cap'ом / нулём.
  const [now, setNow] = useState(() => Date.now());
  useEffect(() => {
    const id = setInterval(() => setNow(Date.now()), 1000);
    return () => clearInterval(id);
  }, []);

  function tick(
    base: number,
    perSec: number,
    cap: number,
    lastUpdateIso: string,
  ): number {
    const lastMs = Date.parse(lastUpdateIso);
    if (Number.isNaN(lastMs)) return Math.floor(base);
    const dt = Math.max(0, (now - lastMs) / 1000);
    let val = base + perSec * dt;
    if (perSec >= 0) {
      // Растущий ресурс: legacy не показывает тик когда storage <= current,
      // т.е. cap уже достигнут. Если ещё не достигнут — потолок = cap.
      if (val > cap) val = cap;
    } else {
      // Потребление: legacy ограничивает 0.
      if (val < 0) val = 0;
    }
    return Math.floor(val);
  }

  const metalVal = planet
    ? tick(planet.metal, planet.metal_per_sec, planet.metal_cap, planet.last_res_update)
    : 0;
  const siliconVal = planet
    ? tick(planet.silicon, planet.silicon_per_sec, planet.silicon_cap, planet.last_res_update)
    : 0;
  const hydrogenVal = planet
    ? tick(planet.hydrogen, planet.hydrogen_per_sec, planet.hydrogen_cap, planet.last_res_update)
    : 0;

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
                    {formatCoords(planet.galaxy, planet.system, planet.position)}
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
              {planet ? (
                <TickValue value={formatNumber(metalVal)} extraClass={metalFull ? 'false' : undefined} />
              ) : (
                <span id="header_layout_metal">—</span>
              )}
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
              {planet ? (
                <TickValue value={formatNumber(siliconVal)} extraClass={siliconFull ? 'false' : undefined} />
              ) : (
                <span id="header_layout_silicon">—</span>
              )}
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
              {planet ? (
                <TickValue value={formatNumber(hydrogenVal)} extraClass={hydrogenFull ? 'false' : undefined} />
              ) : (
                <span id="header_layout_hydrogen">—</span>
              )}
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
              {planet ? (
                <TickValue
                  value={`${formatNumber(Math.floor(energyRemaining))} (${formatNumber(Math.floor(energyProd))})`}
                  extraClass={energyRemaining < 0 ? 'false' : undefined}
                />
              ) : (
                <span id="header_layout_energy">—</span>
              )}
            </td>

            {/* Кредиты */}
            <td className="header-resource">
              <img src="/assets/origin/images/credit.gif" alt="" />
              <br />
              <span className="ressource">{t('global', 'credits')}</span>
              <br />
              <TickValue value={formatNumber(Math.floor(credit))} />
              <br />
              <Link to="/payment">{t('global', 'creditPay')}</Link>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  );
}
