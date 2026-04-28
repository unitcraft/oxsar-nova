// Корневой компонент origin-фронта (план 72 Ф.1).
//
// На Ф.1 рендерится только Bootstrap-заглушка внутри AppShell —
// она проверяет, что 3-frame layout, тема, шрифты и ассеты
// загружаются корректно. Реальные экраны добавляются в Ф.2-Ф.6
// (Spring 1-5).

import { AppShell } from './layout/AppShell';

export function App() {
  return (
    <AppShell>
      <div className="topbox">Origin-frontend Bootstrap (план 72 Ф.1)</div>
      <div className="idiv">
        Каркас 3-frame layout. Экраны Spring 1-5 добавляются отдельными
        итерациями. Тема: pixel-perfect клон legacy-PHP standard.
      </div>
    </AppShell>
  );
}
