// API-модуль notepad origin-фронта (план 72 Ф.5 Spring 4 — S-038).
//
// Endpoints (openapi.yaml):
//   GET /api/notepad
//   PUT /api/notepad   body: { content }    Idempotency-Key: required
//
// Лимит длины 50_000 символов (план 69, см. internal/notepad/handler.go).

import { api } from './client';
import { newIdempotencyKey } from './idempotency';
import type { NotepadContent } from './types';

export const NOTEPAD_MAX_LENGTH = 50_000;

export function fetchNotepad(): Promise<NotepadContent> {
  return api.get<NotepadContent>('/api/notepad');
}

export function saveNotepad(content: string): Promise<NotepadContent> {
  return api.put<NotepadContent>(
    '/api/notepad',
    { content },
    { idempotencyKey: newIdempotencyKey() },
  );
}
