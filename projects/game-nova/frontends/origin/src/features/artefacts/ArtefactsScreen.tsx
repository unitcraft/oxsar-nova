// S-013 Artefacts (план 72 Ф.4 Spring 3).
//
// Pixel-perfect зеркало legacy `templates/standard/artefacts.tpl`:
// инвентарь артефактов игрока с группировкой по статусу + действия
// activate / deactivate / переход на ArtefactInfo.
//
// Endpoints (openapi.yaml):
//   GET  /api/artefacts                   → { artefacts: Artefact[] }
//   POST /api/artefacts/{id}/activate     → Artefact (Idempotency-Key R9)
//   POST /api/artefacts/{id}/deactivate   → 204     (Idempotency-Key R9)
//
// Sell идёт через legacy artefact-market: `POST /api/artefacts/{id}/sell`
// — реализовано в существующем `api/market.ts` (S-021).
// Переход на P2P-биржу (план 68) — отдельная сессия Spring 5.

import { useState, useMemo } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import {
  activateArtefact,
  deactivateArtefact,
  fetchArtefacts,
} from '@/api/artefacts';
import type { ApiError } from '@/api/client';
import { QK } from '@/api/query-keys';
import type { Artefact, ArtefactState } from '@/api/types';
import { useTranslation } from '@/i18n/i18n';
import { findArtefactCatalog } from '../common/artefact-catalog';

type GroupKey = 'active' | 'held' | 'other';

interface Group {
  key: GroupKey;
  titleKey: string;
  items: Artefact[];
}

function stateLabelKey(state: ArtefactState): string {
  switch (state) {
    case 'held':
      return 'stateHeld';
    case 'active':
      return 'stateActive';
    case 'delayed':
      return 'stateDelayed';
    case 'expired':
      return 'stateExpired';
    case 'consumed':
      return 'stateConsumed';
  }
}

function groupArtefacts(items: Artefact[]): Group[] {
  const active: Artefact[] = [];
  const held: Artefact[] = [];
  const other: Artefact[] = [];
  for (const a of items) {
    if (a.state === 'active' || a.state === 'delayed') {
      active.push(a);
    } else if (a.state === 'held') {
      held.push(a);
    } else {
      other.push(a);
    }
  }
  const groups: Group[] = [
    { key: 'active', titleKey: 'groupActive', items: active },
    { key: 'held', titleKey: 'groupHeld', items: held },
    { key: 'other', titleKey: 'groupOther', items: other },
  ];
  return groups.filter((g) => g.items.length > 0);
}

export function ArtefactsScreen() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  const [errMsg, setErrMsg] = useState<string | null>(null);

  const listQ = useQuery({
    queryKey: QK.artefacts(),
    queryFn: fetchArtefacts,
    refetchInterval: 30_000,
  });

  const activate = useMutation({
    mutationFn: (id: string) => activateArtefact(id),
    onSuccess: () => void qc.invalidateQueries({ queryKey: QK.artefacts() }),
    onError: (e) => setErrMsg((e as ApiError).message),
  });

  const deactivate = useMutation({
    mutationFn: (id: string) => deactivateArtefact(id),
    onSuccess: () => void qc.invalidateQueries({ queryKey: QK.artefacts() }),
    onError: (e) => setErrMsg((e as ApiError).message),
  });

  const items = listQ.data?.artefacts ?? [];
  const groups = useMemo(() => groupArtefacts(items), [items]);

  if (listQ.isLoading) return <div className="idiv">…</div>;

  return (
    <table className="ntable">
      <thead>
        <tr>
          <th colSpan={4}>{t('artefacts', 'title')}</th>
        </tr>
      </thead>
      <tbody>
        {items.length === 0 && (
          <tr>
            <td colSpan={4} className="center">
              {t('artefacts', 'empty')}
            </td>
          </tr>
        )}
        {groups.map((group) => (
          <ArtefactGroupRow
            key={group.key}
            group={group}
            onActivate={(id) => activate.mutate(id)}
            onDeactivate={(id) => deactivate.mutate(id)}
            disabled={activate.isPending || deactivate.isPending}
          />
        ))}
        {errMsg && (
          <tr>
            <td colSpan={4} className="center">
              <span className="false">{errMsg}</span>
            </td>
          </tr>
        )}
      </tbody>
    </table>
  );
}

interface GroupRowProps {
  group: Group;
  onActivate: (id: string) => void;
  onDeactivate: (id: string) => void;
  disabled: boolean;
}

function ArtefactGroupRow({
  group,
  onActivate,
  onDeactivate,
  disabled,
}: GroupRowProps) {
  const { t } = useTranslation();
  return (
    <>
      <tr>
        <th colSpan={4}>{t('artefacts', group.titleKey)}</th>
      </tr>
      {group.items.map((a) => {
        const cat = findArtefactCatalog(a.unit_id);
        const name = cat
          ? t('info', cat.i18nName.split('.')[1] ?? cat.i18nName)
          : `${t('artefacts', 'toastArtefact')} #${a.unit_id}`;
        const desc =
          cat && cat.i18nDesc
            ? t('info', cat.i18nDesc.split('.')[1] ?? cat.i18nDesc)
            : '';
        return (
          <tr key={a.id}>
            <td style={{ width: '1%' }}>
              <Link to={`/artefact/${a.unit_id}`}>#{a.unit_id}</Link>
            </td>
            <td>
              <div style={{ width: '100%' }}>
                <b>
                  <Link to={`/artefact/${a.unit_id}`}>{name}</Link>
                </b>
              </div>
              {desc && (
                <div style={{ clear: 'both', fontSize: 'smaller' }}>{desc}</div>
              )}
              <div style={{ fontSize: 'smaller', margin: 5 }}>
                {t('artefacts', stateLabelKey(a.state))}
                {a.expire_at && (
                  <>
                    {' · '}
                    {t('artefacts', 'expires')}{' '}
                    {new Date(a.expire_at).toLocaleString('ru-RU')}
                  </>
                )}
              </div>
            </td>
            <td className="center" style={{ width: '15%' }}>
              {a.state === 'held' && (
                <input
                  type="button"
                  className="button"
                  value={t('artefacts', 'activate')}
                  disabled={disabled}
                  onClick={() => onActivate(a.id)}
                />
              )}
              {a.state === 'active' && (
                <input
                  type="button"
                  className="button"
                  value={t('artefacts', 'deactivate')}
                  disabled={disabled}
                  onClick={() => onDeactivate(a.id)}
                />
              )}
            </td>
            <td className="center" style={{ width: '10%' }}>
              {a.state === 'held' && (
                <Link to="/market">{t('artefacts', 'sell')}</Link>
              )}
            </td>
          </tr>
        );
      })}
    </>
  );
}
