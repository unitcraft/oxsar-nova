import { useEffect, useRef, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '@/api/client';
import { useAuthStore } from '@/stores/auth';

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
  const [kind, setKind] = useState<ChannelKind>('global');
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [input, setInput] = useState('');
  const [wsError, setWsError] = useState('');
  const [showEmoji, setShowEmoji] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editBody, setEditBody] = useState('');
  const wsRef = useRef<WebSocket | null>(null);
  const scrollBoxRef = useRef<HTMLDivElement | null>(null);
  const prevLenRef = useRef(0);
  const inputRef = useRef<HTMLInputElement | null>(null);

  const token = useAuthStore((s) => s.accessToken);
  const me = useQuery({
    queryKey: ['me'],
    queryFn: () => api.get<{ user_id: string; username: string }>('/api/me'),
    staleTime: 60000,
  });

  const { data: history } = useQuery<ChatMessage[]>({
    queryKey: ['chat-history', kind],
    queryFn: () => api.get<ChatMessage[]>(`/api/chat/${kind}/history`),
    staleTime: 0,
  });

  useEffect(() => {
    if (history) setMessages(history);
  }, [history]);

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
          // kind === "msg" или не задан (старые сообщения)
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
          setWsError('Переподключение…');
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
  }, [kind, token]);

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
      author_id: me.data?.user_id ?? '',
      author_name: me.data?.username ?? '…',
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
    api.delete(`/api/chat/messages/${id}`)
      .then(() => setMessages((prev) => prev.filter((m) => m.id !== id)))
      .catch(() => null);
  }

  function canModify(m: ChatMessage): boolean {
    if (m.id.startsWith('tmp-')) return false;
    if (m.author_id !== me.data?.user_id) return false;
    return Date.now() - new Date(m.created_at).getTime() < EDIT_WINDOW_MS;
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '70vh' }}>
      <div style={{ marginBottom: 8, display: 'flex', gap: 8 }}>
        <button
          onClick={() => setKind('global')}
          style={{ fontWeight: kind === 'global' ? 'bold' : 'normal' }}
        >
          Глобальный
        </button>
        <button
          onClick={() => setKind('alliance')}
          style={{ fontWeight: kind === 'alliance' ? 'bold' : 'normal' }}
        >
          Альянс
        </button>
      </div>

      {wsError && <div style={{ color: 'orange', marginBottom: 4 }}>{wsError}</div>}

      <div
        ref={scrollBoxRef}
        style={{ flex: 1, overflowY: 'auto', border: '1px solid #333', padding: '8px 12px', marginBottom: 8 }}
      >
        {messages.map((m) => {
          const isOwn = m.author_id === me.data?.user_id;
          const modifiable = canModify(m);
          const isEditing = editingId === m.id;

          return (
            <div
              key={m.id}
              style={{
                display: 'flex',
                flexDirection: 'column',
                alignItems: isOwn ? 'flex-end' : 'flex-start',
                marginBottom: 8,
              }}
            >
              {!isOwn && (
                <span style={{ fontSize: 11, color: 'var(--ox-accent)', marginBottom: 2 }}>
                  {m.author_name || '???'}
                </span>
              )}
              <div
                style={{
                  maxWidth: '70%',
                  padding: '6px 10px',
                  borderRadius: isOwn ? '12px 12px 2px 12px' : '12px 12px 12px 2px',
                  background: isOwn ? 'rgba(79,195,247,0.2)' : 'rgba(255,255,255,0.07)',
                  border: `1px solid ${isOwn ? 'rgba(79,195,247,0.4)' : 'rgba(255,255,255,0.1)'}`,
                  wordBreak: 'break-word',
                  lineHeight: 1.4,
                }}
              >
                {isEditing ? (
                  <div style={{ display: 'flex', gap: 4 }}>
                    <input
                      value={editBody}
                      onChange={(e) => setEditBody(e.target.value)}
                      onKeyDown={(e) => { if (e.key === 'Enter') submitEdit(m.id); if (e.key === 'Escape') cancelEdit(); }}
                      style={{ flex: 1, fontSize: 13, background: 'rgba(0,0,0,0.3)', border: '1px solid #4fc3f7', color: 'inherit', padding: '2px 6px' }}
                      autoFocus
                      maxLength={500}
                    />
                    <button type="button" onClick={() => submitEdit(m.id)} style={{ padding: '2px 6px', fontSize: 12 }}>✓</button>
                    <button type="button" onClick={cancelEdit} style={{ padding: '2px 6px', fontSize: 12 }}>✕</button>
                  </div>
                ) : (
                  m.body
                )}
              </div>
              <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginTop: 2 }}>
                <span style={{ fontSize: 10, color: '#666' }}>
                  {new Date(m.created_at).toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })}
                  {m.edited_at && <span style={{ marginLeft: 4, fontStyle: 'italic' }}>изм.</span>}
                </span>
                {modifiable && !isEditing && (
                  <>
                    <button
                      type="button"
                      onClick={() => startEdit(m)}
                      style={{ fontSize: 10, padding: '1px 5px', opacity: 0.6 }}
                      title="Редактировать"
                    >✏️</button>
                    <button
                      type="button"
                      onClick={() => deleteMsg(m.id)}
                      style={{ fontSize: 10, padding: '1px 5px', opacity: 0.6 }}
                      title="Удалить"
                    >🗑️</button>
                  </>
                )}
              </div>
            </div>
          );
        })}
      </div>

      {showEmoji && (
        <div style={{
          display: 'flex', flexWrap: 'wrap', gap: 4, padding: '6px 8px',
          background: 'rgba(16,28,44,0.95)', border: '1px solid #1e3a5a',
          marginBottom: 6, borderRadius: 6,
        }}>
          {EMOJIS.map((e) => (
            <button
              key={e}
              type="button"
              onClick={() => insertEmoji(e)}
              style={{ fontSize: 20, background: 'none', border: 'none', cursor: 'pointer', padding: '2px 4px' }}
            >
              {e}
            </button>
          ))}
        </div>
      )}

      <div style={{ display: 'flex', gap: 6 }}>
        <button
          type="button"
          onClick={() => setShowEmoji((v) => !v)}
          style={{ fontSize: 18, padding: '4px 8px', flexShrink: 0 }}
          title="Смайлики"
        >
          😊
        </button>
        <input
          ref={inputRef}
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={handleKey}
          placeholder="Сообщение…"
          maxLength={500}
          style={{ flex: 1 }}
        />
        <button type="button" onClick={send}>Отправить</button>
      </div>
    </div>
  );
}
