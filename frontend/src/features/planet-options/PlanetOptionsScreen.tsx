import { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';
import type { Planet } from '@/api/types';
import { Confirm } from '@/ui/Confirm';
import { useToast } from '@/ui/Toast';

interface PlanetOptionsScreenProps {
  planet: Planet;
  planets: Planet[];
  homePlanetId: string | null;
  onBack: () => void;
}

export function PlanetOptionsScreen({
  planet,
  planets,
  homePlanetId,
  onBack,
}: PlanetOptionsScreenProps) {
  const qc = useQueryClient();
  const toast = useToast();

  const [newName, setNewName] = useState(planet.name);
  const [confirmSetHome, setConfirmSetHome] = useState(false);
  const [confirmAbandon, setConfirmAbandon] = useState(false);
  const [abandonCode, setAbandonCode] = useState('');

  const rename = useMutation({
    mutationFn: (name: string) =>
      api.patch<{ status: string }>(`/api/planets/${planet.id}`, { name }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['planets'] });
      toast.show('success', 'Планета переименована', `Новое имя: ${newName}`);
    },
    onError: (err) => {
      toast.show('danger', 'Ошибка переименования', err instanceof Error ? err.message : '');
    },
  });

  const setHome = useMutation({
    mutationFn: () => api.post(`/api/planets/${planet.id}/set-home`, {}),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['planets'] });
      setConfirmSetHome(false);
      toast.show('success', 'Главная планета', `${planet.name} теперь ваша главная планета`);
      onBack();
    },
    onError: (err) => {
      toast.show('danger', 'Ошибка', err instanceof Error ? err.message : '');
    },
  });

  const abandon = useMutation({
    mutationFn: () => api.delete(`/api/planets/${planet.id}`),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['planets'] });
      setConfirmAbandon(false);
      setAbandonCode('');
      toast.show('success', 'Планета покинута', `${planet.name} была удалена`);
      onBack();
    },
    onError: (err) => {
      toast.show('danger', 'Ошибка при удалении', err instanceof Error ? err.message : '');
    },
  });

  const isHome = planet.id === homePlanetId;
  const canSetHome = !planet.is_moon && !isHome;
  const canAbandon = !planet.is_moon;
  const expectedCode = `LEAVE${planet.id.substring(0, 4).toUpperCase()}`;

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20, padding: '16px 0' }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 8 }}>
        <button
          type="button"
          className="btn btn-ghost btn-sm"
          onClick={onBack}
          style={{ padding: '4px 8px', fontSize: 16 }}
        >
          ← Назад
        </button>
        <h2 style={{ margin: 0, fontSize: 18, fontFamily: 'var(--ox-font)', fontWeight: 700 }}>
          Параметры: {planet.name}
        </h2>
      </div>

      {/* Переименование */}
      <div className="ox-panel" style={{ padding: 20 }}>
        <div style={{ fontSize: 14, fontWeight: 700, letterSpacing: '0.06em', textTransform: 'uppercase', color: 'var(--ox-fg-muted)', marginBottom: 12 }}>
          Переименовать
        </div>
        <div style={{ display: 'flex', gap: 8, alignItems: 'flex-end' }}>
          <div style={{ flex: 1, minWidth: 0 }}>
            <input
              type="text"
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              maxLength={50}
              style={{ width: '100%', boxSizing: 'border-box' }}
              placeholder="Новое имя"
            />
          </div>
          <button
            type="button"
            className={`btn${rename.isPending || newName === planet.name || newName.trim().length === 0 ? ' btn-ghost' : ''} btn-sm`}
            disabled={rename.isPending || newName === planet.name || newName.trim().length === 0}
            onClick={() => rename.mutate(newName.trim())}
          >
            {rename.isPending ? '…' : 'Сохранить'}
          </button>
        </div>
      </div>

      {/* Установить главной */}
      {canSetHome && (
        <div className="ox-panel" style={{ padding: 20 }}>
          <div style={{ fontSize: 14, fontWeight: 700, letterSpacing: '0.06em', textTransform: 'uppercase', color: 'var(--ox-fg-muted)', marginBottom: 12 }}>
            Главная планета
          </div>
          <div style={{ fontSize: 15, color: 'var(--ox-fg-dim)', marginBottom: 12 }}>
            Текущая главная: {planets.find((p) => p.id === homePlanetId)?.name || '—'}
          </div>
          <button
            type="button"
            className="btn btn-sm"
            onClick={() => setConfirmSetHome(true)}
            disabled={setHome.isPending}
          >
            Установить как главную
          </button>
        </div>
      )}

      {isHome && (
        <div className="ox-panel" style={{ padding: 20, background: 'rgba(76, 175, 80, 0.05)', border: '1px solid rgba(76, 175, 80, 0.2)' }}>
          <div style={{ fontSize: 15, color: 'var(--ox-success)', fontWeight: 500 }}>
            ✓ Это ваша главная планета
          </div>
        </div>
      )}

      {planet.is_moon && !canSetHome && (
        <div className="ox-panel" style={{ padding: 20, background: 'rgba(158, 158, 158, 0.05)', border: '1px solid rgba(158, 158, 158, 0.2)' }}>
          <div style={{ fontSize: 15, color: 'var(--ox-fg-dim)', fontWeight: 500 }}>
            ℹ Луны не могут быть главной планетой
          </div>
        </div>
      )}

      {/* Покинуть планету */}
      {canAbandon && (
        <div className="ox-panel" style={{ padding: 20, background: 'rgba(239, 68, 68, 0.05)', border: '1px solid rgba(239, 68, 68, 0.2)' }}>
          <div style={{ fontSize: 14, fontWeight: 700, letterSpacing: '0.06em', textTransform: 'uppercase', color: 'var(--ox-danger)', marginBottom: 12 }}>
            Покинуть планету
          </div>
          <div style={{ fontSize: 15, color: 'var(--ox-fg-dim)', marginBottom: 12 }}>
            Это действие необратимо. Введите код подтверждения для удаления.
          </div>
          <div style={{ fontSize: 14, color: 'var(--ox-fg-muted)', fontFamily: 'var(--ox-mono)', marginBottom: 12, padding: '8px 12px', background: 'var(--ox-bg)', borderRadius: 4 }}>
            Код: <span style={{ fontWeight: 700 }}>{expectedCode}</span>
          </div>
          <div style={{ display: 'flex', gap: 8, alignItems: 'flex-end' }}>
            <div style={{ flex: 1, minWidth: 0 }}>
              <input
                type="text"
                value={abandonCode}
                onChange={(e) => setAbandonCode(e.target.value.toUpperCase())}
                placeholder={expectedCode}
                style={{ width: '100%', boxSizing: 'border-box' }}
              />
            </div>
            <button
              type="button"
              className={`btn btn-danger${abandonCode !== expectedCode ? ' btn-ghost' : ''} btn-sm`}
              disabled={abandon.isPending || abandonCode !== expectedCode}
              onClick={() => setConfirmAbandon(true)}
            >
              {abandon.isPending ? '…' : 'Покинуть'}
            </button>
          </div>
        </div>
      )}

      {!canAbandon && planet.is_moon && (
        <div className="ox-panel" style={{ padding: 20, background: 'rgba(158, 158, 158, 0.05)', border: '1px solid rgba(158, 158, 158, 0.2)' }}>
          <div style={{ fontSize: 15, color: 'var(--ox-fg-dim)', fontWeight: 500 }}>
            ℹ Луны удаляются автоматически при уничтожении планеты
          </div>
        </div>
      )}

      {/* Confirm dialogs */}
      {confirmSetHome && (
        <Confirm
          message={`Установить "${planet.name}" как главную планету?`}
          confirmLabel="Да"
          cancelLabel="Отмена"
          onConfirm={() => setHome.mutate()}
          onCancel={() => setConfirmSetHome(false)}
        />
      )}

      {confirmAbandon && (
        <Confirm
          message={`Вы уверены? Планета "${planet.name}" будет удалена без возможности восстановления.`}
          danger
          confirmLabel="Удалить"
          cancelLabel="Отмена"
          onConfirm={() => abandon.mutate()}
          onCancel={() => setConfirmAbandon(false)}
        />
      )}
    </div>
  );
}
