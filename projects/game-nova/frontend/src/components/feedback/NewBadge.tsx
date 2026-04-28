// X-021: счётчик новых ачивок (или иных новинок) — зелёный жирный
// бейдж с числом.
// Origin (achievements.tpl): `+{real_new_achieve_count}`.
// У nova achievements реактивируются по плану 70 (отложено) — пока
// рендерим только если count > 0, источник count подключим когда
// поле появится в DTO.

interface NewBadgeProps {
  count: number;
  // ariaLabel — для скринридеров; например «3 новых достижения».
  ariaLabel?: string;
}

export function NewBadge({ count, ariaLabel }: NewBadgeProps) {
  if (count <= 0) return null;
  return (
    <span
      aria-label={ariaLabel}
      style={{
        display: 'inline-block',
        marginLeft: 6,
        padding: '0 5px',
        fontSize: 11,
        fontWeight: 700,
        lineHeight: '14px',
        color: 'var(--ox-bg)',
        background: 'var(--ox-success)',
        borderRadius: 4,
        fontFamily: 'var(--ox-mono)',
      }}
    >
      +{count}
    </span>
  );
}
