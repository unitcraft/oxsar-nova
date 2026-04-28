# Nova UI Backlog (U-NNN + X-NNN)

**Дата сборки**: 2026-04-28
**Контекст**: артефакт плана 62. Самостоятельный долгоживущий
бэклог. Каждая запись — функция/UX-сигнал, доступный игроку в
origin, но **отсутствующий или урезанный в game-nova-frontend**.
Не зависит от вердикта по origin: даже если origin останется на
PHP, каждая запись — кандидат для эволюции nova-frontend.

Связь:
- **U-NNN** — UI-функции (целые экраны или крупные элементы)
- **X-NNN** — UX-микрологика (цвета, disabled, tooltip'ы)
- Если U-NNN/X-NNN требует нового endpoint'а или поля DTO — есть
  ссылка на D-NNN в `divergence-log.md`.

---

## Часть I. UI-функции (U-NNN)

### U-001. Биржа артефактов (Stock/StockNew/Exchange)

- **Где в origin**: `src/game/page/Stock.class.php` (757 стр),
  `StockNew.class.php` (850 стр), `Exchange.class.php` (1220 стр),
  шаблоны `stock.tpl`, `stock_new_*.tpl`, `lot_details.tpl`,
  `exchange.tpl`, `artefactmarket.tpl`, `artefactmarket2.tpl`
- **Что делает игрок**: создаёт лоты артефактов на продажу за
  ресурсы/кредиты, покупает чужие лоты, устанавливает TTL (3-7
  дней), история торговли, premium-лоты
- **Backend**: ОТСУТСТВУЕТ (`D-EXCHANGE`)
- **Аналог в nova**: `artmarket` есть как урезанный (продажа от
  системы), но **player-to-player exchange** — нет
- **Кандидат на добавление**: ⭐ да, ключевая legacy-фича
- **Связь с origin**: обязательно для legacy01
- **Объём**: backend (новый модуль `internal/exchange/` ~2000 стр)
  + frontend (3 экрана — список, детали лота, создание); 2-3
  недели

### U-002. Турниры

- **Где в origin**: `src/templates/standard/tournament.tpl` (если
  есть); EVENT_TOURNAMENT_SCHEDULE / RESCHEDULE / PARTICIPANT
  (зарезервированы в EventHandler, не реализованы)
- **Что делает игрок**: участвует в турнирах с расписанием
- **Backend**: ОТСУТСТВУЕТ полностью в nova и origin
- **Аналог в nova**: нет
- **Кандидат**: ⭐ да, для всех вселенных
- **Связь с origin**: только если решим включить (опционально)
- **Объём**: новый модуль с нуля ~3-4 недели

### U-003. Расширенная дипломатия альянсов

- **Где в origin**: `Alliance.class.php` actions
  `applyRelationship`, `acceptRelation`, `refuseRelation`,
  `diplomacy`, `relApplications`, `determineRelation`,
  `getDiploStatus`; шаблоны `ally_diplomacy.tpl`,
  `relation_applications.tpl`
- **Что делает игрок**: устанавливает NAP/союз/война с другими
  альянсами через двухсторонний процесс заявок
- **Backend (nova)**: `GET /api/alliances/{id}/relations`,
  `PUT /api/alliances/{id}/relations/{target_id}`,
  `POST /accept`, `DELETE` — есть, но статусы и UI урезаны
- **Аналог в nova**: 4 endpoint'а есть, но нет UI на frontend (нет
  feature `alliance/diplomacy`)
- **Кандидат**: да
- **Объём**: только frontend ~3-5 дней

### U-004. Передача лидерства альянса (abandonAlly)

- **Где в origin**: `Alliance.class.php::abandonAlly` +
  `referFounderStatus`
- **Что делает игрок**: текущий owner может передать лидерство
  другому члену
- **Backend (nova)**: ОТСУТСТВУЕТ (см. D-NNN-ALLIANCE-OWNERSHIP)
- **Кандидат**: да
- **Объём**: backend endpoint
  `POST /api/alliances/{id}/transfer-ownership/{userID}` + UI
  кнопка; 1 день

### U-005. Гранулярные права рангов альянса

