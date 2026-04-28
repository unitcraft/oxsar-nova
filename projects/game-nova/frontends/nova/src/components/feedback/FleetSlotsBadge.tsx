// X-007: счётчик слотов «X из Y» с явным сигналом «нет свободных».
// Origin (missions3.tpl): если can_send_fleet=false, показывает
// текст NO_FREE_FLEET_SLOTS. У nova поля slots_used/slots_max уже
// есть в /api/fleet, и этот badge даёт единый стиль с
// предупреждением при approaching и full.

import { slotsState } from './feedback';

interface SlotsBadgeProps {
  used: number;
  max: number;
  // labelTitle — название группы слотов («Флот», «Экспедиция»).
  labelTitle: string;
  // labelFull — текст при заполнении (например, «нет свободных»).
  labelFull?: string;
  // labelHint — мелкая подсказка под бейджем (например, «увеличить
  // computer_tech»). Опциональная.
  labelHint?: string;
}

export function FleetSlotsBadge({ used, max, labelTitle, labelFull, labelHint }: SlotsBadgeProps) {
  const state = slotsState(used, max);
  const color =
    state === 'full'   ? 'var(--ox-danger)' :
    state === 'almost' ? 'var(--ox-warn, #f59e0b)' :
                         'var(--ox-fg)';
  return (
    <div className="ox-panel" style={{ padding: '8px 16px', fontSize: 13, color: 'var(--ox-fg-muted)' }}>
      {labelTitle}{' '}
      <strong style={{ color, fontFamily: 'var(--ox-mono)' }}>{used} / {max}</strong>
      {state === 'full' && labelFull && (
        <span style={{ marginLeft: 8, color: 'var(--ox-danger)' }}>· {labelFull}</span>
      )}
      {labelHint && <span style={{ marginLeft: 8, opacity: 0.7 }}>{labelHint}</span>}
    </div>
  );
}
