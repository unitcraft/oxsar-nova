// S-036 Chat (universal) и S-037 ChatAlly — общий и альянс-чат
// (план 72 Ф.5 Spring 4).
//
// Pixel-perfect зеркало legacy `templates/standard/chat.tpl` и
// `chatally.tpl` (упрощённо: без BBcode-toolbar — план 72 P72.S4.BBCODE).
// Origin-фронт реализует чат через WebSocket /api/chat/{kind}/ws с
// REST-fallback для send (POST /api/chat/{kind}/send + Idempotency-Key).
//
// Backend: internal/chat/handler.go — план 32 Ф.5 (Redis pub/sub),
// план 46 Ф.4 (UGC-blacklist + rate-limit 10/min), план 69 D-020
// (read-marker per kind).
//
// trade-off P72.S4.BBCODE (см. simplifications.md): BBCode из legacy
// (`[b]`, `[url]…[/url]`) рендерится как plain text. TipTap в Ф.8.

import { useEffect, useRef, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  chatWsUrl,
  deleteChatMessage,
  editChatMessage,
  fetchChatHistory,
  markChatRead,
  sendChatMessage,
} from '@/api/chat';
import { QK } from '@/api/query-keys';
import type { ApiError } from '@/api/client';
import type { ChatChannelKind, ChatMessage } from '@/api/types';
import { useAuthStore } from '@/stores/auth';
import { useTranslation } from '@/i18n/i18n';

const EDIT_WINDOW_MS = 5 * 60 * 1000;

interface ChatProps {
  kind: ChatChannelKind;
}