- **Где в origin**: `Alliance.class.php::manageRanks`,
  `getRankSelect`, `getRights`; шаблон `manage_ranks.tpl`
- **Что делает игрок**: создаёт кастомные ранги с битовыми правами
  (canManageApps, canSendGlobalMail, canManageDiplomacy, canKick,
  canManageRanks, canEditDescription)
- **Backend (nova)**: rank = enum (`'owner'|'member'`) — без прав
- **Аналог в nova**: только rank_name (произвольная строка из
  миграции 0034) — но это просто имя, без прав
- **Кандидат**: да (для legacy01 особенно)
- **Объём**: новая таблица `alliance_ranks (id, alliance_id, name,
  permissions JSONB)` + миграция + backend handler + frontend
  ~1-2 недели

### U-006. Global mail членам альянса

- **Где в origin**: `Alliance.class.php::globalMail`; шаблон
  `globalmail.tpl`
- **Что делает**: лидер шлёт письмо всем членам с TipTap-редактором
- **Backend**: блокирован планом 57 (mail-service). После 57 —
  endpoint
- **Кандидат**: да, ПОСЛЕ плана 57
- **Объём**: после 57 — frontend + endpoint ~3-5 дней

### U-007. Симулятор боя как UI-фича

- **Где в origin**: `Simulator.class.php` (749 стр), шаблон
  `simulator.tpl`
- **Что делает**: пользовательский UI для расчёта прогноза боя
- **Backend (nova)**: `POST /api/battle-sim` есть
- **Аналог в nova**: `frontend/src/features/battle-sim/` есть
- **Статус**: похоже **уже есть** в nova-frontend, но проверить
  паритет с origin (разные UI могут отличаться)
- **Кандидат**: возможно — улучшение существующего

### U-008. Buddy-list (друзья)

- **Где в origin**: `Friends.class.php`, шаблон `buddylist.tpl`,
  таблица `na_buddylist`
- **Что делает**: список друзей, заявки в друзья
- **Backend (nova)**: `GET/POST/DELETE /api/friends{,/{id}}` есть
- **Аналог в nova**: `frontend/src/features/friends/` (1 файл)
- **Статус**: вероятно есть в nova, верифицировать UI

### U-009. Telepor планеты (TELEPORT_PLANET)

- **Где в origin**: EVENT_TELEPORT_PLANET, поле
  `na_user.planet_teleport_time` (rate-limit), Mission action
  `canTeleportCurPlanet`
- **Что делает**: артефакт перемещает планету на новые координаты
- **Backend (nova)**: ОТСУТСТВУЕТ (см. D-NNN-TELEPORT)
- **Кандидат**: да (особенность legacy)
- **Объём**: новый Kind в event/kinds.go + handler + UI ~5-7 дней

### U-010. Историй атак на планету (Battlestats детально)

- **Где в origin**: `Battlestats.class.php`, шаблон `battlestats.tpl`
  + `assault_report.tpl`
- **Что делает**: длинный лог боёв с фильтрами
- **Backend (nova)**: `GET /api/battlestats` есть
- **Аналог**: есть `frontend/src/features/battlestats/`
- **Статус**: верифицировать паритет фильтров

### U-011. Custom logo альянса

- **Где в origin**: вероятно через `na_alliance` поле logo (uploaded)
- **Что делает**: лидер загружает logo
- **Backend (nova)**: ОТСУТСТВУЕТ
- **Кандидат**: да (но требует storage и moderation для UGC)
- **Объём**: storage + S3-uploader + moderation в админке
  ~1-2 недели

### U-012. Полнотекстовый поиск альянсов

- **Где в origin**: `Alliance.class.php::allySearch`,
  шаблон `allysearch.tpl`, `ally_search_result.tpl`
- **Что делает**: поиск с фильтрами (открытый/закрытый, размер)
- **Backend (nova)**: `GET /api/search` (общий поиск) — есть
- **Аналог**: есть `features/search/`, но фильтры по альянсам
  нужны
- **Кандидат**: да, расширить общий поиск
- **Объём**: 2-3 дня

### U-013. Альянсный лог активности

