import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';
import { SHIPS, DEFENSE, nameOf } from '@/api/catalog';
import { useTranslation } from '@/i18n/i18n';
import type { Inventory, Planet } from '@/api/types';

// RepairScreen: две панели — «Ремонт повреждённых» (damaged>0)
// и «Разбор здоровых юнитов» (disassemble). Очередь одна общая
// (repair_queue), показывается снизу.
//
// REPAIR: кнопка «Починить» чинит всех damaged одного unit_id сразу
// (batch). Стоимость считается на сервере по legacy-формуле,
// клиенту показываем только доступность (есть ли damaged).

interface RepairQueueItem {
  id: string;
  planet_id: string;
  user_id: string;
  unit_id: number;
  is_defense: boolean;
  mode: 'disassemble' | 'repair';
  count: number;
  return_metal: number;
  return_silicon: number;
  return_hydrogen: number;
  per_unit_seconds: number;
  start_at: string;
  end_at: string;
  status: string;
}

interface DamagedUnit {
  unit_id: number;
  count: number;
  damaged: number;
  shell_percent: number;
}

export function RepairScreen({ planet }: { planet: Planet }) {
  const { t, tf } = useTranslation();
  const qc = useQueryClient();

  const inventory = useQuery({
    queryKey: ['shipyard-inventory', planet.id],
    queryFn: () => api.get<Inventory>(`/api/planets/${planet.id}/shipyard/inventory`),
  });
  const queue = useQuery({
    queryKey: ['repair-queue', planet.id],
    queryFn: () =>
      api.get<{ queue: RepairQueueItem[] | null }>(`/api/planets/${planet.id}/repair/queue`),
    refetchInterval: 2000,
  });

  const damaged = useQuery({
    queryKey: ['repair-damaged', planet.id],
    queryFn: () =>
      api.get<{ damaged: DamagedUnit[] | null }>(`/api/planets/${planet.id}/repair/damaged`),
    refetchInterval: 5000,
  });

  const repair = useMutation({
    mutationFn: (unitId: number) =>
      api.post<RepairQueueItem>(`/api/planets/${planet.id}/repair/repair`, {
        unit_id: unitId,
      }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['repair-queue', planet.id] });
      void qc.invalidateQueries({ queryKey: ['repair-damaged', planet.id] });
      void qc.invalidateQueries({ queryKey: ['shipyard-inventory', planet.id] });
      void qc.invalidateQueries({ queryKey: ['planets'] });
    },
  });

  const disassemble = useMutation({
    mutationFn: (p: { unitId: number; count: number }) =>
      api.post<RepairQueueItem>(`/api/planets/${planet.id}/repair/disassemble`, {
        unit_id: p.unitId,
        count: p.count,
      }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['repair-queue', planet.id] });
      void qc.invalidateQueries({ queryKey: ['shipyard-inventory', planet.id] });
      void qc.invalidateQueries({ queryKey: ['planets'] });
    },
  });

  const list = queue.data?.queue ?? [];
  const damagedList = damaged.data?.damaged ?? [];

  return (
    <section>
      <h2>
        {t('global', 'MENU_REPAIR')} — {planet.name}
      </h2>

      <h3>{tf('Main', 'REPAIR_DAMAGED_HEADER', 'Повреждённые корабли')}</h3>
      {damagedList.length === 0 ? (
        <p>{tf('Main', 'REPAIR_NO_DAMAGED', 'Нет повреждённых кораблей.')}</p>
      ) : (
        <table className="ox-table">
          <thead>
            <tr>
              <th>{tf('Main', 'UNIT', 'Юнит')}</th>
              <th>{tf('Main', 'IN_STOCK', 'В наличии')}</th>
              <th>{tf('Main', 'DAMAGED', 'Повреждено')}</th>
              <th>{tf('Main', 'SHELL_PERCENT', 'Броня (%)')}</th>
              <th>{tf('Main', 'ACTION', 'Действие')}</th>
            </tr>
          </thead>
          <tbody>
            {damagedList.map((d) => (
              <tr key={d.unit_id}>
                <td>{nameOf(d.unit_id)}</td>
                <td className="num">{d.count}</td>
                <td className="num">{d.damaged}</td>
                <td className="num">{Math.round(d.shell_percent)}</td>
                <td>
                  <button
                    type="button"
                    disabled={repair.isPending}
                    onClick={() => repair.mutate(d.unit_id)}
                  >
                    {tf('Main', 'REPAIR_FIX', 'Починить')}
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
      {repair.isError && (
        <div className="ox-error">
          {repair.error instanceof Error ? repair.error.message : t('global', 'ERROR')}
        </div>
      )}

      <h3>{tf('Main', 'REPAIR_DISASSEMBLE_HEADER', 'Разбор здоровых юнитов')}</h3>
      <p>
        {tf(
          'Main',
          'REPAIR_DISASSEMBLE_HINT',
          'Разбор здоровых юнитов возвращает ~70% стоимости (предварительно списываются 20% + юниты со стока, в конце очереди зачисляются 90%).',
        )}
      </p>

      <UnitList
        title={tf('Main', 'UNITS_SHIPS', 'Корабли')}
        units={SHIPS}
        stock={inventory.data?.ships}
        onGo={(id, n) => disassemble.mutate({ unitId: id, count: n })}
        pending={disassemble.isPending}
      />
      <UnitList
        title={tf('Main', 'UNITS_DEFENSE', 'Оборона')}
        units={DEFENSE}
        stock={inventory.data?.defense}
        onGo={(id, n) => disassemble.mutate({ unitId: id, count: n })}
        pending={disassemble.isPending}
      />

      {disassemble.isError && (
        <div className="ox-error">
          {disassemble.error instanceof Error ? disassemble.error.message : t('global', 'ERROR')}
        </div>
      )}

      <h3>{tf('Main', 'REPAIR_QUEUE', 'Очередь фабрики')}</h3>
      {list.length === 0 ? (
        <p>{tf('Main', 'QUEUE_EMPTY', 'Очередь пуста.')}</p>
      ) : (
        <table className="ox-table">
          <thead>
            <tr>
              <th>{tf('Main', 'UNIT', 'Юнит')}</th>
              <th>{tf('Main', 'COUNT', 'Количество')}</th>
              <th>{tf('Main', 'REPAIR_RETURN', 'Возврат (M/Si/H)')}</th>
              <th>{tf('Main', 'UNTIL', 'до')}</th>
            </tr>
          </thead>
          <tbody>
            {list.map((q) => (
              <tr key={q.id}>
                <td>{nameOf(q.unit_id)}</td>
                <td className="num">{q.count}</td>
                <td className="num">
                  {q.return_metal} / {q.return_silicon} / {q.return_hydrogen}
                </td>
                <td>{new Date(q.end_at).toLocaleTimeString('ru-RU')}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </section>
  );
}

function UnitList({
  title,
  units,
  stock,
  onGo,
  pending,
}: {
  title: string;
  units: { id: number; name: string }[];
  stock: Record<string, number> | undefined;
  onGo: (unitId: number, count: number) => void;
  pending: boolean;
}) {
  const { tf } = useTranslation();
  const [drafts, setDrafts] = useState<Record<number, number>>({});
  return (
    <>
      <h3>{title}</h3>
      <table className="ox-table">
        <thead>
          <tr>
            <th>{tf('Main', 'UNIT', 'Юнит')}</th>
            <th>{tf('Main', 'IN_STOCK', 'В наличии')}</th>
            <th>{tf('Main', 'COUNT', 'Количество')}</th>
            <th>{tf('Main', 'ACTION', 'Действие')}</th>
          </tr>
        </thead>
        <tbody>
          {units.map((u) => {
            const have = stock?.[u.id.toString()] ?? 0;
            const draft = drafts[u.id] ?? 0;
            return (
              <tr key={u.id}>
                <td>{u.name}</td>
                <td className="num">{have}</td>
                <td>
                  <input
                    type="number"
                    min={0}
                    max={have}
                    value={draft}
                    onChange={(e) =>
                      setDrafts({
                        ...drafts,
                        [u.id]: Math.max(0, Math.min(have, Number(e.target.value))),
                      })
                    }
                    style={{ width: 80 }}
                  />
                </td>
                <td>
                  <button
                    type="button"
                    disabled={pending || draft < 1 || draft > have}
                    onClick={() => onGo(u.id, draft)}
                  >
                    {tf('Main', 'REPAIR_DISASSEMBLE', 'Разобрать')}
                  </button>
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </>
  );
}
