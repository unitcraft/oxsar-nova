// S-035 MSG — личные сообщения (план 72 Ф.5 Spring 4,
// расширен 72.1.17: folder routing).
//
// Pixel-perfect зеркало legacy `templates/standard/folder.tpl`.
//
// Поддерживаемые URL:
//   /msg/inbox            — folder=1 (INBOX, легаси)
//   /msg/sent             — folder=2 (SENT, легаси)
//   /msg/folder/<N>       — произвольная папка (план 72.1.17)
//
// Endpoints:
//   GET /api/messages?folder=N   — содержимое папки
//   GET /api/messages/sent       — отправленные (legacy backwards-compat)
//   DELETE /api/messages/{id}    — удалить одно
//   POST /api/messages/{id}/read — пометить прочитанным
//   POST /api/messages           — отправить (compose)

import { useState } from 'react';
import { Link, useParams, useSearchParams } from 'react-router-dom';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  deleteAllMessages,
  deleteMessage,
  fetchMessages,
  markMessageRead,
  sendMessage,
} from '@/api/messages';
import type { ApiError } from '@/api/client';
import type { MessageFolder } from '@/api/types';
import { QK } from '@/api/query-keys';
import { useTranslation } from '@/i18n/i18n';

const MAX_PM_LENGTH = 2000;

