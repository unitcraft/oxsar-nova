<?php
/**
* Starting rocket attacks.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class RocketAttack extends Page
{
	/**
	* Target id.
	*
	* @var integer
	*/
	protected $target = 0;

	/**
	* The target data.
	*
	* @var array
	*/
	protected $t = array();

	/**
	* Number of available rockets.
	*
	* @var integer
	*/
	protected $rockets = 0;

	/**
	* Constructor: Shows form to start rocket attack.
	*
	* @return void
	*/
	public function __construct()
	{
		if(NS::getResearch(UNIT_COMPUTER_TECH) + 1 <= NS::getEH()->getUsedFleetSlots() /*count(NS::getEH()->getOwnFleetEvents())*/)
		{
			Core::getLanguage()->load("mission");
			Logger::dieMessage("NO_FREE_FLEET_SLOTS");
		}

		$this->target = Core::getRequest()->getGET("id");
		Core::getLanguage()->load("Galaxy,info");
		$joins	= "LEFT JOIN ".PREFIX."user u ON (u.userid = p.userid)";
		$joins .= "LEFT JOIN ".PREFIX."ban_u b ON (u.userid = b.userid)";
		if(Core::getRequest()->getGET("moon"))
		{
			$joins .= "LEFT JOIN ".PREFIX."galaxy g ON (g.moonid = p.planetid)";
		}
		else
		{
			$joins .= "LEFT JOIN ".PREFIX."galaxy g ON (g.planetid = p.planetid)";
		}
		$select = array("p.planetname", "g.galaxy", "g.system", "g.position", "g.position", "u.points", "u.last", "u.umode", "u.observer", "u.protection_time", "b.to", "b.banid");
		$where = "p.planetid = ".sqlVal($this->target);
		$this->t = sqlSelectRow("planet p", $select, $joins, $where);
		if($this->t)
		{
			$ignoreNP = false;
			if($this->t["last"] <= time() - 604800)
			{
				$ignoreNP = true;
			}
			else if($this->t["banid"] && (is_null($this->t["to"]) || $this->t["to"] >= time()))
			{
				$ignoreNP = true;
				Logger::dieMessage("TARGET_BANNED");
			}

			// Hook::event("ROCKET_ATTACK_FORM_LOAD", array(&$this, &$ignoreNP));

			// Check for newbie protection
			if($ignoreNP === false)
			{
				$isProtected = isNewbieProtected(NS::getUser()->get("points"), $this->t["points"]);
				if($isProtected == 1)
				{
					Logger::dieMessage("TARGET_TOO_WEAK");
				}
				else if($isProtected == 2)
				{
					Logger::dieMessage("TARGET_TOO_STRONG");
				}
			}

			// Check for vacation mode
			if($this->t["umode"] || $this->t["observer"] || NS::getUser()->get("umode") || NS::getUser()->get("observer"))
			{
				Logger::dieMessage("TARGET_IN_UMODE");
			}

            $bashing_info = array();
            if(!NS::checkTargetBashing($this->target, null, $bashing_info)){
                Logger::dieMessage("TARGET_BASHED");
            }
            if(!NS::checkTargetValidByAllyAttack($this->target)){
                Logger::dieMessage("TARGET_BLOCKED_BY_ALLY_ATTACK_USED");
            }
            if(NS::checkProtectionTime(NS::getUser()->get("protection_time"))){
                Logger::dieMessage("TARGET_TOO_STRONG");
            }
            if(NS::checkProtectionTime($this->t["protection_time"])){
                Logger::dieMessage("TARGET_TOO_WEAK");
            }

			$this->rockets = (int)sqlSelectField("unit2shipyard", "quantity", "", "unitid = ".UNIT_INTERPLANETARY_ROCKET." AND planetid = ".sqlPlanet());

			parent::__construct();
			$this->setPostAction("start", "sendRockets")
				->addPostArg("sendRockets", "quantity")
				->addPostArg("sendRockets", "target")
				->proceedRequest();
		}
		return;
	}

	/**
	* Index action.
	*
	* @return RocketAttack
	*/
	protected function index()
	{
		$i = 0;
		$d = array();
		$result = sqlSelect("unit2shipyard u2s", array("u2s.unitid", "b.name"), "LEFT JOIN ".PREFIX."construction b ON (b.buildingid = u2s.unitid)", "b.mode = ".UNIT_TYPE_DEFENSE." AND u2s.planetid = ".sqlVal($this->target));
		while($dest = sqlFetch($result))
		{
			if($dest["unitid"] == UNIT_INTERCEPTOR_ROCKET || $dest["unitid"] == UNIT_INTERPLANETARY_ROCKET) { continue; }
			$d[$i]["name"] = Core::getLanguage()->getItem($dest["name"]);
			$d[$i]["unitid"] = $dest["unitid"];
			$i++;
		}
		sqlEnd($result);
		// Hook::event("ROCKET_ATTACK_SHOW_FORM", array(&$this, &$d));
		Core::getTPL()->assign("target", $this->t["planetname"]." ".getCoordLink($this->t["galaxy"], $this->t["system"], $this->t["position"]));
		Core::getTPL()->assign("rockets", $this->rockets);
		Core::getTPL()->addLoop("destionations", $d);
		Core::getTPL()->display("rocket_attack");
		return $this;
	}

	/**
	* Starts the actual rocket attack event.
	*
	* @param integer	Rocket quantity to send
	* @param integer	Primary target
	*
	* @return RocketAttack
	*/
	public function sendRockets($quantity, $primaryTarget)
	{
		if(!NS::isFirstRun("RocketAttack::sendRockets:".NS::getUser()->get("userid")))
		{
			Logger::dieMessage('TOO_MANY_REQUESTS');
		}

		if(NS::isPlanetUnderAttack())
		{
			Logger::dieMessage('PLANET_UNDER_ATTACK');
		}

		$diff = NS::getSystemsDiff($this->t["system"]); // abs($this->t["system"] - NS::getPlanet()->getData("system"));

		// Check max. range
		if(!Core::getOptions()->get("ATTACKING_STOPPAGE") && $this->t["galaxy"] == NS::getPlanet()->getData("galaxy") && $diff <= NS::getRocketRange())
		{
			$quantity = floor($quantity);
			if($quantity <= $this->rockets)
			{
				$this->rockets = $quantity;
			}

			if($this->rockets > 0)
			{
				// Load attacking value for interplanetary rocket
				/* $data = sqlSelectRow("ship_datasheet sds", array("sds.attack", "basic_metal", "basic_silicon", "basic_hydrogen"),
					"LEFT JOIN ".PREFIX."construction b ON (b.buildingid = sds.unitid)",
					"sds.unitid = ".UNIT_INTERPLANETARY_ROCKET); */

				$data = array();
				$data["rockets"] = $this->rockets;
				$data["ships"][UNIT_INTERPLANETARY_ROCKET]["id"] = UNIT_INTERPLANETARY_ROCKET;
				$data["ships"][UNIT_INTERPLANETARY_ROCKET]["quantity"] = $this->rockets;
				$data["ships"][UNIT_INTERPLANETARY_ROCKET]["name"] = "INTERPLANETARY_ROCKET";
				$data["sgalaxy"] = NS::getPlanet()->getData("galaxy");
				$data["ssystem"] = NS::getPlanet()->getData("system");
				$data["sposition"] = NS::getPlanet()->getData("position");
				$data["galaxy"] = $this->t["galaxy"];
				$data["system"] = $this->t["system"];
				$data["position"] = $this->t["position"];
				$data["planetname"] = $this->t["planetname"];
				$data["primary_target"] = floor($primaryTarget);
				NS::getEH()->addEvent(EVENT_ROCKET_ATTACK, time() + getRocketFlightDuration($diff), Core::getUser()->get("curplanet"), NS::getUser()->get("userid"), $this->target, $data);
			}
			doHeaderRedirection("game.php/Main", false);
		}
		return $this->index();
	}
}
?>