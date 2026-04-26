import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';
import { nameOf } from '@/api/catalog';
import { Confirm } from '@/ui/Confirm';
import { useToast } from '@/ui/Toast';
import { useTranslation } from '@/i18n/i18n';

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

type FolderKey = number | null | 'sent';
const FOLDER_DEFS: { folder: FolderKey; tkey: string; icon: string }[] = [
  { folder: null,   tkey: 'folderAll',       icon: '📬' },
  { folder: 1,      tkey: 'folderPersonal',  icon: '✉️' },
  { folder: 2,      tkey: 'folderBattle',    icon: '⚔️' },
  { folder: 3,      tkey: 'folderSpy',       icon: '🔭' },
  { folder: 4,      tkey: 'folderExpedition',icon: '🌌' },
  { folder: 11,     tkey: 'folderPhalanx',   icon: '📡' },
  { folder: 6,      tkey: 'folderAlliance',  icon: '🤝' },
  { folder: 7,      tkey: 'folderArtefacts', icon: '💎' },
  { folder: 8,      tkey: 'folderCredits',   icon: '💳' },
  { folder: 13,     tkey: 'folderSystem',    icon: '⚙️' },
  { folder: 'sent', tkey: 'folderSent',      icon: '📤' },
];

type FleetMissionCb = (g: number, s: number, pos: number, isMoon: boolean, mission: number) => void;

export function MessagesScreen({ onFleetMission }: { onFleetMission?: FleetMissionCb }) {
  const { t } = useTranslation('messagesUi');
  const qc = useQueryClient();
  const toast = useToast();
  const [selectedID, setSelectedID] = useState<string | null>(null);
  const [composing, setComposing] = useState(false);
  const [replyInit, setReplyInit] = useState<ReplyInit | undefined>(undefined);
  const [activeFolder, setActiveFolder] = useState<FolderKey>(null);
  const [confirmDelAll, setConfirmDelAll] = useState(false);

  const list = useQuery({
    queryKey: ['messages'],
    queryFn: () => api.get<{ messages: Message[] | null }>('/api/messages'),
    refetchInterval: 10000,
    enabled: activeFolder !== 'sent',
  });

  const sentList = useQuery({
    queryKey: ['messages', 'sent'],
    queryFn: () => api.get<{ messages: Message[] | null }>('/api/messages/sent'),
    refetchInterval: 30000,
    enabled: activeFolder === 'sent',
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
      toast.show('info', t('toastDeleted'));
    },
  });

  const delAll = useMutation({
    mutationFn: () => {
      const qs = typeof activeFolder === 'number' ? `?folder=${activeFolder}` : '';
      return api.delete<void>(`/api/messages${qs}`);
    },
    onSuccess: () => {
      setSelectedID(null);
      void qc.invalidateQueries({ queryKey: ['messages'] });
      void qc.invalidateQueries({ queryKey: ['messages-unread'] });
      toast.show('info', t('toastDeletedAll'));
    },
  });

  const allMsgs = list.data?.messages ?? [];
  const sentMsgs = sentList.data?.messages ?? [];
  const msgs = activeFolder === 'sent'
    ? sentMsgs
    : activeFolder === null
    ? allMsgs
    : allMsgs.filter((m) => m.folder === activeFolder);
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
          {t('title')} {unreadCount > 0 && <span style={{ fontSize: 16, color: 'var(--ox-accent)' }}>{t('unread', { count: String(unreadCount) })}</span>}
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
              🗑 {t('deleteAll')}
            </button>
          )}
          <button type="button" className="btn btn-sm" onClick={() => { setComposing(true); setSelectedID(null); }}>
            ✉ {t('compose')}
          </button>
        </div>
      </div>

      {/* Folder tabs */}
      <div style={{ display: 'flex', gap: 4, flexWrap: 'wrap' }}>
        {FOLDER_DEFS.map(({ folder, tkey, icon }) => {
          const label = t(tkey);
          const isActive = activeFolder === folder;
          const count = folder === null
            ? allMsgs.length
            : folder === 'sent'
            ? (sentMsgs.length || undefined)
            : allMsgs.filter((m) => m.folder === folder).length;
          return (
            <button
              key={folder === null ? 'all' : String(folder)}
              type="button"
              className={`btn-ghost btn-sm${isActive ? ' btn-active' : ''}`}
              style={{ fontWeight: isActive ? 700 : 400, opacity: isActive ? 1 : 0.7, borderColor: isActive ? 'var(--ox-accent)' : 'transparent' }}
              onClick={() => { setActiveFolder(folder); setSelectedID(null); }}
            >
              {icon} {label}
              {typeof count === 'number' && count > 0 && <span style={{ marginLeft: 4, fontSize: 13, color: 'var(--ox-fg-muted)' }}>({count})</span>}
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
            toast.show('success', t('toastSent'));
          }}
          onCancel={() => { setComposing(false); setReplyInit(undefined); }}
        />
      )}

      {!composing && msgs.length === 0 && (
        <div style={{ color: 'var(--ox-fg-dim)', fontSize: 16, padding: '16px 0' }}>📭 {t('empty')}</div>
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
                    <div style={{ fontWeight: unread ? 700 : 400, fontSize: 15, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                      {m.subject}
                    </div>
                    <div style={{ fontSize: 13, color: 'var(--ox-fg-muted)', marginTop: 2 }}>
                      {m.from_username || t('senderSystem')} · {new Date(m.created_at).toLocaleString('ru-RU')}
                    </div>
                  </div>
                  {unread && <div style={{ width: 6, height: 6, borderRadius: '50%', background: 'var(--ox-accent)', flexShrink: 0 }} />}
                  <button
                    type="button"
                    className="btn-ghost btn-sm btn-icon"
                    style={{ flexShrink: 0, opacity: 0.5 }}
                    disabled={del.isPending}
                    onClick={(e) => { e.stopPropagation(); del.mutate(m.id); }}
                    title={t('deleteTitle')}
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
          title={t('confirmDelAllTitle')}
          message={activeFolder === null ? t('confirmDelAllAll') : t('confirmDelAllFolder')}
          confirmLabel={t('confirmDelLabel')}
          danger
          onConfirm={() => { setConfirmDelAll(false); delAll.mutate(); }}
          onCancel={() => setConfirmDelAll(false)}
        />
      )}
    </div>
  );
}

