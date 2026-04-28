// X-001: дефицит ресурсов с явной пометкой `(нужно X)`.
//
// Origin (required_res_table.tpl): рядом с каждой суммой ресурса,
// если её не хватает — добавляется `(дефицит)` маленьким шрифтом.
// Отображаем то же: иконка ресурса, число, и при дефиците — скобки
// со знаком минус. Цвета через CSS-переменные nova theme.
//
// Применяется в BuildingsScreen, ResearchScreen, ShipyardScreen,
// RepairScreen — везде, где есть стоимость и проверка «хватит ли».

import { computeDeficit, type Cost, type ResourcesAvailable } from './feedback';

interface ResourceCostLineProps {
  cost: Cost;
  have: ResourcesAvailable;
}

// ResourceCostLine — основная стоимость в одну строку с раскраской
// каждого ресурса по «хватает/не хватает». X-001 базовый случай.
export function ResourceCostLine({ cost, have }: ResourceCostLineProps) {
  return (
    <div style={{ fontSize: 13, fontFamily: 'var(--ox-mono)', lineHeight: 1.6 }}>
      {cost.metal > 0 && (
        <span style={{ marginRight: 6, color: have.metal >= cost.metal ? 'var(--ox-fg-dim)' : 'var(--ox-danger)' }}>
          🟠{cost.metal.toLocaleString('ru-RU')}
        </span>
      )}
      {cost.silicon > 0 && (
        <span style={{ marginRight: 6, color: have.silicon >= cost.silicon ? 'var(--ox-fg-dim)' : 'var(--ox-danger)' }}>
          💎{cost.silicon.toLocaleString('ru-RU')}
        </span>
      )}
      {cost.hydrogen > 0 && (
        <span style={{ color: have.hydrogen >= cost.hydrogen ? 'var(--ox-fg-dim)' : 'var(--ox-danger)' }}>
          💧{cost.hydrogen.toLocaleString('ru-RU')}
        </span>
      )}
    </div>
  );
}

// ResourceDeficitBadge — вторая строка под стоимостью, видна только
// при дефиците хотя бы по одному ресурсу. Origin показывает дефицит
// в скобках; здесь — отдельной строкой с явным минусом, читается
// мгновенно: 🟠−2,500 💎−1,800.
export function ResourceDeficitBadge({ cost, have }: ResourceCostLineProps) {
  const d = computeDeficit(cost, have);
  if (d.metal === 0 && d.silicon === 0 && d.hydrogen === 0) return null;
  const parts: string[] = [];
  if (d.metal > 0)    parts.push(`🟠−${d.metal.toLocaleString('ru-RU')}`);
  if (d.silicon > 0)  parts.push(`💎−${d.silicon.toLocaleString('ru-RU')}`);
  if (d.hydrogen > 0) parts.push(`💧−${d.hydrogen.toLocaleString('ru-RU')}`);
  return (
    <div style={{ fontSize: 10, color: 'var(--ox-danger)', marginTop: 2, fontFamily: 'var(--ox-mono)' }}>
      {parts.join(' ')}
    </div>
  );
}
