import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';
import { Confirm } from '@/ui/Confirm';
import { useToast } from '@/ui/Toast';

interface Message {
  id: string;
  from_user_id?: string | null;
  from_username?: string;
  subject: string;
  body: string;
  folder: number;
  created_at: string;
  read_at?: string | null;
  battle_report_id?: string | null;
  espionage_report_id?: string | null;
  expedition_report_id?: string | null;
}

interface ExpeditionReportFull {
  id: string;
  user_id?: string | null;
  fleet_id?: string | null;
  outcome: 'resources' | 'artefact' | 'extra_planet' | 'pirates' | 'loss' | 'nothing';
  at: string;
  report: Record<string, unknown>;
}

interface EspionagePayload {
  ratio: number;
  probes: number;
  metal: number;
  silicon: number;
  hydrogen: number;
  ships?: Record<string, number>;
  defense?: Record<string, number>;
  buildings?: Record<string, number>;
}

interface EspionageReportFull {
  id: string;
  spy_user_id?: string | null;
  target_user_id?: string | null;
  planet_id?: string | null;
  ratio: number;
  probes: number;
  at: string;
  report: EspionagePayload;
}

interface UnitResult {
  unit_id: number;
  quantity_start: number;
  quantity_end: number;
  damaged_end?: number;
  shell_percent_end?: number;
}

interface SideResult {
  user_id: string;
  username?: string;
  lost_metal: number;
  lost_silicon: number;
  lost_hydrogen: number;
  units: UnitResult[];
}

interface RoundTrace {
  index: number;
  attackers_alive: number;
  defenders_alive: number;
}

interface BattleReportPayload {
  seed?: number;
  winner: 'attackers' | 'defenders' | 'draw';
  rounds: number;
  rounds_trace?: RoundTrace[];
  attackers?: SideResult[];
  defenders?: SideResult[];
}

interface BattleReportFull {
  id: string;
  attacker_user_id?: string | null;
  defender_user_id?: string | null;
  attacker_username?: string;
  defender_username?: string;
  planet_id?: string | null;
  dst_galaxy?: number | null;
  dst_system?: number | null;
  dst_position?: number | null;
  seed: number;
  winner: string;
  rounds: number;
  loot_metal: number;
  loot_silicon: number;
  loot_hydrogen: number;
  at: string;
  report: BattleReportPayload;
}

interface ReplyInit {
  to: string;
  subject: string;
}

const FOLDERS: { folder: number | null; label: string; icon: string }[] = [
  { folder: null,  label: 'Все',        icon: '📬' },
  { folder: 1,     label: 'Личные',     icon: '✉️' },
  { folder: 2,     label: 'Бой',        icon: '⚔️' },
  { folder: 3,     label: 'Шпионаж',    icon: '🔭' },
  { folder: 4,     label: 'Экспедиции', icon: '🌌' },
  { folder: 13,    label: 'Система',    icon: '⚙️' },
];

type FleetMissionCb = (g: number, s: number, pos: number, isMoon: boolean, mission: number) => void;

