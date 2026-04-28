// План 67 Ф.5.1 — три описания альянса (D-041, U-015).
//
// Backend: GET /api/alliances/{id}/descriptions, PATCH с Idempotency-Key.
// Видимость viewer-контекста определяется бэком:
//   - outsider: только description_external (для гостей).
//   - applicant: external + apply.
//   - member: external + internal.
// Owner альянса видит и редактирует все три (через PATCH; backend
// проверит can_change_description / owner-fallback).

import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api, genIdempotencyKey } from '@/api/client';
import { useToast } from '@/ui/Toast';
import { useTranslation } from '@/i18n/i18n';

type ViewerKind = 'member' | 'applicant' | 'outsider';

interface DescriptionView {
  description_external: string;
  description_internal: string;
  description_apply: string;
  description: string;
  viewer: ViewerKind;
}

type Tab = 'external' | 'internal' | 'apply';

export function DescriptionsPanel({
  allianceID,
  canEdit,
}: {
  allianceID: string;
  canEdit: boolean;
}) {
  const { t } = useTranslation('alliance');
  const qc = useQueryClient();
  const toast = useToast();

  const desc = useQuery({
    queryKey: ['alliances', allianceID, 'descriptions'],
    queryFn: () => api.get<DescriptionView>(`/api/alliances/${allianceID}/descriptions`),
  });

  const viewer: ViewerKind = desc.data?.viewer ?? 'outsider';

  // Какие табы показываем — определяется viewer'ом, но если canEdit
  // (owner / can_change_description) — показываем все три, чтобы
  // редактировать тексты для applicant/member из owner-сессии.
  const tabs: Tab[] =
    canEdit
      ? ['external', 'internal', 'apply']
      : viewer === 'member'
        ? ['external', 'internal']
        : viewer === 'applicant'
          ? ['external', 'apply']
          : ['external'];

  const [active, setActive] = useState<Tab>('external');
  const visibleActive: Tab = tabs.includes(active) ? active : 'external';

  const [draft, setDraft] = useState<{ external: string; internal: string; apply: string }>({
    external: '',
    internal: '',
    apply: '',
  });
  const [editing, setEditing] = useState(false);

  // Lazy-init draft from server. Avoid setState-in-render: useState
  // initializer не годится (desc.data приходит позже). Срабатывает
  // один раз когда editing открывается.
  const onStartEdit = () => {
    if (!desc.data) return;
    setDraft({
      external: desc.data.description_external,
      internal: desc.data.description_internal,
      apply: desc.data.description_apply,
    });
    setEditing(true);
  };

  const save = useMutation({
    mutationFn: () =>
      api.patch<void>(
        `/api/alliances/${allianceID}/descriptions`,
        {
          description_external: draft.external,
          description_internal: draft.internal,
          description_apply: draft.apply,
        },
        { idempotencyKey: genIdempotencyKey() },
      ),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['alliances', allianceID, 'descriptions'] });
      setEditing(false);
      toast.show('success', t('descriptions.title'), t('descriptions.saved'));
    },
    onError: (e) => toast.show('danger', t('createErr'), e instanceof Error ? e.message : ''),
  });

  if (desc.isLoading) {
    return (
      <div className="ox-panel" style={{ padding: 12 }}>
        <div className="ox-skeleton" style={{ height: 80 }} />
      </div>
    );
  }

  const labelKey: Record<Tab, string> = {
    external: 'descriptions.tabExternal',
    internal: 'descriptions.tabInternal',
    apply: 'descriptions.tabApply',
  };
  const placeholderKey: Record<Tab, string> = {
    external: 'descriptions.placeholderExternal',
    internal: 'descriptions.placeholderInternal',
    apply: 'descriptions.placeholderApply',
  };
  const text: Record<Tab, string> = {
    external: desc.data?.description_external ?? '',
    internal: desc.data?.description_internal ?? '',
    apply: desc.data?.description_apply ?? '',
  };

  return (
    <div className="ox-panel" style={{ padding: '12px 16px', display: 'flex', flexDirection: 'column', gap: 10 }}>
      <div
        style={{
          fontSize: 13,
          fontWeight: 700,
          letterSpacing: '0.08em',
          textTransform: 'uppercase',
          color: 'var(--ox-fg-muted)',
        }}
      >
        {t('descriptions.title')}
      </div>

      <div className="ox-tabs" role="tablist">
        {tabs.map((tab) => (
          <button
            key={tab}
            type="button"
            role="tab"
            aria-pressed={visibleActive === tab}
            onClick={() => setActive(tab)}
          >
            {t(labelKey[tab])}
          </button>
        ))}
      </div>

      {!editing && (
        <>
          <div
            style={{
              fontSize: 15,
              color: 'var(--ox-fg-dim)',
              minHeight: 60,
              whiteSpace: 'pre-wrap',
              padding: '8px 0',
            }}
          >
            {text[visibleActive] || (
              <span style={{ color: 'var(--ox-fg-muted)', fontStyle: 'italic' }}>
                {t('descriptions.empty')}
              </span>
            )}
          </div>
          {canEdit && (
            <div>
              <button type="button" className="btn-ghost btn-sm" onClick={onStartEdit}>
                ✎ {t('descriptions.editBtn')}
              </button>
            </div>
          )}
        </>
      )}

      {editing && canEdit && (
        <>
          <textarea
            value={draft[visibleActive]}
            onChange={(e) => setDraft({ ...draft, [visibleActive]: e.target.value })}
            rows={6}
            maxLength={4000}
            placeholder={t(placeholderKey[visibleActive])}
            style={{ width: '100%', boxSizing: 'border-box', resize: 'vertical' }}
          />
          <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
            <button
              type="button"
              className="btn btn-sm"
              disabled={save.isPending}
              onClick={() => save.mutate()}
            >
              {save.isPending ? '…' : t('descriptions.saveBtn')}
            </button>
            <button
              type="button"
              className="btn-ghost btn-sm"
              disabled={save.isPending}
              onClick={() => setEditing(false)}
            >
              {t('cancelBtn')}
            </button>
            <span style={{ marginLeft: 'auto', fontSize: 13, color: 'var(--ox-fg-muted)' }}>
              {draft[visibleActive].length} / 4000
            </span>
          </div>
        </>
      )}
    </div>
  );
}
