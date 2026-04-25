import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '../../api/client';
import { useToast } from '../../shared/Toast';

type Quest = {
  def_id: number;
  key: string;
  title: string;
  condition_type: string;
  target_progress: number;
  progress: number;
  reward_credits: number;
  reward_metal: number;
  reward_silicon: number;
  reward_hydrogen: number;
  date: string;
  completed: boolean;
  claimed: boolean;
};

export function DailyQuestScreen() {
  const qc = useQueryClient();
  const toast = useToast();

  const quests = useQuery({
    queryKey: ['daily-quests'],
    queryFn: () => api.get<{ quests: Quest[] }>('/api/daily-quests'),
    refetchInterval: 30000,
  });

  const claim = useMutation({
    mutationFn: (defID: number) =>
      api.post<{ reward_credits: number; reward_metal: number; reward_silicon: number; reward_hydrogen: number }>(
        `/api/daily-quests/${defID}/claim`
      ),
    onSuccess: (res) => {
      void qc.invalidateQueries({ queryKey: ['daily-quests'] });
      void qc.invalidateQueries({ queryKey: ['planets'] });
      void qc.invalidateQueries({ queryKey: ['me'] });
      const parts: string[] = [];
      if (res.reward_credits > 0) parts.push(`+${res.reward_credits} кр`);
      if (res.reward_metal > 0) parts.push(`+${res.reward_metal} M`);
      if (res.reward_silicon > 0) parts.push(`+${res.reward_silicon} Si`);
      if (res.reward_hydrogen > 0) parts.push(`+${res.reward_hydrogen} H`);
      toast.show('success', 'Награда получена', parts.join(', ') || '—');
    },
    onError: (err) => {
      toast.show('danger', 'Ошибка', err instanceof Error ? err.message : 'Не удалось получить награду');
    },
  });

  const list = quests.data?.quests ?? [];

  return (
    <div style={{ padding: 16, display: 'flex', flexDirection: 'column', gap: 12 }}>
      <h2 style={{ margin: 0 }}>Ежедневные задания</h2>
      <div style={{ fontSize: 13, color: 'var(--ox-fg-muted)' }}>
        Сбрасываются в 00:00 UTC. Завершите — заберите награду.
      </div>

      {quests.isLoading && <div>Загрузка…</div>}
      {!quests.isLoading && list.length === 0 && (
        <div className="ox-panel" style={{ padding: 16 }}>Заданий пока нет.</div>
      )}

      {list.map((q) => {
        const pct = Math.min(100, Math.round((q.progress / q.target_progress) * 100));
        const rewardParts: string[] = [];
        if (q.reward_credits > 0) rewardParts.push(`${q.reward_credits} кр`);
        if (q.reward_metal > 0) rewardParts.push(`${q.reward_metal} M`);
        if (q.reward_silicon > 0) rewardParts.push(`${q.reward_silicon} Si`);
        if (q.reward_hydrogen > 0) rewardParts.push(`${q.reward_hydrogen} H`);

        return (
          <div key={q.def_id} className="ox-panel" style={{ padding: 12 }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 6 }}>
              <strong>{q.title}</strong>
              {q.claimed ? (
                <span style={{ color: 'var(--ox-fg-muted)', fontSize: 13 }}>Получено</span>
              ) : q.completed ? (
                <button
                  className="ox-btn ox-btn-primary"
                  onClick={() => claim.mutate(q.def_id)}
                  disabled={claim.isPending}
                >
                  Забрать награду
                </button>
              ) : (
                <span style={{ fontSize: 13, color: 'var(--ox-fg-muted)' }}>
                  {q.progress} / {q.target_progress}
                </span>
              )}
            </div>
            <div style={{ height: 6, background: 'var(--ox-bg-deep)', borderRadius: 3, overflow: 'hidden' }}>
              <div
                style={{
                  width: `${pct}%`,
                  height: '100%',
                  background: q.completed ? 'var(--ox-success, #4caf50)' : 'var(--ox-accent, #4a90e2)',
                  transition: 'width 0.3s',
                }}
              />
            </div>
            {rewardParts.length > 0 && (
              <div style={{ marginTop: 6, fontSize: 12, color: 'var(--ox-fg-muted)' }}>
                Награда: {rewardParts.join(', ')}
              </div>
            )}
          </div>
        );
      })}
    </div>
  );
}
