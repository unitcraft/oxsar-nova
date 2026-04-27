<?php
/**
* This class handles combats by starting the Java-App.
*
* Oxsar http://oxsar.ru
*
*
*/

/*

How to install MySql Connector/J:

1. copy mysql-connector-java-5.1.7-bin.jar to /usr/lib/jvm/java-6-openjdk/jre/lib/ext/
* /usr/lib/jvm/java-6-sun-1.6.0.26/jre/lib/ext/
 * 
2. export set CLASSPATH=/usr/lib/jvm/java-6-openjdk/jre/lib/ext/mysql-connector-java-5.1.7-bin.jar:$CLASSPATH

How to check MySql Connector/J:

java -jar Assault.jar 'localhost' 'db-user' 'db-passwd' 'db-name' 'na_' 'N'
java -cp Assault.jar assault.Assault 'localhost' 'db-user' 'db-passwd' 'db-name' 'na_' 'N'

*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

require_once(APP_ROOT_DIR."game/Participant.class.php");

class Assault
{
	/**
	* Event handler object.
	*
	* @var EventHandler
	*/
	protected $EH = null;

	protected $eventid = null;

	/**
	* Assault planetid
	*
	* @var integer
	*/
	protected $planetid = 0;

	/**
	* User id of the planet owner.
	*
	* @var integer
	*/
	protected $userid = 0;

	/**
	* The assault id.
	*
	* @var integer
	*/
	protected $assaultid = 0;

	/**
	* The main defender participant id.
	*
	* @var integer
	*/
	protected $mainDefenderParticipantId = 0;

	/**
	* Holds a list of all attackers and defenders.
	*
	* @var Map
	*/
	protected $attackers = null, $defenders = null;

	/**
	* Creates a new Assault Object.
	*
	* @param integer	Planet id of the assault planetid
	*
	* @return void
	*/
	public function __construct(EventHandler $EH, $eventid, $planetid = null, $userid = null, $is_rocket_attack = false)
	{
		$this->attackers 		= new Map();
		$this->defenders 		= new Map();
		$this->EH 				= $EH;
		$this->eventid 			= $eventid;
		$this->planetid 		= $planetid;
		$this->userid 			= $userid;
		$this->is_rocket_attack = $is_rocket_attack;

		// Create a new assault in the database.
		$this->assaultid = sqlInsert("assault", array("planetid" => $this->planetid, "time" => time()));

		// If the planetid is a planet, we have to update the ressource production.
		if($this->planetid && $this->userid)
		{
			$this->planet = new Planet($this->planetid, $this->userid);

			// Update By Pk
			sqlUpdate('assault', array(
				'planet_metal' => max(0, $this->planet->getData('metal') - $this->planet->getStorage('metal') * STORAGE_SAVE_FACTOR),
				'planet_silicon' => max(0, $this->planet->getData('silicon') - $this->planet->getStorage('silicon') * STORAGE_SAVE_FACTOR),
				'planet_hydrogen' => max(0, $this->planet->getData('hydrogen') - $this->planet->getStorage('hydrogen') * STORAGE_SAVE_FACTOR),
				), 'assaultid='.sqlVal($this->assaultid)
			);

			$this->loadDefenders();
		}
	}

	/**
	* Adds a new participant.
	*
	* @param integer	Participant mode (0: defender, 1: attacker)
	* @param integer	User id
	* @param array		The ships
	*
	* @return Assault
	*/
	public function addParticipant($mode, $userid, $planetid, $time, array $data, $parent_eventid, $primary_target = 0, $event_id = 0)
	{
		$Participant = new Participant($this->EH);

		$Participant->setAssaultId($this->assaultid);
		$Participant->setUserId($userid);
		$Participant->setMode($mode);
		$Participant->setPlanetId($planetid);
		$Participant->setLocation($this->planetid);
		$Participant->setTime($time);
		$Participant->setData($data);
		$Participant->setParentEventid($parent_eventid);
		$Participant->setPrimaryTarget($primary_target);
		$Participant->setEventId($event_id);

		$Participant->setDBEntry();
		if($mode === 1)
		{
			$this->attackers->push($Participant);
		}
		else
		{
			$this->defenders->push($Participant);
		}
		return $this;
	}

	/**
	* Loads the ships of the defender including the
	* participants halting at the planet.
	*
	* @return Assault
	*/
	protected function loadDefenders()
	{
		// Hook::event("LOADING_DEFENDER", array($this));

		$participantid = sqlInsert(
			"assaultparticipant",
			array(
				"assaultid" => $this->assaultid,
				"userid" => $this->userid,
				"planetid" => $this->planetid,
				"mode" => 0
			)
		);
		$this->mainDefenderParticipantId = $participantid;

//		This is not used. Why ?
//		$joins = "LEFT JOIN ".PREFIX."planet p ON p.planetid = u2s.planetid ";
//		if($this->is_rocket_attack)
//		{
//			$joins = "LEFT JOIN ".PREFIX."construction c ON c.buildingid = u2s.unitid AND c.mode=".UNIT_TYPE_DEFENSE." ";
//		}

		$result = sqlSelect(
			"unit2shipyard u2s",
			array(
				"u2s.unitid", "u2s.quantity", "u2s.damaged", "u2s.shell_percent"
			),
//			$joins,
			'',
			"u2s.planetid = ".sqlVal($this->planetid)
		);
		while($row = sqlFetch($result))
		{
			sqlInsert(
				"fleet2assault",
				array(
					"assaultid" => $this->assaultid,
					"participantid" => $participantid,
					"userid" => $this->userid,
					"unitid" => $row["unitid"],
					"mode" => 0,
					"quantity" => $row["quantity"],
					"damaged" => $row["damaged"],
					"shell_percent" => $row["shell_percent"],
					"org_quantity" => $row["quantity"],
					"org_damaged" => $row["damaged"],
					"org_shell_percent" => $row["shell_percent"]
				)
			);
		}
		sqlEnd($result);

		if(!$this->is_rocket_attack && $this->planetid)
		{
			// Get holding fleets.
			$result = sqlSelect("events e", array("e.eventid", "e.user AS userid", "e.planetid", "e.time", "e.data"),
				"INNER JOIN ".PREFIX."user u ON u.userid = e.user
                 INNER JOIN ".PREFIX."planet p ON p.planetid = e.planetid",
				"mode IN (".sqlArray(EVENT_HOLDING, EVENT_ALIEN_HOLDING).") AND destination = ".sqlVal($this->planetid)." AND processed=".EVENT_PROCESSED_WAIT
			);
			while($row = sqlFetch($result)){
				$this->addParticipant( 0, $row["userid"], $row["planetid"], $row["time"], unserialize($row["data"]), null, 0, $row["eventid"]);
			}
			sqlEnd($result);
		}
		return $this;
	}

	protected function getRandomTargetBuilding()
	{
		if($this->is_rocket_attack || !$this->planetid || !$this->userid)
		{
			return false;
		}

		$defender_builds = array();
		$result = sqlSelect("building2planet", array("buildingid", "level - added as level"), "",
		"planetid=".sqlVal($this->planetid)." AND buildingid NOT IN (".sqlArray(UNIT_EXCHANGE, UNIT_NANO_FACTORY).")");
		while($row = sqlFetch($result))
		{
			$defender_builds[$row["buildingid"]] = array("buildingid" => $row["buildingid"], "level" => $row["level"]);
		}
		sqlEnd($result);
		// debug_var($defender_builds, "[getRandomTargetBuilding] defender_builds");

		if(count($defender_builds) <= 0)
		{
			return false;
		}

		$user_checked = array();
		$this->attackers->rewind();
		while($this->attackers->next())
		{
			$p = $this->attackers->current();
			if($p instanceof Participant && !isset($user_checked[$p->getUserId()]))
			{
				$user_checked[$p->getUserId()] = true;
				$result = sqlSelect("building2planet b2p", array("b2p.buildingid", "b2p.level - b2p.added as level"),
					"JOIN ".PREFIX."planet p ON p.planetid = b2p.planetid",
					"p.userid=".sqlVal($p->getUserId()));
				while($row = sqlFetch($result))
				{
					if(isset($defender_builds[$row["buildingid"]]))
					{
						$defender_builds[$row["buildingid"]]["checked"] = true;
						$min_result_level = $row["level"] + DESTROY_BUILD_RESULT_MIN_OFFS_LEVEL;
						$new_result_level = $defender_builds[$row["buildingid"]]["level"] - 1;
						if($new_result_level < $min_result_level)
						{
							unset($defender_builds[$row["buildingid"]]);
							// debug_var($defender_builds, "[getRandomTargetBuilding] removed id {$row['buildingid']} for Participant: ".$p->getUserId().", $build_result_level < $min_result_level");
						}
					}
				}
				sqlEnd($result);
			}
		}
		// $this->attackers->rewind();
		foreach($defender_builds as $buildingid => $value)
		{
			if(empty($value["checked"]))
			{
				unset($defender_builds[$buildingid]);
				// debug_var($defender_builds, "[getRandomTargetBuilding] removed unchecked id: $buildingid");
			}
		}
		// debug_var($defender_builds, "[getRandomTargetBuilding] defender_builds before random");

		if(count($defender_builds) > 0)
		{
			$mines = array();
			$builds = array();
			foreach($defender_builds as $buildingid => $value)
			{
				if($value["level"] > 20 && in_array($buildingid, array(UNIT_METALMINE, UNIT_SILICON_LAB, UNIT_HYDROGEN_LAB)))
				{
					$mines[] = $buildingid;
				}
				else
				{
					$builds[] = $buildingid;
				}
			}
			// shuffle($mines);
			shuffle($builds);
			$builds = array_slice(array_merge($mines, $builds), 0, 2);

			$buildingid = $builds[array_rand($builds)];
			$level = $defender_builds[$buildingid]["level"];

			// debug_var(array("buildingid" => $buildingid, "level" => $level), "[getRandomTargetBuilding] result");

			return self::getTargetBuildingRes($buildingid, $level);
		}
		return false;
	}

	public static function getTargetBuildingRes($buildingid, $level)
	{
		$metal = $silicon = $hydrogen = 0;
		$row = getConstructionDesc($buildingid);
		for($i = $level; $i <= $level; $i++)
		{
			if($row["basic_metal"] > 0)
			{
				$metal += parseChargeFormula($row["charge_metal"], $row["basic_metal"], $i);
			}
			if($row["basic_silicon"] > 0)
			{
				$silicon += parseChargeFormula($row["charge_silicon"], $row["basic_silicon"], $i);
			}
			if($row["basic_hydrogen"] > 0)
			{
				$hydrogen += parseChargeFormula($row["charge_hydrogen"], $row["basic_hydrogen"], $i);
			}
		}

		$target_building = array(
				"id" => $buildingid,
				"level" => $level,
				"metal" => round($metal),
				"silicon" => round($silicon),
				"hydrogen" => round($hydrogen),
				"points" => round(($metal + $silicon + $hydrogen) * RES_TO_BUILD_POINTS, POINTS_PRECISION),
			);
		// debug_var($target_building, "[getTargetBuildingRes] target_buildingid");
		return $target_building;
	}

	protected function processAchievements( $user_id, $planet_id )
	{
		if($user_id && $planet_id)
		{
			AchievementsService::processAchievements($user_id, $planet_id);
		}
	}

	/**
	 * Checks if there were moon created on planet within time period.
	 * @param int $planet_id Planet ID.
	 * @param int $time Time.
	 */
	/*
	protected function checkCreatedMoonOnPlanet( $planet_id, $time = MOON_CREATION_PLANET_INTERVAL )
	{
		$crit = new CDbCriteria();
		$crit->addCondition('planetid = :p');
		$crit->addCondition('moon = 1');
		$crit->addCondition('time >= :t');
		$crit->params[':p'] = $planet_id;
		$crit->params[':t'] = time() - $time;
		$planets = Assault_YII::model()->find($crit);
		if( $planets )
		{
			return true;
		}
		return false;
	}
	*/

	/**
	 * Checks if there were moon created in this system within time period.
	 * @param int $planet_id Planet ID wich system should we check.
	 * @param int $time Time.
	 */
	protected function checkCreatedMoonInSystem( $planet_id, $time = MOON_CREATION_SYSTEM_INTERVAL )
	{
		// План 37.5d.5#3: replaced Galaxy_YII + Assault_YII (CDbCriteria → raw SQL).
		$planet_id_q = sqlVal($planet_id);
		$planet_galaxy = sqlSelectRow("galaxy", array("system", "galaxy"), "",
			"planetid = ".$planet_id_q." OR moonid = ".$planet_id_q);
		if( !$planet_galaxy )
		{
			return false;
		}

		$planets_search = array();
		$result = sqlSelect("galaxy", array("planetid", "moonid"), "",
			"system = ".sqlVal($planet_galaxy["system"])
			." AND galaxy = ".sqlVal($planet_galaxy["galaxy"])
			." AND destroyed = 0");
		while($row = sqlFetch($result))
		{
			if( !empty($row["planetid"]) )
			{
				$planets_search[] = $row["planetid"];
			}
			elseif( !empty($row["moonid"]) )
			{
				$planets_search[] = $row["planetid"];
			}
		}
		sqlEnd($result);

		if( !empty($planets_search) )
		{
			$assault = sqlSelectField("assault", "assaultid", "",
				"planetid IN (".sqlArray($planets_search).")"
				." AND moon = 1"
				." AND time >= ".sqlVal(time() - $time));
			if( $assault )
			{
				return true;
			}
		}
		return false;
	}

    protected function isMoonPosibility()
    {
        return sqlSelectField("artefact_probobility", "probobility", "", "type=".sqlVal(ARTEFACT_MOON_CREATOR)) > 0;
    }

    /**
	 * Checks if there were moon created in this system within time period.
	 * @param int $planet_id Planet ID wich system should we check.
	 * @param int $time Time.
	 */
	protected function checkCreatedMoonForUser( $user_id, $time = MOON_CREATION_USER_INTERVAL )
	{
		// План 37.5d.5#3: replaced Planet_YII + Assault_YII (CDbCriteria → raw SQL).
		$planets_search = array();
		$result = sqlSelect("planet", "planetid", "", "userid = ".sqlVal($user_id));
		while($row = sqlFetch($result))
		{
			$planets_search[] = $row["planetid"];
		}
		sqlEnd($result);

		if( !empty($planets_search) )
		{
			$assault = sqlSelectField("assault", "assaultid", "",
				"planetid IN (".sqlArray($planets_search).")"
				." AND moon = 1"
				." AND time >= ".sqlVal(time() - $time));
			if( $assault )
			{
				return true;
			}
		}
		return false;
	}

	/**
	* Starts the assault.
	*
	* @return Assault
	*/
	public function startAssault($galaxy = 0, $system = 0, $position = 0, $params = null) // $target_building = null)
	{
		$on_planet = $galaxy && $this->planetid && $this->userid ? true : false;

		// Update By Pk
		sqlUpdate(
			'assault',
			array('advanced_system' => ( ( self::isAdvanced($galaxy, $system, $position, $params) != 0 ) ? 1 : 0 )),
			'assaultid='.sqlVal($this->assaultid)
		);

		if( $on_planet )
		{
			$mode = isset($params['mode']) ? $params['mode'] : null;
			if(in_array($mode, array(EVENT_ATTACK_DESTROY_MOON, EVENT_ATTACK_ALLIANCE_DESTROY_MOON)))
			{
				if($this->planet->getData('ismoon'))
				{
					// Update By Pk
					sqlUpdate('assault', array(
						'target_moon' => 1,
						'target_destroyed' => 0
						), 'assaultid='.sqlVal($this->assaultid));
				}
			}
			else
			{
				$target_building = null;
				if(in_array($mode, array(EVENT_ATTACK_DESTROY_BUILDING, EVENT_ATTACK_ALLIANCE_DESTROY_BUILDING)))
				{
					$target_building = $this->getRandomTargetBuilding();
				}
				else
				{
					$target_building = isset($params['target_building']) ? $params['target_building'] : null;
					if($target_building === true)
					{
						$target_building = $this->getRandomTargetBuilding();
					}
				}
				if(is_array($target_building) && isset($target_building['id']))
				{
					if(!isset($target_building['metal']))
					{
						$target_building = self::getTargetBuildingRes($target_building['id'], $target_building['level']);
					}

					// Update By Pk
					sqlUpdate(
						'assault',
						array(
							'target_buildingid'	=> $target_building['id'],
							'building_level'	=> $target_building['level'],
							'building_metal'	=> $target_building['metal'],
							'building_silicon'	=> $target_building['silicon'],
							'building_hydrogen'	=> $target_building['hydrogen'],
							'target_destroyed'	=> 0,
						),
						'assaultid='.sqlVal($this->assaultid)
					);
				}
			}

			$moon_allow_type = 0;
            if( !$this->isMoonPosibility() )
            {
				$moon_allow_type = 3;
            }
			elseif( $this->checkCreatedMoonForUser($this->userid) )
			{
				$moon_allow_type = 2;
			}
			elseif( $this->checkCreatedMoonInSystem($this->planetid) )
			{
				$moon_allow_type = 1;
			}

			if( $moon_allow_type != 0 )
			{
				// Update By Pk
				sqlUpdate('assault', array( 'moon_allow_type' => $moon_allow_type ), 'assaultid='.sqlVal($this->assaultid));
			}
		}


		// Include database access
		$database = array();
		require(APP_ROOT_DIR."config.inc.php");

		// Start assault
		$start_time = time();
		try
		{
			$exec_params = array(
				APP_ROOT_DIR.'game/Assault.jar',
				'assault.Assault',
				$database["host"],
				$database["user"],
				$database["userpw"],
				$database["databasename"],
				$database["tableprefix"],
				$this->assaultid
			);
			$s = "java -cp '".implode('\' \'', $exec_params).'\'';
			exec($s);
		}
		catch(Exception $e)
		{
			error_log('Failed in launching JAVA Assault. Assault_id = ' . $this->assaultid, 'error');
		}
		unset($database);

		// if(time() - $start_time > 60)
		{
			// reconnect
			global $__pdo_db; $db = $__pdo_db;
			$db->setActive(false);
			$db->setActive(true);
		}

		$row = sqlSelectRow(
			'assault',
			array(
				'result', 'moonchance', 'moon',
				'accomplished', 'target_moon', 'target_destroyed',
				'building_level', 'target_buildingid'
			),
			'',
			'assaultid = '.sqlVal($this->assaultid)
		);
		if( !$row["accomplished"] )
		{
			error_log('JAVA failed in processing assault. Assault_id = ' . $this->assaultid, 'error');
			Logger::addMessage("Sorry, could not start battle (".$this->assaultid.").");
		}
		if( $on_planet && $row["target_destroyed"] )
		{
			if($row["target_moon"])
			{
				if($this->planet->getData("ismoon"))
				{
					logMessage("[MOON DESTROYED] userid: {$this->userid}, planetid: {$this->planetid}, assaultid: {$this->assaultid}");

					// Update By Pk
					sqlUpdate("planet", array("userid" => null), "planetid=".sqlVal($this->planetid));
					// Update By Pk
					sqlUpdate("galaxy", array("moonid" => null), "moonid=".sqlVal($this->planetid));

					new AutoMsg(MSG_MOON_DESTROYED, $this->userid, time(), array(
								"planetid" => $this->planetid,
								));
				}
			}
			else
			{
				$build_row = sqlSelectRow("building2planet b2p", "b2p.*, b.name",
					"INNER JOIN ".PREFIX."construction b ON b.buildingid = b2p.buildingid ",
					"b2p.planetid=".sqlVal($this->planetid)." AND b2p.buildingid=".sqlVal($row["target_buildingid"]));
				$real_level = $build_row["level"] - $build_row["added"];
				if($real_level == $row["building_level"])
				{
					if($build_row["level"]-1 > 0)
					{
						// Update By Pk
						sqlUpdate("building2planet", array("level" => $build_row["level"]-1),
							"planetid=".sqlVal($this->planetid)." AND buildingid=".sqlVal($row["target_buildingid"]));
					}
					else
					{
						sqlDelete("building2planet",
							"planetid=".sqlVal($this->planetid)." AND buildingid=".sqlVal($row["target_buildingid"]));
					}

					// Update By Pk
					sqlQuery("UPDATE ".PREFIX."user SET "
						. "	points = points-".sqlVal($target_building["points"])
						. ", b_points = b_points-".sqlVal($target_building["points"])
						. ", b_count = b_count - 1"
                        . ", ".updateDmPointsSetSql()
						. " WHERE userid=".sqlVal($this->userid));

					new AutoMsg(MSG_BUILDING_DESTROYED, $this->userid, time(), array(
								"planetid" => $this->planetid,
								"buildingid" => $build_row["buildingid"],
								"name" => $build_row["name"],
								"level" => $build_row["level"],
								));
				}
			}
		}

		if($on_planet && $row["moon"])
		{
			Artefact::appear(
				mt_rand(0, 70) ? ARTEFACT_MOON_CREATOR : ARTEFACT_PLANET_CREATOR,
				$this->userid,
				$this->planetid,
				array("delay" => 0)
			);
		}

		$this->defenders->rewind();
		while($this->defenders->next()) //defenders _before_ attackers, 'cause of artefacts
		{
			$Participant = $this->defenders->current();
			if($Participant instanceof Participant)
			{
				$Participant->finish();
			}
		}

		if( /*$on_planet*/  $this->mainDefenderParticipantId)
		{
			$this->updateMainDefender(); // $row["lostunits_defender"]);
		}

		if( $on_planet && $row["result"] == 1 ) // victory - attacker has a chance to capture artefacts
		{
			$def_lost_artefacts = array();
			$alive_attackers 	= array();
			$result = sqlSelect(
				"fleet2assault f2a",
				array("f2a.participantid", "f2a.userid"),
				" LEFT JOIN ".PREFIX."assaultparticipant ap ON ap.participantid = f2a.participantid AND ap.assaultid = f2a.assaultid AND ap.userid = f2a.userid"
					. " LEFT JOIN ".PREFIX."construction b ON b.buildingid = f2a.unitid",
				"f2a.assaultid=".sqlVal($this->assaultid)
					. " AND f2a.quantity > 0 "
					. " AND f2a.mode = 1 "
					. " AND b.mode != ".UNIT_TYPE_ARTEFACT,
				"", // Order
				"", // Limit
				"f2a.participantid, f2a.userid" // Group By
			);
			while($row = sqlFetch($result))
			{
				$alive_attackers[ $row["participantid"] ] = $row;
			}
			sqlEnd($result);

			if($alive_attackers)
			{
				$query = "SELECT a2u.artid, ads.trophy_chance, u.last"
					. " FROM ".PREFIX."artefact2user a2u "
					. " INNER JOIN ".PREFIX."user u ON u.userid = a2u.userid "
					. " INNER JOIN ".PREFIX."artefact_datasheet ads ON a2u.typeid = ads.typeid "
					. " INNER JOIN ".PREFIX."events e ON e.eventid = a2u.lifetime_eventid "
					. " WHERE a2u.planetid=".sqlVal($this->planetid)
					. "	AND a2u.deleted=0 "
					// . "	AND ads.trophy_chance > 0 "
					. "	AND a2u.times_left > 0 "
					. "	AND e.time-".time()." > 60*60*4";
				$result = sqlQuery($query);
				while($row = sqlFetch($result))
				{
                    if($row['last'] < time()-60*60*24*7){
                        $row["trophy_chance"] = clampVal($row["trophy_chance"] * ABANDONED_USER_ARTEFACT_CAPTURE_CHANCE_SCALE,
                                ABANDONED_USER_MIN_ARTEFACT_CAPTURE_CHANCE, ABANDONED_USER_MAX_ARTEFACT_CAPTURE_CHANCE);
                    }
					$chance = mt_rand(0, 100);
					if($row["trophy_chance"] > 0 && $chance <= $row["trophy_chance"])
					{
						$attacker = $alive_attackers[ array_rand( $alive_attackers ) ];
						$this->captureArtefact($row['artid'], $attacker['userid'], $attacker['participantid'], $def_lost_artefacts);
					}
				}
				sqlEnd($result);

				if($def_lost_artefacts)
				{
					new AutoMsg(MSG_LOST_ARTEFACTS, $this->userid, time(), array(
						"planetid" => $this->planetid,
						"ships" => $def_lost_artefacts,
					));
				}
			}
		}

		$this->attackers->rewind();
		while($this->attackers->next())
		{
			$Participant = $this->attackers->current();
			if($Participant instanceof Participant)
			{
				$Participant->finish();
			}
		}

		if( $on_planet && $this->planetid && ACHIEVEMENTS_ENABLED )
		{
			$assault_participants = array();
			$result = sqlSelect(
				'assaultparticipant ap',
				array('distinct(ap.userid) AS user_id', 'ap.planetid AS planet_id'),
				'',
				'ap.assaultid='.sqlVal( $this->assaultid )
			);
			while($row = sqlFetch($result))
			{
				$this->processAchievements($row['user_id'], $row['planet_id']);
			}
			sqlEnd($result);
		}
		return $this;
	}

	/**
	 * Captures artefact
	 * @param int $id
	 * @param int $a_user
	 * @param int $a_participant
	 * @param array $d_artefacts
	 */
	protected function captureArtefact( $id, $a_user, $a_participant, &$d_artefacts )
	{
		$art_data = sqlSelectRow(
			'artefact2user a2u',
			array(
				'a2u.artid as artid',
				'a2u.level as level',
				'a2u.construction_id as con_id',
				'ac.name as con_name',
				'a2u.typeid as type',
				'c.name as name',
			),
			array(
				' LEFT JOIN '.PREFIX.'construction ac ON a2u.construction_id = ac.buildingid ',
				' LEFT JOIN '.PREFIX.'construction c ON a2u.typeid = c.buildingid ',
			),
			'a2u.artid = ' . sqlVal($id)
		);

		if ( $art_data && Artefact::onOwnerChange($id, $a_user) )
		{
			$type = $art_data['type'];
			$name = $art_data['name'];
			if( !isset( $d_artefacts[$type] ) )
			{
				$d_artefacts[$type] = array(
					'name'		=> $name,
					'quantity'	=> 1,
				);
				$d_artefacts[$type]['art_ids'][$id] = $art_data;
			}
			else
			{
				$d_artefacts[$type]['quantity']++;
				$d_artefacts[$type]['art_ids'][$id] = $art_data;
			}

			$fleet = sqlSelectRow(
				'fleet2assault',
				'*',
				'',
				array(
					'participantid='. sqlVal($a_participant),
					'unitid=' 		. sqlVal($type)
				)
			);

			if($fleet)
			{
				$etc = unserialize($fleet['etc']);
				$etc[ $id ] = $art_data;
				// Update By Pk
				sqlUpdate(
					'fleet2assault',
					array(
						'quantity'	=> $fleet['quantity']+ 1,
						'grasped'	=> $fleet['grasped'] + 1,
						'etc'		=> serialize($etc),
					),
					' participantid='.sqlVal($a_participant)
						. ' AND unitid='.sqlVal( $type )
				);
			}
			else
			{
				$etc = array();
				$etc[$id] = $art_data;
				sqlInsert(
					'fleet2assault',
					array(
						'assaultid'		=> $this->assaultid,
						'participantid' => $a_participant,
						'userid'		=> $a_user,
						'unitid'		=> $type,
						'mode'			=> 1,
						'quantity'		=> 1,
						'grasped'		=> 1,
						'etc'			=> serialize($etc),
					)
				);
			}
		}
		else
		{
			error_log(
				'Can\'t change owner of artefact. Artefact ID -> ' . $id
					. '. Assault ID -> ' . $this->assaultid
					. '. New Owner ID -> ' . $a_user
					. '. Old Owner ID -> ' . $this->userid,
				'error',
				'Critical'
			);
			return false;
		}
	}

	/**
	 *
	 * Checks if needed to use advaned assault system.
	 * @param unknown_type $galaxy
	 * @param unknown_type $system
	 * @param unknown_type $position
	 * @param unknown_type $params
	 */
	protected function isAdvanced($galaxy = 0, $system = 0, $position = 0, $params = null)
	{
		return NS::getGalaxyParam($galaxy, "ADVANCED_BATTLE", 0);
	}

	/**
	* Updates units after the combat for the main defender.
	*
	* @param integer	Total number of lost units
	* @param integer	User id
	* @param integer	Planet id
	*
	* @return Assault
	*/
	protected function updateMainDefender()
	{
		$result = sqlSelect(
			"fleet2assault f2a",
			array("f2a.unitid", "f2a.quantity", "f2a.damaged", "f2a.shell_percent", "b.mode"),
			" LEFT JOIN ".PREFIX."construction b ON b.buildingid = f2a.unitid",
			"f2a.userid = ".sqlVal($this->userid)." AND f2a.assaultid = ".sqlVal($this->assaultid)." AND f2a.participantid = ".sqlVal($this->mainDefenderParticipantId)
		);
		while($row = sqlFetch($result))
		{
			if($row["mode"] != UNIT_TYPE_ARTEFACT)
			{
				logShipsChanging($row["unitid"], $this->planetid, $row["quantity"], 0, "[updateMainDefender]");
				if($row["quantity"] > 0)
				{
					Core::getQuery()->replace(
						"unit2shipyard",
						array("quantity", "damaged", "shell_percent", "unitid", "planetid"),
						array($row["quantity"], $row["damaged"], $row["shell_percent"], $row["unitid"], $this->planetid)
					);
				}
				else
				{
					Core::getQuery()->delete(
						"unit2shipyard",
						"unitid = ".sqlVal($row["unitid"])." AND planetid = ".sqlVal($this->planetid)
					);
				}
			}
		}
		sqlEnd($result);
		return $this;
	}

	/**
	* Returns the assault id.
	*
	* @return integer
	*/
	public function getAssaultId()
	{
		return $this->assaultid;
	}

	/**
	* Returns the assault planetid.
	*
	* @return integer
	*/
	public function getLocation()
	{
		return $this->planetid;
	}
}
?>