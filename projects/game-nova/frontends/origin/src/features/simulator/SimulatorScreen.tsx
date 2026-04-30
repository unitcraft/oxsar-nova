// S-046 Simulator — боевой симулятор (план 72.1 ч.20.7).
// Pixel-perfect клон legacy simulator.tpl + ADR-0002 порт
// oxsar2-java/Assault.java (rendering на frontend).
//
// Структура (legacy 1:1):
// 1. «Установки» — tech-уровни (gun/shield/shell/ballistics/masking/
//    shipyard для атакующего, defense_factory для защитника) + поле
//    «Количество симуляций».
// 2. Таблица юнитов: колонки Тип корабля / Атакующий / Защитник.
//    Каждая ячейка: <+> / <quantity> [<damaged> - <shell_%>%] / <->
//    Bulk-кнопки в шапке: «Установить флот», «100%», «90%», «80%», «70%», «60%».
// 3. Селект «Уничтожить» здание + поле «Уровень» (для гравитона).
// 4. Submit «Симулировать».
// 5. Report — раунды, потери, обломки.

import { useState } from 'react';
import { useMutation, useQuery } from '@tanstack/react-query';
import {
  runSimulation,
  type SimInput,
  type SimRunResponse,
  type SimSide,
  type SimStats,
  type SimUnit,
} from '@/api/simulator';
import { fetchShipyardInventory } from '@/api/shipyard';
import { QK } from '@/api/query-keys';
import { catalogByGroup, findCatalog } from '@/features/common/catalog';
import { useResolvedPlanet } from '@/features/common/useResolvedPlanet';
import { Link } from 'react-router-dom';
import { useTranslation } from '@/i18n/i18n';

// Базовые статы юнитов из configs/{ships,defense}.yml.
// Backend Calculate сам применит tech-модификаторы.
interface UnitStats {
  attack: number;
  shield: number;
  shell: number;
}

const UNIT_STATS: Record<number, UnitStats> = {
  // ships (id из ships.yml)
  29: { attack: 5,    shield: 10,    shell: 4000 },
  30: { attack: 5,    shield: 25,    shell: 12000 },
  31: { attack: 50,   shield: 10,    shell: 4000 },
  32: { attack: 150,  shield: 25,    shell: 10000 },
  33: { attack: 400,  shield: 50,    shell: 27000 },
  34: { attack: 1000, shield: 200,   shell: 60000 },
  35: { attack: 700,  shield: 400,   shell: 70000 },
  36: { attack: 50,   shield: 100,   shell: 30000 },
  37: { attack: 1,    shield: 10,    shell: 16000 },
  38: { attack: 0,    shield: 0.01,  shell: 1000 },
  39: { attack: 1,    shield: 1,     shell: 2000 },
  40: { attack: 1000, shield: 500,   shell: 75000 },
  41: { attack: 2000, shield: 500,   shell: 110000 },
  42: { attack: 200000, shield: 50000, shell: 9000000 },
  // defense (id из defense.yml)
  43: { attack: 80,   shield: 20,    shell: 2000 },
  44: { attack: 100,  shield: 25,    shell: 2000 },
  45: { attack: 250,  shield: 100,   shell: 8000 },
  46: { attack: 150,  shield: 500,   shell: 8000 },
  47: { attack: 1100, shield: 200,   shell: 35000 },
  48: { attack: 3000, shield: 300,   shell: 100000 },
  49: { attack: 1,    shield: 2000,  shell: 20000 },
  50: { attack: 1,    shield: 10000, shell: 100000 },
};

interface TechSet {
  gun: number;
  shield: number;
  shell: number;
  ballistics: number;
  masking: number;
  shipyard: number;     // только для атакующего (не используется в bой)
  defenseFactory: number; // только для защитника
}

const ZERO_TECH: TechSet = {
  gun: 0, shield: 0, shell: 0, ballistics: 0, masking: 0,
  shipyard: 0, defenseFactory: 0,
};

interface UnitForm {
  qty: string;
  damaged: string;
  percent: string;
}

const ZERO_UNIT: UnitForm = { qty: '0', damaged: '0', percent: '100' };