function ComposeForm({ init, onSent, onCancel }: { init?: ReplyInit | undefined; onSent: () => void; onCancel: () => void }) {
  const { t } = useTranslation('messagesUi');
  const [to, setTo] = useState(init?.to ?? '');
  const [subject, setSubject] = useState(init?.subject ?? '');
  const [body, setBody] = useState('');
  const [error, setError] = useState('');

  const send = useMutation({
    mutationFn: () => api.post<void>('/api/messages', { to, subject, body }),
    onSuccess: onSent,
    onError: (e) => setError(e instanceof Error ? e.message : t('composeError')),
  });

  return (
    <div className="ox-panel" style={{ padding: '16px 20px', display: 'flex', flexDirection: 'column', gap: 12 }}>
      <div style={{ fontSize: 15, fontWeight: 700 }}>{t('composeTitle')}</div>
      <div>
        <label style={{ fontSize: 14, color: 'var(--ox-fg-dim)', display: 'block', marginBottom: 4 }}>{t('composeTo')}</label>
        <input value={to} onChange={(e) => setTo(e.target.value)} placeholder={t('composeToPh')} style={{ width: '100%' }} />
      </div>
      <div>
        <label style={{ fontSize: 14, color: 'var(--ox-fg-dim)', display: 'block', marginBottom: 4 }}>{t('composeSubject')}</label>
        <input value={subject} onChange={(e) => setSubject(e.target.value)} placeholder={t('composeSubjectPh')} style={{ width: '100%' }} />
      </div>
      <div>
        <textarea value={body} onChange={(e) => setBody(e.target.value)} rows={5} style={{ width: '100%', boxSizing: 'border-box' }} placeholder={t('composeBodyPh')} />
      </div>
      {error && <div className="ox-alert ox-alert-danger">{error}</div>}
      <div style={{ display: 'flex', gap: 8 }}>
        <button type="button" className="btn" disabled={send.isPending || !to || !subject} onClick={() => send.mutate()}>
          {send.isPending ? '…' : t('composeSend')}
        </button>
        <button type="button" className="btn-ghost" onClick={onCancel}>{t('composeCancel')}</button>
      </div>
    </div>
  );
}