- **Где в origin**: вероятно через сообщения/события (не явно)
- **Что делает**: кто вступил/вышел/изгнан/повышен
- **Backend (nova)**: ОТСУТСТВУЕТ
- **Кандидат**: да
- **Объём**: новая таблица `alliance_audit_log` + endpoint + UI
  ~1 неделя

### U-014. Tournament UI (если включаем U-002)

- зависит от U-002

### U-015. Multiple alliance descriptions (3 описания)

- **Где в origin**: `Alliance.class.php::updateAllyPrefs`
- **Что делает**: лидер устанавливает 3 разных описания —
  external (для всех), internal (для членов), application (для
  заявок)
- **Backend (nova)**: `description` одно поле
- **Кандидат**: да
- **Объём**: миграция + backend handler + frontend; 2-3 дня

---

## Часть II. UX-микрологика (X-NNN)

### X-001. Индикация дефицита ресурсов при постройке

- **Где**: `templates/standard/required_res_info.tpl:5-40`,
  `required_res_table.tpl:6-41`
  ```smarty
  <td class='{if[$row["metal_notavailable"]]}notavailable{else}true{/if}'>
    {loop=metal_required}</td>
  <td>{if[$row["metal_notavailable"]]}({loop=metal_notavailable}){/if}</td>
  ```
