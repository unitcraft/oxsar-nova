// План 67 Ф.5.2 — кастомные ранги с гранулярными правами (D-014, U-005).
//
// Backend endpoints:
//   GET    /api/alliances/{id}/ranks
//   POST   /api/alliances/{id}/ranks                   { name, position, permissions }
//   PATCH  /api/alliances/{id}/ranks/{rank_id}         { name?, position?, permissions? }
//   DELETE /api/alliances/{id}/ranks/{rank_id}
//
// Право: can_manage_ranks (или owner) — см. permissions.go. Все мутации
// идемпотентны через Idempotency-Key (R9).

import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api, genIdempotencyKey } from '@/api/client';
import { Confirm } from '@/ui/Confirm';
import { useToast } from '@/ui/Toast';
import { useTranslation } from '@/i18n/i18n';
import { PERMISSION_KEYS, type PermissionKey, type PermissionMap } from './permissions';

interface Rank {
  id: string;
  alliance_id: string;
  name: string;
  position: number;
  permissions: PermissionMap;
  created_at: string;
}

const DEFAULT_DRAFT = (): { name: string; position: number; perms: PermissionMap } => ({
  name: '',
  position: 100,
  perms: {},
});

export function RanksPanel({
  allianceID,
  canManage,
}: {
  allianceID: string;
  canManage: boolean;
}) {
  const { t } = useTranslation('alliance');
  const qc = useQueryClient();
  const toast = useToast();

  const ranks = useQuery({
    queryKey: ['alliances', allianceID, 'ranks'],
    queryFn: () =>
      api.get<{ ranks: Rank[] | null }>(`/api/alliances/${allianceID}/ranks`).then(
        (r) => r.ranks ?? [],
      ),
  });

  const [editingID, setEditingID] = useState<string | null>(null); // null = создание | id = edit | '' = форма закрыта
  const [draft, setDraft] = useState<{ name: string; position: number; perms: PermissionMap }>(
    DEFAULT_DRAFT(),
  );
  const [confirmDeleteID, setConfirmDeleteID] = useState<string | null>(null);

  const create = useMutation({
    mutationFn: () =>
      api.post<{ rank: Rank }>(
        `/api/alliances/${allianceID}/ranks`,
        { name: draft.name.trim(), position: draft.position, permissions: draft.perms },
        { idempotencyKey: genIdempotencyKey() },
      ),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['alliances', allianceID, 'ranks'] });
      setEditingID('');
      setDraft(DEFAULT_DRAFT());
      toast.show('success', t('ranks.title'), t('ranks.created'));
    },
    onError: (e) => toast.show('danger', t('createErr'), e instanceof Error ? e.message : ''),
  });

  const update = useMutation({
    mutationFn: ({ id }: { id: string }) =>
      api.patch<void>(
        `/api/alliances/${allianceID}/ranks/${id}`,
        { name: draft.name.trim(), position: draft.position, permissions: draft.perms },
        { idempotencyKey: genIdempotencyKey() },
      ),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['alliances', allianceID, 'ranks'] });
      setEditingID('');
      setDraft(DEFAULT_DRAFT());
      toast.show('success', t('ranks.title'), t('ranks.updated'));
    },
    onError: (e) => toast.show('danger', t('createErr'), e instanceof Error ? e.message : ''),
  });

  const remove = useMutation({
    mutationFn: (id: string) =>
      api.delete<void>(`/api/alliances/${allianceID}/ranks/${id}`),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['alliances', allianceID, 'ranks'] });
      setConfirmDeleteID(null);
      toast.show('info', t('ranks.title'), t('ranks.deleted'));
    },
    onError: (e) => toast.show('danger', t('createErr'), e instanceof Error ? e.message : ''),
  });

  const startCreate = () => {
    setEditingID(null);
    setDraft(DEFAULT_DRAFT());
  };
  const startEdit = (r: Rank) => {
    setEditingID(r.id);
    setDraft({ name: r.name, position: r.position, perms: { ...r.permissions } });
  };
  const cancel = () => {
    setEditingID('');
    setDraft(DEFAULT_DRAFT());
  };

  const togglePerm = (k: PermissionKey) =>
    setDraft((d) => ({ ...d, perms: { ...d.perms, [k]: !d.perms[k] } }));

  const list = ranks.data ?? [];
  const sorted = [...list].sort((a, b) => a.position - b.position);
  const formOpen = editingID !== '';
  const isCreating = editingID === null;

  return (
    <div className="ox-panel" style={{ padding: '12px 16px', display: 'flex', flexDirection: 'column', gap: 10 }}>
      <div
        style={{
          fontSize: 13,
          fontWeight: 700,
          letterSpacing: '0.08em',
          textTransform: 'uppercase',
          color: 'var(--ox-fg-muted)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
        }}
      >
        <span>{t('ranks.title')}</span>
        {canManage && !formOpen && (
          <button type="button" className="btn-ghost btn-sm" onClick={startCreate}>
            + {t('ranks.createBtn')}
          </button>
        )}
      </div>

      {ranks.isLoading && <div className="ox-skeleton" style={{ height: 60 }} />}

      {!ranks.isLoading && sorted.length === 0 && !formOpen && (
        <div style={{ fontSize: 15, color: 'var(--ox-fg-dim)', fontStyle: 'italic' }}>
          {t('ranks.empty')}
        </div>
      )}

      {sorted.length > 0 && (
        <table className="ox-table" style={{ margin: 0, fontSize: 14 }}>
          <thead>
            <tr>
              <th>{t('ranks.colPosition')}</th>
              <th>{t('ranks.colName')}</th>
              <th>{t('ranks.colPerms')}</th>
              {canManage && <th />}
            </tr>
          </thead>
          <tbody>
            {sorted.map((r) => (
              <tr key={r.id}>
                <td style={{ fontFamily: 'var(--ox-mono)', color: 'var(--ox-fg-muted)' }}>
                  {r.position}
                </td>
                <td style={{ fontWeight: 700 }}>{r.name}</td>
                <td style={{ color: 'var(--ox-fg-dim)', fontSize: 13 }}>
                  {permsSummary(r.permissions, t)}
                </td>
                {canManage && (
                  <td>
                    <div style={{ display: 'flex', gap: 4 }}>
                      <button
                        type="button"
                        className="btn-ghost btn-sm"
                        onClick={() => startEdit(r)}
                        aria-label={t('ranks.editBtn')}
                      >
                        ✎
                      </button>
                      <button
                        type="button"
                        className="btn-ghost btn-sm"
                        style={{ color: 'var(--ox-danger)' }}
                        onClick={() => setConfirmDeleteID(r.id)}
                        aria-label={t('ranks.deleteBtn')}
                      >
                        ✕
                      </button>
                    </div>
                  </td>
                )}
              </tr>
            ))}
          </tbody>
        </table>
      )}

      {canManage && formOpen && (
        <div
          style={{
            borderTop: '1px solid var(--ox-border)',
            paddingTop: 10,
            display: 'flex',
            flexDirection: 'column',
            gap: 8,
          }}
        >
          <div style={{ fontSize: 14, fontWeight: 700 }}>
            {isCreating ? t('ranks.formCreate') : t('ranks.formEdit')}
          </div>
          <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
            <label style={{ flex: '1 1 220px', display: 'flex', flexDirection: 'column', gap: 2 }}>
              <span style={{ fontSize: 13, color: 'var(--ox-fg-dim)' }}>{t('ranks.colName')}</span>
              <input
                value={draft.name}
                onChange={(e) => setDraft({ ...draft, name: e.target.value })}
                maxLength={32}
                style={{ width: '100%' }}
                placeholder={t('ranks.namePlaceholder')}
              />
            </label>
            <label style={{ flex: '0 0 100px', display: 'flex', flexDirection: 'column', gap: 2 }}>
              <span style={{ fontSize: 13, color: 'var(--ox-fg-dim)' }}>
                {t('ranks.colPosition')}
              </span>
              <input
                type="number"
                value={draft.position}
                onChange={(e) =>
                  setDraft({ ...draft, position: Number.parseInt(e.target.value || '0', 10) })
                }
                style={{ width: '100%' }}
              />
            </label>
          </div>

          <fieldset
            style={{
              border: '1px solid var(--ox-border)',
              borderRadius: 4,
              padding: '6px 10px',
              display: 'flex',
              flexDirection: 'column',
              gap: 4,
            }}
          >
            <legend style={{ fontSize: 13, color: 'var(--ox-fg-muted)', padding: '0 4px' }}>
              {t('ranks.permsLegend')}
            </legend>
            {PERMISSION_KEYS.map((p) => (
              <label
                key={p}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 8,
                  fontSize: 14,
                  cursor: 'pointer',
                }}
              >
                <input
                  type="checkbox"
                  checked={draft.perms[p] === true}
                  onChange={() => togglePerm(p)}
                />
                <span>{t(`perm.${p}`)}</span>
              </label>
            ))}
          </fieldset>

          <div style={{ display: 'flex', gap: 8 }}>
            <button
              type="button"
              className="btn btn-sm"
              disabled={
                draft.name.trim().length === 0 ||
                create.isPending ||
                update.isPending
              }
              onClick={() =>
                isCreating ? create.mutate() : update.mutate({ id: editingID as string })
              }
            >
              {create.isPending || update.isPending ? '…' : t('ranks.saveBtn')}
            </button>
            <button type="button" className="btn-ghost btn-sm" onClick={cancel}>
              {t('cancelBtn')}
            </button>
          </div>
        </div>
      )}

      {confirmDeleteID && (
        <Confirm
          title={t('ranks.deleteBtn')}
          message={t('ranks.deleteConfirm', {
            name: list.find((r) => r.id === confirmDeleteID)?.name ?? '',
          })}
          confirmLabel={t('confirmBtn')}
          danger
          onConfirm={() => remove.mutate(confirmDeleteID)}
          onCancel={() => setConfirmDeleteID(null)}
        />
      )}
    </div>
  );
}

function permsSummary(perms: PermissionMap, t: (k: string) => string): string {
  const enabled = PERMISSION_KEYS.filter((k) => perms[k] === true);
  if (enabled.length === 0) return t('ranks.permsNone');
  if (enabled.length === PERMISSION_KEYS.length) return t('ranks.permsAll');
  return enabled.map((k) => t(`perm.${k}.short`)).join(', ');
}
