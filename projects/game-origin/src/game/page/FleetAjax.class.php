<?php
/**
* Sends fleet via AjaxRequest.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

require_once(RECIPE_ROOT_DIR."AjaxRequestHelper.abstract_class.php");

class FleetAjax extends AjaxRequestHelper
{
	/**
	* Main method of sending fleet via Ajax.
	*
	* @return void
	*/
	public function __construct()
	{
		parent::__construct();
		Core::getLanguage()->load("Galaxy");

		if(NS::getResearch(UNIT_COMPUTER_TECH) + 1 <= NS::getEH()->getUsedFleetSlots() /*count(NS::getEH()->getOwnFleetEvents())*/)
		{
			Core::getLanguage()->load("mission");
			$this->display($this->format(Core::getLanguage()->getItem("NO_FREE_FLEET_SLOTS")));
		}
		// Hook::event("AJAX_SEND_FLEET", array(&$this));
		$this->setGetAction("mode", EVENT_SPY, "espionage")
			->addGetArg("espionage", "target")
			->proceedRequest();
		return;
	}

	/**
	* Index action.
	*
	* @return FleetAjax
	*/
	protected function index()
	{
		$this->display("Unkown request.");
		return $this;
	}

	/**
	* Sends espionage probes.
	*
	* @return FleetAjax
	*/
	protected function espionage($target)
	{
		if(NS::isPlanetUnderAttack())
		{
			$this->display($this->format(Core::getLanguage()->getItem("PLANET_UNDER_ATTACK")));
			return $this;
		}

		$is_moon = NS::getPlanet()->getData("ismoon");
		$select = array("p.planetid", "g.galaxy", "g.system", "g.position", "u2s.quantity", "u2s.damaged", "u2s.shell_percent", "sd.capicity", "sd.speed", "sd.consume", "b.name AS shipname");
		$joins	= "LEFT JOIN ".PREFIX."galaxy g ON (g.".($is_moon ? "moonid" : "planetid")." = p.planetid)";
		$joins .= "LEFT JOIN ".PREFIX."unit2shipyard u2s ON (u2s.planetid = p.planetid)";
		$joins .= "LEFT JOIN ".PREFIX."ship_datasheet sd ON (sd.unitid = u2s.unitid)";
		$joins .= "LEFT JOIN ".PREFIX."construction b ON (b.buildingid = u2s.unitid)";
		$row = sqlSelectRow("planet p", $select, $joins, "p.planetid = ".sqlPlanet()." AND u2s.unitid = ".UNIT_ESPIONAGE_SENSOR);
		if($row)
		{
			$select = array("g.galaxy", "g.system", "g.position", "u.points", "u.last", "u.umode", "u.observer", "u.protection_time", "b.to", "b.banid");
			$joins	= "LEFT JOIN ".PREFIX."planet p ON (g.planetid = p.planetid)";
			$joins .= "LEFT JOIN ".PREFIX."user u ON (u.userid = p.userid)";
			$joins .= "LEFT JOIN ".PREFIX."ban_u b ON (u.userid = b.userid)";
			$tar = sqlSelectRow("galaxy g", $select, $joins, "g.planetid = ".sqlVal($target));
			if(!$tar)
			{
				$joins	= "LEFT JOIN ".PREFIX."planet p ON (g.moonid = p.planetid)";
				$joins .= "LEFT JOIN ".PREFIX."user u ON (u.userid = p.userid)";
				$joins .= "LEFT JOIN ".PREFIX."ban_u b ON (u.userid = b.userid)";
				$tar = sqlSelectRow("galaxy g", $select, $joins, "g.moonid = ".sqlVal($target));
			}
			if($tar)
			{
				// Hook::event("AJAX_SEND_FLEET_ESP", array(&$this, &$row, &$tar));
				$ignoreNP = false;
				if($tar["last"] <= time() - 604800)
				{
					$ignoreNP = true;
				}
				if($tar["banid"] && (is_null($tar["to"]) || $tar["to"] >= time()))
				{
					$ignoreNP = true;
					$this->display($this->format(Core::getLanguage()->getItem("TARGET_BANNED")));
				}

				$data = array();
				// Check for newbie protection
				if($ignoreNP === false)
				{
					$isProtected = isNewbieProtected(NS::getUser()->get("points"), $tar["points"]);
					if($isProtected == 1)
					{
						$this->display($this->format(Core::getLanguage()->getItem("TARGET_TOO_WEAK")));
					}
					else if($isProtected == 2)
					{
						$this->display($this->format(Core::getLanguage()->getItem("TARGET_TOO_STRONG")));
					}
				}

				// Check for vacation mode
				if($tar["umode"] || $tar["observer"] || NS::getUser()->get("umode") || NS::getUser()->get("observer"))
				{
					$this->display($this->format(Core::getLanguage()->getItem("TARGET_IN_UMODE")));
				}

				// Get quantity
				if($row["quantity"] >= NS::getUser()->get("esps"))
				{
					$data["ships"][UNIT_ESPIONAGE_SENSOR]["quantity"] = (NS::getUser()->get("esps") > 0) ? NS::getUser()->get("esps") : 1;
				}
				else
				{
					$data["ships"][UNIT_ESPIONAGE_SENSOR]["quantity"] = $row["quantity"];
				}
				$data["ships"][UNIT_ESPIONAGE_SENSOR]["id"] = UNIT_ESPIONAGE_SENSOR;
				$data["ships"][UNIT_ESPIONAGE_SENSOR]["name"] = $row["shipname"];
				extractUnits($row, $data["ships"][UNIT_ESPIONAGE_SENSOR]["quantity"], $data["ships"][UNIT_ESPIONAGE_SENSOR]);

				$data["maxspeed"] = NS::getSpeed(UNIT_ESPIONAGE_SENSOR, $row["speed"]);
				$distance = NS::getDistance($tar["galaxy"], $tar["system"], $tar["position"]);
				$time = NS::getFlyTime($distance, $data["maxspeed"]);

				$data["consumption"] = NS::getFlyConsumption($data["ships"][UNIT_ESPIONAGE_SENSOR]["quantity"] * $row["consume"], $distance);

				if(NS::getPlanet()->getData("hydrogen") < $data["consumption"])
				{
					$this->display($this->format(Core::getLanguage()->getItem("DEFICIENT_CONSUMPTION")));
				}
				if($row["capicity"] * $data["ships"][UNIT_ESPIONAGE_SENSOR]["quantity"] - $data["consumption"] < 0)
				{
					$this->display($this->format(Core::getLanguage()->getItem("DEFICIENT_CAPACITY")));
				}

				NS::getPlanet()->setData("hydrogen", NS::getPlanet()->getData("hydrogen") - $data["consumption"]);

				$data["time"] = $time;
				$data["galaxy"] = $tar["galaxy"];
				$data["system"] = $tar["system"];
				$data["position"] = $tar["position"];
				$data["sgalaxy"] = $row["galaxy"];
				$data["ssystem"] = $row["system"];
				$data["sposition"] = $row["position"];
				$data["metal"] = 0;
				$data["silicon"] = 0;
				$data["hydrogen"] = 0;
				// Hook::event("AJAX_SEND_FLEET_ESP_START_EVENT", array($row, $tar, &$data));
				NS::getEH()->addEvent(EVENT_SPY, $time + time(), NS::getUser()->get("curplanet"), NS::getUser()->get("userid"), $target, $data);
				$this->display($this->format(sprintf(Core::getLanguage()->getItem("ESPS_SENT"), $data["ships"][UNIT_ESPIONAGE_SENSOR]["quantity"], $data["galaxy"], $data["system"], $data["position"]), 1));
			}
			else
			{
				$this->display($this->format("Unkown destination"));
			}
		}
		else
		{
			$this->display($this->format(Core::getLanguage()->getItem("DEFICIENT_ESPS")));
		}
		return $this;
	}

	/**
	* Formats the output text.
	*
	* @param string Text to format.
	* @param boolean Success or error format.
	*
	* @return string Formatted text.
	*/
	protected function format($text, $success = 0)
	{
		$class = ($success == 1) ? "success" : "notavailable";
		return "<span class=\"".$class."\" style=\"font-weight: bold;\">".$text."</span>";
	}
}
?>