- **Триггер**: `metal_notavailable` (недостаточно ресурса)
- **Визуальное**: `.notavailable` (#fd7171) или `.true` (#00ff00),
  скобки показывают дефицит `(нужно X)`
- **Аналог nova**: частично (в ресурс-панели)
- **Кандидат**: да, **полная поддержка с дефицитом в скобках**
- **Связь**: Constructions, Research, Shipyard, Repair, ArtefactMarket

### X-002. Производство положительное/отрицательное

- **Где**: `templates/standard/resource.tpl:35-65`
  ```smarty
  {if[$row["metal"] > 0]}<span class="true">{loop}metal{/loop}</span>
  {else if[$row["metalCons"] > 0]}<span class="false">{loop}metalCons{/loop}</span>
  {else}0{/if}
  ```
- **Триггер**: `metal > 0` (производство) vs `metalCons > 0` (потребление)
- **Визуальное**: `.true` зелёный, `.false` красный (`#fc3232`)
- **Аналог nova**: нет явной индикации потребления
- **Кандидат**: да (важно для энергодефицита)

### X-003. Условная видимость кнопок постройки

- **Где**: `research.tpl:88,96`, `constructions.tpl:94,102`,
  `shipyard.tpl:72`
  ```smarty
  {if[$row["can_build"]]}{include}"required_res_table"{/include}
  {else}<span class="normal">{lang}REQUIRED_CONSTRUCTIONS{/lang}</span>
  <br />{loop=required_constructions}{/if}
  ```
- **Триггер**: `can_build = false` — нет требований
- **Визуальное**: показ требований вместо кнопки
- **Аналог nova**: ❌ отсутствует — нет явного отображения причины
- **Кандидат**: ⭐ да, критично для UX

### X-004. Прогресс-бар заряда артефактов

- **Где**: `artefacts.tpl:110-119`
  ```smarty
  {else if[$row['delay_counter']]}<span class='false'>заряжается</span>
  {else if[$row['disappear_counter']]}
    {if[$row['active']]}<span class='true'>активирован</span>
  ```
- **Триггер**: `delay_counter > 0`, `active = true`
- **Визуальное**: `.false`/`.true` + полоса заряда
- **Аналог nova**: ❌ отсутствует
- **Кандидат**: да

### X-005. Хранилище переполнено (free ≤ 0)

- **Где**: `artefacts.tpl:26-27`, `disassemble.tpl:188-189`
  ```smarty
  <td{if[{var=artefact_tech_free} <= 0]} class='false'{/if}>...</td>
  ```
- **Триггер**: `free <= 0`
- **Визуальное**: `.false` + прогресс-бар (rep_destroyed/rep_alive)
- **Аналог nova**: ❌
- **Кандидат**: да

### X-006. Блокировка ввода кораблей в миссии (blocked)

- **Где**: `missions.tpl:108,221`
- **Триггер**: `blocked = true` (скорость 0, в брони и т.п.)
- **Визуальное**: input скрывается
- **Аналог nova**: ❌
- **Кандидат**: да, с tooltip причины

### X-007. Слоты флота/экспедиции (no free slot)

- **Где**: `missions3.tpl:118-125`
  ```smarty
  {if[!{var}can_send_fleet{/var}]}{lang}NO_FREE_FLEET_SLOTS{/lang}{/if}
  {if[!{var}can_send_expo{/var} && NS::getResearch(UNIT_EXPO_TECH) > 0]}
    {lang}NO_FREE_EXPO_SLOTS{/lang}{/if}
  ```
- **Триггер**: `can_send_fleet = false`
- **Визуальное**: текст причины
- **Аналог nova**: ❌
- **Кандидат**: да, с подсчётом занятых слотов

### X-008. Состояние артефакта (активирован/заряжается/существует)

- **Где**: `artefacts.tpl:108-115`
  ```smarty
  {if[$row['isInLot']]}существует
  {else if[$row['expire_counter']]}<span class='true'>активирован</span>
  {else if[$row['delay_counter']]}<span class='false'>заряжается</span>
  ```
- **Триггер**: счётчики `expire_counter`, `delay_counter`,
  `disappear_counter`, флаг `isInLot`
- **Аналог nova**: ❌
- **Кандидат**: да

### X-009. Tooltip с требуемым уровнем (helptip)

- **Где**: `resource.tpl:35`
  ```smarty
  <td><b{if[$row['helptip']]} class="helptip"
       onmouseover="Tip('{loop=helptip}', FADEIN, 500);"
       onmouseout="UnTip();"{/if}>{loop}name{/loop}</b></td>
  ```
- **Триггер**: `helptip` непуста
- **Визуальное**: JS tooltip (wz_tooltip.js)
- **Аналог nova**: tooltip есть, но логика отличается
- **Кандидат**: да, расширить покрытие подсказок

### X-010. Энергодефицит (totalEnergy <= 0)

- **Где**: `resource.tpl:54-58`
  ```smarty
  <span class="{if[{var}totalEnergy{/var} <= 0]}false{else}true{/if}">
    {@totalEnergy}</span>
  ```
- **Триггер**: `totalEnergy <= 0`
- **Аналог nova**: ❌
- **Кандидат**: ⭐ да (критично для геймплея)

### X-011. Полоса здоровья сооружения (rep_destroyed/rep_alive)

- **Где**: `disassemble.tpl:13`, `repair.tpl:13`, `artefacts.tpl:14`
  ```smarty
  <div class='rep_destroyed_back_div' style="clear:both">
    <div class='rep_alive_over_div' style='width: {@construction_free_percent}%' />
  </div>
  ```
- **Триггер**: `construction_free_percent`
- **Визуальное**: красная полоса (повреждено) + голубая (здорово)
- **Аналог nova**: ❌
- **Кандидат**: да

### X-012. Форма "требуемые конструкции" vs "таблица ресурсов"

- **Где**: `research.tpl:88-93`, `constructions.tpl:94-99`
- **Триггер**: `can_build = false`
- **Визуальное**: условная смена контента
- **Кандидат**: да (часть X-003)

### X-013. Состояние уровня (added_level +/-)

- **Где**: `research.tpl:68-69`, `constructions.tpl:74-75`
  ```smarty
  {if[$row['added_level']>0]} <span class='true'>(+{loop=added_level})</span>
  {else if[$row['added_level']<0]} <span class='false'>({loop=added_level})</span>{/if}
  ```
- **Триггер**: `added_level > 0` (бонус) / `< 0` (штраф)
- **Визуальное**: `(+2)` зелёное / `(-1)` красное
- **Аналог nova**: ❌
- **Кандидат**: да

### X-014. Недостаточные ремонтные поля

- **Где**: `disassemble.tpl:143-145`
- **Триггер**: `no_free_repair_fields > 0`
- **Визуальное**: `.notavailable`
- **Кандидат**: да

### X-015. Active/inactive артефакты (5 (2/3))

- **Где**: `artefacts.tpl:82`
  ```smarty
  {loop}quantity{/loop}{if[$row['active_count'] || $row['inactive_count']]}
    (<span class="active">{loop}active_count{/loop}</span>
    /<span class="inactive">{loop}inactive_count{/loop}</span>){/if}
  ```
- **Кандидат**: да

### X-016. Максимум активированных артефактов (false2 окраска)

- **Где**: `artefact_row_info.tpl:33-34`
- **Триггер**: `active_count >= max_active`
- **Кандидат**: да

### X-017. Скидки на бирже (trade-union класс)

- **Где**: `stock.tpl:206,208`
- **Триггер**: `disc_price` непуста
- **Кандидат**: да (для биржи)

### X-018. Ошибка авторизации (error div)

- **Где**: `login.tpl:45`, `login_sub.tpl:33`
- **Триггер**: `errorMsg` непуста
- **Аналог nova**: есть
- **Кандидат**: улучшить стилизацию

### X-019. Real-time валидация формы (field_warning)

- **Где**: `signup_sub.tpl:203-206`
- **Триггер**: JS валидация
- **Аналог nova**: ❌ нет явной realtime-валидации
- **Кандидат**: да

### X-020. Знак торговца (false2 для обычных)

- **Где**: `stock.tpl:138,144`
- **Триггер**: нет premium-marker (Знак торговца)
- **Кандидат**: да (для биржи)

### X-021. Новые достижения (зелёный бейдж)

- **Где**: `achievements.tpl:113`
  ```smarty
  {if[{var=real_new_achieve_count} > 0]}
    <span class="true"><b>+{@real_new_achieve_count}</b></span>{/if}
  ```
- **Триггер**: `real_new_achieve_count > 0`
- **Визуальное**: `.true` зелёное жирное
- **Аналог nova**: ❌ счётчика новых достижений нет
- **Кандидат**: да

### X-022. ui-state-error панель (jQuery UI)

- **Где**: `achievements.tpl:69`, `paymentA1step2.tpl:38,87`
- **Аналог nova**: jQuery UI не используется — заменить на свою
  систему
- **Кандидат**: зависит (своя дизайн-система)

---

## Цветовая палитра origin (для воспроизведения)

### Красный (ошибка/дефицит/отрицательное)

- `.false` — `color: #fc3232 !important;` (основной)
- `.notavailable`, `.false2` — `color: #fd7171` (светлее)
- `.rep_destroyed_back_div` — `background: #fc3232;` (полоса повреждения)
- Odd-строки `.false` — `color: #ff5555` (для чередования)

### Зелёный (успех/положительное)

- `.available`, `.true` — `color: #00ff00 !important;`
- `.rep_alive_over_div` — `background: #6cd8bc;` (живой)

### Оранжевый/жёлтый (предупреждение)

- `.rep_quantity_damage_low` — `color: #d57c08 !important;`

### Специальные

- `.trade-union` — премиум скидки (вероятно жёлтый/оранжевый)
- `.active`/`.inactive` — артефакты (разные цвета)

---

## Сводка

- **U-NNN** (UI-функции): 15 записей. Ключевые: U-001 (биржа),
  U-005 (гранулярные права), U-009 (telepor планеты)
- **X-NNN** (UX-микрологика): 22 записи. Ключевые: X-003 (cause-buttons),
  X-010 (энергодефицит), X-001 (дефицит ресурсов)

Минимальные пороги плана 62 (≥10 U-NNN + ≥20 X-NNN) — выполнены.

## Связь с D-NNN

- U-001 (биржа) ↔ `D-EXCHANGE` в divergence-log
- U-004 (передача лидерства) ↔ `D-ALLIANCE-OWNERSHIP`
- U-005 (гранулярные ранги) ↔ `D-014` Alliance ranks
- U-009 (telepor) ↔ `D-TELEPORT`
- U-015 (3 описания) ↔ `D-ALLIANCE-DESC`

---

## References

- 125 шаблонов: `projects/game-origin/src/templates/standard/`
- 55 контроллеров: `projects/game-origin/src/game/page/`
- nova frontend: `projects/game-nova/frontend/src/features/`
- [origin-ui-replication.md](origin-ui-replication.md) —
  S-NNN экраны (откуда берутся U-NNN)
- [divergence-log.md](divergence-log.md) — D-NNN расхождения
