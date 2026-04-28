<?php
/**
* Administrator interface to modify unit data.
*
* Oxsar http://oxsar.ru
*
* 
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class EditUnit extends Page
{
	/**
	* Holds a list of all available engines.
	*
	* @var array
	*/
	protected $engines = array();

	/**
	* The Id of the unit.
	*
	* @var integer
	*/
	protected $id = 0;

	/**
	* Constructor: Handles form and post actions.
	*
	* @return void
	*/
	public function __construct()
	{
		NS::getUser()->checkPermissions("CAN_EDIT_CONSTRUCTIONS");
		Core::getLanguage()->load(array("Administrator", "UnitInfo", "buildings", "info"));

		$this->id = Core::getRequest()->getGET("id");

		parent::__construct();
		$this->setPostAction("saveunit", "saveConstruction")
			->addPostArg("saveConstruction", "unitid")
			->addPostArg("saveConstruction", "name_id")
			->addPostArg("saveConstruction", "unit_name")
			->addPostArg("saveConstruction", "desc")
			->addPostArg("saveConstruction", "full_desc")
			->addPostArg("saveConstruction", "basic_metal")
			->addPostArg("saveConstruction", "basic_silicon")
			->addPostArg("saveConstruction", "basic_hydrogen")
			->addPostArg("saveConstruction", "basic_energy")
			->addPostArg("saveConstruction", "capicity")
			->addPostArg("saveConstruction", "speed")
			->addPostArg("saveConstruction", "consume")
			->addPostArg("saveConstruction", "attack")
			->addPostArg("saveConstruction", "shield")
			->addPostArg("saveConstruction", "baseEngine")
			->addPostArg("saveConstruction", "extentedEngine")
			->addPostArg("saveConstruction", "extentedEngineLevel")
			->addPostArg("saveConstruction", "extentedEngineSpeed")
			->addPostArg("saveConstruction", "del_rf")
			->addPostArg("saveConstruction", "rf_new")
			->addPostArg("saveConstruction", "rf_new_value")
			->setPostAction("addreq", "addRequirement")
			->addPostArg("addRequirement", "level")
			->addPostArg("addRequirement", "needs")
			->setGetAction("do", "delete", "deleteRequirement")
			->addGetArg("deleteRequirement", "id")
			->proceedRequest();

		return;
	}

	/**
	* Adds a new requirement.
	*
	* @param integer	Level
	* @param integer	Required construction
	*
	* @return EditUnit
	*/
	protected function addRequirement($level, $needs)
	{
		if(!is_numeric($level) || $level < 0) { $level = 1; }
		Core::getQuery()->insert("requirements", array("buildingid", "needs", "level"), array(Core::getRequest()->getGET("delete"), $needs, $level));
		Core::getCache()->flushObject("requirements");
		return $this;
	}

	/**
	* Shows the editing form.
	*
	* @return EditUnit
	*/
	protected function index()
	{
		$languageid = Core::getLang()->getOpt("languageid");
		$select = array("c.name AS name_id", "p.content AS name", "c.basic_metal", "c.basic_silicon", "c.basic_hydrogen", "c.basic_energy", "sds.unitid", "sds.capicity AS capacity", "sds.speed", "sds.consume", "sds.attack", "sds.shield");
		$joins	= "LEFT JOIN ".PREFIX."phrases p ON (p.title = c.name)";
		$joins .= "LEFT JOIN ".PREFIX."ship_datasheet sds ON (sds.unitid = c.buildingid)";
		$result = sqlSelect("construction c", $select, $joins, "c.buildingid = ".sqlVal($this->id)." AND p.languageid = ".sqlVal($languageid));
		if($row = sqlFetch($result))
		{
			sqlEnd($result);
			// Hook::event("EDIT_UNIT_DATA_LOADED", array(&$row));
			foreach($row as $key => $value)
			{
				Core::getTPL()->assign($key, $value);
			}
			Core::getTPL()->assign("shell", fNumber(($row["basic_metal"] + $row["basic_silicon"]) / 10));
			$result = sqlSelect("phrases", "content", "", "languageid = ".sqlVal($languageid)." AND title = ".sqlVal($row["name_id"]."_DESC"));
			$_row = sqlFetch($result);
			sqlEnd($result);
			Core::getTPL()->assign("description", $_row["content"]);
			$result = sqlSelect("phrases", "content", "", "languageid = ".sqlVal($languageid)." AND title = ".sqlVal($row["name_id"]."_FULL_DESC"));
			$_row = sqlFetch($result);
			sqlEnd($result);
			Core::getTPL()->assign("full_description", $_row["content"]);

			// Engine selection
			Core::getTPL()->assign("extentedEngine", $this->getEnginesList(0));
			Core::getTPL()->assign("extentedEngineLevel", 0);
			Core::getTPL()->assign("extentedEngineSpeed", "");
			$engines = array();
			$result = sqlSelect("ship2engine s2e", array("s2e.engineid", "s2e.level", "s2e.base_speed", "s2e.base"), "", "s2e.unitid = ".sqlVal($this->id));
			while($_row = sqlFetch($result))
			{
				if($_row["base"] == 1)
				{
					Core::getTPL()->assign("baseEngine", $this->getEnginesList($_row["engineid"]));
				}
				else
				{
					Core::getTPL()->assign("extentedEngine", $this->getEnginesList($_row["engineid"]));
					Core::getTPL()->assign("extentedEngineLevel", $_row["level"]);
					Core::getTPL()->assign("extentedEngineSpeed", $_row["base_speed"]);
				}
			}

			$req = array(); $i = 0;
			$result = sqlSelect("requirements r", array("r.requirementid", "r.needs", "r.level", "p.content"), "LEFT JOIN ".PREFIX."construction b ON (b.buildingid = r.needs) LEFT JOIN ".PREFIX."phrases p ON (p.title = b.name)", "r.buildingid = ".sqlVal($this->id)." AND p.languageid = ".sqlVal($languageid));
			while($row = sqlFetch($result))
			{
				$req[$i]["delete"] = Link::get("game.php/sid:".SID."/go:EditUnit/do:delete/id:".$row["requirementid"], "[".Core::getLanguage()->getItem("DELETE")."]");
				$req[$i]["name"] = Link::get("game.php/EditUnit/".$row["needs"], $row["content"]);
				$req[$i]["level"] = $row["level"];
				$i++;
			}
			sqlEnd($result);
			Core::getTPL()->addLoop("requirements", $req);

			$const = array(); $i = 0;
			$result = sqlSelect("construction b", array("b.buildingid", "p.content"), "LEFT JOIN ".PREFIX."phrases p ON (p.title = b.name)", "(b.mode = ".UNIT_TYPE_CONSTRUCTION." OR b.mode = ".UNIT_TYPE_RESEARCH." OR b.mode = ".UNIT_TYPE_MOON_CONSTRUCTION.") AND p.languageid = ".sqlVal($languageid), "p.content ASC");
			while($row = sqlFetch($result))
			{
				$const[$i]["name"] = $row["content"];
				$const[$i]["id"] = $row["buildingid"];
				$i++;
			}
			sqlEnd($result);
			Core::getTPL()->addLoop("constructions", $const);
			Core::getTPL()->addLoop("rapidfire", $this->getRapidFire());
			Core::getTPL()->assign("rfSelect", $this->getShipSelect());
			Core::getTPL()->display("edit_unit");
		}
		return $this;
	}

	/**
	* Deletes requirements.
	*
	* @param integer	Requirement id to delete
	*
	* @return EditUnit
	*/
	protected function deleteRequirement($delete)
	{
		Core::getQuery()->delete("requirements", "requirementid = ".sqlVal($delete));
		Core::getCache()->flushObject("requirements");
		return $this->index();
	}

	/**
	* Saves the entered data.
	*
	* @param integer	Unit id
	* @param string	Name id
	* @param string	Name
	* @param string	Description
	* @param string	Long description
	* @param integer	Basic metal cost
	* @param integer	Basic silicon cost
	* @param integer	Basic hydrogen cost
	* @param integer	Basic energy cost
	* @param integer	Capacity
	* @param integer	Basic speed
	* @param integer	Fuel consumption
	* @param integer	Attack power
	* @param integer	Shield power
	* @param integer	Base engine id
	* @param integer	Extented engine id
	* @param integer	Extented engine from level
	* @param integer	Extented engine speed
	*
	* @return EditUnit
	*/
	protected function saveConstruction(
		$unitid, $nameId, $name, $desc, $fullDesc,
		$basicMetal, $basicSilicon, $basicHydrogen, $basicEnergy,
		$capacity, $speed, $consumption, $attack, $shield,
		$baseEngine, $extentedEngine, $extentedEngineLevel, $extentedEngineSpeed,
		$rfDelete, $rfNew, $rfNewValue
		)
	{
		// Hook::event("EDIT_UNIT_SAVE");
		$languageid = Core::getLang()->getOpt("languageid");
		$atts = array("basic_metal", "basic_silicon", "basic_hydrogen", "basic_energy");
		$vals = array($basicMetal, $basicSilicon, $basicHydrogen, $basicEnergy);
		Core::getQuery()->update("construction", $atts, $vals, "name = ".sqlVal($nameId));
		$atts = array("capicity", "speed", "consume", "attack", "shield");
		$vals = array($capacity, $speed, $consumption, $attack, $shield);
		Core::getQuery()->update("ship_datasheet", $atts, $vals, "unitid = ".sqlVal($unitid));
		Core::getQuery()->update("ship2engine", array("engineid"), array($baseEngine), "unitid = ".sqlVal($unitid)." AND base = '1'");
		Core::getQuery()->delete("ship2engine", "unitid = ".sqlVal($unitid)." AND base = '0'");
		if($extentedEngineLevel > 0)
		{
			Core::getQuery()->insert("ship2engine", array("engineid", "unitid", "level", "base_speed", "base"), array($extentedEngine, $unitid, $extentedEngineLevel, $extentedEngineSpeed, 0));
		}

		if(Str::length($name) > 0)
		{
			$result_count = sqlSelectField("phrases", "count(*)", "", "title = ".sqlVal($nameId));
			if($result_count > 0)
			{
				Core::getQuery()->update("phrases", array("content"), array(convertSpecialChars($name)), "title = ".sqlVal($nameId));
			}
			else
			{
				Core::getQuery()->insert("phrases", array("languageid", "phrasegroupid", "title", "content"), array($languageid, 4, $nameId, convertSpecialChars($name)));
			}
		}
		if(Str::length($desc) > 0)
		{
			$result_count = sqlSelectField("phrases", "count(*)", "", "title = ".sqlVal($nameId."_DESC"));
			if($result_count > 0)
			{
				Core::getQuery()->update("phrases", array("content"), array(convertSpecialChars($desc)), "title = ".sqlVal($nameId."_DESC"));
			}
			else
			{
				Core::getQuery()->insert("phrases", array("languageid", "phrasegroupid", "title", "content"), array($languageid, 4, $nameId."_DESC", convertSpecialChars($desc)));
			}
		}
		if(Str::length($fullDesc) > 0)
		{
			$result_count = sqlSelectField("phrases", "count(*)", "", "title = ".sqlVal($nameId."_FULL_DESC"));
			if($result_count > 0)
			{
				Core::getQuery()->update("phrases", array("content"), array(convertSpecialChars($fullDesc)), "title = ".sqlVal($nameId."_FULL_DESC"));
			}
			else
			{
				Core::getQuery()->insert("phrases", array("languageid", "phrasegroupid", "title", "content"), array($languageid, 4, $nameId."_FULL_DESC", convertSpecialChars($fullDesc)));
			}
		}
		Core::getLang()->rebuild("info");

		// Rapidfire
		$result = sqlSelect("rapidfire", array("target", "value"), "", "unitid = ".sqlVal($this->id));
		while($row = sqlFetch($result))
		{
			if(is_array($rfDelete) && in_array($row["target"], $rfDelete))
			{
				Core::getQuery()->delete("rapidfire", "unitid = ".sqlVal($this->id)." AND target = ".sqlVal($row["target"]));
			}
			else if(Core::getRequest()->getPOST("rf_".$row["target"]) != $row["value"])
			{
				Core::getQuery()->update("rapidfire", array("value"), array(Core::getRequest()->getPOST("rf_".$row["target"])), "unitid = ".sqlVal($this->id)." AND target = ".sqlVal($row["target"]));
			}
		}
		if($rfNew > 0 && $rfNewValue > 0)
		{
			Core::getQuery()->delete("rapidfire", "unitid = ".sqlVal($this->id)." AND target = ".sqlVal($rfNew));
			Core::getQuery()->insert("rapidfire", array("unitid", "target", "value"), array($this->id, $rfNew, $rfNewValue));
		}

		return $this->index();
	}

	/**
	* Generates the HTML for the engine list.
	*
	* @param integer	Selected engine
	*
	* @return string	HTML code (only option-tags)
	*/
	protected function getEnginesList($engineid)
	{
		if(count($this->engines) <= 0)
		{
			$joins	= "LEFT JOIN ".PREFIX."construction c ON (c.buildingid = e.engineid)";
			$joins .= "LEFT JOIN ".PREFIX."phrases p ON (p.title = c.name)";
			$result = sqlSelect("engine e", array("e.engineid", "p.content AS name"), $joins, "p.languageid = ".sqlVal(Core::getLang()->getOpt("languageid")), "c.display_order ASC, c.buildingid ASC");
			while($row = sqlFetch($result))
			{
				$this->engines[] = $row;
			}
		}
		$select = "";
		foreach($this->engines as $engine)
		{
			if($engine["engineid"] == $engineid) { $s = 1; }
			else { $s = 0; }
			$select .= createOption($engine["engineid"], $engine["name"], $s);
		}
		return $select;
	}

	/**
	* Returns the rapidfire of the unit.
	*
	* @return array
	*/
	protected function getRapidFire()
	{
		$rf = array();
		$sel = array("r.target", "r.value", "p.content AS name");
		$joins	= "LEFT JOIN ".PREFIX."construction b ON (b.buildingid = r.target) ";
		$joins .= "LEFT JOIN ".PREFIX."phrases p ON (b.name = p.title)";
		$result = sqlSelect("rapidfire r", $sel, $joins, "r.unitid = ".sqlVal($this->id)." AND p.languageid = ".sqlVal(NS::getUser()->get("languageid")));
		while($row = sqlFetch($result))
		{
			$rf[] = $row;
		}
		return $rf;
	}

	/**
	* Fetches all ships returns them as an option list.
	*
	* @param integer	Pre-selected ship
	*
	* @return string	Options
	*/
	protected function getShipSelect($unit = 0)
	{
		$ret = "";
		$sel = array("b.buildingid", "p.content AS name");
		$join = "LEFT JOIN ".PREFIX."phrases p ON (b.name = p.title)";
		$result = sqlSelect("construction b", $sel, $join, "(b.mode = ".UNIT_TYPE_FLEET." OR b.mode = ".UNIT_TYPE_DEFENSE.") AND p.languageid = ".sqlVal(NS::getUser()->get("languageid")));
		while($row = sqlFetch($result))
		{
			$ret .= createOption($row["buildingid"], $row["name"], ($row["buildingid"] == $unit) ? 1 : 0);
		}
		return $ret;
	}
}
?>