export function MessagesScreen({ onFleetMission }: { onFleetMission?: FleetMissionCb }) {
  const qc = useQueryClient();
  const toast = useToast();
  const [selectedID, setSelectedID] = useState<string | null>(null);
  const [composing, setComposing] = useState(false);
  const [replyInit, setReplyInit] = useState<ReplyInit | undefined>(undefined);
  const [activeFolder, setActiveFolder] = useState<number | null>(null);
  const [confirmDelAll, setConfirmDelAll] = useState(false);

  const list = useQuery({
    queryKey: ['messages'],
    queryFn: () => api.get<{ messages: Message[] | null }>('/api/messages'),
    refetchInterval: 10000,
  });

  const markRead = useMutation({
    mutationFn: (id: string) => api.post<void>(`/api/messages/${id}/read`),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['messages'] }),
  });

  const del = useMutation({
    mutationFn: (id: string) => api.delete<void>(`/api/messages/${id}`),
    onSuccess: () => {
      setSelectedID(null);
      void qc.invalidateQueries({ queryKey: ['messages'] });
      void qc.invalidateQueries({ queryKey: ['messages-unread'] });
      toast.show('info', 'Сообщение удалено');
    },
  });

  const delAll = useMutation({
    mutationFn: () => {
      const qs = activeFolder != null ? `?folder=${activeFolder}` : '';
      return api.delete<void>(`/api/messages${qs}`);
    },
    onSuccess: () => {
      setSelectedID(null);
      void qc.invalidateQueries({ queryKey: ['messages'] });
      void qc.invalidateQueries({ queryKey: ['messages-unread'] });
      toast.show('info', 'Все сообщения удалены');
    },
  });

  const allMsgs = list.data?.messages ?? [];
  const msgs = activeFolder === null ? allMsgs : allMsgs.filter((m) => m.folder === activeFolder);
  const selected = msgs.find((m) => m.id === selectedID) ?? null;
  const unreadCount = allMsgs.filter((m) => !m.read_at).length;

  function onSelect(m: Message) {
    setSelectedID(m.id);
    setComposing(false);
    setReplyInit(undefined);
    if (!m.read_at) markRead.mutate(m.id);
  }

  function onReply(m: Message) {
    setReplyInit({
      to: m.from_username ?? '',
      subject: m.subject.startsWith('Re: ') ? m.subject : `Re: ${m.subject}`,
    });
    setComposing(true);
    setSelectedID(null);
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: 8 }}>
        <h2 style={{ margin: 0, fontSize: 18, fontFamily: 'var(--ox-font)', fontWeight: 700 }}>
          Сообщения {unreadCount > 0 && <span style={{ fontSize: 14, color: 'var(--ox-accent)' }}>({unreadCount} новых)</span>}
        </h2>
        <div style={{ display: 'flex', gap: 8 }}>
          {msgs.length > 0 && (
            <button
              type="button"
              className="btn-ghost btn-sm"
              style={{ color: 'var(--ox-danger)', opacity: 0.7 }}
              disabled={delAll.isPending}
              onClick={() => setConfirmDelAll(true)}
            >
              🗑 Удалить все
            </button>
          )}
          <button type="button" className="btn btn-sm" onClick={() => { setComposing(true); setSelectedID(null); }}>
            ✉ Написать
          </button>
        </div>
      </div>

      {/* Folder tabs */}
      <div style={{ display: 'flex', gap: 4, flexWrap: 'wrap' }}>
        {FOLDERS.map(({ folder, label, icon }) => {
          const isActive = activeFolder === folder;
          const count = folder === null ? allMsgs.length : allMsgs.filter((m) => m.folder === folder).length;
          return (
            <button
              key={folder ?? 'all'}
              type="button"
              className={`btn-ghost btn-sm${isActive ? ' btn-active' : ''}`}
              style={{ fontWeight: isActive ? 700 : 400, opacity: isActive ? 1 : 0.7, borderColor: isActive ? 'var(--ox-accent)' : 'transparent' }}
              onClick={() => { setActiveFolder(folder); setSelectedID(null); }}
            >
              {icon} {label}
              {count > 0 && <span style={{ marginLeft: 4, fontSize: 11, color: 'var(--ox-fg-muted)' }}>({count})</span>}
            </button>
          );
        })}
      </div>

      {composing && (
        <ComposeForm
          init={replyInit}
          onSent={() => {
            setComposing(false);
            setReplyInit(undefined);
            void qc.invalidateQueries({ queryKey: ['messages'] });
            toast.show('success', 'Отправлено');
          }}
          onCancel={() => { setComposing(false); setReplyInit(undefined); }}
        />
      )}

      {!composing && msgs.length === 0 && (
        <div style={{ color: 'var(--ox-fg-dim)', fontSize: 14, padding: '16px 0' }}>📭 Нет сообщений</div>
      )}

      {!composing && msgs.length > 0 && (
        <div style={{ display: 'grid', gridTemplateColumns: selected ? '1fr 1.5fr' : '1fr', gap: 16 }}>
          {/* Message list */}
          <div className="ox-panel" style={{ overflow: 'hidden' }}>
            {msgs.map((m) => {
              const unread = !m.read_at;
              const isSelected = selectedID === m.id;
              return (
                <div
                  key={m.id}
                  onClick={() => onSelect(m)}
                  style={{
                    display: 'flex', alignItems: 'center', gap: 10,
                    padding: '10px 14px',
                    cursor: 'pointer',
                    borderBottom: '1px solid var(--ox-border)',
                    background: isSelected ? 'var(--ox-bg-active)' : unread ? 'rgba(99,217,255,0.04)' : 'transparent',
                    transition: 'background var(--ox-tr)',
                  }}
                >
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div style={{ fontWeight: unread ? 700 : 400, fontSize: 13, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                      {m.subject}
                    </div>
                    <div style={{ fontSize: 11, color: 'var(--ox-fg-muted)', marginTop: 2 }}>
                      {m.from_username || 'Система'} · {new Date(m.created_at).toLocaleString('ru-RU')}
                    </div>
                  </div>
                  {unread && <div style={{ width: 6, height: 6, borderRadius: '50%', background: 'var(--ox-accent)', flexShrink: 0 }} />}
                  <button
                    type="button"
                    className="btn-ghost btn-sm btn-icon"
                    style={{ flexShrink: 0, opacity: 0.5 }}
                    disabled={del.isPending}
                    onClick={(e) => { e.stopPropagation(); del.mutate(m.id); }}
                    title="Удалить"
                  >
                    ✕
                  </button>
                </div>
              );
            })}
          </div>

          {/* Message detail */}
          {selected && (
            <div className="ox-panel" style={{ padding: '16px 20px' }}>
              <MessageDetail message={selected} onReply={onReply} onFleetMission={onFleetMission} />
            </div>
          )}
        </div>
      )}

      {confirmDelAll && (
        <Confirm
          title="Удалить сообщения"
          message={activeFolder === null ? 'Удалить все сообщения?' : 'Удалить все сообщения в этой папке?'}
          confirmLabel="Удалить"
          danger
          onConfirm={() => { setConfirmDelAll(false); delAll.mutate(); }}
          onCancel={() => setConfirmDelAll(false)}
        />
      )}
    </div>
  );
}

