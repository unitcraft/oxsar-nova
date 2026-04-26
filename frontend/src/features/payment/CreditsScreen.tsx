import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';
import { useToast } from '@/ui/Toast';
import { useTranslation } from '@/i18n/i18n';

interface CreditPackage {
  key: string;
  label: string;
  credits: number;
  bonus_credits: number;
  total_credits: number;
  price_rub: number;
}

interface PackagesResponse {
  packages: CreditPackage[];
  test_mode: boolean;
}

interface Purchase {
  id: string;
  package_key: string;
  package_label: string;
  credits: number;
  price_rub: number;
  status: 'pending' | 'paid' | 'failed' | 'refunded';
  created_at: string;
  paid_at?: string | null;
}

const STATUS_KEY: Record<string, string> = {
  paid:     'statusPaid',
  pending:  'statusPending',
  failed:   'statusFailed',
  refunded: 'statusCancelled',
};

function fmtDate(iso: string) {
  return new Date(iso).toLocaleDateString('ru-RU', { day: '2-digit', month: '2-digit', year: 'numeric' });
}

function packageHint(total: number, t: (key: string) => string): string | null {
  if (total >= 5000) return t('hint5000');
  if (total >= 2000) return t('hint2000');
  if (total >= 1000) return t('hint1000');
  if (total >= 500)  return t('hint500');
  if (total >= 100)  return t('hint100');
  return null;
}

export function CreditsScreen() {
  const { t } = useTranslation('creditsUi');
  const qc = useQueryClient();
  const { show: showToast } = useToast();

  const me = useQuery({
    queryKey: ['me'],
    queryFn: () => api.get<{ credit: number }>('/api/me'),
  });

  const packages = useQuery({
    queryKey: ['payment', 'packages'],
    queryFn: () => api.get<PackagesResponse>('/api/payment/packages'),
    staleTime: Infinity,
  });

  const testMode = packages.data?.test_mode ?? false;

  const history = useQuery({
    queryKey: ['payment', 'history'],
    queryFn: () => api.get<Purchase[]>('/api/payment/history'),
  });

  const buyMutation = useMutation({
    mutationFn: (packageKey: string) =>
      api.post<{ order_id: string; pay_url: string }>('/api/payment/order', { package_key: packageKey }),
    onSuccess: (data) => {
      // В mock-режиме pay_url — локальный эндпоинт, который редиректит обратно
      // с ?payment=success/fail. В prod — внешний сайт шлюза в новой вкладке.
      if (testMode) {
        window.location.href = data.pay_url;
      } else {
        window.open(data.pay_url, '_blank', 'noopener,noreferrer');
        void qc.invalidateQueries({ queryKey: ['payment', 'history'] });
      }
    },
    onError: () => {
      showToast('danger', t('sectionPackages'));
    },
  });

  const balance = me.data?.credit ?? 0;

  return (
    <div className="screen">
      <h2>{t('sectionPackages')}</h2>

      {testMode && (
        <div
          role="alert"
          style={{
            padding: '10px 14px',
            marginBottom: 12,
            borderRadius: 6,
            background: 'rgba(245,158,11,0.12)',
            border: '1px solid rgba(245,158,11,0.6)',
            color: 'var(--ox-warn, #f59e0b)',
            fontSize: 15,
            fontWeight: 600,
          }}
        >
          ⚠️ {t('testModeBanner')}
        </div>
      )}

      <p className="credits-balance">
        {t('balanceLabel')} <strong>💳 {balance.toLocaleString('ru-RU')} {t('creditsUnit')}</strong>
      </p>

      {packages.isLoading && <p>{t('historyEmpty')}</p>}
      {packages.isError && <p className="error">{t('historyEmpty')}</p>}

      {packages.data && (
        <div className="credit-packages">
          {packages.data.packages.map((pkg) => {
            const hint = packageHint(pkg.total_credits, t);
            return (
              <div key={pkg.key} className="credit-package-card">
                <div className="credit-package-label">{pkg.label}</div>
                <div className="credit-package-credits">
                  {pkg.total_credits.toLocaleString('ru-RU')} {t('creditsUnit')}
                  {pkg.bonus_credits > 0 && (
                    <span className="credit-package-bonus"> (+{pkg.bonus_credits.toLocaleString('ru-RU')} {t('bonusLabel')})</span>
                  )}
                </div>
                <div className="credit-package-price">{pkg.price_rub.toLocaleString('ru-RU')} ₽</div>
                {hint && (
                  <div style={{ fontSize: 13, color: 'var(--ox-fg-dim)', marginBottom: 8, lineHeight: 1.4 }}>
                    {hint}
                  </div>
                )}
                <button
                  className="btn-primary"
                  disabled={buyMutation.isPending}
                  onClick={() => buyMutation.mutate(pkg.key)}
                >
                  {t('buyBtn')}
                </button>
              </div>
            );
          })}
        </div>
      )}

      <details style={{ marginTop: 16, marginBottom: 16 }}>
        <summary style={{ cursor: 'pointer', fontWeight: 600, fontSize: 16, padding: '8px 0' }}>
          {t('sectionSpendHints')}
        </summary>
        <div style={{ padding: 12, background: 'var(--ox-bg-panel)', borderRadius: 6, marginTop: 6 }}>
          {(['Officer', 'Profession', 'Rename', 'Artefact'] as const).map((k) => (
            <div key={k} style={{ display: 'flex', gap: 8, padding: '4px 0', fontSize: 15 }}>
              <span style={{ flex: 1, color: 'var(--ox-fg)' }}>{t(`priceHint${k}.label`)}</span>
              <span style={{ fontFamily: 'var(--ox-mono)', color: 'var(--ox-accent)' }}>{t(`priceHint${k}.value`)}</span>
            </div>
          ))}
        </div>
      </details>

      <h3>{t('sectionHistory')}</h3>

      {history.isLoading && <p>{t('historyEmpty')}</p>}
      {history.isError && <p className="error">{t('historyEmpty')}</p>}
      {history.data && history.data.length === 0 && <p className="muted">{t('historyEmpty')}</p>}

      {history.data && history.data.length > 0 && (
        <table className="data-table">
          <thead>
            <tr>
              <th>{t('colDate')}</th>
              <th>{t('colPackage')}</th>
              <th>{t('colCredits')}</th>
              <th>{t('colPrice')}</th>
              <th>{t('colStatus')}</th>
            </tr>
          </thead>
          <tbody>
            {history.data.map((p) => (
              <tr key={p.id}>
                <td>{fmtDate(p.created_at)}</td>
                <td>{p.package_label}</td>
                <td>+{p.credits.toLocaleString('ru-RU')} {t('creditsUnit')}</td>
                <td>{p.price_rub.toLocaleString('ru-RU')} ₽</td>
                <td>{STATUS_KEY[p.status] ? t(STATUS_KEY[p.status]!) : p.status}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}
