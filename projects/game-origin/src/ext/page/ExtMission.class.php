<?php
/**
* Allows the user to send fleets to a mission.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("RECIPE_ROOT_DIR")) { die("Hacking attempt detected."); }

class ExtMission extends Mission
{
	/**
	* Executes star gate jump and update shipyard informations.
	*
	* @param integer	Target moon id
	*
	* @return Mission
	*/
	protected function executeJump($moonid)
	{
		if(!NS::isFirstRun("Mission::executeJump:".NS::getUser()->get("userid")))
		{
			Logger::dieMessage('TOO_MANY_REQUESTS');
		}

		$temp = sqlSelectRow("temp_fleet", "data", "", "planetid = ".sqlPlanet());
		sqlDelete("temp_fleet", "planetid = ".sqlPlanet());
		if($temp)
		{
			if(NS::isPlanetUnderAttack())
			{
				Logger::dieMessage('PLANET_UNDER_ATTACK');
			}

			$temp = unserialize($temp["data"]);
			if(!in_array($moonid, $temp["moons"]))
			{
				Logger::dieMessage('UNKOWN_MISSION');
				// throw new GenericException("Unkown moon.");
			}

			$ships = array();
			foreach($temp["ships"] as $key => $value)
			{
				$ships[$key] = $value["quantity"];
			}
			// debug_var($ships, "[executeJump] ships");
			unset($temp["ships"]);

			$data = array();
			$fleet_size = 0;
			$is_defense_jump = false;
			$select = array("u2s.unitid", "u2s.quantity", "u2s.damaged", "u2s.shell_percent", "d.capicity", "d.speed", "d.consume", "b.name", "b.mode");
			$joins	= "LEFT JOIN ".PREFIX."construction b ON (b.buildingid = u2s.unitid)";
			$joins .= "LEFT JOIN ".PREFIX."ship_datasheet d ON (d.unitid = u2s.unitid)";
			$result = sqlSelect("unit2shipyard u2s", $select, $joins, "b.mode IN (".sqlArray(UNIT_TYPE_FLEET, UNIT_TYPE_DEFENSE).")
				AND b.buildingid NOT IN (".sqlArray($GLOBALS["BLOCKED_STARGATE_UNITS"]).")
				AND u2s.planetid = ".sqlPlanet());
			while($row = sqlFetch($result))
			{
				$id = $row["unitid"];
				$quantity = min($row["quantity"], $ships[$id]);
				if($quantity > 0)
				{
					$temp["ships"][$id]["name"] = $row["name"];
					$temp["ships"][$id]["id"] = $id;
					// $data["ships"][$id]["quantity"] = $quantity;
					extractUnits($row, $quantity, $temp["ships"][$id]);
					$fleet_size += $temp["ships"][$id]["quantity"];
					if($row["mode"] == UNIT_TYPE_DEFENSE)
					{
						$is_defense_jump = true;
					}
				}
			}
			sqlEnd($result);
			// debug_var($temp, "[executeJump] checked data"); exit;

			if( true ) // NS::getUser()->get("userid") == 1)
			{
				$long_jump_time = 60*10;
				$row = sqlSelectRow("galaxy", "*", "", "moonid=".sqlVal($moonid));
				if($row)
				{
					$data["time"] = NS::getPlanet()->isMoon() && !$is_defense_jump ? 0 : $long_jump_time; // duration
				}
				else
				{
					$row = sqlSelectRow("galaxy", "*", "", "planetid=".sqlVal($moonid));
					if(!$row)
					{
						Logger::dieMessage('UNKOWN_MISSION');
					}
					$data["time"] = $long_jump_time; // duration
				}

				$data["ships"] = $temp["ships"];
				$data["galaxy"] = $row["galaxy"];
				$data["system"] = $row["system"];
				$data["position"] = $row["position"];
				$data["sgalaxy"] = NS::getPlanet()->getData("galaxy");
				$data["ssystem"] = NS::getPlanet()->getData("system");
				$data["sposition"] = NS::getPlanet()->getData("position");
				$data["maxspeed"] = MAX_SPEED_POSSIBLE;
				$data["expeditionMode"] = 0;
				$data["fleet_size"] = $fleet_size;
				$data["consumption"] = 0;
				$data["metal"] = 0;
				$data["silicon"] = 0;
				$data["hydrogen"] = 0;
				$data["capacity"] = 0;
				// $data["time"] = 0; // duration

				NS::getEH()->addEvent(EVENT_STARGATE_JUMP, time() + $data["time"], NS::getUser()->get("curplanet"), NS::getUser()->get("userid"), $moonid, $data);

			}
			else
			{
				foreach($temp["ships"] as $key => $value)
				{
					logShipsChanging($key, Core::getUser()->get("curplanet"), - $value["quantity"], 1, "[executeJump] begin teleport to the moonid: $moonid");
					sqlQuery("UPDATE ".PREFIX."unit2shipyard SET "
						. subQuantitySetSql($value)
						. " WHERE unitid = ".sqlVal($key)." AND planetid = ".sqlPlanet()
						. ' ORDER BY unitid DESC, planetid DESC');

					logShipsChanging($key, $moonid, $value["quantity"], 1, "[executeJump] end teleport from planetid: " . Core::getUser()->get("curplanet"));

					$exist_quantity = sqlSelectField("unit2shipyard", "quantity", "", "unitid = ".sqlVal($key)." AND planetid = ".sqlVal($moonid));
					if($exist_quantity > 0)
					{
						$sql = "UPDATE ".PREFIX."unit2shipyard SET "
							. addQuantitySetSql($value)
							. " WHERE unitid = ".sqlVal($key)." AND planetid = ".sqlVal($moonid)
							. ' ORDER BY unitid DESC, planetid DESC';

						// debug_var($event, "[executeJump] sql: $sql"); exit;

						sqlQuery($sql);
					}
					else
					{
						Core::getQuery()->insert("unit2shipyard",
							array("unitid", "planetid", "quantity", "damaged", "shell_percent"),
							array($key, $moonid, $value["quantity"], $value["damaged"], $value["shell_percent"]));
					}
				}

				sqlDelete("unit2shipyard", "quantity = '0'");
			}

			// Desable curr moon stargate
			if( $GLOBALS['STARGATE']['START_DISABLE'] )
			{
				Core::getQuery()->delete("stargate_jump", "planetid = ".sqlVal(NS::getPlanet()->getData("moonid")));
				Core::getQuery()->insert("stargate_jump", array("planetid", "time", "data"), array(NS::getPlanet()->getData("moonid"), time(), serialize($temp)));
			}
			// Desable target moon stargate
			if( $GLOBALS['STARGATE']['END_DISABLE'] )
			{
				$moonid = NS::getMoonid($moonid);
				Core::getQuery()->delete("stargate_jump", "planetid = ".sqlVal($moonid));
				Core::getQuery()->insert("stargate_jump", array("planetid", "time", "data"), array($moonid, time(), serialize($temp)));
			}

			Logger::addMessage("JUMP_STARTED_SUCCESSFULLY", "success"); // mt_rand(0, 100) ? "JUMP_SUCCESSFUL" : "JUMP_SUCCESSFUL_HITCHED", "success");
		}
		return $this->index();
	}

	/**
	* Select the ships for jump.
	*
	* @param array		Ships for jump
	*
	* @return Mission
	*/
	protected function starGateDefenseJump($ships)
	{
		return $this->starGateJump($ships, true);
	}

	protected function starGateJump($ships, $defense_jump = false)
	{
		/* if(NS::getPlanet()->getBuilding(UNIT_STAR_GATE) <= 0)
		{
			return $this->index();
		} */
		/* if(!NS::isFirstRun("Mission::starGateJump:" . md5(serialize($ships)) . "-".time()))
		{
			Logger::dieMessage('TOO_MANY_REQUESTS');
		} */

		$data = array();
		sqlDelete("temp_fleet", "planetid = ".sqlPlanet());

		$allow_mode = array(UNIT_TYPE_FLEET);
		if($defense_jump)
		{
			$allow_mode[] = UNIT_TYPE_DEFENSE;
		}

		$is_defense_jump = false;
		$select = array("u2s.unitid", "u2s.quantity", "u2s.damaged", "u2s.shell_percent", "d.capicity", "d.speed", "d.consume", "b.name", "b.mode");
		$joins	= "LEFT JOIN ".PREFIX."construction b ON (b.buildingid = u2s.unitid)";
		$joins .= "LEFT JOIN ".PREFIX."ship_datasheet d ON (d.unitid = u2s.unitid)";
		$result = sqlSelect("unit2shipyard u2s", $select, $joins, "b.mode IN (".sqlArray($allow_mode).")
			AND u2s.planetid = ".sqlPlanet()."
			AND b.buildingid NOT IN (".sqlArray($GLOBALS["BLOCKED_STARGATE_UNITS"]).")");
		while($row = sqlFetch($result))
		{
			$id = $row["unitid"];
			$quantity = min($row["quantity"], $ships[$id]);
			if($quantity > 0)
			{
				$data["ships"][$id]["name"] = $row["name"];
				// $data["ships"][$id]["quantity"] = $quantity;
				extractUnits($row, $quantity, $data["ships"][$id]);
				if($row["mode"] == UNIT_TYPE_DEFENSE)
				{
					$is_defense_jump = true;
				}
			}
		}
		sqlEnd($result);

		if(count($data["ships"]) < 1)
		{
			// Logger::addMessage("EMPTY_FLEET_FOR_STAR_GATE");
			return $this->index();
		}

		$stargate_time = 0;
		if(NS::getPlanet()->isMoon())
		{
			$stargate_time = NS::getPlanet()->getStarGateTime();
			$stargate_level = NS::getPlanet()->getBuilding(UNIT_STAR_GATE);
		}
		else // if(isAdmin(null, true)/*2.6*/)
		{
			// $moonlab_level = NS::getPlanet()->getExtBuilding(UNIT_MOON_LAB);
			$stargate_level = NS::getPlanet()->getExtBuilding(UNIT_STAR_GATE);
			if(/*$moonlab_level > 0 ||*/ $stargate_level >= 3)
			{
				$stargate_time = NS::getPlanet()->getStarGateTime(true);
			}
			else
			{
				Logger::dieMessage("UNKOWN_MISSION");
			}
		}

		// $stargate_level = NS::getPlanet()->getBuilding(UNIT_STAR_GATE);
		// $stargate_time = NS::getPlanet()->getStarGateTime(isAdmin(null, true)/*2.6*/ ? $stargate_level >= 2 : false);
		// $row = sqlSelectRow("stargate_jump", "time", "", "planetid = ".sqlPlanet());
		$show_ext_message = $is_defense_jump;
		$gate_cycle_secs = starGateRecycleSecs($stargate_level);
		if(!$stargate_time || $stargate_time < time() - $gate_cycle_secs || !OXSAR_RELEASED)
		{
			$moons = array();
			// if(isAdmin(null, true)/*2.6*/)
			{
				$select = array("p.planetid", "p.planetname", "g.galaxy", "g.system", "g.position", "sj.time", "b2p.level");
				$joins	= "JOIN ".PREFIX."galaxy g ON g.moonid = p.planetid
						   JOIN ".PREFIX."building2planet b2p ON b2p.planetid = g.moonid
						   LEFT JOIN ".PREFIX."stargate_jump sj ON sj.planetid = g.moonid";
				$where = "p.userid = ".sqlUser()
					// . " AND p.ismoon = '1' "
					// . " AND p.planetid != ".sqlPlanet()
					. " AND b2p.buildingid = ".UNIT_STAR_GATE;
				if($stargate_level >= 2)
				{
					$show_ext_message = true;
					if(NS::getPlanet()->isMoon())
					{
						$only_near_planet = /*!NS::getPlanet()->getBuilding(UNIT_MOON_LAB) &&*/ $stargate_level < 4;
						$joins .= " LEFT JOIN ".PREFIX."planet p2 ON p2.planetid = g.planetid".($only_near_planet ? " AND g.moonid=".sqlPlanet() : "");
						$select[] = "p2.planetid as planetid2";
						$select[] = "p2.planetname as planetname2";
					}
					else if(/*!NS::getPlanet()->getBuilding(UNIT_MOON_LAB) &&*/ $stargate_level < 3)
					{
						Logger::dieMessage("UNKOWN_MISSION");
						$where = "1=0";
					}
					else
					{
						$only_near_moon = /*!NS::getPlanet()->getBuilding(UNIT_MOON_LAB) &&*/ $stargate_level < 4;
						$joins	= "
							JOIN ".PREFIX."galaxy g ON g.planetid = p.planetid".($only_near_moon ? " AND g.planetid=".sqlPlanet() : "")."
							JOIN ".PREFIX."planet m ON m.planetid = g.moonid
							JOIN ".PREFIX."building2planet b2p ON b2p.planetid = g.moonid
							LEFT JOIN ".PREFIX."stargate_jump sj ON sj.planetid = g.moonid
							";
						$select[] = "m.planetid as planetid2";
						$select[] = "m.planetname as planetname2";
					}
				}
			}
			/* else // old way
			{
				$select = array("p.planetid", "p.planetname", "g.galaxy", "g.system", "g.position", "sj.time", "b2p.level");
				$joins	= "JOIN ".PREFIX."galaxy g ON g.moonid = p.planetid
						   JOIN ".PREFIX."building2planet b2p ON b2p.planetid = g.moonid
						   LEFT JOIN ".PREFIX."stargate_jump sj ON sj.planetid = g.moonid";
				$where = "p.userid = ".sqlUser()
					. " AND p.ismoon = '1' "
					. " AND p.planetid != ".sqlPlanet()
					. " AND b2p.buildingid = ".UNIT_STAR_GATE;
			} */
			$planetid = NS::getPlanet()->getPlanetId();
			$result = sqlSelect("planet p", $select, $joins, $where);
			while($row = sqlFetch($result))
			{
				$gate_cycle_secs = starGateRecycleSecs($row["level"]);
				if($row["time"] < time() - $gate_cycle_secs || !OXSAR_RELEASED)
				{
					if($row["planetid"] != $planetid)
					{
						$moons[] = $row;
						$data["moons"][] = $row["planetid"];
					}
					if($row["planetid2"] && $row["planetid2"] != $planetid)
					{
						$row["planetid"] = $row["planetid2"];
						$row["planetname"] = $row["planetname2"];
						$moons[] = $row;
						$data["moons"][] = $row["planetid"];
					}
				}
			}
			sqlEnd($result);

			// Hook::event("SHOW_STAR_GATES", array(&$moons, &$data));
			Core::getTPL()->addLoop("moons", $moons);
			Core::getTPL()->assign("show_ext_message", $show_ext_message);
			Core::getTPL()->assign("fleet", getUnitListStr($data["ships"]));

			$data = serialize($data);
			Core::getQuery()->insert("temp_fleet", array("planetid", "data"), array(Core::getUser()->get("curplanet"), $data));
			Core::getTPL()->display("stargatejump");
			exit();
		}
		else
		{
			Logger::addMessage("RELOADING_STAR_GATE");
		}
		return $this->index();
	}

	protected function getControlResource($res, $comis)
	{
		return max(0, floor($res * 100 / ($comis + 100)));
	}

	protected function getControlComis($is_event_owner)
	{
		// return $is_event_owner ? 5 : round((NS::getUser()->get("exchange_rate") - 1) * 100);
		return round((NS::getUser()->get("exchange_rate") - 1) * 100);
	}

	protected function unloadResourcesFromFleet($eventid, $metal, $silicon, $hydrogen)
	{
		$eventid = max(0, (int)$eventid);
		$metal = max(0, (int)$metal);
		$silicon = max(0, (int)$silicon);
		$hydrogen = max(0, (int)$hydrogen);

		if(!NS::isFirstRun("Mission::unloadResourcesFromFleet:{$eventid},{$metal},{$silicon},{$hydrogen}"))
		{
			Logger::dieMessage('TOO_MANY_REQUESTS');
		}

		$event = NS::getEH()->getEvent($eventid);
		if( $event && $event["mode"] == EVENT_HOLDING ) // && $event["destination"] == NS::getPlanet()->getPlanetId() )
		{
			if(!isset($event["data"]["back_consumption"]))
			{
				$event["data"]["back_consumption"] = max(0, ceil($event["data"]["consumption"] / 2));
			}
			$event["data"]["hydrogen"] += $event["data"]["consumption"] > 0 ? $event["data"]["back_consumption"] : 0;

			$is_event_owner = $event["user"] == NS::getUser()->get("userid");
			$comis = $this->getControlComis($is_event_owner);
			// $consumption = $event["data"]["back_consumption"];

			$metal = min($metal, $this->getControlResource($event["data"]["metal"], $comis));
			$silicon = min($silicon, $this->getControlResource($event["data"]["silicon"], $comis));
			$hydrogen = min($hydrogen, $this->getControlResource($event["data"]["hydrogen"], $comis));

			foreach(array("metal", "silicon", "hydrogen") as $res_name)
			{
				/*
				if($$res_name > $capacity)
				{
					$$res_name = $capacity;
				}
				$capacity -= $$res_name;
				*/
				${"used_".$res_name} = $$res_name + ceil($$res_name * $comis / 100);
				$event["data"][$res_name] = max(0, floor($event["data"][$res_name] - ${"used_".$res_name}));
			}

			$fleet_params = NS::calcFleetParams($event["data"]["ships"]);
			$event["data"]["capacity"] = max(0, $fleet_params["capacity"]
													- $event["data"]["metal"]
													- $event["data"]["silicon"]
													- $event["data"]["hydrogen"]);

			$event["data"]["consumption"] = 0;
			if($event["data"]["hydrogen"] >= $event["data"]["back_consumption"])
			{
				$event["data"]["hydrogen"] -= $event["data"]["back_consumption"];
				$event["data"]["consumption"] = $event["data"]["back_consumption"] * 2;
			}

			if($metal > 0 || $silicon > 0 || $hydrogen > 0)
			{
				$res_log = NS::updateUserRes(array(
					"block_minus" => true,
					"type" => RES_UPDATE_UNLOAD_FLEET,
					// "event_mode" => $mode,
					"userid" => NS::getUser()->get("userid"),
					"planetid" => $event["destination"], // NS::getPlanet()->getPlanetId(),
					"metal" => $metal,
					"silicon" => $silicon,
					"hydrogen" => $hydrogen,
				));

				sqlUpdate("events", array(
					"data" => serialize($event["data"])
				), "processed=".EVENT_PROCESSED_WAIT." AND eventid=".sqlVal($eventid));
			}
		}
		doHeaderRedirection("game.php/ControlFleet/".$eventid, false);
	}

	protected function loadResourcesToFleet($eventid, $metal, $silicon, $hydrogen)
	{
		$eventid = max(0, (int)$eventid);
		$metal = max(0, (int)$metal);
		$silicon = max(0, (int)$silicon);
		$hydrogen = max(0, (int)$hydrogen);

		if(!NS::isFirstRun("Mission::loadResourcesToFleet:{$eventid},{$metal},{$silicon},{$hydrogen}"))
		{
			Logger::dieMessage('TOO_MANY_REQUESTS');
		}

		$event = NS::getEH()->getEvent($eventid);
		if( $event && $event["mode"] == EVENT_HOLDING && $event["destination"] == NS::getPlanet()->getPlanetId() && $event["destination"] != $event["planetid"] )
		{
			if(!isset($event["data"]["back_consumption"]))
			{
				$event["data"]["back_consumption"] = max(0, ceil($event["data"]["consumption"] / 2));
			}
			$event["data"]["hydrogen"] += $event["data"]["consumption"] > 0 ? $event["data"]["back_consumption"] : 0;

			$is_event_owner = $event["user"] == NS::getUser()->get("userid");
			$comis = $this->getControlComis($is_event_owner);

			$metal = min($metal, $this->getControlResource(NS::getPlanet()->getData("metal"), $comis));
			$silicon = min($silicon, $this->getControlResource(NS::getPlanet()->getData("silicon"), $comis));
			$hydrogen = min($hydrogen, $this->getControlResource(NS::getPlanet()->getData("hydrogen"), $comis));

			$fleet_params = NS::calcFleetParams($event["data"]["ships"]);
			$capacity = max(0, $fleet_params["capacity"]
									- $event["data"]["metal"]
									- $event["data"]["silicon"]
									- $event["data"]["hydrogen"]);

			foreach(array("metal", "silicon", "hydrogen") as $res_name)
			{
				if($$res_name > $capacity)
				{
					$$res_name = $capacity;
				}
				$capacity -= $$res_name;
				${"planet_".$res_name} = $$res_name + ceil($$res_name * $comis / 100);
			}

			if($metal > 0 || $silicon > 0 || $hydrogen > 0)
			{
				$res_log = NS::updateUserRes(array(
					"block_minus" => true,
					"type" => RES_UPDATE_LOAD_FLEET,
					// "event_mode" => $mode,
					"userid" => NS::getUser()->get("userid"),
					"planetid" => NS::getPlanet()->getPlanetId(),
					"metal" => - $planet_metal,
					"silicon" => - $planet_silicon,
					"hydrogen" => - $planet_hydrogen,
					"auto_fix" => array(
						"metal" => $planet_metal,
						"silicon" => $planet_silicon,
						"hydrogen" => $planet_hydrogen,
						),
				));

				if(empty($res_log["minus_blocked"]))
				{
					if(!empty($res_log["auto_fixed"]))
					{
						$metal = $this->getControlResource($planet_metal - $res_log["auto_fixed"]["metal"], $comis);
						$silicon = $this->getControlResource($planet_silicon - $res_log["auto_fixed"]["silicon"], $comis);
						$hydrogen = $this->getControlResource($planet_hydrogen - $res_log["auto_fixed"]["hydrogen"], $comis);
					}
					$event["data"]["metal"] += $metal;
					$event["data"]["silicon"] += $silicon;
					$event["data"]["hydrogen"] += $hydrogen;

					$event["data"]["capacity"] = max(0, $fleet_params["capacity"]
													- $event["data"]["metal"]
													- $event["data"]["silicon"]
													- $event["data"]["hydrogen"]);

					$event["data"]["consumption"] = 0;
					if($event["data"]["hydrogen"] >= $event["data"]["back_consumption"])
					{
						$event["data"]["hydrogen"] -= $event["data"]["back_consumption"];
						$event["data"]["consumption"] = $event["data"]["back_consumption"] * 2;
					}

					sqlUpdate("events", array(
						"data" => serialize($event["data"])
					), "processed=".EVENT_PROCESSED_WAIT." AND eventid=".sqlVal($eventid));
				}
			}
		}
		// Core::getTPL()->display("mission_control");
		doHeaderRedirection("game.php/ControlFleet/".$eventid, false);
	}

	protected function controlFleetPost($eventid)
	{
		// return $this->controlFleet($eventid);
		doHeaderRedirection("game.php/ControlFleet/".$eventid, false);
	}

	protected function controlFleet($eventid)
	{
		$event = NS::getEH()->getEvent($eventid);
		if( $event && $event["mode"] == EVENT_HOLDING )
		{
			if(!isset($event["data"]["back_consumption"]))
			{
				$event["data"]["back_consumption"] = max(0, ceil($event["data"]["consumption"] / 2));
			}
			$is_event_owner = $event["user"] == NS::getUser()->get("userid");
			$comis = $this->getControlComis($is_event_owner); // $is_event_owner ? 5 : round((NS::getUser()->get("exchange_rate") - 1) * 100);
			$consumption = isset($event["data"]["consumption"]) ? ceil($event["data"]["consumption"]/2) : 0;

			$fleet_params = NS::calcFleetParams($event["data"]["ships"]);
			$capacity = max(0, $fleet_params["capacity"] - $consumption
									- $event["data"]["metal"]
									- $event["data"]["silicon"]
									- $event["data"]["hydrogen"]);

			if(empty($event["destname"]))
			{
				$event["destname"] = sqlSelectField("planet p", "username",
					"JOIN ".PREFIX."user u ON u.userid = p.userid",
					"p.planetid=".sqlVal($event["destination"]));
			}
			Core::getTPL()->assign("eventid", $eventid);
			Core::getTPL()->assign("fleet", Core::getLanguage()->getItemWith("FLEET_MESSAGE_HOLDING_FUEL", array(
				"fleet" => getUnitListStr($event["data"]["ships"]),
				"username" => $event["username"],
				"message" => Link::get("game.php/MSG/Write/Receiver:".$event["username"], Image::getImage("pm.gif", Core::getLanguage()->getItem("WRITE_MESSAGE"))),
				"planet" => $event["planetname"],
				"coords" => getCoordLink($event["galaxy"], $event["system"], $event["position"], false, $event["planetid"]),
				"target" => $event["destplanet"],
				"targetcoords" => getCoordLink($event["galaxy2"], $event["system2"], $event["position2"], false, $event["destination"]),
				"target_username" => $event["destname"],
				"target_message" => Link::get("game.php/MSG/Write/Receiver:".$event["destname"], Image::getImage("pm.gif", Core::getLanguage()->getItem("WRITE_MESSAGE"))),
				"metal" => $event["data"]["metal"] ? fNumber($event["data"]["metal"]) : 0,
				"silicon" => $event["data"]["silicon"] ? fNumber($event["data"]["silicon"]) : 0,
				"hydrogen" => $event["data"]["hydrogen"] ? fNumber($event["data"]["hydrogen"]) : 0,
				"fuel_hydrogen" => fNumber($consumption),
				"mission" => NS::getMissionName($event["mode"], $is_event_owner),
				)));
			// Core::getTPL()->assign("retreat_enabled", true);
			$holding_planet_owner = sqlSelectField("planet", "count(*)", "", "planetid=".sqlVal($event["destination"])." AND userid=".sqlUser());
			Core::getTPL()->assign("holding_planet_owner", $holding_planet_owner);
			Core::getTPL()->assign("load_possible", $event["destination"] != $event["planetid"] );
			Core::getTPL()->assign("back_possible", !( empty($event["data"]["consumption"]) && !empty($event["data"]["back_consumption"]) && $event["destination"] != $event["planetid"] ));
			Core::getTPL()->assign("back_consumption", fNumber($event["data"]["back_consumption"]));
			Core::getTPL()->assign("back_consumption_needed", fNumber($consumption > 0 ? 0 : $event["data"]["back_consumption"] - $event["data"]["hydrogen"]));
			Core::getTPL()->assign("holding_planet_selected", $event["destination"] == NS::getPlanet()->getPlanetId());
			Core::getTPL()->assign("holding_planet_select", getCoordLink(false, false, false, $holding_planet_owner ? "select" : false, $event["destination"], true));
			Core::getTPL()->assign("holding_select_coords_action", socialUrl(RELATIVE_URL . "game.php/Mission"));
			Core::getTPL()->assign("holding_planet_username", $event["destname"]);
			Core::getTPL()->assign("holding_planet_message", Link::get("game.php/MSG/Write/Receiver:".$event["destname"], Image::getImage("pm.gif", Core::getLanguage()->getItem("WRITE_MESSAGE"))));
			Core::getTPL()->assign("comis", $comis);
			Core::getTPL()->assign("fleet_metal", $this->getControlResource($event["data"]["metal"], $comis));
			Core::getTPL()->assign("fleet_silicon", $this->getControlResource($event["data"]["silicon"], $comis));
			Core::getTPL()->assign("fleet_hydrogen", $this->getControlResource($event["data"]["hydrogen"] + $consumption, $comis));
			if($event["destination"] == NS::getPlanet()->getPlanetId())
			{
				Core::getTPL()->assign("metal", $this->getControlResource(NS::getPlanet()->getData("metal"), $comis));
				Core::getTPL()->assign("silicon", $this->getControlResource(NS::getPlanet()->getData("silicon"), $comis));
				Core::getTPL()->assign("hydrogen", $this->getControlResource(NS::getPlanet()->getData("hydrogen"), $comis));
			}
			else
			{
				Core::getTPL()->assign("metal", 0);
				Core::getTPL()->assign("silicon", 0);
				Core::getTPL()->assign("hydrogen", 0);
			}
			Core::getTPL()->assign("capacity", $capacity);
			Core::getTPL()->assign("rest", fNumber($capacity));
			Core::getTPL()->assign("merchant_mark_used", Artefact::getMerchantMark());
			Core::getTPL()->assign("remain_controls", NS::getRemainFleetControls($event));
			Core::getTPL()->display("mission_control");
			exit();
		}
		doHeaderRedirection("game.php/Main", false);
	}

	protected function holdingSelectCoords($eventid)
	{
		$event = NS::getEH()->getEvent($eventid);
		if( $event && $event["mode"] == EVENT_HOLDING && $event["destination"] == NS::getPlanet()->getPlanetId() )
		{
			$is_event_owner = $event["user"] == NS::getUser()->get("userid");
			$comis = $this->getControlComis($is_event_owner); // $is_event_owner ? 5 : round((NS::getUser()->get("exchange_rate") - 1) * 100);
			$consumption = isset($event["data"]["consumption"]) ? ceil($event["data"]["consumption"]/2) : 0;
			$capacity = floor($event["data"]["hydrogen"] + $consumption);

			if( $capacity <= 0 ) // || $consumption <= 0 )
			{
				Logger::addFlashMessage("NOT_ENOUGH_FUEL");
				doHeaderRedirection("game.php/ControlFleet/".$eventid, false);
			}

			if(empty($event["destname"]))
			{
				$event["destname"] = sqlSelectField("planet p", "username",
					"JOIN ".PREFIX."user u ON u.userid = p.userid",
					"p.planetid=".sqlVal($event["destination"]));
			}
			Core::getTPL()->assign("holding_eventid", $eventid);
			Core::getTPL()->assign("fleet", Core::getLanguage()->getItemWith("FLEET_MESSAGE_HOLDING", array(
				"fleet" => getUnitListStr($event["data"]["ships"]),
				"username" => $event["username"],
				"message" => Link::get("game.php/MSG/Write/Receiver:".$event["username"], Image::getImage("pm.gif", Core::getLanguage()->getItem("WRITE_MESSAGE"))),
				"planet" => $event["planetname"],
				"coords" => getCoordLink($event["galaxy"], $event["system"], $event["position"], false, $event["planetid"]),
				"target" => $event["destplanet"],
				"targetcoords" => getCoordLink($event["galaxy2"], $event["system2"], $event["position2"], false, $event["destination"]),
				"target_username" => $event["destname"],
				"target_message" => Link::get("game.php/MSG/Write/Receiver:".$event["destname"], Image::getImage("pm.gif", Core::getLanguage()->getItem("WRITE_MESSAGE"))),
				"metal" => $event["data"]["metal"] ? fNumber($event["data"]["metal"]) : 0,
				"silicon" => $event["data"]["silicon"] ? fNumber($event["data"]["silicon"]) : 0,
				"hydrogen" => $event["data"]["hydrogen"] ? fNumber($event["data"]["hydrogen"]) : 0,
				"mission" => NS::getMissionName($event["mode"], $is_event_owner),
				)));

			Core::getTPL()->addHTMLHeaderFile("lib/jquery.countdown.js?".CLIENT_VERSION, "js");

			if( 1 ) // $event["destination"] == NS::getPlanet()->getPlanetId())
			{
				$galaxy = NS::getPlanet()->getData("galaxy");
				$system = NS::getPlanet()->getData("system");
				$position = NS::getPlanet()->getData("position");
			}
			else
			{
				$galaxy = $event["galaxy2"];
				$system = $event["system2"];
				$position = $event["position2"];
			}

			sqlDelete("temp_fleet", "planetid = ".sqlPlanet());

			// $consumption = $event["data"]["consumption"];
			$consumption = max($event["data"]["consumption"], (isset($event["data"]["back_consumption"]) ? $event["data"]["back_consumption"]*2 : 0));
			$speed = $event["data"]["maxspeed"];
			$speed_factor = 1; // $fleet_params["speed_factor"];
			$fleet_size = $event["data"]["fleet_size"];

			if(count($event["data"]["ships"]) > 0 && $speed > 0)
			{
				$distance = NS::getDistance($galaxy, $system, $position);
				$time = NS::getFlyTime($distance, $speed);

				Core::getTPL()->assign("can_send_fleet", true);
				Core::getTPL()->assign("can_send_expo", false);

				Core::getTPL()->assign("galaxy", $galaxy);
				Core::getTPL()->assign("system", $system);
				Core::getTPL()->assign("position", $position);
				Core::getTPL()->assign("oGalaxy", $galaxy); // NS::getPlanet()->getData("galaxy"));
				Core::getTPL()->assign("oSystem", $system); // NS::getPlanet()->getData("system"));
				Core::getTPL()->assign("oPos", $position); // NS::getPlanet()->getData("position"));
				Core::getTPL()->assign("maxspeedVar", $speed);
				Core::getTPL()->assign("distance", $distance);
				Core::getTPL()->assign("capicity_raw", $capacity);
				Core::getTPL()->assign("basicConsumption", $consumption);
				Core::getTPL()->assign("fleetSize", $fleet_size);
				Core::getTPL()->assign("time", getTimeTerm($time));
				Core::getTPL()->assign("maxspeed", fNumber($speed));
				Core::getTPL()->assign("capicity", fNumber($capacity - NS::getFlyConsumption($consumption, $distance)));
				Core::getTPL()->assign("fuel", fNumber(NS::getFlyConsumption($consumption, $distance)));
				// Core::getTPL()->assign("fleet", getUnitListStr($data));
				// Core::getTPL()->assign("allow_stargate_transport", (int)NS::getGalaxyParam(NS::getPlanet()->getData("galaxy"), "ALLOW_STARGATE_TRANSPORT", isset($event["data"]["ships"][ARTEFACT_IGLA_MORI])));
				Core::getTPL()->assign("allow_stargate_transport", (int)isset($event["data"]["ships"][ARTEFACT_IGLA_MORI]));

				if($speed_factor != 1 && $speed_factor > 0)
				{
					$old_speed = $speed / $speed_factor;
					$speed_delta = $speed - $old_speed;
					Core::getTPL()->assign("speedbonus",
						$speed_delta > 0
						? "<span class='true'>( +".fNumber($speed_delta)." )</span>"
						: "<span class='false'>( ".fNumber($speed_delta)." )</span>"
					);
					// debug_var(array($speed, $old_speed, $speed_delta), "[speeddelta]");
				}

				$data["consumption"] = $consumption;
				$data["maxspeed"] = $speed;
				$data["capacity"] = $capicity;
				$data["galaxy"] = $galaxy;
				$data["system"] = $system;
				$data["position"] = $position;
				$data["holding_eventid"] = $eventid;

				$data = serialize($data);
				Core::getQuery()->insert("temp_fleet", array("planetid", "data"), array(NS::getUser()->get("curplanet"), $data));

				// Short speed selection
				$selectbox = "";
				for($n = 10; $n > 0; $n--)
				{
					$selectbox .= createOption($n * 10, $n * 10, 0);
				}
				Core::getTPL()->assign("speedFromSelectBox", $selectbox);

				// Invatations for alliance attack
				$invitations = array();
				$joins	= "LEFT JOIN ".PREFIX."events e ON (e.eventid = fi.eventid)";
				$joins .= "LEFT JOIN ".PREFIX."attack_formation af ON (e.eventid = af.eventid)";
				$joins .= "LEFT JOIN ".PREFIX."galaxy g ON (g.planetid = e.destination)";
				$joins .= "LEFT JOIN ".PREFIX."galaxy m ON (m.moonid = e.destination)";
				$select = array("af.eventid", "af.name", "af.time", "g.galaxy", "g.system", "g.position", "m.galaxy AS moongala", "m.system AS moonsys", "m.position AS moonpos");
				$_result = sqlSelect("formation_invitation fi", $select, $joins,
					"fi.userid = ".sqlUser()." AND af.time > ".sqlVal(time())." AND e.processed=".EVENT_PROCESSED_WAIT
					);
				while($_row = sqlFetch($_result))
				{
					$_row["type"] = 0;
					if(!empty($_row["moongala"]) && !empty($_row["moonsys"]) && !empty($_row["moonpos"]))
					{
						$_row["galaxy"] = $_row["moongala"];
						$_row["system"] = $_row["moonsys"];
						$_row["position"] = $_row["moonpos"];
						$_row["type"] = 2;
					}
					$_row["time_r"] = $_row["time"] - time();
					$_row["formatted_time"] = getTimeTerm($_row["time_r"]);
					$invitations[] = $_row;
				}
				sqlEnd($_result);
				Core::getTPL()->addLoop("invitations", $invitations);

				// Planet shortlinks
				$i = 1;
				$sl = array();
				$order = getPlanetOrder(NS::getUser()->get("planetorder"));
				$_result = sqlSelect("planet p", array("p.ismoon", "p.planetname", "g.galaxy", "g.system", "g.position", "gm.galaxy moongala", "gm.system as moonsys", "gm.position as moonpos"),
					"LEFT JOIN ".PREFIX."galaxy g ON (g.planetid = p.planetid) LEFT JOIN ".PREFIX."galaxy gm ON (gm.moonid = p.planetid)",
					"p.userid = ".sqlUser()." AND p.planetid != ".sqlPlanet(),
					// "p.userid = ".sqlUser()." AND p.planetid != ".sqlVal($event["destination"]),
					$order);
				while($_row = sqlFetch($_result))
				{
					$sl[$i]["planetname"] = $_row["planetname"];
					$sl[$i]["galaxy"] = ($_row["ismoon"]) ? $_row["moongala"] : $_row["galaxy"];
					$sl[$i]["system"] = ($_row["ismoon"]) ? $_row["moonsys"] : $_row["system"];
					$sl[$i]["position"] = ($_row["ismoon"]) ? $_row["moonpos"] : $_row["position"];
					$sl[$i]["type"] = ($_row["ismoon"]) ? 2 : 0;
					$i++;
				}
				sqlEnd($_result);
				// Hook::event("MISSION_PLANET_QUICK_LINKS", array(&$sl));
				Core::getTPL()->addLoop("shortlinks", $sl);

				Core::getTPL()->display("missions2");
				exit();
			}
		}
		doHeaderRedirection("game.php/Main", false);
	}

	protected function holdingSendFleet($post) // $mode, $metal, $silicon, $hydrogen, $holdingtime)
	{
		if(!NS::isFirstRun("Mission::holdingSendFleet:" . md5(serialize($post)) . "-" . $_SESSION["userid"] ?? 0))
		{
			Logger::dieMessage('TOO_MANY_REQUESTS');
		}

		$curr_planet_id = NS::getUser()->get("curplanet");
		$row = sqlSelectRow("temp_fleet", "data", "", "planetid = ".sqlVal($curr_planet_id));
		sqlDelete('temp_fleet', "planetid = ".sqlVal($curr_planet_id));

		if(empty($row))
		{
			Logger::dieMessage("UNKOWN_MISSION");
		}

		if(NS::isPlanetUnderAttack())
		{
			Logger::dieMessage('PLANET_UNDER_ATTACK');
		}

		$data = unserialize($row["data"]);
		if(!empty($data) && isset($data["holding_eventid"]))
		{
			$holding_event = NS::getEH()->getEvent($data["holding_eventid"]);
			if( $holding_event && $holding_event["mode"] == EVENT_HOLDING && $holding_event["destination"] == NS::getPlanet()->getPlanetId() )
			{
				$ef_rules = array(
					'fleet' => true,
					'expo' 	=> false,
				);
				Core::getTPL()->assign("can_send_fleet", $ef_rules["fleet"]);
				Core::getTPL()->assign("can_send_expo", $ef_rules["expo"]);
				Core::getTPL()->assign("holding_eventid", $holding_event["eventid"]);
			}
		}
		if( !isset($holding_event) )
		{
			Logger::dieMessage('UNKOWN_MISSION');
		}

		$mode = max(0, (int)$post["mode"]);
		$metal = 0; // max(0, (int)$post["metal"]);
		$silicon = 0; // max(0, (int)$post["silicon"]);
		$hydrogen = 0; // max(0, (int)$post["hydrogen"]);
		$holdingtime = max(0, (int)$post["holdingtime"]);

		$temp = $data; // unserialize($row["data"]);
		unset($data);

		if(empty($temp["amissions"]) || !is_array($temp["amissions"]) || !in_array($mode, $temp["amissions"]))
		{
			Logger::dieMessage("UNKOWN_MISSION");
		}

		$data = $holding_event["data"];
		if(!isset($data["pathstack"]))
		{
			$data["pathstack"] = array();
		}
		array_push($data["pathstack"], array(
			"planetid" => $holding_event["planetid"],
			"destination" => $holding_event["destination"],
			"galaxy" => $holding_event["data"]["galaxy"],
			"system" => $holding_event["data"]["system"],
			"position" => $holding_event["data"]["position"],
			"sgalaxy" => $holding_event["data"]["sgalaxy"],
			"ssystem" => $holding_event["data"]["ssystem"],
			"sposition" => $holding_event["data"]["sposition"],
			"maxspeed" => $holding_event["data"]["maxspeed"],
			"consumption" => max($holding_event["data"]["consumption"], (isset($holding_event["data"]["back_consumption"]) ? $holding_event["data"]["back_consumption"]*2 : 0)),
			"time" => $holding_event["data"]["time"],
			));
		// $data["ships"] = $temp["ships"];
		$data["control_times"] = (isset($holding_event["data"]["control_times"]) ? $holding_event["data"]["control_times"] : 0) + 1;
		$data["galaxy"] = $temp["galaxy"];
		$data["system"] = $temp["system"];
		$data["position"] = $temp["position"];
		$data["sgalaxy"] = NS::getPlanet()->getData("galaxy");
		$data["ssystem"] = NS::getPlanet()->getData("system");
		$data["sposition"] = NS::getPlanet()->getData("position");
		$data["maxspeed"] = $temp["maxspeed"];
		$data["expeditionMode"] = 0; // $temp["expeditionMode"];

		$fleet_params = NS::calcFleetParams($data["ships"]); // , 1, $holding_event["userid"]);

		// $data['maxspeed'] = $fleet_params['maxspeed'];
		$data['fleet_size'] = $fleet_params['fleet_size'];

		if($mode == EVENT_ALLIANCE_ATTACK_ADDITIONAL)
		{
			$data["alliance_attack"] = $temp["alliance_attack"];
		}

		$distance = NS::getDistance($data["galaxy"], $data["system"], $data["position"]);
		$data["consumption"] = NS::getFlyConsumption($temp["consumption"], $distance, $temp["speed"]);

		$same_pos_halting = false;
		if($mode == EVENT_HALT)
		{
			$data["duration"] = min(99, max(0, (int)$holdingtime)) * 3600;
			if($temp["destination"] == NS::getUser()->get("curplanet"))
			{
				Logger::dieMessage("UNKOWN_MISSION");

				$data["consumption"] = 0;
				$distance = 0;
				$metal = 0;
				$silicon = 0;
				$hydrogen = 0;
				$same_pos_halting = true;
			}
		}

		// fleet size consumption
		$groupConsumption = unitGroupConsumptionPerHour( UNIT_VIRT_FLEET, $data['fleet_size'], true ) * $time * 2 / (60 * 60);

		$time = $same_pos_halting ? 1 : NS::getFlyTime($distance, $data["maxspeed"], $temp["speed"]);
		if($mode == EVENT_STARGATE_TRANSPORT)
		{
			$data["org_time"] = $time;
			$data["org_consumption"] = max($data["consumption"], $groupConsumption);

			$time = ceil($time * STARGATE_TRANSPORT_TIME_SCALE);
			$data["consumption"] = ceil($data["consumption"] * STARGATE_TRANSPORT_CONSUMPTION_SCALE);
		}

		$data["consumption"] = max($data["consumption"], $groupConsumption);

		$data["metal"] = $holding_event["data"]["metal"];
		$data["silicon"] = $holding_event["data"]["silicon"];
		$data["hydrogen"] = $holding_event["data"]["hydrogen"]
			+ (isset($holding_event["data"]["consumption"]) ? ceil($holding_event["data"]["consumption"]/2) : 0)
			- $data["consumption"];
		if($data["hydrogen"] < 0)
		{
			Logger::dieMessage("NOT_ENOUGH_FUEL");
		}

		$data["capacity"] = $fleet_params["capacity"] - $data["consumption"]
								- $data["metal"]
								- $data["silicon"]
								- $data["hydrogen"];
		if($data["capacity"] < 0)
		{
			Logger::dieMessage("NOT_ENOUGH_CAPACITY");
		}

		// If mission is recycling, get just the capacity of the recyclers.
		if($mode == EVENT_RECYCLING && $data["capacity"] > 0)
		{
			$recycle_capacity = 0;
			foreach($GLOBALS["RECYCLER_UNITS"] as $recycler_unit_id)
			{
				if(isset($data["ships"][$recycler_unit_id]) && $data["ships"][$recycler_unit_id]["quantity"] > 0)
				{
					$_row = sqlSelectRow("ship_datasheet", "capicity", "", "unitid = ".sqlVal($recycler_unit_id)); // It is __capacity__, not capicity
					$recycle_capacity += $_row["capicity"] * $data["ships"][$recycler_unit_id]["quantity"];
				}
			}
			$data["capacity"] = clampVal($recycle_capacity, 0, $data["capacity"]); // max(0, min($data["capacity"], $recycle_capacity));
		}
		// $time = $same_pos_halting ? 1 : NS::getFlyTime($distance, $data["maxspeed"], $temp["speed"]);

		if ( $mode == EVENT_STARGATE_TRANSPORT )
		{
			$igla_activated = false;
			foreach($data["ships"] as $key => $value)
			{
				if( $key == ARTEFACT_IGLA_MORI )
				{
					foreach($value["art_ids"] as $art)
					{
						if(Artefact::activate($art["artid"], $holding_event["userid"], null))
						{
							$igla_activated = true;
							break 2;
						}
					}
				}
			}
			if(!$igla_activated)
			{
				Logger::dieMessage("IGLA_NOT_ACTIVATED");
			}
		}

		if ( $mode == EVENT_STARGATE_TRANSPORT || $mode == EVENT_STARGATE_JUMP )
		{
			if( $GLOBALS['STARGATE']['START_DISABLE'] )
			{
				Core::getQuery()->delete("stargate_jump", "planetid = " . sqlVal(Core::getUser()->get("curplanet")));
				Core::getQuery()->insert(
					"stargate_jump",
					array("planetid", "time", "data"),
					array(Core::getUser()->get("curplanet"), time(), serialize($data))
					);
			}
			if( $GLOBALS['STARGATE']['END_DISABLE'] )
			{
				Core::getQuery()->delete("stargate_jump", "planetid = " . sqlVal($temp["destination"]));
				Core::getQuery()->insert(
					"stargate_jump",
					array("planetid", "time", "data"),
					array($temp["destination"] , time(), serialize($data))
					);
			}
		}
		$data["time"] = $time;

		$new_eventid = NS::getEH()->addEvent($mode, time() + $time,
			$holding_event["destination"], // NS::getUser()->get("curplanet"),
			$holding_event["userid"], // NS::getUser()->get("userid"),
			$temp["destination"],
			$data,
			null, // $protected
			null, // $start_time
			$holding_event["eventid"] // $parent_eventid
			);

		if($new_eventid)
		{
			sqlUpdate("events", array(
					"processed" => EVENT_PROCESSED_OK,
					"processed_time" => time(),
					"error_message" => "next eventid: $new_eventid",
					// "data" => serialize($event["data"])
				), "processed=".EVENT_PROCESSED_WAIT." AND eventid=".sqlVal($holding_event["eventid"]));
		}
		else
		{
			Logger::dieMessage("UNKOWN_MISSION");
		}

		Core::getTPL()->assign("mission", NS::getMissionName($mode));
		Core::getTPL()->assign("mode", $mode);
		Core::getTPL()->assign("distance", fNumber($distance));
		Core::getTPL()->assign("speed", fNumber($temp["maxspeed"]));
		Core::getTPL()->assign("consume", fNumber($data["consumption"]));
		Core::getTPL()->assign("start", NS::getPlanet()->getCoords(false));
		Core::getTPL()->assign("target", $data["galaxy"].":".$data["system"].":".$data["position"]);
		Core::getTPL()->assign("arrival", Date::timeToString(1, $time + time()));
		Core::getTPL()->assign("return", Date::timeToString(1, $time * 2 + time()));

		$fleet = array();
		foreach($data["ships"] as $key => $value)
		{
			$fleet[$key]["name"] = Core::getLanguage()->getItem($value["name"]);
			$fleet[$key]["quantity"] = getUnitQuantityStr($value);

			/*
			if($value["mode"] == UNIT_TYPE_ARTEFACT)
			{
				foreach($value["art_ids"] as $art)
				{
					Artefact::attachToFleet($art["artid"], NS::getUser()->get("userid"), NS::getUser()->get("curplanet"));
				}
			}
			*/
		}
		Core::getTPL()->addLoop("fleet", $fleet);
		Core::getTPL()->display("missions4");
		return $this;
	}
}
?>