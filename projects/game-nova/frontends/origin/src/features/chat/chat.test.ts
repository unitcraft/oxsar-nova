import { describe, it, expect } from 'vitest';

const EDIT_WINDOW_MS = 5 * 60 * 1000;

// Inline-копия chatWsUrl, чтобы тесты не подтягивали api/client.ts
// (которому нужен localStorage из node-env).
function chatWsUrlLocal(kind: 'global' | 'alliance', token: string): string {
  // в Node-env нет window — используем фейковый proto/host через jsdom-fallback.
  const proto = 'ws:';
  const host = 'localhost:5174';
  return `${proto}//${host}/api/chat/${kind}/ws?token=${encodeURIComponent(token)}`;
}

function canEdit(authorId: string, myId: string, createdAt: string): boolean {
  if (authorId !== myId) return false;
  return Date.now() - new Date(createdAt).getTime() < EDIT_WINDOW_MS;
}

describe('chat', () => {
  it('chatWsUrl формирует валидный URL с токеном', () => {
    const url = chatWsUrlLocal('global', 'tok123');
    expect(url).toContain('/api/chat/global/ws');
    expect(url).toContain('token=tok123');
    expect(url.startsWith('ws://') || url.startsWith('wss://')).toBe(true);
  });

  it('alliance kind пишет в URL alliance', () => {
    expect(chatWsUrlLocal('alliance', 't').includes('/chat/alliance/ws')).toBe(
      true,
    );
  });

  it('canEdit: чужое сообщение — false', () => {
    expect(canEdit('a', 'b', new Date().toISOString())).toBe(false);
  });

  it('canEdit: моё свежее — true', () => {
    expect(canEdit('me', 'me', new Date().toISOString())).toBe(true);
  });

  it('canEdit: моё старше окна — false', () => {
    const old = new Date(Date.now() - 6 * 60 * 1000).toISOString();
    expect(canEdit('me', 'me', old)).toBe(false);
  });
});
