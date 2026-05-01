// S-020 Resource market (план 72 Ф.3 Spring 2 ч.2).
//
// Pixel-perfect зеркало legacy `templates/standard/market.tpl` —
// «Галактический обменник» с курсом ресурсов и комиссией.
//
// Endpoint (openapi.yaml):
//   GET  /api/market/rates                       → MarketRates
//                                                   (metal/silicon/hydrogen + user_rate)
//   POST /api/planets/{id}/market/exchange       → ExchangeResult
//                                                   (Idempotency-Key R9)
//
// UX-flow: игрок выбирает direction (from/to из {metal,silicon,hydrogen}) +
// amount → видит ожидаемый to_amount через клиентскую формулу
// rate_from/rate_to. После POST показываем фактический результат.
//
// Замечание про CREDIT: legacy показывает 4-й столбец «кредиты», но в
// nova-API exchange-endpoint поддерживает только {metal,silicon,hydrogen}
// (см. ResourceKind). Покупка ресурсов за кредиты — отдельный механизм
// (биллинг плана 36 + market.fleet_lots), не входит в S-020.

import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  exchangeCreditMulti,
  exchangeResource,
  fetchMarketRates,
  multiCreditCost,
  type CreditExchangeMultiResult,
} from '@/api/market';
import { QK } from '@/api/query-keys';
import type { ApiError } from '@/api/client';
import type { ExchangeResult, ResourceKind } from '@/api/types';
import { useResolvedPlanet } from '@/features/common/useResolvedPlanet';
import { formatNumber } from '@/lib/format';
import { useTranslation } from '@/i18n/i18n';

const RESOURCES: ResourceKind[] = ['metal', 'silicon', 'hydrogen'];

