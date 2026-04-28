<?php
/**
* Generates auto-report after a fleet event.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class AutoMsg
{
	/**
	* Mode id. Identifies what message should be sent.
	*
	* @var integer
	*/
	protected $mode;

	/**
	* Target user.
	*
	* @var integer
	*/
	protected $userid;

	/**
	* Time when message has actually been sent.
	*
	* @var integer
	*/
	protected $time = 0;

	/**
	* The datafield holds the fleet informations.
	*
	* @var array
	*/
	protected $data = array();

	/**
	* Start the message generator.
	*
	* @param integer
	* @param integer
	* @param integer
	* @param array
	*
	* @return void
	*/
	public function __construct($mode, $userid, $time, $data)
	{
		$this->mode = $mode;
		$this->userid = $userid;
		$this->time = $time;
		$this->data = $data;
		Core::pushLanguage();
		Core::selectLanguage(NS::getUserLanguageId($userid));
		Core::getLanguage()->load(array("AutoMessages", "info"));
		// Hook::event("AUTO_MSG_START", array(&$this));
		switch($this->mode)
		{
			case MSG_CREDIT:
				$this->folder = MSG_FOLDER_CREDIT;
				$this->creditReport();
				break;
			case MSG_POSITION_REPORT:
			case MSG_DELIVERY_UNITS:
			case MSG_STARGATE_JUMP_REPORT:
				$this->folder = MSG_FOLDER_FLEET;
				$this->positionReport();
				break;
			case MSG_TRANSPORT_REPORT:
				$this->folder = MSG_FOLDER_FLEET;
				$this->transportReport();
				break;
			case MSG_COLONIZE_REPORT:
				$this->folder = MSG_FOLDER_FLEET;
				$this->colonizeReport();
				break;
			case MSG_RECYCLING_REPORT:
				$this->folder = MSG_FOLDER_RECYCLER;
				$this->recyclingReport();
				break;
			case MSG_ESPIONAGE_COMMITTED:
				$this->folder = MSG_FOLDER_FLEET;
				$this->espionageCommitted();
				break;
			case MSG_RETURN_REPORT:
				$this->folder = MSG_FOLDER_FLEET;
				$this->returnReport();
				break;
			case MSG_TRANSPORT_REPORT_OTHER:
			case MSG_DELIVERY_RESOURSES:
				$this->folder = MSG_FOLDER_FLEET;
				$this->transportReport_other();
				break;
			case MSG_ASTEROID:
				$this->folder = MSG_FOLDER_FLEET;
				$this->asteroid();
				break;
			case MSG_ALLY_ABANDONED:
				$this->folder = MSG_FOLDER_ALLIANCE;
				$this->allyAbandoned();
				break;
			case MSG_MEMBER_RECEIPTED:
				$this->folder = MSG_FOLDER_ALLIANCE;
				$this->memberReceipted();
				break;
			case MSG_MEMBER_REFUSED:
				$this->folder = MSG_FOLDER_ALLIANCE;
				$this->memberRefused();
				break;
			case MSG_NEW_MEMBER:
				$this->folder = MSG_FOLDER_ALLIANCE;
				$this->newMember();
				break;
			case MSG_MEMBER_LEFT:
				$this->folder = MSG_FOLDER_ALLIANCE;
				$this->memberLeft();
				break;
			case MSG_MEMBER_KICKED:
				$this->folder = MSG_FOLDER_ALLIANCE;
				$this->memberKicked();
				break;
			case MSG_EXPEDITION_REPORT:
				$this->folder = MSG_FOLDER_EXPEDITION;
				$this->expeditionReport();
				break;
			case MSG_EXPEDITION_SENSOR:
				$this->folder = MSG_FOLDER_EXPEDITION;
				$this->expeditionSensor();
				break;

			case MSG_GRASPED_REPORT:
				$this->folder = MSG_FOLDER_FLEET;
				$this->graspedReport();
				break;

			case MSG_RETREAT_OTHER:
				$this->folder = MSG_FOLDER_FLEET;
				$this->retreatOther();
				break;

			case MSG_RETREAT_TRANSPORT:
				$this->folder = MSG_FOLDER_FLEET;
				$this->retreatTransport();
				break;

			case MSG_BUILDING_DESTROYED:
				$this->folder = MSG_FOLDER_FLEET;
				$this->buildingDestroyed();
				break;

			case MSG_LOST_ARTEFACTS:
				$this->folder = MSG_FOLDER_FLEET;
				$this->lostArtefacts();
				break;

			case MSG_GRASPED_ARTEFACTS:
				$this->folder = MSG_FOLDER_FLEET;
				$this->graspedArtefactsReport();
				break;

			case MSG_MOON_DESTROYED:
				$this->folder = MSG_FOLDER_FLEET;
				$this->moonDestroyed();
				break;

			case MSG_TRANSPORT_REPORT_ARTEFACT:
				$this->folder = MSG_FOLDER_FLEET;
				$this->transportReport_artefact();
				break;

			case MSG_ARTEFACT:
				$this->folder = MSG_FOLDER_ARTEFACTS;
				$this->artefactReport();
				break;

			case MSG_EXPEDITION_NEW_PLANET:
				$this->folder = MSG_FOLDER_EXPEDITION;
				$this->expeditionNewPlanetReport();
				break;

			case MSG_UFO_PLANET_DIE:
				$this->folder = MSG_FOLDER_FLEET;
				$this->expeditionPlanetDie();
				break;

			case MSG_TEMP_PLANET_DIE:
				$this->folder = MSG_FOLDER_FLEET;
				$this->tempPlanetDie();
				break;

			case MSG_SURVEILLANCE_DETECTED:
				$this->folder = MSG_FOLDER_SURVEILLANCE_DETECTED;
				$this->surveillanceDetected();
				break;

			case MSG_ALIEN_HALTING:
				$this->folder = MSG_FOLDER_FLEET;
				$this->alienHalting();
				break;

			case MSG_EXCH_LOT_BACK_RESOURSES:
				$this->folder = MSG_FOLDER_FLEET;
				$this->exchangeLotBackResourses();
				break;

			case MSG_ALIEN_RESOURSES_GIFT:
				$this->folder = MSG_FOLDER_FLEET;
				$this->alienResoursesGift();
				break;

			case MSG_PLANET_TELEPORTED:
				$this->folder = MSG_FOLDER_FLEET;
				$this->planetTeleported();
				break;

			case MSG_PLANET_NOT_TELEPORTED:
				$this->folder = MSG_FOLDER_FLEET;
				$this->planetNotTeleported();
				break;
		}
		Core::popLanguage();
	}

	protected function planetNotTeleported()
	{
		$planet_new = getCoordLink($this->data["galaxy"], $this->data["system"], $this->data["position"], true, $this->data["planetid"], true);
		$planet_old = getCoordLink($this->data["sgalaxy"], $this->data["ssystem"], $this->data["sposition"], true, $this->data["planetid"], false);

		$msg = Core::getLang()->getItemWith(
			'MSG_PLANET_NOT_TELEPORTED',
			array(
				'planet_new' => $planet_new,
				'planet_old' => $planet_old,
			)
		);
		$subject = Core::getLanguage()->getItem('MSG_PLANET_NOT_TELEPORTED_SUBJ');
		$this->sendMsg($subject, $msg);
		return $this;
	}

	protected function planetTeleported()
	{
		$planet_new = getCoordLink($this->data["galaxy"], $this->data["system"], $this->data["position"], true, $this->data["planetid"], true);
		$planet_old = getCoordLink($this->data["sgalaxy"], $this->data["ssystem"], $this->data["sposition"], true, $this->data["planetid"], false);

		$msg = Core::getLang()->getItemWith(
			'MSG_PLANET_TELEPORTED',
			array(
				'planet_new' => $planet_new,
				'planet_old' => $planet_old,
			)
		);
		$subject = Core::getLanguage()->getItem('MSG_PLANET_TELEPORTED_SUBJ');
		$this->sendMsg($subject, $msg);
		return $this;
	}

	protected function alienResoursesGift()
	{
		$metal = fNumber($this->data["metal"]);
		$silicon = fNumber($this->data["silicon"]);
		$hydrogen = fNumber($this->data["hydrogen"]);
		$planet = getCoordLink(null, null, null, true, $this->data["planetid"], true);

		$msg = Core::getLang()->getItemWith(
			'MSG_ALIEN_RESOURSES_GIFT',
			array(
				'planet' => $planet,
				'metal' => $metal,
				'silicon' => $silicon,
				'hydrogen' => $hydrogen,
			)
		);
		$subject = Core::getLanguage()->getItem('MSG_ALIEN_RESOURSES_GIFT_SUBJ');
		$this->sendMsg($subject, $msg);
		return $this;
	}

	protected function exchangeLotBackResourses()
	{
		$metal = fNumber($this->data["metal"]);
		$silicon = fNumber($this->data["silicon"]);
		$hydrogen = fNumber($this->data["hydrogen"]);
		$planet = getCoordLink(null, null, null, true, $this->data["planetid"], true);

		$msg = Core::getLang()->getItemWith(
			'MSG_EXCH_LOT_BACK_RESOURSES',
			array(
				'planet' => $planet,
				'metal' => $metal,
				'silicon' => $silicon,
				'hydrogen' => $hydrogen,
				'lot' => $this->data["lot"],
			)
		);
		$subject = Core::getLanguage()->getItem('MSG_EXCH_LOT_BACK_RESOURSES_SUBJ');
		$this->sendMsg($subject, $msg);
		return $this;
	}

	protected function alienHalting()
	{
		$coordinates = getCoordLink($this->data["galaxy"], $this->data["system"], $this->data["position"], true, $this->data["planetid"], true);
		$msg = Core::getLang()->getItemWith(
			'ALIEN_HALTING_MSG',
			array(
				'coords_planet' => $coordinates,
			)
		);
		$subject = Core::getLanguage()->getItem('ALIEN_HALTING_SUBJ');
		$this->sendMsg($subject, $msg);
		return $this;
	}

	protected function surveillanceDetected()
	{
		$coordinates = getCoordLink($this->data["galaxy"], $this->data["system"], $this->data["position"], true, $this->data["planetid"], true);
		// $coordinates = "[{$this->data['galaxy']}:{$this->data['system']}:{$this->data['position']}]";
		$msg = Core::getLang()->getItemWith(
			'SURVEILLANCE_DETECTED',
			array(
				'coords_planet' => $coordinates,
			)
		);
		$subject = Core::getLanguage()->getItem('SURVEILLANCE_DETECTED_SUBJ');
		$this->sendMsg($subject, $msg);
		return $this;
	}

	protected function tempPlanetDie()
	{
		$coordinates = "[{$this->data['galaxy']}:{$this->data['system']}:{$this->data['position']}]";
		$msg = Core::getLang()->getItemWith(
			'TEMP_PLANET_DIE',
			array(
				'planet'	=> $coordinates,
			)
		);
		$subject = Core::getLanguage()->getItem('TEMP_PLANET_DIE_SUBJ');
		$this->sendMsg($subject, $msg);
		return $this;
	}

	protected function expeditionPlanetDie()
	{
		$coordinates = "[{$this->data['galaxy']}:{$this->data['system']}:{$this->data['position']}]";
		$msg = Core::getLang()->getItemWith(
			'NEUTRAL_PLANET_DIE',
			array(
				'planet'	=> $coordinates,
			)
		);
		$subject = Core::getLanguage()->getItem('EXPEDITION_NEUTRAL_PLANET_SUBJ');
		$this->sendMsg($subject, $msg);
		return $this;
	}

	protected function expeditionNewPlanetReport()
	{
		$coordinates = $this->data['name'].' ' .Link::get(
			"game.php/go:Mission/g:".$this->data["galaxy"]."/s:".$this->data["system"]."/p:".$this->data["position"],
			"[{$this->data['galaxy']}:{$this->data['system']}:{$this->data['position']}]"
		);
		$msg = Core::getLang()->getItemWith(
			'EXPEDITION_NEUTRAL_PLANET',
			array(
				'planet'	=> $coordinates,
				'm_scraps'	=> fNumber($this->data['m_scraps']),
				's_scraps'	=> fNumber($this->data['s_scraps']),
			)
		);
		$subject = Core::getLanguage()->getItem('EXPEDITION_NEUTRAL_PLANET_SUBJ');
		$this->sendMsg($subject, $msg);
		return $this;
	}

	/**
	* When a fleet reached a planet to position.
	*
	* @return AutoMsg
	*/
	protected function positionReport()
	{
		switch($this->mode)
		{
		case MSG_DELIVERY_UNITS:
			$msgId = "DELIVERY_UNITS_REPORT";
			$subjId = "DELIVERY_UNITS_REPORT_SUBJECT";
			break;

		case MSG_STARGATE_JUMP_REPORT:
			$msgId = "STARGATE_JUMP_REPORT";
			$subjId = "STARGATE_JUMP_REPORT_SUBJECT";
			break;

		default:
			$msgId = "POSITION_REPORT";
			$subjId = "POSITION_REPORT_SUBJECT";
			break;
		}
		$msg = Core::getLanguage()->getItem($msgId);

		// Get the important data
		$metal = fNumber($this->data["metal"]);
		$silicon = fNumber($this->data["silicon"]);
		$hydrogen = fNumber($this->data["hydrogen"]);
		$rShips = $this->getShipsString();
		$row = sqlSelectRow("planet", array("planetname", "userid"), "", "planetid = ".sqlVal($this->data["destination"]));
		$planet = $row["planetname"];
		$coordinates = getCoordLink($this->data["galaxy"], $this->data["system"], $this->data["position"], true, $this->data["destination"]);
		$msg = sprintf($msg, $rShips, $planet, $coordinates, $metal, $silicon, $hydrogen);
		// Hook::event("AUTO_MSG_POSITION_REPORT", array(&$this, &$msg));

		$rel_userid = $this->userid = $row["userid"]; // sqlSelectRow("planet", "userid", "", "planetname=".sqlVal($planet));
		$this->sendMsg(Core::getLanguage()->getItem($subjId), $msg, $rel_userid);
		return $this;
	}

	/**
	* Message after transport to own planet has completed.
	*
	* @return AutoMsg
	*/
	protected function transportReport()
	{
		$msg = Core::getLanguage()->getItem("TRANSPORT_ACCOMPLISHED");

		// Get the important data
		$metal = fNumber($this->data["metal"]);
		$silicon = fNumber($this->data["silicon"]);
		$hydrogen = fNumber($this->data["hydrogen"]);
		$planet = $this->data["targetplanet"];
		$rel_userid = $this->data["targetuser"]; // sqlSelectRow("planet", "userid", "", "planetname=".sqlVal($this->data["targetplanet"]));
		$coordinates = getCoordLink($this->data["galaxy"], $this->data["system"], $this->data["position"], true, $this->data["destination"]);
		$msg = sprintf($msg, $planet, $coordinates, $metal, $silicon, $hydrogen);
		// Hook::event("AUTO_MSG_TRANSPORT_REPORT", array(&$this, &$msg));
		$this->sendMsg(Core::getLanguage()->getItem("TRANPORT_SUBJECT"), $msg, $rel_userid);
		return $this;
	}

	/**
	* Message after transport to another has completed.
	*
	* @return AutoMsg
	*/
	protected function transportReport_other()
	{
		$msg = Core::getLanguage()->getItem("TRANSPORT_ACCOMPLISHED_TO_OTHER");

		$metal = fNumber($this->data["metal"]);
		$silicon = fNumber($this->data["silicon"]);
		$hydrogen = fNumber($this->data["hydrogen"]);
		$planet = $this->data["startplanet"];
		$planetCoords = getCoordLink($this->data["sgalaxy"], $this->data["ssystem"], $this->data["sposition"], true, $this->data["planetid"]);
		$tplanet = $this->data["targetplanet"];
		$tplanetCoords = getCoordLink($this->data["galaxy"], $this->data["system"], $this->data["position"], true, $this->data["destination"]);

		$tuser = $this->data["targetuser"]; // sqlSelectRow("planet", "userid", "", "planetname=".sqlVal($this->data["targetplanet"]));
		$msg = sprintf($msg, $planet, $planetCoords, $tplanet, $tplanetCoords, $metal, $silicon, $hydrogen);
		$this->sendMsg(Core::getLanguage()->getItem("TRANPORT_SUBJECT"), $msg, $tuser);

		$this->userid = $this->data["targetuser"];
		$msg = Core::getLanguage()->getItem("TRANSPORT_ACCOMPLISHED_OTHER");
		$user = $this->data["startuser"];

		$msg = sprintf($msg, $planet, $planetCoords, $user, $tplanet, $tplanetCoords, $metal, $silicon, $hydrogen);
		// Hook::event("AUTO_MSG_TRANSPORT_REPORT_BY_OTHER", array(&$this, &$msg));
		$suser = sqlSelectRow("user", "userid", "", "username=".sqlVal($this->data["startuser"]));
		$this->sendMsg(Core::getLanguage()->getItem("TRANPORT_SUBJECT"), $msg, $suser["userid"]);
		return $this;
	}

	/**
	* Message after transport to another has completed.
	*
	* @return AutoMsg
	*/
	protected function transportReport_artefact()
	{
		$this->userid = $this->data["targetuser"];
		$msg = Core::getLanguage()->getItem("TRANSPORT_ACCOMPLISHED_ARTEFACT");
		$temp = $this->getArtefactNameAndPosition($this->data['artefact']);
		$msg = sprintf($msg, $temp['name']);
		$this->sendMsg( Core::getLanguage()->getItem("TRANPORT_SUBJECT"), $msg );
		return $this;
	}

	/**
	* Generates a string for all ships.
	*
	* @return string	The ships
	*/
	protected function getShipsString($ships = null)
	{
		// Core::getLanguage()->load("info");
		return getUnitListStr(!is_null($ships) ? $ships : $this->data["ships"]);
	}

	protected function expeditionReport()
	{
		switch ($this->data["exp_type"])
		{
		case 'resourceDiscovery':
		case 'asteroidDiscovery':
			if($this->data["exp_type"] == 'resourceDiscovery')
			{
				$name = "EXPEDITION_RESOURCE_DISCOVERY_" . mt_rand(1, 2);
			}
			else
			{
				$is_recycler_units_exist = false;
				foreach($GLOBALS["RECYCLER_UNITS"] as $recycler_unit_id)
				{
					if(isset($this->data["ships"][$recycler_unit_id]))
					{
						$is_recycler_units_exist = true;
						break;
					}
				}
				if($is_recycler_units_exist) // isset($this->data["ships"][UNIT_RECYCLER]))
				{
					$name = "EXPEDITION_ASTEROID_DISCOVERY_" . mt_rand(1, 2);
				}
				else
				{
					$name = "EXPEDITION_ASTEROID_DISCOVERY_NO_CARGO";
				}
			}
			$msg = Core::getLanguage()->getItemWith($name, array(
				"metal" => fNumber($this->data["recycledmetal"]),
				"silicon" => fNumber($this->data["recycledsilicon"]),
				"hydrogen" => fNumber($this->data["recycledhydrogen"]),
				"debrismetal" => fNumber($this->data["debrismetal"]),
				"debrissilicon" => fNumber($this->data["debrissilicon"]),
				"debrishydrogen" => fNumber($this->data["debrishydrogen"]),
				));
			break;

		case 'visitePlanetDiscovery':
			$name = 'NEUTRAL_PLANET_FOUND';
			$coordinates = Link::get(
				"game.php/go:Mission/g:".
					$this->data['planet_found']['galaxy'].
					"/s:".$this->data['planet_found']['system'].
					"/p:".$this->data['planet_found']['position'],
				"[{$this->data['planet_found']['galaxy']}:{$this->data['planet_found']['system']}:{$this->data['planet_found']['position']}]"
			);
			$msg = Core::getLanguage()->getItemWith(
				$name,
				array(
					"planet" => $coordinates,
					"metal" => fNumber($this->data['planet_found']['metal']),
					"silicon" => fNumber($this->data['planet_found']['silicon']),
				)
			);
			break;

		case 'shipsDiscovery':
		case 'battlefieldDiscovery':
			if($this->data["exp_type"] == 'shipsDiscovery')
			{
				$name = "EXPEDITION_SHIPS_DISCOVERY";
			}
			else
			{
				$name = "EXPEDITION_BATTLEFIELD_DISCOVERY";
			}
			$new_ships_str = $this->getShipsString($this->data["newShips"]);
			$msg = Core::getLanguage()->getItemWith($name, array(
				"new_ships" => $new_ships_str
				));
			break;

		case 'delayReturn':
		case 'fastReturn':
			if($this->data["exp_type"] == 'delayReturn')
			{
				$name = "EXPEDITION_DELAY_RETURN";
			}
			else
			{
				$name = "EXPEDITION_FAST_RETURN_" . mt_rand(1, 3);
			}
			$msg = Core::getLanguage()->getItemWith($name, array(
				"percent" => round($this->data["time_k"] * 100)
				));
			break;

		case 'creditDiscovery':
			$name = "EXPEDITION_CREDITS_DISCOVERY_" . mt_rand(1, 6);
			$msg = Core::getLanguage()->getItemWith($name, array(
				"credits" => fNumber($this->data["credit"]),
				"anabios_years" => fNumber(mt_rand(12, 123)*10 + mt_rand(5, 9)),
				));
			break;

		case 'nothing':
			$name = "EXPEDITION_NOTHING_" . mt_rand(1, 7);
			$msg = Core::getLanguage()->getItem($name);
			break;

		case 'xSkirmish':
			$msg = Core::getLanguage()->getItem("EXPEDITION_XSKIRMISH");
			break;

		case 'expeditionLost':
			$name = "EXPEDITION_LOST_" . mt_rand(1, 2);
			$msg = Core::getLanguage()->getItem($name);
			break;

		case 'artefactDiscovery':
			$name = "EXPEDITION_ARTEFACT";
			$temp = getArtefactNameAndPosition($this->data['art_id_desc']);
			$msg = Core::getLanguage()->getItemWith($name, array(
				"art_name" => $temp["name"]
				));
			break;

		default:
			$msg = "ERR: ".$this->data["exp_type"];
			break;
		}

		//$msg = sprintf($msg, $this->getShipsString(), $metal, $silicon, $hydrogen);
		// Hook::event("AUTO_MSG_EXPEDITION_REPORT", array(&$this, &$msg));
		$this->sendMsg(Core::getLanguage()->getItem("EXPEDITION_SUBJECT"), $msg);
		return $this;
	}

	protected function expeditionSensor()
	{
		if($this->data["times_visited"] >= 15)
		{
			$name = "EXPEDITION_SENSOR_POOR_REPORT";
		}
		else if($this->data["times_visited"] > 0)
		{
			$name = "EXPEDITION_SENSOR_REPORT";
		}
		else
		{
			$name = "EXPEDITION_SENSOR_CLEAR_REPORT";
		}
		$msg = Core::getLanguage()->getItemWith($name, array(
			"days" => $this->data["visit_days"],
			"visits" => $this->data["times_visited"],
			"exp_tech" => $this->data["exp_tech"],
			"spy_tech" => $this->data["spy_tech"],
			"spy_units" => fNumber($this->data["spy_units"]),
			"exp_power" => fNumber($this->data["exp_power"], 1),
			"exp_hours" => fNumber($this->data["exp_hours"]),
			"exp_percent" => fNumber($this->data["exp_percent"], 1),
			"exp_sector" => getCoordLink($this->data["galaxy"], $this->data["system"], EXPED_PLANET_POSITION, true),
			"p_resourceDiscovery" => round($this->data["types_percent"]["resourceDiscovery"]),
			"p_asteroidDiscovery" => round($this->data["types_percent"]["asteroidDiscovery"]),
			"p_shipsDiscovery" => round($this->data["types_percent"]["shipsDiscovery"]),
			"p_battlefieldDiscovery" => round($this->data["types_percent"]["battlefieldDiscovery"]),
			"p_nothing" => round($this->data["types_percent"]["nothing"]),
			"p_delayReturn" => round($this->data["types_percent"]["delayReturn"]),
			"p_fastReturn" => round($this->data["types_percent"]["fastReturn"]),
			"p_expeditionLost" => round($this->data["types_percent"]["expeditionLost"]),
			"p_artefactDiscovery" => round($this->data["types_percent"]["artefactDiscovery"]),
			"p_creditDiscovery" => round($this->data["types_percent"]["creditDiscovery"]),
			"p_xSkirmish" => round($this->data["types_percent"]["xSkirmish"]),
			));

		// Hook::event("AUTO_MSG_EXPEDITION_SENSOR", array(&$this, &$msg));
		$this->sendMsg(Core::getLanguage()->getItem("EXPEDITION_SENSOR_SUBJECT"), $msg);
		return $this;
	}

	protected function lostArtefacts()
	{
		// � �������� �� ������� {@coords_planet} ��������� �������� ���� ���������: {@artefacts}.

		$row = getCoordsAndPlanetname($this->data["planetid"]);
		$planet = $row["planetname"];
		$coords = getCoordLink($row["galaxy"], $row["system"], $row["position"], true, $this->data["planetid"]);

		$artefacts_str = $this->getShipsString();

		$msg = Core::getLanguage()->getItemWith("LOST_ARTEFACTS_MSG", array(
			"coords_planet" => $coords,
			"artefacts" => $artefacts_str,
			));

		$rel_userid["userid"] = null;
		$this->sendMsg(Core::getLanguage()->getItem("LOST_ARTEFACTS_SUBJECT"), $msg, $rel_userid["userid"]);

		return $this;
	}

	protected function buildingDestroyed()
	{
		// � �������� �� ������� {@coords_planet} ���� ���������� ���� ��������� <b>{@building_name}</b>.

		$row = getCoordsAndPlanetname($this->data["planetid"]);
		$planet = $row["planetname"];
		$coords = getCoordLink($row["galaxy"], $row["system"], $row["position"], true, $this->data["planetid"]);

		$building_name = Core::getLanguage()->getItem($this->data["name"])." ".$this->data["level"];

		$msg = Core::getLanguage()->getItemWith("BUILDING_DESTROYED_MSG", array(
			"coords_planet" => $coords,
			"building_name" => $building_name,
			));

		$rel_userid["userid"] = null;
		$this->sendMsg(Core::getLanguage()->getItem("BUILDING_DESTROYED_SUBJECT"), $msg, $rel_userid["userid"]);

		return $this;
	}

	protected function moonDestroyed()
	{
		// ������� {@coords_planet} ���� ����������.

		$row = getCoordsAndPlanetname($this->data["planetid"]);
		$planet = $row["planetname"];
		$coords = getCoordLink($row["galaxy"], $row["system"], $row["position"], true, $this->data["planetid"]);

		$msg = Core::getLanguage()->getItemWith("MOON_DESTROYED_MSG", array(
			"coords_planet" => $coords,
			));

		$rel_userid["userid"] = null;
		$this->sendMsg(Core::getLanguage()->getItem("MOON_DESTROYED_SUBJECT"), $msg, $rel_userid["userid"]);

		return $this;
	}

	protected function retreatOther()
	{
		// {@username} ������� ��������� �� {@planet_from} {@coords_from}.
		// ��� ���� ({@fleet}) ������������ �� {@planet} {@coords}.

		$fleet_str = $this->getShipsString();

		$row = getCoordsAndPlanetname($this->data["planetid"]);
		$planet = $row["planetname"];
		$coords = getCoordLink($row["galaxy"], $row["system"], $row["position"], true, $this->data["planetid"]);

		$row = getCoordsAndPlanetname($this->data["destination"]);
		$planet_from = $row["planetname"];
		$coords_from = getCoordLink($row["galaxy"], $row["system"], $row["position"], true, $this->data["destination"]);

		$msg = Core::getLanguage()->getItemWith("RETREAT_OTHER", array(
			"username" => NS::getUser()->get("username"),
			"fleet" => $fleet_str,
			"coords_from" => $coords_from,
			"planet_from" => $planet_from,
			"coords" => $coords,
			"planet" => $planet,
			));

		$rel_userid["userid"] = NS::getUser()->get("userid");

		// Hook::event("AUTO_MSG_GRASPED_REPORT", array(&$this, &$msg));
		$this->sendMsg(Core::getLanguage()->getItem("RETREAT_OTHER_SUBJECT"), $msg, $rel_userid["userid"]);

		// if(NS::getUser()->get("userid") == 2){ debug_var($this, "[AutoMsg::graspedReport] msg: $msg"); exit; }

		return $this;
	}

	protected function retreatTransport()
	{
		$fleet_str = $this->getShipsString();

		$row = getCoordsAndPlanetname($this->data["planetid"]);
		$planet = $row["planetname"];
		$coords = getCoordLink($row["galaxy"], $row["system"], $row["position"], true, $this->data["planetid"]);

		$row = getCoordsAndPlanetname($this->data["destination"]);
		$planet_from = $row["planetname"];
		$coords_from = getCoordLink($row["galaxy"], $row["system"], $row["position"], true, $this->data["destination"]);

		$msg = Core::getLanguage()->getItemWith("RETREAT_TRANSPORT", array(
			"username" => NS::getUser()->get("username"),
			"fleet" => $fleet_str,
			"coords_from" => $coords_from,
			"planet_from" => $planet_from,
			"coords" => $coords,
			"planet" => $planet,
			));

		$rel_userid["userid"] = NS::getUser()->get("userid");

		// Hook::event("AUTO_MSG_GRASPED_REPORT", array(&$this, &$msg));
		$this->sendMsg(Core::getLanguage()->getItem("RETREAT_TRANSPORT_SUBJECT"), $msg, $rel_userid["userid"]);

		// if(NS::getUser()->get("userid") == 2){ debug_var($this, "[AutoMsg::graspedReport] msg: $msg"); exit; }

		return $this;
	}

	protected function graspedArtefactsReport()
	{
		// � �������� �� ������� {@coords_from} ���� ��������� ���������: {@artefacts}.
		// ��� ���� ������ � ����������� �������� �� ������� {@planet} {@coords}.

		// Get important data
		$artefacts_str = $this->getShipsString();

//		$row = sqlSelectRow("planet p", array("p.planetname", "g.galaxy", "g.system", "g.position"),
//			"LEFT JOIN ".PREFIX."galaxy g ON g.planetid = p.planetid",
//			"p.planetid = ".sqlVal($this->data["destination"]));
		$row = getCoordsAndPlanetname($this->data["destination"]);
		$planet = $row["planetname"];
		$coords = getCoordLink($row["galaxy"], $row["system"], $row["position"], true, $this->data["destination"]);

//		$row = sqlSelectRow("planet p", array("p.planetname", "g.galaxy", "g.system", "g.position"),
//			"LEFT JOIN ".PREFIX."galaxy g ON g.planetid = p.planetid",
//			"p.planetid = ".sqlVal($this->data["planetid"]));
		$row = getCoordsAndPlanetname($this->data["planetid"]);
		$planet_from = $row["planetname"];
		$coords_from = getCoordLink($row["galaxy"], $row["system"], $row["position"], true, $this->data["planetid"]);

		$rel_userid = sqlSelectRow("planet", "userid", "", "planetid=".sqlVal($this->data["planetid"]));

		$msg = Core::getLanguage()->getItemWith("ARTEFACTS_GRASPED_MSG", array(
			"artefacts" => $artefacts_str,
			"coords_from" => $coords_from,
			"planet_from" => $planet_from,
			"coords" => $coords,
			"planet" => $planet,
			));

		// Hook::event("AUTO_MSG_GRASPED_REPORT", array(&$this, &$msg));
		$this->sendMsg(Core::getLanguage()->getItem("ARTEFACTS_GRASPED_SUBJECT"), $msg, $rel_userid["userid"]);

		return $this;
	}

	protected function graspedReport()
	{
		// � �������� �� ������� {@coords_from} ��� �������� ���� ���������� ({@fleet}).
		// ������������ ��� ������ ������ ����������� ������������ ������� ���������������� � ����������,
		// ����������� ���� �������� ��������� � ����� ������ �� ������� {@planet} {@coords}.

		// Get important data
		$fleet_str = $this->getShipsString();

//		$row = sqlSelectRow("planet p", array("p.planetname", "g.galaxy", "g.system", "g.position"),
//			"LEFT JOIN ".PREFIX."galaxy g ON g.planetid = p.planetid",
//			"p.planetid = ".sqlVal($this->data["destination"]));
		$row = getCoordsAndPlanetname($this->data["destination"]);
		$planet = $row["planetname"];
		$coords = getCoordLink($row["galaxy"], $row["system"], $row["position"], true, $this->data["destination"]);
		// if(NS::getUser()->get("userid") == 2) debug_var($row, "[AutoMsg::graspedReport] destination: ".$this->data["destination"]
		// . ", fleet: $fleet_str, planet: $planet, coords: $coords");

//		$row = sqlSelectRow("planet p", array("p.planetname", "g.galaxy", "g.system", "g.position"),
//			"LEFT JOIN ".PREFIX."galaxy g ON g.planetid = p.planetid",
//			"p.planetid = ".sqlVal($this->data["planetid"]));
		$row = getCoordsAndPlanetname($this->data["planetid"]);
		$planet_from = $row["planetname"];
		$coords_from = getCoordLink($row["galaxy"], $row["system"], $row["position"], true, $this->data["planetid"]);
		// if(NS::getUser()->get("userid") == 2) debug_var($row, "[AutoMsg::graspedReport] planetid: ".$this->data["planetid"]
		// . ", planet: $planet_from, coords: $coords_from");

		$rel_userid = sqlSelectRow("planet", "userid", "", "planetid=".sqlVal($this->data["planetid"]));

		$msg = Core::getLanguage()->getItemWith("FLEET_GRASPED", array(
			"fleet" => $fleet_str,
			"coords_from" => $coords_from,
			"planet_from" => $planet_from,
			"coords" => $coords,
			"planet" => $planet,
			));

		// Hook::event("AUTO_MSG_GRASPED_REPORT", array(&$this, &$msg));
		$this->sendMsg(Core::getLanguage()->getItem("FLEET_GRASPED_SUBJECT"), $msg, $rel_userid["userid"]);
		// if(NS::getUser()->get("userid") == 2){ debug_var($this, "[AutoMsg::graspedReport] msg: $msg"); exit; }
		return $this;
	}

	protected function artefactReport()
	{
		$mode 		= $this->data['mode'];
		$art_data	= $this->getArtefactNameAndPosition();
		$msg 		= Core::getLanguage()->getItemWith( $mode, array("coords" => $art_data['pos'], "name" => $art_data['name']) );
		if( !empty($art_data['user']) )
		{
			$this->userid = $art_data['user'];
			$this->sendMsg(
				Core::getLanguage()->getItem("MSG_ARTEFACT"),
				$msg,
				$art_data['user']
			);
		}
		else
		{
			error_log('Error while sendind artifact related message. No owner(userid) in artefact ' . $this->data['art_id'] . '.');
		}
	}

	/**
	 *
	 * Sends message to user when his credits are changed
	 */
	protected function creditReport()
	{
		$mode	= $this->data['msg'];
		$credit	= fNumber($this->data['credits'], 2);
		if( isset($this->data['content']['exchange']) ) // Then it's exchange. Need lot content
		{
			$content = $this->data['content'];
			switch ( $content['exchange'] )
			{
				case ETYPE_FLEET:
					$name['name'] = $this->getShipsString($content['ships']);
					break;
				case ETYPE_RESOURCE:
					$name['name'] = Core::getLanguage()->getItem( strtoupper($content['res_type']) ) . ' ' . $content[ $content['res_type'] ];
					break;
				case ETYPE_ARTEFACT:
					$name = $this->getArtefactNameAndPosition($content['artef_id']);
					break;
			}
			$msg = Core::getLanguage()->getItemWith( $mode.'_LOT', array("credits" => $credit, 'lot' => $name['name']) );
		}
		else if( isset($this->data['content']['level']) )
		{
			$content = $this->data['content'];
			$name = Core::getLanguage()->getItem( strtoupper($content['name']) ) . ' ' . fNumber($content['level']);
			$msg = Core::getLanguage()->getItemWith( $mode, array("credits" => $credit, 'name' => $name) );
		}
		else if( isset($this->data['content']['name']) )
		{
			$content = $this->data['content'];
			$msg = Core::getLanguage()->getItemWith( $mode, array("credits" => $credit, 'name' => $content['name']) );
		}
		else
		{
			$msg = Core::getLanguage()->getItemWith( $mode, array("credits" => $credit) );
		}
		$this->sendMsg( Core::getLanguage()->getItem("MSG_CREDIT"), $msg, $this->userid );
	}

	/**
	 *
	 * Generates artefact string using key from 'artefact2user' table
	 * @param int $art_id Default key - $data['art_id']. If $art_id != null uses $art_id.
	 * @param int $get_coords If true then returns name with coordinates.
	 */
	protected function getArtefactNameAndPosition( $art_id = null, $get_coords = true)
	{
		if ( !empty($art_id) )
		{
			$this->data['art_id'] = $art_id;
		}
		$art_data = sqlSelectRow(
			'artefact2user AS a2u',
			'a2u.userid, c.name AS art_Name, g.galaxy, g.system, g.position, co.name AS pack_Name, a2u.level, a2u.planetid AS currPlanet',
			'LEFT JOIN '. PREFIX .'construction AS c ON c.buildingid = a2u.typeid '.
				'LEFT JOIN '. PREFIX .'galaxy AS g ON g.planetid = a2u.planetid '.
				'LEFT JOIN '. PREFIX .'construction AS co ON co.buildingid = a2u.construction_id ',
			'a2u.artid = '.$this->data['art_id']
		);
		$result = array();
		$result['user'] = $art_data['userid'];
		if( $get_coords )
		{
			if ( empty($art_data['currPlanet']) || $art_data['currPlanet'] == 0 )
			{
				$result['pos'] = Core::getLanguage()->getItem('IN_FLIGHT');
			}
			else
			{
				$result['pos'] = getCoordLink(0, 0, 0, false, $art_data['currPlanet'], true);
			}
		}
		$result['name'] = Core::getLanguage()->getItem($art_data['art_Name']);
		if ( !empty($art_data['pack_Name']) && $art_data['pack_Name'] !== 0 )
		{
			$result['name'] .= ' "';
			$result['name'] .= Core::getLanguage()->getItem($art_data['pack_Name']);
			if ( !empty($art_data['level']) && $art_data['level'] != 0 )
			{
				$result['name'] .= ': ';
				$result['name'] .= $art_data['level'];
			}
			$result['name'] .= '"';
		}
		return $result;
	}

	/**
	* Sends a message that a fleet has returned to planet.
	*
	* @return AutoMsg
	*/
	protected function returnReport()
	{
		$msg = Core::getLanguage()->getItem("FLEET_RETURNED");

		// Get important data
		$metal = fNumber($this->data["metal"]);
		$silicon = fNumber($this->data["silicon"]);
		$hydrogen = fNumber($this->data["hydrogen"] + (isset($this->data["ret_consumption"]) ? $this->data["ret_consumption"] : 0));
		$rShips = $this->getShipsString();
		$coordinates = getCoordLink($this->data["galaxy"], $this->data["system"], $this->data["position"], true, $this->data["planetid"]);

		$planet = sqlSelectField("planet", array("planetname"), "", "planetid = ".sqlVal($this->data["destination"]));
		$coords = getCoordLink($this->data["sgalaxy"], $this->data["ssystem"], $this->data["sposition"], true, $this->data["destination"]);
		$msg = sprintf($msg, $rShips, $coordinates, $planet, $coords, $metal, $silicon, $hydrogen);
		// Hook::event("AUTO_MSG_RETURN_REPORT", array(&$this, &$msg));
//		$join = "left join ".PREFIX."galaxy g on g.planetid = p.planetid";
//		$rel_userid = sqlSelectRow("planet p", "userid", $join, "(galaxy=".sqlVal($this->data["galaxy"]).") and (system=".sqlVal($this->data["system"]).") and (position=".sqlVal($this->data["position"]).")");
		$this->sendMsg(Core::getLanguage()->getItem("FLEET_RETURNED_SUBJECT"), $msg);
		return $this;
	}

	/**
	* When recycling event is completed.
	*
	* @return AutoMsg
	*/
	protected function recyclingReport()
	{
		// $capacity = fNumber($this->data["capacity"]);
		$capacity = fNumber($this->data["recycledmetal"] + $this->data["recycledsilicon"]);
		$debrismetal = fNumber($this->data["debrismetal"]);
		$debrissilicon = fNumber($this->data["debrissilicon"]);
		$recycledmetal = fNumber($this->data["recycledmetal"]);
		$recycledsilicon = fNumber($this->data["recycledsilicon"]);

		/*
		$recycler = fNumber($this->data["ships"][UNIT_RECYCLER]["quantity"]);
		$msg = Core::getLanguage()->getItem("RECYLCING_REPORT");
		$msg = sprintf($msg, $recycler, $capacity, $debrismetal, $debrissilicon, $recycledmetal, $recycledsilicon);
		$subject = sprintf(Core::getLanguage()->getItem("RECYLCING_REPORT_SUBJECT"), getCoordLink($this->data["galaxy"], $this->data["system"], $this->data["position"], true, $this->data["destination"]));
		*/

		$fleet_str = $this->getShipsString();
		$msg = Core::getLanguage()->getItemWith("RECYLCING_REPORT_EXT", array(
				'fleet' => $fleet_str,
				'debrismetal' => $debrismetal,
				'debrissilicon' => $debrissilicon,
				'recycledmetal' => $recycledmetal,
				'recycledsilicon' => $recycledsilicon,
			));
		$subject = Core::getLanguage()->getItemWith("RECYLCING_REPORT_SUBJECT_EXT", array(
				'coods' => getCoordLink($this->data["galaxy"], $this->data["system"], $this->data["position"], true, $this->data["destination"])
			));

		$this->sendMsg($subject, $msg);
		return $this;
	}

	/**
	* Sends message after a colony ship reached at a position.
	*
	* @return AutoMsg
	*/
	protected function colonizeReport()
	{
		if($this->data["success"] == "occupied")
		{
			$msg = Core::getLanguage()->getItem("PLANET_OCCUPIED");
		}
		else if($this->data["success"] == "empire")
		{
			$msg = Core::getLanguage()->getItem("TOO_MANY_PLANETS");
		}
		else
		{
			$msg = Core::getLanguage()->getItem("PLANET_COLONIZED");
		}
		$coordinates = getCoordLink($this->data["galaxy"], $this->data["system"], $this->data["position"], true);
		$msg = sprintf($msg, $coordinates);
		// Hook::event("AUTO_MSG_COLONIZE_REPORT", array(&$this, &$msg));
		$this->sendMsg(Core::getLanguage()->getItem("COLONY_ATTEMPT_SUBJECT"), $msg);
		return $this;
	}

	/**
	* Write the message into database.
	*
	* @param string	Message subject
	* @param string	Message
	*
	* @return AutoMsg
	*/
	protected function sendMsg($subject, $msg, $related_user = null)
	{
		sqlInsert("message", array(
			"mode" => $this->folder,
			"time" => $this->time,
			"sender" => null,
			"receiver" => $this->userid,
			"message" => $msg,
			"subject" => $subject,
			"readed" => 0,
			"related_user" => $related_user));
		return $this;
	}

	/**
	* Generates message for an asteroid event.
	*
	* @return AutoMsg
	*/
	protected function asteroid()
	{
		$msg = Core::getLanguage()->getItem("ASTEROID_IMPACT");
		$coordinates = getCoordLink($this->data["galaxy"], $this->data["system"], $this->data["position"], true);
		$msg = sprintf($msg, $this->data["planet"], $coordinates, fNumber($this->data["metal"]), fNumber($this->data["silicon"]));
		// Hook::event("AUTO_MSG_ASTEROID", array(&$this, &$msg));
		$this->sendMsg(Core::getLanguage()->getItem("ASTEROID_IMPACT_SUBJECT"), $msg);
		return $this;
	}

	/**
	* Generates message for members if alliance has been abandoned.
	*
	* @return AutoMsg
	*/
	protected function allyAbandoned()
	{
		$msg = Core::getLanguage()->getItem("ALLIANCE_ABANDONED");
		$msg = sprintf($msg, $this->data["tag"]);
		// Hook::event("AUTO_MSG_ALLY_ABANDONED", array(&$this, &$msg));
		$this->sendMsg(Core::getLanguage()->getItem("ALLIANCE_ABANDONED_SUBJECT"), $msg);
		return $this;
	}

	/**
	* Generates message if user's application has been receipted.
	*
	* @return AutoMsg
	*/
	protected function memberReceipted()
	{
		$msg = Core::getLanguage()->getItem("MEMBER_RECEIPTED");
		$msg = sprintf($msg, $this->data["tag"]);
		// Hook::event("AUTO_MSG_MEMBER_RECEIPTED", array(&$this, &$msg));
		$this->sendMsg(Core::getLanguage()->getItem("ALLIANCE_APPLICATION"), $msg);
		return $this;
	}

	/**
	* Generates message if user's application has been refused.
	*
	* @return AutoMsg
	*/
	protected function memberRefused()
	{
		$msg = Core::getLanguage()->getItem("MEMBER_REFUSED");
		$msg = sprintf($msg, $this->data["tag"]);
		// Hook::event("AUTO_MSG_REFUSED", array(&$this, &$msg));
		$this->sendMsg(Core::getLanguage()->getItem("ALLIANCE_APPLICATION"), $msg);
		return $this;
	}

	/**
	* Generates message if user was target of espionage.
	*
	* @return AutoMsg
	*/
	protected function espionageCommitted()
	{
		$msg = Core::getLanguage()->getItem("ESPIOANGE_COMMITTED");
		$coordss = getCoordLink($this->data["sgalaxy"], $this->data["ssystem"], $this->data["sposition"], true, $this->data["planetid"]);
		$coordinates = getCoordLink($this->data["galaxy"], $this->data["system"], $this->data["position"], true, $this->data["destination"]);
		$espLost = ($this->data["probes_lost"]) ? Core::getLanguage()->getItem("ESP_PROBES_DESTROYED_1") : "";
		$msg = sprintf($msg, $this->data["suser"], $this->data["planetname"], $coordss, $this->data["destinationplanet"], $coordinates, $this->data["defending_chance"], $espLost);
		$suserid = sqlSelectRow("user", "userid", "", "username='{$this->data["suser"]}'");
		// Hook::event("AUTO_MSG_ESPIONAGE_COMMITTED", array(&$this, &$msg));
		$this->sendMsg(Core::getLanguage()->getItem("ESPIOANGE_COMMITTED_SUBJECT"), $msg, $suserid["userid"]);
		return $this;
	}

	/**
	* Generates message for alliance members if someone joined the alliance.
	*
	* @return AutoMsg
	*/
	protected function newMember()
	{
		$msg = Core::getLanguage()->getItem("NEW_MEMBER_JOINED");
		$msg = sprintf($msg, $this->data["username"]);
		$rel_userid = sqlSelectRow("user", "userid", "", "username=".sqlVal($this->data["username"]));
		// Hook::event("AUTO_MSG_NEW_MEMBER", array(&$this, &$msg));
		$this->sendMsg(Core::getLanguage()->getItem("NEW_MEMBER_JOINED_SUBJECT"), $msg, $rel_userid["userid"]);
		return $this;
	}

	/**
	* Generates message for alliance members if someone has left the alliance.
	*
	* @return AutoMsg
	*/
	protected function memberLeft()
	{
		$msg = Core::getLanguage()->getItem("MEMBER_LEFT");
		$msg = sprintf($msg, $this->data["username"]);
		$rel_userid = sqlSelectRow("user", "userid", "", "username=".sqlVal($this->data["username"]));
		// Hook::event("AUTO_MSG_MEMBER_LEFT", array(&$this, &$msg));
		$this->sendMsg(Core::getLanguage()->getItem("MEMBER_LEFT_SUBJECT"), $msg, $rel_userid["userid"]);
		return $this;
	}

	/**
	* Generates message for alliance members if someone has been kicked off.
	*
	* @return AutoMsg
	*/
	protected function memberKicked()
	{
		$msg = Core::getLanguage()->getItem("MEMBER_KICKED");
		$msg = sprintf($msg, $this->data["username"]);
		$rel_userid = isset($this->data["userid"]) ? $this->data["userid"] : sqlSelectField("user", "userid", "", "username=".sqlVal($this->data["username"]));
		// Hook::event("AUTO_MSG_NEW_MEMBER", array(&$this, &$msg));
		$this->sendMsg(Core::getLanguage()->getItem("MEMBER_KICKED_SUBJECT"), $msg, $rel_userid);

		// Hook::event("AUTO_MSG_MEMBER_KICKED", array(&$this));
		// $this->sendMsg(Core::getLanguage()->getItem("MEMBER_KICKED_SUBJECT"), Core::getLanguage()->getItem("MEMBER_KICKED"));
		return $this;
	}
}
?>
