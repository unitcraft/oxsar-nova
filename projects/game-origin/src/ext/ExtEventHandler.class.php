<?php
/**
* Handles all events.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("RECIPE_ROOT_DIR")) { die("Hacking attempt detected."); }

class ExtEventHandler extends EventHandler
{
	/* protected function genPrevRaceConditionKey($size = 16)
	{
        $s = substr(str_replace('.', '', uniqid('', true)), 4, $size).md5(microtime(true).mt_rand());;
		return substr($s, 0, $size);
	} */

	protected function queryExpiredEvents($cur_time, $max_batch_time = 10, $limit = 5)
	{
		Core::getQuery()->update("events",
			array( "prev_rc", /* "time", */ "processed", "processed_mode", "processed_time" ),
			array( $this->raceConditionKey, /* $cur_time, */ EVENT_PROCESSED_WAIT, null, $cur_time ),
			"(time <= ".sqlVal($cur_time)." AND processed='".EVENT_PROCESSED_WAIT."' AND prev_rc IS NULL AND (processed_time IS NULL OR processed_time <= ".sqlVal($cur_time-5)."))"
			. " OR (time < ".sqlVal($cur_time - ($max_batch_time + 10))
			// .				" and processed in ('".EVENT_PROCESSED_WAIT."', '".EVENT_PROCESSED_START."')"
			. " AND processed = '".EVENT_PROCESSED_WAIT."'" // look to DB to find unfinished EVENT_PROCESSED_START
			. " AND prev_rc IS NOT NULL)"
			. " ORDER BY time ASC, eventid ASC" // avoid deadlock
			. " LIMIT ".max(1, (int)$limit)
		);
		$select = array(
			"e.eventid",
			"e.mode",
			"e.start",
			"e.time",
			"e.planetid",
			"e.destination",
			"e.data",
			"e.user AS userid",
			"e.parent_eventid",
			"e.artid",
			"d.userid AS destuser",
			"d.username AS destname",
			"u.points",
			"u.username",
			"p.planetname",
			"p2.planetname AS destplanet"
		);
		$joins	= "LEFT JOIN ".PREFIX."planet p ON (p.planetid = e.planetid)";
		$joins .= "LEFT JOIN ".PREFIX."planet p2 ON (p2.planetid = e.destination)";
		$joins .= "LEFT JOIN ".PREFIX."user u ON (u.userid = e.user)";
		$joins .= "LEFT JOIN ".PREFIX."user d ON (d.userid = p2.userid)";
		$result = sqlSelect("events e", $select, $joins,
			// "e.time = ".sqlVal($cur_time)
			"e.processed_time = ".sqlVal($cur_time)
			. " and e.processed=".EVENT_PROCESSED_WAIT
			. " and e.prev_rc = " . sqlVal($this->raceConditionKey)
			. " and e.mode != " . EVENT_ALLIANCE_ATTACK_ADDITIONAL // ally event will be served by EVENT_ATTACK_ALLIANCE
			. " ORDER BY e.time ASC, e.eventid ASC");
		return $result;
	}

	public function removePlanetEvents($planetid, $params = array())
	{
		$force_remove = isset($params["force_remove"]) ? $params["force_remove"] : true;
		$save_scale = isset($params["save_scale"]) ? $params["save_scale"] : true;
		$except_modes = array(EVENT_RETURN, EVENT_RESEARCH);
		if(isset($params["except_modes"]))
		{
			$except_modes = array_unique( $except_modes + (array)$params["except_modes"] );
		}
		$result = sqlSelect("events", "eventid", "",
			"processed IN (".sqlArray(EVENT_PROCESSED_WAIT, EVENT_PROCESSED_START).")"
				. " AND mode NOT IN (".sqlArray($except_modes).") "
				. " AND (planetid=".sqlVal($planetid)." OR destination=".sqlVal($planetid).")"
		);
		while($row = sqlFetch($result))
		{
			$this->removeEvent($row["eventid"], $force_remove, $save_scale);
		}
		sqlEnd($result);

		return $this;
	}

	//#########################################################################//

	function startConstructionEventVIP($batchkey, $unit_type)
	{
		if(!NS::isFirstRun("EH:startConstructionEventVIP:$batchkey:$unit_type"))
		{
			return false;
		}

		$is_shipyard_event = false;
		switch($unit_type)
		{
			case UNIT_TYPE_CONSTRUCTION:
			case UNIT_TYPE_MOON_CONSTRUCTION:
				$events = $this->getBuildingEvents();
				break;
			case UNIT_TYPE_RESEARCH:
				$events = $this->getResearchEvents();
				break;
			case UNIT_TYPE_FLEET:
				$is_shipyard_event = true;
				$events = $this->getShipyardModeEvents(EVENT_BUILD_FLEET);
				break;
			case UNIT_TYPE_DEFENSE:
				$is_shipyard_event = true;
				$events = $this->getShipyardModeEvents(EVENT_BUILD_DEFENSE);
				break;
			case UNIT_TYPE_REPAIR:
			case UNIT_TYPE_DISASSEMBLE:
				$is_shipyard_event = true;
				$events = $this->getRepairEvents();
				break;
			default:
				return false;
		}

		$cur_event = false;
		$queue_end_time = time();
		foreach($events as $event)
		{
			if($event["data"]["batchkey"] == $batchkey)
			{
				$cur_event = $event;
				if(!$is_shipyard_event)
				{
					foreach($events as $event)
					{
						if($cur_event["data"]["batchkey"] != $event["data"]["batchkey"]
							&& $cur_event["data"]["buildingid"] == $event["data"]["buildingid"]
							&& $cur_event["data"]["level"] > $event["data"]["level"])
						{
							unset($cur_event);
							break;
						}
					}
				}
				break;
			}
			if(0 && /*$event["start"] <= time() &&*/ empty($event["data"]["vip"]))
			{
				$end_time = $event["time"];
				if($is_shipyard_event && !empty($event["data"]["new_format"])
					&& $event["data"]["quantity"] > 0)
				{
					$end_time = $event["start"] + $event["data"]["duration"] * $event["data"]["quantity"];
				}
				$queue_end_time = max($queue_end_time, $end_time);
			}
		}

		if( isset($cur_event["eventid"]) && empty($cur_event["data"]["vip"]) && $cur_event["start"] > time()+5 )
		{
			switch($unit_type)
			{
				case UNIT_TYPE_CONSTRUCTION:
				case UNIT_TYPE_MOON_CONSTRUCTION:
					$credit_req = getCreditImmStartConstructions($cur_event["data"]["level"]);
					break;

				case UNIT_TYPE_RESEARCH:
					$credit_req = getCreditImmStartResearch($cur_event["data"]["level"]);
					break;

				case UNIT_TYPE_FLEET:
				case UNIT_TYPE_DEFENSE:
					$credit_req = getCreditImmStartShipyard($cur_event["data"]["quantity"]);
					break;

				case UNIT_TYPE_REPAIR:
				case UNIT_TYPE_DISASSEMBLE:
					$credit_req = getCreditImmStartRepair($cur_event["data"]["quantity"]);
					break;
			}

			$credit = NS::getUser()->get("credit");
			if($credit_req > $credit)
			{
				Logger::dieMessage("NOT_ENOUGH_CREDITS");
				return false;
			}

			$cur_duration = $cur_event["time"] - $cur_event["start"];
			if(0)
			{
				$cur_end_time = time() + $cur_duration;
				if($is_shipyard_event && !empty($cur_event["data"]["new_format"])
					&& $cur_event["data"]["quantity"] > 0)
				{
					$cur_duration = $cur_event["data"]["duration"];
					$cur_end_time = time() + $cur_duration * $cur_event["data"]["quantity"];
				}
				$queue_end_time = $cur_end_time;
			}

			foreach($events as $event)
			{
				if($cur_event["data"]["batchkey"] == $event["data"]["batchkey"])
				{
					// $queue_end_time = max($queue_end_time, $cur_end_time);
				}
				elseif($event["start"] <= time())
				{
					if(empty($event["data"]["vip"]))
					{
						$end_time = $event["time"];
						if($is_shipyard_event && !empty($event["data"]["new_format"])
							&& $event["data"]["quantity"] > 0)
						{
							$end_time = $event["start"] + $event["data"]["duration"] * $event["data"]["quantity"];
						}
						$queue_end_time = max($queue_end_time, $end_time);
					}
				}
				elseif(empty($event["data"]["vip"]))
				{
					$end_time = $queue_end_time + $event["time"] - $event["start"];

					Core::getQuery()->update(
						"events",
						array("start", "time"),
						array($queue_end_time, $end_time),
						"eventid = ".sqlVal($event["eventid"])." AND processed=".EVENT_PROCESSED_WAIT
							. ' ORDER BY eventid'
					);

					if($is_shipyard_event && !empty($event["data"]["new_format"])
						&& $event["data"]["quantity"] > 0)
					{
						$end_time = $queue_end_time + $event["data"]["duration"] * $event["data"]["quantity"];
					}
					$queue_end_time = max($queue_end_time, $end_time);
				}
			}

			$cur_event["data"]["vip"] = 1;
			Core::getQuery()->update(
				"events",
				array("start", "time", "data"),
				array(time(), time() + $cur_duration, serialize($cur_event["data"]) ),
				"eventid = ".sqlVal($cur_event["eventid"])." AND processed=".EVENT_PROCESSED_WAIT
					. ' ORDER BY eventid'
			);

			if( $credit_req > 0 )
			{
				new AutoMsg(MSG_CREDIT, $cur_event["user"], time(), array('credits' => $credit_req, 'msg' => 'MSG_CREDIT_FOR_VIP' ));
			}

			NS::updateUserRes(array(
				"type" 			=> RES_UPDATE_VIP_START,
				"event_mode" 	=> $cur_event["mode"],
				"ownerid" 		=> $cur_event["eventid"],
				"userid" 		=> $cur_event["user"],// I think that user is now used instead of userid.	//$cur_event["userid"],
				"credit" 		=> - $credit_req,
			));
			return true;
		}
		return false;
	}

	function abortConstructionEvent($batchkey, $unit_type)
	{
		if(!NS::isFirstRun("EH:abortConstructionEvent:$batchkey:$unit_type"))
		{
			return false;
		}

		if(NS::isPlanetUnderAttack())
		{
			Logger::dieMessage('PLANET_UNDER_ATTACK');
		}

		$is_shipyard_event = false;
		switch($unit_type)
		{
			case UNIT_TYPE_CONSTRUCTION:
			case UNIT_TYPE_MOON_CONSTRUCTION:
				$events = $this->getBuildingEvents();
				break;
			case UNIT_TYPE_RESEARCH:
				$events = $this->getResearchEvents();
				break;
			case UNIT_TYPE_FLEET:
				$is_shipyard_event = true;
				$events = $this->getShipyardModeEvents(EVENT_BUILD_FLEET);
				break;
			case UNIT_TYPE_DEFENSE:
				$is_shipyard_event = true;
				$events = $this->getShipyardModeEvents(EVENT_BUILD_DEFENSE);
				break;
			case UNIT_TYPE_REPAIR:
			case UNIT_TYPE_DISASSEMBLE:
				$is_shipyard_event = true;
				$events = $this->getRepairEvents();
				break;
			default:
				return false;
		}
		$cur_event = false;
		$queue_end_time = time();
		foreach($events as $event)
		{
			if($event["data"]["batchkey"] == $batchkey)
			{
				$cur_event = $event;
				if(!$is_shipyard_event)
				{
					foreach($events as $event)
					{
						if($cur_event["data"]["batchkey"] != $event["data"]["batchkey"]
							&& $cur_event["data"]["buildingid"] == $event["data"]["buildingid"]
							&& $cur_event["data"]["level"] < $event["data"]["level"])
						{
							// debug_var($cur_event, "[startVIP] found level: ".$event["data"]["level"]);
							unset($cur_event);
							break;
						}
					}
				}
				break;
			}
			if(0 && /*$event["start"] <= time() &&*/ empty($event["data"]["vip"]))
			{
				$end_time = $event["time"];
				if(
					$is_shipyard_event
					&& !empty($event["data"]["new_format"])
					&& $event["data"]["quantity"] > 0
				)
				{
					$end_time = $event["start"] + $event["data"]["duration"] * $event["data"]["quantity"];
				}
				$queue_end_time = max($queue_end_time, $end_time);
			}
		}
		if( isset($cur_event["eventid"]) )
		{
			// $queue_end_time = time();
			foreach($events as $event)
			{
				if( $cur_event["data"]["batchkey"] == $event["data"]["batchkey"] ) // Recalculate other events
				{
					;
				}
				elseif( $event["start"] <= time() )
				{
					if(empty($event["data"]["vip"]))
					{
						$end_time = $event["time"];
						if(
							$is_shipyard_event
							&& !empty($event["data"]["new_format"])
							&& $event["data"]["quantity"] > 0
						)
						{
							$end_time = $event["start"] + $event["data"]["duration"] * $event["data"]["quantity"];
						}
						$queue_end_time = max($queue_end_time, $end_time);
					}
				}
				elseif(empty($event["data"]["vip"]))
				{
					$end_time = $queue_end_time + $event["time"] - $event["start"];

					Core::getQuery()->update("events",
						array("start", "time"),
						array($queue_end_time, $end_time),
						"eventid = ".sqlVal($event["eventid"])." AND processed=".EVENT_PROCESSED_WAIT
							. ' ORDER BY eventid');

					if($is_shipyard_event && !empty($event["data"]["new_format"])
						&& $event["data"]["quantity"] > 0)
					{
						$end_time = $queue_end_time + $event["data"]["duration"] * $event["data"]["quantity"];
					}
					$queue_end_time = max($queue_end_time, $end_time);
				}
			}
			$event_build_data = $this->getEventBuildingData($cur_event);
			if( $this->removeEvent( $cur_event["eventid"], false, $event_build_data["abort_res_scale"] )
                    && !empty($cur_event["data"]["vip"]) )
			{
				switch($unit_type)
				{
					case UNIT_TYPE_CONSTRUCTION:
					case UNIT_TYPE_MOON_CONSTRUCTION:
						$credit_req = max(1, round(getCreditImmStartConstructions($cur_event["data"]["level"]) * $event_build_data["abort_res_scale"]));
						break;
					case UNIT_TYPE_RESEARCH:
						$credit_req = max(1, round(getCreditImmStartResearch($cur_event["data"]["level"]) * $event_build_data["abort_res_scale"]));
						break;
					case UNIT_TYPE_FLEET:
					case UNIT_TYPE_DEFENSE:
						$credit_req = getCreditAbortShipyard($cur_event["data"]["quantity"], $event_build_data["abort_res_scale"]);
						break;
					case UNIT_TYPE_REPAIR:
					case UNIT_TYPE_DISASSEMBLE:
						$credit_req = getCreditAbortRepair($cur_event["data"]["quantity"], $event_build_data["abort_res_scale"]);
						break;
				}

				NS::updateUserRes(array(
					"type" => RES_UPDATE_CANCEL,
					"event_mode" => $cur_event["mode"],
					"ownerid" => $cur_event["eventid"],
					"userid" => $cur_event["userid"],
					"credit" => $credit_req,
				));
			}
		}
	}

	//#########################################################################//
	//############################# EVENT METHODS #############################//
	//#########################################################################//

	protected function fReturn($row, $data)
	{
		if( !isPlanetOccupied($row["destination"]) )
		{
			$new_target = NS::findFleetTarget($data["sgalaxy"], $data["ssystem"], $data["sposition"], $data["maxspeed"], $row["userid"], $row["destination"]);
			if($new_target)
			{
				$data["sgalaxy"] = $new_target["galaxy"];
				$data["ssystem"] = $new_target["system"];
				$data["sposition"] = $new_target["position"];
				$data["time"] = $new_target["time"];

				sqlInsert("events", array(
					"mode" 				=> $row["mode"], // EVENT_RETURN,
					"start" 			=> time(),
					"time" 				=> time() + $data["time"],
					"planetid" 			=> $row["destination"],
					"user" 				=> $row["userid"],
					"destination" 		=> $new_target["planetid"],
					"data" 				=> serialize($data),
					"parent_eventid" 	=> $row["eventid"],
					"protected" 		=> 0,
				));
			}
			return $this;
		}

		return parent::fReturn($row, $data);
	}

	protected function repair($row, $data)
	{
		return $this->shipyard($row, $data);
	}

	protected function disassemble($row, $data)
	{
        $start_event_process_time = microtime(true);
        if($data["quantity"] >= 1)
        {
            $data["quantity"] = max(1, floor($data["quantity"]));
            // $data["duration"] = max(1, floor($data["duration"]));
            $data["duration"] = max(0.001, floatval($data["duration"]));
            $start_time = $row["start"]; // + $data["duration"];
            $end_time = time();
            $build_quantity = max(1, min($data["quantity"], round(($end_time - $start_time) / $data["duration"])));

			NS::updateUserRes(array(
				"type" => RES_UPDATE_DISASSEMBLE,
				"event_mode" => $row["mode"],
				"reload_planet" => false,
				"update_production" => true, // $processed_quantity == 0,
				"ownerid" => $row["eventid"],
				"userid" => $row["userid"],
				"planetid" => $row["planetid"],
				"metal" => $data["earn_metal"] * $build_quantity,
				"silicon" => $data["earn_silicon"] * $build_quantity,
				"hydrogen" => $data["earn_hydrogen"] * $build_quantity,
			));

			sqlQuery("UPDATE ".PREFIX."events SET processed_quantity = case when processed_quantity is null then ".sqlVal($build_quantity)." else processed_quantity + ".sqlVal($build_quantity)." end "
				. " WHERE eventid=".sqlVal($row["eventid"])
					. ' ORDER BY eventid');

            $data["quantity"] -= $build_quantity;
			if($data["quantity"] >= 1)
			{
				$basic_metal = $data["basic_metal"];
				$basic_silicon = $data["basic_silicon"];
				$basic_hydrogen = $data["basic_hydrogen"];
				$basic_energy = $data["basic_energy"];

				if($basic_metal > 0 || $basic_silicon > 0 || $basic_hydrogen > 0 || $basic_energy > 0)
				{
					if(isset($data["damaged"]))
					{
						$data["damaged"] = max(0, floor($data["damaged"])-1);
					}

					$data["metal"] -= $basic_metal * $build_quantity;
					$data["silicon"] -= $basic_silicon * $build_quantity;
					$data["hydrogen"] -= $basic_hydrogen * $build_quantity;
					$data["energy"] -= $basic_energy * $build_quantity;

					$data["repair_usage"] -= $data["unit_fields"] * $build_quantity;

					// getBuildTime is not valid here, event owner is not owner of the current planet
					// $data["duration"] = getBuildTime($basic_metal, $basic_silicon, 2);
					$data["paid"] = 1;

                    $start_time = time();
                    $end_time = $start_time + NS::getUnitsBuildTime($data);

                    // Update By Pk
                    sqlUpdate("events", array(
                        "start" => $start_time,
                        "time" => $end_time,
                        "data" => serialize($data),
                        "prev_rc" => null,
                        "processed" => EVENT_PROCESSED_WAIT,
                        "processed_mode" => $row["mode"],
                        "processed_time" => $cur_time,
                        "processed_dt" => round(microtime(true) - $start_event_process_time, 5),
                        ), "eventid=".sqlVal($row["eventid"])
                    );
				}
			}
		}
		if( !defined('YII_CONSOLE') && $row["planetid"] == Core::getUser()->get("curplanet"))
		{
			NS::reloadPlanet();
		}
		return $this;
	}

	protected function disassembleOld($row, $data)
	{
        $start_event_process_time = microtime(true);
		for($processed_quantity = 0;;)
		{
			NS::updateUserRes(array(
				"type" => RES_UPDATE_DISASSEMBLE,
				"event_mode" => $row["mode"],
				"reload_planet" => false,
				"update_production" => $processed_quantity == 0,
				"ownerid" => $row["eventid"],
				"userid" => $row["userid"],
				"planetid" => $row["planetid"],
				"metal" => $data["earn_metal"],
				"silicon" => $data["earn_silicon"],
				"hydrogen" => $data["earn_hydrogen"],
			));

			sqlQuery("UPDATE ".PREFIX."events SET processed_quantity = case when processed_quantity is null then 1 else processed_quantity + 1 end "
				. " WHERE eventid=".sqlVal($row["eventid"])
					. ' ORDER BY eventid');
			$processed_quantity++;

			if($data["quantity"] > 1)
			{
				$data["quantity"] = floor($data["quantity"]);

				$basic_metal = $data["basic_metal"];
				$basic_silicon = $data["basic_silicon"];
				$basic_hydrogen = $data["basic_hydrogen"];
				$basic_energy = $data["basic_energy"];

				if($basic_metal > 0 || $basic_silicon > 0 || $basic_hydrogen > 0 || $basic_energy > 0)
				{
                    $data["duration"] = max(1, round($data["duration"]));
					$start_time = $row["start"] + $data["duration"];

					$data["quantity"] -= 1;
					if(isset($data["damaged"]))
					{
						$data["damaged"] = max(0, floor($data["damaged"])-1);
					}

					$data["metal"] -= $basic_metal; // * $data["quantity"];
					$data["silicon"] -= $basic_silicon; // * $data["quantity"];
					$data["hydrogen"] -= $basic_hydrogen; // * $data["quantity"];
					$data["energy"] -= $basic_energy; // * $data["quantity"];

					$data["repair_usage"] -= $data["unit_fields"];

					// getBuildTime is not valid here, event owner is not owner of the current planet
					// $data["duration"] = getBuildTime($basic_metal, $basic_silicon, 2);
					$data["paid"] = 1;

					$end_time = $start_time + $data["duration"];

					if($end_time <= time() && $processed_quantity < 100)
					{
						$row["start"] = $start_time;
						$row["time"] = $end_time;
						continue;
					}

					$cur_time = time();
					// $end_time = max($end_time, $cur_time+1);
					// $start_time = $end_time - $data["duration"];
                    $start_time = $cur_time;
                    $end_time = $start_time + $data["duration"] * clampVal(round($data["quantity"] / 100), 1, 100);
					// Update By Pk
					sqlUpdate("events", array(
						"start" => $start_time,
						"time" => $end_time+10,
						"data" => serialize($data),
						"prev_rc" => null,
						"processed" => EVENT_PROCESSED_WAIT,
						"processed_mode" => $row["mode"],
						"processed_time" => $cur_time,
                        "processed_dt" => round(microtime(true) - $start_event_process_time, 5),
						), "eventid=".sqlVal($row["eventid"])
					);
				}
			}
			break;
		}
		if( !defined('YII_CONSOLE') && $row["planetid"] == Core::getUser()->get("curplanet"))
		{
			NS::reloadPlanet();
		}
		return $this;
	}

	protected function teleportPlanet($row, $data)
	{
		$data["galaxy"] = clampVal($data["galaxy"], 1, NUM_GALAXYS);
		$data["system"] = clampVal($data["system"], 1, NUM_SYSTEMS);
		$data["position"] = clampVal($data["position"], 1, MAX_NORMAL_PLANET_POSITION);

		if(!isset($data["ships"][ARTEFACT_PLANET_TELEPORTER])
			|| !sqlSelectRow("planet", "planetid", "", "planetid=".sqlVal($row["planetid"])." AND userid=".sqlVal($row["userid"])." AND ismoon=1"))
		{
			$this->removeEvent($row["eventid"], true);
			return $this;
		}

		$galaxy_row = sqlSelectRow("galaxy", "*", "",
			"galaxy=".sqlVal($data["galaxy"])
			." AND system=".sqlVal($data["system"])
			." AND position=".sqlVal($data["position"]));
		if(!$galaxy_row)
		{
			$error = false;
			try{
				sqlUpdate("galaxy", array(
					"galaxy" => $data["galaxy"],
					"system" => $data["system"],
					"position" => $data["position"],
				), "moonid=".sqlVal($row["planetid"])." LIMIT 1");
			}catch(Exception $e){
				$error = true;
			}

			if(!$error)
			{
				$galaxy_row = sqlSelectRow("galaxy", "planetid", "",
					"moonid=".sqlVal($row["planetid"])
					." AND galaxy=".sqlVal($data["galaxy"])
					." AND system=".sqlVal($data["system"])
					." AND position=".sqlVal($data["position"])
				);
				if($galaxy_row)
				{
					$art = reset($data["ships"][ARTEFACT_PLANET_TELEPORTER]["art_ids"]);
					$artid = $art["artid"];
					if( Artefact::activate($artid, $row["userid"], $row["planetid"]) )
					{
						unset( $data["ships"][ARTEFACT_PLANET_TELEPORTER] );
					}

					$temperature = PlanetCreator::getTemperature($data["position"]);
					sqlUpdate("planet", array("temperature" => $temperature), "planetid=".sqlVal($galaxy_row["planetid"])." LIMIT 1"); // planet
					sqlUpdate("planet", array("temperature" => $temperature - mt_rand(15, 35)), "planetid=".sqlVal($row["planetid"])." LIMIT 1"); // moon
					sqlUpdate("user", array("planet_teleport_time" => time()), "userid=".sqlVal($row["userid"])." LIMIT 1");

					Core::getQuery()->delete("stargate_jump", "planetid = " . sqlVal($row["planetid"]));
					Core::getQuery()->insert(
						"stargate_jump",
						array("planetid", "time", "data"),
						array($row["planetid"], time(), serialize($data))
					);

					new AutoMsg(MSG_PLANET_TELEPORTED, $row["userid"], time(), $data);

					$row['destuser'] = $row["userid"];
					$row['destination'] = $row["planetid"];
					$row['data'] = $data;
					return $this->position($row, $data);
				}
			}
		}
		new AutoMsg(MSG_PLANET_NOT_TELEPORTED, $row["userid"], time(), $data);

		$this->removeEvent($row["eventid"], true);
		return $this;
	}

	protected function allianceAttack($row, $data)
	{
		$parent_eventid = $row["eventid"];

		require_once(APP_ROOT_DIR."game/Assault.class.php");
		$Assault = new Assault($this, $row["eventid"], $row["destination"], $row["destuser"]);
		$Assault->addParticipant(1, $row["userid"], $row["planetid"], $row["time"], $data, $parent_eventid);

		$processed_time = time();
		Core::getQuery()->update("events", array("processed", "processed_time"),
			array(EVENT_PROCESSED_START, $processed_time),
			"parent_eventid = ".sqlVal($parent_eventid)." AND mode = ".EVENT_ALLIANCE_ATTACK_ADDITIONAL." AND processed=".EVENT_PROCESSED_WAIT
				. ' ORDER BY eventid');

		// Load allied fleets
		$_result = sqlSelect("events", array("user AS userid", "planetid", "data"), "",
			"parent_eventid = ".sqlVal($parent_eventid)." AND mode = ".EVENT_ALLIANCE_ATTACK_ADDITIONAL." AND processed_time=".sqlVal($processed_time));
		while($_row = sqlFetch($_result))
		{
			$Assault->addParticipant(1, $_row["userid"], $_row["planetid"], $row["time"], unserialize($_row["data"]), $parent_eventid);
		}

		$Assault->startAssault($data["galaxy"], $data["system"], $data["position"], array("mode" => $row["mode"])); // == EVENT_ATTACK_ALLIANCE_DESTROY_BUILDING);

		Core::getQuery()->update("events", array("processed", "processed_mode", "processed_time"),
			array(EVENT_PROCESSED_OK, EVENT_ALLIANCE_ATTACK_ADDITIONAL, time()),
			"parent_eventid = ".sqlVal($parent_eventid)." AND mode = ".EVENT_ALLIANCE_ATTACK_ADDITIONAL." AND processed_time=".sqlVal($processed_time)
				. ' ORDER BY eventid');

		Core::getQuery()->delete("formation_invitation", "eventid = ".sqlVal($parent_eventid));
		Core::getQuery()->delete("attack_formation", "eventid = ".sqlVal($parent_eventid));
		return $this;
	}

	protected function rocketAttack($row, $data)
	{
		Core::getLanguage()->load(array("info", "AutoMessages"));
		if($data["ships"][UNIT_INTERPLANETARY_ROCKET]["quantity"] >= 1)
		{
			$data["ships"] = array(UNIT_INTERPLANETARY_ROCKET => $data["ships"][UNIT_INTERPLANETARY_ROCKET]);

			require_once(APP_ROOT_DIR."game/Assault.class.php");
			$Assault = new Assault($this, $row["eventid"], $row["destination"], $row["destuser"], true);
			$Assault->addParticipant(
				1,
				$row["userid"],
				$row["planetid"],
				$row["time"],
				$data,
				0,
				$data["primary_target"]
			);
			$Assault->startAssault($data["galaxy"], $data["system"], $data["position"]);
		}
		return $this;
	}

	/**
	* Loads the fleets of an alliance attack.
	*
	* @param integer	Destination planet id
	* @param integer	Arrival time
	*
	* @return array	List of all formation fleets
	*/
	public function getFormationAllyFleets($parent_eventid) // $planetid, $time)
	{
		$fleets = array();
		$joins	= "LEFT JOIN ".PREFIX."planet p ON (p.planetid = e.planetid) ";
		$joins .= "LEFT JOIN ".PREFIX."galaxy g ON (g.planetid = e.planetid) ";
		$joins .= "LEFT JOIN ".PREFIX."user u ON (u.userid = p.userid) ";
		$select = array("e.*", "u.username", "p.planetname", "g.galaxy", "g.system", "g.position");
		$result = sqlSelect("events e", $select, $joins,
			"e.parent_eventid = ".sqlVal($parent_eventid)." AND e.mode = ".EVENT_ALLIANCE_ATTACK_ADDITIONAL." AND processed=".EVENT_PROCESSED_WAIT,
			"e.time ASC");
		while($row = sqlFetch($result))
		{
			$row["data"] = unserialize($row["data"]);
			$fleets[] = $row;
		}
		sqlEnd($result);
		return $fleets;
	}

	protected function haltPosition($event)
	{
		$data = $event["data"];

		$data["time"] = $data["org_time"];

		$data["back_consumption"] = ceil($data["org_consumption"] / 2);
		$data["hydrogen"] += $data["back_consumption"];

		$data["consumption"] = 0;
		if($data["hydrogen"] >= $data["back_consumption"])
		{
			$data["hydrogen"] -= $data["back_consumption"];
			$data["consumption"] = $data["back_consumption"] * 2;
		}

		$data["duration"] = 60*60*1;

		unset($data["org_time"]);
		unset($data["org_consumption"]);

		$this->addEvent(EVENT_HOLDING, time() + $data["duration"], $event["planetid"], $event["userid"], $event["destination"],
			$data,
			null,
			null, // $event["time"],
			$event["eventid"] // $parent_eventid
			);

		new AutoMsg(MSG_POSITION_REPORT, $event["userid"], time(), $data);

		return $this;
	}

	protected function haltReturn($event)
	{
		if(isset($event["data"]["pathstack"]) && count($event["data"]["pathstack"]) > 0)
		{
			$fleet_params = NS::calcFleetParams($event["data"]["ships"]); // , 1, 1, $event["userid"]);

			$stack_data = array_pop($event["data"]["pathstack"]);
			$data = array(
				"ships" => $event["data"]["ships"],
				"metal" => $event["data"]["metal"],
				"silicon" => $event["data"]["silicon"],
				"hydrogen" => $event["data"]["hydrogen"] + ( isset($event["data"]["ret_consumption"]) ? $event["data"]["ret_consumption"] : 0 ),
				"galaxy" => $stack_data["galaxy"],
				"system" => $stack_data["system"],
				"position" => $stack_data["position"],
				"sgalaxy" => $stack_data["sgalaxy"],
				"ssystem" => $stack_data["ssystem"],
				"sposition" => $stack_data["sposition"],
				"maxspeed" => $stack_data["maxspeed"],
				"time" => $stack_data["time"],
				// "consumption" => $fleet_params["consumption"],
				"consumption" => $stack_data["consumption"],
				// "back_consumption" => ceil($stack_data["consumption"]/2),
				"fleet_size" => $fleet_params["fleet_size"],
				"control_times" => $event["data"]["control_times"],
				"duration" => 60*60*1,
				"pathstack" => $event["data"]["pathstack"],
			);

			$data["capacity"] = $fleet_params["capacity"] // - ceil($data["consumption"]/2)
									- $data["metal"]
									- $data["silicon"]
									- $data["hydrogen"];

			/*
			$distance = NS::getDistance($data["galaxy"], $data["system"], $data["position"],
								$data["sgalaxy"], $data["ssystem"], $data["sposition"]);
			$data["consumption"] = NS::getFlyConsumption($move_data["consumption"], $distance);
			*/

			// fleet size consumption
			$time = $data["time"]; // NS::getFlyTime($distance, $data["maxspeed"]);
			$groupConsumption = unitGroupConsumptionPerHour( UNIT_VIRT_FLEET, $data["fleet_size"], true ) * $time * 2 / (60 * 60);
			$data["back_consumption"] = ceil(max($data["consumption"], $groupConsumption) / 2);

			$data["consumption"] = 0;
			if($data["hydrogen"] >= $data["back_consumption"])
			{
				$data["hydrogen"] -= $data["back_consumption"];
				$data["consumption"] = $data["back_consumption"] * 2;
			}

			$this->addEvent(EVENT_HOLDING, time() + $data["duration"], $stack_data["planetid"], $event["userid"], $stack_data["destination"],
				$data,
				null,
				null, // $event["time"],
				$event["eventid"] // $parent_eventid
				);

			new AutoMsg(MSG_RETURN_REPORT, $event["userid"], time(), $data);
		}

		return $this;
	}

	protected function alienFlyUnknown($event)
	{
		AlienAI::onFlyUnknownEvent($event);
		return $this;
	}

	protected function alienGrabCredit($event)
	{
		AlienAI::onGrabCreditEvent($event);
		return $this;
	}

	protected function alienHolding($event)
	{
		AlienAI::onHoldingEvent($event);
		return $this;
	}

	protected function alienHoldingAI($event)
	{
		AlienAI::onHoldingAIEvent($event);
		return $this;
	}

	protected function alienChangeMissionAI($event)
	{
		AlienAI::onChangeMissionAIEvent($event);
		return $this;
	}

	protected function alienAttack($event)
	{
		AlienAI::onAttackEvent($event);
		return $this;
	}

	protected function alienHalt($event)
	{
		AlienAI::onHaltEvent($event);
		return $this;
	}
}
?>