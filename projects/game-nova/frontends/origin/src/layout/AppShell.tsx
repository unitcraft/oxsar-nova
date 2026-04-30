// AppShell — обёртка origin-фронта.
//
// Структура повторяет layout legacy-PHP:
//   #leftMenu  (fixed top-left, 5/5)
//   #planets   (fixed top-right, 5/5)
//   #topHeader + #content (центральная колонка)
//   .oxsar-footer (fixed bottom)
//
// Для анонимных гостей (без токена, на публичных route типа
// /battle-report/:id) скрываем sidebar/planets/topheader — рисуем
// только #content + footer (план 72.1 ч.20.11).

import type { ReactNode } from 'react';
import { TopHeader } from './TopHeader';
import { LeftMenu } from './LeftMenu';
import { PlanetsList } from './PlanetsList';
import { Footer } from './Footer';
import { useAuthStore } from '@/stores/auth';

interface AppShellProps {
  children: ReactNode;
}

export function AppShell({ children }: AppShellProps) {
  const accessToken = useAuthStore((s) => s.accessToken);
  const isAuth = accessToken !== null;

  if (!isAuth) {
    // Анонимный режим: только контент + footer.
    return (
      <>
        <div id="content">
          <div id="contentHtml">{children}</div>
        </div>
        <Footer />
      </>
    );
  }

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
