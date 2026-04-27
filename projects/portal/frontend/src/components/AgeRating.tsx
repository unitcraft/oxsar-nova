import styles from './AgeRating.module.css';

// AgeRating — возрастная маркировка по 436-ФЗ.
// "12+" по методическим рекомендациям РКН для стратегий
// без натуралистичных сцен жестокости. См. план 46.
export function AgeRating({ size = 'sm' }: { size?: 'sm' | 'md' }) {
  return (
    <span
      className={`${styles.badge} ${styles[size]}`}
      title="Информационная продукция для лиц старше 12 лет (ФЗ № 436)"
      aria-label="Возрастная маркировка: 12 плюс"
    >
      12+
    </span>
  );
}
