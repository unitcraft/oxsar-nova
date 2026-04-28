// План 67 Ф.5.3 — расширенная дипломатия: 5 enum-статусов
// (friend, neutral, hostile_neutral, nap, war), filter-табы,
// status-badge с цветом, action-кнопки propose/accept/reject/break.
//
// Backend:
//   GET    /api/alliances/{id}/relations
//   PUT    /api/alliances/{id}/relations/{targetID}            { relation }
//   POST   /api/alliances/{id}/relations/{initiatorID}/accept
//   DELETE /api/alliances/{id}/relations/{initiatorID}         (reject / break)
//
// Право на мутации: can_manage_diplomacy (или owner).
// Правило accept: friend/neutral/nap двусторонние (status=pending), war/hostile_neutral
// односторонние.

import { useState, useMemo } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api, genIdempotencyKey } from '@/api/client';
import { useToast } from '@/ui/Toast';
import { useTranslation } from '@/i18n/i18n';
import {
  RELATIONS,
  RELATION_LABEL_KEY,
  RELATION_COLOR,
  isKnownRelation,
  type Relation,
} from './relations';

interface RelationItem {
  target_alliance_id: string;
  target_tag: string;
  target_name: string;
  relation: string;
  status: string;
  initiator: boolean;
  set_at: string;
}

type Filter = 'all' | Relation;

