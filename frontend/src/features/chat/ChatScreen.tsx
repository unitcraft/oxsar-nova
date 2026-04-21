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

    const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;
    const ws = new WebSocket(`${proto}//${host}/api/chat/${kind}/ws?token=${encodeURIComponent(token)}`);
    wsRef.current = ws;

    ws.onmessage = (ev: MessageEvent) => {
      try {
        const msg = JSON.parse(ev.data as string) as ChatMessage;
        setMessages((prev) => [...prev, msg]);
      } catch {
        // ignore malformed frames
      }
    };
    ws.onerror = () => setWsError('Соединение прервано');
    ws.onclose = () => setWsError('Соединение закрыто');

    return () => {
      ws.close();
      wsRef.current = null;
    };
  }, [kind, token]);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  function send() {
    const body = input.trim();
    if (!body) return;
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ body }));
    } else {
      // fallback REST
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

      <div style={{ flex: 1, overflowY: 'auto', border: '1px solid #333', padding: 8, marginBottom: 8 }}>
        {messages.map((m) => (
          <div key={m.id} style={{ marginBottom: 4 }}>
            <span style={{ color: '#aaa', fontSize: 11 }}>
              {new Date(m.created_at).toLocaleTimeString()}
            </span>{' '}
            <strong>{m.author_name || '???'}</strong>:{' '}
            {m.body}
          </div>
        ))}
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
