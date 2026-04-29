// S-044 Changelog — история обновлений (план 72 Ф.5 Spring 4 ч.2).
//
// Pixel-perfect зеркало legacy `templates/standard/changelog.tpl`:
//   <table class="ntable">
//     <thead><tr><th>RELEASE</th><th>CHANGES</th></tr></thead>
//     <tbody>
//       {foreach release}<tr><td>version</td><td><pre>changes</pre></td></tr>
//
// Источник данных — bundled markdown (CHANGELOG.md рядом). Endpoint
// `/api/changelog` в game-nova-backend отсутствует — это намеренно
// (план 72 Ф.5 Spring 4 ч.2 simplifications P72.S4.CHANGELOG):
// changelog меняется при релизах, в БД таблицу заводить нет смысла,
// markdown в bundle — стандарт для редко-меняющегося контента.
//
// Lightweight markdown-парсер: парсим `## version — date` как заголовки
// релизов, остальное — body. Без полноценного MD-renderer'а (react-
// markdown добавит ~30 KB), потому что у нас фиксированный формат
// CHANGELOG.md, который мы сами и пишем.

import rawChangelog from './CHANGELOG.md?raw';
import { parseChangelog } from './parse';
import { useTranslation } from '@/i18n/i18n';

export function ChangelogScreen() {
  const { t } = useTranslation();
  const releases = parseChangelog(rawChangelog);

  return (
    <table className="ntable">
      <thead>
        <tr>
          <th>{t('main', 'release')}</th>
          <th>{t('main', 'changes')}</th>
        </tr>
      </thead>
      <tbody>
        {releases.map((r) => (
          <tr key={r.version}>
            <td>{r.version}</td>
            <td>
              <pre style={{ whiteSpace: 'pre-wrap', margin: 0 }}>
                {r.changes}
              </pre>
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}