function MessageDetail({ message, onReply, onFleetMission }: { message: Message; onReply: (m: Message) => void; onFleetMission?: FleetMissionCb }) {
  const { t } = useTranslation('messagesUi');
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
      <div style={{ fontSize: 14, color: 'var(--ox-fg-muted)' }}>
        {t('from')} <span style={{ color: 'var(--ox-fg-dim)' }}>{message.from_username || t('senderSystem')}</span>
        {' · '}
        {new Date(message.created_at).toLocaleString('ru-RU')}
      </div>
      <div style={{ fontSize: 16, lineHeight: 1.6, whiteSpace: 'pre-wrap' }}>{message.body}</div>
      {message.from_user_id && (
        <div>
          <button type="button" className="btn-ghost btn-sm" onClick={() => onReply(message)}>
            {t('reply')}
          </button>
        </div>
      )}

      {message.battle_report_id && (
        <div style={{ borderTop: '1px solid var(--ox-border)', paddingTop: 12 }}>
          {report.isLoading && <div style={{ color: 'var(--ox-fg-muted)' }}>{t('loading')}</div>}
          {report.data && <BattleReportView data={report.data} onFleetMission={onFleetMission} />}
        </div>
      )}
      {message.espionage_report_id && (
        <div style={{ borderTop: '1px solid var(--ox-border)', paddingTop: 12 }}>
          {espionage.isLoading && <div style={{ color: 'var(--ox-fg-muted)' }}>{t('loadingShort')}</div>}
          {espionage.data && <EspionageReportView data={espionage.data} />}
        </div>
      )}
      {message.expedition_report_id && (
        <div style={{ borderTop: '1px solid var(--ox-border)', paddingTop: 12 }}>
          {expedition.isLoading && <div style={{ color: 'var(--ox-fg-muted)' }}>{t('loadingShort')}</div>}
          {expedition.data && <ExpeditionReportView data={expedition.data} />}
        </div>
      )}
    </div>
  );
}

function ExpeditionReportView({ data }: { data: ExpeditionReportFull }) {
  const { t } = useTranslation('messagesUi');
  const outcomeText: Record<string, string> = {
    resources: t('expedResources'), artefact: t('expedArtefact'),
    extra_planet: t('expedExtraPlanet'), pirates: t('expedPirates'),
    loss: t('expedLoss'), nothing: t('expedNothing'),
  };
  return (
    <div>
      <div style={{ fontWeight: 700, marginBottom: 8 }}>{t('expedTitle')}</div>
      <div style={{ marginBottom: 8 }}>{t('expedResult')} <b>{outcomeText[data.outcome] ?? data.outcome}</b></div>
      <pre style={{ background: 'var(--ox-bg-hover)', padding: 10, borderRadius: 6, fontSize: 14, overflow: 'auto' }}>
        {JSON.stringify(data.report, null, 2)}
      </pre>
    </div>
  );
}

