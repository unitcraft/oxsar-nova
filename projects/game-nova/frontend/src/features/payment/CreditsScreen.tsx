import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';
import { useToast } from '@/ui/Toast';
import { useTranslation } from '@/i18n/i18n';

/**
 * CreditsScreen — магазин кредитов и история транзакций.
 *
 * План 38 Ф.7: переписан с /api/payment/* (game-nova) на /billing/* (billing-service).
 * Vite проксирует /billing/* → billing-service:9100.
 *
 * Endpoints:
 *   GET  /billing/packages         — каталог (публичный).
 *   POST /billing/orders           — создать заказ → pay_url.
 *   GET  /billing/wallet/balance   — баланс OXC.
 *   GET  /billing/wallet/history   — история транзакций.
 */

interface BillingPackage {
  id: string;
  title: string;
  amount_kop: number;          // 50000 = 500 RUB
  credits: number;             // base credits в OXC
  bonus?: number;              // бонус сверх base
  is_best?: boolean;
}

interface PackagesResponse {
  packages: BillingPackage[];
}

interface Balance {
  balance: number;
  currency_code: string;
  frozen: boolean;
}

interface Transaction {
  id: string;
  delta: number;               // +N для пополнений, -N для трат
  balance_after: number;
  reason: string;              // 'top_up' | 'feedback_vote' | ...
  ref_id?: string;
  from_account: string;
  to_account: string;
  created_at: string;
}

interface HistoryResponse {
  transactions: Transaction[];
}

const REASON_LABEL: Record<string, string> = {
  top_up:        'Пополнение',
  feedback_vote: 'Голос за предложение',
  shop_purchase: 'Покупка',
  refund:        'Возврат',
  admin_grant:   'Начисление администратора',
};

function fmtDate(iso: string) {
  return new Date(iso).toLocaleDateString('ru-RU', { day: '2-digit', month: '2-digit', year: 'numeric' });
}

function totalCredits(p: BillingPackage): number {
  return p.credits + (p.bonus ?? 0);
}

export function CreditsScreen() {
  const { t } = useTranslation('credits');
  const qc = useQueryClient();
  const { show: showToast } = useToast();

  const balance = useQuery({
    queryKey: ['billing', 'balance'],
    queryFn: () => api.get<Balance>('/billing/wallet/balance'),
  });

  const packages = useQuery({
    queryKey: ['billing', 'packages'],
    queryFn: () => api.get<PackagesResponse>('/billing/packages'),
    staleTime: Infinity,
  });

  const history = useQuery({
    queryKey: ['billing', 'history'],
    queryFn: () => api.get<HistoryResponse>('/billing/wallet/history?limit=50'),
  });

  const buyMutation = useMutation({
    mutationFn: (packageId: string) =>
      api.post<{ order: { id: string; status: string }; pay_url: string }>(
        '/billing/orders',
        { package_id: packageId },
      ),
    onSuccess: (data) => {
      // pay_url ведёт на mock pay-handler в dev (или внешний шлюз в prod).
      // Mock в Ф.4 не имеет pay-page — webhook вызывается напрямую через
      // curl/test-helper, а fronend просто открывает URL (попадёт в 404
      // от billing, что в dev ОК — реальная страница будет с Robokassa).
      window.location.href = data.pay_url;
    },
    onError: (err) => {
      showToast('danger', err instanceof Error ? err.message : 'Ошибка покупки');
    },
  });

  // refresh после возврата с ?payment=success в Header (см. App.tsx).
  void qc;

  const balanceVal = balance.data?.balance ?? 0;

  return (
    <div className="screen">
      <h2>Магазин кредитов</h2>

      <p className="credits-balance">
        Баланс: <strong>💳 {balanceVal.toLocaleString('ru-RU')} OXC</strong>
        {balance.data?.frozen && (
          <span className="ox-error" style={{ marginLeft: 12 }}>⚠ кошелёк заморожен</span>
        )}
      </p>

      {packages.isLoading && <p>Загрузка пакетов…</p>}
      {packages.isError && <p className="ox-error">Не удалось загрузить пакеты</p>}

      {packages.data && (
        <div className="credit-packages">
          {packages.data.packages.map((pkg) => (
            <div
              key={pkg.id}
              className={`credit-package-card ${pkg.is_best ? 'credit-package-best' : ''}`}
            >
              {pkg.is_best && (
                <div style={{
                  position: 'absolute',
                  top: -10,
                  right: 12,
                  padding: '2px 10px',
                  fontSize: 11,
                  fontWeight: 700,
                  background: 'var(--ox-accent)',
                  color: 'var(--ox-bg)',
                  borderRadius: 4,
                  textTransform: 'uppercase',
                }}>
                  Выгодно
                </div>
              )}
              <div className="credit-package-label">{pkg.title}</div>
              <div className="credit-package-credits">
                {totalCredits(pkg).toLocaleString('ru-RU')} OXC
                {pkg.bonus && pkg.bonus > 0 && (
                  <span className="credit-package-bonus">
                    {' '}(+{pkg.bonus.toLocaleString('ru-RU')} бонус)
                  </span>
                )}
              </div>
              <div className="credit-package-price">
                {(pkg.amount_kop / 100).toLocaleString('ru-RU')} ₽
              </div>
              <button
                className="btn-primary"
                disabled={buyMutation.isPending || balance.data?.frozen}
                onClick={() => buyMutation.mutate(pkg.id)}
              >
                Купить
              </button>
            </div>
          ))}
        </div>
      )}

      <h3 style={{ marginTop: 24 }}>История транзакций</h3>

      {history.isLoading && <p>Загрузка истории…</p>}
      {history.isError && <p className="ox-error">Не удалось загрузить историю</p>}
      {history.data && history.data.transactions.length === 0 && (
        <p className="muted">История пуста</p>
      )}

      {history.data && history.data.transactions.length > 0 && (
        <table className="data-table">
          <thead>
            <tr>
              <th>Дата</th>
              <th>Операция</th>
              <th style={{ textAlign: 'right' }}>Сумма</th>
              <th style={{ textAlign: 'right' }}>Баланс после</th>
            </tr>
          </thead>
          <tbody>
            {history.data.transactions.map((tx) => (
              <tr key={tx.id}>
                <td>{fmtDate(tx.created_at)}</td>
                <td>{REASON_LABEL[tx.reason] ?? tx.reason}</td>
                <td style={{
                  textAlign: 'right',
                  fontFamily: 'var(--ox-mono)',
                  color: tx.delta > 0 ? 'var(--ox-success, #10b981)' : 'var(--ox-danger, #ef4444)',
                }}>
                  {tx.delta > 0 ? '+' : ''}{tx.delta.toLocaleString('ru-RU')}
                </td>
                <td style={{
                  textAlign: 'right',
                  fontFamily: 'var(--ox-mono)',
                  color: 'var(--ox-fg-dim)',
                }}>
                  {tx.balance_after.toLocaleString('ru-RU')}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}

      {/* TS-suppress: t используется для legacy i18n-ключей, сейчас всё хардкод */}
      <span style={{ display: 'none' }}>{t('balanceLabel')}</span>
    </div>
  );
}
