import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';

// План 46 Ф.3 (149-ФЗ): админская модерация UGC-жалоб.
// Поток: list (фильтр new/resolved/rejected) → detail-модалка с
// текстом + комментарий → POST /resolve со статусом и пометкой.
//
// Действие (warn/mute/rename/ban) выполняется отдельно через
// существующее /api/admin/users/{id}/* (план 14). Здесь — только
// статус жалобы и пометка о принятом решении.

interface Report {
  id: string;
  reporter_id: string;
  reporter_name: string;
  target_type: 'user' | 'alliance' | 'chat_msg' | 'planet';
  target_id: string;
  reason: string;
  comment: string;
  status: 'new' | 'resolved' | 'rejected';
  resolved_by?: string;
  resolver_name?: string;
  resolution_note?: string;
  created_at: string;
  resolved_at?: string;
}

type StatusFilter = '' | 'new' | 'resolved' | 'rejected';

export function AdminReportsTab() {
  const qc = useQueryClient();
  const [status, setStatus] = useState<StatusFilter>('new');
  const [active, setActive] = useState<Report | null>(null);

  const list = useQuery({
    queryKey: ['admin-reports', status],
    queryFn: () =>
      api.get<{ reports: Report[] }>(`/api/admin/reports?status=${status}&limit=100`),
    refetchInterval: 30000,
  });

  return (
    <div>
      <div style={{ display: 'flex', gap: 8, marginBottom: 12, alignItems: 'center' }}>
        <span style={{ fontSize: 13, color: 'var(--ox-fg-muted)' }}>Статус:</span>
        {(['new', 'resolved', 'rejected', ''] as const).map((s) => (
          <button
            key={s || 'all'}
            type="button"
            className={status === s ? 'btn' : 'btn-ghost'}
            onClick={() => setStatus(s)}
          >
            {s === '' ? 'Все' : s === 'new' ? 'Новые' : s === 'resolved' ? 'Решённые' : 'Отклонённые'}
          </button>
        ))}
        <span style={{ marginLeft: 'auto', fontSize: 12, color: 'var(--ox-fg-muted)' }}>
          {list.data?.reports?.length ?? 0} запис.
        </span>
      </div>

      <table className="ox-table" style={{ width: '100%', fontSize: 13 }}>
        <thead>
          <tr>
            <th style={{ textAlign: 'left' }}>Дата</th>
            <th style={{ textAlign: 'left' }}>От</th>
            <th style={{ textAlign: 'left' }}>Тип</th>
            <th style={{ textAlign: 'left' }}>Цель</th>
            <th style={{ textAlign: 'left' }}>Причина</th>
            <th style={{ textAlign: 'left' }}>Статус</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          {(list.data?.reports ?? []).map((r) => (
            <tr key={r.id}>
              <td style={{ fontFamily: 'var(--ox-mono)', fontSize: 12 }}>
                {new Date(r.created_at).toLocaleString('ru-RU')}
              </td>
              <td>{r.reporter_name || r.reporter_id.slice(0, 8)}</td>
              <td><code>{r.target_type}</code></td>
              <td style={{ fontFamily: 'var(--ox-mono)', fontSize: 11 }}>{r.target_id.slice(0, 12)}…</td>
              <td>{r.reason}</td>
              <td>
                <span style={{
                  padding: '2px 6px', borderRadius: 4, fontSize: 11,
                  background: r.status === 'new' ? 'rgba(239,68,68,0.15)' : r.status === 'resolved' ? 'rgba(34,197,94,0.15)' : 'rgba(148,163,184,0.15)',
                  color: r.status === 'new' ? 'var(--ox-danger)' : r.status === 'resolved' ? 'var(--ox-success, #22c55e)' : 'var(--ox-fg-muted)',
                }}>
                  {r.status}
                </span>
              </td>
              <td>
                <button type="button" className="btn-ghost btn-sm" onClick={() => setActive(r)}>
                  Открыть
                </button>
              </td>
            </tr>
          ))}
          {(list.data?.reports?.length ?? 0) === 0 && (
            <tr>
              <td colSpan={7} style={{ textAlign: 'center', color: 'var(--ox-fg-muted)', padding: 16 }}>
                Жалоб нет
              </td>
            </tr>
          )}
        </tbody>
      </table>

      {active && <ReportDetailModal report={active} onClose={() => { setActive(null); void qc.invalidateQueries({ queryKey: ['admin-reports'] }); }} />}
    </div>
  );
}

