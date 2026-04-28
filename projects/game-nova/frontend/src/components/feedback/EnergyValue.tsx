// X-010: энергодефицит — totalEnergy <= 0 красным.
// X-002: производство положительное → зелёное, потребление → красное.
//
// Origin (resource.tpl): класс .true/.false для каждой ячейки
// производства и отдельная логика для общего баланса энергии.
// Единый компонент NumValue делает то же через CSS-переменные.

import { numKind, energyKind } from './feedback';

interface NumValueProps {
  value: number;
  // formatter — отображение числа (без знака рендерим, чтобы
  // могли подставить '—' для нуля или сократить до K/M).
  formatter?: (v: number) => string;
  // muted — приглушить цвет нуля (по умолчанию dim-серый).
  muted?: boolean;
}

const COLOR: Record<'positive' | 'negative' | 'zero', string> = {
  positive: 'var(--ox-success)',
  negative: 'var(--ox-danger)',
  zero:     'var(--ox-fg-dim)',
};

// NumValue — общий компонент для X-002 (потребление/производство).
// X-002 в origin различает producer (true=зелёный) и consumer
// (false=красный) — у нас это знак числа: positive = производство,
// negative = потребление.
export function NumValue({ value, formatter, muted = false }: NumValueProps) {
  const kind = numKind(value);
  const text = formatter ? formatter(value) : value.toLocaleString('ru-RU');
  return (
    <span style={{ color: muted && kind === 'zero' ? 'var(--ox-fg-muted)' : COLOR[kind] }}>
      {text}
    </span>
  );
}

interface EnergyValueProps {
  totalEnergy: number;
  formatter?: (v: number) => string;
}

// EnergyValue — X-010. Отличие от NumValue: при totalEnergy === 0
// тоже красим в красный (origin: `<= 0`), потому что нулевой баланс
// при любом потреблении — уже срыв.
export function EnergyValue({ totalEnergy, formatter }: EnergyValueProps) {
  const kind = energyKind(totalEnergy);
  const text = formatter ? formatter(totalEnergy) : Math.round(totalEnergy).toLocaleString('ru-RU');
  return (
    <span style={{
      color: kind === 'deficit' ? 'var(--ox-danger)' : 'var(--ox-success)',
      fontWeight: kind === 'deficit' ? 700 : 500,
    }}>
      {text}
    </span>
  );
}
