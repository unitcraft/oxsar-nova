<?php
/**
* Shows Infos about the building and demolish function.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class BuildingInfo extends Page
{
	/**
	* General informations on buildings.
	*
	* @return void
	*/
	public function __construct()
	{
		parent::__construct();
			$this->setGetAction("go", "BuildingInfo", "showInfo")
				->addGetArg("showInfo", "id")
				->setGetAction("go", "PackConstruction", "packCurrentConstruction")
				->setGetAction("go", "PackResearch", "packCurrentResearch")
				->addGetArg("packCurrentConstruction", "id")
				->addGetArg("packCurrentConstruction", "Artefact")
				->addGetArg("packCurrentResearch", "id")
				->addGetArg("packCurrentResearch", "Artefact");

		$this->proceedRequest();

		return;
	}

	/**
	* Index action.
	*
	* @return BuildingInfo
	*/
	protected function index()
	{
		$this->showInfo(Core::getRequest()->getGET("id"));
		return $this;
	}

	/**
	* Shows all building information.
	*
	* @param integer	Building id
	*
	* @return BuildingInfo
	*/
	protected function showInfo($id)
	{
		$events = NS::getEH()->getBuildingEvents();
		$free_queue_size = getMaxCountListConstructions() - count($events);

		$select = array(
			"name", "demolish", "mode",
			"basic_metal", "basic_silicon", "basic_hydrogen", "basic_energy",
			"prod_metal", "prod_silicon", "prod_hydrogen", "prod_energy", "special",
			"cons_metal", "cons_silicon", "cons_hydrogen", "cons_energy",
			"charge_metal", "charge_silicon", "charge_hydrogen", "charge_energy"
		);
		$row = sqlSelectRow("construction", $select, "", "buildingid = ".sqlVal($id)." AND mode in (".sqlArray(UNIT_TYPE_CONSTRUCTION, UNIT_TYPE_RESEARCH, UNIT_TYPE_MOON_CONSTRUCTION).")");
		if($row)
		{
			Core::getLanguage()->load("info,Resource");

			$is_building = $row["mode"] != UNIT_TYPE_RESEARCH;
			if(isset($GLOBALS["MAX_UNIT_LEVELS"][$id])){
				$max_level = $GLOBALS["MAX_UNIT_LEVELS"][$id];
			}else{
				$max_level = $is_building ? MAX_BUILDING_LEVEL : MAX_RESEARCH_LEVEL;
			}
			$max_level_text = Core::getLanguage()->getItemWith("MAX_LEVEL", array("level" => $max_level));

			// Assign general building data
			Core::getTPL()->assign("building_name", Core::getLanguage()->getItem($row["name"]));
			Core::getTPL()->assign("building_desc", Core::getLanguage()->getItem($row["name"]."_FULL_DESC")."<p />".$max_level_text);
			Core::getTPL()->assign("building_image", Image::getImage(getUnitImage($row["name"]), Core::getLanguage()->getItem($row["name"]), null, null, "leftImage"));
			Core::getTPL()->assign("edit", Link::get("game.php/EditConstruction/".$id, "[".Core::getLanguage()->getItem("EDIT")."]"));

			// Production and consumption of the building
			$prodFormula = false;
			$prodFactor = 1;
			if(!empty($row["prod_metal"]))
			{
				$prodFormula = $row["prod_metal"];
				$baseCost = $row["basic_metal"];
				$prodFactor = NS::getPlanet()->getData("produce_factor") * Core::getConfig()->get("PRODUCTION_FACTOR");
			}
			else if(!empty($row["prod_silicon"]))
			{
				$prodFormula = $row["prod_silicon"];
				$baseCost = $row["basic_metal"];
				$prodFactor = NS::getPlanet()->getData("produce_factor") * Core::getConfig()->get("PRODUCTION_FACTOR");
			}
			else if(!empty($row["prod_hydrogen"]))
			{
				$prodFormula = $row["prod_hydrogen"];
				$baseCost = $row["basic_hydrogen"];
				$prodFactor = NS::getPlanet()->getData("produce_factor") * Core::getConfig()->get("PRODUCTION_FACTOR");
			}
			else if(!empty($row["prod_energy"]))
			{
				$prodFormula = $row["prod_energy"];
				$baseCost = $row["basic_energy"];
				$prodFactor = NS::getPlanet()->getData("energy_factor");
			}
			else if(!empty($row["special"]))
			{
				$prodFormula = $row["special"];
				$baseCost = 0;
				switch($id)
				{
					case UNIT_METAL_STORAGE:
					case UNIT_SILICON_STORAGE:
					case UNIT_HYDROGEN_STORAGE:
					// case UNIT_REPAIR_FACTORY:
                    // case UNIT_MOON_REPAIR_FACTORY:
						$prodFactor = NS::getPlanet()->getData("storage_factor");
						break;

                    default:
                        $prodFactor = 1;
                        break;
				}
			}
			$consFormula = false;
			if(!empty($row["cons_metal"]))
			{
				$consFormula = $row["cons_metal"];
			}
			else if(!empty($row["cons_silicon"]))
			{
				$consFormula = $row["cons_silicon"];
			}
			else if(!empty($row["cons_hydrogen"]))
			{
				$consFormula = $row["cons_hydrogen"];
			}
			else if(!empty($row["cons_energy"]))
			{
				$consFormula = $row["cons_energy"];
			}

			$level = $is_building ? NS::getPlanet()->getBuilding($id) : NS::getResearch($id);

			// Production and consumption chart
			$chartType = "error";
			if($prodFormula != false || $consFormula != false)
			{
				$chart = array();
				$chartType = "cons_chart";
				if($prodFormula && $consFormula)
				{
					$chartType = "prod_and_cons_chart";
				}
				else if($prodFormula)
				{
					$chartType = "prod_chart";
				}

				// $start = max(0, $level - 7);
				$end = isset($GLOBALS["MAX_UNIT_LEVELS"][$id]) ? $GLOBALS["MAX_UNIT_LEVELS"][$id] : ($is_building ? MAX_BUILDING_LEVEL : MAX_RESEARCH_LEVEL);
				$end = min($end, $level + 7);
				$start = max(0, $end - 14);
				for($i = $start; $i <= $end; $i++)
				{
					$chart[$i]["level"]		= $i;
					$chart[$i]["s_prod"]	= ($prodFormula) ? round(NS::parseFormula($prodFormula, $baseCost, $i)*$prodFactor) : 0;
					$chart[$i]["s_diffProd"]= ($prodFormula) ? $chart[$i]["s_prod"] - round(NS::parseFormula($prodFormula, $baseCost, $level)*$prodFactor) : 0;
					$chart[$i]["s_cons"]	= ($consFormula) ? NS::parseFormula($consFormula, 0, $i) : 0;
					$chart[$i]["s_diffCons"]= ($consFormula) ? NS::parseFormula($consFormula, 0, $level) - $chart[$i]["s_cons"] : 0;
					$chart[$i]["prod"]		= fNumber($chart[$i]["s_prod"]);
					$chart[$i]["diffProd"]	= fNumber($chart[$i]["s_diffProd"]);
					$chart[$i]["cons"]		= fNumber($chart[$i]["s_cons"]);
					$chart[$i]["diffCons"]	= fNumber($chart[$i]["s_diffCons"]);
				}
				Core::getTPL()->addLoop("chart", $chart);
			}
			Core::getTPL()->assign("chartType", $chartType);
			// Show demolish function
			$factor = floatval($row["demolish"]);
			if($is_building)
			{
				if($level > 0 && $factor > 0.0)
				{
					$true_level = $level - NS::getPlanet()->getAddedBuilding($id);
					Core::getTPL()->assign("demolish", $true_level > 0);
					Core::getTPL()->assign("building_level", $true_level);

					if($row["basic_metal"] > 0)
					{
						$d_metal = NS::getPlanet()->getData("metal");
						$n_metal = (1 / $factor) * parseChargeFormula($row["charge_metal"], $row["basic_metal"], $true_level);
						$_metal = $d_metal - $n_metal;
						$metal = Core::getLanguage()->getItem("METAL") . ": "
							. ($_metal >= 0 ? "<span class='true'>".fNumber($n_metal)."</span>" : "<span class='notavailable'>".fNumber($n_metal)."</span> (".fNumber($_metal).")");
					}
					else { $metal = ""; }
					Core::getTPL()->assign("metal", $metal);

					if($row["basic_silicon"] > 0)
					{
						$d_silicon = NS::getPlanet()->getData("silicon");
						$n_silicon = (1 / $factor) * parseChargeFormula($row["charge_silicon"], $row["basic_silicon"], $true_level);
						$_silicon = $d_silicon - $n_silicon;
						$silicon = Core::getLanguage()->getItem("SILICON") . ": "
							. ($_silicon >= 0 ? "<span class='true'>".fNumber($n_silicon)."</span>" : "<span class='notavailable'>".fNumber($n_silicon)."</span> (".fNumber($_silicon).")");
					}
					else { $silicon = ""; }
					Core::getTPL()->assign("silicon", $silicon);

					if($row["basic_hydrogen"] > 0)
					{
						$d_hydrogen = NS::getPlanet()->getData("hydrogen");
						$n_hydrogen = (1 / $factor) * parseChargeFormula($row["charge_hydrogen"], $row["basic_hydrogen"], $true_level);
						$_hydrogen = $d_hydrogen - $n_hydrogen;
						$hydrogen = Core::getLanguage()->getItem("HYDROGEN") . ": "
							. ($_hydrogen >= 0 ? "<span class='true'>".fNumber($n_hydrogen)."</span>" : "<span class='notavailable'>".fNumber($n_hydrogen)."</span> (".fNumber($_hydrogen).")");
					}
					else { $hydrogen = ""; }
					Core::getTPL()->assign("hydrogen", $hydrogen);

					$time = NS::getBuildingTime($n_metal, $n_silicon, UNIT_TYPE_CONSTRUCTION);
					Core::getTPL()->assign("dimolishTime", getTimeTerm($time));

					$is_in_queue = false;
					foreach($events as $event)
					{
						if($event["data"]["buildingid"] == $id)
						{
							$is_in_queue = true;
							break;
						}
					}
					if(!$is_in_queue && $free_queue_size > 0 && ( $_metal >= 0 && $_silicon >= 0 && $_hydrogen >= 0 ) )
					{
						Core::getTPL()->assign("demolish_now", Link::get("game.php/DemolishConstruction/$id", Core::getLanguage()->getItem("DEMOLISH_NOW")));
					}

					$can_pack_building = Artefact::canPackBuilding(
							NS::getPlanet()->getPlanetId(),
							$id,
							NS::getPlanet()->getData("userid")
						);
					if ( $can_pack_building ) {
						Core::getTPL()->assign("pack_building", Link::get("game.php/PackConstruction/$id/Artefact:$can_pack_building", Core::getLanguage()->getItem("PACK_BUILDING_NOW")));;
					}
				}
			}
			else
			{
				$true_level = $level - NS::getAddedResearch($id);
				Core::getTPL()->assign("demolish", false);
				Core::getTPL()->assign("building_level", $true_level);

				// $true_level = NS::getResearch($id) - NS::getAddedResearch($id);
				$can_pack_building = Artefact::canPackResearch(
						NS::getPlanet()->getPlanetId(),
						$id,
						$_SESSION["userid"] ?? 0
					);
				if ( $can_pack_building )
				{
					Core::getTPL()->assign("pack_research", Link::get("game.php/PackResearch/$id/Artefact:$can_pack_building", Core::getLanguage()->getItem("PACK_RESEARCH_NOW")));;
				}
			}

			Core::getTPL()->assign("prod_invert", (int)in_array($id, array(UNIT_GRAVI, UNIT_STAR_GATE)));

			// Hook::event("BUILDING_INFO_AFTER", array(&$row));
			Core::getTPL()->display("buildinginfo");
		}
		else
		{
			throw new GenericException("Unkown building. You'd better don't manipulate the URL. We see everything ;)");
		}
		return $this;
	}

	protected function packCurrentConstruction($id, $aid)
	{
		$can_pack_building = Artefact::canPackBuilding(
				NS::getPlanet()->getPlanetId(),
				$id,
				NS::getPlanet()->getData("userid")
			);
		if($can_pack_building && $can_pack_building == $aid)
		{
			$level = sqlSelectField('building2planet', 'level - added', '', 'planetid = ' . sqlVal(NS::getPlanet()->getPlanetId()) . ' AND buildingid = ' . sqlVal($id));
			if($level > 0)
			{
				$params = array(
					'construction_id' => $id,
					'level' => $level,
				);
				Artefact::activate($aid, NS::getPlanet()->getData("userid"), NS::getPlanet()->getPlanetId(), true, $params);
                NS::updateProfession(NS::getPlanet()->getData("userid"));
			}
		}
		doHeaderRedirection("game.php/BuildingInfo/$id", false);
	}

	protected function packCurrentResearch($id, $aid)
	{
		$can_pack_building = Artefact::canPackResearch(
				NS::getPlanet()->getPlanetId(),
				$id,
				NS::getPlanet()->getData("userid")
			);
		if($can_pack_building && $can_pack_building == $aid)
		{
			$level = sqlSelectField('research2user', 'level - added', '', 'userid='.sqlUser().' AND buildingid='.sqlVal($id));
			if($level > 0)
			{
				$params = array(
					'construction_id' => $id,
					'level' => $level,
				);
				Artefact::activate($aid, NS::getPlanet()->getData("userid"), NS::getPlanet()->getPlanetId(), true, $params);
                NS::updateProfession(NS::getPlanet()->getData("userid"));
			}
		}
		doHeaderRedirection("game.php/BuildingInfo/$id", false);
	}
}
?>