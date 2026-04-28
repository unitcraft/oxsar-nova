// S-015 Alliance diplomacy (план 72 Ф.3 Spring 2 ч.1).
//
// Pixel-perfect зеркало legacy `templates/standard/ally_diplomacy.tpl`.
// Таблица текущих дипотношений + форма «применить статус» (5 enum
// статусов из плана 67 D-014 B1: protection/confederation/war/trade/
// ceasefire).
//
// Endpoint:
//   GET /api/alliances/{id}/relations         → relations[]
//   PUT /api/alliances/{id}/relations         → propose
//   PUT /api/alliances/{id}/relations/{i}/accept
//   POST /api/alliances/{id}/relations/{i}/reject
//   DELETE /api/alliances/{id}/relations/{i}
//
// Право: can_manage_diplomacy (или owner). UI открывает действия
// только owner'у (см. P67.S5.B про rank_id), backend дополнительно
// валидирует.

import { useState } from 'react';
import { Navigate, Link } from 'react-router-dom';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  acceptRelation,
  breakRelation,
  fetchRelations,
  proposeRelation,
  rejectRelation,
} from '@/api/alliance';
import type { ApiError } from '@/api/client';
import { QK } from '@/api/query-keys';
import type { AllianceRelation, AllianceRelationStatus } from '@/api/types';
import { useAuthStore } from '@/stores/auth';
import { useTranslation } from '@/i18n/i18n';
import { relationStatusKey, useMyAlliance } from './common';

const STATUSES: AllianceRelationStatus[] = [
  'protection',
  'confederation',
  'war',
  'trade',
  'ceasefire',
];

export function AllianceDiplomacyScreen() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  const my = useMyAlliance();
  const userId = useAuthStore((s) => s.userId);
  const [target, setTarget] = useState('');
  const [status, setStatus] = useState<AllianceRelationStatus>('protection');
  const [message, setMessage] = useState('');
  const [errMsg, setErrMsg] = useState<string | null>(null);

  const allianceID = my.data?.alliance.id ?? '';

  const relsQ = useQuery({
    queryKey: QK.allianceRelations(allianceID),
    queryFn: () => fetchRelations(allianceID),
    enabled: !!allianceID,
  });

  const propose = useMutation({
    mutationFn: () =>
      proposeRelation(allianceID, {
        target_id: target,
        status,
        message: message || undefined,
      }),
    onSuccess: () => {
      setTarget('');
      setMessage('');
      void qc.invalidateQueries({ queryKey: QK.allianceRelations(allianceID) });
    },
    onError: (e) => setErrMsg((e as ApiError).message),
  });

  const accept = useMutation({
    mutationFn: (initiatorID: string) => acceptRelation(allianceID, initiatorID),
    onSuccess: () =>
      void qc.invalidateQueries({ queryKey: QK.allianceRelations(allianceID) }),
  });
  const rejectFn = useMutation({
    mutationFn: (initiatorID: string) => rejectRelation(allianceID, initiatorID),
    onSuccess: () =>
      void qc.invalidateQueries({ queryKey: QK.allianceRelations(allianceID) }),
  });
  const breakFn = useMutation({
    mutationFn: (initiatorID: string) => breakRelation(allianceID, initiatorID),
    onSuccess: () =>
      void qc.invalidateQueries({ queryKey: QK.allianceRelations(allianceID) }),
  });

  if (my.isLoading) return <div className="idiv">…</div>;
  if (!my.data) return <Navigate to="/alliance" replace />;

  const al = my.data.alliance;
  const isOwner = !!userId && userId === al.owner_id;
  const relations = relsQ.data?.relations ?? [];

  return (
    <>
      <div className="idiv">
        <Link to="/alliance/me">← {al.tag}</Link>
      </div>

      <table className="ntable">
        <thead>
          <tr>
            <th colSpan={5}>
              {t('alliance', 'allianceDiplomacyRelationships')}
            </th>
          </tr>
          <tr>
            <th>#</th>
            <th>{t('alliance', 'alliances')}</th>
            <th>{t('alliance', 'founder')}</th>
            <th>{t('alliance', 'established')}</th>
            <th>{t('alliance', 'status')}</th>
          </tr>
        </thead>
        <tbody>
          {relations.length === 0 && (
            <tr>
              <td className="center" colSpan={5}>
                {t('alliance', 'diplomacy.empty')}
              </td>
            </tr>
          )}
          {relations.map((r, i) => (
            <DiplomacyRow
              key={`${r.initiator_id}-${r.target_id}`}
              n={i + 1}
              rel={r}
              myAllianceID={al.id}
              canAct={isOwner}
              onAccept={() => accept.mutate(r.initiator_id)}
              onReject={() => rejectFn.mutate(r.initiator_id)}
              onBreak={() => breakFn.mutate(r.initiator_id)}
            />
          ))}
        </tbody>
      </table>

      {isOwner && (
        <form
          method="post"
          onSubmit={(ev) => {
            ev.preventDefault();
            if (!propose.isPending && target) propose.mutate();
          }}
        >
          <table className="ntable">
            <thead>
              <tr>
                <th colSpan={2}>
                  {t('alliance', 'applicationForRelationship')}
                </th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td>
                  <label htmlFor="tag">{t('alliance', 'tag')}</label>
                </td>
                <td>
                  <input
                    type="text"
                    name="tag"
                    id="tag"
                    placeholder={t('alliance', 'diplomacy.targetPlaceholder')}
                    value={target}
                    onChange={(e) => setTarget(e.target.value)}
                  />
                </td>
              </tr>
              <tr>
                <td>
                  <label htmlFor="status">{t('alliance', 'status')}</label>
                </td>
                <td>
                  <select
                    name="status"
                    id="status"
                    value={status}
                    onChange={(e) =>
                      setStatus(e.target.value as AllianceRelationStatus)
                    }
                  >
                    {STATUSES.map((s) => (
                      <option key={s} value={s}>
                        {t('alliance', relationStatusKey(s))}
                      </option>
                    ))}
                  </select>
                </td>
              </tr>
              <tr>
                <td>
                  <label htmlFor="message">{t('alliance', 'message')}</label>
                </td>
                <td>
                  <textarea
                    name="message"
                    id="message"
                    cols={35}
                    rows={4}
                    maxLength={2000}
                    value={message}
                    onChange={(e) => setMessage(e.target.value)}
                  />
                </td>
              </tr>
              {errMsg && (
                <tr>
                  <td className="center" colSpan={2}>
                    <span className="false">{errMsg}</span>
                  </td>
                </tr>
              )}
              <tr>
                <td className="center" colSpan={2}>
                  <input
                    type="submit"
                    name="SendContract"
                    value={t('alliance', 'diplomacy.proposeBtn')}
                    className="button"
                    disabled={!target || propose.isPending}
                  />
                </td>
              </tr>
            </tbody>
          </table>
        </form>
      )}
    </>
  );
}

