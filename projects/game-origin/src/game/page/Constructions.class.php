<?php
/**
* Construction & builings page.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class Constructions extends Construction
{
	/**
	* Displays list of available buildings.
	*
	* @return void
	*/
	public function __construct()
	{
		parent::__construct();
		if(!NS::getUser()->get("umode") && Core::getRequest()->getGET("id") > 0)
		{
			$this->setGetAction("go", "UpgradeConstruction", "upgradeConstruction")
				->addGetArg("upgradeConstruction", "id")
				->setGetAction("go", "AbortConstruction", "abort")
				->addGetArg("abort", "id")
				->setGetAction("go", "DemolishConstruction", "demolish")
				->addGetArg("demolish", "id")
				->setGetAction("go", "UpgradeConstructionImm", "upgradeConstructionVIP")
				->addGetArg("upgradeConstructionVIP", "id")
				->setGetAction("go", "ConstructionInfo", "constructionInfo")
				->addGetArg("constructionInfo", "id")
                ;
			;
		}
		$this->proceedRequest();
		return;
	}

    protected function constructionInfo($id)
    {
        return $this->index($id);
    }

    /**
	* Index action.
	*
	* @return Constructions
	*/
	protected function index($id = null)
	{
        $view_id = max(0, (int)$id);

		Core::getLanguage()->load("info,buildings,Resource");

		$events = NS::getEH()->getBuildingEvents();

        // CVarDumper::dump(array(isDeathmatchStarted(), isDeathmatchEnd()), 10, 1); exit;

		$credit = NS::getUser()->get("credit");
        $observer = NS::getUser()->get("observer") || (DEATHMATCH && (!isDeathmatchStarted() || isDeathmatchEnd()));

        if(NS::getUser()->get("observer")){
            Logger::addMessage('CANT_BUILD_AND_RESEARCH_DUE_OBSERVER');
        }

		$counter = 1;
		$queue = array();
		$items = array();
		$vip_image = Image::getImage("vip_queue.gif", "");

		Core::getTPL()->addHTMLHeaderFile("lib/jquery.countdown.js?".CLIENT_VERSION, "js");
		Core::getTPL()->addHTMLHeaderFile("lib/jquery.countdown-ru.js?".CLIENT_VERSION, "js");

		$free_queue_size = getMaxCountListConstructions() - count($events);

		$queue_levels = array();
		foreach($events as $event)
		{
			$id = $event["data"]["buildingid"];
			if(!isset($queue_levels[$id]))
			{
				$queue_levels[$id]["min"] = $event["data"]["level"];
				$queue_levels[$id]["max"] = $event["data"]["level"];
			}
			else
			{
				$queue_levels[$id]["min"] = min($event["data"]["level"], $queue_levels[$id]["min"]);
				$queue_levels[$id]["max"] = max($event["data"]["level"], $queue_levels[$id]["max"]);
			}
		}

		// If building is under construction.
		foreach($events as $event)
		{
			$data = &$event["data"];

			$queue[$counter]["eventid"] = $event["eventid"];
			$queue[$counter]["number"] = $counter;
			$queue[$counter]["name"] = Core::getLanguage()->getItem($data["buildingname"]);
			$queue[$counter]["level"] = $data["level"];

			if(empty($data["batchkey"]))
			{
				$queue[$counter]["name"] .= " [".Core::getLanguage()->getItem("OLD_QUEUE_ITEM")."]";
			}
			else if(!empty($data["vip"]))
			{
				$queue[$counter]["name"] = $vip_image . "&nbsp;" . $queue[$counter]["name"];
			}
			if($data["level"] < NS::getPlanet()->getBuilding($data["buildingid"]) - NS::getPlanet()->getAddedBuilding($data["buildingid"]))
			{
				$queue[$counter]["name"] = "[".Core::getLanguage()->getItem("DEMOLISH_CONSTRUCTION")."] ".$queue[$counter]["name"];
			}

			$abort = "";
			$vip = "";

			if($event["data"]["level"] == $queue_levels[$data["buildingid"]]["max"]
			&& !empty($data["batchkey"]))
			{
				$abort = Link::get("game.php/AbortConstruction/".$data["batchkey"], "<span class='false'>".Core::getLanguage()->getItem("ABORT")."</span>");
			}

			if($event["start"] > time() && empty($event["data"]["vip"])
				&& $event["data"]["level"] == $queue_levels[$event["data"]["buildingid"]]["min"]
                && !empty($data["batchkey"]))
			{
				$credit_req = getCreditImmStartConstructions($data["level"]);
				$vip = sprintf(Core::getLanguage()->getItem("START_IMM"), fNumber($credit_req));
				if($credit >= $credit_req)
				{
					$vip = Link::get("game.php/UpgradeConstructionImm/".$data["batchkey"], $vip);
				}
				else
				{
					$vip = "<span class='false2'>{$vip}</span>";
				}
			}

			$continue = "-";
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
			elseif( $event["time"] > time() )
			{
				$timeleft = max(1, $event["time"] - time());
				$queue[$counter]["event_start"] = $event["start"];
				$queue[$counter]["event_end"] = $event["time"];
				// $queue[$counter]["event_timeleft"] = max(0, $event["time"] - time());
				$queue[$counter]["event_percent_timeout"] = $event["time"] > $event["start"] ? ceil(($event["time"] - $event["start"]) * 1000 / 100.0) : 1000*10;
				$queue[$counter]["event_pb_value"] = $event["time"] > $event["start"] ? max(1, floor((time() - $event["start"]) * 100 / ($event["time"] - $event["start"]))) : 0;
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
		$mobile_skin = (int)isMobileSkin();
		$result = sqlSelect("construction", array("buildingid", "name", "mode", "demolish",
			"basic_metal", "basic_silicon", "basic_hydrogen", "basic_energy", "basic_credit", "basic_points",
			"charge_metal", "charge_silicon", "charge_hydrogen", "charge_energy", "charge_credit", "charge_points",
			"prod_metal", "prod_silicon", "prod_hydrogen", "prod_energy", "special",
			"cons_metal", "cons_silicon", "cons_hydrogen", "cons_energy",
            ), "",
			"mode = ".sqlVal($this->unit_type)
                . (!isAdmin() ? " AND test = 0" : "")
                . ($view_id > 0 ? " AND buildingid=".sqlVal($view_id) : ""),
            "display_order ASC, buildingid ASC");
		while($row = sqlFetch($result))
		{
			$id = $row["buildingid"];
			$can_build = !$observer && NS::checkRequirements($id);

			$can_show_all = $view_id > 0 || $this->canShowAllUnits();
			Core::getTPL()->assign("show_all_units", $can_show_all);

			if($can_build || $can_show_all || NS::getPlanet()->getBuilding($id) || NS::getPlanet()->getAddedBuilding($id))
			{
                if($view_id > 0){
                    Core::getTPL()->assign("info_id", $view_id);

                    $info = $this->getChartData($row);
                    foreach($info as $info_key => $info_value){
                        if(is_array($info_value)){
                            Core::getTPL()->addLoop($info_key, $info_value);
                        }elseif($info_key == 'prod_invert'){
                            Core::getTPL()->assign($info_key, $info_value);
                        }else{
                            Core::getTPL()->assign("ext_".$info_key, $info_value);
                        }
                    }
                    $max_level_text = Core::getLanguage()->getItemWith("MAX_LEVEL", array("level" => $info['max_level']));
                }

				// Common building data
				$items[$id]["name"] = Link::get("game.php/ConstructionInfo/".$id, Core::getLanguage()->getItem($row["name"]));
				$items[$id]["description"] = Core::getLanguage()->getItem($row["name"].($view_id > 0 ? "_FULL_DESC" : "_DESC")).($view_id > 0 ? "<p />".$max_level_text : "");
				$items[$id]["image"] = Link::get("game.php/ConstructionInfo/".$id, Image::getImage(getUnitImage($row["name"]), Core::getLanguage()->getItem($row["name"])));
				$items[$id]["edit"] = Link::get("game.php/EditConstruction/".$id, "[".Core::getLanguage()->getItem("EDIT")."]");

				$added_level = NS::getPlanet()->getAddedBuilding($id);
				$next_level = NS::getPlanet()->getBuilding($id) - $added_level + 1;
				if(isset($queue_levels[$id]["max"]))
				{
					$next_level = max($next_level, $queue_levels[$id]["max"] + 1);
				}
				$this->setRequieredResources($next_level, $row);
				$items[$id]["level"] = NS::getPlanet()->getBuilding($id); // $next_level-1;
				$items[$id]["added_level"] = $added_level;

				Core::getTPL()->assign( "free_queue_size", ( $free_queue_size <= 0 ) );

				if ($can_build)
				{
					// Upgrade output
					$is_storage_unit = $id == UNIT_METAL_STORAGE || $id == UNIT_SILICON_STORAGE || $id == UNIT_HYDROGEN_STORAGE;

					if(NS::getUser()->get("umode")
						|| (!NS::getPlanet()->planetFree()
								&& ($id != UNIT_TERRA_FORMER) // || (NS::getPlanet()->getMaxFields(true) - NS::getPlanet()->getFields() <= 1))
								&& ($id != UNIT_MOON_LAB) // || (NS::getPlanet()->getMaxFields(true) - NS::getPlanet()->getFields() <= 1))
								&& ($id != UNIT_MOON_BASE || (NS::getPlanet()->getMaxFields(true) - NS::getPlanet()->getFields() <= 1))
								)
						|| $free_queue_size <= 0
						|| (isset($queue_levels[$id]["min"]) && NS::getPlanet()->getBuilding($id) > $queue_levels[$id]["min"])
						)
					{
						$items[$id]["upgrade"] = "";
					}
					else if($next_level > MAX_BUILDING_LEVEL || (isset($GLOBALS["MAX_UNIT_LEVELS"][$id]) && $next_level > $GLOBALS["MAX_UNIT_LEVELS"][$id]))
					{
						$items[$id]["upgrade"] = "<span class='false'>".Core::getLanguage()->getItem("BUILDING_MAX_LEVEL_REACHED")."</span>";
					}
					else if(!NS::canUpgradeConstruction($id))
					{
						$items[$id]["upgrade"] = "<span class='false'>".Core::getLanguage()->getItem("BUILDING_AT_WORK")."</span>";
					}
					else if($id == UNIT_EXCHANGE && !EXCH_ENABLED)
					{
						$items[$id]["upgrade"] = "<span class='false'>".Core::getLanguage()->getItem("BUILDING_NOT_AVAIBLE_EXCHANGE_DISABLED")."</span>";
					}
					else if($this->checkResources(!$is_storage_unit))
					{
						if($next_level > 1)
						{
							$name_item = $counter >= 2 ? "IN_LIST_LEVEL" : "UPGRADE_TO_LEVEL";
							$name_item = Core::getLanguage()->getItem($name_item)." ".$next_level;
							// $items[$id]["upgrade"] = Link::get("game.php/UpgradeConstruction/".$id, Core::getLanguage()->getItem($name_item)." ".$next_level, "", "true");
						}
						else
						{
							$name_item = $counter >= 2 ? "IN_LIST" : "BUILD";
							$name_item = Core::getLanguage()->getItem($name_item);
							// $items[$id]["upgrade"] = Link::get("game.php/UpgradeConstruction/".$id, Core::getLanguage()->getItem($name_item), "", "true");
						}
						if($this->requiredCredit > 0)
						{
							$confirm_message = Core::getLanguage()->getItemWith("CONFIRM_BUILDING_USE_CREDIT", array(
								"name" => Core::getLanguage()->getItem($row["name"])." ".$next_level,
								"credit" => fNumber($this->requiredCredit),
								));
							$url = socialUrl(RELATIVE_URL."game.php/UpgradeConstruction/".$id);
							/* $items[$id]["upgrade"] = CHtml::link($confirm_message, "#", array(
								"onlick" =>
								)); */
							$items[$id]["upgrade"] = Link::get("#", $name_item, "", "true", "onclick=\"confirm_dialog('$confirm_message', {id: '$id', url: '$url', mobile: $mobile_skin}); return false\"");
						}
						else
						{
							$items[$id]["upgrade"] = Link::get("game.php/UpgradeConstruction/".$id, $name_item, "", "true");
						}
					}
					else
					{
						if($next_level > 1)
						{
							$name_item = $counter >= 2 ? "IN_LIST_LEVEL" : "UPGRADE_TO_LEVEL";
							$items[$id]["upgrade"] = "<span class='false'>".Core::getLanguage()->getItem($name_item)." ".$next_level."</span>";
						}
						else
						{
							$name_item = $counter >= 2 ? "IN_LIST" : "BUILD";
							$items[$id]["upgrade"] = "<span class='false'>".Core::getLanguage()->getItem($name_item)."</span>";
						}
					}
					// End upgrade output

					// Data output
					foreach(array("metal", "silicon", "hydrogen", "energy", "credit", "points") as $res_name)
					{
						switch($res_name)
						{
						case "energy":
							$available_res = NS::getPlanet()->getEnergy();
							break;
						case "credit":
							$available_res = NS::getUser()->get("credit");
							break;
						case "points":
							$available_res = NS::getUser()->get("points");
							break;
						default:
							$available_res = NS::getPlanet()->getAvailableRes($res_name, !$is_storage_unit);
						}
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
					$items[$id]["upgrade"] = "";
					$items[$id]["required_constructions"] = NS::requiremtentsList($id);
					$items[$id]["can_build"] = false;
				}
			}
		}
		sqlEnd($result);
		Core::getTPL()->addLoop("imagePacks", getImgPacks());
		Core::getTPL()->addLoop("events", $queue);
		Core::getTPL()->addLoop("constructions", $items);
		Core::getTPL()->display("constructions");
		return $this;
	}

	/**
	* Check for sufficient resources and start to upgrade building.
	*
	* @param integer Building id to upgrade
	*
	* @return Constructions
	*/
	protected function upgradeConstruction($id)
	{
		$id = max(0, (int)$id);

		Core::getLanguage()->load("info,buildings");

		$events = NS::getEH()->getBuildingEvents();

		$row = getConstructionDesc($id, $this->unit_type);
		if(!$row || NS::getUser()->get("observer") || (DEATHMATCH && (!isDeathmatchStarted() || isDeathmatchEnd())))
		{
			throw new GenericException("Unkown building $id, mode: $this->unit_type :(");
		}

		if(!NS::canUpgradeConstruction($id))
		{
			throw new GenericException("Do not mess with the url.");
		}
        if($id == UNIT_EXCHANGE && !EXCH_ENABLED){
            throw new GenericException("Do not mess with the url.");
        }

		// Check fields
		if(!NS::getPlanet()->planetFree()
								&& ($id != UNIT_TERRA_FORMER) // || (NS::getPlanet()->getMaxFields(true) - NS::getPlanet()->getFields() <= 1))
								&& ($id != UNIT_MOON_LAB)
								&& ($id != UNIT_MOON_BASE || (NS::getPlanet()->getMaxFields(true) - NS::getPlanet()->getFields() <= 1))
								)
		// if(!NS::getPlanet()->planetFree())
		{
			Logger::dieMessage("PLANET_FULLY_DEVELOPED");
		}

		$free_queue_size = getMaxCountListConstructions() - count($events);
		if($free_queue_size <= 0)
		{
			Logger::dieMessage("TOO_MANY_REQUESTS");
			doHeaderRedirection($this->main_page, false);
			return $this;
		}

		// Check for requirements
		if(!NS::checkRequirements($id))
		{
			throw new GenericException("You does not fulfil the requirements to build this.");
		}

		$next_level = NS::getPlanet()->getBuilding($id) - NS::getPlanet()->getAddedBuilding($id) + 1;
		$start_time = time();

		foreach($events as $event)
		{
			if($event["data"]["buildingid"] == $id)
			{
				if($event["start"] == time())
				{
					Logger::dieMessage("TOO_MANY_REQUESTS");
					doHeaderRedirection($this->main_page, false);
					return $this;
				}
				if($event["data"]["level"] < $next_level-1) // demolish in queue
				{
					throw new GenericException("Do not mess with the url.");
				}
				$next_level = max($next_level, $event["data"]["level"] + 1);
			}
		}

		if($next_level > MAX_BUILDING_LEVEL || (isset($GLOBALS["MAX_UNIT_LEVELS"][$id]) && $next_level > $GLOBALS["MAX_UNIT_LEVELS"][$id]))
		// if($next_level > MAX_BUILDING_LEVEL)
		{
			throw new GenericException("You have already reached maximum builing level.");
		}

		foreach($events as $event)
		{
			if($event["start"] == time() && $event["data"]["buildingid"] == $id) // already in
			{
				Logger::dieMessage("TOO_MANY_REQUESTS");
				doHeaderRedirection($this->main_page, false);
				return $this;
			}

			//Find the end time of last event
			$end_time = $event["time"];
			if($end_time > $start_time && empty($event["data"]["vip"]))
			{
				$start_time = $end_time;
			}
		}

		$this->setRequieredResources($next_level, $row);

		$is_storage_unit = $id == UNIT_METAL_STORAGE || $id == UNIT_SILICON_STORAGE || $id == UNIT_HYDROGEN_STORAGE;

		// Check resources
		if($this->checkResources(!$is_storage_unit))
		{
			$data["metal"] 		= $this->requiredMetal;
			$data["silicon"] 	= $this->requiredSilicon;
			$data["hydrogen"] 	= $this->requiredHydrogen;
			$data["energy"] 	= $this->requiredEnergy;
			$data["credit"] 	= $this->requiredCredit;
			// $data["points"] 	= $this->requiredPoints;

			$is_moon = NS::getPlanet()->getData("ismoon");
			$galaxy = NS::getPlanet()->getData("galaxy");

			$time = NS::getBuildingTime($data["metal"], $data["silicon"], $this->unit_type);
			$data["level"] = $next_level;
			$data["buildingid"] = $id;
			$data["buildingname"] = $row["name"];
			$data["batchkey"] = time() . mt_rand(1000, 9999);
			//Add in list or start immediately
			NS::getEH()->addEvent($this->eventType(), ($start_time > 0 ? $start_time : time()) + $time,
				NS::getUser()->get("curplanet"), NS::getUser()->get("userid"), null, $data, null,
				$start_time > 0 ? $start_time : null);

			doHeaderRedirection($this->main_page, false);
		}
		else
		{
			throw new GenericException("Not enough resources to build this.");
		}
		return $this;
	}

	/**
	* Check for sufficient credit and immediately start to upgrade building.
	*
	* @param integer Building id to upgrade
	*
	* @return Constructions
	*/
	protected function upgradeConstructionVIP($batchkey)
	{
		// $batchkey is a string
		NS::getEH()->startConstructionEventVIP($batchkey, $this->unit_type);
		doHeaderRedirection($this->main_page, false);
		return $this;
	}

	/**
	* Aborts the current building event.
	*
	* @param integer Building id
	*
	* @return Constructions
	*/
	protected function abort($batchkey)
	{
		// $batchkey is a string

		NS::getEH()->abortConstructionEvent($batchkey, $this->unit_type);

		doHeaderRedirection($this->main_page, false);
		return $this;
	}

	/**
	* Demolish a building ...
	*
	* @param integer Building id
	*
	* @return Constructions
	*/
	protected function demolish($id)
	{
		$events = NS::getEH()->getBuildingEvents();

		$row = getConstructionDesc($id, $this->unit_type);
		if(!$row)
		{
			throw new GenericException("Unkown building $id, mode: $this->unit_type :(");
		}

		$factor = floatval($row["demolish"]);
		if($factor <= 0.0)
		{
			throw new GenericException("The building cannot be demolished.");
		}

		$level = NS::getPlanet()->getBuilding($id) - NS::getPlanet()->getAddedBuilding($id);
		if($level < 1)
		{
			throw new GenericException("Wut?");
		}

		$start_time = 0;
		foreach($events as $event)
		{
			if($event["data"]["buildingid"] == $id)
			{
				throw new GenericException("You have to delete queue item for this build.");
			}

			if($event["start"] == time() && $event["data"]["buildingid"] == $id) // already in
			{
				Logger::dieMessage("TOO_MANY_REQUESTS");
				doHeaderRedirection($this->main_page, false);
				return $this;
			}

			//Find the end time of last event
			$end_time = $event["time"];
			if($end_time > $start_time)
			{
				$start_time = $end_time;
			}
		}
		$data["level"] = $level - 1;
		if($row["basic_metal"] > 0) { $data["metal"] = parseChargeFormula($row["charge_metal"], $row["basic_metal"], $level); }
		if($row["basic_silicon"] > 0) { $data["silicon"] = parseChargeFormula($row["charge_silicon"], $row["basic_silicon"], $level); }
		if($row["basic_hydrogen"] > 0) { $data["hydrogen"] = parseChargeFormula($row["charge_hydrogen"], $row["basic_hydrogen"], $level); }
		$data["metal"] = (1 / $factor) * $data["metal"];
		$data["silicon"] = (1 / $factor) * $data["silicon"];
		$data["hydrogen"] = (1 / $factor) * $data["hydrogen"];
		if($data["metal"] <= NS::getPlanet()->getData("metal") && $data["silicon"] <= NS::getPlanet()->getData("silicon") && $data["hydrogen"] <= NS::getPlanet()->getData("hydrogen"))
		{
			$data["buildingid"] = $id;
			$data["buildingname"] = $row["name"];
			$data["batchkey"] = time() . mt_rand(1000, 9999);
			$time = NS::getBuildingTime($data["metal"], $data["silicon"], $this->unit_type);
			NS::getEH()->addEvent(EVENT_DEMOLISH_CONSTRUCTION, ($start_time > 0 ? $start_time : time()) + $time,
				NS::getUser()->get("curplanet"), NS::getUser()->get("userid"), null, $data, null,
				$start_time > 0 ? $start_time : null);

			doHeaderRedirection($this->main_page, false);
		}
		else
		{
			throw new GenericException("Not enough resources to build this.");
		}
		return $this;
	}
}

?>