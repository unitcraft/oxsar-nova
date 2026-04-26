<?php
/**
* Allows the user to send fleets to a mission.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class Mission extends Page
{
	/**
	* Constructor. Handles requests for this page.
	*
	* @return void
	*/
	public function __construct()
	{
		parent::__construct();
		Core::getTPL()->addHTMLHeaderFile("fleet.js?".CLIENT_VERSION, "js");
		if(!NS::getUser()->get("umode") && !NS::getUser()->get("observer") && !(DEATHMATCH && (!isDeathmatchStarted() || isDeathmatchEnd())))
		{
			$this
				->setPostAction("stargatejump", "starGateJump")
				->setPostAction("stargatejump_defense", "starGateDefenseJump")
				->setPostAction("execjump", "executeJump")
				->setPostAction("retreat", "retreatFleet")
				->setPostAction("formation", "formation")
				->setPostAction("invite", "invite")
				->setPostAction("step2", "selectCoordinates")
				->setPostAction("step3", "selectMission")
				->setPostAction("step4", "sendFleet")
				->addPostArg("selectMission", "galaxy")
				->addPostArg("selectMission", "system")
				->addPostArg("selectMission", "position")
				->addPostArg("selectMission", "targetType")
				->addPostArg("selectMission", "speed")
				->addPostArg("selectCoordinates", "galaxy")
				->addPostArg("selectCoordinates", "system")
				->addPostArg("selectCoordinates", "position")
				->addPostArg("selectCoordinates", null)
				->addPostArg("sendFleet", null)
				->addPostArg("starGateJump", null)
				->addPostArg("starGateDefenseJump", null)
				->addPostArg("retreatFleet", "id")
				->addPostArg("invite", "id")
				->addPostArg("invite", "name")
				->addPostArg("invite", "username")
				->addPostArg("formation", "id")
				->addPostArg("executeJump", "moonid")
				->setPostAction("control", "controlFleetPost")
				->addPostArg("controlFleetPost", "id")
				->setGetAction("go", "ControlFleet", "controlFleet")
				->addGetArg("controlFleet", "id")
				->setPostAction("load_resources", "loadResourcesToFleet")
				->addPostArg("loadResourcesToFleet", "id")
				->addPostArg("loadResourcesToFleet", "metal")
				->addPostArg("loadResourcesToFleet", "silicon")
				->addPostArg("loadResourcesToFleet", "hydrogen")
				->setPostAction("unload_resources", "unloadResourcesFromFleet")
				->addPostArg("unloadResourcesFromFleet", "id")
				->addPostArg("unloadResourcesFromFleet", "fleetMetal")
				->addPostArg("unloadResourcesFromFleet", "fleetSilicon")
				->addPostArg("unloadResourcesFromFleet", "fleetHydrogen")
				->setPostAction("holding_select_coords", "holdingSelectCoords")
				->addPostArg("holdingSelectCoords", "id")
				->setPostAction("holding_send_fleet", "holdingSendFleet")
				->addPostArg("holdingSendFleet", null)
				/*
				->setPostAction("holding_select_mission", "holdingSelectMission")
				->addPostArg("holdingSelectMission", "id")
				->addPostArg("holdingSelectMission", "galaxy")
				->addPostArg("holdingSelectMission", "system")
				->addPostArg("holdingSelectMission", "position")
				->addPostArg("holdingSelectMission", "targetType")
				->addPostArg("holdingSelectMission", "speed")
				*/
				;
		}

		Core::getLanguage()->load("Main,info,mission,ArtefactInfo");
//		try {
			$this->proceedRequest();
//		} catch(Exception $e) {
//			$e->printError();
//		}
		return;
	}

	/**
	* Index action. Shows fleet and artefacts
	*
	* @return Mission
	*/
	protected function index()
	{
        $observer = NS::getUser()->get("observer") || (DEATHMATCH && (!isDeathmatchStarted() || isDeathmatchEnd()));
        if($observer){
            Logger::addMessage('CANT_SEND_FLEET_DUE_OBSERVER');
        }

		Core::getTPL()->addHTMLHeaderFile("lib/jquery.jclock.js?".CLIENT_VERSION, "js");
		$clock = "<script type=\"text/javascript\">
		//<![CDATA[
		$(function($) {
			var options = {
				seedTime: ".time()." * 1000
			}
			$('.jclock').jclock(options);
		});
		//]]>
		</script>
		<span class=\"jclock\"></span>";
		$time_now = date("d.m.Y", time())." ".$clock;
		Core::getTPL()->assign("serverClock", $time_now);

		Core::getTPL()->assign("max_ships", MAX_SHIPS*1000);
		Core::getTPL()->assign("max_ships_grade", MAX_SHIPS_GRADE);
		Core::getTPL()->assign("max_ships_size", (MAX_SHIPS_GRADE - 0));
		$missions = array();
		$ef_rules = $this->setExpoAndFleetRules();
		$fleetEvents = NS::getEH()->getOwnFleetEvents();
		if(is_array($fleetEvents))
		{
			$num = 0;
			$num_f = 0;
			$num_e = 0;
			foreach($fleetEvents as $event)
			{
				$num++;
				$id = $event["eventid"];
				$data = $event["data"];

				$missions[$id]["eventid"] = $event["eventid"];
				$missions[$id]["event_start"] = $event["start"];
				$missions[$id]["event_end"] = $event["time"];
				// $missions[$id]["event_timeleft"] = max(0, $event["time"] - time());
				$missions[$id]["event_percent_timeout"] = $event["time"] > $event["start"] ? ceil(($event["time"] - $event["start"]) * 1000 / 100.0) : 1000*10;
				$missions[$id]["event_pb_value"] = $event["time"] > $event["start"] ? max(1, floor((time() - $event["start"]) * 100 / ($event["time"] - $event["start"]))) : 0;

				$missions[$id]["mode"] = NS::getMissionName($event["mode"]);
				$missions[$id]["mode_r"] = $event["mode"];
				$missions[$id]["mode_o"] = ($data["oldmode"] != "") ? "(".NS::getMissionName($data["oldmode"]).")" : "";
				$missions[$id]["is_exchange"] = (int)isset($event["data"]["exchange"]);
				$missions[$id]["arrival"] = Date::timeToString(1, $event["time"]);
				$missions[$id]["return"] = Date::timeToString(1, $event["time"] + $data["time"]);
				$missions[$id]["return_exists"] = true;
				if( in_array($event["mode"], array(EVENT_POSITION, EVENT_RETURN, EVENT_ROCKET_ATTACK, EVENT_DELIVERY_UNITS)) )
				{
					$missions[$id]["return"] = $missions[$id]["arrival"];
					$missions[$id]["return_exists"] = false;
				}
				$missions[$id]["target"] = getCoordLink($data["galaxy"], $data["system"], $data["position"], false, $event["destination"]);
				$missions[$id]["start"] = getCoordLink($data["sgalaxy"], $data["ssystem"], $data["sposition"], false, $event["planetid"]);
				$missions[$id]["fleet"] = "";
				$missions[$id]["quantity"] = 0;
				$missions[$id]["time"] = $event["time"];
				$missions[$id]["id"] = $id;
				$missions[$id]["num"] = $num;
				$missions[$id]["planet"] = $event["planetid"];
				$missions[$id]['exped_event'] = false;
				$missions[$id]['exped_slot']	= 0;
				$missions[$id]['fleet_event'] = false;
				$missions[$id]['fleet_slot']	= 0;
				if( $event['fleet_slot'] )
				{
					$num_f++;
					$missions[$id]['fleet_event'] = true;
					$missions[$id]['fleet_slot']	= $num_f;
				}
				if( $event['exped_slot'] )
				{
					$num_e++;
					$missions[$id]['exped_event'] = true;
					$missions[$id]['exped_slot']	= $num_e;
				}
				if(is_array($data["ships"]))
				{
					foreach($data["ships"] as $ships)
					{
						$missions[$id]["quantity"] += floor($ships["quantity"]);
					}
					$missions[$id]["quantity"] = fNumber($missions[$id]["quantity"]);
				}
				$missions[$id]["fleet"] = getUnitListStr($data["ships"]);
			}
		}
		Core::getTPL()->addLoop("missions", $missions);

		sqlDelete("temp_fleet", "planetid = ".sqlPlanet());

		$stargate_possible = false;
		$stargate_time = 0;
		if(NS::getPlanet()->isMoon())
		{
			$stargate_possible = true;
			$stargate_time = NS::getPlanet()->getStarGateTime();
			$stargate_level = NS::getPlanet()->getBuilding(UNIT_STAR_GATE);
			Core::getTPL()->assign("stargate_level", $stargate_level);
		}
		else // if(isAdmin(null, true)/*2.6*/)
		{
			// $moonlab_level = NS::getPlanet()->getExtBuilding(UNIT_MOON_LAB);
			$stargate_level = NS::getPlanet()->getExtBuilding(UNIT_STAR_GATE);
			if(/*$moonlab_level > 0 ||*/ $stargate_level >= 3)
			{
				$stargate_possible = true;
				$stargate_time = NS::getPlanet()->getStarGateTime(true);
				Core::getTPL()->assign("stargate_level", $stargate_level);
			}
		}
		if($stargate_time > 0 && OXSAR_RELEASED)
		{
			$gate_cycle_secs = starGateRecycleSecs( $stargate_level );
			$stargate_timeleft = $stargate_time + $gate_cycle_secs - time();
			if( $stargate_timeleft > 0 )
			{
				$stargate_possible = false;
				Core::getTPL()->addHTMLHeaderFile("lib/jquery.countdown.js?".CLIENT_VERSION, "js");
				$stargateCountDown = "<script type='text/javascript'>
					//<![CDATA[
					$(function () {
						$('#stargateCountDown').countdown({until: ".$stargate_timeleft.", compact: true, onExpiry: function() {
							$('#stargateCountDown').text('-');
						}});
				});
				//]]>
				</script>
				<span id='stargateCountDown'>".getTimeTerm($stargate_timeleft)."</span>";
				Core::getTPL()->assign("stargateCountDown", $stargateCountDown);
			}
		}


		$select = array("b.buildingid", "b.name", "b.mode", "u2s.quantity", "u2s.damaged", "u2s.shell_percent", "d.capicity", "d.speed");
		$joins	= "LEFT JOIN ".PREFIX."construction b ON (b.buildingid = u2s.unitid)";
		$joins .= "LEFT JOIN ".PREFIX."ship_datasheet d ON (d.unitid = u2s.unitid)";
		$result = sqlSelect("unit2shipyard u2s", $select, $joins,
			"b.mode IN (".sqlArray(UNIT_TYPE_FLEET, UNIT_TYPE_DEFENSE).") AND u2s.planetid = ".sqlPlanet()
				// . " AND b.buildingid NOT IN (".sqlArray($GLOBALS["BLOCKED_STARGATE_UNITS"]).")"
                ,
			"b.mode, b.display_order, b.buildingid");
		$f = $d = array();
		while($row = sqlFetch($result))
		{
			$id = $row["buildingid"];

            $is_jump_blocked = in_array($id, $GLOBALS["BLOCKED_STARGATE_UNITS"]);
            if($row["mode"] == UNIT_TYPE_DEFENSE && $is_jump_blocked)
            {
                continue;
            }

			$link = "game.php/UnitInfo/".$id;
			$item = array();
			$item["name"] = Link::get($link, Core::getLanguage()->getItem($row["name"]));
			$item["image"] = Link::get($link, Image::getImage(getUnitImage($row["name"]), Core::getLanguage()->getItem($row["name"]), 60));
			$item["mode"] = $row["mode"];
			$item["quantity_raw"] = $row["quantity"];
			$item["quantity"] = getUnitQuantityStr($row, array("splitter" => "<br />", "bracket" => false));
			$item["speed"] = fNumber($speed = NS::getSpeed($id, $row["speed"]));
			$item["capicity"] = fNumber($row["capicity"] * $row["quantity"]);
			$item["id"] = $id;
            $item["blocked"] = false;

			if($item["mode"] == UNIT_TYPE_FLEET && ($speed > 0)) // || !$stargate_possible))
			{
				$f[$id] = $item;
			}
            elseif($item["mode"] == UNIT_TYPE_FLEET && $is_jump_blocked)
            {
                $item["blocked"] = true;
				$f[$id] = $item;
            }
			else // if(isAdmin()/*2.6*/ || $item["mode"] == UNIT_TYPE_FLEET)
			{
				$d[$id] = $item;
			}
		}
		sqlEnd($result);
		Core::getTPL()->addLoop("fleet", $f);
		Core::getTPL()->addLoop("defense", $d);

		$select = array("b.buildingid", "b.name",
			"a2u.active", "a2u.times_left", "a2u.delay_eventid", "a2u.expire_eventid", "a2u.lifetime_eventid",
			"ads.movable", "ads.unique", "ads.usable", "ads.use_duration", "ads.delay", "ads.use_times", "ads.lifetime", "ads.max_active", "ads.effect_type", "ads.trophy_chance",
			"a2u.planetid", "a2u.artid", "e.time as cur_lifetime",
			);
		$joins	= "INNER JOIN ".PREFIX."construction b ON (b.buildingid = a2u.typeid)";
		$joins .= "INNER JOIN ".PREFIX."artefact_datasheet ads ON (a2u.typeid = ads.typeid)";
		$joins .= "LEFT JOIN ".PREFIX."events e ON e.eventid = a2u.lifetime_eventid";
		$result = sqlSelect("artefact2user a2u", $select, $joins,
			"a2u.deleted=0 AND a2u.delay_eventid=0 AND a2u.expire_eventid=0 "
                . " AND ads.movable=1 AND a2u.active=0 AND a2u.times_left > 0 "
                . " AND a2u.userid=".sqlUser()." AND a2u.planetid=".sqlPlanet(),
			"b.display_order ASC, ads.unique DESC, ads.typeid ASC");
		$artefacts = array();
		while($row = sqlFetch($result))
		{
			$id = $row["artid"];
			$type = $row["buildingid"];
			if ( $type == ARTEFACT_PACKED_BUILDING || $type == ARTEFACT_PACKED_RESEARCH )
			{
				$temp 		= sqlSelectRow('artefact2user', '*', '', "artid = ".sqlVal($id));
				$img_link = 'cid='.$temp['construction_id'] . '&level=' . $temp['level'] . '&typeid=' . $type;
				$row['link'] = $type . '_' . $temp['level'] . '_' . $temp['construction_id'];
				Artefact::setViewParams($artefacts[$id], $row, 60,
					// YII_GAME_DIR.'/index.php?r=artefact2user_YII/image_new&'.$img_link.''
					artImageUrl("image_new", $img_link, false)
					);
			}
			else
			{
				Artefact::setViewParams($artefacts[$id], $row, 60);
			}
			if($row["cur_lifetime"] > 0 && $row["cur_lifetime"] > time())
			{
				$timeleft = max(1, $row["cur_lifetime"] - time());
				$artefacts[$id]["disappear_counter"] = "<script type='text/javascript'>
					$(function () {
						$('#disappear_counter{$artid}').countdown({until: {$timeleft}, compact: true, onExpiry: function() {
							$('#disappear_text{$id}').html('-');
						}});
					});
				</script>
				<span id='disappear_text{$artid}'><span id='disappear_counter{$artid}'>".getTimeTerm($timeleft)."</span></span>";
			}
		}
		sqlEnd($result);
		Core::getTPL()->addLoop("artefacts", $artefacts);
		$canSendFleet = false;
		if(NS::getResearch(UNIT_COMPUTER_TECH) + 1 > NS::getEH()->getUsedFleetSlots() /*count(NS::getEH()->getOwnFleetEvents())*/)
		{
			$canSendFleet = true;
		}
		Core::getTPL()->assign("canSendFleet", $canSendFleet);
		Core::getTPL()->assign("observer", $observer);

		Core::getTPL()->display("missions");
		return $this;
	}

	protected function canUseFleetSlot( $mode, $can_use_fleet )
	{
		if( NS::getEH()->isFleetSlotUsed( array( 'mode' => $mode ) ) )
		{
			return $can_use_fleet;
		}
		return true;
	}

	protected function setExpoAndFleetRules()
	{
		$fleet = NS::getEH()->getUsedFleetSlots() <= NS::getResearch(UNIT_COMPUTER_TECH);
		$expo = NS::getEH()->getUsedExpeditionSlots() < NS::getResearch(UNIT_EXPO_TECH);

		Core::getTPL()->assign("can_send_fleet", $fleet);
		Core::getTPL()->assign("can_send_expo", $expo);
		return array(
			'fleet' => $fleet,
			'expo' 	=> $expo,
		);
	}

	public static function calcMoveToCoords($galaxy, $system, $position, $ships, $not_damaged = false, $flag = 0, $speed_modifier = 1)
	{
		$ndamaged = $not_damaged ? "u2s.quantity - u2s.damaged as quantity" : "u2s.quantity";

		$data = array();
		$select = array("u2s.unitid", $ndamaged, "u2s.damaged", "u2s.shell_percent", "d.capicity", "d.speed", "d.consume", "b.name", "b.mode");
		$joins	= "LEFT JOIN ".PREFIX."construction b ON (b.buildingid = u2s.unitid)";
		$joins .= "LEFT JOIN ".PREFIX."ship_datasheet d ON (d.unitid = u2s.unitid)";
		$result = sqlSelect("unit2shipyard u2s", $select, $joins, "b.mode=".UNIT_TYPE_FLEET." AND u2s.planetid = ".sqlPlanet());
		$capicity = 0;
		$consumption = 0;
		$speed = MAX_SPEED_POSSIBLE;
		$fleet_size = 0;
		while($row = sqlFetch($result))
		{
			$id = $row["unitid"];
			$quantity = floor(min($row["quantity"], $flag ? $ships[$id]["quantity"] : $ships[$id]));
			if($quantity > 0)
			{
				$data[$id]["name"] = $row["name"];
				$data[$id]["mode"] = $row["mode"];
				extractUnits($row, $quantity, $data[$id]);

				$speed = min($speed, NS::getSpeed($id, $row["speed"]) * $speed_modifier );
				$consumption += ceil($row["consume"] * FLEET_FUEL_CONSUMPTION * NS::getGraviFuelConsumeScale()) * $data[$id]["quantity"];
				$capicity += $row["capicity"] * $data[$id]["quantity"];
				$fleet_size += $row["consume"] > 0 ? $data[$id]["quantity"] : 0;
			}
		}
		sqlEnd($result);

		if($speed == MAX_SPEED_POSSIBLE) //no ships
		{
			$speed = 0;
		}

		NS::$fleet_speed_factor = 1.0;
		if($speed > 0 && ($flag || isset($ships["art"])))
		{
			$select = array("b.buildingid", "b.name", "b.mode",
				"a2u.active", "a2u.times_left", "a2u.delay_eventid", "a2u.expire_eventid", "a2u.lifetime_eventid",
				"ads.movable", "ads.unique", "ads.usable", "ads.use_duration", "ads.delay", "ads.use_times",
				"ads.lifetime", "ads.max_active", "ads.effect_type", "ads.trophy_chance",
				"a2u.planetid", "a2u.artid", "a2u.typeid, ac.name AS con_name, a2u.construction_id, a2u.level", // "e.time as cur_lifetime",
				);
			$joins	= "INNER JOIN ".PREFIX."construction b ON (b.buildingid = a2u.typeid)";
			$joins .= "INNER JOIN ".PREFIX."artefact_datasheet ads ON (a2u.typeid = ads.typeid)";
			$joins .= "LEFT JOIN ".PREFIX."construction ac ON (a2u.construction_id = ac.buildingid)";
			$result = sqlSelect("artefact2user a2u", $select, $joins,
				"a2u.deleted=0 AND a2u.delay_eventid=0 AND a2u.expire_eventid=0 AND ads.movable=1 AND a2u.active=0 AND a2u.times_left > 0 AND a2u.userid=".sqlUser()." AND a2u.planetid=".sqlPlanet(),
				"b.display_order ASC, ads.unique DESC, ads.typeid ASC");
			while($row = sqlFetch($result))
			{
				$id = $row["buildingid"];
				$artid = $row["artid"];
				if(!$flag)
				{
					if(!isset($ships["art"][$artid]))
					{
						continue;
					}
				}
				else
				{
					if(!isset($ships[$id]["art_ids"][$artid]))
					{
						continue;
					}
				}
				$data[$id]["id"] = $id;
				$data[$id]["name"] = $row["name"];
				$data[$id]["mode"] = $row["mode"];
				$data[$id]["art_ids"][$artid] = array(
					"artid" 	=> $artid,
						'level'		=> $row["level"],
						'con_id'	=> $row["construction_id"],
						'con_name'	=> ( isset($row["con_name"]) && !empty($row["con_name"]) ) ? $row["con_name"] : ''
				);
				$data[$id]["quantity"]++;

				if( $row["effect_type"] == ARTEFACT_EFFECT_TYPE_FLEET )
				{
					Artefact::updateEffect($row, "test");
				}
			}
			sqlEnd($result);
			$speed *= NS::$fleet_speed_factor;
		}

		return array(
			"ships" => $data,
			"capicity" => $capicity,
			"consumption" => $consumption,
			"maxspeed" => $speed,
			"speed_factor" => NS::$fleet_speed_factor,
			"fleet_size" => $fleet_size,
			);
	}

	/**
	* Select the mission's target and speed.
	*
	* @param integer	Galaxy
	* @param integer	System
	* @param integer	Position
	* @param array		Fleet to send
	*
	* @return Mission
	*/
	protected function selectCoordinates($galaxy, $system, $position, $ships)
	{
		$ef_rules = $this->setExpoAndFleetRules();
		Core::getTPL()->addHTMLHeaderFile("lib/jquery.countdown.js?".CLIENT_VERSION, "js");

		$galaxy = $galaxy > 0 ? (int)$galaxy : NS::getPlanet()->getData("galaxy");
		$system = $system > 0 ? (int)$system : NS::getPlanet()->getData("system");
		$position = $position > 0 ? (int)$position : NS::getPlanet()->getData("position");

		sqlDelete("temp_fleet", "planetid = ".sqlPlanet());

		$select = array("u2s.unitid");
		$joins	= "LEFT JOIN ".PREFIX."construction b ON (b.buildingid = u2s.unitid)";
		$joins .= "LEFT JOIN ".PREFIX."ship_datasheet d ON (d.unitid = u2s.unitid)";
		$result = sqlSelect("unit2shipyard u2s", $select, $joins, "b.mode=".UNIT_TYPE_FLEET." AND u2s.planetid = ".sqlPlanet());
		while($row = sqlFetch($result))
		{
			if( isset($ships[$row['unitid']]) )
			{
				// $ships[$row['unitid']] = min( MAX_SHIPS, (int)$ships[$row['unitid']] );
				$ships[$row['unitid']] = (int)$ships[$row['unitid']];
			}
		}
		sqlEnd($result);

		$move_data = self::calcMoveToCoords($galaxy, $system, $position, $ships);
		$data = $move_data["ships"];
		$capicity = $move_data["capicity"];
		$consumption = $move_data["consumption"];
		$speed = $move_data["maxspeed"];
		$speed_factor = $move_data["speed_factor"];
		$fleet_size = $move_data["fleet_size"];

		if(count($data) > 0 && $speed > 0)
		{
			$distance = NS::getDistance($galaxy, $system, $position);
			$time = NS::getFlyTime($distance, $speed);

			Core::getTPL()->assign("galaxy", $galaxy);
			Core::getTPL()->assign("system", $system);
			Core::getTPL()->assign("position", $position);
			Core::getTPL()->assign("oGalaxy", NS::getPlanet()->getData("galaxy"));
			Core::getTPL()->assign("oSystem", NS::getPlanet()->getData("system"));
			Core::getTPL()->assign("oPos", NS::getPlanet()->getData("position"));
			Core::getTPL()->assign("maxspeedVar", $speed);
			Core::getTPL()->assign("distance", $distance);
			Core::getTPL()->assign("capicity_raw", $capicity);
			Core::getTPL()->assign("basicConsumption", $consumption);
			Core::getTPL()->assign("fleetSize", $fleet_size);
			Core::getTPL()->assign("time", getTimeTerm($time));
			Core::getTPL()->assign("maxspeed", fNumber($speed));
			Core::getTPL()->assign("capicity", fNumber($capicity - NS::getFlyConsumption($consumption, $distance)));
			Core::getTPL()->assign("fuel", fNumber(NS::getFlyConsumption($consumption, $distance)));
			Core::getTPL()->assign("fleet", getUnitListStr($data));
			Core::getTPL()->assign("allow_stargate_transport", (int)NS::getGalaxyParam(NS::getPlanet()->getData("galaxy"), "ALLOW_STARGATE_TRANSPORT", isset($data[ARTEFACT_IGLA_MORI])));

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
			$_result = sqlSelect("planet p", array("p.ismoon", "p.planetname", "g.galaxy", "g.system", "g.position", "gm.galaxy moongala", "gm.system as moonsys", "gm.position as moonpos"), "LEFT JOIN ".PREFIX."galaxy g ON (g.planetid = p.planetid) LEFT JOIN ".PREFIX."galaxy gm ON (gm.moonid = p.planetid)", "p.userid = ".sqlUser()." AND p.planetid != ".sqlPlanet(), $order);
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
		return $this->index();
	}

	protected function controlFleet($eventid)
	{
		doHeaderRedirection("game.php/Main", false);
	}

	function isTransportPossible($target_userid, $owner_userid = null)
	{
		return true;

		if(!$owner_userid)
		{
			$owner_userid = NS::getUser()->get("userid");
		}

		if(!OXSAR_RELEASED || $target_userid == $owner_userid)
		{
			return true;
		}

		$result = sqlQuery("SELECT * FROM ".PREFIX."events "
			. " WHERE mode=".sqlVal(EVENT_TRANSPORT)
			. "	 AND user=".sqlUser()
			. "	 AND destination in (SELECT planetid FROM ".PREFIX."planet WHERE userid=".sqlVal($target_userid).")"
			. "	 AND time > ".sqlVal(time()-60*60*24*7)
			. " LIMIT 1");
		$row = sqlFetch($result);
		sqlEnd($result);

		return $row ? false : true;
	}

	function getHoldingMasterUserPoints($target_planet_id)
	{
		$row = sqlQueryRow("SELECT max(u.points) as points FROM ".PREFIX."events e"
			. " JOIN ".PREFIX."user u ON u.userid = e.user"
			. " WHERE e.mode=".sqlVal(EVENT_HOLDING)
			. "	 AND e.destination=".sqlVal($target_planet_id)
			. "	 AND e.processed=".EVENT_PROCESSED_WAIT);
		return $row ? $row["points"] : 0;
	}

	function isHoldingTarget($target_planet_id, $owner_userid = null)
	{
		$users = array( NS::getUser()->get("userid") );
		if( $owner_userid && $owner_userid != NS::getUser()->get("userid") )
		{
			$users[] = $owner_userid;
		}
		$count = sqlQueryField("SELECT count(*) FROM ".PREFIX."events e"
			. " WHERE e.mode in (".sqlArray(EVENT_HALT, EVENT_HOLDING).")"
			. "	 AND e.user in (".sqlArray($users).") "
			. "	 AND e.destination=".sqlVal($target_planet_id)
			. "	 AND e.processed=".EVENT_PROCESSED_WAIT);
		return $count > 0;
	}

	function isAttackingTarget($target_planet_id, $owner_userid = null)
	{
		$users = array( NS::getUser()->get("userid") );
		if( $owner_userid && $owner_userid != NS::getUser()->get("userid") )
		{
			$users[] = $owner_userid;
		}
		$count = sqlQueryField("SELECT count(*) FROM ".PREFIX."events e"
			. " WHERE e.mode in (".sqlArray(EVENT_ATTACK_SINGLE, EVENT_ATTACK_ALLIANCE, EVENT_ROCKET_ATTACK, EVENT_ALLIANCE_ATTACK_ADDITIONAL,
					EVENT_ATTACK_DESTROY_BUILDING, EVENT_ATTACK_ALLIANCE_DESTROY_BUILDING,
					EVENT_ATTACK_DESTROY_MOON, EVENT_ATTACK_ALLIANCE_DESTROY_MOON
					).")"
			. "	 AND e.user in (".sqlArray($users).") "
			. "	 AND e.destination=".sqlVal($target_planet_id)
			. "	 AND e.processed=".EVENT_PROCESSED_WAIT);
		return $count > 0;
	}

	/**
	* Select the mission to start and stored resources.
	*
	* @param integer	Galaxy
	* @param integer	System
	* @param integer	Position
	* @param string	Target type
	* @param integer	Speed
	*
	* @return Mission
	*/
	protected function selectMission($galaxy, $system, $position, $targetType, $speed)
	{
        $bashing_info = array();
        $target_valid_by_ally_attack = true;
		$row = sqlSelectRow("temp_fleet", "data", "", "planetid = ".sqlPlanet());
		if(!$row)
		{
			Logger::dieMessage('UNKOWN_MISSION');
		}
		$data = unserialize($row["data"]);
		if(!empty($data) && isset($data["holding_eventid"]))
		{
			$holding_event = NS::getEH()->getEvent($data["holding_eventid"]);
			if( $holding_event && $holding_event["mode"] == EVENT_HOLDING && $holding_event["destination"] == NS::getPlanet()->getPlanetId() )
			{
				$remain_controls = NS::getRemainFleetControls($holding_event);
				$ef_rules = array(
					'fleet' => true,
					'expo' 	=> false,
				);
				Core::getTPL()->assign("can_send_fleet", $ef_rules["fleet"]);
				Core::getTPL()->assign("can_send_expo", $ef_rules["expo"]);
				Core::getTPL()->assign("holding_eventid", $holding_event["eventid"]);
			}
			else
			{
				Logger::dieMessage('UNKOWN_MISSION');
			}
		}
		else
		{
			$remain_controls = 1;
			$ef_rules = $this->setExpoAndFleetRules();
		}
		// if($row)
		{
			$targetMode = $targetType == "moon" ? "moonid" : "planetid";

			// Set spaceships
			if( !isset($holding_event) )
			{
				foreach($data as $key => $value)
				{
					if(isset($value["art_ids"]))
					{
						$data["ships"][$key] = $value;
					}
					else if(is_numeric($key))
					{
						fixUnitDamagedVars($value);

						$data["ships"][$key]["id"] = $key;
						$data["ships"][$key]["quantity"] = floor($value["quantity"]);
						$data["ships"][$key]["damaged"] = floor($value["damaged"]);
						$data["ships"][$key]["shell_percent"] = $value["shell_percent"];
						$data["ships"][$key]["name"] = $value["name"];
						$data["ships"][$key]["mode"] = $value["mode"];
						// $fleet[$key]["name"] = Core::getLanguage()->getItem($value["name"]);
						// $fleet[$key]["quantity"] = getUnitQuantityStr($value);
					}
				}
			}
			else
			{
				$data["ships"] = $holding_event["data"]["ships"];
			}

			$data["galaxy"] = $galaxy;
			$data["system"] = $system;
			$data["position"] = $position;
			if($speed < 1) { $speed = 1; }
			else if($speed > 100) { $speed = 100; }
			$data["speed"] = $speed;

			$select = array("p.planetid", "p.planetname", "p.ismoon", "p.diameter", "u.userid", "u.points", "u.last", "u.umode", "u.observer", "u.protection_time", "b.to", "b.banid", "u2a.aid");
			$joins	= "LEFT JOIN ".PREFIX."planet p ON (p.planetid = g.".$targetMode.")";
			$joins .= "LEFT JOIN ".PREFIX."user u ON (p.userid = u.userid)";
			$joins .= "LEFT JOIN ".PREFIX."ban_u b ON (u.userid = b.userid)";
			$joins .= "LEFT JOIN ".PREFIX."user2ally u2a ON (u2a.userid = u.userid)";
			$target = sqlSelectRow("galaxy g", $select, $joins, "galaxy = ".sqlVal($galaxy)." AND system = ".sqlVal($system)." AND position = ".sqlVal($position));
			$targetName = !empty($target["planetid"]) ? $target["planetname"] : Core::getLanguage()->getItem("UNKOWN_PLANET");
			if($targetType == "tf")
			{
				$targetName .= " (".Core::getLanguage()->getItem("TF").")";
			}
			else
			{
				$data["destination"] = $target["planetid"];
			}

			$samePos = $samePlanetSystem = false;
			if(
				$galaxy == NS::getPlanet()->getData("galaxy") &&
				$system == NS::getPlanet()->getData("system") &&
				$position == NS::getPlanet()->getData("position")
			)
			{
				$samePlanetSystem = true;
				if(NS::getPlanet()->getData("ismoon") && $targetType == "moon")
				{
					$samePos = true;
				}
				else if(!NS::getPlanet()->getData("ismoon") && $targetType == "planet")
				{
					$samePos = true;
				}
			}
			$data["expeditionMode"] = !isset($holding_event) && $position == EXPED_PLANET_POSITION && EXPEDITION_ENABLED ? 1 : 0;

			require_once(APP_ROOT_DIR."game/Relation.class.php");
			if(!isset($holding_event) || $holding_event["userid"] == NS::getUser()->get("userid"))
			{
				$relation = new Relation(NS::getUser()->get("userid"), NS::getUser()->get("aid"));
			}
			else
			{
				$relation = new Relation($holding_event["userid"], sqlSelectField("user2ally", "aid", "", "userid=".sqlVal($holding_event["userid"])));
			}

			if( !isset($holding_event) )
			{
				$owner_userid = NS::getUser()->get("userid");
				$owner_points = NS::getUser()->get("points");
			}
			else
			{
				$owner_userid = $holding_event["userid"];
				$owner_points = sqlSelectField("user", "points", "", "userid=".sqlVal($owner_userid));
			}

			// Check for available missions
			$showHoldingTime = false;
			$missions = array();
			if($samePos)
			{
				// Prevend flights to own planet
				if(
					$this->canRecycle($data["ships"]) // isset($data["ships"][UNIT_RECYCLER])
					&& $remain_controls > 0
					&& !empty($target["planetid"])
					&& (isset($holding_event) || $this->canUseFleetSlot(EVENT_RECYCLING, $ef_rules['fleet']))
				)
				{
					$missions[EVENT_RECYCLING]["mode"] = EVENT_RECYCLING;
					$missions[EVENT_RECYCLING]["mission"] = Core::getLanguage()->getItem("RECYCLING");
				}

				if(
					$targetType != "tf"
					&& $remain_controls > 0
					&& !isset($holding_event)
					&& $this->canHalt($data["ships"])
					&& (isset($holding_event) || $this->canUseFleetSlot(EVENT_HALT, $ef_rules['fleet']))
				)
				{
					$missions[EVENT_HALT]["mode"] = EVENT_HALT;
					$missions[EVENT_HALT]["mission"] = Core::getLanguage()->getItem("HALT");
					$showHoldingTime = true;
				}
			}
			else if(($target["umode"] || $target["observer"]) && OXSAR_RELEASED) // Vacation mode enabled?
			{
				if(/*$targetType == "tf" &&*/
					$this->canRecycle($data["ships"]) // isset($data["ships"][UNIT_RECYCLER])
					&& $remain_controls > 0
					&& !empty($target["planetid"])
					)
				{
					$missions[EVENT_RECYCLING]["mode"] = EVENT_RECYCLING;
					$missions[EVENT_RECYCLING]["mission"] = Core::getLanguage()->getItem("RECYCLING");
				}
			}
			else if(
				$position > MAX_NORMAL_PLANET_POSITION
				&& $remain_controls > 0
				&& $this->canRecycle($data["ships"]) // isset($data["ships"][UNIT_RECYCLER])
				// !empty($target["planetid"])
				&& (isset($holding_event) || $this->canUseFleetSlot(EVENT_RECYCLING, $ef_rules['fleet'])
				)
			)
			{
				$missions[EVENT_RECYCLING]["mode"] = EVENT_RECYCLING;
				$missions[EVENT_RECYCLING]["mission"] = Core::getLanguage()->getItem("RECYCLING");
			}
			else if( $position <= MAX_NORMAL_PLANET_POSITION )
			{
				$ignoreNP = false;
				$is_banned = false;
				if($target["userid"] == NS::getUser()->get("userid") || $target["userid"] == $owner_userid)
				{
					$ignoreNP = true;
					$newbie_protected = false;
				}
                elseif(NS::checkProtectionTime(NS::getUser()->get("protection_time")))
                {
					$ignoreNP = true;
					$newbie_protected = 2;
                }
				else
				{
					if($target["last"] <= time() - 604800)
					{
						$ignoreNP = true;
					}
					if($target["banid"] && (is_null($target["to"]) || $target["to"] >= time()))
					{
						$ignoreNP = true;
						$is_banned = true;
					}
					else if(defined("ADMIN_IP") && IPADDRESS == ADMIN_IP)
					{
						$ignoreNP = true;
					}
                    if(NS::checkProtectionTime($target["protection_time"]))
                    {
                        $ignoreNP = true;
                        $newbie_protected = 1;
                    }
					if($ignoreNP === false && !$is_banned && !$newbie_protected)
					{
						// $newbie_protected = isNewbieProtected(NS::getUser()->get("points"), max($target["points"], $holding_master_points));
						$newbie_protected = isNewbieProtected($owner_points, $target["points"]);
						if($newbie_protected)
						{
							$holding_master_points = $this->getHoldingMasterUserPoints($target["planetid"]);
							if($holding_master_points > $target["points"])
							{
								$newbie_protected = isNewbieProtected($owner_points, $holding_master_points);
							}
						}
						if($newbie_protected == 2) // strong
						{
							if($this->isNeutronAffectorFound($data["ships"]))
							{
								$newbie_protected = false;
							}
						}
					}
				}
				if(
					$target["userid"] == $owner_userid
						// && $remain_controls > 0 - position is only possible when remain_controls == 0
						&& $targetType != "tf"
						&& (isset($holding_event) || $this->canUseFleetSlot(EVENT_POSITION, $ef_rules['fleet']))
				)
				{
					$missions[EVENT_POSITION]["mode"] = EVENT_POSITION;
					$missions[EVENT_POSITION]["mission"] = Core::getLanguage()->getItem("STATIONATE");
				}

				if(
					NS::getPlanet()->getData('ismoon') && $targetType == "moon"
					&& ($remain_controls > 0 || (isset($holding_event) && $target["userid"] == $owner_userid))
					// && (!isset($holding_event) || isAdmin()) // landing needs to be refactored to set in halting mode
					// && NS::getGalaxyParam(NS::getPlanet()->getData("galaxy"), "ALLOW_STARGATE_TRANSPORT", isset($data["ships"][ARTEFACT_IGLA_MORI]))
					&& (
						(isset($holding_event) && isset($data["ships"][ARTEFACT_IGLA_MORI]))
						|| (!isset($holding_event) && NS::getGalaxyParam(NS::getPlanet()->getData("galaxy"), "ALLOW_STARGATE_TRANSPORT", isset($data["ships"][ARTEFACT_IGLA_MORI])))
						)
					&& (isset($holding_event) || $this->canUseFleetSlot(EVENT_STARGATE_TRANSPORT, $ef_rules['fleet']))
					&& ($target["userid"] == NS::getUser()->get("userid") || $target["userid"] == $owner_userid)
					&& $this->canMakeStarJump( NS::getPlanet()->getPlanetId(), $target['planetid'], $owner_userid ) != false
				)
				{
					$missions[EVENT_STARGATE_TRANSPORT]["mode"] = EVENT_STARGATE_TRANSPORT;
					$missions[EVENT_STARGATE_TRANSPORT]["mission"] = Core::getLanguage()->getItem("STARGATE_TRANSPORT");
				}

				if(
					!$samePlanetSystem
					// && $remain_controls > 0
					&& (!isset($holding_event) && $this->canUseFleetSlot(EVENT_TELEPORT_PLANET, $ef_rules['fleet']))
					&& $this->canTeleportCurPlanet($data["ships"])
				)
				{
					$missions[EVENT_TELEPORT_PLANET]["mode"] = EVENT_TELEPORT_PLANET;
					$missions[EVENT_TELEPORT_PLANET]["mission"] = Core::getLanguage()->getItem("TELEPORT_PLANET");
				}

				if(
					!empty($target["userid"])
					&& $remain_controls > 0
					&& $targetType != "tf"
					&& $this->isTransportPossible($target["userid"], $owner_userid)
					&& (isset($holding_event) || $this->canUseFleetSlot(EVENT_TRANSPORT, $ef_rules['fleet']))
				)
				{
					$missions[EVENT_TRANSPORT]["mode"] = EVENT_TRANSPORT;
					$missions[EVENT_TRANSPORT]["mission"] = Core::getLanguage()->getItem("TRANSPORT");
				}
				if(
					empty($target["planetid"])
					&& $remain_controls > 0
					&& !isset($holding_event)
					&& isset($data["ships"][UNIT_COLONY_SHIP]) // || (count($this->getRealShips($data["ships"])) > 0 && isset($data[ARTEFACT_PLANET_CREATOR])))
					&& $targetType != "tf"
					&& (isset($holding_event) || $this->canUseFleetSlot(EVENT_COLONIZE, $ef_rules['fleet']))
				)
				{
					$missions[EVENT_COLONIZE]["mode"] = EVENT_COLONIZE;
					$missions[EVENT_COLONIZE]["mission"] = Core::getLanguage()->getItem("COLONIZE");
				}
				if(/*$targetType == "tf" &&*/
					$this->canRecycle($data["ships"]) // isset($data["ships"][UNIT_RECYCLER])
					&& $remain_controls > 0
					&& !empty($target["planetid"])
					&& (isset($holding_event) || $this->canUseFleetSlot(EVENT_RECYCLING, $ef_rules['fleet']))
				)
				{
					$missions[EVENT_RECYCLING]["mode"] = EVENT_RECYCLING;
					$missions[EVENT_RECYCLING]["mission"] = Core::getLanguage()->getItem("RECYCLING");
				}
				if(
					!$is_banned
					&& Core::getOptions()->get("ATTACKING_STOPPAGE") != 1
					&& !empty($target["userid"])
					&& $remain_controls > 0
					&& $target["userid"] != NS::getUser()->get("userid")
					&& $target["userid"] != $owner_userid
					&& $targetType != "tf"
					&& $this->canAttack($data["ships"])
				)
				{
					if($newbie_protected == 2)
					{
						$data["alliance_attack"] = $this->getFormations($target["planetid"]);
						if(
							$data["alliance_attack"]
							&& $this->isNeutronAffectorFound($data["alliance_attack"]["data"]["ships"])
							&& (isset($holding_event) || $this->canUseFleetSlot(EVENT_ALLIANCE_ATTACK_ADDITIONAL, $ef_rules['fleet']))
						)
						{
							unset($data["alliance_attack"]["data"]);
							$missions[EVENT_ALLIANCE_ATTACK_ADDITIONAL]["mode"] = EVENT_ALLIANCE_ATTACK_ADDITIONAL;
							$missions[EVENT_ALLIANCE_ATTACK_ADDITIONAL]["mission"] = Core::getLanguage()->getItem("ALLIANCE_ATTACK");
						}else{
							unset($data["alliance_attack"]);
						}
					}
					else if(!$newbie_protected)
					{
						if(!$this->isHoldingTarget($target["planetid"], $owner_userid))
						{
							$isNormalFleet = $this->isNormalFleet($data["ships"]);
                            $bashingCheck = NS::checkTargetBashing($target["planetid"], $owner_userid, $bashing_info);
                            $target_valid_by_ally_attack = NS::checkTargetValidByAllyAttack($target["planetid"], $owner_userid);
							if( $target_valid_by_ally_attack && $bashingCheck && $isNormalFleet && (isset($holding_event) || $this->canUseFleetSlot(EVENT_ATTACK_SINGLE, $ef_rules['fleet'])) )
							{
								$missions[EVENT_ATTACK_SINGLE]["mode"] = EVENT_ATTACK_SINGLE;
								$missions[EVENT_ATTACK_SINGLE]["mission"] = Core::getLanguage()->getItem("ATTACK");
							}

							$data["alliance_attack"] = $this->getFormations($target["planetid"]);
							if(
								$data["alliance_attack"]
								&& (isset($holding_event) || $this->canUseFleetSlot(EVENT_ALLIANCE_ATTACK_ADDITIONAL, $ef_rules['fleet']))
							)
							{
								unset($data["alliance_attack"]["data"]);
								$missions[EVENT_ALLIANCE_ATTACK_ADDITIONAL]["mode"] = EVENT_ALLIANCE_ATTACK_ADDITIONAL;
								$missions[EVENT_ALLIANCE_ATTACK_ADDITIONAL]["mission"] = Core::getLanguage()->getItem("ALLIANCE_ATTACK");
							}

							if($target_valid_by_ally_attack && $bashingCheck && $this->canDestroyAttack($data["ships"]))
							{
								if(
									NS::getGalaxyParam($galaxy, "ALLOW_DESTROY_BUILDING", 0)
									&& (isset($holding_event) || $this->canUseFleetSlot(EVENT_ATTACK_DESTROY_BUILDING, $ef_rules['fleet']))
								)
								{
									$missions[EVENT_ATTACK_DESTROY_BUILDING]["mode"] = EVENT_ATTACK_DESTROY_BUILDING;
									$missions[EVENT_ATTACK_DESTROY_BUILDING]["mission"] = Core::getLanguage()->getItem("DESTROY_ATTACK");
								}

								if(
									$targetType == "moon"
									&& NS::getGalaxyParam($galaxy, "ALLOW_DESTROY_MOON", $target["diameter"] <= TEMP_MOON_SIZE_MAX)
									&& (isset($holding_event) || $this->canUseFleetSlot(EVENT_ATTACK_DESTROY_BUILDING, $ef_rules['fleet']))
								)
								{
									$missions[EVENT_ATTACK_DESTROY_MOON]["mode"] = EVENT_ATTACK_DESTROY_MOON;
									$missions[EVENT_ATTACK_DESTROY_MOON]["mission"] = Core::getLanguage()->getItem("MOON_DESTROY_ATTACK");
								}
							}
						}
					}
				}

/*
if($_SERVER['REMOTE_ADDR'] == '95.221.91.215')
{
    CVarDumper::dump( array(
"target" => $target,
"ships" => $data["ships"],
"canAttack" => $this->canAttack($data["ships"]),
"canHalt" => $this->canHalt($data["ships"]),
"holding_event" => $holding_event,
    ), 10, 1 );
}
*/
				if(
					!empty($target["userid"])
                    && MISSION_HALTING_OTHER_ENABLED
					&& $remain_controls > 0
					// && $target["userid"] != NS::getUser()->get("userid")
					// && $target["userid"] != $owner_userid
					&& $targetType != "tf"
					&& $this->canAttack($data["ships"])
				)
				{
					if(!OXSAR_RELEASED
						|| $target["userid"] == $owner_userid
						|| $target["userid"] == NS::getUser()->get("userid")
						|| $relation->hasRelation($target["userid"], $target["aid"]))
					{
						if(
							$this->canHalt($data["ships"])
							&& (isset($holding_event) || $this->canUseFleetSlot(EVENT_HALT, $ef_rules['fleet']))
							&& !$this->isAttackingTarget($target["planetid"], $owner_userid)
						)
						{
							$missions[EVENT_HALT]["mode"] = EVENT_HALT;
							$missions[EVENT_HALT]["mission"] = Core::getLanguage()->getItem("HALT");
							$showHoldingTime = true;
						}
					}
				}
				if(!empty($target["userid"])
					&& $remain_controls > 0
					&& !isset($holding_event)
					&& $target["userid"] != NS::getUser()->get("userid")
					&& $target["userid"] != $owner_userid
					&& isset($data["ships"][UNIT_ESPIONAGE_SENSOR])
					&& $targetType != "tf"
					&& $this->canSpy($data["ships"]))
				{
					if(
						!$is_banned
						&& !$newbie_protected
						// && Core::getOptions()->get("ATTACKING_STOPPAGE") != 1
						&& (isset($holding_event) || $this->canUseFleetSlot(EVENT_SPY, $ef_rules['fleet']))
					)
					{
						$missions[EVENT_SPY]["mode"] = EVENT_SPY;
						$missions[EVENT_SPY]["mission"] = Core::getLanguage()->getItem("SPY");
					}
				}
			}
			if (!isset($holding_event)
					&& $data["expeditionMode"] == '1'
					&& $remain_controls > 0
				)
			{
				$lvl = NS::getResearch(UNIT_EXPO_TECH);
				$missions = array();
				if ($lvl < 1)
				{
					Logger::dieMessage("RESEARCH_EXPO_TECH");
				}
				$expeditions = NS::getEH()->getUsedExpeditionSlots();
				if ($expeditions >= $lvl)
				{
					Logger::dieMessage("TOO_MANY_EXPOS");
				}

				if (EXPEDITION_ENABLED && $this->canExpedition($data["ships"], $ef_rules['expo']))
				{
					$exp_time = "";
					for ($i = 0; $i <= $lvl; $i++)
					{
						$exp_time .= createOption($i, $i, false);
					}
					$missions[EVENT_EXPEDITION]["mode"] = EVENT_EXPEDITION;
					$missions[EVENT_EXPEDITION]["mission"] = Core::getLanguage()->getItem("EVENT_MODE_EXPEDITION");
					Core::getTPL()->assign("expeditionMode", true);
					Core::getTPL()->assign("expTime", $exp_time);
				}

				if(
					empty($target["planetid"])
					&& !isset($holding_event)
					&& isset($data["ships"][UNIT_COLONY_SHIP]) // || (count($this->getRealShips($data["ships"])) > 0 && isset($data[ARTEFACT_PLANET_CREATOR])))
					&& $targetType != "tf"
					// && (NS::getUser()->get("userid") == 2 || NS::getUser()->get("userid") == 312247)
					&& (isset($holding_event) || $this->canUseFleetSlot(EVENT_COLONIZE_RANDOM_PLANET, $ef_rules['fleet']))
				)
				{
					$missions[EVENT_COLONIZE_RANDOM_PLANET]["mode"] = EVENT_COLONIZE_RANDOM_PLANET;
					$missions[EVENT_COLONIZE_RANDOM_PLANET]["mission"] = Core::getLanguage()->getItem("COLONIZE_RANDOM_PLANET");
				}
			}
			
			$planet_under_attack = NS::isPlanetUnderAttack();
			if($planet_under_attack){
				$missions = array();
			}

			foreach($missions as $key => $value)
			{
				$data["amissions"][] = $key;
			}

			// debug_var($data, "[selectMission] result data");

			Core::getQuery()->update(
				"temp_fleet",
				"data",
				serialize($data),
				"planetid = ".sqlPlanet()
					. ' ORDER BY planetid'
			);

			$distance = NS::getDistance($galaxy, $system, $position);
			// $consumption = NS::getFlyConsumption($data["consumption"], $distance);
			$consumption = NS::getFlyConsumption($data["consumption"], $distance, $speed);

			if(!isset($holding_event)
				&& (isset($missions[EVENT_ATTACK_SINGLE]) || isset($missions[EVENT_HALT])))
			{
				if( NS::getGalaxyParam($galaxy, "ADVANCED_BATTLE", false) )
				{
					$aval_tech = sqlArray(
						UNIT_GUN_TECH, UNIT_SHIELD_TECH,
						UNIT_SHELL_TECH, UNIT_BALLISTICS_TECH, UNIT_MASKING_TECH,
						UNIT_LASER_TECH, UNIT_ION_TECH, UNIT_PLASMA_TECH
					);
				}
				else
				{
					$aval_tech = sqlArray( UNIT_GUN_TECH, UNIT_SHIELD_TECH, UNIT_SHELL_TECH, UNIT_BALLISTICS_TECH, UNIT_MASKING_TECH);
				}
				$balanceAttackLevels = array();
				$exp_active_levels = min(MAX_ADD_TECH_LEVELS, floor(NS::getUser()->get("be_points") / POINTS_PER_ADD_TECH_LEVEL));
				if($exp_active_levels > 0)
				{
					$t_result = sqlSelect("construction", array("buildingid", "name"), "",
						"buildingid in (".$aval_tech.")",
						"display_order ASC, buildingid ASC");
					while($t_row = sqlFetch($t_result))
					{
						$attackLevelsItem = array(
							"tech_id" => $t_row["buildingid"],
							"tech_name" => Core::getLanguage()->getItem($t_row["name"]),
							"select_options" => "",
							);
						if($exp_active_levels > 0)
						{
							$is_adv_tech = in_array( $t_row["buildingid"], array(UNIT_LASER_TECH, UNIT_ION_TECH, UNIT_PLASMA_TECH) );
							$min_value = $is_adv_tech ? -$exp_active_levels : 0;
							for($value = $min_value; $value <= $exp_active_levels; $value++)
							{
								$attackLevelsItem["select_options"] .= "<option value='$value' class='center' ".($value == 0 ? 'selected' : '').">".($value < 0 ? $value : "+".$value)."</option>";
							}
						}
						$balanceAttackLevels[] = $attackLevelsItem;
					}
					sqlEnd($t_result);
				}

				Core::getTPL()->assign("exp_points", NS::getUser()->get("be_points"));
				Core::getTPL()->assign("exp_active_levels", min(MAX_ADD_TECH_LEVELS, floor(NS::getUser()->get("be_points") / POINTS_PER_ADD_TECH_LEVEL)));
				Core::getTPL()->assign("max_add_levels", MAX_ADD_TECH_LEVELS);
				Core::getTPL()->assign("points_per_add_level", POINTS_PER_ADD_TECH_LEVEL);

				Core::getTPL()->addLoop("balanceAttackLevels", $balanceAttackLevels);
			}
			else
			{
				// This is needed for correct work of javascript in template. See BagFix#2
				Core::getTPL()->assign("exp_active_levels", 0);
				Core::getTPL()->assign("exp_points", NS::getUser()->get("be_points"));
				Core::getTPL()->assign("max_add_levels", MAX_ADD_TECH_LEVELS);
				Core::getTPL()->assign("points_per_add_level", POINTS_PER_ADD_TECH_LEVEL);
			}

			// Core::getTPL()->assign("fleet", getUnitListStr($data["ships"]));
            if(!$target_valid_by_ally_attack){
                Core::getTPL()->assign("target_invalid_by_ally_attack", true);
                Core::getTPL()->assign("target_check_ally_attack_days", ceil(BLOCK_TARGET_BY_ALLY_ATTACK_USED_PERIOD / (60*60*24)));
            }
            if($bashing_info){
                Core::getTPL()->assign("bashing_hours", ceil(BASHING_PERIOD / (60*60)));
                Core::getTPL()->assign("max_bashing_count", BASHING_MAX_ATTACKS);
                Core::getTPL()->assign("cur_bashing_count", $bashing_info['cur_attack_count']);
            }
			Core::getTPL()->assign("planet_under_attack", $planet_under_attack);
			Core::getTPL()->assign("metal", max(0, NS::getPlanet()->getData("metal")));
			Core::getTPL()->assign("silicon", max(0, NS::getPlanet()->getData("silicon")));
			Core::getTPL()->assign("hydrogen", max(0, NS::getPlanet()->getData("hydrogen") - $consumption));
			Core::getTPL()->assign("capacity", max(0, $data["capacity"] - $consumption));
			Core::getTPL()->assign("rest", fNumber(max(0, $data["capacity"] - $consumption)));
			Core::getTPL()->assign("showHoldingTime", $showHoldingTime);
			Core::getTPL()->addLoop("missions", $missions);
			Core::getTPL()->assign("targetName", $galaxy.":".$system.":".$position." ".$targetName);
			Core::getTPL()->display("missions3");
			exit();
		}
		return $this;
	}

	/**
	 *
	 * Checks if user can jump from one moon to another. BONUS: returns all planets where user can also jump.
	 *
	 * @param integer	Start planet ID
	 * @param integer	End planet ID
	 * @param integer	Owner ID
	 *
	 * @return array
	 */
	protected function canMakeStarJump( $start_planet, $end_planet, $owner_userid = null)
	{
		$users = array();
		if( NS::getUser() )
		{
			$users[] = NS::getUser()->get("userid");
		}
		if( $owner_userid && !in_array($owner_userid, $users) )
		{
			$users[] = $owner_userid;
		}
		if( count($users) == 0 )
		{
			return false;
		}

		$data = array();
		$moons = array();
		$joins	= " LEFT JOIN ".PREFIX."building2planet b2p ON (b2p.planetid = p.planetid)";
		$joins .= " LEFT JOIN ".PREFIX."galaxy g ON (g.moonid = p.planetid)";
		$joins .= " LEFT JOIN ".PREFIX."stargate_jump sj ON (sj.planetid = p.planetid)";
		$select = array("p.planetid", "p.planetname", "g.galaxy", "g.system", "g.position", "sj.time", "b2p.level");
		$result = sqlSelect("planet p",
			$select,
			$joins,
			"p.userid in (".sqlArray($users).") "
				. " AND p.ismoon = '1' "
				. " AND p.planetid != ".sqlVal($start_planet)
				. " AND b2p.buildingid = ".UNIT_STAR_GATE
		);
		while($row = sqlFetch($result))
		{
			$gate_cycle_secs = starGateRecycleSecs($row["level"]);
			if( ($row["time"] < time() - $gate_cycle_secs) || !OXSAR_RELEASED)
			{
				$moons[] = $row;
				$data[] = $row["planetid"];
			}
		}
		sqlEnd($result);
		if ( in_array($end_planet, $data) )
		{
			return $data;
		}
		return false;
	}

	/**
	* This starts the missions and shows a quick overview of the flight.
	*
	* @param integer	Mission type
	* @param integer	Metal
	* @param integer	Silicon
	* @param integer	Hydrogen
	* @param integer	Holding time
	*
	* @return Mission
	*/
	protected function sendFleet($post) // $mode, $metal, $silicon, $hydrogen, $holdingtime)
	{
		if(!NS::isFirstRun("Mission::sendFleet:" . md5(serialize($post)) . "-" . $_SESSION["userid"] ?? 0))
		{
			Logger::dieMessage('TOO_MANY_REQUESTS');
		}
		$ef_rules = $this->setExpoAndFleetRules();
		$mode = max(0, (int)$post["mode"]);
		$metal = max(0, (int)$post["metal"]);
		$silicon = max(0, (int)$post["silicon"]);
		$hydrogen = max(0, (int)$post["hydrogen"]);
		$holdingtime = max(0, (int)$post["holdingtime"]);

		if( $mode == EVENT_EXPEDITION && !EXPEDITION_ENABLED )
        {
			Logger::dieMessage("UNKOWN_MISSION");
        }
        elseif( $mode == EVENT_EXPEDITION && !$ef_rules['expo'] )
		{
			Logger::dieMessage('TOO_MANY_FLEETS_ON_EXPEDITIONS');
		}
		elseif( !(isset($holding_event) || $this->canUseFleetSlot($mode, $ef_rules[$mode != EVENT_EXPEDITION ? 'fleet' : 'expo'])) )
		{
			Logger::dieMessage('TOO_MANY_FLEETS_ON_MISSION');
		}

		if(!isset($holding_event) && NS::getResearch(UNIT_COMPUTER_TECH) + 1 <= NS::getEH()->getUsedFleetSlots())
		{
			Logger::dieMessage('TOO_MANY_FLEETS_ON_MISSION');
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

		$temp = unserialize($row["data"]);
		if(empty($temp["amissions"]) || !is_array($temp["amissions"]) || !in_array($mode, $temp["amissions"]))
		{
			Logger::dieMessage("UNKOWN_MISSION");
		}

		$data["ships"] = $temp["ships"];
		$data["galaxy"] = $temp["galaxy"];
		$data["system"] = $temp["system"];
		$data["position"] = $temp["position"];
		$data["sgalaxy"] = NS::getPlanet()->getData("galaxy");
		$data["ssystem"] = NS::getPlanet()->getData("system");
		$data["sposition"] = NS::getPlanet()->getData("position");
		$data["maxspeed"] = $temp["maxspeed"];
		$data["expeditionMode"] = $temp["expeditionMode"];

		if(in_array($mode, array(EVENT_ATTACK_SINGLE, EVENT_ATTACK_ALLIANCE, EVENT_HALT, EVENT_MOON_DESTRUCTION, 
			EVENT_EXPEDITION, EVENT_ROCKET_ATTACK, EVENT_HOLDING, EVENT_ALLIANCE_ATTACK_ADDITIONAL,
			EVENT_ATTACK_DESTROY_BUILDING, EVENT_ATTACK_ALLIANCE_DESTROY_BUILDING, 
			EVENT_ATTACK_DESTROY_MOON, EVENT_ATTACK_ALLIANCE_DESTROY_MOON)))
		{
			foreach($data["ships"] as $id => $ship){
				if(isset($data["ships"][$id]["quantity"])){
					$data["ships"][$id]["quantity"] = min(MAX_SHIPS, (int)$data["ships"][$id]["quantity"]);
					if(isset($data["ships"][$id]["damaged"])){					
						$data["ships"][$id]["damaged"] = min($data["ships"][$id]["quantity"], (int)$data["ships"][$id]["damaged"]);
					}
				}else{
					// $data["ships"][$id] = min(MAX_SHIPS, (int)$data["ships"][$id]);
				}
			}		
		}
		
		$move_data = self::calcMoveToCoords(
			$data["galaxy"],
			$data["system"],
			$data["position"],
			$data["ships"],
			false,
			1,
			1 // ( $mode == EVENT_STARGATE_TRANSPORT ? STARGATE_TRANSPORT_SPEED : 1 )
			);
		$data['maxspeed'] = $move_data['maxspeed'];
		$data['fleet_size'] = $move_data['fleet_size'];

		foreach($data["ships"] as $id => &$ship)
		{
			if(!isset($move_data["ships"][$id]))
			{
				unset($data["ships"][$id]);
			}
			else if($ship["quantity"] > $move_data["ships"][$id]["quantity"])
			{
				$ship["quantity"] = $move_data["ships"][$id]["quantity"];
			}
			if($move_data["ships"][$id]["mode"] == UNIT_TYPE_ARTEFACT)
			{
				$data["ships"][$id]["art_ids"] = $move_data["ships"][$id]["art_ids"];
			}
		}

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
				$data["consumption"] = 0;
				$distance = 0;
				$metal = 0;
				$silicon = 0;
				$hydrogen = 0;
				$same_pos_halting = true;
			}
		}

		$time = $same_pos_halting ? 1 : NS::getFlyTime($distance, $data["maxspeed"], $temp["speed"]);
		if($mode == EVENT_STARGATE_TRANSPORT) // || $mode == EVENT_TELEPORT_PLANET)
		{
			$time = ceil($time * STARGATE_TRANSPORT_TIME_SCALE);
			$data["consumption"] = ceil($data["consumption"] * STARGATE_TRANSPORT_CONSUMPTION_SCALE);
		}

		// fleet size consumption
		$groupConsumption = unitGroupConsumptionPerHour( UNIT_VIRT_FLEET, $data['fleet_size'], true ) * $time * 2 / (60 * 60);
		$data["consumption"] = max($data["consumption"], $groupConsumption);

		if(NS::getPlanet()->getData("hydrogen") - $data["consumption"] < 0)
		{
			Logger::dieMessage("NOT_ENOUGH_FUEL");
		}

		NS::getPlanet()->setData("hydrogen", NS::getPlanet()->getData("hydrogen") - $data["consumption"]);
		if($temp["capacity"] < $data["consumption"])
		{
			Logger::dieMessage("NOT_ENOUGH_CAPACITY");
		}

		$data["metal"] = abs($metal);
		$data["silicon"] = abs($silicon);
		$data["hydrogen"] = abs($hydrogen);

		if($data["metal"] > NS::getPlanet()->getData("metal"))
		{
			$data["metal"] = NS::getPlanet()->getData("metal");
		}
		if($data["silicon"] > NS::getPlanet()->getData("silicon"))
		{
			$data["silicon"] = NS::getPlanet()->getData("silicon");
		}
		if($data["hydrogen"] > NS::getPlanet()->getData("hydrogen"))
		{
			$data["hydrogen"] = NS::getPlanet()->getData("hydrogen");
		}

		$capa = $temp["capacity"] - $data["consumption"] - $data["metal"] - $data["silicon"] - $data["hydrogen"];
		// Reduce used capacity automatically
		if($capa < 0)
		{
			if($capa + $data["hydrogen"] > 0)
			{
				$data["hydrogen"] -= abs($capa);
			}
			else
			{
				$capa += $data["hydrogen"];
				$data["hydrogen"] = 0;
				if($capa + $data["silicon"] > 0 && $capa < 0)
				{
					$data["silicon"] -= abs($capa);
				}
				else if($capa < 0)
				{
					$capa += $data["silicon"];
					$data["silicon"] = 0;
					if($capa + $data["metal"] && $capa < 0)
					{
						$data["metal"] -= abs($capa);
					}
					else if($capa < 0)
					{
						$data["metal"] = 0;
					}
				}
			}
		}

		$data["capacity"] = $temp["capacity"] - $data["metal"] - $data["silicon"] - $data["hydrogen"];
		if($data["capacity"] < 0)
		{
			Logger::dieMessage("NOT_ENOUGH_CAPACITY");
		}

		// If mission is recycling, get just the capacity of the recyclers.
		if($mode == EVENT_RECYCLING && $data["capacity"] > 0)
		{
			/*
			$_row = sqlSelectRow("ship_datasheet", "capicity", "", "unitid = ".UNIT_RECYCLER); // It is __capacity__ and not capicity
			$recCapa = $_row["capicity"] * $data["ships"][UNIT_RECYCLER]["quantity"];
			$data["capacity"] = max(0, min($data["capacity"], $recCapa));
			*/
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

		$is_stargate_used = false;
		if ( $mode == EVENT_STARGATE_TRANSPORT || $mode == EVENT_STARGATE_JUMP || $mode == EVENT_TELEPORT_PLANET )
		{
			$is_stargate_used = true;
			if( $mode == EVENT_TELEPORT_PLANET || $GLOBALS['STARGATE']['START_DISABLE'] )
			{
				Core::getQuery()->delete("stargate_jump", "planetid = " . sqlVal(Core::getUser()->get("curplanet")));
				Core::getQuery()->insert(
					"stargate_jump",
					array("planetid", "time", "data"),
					array(Core::getUser()->get("curplanet"), time(), serialize($data))
					);
			}
			if( $mode != EVENT_TELEPORT_PLANET && $GLOBALS['STARGATE']['END_DISABLE'] )
			{
				Core::getQuery()->delete("stargate_jump", "planetid = " . sqlVal($temp["destination"]));
				Core::getQuery()->insert(
					"stargate_jump",
					array("planetid", "time", "data"),
					array($temp["destination"] , time(), serialize($data))
					);
			}
		}

		if ($mode == EVENT_EXPEDITION)
		{
			$temp["destination"] = NULL;
			$lvl = NS::getResearch(UNIT_EXPO_TECH);
			$data["expedition_hours"] = $holdingtime = (int)clampVal($holdingtime, 0, $lvl);

			$data["exp_start_time"] = time();

			if(OXSAR_RELEASED)
			{
				$data["exp_duration"] = $data["expedition_hours"] * 3600;
				$data["exp_end_time"] = time() + $time + $data["exp_duration"];
			}
			else
			{
				$data["exp_duration"] = $data["expedition_hours"] * 30;
				$data["exp_end_time"] = time() + 30 + $data["exp_duration"];
			}
			$time = $data["exp_end_time"] - $data["exp_start_time"];
		}
		$data["time"] = $time;

		if(in_array($mode, array(EVENT_ATTACK_SINGLE, EVENT_ATTACK_ALLIANCE, EVENT_ALLIANCE_ATTACK_ADDITIONAL,
				EVENT_ATTACK_DESTROY_BUILDING, EVENT_ATTACK_ALLIANCE_DESTROY_BUILDING, EVENT_HALT)))
		{
			$used_exp_levels = 0;
			$exp_active_levels = min(MAX_ADD_TECH_LEVELS, floor(NS::getUser()->get("be_points") / POINTS_PER_ADD_TECH_LEVEL));

			if( NS::getGalaxyParam($data["galaxy"], "ADVANCED_BATTLE", false) )
			{
				$aval_tech = array(
					UNIT_GUN_TECH, UNIT_SHIELD_TECH,
					UNIT_SHELL_TECH, UNIT_BALLISTICS_TECH, UNIT_MASKING_TECH,
					UNIT_LASER_TECH, UNIT_ION_TECH, UNIT_PLASMA_TECH
				);
			}
			else
			{
				$aval_tech = array( UNIT_GUN_TECH, UNIT_SHIELD_TECH, UNIT_SHELL_TECH, UNIT_BALLISTICS_TECH, UNIT_MASKING_TECH);
			}

			foreach( $aval_tech as $techid )
			{
				$value = (int)$post["attack_lvl_".$techid];

				$is_adv_tech = in_array( $techid, array(UNIT_LASER_TECH, UNIT_ION_TECH, UNIT_PLASMA_TECH) );
				$min_value = $is_adv_tech ? -$exp_active_levels : 0;
				$value = clampVal($value, $min_value, $exp_active_levels);

				$exp_active_levels -= abs($value);
				$used_exp_levels += abs($value);

				$data["add_tech_".$techid] = $value;
			}
			$data["used_exp_points"] = $used_exp_levels * POINTS_PER_ADD_TECH_LEVEL;
		}

		// are they the same?
		// Core::getQuery()->delete("temp_fleet", "planetid = ".sqlPlanet());
		// sqlDelete('temp_fleet', "planetid = ".sqlVal($curr_planet_id));

		NS::getEH()->addEvent($mode, time() + $time, NS::getUser()->get("curplanet"), NS::getUser()->get("userid"), $temp["destination"], $data);

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

			if($value["mode"] == UNIT_TYPE_ARTEFACT)
			{
				foreach($value["art_ids"] as $art)
				{
					$activate_auto = true;
					if($key == ARTEFACT_IGLA_MORI)
					{
						if($is_stargate_used)
						{
							$is_stargate_used = false;
							$activate_auto = true;
						}
						else
						{
							$activate_auto = false;
						}
					}
					Artefact::attachToFleet($art["artid"], NS::getUser()->get("userid"), NS::getUser()->get("curplanet"), $activate_auto);
				}
			}
		}
		Core::getTPL()->addLoop("fleet", $fleet);
		Core::getTPL()->display("missions4");
		return $this;
	}

	/**
	* Retreats a fleet back to its origin planet.
	*
	* @param integer	Event id to retreat
	*
	* @return Mission
	*/
	protected function retreatFleet($id)
	{
		if( NS::retreatFleet($id) == RETREAT_FLEET_PLANET_UNDER_ATTACK )
		{
			Logger::dieMessage('PLANET_UNDER_ATTACK');
		}
		doHeaderRedirection("game.php/Mission", false);
		return $this;
	}

	/**
	* Invite friends for alliance attack.
	*
	* @param integer	Event id
	*
	* @return Mission
	*/
	protected function formation($eventid)
	{
		$row = sqlSelectRow("events", array("mode", "time"), "",
			"mode in (".sqlArray(EVENT_ATTACK_SINGLE, EVENT_ATTACK_ALLIANCE,
				EVENT_ATTACK_DESTROY_BUILDING, EVENT_ATTACK_ALLIANCE_DESTROY_BUILDING,
				EVENT_ATTACK_DESTROY_MOON, EVENT_ATTACK_ALLIANCE_DESTROY_MOON).") "
			." AND user = ".sqlUser()." AND eventid = ".sqlVal($eventid));
		if($row)
		{
			$invitation = array();
			$joins	= "LEFT JOIN ".PREFIX."formation_invitation fi ON fi.eventid = af.eventid ";
			$joins .= "LEFT JOIN ".PREFIX."user u ON u.userid = fi.userid ";
			$result = sqlSelect("attack_formation af", array("af.name", "fi.userid", "u.username"), $joins, "af.eventid = ".sqlVal($eventid));
			while($_row = sqlFetch($result))
			{
				$name = $_row["name"];
				$invitation[] = $_row;
			}
			sqlEnd($result);

			if(count($invitation) <= 0)
			{
				$attack_type = array(
							EVENT_ATTACK_SINGLE => EVENT_ATTACK_ALLIANCE,
							EVENT_ATTACK_DESTROY_BUILDING => EVENT_ATTACK_ALLIANCE_DESTROY_BUILDING,
							EVENT_ATTACK_DESTROY_MOON => EVENT_ATTACK_ALLIANCE_DESTROY_MOON,
							);
				if(isset($attack_type[$row["mode"]]))
				{
					sqlUpdate("events", array(
						"mode" => $attack_type[$row["mode"]],
						"parent_eventid" => $eventid
						), "eventid = ".sqlVal($eventid)
							. ' ORDER BY eventid');
				}
				sqlInsert("attack_formation", array("eventid" => $eventid, "time" => $row["time"]));
				$name = $eventid;

				sqlInsert("formation_invitation", array("eventid" => $eventid, "userid" => NS::getUser()->get("userid")));

				$invitation[0]["userid"] = NS::getUser()->get("userid");
				$invitation[0]["username"] = NS::getUser()->get("username");
			}
			Core::getTPL()->addLoop("invitation", $invitation);
			Core::getTPL()->assign("formationName", $name);

			Core::getTPL()->display("alliance_attack");
		}
		return $this;
	}

	/**
	* Executes an invitation.
	*
	* @param integer	Event id
	* @param string	Formation name
	* @param string	Invited username
	*
	* @return Mission
	*/
	protected function invite($eventid, $name, $username)
	{
		require_once(APP_ROOT_DIR."game/Relation.class.php");

		$row = sqlSelectRow("events", "time", "",
			"mode in (".sqlArray(EVENT_ATTACK_SINGLE, EVENT_ATTACK_ALLIANCE, EVENT_ATTACK_DESTROY_BUILDING, EVENT_ATTACK_ALLIANCE_DESTROY_BUILDING).") AND user = ".sqlUser()
			. " AND eventid = ".sqlVal($eventid));
		if($row)
		{
			$error = "";
			$time = $row["time"];
			$username = trim($username);
			if($username)
			{
				$row = sqlSelectRow("user u", array("u.userid", "u2a.aid"), "LEFT JOIN ".PREFIX."user2ally u2a ON (u2a.userid = u.userid)", "u.username = ".sqlVal($username));
				$userid = $row["userid"];
				$aid = $row["aid"];
				$Relation = new Relation(NS::getUser()->get("userid"), NS::getUser()->get("aid"));
				if(!$Relation->hasRelation($userid, $aid))
				{
					$error[] = "UNABLE_TO_INVITE_USER";
				}
				unset($Relation);
			}

			$name = trim($name);
			if(Str::length($name) > 0 && Str::length($name) <= 128)
			{
				$name = Str::validateXHTML($name);
				sqlUpdate("attack_formation", array("name" => $name), "eventid = ".sqlVal($eventid) . ' ORDER BY eventid');
			}
			else
			{
				$error[] = "ENTER_FORMATION_NAME";
			}
			if(empty($error))
			{
				if($userid)
				{
					sqlInsert("formation_invitation", array("eventid" => $eventid, "userid" => $userid));
				}
			}
			else
			{
				foreach($error as $error)
				{
					Logger::addMessage($error);
				}
			}
		}
		$this->formation($eventid);
		return $this;
	}

	/**
	* Executes star gate jump and update shipyard informations.
	*
	* @param integer	Target moon id
	*
	* @return Mission
	*/
	protected function executeJump($moonid)
	{
		return $this->index();
	}

	/**
	* Select the ships for jump.
	*
	* @param array		Ships for jump
	*
	* @return Mission
	*/
	protected function starGateJump($ships)
	{
		return $this->index();
	}

	/**
	* Checks the ships for an attack.
	*
	* @param array		All ships
	*
	* @return boolean
	*/
	protected function canDestroyAttack($ships)
	{
		if(isset($ships[UNIT_DEATH_STAR]["quantity"]) && $ships[UNIT_DEATH_STAR]["quantity"] >= 1)
		{
			return true;
		}
		return false;
	}

	protected function getRealShips($ships)
	{
		$real_ships = array();
		foreach((array)$ships as $shipid => $ship)
		{
			if(is_array($ship) && isset($ship["quantity"]) && $ship["quantity"] >= 1 && !isset($ship["art_ids"]))
			{
				$real_ships[$shipid] = $ship;
			}
		}
		return $real_ships;
	}

	protected function isNeutronAffectorFound($ships)
	{
		$ships = (array)$ships;
		return isset($ships[ARTEFACT_BATTLE_NEUTRON_AFFECTOR]);
	}

	protected function canTeleportCurPlanet($ships)
	{
		$ships = (array)$ships;
		if(isset($ships[ARTEFACT_PLANET_TELEPORTER]))
		{
			$ships = $this->getRealShips($ships);
			if(count($ships) > 0 && NS::getPlanet()->isMoon() && NS::getPlanet()->getBuilding(UNIT_STAR_GATE) > 0)
			{
				$stargate_time = NS::getPlanet()->getStarGateTime();
				$stargate_level = NS::getPlanet()->getBuilding(UNIT_STAR_GATE);
				$gate_cycle_secs = starGateRecycleSecs($stargate_level);
				if(!$stargate_time || $stargate_time < time() - $gate_cycle_secs || !OXSAR_RELEASED)
				{
					return NS::getEH()->canTeleportPlanet();
				}
			}
		}
		return false;
	}

	/**
	* Checks the ships for an attack.
	*
	* @param array		All ships
	*
	* @return boolean
	*/
	protected function canAttack($ships)
	{
		$ships = $this->getRealShips($ships);
		if(ATTACK_BY_ESPIONAGE_UNIT_ENABLED)
		{
			return count($ships) > 0;
		}
		if(count($ships) >= 2)
		{
			return true;
		}
		if(isset($ships[UNIT_ESPIONAGE_SENSOR]))
		{
			return false;
		}
		return true;
	}

	/**
	* Checks normal fleet
	*
	* @param array		All ships
	*
	* @return boolean
	*/
	protected function isNormalFleet($ships)
	{
		$ships = $this->getRealShips($ships);
		if(count($ships) >= 2) // maybe UNIT_ESPIONAGE_SENSOR but something alse
		{
			return true;
		}
		// count == 1 || count == 0
		return !isset($ships[UNIT_ESPIONAGE_SENSOR]);
	}

	/**
	* Checks the ships for an expedition.
	*
	* @param array		All ships
	*
	* @return boolean
	*/
	protected function canExpedition($ships, $expo_slot = true)
	{
		return ($this->isNormalFleet($ships) && $expo_slot);
	}

	/**
	* Checks the ships for an recycle.
	*
	* @param array		All ships
	*
	* @return boolean
	*/
	protected function canRecycle($ships)
	{
		$ships = $this->getRealShips($ships);

		foreach($GLOBALS["RECYCLER_UNITS"] as $recycler_unit_id)
		{
			if(isset($ships[$recycler_unit_id]))
			{
				return true;
			}
		}
		return false;
	}

	/**
	* Checks the ships for an halt.
	*
	* @param array		All ships
	*
	* @return boolean
	*/
	protected function canHalt($ships)
	{
		$ships = $this->getRealShips($ships);
		return count($ships) > 0; // $this->isNormalFleet($ships);
	}

	/**
	* Checks the ships for an espioange action.
	*
	* @param array		All ships
	*
	* @return boolean
	*/
	protected function canSpy($ships)
	{
		$ships = $this->getRealShips($ships);
		if(count($ships) > 1 || !isset($ships[UNIT_ESPIONAGE_SENSOR]) || $ships[UNIT_ESPIONAGE_SENSOR]["quantity"] < 1)
		{
			return false;
		}
		return true;
	}

	/**
	* Returns event id and time.
	*
	* @param integer	Destination planet id
	*
	* @return array	Event id and time
	*/
	protected function getFormations($planetid)
	{
		$joins	= "LEFT JOIN ".PREFIX."events e ON (e.eventid = fi.eventid)";
		$joins .= "LEFT JOIN ".PREFIX."attack_formation af ON (e.eventid = af.eventid)";
		$select = array("af.eventid", "af.time", "e.data");
		$row = sqlSelectRow("formation_invitation fi", $select, $joins,
			"fi.userid = ".sqlUser()." AND af.time > ".sqlVal(time())." AND e.destination = ".sqlVal($planetid)
			. " AND e.processed=".EVENT_PROCESSED_WAIT
			. " ORDER BY e.eventid DESC"
		);
		if($row){
			$row['data'] = unserialize($row['data']);
		}
		return $row;
	}
}
?>