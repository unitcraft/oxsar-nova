import { useState } from 'react';
import { useMutation } from '@tanstack/react-query';
import { api } from '@/api/client';
import { SHIPS, DEFENSE, nameOf, imageOf, type CombatEntry } from '@/api/catalog';
import { useTranslation } from '@/i18n/i18n';

type UnitMap = Record<number, number>;

interface UnitResult {
  unit_id: number;
  quantity_start: number;
  quantity_end: number;
  damaged_end?: number;
}

interface SideResult {
  user_id: string;
  lost_metal: number;
  lost_silicon: number;
  lost_hydrogen: number;
  units: UnitResult[];
}

interface SimReport {
  seed: number;
  winner: 'attackers' | 'defenders' | 'draw';
  rounds: number;
  debris_metal: number;
  debris_silicon: number;
  attackers: SideResult[];
  defenders: SideResult[];
}

interface SimStats {
  num_sim: number;
  win_rate: number;
  draw_rate: number;
  avg_rounds: number;
}

// COMBAT_UNITS — корабли + оборона с боевыми характеристиками.
const COMBAT_SHIPS = SHIPS.filter((u) => u.attack > 0 || u.shell > 0);
const COMBAT_DEFENSE = DEFENSE;

function makeUnit(entry: CombatEntry, qty: number) {
  return {
    unit_id: entry.id,
    quantity: qty,
    front: entry.front ?? 10,
    ballistics: entry.ballistics ?? 0,
    masking: entry.masking ?? 0,
    attack: entry.attack,
    shield: entry.shield,
    shell: entry.shell,
  };
}

// Строим глобальную таблицу rapidfire из всех юнитов каталога.
function buildRapidfire(entries: CombatEntry[]): Record<number, Record<number, number>> {
  const table: Record<number, Record<number, number>> = {};
  for (const e of entries) {
    if (e.rapidfire && Object.keys(e.rapidfire).length > 0) {
      table[e.id] = e.rapidfire;
    }
  }
  return table;
}
const ALL_COMBAT = [...SHIPS, ...DEFENSE];
const RAPIDFIRE_TABLE = buildRapidfire(ALL_COMBAT);

export function BattleSimScreen() {
  const { t } = useTranslation();
  const { t: ti } = useTranslation('info');
  const [attackers, setAttackers] = useState<UnitMap>({});
  const [defenders, setDefenders] = useState<UnitMap>({});
  const [seed, setSeed] = useState<number>(42);
  const [numSim, setNumSim] = useState<number>(1);

  const sim = useMutation({
    mutationFn: (body: unknown) =>
      api.post<SimStats | SimReport>('/api/battle-sim', body),
  });

  function runSim() {
    const toSide = (map: UnitMap, entries: CombatEntry[]) =>
      [
        {
          user_id: 'sim',
          units: entries
            .filter((e) => (map[e.id] ?? 0) > 0)
            .map((e) => makeUnit(e, map[e.id] ?? 0)),
        },
      ];

    sim.mutate({
      seed,
      num_sim: numSim >= 2 ? numSim : undefined,
      rapidfire: RAPIDFIRE_TABLE,
      attackers: toSide(attackers, COMBAT_SHIPS),
      defenders: toSide(defenders, [...COMBAT_SHIPS, ...COMBAT_DEFENSE]),
    });
  }

  const r = sim.data as SimReport | SimStats | undefined;
  const isStats = r != null && 'num_sim' in r;

  return (
    <section>
      <h2>{t('global', 'menuSimulator')}</h2>

      <div style={{ display: 'flex', gap: 24, marginBottom: 16, flexWrap: 'wrap' }}>
        <UnitPicker
          title={t('main', 'battleSimAttackers')}
          units={COMBAT_SHIPS}
          value={attackers}
          onChange={setAttackers}
          tInfo={ti}
        />
        <UnitPicker
          title={t('main', 'battleSimDefenders')}
          units={[...COMBAT_SHIPS, ...COMBAT_DEFENSE]}
          value={defenders}
          onChange={setDefenders}
          tInfo={ti}
        />
      </div>

      <div style={{ display: 'flex', gap: 24, marginBottom: 12 }}>
        <label>
          {t('main', 'battleSimSeed')}&nbsp;
          <input
            type="number"
            value={seed}
            onChange={(e) => setSeed(Number(e.target.value))}
            style={{ width: 120 }}
          />
        </label>
        <label>
          {t('main', 'battleSimNumsim')}&nbsp;
          <input
            type="number"
            min={1}
            max={20}
            value={numSim}
            onChange={(e) => setNumSim(Math.min(20, Math.max(1, Number(e.target.value))))}
            style={{ width: 60 }}
          />
        </label>
      </div>

      <button type="button" disabled={sim.isPending} onClick={runSim}>
        {sim.isPending
          ? t('main', 'battleSimCalculating')
          : t('main', 'battleSimRun')}
      </button>

      {sim.isError && (
        <div className="ox-error" style={{ marginTop: 8 }}>
          {sim.error instanceof Error ? sim.error.message : t('global', 'error')}
        </div>
      )}

      {isStats && (
        <div style={{ marginTop: 16 }}>
          <h3>{t('main', 'battleSimResult')}</h3>
          <p>
            <b>{t('main', 'battleSimRuns')}:</b> {(r as SimStats).num_sim}
            {' · '}
            <b>{t('main', 'battleWinRate')}:</b>{' '}
            {((r as SimStats).win_rate * 100).toFixed(1)}%
            {' · '}
            <b>{t('main', 'battleDrawRate')}:</b>{' '}
            {((r as SimStats).draw_rate * 100).toFixed(1)}%
            {' · '}
            <b>{t('main', 'battleAvgRounds')}:</b>{' '}
            {(r as SimStats).avg_rounds.toFixed(1)}
          </p>
        </div>
      )}

      {!isStats && r && (
        <div style={{ marginTop: 16 }}>
          <h3>{t('main', 'battleSimResult')}</h3>
          <p>
            <b>{t('main', 'battleWinner')}:</b>{' '}
            {(r as SimReport).winner === 'attackers'
              ? t('main', 'battleWinAtt')
              : (r as SimReport).winner === 'defenders'
                ? t('main', 'battleWinDef')
                : t('main', 'battleDraw')}
            {' · '}
            <b>{t('main', 'battleRounds')}:</b> {(r as SimReport).rounds}
            {' · '}
            <b>{t('main', 'seed')}:</b> {(r as SimReport).seed}
          </p>
          {((r as SimReport).debris_metal > 0 || (r as SimReport).debris_silicon > 0) && (
            <p>
              <b>{t('main', 'debris')}:</b>{' '}
              {(r as SimReport).debris_metal} M / {(r as SimReport).debris_silicon} Si
            </p>
          )}

          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16, marginTop: 8 }}>
            {(r as SimReport).attackers[0] && (
              <LossesTable
                title={t('main', 'attackers')}
                side={(r as SimReport).attackers[0]!}
                original={attackers}
              />
            )}
            {(r as SimReport).defenders[0] && (
              <LossesTable
                title={t('main', 'defenders')}
                side={(r as SimReport).defenders[0]!}
                original={defenders}
              />
            )}
          </div>
        </div>
      )}
    </section>
  );
}

