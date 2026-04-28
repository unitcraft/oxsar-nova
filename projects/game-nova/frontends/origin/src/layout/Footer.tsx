// Footer origin-фронта (план 72 Ф.1).
//
// Реализует требование плана 50 Ф.5 — юр-ссылки + 12+ маркировка.
// Стилизация (.oxsar-footer) — fixed bottom между #leftMenu и
// #planets, см. layout.css.
//
// Юридические ссылки ведут на portal (внешний домен в проде —
// oxsar-nova.ru). Origin-фронт не реплицирует юр-документы.

export function Footer() {
  return (
    <div className="oxsar-footer">
      <a href="https://oxsar-nova.ru/legal/terms" target="_blank" rel="noreferrer">
        Условия
      </a>
      {' · '}
      <a href="https://oxsar-nova.ru/legal/privacy" target="_blank" rel="noreferrer">
        Конфиденциальность
      </a>
      {' · '}
      <a href="https://oxsar-nova.ru/legal/contacts" target="_blank" rel="noreferrer">
        Контакты
      </a>
      <span className="age-rating">12+</span>
    </div>
  );
}
