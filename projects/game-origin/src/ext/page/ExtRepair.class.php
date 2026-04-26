<?php
/**
* Repair page. Shows ships and defense and allows to repair them.
*
*
*
*/

if(!defined("RECIPE_ROOT_DIR")) { die("Hacking attempt detected."); }

class ExtRepair extends Construction
{
	/**
	* If units can be build.
	*
	* @var boolean
	*/
	protected $canRepairUnits;

	/**
	* Constructor: Shows unit selection.
	*
	* @return void
	*/
	public function __construct()
	{
		parent::__construct();

		// debug_var(Core::getRequest(), "[Core::getRequest()]"); exit;

		$this->canRepairUnits = NS::getEH()->canRepairUnits();

		if(!Core::getUser()->get("umode"))
		{
			if($this->isRepair())
			{
				$this
					->setPostAction("sendmission", "order")
					->addPostArg("order", null)
					->setGetAction("go", "AbortRepair", "abortRepair")
					->addGetArg("abortRepair", "id")
					->setGetAction("go", "StartRepairVIP", "startRepairVIP")
					->addGetArg("startRepairVIP", "id");
			}
			else
			{
				$this
					->setPostAction("sendmission", "order")
					->addPostArg("order", null)
					->setGetAction("go", "AbortDisassemble", "abortDisassemble")
					->addGetArg("abortDisassemble", "id")
					->setGetAction("go", "StartDisassembleVIP", "startDisassembleVIP")
					->addGetArg("startDisassembleVIP", "id");
			}
		}
//		try
//		{
			$this->proceedRequest();
//		}
//		catch(Exception $e)
//		{
//			$e->printError();
//		}
	}

	protected function isRepair()
	{
		return $this->unit_type == UNIT_TYPE_REPAIR;
	}

	protected function setRepairUnitRequirements($row)
	{
		$this->setRequieredResources(0, $row);

		$repair_struct_add_k = $row["mode"] == UNIT_TYPE_FLEET ? 0.1 : 0.05;
		$buildTimeMetalAdd = $this->requiredMetal * $repair_struct_add_k;
		$buildTimeSiliconAdd = $this->requiredSilicon * $repair_struct_add_k;

		$struct_scale = 0.1 * (100 - $row["shell_percent"]) / 100.0;
		$this->requiredMetal = ceil($this->requiredMetal * $struct_scale / 10) * 10;
		$this->requiredSilicon = ceil($this->requiredSilicon * $struct_scale / 10) * 10;
		$this->requiredHydrogen = ceil($this->requiredHydrogen * $struct_scale / 10) * 10;
		$this->requiredEnergy = ceil($this->requiredEnergy * $struct_scale / 10) * 10;

		if(!OXSAR_RELEASED)
		{
			$this->requiredTime = 15;
			return;
		}
		$this->requiredTime = NS::getBuildingTime($this->requiredMetal + $buildTimeMetalAdd, $this->requiredSilicon + $buildTimeSiliconAdd, $row["mode"]);
	}

	protected function setDisassembleUnitRequirements($row)
	{
		$this->setRequieredResources(0, $row);

		$repair_struct_add_k = 0.1; // $row["mode"] == UNIT_TYPE_FLEET ? 0.1 : 0.05;
		$buildTimeMetalAdd = $this->requiredMetal * $repair_struct_add_k;
		$buildTimeSiliconAdd = $this->requiredSilicon * $repair_struct_add_k;

		$struct_scale = 0.9;
		$this->returnMetal = ceil($this->requiredMetal * $struct_scale / 10) * 10;
		$this->returnSilicon = ceil($this->requiredSilicon * $struct_scale / 10) * 10;
		$this->returnHydrogen = 0; // ceil($this->requiredHydrogen * $struct_scale / 10) * 10;
		// $this->returnEnergy = 0; // ceil($this->requiredEnergy * $struct_scale / 10) * 10;

		$struct_scale = 0.2;
		$this->requiredMetal = ceil($this->requiredMetal * $struct_scale / 10) * 10;
		$this->requiredSilicon = ceil($this->requiredSilicon * $struct_scale / 10) * 10;
		$this->requiredHydrogen = ceil($this->requiredHydrogen * $struct_scale / 10) * 10;
		$this->requiredEnergy = ceil($this->requiredEnergy * $struct_scale / 10) * 10;

		$this->earnMetal = $this->returnMetal - $this->requiredMetal;
		$this->earnSilicon = $this->returnSilicon - $this->requiredSilicon;
		$this->earnHydrogen = $this->returnHydrogen - $this->requiredHydrogen;
		// $this->earnEnergy = $this->returnEnergy - $this->requiredEnergy;

		if(!OXSAR_RELEASED)
		{
			$this->requiredTime = 15;
			return;
		}
		$this->requiredTime = NS::getBuildingTime($buildTimeMetalAdd, $buildTimeSiliconAdd, $row["mode"]);
	}

