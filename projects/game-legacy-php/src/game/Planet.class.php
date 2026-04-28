<?php
/**
* Planet class. Loads planet data, updates production.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class Planet
{
	/**
	* Planet data.
	*
	* @var array
	*/
	protected $data = array();

	/**
	* Id of current used planet.
	*
	* @var integer
	*/
	protected $planetid = 0;

	/**
	* Production.
	*
	* @var array
	*/
	protected $prod = array();

	/**
	* Remaining energy.
	*
	* @var integer
	*/
	protected $energy = 0;

	/**
	* Available builings and their level.
	*
	* @var array
	*/
	protected $building = array();
	protected $building_added = array();

	protected $ext_building = array();
	protected $ext_building_added = array();

	protected $halting_fleets = array();
	protected $halting_fleet_size = 0;

	/**
	* Energy consumption of all buildings.
	*
	* @var array
	*/
	protected $consumption = array();

	/**
	* Production to building relation.
	*
	* @var array
	*/
	protected $building_prod = array();

	/**
	* Consumption to building relation.
	*
	* @var array
	*/
	protected $building_cons = array();

	/**
	* Production factor of a building.
	*
	* @var array
	*/
	protected $factors = array();

	/**
	* Metal, silicon and hydrogen storage.
	*
	* @var array
	*/
	protected $storage = array();

	/**
	* The planet's owner.
	*
	* @var integer
	*/
	protected $userid = 0;

	/**
	* Holds a list of all available research labs.
	*
	* @var array
	*/
	// protected $researchLabs = array();

	protected $reseach_virt_lab = false;
	protected $moon_virt_nano_factory = false;

	/**
	* Constructor: Starts planet functions.
	*
	* @return void
	*/
	public function __construct($planetid, $userid)
	{
		$this->planetid = $planetid;
		$this->userid = $userid;
		$this->loadData();
		$this->getProduction();
		$this->addProd();
		return;
	}

	/**
	* Loads all planet data.
	*
	* @return Planet
	*
	* @throws GenericException
	*/
	protected function loadData($genericLoad = true)
	{
		$atts = array("p.planetname", "p.userid", "p.ismoon", "p.picture", "p.temperature", "p.diameter",
			"p.metal", "p.silicon", "p.hydrogen", "p.last", "p.solar_satellite_prod", "p.destroy_eventid",
			"p.build_factor", "u.research_factor",
			"p.produce_factor", "p.energy_factor", "p.storage_factor",
			// "g.galaxy", "g.system", "g.position",
			"p.umi",
			"gm.galaxy AS moongala", "gm.system AS moonsys", "gm.position AS moonpos",
			"IFNULL(g.galaxy, gm.galaxy) as galaxy",
			"IFNULL(g.system, gm.system) as system",
			"IFNULL(g.position, gm.position) as position",
			"IFNULL(g.moonid, gm.moonid) as moonid",
			"b.to AS banto", "b.banid", "u.last AS user_last");
		$this->data = sqlSelectRow("planet p", $atts, "
			LEFT JOIN ".PREFIX."galaxy g ON g.planetid = p.planetid
			LEFT JOIN ".PREFIX."galaxy gm ON gm.moonid = p.planetid
			LEFT JOIN ".PREFIX."ban_u b ON b.userid = p.userid
			LEFT JOIN ".PREFIX."user u ON u.userid = p.userid
			",
			"p.planetid = ".sqlVal($this->planetid)
			. (is_null($this->userid) ? "" : " AND p.userid = ".sqlVal($this->userid)));
		if(!$this->data)
		{
			if($this->planetid)
			{
				if($genericLoad && $this->userid
						&& NS::getUser()
						&& NS::getUser()->get("curplanet") == $this->planetid
						&& NS::getUser()->get("userid") == $this->userid
						)
				{
					$planetid = sqlSelectField("planet", "planetid", "", "userid=".sqlVal($this->userid), "planetid ASC", 1);
					if($planetid > 0)
					{
						if( NS::getUser()->get("curplanet") == NS::getUser()->get("hp") )
						{
							NS::getUser()->set("hp", $planetid);
						}
						NS::getUser()->set("curplanet", $planetid);

						$this->planetid = $planetid;
						return $this->loadData(false);
					}
				}
				throw new GenericException("Could not found planet ".$this->planetid.".");
			}
			// Core::getLanguage()->load("global");
			$this->data = array(
				"picture" => "unformed",
				// "planetname" => Core::getLanguage()->getItem("UNFORMED_PLANET_NAME"),
				);
		}
		if(is_null($this->userid))
		{
			$this->userid = $this->data["userid"];
			// debug_var($this, "[planet]"); exit;
		}
		/* if($this->data["ismoon"])
		{
			$this->data["galaxy"] = $this->data["moongala"];
			$this->data["system"] = $this->data["moonsys"];
			$this->data["position"] = $this->data["moonpos"];
		} */
		$this->data["produce_factor"] *= NS::getGalaxyParam($this->data["galaxy"], "RESOURCES_PRODUCTION_FACTOR", 1);
		$this->data["energy_factor"] *= NS::getGalaxyParam($this->data["galaxy"], "ENEGRY_PRODUCTION_FACTOR", 1);
		$this->data["storage_factor"] *= NS::getGalaxyParam($this->data["galaxy"], "STORAGE_FACTOR", 1);

		return $this;
	}

	public function isBanned()
	{
		return $this->data["banid"] && (is_null($this->data["banto"]) || $this->data["banto"] >= time());
	}

	public function isMoon()
	{
		return $this->data["ismoon"] == 1;
	}

	public function isProdPossible()
	{
		return $this->planetid && !$this->isBanned(); // && !$this->data["ismoon"]; // && (time() - $this->data["user_last"]) < 60*60*24*3;
	}

	/**
	* Sets some standard values to avoid confusion.
	*
	* @return Planet
	*/
	protected function setStandardValues()
	{
		$this->prod["metal"] = 0;
		$this->prod["silicon"] = 0;
		$this->prod["hydrogen"] = 0;
		$this->prod["energy"] = 0;
		$this->prod["mine_metal"] = 0;
		$this->prod["mine_silicon"] = 0;
		$this->prod["mine_hydrogen"] = 0;
		$this->prod["mine_energy"] = 0;
		$this->prod["basic_metal"] = $this->data["ismoon"] ? 0 : Core::getOptions()->get("METAL_BASIC_PROD") * $this->data["produce_factor"];
		$this->prod["basic_silicon"] = $this->data["ismoon"] ? 0 : Core::getOptions()->get("SILICON_BASIC_PROD") * $this->data["produce_factor"];
		$this->prod["basic_hydrogen"] = $this->data["ismoon"] ? 0 : Core::getOptions()->get("HYDROGEN_BASIC_PROD") * $this->data["produce_factor"];
		$this->consumption["metal"] = 0;
		$this->consumption["silicon"] = 0;
		$this->consumption["hydrogen"] = 0;
		$this->consumption["energy"] = 0;
		$this->storage["metal"] = 100000 * $this->data["storage_factor"];
		$this->storage["silicon"] = 100000 * $this->data["storage_factor"];
		$this->storage["hydrogen"] = 100000 * $this->data["storage_factor"];
		$this->storage["repair"] = 0;
		return $this;
	}

	private function updateProduction($row, $level, $produce_factor, $energy_factor)
	{
		$id = intval($row["buildingid"]);
		$factor = isset($row["factor"]) ? $row["factor"] : 100;

		// Production
		if($row["prod_metal"] != "")
		{
			$this->building_prod["metal"][$id] = round($this->parseSpecialFormula($row["prod_metal"], $level) / 100 * $factor * $produce_factor);
			$this->prod["metal"] += $this->building_prod["metal"][$id];
			$this->factors[$id] = $factor;
		}
		else if($row["prod_silicon"] != "")
		{
			$this->building_prod["silicon"][$id] = round($this->parseSpecialFormula($row["prod_silicon"], $level) / 100 * $factor * $produce_factor);
			$this->prod["silicon"] += $this->building_prod["silicon"][$id];
			$this->factors[$id] = $factor;
		}
		else if($row["prod_hydrogen"] != "")
		{
			$this->building_prod["hydrogen"][$id] = round($this->parseSpecialFormula($row["prod_hydrogen"], $level) / 100 * $factor * $produce_factor);
			$this->prod["hydrogen"] += $this->building_prod["hydrogen"][$id];
			$this->factors[$id] = $factor;
		}
		else if($row["prod_energy"] != "")
		{
			$this->building_prod["energy"][$id] = round($this->parseSpecialFormula($row["prod_energy"], $level) / 100 * $factor * $energy_factor);
			$this->prod["energy"] += $this->building_prod["energy"][$id];
			$this->factors[$id] = $factor;
		}

		// Consumption
		if($row["cons_metal"] != "")
		{
			$this->building_cons["metal"][$id] = $this->parseSpecialFormula($row["cons_metal"], $level) / 100 * $factor;
			$this->consumption["metal"] += $this->building_cons["metal"][$id];
		}
		else if($row["cons_silicon"] != "")
		{
			$this->building_cons["silicon"][$id] = $this->parseSpecialFormula($row["cons_silicon"], $level) / 100 * $factor;
			$this->consumption["silicon"] += $this->building_cons["silicon"][$id];
		}
		else if($row["cons_hydrogen"] != "")
		{
			$this->building_cons["hydrogen"][$id] = $this->parseSpecialFormula($row["cons_hydrogen"], $level) / 100 * $factor;
			$this->consumption["hydrogen"] += $this->building_cons["hydrogen"][$id];
		}
		else if($row["cons_energy"] != "")
		{
			$this->building_cons["energy"][$id] = $this->parseSpecialFormula($row["cons_energy"], $level) / 100 * $factor;
			$this->consumption["energy"] += $this->building_cons["energy"][$id];
		}
	}

	/**
	* Fetches building and production data.
	*
	* @return Planet
	*/
	protected function getProduction()
	{
		$produce_factor = (float)($this->data["produce_factor"]  ?? 1);
		$energy_factor  = (float)($this->data["energy_factor"]   ?? 1);
		$storage_factor = (float)($this->data["storage_factor"]  ?? 1);

		$produce_factor *= (float)(Core::getConfig()->get("PRODUCTION_FACTOR") ?: 1);

		$this->setStandardValues();

		$is_prod_possible = $this->isProdPossible();

		$this->building[UNIT_SOLAR_SATELLITE] = getShipyardQuantity(UNIT_SOLAR_SATELLITE, false, $this->planetid);
		$this->building[UNIT_VIRT_FLEET] = 0;
		$this->building[UNIT_VIRT_STOCK_FLEET] = 0;
		$this->building[UNIT_VIRT_DEFENSE] = 0;

		$atts = array("planetid", "c.mode", "sum(`quantity`) as quantity");
		$join = "JOIN ".PREFIX."construction c ON u.unitid=c.buildingid";
		$result = sqlSelect("unit2shipyard u", $atts, $join, "planetid=".sqlVal($this->planetid)." AND u.unitid!=".UNIT_SOLAR_SATELLITE." GROUP BY planetid, c.mode");
		while($row = sqlFetch($result))
		{
			if($row["mode"] == UNIT_TYPE_FLEET)
			{
				$this->building[UNIT_VIRT_FLEET] = $row["quantity"];
			}
			else // if($row["mode"] == UNIT_TYPE_DEFENSE)
			{
				$this->building[UNIT_VIRT_DEFENSE] = $row["quantity"];
			}
		}
		sqlEnd($result);

		$result = sqlSelect("exchange_lots", "data", "", "planetid=".sqlVal($this->planetid)." AND status=".ESTATUS_OK);
		while($row = sqlFetch($result))
		{
			$data = unserialize($row["data"]);
			if(is_array($data["ships"]))
			{
				foreach($data["ships"] as $ships)
				{
					$this->building[UNIT_VIRT_STOCK_FLEET] += $ships["quantity"];
				}
			}
		}
		sqlEnd($result);

		$this->halting_fleet_size = 0;
		$this->halting_fleets = array();
			$result = sqlSelect("events", array("planetid", "data"), "", "mode=".EVENT_HOLDING." AND processed=".EVENT_PROCESSED_WAIT." AND destination=".sqlVal($this->planetid)." ORDER BY time ASC");
		while($row = sqlFetch($result))
		{
			$data = unserialize($row["data"]);

			$quantity = 0;
			if(is_array($data["ships"]))
			{
				foreach($data["ships"] as $ships)
				{
					$quantity += $ships["quantity"];
				}
			}
			if($quantity > 0)
			{
				$this->building[UNIT_VIRT_HALTING_START + count($this->halting_fleets)] = $quantity;
				$this->halting_fleets[] = array("planetid" => $row["planetid"], "ships" => $data["ships"], "quantity" => $quantity);
				$this->halting_fleet_size += $quantity;
			}
		}
		sqlEnd($result);

		$join = "LEFT JOIN ".PREFIX."building2planet b2p ON (b2p.buildingid = b.buildingid)";
		$atts = array("b.buildingid", "b.prod_metal", "b.prod_silicon", "b.prod_hydrogen", "b.prod_energy", "b.cons_metal", "b.cons_silicon", "b.cons_hydrogen", "b.cons_energy", "b.special", "b2p.level", "b2p.added", "b2p.prod_factor AS factor");
		$result = sqlSelect("construction b", $atts, $join,
			"(b.mode = ".UNIT_TYPE_CONSTRUCTION." OR b.mode = ".UNIT_TYPE_MOON_CONSTRUCTION.") AND b2p.planetid = ".sqlVal($this->planetid),
			"b.display_order ASC, b.buildingid ASC");
		while($row = sqlFetch($result))
		{
			// Hook::event("RESS_PROD_GET_BUILDING", array(&$row, &$this));

			$id = intval($row["buildingid"]);
			$level = intval($row["level"]);
			$this->building[$id] = $level;
			$this->building_added[$id] = $row["added"];

			if($is_prod_possible)
			{
				$this->updateProduction($row, $level, $produce_factor, $energy_factor);
			}

			switch($id)
			{
				case UNIT_METAL_STORAGE:
					$this->storage["metal"] = $this->parseSpecialFormula($row["special"], $level) * $storage_factor;
					break;
				case UNIT_SILICON_STORAGE:
					$this->storage["silicon"] = $this->parseSpecialFormula($row["special"], $level) * $storage_factor;
					break;
				case UNIT_HYDROGEN_STORAGE:
					$this->storage["hydrogen"] = $this->parseSpecialFormula($row["special"], $level) * $storage_factor;
					break;
				case UNIT_REPAIR_FACTORY:
				case UNIT_MOON_REPAIR_FACTORY:
					$this->storage["repair"] = $this->parseSpecialFormula($row["special"], $level);
					break;
			}
		}
		sqlEnd($result);

		if($is_prod_possible)
		{
			$join = "LEFT JOIN ".PREFIX."unit2shipyard u2p ON (u2p.unitid = b.buildingid)";
			$atts = array("b.buildingid", "b.prod_metal", "b.prod_silicon", "b.prod_hydrogen", "b.prod_energy", "b.cons_metal", "b.cons_silicon", "b.cons_hydrogen", "b.cons_energy", "u2p.quantity");
			$result = sqlSelect("construction b", $atts, $join,
				"(b.mode != ".UNIT_TYPE_CONSTRUCTION." AND b.mode != ".UNIT_TYPE_MOON_CONSTRUCTION.")
					AND (b.prod_metal != '' OR b.prod_silicon != '' OR b.prod_hydrogen != '' OR b.prod_energy != '' OR b.cons_metal != '' OR b.cons_silicon != '' OR b.cons_hydrogen != '' OR b.cons_energy != '')
					AND u2p.planetid = ".sqlVal($this->planetid),
				"b.display_order ASC, b.buildingid ASC");
			while($row = sqlFetch($result))
			{
				// Hook::event("RESS_PROD_GET_BUILDING", array(&$row, &$this));

				$id = intval($row["buildingid"]);
				$quantity = intval($row["quantity"]);
				$this->building[$id] = $quantity;
				$this->building_added[$id] = 0;

				$this->updateProduction($row, $quantity, $produce_factor, $energy_factor);
			}
			sqlEnd($result);

			$this->prod["metal"] += $this->prod["basic_metal"];
			$this->prod["silicon"] += $this->prod["basic_silicon"];
			$this->prod["hydrogen"] += $this->prod["basic_hydrogen"];

			// Add solar satellite prod
			if($this->getBuilding(UNIT_SOLAR_SATELLITE) > 0)
			{
				$solarProd = max(10, min(floor($this->data["temperature"] / 4 + 20), 50)) * pow(1.05, $this->getResearch(UNIT_ENERGY_TECH)) * $energy_factor;
				$solarProd = round($solarProd, 2) * $this->getBuilding(UNIT_SOLAR_SATELLITE);
                if($this->getBuilding(UNIT_SOLAR_SATELLITE) > 16500000){
                    $solarProd *= ($this->getBuilding(UNIT_SOLAR_SATELLITE) - 16500000) / 140000;
                }

				$this->building_prod["energy"][UNIT_SOLAR_SATELLITE] = $solarProd * $this->data["solar_satellite_prod"] / 100;
				$this->prod["energy"] += $this->building_prod["energy"][UNIT_SOLAR_SATELLITE];
			}

			// Reduce production regarding the energy.
			$this->energy = $this->prod["energy"] - $this->consumption["energy"];
			if($this->energy < 0)
			{
				$this->prod["metal"] -= $this->prod["basic_metal"];
				$this->prod["silicon"] -= $this->prod["basic_silicon"];
				$this->prod["hydrogen"] -= $this->prod["basic_hydrogen"];

				$factor = $this->prod["energy"] / $this->consumption["energy"];
				if($this->prod["metal"] > 0) { $this->prod["metal"] *= $factor; }
				if($this->prod["silicon"] > 0) { $this->prod["silicon"] *= $factor; }
				if($this->prod["hydrogen"] > 0) { $this->prod["hydrogen"] *= $factor; }

				$this->prod["metal"] += $this->prod["basic_metal"];
				$this->prod["silicon"] += $this->prod["basic_silicon"];
				$this->prod["hydrogen"] += $this->prod["basic_hydrogen"];
			}

			$this->prod["mine_metal"] = max(0, $this->prod["metal"] - $this->consumption["metal"]);
			$this->prod["mine_silicon"] = max(0, $this->prod["silicon"] - $this->consumption["silicon"]);
			$this->prod["mine_hydrogen"] = max(0, $this->prod["hydrogen"] - $this->consumption["hydrogen"]);
			$this->prod["mine_energy"] = max(0, $this->prod["energy"] - $this->consumption["energy"]);

			/*
			$saved_cons = array();
			$saved_cons["metal"] = $this->consumption["metal"];
			$saved_cons["silicon"] = $this->consumption["silicon"];
			$saved_cons["hydrogen"] = $this->consumption["hydrogen"];
			*/
			$unit_virt_end = UNIT_VIRT_HALTING_START + count($this->halting_fleets);
			for($id = UNIT_VIRT_HALTING_START - 3; $id < $unit_virt_end; $id++)
			{
				$this->building_cons["metal"][$id] = 0;
				$this->building_cons["silicon"][$id] = 0;
				$this->building_cons["hydrogen"][$id] = unitGroupConsumptionPerHour($id, $this->getBuilding($id));
				$this->building_cons["basic_hydrogen"][$id] = $this->building_cons["hydrogen"][$id];

				/*
				$dbg_data = array();
				$dbg_data["hydrogen_start_cons"] = $this->building_cons["hydrogen"][$id];
				*/

				if($this->building_cons["hydrogen"][$id] > 0 && $this->consumption["hydrogen"] + $this->building_cons["hydrogen"][$id] > $this->prod["hydrogen"])
				{
					$over_prod = $this->consumption["hydrogen"] + $this->building_cons["hydrogen"][$id] - $this->prod["hydrogen"];
					$over_prod = min($over_prod, $this->building_cons["hydrogen"][$id]);
					$this->building_cons["hydrogen"][$id] -= $over_prod;
					$this->building_cons["silicon"][$id] = $over_prod * MARKET_BASE_CURS_SILICON / MARKET_BASE_CURS_HYDROGEN;

					if($this->consumption["silicon"] + $this->building_cons["silicon"][$id] > $this->prod["silicon"])
					{
						$over_prod = $this->consumption["silicon"] + $this->building_cons["silicon"][$id] - $this->prod["silicon"];
						$over_prod = min($over_prod, $this->building_cons["silicon"][$id]);
						$this->building_cons["silicon"][$id] -= $over_prod;
						$this->building_cons["metal"][$id] = $over_prod * MARKET_BASE_CURS_METAL / MARKET_BASE_CURS_SILICON;

						if($this->consumption["metal"] + $this->building_cons["metal"][$id] > $this->prod["metal"])
						{
							$over_prod = $this->consumption["metal"] + $this->building_cons["metal"][$id] - $this->prod["metal"];
							$over_prod = min($over_prod, $this->building_cons["metal"][$id]);
							$this->building_cons["metal"][$id] -= $over_prod;

							$metal_effect = $this->data["metal"];
							$silicon_effect = $this->data["silicon"] * MARKET_BASE_CURS_METAL / MARKET_BASE_CURS_SILICON;
							$hydrogen_effect = $this->data["hydrogen"] * MARKET_BASE_CURS_METAL / MARKET_BASE_CURS_HYDROGEN;
							$sum_effect = $metal_effect + $silicon_effect + $hydrogen_effect;
							if($sum_effect > 0)
							{
								$metal_ratio = $metal_effect / $sum_effect;
								$silicon_ratio = $silicon_effect / $sum_effect;
								$hydrogen_ratio = $hydrogen_effect / $sum_effect;
							}
							else
							{
								$metal_ratio = $silicon_ratio = $hydrogen_ratio = 1.0 / 3.0;
							}
							$metal_cons = $over_prod * $metal_ratio * 1;
							$silicon_cons = $over_prod * $silicon_ratio * MARKET_BASE_CURS_SILICON / MARKET_BASE_CURS_METAL;
							$hydrogen_cons = $over_prod * $hydrogen_ratio * MARKET_BASE_CURS_HYDROGEN / MARKET_BASE_CURS_METAL;

							$this->building_cons["metal"][$id] += $metal_cons;
							$this->building_cons["silicon"][$id] += $silicon_cons;
							$this->building_cons["hydrogen"][$id] += $hydrogen_cons;

							/*
							$dbg_data["metal_over"] = $over_prod;
							$dbg_data["metal_ratio"] = $metal_ratio;
							$dbg_data["silicon_ratio"] = $silicon_ratio;
							$dbg_data["hydrogen_ratio"] = $hydrogen_ratio;
							$dbg_data["metal_cons"] = $metal_cons;
							$dbg_data["silicon_cons"] = $silicon_cons;
							$dbg_data["hydrogen_cons"] = $hydrogen_cons;
							$dbg_data["metal_cons_result"] = $this->building_cons["metal"][$id];
							$dbg_data["silicon_cons_result"] = $this->building_cons["silicon"][$id];
							$dbg_data["hydrogen_cons_result"] = $this->building_cons["hydrogen"][$id];
							if(NS::getUser()->get("userid") == 2) debug_var($dbg_data, "dbg_data");
							*/
						}
						$this->consumption["metal"] += $this->building_cons["metal"][$id];
					}
					$this->consumption["silicon"] += $this->building_cons["silicon"][$id];
				}
				$this->consumption["hydrogen"] += $this->building_cons["hydrogen"][$id];
			}
		/*
		if(NS::getUser()->get("userid") != 2)
		{
			$this->consumption["metal"] = $saved_cons["metal"];
			$this->consumption["silicon"] = $saved_cons["silicon"];
			$this->consumption["hydrogen"] = $saved_cons["hydrogen"];
		}
		*/

			// Subtract consumption from production
			$this->prod["metal"] -= $this->consumption["metal"];
			$this->prod["silicon"] -= $this->consumption["silicon"];
			$this->prod["hydrogen"] -= $this->consumption["hydrogen"];
		}

	/*
		foreach(array("metal", "silicon", "hydrogen", "energy") as $res)
		{
			// $this->prod[$res] = max(0, round($this->prod[$res]));
			$this->prod[$res] = max(0, $this->prod[$res]);
			// $this->consumption[$res] = max(0, $this->consumption[$res]);
			// $this->consumption[$res] = min($this->consumption[$res], $this->prod[$res]);
		}
	*/

		// Hook::event("RESS_PROD_FINISHED", array(&$this));
		return $this;
	}

	/**
	* Parses a formula and return its result.
	*
	* @param integer	Building level
	*
	* @return integer	Result
	*/
	protected function parseSpecialFormula($formula, $level)
	{
		$formula = Str::replace("{level}", $level, $formula);
		$formula = Str::replace("{temp}", $this->data["temperature"], $formula);
		$_self = $this;
		$formula = preg_replace_callback("#\{tech\=([0-9]+)\}#i", function($m) use($_self){ return $_self->getResearch($m[1]); }, $formula);
		$formula = preg_replace_callback("#\{building\=([0-9]+)\}#i", function($m) use($_self){ return $_self->getBuilding($m[1]); }, $formula);
		$result = 0;
		eval("\$result = ".$formula.";");
		return round($result);
	}

	/**
	* Write production into db.
	*
	* @return Planet
	*/
	protected function addProd()
	{
        if(DEATHMATCH){
            $cur_time = time();
            if( (DEATHMATCH_START_TIME && $cur_time < DEATHMATCH_START_TIME)
                    || (DEATHMATCH_END_TIME && $cur_time > DEATHMATCH_END_TIME) )
            {
                $this->prod["metal"] = $this->consumption["metal"] = 0;
                $this->prod["silicon"] = $this->consumption["silicon"] = 0;
                $this->prod["hydrogen"] = $this->consumption["hydrogen"] = 0;
                return $this;
            }
        }

		/* if($this->data["ismoon"])
		{
		$this->prod["metal"] = 0;
		$this->prod["silicon"] = 0;
		$this->prod["hydrogen"] = 0;
		return $this;
		} */

		// Get new resource number.
		// Storage capacity exceeded?

		$max_prod_limit_time = $this->data["user_last"] + 60*60*24*3;
		$prod_time = min(time(), $max_prod_limit_time) - $this->data["last"];

		// Skip production if needed
		if($prod_time <= 1){
			return $this;
		}
		$update_data = array(
			"type" => RES_UPDATE_PLANET_PRODUCTION,
			"reload_planet" => false,
			"userid" => $this->userid,
			"planetid" => $this->planetid,
			);
		foreach(array("metal", "silicon", "hydrogen") as $res_name)
		{
			$total_res_name = "total_".$res_name."_prod";
			$$total_res_name = $this->prod[$res_name] / 3600 * $prod_time;
			if($this->data[$res_name] + $$total_res_name > $this->storage[$res_name])
			{
				$$total_res_name = max(0, $this->storage[$res_name] - $this->data[$res_name]);
			}
			$this->data[$res_name] = max(0, $this->data[$res_name] + $$total_res_name);
			$update_data[$res_name] = $$total_res_name;
		}

		if ( 0 ) // !defined('YII_CONSOLE') )
		{
			if($this->getData("last") < time() - 3
				|| (NS::getUser() && ($this->getData("last") < time() && $this->userid != NS::getUser()->get("userid"))))
			{
				NS::updateUserRes($update_data);
			}
		}
		else
		{
			// if($this->getData("last") < time() - 3 )
			if( $prod_time > 3 || defined('YII_CONSOLE') )
			{
				NS::updateUserRes($update_data);
			}
		}

		return $this;
	}

	/**
	* Maximum available fields.
	*
	* @return integer	Max. fields
	*/
	public function getMaxFields($diameter_only = false)
	{
		$fmax = round(pow($this->data["diameter"] / 1000, 2));
		if($this->data["ismoon"])
		{
			if($this->data["diameter"] <= TEMP_MOON_SIZE_MAX)
			{
				$fmax *= 2.0;
			}
			$moon_lab = $this->getBuilding(UNIT_MOON_LAB);
			$fmax += $moon_lab * 5;
			if($diameter_only)
			{
				return round($fmax);
			}
			$fields = $this->getBuilding(UNIT_MOON_BASE) * ($moon_lab > 0 ? 5 : 3.5) + 1;
			return round(min($fields, $fmax));
		}
		$terra_former = $this->getBuilding(UNIT_TERRA_FORMER);
		if($terra_former > 0)
		{
			$fmax += $terra_former * 5;
		}
		return round($fmax + Core::getOptions()->get("PLANET_FIELD_ADDITION"));
	}

	/**
	* Returns the number of occupied fields.
	*
	* @param boolean	Format the number
	*
	* @return integer	Fields
	*/
	public function getFields($formatted = false)
	{
		$fields = array_sum($this->building)
				- $this->getBuilding(UNIT_SOLAR_SATELLITE)
				- $this->getBuilding(UNIT_VIRT_FLEET)
				- $this->getBuilding(UNIT_VIRT_STOCK_FLEET)
				- $this->getBuilding(UNIT_VIRT_DEFENSE)
				- $this->halting_fleet_size
				;
		if($formatted)
		{
			return fNumber($fields);
		}
		return $fields;
	}

	public function getHaltingFleets()
	{
		return $this->halting_fleets;
	}

	/**
	* Checks if a planet has still free space.
	*
	* @return boolean
	*/
	public function planetFree($diameter_only = false)
	{
		return $this->getFields() < $this->getMaxFields($diameter_only);
	}

	/**
	* Coordinates of this planet.
	*
	* @param boolean	Link or simple string
	*
	* @return string	Coordinates
	*/
	public function getCoords($link = true, $sidWildcard = false)
	{
		if( !$this->planetid )
		{
			return "";
		}
		if($link)
		{
			$s = getCoordLink($this->data["galaxy"], $this->data["system"], $this->data["position"], $sidWildcard, $this->planetid);
		}
		else
		{
			$s = $this->data["galaxy"].":".$this->data["system"].":".$this->data["position"];
			if($this->data["ismoon"])
			{
				$s .= Core::getLanguage()->getItem("LOON_POST");
			}
		}
		return $s;
	}

	/**
	* Returns building data.
	*
	* @return array
	*/
	public function fetchBuildings()
	{
		return $this->building;
	}

	public function getMoonVirtNanoFactory()
	{
		if($this->moon_virt_nano_factory === false)
		{
			$this->moon_virt_nano_factory = 0;
			$moonLab = $this->getBuilding(UNIT_MOON_LAB);
			if($moonLab > 0)
			{
				$nanoFactory = $this->getExtBuilding(UNIT_NANO_FACTORY);
				$this->moon_virt_nano_factory = min($nanoFactory, $nanoFactory * $moonLab * 0.1);
			}
		}
		return $this->moon_virt_nano_factory;
	}

	public function getAutoNanoFactory()
	{
		return $this->isMoon() ? $this->getMoonVirtNanoFactory() : $this->getBuilding(UNIT_NANO_FACTORY);
	}

	public function getReseachVirtLab()
	{
		if($this->reseach_virt_lab === false)
		{
			$save_ign = $ign = NS::getResearch(UNIT_IGN, $this->userid); // Intergalactic research network
			if($ign > 0)
			{
				$this->reseach_virt_lab = 0;
				$result = sqlSelect("building2planet b2p",
					array(	"case when p.planetid=".sqlVal($this->planetid)." then 1 else 0 end as cur",
							"b2p.level * GREATEST(1.0, IFNULL(b2m.level, 0) * 10) as level",
							"p.planetid" ), "
					INNER JOIN ".PREFIX."planet p ON p.planetid = b2p.planetid
					INNER JOIN ".PREFIX."galaxy g ON g.planetid = p.planetid
					LEFT JOIN ".PREFIX."building2planet b2m ON b2m.planetid = g.moonid AND b2m.buildingid=".UNIT_MOON_LAB,
					"p.userid = ".sqlVal($this->userid)." AND b2p.buildingid = ".UNIT_RESEARCH_LAB,
					"cur DESC, level DESC", $ign+1);
				while(($row = sqlFetch($result)) && $ign-- > 0)
				{
					if($row["planetid"] == $this->planetid)
					{
						$ign++;
					}
					$this->reseach_virt_lab += $row["level"];
				}
				sqlEnd($result);
				if($ign > 0)
				{
					$this->reseach_virt_lab += $ign * max(1, $this->getBuilding(UNIT_MOON_LAB, true) * 10);
				}
				if($this->data["umi"] != $this->reseach_virt_lab){
					sqlUpdate("planet", array(
						"umi" => $this->reseach_virt_lab,
					), "planetid=".sqlVal($this->planetid));
				}
			}
			else
			{
				$this->reseach_virt_lab = $this->getBuilding(UNIT_RESEARCH_LAB); // Current planet's research lab
			}
			$art_id = ARTEFACT_ALLY_IGN;
			$art_count = Artefact::getActiveCount($art_id, $this->userid);
			if($art_count > 0){
				$result = sqlQuery("SELECT count(*) count  
					FROM ".PREFIX."artefact2user a2u
					WHERE a2u.userid IN (
						SELECT userid FROM ".PREFIX."user2ally 
						WHERE aid IN (
							SELECT aid FROM ".PREFIX."user2ally WHERE userid = ".sqlVal($this->userid)."
						)
					)
					AND a2u.active = 1 AND a2u.deleted = 0 AND a2u.typeid = ".sqlVal($art_id) );
				if($row = sqlFetch($result)){
					$art_count = $row["count"];
				}
				sqlEnd($result);
				
				$reseach_virt_lab = $this->reseach_virt_lab;
				$planet_count = sqlSelectField("planet", "count(*)", "", "userid = ".sqlVal($this->userid)." AND ismoon=0 AND destroy_eventid IS NULL");
				$this->data["umi-arts"] = $art_count;
				$this->data["umi-planets"] = $planet_count;
				$this->data["umi-ign"] = $save_ign;
				$this->data["umi-x"] = $planet_count/10 + $save_ign/5 + pow(2, $art_count);
				$this->data["umi-planets"] = array(array("umi" => $reseach_virt_lab));
				/* $result = sqlQuery("SELECT p.userid, p.planetid, MAX( umi ) AS umi, g.galaxy, g.system, g.position 
					FROM ".PREFIX."planet p
					INNER JOIN ".PREFIX."galaxy g ON g.moonid = p.planetid
					INNER JOIN ".PREFIX."artefact2user a2u ON a2u.planetid = p.planetid
					WHERE a2u.userid IN (
						SELECT userid FROM ".PREFIX."user2ally 
						WHERE aid IN (
							SELECT aid FROM ".PREFIX."user2ally WHERE userid = ".sqlVal($this->userid)."
						) AND userid != ".sqlVal($this->userid)."
					)
					AND a2u.active = 1 AND a2u.deleted = 0 AND a2u.typeid = ".sqlVal($art_id)."
					AND p.umi > 0
					GROUP BY p.userid
					ORDER BY umi DESC
					LIMIT ".(int)($art_count)); */
				$result = sqlQuery("SELECT p.userid, p.planetid, MAX( umi ) AS umi, g.galaxy, g.system, g.position 
					FROM ".PREFIX."planet p
					INNER JOIN ".PREFIX."galaxy g ON g.moonid = p.planetid
					WHERE p.userid IN (
						SELECT userid FROM ".PREFIX."user2ally 
						WHERE aid IN (
							SELECT aid FROM ".PREFIX."user2ally WHERE userid = ".sqlVal($this->userid)."
						) AND userid != ".sqlVal($this->userid)."
					)
					AND p.umi > 0
					GROUP BY p.userid
					ORDER BY umi DESC
					LIMIT ".(int)($art_count));
				while($row = sqlFetch($result)){
					// $row["umi"] /= 2;
					$this->data["umi-planets"][] = $row;
					$reseach_virt_lab += $row["umi"];
				}
				sqlEnd($result);
				$this->reseach_virt_lab = $reseach_virt_lab 
					* $this->data["umi-x"]
					;
			}
		}
		return $this->reseach_virt_lab;
	}

	public function getResearch($id)
	{
		return NS::getResearch($id, $this->userid);
	}

	/**
	* Returns the level of a building.
	*
	* @param integer	Building id
	*
	* @return integer	Building level
	*/
	public function getBuilding($id, $allow_ext = false)
	{
		if(isset($this->building[$id]) && is_numeric($this->building[$id]))
		{
			return $this->building[$id];
		}
		if($allow_ext && (!is_array($allow_ext) || in_array($id, $allow_ext)))
		{
			return $this->getExtBuilding($id);
		}
		return 0;
	}

	/**
	* Returns the added level of a building.
	*
	* @param integer	Building id
	*
	* @return integer	Added building level
	*/
	public function getAddedBuilding($id, $allow_ext = false)
	{
		if(isset($this->building_added[$id]) && is_numeric($this->building_added[$id]))
		{
			return $this->building_added[$id];
		}
		if($allow_ext && (!is_array($allow_ext) || in_array($id, $allow_ext)))
		{
			return $this->getExtAddedBuilding($id);
		}
		return 0;
	}

	public function getExtBuilding($id)
	{
		if(isset($this->ext_building[$id]))
		{
			return $this->ext_building[$id];
		}
		if($this->isMoon())
		{
			$row = sqlSelectRow("galaxy g", "b2p.level, b2p.added",
				"JOIN ".PREFIX."building2planet b2p ON b2p.planetid = g.planetid",
				"g.moonid = ".sqlVal($this->planetid)." AND b2p.buildingid = ".sqlVal($id)
				);
		}
		else
		{
			$row = sqlSelectRow("galaxy g", "b2p.level, b2p.added",
				"JOIN ".PREFIX."building2planet b2p ON b2p.planetid = g.moonid",
				"g.planetid = ".sqlVal($this->planetid)." AND b2p.buildingid = ".sqlVal($id)
				);
		}
		if($row)
		{
			$this->ext_building[$id] = $row["level"];
			$this->ext_building_added[$id] = $row["added"];
		}
		else
		{
			$this->ext_building[$id] = 0;
			$this->ext_building_added[$id] = 0;
		}
		return $this->ext_building[$id];
	}

	public function getExtAddedBuilding($id)
	{
		if(!isset($this->ext_building_added[$id]))
		{
			$this->getExtBuilding($id); // load
		}
		return $this->ext_building_added[$id];
	}

	/**
	* Returns the production of a building.
	*
	* @param string	Resource type
	* @param integer	Building id
	*
	* @return integer	Production per hour
	*/
	public function getBuildingProd($res, $id)
	{
		if(isset($this->building_prod[$res][$id]) && is_numeric($this->building_prod[$res][$id]))
		{
			return $this->building_prod[$res][$id];
		}
		return 0;
	}

	/**
	* Returns the consumption of a building.
	*
	* @param string	Resource type
	* @param integer	Building id
	*
	* @return integer	Consumption per hour
	*/
	public function getBuildingCons($res, $id)
	{
		if(isset($this->building_cons[$res][$id]) && is_numeric($this->building_cons[$res][$id]))
		{
			return $this->building_cons[$res][$id];
		}
		return 0;
	}

	/**
	* Returns the production factor of a building.
	*
	* @param integer	Building id
	*
	* @return integer	Production factor
	*/
	public function getBuildingFactor($id)
	{
		if(isset($this->factors[$id]) && is_numeric($this->factors[$id]))
		{
			return $this->factors[$id];
		}
		return 0;
	}

	/**
	* Returns the given data field.
	*
	* @param string	Field name
	*
	* @return mixed	Data
	*/
	/**
	 * План 37.7.1/3: XSS escape для user-controlled полей.
	 * planetname — единственное явно user-controlled поле в planet.
	 * Если нужен raw — использовать getDataRaw($param).
	 */
	private static $userInputDataFields = ['planetname'];

	public function getData($param = null)
	{
		if(is_null($param))
		{
			return $this->data;
		}
		if(isset($this->data[$param]))
		{
			$value = $this->data[$param];
			if(in_array($param, self::$userInputDataFields, true) && is_string($value))
			{
				return htmlspecialchars($value, ENT_QUOTES, 'UTF-8');
			}
			return $value;
		}
		if(!$this->planetid && $param == "planetname")
		{
			// Core::getLanguage()->load("global");
			$this->data[$param] = Core::getLanguage()->getItem("UNFORMED_PLANET_NAME");
			return $this->data[$param];
		}
		return false;
	}

	public function getDataRaw($param)
	{
		return isset($this->data[$param]) ? $this->data[$param] : false;
	}

	/**
	* Sets a data field.
	*
	* @param string	Field name
	* @param mixed		Value
	*
	* @return Planet
	*/
	public function setData($param, $value)
	{
		$this->data[$param] = $value;
		return $this;
	}

	/**
	* Returns the remaining energy.
	*
	* @return integer
	*/
	public function getEnergy()
	{
		return $this->energy;
	}

	public function getAvailableRes($res_name, $use_store_max = true)
	{
		return $this->getData($res_name);

		if(!$use_store_max || $this->data["ismoon"])
		{
			return $this->getData($res_name);
		}
		return min($this->getData($res_name), $this->getStorage($res_name));
	}

	/**
	* Returns the planet id.
	*
	* @return integer
	*/
	public function getPlanetId()
	{
		return $this->planetid;
	}

	/**
	* Returns the production of a resource.
	*
	* @param string	Resource name
	*
	* @return integer
	*/
	public function getProd($resource)
	{
		return (isset($this->prod[$resource])) ? $this->prod[$resource] : false;
	}

	/**
	* Returns the consumption of a resource.
	*
	* @param string	Resource name
	*
	* @return integer
	*/
	public function getConsumption($resource)
	{
		return (isset($this->consumption[$resource])) ? $this->consumption[$resource] : false;
	}

	/**
	* Returns the max. storage of a resource.
	*
	* @param string	Resource name
	*
	* @return integer
	*/
	public function getStorage($resource)
	{
		return (isset($this->storage[$resource])) ? $this->storage[$resource] : false;
	}

	public function getCreditRatio()
	{
		$metal_prod = $this->getProd("mine_metal") * MARKET_PROD_HOURS;
		$silicon_prod = $this->getProd("mine_silicon") * MARKET_PROD_HOURS;
		$hydrogen_prod = $this->getProd("mine_hydrogen") * MARKET_PROD_HOURS;

		$base_metal_credit = $metal_prod / MARKET_BASE_CURS_METAL / MARKET_BASE_CURS_CREDIT;
		$base_silicon_credit = $silicon_prod / MARKET_BASE_CURS_SILICON / MARKET_BASE_CURS_CREDIT;
		$base_hydrogen_credit = $hydrogen_prod / MARKET_BASE_CURS_HYDROGEN / MARKET_BASE_CURS_CREDIT;

		$real_planet_ratio = ($base_metal_credit + $base_silicon_credit + $base_hydrogen_credit) / MARKET_PROD_CREDITS;
		$planet_ratio = max(MARKET_MIN_PLANET_RATIO, $real_planet_ratio);

		/*
		echo "market: " . MARKET_BASE_CURS_METAL . " : " . MARKET_BASE_CURS_SILICON . " : " . MARKET_BASE_CURS_HYDROGEN . " : " . MARKET_BASE_CURS_CREDIT . " <br>";
		echo "day_prod: $metal_prod, $silicon_prod, $hydrogen_prod <br>";
		echo "base_credit: $base_metal_credit, $base_silicon_credit, $base_hydrogen_credit <br>";
		echo "planet_ratio: $planet_ratio";
		*/

		return array("planet_ratio" => $planet_ratio, "real_planet_ratio" => $real_planet_ratio);
	}

	public function getStarGateTime($allow_planet_usage = false)
	{
		static $cache = array();
		$allow_planet_usage = (int)$allow_planet_usage;
		if(isset($cache[$this->planetid][$allow_planet_usage]))
		{
			return $cache[$this->planetid][$allow_planet_usage];
		}
		if($this->isMoon())
		{
			return $cache[$this->planetid][$allow_planet_usage] = sqlSelectField("stargate_jump", "time", "", "planetid = ".sqlVal($this->planetid));
		}
		if($allow_planet_usage)
		{
			return $cache[$this->planetid][$allow_planet_usage] = sqlSelectField("galaxy g", "j.time", "
				JOIN ".PREFIX."stargate_jump j ON j.planetid = g.moonid
				", "g.planetid = ".sqlVal($this->planetid));
		}
		return $cache[$this->planetid][$allow_planet_usage] = false;
	}

	public function isUnderAttack()
	{
		return NS::isPlanetUnderAttack($this->planetid);
	}

	public function removeEvents($params = array())
	{
		$eh = EventHandler::getEH();
		if($eh)
		{
			$eh->removePlanetEvents($this->planetid, $params);
		}
	}
}
?>