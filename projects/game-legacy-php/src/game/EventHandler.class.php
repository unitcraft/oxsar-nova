<?php
/**
* Handles all events.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

require_once(APP_ROOT_DIR."game/Expedition.class.php");

class EventHandler
{
	/**
	* Event stack holds all user-related events.
	*/
	protected $eventStack = array();

	/**
	* Unique key to prevent race conditions.
	*
	* @var string
	*/
	protected $raceConditionKey = null;

	protected $events_processed = false;

	public $externalMonitor = null; // NewEHMonitorCommand

	protected static $eh = null;

	public static function getEH()
	{
		return self::$eh;
	}

	/**
	* Constructor.
	*/
	public function __construct()
	{
		self::$eh = $this;
		// $this->externalMonitor = $externalMonitor;
		$this->raceConditionKey = $this->genPrevRaceConditionKey();
	}

	public function startEvents()
	{
		$cache_name = "NS:UserLast:".NS::getUser()->get("userid");
		if(!NS::getMCH()->get($cache_name, $user_last) || $user_last < time()-1)
		{
			$user_last = time();
			NS::getMCH()->set($cache_name, $user_last, 60);
			if ( is_null($this->externalMonitor)
					&& (!defined('YII_CONSOLE') || !YII_CONSOLE)
					&& (!defined('YII_CONSOLE_IS_RUNNING') || !YII_CONSOLE_IS_RUNNING) )
			{
				$this->goThroughEvents();
			}
		}
		$this->events_processed = true;
		$this->setEventStack();
	}

	/**
	* Append an event onto the event stack.
	*
	* @param integer	Mode id (see list)
	* @param integer	Time when event will be triggered
	* @param integer	Planet where event has been triggered
	* @param integer	Destination planet (just for fleet events)
	* @param array	Event-related data
	*
	* @return EventHandler
	*/
	public function addEvent($mode, $time, $planetid, $userid, $destination, $data, $protected = null, $start_time = null, $parent_eventid = null, $artid = null)
	{
		$start_time = empty($start_time) ? time() : $start_time;

		$max_galaxy = NUM_GALAXYS;
		$max_system = NUM_SYSTEMS;
		$max_position = MAX_POSITION;
		foreach(array(
			"galaxy" => $max_galaxy,
			"system" => $max_system,
			"position" => $max_position,
			) as $name => $max)
		{
			if(isset($data[$name]))
			{
				$data[$name] = clampVal(floor($data[$name]), 1, $max);
			}
		}
		foreach(array("quantity", "metal", "silicon", "hydrogen",
            "consumption", "credit", "fleet_size", "capacity",
            ) as $name)
        {
            if(isset($data[$name])){
                $data[$name] = floor($data[$name]);
            }
        }
        if(isset($data["ships"])){
            foreach($data["ships"] as &$ship_data)
            {
                foreach(array("quantity", "damaged", // "shell_percent",
                    ) as $name)
                {
                    if(isset($ship_data[$name])){
                        $ship_data[$name] = floor($ship_data[$name]);
                    }
                }
            }
            unset($ship_data);
        }

		$event_key = md5(serialize(array(
			"mode" => $mode,
			"time" => $time,
			"planetid" => $planetid,
			"userid" => $userid,
			"destination" => $destination,
			"data" => $data,
			"protected" => $protected,
			"start_time" => $start_time,
			"parent_eventid" => $parent_eventid,
			"artid" => $artid
		)));

		if(!NS::isFirstRun("EV:addEvent:{$event_key}"))
		{
			return;
		}
		if(NS::getUser() && NS::getUser()->get("umode") && in_array($mode, $GLOBALS["VACATION_BLOCKING_EVENTS"]))
		{
			return;
		}

		$res_log = false;
		switch($mode)
		{
			case EVENT_RETURN:
			case EVENT_DELIVERY_UNITS:
			case EVENT_DELIVERY_RESOURSES:
			case EVENT_HOLDING:
			case EVENT_ARTEFACT_EXPIRE:
			case EVENT_ARTEFACT_DISAPPEAR:
			case EVENT_ARTEFACT_DELAY:
			case EVENT_EXCH_EXPIRE:
			case EVENT_EXCH_BAN:
			case EVENT_TEMP_PLANET_DISAPEAR:
			case EVENT_RUN_SIM_ASSAULT:
				break;

			case EVENT_BUILD_CONSTRUCTION:
			case EVENT_DEMOLISH_CONSTRUCTION:
			case EVENT_RESEARCH:
			case EVENT_BUILD_FLEET:
			case EVENT_BUILD_DEFENSE:
			case EVENT_REPAIR:
			case EVENT_DISASSEMBLE:
				/* if( NS::getUser()->get("umode") )
				{
					return;
				} */
				if(empty($data["paid"]))
				{
					$res_log = NS::updateUserRes(array(
						"block_minus"	=> true,
						"event_mode"	=> $mode,
						"userid"		=> $userid,
						"planetid"		=> $planetid,
						"metal"			=> - $data["metal"],
						"silicon"		=> - $data["silicon"],
						"hydrogen"		=> - $data["hydrogen"] - (isset($data["consumption"]) ? $data["consumption"] : 0),
						"credit"		=> - $data["credit"],
					));

					if(!empty($res_log["minus_blocked"]))
					{
						return;
					}

					if( $data["credit"] > 0 )
					{
						new AutoMsg(
							MSG_CREDIT,
							$userid,
							time(),
							array(
								'credits'	=> $data["credit"],
								'msg' 		=> $mode == EVENT_RESEARCH ? 'MSG_CREDIT_RESEARCHING' : 'MSG_CREDIT_BUILDING_UNIT',
								'content'	=> array('id' => $data["buildingid"], 'name' => $data["buildingname"], 'level' => $data["level"]) )
						);
					}

					if($mode == EVENT_REPAIR)
					{
						logShipsChanging($data["buildingid"], $planetid, - $data["quantity"], 1, "[addEvent] repair");

						// Update By Pk
						sqlQuery("UPDATE ".PREFIX."unit2shipyard SET "
							. subQuantitySetSql(array("damaged" => $data["quantity"]))
							. " WHERE unitid = ".sqlVal($data["buildingid"])." AND planetid = ".sqlVal($planetid));

						Core::getQuery()->delete("unit2shipyard", "quantity = '0'");

						$points = $data["points"] * $data["quantity"];

						// Update By Pk
						sqlQuery("UPDATE ".PREFIX."user SET "
							. " points = points - ".sqlVal($points)
							. ", u_points = u_points - ".sqlVal($points)
							. ", u_count = u_count - ".sqlVal($data["quantity"])
							. " WHERE userid = ".sqlVal($userid));
					}
					elseif($mode == EVENT_DISASSEMBLE)
					{
						logShipsChanging($data["buildingid"], $planetid, - $data["quantity"], 1, "[addEvent] disassemble");

						// Update By Pk
						sqlQuery("UPDATE ".PREFIX."unit2shipyard SET "
							. subQuantitySetSql($data["quantity"])
							. " WHERE unitid = ".sqlVal($data["buildingid"])." AND planetid = ".sqlVal($planetid));

						Core::getQuery()->delete("unit2shipyard", "quantity = '0'");

						$points = $data["points"] * $data["quantity"];
						// $fpoints = intval($data["buildingtype"] == UNIT_TYPE_FLEET) * $data["quantity"];
						// Update By Pk
						sqlQuery("UPDATE ".PREFIX."user SET "
							. " points = points - ".sqlVal($points)
							. ", u_points = u_points - ".sqlVal($points)
							. ", u_count = u_count - ".sqlVal($data["quantity"])
							. " WHERE userid = ".sqlVal($userid));
					}
				}
				break;

			case EVENT_ROCKET_ATTACK:
				logShipsChanging(52, $planetid, - $data["rockets"], 1, "[addEvent] $mode - rockets");
				// Update By Pk
				sqlQuery("UPDATE ".PREFIX."unit2shipyard SET "
					. subQuantitySetSql($data["rockets"])
					. " WHERE unitid = ".UNIT_INTERPLANETARY_ROCKET." AND planetid = ".sqlVal($planetid));
				Core::getQuery()->delete("unit2shipyard", "quantity = '0'");
				break;

			default:
				if($mode == EVENT_ALLIANCE_ATTACK_ADDITIONAL)
				{
					$result = sqlSelect("events", array("eventid", "time"), "",
						"eventid = '".$data["alliance_attack"]["eventid"]."' "
						. " AND processed='".EVENT_PROCESSED_WAIT."'"
						. " AND time > '".(time() + 10)."'"
						);
					$check_alliance_event = sqlFetch($result);
					sqlEnd($result);
					if(!$check_alliance_event)
					{
						Logger::dieMessage("UNKOWN_MISSION");
						return $this;
					}
					$data["alliance_attack"]["time"] = $check_alliance_event["time"];
				}

				if(!isset($data["control_times"]))
				{
					$res_log = NS::updateUserRes(array(
						"block_minus" => true,
						"event_mode" => $mode,
						"userid" => $userid,
						"planetid" => $planetid,
						"metal" => - $data["metal"],
						"silicon" => - $data["silicon"],
						"hydrogen" => - $data["hydrogen"] - (isset($data["consumption"]) ? $data["consumption"] : 0),
						"auto_fix" => array(
							"metal" => $data["metal"],
							"silicon" => $data["silicon"],
							"hydrogen" => $data["hydrogen"],
							),
					));

					if(!empty($res_log["minus_blocked"]))
					{
						return;
					}

					if(!empty($res_log["auto_fixed"]))
					{
						$data["metal"] -= $res_log["auto_fixed"]["metal"];
						$data["silicon"] -= $res_log["auto_fixed"]["silicon"];
						$data["hydrogen"] -= $res_log["auto_fixed"]["hydrogen"];
					}

					foreach($data["ships"] as $ships)
					{
						$ships["quantity"] = floor($ships["quantity"]);
						fixUnitDamagedVars($ships);

						logShipsChanging($ships["id"], $planetid, - $ships["quantity"], 1, "[addEvent] $mode");

						// Update By Pk
						sqlQuery("UPDATE ".PREFIX."unit2shipyard SET "
							. subQuantitySetSql($ships)
							. " WHERE unitid = ".sqlVal($ships["id"])." AND planetid = ".sqlVal($planetid));
					}
					Core::getQuery()->delete("unit2shipyard", "quantity = '0'");
				}

				if($mode == EVENT_ALLIANCE_ATTACK_ADDITIONAL)
				{
					if($time > $data["alliance_attack"]["time"])
					{
						sqlUpdate("events", array("time" => $time),
							"parent_eventid = ".sqlVal($data["alliance_attack"]["eventid"])." AND processed=".EVENT_PROCESSED_WAIT . ' ORDER BY eventid');
						// Update By Pk
						sqlUpdate("attack_formation", array("time" => $time),
							"eventid = ".sqlVal($data["alliance_attack"]["eventid"]));
					}
					else
					{
						$time = $data["alliance_attack"]["time"];
					}
				}
				break;
		}

		if( !isset($data["created_time"]) )
		{
			$data["created_time"] = time();
		}
		// unset($data["back_consumption"]);

		$serialized 	= serialize($data);
		$is_long_event 	= in_array($mode, array(EVENT_BUILD_FLEET, EVENT_BUILD_DEFENSE, EVENT_REPAIR, EVENT_DISASSEMBLE, EVENT_RUN_SIM_ASSAULT));
		$arrayToInsert 	= array(
			"mode" 				=> $mode,
			"start" 			=> $start_time,
			"time" 				=> $time,
			"user" 				=> $userid,
			"destination" 		=> $destination,
			"data" 				=> $serialized,
			"required_quantity" => $is_long_event ? $data["quantity"] : null,
			"org_data" 			=> $is_long_event ? $serialized : null,
			"parent_eventid" 	=> $mode == EVENT_ALLIANCE_ATTACK_ADDITIONAL ? $data["alliance_attack"]["eventid"] : $parent_eventid,
			"artid"				=> $artid,
		);
		// todo: is it real needed?
		if ( !empty($planetid) ) {
			$arrayToInsert['planetid'] = $planetid;
		}
		$eventid = sqlInsert( "events", $arrayToInsert );

		if(!isset($data["control_times"]))
		{
			$this->addEventIDToArtefacts($eventid, $data);

			if( isset($data["used_exp_points"]) )
			{
				// Update By Pk
				sqlQuery("UPDATE ".PREFIX."user SET be_points = be_points - ".sqlVal($data["used_exp_points"])." WHERE userid=".sqlVal($userid));
			}

			if($res_log)
			{
				// Update By Pk
				sqlUpdate("res_log", array("ownerid" => $eventid), "id=".sqlVal($res_log["id"]));
			}
		}
		return $eventid;
	}

	protected function addEventIDToArtefacts( $event_id, $data )
	{
		if( !is_array($data) || !isset($data['ships']) || empty($data['ships']) )
		{
			return;
		}
		foreach( $data['ships'] as $key => $val )
		{
			if( ( isset($val['art_ids']) ) && ( !empty($val['art_ids']) ) )
			{
				foreach( $val['art_ids'] as $id => $a_data )
				{
					Artefact::setTransportID($id, $event_id);
				}
			}
		}
	}

	protected function removeEventIDFromArtefacts( $event_id, $data )
	{
		if( !is_array($data) || !isset($data['ships']) || empty($data['ships']) )
		{
			return;
		}
		foreach( $data['ships'] as $key => $val )
		{
			if( ( isset($val['art_ids']) ) && ( !empty($val['art_ids']) ) )
			{
				foreach( $val['art_ids'] as $id => $a_data )
				{
					Artefact::setTransportID($id, 0);
				}
			}
		}
	}

	/**
	 * Sends a fleet back.
	 *
	 * @param integer $time				Time when event will be triggered
	 * @param integer $start_planet		Planet where event has been triggered
	 * @param integer $userid			MUST BE User that fleet belongs to.
	 * @param integer $target_planet	Destination planet (just for fleet events)
	 * @param array $data				Fleet Data
	 * @param integer $parent_eventid	ID of event that gave birth to this sendBack event. Defaults to null
	 *
	 * @return EventHandler if fail and event_id of added event if success.
	 */
	public function sendBack($time, $start_planet, $userid, $target_planet, array $data, $parent_eventid = null)
	{
		if(!empty($data["alien_actor"]))
		{
			return $this;
		}
		$ifr_params = array(
			"mode" 				=> EVENT_RETURN,
			"planetid" 			=> $start_planet,
			"user" 				=> $userid,
			"destination" 		=> $target_planet,
			"data" 				=> serialize($data),
			"parent_eventid" 	=> $parent_eventid,
			"protected" 		=> 0,
		);
		$params = array(
			"mode" 				=> EVENT_RETURN,
			"start" 			=> time(),
			"time" 				=> $time,
			"planetid" 			=> $start_planet,
			"user" 				=> $userid,
			"destination" 		=> $target_planet,
			"data" 				=> serialize($data),
			"parent_eventid" 	=> $parent_eventid,
			"protected" 		=> 0,
		);
		$key = md5(serialize($ifr_params));
		if(NS::isFirstRun("EV:sendBack:{$key}"))
		{
			$new_event_id = sqlInsert(
				"events",
				$params
			);
			$this->addEventIDToArtefacts($new_event_id, $data);
			return $new_event_id;
		}
		return $this;
	}


	/**
	* Generates an unique key to prevend race conditions.
	*
	* @return string
	*/
	protected function genPrevRaceConditionKey($size = 16)
	{
		return substr(md5(microtime(true) . mt_rand() . mt_rand()), $size);

		// $sid_len = ceil($size / 4);
		// return Str::substring(SID, 0, $sid_len) . Str::substring(md5(microtime(true) . mt_rand(0, 0xfff)), 0, $size - $sid_len);
	}

	protected function queryExpiredEvents($cur_time, $max_batch_time = 10, $limit = 5)
	{
		Core::getQuery()->update("events",
			array( "prev_rc", /* "time", */ "processed", "processed_mode", "processed_time" ),
			array( $this->raceConditionKey, /* $cur_time, */ EVENT_PROCESSED_WAIT, null, $cur_time ),
			"(time <= ".sqlVal($cur_time)." AND processed='".EVENT_PROCESSED_WAIT."' AND prev_rc IS NULL AND (processed_time IS NULL OR processed_time <= ".sqlVal($cur_time-5)."))"
			. " OR (time < ".sqlVal($cur_time - ($max_batch_time + 10))
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
			"e.processed_time = ".sqlVal($cur_time)
			. " and e.processed=".EVENT_PROCESSED_WAIT
			. " and e.prev_rc = " . sqlVal($this->raceConditionKey)
			. " and e.mode != " . EVENT_ALLIANCE_ATTACK_ADDITIONAL // ally event will be served by EVENT_ATTACK_ALLIANCE
			. " ORDER BY e.time ASC, e.eventid ASC");
		return $result;
	}

	/**
	* Executes expired events.
	*
	* @return EventHandler
	*/
	public function goThroughEvents($limit = 10)
	{
		$count = 0;
		$is_external_handler = is_object($this->externalMonitor); // (defined('YII_CONSOLE') && YII_CONSOLE) || (defined('YII_CONSOLE_IS_RUNNING') && YII_CONSOLE_IS_RUNNING);

		$is_console_always = defined('YII_CONSOLE_IS_RUNNING') && YII_CONSOLE_IS_RUNNING;
		$max_batch_time = $is_console_always ? EVENT_BATCH_CONSOLE_PROCESS_TIME : EVENT_BATCH_PROCESS_TIME;

		$cur_time = time();
		$result = $this->queryExpiredEvents($cur_time, $max_batch_time, $limit);

		while( $row = sqlFetch( $result ) )
		{
			$count++;
			if($is_external_handler)
			{
				echo "[EVENT {$row['eventid']}] m:{$row['mode']} u:{$row['userid']}\n";
			}
			$eventid = $row['eventid'];
			if( !$is_console_always && time() - $cur_time >= $max_batch_time )
			{
				if($is_external_handler)
				{
					echo "[EVENT {$row['eventid']}] process_time >= $max_batch_time, break\n";
				}
				break;
			}
			$start_event_process_time = microtime(true);
			// Update By Pk
			sqlUpdate(
				"events",
				array(
					"processed" 		=> EVENT_PROCESSED_START,
					"processed_mode" 	=> $row['mode'],
					"processed_time" 	=> time()
				),
				"eventid = ".sqlVal($eventid)
			);

			$processed_result = EVENT_PROCESSED_OK;
			try
			{
				$data = unserialize($row["data"]);
				$row["data"] = &$data;

				$event_mode = intval($row["mode"]);

				$check_planet_field = false;
				switch($event_mode)
				{
					case EVENT_RESEARCH:
					case EVENT_COLONIZE:
					case EVENT_COLONIZE_RANDOM_PLANET:
					case EVENT_COLONIZE_NEW_USER_PLANET:
					case EVENT_RECYCLING:
					case EVENT_MOON_DESTRUCTION:
					case EVENT_EXPEDITION:
					case EVENT_ROCKET_ATTACK:
					case EVENT_HOLDING:
					case EVENT_ALLIANCE_ATTACK_ADDITIONAL: // Serves as referer to alliance attack
					case EVENT_RETURN:
					case EVENT_TEMP_PLANET_DISAPEAR:
					case EVENT_RUN_SIM_ASSAULT:
					case EVENT_ALIEN_HOLDING:
					case EVENT_ALIEN_HOLDING_AI:
					case EVENT_ALIEN_CHANGE_MISSION_AI:
						break;
					//### Building missions ###//
					case EVENT_BUILD_CONSTRUCTION:
					case EVENT_DEMOLISH_CONSTRUCTION:
					case EVENT_BUILD_FLEET:
					case EVENT_BUILD_DEFENSE:
					case EVENT_REPAIR:
					case EVENT_DISASSEMBLE:
						$check_planet_field = "planetid";
						break;
					//### Fleet missions ###//
					case EVENT_POSITION:
					case EVENT_DELIVERY_UNITS:
					case EVENT_TRANSPORT:
					case EVENT_DELIVERY_RESOURSES:
					case EVENT_DELIVERY_ARTEFACTS:
					case EVENT_ATTACK_SINGLE:
					case EVENT_ATTACK_DESTROY_BUILDING:
					case EVENT_ATTACK_DESTROY_MOON:
					case EVENT_SPY:
					case EVENT_ATTACK_ALLIANCE:
					case EVENT_ATTACK_ALLIANCE_DESTROY_BUILDING:
					case EVENT_ATTACK_ALLIANCE_DESTROY_MOON:
					case EVENT_HALT:
					case EVENT_ALIEN_FLY_UNKNOWN:
					case EVENT_ALIEN_GRAB_CREDIT:
					// case EVENT_ALIEN_HOLDING:
					case EVENT_ALIEN_ATTACK:
					case EVENT_ALIEN_ATTACK_CUSTOM:
					case EVENT_ALIEN_HALT:
						$check_planet_field = "destination";
						break;
				}

				if( $check_planet_field && !isPlanetOccupied( $row[$check_planet_field] ) )
				{
					$this->removeEvent($row["eventid"], true);
					if($is_external_handler)
					{
						echo "[EVENT {$row['eventid']}] planet {$row[$check_planet_field]} ($check_planet_field) is not occupied, event is removed\n";
					}
					continue;
				}

				switch($event_mode)
				{
					//### Building missions ###//
					case EVENT_BUILD_CONSTRUCTION:
						$this->build($row, $data);
						break;
					case EVENT_DEMOLISH_CONSTRUCTION:
						$this->demolish($row, $data);
						break;
					case EVENT_RESEARCH:
						$this->research($row, $data);
						break;
					case EVENT_BUILD_FLEET:
					case EVENT_BUILD_DEFENSE:
						$this->shipyard($row, $data);
						break;
					case EVENT_REPAIR:
						$this->repair($row, $data);
						break;
					case EVENT_DISASSEMBLE:
						$this->disassemble($row, $data);
						break;
					//### Fleet missions ###//
					case EVENT_POSITION:
					case EVENT_DELIVERY_UNITS:
					case EVENT_STARGATE_TRANSPORT:
					case EVENT_STARGATE_JUMP:
						$this->position($row, $data);
						break;
					case EVENT_TELEPORT_PLANET:
						$this->teleportPlanet($row, $data);
						break;
					case EVENT_TRANSPORT:
					case EVENT_DELIVERY_RESOURSES:
					case EVENT_DELIVERY_ARTEFACTS:
						$this->transport($row, $data);
						break;
					case EVENT_COLONIZE:
					case EVENT_COLONIZE_RANDOM_PLANET:
					case EVENT_COLONIZE_NEW_USER_PLANET:
						$this->colonize($row, $data);
						break;
					case EVENT_RECYCLING:
						$this->recycling($row, $data);
						break;
					case EVENT_ATTACK_SINGLE:
					case EVENT_ATTACK_DESTROY_BUILDING:
					case EVENT_ATTACK_DESTROY_MOON:
						$this->attack($row, $data);
						break;
					case EVENT_SPY:
						$this->spy($row, $data);
						break;
					case EVENT_ATTACK_ALLIANCE:
					case EVENT_ATTACK_ALLIANCE_DESTROY_BUILDING:
					case EVENT_ATTACK_ALLIANCE_DESTROY_MOON:
						$this->allianceAttack($row, $data);
						break;
					case EVENT_HALT:
						$this->halt($row, $data);
						break;
					case EVENT_MOON_DESTRUCTION:
						$this->moonDestruction($row, $data);
						break;
					case EVENT_EXPEDITION:
						$this->expedition($row, $data);
						break;
					case EVENT_ROCKET_ATTACK:
						$this->rocketAttack($row, $data);
						break;
					case EVENT_HOLDING:
						$this->holding($row, $data);
						break;
					case EVENT_ALLIANCE_ATTACK_ADDITIONAL: // Serves as referer to alliance attack
						break;
					case EVENT_RETURN:
						$this->fReturn($row, $data);
						break;
					//### Others ###//
					case EVENT_TEMP_PLANET_DISAPEAR:
						$this->destroyPlanet($row, $data);
						break;
					case EVENT_RUN_SIM_ASSAULT:
						$this->runSimAssault($row, $data);
						break;
					//### Artefacts ###//
					case EVENT_ARTEFACT_EXPIRE:
						$this->artefactExpire($row, $data);
						break;
					case EVENT_ARTEFACT_DISAPPEAR:
						$this->artefactDisappear($row, $data);
						break;
					case EVENT_ARTEFACT_DELAY:
						$this->artefactDelay($row, $data);
						break;
					//### Exchange ###//
					case EVENT_EXCH_EXPIRE:
						$this->exchangeExpire($row, $data);
						break;
					case EVENT_EXCH_BAN:
						break;

					case EVENT_ALIEN_ATTACK_CUSTOM:
					case EVENT_ALIEN_FLY_UNKNOWN:
						$this->alienFlyUnknown($row, $data);
						break;

					case EVENT_ALIEN_GRAB_CREDIT:
						$this->alienGrabCredit($row, $data);
						break;

					case EVENT_ALIEN_HOLDING:
						$this->alienHolding($row, $data);
						break;

					case EVENT_ALIEN_HOLDING_AI:
						$this->alienHoldingAI($row, $data);
						break;

					case EVENT_ALIEN_CHANGE_MISSION_AI:
						$this->alienChangeMissionAI($row, $data);
						break;

					case EVENT_ALIEN_ATTACK:
						$this->alienAttack($row, $data);
						break;

					case EVENT_ALIEN_HALT:
						$this->alienHalt($row, $data);
						break;

					default:
						{
							$processed_result 	= EVENT_PROCESSED_ERROR;
							$error_message 		= "Fatal Error: EH-Mode (".intval($row["mode"]).") unknown. Please copy this error and report it to the developer. Error description: ".print_r($row, true)."<br />".print_r($data, true);
							// Update By Pk
							Core::getQuery()->update("events",
								array("processed", "processed_time", "error_message", "processed_dt"),
								array($processed_result, time(), $error_message, round(microtime(true) - $start_event_process_time, 5)),
								"eventid = ".sqlVal($eventid));

							if($is_external_handler)
							{
								echo "[EVENT {$row['eventid']}] unknown event mode {$event_mode}\n";
							}
							// Logger::addMessage($error_message);
						}
						break;
				}
			}
			catch( \Throwable $e )
			{
				// 37.8 STUCK-001: было `Exception` — на PHP 8 не ловит \Error (TypeError, OOM и пр.),
				// событие застревало в PROCESSED_START. \Throwable ловит и Exception, и Error.
				$processed_result 	= EVENT_PROCESSED_ERROR;
				$error_message 		= "EH Exception: " . $e->__toString();
				error_log( "[EVENT {$row['eventid']}] CATCH " . $error_message);
				// Update By Pk
				Core::getQuery()->update("events",
					array("processed", "processed_time", "error_message", "processed_dt"),
					array($processed_result, time(), $error_message, round(microtime(true) - $start_event_process_time, 5)),
					"eventid = ".sqlVal($eventid));
				if($is_external_handler)
				{
					echo "[EVENT {$row['eventid']}] CATCH ".$e->__toString()."\n";
				}
			}
			if( $processed_result == EVENT_PROCESSED_OK )
			{
				// Update By Pk
				sqlUpdate(
					"events",
					array(
						"processed" => EVENT_PROCESSED_OK,
						"processed_time" => time(),
						"processed_dt" => round(microtime(true) - $start_event_process_time, 5),
					),
					"eventid = ".sqlVal($eventid)
						. " AND processed = " . EVENT_PROCESSED_START
						. " AND prev_rc is not null "
						. " AND prev_rc = " . sqlVal( $this->raceConditionKey )
				);
				if($is_external_handler)
				{
					echo "[EVENT {$row['eventid']}] OK\n";
				}
			}
		}
		sqlEnd($result);

		if( !mt_rand(0, 1000) )
		{
			// 37.8 STUCK-003: было 14 дней для PROCESSED_START — слишком долго,
			// игрок ждал две недели до auto-recovery зависшего флота. 3 дня
			// безопасно покрывает max event duration (экспедиция ~24-48ч).
			// NB: cleanup только удаляет event, НЕ возвращает ресурсы/корабли —
			// полноценный recovery остаётся отдельной задачей (см. bugfix-log).
			sqlDelete(
				"events",
				"(processed = ".EVENT_PROCESSED_OK." AND processed_time < ".sqlVal(time() - 60*60*24*7).")"
				. " OR (processed = ".EVENT_PROCESSED_ERROR." AND processed_time < ".sqlVal(time() - 60*60*24*10).")"
				. " OR (processed = ".EVENT_PROCESSED_START." AND processed_time < ".sqlVal(time() - 60*60*24*3).")"
			);
			if($is_external_handler)
			{
				echo "[EVENT {$row['eventid']}] old events deleted\n";
			}
		}

		return $count;
	}

	public function getEventBuildingData($event, $result_field = false)
	{
		$res["duration"] = $event["time"] - $event["start"];
		if(!empty($event["data"]["new_format"]) && $event["data"]["quantity"] > 0)
		{
			$res["duration"] = $event["data"]["duration"] * $event["data"]["quantity"];
		}
		$res["building_time"] = $res["duration"] > 0 ? time() - $event["start"] : 0;
		$res["building_scale"] = $res["duration"] > 0 ? $res["building_time"] / $res["duration"] : 0;

		$created_time = isset($event["data"]["created_time"]) ? $event["data"]["created_time"] : 0;
		$res["event_live_time"] = time() - $created_time;
		if($res["event_live_time"] > EV_ABORT_SAVE_TIME 
			// && $event["mode"] != EVENT_RESEARCH
			)
		{
			$res["abort_res_scale"] = min(1, max(0, 1 - $res["building_scale"]));
			switch($event["mode"])
			{
			case EVENT_BUILD_CONSTRUCTION: // building
			case EVENT_DEMOLISH_CONSTRUCTION: // building
			case EVENT_RESEARCH: // research
				$res["abort_res_scale"] = min($res["abort_res_scale"], EV_ABORT_MAX_BUILD_PERCENT / 100.0);
				break;

			case EVENT_BUILD_FLEET:
			case EVENT_BUILD_DEFENSE:
				$res["abort_res_scale"] = min($res["abort_res_scale"], EV_ABORT_MAX_SHIPYARD_PERCENT / 100.0);
				break;

			case EVENT_REPAIR:
				$res["abort_res_scale"] = min($res["abort_res_scale"], EV_ABORT_MAX_REPAIR_PERCENT / 100.0);
				break;

			case EVENT_DISASSEMBLE:
				$res["abort_res_scale"] = min($res["abort_res_scale"], EV_ABORT_MAX_DISASSEMBLE_PERCENT / 100.0);
				break;

			case EVENT_POSITION:
			case EVENT_TRANSPORT:
			case EVENT_COLONIZE:
			case EVENT_COLONIZE_RANDOM_PLANET:
			case EVENT_COLONIZE_NEW_USER_PLANET:
			case EVENT_RECYCLING:
			case EVENT_ATTACK_SINGLE:
			case EVENT_ATTACK_DESTROY_BUILDING:
			case EVENT_ATTACK_DESTROY_MOON:
			case EVENT_SPY:
			case EVENT_ATTACK_ALLIANCE:
			case EVENT_ATTACK_ALLIANCE_DESTROY_BUILDING:
			case EVENT_ATTACK_ALLIANCE_DESTROY_MOON:
			case EVENT_ALLIANCE_ATTACK_ADDITIONAL:
				$res["abort_res_scale"] = min($res["abort_res_scale"], EV_ABORT_MAX_FLY_PERCENT / 100.0);
				break;
			}
		}
		else
		{
			$res["abort_res_scale"] = 1;
		}

		return $result_field ? $res[$result_field] : $res;
	}

	/**
	* Removes an event from the event stack.
	*
	* @param integer	Event id
	*
	* @return EventHandler
	*/
	public function removeEvent($eventid, $force_remove = false, $save_scale = true)
	{
		if(!NS::isFirstRun("EV:removeEvent:$eventid"))
		{
			return false;
		}
		if( $force_remove || !NS::getUser()->get("umode") ) // if force or not on vacation
		{
			if( 0 && $this->events_processed )
			{
				$event = $this->eventStack[$eventid];
				if( !isset($event) )
				{
					return false;
				}
			}
			else
			{
				$joins	= "LEFT JOIN ".PREFIX."planet p1 ON (p1.planetid = e.planetid) ";
				$joins .= "LEFT JOIN ".PREFIX."planet p2 ON (p2.planetid = e.destination) ";
				$joins .= "LEFT JOIN ".PREFIX."galaxy g1 ON (g1.planetid = e.planetid) ";
				$joins .= "LEFT JOIN ".PREFIX."galaxy g2 ON (g2.planetid = e.destination) ";
				$joins .= "LEFT JOIN ".PREFIX."user u1 ON (u1.userid = p1.userid) ";
				$joins .= "LEFT JOIN ".PREFIX."user u2 ON (u2.userid = p2.userid)";
				$select = array(
					"e.*", "e.user as userid", "u1.username",
					"u2.username AS destname", "p1.planetname",
					"p2.planetname AS destplanet", "g1.galaxy",
					"g1.system", "g1.position", "g2.galaxy AS galaxy2",
					"g2.system AS system2", "g2.position AS position2"
				);
				$event = sqlSelectRow(
					"events e",
					$select,
					$joins,
					"e.eventid = ".sqlVal($eventid)
						// ." AND processed in (".sqlArray(EVENT_PROCESSED_WAIT, EVENT_PROCESSED_START).")"
						." AND processed in (".sqlArray(EVENT_PROCESSED_WAIT).")"
                        // ." AND e.time > ".sqlVal(time()+RETREAT_EVENT_BLOCK_END_TIME)
				);
				if($event["time"] <= time()+RETREAT_EVENT_BLOCK_END_TIME){
					$skip_event = true;
					if(time() > $event["time"] + 60*60*12){
						$parent_event_exist = sqlSelectField("events", "count(*)", null, "eventid=".sqlVal($f["parent_eventid"]));
						// echo "parent_event_exist <pre>"; print_r($parent_event_exist); echo "</pre>"; exit;
						if(!$parent_event_exist){
							$skip_event = false;
							$event["start"] = time()-5;
							$event["time"] = time();
						}
					}
					if($skip_event){
						return false;
					}
				}
				// echo "event <pre>"; print_r($event); echo "</pre>"; exit;
				if( !isset($event["data"]) )
				{
					return false;
				}

				$event["data"] = unserialize($event["data"]);
			}

			$userid		= $event["user"];
			$planetid	= $event["planetid"];
			$data		= &$event["data"];
			switch($event["mode"])
			{
				case EVENT_BUILD_CONSTRUCTION:
				case EVENT_DEMOLISH_CONSTRUCTION:
				case EVENT_RESEARCH:
				case EVENT_BUILD_FLEET:
				case EVENT_BUILD_DEFENSE:
				case EVENT_REPAIR:
				case EVENT_DISASSEMBLE:
					if( $save_scale === true )
					{
						$save_scale = $this->getEventBuildingData($event, "abort_res_scale");
					}
					$data["metal"]		= round($data["metal"] * $save_scale);
					$data["silicon"]	= round($data["silicon"] * $save_scale);
					$data["hydrogen"]	= round($data["hydrogen"] * $save_scale);

					if($event["mode"] == EVENT_REPAIR || $event["mode"] == EVENT_DISASSEMBLE)
					{
						logShipsChanging($data["buildingid"], $event["planetid"], $data["quantity"], 1, "[removeEvent] repair");
						$exist_quantity = sqlSelectField(
							"unit2shipyard",
							"quantity",
							"",
							"unitid = ".sqlVal($data["buildingid"])
								. " AND planetid = ".sqlVal($event["planetid"])
						);
						if($exist_quantity > 0)
						{
							// Update By Pk
							$sql = "UPDATE ".PREFIX."unit2shipyard SET "
								. addQuantitySetSql(array("damaged" => $data["damaged"], "quantity" => $data["quantity"], "shell_percent" => $data["shell_percent"]))
								. " WHERE unitid = ".sqlVal($data["buildingid"])." AND planetid = ".sqlVal($event["planetid"]);
							sqlQuery($sql);
						}
						else
						{
							Core::getQuery()->insert(
								"unit2shipyard",
								array("unitid", "planetid", "quantity", "damaged", "shell_percent"),
								array($data["buildingid"], $event["planetid"], $data["quantity"], $data["damaged"], $data["shell_percent"])
							);
						}
						$points = $data["points"] * $data["quantity"];
						// Update By Pk
						sqlQuery("UPDATE ".PREFIX."user SET "
							. " points = points + ".sqlVal($points)
							. ", u_points = u_points + ".sqlVal($points)
							. ", u_count = u_count + ".sqlVal($data["quantity"])
                            . ", ".updateDmPointsSetSql()
							. " WHERE userid = ".sqlVal($userid));
					}
					NS::updateUserRes(array(
						"type" => RES_UPDATE_CANCEL,
						"event_mode" => $event["mode"],
						"reload_planet" => false,
						"ownerid" => $eventid,
						"userid" => $userid,
						"planetid" => $planetid,
						"metal" => $data["metal"],
						"silicon" => $data["silicon"],
						"hydrogen" => $data["hydrogen"],
						));
					break;

				case EVENT_ATTACK_ALLIANCE:
					$data["ret_consumption"] = round($data["consumption"] *
						($save_scale === true ? $this->getEventBuildingData($event, "abort_res_scale") : $save_scale));

					$result = sqlSelect("events",
						"*", // array("eventid", "start", "destination", "user", "planetid", "data"),
						"",
						"mode = ".EVENT_ALLIANCE_ATTACK_ADDITIONAL." AND parent_eventid = ".sqlVal($eventid)." AND processed=".EVENT_PROCESSED_WAIT
						);

					// if(NS::getUser()->get("userid") == 2){ debug_var($result, "[removeEvent] result"); exit; }

					while($row = sqlFetch($result))
					{
						$row["data"] = unserialize($row["data"]);

						$row["data"]["ret_consumption"] = round($row["data"]["consumption"] *
							($save_scale === true ? $this->getEventBuildingData($row, "abort_res_scale") : $save_scale));
						$row["data"]["oldmode"] = $row["mode"];

						// if(NS::getUser()->get("userid") == 2){ debug_var($row["data"], "[removeEvent] row data"); }

						$this->sendBack(time() + (time() - $row["start"]), $row["destination"], $row["user"], $row["planetid"],
							$row["data"], $row["eventid"]);

						// Hook::event("EH_REMOVE_EVENT", array($row["eventid"], &$row, &$row["data"]));
						// Update By Pk
						Core::getQuery()->update("events", array("processed", "processed_time", "error_message"),
							array(EVENT_PROCESSED_OK, time(), "removed"),
							"eventid = ".sqlVal($row["eventid"]));
					}
					sqlEnd($result);

					$data["oldmode"] = $event["mode"];
					$this->sendBack(time() + (time() - $event["start"]), $event["destination"], $event["user"], $event["planetid"],
						$data, $event["eventid"]);

					// if(NS::getUser()->get("userid") == 2){ debug_var($event, "[removeEvent]"); exit; }

					break;

				case EVENT_RETURN:
				case EVENT_DELIVERY_UNITS:
				case EVENT_DELIVERY_RESOURSES:
					// these events can't be removed
					return false;
				case EVENT_ARTEFACT_DELAY:
				case EVENT_ARTEFACT_EXPIRE:
				case EVENT_ARTEFACT_DISAPPEAR:
					break;
				case EVENT_EXCH_EXPIRE:
				case EVENT_EXCH_BAN:
					break;
				case EVENT_ALIEN_FLY_UNKNOWN:
				case EVENT_ALIEN_HOLDING:
				case EVENT_ALIEN_ATTACK:
				case EVENT_ALIEN_HALT:
				case EVENT_ALIEN_GRAB_CREDIT:
				case EVENT_ALIEN_ATTACK_CUSTOM:
				case EVENT_TEMP_PLANET_DISAPEAR:
				case EVENT_ALIEN_HOLDING_AI:
				case EVENT_ALIEN_CHANGE_MISSION_AI:
				case EVENT_RUN_SIM_ASSAULT:
				case EVENT_TOURNAMENT_SCHEDULE:
				case EVENT_TOURNAMENT_RESCHEDULE:
				case EVENT_TOURNAMENT_PARTICIPANT:
					break;

				case EVENT_HOLDING:
					if( !$force_remove && empty($data["consumption"]) && !empty($data["back_consumption"]) && $event["destination"] != $event["planetid"] )
					{
						return false;
					}
					$data["oldmode"] = $event["mode"];
					$this->sendBack(time() + $data["time"], $event["destination"], $event["user"], $event["planetid"], $data, $event["eventid"]);
					break;

				default:
					if(empty($event["planetid"]))
					{
						return false;
					}
					if(isset($data["consumption"]))
					{
						$data["ret_consumption"] = round($data["consumption"] *
							($save_scale === true ? $this->getEventBuildingData($event, "abort_res_scale") : $save_scale));
					}
					$data["oldmode"] = $event["mode"];
					$this->sendBack(time() + (time() - $event["start"]), $event["destination"], $event["user"], $event["planetid"], $data, $event["eventid"]);
					break;
			}
			// if($event["mode"] != EVENT_RETURN)
			{
				// Hook::event("EH_REMOVE_EVENT", array($eventid, &$event, &$data));
				// Core::getQuery()->delete("events", "eventid = '".$event["eventid"]."'");
				if ( !is_string($force_remove) )
				{
					$force_remove = $force_remove ? "force_removed" : "removed";
				}
				// Update By Pk
				Core::getQuery()->update("events", array("processed", "processed_time", "error_message"),
					array(EVENT_PROCESSED_OK, time(), $force_remove),
					"eventid = ".sqlVal($event["eventid"]));
				$this->removeEventIDFromArtefacts($eventid, $data);
				if(isset($data["used_exp_points"]))
				{
					// Update By Pk
					sqlQuery("UPDATE ".PREFIX."user SET be_points = be_points + ".sqlVal($data["used_exp_points"])." WHERE userid=".sqlVal($event["user"]));
				}

			}
		}
		return true;
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

	/**
	* Sets all user-related events.
	*
	* @return EventHandler
	*/
	protected function setEventStack()
	{
		/*
		$joins	= "LEFT JOIN ".PREFIX."planet p1 ON p1.planetid = e.planetid ";
		$joins .= "LEFT JOIN ".PREFIX."planet p2 ON p2.planetid = e.destination ";
		$joins .= "LEFT JOIN ".PREFIX."galaxy g1 ON g1.planetid = e.planetid ";
		$joins .= "LEFT JOIN ".PREFIX."galaxy g2 ON g2.planetid = e.destination ";
		$joins .= "LEFT JOIN ".PREFIX."galaxy g1m ON g1m.moonid = e.planetid ";
		$joins .= "LEFT JOIN ".PREFIX."galaxy g2m ON g2m.moonid = e.destination ";
		$joins .= "LEFT JOIN ".PREFIX."user u1 ON u1.userid = p1.userid ";
		$joins .= "LEFT JOIN ".PREFIX."user u2 ON u2.userid = p2.userid";
		$select = array("e.*", "u1.username", "u2.username AS destname", "p1.planetname", "p2.planetname AS destplanet",
			"IFNULL(g1.galaxy, g1m.galaxy) as galaxy",
			"IFNULL(g1.system, g1m.system) as system",
			"IFNULL(g1.position, g1m.position) as position",
			"IFNULL(g2.galaxy, g2m.galaxy) as galaxy2",
			"IFNULL(g2.system, g2m.system) as system2",
			"IFNULL(g2.position, g2m.position) as position2",
		);
		$result = sqlSelect("events e", $select, $joins,
			"(u1.userid = ".sqlUser()." OR u2.userid = ".sqlUser().")"
			. " AND processed=".EVENT_PROCESSED_WAIT, "e.time ASC");
		*/
		$result = sqlQuery("
			SELECT *
			FROM na_event_src
			WHERE userid = ".sqlUser()."
			AND processed = ".EVENT_PROCESSED_WAIT."
			UNION
			SELECT *
			FROM na_event_src
			WHERE startuserid = ".sqlUser()." AND userid != ".sqlUser()." AND mode != ".EVENT_RETURN."
			AND processed = ".EVENT_PROCESSED_WAIT."
			UNION
			SELECT *
			FROM na_event_dest
			WHERE destuserid = ".sqlUser()."
			AND processed = ".EVENT_PROCESSED_WAIT."
			ORDER BY `time`, `eventid`
			");
		while($row = sqlFetch($result))
		{
			$row["data"] = unserialize($row["data"]);
			$this->eventStack[$row["eventid"]] = $row;
		}
		sqlEnd($result);
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

	/**
	* Loads the main fleet of an alliance attack.
	*
	* @param integer	Event id
	*
	* @return array	Event data
	*/
	public function getMainFormationFleet($eventid)
	{
		$joins	= "LEFT JOIN ".PREFIX."planet p ON (p.planetid = e.planetid) ";
		$joins .= "LEFT JOIN ".PREFIX."galaxy g ON (g.planetid = e.planetid) ";
		$joins .= "LEFT JOIN ".PREFIX."user u ON (u.userid = p.userid) ";
		$select = array("e.*", "u.username", "p.planetname", "g.galaxy", "g.system", "g.position");
		$row = sqlSelectRow("events e", $select, $joins, "e.eventid = ".sqlVal($eventid), "e.time ASC");
		if($row)
		{
			$row["data"] = unserialize($row["data"]);
			return $row;
		}
		return false;
	}

	//#########################################################################//
	//############################# EVENT GETTERS #############################//
	//#########################################################################//

	/**
	* Building events of the current used planet.
	*
	* @return mixed Event data or false.
	*/
	public function getFirstBuildingEvent()
	{
		$ev = $this->getBuildingEvents();
		return reset($ev);
	}

	/**
	* Building current events of the current used planet.
	*
	* @return mixed Event data or false.
	*/
	public function getBuildingEvents()
	{
		$return = array();
		$cur_user_id = NS::getUser()->get("userid");
		$cur_planet_id = NS::getUser()->get("curplanet");
		// debug_var($this->eventStack, "[eventStack] user: $cur_user_id, planet: $cur_planet_id");
		foreach($this->eventStack as $value)
		{
			if(($value["mode"] == EVENT_BUILD_CONSTRUCTION || $value["mode"] == EVENT_DEMOLISH_CONSTRUCTION)
				&& $value["user"] == $cur_user_id
				&& $value["planetid"] == $cur_planet_id)
			{
				$return[$value["eventid"]] = $value;
			}
		}
		// if(count($return) == 0) { return false; }
		//print_r($return);die;
		return $return;
	}

	/**
	 *
	 * Returns array of exchanges, where curr user was banned
	 */
	public function getExchBans()
	{
		if (empty($this->eventStack))
		{
			self::setEventStack();
		}
		$return = array();
		foreach ( $this->eventStack as $value )
		{
			if ( $value["mode"] == EVENT_EXCH_BAN )
			{
				$return[] = $value['data']['exch_id'];
			}
		}
		return $return;
	}

	/**
	* Repairing current events of the current used planet.
	*
	* @return mixed Event data or false.
	*/
	public function getRepairEvents()
	{
		$return = array();
		$cur_user_id = NS::getUser()->get("userid");
		$cur_planet_id = NS::getUser()->get("curplanet");
		// debug_var($this->eventStack, "[eventStack] user: $cur_user_id, planet: $cur_planet_id");
		foreach($this->eventStack as $value)
		{
			if(($value["mode"] == EVENT_REPAIR || $value["mode"] == EVENT_DISASSEMBLE)
				&& $value["user"] == $cur_user_id
				&& $value["planetid"] == $cur_planet_id)
			{
				$return[$value["eventid"]] = $value;
			}
		}
		return $return;
	}

	public function getFirstPlanetBuildingEvent($planet_id)
	{
		$cur_user_id = NS::getUser()->get("userid");
		foreach($this->eventStack as $value)
		{
			if(($value["mode"] == EVENT_BUILD_CONSTRUCTION || $value["mode"] == EVENT_DEMOLISH_CONSTRUCTION)
				&& $value["user"] == $cur_user_id
				&& $value["planetid"] == $planet_id)
			{
				return $value;
			}
		}
		return false;
	}

	/**
	* Current research event.
	*
	* @return mixed Event data or false.
	*/
	public function getFirstResearchEvent()
	{
		$ev = $this->getResearchEvents();
		return reset($ev);
	}

	/**
	* Current research events.
	*
	* @return mixed Event data or false.
	*/
	public function getResearchEvents()
	{
		$return = array();
		$cur_user_id = NS::getUser()->get("userid");
		foreach($this->eventStack as $value)
		{
			if($value["mode"] == EVENT_RESEARCH && $value["user"] == $cur_user_id)
			{
				$return[$value["eventid"]] = $value;
			}
		}
		// if(count($return) == 0) { return false; }
		return $return;
	}

	/**
	* Shipyard events for the current used planet.
	*
	* @return mixed Event data or false.
	*/
	public function getShipyardEvents()
	{
		$return = array();
		$cur_user_id = NS::getUser()->get("userid");
		$cur_planet_id = NS::getUser()->get("curplanet");
		foreach($this->eventStack as $value)
		{
			if(($value["mode"] == EVENT_BUILD_FLEET || $value["mode"] == EVENT_BUILD_DEFENSE)
				&& $value["user"] == $cur_user_id
				&& $value["planetid"] == $cur_planet_id
				)
			{
				$return[$value["eventid"]] = $value;
			}
		}
		// if(count($return) == 0) { return false; }
		return $return;
	}

	/**
	* Shipyard events for the current used planet.
	*
	* @return mixed Event data or false.
	*/
	public function getShipyardModeEvents($mode)
	{
		$return = array();
		if($mode == EVENT_BUILD_FLEET || $mode == EVENT_BUILD_DEFENSE)
		{
			$cur_user_id = NS::getUser()->get("userid");
			$cur_planet_id = NS::getUser()->get("curplanet");
			foreach($this->eventStack as $value)
			{
				if($value["mode"] == $mode
					&& $value["user"] == $cur_user_id
					&& $value["planetid"] == $cur_planet_id)
				{
					$return[$value["eventid"]] = $value;
				}
			}
		}
		// if(count($return) == 0) { return false; }
		return $return;
	}

	/**
	* Own Fleet events.
	*
	* @return mixed Event data or false.
	*/
	public function getOwnFleetEvents()
	{
		$return = array();
		$cur_user_id = NS::getUser()->get("userid");
		foreach($this->eventStack as $value)
		{
			if( ( $value["mode"] >= EVENT_MARK_FIRST_FLEET && $value["mode"] <= EVENT_MARK_LAST_FLEET ) && $value["user"] == $cur_user_id)
			{
				$value['exped_slot'] = false;
				$value['fleet_slot'] = false;
				if( $this->isFleetSlotUsed($value) )
				{
					$value['fleet_slot'] = true;
				}
				if( $this->isExpeditionSlotUsed($value) )
				{
					$value['exped_slot'] = true;
				}
				$return[$value["eventid"]] = $value;
			}
		}
		return $return;
	}

	public function isFleetSlotUsed($event)
	{
		static $skip_events = array(
			EVENT_EXPEDITION => 1,
			EVENT_DELIVERY_UNITS => 1,
			EVENT_DELIVERY_RESOURSES => 1,
			EVENT_DELIVERY_ARTEFACTS => 1,
		);
		if($event["mode"] < EVENT_MARK_FIRST_FLEET || $event["mode"] > EVENT_MARK_LAST_FLEET)
		{
			return false;
		}
		if(isset($skip_events[$event["mode"]]))
		{
			return false;
		}
		if($event["mode"] == EVENT_RETURN && isset($event['data']["oldmode"]) && isset($skip_events[$event['data']["oldmode"]]))
		{
			return false;
		}
		return true;
	}

	public function getUsedFleetSlots()
	{
		$count = 0;
		$cur_user_id = NS::getUser()->get("userid");
		foreach($this->eventStack as $event)
		{
			if( $event["user"] == $cur_user_id)
			{
				$count += (int)$this->isFleetSlotUsed($event);
			}
		}
		return $count;
	}

	public function isExpeditionSlotUsed($event)
	{
		if($event["mode"] == EVENT_EXPEDITION)
		{
			return true;
		}
		// the return from EXP does not use exp slot
		if(0 && $event["mode"] == EVENT_RETURN && isset($event['data']["oldmode"]) && $event['data']["oldmode"] == EVENT_EXPEDITION)
		{
			return true;
		}
		return false;
	}

	public function getUsedExpeditionSlots()
	{
		$count = 0;
		$cur_user_id = NS::getUser()->get("userid");
		foreach($this->eventStack as $event)
		{
			if($event["user"] == $cur_user_id)
			{
				$count += (int)$this->isExpeditionSlotUsed($event);
			}
		}
		return $count;
	}

	/**
	* Fleet events.
	*
	* @return mixed Event data or false.
	*/
	public function getFleetEvents()
	{
		$return = array();
		$cur_user_id = NS::getUser()->get("userid");
		foreach($this->eventStack as $value)
		{
			if(
				($value["mode"] >= EVENT_MARK_FIRST_FLEET && $value["mode"] <= EVENT_MARK_LAST_FLEET)
				|| $value["mode"] == EVENT_TEMP_PLANET_DISAPEAR
			)
			{
				if((/*$value["mode"] == EVENT_RETURN ||*/ $value["mode"] == EVENT_RECYCLING) && $value["user"] != $cur_user_id)
				{
				}
				else
				{
					$return[$value["eventid"]] = $value;
				}
			}
		}
		// if(count($return) == 0) { return false; }
		return $return;
	}

	/**
	* Return fleet events of the current planet.
	* (Used to check if colony can be deleted).
	*
	* @return mixed The events or false.
	*/
	public function getPlanetFleetEvents()
	{
		$return = array();
		$cur_planet_id = NS::getUser()->get("curplanet");
		foreach($this->eventStack as $value)
		{
			if($value["mode"] >= EVENT_MARK_FIRST_FLEET && $value["mode"] <= EVENT_MARK_LAST_FLEET
				&& ($value["planetid"] == $cur_planet_id || $value["destination"] == $cur_planet_id))
			{
				$return[$value["eventid"]] = $value;
			}
		}
		// if(count($return) == 0) { return false; }
		return $return;
	}

	public function getEvent($eventid)
	{
		if( !isset($this->eventStack[$eventid]) )
		{
			return null;
		}
		return $this->eventStack[$eventid];
	}

	/**
	* Checks if research lab is currently upgrading.
	*
	* @return boolean
	*/
	public function canReasearch()
	{
		foreach($this->eventStack as $value)
		{
			if($value["mode"] == EVENT_BUILD_CONSTRUCTION && $value["data"]["buildingid"] == UNIT_RESEARCH_LAB) // && $value["start"] < time())
			{
				return false;
			}
		}
		return true;
	}

	/**
	* Checks if shipyard or nanit factory is currently upgrading.
	*
	* @return boolean
	*/
	public function canBuildShipyardUnits()
	{
		$cur_planet_id = NS::getUser()->get("curplanet");
		foreach($this->eventStack as $value)
		{
			if($value["planetid"] == $cur_planet_id
				&& $value["mode"] == EVENT_BUILD_CONSTRUCTION
				&& ($value["data"]["buildingid"] == UNIT_NANO_FACTORY || $value["data"]["buildingid"] == UNIT_SHIPYARD))
			{
				return false;
			}
		}
		return true;
	}

	/**
	* Checks if shipyard or nanit factory is currently upgrading.
	*
	* @return boolean
	*/
	public function canBuildDefenseUnits()
	{
		$cur_planet_id = NS::getUser()->get("curplanet");
		foreach($this->eventStack as $value)
		{
			if($value["planetid"] == $cur_planet_id
				&& $value["mode"] == EVENT_BUILD_CONSTRUCTION
				&& ($value["data"]["buildingid"] == UNIT_NANO_FACTORY || $value["data"]["buildingid"] == UNIT_DEFENSE_FACTORY))
			{
				return false;
			}
		}
		return true;
	}

	/**
	* Checks if repair is possible.
	*
	* @return boolean
	*/
	public function canRepairUnits()
	{
		$cur_planet_id = NS::getUser()->get("curplanet");
		foreach($this->eventStack as $value)
		{
			if($value["planetid"] == $cur_planet_id
				&& $value["mode"] == EVENT_BUILD_CONSTRUCTION
				&& ($value["data"]["buildingid"] == UNIT_NANO_FACTORY || $value["data"]["buildingid"] == UNIT_REPAIR_FACTORY || $value["data"]["buildingid"] == UNIT_MOON_REPAIR_FACTORY))
			{
				return false;
			}
		}
		return true;
	}

	public function canTeleportPlanet()
	{
		if(time() - NS::getUser()->get("planet_teleport_time") < PLANET_TELEPORT_MIN_INTERVAL_TIME)
		{
			return false;
		}
		$cur_user_id = NS::getUser()->get("userid");
		foreach($this->eventStack as $value)
		{
			if($value["mode"] == EVENT_TELEPORT_PLANET && $value["user"] == $cur_user_id)
			{
				return false;
			}
		}
		return true;
	}

	/**
	* Returns working rockets.
	*
	* @return integer Number of working rockets in shipyard queue
	*/
	public function getWorkingRockets()
	{
		$rockets = 0;
		$cur_user_id = NS::getUser()->get("userid");
		$cur_planet_id = NS::getUser()->get("curplanet");
		foreach($this->eventStack as $value)
		{
			if(($value["mode"] == EVENT_BUILD_DEFENSE || $value["mode"] == EVENT_REPAIR || $value["mode"] == EVENT_DISASSEMBLE)
				&& $value["user"] == $cur_user_id
				&& $value["planetid"] == $cur_planet_id
				)
			{
				if($value["data"]["buildingid"] == UNIT_INTERCEPTOR_ROCKET)
				{
					$rockets += !isset($value["data"]["quantity"]) ? 1 : $value["data"]["quantity"];
				}
				else if($value["data"]["buildingid"] == UNIT_INTERPLANETARY_ROCKET)
				{
					$rockets += !isset($value["data"]["quantity"]) ? 2 : $value["data"]["quantity"] * 2;
				}
			}
		}
		return $rockets;
	}

	/**
	* Returns working shields.
	*
	* @return integer Number of working rockets in shipyard queue
	*/
	public function getWorkingShields()
	{
		$shields = 0;
		$cur_user_id = NS::getUser()->get("userid");
		$cur_planet_id = NS::getUser()->get("curplanet");
		foreach($this->eventStack as $value)
		{
			if(($value["mode"] == EVENT_BUILD_DEFENSE || $value["mode"] == EVENT_REPAIR || $value["mode"] == EVENT_DISASSEMBLE)
				&& $value["user"] == $cur_user_id
				&& $value["planetid"] == $cur_planet_id
				)
			{
				if($value["data"]["buildingid"] == UNIT_SMALL_SHIELD)
				{
					$shields += !isset($value["data"]["quantity"]) ? 1 : $value["data"]["quantity"] * 1;
				}
				else if($value["data"]["buildingid"] == UNIT_LARGE_SHIELD)
				{
					$shields += !isset($value["data"]["quantity"]) ? 5 : $value["data"]["quantity"] * 5;
				}
				else if($value["data"]["buildingid"] == UNIT_SMALL_PLANET_SHIELD)
				{
					$shields += !isset($value["data"]["quantity"]) ? 10 : $value["data"]["quantity"] * 10;
				}
				else if($value["data"]["buildingid"] == UNIT_LARGE_PLANET_SHIELD)
				{
					$shields += !isset($value["data"]["quantity"]) ? 20 : $value["data"]["quantity"] * 40;
				}
			}
		}
		return $shields;
	}

	/**
	* Returns working exchange units.
	*
	* @return integer Number of working rockets in shipyard queue
	*/
	public function getWorkingExchangeUnits()
	{
		$quantity = 0;
		$cur_user_id = NS::getUser()->get("userid");
		$cur_planet_id = NS::getUser()->get("curplanet");
		foreach($this->eventStack as $value)
		{
			if(($value["mode"] == EVENT_BUILD_DEFENSE || $value["mode"] == EVENT_BUILD_FLEET
				|| $value["mode"] == EVENT_REPAIR || $value["mode"] == EVENT_DISASSEMBLE)
				&& $value["user"] == $cur_user_id
				&& $value["planetid"] == $cur_planet_id
				&& ($value["data"]["buildingid"] == UNIT_EXCH_SUPPORT_SLOT || $value["data"]["buildingid"] == UNIT_EXCH_SUPPORT_RANGE)
				)
			{
				$quantity += !isset($value["data"]["quantity"]) ? 1 : $value["data"]["quantity"];
			}
		}
		return $quantity;
	}

	public function getWorkingUnits($unitid, $mode)
	{
		//$wmode = EVENT_BUILD_DEFENSE;
		$units = 0;
		$cur_user_id = NS::getUser()->get("userid");
		$cur_planet_id = NS::getUser()->get("curplanet");
		foreach($this->eventStack as $value)
		{
			if(($value["mode"] == $mode || $value["mode"] == EVENT_REPAIR || $value["mode"] == EVENT_DISASSEMBLE)
				&& $value["user"] == $cur_user_id
				&& $value["planetid"] == $cur_planet_id
				&& $value["data"]["buildingid"] == $unitid
				)
			{
				$units += !isset($value["data"]["quantity"]) ? 1 : $value["data"]["quantity"];;
			}
		}
		return $units;
	}

	public function getRepairUsage()
	{
		$repair_usage = 0;
		$cur_user_id = NS::getUser()->get("userid");
		$cur_planet_id = NS::getUser()->get("curplanet");
		foreach($this->eventStack as $value)
		{
			if(($value["mode"] == EVENT_REPAIR || $value["mode"] == EVENT_DISASSEMBLE)
				&& $value["user"] == $cur_user_id
				&& $value["planetid"] == $cur_planet_id)
			{
				$repair_usage += $value["data"]["repair_usage"];
			}
		}
		return $repair_usage;
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
				"userid" 		=> $cur_event["user"],
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
			foreach($events as $event)
			{
				if( $cur_event["data"]["batchkey"] == $event["data"]["batchkey"] )
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

	/**
	* For all event methods. Used from AlienAI
	*
	* @param array Database row of the current event.
	* @param array Specific data of the current event.
	*
	* @return EventHandler
	*/

	public function attack($row, $data)
	{
		require_once(APP_ROOT_DIR."game/Assault.class.php");
		$Assault = new Assault($this, $row["eventid"], $row["destination"], $row["destuser"]);
		$Assault->addParticipant(1, $row["userid"], $row["planetid"], $row["time"], $data, $row["eventid"]);
		$Assault->startAssault(
			$data["galaxy"],
			$data["system"],
			$data["position"],
			array("mode" => $row["mode"])
		); // alliance attack is served by allianceAttack
		return $this;
	}

	protected function build($row, $data)
	{
		// Hook::event("EH_UPGRADE_BUILDING", array(&$row, &$data));
		$points = round(($data["metal"] + $data["silicon"] + $data["hydrogen"]) * RES_TO_BUILD_POINTS, POINTS_PRECISION);

		$add_row = sqlSelectRow("building2planet", "added", "", "buildingid = ".sqlVal($data["buildingid"])." AND planetid = ".sqlVal($row["planetid"]));
		$insert = $add_row ? false : true;
		if($insert)
		{
			sqlInsert("building2planet", array(
				"planetid" => $row["planetid"],
				"buildingid" => $data["buildingid"],
				"level" => 1));

			if($data["buildingid"] == UNIT_EXCHANGE)
			{
				sqlInsert("exchange", array(
					"eid" => $row["planetid"],
					"uid" => $row["userid"],
					"title" => Core::getLang()->get("MENU_STOCK"),
					"fee" => EXCH_FEE_MIN,
					"def_fee" => EXCH_FEE_MIN,
					"comission" => EXCH_COMMISSION_MIN));
			}
		}
		else
		{
			// Update By Pk
			sqlUpdate("building2planet", array(
				"level" => $data["level"] + $add_row["added"]
			), "buildingid = ".sqlVal($data["buildingid"])." AND planetid = ".sqlVal($row["planetid"]));
		}

		// Update By Pk
		sqlQuery("UPDATE ".PREFIX."user SET "
			. " points = points + ".sqlVal($points)
			. ", b_points = b_points + ".sqlVal($points)
			. ", b_count = b_count + 1"
            . ", ".updateDmPointsSetSql()
			. " WHERE userid = ".sqlVal($row["userid"]));

		if(in_array($data["buildingid"], array(UNIT_METALMINE, UNIT_SILICON_LAB, UNIT_HYDROGEN_LAB)))
		{
			updateMinerPoints($row["userid"], $points, true);
		}

        NS::updateProfession($row['userid']);
		AchievementsService::processAchievements($row["userid"], $row["planetid"]);

		return $this;
	}

	protected function demolish($row, $data)
	{
		// Hook::event("EH_DOWNGRADE_BUILDING", array(&$row, &$data));
		// 37.8 OVF-004: было `metal+metal+metal` (typo legacy oxsar2), теперь как в строке 2201/2853/3021
		$points = round(($data["metal"] + $data["silicon"] + $data["hydrogen"]) * RES_TO_BUILD_POINTS, POINTS_PRECISION);

		$add_row = sqlSelectRow("building2planet", "added", "", "buildingid = ".sqlVal($data["buildingid"])." AND planetid = ".sqlVal($row["planetid"]));

		if($data["level"] + $add_row["added"] > 0)
		{
			// Update By Pk
			sqlUpdate("building2planet", array(
				"level" => $data["level"] + $add_row["added"]
			), "buildingid = ".sqlVal($data["buildingid"])." AND planetid = ".sqlVal($row["planetid"]));
		}
		else
		{
			sqlDelete("building2planet", "buildingid = ".sqlVal($data["buildingid"])." AND planetid = ".sqlVal($row["planetid"]));
			if ($data["buildingid"] == UNIT_EXCHANGE)
			{
				sqlDelete("exchange", "eid=".sqlVal($row["planetid"]));
			}
		}
		// Update By Pk
		sqlQuery("UPDATE ".PREFIX."user SET "
			. " points = points - ".sqlVal($points)
			. ", b_points = b_points - ".sqlVal($points)
			. ", b_count = b_count - 1"
			. " WHERE userid = ".sqlVal($row["userid"]));

        NS::updateProfession($row['userid']);
		return $this;
	}

	protected function research($row, $data)
	{
		// Hook::event("EH_UPGRADE_RESEARCH", array(&$row, &$data));
		$points = round(($data["metal"] + $data["silicon"] + $data["hydrogen"]) * RES_TO_RESEARCH_POINTS, POINTS_PRECISION);

		$add_row = sqlSelectRow("research2user", "added", "", "buildingid = ".sqlVal($data["buildingid"])." AND userid = ".sqlVal($row["userid"]));
		$insert = $add_row ? false : true;

		if($insert)
		{
			sqlInsert("research2user", array(
				"buildingid" => $data["buildingid"],
				"userid" => $row["userid"],
				"level" => 1,
				"added" => 0));
		}
		else
		{
			// Update By Pk
			sqlUpdate("research2user", array(
				"level" => $data["level"]+$add_row["added"]
			), "buildingid = ".sqlVal($data["buildingid"])." AND userid = ".sqlVal($row["userid"]));
		}

		// Update By Pk
		sqlQuery("UPDATE ".PREFIX."user SET "
			. " points = points + ".sqlVal($points)
			. ", r_points = r_points + ".sqlVal($points)
			. ", r_count = r_count + 1"
            . ", ".updateDmPointsSetSql()
			. " WHERE userid = ".sqlVal($row["userid"]));

        NS::updateProfession($row['userid']);
		AchievementsService::processAchievements($row["userid"], $row["planetid"]);

		return $this;
	}

	protected function shipyard($row, $data)
	{
        $start_event_process_time = microtime(true);
        $row["data"] = &$data;
        if(!empty($data["new_format"]) && $data["quantity"] >= 1)
        {
            $data["quantity"] = max(1, floor($data["quantity"]));
            $data["duration"] = max(0.001, floatval($data["duration"]));
            $start_time = $row["start"]; // + $data["duration"];
            $end_time = time();
            $build_quantity = max(1, min($data["quantity"], round(($end_time - $start_time) / $data["duration"])));

			// logShipsChanging($data["buildingid"], $row["planetid"], 1, 1, $row["mode"] == EVENT_REPAIR ? "[repair]" : "[shipyard]");

			$points = $data["points"] * $build_quantity;
			if(sqlSelectField("unit2shipyard", "count(*)", "", "unitid = ".sqlVal($data["buildingid"])." AND planetid = ".sqlVal($row["planetid"])))
			{
				// Update By Pk
				sqlQuery("UPDATE ".PREFIX."unit2shipyard SET quantity = quantity + ".sqlVal($build_quantity)." WHERE planetid = ".sqlVal($row["planetid"])." AND unitid = ".sqlVal($data["buildingid"]));
			}
			else
			{
				sqlInsert("unit2shipyard", array("unitid" => $data["buildingid"], "planetid" => $row["planetid"], "quantity" => $build_quantity));
			}

			// $fpoints = intval($row["mode"] == EVENT_BUILD_FLEET || $data["buildingtype"] == UNIT_TYPE_FLEET);
			// Update By Pk
			sqlQuery("UPDATE ".PREFIX."user SET "
				. " points = points + ".sqlVal($points)
				. ", u_points = u_points + ".sqlVal($points)
				. ", u_count = u_count + ".  sqlVal($build_quantity)
                . ", ".updateDmPointsSetSql()
				. " WHERE userid = ".sqlVal($row["userid"]));

			// sqlUpdate("events", array("processed_quantity" => ++$processed_quantity), "eventid=".sqlVal($row["eventid"]));
			// Update By Pk
			sqlQuery("UPDATE ".PREFIX."events SET processed_quantity = case when processed_quantity is null then ".sqlVal($build_quantity)." else processed_quantity + ".sqlVal($build_quantity)." end "
				. " WHERE eventid=".sqlVal($row["eventid"]));

            $data["quantity"] -= $build_quantity;
            if($data["quantity"] >= 1)
            {
                if($data["new_format"] == 2)
                {
                    $basic_metal = $data["basic_metal"];
                    $basic_silicon = $data["basic_silicon"];
                    $basic_hydrogen = $data["basic_hydrogen"];
                    $basic_energy = $data["basic_energy"];
                }
                else
                {
                    $basic_metal = round(max(0, $data["metal"]) / $data["quantity"]);
                    $basic_silicon = round(max(0, $data["silicon"]) / $data["quantity"]);
                    $basic_hydrogen = round(max(0, $data["hydrogen"]) / $data["quantity"]);
                    $basic_energy = round(max(0, $data["energy"]) / $data["quantity"]);
                }

                if($basic_metal > 0 || $basic_silicon > 0 || $basic_hydrogen > 0 || $basic_energy > 0)
                {
                    if(isset($data["damaged"]))
                    {
                        $data["damaged"] = max(0, floor($data["damaged"]) - $build_quantity);
                    }

                    $data["metal"] -= $basic_metal * $build_quantity;
                    $data["silicon"] -= $basic_silicon * $build_quantity;
                    $data["hydrogen"] -= $basic_hydrogen * $build_quantity;
                    $data["energy"] -= $basic_energy * $build_quantity;

                    if($row["mode"] == EVENT_REPAIR)
                    {
                        $data["repair_usage"] -= $data["unit_fields"] * $build_quantity;
                    }
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
        AchievementsService::processAchievements($row["userid"], $row["planetid"]);
		return $this;
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
			$data["duration"] = max(0.001, floatval($data["duration"]));
			$start_time = $row["start"];
			$end_time = time();
			$build_quantity = max(1, min($data["quantity"], round(($end_time - $start_time) / $data["duration"])));

			NS::updateUserRes(array(
				"type" => RES_UPDATE_DISASSEMBLE,
				"event_mode" => $row["mode"],
				"reload_planet" => false,
				"update_production" => true,
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

					$data["paid"] = 1;

					$start_time = time();
					$end_time = $start_time + NS::getUnitsBuildTime($data);

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
					sqlUpdate("planet", array("temperature" => $temperature), "planetid=".sqlVal($galaxy_row["planetid"])." LIMIT 1");
					sqlUpdate("planet", array("temperature" => $temperature - mt_rand(15, 35)), "planetid=".sqlVal($row["planetid"])." LIMIT 1");
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

	protected function position($row, $data)
	{
		if(isset($data["pathstack"]) && count($data["pathstack"]) > 0 && $row["userid"] != $row["destuser"])
		{
			return $this->haltPosition($row);
		}

		$data["metal"] = max(0, $data["metal"]);
		$data["silicon"] = max(0, $data["silicon"]);
		$data["hydrogen"] = max(0, $data["hydrogen"]);

		// Hook::event("EH_POSITION", array(&$row, &$data));
		foreach($data["ships"] as $type => $ship)
		{
			if($ship["mode"] == UNIT_TYPE_ARTEFACT)
			{
				foreach($ship["art_ids"] as $art)
				{
					Artefact::detachFromFleet($art["artid"], $row["userid"], $row["destination"]);
				}
			}
			else
			{
				fixUnitDamagedVars($ship);

				logShipsChanging($ship["id"], $row["destination"], $ship["quantity"], 1, "[position]");

				$result1_count = sqlSelectField("unit2shipyard", "count(*)", "", "unitid = ".sqlVal($ship["id"])." AND planetid = ".sqlVal($row["destination"]));
				if($result1_count > 0)
				{
					// Update By Pk
					sqlQuery("UPDATE ".PREFIX."unit2shipyard SET "
						. addQuantitySetSql($ship)
						. " WHERE unitid = ".sqlVal($ship["id"])." AND planetid = ".sqlVal($row["destination"]));
				}
				else
				{
					Core::getQuery()->insert("unit2shipyard",
						array("unitid", "planetid", "quantity", "damaged", "shell_percent"),
						array($ship["id"], $row["destination"], $ship["quantity"], $ship["damaged"], $ship["shell_percent"]));
				}
			}
		}
		$data["destination"] = $row["destination"];
		$data["hydrogen"] += round($data["consumption"] / 2);

		NS::updateUserRes(array(
			"type" => RES_UPDATE_UNLOAD,
			"event_mode" => $row["mode"],
			"reload_planet" => false,
			"ownerid" => $row["eventid"],
			"userid" => $row["userid"],
			"planetid" => $row["destination"],
			"metal" => $data["metal"],
			"silicon" => $data["silicon"],
			"hydrogen" => $data["hydrogen"],
			// "max_metal" => NS::getPlanet()->getStorage("metal"),
			// "max_silicon" => NS::getPlanet()->getStorage("silicon"),
			// "max_hydrogen" => NS::getPlanet()->getStorage("hydrogen"),
			));
		/*
		sqlQuery("UPDATE ".PREFIX."planet SET "
		. " metal = metal + ".sqlVal($data["metal"])
		. ",silicon = silicon + ".sqlVal($data["silicon"])
		. ",hydrogen = hydrogen + ".sqlVal($data["hydrogen"])
		. ", last = ".sqlVal(time()+1)
		. " WHERE planetid = ".sqlVal($row["destination"]));
		*/

		AchievementsService::processAchievements($row["userid"], $row["destination"]);

		switch($row["mode"])
		{
		case EVENT_DELIVERY_UNITS:
			$msg_type = MSG_DELIVERY_UNITS;
			break;

		case EVENT_COLONIZE:
		case EVENT_COLONIZE_RANDOM_PLANET:
		case EVENT_COLONIZE_NEW_USER_PLANET:
			if(!count($data["ships"]) && !$data["metal"] && !$data["silicon"] && !$data["hydrogen"])
			{
				return $this;
			}
			$msg_type = MSG_POSITION_REPORT;
			break;

		case EVENT_STARGATE_JUMP:
			$msg_type = MSG_STARGATE_JUMP_REPORT;
			break;

		default:
			$msg_type = MSG_POSITION_REPORT;
		}

		new AutoMsg($msg_type, $row["userid"], time() /*$row["time"]*/, $data);

		return $this;
	}

	protected function transport($row, $data)
	{
		error_log("Transport event.");
		$data["metal"] 		= max(0, $data["metal"]);
		$data["silicon"] 	= max(0, $data["silicon"]);
		$data["hydrogen"] 	= max(0, $data["hydrogen"]);

		if ( $row["mode"] != EVENT_DELIVERY_ARTEFACTS )
		{
			// error_log("Updating ress, user {$row["destuser"]}, planetid {$row["destination"]}, event_mode {$row["mode"]}");
			NS::updateUserRes(array(
				"type" 			=> RES_UPDATE_UNLOAD,
				"event_mode" 	=> $row["mode"],
				"reload_planet" => false,
				"ownerid" 		=> $row["eventid"],
				"userid" 		=> $row["destuser"],
				"planetid" 		=> $row["destination"],
				"metal" 		=> $data["metal"],
				"silicon" 		=> $data["silicon"],
				"hydrogen" 		=> $data["hydrogen"],
			));
		}

		$data["targetplanet"] 	= $row["destplanet"];
		$data["targetuser"] 	= $row["destuser"];
		$data["startplanet"] 	= $row["planetname"];
		$data["startuser"] 		= $row["username"];
		$data["destination"] 	= $row["destination"];
		$data["planetid"] 		= $row["planetid"];

		if( $row["userid"] == $row["destuser"] )
		{
				// error_log("MSG_TRANSPORT_REPORT");
				new AutoMsg(MSG_TRANSPORT_REPORT, $row["userid"], time() /*$row["time"]*/, $data);
		}
		else
		{
			if ( $row["mode"] != EVENT_DELIVERY_ARTEFACTS )
			{
				if( ($data["metal"] || $data["silicon"] || $data["hydrogen"]) && $row["mode"] != EVENT_DELIVERY_RESOURSES )
				{
					Core::getQuery()->insert(
						"res_transfer",
						array("time", "userid", "senderid", "metal", "silicon", "hydrogen", "resum"),
						array($row["time"], $data["targetuser"], $row["userid"], $data["metal"], $data["silicon"], $data["hydrogen"], $data["metal"] + $data["silicon"] + $data["hydrogen"]));
				}
				if ( $row["mode"] == EVENT_DELIVERY_RESOURSES )
				{
					$msg_type = MSG_DELIVERY_RESOURSES;
				}
				else
				{
					$msg_type = MSG_TRANSPORT_REPORT_OTHER;
				}
				// error_log("Creating message row['mode'] = {$row["mode"]}, message_type = $msg_type");
			}
			else
			{
				foreach($data["ships"] as $type => $ship)
				{
					if( isset($ship['art_ids']) )
					{
						// todo: here something wrong, if $ship["art_ids"] is array then why is only one art stored: $data['artefact'] = $art['artid']
						foreach($ship["art_ids"] as $key => $art)
						{
							Artefact::onOwnerChange($art["artid"], $row["destuser"], $row["destination"]);
						}
						$data['artefact'] = $art['artid'];
						unset($data['ships'][$type]);
					}
				}
				$msg_type = MSG_TRANSPORT_REPORT_ARTEFACT;
			}
			new AutoMsg($msg_type, $row["userid"], time() /*$row["time"]*/, $data);
		}
		$data["metal"] = 0;
		$data["silicon"] = 0;
		$data["hydrogen"] = 0;
		$data["oldmode"] = $row["mode"];

		unset($data['artefact']);

		// error_log("sendBack	");
		$this->sendBack(time() + $data["time"], $row["destination"], $row["userid"], $row["planetid"], $data, $row["eventid"]);

		return $this;
	}

	protected function colonize($row, $data)
	{
		// Hook::event("EH_COLONIZE", array(&$row, &$data));

		unset($data["success"]);

		if( $row["mode"] == EVENT_COLONIZE_RANDOM_PLANET || $row["mode"] == EVENT_COLONIZE_NEW_USER_PLANET )
		{
			$is_colonize_random_planet_event = true;
			$planetid = false;
		}
		else
		{
			$is_colonize_random_planet_event = false;
			$data["galaxy"] = clampVal($data["galaxy"], 1, NUM_GALAXYS);
			$data["system"] = clampVal($data["system"], 1, NUM_SYSTEMS);
			$data["position"] = clampVal($data["position"], 1, MAX_NORMAL_PLANET_POSITION);
			$planetid = sqlSelectField("galaxy", "planetid", "", "galaxy = ".sqlVal($data["galaxy"])." AND system = ".sqlVal($data["system"])." AND position = ".sqlVal($data["position"])." AND planetid != '0'");;
		}
		if(!$planetid)
		{
			$is_art_used = false;
			$is_valid = false;
			$is_temp = false;
			$planets_number = sqlSelectField("planet", "count(*)", "", "userid = ".sqlVal($row["userid"])." AND ismoon = 0 AND destroy_eventid IS NULL");
			// debug_var($data, "[colonize] planets already: $planets_number");
			if( isset($data["ships"][ARTEFACT_PLANET_TEMP_CREATOR]) )
			{
				$temp_planets_number = sqlSelectField("planet", "count(*)", "", "userid = ".sqlVal($row["userid"])." AND ismoon = 0 AND destroy_eventid IS NOT NULL");
				if($temp_planets_number < TEMP_PLANETS_NUMBER)
				{
					$art = reset($data["ships"][ARTEFACT_PLANET_TEMP_CREATOR]["art_ids"]);
					$artid = $art["artid"];
					if( Artefact::activate($artid, $row["userid"], $row["planetid"]) )
					{
						$is_valid = true;
						$is_art_used = true;
						$is_temp = true;
						unset( $data["ships"][ARTEFACT_PLANET_TEMP_CREATOR] );
					}
				}
			}
			else if($planets_number < MAX_PLANETS) // it's always valid for new user: planets_number == 0
			{
				$is_valid = true;
			}
			else if( isset($data["ships"][ARTEFACT_PLANET_CREATOR])
						&& ( MAX_PLANETS + ADDITIONAL_ARTEFACT_PLANETS_NUMBER > $planets_number ) )
			{
				$art = reset($data["ships"][ARTEFACT_PLANET_CREATOR]["art_ids"]);
				$artid = $art["artid"];
				if( Artefact::activate($artid, $row["userid"], $row["planetid"]) )
				{
					$is_valid = true;
					$is_art_used = true;
					unset( $data["ships"][ARTEFACT_PLANET_CREATOR] );
				}
			}

			if($is_valid)
			{
				if( $planets_number == 0 )
				{
					$planet_percent = 99;
				}
				else
				{
					$planet_percent = $is_art_used ? ($is_temp ? 5 : 100) : 0;
				}
				if( $is_colonize_random_planet_event )
				{
					$colony = new PlanetCreator( $row["userid"], null, null, null, 0, $planet_percent );
					$colony_pos = $colony->getPosition();
					$data["galaxy"] = $colony_pos["galaxy"];
					$data["system"] = $colony_pos["system"];
					$data["position"] = $colony_pos["position"];
				}
				else
				{
					$colony = new PlanetCreator( $row["userid"], $data["galaxy"], $data["system"], $data["position"], 0, $planet_percent );
				}
				if(TEMP_MOON_ENABLED && $is_temp)
				{
					new PlanetCreator( $row["userid"], $data["galaxy"], $data["system"], $data["position"], 1, 0 );
				}

				// if( !isset($data["ships"][ARTEFACT_PLANET_CREATOR]) )
				if( isset($data["ships"][UNIT_COLONY_SHIP]) )
				{
					if( $planets_number > 0 )
					{
						$shipData = sqlSelectRow("construction", array("basic_metal", "basic_silicon", "basic_hydrogen"), "", "buildingid = ".UNIT_COLONY_SHIP);
						$points = round(($shipData["basic_metal"] + $shipData["basic_silicon"] + $shipData["basic_hydrogen"]) * RES_TO_UNIT_POINTS, POINTS_PRECISION);
						// Update By Pk
						sqlQuery("UPDATE ".PREFIX."user SET "
							. " points = GREATEST(0, points - ".sqlVal($points).") "
							. ", u_points = GREATEST(0, u_points - ".sqlVal($points).") "
							. ", u_count = GREATEST(0, u_count - 1)"
							. " WHERE userid = ".sqlVal($row["userid"]));
					}
					if(--$data["ships"][UNIT_COLONY_SHIP]["quantity"] <= 0)
					{
						unset($data["ships"][UNIT_COLONY_SHIP]);
					}
				}

				$row["destination"] = $colony->getPlanetId();

				if( $planets_number == 0 && $row["destination"] )
				{
					sqlUpdate("user", array(
							"hp" => $row["destination"],
							"curplanet" => $row["destination"],
						), "userid = ".sqlVal($row["userid"]));
				}

				if( $is_temp )
				{
					$destroy_eventid = $this->addEvent(
						EVENT_TEMP_PLANET_DISAPEAR,
						time() + TEMP_PLANET_LIFETIME,
						$row["destination"],
						$row["userid"],
						null,
						array(
							'planet_id'	=> $row["destination"],
							'user_id'	=> $row["userid"],
							'galaxy'	=> $data["galaxy"],
							'system'	=> $data["system"],
							'position'	=> $data["position"],
						)
					);
					sqlUpdate("planet", array(
						"destroy_eventid" => $destroy_eventid
					), "planetid = ".sqlVal($row["destination"]));
				}
			}
			else // if( $row["planetid"] )
			{
				$data["oldmode"] = $row["mode"]; // EVENT_COLONIZE;
				$this->sendBack(time() + $data["time"], $row["destination"], $row["userid"], $row["planetid"], $data, $row["eventid"]);
				$data["success"] = "empire";
			}
		}
		else
		{
			$data["oldmode"] = $row["mode"]; // EVENT_COLONIZE;
			$this->sendBack(time() + $data["time"], $row["destination"], $row["userid"], $row["planetid"], $data, $row["eventid"]);
			$data["success"] = "occupied";
		}
		new AutoMsg(MSG_COLONIZE_REPORT, $row["userid"], time() /*$row["time"]*/, $data);

		if(empty($data["success"]))
		{
			$this->position($row, $data);
		}

		if( $row["mode"] == EVENT_COLONIZE_NEW_USER_PLANET )
		{
            if(isset($GLOBALS["INITIAL_BUILDINGS"])){
                foreach($GLOBALS["INITIAL_BUILDINGS"] as $id => $level){
                    sqlInsert('building2planet', array(
                        'planetid' => $row["destination"],
                        'buildingid' => $id,
                        'level' => $level,
                    ));
                }
            }

            if(isset($GLOBALS["INITIAL_RESEARCHES"])){
                foreach($GLOBALS["INITIAL_RESEARCHES"] as $id => $level){
                    sqlInsert('research2user', array(
                        'userid' => $row["userid"],
                        'buildingid' => $id,
                        'level' => $level,
                    ));
                }
            }

            if(isset($GLOBALS["INITIAL_UNITS"])){
                sqlDelete('unit2shipyard', 'planetid='.sqlVal($row["destination"]));
                foreach($GLOBALS["INITIAL_UNITS"] as $id => $quantity){
                    sqlInsert('unit2shipyard', array(
                        'planetid' => $row["destination"],
                        'unitid' => $id,
                        'quantity' => $quantity,
                    ));
                }
            }

            PointRenewer::updateUserPoints($row["userid"]);

            if(ALIEN_ENABLED){
                AlienAI::generateAttack($row["userid"], $row["destination"], NEW_USER_ALIEN_ATTACK_FLYTIME, array(UNIT_A_PALADIN => 1));
            }
		}

		return $this;
	}

	protected function recycling($row, $data)
	{
		$result = sqlSelect(
			"galaxy",
			array("metal", "silicon"),
			"",
			"galaxy = ".sqlVal($data["galaxy"])." AND system = ".sqlVal($data["system"])." AND position = ".sqlVal($data["position"])
		);
		if($_row = sqlFetch($result))
		{
			sqlEnd($result);

			$capacity = (int)$data["capacity"];

			$data["debrismetal"] = (int)$_row["metal"];
			$data["debrissilicon"] = (int)$_row["silicon"];
			$data["debrishydrogen"] = 0; // (int)$_row["hydrogen"];

			recycleDebris($data);

			if($_row["silicon"] != 0 || $_row["metal"] != 0)
			{
				// Update By Pk
				Core::getQuery()->update("galaxy",
					array("metal", "silicon"),
					array($data["debrismetal"] - $data["recycledmetal"], $data["debrissilicon"] - $data["recycledsilicon"]),
					"galaxy = ".sqlVal($data["galaxy"])." AND system = ".sqlVal($data["system"])." AND position = ".sqlVal($data["position"]));
			}
		}
		$data["oldmode"] = EVENT_RECYCLING;
		$data["destination"] = $row["destination"];
		new AutoMsg(MSG_RECYCLING_REPORT, $row["userid"], time() /*$row["time"]*/, $data);
		$this->sendBack(time() + $data["time"], $row["destination"], $row["userid"], $row["planetid"], $data, $row["eventid"]);
		return $this;
	}

	protected function spy($row, $data)
	{
		// Hook::event("EH_ESPIONAGE", array(&$row, &$data));
		$espReport = new EspionageReport($row["destination"], $row["userid"], $row["destuser"], $row["destname"], $data["ships"][UNIT_ESPIONAGE_SENSOR]["quantity"]);

		$row0 = sqlSelectRow("planet p", array("u.userid", "u.username", "p.planetname"), "LEFT JOIN ".PREFIX."user u ON (u.userid = p.userid)", "p.planetid = ".sqlVal($row["planetid"]));

		$data["destinationplanet"] = $espReport->getPlanetname();
		$data["suser"] = $row0["username"];
		$data["planetname"] = $row0["planetname"];
		$data["defending_chance"] = $espReport->getChance();
		$data["probes_lost"] = $espReport->getProbesLost();
		$data["destination"] = $row["destination"];
		$data["planetid"] = $row["planetid"];

		new AutoMsg(MSG_ESPIONAGE_COMMITTED, $espReport->getTargetUserId(), time() /*$row["time"]*/, $data);

		if($espReport->getProbesLost())
		{
			$points = 0; $u_count = 0;
			foreach($data["ships"] as $key => $ship)
			{
				$shipData = sqlSelectRow("construction", array("basic_metal", "basic_silicon", "basic_hydrogen"), "", "buildingid = ".sqlVal($key));

				$points += round(($shipData["basic_metal"] + $shipData["basic_silicon"] + $shipData["basic_hydrogen"]) * RES_TO_UNIT_POINTS * $ship["quantity"], POINTS_PRECISION);
				$u_count += $ship["quantity"];

				$tfMetal = $shipData["basic_metal"] * $ship["quantity"] * FLEET_BULK_INTO_DEBRIS;
				$tfSilicon = $shipData["basic_silicon"] * $ship["quantity"] * FLEET_BULK_INTO_DEBRIS;
			}
			// Update By Pk
			sqlQuery("UPDATE ".PREFIX."galaxy SET metal = metal + ".sqlVal($tfMetal).", silicon = silicon + ".sqlVal($tfSilicon)." WHERE planetid = ".sqlVal($row["destination"])." OR moonid = ".sqlVal($row["destination"]));
			// Update By Pk
			sqlQuery("UPDATE ".PREFIX."user SET "
				. " points = points - ".sqlVal($points)
				. ", u_points = u_points - ".sqlVal($points)
				. ", u_count = u_count - ".sqlVal($u_count)
				. " WHERE userid = ".sqlVal($row["userid"]));
		}
		else
		{
			$data["nomessage"] = true;
			$data["oldmode"] = EVENT_SPY;
			$this->sendBack(time() + $data["time"], $row["destination"], $row["userid"], $row["planetid"], $data, $row["eventid"]);
		}
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

		$Assault->startAssault($data["galaxy"], $data["system"], $data["position"], array("mode" => $row["mode"]));

		Core::getQuery()->update("events", array("processed", "processed_mode", "processed_time"),
			array(EVENT_PROCESSED_OK, EVENT_ALLIANCE_ATTACK_ADDITIONAL, time()),
			"parent_eventid = ".sqlVal($parent_eventid)." AND mode = ".EVENT_ALLIANCE_ATTACK_ADDITIONAL." AND processed_time=".sqlVal($processed_time)
				. ' ORDER BY eventid');

		Core::getQuery()->delete("formation_invitation", "eventid = ".sqlVal($parent_eventid));
		Core::getQuery()->delete("attack_formation", "eventid = ".sqlVal($parent_eventid));
		return $this;
	}

	protected function halt($row, $data)
	{
		// Hook::event("EH_HALT", array(&$row, &$data));
		$time = time() /* $row["time"] */ + max(1, $data["duration"]);
		$this->addEvent(EVENT_HOLDING, $time, $row["planetid"], $row["userid"], $row["destination"], $data, null, $row["time"],
			$row["eventid"] // $parent_eventid
			);
		return $this;
	}

	protected function holding($row, $data)
	{
		// Hook::event("EH_HOLDING", array(&$row, &$data));
		if( empty($data["consumption"]) && !empty($data["back_consumption"]) && $row["destination"] != $row["planetid"] )
		{
			// Update By Pk
			sqlUpdate("events", array(
				"start" => time(),
				"time" => time() + 60*60*1,
				// "data" => serialize($data),
				"prev_rc" => null,
				"processed" => EVENT_PROCESSED_WAIT,
				"processed_mode" => $row["mode"],
				"processed_time" => time(),
				), "eventid=".sqlVal($row["eventid"]));
		}
		else
		{
			$this->sendBack(time() + $data["time"], $row["destination"], $row["userid"], $row["planetid"], $data, $row["eventid"]);
		}
		return $this;
	}

	protected function moonDestruction($row, $data)
	{
		// Hook::event("EH_MOON_DESTRUCTION", array(&$row, &$data));
		return;
	}

	protected function expedition($row, $data)
	{
		require_once(APP_ROOT_DIR."game/Expedition.class.php");
		new Expedition($this, $row);
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

	protected function fReturn($row, $data)
	{
		// Из ExtEventHandler: попытка перенаправления возврата при незанятой destination
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
					"mode" 				=> $row["mode"],
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

		if(isset($data["pathstack"]) && count($data["pathstack"]) > 0)
		{
			return $this->haltReturn($row);
		}

		if ( $data["oldmode"] == EVENT_DELIVERY_RESOURSES && !$data["sold"] )
		{
			// Update By Pk
			Core::getQuery()->update("exchange_lots", "status", ESTATUS_OK, "lid = ".sqlVal($data["lid"]));
			return $this;
		}
		foreach($data["ships"] as $type => $ship)
		{
			if($ship["mode"] == UNIT_TYPE_ARTEFACT)
			{
				foreach($ship["art_ids"] as $art)
				{
					Artefact::detachFromFleet($art["artid"], $row["userid"], $row["destination"]);
				}
			}
			else
			{
				fixUnitDamagedVars($ship);
				logShipsChanging($ship["id"], $row["destination"], $ship["quantity"], 1, "[fReturn]");
				$result1_count = sqlSelectField(
					"unit2shipyard",
					"count(*)",
					"",
					"unitid = ".sqlVal($ship["id"])
						. " AND planetid = ".sqlVal($row["destination"])
				);
				if( $result1_count > 0 )
				{
					// Update By Pk
					sqlQuery("UPDATE " . PREFIX . "unit2shipyard SET "
						. addQuantitySetSql($ship)
						. " WHERE unitid = " . sqlVal($ship["id"]) . " AND planetid = " . sqlVal($row["destination"])
					);
				}
				else
				{
					Core::getQuery()->insert("unit2shipyard",
						array("unitid", "planetid", "quantity", "damaged", "shell_percent"),
						array($ship["id"], $row["destination"], $ship["quantity"], $ship["damaged"], $ship["shell_percent"]));
				}
			}
		}
		$data["destination"]	= $row["destination"];
		$data["planetid"]		= $row["planetid"];
		NS::updateUserRes(array(
			"type"			=> RES_UPDATE_UNLOAD,
			"event_mode"	=> $row["mode"],
			"reload_planet" => false,
			"ownerid"		=> $row["eventid"],
			"userid"		=> $row["userid"],
			"planetid"		=> $row["destination"],
			"metal"			=> $data["metal"],
			"silicon"		=> $data["silicon"],
			"hydrogen"		=> $data["hydrogen"] + ( isset($data["ret_consumption"]) ? $data["ret_consumption"] : 0 ),
		));

		if(!isset($data["nomessage"]))
		{
			new AutoMsg(MSG_RETURN_REPORT, $row["userid"], time() /*$row["time"]*/, $data);
		}

		return $this;
	}

	protected function artefactExpire($row, $data)
	{
		Artefact::onExpireEvent($row); // ["artid"], $row["userid"], $row["planetid"]); // $data["artid"],$data["typeid"],$row["userid"],$row["planetid"]);
		return $this;
	}

	protected function artefactDisappear($row, $data)
	{
		Artefact::onDisappearEvent($row); // ["artid"], $row["userid"], $row["planetid"]); // $data["artid"],$data["typeid"],$row["userid"],$row["planetid"]);
		return $this;
	}

	protected function artefactDelay($row, $data)
	{
		Artefact::onDelayEvent($row); // ["artid"], $row["userid"], $row["planetid"]); //
		// sqlUpdate("artefact2user", array("active" => 1), "artid=".sqlVal($data["artid"]));
		return $this;
	}

	protected function exchangeExpire($row, $data)
	{
		$id = ( ( isset($data['exchid']) && !empty($data['exchid']) )
		 	? ($data['exchid'])
		 	: ($data['lot_id'])
		);
		$result = Exchange::sendBackLot($id);
		/* if( !$result )
		{
			error_log('Lot not found or this is no first run. LID - ' . $id . ' Event ID - ' . $row['eventid']);
			throw new Exception("Can't send lot back $id");
		} */
		return $this;
	}

	protected function destroyPlanet($row, $data)
	{
		$row = sqlSelectRow("galaxy", "*", "", "planetid = ".sqlVal($data['planet_id']));
		if( !empty($row["planetid"]) )
		{
			NS::deletePlanet($row["planetid"], $row["userid"], $row["ismoon"], true);

			/*
			// $moon_id = sqlSelectField("galaxy", "moonid", "", "planetid = ".sqlVal($data['planet_id']));
			if($row["moonid"])
			{
				// Update By Pk
				sqlUpdate('planet', array(
					'userid' => null
				), 'planetid = ' . sqlVal($row["moonid"]) );
			}

			// Update By Pk
			sqlUpdate('planet', array(
				'userid' => null
			), 'planetid = ' . sqlVal($data['planet_id']) );

			// Update By Pk
			sqlUpdate('galaxy', array(
				'destroyed' => 1
			), 'planetid = ' . sqlVal($data['planet_id']) );
			*/

			new AutoMsg(
				// $row["destination"] ? MSG_UFO_PLANET_DIE : MSG_TEMP_PLANET_DIE,
				$row["userid"] ? MSG_TEMP_PLANET_DIE : MSG_UFO_PLANET_DIE,
				$data['user_id'],
				time(),
				array(
					'planet_id' => $data['planet_id'],
					'galaxy'	=> $data['galaxy'],
					'system'	=> $data['system'],
					'position'	=> $data['position'],
				)
			);
		}
		return $this;
	}

	protected function runSimAssault($row, $data)
	{
		$database = array();
		require(APP_ROOT_DIR."config.inc.php");

		$temp_array = array(
			$database["host"],
			$database["user"],
			$database["userpw"],
			$database["databasename"],
			$database["tableprefix"]."sim_",
			$data["assaultid"],
		);
		$s = '/usr/bin/java -cp '.APP_ROOT_DIR.'game/'.SIMULATOR_ASSAULT_JAR.' assault.Assault "' . implode('" "', $temp_array) . '"';
		exec( $s );

		// Update By Pk
		sqlQuery("UPDATE ".PREFIX."events SET processed_quantity = case when processed_quantity is null then 1 else processed_quantity + 1 end "
			. " WHERE eventid=".sqlVal($row["eventid"]));

		return $this;
	}

	// === Из ExtEventHandler: новые обработчики событий ===

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

					$data["metal"] -= $basic_metal;
					$data["silicon"] -= $basic_silicon;
					$data["hydrogen"] -= $basic_hydrogen;
					$data["energy"] -= $basic_energy;

					$data["repair_usage"] -= $data["unit_fields"];

					$data["paid"] = 1;

					$end_time = $start_time + $data["duration"];

					if($end_time <= time() && $processed_quantity < 100)
					{
						$row["start"] = $start_time;
						$row["time"] = $end_time;
						continue;
					}

					$cur_time = time();
                    $start_time = $cur_time;
                    $end_time = $start_time + $data["duration"] * clampVal(round($data["quantity"] / 100), 1, 100);
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
			null,
			$event["eventid"]
			);

		new AutoMsg(MSG_POSITION_REPORT, $event["userid"], time(), $data);

		return $this;
	}

	protected function haltReturn($event)
	{
		if(isset($event["data"]["pathstack"]) && count($event["data"]["pathstack"]) > 0)
		{
			$fleet_params = NS::calcFleetParams($event["data"]["ships"]);

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
				"consumption" => $stack_data["consumption"],
				"fleet_size" => $fleet_params["fleet_size"],
				"control_times" => $event["data"]["control_times"],
				"duration" => 60*60*1,
				"pathstack" => $event["data"]["pathstack"],
			);

			$data["capacity"] = $fleet_params["capacity"]
									- $data["metal"]
									- $data["silicon"]
									- $data["hydrogen"];

			// fleet size consumption
			$time = $data["time"];
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
				null,
				$event["eventid"]
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