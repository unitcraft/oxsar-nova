// S-008 Alliance overview (для члена) + S-012 Members (план 72 Ф.3
// Spring 2 ч.1).
//
// Pixel-perfect зеркало `templates/standard/ally.tpl` (содержательный
// блок для членов: tag/name/member-count/founder + список действий) +
// applications.tpl (список заявок для admin'а с can_see_applications).
//
// Поведение:
//   - Загружает /api/alliances/me — alliance + members.
//   - Owner видит блок «Заявки» (если closed) с принять/отклонить.
//   - Не-owner видит [Покинуть альянс].
//   - Owner видит [Распустить] / [Передать лидерство] / [Управление].
//   - Ссылки внизу — на разделы /alliance/members, /alliance/manage,
//     /alliance/diplomacy, /alliance/ranks, /alliance/audit.

import { useState } from 'react';
import { Link, Navigate, useNavigate } from 'react-router-dom';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  approveApplication,
  disbandAlliance,
  fetchAllianceApplications,
  fetchDescriptions,
  leaveAlliance,
  rejectApplication,
} from '@/api/alliance';
import type { ApiError } from '@/api/client';
import { QK } from '@/api/query-keys';
import { useAuthStore } from '@/stores/auth';
import { useTranslation } from '@/i18n/i18n';
import { useMyAlliance } from './common';

export function AllianceMyScreen() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  const navigate = useNavigate();
  const my = useMyAlliance();
  const userId = useAuthStore((s) => s.userId);

  const [errMsg, setErrMsg] = useState<string | null>(null);

  if (my.isLoading) return <div className="idiv">…</div>;
  if (!my.data) return <Navigate to="/alliance" replace />;

  const al = my.data.alliance;
  const members = my.data.members;
  const isOwner = !!userId && userId === al.owner_id;

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
            <td style={{ width: '30%' }}>{t('alliance', 'tag')}</td>
            <td>{al.tag}</td>
          </tr>
          <tr>
            <td>{t('alliance', 'name')}</td>
            <td>{al.name}</td>
          </tr>
          <tr>
            <td>{t('alliance', 'member')}</td>
            <td>
              {al.member_count} ({members.length}){' '}
              <Link to="/alliance/members">{t('alliance', 'memberList')}</Link>
            </td>
          </tr>
          <tr>
            <td>{t('alliance', 'founder')}</td>
            <td>{al.owner_name}</td>
          </tr>
          <ExternalDescriptionRow allianceID={al.id} />
          <tr>
            <td colSpan={2} className="center">
              [{' '}
              <Link to="/alliance/diplomacy">
                {t('alliance', 'diplomacy')}
              </Link>{' '}
              ] [{' '}
              <Link to="/alliance/ranks">
                {t('alliance', 'rightManagement')}
              </Link>{' '}
              ] [{' '}
              <Link to="/alliance/audit">{t('alliance', 'audit.title')}</Link>{' '}
              ]
              {isOwner && (
                <>
                  {' '}
                  [{' '}
                  <Link to="/alliance/manage">
                    {t('alliance', 'allianceManagement')}
                  </Link>{' '}
                  ]
                </>
              )}
            </td>
          </tr>
          {!isOwner && (
            <tr>
              <td colSpan={2} className="center">
                <button
                  type="button"
                  className="button"
                  onClick={() => {
                    if (
                      window.confirm(
                        t('alliance', 'leaveConfirm', { name: al.name }),
                      )
                    ) {
                      leaveAlliance()
                        .then(() => {
                          void qc.invalidateQueries({
                            queryKey: ['alliances'],
                          });
                          navigate('/alliance');
                        })
                        .catch((e: unknown) =>
                          setErrMsg((e as ApiError).message),
                        );
                    }
                  }}
                >
                  {t('alliance', 'leaveBtn')}
                </button>
              </td>
            </tr>
          )}
          {isOwner && (
            <tr>
              <td colSpan={2} className="center">
                <button
                  type="button"
                  className="button"
                  onClick={() => {
                    if (
                      window.confirm(
                        t('alliance', 'disbandConfirm', { name: al.name }),
                      )
                    ) {
                      disbandAlliance(al.id)
                        .then(() => {
                          void qc.invalidateQueries({
                            queryKey: ['alliances'],
                          });
                          navigate('/alliance');
                        })
                        .catch((e: unknown) =>
                          setErrMsg((e as ApiError).message),
                        );
                    }
                  }}
                >
                  {t('alliance', 'abandonAlliance')}
                </button>{' '}
                <Link to="/alliance/transfer">
                  {t('alliance', 'referFounderStatus')}
                </Link>
              </td>
            </tr>
          )}
          {errMsg && (
            <tr>
              <td colSpan={2} className="center">
                <span className="false">{errMsg}</span>
              </td>
            </tr>
          )}
        </tbody>
      </table>

      {isOwner && !al.is_open && <ApplicationsTable allianceID={al.id} />}
    </>
  );
}

