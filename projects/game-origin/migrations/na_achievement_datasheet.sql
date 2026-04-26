-- phpMyAdmin SQL Dump
-- version 3.3.9.1
-- http://www.phpmyadmin.net
--
-- Хост: localhost
-- Время создания: Мар 15 2011 г., 17:23
-- Версия сервера: 5.1.54
-- Версия PHP: 5.3.5-0.dotdeb.0

SET SQL_MODE="NO_AUTO_VALUE_ON_ZERO";

--
-- База данных: `oxsar2-srv-01`
--

-- --------------------------------------------------------

--
-- Структура таблицы `na_achievement_datasheet`
--

CREATE TABLE IF NOT EXISTS `na_achievement_datasheet` (
  `achievement_id` int(10) unsigned NOT NULL,
  `req_points` int(10) unsigned NOT NULL DEFAULT '0',
  `req_u_points` int(10) unsigned NOT NULL DEFAULT '0',
  `req_r_points` int(10) unsigned NOT NULL DEFAULT '0',
  `req_b_points` int(10) unsigned NOT NULL DEFAULT '0',
  `req_u_count` int(10) unsigned NOT NULL DEFAULT '0',
  `req_r_count` int(10) unsigned NOT NULL DEFAULT '0',
  `req_b_count` int(10) unsigned NOT NULL DEFAULT '0',
  `req_e_points` int(10) unsigned NOT NULL DEFAULT '0',
  `req_be_points` int(10) unsigned NOT NULL DEFAULT '0',
  `req_of_points` int(10) unsigned NOT NULL DEFAULT '0',
  `req_of_level` int(10) unsigned NOT NULL DEFAULT '0',
  `req_battles` int(10) unsigned NOT NULL DEFAULT '0',
  `req_credit` int(10) unsigned NOT NULL DEFAULT '0',
  `req_a_points` int(10) unsigned NOT NULL DEFAULT '0',
  `req_a_count` int(10) unsigned NOT NULL DEFAULT '0',
  `bonus_metal` int(10) unsigned NOT NULL DEFAULT '0',
  `bonus_silicon` int(10) unsigned NOT NULL DEFAULT '0',
  `bonus_hydrogen` int(10) unsigned NOT NULL DEFAULT '0',
  `bonus_credit` int(10) unsigned NOT NULL DEFAULT '0',
  `bonus_1_unit_id` int(10) unsigned DEFAULT NULL,
  `bonus_1_unit_level` int(10) unsigned DEFAULT NULL,
  `bonus_2_unit_id` int(10) unsigned DEFAULT NULL,
  `bonus_2_unit_level` int(10) unsigned DEFAULT NULL,
  `bonus_3_unit_id` int(10) unsigned DEFAULT NULL,
  `bonus_3_unit_level` int(10) unsigned DEFAULT NULL,
  `image` varchar(255) DEFAULT NULL,
  `points` int(10) unsigned NOT NULL DEFAULT '0',
  `type` int(10) unsigned NOT NULL DEFAULT '0',
  `time` int(10) unsigned NOT NULL DEFAULT '31104000' COMMENT 'default - year',
  `custom_req_1` varchar(255) DEFAULT NULL,
  `custom_req_2` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`achievement_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

--
-- Дамп данных таблицы `na_achievement_datasheet`
--

INSERT INTO `na_achievement_datasheet` (`achievement_id`, `req_points`, `req_u_points`, `req_r_points`, `req_b_points`, `req_u_count`, `req_r_count`, `req_b_count`, `req_e_points`, `req_be_points`, `req_of_points`, `req_of_level`, `req_battles`, `req_credit`, `req_a_points`, `req_a_count`, `bonus_metal`, `bonus_silicon`, `bonus_hydrogen`, `bonus_credit`, `bonus_1_unit_id`, `bonus_1_unit_level`, `bonus_2_unit_id`, `bonus_2_unit_level`, `bonus_3_unit_id`, `bonus_3_unit_level`, `image`, `points`, `type`, `time`, `custom_req_1`, `custom_req_2`) VALUES
(330, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 100, 125, 150, 0, NULL, NULL, NULL, NULL, NULL, NULL, 'small', 0, 0, 31104000, NULL, NULL),
(331, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 75, 40, 35, 0, 1, 2, NULL, NULL, NULL, NULL, 'small', 0, 0, 31104000, NULL, NULL),
(332, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 50, 40, 35, 0, 2, 2, NULL, NULL, NULL, NULL, 'small', 0, 0, 31104000, NULL, NULL),
(333, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 500, 250, 150, 0, 3, 2, NULL, NULL, NULL, NULL, 'small', 0, 0, 31104000, NULL, NULL),
(334, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 100, 75, 45, 0, 4, 2, 39, 2, NULL, NULL, 'small', 0, 0, 31104000, NULL, NULL),
(335, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 400, 300, 200, 0, 6, 2, NULL, NULL, NULL, NULL, 'small', 0, 0, 31104000, NULL, NULL),
(336, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1000, 1000, 1000, 0, 8, 1, 12, 1, 101, 1, 'med', 0, 0, 31104000, NULL, NULL),
(337, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 32, 2, 0, 0, 0, 0, 'small', 0, 0, 31104000, NULL, NULL),
(338, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2000, 500, 500, 0, 31, 2, 29, 3, 0, 0, 'small', 0, 0, 31104000, NULL, NULL),
(339, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 30, 2, 0, 0, 0, 0, 'small', 0, 0, 31104000, NULL, NULL),
(340, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2000, 0, 0, 0, 44, 1, 43, 2, 0, 0, 'small', 0, 0, 31104000, NULL, NULL),
(341, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 500, 400, 100, 0, 29, 3, 20, 2, 0, 0, 'small', 0, 0, 31104000, NULL, NULL),
(342, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 45, 1, 0, 0, 0, 0, 'small', 0, 0, 31104000, NULL, NULL),
(343, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 5000, 3000, 2000, 0, 9, 1, 10, 1, 31, 3, 'med', 0, 0, 31104000, NULL, NULL),
(344, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 500, 0, NULL, NULL, NULL, NULL, NULL, NULL, 'small', 0, 0, 31104000, 'GotSim', NULL),
(345, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 600, 470, 100, 0, NULL, NULL, NULL, NULL, NULL, NULL, 'small', 0, 0, 31104000, 'ExchRes', NULL),
(346, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 100, 500, 700, 0, 8, 2, 300, 1, 111, 1, 'small', 0, 0, 31104000, NULL, NULL);

--
-- Ограничения внешнего ключа сохраненных таблиц
--

--
-- Ограничения внешнего ключа таблицы `na_achievement_datasheet`
--
ALTER TABLE `na_achievement_datasheet`
  ADD CONSTRAINT `na_achievement_datasheet_ibfk_1` FOREIGN KEY (`achievement_id`) REFERENCES `na_construction` (`buildingid`) ON DELETE CASCADE ON UPDATE CASCADE;