	/**
	* Index action.
	*
	* @return Repair
	*/
	protected function index()
	{
		Core::getLanguage()->load("info,buildings");

		$credit = NS::getUser()->get("credit");
        $observer = NS::getUser()->get("observer") || (DEATHMATCH && (!isDeathmatchStarted() || isDeathmatchEnd()));

		$counter = 1;
		$queue = array();
		$items = array();
		$vip_image = Image::getImage("vip_queue.gif", "");

		Core::getTPL()->addHTMLHeaderFile("lib/jquery.countdown.js?".CLIENT_VERSION, "js");
		Core::getTPL()->addHTMLHeaderFile("lib/jquery.countdown-ru.js?".CLIENT_VERSION, "js");

		// If unit is under repair.
		$repair_units = array();
		$events = NS::getEH()->getRepairEvents();
		foreach($events as &$event)
		{
			$event["real_end_time"] = $event["time"];
			if(!empty($event["data"]["new_format"]) && $event["data"]["quantity"] > 1)
			{
				$event["real_end_time"] = $event["start"] + $event["data"]["quantity"] * $event["data"]["duration"];
			}
			$repair_units[$event["data"]["buildingid"]] = 1;
		}
		unset($event);
		function compare_repair_events($a, $b)
		{
			if($a["real_end_time"] < $b["real_end_time"]) return -1;
			if($a["real_end_time"] > $b["real_end_time"]) return 1;
			return 0;
		}
		usort($events, "compare_repair_events");

		$free_queue_size = getMaxCountListRepair() - count($events);

		foreach($events as $event)
		{
			$data = &$event["data"];
			if(empty($data["new_format"]))
			{
				continue;
			}

			$queue[$counter]["eventid"] = $event["eventid"];
			$queue[$counter]["number"] = $counter;
			$queue[$counter]["name"] = Core::getLanguage()->getItem($data["buildingname"]);
			if($event["mode"] == EVENT_REPAIR)
			{
				$queue[$counter]["quantity"] = getUnitQuantityStr(array("damaged" => $data["quantity"], "shell_percent" => $data["shell_percent"]), array("bracket" => false, "no_quantity" => true));
			}
			else
			{
				$queue[$counter]["quantity"] = getUnitQuantityStr(array("quantity" => $data["quantity"]), array("bracket" => false, "no_quantity" => true));
			}
			if(!empty($data["vip"]))
			{
				$queue[$counter]["name"] = $vip_image . "&nbsp;" . $queue[$counter]["name"];
			};
			$queue[$counter]["name"] =
				"[" .
				Core::getLanguage()->getItem(
					$event["mode"] == EVENT_REPAIR ? "QUEUE_REPAIR_PREFIX" : "QUEUE_DISASSEMBLE_PREFIX"
				) .
				"] " .
				$queue[$counter]["name"];
			$abortCmd = $event["mode"] == EVENT_REPAIR ? "AbortRepair" : "AbortDisassemble";
			$abort = Link::get(
				"game.php/{$abortCmd}/".$data["batchkey"],
				"<span class='false'>".Core::getLanguage()->getItem("ABORT")."</span>"
			);
			$vip = "";
			if($event["start"] > time() && empty($event["data"]["vip"]))
			{
				if($event["mode"] == EVENT_REPAIR)
				{
					$credit_req = getCreditImmStartRepair($data["quantity"]);
				}
				else
				{
					$credit_req = getCreditImmStartDisassemble($data["quantity"]);
				}
				$vip = sprintf(Core::getLanguage()->getItem("START_IMM"), fNumber($credit_req));
				if($credit > $credit_req)
				{
					$startImmCmd = $event["mode"] == EVENT_REPAIR ? "StartRepairVIP" : "StartDisassembleVIP";
					$vip = Link::get("game.php/{$startImmCmd}/".$data["batchkey"], $vip);
				}
				else
				{
					$vip = "<span class='false2'>{$vip}</span>";
				}
			}
			$continue = "-";
			if( $event["start"] > time() )
			{
				$timeleft = max(1, $event["start"] - time());
				$text = Core::getLanguage()->getItem("START_IN");
				$continue = Link::get($this_page, Core::getLanguage()->getItem("STARTED"));
				$js = "<script type='text/javascript'>
					$(function () {
						$('#queueCountDown{$counter}').countdown({until: {$timeleft}, compact: true, onExpiry: function() {
							$('#queueItem{$counter}').html('{$continue}');
						}});
				});
				</script>
					<span id='queueItem" . $counter . "'>{$text}<br><span id='queueCountDown" . $counter . "'>". getTimeTerm($timeleft)."</span><br />{$abort}</span>";
					$queue[$counter]["cancel_link"] = $js;
			}
			elseif( (round(($event_end_time = $event["start"] + $event["data"]["quantity"] * $event["data"]["duration"]) - time())) > 0 )
			{
				$timeleft = max(1, round($event_end_time - time()));
				$queue[$counter]["event_start"] = $event["start"];
				$queue[$counter]["event_end"] = $event_end_time;
				// $queue[$counter]["event_timeleft"] = max(0, $event_end_time - time());
				$queue[$counter]["event_percent_timeout"] = $event_end_time > $event["start"] ? ceil(($event_end_time - $event["start"]) * 1000 / 100.0) : 1000*10;
				$queue[$counter]["event_pb_value"] = $event_end_time > $event["start"] ? max(1, floor((time() - $event["start"]) * 100 / ($event_end_time - $event["start"]))) : 0;
				$js = "<script type='text/javascript'>
					$(function () {
						$('#queueCountDown{$counter}').countdown({until: {$timeleft}, compact: true, onExpiry: function() {
							$('#queueItem{$counter}').html('{$continue}');
						}});
				});
				</script>
					<span id='queueItem" . $counter . "'><span id='queueCountDown" . $counter . "'>". getTimeTerm($timeleft)."</span><br />{$abort}</span>";
					$queue[$counter]["cancel_link"] = $js;
			}
			else
			{
				$queue[$counter]["cancel_link"] = (
					(defined('TW'))
						? Core::getLang()->getItem('COUNTDOWN_TW')
						: Core::getLang()->getItem('COUNTDOWN_WAIT')
				);
			}
			$queue[$counter]["vip_link"] = $vip;
			$counter++;
		}

		$repair_storage	= NS::getRepairStorage();
		$repair_usage	= NS::getEH()->getRepairUsage();
		$repair_free	= $repair_storage - $repair_usage;

		$repair_building_id = NS::getPlanet()->isMoon() ? UNIT_MOON_REPAIR_FACTORY : UNIT_REPAIR_FACTORY;
		$repair_building_level = NS::getPlanet()->getBuilding($repair_building_id);
		$row = sqlSelectRow("construction", "*", "", "buildingid=".sqlVal($repair_building_id));

		Core::getTPL()->assign("construction_level", $repair_building_level);
		Core::getTPL()->assign("construction_name", Core::getLanguage()->getItem($this->isRepair() ? "MENU_REPAIR" : "MENU_DISASSEMBLE"));
		Core::getTPL()->assign("construction_image", Link::get("game.php/ConstructionInfo/".$row["buildingid"], Image::getImage(getUnitImage($row["name"]), Core::getLanguage()->getItem($row["name"]))));
		Core::getTPL()->assign("construction_description", Core::getLanguage()->getItem($row["name"]."_DESC"));
		Core::getTPL()->assign("construction_storage", fNumber($repair_storage));
		Core::getTPL()->assign("construction_used", fNumber($repair_usage));
		Core::getTPL()->assign("construction_free", fNumber($repair_free));
		Core::getTPL()->assign("construction_free_percent", $repair_storage > 0 ? min(100, round($repair_free * 100 / $repair_storage)) : 100);

		$temp = array("storage" => $repair_storage, "used" => $repair_usage, "free" => $repair_free);

		$is_fleet_started = false;
		$is_defense_started = false;

		if( $this->isRepair() )
		{
			$result = sqlSelect("construction c",
				array("buildingid", "mode", "name", "basic_metal", "basic_silicon", "basic_hydrogen", "basic_energy", "us.quantity", "us.damaged", "us.shell_percent"),
				"JOIN ".PREFIX."unit2shipyard us ON us.unitid=c.buildingid AND us.planetid=".sqlPlanet(),
				"us.planetid=".sqlPlanet()." AND (us.damaged > 0 OR c.buildingid in (".sqlArray(array_keys($repair_units))."))",
				"c.mode ASC, c.display_order ASC, c.buildingid ASC");
		}
		else
		{
			$result = sqlSelect("construction c",
				array("buildingid", "mode", "name", "basic_metal", "basic_silicon", "basic_hydrogen", "basic_energy", "us.quantity", "us.damaged", "us.shell_percent"),
				"JOIN ".PREFIX."unit2shipyard us ON us.unitid=c.buildingid AND us.planetid=".sqlPlanet(),
				"us.planetid=".sqlPlanet(),
				"c.mode ASC, c.display_order ASC, c.buildingid ASC");
		}
		while($row = sqlFetch($result))
		{
			$id = $row["buildingid"];
			{
				// Common building data
				$items[$id]["name"] = Link::get("game.php/UnitInfo/".$id, Core::getLanguage()->getItem($row["name"]));
				$items[$id]["description"] = Core::getLanguage()->getItem($row["name"]."_DESC");
				$items[$id]["image"] = Link::get("game.php/UnitInfo/".$id, Image::getImage(getUnitImage($row["name"]), Core::getLanguage()->getItem($row["name"])));
				$items[$id]["edit"] = Link::get("game.php/EditUnit/".$id, "[".Core::getLanguage()->getItem("EDIT")."]");

				$items[$id]["quantity"] = getUnitQuantityStr($row, array("splitter" => "<br />", "bracket" => false));
				// Required resources
				if($this->isRepair())
				{
					$this->setRepairUnitRequirements($row);
				}
				else
				{
					$this->setDisassembleUnitRequirements($row);
				}
				$items[$id]["max_dock_units"]		= getDockUnitsCapacity($row, $repair_free);
				$items[$id]["dock_capacity"]		= getDockUnitsCapacity($row, $repair_storage);
				$items[$id]["unit_fields"]			= getUnitFields($row);
				$items[$id]["no_free_repair_fields"]= $items[$id]["unit_fields"] <= $repair_free ? "" : $repair_free - $items[$id]["unit_fields"];
				$items[$id]["can_build"]			= !$observer && ($this->isRepair() ? NS::checkRequirements($id, null, null, true /*array(UNIT_SHIPYARD, UNIT_DEFENSE_FACTORY)*/) : true);
				//Construction output
				if( Core::getUser()->get("umode")
					|| ($this->isRepair() && $row["damaged"] < 1)
					|| $free_queue_size < 1 || $items[$id]["max_dock_units"] < 1
				)
				{
					$items[$id]["construct"] = "";
				}
				else if(!$items[$id]["can_build"])
				{
					$items[$id]["construct"] = Core::getLanguage()->getItem($this->isRepair() ? "REPAIR_NOT_AVAILABLE" : "DISASSEMBLE_NOT_AVAILABLE");
				}
				else if( $this->checkResources() && $this->canRepairUnits )
				{
					$items[$id]["construct"] = "<input type='text' name='".$id."' value='0' size='3' maxlength='".MAX_BUILDING_ORDER_UNITS_GRADE."' class='center' />";
				}
				else
				{
					$items[$id]["construct"] = "";
				}
				// End upgrade output

				// Data output
				foreach(array("metal", "silicon", "hydrogen", "energy") as $res_name)
				{
					$available_res = $res_name == "energy" ? NS::getPlanet()->getEnergy() : NS::getPlanet()->getAvailableRes($res_name);
					$required_field = "required" . ucfirst($res_name);
					$required_res = $this->$required_field;
					$items[$id][$res_name."_required"] = $required_res > 0 ? fNumber($required_res) : "";
					$items[$id][$res_name."_notavailable"] = $required_res > 0 && $available_res < $required_res ? fNumber($available_res - $required_res) : "";

					if(!$this->isRepair())
					{
						$field = "earn" . ucfirst($res_name);
						$items[$id][$res_name."_earned"] = fNumber($this->$field);
					}
				}
				$items[$id]["productiontime"] = getTimeTerm($this->requiredTime);

				if($row["mode"] == UNIT_TYPE_DEFENSE && !$is_defense_started)
				{
					$is_defense_started = true;
					$items[$id]["defense_started"] = $is_fleet_started;
				}
				if($row["mode"] == UNIT_TYPE_FLEET && !$is_fleet_started)
				{
					$is_fleet_started = true;
					$items[$id]["defense_started"] = $is_defense_started;
				}

				$temp[$id] = $items[$id]["unit_fields"]; // $items[$id];
			}
		}
		sqlEnd($result);
		if(!$repair_building_level)
		{
			Logger::dieMessage("REPAIR_REQUIRED");
		}
		Core::getTPL()->addLoop("imagePacks", getImgPacks());
		Core::getTPL()->addLoop("events", $queue);
		Core::getTPL()->addLoop("items", $items);
		Core::getTPL()->display($this->isRepair() ? "repair" : "disassemble");
		return $this;
	}

