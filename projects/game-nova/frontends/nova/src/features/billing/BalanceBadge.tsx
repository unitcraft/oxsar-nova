import { useQuery } from '@tanstack/react-query';
import { api } from '@/api/client';

/**
 * BalanceBadge — индикатор баланса OXC в шапке игры.
 *
 * Дёргает /billing/wallet/balance (vite proxy → billing-service:9100).
 * Обновляется каждые 30s (чтобы видеть top-up после webhook'а).
 *
 * При клике ведёт на shop (#credits) — там можно купить пакет кредитов.
 *
 * План 38 Ф.7.
 */
interface Balance {
  balance: number;
  currency_code: string;
  frozen: boolean;
}

export function BalanceBadge() {
  const { data, isLoading } = useQuery({
    queryKey: ['billing', 'balance'],
    queryFn: () => api.get<Balance>('/billing/wallet/balance'),
    refetchInterval: 30_000,
    staleTime: 10_000,
  });

  const balance = data?.balance ?? 0;
  const frozen = data?.frozen ?? false;

  return (
    <a
      href="#credits"
      title={frozen ? 'Кошелёк заморожен — обратитесь в поддержку' : 'Купить кредиты'}
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: 4,
        padding: '4px 10px',
        borderRadius: 4,
        background: frozen ? 'var(--ox-danger-bg, rgba(239,68,68,0.15))' : 'var(--ox-panel)',
        border: `1px solid ${frozen ? 'var(--ox-danger)' : 'var(--ox-border)'}`,
        color: frozen ? 'var(--ox-danger)' : 'var(--ox-fg)',
        fontFamily: 'var(--ox-mono)',
        fontSize: 13,
        textDecoration: 'none',
      }}
    >
      <span>💳</span>
      <span>{isLoading ? '…' : balance.toLocaleString('ru-RU')}</span>
      {frozen && <span style={{ fontSize: 10, opacity: 0.7 }}>заморожен</span>}
    </a>
  );
}
