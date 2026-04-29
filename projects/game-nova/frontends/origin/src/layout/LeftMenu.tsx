// Левое меню origin-фронта (план 72 Ф.1 + Ф.2 Spring 1 + Ф.3 Spring 2 + Ф.5 Spring 4).
//
// Pixel-perfect клон legacy layout.tpl + Menu.class.php:
//   - Секция 1 (Планета): пустой li-разделитель, плоские ссылки
//   - Секции 2-8: li.menu-info.menuN-max + li#menudivN > ul > ссылки
//
// Намеренные расхождения с legacy:
//   - Achievements, Tutorial — скрыты (план 72 §«не делаем», план 70 отложен)
//   - Реклама/баннеры — не переносятся
//   - Реферальный экран — заглушка (план 59 portal)
//   - Widgets, UserAgreement, Support — заглушки (Spring 4)
//   - JS-разворачивание секций (menuSoH) — не реализуем (ADR: статическое меню)

import { Link } from 'react-router-dom';

export function LeftMenu() {
  return (
    <div id="leftMenu">
      <ul>
        {/* Секция 1: Планета — без заголовка, плоский список */}
        <li></li>
        <li><Link to="/">Обзор</Link></li>
        <li><Link to="/resource">Сырьё</Link></li>
        <li><Link to="/constructions" id="menu_build">Постройки</Link></li>
        <li><Link to="/research" id="menu_research">Исследования</Link></li>
        <li><Link to="/shipyard">Верфь</Link></li>
        <li><Link to="/defense">Оборона</Link></li>
        <li><Link to="/mission" id="menu_fleet">Флот</Link></li>
        <li><Link to="/artefacts">Артефакты</Link></li>
        <li><Link to="/repair">Ремонт</Link></li>
        <li><Link to="/disassemble">Утилизация</Link></li>
        <li><Link to="/stock">Биржа</Link></li>

        {/* Секции 2-8: заголовки — цветные полоски (свёрнуты по умолчанию через CSS) */}
        <li className="menu-info menu2-max" id="menuli2" title="Коммуникации"></li>
        <li id="menudiv2"><ul>
          <li><Link to="/chat" id="menu_chat">Чат</Link></li>
          <li><Link to="/messages">Сообщения</Link></li>
          <li><Link to="/notepad">Блокнот</Link></li>
          <li><Link to="/alliance">Альянс</Link></li>
          <li><Link to="/friends">Друзья</Link></li>
          <li><Link to="/search">Поиск</Link></li>
        </ul></li>

        <li className="menu-info menu3-max" id="menuli3" title="Галактика"></li>
        <li id="menudiv3"><ul>
          <li><Link to="/galaxy" id="menu_galaxy">Галактика</Link></li>
          <li><Link to="/empire">Империя</Link></li>
          <li><Link to="/techtree">Технологии</Link></li>
          <li><Link to="/profession" id="menu_profession">Профессия</Link></li>
        </ul></li>

        <li className="menu-info menu4-max" id="menuli4" title="Статистика"></li>
        <li id="menudiv4"><ul>
          <li><Link to="/ranking">Статистика</Link></li>
          <li><Link to="/records">Рекорды</Link></li>
          <li><Link to="/battlestats">Сражения</Link></li>
          <li><Link to="/simulator" id="menu_simulator">Симулятор боя</Link></li>
        </ul></li>

        <li className="menu-info menu5-max" id="menuli5" title="Рынок ресурсов"></li>
        <li id="menudiv5"><ul>
          <li><Link to="/market">Обменник</Link></li>
          <li><Link to="/artefact-market">Магазин артефактов</Link></li>
          <li><Link to="/payment">Пополнить кредиты</Link></li>
        </ul></li>

        <li className="menu-info menu6-max" id="menuli6" title="Поиск"></li>
        <li id="menudiv6"><ul>
        </ul></li>

        <li className="menu-info menu7-max" id="menuli7" title="Настройки"></li>
        <li id="menudiv7"><ul>
          <li><Link to="/planet-options">Планета</Link></li>
          <li><Link to="/prefs">Настройки</Link></li>
          <li><Link to="/support">Регламент тех. поддержки</Link></li>
        </ul></li>

        <li className="menu-info menu8-max" id="menuli8" title="Выход"></li>
        <li id="menudiv8"><ul>
          <li><Link to="/logout">Выход</Link></li>
        </ul></li>

        <li><span style={{ whiteSpace: 'nowrap' }}>Oxsar Nova</span></li>
      </ul>
    </div>
  );
}
