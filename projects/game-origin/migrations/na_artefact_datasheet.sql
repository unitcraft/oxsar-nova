-- phpMyAdmin SQL Dump
-- version 3.3.9.1
-- http://www.phpmyadmin.net
--
-- Хост: localhost
-- Время создания: Мар 15 2011 г., 17:25
-- Версия сервера: 5.1.54
-- Версия PHP: 5.3.5-0.dotdeb.0

SET SQL_MODE="NO_AUTO_VALUE_ON_ZERO";

--
-- База данных: `oxsar2-srv-01`
--

-- --------------------------------------------------------

--
-- Структура таблицы `na_artefact_datasheet`
--

CREATE TABLE IF NOT EXISTS `na_artefact_datasheet` (
  `typeid` int(11) unsigned NOT NULL DEFAULT '0',
  `buyable` tinyint(1) unsigned NOT NULL DEFAULT '1',
  `auto_active` tinyint(1) unsigned NOT NULL DEFAULT '0',
  `movable` tinyint(1) unsigned NOT NULL DEFAULT '1',
  `unique` tinyint(1) unsigned NOT NULL DEFAULT '0',
  `usable` tinyint(1) unsigned NOT NULL DEFAULT '0',
  `trophy_chance` tinyint(1) unsigned NOT NULL DEFAULT '1',
  `delay` int(11) unsigned NOT NULL DEFAULT '0',
  `use_times` int(11) unsigned NOT NULL DEFAULT '0',
  `use_duration` int(11) unsigned NOT NULL DEFAULT '0',
  `lifetime` int(11) unsigned NOT NULL DEFAULT '0',
  `effect_type` tinyint(1) unsigned NOT NULL DEFAULT '0',
  `max_active` int(10) unsigned NOT NULL DEFAULT '0',
  `quota` float NOT NULL DEFAULT '1',
  PRIMARY KEY (`typeid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

--
-- Дамп данных таблицы `na_artefact_datasheet`
--

INSERT INTO `na_artefact_datasheet` (`typeid`, `buyable`, `auto_active`, `movable`, `unique`, `usable`, `trophy_chance`, `delay`, `use_times`, `use_duration`, `lifetime`, `effect_type`, `max_active`, `quota`) VALUES
(300, 1, 0, 1, 0, 1, 5, 0, 1, 604800, 2592000, 1, 1, 0.3),
(301, 1, 0, 1, 0, 1, 5, 0, 1, 604800, 2592000, 1, 5, 0.2),
(302, 1, 0, 1, 0, 1, 5, 0, 1, 604800, 2592000, 1, 5, 0.3),
(303, 1, 0, 1, 0, 1, 5, 0, 1, 604800, 2592000, 1, 5, 0.1),
(304, 1, 0, 1, 0, 0, 20, 3600, 10, 0, 2592000, 2, 0, 0.3),
(305, 1, 0, 1, 0, 1, 20, 3600, 2, 259200, 2592000, 1, 2, 0.08),
(315, 1, 0, 1, 0, 1, 20, 3600, 2, 604800, 2592000, 0, 2, 0.1),
(316, 1, 0, 1, 0, 0, 20, 3600, 5, 0, 2592000, 3, 0, 0.1),
(317, 1, 0, 1, 0, 0, 20, 3600, 5, 0, 2592000, 3, 0, 0.1),
(318, 1, 0, 1, 0, 0, 20, 3600, 5, 0, 2592000, 3, 0, 0.1),
(319, 1, 0, 1, 0, 1, 0, 0, 1, 0, 5184000, 0, 0, 0.01),
(320, 1, 0, 1, 0, 0, 0, 0, 1, 0, 5184000, 4, 0, 0.005),
(321, 1, 0, 1, 0, 1, 1, 0, 1, 0, 2592000, 0, 0, 0.01),
(322, 1, 0, 1, 0, 1, 1, 0, 1, 0, 2592000, 1, 0, 0.01),
(323, 1, 0, 1, 0, 0, 1, 0, 1, 0, 2592000, 0, 0, 0.05),
(324, 1, 0, 1, 0, 0, 1, 0, 1, 0, 2592000, 1, 0, 0.05);

--
-- Ограничения внешнего ключа сохраненных таблиц
--

--
-- Ограничения внешнего ключа таблицы `na_artefact_datasheet`
--
ALTER TABLE `na_artefact_datasheet`
  ADD CONSTRAINT `na_artefact_datasheet_ibfk_1` FOREIGN KEY (`typeid`) REFERENCES `na_construction` (`buildingid`) ON DELETE CASCADE ON UPDATE CASCADE;