function ComposeForm({ init, onSent, onCancel }: { init?: ReplyInit | undefined; onSent: () => void; onCancel: () => void }) {
  const [to, setTo] = useState(init?.to ?? '');
  const [subject, setSubject] = useState(init?.subject ?? '');
  const [body, setBody] = useState('');
  const [error, setError] = useState('');

  const send = useMutation({
    mutationFn: () => api.post<void>('/api/messages', { to, subject, body }),
    onSuccess: onSent,
    onError: (e) => setError(e instanceof Error ? e.message : 'Ошибка'),
  });

  return (
    <div className="ox-panel" style={{ padding: '16px 20px', display: 'flex', flexDirection: 'column', gap: 12 }}>
      <div style={{ fontSize: 13, fontWeight: 700 }}>Написать сообщение</div>
      <div>
        <label style={{ fontSize: 12, color: 'var(--ox-fg-dim)', display: 'block', marginBottom: 4 }}>Кому</label>
        <input value={to} onChange={(e) => setTo(e.target.value)} placeholder="имя игрока" style={{ width: '100%' }} />
      </div>
      <div>
        <label style={{ fontSize: 12, color: 'var(--ox-fg-dim)', display: 'block', marginBottom: 4 }}>Тема</label>
        <input value={subject} onChange={(e) => setSubject(e.target.value)} placeholder="тема сообщения" style={{ width: '100%' }} />
      </div>
      <div>
        <textarea value={body} onChange={(e) => setBody(e.target.value)} rows={5} style={{ width: '100%', boxSizing: 'border-box' }} placeholder="текст сообщения…" />
      </div>
      {error && <div className="ox-alert ox-alert-danger">{error}</div>}
      <div style={{ display: 'flex', gap: 8 }}>
        <button type="button" className="btn" disabled={send.isPending || !to || !subject} onClick={() => send.mutate()}>
          {send.isPending ? '…' : 'Отправить'}
        </button>
        <button type="button" className="btn-ghost" onClick={onCancel}>Отмена</button>
      </div>
    </div>
  );
}

