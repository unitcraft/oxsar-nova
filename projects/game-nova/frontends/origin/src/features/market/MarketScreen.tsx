// S-021 Artefact market (план 72 Ф.3 Spring 2 ч.2).
//
// Pixel-perfect зеркало legacy `templates/standard/artefactmarket.tpl` —
// это **EXT_MODE legacy market** артефактов за credit (НЕ биржа из плана 68
// — та P2P-биржа лотов; здесь fixed-price offers через одиночное `buy`).
//
// Endpoint (openapi.yaml):
//   GET    /api/artefact-market/offers              → { offers: [] }
//   GET    /api/artefact-market/credit              → { credit: number }
//   POST   /api/artefact-market/offers/{id}/buy
//   DELETE /api/artefact-market/offers/{id}         (отмена своего)
//
// Замечание (P72.S2.G — см. simplifications.md):
// `ArtMarketOffer` в openapi содержит только { id, seller_id, unit_id,
// price, created_at }. Полное имя артефакта / описание / иконка — нет.
// В MVP отображаем `unit_id` как номер; имя берём из i18n-ключа
// `artefact.unit.{unit_id}` если есть, иначе показываем `Артефакт #{id}`.
// Расширение DTO — отдельный план.

import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  buyArtMarketOffer,
  cancelArtMarketOffer,
  fetchArtMarketCredit,
  fetchArtMarketOffers,
} from '@/api/market';
import type { ApiError } from '@/api/client';
import { QK } from '@/api/query-keys';
import { useAuthStore } from '@/stores/auth';
import { formatNumber } from '@/lib/format';
import { useTranslation } from '@/i18n/i18n';

export function MarketScreen() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  const userId = useAuthStore((s) => s.userId);
  const [errMsg, setErrMsg] = useState<string | null>(null);

  const offersQ = useQuery({
    queryKey: QK.artMarketOffers(),
    queryFn: fetchArtMarketOffers,
    refetchInterval: 30_000,
  });

  const creditQ = useQuery({
    queryKey: QK.artMarketCredit(),
    queryFn: fetchArtMarketCredit,
    staleTime: 30_000,
  });

  const buy = useMutation({
    mutationFn: (offerID: string) => buyArtMarketOffer(offerID),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: QK.artMarketOffers() });
      void qc.invalidateQueries({ queryKey: QK.artMarketCredit() });
    },
    onError: (e) => setErrMsg((e as ApiError).message),
  });

  const cancel = useMutation({
    mutationFn: (offerID: string) => cancelArtMarketOffer(offerID),
    onSuccess: () =>
      void qc.invalidateQueries({ queryKey: QK.artMarketOffers() }),
    onError: (e) => setErrMsg((e as ApiError).message),
  });

  const offers = offersQ.data?.offers ?? [];
  const credit = creditQ.data?.credit ?? 0;

  if (offersQ.isLoading) return <div className="idiv">…</div>;

  return (
    <table className="ntable">
      <thead>
        <tr>
          <th colSpan={3}>{t('artefacts', 'title')}</th>
        </tr>
        <tr>
          <td colSpan={3}>
            {t('officers', 'balance')} <b>{formatNumber(credit)}</b>{' '}
            {t('messages', 'folderCredits')}
          </td>
        </tr>
      </thead>
      <tbody>
        {offers.length === 0 && (
          <tr>
            <td colSpan={3} className="center">
              {t('alliance', 'nothing')}
            </td>
          </tr>
        )}
        {offers.map((offer) => {
          const mine = !!userId && userId === offer.seller_id;
          const canBuy = !mine && credit >= offer.price;
          return (
            <tr key={offer.id}>
              <td style={{ width: '1%' }}>#{offer.unit_id}</td>
              <td>
                <div style={{ width: '100%' }}>
                  <b>{t('artefacts', 'title')} #{offer.unit_id}</b>
                </div>
                <div style={{ fontSize: 'smaller', margin: 5 }}>
                  {new Date(offer.created_at).toLocaleString('ru-RU')}
                </div>
              </td>
              <td className="center" style={{ width: '10%' }}>
                <span className={canBuy ? 'true' : 'false'}>
                  {formatNumber(offer.price)}
                </span>
                <br />
                {mine ? (
                  <input
                    type="button"
                    className="button"
                    value={t('messages', 'deleteAll')}
                    disabled={cancel.isPending}
                    onClick={() => cancel.mutate(offer.id)}
                  />
                ) : (
                  <input
                    type="button"
                    className="button"
                    value={t('artefacts', 'sell')}
                    disabled={!canBuy || buy.isPending}
                    onClick={() => buy.mutate(offer.id)}
                    title={!canBuy ? t('officers', 'notEnoughCr') : undefined}
                  />
                )}
              </td>
            </tr>
          );
        })}
        {errMsg && (
          <tr>
            <td colSpan={3} className="center">
              <span className="false">{errMsg}</span>
            </td>
          </tr>
        )}
      </tbody>
    </table>
  );
}
