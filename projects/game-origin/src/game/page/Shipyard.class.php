<?php
/**
* Shipyard page. Shows ships and defense and allows to construct them.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class Shipyard extends Construction
{
	/**
	* If units can be build.
	*
	* @var boolean
	*/
	protected $canBuildUnits;

	protected $freeRocketFields = 0;
	protected $freeShieldFields = 0;
	protected $freeExchangeSlots = 0;

	/**
	* Constructor: Shows unit selection.
	*
	* @return void
	*/
	public function __construct()
	{
		parent::__construct();

		// debug_var(Core::getRequest(), "[Core::getRequest()]"); exit;

		if($this->unit_type == UNIT_TYPE_DEFENSE)
		{
			$this->canBuildUnits = NS::getEH()->canBuildDefenseUnits()
                    && NS::getPlanet()->getBuilding(UNIT_DEFENSE_FACTORY) > 0;

			if(NS::getPlanet()->getBuilding(UNIT_ROCKET_STATION) > 0)
			{
				$this->freeRocketFields = NS::getRocketStationSize()
					- getShipyardQuantity(UNIT_INTERCEPTOR_ROCKET) - getShipyardQuantity(UNIT_INTERPLANETARY_ROCKET)*2
					- NS::getEH()->getWorkingRockets();
			}

			$this->freeShieldFields = NS::getShieldFields()
				- getShipyardQuantity(UNIT_SMALL_SHIELD) - getShipyardQuantity(UNIT_LARGE_SHIELD)*5
				- getShipyardQuantity(UNIT_SMALL_PLANET_SHIELD)*10 - getShipyardQuantity(UNIT_LARGE_PLANET_SHIELD)*40
				- NS::getEH()->getWorkingShields();
		}
		else
		{
			$this->canBuildUnits = NS::getEH()->canBuildShipyardUnits()
                    && NS::getPlanet()->getBuilding(UNIT_SHIPYARD) > 0;
		}
		$this->freeExchangeSlots = Exchange::freeUnitSlots();

		if(!NS::getUser()->get("umode"))
		{
			$this->setPostAction("sendmission", "order")
				->addPostArg("order", null)
				->setGetAction("go", "AbortShipyard", "abortShipyard")
				->addGetArg("abortShipyard", "id")
				->setGetAction("go", "AbortDefense", "abortDefense")
				->addGetArg("abortDefense", "id")
				->setGetAction("go", "StartShipyardVIP", "startShipyardVIP")
				->addGetArg("startShipyardVIP", "id")
				->setGetAction("go", "StartDefenseVIP", "startDefenseVIP")
				->addGetArg("startDefenseVIP", "id")
				;
		}
			$this->proceedRequest();
		return;
	}

	/**
	* Index action.
	*
	* @return Shipyard
	*/
	protected function index()
	{
		Core::getLanguage()->load("info,buildings");

		$credit = NS::getUser()->get("credit");
        $observer = NS::getUser()->get("observer") || (DEATHMATCH && (!isDeathmatchStarted() || isDeathmatchEnd()));
		$user_race = NS::getUser()->get("race");

        if(NS::getUser()->get("observer")){
            Logger::addMessage('CANT_BUILD_AND_RESEARCH_DUE_OBSERVER');
        }

		$counter = 1;
		$queue = array();
		$items = array();
		$vip_image = Image::getImage("vip_queue.gif", "");
		$abort_action_name = $this->unit_type == UNIT_TYPE_FLEET ? "AbortShipyard" : "AbortDefense";
		$start_action_name = $this->unit_type == UNIT_TYPE_FLEET ? "StartShipyardVIP" : "StartDefenseVIP";

		Core::getTPL()->addHTMLHeaderFile("lib/jquery.countdown.js?".CLIENT_VERSION, "js");
		Core::getTPL()->addHTMLHeaderFile("lib/jquery.countdown-ru.js?".CLIENT_VERSION, "js");

		// If shipyard is under construction.
		// while($row = sqlFetch($result))
		$events = NS::getEH()->getShipyardModeEvents($this->eventType());
		foreach($events as &$event)
		{
			$event["real_end_time"] = $event["time"];
			if(!empty($event["data"]["new_format"]) && $event["data"]["quantity"] > 1)
			{
				$event["real_end_time"] = $event["start"] + $event["data"]["quantity"] * $event["data"]["duration"];
			}
		}
		unset($event);
		usort($events, function($a, $b){
			if($a["real_end_time"] < $b["real_end_time"]) return -1;
			if($a["real_end_time"] > $b["real_end_time"]) return 1;
			return 0;
		});

		$free_queue_size = $this->unit_type == UNIT_TYPE_FLEET ? getMaxCountListShipyard() : getMaxCountListDefense();
		$free_queue_size -= count($events);
		foreach($events as $event)
		{
			// $data = unserialize($event["data"]);
			$data = &$event["data"];

			if(empty($data["new_format"]))
			{
				continue;
			}

			$queue[$counter]["eventid"] = $event["eventid"];
			$queue[$counter]["number"] = $counter;
			$queue[$counter]["name"] = Core::getLanguage()->getItem($data["buildingname"]);
			$queue[$counter]["quantity"] = $data["quantity"];

			$abort = "";
			$vip = "";

			if(!empty($data["vip"]))
			{
				$queue[$counter]["name"] = $vip_image . "&nbsp;" . $queue[$counter]["name"];
			};
			$abort = Link::get("game.php/{$abort_action_name}/".$data["batchkey"], "<span class='false'>".Core::getLanguage()->getItem("ABORT")."</span>");

			if($event["start"] > time() && empty($event["data"]["vip"]))
			{
				$credit_req = getCreditImmStartShipyard($data["quantity"], $this->unit_type);
				$vip = sprintf(Core::getLanguage()->getItem("START_IMM"), fNumber($credit_req));
				if($credit >= $credit_req)
				{
					$vip = Link::get("game.php/{$start_action_name}/".$data["batchkey"], $vip);
				}
				else
				{
					$vip = "<span class='false2'>{$vip}</span>";
				}
			}

			$continue = "-"; // Link::get($this_page, Core::getLanguage()->getItem("CONTINUE"));
			if($event["start"] > time())
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
				// $event_end_time = $event["start"] + $event["data"]["quantity"] * $event["data"]["duration"];
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
			// $queue_units[$data["buildingid"]] = 1; // array("level" => $data["level"]);
			$counter++;
		}
		// sqlEnd($result);
		Core::getTPL()->assign("can_add_to_queue", $free_queue_size > 0 && $this->canBuildUnits);

		// $result = sqlSelect("construction", array("buildingid", "name", "basic_metal", "basic_silicon", "basic_hydrogen", "basic_energy"), "", "mode = ".sqlVal($this->unit_type), "display_order ASC, buildingid ASC");
		$result = sqlSelect("construction c",
			array("buildingid", "mode", "name", "basic_metal", "basic_silicon", "basic_hydrogen", "basic_energy", "race", "us.quantity", "us.damaged", "us.shell_percent"),
			"LEFT JOIN ".PREFIX."unit2shipyard us ON us.unitid=c.buildingid AND us.planetid=".sqlPlanet(),
			"c.mode=".sqlVal($this->unit_type).(!isAdmin() ? " AND c.test = 0" : ""),
			"c.display_order ASC, c.buildingid ASC");
		while($row = sqlFetch($result))
		{
			$id = $row["buildingid"];

			$can_build = !$observer && NS::checkRequirements($id);
			if($row["race"] != $user_race && !$can_build)
			{
				continue;
			}

			$can_show_all = $this->canShowAllUnits();
			Core::getTPL()->assign("show_all_units", $can_show_all);

			if($can_build || $can_show_all || $row["quantity"] > 0)
			{
				// Hook::event("SHIPYARD_ITEM", array(&$row));

				// Common building data
				$items[$id]["name"] = Link::get("game.php/UnitInfo/".$id, Core::getLanguage()->getItem($row["name"]));
				$items[$id]["description"] = Core::getLanguage()->getItem($row["name"]."_DESC");
				$items[$id]["image"] = Link::get("game.php/UnitInfo/".$id, Image::getImage(getUnitImage($row["name"]), Core::getLanguage()->getItem($row["name"])));
				$items[$id]["edit"] = Link::get("game.php/EditUnit/".$id, "[".Core::getLanguage()->getItem("EDIT")."]");

				$quantity = $row["quantity"]; // getShipyardQuantity($id);
				$items[$id]["quantity_num"] = $quantity;
				$items[$id]["quantity"] = getUnitQuantityStr($row, array("splitter" => "<br />", "bracket" => false)); // fNumber($quantity);
				// $items[$id]["quantity"] = sprintf(Core::getLanguage()->getItem("SHIPS_EXIST"), fNumber($items[$id]["quantity"]));

				if ($can_build)
				{
					// Required resources
					$this->setRequieredResources(0, $row);

					//Construction output
					if(NS::getUser()->get("umode") || $free_queue_size <= 0) // || isset($queue_units[$id]))
					{
						$items[$id]["construct"] = "";
					}
					else if(($id == UNIT_SMALL_SHIELD && $this->freeShieldFields < 1) || ($id == UNIT_LARGE_SHIELD && $this->freeShieldFields < 2)
								|| ($id == UNIT_SMALL_PLANET_SHIELD && $this->freeShieldFields < 4) || ($id == UNIT_LARGE_PLANET_SHIELD && $this->freeShieldFields < 8)
							)
					{
						$items[$id]["construct"] = Core::getLanguage()->getItem("NO_SHIELDS_FIELDS");
					}
					else if(($id == UNIT_INTERCEPTOR_ROCKET && $this->freeRocketFields < 1) || ($id == UNIT_INTERPLANETARY_ROCKET && $this->freeRocketFields < 2))
					{
						$items[$id]["construct"] = Core::getLanguage()->getItem("ROCKET_STATION_FULL");
					}
					else if (($id == UNIT_EXCH_SUPPORT_RANGE || $id == UNIT_EXCH_SUPPORT_SLOT)
						&& $this->freeExchangeSlots < 1)
					{
						$items[$id]["construct"] = Core::getLanguage()->getItem("NO_EXCHANGE_SLOTS");
					}
					else if($this->checkResources() && $this->canBuildUnits && !NS::getUser()->get("umode"))
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
					}
					$time = NS::getBuildingTime($this->requiredMetal, $this->requiredSilicon, $this->unit_type);
					$items[$id]["productiontime"] = getTimeTerm($time);
					$items[$id]["can_build"] = true;
				}
				else
				{
					$items[$id]["construct"] = "";
					$items[$id]["required_constructions"] = NS::requiremtentsList($id);
					$items[$id]["can_build"] = false;
				}

				// Hook::event("SHOW_SHIPYARD_UNIT_LAST", array($row, &$items[$id]));
			}
		}
		sqlEnd($result);

        $can_build = true;
		if($this->unit_type == UNIT_TYPE_FLEET)
		{
			if(NS::getPlanet()->getBuilding(UNIT_SHIPYARD) < 1)
			{
                if($queue || $items){
                    Logger::addMessage("SHIPYARD_REQUIRED");
                }else{
                    Logger::dieMessage("SHIPYARD_REQUIRED");
                }
                $can_build = false;
			}
			Core::getTPL()->assign("shipyard", Core::getLanguage()->getItem("SHIP_CONSTRUCTION"));
		}
		else
		{
			if(NS::getPlanet()->getBuilding(UNIT_DEFENSE_FACTORY) < 1)
			{
                if($queue || $items){
                    Logger::addMessage("DEFENSE_FACTORY_REQUIRED");
                }else{
                    Logger::dieMessage("DEFENSE_FACTORY_REQUIRED");
                }
			}
			Core::getTPL()->assign("shipyard", Core::getLanguage()->getItem("DEFENSE"));
		}

        if(!$can_build){
            foreach($items as &$item){
                // $item["can_build"] = false;
                $item["construct"] = "";
            }
            unset($item);
        }
		Core::getTPL()->addLoop("imagePacks", getImgPacks());
		Core::getTPL()->addLoop("events", $queue);
		Core::getTPL()->addLoop("shipyard", $items);
		Core::getTPL()->display("shipyard");
		return $this;
	}

	/**
	* Starts an event for building units.
	*
	* @param array		_POST data
	*
	* @return Shipyard
	*/
	protected function order($post)
	{
		// debug_var($post, "[order]");

		// Vacation enabled?
		if(NS::getUser()->get("umode"))
		{
			throw new GenericException("Your account is still in vacation mode.");
		}
        if(NS::getUser()->get("observer") || (DEATHMATCH && (!isDeathmatchStarted() || isDeathmatchEnd()))){
            throw new GenericException("The observer mode enabled");
        }
		if(!$this->canBuildUnits)
		{
			throw new GenericException("Shipyard or nano factory in progress.");
		}
		if($this->unit_type != UNIT_TYPE_FLEET && $this->unit_type != UNIT_TYPE_DEFENSE)
		{
			throw new GenericException("Error building mode: $this->unit_type");
			// doHeaderRedirection($this->main_page);
			// return $this;
		}
		$events = NS::getEH()->getShipyardModeEvents($this->eventType());

		$free_queue_size = $this->unit_type == UNIT_TYPE_FLEET ? getMaxCountListShipyard() : getMaxCountListDefense();
		$free_queue_size -= count($events);

		if($free_queue_size <= 0)
		{
			//throw new GenericException("There is no empty slots in queue.");
			return $this->index();
		}

		$queue_end_time = time();
		foreach($events as $event)
		{
			if(empty($event["data"]["vip"]))
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

			$id	= max(0, (int)$id);
			$quantity = max(0, (int)$quantity);

			if(!$id || !$quantity)
			{
				continue;
			}

			// Load shipyard data
			$row = getConstructionDesc($id, $this->unit_type);
			if(!$row)
			{
				throw new GenericException("Unkown shipyard :(");
			}

			// Check for requirements
			if(!NS::checkRequirements($id))
			{
				throw new GenericException("You cannot build this.");
			}

			if($this->unit_type == UNIT_TYPE_FLEET)
			{
				$quantity = min($quantity, getMaxCountOrderShipyard());
			}
			else
			{
				$quantity = min($quantity, getMaxCountOrderDefense());
			}

			// Decrease quantity, if necessary
			if($id == UNIT_SMALL_SHIELD){
				$quantity = floor(min($quantity, $this->freeShieldFields / 1));
			}else if($id == UNIT_LARGE_SHIELD){
				$quantity = floor(min($quantity, $this->freeShieldFields / 2));
			}else if($id == UNIT_SMALL_PLANET_SHIELD){
				$quantity = floor(min($quantity, $this->freeShieldFields / 4));
			}else if($id == UNIT_LARGE_PLANET_SHIELD){
				$quantity = floor(min($quantity, $this->freeShieldFields / 8));
			}

			// Decrease quantity, if necessary
			if($id == UNIT_INTERCEPTOR_ROCKET){
				$quantity = floor(min($quantity, $this->freeRocketFields / 1));
			}else if($id == UNIT_INTERPLANETARY_ROCKET){
				$quantity = floor(min($quantity, $this->freeRocketFields / 2));
			}

			if ($id == UNIT_EXCH_SUPPORT_RANGE || $id == UNIT_EXCH_SUPPORT_SLOT)
			{
				$quantity = min($quantity, $this->freeExchangeSlots);
			}

			$quantity = floor($quantity);
			if($quantity < 1)
			{
				continue;
			}

			// Hook::event("UNIT_ORDER_START", array(&$row, &$quantity));

			foreach($events as $event)
			{
				// $event["data"] = unserialize($event["data"]);
				if($event["start"] == time() && $event["data"]["buildingid"] == $id
					&& empty($event["data"]["paid"])) // already in
				{
					Logger::dieMessage("TOO_MANY_REQUESTS");
					doHeaderRedirection($this->main_page, false);
					return $this;
				}
			}
			$start_time = $queue_end_time;

			// Check resources
			$required["metal"] = $row["basic_metal"] * $quantity;
			$required["silicon"] = $row["basic_silicon"] * $quantity;
			$required["hydrogen"] = $row["basic_hydrogen"] * $quantity;
			$required["energy"] = $row["basic_energy"] * $quantity;

			if($required["metal"] > NS::getPlanet()->getData("metal")
				|| $required["silicon"] > NS::getPlanet()->getData("silicon")
				|| $required["hydrogen"] > NS::getPlanet()->getData("hydrogen")
				|| $required["energy"] > NS::getPlanet()->getEnergy())
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
				$data["basic_metal"] = $row["basic_metal"];
				$data["basic_silicon"] = $row["basic_silicon"];
				$data["basic_hydrogen"] = $row["basic_hydrogen"];
				$data["basic_energy"] = $row["basic_energy"];
				$data["metal"] = $required["metal"];
				$data["silicon"] = $required["silicon"];
				$data["hydrogen"] = $required["hydrogen"];
				$data["energy"] = $required["energy"];
				$data["duration"] = NS::getBuildingTime($row["basic_metal"], $row["basic_silicon"], $this->unit_type);
				$data["points"] = round(($row["basic_metal"] + $row["basic_silicon"] + $row["basic_hydrogen"]) * RES_TO_UNIT_POINTS, POINTS_PRECISION);
				$data["quantity"] = $quantity;
				$data["buildingid"] = $id;
				$data["buildingname"] = $row["name"];
				$data["batchkey"] = time() . mt_rand(1000, 9999);

				// Hook::event("UPGRADE_ORDER_LAST", array($row, &$data));

				// debug_var($data, "[order]"); exit;

				//Add in list or start immediately
				NS::getEH()->addEvent($this->eventType(), $start_time + NS::getUnitsBuildTime($data),
					NS::getUser()->get("curplanet"), NS::getUser()->get("userid"), null, $data, null,
					$start_time);

				$queue_end_time += $data["duration"] * $data["quantity"];
				$free_queue_size--;
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

	protected function startShipyardVIP($batchkey)
	{
		// $this->unit_type = UNIT_TYPE_FLEET;
		return $this->startEventVIP($batchkey);
	}

	protected function startDefenseVIP($batchkey)
	{
		// $this->unit_type = UNIT_TYPE_DEFENSE;
		return $this->startEventVIP($batchkey);
	}

	protected function startEventVIP($batchkey)
	{
		// $batchkey is a string

		NS::getEH()->startConstructionEventVIP($batchkey, $this->unit_type);
		doHeaderRedirection($this->main_page, false);
		return $this;

		// Hook::event("ABORT_SHIPYARD", array(&$this, $this->unit_type));
	}

	// ===========================================================

	protected function abortShipyard($batchkey)
	{
		// $this->unit_type = UNIT_TYPE_FLEET;
		return $this->abortEvent($batchkey);
	}

	protected function abortDefense($batchkey)
	{
		// $this->unit_type = UNIT_TYPE_DEFENSE;
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