// Админ-drawer с полной карточкой игрока (план 14 Ф.2).
//
// Открывается кликом по строке в таблице игроков (AdminScreen).
// Backend: GET /api/admin/users/{id} — один агрегированный запрос.

import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';
import { useTranslation } from '@/i18n/i18n';

interface Planet {
  id: string; name: string; galaxy: number; system: number; position: number;
  is_moon: boolean; metal: number; silicon: number; hydrogen: number;
}
interface Fleet {
  id: string; mission: number; state: string;
  dst_galaxy: number; dst_system: number; dst_position: number; dst_is_moon: boolean;
  arrive_at: string;
}
interface MarketLot {
  id: string; sell_resource: string; sell_amount: number;
  buy_resource: string; buy_amount: number; state: string;
}
interface ArtefactLot {
  id: string; artefact_id: string; unit_id: number; price_credit: number;
}
interface OfficerActive { key: string; expires_at: string }
interface ArtefactItem { id: string; unit_id: number; state: string }
interface ResLogEntry {
  reason: string; planet_id?: string | null;
  d_metal: number; d_silicon: number; d_hydrogen: number; created_at: string;
}
interface Purchase {
  id: string; package_key: string; credits: number; price_rub: number;
  status: string; created_at: string; paid_at?: string | null;
}
interface MessageShort { id: string; folder: number; subject: string; created_at: string; read: boolean }
interface ReportShort { id: string; kind: 'battle' | 'espionage' | 'expedition'; created_at: string }

interface UserProfile {
  id: string; username: string; email: string; role: string;
  credit: number; score: number;
  banned_at?: string | null; created_at: string; last_seen_at: string;
  planets: Planet[];
  fleets: Fleet[];
  market_lots: MarketLot[];
  artefact_lots: ArtefactLot[];
  officers: OfficerActive[];
  artefacts: ArtefactItem[];
  res_log: ResLogEntry[];
  purchases: Purchase[];
  messages_recent: MessageShort[];
  reports_recent: ReportShort[];
}

