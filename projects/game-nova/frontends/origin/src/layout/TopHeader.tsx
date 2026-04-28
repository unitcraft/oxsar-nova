// Top header origin-фронта (план 72 Ф.1).
//
// Воспроизводит legacy `topHeader` (см.
// projects/game-legacy-php/src/templates/standard/before_content.tpl
// + layout.css). Содержимое: ресурсы (баланс), username, language,
// logout. На Ф.1 — каркас с placeholder'ами; реальные данные
// привяжутся в Spring 1 (Main экран).

import { useAuthStore } from '@/stores/auth';

export function TopHeader() {
  const userId = useAuthStore((s) => s.userId);
  const logout = useAuthStore((s) => s.logout);

  return (
    <div id="topHeader">
      <ul>
        <li className="ressource">
          <span>—</span>
        </li>
        <li>
          <span>{userId ?? '—'}</span>
        </li>
        <li>
          <select defaultValue="ru" disabled>
            <option value="ru">RU</option>
          </select>
        </li>
        <li>
          <button type="button" className="button" onClick={logout}>
            Logout
          </button>
        </li>
      </ul>
    </div>
  );
}
