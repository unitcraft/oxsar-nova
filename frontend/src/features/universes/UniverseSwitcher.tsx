import { useState } from 'react';
import { useQuery, useMutation } from '@tanstack/react-query';
import styles from './UniverseSwitcher.module.css';

interface Universe {
  id: string;
  name: string;
  description: string;
  subdomain: string;
  status: string;
  speed: number;
  online_players?: number;
  current?: boolean;
}

interface UniversesResponse {
  universes: Universe[];
  current: string;
}

interface SwitchResponse {
  redirect_url: string;
  universe_id: string;
  universe_name: string;
}

async function fetchUniverses(): Promise<UniversesResponse> {
  const res = await fetch('/api/universes');
  if (!res.ok) throw new Error('Failed to load universes');
  return res.json() as Promise<UniversesResponse>;
}

async function switchUniverse(targetId: string): Promise<SwitchResponse> {
  const token = localStorage.getItem('access_token') ?? '';
  const res = await fetch(`/api/universes/switch?target=${encodeURIComponent(targetId)}`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok) throw new Error('Failed to switch universe');
  return res.json() as Promise<SwitchResponse>;
}

export function UniverseSwitcher() {
  const [open, setOpen] = useState(false);
  const { data } = useQuery({ queryKey: ['universes'], queryFn: fetchUniverses });

  const mut = useMutation({
    mutationFn: switchUniverse,
    onSuccess: (data) => {
      window.location.href = data.redirect_url;
    },
  });

  if (!data || data.universes.length <= 1) return null;

  const current = data.universes.find((u) => u.id === data.current);

  return (
    <div className={styles.root}>
      <button className={styles.trigger} onClick={() => setOpen((v) => !v)}>
        {current?.name ?? 'Вселенная'} ▾
      </button>
      {open && (
        <div className={styles.dropdown}>
          {data.universes.map((u) => (
            <button
              key={u.id}
              className={`${styles.item} ${u.id === data.current ? styles.active : ''} ${u.status !== 'active' ? styles.disabled : ''}`}
              disabled={u.id === data.current || u.status !== 'active' || mut.isPending}
              onClick={() => { mut.mutate(u.id); setOpen(false); }}
            >
              <span className={styles.itemName}>{u.name}</span>
              <span className={styles.itemMeta}>
                ×{u.speed}
                {u.online_players !== undefined && ` · ${u.online_players} онлайн`}
                {u.status === 'upcoming' && ' · Скоро'}
              </span>
              {u.id === data.current && <span className={styles.badge}>здесь</span>}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