function LossesTable({
  title, side, original,
}: {
  title: string;
  side: SideResult;
  original: UnitMap;
}) {
  const { t } = useTranslation();
  const { t: ti } = useTranslation('info');
  const changed = side.units.filter((u) => u.quantity_end !== u.quantity_start || u.damaged_end);
  return (
    <div>
      <h4>{title}</h4>
      <p>
        {t('main', 'battleLosses')}:{' '}
        <b>{side.lost_metal}</b> M / <b>{side.lost_silicon}</b> Si
      </p>
      {changed.length > 0 ? (
        <table className="ox-table">
          <thead>
            <tr>
              <th>{t('main', 'unitId')}</th>
              <th>{t('main', 'before')}</th>
              <th>{t('main', 'after')}</th>
              <th>{t('main', 'damaged')}</th>
            </tr>
          </thead>
          <tbody>
            {changed.map((u) => {
              const unitKey = [...SHIPS, ...DEFENSE].find((s) => s.id === u.unit_id)?.key ?? '';
              return (
                <tr key={u.unit_id}>
                  <td style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                    {unitKey && <img src={imageOf(unitKey)} alt="" width={32} height={32} style={{ imageRendering: 'pixelated' }} />}
                    {nameOf(u.unit_id, ti)}
                  </td>
                  <td className="num">{original[u.unit_id] ?? u.quantity_start}</td>
                  <td className="num">{u.quantity_end}</td>
                  <td className="num">{u.damaged_end ?? 0}</td>
                </tr>
              );
            })}
          </tbody>
        </table>
      ) : (
        <p>{t('main', 'noLosses')}</p>
      )}
    </div>
  );
}

function UnitPicker({
  title, units, value, onChange, tInfo,
}: {
  title: string;
  units: CombatEntry[];
  value: UnitMap;
  onChange: (v: UnitMap) => void;
  tInfo: (key: string) => string;
}) {
  return (
    <div style={{ minWidth: 260 }}>
      <h3 style={{ marginTop: 0 }}>{title}</h3>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
        {units.map((u) => (
          <label key={u.id} style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <img src={imageOf(u.key)} alt="" width={32} height={32} style={{ imageRendering: 'pixelated', flexShrink: 0 }} />
            <span style={{ flex: 1, fontSize: 15 }}>{nameOf(u.id, tInfo)}</span>
            <input
              type="number"
              min={0}
              value={value[u.id] ?? 0}
              onChange={(e) => onChange({ ...value, [u.id]: Math.max(0, Number(e.target.value)) })}
              style={{ width: 70, flexShrink: 0 }}
            />
          </label>
        ))}
      </div>
    </div>
  );
}
