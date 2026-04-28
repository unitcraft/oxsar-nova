<?php
/**
* Generates espionage report.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class EspionageReport extends Planet
{
	/**
	* Buildings, research, fleet and defense data of planet.
	*
	* @var array
	*/
	protected $research = array(), $fleet = array(), $defense = array();

	/**
	* Number of probes.
	*
	* @var integer
	*/
	protected $probes;

	/**
	* Espionage tech of the commiting user.
	*
	* @var integer
	*/
	protected $espTech = 0;

	/**
	* Number of block that will be shown in report.
	*
	* @var integer
	*/
	protected $blocks = 1;

	/**
	* Final HTML string of report.
	*
	* @var string
	*/
	protected $espReport = "";

	/**
	* Chance that probes will be detected.
	*
	* @var integer
	*/
	protected $chance = 0;

	/**
	* If espionage probes were destroyed.
	*
	* @var boolean
	*/
	protected $probesLost = false;

	/**
	* Galaxy link of the target planet.
	*
	* @var string
	*/
	protected $position = "";

	/**
	* Target username
	*
	* @var string
	*/
	protected $targetName = "";

  protected $skip_unit_keys = array();

	/**
	* Sets essentail data for the espionage report.
	*
	* @param integer	Target planet id
	* @param integer	Committing user
	* @param integer	Target user id
	* @param string	Traget user name
	* @param integer	Number of espionage probes
	*
	* @return void
	*/
	public function __construct($planetid, $userid, $targetUser, $targetName, $probes)
	{
		parent::__construct($planetid, $targetUser);
		$this->userid = $userid;
		$this->dataEH = $data;
		$this->targetName = $targetName;
		$this->probes = max(1, $probes);
		
		Core::pushLanguage();
		Core::selectLanguage(NS::getUserLanguageId($userid));
		Core::getLanguage()->load(array("info" ,"EspionageReport"));

		$this->skip_unit_keys = array_flip(array(UNIT_SOLAR_SATELLITE, UNIT_INTERCEPTOR_ROCKET, UNIT_INTERPLANETARY_ROCKET, UNIT_EXCH_SUPPORT_RANGE, UNIT_EXCH_SUPPORT_SLOT));

		$this->building = array();
		$this->position = $this->getCoords(true, true); // getCoordLink($this->getData("galaxy"), $this->getData("system"), $this->getData("position"), true);
		$this->loadEspData();
		$this->generateReport();
		$this->sendESPR();
		
		Core::popLanguage();
	}

	/**
	* Saves the espionage report.
	*
	* @return EspionageReport
	*/
	protected function sendESPR()
	{
		Core::getQuery()->insert("message", 
			array("mode", "time", "sender", "receiver", "message", "subject", "readed", "related_user"), 
			array("4", time(), null, $this->userid, $this->espReport, Core::getLanguage()->getItem("ESP_REPORT_SUBJECT"), 0, $this->getData("userid")));
		return $this;
	}

	/**
	* Loads all needed data to calculate report data.
	*
	* @return EspionageReport
	*/
	protected function loadEspData()
	{
		// Load research
		$this->espTech = intval(sqlSelectField("research2user", "level", "", "userid = ".sqlVal($this->userid)." AND buildingid = ".UNIT_SPYWARE));

		$result = sqlSelect("research2user r2u", array("r2u.buildingid", "r2u.level", "b.name"), 
			"LEFT JOIN ".PREFIX."construction b ON (b.buildingid = r2u.buildingid)", 
			"r2u.userid = ".sqlVal($this->getData("userid")), 
			"b.display_order ASC, b.buildingid ASC");
		while($row = sqlFetch($result))
		{
			$this->research[$row["buildingid"]]["level"] = intval($row["level"]);
			$this->research[$row["buildingid"]]["name"] = Core::getLanguage()->getItem($row["name"]);
		}
		sqlEnd($result);
		// Get the blocks, wich will be viewed.
		if(!$this->research[UNIT_SPYWARE]["level"])
		{
			$tEspTech = 0;
		}
		else
		{
			$tEspTech = $this->research[UNIT_SPYWARE]["level"];
		}

		$result = sqlSelect(
			"unit2shipyard u2s",
			array("u2s.unitid", "u2s.quantity", "b.name"),
			"LEFT JOIN ".PREFIX."construction b ON (b.buildingid = u2s.unitid)",
			"u2s.planetid = ".sqlVal($this->planetid)
			. " AND b.mode = ".UNIT_TYPE_FLEET,
			"b.display_order ASC, b.buildingid ASC");
			while($row = sqlFetch($result))
			{
				$this->fleet[$row["unitid"]]["quantity"] = $row["quantity"];
				$this->fleet[$row["unitid"]]["name"] = Core::getLanguage()->getItem($row["name"]);
			}
			sqlEnd($result);
		$this->loadHoldingFleet();
		
		$units = 0;
		foreach($this->fleet as $unitid => $row)
		{
			if(!isset($this->skip_unit_keys[$unitid]))
			{
				$units += $row["quantity"];
			}
		}
		
		$result = sqlSelect(
			"unit2shipyard u2s",
			array("u2s.unitid", "u2s.quantity", "b.name"),
			"LEFT JOIN ".PREFIX."construction b ON (b.buildingid = u2s.unitid)",
			"u2s.planetid = ".sqlVal($this->planetid)
			. " AND b.mode = ".UNIT_TYPE_DEFENSE,
			"b.display_order ASC, b.buildingid ASC");
		while($row = sqlFetch($result))
		{
			$this->defense[$row["unitid"]]["quantity"] = $row["quantity"];
			$this->defense[$row["unitid"]]["name"] = Core::getLanguage()->getItem($row["name"]);
		}
		sqlEnd($result);
		
		foreach($this->defense as $unitid => $row)
		{
			if(!isset($this->skip_unit_keys[$unitid]))
			{
				$units += $row["quantity"];
			}
		}
	
		// Get espionage defense chance.
		$this->chance = round(50 * pow(max(0, 1 + ($tEspTech - $this->espTech) / 10), 5) * pow(max(1, $units), 0.02) / pow($this->probes, 0.3));
		$this->chance = max(0, $this->chance);

		if($units > 0 && mt_rand(0, 100) < $this->chance)
		{
			$this->probesLost = true;
		}
		$show_chance = max(0, 100 - $this->chance);
		$this->blocks = ceil($show_chance / 20);

		if($this->blocks > 2)
		{
			if($this->blocks > 3)
			{
				$result = sqlSelect(
					"building2planet b2p",
					array("b2p.buildingid", "b2p.level", "b.name"),
					"LEFT JOIN ".PREFIX."construction b ON (b.buildingid = b2p.buildingid)",
					"b2p.planetid = ".sqlVal($this->planetid)." AND b.mode in (".sqlArray(UNIT_TYPE_CONSTRUCTION, UNIT_TYPE_MOON_CONSTRUCTION).")",
					"b.display_order ASC, b.buildingid ASC");
				while($row = sqlFetch($result))
				{
					$this->building[$row["buildingid"]]["level"] = $row["level"];
					$this->building[$row["buildingid"]]["name"] = Core::getLanguage()->getItem($row["name"]);
				}
				sqlEnd($result);
			}
		}

		return $this;
	}

	/**
	* Generates the HTML code for the report.
	*
	* @return EspionageReport
	*/
	protected function generateReport()
	{
		$rHeadline = Core::getLanguage()->getItem("ESP_REPORT_RESSOURCE_HEADLINE");
		$rHeadline = sprintf($rHeadline, $this->getData("planetname"), $this->position, $this->targetName);
		$this->espReport	= "<table class=\"msgEspionage\">";
		$this->espReport .= "<tr><th colspan=\"4\">".$rHeadline."</th></tr>";
		$this->espReport .= "<tr><td>".Core::getLanguage()->getItem("ESP_REPORT_METAL")."</td><td>".fNumber($this->getData("metal"))."</td><td>".Core::getLanguage()->getItem("ESP_REPORT_SILICON")."</td><td>".fNumber($this->getData("silicon"))."</td></tr>";
		$this->espReport .= "<tr><td>".Core::getLanguage()->getItem("ESP_REPORT_HYDROGEN")."</td><td>".fNumber($this->getData("hydrogen"))."</td><td>".Core::getLanguage()->getItem("ESP_REPORT_ENERGY")."</td><td>".fNumber($this->getEnergy())."</td></tr>";
		if($this->blocks > 1 && (true || !$this->probesLost))
		{
			// Set fleet
			$this->espReport .= "<tr><th colspan=\"4\">".Core::getLanguage()->getItem("ESP_REPORT_FLEET")."</th></tr>";
			$i = 0;
			if(count($this->fleet) > 0)
			{
				foreach($this->fleet as $fleet)
				{
					if($i % 2 == 0) { $this->espReport .= "<tr>"; }
					$this->espReport .= "<td>".$fleet["name"]."</td><td>".$fleet["quantity"]."</td>";
					if(count($this->fleet) == $i + 1 && $i % 2 == 0)
					{
						$this->espReport .= "<td></td><td></td></tr>";
					}
					if($i % 2 == 1) { $this->espReport .= "</tr>"; }
					$i++;
				}
			}

			// Set defense
			if($this->blocks > 2)
			{
				$this->espReport .= "<tr><th colspan=\"4\">".Core::getLanguage()->getItem("ESP_REPORT_DEFENSE")."</th></tr>";
				$i = 0;
				if(count($this->defense) > 0)
				{
					foreach($this->defense as $def)
					{
						if($i % 2 == 0) { $this->espReport .= "<tr>"; }
						$this->espReport .= "<td>".$def["name"]."</td><td>".$def["quantity"]."</td>";
						if(count($this->defense) == $i + 1 && $i % 2 == 0)
						{
							$this->espReport .= "<td></td><td></td></tr>";
						}
						if($i % 2 == 1) { $this->espReport .= "</tr>"; }
						$i++;
					}
				}

				// Set buildings
				if($this->blocks > 3)
				{
					$this->espReport .= "<tr><th colspan=\"4\">".Core::getLanguage()->getItem("ESP_REPORT_BUILDINGS")."</th></tr>";
					$i = 0;
					if(count($this->building) > 0)
					{
						foreach($this->building as $b)
						{
							if($i % 2 == 0) { $this->espReport .= "<tr>"; }
							$this->espReport .= "<td>".$b["name"]."</td><td>".$b["level"]."</td>";
							if(count($this->building) == $i + 1 && $i % 2 == 0)
							{
								$this->espReport .= "<td></td><td></td></tr>";
							}
							if($i % 2 == 1) { $this->espReport .= "</tr>"; }
							$i++;
						}
					}

					// Set research
					if($this->blocks > 4)
					{
						$this->espReport .= "<tr><th colspan=\"4\">".Core::getLanguage()->getItem("ESP_REPORT_RESEARCH")."</th></tr>";
						$i = 0;
						if(count($this->research) > 0)
						{
							foreach($this->research as $r)
							{
								if($i % 2 == 0) { $this->espReport .= "<tr>"; }
								$this->espReport .= "<td>".$r["name"]."</td><td>".$r["level"]."</td>";
								if(count($this->research) == $i + 1 && $i % 2 == 0)
								{
									$this->espReport .= "<td></td><td></td></tr>";
								}
								if($i % 2 == 1) { $this->espReport .= "</tr>"; }
								$i++;
							}
						}
					}
				}
			}
		}
		$this->espReport .= "<tr><td colspan=\"4\" style=\"text-align: center;\">".sprintf(Core::getLanguage()->getItem("ESP_DEFENDING_CHANCE"), round($this->chance))."</td></tr>";
		if($this->probesLost)
		{
			$this->espReport .= "<tr><td colspan=\"4\" style=\"text-align: center; font-weight: bold;\" class=\"notavailable\">".Core::getLanguage()->getItem("ESP_PROBES_DESTROYED")."</td></tr>";
		}
		$this->espReport .= "<tr><td colspan=\"4\" style=\"text-align: center;\">".
			Link::get("game.php/go:Mission/g:".$this->getData("galaxy")."/s:".$this->getData("system")."/p:".$this->getData("position"), Core::getLanguage()->getItem("ATTACK"))."</td></tr>";
		$this->espReport .= "</table>";
		// Hook::event("ESPIONAGE_REPORT_GENERATOR", array(&$this));
		return $this;
	}

	/**
	* Loads the holding fleets.
	*
	* @return EspionageReport
	*/
	protected function loadHoldingFleet()
	{
		$result = sqlSelect("events", array("data"), "",
			"mode IN (".sqlArray(EVENT_HOLDING, EVENT_ALIEN_HOLDING).") AND destination = ".sqlVal($this->planetid)." AND processed=".EVENT_PROCESSED_WAIT);
		while($row = sqlFetch($result))
		{
			$row["data"] = unserialize($row["data"]);
			foreach($row["data"]["ships"] as $ship)
			{
				if(isset($this->fleet[$ship["id"]]))
				{
					$this->fleet[$ship["id"]]["quantity"] += $ship["quantity"];
				}
				else
				{
					$this->fleet[$ship["id"]]["name"] = Core::getLanguage()->getItem($ship["name"]);
					$this->fleet[$ship["id"]]["quantity"] = $ship["quantity"];
				}
			}
		}
		sqlEnd($result);
		return $this;
	}

	/**
	* Returns the name of the planet.
	*
	* @return string
	*/
	public function getPlanetname()
	{
		return $this->getData("planetname");
	}

	/**
	* Returns the chance to be discovered.
	*
	* @return integer
	*/
	public function getChance()
	{
		return $this->chance;
	}

	/**
	* Returns the lost probes.
	*
	* @return boolean
	*/
	public function getProbesLost()
	{
		return $this->probesLost;
	}

	/**
	* Returns the target user id.
	*
	* @return integer
	*/
	public function getTargetUserId()
	{
		return $this->getData("userid");
	}
}
?>