export function AdminUserProfilePanel({ userID, onClose }: { userID: string; onClose: () => void }) {
  const { t } = useTranslation('adminUi');
  const qc = useQueryClient();
  const [granularity, setGranularity] = useState<'summary' | 'economy' | 'combat' | 'audit'>('summary');

  const profile = useQuery({
    queryKey: ['admin-user-profile', userID],
    queryFn: () => api.get<UserProfile>(`/api/admin/users/${userID}`),
  });

  const p = profile.data;

  return (
    <div
      role="dialog"
      aria-modal="true"
      style={{
        position: 'fixed', top: 0, right: 0, bottom: 0, width: 'min(780px, 100vw)',
        background: 'var(--ox-bg-panel-2, #0b132b)', borderLeft: '1px solid #444',
        boxShadow: '-8px 0 24px rgba(0,0,0,0.5)',
        zIndex: 100, overflow: 'auto', padding: 20,
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', marginBottom: 16, gap: 12 }}>
        <h3 style={{ margin: 0, flex: 1 }}>
          {p ? (<>👤 {p.username} <span style={{ color: 'var(--ox-fg-muted)', fontWeight: 400, fontSize: 15 }}>{p.email}</span></>) : t('profileLoading')}
        </h3>
        <button type="button" className="btn-ghost btn-sm" onClick={onClose}>{t('profileCloseBtn')}</button>
      </div>

      {profile.isLoading && <p>{t('profileLoading')}</p>}
      {profile.isError && <p style={{ color: 'var(--ox-danger)' }}>{t('profileLoadErr')}</p>}

      {p && (
        <>
          <div style={{ display: 'flex', gap: 16, marginBottom: 12, flexWrap: 'wrap', fontSize: 14 }}>
            <span><b>{t('profileLabelRole')}</b> {p.role || 'player'}</span>
            <span><b>{t('profileLabelCredits')}</b> {p.credit.toLocaleString('ru-RU')}</span>
            <span><b>{t('profileLabelScore')}</b> {p.score.toLocaleString('ru-RU')}</span>
            <span><b>{t('profileLabelReg')}</b> {fmtDate(p.created_at)}</span>
            <span><b>{t('profileLabelOnline')}</b> {fmtDate(p.last_seen_at)}</span>
            {p.banned_at && <span style={{ color: 'var(--ox-danger)' }}>{t('profileBanned')} {fmtDate(p.banned_at)}</span>}
          </div>

          <div className="ox-tabs" style={{ marginBottom: 16 }}>
            <button type="button" aria-pressed={granularity === 'summary'} onClick={() => setGranularity('summary')}>{t('profileTabSummary')}</button>
            <button type="button" aria-pressed={granularity === 'economy'} onClick={() => setGranularity('economy')}>{t('profileTabEconomy')}</button>
            <button type="button" aria-pressed={granularity === 'combat'} onClick={() => setGranularity('combat')}>{t('profileTabCombat')}</button>
            <button type="button" aria-pressed={granularity === 'audit'} onClick={() => setGranularity('audit')}>{t('profileTabAudit')}</button>
          </div>

          {granularity === 'summary' && (
            <SummaryTab p={p} userID={userID} onChanged={() => qc.invalidateQueries({ queryKey: ['admin-user-profile', userID] })} />
          )}
          {granularity === 'economy' && <EconomyTab p={p} />}
          {granularity === 'combat' && <CombatTab p={p} />}
          {granularity === 'audit' && <AuditTab userID={userID} />}
        </>
      )}
    </div>
  );
}

function SummaryTab({ p, userID, onChanged }: { p: UserProfile; userID: string; onChanged: () => void }) {
  const { t } = useTranslation('adminUi');
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      <Section title={`${t('sectionPlanets')} (${p.planets.length})`}>
        {p.planets.length === 0 ? <Empty /> : (
          <table style={tableStyle}>
            <thead>
              <tr><Th>{t('colName')}</Th><Th>{t('colCoords')}</Th><Th>🟠</Th><Th>💎</Th><Th>💧</Th></tr>
            </thead>
            <tbody>
              {p.planets.map((pl) => (
                <tr key={pl.id}>
                  <Td>{pl.is_moon ? '🌑 ' : ''}{pl.name}</Td>
                  <Td><code>[{pl.galaxy}:{pl.system}:{pl.position}]</code></Td>
                  <Td>{pl.metal.toLocaleString('ru-RU')}</Td>
                  <Td>{pl.silicon.toLocaleString('ru-RU')}</Td>
                  <Td>{pl.hydrogen.toLocaleString('ru-RU')}</Td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </Section>

      <Section title={t('sectionResources')}>
        <ResourceGranter userID={userID} planets={p.planets} onDone={onChanged} />
      </Section>

      <Section title={`${t('sectionFleets')} (${p.fleets.length})`}>
        {p.fleets.length === 0 ? <Empty /> : (
          <table style={tableStyle}>
            <thead><tr><Th>{t('colMission')}</Th><Th>State</Th><Th>{t('colDest')}</Th><Th>{t('colArrival')}</Th></tr></thead>
            <tbody>
              {p.fleets.map((f) => (
                <tr key={f.id}>
                  <Td>{missionLabel(f.mission, t)}</Td>
                  <Td>{f.state}</Td>
                  <Td><code>[{f.dst_galaxy}:{f.dst_system}:{f.dst_position}]{f.dst_is_moon ? ' 🌑' : ''}</code></Td>
                  <Td>{fmtDate(f.arrive_at)}</Td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </Section>

      <Section title={`${t('sectionOfficers')} (${p.officers.length})`}>
        {p.officers.length === 0 ? <Empty /> : (
          <ul>{p.officers.map((o) => <li key={o.key}><b>{o.key}</b> {t('officerUntil')} {fmtDate(o.expires_at)}</li>)}</ul>
        )}
      </Section>

      <Section title={`${t('sectionArtefacts')} (${p.artefacts.length})`}>
        <ArtefactsBlock userID={userID} items={p.artefacts} onChanged={onChanged} />
      </Section>
    </div>
  );
}

function EconomyTab({ p }: { p: UserProfile }) {
  const { t } = useTranslation('adminUi');
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      <Section title={`${t('sectionMarketLots')} (${p.market_lots.length})`}>
        {p.market_lots.length === 0 ? <Empty /> : (
          <table style={tableStyle}>
            <thead><tr><Th>{t('colSells')}</Th><Th>{t('colWants')}</Th><Th>State</Th></tr></thead>
            <tbody>
              {p.market_lots.map((l) => (
                <tr key={l.id}>
                  <Td>{l.sell_amount.toLocaleString('ru-RU')} {l.sell_resource}</Td>
                  <Td>{l.buy_amount.toLocaleString('ru-RU')} {l.buy_resource}</Td>
                  <Td>{l.state}</Td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </Section>

      <Section title={`${t('sectionArtefactLots')} (${p.artefact_lots.length})`}>
        {p.artefact_lots.length === 0 ? <Empty /> : (
          <table style={tableStyle}>
            <thead><tr><Th>{t('colArtefact')}</Th><Th>unit_id</Th><Th>{t('colPrice')}</Th></tr></thead>
            <tbody>
              {p.artefact_lots.map((l) => (
                <tr key={l.id}>
                  <Td><code>{l.artefact_id.slice(0, 8)}</code></Td>
                  <Td>{l.unit_id}</Td>
                  <Td>{l.price_credit.toLocaleString('ru-RU')} 💳</Td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </Section>

      <Section title={`${t('sectionResLog')} (${p.res_log.length})`}>
        {p.res_log.length === 0 ? <Empty /> : (
          <table style={tableStyle}>
            <thead><tr><Th>{t('colDate')}</Th><Th>Reason</Th><Th>🟠</Th><Th>💎</Th><Th>💧</Th></tr></thead>
            <tbody>
              {p.res_log.map((l, i) => (
                <tr key={i}>
                  <Td>{fmtDate(l.created_at)}</Td>
                  <Td><code>{l.reason}</code></Td>
                  <Td>{sign(l.d_metal)}</Td>
                  <Td>{sign(l.d_silicon)}</Td>
                  <Td>{sign(l.d_hydrogen)}</Td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </Section>

      <Section title={`${t('sectionPurchases')} (${p.purchases.length})`}>
        {p.purchases.length === 0 ? <Empty /> : (
          <table style={tableStyle}>
            <thead><tr><Th>{t('colDate')}</Th><Th>{t('colPackage')}</Th><Th>{t('colCreditsShort')}</Th><Th>{t('colPrice')}</Th><Th>{t('colStatus')}</Th></tr></thead>
            <tbody>
              {p.purchases.map((pu) => (
                <tr key={pu.id}>
                  <Td>{fmtDate(pu.created_at)}</Td>
                  <Td>{pu.package_key}</Td>
                  <Td>+{pu.credits}</Td>
                  <Td>{pu.price_rub} ₽</Td>
                  <Td>{pu.status}</Td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </Section>
    </div>
  );
}

function CombatTab({ p }: { p: UserProfile }) {
  const { t } = useTranslation('adminUi');
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      <Section title={`${t('sectionReports')} (${p.reports_recent.length})`}>
        {p.reports_recent.length === 0 ? <Empty /> : (
          <ul>
            {p.reports_recent.map((r) => (
              <li key={`${r.kind}-${r.id}`}>
                <b>{reportLabel(r.kind, t)}</b> <code>{r.id.slice(0, 8)}</code> — {fmtDate(r.created_at)}
              </li>
            ))}
          </ul>
        )}
      </Section>

      <Section title={`${t('sectionMessages')} (${p.messages_recent.length})`}>
        {p.messages_recent.length === 0 ? <Empty /> : (
          <table style={tableStyle}>
            <thead><tr><Th>{t('colDate')}</Th><Th>{t('colFolder')}</Th><Th>{t('colSubject')}</Th><Th>{t('colRead')}</Th></tr></thead>
            <tbody>
              {p.messages_recent.map((m) => (
                <tr key={m.id}>
                  <Td>{fmtDate(m.created_at)}</Td>
                  <Td>{m.folder}</Td>
                  <Td>{m.subject}</Td>
                  <Td>{m.read ? '✓' : '—'}</Td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </Section>
    </div>
  );
}

function AuditTab({ userID }: { userID: string }) {
  const { t } = useTranslation('adminUi');
  const q = useQuery({
    queryKey: ['admin-audit-target', userID],
    queryFn: () => api.get<{ entries: Array<{ id: string; action: string; admin_name: string; created_at: string; status: number }> }>(
      `/api/admin/audit?target_id=${encodeURIComponent(userID)}&limit=100`,
    ),
  });
  const entries = q.data?.entries ?? [];
  return (
    <Section title={`${t('sectionAdminAudit')} (${entries.length})`}>
      {entries.length === 0 ? <Empty /> : (
        <table style={tableStyle}>
          <thead><tr><Th>{t('colDate')}</Th><Th>{t('colAdmin')}</Th><Th>{t('colAction')}</Th><Th>{t('colStatus')}</Th></tr></thead>
          <tbody>
            {entries.map((e) => (
              <tr key={e.id}>
                <Td>{fmtDate(e.created_at)}</Td>
                <Td>{e.admin_name}</Td>
                <Td><code>{e.action}</code></Td>
                <Td>{e.status}</Td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </Section>
  );
}

function ResourceGranter({ userID, planets, onDone }: { userID: string; planets: Planet[]; onDone: () => void }) {
  const { t } = useTranslation('adminUi');
  const [planetID, setPlanetID] = useState(planets[0]?.id ?? '');
  const [m, setM] = useState(0);
  const [s, setS] = useState(0);
  const [h, setH] = useState(0);

  const mut = useMutation({
    mutationFn: () => api.post(`/api/admin/users/${userID}/resources`, {
      planet_id: planetID, metal: m, silicon: s, hydrogen: h,
    }),
    onSuccess: () => {
      setM(0); setS(0); setH(0);
      onDone();
    },
  });

  if (planets.length === 0) return <Empty />;

  return (
    <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap', alignItems: 'center' }}>
      <select value={planetID} onChange={(e) => setPlanetID(e.target.value)}>
        {planets.map((pl) => (
          <option key={pl.id} value={pl.id}>{pl.name} [{pl.galaxy}:{pl.system}:{pl.position}]</option>
        ))}
      </select>
      <NumInput label="🟠" value={m} onChange={setM} />
      <NumInput label="💎" value={s} onChange={setS} />
      <NumInput label="💧" value={h} onChange={setH} />
      <button type="button" disabled={mut.isPending || (m === 0 && s === 0 && h === 0) || !planetID}
        onClick={() => mut.mutate()}>
        {t('resourceApplyBtn')}
      </button>
      {mut.isSuccess && <span style={{ color: 'var(--ox-success)' }}>✓</span>}
      {mut.isError && <span style={{ color: 'var(--ox-danger)' }}>✗</span>}
    </div>
  );
}

function ArtefactsBlock({ userID, items, onChanged }: { userID: string; items: ArtefactItem[]; onChanged: () => void }) {
  const { t } = useTranslation('adminUi');
  const [unitID, setUnitID] = useState(0);
  const grant = useMutation({
    mutationFn: () => api.post(`/api/admin/users/${userID}/artefacts/grant`, { unit_id: unitID }),
    onSuccess: () => { setUnitID(0); onChanged(); },
  });
  const del = useMutation({
    mutationFn: (aid: string) => api.delete(`/api/admin/users/${userID}/artefacts/${aid}`),
    onSuccess: onChanged,
  });

  return (
    <div>
      <div style={{ display: 'flex', gap: 8, alignItems: 'center', marginBottom: 8 }}>
        <NumInput label="unit_id" value={unitID} onChange={setUnitID} />
        <button type="button" disabled={grant.isPending || unitID <= 0} onClick={() => grant.mutate()}>{t('artefactGrantBtn')}</button>
      </div>
      {items.length === 0 ? <Empty /> : (
        <table style={tableStyle}>
          <thead><tr><Th>id</Th><Th>unit</Th><Th>state</Th><Th /></tr></thead>
          <tbody>
            {items.map((a) => (
              <tr key={a.id}>
                <Td><code>{a.id.slice(0, 8)}</code></Td>
                <Td>{a.unit_id}</Td>
                <Td>{a.state}</Td>
                <Td>
                  <button type="button" className="btn-ghost btn-sm"
                    disabled={del.isPending} onClick={() => del.mutate(a.id)}
                    style={{ color: 'var(--ox-danger)' }}>
                    ✕
                  </button>
                </Td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}

/* ── utils ── */

function NumInput({ label, value, onChange }: { label: string; value: number; onChange: (v: number) => void }) {
  return (
    <label style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
      <span>{label}</span>
      <input type="number" value={value} onChange={(e) => onChange(Number(e.target.value))}
        style={{ width: 120 }} />
    </label>
  );
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div>
      <h4 style={{ margin: '0 0 8px 0', fontSize: 15, color: 'var(--ox-fg-dim)' }}>{title}</h4>
      {children}
    </div>
  );
}

function Empty() {
  return <span style={{ color: 'var(--ox-fg-muted)' }}>—</span>;
}

function Th({ children }: { children: React.ReactNode }) {
  return <th style={{ textAlign: 'left', padding: '4px 8px', borderBottom: '1px solid #444', fontSize: 13 }}>{children}</th>;
}
function Td({ children }: { children?: React.ReactNode }) {
  return <td style={{ padding: '4px 8px', borderBottom: '1px solid #2a2a2a', fontSize: 14 }}>{children}</td>;
}

const tableStyle: React.CSSProperties = { width: '100%', borderCollapse: 'collapse' };

function fmtDate(iso: string) {
  if (!iso) return '';
  return new Date(iso).toLocaleString('ru-RU', { day: '2-digit', month: '2-digit', year: '2-digit', hour: '2-digit', minute: '2-digit' });
}

function sign(n: number): string {
  if (n === 0) return '—';
  const s = n.toLocaleString('ru-RU');
  return n > 0 ? `+${s}` : s;
}

const MISSION_KEYS: Record<number, string> = {
  1: 'missionAttack', 2: 'missionEspionage', 7: 'missionTransport', 8: 'missionColonize',
  9: 'missionHarvest', 10: 'missionExpedition', 11: 'missionInvasion',
};

function missionLabel(mission: number, t: (key: string) => string): string {
  const key = MISSION_KEYS[mission];
  return key ? t(key) : String(mission);
}

const REPORT_KEYS: Record<ReportShort['kind'], string> = {
  battle: 'reportBattle', espionage: 'reportEspionage', expedition: 'reportExpedition',
};

function reportLabel(kind: ReportShort['kind'], t: (key: string) => string): string {
  return t(REPORT_KEYS[kind]);
}
