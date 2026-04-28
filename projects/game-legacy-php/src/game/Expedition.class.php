<?php
/**
* Expedition simulator.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class Expedition
{
	private $EH;
	private $event;

	private $sendBack = true;
	private $ships = array();
	private $capacity = 0;
	private $data = array();
	private $planetid = 0;
	private $galaxy = 0;
	private $system = 0;
	private $userid = 0;
	private $time = 0;
	private $research = array();
	private $method = '';
	private $message = '';
	private $art_data = array();
	private $additional_msg = array();
	private $statid = 0;

	private $found_credit	= null;
	private $found_metal	= null;
	private $found_silicon	= null;
	private $found_hydrogen = null;
	private $found_fleet	= array();


	private $times_visited = 0;
	private $xUnitsAvailable = array(); // UNIT_A_CORVETTE, UNIT_A_SCREEN, UNIT_A_PALADIN, UNIT_A_FRIGATE, UNIT_A_TORPEDOCARIER);

	/**
	* Starts an expedition.
	*
	* @param integer
	* @param integer
	* @param integer
	* @param array
	*
	* @return void
	*/
	public function __construct(EventHandler $EH, $event) // $userid, $planetid, $time, array &$data, EventHandler $EH)
	{
		Core::getLanguage()->load("info");

		$this->EH				= $EH;
		$this->event			= $event;
		$this->userid			= $this->event["userid"];
		$this->planetid			= $this->event["planetid"];
		$this->time				= $this->event["time"];
		$this->data				= & $this->event["data"];
		$this->ships			= & $this->data["ships"];
		$this->xUnitsAvailable	= array_flip(array(UNIT_A_CORVETTE, UNIT_A_SCREEN, UNIT_A_PALADIN, UNIT_A_FRIGATE, UNIT_A_TORPEDOCARIER));

		$days_const = 7*1;
		$time_const = $this->time - $days_const * 60*60*24;
		// Galaxy
		$this->galaxy = $this->data["galaxy"] = clampVal($this->data["galaxy"], 1, NUM_GALAXYS);
		// System
		$this->system = $this->data["system"] = clampVal($this->data["system"], 1, NUM_SYSTEMS);

		// Total times visited by anybody
		$this->times_visited = sqlSelectField(
			"expedition_stats",
			"count(*)",
			"",
			"`galaxy` = ".sqlVal($this->galaxy)." AND `system`=".sqlVal($this->system)." AND `time` > ".sqlVal($time_const)
		);

		// Getting research levels of curr user
		$tech_list = array(UNIT_EXPO_TECH, UNIT_GUN_TECH, UNIT_SHIELD_TECH, UNIT_SHELL_TECH, UNIT_SPYWARE,
			UNIT_BALLISTICS_TECH, UNIT_MASKING_TECH, UNIT_LASER_TECH, UNIT_ION_TECH, UNIT_PLASMA_TECH);
		foreach($tech_list as $tech_id)
		{
			$this->research[$tech_id] = 0;
		}

		$result = sqlSelect("research2user", array("buildingid", "level"), "",
			"userid = ".sqlVal($this->userid)." AND buildingid IN (" . sqlArray($tech_list) . ")");
		while ($row = sqlFetch($result))
		{
			$this->research[$row["buildingid"]] = $row["level"];
		}
		sqlEnd($result);

		// Time of expedition
		$exp_tech = $this->research[UNIT_EXPO_TECH];
		$exp_hours = $this->data["expedition_hours"];
		$spy_tech = $this->research[UNIT_SPYWARE];
		$spy_units = isset( $this->ships[UNIT_ESPIONAGE_SENSOR] ) ? floor($this->ships[UNIT_ESPIONAGE_SENSOR]["quantity"]) : 0;

		// Metaphysical const, representing power of this expedition
		$exp_power = $exp_tech + $exp_hours * 2 + $spy_tech/10.0 * pow($spy_units, 0.4);

		$visited_scale = 1;
		if($this->times_visited >= 20)
		{
			$visited_scale = pow($this->times_visited, -0.7);
		}
		else if($this->times_visited >= 10)
		{
			$visited_scale = pow($this->times_visited, -0.5);
		}
		else if($this->times_visited >= 5)
		{
			$visited_scale = pow($this->times_visited, -0.3);
		}
		else if($this->times_visited >= 3)
		{
			$visited_scale = pow($this->times_visited, -0.2);
		}
		$visited_scale *= pow(($exp_hours + 1)/6.0, 1.5);

		$this->data["exp_tech"]			= $this->research[UNIT_EXPO_TECH];
		$this->data["spy_tech"]			= $this->research[UNIT_SPYWARE];
		$this->data["spy_units"]		= $spy_units;
		$this->data["exp_power"]		= round($exp_power, 2);
		$this->data["exp_hours"]		= $exp_hours;
		$this->data["times_visited"] 	= $this->times_visited;
		$this->data["visit_days"]		= $days_const;
		$this->data["visited_scale"] 	= $visited_scale;
		$this->data["types"]			= array();
		$this->data["types_percent"]	= array();

		$this->message = ""
			. "exp_power: ".round($this->data["exp_power"],2)
			. " exp_tech: ".$this->data["exp_tech"]
			. " spy_tech: ".$this->data["spy_tech"]
			. " spy_units: ".$this->data["spy_units"]
			. " hours: ".$this->data["expedition_hours"]
			. " visited: ".$this->data["times_visited"]
			. "\n";

		$this->types = & $this->data["types"];

		$this->types["resourceDiscovery"]		= ceil( 100 * pow(1.22, $exp_power) );
		$this->types["asteroidDiscovery"]		= ceil( (70 + ($exp_hours == 0 ? 30 : 0)) * pow(1.22, $exp_power) );
		$this->types["shipsDiscovery"]			= $exp_hours >= 2 ? ceil( 20 * pow(1.25, $exp_power) ) : 0;
		$this->types["battlefieldDiscovery"]	= $exp_hours >= 1 ? ceil( 20 * pow(1.25, $exp_power) ) : 0;
		$this->types["xSkirmish"]				= min( ceil( ($exp_hours >= 3 ? 10 : 0.1) * pow(1.26, $exp_power) ), $this->types["resourceDiscovery"] );
		$this->types["visitePlanetDiscovery"]	= min( ceil( ($exp_hours >= 7 ? 10 : 0) * pow(1.26, $exp_power) ), $this->types["resourceDiscovery"] );
		$this->types["nothing"]					= ceil(	40 * pow(1.25, $this->times_visited + $exp_power / 4) );
		$this->types["delayReturn"]				= $exp_hours >= 1 ? ceil( 30 * pow(1.25, $this->times_visited + $exp_power / 4) ) : 0;
		$this->types["fastReturn"]				= ceil(	60 * pow(1.25, $this->times_visited + $exp_power / 4) );
		$this->types["expeditionLost"]			= ceil(	10 ); // will be fixed to 1% later
		$this->types["artefactDiscovery"]		= min( $exp_hours >= 4 ? ceil( 3 * pow(1.28, $exp_power) ) : 0, $this->types["xSkirmish"] / 2 );
		$this->types["creditDiscovery"]			= min( $exp_hours >= 4 ? ceil( 4 * pow(1.28, $exp_power) ) : 0, $this->types["xSkirmish"] / 2 );

        if($exp_power > 20){
            $this->types["nothing"] *= 20 / $exp_power;
            $this->types["delayReturn"] *= 20 / $exp_power;
            $this->types["fastReturn"] *= 20 / $exp_power;
        }

		foreach($this->ships as $ship)
		{
			if($ship["quantity"] > 10000)
			{
				$this->types["xSkirmish"] *= 0.5;
				$this->types["shipsDiscovery"] *= 0.5;
				$this->types["battlefieldDiscovery"] *= 0.5;
			}
		}
		if(isset($this->ships[UNIT_DEATH_STAR]))
		{
			if($this->ships[UNIT_DEATH_STAR]["quantity"] > 100){
				$this->types["xSkirmish"] *= 0.1;
				$this->types["shipsDiscovery"] *= 0.5;
				$this->types["battlefieldDiscovery"] *= 0.5;
			}else if($this->ships[UNIT_DEATH_STAR]["quantity"] > 50){
				$this->types["xSkirmish"] *= 0.5;
				$this->types["shipsDiscovery"] *= 0.5;
				$this->types["battlefieldDiscovery"] *= 0.5;
			}
		}

		$sum_prob = 0;
		$types_usage = array();
		$types_fixed = array();
		foreach ( $this->types as $key => $value )
		{
			if($value <= 0)
			{
				continue;
			}
			$this->types[$key] = $value = ceil($value * (1 + randSignFloat()*0.05));
			if($key != "expeditionLost")
			{
				switch(mt_rand(0, 100))
				{
				case 0:
					$types_fixed[$key] = $value;
					$this->types[$key] = $value = $value*10;
					break;

				case 1:
					$types_fixed[$key] = $value;
					$this->types[$key] = $value = 0;
					break;
				}
			}
			$types_usage[$key] = 0;
			$sum_prob += $value;
		}
		foreach(array(
			"nothing" => 0.01,
			"fastReturn" => 0.01,
			"delayReturn" => 0.01,
			"expeditionLost" => 0.001) as $name => $scale)
		{
			$new_prob = ceil($sum_prob * $scale);
			if($this->types[$name] < $new_prob)
			{
				$types_fixed[$name] = $this->types[$name];
				$sum_prob += $new_prob - $this->types[$name];
				$this->types[$name] = $new_prob;
			}
		}
        if(!EXPED_LOST_ENABLED){
            foreach(array("expeditionLost", "nothing", "fastReturn", "delayReturn") as $name){
                $sum_prob -= $this->types[$name];
                unset($this->types[$name]);
                unset($types_fixed[$name]);
            }
        }

        $sum_prob = max(1, $sum_prob);
		$this->types_percent = & $this->data["types_percent"];
		foreach($this->types as $key => $value)
		{
			$percent 					= $value * 100 / $sum_prob;
			$this->types_percent[$key] 	= round($percent, 3);
			$percent 					= round($percent, 1);
			if(!isset($types_fixed[$key]))
			{
				$this->message .= " $key: {$percent}% ($value)\n";
			}
			else
			{
				$this->message .= " $key: {$percent}% ($value <- {$types_fixed[$key]})\n";
			}
		}

		$this->message .= "Fleet: ".strip_tags(getUnitListStr($this->ships))."\n";

		$method = 'resourceDiscovery';
//		Used for debuging, DO NOT DELETE
		$stat_size = 1;
		for($i = 0; $i < $stat_size; $i++)
		{
			$rand = randRoundRange(1, $sum_prob);
			foreach($this->types as $key => $value)
			{
				if($rand <= $value && $value > 0)
				{
					$types_usage[$key]++;
					$method = $key;
					break;
				}
				$rand -= $value;
			}
		}
		// it can be used to check type percents
		if( $stat_size > 1 )
		{
			// $method = 'resourceDiscovery';
			foreach($types_usage as $key => $value)
			{
					Logger::addMessage("$key: $value (".round($stat_size * $this->types[$key] / $sum_prob).")");
			}
		}
		// if($this->userid == 1) $method = 'xSkirmish';
		//	$method = 'artefactDiscovery';
		$this->data["exp_type"] = $method;
		$this->data["exp_percent"] = $this->types_percent[$method];

		// $this->message .= "system: {$this->system}, max: ".NUM_SYSTEMS.", app params:\n".var_export([], true)."\n";

		$this->statid = sqlInsert(
			"expedition_stats",
			array(
				"userid"	=> $this->userid,
				"time"		=> $this->time,
				"galaxy"	=> $this->galaxy,
				"system"	=> $this->system,
				"type"		=> $method,
				"percent"	=> $this->types_percent[$method],
				"message"	=> $this->message,
				"completed"	=> 0,
				"event_id"	=> $this->event["eventid"],
			)
		);

		if ( isset($this->ships[UNIT_ESPIONAGE_SENSOR]) )
		{
			new AutoMsg( MSG_EXPEDITION_SENSOR, $this->userid, $this->time, $this->data );
		}

		$log_data = array();
		$log_data = $this->$method();
		$art_type = 0;
		if( $method == 'artefactDiscovery' )
		{
			if( !isset($log_data['a_type_id']) || empty($log_data['a_type_id']) )
			{
				$method = $this->method;
			}
			else
			{
				$art_type = $log_data['a_type_id'];
			}
			$temp = $this->data;
			$temp['art_id_desc'] = $this->art_data['art_id_desc'];
			new AutoMsg(MSG_EXPEDITION_REPORT, $this->userid, $this->time, $temp );
		}
		else
		{
			new AutoMsg(MSG_EXPEDITION_REPORT, $this->userid, $this->time, $this->data);
		}

		if( $this->sendBack )
		{
			$this->data["oldmode"] = EVENT_EXPEDITION;
			$this->EH->sendBack(
				$this->time + $this->data["time"],
				NULL,
				$this->userid,
				$this->planetid,
				$this->data,
				$this->event["eventid"]
			);
		}
		if( !empty($this->additional_msg) )
		{
			new AutoMsg(
				$this->additional_msg['mode'],
				$this->additional_msg['userid'],
				$this->additional_msg['time'],
				$this->additional_msg['data']
			);
		}
		$data_to_update = array(
			'message'		=> $this->message,
			'completed'		=> 1,
			'type'			=> !empty($this->data['exp_type']) ? $this->data['exp_type'] : $method,
			'artefact_type' => $art_type
		);
		if( !empty($this->found_credit) )
		{
			$data_to_update['found_credit'] = $this->found_credit;
		}
		if( !empty($this->found_metal) )
		{
			$data_to_update['found_metal'] = $this->found_metal;
		}
		if( !empty($this->found_silicon) )
		{
			$data_to_update['found_silicon'] = $this->found_silicon;
		}
		if( !empty($this->found_hydrogen) )
		{
			$data_to_update['found_hydrogen'] = $this->found_hydrogen;
		}
		// Update By Pk
		sqlUpdate(
			'expedition_stats',
			$data_to_update,
			'statid = '.sqlVal($this->statid)
		);
		if( !empty($this->found_fleet) )
		{
			$this->logExpeditionFleet();
		}
	}

	protected function logExpeditionFleet()
	{
		foreach( $this->found_fleet as $unit_id => $data )
		{
			sqlInsert(
				'expedition_found_units',
				array(
					'unit_id'		=> $unit_id,
					'expedition_id' => $this->statid,
					'quantity'		=> $data['quantity'],
				)
			);
		}
	}

	private function creditDiscovery()
	{
		$buy_credit = sqlSelectField(
			"payments",
			"sum(pay_credit)",
			"",
			"pay_user_id=".sqlVal($this->userid)." AND pay_date > ".sqlVal(date("Y-m-d", time()-60*60*24*3))." AND pay_status=1"
		);

		$this->data["credit"] = mt_rand(10 + min(100, $buy_credit / 10) + $this->data["exp_power"]/2, 29 + min(300, $buy_credit) + $this->data["exp_power"]*2);
		$this->data["credit"] = max(19, ceil( $this->data["credit"] * $this->data["visited_scale"] )) * 0.7;
		$this->data["credit"] -= $this->data["credit"] % 10;
		$this->data["credit"] += mt_rand(5, 9); // allow string form support

		NS::updateUserRes(array(
			"type"			=> RES_UPDATE_EXPEDITION_CREDITS,
			"reload_planet" => false,
			"userid"		=> $this->userid,
			"credit"		=> $this->data["credit"],
		));
		$this->found_credit = $this->data['credit'];
		$this->message .= "Credit: {$this->data['credit']}\n";
	}

	private function asteroidDiscovery()
	{
		return $this->resourceDiscovery(true);
	}

	private function getResourcesFound()
	{
		$exp_res_scale = max($this->data["expedition_hours"] > 0 ? 0.5 : 0.25, (1 + pow($this->data["expedition_hours"], 1.1)) * $this->data["exp_power"] / 40 * $this->data["visited_scale"]) * (1 + randSignFloat()*0.05);
		$res_k = randRoundRange(500000 * $exp_res_scale, 1000000 * $exp_res_scale) * 2;
		if(!mt_rand(0, 50))
		{
			$res_k *= 100;
		}
		$res_k = min(1000000 * $exp_res_scale * 10, $res_k);

		return array(
			"exp_res_scale" => $exp_res_scale,
			"debrismetal" => ceil($res_k),
			"debrissilicon" => ceil($res_k / 2 * (1 + randSignFloat()*0.1)),
			"debrishydrogen" => ceil($res_k / 3 * (1 + randSignFloat()*0.1)),
		);
	}

	private function resourceDiscovery($is_asteroid = false)
	{
		$res_found = $this->getResourcesFound();

		$data = &$this->data;
		$capacity = (int)$data["capacity"];
		$exp_res_scale = $res_found["exp_res_scale"];
		$data["debrismetal"] = $res_found["debrismetal"];
		$data["debrissilicon"] = $res_found["debrissilicon"];
		$data["debrishydrogen"] = $res_found["debrishydrogen"];

		/*
		$data = &$this->data;
		$capacity = (int)$data["capacity"];

		$exp_res_scale = max($this->data["expedition_hours"] > 0 ? 0.5 : 0.25, (1 + pow($this->data["expedition_hours"], 1.1)) * $this->data["exp_power"] / 40 * $this->data["visited_scale"]) * (1 + randSignFloat()*0.05);
		$res_k = randRoundRange(100000 * $exp_res_scale, 1000000 * $exp_res_scale);
		if(!mt_rand(0, 50))
		{
			$res_k *= 100;
		}
		$res_k = min(1000000 * $exp_res_scale * 3, $res_k);

		$data["debrismetal"]	= ceil($res_k);
		$data["debrissilicon"]	= ceil($res_k / 2 * (1 + randSignFloat()*0.1));
		$data["debrishydrogen"] = ceil($res_k / 3 * (1 + randSignFloat()*0.1));
		*/

		if($is_asteroid)
		{
			$data["debrishydrogen"] = 0;

			$recycle_capacity = 0;
			foreach($GLOBALS["RECYCLER_UNITS"] as $recycler_unit_id)
			{
				if(isset($data["ships"][$recycler_unit_id]) && $data["ships"][$recycler_unit_id]["quantity"] > 0)
				{
					$_row = sqlSelectRow("ship_datasheet", "capicity", "", "unitid = ".sqlVal($recycler_unit_id)); // It is __capacity__, not capicity
					$recycle_capacity += $_row["capicity"] * $data["ships"][$recycler_unit_id]["quantity"];
				}
			}
			$capacity = clampVal($recycle_capacity, 0, $capacity); // max(0, min($data["capacity"], $recycle_capacity));
			/*
			if(isset($this->ships[UNIT_RECYCLER]))
			{
				$recycle_capacity	= sqlSelectField("ship_datasheet", "capicity", "", "unitid = ".UNIT_RECYCLER);
				$recycle_capacity	*= $this->ships[UNIT_RECYCLER]["quantity"];
				$capacity			= max(0, min($capacity, $recycle_capacity));
			}
			else
			{
				$capacity = 0;
			}
			*/
		}
		else
		{
		}
		$data["capacity"] = $capacity;
		recycleDebris($data);
		$this->message .= "ResScale: ".round($exp_res_scale, 2)."\n";
		$this->message .= "Debris: ".$this->data["debrismetal"]." ".$this->data["debrissilicon"]." ".$this->data["debrishydrogen"]."\n";

		$this->found_metal		= $this->data["recycledmetal"];
		$this->found_silicon	= $this->data["recycledsilicon"];
		$this->found_hydrogen	= $this->data["recycledhydrogen"];

		$this->message .= "Recycled: ".$this->data["recycledmetal"]." ".$this->data["recycledsilicon"]." ".$this->data["recycledhydrogen"]."\n";
	}

	private function autoDiscovery()
	{
		if(mt_rand(0, 1))
		{
			$this->data["exp_type"] = "resourceDiscovery";
			return $this->resourceDiscovery();
		}
		$this->data["exp_type"] = "asteroidDiscovery";
		return $this->asteroidDiscovery();
	}

	private function artefactDiscovery()
	{
		$total_probability = 0;
		$a_result = sqlSelect('artefact_probobility', '*', '', 'probobility != 0');
		$artefacts = array();
		while($row = sqlFetch($a_result))
		{
			$new_key = $row['type'];
			$artefacts[$new_key] = $row;
			$artefacts[$new_key]['a_type_id'] = $row['type'];
			$total_probability += $row["probobility"];
		}
		sqlEnd($a_result);

		if( count($artefacts) == 0 )
		{
			return $this->autoDiscovery();
		}

		// Generate optimal artefact
		$probability = randRoundRange(1, $total_probability);
		foreach ($artefacts as $key => $value)
		{
			if($probability <= $value["probobility"])
			{
				break;
			}
			$probability -= $value["probobility"];
		}
		$artefact = $value;

		$level_types = array(
//			array( 'probobility' => ALIEN_PROB,		'type' => "alien_tech" ),
			array( 'probobility' => USER_READY_PROB,'type' => "user_level" ), // User tech or building
			// array( 'probobility' => HIGH_LEVEL_PROB,'type' => "high_level" ), // Level 21-40
			// array( 'probobility' => MID_LEVEL_PROB,	'type' => "mid_level" ), // Level 11-20
			array( 'probobility' => LOW_LEVEL_PROB,	'type' => "low_level" ), // Level 1-10
		);
		$total_probability = 0;
		foreach($level_types as $key => $value)
		{
			$total_probability += $value["probobility"];
		}
		$probability = randRoundRange(1, $total_probability);
		foreach($level_types as $akey => $value)
		{
			if($probability <= $value["probobility"])
			{
				break;
			}
			$probability -= $value["probobility"];
		}
		$artefact['level'] 			 = 0;
		$artefact['construction_id'] = 0;
		$artefact['con_name'] 		 = '';

		if ( $artefact['a_type_id'] == ARTEFACT_PACKED_BUILDING || $artefact['a_type_id'] == ARTEFACT_PACKED_RESEARCH )
		{
			switch ( $level_types[$akey]['type'] )
			{
				case "alien_tech":
					if( $artefact['a_type_id'] == ARTEFACT_PACKED_BUILDING )
					{
						$artefact = $this->searchPlayerLevel( $artefact );
					}
					else
					{
						$artefact = $this->alientResearchArtefact($artefact);
					}
					break;

				case "user_level":
					$artefact = $this->searchPlayerLevel( $artefact );
					break;

				default:
					$artefact = $this->makePackedArtefact($artefact, $level_types[$akey]['type']);
					break;
			}
		}
		// Create Artefact
		$artefact['flying']	= true;
		$aid				= Artefact::appear($artefact['a_type_id'], $this->userid, null, $artefact);

		$artefact["level"] = min($artefact["level"], $artefact['a_type_id'] == ARTEFACT_PACKED_BUILDING ? MAX_BUILDING_LEVEL*0.9 : MAX_RESEARCH_LEVEL*0.9);
		if(isset($GLOBALS["MAX_UNIT_LEVELS"][$artefact["construction_id"]]))
		{
			$artefact["level"] = min($artefact["level"], $GLOBALS["MAX_UNIT_LEVELS"][$artefact["construction_id"]]*0.9);
		}
		$artefact["level"] = max(1, floor($artefact["level"]));

		// Add artefact to fleet
		$fleet_artefact = array();
		$fleet_artefact['mode'] = UNIT_TYPE_ARTEFACT;
		$fleet_artefact['id'] = $artefact['a_type_id'];
		$fleet_artefact['quantity'] = 1;
		$fleet_artefact['name'] = sqlSelectField('construction', 'name', '', 'buildingid = '.sqlVal($artefact['a_type_id']));
		$fleet_artefact['art_ids'][$aid] = array(
			'artid' 	=> $aid,
			'level'		=> $artefact["level"],
			'con_id'	=> $artefact["construction_id"],
			'con_name'	=> $artefact["con_name"],
		);
		$this->data['ships'][$artefact['a_type_id']] = $this->ships[$artefact['a_type_id']] = $fleet_artefact;
		$data = getArtefactNameAndPosition($aid);
		$this->art_data['art_name'] = $data['name'];
		$this->art_data['art_id_desc'] = $aid;
		$this->message .= strip_tags( $data['name'] );
		return $artefact;
	}

	/**
	 *
	 * Generates artefact
	 * @param array $artefact Artefact info
	 * @param int $type Level type
	 */
	private function makePackedArtefact($artefact, $type)
	{
		if ( $artefact['a_type_id'] == ARTEFACT_PACKED_BUILDING )
		{
			$where = "mode in (".sqlArray(UNIT_TYPE_CONSTRUCTION, UNIT_TYPE_MOON_CONSTRUCTION).")"
						. " AND buildingid not in (".sqlArray(UNIT_EXCHANGE, UNIT_NANO_FACTORY).")";
		}
		else
		{
			$where = 'mode = '.UNIT_TYPE_RESEARCH." AND buildingid != ".UNIT_ALIEN_TECH;
		}
		$result = sqlSelect('construction', 'buildingid, name', '', $where);
		while($row = sqlFetch($result))
		{
			$all_buildings[] = $row;
		}
		sqlEnd($result);
		$rand_key = array_rand($all_buildings);
		$artefact['construction_id'] = $all_buildings[ $rand_key ]['buildingid'];
		$artefact['con_name'] = $all_buildings[ $rand_key ]['name'];
		switch ($type)
		{
			default:
			case "low_level": $artefact['level'] = mt_rand(1, 10);break;
			case "mid_level": $artefact['level'] = mt_rand(11, 20);break;
			case "high_level": $artefact['level'] = mt_rand(21, 40);break;
		}
		return $artefact;
	}

	/**
	 *
	 * Creates Artefact that user can use now
	 * @param array $artefact
	 */
	private function searchPlayerLevel($artefact)
	{
		$user_id = $this->userid;
		if ( $artefact['a_type_id'] == ARTEFACT_PACKED_BUILDING )
		{
			$result = sqlSelect(
				'building2planet AS b',
				'b.buildingid, b.level, b.added, c.name',
				'LEFT JOIN '.PREFIX.'planet AS p ON p.planetid = b.planetid '
					. ' LEFT JOIN '.PREFIX.'construction AS c ON c.buildingid = b.buildingid',
				'p.userid ='.sqlVal($user_id).' AND b.buildingid != '.UNIT_EXCHANGE
			);
		}
		else
		{
			$result = sqlSelect('research2user', 'buildingid, level, added', '', 'userid ='.sqlVal($user_id).' AND buildingid != '.UNIT_ALIEN_TECH);
		}
		while($row = sqlFetch($result))
		{
			$all_user_buildings[] = $row;
		}
		sqlEnd($result);

		$row = $all_user_buildings[ array_rand($all_user_buildings) ];
		$artefact['level'] = $row['level'] - $row['added'] + 1;
		$artefact['construction_id'] = $row['buildingid'];
		$artefact['con_name'] = $row['name'];

		return $artefact;
	}

	/**
	 *
	 * Creates Alien Research tech Artefact
	 * @param array $artefact
	 * @param int $user_count
	 */
	private function alientResearchArtefact($artefact)
	{
		$user_count = sqlSelectField(
			'user',
			'count(*)',
			'',
			'`points` > 0 and `last` > ' . sqlVal(time()-7*24*60*60)
		);
		$max_alient_tech_count = $user_count * MAX_ALIEN_TECH_FACTOR;
		$alien_tech_count = sqlSelectField('research2user', 'count(*)', '', 'buildingid = '.UNIT_ALIEN_TECH);
		if ( $alien_tech_count < $max_alient_tech_count )
		{
			$artefact['level'] = 1;
			$artefact['construction_id'] = UNIT_ALIEN_TECH;
			$artefact['con_name'] = 'UNIT_ALIEN_TECH';
			return $artefact;
		}
		return $this->searchPlayerLevel($artefact);
	}

	private function shipsDiscovery($battlefield = false)
	{
		$res_found = $this->getResourcesFound();
		$res_found["debrismetal"] *= 3;
		$res_found["debrissilicon"] *= 3;
		$res_found["debrishydrogen"] *= 3;

		$alliense_ships = mt_rand(0, 9) == 0;
		if ($alliense_ships)
		{
			$res_found["scale"] = 0.1 * mt_rand(1, 3) * $this->data["visited_scale"];
			$this->data["newShips"] = $this->createOponentFleet($this->ships, $this->xUnitsAvailable, $res_found, true);
		}
		else
		{
			$res_found["scale"] = 0.1 * mt_rand(3, 5) * $this->data["visited_scale"];
			$this->data["newShips"] = $this->createOponentFleet($this->ships, $this->ships, $res_found, true);
			if(sizeof($this->data["newShips"]) == 0 && mt_rand(0, 1))
			{
				$alliense_ships = true;
				$res_found["scale"] = 0.1 * mt_rand(1, 3) * $this->data["visited_scale"];
				$this->data["newShips"] = $this->createOponentFleet($this->ships, $this->xUnitsAvailable, $res_found, true);
			}
		}

		if(sizeof($this->data["newShips"]) == 0)
		{
			return $this->autoDiscovery();
		}

		foreach($this->data["newShips"] as $id => $ship)
		{
			if ($battlefield)
			{
				if(mt_rand(0, 9))
				{
					$this->data["newShips"][$id]["damaged"] = $this->data["newShips"][$id]["quantity"];
					$this->data["newShips"][$id]["shell_percent"] = mt_rand(10, 35);
				}
				else
				{
					$this->data["newShips"][$id]["damaged"] = mt_rand(1, $this->data["newShips"][$id]["quantity"]);
					$this->data["newShips"][$id]["shell_percent"] = mt_rand(90, 99);
				}
			}
			if (!isset($this->ships[$id]))
			{
				$this->ships[$id] = $this->data["newShips"][$id];
			}
			else
			{
				if ($battlefield)
				{

					$damage = $this->data["newShips"][$id]["damaged"] * $this->data["newShips"][$id]["shell_percent"] +
						$this->ships[$id]["damaged"] * $this->ships[$id]["shell_percent"];
					$this->ships[$id]["damaged"] = $this->data["newShips"][$id]["damaged"] + $this->ships[$id]["damaged"];
					$this->ships[$id]["shell_percent"] = $damage / $this->ships[$id]["damaged"];
				}
				$this->ships[$id]["quantity"] += $this->data["newShips"][$id]["quantity"];
			}
		}
		$this->found_fleet = $this->data["newShips"];
		$this->message .= "Found: ".strip_tags(getUnitListStr($this->data["newShips"]))."\n";
	}

	private function battlefieldDiscovery()
	{
		return $this->shipsDiscovery(true);
	}

	private function nothing()
	{
		return;
	}

	private function delayReturn()
	{
		$this->data["time_k"] = round(0.1 + randFloat() * 0.2, 3);
		$this->data["delay_time"] = round($this->data["time"] * $this->data["time_k"]);
		$this->data["time"] += $this->data["delay_time"];
		$this->message .= "Percent: ".round($this->data["time_k"] * 100, 1)."%, Time: ".$this->data["delay_time"]."\n";
	}

	private function fastReturn()
	{
		$this->data["time_k"] = round(0.1 + randFloat() * 0.5, 3);
		$this->data["fast_time"] = round($this->data["time"] * $this->data["time_k"]);
		$this->data["time"] -= $this->data["fast_time"];
		$this->message .= "Percent: ".round($this->data["time_k"] * 100, 1)."%, Time: ".$this->data["fast_time"]."\n";
	}

	private function pirates()
	{
	}

	private function blackHole()
	{
	}

	private function expeditionLost()
	{
		$this->sendBack = false;
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

	private function xSkirmish()
	{
		$data = &$this->data;

		$research_ids = array(
			UNIT_GUN_TECH, UNIT_SHIELD_TECH, UNIT_SHELL_TECH,
			UNIT_BALLISTICS_TECH, UNIT_MASKING_TECH,
			UNIT_LASER_TECH, UNIT_ION_TECH, UNIT_PLASMA_TECH
		);
		$research = array();
		$result = sqlSelect("research2user", array("buildingid", "level"), "", "userid = ".sqlVal($this->userid)." AND buildingid in (".sqlArray($research_ids).")");
		while($row = sqlFetch($result))
		{
			$research[$row["buildingid"]] = $row["level"];
		}
		sqlEnd($result);

		foreach($research_ids as $techid)
		{
			if(empty($research[$techid]))
			{
				$research[$techid] = 0;
			}
		}
		self::shuffleKeyValues( $research, array(UNIT_GUN_TECH, UNIT_SHIELD_TECH, UNIT_SHELL_TECH) );
		self::shuffleKeyValues( $research, array(UNIT_BALLISTICS_TECH, UNIT_MASKING_TECH) );
		self::shuffleKeyValues( $research, array(UNIT_LASER_TECH, UNIT_ION_TECH, UNIT_PLASMA_TECH) );

		$attack_types = array(
			array("probabilit" => 55, "scale" => 0.5, "aliens_attack" => mt_rand(0, 1), "tech_scale" => mt_rand(50, 70) / 100.0),
			array("probabilit" => 35, "scale" => 0.7, "aliens_attack" => mt_rand(0, 1), "tech_scale" => mt_rand(80, 110) / 100.0),
			array("probabilit" => 9, "scale" => 1.0, "aliens_attack" => mt_rand(0, 1), "tech_scale" => mt_rand(100, 110) / 100.0),
			array("probabilit" => 1, "scale" => 100.0, "aliens_attack" => true, "tech_scale" => mt_rand(100, 110) / 100.0),
			);

		$prob_sum = 0;
		foreach($attack_types as $key => $value)
		{
			$prob_sum += $value["probabilit"];
		}

		$prob = mt_rand(1, $prob_sum);
		foreach($attack_types as $akey => $value)
		{
			if($prob <= $value["probabilit"])
			{
				break;
			}
			$prob -= $value["probabilit"];
		}

		$aliens_role = $attack_types[$akey]["aliens_attack"] ? 1 : 0;
		$player_role = 1 - $aliens_role;

		require_once(APP_ROOT_DIR."game/Assault.class.php");
		$Assault = new Assault($this->EH, $this->event["eventid"], NULL);

		$res_found = $this->getResourcesFound();
		$res_found["scale"] = $attack_types[$akey]["scale"];
		$res_found["debrismetal"] *= 5;
		$res_found["debrissilicon"] *= 5;
		$res_found["debrishydrogen"] *= 5;

		$dataXeno["ships"] = $this->createOponentFleet($this->ships, $this->xUnitsAvailable, $res_found);

		if(sizeof($dataXeno["ships"]) == 0)
		{
			return $this->autoDiscovery();
		}

		$this->message .= "Role: ".$player_role.", Prob: ".$attack_types[$akey]["probabilit"].", Scale: ".$attack_types[$akey]["scale"]."\n";

		$tech_scale = round($attack_types[$akey]["tech_scale"], 2);
		$this->message .= "Alien tech scale: $tech_scale\n";
		$this->message .= "Alien: ".strip_tags(getUnitListStr($dataXeno["ships"]))."\n";
		foreach($research_ids as $techid)
		{
			$dataXeno["add_tech_".$techid] = max(1, round($research[$techid] * $tech_scale + randSignFloat()*1));
            if(!EXPED_LOST_ENABLED){
                $dataXeno["add_tech_".$techid] *= 0.7;
            }
		}
		$Assault->addParticipant($aliens_role, NULL, NULL, $this->time, $dataXeno, null);

		if(!mt_rand(0, 4))
		{
			$res_found["scale"] /= 2;
			$dataXeno["ships"] = $this->createOponentFleet($this->ships, $this->xUnitsAvailable, $res_found);
			if(sizeof($dataXeno["ships"]) > 1)
			{
				$tech_scale = round(mt_rand(40, 120) / 100.0, 2);
				$this->message .= "Alien #2 tech scale: $tech_scale\n";
				$this->message .= "Alien #2: ".strip_tags(getUnitListStr($dataXeno["ships"]))."\n";
				foreach($research_ids as $techid)
				{
					$dataXeno["add_tech_".$techid] = max(1, round($research[$techid] * $tech_scale + randSignFloat()*1));
                    if(!EXPED_LOST_ENABLED){
                        $dataXeno["add_tech_".$techid] *= 0.7;
                    }
				}
				$Assault->addParticipant($aliens_role, NULL, NULL, $this->time, $dataXeno, null);

				if(!mt_rand(0, 4))
				{
					$res_found["scale"] /= 2;
					$dataXeno["ships"] = $this->createOponentFleet($this->ships, $this->xUnitsAvailable, $res_found);
					if(sizeof($dataXeno["ships"]) > 1)
					{
						$tech_scale = round(mt_rand(40, 120) / 100.0, 2);
						$this->message .= "Alien #3 tech scale: $tech_scale\n";
						$this->message .= "Alien #3: ".strip_tags(getUnitListStr($dataXeno["ships"]))."\n";
						foreach($research_ids as $techid)
						{
							$dataXeno["add_tech_".$techid] = max(1, round($research[$techid] * $tech_scale + randSignFloat()*2));
                            if(!EXPED_LOST_ENABLED){
                                $dataXeno["add_tech_".$techid] *= 0.7;
                            }
						}
						$Assault->addParticipant($aliens_role, NULL, NULL, $this->time, $dataXeno, null);
					}
				}
			}
		}

		$this->data["xSkirmish"]		= true;
		$this->data["expedition_id"]	= $this->statid;
		$this->data['oldmode']			= EVENT_EXPEDITION;
		$Assault->addParticipant($player_role, $this->userid, $this->planetid, $this->time, $this->data, $this->event["eventid"]);

		$Assault->startAssault( $this->galaxy );

		$assault_id = $Assault->getAssaultId();

		$assault_data = sqlSelectRow('assault', '*', '', 'assaultid = ' . sqlVal($assault_id));

		if( ($assault_data['result'] != 2 && $player_role == 1) || ($assault_data['result'] != 1 && $player_role == 0) )
		{
			// If player didn't lose as attacker OR player didn't lose as deffender
			$this->createNeutralPlanet($assault_data);
		}

		$this->sendBack = false;
		return $this;
	}

	private function visitePlanetDiscovery()
	{
		$planets = array();
		$system1 = clampVal($this->system - 3, 1, NUM_SYSTEMS);
		$system2 = clampVal($this->system + 3, 1, NUM_SYSTEMS);
		$result = sqlSelect(
			'galaxy',
			array('galaxy', 'system', 'position', 'metal', 'silicon', 'planetid'),
			'',
			'galaxy = ' . sqlVal($this->galaxy) .
				' AND system BETWEEN ' . sqlVal($system1) . ' AND ' . sqlVal($system2) .
				' AND position >= ' . EXPED_START_CREATE_PLANET .
				' AND metal > 1000000 AND silicon > 1000000 ' .
				' AND destroyed = 0 ' .
			'ORDER BY silicon DESC ' .
			'LIMIT 10'
		);
		while ( $row = sqlFetch($result) )
		{
			$planets[$row['planetid']] = $row;
		}
		sqlEnd($result);
		if( empty($planets) )
		{
			$this->data["exp_type"] = "xSkirmish";
			return $this->xSkirmish(); // autoDiscovery();
		}
		$this->data['planet_found'] = $planets[array_rand($planets)];

		$planet = $this->data['planet_found'];
		$this->message .= "Found planet: [{$planet['galaxy']}:{$planet['system']}:{$planet['position']}]\n";
		$this->message .= "Debris: {$planet['metal']} {$planet['silicon']}\n";
	}

	/**
	 *
	 * Creates neutral planet and fills it's orbit with scraps
	 * @param array $data Row from assault table
	 */
	protected function createNeutralPlanet($data)
	{
		// Get all scraps
		$metal_scraps	= $data['debris_metal'];
		$silicon_scraps	= $data['debris_silicon'];
		// Create neutral planet
		// Plant scraps on it's orbit
		$position = $this->getPositionForNewPlanet();
		$planet_data = array(
			'percent'		=> '100',
			'galaxy'		=> $this->galaxy,
			'system'		=> $this->system,
			'position'		=> $position,
			'name'			=> Core::getLang()->getItem('NEUTRAL_PLANET_NAME'),
			'metal_scraps'	=> $metal_scraps,
			'silicon_scraps'=> $silicon_scraps,
		);
		$EPC = new ExpedPlanetCreator(null, $planet_data);
		$new_planet_id = $EPC->getPlanetId();
		$new_planet_name = $EPC->getPlanetName();
		$event_id = $this->EH->addEvent(
			EVENT_TEMP_PLANET_DISAPEAR,
			time() + mt_rand(EXPED_PLANET_LIFETIME_MIN, EXPED_PLANET_LIFETIME_MAX),
			$new_planet_id,
			$this->userid,
			$this->planetid,
			array(
				'planet_id'	=> $new_planet_id,
				'user_id'	=> $this->userid,
				'galaxy'	=> $this->galaxy,
				'system'	=> $this->system,
				'position'	=> $position,
			)
		);
		// Send MSG to user about this planet
		$this->additional_msg = array(
			'mode'		=> MSG_EXPEDITION_NEW_PLANET,
			'userid'	=> $this->userid,
			'time'		=> $this->time,
			'data'		=> array(
				'planet_id' => $new_planet_id,
				'galaxy'	=> $this->galaxy,
				'system'	=> $this->system,
				'position'	=> $position,
				'm_scraps'	=> $metal_scraps,
				's_scraps'	=> $silicon_scraps,
				'name'		=> $new_planet_name,
			)
		);
		$this->found_metal		= $metal_scraps;
		$this->found_silicon	= $silicon_scraps;
		$this->message .= "Planet created: [{$this->galaxy}:{$this->system}:$position].\n" .
			"Debris: {$metal_scraps} {$silicon_scraps}\n" .
			"Planet ID: {$new_planet_id}\n" .
			"Planet event ID: {$event_id}\n";
	}

	/**
	 *
	 * Returns free position in current galaxy and system
	 */
	protected function getPositionForNewPlanet()
	{
		$result = sqlSelect(
			'galaxy',
			'position',
			'',
			' galaxy = ' . sqlVal($this->galaxy) .
				' AND system = ' . sqlVal($this->system) .
				' AND position >= ' . EXPED_START_CREATE_PLANET
		);
		$max = 0;
		$positions = array_flip(range(EXPED_START_CREATE_PLANET, EXPED_END_CREATE_PLANET));
		while ( $row = sqlFetch($result) )
		{
			unset($positions[ $row['position'] ]);
			$max = max( $max, $row['position'] );
		}
		sqlEnd($result);
		if($max > 0 && !$positions)
		{
			return $max + 1;
			/*
			for($i = 1; $i <= 4; $i++)
			{
				$positions[$max + $i] = $max + $i;
			}
			*/
		}
		return array_rand($positions, 1);
	}

	public function createOponentFleet($pShips, $xUnitsAvailable, $scale = 1, $find_mode = false)
	{
		$params = array(
			"find_mode" => $find_mode,
			);
		if(is_array($scale))
		{
			$metal = isset($scale["metal"]) ? $scale["metal"] : (isset($scale["debrismetal"]) ? $scale["debrismetal"] : null);
			$silicon = isset($scale["silicon"]) ? $scale["silicon"] : (isset($scale["debrissilicon"]) ? $scale["debrissilicon"] : null);
			// $hydrogen = isset($scale["hydrogen"]) ? $scale["hydrogen"] : (isset($scale["debrishydrogen"]) ? $scale["debrishydrogen"] : null);
			$params["max_derbis"] = min((int)$metal + (int)$silicon, 1000*1000*1000);
			$scale = max(0.001, (float)$scale["scale"]);
		}

		if(!mt_rand(0, 20))
		{
			$params["damaged"] = array(
				"chance" => mt_rand(30, 100),
				"quantity_percent" => array(20, mt_rand(80, 100)),
				"shell_percent" => array(50, mt_rand(80, 100)),
				);
		}

		return AlienAI::generateFleet($pShips, $xUnitsAvailable, $scale, $params);
	}
}
?>