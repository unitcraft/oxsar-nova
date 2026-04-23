import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';
import { useToast } from '@/ui/Toast';

interface CreditPackage {
  key: string;
  label: string;
  credits: number;
  bonus_credits: number;
  total_credits: number;
  price_rub: number;
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

const STATUS_LABEL: Record<string, string> = {
  paid:     '✅ оплачен',
  pending:  '⏳ ожидает',
  failed:   '❌ ошибка',
  refunded: '↩️ возврат',
};

function fmtDate(iso: string) {
  return new Date(iso).toLocaleDateString('ru-RU', { day: '2-digit', month: '2-digit', year: 'numeric' });
}

export function CreditsScreen() {
  const qc = useQueryClient();
  const { show: showToast } = useToast();

  const me = useQuery({
    queryKey: ['me'],
    queryFn: () => api.get<{ credit: number }>('/api/me'),
  });

  const packages = useQuery({
    queryKey: ['payment', 'packages'],
    queryFn: () => api.get<CreditPackage[]>('/api/payment/packages'),
    staleTime: Infinity,
  });

  const history = useQuery({
    queryKey: ['payment', 'history'],
    queryFn: () => api.get<Purchase[]>('/api/payment/history'),
  });

  const buyMutation = useMutation({
    mutationFn: (packageKey: string) =>
      api.post<{ order_id: string; pay_url: string }>('/api/payment/order', { package_key: packageKey }),
    onSuccess: (data) => {
      window.open(data.pay_url, '_blank', 'noopener,noreferrer');
      void qc.invalidateQueries({ queryKey: ['payment', 'history'] });
    },
    onError: () => {
      showToast('danger', 'Не удалось создать заказ. Попробуйте позже.');
    },
  });

  const balance = me.data?.credit ?? 0;

  return (
    <div className="screen">
      <h2>Пополнение кредитов</h2>

      <p className="credits-balance">
        Баланс: <strong>💳 {balance.toLocaleString('ru-RU')} кр</strong>
      </p>

      {packages.isLoading && <p>Загрузка пакетов…</p>}
      {packages.isError && <p className="error">Ошибка загрузки пакетов</p>}

      {packages.data && (
        <div className="credit-packages">
          {packages.data.map((pkg) => (
            <div key={pkg.key} className="credit-package-card">
              <div className="credit-package-label">{pkg.label}</div>
              <div className="credit-package-credits">
                {pkg.total_credits.toLocaleString('ru-RU')} кр
                {pkg.bonus_credits > 0 && (
                  <span className="credit-package-bonus"> (+{pkg.bonus_credits.toLocaleString('ru-RU')} бонус)</span>
                )}
              </div>
              <div className="credit-package-price">{pkg.price_rub.toLocaleString('ru-RU')} ₽</div>
              <button
                className="btn-primary"
                disabled={buyMutation.isPending}
                onClick={() => buyMutation.mutate(pkg.key)}
              >
                Купить
              </button>
            </div>
          ))}
        </div>
      )}

      <h3>История покупок</h3>

      {history.isLoading && <p>Загрузка…</p>}
      {history.isError && <p className="error">Ошибка загрузки истории</p>}
      {history.data && history.data.length === 0 && <p className="muted">Покупок пока нет.</p>}

      {history.data && history.data.length > 0 && (
        <table className="data-table">
          <thead>
            <tr>
              <th>Дата</th>
              <th>Пакет</th>
              <th>Кредиты</th>
              <th>Сумма</th>
              <th>Статус</th>
            </tr>
          </thead>
          <tbody>
            {history.data.map((p) => (
              <tr key={p.id}>
                <td>{fmtDate(p.created_at)}</td>
                <td>{p.package_label}</td>
                <td>+{p.credits.toLocaleString('ru-RU')} кр</td>
                <td>{p.price_rub.toLocaleString('ru-RU')} ₽</td>
                <td>{STATUS_LABEL[p.status] ?? p.status}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}
