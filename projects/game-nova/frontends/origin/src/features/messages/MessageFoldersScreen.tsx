// План 72.1.17 — Корневой экран /msg: список папок legacy MSG.
//
// Pixel-perfect зеркало legacy `templates/standard/messages.tpl`:
// для каждой папки строка с image (READ/UNREAD), label-link, total,
// new-count, размер. Origin не считает byte-storage (legacy показывал
// `SUM(LENGTH(message))` — для нашего use-case не критично).
//
// Endpoint: GET /api/messages/folders.

import { Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { fetchMessageFolders } from '@/api/messages';
import { QK } from '@/api/query-keys';
import { useTranslation } from '@/i18n/i18n';

export function MessageFoldersScreen() {
  const { t } = useTranslation();

  const q = useQuery({
    queryKey: QK.messageFolders(),
    queryFn: fetchMessageFolders,
  });

  const folders = q.data?.folders ?? [];

  return (
    <>
      <div className="idiv">
        <Link to="/msg/compose">{t('message', 'createNewMessage')}</Link>
      </div>
      <table className="ntable">
        <thead>
          <tr>
            <th colSpan={4}>{t('message', 'folders') || 'Папки'}</th>
          </tr>
          <tr>
            <th></th>
            <th>{t('message', 'folder')}</th>
            <th className="center">{t('message', 'total') || 'Всего'}</th>
            <th className="center">{t('message', 'unread') || 'Новые'}</th>
          </tr>
        </thead>
        <tbody>
          {q.isLoading ? (
            <tr>
              <td colSpan={4} className="center">
                …
              </td>
            </tr>
          ) : folders.length === 0 ? (
            <tr>
              <td colSpan={4} className="center">
                {t('search', 'notFound')}
              </td>
            </tr>
          ) : (
            folders.map((f) => {
              const isUnread = f.unread > 0;
              const linkPath =
                f.folder_id === 1
                  ? '/msg/inbox'
                  : f.folder_id === 2
                    ? '/msg/sent'
                    : `/msg/folder/${f.folder_id}`;
              const labelText = t('msgFolder', f.label_key);
              return (
                <tr key={f.folder_id} className={isUnread ? 'true' : ''}>
                  <td className="center" style={{ width: 24 }}>
                    {isUnread ? '✉' : '·'}
                  </td>
                  <td>
                    <Link to={linkPath}>{labelText}</Link>
                  </td>
                  <td className="center">{f.total}</td>
                  <td className="center">
                    {isUnread ? <b>{f.unread}</b> : '—'}
                  </td>
                </tr>
              );
            })
          )}
        </tbody>
      </table>
    </>
  );
}
