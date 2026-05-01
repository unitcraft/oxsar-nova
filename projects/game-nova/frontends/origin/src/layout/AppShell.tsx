// AppShell — обёртка origin-фронта.
//
// Структура повторяет layout legacy-PHP:
//   #leftMenu  (fixed top-left, 5/5)
//   #planets   (fixed top-right, 5/5)
//   #topHeader + #content (центральная колонка)
//   .oxsar-footer (fixed bottom)
//
// «Bare»-режим (без обвязки) включается для:
//   - анонимных гостей (нет accessToken) — план 72.1 ч.20.11.6;
//   - просмотра боевого отчёта /battle-report/:id даже у
//     авторизованных юзеров (план 72.1 ч.20.11.12) — legacy
//     показывает страницу отчёта без меню/ресурсов/планет.

import type { ReactNode } from 'react';
import { useLocation } from 'react-router-dom';
import { TopHeader } from './TopHeader';
import { LeftMenu } from './LeftMenu';
import { PlanetsList } from './PlanetsList';
import { Footer } from './Footer';
import { useAuthStore } from '@/stores/auth';

interface AppShellProps {
  children: ReactNode;
}

// BARE_ROUTES — pathname'ы, на которых обвязка скрывается всегда.
const BARE_ROUTES: RegExp[] = [
  /^\/battle-report\/[0-9a-f-]+$/,
];

function isBareRoute(pathname: string): boolean {
  return BARE_ROUTES.some((re) => re.test(pathname));
}

export function AppShell({ children }: AppShellProps) {
  const accessToken = useAuthStore((s) => s.accessToken);
  const { pathname } = useLocation();
  const isAuth = accessToken !== null;
  const bare = !isAuth || isBareRoute(pathname);

  if (bare) {
    // Bare-режим: только контент + footer.
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
