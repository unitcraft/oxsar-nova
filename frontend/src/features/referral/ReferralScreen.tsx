import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '@/api/client';

interface ReferredUser {
  user_id: string;
  username: string;
  points: number;
  reg_time: string;
}

interface ReferralData {
  invited_count: number;
  bonus_points: number;
  max_bonus_points: number;
  credit_percent: number;
  referred: ReferredUser[];
}

export function ReferralScreen() {
  const me = useQuery({
    queryKey: ['me'],
    queryFn: () => api.get<{ user_id: string }>('/api/me'),
    staleTime: 60000,
  });

  const data = useQuery({
    queryKey: ['referrals'],
    queryFn: () => api.get<ReferralData>('/api/referrals'),
  });

  const [copied, setCopied] = useState(false);
  const [shareError, setShareError] = useState('');

  const userId = me.data?.user_id ?? '';
  const url = userId ? `${window.location.origin}/?ref=${userId}` : '';

  async function handleCopy() {
    if (!url) return;
    try {
      await navigator.clipboard.writeText(url);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      setShareError('Не удалось скопировать');
    }
  }

  async function handleShare() {
    if (!url) return;
    setShareError('');
    if (typeof navigator.share === 'function') {
      try {
        await navigator.share({ title: 'oxsar — космическая стратегия', url });
      } catch {
        // cancelled — no-op
      }
    } else {
      await handleCopy();
    }
  }

  if (data.isLoading) {
    return <div style={{ padding: 24 }}><div className="ox-skeleton" style={{ height: 300, borderRadius: 8 }} /></div>;
  }

  const d = data.data;

  return (
    <div style={{ maxWidth: 760, margin: '0 auto', padding: 24, display: 'flex', flexDirection: 'column', gap: 20 }}>
      <h2 style={{ margin: 0, fontSize: 20, fontWeight: 700 }}>🎁 Реферальная программа</h2>

      <section className="ox-panel" style={{ padding: 20, display: 'flex', flexDirection: 'column', gap: 12 }}>
        <h3 style={{ margin: 0, fontSize: 14, fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.06em', color: 'var(--ox-fg-muted)' }}>Ваша ссылка</h3>
        <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
          <input
            type="text"
            className="ox-input"
            readOnly
            value={url}
            onFocus={(e) => e.currentTarget.select()}
            style={{ flex: 1, minWidth: 240, fontFamily: 'var(--ox-mono)', fontSize: 12 }}
          />
          <button type="button" className="btn" onClick={() => void handleCopy()}>
            {copied ? '✓ Скопировано' : '📋 Копировать'}
          </button>
          <button type="button" className="btn-ghost" onClick={() => void handleShare()}>
            📤 Поделиться
          </button>
        </div>
        {shareError && <span style={{ fontSize: 12, color: 'var(--ox-danger)' }}>{shareError}</span>}
        <p style={{ margin: 0, fontSize: 12, color: 'var(--ox-fg-dim)', lineHeight: 1.6 }}>
          Отправьте эту ссылку друзьям. Когда они зарегистрируются и начнут играть —
          вы получите {Math.round((d?.credit_percent ?? 0) * 100)}% от всех их покупок кредитов
          и очки за каждого активного реферала.
        </p>
      </section>

      <section style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))', gap: 12 }}>
        <StatCard label="Приглашено" value={String(d?.invited_count ?? 0)} icon="👥" />
        <StatCard
          label="Бонусные очки"
          value={Math.min(d?.bonus_points ?? 0, d?.max_bonus_points ?? 0).toLocaleString('ru-RU')}
          icon="🏆"
          hint={`макс. ${(d?.max_bonus_points ?? 0).toLocaleString('ru-RU')}`}
        />
        <StatCard
          label="Бонус от покупок"
          value={`${Math.round((d?.credit_percent ?? 0) * 100)}%`}
          icon="💳"
          hint="с каждой покупки кредитов реферала"
        />
      </section>

      <section className="ox-panel" style={{ padding: 20 }}>
        <h3 style={{ margin: '0 0 12px', fontSize: 14, fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.06em', color: 'var(--ox-fg-muted)' }}>Приглашённые игроки</h3>
        {(!d || d.referred.length === 0) ? (
          <div style={{ textAlign: 'center', padding: 24, color: 'var(--ox-fg-muted)' }}>
            Пока никого. Поделитесь ссылкой — бонусы приходят сразу после регистрации.
          </div>
        ) : (
          <div className="ox-table-responsive">
            <table className="ox-table" style={{ margin: 0 }}>
              <thead>
                <tr>
                  <th>Игрок</th>
                  <th className="num">Очки</th>
                  <th>Зарегистрирован</th>
                </tr>
              </thead>
              <tbody>
                {d.referred.map((u) => (
                  <tr key={u.user_id}>
                    <td>{u.username}</td>
                    <td className="num" style={{ color: 'var(--ox-accent)' }}>{Math.round(u.points).toLocaleString('ru-RU')}</td>
                    <td style={{ fontSize: 12, color: 'var(--ox-fg-muted)', fontFamily: 'var(--ox-mono)' }}>
                      {u.reg_time ? new Date(u.reg_time).toLocaleDateString('ru-RU') : '—'}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>
    </div>
  );
}

function StatCard({ label, value, icon, hint }: { label: string; value: string; icon: string; hint?: string }) {
  return (
    <div className="ox-panel" style={{ padding: 16, display: 'flex', flexDirection: 'column', gap: 4 }}>
      <div style={{ fontSize: 12, color: 'var(--ox-fg-muted)', textTransform: 'uppercase', letterSpacing: '0.06em' }}>
        {icon} {label}
      </div>
      <div style={{ fontSize: 22, fontWeight: 700, fontFamily: 'var(--ox-mono)', color: 'var(--ox-accent)' }}>{value}</div>
      {hint && <div style={{ fontSize: 11, color: 'var(--ox-fg-muted)' }}>{hint}</div>}
    </div>
  );
}
