<?php
/**
* Shows galaxy
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

require_once(APP_ROOT_DIR."game/UserList.class.php");

class Galaxy extends Page
{
	/**
	* Current viewing galaxy.
	*
	* @var integer
	*/
	protected $galaxy = 0;

	/**
	* Current viewing system.
	*
	* @var integer
	*/
	protected $system = 0;

	/**
	* Missile range.
	*
	* @var integer
	*/
	protected $missileRange = 0;

	/**
	* Main method to display a sun system.
	*
	* @return void
	*/
	public function __construct()
	{
		parent::__construct();
		Core::getLanguage()->load(array("Galaxy", "Statistics"));
		$this->missileRange = NS::getRocketRange();
		if(Core::getRequest()->getGET("galaxy") && Core::getRequest()->getGET("system"))
		{
			$this->setCoordinatesByGet(Core::getRequest()->getGET("galaxy"), Core::getRequest()->getGET("system"));
		}

		$this->setPostAction("submittype", "setCoordinatesByPost")
			->setPostAction("jump", "setCoordinatesByPost")
			->addPostArg("setCoordinatesByPost", "galaxy")
			->addPostArg("setCoordinatesByPost", "system")
			->addPostArg("setCoordinatesByPost", "submittype");
		$this->proceedRequest();
		return;
	}

	/**
	* Sets the coordinates by POST.
	*
	* @param integer	Galaxy
	* @param integer	System
	* @param string	Submit type
	*
	* @return Galaxy
	*/
	protected function setCoordinatesByPost($galaxy, $system, $submitType)
	{
		$this->galaxy = $galaxy;
		$this->system = $system;
		switch($submitType)
		{
		case "prevgalaxy":
			$this->galaxy--;
			break;
		case "nextgalaxy":
			$this->galaxy++;
			break;
		case "prevsystem":
			$this->system--;
			break;
		case "nextsystem":
			$this->system++;
			break;
		}
		return $this->index();
	}

	/**
	* Validates the inputted galaxy and system.
	*
	* @return Galaxy
	*/
	protected function validateInputs()
	{
		if(empty($this->galaxy))
		{
			$this->galaxy = NS::getPlanet()->getData("galaxy");
		}
		if(empty($this->system))
		{
			$this->system = NS::getPlanet()->getData("system");
		}

		if($this->galaxy < 1) { $this->galaxy = 1; }
		else if($this->galaxy > NUM_GALAXYS) { $this->galaxy = NUM_GALAXYS; }
		if($this->system < 1) { $this->system = 1; }
		else if($this->system > NUM_SYSTEMS) { $this->system = NUM_SYSTEMS; }
		return $this;
	}

	/**
	* Sets the coordinates by GET.
	*
	* @param integer	Galaxy
	* @param integer	System
	*
	* @return Galaxy
	*/
	protected function setCoordinatesByGet($galaxy, $system)
	{
		if($galaxy && $system)
		{
			$this->galaxy = $galaxy;
			$this->system = $system;
		}
		return $this;
	}

	protected function subtractHydrogen()
	{
		if($this->galaxy != NS::getPlanet()->getData("galaxy") || $this->system != NS::getPlanet()->getData("system"))
		{
			if(NS::getPlanet()->getData("hydrogen") - 10 < 0)
			{
				Logger::dieMessage("DEFICIENT_CONSUMPTION");
			}

			NS::updateUserRes(array(
				"type" => RES_UPDATE_VIEW_GALAXY,
				"userid" => NS::getUser()->get("userid"),
				"planetid" => NS::getPlanet()->getPlanetId(),
				"hydrogen" => -10,
				));
		}
		return $this;
	}

	/**
	* Index action.
	*
	* @return Galaxy
	*/
	protected function index()
	{
		$this->validateInputs()->subtractHydrogen();

		// Star surveillance
		$canMonitorActivity = false;
		if($this->galaxy == NS::getPlanet()->getData("galaxy") && NS::getPlanet()->getBuilding(UNIT_STAR_SURVEILLANCE) > 0)
		{
			$range = NS::getMonitorActivityRange();
			// $diff = abs(NS::getPlanet()->getData("system") - $this->system);
			if($range >= NS::getSystemsDiff($this->system)) // $diff)
			{
				$canMonitorActivity = true;
			}
		}
		if(!$canMonitorActivity && isAdmin())
		{
			$canMonitorActivity = true;
		}
		Core::getTPL()->assign("canMonitorActivity", $canMonitorActivity);

		// Images
		$rockimg = Image::getImage("rocket.gif", Core::getLanguage()->getItem("ROCKET_ATTACK"));

		// Get sunsystem data
		$select = array(
			"g.planetid", "g.position", "g.destroyed", "g.metal", "g.silicon",
			"p.picture", "p.planetname", "p.last as planetactivity",
			"u.username", "u.userid", "u.dm_points", "u.points", "u.e_points", "u.last as useractivity", "u.umode", "u.observer",
			"m.planetid AS moon", "m.planetid AS moonid", "m.picture AS moonpic", "m.planetname AS moonname", "m.diameter AS moonsize", "m.temperature", "m.last as moonactivity",
			"a.tag", "a.name", "a.showmember", "a.homepage", "a.showhomepage",
			"u2a.aid", "b.to", "b.banid"
		);
		$joins	= "LEFT JOIN ".PREFIX."planet p ON (p.planetid = g.planetid)";
		$joins .= "LEFT JOIN ".PREFIX."user u ON (u.userid = p.userid)";
		$joins .= "LEFT JOIN ".PREFIX."planet m ON (m.planetid = g.moonid)";
		$joins .= "LEFT JOIN ".PREFIX."user2ally u2a ON (u2a.userid = u.userid)";
		$joins .= "LEFT JOIN ".PREFIX."alliance a ON (a.aid = u2a.aid)";
		$joins .= "LEFT JOIN ".PREFIX."ban_u b ON (b.userid = u.userid)";
		$result = sqlSelect("galaxy g", $select, $joins, "g.galaxy = ".sqlVal($this->galaxy)." AND g.system = ".sqlVal($this->system).' AND g.position <= '.MAX_NORMAL_PLANET_POSITION );
		$systemData = array();
		$UserList = new UserList();
		$UserList->setKey("position");
		$UserList->setNewbieProtection(true);
        $UserList->setPointType(SHOW_DM_POINTS ? 'dm_points' : 'points');
		$UserList->setFetchRank(true);
		$UserList->setTagAsLink(false);
		$UserList->load($result);
		$sys = $UserList->getArray();
		for($i = 1; $i <= MAX_NORMAL_PLANET_POSITION; $i++)
		{
			if(isset($sys[$i]) && !$sys[$i]["destroyed"])
			{
				$sys[$i]["systempos"] = $i;
				if($sys[$i]["tag"] != "")
				{
					$sys[$i]["allydesc"] = sprintf(Core::getLanguage()->getItem("GALAXY_ALLY_HEADLINE"), $sys[$i]["tag"], $sys[$i]["alliance_rank"]);
				}

				// $sys[$i]["points"] = Link::get("game.php/Ranking", $sys[$i]["points"], Core::getLanguage()->getItem("POINTS"));
				$sys[$i]["e_points"] = Link::get("game.php/Ranking/id:e_points", $sys[$i]["e_points"], Core::getLanguage()->getItem("BATTLE_EXPERIENCE"));

				$sys[$i]["metal"] = fNumber($sys[$i]["metal"]);
				$sys[$i]["silicon"] = fNumber($sys[$i]["silicon"]);
				$sys[$i]["picture"] = Image::getImage("planets/small/s_".$sys[$i]["picture"].Core::getConfig()->get("PLANET_IMG_EXT"), $sys[$i]["planetname"], 30, 30);
				$sys[$i]["picture"] = Link::get("game.php/go:Mission/g:".$this->galaxy."/s:".$this->system."/p:".$i, $sys[$i]["picture"]);
				$sys[$i]["planetname"] = Link::get("game.php/go:Mission/g:".$this->galaxy."/s:".$this->system."/p:".$i, $sys[$i]["planetname"]);
				$sys[$i]["moonpicture"] = ($sys[$i]["moonpic"] != "") ? Image::getImage("planets/small/s_".$sys[$i]["moonpic"].Core::getConfig()->get("PLANET_IMG_EXT"), $sys[$i]["moonname"], 22, 22) : "";
				$sys[$i]["moonid"] = $sys[$i]["moonid"];
				$sys[$i]["moon"] = sprintf(Core::getLanguage()->getItem("MOON_DESC"), $sys[$i]["moonname"]);
				$sys[$i]["moonsize"] = fNumber($sys[$i]["moonsize"]);
				$sys[$i]["moontemp"] = fNumber($sys[$i]["temperature"]);

				if($sys[$i]["moonactivity"] > $sys[$i]["planetactivity"])
				{
					$activity = $sys[$i]["moonactivity"];
				}
				else
				{
					$activity = $sys[$i]["planetactivity"];
				}
				if($activity > time() - 900 && $sys[$i]["userid"] != NS::getUser()->get("userid")) { $sys[$i]["activity"] = "(*)"; }
				else if($activity > time() - 3600 && $sys[$i]["userid"] != NS::getUser()->get("userid")) { $sys[$i]["activity"] = "(".floor((time() - $activity) / 60)." min)"; }
				else { $sys[$i]["activity"] = ""; }

				if($sys[$i]["userid"] != NS::getUser()->get("userid") && $this->inMissileRange())
				{
					$sys[$i]["rocketattack"] = Link::get("game.php/RocketAttack/".$sys[$i]["planetid"], $rockimg);
					// $sys[$i]["moonrocket"] = "<tr><td colspan=&quot;3&quot;>".Str::replace("\"", "", Link::get("game.php/RocketAttack/".$sys[$i]["moon"]."/moon:1", Core::getLanguage()->getItem("ROCKET_ATTACK")))."</td></tr>";
					$sys[$i]["moonrocket"] = Str::replace("\"", "", Link::get("game.php/RocketAttack/".$sys[$i]["moonid"]."/moon:1", Core::getLanguage()->getItem("ROCKET_ATTACK")));
				}
				else
				{
					$sys[$i]["rocketattack"] = "";
					// $sys[$i]["moonrocket"] = "";
				}

				$sys[$i]["allypage"] = Str::replace("\"", "", Link::get("game.php/AlliancePage/".$sys[$i]["aid"], Core::getLanguage()->getItem("ALLIANCE_PAGE"), 1));
				if($sys[$i]["showhomepage"] && $sys[$i]["homepage"] != "" || $sys[$i]["aid"] == NS::getUser()->get("aid"))
				{
					$sys[$i]["homepage"] = "<tr><td>".Str::replace("\"", "", Link::get($sys[$i]["homepage"], Core::getLanguage()->getItem("HOMEPAGE"), 2))."</td></tr>";
				}
				else { $sys[$i]["homepage"] = ""; }
				if($sys[$i]["showmember"])
				{
					$sys[$i]["memberlist"] = "<tr><td>".Str::replace("\"", "", Link::get("game.php/MemberList/".$sys[$i]["aid"], Core::getLanguage()->getItem("SHOW_MEMBERLIST"), 3))."</td></tr>";
				}
				$sys[$i]["debris"] = Image::getImage("debris.jpg", "", 25, 25);
                if(!isMobileSkin()){
                    $sys[$i]["debris"] = Link::get("game.php/go:Mission/g:".$this->galaxy."/s:".$this->system."/p:".$i, $sys[$i]["debris"]);
                }
			}
			else
			{
				if(!isset($sys[$i]["destroyed"])) { $sys[$i] = array(); }
				$sys[$i]["systempos"] = $i;
				$sys[$i]["userid"] = null;
				$sys[$i]["metal"] = ($sys[$i]["destroyed"]) ? fNumber($sys[$i]["metal"]) : "";
				$sys[$i]["silicon"] = ($sys[$i]["destroyed"]) ? fNumber($sys[$i]["silicon"]) : "";
				$sys[$i]["picture"] = ($sys[$i]["destroyed"]) ? Image::getImage("planets/small/s_".$sys[$i]["picture"].Core::getConfig()->get("PLANET_IMG_EXT"), $sys[$i]["planetname"], 30, 30) : "";
				$sys[$i]["planetname"] = ($sys[$i]["destroyed"]) ? Core::getLanguage()->getItem("DESTROYED_PLANET") : "";
				$sys[$i]["moon"] = "";
				$sys[$i]["username"] = "";
				$sys[$i]["alliance"] = "";
			}
		}
		ksort($sys);
		Core::getTPL()->assign("goMissionLink", socialUrl(RELATIVE_URL."game.php/".Core::getRequest()->getGET("sid")."/go:Mission/")); // . ( (defined('SN')) ? ('?' . ''): ('') ) );
		Core::getTPL()->assign("sendesp", Image::getImage("esp.gif", Core::getLanguage()->getItem("SEND_ESPIONAGE_PROBE")));
		Core::getTPL()->assign("monitorfleet", Image::getImage("binocular.gif", Core::getLanguage()->getItem("MONITOR_FLEET_ACTIVITY")));
		Core::getTPL()->assign("moon", Str::replace("\"", "", Image::getImage("planets/mond".Core::getConfig()->get("PLANET_IMG_EXT"), Core::getLanguage()->getItem("MOON"), 75, 75)));
		Core::getTPL()->addLoop("sunsystem", $sys);
		Core::getTPL()->assign("debris", Str::replace("\"", "", Image::getImage("debris.jpg", Core::getLanguage()->getItem("DEBRIS"), 75, 75)));
		Core::getTPL()->assign("galaxy", $this->galaxy);
		Core::getTPL()->assign("system", $this->system);
		Core::getTPL()->display("galaxy");
		return $this;
	}

	/**
	* Checks if the current solar system is in missile range.
	*
	* @return boolean
	*/
	protected function inMissileRange()
	{
		if(Core::getOptions()->get("ATTACKING_STOPPAGE") == 1)
		{
			return false;
		}
		if($this->galaxy != NS::getPlanet()->getData("galaxy"))
		{
			return false;
		}
		// if(abs(NS::getPlanet()->getData("system") - $this->system) <= $this->missileRange)
		if(NS::getSystemsDiff($this->system) <= $this->missileRange)
		{
			return true;
		}
		return false;
	}
}
?>