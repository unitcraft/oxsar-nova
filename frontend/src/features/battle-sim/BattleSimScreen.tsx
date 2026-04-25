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
    attack: [entry.attack, 0, 0] as [number, number, number],
    shield: [entry.shield, 0, 0] as [number, number, number],
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
  const { t, tf } = useTranslation();
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
      <h2>{t('global', 'MENU_SIMULATOR')}</h2>

      <div style={{ display: 'flex', gap: 24, marginBottom: 16, flexWrap: 'wrap' }}>
        <UnitPicker
          title={tf('Main', 'BATTLE_SIM_ATTACKERS', 'Атакующий флот')}
          units={COMBAT_SHIPS}
          value={attackers}
          onChange={setAttackers}
        />
        <UnitPicker
          title={tf('Main', 'BATTLE_SIM_DEFENDERS', 'Защита + флот')}
          units={[...COMBAT_SHIPS, ...COMBAT_DEFENSE]}
          value={defenders}
          onChange={setDefenders}
        />
      </div>

      <div style={{ display: 'flex', gap: 24, marginBottom: 12 }}>
        <label>
          {tf('Main', 'BATTLE_SIM_SEED', 'Seed:')}&nbsp;
          <input
            type="number"
            value={seed}
            onChange={(e) => setSeed(Number(e.target.value))}
            style={{ width: 120 }}
          />
        </label>
        <label>
          {tf('Main', 'BATTLE_SIM_NUMSIM', 'Прогонов (1–20):')}&nbsp;
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
          ? tf('Main', 'BATTLE_SIM_CALCULATING', 'Считаем…')
          : tf('Main', 'BATTLE_SIM_RUN', 'Рассчитать')}
      </button>

      {sim.isError && (
        <div className="ox-error" style={{ marginTop: 8 }}>
          {sim.error instanceof Error ? sim.error.message : t('global', 'ERROR')}
        </div>
      )}

      {isStats && (
        <div style={{ marginTop: 16 }}>
          <h3>{tf('Main', 'BATTLE_SIM_RESULT', 'Результат')}</h3>
          <p>
            <b>{tf('Main', 'BATTLE_SIM_RUNS', 'Прогонов')}:</b> {(r as SimStats).num_sim}
            {' · '}
            <b>{tf('Main', 'BATTLE_WIN_RATE', 'Победы атак.')}:</b>{' '}
            {((r as SimStats).win_rate * 100).toFixed(1)}%
            {' · '}
            <b>{tf('Main', 'BATTLE_DRAW_RATE', 'Ничьи')}:</b>{' '}
            {((r as SimStats).draw_rate * 100).toFixed(1)}%
            {' · '}
            <b>{tf('Main', 'BATTLE_AVG_ROUNDS', 'Ср. раундов')}:</b>{' '}
            {(r as SimStats).avg_rounds.toFixed(1)}
          </p>
        </div>
      )}

      {!isStats && r && (
        <div style={{ marginTop: 16 }}>
          <h3>{tf('Main', 'BATTLE_SIM_RESULT', 'Результат')}</h3>
          <p>
            <b>{tf('Main', 'BATTLE_WINNER', 'Победитель')}:</b>{' '}
            {(r as SimReport).winner === 'attackers'
              ? tf('Main', 'BATTLE_WIN_ATT', 'Атакующие')
              : (r as SimReport).winner === 'defenders'
                ? tf('Main', 'BATTLE_WIN_DEF', 'Защитники')
                : tf('Main', 'BATTLE_DRAW', 'Ничья')}
            {' · '}
            <b>{tf('Main', 'BATTLE_ROUNDS', 'Раундов')}:</b> {(r as SimReport).rounds}
            {' · '}
            <b>{tf('Main', 'SEED', 'Сид')}:</b> {(r as SimReport).seed}
          </p>
          {((r as SimReport).debris_metal > 0 || (r as SimReport).debris_silicon > 0) && (
            <p>
              <b>{tf('Main', 'DEBRIS', 'Обломки')}:</b>{' '}
              {(r as SimReport).debris_metal} M / {(r as SimReport).debris_silicon} Si
            </p>
          )}

          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16, marginTop: 8 }}>
            {(r as SimReport).attackers[0] && (
              <LossesTable
                title={tf('Main', 'ATTACKERS', 'Атакующие')}
                side={(r as SimReport).attackers[0]!}
                original={attackers}
              />
            )}
            {(r as SimReport).defenders[0] && (
              <LossesTable
                title={tf('Main', 'DEFENDERS', 'Защитники')}
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
  const { tf } = useTranslation();
  const changed = side.units.filter((u) => u.quantity_end !== u.quantity_start || u.damaged_end);
  return (
    <div>
      <h4>{title}</h4>
      <p>
        {tf('Main', 'BATTLE_LOSSES', 'Потери')}:{' '}
        <b>{side.lost_metal}</b> M / <b>{side.lost_silicon}</b> Si
      </p>
      {changed.length > 0 ? (
        <table className="ox-table">
          <thead>
            <tr>
              <th>{tf('Main', 'UNIT_ID', 'Юнит')}</th>
              <th>{tf('Main', 'BEFORE', 'Было')}</th>
              <th>{tf('Main', 'AFTER', 'Стало')}</th>
              <th>{tf('Main', 'DAMAGED', 'Повреждено')}</th>
            </tr>
          </thead>
          <tbody>
            {changed.map((u) => {
              const unitKey = [...SHIPS, ...DEFENSE].find((s) => s.id === u.unit_id)?.key ?? '';
              return (
                <tr key={u.unit_id}>
                  <td style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                    {unitKey && <img src={imageOf(unitKey)} alt="" width={32} height={32} style={{ imageRendering: 'pixelated' }} />}
                    {nameOf(u.unit_id)}
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
        <p>{tf('Main', 'NO_LOSSES', 'Потерь нет.')}</p>
      )}
    </div>
  );
}

function UnitPicker({
  title, units, value, onChange,
}: {
  title: string;
  units: CombatEntry[];
  value: UnitMap;
  onChange: (v: UnitMap) => void;
}) {
  return (
    <div style={{ minWidth: 260 }}>
      <h3 style={{ marginTop: 0 }}>{title}</h3>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
        {units.map((u) => (
          <label key={u.id} style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <img src={imageOf(u.key)} alt="" width={32} height={32} style={{ imageRendering: 'pixelated', flexShrink: 0 }} />
            <span style={{ flex: 1, fontSize: 15 }}>{nameOf(u.id)}</span>
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
