<?php
/**
 * Class for handling the artefacts.
 *
 * Oxsar http://oxsar.ru
 *
 *
 */

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class Artefact
{
	/**
	 * Check if any artefact of a given type is active
	 */
	public static function getActiveCount($type_id, $user_id, $excludeid = 0, $use_cached_result = true)
	{
		$count = sqlSelectField("artefact2user", "count(*)", "",
			"deleted=0 AND active=1 AND typeid=".sqlVal($type_id)." AND userid=".sqlVal($user_id).($excludeid ? " AND artid !=".sqlVal($excludeid) : ""));
		// debug_var($count, "[getActiveCount] $type_id, $user_id, $excludeid");
		return $count;
	}

	public static function getMerchantMark($user_id = null)
	{
		static $cache = array();
		if(is_null($user_id))
		{
			$user_id = NS::getUser()->get("userid");
		}
		if(!isset($cache[$user_id]))
		{
			$cache[$user_id] = self::getActiveCount(ARTEFACT_MERCHANTS_MARK, $user_id);
		}
		return $cache[$user_id];
	}

	protected static function getMaxActiveCount($type_id)
	{
		$count = sqlSelectField("artefact_datasheet", "max_active", "", "typeid=".sqlVal($type_id));
		// debug_var($count, "[getMaxActiveCount] $id");
		return $count;
	}

	/**
	 * Check artefact's requirements
	 */
	public static function checkRequirements($type_id, $user_id = null, $planet_id = null, $art_id = null)
	{
		if(!Artefact::hasFreeSlots($user_id, $planet_id))
		{
			return false;
		}

		if(is_null($user_id))
		{
			$user_id = NS::getUser()->get("userid");
		}
		if(is_null($planet_id))
		{
			$planet_id = NS::getUser()->get("curplanet");
		}

		$max_active = sqlSelectField(
			'research2user',
			'level',
			'',
			' buildingid = ' . UNIT_ARTEFACTS_TECH . ' AND userid = ' . sqlVal($user_id)
		);

		$curr_active = sqlSelectField(
			'artefact2user',
			'COUNT(*) AS curr_active',
			'',
			' active = 1'.
			' AND deleted = 0 '.
			' AND expire_eventid != 0 '.
			' AND userid = '. sqlVal($user_id)
		);

		if( empty($max_active) || $curr_active >= $max_active )
		{
			return false;
		}

		if(!NS::checkRequirements($type_id, $user_id, $planet_id))
		{
			return false;
		}

		$max_active = self::getMaxActiveCount($type_id);
		if($max_active > 0 && self::getActiveCount($type_id, $user_id) >= $max_active)
		{
			return false;
		}

		switch($type_id)
		{
			case ARTEFACT_MERCHANTS_MARK:
			case ARTEFACT_CATALYST:
			case ARTEFACT_POWER_GENERATOR:
			case ARTEFACT_ATOMIC_DENSIFIER:
				break;

			case ARTEFACT_PACKING_BUILDING:
				return self::checkBuildingBusy($type_id, $planet_id);

			case ARTEFACT_PACKING_RESEARCH:
				return self::checkResearchBusy($type_id, $user_id);

			case ARTEFACT_PACKED_BUILDING:
				if ( !empty($art_id) )
				{
					$curr_art_data = sqlSelectRow('artefact2user AS a', 'a.*, c.mode', 'LEFT JOIN ' . PREFIX . 'construction AS c ON c.buildingid = a.construction_id', "artid = ".sqlVal($art_id));
					if(empty($curr_art_data))
					{
						return false;
					}

					$art_id = $curr_art_data["artid"];
					$level= $curr_art_data["level"];
					$construction_id = $curr_art_data["construction_id"];
					$isMoon = NS::getPlanet()->getData('ismoon');
					$mode = $curr_art_data['mode'];
					if
					(
						( $isMoon && $mode != UNIT_TYPE_MOON_CONSTRUCTION ) ||
						( !$isMoon && $mode != UNIT_TYPE_CONSTRUCTION )
					)
					{
						return false;
					}
					if ( self::checkBuildingBusy($construction_id, $planet_id) )
					{
						$add_row = sqlSelectRow("building2planet", "*", "", "planetid = ".sqlVal($planet_id)." AND buildingid = ".sqlVal($construction_id));
						$insert = $add_row ? false : true;
						if( $insert ) // && $level == 1 )
						{
							$add_row = array( "level" => 0 );
							// return 1;
						}
						// elseif( !$insert )
						{
							$new_level = self::getUpgradedLevel($add_row, $level);
							return $new_level > 0 ? $new_level : false;
						}
						/* else
						{
							return false;
						} */
					}
					else
					{
						return false;
					}
				}
				else
				{
					return false;
				}
				break;

			case ARTEFACT_PACKED_RESEARCH:
				if ( !empty($art_id) )
				{
					// t('dsfsdfsdfsdf');
					$curr_art_data = sqlSelectRow('artefact2user', '*', '', "artid = ".sqlVal($art_id));
					if(empty($curr_art_data))
					{
						return false;
					}
					$art_id = $curr_art_data["artid"];
					$level = $curr_art_data["level"];
					$construction_id = $curr_art_data["construction_id"];
					if ( self::checkResearchBusy($construction_id, $user_id) )
					{
						$add_row = sqlSelectRow("research2user", "*", "", "userid = ".sqlVal($user_id)." AND buildingid = ".sqlVal($construction_id));
						$insert = $add_row ? false : true;
						// t('');
						if( $insert ) // && $level == 1 )
						{
							$add_row = array( "level" => 0 );
							// return 1;
						}
						// elseif( !$insert )
						{
							$new_level = self::getUpgradedLevel($add_row, $level);
							return $new_level > 0 ? $new_level : false;
						}
						/* else
						{
							return false;
						} */
					}
					else
					{
						return false;
					}
				}
				else
				{
					return false;
				}
				break;
				
			case ARTEFACT_NANOBOT_REPAIR_SYSTEM:
				if(!empty($art_id)){
					$row = sqlSelectRow("unit2shipyard", "*", "", "planetid = ".sqlVal($planet_id)." AND damaged>0");
					return $row ? true : false;
				}else{
					return false;
				}
				break;
		}

		return true;
	}

	public static function addEvent($mode, $time, $art_id, $user_id, $planet_id, $start_time = null)
	{
		// TODO: Does it occur? Probably NS::getEH() has to be refactored. ExtEventHandler has to be used if exists
		$EH = NS::getEH();
		if( !is_object($EH) )
		{
			$EH = new EventHandler();
		}
		return $EH->addEvent($mode, $time, $planet_id, $user_id, null, null, null, $start_time, null, $art_id);
	}

	/**
	 *
	 * Creates artefact.
	 * @param int Type of artefact.
	 * @param int ID of user
	 * @param int ID of planet on which we create artefact. If it is created in space set $params['flying'] not empty.
	 * @param array Additional parametrs of artefact creation.
	 */
	public static function appear($type_id, $user_id, $planet_id, $params) // $buyed = false, $assaultid = 0)
	{
		$row = sqlQueryRow("SELECT buyable, auto_active, movable, `unique`, usable, trophy_chance, delay, use_times "
			. ", use_duration, lifetime, effect_type, max_active FROM ".PREFIX."artefact_datasheet "
			. " WHERE typeid=".sqlVal($type_id));
		if($row && (!isset($params["times_left"]) || $params["times_left"] > 0))
		{
			$art_id = sqlInsert("artefact2user", array(
				"typeid" => $type_id,
				"userid" => $user_id,
				"planetid" => empty($params["flying"]) ? $planet_id : 0,
				"active" => 0,
				"times_left" => isset($params["times_left"]) ? $params["times_left"] : $row["use_times"],
				"delay_eventid" => 0,
				"expire_eventid" => 0,
				"lifetime_eventid" => 0,
				"level" => isset($params["level"]) ? $params["level"] : 0,
				"construction_id" => isset($params["construction_id"]) ? $params["construction_id"] : 0,
                "bought" => empty($params["bought"]) ? 0 : 1,
                // "artid" => $art_id,
			));

			$delay_eventid = 0;
			$lifetime_eventid = 0;

			if(isset($params["lifetime"]))
			{
				$row["lifetime"] = $params["lifetime"];
			}
			if($row["lifetime"] > 0)
			{
				$lifetime_eventid = self::addEvent(EVENT_ARTEFACT_DISAPPEAR, max(10, $row["lifetime"])+time(), $art_id, $user_id, $planet_id);
			}

			if(empty($params["flying"]))
			{
				if(isset($params["delay"]))
				{
					$row["delay"] = $params["delay"];
				}
				if($row["delay"] > 0)
				{
					$delay_eventid = self::addEvent(EVENT_ARTEFACT_DELAY, max(10, $row["delay"])+time(), $art_id, $user_id, $planet_id);
				}
			}

			if($lifetime_eventid || $delay_eventid)
			{
				// Updating by PK
				sqlUpdate("artefact2user", array(
					"lifetime_eventid" => $lifetime_eventid,
					"delay_eventid" => $delay_eventid,
				), "artid=".sqlVal($art_id));
			}

			if($row["unique"])
			{
				sqlInsert("artefact_history", array(
					"typeid" => $type_id,
					"userid" => $user_id,
					"assaultid" => isset($params["assaultid"]) ? $params["assaultid"] : 0,
					"time" => time()));
			}

			if( empty($params["flying"]) && $row["auto_active"] )
			{
				self::activate($art_id, $user_id, $planet_id, false);
			}

			return $art_id;
		}
		return false;
	}

	/**
	 * Call on expiration of the artefact's lifetime
	 */
	public static function onDisappearEvent($event) // $art_id, $user_id, $planet_id, $reason = "lifetime")
	{
		$row = sqlQueryRow("SELECT buyable, auto_active, movable, `unique`, usable, trophy_chance, delay, use_times "
			. ", use_duration, lifetime, effect_type, max_active, construction_id, `level`"
			. ", artid, a2u.typeid, userid, planetid, active, times_left, delay_eventid, expire_eventid, lifetime_eventid "
			. " FROM ".PREFIX."artefact2user a2u "
			. " INNER JOIN ".PREFIX."artefact_datasheet ads ON a2u.typeid=ads.typeid "
			. " WHERE a2u.deleted=0 AND a2u.lifetime_eventid=".sqlVal($event["eventid"])
			. "	 AND (a2u.userid=".sqlVal($event["userid"])
			. "	 OR a2u.artid=".sqlVal($event["artid"]). ')'
		);
		if($row)
		{
			$art_id	 = $row["artid"];
			$type_id = $row["typeid"];

			if($row["planetid"] == 0)
			{
			}
			else if($row["active"])
			{
				self::updateEffect($row, false);
			}

				// Updating by PK
			sqlUpdate("artefact2user", array(
				"active" => 0,
				"deleted" => time(),
				"reason" => isset($event["new_userid"]) ? "transfer to ".$event["new_userid"] : "lifetime"
				), "artid=".sqlVal($art_id));

				if ( !isset($event['called_from_participants']) )
				{
					new AutoMsg(MSG_ARTEFACT, $row['userid'], time(), array('art_id' => $row['artid'], 'mode' => 'MSG_DISAPEAR_ARTEFACT'));
				}
				else
				{
					new AutoMsg(MSG_ARTEFACT, $row['userid'], time(), array('art_id' => $row['artid'], 'mode' => 'MSG_CAPTURE_ARTEFACT'));
				}

				return true;
		}
		return false;
	}

	public static function setTransportID( $id, $event_id )
	{
		if( sqlSelectRow('artefact2user', '*', '', 'artid=' . sqlVal($id)) )
		{
			// Updating by PK
			sqlUpdate('artefact2user', array('transport_eventid' => $event_id), 'artid=' . sqlVal($id));
			return true;
		}
		return false;
	}

	/**
	 *
	 * Call on artefact changing owner
	 * @param int $id Artefact ID
	 * @param int $user New User ID
	 * @param int $user New Planet ID
	 */
	public static function onOwnerChange( $id, $user, $planet = 0 )
	{
		$row = sqlQueryRow("SELECT buyable, auto_active, movable, `unique`, usable, trophy_chance, delay, use_times "
		. ", use_duration, lifetime, effect_type, max_active, construction_id, `level`"
		. ", artid, a2u.typeid, userid, planetid, active, times_left, delay_eventid, expire_eventid, lifetime_eventid "
		. " FROM ".PREFIX."artefact2user a2u "
		. " INNER JOIN ".PREFIX."artefact_datasheet ads ON a2u.typeid=ads.typeid "
		. " WHERE a2u.deleted = 0 " // AND a2u.lifetime_eventid != 0" - will be used later, lifetime_eventid is not set for all current artefacts
		. " AND a2u.artid=".sqlVal($id));
		if($row)
		{
				// Updating by PK
			$temp = sqlUpdate(
				'artefact2user',
				array(
					'userid'			=> $user,
					'planetid'			=> $planet,
					'delay_eventid'		=> 0, // is it OK?
					'expire_eventid'	=> 0, // is it OK?
					'transport_eventid' => 0, // is it OK?
					'lot_id'			=> 0,
					'active'			=> 0,
				),
				' artid = ' . sqlVal($id)
			);
			if( !$temp )
			{
				return false;
			}
			if( $row["active"] )
			{
				self::updateEffect( $row, false );
			}
			if ( $row['delay_eventid'] != 0 )
			{
				NS::getEH()->removeEvent($row['delay_eventid'], 'owner changed to ' . $user);
			}
			if ( $row['expire_eventid'] != 0 )
			{
				NS::getEH()->removeEvent($row['expire_eventid'], 'owner changed to ' . $user);
			}
			return true;
		}
		return false;
	}

	/**
	 *
	 * Supposedly called when delay time runs out
	 * @param aray $event. Params.
	 */
	public static function onDelayEvent($event)
	{
		$row = sqlQueryRow("SELECT buyable, auto_active, movable, `unique`, usable, trophy_chance, delay, use_times "
		. ", use_duration, lifetime, effect_type, max_active, construction_id, `level`"
		. ", artid, a2u.typeid, userid, planetid, active, times_left, delay_eventid, expire_eventid, lifetime_eventid "
		. " FROM ".PREFIX."artefact2user a2u "
		. " INNER JOIN ".PREFIX."artefact_datasheet ads ON a2u.typeid=ads.typeid "
		. " WHERE a2u.deleted=0 AND a2u.delay_eventid=".sqlVal($event["eventid"])
//			." AND a2u.userid=".sqlVal($event["userid"])
//			." AND a2u.planetid=".sqlVal($event["planetid"])
			. "	 AND (a2u.userid=".sqlVal($event["userid"])
			. "	 OR a2u.artid=".sqlVal($event["artid"]). ')'
		);
		if($row)
		{
			$art_id = $row["artid"];
			$type_id = $row["typeid"];
			$user_id = $row["userid"];
			$planet_id = $row["planetid"];

			// Artefact::updateEffect($row, "delay_finished");

			if(sqlSelectField("planet", "count(*)", "", "planetid=".sqlVal($planet_id)) == 0)
			{
				// Updating by PK
				sqlUpdate("artefact2user", array(
					"active" => 0,
					"deleted" => time(),
					"reason" => "planet is not found, onDelayEvent"
					), "artid=".sqlVal($art_id));

				return true;
			}

				// Updating by PK
			sqlUpdate("artefact2user", array(
			// "active" => 0,
				"delay_eventid" => 0,
			), "artid=".sqlVal($event["artid"]));
			new AutoMsg(MSG_ARTEFACT, $row['userid'], time(), array('art_id' => $row['artid'], 'mode' => 'MSG_DELAY_ARTEFACT'));
			return true;
		}
		return false;
	}

	/**
	 * Call on expiration of the artefact's effect
	 */
	public static function onExpireEvent($event) // $art_id, $user_id, $planet_id)
	{
		$row = sqlQueryRow("SELECT buyable, auto_active, movable, `unique`, usable, trophy_chance, delay, use_times "
			. ", use_duration, lifetime, effect_type, max_active, construction_id, `level`"
			. ", artid, a2u.typeid, userid, planetid, active, times_left, delay_eventid, expire_eventid, lifetime_eventid "
			. " FROM ".PREFIX."artefact2user a2u "
			. " INNER JOIN ".PREFIX."artefact_datasheet ads ON a2u.typeid=ads.typeid "
			. " WHERE a2u.deleted=0 AND a2u.expire_eventid=".sqlVal($event["eventid"])
//				." AND a2u.userid=".sqlVal($event["userid"]).
//				." AND a2u.planetid=".sqlVal($event["planetid"]).
//				." AND a2u.artid=".sqlVal($event["artid"])
			. "	 AND (a2u.userid=".sqlVal($event["userid"])
			. "	 OR a2u.artid=".sqlVal($event["artid"]). ')'
		);
		if($row)
		{
			$art_id = $row["artid"];
			$type_id = $row["typeid"];
			$user_id = $row["userid"];
			$planet_id = $row["planetid"];

			if($row["active"])
			{
				Artefact::updateEffect($row, false);
			}

			if(sqlSelectField("planet", "count(*)", "", "planetid=".sqlVal($planet_id)) == 0)
			{
				// Updating by PK
				sqlUpdate("artefact2user", array(
					"active" => 0,
					"deleted" => time(),
					"reason" => "planet is not found, onExpireEvent"
					), "artid=".sqlVal($art_id));

				return true;
			}
			else if($row["times_left"] > 0)
			{
				$msg_mode = 'MSG_EXPIRE_ARTEFACT';
				$delay_eventid = 0;
				if($row["delay"] > 0)
				{
					$delay_eventid = self::addEvent(EVENT_ARTEFACT_DELAY, max(10, $row["delay"]) + time(), $art_id, $user_id, $planet_id);
				}
				// Updating by PK
				sqlUpdate("artefact2user", array(
					"active" => 0,
					"expire_eventid" => 0,
					"delay_eventid" => $delay_eventid,
				), "artid=".sqlVal($art_id));
			}
			else
			{
				// Updating by PK
				$msg_mode = 'MSG_DISAPEAR_ARTEFACT_USED';
				sqlUpdate("artefact2user", array(
					"active" => 0,
					"deleted" => time(),
					"reason" => "times_left, expired"
					), "artid=".sqlVal($art_id));
			}
			new AutoMsg(MSG_ARTEFACT, $row['userid'], time(), array('art_id' => $row['artid'], 'mode' => $msg_mode));

			return true;
		}
		return false;
	}

	/**
	 * Call to activate the artefact
	 */
	public static function activate($art_id, $user_id, $planet_id, $user_action = true, $params = array() )
	{
		$row = sqlQueryRow("SELECT buyable, auto_active, movable, `unique`, usable, trophy_chance, delay, use_times "
		. ", use_duration, lifetime, effect_type, max_active, construction_id, `level`"
		. ", artid, a2u.typeid, userid, planetid, active, times_left, delay_eventid, expire_eventid, lifetime_eventid "
		. " FROM ".PREFIX."artefact2user a2u "
		. " INNER JOIN ".PREFIX."artefact_datasheet ads ON a2u.typeid=ads.typeid "
		. " WHERE a2u.deleted=0 AND a2u.userid=".sqlVal($user_id) // ." AND a2u.planetid=".sqlVal($planet_id)
		. " AND a2u.artid=".sqlVal($art_id)." AND active=0 AND expire_eventid=0 AND times_left>0"
		. ($user_action ? " AND delay_eventid=0 AND auto_active=0 " : "")
		// . " AND effect_type in (".sqlArray(ARTEFACT_EFFECT_TYPE_PLANET, ARTEFACT_EFFECT_TYPE_EMPIRE, ARTEFACT_EFFECT_TYPE_AUTO).")"
		);
		if($row)
		{
			// if(NS::getUser()->get("userid") == 2){ debug_var($row, "[activate]"); exit; }

			$art_id = $row["artid"];
			$type_id = $row["typeid"];
			$user_id = $row["userid"];
			$planet_id = $row["planetid"];

			switch($type_id)
			{
			case ARTEFACT_PACKING_BUILDING:
			case ARTEFACT_PACKING_RESEARCH:
				if ( empty($params['level']) || empty($params['construction_id']) )
				{
					return false;
				};
				$row['level'] = $params['level'];
				$row['construction_id'] = $params['construction_id'];
				break;
			}

			if(!self::updateEffect($row))
			{
				return false;
			}

			$msg_mode = 'MSG_ACTIVATE_ARTEFACT';

			$times_left = $row["times_left"]-1;
			if($row["use_duration"] > 0)
			{
				$expire_eventid = 0;
				if($planet_id > 0)
				{
					$expire_eventid = self::addEvent(EVENT_ARTEFACT_EXPIRE, max(10, $row["use_duration"])+time(), $art_id, $user_id, $planet_id);
				}

				// Updating by PK
				sqlUpdate("artefact2user", array(
					"active" => 1,
					"times_left" => $times_left,
					"expire_eventid" => $expire_eventid,
					"delay_eventid" => 0,
				), "artid=".sqlVal($art_id));
			}
			else
			{
				$row["active"] = 1;

				Artefact::updateEffect($row, false);

				if($times_left > 0)
				{
					$delay_eventid = 0;
					if($row["delay"] > 0 && $planet_id > 0)
					{
						$delay_eventid = self::addEvent(EVENT_ARTEFACT_DELAY, max(10, $row["delay"]) + time(), $art_id, $user_id, $planet_id);
					}

				// Updating by PK
					sqlUpdate("artefact2user", array(
						"active" => 0,
						"times_left" => $times_left,
						"expire_eventid" => 0,
						"delay_eventid" => $delay_eventid,
					), "artid=".sqlVal($art_id));
				}
				else
				{
				// Updating by PK
					sqlUpdate("artefact2user", array(
						"active" 	 => 0,
						"times_left" => 0,
						"deleted" 	 => time(),
						"reason" 	 => "times_left, activate"
						), "artid=".sqlVal($art_id));

					$msg_mode = 'MSG_DISAPEAR_ARTEFACT_USED';
				}
			}

			new AutoMsg(MSG_ARTEFACT, $row['userid'], time(), array('art_id' => $art_id, 'mode' => $msg_mode));

			return true;
		}
		return false;
	}

	/**
	 *
	 * Can pack building into artefact
	 * @param int $planet_id Planet where building is located
	 * @param int $building_id
	 * @param int $user_id
	 */
	public static function canPackBuilding($planet_id, $building_id, $user_id)
	{
		if(in_array($building_id, $GLOBALS["CANT_PACK_UNITS"]))
		{
			return false;
		}
		$row = sqlQueryRow("SELECT buyable, auto_active, movable, `unique`, usable, trophy_chance, delay, use_times "
		. ", use_duration, lifetime, effect_type, max_active, construction_id, `level`"
		. ", artid, a2u.typeid, userid, planetid, active, times_left, delay_eventid, expire_eventid, lifetime_eventid "
		. " FROM ".PREFIX."artefact2user a2u "
		. " INNER JOIN ".PREFIX."artefact_datasheet ads ON a2u.typeid=ads.typeid "
		. " WHERE a2u.deleted=0 AND a2u.userid=".sqlVal($user_id) ." AND a2u.planetid=".sqlVal($planet_id)
		. " AND active=0"
		. " AND delay_eventid=0 "
		. " AND a2u.typeid = " . ARTEFACT_PACKING_BUILDING
		);
		if($row)
		{
			if( !self::checkRequirements($building_id, $user_id, $planet_id, $row['artid']) )
			// if ( !self::checkBuildingBusy($building_id, $planet_id) )
			{
				return false;
			}
			return $row['artid'];
		}
		else
		{
			return false;
		}
	}

	/**
	 *
	 * Can pack research into artefact
	 * @param int $building_id
	 * @param int $user_id
	 */
	public static function canPackResearch($planet_id, $building_id, $user_id)
	{
		if(in_array($building_id, $GLOBALS["CANT_PACK_UNITS"]))
		{
			return false;
		}
		$row = sqlQueryRow("SELECT buyable, auto_active, movable, `unique`, usable, trophy_chance, delay, use_times "
		. ", use_duration, lifetime, effect_type, max_active, construction_id, `level`"
		. ", artid, a2u.typeid, userid, planetid, active, times_left, delay_eventid, expire_eventid, lifetime_eventid "
		. " FROM ".PREFIX."artefact2user a2u "
		. " INNER JOIN ".PREFIX."artefact_datasheet ads ON a2u.typeid=ads.typeid "
		. " WHERE a2u.deleted=0 AND a2u.userid=".sqlVal($user_id) ." AND a2u.planetid=".sqlVal($planet_id)
		. " AND active=0"
		. " AND delay_eventid=0 "
		. " AND a2u.typeid = " . ARTEFACT_PACKING_RESEARCH
		);
		if($row)
		{
			if( !self::checkRequirements($building_id, $user_id, $planet_id, $row['artid']) )
			// if ( !self::checkResearchBusy($building_id, $user_id) )
			{
				return false;
			}
			return $row['artid'];
		}
		else
		{
			return false;
		}
	}

	/**
	 * Call to deactivate the artefact
	 */
	public static function deactivate($art_id, $user_id, $planet_id, $user_action = true)
	{
		$row = sqlQueryRow("SELECT buyable, auto_active, movable, `unique`, usable, trophy_chance, delay, use_times "
		. ", use_duration, lifetime, effect_type, max_active, construction_id, `level`"
		. ", artid, a2u.typeid, userid, planetid, active, times_left, delay_eventid, expire_eventid, lifetime_eventid "
		. " FROM ".PREFIX."artefact2user a2u "
		. " INNER JOIN ".PREFIX."artefact_datasheet ads ON a2u.typeid=ads.typeid "
		. " WHERE a2u.deleted=0 AND a2u.userid=".sqlVal($user_id) // ." AND a2u.planetid=".sqlVal($planet_id)
		. " AND a2u.artid=".sqlVal($art_id)." AND active=1 AND expire_eventid!=0"
		. ($user_action ? " AND delay_eventid=0 AND auto_active=0 " : "")
		. " AND effect_type in (".sqlArray(ARTEFACT_EFFECT_TYPE_PLANET, ARTEFACT_EFFECT_TYPE_EMPIRE, ARTEFACT_EFFECT_TYPE_AUTO).")"
		);
		if($row)
		{
			$art_id = $row["artid"];
			$type_id = $row["typeid"];
			$user_id = $row["userid"];
			$planet_id = $row["planetid"];

			if($row["active"])
			{
				Artefact::updateEffect($row, false);
			}

			if(sqlSelectField("planet", "count(*)", "", "planetid=".sqlVal($planet_id)) == 0)
			{
				// Updating by PK
				sqlUpdate("artefact2user", array(
					"active" => 0,
					"deleted" => time(),
					"reason" => !$user_action ? "planet is not found" : "planet is not found, deactivate"
					), "artid=".sqlVal($art_id));

				return true;
			}
			else if($row["times_left"] > 0)
			{
				$msg_mode = 'MSG_DEACTIVATE_ARTEFACT';
				$delay_eventid = 0;
				if($row["delay"] > 0)
				{
					$delay_eventid = self::addEvent(EVENT_ARTEFACT_DELAY, max(10, $row["delay"]) + time(), $art_id, $user_id, $planet_id);
				}
				// Updating by PK
				sqlUpdate("artefact2user", array(
					"active" => 0,
					"expire_eventid" => 0,
					"delay_eventid" => $delay_eventid,
				), "artid=".sqlVal($art_id));
			}
			else
			{
				// Updating by PK
				$msg_mode = 'MSG_DISAPEAR_ARTEFACT_USED';
				sqlUpdate("artefact2user", array(
					"active" => 0,
					"deleted" => time(),
					"reason" => !$user_action ? "times_left" : "times_left, deactivate"
					), "artid=".sqlVal($art_id));
			}
			new AutoMsg(MSG_ARTEFACT, $row['userid'], time(), array('art_id' => $art_id, 'mode' => $msg_mode));

			return true;
		}
		return false;
	}

	/**
	 * Call on attaching the artefact to fleet
	 */
	public static function attachToFleet($art_id, $user_id, $planet_id, $activate_auto = true)
	{
		$row = sqlQueryRow("SELECT buyable, auto_active, movable, `unique`, usable, trophy_chance, delay, use_times "
			. ", use_duration, lifetime, effect_type, max_active, construction_id, `level`"
			. ", artid, a2u.typeid, userid, planetid, active, times_left, delay_eventid, expire_eventid, lifetime_eventid "
			. " FROM ".PREFIX."artefact2user a2u "
			. " INNER JOIN ".PREFIX."artefact_datasheet ads ON a2u.typeid=ads.typeid "
			. " WHERE a2u.deleted=0 AND a2u.userid=".sqlVal($user_id)." AND a2u.planetid=".sqlVal($planet_id)
			. " AND a2u.artid=".sqlVal($art_id)." AND movable=1 AND active=0 AND delay_eventid=0 AND expire_eventid=0 AND times_left>0"
//			. " AND effect_type in (".sqlArray(ARTEFACT_EFFECT_TYPE_FLEET, ARTEFACT_EFFECT_TYPE_BATTLE).")"
		);
		if($row)
		{
			if($row["effect_type"] == ARTEFACT_EFFECT_TYPE_FLEET && $activate_auto)
			{
				self::updateEffect($row);

				// Updating by PK
				sqlUpdate("artefact2user", array(
					"planetid" => 0,
					"active" => 1,
					"times_left" => $row["times_left"] - 1,
//					"active" => 0,
//					"expire_eventid" => 0,
//					"delay_eventid" => $delay_eventid,
				), "artid=".sqlVal($art_id));
			}
			else
			{
				// Updating by PK
				sqlUpdate("artefact2user", array(
					"planetid" => 0,
				), "artid=".sqlVal($art_id));
			}
			return true;
		}
		return false;
	}

	/**
	 *
	 * Enter description here ...
	 * @param unknown_type $art_id
	 */
	public static function getArtefactCost($art_id = NULL, $type = NULL, $level = NULL, $const_id = NULL)
	{
		if ( $art_id != NULL )
		{
			$data = sqlSelectRow(
				'artefact2user AS a',
				'c.basic_credit, a.level',
				'LEFT JOIN ' . PREFIX . 'construction AS c ON c.buildingid = a.typeid',
				'a.artid = '.sqlVal($art_id)
			);
			if ($data['level'] != 0)
			{
				return $data['level'] * $data['basic_credit'];
			}
			else
			{
				return $data['basic_credit'];
			}
		}
		else
		{
			return 1;
		}
	}

	/**
	 * Call on arrival of the artefact to planet
	 */
	public static function detachFromFleet($art_id, $user_id, $planet_id)
	{
		$row = sqlQueryRow("SELECT buyable, auto_active, movable, `unique`, usable, trophy_chance, delay, use_times "
			. ", use_duration, lifetime, effect_type, max_active, construction_id, `level`"
			. ", artid, a2u.typeid, userid, planetid, active, times_left, delay_eventid, expire_eventid, lifetime_eventid "
			. " FROM ".PREFIX."artefact2user a2u "
			. " INNER JOIN ".PREFIX."artefact_datasheet ads ON a2u.typeid=ads.typeid "
			. " WHERE a2u.deleted=0 AND a2u.userid=".sqlVal($user_id)." AND a2u.planetid=0"
			. " AND a2u.artid=".sqlVal($art_id)." AND delay_eventid=0 AND expire_eventid=0"
//			. " AND effect_type in (".sqlArray(ARTEFACT_EFFECT_TYPE_FLEET, ARTEFACT_EFFECT_TYPE_BATTLE).")"
		);
		if($row)
		{
			// Updating by PK
			sqlUpdate('artefact2user', array('transport_eventid' => 0), 'artid='.sqlVal($art_id));
			if($row["times_left"] > 0)
			{
				$new_userid = sqlSelectField("planet", "userid", "", "planetid=".sqlVal($planet_id));
				$delay_eventid = 0;
				if($new_userid == $user_id)
				{
					if($row["delay"] > 0)
					{
						$delay_eventid = self::addEvent(EVENT_ARTEFACT_DELAY, max(10, $row["delay"]) + time(), $art_id, $user_id, $planet_id);
					}

					// Updating by PK
					sqlUpdate("artefact2user", array(
						"planetid"		=> $planet_id,
						"active"		=> 0,
						"delay_eventid" => $delay_eventid,
					// "expire_eventid" => 0,
					), "artid=".sqlVal($art_id));
					if($row["auto_active"])
					{
						self::activate($art_id, $user_id, $planet_id, false);
					}
				}
				else
				{
					if($row["delay"] > 0)
					{
						$delay_eventid = self::addEvent(EVENT_ARTEFACT_DELAY, max(10, $row["delay"]) + time(), $art_id, $user_id, $planet_id);
					}
					// Updating by PK
					sqlUpdate("artefact2user", array(
						"planetid"		=> $planet_id,
						"active"		=> 0,
						"userid"		=> $user_id,
						"delay_eventid" => $delay_eventid,
					), "artid=".sqlVal($art_id));

					/*					self::onDisappearEvent(array(
						"new_userid" => $new_userid,
						"userid" => $user_id,
						"artid" => $art_id,
						"eventid" => $row["lifetime_eventid"],
						));

						self::appear($row["typeid"], $new_userid, $planet_id, array("delay" => 0));*/
				}
			}
			else
			{
				// Updating by PK
				sqlUpdate("artefact2user", array(
					"active" => 0,
					"deleted" => time(),
					"reason" => "times_left, detach"
					), "artid=".sqlVal($art_id));
			}
			//some action depending on $type_id
			return true;
		}
		return false;
	}

	/**
	 * Call on attaching the artefact to fleet
	 */
	public static function triggerBattle($art_id)
	{
		$row = sqlQueryRow("SELECT buyable, auto_active, movable, `unique`, usable, trophy_chance, delay, use_times "
		. ", use_duration, lifetime, effect_type, max_active, construction_id, `level`"
		. ", artid, a2u.typeid, userid, planetid, active, times_left, delay_eventid, expire_eventid, lifetime_eventid "
		. " FROM ".PREFIX."artefact2user a2u "
		. " INNER JOIN ".PREFIX."artefact_datasheet ads ON a2u.typeid=ads.typeid "
		. " WHERE a2u.deleted=0 AND a2u.planetid=0"
		. " AND a2u.artid=".sqlVal($art_id)
        // . " AND active=0 "
        . " AND delay_eventid=0 AND expire_eventid=0"
		// . " AND times_left>0 AND effect_type=".ARTEFACT_EFFECT_TYPE_BATTLE
		);
		if($row)
		{
            if($row['times_left'] < 1){
                return false;
            }
            if($row['effect_type'] == ARTEFACT_EFFECT_TYPE_BATTLE){
                // Updating by PK
                sqlUpdate(
                    "artefact2user",
                    array(
                        "times_left" => $row["times_left"] - 1,
                    ),
                    "artid=".sqlVal($art_id)
                );
            }
			return true;
		}
		return false;
	}

	/**
	 *
	 * Call on activating the artefact
	 * @param array All data about this artefact
	 * @param bool Is artefact activated or disactivated
	 */
	public static function updateEffect($row, $activate = true) // artid, $type_id, $user_id, $planet_id, $resync = false)
	{
		$type_id 		 = $row["typeid"];
		$user_id 		 = $row["userid"];
		$planet_id 		 = $row["planetid"];
		$art_id			 = $row["artid"];
		$level			 = $row["level"];
		$construction_id = $row["construction_id"];

		$op = $activate ? "+" : "-";

		switch($type_id)
		{
			case ARTEFACT_ANNIHILATION_ENGINE:
				if($activate)
				{
					NS::$fleet_speed_factor += 0.1;
				}
				break;

			case ARTEFACT_ANNIHILATION_ENGINE_10:
				if($activate)
				{
					NS::$fleet_speed_factor += 1;
				}
				break;
		}

		if(is_bool($activate))
		{
			switch($type_id)
			{
				case ARTEFACT_MOON_CREATOR:
					if($activate)
					{
						$row = sqlSelectRow("galaxy", "*", "", "planetid=".sqlVal($planet_id)." OR moonid=".sqlVal($planet_id));
						if(!$row["moonid"] && !$row["destroy_eventid"])
						{
							$percent = mt_rand(0, 100) ? mt_rand(14, 20) : 22;
							new PlanetCreator($user_id, $row["galaxy"], $row["system"], $row["position"], 1, $percent);
						}
						else
						{
							return false;
						}
					}
					break;

				case ARTEFACT_SUPERCOMPUTER:
					// Update By PK
                    $level_diff = 1;
					sqlQuery("UPDATE ".PREFIX."research2user SET
                        added = added + GREATEST(-level, $op".sqlVal($level_diff)."),
                        level = GREATEST(0, level $op ".sqlVal($level_diff).")
						WHERE buildingid=".UNIT_COMPUTER_TECH." AND userid=".sqlVal($user_id)
					);
					// Update By PK
					sqlQuery("UPDATE ".PREFIX."user SET research_factor = research_factor $op 1 WHERE userid=".sqlVal($user_id));
					break;

				case ARTEFACT_ROBOT_CONTROL_SYSTEM:
					// Update By PK
					sqlQuery("UPDATE ".PREFIX."planet SET build_factor = build_factor $op 1 WHERE planetid=".sqlVal($planet_id));
					break;

				case ARTEFACT_MERCHANTS_MARK:
					if($activate && self::getActiveCount($type_id, $user_id, $art_id) > 0)
					{
						return false;
					}
					// Update By PK
					sqlQuery("UPDATE ".PREFIX."user SET exchange_rate = ".($activate ? 1.03 : 1.2)." WHERE userid=".sqlVal($user_id));
					break;

				case ARTEFACT_CATALYST:
					sqlQuery("UPDATE ".PREFIX."planet SET produce_factor = produce_factor $op 0.1 WHERE userid=".sqlVal($user_id) . ' ORDER BY planetid');
					break;

				case ARTEFACT_POWER_GENERATOR:
					// Update By PK
                    $level_diff = 1;
					sqlQuery("UPDATE ".PREFIX."research2user SET
                        added = added + GREATEST(-level, $op".sqlVal($level_diff)."),
                        level = GREATEST(0, level $op ".sqlVal($level_diff).")
						WHERE buildingid=".UNIT_SHIELD_TECH." AND userid=".sqlVal($user_id)
					);
					sqlQuery("UPDATE ".PREFIX."planet SET energy_factor = energy_factor $op 0.15 WHERE userid=".sqlVal($user_id) . ' ORDER BY planetid');
					break;

				case ARTEFACT_ATOMIC_DENSIFIER:
					// Update By PK
                    $level_diff = 1;
					sqlQuery("UPDATE ".PREFIX."research2user SET
                        added = added + GREATEST(-level, $op".sqlVal($level_diff)."),
                        level = GREATEST(0, level $op ".sqlVal($level_diff).")
						WHERE buildingid=".UNIT_SHELL_TECH." AND userid=".sqlVal($user_id)
					);
					sqlQuery("UPDATE ".PREFIX."planet SET storage_factor = storage_factor $op 0.15 WHERE userid=".sqlVal($user_id) . ' ORDER BY planetid');
					break;

				case ARTEFACT_PACKED_BUILDING:
					if($activate)
					{
						return self::usePackedBuilding($construction_id, $planet_id, $user_id, $level);
					}
					break;

				case ARTEFACT_PACKED_RESEARCH:
					if($activate)
					{
						return self::usePackedResearch($construction_id, $user_id, $level);
					}
					break;

				case ARTEFACT_PACKING_BUILDING:
					if($activate)
					{
						return self::packBuilding($planet_id, $construction_id, $level, $user_id);
					}
					break;

				case ARTEFACT_PACKING_RESEARCH:
					if($activate)
					{
						return self::packResearch($planet_id, $construction_id, $level, $user_id);
					}
					break;
					
				case ARTEFACT_NANOBOT_REPAIR_SYSTEM:
					if($activate)
					{
						sqlQuery("UPDATE ".PREFIX."unit2shipyard SET shell_percent = 0, damaged = 0 WHERE planetid=".sqlVal($planet_id));
					}
					break;					
					
				case ARTEFACT_BUG:
					if($activate)
					{
						sqlQuery("UPDATE ".PREFIX."unit2shipyard SET 
							damaged = CEIL( damaged + ( quantity - damaged ) * ( 0.45 + 0.1 * RAND() ) ),
							shell_percent = shell_percent * ( 0.45 + 0.1 * RAND() )
							WHERE planetid=".sqlVal($planet_id));
					}
					break;					
			}
		}
		return true;
	}

	public static function resyncUpdateEffect($row, $activate = true) // artid, $typeid, $userid, $planetid, $resync = false)
	{
		$typeid 		 = $row["typeid"];
		$userid 		 = $row["userid"];
		$planetid 		 = $row["planetid"];
		$artid			 = $row["artid"];
		$level			 = $row["level"];
		$construction_id = $row["construction_id"];

		$op = $activate ? "+" : "-";

		switch($typeid)
		{
			case ARTEFACT_ANNIHILATION_ENGINE:
				if($activate)
				{
					NS::$fleet_speed_factor += 0.1;
				}
				break;
		}

		if(is_bool($activate))
		{
			switch($typeid)
			{
				case ARTEFACT_SUPERCOMPUTER:
					// Update By PK
                    $level_diff = 1;
					sqlQuery("UPDATE ".PREFIX."research2user SET
                        added = added + GREATEST(-level, $op".sqlVal($level_diff)."),
                        level = GREATEST(0, level $op ".sqlVal($level_diff).")
						WHERE buildingid=".UNIT_COMPUTER_TECH." AND userid=".sqlVal($userid)
					);
					// Update By PK
					sqlQuery("UPDATE ".PREFIX."user SET research_factor = research_factor $op 1 WHERE userid=".sqlVal($userid));
					break;

				case ARTEFACT_ROBOT_CONTROL_SYSTEM:
					// Update By PK
					sqlQuery("UPDATE ".PREFIX."planet SET build_factor = build_factor $op 1 WHERE planetid=".sqlVal($planetid));
					break;

				case ARTEFACT_MERCHANTS_MARK:
					if($activate && self::getActiveCount($typeid, $userid, $artid) > 0)
					{
						return false;
					}
					// Update By PK
					sqlQuery("UPDATE ".PREFIX."user SET exchange_rate = ".($activate ? 1.03 : 1.2)." WHERE userid=".sqlVal($userid));
					break;

				case ARTEFACT_CATALYST:
					sqlQuery("UPDATE ".PREFIX."planet SET produce_factor = produce_factor $op 0.1 WHERE userid=".sqlVal($userid) . ' ORDER BY planetid');
					break;

				case ARTEFACT_POWER_GENERATOR:
					// Update By PK
                    $level_diff = 1;
					sqlQuery("UPDATE ".PREFIX."research2user SET
                        added = added + GREATEST(-level, $op".sqlVal($level_diff)."),
                        level = GREATEST(0, level $op ".sqlVal($level_diff).")
						WHERE buildingid=".UNIT_SHIELD_TECH." AND userid=".sqlVal($userid)
					);
					sqlQuery("UPDATE ".PREFIX."planet SET energy_factor = energy_factor $op 0.15 WHERE userid=".sqlVal($userid) . ' ORDER BY planetid');
					break;

				case ARTEFACT_ATOMIC_DENSIFIER:
					// Update By PK
                    $level_diff = 1;
					sqlQuery("UPDATE ".PREFIX."research2user SET
                        added = added + GREATEST(-level, $op".sqlVal($level_diff)."),
                        level = GREATEST(0, level $op ".sqlVal($level_diff).")
						WHERE buildingid=".UNIT_SHELL_TECH." AND userid=".sqlVal($userid)
					);
					sqlQuery("UPDATE ".PREFIX."planet SET storage_factor = storage_factor $op 0.15 WHERE userid=".sqlVal($userid) . ' ORDER BY planetid');
					break;
			}
		}
		return true;
	}

    public static function resyncUser($userid)
    {
        // !!! IF CHANGED SYNC WITH CronCommand::actionResyncArtefacts
        Lib::beginTransaction();
        try{
            sqlQuery("UPDATE ".PREFIX."user SET exchange_rate = 1.2, research_factor = 1 where userid=".sqlVal($userid));
            sqlQuery("UPDATE ".PREFIX."planet SET build_factor = 1, produce_factor = 1, energy_factor = 1, storage_factor = 1 where userid=".sqlVal($userid));
            sqlQuery("UPDATE ".PREFIX."research2user SET level = level - added, added = 0 where userid=".sqlVal($userid));
            sqlQuery("UPDATE ".PREFIX."building2planet SET level = level - added, added = 0
                where planetid IN (select planetid from ".PREFIX."planet where userid=".sqlVal($userid).") ORDER BY planetid, buildingid");

            $result = sqlSelect('user', 'userid, profession', '', 'userid='.sqlVal($userid).' AND profession > 0');
            while($row = sqlFetch($result)){
                NS::applyProfession($row['userid'], $row['profession']);
            }
            sqlEnd($result);

            $result = sqlQuery(
                "SELECT buyable, auto_active, movable, `unique`, usable, trophy_chance, delay, use_times "
                . ", use_duration, lifetime, effect_type, max_active "
                . ", a2u.artid, a2u.typeid, a2u.userid, a2u.planetid, a2u.active, a2u.times_left, a2u.delay_eventid, a2u.expire_eventid, a2u.lifetime_eventid "
                . ', e_e.mode AS e_e_mode, e_e.start AS e_e_start, e_e.processed AS e_e_processed, e_e.artid AS e_e_artid, e_e.eventid AS e_e_eventid '
                . ', e_d.mode AS e_d_mode, e_d.start AS e_d_start, e_d.processed AS e_d_processed, e_d.artid AS e_d_artid, e_d.eventid AS e_d_eventid '
                . ', e_l.mode AS e_l_mode, e_l.start AS e_l_start, e_l.processed AS e_l_processed, e_l.artid AS e_l_artid, e_l.eventid AS e_l_eventid '
                . " FROM ".PREFIX."artefact2user a2u "
                . " INNER JOIN ".PREFIX."artefact_datasheet ads ON a2u.typeid=ads.typeid "
                . " LEFT JOIN ".PREFIX."events e_e ON e_e.eventid = a2u.expire_eventid "
                . " LEFT JOIN ".PREFIX."events e_d ON e_d.eventid = a2u.delay_eventid "
                . " LEFT JOIN ".PREFIX."events e_l ON e_l.eventid = a2u.lifetime_eventid "
                . " WHERE a2u.userid=".sqlVal($userid)." AND a2u.deleted=0 AND a2u.active = 1 AND a2u.lot_id=0"
            );
            while($row = sqlFetch($result))
            {
                if(
                    $row['typeid'] != ARTEFACT_PACKED_BUILDING &&
                    $row['typeid'] != ARTEFACT_PACKED_RESEARCH &&
                    $row['typeid'] != ARTEFACT_PACKING_BUILDING &&
                    $row['typeid'] != ARTEFACT_PACKING_RESEARCH
                )
                {
                    Artefact::resyncUpdateEffect($row);
                }
            }
            sqlEnd($result);

            Lib::commitTransaction();
        }catch(Exception $e){
            Lib::rollbackTransaction();
            throw $e;
        }
    }

    /**
	 * Resynchronize effects of all active artefacts//moved to Crone
	 */
	public static function Resync()
	{
		sqlQuery("UPDATE ".PREFIX."user SET exchange_rate=1.2,research_factor=1 ORDER BY userid");
		sqlQuery("UPDATE ".PREFIX."planet SET build_factor=1,produce_factor=1,energy_factor=1,storage_factor=1 ORDER BY planetid");
		sqlQuery("UPDATE ".PREFIX."research2user SET level=GREATEST(0, level-added),added=0 ORDER BY buildingid, userid");
		sqlQuery("UPDATE ".PREFIX."building2planet SET level=GREATEST(0, level-added),added=0 ORDER BY planetid, buildingid");

		$result = sqlQuery("SELECT buyable, auto_active, movable, `unique`, usable, trophy_chance, delay, use_times "
		. ", use_duration, lifetime, effect_type, max_active "
		. ", artid, a2u.typeid, userid, planetid, active, times_left, delay_eventid, expire_eventid, lifetime_eventid "
		. ', e_e.mode AS e_e_mode, e_e.start AS e_e_start, e_e.processed AS e_e_processed, e_e.artid AS e_e_artid, e_e.eventid AS e_e_eventid '
		. ', e_d.mode AS e_d_mode, e_d.start AS e_d_start, e_d.processed AS e_d_processed, e_d.artid AS e_d_artid, e_d.eventid AS e_d_eventid '
		. ', e_l.mode AS e_l_mode, e_l.start AS e_l_start, e_l.processed AS e_l_processed, e_l.artid AS e_l_artid, e_l.eventid AS e_l_eventid '
		. " FROM ".PREFIX."artefact2user a2u "
		. " INNER JOIN ".PREFIX."artefact_datasheet ads ON a2u.typeid=ads.typeid "
		. " LEFT	JOIN ".PREFIX."events e_e ON e_e.eventid = a2u.expire_eventid "
		. " LEFT	JOIN ".PREFIX."events e_d ON e_d.eventid = a2u.delay_eventid "
		. " LEFT	JOIN ".PREFIX."events e_l ON e_l.eventid = a2u.lifetime_eventid "
		. " WHERE a2u.deleted=0 AND active=1");
		while($row = sqlFetch($result))
		{
			if ( empty($row['e_e_eventid']) )
			{
				$e_id = self::addEvent(EVENT_ARTEFACT_EXPIRE, max(10, $row["lifetime"]) + time(), $row['artid'], $row['userid'], $row['planetid']);
				sqlUpdate('artefact2user', array('expire_eventid' => $e_id), 'artid = ' . $row['artid']);
				error_log("No Expire event on active artefact. Artefact ID -> {$row['artid']}");
				// Create expire Event
			}
			if ( !empty($row['e_d_eventid']) )
			{
				sqlUpdate( 'events', array('processed' => EVENT_PROCESSED_ERROR), 'eventid = ' . $row['e_d_eventid'] );
				error_log("Found delay event. Artefact ID -> {$row['artid']} Event ID -> {$row['e_d_eventid']}");
				// Delete delay Event
			}
			if ( empty($row['e_l_eventid']) )
			{
				$e_id = self::addEvent(EVENT_ARTEFACT_DISAPPEAR, max(10, $row["lifetime"]) + time(), $row['artid'], $row['userid'], $row['planetid']);
				sqlUpdate('artefact2user', array('lifetime_eventid' => $e_id), 'artid = ' . $row['artid']);
				error_log("No Lifetime event on active artefact. Artefact ID -> {$row['artid']}");
				// Create lifetime Event
			}
			if(
				$row['typeid'] != ARTEFACT_PACKED_BUILDING &&
				$row['typeid'] != ARTEFACT_PACKED_RESEARCH &&
				$row['typeid'] != ARTEFACT_PACKING_BUILDING &&
				$row['typeid'] != ARTEFACT_PACKING_RESEARCH
			)
			{
				self::updateEffect($row);
			}
		}
		sqlEnd($result);
	}

	public static function setViewParams(&$artefact, $row, $width = null, $image_path = null, $adv_names = true, $personal = true)
	{
		$name = $row["name"];
		$id	 = $row["buildingid"];
		$link = ( empty( $row["link"] ) ? $row["buildingid"].($personal ? '' : '___1') : $row["link"] );

		$artefact["id"] = $id;
		$artefact["artid"] = $row["artid"];

		$artefact["unique"] = $row["unique"];
		$artefact["quota"] = ceil($row["quota"] * 1000);
		$artefact["usable"] = $row["usable"];
		$artefact["auto_active"] = $row["auto_active"];
		$artefact["movable"] = $row["movable"];
		$artefact["static"] = intval(!$row["movable"]);
		$artefact["trophy_chance"] = $row["trophy_chance"];
		$artefact["planet"] = intval($row["effect_type"] == ARTEFACT_EFFECT_TYPE_PLANET);
		$artefact["empire"] = intval($row["effect_type"] == ARTEFACT_EFFECT_TYPE_EMPIRE);
		$artefact["fleet"] = intval($row["effect_type"] == ARTEFACT_EFFECT_TYPE_FLEET);
		$artefact["battle"] = intval($row["effect_type"] == ARTEFACT_EFFECT_TYPE_BATTLE);

		if($row["lifetime"]) $artefact["lifetime"] = getTimeTerm($row["lifetime"]);
		if($row["use_duration"]) $artefact["duration"] = getTimeTerm($row["use_duration"]);
		if($row["delay"]) $artefact["delay"] = getTimeTerm($row["delay"]);
		$artefact["times"] = $row["use_times"];
		$artefact["max_active"] = $row["max_active"];
		$artefact["times_left"] = $row["times_left"];

		if( ($row['buildingid'] == ARTEFACT_PACKED_BUILDING || $row['buildingid'] == ARTEFACT_PACKED_RESEARCH) )
		{
			if( $adv_names )
			{
				$temp_art_id = ((isset($row["artid"]) && !empty($row["artid"])) ? $row["artid"] : $row["art_id"] );

				$temp = getArtefactNameAndPosition( $temp_art_id );
				if($personal) $temp['link'] .= '_1';

				$artefact["name"] = Link::get("game.php/ArtefactInfo/".$temp['link'], $temp['name']);
				$artefact["image"] = Link::get(
					"game.php/ArtefactInfo/".$temp['link'],
					Image::getImage(
						( !empty($image_path) )? $image_path : getUnitImage($name) ,
						Core::getLanguage()->getItem($name),
						$width,
						null,
						'',
						!empty($image_path) ? true : false
					)
				);
			}
			else
			{
				$artefact["name"] = Link::get("game.php/ArtefactInfo/".$link, Core::getLanguage()->getItem($name));
				$artefact["image"] = Link::get("game.php/ArtefactInfo/".$link, Image::getImage( ( !empty($image_path) )? $image_path :getUnitImage($name) , Core::getLanguage()->getItem($name), $width, null, '', ( !empty($image_path) )? true : false));
			}
		}
		else
		{
			$artefact["name"] = Link::get("game.php/ArtefactInfo/".$link, Core::getLanguage()->getItem($name));
			$artefact["image"] = Link::get("game.php/ArtefactInfo/".$link, Image::getImage( ( !empty($image_path) )? $image_path :getUnitImage($name) , Core::getLanguage()->getItem($name), $width, null, '', ( !empty($image_path) )? true : false));
		}
		$artefact["history"] = Link::get("game.php/ArtefactInfo/".$id, Core::getLanguage()->getItem("HISTORY"));
		$artefact["description"] = Core::getLanguage()->getItem($name."_DESC");
		$artefact["description_full"] = Core::getLanguage()->getItem($name."_FULL_DESC");

		foreach(array("lifetime","unique","usable","trophy_chance","static","auto_active","movable","planet","empire","fleet","battle") as $flag)
		{
			if($artefact[$flag])
			{
				$artefact["flags"] .= Artefact::getFlagImageHtml($flag);
			}
		}
	}

	public static function getTechLevel($user_id = null)
	{
		if(is_null($user_id))
		{
			$user_id = NS::getUser()->get("userid");
		}
		return NS::getResearch(UNIT_ARTEFACTS_TECH, $user_id);
	}

	public static function getStorageSlots($user_id = null)
	{
		return NS::getResearch(UNIT_ARTEFACTS_TECH, $user_id);
	}

	public static function getUsedSlots($user_id = null, $planet_id = null)
	{
		if(is_null($user_id))
		{
			$user_id = NS::getUser()->get("userid");
		}
		static $cache = array();
		if(!isset($cache[$user_id]))
		{
			$cache[$user_id] = sqlSelectField("artefact2user a2u", "count(*)", ""
			. "INNER JOIN ".PREFIX."artefact_datasheet ads ON a2u.typeid = ads.typeid "
			. "INNER JOIN ".PREFIX."construction c ON a2u.typeid = c.buildingid",
				"a2u.deleted=0 AND a2u.active=1 AND a2u.userid=".sqlVal($user_id));
		}
		return $cache[$user_id];
	}

	public static function hasFreeSlots($user_id = null, $planet_id = null) // canActivateArtefact
	{
		return self::getUsedSlots($user_id, $planet_id) < self::getStorageSlots($user_id);
	}

	public static function getFlagImage($name)
	{
		$image = "flag_".strtolower($name).".png";
		if(!file_exists(APP_ROOT_DIR."images/".$image))
		{
			$image = "buildings/empty/empty.gif";
		}
		return $image;
	}

	public static function getFlagImageHtml($flag)
	{
		return Image::getImage(Artefact::getFlagImage($flag), Core::getLanguage()->getItem("FLAG_".strtoupper($flag)));
	}

	protected static function checkBuildingBusy($id, $planet_id)
	{
		/* if( ( NS::getBuilding($id, $planet_id) ) <= (1 - $modif) )
		{
			return false;
		} */
		if( NS::getEH() && NS::getPlanet() && NS::getPlanet()->getPlanetId() == $planet_id )
		{
			$events = NS::getEH()->getBuildingEvents();
		}
		else
		{
			// TODO: have to load events from db
			$events = array();

		}
		foreach($events as $event)
		{
			if( $event["data"]["buildingid"] == $id )
			{
				return false;
			}
		}
		return true;
	}

	protected static function usePackedBuilding($id, $planet_id, $user_id, $packed_level)
	{
		if ( $id <= 0 || $planet_id <= 0 || $user_id <= 0 || $packed_level <= 0 || !self::checkBuildingBusy($id, $planet_id) )
		{
			return false;
		}
		$data = sqlSelectRow('construction', '*', '', "buildingid = ".sqlVal($id)." AND mode in (".sqlArray(UNIT_TYPE_CONSTRUCTION, UNIT_TYPE_MOON_CONSTRUCTION).")");
		if(empty($data))
		{
			return false;
		}
		$add_row = sqlSelectRow("building2planet b", "b.*",
			"JOIN ".PREFIX."planet p ON b.planetid = p.planetid",
			"b.planetid = ".sqlVal($planet_id)." AND buildingid = ".sqlVal($id)." AND p.userid=".sqlVal($user_id));
		$insert = $add_row ? false : true;
		$cur_level = 0;
		$new_level = 0;
		if( $insert ) // && $packed_level == 1 )
		{
			$add_row = array( 'level' => 0 );
			$cur_level = $add_row['level'];
			$new_level = self::getUpgradedLevel($add_row, $packed_level);
			sqlInsert(
				'building2planet',
				array(
					'planetid'	=> $planet_id,
					'buildingid'=> $id,
					'level'		=> $new_level
				)
			);
			if( $id == UNIT_EXCHANGE )
			{
				sqlInsert(
					'exchange',
					array(
						'eid' => $planet_id,
						'uid' => $user_id,
						'title' => Core::getLang()->get('MENU_STOCK'),
						'fee' => EXCH_FEE_MIN,
						'def_fee' => EXCH_FEE_MIN,
						'comission' => EXCH_COMMISSION_MIN
					)
				);
			}
		}
		else if( !$insert )
		{
			$cur_level = $add_row['level'];
			$new_level = self::getUpgradedLevel($add_row, $packed_level);
			if( $new_level > 0 )
			{
				// Update By Pk
				sqlUpdate(
					'building2planet',
					array(
						'level' => $new_level
					),
					'buildingid = '.sqlVal($id).' AND planetid = '.sqlVal($planet_id)
				);
			}
		}
		if($new_level == 0)
		{
			return false;
		}

		$count = 0;
		$points = 0;
		for($i = $cur_level + 1; $i <= $new_level; $i++)
		{
			if($data["basic_metal"] > 0)
			{
			  $points += parseChargeFormula($data["charge_metal"], $data["basic_metal"], $i);
			}
			if($data["basic_silicon"] > 0)
			{
			  $points += parseChargeFormula($data["charge_silicon"], $data["basic_silicon"], $i);
			}
			if($data["basic_hydrogen"] > 0)
			{
			  $points += parseChargeFormula($data["charge_hydrogen"], $data["basic_hydrogen"], $i);
			}
			$count++;
		}
		$points = round($points * RES_TO_BUILD_POINTS, POINTS_PRECISION);

		// Update By Pk
		sqlQuery("UPDATE ".PREFIX."user SET "
			. " points = points + ".sqlVal($points)
			. ", b_points = b_points + ".sqlVal($points)
			. ", b_count = b_count + ".sqlVal($count)
            . ", ".updateDmPointsSetSql()
			. " WHERE userid = ".sqlVal($user_id));

		if(in_array($id, array(UNIT_METALMINE, UNIT_SILICON_LAB, UNIT_HYDROGEN_LAB)))
		{
			updateMinerPoints($user_id, $points, false);
		}

		return true;
	}

	protected static function checkResearchBusy($id, $user_id)
	{
		/* if( ( NS::getResearch($id, $user_id) ) <= (1 - $modif) )
		{
			return false;
		} */
		if( NS::getEH() && NS::getUser() && NS::getUser()->get("userid") == $user_id )
		{
			$events = NS::getEH()->getResearchEvents();
		}
		else
		{
			// TODO: have to load events from db
			$events = array();

		}
		foreach($events as $event)
		{
			if( $event["data"]["buildingid"] == $id )
			{
				return false;
			}
		}
		return true;
	}

	protected static function usePackedResearch($id, $user_id, $packed_level)
	{
		if ( $id <= 0 || $user_id <= 0 || $packed_level <= 0 || !self::checkResearchBusy($id, $user_id) )
		{
			return false;
		}
		$data = sqlSelectRow('construction', '*', '', 'buildingid = '.sqlVal($id)." AND mode = ".sqlVal(UNIT_TYPE_RESEARCH));
		if(empty($data))
		{
			return false;
		}
		$add_row = sqlSelectRow('research2user', '*', '', 'buildingid = '.sqlVal($id).' AND userid = '.sqlVal($user_id));
		$insert	= $add_row ? false : true;
		$cur_level = 0;
		$new_level = 0;
		if ( $insert ) // && $packed_level == 1 )
		{
			$add_row = array( 'level' => 0 );
			$cur_level = $add_row['level'];
			$new_level = self::getUpgradedLevel($add_row, $packed_level);
			sqlInsert('research2user', array(
				'buildingid' => $id,
				'userid' => $user_id,
				'level' => $new_level,
				'added' => 0,
			));
		}
		elseif( !$insert )
		{
			$cur_level = $add_row['level'];
			$new_level = self::getUpgradedLevel($add_row, $packed_level);
			if($new_level > 0)
			{
				// Update By Pk
				sqlUpdate('research2user', array(
						'level' => $new_level
					),
					'buildingid = '.sqlVal($id).' AND userid = '.sqlVal($user_id)
				);
			}
		}
		if($new_level == 0)
		{
			return false;
		}

		$count = 0;
		$points = 0;
		for($i = $cur_level + 1; $i <= $new_level; $i++)
		{
			if($data["basic_metal"] > 0)
			{
			  $points += parseChargeFormula($data["charge_metal"], $data["basic_metal"], $i);
			}
			if($data["basic_silicon"] > 0)
			{
			  $points += parseChargeFormula($data["charge_silicon"], $data["basic_silicon"], $i);
			}
			if($data["basic_hydrogen"] > 0)
			{
			  $points += parseChargeFormula($data["charge_hydrogen"], $data["basic_hydrogen"], $i);
			}
			$count++;
		}
		$points = round($points * RES_TO_RESEARCH_POINTS, POINTS_PRECISION);

		// Update By Pk
		sqlQuery("UPDATE ".PREFIX."user SET "
			. " points = points + ".sqlVal($points)
			. ", r_points = r_points + ".sqlVal($points)
			. ", r_count = r_count + ".sqlVal($count)
            . ", ".updateDmPointsSetSql()
			. " WHERE userid = ".sqlVal($user_id));

		return true;
	}

	protected static function packBuilding($planet_id, $construction_id, $level, $user_id)
	{
		if ( $planet_id <= 0 || $construction_id <= 0 || $user_id <= 0
				|| !self::checkBuildingBusy($construction_id, $planet_id) )
		{
			return false;
		}
		$data = sqlSelectRow('construction', '*', '', "buildingid = ".sqlVal($construction_id)." AND mode in (".sqlArray(UNIT_TYPE_CONSTRUCTION, UNIT_TYPE_MOON_CONSTRUCTION).")");
		if(empty($data))
		{
			return false;
		}
		$row = sqlSelectRow("building2planet b", "b.*",
			"JOIN ".PREFIX."planet p ON b.planetid = p.planetid",
			"b.planetid = ".sqlVal($planet_id)." AND buildingid = ".sqlVal($construction_id)." AND p.userid=".sqlVal($user_id));
		if(empty($row) || $row["level"] - $row["added"] != $level)
		{
			return false;
		}

		$params = array(
			'level'	=> $level,
			'construction_id' => $construction_id,
		);
		$new_art_id = self::appear(ARTEFACT_PACKED_BUILDING, $user_id, $planet_id, $params);

		$level = $row["level"];
		if($level - 1 > 0)
		{
			// Update By Pk
			sqlUpdate('building2planet',
				array( 'level' => $level - 1 ),
				'planetid = ' . sqlVal($planet_id) . ' AND buildingid = ' . sqlVal($construction_id)
			);
		}
		else
		{
			// Delete By Pk
			sqlDelete('building2planet',
				'planetid = ' . sqlVal($planet_id) . ' AND buildingid = ' . sqlVal($construction_id)
			);
			if ($construction_id == UNIT_EXCHANGE)
			{
				sqlDelete("exchange", "eid=".sqlVal($planet_id));
			}
		}

		$points = 0;
		if($data["basic_metal"] > 0)
		{
		  $points += parseChargeFormula($data["charge_metal"], $data["basic_metal"], $level);
		}
		if($data["basic_silicon"] > 0)
		{
		  $points += parseChargeFormula($data["charge_silicon"], $data["basic_silicon"], $level);
		}
		if($data["basic_hydrogen"] > 0)
		{
		  $points += parseChargeFormula($data["charge_hydrogen"], $data["basic_hydrogen"], $level);
		}
		$points = round($points * RES_TO_BUILD_POINTS, POINTS_PRECISION);

		// Update By Pk
		sqlQuery("UPDATE ".PREFIX."user SET "
			. " points = points - ".sqlVal($points)
			. ", b_points = b_points - ".sqlVal($points)
			. ", b_count = b_count - 1"
            . ", ".updateDmPointsSetSql()
			. " WHERE userid = ".sqlVal($user_id));

		return true;
	}

	protected static function packResearch($planet_id, $construction_id, $level, $user_id)
	{
		if ( $planet_id <= 0 || $construction_id <= 0 || $user_id <= 0
				|| !self::checkResearchBusy($construction_id, $user_id) )
		{
			return false;
		}
		$data = sqlSelectRow('construction', '*', '', 'buildingid = '.sqlVal($construction_id)." AND mode = ".sqlVal(UNIT_TYPE_RESEARCH));
		if(empty($data))
		{
			return false;
		}
		$row = sqlSelectRow('research2user', '*', '', 'buildingid = '.sqlVal($construction_id).' AND userid = '.sqlVal($user_id));
		if(empty($row) || $row["level"] - $row["added"] != $level)
		{
			return false;
		}
		$params = array(
			'level'	=> $level,
			'construction_id' => $construction_id,
		);
		$new_art_id = self::appear(ARTEFACT_PACKED_RESEARCH, $user_id, $planet_id, $params);

		$level = $row["level"];
		if($level - 1 > 0)
		{
			// Update By Pk
			sqlUpdate('research2user',
				array( 'level' => $level - 1 ),
				'userid = ' . sqlVal($user_id) . ' AND buildingid = ' . sqlVal($construction_id)
			);
		}
		else
		{
			// Delete By Pk
			sqlDelete('research2user',
				'userid = ' . sqlVal($user_id) . ' AND buildingid = ' . sqlVal($construction_id)
			);
		}

		$points = 0;
		if($data["basic_metal"] > 0)
		{
		  $points += parseChargeFormula($data["charge_metal"], $data["basic_metal"], $level);
		}
		if($data["basic_silicon"] > 0)
		{
		  $points += parseChargeFormula($data["charge_silicon"], $data["basic_silicon"], $level);
		}
		if($data["basic_hydrogen"] > 0)
		{
		  $points += parseChargeFormula($data["charge_hydrogen"], $data["basic_hydrogen"], $level);
		}
		$points = round($points * RES_TO_RESEARCH_POINTS, POINTS_PRECISION);

		// Update By Pk
		sqlQuery("UPDATE ".PREFIX."user SET "
			. " points = points - ".sqlVal($points)
			. ", r_points = r_points - ".sqlVal($points)
			. ", r_count = r_count - 1"
            . ", ".updateDmPointsSetSql()
			. " WHERE userid = ".sqlVal($user_id));

		return true;
	}

	public static function getUpgradedLevel($row, $packed_level)
	{
		// t($row);
		// t($packed_level);
		// $new_level = 0;
		if(isset($row['level']))
		{
			$added = isset($row['added']) ? $row['added'] : 0;
			$true_level = $row['level'] - $added;
			if($packed_level > $true_level)
			{
				$diff = $packed_level - $true_level;
				if($diff > 1)
				{
					$new_level = floor( $true_level + $diff / 2 ) + (int)( $row['level'] > 0 );
				}
				else
				{
					$new_level = $packed_level;
				}
				return $new_level + $added;
			}

			/*
			if($packed_level > $row['level'])
			{
				$diff = $packed_level - $row['level'];
				// $new_level = round( $row['level'] + max(pow($diff, 0.7), $diff/2) );
				if($diff > 1)
				{
					$new_level = floor( $row['level'] + $diff / 2 ) + (int)( $row['level'] > 0 );
				}
				else
				{
					$new_level = $packed_level;
				}
			}
			if(isset($row['added']) && $row['added'] > 0 && $packed_level > $row['level'] - $row['added'])
			{
				$diff = $packed_level - ($row['level'] - $row['added']);
				// $new_level_2 = round( $row['level'] + max(pow($diff, 0.7), $diff/2) );
				if($diff > 1)
				{
					$new_level_2 = floor( $row['level'] + $diff / 2 ) + (int)( $row['level'] > 0 ) + $row['added'];
					$new_level = max($new_level, $new_level_2);
				}
			}
			*/
		}
		// $new_level = min($new_level, $packed_level); // make it safely
		return 0; // $new_level;
	}
}
?>