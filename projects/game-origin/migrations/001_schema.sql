
/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;
DROP TABLE IF EXISTS `na_achievement_datasheet`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_achievement_datasheet` (
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
  PRIMARY KEY (`achievement_id`),
  CONSTRAINT `na_achievement_datasheet_ibfk_1` FOREIGN KEY (`achievement_id`) REFERENCES `na_construction` (`buildingid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_achievements2user`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_achievements2user` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `user_id` int(10) unsigned NOT NULL,
  `achievement_id` int(10) unsigned NOT NULL,
  `created` int(10) unsigned NOT NULL,
  `granted` int(10) unsigned NOT NULL DEFAULT '0',
  `state` tinyint(3) unsigned NOT NULL DEFAULT '0',
  `quantity` int(10) unsigned NOT NULL DEFAULT '0',
  `granted_planet_id` int(10) unsigned DEFAULT NULL,
  `bonus_planet_id` int(10) unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `user_id_3` (`user_id`,`achievement_id`),
  KEY `achievement_id` (`achievement_id`),
  KEY `user_id` (`user_id`,`achievement_id`,`granted`),
  KEY `user_id_2` (`user_id`,`state`),
  KEY `user_id_4` (`user_id`,`achievement_id`,`state`,`granted`),
  KEY `user_id_5` (`user_id`,`granted`,`created`),
  CONSTRAINT `na_achievements2user_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `na_user` (`userid`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `na_achievements2user_ibfk_2` FOREIGN KEY (`achievement_id`) REFERENCES `na_achievement_datasheet` (`achievement_id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=2029055 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_alliance`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_alliance` (
  `aid` int(8) unsigned NOT NULL AUTO_INCREMENT,
  `tag` varchar(255) CHARACTER SET utf8 NOT NULL,
  `name` varchar(255) CHARACTER SET utf8 NOT NULL,
  `logo` varbinary(255) NOT NULL,
  `founder` int(10) unsigned NOT NULL,
  `foundername` varbinary(128) NOT NULL,
  `textextern` blob NOT NULL,
  `textintern` blob NOT NULL,
  `applicationtext` blob NOT NULL,
  `homepage` varbinary(255) NOT NULL,
  `showmember` tinyint(1) unsigned NOT NULL,
  `showhomepage` tinyint(1) unsigned NOT NULL,
  `memberlistsort` int(2) unsigned NOT NULL,
  `open` tinyint(1) unsigned NOT NULL,
  PRIMARY KEY (`aid`),
  UNIQUE KEY `tag` (`tag`),
  UNIQUE KEY `name` (`name`),
  KEY `founder` (`founder`),
  CONSTRAINT `na_alliance_ibfk_1` FOREIGN KEY (`founder`) REFERENCES `na_user` (`userid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=3170 DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_alliance_tmp`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_alliance_tmp` (
  `aid` int(8) unsigned NOT NULL AUTO_INCREMENT,
  `tag` varchar(255) CHARACTER SET utf8 NOT NULL,
  `name` varchar(255) CHARACTER SET utf8 NOT NULL,
  `logo` varbinary(255) NOT NULL,
  `founder` int(9) unsigned NOT NULL,
  `foundername` varbinary(128) NOT NULL,
  `textextern` blob NOT NULL,
  `textintern` blob NOT NULL,
  `applicationtext` blob NOT NULL,
  `homepage` varbinary(255) NOT NULL,
  `showmember` tinyint(1) unsigned NOT NULL,
  `showhomepage` tinyint(1) unsigned NOT NULL,
  `memberlistsort` int(2) unsigned NOT NULL,
  `open` tinyint(1) unsigned NOT NULL,
  PRIMARY KEY (`aid`),
  UNIQUE KEY `tag` (`tag`),
  UNIQUE KEY `name` (`name`),
  KEY `founder` (`founder`)
) ENGINE=InnoDB DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_ally_relationships`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_ally_relationships` (
  `relid` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `rel1` int(8) unsigned NOT NULL,
  `rel2` int(8) unsigned NOT NULL,
  `time` int(10) unsigned NOT NULL,
  `mode` tinyint(3) unsigned NOT NULL,
  PRIMARY KEY (`relid`),
  KEY `rel1` (`rel1`),
  KEY `rel2` (`rel2`),
  CONSTRAINT `na_ally_relationships_ibfk_1` FOREIGN KEY (`rel1`) REFERENCES `na_alliance` (`aid`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `na_ally_relationships_ibfk_2` FOREIGN KEY (`rel2`) REFERENCES `na_alliance` (`aid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=2565 DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_ally_relationships_application`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_ally_relationships_application` (
  `candidate_ally` int(8) unsigned NOT NULL,
  `request_ally` int(8) unsigned NOT NULL,
  `userid` int(9) unsigned NOT NULL,
  `mode` tinyint(2) NOT NULL,
  `application` blob NOT NULL,
  `time` int(10) unsigned NOT NULL,
  KEY `candidate_ally` (`candidate_ally`),
  KEY `request_ally` (`request_ally`),
  KEY `userid` (`userid`),
  CONSTRAINT `na_ally_relationships_application_ibfk_1` FOREIGN KEY (`candidate_ally`) REFERENCES `na_alliance` (`aid`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `na_ally_relationships_application_ibfk_2` FOREIGN KEY (`request_ally`) REFERENCES `na_alliance` (`aid`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `na_ally_relationships_application_ibfk_3` FOREIGN KEY (`userid`) REFERENCES `na_user` (`userid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_allyapplication`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_allyapplication` (
  `userid` int(9) unsigned NOT NULL,
  `aid` int(8) unsigned NOT NULL,
  `date` int(10) unsigned NOT NULL,
  `application` blob NOT NULL,
  KEY `userid` (`userid`),
  KEY `aid` (`aid`),
  CONSTRAINT `na_allyapplication_ibfk_1` FOREIGN KEY (`userid`) REFERENCES `na_user` (`userid`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `na_allyapplication_ibfk_2` FOREIGN KEY (`aid`) REFERENCES `na_alliance` (`aid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_allyrank`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_allyrank` (
  `rankid` int(12) unsigned NOT NULL AUTO_INCREMENT,
  `aid` int(8) unsigned NOT NULL,
  `name` varbinary(255) NOT NULL,
  `CAN_SEE_MEMBERLIST` tinyint(1) unsigned NOT NULL,
  `CAN_SEE_APPLICATIONS` tinyint(1) unsigned NOT NULL,
  `CAN_MANAGE` tinyint(1) unsigned NOT NULL,
  `CAN_BAN_MEMBER` tinyint(1) unsigned NOT NULL,
  `CAN_SEE_ONLINE_STATE` tinyint(1) unsigned NOT NULL,
  `CAN_WRITE_GLOBAL_MAILS` tinyint(1) unsigned NOT NULL,
  PRIMARY KEY (`rankid`),
  KEY `aid` (`aid`),
  CONSTRAINT `na_allyrank_ibfk_1` FOREIGN KEY (`aid`) REFERENCES `na_alliance` (`aid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=3847 DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_artefact2user`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_artefact2user` (
  `artid` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `typeid` int(10) unsigned NOT NULL DEFAULT '0',
  `userid` int(10) unsigned NOT NULL DEFAULT '0',
  `planetid` int(11) unsigned NOT NULL DEFAULT '0',
  `active` int(1) unsigned NOT NULL DEFAULT '0',
  `times_left` int(11) unsigned NOT NULL DEFAULT '0',
  `delay_eventid` int(11) unsigned NOT NULL DEFAULT '0',
  `expire_eventid` int(11) unsigned NOT NULL DEFAULT '0',
  `lifetime_eventid` int(11) unsigned NOT NULL DEFAULT '0',
  `transport_eventid` int(10) unsigned NOT NULL DEFAULT '0',
  `deleted` int(11) unsigned NOT NULL DEFAULT '0',
  `reason` varchar(64) DEFAULT NULL,
  `construction_id` int(11) unsigned NOT NULL DEFAULT '0',
  `level` int(11) unsigned NOT NULL DEFAULT '0',
  `bought` tinyint(1) NOT NULL DEFAULT '0',
  `lot_id` int(11) unsigned NOT NULL DEFAULT '0',
  PRIMARY KEY (`artid`),
  KEY `deleted` (`deleted`,`userid`,`planetid`),
  KEY `userid` (`userid`),
  KEY `planetid` (`planetid`),
  KEY `type` (`typeid`,`deleted`)
) ENGINE=InnoDB AUTO_INCREMENT=317904 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_artefact2user_tmp`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_artefact2user_tmp` (
  `artid` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `typeid` int(10) unsigned NOT NULL DEFAULT '0',
  `userid` int(10) unsigned NOT NULL DEFAULT '0',
  `planetid` int(11) unsigned NOT NULL DEFAULT '0',
  `active` int(1) unsigned NOT NULL DEFAULT '0',
  `times_left` int(11) unsigned NOT NULL DEFAULT '0',
  `delay_eventid` int(11) unsigned NOT NULL DEFAULT '0',
  `expire_eventid` int(11) unsigned NOT NULL DEFAULT '0',
  `lifetime_eventid` int(11) unsigned NOT NULL DEFAULT '0',
  `transport_eventid` int(10) unsigned NOT NULL DEFAULT '0',
  `deleted` int(11) unsigned NOT NULL DEFAULT '0',
  `reason` varchar(64) DEFAULT NULL,
  `construction_id` int(11) unsigned NOT NULL DEFAULT '0',
  `level` int(11) unsigned NOT NULL DEFAULT '0',
  `lot_id` int(11) unsigned NOT NULL DEFAULT '0',
  PRIMARY KEY (`artid`),
  KEY `deleted` (`deleted`,`userid`,`planetid`),
  KEY `userid` (`userid`),
  KEY `type` (`typeid`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_artefact_datasheet`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_artefact_datasheet` (
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
  PRIMARY KEY (`typeid`),
  CONSTRAINT `na_artefact_datasheet_ibfk_1` FOREIGN KEY (`typeid`) REFERENCES `na_construction` (`buildingid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_artefact_history`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_artefact_history` (
  `typeid` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `userid` int(10) unsigned NOT NULL DEFAULT '0',
  `assaultid` int(10) unsigned NOT NULL DEFAULT '0',
  `time` int(10) NOT NULL DEFAULT '0',
  PRIMARY KEY (`typeid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_artefact_probobility`;
/*!50001 DROP VIEW IF EXISTS `na_artefact_probobility`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_artefact_probobility` AS SELECT 
 1 AS `type`,
 1 AS `name`,
 1 AS `quota`,
 1 AS `art_count`,
 1 AS `exp_count`,
 1 AS `total_user`,
 1 AS `deleted_count`,
 1 AS `quota_count`,
 1 AS `probobility`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_artefact_used`;
/*!50001 DROP VIEW IF EXISTS `na_artefact_used`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_artefact_used` AS SELECT 
 1 AS `id`,
 1 AS `name`,
 1 AS `quota`,
 1 AS `all_count`,
 1 AS `buy_count`,
 1 AS `art_count`,
 1 AS `exp_count`,
 1 AS `total_user`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_assault`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_assault` (
  `assaultid` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `key` varbinary(4) NOT NULL,
  `key2` varbinary(4) NOT NULL,
  `result` tinyint(1) unsigned NOT NULL,
  `planetid` int(10) unsigned DEFAULT NULL,
  `time` int(10) unsigned NOT NULL,
  `target_moon` tinyint(1) unsigned NOT NULL DEFAULT '0',
  `target_buildingid` int(11) DEFAULT NULL,
  `building_level` smallint(11) DEFAULT NULL,
  `building_metal` double(15,0) DEFAULT NULL,
  `building_silicon` double(15,0) DEFAULT NULL,
  `building_hydrogen` double(15,0) DEFAULT NULL,
  `building_destroyed` tinyint(1) unsigned DEFAULT NULL,
  `target_destroyed` tinyint(1) unsigned NOT NULL DEFAULT '0',
  `attacker_explode` tinyint(1) unsigned NOT NULL DEFAULT '0',
  `moon_allow_type` tinyint(3) unsigned NOT NULL DEFAULT '0',
  `moonchance` tinyint(3) unsigned NOT NULL,
  `moon` tinyint(1) unsigned NOT NULL,
  `attacker_lost_res` double(15,0) NOT NULL,
  `attacker_lost_metal` double(15,0) NOT NULL,
  `attacker_lost_silicon` double(15,0) NOT NULL,
  `attacker_lost_hydrogen` double(15,0) NOT NULL,
  `defender_lost_res` double(15,0) NOT NULL,
  `defender_lost_metal` double(15,0) NOT NULL,
  `defender_lost_silicon` double(15,0) NOT NULL,
  `defender_lost_hydrogen` double(15,0) NOT NULL,
  `debris_metal` double(15,0) NOT NULL,
  `debris_silicon` double(15,0) NOT NULL,
  `planet_metal` double(15,0) DEFAULT NULL,
  `planet_silicon` double(15,0) DEFAULT NULL,
  `planet_hydrogen` double(15,0) DEFAULT NULL,
  `haul_metal` double(15,0) DEFAULT NULL,
  `haul_silicon` double(15,0) DEFAULT NULL,
  `haul_hydrogen` double(15,0) DEFAULT NULL,
  `lostunits_attacker` int(10) unsigned NOT NULL,
  `lostunits_defender` int(10) unsigned NOT NULL,
  `attacker_exp` float DEFAULT NULL,
  `defender_exp` float DEFAULT NULL,
  `turns` tinyint(3) unsigned DEFAULT NULL,
  `gentime` int(8) unsigned NOT NULL,
  `report` mediumblob NOT NULL,
  `accomplished` tinyint(1) unsigned NOT NULL,
  `message` blob,
  `advanced_system` tinyint(1) unsigned NOT NULL DEFAULT '0',
  PRIMARY KEY (`assaultid`),
  KEY `planetid` (`planetid`),
  KEY `time` (`time`),
  KEY `attacker_lost_res` (`attacker_lost_res`),
  KEY `defender_lost_res` (`defender_lost_res`),
  KEY `target_moon` (`target_moon`),
  CONSTRAINT `na_assault_ibfk_1` FOREIGN KEY (`planetid`) REFERENCES `na_planet` (`planetid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=2048388 DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_assault_ext`;
/*!50001 DROP VIEW IF EXISTS `na_assault_ext`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_assault_ext` AS SELECT 
 1 AS `t`,
 1 AS `g`,
 1 AS `s`,
 1 AS `p`,
 1 AS `assaultid`,
 1 AS `key`,
 1 AS `key2`,
 1 AS `result`,
 1 AS `planetid`,
 1 AS `time`,
 1 AS `target_moon`,
 1 AS `target_buildingid`,
 1 AS `building_level`,
 1 AS `building_metal`,
 1 AS `building_silicon`,
 1 AS `building_hydrogen`,
 1 AS `building_destroyed`,
 1 AS `target_destroyed`,
 1 AS `attacker_explode`,
 1 AS `moon_allow_type`,
 1 AS `moonchance`,
 1 AS `moon`,
 1 AS `attacker_lost_res`,
 1 AS `attacker_lost_metal`,
 1 AS `attacker_lost_silicon`,
 1 AS `attacker_lost_hydrogen`,
 1 AS `defender_lost_res`,
 1 AS `defender_lost_metal`,
 1 AS `defender_lost_silicon`,
 1 AS `defender_lost_hydrogen`,
 1 AS `debris_metal`,
 1 AS `debris_silicon`,
 1 AS `planet_metal`,
 1 AS `planet_silicon`,
 1 AS `planet_hydrogen`,
 1 AS `haul_metal`,
 1 AS `haul_silicon`,
 1 AS `haul_hydrogen`,
 1 AS `lostunits_attacker`,
 1 AS `lostunits_defender`,
 1 AS `attacker_exp`,
 1 AS `defender_exp`,
 1 AS `turns`,
 1 AS `gentime`,
 1 AS `report`,
 1 AS `accomplished`,
 1 AS `message`,
 1 AS `advanced_system`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_assault_ext2`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_assault_ext2` (
  `t` datetime DEFAULT NULL,
  `assaultid` int(10) unsigned DEFAULT NULL,
  `key` varbinary(4) DEFAULT NULL,
  `key2` varbinary(4) DEFAULT NULL,
  `result` tinyint(1) unsigned DEFAULT NULL,
  `planetid` int(10) unsigned DEFAULT NULL,
  `time` int(10) unsigned DEFAULT NULL,
  `moonchance` tinyint(3) unsigned DEFAULT NULL,
  `moon` tinyint(1) unsigned DEFAULT NULL,
  `attacker_lost_res` double(15,0) DEFAULT NULL,
  `attacker_lost_metal` double(15,0) DEFAULT NULL,
  `attacker_lost_silicon` double(15,0) DEFAULT NULL,
  `attacker_lost_hydrogen` double(15,0) DEFAULT NULL,
  `defender_lost_res` double(15,0) DEFAULT NULL,
  `defender_lost_metal` double(15,0) DEFAULT NULL,
  `defender_lost_silicon` double(15,0) DEFAULT NULL,
  `defender_lost_hydrogen` double(15,0) DEFAULT NULL,
  `debris_metal` double(15,0) DEFAULT NULL,
  `debris_silicon` double(15,0) DEFAULT NULL,
  `lostunits_attacker` int(10) unsigned DEFAULT NULL,
  `lostunits_defender` int(10) unsigned DEFAULT NULL,
  `attacker_exp` float DEFAULT NULL,
  `defender_exp` float DEFAULT NULL,
  `turns` tinyint(3) unsigned DEFAULT NULL,
  `gentime` int(8) unsigned DEFAULT NULL,
  `report` mediumblob,
  `accomplished` tinyint(1) unsigned DEFAULT NULL,
  `message` blob
) ENGINE=MyISAM DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_assault_stat`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_assault_stat` (
  `d` date DEFAULT NULL,
  `participants` decimal(41,0) DEFAULT NULL,
  `battles` bigint(21) DEFAULT NULL,
  `mp_sum` decimal(41,0) DEFAULT NULL,
  `mp_battles` decimal(23,0) DEFAULT NULL
) ENGINE=MyISAM DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_assaultparticipant`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_assaultparticipant` (
  `participantid` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `assaultid` int(10) unsigned NOT NULL,
  `userid` int(10) unsigned DEFAULT NULL,
  `planetid` int(10) unsigned DEFAULT NULL,
  `mode` tinyint(1) unsigned NOT NULL,
  `consumption` double(15,0) DEFAULT NULL,
  `preloaded` double(15,0) DEFAULT NULL,
  `capacity` double(15,0) DEFAULT NULL,
  `haul_metal` double(15,0) NOT NULL,
  `haul_silicon` double(15,0) NOT NULL,
  `haul_hydrogen` double(15,0) NOT NULL,
  `target_unitid` smallint(5) unsigned NOT NULL DEFAULT '0',
  `add_gun_tech` tinyint(3) unsigned NOT NULL DEFAULT '0',
  `add_shield_tech` tinyint(3) unsigned NOT NULL DEFAULT '0',
  `add_shell_tech` tinyint(3) unsigned NOT NULL DEFAULT '0',
  `add_ballistics_tech` tinyint(3) unsigned NOT NULL DEFAULT '0',
  `add_masking_tech` tinyint(3) unsigned NOT NULL DEFAULT '0',
  `add_laser_tech` tinyint(3) unsigned NOT NULL DEFAULT '0',
  `add_ion_tech` tinyint(3) unsigned NOT NULL DEFAULT '0',
  `add_plasma_tech` tinyint(3) unsigned NOT NULL DEFAULT '0',
  PRIMARY KEY (`participantid`),
  KEY `assaultid` (`assaultid`),
  KEY `userid` (`userid`),
  KEY `planetid` (`planetid`),
  CONSTRAINT `na_assaultparticipant_ibfk_1` FOREIGN KEY (`assaultid`) REFERENCES `na_assault` (`assaultid`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `na_assaultparticipant_ibfk_2` FOREIGN KEY (`userid`) REFERENCES `na_user` (`userid`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `na_assaultparticipant_ibfk_3` FOREIGN KEY (`planetid`) REFERENCES `na_planet` (`planetid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=4779942 DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_attack_formation`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_attack_formation` (
  `eventid` int(10) unsigned NOT NULL,
  `name` varbinary(128) NOT NULL,
  `time` int(10) unsigned NOT NULL,
  PRIMARY KEY (`eventid`)
) ENGINE=InnoDB DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_ban`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_ban` (
  `banid` int(8) unsigned DEFAULT NULL,
  `ipaddress` varbinary(40) DEFAULT NULL,
  `reason` varbinary(255) DEFAULT NULL,
  `timebegin` int(10) unsigned DEFAULT NULL,
  `timeend` int(10) unsigned DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_ban_u`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_ban_u` (
  `banid` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `userid` int(9) unsigned NOT NULL,
  `from` int(10) unsigned NOT NULL,
  `to` int(10) unsigned DEFAULT NULL,
  `reason` varchar(255) NOT NULL,
  `admin_comment` varchar(2000) DEFAULT NULL,
  PRIMARY KEY (`banid`),
  UNIQUE KEY `userid` (`userid`),
  CONSTRAINT `na_ban_u_ibfk_1` FOREIGN KEY (`userid`) REFERENCES `na_user` (`userid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_ban_u_ext`;
/*!50001 DROP VIEW IF EXISTS `na_ban_u_ext`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_ban_u_ext` AS SELECT 
 1 AS `r`,
 1 AS `username`,
 1 AS `banid`,
 1 AS `userid`,
 1 AS `from`,
 1 AS `to`,
 1 AS `reason`,
 1 AS `admin_comment`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_buddylist`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_buddylist` (
  `relid` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `friend1` int(9) unsigned NOT NULL,
  `friend2` int(9) unsigned NOT NULL,
  `accepted` tinyint(4) unsigned NOT NULL,
  PRIMARY KEY (`relid`),
  KEY `friend2` (`friend2`),
  KEY `friend1` (`friend1`),
  CONSTRAINT `na_buddylist_ibfk_1` FOREIGN KEY (`friend1`) REFERENCES `na_user` (`userid`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `na_buddylist_ibfk_2` FOREIGN KEY (`friend2`) REFERENCES `na_user` (`userid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=35291 DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_building2planet`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_building2planet` (
  `planetid` int(10) unsigned NOT NULL,
  `buildingid` int(4) unsigned NOT NULL,
  `level` int(3) unsigned NOT NULL,
  `added` int(3) NOT NULL DEFAULT '0',
  `prod_factor` int(3) unsigned NOT NULL DEFAULT '100',
  PRIMARY KEY (`planetid`,`buildingid`),
  KEY `buildingid` (`buildingid`),
  CONSTRAINT `na_building2planet_ibfk_1` FOREIGN KEY (`planetid`) REFERENCES `na_planet` (`planetid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_building2planet_tmp`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_building2planet_tmp` (
  `planetid` int(10) unsigned NOT NULL,
  `buildingid` int(4) unsigned NOT NULL,
  `level` int(3) unsigned NOT NULL,
  `added` int(3) unsigned NOT NULL DEFAULT '0',
  `prod_factor` int(3) unsigned NOT NULL DEFAULT '100',
  PRIMARY KEY (`planetid`,`buildingid`),
  KEY `buildingid` (`buildingid`)
) ENGINE=MyISAM DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_building2planet_tmp2`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_building2planet_tmp2` (
  `planetid` int(10) unsigned NOT NULL,
  `buildingid` int(4) unsigned NOT NULL,
  `level` int(3) unsigned NOT NULL,
  `added` int(3) unsigned NOT NULL DEFAULT '0',
  `prod_factor` int(3) unsigned NOT NULL DEFAULT '100',
  PRIMARY KEY (`planetid`,`buildingid`),
  KEY `buildingid` (`buildingid`)
) ENGINE=MyISAM DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_chat`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_chat` (
  `messageid` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `time` int(10) unsigned DEFAULT NULL,
  `userid` int(10) unsigned DEFAULT NULL,
  `tinymce` tinyint(1) NOT NULL DEFAULT '0',
  `message` text,
  PRIMARY KEY (`messageid`),
  KEY `time` (`time`),
  KEY `userid` (`userid`,`messageid`)
) ENGINE=InnoDB AUTO_INCREMENT=1283133 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_chat2ally`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_chat2ally` (
  `messageid` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `time` int(10) unsigned NOT NULL,
  `userid` int(10) unsigned NOT NULL,
  `allyid` int(10) unsigned NOT NULL,
  `tinymce` tinyint(1) NOT NULL DEFAULT '0',
  `message` text,
  PRIMARY KEY (`messageid`),
  KEY `allyid` (`allyid`),
  KEY `userid` (`userid`,`messageid`)
) ENGINE=InnoDB AUTO_INCREMENT=2244760 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_chat2ally_ext`;
/*!50001 DROP VIEW IF EXISTS `na_chat2ally_ext`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_chat2ally_ext` AS SELECT 
 1 AS `t`,
 1 AS `username`,
 1 AS `messageid`,
 1 AS `time`,
 1 AS `userid`,
 1 AS `allyid`,
 1 AS `message`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_chat2ally_stat`;
/*!50001 DROP VIEW IF EXISTS `na_chat2ally_stat`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_chat2ally_stat` AS SELECT 
 1 AS `allyid`,
 1 AS `cnt`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_chat_ext`;
/*!50001 DROP VIEW IF EXISTS `na_chat_ext`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_chat_ext` AS SELECT 
 1 AS `t`,
 1 AS `username`,
 1 AS `messageid`,
 1 AS `time`,
 1 AS `userid`,
 1 AS `message`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_chat_ro`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_chat_ro` (
  `roid` int(8) unsigned NOT NULL AUTO_INCREMENT,
  `ipaddress` varbinary(40) NOT NULL,
  `reason` varbinary(255) NOT NULL,
  `timebegin` int(10) unsigned NOT NULL,
  `timeend` int(10) unsigned NOT NULL,
  PRIMARY KEY (`roid`)
) ENGINE=InnoDB DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_chat_ro_u`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_chat_ro_u` (
  `roid` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `userid` int(10) unsigned NOT NULL,
  `from` int(10) unsigned NOT NULL,
  `to` int(10) unsigned NOT NULL,
  `reason` varbinary(255) NOT NULL,
  PRIMARY KEY (`roid`),
  KEY `userid` (`userid`),
  CONSTRAINT `na_chat_ro_u_ibfk_1` FOREIGN KEY (`userid`) REFERENCES `na_user` (`userid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_chat_tmp`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_chat_tmp` (
  `messageid` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `time` int(10) unsigned DEFAULT NULL,
  `userid` int(10) unsigned DEFAULT NULL,
  `message` blob,
  PRIMARY KEY (`messageid`),
  KEY `userid` (`userid`),
  KEY `time` (`time`)
) ENGINE=MyISAM AUTO_INCREMENT=41615 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_config`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_config` (
  `var` varbinary(255) NOT NULL,
  `value` varbinary(10000) NOT NULL,
  `type` varbinary(10) NOT NULL,
  `description` varbinary(5000) NOT NULL,
  `options` varbinary(1000) NOT NULL,
  `groupid` int(4) unsigned NOT NULL,
  `islisted` tinyint(1) unsigned NOT NULL,
  PRIMARY KEY (`var`),
  KEY `groupid` (`groupid`)
) ENGINE=InnoDB DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_configgroups`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_configgroups` (
  `groupid` int(4) unsigned NOT NULL AUTO_INCREMENT,
  `groupname` varbinary(64) NOT NULL,
  PRIMARY KEY (`groupid`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_construction`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_construction` (
  `buildingid` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `race` tinyint(3) unsigned NOT NULL DEFAULT '1',
  `mode` tinyint(3) unsigned NOT NULL,
  `name` varbinary(255) NOT NULL,
  `test` tinyint(1) unsigned NOT NULL DEFAULT '0',
  `front` tinyint(3) unsigned NOT NULL DEFAULT '10',
  `ballistics` tinyint(3) unsigned NOT NULL DEFAULT '0',
  `masking` tinyint(3) unsigned NOT NULL DEFAULT '0',
  `basic_metal` double(15,0) unsigned NOT NULL,
  `basic_silicon` double(15,0) unsigned NOT NULL,
  `basic_hydrogen` double(15,0) unsigned NOT NULL,
  `basic_energy` double(15,0) unsigned NOT NULL,
  `basic_credit` int(10) unsigned NOT NULL DEFAULT '0',
  `basic_points` int(10) unsigned NOT NULL DEFAULT '0',
  `prod_metal` varbinary(255) NOT NULL,
  `prod_silicon` varbinary(255) NOT NULL,
  `prod_hydrogen` varbinary(255) NOT NULL,
  `prod_energy` varbinary(255) NOT NULL,
  `cons_metal` varbinary(255) NOT NULL,
  `cons_silicon` varbinary(255) NOT NULL,
  `cons_hydrogen` varbinary(255) NOT NULL,
  `cons_energy` varbinary(255) NOT NULL,
  `charge_metal` varbinary(255) NOT NULL,
  `charge_silicon` varbinary(255) NOT NULL,
  `charge_hydrogen` varbinary(255) NOT NULL,
  `charge_energy` varbinary(255) NOT NULL,
  `charge_credit` varbinary(255) NOT NULL DEFAULT '',
  `charge_points` varbinary(255) NOT NULL,
  `special` varbinary(255) NOT NULL,
  `demolish` float NOT NULL,
  `display_order` int(10) unsigned NOT NULL,
  PRIMARY KEY (`buildingid`),
  KEY `mode` (`mode`,`test`),
  KEY `display_order` (`display_order`,`buildingid`)
) ENGINE=InnoDB AUTO_INCREMENT=366 DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_credit_bonus_item`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_credit_bonus_item` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `unitid` int(10) unsigned NOT NULL,
  `userid` int(10) unsigned NOT NULL,
  `date` datetime NOT NULL,
  `credit` float(11,2) NOT NULL,
  `done` tinyint(1) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `userid` (`userid`,`done`,`date`)
) ENGINE=InnoDB AUTO_INCREMENT=96 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_cronjob`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_cronjob` (
  `cronid` int(5) unsigned NOT NULL AUTO_INCREMENT,
  `script` varbinary(255) NOT NULL,
  `month` varbinary(26) NOT NULL,
  `day` varbinary(83) NOT NULL,
  `weekday` varbinary(13) NOT NULL,
  `hour` varbinary(62) NOT NULL,
  `minute` varbinary(34) NOT NULL,
  `xtime` int(10) unsigned NOT NULL,
  `last` int(10) unsigned NOT NULL,
  `active` tinyint(1) unsigned NOT NULL,
  PRIMARY KEY (`cronid`),
  KEY `xtime` (`xtime`,`active`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_cronjob_ext`;
/*!50001 DROP VIEW IF EXISTS `na_cronjob_ext`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_cronjob_ext` AS SELECT 
 1 AS `x`,
 1 AS `l`,
 1 AS `cronid`,
 1 AS `script`,
 1 AS `month`,
 1 AS `day`,
 1 AS `weekday`,
 1 AS `hour`,
 1 AS `minute`,
 1 AS `xtime`,
 1 AS `last`,
 1 AS `active`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_empty_systems`;
/*!50001 DROP VIEW IF EXISTS `na_empty_systems`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_empty_systems` AS SELECT 
 1 AS `galaxy`,
 1 AS `system`,
 1 AS `cnt`,
 1 AS `free_pos`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_engine`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_engine` (
  `engineid` int(4) unsigned NOT NULL,
  `factor` int(3) unsigned NOT NULL,
  PRIMARY KEY (`engineid`)
) ENGINE=InnoDB DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_event_aliens`;
/*!50001 DROP VIEW IF EXISTS `na_event_aliens`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_event_aliens` AS SELECT 
 1 AS `s`,
 1 AS `t`,
 1 AS `p`,
 1 AS `eventid`,
 1 AS `mode`,
 1 AS `start`,
 1 AS `time`,
 1 AS `planetid`,
 1 AS `user`,
 1 AS `destination`,
 1 AS `data`,
 1 AS `protected`,
 1 AS `prev_rc`,
 1 AS `processed`,
 1 AS `processed_mode`,
 1 AS `processed_time`,
 1 AS `processed_dt`,
 1 AS `processed_quantity`,
 1 AS `required_quantity`,
 1 AS `error_message`,
 1 AS `ally_eventid`,
 1 AS `parent_eventid`,
 1 AS `org_data`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_event_dest`;
/*!50001 DROP VIEW IF EXISTS `na_event_dest`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_event_dest` AS SELECT 
 1 AS `eventid`,
 1 AS `mode`,
 1 AS `start`,
 1 AS `time`,
 1 AS `planetid`,
 1 AS `user`,
 1 AS `destination`,
 1 AS `data`,
 1 AS `protected`,
 1 AS `prev_rc`,
 1 AS `processed`,
 1 AS `processed_mode`,
 1 AS `processed_time`,
 1 AS `processed_dt`,
 1 AS `processed_quantity`,
 1 AS `required_quantity`,
 1 AS `error_message`,
 1 AS `ally_eventid`,
 1 AS `parent_eventid`,
 1 AS `artid`,
 1 AS `org_data`,
 1 AS `startuserid`,
 1 AS `startusername`,
 1 AS `userid`,
 1 AS `username`,
 1 AS `planetname`,
 1 AS `galaxy`,
 1 AS `system`,
 1 AS `position`,
 1 AS `destuserid`,
 1 AS `destname`,
 1 AS `destplanet`,
 1 AS `galaxy2`,
 1 AS `system2`,
 1 AS `position2`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_event_src`;
/*!50001 DROP VIEW IF EXISTS `na_event_src`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_event_src` AS SELECT 
 1 AS `eventid`,
 1 AS `mode`,
 1 AS `start`,
 1 AS `time`,
 1 AS `planetid`,
 1 AS `user`,
 1 AS `destination`,
 1 AS `data`,
 1 AS `protected`,
 1 AS `prev_rc`,
 1 AS `processed`,
 1 AS `processed_mode`,
 1 AS `processed_time`,
 1 AS `processed_dt`,
 1 AS `processed_quantity`,
 1 AS `required_quantity`,
 1 AS `error_message`,
 1 AS `ally_eventid`,
 1 AS `parent_eventid`,
 1 AS `artid`,
 1 AS `org_data`,
 1 AS `startuserid`,
 1 AS `startusername`,
 1 AS `userid`,
 1 AS `username`,
 1 AS `planetname`,
 1 AS `galaxy`,
 1 AS `system`,
 1 AS `position`,
 1 AS `destuserid`,
 1 AS `destname`,
 1 AS `destplanet`,
 1 AS `galaxy2`,
 1 AS `system2`,
 1 AS `position2`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_events`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_events` (
  `eventid` int(10) NOT NULL AUTO_INCREMENT,
  `mode` int(2) unsigned NOT NULL,
  `start` int(10) unsigned NOT NULL,
  `time` int(10) unsigned NOT NULL,
  `planetid` int(10) unsigned DEFAULT NULL,
  `user` int(9) NOT NULL,
  `destination` int(10) unsigned DEFAULT NULL,
  `data` mediumblob NOT NULL,
  `protected` tinyint(1) unsigned NOT NULL DEFAULT '0',
  `prev_rc` varbinary(16) DEFAULT NULL,
  `processed` tinyint(4) NOT NULL DEFAULT '0',
  `processed_mode` int(11) DEFAULT NULL,
  `processed_time` int(11) DEFAULT NULL,
  `processed_dt` float NOT NULL DEFAULT '0',
  `processed_quantity` int(11) DEFAULT NULL,
  `required_quantity` int(11) DEFAULT NULL,
  `error_message` mediumblob,
  `ally_eventid` int(11) DEFAULT NULL,
  `parent_eventid` int(11) DEFAULT NULL,
  `artid` int(11) DEFAULT NULL,
  `org_data` mediumblob,
  PRIMARY KEY (`eventid`),
  KEY `planetid` (`planetid`),
  KEY `user` (`user`),
  KEY `prev_rc` (`prev_rc`),
  KEY `processed` (`processed`,`time`),
  KEY `time` (`time`,`processed`,`prev_rc`),
  KEY `effect_time` (`processed`,`prev_rc`,`time`),
  KEY `ally_eventid` (`ally_eventid`,`mode`),
  KEY `parent_eventid` (`parent_eventid`),
  KEY `destination` (`destination`,`processed`,`mode`,`time`)
) ENGINE=InnoDB AUTO_INCREMENT=78785189 DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_events_ext`;
/*!50001 DROP VIEW IF EXISTS `na_events_ext`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_events_ext` AS SELECT 
 1 AS `s`,
 1 AS `t`,
 1 AS `p`,
 1 AS `eventid`,
 1 AS `mode`,
 1 AS `start`,
 1 AS `time`,
 1 AS `planetid`,
 1 AS `user`,
 1 AS `destination`,
 1 AS `data`,
 1 AS `protected`,
 1 AS `prev_rc`,
 1 AS `processed`,
 1 AS `processed_mode`,
 1 AS `processed_time`,
 1 AS `processed_dt`,
 1 AS `processed_quantity`,
 1 AS `required_quantity`,
 1 AS `error_message`,
 1 AS `ally_eventid`,
 1 AS `parent_eventid`,
 1 AS `org_data`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_events_stat`;
/*!50001 DROP VIEW IF EXISTS `na_events_stat`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_events_stat` AS SELECT 
 1 AS `d`,
 1 AS `mode`,
 1 AS `cnt`,
 1 AS `qnt`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_events_tmp`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_events_tmp` (
  `eventid` int(10) NOT NULL AUTO_INCREMENT,
  `mode` int(2) unsigned NOT NULL,
  `start` int(10) unsigned NOT NULL,
  `time` int(10) unsigned NOT NULL,
  `planetid` int(10) unsigned DEFAULT NULL,
  `user` int(9) NOT NULL,
  `destination` int(10) unsigned DEFAULT NULL,
  `data` mediumblob NOT NULL,
  `protected` tinyint(1) unsigned NOT NULL DEFAULT '0',
  `prev_rc` varbinary(16) DEFAULT NULL,
  `processed` tinyint(4) NOT NULL DEFAULT '0',
  `processed_mode` int(11) DEFAULT NULL,
  `processed_time` int(11) DEFAULT NULL,
  `processed_dt` float NOT NULL DEFAULT '0',
  `processed_quantity` int(11) DEFAULT NULL,
  `required_quantity` int(11) DEFAULT NULL,
  `error_message` tinyblob,
  `ally_eventid` int(11) DEFAULT NULL,
  `parent_eventid` int(11) DEFAULT NULL,
  `artid` int(11) DEFAULT NULL,
  `org_data` mediumblob,
  PRIMARY KEY (`eventid`),
  KEY `planetid` (`planetid`),
  KEY `user` (`user`),
  KEY `prev_rc` (`prev_rc`),
  KEY `processed` (`processed`,`time`),
  KEY `time` (`time`,`processed`,`prev_rc`),
  KEY `effect_time` (`processed`,`prev_rc`,`time`),
  KEY `ally_eventid` (`ally_eventid`,`mode`),
  KEY `parent_eventid` (`parent_eventid`),
  KEY `destination` (`destination`,`processed`,`mode`,`time`)
) ENGINE=MyISAM DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_exchange`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_exchange` (
  `eid` int(10) unsigned NOT NULL,
  `uid` int(10) unsigned NOT NULL,
  `title` varchar(50) NOT NULL,
  `fee` int(11) NOT NULL,
  `def_fee` int(11) NOT NULL,
  `comission` int(11) NOT NULL,
  `featured_time` int(10) unsigned DEFAULT NULL,
  PRIMARY KEY (`eid`),
  KEY `uid` (`uid`),
  KEY `featured_time` (`featured_time`),
  CONSTRAINT `na_exchange_ibfk_1` FOREIGN KEY (`eid`) REFERENCES `na_planet` (`planetid`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `na_exchange_ibfk_2` FOREIGN KEY (`uid`) REFERENCES `na_user` (`userid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_exchange_lots`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_exchange_lots` (
  `lid` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `planetid` int(10) unsigned NOT NULL,
  `brokerid` int(10) unsigned NOT NULL COMMENT 'exchange_id = planet_id of the exchange',
  `buyerid` int(10) unsigned DEFAULT NULL,
  `buyerplanet` int(10) unsigned DEFAULT NULL,
  `raising_date` int(10) unsigned NOT NULL,
  `sold_date` int(10) unsigned DEFAULT NULL,
  `expiry_date` int(10) unsigned NOT NULL,
  `featured_date` int(10) unsigned DEFAULT NULL,
  `delivery_hydro` int(10) unsigned NOT NULL DEFAULT '0',
  `delivery_percent` tinyint(10) unsigned NOT NULL DEFAULT '0',
  `data` mediumblob NOT NULL,
  `type` int(11) NOT NULL,
  `lot` int(10) NOT NULL,
  `amount` double(15,0) unsigned NOT NULL,
  `lot_min_amount` double(15,0) unsigned NOT NULL DEFAULT '1',
  `price` double(15,2) NOT NULL,
  `fee` tinyint(4) NOT NULL,
  `ally_discount` tinyint(4) NOT NULL DEFAULT '0',
  `status` int(11) NOT NULL,
  `lot_amount` double(15,0) unsigned NOT NULL DEFAULT '0' COMMENT 'Amount in lot at the very begining',
  `lot_price` double(15,2) unsigned NOT NULL DEFAULT '0.00' COMMENT 'Price at the very begining',
  `lot_unit_price` double(15,5) DEFAULT NULL,
  `lot_parent_id` int(10) unsigned DEFAULT NULL COMMENT 'Parent id',
  `payed_seller` double(15,2) unsigned NOT NULL DEFAULT '0.00' COMMENT 'Payed to seller.',
  `payed_buyer` double(15,2) unsigned NOT NULL DEFAULT '0.00' COMMENT 'Payed to buyer.',
  `payed_exchange` double(15,2) unsigned NOT NULL DEFAULT '0.00' COMMENT 'Payed to exchange(comission).',
  `payed_fuel` double(15,2) unsigned NOT NULL DEFAULT '0.00' COMMENT 'Payed for fuel.',
  `used_fuel` double(15,0) unsigned NOT NULL DEFAULT '0' COMMENT 'Fuel from this lot used to delivery this lot.',
  `rest_fuel` double(15,0) unsigned NOT NULL DEFAULT '0' COMMENT 'Start fuel - used fuel.',
  PRIMARY KEY (`lid`),
  KEY `sellerid` (`planetid`,`brokerid`,`type`),
  KEY `buyerid` (`buyerid`),
  KEY `buyerplanet` (`buyerplanet`),
  KEY `planetid` (`status`,`planetid`),
  KEY `status` (`status`,`brokerid`),
  KEY `status_2` (`status`,`featured_date`),
  KEY `brokerid` (`brokerid`,`sold_date`,`status`),
  KEY `lot_unit_price` (`lot_unit_price`)
) ENGINE=InnoDB AUTO_INCREMENT=574821 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_exchange_stats`;
/*!50001 DROP VIEW IF EXISTS `na_exchange_stats`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_exchange_stats` AS SELECT 
 1 AS `d`,
 1 AS `price`,
 1 AS `payed_exchange`,
 1 AS `cnt`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_exchange_tmp`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_exchange_tmp` (
  `eid` int(10) unsigned NOT NULL,
  `uid` int(10) unsigned NOT NULL,
  `title` varchar(50) NOT NULL,
  `fee` int(11) NOT NULL,
  `def_fee` int(11) NOT NULL,
  `comission` int(11) NOT NULL,
  PRIMARY KEY (`eid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_expedition_found_units`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_expedition_found_units` (
  `unit_id` mediumint(8) unsigned NOT NULL,
  `expedition_id` int(11) NOT NULL,
  `quantity` int(11) DEFAULT NULL,
  UNIQUE KEY `unit_id` (`unit_id`,`expedition_id`),
  KEY `expedition_id` (`expedition_id`),
  CONSTRAINT `na_expedition_found_units_ibfk_1` FOREIGN KEY (`expedition_id`) REFERENCES `na_expedition_stats` (`statid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_expedition_log`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_expedition_log` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `type_id` int(10) unsigned NOT NULL DEFAULT '0',
  `user_id` int(10) unsigned NOT NULL DEFAULT '0',
  `time` int(10) unsigned NOT NULL DEFAULT '0',
  `data` mediumblob NOT NULL,
  PRIMARY KEY (`id`),
  KEY `user` (`user_id`)
) ENGINE=InnoDB AUTO_INCREMENT=664 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_expedition_stats`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_expedition_stats` (
  `statid` int(11) NOT NULL AUTO_INCREMENT,
  `userid` int(10) unsigned NOT NULL,
  `time` int(10) unsigned NOT NULL,
  `completed` tinyint(4) NOT NULL DEFAULT '0',
  `galaxy` int(3) unsigned NOT NULL,
  `system` int(4) unsigned NOT NULL,
  `type` varchar(100) DEFAULT NULL,
  `percent` float NOT NULL,
  `message` mediumtext,
  `artefact_type` int(10) unsigned NOT NULL DEFAULT '0',
  `found_credit` double(15,2) unsigned DEFAULT NULL,
  `found_metal` double(15,2) unsigned DEFAULT NULL,
  `found_silicon` double(15,2) unsigned DEFAULT NULL,
  `found_hydrogen` double(15,2) unsigned DEFAULT NULL,
  `event_id` int(11) unsigned DEFAULT NULL,
  PRIMARY KEY (`statid`),
  KEY `time` (`time`,`system`,`galaxy`),
  KEY `artefact_type` (`artefact_type`,`time`)
) ENGINE=InnoDB AUTO_INCREMENT=4264105 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_expedition_stats_day`;
/*!50001 DROP VIEW IF EXISTS `na_expedition_stats_day`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_expedition_stats_day` AS SELECT 
 1 AS `d`,
 1 AS `type`,
 1 AS `cnt`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_expedition_stats_ext`;
/*!50001 DROP VIEW IF EXISTS `na_expedition_stats_ext`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_expedition_stats_ext` AS SELECT 
 1 AS `t`,
 1 AS `username`,
 1 AS `art_name`,
 1 AS `statid`,
 1 AS `userid`,
 1 AS `time`,
 1 AS `completed`,
 1 AS `galaxy`,
 1 AS `system`,
 1 AS `type`,
 1 AS `percent`,
 1 AS `message`,
 1 AS `artefact_type`,
 1 AS `found_credit`,
 1 AS `found_metal`,
 1 AS `found_silicon`,
 1 AS `found_hydrogen`,
 1 AS `event_id`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_expedition_stats_old2`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_expedition_stats_old2` (
  `statid` int(11) NOT NULL AUTO_INCREMENT,
  `userid` int(10) unsigned NOT NULL,
  `time` int(10) unsigned NOT NULL,
  `completed` tinyint(4) NOT NULL DEFAULT '0',
  `galaxy` int(3) unsigned NOT NULL,
  `system` int(4) unsigned NOT NULL,
  `type` varchar(100) DEFAULT NULL,
  `percent` float NOT NULL,
  `message` mediumtext,
  `artefact_type` int(10) unsigned NOT NULL DEFAULT '0',
  `event_id` int(11) unsigned DEFAULT NULL,
  PRIMARY KEY (`statid`),
  KEY `time` (`time`,`system`,`galaxy`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_expedition_stats_old3`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_expedition_stats_old3` (
  `statid` int(11) NOT NULL AUTO_INCREMENT,
  `userid` int(10) unsigned NOT NULL,
  `time` int(10) unsigned NOT NULL,
  `completed` tinyint(4) NOT NULL DEFAULT '0',
  `galaxy` int(3) unsigned NOT NULL,
  `system` int(4) unsigned NOT NULL,
  `type` varchar(100) DEFAULT NULL,
  `percent` float NOT NULL,
  `message` mediumtext,
  `artefact_type` int(10) unsigned NOT NULL DEFAULT '0',
  `event_id` int(11) unsigned DEFAULT NULL,
  PRIMARY KEY (`statid`),
  KEY `time` (`time`,`system`,`galaxy`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_expedition_stats_olddata`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_expedition_stats_olddata` (
  `statid` int(11) NOT NULL AUTO_INCREMENT,
  `userid` int(10) unsigned NOT NULL,
  `time` int(10) unsigned NOT NULL,
  `completed` tinyint(4) NOT NULL DEFAULT '0',
  `galaxy` int(3) unsigned NOT NULL,
  `system` int(4) unsigned NOT NULL,
  `type` varchar(100) DEFAULT NULL,
  `percent` float NOT NULL,
  `message` mediumtext,
  PRIMARY KEY (`statid`),
  KEY `time` (`time`,`system`,`galaxy`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_expedition_used`;
/*!50001 DROP VIEW IF EXISTS `na_expedition_used`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_expedition_used` AS SELECT 
 1 AS `d`,
 1 AS `type`,
 1 AS `cnt`,
 1 AS `percent`,
 1 AS `avg_percent`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_fleet2assault`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_fleet2assault` (
  `assaultid` int(10) unsigned NOT NULL,
  `participantid` int(10) unsigned NOT NULL,
  `userid` int(10) unsigned DEFAULT NULL,
  `unitid` smallint(5) unsigned NOT NULL,
  `mode` tinyint(3) unsigned NOT NULL,
  `quantity` int(10) unsigned NOT NULL,
  `damaged` int(10) unsigned NOT NULL DEFAULT '0',
  `shell_percent` tinyint(3) unsigned NOT NULL DEFAULT '0',
  `grasped` int(10) unsigned NOT NULL DEFAULT '0',
  `org_quantity` int(10) unsigned NOT NULL DEFAULT '0',
  `org_damaged` int(10) unsigned NOT NULL DEFAULT '0',
  `org_shell_percent` tinyint(3) unsigned NOT NULL DEFAULT '0',
  `etc` blob,
  UNIQUE KEY `participantid` (`participantid`,`unitid`),
  KEY `assaultid` (`assaultid`),
  KEY `userid` (`userid`),
  KEY `unitid` (`unitid`),
  CONSTRAINT `na_fleet2assault_ibfk_1` FOREIGN KEY (`assaultid`) REFERENCES `na_assault` (`assaultid`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `na_fleet2assault_ibfk_3` FOREIGN KEY (`userid`) REFERENCES `na_user` (`userid`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `na_fleet2assault_ibfk_4` FOREIGN KEY (`participantid`) REFERENCES `na_assaultparticipant` (`participantid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_folder`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_folder` (
  `folder_id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `label` varbinary(128) NOT NULL,
  `userid` int(9) unsigned DEFAULT NULL,
  `is_standard` tinyint(1) unsigned NOT NULL,
  `display_order` int(10) unsigned NOT NULL,
  PRIMARY KEY (`folder_id`),
  KEY `userid` (`userid`),
  KEY `display_order` (`display_order`,`folder_id`),
  CONSTRAINT `na_folder_ibfk_1` FOREIGN KEY (`userid`) REFERENCES `na_user` (`userid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=12 DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_formation_invitation`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_formation_invitation` (
  `eventid` int(10) NOT NULL,
  `userid` int(9) unsigned NOT NULL,
  KEY `eventid` (`eventid`),
  KEY `userid` (`userid`),
  CONSTRAINT `na_formation_invitation_ibfk_1` FOREIGN KEY (`userid`) REFERENCES `na_user` (`userid`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `na_formation_invitation_ibfk_2` FOREIGN KEY (`eventid`) REFERENCES `na_events` (`eventid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_free_planets`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_free_planets` (
  `galaxy` int(3) unsigned DEFAULT NULL,
  `system` int(4) unsigned DEFAULT NULL,
  `cnt` bigint(21) DEFAULT NULL,
  `free_pos` bigint(11) DEFAULT NULL
) ENGINE=MyISAM DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_galaxy`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_galaxy` (
  `galaxy` int(3) unsigned NOT NULL,
  `system` int(4) unsigned NOT NULL,
  `position` int(2) unsigned NOT NULL,
  `metal` double(15,0) NOT NULL,
  `silicon` double(15,0) NOT NULL,
  `planetid` int(10) unsigned NOT NULL,
  `moonid` int(10) unsigned DEFAULT NULL,
  `destroyed` tinyint(1) unsigned NOT NULL,
  UNIQUE KEY `galaxy` (`galaxy`,`system`,`position`),
  UNIQUE KEY `planetid` (`planetid`),
  UNIQUE KEY `moonid` (`moonid`),
  KEY `system` (`system`),
  KEY `galaxy_2` (`galaxy`,`destroyed`),
  KEY `galaxy_3` (`galaxy`,`position`,`destroyed`),
  CONSTRAINT `na_galaxy_ibfk_1` FOREIGN KEY (`planetid`) REFERENCES `na_planet` (`planetid`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `na_galaxy_ibfk_2` FOREIGN KEY (`moonid`) REFERENCES `na_planet` (`planetid`) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_galaxy_active`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_galaxy_active` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=9 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_galaxy_empty_new_pos`;
/*!50001 DROP VIEW IF EXISTS `na_galaxy_empty_new_pos`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_galaxy_empty_new_pos` AS SELECT 
 1 AS `galaxy`,
 1 AS `system`,
 1 AS `position`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_galaxy_empty_new_pos2`;
/*!50001 DROP VIEW IF EXISTS `na_galaxy_empty_new_pos2`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_galaxy_empty_new_pos2` AS SELECT 
 1 AS `galaxy`,
 1 AS `system`,
 1 AS `position`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_galaxy_empty_new_pos_all`;
/*!50001 DROP VIEW IF EXISTS `na_galaxy_empty_new_pos_all`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_galaxy_empty_new_pos_all` AS SELECT 
 1 AS `galaxy`,
 1 AS `system`,
 1 AS `position`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_galaxy_empty_new_pos_sum`;
/*!50001 DROP VIEW IF EXISTS `na_galaxy_empty_new_pos_sum`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_galaxy_empty_new_pos_sum` AS SELECT 
 1 AS `sum_empty`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_galaxy_err_rownum`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_galaxy_err_rownum` (
  `rownum` int(11) NOT NULL AUTO_INCREMENT,
  `galaxy` int(11) NOT NULL,
  `system` int(11) NOT NULL,
  `position` int(11) NOT NULL,
  PRIMARY KEY (`rownum`),
  UNIQUE KEY `galaxy` (`galaxy`,`system`,`position`)
) ENGINE=MyISAM AUTO_INCREMENT=3 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_galaxy_fix`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_galaxy_fix` (
  `galaxy` int(11) NOT NULL,
  `system` int(11) NOT NULL,
  `position` int(11) NOT NULL,
  `new_system` int(11) NOT NULL,
  UNIQUE KEY `galazy` (`galaxy`,`new_system`,`position`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_galaxy_free`;
/*!50001 DROP VIEW IF EXISTS `na_galaxy_free`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_galaxy_free` AS SELECT 
 1 AS `galaxy`,
 1 AS `system`,
 1 AS `cnt`,
 1 AS `free_pos`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_galaxy_free_pos`;
/*!50001 DROP VIEW IF EXISTS `na_galaxy_free_pos`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_galaxy_free_pos` AS SELECT 
 1 AS `galaxy`,
 1 AS `system`,
 1 AS `cnt`,
 1 AS `position`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_galaxy_free_pos2`;
/*!50001 DROP VIEW IF EXISTS `na_galaxy_free_pos2`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_galaxy_free_pos2` AS SELECT 
 1 AS `galaxy`,
 1 AS `system`,
 1 AS `position`,
 1 AS `cnt`,
 1 AS `free_cnt`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_galaxy_free_pos_rnd2`;
/*!50001 DROP VIEW IF EXISTS `na_galaxy_free_pos_rnd2`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_galaxy_free_pos_rnd2` AS SELECT 
 1 AS `galaxy`,
 1 AS `system`,
 1 AS `position`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_galaxy_free_pos_rnd_cut2`;
/*!50001 DROP VIEW IF EXISTS `na_galaxy_free_pos_rnd_cut2`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_galaxy_free_pos_rnd_cut2` AS SELECT 
 1 AS `galaxy`,
 1 AS `system`,
 1 AS `position`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_galaxy_new_active`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_galaxy_new_active` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=9 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_galaxy_new_pos`;
/*!50001 DROP VIEW IF EXISTS `na_galaxy_new_pos`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_galaxy_new_pos` AS SELECT 
 1 AS `galaxy`,
 1 AS `system`,
 1 AS `position`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_galaxy_new_pos_cut2`;
/*!50001 DROP VIEW IF EXISTS `na_galaxy_new_pos_cut2`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_galaxy_new_pos_cut2` AS SELECT 
 1 AS `galaxy`,
 1 AS `system`,
 1 AS `position`,
 1 AS `type`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_galaxy_new_pos_sum2`;
/*!50001 DROP VIEW IF EXISTS `na_galaxy_new_pos_sum2`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_galaxy_new_pos_sum2` AS SELECT 
 1 AS `destroyed_planet_cnt`,
 1 AS `free_planet_cnt`,
 1 AS `empty_galaxy_cnt`,
 1 AS `empty_planet_cnt`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_galaxy_new_pos_union2`;
/*!50001 DROP VIEW IF EXISTS `na_galaxy_new_pos_union2`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_galaxy_new_pos_union2` AS SELECT 
 1 AS `galaxy`,
 1 AS `system`,
 1 AS `position`,
 1 AS `type`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_galaxy_tmp`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_galaxy_tmp` (
  `galaxy` int(3) unsigned NOT NULL,
  `system` int(4) unsigned NOT NULL,
  `position` int(2) unsigned NOT NULL,
  `metal` double(15,0) NOT NULL,
  `silicon` double(15,0) NOT NULL,
  `planetid` int(10) unsigned NOT NULL,
  `moonid` int(10) unsigned DEFAULT NULL,
  `destroyed` tinyint(1) unsigned NOT NULL,
  UNIQUE KEY `galaxy` (`galaxy`,`system`,`position`),
  KEY `planetid` (`planetid`),
  KEY `moonid` (`moonid`),
  KEY `system` (`system`)
) ENGINE=MyISAM DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_galaxy_with_destroyed`;
/*!50001 DROP VIEW IF EXISTS `na_galaxy_with_destroyed`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_galaxy_with_destroyed` AS SELECT 
 1 AS `galaxy`,
 1 AS `cnt`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_galaxy_with_destroyed2`;
/*!50001 DROP VIEW IF EXISTS `na_galaxy_with_destroyed2`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_galaxy_with_destroyed2` AS SELECT 
 1 AS `galaxy`,
 1 AS `generic_cnt`,
 1 AS `ufo_cnt`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_galaxy_with_free_pos`;
/*!50001 DROP VIEW IF EXISTS `na_galaxy_with_free_pos`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_galaxy_with_free_pos` AS SELECT 
 1 AS `galaxy`,
 1 AS `system`,
 1 AS `cnt`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_galaxy_with_free_pos2`;
/*!50001 DROP VIEW IF EXISTS `na_galaxy_with_free_pos2`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_galaxy_with_free_pos2` AS SELECT 
 1 AS `galaxy`,
 1 AS `system`,
 1 AS `cnt`,
 1 AS `free_cnt`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_galaxy_with_free_pos_all`;
/*!50001 DROP VIEW IF EXISTS `na_galaxy_with_free_pos_all`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_galaxy_with_free_pos_all` AS SELECT 
 1 AS `cnt`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_galaxy_with_free_pos_sum`;
/*!50001 DROP VIEW IF EXISTS `na_galaxy_with_free_pos_sum`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_galaxy_with_free_pos_sum` AS SELECT 
 1 AS `sum_free`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_group2permission`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_group2permission` (
  `permissionid` int(4) unsigned NOT NULL,
  `groupid` int(3) unsigned NOT NULL,
  `value` tinyint(1) unsigned NOT NULL,
  KEY `permissionid` (`permissionid`),
  KEY `groupid` (`groupid`),
  CONSTRAINT `na_group2permission_ibfk_1` FOREIGN KEY (`permissionid`) REFERENCES `na_permissions` (`permissionid`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `na_group2permission_ibfk_2` FOREIGN KEY (`groupid`) REFERENCES `na_usergroup` (`usergroupid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_languages`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_languages` (
  `languageid` int(4) unsigned NOT NULL AUTO_INCREMENT,
  `title` varbinary(128) NOT NULL,
  `langcode` varbinary(12) NOT NULL,
  `charset` varbinary(15) NOT NULL,
  `display_order` int(11) NOT NULL DEFAULT '0',
  PRIMARY KEY (`languageid`),
  KEY `langcode` (`langcode`),
  KEY `display_order` (`display_order`)
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_log`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_log` (
  `logid` int(11) NOT NULL AUTO_INCREMENT,
  `time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `dt` float NOT NULL DEFAULT '0',
  `message` mediumtext NOT NULL,
  PRIMARY KEY (`logid`),
  KEY `time` (`time`)
) ENGINE=InnoDB AUTO_INCREMENT=167 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_log_error_fb_external`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_log_error_fb_external` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `level` varchar(128) DEFAULT NULL,
  `category` varchar(128) DEFAULT NULL,
  `logtime` int(11) DEFAULT NULL,
  `message` text,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM AUTO_INCREMENT=3328 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_log_error_fb_external_chat`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_log_error_fb_external_chat` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `level` varchar(128) DEFAULT NULL,
  `category` varchar(128) DEFAULT NULL,
  `logtime` int(11) DEFAULT NULL,
  `message` text,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM AUTO_INCREMENT=3 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_log_error_google_external`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_log_error_google_external` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `level` varchar(128) DEFAULT NULL,
  `category` varchar(128) DEFAULT NULL,
  `logtime` int(11) DEFAULT NULL,
  `message` text,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM AUTO_INCREMENT=3 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_log_error_index`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_log_error_index` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `level` varchar(128) DEFAULT NULL,
  `category` varchar(128) DEFAULT NULL,
  `logtime` int(11) DEFAULT NULL,
  `message` text,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_log_error_index_ext`;
/*!50001 DROP VIEW IF EXISTS `na_log_error_index_ext`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_log_error_index_ext` AS SELECT 
 1 AS `t`,
 1 AS `id`,
 1 AS `level`,
 1 AS `category`,
 1 AS `logtime`,
 1 AS `message`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_log_error_mailru`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_log_error_mailru` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `level` varchar(128) DEFAULT NULL,
  `category` varchar(128) DEFAULT NULL,
  `logtime` int(11) DEFAULT NULL,
  `message` text,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM AUTO_INCREMENT=3801 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_log_error_mailru_chat`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_log_error_mailru_chat` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `level` varchar(128) DEFAULT NULL,
  `category` varchar(128) DEFAULT NULL,
  `logtime` int(11) DEFAULT NULL,
  `message` text,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM AUTO_INCREMENT=31296 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_log_error_mailru_external`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_log_error_mailru_external` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `level` varchar(128) DEFAULT NULL,
  `category` varchar(128) DEFAULT NULL,
  `logtime` int(11) DEFAULT NULL,
  `message` text,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM AUTO_INCREMENT=2295 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_log_error_mailru_external_chat`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_log_error_mailru_external_chat` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `level` varchar(128) DEFAULT NULL,
  `category` varchar(128) DEFAULT NULL,
  `logtime` int(11) DEFAULT NULL,
  `message` text,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM AUTO_INCREMENT=42 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_log_error_main`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_log_error_main` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `level` varchar(128) DEFAULT NULL,
  `category` varchar(128) DEFAULT NULL,
  `logtime` int(11) DEFAULT NULL,
  `message` text,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM AUTO_INCREMENT=236936 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_log_error_main_chat`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_log_error_main_chat` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `level` varchar(128) DEFAULT NULL,
  `category` varchar(128) DEFAULT NULL,
  `logtime` int(11) DEFAULT NULL,
  `message` text,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM AUTO_INCREMENT=1337 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_log_error_main_ext`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_log_error_main_ext` (
  `t` datetime DEFAULT NULL,
  `id` int(11) DEFAULT NULL,
  `level` varchar(128) DEFAULT NULL,
  `category` varchar(128) DEFAULT NULL,
  `logtime` int(11) DEFAULT NULL,
  `message` text
) ENGINE=InnoDB DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_log_error_odnk_external`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_log_error_odnk_external` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `level` varchar(128) DEFAULT NULL,
  `category` varchar(128) DEFAULT NULL,
  `logtime` int(11) DEFAULT NULL,
  `message` text,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM AUTO_INCREMENT=6948 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_log_error_odnk_external_chat`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_log_error_odnk_external_chat` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `level` varchar(128) DEFAULT NULL,
  `category` varchar(128) DEFAULT NULL,
  `logtime` int(11) DEFAULT NULL,
  `message` text,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM AUTO_INCREMENT=2226 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_log_error_odnoklassniki`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_log_error_odnoklassniki` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `level` varchar(128) DEFAULT NULL,
  `category` varchar(128) DEFAULT NULL,
  `logtime` int(11) DEFAULT NULL,
  `message` text,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM AUTO_INCREMENT=20372 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_log_error_odnoklassniki_chat`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_log_error_odnoklassniki_chat` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `level` varchar(128) DEFAULT NULL,
  `category` varchar(128) DEFAULT NULL,
  `logtime` int(11) DEFAULT NULL,
  `message` text,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM AUTO_INCREMENT=22310 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_log_error_odnoklassniki_ext`;
/*!50001 DROP VIEW IF EXISTS `na_log_error_odnoklassniki_ext`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_log_error_odnoklassniki_ext` AS SELECT 
 1 AS `t`,
 1 AS `id`,
 1 AS `level`,
 1 AS `category`,
 1 AS `logtime`,
 1 AS `message`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_log_error_vkontakte`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_log_error_vkontakte` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `level` varchar(128) DEFAULT NULL,
  `category` varchar(128) DEFAULT NULL,
  `logtime` int(11) DEFAULT NULL,
  `message` text,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM AUTO_INCREMENT=16843 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_log_error_vkontakte_chat`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_log_error_vkontakte_chat` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `level` varchar(128) DEFAULT NULL,
  `category` varchar(128) DEFAULT NULL,
  `logtime` int(11) DEFAULT NULL,
  `message` text,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM AUTO_INCREMENT=2427 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_log_error_vkontakte_external`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_log_error_vkontakte_external` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `level` varchar(128) DEFAULT NULL,
  `category` varchar(128) DEFAULT NULL,
  `logtime` int(11) DEFAULT NULL,
  `message` text,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM AUTO_INCREMENT=2106 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_log_error_vkontakte_external_chat`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_log_error_vkontakte_external_chat` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `level` varchar(128) DEFAULT NULL,
  `category` varchar(128) DEFAULT NULL,
  `logtime` int(11) DEFAULT NULL,
  `message` text,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM AUTO_INCREMENT=3 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_log_warning_fb_external`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_log_warning_fb_external` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `level` varchar(128) DEFAULT NULL,
  `category` varchar(128) DEFAULT NULL,
  `logtime` int(11) DEFAULT NULL,
  `message` text,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM AUTO_INCREMENT=12 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_log_warning_fb_external_chat`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_log_warning_fb_external_chat` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `level` varchar(128) DEFAULT NULL,
  `category` varchar(128) DEFAULT NULL,
  `logtime` int(11) DEFAULT NULL,
  `message` text,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM AUTO_INCREMENT=3140 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_log_warning_google_external`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_log_warning_google_external` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `level` varchar(128) DEFAULT NULL,
  `category` varchar(128) DEFAULT NULL,
  `logtime` int(11) DEFAULT NULL,
  `message` text,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_log_warning_index`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_log_warning_index` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `level` varchar(128) DEFAULT NULL,
  `category` varchar(128) DEFAULT NULL,
  `logtime` int(11) DEFAULT NULL,
  `message` text,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM AUTO_INCREMENT=66 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_log_warning_index_ext`;
/*!50001 DROP VIEW IF EXISTS `na_log_warning_index_ext`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_log_warning_index_ext` AS SELECT 
 1 AS `t`,
 1 AS `id`,
 1 AS `level`,
 1 AS `category`,
 1 AS `logtime`,
 1 AS `message`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_log_warning_mailru`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_log_warning_mailru` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `level` varchar(128) DEFAULT NULL,
  `category` varchar(128) DEFAULT NULL,
  `logtime` int(11) DEFAULT NULL,
  `message` text,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM AUTO_INCREMENT=22 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_log_warning_mailru_chat`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_log_warning_mailru_chat` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `level` varchar(128) DEFAULT NULL,
  `category` varchar(128) DEFAULT NULL,
  `logtime` int(11) DEFAULT NULL,
  `message` text,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM AUTO_INCREMENT=4032 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_log_warning_mailru_external`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_log_warning_mailru_external` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `level` varchar(128) DEFAULT NULL,
  `category` varchar(128) DEFAULT NULL,
  `logtime` int(11) DEFAULT NULL,
  `message` text,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM AUTO_INCREMENT=23 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_log_warning_mailru_external_chat`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_log_warning_mailru_external_chat` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `level` varchar(128) DEFAULT NULL,
  `category` varchar(128) DEFAULT NULL,
  `logtime` int(11) DEFAULT NULL,
  `message` text,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM AUTO_INCREMENT=14143 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_log_warning_main`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_log_warning_main` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `level` varchar(128) DEFAULT NULL,
  `category` varchar(128) DEFAULT NULL,
  `logtime` int(11) DEFAULT NULL,
  `message` text,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM AUTO_INCREMENT=2336 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_log_warning_main_chat`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_log_warning_main_chat` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `level` varchar(128) DEFAULT NULL,
  `category` varchar(128) DEFAULT NULL,
  `logtime` int(11) DEFAULT NULL,
  `message` text,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM AUTO_INCREMENT=605679 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_log_warning_main_ext`;
/*!50001 DROP VIEW IF EXISTS `na_log_warning_main_ext`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_log_warning_main_ext` AS SELECT 
 1 AS `t`,
 1 AS `id`,
 1 AS `level`,
 1 AS `category`,
 1 AS `logtime`,
 1 AS `message`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_log_warning_odnk_external`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_log_warning_odnk_external` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `level` varchar(128) DEFAULT NULL,
  `category` varchar(128) DEFAULT NULL,
  `logtime` int(11) DEFAULT NULL,
  `message` text,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM AUTO_INCREMENT=408 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_log_warning_odnk_external_chat`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_log_warning_odnk_external_chat` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `level` varchar(128) DEFAULT NULL,
  `category` varchar(128) DEFAULT NULL,
  `logtime` int(11) DEFAULT NULL,
  `message` text,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM AUTO_INCREMENT=294578 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_log_warning_odnoklassniki`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_log_warning_odnoklassniki` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `level` varchar(128) DEFAULT NULL,
  `category` varchar(128) DEFAULT NULL,
  `logtime` int(11) DEFAULT NULL,
  `message` text,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM AUTO_INCREMENT=127443 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_log_warning_odnoklassniki_chat`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_log_warning_odnoklassniki_chat` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `level` varchar(128) DEFAULT NULL,
  `category` varchar(128) DEFAULT NULL,
  `logtime` int(11) DEFAULT NULL,
  `message` text,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM AUTO_INCREMENT=75996 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_log_warning_odnoklassniki_ext`;
/*!50001 DROP VIEW IF EXISTS `na_log_warning_odnoklassniki_ext`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_log_warning_odnoklassniki_ext` AS SELECT 
 1 AS `t`,
 1 AS `id`,
 1 AS `level`,
 1 AS `category`,
 1 AS `logtime`,
 1 AS `message`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_log_warning_vkontakte`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_log_warning_vkontakte` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `level` varchar(128) DEFAULT NULL,
  `category` varchar(128) DEFAULT NULL,
  `logtime` int(11) DEFAULT NULL,
  `message` text,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM AUTO_INCREMENT=44 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_log_warning_vkontakte_chat`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_log_warning_vkontakte_chat` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `level` varchar(128) DEFAULT NULL,
  `category` varchar(128) DEFAULT NULL,
  `logtime` int(11) DEFAULT NULL,
  `message` text,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM AUTO_INCREMENT=6469 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_log_warning_vkontakte_external`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_log_warning_vkontakte_external` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `level` varchar(128) DEFAULT NULL,
  `category` varchar(128) DEFAULT NULL,
  `logtime` int(11) DEFAULT NULL,
  `message` text,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM AUTO_INCREMENT=12 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_log_warning_vkontakte_external_chat`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_log_warning_vkontakte_external_chat` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `level` varchar(128) DEFAULT NULL,
  `category` varchar(128) DEFAULT NULL,
  `logtime` int(11) DEFAULT NULL,
  `message` text,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM AUTO_INCREMENT=3610 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_loginattempts`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_loginattempts` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `time` int(10) unsigned NOT NULL,
  `ip` varbinary(16) NOT NULL,
  `username` varbinary(255) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2617828 DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_message`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_message` (
  `msgid` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `mode` int(1) unsigned NOT NULL,
  `time` int(10) unsigned NOT NULL,
  `sender` int(10) unsigned DEFAULT NULL,
  `receiver` int(10) unsigned NOT NULL,
  `related_user` int(10) unsigned DEFAULT NULL,
  `message` text NOT NULL,
  `subject` varchar(255) NOT NULL,
  `readed` tinyint(1) unsigned NOT NULL,
  PRIMARY KEY (`msgid`),
  KEY `sender` (`sender`),
  KEY `related_user` (`related_user`),
  KEY `receiver` (`receiver`,`readed`),
  KEY `receiver_2` (`receiver`,`mode`,`time`,`msgid`),
  CONSTRAINT `na_message_ibfk_1` FOREIGN KEY (`sender`) REFERENCES `na_user` (`userid`) ON DELETE SET NULL ON UPDATE CASCADE,
  CONSTRAINT `na_message_ibfk_2` FOREIGN KEY (`receiver`) REFERENCES `na_user` (`userid`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `na_message_ibfk_3` FOREIGN KEY (`related_user`) REFERENCES `na_user` (`userid`) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=68119741 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_message_ext`;
/*!50001 DROP VIEW IF EXISTS `na_message_ext`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_message_ext` AS SELECT 
 1 AS `t`,
 1 AS `s`,
 1 AS `r`,
 1 AS `msgid`,
 1 AS `mode`,
 1 AS `time`,
 1 AS `sender`,
 1 AS `receiver`,
 1 AS `related_user`,
 1 AS `message`,
 1 AS `subject`,
 1 AS `readed`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_moon_creation_stats`;
/*!50001 DROP VIEW IF EXISTS `na_moon_creation_stats`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_moon_creation_stats` AS SELECT 
 1 AS `moonchance`,
 1 AS `cnt`,
 1 AS `calc_moons`,
 1 AS `real_moons`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_moon_destroy_stats`;
/*!50001 DROP VIEW IF EXISTS `na_moon_destroy_stats`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_moon_destroy_stats` AS SELECT 
 1 AS `cnt`,
 1 AS `target_destroyed`,
 1 AS `ptd`,
 1 AS `attacker_explode`,
 1 AS `pae`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_moon_destroy_stats_old`;
/*!50001 DROP VIEW IF EXISTS `na_moon_destroy_stats_old`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_moon_destroy_stats_old` AS SELECT 
 1 AS `cnt`,
 1 AS `target_destroyed`,
 1 AS `ptd`,
 1 AS `attacker_explode`,
 1 AS `pae`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_moon_destroy_stats_old2`;
/*!50001 DROP VIEW IF EXISTS `na_moon_destroy_stats_old2`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_moon_destroy_stats_old2` AS SELECT 
 1 AS `cnt`,
 1 AS `target_destroyed`,
 1 AS `ptd`,
 1 AS `attacker_explode`,
 1 AS `pae`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_notes`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_notes` (
  `user_id` int(10) unsigned NOT NULL,
  `notes` text,
  PRIMARY KEY (`user_id`),
  CONSTRAINT `na_notes_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `na_user` (`userid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_officer`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_officer` (
  `userid` int(10) unsigned NOT NULL,
  `of_points` int(8) NOT NULL DEFAULT '0',
  `of_batle` int(8) NOT NULL DEFAULT '0',
  `of_level` int(8) NOT NULL DEFAULT '0',
  `credit` double(15,2) NOT NULL DEFAULT '3.00',
  `of_1` int(10) unsigned NOT NULL DEFAULT '0',
  `of_2` int(10) unsigned NOT NULL DEFAULT '0',
  `of_3` int(10) unsigned NOT NULL DEFAULT '0',
  `of_4` int(10) unsigned NOT NULL DEFAULT '0',
  `tmp` tinyint(4) NOT NULL DEFAULT '0',
  KEY `userid` (`userid`),
  CONSTRAINT `na_officer_ibfk_1` FOREIGN KEY (`userid`) REFERENCES `na_user` (`userid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_officer_ext`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_officer_ext` (
  `l` datetime DEFAULT NULL,
  `username` varchar(32) DEFAULT NULL,
  `userid` int(10) unsigned DEFAULT NULL,
  `of_points` int(8) DEFAULT NULL,
  `of_batle` int(8) DEFAULT NULL,
  `of_level` int(8) DEFAULT NULL,
  `credit` double(15,2) DEFAULT NULL,
  `of_1` int(10) unsigned DEFAULT NULL,
  `of_2` int(10) unsigned DEFAULT NULL,
  `of_3` int(10) unsigned DEFAULT NULL,
  `of_4` int(10) unsigned DEFAULT NULL,
  `tmp` tinyint(4) DEFAULT NULL
) ENGINE=MyISAM DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_page`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_page` (
  `pageid` int(4) unsigned NOT NULL AUTO_INCREMENT,
  `position` char(1) CHARACTER SET utf8 NOT NULL,
  `languageid` int(4) unsigned NOT NULL,
  `displayorder` int(4) unsigned NOT NULL,
  `title` varbinary(32) NOT NULL,
  `label` varbinary(32) NOT NULL,
  `link` varbinary(128) NOT NULL,
  `content` text CHARACTER SET utf8 COLLATE utf8_bin NOT NULL,
  PRIMARY KEY (`pageid`),
  KEY `languageid` (`languageid`),
  CONSTRAINT `na_page_ibfk_1` FOREIGN KEY (`languageid`) REFERENCES `na_languages` (`languageid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_password`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_password` (
  `userid` int(9) unsigned NOT NULL,
  `password` varbinary(32) NOT NULL,
  `password_sha1` varbinary(40) NOT NULL,
  `time` int(10) unsigned NOT NULL,
  UNIQUE KEY `userid` (`userid`),
  CONSTRAINT `na_password_ibfk_1` FOREIGN KEY (`userid`) REFERENCES `na_user` (`userid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_payment_stats`;
/*!50001 DROP VIEW IF EXISTS `na_payment_stats`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_payment_stats` AS SELECT 
 1 AS `d`,
 1 AS `credit`,
 1 AS `amount_r`,
 1 AS `cnt`,
 1 AS `odnk_r`,
 1 AS `vk_r`,
 1 AS `mailru_r`,
 1 AS `oxsar_r`,
 1 AS `real_r`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_payment_stats_month`;
/*!50001 DROP VIEW IF EXISTS `na_payment_stats_month`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_payment_stats_month` AS SELECT 
 1 AS `d`,
 1 AS `credit`,
 1 AS `amount_r`,
 1 AS `cnt`,
 1 AS `odnk_r`,
 1 AS `vk_r`,
 1 AS `mailru_r`,
 1 AS `oxsar_r`,
 1 AS `real_r`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_payment_user_stats`;
/*!50001 DROP VIEW IF EXISTS `na_payment_user_stats`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_payment_user_stats` AS SELECT 
 1 AS `d`,
 1 AS `userid`,
 1 AS `username`,
 1 AS `credit`,
 1 AS `amount_r`,
 1 AS `cnt`,
 1 AS `odnk_r`,
 1 AS `vk_r`,
 1 AS `mailru_r`,
 1 AS `oxsar_r`,
 1 AS `real_r`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_payment_user_stats_month`;
/*!50001 DROP VIEW IF EXISTS `na_payment_user_stats_month`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_payment_user_stats_month` AS SELECT 
 1 AS `d`,
 1 AS `userid`,
 1 AS `username`,
 1 AS `credit`,
 1 AS `amount_r`,
 1 AS `cnt`,
 1 AS `odnk_r`,
 1 AS `vk_r`,
 1 AS `mailru_r`,
 1 AS `oxsar_r`,
 1 AS `real_r`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_payments`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_payments` (
  `pay_id` int(11) NOT NULL AUTO_INCREMENT,
  `pay_user_id` int(11) NOT NULL,
  `pay_type` varchar(100) NOT NULL,
  `pay_from` varchar(100) NOT NULL,
  `pay_amount` float(11,2) DEFAULT NULL,
  `pay_amount_r` float(11,2) DEFAULT NULL,
  `pay_credit` float(11,2) NOT NULL,
  `pay_date` datetime NOT NULL,
  `pay_status` int(1) NOT NULL,
  `pay_domain` varchar(50) DEFAULT NULL,
  `pay_extra_info` text,
  `pay_ext_transaction` int(10) unsigned DEFAULT NULL,
  PRIMARY KEY (`pay_id`),
  KEY `pay_user_id` (`pay_user_id`,`pay_date`),
  KEY `pay_user_id_2` (`pay_user_id`,`pay_from`)
) ENGINE=InnoDB AUTO_INCREMENT=52909 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_payments_ext`;
/*!50001 DROP VIEW IF EXISTS `na_payments_ext`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_payments_ext` AS SELECT 
 1 AS `username`,
 1 AS `credit`,
 1 AS `pay_id`,
 1 AS `pay_user_id`,
 1 AS `pay_type`,
 1 AS `pay_from`,
 1 AS `pay_amount`,
 1 AS `pay_amount_r`,
 1 AS `pay_credit`,
 1 AS `pay_date`,
 1 AS `pay_status`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_permissions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_permissions` (
  `permissionid` int(4) unsigned NOT NULL AUTO_INCREMENT,
  `permission` varbinary(255) NOT NULL,
  PRIMARY KEY (`permissionid`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_phrases`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_phrases` (
  `phraseid` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `languageid` int(4) unsigned NOT NULL,
  `phrasegroupid` int(4) unsigned NOT NULL,
  `title` varchar(128) NOT NULL,
  `content` varchar(10000) NOT NULL,
  `translated` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`phraseid`),
  UNIQUE KEY `title_2` (`title`,`phrasegroupid`,`languageid`),
  KEY `languageid` (`languageid`,`phrasegroupid`),
  KEY `phrasegroupid` (`phrasegroupid`),
  KEY `title` (`title`)
) ENGINE=InnoDB AUTO_INCREMENT=9507 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_phrases_tmp`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_phrases_tmp` (
  `phraseid` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `languageid` int(4) unsigned NOT NULL,
  `phrasegroupid` int(4) unsigned NOT NULL,
  `title` varchar(128) NOT NULL,
  `content` varchar(10000) NOT NULL,
  PRIMARY KEY (`phraseid`),
  KEY `languageid` (`languageid`,`phrasegroupid`),
  KEY `phrasegroupid` (`phrasegroupid`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_phrasesgroups`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_phrasesgroups` (
  `phrasegroupid` int(4) unsigned NOT NULL AUTO_INCREMENT,
  `title` varbinary(255) NOT NULL,
  PRIMARY KEY (`phrasegroupid`),
  KEY `title` (`title`,`phrasegroupid`)
) ENGINE=InnoDB AUTO_INCREMENT=26 DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_planet`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_planet` (
  `planetid` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `userid` int(9) unsigned DEFAULT NULL,
  `ismoon` tinyint(1) unsigned NOT NULL,
  `umi` float NOT NULL,
  `planetname` varchar(255) CHARACTER SET utf8 NOT NULL,
  `diameter` int(10) unsigned NOT NULL,
  `picture` varbinary(255) NOT NULL,
  `temperature` smallint(6) NOT NULL,
  `last` int(10) unsigned NOT NULL,
  `metal` double NOT NULL DEFAULT '500',
  `silicon` double NOT NULL DEFAULT '500',
  `hydrogen` double NOT NULL DEFAULT '0',
  `solar_satellite_prod` int(3) unsigned NOT NULL DEFAULT '100',
  `build_factor` float NOT NULL DEFAULT '1',
  `research_factor` float NOT NULL DEFAULT '1',
  `produce_factor` float NOT NULL DEFAULT '1',
  `energy_factor` float NOT NULL DEFAULT '1',
  `storage_factor` float NOT NULL DEFAULT '1',
  `destroy_eventid` int(11) DEFAULT NULL,
  PRIMARY KEY (`planetid`),
  KEY `userid` (`userid`),
  KEY `umi` (`umi`),
  CONSTRAINT `na_planet_ibfk_1` FOREIGN KEY (`userid`) REFERENCES `na_user` (`userid`) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=1764988 DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_planet_err_rownum`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_planet_err_rownum` (
  `rownum` int(11) NOT NULL AUTO_INCREMENT,
  `planetid` int(11) NOT NULL,
  PRIMARY KEY (`rownum`),
  KEY `planetid` (`planetid`)
) ENGINE=MyISAM AUTO_INCREMENT=3 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_planet_free`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_planet_free` (
  `planet_free_pos` int(11) NOT NULL,
  PRIMARY KEY (`planet_free_pos`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_planet_new_active`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_planet_new_active` (
  `position` int(11) NOT NULL,
  PRIMARY KEY (`position`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_planet_tmp`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_planet_tmp` (
  `planetid` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `userid` int(9) unsigned DEFAULT NULL,
  `ismoon` tinyint(1) unsigned NOT NULL,
  `planetname` varchar(255) CHARACTER SET utf8 NOT NULL,
  `diameter` int(10) unsigned NOT NULL,
  `picture` varbinary(255) NOT NULL,
  `temperature` smallint(6) NOT NULL,
  `last` int(10) unsigned NOT NULL,
  `metal` double NOT NULL DEFAULT '500',
  `silicon` double NOT NULL DEFAULT '500',
  `hydrogen` double NOT NULL DEFAULT '0',
  `solar_satellite_prod` int(3) unsigned NOT NULL DEFAULT '100',
  `build_factor` float NOT NULL DEFAULT '1',
  `research_factor` float NOT NULL DEFAULT '1',
  `produce_factor` float NOT NULL DEFAULT '1',
  `energy_factor` float NOT NULL DEFAULT '1',
  `storage_factor` float NOT NULL DEFAULT '1',
  PRIMARY KEY (`planetid`),
  KEY `userid` (`userid`)
) ENGINE=MyISAM DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_rapidfire`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_rapidfire` (
  `unitid` int(4) unsigned NOT NULL,
  `target` int(4) unsigned NOT NULL,
  `value` int(4) unsigned NOT NULL,
  PRIMARY KEY (`unitid`,`target`),
  KEY `target` (`target`)
) ENGINE=InnoDB DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_referral`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_referral` (
  `userid` int(10) unsigned NOT NULL,
  `ref_id` int(10) unsigned NOT NULL,
  `ref_time` int(10) NOT NULL,
  `ref_ip` varbinary(40) NOT NULL,
  `bonus` tinyint(4) NOT NULL DEFAULT '0',
  `bonus_time` int(10) unsigned NOT NULL DEFAULT '0',
  `bonus_metal` int(10) unsigned NOT NULL DEFAULT '0',
  `bonus_silicon` int(10) unsigned NOT NULL DEFAULT '0',
  `bonus_hydrogen` int(10) unsigned NOT NULL DEFAULT '0',
  `bonus_credit` float unsigned NOT NULL DEFAULT '0',
  UNIQUE KEY `ref_id` (`ref_id`),
  KEY `user_id` (`userid`),
  CONSTRAINT `na_referral_ibfk_1` FOREIGN KEY (`userid`) REFERENCES `na_user` (`userid`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `na_referral_ibfk_2` FOREIGN KEY (`ref_id`) REFERENCES `na_user` (`userid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_referral_ext`;
/*!50001 DROP VIEW IF EXISTS `na_referral_ext`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_referral_ext` AS SELECT 
 1 AS `u`,
 1 AS `r`,
 1 AS `rt`,
 1 AS `rsn`,
 1 AS `userid`,
 1 AS `ref_id`,
 1 AS `ref_points`,
 1 AS `ref_time`,
 1 AS `ref_ip`,
 1 AS `bonus`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_registration`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_registration` (
  `time` int(10) unsigned NOT NULL,
  `ipaddress` varbinary(40) NOT NULL,
  `useragent` varbinary(255) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_requirements`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_requirements` (
  `requirementid` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `buildingid` int(10) unsigned NOT NULL,
  `needs` int(10) unsigned NOT NULL,
  `level` int(10) unsigned NOT NULL,
  `level_limit` int(10) unsigned DEFAULT NULL,
  PRIMARY KEY (`requirementid`),
  KEY `buildingid` (`buildingid`),
  KEY `needs` (`needs`)
) ENGINE=InnoDB AUTO_INCREMENT=274 DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_res_log`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_res_log` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `type` int(11) NOT NULL,
  `userid` int(11) NOT NULL,
  `planetid` int(11) DEFAULT NULL,
  `cnt` int(11) NOT NULL DEFAULT '1',
  `metal` double(15,2) NOT NULL DEFAULT '0.00',
  `silicon` double(15,2) NOT NULL DEFAULT '0.00',
  `hydrogen` double(15,2) NOT NULL DEFAULT '0.00',
  `credit` double(15,2) NOT NULL DEFAULT '0.00',
  `result_metal` double(15,2) DEFAULT NULL,
  `result_silicon` double(15,2) DEFAULT NULL,
  `result_hydrogen` double(15,2) DEFAULT NULL,
  `result_credit` double(15,2) DEFAULT NULL,
  `game_credit` double(15,2) DEFAULT NULL,
  `ownerid` int(11) DEFAULT NULL,
  `event_mode` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `time` (`time`),
  KEY `planetid` (`planetid`),
  KEY `userid` (`userid`,`planetid`,`id`),
  KEY `userid_2` (`userid`,`type`,`time`)
) ENGINE=InnoDB AUTO_INCREMENT=206384334 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_res_log_game_credit`;
/*!50001 DROP VIEW IF EXISTS `na_res_log_game_credit`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_res_log_game_credit` AS SELECT 
 1 AS `d`,
 1 AS `min_game_credit`,
 1 AS `max_game_credit`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_res_log_gift_stats`;
/*!50001 DROP VIEW IF EXISTS `na_res_log_gift_stats`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_res_log_gift_stats` AS SELECT 
 1 AS `d`,
 1 AS `max`,
 1 AS `c`,
 1 AS `cnt`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_res_log_grab_stats`;
/*!50001 DROP VIEW IF EXISTS `na_res_log_grab_stats`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_res_log_grab_stats` AS SELECT 
 1 AS `d`,
 1 AS `min`,
 1 AS `c`,
 1 AS `cnt`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_res_log_hack`;
/*!50001 DROP VIEW IF EXISTS `na_res_log_hack`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_res_log_hack` AS SELECT 
 1 AS `id`,
 1 AS `time`,
 1 AS `type`,
 1 AS `userid`,
 1 AS `planetid`,
 1 AS `cnt`,
 1 AS `metal`,
 1 AS `silicon`,
 1 AS `hydrogen`,
 1 AS `credit`,
 1 AS `result_metal`,
 1 AS `result_silicon`,
 1 AS `result_hydrogen`,
 1 AS `result_credit`,
 1 AS `ownerid`,
 1 AS `event_mode`,
 1 AS `planetid1`,
 1 AS `userid1`,
 1 AS `planetname1`,
 1 AS `username1`,
 1 AS `userid_r1`,
 1 AS `destination`,
 1 AS `planet2`,
 1 AS `username2`,
 1 AS `user_r2`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_res_log_hack_dub`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_res_log_hack_dub` (
  `min_t` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `max_t` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
  `cnt` bigint(21) DEFAULT NULL,
  `min_evid` int(11) DEFAULT NULL,
  `max_evid` int(11) DEFAULT NULL,
  `type` int(11) DEFAULT NULL,
  `userid` int(11) DEFAULT NULL,
  `planetid` int(11) DEFAULT NULL,
  `result_metal` double(15,2) DEFAULT NULL,
  `result_silicon` double(15,2) DEFAULT NULL,
  `result_hydrogen` double(15,2) DEFAULT NULL
) ENGINE=MyISAM DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_res_log_premium_stats`;
/*!50001 DROP VIEW IF EXISTS `na_res_log_premium_stats`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_res_log_premium_stats` AS SELECT 
 1 AS `d`,
 1 AS `credit`,
 1 AS `cnt`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_res_log_stats`;
/*!50001 DROP VIEW IF EXISTS `na_res_log_stats`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_res_log_stats` AS SELECT 
 1 AS `d`,
 1 AS `plus`,
 1 AS `minus`,
 1 AS `summary`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_res_log_stats_month`;
/*!50001 DROP VIEW IF EXISTS `na_res_log_stats_month`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_res_log_stats_month` AS SELECT 
 1 AS `d`,
 1 AS `plus`,
 1 AS `minus`,
 1 AS `summary`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_res_log_type_stats`;
/*!50001 DROP VIEW IF EXISTS `na_res_log_type_stats`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_res_log_type_stats` AS SELECT 
 1 AS `d`,
 1 AS `type`,
 1 AS `c`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_res_transfer`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_res_transfer` (
  `tid` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `time` int(10) unsigned NOT NULL,
  `userid` int(9) unsigned NOT NULL,
  `senderid` int(9) unsigned NOT NULL,
  `metal` double(15,0) NOT NULL,
  `silicon` double(15,0) NOT NULL,
  `hydrogen` double(15,0) NOT NULL,
  `resum` double(15,0) NOT NULL,
  PRIMARY KEY (`tid`),
  KEY `time` (`time`,`userid`,`senderid`),
  KEY `resum` (`resum`)
) ENGINE=InnoDB AUTO_INCREMENT=209122 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_research2user`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_research2user` (
  `buildingid` int(4) unsigned NOT NULL,
  `userid` int(9) unsigned NOT NULL,
  `level` int(3) unsigned NOT NULL,
  `added` tinyint(3) NOT NULL DEFAULT '0',
  PRIMARY KEY (`buildingid`,`userid`),
  KEY `userid` (`userid`),
  CONSTRAINT `na_research2user_ibfk_1` FOREIGN KEY (`userid`) REFERENCES `na_user` (`userid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_research2user_tmp`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_research2user_tmp` (
  `buildingid` int(4) unsigned NOT NULL,
  `userid` int(9) unsigned NOT NULL,
  `level` int(3) unsigned NOT NULL,
  `added` tinyint(3) unsigned NOT NULL DEFAULT '0',
  PRIMARY KEY (`buildingid`,`userid`),
  KEY `userid` (`userid`)
) ENGINE=MyISAM DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_sendreport`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_sendreport` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `userid` int(10) unsigned NOT NULL,
  `ip` char(15) NOT NULL,
  `email` varchar(64) NOT NULL,
  `time` int(11) NOT NULL,
  `status` tinyint(4) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=177 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_sessions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_sessions` (
  `sessionid` varbinary(40) NOT NULL,
  `userid` int(9) unsigned NOT NULL,
  `ipaddress` varbinary(40) NOT NULL,
  `useragent` varbinary(255) NOT NULL,
  `time` int(10) unsigned NOT NULL,
  `logged` tinyint(1) unsigned NOT NULL DEFAULT '0',
  `is_admin` tinyint(1) unsigned NOT NULL DEFAULT '0',
  PRIMARY KEY (`sessionid`),
  KEY `userid` (`userid`),
  KEY `ipaddress` (`ipaddress`),
  KEY `time` (`time`),
  CONSTRAINT `na_sessions_ibfk_1` FOREIGN KEY (`userid`) REFERENCES `na_user` (`userid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_sessions_copy`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_sessions_copy` (
  `sessionid` varbinary(40) NOT NULL,
  `userid` int(9) unsigned NOT NULL,
  `ipaddress` varbinary(40) NOT NULL,
  `useragent` varbinary(255) NOT NULL,
  `time` int(10) unsigned NOT NULL,
  `logged` tinyint(1) unsigned NOT NULL DEFAULT '0',
  PRIMARY KEY (`sessionid`),
  KEY `userid` (`userid`)
) ENGINE=InnoDB DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_sessions_ext`;
/*!50001 DROP VIEW IF EXISTS `na_sessions_ext`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_sessions_ext` AS SELECT 
 1 AS `from_unixtime(s.time)`,
 1 AS `username`,
 1 AS `sessionid`,
 1 AS `userid`,
 1 AS `ipaddress`,
 1 AS `useragent`,
 1 AS `time`,
 1 AS `logged`,
 1 AS `is_admin`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_ship2engine`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_ship2engine` (
  `unitid` int(4) unsigned NOT NULL,
  `engineid` int(4) unsigned NOT NULL,
  `level` int(3) unsigned NOT NULL,
  `base_speed` int(6) unsigned NOT NULL,
  `base` tinyint(1) unsigned NOT NULL,
  KEY `unitid` (`unitid`),
  KEY `engineid` (`engineid`),
  CONSTRAINT `na_ship2engine_ibfk_1` FOREIGN KEY (`engineid`) REFERENCES `na_engine` (`engineid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_ship_datasheet`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_ship_datasheet` (
  `unitid` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `capicity` int(10) unsigned NOT NULL,
  `speed` int(10) unsigned NOT NULL,
  `consume` int(10) unsigned NOT NULL,
  `attack` int(10) unsigned NOT NULL,
  `shield` int(10) unsigned NOT NULL,
  `front` int(10) unsigned NOT NULL,
  `ballistics` int(10) unsigned NOT NULL,
  `masking` int(10) unsigned NOT NULL,
  `attacker_attack` int(10) unsigned NOT NULL,
  `attacker_shield` int(10) unsigned NOT NULL,
  `attacker_front` int(10) unsigned NOT NULL,
  `attacker_ballistics` int(10) unsigned NOT NULL,
  `attacker_masking` int(10) unsigned NOT NULL,
  PRIMARY KEY (`unitid`)
) ENGINE=InnoDB AUTO_INCREMENT=359 DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_ships_log`;
/*!50001 DROP VIEW IF EXISTS `na_ships_log`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_ships_log` AS SELECT 
 1 AS `created`,
 1 AS `old_quantity`,
 1 AS `quantity`,
 1 AS `is_adding`,
 1 AS `new_quantity`,
 1 AS `message`,
 1 AS `unitid`,
 1 AS `name`,
 1 AS `content`,
 1 AS `planetid`,
 1 AS `planetname`,
 1 AS `userid`,
 1 AS `username`,
 1 AS `galaxy`,
 1 AS `system`,
 1 AS `position`,
 1 AS `moonid`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_shipyard`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_shipyard` (
  `orderid` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `planetid` int(10) unsigned NOT NULL,
  `unitid` int(4) unsigned NOT NULL,
  `quantity` int(10) unsigned NOT NULL,
  `one` int(6) unsigned NOT NULL,
  `time` int(10) unsigned NOT NULL,
  `finished` int(10) unsigned NOT NULL,
  PRIMARY KEY (`orderid`),
  KEY `planetid` (`planetid`),
  KEY `unitid` (`unitid`),
  KEY `time` (`time`,`unitid`),
  CONSTRAINT `na_shipyard_ibfk_1` FOREIGN KEY (`planetid`) REFERENCES `na_planet` (`planetid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_sim_assault`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_sim_assault` (
  `assaultid` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `key` varbinary(4) DEFAULT NULL,
  `key2` varbinary(4) DEFAULT NULL,
  `result` tinyint(1) unsigned DEFAULT NULL,
  `planetid` int(10) unsigned DEFAULT NULL,
  `time` int(10) unsigned DEFAULT NULL,
  `target_moon` tinyint(1) unsigned NOT NULL DEFAULT '0',
  `target_buildingid` int(11) DEFAULT NULL,
  `building_level` smallint(6) unsigned DEFAULT NULL,
  `building_metal` double(15,0) DEFAULT NULL,
  `building_silicon` double(15,0) DEFAULT NULL,
  `building_hydrogen` double(15,0) DEFAULT NULL,
  `building_destroyed` tinyint(1) unsigned DEFAULT NULL,
  `target_destroyed` tinyint(1) unsigned NOT NULL DEFAULT '0',
  `attacker_explode` tinyint(1) unsigned NOT NULL DEFAULT '0',
  `building_destroy_chance` tinyint(3) unsigned DEFAULT NULL,
  `moon_allow_type` tinyint(3) unsigned NOT NULL DEFAULT '0',
  `moonchance` tinyint(3) unsigned DEFAULT NULL,
  `moon` tinyint(1) unsigned DEFAULT NULL,
  `attacker_lost_res` double(15,0) DEFAULT NULL,
  `attacker_lost_metal` double(15,0) DEFAULT NULL,
  `attacker_lost_silicon` double(15,0) DEFAULT NULL,
  `attacker_lost_hydrogen` double(15,0) DEFAULT NULL,
  `defender_lost_res` double(15,0) DEFAULT NULL,
  `defender_lost_metal` double(15,0) DEFAULT NULL,
  `defender_lost_silicon` double(15,0) DEFAULT NULL,
  `defender_lost_hydrogen` double(15,0) DEFAULT NULL,
  `debris_metal` double(15,0) DEFAULT NULL,
  `debris_silicon` double(15,0) DEFAULT NULL,
  `planet_metal` double(15,0) DEFAULT NULL,
  `planet_silicon` double(15,0) DEFAULT NULL,
  `planet_hydrogen` double(15,0) DEFAULT NULL,
  `haul_metal` double(15,0) DEFAULT NULL,
  `haul_silicon` double(15,0) DEFAULT NULL,
  `haul_hydrogen` double(15,0) DEFAULT NULL,
  `lostunits_attacker` int(10) unsigned NOT NULL,
  `lostunits_defender` int(10) unsigned NOT NULL,
  `attacker_exp` float DEFAULT NULL,
  `defender_exp` float DEFAULT NULL,
  `turns` tinyint(3) unsigned DEFAULT NULL,
  `turns_min` tinyint(3) unsigned DEFAULT NULL,
  `turns_max` tinyint(3) unsigned DEFAULT NULL,
  `attacker_win_percent` tinyint(3) unsigned NOT NULL,
  `defender_win_percent` tinyint(3) unsigned NOT NULL,
  `draw_percent` tinyint(3) unsigned NOT NULL,
  `gentime` float unsigned DEFAULT NULL,
  `report` mediumblob,
  `accomplished` tinyint(1) unsigned DEFAULT NULL,
  `message` blob,
  `advanced_system` tinyint(1) unsigned NOT NULL DEFAULT '0',
  PRIMARY KEY (`assaultid`),
  KEY `planetid` (`planetid`),
  KEY `time` (`time`),
  KEY `attacker_lost_res` (`attacker_lost_res`),
  KEY `defender_lost_res` (`defender_lost_res`)
) ENGINE=InnoDB AUTO_INCREMENT=184259 DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_sim_assaultparticipant`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_sim_assaultparticipant` (
  `participantid` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `assaultid` int(10) unsigned DEFAULT NULL,
  `userid` int(10) unsigned DEFAULT NULL,
  `planetid` int(10) unsigned DEFAULT NULL,
  `mode` tinyint(1) unsigned DEFAULT NULL,
  `consumption` double(15,0) DEFAULT NULL,
  `preloaded` double(15,0) DEFAULT NULL,
  `capacity` double(15,0) DEFAULT NULL,
  `haul_metal` double(15,0) DEFAULT NULL,
  `haul_silicon` double(15,0) DEFAULT NULL,
  `haul_hydrogen` double(15,0) DEFAULT NULL,
  `target_unitid` smallint(5) unsigned NOT NULL DEFAULT '0',
  `add_gun_tech` tinyint(3) unsigned NOT NULL DEFAULT '0',
  `add_shield_tech` tinyint(3) unsigned NOT NULL DEFAULT '0',
  `add_shell_tech` tinyint(3) unsigned NOT NULL DEFAULT '0',
  `add_ballistics_tech` tinyint(3) unsigned NOT NULL DEFAULT '0',
  `add_masking_tech` tinyint(3) unsigned NOT NULL DEFAULT '0',
  `add_laser_tech` tinyint(3) unsigned NOT NULL DEFAULT '0',
  `add_ion_tech` tinyint(3) unsigned NOT NULL DEFAULT '0',
  `add_plasma_tech` tinyint(3) unsigned NOT NULL DEFAULT '0',
  PRIMARY KEY (`participantid`),
  KEY `assaultid` (`assaultid`),
  KEY `userid` (`userid`),
  KEY `planetid` (`planetid`),
  CONSTRAINT `na_sim_assaultparticipant_ibfk_2` FOREIGN KEY (`userid`) REFERENCES `na_sim_user` (`userid`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `na_sim_assaultparticipant_ibfk_3` FOREIGN KEY (`planetid`) REFERENCES `na_sim_planet` (`planetid`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `na_sim_assaultparticipant_ibfk_4` FOREIGN KEY (`assaultid`) REFERENCES `na_sim_assault` (`assaultid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=357760 DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_sim_construction`;
/*!50001 DROP VIEW IF EXISTS `na_sim_construction`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_sim_construction` AS SELECT 
 1 AS `buildingid`,
 1 AS `race`,
 1 AS `mode`,
 1 AS `name`,
 1 AS `front`,
 1 AS `ballistics`,
 1 AS `masking`,
 1 AS `basic_metal`,
 1 AS `basic_silicon`,
 1 AS `basic_hydrogen`,
 1 AS `basic_energy`,
 1 AS `prod_metal`,
 1 AS `prod_silicon`,
 1 AS `prod_hydrogen`,
 1 AS `prod_energy`,
 1 AS `cons_metal`,
 1 AS `cons_silicon`,
 1 AS `cons_hydrogen`,
 1 AS `cons_energy`,
 1 AS `charge_metal`,
 1 AS `charge_silicon`,
 1 AS `charge_hydrogen`,
 1 AS `charge_energy`,
 1 AS `special`,
 1 AS `demolish`,
 1 AS `display_order`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_sim_fleet2assault`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_sim_fleet2assault` (
  `assaultid` int(10) unsigned DEFAULT NULL,
  `participantid` int(10) unsigned DEFAULT NULL,
  `userid` int(10) unsigned DEFAULT NULL,
  `unitid` smallint(5) unsigned DEFAULT NULL,
  `mode` tinyint(3) unsigned DEFAULT NULL,
  `quantity` int(10) unsigned DEFAULT NULL,
  `damaged` int(10) unsigned DEFAULT '0',
  `shell_percent` tinyint(3) unsigned DEFAULT '0',
  `grasped` int(10) unsigned DEFAULT '0',
  `org_quantity` int(10) unsigned DEFAULT '0',
  `org_damaged` int(10) unsigned DEFAULT '0',
  `org_shell_percent` tinyint(3) unsigned DEFAULT '0',
  `etc` blob,
  UNIQUE KEY `participantid` (`participantid`,`unitid`),
  KEY `assaultid` (`assaultid`),
  KEY `userid` (`userid`),
  KEY `unitid` (`unitid`),
  CONSTRAINT `na_sim_fleet2assault_ibfk_2` FOREIGN KEY (`participantid`) REFERENCES `na_sim_assaultparticipant` (`participantid`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `na_sim_fleet2assault_ibfk_3` FOREIGN KEY (`userid`) REFERENCES `na_sim_user` (`userid`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `na_sim_fleet2assault_ibfk_4` FOREIGN KEY (`assaultid`) REFERENCES `na_sim_assault` (`assaultid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_sim_galaxy`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_sim_galaxy` (
  `galaxy` int(3) unsigned DEFAULT NULL,
  `system` int(4) unsigned DEFAULT NULL,
  `position` int(2) unsigned DEFAULT NULL,
  `metal` double(15,0) DEFAULT NULL,
  `silicon` double(15,0) DEFAULT NULL,
  `planetid` int(10) unsigned DEFAULT NULL,
  `moonid` int(10) unsigned DEFAULT NULL,
  `destroyed` tinyint(1) unsigned DEFAULT NULL,
  UNIQUE KEY `galaxy` (`galaxy`,`system`,`position`),
  KEY `planetid` (`planetid`),
  KEY `moonid` (`moonid`),
  KEY `system` (`system`),
  CONSTRAINT `na_sim_galaxy_ibfk_1` FOREIGN KEY (`planetid`) REFERENCES `na_sim_planet` (`planetid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_sim_message`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_sim_message` (
  `msgid` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `mode` int(1) unsigned DEFAULT NULL,
  `time` int(10) unsigned DEFAULT NULL,
  `sender` int(10) unsigned DEFAULT NULL,
  `receiver` int(10) unsigned DEFAULT NULL,
  `message` blob,
  `subject` varbinary(255) DEFAULT NULL,
  `readed` tinyint(1) unsigned DEFAULT NULL,
  PRIMARY KEY (`msgid`),
  KEY `sender` (`sender`),
  KEY `receiver` (`receiver`),
  CONSTRAINT `na_sim_message_ibfk_2` FOREIGN KEY (`receiver`) REFERENCES `na_sim_user` (`userid`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `na_sim_message_ibfk_3` FOREIGN KEY (`sender`) REFERENCES `na_sim_user` (`userid`) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_sim_planet`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_sim_planet` (
  `planetid` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `userid` int(9) unsigned DEFAULT NULL,
  `ismoon` tinyint(1) unsigned DEFAULT NULL,
  `planetname` varchar(255) CHARACTER SET utf8 DEFAULT NULL,
  `diameter` int(6) unsigned DEFAULT NULL,
  `picture` varbinary(255) DEFAULT NULL,
  `temperature` int(3) DEFAULT NULL,
  `last` int(10) unsigned DEFAULT NULL,
  `metal` double(128,8) NOT NULL DEFAULT '500.00000000',
  `silicon` double(128,8) NOT NULL DEFAULT '500.00000000',
  `hydrogen` double(128,8) NOT NULL DEFAULT '0.00000000',
  `solar_satellite_prod` int(3) unsigned NOT NULL DEFAULT '100',
  PRIMARY KEY (`planetid`),
  KEY `userid` (`userid`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_sim_rapidfire`;
/*!50001 DROP VIEW IF EXISTS `na_sim_rapidfire`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_sim_rapidfire` AS SELECT 
 1 AS `unitid`,
 1 AS `target`,
 1 AS `value`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_sim_res_log`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_sim_res_log` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `type` int(11) NOT NULL,
  `userid` int(10) unsigned NOT NULL,
  `planetid` int(11) DEFAULT NULL,
  `cnt` int(11) NOT NULL DEFAULT '1',
  `metal` double(15,2) NOT NULL DEFAULT '0.00',
  `silicon` double(15,2) NOT NULL DEFAULT '0.00',
  `hydrogen` double(15,2) NOT NULL DEFAULT '0.00',
  `credit` double(15,2) NOT NULL DEFAULT '0.00',
  `result_metal` double(15,2) DEFAULT NULL,
  `result_silicon` double(15,2) DEFAULT NULL,
  `result_hydrogen` double(15,2) DEFAULT NULL,
  `result_credit` double(15,2) DEFAULT NULL,
  `ownerid` int(11) DEFAULT NULL,
  `event_mode` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `time` (`time`),
  KEY `planetid` (`planetid`),
  KEY `userid` (`userid`,`planetid`,`id`),
  CONSTRAINT `na_sim_res_log_ibfk_1` FOREIGN KEY (`userid`) REFERENCES `na_sim_user` (`userid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_sim_research2user`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_sim_research2user` (
  `buildingid` int(4) unsigned NOT NULL,
  `userid` int(10) unsigned NOT NULL,
  `level` int(3) unsigned NOT NULL,
  PRIMARY KEY (`buildingid`,`userid`),
  KEY `userid` (`userid`),
  CONSTRAINT `na_sim_research2user_ibfk_1` FOREIGN KEY (`userid`) REFERENCES `na_sim_user` (`userid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_sim_ship_datasheet`;
/*!50001 DROP VIEW IF EXISTS `na_sim_ship_datasheet`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_sim_ship_datasheet` AS SELECT 
 1 AS `unitid`,
 1 AS `capicity`,
 1 AS `speed`,
 1 AS `consume`,
 1 AS `attack`,
 1 AS `shield`,
 1 AS `front`,
 1 AS `ballistics`,
 1 AS `masking`,
 1 AS `attacker_attack`,
 1 AS `attacker_shield`,
 1 AS `attacker_front`,
 1 AS `attacker_ballistics`,
 1 AS `attacker_masking`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_sim_user`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_sim_user` (
  `userid` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `username` varchar(32) DEFAULT NULL,
  `email` varchar(255) DEFAULT NULL,
  `temp_email` varchar(255) DEFAULT NULL,
  `languageid` int(4) unsigned DEFAULT NULL,
  `timezone` varbinary(25) DEFAULT NULL,
  `templatepackage` varbinary(128) DEFAULT NULL,
  `imagepackage` varbinary(128) NOT NULL DEFAULT 'std',
  `theme` varbinary(255) DEFAULT NULL,
  `curplanet` int(10) DEFAULT NULL,
  `points` int(10) unsigned DEFAULT NULL,
  `u_points` double NOT NULL DEFAULT '0',
  `r_points` double NOT NULL DEFAULT '0',
  `b_points` double NOT NULL DEFAULT '0',
  `u_count` int(10) unsigned NOT NULL DEFAULT '0',
  `r_count` int(10) unsigned NOT NULL DEFAULT '0',
  `b_count` int(10) unsigned NOT NULL DEFAULT '0',
  `e_points` int(10) unsigned NOT NULL DEFAULT '0',
  `be_points` int(10) unsigned NOT NULL DEFAULT '873',
  `hp` int(10) DEFAULT NULL,
  `battles` int(10) unsigned NOT NULL DEFAULT '0',
  `ipcheck` tinyint(1) unsigned NOT NULL DEFAULT '1',
  `activation` varbinary(32) DEFAULT NULL,
  `regtime` int(10) unsigned DEFAULT NULL,
  `last` int(10) unsigned DEFAULT NULL,
  `asteroid` int(10) unsigned DEFAULT NULL,
  `umode` tinyint(1) unsigned DEFAULT NULL,
  `umodemin` int(10) unsigned DEFAULT NULL,
  `planetorder` tinyint(1) unsigned DEFAULT NULL,
  `delete` int(10) unsigned DEFAULT NULL,
  `esps` tinyint(2) unsigned NOT NULL DEFAULT '1',
  `show_all_constructions` tinyint(3) unsigned NOT NULL DEFAULT '1',
  `show_all_research` tinyint(3) unsigned NOT NULL DEFAULT '1',
  `show_all_shipyard` tinyint(3) unsigned NOT NULL DEFAULT '1',
  `show_all_defense` tinyint(3) unsigned NOT NULL DEFAULT '1',
  PRIMARY KEY (`userid`)
) ENGINE=InnoDB AUTO_INCREMENT=368516 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_sim_user_experience`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_sim_user_experience` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `time` int(11) NOT NULL,
  `isatter` tinyint(3) unsigned NOT NULL,
  `userid` int(10) unsigned NOT NULL,
  `experience` int(11) NOT NULL,
  `assaultid` int(10) unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `assaultid` (`assaultid`),
  KEY `userid` (`userid`),
  KEY `experience` (`experience`),
  CONSTRAINT `na_sim_user_experience_ibfk_1` FOREIGN KEY (`userid`) REFERENCES `na_sim_user` (`userid`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `na_sim_user_experience_ibfk_2` FOREIGN KEY (`assaultid`) REFERENCES `z_na_sim_assault_break` (`assaultid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_social_network_user`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_social_network_user` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `user_id` int(10) unsigned NOT NULL,
  `network_id` int(10) unsigned NOT NULL,
  `network_user_id` varchar(200) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `network_user_id` (`network_user_id`,`user_id`),
  KEY `user_id` (`user_id`),
  CONSTRAINT `na_social_network_user_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `na_user` (`userid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=1029888 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_stargate_jump`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_stargate_jump` (
  `jumpid` int(8) unsigned NOT NULL AUTO_INCREMENT,
  `planetid` int(10) unsigned NOT NULL,
  `time` int(10) unsigned NOT NULL,
  `data` mediumblob NOT NULL,
  PRIMARY KEY (`jumpid`),
  KEY `planetid` (`planetid`),
  CONSTRAINT `na_stargate_jump_ibfk_1` FOREIGN KEY (`planetid`) REFERENCES `na_planet` (`planetid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=1487794 DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_system_free`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_system_free` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=201 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_system_new_active`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_system_new_active` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=601 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_temp_fleet`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_temp_fleet` (
  `planetid` int(10) unsigned NOT NULL,
  `data` blob NOT NULL,
  PRIMARY KEY (`planetid`),
  CONSTRAINT `na_temp_fleet_ibfk_1` FOREIGN KEY (`planetid`) REFERENCES `na_planet` (`planetid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_tournament`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_tournament` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `type` tinyint(3) unsigned NOT NULL,
  `start_time` int(10) unsigned DEFAULT NULL,
  `end_time` int(10) unsigned DEFAULT NULL,
  `start_fleets` smallint(5) unsigned NOT NULL,
  `end_fleets` smallint(5) unsigned DEFAULT NULL,
  `credit_fund` int(10) unsigned NOT NULL,
  `comission` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_tracks`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_tracks` (
  `trackid` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `src` varchar(255) NOT NULL,
  `path` varchar(250) NOT NULL,
  `title` varchar(100) NOT NULL,
  `album` varchar(100) DEFAULT NULL,
  `composer` varchar(100) DEFAULT NULL,
  `original_url` varchar(250) DEFAULT NULL,
  PRIMARY KEY (`trackid`),
  KEY `src` (`src`)
) ENGINE=InnoDB AUTO_INCREMENT=82 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_tutorial_states`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_tutorial_states` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  `display_order` int(10) unsigned NOT NULL,
  `arrow_name` varchar(255) NOT NULL DEFAULT 'left.png',
  `arrow_of` varchar(1023) DEFAULT NULL,
  `arrow_my` varchar(255) NOT NULL DEFAULT 'left center',
  `arrow_at` varchar(255) NOT NULL DEFAULT 'right center',
  `menu_div` varchar(255) NOT NULL,
  `formaction` varchar(255) NOT NULL,
  `dialog_vert` varchar(1023) DEFAULT 'top',
  `dialog_hori` varchar(255) NOT NULL DEFAULT 'center',
  `category` int(10) unsigned NOT NULL DEFAULT '1',
  `modal` tinyint(1) unsigned NOT NULL DEFAULT '0',
  `enabled` tinyint(1) unsigned NOT NULL DEFAULT '1',
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`),
  KEY `display_order` (`display_order`)
) ENGINE=InnoDB AUTO_INCREMENT=22 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_tutorial_states_category`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_tutorial_states_category` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `title` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_unit2shipyard`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_unit2shipyard` (
  `unitid` mediumint(8) unsigned NOT NULL,
  `planetid` int(10) unsigned NOT NULL,
  `quantity` int(10) unsigned NOT NULL,
  `damaged` int(10) unsigned NOT NULL DEFAULT '0',
  `shell_percent` float unsigned NOT NULL DEFAULT '0',
  UNIQUE KEY `unitid` (`unitid`,`planetid`),
  KEY `planetid` (`planetid`),
  KEY `quantity` (`quantity`),
  CONSTRAINT `na_unit2shipyard_ibfk_1` FOREIGN KEY (`planetid`) REFERENCES `na_planet` (`planetid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_unit2shipyard_log`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_unit2shipyard_log` (
  `created` datetime NOT NULL,
  `unitid` int(4) unsigned NOT NULL,
  `planetid` int(10) unsigned NOT NULL,
  `quantity` int(11) NOT NULL,
  `is_adding` tinyint(4) unsigned NOT NULL,
  `new_quantity` int(11) NOT NULL,
  `old_quantity` int(11) NOT NULL,
  `message` varbinary(250) NOT NULL,
  KEY `created` (`created`),
  KEY `unitid` (`unitid`,`planetid`),
  KEY `planetid` (`planetid`,`unitid`)
) ENGINE=InnoDB DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_unit2shipyard_tmp`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_unit2shipyard_tmp` (
  `unitid` mediumint(8) unsigned NOT NULL,
  `planetid` int(10) unsigned NOT NULL,
  `quantity` mediumint(8) unsigned NOT NULL,
  `damaged` mediumint(8) unsigned NOT NULL DEFAULT '0',
  `shell_percent` float unsigned NOT NULL DEFAULT '0',
  UNIQUE KEY `unitid` (`unitid`,`planetid`),
  KEY `planetid` (`planetid`),
  KEY `quantity` (`quantity`)
) ENGINE=MyISAM DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_units_destroyed_stats`;
/*!50001 DROP VIEW IF EXISTS `na_units_destroyed_stats`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_units_destroyed_stats` AS SELECT 
 1 AS `d`,
 1 AS `mode`,
 1 AS `qnt`,
 1 AS `ufo_qnt`,
 1 AS `all_qnt`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_user`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_user` (
  `userid` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `username` varchar(32) CHARACTER SET utf8 NOT NULL,
  `email` varchar(255) CHARACTER SET utf8 NOT NULL,
  `temp_email` varchar(255) CHARACTER SET utf8 NOT NULL,
  `profession` tinyint(3) unsigned NOT NULL DEFAULT '0',
  `prof_time` int(10) unsigned NOT NULL DEFAULT '0',
  `languageid` int(4) unsigned NOT NULL,
  `timezone` varbinary(25) NOT NULL,
  `templatepackage` varbinary(128) NOT NULL,
  `imagepackage` varbinary(128) NOT NULL DEFAULT 'std',
  `theme` varbinary(255) NOT NULL,
  `curplanet` int(10) unsigned DEFAULT NULL,
  `dm_points` double unsigned NOT NULL,
  `points` double NOT NULL DEFAULT '0',
  `max_points` double unsigned NOT NULL DEFAULT '0',
  `u_points` double NOT NULL DEFAULT '0',
  `r_points` double NOT NULL DEFAULT '0',
  `b_points` double NOT NULL DEFAULT '0',
  `u_count` int(10) unsigned NOT NULL DEFAULT '0',
  `r_count` int(10) unsigned NOT NULL DEFAULT '0',
  `b_count` int(10) unsigned NOT NULL DEFAULT '0',
  `e_points` int(10) unsigned NOT NULL DEFAULT '0',
  `be_points` int(10) unsigned NOT NULL DEFAULT '0',
  `of_points` int(10) unsigned NOT NULL DEFAULT '0',
  `of_level` int(10) unsigned NOT NULL DEFAULT '0',
  `a_points` int(10) unsigned NOT NULL DEFAULT '0',
  `a_count` int(10) unsigned NOT NULL DEFAULT '0',
  `hp` int(10) unsigned DEFAULT NULL,
  `battles` int(10) unsigned NOT NULL DEFAULT '0',
  `credit` double(15,2) NOT NULL DEFAULT '5.00',
  `exchange_rate` float NOT NULL DEFAULT '1.2',
  `research_factor` float NOT NULL DEFAULT '1',
  `ipcheck` tinyint(1) unsigned NOT NULL DEFAULT '0',
  `activation` varbinary(32) NOT NULL,
  `password_activation` varbinary(32) NOT NULL DEFAULT '',
  `email_activation` varbinary(32) NOT NULL DEFAULT '',
  `regtime` int(10) unsigned NOT NULL,
  `last` int(10) unsigned NOT NULL,
  `asteroid` int(10) unsigned NOT NULL,
  `umode` tinyint(1) unsigned NOT NULL,
  `umodemin` int(10) unsigned NOT NULL,
  `planetorder` tinyint(1) unsigned NOT NULL,
  `delete` int(10) unsigned NOT NULL,
  `esps` tinyint(2) unsigned NOT NULL DEFAULT '1',
  `show_all_constructions` tinyint(3) unsigned NOT NULL DEFAULT '0',
  `show_all_research` tinyint(3) unsigned NOT NULL DEFAULT '0',
  `show_all_shipyard` tinyint(3) unsigned NOT NULL DEFAULT '0',
  `show_all_defense` tinyint(3) unsigned NOT NULL DEFAULT '0',
  `user_bg_style` varbinary(128) NOT NULL DEFAULT 'us_bg/a-bg-47.css',
  `user_table_style` varbinary(128) NOT NULL DEFAULT 'us_table/table_td2_bg_90.css',
  `skin_type` varbinary(64) NOT NULL DEFAULT '',
  `race` tinyint(3) unsigned NOT NULL DEFAULT '1',
  `user_agreement_read` int(10) unsigned NOT NULL DEFAULT '0',
  `tutorial_state` int(10) unsigned NOT NULL DEFAULT '1',
  `tutorial_show` tinyint(1) unsigned NOT NULL DEFAULT '1',
  `last_chat` int(10) unsigned NOT NULL DEFAULT '0',
  `last_chatally` int(10) unsigned NOT NULL DEFAULT '0',
  `chat_languageid` int(4) unsigned DEFAULT NULL,
  `planet_teleport_time` int(11) DEFAULT NULL,
  `observer` tinyint(1) unsigned NOT NULL DEFAULT '0',
  `protection_time` int(10) unsigned NOT NULL DEFAULT '0',
  PRIMARY KEY (`userid`),
  UNIQUE KEY `username` (`username`),
  UNIQUE KEY `email` (`email`),
  KEY `curplanet` (`curplanet`),
  KEY `regtime` (`regtime`),
  KEY `battles` (`battles`),
  KEY `hp` (`hp`),
  KEY `epoints` (`e_points`),
  KEY `fpoints` (`u_count`),
  KEY `rpoints` (`r_count`),
  KEY `last` (`last`),
  KEY `points` (`points`,`last`),
  KEY `last_chat` (`last_chat`),
  KEY `last_chatally` (`last_chatally`),
  KEY `credit` (`credit`),
  KEY `dm_points` (`dm_points`),
  KEY `observer` (`observer`),
  KEY `umode` (`umode`,`observer`,`userid`),
  KEY `max_points` (`max_points`),
  CONSTRAINT `na_user_ibfk_1` FOREIGN KEY (`curplanet`) REFERENCES `na_planet` (`planetid`) ON DELETE SET NULL ON UPDATE CASCADE,
  CONSTRAINT `na_user_ibfk_2` FOREIGN KEY (`hp`) REFERENCES `na_planet` (`planetid`) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=2094833 DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_user2ally`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_user2ally` (
  `userid` int(9) unsigned NOT NULL,
  `aid` int(8) unsigned NOT NULL,
  `joindate` int(10) unsigned NOT NULL,
  `rank` int(12) unsigned NOT NULL,
  UNIQUE KEY `userid` (`userid`),
  KEY `aid` (`aid`),
  CONSTRAINT `na_user2ally_ibfk_1` FOREIGN KEY (`userid`) REFERENCES `na_user` (`userid`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `na_user2ally_ibfk_2` FOREIGN KEY (`aid`) REFERENCES `na_alliance` (`aid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_user2exchange`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_user2exchange` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `user_id` int(10) unsigned NOT NULL DEFAULT '0',
  `exchange_id` int(10) unsigned NOT NULL DEFAULT '0',
  `ban_event_id` int(10) unsigned NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_user2group`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_user2group` (
  `usergroupid` int(3) unsigned NOT NULL,
  `userid` int(10) unsigned NOT NULL,
  `data` varbinary(255) NOT NULL,
  KEY `userid` (`userid`),
  KEY `usergroup` (`usergroupid`),
  CONSTRAINT `na_user2group_ibfk_1` FOREIGN KEY (`usergroupid`) REFERENCES `na_usergroup` (`usergroupid`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `na_user2group_ibfk_2` FOREIGN KEY (`userid`) REFERENCES `na_user` (`userid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_user_agreement`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_user_agreement` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `type` tinyint(3) unsigned NOT NULL DEFAULT '1',
  `text` text NOT NULL,
  `lang` int(10) unsigned NOT NULL DEFAULT '1',
  `date` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
  `display_order` int(10) unsigned NOT NULL DEFAULT '10',
  `parent_id` int(10) unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `parent_id` (`parent_id`),
  KEY `display_order` (`display_order`,`id`),
  CONSTRAINT `na_user_agreement_ibfk_1` FOREIGN KEY (`parent_id`) REFERENCES `na_user_agreement` (`id`) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=150 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_user_copy`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_user_copy` (
  `userid` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `username` varchar(32) CHARACTER SET utf8 NOT NULL,
  `email` varchar(255) CHARACTER SET utf8 NOT NULL,
  `temp_email` varchar(255) CHARACTER SET utf8 NOT NULL,
  `profession` tinyint(3) unsigned NOT NULL DEFAULT '0',
  `prof_time` int(10) unsigned NOT NULL DEFAULT '0',
  `languageid` int(4) unsigned NOT NULL,
  `timezone` varbinary(25) NOT NULL,
  `templatepackage` varbinary(128) NOT NULL,
  `imagepackage` varbinary(128) NOT NULL DEFAULT 'std',
  `theme` varbinary(255) NOT NULL,
  `curplanet` int(10) NOT NULL,
  `points` double NOT NULL DEFAULT '0',
  `u_points` double NOT NULL DEFAULT '0',
  `r_points` double NOT NULL DEFAULT '0',
  `b_points` double NOT NULL DEFAULT '0',
  `u_count` int(10) unsigned NOT NULL DEFAULT '0',
  `r_count` int(10) unsigned NOT NULL DEFAULT '0',
  `b_count` int(10) unsigned NOT NULL DEFAULT '0',
  `e_points` int(10) unsigned NOT NULL DEFAULT '0',
  `be_points` int(10) unsigned NOT NULL DEFAULT '0',
  `of_points` int(10) unsigned NOT NULL DEFAULT '0',
  `of_level` int(10) unsigned NOT NULL DEFAULT '0',
  `a_points` int(10) unsigned NOT NULL DEFAULT '0',
  `a_count` int(10) unsigned NOT NULL DEFAULT '0',
  `hp` int(10) NOT NULL,
  `battles` int(10) unsigned NOT NULL DEFAULT '0',
  `credit` double(15,2) NOT NULL DEFAULT '5.00',
  `credit_check` double(15,2) DEFAULT NULL,
  `credit_check2` double(15,2) DEFAULT NULL,
  `exchange_rate` float NOT NULL DEFAULT '1.2',
  `research_factor` float NOT NULL DEFAULT '1',
  `ipcheck` tinyint(1) unsigned NOT NULL DEFAULT '0',
  `activation` varbinary(32) NOT NULL,
  `password_activation` varbinary(32) NOT NULL DEFAULT '',
  `email_activation` varbinary(32) NOT NULL DEFAULT '',
  `regtime` int(10) unsigned NOT NULL,
  `last` int(10) unsigned NOT NULL,
  `asteroid` int(10) unsigned NOT NULL,
  `umode` tinyint(1) unsigned NOT NULL,
  `umodemin` int(10) unsigned NOT NULL,
  `planetorder` tinyint(1) unsigned NOT NULL,
  `delete` int(10) unsigned NOT NULL,
  `esps` tinyint(2) unsigned NOT NULL DEFAULT '1',
  `show_all_constructions` tinyint(3) unsigned NOT NULL DEFAULT '0',
  `show_all_research` tinyint(3) unsigned NOT NULL DEFAULT '0',
  `show_all_shipyard` tinyint(3) unsigned NOT NULL DEFAULT '0',
  `show_all_defense` tinyint(3) unsigned NOT NULL DEFAULT '0',
  `user_bg_style` varbinary(128) NOT NULL DEFAULT 'us_bg/a-bg-47.css',
  `user_table_style` varbinary(128) NOT NULL DEFAULT 'us_table/table_td2_bg_90.css',
  `skin_type` varbinary(64) NOT NULL DEFAULT '',
  `race` tinyint(3) unsigned NOT NULL DEFAULT '1',
  `user_agreement_read` int(10) unsigned NOT NULL DEFAULT '0',
  `tutorial_state` int(10) unsigned NOT NULL DEFAULT '1',
  `tutorial_show` tinyint(1) unsigned NOT NULL DEFAULT '0',
  `last_chat` int(10) unsigned NOT NULL DEFAULT '0',
  `last_chatally` int(10) unsigned NOT NULL DEFAULT '0',
  `chat_languageid` int(4) unsigned DEFAULT NULL,
  `planet_teleport_time` int(11) DEFAULT NULL,
  PRIMARY KEY (`userid`),
  UNIQUE KEY `username` (`username`),
  UNIQUE KEY `email` (`email`),
  KEY `curplanet` (`curplanet`),
  KEY `regtime` (`regtime`),
  KEY `battles` (`battles`),
  KEY `hp` (`hp`),
  KEY `epoints` (`e_points`),
  KEY `fpoints` (`u_count`),
  KEY `rpoints` (`r_count`),
  KEY `last` (`last`),
  KEY `points` (`points`,`last`),
  KEY `last_chat` (`last_chat`),
  KEY `last_chatally` (`last_chatally`),
  KEY `credit` (`credit`)
) ENGINE=InnoDB AUTO_INCREMENT=1339042 DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_user_experience`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_user_experience` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `time` int(11) NOT NULL,
  `isatter` tinyint(3) unsigned NOT NULL,
  `userid` int(10) unsigned NOT NULL,
  `experience` int(11) NOT NULL,
  `assaultid` int(10) unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `assaultid` (`assaultid`),
  KEY `userid` (`userid`),
  KEY `experience` (`experience`),
  CONSTRAINT `na_user_experience_ibfk_1` FOREIGN KEY (`userid`) REFERENCES `na_user` (`userid`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `na_user_experience_ibfk_2` FOREIGN KEY (`assaultid`) REFERENCES `na_assault` (`assaultid`) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=1497793 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_user_ext`;
/*!50001 DROP VIEW IF EXISTS `na_user_ext`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_user_ext` AS SELECT 
 1 AS `r`,
 1 AS `l`,
 1 AS `snid`,
 1 AS `userid`,
 1 AS `username`,
 1 AS `email`,
 1 AS `temp_email`,
 1 AS `languageid`,
 1 AS `timezone`,
 1 AS `templatepackage`,
 1 AS `imagepackage`,
 1 AS `theme`,
 1 AS `curplanet`,
 1 AS `points`,
 1 AS `u_points`,
 1 AS `r_points`,
 1 AS `b_points`,
 1 AS `u_count`,
 1 AS `r_count`,
 1 AS `b_count`,
 1 AS `e_points`,
 1 AS `be_points`,
 1 AS `of_points`,
 1 AS `of_level`,
 1 AS `a_points`,
 1 AS `a_count`,
 1 AS `hp`,
 1 AS `battles`,
 1 AS `credit`,
 1 AS `exchange_rate`,
 1 AS `research_factor`,
 1 AS `ipcheck`,
 1 AS `activation`,
 1 AS `password_activation`,
 1 AS `email_activation`,
 1 AS `regtime`,
 1 AS `last`,
 1 AS `asteroid`,
 1 AS `umode`,
 1 AS `umodemin`,
 1 AS `planetorder`,
 1 AS `delete`,
 1 AS `esps`,
 1 AS `show_all_constructions`,
 1 AS `show_all_research`,
 1 AS `show_all_shipyard`,
 1 AS `show_all_defense`,
 1 AS `user_bg_style`,
 1 AS `user_table_style`,
 1 AS `skin_type`,
 1 AS `race`,
 1 AS `user_agreement_read`,
 1 AS `tutorial_state`,
 1 AS `tutorial_show`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_user_imgpak_ext`;
/*!50001 DROP VIEW IF EXISTS `na_user_imgpak_ext`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_user_imgpak_ext` AS SELECT 
 1 AS `imagepackage`,
 1 AS `cnt`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_user_online`;
/*!50001 DROP VIEW IF EXISTS `na_user_online`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_user_online` AS SELECT 
 1 AS `online_5_h`,
 1 AS `online_10_h`,
 1 AS `online_15_h`,
 1 AS `online_30_h`,
 1 AS `core_1_h`,
 1 AS `core_2_h`,
 1 AS `core_7_h`,
 1 AS `online_5`,
 1 AS `online_10`,
 1 AS `online_15`,
 1 AS `online_30`,
 1 AS `core_1`,
 1 AS `core_2`,
 1 AS `core_7`,
 1 AS `all_users`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_user_reg_stats`;
/*!50001 DROP VIEW IF EXISTS `na_user_reg_stats`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_user_reg_stats` AS SELECT 
 1 AS `d`,
 1 AS `snid`,
 1 AS `cnt`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_user_reg_stats_month`;
/*!50001 DROP VIEW IF EXISTS `na_user_reg_stats_month`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_user_reg_stats_month` AS SELECT 
 1 AS `d`,
 1 AS `snid`,
 1 AS `cnt`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_user_states`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_user_states` (
  `user_id` int(10) unsigned NOT NULL,
  `simulated_assault` int(10) unsigned NOT NULL DEFAULT '0',
  `exchanged_ress` int(10) unsigned NOT NULL DEFAULT '0',
  `unknown` int(10) unsigned NOT NULL DEFAULT '0',
  PRIMARY KEY (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_user_tmp`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_user_tmp` (
  `userid` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `username` varchar(32) CHARACTER SET utf8 NOT NULL,
  `email` varchar(255) CHARACTER SET utf8 NOT NULL,
  `temp_email` varchar(255) CHARACTER SET utf8 NOT NULL,
  `languageid` int(4) unsigned NOT NULL,
  `timezone` varbinary(25) NOT NULL,
  `templatepackage` varbinary(128) NOT NULL,
  `imagepackage` varbinary(128) NOT NULL DEFAULT 'std',
  `theme` varbinary(255) NOT NULL,
  `curplanet` int(10) NOT NULL,
  `points` double NOT NULL DEFAULT '0',
  `u_points` double NOT NULL DEFAULT '0',
  `r_points` double NOT NULL DEFAULT '0',
  `b_points` double NOT NULL DEFAULT '0',
  `u_count` int(10) unsigned NOT NULL DEFAULT '0',
  `r_count` int(10) unsigned NOT NULL DEFAULT '0',
  `b_count` int(10) unsigned NOT NULL DEFAULT '0',
  `e_points` int(10) unsigned NOT NULL DEFAULT '0',
  `be_points` int(10) unsigned NOT NULL DEFAULT '0',
  `of_points` int(10) unsigned NOT NULL DEFAULT '0',
  `of_level` int(10) unsigned NOT NULL DEFAULT '0',
  `a_points` int(10) unsigned NOT NULL DEFAULT '0',
  `a_count` int(10) unsigned NOT NULL DEFAULT '0',
  `hp` int(10) NOT NULL,
  `battles` int(10) unsigned NOT NULL DEFAULT '0',
  `credit` double(15,2) NOT NULL DEFAULT '5.00',
  `exchange_rate` float NOT NULL DEFAULT '1.2',
  `research_factor` float NOT NULL DEFAULT '1',
  `ipcheck` tinyint(1) unsigned NOT NULL DEFAULT '0',
  `activation` varbinary(32) NOT NULL,
  `password_activation` varbinary(32) NOT NULL DEFAULT '',
  `email_activation` varbinary(32) NOT NULL DEFAULT '',
  `regtime` int(10) unsigned NOT NULL,
  `last` int(10) unsigned NOT NULL,
  `asteroid` int(10) unsigned NOT NULL,
  `umode` tinyint(1) unsigned NOT NULL,
  `umodemin` int(10) unsigned NOT NULL,
  `planetorder` tinyint(1) unsigned NOT NULL,
  `delete` int(10) unsigned NOT NULL,
  `esps` tinyint(2) unsigned NOT NULL DEFAULT '1',
  `show_all_constructions` tinyint(3) unsigned NOT NULL DEFAULT '1',
  `show_all_research` tinyint(3) unsigned NOT NULL DEFAULT '1',
  `show_all_shipyard` tinyint(3) unsigned NOT NULL DEFAULT '1',
  `show_all_defense` tinyint(3) unsigned NOT NULL DEFAULT '1',
  `user_bg_style` varbinary(128) NOT NULL DEFAULT 'us_bg/a-bg-25.css',
  `user_table_style` varbinary(128) NOT NULL DEFAULT 'us_table/table_std_bg_80.css',
  `skin_type` varbinary(64) NOT NULL DEFAULT '',
  `race` tinyint(3) unsigned NOT NULL DEFAULT '1',
  `user_agreement_read` int(10) unsigned NOT NULL DEFAULT '0',
  PRIMARY KEY (`userid`),
  UNIQUE KEY `username` (`username`),
  UNIQUE KEY `email` (`email`),
  KEY `curplanet` (`curplanet`),
  KEY `regtime` (`regtime`),
  KEY `battles` (`battles`),
  KEY `hp` (`hp`),
  KEY `epoints` (`e_points`),
  KEY `fpoints` (`u_count`),
  KEY `rpoints` (`r_count`),
  KEY `last` (`last`),
  KEY `points` (`points`,`last`)
) ENGINE=MyISAM DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_usergroup`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_usergroup` (
  `usergroupid` int(3) unsigned NOT NULL AUTO_INCREMENT,
  `grouptitle` varbinary(128) NOT NULL,
  `standard` tinyint(1) unsigned NOT NULL,
  PRIMARY KEY (`usergroupid`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=binary;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_view_max_building_stats`;
/*!50001 DROP VIEW IF EXISTS `na_view_max_building_stats`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_view_max_building_stats` AS SELECT 
 1 AS `buildingid`,
 1 AS `userid`,
 1 AS `max_level`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_view_sum_unit_stats`;
/*!50001 DROP VIEW IF EXISTS `na_view_sum_unit_stats`*/;
SET @saved_cs_client     = @@character_set_client;
SET character_set_client = utf8;
/*!50001 CREATE VIEW `na_view_sum_unit_stats` AS SELECT 
 1 AS `unitid`,
 1 AS `userid`,
 1 AS `sum_quantity`*/;
SET character_set_client = @saved_cs_client;
DROP TABLE IF EXISTS `na_yii_cron`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_yii_cron` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `level` varchar(128) DEFAULT NULL,
  `category` varchar(128) DEFAULT NULL,
  `logtime` int(11) DEFAULT NULL,
  `message` text,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM AUTO_INCREMENT=13 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_yii_log`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_yii_log` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `level` varchar(128) DEFAULT NULL,
  `category` varchar(128) DEFAULT NULL,
  `logtime` int(11) DEFAULT NULL,
  `message` text,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=24 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `na_yii_log_info`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `na_yii_log_info` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `level` varchar(128) DEFAULT NULL,
  `category` varchar(128) DEFAULT NULL,
  `logtime` int(11) DEFAULT NULL,
  `message` text,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=32677 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `o_yii_error_log`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `o_yii_error_log` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `level` varchar(128) DEFAULT NULL,
  `category` varchar(128) DEFAULT NULL,
  `logtime` int(11) DEFAULT NULL,
  `message` text,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `o_yii_warning_log`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `o_yii_warning_log` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `level` varchar(128) DEFAULT NULL,
  `category` varchar(128) DEFAULT NULL,
  `logtime` int(11) DEFAULT NULL,
  `message` text,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `yii_error_log`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `yii_error_log` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `level` varchar(128) DEFAULT NULL,
  `category` varchar(128) DEFAULT NULL,
  `logtime` int(11) DEFAULT NULL,
  `message` text,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `yii_warning_log`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `yii_warning_log` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `level` varchar(128) DEFAULT NULL,
  `category` varchar(128) DEFAULT NULL,
  `logtime` int(11) DEFAULT NULL,
  `message` text,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!50001 DROP VIEW IF EXISTS `na_artefact_probobility`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_artefact_probobility` AS select `a`.`id` AS `type`,`a`.`name` AS `name`,`a`.`quota` AS `quota`,`a`.`art_count` AS `art_count`,`a`.`exp_count` AS `exp_count`,`a`.`total_user` AS `total_user`,(select count(0) from `na_artefact2user` `a2u` where ((`a2u`.`typeid` = `a`.`id`) and (`a2u`.`deleted` > (unix_timestamp(now()) - (60 * ((60 * 5) + ((60 * 12) * rand()))))))) AS `deleted_count`,round((`a`.`total_user` * `a`.`quota`),0) AS `quota_count`,greatest(0,((round((`a`.`total_user` * `a`.`quota`),0) - greatest(`a`.`art_count`,`a`.`exp_count`)) - (select count(0) from `na_artefact2user` `a2u` where ((`a2u`.`typeid` = `a`.`id`) and (`a2u`.`deleted` > (unix_timestamp(now()) - (60 * ((60 * 5) + ((60 * 12) * rand()))))))))) AS `probobility` from `na_artefact_used` `a` */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_artefact_used`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_artefact_used` AS select `c`.`buildingid` AS `id`,`c`.`name` AS `name`,`d`.`quota` AS `quota`,(select count(0) AS `count(*)` from `na_artefact2user` `a` where ((`a`.`typeid` = `d`.`typeid`) and (`a`.`deleted` = 0))) AS `all_count`,(select count(0) AS `count(*)` from `na_artefact2user` `a` where ((`a`.`typeid` = `d`.`typeid`) and (`a`.`deleted` = 0) and (`a`.`bought` = 1))) AS `buy_count`,(select count(0) AS `count(*)` from `na_artefact2user` `a` where ((`a`.`typeid` = `d`.`typeid`) and (`a`.`deleted` = 0) and (`a`.`bought` <> 1))) AS `art_count`,(select count(0) AS `count(*)` from `na_expedition_stats` `e` where ((`e`.`artefact_type` = `d`.`typeid`) and (`e`.`time` > (unix_timestamp() - 604800)))) AS `exp_count`,(select count(0) AS `count(*)` from `na_user` `u` where ((`u`.`points` > 1000) and (`u`.`last` > (unix_timestamp() - (((7 * 24) * 60) * 60))))) AS `total_user` from (`na_artefact_datasheet` `d` left join `na_construction` `c` on((`c`.`buildingid` = `d`.`typeid`))) */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_assault_ext`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_assault_ext` AS select from_unixtime(`a`.`time`) AS `t`,ifnull(`g`.`galaxy`,`gm`.`galaxy`) AS `g`,ifnull(`g`.`system`,`gm`.`system`) AS `s`,ifnull(`g`.`position`,`gm`.`position`) AS `p`,`a`.`assaultid` AS `assaultid`,`a`.`key` AS `key`,`a`.`key2` AS `key2`,`a`.`result` AS `result`,`a`.`planetid` AS `planetid`,`a`.`time` AS `time`,`a`.`target_moon` AS `target_moon`,`a`.`target_buildingid` AS `target_buildingid`,`a`.`building_level` AS `building_level`,`a`.`building_metal` AS `building_metal`,`a`.`building_silicon` AS `building_silicon`,`a`.`building_hydrogen` AS `building_hydrogen`,`a`.`building_destroyed` AS `building_destroyed`,`a`.`target_destroyed` AS `target_destroyed`,`a`.`attacker_explode` AS `attacker_explode`,`a`.`moon_allow_type` AS `moon_allow_type`,`a`.`moonchance` AS `moonchance`,`a`.`moon` AS `moon`,`a`.`attacker_lost_res` AS `attacker_lost_res`,`a`.`attacker_lost_metal` AS `attacker_lost_metal`,`a`.`attacker_lost_silicon` AS `attacker_lost_silicon`,`a`.`attacker_lost_hydrogen` AS `attacker_lost_hydrogen`,`a`.`defender_lost_res` AS `defender_lost_res`,`a`.`defender_lost_metal` AS `defender_lost_metal`,`a`.`defender_lost_silicon` AS `defender_lost_silicon`,`a`.`defender_lost_hydrogen` AS `defender_lost_hydrogen`,`a`.`debris_metal` AS `debris_metal`,`a`.`debris_silicon` AS `debris_silicon`,`a`.`planet_metal` AS `planet_metal`,`a`.`planet_silicon` AS `planet_silicon`,`a`.`planet_hydrogen` AS `planet_hydrogen`,`a`.`haul_metal` AS `haul_metal`,`a`.`haul_silicon` AS `haul_silicon`,`a`.`haul_hydrogen` AS `haul_hydrogen`,`a`.`lostunits_attacker` AS `lostunits_attacker`,`a`.`lostunits_defender` AS `lostunits_defender`,`a`.`attacker_exp` AS `attacker_exp`,`a`.`defender_exp` AS `defender_exp`,`a`.`turns` AS `turns`,`a`.`gentime` AS `gentime`,`a`.`report` AS `report`,`a`.`accomplished` AS `accomplished`,`a`.`message` AS `message`,`a`.`advanced_system` AS `advanced_system` from ((`na_assault` `a` left join `na_galaxy` `g` on((`g`.`planetid` = `a`.`planetid`))) left join `na_galaxy` `gm` on((`gm`.`moonid` = `a`.`planetid`))) order by `a`.`assaultid` desc */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_ban_u_ext`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_ban_u_ext` AS select `u`.`r` AS `r`,`u`.`username` AS `username`,`b`.`banid` AS `banid`,`b`.`userid` AS `userid`,`b`.`from` AS `from`,`b`.`to` AS `to`,`b`.`reason` AS `reason`,`b`.`admin_comment` AS `admin_comment` from (`na_ban_u` `b` left join `na_user_ext` `u` on((`u`.`userid` = `b`.`userid`))) order by `b`.`banid` desc */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_chat2ally_ext`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_chat2ally_ext` AS select from_unixtime(`a`.`time`) AS `t`,`u`.`username` AS `username`,`a`.`messageid` AS `messageid`,`a`.`time` AS `time`,`a`.`userid` AS `userid`,`a`.`allyid` AS `allyid`,`a`.`message` AS `message` from (`na_chat2ally` `a` join `na_user` `u` on((`u`.`userid` = `a`.`userid`))) order by `a`.`time` desc */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_chat2ally_stat`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_chat2ally_stat` AS select `na_chat2ally`.`allyid` AS `allyid`,count(0) AS `cnt` from `na_chat2ally` where (`na_chat2ally`.`time` > unix_timestamp((now() - interval 3 day))) group by `na_chat2ally`.`allyid` order by count(0) desc */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_chat_ext`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_chat_ext` AS select from_unixtime(`c`.`time`) AS `t`,`u`.`username` AS `username`,`c`.`messageid` AS `messageid`,`c`.`time` AS `time`,`c`.`userid` AS `userid`,`c`.`message` AS `message` from (`na_chat` `c` left join `na_user` `u` on((`u`.`userid` = `c`.`userid`))) order by `c`.`messageid` desc */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_cronjob_ext`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_cronjob_ext` AS select from_unixtime(`c`.`xtime`) AS `x`,from_unixtime(`c`.`last`) AS `l`,`c`.`cronid` AS `cronid`,`c`.`script` AS `script`,`c`.`month` AS `month`,`c`.`day` AS `day`,`c`.`weekday` AS `weekday`,`c`.`hour` AS `hour`,`c`.`minute` AS `minute`,`c`.`xtime` AS `xtime`,`c`.`last` AS `last`,`c`.`active` AS `active` from `na_cronjob` `c` order by `c`.`last` desc */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_empty_systems`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_empty_systems` AS select `g`.`id` AS `galaxy`,`s`.`id` AS `system`,0 AS `cnt`,(select min(`na_planet_free`.`planet_free_pos`) AS `min(planet_free_pos)` from `na_planet_free`) AS `free_pos` from ((`na_system_free` `s` join `na_galaxy_active` `g` on((1 = 1))) left join `na_galaxy` `g2` on(((`g2`.`system` = `s`.`id`) and (`g2`.`galaxy` = `g`.`id`)))) where isnull(`g2`.`system`) order by `g`.`id`,`s`.`id` limit 2 */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_event_aliens`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_event_aliens` AS select `na_events_ext`.`s` AS `s`,`na_events_ext`.`t` AS `t`,`na_events_ext`.`p` AS `p`,`na_events_ext`.`eventid` AS `eventid`,`na_events_ext`.`mode` AS `mode`,`na_events_ext`.`start` AS `start`,`na_events_ext`.`time` AS `time`,`na_events_ext`.`planetid` AS `planetid`,`na_events_ext`.`user` AS `user`,`na_events_ext`.`destination` AS `destination`,`na_events_ext`.`data` AS `data`,`na_events_ext`.`protected` AS `protected`,`na_events_ext`.`prev_rc` AS `prev_rc`,`na_events_ext`.`processed` AS `processed`,`na_events_ext`.`processed_mode` AS `processed_mode`,`na_events_ext`.`processed_time` AS `processed_time`,`na_events_ext`.`processed_dt` AS `processed_dt`,`na_events_ext`.`processed_quantity` AS `processed_quantity`,`na_events_ext`.`required_quantity` AS `required_quantity`,`na_events_ext`.`error_message` AS `error_message`,`na_events_ext`.`ally_eventid` AS `ally_eventid`,`na_events_ext`.`parent_eventid` AS `parent_eventid`,`na_events_ext`.`org_data` AS `org_data` from `na_events_ext` where ((`na_events_ext`.`mode` in (33,34,35,36)) and (`na_events_ext`.`user` = 0) and (`na_events_ext`.`processed` = 0)) */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_event_dest`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_event_dest` AS select `e`.`eventid` AS `eventid`,`e`.`mode` AS `mode`,`e`.`start` AS `start`,`e`.`time` AS `time`,`e`.`planetid` AS `planetid`,`e`.`user` AS `user`,`e`.`destination` AS `destination`,`e`.`data` AS `data`,`e`.`protected` AS `protected`,`e`.`prev_rc` AS `prev_rc`,`e`.`processed` AS `processed`,`e`.`processed_mode` AS `processed_mode`,`e`.`processed_time` AS `processed_time`,`e`.`processed_dt` AS `processed_dt`,`e`.`processed_quantity` AS `processed_quantity`,`e`.`required_quantity` AS `required_quantity`,`e`.`error_message` AS `error_message`,`e`.`ally_eventid` AS `ally_eventid`,`e`.`parent_eventid` AS `parent_eventid`,`e`.`artid` AS `artid`,`e`.`org_data` AS `org_data`,NULL AS `startuserid`,NULL AS `startusername`,`u1`.`userid` AS `userid`,`u1`.`username` AS `username`,`p1`.`planetname` AS `planetname`,ifnull(`g1`.`galaxy`,`g1m`.`galaxy`) AS `galaxy`,ifnull(`g1`.`system`,`g1m`.`system`) AS `system`,ifnull(`g1`.`position`,`g1m`.`position`) AS `position`,`u2`.`userid` AS `destuserid`,`u2`.`username` AS `destname`,`p2`.`planetname` AS `destplanet`,ifnull(`g2`.`galaxy`,`g2m`.`galaxy`) AS `galaxy2`,ifnull(`g2`.`system`,`g2m`.`system`) AS `system2`,ifnull(`g2`.`position`,`g2m`.`position`) AS `position2` from ((((((((`na_events` `e` join `na_planet` `p2` on((`p2`.`planetid` = `e`.`destination`))) join `na_user` `u2` on((`u2`.`userid` = `p2`.`userid`))) left join `na_galaxy` `g2` on((`g2`.`planetid` = `e`.`destination`))) left join `na_galaxy` `g2m` on((`g2m`.`moonid` = `e`.`destination`))) left join `na_planet` `p1` on((`p1`.`planetid` = `e`.`planetid`))) left join `na_user` `u1` on((`u1`.`userid` = `e`.`user`))) left join `na_galaxy` `g1` on((`g1`.`planetid` = `e`.`planetid`))) left join `na_galaxy` `g1m` on((`g1m`.`moonid` = `e`.`planetid`))) */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_event_src`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_event_src` AS select `e`.`eventid` AS `eventid`,`e`.`mode` AS `mode`,`e`.`start` AS `start`,`e`.`time` AS `time`,`e`.`planetid` AS `planetid`,`e`.`user` AS `user`,`e`.`destination` AS `destination`,`e`.`data` AS `data`,`e`.`protected` AS `protected`,`e`.`prev_rc` AS `prev_rc`,`e`.`processed` AS `processed`,`e`.`processed_mode` AS `processed_mode`,`e`.`processed_time` AS `processed_time`,`e`.`processed_dt` AS `processed_dt`,`e`.`processed_quantity` AS `processed_quantity`,`e`.`required_quantity` AS `required_quantity`,`e`.`error_message` AS `error_message`,`e`.`ally_eventid` AS `ally_eventid`,`e`.`parent_eventid` AS `parent_eventid`,`e`.`artid` AS `artid`,`e`.`org_data` AS `org_data`,`p1user`.`userid` AS `startuserid`,`p1user`.`username` AS `startusername`,`e`.`user` AS `userid`,`u1`.`username` AS `username`,`p1`.`planetname` AS `planetname`,ifnull(`g1`.`galaxy`,`g1m`.`galaxy`) AS `galaxy`,ifnull(`g1`.`system`,`g1m`.`system`) AS `system`,ifnull(`g1`.`position`,`g1m`.`position`) AS `position`,NULL AS `destuserid`,NULL AS `destname`,NULL AS `destplanet`,NULL AS `galaxy2`,NULL AS `system2`,NULL AS `position2` from (((((`na_events` `e` left join `na_user` `u1` on((`u1`.`userid` = `e`.`user`))) left join `na_planet` `p1` on((`p1`.`planetid` = `e`.`planetid`))) left join `na_user` `p1user` on((`p1user`.`userid` = `p1`.`userid`))) left join `na_galaxy` `g1` on((`g1`.`planetid` = `e`.`planetid`))) left join `na_galaxy` `g1m` on((`g1m`.`moonid` = `e`.`planetid`))) */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_events_ext`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_events_ext` AS select from_unixtime(`e`.`start`) AS `s`,from_unixtime(`e`.`time`) AS `t`,from_unixtime(`e`.`processed_time`) AS `p`,`e`.`eventid` AS `eventid`,`e`.`mode` AS `mode`,`e`.`start` AS `start`,`e`.`time` AS `time`,`e`.`planetid` AS `planetid`,`e`.`user` AS `user`,`e`.`destination` AS `destination`,`e`.`data` AS `data`,`e`.`protected` AS `protected`,`e`.`prev_rc` AS `prev_rc`,`e`.`processed` AS `processed`,`e`.`processed_mode` AS `processed_mode`,`e`.`processed_time` AS `processed_time`,`e`.`processed_dt` AS `processed_dt`,`e`.`processed_quantity` AS `processed_quantity`,`e`.`required_quantity` AS `required_quantity`,`e`.`error_message` AS `error_message`,`e`.`ally_eventid` AS `ally_eventid`,`e`.`parent_eventid` AS `parent_eventid`,`e`.`org_data` AS `org_data` from `na_events` `e` order by `e`.`eventid` desc */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_events_stat`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_events_stat` AS select cast(from_unixtime(`e`.`processed_time`) as date) AS `d`,`e`.`mode` AS `mode`,count(0) AS `cnt`,sum(`e`.`processed_quantity`) AS `qnt` from `na_events` `e` where ((`e`.`mode` <> 20) and (`e`.`processed` = 3)) group by cast(from_unixtime(`e`.`processed_time`) as date),`e`.`mode` order by cast(from_unixtime(`e`.`processed_time`) as date) desc,sum(`e`.`processed_quantity`) desc,count(0) desc,`e`.`mode` */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_exchange_stats`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_exchange_stats` AS select cast(from_unixtime(`na_exchange_lots`.`sold_date`) as date) AS `d`,sum(`na_exchange_lots`.`price`) AS `price`,sum(`na_exchange_lots`.`payed_exchange`) AS `payed_exchange`,count(0) AS `cnt` from `na_exchange_lots` where (`na_exchange_lots`.`type` = 2) group by cast(from_unixtime(`na_exchange_lots`.`sold_date`) as date) order by cast(from_unixtime(`na_exchange_lots`.`sold_date`) as date) desc */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_expedition_stats_day`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_expedition_stats_day` AS select cast(from_unixtime(`na_expedition_stats`.`time`) as date) AS `d`,`na_expedition_stats`.`type` AS `type`,count(0) AS `cnt` from `na_expedition_stats` group by cast(from_unixtime(`na_expedition_stats`.`time`) as date),`na_expedition_stats`.`type` order by cast(from_unixtime(`na_expedition_stats`.`time`) as date) desc,count(0) desc */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_expedition_stats_ext`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_expedition_stats_ext` AS select from_unixtime(`s`.`time`) AS `t`,`u`.`username` AS `username`,`c`.`name` AS `art_name`,`s`.`statid` AS `statid`,`s`.`userid` AS `userid`,`s`.`time` AS `time`,`s`.`completed` AS `completed`,`s`.`galaxy` AS `galaxy`,`s`.`system` AS `system`,`s`.`type` AS `type`,`s`.`percent` AS `percent`,`s`.`message` AS `message`,`s`.`artefact_type` AS `artefact_type`,`s`.`found_credit` AS `found_credit`,`s`.`found_metal` AS `found_metal`,`s`.`found_silicon` AS `found_silicon`,`s`.`found_hydrogen` AS `found_hydrogen`,`s`.`event_id` AS `event_id` from ((`na_expedition_stats` `s` join `na_user` `u` on((`u`.`userid` = `s`.`userid`))) left join `na_construction` `c` on((`c`.`buildingid` = `s`.`artefact_type`))) order by `s`.`statid` desc */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_expedition_used`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_expedition_used` AS select cast(from_unixtime(`e`.`time`) as date) AS `d`,`e`.`type` AS `type`,count(0) AS `cnt`,((count(0) * 100) / (select count(0) from `na_expedition_stats` where (cast(from_unixtime(`na_expedition_stats`.`time`) as date) = `d`))) AS `percent`,avg(`e`.`percent`) AS `avg_percent` from `na_expedition_stats` `e` group by cast(from_unixtime(`e`.`time`) as date),`e`.`type` order by cast(from_unixtime(`e`.`time`) as date) desc,count(0) desc limit 100 */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_galaxy_empty_new_pos`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_galaxy_empty_new_pos` AS select `g`.`id` AS `galaxy`,`s`.`id` AS `system`,(select `na_planet_new_active`.`position` from `na_planet_new_active` order by rand() limit 1) AS `position` from ((`na_system_new_active` `s` join `na_galaxy_new_active` `g` on((1 = 1))) left join `na_galaxy` `g2` on(((`g2`.`system` = `s`.`id`) and (`g2`.`galaxy` = `g`.`id`)))) where isnull(`g2`.`system`) order by `g`.`id`,`s`.`id` limit 2 */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_galaxy_empty_new_pos2`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_galaxy_empty_new_pos2` AS select `g`.`id` AS `galaxy`,`s`.`id` AS `system`,(select `na_planet_new_active`.`position` from `na_planet_new_active` order by rand() limit 1) AS `position` from ((`na_system_new_active` `s` join `na_galaxy_new_active` `g` on((1 = 1))) left join `na_galaxy` `g2` on(((`g2`.`system` = `s`.`id`) and (`g2`.`galaxy` = `g`.`id`)))) where isnull(`g2`.`system`) */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_galaxy_empty_new_pos_all`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_galaxy_empty_new_pos_all` AS select `g`.`id` AS `galaxy`,`s`.`id` AS `system`,(select `na_planet_new_active`.`position` from `na_planet_new_active` order by rand() limit 1) AS `position` from ((`na_system_new_active` `s` join `na_galaxy_new_active` `g` on((1 = 1))) left join `na_galaxy` `g2` on(((`g2`.`system` = `s`.`id`) and (`g2`.`galaxy` = `g`.`id`)))) where isnull(`g2`.`system`) order by `g`.`id`,`s`.`id` */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_galaxy_empty_new_pos_sum`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_galaxy_empty_new_pos_sum` AS select sum(`na_galaxy_empty_new_pos_all`.`position`) AS `sum_empty` from `na_galaxy_empty_new_pos_all` */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_galaxy_free`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_galaxy_free` AS select `na_free_planets`.`galaxy` AS `galaxy`,`na_free_planets`.`system` AS `system`,`na_free_planets`.`cnt` AS `cnt`,`na_free_planets`.`free_pos` AS `free_pos` from `na_free_planets` union select `na_empty_systems`.`galaxy` AS `galaxy`,`na_empty_systems`.`system` AS `system`,`na_empty_systems`.`cnt` AS `cnt`,`na_empty_systems`.`free_pos` AS `free_pos` from `na_empty_systems` order by `galaxy`,rand() */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_galaxy_free_pos`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_galaxy_free_pos` AS select `gf`.`galaxy` AS `galaxy`,`gf`.`system` AS `system`,`gf`.`cnt` AS `cnt`,`pf`.`position` AS `position` from ((`na_galaxy_with_free_pos` `gf` join `na_planet_new_active` `pf` on((1 = 1))) left join `na_galaxy` `g` on(((`g`.`galaxy` = `gf`.`galaxy`) and (`g`.`system` = `gf`.`system`) and (`pf`.`position` = `g`.`position`)))) where isnull(`g`.`position`) order by rand() limit 2 */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_galaxy_free_pos2`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_galaxy_free_pos2` AS select `gf`.`galaxy` AS `galaxy`,`gf`.`system` AS `system`,`pf`.`position` AS `position`,`gf`.`cnt` AS `cnt`,`gf`.`free_cnt` AS `free_cnt` from ((`na_galaxy_with_free_pos2` `gf` join `na_planet_new_active` `pf` on((1 = 1))) left join `na_galaxy` `g` on(((`g`.`galaxy` = `gf`.`galaxy`) and (`g`.`system` = `gf`.`system`) and (`pf`.`position` = `g`.`position`)))) where isnull(`g`.`position`) */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_galaxy_free_pos_rnd2`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_galaxy_free_pos_rnd2` AS select `na_galaxy_free_pos2`.`galaxy` AS `galaxy`,`na_galaxy_free_pos2`.`system` AS `system`,`na_galaxy_free_pos2`.`position` AS `position` from `na_galaxy_free_pos2` order by rand() */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_galaxy_free_pos_rnd_cut2`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_galaxy_free_pos_rnd_cut2` AS select `na_galaxy_free_pos2`.`galaxy` AS `galaxy`,`na_galaxy_free_pos2`.`system` AS `system`,`na_galaxy_free_pos2`.`position` AS `position` from `na_galaxy_free_pos2` order by rand() limit 50 */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_galaxy_new_pos`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_galaxy_new_pos` AS select `na_galaxy_free_pos`.`galaxy` AS `galaxy`,`na_galaxy_free_pos`.`system` AS `system`,`na_galaxy_free_pos`.`position` AS `position` from `na_galaxy_free_pos` union select `na_galaxy_empty_new_pos`.`galaxy` AS `galaxy`,`na_galaxy_empty_new_pos`.`system` AS `system`,`na_galaxy_empty_new_pos`.`position` AS `position` from `na_galaxy_empty_new_pos` order by rand() */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_galaxy_new_pos_cut2`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_galaxy_new_pos_cut2` AS select `na_galaxy_new_pos_union2`.`galaxy` AS `galaxy`,`na_galaxy_new_pos_union2`.`system` AS `system`,`na_galaxy_new_pos_union2`.`position` AS `position`,`na_galaxy_new_pos_union2`.`type` AS `type` from `oxsar_db`.`na_galaxy_new_pos_union2` order by rand() limit 100 */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_galaxy_new_pos_sum2`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_galaxy_new_pos_sum2` AS select (select sum(`na_galaxy_with_destroyed2`.`generic_cnt`) from `na_galaxy_with_destroyed2` where `na_galaxy_with_destroyed2`.`galaxy` in (select `na_galaxy_new_active`.`id` from `na_galaxy_new_active`)) AS `destroyed_planet_cnt`,(select sum(`na_galaxy_with_free_pos2`.`free_cnt`) from `na_galaxy_with_free_pos2`) AS `free_planet_cnt`,(select count(0) from `na_galaxy_empty_new_pos2`) AS `empty_galaxy_cnt`,((select count(0) from `na_galaxy_empty_new_pos2`) * (select ceiling((count(0) / 2)) from `na_planet_new_active`)) AS `empty_planet_cnt` */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_galaxy_new_pos_union2`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_galaxy_new_pos_union2` AS select `a`.`galaxy` AS `galaxy`,`a`.`system` AS `system`,`a`.`position` AS `position`,`a`.`type` AS `type` from (select `na_galaxy_empty_new_pos2`.`galaxy` AS `galaxy`,`na_galaxy_empty_new_pos2`.`system` AS `system`,`na_galaxy_empty_new_pos2`.`position` AS `position`,'empty' AS `type` from `oxsar_db`.`na_galaxy_empty_new_pos2` limit 40) `a` union select `b`.`galaxy` AS `galaxy`,`b`.`system` AS `system`,`b`.`position` AS `position`,`b`.`type` AS `type` from (select `na_galaxy_free_pos_rnd_cut2`.`galaxy` AS `galaxy`,`na_galaxy_free_pos_rnd_cut2`.`system` AS `system`,`na_galaxy_free_pos_rnd_cut2`.`position` AS `position`,'free' AS `type` from `oxsar_db`.`na_galaxy_free_pos_rnd_cut2` limit 50) `b` */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_galaxy_with_destroyed`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_galaxy_with_destroyed` AS select `na_galaxy`.`galaxy` AS `galaxy`,count(0) AS `cnt` from `na_galaxy` where (`na_galaxy`.`destroyed` = 1) group by `na_galaxy`.`galaxy` */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_galaxy_with_destroyed2`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_galaxy_with_destroyed2` AS select `g`.`galaxy` AS `galaxy`,sum((case when (`p`.`position` is not null) then 1 else 0 end)) AS `generic_cnt`,sum((case when (`p`.`position` is not null) then 0 else 1 end)) AS `ufo_cnt` from (`na_galaxy` `g` left join `na_planet_new_active` `p` on((`g`.`position` = `p`.`position`))) where (`g`.`destroyed` = 1) group by `g`.`galaxy` */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_galaxy_with_free_pos`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_galaxy_with_free_pos` AS select `g`.`galaxy` AS `galaxy`,`g`.`system` AS `system`,count(0) AS `cnt` from ((`na_galaxy` `g` join `na_planet_new_active` `pf` on((`pf`.`position` = `g`.`position`))) join `na_galaxy_new_active` `gf` on((`gf`.`id` = `g`.`galaxy`))) group by `g`.`galaxy`,`g`.`system` having (count(0) <= (select ceiling((count(0) / 2)) from `na_planet_new_active`)) order by rand() limit 2 */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_galaxy_with_free_pos2`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_galaxy_with_free_pos2` AS select `g`.`galaxy` AS `galaxy`,`g`.`system` AS `system`,count(0) AS `cnt`,(((select ceiling((count(0) / 2)) from `na_planet_new_active`) - count(0)) + 1) AS `free_cnt` from ((`na_galaxy` `g` join `na_planet_new_active` `pf` on((`pf`.`position` = `g`.`position`))) join `na_galaxy_new_active` `gf` on((`gf`.`id` = `g`.`galaxy`))) group by `g`.`galaxy`,`g`.`system` having (count(0) <= (select ceiling((count(0) / 2)) from `na_planet_new_active`)) */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_galaxy_with_free_pos_all`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_galaxy_with_free_pos_all` AS select (((select ceiling((count(0) / 2)) from `na_planet_new_active`) - count(0)) + 1) AS `cnt` from ((`na_galaxy` `g` join `na_planet_new_active` `pf` on((`pf`.`position` = `g`.`position`))) join `na_galaxy_new_active` `gf` on((`gf`.`id` = `g`.`galaxy`))) group by `g`.`galaxy`,`g`.`system` having (count(0) <= (select ceiling((count(0) / 2)) from `na_planet_new_active`)) */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_galaxy_with_free_pos_sum`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_galaxy_with_free_pos_sum` AS select sum(`na_galaxy_with_free_pos_all`.`cnt`) AS `sum_free` from `na_galaxy_with_free_pos_all` */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_log_error_index_ext`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_log_error_index_ext` AS select from_unixtime(`l`.`logtime`) AS `t`,`l`.`id` AS `id`,`l`.`level` AS `level`,`l`.`category` AS `category`,`l`.`logtime` AS `logtime`,`l`.`message` AS `message` from `na_log_error_index` `l` order by `l`.`id` desc */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_log_error_odnoklassniki_ext`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_log_error_odnoklassniki_ext` AS select from_unixtime(`l`.`logtime`) AS `t`,`l`.`id` AS `id`,`l`.`level` AS `level`,`l`.`category` AS `category`,`l`.`logtime` AS `logtime`,`l`.`message` AS `message` from `na_log_error_odnoklassniki` `l` order by `l`.`id` desc */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_log_warning_index_ext`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_log_warning_index_ext` AS select from_unixtime(`l`.`logtime`) AS `t`,`l`.`id` AS `id`,`l`.`level` AS `level`,`l`.`category` AS `category`,`l`.`logtime` AS `logtime`,`l`.`message` AS `message` from `na_log_warning_index` `l` order by `l`.`id` desc */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_log_warning_main_ext`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_log_warning_main_ext` AS select from_unixtime(`l`.`logtime`) AS `t`,`l`.`id` AS `id`,`l`.`level` AS `level`,`l`.`category` AS `category`,`l`.`logtime` AS `logtime`,`l`.`message` AS `message` from `na_log_warning_main` `l` order by `l`.`id` desc */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_log_warning_odnoklassniki_ext`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_log_warning_odnoklassniki_ext` AS select from_unixtime(`l`.`logtime`) AS `t`,`l`.`id` AS `id`,`l`.`level` AS `level`,`l`.`category` AS `category`,`l`.`logtime` AS `logtime`,`l`.`message` AS `message` from `na_log_warning_odnoklassniki` `l` order by `l`.`id` desc */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_message_ext`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_message_ext` AS select from_unixtime(`m`.`time`) AS `t`,`u1`.`username` AS `s`,`u2`.`username` AS `r`,`m`.`msgid` AS `msgid`,`m`.`mode` AS `mode`,`m`.`time` AS `time`,`m`.`sender` AS `sender`,`m`.`receiver` AS `receiver`,`m`.`related_user` AS `related_user`,`m`.`message` AS `message`,`m`.`subject` AS `subject`,`m`.`readed` AS `readed` from ((`na_message` `m` left join `na_user` `u1` on((`u1`.`userid` = `m`.`sender`))) left join `na_user` `u2` on((`u2`.`userid` = `m`.`receiver`))) order by `m`.`msgid` desc */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_moon_creation_stats`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_moon_creation_stats` AS select `na_assault`.`moonchance` AS `moonchance`,count(0) AS `cnt`,round(((count(0) * `na_assault`.`moonchance`) / 100),0) AS `calc_moons`,sum(`na_assault`.`moon`) AS `real_moons` from `na_assault` where (`na_assault`.`moonchance` > 0) group by `na_assault`.`moonchance` order by `na_assault`.`moonchance` desc */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_moon_destroy_stats`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_moon_destroy_stats` AS select count(0) AS `cnt`,sum(`na_assault_ext`.`target_destroyed`) AS `target_destroyed`,round(((sum(`na_assault_ext`.`target_destroyed`) * 100) / count(0)),0) AS `ptd`,sum(`na_assault_ext`.`attacker_explode`) AS `attacker_explode`,round(((sum(`na_assault_ext`.`attacker_explode`) * 100) / count(0)),0) AS `pae` from `na_assault_ext` where ((`na_assault_ext`.`target_moon` = 1) and (`na_assault_ext`.`t` >= '20110623')) */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_moon_destroy_stats_old`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_moon_destroy_stats_old` AS select count(0) AS `cnt`,sum(`na_assault_ext`.`target_destroyed`) AS `target_destroyed`,round(((sum(`na_assault_ext`.`target_destroyed`) * 100) / count(0)),0) AS `ptd`,sum(`na_assault_ext`.`attacker_explode`) AS `attacker_explode`,round(((sum(`na_assault_ext`.`attacker_explode`) * 100) / count(0)),0) AS `pae` from `na_assault_ext` where ((`na_assault_ext`.`target_moon` = 1) and (`na_assault_ext`.`t` >= '20110518') and (`na_assault_ext`.`t` < '20110609')) */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_moon_destroy_stats_old2`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_moon_destroy_stats_old2` AS select count(0) AS `cnt`,sum(`na_assault_ext`.`target_destroyed`) AS `target_destroyed`,round(((sum(`na_assault_ext`.`target_destroyed`) * 100) / count(0)),0) AS `ptd`,sum(`na_assault_ext`.`attacker_explode`) AS `attacker_explode`,round(((sum(`na_assault_ext`.`attacker_explode`) * 100) / count(0)),0) AS `pae` from `na_assault_ext` where ((`na_assault_ext`.`target_moon` = 1) and (`na_assault_ext`.`t` >= '20110609') and (`na_assault_ext`.`t` < '20110623')) */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_payment_stats`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_payment_stats` AS select cast(`p`.`pay_date` as date) AS `d`,sum(`p`.`pay_credit`) AS `credit`,round(sum((case when (`p`.`pay_type` = 'Mailru') then (`p`.`pay_amount_r` / 0.7) when (`p`.`pay_type` = 'Vkontakte') then (`p`.`pay_amount_r` * 6.4) else `p`.`pay_amount_r` end)),2) AS `amount_r`,count(0) AS `cnt`,sum((case when (`p`.`pay_type` = 'Odnoklassniki') then `p`.`pay_amount_r` else 0 end)) AS `odnk_r`,sum((case when (`p`.`pay_type` = 'Vkontakte') then `p`.`pay_amount_r` else 0 end)) AS `vk_r`,sum((case when (`p`.`pay_type` = 'Mailru') then `p`.`pay_amount_r` else 0 end)) AS `mailru_r`,sum((case when (`p`.`pay_type` not in ('Odnoklassniki','Mailru','Vkontakte')) then `p`.`pay_amount_r` else 0 end)) AS `oxsar_r`,sum((case when (`p`.`pay_type` = 'Odnoklassniki') then round((`p`.`pay_amount_r` * (case when (`p`.`pay_date` < '2011-09-01') then 0.5 else 0.42 end)),2) when (`p`.`pay_type` = 'Mailru') then round((`p`.`pay_amount_r` * 1.0),2) when (`p`.`pay_type` = 'Vkontakte') then round((`p`.`pay_amount_r` * 3.2),2) else `p`.`pay_amount_r` end)) AS `real_r` from `na_payments_ext` `p` where (`p`.`pay_status` = 1) group by cast(`p`.`pay_date` as date) order by cast(`p`.`pay_date` as date) desc */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_payment_stats_month`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_payment_stats_month` AS select extract(year_month from `p`.`pay_date`) AS `d`,sum(`p`.`pay_credit`) AS `credit`,round(sum((case when (`p`.`pay_type` = 'Mailru') then (`p`.`pay_amount_r` / 0.7) when (`p`.`pay_type` = 'Vkontakte') then (`p`.`pay_amount_r` * 6.4) else `p`.`pay_amount_r` end)),2) AS `amount_r`,count(0) AS `cnt`,sum((case when (`p`.`pay_type` = 'Odnoklassniki') then `p`.`pay_amount_r` else 0 end)) AS `odnk_r`,sum((case when (`p`.`pay_type` = 'Vkontakte') then `p`.`pay_amount_r` else 0 end)) AS `vk_r`,sum((case when (`p`.`pay_type` = 'Mailru') then `p`.`pay_amount_r` else 0 end)) AS `mailru_r`,sum((case when (`p`.`pay_type` not in ('Odnoklassniki','Mailru','Vkontakte')) then `p`.`pay_amount_r` else 0 end)) AS `oxsar_r`,sum((case when (`p`.`pay_type` = 'Odnoklassniki') then round((`p`.`pay_amount_r` * (case when (`p`.`pay_date` < '2011-09-01') then 0.5 else 0.42 end)),2) when (`p`.`pay_type` = 'Mailru') then round((`p`.`pay_amount_r` * 1.0),2) when (`p`.`pay_type` = 'Vkontakte') then round((`p`.`pay_amount_r` * 3.2),2) else `p`.`pay_amount_r` end)) AS `real_r` from `na_payments_ext` `p` where (`p`.`pay_status` = 1) group by extract(year_month from `p`.`pay_date`) order by extract(year_month from `p`.`pay_date`) desc */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_payment_user_stats`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_payment_user_stats` AS select cast(`p`.`pay_date` as date) AS `d`,`u`.`userid` AS `userid`,`u`.`username` AS `username`,sum(`p`.`pay_credit`) AS `credit`,round(sum((case when (`p`.`pay_type` = 'Mailru') then (`p`.`pay_amount_r` / 0.7) when (`p`.`pay_type` = 'Vkontakte') then (`p`.`pay_amount_r` * 6.4) else `p`.`pay_amount_r` end)),2) AS `amount_r`,count(0) AS `cnt`,sum((case when (`p`.`pay_type` = 'Odnoklassniki') then `p`.`pay_amount_r` else 0 end)) AS `odnk_r`,sum((case when (`p`.`pay_type` = 'Vkontakte') then `p`.`pay_amount_r` else 0 end)) AS `vk_r`,sum((case when (`p`.`pay_type` = 'Mailru') then `p`.`pay_amount_r` else 0 end)) AS `mailru_r`,sum((case when (`p`.`pay_type` not in ('Odnoklassniki','Mailru','Vkontakte')) then `p`.`pay_amount_r` else 0 end)) AS `oxsar_r`,sum((case when (`p`.`pay_type` = 'Odnoklassniki') then round((`p`.`pay_amount_r` * (case when (`p`.`pay_date` < '2011-09-01') then 0.5 else 0.42 end)),2) when (`p`.`pay_type` = 'Mailru') then round((`p`.`pay_amount_r` * 1.0),2) when (`p`.`pay_type` = 'Vkontakte') then round((`p`.`pay_amount_r` * 3.2),2) else `p`.`pay_amount_r` end)) AS `real_r` from (`na_payments_ext` `p` left join `na_user` `u` on((`u`.`userid` = `p`.`pay_user_id`))) where (`p`.`pay_status` = 1) group by cast(`p`.`pay_date` as date),`p`.`pay_user_id` order by cast(`p`.`pay_date` as date) desc,sum((case when (`p`.`pay_type` = 'Odnoklassniki') then round((`p`.`pay_amount_r` * (case when (`p`.`pay_date` < '2011-09-01') then 0.5 else 0.42 end)),2) when (`p`.`pay_type` = 'Mailru') then round((`p`.`pay_amount_r` * 1.0),2) when (`p`.`pay_type` = 'Vkontakte') then round((`p`.`pay_amount_r` * 3.2),2) else `p`.`pay_amount_r` end)) desc */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_payment_user_stats_month`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_payment_user_stats_month` AS select extract(year_month from `p`.`pay_date`) AS `d`,`u`.`userid` AS `userid`,`u`.`username` AS `username`,sum(`p`.`pay_credit`) AS `credit`,round(sum((case when (`p`.`pay_type` = 'Mailru') then (`p`.`pay_amount_r` / 0.7) when (`p`.`pay_type` = 'Vkontakte') then (`p`.`pay_amount_r` * 6.4) else `p`.`pay_amount_r` end)),2) AS `amount_r`,count(0) AS `cnt`,sum((case when (`p`.`pay_type` = 'Odnoklassniki') then `p`.`pay_amount_r` else 0 end)) AS `odnk_r`,sum((case when (`p`.`pay_type` = 'Vkontakte') then `p`.`pay_amount_r` else 0 end)) AS `vk_r`,sum((case when (`p`.`pay_type` = 'Mailru') then `p`.`pay_amount_r` else 0 end)) AS `mailru_r`,sum((case when (`p`.`pay_type` not in ('Odnoklassniki','Mailru','Vkontakte')) then `p`.`pay_amount_r` else 0 end)) AS `oxsar_r`,sum((case when (`p`.`pay_type` = 'Odnoklassniki') then round((`p`.`pay_amount_r` * (case when (`p`.`pay_date` < '2011-09-01') then 0.5 else 0.42 end)),2) when (`p`.`pay_type` = 'Mailru') then round((`p`.`pay_amount_r` * 1.0),2) when (`p`.`pay_type` = 'Vkontakte') then round((`p`.`pay_amount_r` * 3.2),2) else `p`.`pay_amount_r` end)) AS `real_r` from (`na_payments_ext` `p` left join `na_user` `u` on((`u`.`userid` = `p`.`pay_user_id`))) where (`p`.`pay_status` = 1) group by extract(year_month from `p`.`pay_date`),`p`.`pay_user_id` order by extract(year_month from `p`.`pay_date`) desc,sum((case when (`p`.`pay_type` = 'Odnoklassniki') then round((`p`.`pay_amount_r` * (case when (`p`.`pay_date` < '2011-09-01') then 0.5 else 0.42 end)),2) when (`p`.`pay_type` = 'Mailru') then round((`p`.`pay_amount_r` * 1.0),2) when (`p`.`pay_type` = 'Vkontakte') then round((`p`.`pay_amount_r` * 3.2),2) else `p`.`pay_amount_r` end)) desc */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_payments_ext`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_payments_ext` AS select `u`.`username` AS `username`,`u`.`credit` AS `credit`,`p`.`pay_id` AS `pay_id`,`p`.`pay_user_id` AS `pay_user_id`,`p`.`pay_type` AS `pay_type`,`p`.`pay_from` AS `pay_from`,`p`.`pay_amount` AS `pay_amount`,`p`.`pay_amount_r` AS `pay_amount_r`,`p`.`pay_credit` AS `pay_credit`,`p`.`pay_date` AS `pay_date`,`p`.`pay_status` AS `pay_status` from (`na_payments` `p` left join `na_user` `u` on((`p`.`pay_user_id` = `u`.`userid`))) order by `p`.`pay_date` desc */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_referral_ext`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_referral_ext` AS select `u1`.`username` AS `u`,`u2`.`username` AS `r`,from_unixtime(`r`.`ref_time`) AS `rt`,`sn2`.`network_id` AS `rsn`,`r`.`userid` AS `userid`,`r`.`ref_id` AS `ref_id`,`u2`.`points` AS `ref_points`,`r`.`ref_time` AS `ref_time`,`r`.`ref_ip` AS `ref_ip`,`r`.`bonus` AS `bonus` from (((`na_referral` `r` left join `na_user` `u1` on((`u1`.`userid` = `r`.`userid`))) left join `na_user` `u2` on((`u2`.`userid` = `r`.`ref_id`))) left join `na_social_network_user` `sn2` on((`sn2`.`user_id` = `r`.`ref_id`))) order by `r`.`ref_time` desc */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_res_log_game_credit`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_res_log_game_credit` AS select cast(`na_res_log`.`time` as date) AS `d`,min(`na_res_log`.`game_credit`) AS `min_game_credit`,max(`na_res_log`.`game_credit`) AS `max_game_credit` from `na_res_log` where (`na_res_log`.`result_credit` is not null) group by cast(`na_res_log`.`time` as date) desc */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_res_log_gift_stats`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_res_log_gift_stats` AS select cast(`na_res_log`.`time` as date) AS `d`,max(`na_res_log`.`credit`) AS `max`,sum(`na_res_log`.`credit`) AS `c`,count(0) AS `cnt` from `na_res_log` where ((`na_res_log`.`type` = 28) and (`na_res_log`.`credit` > 1)) group by cast(`na_res_log`.`time` as date),`na_res_log`.`type` order by cast(`na_res_log`.`time` as date) desc */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_res_log_grab_stats`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_res_log_grab_stats` AS select cast(`na_res_log`.`time` as date) AS `d`,min(`na_res_log`.`credit`) AS `min`,sum(`na_res_log`.`credit`) AS `c`,count(0) AS `cnt` from `na_res_log` where (`na_res_log`.`type` = 27) group by cast(`na_res_log`.`time` as date),`na_res_log`.`type` order by cast(`na_res_log`.`time` as date) desc */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_res_log_hack`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_res_log_hack` AS select `l`.`id` AS `id`,`l`.`time` AS `time`,`l`.`type` AS `type`,`l`.`userid` AS `userid`,`l`.`planetid` AS `planetid`,`l`.`cnt` AS `cnt`,`l`.`metal` AS `metal`,`l`.`silicon` AS `silicon`,`l`.`hydrogen` AS `hydrogen`,`l`.`credit` AS `credit`,`l`.`result_metal` AS `result_metal`,`l`.`result_silicon` AS `result_silicon`,`l`.`result_hydrogen` AS `result_hydrogen`,`l`.`result_credit` AS `result_credit`,`l`.`ownerid` AS `ownerid`,`l`.`event_mode` AS `event_mode`,`e`.`planetid` AS `planetid1`,`e`.`user` AS `userid1`,`p1`.`planetname` AS `planetname1`,`u1`.`username` AS `username1`,`u1`.`userid` AS `userid_r1`,`e`.`destination` AS `destination`,`p2`.`planetname` AS `planet2`,`u2`.`username` AS `username2`,`u2`.`userid` AS `user_r2` from (((((`na_res_log` `l` left join `na_events` `e` on((`e`.`eventid` = `l`.`ownerid`))) left join `na_planet` `p1` on((`p1`.`planetid` = `e`.`planetid`))) left join `na_user` `u1` on((`u1`.`userid` = `p1`.`userid`))) left join `na_planet` `p2` on((`p2`.`planetid` = `e`.`destination`))) left join `na_user` `u2` on((`u2`.`userid` = `p2`.`userid`))) where (((`l`.`metal` < -(1)) or (`l`.`silicon` < -(1)) or (`l`.`hydrogen` < -(1))) and (`l`.`type` in (1,5,6,9))) order by `l`.`time` desc */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_res_log_premium_stats`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_res_log_premium_stats` AS select cast(`na_res_log`.`time` as date) AS `d`,abs(sum(`na_res_log`.`credit`)) AS `credit`,count(0) AS `cnt` from `na_res_log` where (`na_res_log`.`type` = 29) group by cast(`na_res_log`.`time` as date) order by cast(`na_res_log`.`time` as date) desc */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_res_log_stats`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_res_log_stats` AS select cast(`na_res_log`.`time` as date) AS `d`,sum(greatest(0,`na_res_log`.`credit`)) AS `plus`,sum(least(0,`na_res_log`.`credit`)) AS `minus`,sum(`na_res_log`.`credit`) AS `summary` from `na_res_log` where (`na_res_log`.`credit` <> 0) group by cast(`na_res_log`.`time` as date) order by cast(`na_res_log`.`time` as date) desc */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_res_log_stats_month`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_res_log_stats_month` AS select extract(year_month from `na_res_log`.`time`) AS `d`,sum(greatest(0,`na_res_log`.`credit`)) AS `plus`,sum(least(0,`na_res_log`.`credit`)) AS `minus`,sum(`na_res_log`.`credit`) AS `summary` from `na_res_log` where (`na_res_log`.`credit` <> 0) group by extract(year_month from `na_res_log`.`time`) order by extract(year_month from `na_res_log`.`time`) desc */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_res_log_type_stats`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_res_log_type_stats` AS select cast(`na_res_log`.`time` as date) AS `d`,`na_res_log`.`type` AS `type`,sum(`na_res_log`.`credit`) AS `c` from `na_res_log` where (`na_res_log`.`credit` <> 0) group by cast(`na_res_log`.`time` as date),`na_res_log`.`type` order by cast(`na_res_log`.`time` as date) desc,abs(sum(`na_res_log`.`credit`)) desc */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_sessions_ext`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_sessions_ext` AS select from_unixtime(`s`.`time`) AS `from_unixtime(s.time)`,`u`.`username` AS `username`,`s`.`sessionid` AS `sessionid`,`s`.`userid` AS `userid`,`s`.`ipaddress` AS `ipaddress`,`s`.`useragent` AS `useragent`,`s`.`time` AS `time`,`s`.`logged` AS `logged`,`s`.`is_admin` AS `is_admin` from (`na_sessions` `s` join `na_user` `u` on((`u`.`userid` = `s`.`userid`))) order by `s`.`time` desc */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_ships_log`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_ships_log` AS select `l`.`created` AS `created`,`l`.`old_quantity` AS `old_quantity`,`l`.`quantity` AS `quantity`,`l`.`is_adding` AS `is_adding`,`l`.`new_quantity` AS `new_quantity`,`l`.`message` AS `message`,`l`.`unitid` AS `unitid`,`c`.`name` AS `name`,`ph`.`content` AS `content`,`l`.`planetid` AS `planetid`,`pl`.`planetname` AS `planetname`,`u`.`userid` AS `userid`,`u`.`username` AS `username`,`ga`.`galaxy` AS `galaxy`,`ga`.`system` AS `system`,`ga`.`position` AS `position`,`ga`.`moonid` AS `moonid` from (((((`na_unit2shipyard_log` `l` left join `na_construction` `c` on((`l`.`unitid` = `c`.`buildingid`))) left join `na_phrases` `ph` on(((`ph`.`title` = `c`.`name`) and (`ph`.`phrasegroupid` = 4)))) left join `na_planet` `pl` on((`pl`.`planetid` = `l`.`planetid`))) left join `na_galaxy` `ga` on((`ga`.`planetid` = `l`.`planetid`))) left join `na_user` `u` on((`u`.`userid` = `pl`.`userid`))) order by `l`.`created` desc */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_sim_construction`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_sim_construction` AS select `na_construction`.`buildingid` AS `buildingid`,`na_construction`.`race` AS `race`,`na_construction`.`mode` AS `mode`,`na_construction`.`name` AS `name`,`na_construction`.`front` AS `front`,`na_construction`.`ballistics` AS `ballistics`,`na_construction`.`masking` AS `masking`,`na_construction`.`basic_metal` AS `basic_metal`,`na_construction`.`basic_silicon` AS `basic_silicon`,`na_construction`.`basic_hydrogen` AS `basic_hydrogen`,`na_construction`.`basic_energy` AS `basic_energy`,`na_construction`.`prod_metal` AS `prod_metal`,`na_construction`.`prod_silicon` AS `prod_silicon`,`na_construction`.`prod_hydrogen` AS `prod_hydrogen`,`na_construction`.`prod_energy` AS `prod_energy`,`na_construction`.`cons_metal` AS `cons_metal`,`na_construction`.`cons_silicon` AS `cons_silicon`,`na_construction`.`cons_hydrogen` AS `cons_hydrogen`,`na_construction`.`cons_energy` AS `cons_energy`,`na_construction`.`charge_metal` AS `charge_metal`,`na_construction`.`charge_silicon` AS `charge_silicon`,`na_construction`.`charge_hydrogen` AS `charge_hydrogen`,`na_construction`.`charge_energy` AS `charge_energy`,`na_construction`.`special` AS `special`,`na_construction`.`demolish` AS `demolish`,`na_construction`.`display_order` AS `display_order` from `na_construction` */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_sim_rapidfire`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_sim_rapidfire` AS select `na_rapidfire`.`unitid` AS `unitid`,`na_rapidfire`.`target` AS `target`,`na_rapidfire`.`value` AS `value` from `na_rapidfire` */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_sim_ship_datasheet`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_sim_ship_datasheet` AS select `na_ship_datasheet`.`unitid` AS `unitid`,`na_ship_datasheet`.`capicity` AS `capicity`,`na_ship_datasheet`.`speed` AS `speed`,`na_ship_datasheet`.`consume` AS `consume`,`na_ship_datasheet`.`attack` AS `attack`,`na_ship_datasheet`.`shield` AS `shield`,`na_ship_datasheet`.`front` AS `front`,`na_ship_datasheet`.`ballistics` AS `ballistics`,`na_ship_datasheet`.`masking` AS `masking`,`na_ship_datasheet`.`attacker_attack` AS `attacker_attack`,`na_ship_datasheet`.`attacker_shield` AS `attacker_shield`,`na_ship_datasheet`.`attacker_front` AS `attacker_front`,`na_ship_datasheet`.`attacker_ballistics` AS `attacker_ballistics`,`na_ship_datasheet`.`attacker_masking` AS `attacker_masking` from `na_ship_datasheet` */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_units_destroyed_stats`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_units_destroyed_stats` AS select cast(from_unixtime(`a`.`time`) as date) AS `d`,`c`.`mode` AS `mode`,(sum((case when (`f`.`userid` is not null) then `f`.`org_quantity` else 0 end)) - sum((case when (`f`.`userid` is not null) then `f`.`quantity` else 0 end))) AS `qnt`,(sum((case when isnull(`f`.`userid`) then `f`.`org_quantity` else 0 end)) - sum((case when isnull(`f`.`userid`) then `f`.`quantity` else 0 end))) AS `ufo_qnt`,(sum(`f`.`org_quantity`) - sum(`f`.`quantity`)) AS `all_qnt` from ((`na_fleet2assault` `f` join `na_assault` `a` on((`a`.`assaultid` = `f`.`assaultid`))) join `na_construction` `c` on((`c`.`buildingid` = `f`.`unitid`))) group by cast(from_unixtime(`a`.`time`) as date),`c`.`mode` order by cast(from_unixtime(`a`.`time`) as date) desc */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_user_ext`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_user_ext` AS select from_unixtime(`u`.`regtime`) AS `r`,from_unixtime(`u`.`last`) AS `l`,`sn`.`network_id` AS `snid`,`u`.`userid` AS `userid`,`u`.`username` AS `username`,`u`.`email` AS `email`,`u`.`temp_email` AS `temp_email`,`u`.`languageid` AS `languageid`,`u`.`timezone` AS `timezone`,`u`.`templatepackage` AS `templatepackage`,`u`.`imagepackage` AS `imagepackage`,`u`.`theme` AS `theme`,`u`.`curplanet` AS `curplanet`,`u`.`points` AS `points`,`u`.`u_points` AS `u_points`,`u`.`r_points` AS `r_points`,`u`.`b_points` AS `b_points`,`u`.`u_count` AS `u_count`,`u`.`r_count` AS `r_count`,`u`.`b_count` AS `b_count`,`u`.`e_points` AS `e_points`,`u`.`be_points` AS `be_points`,`u`.`of_points` AS `of_points`,`u`.`of_level` AS `of_level`,`u`.`a_points` AS `a_points`,`u`.`a_count` AS `a_count`,`u`.`hp` AS `hp`,`u`.`battles` AS `battles`,`u`.`credit` AS `credit`,`u`.`exchange_rate` AS `exchange_rate`,`u`.`research_factor` AS `research_factor`,`u`.`ipcheck` AS `ipcheck`,`u`.`activation` AS `activation`,`u`.`password_activation` AS `password_activation`,`u`.`email_activation` AS `email_activation`,`u`.`regtime` AS `regtime`,`u`.`last` AS `last`,`u`.`asteroid` AS `asteroid`,`u`.`umode` AS `umode`,`u`.`umodemin` AS `umodemin`,`u`.`planetorder` AS `planetorder`,`u`.`delete` AS `delete`,`u`.`esps` AS `esps`,`u`.`show_all_constructions` AS `show_all_constructions`,`u`.`show_all_research` AS `show_all_research`,`u`.`show_all_shipyard` AS `show_all_shipyard`,`u`.`show_all_defense` AS `show_all_defense`,`u`.`user_bg_style` AS `user_bg_style`,`u`.`user_table_style` AS `user_table_style`,`u`.`skin_type` AS `skin_type`,`u`.`race` AS `race`,`u`.`user_agreement_read` AS `user_agreement_read`,`u`.`tutorial_state` AS `tutorial_state`,`u`.`tutorial_show` AS `tutorial_show` from (`na_user` `u` left join `na_social_network_user` `sn` on((`u`.`userid` = `sn`.`user_id`))) order by `u`.`userid` desc */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_user_imgpak_ext`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_user_imgpak_ext` AS select `na_user`.`imagepackage` AS `imagepackage`,count(0) AS `cnt` from `na_user` where (`na_user`.`activation` = _utf8'') group by `na_user`.`imagepackage` order by count(0) desc */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_user_online`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_user_online` AS select (select count(0) AS `count(*)` from `na_user` where ((from_unixtime(`na_user`.`last`) > subtime(now(),_utf8'00:05')) and (from_unixtime(`na_user`.`regtime`) < subtime(now(),_utf8'01:00')))) AS `online_5_h`,(select count(0) AS `count(*)` from `na_user` where ((from_unixtime(`na_user`.`last`) > subtime(now(),_utf8'00:10')) and (from_unixtime(`na_user`.`regtime`) < subtime(now(),_utf8'01:00')))) AS `online_10_h`,(select count(0) AS `count(*)` from `na_user` where ((from_unixtime(`na_user`.`last`) > subtime(now(),_utf8'00:15')) and (from_unixtime(`na_user`.`regtime`) < subtime(now(),_utf8'01:00')))) AS `online_15_h`,(select count(0) AS `count(*)` from `na_user` where ((from_unixtime(`na_user`.`last`) > subtime(now(),_utf8'00:30')) and (from_unixtime(`na_user`.`regtime`) < subtime(now(),_utf8'01:00')))) AS `online_30_h`,(select count(0) AS `count(*)` from `na_user` where ((from_unixtime(`na_user`.`last`) > (now() - interval 1 day)) and (from_unixtime(`na_user`.`regtime`) < subtime(now(),_utf8'01:00')))) AS `core_1_h`,(select count(0) AS `count(*)` from `na_user` where ((from_unixtime(`na_user`.`last`) > (now() - interval 2 day)) and (from_unixtime(`na_user`.`regtime`) < subtime(now(),_utf8'01:00')))) AS `core_2_h`,(select count(0) AS `count(*)` from `na_user` where ((from_unixtime(`na_user`.`last`) > (now() - interval 7 day)) and (from_unixtime(`na_user`.`regtime`) < subtime(now(),_utf8'01:00')))) AS `core_7_h`,(select count(0) AS `count(*)` from `na_user` where (from_unixtime(`na_user`.`last`) > subtime(now(),_utf8'00:05'))) AS `online_5`,(select count(0) AS `count(*)` from `na_user` where (from_unixtime(`na_user`.`last`) > subtime(now(),_utf8'00:10'))) AS `online_10`,(select count(0) AS `count(*)` from `na_user` where (from_unixtime(`na_user`.`last`) > subtime(now(),_utf8'00:15'))) AS `online_15`,(select count(0) AS `count(*)` from `na_user` where (from_unixtime(`na_user`.`last`) > subtime(now(),_utf8'00:30'))) AS `online_30`,(select count(0) AS `count(*)` from `na_user` where (from_unixtime(`na_user`.`last`) > (now() - interval 1 day))) AS `core_1`,(select count(0) AS `count(*)` from `na_user` where (from_unixtime(`na_user`.`last`) > (now() - interval 2 day))) AS `core_2`,(select count(0) AS `count(*)` from `na_user` where (from_unixtime(`na_user`.`last`) > (now() - interval 7 day))) AS `core_7`,(select count(0) AS `count(*)` from `na_user`) AS `all_users` */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_user_reg_stats`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_user_reg_stats` AS select cast(`na_user_ext`.`r` as date) AS `d`,`na_user_ext`.`snid` AS `snid`,count(0) AS `cnt` from `na_user_ext` group by cast(`na_user_ext`.`r` as date),`na_user_ext`.`snid` order by cast(`na_user_ext`.`r` as date) desc,`na_user_ext`.`snid` */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_user_reg_stats_month`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_user_reg_stats_month` AS select extract(year_month from `na_user_ext`.`r`) AS `d`,`na_user_ext`.`snid` AS `snid`,count(0) AS `cnt` from `na_user_ext` group by extract(year_month from `na_user_ext`.`r`),`na_user_ext`.`snid` order by extract(year_month from `na_user_ext`.`r`) desc,`na_user_ext`.`snid` */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_view_max_building_stats`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_view_max_building_stats` AS select `na_building2planet`.`buildingid` AS `buildingid`,`na_planet`.`userid` AS `userid`,max(`na_building2planet`.`level`) AS `max_level` from ((`na_building2planet` join `na_planet` on((`na_building2planet`.`planetid` = `na_planet`.`planetid`))) left join `na_ban_u` on((`na_ban_u`.`userid` = `na_planet`.`userid`))) where ((`na_planet`.`userid` is not null) and isnull(`na_ban_u`.`userid`)) group by `na_building2planet`.`buildingid`,`na_planet`.`userid` */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!50001 DROP VIEW IF EXISTS `na_view_sum_unit_stats`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8 */;
/*!50001 SET character_set_results     = utf8 */;
/*!50001 SET collation_connection      = utf8_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=CURRENT_USER SQL SECURITY DEFINER */
/*!50001 VIEW `na_view_sum_unit_stats` AS select `na_unit2shipyard`.`unitid` AS `unitid`,`na_planet`.`userid` AS `userid`,sum(`na_unit2shipyard`.`quantity`) AS `sum_quantity` from ((`na_unit2shipyard` join `na_planet` on((`na_unit2shipyard`.`planetid` = `na_planet`.`planetid`))) left join `na_ban_u` on((`na_ban_u`.`userid` = `na_planet`.`userid`))) where ((`na_planet`.`userid` is not null) and isnull(`na_ban_u`.`userid`)) group by `na_unit2shipyard`.`unitid`,`na_planet`.`userid` */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

