import { useEffect, useRef, useState } from 'react';
import { api } from '@/api/client';
import { useAuthStore } from '@/stores/auth';
import { Confirm } from '@/ui/Confirm';
import { useTranslation } from '@/i18n/i18n';

interface ChatMessage {
  id: string;
  channel: string;
  author_id: string;
  author_name: string;
  body: string;
  created_at: string;
  edited_at?: string;
  kind?: string; // "msg" | "edit" | "delete"
}

type ChannelKind = 'global' | 'alliance';

const EMOJIS = ['😀','😂','😍','🤔','👍','👎','❤️','🔥','🎉','😎','😢','🤣','😡','🙏','💀','🚀','⚔️','🛡️','🌟','💰'];
const EDIT_WINDOW_MS = 5 * 60 * 1000; // 5 минут

export function ChatScreen() {
  const { t } = useTranslation('chat');
  const [kind, setKind] = useState<ChannelKind>('global');
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [input, setInput] = useState('');
  const [wsError, setWsError] = useState('');
  const [showEmoji, setShowEmoji] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editBody, setEditBody] = useState('');
  const [confirmDeleteId, setConfirmDeleteId] = useState<string | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const scrollBoxRef = useRef<HTMLDivElement | null>(null);
  const prevLenRef = useRef(0);
  const inputRef = useRef<HTMLInputElement | null>(null);

  const token = useAuthStore((s) => s.accessToken);
  const myId = useAuthStore((s) => s.userId);

  useEffect(() => {
    api.get<ChatMessage[] | null>(`/api/chat/${kind}/history`)
      .then((list) => setMessages(list ?? []))
      .catch(() => null);
  }, [kind]);

  useEffect(() => {
    if (!token) return;
    setWsError('');

    let stopped = false;
    let retryTimer: ReturnType<typeof setTimeout> | null = null;

    function connect() {
      if (stopped) return;
      const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const host = window.location.host;
      const ws = new WebSocket(`${proto}//${host}/api/chat/${kind}/ws?token=${encodeURIComponent(token!)}`);
      wsRef.current = ws;

      ws.onmessage = (ev: MessageEvent) => {
        try {
          const msg = JSON.parse(ev.data as string) as ChatMessage;

          if (msg.kind === 'delete') {
            setMessages((prev) => prev.filter((m) => m.id !== msg.id));
            return;
          }
          if (msg.kind === 'edit') {
            setMessages((prev) => prev.map((m) => m.id === msg.id ? { ...m, body: msg.body, ...(msg.edited_at !== undefined ? { edited_at: msg.edited_at } : {}) } : m));
            return;
          }
          setMessages((prev) => {
            const tmpIdx = prev.findIndex(
              (m) => m.id.startsWith('tmp-') && m.author_id === msg.author_id && m.body === msg.body,
            );
            if (tmpIdx !== -1) {
              const next = [...prev];
              next[tmpIdx] = msg;
              return next;
            }
            return [...prev, msg];
          });
        } catch {
          // ignore malformed frames
        }
      };
      ws.onerror = () => { /* закрытие придёт в onclose */ };
      ws.onclose = (ev) => {
        if (stopped) return;
        if (ev.wasClean) {
          setWsError('');
        } else {
          setWsError(t('reconnecting'));
          retryTimer = setTimeout(() => {
            setWsError('');
            connect();
          }, 3000);
        }
      };
    }

    connect();

    return () => {
      stopped = true;
      if (retryTimer !== null) clearTimeout(retryTimer);
      wsRef.current?.close();
      wsRef.current = null;
    };
  }, [kind, token, t]);

  useEffect(() => {
    if (messages.length > prevLenRef.current) {
      const box = scrollBoxRef.current;
      if (box) box.scrollTop = box.scrollHeight;
    }
    prevLenRef.current = messages.length;
  }, [messages]);

  function send() {
    const body = input.trim();
    if (!body) return;
    const optimistic: ChatMessage = {
      id: `tmp-${Date.now()}`,
      channel: kind,
      author_id: myId ?? '',
      author_name: '…',
      body,
      created_at: new Date().toISOString(),
      kind: 'msg',
    };
    setMessages((prev) => [...prev, optimistic]);
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ body }));
    } else {
      api.post(`/api/chat/${kind}/send`, { body }).catch(() => null);
    }
    setInput('');
    setShowEmoji(false);
  }

  function handleKey(e: React.KeyboardEvent<HTMLInputElement>) {
    if (e.key === 'Enter') send();
    if (e.key === 'Escape') setShowEmoji(false);
  }

  function insertEmoji(emoji: string) {
    setInput((prev) => prev + emoji);
    inputRef.current?.focus();
  }

  function startEdit(m: ChatMessage) {
    setEditingId(m.id);
    setEditBody(m.body);
  }

  function cancelEdit() {
    setEditingId(null);
    setEditBody('');
  }

  function submitEdit(id: string) {
    const body = editBody.trim();
    if (!body) return;
    api.patch<ChatMessage>(`/api/chat/messages/${id}`, { body })
      .then((updated) => {
        setMessages((prev) => prev.map((m) => m.id === id ? { ...m, body: updated.body, ...(updated.edited_at !== undefined ? { edited_at: updated.edited_at } : {}) } : m));
      })
      .catch(() => null);
    cancelEdit();
  }

  function deleteMsg(id: string) {
    setConfirmDeleteId(id);
  }

  function confirmDelete() {
    if (!confirmDeleteId) return;
    const id = confirmDeleteId;
    setConfirmDeleteId(null);
    api.delete(`/api/chat/messages/${id}`)
      .then(() => setMessages((prev) => prev.filter((m) => m.id !== id)))
      .catch(() => null);
  }

  function canModify(m: ChatMessage): boolean {
    if (m.id.startsWith('tmp-')) return false;
    if (m.author_id !== myId) return false;
    return Date.now() - new Date(m.created_at).getTime() < EDIT_WINDOW_MS;
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '72vh', minHeight: 400 }}>
      <div className="ox-tabs" style={{ marginBottom: 8 }}>
        <button type="button" aria-pressed={kind === 'global'} onClick={() => setKind('global')}>
          {t('tabGlobal')}
        </button>
        <button type="button" aria-pressed={kind === 'alliance'} onClick={() => setKind('alliance')}>
          {t('tabAlliance')}
        </button>
      </div>

      {wsError && (
        <div style={{ fontSize: 14, color: 'var(--ox-warning)', marginBottom: 6, padding: '4px 8px', background: 'rgba(255,183,77,0.08)', borderRadius: 4 }}>
          ⚡ {wsError}
        </div>
      )}

      <div
        ref={scrollBoxRef}
        className="ox-panel"
        style={{ flex: 1, overflowY: 'auto', padding: '10px 12px', marginBottom: 8, display: 'flex', flexDirection: 'column', gap: 2 }}
      >
        {messages.map((m) => {
          const isOwn = m.author_id === myId;
          const modifiable = canModify(m);
          const isEditing = editingId === m.id;

          return (
            <div
              key={m.id}
              className="chat-msg-row"
              style={{
                display: 'flex',
                flexDirection: 'column',
                alignItems: isOwn ? 'flex-end' : 'flex-start',
                marginBottom: 4,
              }}
            >
              {!isOwn && (
                <span style={{ fontSize: 13, color: 'var(--ox-accent)', marginBottom: 2, paddingLeft: 4 }}>
                  {m.author_name || '???'}
                </span>
              )}

              <div style={{ display: 'flex', alignItems: 'flex-end', gap: 4, flexDirection: isOwn ? 'row-reverse' : 'row', maxWidth: '78%' }}>
                {modifiable && !isEditing && (
                  <div style={{ display: 'flex', flexDirection: 'column', gap: 2, flexShrink: 0 }}>
                    <button
                      type="button"
                      onClick={() => startEdit(m)}
                      title={t('editTitle')}
                      style={{
                        fontSize: 13, lineHeight: 1, padding: '3px 5px',
                        background: 'rgba(99,217,255,0.12)', border: '1px solid rgba(99,217,255,0.25)',
                        borderRadius: 4, cursor: 'pointer', color: 'var(--ox-accent)',
                      }}
                    >✏️</button>
                    <button
                      type="button"
                      onClick={() => deleteMsg(m.id)}
                      title={t('deleteTitle')}
                      style={{
                        fontSize: 13, lineHeight: 1, padding: '3px 5px',
                        background: 'rgba(244,67,54,0.12)', border: '1px solid rgba(244,67,54,0.25)',
                        borderRadius: 4, cursor: 'pointer', color: 'var(--ox-danger)',
                      }}
                    >🗑️</button>
                  </div>
                )}

                <div
                  style={{
                    padding: '7px 11px',
                    borderRadius: isOwn ? '14px 14px 3px 14px' : '14px 14px 14px 3px',
                    background: isOwn ? 'rgba(99,217,255,0.18)' : 'rgba(255,255,255,0.06)',
                    border: `1px solid ${isOwn ? 'rgba(99,217,255,0.35)' : 'rgba(255,255,255,0.1)'}`,
                    wordBreak: 'break-word',
                    lineHeight: 1.5,
                    fontSize: 16,
                  }}
                >
                  {isEditing ? (
                    <div style={{ display: 'flex', gap: 4, alignItems: 'center' }}>
                      <input
                        value={editBody}
                        onChange={(e) => setEditBody(e.target.value)}
                        onKeyDown={(e) => { if (e.key === 'Enter') submitEdit(m.id); if (e.key === 'Escape') cancelEdit(); }}
                        style={{ flex: 1, minWidth: 120, fontSize: 15, background: 'rgba(0,0,0,0.3)', border: '1px solid var(--ox-accent)', color: 'inherit', padding: '3px 7px', borderRadius: 4 }}
                        autoFocus
                        maxLength={500}
                      />
                      <button
                        type="button"
                        onClick={() => submitEdit(m.id)}
                        style={{ padding: '3px 8px', fontSize: 15, background: 'var(--ox-accent)', color: '#000', border: 'none', borderRadius: 4, cursor: 'pointer', fontWeight: 700 }}
                      >✓</button>
                      <button
                        type="button"
                        onClick={cancelEdit}
                        style={{ padding: '3px 8px', fontSize: 15, background: 'rgba(255,255,255,0.1)', color: 'inherit', border: 'none', borderRadius: 4, cursor: 'pointer' }}
                      >✕</button>
                    </div>
                  ) : (
                    <span style={{ userSelect: 'text' }}>{m.body}</span>
                  )}
                </div>
              </div>

              <div style={{ fontSize: 10, color: 'var(--ox-fg-muted)', marginTop: 2, paddingLeft: isOwn ? 0 : 4, paddingRight: isOwn ? 4 : 0 }}>
                {new Date(m.created_at).toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })}
                {m.edited_at && <span style={{ marginLeft: 4, fontStyle: 'italic', color: 'var(--ox-fg-muted)' }}>{t('editedMark')}</span>}
              </div>
            </div>
          );
        })}
      </div>

      {showEmoji && (
        <div style={{
          display: 'flex', flexWrap: 'wrap', gap: 2, padding: '6px 8px',
          background: 'var(--ox-bg-panel)', border: '1px solid var(--ox-border)',
          marginBottom: 6, borderRadius: 6,
        }}>
          {EMOJIS.map((e) => (
            <button
              key={e}
              type="button"
              onClick={() => insertEmoji(e)}
              style={{ fontSize: 18, background: 'none', border: 'none', cursor: 'pointer', padding: '2px 4px', borderRadius: 4 }}
            >
              {e}
            </button>
          ))}
        </div>
      )}

      {confirmDeleteId && (
        <Confirm
          title={t('deleteConfirmTitle')}
          message={t('deleteConfirmMsg')}
          confirmLabel={t('deleteConfirmBtn')}
          danger
          onConfirm={confirmDelete}
          onCancel={() => setConfirmDeleteId(null)}
        />
      )}

      <div style={{ display: 'flex', gap: 6 }}>
        <button
          type="button"
          onClick={() => setShowEmoji((v) => !v)}
          style={{ fontSize: 18, padding: '0 8px', flexShrink: 0, background: showEmoji ? 'var(--ox-bg-active)' : undefined, borderRadius: 6 }}
          title={t('emojiTitle')}
        >😊</button>
        <input
          ref={inputRef}
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={handleKey}
          placeholder={t('inputPlaceholder')}
          maxLength={500}
          style={{ flex: 1 }}
        />
        <button type="button" className="btn" onClick={send} disabled={!input.trim()}>
          {t('sendBtn')}
        </button>
      </div>
    </div>
  );
}
