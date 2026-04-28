// S-008 Alliance overview (план 72 Ф.3 Spring 2 ч.1).
//
// Pixel-perfect зеркало legacy `templates/standard/ally.tpl` —
// landing-страница раздела «Альянсы».
//
// Поведение:
//   - Если игрок в альянсе — редирект на /alliance/me.
//   - Иначе: 2 кнопки [Создать альянс] / [Найти альянс] и блок
//     «Текущие заявки» (legacy applications list).
//
// Endpoint: GET /api/alliances/me. Заявки текущего игрока (own
// applications) — в legacy отдельный controller и table; в nova-API
// аналога нет → отмечено в simplifications.md (P72.S2.A) и блок
// заявок в первой итерации остаётся пустым placeholder'ом.

import { Link, Navigate } from 'react-router-dom';
import { useTranslation } from '@/i18n/i18n';
import { useMyAlliance } from './common';

export function AllianceOverviewScreen() {
  const { t } = useTranslation();
  const { data, isLoading } = useMyAlliance();

  if (isLoading) {
    return <div className="idiv">…</div>;
  }
  if (data) {
    return <Navigate to="/alliance/me" replace />;
  }

  return (
    <table className="ntable">
      <thead>
        <tr>
          <th colSpan={2}>{t('alliance', 'alliances')}</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td className="center">
            <Link to="/alliance/create">{t('alliance', 'foundAlliance')}</Link>
          </td>
          <td className="center">
            <Link to="/alliance/list">{t('alliance', 'joinAlliance')}</Link>
          </td>
        </tr>
      </tbody>
    </table>
  );
}
