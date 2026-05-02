// S-013 Alliance ranks (план 72 Ф.3 Spring 2 ч.1).
//
// Pixel-perfect зеркало legacy `templates/standard/manage_ranks.tpl` —
// таблица со столбцами CAN_MANAGE / CAN_SEE_APPLICATIONS / ... и формой
// «создать новый ранг». В origin расширено гранулярными permissions
// плана 67 D-014 / U-005 (7 ключей вместо legacy-набора).
//
// Endpoint: GET /api/alliances/{id}/ranks (требует can_manage_ranks или owner).

import { useEffect, useState } from 'react';
import { Navigate, Link } from 'react-router-dom';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  createRank,
  deleteRank,
  fetchRanks,
  updateRank,
} from '@/api/alliance';
import type { ApiError } from '@/api/client';
import { QK } from '@/api/query-keys';
import type {
  AlliancePermissionKey,
  AlliancePermissionMap,
  AllianceRank,
} from '@/api/types';
import { useAuthStore } from '@/stores/auth';
import { useTranslation } from '@/i18n/i18n';
import { PERMISSION_KEYS, useMyAlliance } from './common';

export function AllianceRanksScreen() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  const my = useMyAlliance();
  const userId = useAuthStore((s) => s.userId);
  const [errMsg, setErrMsg] = useState<string | null>(null);
  const [newName, setNewName] = useState('');

  const ranksQ = useQuery({
    queryKey: QK.allianceRanks(my.data?.alliance.id ?? ''),
    queryFn: () => fetchRanks(my.data?.alliance.id ?? ''),
    enabled: !!my.data,
  });

  const create = useMutation({
    mutationFn: (name: string) =>
      createRank(my.data?.alliance.id ?? '', { name }),
    onSuccess: () => {
      setNewName('');
      void qc.invalidateQueries({
        queryKey: QK.allianceRanks(my.data?.alliance.id ?? ''),
      });
    },
    onError: (e) => setErrMsg((e as ApiError).message),
  });

  const remove = useMutation({
    mutationFn: (rankID: string) =>
      deleteRank(my.data?.alliance.id ?? '', rankID),
    onSuccess: () =>
      void qc.invalidateQueries({
        queryKey: QK.allianceRanks(my.data?.alliance.id ?? ''),
      }),
    onError: (e) => setErrMsg((e as ApiError).message),
  });

  if (my.isLoading) return <div className="idiv">…</div>;
  if (!my.data) return <Navigate to="/alliance" replace />;

  const al = my.data.alliance;
  const isOwner = !!userId && userId === al.owner_id;

  if (!isOwner) {
    return (
      <div className="idiv">
        <span className="false">{t('alliance', 'rightManagement')}</span>
      </div>
    );
  }

  const ranks = ranksQ.data?.ranks ?? [];

  return (
    <>
      <div className="idiv">
        <Link to="/alliance/me">← {al.tag}</Link>
      </div>

      <table className="ntable">
        <thead>
          <tr>
            <th colSpan={PERMISSION_KEYS.length + 2}>
              {t('alliance', 'setRankRights')}
            </th>
          </tr>
          <tr>
            <td>{t('alliance', 'rankName')}</td>
            {PERMISSION_KEYS.map((p) => (
              <td className="center" key={p} title={t('alliance', `perm.${p}`)}>
                {t('alliance', `perm.${p}.short`)}
              </td>
            ))}
            <td />
          </tr>
        </thead>
        <tbody>
          {ranks.length === 0 && (
            <tr>
              <td className="center" colSpan={PERMISSION_KEYS.length + 2}>
                {t('alliance', 'ranks.empty')}
              </td>
            </tr>
          )}
          {ranks.map((r) => (
            <RankRow
              key={r.id}
              allianceID={al.id}
              rank={r}
              onDelete={() => {
                if (
                  window.confirm(
                    t('alliance', 'ranks.deleteConfirm', { name: r.name }),
                  )
                ) {
                  remove.mutate(r.id);
                }
              }}
            />
          ))}
        </tbody>
      </table>

      <table className="ntable">
        <thead>
          <tr>
            <th>{t('alliance', 'createNewRank')}</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>
              <label htmlFor="name">{t('alliance', 'rankName')}:</label>{' '}
              <input
                type="text"
                name="name"
                id="name"
                maxLength={32}
                value={newName}
                onChange={(e) => setNewName(e.target.value)}
              />{' '}
              <input
                type="button"
                className="button"
                value={t('alliance', 'ranks.createBtn')}
                disabled={newName.length === 0 || create.isPending}
                onClick={() => create.mutate(newName)}
              />
            </td>
          </tr>
        </tbody>
      </table>

      {errMsg && (
        <div className="idiv">
          <span className="false">{errMsg}</span>
        </div>
      )}
    </>
  );
}

function RankRow({
  allianceID,
  rank,
  onDelete,
}: {
  allianceID: string;
  rank: AllianceRank;
  onDelete: () => void;
}) {
  const qc = useQueryClient();
  const { t } = useTranslation();
  const [perms, setPerms] = useState<AlliancePermissionMap>(
    rank.permissions ?? {},
  );
  const [name, setName] = useState(rank.name);
  const [dirty, setDirty] = useState(false);

  // Если backend вернёт обновлённый rank — локальный state ресинкаем.
  // Не убирает локальные несохранённые правки во время мутации, потому
  // что dirty=true.
  useEffect(() => {
    if (!dirty) {
      setPerms(rank.permissions ?? {});
      setName(rank.name);
    }
  }, [rank.permissions, rank.name, dirty]);

  const save = useMutation({
    mutationFn: () => updateRank(allianceID, rank.id, { name, permissions: perms }),
    onSuccess: () => {
      setDirty(false);
      void qc.invalidateQueries({ queryKey: QK.allianceRanks(allianceID) });
    },
  });

  function toggle(p: AlliancePermissionKey) {
    setDirty(true);
    setPerms((prev) => ({ ...prev, [p]: !prev[p] }));
  }

  return (
    <tr>
      <td>
        <input
          type="text"
          value={name}
          maxLength={32}
          onChange={(e) => {
            setDirty(true);
            setName(e.target.value);
          }}
        />
      </td>
      {PERMISSION_KEYS.map((p) => (
        <td key={p} className="center">
          <input
            type="checkbox"
            checked={!!perms[p]}
            onChange={() => toggle(p)}
            aria-label={t('alliance', `perm.${p}`)}
          />
        </td>
      ))}
      <td className="center">
        {/* План 72.1.52 (72.1.6 P3 closure): client-side validation
            длины имени ранга. Backend-сентинель `ErrRankNameInvalid`
            разрешает 1-32 символа, frontend дублирует чтобы не
            показывать пустую кнопку «Сохранить». */}
        <input
          type="button"
          className="button"
          value={t('alliance', 'ranks.saveBtn')}
          disabled={
            !dirty ||
            save.isPending ||
            name.trim().length < 1 ||
            name.trim().length > 32
          }
          onClick={() => save.mutate()}
        />{' '}
        <input
          type="button"
          className="button"
          value="✕"
          title={t('alliance', 'remove')}
          onClick={onDelete}
        />
      </td>
    </tr>
  );
}
