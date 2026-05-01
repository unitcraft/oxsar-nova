<?php
/**
* Represents an assault participant.
*
* Oxsar http://oxsar.ru
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class Participant
{
	/**
	* Event handler object.
	*
	* @var EventHandler
	*/
	protected $EH = null;

	/**
	* Assault id.
	*
	* @var integer
	*/
	protected $assaultid = 0;

	/**
	* Event id (When defender).
	*
	* @var integer
	*/
	protected $eventid = 0;

	protected $parent_eventid = null;

	/**
	* User id.
	*
	* @var integer
	*/
	protected $userid = 0;

	/**
	* Participant id (0: defender, 1: attacker).
	*
	* @var integer
	*/
	protected $mode = 0;

	/**
	* Source planet
	*
	* @var integer
	*/
	protected $planetid = 0;

	/**
	* Planet planetid of the combat.
	*
	* @var integer
	*/
	protected $location = 0;

	/**
	* Assault start time.
	*
	* @var integer
	*/
	protected $time = 0;

	/**
	* Event data field.
	*
	* @var array
	*/
	protected $data = array();

	protected $primary_target = 0;

	/**
	* The participant's ships.
	*
	* @var array
	*/
	protected $ships = array();

	/**
	* Participant id.
	*
	* @var integer
	*/
	protected $participantid = 0;

	/**
	* Creates a new participant.
	*
	* @param integer	Assault id
	* @param integer	User id
	* @param integer	Participant id (0: defender, 1: attacker)
	* @param integer	Source planet
	* @param integer	Time when the assault has started
	* @param array		Event data field
	*
	* @return void
	*/
	public function __construct(EventHandler $EH)
	{
		$this->EH = $EH;
	}

	/**
	* Sets the data and ships.
	*
	* @param array		Data field
	*
	* @return Participant
	*/
	public function setData(array $data)
	{
		$this->ships = $data["ships"];
		$data["ships"] = null;
		$this->data = $data;
		return $this;
	}

	/**
	* Retuns the data.
	*
	* @param string	Data parameter [optional]
	*
	* @return mixed	Data value or false
	*/
	public function getData($param = null)
	{
		if(is_null($param))
		{
			return $this->data;
		}
		return (isset($this->data[$param])) ? $this->data[$param] : false;
	}

	/**
	* Returns the ships.
	*
	* @return array
	*/
	public function getShips()
	{
		return $this->ships;
	}

	/**
	* Saves the participant data into database.
	*
	* @return Participant
	*/
	public function setDBEntry()
	{
		$preloaded = $this->data["metal"] + $this->data["silicon"] + $this->data["hydrogen"];
		$this->participantid = sqlInsert(
			"assaultparticipant",
			array(
				"assaultid"				=> $this->assaultid,
				"userid"				=> $this->userid,
				"planetid"				=> $this->planetid,
				"mode"					=> $this->mode,
				"consumption"			=> $this->data["consumption"],
				"preloaded"				=> $preloaded,
				"target_unitid"			=> $this->primary_target,
				"add_gun_tech"			=> intval($this->data["add_tech_".UNIT_GUN_TECH]),
				"add_shield_tech"		=> intval($this->data["add_tech_".UNIT_SHIELD_TECH]),
				"add_shell_tech"		=> intval($this->data["add_tech_".UNIT_SHELL_TECH]),
				"add_ballistics_tech"	=> intval($this->data["add_tech_".UNIT_BALLISTICS_TECH]),
				"add_masking_tech" 		=> intval($this->data["add_tech_".UNIT_MASKING_TECH]),
				"add_laser_tech"		=> intval($this->data["add_tech_".UNIT_LASER_TECH]),
				"add_ion_tech"			=> intval($this->data["add_tech_".UNIT_ION_TECH]),
				"add_plasma_tech"		=> intval($this->data["add_tech_".UNIT_PLASMA_TECH]),
			)
		);
		// План 86 audit: продакшн-аналог Simulator-бага. participantid
		// идёт FK в fleet2assault.participantid (foreach ниже). При
		// провале INSERT (FK на assault/user/planet) early-return,
		// чтобы не создать каскадный мусор в fleet2assault.
		if ($this->participantid === false) {
			return $this;
		}

		foreach($this->ships as $ship)
		{
			$etc = null;
			if($ship["mode"] == UNIT_TYPE_ARTEFACT)
			{
				$etc = $ship["art_ids"];
				foreach($etc as $art_key => $art_data)
				{
					if(!Artefact::triggerBattle($art_key)){
                        $ship["quantity"] = $ship["quantity"]-1;

                        // Updating by PK
                        sqlUpdate("artefact2user", array(
                            "active" => 0,
                            "deleted" => time(),
                            "reason" => "found no times_left, fixed by battle (new 2)"
                            ), "artid=".sqlVal($art_key)." AND deleted=0");
                    }
				}
			}
            if($ship["quantity"] >= 1){
                fixUnitDamagedVars($ship);

                sqlInsert(
                    "fleet2assault",
                    array(
                        "assaultid"			=> $this->assaultid,
                        "participantid"		=> $this->participantid,
                        "userid"			=> $this->userid,
                        "unitid"			=> $ship["id"],
                        "mode"				=> $this->mode,
                        "quantity"			=> floor($ship["quantity"]),
                        "damaged"			=> floor($ship["damaged"]),
                        "shell_percent"		=> floor($ship["shell_percent"]),
                        "org_quantity"		=> floor($ship["quantity"]),
                        "org_damaged"		=> floor($ship["damaged"]),
                        "org_shell_percent"	=> floor($ship["shell_percent"]),
                        "etc"				=> is_null($etc) ? null : serialize($etc)
                    )
                );
            }
		}
		return $this;
	}

	/**
	* Drops all artefacts of a defeated fleet on the planet.
	*
	* @param array		artefact artid
	* @param integer	artefact type
	*/
	protected function dropArtefacts($art_ids, $typeid, $owner)
	{
		$artid_list = array();
		foreach($art_ids as $art)
		{
			$artid_list[] = $art["artid"];
		}

		$new_userid = $owner; // sqlSelectField("planet", "userid", "", "planetid=".sqlVal($this->location));

		$result = sqlQuery("SELECT a2u.artid, a2u.userid, ads.typeid, ads.unique, a2u.active, ads.usable, a2u.times_left, e.time, a2u.lifetime_eventid "
			. " FROM ".PREFIX."artefact2user a2u "
			. " INNER JOIN ".PREFIX."artefact_datasheet ads ON a2u.typeid = ads.typeid "
			. " INNER JOIN ".PREFIX."events e ON e.eventid = a2u.lifetime_eventid "
			. " WHERE a2u.artid in (".sqlArray($artid_list).") AND a2u.deleted=0 AND times_left>0");
		while($row = sqlFetch($result))
		{
			Artefact::onOwnerChange($row["artid"], $new_userid, $this->location);
			if($row["typeid"] == ARTEFACT_BUG){
				Artefact::activate($row["artid"], $new_userid, $this->location, false);
			}
		}
		sqlEnd($result);
	}

	/**
	* Makes final calculations and sends the fleet back.
	*
	* @param integer	Assault result
	*
	* @return Participant
	*/
	public function finish()
	{
		// Update units after battle.
		// if((($result == 0 || $result == 1) && $this->mode == 1) || (($result == 1 || $result == 2) && $this->mode == 0))
		//1- win, 2 - lose, 0 - draw
		$grasped_data["ships"]		= array();
		$grasped_artefacts["ships"]	= array();
		$this->data["ships"]		= array();
		$artefacts_grasped 			= array();
		$metal		= 0;
		$silicon	= 0;
		$hydrogen	= 0;
		$artefacts	= 0;
		$owner = sqlQueryField('SELECT userid FROM '.PREFIX.'planet WHERE planetid='.sqlVal( $this->location ));
		$result = sqlSelect(
			'fleet2assault f2a',
			array(
				'f2a.unitid', 'f2a.quantity', 'f2a.org_quantity',
				'f2a.damaged', 'f2a.grasped', 'f2a.shell_percent',
				'b.name', 'b.mode', 'f2a.etc',
				'ap.haul_metal', 'ap.haul_silicon', 'ap.haul_hydrogen'
			),
			' LEFT JOIN '.PREFIX.'assaultparticipant ap ON ap.participantid = f2a.participantid AND ap.assaultid = f2a.assaultid AND ap.userid = f2a.userid'
				. ' LEFT JOIN '.PREFIX.'construction b ON b.buildingid = f2a.unitid',
			'f2a.participantid = '.sqlVal($this->participantid)
		);
		while( $row = sqlFetch($result) )
		{
			$id = $row["unitid"];
			if($row["mode"] == UNIT_TYPE_ARTEFACT)
			{
				$artefacts++;
				$this->data["ships"][$id]["id"]			= $row["unitid"];
				$this->data["ships"][$id]["quantity"]	= $row["quantity"];
				$this->data["ships"][$id]["grasped"]	= $row["grasped"];
				$this->data["ships"][$id]["damaged"]	= $row["damaged"];
				$this->data["ships"][$id]["shell_percent"]	= $row["shell_percent"];
				$this->data["ships"][$id]["name"]		= $row["name"];
				$this->data["ships"][$id]["mode"]		= $row["mode"];
				$this->data["ships"][$id]["art_ids"]	= unserialize($row["etc"]);
				if( $row["grasped"] > 0 )
				{
					if ( isset($grasped_artefacts["ships"][$id]) ) // Такой тип артефакта уже есть. Но такого быть не должно, так как уникальность unit id
					{
						error_log('Found duplicated artefact(2 or more rows in fleet2assault have same unit_id for this participant).', 'error');
					}
					else // Первый раз встречаем такой тип артефакта.
					{
						$grasped_artefacts["ships"][$id] = $this->data["ships"][$id];
					}
				}
			}
			else
			{
				if( $this->data["xSkirmish"] && ( ( $row["quantity"] - $row['org_quantity'] ) != 0 ) )
				{
					sqlInsert(
						'expedition_found_units',
						array(
							'unit_id'		=> $id,
							'expedition_id'	=> $this->data['expedition_id'],
							'quantity'		=> $row['quantity'] - $row['org_quantity'],
						)
					);
				}

				if($row["quantity"] > 0 && $id != UNIT_INTERPLANETARY_ROCKET)
				{
					$this->data["ships"][$id]["id"]				= $row["unitid"];
					$this->data["ships"][$id]["quantity"]		= $row["quantity"];
					$this->data["ships"][$id]["grasped"]		= $row["grasped"];
					$this->data["ships"][$id]["damaged"]		= $row["damaged"];
					$this->data["ships"][$id]["shell_percent"]	= $row["shell_percent"];
					$this->data["ships"][$id]["name"]			= $row["name"];
					$this->data["ships"][$id]["mode"]			= $row["mode"];

					if($row["grasped"] > 0)
					{
						$grasped_data["ships"][$id]				= $this->data["ships"][$id];
						$grasped_data["ships"][$id]["quantity"]	= $grasped_data["ships"][$id]["grasped"];
						$grasped_data["ships"][$id]["damaged"]	= $grasped_data["ships"][$id]["grasped"];
					}
				}
				else
				{
					logShipsChanging($row["unitid"], $this->planetid, $row["quantity"], 1, "[killed] at planet: {$this->location}, participantid: {$this->participantid}");
				}
				$metal		= $row["haul_metal"];
				$silicon	= $row["haul_silicon"];
				$hydrogen	= $row["haul_hydrogen"];
			}
		}
		sqlEnd($result);

		$this->data["metal"]	+= $metal;
		$this->data["silicon"]	+= $silicon;
		$this->data["hydrogen"] += $hydrogen;
		$this->data["oldmode"]	= (
			(isset( $this->data['oldmode'] ) && !empty( $this->data['oldmode'] ))
				? ($this->data['oldmode'])
				: EVENT_ATTACK_SINGLE
		);

		if( $this->mode == 1 || $this->data["xSkirmish"] ) // Attacker
		{
			if( !empty($this->userid) )
			{
				if(count($this->data["ships"]) - $artefacts > 0)
				{
					$this->EH->sendBack(
						time() + $this->data["time"], // + $this->time,
						$this->location,
						$this->userid,
						$this->planetid,
						$this->data,
						$this->parent_eventid
					);
				}
				else if($artefacts > 0) //fleet destroyed, dropping artefacts
				{
					foreach($this->data["ships"] as $typeid => $ship)
					{
						if($ship["mode"] == UNIT_TYPE_ARTEFACT)
						{
							$this->dropArtefacts($ship["art_ids"], $typeid, $owner);
						}
					}
				}
			}
		}
		else if($this->mode == 0 && $this->eventid > 0) // Defender
		{
			// Fleet contains more than just artefacts
			if( count($this->data["ships"]) - $artefacts > 0 )
			{
				// Updated By Pk
				sqlUpdate(
					"events",
					array(
						"data"		=> serialize($this->data),
						"prev_rc"	=> null,
						"processed" => EVENT_PROCESSED_WAIT
					),
					"eventid = ".sqlVal($this->eventid)
				);
			}
			else
			{
				if( $artefacts > 0 ) //fleet destroyed, dropping artefacts
				{
					foreach($this->data["ships"] as $typeid => $ship)
					{
						if($ship["mode"] == UNIT_TYPE_ARTEFACT)
						{
							$this->dropArtefacts($ship["art_ids"], $typeid, $owner);
						}
					}
				}
				// Updated By Pk
				Core::getQuery()->update("events", array("prev_rc", "processed", "processed_time", "error_message"),
					array(null, EVENT_PROCESSED_OK, time(), "killed, participantid: {$this->participantid}"),
					"eventid = ".sqlVal($this->eventid));
			}
		}

		if( !empty($this->location) && !empty($this->planetid) && !empty($this->userid) )
		{
			if(count($grasped_artefacts["ships"]) > 0)
			{
				$grasped_artefacts["destination"]	= $this->planetid;
				$grasped_artefacts["planetid"]		= $this->location;
				new AutoMsg(MSG_GRASPED_ARTEFACTS, $this->userid, time(), $grasped_artefacts);
			}
			if(count($grasped_data["ships"]) > 0)
			{
				$grasped_data["destination"] = $this->planetid;
				$grasped_data["planetid"] = $this->location;
				new AutoMsg(MSG_GRASPED_REPORT, $this->userid, time() /*$this->time*/, $grasped_data);
			}
		}

        if( !empty($this->userid) ){
            NS::updateUserDmPoints($this->userid);
        }
		return $this;
	}

	/**
	 * Sets transport_event for all artefacts after assault
	 * @param int $event_id SendBack event id
	 * @param array $arts_id Array of artefact IDs
	 */
	protected function setArtefactTransportEvent( $event_id, $arts_id = array() )
	{
		if( empty($arts_id) || !is_array($arts_id) )
		{
			return;
		}
		foreach( $arts_id as $id => $trash )
		{
			Artefact::setTransportID($id, $event_id);
		}
		return;
	}

	/**
	* Setter-method for event id.
	*
	* @param integer
	*
	* @return Participant
	*/
	public function setEventId($eventid)
	{
		$this->eventid = $eventid;
		return $this;
	}

	/**
	* Sets the participant fleet planet id
	*
	* @param integer
	*
	* @return Participant
	*/
	public function setPlanetId($planetid)
	{
		$this->planetid = $planetid;
		return $this;
	}

	/**
	* Set the combat planetid.
	*
	* @param integer	Planet id
	*
	* @return Participant
	*/
	public function setLocation($planetid)
	{
		$this->location = $planetid;
		return $this;
	}

	/**
	* Getter-method for the participant id.
	*
	* @return integer
	*/
	public function getParticipantId()
	{
		return $this->participantid;
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
	* Returns the mode.
	*
	* @return integer
	*/
	public function getMode()
	{
		return $this->mode;
	}

	/**
	* Returns the user id.
	*
	* @return integer
	*/
	public function getUserId()
	{
		return $this->userid;
	}

	/**
	* Returns the planet id.
	*
	* @return integer
	*/
	public function getPlanetId()
	{
		return $this->planetid;
	}

	public function getPrimaryTarget()
	{
		return $this->primary_target;
	}

	/**
	* Returns the time.
	*
	* @return integer
	*/
	public function getTime($time)
	{
		return $this->time;
	}

	/**
	* Sets the assault id.
	*
	* @param integer
	*
	* @return Participant
	*/
	public function setAssaultId($assaultid)
	{
		$this->assaultid = $assaultid;
		return $this;
	}

	/**
	* Sets the mode.
	*
	* @param integer
	*
	* @return Participant
	*/
	public function setMode($mode)
	{
		$this->mode = $mode;
		return $this;
	}

	/**
	* Sets the user id.
	*
	* @param integer
	*
	* @return Participant
	*/
	public function setUserId($userid)
	{
		$this->userid = $userid > 0 ? $userid : null;
		return $this;
	}

	/**
	* Sets the time.
	*
	* @param integer
	*
	* @return Participant
	*/
	public function setTime($time)
	{
		$this->time = $time;
		return $this;
	}

	public function setParentEventid($parent_eventid)
	{
		$this->parent_eventid = $parent_eventid;
		return $this;
	}

	public function setPrimaryTarget($primary_target)
	{
		$this->primary_target = $primary_target;
		return $this;
	}
}
?>