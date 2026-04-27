<?php
/**
* Research page. Shows research list.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class Research extends Construction
{
	/**
	* Current research event.
	*
	* @var mixed
	*/
	// protected $event = false;

	/**
	* Displays list of available researches.
	*
	* @return void
	*/
	public function __construct()
	{
		parent::__construct();
		// Get event
		// $this->events = NS::getEH()->getResearchEvents();
		// debug_var($this->events, "[events]");

		if(!NS::getUser()->get("umode") && Core::getRequest()->getGET("id") > 0)
		{
			$this->setGetAction("go", "UpgradeResearch", "upgradeResearch")
				->setGetAction("go", "AbortResearch", "abort")
				->addGetArg("upgradeResearch", "id")
				->addGetArg("abort", "id")
				->setGetAction("go", "UpgradeResearchImm", "upgradeResearchVIP")
				->addGetArg("upgradeResearchVIP", "id")
				->setGetAction("go", "ResearchInfo", "researchInfo")
				->addGetArg("researchInfo", "id")
                ;
		}

//		try {
			$this->proceedRequest();
//		} catch(Exception $e) {
//			$e->printError();
//		}
		return;
	}

    protected function researchInfo($id)
    {
        return $this->index($id);
    }

	/**
	* Index action.
	*
	* @return Research
	*/
	protected function index($id = null)
	{
        $view_id = max(0, (int)$id);

		Core::getLanguage()->load("info,buildings,Resource");

		$events = NS::getEH()->getResearchEvents();

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

		Core::getTPL()->addHTMLHeaderFile("lib/jquery.countdown.js?".CLIENT_VERSION, "js");
		Core::getTPL()->addHTMLHeaderFile("lib/jquery.countdown-ru.js?".CLIENT_VERSION, "js");

		$free_queue_size = getMaxCountListResearch() - count($events);

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

		// If research is in work
		// while($row = sqlFetch($result))
		foreach($events as $event)
		{
			// $data = unserialize($row["data"]);
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

			$abort = "";
			$vip = "";

			if($event["data"]["level"] == $queue_levels[$event["data"]["buildingid"]]["max"]
					&& !empty($data["batchkey"]))
			{
				$abort = Link::get("game.php/AbortResearch/".$data["batchkey"], "<span class='false'>".Core::getLanguage()->getItem("ABORT")."</span>");
			}

			if($event["start"] > time() && empty($event["data"]["vip"])
				&& $event["data"]["level"] == $queue_levels[$event["data"]["buildingid"]]["min"]
				&& !empty($data["batchkey"]))
			{
				$credit_req = getCreditImmStartResearch($data["level"]);
				$vip = sprintf(Core::getLanguage()->getItem("START_IMM"), fNumber($credit_req));
				if($credit >= $credit_req)
				{
					$vip = Link::get("game.php/UpgradeResearchImm/".$data["batchkey"], $vip);
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
			elseif($event["time"] > time())
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
			// $current_levels[ $data["buildingid"] ] = array("level" => $data["level"]);
			$counter++;
		}
		// sqlEnd($result);

		$is_lab_exist = (bool)(NS::getPlanet()->getBuilding(UNIT_RESEARCH_LAB) || NS::getPlanet()->getBuilding(UNIT_MOON_LAB));

		$result = sqlSelect("construction", array("buildingid", "name", "mode", "demolish",
            "basic_metal", "basic_silicon", "basic_hydrogen", "basic_energy",
            "charge_metal", "charge_silicon", "charge_hydrogen", "charge_energy",
			"prod_metal", "prod_silicon", "prod_hydrogen", "prod_energy", "special",
			"cons_metal", "cons_silicon", "cons_hydrogen", "cons_energy",
            "race"),
			"",
            "mode = ".UNIT_TYPE_RESEARCH
                . (!isAdmin() ? " AND test = 0" : "")
                . ($view_id > 0 ? " AND buildingid=".sqlVal($view_id) : "")
                . (!EXPEDITION_ENABLED ? " AND buildingid!=".sqlVal(UNIT_EXPO_TECH) : "")
            ,
            "display_order ASC, buildingid ASC");
		while($row = sqlFetch($result))
		{
			$id = $row["buildingid"];
			/* if($id == UNIT_EXPO_TECH)
			{
			continue;
			} */

			$can_build = !$observer && NS::checkRequirements($id, null, null, $is_lab_exist) && $row["race"] == $user_race;
			if($view_id <= 0 && ($row["race"] != $user_race && NS::getResearch($id) == 0))
			{
				continue;
			}

			$can_show_all = $view_id > 0 || $this->canShowAllUnits();
			Core::getTPL()->assign("show_all_units", $can_show_all);

			if($can_build || $can_show_all || NS::getResearch($id) ||  NS::getAddedResearch($id))
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

				// Common research data
				$items[$id]["name"] = Link::get("game.php/ResearchInfo/".$id, Core::getLanguage()->getItem($row["name"]));
				$items[$id]["description"] = Core::getLanguage()->getItem($row["name"].($view_id > 0 ? "_FULL_DESC" : "_DESC")).($view_id > 0 ? "<p />".$max_level_text : "");
				$items[$id]["image"] = Link::get("game.php/ResearchInfo/".$id, Image::getImage(getUnitImage($row["name"]), Core::getLanguage()->getItem($row["name"])));
				$items[$id]["edit"] = Link::get("game.php/EditConstruction/".$id, "[".Core::getLanguage()->getItem("EDIT")."]");

				$added_level = NS::getAddedResearch($id);
				$next_level = NS::getResearch($id) - $added_level + 1;
				if(isset($queue_levels[$id]["max"]))
				{
					$next_level = max($next_level, $queue_levels[$id]["max"] + 1);
				}
				$this->setRequieredResources($next_level, $row);
				$items[$id]["level"] = NS::getResearch($id); // $next_level-1;
				$items[$id]["added_level"] = $added_level;

				if ($can_build)
				{
					// Upgrade output
					if(NS::getUser()->get("umode") || !NS::getEH()->canReasearch() || $free_queue_size <= 0)
					{
						$items[$id]["upgrade"] = "";
					}
					else if($next_level > MAX_RESEARCH_LEVEL || (isset($GLOBALS["MAX_UNIT_LEVELS"][$id]) && $next_level > $GLOBALS["MAX_UNIT_LEVELS"][$id]))
					{
						$items[$id]["upgrade"] = "<span class='false'>".Core::getLanguage()->getItem("RESEARCH_MAX_LEVEL_REACHED")."</span>";
					}
					else if($this->checkResources())
					{
						if($next_level > 1)
						{
							$name_item = $counter >= 2 ? "IN_LIST_LEVEL" : "RESEARCH_OF_LEVEL";
							$items[$id]["upgrade"] = Link::get("game.php/UpgradeResearch/".$id, Core::getLanguage()->getItem($name_item)." ".$next_level, "", "true");
						}
						else
						{
							$name_item = $counter >= 2 ? "IN_LIST" : "RESEARCH";
							$items[$id]["upgrade"] = Link::get("game.php/UpgradeResearch/".$id, Core::getLanguage()->getItem($name_item), "", "true");
						}
					}
					else
					{
						if($next_level > 1)
						{
							$name_item = $counter >= 2 ? "IN_LIST_LEVEL" : "RESEARCH_OF_LEVEL";
							$items[$id]["upgrade"] = "<span class='false'>".Core::getLanguage()->getItem($name_item)." ".$next_level."</span>";
						}
						else
						{
							$name_item = $counter >= 2 ? "IN_LIST" : "RESEARCH";
							$items[$id]["upgrade"] = "<span class='false'>".Core::getLanguage()->getItem($name_item)."</span>";
						}
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
					$items[$id]["upgrade"] = "";
					$items[$id]["required_constructions"] = NS::requiremtentsList($id);
					$items[$id]["can_build"] = false;
				}
			}
		}
		sqlEnd($result);

		if(count($queue) > 0)
		{
			if(!$is_lab_exist)
			{
				Logger::addMessage("RESEARCH_LAB_REQUIRED");
			}
		}
		else if(count($items) == 0 || !$is_lab_exist)
		{
			Logger::dieMessage("RESEARCH_LAB_REQUIRED");
		}
		// debug_var($queue, "ress");
		// Hook::event("RESEARCH_LOADED", array(&$items));

		Core::getTPL()->addLoop("imagePacks", getImgPacks());
		Core::getTPL()->addLoop("events", $queue);
		Core::getTPL()->addLoop("constructions", $items);
		Core::getTPL()->display("research");
		return $this;
	}

	/**
	* Check for sufficient resources and start research upgrade.
	*
	* @param integer	Building id to upgrade
	*
	* @return Research
	*/
	protected function upgradeResearch($id)
	{
		$id = max(0, (int)$id);
        if($id == UNIT_EXPO_TECH && !EXPEDITION_ENABLED){
            throw new GenericException("Expedition is disabled");
        }

		// 37.8 RACE-003: lock на 5 сек предотвращает двойную постройку через
		// двойной клик / параллельные HTTP. Ключ зависит от userid+research_id —
		// разные исследования не блокируют друг друга.
		$_userid_lock = $_SESSION["userid"] ?? 0;
		if( !NS::acquireLock("Research::upgrade:{$_userid_lock}:{$id}", 5) )
		{
			Logger::dieMessage("TOO_MANY_REQUESTS");
			doHeaderRedirection($this->main_page, false);
			return $this;
		}

		Core::getLanguage()->load("info,buildings");

		$events = NS::getEH()->getResearchEvents();

		$row = getConstructionDesc($id, $this->unit_type);
		if(!$row || NS::getUser()->get("observer") || (DEATHMATCH && (!isDeathmatchStarted() || isDeathmatchEnd())))
		{
			throw new GenericException("Unkown research $id, mode: $this->unit_type :(");
		}
		if($row["race"] != NS::getUser()->get("race"))
		{
			throw new GenericException("Error race to research $id :(");
		}

		$free_queue_size = getMaxCountListResearch() - count($events);
		if($free_queue_size <= 0)
		{
			// Logger::dieMessage("QUEUE_IS_FULL");
			// throw new GenericException("There is no empty slots in queue.");
			Logger::dieMessage("TOO_MANY_REQUESTS");
			doHeaderRedirection($this->main_page, false);
			return $this;
		}

		// Check for requirements
		$is_lab_exist = NS::getPlanet()->getBuilding(UNIT_RESEARCH_LAB) || NS::getPlanet()->getBuilding(UNIT_MOON_LAB);
		if(!NS::checkRequirements($id, null, null, $is_lab_exist) || !$is_lab_exist)
		{
			throw new GenericException("You does not fulfil the requirements to research this.");
		}

		// Check if research labor is not in progress
		if(!NS::getEH()->canReasearch())
		{
			throw new GenericException("Research labor in progress.");
		}

		// Hook::event("UPGRADE_RESEARCH_FIRST", array(&$row));

		$next_level = NS::getResearch($id) - NS::getAddedResearch($id) + 1;
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
				$next_level = max($next_level, $event["data"]["level"] + 1);
			}
		}

		if($next_level > MAX_RESEARCH_LEVEL || (isset($GLOBALS["MAX_UNIT_LEVELS"][$id]) && $next_level > $GLOBALS["MAX_UNIT_LEVELS"][$id]))
		{
			throw new GenericException("You have already reached maximum research level.");
		}

		foreach($events as $event)
		{
			// $event["data"] = unserialize($event["data"]);
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

		//debug_var(array($next_level, $row), "test");

		// Check resources
		if($this->checkResources())
		{
			$data["metal"] = $this->requiredMetal;
			$data["silicon"] = $this->requiredSilicon;
			$data["hydrogen"] = $this->requiredHydrogen;
			$data["energy"] = $this->requiredEnergy;
			$time = NS::getBuildingTime($data["metal"], $data["silicon"], $this->unit_type);
			$data["level"] = $next_level;
			$data["buildingid"] = $id;
			$data["buildingname"] = $row["name"];
			$data["batchkey"] = time() . mt_rand(1000, 9999);

			// Hook::event("UPGRADE_RESEARCH_LAST", array($row, &$data, &$time));

			//debug_var($data, "pre event");

			//Add in list or start immediately
			NS::getEH()->addEvent($this->eventType(), ($start_time > 0 ? $start_time : time()) + $time,
				NS::getUser()->get("curplanet"), NS::getUser()->get("userid"), null, $data, null,
				$start_time > 0 ? $start_time : null);

			doHeaderRedirection($this->main_page, false);
		}
		else
		{
			throw new GenericException("Not enough resources to research this.");
		}
		return $this;
	}

	/**
	* Check for sufficient credit and immediately start research upgrade
	*
	* @param integer	Building id to upgrade
	*
	* @return Research
	*/
	protected function upgradeResearchVIP($batchkey)
	{
		// $batchkey is a string

		NS::getEH()->startConstructionEventVIP($batchkey, $this->unit_type);

		doHeaderRedirection($this->main_page, false);
		return $this;
	}

	/**
	* Aborts the current research event.
	*
	* @param integer	Building id
	*
	* @return Research
	*/
	protected function abort($batchkey)
	{
		// $batchkey is a string

		NS::getEH()->abortConstructionEvent($batchkey, $this->unit_type);

		doHeaderRedirection($this->main_page, false);
		return $this;
	}
}

?>