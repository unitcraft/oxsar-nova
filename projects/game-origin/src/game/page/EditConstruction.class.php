<?php
/**
* Administrator interface to modify construction data.
*
* Oxsar http://oxsar.ru
*
* 
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class EditConstruction extends Page
{
	/**
	* List of resources.
	*
	* @var array
	*/
	protected $resources = array(
		"metal"		=> "METAL",
		"silicon"	=> "SILICON",
		"hydrogen"	=> "HYDROGEN",
		"energy"	=> "ENERGY"
		);

	/**
	* Shows edit form.
	*
	* @return void
	*/
	public function __construct()
	{
		NS::getUser()->checkPermissions("CAN_EDIT_CONSTRUCTIONS");
		Core::getLanguage()->load("Administrator");
		parent::__construct();

		$this->setPostAction("saveconstruction", "saveConstruction")
			->addPostArg("saveConstruction", "name")
			->addPostArg("saveConstruction", "name_id")
			->addPostArg("saveConstruction", "desc")
			->addPostArg("saveConstruction", "full_desc")
			->addPostArg("saveConstruction", "prod_what")
			->addPostArg("saveConstruction", "prod")
			->addPostArg("saveConstruction", "cons_what")
			->addPostArg("saveConstruction", "consumption")
			->addPostArg("saveConstruction", "special")
			->addPostArg("saveConstruction", "basic_metal")
			->addPostArg("saveConstruction", "basic_silicon")
			->addPostArg("saveConstruction", "basic_hydrogen")
			->addPostArg("saveConstruction", "basic_energy")
			->addPostArg("saveConstruction", "charge_metal")
			->addPostArg("saveConstruction", "charge_silicon")
			->addPostArg("saveConstruction", "charge_hydrogen")
			->addPostArg("saveConstruction", "charge_energy")
			->setPostAction("addreq", "addRequirement")
			->addPostArg("addRequirement", "level")
			->addPostArg("addRequirement", "needs")
			->setGetAction("do", "delete", "deleteRequirement")
			->addGetArg("deleteRequirement", "id")
			->proceedRequest();

		return;
	}

	/**
	* Index action.
	*
	* @return EditConstruction
	*/
	protected function index()
	{
		$id = Core::getRequest()->getGET("id");
		$select = array(
			"c.name AS name_id", "p.content AS name", "c.special",
			"c.basic_metal", "c.basic_silicon", "c.basic_hydrogen", "c.basic_energy",
			"c.prod_metal", "c.prod_silicon", "c.prod_hydrogen", "c.prod_energy",
			"c.cons_metal", "c.cons_silicon", "c.cons_hydrogen", "c.cons_energy",
			"c.charge_metal", "c.charge_silicon", "c.charge_hydrogen", "c.charge_energy"
			);
		$joins	= "LEFT JOIN ".PREFIX."phrases p ON (p.title = c.name)";
		$result = sqlSelect("construction c", $select, $joins, "c.buildingid = ".sqlVal($id)." AND p.languageid = ".sqlVal(Core::getLanguage()->getOpt("languageid")));
		if($row = sqlFetch($result))
		{
			sqlEnd($result);
			// Hook::event("EDIT_UNIT_DATA_LOADED", array(&$row));

			// Set production
			if(!empty($row["prod_metal"]))
			{
				$row["prod"] = $row["prod_metal"];
				$prodWhat = "metal";
			}
			else if(!empty($row["prod_silicon"]))
			{
				$row["prod"] = $row["prod_silicon"];
				$prodWhat = "silicon";
			}
			else if(!empty($row["prod_hydrogen"]))
			{
				$row["prod"] = $row["prod_hydrogen"];
				$prodWhat = "hydrogen";
			}
			else if(!empty($row["prod_energy"]))
			{
				$row["prod"] = $row["prod_energy"];
				$prodWhat = "energy";
			}

			// Set Consumption
			if(!empty($row["cons_metal"]))
			{
				$row["consumption"] = $row["cons_metal"];
				$consWhat = "metal";
			}
			else if(!empty($row["cons_silicon"]))
			{
				$row["consumption"] = $row["cons_silicon"];
				$consWhat = "silicon";
			}
			else if(!empty($row["cons_hydrogen"]))
			{
				$row["consumption"] = $row["cons_hydrogen"];
				$consWhat = "hydrogen";
			}
			else if(!empty($row["cons_energy"]))
			{
				$row["consumption"] = $row["cons_energy"];
				$consWhat = "energy";
			}

			Core::getTPL()->assign("prodWhat", $this->getResourceSelect($prodWhat));
			Core::getTPL()->assign("consWhat", $this->getResourceSelect($consWhat));

			foreach($row as $key => $value)
			{
				Core::getTPL()->assign($key, $value);
			}

			$result = sqlSelect("phrases", "content", "", "languageid = ".sqlVal(Core::getLanguage()->getOpt("languageid"))." AND title = ".sqlVal($row["name_id"]."_DESC"));
			$_row = sqlFetch($result);
			sqlEnd($result);
			Core::getTPL()->assign("description", $_row["content"]);
			$result = sqlSelect("phrases", "content", "", "languageid = ".sqlVal(Core::getLanguage()->getOpt("languageid"))." AND title = ".sqlVal($row["name_id"]."_FULL_DESC"));
			$_row = sqlFetch($result);
			sqlEnd($result);
			Core::getTPL()->assign("full_description", $_row["content"]);

			$req = array(); $i = 0;
			$result = sqlSelect("requirements r", array("r.requirementid", "r.needs", "r.level", "p.content"), "LEFT JOIN ".PREFIX."construction b ON (b.buildingid = r.needs) LEFT JOIN ".PREFIX."phrases p ON (p.title = b.name)", "r.buildingid = ".sqlVal($id)." AND p.languageid = ".sqlVal(Core::getLanguage()->getOpt("languageid")));
			while($row = sqlFetch($result))
			{
				$req[$i]["delete"] = Link::get("game.php/sid:".SID."/go:EditUnit/do:delete/id:".$row["requirementid"], "[".Core::getLanguage()->getItem("DELETE")."]");
				$req[$i]["name"] = Link::get("game.php/EditUnit/".$row["needs"], $row["content"]);
				$req[$i]["level"] = $row["level"];
				$i++;
			}
			Core::getTPL()->addLoop("requirements", $req);

			$const = array(); $i = 0;
			$result = sqlSelect("construction b", array("b.buildingid", "p.content"), "LEFT JOIN ".PREFIX."phrases p ON (p.title = b.name)", "(b.mode = ".UNIT_TYPE_CONSTRUCTION." OR b.mode = ".UNIT_TYPE_RESEARCH." OR b.mode = ".UNIT_TYPE_MOON_CONSTRUCTION.") AND p.languageid = ".sqlVal(Core::getLanguage()->getOpt("languageid")), "p.content ASC");
			while($row = sqlFetch($result))
			{
				$const[$i]["name"] = $row["content"];
				$const[$i]["id"] = $row["buildingid"];
				$i++;
			}
			sqlEnd($result);
			Core::getTPL()->addLoop("constructions", $const);
			Core::getTPL()->display("edit_construction");
		}
		return $this;
	}

	/**
	* Adds Requirements for a construction.
	*
	* @param integer	Level
	* @param integer	Required construction
	*
	* @return EditConstruction
	*/
	protected function addRequirement($level, $needs)
	{
		if(!is_numeric($level) || $level < 0) { $level = 1; }
		Core::getQuery()->insert("requirements", array("buildingid", "needs", "level"), array(Core::getRequest()->getGET("id"), $needs, $level));
		Core::getCache()->flushObject("requirements");
		return $this;
	}

	/**
	* Deletes the stated requirement.
	*
	* @param integer	Requirement id
	*
	* @return EditConstruction
	*/
	protected function deleteRequirement($delete)
	{
		Core::getQuery()->delete("requirements", "requirementid = ".sqlVal($delete));
		Core::getCache()->flushObject("requirements");
		return $this->index();
	}

	/**
	* Saves the construction data.
	*
	* @param string	Name
	* @param string	Name id
	* @param string	Description
	* @param string	Full description
	* @param string	Production resource
	* @param string	Production formula
	* @param string	Consumption resource
	* @param string	Consumption formula
	* @param string	Special formula
	* @param string	Basic metal cost
	* @param string	Basic silicon cost
	* @param string	Basic hydrogen cost
	* @param string	Basic energy cost
	* @param string	Metal cost increase
	* @param string	Silicon cost increase
	* @param string	Hydrogen cost increase
	* @param string	Energy cost increase
	*
	* @return EditConstruction
	*/
	protected function saveConstruction(
		$name, $nameId, $desc, $fullDesc,
		$prodWhat, $prod, $consWhat, $consumption, $special,
		$basicMetal, $basicSilicon, $basicHydrogen, $basicEnergy,
		$chargeMetal, $chargeSilicon, $chargeHydrogen, $chargeEnergy
		)
	{
		// Hook::event("EDIT_UNIT_SAVE");

		// Fetch production from form
		$prodMetal = ""; $prodSilicon = ""; $prodHydrogen = ""; $prodEnergy = "";
		if($prodWhat == "metal")
		{
			$prodMetal = $prod;
		}
		else if($prodWhat == "silicon")
		{
			$prodSilicon = $prod;
		}
		else if($prodWhat == "hydrogen")
		{
			$prodHydrogen = $prod;
		}
		else if($prodWhat == "energy")
		{
			$prodEnergy = $prod;
		}

		// Fetch consumption from form
		$consMetal = ""; $consSilicon = ""; $consHydrogen = ""; $consEnergy = "";
		if($consWhat == "metal")
		{
			$consMetal = $consumption;
		}
		else if($consWhat == "silicon")
		{
			$consSilicon = $consumption;
		}
		else if($consWhat == "hydrogen")
		{
			$consHydrogen = $consumption;
		}
		else if($consWhat == "energy")
		{
			$consEnergy = $consumption;
		}

		// Now generate the sql query.
		$atts = array("special",
			"basic_metal", "basic_silicon", "basic_hydrogen", "basic_energy",
			"prod_metal", "prod_silicon", "prod_hydrogen", "prod_energy",
			"cons_metal", "cons_silicon", "cons_hydrogen", "cons_energy",
			"charge_metal", "charge_silicon", "charge_hydrogen", "charge_energy"
			);
		$vals = array($special,
			$basicMetal, $basicSilicon, $basicHydrogen, $basicEnergy,
			$prodMetal, $prodSilicon, $prodHydrogen, $prodEnergy,
			$consMetal, $consSilicon, $consHydrogen, $consEnergy,
			$chargeMetal, $chargeSilicon, $chargeHydrogen, $chargeEnergy
			);
		Core::getQuery()->update("construction", $atts, $vals, "name = ".sqlVal($nameId));

		// Save the name and description
		$languageId = Core::getLang()->getOpt("languageid");
		if(Str::length($name) > 0)
		{
			$result = sqlSelect("phrases", "phraseid", "", "title = ".sqlVal($nameId));
			if(Core::getDB()->num_rows($result) > 0)
			{
				Core::getQuery()->update("phrases", array("content"), array(convertSpecialChars($name)), "title = ".sqlVal($nameId));
			}
			else
			{
				Core::getQuery()->insert("phrases", array("languageid", "phrasegroupid", "title", "content"), array($languageId, 4, $nameId, convertSpecialChars($name)));
			}
			sqlEnd($result);
		}
		if(Str::length($desc) > 0)
		{
			$result = sqlSelect("phrases", "phraseid", "", "title = ".sqlVal($nameId."_DESC"));
			if(Core::getDB()->num_rows($result) > 0)
			{
				Core::getQuery()->update("phrases", array("content"), array(convertSpecialChars($desc)), "title = ".sqlVal($nameId."_DESC"));
			}
			else
			{
				Core::getQuery()->insert("phrases", array("languageid", "phrasegroupid", "title", "content"), array($languageId, 4, $nameId."_DESC", convertSpecialChars($desc)));
			}
			sqlEnd($result);
		}
		if(Str::length($fullDesc) > 0)
		{
			$result = sqlSelect("phrases", "phraseid", "", "title = ".sqlVal($nameId."_FULL_DESC"));
			if(Core::getDB()->num_rows($result) > 0)
			{
				Core::getQuery()->update("phrases", array("content"), array(convertSpecialChars($fullDesc)), "title = ".sqlVal($nameId."_FULL_DESC"));
			}
			else
			{
				Core::getQuery()->insert("phrases", array("languageid", "phrasegroupid", "title", "content"), array($languageId, 4, $nameId."_FULL_DESC", convertSpecialChars($fullDesc)));
			}
			sqlEnd($result);
		}

		// Rebuild language cache
		Core::getLang()->rebuild("info");
		return $this->index();
	}

	/**
	* Creates the options of all resources.
	*
	* @param string	Pre-selected entry
	*
	* @return string	Option list
	*/
	protected function getResourceSelect($what)
	{
		$options = "";
		foreach($this->resources as $key => $value)
		{
			if($what == $key) { $s = 1; } else { $s = 0; }
			$options .= createOption($key, Core::getLang()->getItem($value), $s);
		}
		return $options;
	}
}
?>