import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';
import { useTranslation } from '@/i18n/i18n';

// MessagesScreen — inbox + inline battle-report detail.
//
// Структура:
//   [список строк inbox] — клик открывает detail справа/снизу.
//   Если у строки battle_report_id — подгружаем полный report и
//   рендерим BattleReportView; иначе просто body.
//
// Пагинация: серверный лимит 100, клиенту достаточно (см. §4
// service.go). Refetch каждые 10s для новых входящих.

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

const FOLDERS: { folder: number | null; label: string }[] = [
  { folder: null, label: 'Все' },
  { folder: 1, label: 'Личные' },
  { folder: 2, label: 'Бой' },
  { folder: 3, label: 'Шпионаж' },
  { folder: 4, label: 'Экспедиции' },
  { folder: 13, label: 'Система' },
];

export function MessagesScreen() {
  const { t, tf } = useTranslation();
  const qc = useQueryClient();
  const [selectedID, setSelectedID] = useState<string | null>(null);
  const [composing, setComposing] = useState(false);
  const [replyInit, setReplyInit] = useState<ReplyInit | undefined>(undefined);
  const [activeFolder, setActiveFolder] = useState<number | null>(null);

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
      void qc.invalidateQueries({ queryKey: ['messages', 'unread-count'] });
    },
  });

  const allMsgs = list.data?.messages ?? [];
  const msgs = activeFolder === null ? allMsgs : allMsgs.filter((m) => m.folder === activeFolder);
  const selected = msgs.find((m) => m.id === selectedID) ?? null;

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
    <section>
      <h2>{tf('global', 'MENU_MESSAGES', 'Сообщения')}</h2>

      <div style={{ marginBottom: 8, display: 'flex', gap: 4, flexWrap: 'wrap' }}>
        {FOLDERS.map(({ folder, label }) => (
          <button
            key={folder ?? 'all'}
            type="button"
            onClick={() => { setActiveFolder(folder); setSelectedID(null); }}
            style={{ fontWeight: activeFolder === folder ? 700 : 400, opacity: activeFolder === folder ? 1 : 0.6 }}
          >
            {label}
          </button>
        ))}
      </div>

      <div style={{ marginBottom: 12 }}>
        <button type="button" onClick={() => { setComposing(true); setSelectedID(null); }}>
          {tf('Main', 'MSG_COMPOSE', '✉ Написать')}
        </button>
      </div>

      {list.isLoading && <p>…</p>}
      {list.error && (
        <p className="ox-error">
          {t('global', 'ERROR')}:{' '}
          {list.error instanceof Error ? list.error.message : ''}
        </p>
      )}

      {composing && (
        <ComposeForm
          init={replyInit}
          onSent={() => {
            setComposing(false);
            setReplyInit(undefined);
            void qc.invalidateQueries({ queryKey: ['messages'] });
          }}
          onCancel={() => { setComposing(false); setReplyInit(undefined); }}
        />
      )}

      {!composing && msgs.length === 0 && (
        <p>{tf('Main', 'INBOX_EMPTY', 'Нет сообщений.')}</p>
      )}

      {!composing && msgs.length > 0 && (
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 2fr', gap: 16 }}>
          <table className="ox-table">
            <thead>
              <tr>
                <th>{tf('Main', 'MSG_FROM', 'От')}</th>
                <th>{tf('Main', 'MSG_SUBJECT', 'Тема')}</th>
                <th>{tf('Main', 'MSG_WHEN', 'Когда')}</th>
                <th />
              </tr>
            </thead>
            <tbody>
              {msgs.map((m) => {
                const unread = !m.read_at;
                return (
                  <tr
                    key={m.id}
                    onClick={() => onSelect(m)}
                    style={{
                      cursor: 'pointer',
                      fontWeight: unread ? 700 : 400,
                      background: selectedID === m.id ? 'rgba(255,255,255,0.05)' : undefined,
                    }}
                  >
                    <td>{m.from_username || '—'}</td>
                    <td>{m.subject}</td>
                    <td>{new Date(m.created_at).toLocaleString('ru-RU')}</td>
                    <td>
                      <button
                        type="button"
                        disabled={del.isPending}
                        onClick={(e) => { e.stopPropagation(); del.mutate(m.id); }}
                        title={tf('Main', 'MSG_DELETE', 'Удалить')}
                        style={{ opacity: 0.6 }}
                      >
                        ✕
                      </button>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
          <div>{selected ? <MessageDetail message={selected} onReply={onReply} /> : <p>—</p>}</div>
        </div>
      )}
    </section>
  );
}

function ComposeForm({ init, onSent, onCancel }: { init?: ReplyInit | undefined; onSent: () => void; onCancel: () => void }) {
  const { tf } = useTranslation();
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
    <div style={{ marginBottom: 16, padding: 12, border: '1px solid rgba(255,255,255,0.15)', borderRadius: 4 }}>
      <h3 style={{ marginTop: 0 }}>{tf('Main', 'MSG_COMPOSE', 'Написать сообщение')}</h3>
      <div style={{ marginBottom: 8 }}>
        <label>
          {tf('Main', 'MSG_TO', 'Кому')}:{' '}
          <input
            value={to}
            onChange={(e) => setTo(e.target.value)}
            placeholder={tf('Main', 'MSG_TO_PLACEHOLDER', 'имя игрока')}
            style={{ width: 200 }}
          />
        </label>
      </div>
      <div style={{ marginBottom: 8 }}>
        <label>
          {tf('Main', 'MSG_SUBJECT', 'Тема')}:{' '}
          <input
            value={subject}
            onChange={(e) => setSubject(e.target.value)}
            placeholder={tf('Main', 'MSG_SUBJECT_PLACEHOLDER', 'тема')}
            style={{ width: 300 }}
          />
        </label>
      </div>
      <div style={{ marginBottom: 8 }}>
        <textarea
          value={body}
          onChange={(e) => setBody(e.target.value)}
          rows={5}
          style={{ width: '100%', boxSizing: 'border-box' }}
          placeholder={tf('Main', 'MSG_BODY_PLACEHOLDER', 'текст сообщения…')}
        />
      </div>
      {error && <p className="ox-error">{error}</p>}
      <button type="button" disabled={send.isPending || !to || !subject} onClick={() => send.mutate()}>
        {tf('Main', 'MSG_SEND', 'Отправить')}
      </button>{' '}
      <button type="button" onClick={onCancel}>
        {tf('Main', 'CANCEL', 'Отмена')}
      </button>
    </div>
  );
}

function MessageDetail({ message, onReply }: { message: Message; onReply: (m: Message) => void }) {
  const { tf } = useTranslation();
  const report = useQuery({
    queryKey: ['battle-report', message.battle_report_id],
    queryFn: () =>
      api.get<BattleReportFull>(`/api/battle-reports/${message.battle_report_id}`),
    enabled: !!message.battle_report_id,
  });
  const espionage = useQuery({
    queryKey: ['espionage-report', message.espionage_report_id],
    queryFn: () =>
      api.get<EspionageReportFull>(`/api/espionage-reports/${message.espionage_report_id}`),
    enabled: !!message.espionage_report_id,
  });
  const expedition = useQuery({
    queryKey: ['expedition-report', message.expedition_report_id],
    queryFn: () =>
      api.get<ExpeditionReportFull>(`/api/expedition-reports/${message.expedition_report_id}`),
    enabled: !!message.expedition_report_id,
  });

  return (
    <div>
      <h3>{message.subject}</h3>
      <p>
        <b>{tf('Main', 'MSG_FROM', 'От')}:</b> {message.from_username || '—'}{' '}
        <span style={{ opacity: 0.6 }}>
          ({new Date(message.created_at).toLocaleString('ru-RU')})
        </span>
      </p>
      <p>{message.body}</p>

      {message.from_user_id && (
        <button type="button" onClick={() => onReply(message)} style={{ marginBottom: 12 }}>
          {tf('Main', 'MSG_REPLY', '↩ Ответить')}
        </button>
      )}

      {message.battle_report_id && (
        <>
          <hr />
          {report.isLoading && <p>…</p>}
          {report.error && (
            <p className="ox-error">
              {tf('global', 'ERROR', 'Ошибка')}:{' '}
              {report.error instanceof Error ? report.error.message : ''}
            </p>
          )}
          {report.data && <BattleReportView data={report.data} />}
        </>
      )}

      {message.espionage_report_id && (
        <>
          <hr />
          {espionage.isLoading && <p>…</p>}
          {espionage.error && (
            <p className="ox-error">
              {tf('global', 'ERROR', 'Ошибка')}:{' '}
              {espionage.error instanceof Error ? espionage.error.message : ''}
            </p>
          )}
          {espionage.data && <EspionageReportView data={espionage.data} />}
        </>
      )}

      {message.expedition_report_id && (
        <>
          <hr />
          {expedition.isLoading && <p>…</p>}
          {expedition.error && (
            <p className="ox-error">
              {tf('global', 'ERROR', 'Ошибка')}:{' '}
              {expedition.error instanceof Error ? expedition.error.message : ''}
            </p>
          )}
          {expedition.data && <ExpeditionReportView data={expedition.data} />}
        </>
      )}
    </div>
  );
}

function ExpeditionReportView({ data }: { data: ExpeditionReportFull }) {
  const { tf } = useTranslation();
  const outcomeText: Record<string, string> = {
    resources: tf('Main', 'EXP_RESOURCES', 'Найдены ресурсы'),
    artefact: tf('Main', 'EXP_ARTEFACT', 'Найден артефакт'),
    extra_planet: tf('Main', 'EXP_EXTRA_PLANET', 'Обнаружена новая планета'),
    pirates: tf('Main', 'EXP_PIRATES', 'Столкновение с пиратами'),
    loss: tf('Main', 'EXP_LOSS', 'Потери'),
    nothing: tf('Main', 'EXP_NOTHING', 'Ничего не найдено'),
  };
  return (
    <div>
      <h4>{tf('Main', 'EXPEDITION_REPORT', 'Отчёт экспедиции')}</h4>
      <p>
        <b>{tf('Main', 'EXP_OUTCOME', 'Результат')}:</b> {outcomeText[data.outcome] ?? data.outcome}
      </p>
      <pre
        style={{
          background: 'rgba(255,255,255,0.05)',
          padding: 8,
          borderRadius: 4,
          fontSize: 12,
        }}
      >
        {JSON.stringify(data.report, null, 2)}
      </pre>
    </div>
  );
}

function EspionageReportView({ data }: { data: EspionageReportFull }) {
  const { tf } = useTranslation();
  const r = data.report;
  return (
    <div>
      <h4>{tf('Main', 'SPY_REPORT', 'Шпионский отчёт')}</h4>
      <p>
        <b>{tf('Main', 'SPY_RATIO', 'Соотношение')}:</b> {r.ratio}{' · '}
        <b>{tf('Main', 'SPY_PROBES', 'Зонды')}:</b> {r.probes}
      </p>
      <p>
        <b>{tf('Main', 'RESOURCES', 'Ресурсы')}:</b>{' '}
        {r.metal} M / {r.silicon} Si / {r.hydrogen} H
      </p>
      {r.ships && (
        <UnitMapTable title={tf('Main', 'UNITS_SHIPS', 'Корабли')} data={r.ships} />
      )}
      {r.defense && (
        <UnitMapTable title={tf('Main', 'UNITS_DEFENSE', 'Оборона')} data={r.defense} />
      )}
      {r.buildings && (
        <UnitMapTable
          title={tf('global', 'MENU_CONSTRUCTIONS', 'Здания')}
          data={r.buildings}
        />
      )}
    </div>
  );
}

function UnitMapTable({ title, data }: { title: string; data: Record<string, number> }) {
  const entries = Object.entries(data);
  if (entries.length === 0) return null;
  return (
    <>
      <h5>{title}</h5>
      <table className="ox-table">
        <tbody>
          {entries.map(([id, n]) => (
            <tr key={id}>
              <td>#{id}</td>
              <td className="num">{n}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </>
  );
}

function BattleReportView({ data }: { data: BattleReportFull }) {
  const { tf } = useTranslation();
  const r = data.report;
  const winnerText =
    r.winner === 'attackers'
      ? tf('Main', 'BATTLE_WIN_ATT', 'Победа атакующих')
      : r.winner === 'defenders'
        ? tf('Main', 'BATTLE_WIN_DEF', 'Победа защитников')
        : tf('Main', 'BATTLE_DRAW', 'Ничья');
  return (
    <div>
      <h4>{tf('Main', 'BATTLE_REPORT', 'Боевой отчёт')}</h4>
      <p>
        <b>{tf('Main', 'BATTLE_WINNER', 'Победитель')}:</b> {winnerText}
        {' · '}
        <b>{tf('Main', 'BATTLE_ROUNDS', 'Раундов')}:</b> {r.rounds}
      </p>
      <p>
        <b>{tf('Main', 'BATTLE_LOOT', 'Добыча')}:</b>{' '}
        {data.loot_metal} M / {data.loot_silicon} Si / {data.loot_hydrogen} H
      </p>

      {r.rounds_trace && r.rounds_trace.length > 0 && (
        <>
          <h4>{tf('Main', 'BATTLE_ROUNDS_HEADER', 'Ход боя')}</h4>
          <table className="ox-table">
            <thead>
              <tr>
                <th>#</th>
                <th>{tf('Main', 'ATTACKERS_ALIVE', 'Атакующих')}</th>
                <th>{tf('Main', 'DEFENDERS_ALIVE', 'Защитников')}</th>
              </tr>
            </thead>
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
        </>
      )}

      {r.attackers && r.attackers[0] && (
        <SideLosses title={tf('Main', 'ATTACKERS', 'Атакующие')} side={r.attackers[0]} />
      )}
      {r.defenders && r.defenders[0] && (
        <SideLosses title={tf('Main', 'DEFENDERS', 'Защитники')} side={r.defenders[0]} />
      )}
    </div>
  );
}

function SideLosses({ title, side }: { title: string; side: SideResult }) {
  const { tf } = useTranslation();
  return (
    <>
      <h4>{title}</h4>
      <p>
        {tf('Main', 'BATTLE_LOSSES', 'Потери')}:{' '}
        <b>{side.lost_metal}</b> M / <b>{side.lost_silicon}</b> Si /{' '}
        <b>{side.lost_hydrogen}</b> H
      </p>
      <table className="ox-table">
        <thead>
          <tr>
            <th>{tf('Main', 'UNIT_ID', 'Юнит #')}</th>
            <th>{tf('Main', 'BEFORE', 'Было')}</th>
            <th>{tf('Main', 'AFTER', 'Стало')}</th>
            <th>{tf('Main', 'DAMAGED', 'Повреждено')}</th>
          </tr>
        </thead>
        <tbody>
          {side.units.map((u) => (
            <tr key={u.unit_id}>
              <td>{u.unit_id}</td>
              <td className="num">{u.quantity_start}</td>
              <td className="num">{u.quantity_end}</td>
              <td className="num">
                {u.damaged_end ? `${u.damaged_end} (${Math.round(u.shell_percent_end ?? 0)}%)` : '—'}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </>
  );
}
