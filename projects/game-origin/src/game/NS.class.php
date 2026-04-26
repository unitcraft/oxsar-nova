<?php
/**
* Main programm. Handles all application-related classes.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

define("NS_VERSION", OXSAR_VERSION);
define("NS_REVISION", 0);

class NS
{
	/**
	* MemCacheHandler object.
	*
	* @var MemCacheHandler
	*/
	protected static $mch = null;

	/**
	* Planet object.
	*
	* @var Planet
	*/
	protected static $planet = null;

	/**
	* EventHandler object.
	*
	* @var EventHandler
	*/
	// protected static $eh = null;

	/**
	* Holds a list of all user's planets.
	*
	* @var array
	*/
	protected static $planetStack = array();

	/**
	* Holds a list of all available research levels.
	*
	* @var array
	*/
	protected static $research = array();
	protected static $research_added = array();
	protected static $research_special = array();

	/**
	* Holds all requirements.
	*
	* @var array
	*/
	protected static $requirements = array();

	/**
	* Fleet speed factor. It can be changed ataching fleet artefact to fleet.
	*
	* @var float
	*/
	public static $fleet_speed_factor = 1.0;

	/* public static function getFleetSpeedFactor()
	{
		return
	} */

	/**
	* Starts Net-Assault.
	*
	* @return void
	*/
	public function __construct()
	{
		if( !defined('YII_CONSOLE') && !NS::getUser()->get("userid") )
		{
			doHeaderRedirection("login.php", false);
		}
		
		self::$mch = new MemCacheHandler();

		if( !defined('YII_CONSOLE') )
		{
			if( $planetid = Core::getRequest()->getPOST("planetid") )
			{
				$userid = sqlSelectField("planet", "userid", "", "planetid = ".sqlVal($planetid)." AND userid=".sqlUser());
				if($userid)
				{
					NS::getUser()->set("curplanet", $planetid);
				}
			}
            if(ACHIEVEMENTS_ENABLED){
                if( $achievement_id = Core::getRequest()->getPOST("achievement_get_bonus_id") ){
                    Achievements::setAchievementState( null, null, max(0, (int)$achievement_id), ACHIEV_STATE_BONUS_GIVEN );
                }
                else if( $achievement_id = Core::getRequest()->getPOST("achievement_process_id") ){
                    Achievements::setAchievementState( null, null, max(0, (int)$achievement_id), ACHIEV_STATE_PROCESSED );
                }
            }
		}
		// Hook::event("NS_START");

		// $this->loadResearch();
		if( !defined('YII_CONSOLE') )
		{
			$this->setEventHandler();
			$this->setPlanet();
			$this->globalTPLAssigns();
			// Asteroid event (Once per week).
			if(!mt_rand(0, 10000) && NS::getUser()->get("asteroid") < time() - 604800)
			{
				$data = array();
				$data["metal"] = mt_rand(1, Core::getOptions()->get("MAX_ASTEROID_SIZE")) * 1000;
				$data["silicon"] = mt_rand(1, $data["metal"]/1000) * 1000;
				$data["galaxy"] = self::getPlanet()->getData("galaxy");
				$data["system"] = self::getPlanet()->getData("system");
				$data["position"] = self::getPlanet()->getData("position");
				$data["planet"] = self::getPlanet()->getData("planetname");
				// Hook::event("NS_ASTEROID_EVENT", array(&$data));
				// Updated By Pk
				Core::getQuery()->update("user", "asteroid", time(), "userid = ".sqlUser());
				// Updated By Pk
				sqlQuery("UPDATE ".PREFIX."galaxy SET "
					. " metal = metal + ".sqlVal($data["metal"])
					. ", silicon = silicon + ".sqlVal($data["silicon"])
					. " WHERE galaxy = ".sqlVal(self::getPlanet()->getData("galaxy"))
					. " AND system = ".sqlVal(self::getPlanet()->getData("system"))
					. " AND position = ".sqlVal(self::getPlanet()->getData("position")));
				new AutoMsg(MSG_ASTEROID, NS::getUser()->get("userid"), time(), $data);
				NS::getUser()->rebuild();
			}
			// Update last activity
			// Updated By Pk
			Core::getQuery()->update("user", array("last"), array(time()), "userid = ".sqlUser());
			$this->loadPage(Core::getRequest()->getGET("go"));
		}
	}

	/**
	* Sets EventHandler object.
	*
	* @return NS
	*/
	private function setEventHandler()
	{
		$classname = self::getSystemClassname("EventHandler");
		$eh = new $classname();
		if ( !defined('YII_CONSOLE') )
		{
			$eh->startEvents();
		}
	}

	/**
	* Returns EventHandler objeckt.
	*
	* @return EventHandler
	*/
	public static function getEH()
	{
		// return self::$eh;
		return EventHandler::getEH();
	}

	public static function getMCH()
	{
		return self::$mch;
	}

	/**
	* Sets planet object.
	*
	* @return NS
	*/
	protected function setPlanet()
	{
		self::$planet = null;
		self::$planet = new Planet(NS::getUser()->get("curplanet"), NS::getUser()->get("userid"));
		return $this;
	}

	public static function reloadPlanet()
	{
		self::$planet = null;
		self::$planet = new Planet(NS::getUser()->get("curplanet"), NS::getUser()->get("userid"));
	}

	/**
	* Returns Planet object.
	*
	* @return Planet
	*/
	public static function getPlanet()
	{
		return self::$planet;
	}

	/**
	* Returns User object.
	*
	* @return User
	*/
	public static function getUser()
	{
		return Core::getUser();
	}

	public static function getLanguageId()
	{
		$lang = Core::getLanguage();
		return $lang ? $lang->getOpt("languageid") : DEF_LANGUAGE_ID;
	}

	/**
	* Generic template assignments.
	*
	* @return NS
	*/
	protected function globalTPLAssigns()
	{
		Core::getTPL()->assign("charset", Core::getLanguage()->getOpt("charset"));

		// Set planets for right menu and fill planet stack.
		$planets= array();
		$i		= 0;
		$order	= getPlanetOrder(NS::getUser()->get("planetorder"));
		$joins	= "LEFT JOIN ".PREFIX."galaxy g ON (g.planetid = p.planetid)";
		$joins .= "LEFT JOIN ".PREFIX."planet m ON (g.moonid = m.planetid)";
		$atts	= array(
			"p.planetid", "p.ismoon", "p.planetname",
			"p.picture", "g.galaxy", "g.system",
			"g.position", "m.planetid AS moonid", "m.planetname AS moon",
			"m.picture AS mpicture"
		);
		$result = sqlSelect(
			"planet p",
			$atts,
			$joins,
			"p.userid = ".sqlUser()." AND p.ismoon = 0" /* AND g.destroyed = 0 */,
			$order
		);
		unset($order);
		while($row = sqlFetch($result))
		{
			$planets[$i]			= $row;
			$coords					= $row["galaxy"].":".$row["system"].":".$row["position"];
			$coords					= "[".$coords."]";
			$planets[$i]["coords"]	= $coords;
			$planets[$i]["picture"] = Image::getImage("planets/small/s_".$row["picture"].Core::getConfig()->get("PLANET_IMG_EXT"), $row["planetname"]." ".$coords, 60, 60);
			if($row["moonid"])
			{
				$planets[$i]["mpicture"] = Image::getImage("planets/small/s_".$row["mpicture"].Core::getConfig()->get("PLANET_IMG_EXT"), $row["moon"]." ".$coords, 20, 20);
			}
			array_push(self::$planetStack, $row["planetid"]);
			$i++;
		}
		sqlEnd($result);
		Core::getTPL()->addLoop("planetHeaderList", $planets);

		// Menu
		if( isFacebookSkin() || isMobileSkin() )
		{
			$menu_class = isMobileSkin() ? "ExtMenu" : "Menu";
			$menu = new $menu_class("Menu");
			Core::getTPL()->addLoop("menu_headers", $menu->getMenuTitles($planets));
			Core::getTPL()->addLoop("navigation", $menu);
		}
		else
		{
			Core::getTPL()->addLoop("navigation", new Menu("Menu"));
		}

		// Assignments
		Core::getTPL()->assign("themePath", (NS::getUser()->get("theme")) ? NS::getUser()->get("theme") : RELATIVE_URL);
		// if( NS::getPlanet()->getPlanetId() )
		Core::getTPL()->assign("planetImageSmall", Image::getImage("planets/small/s_".self::getPlanet()->getData("picture").Core::getConfig()->get("PLANET_IMG_EXT"), self::getPlanet()->getData("planetname")));
		Core::getTPL()->assign("currentPlanet", NS::getPlanet()->getPlanetId() ? Link::get("game.php/PlanetOptions", self::getPlanet()->getData("planetname")) : self::getPlanet()->getData("planetname"));
		Core::getTPL()->assign("currentCoords", self::getPlanet()->getCoords());
		Core::getTPL()->assign("planetMetal", fNumber(self::getPlanet()->getData("metal")));
		Core::getTPL()->assign("planetSilicon", fNumber(self::getPlanet()->getData("silicon")));
		Core::getTPL()->assign("planetHydrogen", fNumber(self::getPlanet()->getData("hydrogen")));
		Core::getTPL()->assign("planetName", self::getPlanet()->getData("planetname"));
		Core::getTPL()->assign("planetAEnergy", fNumber(self::getPlanet()->getConsumption("energy")));
		Core::getTPL()->assign("planetEnergy", fNumber(self::getPlanet()->getProd("energy")));
		// $formaction = !empty($_SERVER['DOCUMENT_URI']) ? $_SERVER['DOCUMENT_URI'] : $_SERVER['PHP_SELF'];
		// $formaction = !empty($_SERVER['REQUEST_URI']) ? $_SERVER['REQUEST_URI'] : $_SERVER['PHP_SELF'];
		$formaction = socialUrl($_SERVER['PHP_SELF']);
		/*
		if( defined('SN') )
		{
			$formaction .= '?' . '';
		}
		*/
		Core::getTPL()->assign("formaction", $formaction);

		// Show message if user is in vacation or deletion mode.
		$delete = false;
		if(NS::getUser()->get("delete") > 0)
		{
			$delete = Core::getLanguage()->getItem("ACCOUNT_WILL_BE_DELETED");
		}
		$umode = false;
		if(NS::getUser()->get("umode"))
		{
			$umode = Core::getLanguage()->getItem("UMODE_ENABLED");
		}
		Core::getTPL()->assign("delete", $delete);
		Core::getTPL()->assign("umode", $umode);

		// Format ressources.
		if(self::getPlanet()->getData("metal") >= self::getPlanet()->getStorage("metal"))
		{
			Core::getTPL()->assign("metalClass", "false");
		}
		else { Core::getTPL()->assign("metalClass", ""); }
		if(self::getPlanet()->getData("silicon") >= self::getPlanet()->getStorage("silicon"))
		{
			Core::getTPL()->assign("siliconClass", "false");
		}
		else { Core::getTPL()->assign("siliconClass", ""); }
		if(self::getPlanet()->getData("hydrogen") >= self::getPlanet()->getStorage("hydrogen"))
		{
			Core::getTPL()->assign("hydrogenClass", "false");
		}
		else { Core::getTPL()->assign("hydrogenClass", ""); }
		if(self::getPlanet()->getConsumption("energy") >= self::getPlanet()->getProd("energy"))
		{
			Core::getTPL()->assign("energyClass", "false");
		}
		else { Core::getTPL()->assign("energyClass", ""); }
		Core::getTPL()->assign("remainingEnergy", fNumber(self::getPlanet()->getEnergy()));
		return $this;
	}

	/**
	* Loads the requested page.
	*
	* @param string	Page name
	*
	* @return NS
	*/
	protected function loadPage($page)
	{
		// Special pages
		switch($page)
		{
			case "StockLotRecall": $page = "Stock"; break;
			case "StockLotPremium": $page = "Stock"; break;
			case "StockBan": $page = "Stock"; break;
			case "PackConstruction": $page = "BuildingInfo"; break;
			case "PackResearch": $page = "BuildingInfo"; break;
			case "UpgradeConstruction": $page = "Constructions"; break;
			case "UpgradeConstructionImm": $page = "Constructions"; break;
			case "AbortConstruction": $page = "Constructions"; break;
			case "DemolishConstruction": $page = "Constructions"; break;
			case "ConstructionInfo": $page = "Constructions"; break;
			case "ResearchInfo": $page = "Research"; break;
			case "PlanetOptions": $page = "Main"; break;
			case "HomePlanetRequired": $page = "Main"; break;
			case "Prefs": $page = "Preferences"; break;
			case "UpgradeResearch": $page = "Research"; break;
			case "UpgradeResearchImm": $page = "Research"; break;
			case "AbortResearch": $page = "Research"; break;
			case "Defense": $page = "Shipyard"; break;
			case "AbortShipyard": $page = "Shipyard"; break;
			case "AbortDefense": $page = "Shipyard"; break;
			case "StartShipyardVIP": $page = "Shipyard"; break;
			case "StartDefenseVIP": $page = "Shipyard"; break;
			case "AbortRepair": $page = "Repair"; break;
			case "StartRepairVIP": $page = "Repair"; break;
			case "Disassemble": $page = "Repair"; break;
			case "AbortDisassemble": $page = "Repair"; break;
			case "StartDisassembleVIP": $page = "Repair"; break;
			case "MemberList": $page = "Alliance"; break;
			case "AlliancePage": $page = "Alliance"; break;
			case "BuyArtefact": $page = "ArtefactMarket"; break;
			case "ArtefactOn": $page = "Artefacts"; break;
			case "ArtefactOff": $page = "Artefacts"; break;
			case "ArtefactInfo": $page = "Artefacts"; break;
			case "AchievementInfo": $page = "Achievements"; break;
			case "AchievementHideAjax": $page = "Achievements"; break;
			case "AchievementGetBonus": $page = "Achievements"; break;
			case "AchievementProcess": $page = "Achievements"; break;
			case "AchievementsRecalc": $page = "Achievements"; break;
			case "AchievementsProfile": $page = "Achievements"; break;
			case "AchievementsDone": $page = "Achievements"; break;
			case "AchievementsAvaliable": $page = "Achievements"; break;
			case "SaveNotes": $page = "Notepad"; break;
			case "PaymentVkontakte": $page = "Payment"; break;
			case "PaymentWebmoney": $page = "Payment"; break;
			case "PaymentRobokassa": $page = "Payment"; break;
			case "PaymentSMS": $page = "Payment"; break;
			case "ControlFleet": $page = "Mission"; break;
		}

		self::getPage($page);
		return $this;
	}

	private static function fixAlienTechQuery($id, $uid)
	{
		static $alienTechFixed = false;
		if($id == UNIT_ALIEN_TECH
			// && isAdmin() // debug
			&& $alienTechFixed === false
			&& empty(self::$research[$uid][$id])
			&& NS::getUser()
			&& NS::getPlanet()
			&& $uid == NS::getUser()->get("userid")
			)
		{
			$alienTechFixed = true;
			// self::$research[$uid][$id] = 0;
			// self::$research_added[$uid][$id] = 0;
			if(AlienAI::isAlienPosition( NS::getPlanet()->getData("galaxy"), NS::getPlanet()->getData("system") ))
			{
				self::$research[$uid][$id] = self::$research_added[$uid][$id] = 1;
			}
		}
	}

	/**
	* Returns the research level.
	*
	* @param integer	Id of requested research
	*
	* @return integer	Level of requested research
	*/
	public static function getResearch($id, $uid = null)
	{
		if(is_null($uid))
		{
            if(!NS::getUser()){
                // error!
                return 0;
            }
			$uid = NS::getUser()->get("userid");
		}
		if(!isset(self::$research[$uid]))
		{
			self::loadResearch($uid);
		}
		self::fixAlienTechQuery($id, $uid);
		if(isset(self::$research[$uid][$id]))
		{
			return self::$research[$uid][$id];
		}
		return 0;
	}

	/**
	* Returns the temporary added research level.
	*
	* @param integer	Id of requested research
	*
	* @return integer	Level of requested research
	*/
	public static function getAddedResearch($id, $uid = null)
	{
		if(is_null($uid))
		{
			$uid = NS::getUser()->get("userid");
		}
		if(!isset(self::$research_added[$uid]))
		{
			self::loadResearch($uid);
		}
		self::fixAlienTechQuery($id, $uid);
		if(isset(self::$research_added[$uid][$id]))
		{
			return self::$research_added[$uid][$id];
		}
		return 0;
	}

	/**
	* Returns the temporary added research level.
	*
	* @param integer	Id of requested research
	*
	* @return integer	Level of requested research
	*/
	public static function getResearchSpecial($id, $uid = null)
	{
		if(is_null($uid))
		{
			$uid = NS::getUser()->get("userid");
		}
		if(!isset(self::$research_added[$uid]))
		{
			self::loadResearch($uid);
		}
		if(isset(self::$research_special[$uid][$id]))
		{
			return self::$research_special[$uid][$id];
		}
		return "";
	}

	/**
	* Returns the level of a building.
	*
	* @param integer	Building id
	*
	* @return integer	Building level
	*/
	public static function getBuilding($id, $planetid = null, $allow_ext = false)
	{
		if(is_null($planetid))
		{
			if(!NS::getUser())
			{
				return 0;
			}
			$planetid = NS::getUser()->get("curplanet");
		}
		static $cache = array();
		// $allow_ext = (int)$allow_ext;
		if(isset($cache[$planetid][$id]))
		{
			return $cache[$planetid][$id];
		}
		if(NS::getPlanet() && $planetid == NS::getPlanet()->getPlanetId())
		{
			return $cache[$planetid][$id] = NS::getPlanet()->getBuilding($id, $allow_ext);
		}
		$level = sqlQueryField("SELECT level FROM ".PREFIX."building2planet WHERE buildingid=".sqlVal($id)." AND planetid=".sqlVal($planetid));
		if($level > 0)
		{
			return $cache[$planetid][$id] = $level;
		}
		if($allow_ext && (!is_array($allow_ext) || in_array($id, $allow_ext)))
		{
			$level = sqlSelectField("galaxy g", "b2p.level",
				"JOIN ".PREFIX."building2planet b2p ON b2p.planetid = g.planetid",
				"g.moonid = ".sqlVal($planetid)." AND b2p.buildingid = ".sqlVal($id)
				);
			if($level > 0)
			{
				return $cache[$planetid][$id] = $level;
			}
			$level = sqlSelectField("galaxy g", "b2p.level",
				"JOIN ".PREFIX."building2planet b2p ON b2p.planetid = g.moonid",
				"g.planetid = ".sqlVal($planetid)." AND b2p.buildingid = ".sqlVal($id)
				);
			if($level > 0)
			{
				return $cache[$planetid][$id] = $level;
			}
		}
		return $cache[$planetid][$id] = 0;
	}

	/**
	* Find out whether contruction is available.
	*
	* @param integer	Construction id
	*
	* @return boolean	True, if construction can be upgraded, false if not
	*/
	public static function checkRequirements($bid, $userid = null, $planetid = null, $allow_ext = false)
	{
		self::loadRequirements();
		if(!isset(self::$requirements[$bid]))
		{
			return true;
		}
		foreach(self::$requirements[$bid] as $r)
		{
			if($r["mode"] == UNIT_TYPE_CONSTRUCTION || $r["mode"] == UNIT_TYPE_MOON_CONSTRUCTION)
			{
				$rLevel = self::getBuilding($r["needs"], $planetid, $allow_ext);
			}
			else if($r["mode"] == UNIT_TYPE_RESEARCH)
			{
				$rLevel = self::getResearch($r["needs"], $userid);
			}
			else if($r["mode"] == UNIT_TYPE_FLEET || $r["mode"] == UNIT_TYPE_DEFENSE)
			{
				$rLevel = getShipyardQuantity($r["needs"], false, $planetid);
			}
			else if($r["mode"] == UNIT_TYPE_ACHIEVEMENT && ACHIEVEMENTS_ENABLED)
			{
				// $rLevel = ( ( Achievements::isGrantedAchievement($userid, $r['needs']) ) ? 1000 : 0 );
				$rLevel = Achievements::getProcessedAchievementQuantity($userid, $r['needs']);
			}
			else
			{
				continue;
			}

			if($rLevel < $r["level"])
			{
				return false;
			}
			if($r["level_limit"] > 0 && $rLevel > $r["level_limit"])
			{
				return false;
			}
		}
		return true;
	}

	public static function canUpgradeConstruction($id)
	{
		switch($id)
		{
		case UNIT_NANO_FACTORY:
			return !NS::getEH()->getShipyardEvents() && !NS::getEH()->getRepairEvents();

		case UNIT_SHIPYARD:
			return !NS::getEH()->getShipyardModeEvents(EVENT_BUILD_FLEET);

		case UNIT_DEFENSE_FACTORY:
			return !NS::getEH()->getShipyardModeEvents(EVENT_BUILD_DEFENSE);

		case UNIT_REPAIR_FACTORY:
			return !NS::getEH()->getRepairEvents();

		case UNIT_RESEARCH_LAB:
			return !NS::getEH()->getFirstResearchEvent();
		}
		return true;
	}

	public static function requiremtentsList($bid, $userid = null, $planetid = null, $show_all = false, $safe_mode = false)
	{
		Core::getLanguage()->load(array('Achievements', 'info'));
		self::loadRequirements();
		if(!isset(self::$requirements[$bid]))
		{
			return "";
		}
		$requirements = array();
		foreach(self::$requirements[$bid] as $r)
		{
			if($r["mode"] == UNIT_TYPE_CONSTRUCTION || $r["mode"] == UNIT_TYPE_MOON_CONSTRUCTION)
			{
				$rLevel = /* empty($planetid) ? 0 : */ self::getBuilding($r["needs"], $planetid);
                $page = 'ConstructionInfo';
			}
			else if($r["mode"] == UNIT_TYPE_RESEARCH)
			{
				$rLevel = /* empty($planetid) ? 0 : */ self::getResearch($r["needs"], $userid);
                $page = 'ResearchInfo';
			}
			else if($r["mode"] == UNIT_TYPE_FLEET || $r["mode"] == UNIT_TYPE_DEFENSE)
			{
				$rLevel = /* empty($planetid) ? 0 : */ getShipyardQuantity($r["needs"], false, $planetid);
			}
			else if($r["mode"] == UNIT_TYPE_ACHIEVEMENT && ACHIEVEMENTS_ENABLED)
			{
				// $rLevel = ( ( Achievements::isGrantedAchievement($userid, $r['needs']) ) ? 1000 : 0 );
				$rLevel = /* empty($planetid) ? 0 : */ Achievements::getProcessedAchievementQuantity($userid, $r['needs']);
			}
			else
			{
				continue;
			}

			if( $safe_mode )
			{
				if( $r["mode"] == UNIT_TYPE_ACHIEVEMENT )
				{
					$style_class = "";
					$requirements[] = "<span class='$style_class'>"
						. Link::get('game.php/AchievementInfo/' . $r['needs'], Core::getLanguage()->getItem($r['name']), '', "$style_class")
						. ($r["level"] > 1 ? " <nobr>(" . Core::getLanguage()->getItem("QUANTITY") . " " . fNumber($r["level"]) . ")</nobr>" : "")
						. "</span>";
				}
				else if( $r["mode"] == UNIT_TYPE_FLEET || $r["mode"] == UNIT_TYPE_DEFENSE)
				{
					$style_class = "";
					$requirements[] = "<span class='$style_class'>"
						. Link::get("game.php/UnitInfo/".$r["needs"], Core::getLanguage()->getItem($r["name"]), "", "$style_class")
						. " <nobr>(" . Core::getLanguage()->getItem("QUANTITY") . " " . fNumber($r["level"])
						. ")</nobr></span>";
				}
				else
				{
					$style_class = "";
					$requirements[] = "<span class='$style_class'>"
						. Link::get("game.php/{$page}/".$r["needs"], Core::getLanguage()->getItem($r["name"]), "", "$style_class")
						. " <nobr>(" . Core::getLanguage()->getItem("LEVEL") . " " . fNumber($r["level"])
						. ")</nobr></span>";
				}
			}
			else
			{
				if( $r["mode"] == UNIT_TYPE_ACHIEVEMENT )
				{
					if($rLevel < $r["level"])
					{
						$style_class = "false2";
						$requirements[] = "<span class='$style_class'>"
							. Link::get("game.php/AchievementInfo/".$r["needs"], Core::getLanguage()->getItem($r["name"]), "", "$style_class")
							. ($r["level"] > 1 ? " <nobr>(" . Core::getLanguage()->getItem("QUANTITY") . " " . fNumber($r["level"])
							. ($rLevel < $r["level"] ? ", +" . fNumber($r["level"] - $rLevel) : "")
							. ")</nobr></span>" : "</span>");
					}
					else if($r["level_limit"] > 0 && $rLevel > $r["level_limit"])
					{
						$style_class = "false2";
						$requirements[] = "<span class='$style_class'>"
							. Link::get("game.php/AchievementInfo/".$r["needs"], Core::getLanguage()->getItem($r["name"]), "", "$style_class")
							. ($r["level"] > 1 ? " <nobr>(" . Core::getLanguage()->getItem("QUANTITY") . " " . fNumber($r["level"])
							. ($rLevel > $r["level_limit"] ? ", " . fNumber($r["level_limit"] - $rLevel) : "")
							. ")</nobr></span>" : "</span>");
					}
					else if($show_all)
					{
						$style_class = "true";
						$requirements[] = "<span class='$style_class'>"
							. Link::get("game.php/AchievementInfo/".$r["needs"], Core::getLanguage()->getItem($r["name"]), "", "$style_class")
							. ($r["level"] > 1 ? " <nobr>(" . Core::getLanguage()->getItem("QUANTITY") . " " . fNumber($r["level"])
							. ")</nobr></span>" : "</span>");
					}
				}
				else if( $r["mode"] == UNIT_TYPE_FLEET || $r["mode"] == UNIT_TYPE_DEFENSE )
				{
					if($rLevel < $r["level"])
					{
						$style_class = "false2";
						$requirements[] = "<span class='$style_class'>"
							. Link::get("game.php/UnitInfo/".$r["needs"], Core::getLanguage()->getItem($r["name"]), "", "$style_class")
							. " <nobr>(" . Core::getLanguage()->getItem("QUANTITY") . " " . fNumber($r["level"])
							. ($rLevel < $r["level"] ? ", +" . fNumber($r["level"] - $rLevel) : "")
							. ")</nobr></span>";
					}
					else if($r["level_limit"] > 0 && $rLevel > $r["level_limit"])
					{
						$style_class = "false2";
						$requirements[] = "<span class='$style_class'>"
							. Link::get("game.php/UnitInfo/".$r["needs"], Core::getLanguage()->getItem($r["name"]), "", "$style_class")
							. " <nobr>(" . Core::getLanguage()->getItem("QUANTITY") . " " . fNumber($r["level"])
							. ($rLevel > $r["level_limit"] ? ", " . fNumber($r["level_limit"] - $rLevel) : "")
							. ")</nobr></span>";
					}
					else if($show_all)
					{
						$style_class = "true";
						$requirements[] = "<span class='$style_class'>"
							. Link::get("game.php/UnitInfo/".$r["needs"], Core::getLanguage()->getItem($r["name"]), "", "$style_class")
							. " <nobr>(" . Core::getLanguage()->getItem("QUANTITY") . " " . fNumber($r["level"])
							. ")</nobr></span>";
					}
				}
				else
				{
					if($rLevel < $r["level"])
					{
						$style_class = "false2";
						$requirements[] = "<span class='$style_class'>"
							. Link::get("game.php/{$page}/".$r["needs"], Core::getLanguage()->getItem($r["name"]), "", "$style_class")
							. " <nobr>(" . Core::getLanguage()->getItem("LEVEL") . " " . fNumber($r["level"])
							. ($rLevel < $r["level"] ? ", +" . fNumber($r["level"] - $rLevel) : "")
							. ")</nobr></span>";
					}
					else if($r["level_limit"] > 0 && $rLevel > $r["level_limit"])
					{
						$style_class = "false2";
						$requirements[] = "<span class='$style_class'>"
							. Link::get("game.php/{$page}/".$r["needs"], Core::getLanguage()->getItem($r["name"]), "", "$style_class")
							. " <nobr>(" . Core::getLanguage()->getItem("LEVEL") . " " . fNumber($r["level"])
							. ($rLevel > $r["level_limit"] ? ", " . fNumber($r["level_limit"] - $rLevel) : "")
							. ")</nobr></span>";
					}
					else if($show_all)
					{
						$style_class = "true";
						$requirements[] = "<span class='$style_class'>"
							. Link::get("game.php/{$page}/".$r["needs"], Core::getLanguage()->getItem($r["name"]), "", "$style_class")
							. " <nobr>(" . Core::getLanguage()->getItem("LEVEL") . " " . fNumber($r["level"])
							. ")</nobr></span>";
					}
				}
			}
		}
		return implode("<br />", $requirements);
	}
	/**
	* Loads all requirements from the cache.
	*
	* @return NS
	*/
	protected static function loadRequirements()
	{
		if( empty(self::$requirements) )
		{
//			if(!Core::getCache()->objectExists("requirements"))
//			{
			$c_name = 'NS.loadRequirements';
			$requirements = false;
			if( $requirements === false )
			{
				$requirements_records = Requirements_YII::model()
					->with('need_const')
					->findAll(array('order' => 'need_const.display_order ASC, t.buildingid ASC'));
				if( $requirements_records )
				{
					foreach($requirements_records as $requirement_records)
					{
						$data = $requirement_records->getAttributes();
						if( $requirement_records->need_const )
						{
							$data = array_merge($data, $requirement_records->need_const->getAttributes());
						}
						$requirements[intval($requirement_records->buildingid)][] = $data;
					}
				}
				// cache disabled
			}
			self::$requirements = $requirements;
//				$result = sqlSelect(
//					"requirements r",
//					array("r.buildingid", "r.needs", "b.name", "b.mode", "r.level", "r.level_limit"),
//					"LEFT JOIN ".PREFIX."construction b ON b.buildingid = r.needs",
//					"",
//					"b.display_order ASC, r.buildingid ASC"
//				);
//				Core::getCache()->buildObject("requirements", $result, "buildingid");
//			}
//			self::$requirements = Core::getCache()->readObject("requirements");
		}
	}

	/**
	* Loads all research items.
	*
	* @return NS
	*/
	protected static function loadResearch($uid = null)
	{
		if(is_null($uid))
		{
			$uid = NS::getUser()->get("userid");
		}

		$result = sqlSelect("research2user r", array("r.buildingid", "r.level", "r.added", "c.special"),
			"LEFT JOIN ".PREFIX."construction c ON c.buildingid = r.buildingid", "r.userid = ".sqlVal($uid));
		while($row = sqlFetch($result))
		{
			self::$research[$uid][$row["buildingid"]] = $row["level"];
			self::$research_added[$uid][$row["buildingid"]] = $row["added"];
			self::$research_special[$uid][$row["buildingid"]] = $row["special"];
		}
		sqlEnd($result);
	}

	/**
	* Calculates full speed of a spaceship due to the engine level.
	*
	* @param integer	Ship id
	* @param integer	basic speed
	*
	* @return integer	Full speed
	*/
	public static function getSpeed($id, $speed, $uid = null)
	{
		if(is_null($uid))
		{
			$uid = NS::getUser()->get("userid");
		}

		$select = array(
			"s2e.engineid",
			"s2e.level",
			"s2e.base_speed",
			"e.factor"
			);
		$joins	= "LEFT JOIN ".PREFIX."engine e ON e.engineid = s2e.engineid ";
		$joins .= "LEFT JOIN ".PREFIX."research2user r2u ON r2u.buildingid = s2e.engineid ";
		$where	= "s2e.unitid = ".sqlVal($id)." AND r2u.level >= s2e.level AND r2u.userid = ".sqlVal($uid);
		$order	= "s2e.level DESC";
		$row = sqlSelectRow("ship2engine s2e", $select, $joins, $where, $order, "1");
		if($row)
		{
			// Hook::event("NS_GET_FLY_SPEED_FIRST", array($id, &$speed, &$row));
			if($row["base_speed"] > 0)
			{
				$speed = intval($row["base_speed"]);
			}
			$level = self::getResearch($row["engineid"], $uid);
			$multiplier = $level * intval($row["factor"]);
			if($multiplier > 0)
			{
				$speed += $speed / 100 * $multiplier;
			}
		}
		$speed = round($speed * NS::getGalaxyParam(null, "FLEET_SPEED_FACTOR", 1));
		// Hook::event("NS_GET_FLY_SPEED_LAST", array($id, &$speed));
		return $speed;
	}

	public static function getUnitEngines($id, $unit_base_speed)
	{
		$r = array();
		$select = array(
			"s2e.engineid",
			"s2e.level",
			"s2e.base_speed",
			"e.factor",
			"r2u.level as tech_level",
			"c.name",
			);
		$joins	= "LEFT JOIN ".PREFIX."engine e ON e.engineid = s2e.engineid ";
		$joins .= "LEFT JOIN ".PREFIX."research2user r2u ON r2u.buildingid = s2e.engineid ";
		$joins .= "LEFT JOIN ".PREFIX."construction c ON e.engineid = c.buildingid ";
		$where	= "s2e.unitid = ".sqlVal($id)." AND r2u.userid = ".sqlUser();
		$order	= "s2e.level DESC";
		$active = 0;
		$result = sqlSelect("ship2engine s2e", $select, $joins, $where, $order);
		while($row = sqlFetch($result))
		{
			// Hook::event("NS_GET_FLY_SPEED_FIRST", array($id, &$speed, &$row));
			$base_speed = $unit_base_speed;
			if($row["base_speed"] > 0)
			{
				$base_speed = intval($row["base_speed"]);
			}
			$speed = $base_speed;
			if($row["tech_level"] >= $row["level"])
			{
				$active++;
			}
			$level = max($row["tech_level"], $row["level"]); // self::getResearch($row["engineid"]);
			$multiplier = $level * intval($row["factor"]);
			if($multiplier > 0)
			{
				$speed = $base_speed / 100 * $multiplier + $base_speed;
			}

			$r[] = array(
				"engineid" => $row["engineid"],
				"engine_name" => $row["name"],
				"engine_level" => $row["level"],
				"tech_level" => $row["tech_level"],
				"base_speed" => $base_speed,
				"speed" => $speed,
				"active" => $active == 1,
				);
		}
		sqlEnd($result);

		if(!$r || !$active)
		{
			$r[] = array(
				"engineid" => false,
				"engine_name" => "",
				"engine_level" => 0,
				"tech_level" => 0,
				"base_speed" => $unit_base_speed,
				"speed" => $unit_base_speed,
				"active" => $active == 0,
				);
		}
		// debug_var($r, "[getUnitEngines]");
		// Hook::event("NS_GET_FLY_SPEED_LAST", array($id, &$speed));
		return array_reverse($r);
	}

	public static function getGalaxyParam($galaxy, $param_name, $def_value)
	{
		global $GALAXY_SPEC_PARAMS;
		if(empty($galaxy) && NS::getPlanet())
		{
			$galaxy = NS::getPlanet()->getData("galaxy");
		}
		return isset($GALAXY_SPEC_PARAMS[$galaxy][$param_name]) ? $GALAXY_SPEC_PARAMS[$galaxy][$param_name] : $def_value;
	}

	/**
	* Calculates the flying time.
	*
	* @param integer	Distance to destination
	* @param integer	Maximum speed of the slowest ship
	* @param integer	Speed factor (Can throttel the max. speed)
	*
	* @return integer	Flying time in seconds
	*/
	public static function getFlyTime($distance, $maxspeed, $speed = 100, $premium = false)
	{
		$speed /= 10;

		if($speed < 0.1) $speed = 0.1;
		else if($speed > 10) $speed = 10;

		$maxspeed = max(0.01, $maxspeed);
		$time = round((35000 / $speed) * sqrt($distance * 10 / $maxspeed) + 10);
		if(GAMESPEED > 0)
		{
			$time *= floatval(GAMESPEED);
		}
		if($premium)
		{
			$time *= EXCH_PREMIUM_LOT_FLY_TIME_MULT;
		}
		return ceil($time);
	}

	/**
	* Calculates the flying consumption.
	*
	* @param integer	Basic consumption
	* @param integer	Distance to destination
	* @param integer	Speed factor (Can decrease the consumption)
	*
	* @return integer	Full consumption
	*/
	public static function getFlyConsumption($basicConsumption, $dist, $speed = 100)
	{
		$speed /= 10;

		if($speed < 0.1) $speed = 0.1;
		else if($speed > 10) $speed = 10;

		return ceil($basicConsumption * $dist / 35000 * (($speed / 10) + 1) * (($speed / 10) + 1));
	}

	/**
	* Calculates the distance.
	*
	* @param integer	Destination galaxy
	* @param integer	Destination system
	* @param integer	Destination position
	*
	* @return integer	Distance
	*/
	public static function getDistance($galaxy, $system, $pos, $oGalaxy = null, $oSystem = null, $oPos = null)
	{
		if( is_null($oGalaxy) )
		{
			$oGalaxy = self::getPlanet()->getData("galaxy");
			$oSystem = self::getPlanet()->getData("system");
			$oPos = self::getPlanet()->getData("position");
		}

		if($galaxy - $oGalaxy != 0)
		{
			return abs($galaxy - $oGalaxy) * (defined('GALAXY_DISTANCE_MULT') ? GALAXY_DISTANCE_MULT : 1);
		}
		else if( $system - $oSystem != 0 )
		{
			return self::getSystemsDiff($system, $oSystem) * 5 * 19 + 2700;
		}
		else if( $pos - $oPos != 0 )
		{
			if( $pos == EXPED_PLANET_POSITION && defined('POSITION_TO_CALC_EXP_TO') )
			{
				$pos = POSITION_TO_CALC_EXP_TO;
			}
			return abs($pos - $oPos) * 5 + 1000;
		}
		return 5;
	}

	/**
	* Calculates the distance between systems.
	*
	* @param integer	Destination system
	*
	* @return integer	Distance
	*/
	public static function getSystemsDiff($system, $oSystem = null)
	{
		if( is_null($oSystem) )
		{
			$oSystem = self::getPlanet()->getData("system");
		}

		$between = abs($system - $oSystem);
		if( $between > 0 )
		{
			$systems = NUM_SYSTEMS;
			if($systems > 0)
			{
				$around = $systems - max( $system, $oSystem ) + min( $system, $oSystem );
				return min( $between, $around );
			}
			return $between;
		}
		return 0;
	}

	/**
	* Возвращает +-N.
	* Если N >= 0, расстояние указано в системах.
	* Если N < 0, расстояние в галактиках.
	*
	* @param int $galaxy
	* @param int $system
	* @param int $pos
	*/
	public static function getDistanceSystems($galaxy, $system, $pos)
	{
		$oSystem = self::getPlanet()->getData("system");
		$oGalaxy = self::getPlanet()->getData("galaxy");
		if ($galaxy != $oGalaxy)
			return -abs($galaxy - $oGalaxy);

		return self::getSystemsDiff($system);// abs($system - $oSystem);
	}

	public static function getMonitorActivityRange()
	{
		$star_surv = NS::getPlanet()->getBuilding(UNIT_STAR_SURVEILLANCE);
		$hyper_tech = NS::getResearch(UNIT_HYPERSPACE_TECH);
		return round((pow($star_surv, 2) - 1) * (1 + $hyper_tech / 10.0));
	}

	/**
	* Returns the mission name due to an id.
	*
	* @param integer	Mission id
	*
	* @return string	Mission name
	*/
	public static function getMissionName($id, $is_owner = false)
	{
		switch($id)
		{
		case EVENT_POSITION: $return = "STATIONATE"; break;
		case EVENT_DELIVERY_UNITS: $return = $is_owner ? "DELIVERY_UNITS_SELLER" : "DELIVERY_UNITS"; break;
		case EVENT_TRANSPORT: $return = "TRANSPORT"; break;
		case EVENT_DELIVERY_ARTEFACTS: $return = "DELIVERY_ARTEFACTS"; break;
		case EVENT_STARGATE_TRANSPORT: $return = "STARGATE_TRANSPORT"; break;
		case EVENT_STARGATE_JUMP: $return = "STARGATE_JUMP"; break;
		case EVENT_DELIVERY_RESOURSES: $return = "DELIVERY_RESOURSES"; break;
		case EVENT_COLONIZE: $return = "COLONIZE"; break;
		case EVENT_COLONIZE_RANDOM_PLANET:
		case EVENT_COLONIZE_NEW_USER_PLANET:
			$return = "COLONIZE_RANDOM_PLANET";
			break;
		case EVENT_RECYCLING: $return = "RECYCLING"; break;
		case EVENT_ALIEN_ATTACK:
		// case EVENT_ALIEN_ATTACK_CUSTOM: - unknown
		case EVENT_ATTACK_SINGLE: $return = "ATTACK"; break;
		case EVENT_ATTACK_DESTROY_BUILDING: $return = "DESTROY_ATTACK"; break;
		case EVENT_ATTACK_ALLIANCE_DESTROY_BUILDING: $return = "ALLIANCE_DESTROY_ATTACK"; break;
		case EVENT_ATTACK_DESTROY_MOON: $return = "MOON_DESTROY_ATTACK"; break;
		case EVENT_ATTACK_ALLIANCE_DESTROY_MOON: $return = "ALLIANCE_MOON_DESTROY_ATTACK"; break;
		case EVENT_SPY: $return = "SPY"; break;
		case EVENT_ATTACK_ALLIANCE:
		case EVENT_ALLIANCE_ATTACK_ADDITIONAL: $return = "ALLIANCE_ATTACK"; break;
		case EVENT_ALIEN_HALT:
		case EVENT_ALIEN_HOLDING:
		case EVENT_HALT:
		case EVENT_HOLDING: $return = "HALT"; break;
		case EVENT_MOON_DESTRUCTION: $return = "MOON_DESTRUCTION"; break;
		case EVENT_EXPEDITION: $return = "EXPEDITION"; break;
		case EVENT_ROCKET_ATTACK: $return = "ROCKET_ATTACK"; break;
		case EVENT_RETURN: $return = "RETURN_FLY"; break;
		case EVENT_TELEPORT_PLANET: $return = "TELEPORT_PLANET"; break;
		default: $return = "UNKOWN"; break;
		}
		// Hook::event("NS_GET_MISSION_NAME", array(&$return));
		return Core::getLanguage()->getItem($return);
	}

	/**
	* Returns a list of user's planets.
	*
	* @return array
	*/
	public static function getPlanetStack()
	{
		return self::$planetStack;
	}

	/**
	* Returns building time due to the used resources.
	*
	* @param integer	Used metal
	* @param integer	Used silicon
	* @param integer	Building mode (1: building, 2: unit, 3: research)
	*
	* @return integer	Time in seconds
	*/
	public static function getBuildingTime($metal, $silicon, $mode)
	{
		switch($mode)
		{
		case UNIT_TYPE_CONSTRUCTION:
			$robo_factory = NS::getPlanet()->getBuilding(UNIT_ROBOTIC_FACTORY);
			$time = (($metal + $silicon) / 2500.0) * (1 / ($robo_factory + 1)) * pow(0.5, NS::getPlanet()->getAutoNanoFactory());
			$time /= NS::getGalaxyParam(null, "PLANET_CONSTRUCTION_SPEED_FACTOR", 1);
			$time /= pow(2, (NS::getPlanet()->getData("build_factor")-1));
			break;

		case UNIT_TYPE_MOON_CONSTRUCTION:
			$robo_factory = NS::getPlanet()->getBuilding(UNIT_MOON_ROBOTIC_FACTORY);
			$time = (($metal + $silicon) / 2500.0) * (1 / ($robo_factory + 1)) * pow(0.5, NS::getPlanet()->getAutoNanoFactory());
			$time /= pow(2, (NS::getPlanet()->getData("build_factor")-1));

			$diameter = NS::getPlanet()->getData("diameter");
			if($diameter > 0 && $diameter <= TEMP_MOON_SIZE_MAX)
			{
				$time /= NS::getGalaxyParam(null, "TEMP_MOON_CONSTRUCTION_SPEED_FACTOR", 1);
			}
			else
			{
				$time /= NS::getGalaxyParam(null, "MOON_CONSTRUCTION_SPEED_FACTOR", 1);
			}
			break;

		case UNIT_TYPE_FLEET:
			$shipyard_level = NS::getPlanet()->getBuilding(UNIT_SHIPYARD, true);
			$time = (($metal + $silicon) / 5000.0) * (2 / ($shipyard_level + 1)) * pow(0.5, NS::getPlanet()->getAutoNanoFactory());
			$time /= NS::getGalaxyParam(null, "FLEET_BUILDING_SPEED_FACTOR", 1);
			$time /= pow(2, (NS::getPlanet()->getData("build_factor")-1));
			break;

		case UNIT_TYPE_DEFENSE:
			$shipyard_level = NS::getPlanet()->getBuilding(UNIT_DEFENSE_FACTORY, true);
			$time = (($metal + $silicon) / 5000.0) * (2 / ($shipyard_level + 1)) * pow(0.5, NS::getPlanet()->getAutoNanoFactory());
			$time /= NS::getGalaxyParam(null, "DEFENSE_BUILDING_SPEED_FACTOR", 1);
			$time /= pow(2, (NS::getPlanet()->getData("build_factor")-1));
			break;

		case UNIT_TYPE_RESEARCH:
			// $time = NS::getResearchTime($metal + $silicon);
			$time = ($metal + $silicon) / (1000.0 * (1 + NS::getPlanet()->getReseachVirtLab()));
			$time /= NS::getGalaxyParam(null, "RESEARCH_SPEED_FACTOR", 1);
			$time /= pow(2, (NS::getUser()->get("research_factor")-1));
			break;
		}

		$time *= 3600;
		// $time /= $factor;
		if(GAMESPEED > 0)
		{
			$time *= floatval(GAMESPEED);
		}
		// Hook::event("CONSTRUCTION_TIME", array(&$time, $metal, $silicon, $mode));
		if($time > 1)
			return max(1, round($time));
		return max(0.001, $time);
	}

	/**
	* Returns the range for rocket attacks.
	*
	* @return integer
	*/
	public static function getRocketRange()
	{
		return self::getResearch(UNIT_IMPULSE_ENGINE) * /* 5 */ 8 - 1;
	}

	/**
	* Returns the capacity of the current rocket station.
	*
	* @return integer
	*/
	public static function getRocketStationSize()
	{
		return self::$planet->getBuilding(UNIT_ROCKET_STATION) * 15;
	}

	/**
	* Returns the capacity of the shield units.
	*
	* @return integer
	*/
	public static function getShieldFields()
	{
		return self::getResearch(UNIT_SHIELD_TECH)*10;
	}

	/**
	* Returns the capacity of the repair factory.
	*
	* @return integer
	*/
	public static function getRepairStorage()
	{
		return self::$planet->getStorage("repair");
	}

	/**
	* Returns the loaded requirements.
	*
	* @return array
	*/
	public static function getAllRequirements()
	{
		self::loadRequirements();
		return self::$requirements;
	}

	/**
	* Instantiates a page class.
	*
	* @param string	Page name
	*
	* @return Page
	*/
	public static function getPage($page)
	{
		// currentPage removed
		if($pageClass = self::factory("page", $page))
		{
			return new $pageClass();
		}
		/* if(!class_exists($page))
		{
		$page = "Main";
		} */
		// $def_page = "Main";
        if(DEATHMATCH){
            $def_page = 'Ranking';
        }else{
            $def_page = (NS::getUser() && NS::getUser()->get("points") >= 50) || mt_rand(1, 100) <= 0 ? "Main" : "Stock";
        }
		if($page != $def_page && $pageClass = self::factory("page", $def_page))
		{
			return new $pageClass();
		}
		throw new GenericException("Unable to load page '".$page."'.");
	}

	/**
	* Executes a formula.
	*
	* @param string	Formular to execute
	* @param integer	First number to replace
	* @param integer	Second number to replace
	*
	* @return integer	Result
	*/
	public static function parseFormula($formula, $basic, $level)
	{
		// Hook::event("PARSE_FORMULA", array(&$formula, &$basic, &$level));
		$formula = Str::replace("{level}", $level, $formula);
		$formula = Str::replace("{basic}", $basic, $formula);
		$formula = Str::replace("{temp}", NS::getPlanet()->getData("temperature"), $formula);
		$formula = preg_replace_callback("#\{tech\=([0-9]+)\}#i", function($m){ return NS::getResearch($m[1]); }, $formula);
		$formula = preg_replace_callback("#\{building\=([0-9]+)\}#i", function($m){ return NS::getPlanet()->getBuilding($m[1]); }, $formula);
		$result = 0;
		eval("\$result = ".$formula.";");
		return round($result);
	}

	public static function getGraviFuelConsumeScale($uid = null)
	{
		if(is_null($uid))
		{
			$uid = NS::getUser()->get("userid");
		}
		$formula = NS::getResearchSpecial(UNIT_GRAVI, $uid);
		if($formula)
		{
			$scale = NS::parseFormula($formula, 0, NS::getResearch(UNIT_GRAVI, $uid)) / 100.0;
			return $scale > 0 ? $scale : 1.0;
		}
		return 1.0;
	}

	/**
	* Loads a model and instantiate it.
	*
	* @param string	Model name
	* @param mixed		Initial loaded resource
	*
	* @return Model
	*/
	public static function getModel($model, $resource = null)
	{
		if($modelClass = self::factory("models", $model))
		{
			return new $modelClass($resource);
		}
		throw new GenericException("Unable to load model '".$model."'.");
	}

	/**
	* Loads a system class.
	*
	* @param string	System class name
	*
	* @return Model
	*/
	public static function getSystemClassname($name)
	{
		if($classname = self::factory(".", $name))
		{
			return $classname;
		}
		throw new GenericException("Unable to load system class '".$name."'.");
	}

	/**
	* Internal class factory to search for extensions.
	*
	* @param string	Class type path
	* @param string	Class path
	*
	* @return string	Class name
	*/
	public static function factory($pathName, $className)
	{
		// $pathName = getClassPath($pathName);
		$className = getClassPath($className);
		$class = $className;
		$extFilename = APP_ROOT_DIR."ext/".$pathName."/Ext".$className.".class.php";
		$stdFilename = APP_ROOT_DIR."game/".$pathName."/".$className.".class.php";
		// debug_var(array($pathName, $className, $ext, $classes), "[factory] $pathName, $className");
		$classFilename = false;
		if(file_exists($extFilename))
		{
			$classFilename = $extFilename;
			$class = "Ext".$className;
			// debug_var(array($pathName, $className, $ext, $classes), "[factory] $pathName, $className");
		}
		else if(file_exists($stdFilename))
		{
			$classFilename = $stdFilename;
		}
		if($classFilename !== false)
		{
			if(Str::inString("/", $class))
			{
				$class = Str::substring(strrchr("/", $class), 1);
			}
			require_once($classFilename);
			return $class;
		}
		return false;
	}

	/**
	* Returns all modorator.
	*
	* @return integer	User id
	*/
	public static function getModerators()
	{
		$mods = array();
		$result = sqlSelect("user2group", array("userid"), "", "usergroupid = '2' OR usergroupid = '4'");
		while($row = sqlFetch($result))
		{
			if(!in_array($row["userid"], $mods))
			{
				$mods[] = $row["userid"];
			}
		}
		return $mods;
	}

	/**
	* Returns a random Modorator.
	*
	* @return integer	User id
	*/
	public static function getRandomModerator()
	{
		$mods = NS::getModerators();
		$i = mt_rand(0, count($mods) - 1);
		return isset($mods[$i]) ? $mods[$i] : 1;
	}

	/**
	* Updates user resourses and credit.
	*
	* @return array
	*/
	public static function updateUserRes($params)
	{
		if(!mt_rand(0, 10000))
		{
			// it's used in ExtPayment, check_max_credit
			sqlDelete("res_log", "time < ".sqlVal(date("Y-m-d H:i:s", time()-60*60*24*5)));
		}

		if(!isset($params["type"])) // && isset($params["event_mode"]))
		{
			$params["type"] = RES_UPDATE_COST;
		}
		if( $params['type'] == RES_UPDATE_EXCHANGE )
		{
			updateUserState($params['userid'], STATE_RES_UPDATE_EXCHANGE);
		}
		$type = $params["type"];
		$event_mode = isset($params["event_mode"]) ? $params["event_mode"] : null;
		$userid = isset($params["userid"]) ? $params["userid"] : null;
		$planetid = isset($params["planetid"]) ? $params["planetid"] : null;
		$metal = isset($params["metal"]) ? $params["metal"] : 0;
		$silicon = isset($params["silicon"]) ? $params["silicon"] : 0;
		$hydrogen = isset($params["hydrogen"]) ? $params["hydrogen"] : 0;
		$credit = isset($params["credit"]) ? $params["credit"] : 0;

		$max_metal = isset($params["max_metal"]) ? max(0, $params["max_metal"]) : null;
		$max_silicon = isset($params["max_silicon"]) ? max(0, $params["max_silicon"]) : null;
		$max_hydrogen = isset($params["max_hydrogen"]) ? max(0, $params["max_hydrogen"]) : null;

		$block_minus = isset($params["block_minus"]) ? $params["block_minus"] : false;

		$params["minus"] = false;
		$params["result_metal"] = null;
		$params["result_silicon"] = null;
		$params["result_hydrogen"] = null;
		$params["result_credit"] = null;
        $params["game_credit"] = null;

		if(!$userid && !$planetid)
		{
			return false;
		}

		if(!isset($params["ownerid"]))
		{
			$params["ownerid"] = null;
		}

		$update_res_types = array(
			RES_UPDATE_PLANET_PRODUCTION => "+-",
			RES_UPDATE_EXCHANGE => "+-",
			RES_UPDATE_COST => "-",
			RES_UPDATE_CANCEL => "+-",
			RES_UPDATE_DISASSEMBLE => "+",
			RES_UPDATE_UNLOAD => "+",
			RES_UPDATE_VIEW_GALAXY => "-",
			RES_UPDATE_MONITOR_PLANET => "-",
			RES_UPDATE_EXPEDITION_CREDITS => "+-",
			RES_UPDATE_BUY_CREDITS => "+",
			RES_UPDATE_ADD_REF_CREDITS => "+",
			// RES_UPDATE_EXCH_LOT_UNLOAD => "+",
			RES_UPDATE_EXCH_LOT_RESERVE => "-",
			RES_UPDATE_EXCH_LOT_BUY => "-",
			RES_UPDATE_EXCH_LOT_SELL => "+",
			RES_UPDATE_EXCH_OWNER_PROFIT => "+",
			RES_UPDATE_EXCH_DEFENDER_PROFIT => "+",
			RES_UPDATE_FIX => "+-",
			RES_UPDATE_EXCH_LOT_COMISSION => "+",
			RES_UPDATE_BUY_ARTEFACT => "-",
			RES_UPDATE_VIP_START => "-",
			RES_UPDATE_ACHIEVEMENT => "+",
			RES_UPDATE_EXCH_FUEL_REST => "+",
			RES_UPDATE_ALIEN_GRAB_CREDIT => "-",
			RES_UPDATE_ALIEN_GIFT_CREDIT => "+",
			RES_UPDATE_EXCH_LOT_PREMIUM => "-",
			);

		if(isset($update_res_types[$params["type"]]))
		{
			switch($update_res_types[$params["type"]])
			{
				case "+":
					$metal = max(0, $metal);
					$silicon = max(0, $silicon);
					$hydrogen = max(0, $hydrogen);
					$credit = max(0, $credit);
					break;

				case "-":
					$metal = min(0, $metal);
					$silicon = min(0, $silicon);
					$hydrogen = min(0, $hydrogen);
					$credit = min(0, $credit);
					break;
			}
		}

		$is_planet_data_changed = false;
		//t($planetid, 'update.$planetid');
		//t($metal, 'update.$$metal');
		//t($silicon, 'update.$$silicon');
		//t($hydrogen, 'update.$$hydrogen');
		if($planetid && ($metal || $silicon || $hydrogen))
		{
			$update_production = isset($params["update_production"]) ? $params["update_production"] : true;
			if($type != RES_UPDATE_PLANET_PRODUCTION && $update_production
				&& (!NS::getPlanet() || $planetid != NS::getPlanet()->getPlanetId())
				)
			{
				new Planet($planetid, null);
			}

			if($block_minus)
			{
				$row = sqlSelectRow("planet", array("metal", "silicon", "hydrogen"), "", "planetid = ".sqlVal($planetid));
				$params["before_metal"] = $row["metal"];
				$params["before_silicon"] = $row["silicon"];
				$params["before_hydrogen"] = $row["hydrogen"];
			}

			$cut_res_sql = $type == RES_UPDATE_PLANET_PRODUCTION ? "GREATEST(0," : "(";
			// Updated By Pk
			sqlQuery("UPDATE ".PREFIX."planet SET "
				. " metal = $cut_res_sql".(isset($max_metal) ? "LEAST(".sqlVal($max_metal)."," : "(")."metal+".sqlVal($metal)."))"
				. ", silicon = $cut_res_sql".(isset($max_silicon) ? "LEAST(".sqlVal($max_silicon)."," : "(")."silicon+".sqlVal($silicon)."))"
				. ", hydrogen = $cut_res_sql".(isset($max_hydrogen) ? "LEAST(".sqlVal($max_hydrogen)."," : "(")."hydrogen+".sqlVal($hydrogen)."))"
				. ", last = ".sqlVal(time())
				. " WHERE planetid = ".sqlVal($planetid));

			$row = sqlSelectRow("planet", array("metal", "silicon", "hydrogen"), "", "planetid = ".sqlVal($planetid));
			$params["result_metal"] = $row["metal"];
			$params["result_silicon"] = $row["silicon"];
			$params["result_hydrogen"] = $row["hydrogen"];

			if($block_minus)
			{
				if($params["result_metal"] < 0 || $params["result_silicon"] < 0 || $params["result_hydrogen"] < 0)
				{
					$auto_fixed = false;
					if(!empty($params["auto_fix"]) && is_array($params["auto_fix"]))
					{
						$auto_fixed = array(
							"metal" => 0,
							"silicon" => 0,
							"hydrogen" => 0,
							);
						foreach($params["auto_fix"] as $res_name => $res_value)
						{
							if($params["result_".$res_name] < 0 && $params["auto_fix"] != 0)
							{
								if( floor(abs($params["result_".$res_name])) <= ceil($res_value) )
								{
									$auto_fixed[$res_name] = min($res_value, abs($params["result_".$res_name]));
									$auto_fixed["processed"] = true;
								}
								else
								{
									$auto_fixed = false;
									break;
								}
							}
						}
					}
					if(is_array($auto_fixed) && !empty($auto_fixed["processed"]))
					{
						unset($auto_fixed["processed"]);
						$params["auto_fixed"] = $auto_fixed;

						// Updated By Pk
						sqlQuery("UPDATE ".PREFIX."planet SET "
							. " metal = GREATEST(0, metal)"
							. ", silicon = GREATEST(0, silicon)"
							. ", hydrogen = GREATEST(0, hydrogen)"
							. " WHERE planetid = ".sqlVal($planetid));

						$params["result_metal"] = max(0, $params["result_metal"]);
						$params["result_silicon"] = max(0, $params["result_silicon"]);
						$params["result_hydrogen"] = max(0, $params["result_hydrogen"]);
					}
					else
					{
						$params["minus_blocked"] = true;

						// Updated By Pk
						sqlQuery("UPDATE ".PREFIX."planet SET "
							. " metal = ".sqlVal(max(0, $params["before_metal"]))
							. ", silicon = ".sqlVal(max(0, $params["before_silicon"]))
							. ", hydrogen = ".sqlVal(max(0, $params["before_hydrogen"]))
							. " WHERE planetid = ".sqlVal($planetid));

						return $params;
					}
				}
			}

			$is_planet_data_changed = true;
		}
		else if($planetid && $type == RES_UPDATE_PLANET_PRODUCTION)
		{
			// Updated By Pk
			sqlQuery("UPDATE ".PREFIX."planet SET "
				. " last = ".sqlVal(time())
				. " WHERE planetid = ".sqlVal($planetid));
		}

		$is_credit_changed = false;
		if($userid && $credit)
		{
			// Updated By Pk
			sqlQuery("UPDATE ".PREFIX."user SET "
				. " credit = credit + ".sqlVal($credit)
				. " WHERE userid = ".sqlVal($userid));

			$params["result_credit"] = sqlSelectField("user", "credit", "", "userid = ".sqlVal($userid));
            $params["game_credit"] = sqlSelectField("user", "sum(credit)");

			$is_credit_changed = true;
		}

		if($is_planet_data_changed || $is_credit_changed)
		{
			$log_row = false;
			if($type == RES_UPDATE_PLANET_PRODUCTION || $type == RES_UPDATE_DISASSEMBLE)
			{
				$where = array(
					!is_null($userid) ? "userid=".sqlVal($userid) : "userid is null",
					!is_null($planetid) ? "planetid=".sqlVal($planetid) : "planetid is null",
					);
				$log_row = sqlSelectRow("res_log", "*", "",
					implode(" AND ", $where),
					"id DESC", 1);
				if($log_row["type"] != $type)
				{
					$log_row = false;
				}
			}

			if(!$log_row)
			{
				$params["id"] = sqlInsert("res_log", array(
					"type" => $type,
					"planetid" => $planetid,
					"userid" => $userid,
					"metal" => $metal,
					"silicon" => $silicon,
					"hydrogen" => $hydrogen,
					"credit" => $credit,
					"result_metal" => $params["result_metal"],
					"result_silicon" => $params["result_silicon"],
					"result_hydrogen" => $params["result_hydrogen"],
					"result_credit" => $params["result_credit"],
                    "game_credit" => $params["game_credit"],
					"ownerid" => $params["ownerid"],
					"event_mode" => $event_mode,
					));
			}
			else
			{
				// Updated By Pk
				$sql = "UPDATE ".PREFIX."res_log SET "
					. " time = NOW()"
					. ", ownerid = ".sqlVal($params["ownerid"])
					. ", cnt = cnt + 1"
					. ", event_mode = ".sqlVal($event_mode)
					;

				if($is_planet_data_changed)
				{
					$sql .= ""
						. ", metal = metal + ".sqlVal($metal)
						. ", silicon = silicon + ".sqlVal($silicon)
						. ", hydrogen = hydrogen + ".sqlVal($hydrogen)
						. ", result_metal=".sqlVal($params["result_metal"])
						. ", result_silicon=".sqlVal($params["result_silicon"])
						. ", result_hydrogen=".sqlVal($params["result_hydrogen"])
						;
				}

				if($is_credit_changed)
				{
					$sql .= ""
						. ", credit = credit + ".sqlVal($credit)
						. ", result_credit=".sqlVal($params["result_credit"])
						. ", game_credit=".sqlVal($params["game_credit"])
						;
				}

				$sql .= " WHERE id=".sqlVal($log_row["id"]);
				sqlQuery($sql);

				$params["id"] = $log_row["id"];
			}

			if($is_planet_data_changed
				// && !empty($params["reload_planet"])
				// && $type != RES_UPDATE_PLANET_PRODUCTION
				&& NS::getPlanet()
				&& $planetid == NS::getPlanet()->getPlanetId()
				)
			{
				// NS::reloadPlanet();
				NS::getPlanet()->setData("metal", $params["result_metal"]);
				NS::getPlanet()->setData("silicon", $params["result_silicon"]);
				NS::getPlanet()->setData("hydrogen", $params["result_hydrogen"]);
			}

			if($is_credit_changed
				&& !empty($params["reload_user"])
				// && $type != RES_UPDATE_PLANET_PRODUCTION
				&& NS::getUser()
				&& $userid == NS::getUser()->get("userid")
				)
			{
				NS::getUser()->rebuild();
			}
		}
		return $params;
	}

    public static function updateUserDmPoints($uid = null)
    {
		if($uid === null){
			$uid = NS::getUser()->get('userid');
		}
        $update = array();
        $update[] = updateDmPointsSetSql();
		// Updated By Pk
		Core::getDB()->query('UPDATE ' . PREFIX . 'user SET ' . implode(', ', $update) . ' WHERE userid = ' . sqlVal($user_id));
    }

	public static function updateUserPoints( $params )
	{
		if( empty($params['user_id']) )
		{
			$params['user_id'] = NS::getUser()->get('userid');
		}
		$user_id	= $params['user_id'];
		$update 	= array( "points" => 0 );
		foreach( array(
				"b",
				"r",
				"u",
				// "a",
			) as $prefix)
		{
			$points_field = $prefix."_points";
			$count_field = $prefix."_count";
			if( isset($params['points'][$points_field]) && isset($params['points'][$count_field]) )
			{
				$update["points"] += $params['points'][$points_field];
				$update[$points_field] = $params['points'][$points_field];
				$update[$count_field] = $params['points'][$count_field];
			}
		}
		$op = !empty($params['inc']) ? "+" : "-";
		foreach($update as $field => $value)
		{
			$update[$field] = "$field = GREATEST(0, $field $op ".sqlVal($value).")";
		}
        $update[] = updateDmPointsSetSql();
		// Updated By Pk
		Core::getDB()->query('UPDATE ' . PREFIX . 'user SET ' . implode(', ', $update) . ' WHERE userid = ' . sqlVal($user_id));
	}

	public static function isFirstRun($name)
	{
		if( defined('YII_CONSOLE') )
		{
			return true;
		}
		$value = true;
		$name = "NS:isFirstRun:".$name;
		if ( !self::$mch->is_valid() )
		{
			return true;
		}
		if(self::$mch->add($name, $value, 2))
		{
			// self::$mch->set($name, $value, 60);
			return true;
		}
		return false;
	}

	public static function getMoonid($planetid)
	{
		static $cache = array();
		if(isset($cache[$planetid]))
		{
			return $cache[$planetid];
		}
		$moonid = sqlSelectField("galaxy", "moonid", "", "planetid=".sqlVal($planetid));
		if($moonid)
		{
			return $cache[$planetid] = $moonid;
		}
		return $cache[$planetid] = $planetid;
	}

	public static function getMaxFleetControls($owner_userid, $location_userid)
	{
		$comp_tech = NS::getResearch(UNIT_COMPUTER_TECH, $owner_userid);
		if( $owner_userid != $location_userid && $location_userid )
		{
			$comp_tech += NS::getResearch(UNIT_COMPUTER_TECH, $location_userid);
		}
		return 1 + floor($comp_tech / 6);
	}

	public static function getRemainFleetControls($event)
	{
		return max(0, NS::getMaxFleetControls($event["user"], NS::getUser()->get("userid")) - (isset($event["data"]["control_times"]) ? $event["data"]["control_times"] : 0));
	}

	public static function calcFleetParams($ships, $params = array())
	{
		// $calc_consumption = isset($params["calc_consumption"]) ? $params["calc_consumption"] : false;
		// $uid = $calc_consumption ? (isset($params["userid"]) ? $params["userid"] : NS::getUser()->get("userid")) : 0;
		$uid = isset($params["userid"]) ? $params["userid"] : 0;
		$speed_modifier = isset($params["speed_modifier"]) ? $params["speed_modifier"] : 1;

		$data = array();
		$select = array("d.unitid", "d.capicity", "d.speed", "d.consume", "b.name", "b.mode");
		$joins	= "LEFT JOIN ".PREFIX."construction b ON (b.buildingid = d.unitid)";
		$result = sqlSelect("ship_datasheet d", $select, $joins, "b.mode=".UNIT_TYPE_FLEET." AND b.buildingid IN (".sqlArray(array_keys($ships)).")");
		$capicity = 0;
		$consumption = 0;
		$speed = MAX_SPEED_POSSIBLE;
		$fleet_size = 0;
		while($row = sqlFetch($result))
		{
			$id = $row["unitid"];
			$quantity = is_array($ships[$id]) ? $ships[$id]["quantity"] : $ships[$id];
			if($quantity > 0)
			{
				$data[$id]["name"] = $row["name"];
				$data[$id]["mode"] = $row["mode"];
				$data[$id]["quantity"] = $quantity;

				if($uid)
				{
					$speed = min($speed, NS::getSpeed($id, $row["speed"], $uid) * $speed_modifier );
					$consumption += ceil($row["consume"] * FLEET_FUEL_CONSUMPTION * NS::getGraviFuelConsumeScale($uid)) * $data[$id]["quantity"];
				}
				$capicity += $row["capicity"] * $data[$id]["quantity"];
				$fleet_size += $row["consume"] > 0 ? $data[$id]["quantity"] : 0;
			}
		}
		sqlEnd($result);

		if($speed == MAX_SPEED_POSSIBLE) // no ships
		{
			$speed = 0;
		}

		return array(
			"ships" => $data,
			"capicity" => $capicity, // compatible
			"capacity" => $capicity,
			"consumption" => $consumption,
			"maxspeed" => $speed,
			"speed_factor" => NS::$fleet_speed_factor,
			"fleet_size" => $fleet_size,
			);
	}

	public static function findFleetTarget($galaxy, $system, $pos, $maxspeed, $userid, $except_planetid)
	{
		$best_dist = null;
		$best_row = null;
		$result = sqlSelect("planet p", array(
			"p.planetid",
			"p.ismoon",
			"IFNULL(g.galaxy, gm.galaxy) as galaxy",
			"IFNULL(g.system, gm.system) as system",
			"IFNULL(g.position, gm.position) as position",
			), "
			LEFT JOIN ".PREFIX."galaxy g ON g.planetid = p.planetid AND g.destroyed=0
			LEFT JOIN ".PREFIX."galaxy gm ON g.moonid = p.planetid AND gm.destroyed=0
			", "p.userid = ".sqlVal($userid).($except_planetid ? " AND p.planetid != ".sqlVal(except_planetid) : ""), "p.ismoon DESC");
		while($row = sqlFetch($result))
		{
			if($row["galaxy"])
			{
				$dist = NS::getDistance($galaxy, $system, $pos, $row["galaxy"], $row["system"], $row["position"]);
				if($best_dist === null || $best_dist > $dist)
				{
					$best_dist = $dist;
					$best_row = $row;
				}
			}
		}
		sqlEnd($result);

		if($best_dist === null)
		{
			return false;
		}
		$best_row["time"] = max(60, NS::getFlyTime($best_dist, $maxspeed));
		return $best_row;
	}

	public static function isPlanetUnderAttack($planetid = null)
	{
		// if(!isAdmin(null, true)) return false;
		if($planetid === null)
		{
			$planetid = NS::getUser()->get("curplanet"); // NS::getPlanet()->getPlanetid();
		}
		static $cache = array();
		if(isset($cache[$planetid]))
		{
			return $cache[$planetid];
		}
		$row = sqlQueryRow("SELECT planetid, accomplished FROM ".PREFIX."assault ORDER BY assaultid DESC LIMIT 1");
		if($row["planetid"] == $planetid && !$row["accomplished"]){
			return $cache[$planetid] = true;
		}
		return $cache[$planetid] = false;
		/*
		$modes = array(EVENT_ATTACK_SINGLE, EVENT_ATTACK_ALLIANCE, EVENT_ROCKET_ATTACK, EVENT_ALLIANCE_ATTACK_ADDITIONAL, EVENT_ATTACK_DESTROY_BUILDING, EVENT_ATTACK_ALLIANCE_DESTROY_BUILDING,
					EVENT_ATTACK_DESTROY_MOON, EVENT_ATTACK_ALLIANCE_DESTROY_MOON, EVENT_ALIEN_FLY_UNKNOWN, EVENT_ALIEN_ATTACK, EVENT_ALIEN_ATTACK_CUSTOM, );
		$cur_time = time();
		$row = sqlSelectRow("events", "*", "",
			"mode in (".sqlArray($modes).") AND destination=".sqlVal($planetid)." AND time BETWEEN ".sqlVal($cur_time-60*5)." AND ".sqlVal($cur_time+1), "time DESC, eventid DESC", 1);
		return $cache[$planetid] = ($row["processed"] == EVENT_PROCESSED_START || ($row["processed"] == EVENT_PROCESSED_WAIT && abs($row["time"] - $cur_time) <= 1));
		*/
	}

	/**
	* Deletes all planet data.
	*
	* @param integer	Planet id
	* @param integer	User id
	* @param boolean	Planet is moon
	*
	* @return void
	*/
	public static function deletePlanet($planetid, $userid, $ismoon, $from_event = false)
	{
		if( $userid )
		{
			$stat = PointRenewer::getPlanetBuildingStats($planetid);
			$build_points = $stat["points"];
			$build_count = $stat["count"];

			$stat = PointRenewer::getPlanetFleetStats($planetid);
			$unit_points = $stat["points"];
			$unit_count = $stat["count"];

			$sum_points = $build_points + $unit_points;
		}

		$eh = EventHandler::getEH();
		if($eh)
		{
			$eh->removePlanetEvents($planetid, $from_event ? array("except_modes" => EVENT_TEMP_PLANET_DISAPEAR) : array());
		}

		if( !$ismoon )
		{
			$row = sqlSelectRow("galaxy", "moonid", "", "planetid = ".sqlVal($planetid));
			if($row["moonid"])
			{
				self::deletePlanet($row["moonid"], $userid, 1);
			}
		}

		if( $userid )
		{
			// Update By Pk
			sqlQuery("UPDATE ".PREFIX."user SET "
				. " points = points - ".sqlVal($sum_points)
				. ", b_points = b_points - ".sqlVal($build_points)
				. ", b_count = b_count - ".sqlVal($build_count)
				. ", u_points = u_points - ".sqlVal($unit_points)
				. ", u_count = u_count - ".sqlVal($unit_count)
				. " WHERE userid = ".sqlVal($userid));
		}

		// Update By Pk
		sqlUpdate("planet", array("userid" => null), "planetid = ".sqlVal($planetid));

		if( $ismoon )
		{
			sqlDelete("planet", "planetid = ".sqlVal($planetid));
			// Update By Pk
			sqlUpdate("galaxy", array("moonid" => null), "moonid = ".sqlVal($planetid));
		}
		else
		{
			// Update By Pk
			sqlUpdate("galaxy", array("destroyed" => 1), "planetid = ".sqlVal($planetid));
		}
		// Hook::event("DELETE_PLANET", array($id, $userid, $ismoon));
	}

	public static function retreatFleet($id)
	{
		if( !NS::isFirstRun('NS.retreatFleet.' . $id) )
		{
			return RETREAT_FLEET_ALREADY_DONE;
		}

		if( NS::isPlanetUnderAttack() )
		{
			return RETREAT_FLEET_PLANET_UNDER_ATTACK;
		}

		$fleetEvents = NS::getEH()->getFleetEvents();
		if( isset($fleetEvents[$id]) && in_array($fleetEvents[$id]["mode"], $GLOBALS["RETREAT_FLEET_EVENTS"]) )
		{
			switch($fleetEvents[$id]["mode"])
			{
				case EVENT_HOLDING:
				case EVENT_HALT:
				case EVENT_TRANSPORT:
                    break;

				default:
					if($fleetEvents[$id]["user"] != NS::getUser()->get("userid") && !isAdmin(null, true)){
						return RETREAT_FLEET_NOT_OWNER;
					}
            }
            if(!NS::getEH()->removeEvent($id)){
                return RETREAT_FLEET_ALREADY_DONE;
            }
			switch($fleetEvents[$id]["mode"])
			{
				case EVENT_HOLDING:
				case EVENT_HALT:
					if( $fleetEvents[$id]["user"] != NS::getUser()->get("userid") )
					{
						$data = $fleetEvents[$id]["data"];
						$data["planetid"] = $fleetEvents[$id]["planetid"];
						$data["destination"] = $fleetEvents[$id]["destination"];
						new AutoMsg(MSG_RETREAT_OTHER, $fleetEvents[$id]["user"], time(), $data);
					}
					break;

				case EVENT_TRANSPORT:
					if( $fleetEvents[$id]["user"] != NS::getUser()->get("userid") )
					{
						$data = $fleetEvents[$id]["data"];
						$data["planetid"] = $fleetEvents[$id]["planetid"];
						$data["destination"] = $fleetEvents[$id]["destination"];
						new AutoMsg(MSG_RETREAT_TRANSPORT, $fleetEvents[$id]["user"], time(), $data);
					}
					break;
			}
            return RETREAT_FLEET_OK;
		}
		return RETREAT_FLEET_ALREADY_DONE;
	}

	public static function getUserLanguageId($userid)
	{
		static $cache = array();
		if(isset($cache[$userid]))
		{
			return $cache[$userid];
		}
		if(NS::getUser() && NS::getUser()->get("userid") == $userid)
		{
			return $cache[$userid] = NS::getUser()->get("languageid");
		}
		return $cache[$userid] = sqlSelectField("user", "languageid", "", "userid=".sqlVal($userid));
	}

    public static function getUnitsBuildTime($data)
    {
        $full_time = $data["duration"] * $data["quantity"];
        $min_time = ceil(5.0 / $data["duration"]) * $data["duration"];
        $max_time = $data["duration"] * clampVal(round($data["quantity"] / 100), 1, 60);
        return max(1, min($full_time, max($min_time, $max_time)));
        /*
        max(1, min($data["duration"] * $data["quantity"],
                    max(ceil(5.0 / $data["duration"]) * $data["duration"],
                        $data["duration"] * clampVal(round($data["quantity"] / 100), 1, 60))));
        */
    }

    public static function getProfessionName()
    {
        $profession = NS::getUser() ? NS::getUser()->get("profession") : 0;
        if(isset($GLOBALS["PROFESSIONS"][$profession])){
            return Core::getLanguage()->getItem($GLOBALS["PROFESSIONS"][$profession]['name']);
        }
        return Core::getLanguage()->getItem('PROFESSION_UNKNOWN');
    }

	public static function getProfessionChangeDaysRemain()
	{
        if(!NS::getUser()){
            return 0;
        }
		return max(0, PROFESSION_CHANGE_MIN_DAYS - floor((time() - NS::getUser()->get("prof_time")) / (60*60*24.0)));
	}

	public static function getProfessionChangeCost()
	{
        if(!NS::getUser()){
            return 0;
        }
		if(time() - NS::getUser()->get("prof_time") >= PROFESSION_CHANGE_MIN_DAYS * 60*60*24){
			return 0;
		}
		return PROFESSION_CHANGE_COST;
	}

    public static function applyProfession($userid, $profession, $add = true)
    {
        if(!isset($GLOBALS["PROFESSIONS"][$profession])){
            return;
        }
        $profession = $GLOBALS["PROFESSIONS"][$profession];
        if(isset($profession['tech_special'])){
            static $tech_list = array();
            $op = $add ? '+' : '-';
            foreach($profession['tech_special'] as $tech_id => $level_diff){
                if(!$level_diff){
                    continue;
                }
                if(!isset($tech_list[$tech_id])){
                    $tech_list[$tech_id] = sqlSelectField('construction',
                        'mode', '', 'buildingid='.sqlVal($tech_id));
                }
                switch($tech_list[$tech_id]){
                case UNIT_TYPE_CONSTRUCTION:
                case UNIT_TYPE_MOON_CONSTRUCTION:
                    sqlQuery("update ".PREFIX."building2planet set
                        added = added + GREATEST(-level, $op".sqlVal($level_diff)."),
                        level = GREATEST(0, level $op ".sqlVal($level_diff).")
                        where buildingid=".sqlVal($tech_id)."
                            AND planetid IN (select planetid from ".PREFIX."planet where userid=".sqlVal($userid).")
                        ORDER BY planetid, buildingid");
                    break;

                case UNIT_TYPE_RESEARCH:
                    sqlQuery("update ".PREFIX."research2user set
                        added = added + GREATEST(-level, $op".sqlVal($level_diff)."),
                        level = GREATEST(0, level $op ".sqlVal($level_diff).")
                        where buildingid=".sqlVal($tech_id)." AND userid=".sqlVal($userid)."
                        ORDER BY buildingid");
                    break;
                }
            }
        }
    }

    public static function updateProfession($userid)
    {
        Artefact::resyncUser($userid);

        /*
        $profession = sqlSelectField('user', 'profession', '', 'userid='.sqlVal($userid));
        if($profession > 0){
            Lib::beginTransaction();
            try{
                self::applyProfession($userid, $profession, false);
                self::applyProfession($userid, $profession, true);
                Lib::commitTransaction();
            }catch(Exception $e){
                Lib::rollbackTransaction();
                throw $e;
            }
        }
         *
         */
    }

    public static function checkTargetValidByAllyAttack($target_planet_id, $owner_userid = null)
    {
        if(!BLOCK_TARGET_BY_ALLY_ATTACK_USED_PERIOD){
            return true;
        }

		$users = array( NS::getUser()->get("userid") );
		if( $owner_userid && $owner_userid != NS::getUser()->get("userid") )
		{
			$users[] = $owner_userid;
		}

        /* select count(*) from na_events
        where user=51
        and eventid in (SELECT distinct parent_eventid FROM na_events
        WHERE user in (22)
        and mode in (12,18))
        limit 1 */

		$eventid = sqlQueryField("SELECT eventid FROM ".PREFIX."events"
            . " WHERE user = (SELECT userid FROM ".PREFIX."planet WHERE planetid=".sqlVal($target_planet_id).")"
            . "  AND parent_eventid IN (SELECT DISTINCT parent_eventid FROM ".PREFIX."events"
                . " WHERE mode IN (".sqlArray(EVENT_ATTACK_ALLIANCE, EVENT_ALLIANCE_ATTACK_ADDITIONAL,
                        EVENT_ATTACK_ALLIANCE_DESTROY_BUILDING, EVENT_ATTACK_ALLIANCE_DESTROY_MOON
                        ).")"
                . "	 AND user IN (".sqlArray($users).") "
                . "	 AND processed != ".EVENT_PROCESSED_WAIT
                . "  AND processed_time > ".sqlVal(time()-BLOCK_TARGET_BY_ALLY_ATTACK_USED_PERIOD).")"
            . " LIMIT 1"
        );
        return !$eventid;
    }

    public static function checkProtectionTime($time)
    {
        if(DEATHMATCH && PROTECTION_PERIOD > 0){
            $time = max($time, DEATHMATCH_START_TIME + PROTECTION_PERIOD);
            $time = min($time, DEATHMATCH_END_TIME - PROTECTION_PERIOD);
        }
        return $time > time();
    }

    public static function checkTargetBashing($target_planet_id, $owner_userid, &$bashing_info)
    {
        $bashing_info = array();
        if(!BASHING_PERIOD || !BASHING_MAX_ATTACKS){
            return true;
        }

		$users = array( NS::getUser()->get("userid") );
		if( $owner_userid && $owner_userid != NS::getUser()->get("userid") )
		{
			$users[] = $owner_userid;
		}
		$bashing_info['cur_attack_count'] = $bashing_info['attacking_count'] = sqlQueryField("SELECT count( DISTINCT e.time ) FROM ".PREFIX."events e"
			. " WHERE e.mode in (".sqlArray(EVENT_ATTACK_SINGLE, EVENT_ATTACK_ALLIANCE, EVENT_ROCKET_ATTACK, EVENT_ALLIANCE_ATTACK_ADDITIONAL,
					EVENT_ATTACK_DESTROY_BUILDING, EVENT_ATTACK_ALLIANCE_DESTROY_BUILDING,
					EVENT_ATTACK_DESTROY_MOON, EVENT_ATTACK_ALLIANCE_DESTROY_MOON
					).")"
			. "	 AND e.user in (".sqlArray($users).") "
			. "	 AND e.destination IN (SELECT planetid FROM ".PREFIX."planet WHERE userid = (SELECT userid FROM ".PREFIX."planet WHERE planetid=".sqlVal($target_planet_id)."))"
			. "	 AND e.processed=".EVENT_PROCESSED_WAIT
			// . "	 AND e.processed!=".EVENT_PROCESSED_ERROR
        );
        if($bashing_info['attacking_count'] >= BASHING_MAX_ATTACKS){
            $bashing_info['reason'] = 'attacking_count';
            return false;
        }

		$bashing_info['finished_attack_count'] = sqlQueryField("SELECT count( DISTINCT e.time ) FROM ".PREFIX."events e"
			. " WHERE e.mode in (".sqlArray(EVENT_ATTACK_SINGLE, EVENT_ATTACK_ALLIANCE, EVENT_ROCKET_ATTACK, EVENT_ALLIANCE_ATTACK_ADDITIONAL,
					EVENT_ATTACK_DESTROY_BUILDING, EVENT_ATTACK_ALLIANCE_DESTROY_BUILDING,
					EVENT_ATTACK_DESTROY_MOON, EVENT_ATTACK_ALLIANCE_DESTROY_MOON
					).")"
			. "	 AND e.user in (".sqlArray($users).") "
			. "	 AND e.destination IN (SELECT planetid FROM ".PREFIX."planet WHERE userid = (SELECT userid FROM ".PREFIX."planet WHERE planetid=".sqlVal($target_planet_id)."))"
			. "	 AND e.processed != ".EVENT_PROCESSED_WAIT
            . "  AND e.processed_time > ".sqlVal(time()-BASHING_PERIOD)
			// . "	 AND e.processed!=".EVENT_PROCESSED_ERROR
        );
        $bashing_info['cur_attack_count'] += $bashing_info['finished_attack_count'];
        if($bashing_info['cur_attack_count'] >= BASHING_MAX_ATTACKS){
            $bashing_info['reason'] = 'finished_attack_count';
            return false;
        }

        return true;
    }
	
    public static function checkSpan($message, $from = '', $auto_ban = true)
    {
        if(preg_match("#http://orion|parallax17|paralax17|parallax|paralax|orion\.parallax17|параллакс17|паралакс17|[pрп]+[aа]+[rр]+[aа]+[lл]+[aа]+[xх]+"
                ."|starwind\.galaxion|galaxion\.ru"
                ."|звездные-миры\.рф"
                ."|wildspace-online"
                ."|4bloka\.com"
                ."|eveonline\.com"
                ."|eve-ru\.com"
                ."|thejam\.ru"
                ."|classicocrack"
                ."|combats\.com"
                ."|supernova\.ws"
                ."|icedice\.org"
                ."|xnova\.us"
                ."|oga\.by"
                ."|moon-hunt\.ru"
                ."|uo\.mazar\.ru"
                ."|lordofultima\.com"
                ."|aceonline\.ru"
                ."|icedice\.org"
                ."|spacerun\.ru"
                ."|rangers\.ru"
                ."|xgame-online"
                ."|victory-online"
                ."|ogame-s"
                ."|Xnova2moons"
                ."|gnova2moons"
                ."|nova2moons"
                ."|galaxion\.su"
                ."|galaxion\.ru"
                ."|galaxion\.com"
                ."|battlespace\.ru"
                ."|battlespace\.su"
                ."|battlespace\.com"
                ."|ws-o\.ru"
                ."|spaceinvasion"
				."|xgame\.net"
				."|cosmars\.ru"
				."|ogame-online\."
				."|xterium\.ru"
				."|thexnova\.ru"
				."|starofwars\."
				."|starghosts\."
                ."#isu", $message))
        {
            if(!$auto_ban){
                return true;
            }
            $ban_u = new BanU_YII();
            $ban_u->userid = $_SESSION["userid"] ?? 0;
            $ban_u->from = 1;
            $ban_u->to = null;
            $ban_u->reason = "Violation of User Agreement, 5.13";
            $ban_u->admin_comment = date('Y-m-d H:i:s').' [AUTO BAN] '.($from ? " [$from] " : '').$message;
            $ban_u->save(false);
            exit();
        }
        return false;
    }
	
	public static function processCreditBonusItems()
	{
		if(!NS::getUser())
		{
			return false;
		}
		$user_id = NS::getUser()->get('userid');
		$planet_id = NS::getUser()->get("curplanet");
		$row = sqlSelectRow("credit_bonus_item", "*", null, "userid=".sqlUser()." AND done=0", "date", "1");
		if(NS::isFirstRun('NS.processCreditBonusItems.' . $row['id'])){
			sqlUpdate("credit_bonus_item", array("done" => 1), "id=".sqlVal($row['id']));
			$art_id = Artefact::appear($row['unitid'], $user_id, $planet_id, array());
			if($art_id !== false){
				new AutoMsg(MSG_ARTEFACT, $user_id, time(), array('art_id' => $art_id, 'mode' => 'MSG_BONUS_ARTEFACT_ADDED'));
			}
		}
	}
}
?>