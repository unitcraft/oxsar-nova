// Список планет (правая колонка) origin-фронта (план 72 Ф.1).
//
// Воспроизводит legacy `#planets` — absolute-блок справа со списком
// планет/лун игрока (cur-planet / cur-moon с фоновым PNG-спрайтом
// в зависимости от состояния).
//
// На Ф.1 — каркас без данных. Подключение к /api/planets — Spring 1
// (Main экран).

export function PlanetsList() {
  return (
    <ul id="planets">
      {/* Spring 1: подгрузить /api/planets и рендерить .cur-planet / .cur-moon */}
    </ul>
  );
}
