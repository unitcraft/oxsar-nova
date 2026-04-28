// Левое меню origin-фронта (план 72 Ф.1).
//
// Воспроизводит legacy `#leftMenu` — фиксированная боковая навигация
// шириной 160px с группами: производство (menu-prod), пользовательские
// разделы (menu-user), прочее (menu-other).
//
// На Ф.1 — каркас со ссылками на основные экраны Spring 1 (Main,
// Constructions, Research, Shipyard, Galaxy, Mission, Empire). Полный
// набор разделов добавляется по мере реализации экранов в Spring 2-5.
//
// Намеренные расхождения с legacy:
//   - **Achievements** — пункт скрыт (план 70 отложен, см. шапку плана 72).
//   - **Tutorial** — пункт скрыт (тот же раздел плана 72 «не делаем»).
//   - **Реклама/баннеры** — не переносятся.
//   - **Реферальный экран** — отдельного пункта нет, ведёт в portal
//     (план 59) — будет реализовано на этапе Spring 4.

import type { ReactNode } from 'react';

interface MenuGroupProps {
  className: string;
  label: string;
  children?: ReactNode;
}

function MenuGroup({ className, label, children }: MenuGroupProps) {
  return (
    <>
      <li className={className}>{label}</li>
      {children}
    </>
  );
}

interface MenuLinkProps {
  href: string;
  label: string;
}

function MenuLink({ href, label }: MenuLinkProps) {
  return (
    <li>
      <a href={href}>{label}</a>
    </li>
  );
}

export function LeftMenu() {
  return (
    <div id="leftMenu">
      <ul>
        <MenuGroup className="menu-info" label="Империя" />
        <MenuLink href="#main" label="Обзор" />
        <MenuLink href="#empire" label="Империя" />
        <MenuLink href="#resource" label="Ресурсы" />

        <MenuGroup className="menu-prod" label="Производство" />
        <MenuLink href="#constructions" label="Строения" />
        <MenuLink href="#research" label="Исследования" />
        <MenuLink href="#shipyard" label="Верфь" />
        <MenuLink href="#repair" label="Ремонт" />

        <MenuGroup className="menu-user" label="Игрок" />
        <MenuLink href="#galaxy" label="Галактика" />
        <MenuLink href="#mission" label="Миссии" />
        <MenuLink href="#fleet" label="Флот" />
        <MenuLink href="#alliance" label="Альянс" />

        <MenuGroup className="menu-other" label="Прочее" />
        <MenuLink href="#chat" label="Чат" />
        <MenuLink href="#msg" label="Сообщения" />
        <MenuLink href="#friends" label="Друзья" />
        <MenuLink href="#statistics" label="Статистика" />
        <MenuLink href="#settings" label="Настройки" />
      </ul>
    </div>
  );
}
