import { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';
import type { Planet } from '@/api/types';
import { Confirm } from '@/ui/Confirm';
import { useToast } from '@/ui/Toast';
import { useTranslation } from '@/i18n/i18n';

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
  const { t } = useTranslation('planetOptionsUi');
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
      toast.show('success', t('toastRenameTitle'), t('toastRenameBody', { name: newName }));
    },
    onError: (err) => {
      toast.show('danger', t('toastRenameErr'), err instanceof Error ? err.message : '');
    },
  });

  const setHome = useMutation({
    mutationFn: () => api.post(`/api/planets/${planet.id}/set-home`, {}),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['planets'] });
      setConfirmSetHome(false);
      toast.show('success', t('toastHomeTitle'), t('toastHomeBody', { name: planet.name }));
      onBack();
    },
    onError: (err) => {
      toast.show('danger', t('toastHomeErr'), err instanceof Error ? err.message : '');
    },
  });

  const abandon = useMutation({
    mutationFn: () => api.delete(`/api/planets/${planet.id}`),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['planets'] });
      setConfirmAbandon(false);
      setAbandonCode('');
      toast.show('success', t('toastAbandonTitle'), t('toastAbandonBody', { name: planet.name }));
      onBack();
    },
    onError: (err) => {
      toast.show('danger', t('toastAbandonErr'), err instanceof Error ? err.message : '');
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
          {t('backBtn')}
        </button>
        <h2 style={{ margin: 0, fontSize: 18, fontFamily: 'var(--ox-font)', fontWeight: 700 }}>
          {t('titlePrefix')} {planet.name}
        </h2>
      </div>

      <div className="ox-panel" style={{ padding: 20 }}>
        <div style={{ fontSize: 14, fontWeight: 700, letterSpacing: '0.06em', textTransform: 'uppercase', color: 'var(--ox-fg-muted)', marginBottom: 12 }}>
          {t('sectionRename')}
        </div>
        <div style={{ display: 'flex', gap: 8, alignItems: 'flex-end' }}>
          <div style={{ flex: 1, minWidth: 0 }}>
            <input
              type="text"
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              maxLength={50}
              style={{ width: '100%', boxSizing: 'border-box' }}
              placeholder={t('renamePlaceholder')}
            />
          </div>
          <button
            type="button"
            className={`btn${rename.isPending || newName === planet.name || newName.trim().length === 0 ? ' btn-ghost' : ''} btn-sm`}
            disabled={rename.isPending || newName === planet.name || newName.trim().length === 0}
            onClick={() => rename.mutate(newName.trim())}
          >
            {rename.isPending ? '…' : t('saveBtn')}
          </button>
        </div>
      </div>

      {canSetHome && (
        <div className="ox-panel" style={{ padding: 20 }}>
          <div style={{ fontSize: 14, fontWeight: 700, letterSpacing: '0.06em', textTransform: 'uppercase', color: 'var(--ox-fg-muted)', marginBottom: 12 }}>
            {t('sectionHome')}
          </div>
          <div style={{ fontSize: 15, color: 'var(--ox-fg-dim)', marginBottom: 12 }}>
            {t('currentHome', { name: planets.find((p) => p.id === homePlanetId)?.name ?? '—' })}
          </div>
          <button
            type="button"
            className="btn btn-sm"
            onClick={() => setConfirmSetHome(true)}
            disabled={setHome.isPending}
          >
            {t('setHomeBtn')}
          </button>
        </div>
      )}

      {isHome && (
        <div className="ox-panel" style={{ padding: 20, background: 'rgba(76, 175, 80, 0.05)', border: '1px solid rgba(76, 175, 80, 0.2)' }}>
          <div style={{ fontSize: 15, color: 'var(--ox-success)', fontWeight: 500 }}>
            {t('isHomeMsg')}
          </div>
        </div>
      )}

      {planet.is_moon && !canSetHome && (
        <div className="ox-panel" style={{ padding: 20, background: 'rgba(158, 158, 158, 0.05)', border: '1px solid rgba(158, 158, 158, 0.2)' }}>
          <div style={{ fontSize: 15, color: 'var(--ox-fg-dim)', fontWeight: 500 }}>
            {t('moonNoHomeMsg')}
          </div>
        </div>
      )}

      {canAbandon && (
        <div className="ox-panel" style={{ padding: 20, background: 'rgba(239, 68, 68, 0.05)', border: '1px solid rgba(239, 68, 68, 0.2)' }}>
          <div style={{ fontSize: 14, fontWeight: 700, letterSpacing: '0.06em', textTransform: 'uppercase', color: 'var(--ox-danger)', marginBottom: 12 }}>
            {t('sectionAbandon')}
          </div>
          <div style={{ fontSize: 15, color: 'var(--ox-fg-dim)', marginBottom: 12 }}>
            {t('abandonDesc')}
          </div>
          <div style={{ fontSize: 14, color: 'var(--ox-fg-muted)', fontFamily: 'var(--ox-mono)', marginBottom: 12, padding: '8px 12px', background: 'var(--ox-bg)', borderRadius: 4 }}>
            {t('abandonCodeLabel')} <span style={{ fontWeight: 700 }}>{expectedCode}</span>
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
              {abandon.isPending ? '…' : t('abandonBtn')}
            </button>
          </div>
        </div>
      )}

      {!canAbandon && planet.is_moon && (
        <div className="ox-panel" style={{ padding: 20, background: 'rgba(158, 158, 158, 0.05)', border: '1px solid rgba(158, 158, 158, 0.2)' }}>
          <div style={{ fontSize: 15, color: 'var(--ox-fg-dim)', fontWeight: 500 }}>
            {t('moonAbandonMsg')}
          </div>
        </div>
      )}

      {confirmSetHome && (
        <Confirm
          message={t('setHomeConfirm', { name: planet.name })}
          confirmLabel={t('setHomeConfirmBtn')}
          cancelLabel={t('cancelBtn')}
          onConfirm={() => setHome.mutate()}
          onCancel={() => setConfirmSetHome(false)}
        />
      )}

      {confirmAbandon && (
        <Confirm
          message={t('abandonConfirm', { name: planet.name })}
          danger
          confirmLabel={t('abandonConfirmBtn')}
          cancelLabel={t('cancelBtn')}
          onConfirm={() => abandon.mutate()}
          onCancel={() => setConfirmAbandon(false)}
        />
      )}
    </div>
  );
}
