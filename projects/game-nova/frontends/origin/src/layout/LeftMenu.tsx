// Левое меню origin-фронта (план 72 Ф.1-Ф.5).
//
// Pixel-perfect клон legacy layout.tpl + Menu.class.php:
//   - Секция 1 (Планета): пустой li-разделитель, плоские ссылки
//   - Секции 2-8: li.menu-info.menuN-max + toggle (show/hide по клику)
//
// Намеренные расхождения с legacy:
//   - Achievements, Tutorial — скрыты (план 72 §«не делаем», план 70 отложен)
//   - Реклама/баннеры — не переносятся
//   - JS menuSoH заменён на React useState toggle (ADR: без jQuery)
//   - /resource → /resource-market (роутер), /defense → заглушка,
//     /disassemble → заглушка, /stock → заглушка

import { useState } from 'react';
import { Link } from 'react-router-dom';

function MenuSection({
  id,
  colorClass,
  title,
  children,
}: {
  id: string;
  colorClass: string;
  title: string;
  children?: React.ReactNode;
}) {
  const [open, setOpen] = useState(false);
  return (
    <>
      <li
        className={`menu-info ${colorClass}-${open ? 'max' : 'min'}`}
        id={`menuli${id}`}
        title={title}
        onClick={() => setOpen((v) => !v)}
        style={{ cursor: 'pointer' }}
      />
      <li id={`menudiv${id}`} style={{ display: open ? '' : 'none' }}>
        <ul>{children}</ul>
      </li>
    </>
  );
}

export function LeftMenu() {
  return (
    <div id="leftMenu">
      <ul>
        {/* Секция 1: Планета — без заголовка, плоский список */}
        <li></li>
        <li><Link to="/">Обзор</Link></li>
        <li><Link to="/resource-market">Сырьё</Link></li>
        <li><Link to="/constructions" id="menu_build">Постройки</Link></li>
        <li><Link to="/research" id="menu_research">Исследования</Link></li>
        <li><Link to="/shipyard">Верфь</Link></li>
        <li><Link to="/defense">Оборона</Link></li>
        <li><Link to="/mission" id="menu_fleet">Флот</Link></li>
        <li><Link to="/artefacts">Артефакты</Link></li>
        <li><Link to="/repair">Ремонт</Link></li>
        <li><Link to="/disassemble">Утилизация</Link></li>
        <li><Link to="/stock">Биржа</Link></li>

        {/* Секция 2: Коммуникации */}
        <MenuSection id="2" colorClass="menu2" title="Коммуникации">
          <li><Link to="/chat" id="menu_chat">Чат</Link></li>
          <li><Link to="/msg">Сообщения</Link></li>
          <li><Link to="/notepad">Блокнот</Link></li>
          <li><Link to="/alliance">Альянс</Link></li>
          <li><Link to="/friends">Друзья</Link></li>
          <li><Link to="/search">Поиск</Link></li>
        </MenuSection>

        {/* Секция 3: Галактика */}
        <MenuSection id="3" colorClass="menu3" title="Галактика">
          <li><Link to="/galaxy" id="menu_galaxy">Галактика</Link></li>
          <li><Link to="/empire">Империя</Link></li>
          <li><Link to="/techtree">Технологии</Link></li>
          <li><Link to="/profession" id="menu_profession">Профессия</Link></li>
        </MenuSection>

        {/* Секция 4: Статистика */}
        <MenuSection id="4" colorClass="menu4" title="Статистика">
          <li><Link to="/ranking">Статистика</Link></li>
          <li><Link to="/records">Рекорды</Link></li>
          <li><Link to="/battlestats">Сражения</Link></li>
          <li><Link to="/tools/tech-calc" id="menu_simulator">Симулятор боя</Link></li>
        </MenuSection>

        {/* Секция 5: Рынок ресурсов */}
        <MenuSection id="5" colorClass="menu5" title="Рынок ресурсов">
          <li><Link to="/resource-market">Обменник</Link></li>
          <li><Link to="/market">Магазин артефактов</Link></li>
          <li><Link to="/payment">Пополнить кредиты</Link></li>
        </MenuSection>

        {/* Секция 6: (форумы — внешние, не переносим) */}
        <MenuSection id="6" colorClass="menu6" title="Поиск">
        </MenuSection>

        {/* Секция 7: Настройки */}
        <MenuSection id="7" colorClass="menu7" title="Настройки">
          <li><Link to="/settings">Планета</Link></li>
          <li><Link to="/settings">Настройки</Link></li>
          <li><Link to="/support">Регламент тех. поддержки</Link></li>
        </MenuSection>

        {/* Секция 8: Выход */}
        <MenuSection id="8" colorClass="menu8" title="Выход">
          <li><Link to="/logout">Выход</Link></li>
        </MenuSection>

        <li><span style={{ whiteSpace: 'nowrap' }}>Oxsar Nova</span></li>
      </ul>
    </div>
  );
}