export function SimulatorScreen() {
  const { t } = useTranslation();
  const { planetId } = useResolvedPlanet();
  const ships = catalogByGroup('ship');
  const defense = catalogByGroup('defense');

  const [aTech, setATech] = useState<TechSet>({ ...ZERO_TECH });
  const [dTech, setDTech] = useState<TechSet>({ ...ZERO_TECH });
  const [numSim, setNumSim] = useState('1');
  const [aUnits, setAUnits] = useState<Record<number, UnitForm>>({});
  const [dUnits, setDUnits] = useState<Record<number, UnitForm>>({});

  // Inventory для bulk-кнопки «Установить флот».
  const invQ = useQuery({
    queryKey: planetId ? QK.shipyardInventory(planetId) : ['noop-sim-inv'],
    queryFn: () =>
      planetId
        ? fetchShipyardInventory(planetId)
        : Promise.resolve({ ships: {}, defense: {} }),
    enabled: planetId !== null,
  });
  const inv = invQ.data ?? { ships: {}, defense: {} };

  // План 72.1 ч.20.11.7: НЕ редиректим на /battle-report. Показываем
  // сводку (Stats) + ссылку «Отчёт о сражении» — юзер запускает
  // симулятор много раз и сравнивает агрегаты.
  const sim = useMutation<SimRunResponse, Error, SimInput>({
    mutationFn: runSimulation,
  });

  function setUnitField(
    target: 'a' | 'd',
    id: number,
    field: keyof UnitForm,
    value: string,
  ) {
    const map = target === 'a' ? aUnits : dUnits;
    const setter = target === 'a' ? setAUnits : setDUnits;
    setter({ ...map, [id]: { ...(map[id] ?? ZERO_UNIT), [field]: value } });
  }

  function applyAutoIncFromInventory(
    target: 'a' | 'd',
    id: number,
    group: 'ships' | 'defense',
  ) {
    const stock = Number(
      (inv as Record<string, Record<string, number>>)[group]?.[String(id)] ?? 0,
    );
    setUnitField(target, id, 'qty', String(stock));
  }

  function setAllUnits(target: 'a' | 'd', group: 'ship' | 'defense') {
    const updates: Record<number, UnitForm> = {};
    const list = group === 'ship' ? ships : defense;
    const invMap =
      group === 'ship'
        ? ((inv.ships ?? {}) as Record<string, number>)
        : ((inv.defense ?? {}) as Record<string, number>);
    for (const entry of list) {
      updates[entry.id] = {
        qty: String(Number(invMap[String(entry.id)] ?? 0)),
        damaged: '0',
        percent: '100',
      };
    }
    const map = target === 'a' ? aUnits : dUnits;
    const setter = target === 'a' ? setAUnits : setDUnits;
    setter({ ...map, ...updates });
  }

  function setAllShellPercent(
    target: 'a' | 'd',
    group: 'ship' | 'defense',
    pct: number,
  ) {
    const map = target === 'a' ? aUnits : dUnits;
    const setter = target === 'a' ? setAUnits : setDUnits;
    const next: Record<number, UnitForm> = { ...map };
    const list = group === 'ship' ? ships : defense;
    for (const entry of list) {
      const cur = next[entry.id] ?? ZERO_UNIT;
      next[entry.id] = { ...cur, percent: String(pct) };
    }
    setter(next);
  }

  function resetAll(target: 'a' | 'd', group: 'ship' | 'defense') {
    const map = target === 'a' ? aUnits : dUnits;
    const setter = target === 'a' ? setAUnits : setDUnits;
    const next: Record<number, UnitForm> = { ...map };
    const list = group === 'ship' ? ships : defense;
    for (const entry of list) {
      next[entry.id] = { ...ZERO_UNIT };
    }
    setter(next);
  }

  function buildSide(units: Record<number, UnitForm>, tech: TechSet): SimUnit[] {
    const out: SimUnit[] = [];
    for (const idStr of Object.keys(units)) {
      const id = Number(idStr);
      const f = units[id];
      if (!f) continue;
      const qty = Math.max(0, Math.floor(Number(f.qty) || 0));
      if (qty <= 0) continue;
      const stats = UNIT_STATS[id];
      if (!stats) continue;
      const damaged = Math.max(0, Math.floor(Number(f.damaged) || 0));
      const percent = Math.max(0, Math.min(100, Math.floor(Number(f.percent) || 100)));
      out.push({
        unit_id: id,
        quantity: qty,
        damaged,
        shell_percent: percent / 100,
        attack: stats.attack,
        shield: stats.shield,
        shell: stats.shell,
      });
    }
    void tech;
    return out;
  }

  function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    const attackerUnits = buildSide(aUnits, aTech);
    const defenderUnits = buildSide(dUnits, dTech);
    if (attackerUnits.length === 0 || defenderUnits.length === 0) return;
    const attackers: SimSide[] = [
      {
        user_id: 'sim-attacker',
        tech: {
          gun: aTech.gun,
          shield: aTech.shield,
          shell: aTech.shell,
          ballistics: aTech.ballistics,
          masking: aTech.masking,
        },
        units: attackerUnits,
      },
    ];
    const defenders: SimSide[] = [
      {
        user_id: 'sim-defender',
        tech: {
          gun: dTech.gun,
          shield: dTech.shield,
          shell: dTech.shell,
          ballistics: dTech.ballistics,
          masking: dTech.masking,
        },
        units: defenderUnits,
      },
    ];
    sim.mutate({
      attackers,
      defenders,
      rounds: 6,
      num_sim: Math.max(1, Math.min(100, Number(numSim) || 1)),
    });
  }

  function entryName(id: number): string {
    const cat = findCatalog(id);
    if (!cat) return `#${id}`;
    const [g, k] = cat.i18n.split('.') as [string, string];
    return t(g, k);
  }

  return (
    <form onSubmit={onSubmit}>
      {/* Заголовок «Установки» */}
      <table className="ntable">
        <tbody>
          <tr>
            <th colSpan={2}>{t('info', 'options')}</th>
          </tr>
          <tr>
            <td>
              <table className="table_no_background" cellSpacing={0} cellPadding={0}>
                <tbody>
                  <tr>
                    <td>&nbsp;</td>
                    <td>{t('assaultReport', 'gunPower')}</td>
                    <td>{t('assaultReport', 'shieldPower')}</td>
                    <td>{t('assaultReport', 'armoring')}</td>
                    <td>{t('assaultReport', 'ballisticsPower')}</td>
                    <td>{t('assaultReport', 'maskingPower')}</td>
                    <td>{t('assaultReport', 'shipyardPower')}</td>
                    <td>{t('assaultReport', 'defenseFactoryPower')}</td>
                  </tr>
                  <tr>
                    <td>{t('mission', 'attacker') ?? 'Атакующий'}</td>
                    <td><TechInput v={aTech.gun} on={(v) => setATech({ ...aTech, gun: v })} /></td>
                    <td><TechInput v={aTech.shield} on={(v) => setATech({ ...aTech, shield: v })} /></td>
                    <td><TechInput v={aTech.shell} on={(v) => setATech({ ...aTech, shell: v })} /></td>
                    <td><TechInput v={aTech.ballistics} on={(v) => setATech({ ...aTech, ballistics: v })} /></td>
                    <td><TechInput v={aTech.masking} on={(v) => setATech({ ...aTech, masking: v })} /></td>
                    <td><TechInput v={aTech.shipyard} on={(v) => setATech({ ...aTech, shipyard: v })} /></td>
                    <td>&nbsp;</td>
                  </tr>
                  <tr>
                    <td>{t('mission', 'defender') ?? 'Защитник'}</td>
                    <td><TechInput v={dTech.gun} on={(v) => setDTech({ ...dTech, gun: v })} /></td>
                    <td><TechInput v={dTech.shield} on={(v) => setDTech({ ...dTech, shield: v })} /></td>
                    <td><TechInput v={dTech.shell} on={(v) => setDTech({ ...dTech, shell: v })} /></td>
                    <td><TechInput v={dTech.ballistics} on={(v) => setDTech({ ...dTech, ballistics: v })} /></td>
                    <td><TechInput v={dTech.masking} on={(v) => setDTech({ ...dTech, masking: v })} /></td>
                    <td>&nbsp;</td>
                    <td><TechInput v={dTech.defenseFactory} on={(v) => setDTech({ ...dTech, defenseFactory: v })} /></td>
                  </tr>
                </tbody>
              </table>
            </td>
            <td valign="top">
              <table className="table_no_background" cellSpacing={0} cellPadding={0}>
                <tbody>
                  <tr><td>{t('mission', 'numSim') ?? 'Количество симуляций'}</td></tr>
                  <tr>
                    <td>
                      <input
                        type="text"
                        size={2}
                        maxLength={2}
                        value={numSim}
                        onChange={(e) => setNumSim(e.target.value)}
                      />
                    </td>
                  </tr>
                </tbody>
              </table>
            </td>
          </tr>
        </tbody>
      </table>

      {/* Юниты */}
      <table className="ntable">
        <thead>
          <tr>
            <th>{t('mission', 'shipName') ?? 'Тип корабля'}</th>
            <th>{t('mission', 'attacker') ?? 'Атакующий'}</th>
            <th>{t('mission', 'defender') ?? 'Защитник'}</th>
          </tr>
        </thead>
        <tfoot>
          <tr>
            <td colSpan={3} className="center">
              <input
                type="submit"
                id="sim"
                name="simulate"
                className="button"
                value={t('mission', 'simulate') ?? 'Симулировать'}
                disabled={sim.isPending}
              />
            </td>
          </tr>
        </tfoot>
        <tbody>
          {/* Bulk-row для флота */}
          <tr>
            <th className="center">
              <a href="#" onClick={(e) => { e.preventDefault(); resetAll('a', 'ship'); resetAll('d', 'ship'); }}>
                {t('mission', 'resetFleet') ?? 'Обнулить флот'}
              </a>
            </th>
            <th className="center">
              <a href="#" onClick={(e) => { e.preventDefault(); setAllUnits('a', 'ship'); }}>
                {t('mission', 'setFleet') ?? 'Установить флот'}
              </a>
              <br />
              {[100, 90, 80, 70, 60].map((p) => (
                <span key={p}>
                  <a href="#" onClick={(e) => { e.preventDefault(); setAllShellPercent('a', 'ship', p); }}>
                    {p}%
                  </a>{' '}
                </span>
              ))}
            </th>
            <th className="center">
              <a href="#" onClick={(e) => { e.preventDefault(); setAllUnits('d', 'ship'); }}>
                {t('mission', 'setFleet') ?? 'Установить флот'}
              </a>
              <br />
              {[100, 90, 80, 70, 60].map((p) => (
                <span key={p}>
                  <a href="#" onClick={(e) => { e.preventDefault(); setAllShellPercent('d', 'ship', p); }}>
                    {p}%
                  </a>{' '}
                </span>
              ))}
            </th>
          </tr>

          {ships.map((entry) => (
            <UnitRow
              key={`ship-${entry.id}`}
              id={entry.id}
              name={entryName(entry.id)}
              aUnit={aUnits[entry.id] ?? ZERO_UNIT}
              dUnit={dUnits[entry.id] ?? ZERO_UNIT}
              onA={(field, v) => setUnitField('a', entry.id, field, v)}
              onD={(field, v) => setUnitField('d', entry.id, field, v)}
              onAInc={() => applyAutoIncFromInventory('a', entry.id, 'ships')}
              onDInc={() => applyAutoIncFromInventory('d', entry.id, 'ships')}
              onAReset={() => setUnitField('a', entry.id, 'qty', '0')}
              onDReset={() => setUnitField('d', entry.id, 'qty', '0')}
            />
          ))}

          {/* Bulk-row для обороны */}
          <tr>
            <th className="center">
              <a href="#" onClick={(e) => { e.preventDefault(); resetAll('d', 'defense'); }}>
                {t('mission', 'resetDefense') ?? 'Обнулить оборону'}
              </a>
            </th>
            <th className="center">&nbsp;</th>
            <th className="center">
              <a href="#" onClick={(e) => { e.preventDefault(); setAllUnits('d', 'defense'); }}>
                {t('mission', 'setDefense') ?? 'Установить оборону'}
              </a>
              <br />
              {[100, 90, 80, 70, 60].map((p) => (
                <span key={p}>
                  <a href="#" onClick={(e) => { e.preventDefault(); setAllShellPercent('d', 'defense', p); }}>
                    {p}%
                  </a>{' '}
                </span>
              ))}
            </th>
          </tr>

          {defense.map((entry) => (
            <UnitRow
              key={`def-${entry.id}`}
              id={entry.id}
              name={entryName(entry.id)}
              aUnit={null}
              dUnit={dUnits[entry.id] ?? ZERO_UNIT}
              onA={() => {}}
              onD={(field, v) => setUnitField('d', entry.id, field, v)}
              onAInc={() => {}}
              onDInc={() => applyAutoIncFromInventory('d', entry.id, 'defense')}
              onAReset={() => {}}
              onDReset={() => setUnitField('d', entry.id, 'qty', '0')}
            />
          ))}
        </tbody>
      </table>

      {/* Ошибки симуляции */}
      {sim.isError && (
        <div style={{ marginTop: 12, textAlign: 'center' }}>
          <span className="false">{(sim.error as Error)?.message ?? 'error'}</span>
        </div>
      )}

      {/* Блок «Результаты» — таблица сводки по N итераций (план 72.1
          ч.20.11.7+8). Выводится ПОСЛЕ кнопки «Симулировать», чтобы
          юзер видел результат под формой и мог запускать ещё. */}
      {sim.data && <SimResultsView resp={sim.data} t={t} />}
    </form>
  );
}

