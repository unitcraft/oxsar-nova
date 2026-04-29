// AppShell — обёртка origin-фронта.
//
// Структура повторяет layout legacy-PHP:
//   #leftMenu  (fixed top-left, 5/5)
//   #planets   (fixed top-right, 5/5)
//   #topHeader + #content (центральная колонка)
//   .oxsar-footer (fixed bottom)

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
    <>
      <LeftMenu />
      <PlanetsList />
      <div id="contentTopAndBody">
        <TopHeader />
        <div id="content">
          <div id="contentHtml">{children}</div>
        </div>
      </div>
      <Footer />
    </>
  );
}
