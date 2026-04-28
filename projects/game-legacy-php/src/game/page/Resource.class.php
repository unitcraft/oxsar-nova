<?php
/**
* Resources page. Shows resource production.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class Resource extends Page
{
	/**
	* Holds the building data.
	*
	* @var array
	*/
	protected $data = array();

	/**
	* Shows overview on resource production and options.
	*
	* @return void
	*/
	public function __construct()
	{
		parent::__construct();
		Core::getLanguage()->load("Resource,info,buildings");

		$this->loadBuildingData();

		if(!NS::getUser()->get("umode"))
		{
			$this->setPostAction("update", "updateResources")
				->addPostArg("updateResources", null);
		}

		$this->proceedRequest();
		return;
	}

	/**
	* Index action.
	*
	* @return Resource
	*/
	protected function index()
	{
		// Hook::event("SHOW_RESOURCES_BEFORE", array(&$this));
		Core::getTPL()->addLoop("data", $this->data);

		// Basic prod
		Core::getTPL()->assign("basicMetal", fNumber(NS::getPlanet()->getProd("basic_metal")));
		Core::getTPL()->assign("basicSilicon", fNumber(NS::getPlanet()->getProd("basic_silicon")));
		Core::getTPL()->assign("basicHydrogen", fNumber(NS::getPlanet()->getProd("basic_hydrogen")));

		// Storage capicity
		Core::getTPL()->assign("storageMetal", fNumber(NS::getPlanet()->getStorage("metal") / 1000)."k");
		Core::getTPL()->assign("storageSilicon", fNumber(NS::getPlanet()->getStorage("silicon") / 1000)."k");
		Core::getTPL()->assign("sotrageHydrogen", fNumber(NS::getPlanet()->getStorage("hydrogen") / 1000)."k");

		// Total prod
		Core::getTPL()->assign("totalMetal", fNumber(NS::getPlanet()->getProd("metal")));
		Core::getTPL()->assign("totalSilicon", fNumber(NS::getPlanet()->getProd("silicon")));
		Core::getTPL()->assign("totalHydrogen", fNumber(NS::getPlanet()->getProd("hydrogen")));
		Core::getTPL()->assign("totalEnergy", fNumber(NS::getPlanet()->getEnergy()));

		// Daily prod
		Core::getTPL()->assign("dailyMetal", fNumber(NS::getPlanet()->getProd("metal") * 24));
		Core::getTPL()->assign("dailySilicon", fNumber(NS::getPlanet()->getProd("silicon") * 24));
		Core::getTPL()->assign("dailyHydrogen", fNumber(NS::getPlanet()->getProd("hydrogen") * 24));

		// Weekly prod
		Core::getTPL()->assign("weeklyMetal", fNumber(NS::getPlanet()->getProd("metal") * 168));
		Core::getTPL()->assign("weeklySilicon", fNumber(NS::getPlanet()->getProd("silicon") * 168));
		Core::getTPL()->assign("weeklyHydrogen", fNumber(NS::getPlanet()->getProd("hydrogen") * 168));

		// Monthly prod
		Core::getTPL()->assign("monthlyMetal", fNumber(NS::getPlanet()->getProd("metal") * 720));
		Core::getTPL()->assign("monthlySilicon", fNumber(NS::getPlanet()->getProd("silicon") * 720));
		Core::getTPL()->assign("monthlyHydrogen", fNumber(NS::getPlanet()->getProd("hydrogen") * 720));

		$selectbox = "";
		for($i = 10; $i >= 0; $i--)
		{
			$selectbox .= createOption($i * 10, $i * 10, 0);
		}
		Core::getTPL()->assign("selectProd", $selectbox);
		// Hook::event("SHOW_RESOURCES_AFTER");
		Core::getTPL()->display("resource");
		return $this;
	}

	/**
	* Loads all production and consumption data of all available buildings.
	*
	* @return Resource
	*/
	protected function loadBuildingData()
	{
		$where = "c.prod_metal != '' OR c.prod_silicon != '' OR c.prod_hydrogen != '' OR c.prod_energy != '' OR c.cons_metal != '' OR c.cons_silicon != '' OR c.cons_hydrogen != '' OR c.cons_energy != ''";
		$result = sqlSelect("construction c", array("c.buildingid", "c.name", "c.mode"), "", $where, "c.display_order ASC, c.buildingid ASC");
		while($row = sqlFetch($result))
		{
			$id = $row["buildingid"];
			$this->data[$id]["id"] = $id;
			// $this->data[$id]["mode"] = $row["mode"];
			$this->data[$id]["name"] = Core::getLang()->getItem($row["name"]);
			$this->data[$id]["level"] = NS::getPlanet()->getBuilding($id);
			$this->data[$id]["factor"] = NS::getPlanet()->getBuildingFactor($id);
			$this->data[$id]["metal"] = fNumber(NS::getPlanet()->getBuildingProd("metal", $id));
			$this->data[$id]["silicon"] = fNumber(NS::getPlanet()->getBuildingProd("silicon", $id));
			$this->data[$id]["hydrogen"] = fNumber(NS::getPlanet()->getBuildingProd("hydrogen", $id));
			$this->data[$id]["energy"] = fNumber(NS::getPlanet()->getBuildingProd("energy", $id));
			$this->data[$id]["metalCons"] = fNumber(NS::getPlanet()->getBuildingCons("metal", $id));
			$this->data[$id]["siliconCons"] = fNumber(NS::getPlanet()->getBuildingCons("silicon", $id));
			$this->data[$id]["hydrogenCons"] = fNumber(NS::getPlanet()->getBuildingCons("hydrogen", $id));
			$this->data[$id]["energyCons"] = fNumber(NS::getPlanet()->getBuildingCons("energy", $id));
			
			$this->data[$id]["allow_factor"] = $row["mode"] == UNIT_TYPE_CONSTRUCTION || $row["mode"] == UNIT_TYPE_MOON_CONSTRUCTION;
		}

		if(!isset($this->data[UNIT_SOLAR_SATELLITE]) && NS::getPlanet()->getBuilding(UNIT_SOLAR_SATELLITE) > 0)
		{
			$id = UNIT_SOLAR_SATELLITE;
			$this->data[$id]["id"] = $id;
			$this->data[$id]["name"] = Core::getLang()->getItem("SOLAR_SATELLITE");
			$this->data[$id]["level"] = NS::getPlanet()->getBuilding($id);
			$this->data[$id]["factor"] = NS::getPlanet()->getData("solar_satellite_prod");
			$this->data[$id]["metal"] = 0; // fNumber(NS::getPlanet()->getBuildingProd("metal", $id));
			$this->data[$id]["silicon"] = 0; // fNumber(NS::getPlanet()->getBuildingProd("silicon", $id));
			$this->data[$id]["hydrogen"] = 0; // fNumber(NS::getPlanet()->getBuildingProd("hydrogen", $id));
			$this->data[$id]["energy"] = fNumber(NS::getPlanet()->getBuildingProd("energy", $id));
			$this->data[$id]["metalCons"] = 0; // fNumber(NS::getPlanet()->getBuildingCons("metal", $id));
			$this->data[$id]["siliconCons"] = 0; // fNumber(NS::getPlanet()->getBuildingCons("silicon", $id));
			$this->data[$id]["hydrogenCons"] = 0; // fNumber(NS::getPlanet()->getBuildingCons("hydrogen", $id));
			$this->data[$id]["energyCons"] = 0; // fNumber(NS::getPlanet()->getBuildingCons("energy", $id));

			$this->data[$id]["helptip"] = "<span class=true>" . Core::getLang()->getItemWith('PROD_ONE_SOLAR_SATELLITE', array(
					'prodPerSat' => fNumber(NS::getPlanet()->getBuildingProd("energy", $id) / NS::getPlanet()->getBuilding($id), 2),
				)) . "</span>";
		}
		
		$halting_fleets = NS::getPlanet()->getHaltingFleets();
		$unit_virt_end = UNIT_VIRT_HALTING_START + count($halting_fleets);
		for($id = UNIT_VIRT_HALTING_START - 3; $id < $unit_virt_end; $id++)
		{
			if(NS::getPlanet()->getBuilding($id) <= 0)
			{
				continue;
			}
			if($id == UNIT_VIRT_FLEET)
			{
				$name = Core::getLang()->getItem("FLEET_CONSUMPTION");
			}
			else if($id == UNIT_VIRT_STOCK_FLEET)
			{
				$name = Core::getLang()->getItem("FLEET_STOCK_CONSUMPTION");
			}
			else if($id == UNIT_VIRT_DEFENSE)
			{
				$name = Core::getLang()->getItem("DEFENSE_CONSUMPTION");
			}
			else
			{
				$name = Core::getLang()->getItemWith("HALTING_CONSUMPTION", array(
						"planet" => getCoordLink(null, null, null, false, $halting_fleets[$id - UNIT_VIRT_HALTING_START]["planetid"], true),
						"fleet" => getUnitListStr($halting_fleets[$id - UNIT_VIRT_HALTING_START]["ships"]),
					));
			}
		
			$this->data[$id]["id"] = $id;
			$this->data[$id]["name"] = $name;
			$this->data[$id]["level"] = NS::getPlanet()->getBuilding($id);
			$this->data[$id]["factor"] = NS::getPlanet()->getData("solar_satellite_prod");
			$this->data[$id]["metal"] = 0; // fNumber(NS::getPlanet()->getBuildingProd("metal", $id));
			$this->data[$id]["silicon"] = 0; // fNumber(NS::getPlanet()->getBuildingProd("silicon", $id));
			$this->data[$id]["hydrogen"] = 0; // fNumber(NS::getPlanet()->getBuildingProd("hydrogen", $id));
			$this->data[$id]["energy"] = 0; // fNumber(NS::getPlanet()->getBuildingProd("energy", $id));
			$this->data[$id]["metalCons"] = fNumber(NS::getPlanet()->getBuildingCons("metal", $id));
			$this->data[$id]["siliconCons"] = fNumber(NS::getPlanet()->getBuildingCons("silicon", $id));
			$this->data[$id]["hydrogenCons"] = fNumber(NS::getPlanet()->getBuildingCons("hydrogen", $id));
			$this->data[$id]["energyCons"] = 0; // fNumber(NS::getPlanet()->getBuildingCons("energy", $id));
			$this->data[$id]["allow_factor"] = false;

			$this->data[$id]["helptip"] = "<span class=false2>" . Core::getLang()->getItemWith(
				($this->data[$id]["metalCons"] || $this->data[$id]["siliconCons"] ? 'ONE_UNIT_CONSUMPTION_EXT' : 'ONE_UNIT_CONSUMPTION'),
				array(
					'metalCons' => fNumber($this->data[$id]["metalCons"] / $this->data[$id]["level"], 4),
					'siliconCons' => fNumber($this->data[$id]["siliconCons"] / $this->data[$id]["level"], 4),
					'hydrogenCons' => fNumber($this->data[$id]["hydrogenCons"] / $this->data[$id]["level"], 4),
					'basicHydrogenCons' => fNumber(NS::getPlanet()->getBuildingCons("basic_hydrogen", $id) / $this->data[$id]["level"], 4),
				)
				) . "</span>";
		}
	
		return $this;
	}

	/**
	* Saves resource production factor.
	*
	* @param array		_POST
	*
	* @return Resource
	*/
	protected function updateResources($post)
	{
		// Hook::event("SAVE_RESOURCES");
		foreach($this->data as $key => $value)
		{
			if($value["level"] > 0)
			{
				$factor = abs($post[$key]);
				$factor = ($factor > 100) ? 100 : $factor;
				Core::getQuery()->update(
					"building2planet",
					array("prod_factor"),
					array($factor),
					"buildingid = ".sqlVal($key)." AND planetid = ".sqlPlanet()
						. ' ORDER BY buildingid DESC, planetid DESC'
				);
			}
		}

		if(NS::getPlanet()->getBuilding(UNIT_SOLAR_SATELLITE) > 0)
		{
			$satelliteProd = abs($post[UNIT_SOLAR_SATELLITE]);
			$satelliteProd = ($satelliteProd > 100) ? 100 : $satelliteProd;
			Core::getQuery()->update(
				"planet",
				array("solar_satellite_prod"),
				array($satelliteProd),
				"planetid = ".sqlPlanet()
					. ' ORDER BY planetid'
			);
		}
		doHeaderRedirection("game.php/Resource", false);
		return $this;
	}
}
?>