export function DiplomacyPanel({
  allianceID,
  canManage,
}: {
  allianceID: string;
  canManage: boolean;
}) {
  const { t } = useTranslation('alliance');
  const qc = useQueryClient();
  const toast = useToast();
  const [filter, setFilter] = useState<Filter>('all');
  const [targetID, setTargetID] = useState('');
  const [proposeRel, setProposeRel] = useState<Relation>('nap');

  const rels = useQuery({
    queryKey: ['alliances', allianceID, 'relations'],
    queryFn: () =>
      api
        .get<{ relations: RelationItem[] | null }>(`/api/alliances/${allianceID}/relations`)
        .then((r) => r.relations ?? []),
    refetchInterval: 30000,
  });

  const propose = useMutation({
    mutationFn: ({ tid, rel }: { tid: string; rel: Relation }) =>
      api.put<void>(
        `/api/alliances/${allianceID}/relations/${tid}`,
        { relation: rel },
        { idempotencyKey: genIdempotencyKey() },
      ),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['alliances', allianceID, 'relations'] });
      setTargetID('');
      toast.show('success', t('sectionRelations'), t('diplomacy.proposed'));
    },
    onError: (e) => toast.show('danger', t('createErr'), e instanceof Error ? e.message : ''),
  });

  const breakRel = useMutation({
    mutationFn: (tid: string) =>
      api.put<void>(
        `/api/alliances/${allianceID}/relations/${tid}`,
        { relation: 'none' },
        { idempotencyKey: genIdempotencyKey() },
      ),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['alliances', allianceID, 'relations'] }),
    onError: (e) => toast.show('danger', t('createErr'), e instanceof Error ? e.message : ''),
  });

  const accept = useMutation({
    mutationFn: (initiatorID: string) =>
      api.post<void>(
        `/api/alliances/${allianceID}/relations/${initiatorID}/accept`,
        undefined,
        { idempotencyKey: genIdempotencyKey() },
      ),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['alliances', allianceID, 'relations'] }),
    onError: (e) => toast.show('danger', t('createErr'), e instanceof Error ? e.message : ''),
  });

  const reject = useMutation({
    mutationFn: (initiatorID: string) =>
      api.delete<void>(`/api/alliances/${allianceID}/relations/${initiatorID}`),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['alliances', allianceID, 'relations'] }),
    onError: (e) => toast.show('danger', t('createErr'), e instanceof Error ? e.message : ''),
  });

  const filtered = useMemo(() => {
    const list = rels.data ?? [];
    if (filter === 'all') return list;
    return list.filter((r) => r.relation === filter);
  }, [rels.data, filter]);

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
        {t('sectionRelations')}
      </div>

      <div className="ox-tabs" role="tablist" style={{ flexWrap: 'wrap' }}>
        <button
          type="button"
          role="tab"
          aria-pressed={filter === 'all'}
          onClick={() => setFilter('all')}
        >
          {t('diplomacy.filterAll')}
        </button>
        {RELATIONS.map((rel) => (
          <button
            key={rel}
            type="button"
            role="tab"
            aria-pressed={filter === rel}
            onClick={() => setFilter(rel)}
          >
            {t(RELATION_LABEL_KEY[rel])}
          </button>
        ))}
      </div>

      {rels.isLoading && <div className="ox-skeleton" style={{ height: 60 }} />}

      {!rels.isLoading && filtered.length === 0 && (
        <div style={{ fontSize: 15, color: 'var(--ox-fg-dim)', fontStyle: 'italic' }}>
          {t('diplomacy.empty')}
        </div>
      )}

      {filtered.length > 0 && (
        <table className="ox-table" style={{ margin: 0, fontSize: 14 }}>
          <thead>
            <tr>
              <th>{t('diplomacy.colTarget')}</th>
              <th>{t('colRelation')}</th>
              <th>{t('diplomacy.colStatus')}</th>
              {canManage && <th>{t('colActions')}</th>}
            </tr>
          </thead>
          <tbody>
            {filtered.map((r) => {
              const rel: Relation | null = isKnownRelation(r.relation) ? r.relation : null;
              const color = rel ? RELATION_COLOR[rel] : 'var(--ox-fg-dim)';
              const label = rel ? t(RELATION_LABEL_KEY[rel]) : r.relation;
              const isPending = r.status === 'pending';
              return (
                <tr key={`${r.initiator ? 'out' : 'in'}-${r.target_alliance_id}`}>
                  <td style={{ fontFamily: 'var(--ox-mono)' }}>
                    [{r.target_tag}] {r.target_name}
                  </td>
                  <td>
                    <span
                      style={{
                        color,
                        fontWeight: 700,
                        padding: '2px 8px',
                        border: `1px solid ${color}`,
                        borderRadius: 4,
                        fontSize: 13,
                      }}
                    >
                      {label}
                    </span>
                  </td>
                  <td style={{ color: isPending ? 'var(--ox-warning)' : 'var(--ox-fg-dim)' }}>
                    {isPending
                      ? r.initiator
                        ? t('diplomacy.statusOutgoing')
                        : t('diplomacy.statusIncoming')
                      : t('diplomacy.statusActive')}
                  </td>
                  {canManage && (
                    <td>
                      <div style={{ display: 'flex', gap: 4, flexWrap: 'wrap' }}>
                        {!r.initiator && isPending ? (
                          <>
                            <button
                              type="button"
                              className="btn btn-sm"
                              disabled={accept.isPending}
                              onClick={() => accept.mutate(r.target_alliance_id)}
                            >
                              ✓ {t('acceptBtn')}
                            </button>
                            <button
                              type="button"
                              className="btn-ghost btn-sm"
                              disabled={reject.isPending}
                              onClick={() => reject.mutate(r.target_alliance_id)}
                            >
                              ✕ {t('rejectBtn')}
                            </button>
                          </>
                        ) : (
                          <button
                            type="button"
                            className="btn-ghost btn-sm"
                            disabled={breakRel.isPending}
                            onClick={() => breakRel.mutate(r.target_alliance_id)}
                          >
                            {t('diplomacy.breakBtn')}
                          </button>
                        )}
                      </div>
                    </td>
                  )}
                </tr>
              );
            })}
          </tbody>
        </table>
      )}

      {canManage && (
        <div
          style={{
            borderTop: '1px solid var(--ox-border)',
            paddingTop: 10,
            display: 'flex',
            gap: 6,
            alignItems: 'center',
            flexWrap: 'wrap',
          }}
        >
          <input
            placeholder={t('diplomacy.targetPlaceholder')}
            value={targetID}
            onChange={(e) => setTargetID(e.target.value)}
            style={{ flex: 1, minWidth: 200, fontFamily: 'var(--ox-mono)', fontSize: '0.85em' }}
          />
          <select
            value={proposeRel}
            onChange={(e) => setProposeRel(e.target.value as Relation)}
          >
            {RELATIONS.map((rel) => (
              <option key={rel} value={rel}>
                {t(RELATION_LABEL_KEY[rel])}
              </option>
            ))}
          </select>
          <button
            type="button"
            className="btn btn-sm"
            disabled={!targetID || propose.isPending}
            onClick={() => propose.mutate({ tid: targetID, rel: proposeRel })}
          >
            {t('diplomacy.proposeBtn')}
          </button>
        </div>
      )}
    </div>
  );
}
