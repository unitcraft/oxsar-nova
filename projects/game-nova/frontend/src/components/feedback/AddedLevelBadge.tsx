// X-013: показывает added_level рядом с базовым уровнем как «(+2)»
// или «(−1)» зелёным/красным.
//
// Origin (research.tpl, constructions.tpl): при наличии бонусов от
// артефактов / технологий уровень показывается как `5 (+2)`. У нас
// nova-DTO планеты содержит build_factor / produce_factor / research_factor,
// но added_level — это разница между base level и effective level
// для конкретного юнита. Этот компонент готов к моменту, когда DTO
// будет расширен (R2 — добавим поля в openapi.yaml). Пока вызывающий
// код передаёт уже посчитанное значение или 0 (компонент не рендерит).

import { addedLevelKind, formatAddedLevel } from './feedback';

interface AddedLevelBadgeProps {
  added: number;
}

export function AddedLevelBadge({ added }: AddedLevelBadgeProps) {
  const kind = addedLevelKind(added);
  if (kind === 'none') return null;
  return (
    <span style={{
      marginLeft: 4,
      fontSize: 12,
      fontFamily: 'var(--ox-mono)',
      color: kind === 'positive' ? 'var(--ox-success)' : 'var(--ox-danger)',
    }}>
      ({formatAddedLevel(added)})
    </span>
  );
}
