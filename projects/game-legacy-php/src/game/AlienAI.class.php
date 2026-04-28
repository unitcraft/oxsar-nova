<?php

/**
 * Class for handling the alien AI.
 *
 * Oxsar http://oxsar.ru
 *
 *
 */

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class AlienAI
{
	public static function isAttackTime($time = null)
	{
		if(is_null($time))
		{
			$time = time();
		}
		return date("w", $time) == 4; // && date("j", $time) <= 20;
	}

	public static function isAlienPosition($galaxy, $system)
	{
		static $cache = array();
		if(isset($cache[$galaxy][$system]))
		{
			return $cache[$galaxy][$system];
		}
		$count = 0;
		$cache_key = "isAlienPosition:$galaxy,$system";
		if(NS::getMCH()->get($cache_key, $count))
		{
			return $cache[$galaxy][$system] = (int)$count;
		}
		$count = sqlSelectField("events e", "count(*)",
			" JOIN ".PREFIX."galaxy g ON g.planetid = e.destination ",
			"g.galaxy = ".sqlVal($galaxy)." AND g.system = ".sqlVal($system)
			. " AND e.mode = ".sqlVal(EVENT_ALIEN_HOLDING)
			. " AND e.processed = ".EVENT_PROCESSED_WAIT
			. " LIMIT 1"
			);
		if(!$count)
		{
			$count = sqlSelectField("events e", "count(*)",
				" JOIN ".PREFIX."galaxy g ON g.moonid = e.destination ",
				"g.galaxy = ".sqlVal($galaxy)." AND g.system = ".sqlVal($system)
				. " AND e.mode = ".sqlVal(EVENT_ALIEN_HOLDING)
				. " AND e.processed = ".EVENT_PROCESSED_WAIT
				. " LIMIT 1"
				);
		}
		$value = $count;
        NS::getMCH()->set($cache_key, $value, 60*10);
		return $cache[$galaxy][$system] = (int)$count;
	}

	protected static function generateMission($target = null, $params = array())
	{
		if(is_null($target))
		{
			if( isset($params["mode"]) && $params["mode"] == EVENT_ALIEN_GRAB_CREDIT )
			{
				$target = self::findCreditTarget();
			}
			else
			{
				$target = self::findTarget();
				if( !$target && !isset($params["mode"]) )
				{
					$params["mode"] = EVENT_ALIEN_GRAB_CREDIT;
					$params["power_scale"] = randFloatRange(1.5, 2.0);
					$target = self::findCreditTarget();
				}
			}
		}
		if(!$target)
		{
			return false;
		}

		$target_ships = self::loadPlanetShips($target["planetid"]);
		$power_scale = isset($params["power_scale"]) ? $params["power_scale"] : randFloatRange(0.9, 1.1);
		if(isset($params["alien_fleet"]))
		{
			$alien_ships = $params["alien_fleet"];
		}
		else
		{
			if(isset($params["alien_available_ships"]))
			{
				$alien_available_ships = $params["alien_available_ships"];
			}
			else
			{
				$alien_available_ships = array_flip(array(UNIT_A_CORVETTE, UNIT_A_SCREEN, UNIT_A_PALADIN, UNIT_A_FRIGATE, UNIT_A_TORPEDOCARIER));
			}
			$alien_ships = self::generateFleet($target_ships, $alien_available_ships, $power_scale);
		}
		if($alien_ships)
		{
			$mode = isset($params["mode"]) ? $params["mode"] : (mt_rand(1, 100) <= 90 ? EVENT_ALIEN_ATTACK : EVENT_ALIEN_FLY_UNKNOWN);
			if( $mode != EVENT_ALIEN_ATTACK && $mode != EVENT_ALIEN_ATTACK_CUSTOM )
			{
				$planet = new Planet($target["planetid"], null);
				$target["metal"] += $planet->getProd("metal") * mt_rand(24*2, 24*7);
				$target["silicon"] += $planet->getProd("silicon") * mt_rand(24*1, 24*4);
				$target["hydrogen"] += $planet->getProd("hydrogen") * mt_rand(12, 24*2);
			}
			$coords = getCoordsAndPlanetname($target["planetid"]);
			$data = array(
				"ships" => $alien_ships,
				"metal" => ceil( (5000000 + $target["metal"]) * randFloatRange(0.9, 1.1) ),
				"silicon" => ceil( (2000000 + $target["silicon"]) * randFloatRange(0.9, 1.1) ),
				"hydrogen" => ceil( (1000000 + $target["hydrogen"]) * randFloatRange(0.9, 1.1) ),
				"galaxy" => $coords["galaxy"],
				"system" => $coords["system"],
				"position" => $coords["position"],
				"sgalaxy" => 0,
				"ssystem" => 0,
				"sposition" => 0,
				"maxspeed" => MAX_SPEED_POSSIBLE,
				"consumption" => 0,
				"capacity" => 0,
				"time" => randRoundRange(ALIEN_FLY_MIN_TIME, ALIEN_FLY_MAX_TIME),
				"duration" => randRoundRange(ALIEN_HALTING_MIN_TIME, ALIEN_HALTING_MAX_TIME),
				"control_times" => 1,
				"alien_actor" => 1,
				);

			$research = self::loadUserResearches($target["userid"]);
			self::shuffleKeyValues( $research, array(UNIT_GUN_TECH, UNIT_SHIELD_TECH, UNIT_SHELL_TECH) );
			self::shuffleKeyValues( $research, array(UNIT_BALLISTICS_TECH, UNIT_MASKING_TECH) );
			self::shuffleKeyValues( $research, array(UNIT_LASER_TECH, UNIT_ION_TECH, UNIT_PLASMA_TECH) );
			foreach($research as $techid => $level)
			{
				$data["add_tech_".$techid] = max(0, mt_rand( floor($level * 0.7), $level < 3 ? $level : $level + 1 ));
			}

			// debug
			if(!empty($params["debug"]))
			{
				$data["real_userid"] = $target["userid"];
				$data["real_username"] = $target["username"];
				$data["real_planetid"] = $target["planetid"];
				$target["userid"] = 1;
				$target["planetid"] = 1;
			}

			return array(
				"userid" => $target["userid"],
				"destination" => $target["planetid"],
				"mode" => $mode,
				"data" => $data,
				);
		}
		return false;
	}

	public static function generateAttack($userid, $planetid, $fly_time = 3600, $alien_fleet = null)
	{
		$mission = self::generateMission(array("userid" => $userid, "planetid" => $planetid), array(
			"mode" => EVENT_ALIEN_ATTACK_CUSTOM,
			"alien_fleet" => self::setupFleet($alien_fleet ? $alien_fleet : array(UNIT_A_CORVETTE => 1)),
			));
		if($mission)
		{
			$mission["data"]["time"] = $fly_time;
			$eventid = EventHandler::getEH()->addEvent($mission["mode"], time() + $mission["data"]["time"],
				null, // NS::getUser()->get("curplanet"),
				0, // NS::getUser()->get("userid"),
				$mission["destination"],
				$mission["data"],
				null, // $protected
				null, // $start_time
				null // $parent_eventid
				);
			return array("mission" => $mission, "eventid" => $eventid);
		}
		return false;
	}

	public static function checkAlientNeeds()
	{
        if(!ALIEN_ENABLED){
            return false;
        }
		$debug = isAdmin();
		$params = array("debug" => $debug);
		$count = sqlSelectField("events", "count(*)", "", "mode in (".sqlArray(EVENT_ALIEN_FLY_UNKNOWN, EVENT_ALIEN_HOLDING, EVENT_ALIEN_ATTACK, EVENT_ALIEN_HALT).") AND processed=".EVENT_PROCESSED_WAIT);
		$is_attack_time = self::isAttackTime();
		if($count < ($is_attack_time ? ALIEN_ATTACK_TIME_FLEETS_NUMBER : ALIEN_NORMAL_FLEETS_NUMBER) || $debug)
		{
		}
		else
		{
			$params["mode"] = EVENT_ALIEN_GRAB_CREDIT;
		}
		if($is_attack_time)
		{
			$params["power_scale"] = randFloatRange(1.5, 2.0);
		}
		$mission = self::generateMission(null, $params);
		if($mission)
		{
			$eventid = EventHandler::getEH()->addEvent($mission["mode"], time() + $mission["data"]["time"],
				null, // NS::getUser()->get("curplanet"),
				0, // NS::getUser()->get("userid"),
				$mission["destination"],
				$mission["data"],
				null, // $protected
				null, // $start_time
				null // $parent_eventid
				);

			if($eventid && $mission["mode"] != EVENT_ALIEN_GRAB_CREDIT && (mt_rand(1, 100) <= 60 || $debug))
			{
				if(mt_rand(1, 100) <= 60)
				{
					$time = randRoundRange(ALIEN_CHANGE_MISSION_MIN_TIME, ALIEN_CHANGE_MISSION_MAX_TIME);
					$time = min($time, $mission["data"]["time"] - 10);
				}
				else
				{
					$time = randRoundRange($mission["data"]["time"] - 30, $mission["data"]["time"] - 10);
				}
				$time += time();
				EventHandler::getEH()->addEvent(EVENT_ALIEN_CHANGE_MISSION_AI, $time, null, 0, $mission["destination"],
					array(
						"control_times" => 1,
						"alien_actor" => 1,
					), null, null,
					$eventid // $parent_eventid
				);
				return array(
					"mission" => $mission,
					"eventid" => $eventid,
					"change_time" => $time,
					"change_eventid" => $eventid,
					);
			}
			return array(
				"mission" => $mission,
				"eventid" => $eventid,
				);
		}
		return false;
	}

	private static function shuffleKeyValues(&$r, $keys)
	{
		$values = array();
		foreach( $keys as $key)
		{
			$values[] = isset($r[$key]) ? $r[$key] : 0;
		}
		shuffle($values);
		$i = 0;
		foreach( $keys as $key)
		{
			$r[$key] = $values[$i++];
		}
	}

	protected static function loadPlanetShips($planetid)
	{
		$ships = array();
		$result = sqlSelect("unit2shipyard", "*", "", "planetid=".sqlVal($planetid)." AND unitid NOT IN (".sqlArray(UNIT_SOLAR_SATELLITE).")");
		while($row = sqlFetch($result))
		{
			$ships[$row["unitid"]] = $row["quantity"];
		}
		sqlEnd($result);
		return $ships;
	}

	protected static function loadUserResearches($userid)
	{
		$tech_list = array(UNIT_EXPO_TECH, UNIT_GUN_TECH, UNIT_SHIELD_TECH, UNIT_SHELL_TECH, UNIT_SPYWARE,
			UNIT_BALLISTICS_TECH, UNIT_MASKING_TECH, UNIT_LASER_TECH, UNIT_ION_TECH, UNIT_PLASMA_TECH);

		$researches = array();
		foreach($tech_list as $tech_id)
		{
			$researches[$tech_id] = 0;
		}

		$result = sqlSelect("research2user", array("buildingid", "level"), "",
			"userid = ".sqlVal($userid)." AND buildingid IN (" . sqlArray($tech_list) . ")");
		while ($row = sqlFetch($result))
		{
			$researches[$row["buildingid"]] = $row["level"];
		}
		sqlEnd($result);
		return $researches;
	}

	protected static function findCreditTarget()
	{
		$user_credit_edge = ALIEN_GRAB_MIN_CREDIT;
		$user_ships_edge = 300000;
		$planet_ships_edge = 10000;
		$row = sqlQueryRow("
			SELECT
			p.userid,
			-- u.username,
			-- from_unixtime(u.last) as l,
			p.metal, p.silicon, p.hydrogen,
			ship.planetid,
			sum(ship.quantity) as quantity
			FROM ".PREFIX."user u
			JOIN ".PREFIX."planet p ON p.userid = u.userid
			JOIN ".PREFIX."unit2shipyard ship ON ship.planetid = p.planetid
			-- LEFT JOIN ".PREFIX."galaxy g ON g.planetid = p.planetid
			-- LEFT JOIN ".PREFIX."galaxy gm ON gm.moonid = p.planetid
			WHERE u.last > unix_timestamp(now()) - 60*30
			AND u.umode = 0
			AND u.credit > ".sqlVal($user_credit_edge)."
			AND u.u_count > ".sqlVal($user_ships_edge)."
			-- AND ship.unitid NOT IN (".sqlArray(UNIT_ESPIONAGE_SENSOR, UNIT_SOLAR_SATELLITE, UNIT_DEATH_STAR).")
			AND u.userid NOT IN (
				SELECT userid FROM ".PREFIX."planet p2
				JOIN ".PREFIX."events e2 ON p2.planetid = e2.destination
				WHERE e2.mode IN (".sqlArray(EVENT_ALIEN_GRAB_CREDIT).") AND e2.start > unix_timestamp(now()) - ".sqlVal(ALIEN_GRAB_CREDIT_INTERVAL)."
				)
			GROUP BY ship.planetid
			HAVING sum(ship.quantity) > ".sqlVal($planet_ships_edge)."
			ORDER BY rand()
			LIMIT 1
			-- sum(ship.quantity) DESC
			");
		return $row;
	}

	protected static function findTarget()
	{
		$user_ships_edge = 1000;
		$planet_ships_edge = 100;
		$row = sqlQueryRow("
			SELECT
			p.userid,
			-- u.username,
			-- from_unixtime(u.last) as l,
			p.metal, p.silicon, p.hydrogen,
			ship.planetid,
			sum(ship.quantity) as quantity
			FROM ".PREFIX."user u
			JOIN ".PREFIX."planet p ON p.userid = u.userid
			JOIN ".PREFIX."unit2shipyard ship ON ship.planetid = p.planetid
			-- LEFT JOIN ".PREFIX."galaxy g ON g.planetid = p.planetid
			-- LEFT JOIN ".PREFIX."galaxy gm ON gm.moonid = p.planetid
			WHERE u.last > unix_timestamp(now()) - 60*30
			AND u.umode = 0
			AND u.u_count > ".sqlVal($user_ships_edge)."
			".(mt_rand(1, 100) <= 10 ? "AND ship.unitid=".UNIT_SOLAR_SATELLITE : "AND ship.unitid!=".UNIT_SOLAR_SATELLITE)."
			-- AND ship.unitid NOT IN (".sqlArray(UNIT_ESPIONAGE_SENSOR, UNIT_SOLAR_SATELLITE, UNIT_DEATH_STAR).")
			AND u.userid NOT IN (
				SELECT userid FROM ".PREFIX."planet p2
				JOIN ".PREFIX."events e2 ON p2.planetid = e2.destination
				WHERE e2.mode IN (".sqlArray(EVENT_ALIEN_FLY_UNKNOWN, EVENT_ALIEN_ATTACK, EVENT_ALIEN_HOLDING, EVENT_ALIEN_HALT).") AND e2.start > unix_timestamp(now()) - ".sqlVal(ALIEN_ATTACK_INTERVAL)."
				)
			GROUP BY ship.planetid
			HAVING sum(ship.quantity) > ".sqlVal($planet_ships_edge)."
			ORDER BY rand()
			LIMIT 1
			-- sum(ship.quantity) DESC
			");
		return $row;
	}

	public static function setupFleet(array $fleet_ships)
	{
		foreach($fleet_ships as $id => $ship)
		{
			if(!is_array($fleet_ships[$id]))
			{
				$fleet_ships[$id] = array("quantity" => $fleet_ships[$id]);
			}
			if($fleet_ships[$id]["quantity"] <= 0)
			{
				unset($fleet_ships[$id]);
			}
		}
		if(!$fleet_ships)
		{
			return array();
		}
		$ships = array();
		$result = sqlSelect("construction", "buildingid, name", "", "buildingid in (".sqlArray(array_keys($fleet_ships)).") AND mode=".UNIT_TYPE_FLEET);
		while($row = sqlFetch($result))
		{
			$id = $row["buildingid"];
			$ships[$id] = array();
			$ships[$id]["id"] = $id;
			$ships[$id]["name"] = $row["name"];
			$ships[$id]["quantity"] = max(1, $fleet_ships[$id]["quantity"]);
			$ships[$id]["damaged"] = isset($fleet_ships[$id]["damaged"]) ? clampVal($fleet_ships[$id]["damaged"], 0, $ships[$id]["quantity"]) : 0;
			$ships[$id]["shell_percent"] = isset($fleet_ships[$id]["shell_percent"]) ? clampVal($fleet_ships[$id]["shell_percent"], 0, 100) : 100;
		}
		sqlEnd($result);
		return $ships;
	}

	public static function generateFleet($target_ships, $available_ships, $scale = 1, $params = array())
	{
		$use_shell_power = false;
		$use_shield_power = true;
        $find_mode = !empty($params["find_mode"]);

		$target_attack = $target_shields = $target_shells = 0;
		$max_debris = isset($params["max_derbis"]) ? $params["max_derbis"] : ALIEN_FLEET_MAX_DERBIS;
		$death_star_debris = 0;
        $ignore_ships = array(UNIT_ESPIONAGE_SENSOR);
        $special_ships = array(UNIT_DEATH_STAR, UNIT_SHIP_TRANSPLANTATOR);
        $target_avg_quantity = 0;
        $target_avg_quantity_count = 0;

		$select = array("d.unitid", "d.attacker_attack as attack", "d.attacker_shield as shield", "c.name", "c.basic_metal", "c.basic_silicon");
		$join = "INNER JOIN ".PREFIX."construction c ON c.buildingid = d.unitid";
		if($target_ships){
			$result = sqlSelect("ship_datasheet d", $select, $join, "unitid in (". sqlArray(array_keys($target_ships)) .")");
			while($row = sqlFetch($result)){
                if($row["unitid"] == UNIT_DEATH_STAR){
                    $death_star_debris = ($row["basic_metal"] + $row["basic_silicon"]) * 0.5; // debris is 50%
                }
                $add_q = $special = false;
                if(in_array($row["unitid"], $ignore_ships)){
					$row["basic_metal"] = 0;
					$row["basic_silicon"] = 0;
					$row["shield"] = 0;
					$row["attack"] = 0;
                }elseif(in_array($row["unitid"], $special_ships)){
                    $special = true;
					$row["basic_metal"] *= 0.2;
					$row["basic_silicon"] *= 0.2;
					$row["shield"] *= 0.2;
					$row["attack"] *= 0.2;
				}elseif($row["unitid"] == UNIT_SHIP_ARMORED_TERRAN){
					$row["basic_metal"] *= 0.001;
					$row["basic_silicon"] *= 0.001;
					$row["shield"] *= 0.001;
					$row["attack"] *= 0.001;
				}else{
                    $add_q = true;
                }
				$quantity = is_numeric($target_ships[$row["unitid"]]) ? $target_ships[$row["unitid"]] : $target_ships[$row["unitid"]]["quantity"];
				if($find_mode)
				{
					$quantity = min($quantity, mt_rand(50, 100));
				}
                if($special){
                    $quantity = min($quantity, mt_rand(50, 100));
                }
                if($add_q){
                    $target_avg_quantity += $quantity;
                    $target_avg_quantity_count++;
                }
				$target_attack += $row["attack"] * $quantity;
				$target_shields += $row["shield"] * $quantity;
				$target_shells += (($row["basic_metal"] + $row["basic_silicon"]) / 10) * 0.3 * $quantity;
			}
			sqlEnd($result);
		}
        if($target_avg_quantity_count > 0){
            $target_avg_quantity /= $target_avg_quantity_count;
        }

		$target_power = $target_attack;
		$target_power += $use_shield_power ? $target_shields : 0;
		$target_power += $use_shell_power ? $target_shells : 0;
		$target_power = max(100, $target_power * $scale);

		// unset($available_ships[UNIT_ESPIONAGE_SENSOR]);

		$max_death_stars = 0;
		$max_espionage_sensors = mt_rand(10, 100);

		if(isset($target_ships[UNIT_DEATH_STAR]))
		{
			$available_ships[UNIT_DEATH_STAR] = 1;
			$quantity = is_numeric($target_ships[UNIT_DEATH_STAR]) ? $target_ships[UNIT_DEATH_STAR] : $target_ships[UNIT_DEATH_STAR]["quantity"];
            if(mt_rand(0, 10)){
                $max_death_stars = min(100, ceil($quantity * 0.3), ceil($max_debris * 0.5 / $death_star_debris));
            }else{
                $max_death_stars = min(100, ceil($quantity * 0.9), ceil($max_debris * 0.9 / $death_star_debris));
            }
		}
		else if(mt_rand(1, 100) <= 10)
		{
			$available_ships[UNIT_DEATH_STAR] = 1;
			$max_death_stars = 1;
		}
        if(mt_rand(0, 50)){
            unset($available_ships[UNIT_SHIP_ARMORED_TERRAN]);
        }

		if(!$available_ships){
			return false;
		}

		$result = sqlSelect("ship_datasheet d", $select, $join, "unitid in (".sqlArray(array_keys($available_ships)).") AND c.mode=".UNIT_TYPE_FLEET);
		$available_ships = array();
		while($row = sqlFetch($result))
		{
			$id = $row["unitid"];
			$available_ships[$id] = array();
			$available_ships[$id]["derbis"] = ($row["basic_metal"] + $row["basic_silicon"]) * 0.5; // debris is 50%
			if(in_array($id, $ignore_ships)){
                $row["basic_metal"] = 0;
                $row["basic_silicon"] = 0;
                $row["shield"] = 0;
                $row["attack"] = 0;
            }elseif(in_array($id, $special_ships)){
				$row["basic_metal"] *= 0.2;
				$row["basic_silicon"] *= 0.2;
				$row["shield"] *= 0.2;
				$row["attack"] *= 0.2;
			}elseif($id == UNIT_SHIP_ARMORED_TERRAN){
				// $row["basic_metal"] *= 0.01;
				// $row["basic_silicon"] *= 0.01;
				// $row["shield"] *= 0.01;
				// $row["attack"] *= 0.01;
            }
			$available_ships[$id]["shell"] = max(10, ($row["basic_metal"] + $row["basic_silicon"]) / 10) * 0.3;
			$available_ships[$id]["shield"] = max(10, $row["shield"]);
			$available_ships[$id]["attack"] = max(10, $row["attack"]);
			$available_ships[$id]["real_attack"] = $row["attack"];
			$available_ships[$id]["name"] = $row["name"];
		}
		sqlEnd($result);

		if(!$available_ships)
		{
			return false;
		}

		$power = 0;
		$debris = 0;
		$fleet = array();
		$max_single_units = 0;
		$ignore_count_units = array_flip(array(UNIT_DEATH_STAR, UNIT_SHIP_ARMORED_TERRAN, UNIT_SHIP_TRANSPLANTATOR,
            UNIT_ESPIONAGE_SENSOR, UNIT_SOLAR_SATELLITE, UNIT_A_SCREEN));
        $max_special_scale = 1.0 / max(1, $scale);
		$max_ship_mass = array(
            UNIT_DEATH_STAR => ($find_mode ? 0.03 : 0.3) * $max_special_scale,
            UNIT_SHIP_TRANSPLANTATOR => ($find_mode ? 0.05 : 0.1) * $max_special_scale,
            UNIT_SHIP_ARMORED_TERRAN => 0.001,
            UNIT_ESPIONAGE_SENSOR => ($find_mode ? 0.2 : 0.5) * $max_special_scale,
            UNIT_A_SCREEN => 0.5 / $max_special_scale);
		do
		{
			$id = array_rand($available_ships);
			if (!isset($fleet[$id]))
			{
				$fleet[$id]["id"] = $id;
				$fleet[$id]["quantity"] = 0;
				$fleet[$id]["damaged"] = 0;
				$fleet[$id]["shell_percent"] = 100;
				$fleet[$id]["name"] = $available_ships[$id]["name"];
			}
			if($fleet[$id]["quantity"] < mt_rand(18000, 20000)
				// && ($id != UNIT_DEATH_STAR || $fleet[$id]["quantity"] < $max_death_stars)
				// && ($id != UNIT_ESPIONAGE_SENSOR || $fleet[$id]["quantity"] < $max_espionage_sensors)
			)
			{
                $target_quantity = 0;
                if($target_ships && isset($target_ships[$id])){
                    if(is_numeric($target_ships[$id])){
                        $target_quantity = $target_ships[$id];
                    }elseif(isset($target_ships[$id]["quantity"])){
                        $target_quantity = $target_ships[$id]["quantity"];
                    }
                    $target_quantity = min($target_quantity, $target_avg_quantity * pow(max(1, $scale), 2) + mt_rand(10, 100));
                }

                $over = false;
				$inc_unit = max(1, mt_rand(max(1, ceil($max_single_units * 0.1)), max(5, ceil($max_single_units * 0.3))));
				if(	($target_quantity > 0 && $fleet[$id]["quantity"] + $inc_unit > $target_quantity * 3)
                    || $fleet[$id]["quantity"] + $inc_unit > 15000
					|| $id == UNIT_SHIP_ARMORED_TERRAN
					|| ( $id == UNIT_DEATH_STAR && $fleet[$id]["quantity"] + $inc_unit > $max_death_stars )
					|| ( $id == UNIT_SHIP_TRANSPLANTATOR && $fleet[$id]["quantity"] + $inc_unit > 1+$max_death_stars*2 )
				){
                    if($id == UNIT_SHIP_ARMORED_TERRAN){
                        $fleet[$id]["quantity"] = 1;
                        $fleet = array(
                            $id => $fleet[$id],
                        );
                        break;
                    }
                    // $over = true;
					$inc_unit = $fleet[$id]["quantity"] > 0 ? 0 : 1;
                    unset($available_ships[$id]);
                }elseif(
                    ( isset($max_ship_mass[$id]) && $fleet[$id]["quantity"] + $inc_unit > ceil($max_single_units * $max_ship_mass[$id]) )
					|| ( $available_ships[$id]["real_attack"] <= 10 && $fleet[$id]["quantity"] + $inc_unit > $max_single_units * 0.01 )
				)
				{
                    $inc_unit = 1;
				}
				$fleet[$id]["quantity"] += $inc_unit;

                if($over){ // in_array($id, $ignore_ships)){
                    $power += $target_power * 0.05;
                }else{
                    $power += $use_shell_power ? $available_ships[$id]["shell"] * $inc_unit : 0;
                    $power += $use_shield_power ? $available_ships[$id]["shield"] * $inc_unit : 0;
                    $power += $available_ships[$id]["attack"] * $inc_unit;
                }
                $debris += $available_ships[$id]["debris"] * $inc_unit;
			}
			else
			{
				$power += $target_power * 0.05;
			}
			if($available_ships[$id]["real_attack"] > 10 && !isset($ignore_count_units[$id]))
			{
				$max_single_units = max($max_single_units, $fleet[$id]["quantity"]);
			}
		}
		while ($power < $target_power && $debris < $max_debris && $available_ships);

		if(isset($params["damaged"]))
		{
			foreach($fleet as $id => $ship)
			{
				if(!isset($params["damaged"]["chance"]) || mt_rand(1, 100) <= $params["damaged"]["chance"])
				{
					$p = mt_rand(round($params["damaged"]["quantity_percent"][0]), round($params["damaged"]["quantity_percent"][1]));
					$fleet[$id]["damaged"] = clampVal(round($p * 0.01 * $fleet[$id]["quantity"]), 0, $fleet[$id]["quantity"]);

					$p = mt_rand(round($params["damaged"]["shell_percent"][0]), round($params["damaged"]["shell_percent"][1]));
					$fleet[$id]["shell_percent"] = clampVal($p, 1, 100);
				}
			}
		}

		return $fleet;
	}

	public static function onAttackEvent($event)
	{
		return EventHandler::getEH()->attack($event, $event["data"]);
	}

	public static function onGrabCreditEvent($event)
	{
		return self::onFlyUnknownEvent($event);
	}

	public static function onFlyUnknownEvent($event)
	{
		$debug = isAdmin($event["destuser"]);
		$user_row = sqlSelectRow("user", "credit, u_count", "", "userid=".sqlVal($event["destuser"]));

		if($event["mode"] == EVENT_ALIEN_ATTACK_CUSTOM && $user_row["u_count"] > 10)
		{
			return self::onAttackEvent($event);
		}

		$user_credit = $user_row[credit];
		if(/*$debug ||*/ ($user_credit > ALIEN_GRAB_MIN_CREDIT
			&& $event["mode"] != EVENT_ALIEN_ATTACK_CUSTOM
			&& ($event["mode"] == EVENT_ALIEN_GRAB_CREDIT || mt_rand(1, 100) <= 10)))
		{
			$grab_credit = round($user_credit * 0.01 * randFloatRange(ALIEN_GRAB_CREDIT_MIN_PERCENT, ALIEN_GRAB_CREDIT_MAX_PERCENT), 2);
			// echo "GRAB CREDIR 2: $grab_credit of $user_credit, userid: ".$event["destuser"]."\n";
			if($grab_credit > 0)
			{
				$res_log = NS::updateUserRes(array(
					"block_minus"	=> true,
					"type" 			=> RES_UPDATE_ALIEN_GRAB_CREDIT,
					"event_mode"	=> $event["mode"],
					"ownerid"		=> $event["eventid"],
					"userid"		=> $event["destuser"],
					// "planetid"		=> $planetid,
					"credit"		=> - $grab_credit,
				));

				if(empty($res_log["minus_blocked"]))
				{
					new AutoMsg(
						MSG_CREDIT,
						$event["destuser"],
						time(),
						array(
							'credits'	=> $grab_credit,
							'msg' 		=> 'MSG_CREDIT_ALIEN_GRAB' )
					);

					if( !$debug && mt_rand(1, 100) <= 90 )
					{
						return true;
					}
				}
			}
		}

		if($debug || $event["mode"] == EVENT_ALIEN_ATTACK_CUSTOM || (!$grab_credit && mt_rand(1, 100) <= 5))
		{
			$gift_res_scale = randFloatRange(0.7, 1.0);
			$res_log = NS::updateUserRes(array(
				"block_minus"	=> true,
				"type" 			=> RES_UPDATE_ALIEN_GIFT_RESOURSES,
				"event_mode"	=> $event["mode"],
				"ownerid"		=> $event["eventid"],
				"userid"		=> $event["destuser"],
				"planetid"		=> $event["destination"],
				"metal"			=> ($gift_metal = round($event["data"]["metal"] * $gift_res_scale)),
				"silicon"		=> ($gift_silicon = round($event["data"]["silicon"] * $gift_res_scale)),
				"hydrogen"		=> ($gift_hydrogen = round($event["data"]["hydrogen"] * $gift_res_scale)),
			));

			$event["data"]["metal"] -= $gift_metal;
			$event["data"]["silicon"] -= $gift_silicon;
			$event["data"]["hydrogen"] -= $gift_hydrogen;

			new AutoMsg(
				MSG_ALIEN_RESOURSES_GIFT,
				$event["destuser"],
				time(),
				array(
					"metal"		=> $gift_metal,
					"silicon"	=> $gift_silicon,
					"hydrogen"	=> $gift_hydrogen,
					"planetid"  => $event["destination"],
				)
			);

			// if($event["mode"] == EVENT_ALIEN_ATTACK_CUSTOM)
			{
				return true;
			}
		}

		if($debug || $event["mode"] == EVENT_ALIEN_ATTACK_CUSTOM || (!$grab_credit && mt_rand(1, 100) <= 5))
		{
			$max_gift_credit = ALIEN_MAX_GIFT_CREDIT * randFloatRange(0.98, 1.02);
			$gift_credit = round(min($max_gift_credit, $user_credit * 0.01 * randFloatRange(ALIEN_GIFT_CREDIT_MIN_PERCENT, ALIEN_GIFT_CREDIT_MAX_PERCENT)), 2);
			if($event["mode"] == EVENT_ALIEN_ATTACK_CUSTOM)
			{
				$gift_credit = max(1, $gift_credit);
			}
			// echo "GRAB CREDIR 2: $grab_credit of $user_credit, userid: ".$event["destuser"]."\n";
			if($gift_credit > 0)
			{
				$res_log = NS::updateUserRes(array(
					"block_minus"	=> true,
					"type" 			=> RES_UPDATE_ALIEN_GIFT_CREDIT,
					"event_mode"	=> $event["mode"],
					"ownerid"		=> $event["eventid"],
					"userid"		=> $event["destuser"],
					// "planetid"		=> $planetid,
					"credit"		=> $gift_credit,
				));

				new AutoMsg(
					MSG_CREDIT,
					$event["destuser"],
					time(),
					array(
						'credits'	=> $gift_credit,
						'msg' 		=> 'MSG_CREDIT_ALIEN_GIFT' )
				);
			}

			if($event["mode"] == EVENT_ALIEN_ATTACK_CUSTOM)
			{
				return true;
			}
		}

		if(mt_rand(1, 100) <= 10)
		{
			$mission = self::generateMission(null, array("debug" => $debug, "mode" => $event["mode"]));
			if($mission)
			{
				$data["control_times"] = $event["data"]["control_times"] + 1;
				sqlUpdate("events", array(
						"start" => time(),
						"time" => time() + $mission["data"]["time"],
						"data" => serialize($mission["data"]),
						"planetid" => $event["destination"],
						"destination" => $mission["destination"],
						"prev_rc" => null,
						"processed" => EVENT_PROCESSED_WAIT,
						"processed_mode" => $event["mode"],
						"processed_time" => time(),
						), "eventid=".sqlVal($event["eventid"])
					);

				if(mt_rand(1, 100) <= 90)
				{
					if(mt_rand(1, 100) <= 60)
					{
						$time = randRoundRange(ALIEN_CHANGE_MISSION_MIN_TIME, ALIEN_CHANGE_MISSION_MAX_TIME);
						$time = min($time, $mission["data"]["time"] - 10);
					}
					else
					{
						$time = randRoundRange($mission["data"]["time"] - 30, $mission["data"]["time"] - 10);
					}
					EventHandler::getEH()->addEvent(EVENT_ALIEN_CHANGE_MISSION_AI, time() + $time, null, 0, $mission["destination"],
						array(
							"control_times" => 1,
							"alien_actor" => 1,
						), null, null,
						$event["eventid"] // $parent_eventid
					);
				}

				return true;
			}
		}

		if( ($debug || (!$gift_credit && !$gift_res_scale)) && mt_rand(1, 100) <= ($grab_credit || self::isAttackTime() ? 90 : 50)
			// || AlienAI::isAttackTime($event["time"])
			)
		{
			return self::onAttackEvent($event);
		}

		return self::onHaltEvent($event);
	}

	public static function onHaltEvent($event)
	{
		$data = $event["data"];
		$time = time() + max(1, $data["duration"]);
		$new_eventid = EventHandler::getEH()->addEvent(EVENT_ALIEN_HOLDING, $time, $event["planetid"], $event["userid"], $event["destination"], $data, null, null,
			$event["eventid"] // $parent_eventid
			);
		if($new_eventid && isAdmin($event["destuser"]))
		{
			EventHandler::getEH()->addEvent(EVENT_ALIEN_HOLDING_AI, time() + mt_rand(5, 10), $event["planetid"], $event["userid"], $event["destination"],
				array(
					"control_times" => 1,
					"alien_actor" => 1,
					"paid_sum_credit" => 0,
					"paid_times" => 0,
				), null, null,
				$new_eventid // $parent_eventid
			);
		}
		new AutoMsg( MSG_ALIEN_HALTING, $event["destuser"], time(), array(
			"galaxy" => $data["galaxy"],
			"system" => $data["system"],
			"position" => $data["position"],
			"planetid" => $data["planetid"],
			) );

		return true;
	}

	public static function onHoldingEvent($event)
	{
		self::checkAlientNeeds();

		// do nothink, alien go away
		return true;
	}

	public static function onChangeMissionAIEvent($event)
	{
		$parent_event = sqlSelectRow("events", "*", "",
			"eventid=".sqlVal($event["parent_eventid"])
			. " AND processed = ".EVENT_PROCESSED_WAIT
			// . " AND mode in (".sqlArray(EVENT_ALIEN_FLY_UNKNOWN, EVENT_ALIEN_ATTACK).")"
			);
		if($parent_event)
		{
			$parent_event["data"] = unserialize($parent_event["data"]);

			$debug = isAdmin($event["destuser"]);
			if($parent_event["time"] - time() >= ALIEN_CHANGE_MISSION_MIN_TIME || !in_array($parent_event["mode"], array(EVENT_ALIEN_ATTACK, EVENT_ALIEN_ATTACK_CUSTOM, EVENT_ALIEN_FLY_UNKNOWN)))
			{
				$mission = self::generateMission(
								array(
									"userid" => isset($parent_event["data"]["real_userid"]) ? $parent_event["data"]["real_userid"] : $event["destuser"],
									"planetid" => isset($parent_event["data"]["real_planetid"]) ? $parent_event["data"]["real_planetid"] : $event["destination"]
								),
								array(
									"power_scale" => 1 + $parent_event["data"]["control_times"] * 1.5,
									"mode" => in_array($parent_event["mode"], array(EVENT_ALIEN_ATTACK, EVENT_ALIEN_ATTACK_CUSTOM, EVENT_ALIEN_FLY_UNKNOWN))
										? (mt_rand(1, 100) <= 50 ? EVENT_ALIEN_ATTACK : EVENT_ALIEN_FLY_UNKNOWN)
										: $parent_event["mode"],
									"debug" => $debug,
								) );
				if($mission)
				{
					$mission["data"]["control_times"] = $parent_event["data"]["control_times"] + 1;
					$mission["data"]["time"] = $parent_event["data"]["time"];

					sqlUpdate("events", array(
							"mode" => $mission["mode"],
							"data" => serialize($mission["data"]),
							), "eventid=".sqlVal($parent_event["eventid"])." AND processed=".EVENT_PROCESSED_WAIT
						);

					return true;
				}
			}
			else if($parent_event["time"] - time() < ALIEN_CHANGE_MISSION_MIN_TIME) // && $parent_event["mode"] == EVENT_ALIEN_ATTACK)
			{
				$mission = array();
				$mission["data"] = $parent_event["data"];
				$mission["data"]["control_times"] += 1;

				$time = randRoundRange(10, 50);
				$mission["data"]["time"] += $time;

				sqlUpdate("events", array(
						"time" => $parent_event["time"] + $time,
						), "eventid=".sqlVal($parent_event["eventid"])." AND processed=".EVENT_PROCESSED_WAIT
					);

				return true;
			}
		}
		return false;
	}

	public static function onHoldingAIEvent($event)
	{
		$parent_event = sqlSelectRow("events", "*", "", "eventid=".sqlVal($event["parent_eventid"])." AND processed = ".EVENT_PROCESSED_WAIT);
		if($parent_event)
		{
			$parent_event["data"] = unserialize($parent_event["data"]);

			$data = &$event["data"];
			$times = max(1, (int)$data["control_times"]);
			// $target_userid = isset($data["target_userid"]) ? $data["target_userid"] : $parent_event["destuser"];
			$paid_credit = isset($data["paid_credit"]) ? $data["paid_credit"] : 0;
			unset($data["paid_credit"]); // , $data["target_userid"]);

			$data["paid_sum_credit"] += $paid_credit;
			$paid_credit > 0 && $data["paid_times"]++;

			$variants["onExtractAlientShipsAI"] = 10;
			$variants["onUnloadAlienResoursesAI"] = 10;
			$variants["onRepairUserUnitsAI"] = 10;
			$variants["onAddUserUnitsAI"] = 10;
			$variants["onAddCreditsAI"] = 10;
			$variants["onAddArtefactAI"] = 10;
			$variants["onGenerateAsteroidAI"] = 10;
			$variants["onFindPlanetAfterBattleAI"] = 10;

			$sum_prob = 0;
			foreach($variants as $prob)
			{
				$sum_prob += $prob;
			}

			$log_data = array();
			$rand = randRoundRange(1, $sum_prob);
			foreach($variants as $method => $prob)
			{
				if($rand <= $prob && $prob > 0)
				{
					$ai = new AlienAI($event, $parent_event);
					$log_data = $ai->{$method}();
					break;
				}
				$rand -= $prob;
			}

			if(mt_rand(1, 100) <= 1)
			{
				// remove 1% of user ship unit to be analysed
			}

			$start_time = time();
			$end_time = $start_time + randRoundRange(min(60*60*12, 60*30*$times), max(60*60*24, 60*60*$times));
			$parent_time = $parent_event["time"]-2;
			if($end_time <= $parent_time || $parent_time - $start_time > 60*30)
			{
				$data["control_times"] = $data["control_times"] + 1;
				sqlUpdate("events", array(
					"start" => $start_time,
					"time" => min($end_time, $parent_time),
					"data" => serialize($data),
					"prev_rc" => null,
					"processed" => EVENT_PROCESSED_WAIT,
					"processed_mode" => $row["mode"],
					"processed_time" => time(),
					), "eventid=".sqlVal($event["eventid"])
				);
			}

			if(!empty($log_data["parent_changed"]) || $paid_credit > 0)
			{
				$end_time = $parent_event["time"] + 60*60*2 * $paid_credit / 50.0;
				$end_time = round(min($end_time, $parent_event["start"] + ALIEN_HALTING_MAX_REAL_TIME));
				sqlUpdate("events", array(
					"time" => $end_time,
					"data" => serialize($parent_event["data"]),
					// "prev_rc" => null,
					// "processed" => EVENT_PROCESSED_WAIT,
					// "processed_mode" => $row["mode"],
					// "processed_time" => time(),
					), "eventid=".sqlVal($parent_event["eventid"])." AND processed=".EVENT_PROCESSED_WAIT
				);
			}

			if(!mt_rand(0, 100))
			{
				self::checkAlientNeeds();
			}

			return true;
		}
		return false;
	}

	protected $event;
	protected $parent_event;

	public function __construct(&$event, &$parent_event)
	{
		$this->event = &$event;
		$this->parent_event = &$parent_event;
	}

	protected function onExtractAlientShipsAI($unload_resourses = false)
	{
		$event = &$this->event;
		$parent_event = &$this->parent_event;

		$alien_ships = &$parent_event["data"]["ships"];
		$id = array_rand($alien_ships);
		if($alien_ships[$id]["quantity"] < 2)
		{
			$found = false;
			foreach($alien_ships as $id => $ship)
			{
				if($ship["quantity"] >= 2)
				{
					$found = true;
					break;
				}
			}
			if(!$found)
			{
				return $this->onRepairUserUnitsAI();
			}
		}
		$data = array(
			"ships" => array(),
			"time" => mt_rand(30, 60),
			);
		$times = max(1, (int)$data["control_times"]);
		if($unload_resourses)
		{
			$good_bonus = false;
			foreach(array("metal", "silicon", "hydrogen") as $res)
			{
				$data[$res] = ceil(min($parent_event["data"][$res] * 0.7, $parent_event["data"][$res] * 0.1 * $times));
				$parent_event["data"][$res] = max(0, $parent_event["data"][$res] - $data[$res]);
				$good_bonus = $good_bonus || $data[$res] > 1000000;
			}
		}
		$quantity = min($alien_ships[$id]["quantity"] - 1, ceil($alien_ships[$id]["quantity"] * 0.01 * pow($times, 2) * ($good_bonus ? 0.3 : 1)));
		extractUnits($alien_ships[$id], $quantity, $data["ships"]);

		EventHandler::getEH()->addEvent(EVENT_POSITION, time() + $data["time"],
			$parent_event["destination"], // NS::getUser()->get("curplanet"),
			0, // NS::getUser()->get("userid"),
			$event["destination"],
			$data,
			null, // $protected
			null, // $start_time
			null // $parent_eventid
			);

		return array(
			"parent_changed" => true,
			);
	}

	protected function onUnloadAlienResoursesAI()
	{
		return $this->onExtractAlientShipsAI(true);
	}

	protected function onRepairUserUnitsAI()
	{
		$event = &$this->event;
		$parent_event = &$this->parent_event;

	}

	protected function onAddUserUnitsAI()
	{
		$event = &$this->event;
		$parent_event = &$this->parent_event;

	}

	protected function onAddCreditsAI()
	{
		$event = &$this->event;
		$parent_event = &$this->parent_event;

	}

	protected function onAddArtefactAI()
	{
		$event = &$this->event;
		$parent_event = &$this->parent_event;

	}

	protected function onGenerateAsteroidAI()
	{
		$event = &$this->event;
		$parent_event = &$this->parent_event;

	}

	protected function onFindPlanetAfterBattleAI()
	{
		$event = &$this->event;
		$parent_event = &$this->parent_event;

	}
}
