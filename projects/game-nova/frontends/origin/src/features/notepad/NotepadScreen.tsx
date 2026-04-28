// S-038 Notepad — личный блокнот игрока (план 72 Ф.5 Spring 4).
//
// Pixel-perfect зеркало legacy `templates/standard/notes.tpl`:
//   ntable c шапкой "Блокнот" + textarea + auto-save с debounce.
//
// Endpoints: GET/PUT /api/notepad — план 69 backend.

import { useEffect, useRef, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  fetchNotepad,
  saveNotepad,
  NOTEPAD_MAX_LENGTH,
} from '@/api/notepad';
import { QK } from '@/api/query-keys';
import { useTranslation } from '@/i18n/i18n';

const DEBOUNCE_MS = 1000;

type SaveStatus = 'idle' | 'saving' | 'saved' | 'error';

export function NotepadScreen() {
  const { t } = useTranslation();
  const qc = useQueryClient();

  const noteQ = useQuery({
    queryKey: QK.notepad(),
    queryFn: fetchNotepad,
  });

  const [draft, setDraft] = useState<string>('');
  const [status, setStatus] = useState<SaveStatus>('idle');
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const initialised = useRef(false);

  useEffect(() => {
    if (!initialised.current && noteQ.data !== undefined) {
      setDraft(noteQ.data.content);
      initialised.current = true;
    }
  }, [noteQ.data]);

  const save = useMutation({
    mutationFn: saveNotepad,
    onMutate: () => setStatus('saving'),
    onSuccess: (data) => {
      setStatus('saved');
      qc.setQueryData(QK.notepad(), data);
    },
    onError: () => setStatus('error'),
  });

  function onChange(e: React.ChangeEvent<HTMLTextAreaElement>) {
    const next = e.target.value.slice(0, NOTEPAD_MAX_LENGTH);
    setDraft(next);
    if (debounceRef.current) clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(() => save.mutate(next), DEBOUNCE_MS);
  }

  if (noteQ.isLoading) {
    return <div className="idiv">{t('notepad', 'statusSaving')}</div>;
  }

  return (
    <table className="ntable">
      <thead>
        <tr>
          <th style={{ textAlign: 'center' }}>{t('notepad', 'title')}</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td style={{ textAlign: 'center' }}>{t('notepad', 'hint')}</td>
        </tr>
        <tr>
          <td style={{ textAlign: 'center' }}>
            <textarea
              rows={20}
              name="notes"
              style={{ width: '90%' }}
              value={draft}
              maxLength={NOTEPAD_MAX_LENGTH}
              placeholder={t('notepad', 'placeholder')}
              onChange={onChange}
            />
            <br />
            <span className="small">
              {t('notepad', 'charCount', {
                count: draft.length,
                max: NOTEPAD_MAX_LENGTH,
              })}
            </span>
          </td>
        </tr>
        <tr>
          <td style={{ textAlign: 'center' }}>
            {status === 'saving' && (
              <span>{t('notepad', 'statusSaving')}</span>
            )}
            {status === 'saved' && (
              <span className="true">{t('notepad', 'statusSaved')}</span>
            )}
            {status === 'error' && (
              <span className="false">{t('notepad', 'statusError')}</span>
            )}
            {noteQ.data?.updated_at && status === 'idle' && (
              <span className="small">
                {t('notepad', 'updatedAt')}{' '}
                {new Date(noteQ.data.updated_at).toLocaleString('ru-RU')}
              </span>
            )}
          </td>
        </tr>
      </tbody>
    </table>
  );
}
