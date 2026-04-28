<?php
/**
* Star surveillance: show planet's events.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class MonitorPlanet extends Page
{
	/**
	* Target planet id.
	*
	* @var integer
	*/
	protected $planetid = 0;

	/**
	* Target planet data.
	*
	* @var array
	*/
	protected $planetData = array();

	/**
	* Shows the fleet events of the stated planet.
	* Firstly, check for validation (Range, consumption, ...)
	*
	* @return void
	*/
	public function __construct()
	{
		$this->planetid = max(0, (int)Core::getRequest()->getGET("id"));

		$this->planetData = sqlSelectRow("galaxy g", array(
			"g.galaxy", "g.system", "g.position", "p.planetname", "u.username", "u.userid"),
			" INNER JOIN ".PREFIX."planet p ON p.planetid = g.planetid "
			. " LEFT JOIN ".PREFIX."user u ON u.userid = p.userid",
			"g.planetid = ".sqlVal($this->planetid));

		$this->proceedRequest();
	}

	/**
	* Checks for validation.
	*
	* @return MonitorPlanet
	*/
	protected function validate()
	{
		if(!$this->planetData)
		{
			throw new GenericException("The selected planet is unavailable.");
		}

		if(isAdmin())
		{
			return $this;
		}

		// Check range
		if($this->planetData["galaxy"] != NS::getPlanet()->getData("galaxy")
			|| NS::getPlanet()->getBuilding(UNIT_STAR_SURVEILLANCE) <= 0)
		{
			throw new GenericException("Range exceeded.");
		}

		$range = NS::getMonitorActivityRange();
		$diff = NS::getSystemsDiff($this->planetData["system"]); // abs(NS::getPlanet()->getData("system") - $this->planetData["system"]);
		if($range < $diff) // && ($this->galaxy == NS::getPlanet()->getData("galaxy")) )
		{
			throw new GenericException("Range exceeded.");
		}

		// Check consumption
		Core::getLanguage()->load(array("Galaxy", "Main", "info"));
		if(NS::getPlanet()->getData("hydrogen") < Core::getOptions()->get("STAR_SURVEILLANCE_CONSUMPTION"))
		{
			Logger::dieMessage("DEFICIENT_CONSUMPTION");
		}
		return $this;
	}

	/**
	* Subtracts hydrogen for surveillance consumption.
	*
	* @return MonitorPlanet
	*/
	protected function subtractHydrogen()
	{
		NS::updateUserRes(array(
			"type" => RES_UPDATE_MONITOR_PLANET,
			// "reload_planet" => false,
			"userid" => NS::getUser()->get("userid"),
			"planetid" => NS::getPlanet()->getPlanetId(),
			// "metal" => - $metal,
			// "silicon" => $silicon,
			"hydrogen" => - Core::getOptions()->get("STAR_SURVEILLANCE_CONSUMPTION"),
			// "max_metal" => NS::getPlanet()->getStorage("metal"),
			// "max_silicon" => NS::getPlanet()->getStorage("silicon"),
			// "max_hydrogen" => NS::getPlanet()->getStorage("hydrogen"),
			));

		// sqlQuery("UPDATE ".PREFIX."planet SET hydrogen = hydrogen - ".sqlVal(Core::getOptions()->get("STAR_SURVEILLANCE_CONSUMPTION"))." WHERE planetid = ".sqlPlanet());
		return $this;
	}

	/**
	* Index action.
	*
	* @return MonitorPlanet
	*/
	protected function index()
	{
		Core::getLanguage()->load(array("Main", "Galaxy", "info"));

		$this->validate()->subtractHydrogen();
		
		// Load events
		$events = array(); $i = 0;
		$joins	= 'LEFT JOIN ' . PREFIX . 'planet p1 ON p1.planetid = e.planetid ';
		$joins .= 'LEFT JOIN ' . PREFIX . 'planet p2 ON p2.planetid = e.destination ';
		$joins .= 'LEFT JOIN ' . PREFIX . 'galaxy g1 ON g1.planetid = e.planetid ';
		$joins .= 'LEFT JOIN ' . PREFIX . 'galaxy g2 ON g2.planetid = e.destination ';
		$joins .= 'LEFT JOIN ' . PREFIX . 'galaxy g1m ON g1m.moonid = e.planetid ';
		$joins .= 'LEFT JOIN ' . PREFIX . 'galaxy g2m ON g2m.moonid = e.destination ';
		$joins .= 'LEFT JOIN ' . PREFIX . 'user u1 ON u1.userid = p1.userid ';
		$joins .= 'LEFT JOIN ' . PREFIX . 'user u2 ON u2.userid = p2.userid';
		$select = array('e.*', 'u1.username', 'u2.username AS destname', 'p1.planetname', 'p2.planetname AS destplanet',
			'IFNULL(g1.galaxy, g1m.galaxy) as galaxy',
			'IFNULL(g1.system, g1m.system) as system',
			'IFNULL(g1.position, g1m.position) as position',
			'IFNULL(g2.galaxy, g2m.galaxy) as galaxy2',
			'IFNULL(g2.system, g2m.system) as system2',
			'IFNULL(g2.position, g2m.position) as position2',
			);
		$result = sqlSelect(
			'events e',
			$select,
			$joins,
			''
				. '('
					. ' (e.mode >= ' . EVENT_MARK_FIRST_FLEET . ' AND e.mode <= ' . EVENT_MARK_LAST_FLEET. ')'
					. ' OR e.mode = ' . EVENT_TEMP_PLANET_DISAPEAR
				. ')'
				. ' AND (e.planetid = ' . sqlVal($this->planetid) . ' OR e.destination = ' . sqlVal($this->planetid) . ')'
				. ' AND processed=' . EVENT_PROCESSED_WAIT,
			'e.time ASC'
		);
		while($row = sqlFetch($result))
		{
			$data			= unserialize($row["data"]);
			$is_event_owner	= ( $row["user"] == $this->planetData["userid"] );
			
			if(
					( $row["mode"] == EVENT_ROCKET_ATTACK													) // rocket attack
				||	( $row["mode"] == EVENT_RETURN			&& $row["destination"]	!= $this->planetid		) // return to not this planet
				||	( $row["mode"] == EVENT_POSITION		&& $row["destination"]	!= $this->planetid		) // position to not this planet
				||	( $row["mode"] == EVENT_DELIVERY_UNITS	&& $row["destination"]	!= $this->planetid		) // delivery to not this planet
				||	( $row["mode"] == EVENT_RECYCLING		&& $row["planetid"] 	!= $this->planetid 		)
				||	( $data["oldmode"] == EVENT_POSITION 													) // recycling from other planet
			)
			{
				continue;
			}

			// vars used in lang strings
			Core::getLanguage()->assign("planet", $row["planetname"]);
			Core::getLanguage()->assign("coords", $row["galaxy"] ? getCoordLink($row["galaxy"], $row["system"], $row["position"], false, $row["planetid"]) : getCoordLink($data["galaxy"], $data["system"], $data["position"], false, $row["planetid"]));
			Core::getLanguage()->assign("target", $row["mode"] != EVENT_RECYCLING ? $row["destplanet"] : Core::getLanguage()->getItem("DEBRIS"));
			Core::getLanguage()->assign("targetcoords", $row["galaxy2"] ? getCoordLink($row["galaxy2"], $row["system2"], $row["position2"], false, $row["destination"]) : getCoordLink($data["galaxy"], $data["system"], $data["position"], false, $row["destination"]));
			Core::getLanguage()->assign("username", $row["username"]);
			Core::getLanguage()->assign(
				"mission",
				$row["mode"] == EVENT_RETURN
					? NS::getMissionName($data["oldmode"], $is_event_owner)
					: NS::getMissionName($row["mode"]), $is_event_owner
			);

			$ships = "";
			if( is_array($data["ships"]) )
			{
				$ships = getUnitListStr($data["ships"]);
				Core::getLanguage()->assign("fleet", $ships);
			}

			if( $row["mode"] == EVENT_HOLDING )
			{
				$events[$i]["message"] = Core::getLanguage()->getItem("FLEET_MESSAGE_HOLDING_2");
			}
			elseif( $row["mode"] == EVENT_ALIEN_HOLDING )
			{
				$events[$i]["message"] = Core::getLanguage()->getItem("FLEET_MESSAGE_ALIEN_HOLDING");
			}
			elseif( $row["mode"] == EVENT_ALIEN_FLY_UNKNOWN || $row["mode"] == EVENT_ALIEN_GRAB_CREDIT 
						|| $row["mode"] == EVENT_ALIEN_ATTACK || $row["mode"] == EVENT_ALIEN_ATTACK_CUSTOM || $row["mode"] == EVENT_ALIEN_HALT )
			{
				$events[$i]["message"] = Core::getLanguage()->getItem("FLEET_MESSAGE_ALIEN_OTHER_2");
			}
			elseif( $row["mode"] == EVENT_RETURN && $is_event_owner )
			{
				$events[$i]["message"] = Core::getLanguage()->getItem("STAR_SUR_MSG_RETURN");
			}
			elseif( $row["mode"] == EVENT_TEMP_PLANET_DISAPEAR && $is_event_owner )
			{
				$coordinates = Link::get(
					"game.php/go:Mission/g:".$row["galaxy"]."/s:".$row["system"]."/p:".$row["position"],
					"[{$row['galaxy']}:{$row['system']}:{$row['position']}]"
				);
				Core::getLanguage()->assign("target_mission", $coordinates);
				$events[$i]["message"] = Core::getLanguage()->getItem("ANIGILATION_TIMER");
			}
			else
			{
				$events[$i]["message"] = Core::getLanguage()->getItem("STAR_SUR_MSG");
			}

			$events[$i]["eventid"]	= $row["eventid"];
			$events[$i]["class"]	= getFleetMessageClass($row["mode"], $is_event_owner);
			
			$events[$i]["time"]		= getTimeTerm($row["time"] - time());
			$events[$i]["time_r"]	= floor($row["time"] - time());
			
//			��� ���� ��� ����, ��� �� ���������� �� ��� ����� ������, � ����� ������ �� ������� ����. �� �������
//			if (1) // $row["destination"] == $this->planetid)
//			{
//				$events[$i]["time"]		= getTimeTerm($row["time"] - time());
//				$events[$i]["time_r"]	= floor($row["time"] - time());
//			}
//			else
//			{
//				$events[$i]["time"] = getTimeTerm($row["time"] + $data["time"] - time());
//				$events[$i]["time_r"] = floor($row["time"] + $data["time"] - time());
//			}
			$i++;
		}
		sqlEnd($result);
		if($i > 0)
		{
			Core::getLanguage()->load("info");
		}

		// vars used in lang strings
		Core::getLanguage()->assign("target", $this->planetData["planetname"]);
		Core::getLanguage()->assign("targetuser", $this->planetData["username"]);
		Core::getLanguage()->assign("targetcoords", getCoordLink($this->planetData["galaxy"], $this->planetData["system"], $this->planetData["position"], false, $this->planetid));

		Core::getTPL()->addLoop("events", $events);
		Core::getTPL()->assign("num_rows", $i);
		Core::getTPL()->display("monitor_planet", true);
		
		if( $this->planetData["userid"] != NS::getUser()->get("userid")
			&& NS::getResearch(UNIT_SPYWARE, $this->planetData["userid"]) >= NS::getResearch(UNIT_SPYWARE)/2 )
		{
			new AutoMsg(MSG_SURVEILLANCE_DETECTED, $this->planetData["userid"], time(), array(
						'galaxy' => $this->planetData["galaxy"], 
						'system' => $this->planetData["system"], 
						'position' => $this->planetData["position"], 
						'planetid' => $this->planetid,
					));
		}
		
		return $this;
	}
}
?>