// S-014 Alliance descriptions (план 72 Ф.3 Spring 2 ч.1).
//
// Pixel-perfect зеркало legacy `templates/standard/manage_ally.tpl`
// (часть с 3 textarea-вкладками: extern/intern/apply).
//
// Endpoint:
//   GET   /api/alliances/{id}/descriptions  → external/internal/apply +
//                                             legacy description + viewer
//   PATCH /api/alliances/{id}/descriptions   — право can_change_description
//                                             (или owner). Idempotency-Key.

import { useEffect, useState } from 'react';
import { Link, Navigate } from 'react-router-dom';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { fetchDescriptions, updateDescriptions } from '@/api/alliance';
import type { ApiError } from '@/api/client';
import { QK } from '@/api/query-keys';
import { useAuthStore } from '@/stores/auth';
import { useTranslation } from '@/i18n/i18n';
import { useMyAlliance } from './common';

type Tab = 'external' | 'internal' | 'apply';

const MAX_LEN = 4000;

export function AllianceDescriptionsScreen() {
  const { t } = useTranslation();
  const my = useMyAlliance();
  const qc = useQueryClient();
  const userId = useAuthStore((s) => s.userId);

  const allianceID = my.data?.alliance.id ?? '';

  const descr = useQuery({
    queryKey: QK.allianceDescriptions(allianceID),
    queryFn: () => fetchDescriptions(allianceID),
    enabled: !!allianceID,
  });

  const [tab, setTab] = useState<Tab>('external');
  const [extern, setExtern] = useState('');
  const [intern, setIntern] = useState('');
  const [apply, setApply] = useState('');
  const [errMsg, setErrMsg] = useState<string | null>(null);

  useEffect(() => {
    if (descr.data) {
      setExtern(descr.data.description_external);
      setIntern(descr.data.description_internal);
      setApply(descr.data.description_apply);
    }
  }, [descr.data]);

  const save = useMutation({
    mutationFn: () =>
      updateDescriptions(allianceID, {
        description_external: extern,
        description_internal: intern,
        description_apply: apply,
      }),
    onSuccess: () =>
      void qc.invalidateQueries({
        queryKey: QK.allianceDescriptions(allianceID),
      }),
    onError: (e) => setErrMsg((e as ApiError).message),
  });

  if (my.isLoading) return <div className="idiv">…</div>;
  if (!my.data) return <Navigate to="/alliance" replace />;

  const al = my.data.alliance;
  const isOwner = !!userId && userId === al.owner_id;
  // can_change_description проверяется бэкендом; UI открывает форму
  // владельцу сразу, прочим — только если viewer === 'member' с правом
  // (rank_id с бэка пока не приходит — план 67 simplifications.md
  // P67.S5.B). На UI здесь — только owner, иначе read-only.
  const canEdit = isOwner;

  return (
    <>
      <div className="idiv">
        <Link to="/alliance/me">← {al.tag}</Link>
      </div>

      <table className="ntable">
        <thead>
          <tr>
            <th>
              <a
                onClick={() => setTab('external')}
                style={{
                  cursor: 'pointer',
                  fontWeight: tab === 'external' ? 'bold' : 'normal',
                  marginRight: 12,
                }}
              >
                {t('alliance', 'externAllianceText')}
              </a>
              <a
                onClick={() => setTab('internal')}
                style={{
                  cursor: 'pointer',
                  fontWeight: tab === 'internal' ? 'bold' : 'normal',
                  marginRight: 12,
                }}
              >
                {t('alliance', 'internAllianceText')}
              </a>
              <a
                onClick={() => setTab('apply')}
                style={{
                  cursor: 'pointer',
                  fontWeight: tab === 'apply' ? 'bold' : 'normal',
                }}
              >
                {t('alliance', 'applicationText')}
              </a>
            </th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>
              {tab === 'external' && (
                <DescTextarea
                  value={extern}
                  onChange={setExtern}
                  readOnly={!canEdit}
                />
              )}
              {tab === 'internal' && (
                <DescTextarea
                  value={intern}
                  onChange={setIntern}
                  readOnly={!canEdit}
                />
              )}
              {tab === 'apply' && (
                <DescTextarea
                  value={apply}
                  onChange={setApply}
                  readOnly={!canEdit}
                />
              )}
            </td>
          </tr>
          {canEdit && (
            <tr>
              <td className="center">
                <input
                  type="button"
                  className="button"
                  value={t('alliance', 'descriptions.saveBtn')}
                  disabled={save.isPending}
                  onClick={() => save.mutate()}
                />
              </td>
            </tr>
          )}
          {errMsg && (
            <tr>
              <td className="center">
                <span className="false">{errMsg}</span>
              </td>
            </tr>
          )}
        </tbody>
      </table>
    </>
  );
}

function DescTextarea({
  value,
  onChange,
  readOnly,
}: {
  value: string;
  onChange: (s: string) => void;
  readOnly: boolean;
}) {
  const { t } = useTranslation();
  return (
    <>
      <textarea
        cols={75}
        rows={15}
        className="center"
        value={value}
        readOnly={readOnly}
        maxLength={MAX_LEN}
        onChange={(e) => onChange(e.target.value)}
      />
      <br />
      {t('alliance', 'marked')} {value.length} / {MAX_LEN}{' '}
      {t('alliance', 'characters') || ''}
    </>
  );
}