function MessageDetail({ message, onReply, onFleetMission }: { message: Message; onReply: (m: Message) => void; onFleetMission?: FleetMissionCb }) {
  const report = useQuery({
    queryKey: ['battle-report', message.battle_report_id],
    queryFn: () => api.get<BattleReportFull>(`/api/battle-reports/${message.battle_report_id}`),
    enabled: !!message.battle_report_id,
  });
  const espionage = useQuery({
    queryKey: ['espionage-report', message.espionage_report_id],
    queryFn: () => api.get<EspionageReportFull>(`/api/espionage-reports/${message.espionage_report_id}`),
    enabled: !!message.espionage_report_id,
  });
  const expedition = useQuery({
    queryKey: ['expedition-report', message.expedition_report_id],
    queryFn: () => api.get<ExpeditionReportFull>(`/api/expedition-reports/${message.expedition_report_id}`),
    enabled: !!message.expedition_report_id,
  });

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
      <h3 style={{ margin: 0, fontSize: 15 }}>{message.subject}</h3>
      <div style={{ fontSize: 12, color: 'var(--ox-fg-muted)' }}>
        От: <span style={{ color: 'var(--ox-fg-dim)' }}>{message.from_username || 'Система'}</span>
        {' · '}
        {new Date(message.created_at).toLocaleString('ru-RU')}
      </div>
      <div style={{ fontSize: 14, lineHeight: 1.6, whiteSpace: 'pre-wrap' }}>{message.body}</div>
      {message.from_user_id && (
        <div>
          <button type="button" className="btn-ghost btn-sm" onClick={() => onReply(message)}>
            ↩ Ответить
          </button>
        </div>
      )}

      {message.battle_report_id && (
        <div style={{ borderTop: '1px solid var(--ox-border)', paddingTop: 12 }}>
          {report.isLoading && <div style={{ color: 'var(--ox-fg-muted)' }}>Загрузка отчёта…</div>}
          {report.data && <BattleReportView data={report.data} onFleetMission={onFleetMission} />}
        </div>
      )}
      {message.espionage_report_id && (
        <div style={{ borderTop: '1px solid var(--ox-border)', paddingTop: 12 }}>
          {espionage.isLoading && <div style={{ color: 'var(--ox-fg-muted)' }}>Загрузка…</div>}
          {espionage.data && <EspionageReportView data={espionage.data} />}
        </div>
      )}
      {message.expedition_report_id && (
        <div style={{ borderTop: '1px solid var(--ox-border)', paddingTop: 12 }}>
          {expedition.isLoading && <div style={{ color: 'var(--ox-fg-muted)' }}>Загрузка…</div>}
          {expedition.data && <ExpeditionReportView data={expedition.data} />}
        </div>
      )}
    </div>
  );
}

function ExpeditionReportView({ data }: { data: ExpeditionReportFull }) {
  const outcomeText: Record<string, string> = {
    resources: 'Найдены ресурсы', artefact: 'Найден артефакт',
    extra_planet: 'Обнаружена новая планета', pirates: 'Столкновение с пиратами',
    loss: 'Потери', nothing: 'Ничего не найдено',
  };
  return (
    <div>
      <div style={{ fontWeight: 700, marginBottom: 8 }}>🌌 Отчёт экспедиции</div>
      <div style={{ marginBottom: 8 }}>Результат: <b>{outcomeText[data.outcome] ?? data.outcome}</b></div>
      <pre style={{ background: 'var(--ox-bg-hover)', padding: 10, borderRadius: 6, fontSize: 12, overflow: 'auto' }}>
        {JSON.stringify(data.report, null, 2)}
      </pre>
    </div>
  );
}

function EspionageReportView({ data }: { data: EspionageReportFull }) {
  const r = data.report;
  return (
    <div>
      <div style={{ fontWeight: 700, marginBottom: 8 }}>🔭 Шпионский отчёт</div>
      <div style={{ fontSize: 13, marginBottom: 4 }}>Соотношение: {r.ratio} · Зонды: {r.probes}</div>
      <div style={{ fontSize: 13, marginBottom: 8 }}>Ресурсы: {r.metal} M / {r.silicon} Si / {r.hydrogen} H</div>
      {r.ships && <UnitMapBlock title="Корабли" data={r.ships} />}
      {r.defense && <UnitMapBlock title="Оборона" data={r.defense} />}
      {r.buildings && <UnitMapBlock title="Здания" data={r.buildings} />}
    </div>
  );
}