function TechInput({ v, on }: { v: number; on: (n: number) => void }) {
  return (
    <input
      type="text"
      size={2}
      maxLength={2}
      value={v}
      onChange={(e) => on(Math.max(0, Math.min(99, Number(e.target.value) || 0)))}
    />
  );
}

interface UnitRowProps {
  id: number;
  name: string;
  aUnit: UnitForm | null;
  dUnit: UnitForm;
  onA: (field: keyof UnitForm, v: string) => void;
  onD: (field: keyof UnitForm, v: string) => void;
  onAInc: () => void;
  onDInc: () => void;
  onAReset: () => void;
  onDReset: () => void;
}

function UnitRow({
  name, aUnit, dUnit, onA, onD, onAInc, onDInc, onAReset, onDReset,
}: UnitRowProps) {
  return (
    <tr>
      <td>{name}</td>
      <td className="center" valign="top">
        {aUnit ? (
          <UnitCell unit={aUnit} on={onA} onInc={onAInc} onReset={onAReset} />
        ) : (
          '—'
        )}
      </td>
      <td className="center" valign="top">
        <UnitCell unit={dUnit} on={onD} onInc={onDInc} onReset={onDReset} />
      </td>
    </tr>
  );
}

function UnitCell({
  unit, on, onInc, onReset,
}: {
  unit: UnitForm;
  on: (field: keyof UnitForm, v: string) => void;
  onInc: () => void;
  onReset: () => void;
}) {
  return (
    <table className="table_no_background" cellSpacing={0} cellPadding={0}>
      <tbody>
        <tr>
          <td>
            <a href="#" onClick={(e) => { e.preventDefault(); onInc(); }}>+</a>
          </td>
          <td rowSpan={2} style={{ whiteSpace: 'nowrap' }}>
            <input
              type="text"
              size={6}
              value={unit.qty}
              onChange={(e) => on('qty', e.target.value)}
            />
            {' ['}
            <input
              type="text"
              size={4}
              value={unit.damaged}
              onChange={(e) => on('damaged', e.target.value)}
            />
            {' - '}
            <input
              type="text"
              size={2}
              maxLength={3}
              value={unit.percent}
              onChange={(e) => on('percent', e.target.value)}
            />
            {'%]'}
          </td>
        </tr>
        <tr>
          <td>
            <a href="#" onClick={(e) => { e.preventDefault(); onReset(); }}>−</a>
          </td>
        </tr>
      </tbody>
    </table>
  );
}