export function ChatScreen({ kind }: ChatProps) {
  const { t } = useTranslation();
  const qc = useQueryClient();
  const token = useAuthStore((s) => s.accessToken);
  const myId = useAuthStore((s) => s.userId);

  const historyQ = useQuery({
    queryKey: QK.chatHistory(kind),
    queryFn: () => fetchChatHistory(kind),
  });

  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [input, setInput] = useState('');
  const [wsOk, setWsOk] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editBody, setEditBody] = useState('');
  const [errMsg, setErrMsg] = useState<string | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const scrollRef = useRef<HTMLDivElement | null>(null);

  // sync messages при загрузке истории и смене kind.
  useEffect(() => {
    if (historyQ.data) setMessages(historyQ.data);
  }, [historyQ.data, kind]);

  // mark-read при заходе на экран.
  useEffect(() => {
    markChatRead(kind).catch(() => null);
    void qc.invalidateQueries({ queryKey: QK.chatUnread(kind) });
  }, [kind, qc]);

  // WebSocket connect.
  useEffect(() => {
    if (!token) return;
    let stopped = false;
    let retry: ReturnType<typeof setTimeout> | null = null;

    function connect() {
      if (stopped) return;
      const ws = new WebSocket(chatWsUrl(kind, token!));
      wsRef.current = ws;
      ws.onopen = () => setWsOk(true);
      ws.onclose = () => {
        setWsOk(false);
        retry = setTimeout(connect, 2000);
      };
      ws.onerror = () => {
        setWsOk(false);
      };
      ws.onmessage = (ev) => {
        try {
          const msg = JSON.parse(ev.data as string) as ChatMessage;
          if (msg.kind === 'delete') {
            setMessages((prev) => prev.filter((m) => m.id !== msg.id));
            return;
          }
          if (msg.kind === 'edit') {
            setMessages((prev) =>
              prev.map((m) =>
                m.id === msg.id
                  ? {
                      ...m,
                      body: msg.body,
                      ...(msg.edited_at !== undefined
                        ? { edited_at: msg.edited_at }
                        : {}),
                    }
                  : m,
              ),
            );
            return;
          }
          // обычное сообщение — добавляем (если ещё нет такого id).
          setMessages((prev) =>
            prev.some((m) => m.id === msg.id) ? prev : [...prev, msg],
          );
        } catch {
          /* ignore */
        }
      };
    }

    connect();

    return () => {
      stopped = true;
      if (retry) clearTimeout(retry);
      wsRef.current?.close();
      wsRef.current = null;
    };
  }, [kind, token]);

  // auto-scroll вниз при новом сообщении.
  useEffect(() => {
    const el = scrollRef.current;
    if (el) el.scrollTop = el.scrollHeight;
  }, [messages.length]);

  const sendMut = useMutation({
    mutationFn: (body: string) => sendChatMessage(kind, body),
    onError: (e) => setErrMsg((e as ApiError).message),
  });

  const editMut = useMutation({
    mutationFn: ({ id, body }: { id: string; body: string }) =>
      editChatMessage(id, body),
    onSuccess: () => {
      setEditingId(null);
      setEditBody('');
    },
    onError: (e) => setErrMsg((e as ApiError).message),
  });

  const deleteMut = useMutation({
    mutationFn: deleteChatMessage,
    onError: (e) => setErrMsg((e as ApiError).message),
  });

  function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    const body = input.trim();
    if (!body) return;
    setErrMsg(null);
    sendMut.mutate(body);
    setInput('');
  }

  return (
    <table className="ntable">
      <thead>
        <tr>
          <th colSpan={2}>
            {kind === 'global'
              ? t('chat', 'tabGlobal')
              : t('chat', 'tabAlliance')}
            {!wsOk && (
              <span className="small" style={{ float: 'right' }}>
                {' '}
                {t('chat', 'reconnecting')}
              </span>
            )}
          </th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td colSpan={2}>
            <div
              ref={scrollRef}
              style={{
                height: 360,
                overflowY: 'auto',
                border: '1px solid #555',
                padding: 4,
                fontFamily: 'monospace',
                fontSize: 13,
              }}
            >
              {messages.length === 0 ? (
                <div className="center small">—</div>
              ) : (
                messages.map((m) => {
                  const isMine = m.author_id === myId;
                  const ts = new Date(m.created_at);
                  const tsStr = ts.toLocaleTimeString('ru-RU', {
                    hour: '2-digit',
                    minute: '2-digit',
                  });
                  const canEdit =
                    isMine &&
                    Date.now() - ts.getTime() < EDIT_WINDOW_MS;
                  return (
                    <div key={m.id}>
                      <span className="small">[{tsStr}]</span>{' '}
                      <b>{m.author_name || '—'}:</b>{' '}
                      {editingId === m.id ? (
                        <>
                          <input
                            type="text"
                            value={editBody}
                            maxLength={500}
                            onChange={(e) => setEditBody(e.target.value)}
                            style={{ width: '60%' }}
                          />{' '}
                          <button
                            type="button"
                            className="button"
                            disabled={editMut.isPending || !editBody.trim()}
                            onClick={() =>
                              editMut.mutate({ id: m.id, body: editBody })
                            }
                          >
                            ✓
                          </button>{' '}
                          <button
                            type="button"
                            className="button"
                            onClick={() => {
                              setEditingId(null);
                              setEditBody('');
                            }}
                          >
                            ✕
                          </button>
                        </>
                      ) : (
                        <>
                          <span style={{ whiteSpace: 'pre-wrap' }}>
                            {m.body}
                          </span>
                          {m.edited_at && (
                            <span className="small">
                              {' '}
                              ({t('chat', 'editedMark')})
                            </span>
                          )}
                          {canEdit && (
                            <span className="small">
                              {' · '}
                              <a
                                href="#"
                                title={t('chat', 'editTitle')}
                                onClick={(e) => {
                                  e.preventDefault();
                                  setEditingId(m.id);
                                  setEditBody(m.body);
                                }}
                              >
                                ✎
                              </a>
                              {' · '}
                              <a
                                href="#"
                                title={t('chat', 'deleteTitle')}
                                onClick={(e) => {
                                  e.preventDefault();
                                  if (
                                    window.confirm(
                                      t('chat', 'deleteConfirmMsg'),
                                    )
                                  ) {
                                    deleteMut.mutate(m.id);
                                  }
                                }}
                              >
                                ✕
                              </a>
                            </span>
                          )}
                        </>
                      )}
                    </div>
                  );
                })
              )}
            </div>
          </td>
        </tr>
        <tr>
          <td colSpan={2}>
            <form onSubmit={onSubmit}>
              <input
                type="text"
                value={input}
                maxLength={500}
                placeholder={t('chat', 'inputPlaceholder')}
                onChange={(e) => setInput(e.target.value)}
                style={{ width: '70%' }}
              />{' '}
              <button
                type="submit"
                className="button"
                disabled={sendMut.isPending || !input.trim()}
              >
                {t('chat', 'sendBtn')}
              </button>
            </form>
            {errMsg && <span className="false">{errMsg}</span>}
          </td>
        </tr>
      </tbody>
    </table>
  );
}

export function ChatGlobalScreen() {
  return <ChatScreen kind="global" />;
}

export function ChatAllyScreen() {
  return <ChatScreen kind="alliance" />;
}