export function MessagesScreen() {
  const { t } = useTranslation();
  const params = useParams<{ folder?: string; folderId?: string }>();

  // Резолвим источник: legacy 'inbox'/'sent' или числовой folderId.
  // folderId перекрывает folder, если задан.
  const numericFolderId =
    params.folderId !== undefined ? Number(params.folderId) : null;
  const legacyFolder: MessageFolder =
    params.folder === 'sent' ? 'sent' : 'inbox';
  const folderKey: MessageFolder | number =
    numericFolderId !== null && Number.isFinite(numericFolderId)
      ? numericFolderId
      : legacyFolder;
  // 'inbox' для UI-логики (delete-all, expand-mark-read) равно folder=1
  // и любой другой numericFolderId, который не 2 (sent). 'sent' — это
  // legacy=2, мы не помечаем читать и не показываем «удалить все».
  const isSent = folderKey === 'sent' || folderKey === 2;
  const folder: MessageFolder = isSent ? 'sent' : 'inbox';

  const qc = useQueryClient();
  const q = useQuery({
    queryKey: QK.messages(folderKey),
    queryFn: () => fetchMessages(folderKey),
  });

  const [expandedId, setExpandedId] = useState<string | null>(null);
  const [errMsg, setErrMsg] = useState<string | null>(null);

  const readMut = useMutation({
    mutationFn: markMessageRead,
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['messages'] }),
  });

  const deleteMut = useMutation({
    mutationFn: deleteMessage,
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['messages'] }),
    onError: (e) => setErrMsg((e as ApiError).message),
  });

  // План 72.1.11: bulk-delete (legacy MSG.class.php::DeleteAll).
  // Backend DELETE /api/messages удаляет все сообщения юзера с
  // to_user_id=uid (folder=0 в DeleteAll service). Для папки 'sent'
  // backend семантику не поддерживает — кнопка показывается только
  // для inbox.
  const deleteAllMut = useMutation({
    mutationFn: deleteAllMessages,
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['messages'] }),
    onError: (e) => setErrMsg((e as ApiError).message),
  });

  const list = q.data?.messages ?? [];

  function expand(id: string, isRead: boolean) {
    setExpandedId((prev) => (prev === id ? null : id));
    if (folder === 'inbox' && !isRead) {
      readMut.mutate(id);
    }
  }

  return (
    <>
      <div className="idiv">
        {/* План 72.1.17: ссылка «← к папкам» (legacy index с списком). */}
        <Link to="/msg">
          <b>← {t('message', 'folder')}</b>
        </Link>{' '}
        |{' '}
        <Link to="/msg/inbox" className={folder === 'inbox' && numericFolderId === null ? 'true' : ''}>
          {t('overview', 'unreadPlural') || 'Входящие'}
        </Link>{' '}
        |{' '}
        <Link to="/msg/sent" className={folder === 'sent' ? 'true' : ''}>
          {t('message', 'outbox')}
        </Link>{' '}
        |{' '}
        <Link to="/msg/compose">{t('message', 'createNewMessage')}</Link>
        {folder === 'inbox' && list.length > 0 && (
          <>
            {' | '}
            <button
              type="button"
              className="button"
              disabled={deleteAllMut.isPending}
              onClick={() => {
                if (window.confirm(t('message', 'deleteAllConfirm'))) {
                  deleteAllMut.mutate();
                }
              }}
            >
              {t('message', 'deleteAllBtn')}
            </button>
          </>
        )}
      </div>

      <table className="ntable">
        <thead>
          <tr>
            <th>
              {folder === 'sent' ? t('message', 'receiver') : t('message', 'from')}
            </th>
            <th>{t('message', 'subject')}</th>
            <th>{t('message', 'date')}</th>
            <th>{t('message', 'action')}</th>
          </tr>
        </thead>
        <tbody>
          {q.isLoading ? (
            <tr>
              <td colSpan={4} className="center">
                …
              </td>
            </tr>
          ) : list.length === 0 ? (
            <tr>
              <td colSpan={4} className="center">
                {t('search', 'notFound')}
              </td>
            </tr>
          ) : (
            list.map((m) => {
              const isRead = m.read_at !== undefined;
              const isOpen = expandedId === m.id;
              return (
                <Row
                  key={m.id}
                  m={m}
                  isOpen={isOpen}
                  isRead={isRead}
                  isSent={folder === 'sent'}
                  onExpand={() => expand(m.id, isRead)}
                  onDelete={() => {
                    if (window.confirm(t('message', 'deleteOneConfirm'))) {
                      deleteMut.mutate(m.id);
                    }
                  }}
                  deletePending={deleteMut.isPending}
                />
              );
            })
          )}
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

function Row({
  m,
  isOpen,
  isRead,
  isSent,
  onExpand,
  onDelete,
  deletePending,
}: {
  m: import('@/api/types').Message;
  isOpen: boolean;
  isRead: boolean;
  isSent: boolean;
  onExpand: () => void;
  onDelete: () => void;
  deletePending: boolean;
}) {
  const dateStr = new Date(m.created_at).toLocaleString('ru-RU');
  const subject = m.subject || `(${'no subject'})`;
  const className = !isRead && !isSent ? 'true' : '';
  return (
    <>
      <tr className={className}>
        <td>{m.from_username || '—'}</td>
        <td>
          <a
            href="#"
            onClick={(e) => {
              e.preventDefault();
              onExpand();
            }}
          >
            {subject}
          </a>
        </td>
        <td>{dateStr}</td>
        <td className="center">
          <button
            type="button"
            className="button"
            disabled={deletePending}
            onClick={onDelete}
          >
            ✕
          </button>
        </td>
      </tr>
      {isOpen && (
        <tr>
          <td colSpan={4} style={{ whiteSpace: 'pre-wrap' }}>
            {m.body}
          </td>
        </tr>
      )}
    </>
  );
}

// Compose-экран — отдельный route /msg/compose.
export function MessageComposeScreen() {
  const { t } = useTranslation();
  const [params] = useSearchParams();
  const initialTo = params.get('to') ?? '';

  const qc = useQueryClient();
  const [to, setTo] = useState(initialTo);
  const [subject, setSubject] = useState('');
  const [body, setBody] = useState('');
  const [okMsg, setOkMsg] = useState<string | null>(null);
  const [errMsg, setErrMsg] = useState<string | null>(null);

  const sendMut = useMutation({
    mutationFn: sendMessage,
    onSuccess: () => {
      setOkMsg(t('message', 'messagesReported') || '✓');
      setTo('');
      setSubject('');
      setBody('');
      void qc.invalidateQueries({ queryKey: ['messages'] });
    },
    onError: (e) => setErrMsg((e as ApiError).message),
  });

  function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setOkMsg(null);
    setErrMsg(null);
    if (!to.trim() || !subject.trim()) {
      setErrMsg(t('search', 'promptEmpty') || 'fill required fields');
      return;
    }
    sendMut.mutate({ to: to.trim(), subject: subject.trim(), body });
  }

  return (
    <>
      <div className="idiv">
        <Link to="/msg/inbox">← {t('message', 'folder')}</Link>
      </div>
      <form onSubmit={onSubmit}>
        <table className="ntable">
          <thead>
            <tr>
              <th colSpan={2}>{t('message', 'newMessage')}</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td>
                <label htmlFor="receiver">{t('message', 'receiver')}</label>
              </td>
              <td>
                <input
                  id="receiver"
                  type="text"
                  name="receiver"
                  maxLength={64}
                  value={to}
                  onChange={(e) => setTo(e.target.value)}
                />
              </td>
            </tr>
            <tr>
              <td>
                <label htmlFor="subject">{t('message', 'subject')}</label>
              </td>
              <td>
                <input
                  id="subject"
                  type="text"
                  name="subject"
                  maxLength={50}
                  value={subject}
                  onChange={(e) => setSubject(e.target.value)}
                />
              </td>
            </tr>
            <tr>
              <td>
                <label htmlFor="body">{t('message', 'message')}</label>
              </td>
              <td>
                <textarea
                  id="body"
                  name="body"
                  cols={35}
                  rows={8}
                  maxLength={MAX_PM_LENGTH}
                  value={body}
                  onChange={(e) => setBody(e.target.value)}
                />
                <br />
                <span className="small">
                  {body.length} / {MAX_PM_LENGTH}
                </span>
              </td>
            </tr>
          </tbody>
          <tfoot>
            <tr>
              <td colSpan={2} className="center">
                <input
                  type="submit"
                  className="button"
                  value={t('message', 'newMessage')}
                  disabled={sendMut.isPending}
                />{' '}
                {okMsg && <span className="true">{okMsg}</span>}
                {errMsg && <span className="false">{errMsg}</span>}
              </td>
            </tr>
          </tfoot>
        </table>
      </form>
    </>
  );
}