	/**
	* Starts an event for repair units.
	*
	* @param array		_POST data
	*/
	protected function order($post)
	{
		// Vacation enabled?
		if(Core::getUser()->get("umode"))
		{
			throw new GenericException("Your account is still in vacation mode.");
		}
		if(!$this->canRepairUnits || NS::getUser()->get("observer") || (DEATHMATCH && (!isDeathmatchStarted() || isDeathmatchEnd())))
		{
			throw new GenericException("You can't repair units.");
		}

		if( !NS::isFirstRun( "ExtRepair::order:" . md5(serialize($post)) ) )
		{
			throw new GenericException("There are too many queries.");
		}

		if(NS::isPlanetUnderAttack())
		{
			Logger::dieMessage('PLANET_UNDER_ATTACK');
		}

		$events = NS::getEH()->getRepairEvents();

		$free_queue_size = getMaxCountListRepair() - count($events);

		if($free_queue_size <= 0)
		{
			throw new GenericException("There is no empty slots in queue.");
		}

		$repair_storage	= NS::getRepairStorage();
		$repair_usage	= NS::getEH()->getRepairUsage();
		$repair_free	= $repair_storage - $repair_usage;
		$queue_end_time	= time();

		foreach($events as $event)
		{
			if(/*$event["start"] <= time() &&*/ empty($event["data"]["vip"]))
			{
				$end_time = $event["time"];
				if(!empty($event["data"]["new_format"]) && $event["data"]["quantity"] > 0)
				{
					$end_time = $event["start"] + $event["data"]["quantity"] * $event["data"]["duration"];
				}
				$queue_end_time = max($queue_end_time, $end_time);
			}
		}

		foreach($post as $id => $quantity)
		{
			if($free_queue_size <= 0)
			{
				break;
			}

			$id			= max(0, (int)$id);
			$quantity	= max(0, (int)$quantity);
			if(!$id || !$quantity)
			{
				continue;
			}
			// Load repair data
			if( $this->isRepair() )
			{
				$row = sqlSelectRow(
					"construction c",
					array(
						"buildingid", "mode",
						"name", "basic_metal",
						"basic_silicon", "basic_hydrogen",
						"basic_energy", "us.quantity",
						"us.damaged", "us.shell_percent"
					),
					"JOIN ".PREFIX."unit2shipyard us ON us.unitid=c.buildingid",
					"c.buildingid = ".sqlVal($id)." AND us.planetid=".sqlPlanet()." AND us.damaged > 0"
					);
			}
			else
			{
				$row = sqlSelectRow(
					"construction c",
					array(
						"buildingid", "mode",
						"name", "basic_metal",
						"basic_silicon", "basic_hydrogen",
						"basic_energy", "us.quantity",
						"us.damaged", "us.shell_percent"
					),
					"JOIN ".PREFIX."unit2shipyard us ON us.unitid=c.buildingid",
					"c.buildingid = ".sqlVal($id)." AND us.planetid=".sqlPlanet()
				);
			}
			if(!$row)
			{
				throw new GenericException("Unkown unit :(");
			}

			// Check for requirements
			if( $this->isRepair() )
			{
				if(!NS::checkRequirements($id, null, null, true /*array(UNIT_SHIPYARD, UNIT_DEFENSE_FACTORY)*/))
				{
					throw new GenericException("You cannot repair this.");
				}
				$quantity = min($quantity, $row["damaged"]);
			}

			$quantity = min($quantity, getDockUnitsCapacity($row, $repair_free), $row['quantity']);

			if($quantity < 1)
			{
				continue;
			}

			//Get current events
			$start_time = 0;
			foreach($events as $event)
			{
				// $event["data"] = unserialize($event["data"]);
				if(
					$event["start"] == time()
					&& $event["data"]["buildingid"] == $id
					&& empty($event["data"]["paid"])
				) // already in
				{
					Logger::dieMessage("TOO_MANY_REQUESTS");
					doHeaderRedirection($this->main_page, false);
					return $this;
				}
			}
			$start_time = $queue_end_time;

			// Check resources
			$points		= round(($row["basic_metal"] + $row["basic_silicon"] + $row["basic_hydrogen"]) * RES_TO_UNIT_POINTS, POINTS_PRECISION);
			$unit_fields= getUnitFields($row);

			if( $this->isRepair() )
			{
				$this->setRepairUnitRequirements($row);
			}
			else
			{
				$this->setDisassembleUnitRequirements($row);
			}
			$row["basic_metal"] = $this->requiredMetal;
			$row["basic_silicon"] = $this->requiredSilicon;
			$row["basic_hydrogen"] = $this->requiredHydrogen;
			$row["basic_energy"] = $this->requiredEnergy;

			$required["metal"] = $row["basic_metal"] * $quantity;
			$required["silicon"] = $row["basic_silicon"] * $quantity;
			$required["hydrogen"] = $row["basic_hydrogen"] * $quantity;
			$required["energy"] = $row["basic_energy"] * $quantity;

			if(
				$required["metal"] > NS::getPlanet()->getData("metal")
				|| $required["silicon"] > NS::getPlanet()->getData("silicon")
				|| $required["hydrogen"] > NS::getPlanet()->getData("hydrogen")
				|| $required["energy"] > NS::getPlanet()->getEnergy()
			)
			{
				$q1 = $row["basic_metal"] > 0 ? floor(NS::getPlanet()->getData("metal") / $row["basic_metal"]) : $quantity;
				$q2 = $row["basic_silicon"] > 0 ? floor(NS::getPlanet()->getData("silicon") / $row["basic_silicon"]) : $quantity;
				$q3 = $row["basic_hydrogen"] > 0 ? floor(NS::getPlanet()->getData("hydrogen") / $row["basic_hydrogen"]) : $quantity;
				$q4 = $row["basic_energy"] > 0 ? floor(NS::getPlanet()->getEnergy() / $row["basic_energy"]) : $quantity;

				$quantity = min($q1, $q2, $q3, $q4);

				$required["metal"] = $row["basic_metal"] * $quantity;
				$required["silicon"] = $row["basic_silicon"] * $quantity;
				$required["hydrogen"] = $row["basic_hydrogen"] * $quantity;
				$required["energy"] = $row["basic_energy"] * $quantity;
			}

			if($quantity > 0)
			{
				$data["new_format"] = 2;
				$data["unit_fields"] = $unit_fields;
				$data["repair_usage"] = $unit_fields * $quantity;
				$data["basic_metal"] = $row["basic_metal"];
				$data["basic_silicon"] = $row["basic_silicon"];
				$data["basic_hydrogen"] = $row["basic_hydrogen"];
				$data["basic_energy"] = $row["basic_energy"];
				$data["metal"] = $required["metal"];
				$data["silicon"] = $required["silicon"];
				$data["hydrogen"] = $required["hydrogen"];
				$data["energy"] = $required["energy"];
				$data["earn_metal"] = $this->returnMetal;
				$data["earn_silicon"] = $this->returnSilicon;
				$data["earn_hydrogen"] = $this->returnHydrogen;
				$data["duration"] = $this->requiredTime;
				$data["points"] = $points;
				$data["quantity"] = $quantity;
				$data["shell_percent"] = $row["shell_percent"];
				$data["damaged"] = min($quantity, $row["damaged"]);
				$data["buildingid"] = $id;
				$data["buildingtype"] = $row["mode"];
				$data["buildingname"] = $row["name"];
				$data["batchkey"] = time() . mt_rand(1000, 9999);
				//Add in list or start immediately
				NS::getEH()->addEvent(
					$this->eventType(),
					$start_time + NS::getUnitsBuildTime($data),
					Core::getUser()->get("curplanet"),
					Core::getUser()->get("userid"),
					null,
					$data,
					null,
					$start_time
				);

				$queue_end_time += $data["duration"] * $data["quantity"];
				$free_queue_size--;
				$repair_free -= $data["repair_usage"];
			}
			else
			{
				// don't throw exception
				// throw new GenericException("Not enough resources to build this.");
			}
		}
		doHeaderRedirection($this->main_page, false);
		return $this;
	}

