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
}

type ChannelKind = 'global' | 'alliance';

export function ChatScreen() {
  const [kind, setKind] = useState<ChannelKind>('global');
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [input, setInput] = useState('');
  const [wsError, setWsError] = useState('');
  const wsRef = useRef<WebSocket | null>(null);
  const bottomRef = useRef<HTMLDivElement | null>(null);
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
          setMessages((prev) => {
            // заменяем первое tmp-сообщение с тем же телом от того же автора
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

  const prevLenRef = useRef(0);
  useEffect(() => {
    if (messages.length > prevLenRef.current) {
      bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
    }
    prevLenRef.current = messages.length;
  }, [messages]);

  function send() {
    const body = input.trim();
    if (!body) return;
    // Optimistic: добавляем сообщение сразу, сервер вернёт настоящее через WS
    const optimistic: ChatMessage = {
      id: `tmp-${Date.now()}`,
      channel: kind,
      author_id: me.data?.user_id ?? '',
      author_name: me.data?.username ?? '…',
      body,
      created_at: new Date().toISOString(),
    };
    setMessages((prev) => [...prev, optimistic]);
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ body }));
    } else {
      api.post(`/api/chat/${kind}/send`, { body }).catch(() => null);
    }
    setInput('');
  }

  function handleKey(e: React.KeyboardEvent<HTMLInputElement>) {
    if (e.key === 'Enter') send();
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

      <div style={{ flex: 1, overflowY: 'auto', border: '1px solid #333', padding: '8px 12px', marginBottom: 8 }}>
        {messages.map((m) => {
          const isOwn = m.author_id === me.data?.user_id;
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
                {m.body}
              </div>
              <span style={{ fontSize: 10, color: '#666', marginTop: 2 }}>
                {new Date(m.created_at).toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })}
              </span>
            </div>
          );
        })}
        <div ref={bottomRef} />
      </div>

      <div style={{ display: 'flex', gap: 8 }}>
        <input
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={handleKey}
          placeholder="Сообщение…"
          maxLength={500}
          style={{ flex: 1 }}
        />
        <button onClick={send}>Отправить</button>
      </div>
    </div>
  );
}
