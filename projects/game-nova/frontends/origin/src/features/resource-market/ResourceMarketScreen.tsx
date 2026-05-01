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
  exchangeCredit,
  exchangeResource,
  fetchMarketRates,
  type CreditExchangeResult,
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

  // План 72.1.21: credit-exchange (legacy `Credit_ex`).
  const [creditResource, setCreditResource] = useState<ResourceKind>('metal');
  const [creditAmount, setCreditAmount] = useState<string>('100');
  const [creditLast, setCreditLast] = useState<CreditExchangeResult | null>(null);

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
      exchangeCredit({
        planetId: planet?.id ?? '',
        resource: creditResource,
        amount: parseAmount(creditAmount),
      }),
    onSuccess: (res) => {
      setCreditLast(res);
      setErrMsg(null);
      void qc.invalidateQueries({ queryKey: QK.planets() });
      // /me invalidate чтобы шапка с credit обновилась.
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

  const creditAmt = parseAmount(creditAmount);

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

    {/* План 72.1.21: legacy `Credit_ex` — покупка ресурса за кредиты. */}
    <form
      method="post"
      onSubmit={(ev) => {
        ev.preventDefault();
        if (!creditExchange.isPending && creditAmt > 0) creditExchange.mutate();
      }}
    >
      <table className="ntable">
        <thead>
          <tr>
            <th colSpan={3}>
              {t('market', 'tabCredit') || 'Купить за кредиты'}
            </th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>
              <label htmlFor="cr-resource">
                {t('market', 'labelTo')}
              </label>
              <br />
              <select
                id="cr-resource"
                value={creditResource}
                onChange={(e) =>
                  setCreditResource(e.target.value as ResourceKind)
                }
              >
                {RESOURCES.map((r) => (
                  <option key={r} value={r}>
                    {t('market', `resLabel${capitalize(r)}`)}
                  </option>
                ))}
              </select>
            </td>
            <td>
              <label htmlFor="cr-amount">
                {t('market', 'creditAmountLabel') || 'Кредитов'}
              </label>
              <br />
              <input
                type="text"
                id="cr-amount"
                name="credit_amount"
                value={creditAmount}
                onChange={(e) => setCreditAmount(e.target.value)}
              />
            </td>
            <td className="center">
              <input
                type="submit"
                className="button"
                value={
                  creditExchange.isPending
                    ? '…'
                    : t('market', 'creditExchangeBtn') || 'Купить'
                }
                disabled={creditAmt <= 0 || creditExchange.isPending}
              />
            </td>
          </tr>
          {creditLast && (
            <tr>
              <td colSpan={3} className="center">
                {t('market', 'creditLastExchange', {
                  credits: formatNumber(-creditLast.credit_delta),
                  amount: formatNumber(creditLast.resource_delta),
                  res: t('market', `resLabel${capitalize(creditLast.resource)}`),
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
