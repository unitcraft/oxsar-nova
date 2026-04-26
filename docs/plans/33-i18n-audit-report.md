# i18n Audit Report

Всего: **1383** хардкод-строк с кириллицей.
Уже в ru.yml: **358**. Нужно добавить: **1025**.

| Файл | Строка | Литерал | Предложенный ключ | В ru.yml? |
|---|---|---|---|---|
| backend/internal/achievement/service.go | 64 | Достижение: %s | `achievement.достижение_s` | ❌ |
| backend/internal/achievement/service.go | 65 | Открыто новое достижение: %s. Начислено %d кредитов. | `achievement.открыто_новое_достижение` | ❌ |
| backend/internal/aiadvisor/service.go | 193 | Планеты игрока:\n | `global.планеты_игрокаn` | ❌ |
| backend/internal/alien/alien.go | 226 | Инопланетяне | `alien.инопланетяне` | ❌ |
| backend/internal/alien/alien.go | 310 | атаковали и победили | `alien.атаковали_и_победили` | ❌ |
| backend/internal/alien/alien.go | 312 | атаковали, но были отбиты | `alien.атаковали_но_были` | ❌ |
| backend/internal/alien/alien.go | 314 | атаковали — ничья | `alien.атаковали_ничья` | ❌ |
| backend/internal/alien/alien.go | 316 | Инопланетяне (тир %d) %s вашу планету. | `alien.инопланетяне_тир_d` | ❌ |
| backend/internal/alien/alien.go | 318 | Похищено %d кредитов. | `alien.похищено_d_кредитов` | ❌ |
| backend/internal/alien/alien.go | 321 | Инопланетяне оставили %d кредитов в знак уважения. | `alien.инопланетяне_оставили_d` | ❌ |
| backend/internal/alien/alien.go | 324 | Их флот остался на орбите — ждите отдельного сообщения об… | `alien.их_флот_остался` | ❌ |
| backend/internal/alien/holding.go | 59 | Инопланетяне (удержание) | `alien.инопланетяне_удержание` | ❌ |
| backend/internal/alien/holding.go | 276 | Инопланетяне (тир %d) установили контроль над вашей плане… | `alien.инопланетяне_тир_d` | ❌ |
| backend/internal/alien/holding.go | 277 | Их флот останется на орбите до %s. Пока они здесь, они от… | `alien.их_флот_останется` | ❌ |
| backend/internal/alien/holding.go | 278 | атаки других игроков — но и сами забирают часть ресурсов. | `alien.атаки_других_игроков` | ❌ |
| backend/internal/alien/holding.go | 315 | Инопланетяне покинули вашу планету — флот ушёл в глубокий… | `alien.инопланетяне_покинули_вашу` | ❌ |
| backend/internal/alien/holding.go | 318 | Инопланетяне покинули вашу планету. За время удержания вы | `alien.инопланетяне_покинули_вашу` | ❌ |
| backend/internal/alien/holding.go | 319 | заплатили им %d кредитов за %d продлений. | `alien.заплатили_им_d` | ❌ |
| backend/internal/alien/holding.go | 431 | Инопланетяне выгрузили на вашу планету ресурсы: | `alien.инопланетяне_выгрузили_на` | ❌ |
| backend/internal/alien/holding.go | 432 | металл +%d, кремний +%d, водород +%d. | `alien.металл_d_кремний` | ❌ |
| backend/internal/alien/holding.go | 511 | Инопланетяне постепенно покинули вашу планету — их флот р… | `alien.инопланетяне_постепенно_покинули` | ❌ |
| backend/internal/alliance/service.go | 308 | Вы вступили в альянс | `alliance.вы_вступили_в` | ❌ |
| backend/internal/alliance/service.go | 309 | Вы стали членом альянса [%s]. | `alliance.вы_стали_членом` | ❌ |
| backend/internal/alliance/service.go | 312 | Заявка на вступление в альянс | `alliance.заявка_на_вступление` | ❌ |
| backend/internal/alliance/service.go | 313 | Игрок %s подал заявку на вступление в ваш альянс [%s]. | `alliance.игрок_s_подал` | ❌ |
| backend/internal/alliance/service.go | 423 | Заявка одобрена | `alliance.заявка_одобрена` | ❌ |
| backend/internal/alliance/service.go | 424 | Ваша заявка на вступление в альянс [%s] одобрена. | `alliance.ваша_заявка_на` | ❌ |
| backend/internal/alliance/service.go | 457 | Заявка отклонена | `alliance.заявка_отклонена` | ❌ |
| backend/internal/alliance/service.go | 458 | Ваша заявка на вступление в альянс [%s] отклонена. | `alliance.ваша_заявка_на` | ❌ |
| backend/internal/artefact/expire.go | 65 | Артефакт истёк | `artefact.артефакт_истёк` | ❌ |
| backend/internal/artefact/expire.go | 66 | Действие артефакта «%s» завершилось. Эффекты отменены. | `artefact.действие_артефакта_s` | ❌ |
| backend/internal/fleet/acs_attack.go | 373 | ACS боевой отчёт: %s | `fleet.acs_боевой_отчёт` | ❌ |
| backend/internal/fleet/acs_attack.go | 374 | Раундов: %d. ACS-группа %s. | `fleet.раундов_d_acsгруппа` | ❌ |
| backend/internal/fleet/attack.go | 790 | Боевой отчёт: %s | `fleet.боевой_отчёт_s` | ❌ |
| backend/internal/fleet/attack.go | 791 | Раундов: %d. Добыча: %d M / %d Si / %d H. | `fleet.раундов_d_добыча` | ❌ |
| backend/internal/fleet/attack.go | 857 | Луна создана в %d:%d:%d | `fleet.луна_создана_в` | ❌ |
| backend/internal/fleet/attack.go | 858 | В результате боя образовалась луна на %d:%d:%d (диаметр %d). | `fleet.в_результате_боя` | ❌ |
| backend/internal/fleet/colonize.go | 77 | Колонизация луны невозможна | `fleet.колонизация_луны_невозможна` | ❌ |
| backend/internal/fleet/colonize.go | 89 | Нет колониального корабля во флоте | `fleet.нет_колониального_корабля` | ❌ |
| backend/internal/fleet/colonize.go | 101 | Координата занята | `fleet.координата_занята` | ❌ |
| backend/internal/fleet/colonize.go | 128 | Достигнут лимит планет (%d/%d). Улучшите computer_tech. | `fleet.достигнут_лимит_планет` | ❌ |
| backend/internal/fleet/colonize.go | 186 | Основана колония %d:%d:%d | `fleet.основана_колония_ddd` | ❌ |
| backend/internal/fleet/colonize.go | 187 | Новая планета «%s» на координатах [%d:%d:%d]. | `fleet.новая_планета_s` | ❌ |
| backend/internal/fleet/colonize.go | 188 | Стартовые ресурсы: %d M / %d Si / %d H. | `fleet.стартовые_ресурсы_d` | ❌ |
| backend/internal/fleet/colonize.go | 207 | Колонизация %d:%d:%d провалена | `fleet.колонизация_ddd_провалена` | ❌ |
| backend/internal/fleet/events.go | 318 | Перебазирование завершено | `fleet.перебазирование_завершено` | ❌ |
| backend/internal/fleet/events.go | 319 | Флот прибыл на [%d:%d:%d]. Корабли перемещены на планету,… | `fleet.флот_прибыл_на` | ❌ |
| backend/internal/fleet/expedition.go | 201 | Ничего не нашли. | `fleet.ничего_не_нашли` | ❌ |
| backend/internal/fleet/expedition.go | 213 | Экспедиция: %s | `fleet.экспедиция_s` | ❌ |
| backend/internal/fleet/expedition.go | 214 | Результат: %s | `fleet.результат_s` | ❌ |
| backend/internal/fleet/expedition.go | 385 | нет места в cargo | `fleet.нет_места_в` | ❌ |
| backend/internal/fleet/expedition.go | 431 | нет места в cargo | `fleet.нет_места_в` | ❌ |
| backend/internal/fleet/expedition.go | 467 | артефакты недоступны | `fleet.артефакты_недоступны` | ❌ |
| backend/internal/fleet/expedition.go | 501 | Обнаружена пригодная планета, но достигнут лимит (%d/%d).… | `fleet.обнаружена_пригодная_планета` | ❌ |
| backend/internal/fleet/expedition.go | 557 | Пригодная планета найдена, но подходящая позиция не обнар… | `fleet.пригодная_планета_найдена` | ❌ |
| backend/internal/fleet/expedition.go | 565 | нет ship'ов | `fleet.нет_shipов` | ❌ |
| backend/internal/fleet/expedition.go | 607 | нет ship'ов | `fleet.нет_shipов` | ❌ |
| backend/internal/fleet/expedition.go | 644 | %d × light_fighter (повреждённые) | `fleet.d_lightfighter_повреждённые` | ❌ |
| backend/internal/fleet/expedition.go | 698 | delay: нет return_event_id | `fleet.delay_нет_returneventid` | ❌ |
| backend/internal/fleet/expedition.go | 713 | fast: нет return_event_id | `fleet.fast_нет_returneventid` | ❌ |
| backend/internal/fleet/moon_destruction.go | 106 | Ваша луна уничтожена | `fleet.ваша_луна_уничтожена` | ❌ |
| backend/internal/fleet/moon_destruction.go | 107 | Атакующий с %d Deathstar разрушил вашу луну (P=%.1f%%, ro… | `fleet.атакующий_с_d` | ❌ |
| backend/internal/fleet/moon_destruction.go | 112 | Луна противника уничтожена | `fleet.луна_противника_уничтожена` | ❌ |
| backend/internal/fleet/moon_destruction.go | 113 | Ваши %d Deathstar успешно разрушили луну (P=%.1f%%, roll=… | `fleet.ваши_d_deathstar` | ❌ |
| backend/internal/fleet/moon_destruction.go | 126 | Ваш Deathstar-флот уничтожен | `fleet.ваш_deathstarфлот_уничтожен` | ❌ |
| backend/internal/fleet/moon_destruction.go | 127 | При попытке разрушить луну все Deathstar взорвались (P=%.… | `fleet.при_попытке_разрушить` | ❌ |
| backend/internal/fleet/moon_destruction.go | 182 | Ваша луна уничтожена | `fleet.ваша_луна_уничтожена` | ❌ |
| backend/internal/fleet/moon_destruction.go | 183 | ACS-атака с %d Deathstar разрушила вашу луну (P=%.1f%%, r… | `fleet.acsатака_с_d` | ❌ |
| backend/internal/fleet/moon_destruction.go | 186 | Луна противника уничтожена (ACS) | `fleet.луна_противника_уничтожена` | ❌ |
| backend/internal/fleet/moon_destruction.go | 187 | Объединённый флот с %d Deathstar разрушил луну (P=%.1f%%,… | `fleet.объединённый_флот_с` | ❌ |
| backend/internal/fleet/moon_destruction.go | 209 | Ваши Deathstar взорвались (ACS) | `fleet.ваши_deathstar_взорвались` | ❌ |
| backend/internal/fleet/moon_destruction.go | 210 | При ACS-попытке разрушить луну все Deathstar взорвались (… | `fleet.при_acsпопытке_разрушить` | ❌ |
| backend/internal/fleet/raid_warning.go | 63 | Предупреждение: вражеский флот! | `fleet.предупреждение_вражеский_флот` | ❌ |
| backend/internal/fleet/raid_warning.go | 65 | Вражеский флот приближается к вашей планете [%d:%d:%d]. О… | `fleet.вражеский_флот_приближается` | ❌ |
| backend/internal/fleet/spy.go | 121 | Зонды перехвачены %d:%d:%d | `fleet.зонды_перехвачены_ddd` | ❌ |
| backend/internal/fleet/spy.go | 122 | Все %d зонд(ов) уничтожены обороной цели. Флот потерян. | `fleet.все_d_зондов` | ❌ |
| backend/internal/fleet/spy.go | 152 | Разведка %d:%d:%d (ratio=%d) | `fleet.разведка_ddd_ratiod` | ❌ |
| backend/internal/fleet/spy.go | 153 | Металл: %d, Кремний: %d, Водород: %d | `fleet.металл_d_кремний` | ❌ |
| backend/internal/fleet/spy.go | 163 | Вас шпионили %d:%d:%d | `fleet.вас_шпионили_ddd` | ❌ |
| backend/internal/fleet/spy.go | 164 | Противник послал %d зонд(ов). Соотношение %d. | `fleet.противник_послал_d` | ❌ |
| backend/internal/goal/notifier.go | 30 | Цель: %s | `global.цель_s` | ❌ |
| backend/internal/goal/notifier.go | 33 | Вы завершили цель «%s». | `global.вы_завершили_цель` | ❌ |
| backend/internal/officer/service.go | 321 | Officer %s продлён автоматически | `officer.officer_s_продлён` | ❌ |
| backend/internal/officer/service.go | 322 | Подписка продлена. Списано %d кредитов. | `officer.подписка_продлена_списано` | ❌ |
| backend/internal/officer/service.go | 335 | Officer %s истёк | `officer.officer_s_истёк` | ❌ |
| backend/internal/officer/service.go | 336 | Срок подписки закончился. Активируйте снова, если нужно. | `officer.срок_подписки_закончился` | ❌ |
| backend/internal/officer/service.go | 338 | Срок подписки закончился. Недостаточно кредитов для авто-… | `officer.срок_подписки_закончился` | ❌ |
| backend/internal/payment/packages.go | 17 | Пробный | `payment.пробный` | ❌ |
| backend/internal/payment/packages.go | 18 | Стартовый | `payment.стартовый` | ❌ |
| backend/internal/payment/packages.go | 19 | Средний | `payment.средний` | ✅ |
| backend/internal/payment/packages.go | 20 | Большой | `payment.большой` | ❌ |
| backend/internal/payment/packages.go | 21 | Максимальный | `payment.максимальный` | ❌ |
| backend/internal/payment/service.go | 158 | Кредиты зачислены | `payment.кредиты_зачислены` | ❌ |
| backend/internal/payment/service.go | 159 | На ваш счёт зачислено %d кредитов (заказ #%s). Спасибо за… | `payment.на_ваш_счёт` | ❌ |
| backend/internal/referral/service.go | 122 | Реферальный бонус | `referral.реферальный_бонус` | ❌ |
| backend/internal/referral/service.go | 123 | Ваш реферал %s совершил покупку. На ваш счёт зачислено %.… | `referral.ваш_реферал_s` | ❌ |
| backend/internal/rocket/events.go | 71 | Ракетный удар %d:%d:%d провалился | `rocket.ракетный_удар_ddd` | ❌ |
| backend/internal/rocket/events.go | 73 | Цель не найдена (уничтожена или не существовала). | `rocket.цель_не_найдена` | ❌ |
| backend/internal/rocket/events.go | 134 | Ракетный удар %d:%d:%d | `rocket.ракетный_удар_ddd` | ❌ |
| backend/internal/rocket/events.go | 137 | %d ракет перехвачено антиракетами (%d ABM). Урон отсутств… | `rocket.d_ракет_перехвачено` | ❌ |
| backend/internal/rocket/events.go | 140 | %d ракет долетели. Оборона отсутствует — урон %d пропал. | `rocket.d_ракет_долетели` | ❌ |
| backend/internal/rocket/events.go | 167 | Ракетный удар %d:%d:%d | `rocket.ракетный_удар_ddd` | ❌ |
| backend/internal/rocket/events.go | 170 | \n- юнит #%d: -%d | `rocket.n_юнит_d` | ❌ |
| backend/internal/rocket/events.go | 174 | (%d сбито ABM) | `rocket.d_сбито_abm` | ❌ |
| backend/internal/rocket/events.go | 176 | %d/%d ракет долетели%s (урон %d). Потери обороны:%s | `rocket.dd_ракет_долетелиs` | ❌ |
| backend/internal/settings/delete.go | 98 | Код для удаления аккаунта | `settings.код_для_удаления` | ❌ |
| backend/internal/settings/delete.go | 100 | Код подтверждения удаления аккаунта: %s\nДействителен до … | `settings.код_подтверждения_удаления` | ❌ |
| backend/internal/wiki/service.go | 47 | Здания | `wiki.здания` | ✅ |
| cmd/tools/battle-sim/main.go | 40 | имя сценария | `global.имя_сценария` | ❌ |
| cmd/tools/battle-sim/main.go | 41 | прогнать все сценарии | `global.прогнать_все_сценарии` | ❌ |
| cmd/tools/battle-sim/main.go | 42 | матрица 1v1: каждый combat-юнит vs каждый при равной meta… | `global.матрица_1v1_каждый` | ❌ |
| cmd/tools/battle-sim/main.go | 43 | metal-eq на сторону для --matrix (default 10M) | `global.metaleq_на_сторону` | ❌ |
| cmd/tools/battle-sim/main.go | 44 | группы vs группы (lite/mid/capital/endgame) | `global.группы_vs_группы` | ❌ |
| cmd/tools/battle-sim/main.go | 45 | число прогонов на сценарий | `global.число_прогонов_на` | ❌ |
| cmd/tools/battle-sim/main.go | 46 | макс раундов в одном бою | `global.макс_раундов_в` | ❌ |
| cmd/tools/battle-sim/main.go | 47 | путь к configs/ | `global.путь_к_configs` | ❌ |
| cmd/tools/battle-sim/main.go | 51 | переопределить стоимость юнита: ship_key=M/Si/H | `global.переопределить_стоимость_юнита` | ❌ |
| cmd/tools/battle-sim/main.go | 52 | переопределить front юнита: ship_key=N | `global.переопределить_front_юнита` | ❌ |
| cmd/tools/battle-sim/main.go | 99 | доступные сценарии: | `global.доступные_сценарии` | ❌ |
| cmd/tools/battle-sim/main.go | 144 | Cell = exchange ratio (def_loss/atk_loss). >1 = атакующий… | `global.cell_exchange_ratio` | ❌ |
| cmd/tools/battle-sim/main.go | 301 | === GROUPS vs GROUPS (равные бюджеты) === | `global.groups_vs_groups` | ❌ |
| cmd/tools/battle-sim/main.go | 448 | BA-002 проверка: 1000 Lancer (25M-eq) vs 1000 Cruiser (29… | `global.ba002_проверка_1000` | ❌ |
| cmd/tools/battle-sim/main.go | 458 | BA-002: 500 Lancer атакуют микс 200 LF + 100 Cruiser + 50… | `global.ba002_500_lancer` | ❌ |
| cmd/tools/battle-sim/main.go | 470 | BA-001/002: 1 DS (10M-eq) vs 300 Lancer (7.5M-eq). DS×100… | `global.ba001002_1_ds` | ❌ |
| cmd/tools/battle-sim/main.go | 480 | BA-001: 1 DS vs 200 Battleship (12M-eq). BS пробивает DS … | `global.ba001_1_ds` | ❌ |
| cmd/tools/battle-sim/main.go | 490 | BA-001: 1 DS vs смесь 100 BS + 50 SD (12.25M-eq). Ни у BS… | `global.ba001_1_ds` | ❌ |
| cmd/tools/battle-sim/main.go | 501 | Balance: 100 Bomber (9M) vs 3000 RocketLauncher (6M). Bom… | `global.balance_100_bomber` | ❌ |
| cmd/tools/battle-sim/main.go | 511 | Balance: 200 Cruiser (5.8M) vs 2000 RocketLauncher (4M). … | `global.balance_200_cruiser` | ❌ |
| cmd/tools/battle-sim/main.go | 521 | Общий sanity: 500 LF + 200 Cruiser + 50 BS vs 1000 RL + 3… | `global.общий_sanity_500` | ❌ |
| cmd/tools/battle-sim/main.go | 539 | Эксплойт: 50000 Solar Satellite (125M-eq) vs 1 DS. SSat д… | `global.эксплойт_50000_solar` | ❌ |
| cmd/tools/battle-sim/main.go | 545 | Защита: 10000 SSat (25M-eq) + 50 BS прикрывает от 5 DS. П… | `global.защита_10000_ssat` | ❌ |
| cmd/tools/battle-sim/main.go | 556 | Frigate-role: 200 Frigate (17M-eq) vs 600 Cruiser (17.4M-… | `global.frigaterole_200_frigate` | ❌ |
| cmd/tools/battle-sim/main.go | 562 | Frigate-role: 200 Frigate (17M-eq) vs 200 BS (12M-eq) — F… | `global.frigaterole_200_frigate` | ❌ |
| cmd/tools/battle-sim/main.go | 576 | Shadow anti-DS (ADR-0007): 100 Shadow (500k-eq) vs 1 DS —… | `global.shadow_antids_adr0007` | ❌ |
| cmd/tools/battle-sim/main.go | 582 | Shadow mass: 1000 Shadow (5M-eq) vs 1 DS — масштабная ата… | `global.shadow_mass_1000` | ❌ |
| cmd/tools/battle-sim/main.go | 588 | Shadow vs средний флот: 500 Shadow (2.5M-eq) vs 100 BS + … | `global.shadow_vs_средний` | ❌ |
| cmd/tools/battle-sim/main.go | 600 | SD vs BS (одинаковая metal-eq): 50 SD (6.25M) vs 100 BS (… | `global.sd_vs_bs` | ❌ |
| cmd/tools/battle-sim/main.go | 606 | SD имеет RF×2 vs Frigate (legacy). 50 SD (6.25M) vs 100 F… | `global.sd_имеет_rf2` | ❌ |
| cmd/tools/battle-sim/main.go | 614 | Mass-BS: 1 DS vs 1000 BS (60M-eq). 6× ресурсов vs DS — до… | `global.massbs_1_ds` | ❌ |
| cmd/tools/battle-sim/main.go | 620 | Bomber vs DS: 200 Bomber (18M) vs 1 DS. У Bomber нет RF v… | `global.bomber_vs_ds` | ❌ |
| cmd/tools/battle-sim/main.go | 626 | Plasma защищает планету: 1 DS vs 50 Plasma Gun (6.5M). Pl… | `global.plasma_защищает_планету` | ❌ |
| cmd/tools/battle-sim/main.go | 632 | Gauss-wall: 1 DS vs 200 Gauss (7.4M). Gauss attack=1100 (… | `global.gausswall_1_ds` | ❌ |
| cmd/tools/battle-sim/main.go | 655 | Bomber специалист по defense: 100 Bomber (9M) vs 50 Plasm… | `global.bomber_специалист_по` | ❌ |
| cmd/tools/battle-sim/main.go | 661 | Bomber vs Gauss: 100 Bomber (9M) vs 100 Gauss (3.7M). У B… | `global.bomber_vs_gauss` | ❌ |
| cmd/tools/battle-sim/main.go | 675 | SF vs LF (одинаковая metal-eq): 250 SF (2.5M) vs 600 LF (… | `global.sf_vs_lf` | ❌ |
| cmd/tools/battle-sim/main.go | 683 | Lancer vs тяжёлый флот: 500 Lancer (55M) vs 1000 BS (60M). | `global.lancer_vs_тяжёлый` | ❌ |
| cmd/tools/battle-sim/main.go | 689 | Lancer vs SF (ловит на дешёвом флоте): 100 Lancer (11M) v… | `global.lancer_vs_sf` | ❌ |
| cmd/tools/battle-sim/main.go | 703 | Уязвимость Recycler: 100 BS + 100 Recycler (7.8M) vs 200 … | `global.уязвимость_recycler_100` | ❌ |
| cmd/tools/battle-sim/main.go | 714 | Endgame: 100 BS + 50 SD + 200 Cruiser vs 100 BS + 50 SD +… | `global.endgame_100_bs` | ❌ |
| cmd/tools/battle-sim/main.go | 742 | Small Shield (front=16) перетягивает огонь: 200 Cruiser v… | `global.small_shield_front16` | ❌ |
| cmd/tools/battle-sim/main.go | 751 | Large Shield (front=17) — ультра-приоритетная цель. 1000 … | `global.large_shield_front17` | ❌ |
| cmd/tools/battle-sim/main.go | 770 | ESensor мягкая цель: 100 BS + 50 ESensor (6.05M) vs 200 C… | `global.esensor_мягкая_цель` | ❌ |
| cmd/tools/battle-sim/main.go | 781 | Plasma stack vs BS: 200 Plasma (26M) vs 1000 BS (60M). Ка… | `global.plasma_stack_vs` | ❌ |
| cmd/tools/battle-sim/main.go | 789 | Эффективность атаки: 50 BS + 100 Cru + 500 LF (10.8M) vs … | `global.эффективность_атаки_50` | ❌ |
| cmd/tools/battle-sim/main.go | 820 | Probe-spam: 1000 ESensor (1M) vs 100 Cruiser. Probe не уг… | `global.probespam_1000_esensor` | ❌ |
| cmd/tools/battle-sim/main.go | 843 | Большой флот атакует DS: 500 BS + 200 SD (55M) vs 5 DS (5… | `global.большой_флот_атакует` | ❌ |
| cmd/tools/battle-sim/main.go | 854 | Transport-уязвимость: 500 LF + 100 LT (12M) vs 200 Cruise… | `global.transportуязвимость_500_lf` | ❌ |
| cmd/tools/battle-sim/main.go | 865 | Mix-test: 200 Shadow + 100 BS (7M) vs 200 Cruiser. Низкий… | `global.mixtest_200_shadow` | ❌ |
| cmd/tools/battle-sim/main.go | 876 | Recycler в смеси: 100 BS + 50 Recycler (6.5M) vs 200 Crui… | `global.recycler_в_смеси` | ❌ |
| cmd/tools/battle-sim/main.go | 886 | Превосходство 3:1 — 1500 BS + 500 SD (152.5M) vs 5 DS (50… | `global.превосходство_31_1500` | ❌ |
| cmd/tools/battle-sim/main.go | 895 | Anti-DS через Shadow: 5000 Shadow (25M) vs 5 DS (50M). Ме… | `global.antids_через_shadow` | ❌ |
| cmd/tools/battle-sim/main.go | 901 | Anti-DS через Lancer: 1000 Lancer (110M) vs 5 DS (50M). R… | `global.antids_через_lancer` | ❌ |
| cmd/tools/battle-sim/main.go | 909 | Паритет 1.2:1 — 3 DS + 50 BS (33M) vs 200 Plasma + 100 Ga… | `global.паритет_121_3` | ❌ |
| cmd/tools/battle-sim/main.go | 921 | Паритет: 100 BS + 200 Cruiser (17.8M) vs 100 Plasma + 50 … | `global.паритет_100_bs` | ❌ |
| cmd/tools/battle-sim/main.go | 935 | Эскорт колонизатора: 5 Colony + 50 BS (3.2M) vs 200 LF. | `global.эскорт_колонизатора_5` | ❌ |
| cmd/tools/i18n-audit/main.go | 72 | корень проекта (где frontend/ и backend/) | `global.корень_проекта_где` | ❌ |
| cmd/tools/i18n-audit/main.go | 73 | путь к ru.yml | `global.путь_к_ruyml` | ❌ |
| cmd/tools/i18n-audit/main.go | 74 | выходной файл отчёта | `global.выходной_файл_отчёта` | ❌ |
| cmd/tools/i18n-audit/main.go | 116 | Всего хардкод-строк: %d\n | `global.всего_хардкодстрок_dn` | ❌ |
| cmd/tools/i18n-audit/main.go | 117 | Уже есть в ru.yml: %d (%.0f%%)\n | `global.уже_есть_в` | ❌ |
| cmd/tools/i18n-audit/main.go | 118 | Нужно добавить: %d\n | `global.нужно_добавить_dn` | ❌ |
| cmd/tools/i18n-audit/main.go | 119 | Отчёт: %s\n | `global.отчёт_sn` | ❌ |
| cmd/tools/i18n-audit/main.go | 365 | Всего: **%d** хардкод-строк с кириллицей.\n | `global.всего_d_хардкодстрок` | ❌ |
| cmd/tools/i18n-audit/main.go | 372 | Уже в ru.yml: **%d**. Нужно добавить: **%d**.\n | `global.уже_в_ruyml` | ❌ |
| cmd/tools/i18n-audit/main.go | 374 | \| Файл \| Строка \| Литерал \| Предложенный ключ \| В ru… | `global.файл_строка_литерал` | ❌ |
| cmd/tools/import-phrases/main.go | 34 | путь к na_phrases.sql (обязательно) | `global.путь_к_naphrasessql` | ❌ |
| cmd/tools/import-phrases/main.go | 35 | папка configs/i18n/ для записи *.yml | `global.папка_configsi18n_для` | ❌ |
| cmd/tools/resync/main.go | 34 | UUID пользователя для пересчёта | `global.uuid_пользователя_для` | ❌ |
| cmd/tools/resync/main.go | 35 | пересчитать всех пользователей (осторожно) | `global.пересчитать_всех_пользователей` | ❌ |
| cmd/tools/testseed/main.go | 52 | truncate игровые таблицы перед сидом | `global.truncate_игровые_таблицы` | ❌ |
| features/achievements/AchievementsScreen.tsx | 20 | 🎓 Старт | `global.старт` | ❌ |
| features/achievements/AchievementsScreen.tsx | 21 | 🥇 Достижения | `global.достижения` | ❌ |
| features/achievements/AchievementsScreen.tsx | 121 | 📋 Все | `global.все` | ❌ |
| features/admin/AdminScreen.tsx | 96 | Панель администратора | `admin.панель_администратора` | ❌ |
| features/admin/AdminScreen.tsx | 116 | Пользователей | `admin.пользователей` | ❌ |
| features/admin/AdminScreen.tsx | 117 | Планет | `admin.планет` | ❌ |
| features/admin/AdminScreen.tsx | 118 | Флотов в пути | `admin.флотов_в_пути` | ❌ |
| features/admin/AdminScreen.tsx | 119 | Событий в очереди | `admin.событий_в_очереди` | ❌ |
| features/admin/AdminScreen.tsx | 123 | Действия | `admin.действия` | ❌ |
| features/admin/AdminScreen.tsx | 126 | Начислить кредиты | `admin.начислить_кредиты` | ❌ |
| features/admin/AdminScreen.tsx | 136 | сумма | `admin.сумма` | ❌ |
| features/admin/AdminScreen.tsx | 146 | Начислить ${creditAmount} кредитов игроку ${creditUserID.… | `admin.начислить_creditamount_кредитов` | ❌ |
| features/admin/AdminScreen.tsx | 158 | Установить роль | `admin.установить_роль` | ❌ |
| features/admin/AdminScreen.tsx | 177 | Назначить роль "${roleValue \|\| 'user'}" игроку ${roleUs… | `admin.назначить_роль_rolevalue` | ❌ |
| features/admin/AdminScreen.tsx | 190 | Пользователи | `admin.пользователи` | ❌ |
| features/admin/AdminScreen.tsx | 196 | Игрок | `admin.игрок` | ✅ |
| features/admin/AdminScreen.tsx | 197 | Роль | `admin.роль` | ❌ |
| features/admin/AdminScreen.tsx | 198 | Кредиты | `admin.кредиты` | ✅ |
| features/admin/AdminScreen.tsx | 199 | Очки | `admin.очки` | ✅ |
| features/admin/AdminScreen.tsx | 200 | Создан | `admin.создан` | ❌ |
| features/admin/AdminScreen.tsx | 201 | Действия | `admin.действия` | ❌ |
| features/admin/AdminScreen.tsx | 220 | Открыть полный профиль игрока | `admin.открыть_полный_профиль` | ❌ |
| features/admin/AdminScreen.tsx | 228 | Снять бан с игрока ${u.username}? | `admin.снять_бан_с` | ❌ |
| features/admin/AdminScreen.tsx | 236 | Забанить игрока ${u.username}? Он не сможет войти в игру,… | `admin.забанить_игрока_uusername` | ❌ |
| features/admin/AdminScreen.tsx | 258 | Подтверждение админского действия | `admin.подтверждение_админского_действия` | ❌ |
| features/admin/AdminScreen.tsx | 260 | Выполнить | `admin.выполнить` | ❌ |
| features/admin/AdminScreen.tsx | 299 | Шаблоны сообщений | `admin.шаблоны_сообщений` | ❌ |
| features/admin/AdminScreen.tsx | 305 | Ключ | `admin.ключ` | ❌ |
| features/admin/AdminScreen.tsx | 306 | Заголовок | `admin.заголовок` | ❌ |
| features/admin/AdminScreen.tsx | 307 | Папка | `admin.папка` | ❌ |
| features/admin/AdminScreen.tsx | 370 | ошибка | `admin.ошибка` | ❌ |
| features/admin/AdminScreen.tsx | 431 | Монитор событий | `admin.монитор_событий` | ❌ |
| features/admin/AdminScreen.tsx | 470 | Загрузка… | `admin.загрузка` | ❌ |
| features/admin/AdminScreen.tsx | 472 | нет событий | `admin.нет_событий` | ❌ |
| features/admin/AdminScreen.tsx | 571 | Загрузка… | `admin.загрузка` | ❌ |
| features/admin/AdminScreen.tsx | 572 | Ошибка загрузки dead-letter | `admin.ошибка_загрузки_deadletter` | ❌ |
| features/admin/AdminScreen.tsx | 574 | Пусто. События попадают сюда после N неудачных попыток. | `admin.пусто_события_попадают` | ❌ |
| features/admin/AdminScreen.tsx | 583 | Ошибка | `admin.ошибка` | ✅ |
| features/admin/AdminScreen.tsx | 609 | Вернуть событие ${e.id.slice(0, 8)} в активную очередь? | `admin.вернуть_событие_eidslice0` | ❌ |
| features/admin/AdminScreen.tsx | 684 | ↻ Обновить | `admin.обновить` | ❌ |
| features/admin/AdminScreen.tsx | 687 | Загрузка… | `admin.загрузка` | ❌ |
| features/admin/AdminScreen.tsx | 688 | Ошибка загрузки журнала | `admin.ошибка_загрузки_журнала` | ❌ |
| features/admin/AdminScreen.tsx | 690 | Журнал пуст. Выполните любое write-действие в админке — з… | `admin.журнал_пуст_выполните` | ❌ |
| features/admin/AdminScreen.tsx | 697 | Дата | `admin.дата` | ✅ |
| features/admin/AdminScreen.tsx | 698 | Админ | `admin.админ` | ❌ |
| features/admin/AdminScreen.tsx | 699 | Действие | `admin.действие` | ✅ |
| features/admin/AdminScreen.tsx | 700 | Цель | `admin.цель` | ✅ |
| features/admin/AdminUserProfilePanel.tsx | 87 | Загрузка… | `admin.загрузка` | ❌ |
| features/admin/AdminUserProfilePanel.tsx | 89 | ✕ Закрыть | `admin.закрыть` | ❌ |
| features/admin/AdminUserProfilePanel.tsx | 92 | Загрузка профиля… | `admin.загрузка_профиля` | ❌ |
| features/admin/AdminUserProfilePanel.tsx | 93 | Ошибка загрузки | `admin.ошибка_загрузки` | ❌ |
| features/admin/AdminUserProfilePanel.tsx | 98 | Роль: | `admin.роль` | ❌ |
| features/admin/AdminUserProfilePanel.tsx | 99 | 💳 Кредитов: | `admin.кредитов` | ❌ |
| features/admin/AdminUserProfilePanel.tsx | 100 | 🏆 Очков: | `admin.очков` | ❌ |
| features/admin/AdminUserProfilePanel.tsx | 101 | Зарегистрирован: | `admin.зарегистрирован` | ❌ |
| features/admin/AdminUserProfilePanel.tsx | 102 | Был онлайн: | `admin.был_онлайн` | ❌ |
| features/admin/AdminUserProfilePanel.tsx | 107 | Общее | `admin.общее` | ❌ |
| features/admin/AdminUserProfilePanel.tsx | 108 | Экономика | `admin.экономика` | ❌ |
| features/admin/AdminUserProfilePanel.tsx | 109 | Бои / отчёты | `admin.бои_отчёты` | ❌ |
| features/admin/AdminUserProfilePanel.tsx | 110 | Журнал | `admin.журнал` | ❌ |
| features/admin/AdminUserProfilePanel.tsx | 128 | 🪐 Планеты (${p.planets.length}) | `admin.планеты_pplanetslength` | ❌ |
| features/admin/AdminUserProfilePanel.tsx | 132 | Имя | `admin.имя` | ❌ |
| features/admin/AdminUserProfilePanel.tsx | 132 | Координаты | `admin.координаты` | ✅ |
| features/admin/AdminUserProfilePanel.tsx | 132 | 🟠 Метал | `admin.метал` | ❌ |
| features/admin/AdminUserProfilePanel.tsx | 132 | 💎 Крем | `admin.крем` | ❌ |
| features/admin/AdminUserProfilePanel.tsx | 132 | 💧 Водор | `admin.водор` | ❌ |
| features/admin/AdminUserProfilePanel.tsx | 149 | ⚙️ Изменить ресурсы (delta, может быть отрицательным) | `admin.изменить_ресурсы_delta` | ❌ |
| features/admin/AdminUserProfilePanel.tsx | 153 | 🛸 Флоты в полёте (${p.fleets.length}) | `admin.флоты_в_полёте` | ❌ |
| features/admin/AdminUserProfilePanel.tsx | 156 | Миссия | `admin.миссия` | ❌ |
| features/admin/AdminUserProfilePanel.tsx | 156 | Куда | `admin.куда` | ❌ |
| features/admin/AdminUserProfilePanel.tsx | 156 | Прилёт | `admin.прилёт` | ❌ |
| features/admin/AdminUserProfilePanel.tsx | 171 | ⭐ Офицеры (${p.officers.length}) | `admin.офицеры_pofficerslength` | ❌ |
| features/admin/AdminUserProfilePanel.tsx | 177 | 💎 Артефакты (${p.artefacts.length}) | `admin.артефакты_partefactslength` | ❌ |
| features/admin/AdminUserProfilePanel.tsx | 187 | 📈 Лоты рынка (${p.market_lots.length}) | `admin.лоты_рынка_pmarketlotslength` | ❌ |
| features/admin/AdminUserProfilePanel.tsx | 190 | Продаёт | `admin.продаёт` | ❌ |
| features/admin/AdminUserProfilePanel.tsx | 190 | Хочет | `admin.хочет` | ❌ |
| features/admin/AdminUserProfilePanel.tsx | 204 | 🏪 Лоты арт-рынка (${p.artefact_lots.length}) | `admin.лоты_артрынка_partefactlotslength` | ❌ |
| features/admin/AdminUserProfilePanel.tsx | 207 | Артефакт | `admin.артефакт` | ✅ |
| features/admin/AdminUserProfilePanel.tsx | 207 | Цена | `admin.цена` | ❌ |
| features/admin/AdminUserProfilePanel.tsx | 221 | 💰 Последние транзакции ресурсов (${p.res_log.length}) | `admin.последние_транзакции_ресурсов` | ❌ |
| features/admin/AdminUserProfilePanel.tsx | 224 | Дата | `admin.дата` | ✅ |
| features/admin/AdminUserProfilePanel.tsx | 240 | 💳 Покупки кредитов (${p.purchases.length}) | `admin.покупки_кредитов_ppurchaseslength` | ❌ |
| features/admin/AdminUserProfilePanel.tsx | 243 | Дата | `admin.дата` | ✅ |
| features/admin/AdminUserProfilePanel.tsx | 243 | Пакет | `admin.пакет` | ❌ |
| features/admin/AdminUserProfilePanel.tsx | 243 | +Кр | `admin.кр` | ❌ |
| features/admin/AdminUserProfilePanel.tsx | 243 | Цена | `admin.цена` | ❌ |
| features/admin/AdminUserProfilePanel.tsx | 243 | Статус | `admin.статус` | ✅ |
| features/admin/AdminUserProfilePanel.tsx | 265 | 📜 Последние отчёты (${p.reports_recent.length}) | `admin.последние_отчёты_preportsrecentlength` | ❌ |
| features/admin/AdminUserProfilePanel.tsx | 277 | 📬 Последние сообщения (${p.messages_recent.length}) | `admin.последние_сообщения_pmessagesrecentlength` | ❌ |
| features/admin/AdminUserProfilePanel.tsx | 280 | Дата | `admin.дата` | ✅ |
| features/admin/AdminUserProfilePanel.tsx | 280 | Тема | `admin.тема` | ✅ |
| features/admin/AdminUserProfilePanel.tsx | 307 | 🗒 Админские действия над этим игроком (${entries.length}) | `admin.админские_действия_над` | ❌ |
| features/admin/AdminUserProfilePanel.tsx | 310 | Действие | `admin.действие` | ✅ |
| features/admin/AdminUserProfilePanel.tsx | 310 | Статус | `admin.статус` | ✅ |
| features/admin/AdminUserProfilePanel.tsx | 310 | Админ | `admin.админ` | ❌ |
| features/admin/AdminUserProfilePanel.tsx | 310 | Дата | `admin.дата` | ✅ |
| features/admin/AdminUserProfilePanel.tsx | 360 | Ошибка | `admin.ошибка` | ✅ |
| features/admin/AdminUserProfilePanel.tsx | 380 | + Выдать | `admin.выдать` | ❌ |
| features/admin/AdminUserProfilePanel.tsx | 453 | Колонизация | `admin.колонизация` | ✅ |
| features/admin/AdminUserProfilePanel.tsx | 453 | Транспорт | `admin.транспорт` | ✅ |
| features/admin/AdminUserProfilePanel.tsx | 453 | Разведка | `admin.разведка` | ❌ |
| features/admin/AdminUserProfilePanel.tsx | 453 | Атака | `admin.атака` | ✅ |
| features/admin/AdminUserProfilePanel.tsx | 454 | Вторжение | `admin.вторжение` | ❌ |
| features/admin/AdminUserProfilePanel.tsx | 454 | Экспедиция | `admin.экспедиция` | ✅ |
| features/admin/AdminUserProfilePanel.tsx | 454 | Сбор обломков | `admin.сбор_обломков` | ❌ |
| features/admin/AdminUserProfilePanel.tsx | 458 | ⚔ Бой | `admin.бой` | ❌ |
| features/admin/AdminUserProfilePanel.tsx | 458 | 🌌 Экспедиция | `admin.экспедиция` | ❌ |
| features/admin/AdminUserProfilePanel.tsx | 458 | 🔭 Шпионаж | `admin.шпионаж` | ❌ |
| features/alien/AlienHoldingPanel.tsx | 27 | истекает | `alien.истекает` | ❌ |
| features/alien/AlienHoldingPanel.tsx | 30 | ${Math.floor(h / 24)}д ${h % 24}ч | `alien.mathfloorh_24д_h` | ❌ |
| features/alien/AlienHoldingPanel.tsx | 31 | ${h}ч ${m}м | `alien.hч_mм` | ❌ |
| features/alien/AlienHoldingPanel.tsx | 32 | ${m}м | `alien.mм` | ❌ |
| features/alien/AlienHoldingPanel.tsx | 62 | Продлено на ${minutes} мин (достигнут лимит 15 дней) | `alien.продлено_на_minutes` | ❌ |
| features/alien/AlienHoldingPanel.tsx | 63 | Продлено на ${minutes} мин | `alien.продлено_на_minutes` | ❌ |
| features/alien/AlienHoldingPanel.tsx | 64 | Платёж принят | `alien.платёж_принят` | ❌ |
| features/alien/AlienHoldingPanel.tsx | 68 | Ошибка платежа | `alien.ошибка_платежа` | ❌ |
| features/alien/AlienHoldingPanel.tsx | 69 | Не удалось оплатить | `alien.не_удалось_оплатить` | ❌ |
| features/alien/AlienHoldingPanel.tsx | 80 | Планеты под захватом пришельцев | `alien.планеты_под_захватом` | ❌ |
| features/alliance/AllianceScreen.tsx | 46 | СОЮЗ | `alliance.союз` | ❌ |
| features/alliance/AllianceScreen.tsx | 46 | ВОЙНА | `alliance.война` | ❌ |
| features/alliance/AllianceScreen.tsx | 46 | НЕН | `alliance.нен` | ❌ |
| features/alliance/AllianceScreen.tsx | 79 | Альянс | `alliance.альянс` | ✅ |
| features/alliance/AllianceScreen.tsx | 79 | Заявка отправлена | `alliance.заявка_отправлена` | ❌ |
| features/alliance/AllianceScreen.tsx | 81 | Ошибка | `alliance.ошибка` | ✅ |
| features/alliance/AllianceScreen.tsx | 89 | Альянс покинут | `alliance.альянс_покинут` | ❌ |
| features/alliance/AllianceScreen.tsx | 91 | Ошибка | `alliance.ошибка` | ✅ |
| features/alliance/AllianceScreen.tsx | 99 | Альянс распущен | `alliance.альянс_распущен` | ✅ |
| features/alliance/AllianceScreen.tsx | 101 | Ошибка | `alliance.ошибка` | ✅ |
| features/alliance/AllianceScreen.tsx | 156 | [Тег] | `alliance.тег` | ❌ |
| features/alliance/AllianceScreen.tsx | 157 | Название | `alliance.название` | ✅ |
| features/alliance/AllianceScreen.tsx | 158 | Игроков | `alliance.игроков` | ❌ |
| features/alliance/AllianceScreen.tsx | 159 | Тип | `alliance.тип` | ❌ |
| features/alliance/AllianceScreen.tsx | 221 | Ошибка | `alliance.ошибка` | ✅ |
| features/alliance/AllianceScreen.tsx | 252 | 🔓 Открытый | `alliance.открытый` | ❌ |
| features/alliance/AllianceScreen.tsx | 252 | 🔒 Закрытый | `alliance.закрытый` | ❌ |
| features/alliance/AllianceScreen.tsx | 269 | 🔓 Открыть (вход) | `alliance.открыть_вход` | ❌ |
| features/alliance/AllianceScreen.tsx | 269 | 🔒 Закрыть (заявки) | `alliance.закрыть_заявки` | ❌ |
| features/alliance/AllianceScreen.tsx | 299 | Нет заявок. | `alliance.нет_заявок` | ❌ |
| features/alliance/AllianceScreen.tsx | 307 | ✓ Принять | `alliance.принять` | ❌ |
| features/alliance/AllianceScreen.tsx | 316 | Распустить альянс | `alliance.распустить_альянс` | ✅ |
| features/alliance/AllianceScreen.tsx | 317 | Распустить альянс? Это действие необратимо. | `alliance.распустить_альянс_это` | ❌ |
| features/alliance/AllianceScreen.tsx | 318 | Распустить | `alliance.распустить` | ❌ |
| features/alliance/AllianceScreen.tsx | 346 | 🔓 Открытый | `alliance.открытый` | ❌ |
| features/alliance/AllianceScreen.tsx | 346 | 🔒 Закрытый | `alliance.закрытый` | ❌ |
| features/alliance/AllianceScreen.tsx | 355 | Ранг | `alliance.ранг` | ❌ |
| features/alliance/AllianceScreen.tsx | 355 | Игрок | `alliance.игрок` | ✅ |
| features/alliance/AllianceScreen.tsx | 375 | Сопроводительное сообщение (необязательно) | `alliance.сопроводительное_сообщение_необязательно` | ❌ |
| features/alliance/AllianceScreen.tsx | 379 | 📨 Подать заявку | `alliance.подать_заявку` | ❌ |
| features/alliance/AllianceScreen.tsx | 379 | 🤝 Вступить | `alliance.вступить` | ❌ |
| features/alliance/AllianceScreen.tsx | 395 | Альянс создан | `alliance.альянс_создан` | ❌ |
| features/alliance/AllianceScreen.tsx | 396 | Ошибка | `alliance.ошибка` | ✅ |
| features/alliance/AllianceScreen.tsx | 401 | Создать альянс | `alliance.создать_альянс` | ✅ |
| features/alliance/AllianceScreen.tsx | 404 | Тег (3–5 символов) | `alliance.тег_35_символов` | ❌ |
| features/alliance/AllianceScreen.tsx | 414 | Название | `alliance.название` | ✅ |
| features/alliance/AllianceScreen.tsx | 418 | Описание | `alliance.описание` | ❌ |
| features/alliance/AllianceScreen.tsx | 424 | Описание альянса… | `alliance.описание_альянса` | ❌ |
| features/alliance/AllianceScreen.tsx | 435 | 🤝 Создать | `alliance.создать` | ❌ |
| features/alliance/AllianceScreen.tsx | 437 | Отмена | `alliance.отмена` | ✅ |
| features/alliance/AllianceScreen.tsx | 459 | Ошибка | `alliance.ошибка` | ✅ |
| features/alliance/AllianceScreen.tsx | 486 | Нет установленных отношений. | `alliance.нет_установленных_отношений` | ❌ |
| features/alliance/AllianceScreen.tsx | 492 | Отношение | `alliance.отношение` | ❌ |
| features/alliance/AllianceScreen.tsx | 492 | Статус | `alliance.статус` | ✅ |
| features/alliance/AllianceScreen.tsx | 492 | Альянс | `alliance.альянс` | ✅ |
| features/alliance/AllianceScreen.tsx | 502 | (ожидает) | `alliance.ожидает` | ❌ |
| features/alliance/AllianceScreen.tsx | 502 | Входящее | `alliance.входящее` | ❌ |
| features/alliance/AllianceScreen.tsx | 502 | Предложено | `alliance.предложено` | ❌ |
| features/alliance/AllianceScreen.tsx | 524 | ID альянса | `alliance.id_альянса` | ❌ |
| features/alliance/AllianceScreen.tsx | 530 | НЕН (ненападение) | `alliance.нен_ненападение` | ❌ |
| features/alliance/AllianceScreen.tsx | 531 | СОЮЗ | `alliance.союз` | ❌ |
| features/alliance/AllianceScreen.tsx | 532 | ВОЙНА | `alliance.война` | ❌ |
| features/alliance/AllianceScreen.tsx | 558 | Ошибка | `alliance.ошибка` | ✅ |
| features/alliance/AllianceScreen.tsx | 569 | Игрок | `alliance.игрок` | ✅ |
| features/alliance/AllianceScreen.tsx | 570 | Ранг | `alliance.ранг` | ❌ |
| features/alliance/AllianceScreen.tsx | 571 | Вступил | `alliance.вступил` | ❌ |
| features/artefacts/ArtefactsScreen.tsx | 11 | 📦 В инвентаре | `global.в_инвентаре` | ❌ |
| features/artefacts/ArtefactsScreen.tsx | 12 | ✅ Активен | `global.активен` | ❌ |
| features/artefacts/ArtefactsScreen.tsx | 13 | ⏳ Активируется | `global.активируется` | ❌ |
| features/artefacts/ArtefactsScreen.tsx | 14 | 🏷 На продаже | `global.на_продаже` | ❌ |
| features/artefacts/ArtefactsScreen.tsx | 15 | 💀 Истёк | `global.истёк` | ❌ |
| features/artefacts/ArtefactsScreen.tsx | 16 | ⚡ Использован | `global.использован` | ❌ |
| features/artefacts/ArtefactsScreen.tsx | 44 | Артефакт | `global.артефакт` | ✅ |
| features/artefacts/ArtefactsScreen.tsx | 44 | ${nameOf(a.unit_id)} активирован | `global.nameofaunitid_активирован` | ❌ |
| features/artefacts/ArtefactsScreen.tsx | 48 | Ошибка активации | `global.ошибка_активации` | ❌ |
| features/artefacts/ArtefactsScreen.tsx | 64 | Артефакт деактивирован | `global.артефакт_деактивирован` | ❌ |
| features/artefacts/ArtefactsScreen.tsx | 68 | Ошибка | `global.ошибка` | ✅ |
| features/artefacts/ArtefactsScreen.tsx | 86 | Артефакт выставлен на продажу | `global.артефакт_выставлен_на` | ❌ |
| features/artefacts/ArtefactsScreen.tsx | 90 | Ошибка продажи | `global.ошибка_продажи` | ❌ |
| features/artefacts/ArtefactsScreen.tsx | 123 | Активные | `global.активные` | ❌ |
| features/artefacts/ArtefactsScreen.tsx | 130 | В инвентаре | `global.в_инвентаре` | ❌ |
| features/artefacts/ArtefactsScreen.tsx | 137 | Прочие | `global.прочие` | ❌ |
| features/artefacts/ArtefactsScreen.tsx | 231 | 💀 Истёк | `global.истёк` | ❌ |
| features/artmarket/ArtefactMarketScreen.tsx | 43 | Куплено | `global.куплено` | ❌ |
| features/artmarket/ArtefactMarketScreen.tsx | 43 | Артефакт добавлен в инвентарь | `global.артефакт_добавлен_в` | ❌ |
| features/artmarket/ArtefactMarketScreen.tsx | 45 | Ошибка покупки | `global.ошибка_покупки` | ❌ |
| features/artmarket/ArtefactMarketScreen.tsx | 52 | Оффер отменён | `global.оффер_отменён` | ❌ |
| features/artmarket/ArtefactMarketScreen.tsx | 68 | Баланс: | `global.баланс` | ❌ |
| features/artmarket/ArtefactMarketScreen.tsx | 94 | Артефакт | `global.артефакт` | ✅ |
| features/artmarket/ArtefactMarketScreen.tsx | 95 | Продавец | `global.продавец` | ✅ |
| features/artmarket/ArtefactMarketScreen.tsx | 96 | Цена | `global.цена` | ❌ |
| features/artmarket/ArtefactMarketScreen.tsx | 97 | Дата | `global.дата` | ✅ |
| features/artmarket/ArtefactMarketScreen.tsx | 108 | Артефакт | `global.артефакт` | ✅ |
| features/artmarket/ArtefactMarketScreen.tsx | 112 | Продавец | `global.продавец` | ✅ |
| features/artmarket/ArtefactMarketScreen.tsx | 113 | Цена | `global.цена` | ❌ |
| features/artmarket/ArtefactMarketScreen.tsx | 116 | Дата | `global.дата` | ✅ |
| features/artmarket/ArtefactMarketScreen.tsx | 129 | Недостаточно кредитов | `global.недостаточно_кредитов` | ❌ |
| features/artmarket/ArtefactMarketScreen.tsx | 132 | Купить | `global.купить` | ✅ |
| features/artmarket/ArtefactMarketScreen.tsx | 132 | Мало cr | `global.мало_cr` | ❌ |
| features/auth/LoginScreen.tsx | 78 | Имя пользователя | `auth.имя_пользователя` | ✅ |
| features/auth/LoginScreen.tsx | 86 | от 3 до 24 символов | `auth.от_3_до` | ❌ |
| features/auth/LoginScreen.tsx | 91 | E-Mail или логин | `auth.email_или_логин` | ❌ |
| features/auth/LoginScreen.tsx | 98 | example@mail.com или alice | `auth.examplemailcom_или_alice` | ❌ |
| features/auth/LoginScreen.tsx | 102 | Пароль | `auth.пароль` | ✅ |
| features/auth/LoginScreen.tsx | 110 | минимум 8 символов | `auth.минимум_8_символов` | ❌ |
| features/auth/LoginScreen.tsx | 115 | Войти | `auth.войти` | ✅ |
| features/auth/LoginScreen.tsx | 115 | Зарегистрироваться | `auth.зарегистрироваться` | ❌ |
| features/battle-sim/BattleSimScreen.tsx | 112 | Атакующий флот | `global.атакующий_флот` | ❌ |
| features/battle-sim/BattleSimScreen.tsx | 118 | Защита + флот | `global.защита_флот` | ❌ |
| features/battle-sim/BattleSimScreen.tsx | 136 | Прогонов (1–20): | `global.прогонов_120` | ❌ |
| features/battle-sim/BattleSimScreen.tsx | 150 | Считаем… | `global.считаем` | ❌ |
| features/battle-sim/BattleSimScreen.tsx | 151 | Рассчитать | `global.рассчитать` | ❌ |
| features/battle-sim/BattleSimScreen.tsx | 162 | Результат | `global.результат` | ❌ |
| features/battle-sim/BattleSimScreen.tsx | 164 | Прогонов | `global.прогонов` | ❌ |
| features/battle-sim/BattleSimScreen.tsx | 166 | Победы атак. | `global.победы_атак` | ❌ |
| features/battle-sim/BattleSimScreen.tsx | 169 | Ничьи | `global.ничьи` | ❌ |
| features/battle-sim/BattleSimScreen.tsx | 172 | Ср. раундов | `global.ср_раундов` | ❌ |
| features/battle-sim/BattleSimScreen.tsx | 180 | Результат | `global.результат` | ❌ |
| features/battle-sim/BattleSimScreen.tsx | 182 | Победитель | `global.победитель` | ❌ |
| features/battle-sim/BattleSimScreen.tsx | 184 | Атакующие | `global.атакующие` | ❌ |
| features/battle-sim/BattleSimScreen.tsx | 186 | Защитники | `global.защитники` | ❌ |
| features/battle-sim/BattleSimScreen.tsx | 187 | Ничья | `global.ничья` | ✅ |
| features/battle-sim/BattleSimScreen.tsx | 189 | Раундов | `global.раундов` | ✅ |
| features/battle-sim/BattleSimScreen.tsx | 191 | Сид | `global.сид` | ❌ |
| features/battle-sim/BattleSimScreen.tsx | 195 | Обломки | `global.обломки` | ✅ |
| features/battle-sim/BattleSimScreen.tsx | 203 | Атакующие | `global.атакующие` | ❌ |
| features/battle-sim/BattleSimScreen.tsx | 210 | Защитники | `global.защитники` | ❌ |
| features/battle-sim/BattleSimScreen.tsx | 235 | Потери | `global.потери` | ❌ |
| features/battle-sim/BattleSimScreen.tsx | 242 | Юнит | `global.юнит` | ❌ |
| features/battle-sim/BattleSimScreen.tsx | 243 | Было | `global.было` | ❌ |
| features/battle-sim/BattleSimScreen.tsx | 244 | Стало | `global.стало` | ❌ |
| features/battle-sim/BattleSimScreen.tsx | 245 | Повреждено | `global.повреждено` | ❌ |
| features/battle-sim/BattleSimScreen.tsx | 266 | Потерь нет. | `global.потерь_нет` | ❌ |
| features/battlestats/BattlestatsScreen.tsx | 54 | ⚔ История боёв | `global.история_боёв` | ❌ |
| features/battlestats/BattlestatsScreen.tsx | 58 | Всего боёв | `global.всего_боёв` | ❌ |
| features/battlestats/BattlestatsScreen.tsx | 59 | Побед | `global.побед` | ❌ |
| features/battlestats/BattlestatsScreen.tsx | 60 | Поражений | `global.поражений` | ❌ |
| features/battlestats/BattlestatsScreen.tsx | 61 | Ничьих | `global.ничьих` | ❌ |
| features/battlestats/BattlestatsScreen.tsx | 67 | Роль | `global.роль` | ❌ |
| features/battlestats/BattlestatsScreen.tsx | 69 | Любая | `global.любая` | ❌ |
| features/battlestats/BattlestatsScreen.tsx | 70 | Атакующий | `global.атакующий` | ✅ |
| features/battlestats/BattlestatsScreen.tsx | 71 | Защитник | `global.защитник` | ❌ |
| features/battlestats/BattlestatsScreen.tsx | 75 | Результат | `global.результат` | ❌ |
| features/battlestats/BattlestatsScreen.tsx | 77 | Любой | `global.любой` | ❌ |
| features/battlestats/BattlestatsScreen.tsx | 78 | Победа | `global.победа` | ❌ |
| features/battlestats/BattlestatsScreen.tsx | 79 | Поражение | `global.поражение` | ❌ |
| features/battlestats/BattlestatsScreen.tsx | 80 | Ничья | `global.ничья` | ✅ |
| features/battlestats/BattlestatsScreen.tsx | 84 | С даты | `global.с_даты` | ❌ |
| features/battlestats/BattlestatsScreen.tsx | 88 | По дату | `global.по_дату` | ❌ |
| features/battlestats/BattlestatsScreen.tsx | 108 | Дата | `global.дата` | ✅ |
| features/battlestats/BattlestatsScreen.tsx | 109 | Роль | `global.роль` | ❌ |
| features/battlestats/BattlestatsScreen.tsx | 110 | Противник | `global.противник` | ❌ |
| features/battlestats/BattlestatsScreen.tsx | 111 | Планета | `global.планета` | ✅ |
| features/battlestats/BattlestatsScreen.tsx | 112 | Раундов | `global.раундов` | ✅ |
| features/battlestats/BattlestatsScreen.tsx | 113 | Результат | `global.результат` | ❌ |
| features/battlestats/BattlestatsScreen.tsx | 114 | Добыча | `global.добыча` | ❌ |
| features/battlestats/BattlestatsScreen.tsx | 115 | Обломки | `global.обломки` | ✅ |
| features/battlestats/BattlestatsScreen.tsx | 123 | ⚖ Ничья | `global.ничья` | ❌ |
| features/battlestats/BattlestatsScreen.tsx | 123 | 🏆 Победа | `global.победа` | ❌ |
| features/battlestats/BattlestatsScreen.tsx | 123 | 💀 Поражение | `global.поражение` | ❌ |
| features/battlestats/BattlestatsScreen.tsx | 131 | ⚔ Атака | `global.атака` | ❌ |
| features/battlestats/BattlestatsScreen.tsx | 131 | 🛡 Защита | `global.защита` | ❌ |
| features/buildings/BuildingsScreen.tsx | 6 | ${secs}с | `buildings.secsс` | ❌ |
| features/buildings/BuildingsScreen.tsx | 10 | ${d}д ${h}ч ${m}м | `buildings.dд_hч_mм` | ❌ |
| features/buildings/BuildingsScreen.tsx | 11 | ${h}ч ${m}м | `buildings.hч_mм` | ❌ |
| features/buildings/BuildingsScreen.tsx | 12 | ${m}м | `buildings.mм` | ❌ |
| features/buildings/BuildingsScreen.tsx | 23 | ${(h / 1_000_000).toFixed(1)}M/ч | `buildings.h_1000000tofixed1mч` | ❌ |
| features/buildings/BuildingsScreen.tsx | 24 | ${Math.round(h / 1_000)}k/ч | `buildings.mathroundh_1000kч` | ❌ |
| features/buildings/BuildingsScreen.tsx | 25 | ${h}/ч | `buildings.hч` | ❌ |
| features/buildings/BuildingsScreen.tsx | 39 | Добыча | `buildings.добыча` | ❌ |
| features/buildings/BuildingsScreen.tsx | 40 | Добыча | `buildings.добыча` | ❌ |
| features/buildings/BuildingsScreen.tsx | 41 | Добыча | `buildings.добыча` | ❌ |
| features/buildings/BuildingsScreen.tsx | 42 | Энергия | `buildings.энергия` | ✅ |
| features/buildings/BuildingsScreen.tsx | 43 | Энергия | `buildings.энергия` | ✅ |
| features/buildings/BuildingsScreen.tsx | 75 | В очередь | `buildings.в_очередь` | ✅ |
| features/buildings/BuildingsScreen.tsx | 75 | ${name} добавлена в очередь строительства | `buildings.name_добавлена_в` | ❌ |
| features/buildings/BuildingsScreen.tsx | 79 | Очередь занята, дождитесь завершения постройки | `buildings.очередь_занята_дождитесь` | ❌ |
| features/buildings/BuildingsScreen.tsx | 80 | Недостаточно ресурсов | `buildings.недостаточно_ресурсов` | ❌ |
| features/buildings/BuildingsScreen.tsx | 81 | Это здание доступно только на луне | `buildings.это_здание_доступно` | ❌ |
| features/buildings/BuildingsScreen.tsx | 82 | Это здание недоступно на луне | `buildings.это_здание_недоступно` | ❌ |
| features/buildings/BuildingsScreen.tsx | 83 | Не удалось добавить в очередь | `buildings.не_удалось_добавить` | ❌ |
| features/buildings/BuildingsScreen.tsx | 84 | Ошибка | `buildings.ошибка` | ✅ |
| features/buildings/BuildingsScreen.tsx | 94 | Отменено | `buildings.отменено` | ❌ |
| features/buildings/BuildingsScreen.tsx | 94 | Строительство отменено, ресурсы возвращены | `buildings.строительство_отменено_ресурсы` | ❌ |
| features/buildings/BuildingsScreen.tsx | 97 | Ошибка | `buildings.ошибка` | ✅ |
| features/buildings/BuildingsScreen.tsx | 97 | Не удалось отменить | `buildings.не_удалось_отменить` | ❌ |
| features/buildings/BuildingsScreen.tsx | 131 | Занято / максимум полей на планете | `buildings.занято_максимум_полей` | ❌ |
| features/buildings/BuildingsScreen.tsx | 161 | 👁 Все здания | `buildings.все_здания` | ❌ |
| features/buildings/BuildingsScreen.tsx | 161 | 🔒 Только доступные | `buildings.только_доступные` | ❌ |
| features/buildings/BuildingsScreen.tsx | 205 | Подробнее | `buildings.подробнее` | ❌ |
| features/buildings/BuildingsScreen.tsx | 215 | Не построено | `buildings.не_построено` | ❌ |
| features/buildings/BuildingsScreen.tsx | 215 | Уровень ${level} | `buildings.уровень_level` | ❌ |
| features/buildings/BuildingsScreen.tsx | 287 | → ур. ${level + 1} | `buildings.ур_level_1` | ❌ |
| features/buildings/BuildingsScreen.tsx | 287 | Построить | `buildings.построить` | ✅ |
| features/buildings/BuildingsScreen.tsx | 287 | 🔒 Заблокировано | `buildings.заблокировано` | ❌ |
| features/buildings/BuildingsScreen.tsx | 287 | ⏳ В очереди | `buildings.в_очереди` | ❌ |
| features/buildings/BuildingsScreen.tsx | 358 | Отменить? | `buildings.отменить` | ❌ |
| features/buildings/BuildingsScreen.tsx | 383 | Отменить (ресурсы вернутся) | `buildings.отменить_ресурсы_вернутся` | ❌ |
| features/chat/ChatScreen.tsx | 93 | Переподключение… | `chat.переподключение` | ❌ |
| features/chat/ChatScreen.tsx | 245 | Редактировать | `chat.редактировать` | ✅ |
| features/chat/ChatScreen.tsx | 255 | Удалить | `chat.удалить` | ✅ |
| features/chat/ChatScreen.tsx | 307 | изм. | `chat.изм` | ❌ |
| features/chat/ChatScreen.tsx | 337 | Удалить сообщение? | `chat.удалить_сообщение` | ❌ |
| features/chat/ChatScreen.tsx | 338 | Сообщение будет удалено безвозвратно. | `chat.сообщение_будет_удалено` | ❌ |
| features/chat/ChatScreen.tsx | 339 | Удалить | `chat.удалить` | ✅ |
| features/chat/ChatScreen.tsx | 352 | Смайлики | `chat.смайлики` | ❌ |
| features/chat/ChatScreen.tsx | 359 | Сообщение… (Enter — отправить) | `chat.сообщение_enter_отправить` | ❌ |
| features/dailyquest/DailyQuestScreen.tsx | 41 | +${res.reward_credits} кр | `global.resrewardcredits_кр` | ❌ |
| features/dailyquest/DailyQuestScreen.tsx | 45 | Награда получена | `global.награда_получена` | ❌ |
| features/dailyquest/DailyQuestScreen.tsx | 48 | Не удалось получить награду | `global.не_удалось_получить` | ❌ |
| features/dailyquest/DailyQuestScreen.tsx | 48 | Ошибка | `global.ошибка` | ✅ |
| features/dailyquest/DailyQuestScreen.tsx | 56 | Ежедневные задания | `global.ежедневные_задания` | ❌ |
| features/dailyquest/DailyQuestScreen.tsx | 61 | Загрузка… | `global.загрузка` | ❌ |
| features/dailyquest/DailyQuestScreen.tsx | 63 | Заданий пока нет. | `global.заданий_пока_нет` | ❌ |
| features/dailyquest/DailyQuestScreen.tsx | 69 | ${q.reward_credits} кр | `global.qrewardcredits_кр` | ❌ |
| features/dailyquest/DailyQuestScreen.tsx | 79 | Получено | `global.получено` | ❌ |
| features/empire/EmpireScreen.tsx | 27 | Добыча | `global.добыча` | ❌ |
| features/empire/EmpireScreen.tsx | 28 | Энергия | `global.энергия` | ✅ |
| features/empire/EmpireScreen.tsx | 29 | Хранилища | `global.хранилища` | ❌ |
| features/empire/EmpireScreen.tsx | 30 | Производство | `global.производство` | ✅ |
| features/empire/EmpireScreen.tsx | 31 | Особые | `global.особые` | ❌ |
| features/empire/EmpireScreen.tsx | 124 | Нет планет. | `global.нет_планет` | ❌ |
| features/empire/EmpireScreen.tsx | 176 | Империя | `global.империя` | ✅ |
| features/empire/EmpireScreen.tsx | 180 | Всего ресурсов: | `global.всего_ресурсов` | ❌ |
| features/empire/EmpireScreen.tsx | 192 | Параметр | `global.параметр` | ❌ |
| features/empire/EmpireScreen.tsx | 206 | Планета | `global.планета` | ✅ |
| features/empire/EmpireScreen.tsx | 209 | 📐 Диаметр | `global.диаметр` | ❌ |
| features/empire/EmpireScreen.tsx | 217 | 🔲 Поля | `global.поля` | ❌ |
| features/empire/EmpireScreen.tsx | 225 | 🌡️ Температура | `global.температура` | ❌ |
| features/empire/EmpireScreen.tsx | 235 | Ресурсы | `global.ресурсы` | ✅ |
| features/empire/EmpireScreen.tsx | 237 | 🟠 Металл | `global.металл` | ❌ |
| features/empire/EmpireScreen.tsx | 237 | 💎 Кремний | `global.кремний` | ❌ |
| features/empire/EmpireScreen.tsx | 237 | 💧 Водород | `global.водород` | ❌ |
| features/empire/EmpireScreen.tsx | 268 | Прочие постройки | `global.прочие_постройки` | ❌ |
| features/empire/EmpireScreen.tsx | 276 | Флот | `global.флот` | ✅ |
| features/empire/EmpireScreen.tsx | 287 | Оборона | `global.оборона` | ✅ |
| features/fleet/FleetScreen.tsx | 29 | ${Math.ceil(secs)}с | `fleet.mathceilsecsс` | ❌ |
| features/fleet/FleetScreen.tsx | 33 | ${d}д ${h}ч ${m}м | `fleet.dд_hч_mм` | ❌ |
| features/fleet/FleetScreen.tsx | 34 | ${h}ч ${m}м | `fleet.hч_mм` | ❌ |
| features/fleet/FleetScreen.tsx | 35 | ${m}м | `fleet.mм` | ❌ |
| features/fleet/FleetScreen.tsx | 55 | Перебазирование | `fleet.перебазирование` | ❌ |
| features/fleet/FleetScreen.tsx | 56 | Транспорт | `fleet.транспорт` | ✅ |
| features/fleet/FleetScreen.tsx | 57 | Колонизация | `fleet.колонизация` | ✅ |
| features/fleet/FleetScreen.tsx | 58 | Переработка | `fleet.переработка` | ✅ |
| features/fleet/FleetScreen.tsx | 59 | Атака | `fleet.атака` | ✅ |
| features/fleet/FleetScreen.tsx | 60 | Шпионаж | `fleet.шпионаж` | ✅ |
| features/fleet/FleetScreen.tsx | 61 | Экспедиция | `fleet.экспедиция` | ✅ |
| features/fleet/FleetScreen.tsx | 65 | → В пути | `fleet.в_пути` | ❌ |
| features/fleet/FleetScreen.tsx | 66 | ← Возврат | `fleet.возврат` | ❌ |
| features/fleet/FleetScreen.tsx | 67 | ✓ Прибыл | `fleet.прибыл` | ❌ |
| features/fleet/FleetScreen.tsx | 114 | ${MISSION_LABELS[mission] ?? 'Миссия'} → [${g}:${s}:${pos}] | `fleet.missionlabelsmission_миссия_gspos` | ❌ |
| features/fleet/FleetScreen.tsx | 114 | Флот отправлен | `fleet.флот_отправлен` | ❌ |
| features/fleet/FleetScreen.tsx | 114 | Миссия | `fleet.миссия` | ❌ |
| features/fleet/FleetScreen.tsx | 117 | Не удалось отправить | `fleet.не_удалось_отправить` | ❌ |
| features/fleet/FleetScreen.tsx | 117 | Ошибка | `fleet.ошибка` | ✅ |
| features/fleet/FleetScreen.tsx | 125 | Флот отозван | `fleet.флот_отозван` | ❌ |
| features/fleet/FleetScreen.tsx | 128 | Ошибка | `fleet.ошибка` | ✅ |
| features/fleet/FleetScreen.tsx | 128 | Не удалось отозвать | `fleet.не_удалось_отозвать` | ❌ |
| features/fleet/FleetScreen.tsx | 170 | (увеличивается с computer_tech) | `fleet.увеличивается_с_computertech` | ❌ |
| features/fleet/FleetScreen.tsx | 184 | Миссия | `fleet.миссия` | ❌ |
| features/fleet/FleetScreen.tsx | 185 | Назначение | `fleet.назначение` | ❌ |
| features/fleet/FleetScreen.tsx | 186 | Состав | `fleet.состав` | ❌ |
| features/fleet/FleetScreen.tsx | 187 | Статус | `fleet.статус` | ✅ |
| features/fleet/FleetScreen.tsx | 188 | Прилёт / Возврат | `fleet.прилёт_возврат` | ❌ |
| features/fleet/FleetScreen.tsx | 253 | Миссия | `fleet.миссия` | ❌ |
| features/fleet/FleetScreen.tsx | 262 | Координаты назначения | `fleet.координаты_назначения` | ❌ |
| features/fleet/FleetScreen.tsx | 279 | Название колонии | `fleet.название_колонии` | ❌ |
| features/fleet/FleetScreen.tsx | 295 | Металл | `fleet.металл` | ✅ |
| features/fleet/FleetScreen.tsx | 296 | Кремний | `fleet.кремний` | ✅ |
| features/fleet/FleetScreen.tsx | 297 | Водород | `fleet.водород` | ✅ |
| features/fleet/FleetScreen.tsx | 348 | 🚀 Отправить флот | `fleet.отправить_флот` | ❌ |
| features/friends/FriendsScreen.tsx | 15 | онлайн | `global.онлайн` | ❌ |
| features/friends/FriendsScreen.tsx | 16 | ${mins} мин назад | `global.mins_мин_назад` | ❌ |
| features/friends/FriendsScreen.tsx | 18 | ${hrs} ч назад | `global.hrs_ч_назад` | ❌ |
| features/friends/FriendsScreen.tsx | 20 | ${days} дн назад | `global.days_дн_назад` | ❌ |
| features/friends/FriendsScreen.tsx | 40 | ⭐ Друзья | `global.друзья` | ❌ |
| features/friends/FriendsScreen.tsx | 57 | Игрок | `global.игрок` | ✅ |
| features/friends/FriendsScreen.tsx | 58 | Альянс | `global.альянс` | ✅ |
| features/friends/FriendsScreen.tsx | 59 | Очки | `global.очки` | ✅ |
| features/friends/FriendsScreen.tsx | 60 | Активность | `global.активность` | ❌ |
| features/galaxy/GalaxyScreen.tsx | 59 | ${Math.ceil(secs)}с | `galaxy.mathceilsecsс` | ❌ |
| features/galaxy/GalaxyScreen.tsx | 63 | ${d}д ${h}ч ${m}м | `galaxy.dд_hч_mм` | ❌ |
| features/galaxy/GalaxyScreen.tsx | 64 | ${h}ч ${m}м | `galaxy.hч_mм` | ❌ |
| features/galaxy/GalaxyScreen.tsx | 65 | ${m}м | `galaxy.mм` | ❌ |
| features/galaxy/GalaxyScreen.tsx | 93 | Забанен | `galaxy.забанен` | ✅ |
| features/galaxy/GalaxyScreen.tsx | 95 | Режим отпуска | `galaxy.режим_отпуска` | ✅ |
| features/galaxy/GalaxyScreen.tsx | 98 | Очень неактивный (21+ дн) | `galaxy.очень_неактивный_21` | ❌ |
| features/galaxy/GalaxyScreen.tsx | 100 | Неактивный (7+ дн) | `galaxy.неактивный_7_дн` | ❌ |
| features/galaxy/GalaxyScreen.tsx | 143 | Ракеты запущены | `galaxy.ракеты_запущены` | ❌ |
| features/galaxy/GalaxyScreen.tsx | 164 | Источник | `galaxy.источник` | ❌ |
| features/galaxy/GalaxyScreen.tsx | 175 | Количество | `galaxy.количество` | ✅ |
| features/galaxy/GalaxyScreen.tsx | 194 | Запустить | `galaxy.запустить` | ❌ |
| features/galaxy/GalaxyScreen.tsx | 195 | Отмена | `galaxy.отмена` | ✅ |
| features/galaxy/GalaxyScreen.tsx | 226 | Расстояние: ${dist}\nВремя (мин. скорость): ${flightTime} | `galaxy.расстояние_distnвремя_мин` | ❌ |
| features/galaxy/GalaxyScreen.tsx | 235 | Шпионаж\n${fuelHint} | `galaxy.шпионажnfuelhint` | ❌ |
| features/galaxy/GalaxyScreen.tsx | 242 | Атака\n${fuelHint} | `galaxy.атакаnfuelhint` | ❌ |
| features/galaxy/GalaxyScreen.tsx | 249 | Транспорт\n${fuelHint} | `galaxy.транспортnfuelhint` | ❌ |
| features/galaxy/GalaxyScreen.tsx | 257 | Переработка обломков\n${fuelHint} | `galaxy.переработка_обломковnfuelhint` | ❌ |
| features/galaxy/GalaxyScreen.tsx | 266 | Шпионаж на луну\n${fuelHint} | `galaxy.шпионаж_на_лунуnfuelhint` | ❌ |
| features/galaxy/GalaxyScreen.tsx | 274 | Ракетный удар | `galaxy.ракетный_удар` | ❌ |
| features/galaxy/GalaxyScreen.tsx | 344 | Система добавлена в наблюдение | `galaxy.система_добавлена_в` | ❌ |
| features/galaxy/GalaxyScreen.tsx | 344 | Система удалена из наблюдения | `galaxy.система_удалена_из` | ❌ |
| features/galaxy/GalaxyScreen.tsx | 393 | Убрать из наблюдения | `galaxy.убрать_из_наблюдения` | ❌ |
| features/galaxy/GalaxyScreen.tsx | 393 | Добавить в наблюдение | `galaxy.добавить_в_наблюдение` | ❌ |
| features/galaxy/GalaxyScreen.tsx | 410 | Наблюдение: | `galaxy.наблюдение` | ❌ |
| features/galaxy/GalaxyScreen.tsx | 436 | неизвестная ошибка | `galaxy.неизвестная_ошибка` | ❌ |
| features/galaxy/GalaxyScreen.tsx | 447 | Планета | `galaxy.планета` | ✅ |
| features/galaxy/GalaxyScreen.tsx | 448 | Игрок | `galaxy.игрок` | ✅ |
| features/galaxy/GalaxyScreen.tsx | 449 | Альянс | `galaxy.альянс` | ✅ |
| features/galaxy/GalaxyScreen.tsx | 450 | Обломки | `galaxy.обломки` | ✅ |
| features/galaxy/GalaxyScreen.tsx | 451 | Миссии | `galaxy.миссии` | ❌ |
| features/galaxy/GalaxyScreen.tsx | 459 | Луна | `galaxy.луна` | ✅ |
| features/galaxy/GalaxyScreen.tsx | 460 | ${c.moon_diameter} км | `galaxy.cmoondiameter_км` | ❌ |
| features/galaxy/GalaxyScreen.tsx | 467 | Обломки\nМеталл: ${c.debris_metal.toLocaleString('ru-RU')… | `galaxy.обломкиnметалл_cdebrismetaltolocalestringrurunкремний_cdebrissilicontolocalestringruru` | ❌ |
| features/galaxy/GalaxyScreen.tsx | 492 | Планета | `galaxy.планета` | ✅ |
| features/galaxy/GalaxyScreen.tsx | 494 | 🌌 Бесконечные дали | `galaxy.бесконечные_дали` | ❌ |
| features/galaxy/GalaxyScreen.tsx | 514 | Игрок | `galaxy.игрок` | ✅ |
| features/galaxy/GalaxyScreen.tsx | 517 | В друзьях | `galaxy.в_друзьях` | ❌ |
| features/galaxy/GalaxyScreen.tsx | 536 | Отправить экспедицию | `galaxy.отправить_экспедицию` | ❌ |
| features/galaxy/GalaxyScreen.tsx | 538 | 🌌 Экспедиция | `galaxy.экспедиция` | ❌ |
| features/galaxy/GalaxyScreen.tsx | 541 | Альянс | `galaxy.альянс` | ✅ |
| features/galaxy/GalaxyScreen.tsx | 547 | Обломки | `galaxy.обломки` | ✅ |
| features/galaxy/GalaxyScreen.tsx | 557 | Миссии | `galaxy.миссии` | ❌ |
| features/galaxy/GalaxyScreen.tsx | 563 | Отправить экспедицию в Бесконечные дали | `galaxy.отправить_экспедицию_в` | ❌ |
| features/galaxy/GalaxyScreen.tsx | 565 | 🌌 Экспедиция | `galaxy.экспедиция` | ❌ |
| features/galaxy/GalaxyScreen.tsx | 582 | Отправить экспедицию с этой планеты | `galaxy.отправить_экспедицию_с` | ❌ |
| features/galaxyevent/GalaxyEventBanner.tsx | 13 | Метеоритный шторм | `global.метеоритный_шторм` | ❌ |
| features/galaxyevent/GalaxyEventBanner.tsx | 13 | +30% к добыче металла | `global.30_к_добыче` | ❌ |
| features/galaxyevent/GalaxyEventBanner.tsx | 14 | Солнечная вспышка | `global.солнечная_вспышка` | ❌ |
| features/galaxyevent/GalaxyEventBanner.tsx | 14 | −20% энергии | `global.20_энергии` | ❌ |
| features/galaxyevent/GalaxyEventBanner.tsx | 15 | Торговый форум | `global.торговый_форум` | ❌ |
| features/galaxyevent/GalaxyEventBanner.tsx | 15 | Льготные курсы рынка | `global.льготные_курсы_рынка` | ❌ |
| features/galaxyevent/GalaxyEventBanner.tsx | 16 | Звёздная туманность | `global.звёздная_туманность` | ❌ |
| features/galaxyevent/GalaxyEventBanner.tsx | 16 | +15% к мощи экспедиций | `global.15_к_мощи` | ❌ |
| features/galaxyevent/GalaxyEventBanner.tsx | 24 | ${h}ч ${m}м | `global.hч_mм` | ❌ |
| features/galaxyevent/GalaxyEventBanner.tsx | 25 | ${m}м | `global.mм` | ❌ |
| features/galaxyevent/GalaxyEventBanner.tsx | 41 | Действует особый эффект | `global.действует_особый_эффект` | ❌ |
| features/market/MarketScreen.tsx | 38 | 🟠 Металл | `market.металл` | ❌ |
| features/market/MarketScreen.tsx | 38 | 💎 Кремний | `market.кремний` | ❌ |
| features/market/MarketScreen.tsx | 38 | 💧 Водород | `market.водород` | ❌ |
| features/market/MarketScreen.tsx | 78 | Обмен | `market.обмен` | ❌ |
| features/market/MarketScreen.tsx | 84 | Ошибка обмена | `market.ошибка_обмена` | ❌ |
| features/market/MarketScreen.tsx | 136 | Отдать | `market.отдать` | ❌ |
| features/market/MarketScreen.tsx | 144 | Получить | `market.получить` | ❌ |
| features/market/MarketScreen.tsx | 152 | Количество | `market.количество` | ✅ |
| features/market/MarketScreen.tsx | 178 | Обменять | `market.обменять` | ❌ |
| features/market/MarketScreen.tsx | 215 | Лот выставлен | `market.лот_выставлен` | ❌ |
| features/market/MarketScreen.tsx | 216 | Ошибка | `market.ошибка` | ✅ |
| features/market/MarketScreen.tsx | 221 | Лот отменён | `market.лот_отменён` | ❌ |
| features/market/MarketScreen.tsx | 226 | Сделка выполнена | `market.сделка_выполнена` | ❌ |
| features/market/MarketScreen.tsx | 227 | Ошибка | `market.ошибка` | ✅ |
| features/market/MarketScreen.tsx | 234 | 💱 Ресурсы | `market.ресурсы` | ❌ |
| features/market/MarketScreen.tsx | 235 | 🛸 Флот | `market.флот` | ❌ |
| features/market/MarketScreen.tsx | 245 | 💱 Ресурсы | `market.ресурсы` | ❌ |
| features/market/MarketScreen.tsx | 246 | 🛸 Флот | `market.флот` | ❌ |
| features/market/MarketScreen.tsx | 255 | Продать | `market.продать` | ❌ |
| features/market/MarketScreen.tsx | 265 | Получить | `market.получить` | ❌ |
| features/market/MarketScreen.tsx | 274 | Выставить | `market.выставить` | ❌ |
| features/market/MarketScreen.tsx | 285 | Продавец | `market.продавец` | ✅ |
| features/market/MarketScreen.tsx | 286 | Продаёт | `market.продаёт` | ❌ |
| features/market/MarketScreen.tsx | 287 | Хочет | `market.хочет` | ❌ |
| features/market/MarketScreen.tsx | 294 | Продавец | `market.продавец` | ✅ |
| features/market/MarketScreen.tsx | 295 | Продаёт | `market.продаёт` | ❌ |
| features/market/MarketScreen.tsx | 296 | Хочет | `market.хочет` | ❌ |
| features/market/MarketScreen.tsx | 299 | Отмена | `market.отмена` | ✅ |
| features/market/MarketScreen.tsx | 301 | Купить | `market.купить` | ✅ |
| features/market/MarketScreen.tsx | 307 | Нет открытых лотов | `market.нет_открытых_лотов` | ❌ |
| features/market/MarketScreen.tsx | 345 | Обмен | `market.обмен` | ❌ |
| features/market/MarketScreen.tsx | 345 | ${amount} кред. → ${res.resource_delta} ${res.resource} | `market.amount_кред_resresourcedelta` | ❌ |
| features/market/MarketScreen.tsx | 347 | Ошибка | `market.ошибка` | ✅ |
| features/market/MarketScreen.tsx | 359 | Ресурс | `market.ресурс` | ✅ |
| features/market/MarketScreen.tsx | 389 | Купить | `market.купить` | ✅ |
| features/market/MarketScreen.tsx | 430 | Лот флота выставлен | `market.лот_флота_выставлен` | ❌ |
| features/market/MarketScreen.tsx | 432 | Ошибка | `market.ошибка` | ✅ |
| features/market/MarketScreen.tsx | 437 | Лот отменён | `market.лот_отменён` | ❌ |
| features/market/MarketScreen.tsx | 445 | Сделка выполнена | `market.сделка_выполнена` | ❌ |
| features/market/MarketScreen.tsx | 447 | Ошибка | `market.ошибка` | ✅ |
| features/market/MarketScreen.tsx | 478 | Хочу получить | `market.хочу_получить` | ❌ |
| features/market/MarketScreen.tsx | 500 | Выставить (${totalShips} кораблей) | `market.выставить_totalships_кораблей` | ❌ |
| features/market/MarketScreen.tsx | 511 | Продавец | `market.продавец` | ✅ |
| features/market/MarketScreen.tsx | 512 | Состав | `market.состав` | ❌ |
| features/market/MarketScreen.tsx | 513 | Цена | `market.цена` | ❌ |
| features/market/MarketScreen.tsx | 522 | (вы) | `market.вы` | ❌ |
| features/market/MarketScreen.tsx | 533 | Отменить | `market.отменить` | ❌ |
| features/market/MarketScreen.tsx | 535 | Купить | `market.купить` | ✅ |
| features/messages/MessagesScreen.tsx | 112 | Все | `global.все` | ✅ |
| features/messages/MessagesScreen.tsx | 113 | Личные | `global.личные` | ❌ |
| features/messages/MessagesScreen.tsx | 114 | Бой | `global.бой` | ✅ |
| features/messages/MessagesScreen.tsx | 115 | Шпионаж | `global.шпионаж` | ✅ |
| features/messages/MessagesScreen.tsx | 116 | Экспедиции | `global.экспедиции` | ❌ |
| features/messages/MessagesScreen.tsx | 117 | Фаланга | `global.фаланга` | ❌ |
| features/messages/MessagesScreen.tsx | 118 | Альянс | `global.альянс` | ✅ |
| features/messages/MessagesScreen.tsx | 119 | Артефакты | `global.артефакты` | ✅ |
| features/messages/MessagesScreen.tsx | 120 | Кредиты | `global.кредиты` | ✅ |
| features/messages/MessagesScreen.tsx | 121 | Система | `global.система` | ✅ |
| features/messages/MessagesScreen.tsx | 122 | Отправленные | `global.отправленные` | ❌ |
| features/messages/MessagesScreen.tsx | 161 | Сообщение удалено | `global.сообщение_удалено` | ❌ |
| features/messages/MessagesScreen.tsx | 174 | Все сообщения удалены | `global.все_сообщения_удалены` | ❌ |
| features/messages/MessagesScreen.tsx | 259 | Отправлено | `global.отправлено` | ❌ |
| features/messages/MessagesScreen.tsx | 266 | 📭 Нет сообщений | `global.нет_сообщений` | ❌ |
| features/messages/MessagesScreen.tsx | 294 | Система | `global.система` | ✅ |
| features/messages/MessagesScreen.tsx | 304 | Удалить | `global.удалить` | ✅ |
| features/messages/MessagesScreen.tsx | 324 | Удалить сообщения | `global.удалить_сообщения` | ❌ |
| features/messages/MessagesScreen.tsx | 325 | Удалить все сообщения? | `global.удалить_все_сообщения` | ❌ |
| features/messages/MessagesScreen.tsx | 325 | Удалить все сообщения в этой папке? | `global.удалить_все_сообщения` | ❌ |
| features/messages/MessagesScreen.tsx | 326 | Удалить | `global.удалить` | ✅ |
| features/messages/MessagesScreen.tsx | 345 | Ошибка | `global.ошибка` | ✅ |
| features/messages/MessagesScreen.tsx | 350 | Написать сообщение | `global.написать_сообщение` | ✅ |
| features/messages/MessagesScreen.tsx | 352 | Кому | `global.кому` | ❌ |
| features/messages/MessagesScreen.tsx | 353 | имя игрока | `global.имя_игрока` | ❌ |
| features/messages/MessagesScreen.tsx | 356 | Тема | `global.тема` | ✅ |
| features/messages/MessagesScreen.tsx | 357 | тема сообщения | `global.тема_сообщения` | ❌ |
| features/messages/MessagesScreen.tsx | 360 | текст сообщения… | `global.текст_сообщения` | ❌ |
| features/messages/MessagesScreen.tsx | 365 | Отправить | `global.отправить` | ✅ |
| features/messages/MessagesScreen.tsx | 367 | Отмена | `global.отмена` | ✅ |
| features/messages/MessagesScreen.tsx | 394 | Система | `global.система` | ✅ |
| features/messages/MessagesScreen.tsx | 409 | Загрузка отчёта… | `global.загрузка_отчёта` | ❌ |
| features/messages/MessagesScreen.tsx | 415 | Загрузка… | `global.загрузка` | ❌ |
| features/messages/MessagesScreen.tsx | 421 | Загрузка… | `global.загрузка` | ❌ |
| features/messages/MessagesScreen.tsx | 431 | Найдены ресурсы | `global.найдены_ресурсы` | ❌ |
| features/messages/MessagesScreen.tsx | 431 | Найден артефакт | `global.найден_артефакт` | ❌ |
| features/messages/MessagesScreen.tsx | 432 | Обнаружена новая планета | `global.обнаружена_новая_планета` | ❌ |
| features/messages/MessagesScreen.tsx | 432 | Столкновение с пиратами | `global.столкновение_с_пиратами` | ❌ |
| features/messages/MessagesScreen.tsx | 433 | Потери | `global.потери` | ❌ |
| features/messages/MessagesScreen.tsx | 433 | Ничего не найдено | `global.ничего_не_найдено` | ❌ |
| features/messages/MessagesScreen.tsx | 437 | 🌌 Отчёт экспедиции | `global.отчёт_экспедиции` | ❌ |
| features/messages/MessagesScreen.tsx | 438 | Результат: | `global.результат` | ✅ |
| features/messages/MessagesScreen.tsx | 450 | 🔭 Шпионский отчёт | `global.шпионский_отчёт` | ❌ |
| features/messages/MessagesScreen.tsx | 455 | Корабли | `global.корабли` | ❌ |
| features/messages/MessagesScreen.tsx | 456 | Оборона | `global.оборона` | ✅ |
| features/messages/MessagesScreen.tsx | 457 | Здания | `global.здания` | ✅ |
| features/messages/MessagesScreen.tsx | 487 | ⚔️ Победа атакующих | `global.победа_атакующих` | ❌ |
| features/messages/MessagesScreen.tsx | 488 | 🛡 Победа защитников | `global.победа_защитников` | ❌ |
| features/messages/MessagesScreen.tsx | 489 | ⚖️ Ничья | `global.ничья` | ❌ |
| features/messages/MessagesScreen.tsx | 517 | АТАКУЮЩИЙ | `global.атакующий` | ❌ |
| features/messages/MessagesScreen.tsx | 521 | ЗАЩИТНИК | `global.защитник` | ❌ |
| features/messages/MessagesScreen.tsx | 529 | ДОБЫЧА | `global.добыча` | ❌ |
| features/messages/MessagesScreen.tsx | 538 | Атакующие | `global.атакующие` | ❌ |
| features/messages/MessagesScreen.tsx | 539 | Защитники | `global.защитники` | ❌ |
| features/messages/MessagesScreen.tsx | 545 | Раунды по раундам | `global.раунды_по_раундам` | ❌ |
| features/messages/MessagesScreen.tsx | 548 | Атакующих | `global.атакующих` | ❌ |
| features/messages/MessagesScreen.tsx | 548 | Защитников | `global.защитников` | ❌ |
| features/notepad/NotepadScreen.tsx | 57 | 📝 Блокнот | `global.блокнот` | ❌ |
| features/notepad/NotepadScreen.tsx | 68 | Координаты целей, заметки по разведке, планы…&#10;&#10;По… | `global.координаты_целей_заметки` | ❌ |
| features/notepad/NotepadScreen.tsx | 96 | 💾 Сохраняется… | `global.сохраняется` | ❌ |
| features/notepad/NotepadScreen.tsx | 97 | ✓ Сохранено | `global.сохранено` | ❌ |
| features/notepad/NotepadScreen.tsx | 98 | ⚠ Ошибка сохранения | `global.ошибка_сохранения` | ❌ |
| features/officers/OfficersScreen.tsx | 19 | Производство | `global.производство` | ✅ |
| features/officers/OfficersScreen.tsx | 20 | Строительство | `global.строительство` | ✅ |
| features/officers/OfficersScreen.tsx | 21 | Исследования | `global.исследования` | ✅ |
| features/officers/OfficersScreen.tsx | 22 | Энергия | `global.энергия` | ✅ |
| features/officers/OfficersScreen.tsx | 23 | Склад | `global.склад` | ❌ |
| features/officers/OfficersScreen.tsx | 78 | Офицер | `global.офицер` | ❌ |
| features/officers/OfficersScreen.tsx | 78 | ${e.title} активирован на ${e.duration_days} дн. | `global.etitle_активирован_на` | ❌ |
| features/officers/OfficersScreen.tsx | 83 | Ошибка | `global.ошибка` | ✅ |
| features/officers/OfficersScreen.tsx | 103 | Баланс: | `global.баланс` | ❌ |
| features/officers/OfficersScreen.tsx | 111 | Нет доступных офицеров. | `global.нет_доступных_офицеров` | ❌ |
| features/officers/OfficersScreen.tsx | 164 | Недостаточно кредитов | `global.недостаточно_кредитов` | ❌ |
| features/officers/OfficersScreen.tsx | 167 | Мало cr | `global.мало_cr` | ❌ |
| features/officers/OfficersScreen.tsx | 167 | Активировать (${e.cost_credit} cr) | `global.активировать_ecostcredit_cr` | ❌ |
| features/overview/ForecastWidget.tsx | 48 | Прогноз через: | `global.прогноз_через` | ❌ |
| features/overview/ForecastWidget.tsx | 65 | Достигнут лимит хранилища | `global.достигнут_лимит_хранилища` | ❌ |
| features/overview/OverviewScreen.tsx | 37 | Транспорт | `global.транспорт` | ✅ |
| features/overview/OverviewScreen.tsx | 37 | Колонизация | `global.колонизация` | ✅ |
| features/overview/OverviewScreen.tsx | 37 | Переработка | `global.переработка` | ✅ |
| features/overview/OverviewScreen.tsx | 38 | Атака | `global.атака` | ✅ |
| features/overview/OverviewScreen.tsx | 38 | Шпионаж | `global.шпионаж` | ✅ |
| features/overview/OverviewScreen.tsx | 38 | Экспедиция | `global.экспедиция` | ✅ |
| features/overview/OverviewScreen.tsx | 131 | Нет планет. Попробуйте перезагрузить страницу. | `global.нет_планет_попробуйте` | ❌ |
| features/overview/OverviewScreen.tsx | 158 | непрочитанное сообщение | `global.непрочитанное_сообщение` | ❌ |
| features/overview/OverviewScreen.tsx | 158 | непрочитанных сообщений | `global.непрочитанных_сообщений` | ❌ |
| features/overview/OverviewScreen.tsx | 165 | Очки | `global.очки` | ✅ |
| features/overview/OverviewScreen.tsx | 166 | Место в рейтинге | `global.место_в_рейтинге` | ❌ |
| features/overview/OverviewScreen.tsx | 168 | Боевой опыт | `global.боевой_опыт` | ❌ |
| features/overview/OverviewScreen.tsx | 171 | Профессия | `global.профессия` | ❌ |
| features/overview/OverviewScreen.tsx | 176 | Сейчас играют | `global.сейчас_играют` | ❌ |
| features/overview/OverviewScreen.tsx | 177 | За 24 часа | `global.за_24_часа` | ❌ |
| features/overview/OverviewScreen.tsx | 283 | Металл | `global.металл` | ✅ |
| features/overview/OverviewScreen.tsx | 284 | Кремний | `global.кремний` | ✅ |
| features/overview/OverviewScreen.tsx | 285 | Водород | `global.водород` | ✅ |
| features/overview/OverviewScreen.tsx | 319 | Флот отозван | `global.флот_отозван` | ❌ |
| features/overview/OverviewScreen.tsx | 321 | Ошибка | `global.ошибка` | ✅ |
| features/overview/OverviewScreen.tsx | 321 | Не удалось отозвать флот | `global.не_удалось_отозвать` | ❌ |
| features/overview/OverviewScreen.tsx | 335 | ← Возврат | `global.возврат` | ❌ |
| features/overview/OverviewScreen.tsx | 335 | → В пути | `global.в_пути` | ❌ |
| features/overview/OverviewScreen.tsx | 454 | луна | `global.луна` | ❌ |
| features/overview/OverviewScreen.tsx | 461 | активность | `global.активность` | ❌ |
| features/overview/OverviewScreen.tsx | 468 | Параметры планеты | `global.параметры_планеты` | ❌ |
| features/overview/OverviewScreen.tsx | 482 | Диаметр | `global.диаметр` | ✅ |
| features/overview/OverviewScreen.tsx | 482 | ${diameter.toLocaleString('ru-RU')} км | `global.diametertolocalestringruru_км` | ❌ |
| features/overview/OverviewScreen.tsx | 485 | Поля | `global.поля` | ❌ |
| features/overview/OverviewScreen.tsx | 488 | Температура | `global.температура` | ✅ |
| features/overview/OverviewScreen.tsx | 501 | Металл | `global.металл` | ✅ |
| features/overview/OverviewScreen.tsx | 502 | Кремний | `global.кремний` | ✅ |
| features/overview/OverviewScreen.tsx | 503 | Водород | `global.водород` | ✅ |
| features/overview/OverviewScreen.tsx | 516 | ${buildingName(item.unit_id)} → ур. ${item.target_level} | `global.buildingnameitemunitid_ур_itemtargetlevel` | ❌ |
| features/overview/OverviewScreen.tsx | 536 | ${nameOf(item.unit_id)} → ур. ${item.target_level} | `global.nameofitemunitid_ур_itemtargetlevel` | ❌ |
| features/overview/OverviewScreen.tsx | 607 | 🔗 Реферальная ссылка: | `global.реферальная_ссылка` | ❌ |
| features/overview/OverviewScreen.tsx | 612 | Скопировать | `global.скопировать` | ❌ |
| features/overview/OverviewScreen.tsx | 612 | ✅ Скопировано | `global.скопировано` | ❌ |
| features/payment/CreditsScreen.tsx | 31 | ✅ оплачен | `payment.оплачен` | ❌ |
| features/payment/CreditsScreen.tsx | 32 | ⏳ ожидает | `payment.ожидает` | ❌ |
| features/payment/CreditsScreen.tsx | 33 | ❌ ошибка | `payment.ошибка` | ❌ |
| features/payment/CreditsScreen.tsx | 34 | ↩️ возврат | `payment.возврат` | ❌ |
| features/payment/CreditsScreen.tsx | 44 | 💼 Все 4 офицера на месяц + запас на артефакты | `payment.все_4_офицера` | ❌ |
| features/payment/CreditsScreen.tsx | 45 | ⭐ Все 4 офицера на 2 недели | `payment.все_4_офицера` | ❌ |
| features/payment/CreditsScreen.tsx | 46 | 👔 Пара офицеров на месяц или смена профессии | `payment.пара_офицеров_на` | ❌ |
| features/payment/CreditsScreen.tsx | 47 | 🎖 Офицер на 2 недели | `payment.офицер_на_2` | ❌ |
| features/payment/CreditsScreen.tsx | 48 | 🔧 Мелкие покупки (смена имени, артефакт) | `payment.мелкие_покупки_смена` | ❌ |
| features/payment/CreditsScreen.tsx | 53 | Адмирал / Инженер / Геолог / Меркуре | `payment.адмирал_инженер_геолог` | ❌ |
| features/payment/CreditsScreen.tsx | 53 | 50 кр/день (1500 кр / месяц) | `payment.50_крдень_1500` | ❌ |
| features/payment/CreditsScreen.tsx | 54 | Смена профессии (после 7 дней) | `payment.смена_профессии_после` | ❌ |
| features/payment/CreditsScreen.tsx | 54 | 1000 кр | `payment.1000_кр` | ❌ |
| features/payment/CreditsScreen.tsx | 55 | Переименование планеты | `payment.переименование_планеты` | ❌ |
| features/payment/CreditsScreen.tsx | 55 | бесплатно | `payment.бесплатно` | ❌ |
| features/payment/CreditsScreen.tsx | 56 | Покупка артефактов на бирже | `payment.покупка_артефактов_на` | ❌ |
| features/payment/CreditsScreen.tsx | 56 | зависит от лота | `payment.зависит_от_лота` | ❌ |
| features/payment/CreditsScreen.tsx | 95 | Не удалось создать заказ. Попробуйте позже. | `payment.не_удалось_создать` | ❌ |
| features/payment/CreditsScreen.tsx | 103 | Пополнение кредитов | `payment.пополнение_кредитов` | ❌ |
| features/payment/CreditsScreen.tsx | 127 | Загрузка пакетов… | `payment.загрузка_пакетов` | ❌ |
| features/payment/CreditsScreen.tsx | 128 | Ошибка загрузки пакетов | `payment.ошибка_загрузки_пакетов` | ❌ |
| features/payment/CreditsScreen.tsx | 178 | История покупок | `payment.история_покупок` | ❌ |
| features/payment/CreditsScreen.tsx | 180 | Загрузка… | `payment.загрузка` | ❌ |
| features/payment/CreditsScreen.tsx | 181 | Ошибка загрузки истории | `payment.ошибка_загрузки_истории` | ❌ |
| features/payment/CreditsScreen.tsx | 182 | Покупок пока нет. | `payment.покупок_пока_нет` | ❌ |
| features/payment/CreditsScreen.tsx | 188 | Дата | `payment.дата` | ✅ |
| features/payment/CreditsScreen.tsx | 189 | Пакет | `payment.пакет` | ❌ |
| features/payment/CreditsScreen.tsx | 190 | Кредиты | `payment.кредиты` | ✅ |
| features/payment/CreditsScreen.tsx | 191 | Сумма | `payment.сумма` | ❌ |
| features/payment/CreditsScreen.tsx | 192 | Статус | `payment.статус` | ✅ |
| features/planet-options/PlanetOptionsScreen.tsx | 34 | Планета переименована | `global.планета_переименована` | ❌ |
| features/planet-options/PlanetOptionsScreen.tsx | 34 | Новое имя: ${newName} | `global.новое_имя_newname` | ❌ |
| features/planet-options/PlanetOptionsScreen.tsx | 37 | Ошибка переименования | `global.ошибка_переименования` | ❌ |
| features/planet-options/PlanetOptionsScreen.tsx | 46 | Главная планета | `global.главная_планета` | ✅ |
| features/planet-options/PlanetOptionsScreen.tsx | 46 | ${planet.name} теперь ваша главная планета | `global.planetname_теперь_ваша` | ❌ |
| features/planet-options/PlanetOptionsScreen.tsx | 50 | Ошибка | `global.ошибка` | ✅ |
| features/planet-options/PlanetOptionsScreen.tsx | 60 | Планета покинута | `global.планета_покинута` | ❌ |
| features/planet-options/PlanetOptionsScreen.tsx | 60 | ${planet.name} была удалена | `global.planetname_была_удалена` | ❌ |
| features/planet-options/PlanetOptionsScreen.tsx | 64 | Ошибка при удалении | `global.ошибка_при_удалении` | ❌ |
| features/planet-options/PlanetOptionsScreen.tsx | 102 | Новое имя | `global.новое_имя` | ❌ |
| features/planet-options/PlanetOptionsScreen.tsx | 111 | Сохранить | `global.сохранить` | ❌ |
| features/planet-options/PlanetOptionsScreen.tsx | 180 | Покинуть | `global.покинуть` | ❌ |
| features/planet-options/PlanetOptionsScreen.tsx | 197 | Установить "${planet.name}" как главную планету? | `global.установить_planetname_как` | ❌ |
| features/planet-options/PlanetOptionsScreen.tsx | 199 | Отмена | `global.отмена` | ✅ |
| features/planet-options/PlanetOptionsScreen.tsx | 207 | Вы уверены? Планета "${planet.name}" будет удалена без во… | `global.вы_уверены_планета` | ❌ |
| features/planet-options/PlanetOptionsScreen.tsx | 209 | Удалить | `global.удалить` | ✅ |
| features/planet-options/PlanetOptionsScreen.tsx | 210 | Отмена | `global.отмена` | ✅ |
| features/profession/ProfessionScreen.tsx | 19 | Рудник металла | `global.рудник_металла` | ❌ |
| features/profession/ProfessionScreen.tsx | 20 | Рудник кремния | `global.рудник_кремния` | ❌ |
| features/profession/ProfessionScreen.tsx | 21 | Солнечная электростанция | `global.солнечная_электростанция` | ✅ |
| features/profession/ProfessionScreen.tsx | 22 | Верфь | `global.верфь` | ✅ |
| features/profession/ProfessionScreen.tsx | 23 | Оружейная технология | `global.оружейная_технология` | ✅ |
| features/profession/ProfessionScreen.tsx | 24 | Щитовая технология | `global.щитовая_технология` | ✅ |
| features/profession/ProfessionScreen.tsx | 25 | Броневая технология | `global.броневая_технология` | ✅ |
| features/profession/ProfessionScreen.tsx | 26 | Баллистика | `global.баллистика` | ✅ |
| features/profession/ProfessionScreen.tsx | 27 | Маскировка | `global.маскировка` | ✅ |
| features/profession/ProfessionScreen.tsx | 28 | Оборонный завод | `global.оборонный_завод` | ✅ |
| features/profession/ProfessionScreen.tsx | 29 | Ракетная шахта | `global.ракетная_шахта` | ✅ |
| features/profession/ProfessionScreen.tsx | 30 | Компьютерная технология | `global.компьютерная_технология` | ✅ |
| features/profession/ProfessionScreen.tsx | 31 | Гравитационная технология | `global.гравитационная_технология` | ✅ |
| features/profession/ProfessionScreen.tsx | 32 | Реактивный двигатель | `global.реактивный_двигатель` | ✅ |
| features/profession/ProfessionScreen.tsx | 33 | Импульсный двигатель | `global.импульсный_двигатель` | ✅ |
| features/profession/ProfessionScreen.tsx | 34 | Гиперпространственный двигатель | `global.гиперпространственный_двигатель` | ✅ |
| features/profession/ProfessionScreen.tsx | 53 | доступно | `global.доступно` | ❌ |
| features/profession/ProfessionScreen.tsx | 56 | ${d}д ${h}ч | `global.dд_hч` | ❌ |
| features/profession/ProfessionScreen.tsx | 58 | ${h}ч ${m}м | `global.hч_mм` | ❌ |
| features/profession/ProfessionScreen.tsx | 83 | Профессия изменена | `global.профессия_изменена` | ❌ |
| features/profession/ProfessionScreen.tsx | 83 | Теперь вы ${prof.label} | `global.теперь_вы_proflabel` | ❌ |
| features/profession/ProfessionScreen.tsx | 86 | Ошибка | `global.ошибка` | ✅ |
| features/profession/ProfessionScreen.tsx | 88 | Слишком рано | `global.слишком_рано` | ❌ |
| features/profession/ProfessionScreen.tsx | 88 | Смена профессии доступна раз в 14 дней | `global.смена_профессии_доступна` | ❌ |
| features/profession/ProfessionScreen.tsx | 90 | Смена профессии стоит 1000 кредитов | `global.смена_профессии_стоит` | ❌ |
| features/profession/ProfessionScreen.tsx | 90 | Недостаточно кредитов | `global.недостаточно_кредитов` | ❌ |
| features/profession/ProfessionScreen.tsx | 92 | Ошибка | `global.ошибка` | ✅ |
| features/profession/ProfessionScreen.tsx | 109 | Профессия | `global.профессия` | ❌ |
| features/profession/ProfessionScreen.tsx | 117 | Нет профессии | `global.нет_профессии` | ❌ |
| features/profession/ProfessionScreen.tsx | 125 | 1 000 кредитов | `global.1_000_кредитов` | ❌ |
| features/profession/ProfessionScreen.tsx | 128 | 14 дней | `global.14_дней` | ❌ |
| features/profession/ProfessionScreen.tsx | 171 | ● активна | `global.активна` | ❌ |
| features/profession/ProfessionScreen.tsx | 215 | (1 000 кредитов) | `global.1_000_кредитов` | ❌ |
| features/profession/ProfessionScreen.tsx | 216 | Сменить профессию на "${p.label}"?${label} | `global.сменить_профессию_на` | ❌ |
| features/profession/ProfessionScreen.tsx | 221 | Смена… | `global.смена` | ❌ |
| features/profession/ProfessionScreen.tsx | 221 | Выбрать${canChangeFree ? '' : ' (1000 💳)'} | `global.выбратьcanchangefree_1000` | ❌ |
| features/records/RecordsScreen.tsx | 19 | Постройки | `global.постройки` | ✅ |
| features/records/RecordsScreen.tsx | 20 | Исследования | `global.исследования` | ✅ |
| features/records/RecordsScreen.tsx | 21 | Флот | `global.флот` | ✅ |
| features/records/RecordsScreen.tsx | 22 | Оборона | `global.оборона` | ✅ |
| features/records/RecordsScreen.tsx | 23 | Очки | `global.очки` | ✅ |
| features/records/RecordsScreen.tsx | 51 | 🏅 Рекорды сервера | `global.рекорды_сервера` | ❌ |
| features/records/RecordsScreen.tsx | 57 | 📊 Все | `global.все` | ❌ |
| features/records/RecordsScreen.tsx | 65 | 🔍 Поиск… | `global.поиск` | ❌ |
| features/records/RecordsScreen.tsx | 83 | Категория | `global.категория` | ❌ |
| features/records/RecordsScreen.tsx | 84 | Позиция | `global.позиция` | ❌ |
| features/records/RecordsScreen.tsx | 85 | Держатель | `global.держатель` | ❌ |
| features/records/RecordsScreen.tsx | 86 | Рекорд | `global.рекорд` | ❌ |
| features/records/RecordsScreen.tsx | 87 | Мой результат | `global.мой_результат` | ❌ |
| features/referral/ReferralScreen.tsx | 45 | Не удалось скопировать | `referral.не_удалось_скопировать` | ❌ |
| features/referral/ReferralScreen.tsx | 54 | oxsar — космическая стратегия | `referral.oxsar_космическая_стратегия` | ❌ |
| features/referral/ReferralScreen.tsx | 71 | 🎁 Реферальная программа | `referral.реферальная_программа` | ❌ |
| features/referral/ReferralScreen.tsx | 74 | Ваша ссылка | `referral.ваша_ссылка` | ❌ |
| features/referral/ReferralScreen.tsx | 85 | ✓ Скопировано | `referral.скопировано` | ❌ |
| features/referral/ReferralScreen.tsx | 85 | 📋 Копировать | `referral.копировать` | ❌ |
| features/referral/ReferralScreen.tsx | 100 | Приглашено | `referral.приглашено` | ❌ |
| features/referral/ReferralScreen.tsx | 102 | Бонусные очки | `referral.бонусные_очки` | ❌ |
| features/referral/ReferralScreen.tsx | 105 | макс. ${(d?.max_bonus_points ?? 0).toLocaleString('ru-RU')} | `referral.макс_dmaxbonuspoints_0tolocalestringruru` | ❌ |
| features/referral/ReferralScreen.tsx | 108 | Бонус от покупок | `referral.бонус_от_покупок` | ❌ |
| features/referral/ReferralScreen.tsx | 111 | с каждой покупки кредитов реферала | `referral.с_каждой_покупки` | ❌ |
| features/referral/ReferralScreen.tsx | 116 | Приглашённые игроки | `referral.приглашённые_игроки` | ❌ |
| features/referral/ReferralScreen.tsx | 126 | Игрок | `referral.игрок` | ✅ |
| features/referral/ReferralScreen.tsx | 127 | Очки | `referral.очки` | ✅ |
| features/referral/ReferralScreen.tsx | 128 | Зарегистрирован | `referral.зарегистрирован` | ❌ |
| features/repair/RepairScreen.tsx | 69 | Ремонт | `repair.ремонт` | ✅ |
| features/repair/RepairScreen.tsx | 69 | ${nameOf(unitId)} отправлен на ремонт | `repair.nameofunitid_отправлен_на` | ❌ |
| features/repair/RepairScreen.tsx | 71 | Ошибка | `repair.ошибка` | ✅ |
| features/repair/RepairScreen.tsx | 71 | Ошибка ремонта | `repair.ошибка_ремонта` | ❌ |
| features/repair/RepairScreen.tsx | 81 | Разбор | `repair.разбор` | ❌ |
| features/repair/RepairScreen.tsx | 81 | ${nameOf(unitId)} × ${count} отправлено на разбор | `repair.nameofunitid_count_отправлено` | ❌ |
| features/repair/RepairScreen.tsx | 83 | Ошибка | `repair.ошибка` | ✅ |
| features/repair/RepairScreen.tsx | 83 | Ошибка разбора | `repair.ошибка_разбора` | ❌ |
| features/repair/RepairScreen.tsx | 94 | Отменено | `repair.отменено` | ❌ |
| features/repair/RepairScreen.tsx | 94 | Задание отменено, ресурсы возвращены | `repair.задание_отменено_ресурсы` | ❌ |
| features/repair/RepairScreen.tsx | 96 | Ошибка | `repair.ошибка` | ✅ |
| features/repair/RepairScreen.tsx | 96 | Не удалось отменить | `repair.не_удалось_отменить` | ❌ |
| features/repair/RepairScreen.tsx | 115 | Хранилище | `repair.хранилище` | ✅ |
| features/repair/RepairScreen.tsx | 139 | Ремонт | `repair.ремонт` | ✅ |
| features/repair/RepairScreen.tsx | 139 | Разбор | `repair.разбор` | ❌ |
| features/repair/RepairScreen.tsx | 151 | Отменить | `repair.отменить` | ❌ |
| features/research/ResearchScreen.tsx | 10 | ${secs}с | `research.secsс` | ❌ |
| features/research/ResearchScreen.tsx | 14 | ${d}д ${h}ч ${m}м | `research.dд_hч_mм` | ❌ |
| features/research/ResearchScreen.tsx | 15 | ${h}ч ${m}м | `research.hч_mм` | ❌ |
| features/research/ResearchScreen.tsx | 16 | ${m}м | `research.mм` | ❌ |
| features/research/ResearchScreen.tsx | 69 | Исследование | `research.исследование` | ✅ |
| features/research/ResearchScreen.tsx | 69 | ${name} запущено | `research.name_запущено` | ❌ |
| features/research/ResearchScreen.tsx | 73 | Лаборатория занята | `research.лаборатория_занята` | ❌ |
| features/research/ResearchScreen.tsx | 74 | Недостаточно ресурсов | `research.недостаточно_ресурсов` | ❌ |
| features/research/ResearchScreen.tsx | 75 | Не удалось запустить | `research.не_удалось_запустить` | ❌ |
| features/research/ResearchScreen.tsx | 76 | Ошибка | `research.ошибка` | ✅ |
| features/research/ResearchScreen.tsx | 138 | Подробнее | `research.подробнее` | ❌ |
| features/research/ResearchScreen.tsx | 143 | Не изучено | `research.не_изучено` | ❌ |
| features/research/ResearchScreen.tsx | 143 | Уровень ${level} | `research.уровень_level` | ❌ |
| features/research/ResearchScreen.tsx | 198 | 🔬 Изучается | `research.изучается` | ❌ |
| features/research/ResearchScreen.tsx | 198 | ⏳ Занято | `research.занято` | ❌ |
| features/research/ResearchScreen.tsx | 198 | Изучить | `research.изучить` | ❌ |
| features/research/ResearchScreen.tsx | 198 | → ур. ${level + 1} | `research.ур_level_1` | ❌ |
| features/resource/ResourceScreen.tsx | 65 | Ошибка | `global.ошибка` | ✅ |
| features/resource/ResourceScreen.tsx | 65 | Не удалось сохранить | `global.не_удалось_сохранить` | ❌ |
| features/resource/ResourceScreen.tsx | 84 | Ошибка загрузки | `global.ошибка_загрузки` | ❌ |
| features/resource/ResourceScreen.tsx | 104 | сохраняю… | `global.сохраняю` | ❌ |
| features/resource/ResourceScreen.tsx | 105 | Выключить всё | `global.выключить_всё` | ❌ |
| features/resource/ResourceScreen.tsx | 106 | Включить всё | `global.включить_всё` | ❌ |
| features/resource/ResourceScreen.tsx | 122 | Здание | `global.здание` | ❌ |
| features/resource/ResourceScreen.tsx | 132 | Естественное | `global.естественное` | ❌ |
| features/resource/ResourceScreen.tsx | 172 | Вместимость хранилищ | `global.вместимость_хранилищ` | ❌ |
| features/resource/ResourceScreen.tsx | 173 | За час | `global.за_час` | ✅ |
| features/resource/ResourceScreen.tsx | 174 | За сутки | `global.за_сутки` | ❌ |
| features/resource/ResourceScreen.tsx | 175 | За неделю | `global.за_неделю` | ✅ |
| features/rockets/RocketsScreen.tsx | 54 | Ракеты запущены | `global.ракеты_запущены` | ❌ |
| features/rockets/RocketsScreen.tsx | 54 | ${res.count} ракет → [${g}:${s}:${pos}] | `global.rescount_ракет_gspos` | ❌ |
| features/rockets/RocketsScreen.tsx | 57 | Ошибка | `global.ошибка` | ✅ |
| features/rockets/RocketsScreen.tsx | 57 | Не удалось запустить | `global.не_удалось_запустить` | ❌ |
| features/rockets/RocketsScreen.tsx | 72 | межпланетарных ракет | `global.межпланетарных_ракет` | ❌ |
| features/rockets/RocketsScreen.tsx | 92 | Координаты цели | `global.координаты_цели` | ❌ |
| features/rockets/RocketsScreen.tsx | 127 | 💥 Запустить ${count} ракет${count > 4 ? '' : count > 1 ? … | `global.запустить_count_ракетcount` | ❌ |
| features/score/ScoreScreen.tsx | 47 | Общий | `score.общий` | ❌ |
| features/score/ScoreScreen.tsx | 48 | Постройки | `score.постройки` | ✅ |
| features/score/ScoreScreen.tsx | 49 | Исследования | `score.исследования` | ✅ |
| features/score/ScoreScreen.tsx | 50 | Флот | `score.флот` | ✅ |
| features/score/ScoreScreen.tsx | 51 | Достижения | `score.достижения` | ✅ |
| features/score/ScoreScreen.tsx | 52 | Боевой | `score.боевой` | ❌ |
| features/score/ScoreScreen.tsx | 143 | 🔍 Фильтр по нику… | `score.фильтр_по_нику` | ❌ |
| features/score/ScoreScreen.tsx | 162 | Игрок | `score.игрок` | ✅ |
| features/score/ScoreScreen.tsx | 163 | Альянс | `score.альянс` | ✅ |
| features/score/ScoreScreen.tsx | 164 | Координаты | `score.координаты` | ✅ |
| features/score/ScoreScreen.tsx | 185 | Игрок | `score.игрок` | ✅ |
| features/score/ScoreScreen.tsx | 186 | Альянс | `score.альянс` | ✅ |
| features/score/ScoreScreen.tsx | 189 | Координаты | `score.координаты` | ✅ |
| features/score/ScoreScreen.tsx | 195 | Перейти в галактику | `score.перейти_в_галактику` | ❌ |
| features/score/ScoreScreen.tsx | 206 | Постройки | `score.постройки` | ✅ |
| features/score/ScoreScreen.tsx | 207 | Исследования | `score.исследования` | ✅ |
| features/score/ScoreScreen.tsx | 208 | Флот | `score.флот` | ✅ |
| features/score/ScoreScreen.tsx | 246 | Альянс | `score.альянс` | ✅ |
| features/score/ScoreScreen.tsx | 247 | Игроков | `score.игроков` | ❌ |
| features/score/ScoreScreen.tsx | 248 | Очки | `score.очки` | ✅ |
| features/score/ScoreScreen.tsx | 296 | Игрок | `score.игрок` | ✅ |
| features/score/ScoreScreen.tsx | 297 | Альянс | `score.альянс` | ✅ |
| features/score/ScoreScreen.tsx | 298 | Очки | `score.очки` | ✅ |
| features/score/ScoreScreen.tsx | 299 | В отпуске с | `score.в_отпуске_с` | ❌ |
| features/score/ScoreScreen.tsx | 351 | 📥 Получатели | `score.получатели` | ❌ |
| features/score/ScoreScreen.tsx | 352 | 📤 Отправители | `score.отправители` | ❌ |
| features/score/ScoreScreen.tsx | 355 | Всё время | `score.всё_время` | ❌ |
| features/score/ScoreScreen.tsx | 356 | Месяц | `score.месяц` | ✅ |
| features/score/ScoreScreen.tsx | 357 | Неделя | `score.неделя` | ✅ |
| features/score/ScoreScreen.tsx | 374 | Отправитель | `score.отправитель` | ✅ |
| features/score/ScoreScreen.tsx | 374 | Получатель | `score.получатель` | ✅ |
| features/score/ScoreScreen.tsx | 375 | Металл | `score.металл` | ✅ |
| features/score/ScoreScreen.tsx | 376 | Кремний | `score.кремний` | ✅ |
| features/score/ScoreScreen.tsx | 377 | Водород | `score.водород` | ✅ |
| features/score/ScoreScreen.tsx | 378 | Всего (у.е.) | `score.всего_уе` | ❌ |
| features/score/ScoreScreen.tsx | 423 | 👤 Игроки | `score.игроки` | ❌ |
| features/score/ScoreScreen.tsx | 424 | 🤝 Альянсы | `score.альянсы` | ❌ |
| features/score/ScoreScreen.tsx | 425 | ✈ В отпуске | `score.в_отпуске` | ❌ |
| features/score/ScoreScreen.tsx | 426 | 📦 Торговля | `score.торговля` | ❌ |
| features/search/GlobalSearch.tsx | 92 | Поиск игроков, альянсов, планет… | `global.поиск_игроков_альянсов` | ❌ |
| features/search/GlobalSearch.tsx | 115 | Поиск… | `global.поиск` | ❌ |
| features/search/GlobalSearch.tsx | 125 | Игроки | `global.игроки` | ✅ |
| features/search/GlobalSearch.tsx | 144 | Альянсы | `global.альянсы` | ❌ |
| features/search/GlobalSearch.tsx | 163 | Планеты | `global.планеты` | ✅ |
| features/settings/SettingsScreen.tsx | 24 | Москва (UTC+3) | `settings.москва_utc3` | ❌ |
| features/settings/SettingsScreen.tsx | 25 | Киев (UTC+2/+3) | `settings.киев_utc23` | ❌ |
| features/settings/SettingsScreen.tsx | 26 | Минск (UTC+3) | `settings.минск_utc3` | ❌ |
| features/settings/SettingsScreen.tsx | 27 | Екатеринбург (UTC+5) | `settings.екатеринбург_utc5` | ❌ |
| features/settings/SettingsScreen.tsx | 28 | Новосибирск (UTC+7) | `settings.новосибирск_utc7` | ❌ |
| features/settings/SettingsScreen.tsx | 29 | Владивосток (UTC+10) | `settings.владивосток_utc10` | ❌ |
| features/settings/SettingsScreen.tsx | 30 | Алматы (UTC+6) | `settings.алматы_utc6` | ❌ |
| features/settings/SettingsScreen.tsx | 31 | Берлин (UTC+1/+2) | `settings.берлин_utc12` | ❌ |
| features/settings/SettingsScreen.tsx | 32 | Лондон (UTC+0/+1) | `settings.лондон_utc01` | ❌ |
| features/settings/SettingsScreen.tsx | 33 | Нью-Йорк (UTC-5/-4) | `settings.ньюйорк_utc54` | ❌ |
| features/settings/SettingsScreen.tsx | 34 | Лос-Анджелес (UTC-8/-7) | `settings.лосанджелес_utc87` | ❌ |
| features/settings/SettingsScreen.tsx | 35 | Токио (UTC+9) | `settings.токио_utc9` | ❌ |
| features/settings/SettingsScreen.tsx | 88 | Ошибка запроса кода | `settings.ошибка_запроса_кода` | ❌ |
| features/settings/SettingsScreen.tsx | 94 | Неверный код | `settings.неверный_код` | ❌ |
| features/settings/SettingsScreen.tsx | 115 | Ошибка сохранения | `settings.ошибка_сохранения` | ❌ |
| features/settings/SettingsScreen.tsx | 122 | Пароли не совпадают | `settings.пароли_не_совпадают` | ❌ |
| features/settings/SettingsScreen.tsx | 123 | Минимум 8 символов | `settings.минимум_8_символов` | ❌ |
| features/settings/SettingsScreen.tsx | 130 | Неверный текущий пароль | `settings.неверный_текущий_пароль` | ❌ |
| features/settings/SettingsScreen.tsx | 136 | Настройки аккаунта | `settings.настройки_аккаунта` | ❌ |
| features/settings/SettingsScreen.tsx | 140 | Профиль | `settings.профиль` | ❌ |
| features/settings/SettingsScreen.tsx | 162 | ✓ Email обновлён | `settings.email_обновлён` | ❌ |
| features/settings/SettingsScreen.tsx | 167 | Язык | `settings.язык` | ✅ |
| features/settings/SettingsScreen.tsx | 173 | Русский | `settings.русский` | ❌ |
| features/settings/SettingsScreen.tsx | 179 | Часовой пояс | `settings.часовой_пояс` | ❌ |
| features/settings/SettingsScreen.tsx | 194 | Безопасность | `settings.безопасность` | ❌ |
| features/settings/SettingsScreen.tsx | 198 | Текущий пароль | `settings.текущий_пароль` | ❌ |
| features/settings/SettingsScreen.tsx | 208 | Новый пароль | `settings.новый_пароль` | ✅ |
| features/settings/SettingsScreen.tsx | 218 | Повторите новый пароль | `settings.повторите_новый_пароль` | ❌ |
| features/settings/SettingsScreen.tsx | 236 | ✓ Пароль изменён | `settings.пароль_изменён` | ❌ |
| features/settings/SettingsScreen.tsx | 243 | Режим отпуска | `settings.режим_отпуска` | ✅ |
| features/settings/SettingsScreen.tsx | 300 | Ошибка | `settings.ошибка` | ✅ |
| features/settings/SettingsScreen.tsx | 349 | Получить код подтверждения | `settings.получить_код_подтверждения` | ❌ |
| features/settings/SettingsScreen.tsx | 381 | 🗑 Удалить аккаунт навсегда | `settings.удалить_аккаунт_навсегда` | ❌ |
| features/settings/SettingsScreen.tsx | 471 | 💾 Сохранение… | `settings.сохранение` | ❌ |
| features/shipyard/ShipyardScreen.tsx | 36 | Отменено | `shipyard.отменено` | ❌ |
| features/shipyard/ShipyardScreen.tsx | 36 | Задание отменено, ресурсы возвращены | `shipyard.задание_отменено_ресурсы` | ❌ |
| features/shipyard/ShipyardScreen.tsx | 39 | Ошибка | `shipyard.ошибка` | ✅ |
| features/shipyard/ShipyardScreen.tsx | 39 | Не удалось отменить | `shipyard.не_удалось_отменить` | ❌ |
| features/shipyard/ShipyardScreen.tsx | 53 | В очередь | `shipyard.в_очередь` | ✅ |
| features/shipyard/ShipyardScreen.tsx | 53 | ${nameOf(unitId)} × ${count} добавлено в верфь | `shipyard.nameofunitid_count_добавлено` | ❌ |
| features/shipyard/ShipyardScreen.tsx | 56 | Ошибка | `shipyard.ошибка` | ✅ |
| features/shipyard/ShipyardScreen.tsx | 56 | Не удалось добавить | `shipyard.не_удалось_добавить` | ❌ |
| features/shipyard/ShipyardScreen.tsx | 104 | 🔒 Скрыть недоступные | `shipyard.скрыть_недоступные` | ❌ |
| features/shipyard/ShipyardScreen.tsx | 104 | 👁 Показать все | `shipyard.показать_все` | ❌ |
| features/shipyard/ShipyardScreen.tsx | 178 | Подробнее о ${u.name} | `shipyard.подробнее_о_uname` | ❌ |
| features/shipyard/ShipyardScreen.tsx | 195 | Грузоподъёмность | `shipyard.грузоподъёмность` | ❌ |
| features/shipyard/ShipyardScreen.tsx | 282 | Отменить | `shipyard.отменить` | ❌ |
| features/techtree/TechtreeGraph.tsx | 102 | Ур. ${p.node.current_level} | `global.ур_pnodecurrentlevel` | ❌ |
| features/techtree/TechtreeGraph.tsx | 103 | доступно | `global.доступно` | ❌ |
| features/techtree/TechtreeGraph.tsx | 103 | закрыто | `global.закрыто` | ❌ |
| features/techtree/TechtreeScreen.tsx | 29 | Постройки | `global.постройки` | ✅ |
| features/techtree/TechtreeScreen.tsx | 30 | Исследования | `global.исследования` | ✅ |
| features/techtree/TechtreeScreen.tsx | 31 | Флот | `global.флот` | ✅ |
| features/techtree/TechtreeScreen.tsx | 32 | Оборона | `global.оборона` | ✅ |
| features/techtree/TechtreeScreen.tsx | 70 | 🌳 Дерево технологий | `global.дерево_технологий` | ❌ |
| features/techtree/TechtreeScreen.tsx | 83 | Поиск по названию… | `global.поиск_по_названию` | ❌ |
| features/techtree/TechtreeScreen.tsx | 89 | Все | `global.все` | ✅ |
| features/techtree/TechtreeScreen.tsx | 90 | ✓ Доступно | `global.доступно` | ❌ |
| features/techtree/TechtreeScreen.tsx | 91 | 🔒 Закрыто | `global.закрыто` | ❌ |
| features/techtree/TechtreeScreen.tsx | 94 | Карточки | `global.карточки` | ❌ |
| features/techtree/TechtreeScreen.tsx | 94 | 🗂 Карточки | `global.карточки` | ❌ |
| features/techtree/TechtreeScreen.tsx | 95 | Граф | `global.граф` | ❌ |
| features/techtree/TechtreeScreen.tsx | 95 | 🌐 Граф | `global.граф` | ❌ |
| features/techtree/TechtreeScreen.tsx | 139 | ✓ доступно | `global.доступно` | ❌ |
| features/unit-info/UnitInfoScreen.tsx | 13 | ${secs}с | `global.secsс` | ❌ |
| features/unit-info/UnitInfoScreen.tsx | 17 | ${d}д ${h}ч ${m}м | `global.dд_hч_mм` | ❌ |
| features/unit-info/UnitInfoScreen.tsx | 18 | ${h}ч ${m}м | `global.hч_mм` | ❌ |
| features/unit-info/UnitInfoScreen.tsx | 19 | ${m}м | `global.mм` | ❌ |
| features/unit-info/UnitInfoScreen.tsx | 45 | 🟠/ч | `global.ч` | ❌ |
| features/unit-info/UnitInfoScreen.tsx | 46 | 💎/ч | `global.ч` | ❌ |
| features/unit-info/UnitInfoScreen.tsx | 47 | 💧/ч | `global.ч` | ❌ |
| features/unit-info/UnitInfoScreen.tsx | 48 | ⚡/ч | `global.ч` | ❌ |
| features/unit-info/UnitInfoScreen.tsx | 49 | ⚡/ч | `global.ч` | ❌ |
| features/unit-info/UnitInfoScreen.tsx | 121 | Ур. | `global.ур` | ❌ |
| features/unit-info/UnitInfoScreen.tsx | 167 | Время следующего уровня с учётом фабрики роботов и нано-ф… | `global.время_следующего_уровня` | ❌ |
| features/unit-info/UnitInfoScreen.tsx | 167 | Время указано без учёта фабрики роботов и нано-фабрики. | `global.время_указано_без` | ❌ |
| features/unit-info/UnitInfoScreen.tsx | 168 | Время указано без учёта уровня исследовательской лаборато… | `global.время_указано_без` | ❌ |
| features/unit-info/UnitInfoScreen.tsx | 229 | Параметр | `global.параметр` | ❌ |
| features/unit-info/UnitInfoScreen.tsx | 230 | В атаке | `global.в_атаке` | ❌ |
| features/unit-info/UnitInfoScreen.tsx | 231 | В обороне | `global.в_обороне` | ❌ |
| features/unit-info/UnitInfoScreen.tsx | 231 | Значение | `global.значение` | ❌ |
| features/unit-info/UnitInfoScreen.tsx | 237 | ⚔ Атака | `global.атака` | ❌ |
| features/unit-info/UnitInfoScreen.tsx | 238 | 🛡 Щит | `global.щит` | ❌ |
| features/unit-info/UnitInfoScreen.tsx | 239 | ❤ Броня | `global.броня` | ❌ |
| features/unit-info/UnitInfoScreen.tsx | 240 | 🎯 Приоритет цели | `global.приоритет_цели` | ❌ |
| features/unit-info/UnitInfoScreen.tsx | 241 | 🎲 Баллистика | `global.баллистика` | ❌ |
| features/unit-info/UnitInfoScreen.tsx | 242 | 👻 Маскировка | `global.маскировка` | ❌ |
| features/unit-info/UnitInfoScreen.tsx | 246 | ⚔ Атака | `global.атака` | ❌ |
| features/unit-info/UnitInfoScreen.tsx | 247 | 🛡 Щит | `global.щит` | ❌ |
| features/unit-info/UnitInfoScreen.tsx | 248 | ❤ Броня | `global.броня` | ❌ |
| features/unit-info/UnitInfoScreen.tsx | 249 | 🎯 Приоритет цели | `global.приоритет_цели` | ❌ |
| features/unit-info/UnitInfoScreen.tsx | 250 | 🎲 Баллистика | `global.баллистика` | ❌ |
| features/unit-info/UnitInfoScreen.tsx | 251 | 👻 Маскировка | `global.маскировка` | ❌ |
| features/unit-info/UnitInfoScreen.tsx | 266 | 📦 Грузоподъёмность | `global.грузоподъёмность` | ❌ |
| features/unit-info/UnitInfoScreen.tsx | 269 | 🚀 Скорость | `global.скорость` | ❌ |
| features/unit-info/UnitInfoScreen.tsx | 272 | ⛽ Расход топлива | `global.расход_топлива` | ❌ |
| features/unit-info/UnitInfoScreen.tsx | 272 | ${entry.fuel}/ед. | `global.entryfuelед` | ❌ |
| features/unit-info/UnitInfoScreen.tsx | 275 | 🔩 Конструкция | `global.конструкция` | ❌ |
| features/unit-info/UnitInfoScreen.tsx | 278 | 🟠 Металл | `global.металл` | ❌ |
| features/unit-info/UnitInfoScreen.tsx | 281 | 💎 Кремний | `global.кремний` | ❌ |
| features/unit-info/UnitInfoScreen.tsx | 284 | 💧 Водород | `global.водород` | ❌ |
| features/unit-info/UnitInfoScreen.tsx | 287 | ⏱ Время постройки (базовое) | `global.время_постройки_базовое` | ❌ |
| features/unit-info/UnitInfoScreen.tsx | 302 | Цель | `global.цель` | ✅ |
| features/unit-info/UnitInfoScreen.tsx | 303 | Выстрелов за раунд | `global.выстрелов_за_раунд` | ❌ |
| features/unit-info/UnitInfoScreen.tsx | 332 | Атакующий | `global.атакующий` | ✅ |
| features/unit-info/UnitInfoScreen.tsx | 333 | Выстрелов за раунд | `global.выстрелов_за_раунд` | ❌ |
| features/wiki/WikiScreen.tsx | 188 | Вики | `wiki.вики` | ❌ |
| features/wiki/WikiScreen.tsx | 189 | Загрузка… | `wiki.загрузка` | ❌ |
| features/wiki/WikiScreen.tsx | 241 | Загрузка статьи… | `wiki.загрузка_статьи` | ❌ |
| features/wiki/WikiScreen.tsx | 242 | Не удалось загрузить статью. | `wiki.не_удалось_загрузить` | ❌ |
| frontend/src/App.tsx | 104 | Оплата прошла успешно, кредиты зачислены | `global.оплата_прошла_успешно` | ❌ |
| frontend/src/App.tsx | 108 | Оплата не прошла, попробуйте снова | `global.оплата_не_прошла` | ❌ |
| frontend/src/App.tsx | 138 | Alt+H — главный экран | `global.alth_главный_экран` | ❌ |
| frontend/src/App.tsx | 144 | Alt+B — постройки | `global.altb_постройки` | ❌ |
| frontend/src/App.tsx | 150 | Alt+R — исследования | `global.altr_исследования` | ❌ |
| frontend/src/App.tsx | 156 | Alt+M — сообщения | `global.altm_сообщения` | ❌ |
| frontend/src/App.tsx | 161 | Esc — вернуться на главный экран | `global.esc_вернуться_на` | ❌ |
| frontend/src/App.tsx | 221 | Ошибка загрузки. Попробуйте обновить страницу. | `global.ошибка_загрузки_попробуйте` | ❌ |
| frontend/src/App.tsx | 222 | Обновить | `global.обновить` | ❌ |
| frontend/src/App.tsx | 435 | Мет | `global.мет` | ❌ |
| frontend/src/App.tsx | 443 | Крем | `global.крем` | ❌ |
| frontend/src/App.tsx | 451 | Водор | `global.водор` | ❌ |
| frontend/src/App.tsx | 472 | Кредиты | `global.кредиты` | ✅ |
| frontend/src/App.tsx | 486 | Поиск (Ctrl+K) | `global.поиск_ctrlk` | ❌ |
| frontend/src/App.tsx | 495 | Профиль / настройки | `global.профиль_настройки` | ❌ |
| frontend/src/App.tsx | 512 | Панель администратора | `global.панель_администратора` | ❌ |
| frontend/src/App.tsx | 519 | Выйти | `global.выйти` | ❌ |
| frontend/src/App.tsx | 628 | Обзор | `global.обзор` | ✅ |
| frontend/src/App.tsx | 629 | Галакт. | `global.галакт` | ❌ |
| frontend/src/App.tsx | 630 | Флот | `global.флот` | ✅ |
| frontend/src/App.tsx | 631 | Сообщ. | `global.сообщ` | ❌ |
| frontend/src/App.tsx | 656 | Ещё | `global.ещё` | ❌ |
| frontend/src/App.tsx | 679 | Обзор | `global.обзор` | ✅ |
| frontend/src/App.tsx | 680 | Сырьё | `global.сырьё` | ✅ |
| frontend/src/App.tsx | 681 | Постройки | `global.постройки` | ✅ |
| frontend/src/App.tsx | 682 | Исследования | `global.исследования` | ✅ |
| frontend/src/App.tsx | 683 | Верфь | `global.верфь` | ✅ |
| frontend/src/App.tsx | 684 | Ремонт | `global.ремонт` | ✅ |
| frontend/src/App.tsx | 685 | Профессия | `global.профессия` | ❌ |
| frontend/src/App.tsx | 686 | Империя | `global.империя` | ✅ |
| frontend/src/App.tsx | 687 | Техдерево | `global.техдерево` | ❌ |
| frontend/src/App.tsx | 688 | Настройки | `global.настройки` | ✅ |
| frontend/src/App.tsx | 689 | Галактика | `global.галактика` | ✅ |
| frontend/src/App.tsx | 690 | Флот | `global.флот` | ✅ |
| frontend/src/App.tsx | 691 | Ракеты | `global.ракеты` | ✅ |
| frontend/src/App.tsx | 692 | Сообщения | `global.сообщения` | ✅ |
| frontend/src/App.tsx | 693 | Чат | `global.чат` | ❌ |
| frontend/src/App.tsx | 694 | Альянс | `global.альянс` | ✅ |
| frontend/src/App.tsx | 695 | Рынок | `global.рынок` | ✅ |
| frontend/src/App.tsx | 696 | Артефакты | `global.артефакты` | ✅ |
| frontend/src/App.tsx | 697 | Арт-рынок | `global.артрынок` | ❌ |
| frontend/src/App.tsx | 698 | Офицеры | `global.офицеры` | ✅ |
| frontend/src/App.tsx | 699 | Рейтинг | `global.рейтинг` | ✅ |
| frontend/src/App.tsx | 700 | Достижения | `global.достижения` | ✅ |
| frontend/src/App.tsx | 701 | Задания дня | `global.задания_дня` | ❌ |
| frontend/src/App.tsx | 702 | Вики | `global.вики` | ❌ |
| frontend/src/App.tsx | 703 | История боёв | `global.история_боёв` | ❌ |
| frontend/src/App.tsx | 704 | Рекорды | `global.рекорды` | ✅ |
| frontend/src/App.tsx | 705 | Симулятор | `global.симулятор` | ❌ |
| frontend/src/App.tsx | 706 | Блокнот | `global.блокнот` | ✅ |
| frontend/src/App.tsx | 707 | Рефералы | `global.рефералы` | ✅ |
| frontend/src/App.tsx | 708 | Друзья | `global.друзья` | ✅ |
| frontend/src/App.tsx | 740 | Планета | `global.планета` | ✅ |
| frontend/src/App.tsx | 747 | Профессия | `global.профессия` | ❌ |
| frontend/src/App.tsx | 748 | Империя | `global.империя` | ✅ |
| frontend/src/App.tsx | 749 | Техдерево | `global.техдерево` | ❌ |
| frontend/src/App.tsx | 750 | Настройки | `global.настройки` | ✅ |
| frontend/src/App.tsx | 752 | Космос | `global.космос` | ❌ |
| frontend/src/App.tsx | 757 | Общение | `global.общение` | ❌ |
| frontend/src/App.tsx | 759 | Чат | `global.чат` | ❌ |
| frontend/src/App.tsx | 760 | Альянс | `global.альянс` | ✅ |
| frontend/src/App.tsx | 761 | Блокнот | `global.блокнот` | ✅ |
| frontend/src/App.tsx | 762 | Рефералы | `global.рефералы` | ✅ |
| frontend/src/App.tsx | 763 | Друзья | `global.друзья` | ✅ |
| frontend/src/App.tsx | 765 | Торговля | `global.торговля` | ❌ |
| frontend/src/App.tsx | 771 | Статистика | `global.статистика` | ✅ |
| frontend/src/App.tsx | 772 | Рейтинг | `global.рейтинг` | ✅ |
| frontend/src/App.tsx | 773 | Достижения | `global.достижения` | ✅ |
| frontend/src/App.tsx | 774 | Задания дня | `global.задания_дня` | ❌ |
| frontend/src/App.tsx | 775 | Вики | `global.вики` | ❌ |
| frontend/src/App.tsx | 776 | История боёв | `global.история_боёв` | ❌ |
| frontend/src/App.tsx | 777 | Рекорды | `global.рекорды` | ✅ |
| frontend/src/App.tsx | 781 | Админ | `global.админ` | ❌ |
| frontend/src/main.tsx | 46 | Ошибка приложения | `global.ошибка_приложения` | ❌ |
| src/api/catalog.ts | 62 | Рудник металла | `global.рудник_металла` | ❌ |
| src/api/catalog.ts | 62 | Основной поставщик сырья для строительства несущих структ… | `global.основной_поставщик_сырья` | ❌ |
| src/api/catalog.ts | 62 | Добывает металл из недр планеты | `global.добывает_металл_из` | ❌ |
| src/api/catalog.ts | 63 | Рудник по добыче кремния | `global.рудник_по_добыче` | ✅ |
| src/api/catalog.ts | 63 | Добывает кремний для строительства и исследований | `global.добывает_кремний_для` | ❌ |
| src/api/catalog.ts | 63 | Основной поставщик сырья для электронных строительных эле… | `global.основной_поставщик_сырья` | ✅ |
| src/api/catalog.ts | 64 | Синтезатор водорода | `global.синтезатор_водорода` | ✅ |
| src/api/catalog.ts | 64 | Синтезирует водород — топливо для флота | `global.синтезирует_водород_топливо` | ❌ |
| src/api/catalog.ts | 64 | Совершенствование синтезатора способствует увеличению его… | `global.совершенствование_синтезатора_способствует` | ✅ |
| src/api/catalog.ts | 65 | Для обеспечения энергией рудников и синтезаторов необходи… | `global.для_обеспечения_энергией` | ✅ |
| src/api/catalog.ts | 65 | Солнечная электростанция | `global.солнечная_электростанция` | ✅ |
| src/api/catalog.ts | 65 | Вырабатывает энергию из солнечного излучения | `global.вырабатывает_энергию_из` | ❌ |
| src/api/catalog.ts | 66 | Мощная электростанция на термоядерном синтезе | `global.мощная_электростанция_на` | ❌ |
| src/api/catalog.ts | 66 | Термоядерная электростанция | `global.термоядерная_электростанция` | ✅ |
| src/api/catalog.ts | 66 | На термоядерных электростанциях при помощи термоядерного … | `global.на_термоядерных_электростанциях` | ❌ |
| src/api/catalog.ts | 67 | Фабрика роботов | `global.фабрика_роботов` | ✅ |
| src/api/catalog.ts | 67 | Ускоряет строительство зданий | `global.ускоряет_строительство_зданий` | ❌ |
| src/api/catalog.ts | 67 | Предоставляет простую рабочую силу, которую можно применя… | `global.предоставляет_простую_рабочую` | ✅ |
| src/api/catalog.ts | 68 | Фабрика нанитов представляет собой венец робототехники. Н… | `global.фабрика_нанитов_представляет` | ❌ |
| src/api/catalog.ts | 68 | Нано-фабрика | `global.нанофабрика` | ❌ |
| src/api/catalog.ts | 68 | Вдвое ускоряет строительство за каждый уровень | `global.вдвое_ускоряет_строительство` | ❌ |
| src/api/catalog.ts | 69 | Производит корабли и оборонительные системы | `global.производит_корабли_и` | ❌ |
| src/api/catalog.ts | 69 | Верфь | `global.верфь` | ✅ |
| src/api/catalog.ts | 69 | В строительной верфи производятся все виды кораблей. Чем … | `global.в_строительной_верфи` | ✅ |
| src/api/catalog.ts | 70 | Огромное хранилище для добытых руд. Чем оно больше, тем б… | `global.огромное_хранилище_для` | ✅ |
| src/api/catalog.ts | 70 | Хранилище металла | `global.хранилище_металла` | ✅ |
| src/api/catalog.ts | 70 | Увеличивает максимальный запас металла | `global.увеличивает_максимальный_запас` | ❌ |
| src/api/catalog.ts | 71 | В этом огромном хранилище складируется ещё не обработанны… | `global.в_этом_огромном` | ✅ |
| src/api/catalog.ts | 71 | Увеличивает максимальный запас кремния | `global.увеличивает_максимальный_запас` | ❌ |
| src/api/catalog.ts | 71 | Хранилище кремния | `global.хранилище_кремния` | ✅ |
| src/api/catalog.ts | 72 | Огромные ёмкости для хранения добытого водорода. Они обыч… | `global.огромные_ёмкости_для` | ✅ |
| src/api/catalog.ts | 72 | Увеличивает максимальный запас водорода | `global.увеличивает_максимальный_запас` | ❌ |
| src/api/catalog.ts | 72 | Емкость для водорода | `global.емкость_для_водорода` | ✅ |
| src/api/catalog.ts | 73 | Исследовательская лаборатория | `global.исследовательская_лаборатория` | ✅ |
| src/api/catalog.ts | 73 | Позволяет проводить научные исследования | `global.позволяет_проводить_научные` | ❌ |
| src/api/catalog.ts | 73 | Для исследования новых технологий необходима работа иссле… | `global.для_исследования_новых` | ❌ |
| src/api/catalog.ts | 74 | Ракетная шахта | `global.ракетная_шахта` | ✅ |
| src/api/catalog.ts | 74 | Хранит межпланетные ракеты для атаки | `global.хранит_межпланетные_ракеты` | ❌ |
| src/api/catalog.ts | 74 | Ракетные шахты служат для хранения ракет. С каждым уровне… | `global.ракетные_шахты_служат` | ✅ |
| src/api/catalog.ts | 75 | Ремонтный ангар | `global.ремонтный_ангар` | ✅ |
| src/api/catalog.ts | 75 | Восстанавливает повреждённые корабли после боя | `global.восстанавливает_повреждённые_корабли` | ❌ |
| src/api/catalog.ts | 75 | Ремонтный ангар необходим для выполнения двух операций: 1… | `global.ремонтный_ангар_необходим` | ❌ |
| src/api/catalog.ts | 76 | Завод обороны | `global.завод_обороны` | ❌ |
| src/api/catalog.ts | 76 | Ускоряет постройку оборонительных сооружений | `global.ускоряет_постройку_оборонительных` | ❌ |
| src/api/catalog.ts | 77 | Обменник | `global.обменник` | ❌ |
| src/api/catalog.ts | 77 | Быстрый обмен ресурсов по фиксированному курсу | `global.быстрый_обмен_ресурсов` | ❌ |
| src/api/catalog.ts | 78 | Биржа | `global.биржа` | ✅ |
| src/api/catalog.ts | 78 | Книга ордеров на обмен ресурсами между игроками | `global.книга_ордеров_на` | ❌ |
| src/api/catalog.ts | 79 | Терраформер | `global.терраформер` | ✅ |
| src/api/catalog.ts | 79 | Расширяет пригодную для постройки площадь планеты | `global.расширяет_пригодную_для` | ❌ |
| src/api/catalog.ts | 83 | Лунная база | `global.лунная_база` | ✅ |
| src/api/catalog.ts | 83 | Основная постройка луны, даёт возможность строить другие … | `global.основная_постройка_луны` | ❌ |
| src/api/catalog.ts | 83 | Луна не располагает атмосферой, поэтому перед заселением … | `global.луна_не_располагает` | ❌ |
| src/api/catalog.ts | 84 | Звёздные сенсоры | `global.звёздные_сенсоры` | ❌ |
| src/api/catalog.ts | 84 | Следит за флотами противника в системе | `global.следит_за_флотами` | ❌ |
| src/api/catalog.ts | 85 | Звёздные врата | `global.звёздные_врата` | ❌ |
| src/api/catalog.ts | 85 | Позволяет мгновенно переместить флот между лунами | `global.позволяет_мгновенно_переместить` | ❌ |
| src/api/catalog.ts | 85 | Ворота — это огромные телепортеры, которые могут пересыла… | `global.ворота_это_огромные` | ❌ |
| src/api/catalog.ts | 86 | Лунная фабрика роботов | `global.лунная_фабрика_роботов` | ✅ |
| src/api/catalog.ts | 86 | Ускоряет строительство зданий на луне | `global.ускоряет_строительство_зданий` | ❌ |
| src/api/catalog.ts | 86 | Предоставляет простую рабочую силу, которую можно применя… | `global.предоставляет_простую_рабочую` | ✅ |
| src/api/catalog.ts | 98 | Шпионаж | `global.шпионаж` | ✅ |
| src/api/catalog.ts | 98 | +1 уровень шпионажа зонда | `global.1_уровень_шпионажа` | ❌ |
| src/api/catalog.ts | 98 | Шпионаж предназначен для исследования новых и более эффек… | `global.шпионаж_предназначен_для` | ❌ |
| src/api/catalog.ts | 99 | Компьютерная технология | `global.компьютерная_технология` | ✅ |
| src/api/catalog.ts | 99 | +1 слот флота | `global.1_слот_флота` | ❌ |
| src/api/catalog.ts | 99 | Компьютерная технология предназначена для расширения имею… | `global.компьютерная_технология_предназначена` | ❌ |
| src/api/catalog.ts | 100 | Оружейная технология | `global.оружейная_технология` | ✅ |
| src/api/catalog.ts | 100 | +2% атака флота и обороны | `global.2_атака_флота` | ❌ |
| src/api/catalog.ts | 100 | Оружейная технология занимается прежде всего дальнейшим р… | `global.оружейная_технология_занимается` | ❌ |
| src/api/catalog.ts | 101 | Щитовая технология | `global.щитовая_технология` | ✅ |
| src/api/catalog.ts | 101 | +2% щит флота и обороны | `global.2_щит_флота` | ❌ |
| src/api/catalog.ts | 101 | Развитие этой технологии позволяет увеличивать снабжение … | `global.развитие_этой_технологии` | ❌ |
| src/api/catalog.ts | 102 | Броневая технология | `global.броневая_технология` | ✅ |
| src/api/catalog.ts | 102 | +2% броня флота и обороны | `global.2_броня_флота` | ❌ |
| src/api/catalog.ts | 102 | Специальные сплавы улучшают броню космических кораблей. К… | `global.специальные_сплавы_улучшают` | ❌ |
| src/api/catalog.ts | 103 | Гравитонная технология (IGN) | `global.гравитонная_технология_ign` | ❌ |
| src/api/catalog.ts | 103 | требование для линкоров | `global.требование_для_линкоров` | ❌ |
| src/api/catalog.ts | 103 | Технология объединения граничных полей открывает доступ к… | `global.технология_объединения_граничных` | ❌ |
| src/api/catalog.ts | 104 | Гравитонная пушка | `global.гравитонная_пушка` | ❌ |
| src/api/catalog.ts | 104 | требование для Звезды смерти | `global.требование_для_звезды` | ❌ |
| src/api/catalog.ts | 104 | Гравитонная пушка стреляет ускоренными гравитонами. Сильн… | `global.гравитонная_пушка_стреляет` | ❌ |
| src/api/catalog.ts | 105 | Астрофизика | `global.астрофизика` | ❌ |
| src/api/catalog.ts | 105 | +1 слот колонии и экспедиции | `global.1_слот_колонии` | ❌ |
| src/api/catalog.ts | 105 | Астрофизика позволяет колонизировать дальние планеты и пр… | `global.астрофизика_позволяет_колонизировать` | ❌ |
| src/api/catalog.ts | 106 | Межгалактическая исследовательская сеть | `global.межгалактическая_исследовательская_сеть` | ✅ |
| src/api/catalog.ts | 106 | объединяет лаборатории планет | `global.объединяет_лаборатории_планет` | ❌ |
| src/api/catalog.ts | 106 | Межгалактическая исследовательская сеть позволяет лаборат… | `global.межгалактическая_исследовательская_сеть` | ❌ |
| src/api/catalog.ts | 107 | Энергетическая технология | `global.энергетическая_технология` | ✅ |
| src/api/catalog.ts | 107 | требование для высоких технологий | `global.требование_для_высоких` | ❌ |
| src/api/catalog.ts | 107 | Обладание различными видами энергии необходимо для многих… | `global.обладание_различными_видами` | ✅ |
| src/api/catalog.ts | 108 | Гиперпространственная технология | `global.гиперпространственная_технология` | ✅ |
| src/api/catalog.ts | 108 | требование для гипердвигателя | `global.требование_для_гипердвигателя` | ❌ |
| src/api/catalog.ts | 108 | Путём сплетения 4-го и 5-го измерения стало возможным исс… | `global.путём_сплетения_4го` | ❌ |
| src/api/catalog.ts | 109 | Реактивный двигатель | `global.реактивный_двигатель` | ✅ |
| src/api/catalog.ts | 109 | +10% скорость транспортов и истребителей | `global.10_скорость_транспортов` | ❌ |
| src/api/catalog.ts | 109 | Реактивный двигатель основывается на принципе отдачи. Мат… | `global.реактивный_двигатель_основывается` | ❌ |
| src/api/catalog.ts | 110 | Импульсный двигатель | `global.импульсный_двигатель` | ✅ |
| src/api/catalog.ts | 110 | +20% скорость крейсеров и зондов | `global.20_скорость_крейсеров` | ❌ |
| src/api/catalog.ts | 110 | Импульсный двигатель основывается на принципе отдачи, при… | `global.импульсный_двигатель_основывается` | ❌ |
| src/api/catalog.ts | 111 | Гиперпространственный двигатель | `global.гиперпространственный_двигатель` | ✅ |
| src/api/catalog.ts | 111 | +30% скорость линкоров и флагманов | `global.30_скорость_линкоров` | ❌ |
| src/api/catalog.ts | 111 | Благодаря пространственно-временному изгибу в непосредств… | `global.благодаря_пространственновременному_изгибу` | ✅ |
| src/api/catalog.ts | 112 | Лазерная технология | `global.лазерная_технология` | ✅ |
| src/api/catalog.ts | 112 | требование для ионной технологии | `global.требование_для_ионной` | ❌ |
| src/api/catalog.ts | 112 | Лазеры (усиление света при помощи индуцированного выброса… | `global.лазеры_усиление_света` | ✅ |
| src/api/catalog.ts | 113 | Ионная технология | `global.ионная_технология` | ✅ |
| src/api/catalog.ts | 113 | требование для плазменной технологии | `global.требование_для_плазменной` | ❌ |
| src/api/catalog.ts | 113 | Поистине смертоносный наводимый луч из ускоренных ионов. … | `global.поистине_смертоносный_наводимый` | ✅ |
| src/api/catalog.ts | 114 | Плазменная технология | `global.плазменная_технология` | ✅ |
| src/api/catalog.ts | 114 | повышенный урон по ресурсам противника | `global.повышенный_урон_по` | ❌ |
| src/api/catalog.ts | 114 | Дальнейшее развитие ионной технологии, которая ускоряет н… | `global.дальнейшее_развитие_ионной` | ✅ |
| src/api/catalog.ts | 115 | Экспедиционная технология охватывает различные технологии… | `global.экспедиционная_технология_охватывает` | ❌ |
| src/api/catalog.ts | 115 | Экспедиционная технология | `global.экспедиционная_технология` | ✅ |
| src/api/catalog.ts | 115 | +1 слот экспедиции за уровень | `global.1_слот_экспедиции` | ❌ |
| src/api/catalog.ts | 116 | +1 ракета в шахте за уровень | `global.1_ракета_в` | ❌ |
| src/api/catalog.ts | 116 | Технология баллистического анализа позволяет компьютерным… | `global.технология_баллистического_анализа` | ❌ |
| src/api/catalog.ts | 116 | Баллистическая технология | `global.баллистическая_технология` | ✅ |
| src/api/catalog.ts | 117 | Маскировочная технология | `global.маскировочная_технология` | ✅ |
| src/api/catalog.ts | 117 | снижение видимости флота для шпионажа | `global.снижение_видимости_флота` | ❌ |
| src/api/catalog.ts | 117 | Технология радио-локационной маскировки создаёт помехи в … | `global.технология_радиолокационной_маскировки` | ❌ |
| src/api/catalog.ts | 121 | Дешёвый транспорт для перевозки ресурсов | `global.дешёвый_транспорт_для` | ❌ |
| src/api/catalog.ts | 121 | Малый транспорт | `global.малый_транспорт` | ✅ |
| src/api/catalog.ts | 122 | Большой транспорт | `global.большой_транспорт` | ✅ |
| src/api/catalog.ts | 122 | Основной грузовоз для перевозки ресурсов | `global.основной_грузовоз_для` | ❌ |
| src/api/catalog.ts | 123 | Легкий истребитель | `global.легкий_истребитель` | ✅ |
| src/api/catalog.ts | 123 | Дешёвый и быстрый — основа атакующего флота | `global.дешёвый_и_быстрый` | ❌ |
| src/api/catalog.ts | 124 | Тяжелый истребитель | `global.тяжелый_истребитель` | ✅ |
| src/api/catalog.ts | 124 | Мощная альтернатива лёгкому истребителю | `global.мощная_альтернатива_лёгкому` | ❌ |
| src/api/catalog.ts | 125 | Крейсер | `global.крейсер` | ✅ |
| src/api/catalog.ts | 125 | Эффективен против ракетных установок | `global.эффективен_против_ракетных` | ❌ |
| src/api/catalog.ts | 126 | Линкор | `global.линкор` | ✅ |
| src/api/catalog.ts | 126 | Мощный боевой корабль с гиперпространственным двигателем | `global.мощный_боевой_корабль` | ❌ |
| src/api/catalog.ts | 127 | Быстрый ударный корабль среднего класса | `global.быстрый_ударный_корабль` | ❌ |
| src/api/catalog.ts | 127 | Фрегат | `global.фрегат` | ❌ |
| src/api/catalog.ts | 128 | Колонизатор | `global.колонизатор` | ✅ |
| src/api/catalog.ts | 128 | Позволяет колонизировать незанятые планеты | `global.позволяет_колонизировать_незанятые` | ❌ |
| src/api/catalog.ts | 129 | Переработчик | `global.переработчик` | ✅ |
| src/api/catalog.ts | 129 | Собирает ресурсы из полей обломков | `global.собирает_ресурсы_из` | ❌ |
| src/api/catalog.ts | 130 | Шпионский зонд | `global.шпионский_зонд` | ✅ |
| src/api/catalog.ts | 130 | Разведывает планеты — при слабом шпионаже может быть пере… | `global.разведывает_планеты_при` | ❌ |
| src/api/catalog.ts | 131 | Солнечный спутник | `global.солнечный_спутник` | ✅ |
| src/api/catalog.ts | 131 | Добавляет энергию без строительства электростанций | `global.добавляет_энергию_без` | ❌ |
| src/api/catalog.ts | 132 | Специализируется на уничтожении оборонительных сооружений | `global.специализируется_на_уничтожении` | ❌ |
| src/api/catalog.ts | 132 | Бомбардировщик | `global.бомбардировщик` | ❌ |
| src/api/catalog.ts | 133 | Звёздный разрушитель | `global.звёздный_разрушитель` | ❌ |
| src/api/catalog.ts | 133 | Тяжёлый боевой корабль выше класса линкора | `global.тяжёлый_боевой_корабль` | ❌ |
| src/api/catalog.ts | 134 | Звезда смерти | `global.звезда_смерти` | ✅ |
| src/api/catalog.ts | 134 | Сильнейший корабль — способен уничтожить луну | `global.сильнейший_корабль_способен` | ❌ |
| src/api/catalog.ts | 135 | Запускается через ракетную шахту — наносит урон обороне н… | `global.запускается_через_ракетную` | ❌ |
| src/api/catalog.ts | 135 | Межпланетная ракета | `global.межпланетная_ракета` | ✅ |
| src/api/catalog.ts | 136 | Лансер | `global.лансер` | ✅ |
| src/api/catalog.ts | 136 | Премиум-корабль с высокой атакой — атакует раньше обычных | `global.премиумкорабль_с_высокой` | ❌ |
| src/api/catalog.ts | 137 | Корабль-призрак | `global.корабльпризрак` | ❌ |
| src/api/catalog.ts | 137 | Стелс-корабль с anti-DS ролью — высокая маскировка | `global.стелскорабль_с_antids` | ❌ |
| src/api/catalog.ts | 139 | Корвет пришельцев | `global.корвет_пришельцев` | ❌ |
| src/api/catalog.ts | 139 | Корабль AI — не строится игроком | `global.корабль_ai_не` | ❌ |
| src/api/catalog.ts | 140 | Корабль AI — не строится игроком | `global.корабль_ai_не` | ❌ |
| src/api/catalog.ts | 140 | Прикрытие пришельцев | `global.прикрытие_пришельцев` | ❌ |
| src/api/catalog.ts | 141 | Паладин пришельцев | `global.паладин_пришельцев` | ❌ |
| src/api/catalog.ts | 141 | Корабль AI — не строится игроком | `global.корабль_ai_не` | ❌ |
| src/api/catalog.ts | 142 | Фрегат пришельцев | `global.фрегат_пришельцев` | ❌ |
| src/api/catalog.ts | 142 | Корабль AI — не строится игроком | `global.корабль_ai_не` | ❌ |
| src/api/catalog.ts | 143 | Торпедоносец пришельцев | `global.торпедоносец_пришельцев` | ❌ |
| src/api/catalog.ts | 143 | Корабль AI — не строится игроком | `global.корабль_ai_не` | ❌ |
| src/api/catalog.ts | 147 | Ракетная установка | `global.ракетная_установка` | ✅ |
| src/api/catalog.ts | 147 | Базовая и дешёвая оборонительная установка | `global.базовая_и_дешёвая` | ❌ |
| src/api/catalog.ts | 148 | Лазерная пушка начального уровня | `global.лазерная_пушка_начального` | ❌ |
| src/api/catalog.ts | 148 | Легкий лазер | `global.легкий_лазер` | ✅ |
| src/api/catalog.ts | 149 | Тяжелый лазер | `global.тяжелый_лазер` | ✅ |
| src/api/catalog.ts | 149 | Усиленная лазерная пушка с большей мощностью | `global.усиленная_лазерная_пушка` | ❌ |
| src/api/catalog.ts | 150 | Ионное орудие | `global.ионное_орудие` | ✅ |
| src/api/catalog.ts | 150 | Высокий щит делает её устойчивой против обычных атак | `global.высокий_щит_делает` | ❌ |
| src/api/catalog.ts | 151 | Пушка Гаусса | `global.пушка_гаусса` | ✅ |
| src/api/catalog.ts | 151 | Мощная пушка — эффективна против тяжёлых кораблей | `global.мощная_пушка_эффективна` | ❌ |
| src/api/catalog.ts | 152 | Плазменное орудие | `global.плазменное_орудие` | ✅ |
| src/api/catalog.ts | 152 | Наиболее разрушительное орудие обороны | `global.наиболее_разрушительное_орудие` | ❌ |
| src/api/catalog.ts | 153 | Малый щитовой купол | `global.малый_щитовой_купол` | ✅ |
| src/api/catalog.ts | 153 | Защищает всю оборону планеты от одного залпа | `global.защищает_всю_оборону` | ❌ |
| src/api/catalog.ts | 154 | Усиленный купол — щит в 5× больше малого | `global.усиленный_купол_щит` | ❌ |
| src/api/catalog.ts | 154 | Большой щитовой купол | `global.большой_щитовой_купол` | ✅ |
| src/api/catalog.ts | 161 | Знак торговца | `global.знак_торговца` | ✅ |
| src/api/catalog.ts | 161 | +3% курс обмена ресурсов | `global.3_курс_обмена` | ❌ |
| src/api/catalog.ts | 161 | 7 дней | `global.7_дней` | ❌ |
| src/api/catalog.ts | 162 | Катализатор | `global.катализатор` | ✅ |
| src/api/catalog.ts | 162 | +10% добыча на всех планетах | `global.10_добыча_на` | ❌ |
| src/api/catalog.ts | 162 | 7 дней | `global.7_дней` | ❌ |
| src/api/catalog.ts | 163 | Энерготранс | `global.энерготранс` | ✅ |
| src/api/catalog.ts | 163 | +15% энергия на всех планетах | `global.15_энергия_на` | ❌ |
| src/api/catalog.ts | 163 | 7 дней | `global.7_дней` | ❌ |
| src/api/catalog.ts | 164 | 7 дней | `global.7_дней` | ❌ |
| src/api/catalog.ts | 164 | +15% ёмкость склада на всех планетах | `global.15_ёмкость_склада` | ❌ |
| src/api/catalog.ts | 164 | Атомный уплотнитель | `global.атомный_уплотнитель` | ✅ |
| src/api/catalog.ts | 165 | Суперкомпьютер | `global.суперкомпьютер` | ✅ |
| src/api/catalog.ts | 165 | +100% скорость исследования | `global.100_скорость_исследования` | ❌ |
| src/api/catalog.ts | 165 | 7 дней | `global.7_дней` | ❌ |
| src/api/catalog.ts | 166 | +100% скорость строительства (планета) | `global.100_скорость_строительства` | ❌ |
| src/api/catalog.ts | 166 | Система управления роботами | `global.система_управления_роботами` | ✅ |
| src/api/catalog.ts | 166 | 7 дней | `global.7_дней` | ❌ |
| src/api/catalog.ts | 198 | ${nameByKey(r.key)} ур.${r.level} | `global.namebykeyrkey_урrlevel` | ❌ |
| src/ui/Confirm.tsx | 14 | Подтверждение | `global.подтверждение` | ❌ |
| src/ui/Confirm.tsx | 16 | Подтвердить | `global.подтвердить` | ❌ |
| src/ui/Confirm.tsx | 17 | Отмена | `global.отмена` | ✅ |
| src/ui/Modal.tsx | 28 | Закрыть | `global.закрыть` | ❌ |
