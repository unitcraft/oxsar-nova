import { useEffect, useRef, useState } from 'react';
import { useQuery, useMutation } from '@tanstack/react-query';
import { api } from '@/api/client';
import { useTranslation } from '@/i18n/i18n';

interface NotepadData {
  content: string;
  updated_at: string;
}

type SaveStatus = 'idle' | 'saving' | 'saved' | 'error';

export function NotepadScreen() {
  const { t } = useTranslation('notepad');
  const q = useQuery({
    queryKey: ['notepad'],
    queryFn: () => api.get<NotepadData>('/api/notepad'),
    staleTime: Infinity,
  });

  const save = useMutation({
    mutationFn: (content: string) => api.put<void>('/api/notepad', { content }),
  });

  const [text, setText] = useState('');
  const [status, setStatus] = useState<SaveStatus>('idle');
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const loaded = useRef(false);

  useEffect(() => {
    if (q.data && !loaded.current) {
      setText(q.data.content);
      loaded.current = true;
    }
  }, [q.data]);

  function onChange(next: string) {
    setText(next);
    setStatus('saving');
    if (debounceRef.current) clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(() => {
      save.mutate(next, {
        onSuccess: () => setStatus('saved'),
        onError: () => setStatus('error'),
      });
    }, 500);
  }

  const charCount = text.length;
  const maxChars = 50000;

  if (q.isLoading) {
    return <div style={{ padding: 24 }}><div className="ox-skeleton" style={{ height: 400, borderRadius: 8 }} /></div>;
  }

  return (
    <div style={{ maxWidth: 900, margin: '0 auto', padding: 24, display: 'flex', flexDirection: 'column', gap: 12 }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: 8 }}>
        <h2 style={{ margin: 0, fontSize: 20, fontWeight: 700 }}>{t('title')}</h2>
        <StatusBadge status={status} />
      </div>

      <p style={{ margin: 0, fontSize: 14, color: 'var(--ox-fg-dim)' }}>
        {t('hint')}
      </p>

      <textarea
        value={text}
        onChange={(e) => onChange(e.target.value)}
        placeholder={t('placeholder')}
        style={{
          width: '100%',
          minHeight: 500,
          padding: 16,
          fontFamily: 'var(--ox-mono)',
          fontSize: 15,
          lineHeight: 1.6,
          color: 'var(--ox-fg)',
          background: 'var(--ox-bg-panel)',
          border: '1px solid var(--ox-border)',
          borderRadius: 6,
          resize: 'vertical',
        }}
        maxLength={maxChars}
      />

      <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 13, color: 'var(--ox-fg-muted)', fontFamily: 'var(--ox-mono)' }}>
        <span>{t('charCount', { count: charCount.toLocaleString('ru-RU'), max: maxChars.toLocaleString('ru-RU') })}</span>
        {q.data?.updated_at && (
          <span>{t('updatedAt')} {new Date(q.data.updated_at).toLocaleString('ru-RU')}</span>
        )}
      </div>
    </div>
  );
}

function StatusBadge({ status }: { status: SaveStatus }) {
  const { t } = useTranslation('notepad');
  if (status === 'saving') return <span style={{ fontSize: 14, color: 'var(--ox-fg-muted)' }}>{t('statusSaving')}</span>;
  if (status === 'saved') return <span style={{ fontSize: 14, color: 'var(--ox-success)' }}>{t('statusSaved')}</span>;
  if (status === 'error') return <span style={{ fontSize: 14, color: 'var(--ox-danger)' }}>{t('statusError')}</span>;
  return null;
}