function DiplomacyRow({
  n,
  rel,
  myAllianceID,
  canAct,
  onAccept,
  onReject,
  onBreak,
}: {
  n: number;
  rel: AllianceRelation;
  myAllianceID: string;
  canAct: boolean;
  onAccept: () => void;
  onReject: () => void;
  onBreak: () => void;
}) {
  const { t } = useTranslation();
  const initiatedByMe = rel.initiator_id === myAllianceID;
  const counterpartTag = initiatedByMe ? rel.target_tag : rel.initiator_tag;
  const counterpartName = initiatedByMe ? rel.target_name : rel.initiator_name;
  const time = initiatedByMe ? rel.proposed_at : rel.proposed_at;
  return (
    <tr>
      <td>{n}</td>
      <td>
        [{counterpartTag}] {counterpartName}
      </td>
      <td>—</td>
      <td>{new Date(time).toLocaleDateString('ru-RU')}</td>
      <td>
        {t('alliance', relationStatusKey(rel.status))} ·{' '}
        {rel.state === 'incoming' && (
          <span className="false">
            {t('alliance', 'diplomacy.statusIncoming')}
          </span>
        )}
        {rel.state === 'outgoing' && (
          <span className="false2">
            {t('alliance', 'diplomacy.statusOutgoing')}
          </span>
        )}
        {rel.state === 'active' && (
          <span className="true">
            {t('alliance', 'diplomacy.statusActive')}
          </span>
        )}
        {canAct && rel.state === 'incoming' && (
          <>
            {' '}
            <button type="button" className="button" onClick={onAccept}>
              {t('alliance', 'accept')}
            </button>{' '}
            <button type="button" className="button" onClick={onReject}>
              {t('alliance', 'refuse')}
            </button>
          </>
        )}
        {canAct && rel.state !== 'incoming' && (
          <>
            {' '}
            <button type="button" className="button" onClick={onBreak}>
              {t('alliance', 'determine')}
            </button>
          </>
        )}
      </td>
    </tr>
  );
}
