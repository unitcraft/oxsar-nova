// Footer origin-фронта (план 72 Ф.1).
//
// Реализует требование плана 50 Ф.5 — юр-ссылки + 12+ маркировка.
// Pixel-perfect клон layout.tpl footer-блока legacy-PHP.

const PORTAL = 'https://oxsar-nova.ru';

export function Footer() {
  return (
    <div className="oxsar-footer">
      <div className="age-rating" title="Возрастная категория 12+">12+</div>
      <div className="legal-links">
        <a href={`${PORTAL}/offer`} target="_blank" rel="noopener noreferrer">Оферта</a>
        <span className="sep">|</span>
        <a href={`${PORTAL}/game-rules`} target="_blank" rel="noopener noreferrer">Правила</a>
        <span className="sep">|</span>
        <a href={`${PORTAL}/refund`} target="_blank" rel="noopener noreferrer">Возврат</a>
        <span className="sep">|</span>
        <a href={`${PORTAL}/privacy`} target="_blank" rel="noopener noreferrer">Конфиденциальность</a>
      </div>
    </div>
  );
}
