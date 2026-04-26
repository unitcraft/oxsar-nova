-- 
-- Structure for table `na_tutorial_states`
-- 

DROP TABLE IF EXISTS `na_tutorial_states`;
CREATE TABLE IF NOT EXISTS `na_tutorial_states` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(255) CHARACTER SET utf8 NOT NULL,
  `display_order` int(10) unsigned NOT NULL,
  `arrow_name` varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT 'left.png',
  `arrow_of` varchar(1023) CHARACTER SET utf8 DEFAULT NULL,
  `arrow_my` varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT 'left center',
  `arrow_at` varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT 'right center',
  `menu_div` varchar(255) CHARACTER SET utf8 NOT NULL,
  `formaction` varchar(255) CHARACTER SET utf8 NOT NULL,
  `dialog_vert` varchar(1023) CHARACTER SET utf8 DEFAULT 'top',
  `dialog_hor` varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT 'center',
  `category` int(10) unsigned NOT NULL DEFAULT '1',
  `modal` tinyint(1) unsigned NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB;

-- 
-- Data for table `na_tutorial_states`
-- 

INSERT INTO `na_tutorial_states` (`id`, `name`, `display_order`, `arrow_name`, `arrow_of`, `arrow_my`, `arrow_at`, `menu_div`, `formaction`, `dialog_vert`, `dialog_hor`, `category`, `modal`) VALUES
  ('1', 'change_name', '10', 'left.png', NULL, 'left center', 'right center', '', '', 'top', 'center', '1', '0'),
  ('2', 'achiv_bonus', '20', 'top.png', 'input[name=process]', 'center top', 'center botom', '', 'Constructions', 'bottom', 'center', '1', '0'),
  ('3', 'menu_achievements', '30', 'left.png', '#menu_achievements', 'left center', 'right center', '3', '', 'top', 'center', '1', '0'),
  ('4', 'menu_achievements_page', '40', 'left.png', '.for_pointer_req', 'left center', 'center center', '', 'Achievements', 'top', 'center', '1', '0'),
  ('5', 'menu_build', '50', 'left.png', '#menu_build', 'left center', 'right center', '', '', 'top', 'center', '1', '0'),
  ('6', 'build_solar', '60', 'left.png', '#build_construction_4', 'left center', 'right center', '', 'Constructions', 'top', 'center', '1', '0'),
  ('7', 'show_me_tutorial', '70', 'top.png', '#show_me_tutorial', 'center top', 'center bottom', '', '', 'bottom', 'center', '1', '0'),
  ('8', 'menu_galaxy', '80', 'left.png', '#menu_galaxy', 'left center', 'right center', '3', '', 'top', 'center', '1', '0'),
  ('9', 'menu_chat', '90', 'left.png', '#menu_chat', 'left center', 'right center', '2', '', 'top', 'center', '1', '0'),
  ('10', 'solarplant_bonus', '100', 'left.png', '.for_pointer_bonus', 'left center', 'center center', '', 'Constructions', 'bottom', 'center', '1', '0'),
  ('11', 'menu_fleet', '110', 'left.png', '#menu_fleet', 'left center', 'right center', '', '', 'top', 'center', '1', '0'),
  ('12', 'menu_simulator', '120', 'left.png', '#menu_simulator', 'left center', 'right center', '4', '', 'top', 'center', '1', '0'),
  ('13', 'menu_galaxy_desc', '81', 'left.png', NULL, 'left center', 'right center', '', 'Galaxy', 'bottom', 'center', '1', '0'),
  ('14', 'menu_chat_desc', '91', 'left.png', '#menu_build', 'left center', 'right center', '', 'Chat', 'top', 'center', '1', '0'),
  ('15', 'menu_fleet_desc', '111', 'left.png', NULL, 'left center', 'right center', '', 'Mission', 'bottom', 'center', '1', '0'),
  ('16', 'menu_simulator_desc', '121', 'left.png', NULL, 'left center', 'right center', '', 'Simulator', 'top', 'center', '1', '0'),
  ('17', 'continue_achiv', '130', 'left.png', NULL, 'left center', 'right center', '', '', 'top', 'center', '1', '0'),
  ('18', 'planet_creation_wait', '11', 'left.png', NULL, 'left center', 'right center', '', '', 'top', 'center', '1', '0');

