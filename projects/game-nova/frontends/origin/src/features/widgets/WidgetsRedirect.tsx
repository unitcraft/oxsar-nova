// S-046 Widgets — закрыт через S-001 Main (план 72 Ф.5 Spring 4 ч.2).
//
// Legacy `templates/standard/widgets.tpl` оказался почти пустым
// (Yii widget CurrentEvents удалён, см. план 37.5d.9). Семантически
// "виджеты на главной" в origin-фронте уже агрегированы в S-001
// MainScreen (Spring 1, коммит 47d1f0ef65) — фронт показывает события,
// флоты, непрочитанные сообщения сразу на /.
//
// Поэтому /widgets делает redirect на /. См. simplifications.md
// P72.S4.WIDGETS — TRADE-OFF (R15 ✅, не упрощение): визуальное
// расхождение с legacy (нет отдельного маршрута), но семантически
// эквивалент сохранён.

import { useEffect } from 'react';
import { Navigate } from 'react-router-dom';

export function WidgetsRedirect() {
  useEffect(() => {
    if (typeof console !== 'undefined') {
      // dev-notice — для разработчиков, объясняющий почему /widgets
      // редиректит на /. В prod-bundle логи не критичны.
      console.info(
        '[origin] S-046 /widgets устарел — данные перенесены в S-001 MainScreen (/).',
      );
    }
  }, []);
  return <Navigate to="/" replace />;
}
