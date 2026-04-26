<?php
/**
* Shows infos about an unit.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class UnitInfo extends Page
{
	/**
	* Constructor: Shows informations about an unit.
	*
	* @return void
	*/
	public function __construct()
	{
		parent::__construct();
//		try {
			$this->setGetAction("go", "UnitInfo", "showInfo")
				->addGetArg("showInfo", "id")
				->proceedRequest();
//		} catch(Exception $e) {
//			$e->printError();
//		}
		return;
	}

	/**
	* Index action.
	*
	* @return UnitInfo
	*/
	protected function index()
	{
		$this->showInfo(Core::getRequest()->getGET("id"));
		return $this;
	}

	/**
	* Shows all unit information.
	*
	* @param integer	Unit id
	*
	* @return UnitInfo
	*/
	protected function showInfo($id)
	{
		// Common unit data
		$select = array("c.buildingid as unitid", "c.name", "c.mode",
			"c.basic_metal", "c.basic_silicon", "c.basic_hydrogen",
			"ds.capicity", "ds.speed", "ds.consume",
			"ds.attack", "ds.shield", "ds.front", "ds.ballistics", "ds.masking",
			"ds.attacker_attack", "ds.attacker_shield", "ds.attacker_front", "ds.attacker_ballistics", "ds.attacker_masking",
			);
		$join = "LEFT JOIN ".PREFIX."ship_datasheet ds ON (ds.unitid = c.buildingid)";
		$result = sqlSelect("construction c", $select, $join, "c.buildingid = ".sqlVal($id));
		if($row = sqlFetch($result))
		{
			sqlEnd($result);
			Core::getLanguage()->load("info,UnitInfo");
			// Hook::event("SHOW_UNIT_INFO", array(&$row));
			Core::getTPL()->assign("mode", $row["mode"]);
			Core::getTPL()->assign("productiontime", getTimeTerm(NS::getBuildingTime($row["basic_metal"], $row["basic_silicon"], $row["mode"])));
			Core::getTPL()->assign("basic_metal", fNumber($row["basic_metal"]));
			Core::getTPL()->assign("basic_silicon", fNumber($row["basic_silicon"]));
			Core::getTPL()->assign("basic_hydrogen", fNumber($row["basic_hydrogen"]));
			Core::getTPL()->assign("structure", fNumber($row["basic_metal"] + $row["basic_silicon"]));

			Core::getTPL()->assign("shield", fNumber($row["shield"]));
			Core::getTPL()->assign("attack", fNumber($row["attack"]));
			Core::getTPL()->assign("shell", fNumber(($row["basic_metal"] + $row["basic_silicon"]) / 10));
			Core::getTPL()->assign("front", fNumber($row["front"]));
			Core::getTPL()->assign("battle_weight", fNumber(pow(2, $row["front"])));
			Core::getTPL()->assign("ballistics", fNumber($row["ballistics"]));
			Core::getTPL()->assign("masking", fNumber($row["masking"]));

			Core::getTPL()->assign("attacker_shield", fNumber($row["attacker_shield"]));
			Core::getTPL()->assign("attacker_attack", fNumber($row["attacker_attack"]));
			Core::getTPL()->assign("attacker_front", fNumber($row["attacker_front"]));
			Core::getTPL()->assign("attacker_battle_weight", fNumber(pow(2, $row["attacker_front"])));
			Core::getTPL()->assign("attacker_ballistics", fNumber($row["attacker_ballistics"]));
			Core::getTPL()->assign("attacker_masking", fNumber($row["attacker_masking"]));

			Core::getTPL()->assign("capacity", fNumber($row["capicity"]));
			Core::getTPL()->assign("speed", fNumber($row["speed"]));
			Core::getTPL()->assign("cur_speed", fNumber(NS::getSpeed($row["unitid"], $row["speed"])));
			Core::getTPL()->assign("base_consume", fNumber( ceil( $row["consume"] * FLEET_FUEL_CONSUMPTION ) ));
			Core::getTPL()->assign("consume", fNumber( ceil( $row["consume"] * FLEET_FUEL_CONSUMPTION * NS::getGraviFuelConsumeScale() ) ));
			Core::getTPL()->assign("gravi_name_link", Link::get("game.php/ResearchInfo/".UNIT_GRAVI, Core::getLanguage()->getItem("GRAVI")." ".NS::getResearch(UNIT_GRAVI)));
			Core::getTPL()->assign("name", Core::getLanguage()->getItem($row["name"]));
			Core::getTPL()->assign("description", Core::getLanguage()->getItem($row["name"]."_FULL_DESC"));
			Core::getTPL()->assign("pic", Image::getImage(getUnitImage($row["name"]), Core::getLanguage()->getItem($row["name"]), null, null, "leftImage"));
			Core::getTPL()->assign("edit", Link::get("game.php/EditUnit/".$id, "[".Core::getLanguage()->getItem("EDIT")."]"));

			$engines = array();
			$speed_factor = NS::getGalaxyParam(null, "FLEET_SPEED_FACTOR", 1);
			$unit_engines = NS::getUnitEngines($row["unitid"], $row["speed"]);
			foreach($unit_engines as $engine)
			{
				$classname = "";
				$value = "";
				if($engine["engine_name"])
				{
					$value = Core::getLanguage()->getItem($engine["engine_name"]);
					if($engine["engine_level"] > $engine["tech_level"])
					{
						$classname = "notavailable";
						$value .= " ".$engine["engine_level"];
						$value .= " (+".($engine["engine_level"] - $engine["tech_level"]).")";
					}
					else
					{
						$classname = $engine["active"] && sizeof($unit_engines) > 1 ? "available" : "";
						$value .= " ".$engine["tech_level"];
					}
					$value .= " : ";
				}
				$value .= fNumber($engine["speed"] * $speed_factor);
				if($engine["speed"] != $engine["base_speed"] && $engine["active"])
				{
					$value .= " ( ".fNumber($engine["base_speed"] * $speed_factor)." )";
				}
				$value = "<span class='{$classname}'><nobr>{$value}</nobr></span>";
				$engines[] = array("value" => $value);
			}
			// debug_var($engines, "[engines]");
			Core::getTPL()->addLoop("engines", $engines);

			// Rapidfire
			$i = 0;
			$_result = sqlSelect("rapidfire rf", array("rf.target", "rf.value", "c.name"), "LEFT JOIN ".PREFIX."construction c ON (c.buildingid = rf.target)", "rf.unitid = ".sqlVal($id));
			while($_row = sqlFetch($_result))
			{
				// Hook::event("SHOW_UNIT_RAPIDFIRE", array($row, &$_row));
				$name = Link::get("game.php/UnitInfo/".$_row["target"], Core::getLanguage()->getItem($_row["name"]));
				$rf[$i]["rapidfire"] = sprintf(Core::getLanguage()->getItem("RAPIDFIRE_TO"), $name);
				$rf[$i]["value"] = "<span class=\"available\">".fNumber($_row["value"])."</span>";
				$i++;
			}
			sqlEnd($_result);

			$_result = sqlSelect("rapidfire rf", array("rf.unitid", "rf.value", "c.name"), "LEFT JOIN ".PREFIX."construction c ON (c.buildingid = rf.unitid)", "rf.target = ".sqlVal($id));
			while($_row = sqlFetch($_result))
			{
				// Hook::event("SHOW_UNIT_RAPIDFIRE", array($row, &$_row));
				$name = Link::get("game.php/UnitInfo/".$_row["unitid"], Core::getLanguage()->getItem($_row["name"]));
				$rf[$i]["rapidfire"] = sprintf(Core::getLanguage()->getItem("RAPIDFIRE_FROM"), $name);
				$rf[$i]["value"] = "<span class=\"notavailable\">".fNumber($_row["value"])."</span>";
				$i++;
			}
			sqlEnd($_result);
			Core::getTPL()->addLoop("rapidfire", $rf);

			Core::getTPL()->display("unitinfo");
		}
		else
		{
			sqlEnd($result);
			throw new GenericException("Unkown unit. You'd better don't mess with the URL.");
		}
		return $this;
	}
}
?>