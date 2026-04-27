// AgeRating — возрастная маркировка по 436-ФЗ (план 46).
// "12+" по методическим рекомендациям РКН для стратегий
// без натуралистичных сцен жестокости.
export function AgeRating({ size = 'sm' }: { size?: 'sm' | 'md' }) {
  const fontSize = size === 'md' ? 13 : 11;
  const padding = size === 'md' ? '2px 7px' : '1px 5px';
  return (
    <span
      title="Информационная продукция для лиц старше 12 лет (ФЗ № 436)"
      aria-label="Возрастная маркировка: 12 плюс"
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        justifyContent: 'center',
        border: '1px solid var(--ox-border)',
        borderRadius: 'var(--ox-r)',
        color: 'var(--ox-fg-muted)',
        fontFamily: 'var(--ox-mono)',
        fontWeight: 700,
        fontSize,
        padding,
        letterSpacing: '0.04em',
        userSelect: 'none',
      }}
    >
      12+
    </span>
  );
}
