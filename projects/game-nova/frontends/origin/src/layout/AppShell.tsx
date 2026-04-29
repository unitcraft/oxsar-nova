// AppShell — 3-frame обёртка origin-фронта (план 72 Ф.1).
//
// Структура повторяет layout legacy-PHP:
//   #topHeader (margin 10px, full width)
//   #leftMenu  (fixed top-left)
//   #planets   (absolute top-right)
//   #content   (main area, между left и right)
//   .oxsar-footer (fixed bottom)
//
// В отличие от nova-фронта (где есть SPA-роутер + tabs), origin
// сохраняет legacy-style hash-навигацию (#main, #constructions, ...).
// Конкретный роутинг подключим в Ф.2 при добавлении первых экранов;
// на Ф.1 #content получает children и рисует Bootstrap-заглушку.

import type { ReactNode } from 'react';
import { TopHeader } from './TopHeader';
import { LeftMenu } from './LeftMenu';
import { PlanetsList } from './PlanetsList';
import { Footer } from './Footer';

interface AppShellProps {
  children: ReactNode;
}

export function AppShell({ children }: AppShellProps) {
  return (
    <div id="contentHtml">
      <div className="main_content">
        <LeftMenu />
        <PlanetsList />
        <div id="contentTopAndBody">
          <TopHeader />
          <div id="content">{children}</div>
        </div>
        <Footer />
      </div>
    </div>
  );
}
