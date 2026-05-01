// ExchangeOptsScreen — план 72.1.45 §9.
//
// Pixel-perfect зеркало legacy `?go=ExchangeOpts` admin-страницы:
// брокер видит и редактирует своё title (название биржи) + fee_percent
// (комиссия за продажу). Endpoint: GET/PATCH /api/exchange/opts.

import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';
import type { ApiError } from '@/api/client';
import { useTranslation } from '@/i18n/i18n';

interface BrokerSettings {
  user_id: string;
  title: string;
  fee_percent: number;
}

function fetchOpts(): Promise<BrokerSettings> {
  return api.get<BrokerSettings>('/api/exchange/opts');
}

function updateOpts(input: { title: string; fee_percent: number }): Promise<BrokerSettings> {
  return api.patch<BrokerSettings>('/api/exchange/opts', input);
}

export function ExchangeOptsScreen() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  const [title, setTitle] = useState('');
  const [feePercent, setFeePercent] = useState('5.0');
  const [errMsg, setErrMsg] = useState<string | null>(null);
  const [okMsg, setOkMsg] = useState<string | null>(null);

  const q = useQuery({ queryKey: ['exchange-opts'], queryFn: fetchOpts });
  useEffect(() => {
    if (q.data) {
      setTitle(q.data.title);
      setFeePercent(String(q.data.fee_percent));
    }
  }, [q.data]);

  const mut = useMutation({
    mutationFn: updateOpts,
    onSuccess: () => {
      setOkMsg(t('exchangeOpts', 'saved') || 'Сохранено');
      setErrMsg(null);
      void qc.invalidateQueries({ queryKey: ['exchange-opts'] });
      void qc.invalidateQueries({ queryKey: ['broker-stats'] });
    },
    onError: (e) => {
      setErrMsg((e as ApiError).message);
      setOkMsg(null);
    },
  });

  if (q.isLoading) return <div className="idiv">…</div>;

  function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    const fee = Number(feePercent);
    if (Number.isNaN(fee) || fee < 0 || fee > 50) {
      setErrMsg(t('exchangeOpts', 'feeOutOfRange') || 'Комиссия должна быть в диапазоне 0..50');
      return;
    }
    mut.mutate({ title: title.trim() || 'My exchange', fee_percent: fee });
  }

  return (
    <>
      <div className="idiv">
        <Link to="/p2p-exchange">← {t('p2pExchange', 'title') || 'Биржа'}</Link>
      </div>
      <form onSubmit={onSubmit}>
        <table className="ntable">
          <thead>
            <tr>
              <th colSpan={2}>{t('exchangeOpts', 'title') || 'Настройки биржи'}</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td>{t('exchangeOpts', 'titleLabel') || 'Название'}</td>
              <td>
                <input
                  type="text"
                  value={title}
                  onChange={(e) => setTitle(e.target.value)}
                  maxLength={100}
                  className="text"
                  style={{ width: '100%' }}
                />
              </td>
            </tr>
            <tr>
              <td>{t('exchangeOpts', 'feeLabel') || 'Комиссия (%)'}</td>
              <td>
                <input
                  type="number"
                  value={feePercent}
                  onChange={(e) => setFeePercent(e.target.value)}
                  min={0}
                  max={50}
                  step={0.1}
                  className="text"
                  style={{ width: '100%' }}
                />
                <div style={{ fontSize: 'smaller', color: '#888' }}>
                  {t('exchangeOpts', 'feeHint') || '0..50%, влияет на расчёт прибыли'}
                </div>
              </td>
            </tr>
            <tr>
              <td colSpan={2} className="center">
                <input
                  type="submit"
                  className="button"
                  value={t('exchangeOpts', 'saveBtn') || 'Сохранить'}
                  disabled={mut.isPending}
                />
              </td>
            </tr>
            {errMsg && (
              <tr>
                <td colSpan={2} className="center">
                  <span className="false">{errMsg}</span>
                </td>
              </tr>
            )}
            {okMsg && (
              <tr>
                <td colSpan={2} className="center">
                  <span className="true">{okMsg}</span>
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </form>
    </>
  );
}
