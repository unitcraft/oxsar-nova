-- Сид одной служебной записи в na_sim_planet (planetid=1).
--
-- Симулятор боя (Simulator.class.php) при формировании участников
-- (sim_assaultparticipant) ставит planetid = SIM_PLANET_ID = 1.
-- Эта запись — общий "контейнер" для всех симулируемых боёв; реальной
-- планетой не является, остальные поля заполнены минимально-валидно.
--
-- Без этой строки FOREIGN KEY na_sim_assaultparticipant.planetid →
-- na_sim_planet.planetid падает на каждой симуляции (подтверждено логами
-- 2026-04-30: "Cannot add or update a child row, FK ibfk_3"),
-- что приводило к пустым отчётам (Java получала assaultid без флотов).
--
-- Источник: oxsar2/sql/new-for-dm/data.sql строка 9838.

INSERT INTO `na_sim_planet`
  (`planetid`, `userid`, `ismoon`, `planetname`, `diameter`,
   `picture`, `temperature`, `last`, `metal`, `silicon`, `hydrogen`,
   `solar_satellite_prod`)
VALUES
  (1, NULL, 0, 'sim', 0, '0', 0, UNIX_TIMESTAMP(),
   500, 500, 0, 100)
ON DUPLICATE KEY UPDATE planetid=planetid;
