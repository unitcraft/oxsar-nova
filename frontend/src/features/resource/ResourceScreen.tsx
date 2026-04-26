import { useEffect, useRef, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { catalog, BUILDINGS } from '@/api/catalog';
import { useTranslation } from '@/i18n/i18n';

const BUILDING_NAMES: Record<number, string> = Object.fromEntries(
  BUILDINGS.map((b) => [b.id, b.name]),
);
import type { ResourceBuilding } from '@/api/types';
import { useToast } from '@/ui/Toast';
import { ScreenSkeleton } from '@/ui/Skeleton';

function fmt(v: number): string {
  const n = Math.round(v);
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`;
  if (n >= 1_000) return `${Math.round(n / 1_000)}k`;
  return n.toLocaleString('ru-RU');
}

function numColor(v: number): string {
  return v > 0 ? 'var(--ox-success)' : v < 0 ? 'var(--ox-danger)' : 'var(--ox-fg-dim)';
}

const TD_NUM: React.CSSProperties = {
  width: '12%',
  textAlign: 'right',
  fontFamily: 'var(--ox-mono)',
  fontSize: 15,
  paddingRight: 10,
  whiteSpace: 'nowrap',
};

const TR_BASE: React.CSSProperties = {
  borderBottom: '1px solid var(--ox-border)',
};

export function ResourceScreen({ planetId }: { planetId: string }) {
  const { t } = useTranslation('resourceUi');
  const { t: tg } = useTranslation('global');
  const qc = useQueryClient();
  const toast = useToast();
  const [factors, setFactors] = useState<Record<string, number>>({});
  const factorsRef = useRef(factors);
  factorsRef.current = factors;
  const [modalBuilding, setModalBuilding] = useState<ResourceBuilding | null>(null);

  const { data: report, isLoading } = useQuery({
    queryKey: ['resource-report', planetId],
    queryFn: () => catalog.getResourceReport(planetId),
  });

  useEffect(() => {
    if (!report) return;
    const init: Record<string, number> = {};
    report.buildings.forEach((b) => { init[b.unit_id] = b.factor; });
    setFactors(init);
  }, [report]);

  const save = useMutation({
    mutationFn: (fs: Record<string, number>) =>
      catalog.updateResourceFactors(planetId, {
        factors: Object.fromEntries(
          Object.entries(fs).map(([k, v]) => [k, Math.max(0, Math.min(100, v))]),
        ),
      }),
    onSuccess: () => { void qc.invalidateQueries({ queryKey: ['resource-report', planetId] }); },
    onError: (err) => {
      toast.show('danger', tg('error'), err instanceof Error ? err.message : t('loadError'));
    },
  });

  const commitFactor = (unitId: string, value: number) => {
    const next = { ...factorsRef.current, [unitId]: value };
    setFactors(next);
    save.mutate(next);
  };

  const setAll = (value: number) => {
    if (!report) return;
    const next: Record<string, number> = {};
    report.buildings.forEach((b) => { next[b.unit_id] = value; });
    setFactors(next);
    save.mutate(next);
  };

  if (isLoading) return <ScreenSkeleton />;
  if (!report) return <div style={{ color: 'var(--ox-danger)', padding: 24 }}>{t('loadError')}</div>;

  const buildings = report.buildings.filter((b) => b.level > 0);

  const ph = report.basic_metal   + buildings.reduce((s, b) => s + b.prod_metal    * (factors[b.unit_id] ?? b.factor) / 100, 0);
  const sh = report.basic_silicon + buildings.reduce((s, b) => s + b.prod_silicon  * (factors[b.unit_id] ?? b.factor) / 100, 0);
  const hh =                        buildings.reduce((s, b) => s + b.prod_hydrogen * (factors[b.unit_id] ?? b.factor) / 100, 0);
  const te =                        buildings.reduce((s, b) => s + b.cons_energy   * (factors[b.unit_id] ?? b.factor) / 100, 0);

  const modalFactor = modalBuilding ? (factors[modalBuilding.unit_id] ?? modalBuilding.factor) : 0;

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20, paddingBottom: modalBuilding ? 140 : 0 }}>

      {/* Header */}
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: 8 }}>
        <h2 style={{ margin: 0, fontSize: 18, fontFamily: 'var(--ox-font)', fontWeight: 700 }}>
          {t('title', { planetName: report.planet_name })}
        </h2>
        <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
          {save.isPending && <span style={{ fontSize: 14, color: 'var(--ox-fg-muted)', fontFamily: 'var(--ox-mono)' }}>{t('saving')}</span>}
          <button type="button" className="btn btn-sm btn-ghost" disabled={save.isPending} onClick={() => setAll(0)}>{t('turnOffAll')}</button>
          <button type="button" className="btn btn-sm btn-ghost" disabled={save.isPending} onClick={() => setAll(100)}>{t('turnOnAll')}</button>
        </div>
      </div>

      {/* Table */}
      <div className="ox-panel" style={{ padding: 0 }}>
        <table style={{ width: '100%', borderCollapse: 'collapse', tableLayout: 'fixed' }}>
          <colgroup>
            <col style={{ width: 'auto' }} />
            <col style={{ width: '12%' }} />
            <col style={{ width: '12%' }} />
            <col style={{ width: '12%' }} />
            <col style={{ width: '12%' }} />
          </colgroup>
          <thead>
            <tr style={{ borderBottom: '2px solid var(--ox-border)' }}>
              <th style={{ textAlign: 'left', fontSize: 13, fontWeight: 700, letterSpacing: '0.08em', textTransform: 'uppercase', color: 'var(--ox-fg-muted)', paddingLeft: 12, paddingTop: 8, paddingBottom: 8 }}>{t('colBuilding')}</th>
              <th style={{ ...TD_NUM, fontSize: 13, fontWeight: 700, color: 'var(--ox-fg-muted)', paddingTop: 8, paddingBottom: 8 }}>🟠</th>
              <th style={{ ...TD_NUM, fontSize: 13, fontWeight: 700, color: 'var(--ox-fg-muted)', paddingTop: 8, paddingBottom: 8 }}>💎</th>
              <th style={{ ...TD_NUM, fontSize: 13, fontWeight: 700, color: 'var(--ox-fg-muted)', paddingTop: 8, paddingBottom: 8 }}>💧</th>
              <th style={{ ...TD_NUM, fontSize: 13, fontWeight: 700, color: 'var(--ox-fg-muted)', paddingTop: 8, paddingBottom: 8 }}>⚡</th>
            </tr>
          </thead>
          <tbody>
            {/* Natural */}
            <tr style={{ ...TR_BASE, background: 'var(--ox-bg-2)' }}>
              <td style={{ paddingLeft: 12, fontSize: 15, paddingTop: 8, paddingBottom: 8, color: 'var(--ox-fg-muted)', fontStyle: 'italic' }}>{t('natural')}</td>
              <td style={{ ...TD_NUM, color: numColor(report.basic_metal) }}>{fmt(report.basic_metal)}</td>
              <td style={{ ...TD_NUM, color: numColor(report.basic_silicon) }}>{fmt(report.basic_silicon)}</td>
              <td style={{ ...TD_NUM, color: 'var(--ox-fg-dim)' }}>—</td>
              <td style={{ ...TD_NUM, color: 'var(--ox-fg-dim)' }}>—</td>
            </tr>

            {/* Buildings */}
            {buildings.map((b) => {
              const factor = factors[b.unit_id] ?? b.factor;
              const metal    = b.prod_metal    * factor / 100;
              const silicon  = b.prod_silicon  * factor / 100;
              const hydrogen = b.prod_hydrogen * factor / 100;
              const energy   = b.cons_energy   * factor / 100;
              return (
                <tr
                  key={b.unit_id}
                  style={{ ...TR_BASE, cursor: b.allow_factor ? 'pointer' : 'default' }}
                  onClick={() => b.allow_factor && setModalBuilding(b)}
                >
                  <td style={{ paddingLeft: 12, paddingTop: 8, paddingBottom: 8, overflow: 'hidden' }}>
                    <span style={{ fontWeight: 500, fontSize: 15, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis', display: 'block' }}>
                      {BUILDING_NAMES[b.unit_id] ?? b.name}
                      <span style={{ fontSize: 13, color: 'var(--ox-fg-muted)', marginLeft: 4 }}>{t('levelAbbr')} {b.level}</span>
                      {b.allow_factor && (
                        <span style={{ fontSize: 13, fontFamily: 'var(--ox-mono)', marginLeft: 4, color: factor < 100 ? 'var(--ox-warn, #f59e0b)' : 'var(--ox-fg-dim)' }}>
                          {factor}%
                        </span>
                      )}
                    </span>
                  </td>
                  <td style={{ ...TD_NUM, color: numColor(metal) }}>{metal !== 0 ? fmt(metal) : <span style={{ color: 'var(--ox-fg-dim)' }}>—</span>}</td>
                  <td style={{ ...TD_NUM, color: numColor(silicon) }}>{silicon !== 0 ? fmt(silicon) : <span style={{ color: 'var(--ox-fg-dim)' }}>—</span>}</td>
                  <td style={{ ...TD_NUM, color: numColor(hydrogen) }}>{hydrogen !== 0 ? fmt(hydrogen) : <span style={{ color: 'var(--ox-fg-dim)' }}>—</span>}</td>
                  <td style={{ ...TD_NUM, color: numColor(energy) }}>{energy !== 0 ? fmt(energy) : <span style={{ color: 'var(--ox-fg-dim)' }}>—</span>}</td>
                </tr>
              );
            })}

            {/* Storage */}
            <SummaryRow label={t('storage')} metal={report.storage_metal} silicon={report.storage_silicon} hydrogen={report.storage_hydrogen} energy={null} topBorder dim />
            <SummaryRow label={t('perHour')}   metal={ph}          silicon={sh}          hydrogen={hh}          energy={te}   topBorder />
            <SummaryRow label={t('perDay')}    metal={ph * 24}     silicon={sh * 24}     hydrogen={hh * 24}     energy={null} />
            <SummaryRow label={t('perWeek')}   metal={ph * 24 * 7} silicon={sh * 24 * 7} hydrogen={hh * 24 * 7} energy={null} />
          </tbody>
        </table>
      </div>

      {/* Bottom sheet */}
      {modalBuilding && (
        <div style={{
          position: 'fixed', bottom: 0, left: 0, right: 0, zIndex: 200,
          background: 'var(--ox-bg-panel-2)',
          borderTop: '1px solid var(--ox-border)',
          backdropFilter: 'blur(12px)',
          padding: '12px 20px 20px',
          boxShadow: '0 -4px 24px rgba(0,0,0,0.5)',
        }}>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 12 }}>
            <div>
              <span style={{ fontWeight: 700, fontSize: 16 }}>{BUILDING_NAMES[modalBuilding.unit_id] ?? modalBuilding.name}</span>
              <span style={{ fontSize: 14, color: 'var(--ox-fg-muted)', marginLeft: 8 }}>{t('levelAbbr')} {modalBuilding.level}</span>
            </div>
            <button type="button" className="btn-ghost btn-sm" onClick={() => setModalBuilding(null)}>✕</button>
          </div>
          <FactorInput
            value={modalFactor}
            onChange={(v) => setFactors((prev) => ({ ...prev, [modalBuilding.unit_id]: v }))}
            onCommit={(v) => commitFactor(String(modalBuilding.unit_id), v)}
            disabled={save.isPending}
          />
        </div>
      )}
    </div>
  );
}

function SummaryRow({ label, metal, silicon, hydrogen, energy, topBorder, dim }: {
  label: string; metal: number; silicon: number; hydrogen: number;
  energy: number | null; topBorder?: boolean; dim?: boolean;
}) {
  return (
    <tr style={{
      borderTop: topBorder ? '2px solid var(--ox-border)' : undefined,
      borderBottom: '1px solid var(--ox-border)',
      background: dim ? 'var(--ox-bg-2)' : undefined,
    }}>
      <td style={{ paddingLeft: 12, paddingTop: 8, paddingBottom: 8, fontSize: 15, fontWeight: 700, color: dim ? 'var(--ox-fg-muted)' : undefined }}>{label}</td>
      <td style={{ ...TD_NUM, color: numColor(metal) }}>{fmt(metal)}</td>
      <td style={{ ...TD_NUM, color: numColor(silicon) }}>{fmt(silicon)}</td>
      <td style={{ ...TD_NUM, color: numColor(hydrogen) }}>{fmt(hydrogen)}</td>
      <td style={{ ...TD_NUM, color: 'var(--ox-fg-dim)' }}>
        {energy !== null ? <span style={{ color: numColor(energy) }}>{fmt(energy)}</span> : '—'}
      </td>
    </tr>
  );
}

const PRESETS = [0, 25, 50, 75, 100];

function FactorInput({ value, onChange, onCommit, disabled }: {
  value: number; onChange: (v: number) => void;
  onCommit: (v: number) => void; disabled: boolean;
}) {
  const clamp = (n: number) => Math.max(0, Math.min(100, Math.round(n)));
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
      <div style={{ display: 'flex', gap: 6, justifyContent: 'center' }}>
        {PRESETS.map((p) => (
          <button
            key={p}
            type="button"
            disabled={disabled}
            onClick={() => onCommit(p)}
            style={{
              padding: '5px 10px', fontSize: 15, fontFamily: 'var(--ox-mono)',
              background: value === p ? 'var(--ox-accent)' : 'var(--ox-bg-3)',
              color: value === p ? '#000' : 'var(--ox-fg-dim)',
              border: '1px solid var(--ox-border)', borderRadius: 4,
              cursor: disabled ? 'default' : 'pointer',
              fontWeight: value === p ? 700 : 400,
            }}
          >
            {p}%
          </button>
        ))}
      </div>
      <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
        <input
          type="range" min={0} max={100} step={1} value={value} disabled={disabled}
          onChange={(e) => onChange(clamp(Number(e.target.value)))}
          onMouseUp={(e) => onCommit(clamp(Number((e.target as HTMLInputElement).value)))}
          onTouchEnd={(e) => onCommit(clamp(Number((e.target as HTMLInputElement).value)))}
          style={{ flex: 1, accentColor: 'var(--ox-accent)', cursor: disabled ? 'default' : 'pointer' }}
        />
        <span style={{ fontSize: 15, fontFamily: 'var(--ox-mono)', fontWeight: 700, color: 'var(--ox-fg)', minWidth: 40, textAlign: 'right' }}>
          {value}%
        </span>
      </div>
    </div>
  );
}
