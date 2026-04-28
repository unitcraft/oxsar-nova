<?php
/**
* Empire module.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class Empire extends Page
{
	public function __construct()
	{
		Core::getLanguage()->load("info,buildings");
		$this->proceedRequest();
		return;
	}

	protected function index()
	{
		// $p_info = "";
		$p_i = 1;
		$planets = array();
		$planetData = array();
		$num = "";
		$num_planet = sqlSelect("planet", array("planetid", "planetname", "diameter", "temperature"), "", "userid = ".sqlUser()." ORDER BY planetid ASC");
		Core::getTPL()->assign("totalnum", Core::getDB()->num_rows($num_planet));
		Core::getTPL()->assign("totalnum1", Core::getDB()->num_rows($num_planet)+1);
		while($planet = sqlFetch($num_planet))
		{
			$planetObject = new Planet($planet["planetid"], NS::getUser()->get("userid"));

			$planets[] = array(
				"num" => $p_i,
				"selected" => $planetObject->getPlanetId() == NS::getPlanet()->getPlanetId(),
				"fields" => $planetObject->getFields(true),
				"maxFields" => $planetObject->getMaxFields(),
				"maxFields2" => $planetObject->getMaxFields(true),
				"ismoon" => $planetObject->getData("ismoon"),
				"diameter" => $planetObject->getData("diameter"),
				"temperature" => $planetObject->getData("temperature"),
				"planetname" => $planetObject->getData("planetname"),
				"coords" => $planetObject->getCoords(true, "select"),
				"metal" => fNumber($planetObject->getData("metal")),
				"silicon" => fNumber($planetObject->getData("silicon")),
				"hydrogen" => fNumber($planetObject->getData("hydrogen")),
				"reseachVirtLab" => fNumber($planetObject->getReseachVirtLab()),
			);

			$planetData[$planet["planetid"]] = array();
			/*
			$p_info .= "<tr><td>".$p_i."</td><td>".$planet["planetname"]."<br />".$planetObject->getCoords(true, 'select')."</td>"
				. "<td>".$planet["diameter"]."</td>"
				. "<td>".$planetObject->getFields(true)." (".$planetObject->getMaxFields().")</td>"
				. "<td>".$planet["temperature"]." &deg;C</td>"
				. "</tr>";
			*/
			$num .= "<th>№".$p_i."</th>";
			$p_i++;
			$result = sqlSelect("building2planet", array("buildingid", "level"), "", "planetid = ".sqlVal($planet["planetid"])." ORDER BY buildingid ASC");
			while($_row = sqlFetch($result))
			{
				$planetData[$planet["planetid"]][$_row["buildingid"]] = $_row["level"];
			}
			sqlEnd($result);

			$result = sqlSelect("unit2shipyard", array("unitid", "quantity", "damaged", "shell_percent"), "", "planetid = ".sqlVal($planet["planetid"])." ORDER BY unitid ASC");
			while($_row = sqlFetch($result))
			{
				$planetData[$planet["planetid"]][$_row["unitid"]] = getUnitQuantityStr($_row, array("splitter" => "<br />", "bracket" => false));
			}
			sqlEnd($result);
		}
		// Core::getTPL()->assign("p_info", $p_info);

		Core::getTPL()->assign("num", $num);

		$reqs = NS::getAllRequirements();
		$cons = array();
		$research = array();
		$ships = array();
		$def = array();
		$moon = array();
		$result = sqlSelect("construction", array("buildingid", "mode", "name"), "", !isAdmin() ? "test = 0" : "", "display_order ASC, buildingid ASC");
		while($row = sqlFetch($result))
		{
			$bid = $row["buildingid"];

			if($row["mode"] == UNIT_TYPE_RESEARCH)
			{
				$requirements = NS::getResearch($row["buildingid"]);
			}
			else
			{
				$requirements = "";
				foreach($planetData as $key => $planet)
				{
					$planet[$row["buildingid"]] = empty($planet[$row["buildingid"]]) ? '-' : $planet[$row["buildingid"]];
					$requirements .= "<td align='center'>".$planet[$row["buildingid"]]."</td>";
				}
			}

			$name = Core::getLanguage()->getItem($row["name"]);
			$image = Image::getImage(getUnitImage($row["name"]), $name, 60);
			switch($row["mode"])
			{
			case UNIT_TYPE_CONSTRUCTION:
			case UNIT_TYPE_MOON_CONSTRUCTION:
				$cons[$bid]["name"] = Link::get("game.php/ConstructionInfo/".$row["buildingid"], $name);
				$cons[$bid]["image"] = Link::get("game.php/ConstructionInfo/".$row["buildingid"], $image);
				$cons[$bid]["requirements"] = $requirements;
				break;
			case UNIT_TYPE_RESEARCH:
				$research[$bid]["name"] = Link::get("game.php/ResearchInfo/".$row["buildingid"], $name);
				$research[$bid]["image"] = Link::get("game.php/ResearchInfo/".$row["buildingid"], $image);
				$research[$bid]["requirements"] = $requirements;
				break;
			case UNIT_TYPE_FLEET:
				$ships[$bid]["name"] = Link::get("game.php/UnitInfo/".$row["buildingid"], $name);
				$ships[$bid]["image"] = Link::get("game.php/UnitInfo/".$row["buildingid"], $image);
				$ships[$bid]["requirements"] = $requirements;
				break;
			case UNIT_TYPE_DEFENSE:
				$def[$bid]["name"] = Link::get("game.php/UnitInfo/".$row["buildingid"], $name);
				$def[$bid]["image"] = Link::get("game.php/UnitInfo/".$row["buildingid"], $image);
				$def[$bid]["requirements"] = $requirements;
				break;
			}
		}
		sqlEnd($result);
		// Hook::event("SHOW_TECHTREE_LOADED", array(&$cons, &$research, &$ships, &$def, &$moon, &$moon));
		Core::getTPL()->addLoop("planets", $planets);
		Core::getTPL()->addLoop("construction", $cons);
		Core::getTPL()->addLoop("research", $research);
		Core::getTPL()->addLoop("shipyard", $ships);
		Core::getTPL()->addLoop("defense", $def);
		Core::getTPL()->addLoop("moon", $moon);
		Core::getTPL()->display("empire");
		return $this;
	}
}
?>