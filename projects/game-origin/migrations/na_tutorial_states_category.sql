-- 
-- Structure for table `na_tutorial_states_category`
-- 

DROP TABLE IF EXISTS `na_tutorial_states_category`;
CREATE TABLE IF NOT EXISTS `na_tutorial_states_category` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `title` varchar(255) CHARACTER SET utf8 DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB;
INSERT INTO `na_tutorial_states_category` (`id`, `title`) VALUES
  ('1', 'BASIC_TUTORIAL');