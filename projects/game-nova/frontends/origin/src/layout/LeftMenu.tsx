// Левое меню origin-фронта (план 72 Ф.1 + Ф.2 Spring 1 + Ф.3 Spring 2).
//
// Spring 1: Обзор, Империя, Строения, Исследования, Верфь, Галактика,
// Миссии.
// Spring 2 ч.1: Альянс. Spring 2 ч.2: Ресурсы (рынок ресурсов), Ремонт,
// Флот (операции), Боевая статистика. Чат/Сообщения/Друзья/Статистика/
// Настройки остаются заглушками до Spring 3-5.
//
// Намеренные расхождения с legacy:
//   - **Achievements**, **Tutorial** — пункты скрыты (см. план 72 §«не
//     делаем» и план 70 отложен).
//   - **Реклама/баннеры** — не переносятся.
//   - **Реферальный экран** — отдельного пункта нет, ведёт в portal
//     (план 59) — будет реализовано на этапе Spring 4.

import type { ReactNode } from 'react';
import { Link } from 'react-router-dom';

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

interface RouterLinkProps {
  to: string;
  label: string;
}

function RouterLink({ to, label }: RouterLinkProps) {
  return (
    <li>
      <Link to={to}>{label}</Link>
    </li>
  );
}

interface AnchorLinkProps {
  href: string;
  label: string;
}

function AnchorLink({ href, label }: AnchorLinkProps) {
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
        <RouterLink to="/" label="Обзор" />
        <RouterLink to="/empire" label="Империя" />
        <RouterLink to="/resource-market" label="Ресурсы" />

        <MenuGroup className="menu-prod" label="Производство" />
        <RouterLink to="/constructions" label="Строения" />
        <RouterLink to="/research" label="Исследования" />
        <RouterLink to="/shipyard" label="Верфь" />
        <RouterLink to="/repair" label="Ремонт" />

        <MenuGroup className="menu-user" label="Игрок" />
        <RouterLink to="/galaxy" label="Галактика" />
        <RouterLink to="/mission" label="Миссии" />
        <RouterLink to="/fleet-operations" label="Флот" />
        <RouterLink to="/alliance" label="Альянс" />
        <RouterLink to="/market" label="Артефакты" />
        <RouterLink to="/battlestats" label="История боёв" />

        <MenuGroup className="menu-other" label="Прочее" />
        <AnchorLink href="#chat" label="Чат" />
        <AnchorLink href="#msg" label="Сообщения" />
        <AnchorLink href="#friends" label="Друзья" />
        <AnchorLink href="#statistics" label="Статистика" />
        <AnchorLink href="#settings" label="Настройки" />
      </ul>
    </div>
  );
}
