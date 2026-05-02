// S-008 Alliance overview (план 72 Ф.3 Spring 2 ч.1).
//
// Pixel-perfect зеркало legacy `templates/standard/ally.tpl` —
// landing-страница раздела «Альянсы».
//
// Поведение:
//   - Если игрок в альянсе — редирект на /alliance/me.
//   - Иначе: 2 кнопки [Создать альянс] / [Найти альянс] и блок
//     «Текущие заявки» (legacy applications list).
//
// План 72.1.55 Task B (P72.S2.A 1:1): блок «Текущие заявки» теперь
// функционален — fetch /api/users/me/applications + cancel-кнопка.

import { Link, Navigate } from 'react-router-dom';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  cancelMyApplication,
  fetchMyApplications,
} from '@/api/alliance';
import { ConfirmDialog, useConfirm } from '@/features/common/ConfirmDialog';
import { useTranslation } from '@/i18n/i18n';
import { useMyAlliance } from './common';

export function AllianceOverviewScreen() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  const { confirm, dialogProps } = useConfirm();
  const { data, isLoading } = useMyAlliance();

  // План 72.1.55 Task B: own pending applications.
  const appsQ = useQuery({
    queryKey: ['my-applications'],
    queryFn: fetchMyApplications,
    enabled: !data, // только если не в альянсе
  });
  const cancelMut = useMutation({
    mutationFn: (id: string) => cancelMyApplication(id),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['my-applications'] }),
  });

  if (isLoading) {
    return <div className="idiv">…</div>;
  }
  if (data) {
    return <Navigate to="/alliance/me" replace />;
  }

  const apps = appsQ.data?.applications ?? [];

  return (
    <>
      <table className="ntable">
        <thead>
          <tr>
            <th colSpan={2}>{t('alliance', 'alliances')}</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td className="center">
              <Link to="/alliance/create">{t('alliance', 'foundAlliance')}</Link>
            </td>
            <td className="center">
              <Link to="/alliance/list">{t('alliance', 'joinAlliance')}</Link>
            </td>
          </tr>
        </tbody>
      </table>

      {/* План 72.1.55 Task B (P72.S2.A 1:1): legacy `ally.tpl` блок
          «Текущие заявки» — pending applications applicant'а с
          возможностью отозвать. Не показываем если 0 заявок. */}
      {apps.length > 0 && (
        <table className="ntable">
          <thead>
            <tr>
              <th colSpan={3}>
                {t('alliance', 'myApplications') || 'Текущие заявки'}
              </th>
            </tr>
            <tr>
              <th>{t('alliance', 'name') || 'Альянс'}</th>
              <th>{t('alliance', 'message') || 'Сообщение'}</th>
              <th>{t('alliance', 'remove') || 'Отозвать'}</th>
            </tr>
          </thead>
          <tbody>
            {apps.map((app) => (
              <tr key={app.id}>
                <td>
                  <Link to={`/alliance/${app.alliance_id}`}>
                    [{app.alliance_tag}] {app.alliance_name}
                  </Link>
                </td>
                <td>{app.message || '—'}</td>
                <td className="center">
                  <button
                    type="button"
                    className="button"
                    disabled={cancelMut.isPending}
                    onClick={async () => {
                      if (await confirm({
                        title: t('alliance', 'remove') || 'Отозвать',
                        message:
                          (t('alliance', 'cancelApplicationConfirm') as string) ||
                          'Отозвать заявку?',
                        destructive: true,
                      })) {
                        cancelMut.mutate(app.id);
                      }
                    }}
                  >
                    ✕
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
      <ConfirmDialog {...dialogProps} />
    </>
  );
}