function UnitMapBlock({ title, data }: { title: string; data: Record<string, number> }) {
  const entries = Object.entries(data);
  if (entries.length === 0) return null;
  return (
    <div style={{ marginBottom: 8 }}>
      <div style={{ fontSize: 12, fontWeight: 700, color: 'var(--ox-fg-dim)', marginBottom: 4 }}>{title}</div>
      <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
        {entries.map(([id, n]) => (
          <span key={id} style={{ fontSize: 12, background: 'var(--ox-bg-hover)', padding: '2px 8px', borderRadius: 4 }}>
            #{id}: {n}
          </span>
        ))}
      </div>
    </div>
  );
}

function BattleReportView({ data, onFleetMission }: { data: BattleReportFull; onFleetMission?: FleetMissionCb }) {
  const r = data.report;
  const winnerText = r.winner === 'attackers' ? '⚔️ Победа атакующих'
    : r.winner === 'defenders' ? '🛡 Победа защитников'
    : '⚖️ Ничья';
  const hasDst = data.dst_galaxy != null && data.dst_system != null && data.dst_position != null;
  return (
    <div>
      <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 8 }}>
        <div style={{ fontWeight: 700 }}>⚔️ Боевой отчёт</div>
        {hasDst && onFleetMission && (
          <button
            type="button"
            className="btn btn-sm"
            style={{ fontSize: 11 }}
            onClick={() => onFleetMission(data.dst_galaxy!, data.dst_system!, data.dst_position!, false, 10)}
          >
            ⚔️ Атаковать [{data.dst_galaxy}:{data.dst_system}:{data.dst_position}]
          </button>
        )}
      </div>
      <div style={{ fontSize: 13, marginBottom: 4 }}>
        <span style={{ fontWeight: 700 }}>{winnerText}</span> · Раундов: {r.rounds}
      </div>
      <div style={{ fontSize: 13, marginBottom: 12 }}>
        Добыча: {data.loot_metal.toLocaleString('ru-RU')} 🟠 / {data.loot_silicon.toLocaleString('ru-RU')} 💎 / {data.loot_hydrogen.toLocaleString('ru-RU')} 💧
      </div>

      {r.rounds_trace && r.rounds_trace.length > 0 && (
        <div style={{ marginBottom: 12, overflowX: 'auto' }}>
          <table className="ox-table" style={{ margin: 0 }}>
            <thead><tr><th>#</th><th>Атакующих</th><th>Защитников</th></tr></thead>
            <tbody>
              {r.rounds_trace.map((row) => (
                <tr key={row.index}>
                  <td>{row.index + 1}</td>
                  <td className="num">{row.attackers_alive}</td>
                  <td className="num">{row.defenders_alive}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {r.attackers?.[0] && <SideLosses title="Атакующие" side={r.attackers[0]} />}
      {r.defenders?.[0] && <SideLosses title="Защитники" side={r.defenders[0]} />}
    </div>
  );
}

function SideLosses({ title, side }: { title: string; side: SideResult }) {
  return (
    <div style={{ marginBottom: 12 }}>
      <div style={{ fontSize: 13, fontWeight: 700, marginBottom: 4 }}>{title}</div>
      <div style={{ fontSize: 12, color: 'var(--ox-fg-dim)', marginBottom: 6 }}>
        Потери: {side.lost_metal} M / {side.lost_silicon} Si / {side.lost_hydrogen} H
      </div>
      <table className="ox-table" style={{ margin: 0 }}>
        <thead><tr><th>Юнит</th><th>Было</th><th>Стало</th><th>Повреждено</th></tr></thead>
        <tbody>
          {side.units.map((u) => (
            <tr key={u.unit_id}>
              <td>#{u.unit_id}</td>
              <td className="num">{u.quantity_start}</td>
              <td className="num">{u.quantity_end}</td>
              <td className="num">{u.damaged_end ? `${u.damaged_end} (${Math.round(u.shell_percent_end ?? 0)}%)` : '—'}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
