<?php
/**
* Shows all available constructions and their requirements.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class Techtree extends Page
{
	/**
	* Main method to display techtree.
	*
	* @return void
	*/
	public function __construct()
	{
		Core::getLanguage()->load("info,buildings");
		$this->proceedRequest();
		return;
	}

	/**
	* Index action.
	*
	* @return Techtree
	*/
	protected function index()
	{
		$reqs = NS::getAllRequirements();
		$cons = array(); $research = array(); $ships = array(); $def = array(); $moon = array();
		$result = sqlSelect("construction", array("buildingid", "mode", "name"), "", "1=1".(!isAdmin() ? " AND test = 0" : "")." ORDER BY display_order ASC, buildingid ASC");
		while($row = sqlFetch($result))
		{
			// Hook::event("LOAD_TECHTREE", array(&$row));
			$bid = $row["buildingid"];
            if($bid == UNIT_EXPO_TECH && !EXPEDITION_ENABLED){
                continue;
            }
			if(!isset($reqs[$bid]))
			{
				$reqs[$bid] = array();
			}
			$requirements = array();
			foreach($reqs[$bid] as $r)
			{
				$rLevel = 0;
				if($r["mode"] == UNIT_TYPE_CONSTRUCTION || $r["mode"] == UNIT_TYPE_MOON_CONSTRUCTION)
				{
					$rLevel = NS::getPlanet()->getBuilding($r["needs"]);
				}
				else if($r["mode"] == UNIT_TYPE_RESEARCH)
				{
					$rLevel = NS::getResearch($r["needs"]);
				}
                switch($r["mode"])
                {
                    case UNIT_TYPE_FLEET:
                    case UNIT_TYPE_DEFENSE:
                        $page = "UnitInfo";
                        break;
                    case UNIT_TYPE_ACHIEVEMENT:
                        $page = "AchievementInfo";
                        break;
                    case UNIT_TYPE_ARTEFACT:
                        $page = "ArtefactInfo";
                        break;
                    case UNIT_TYPE_CONSTRUCTION:
                    case UNIT_TYPE_MOON_CONSTRUCTION:
                        $page = "ConstructionInfo";
                        break;
                    case UNIT_TYPE_RESEARCH:
                        $page = "ResearchInfo";
                        break;
                    default:
                        $page = "UnknownInfo";
                        break;
                }

				$style_class = $rLevel >= $r["level"] ? "true" : "false";
				$requirements[] = "<span class='$style_class'>"
					. Link::get("game.php/{$page}/".$r["needs"], Core::getLanguage()->getItem($r["name"]), "", "$style_class")
					. " <nobr>(" . Core::getLanguage()->getItem("LEVEL") . " " . $r["level"]
				. ($rLevel < $r["level"] ? ", +" . ($r["level"] - $rLevel) : "")
					. ")</nobr></span>";
			}
			$requirements = implode("<br />", $requirements);
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
		Core::getTPL()->addLoop("construction", $cons);
		Core::getTPL()->addLoop("research", $research);
		Core::getTPL()->addLoop("shipyard", $ships);
		Core::getTPL()->addLoop("defense", $def);
		Core::getTPL()->addLoop("moon", $moon);
		Core::getTPL()->display("techtree");
		return $this;
	}
}
?>