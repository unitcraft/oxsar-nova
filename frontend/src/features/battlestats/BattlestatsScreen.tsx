import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '@/api/client';
import { useTranslation } from '@/i18n/i18n';

interface Battle {
  id: string;
  at: string;
  winner: 'attackers' | 'defenders' | 'draw';
  rounds: number;
  role: 'attacker' | 'defender';
  opponent: string;
  opponent_id?: string | null;
  planet_name?: string | null;
  debris_metal: number;
  debris_silicon: number;
  loot_metal: number;
  loot_silicon: number;
  loot_hydrogen: number;
}

interface Stats {
  battles: Battle[];
  total: number;
  wins: number;
  losses: number;
  draws: number;
}

type Role = 'any' | 'attacker' | 'defender';
type Result = 'any' | 'win' | 'loss' | 'draw';

export function BattlestatsScreen() {
  const { t } = useTranslation('battlestatsUi');
  const [role, setRole] = useState<Role>('any');
  const [result, setResult] = useState<Result>('any');
  const [from, setFrom] = useState('');
  const [to, setTo] = useState('');

  const qs = new URLSearchParams();
  if (role !== 'any') qs.set('role', role);
  if (result !== 'any') qs.set('result', result);
  if (from) qs.set('from', from);
  if (to) qs.set('to', to);

  const q = useQuery({
    queryKey: ['battlestats', role, result, from, to],
    queryFn: () => api.get<Stats>(`/api/battlestats${qs.toString() ? '?' + qs.toString() : ''}`),
    refetchInterval: 60000,
  });

  const list = q.data?.battles ?? [];

  return (
    <div style={{ maxWidth: 1100, margin: '0 auto', padding: 24, display: 'flex', flexDirection: 'column', gap: 16 }}>
      <h2 style={{ margin: 0, fontSize: 20, fontWeight: 700 }}>{t('title')}</h2>

      <section style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(120px, 1fr))', gap: 8 }}>
        <StatCell label={t('statTotal')} value={q.data?.total ?? 0} color="var(--ox-accent)" />
        <StatCell label={t('statWins')} value={q.data?.wins ?? 0} color="var(--ox-success)" />
        <StatCell label={t('statLosses')} value={q.data?.losses ?? 0} color="var(--ox-danger)" />
        <StatCell label={t('statDraws')} value={q.data?.draws ?? 0} color="var(--ox-fg-muted)" />
      </section>

      <section className="ox-panel" style={{ padding: 14, display: 'flex', gap: 12, flexWrap: 'wrap', alignItems: 'flex-end' }}>
        <div>
          <label style={{ fontSize: 13, color: 'var(--ox-fg-muted)', display: 'block' }}>{t('filterRole')}</label>
          <select value={role} onChange={(e) => setRole(e.target.value as Role)}>
            <option value="any">{t('filterRoleAny')}</option>
            <option value="attacker">{t('filterRoleAttacker')}</option>
            <option value="defender">{t('filterRoleDefender')}</option>
          </select>
        </div>
        <div>
          <label style={{ fontSize: 13, color: 'var(--ox-fg-muted)', display: 'block' }}>{t('filterResult')}</label>
          <select value={result} onChange={(e) => setResult(e.target.value as Result)}>
            <option value="any">{t('filterResultAny')}</option>
            <option value="win">{t('filterResultWin')}</option>
            <option value="loss">{t('filterResultLoss')}</option>
            <option value="draw">{t('filterResultDraw')}</option>
          </select>
        </div>
        <div>
          <label style={{ fontSize: 13, color: 'var(--ox-fg-muted)', display: 'block' }}>{t('filterFrom')}</label>
          <input type="date" value={from} onChange={(e) => setFrom(e.target.value)} />
        </div>
        <div>
          <label style={{ fontSize: 13, color: 'var(--ox-fg-muted)', display: 'block' }}>{t('filterTo')}</label>
          <input type="date" value={to} onChange={(e) => setTo(e.target.value)} />
        </div>
      </section>

      <div className="ox-panel" style={{ overflow: 'hidden' }}>
        {q.isLoading && <div style={{ padding: 20 }}><div className="ox-skeleton" style={{ height: 200 }} /></div>}

        {!q.isLoading && list.length === 0 && (
          <div style={{ padding: 24, textAlign: 'center', color: 'var(--ox-fg-muted)' }}>
            {t('empty')}
          </div>
        )}

        {list.length > 0 && (
          <div className="ox-table-responsive">
            <table className="ox-table" style={{ margin: 0 }}>
              <thead>
                <tr>
                  <th>{t('colDate')}</th>
                  <th>{t('colRole')}</th>
                  <th>{t('colOpponent')}</th>
                  <th>{t('colPlanet')}</th>
                  <th>{t('colRounds')}</th>
                  <th>{t('colResult')}</th>
                  <th>{t('colLoot')}</th>
                  <th>{t('colDebris')}</th>
                </tr>
              </thead>
              <tbody>
                {list.map((b) => {
                  const isWin = (b.role === 'attacker' && b.winner === 'attackers') || (b.role === 'defender' && b.winner === 'defenders');
                  const isDraw = b.winner === 'draw';
                  const color = isWin ? 'var(--ox-success)' : isDraw ? 'var(--ox-fg-muted)' : 'var(--ox-danger)';
                  const label = isWin ? t('resultWin') : isDraw ? t('resultDraw') : t('resultLoss');
                  const loot = b.loot_metal + b.loot_silicon + b.loot_hydrogen;
                  const debris = b.debris_metal + b.debris_silicon;
                  return (
                    <tr key={b.id}>
                      <td style={{ fontSize: 13, fontFamily: 'var(--ox-mono)', color: 'var(--ox-fg-muted)' }}>
                        {new Date(b.at).toLocaleString('ru-RU', { dateStyle: 'short', timeStyle: 'short' })}
                      </td>
                      <td>{b.role === 'attacker' ? t('roleAttack') : t('roleDefend')}</td>
                      <td style={{ fontWeight: 600 }}>{b.opponent || '—'}</td>
                      <td style={{ fontSize: 14, color: 'var(--ox-fg-muted)' }}>{b.planet_name ?? '—'}</td>
                      <td className="num">{b.rounds}</td>
                      <td style={{ color, fontWeight: 600 }}>{label}</td>
                      <td className="num" style={{ fontSize: 14 }}>
                        {loot > 0 ? Math.round(loot).toLocaleString('ru-RU') : '—'}
                      </td>
                      <td className="num" style={{ fontSize: 14, color: 'var(--ox-warning)' }}>
                        {debris > 0 ? Math.round(debris).toLocaleString('ru-RU') : '—'}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}

function StatCell({ label, value, color }: { label: string; value: number; color: string }) {
  return (
    <div className="ox-panel" style={{ padding: 12 }}>
      <div style={{ fontSize: 13, color: 'var(--ox-fg-muted)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>{label}</div>
      <div style={{ fontSize: 22, fontWeight: 700, fontFamily: 'var(--ox-mono)', color }}>{value}</div>
    </div>
  );
}