function EspionageReportView({ data }: { data: EspionageReportFull }) {
  const { t } = useTranslation('messagesUi');
  const r = data.report;
  return (
    <div>
      <div style={{ fontWeight: 700, marginBottom: 8 }}>{t('spyTitle')}</div>
      <div style={{ fontSize: 15, marginBottom: 4 }}>{t('spyRatio')}: {r.ratio} · {t('spyProbes')}: {r.probes}</div>
      <div style={{ fontSize: 15, marginBottom: 8 }}>
        🟠 {r.metal.toLocaleString('ru-RU')} · 💎 {r.silicon.toLocaleString('ru-RU')} · 💧 {r.hydrogen.toLocaleString('ru-RU')}
      </div>
      {r.ships && <UnitMapBlock title={t('spyShips')} data={r.ships} />}
      {r.defense && <UnitMapBlock title={t('spyDefense')} data={r.defense} />}
      {r.buildings && <UnitMapBlock title={t('spyBuildings')} data={r.buildings} />}
    </div>
  );
}

function UnitMapBlock({ title, data }: { title: string; data: Record<string, number> }) {
  const { t: ti } = useTranslation('info');
  const entries = Object.entries(data);
  if (entries.length === 0) return null;
  return (
    <div style={{ marginBottom: 8 }}>
      <div style={{ fontSize: 14, fontWeight: 700, color: 'var(--ox-fg-dim)', marginBottom: 4 }}>{title}</div>
      <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
        {entries.map(([id, n]) => (
          <span key={id} style={{ fontSize: 14, background: 'var(--ox-bg-hover)', padding: '2px 8px', borderRadius: 4 }}>
            {nameOf(Number(id), ti)}: {n.toLocaleString('ru-RU')}
          </span>
        ))}
      </div>
    </div>
  );
}

function BattleReportView({ data, onFleetMission }: { data: BattleReportFull; onFleetMission?: FleetMissionCb }) {
  const { t } = useTranslation('messagesUi');
  const r = data.report;
  const isAttackerWin = r.winner === 'attackers';
  const isDefenderWin = r.winner === 'defenders';
  const hasDst = data.dst_galaxy != null && data.dst_system != null && data.dst_position != null;
  const hasLoot = data.loot_metal > 0 || data.loot_silicon > 0 || data.loot_hydrogen > 0;

  const winnerColor = isAttackerWin ? 'var(--ox-accent)' : isDefenderWin ? 'var(--ox-success)' : 'var(--ox-warning)';
  const winnerText = isAttackerWin ? t('winnerAttackers')
    : isDefenderWin ? t('winnerDefenders')
    : t('winnerDraw');

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
      {/* Header */}
      <div style={{ display: 'flex', alignItems: 'center', gap: 12, flexWrap: 'wrap' }}>
        <div style={{ fontSize: 16, fontWeight: 700, color: winnerColor }}>{winnerText}</div>
        <span style={{ fontSize: 14, color: 'var(--ox-fg-muted)' }}>{t('rounds')} {r.rounds}</span>
        {hasDst && (
          <span style={{ fontSize: 14, color: 'var(--ox-fg-muted)', fontFamily: 'var(--ox-mono)' }}>
            [{data.dst_galaxy}:{data.dst_system}:{data.dst_position}]
          </span>
        )}
        {hasDst && onFleetMission && (
          <button
            type="button"
            className="btn btn-sm"
            style={{ fontSize: 13 }}
            onClick={() => onFleetMission(data.dst_galaxy!, data.dst_system!, data.dst_position!, false, 10)}
          >
            {t('attackBtn')}
          </button>
        )}
      </div>

      {/* Participants */}
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 8 }}>
        <div style={{ padding: '8px 12px', borderRadius: 6, border: '1px solid var(--ox-accent)', background: 'rgba(99,217,255,0.05)' }}>
          <div style={{ fontSize: 13, fontWeight: 700, color: 'var(--ox-accent)', letterSpacing: '0.08em', marginBottom: 4 }}>{t('labelAttacker')}</div>
          <div style={{ fontSize: 15 }}>{data.attacker_username || '—'}</div>
        </div>
        <div style={{ padding: '8px 12px', borderRadius: 6, border: '1px solid var(--ox-success)', background: 'rgba(80,200,120,0.05)' }}>
          <div style={{ fontSize: 13, fontWeight: 700, color: 'var(--ox-success)', letterSpacing: '0.08em', marginBottom: 4 }}>{t('labelDefender')}</div>
          <div style={{ fontSize: 15 }}>{data.defender_username || '—'}</div>
        </div>
      </div>

      {/* Loot */}
      {hasLoot && (
        <div style={{ padding: '8px 14px', background: 'var(--ox-surface)', borderRadius: 6, border: '1px solid var(--ox-border)' }}>
          <span style={{ fontSize: 13, fontWeight: 700, color: 'var(--ox-fg-muted)', marginRight: 12 }}>{t('labelLoot')}</span>
          {data.loot_metal > 0 && <span style={{ marginRight: 10, fontSize: 15 }}>🟠 {data.loot_metal.toLocaleString('ru-RU')}</span>}
          {data.loot_silicon > 0 && <span style={{ marginRight: 10, fontSize: 15 }}>💎 {data.loot_silicon.toLocaleString('ru-RU')}</span>}
          {data.loot_hydrogen > 0 && <span style={{ fontSize: 15 }}>💧 {data.loot_hydrogen.toLocaleString('ru-RU')}</span>}
        </div>
      )}

      {/* Sides */}
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
        {r.attackers?.[0] && <SideLosses title={t('sidesAttackers')} side={r.attackers[0]} accentColor="var(--ox-accent)" />}
        {r.defenders?.[0] && <SideLosses title={t('sidesDefenders')} side={r.defenders[0]} accentColor="var(--ox-success)" />}
      </div>

      {/* Rounds trace */}
      {r.rounds_trace && r.rounds_trace.length > 0 && (
        <details style={{ fontSize: 14 }}>
          <summary style={{ cursor: 'pointer', color: 'var(--ox-fg-muted)', marginBottom: 6 }}>{t('roundsTrace')}</summary>
          <div style={{ overflowX: 'auto' }}>
            <table className="ox-table" style={{ margin: 0 }}>
              <thead><tr><th>{t('thRound')}</th><th style={{ color: 'var(--ox-accent)' }}>{t('thAttackers')}</th><th style={{ color: 'var(--ox-success)' }}>{t('thDefenders')}</th></tr></thead>
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
        </details>
      )}
    </div>
  );
}

