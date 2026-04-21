import { useState } from 'react';
import { useMutation } from '@tanstack/react-query';
import { api } from '@/api/client';
import { SHIPS, DEFENSE, nameOf } from '@/api/catalog';
import { useTranslation } from '@/i18n/i18n';

// Экран симулятора боя. Пока минимальный: два списка юнитов (атакующий
// и защитник), кнопка «рассчитать». Бек возвращает плейсхолдер-отчёт
// (каркас боя, M4). Когда движок будет полностью портирован, UI
// подхватит без переделок — он работает через /api/battle-sim.

type UnitMap = Record<number, number>;

interface SimReport {
  seed: number;
  winner: 'attackers' | 'defenders' | 'draw';
  rounds: number;
}

export function BattleSimScreen() {
  const { t, tf } = useTranslation();
  const [attackers, setAttackers] = useState<UnitMap>({});
  const [defenders, setDefenders] = useState<UnitMap>({});
  const [seed, setSeed] = useState<number>(42);

  const sim = useMutation({
    mutationFn: (body: unknown) => api.post<SimReport>('/api/battle-sim', body),
  });

  function runSim() {
    const toSide = (m: UnitMap) => [
      {
        user_id: 'sim',
        units: Object.entries(m)
          .filter(([, count]) => count > 0)
          .map(([unitId, count]) => ({
            unit_id: Number(unitId),
            quantity: count,
            front: 10,
            attack: [100, 0, 0],
            shield: [10, 0, 0],
            shell: 4000,
          })),
      },
    ];
    sim.mutate({
      seed,
      attackers: toSide(attackers),
      defenders: toSide(defenders),
    });
  }

  return (
    <section>
      <h2>{t('global', 'MENU_SIMULATOR')}</h2>
      <p>
        {tf(
          'Main',
          'BATTLE_SIM_WIP',
          'Внимание: бой ещё портируется (M4). Сейчас endpoint возвращает «draw» на любом входе — это каркас, а не реальный расчёт. Скрин/API стабильные, логика придёт вместе с портом game/Assault.class.php.',
        )}
      </p>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
        <UnitPicker
          title={tf('Main', 'BATTLE_SIM_ATTACKERS', 'Атакующий флот')}
          units={[...SHIPS]}
          value={attackers}
          onChange={setAttackers}
        />
        <UnitPicker
          title={tf('Main', 'BATTLE_SIM_DEFENDERS', 'Защита + флот')}
          units={[...SHIPS, ...DEFENSE]}
          value={defenders}
          onChange={setDefenders}
        />
      </div>

      <label style={{ display: 'block', margin: '12px 0' }}>
        {tf('Main', 'BATTLE_SIM_SEED', 'Seed:')}&nbsp;
        <input
          type="number"
          value={seed}
          onChange={(e) => setSeed(Number(e.target.value))}
          style={{ width: 120 }}
        />
      </label>

      <button type="button" disabled={sim.isPending} onClick={runSim}>
        {sim.isPending
          ? tf('Main', 'BATTLE_SIM_CALCULATING', 'Считаем…')
          : tf('Main', 'BATTLE_SIM_RUN', 'Рассчитать')}
      </button>

      {sim.isError && (
        <div className="ox-error">
          {sim.error instanceof Error ? sim.error.message : t('global', 'ERROR')}
        </div>
      )}
      {sim.data && (
        <div style={{ marginTop: 12 }}>
          <h3>{tf('Main', 'BATTLE_SIM_RESULT', 'Результат')}</h3>
          <ul>
            <li>{tf('Main', 'BATTLE_SIM_SEED', 'Seed:')} {sim.data.seed}</li>
            <li>{tf('Main', 'BATTLE_SIM_ROUNDS', 'Раундов:')} {sim.data.rounds}</li>
            <li>{tf('Main', 'BATTLE_SIM_WINNER', 'Победитель:')} {sim.data.winner}</li>
          </ul>
        </div>
      )}
    </section>
  );
}

function UnitPicker({
  title,
  units,
  value,
  onChange,
}: {
  title: string;
  units: { id: number; name: string }[];
  value: UnitMap;
  onChange: (v: UnitMap) => void;
}) {
  return (
    <div>
      <h3>{title}</h3>
      <table className="ox-table">
        <thead>
          <tr>
            <th>Юнит</th>
            <th>Кол-во</th>
          </tr>
        </thead>
        <tbody>
          {units.map((u) => (
            <tr key={u.id}>
              <td>{nameOf(u.id)}</td>
              <td>
                <input
                  type="number"
                  min={0}
                  value={value[u.id] ?? 0}
                  onChange={(e) => onChange({ ...value, [u.id]: Math.max(0, Number(e.target.value)) })}
                  style={{ width: 100 }}
                />
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
