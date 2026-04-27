import { useState } from 'react';
import { useAuthStore } from '@/stores/auth';

// План 46 Ф.3 (149-ФЗ): кнопка "Пожаловаться" с модалкой.
// План 56 Ф.6: endpoint переехал в portal-backend (единый реестр жалоб
// для всех вселенных). VITE_PORTAL_BASE_URL — тот же env, что использует
// LoginScreen для ссылок на /privacy и /offer; в dev пустой → vite proxy
// /api/reports на game-nova не сработает, поэтому в dev требуется явно
// указать VITE_PORTAL_BASE_URL=http://localhost:8090. В проде origin
// portal'а должен быть в ALLOWED_ORIGINS portal-backend'а (cross-origin
// fetch с Authorization-header).
// targetType — куда жалуемся ('user' | 'alliance' | 'chat_msg' | 'planet').
// targetId — id объекта жалобы (player.id, alliance.id, chat_message.id, planet.id).
// label — текст кнопки (по умолчанию иконка-только).

type TargetType = 'user' | 'alliance' | 'chat_msg' | 'planet';

const PORTAL_BASE =
  (import.meta.env['VITE_PORTAL_BASE_URL'] as string | undefined) ?? '';
const REPORT_ENDPOINT = `${PORTAL_BASE}/api/reports`;

const REASONS: Array<{ value: string; label: string }> = [
  { value: 'profanity', label: 'Мат / оскорбления' },
  { value: 'extremism', label: 'Экстремизм / разжигание' },
  { value: 'drugs', label: 'Наркотики' },
  { value: 'spam', label: 'Спам / реклама' },
  { value: 'impersonation', label: 'Выдача за другое лицо' },
  { value: 'cheat', label: 'Чит / эксплойт' },
  { value: 'other', label: 'Другое' },
];

export function ReportButton({
  targetType,
  targetId,
  label,
  compact,
}: {
  targetType: TargetType;
  targetId: string;
  label?: string;
  compact?: boolean;
}) {
  const [open, setOpen] = useState(false);
  const [reason, setReason] = useState('profanity');
  const [comment, setComment] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [result, setResult] = useState<'ok' | 'err' | null>(null);
  const [errMsg, setErrMsg] = useState<string>('');

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    if (submitting) return;
    setSubmitting(true);
    setResult(null);
    try {
      const token = useAuthStore.getState().accessToken;
      const res = await fetch(REPORT_ENDPOINT, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
        body: JSON.stringify({
          target_type: targetType,
          target_id: targetId,
          reason,
          comment,
        }),
      });
      if (!res.ok) {
        let msg = `HTTP ${res.status}`;
        try {
          const body = (await res.json()) as { error?: { message?: string } };
          if (body.error?.message) msg = body.error.message;
        } catch {
          // ignore
        }
        throw new Error(msg);
      }
      setResult('ok');
      setComment('');
    } catch (err) {
      setResult('err');
      setErrMsg(err instanceof Error ? err.message : 'unknown error');
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <>
      <button
        type="button"
        onClick={() => setOpen(true)}
        title="Пожаловаться"
        aria-label="Пожаловаться"
        className={compact ? 'btn-ghost btn-sm' : 'btn-ghost'}
        style={{
          display: 'inline-flex',
          alignItems: 'center',
          gap: 4,
          fontSize: compact ? 12 : 13,
          color: 'var(--ox-fg-muted)',
          padding: compact ? '2px 6px' : '4px 10px',
        }}
      >
        🚩{label ? ` ${label}` : ''}
      </button>

      {open && (
        <div
          className="ox-modal-overlay"
          onClick={() => setOpen(false)}
          style={{ position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.5)', zIndex: 1000, display: 'flex', alignItems: 'center', justifyContent: 'center' }}
        >
          <div
            className="ox-modal"
            onClick={(e) => e.stopPropagation()}
            style={{ background: 'var(--ox-bg-panel)', border: '1px solid var(--ox-border)', borderRadius: 8, padding: 24, maxWidth: 420, width: '90%' }}
          >
            <h3 style={{ marginTop: 0, fontSize: 18, color: 'var(--ox-accent)' }}>
              Пожаловаться
            </h3>
            {result === 'ok' ? (
              <>
                <p style={{ color: 'var(--ox-fg-dim)', lineHeight: 1.5 }}>
                  Спасибо. Жалоба принята и будет рассмотрена модератором
                  в течение 24 часов.
                </p>
                <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
                  <button type="button" className="btn" onClick={() => { setOpen(false); setResult(null); }}>
                    Закрыть
                  </button>
                </div>
              </>
            ) : (
              <form onSubmit={submit} style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
                <label style={{ display: 'flex', flexDirection: 'column', gap: 4, fontSize: 13, color: 'var(--ox-fg-dim)' }}>
                  Причина
                  <select
                    value={reason}
                    onChange={(e) => setReason(e.target.value)}
                    required
                    style={{ padding: '6px 10px', background: 'var(--ox-bg-2)', border: '1px solid var(--ox-border)', color: 'var(--ox-fg)', borderRadius: 4 }}
                  >
                    {REASONS.map((r) => (
                      <option key={r.value} value={r.value}>{r.label}</option>
                    ))}
                  </select>
                </label>
                <label style={{ display: 'flex', flexDirection: 'column', gap: 4, fontSize: 13, color: 'var(--ox-fg-dim)' }}>
                  Комментарий (необязательно)
                  <textarea
                    value={comment}
                    onChange={(e) => setComment(e.target.value)}
                    maxLength={1000}
                    rows={4}
                    placeholder="Опишите подробнее, что произошло"
                    style={{ padding: '6px 10px', background: 'var(--ox-bg-2)', border: '1px solid var(--ox-border)', color: 'var(--ox-fg)', borderRadius: 4, fontFamily: 'inherit', resize: 'vertical' }}
                  />
                </label>
                {result === 'err' && (
                  <div style={{ color: 'var(--ox-danger)', fontSize: 13 }}>
                    Не удалось отправить: {errMsg}
                  </div>
                )}
                <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 8 }}>
                  <button type="button" className="btn-ghost" onClick={() => setOpen(false)}>
                    Отмена
                  </button>
                  <button type="submit" className="btn" disabled={submitting}>
                    {submitting ? '…' : 'Отправить'}
                  </button>
                </div>
              </form>
            )}
          </div>
        </div>
      )}
    </>
  );
}