export function ResourceMarketScreen() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  const { planet, isLoading } = useResolvedPlanet();

  const [from, setFrom] = useState<ResourceKind>('metal');
  const [to, setTo] = useState<ResourceKind>('silicon');
  const [amount, setAmount] = useState<string>('1000');
  const [last, setLast] = useState<ExchangeResult | null>(null);
  const [errMsg, setErrMsg] = useState<string | null>(null);

  // План 72.1.28: multi-resource credit-exchange (legacy
  // `Credit_ex(metal, silicon, hydrogen)`). Все три за раз.
  const [creditMetal, setCreditMetal] = useState<string>('0');
  const [creditSilicon, setCreditSilicon] = useState<string>('0');
  const [creditHydrogen, setCreditHydrogen] = useState<string>('0');
  const [creditLast, setCreditLast] = useState<CreditExchangeMultiResult | null>(null);

  const ratesQ = useQuery({
    queryKey: QK.marketRates(),
    queryFn: fetchMarketRates,
    staleTime: 60_000,
  });

  const exchange = useMutation({
    mutationFn: () =>
      exchangeResource({
        planetId: planet?.id ?? '',
        from,
        to,
        amount: parseAmount(amount),
      }),
    onSuccess: (res) => {
      setLast(res);
      setErrMsg(null);
      void qc.invalidateQueries({ queryKey: QK.planets() });
    },
    onError: (e) => setErrMsg((e as ApiError).message),
  });

  const creditExchange = useMutation({
    mutationFn: () =>
      exchangeCreditMulti({
        planetId: planet?.id ?? '',
        metal: parseAmount(creditMetal),
        silicon: parseAmount(creditSilicon),
        hydrogen: parseAmount(creditHydrogen),
      }),
    onSuccess: (res) => {
      setCreditLast(res);
      setErrMsg(null);
      void qc.invalidateQueries({ queryKey: QK.planets() });
      void qc.invalidateQueries({ queryKey: ['me'] });
    },
    onError: (e) => setErrMsg((e as ApiError).message),
  });

  if (isLoading) return <div className="idiv">…</div>;
  if (!planet) return <div className="idiv">—</div>;

  const rates = ratesQ.data;
  const rateFromTo = rates ? rates[from] / rates[to] : 1;
  const userRatePct = rates ? Math.round(rates.user_rate * 100) : 100;
  const commissionPct = Math.max(0, 100 - userRatePct);

  const amt = parseAmount(amount);
  const expected = amt > 0 ? Math.floor(amt * rateFromTo) : 0;

  const cMetal = parseAmount(creditMetal);
  const cSilicon = parseAmount(creditSilicon);
  const cHydrogen = parseAmount(creditHydrogen);
  const totalCreditCost = multiCreditCost(cMetal, cSilicon, cHydrogen);
  const hasCreditAmount = cMetal > 0 || cSilicon > 0 || cHydrogen > 0;

  return (
    <>
    <form
      method="post"
      onSubmit={(ev) => {
        ev.preventDefault();
        if (!exchange.isPending && from !== to && amt > 0) exchange.mutate();
      }}
    >
      <table className="ntable">
        <thead>
          <tr>
            <th colSpan={2}>{t('market', 'tabExchange')}</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td colSpan={2} className="center">
              {t('market', 'globalRate', {
                siRatio: rates ? rates.metal / rates.silicon : 0,
                hRatio: rates ? rates.metal / rates.hydrogen : 0,
              })}
              <br />
              {t('market', 'yourRate')} {userRatePct}% ·{' '}
              {commissionPct > 5 ? (
                <span className="false">
                  <b>{commissionPct}%</b>
                </span>
              ) : (
                <span className="true">
                  <b>{commissionPct}%</b>
                </span>
              )}
            </td>
          </tr>
        </tbody>
      </table>

      <table className="ntable">
        <thead>
          <tr>
            <th colSpan={3}>{t('market', 'exchangeBtn')}</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>
              <label htmlFor="from">{t('market', 'labelFrom')}</label>
              <br />
              <select
                id="from"
                value={from}
                onChange={(e) => setFrom(e.target.value as ResourceKind)}
              >
                {RESOURCES.map((r) => (
                  <option key={r} value={r}>
                    {t('market', `resLabel${capitalize(r)}`)}
                  </option>
                ))}
              </select>
            </td>
            <td>
              <label htmlFor="amount">{t('market', 'labelAmount')}</label>
              <br />
              <input
                type="text"
                id="amount"
                name="amount"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
              />
            </td>
            <td>
              <label htmlFor="to">{t('market', 'labelTo')}</label>
              <br />
              <select
                id="to"
                value={to}
                onChange={(e) => setTo(e.target.value as ResourceKind)}
              >
                {RESOURCES.map((r) => (
                  <option key={r} value={r}>
                    {t('market', `resLabel${capitalize(r)}`)}
                  </option>
                ))}
              </select>
            </td>
          </tr>
          <tr>
            <td colSpan={3} className="center">
              {t('market', 'youWillGet')} <b>{formatNumber(expected)}</b>
            </td>
          </tr>
          {errMsg && (
            <tr>
              <td colSpan={3} className="center">
                <span className="false">{errMsg}</span>
              </td>
            </tr>
          )}
          <tr>
            <td colSpan={3} className="center">
              <input
                type="submit"
                className="button"
                value={
                  exchange.isPending ? '…' : t('market', 'exchangeBtn')
                }
                disabled={
                  from === to || amt <= 0 || exchange.isPending || !rates
                }
              />
            </td>
          </tr>
          {last && (
            <tr>
              <td colSpan={3} className="center">
                {t('market', 'lastExchange', {
                  fa: formatNumber(last.from_amount),
                  fr: t('market', `resLabel${capitalize(last.from)}`),
                  ta: formatNumber(last.to_amount),
                  tr: t('market', `resLabel${capitalize(last.to)}`),
                  rate: last.rate.toFixed(3),
                })}
              </td>
            </tr>
          )}
        </tbody>
      </table>
    </form>

    {/* План 72.1.28: legacy `Credit_ex(metal, silicon, hydrogen)` —
        multi-resource покупка за кредиты. Все три за раз; commission=0
        и storage=UNLIMIT (legacy override в Credit_ex). */}
    <form
      method="post"
      onSubmit={(ev) => {
        ev.preventDefault();
        if (!creditExchange.isPending && hasCreditAmount) creditExchange.mutate();
      }}
    >
      <table className="ntable">
        <thead>
          <tr>
            <th colSpan={4}>
              {t('market', 'tabCredit') || 'Купить за кредиты'}
            </th>
          </tr>
          <tr>
            <th>{t('market', 'resLabelMetal')}</th>
            <th>{t('market', 'resLabelSilicon')}</th>
            <th>{t('market', 'resLabelHydrogen')}</th>
            <th className="center">{t('market', 'creditExchangeBtn') || 'Купить'}</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>
              <input
                type="text"
                id="cr-metal"
                name="credit_metal"
                value={creditMetal}
                onChange={(e) => setCreditMetal(e.target.value)}
                style={{ width: '100%' }}
              />
              {cMetal > 0 && (
                <div style={{ fontSize: 'smaller' }}>
                  {Math.ceil(cMetal / 100)} кр
                </div>
              )}
            </td>
            <td>
              <input
                type="text"
                id="cr-silicon"
                name="credit_silicon"
                value={creditSilicon}
                onChange={(e) => setCreditSilicon(e.target.value)}
                style={{ width: '100%' }}
              />
              {cSilicon > 0 && (
                <div style={{ fontSize: 'smaller' }}>
                  {Math.ceil(cSilicon / 50)} кр
                </div>
              )}
            </td>
            <td>
              <input
                type="text"
                id="cr-hydrogen"
                name="credit_hydrogen"
                value={creditHydrogen}
                onChange={(e) => setCreditHydrogen(e.target.value)}
                style={{ width: '100%' }}
              />
              {cHydrogen > 0 && (
                <div style={{ fontSize: 'smaller' }}>
                  {Math.ceil(cHydrogen / 25)} кр
                </div>
              )}
            </td>
            <td className="center">
              <div>
                {t('market', 'creditTotalLabel') || 'Итого:'}{' '}
                <b>{formatNumber(totalCreditCost)}</b> кр
              </div>
              <input
                type="submit"
                className="button"
                style={{ marginTop: 4 }}
                value={
                  creditExchange.isPending
                    ? '…'
                    : t('market', 'creditExchangeBtn') || 'Купить'
                }
                disabled={!hasCreditAmount || creditExchange.isPending}
              />
            </td>
          </tr>
          {creditLast && (
            <tr>
              <td colSpan={4} className="center">
                {t('market', 'creditLastExchangeMulti', {
                  credits: formatNumber(creditLast.credits),
                  metal: formatNumber(creditLast.metal),
                  silicon: formatNumber(creditLast.silicon),
                  hydrogen: formatNumber(creditLast.hydrogen),
                })}
              </td>
            </tr>
          )}
        </tbody>
      </table>
    </form>
    </>
  );
}

function capitalize(s: string): string {
  return s.charAt(0).toUpperCase() + s.slice(1);
}

function parseAmount(s: string): number {
  const n = parseInt(s.replace(/\D/g, ''), 10);
  return Number.isFinite(n) && n > 0 ? n : 0;
}