function ExternalDescriptionRow({ allianceID }: { allianceID: string }) {
  const descr = useQuery({
    queryKey: QK.allianceDescriptions(allianceID),
    queryFn: () => fetchDescriptions(allianceID),
  });
  const ext = descr.data?.description_external;
  const intern = descr.data?.description_internal;
  const { t } = useTranslation();
  if (!ext && !intern) return null;
  return (
    <>
      {ext && (
        <tr>
          <td colSpan={2} className="center">
            <pre style={{ whiteSpace: 'pre-wrap', margin: 0 }}>{ext}</pre>
          </td>
        </tr>
      )}
      {intern && (
        <>
          <tr>
            <th colSpan={2}>{t('alliance', 'intern')}</th>
          </tr>
          <tr>
            <td colSpan={2} className="center">
              <pre style={{ whiteSpace: 'pre-wrap', margin: 0 }}>{intern}</pre>
            </td>
          </tr>
        </>
      )}
    </>
  );
}

function ApplicationsTable({ allianceID }: { allianceID: string }) {
  const { t } = useTranslation();
  const qc = useQueryClient();
  const apps = useQuery({
    queryKey: QK.allianceApplications(allianceID),
    queryFn: () => fetchAllianceApplications(allianceID),
    refetchInterval: 30_000,
  });

  const approve = useMutation({
    mutationFn: (appID: string) => approveApplication(appID),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['alliances'] }),
  });
  const reject = useMutation({
    mutationFn: (appID: string) => rejectApplication(appID),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['alliances'] }),
  });

  const list = apps.data?.applications ?? [];

  return (
    <table className="ntable">
      <thead>
        <tr>
          <th colSpan={6}>{t('alliance', 'applications')}</th>
        </tr>
        <tr>
          <th>{t('alliance', 'candidate')}</th>
          {/* План 72.1.45 §3: координаты home-планеты + очки. */}
          <th>{t('alliance', 'candidateCoords') || 'Координаты'}</th>
          <th>{t('alliance', 'memberPoints') || 'Очки'}</th>
          <th>{t('alliance', 'applicationContent')}</th>
          <th>{t('alliance', 'applicationTime')}</th>
          <th>{t('alliance', 'operations')}</th>
        </tr>
      </thead>
      <tbody>
        {list.length === 0 && (
          <tr>
            <td colSpan={6} className="center">
              {t('alliance', 'nothing')}
            </td>
          </tr>
        )}
        {list.map((ap) => (
          <tr key={ap.id}>
            <td className="center">{ap.username}</td>
            <td className="center">
              {ap.home_galaxy > 0
                ? `[${ap.home_galaxy}:${ap.home_system}:${ap.home_position}]`
                : '—'}
            </td>
            <td className="center">{ap.points.toLocaleString('ru-RU')}</td>
            <td>{ap.message}</td>
            <td className="center">
              {new Date(ap.created_at).toLocaleString('ru-RU')}
            </td>
            <td className="center">
              <button
                type="button"
                className="button"
                disabled={approve.isPending}
                onClick={() => approve.mutate(ap.id)}
              >
                {t('alliance', 'receipt')}
              </button>{' '}
              <button
                type="button"
                className="button"
                disabled={reject.isPending}
                onClick={() => reject.mutate(ap.id)}
              >
                {t('alliance', 'refuse')}
              </button>
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}