// SimResultsView — блок «Результаты» (pixel-perfect клон legacy
// simulator.tpl, lines 101-160). Показывается после прогона
// num_sim симуляций, до формы — юзер видит агрегат и решает
// стоит ли менять параметры/переоткатить.
function SimResultsView({
  resp,
  t,
}: {
  resp: SimRunResponse;
  t: (g: string, k: string) => string | null;
}) {
  const s: SimStats = resp.stats;
  const fmt = (v: number, digits = 0) =>
    v.toLocaleString('ru-RU', {
      minimumFractionDigits: digits,
      maximumFractionDigits: digits,
    });
  // Время: до 1 с показываем в миллисекундах с округлением, иначе в
  // секундах с двумя знаками — иначе при быстром расчёте (<5 мс)
  // сводка уходила в «0,00 с».
  const fmtTime = (sec: number): string =>
    sec < 1
      ? `${fmt(sec * 1000, sec * 1000 < 10 ? 1 : 0)} мс`
      : `${fmt(sec, 2)} с`;

  const atkLossTotal =
    s.attacker_lost_metal + s.attacker_lost_silicon + s.attacker_lost_hydrogen;
  const defLossTotal =
    s.defender_lost_metal + s.defender_lost_silicon + s.defender_lost_hydrogen;

  return (
    <table className="ntable" style={{ marginTop: 12 }}>
      <thead>
        <tr>
          <th colSpan={3}>{t('assaultReport', 'summary') ?? 'Итоговый результат'}</th>
        </tr>
        <tr>
          <th style={{ width: '40%' }}>&nbsp;</th>
          <th style={{ width: '30%' }}>{t('mission', 'attacker') ?? 'Атакующий'}</th>
          <th style={{ width: '30%' }}>{t('mission', 'defender') ?? 'Защитник'}</th>
        </tr>
      </thead>
      <tbody>
        {/* Победы — в одной строке два значения */}
        <tr>
          <th>{t('assaultReport', 'attackerWon') ?? 'Победа'}</th>
          <td className={s.attacker_win_pct >= s.defender_win_pct ? 'true' : 'false'}>
            {fmt(s.attacker_win_pct, 1)}%
          </td>
          <td className={s.defender_win_pct > s.attacker_win_pct ? 'true' : 'false'}>
            {fmt(s.defender_win_pct, 1)}%
          </td>
        </tr>
        <tr>
          <th>{t('assaultReport', 'battleDraw') ?? 'Ничья'}</th>
          <td colSpan={2}>{fmt(s.draw_pct, 1)}%</td>
        </tr>
        <tr>
          <th>{t('mission', 'roundCount') ?? 'Раундов'}</th>
          <td colSpan={2}>{fmt(s.avg_rounds, 1)}</td>
        </tr>
        <tr>
          <th>{t('mission', 'moonChance') ?? 'Шанс луны'}</th>
          <td colSpan={2}>{fmt(s.avg_moon_chance, 1)}%</td>
        </tr>

        {/* Потери — две колонки */}
        <tr>
          <th>{t('mission', 'losses') ?? 'Потери'}</th>
          <td>
            <b>{fmt(atkLossTotal)}</b>
            <br />
            <small>
              {fmt(s.attacker_lost_metal)} мет /{' '}
              {fmt(s.attacker_lost_silicon)} крем /{' '}
              {fmt(s.attacker_lost_hydrogen)} вод
            </small>
          </td>
          <td>
            <b>{fmt(defLossTotal)}</b>
            <br />
            <small>
              {fmt(s.defender_lost_metal)} мет /{' '}
              {fmt(s.defender_lost_silicon)} крем /{' '}
              {fmt(s.defender_lost_hydrogen)} вод
            </small>
          </td>
        </tr>

        {/* Опыт */}
        <tr>
          <th>{t('mission', 'experience') ?? 'Опыт'}</th>
          <td>{fmt(s.attacker_exp, 1)}</td>
          <td>{fmt(s.defender_exp, 1)}</td>
        </tr>

        {/* Обломки на орбите — общая строка */}
        <tr>
          <th>{t('mission', 'debris') ?? 'Обломки на орбите'}</th>
          <td colSpan={2}>
            {fmt(s.debris_metal)} {t('assaultReport', 'and') ?? 'и'}{' '}
            {fmt(s.debris_silicon)}
            <small> (мет / крем)</small>
          </td>
        </tr>

        {/* Время */}
        <tr>
          <th>{t('mission', 'simTime') ?? 'Время'}</th>
          <td colSpan={2}>
            {fmtTime(s.gen_time_all)}{' · '}
            <small>
              {t('mission', 'oneSimTime') ?? 'одна симуляция'}:{' '}
              {fmtTime(s.gen_time)} ({fmt(s.num_sim)}×)
            </small>
          </td>
        </tr>

        {/* Ссылка на просмотрщик последнего боя */}
        <tr>
          <th>&nbsp;</th>
          <td colSpan={2}>
            <Link
              id="sim_report"
              className="false2"
              to={`/battle-report/${resp.id}`}
              target="_blank"
              rel="noopener noreferrer"
            >
              {t('assaultReport', 'assault') ?? 'Отчёт о сражении'}
            </Link>
          </td>
        </tr>
      </tbody>
    </table>
  );
}