	// ===========================================================

	protected function startRepairVIP($batchkey)
	{
		// $this->unit_type = 3;
		return $this->startEventVIP($batchkey);
	}

	protected function startDisassembleVIP($batchkey)
	{
		return $this->startEventVIP($batchkey);
	}

	protected function startDefenseVIP($batchkey)
	{
		// $this->unit_type = 4;
		return $this->startEventVIP($batchkey);
	}

	protected function startEventVIP($batchkey)
	{
		// $batchkey is a string

		NS::getEH()->startConstructionEventVIP($batchkey, $this->unit_type);
		doHeaderRedirection($this->main_page, false);
		return $this;

		// Hook::event("ABORT_REPAIR", array(&$this, $this->unit_type));
	}

	// ===========================================================

	protected function abortRepair($batchkey)
	{
		return $this->abortEvent($batchkey);
	}

	protected function abortDisassemble($batchkey)
	{
		return $this->abortEvent($batchkey);
	}

	protected function abortDefense($batchkey)
	{
		// $this->unit_type = 4;
		return $this->abortEvent($batchkey);
	}

	protected function abortEvent($batchkey)
	{
		// $batchkey is a string

		NS::getEH()->abortConstructionEvent($batchkey, $this->unit_type);
		doHeaderRedirection($this->main_page, false);
		return $this;
	}
}

?>