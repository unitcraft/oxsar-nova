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

export function MessagesScreen() {
  const { t, tf } = useTranslation();
  const qc = useQueryClient();
  const [selectedID, setSelectedID] = useState<string | null>(null);

  const list = useQuery({
    queryKey: ['messages'],
    queryFn: () => api.get<{ messages: Message[] | null }>('/api/messages'),
    refetchInterval: 10000,
  });

  const markRead = useMutation({
    mutationFn: (id: string) => api.post<void>(`/api/messages/${id}/read`),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['messages'] });
    },
  });

  const msgs = list.data?.messages ?? [];
  const selected = msgs.find((m) => m.id === selectedID) ?? null;

  function onSelect(m: Message) {
    setSelectedID(m.id);
    if (!m.read_at) {
      markRead.mutate(m.id);
    }
  }

  return (
    <section>
      <h2>{tf('global', 'MENU_MESSAGES', 'Сообщения')}</h2>

      {list.isLoading && <p>…</p>}
      {list.error && (
        <p className="ox-error">
          {t('global', 'ERROR')}:{' '}
          {list.error instanceof Error ? list.error.message : ''}
        </p>
      )}

      {msgs.length === 0 ? (
        <p>{tf('Main', 'INBOX_EMPTY', 'Нет сообщений.')}</p>
      ) : (
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 2fr', gap: 16 }}>
          <table className="ox-table">
            <thead>
              <tr>
                <th>{tf('Main', 'MSG_FROM', 'От')}</th>
                <th>{tf('Main', 'MSG_SUBJECT', 'Тема')}</th>
                <th>{tf('Main', 'MSG_WHEN', 'Когда')}</th>
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
                  </tr>
                );
              })}
            </tbody>
          </table>
          <div>{selected ? <MessageDetail message={selected} /> : <p>—</p>}</div>
        </div>
      )}
    </section>
  );
}

function MessageDetail({ message }: { message: Message }) {
  const { tf } = useTranslation();
  const report = useQuery({
    queryKey: ['battle-report', message.battle_report_id],
    queryFn: () =>
      api.get<BattleReportFull>(`/api/battle-reports/${message.battle_report_id}`),
    enabled: !!message.battle_report_id,
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
    </div>
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
