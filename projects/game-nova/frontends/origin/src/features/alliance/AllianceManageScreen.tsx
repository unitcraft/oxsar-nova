// S-018 Alliance management / settings (план 72 Ф.3 Spring 2 ч.1).
//
// Pixel-perfect зеркало legacy `templates/standard/manage_ally.tpl`
// (часть с tag/name + memberlist-prefs + open/closed toggle, без блока
// 3 textarea — он вынесен на /alliance/descriptions).
//
// В origin-фронте сохраняются только настройки, которые есть в backend:
//   - is_open (PATCH /api/alliances/{id}/open)
//   - распуск (на странице /alliance/me, кнопка abandonAlliance)
//   - передача лидерства (на странице /alliance/transfer)
//
// Поля legacy `name`/`tag` PATCH backend пока не поддерживает (см.
// openapi.yaml — для /api/alliances/{id} есть только GET/DELETE), поэтому
// они read-only и помечены simplifications.md (P72.S2.B).

import { Link, Navigate } from 'react-router-dom';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { setAllianceOpen } from '@/api/alliance';
import type { ApiError } from '@/api/client';
import { useAuthStore } from '@/stores/auth';
import { useTranslation } from '@/i18n/i18n';
import { useState } from 'react';
import { useMyAlliance } from './common';

export function AllianceManageScreen() {
  const { t } = useTranslation();
  const my = useMyAlliance();
  const qc = useQueryClient();
  const userId = useAuthStore((s) => s.userId);
  const [errMsg, setErrMsg] = useState<string | null>(null);

  const allianceID = my.data?.alliance.id ?? '';

  const setOpen = useMutation({
    mutationFn: (isOpen: boolean) => setAllianceOpen(allianceID, isOpen),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['alliances'] }),
    onError: (e) => setErrMsg((e as ApiError).message),
  });

  if (my.isLoading) return <div className="idiv">…</div>;
  if (!my.data) return <Navigate to="/alliance" replace />;

  const al = my.data.alliance;
  const isOwner = !!userId && userId === al.owner_id;
  if (!isOwner) {
    return (
      <div className="idiv">
        <span className="false">{t('alliance', 'allianceManagement')}</span>
      </div>
    );
  }

  return (
    <>
      <div className="idiv">
        <Link to="/alliance/me">← {al.tag}</Link>
      </div>

      <table className="ntable">
        <thead>
          <tr>
            <th colSpan={2}>{t('alliance', 'allianceManagement')}</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>
              <label>{t('alliance', 'allianceTag')}</label>
            </td>
            <td>{al.tag}</td>
          </tr>
          <tr>
            <td>
              <label>{t('alliance', 'allianceName')}</label>
            </td>
            <td>{al.name}</td>
          </tr>
          <tr>
            <td colSpan={2} className="center">
              [{' '}
              <Link to="/alliance/descriptions">
                {t('alliance', 'externAllianceText')}
              </Link>{' '}
              ] [{' '}
              <Link to="/alliance/diplomacy">{t('alliance', 'diplomacy')}</Link>{' '}
              ] [{' '}
              <Link to="/alliance/ranks">
                {t('alliance', 'rightManagement')}
              </Link>{' '}
              ]
            </td>
          </tr>
        </tbody>
      </table>

      <table className="ntable">
        <thead>
          <tr>
            <th colSpan={2}>{t('alliance', 'memberList')}</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>
              <label htmlFor="apps">
                {t('alliance', 'enableApplications')}
              </label>
            </td>
            <td>
              <input
                type="checkbox"
                id="apps"
                checked={al.is_open}
                disabled={setOpen.isPending}
                onChange={(e) => setOpen.mutate(e.target.checked)}
              />{' '}
              {al.is_open
                ? t('alliance', 'labelOpen')
                : t('alliance', 'labelClosed')}
            </td>
          </tr>
          {errMsg && (
            <tr>
              <td colSpan={2} className="center">
                <span className="false">{errMsg}</span>
              </td>
            </tr>
          )}
        </tbody>
      </table>

      <table className="ntable">
        <thead>
          <tr>
            <th>{t('alliance', 'referFounderStatus')}</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td className="center">
              <Link to="/alliance/transfer">
                {t('alliance', 'referFounderStatus')}
              </Link>
            </td>
          </tr>
        </tbody>
      </table>
    </>
  );
}