function SideLosses({ title, side, accentColor }: { title: string; side: SideResult; accentColor: string }) {
  const { t } = useTranslation('messagesUi');
  const { t: ti } = useTranslation('info');
  const lost = side.lost_metal + side.lost_silicon + side.lost_hydrogen;
  return (
    <div style={{ padding: '8px 12px', borderRadius: 6, border: `1px solid ${accentColor}`, background: `${accentColor}0d` }}>
      <div style={{ fontSize: 13, fontWeight: 700, color: accentColor, letterSpacing: '0.08em', marginBottom: 6 }}>{title.toUpperCase()}</div>
      {lost > 0 && (
        <div style={{ fontSize: 13, color: 'var(--ox-danger)', marginBottom: 6, fontFamily: 'var(--ox-mono)' }}>
          {t('sideLosses')} {side.lost_metal > 0 && `🟠${side.lost_metal.toLocaleString('ru-RU')} `}
          {side.lost_silicon > 0 && `💎${side.lost_silicon.toLocaleString('ru-RU')} `}
          {side.lost_hydrogen > 0 && `💧${side.lost_hydrogen.toLocaleString('ru-RU')}`}
        </div>
      )}
      <div style={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
        {side.units.map((u) => {
          const lost_ = u.quantity_start - u.quantity_end;
          return (
            <div key={u.unit_id} style={{ display: 'flex', gap: 6, fontSize: 13, fontFamily: 'var(--ox-mono)', alignItems: 'center' }}>
              <span style={{ flex: 1, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', color: 'var(--ox-fg-dim)' }}>
                {nameOf(u.unit_id, ti)}
              </span>
              <span>{u.quantity_start} → {u.quantity_end}</span>
              {lost_ > 0 && <span style={{ color: 'var(--ox-danger)' }}>−{lost_}</span>}
            </div>
          );
        })}
      </div>
    </div>
  );
}