function ReportDetailModal({ report, onClose }: { report: Report; onClose: () => void }) {
  const [note, setNote] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const resolve = useMutation({
    mutationFn: ({ status }: { status: 'resolved' | 'rejected' }) =>
      api.post(`/api/admin/reports/${report.id}/resolve`, { status, note }),
    onSuccess: onClose,
  });

  async function handle(status: 'resolved' | 'rejected') {
    setSubmitting(true);
    try {
      await resolve.mutateAsync({ status });
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div
      onClick={onClose}
      style={{ position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.5)', zIndex: 1000, display: 'flex', alignItems: 'center', justifyContent: 'center' }}
    >
      <div
        onClick={(e) => e.stopPropagation()}
        style={{ background: 'var(--ox-bg-panel)', border: '1px solid var(--ox-border)', borderRadius: 8, padding: 24, maxWidth: 560, width: '90%', maxHeight: '90vh', overflowY: 'auto' }}
      >
        <h3 style={{ marginTop: 0 }}>Жалоба от {report.reporter_name}</h3>

        <dl style={{ display: 'grid', gridTemplateColumns: 'auto 1fr', gap: '4px 12px', fontSize: 13, marginBottom: 16 }}>
          <dt style={{ color: 'var(--ox-fg-muted)' }}>Дата:</dt>
          <dd style={{ margin: 0, fontFamily: 'var(--ox-mono)' }}>{new Date(report.created_at).toLocaleString('ru-RU')}</dd>
          <dt style={{ color: 'var(--ox-fg-muted)' }}>Тип цели:</dt>
          <dd style={{ margin: 0 }}><code>{report.target_type}</code></dd>
          <dt style={{ color: 'var(--ox-fg-muted)' }}>ID цели:</dt>
          <dd style={{ margin: 0, fontFamily: 'var(--ox-mono)', fontSize: 12 }}>{report.target_id}</dd>
          <dt style={{ color: 'var(--ox-fg-muted)' }}>Причина:</dt>
          <dd style={{ margin: 0 }}>{report.reason}</dd>
          <dt style={{ color: 'var(--ox-fg-muted)' }}>Статус:</dt>
          <dd style={{ margin: 0 }}>{report.status}</dd>
        </dl>

        {report.comment && (
          <div style={{ background: 'var(--ox-bg-2)', border: '1px solid var(--ox-border)', padding: 10, borderRadius: 4, fontSize: 13, marginBottom: 16, whiteSpace: 'pre-wrap' }}>
            {report.comment}
          </div>
        )}

        {report.status === 'new' ? (
          <>
            <label style={{ display: 'flex', flexDirection: 'column', gap: 4, fontSize: 13, color: 'var(--ox-fg-dim)' }}>
              Пометка о принятом решении
              <textarea
                value={note}
                onChange={(e) => setNote(e.target.value)}
                rows={3}
                maxLength={1000}
                placeholder="Что сделано (warn/mute/ban/rename) или почему отклонено"
                style={{ padding: '6px 10px', background: 'var(--ox-bg-2)', border: '1px solid var(--ox-border)', color: 'var(--ox-fg)', borderRadius: 4, fontFamily: 'inherit', resize: 'vertical' }}
              />
            </label>
            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 8, marginTop: 12 }}>
              <button type="button" className="btn-ghost" onClick={onClose}>Закрыть</button>
              <button type="button" className="btn-ghost" disabled={submitting} onClick={() => void handle('rejected')}>
                Отклонить
              </button>
              <button type="button" className="btn" disabled={submitting} onClick={() => void handle('resolved')}>
                Принять
              </button>
            </div>
          </>
        ) : (
          <>
            <p style={{ color: 'var(--ox-fg-muted)', fontSize: 13 }}>
              Решено {report.resolver_name || '?'} — {report.resolved_at ? new Date(report.resolved_at).toLocaleString('ru-RU') : ''}
            </p>
            {report.resolution_note && (
              <div style={{ background: 'var(--ox-bg-2)', border: '1px solid var(--ox-border)', padding: 10, borderRadius: 4, fontSize: 13 }}>
                {report.resolution_note}
              </div>
            )}
            <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: 12 }}>
              <button type="button" className="btn" onClick={onClose}>Закрыть</button>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
