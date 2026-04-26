<?php
/**
* Common functions to check constructions.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

abstract class Construction extends Page
{
	/**
	* Holds the required resources of the current building.
	*
	* @var integer
	*/
	protected $requiredMetal = 0, $requiredSilicon = 0, $requiredHydrogen = 0, $requiredEnergy = 0, $requiredCredit = 0;

	/**
	* Constructions
	* 1: Planet, 4: Moon
	*
	* Research
	* 2: Research
	*
	* Shipyard display mode
	* 3: Fleet, 4: Defense
	*
	* @var integer
	*/
	protected $unit_type = false;
	protected $main_page = "unknown";

	protected function eventType()
	{
		switch($this->unit_type)
		{
		case UNIT_TYPE_CONSTRUCTION:
		case UNIT_TYPE_MOON_CONSTRUCTION:
			return EVENT_BUILD_CONSTRUCTION;

		case UNIT_TYPE_RESEARCH:
			return EVENT_RESEARCH;

		case UNIT_TYPE_FLEET:
			return EVENT_BUILD_FLEET;

		case UNIT_TYPE_DEFENSE:
			return EVENT_BUILD_DEFENSE;

		case UNIT_TYPE_REPAIR:
			return EVENT_REPAIR;

		case UNIT_TYPE_DISASSEMBLE:
			return EVENT_DISASSEMBLE;

		default:
			return false;
		}
	}

	public function __construct()
	{
		parent::__construct();

		$page = Core::getRequest()->getGET("go");
		if(preg_match("#construction#is", $page))
		{
			$this->unit_type = NS::getPlanet()->getData("ismoon") ? UNIT_TYPE_MOON_CONSTRUCTION : UNIT_TYPE_CONSTRUCTION;
			$this->main_page = "game.php/Constructions";
		}
		else if(preg_match("#research#is", $page))
		{
			$this->unit_type = UNIT_TYPE_RESEARCH;
			$this->main_page = "game.php/Research";
		}
		else if(preg_match("#shipyard#is", $page))
		{
			$this->unit_type = UNIT_TYPE_FLEET;
			$this->main_page = "game.php/Shipyard";
		}
		else if(preg_match("#defense#is", $page))
		{
			$this->unit_type = UNIT_TYPE_DEFENSE;
			$this->main_page = "game.php/Defense";
		}
		else if(preg_match("#repair#is", $page))
		{
			$this->unit_type = UNIT_TYPE_REPAIR;
			$this->main_page = "game.php/Repair";
		}
		else if(preg_match("#disassemble#is", $page))
		{
			$this->unit_type = UNIT_TYPE_DISASSEMBLE;
			$this->main_page = "game.php/Disassemble";
		}
		else if(preg_match("#artefact#is", $page))
		{
			$this->unit_type = UNIT_TYPE_ARTEFACT;
			$this->main_page = "game.php/ArtefactMarket";
		}
		else
		{
			throw new GenericException("Unkown construction mode: '$page'");
		}

		if(!NS::getUser()->get("umode"))
		{
			$this
				->setPostAction("image_package", "updateUserImagePak")
				->addPostArg("updateUserImagePak", "image_package")
				->setPostAction("show_all_units", "updateUserShowAllUnits")
				->addPostArg("updateUserShowAllUnits", "show_all_units")
				;
		}
	}

	protected function updateUserImagePak($image_package)
	{
		$image_package = empty($image_package) || preg_match("#[^\w\d\-_]#is", $image_package) ? "std" : $image_package;
		if(!is_dir(APP_ROOT_DIR."images/buildings/".$image_package))
		{
			$image_package = "std";
		}
		NS::getUser()->set("imagepackage", $image_package);

		doHeaderRedirection($this->main_page, false);
		return $this;
	}

	protected function updateUserShowAllUnits($show_all_units)
	{
		$field = "";
		switch($this->unit_type)
		{
		case UNIT_TYPE_CONSTRUCTION:
		case UNIT_TYPE_MOON_CONSTRUCTION:
			$field = "show_all_constructions";
			break;

		case UNIT_TYPE_RESEARCH:
			$field = "show_all_research";
			break;

		case UNIT_TYPE_FLEET:
			$field = "show_all_shipyard";
			break;

		case UNIT_TYPE_DEFENSE:
			$field = "show_all_defense";
			break;

		default:
			return false;
		}
		NS::getUser()->set($field, empty($show_all_units) ? 0 : 1);

		doHeaderRedirection($this->main_page, false);
		return $this;
	}

	public function canShowAllUnits()
	{
		switch($this->unit_type)
		{
		case UNIT_TYPE_CONSTRUCTION:
		case UNIT_TYPE_MOON_CONSTRUCTION:
			return NS::getUser()->get("show_all_constructions");

		case UNIT_TYPE_RESEARCH:
			return NS::getUser()->get("show_all_research");

		case UNIT_TYPE_FLEET:
			return NS::getUser()->get("show_all_shipyard");

		case UNIT_TYPE_DEFENSE:
			return NS::getUser()->get("show_all_defense");

		case UNIT_TYPE_ARTEFACT:
			return true;

		default:
			return false;
		}
	}

	/**
	* Stores the required resources for the current building.
	*
	* @param integer	Next level
	* @param integer	Construction data
	*
	* @return Construction
	*/
	protected function setRequieredResources($nextLevel, $row)
	{
		if($nextLevel > 1)
		{
			if($row["basic_metal"] > 0)
			{
				$this->requiredMetal = parseChargeFormula($row["charge_metal"], $row["basic_metal"], $nextLevel);
			}
			else { $this->requiredMetal = 0; }
			if($row["basic_silicon"] > 0)
			{
				$this->requiredSilicon = parseChargeFormula($row["charge_silicon"], $row["basic_silicon"], $nextLevel);
			}
			else { $this->requiredSilicon = 0; }
			if($row["basic_hydrogen"] > 0)
			{
				$this->requiredHydrogen = parseChargeFormula($row["charge_hydrogen"], $row["basic_hydrogen"], $nextLevel);
			}
			else { $this->requiredHydrogen = 0; }
			if($row["basic_energy"] > 0)
			{
				$this->requiredEnergy = parseChargeFormula($row["charge_energy"], $row["basic_energy"], $nextLevel);
			}
			else { $this->requiredEnergy = 0; }
			if($row["basic_credit"] > 0)
			{
				$this->requiredCredit = parseChargeFormula($row["charge_credit"], $row["basic_credit"], $nextLevel);
			}
			else { $this->requiredCredit = 0; }
			if($row["basic_points"] > 0)
			{
				$this->requiredPoints = parseChargeFormula($row["charge_points"], $row["basic_points"], $nextLevel);
			}
			else { $this->requiredPoints = 0; }
		}
		else
		{
			$this->requiredMetal = intval($row["basic_metal"]);
			$this->requiredSilicon = intval($row["basic_silicon"]);
			$this->requiredHydrogen = intval($row["basic_hydrogen"]);
			$this->requiredEnergy = intval($row["basic_energy"]);
			$this->requiredCredit = intval($row["basic_credit"]);
			$this->requiredPoints = intval($row["basic_points"]);
		}
		return $this;
	}

	/**
	* Checks the resources.
	*
	* @return boolean	True if resources suffice, fals if not
	*/
	protected function checkResources($use_store_max = true)
	{
		if($this->requiredMetal > NS::getPlanet()->getAvailableRes("metal", $use_store_max))
		{
			return false;
		}
		if($this->requiredSilicon > NS::getPlanet()->getAvailableRes("silicon", $use_store_max))
		{
			return false;
		}
		if($this->requiredHydrogen > NS::getPlanet()->getAvailableRes("hydrogen", $use_store_max))
		{
			return false;
		}
		if($this->requiredEnergy > max(0, NS::getPlanet()->getEnergy()))
		{
			return false;
		}
		if($this->requiredCredit > NS::getUser()->get("credit"))
		{
			return false;
		}
		if($this->requiredPoints > NS::getUser()->get("points"))
		{
			return false;
		}
		return true;
	}

    protected function getChartData($row)
    {
        $result = array();
        if($row["mode"] != UNIT_TYPE_RESEARCH && $row["mode"] != UNIT_TYPE_CONSTRUCTION && $row["mode"] != UNIT_TYPE_MOON_CONSTRUCTION){
            return $result;
        }

        $is_building = $row["mode"] != UNIT_TYPE_RESEARCH;
        if($is_building){
            $events = NS::getEH()->getBuildingEvents();
            $free_queue_size = getMaxCountListConstructions() - count($events);
        }else{
            $events = NS::getEH()->getResearchEvents();
            $free_queue_size = getMaxCountListResearch() - count($events);
        }

        $id = $row['buildingid'];
        // $result['id'] = $id;

        if(isset($GLOBALS["MAX_UNIT_LEVELS"][$id])){
            $max_level = $GLOBALS["MAX_UNIT_LEVELS"][$id];
        }else{
            $max_level = $is_building ? MAX_BUILDING_LEVEL : MAX_RESEARCH_LEVEL;
        }
        $result["name"] = Core::getLanguage()->getItem($row["name"]);
        $result["max_level"] = $max_level;

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
        // $result['level'] = $level;

        // Production and consumption chart
        $chart_type = "error";
        if($prodFormula != false || $consFormula != false)
        {
            $chart = array();
            $chart_type = "cons_chart";
            if($prodFormula && $consFormula)
            {
                $chart_type = "prod_and_cons_chart";
            }
            else if($prodFormula)
            {
                $chart_type = "prod_chart";
            }
            $result['chart_type'] = $chart_type;

            $end = min($max_level, $level + 7);
            $start = max(0, $end - 14);
            $chart = array();
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
            $result['chart'] = $chart;
        }

        // Show demolish function
        $factor = floatval($row["demolish"]);
        if($is_building)
        {
            if($level > 0 && $factor > 0.0)
            {
                $true_level = $level - NS::getPlanet()->getAddedBuilding($id);
                $result["demolish"] = $true_level > 0;
                $result["level"] = $true_level;

                if($row["basic_metal"] > 0)
                {
                    $d_metal = NS::getPlanet()->getData("metal");
                    $n_metal = (1 / $factor) * parseChargeFormula($row["charge_metal"], $row["basic_metal"], $true_level);
                    $_metal = $d_metal - $n_metal;
                    $metal = Core::getLanguage()->getItem("METAL") . ": "
                        . ($_metal >= 0 ? "<span class='true'>".fNumber($n_metal)."</span>" : "<span class='notavailable'>".fNumber($n_metal)."</span> (".fNumber($_metal).")");
                }
                else { $metal = ""; }
                $result["demolish_metal"] = $metal;

                if($row["basic_silicon"] > 0)
                {
                    $d_silicon = NS::getPlanet()->getData("silicon");
                    $n_silicon = (1 / $factor) * parseChargeFormula($row["charge_silicon"], $row["basic_silicon"], $true_level);
                    $_silicon = $d_silicon - $n_silicon;
                    $silicon = Core::getLanguage()->getItem("SILICON") . ": "
                        . ($_silicon >= 0 ? "<span class='true'>".fNumber($n_silicon)."</span>" : "<span class='notavailable'>".fNumber($n_silicon)."</span> (".fNumber($_silicon).")");
                }
                else { $silicon = ""; }
                $result["demolish_silicon"] = $silicon;

                if($row["basic_hydrogen"] > 0)
                {
                    $d_hydrogen = NS::getPlanet()->getData("hydrogen");
                    $n_hydrogen = (1 / $factor) * parseChargeFormula($row["charge_hydrogen"], $row["basic_hydrogen"], $true_level);
                    $_hydrogen = $d_hydrogen - $n_hydrogen;
                    $hydrogen = Core::getLanguage()->getItem("HYDROGEN") . ": "
                        . ($_hydrogen >= 0 ? "<span class='true'>".fNumber($n_hydrogen)."</span>" : "<span class='notavailable'>".fNumber($n_hydrogen)."</span> (".fNumber($_hydrogen).")");
                }
                else { $hydrogen = ""; }
                $result["demolish_hydrogen"] = $hydrogen;

                $time = NS::getBuildingTime($n_metal, $n_silicon, UNIT_TYPE_CONSTRUCTION);
                $result["demolish_time"] = getTimeTerm($time);

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
                    $result["demolish_now"] = Link::get("game.php/DemolishConstruction/$id", Core::getLanguage()->getItem("DEMOLISH_NOW"));
                }

                $can_pack_building = Artefact::canPackBuilding(
                        NS::getPlanet()->getPlanetId(),
                        $id,
                        NS::getPlanet()->getData("userid")
                    );
                if ( $can_pack_building ) {
                    $result["pack_building"] = Link::get("game.php/PackConstruction/$id/Artefact:$can_pack_building", Core::getLanguage()->getItem("PACK_BUILDING_NOW"));
                }
            }
        }
        else
        {
            $true_level = $level - NS::getAddedResearch($id);
            $result["demolish"] = false;
            $result["level"] = $true_level;

            // $true_level = NS::getResearch($id) - NS::getAddedResearch($id);
            $can_pack_building = Artefact::canPackResearch(
                    NS::getPlanet()->getPlanetId(),
                    $id,
                    $_SESSION["userid"] ?? 0
                );
            if ( $can_pack_building )
            {
                $result["pack_research"] = Link::get("game.php/PackResearch/$id/Artefact:$can_pack_building", Core::getLanguage()->getItem("PACK_RESEARCH_NOW"));
            }
        }

        $result["prod_invert"] = (int)in_array($id, array(UNIT_GRAVI, UNIT_STAR_GATE));

        return $result;
    }
}
?>