<?php
/**
* Creates new planets and moons.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class PlanetCreator
{
	/**
	* User who will get this planet.
	*
	* @var integer
	*/
	protected $userid = 0;

	/**
	* If planet will be a moon.
	*
	* @var boolean
	*/
	protected $moon = 0;

	/**
	* Moon formation percent.
	*
	* @var integer
	*/
	protected $percent = 0;

	/**
	* Galaxy where planet or moon is positionated.
	*
	* @var integer
	*/
	protected $galaxy = 0;

	/**
	* Solar system where planet or moon is positionated.
	*
	* @var integer
	*/
	protected $system = 0;

	/**
	* Planet position in solar system.
	*
	* @var integer
	*/
	protected $position = 0;

	/**
	* Planet name.
	*
	* @var string
	*/
	protected $name = "";

	/**
	* Planet diameter size.
	*
	* @var integer
	*/
	protected $size = 0;

	/**
	* Picture name.
	*
	* @var string
	*/
	protected $picture = "";

	/**
	* Planet temperature.
	*
	* @var integer
	*/
	protected $temperature = 0;

	/**
	* Id of created planet
	*
	* @var integer
	*/
	protected $planetid = 0;

	/**
	* Starts creator.
	*
	* @param integer	User id of this planet/moon
	* @param integer	Galaxy
	* @param integer	System
	* @param integer	Position
	* @param boolean	If planet is moon
	* @param integer	Moon formation percent
	*
	* @return void
	*/
	public function __construct($userid, $galaxy = null, $system = null, $position = null, $ismoon = 0, $percent = 0)
	{
		$this->userid	= $userid;
		$this->moon		= $ismoon;
		$this->percent	= $percent;

		$this->metal	= Core::getOptions()->get("DEFAULT_METAL");
		$this->silicon	= Core::getOptions()->get("DEFAULT_SILICON");
		$this->hydrogen = Core::getOptions()->get("DEFAULT_HYDROGEN");

		if($galaxy == null)
		{
			$this->setRandPos();
			$this->size = Core::getOptions()->get("HOME_PLANET_SIZE");
			$this->name = Core::getLanguage()->getItem("HOME_PLANET");

			$this->metal = Core::getOptions()->get("DEFAULT_START_METAL");
			$this->silicon = Core::getOptions()->get("DEFAULT_START_SILICON");
			$this->hydrogen = Core::getOptions()->get("DEFAULT_START_HYDROGEN");
		}
		else
		{
			$this->name = Core::getLanguage()->getItem("COLONY");
			$this->galaxy = $galaxy;
			$this->system = $system;
			$this->position = $position;
			if($this->moon)
			{
				$this->setRandMoonSize();
			}
			else
			{
				$this->setRandSize();
			}
		}
		$this->setRandTemperature();
		if(!$this->moon)
		{
			$this->setPicture();
		}
		$this->createPlanet();
	}

	/**
	* Sets a random position for this planet.
	*
	* @return PlanetCreator
	*/
	protected function setRandPos()
	{
		$log_str = 'in setRandPos. ';
		if( !$this->planetid && !$this->moon )
		{
			$this->planetid = sqlInsert("planet", array(
				"userid" => null,
				"ismoon" => $this->moon,
				"planetname" => "unknown",
				"diameter" => 8765,
				"temperature" => 1,
				"last" => time(),
				));
			$log_str .= 'Incerted Planet ' . $this->planetid . '. ';
			
			if( sqlSelectField("galaxy", "count(*)", "", "position <= ".MAX_NORMAL_PLANET_POSITION." AND destroyed = 1 AND moonid is NULL") > 10 )
			{			
				$galaxies = array();
				$res = sqlSelect("galaxy_new_active", "id");
				while($row = sqlFetch($res))
				{
					$galaxies[] = $row["id"];
				}
				sqlEnd($res);
				
				shuffle($galaxies);
				$log_str .= 'got ' . count($galaxies) . '. ';
				foreach($galaxies as $galaxy)
				{
					sqlUpdate("galaxy", array(
						"planetid" => $this->planetid,
						"metal" => 0,
						"silicon" => 0,
						"moonid" => null,
						"destroyed" => 0,
						), "galaxy = ".sqlVal($galaxy)." AND position <= ".MAX_NORMAL_PLANET_POSITION." AND destroyed = 1 AND moonid is NULL LIMIT 1");
					$log_str .= 'Updated . ';
					
					$row = sqlSelectRow("galaxy", "*", "", "planetid = ".sqlVal($this->planetid));
					if( $row )
					{
						$this->galaxy = $row["galaxy"];
						$this->system = $row["system"];
						$this->position = $row["position"];
						$log_str .= 'stoped at ' . $this->galaxy . ':' . $this->system . ':' . $this->position . '. ';
						// error_log($log_str, 'warning');
						return $this;
					}
				}
			}
			
			$log_str .= 'galaxy_new_pos_cut2 Way . ';
			/*
			$result = sqlSelect("galaxy_new_pos_cut2", "galaxy, system, position",
				"", // join
				"", // where
				"", // order
				"10");
			while( $row = sqlFetch($result) )
			*/
			for($i = 0; $i < 10; $i++)
			{
				try
				{
					sqlQuery('INSERT INTO '.PREFIX.'galaxy (galaxy, system, position, planetid)
						SELECT galaxy, system, position, '.sqlVal($this->planetid).'
						FROM '.PREFIX.'galaxy_new_pos_cut2
						WHERE galaxy between 1 and '.sqlVal(NUM_GALAXYS).'
							AND system between 1 and '.sqlVal(NUM_SYSTEMS).'
						LIMIT 1');
					
					/*
					sqlInsert("galaxy", array(
						"galaxy" => $row["galaxy"],
						"system" => $row["system"],
						"position" => $row["position"],
						"planetid" => $this->planetid,
						));
					*/

					$row = sqlSelectRow("galaxy", "*", "", "planetid = ".sqlVal($this->planetid));
					if( $row )
					{
						$this->galaxy = $row["galaxy"];
						$this->system = $row["system"];
						$this->position = $row["position"];
						$log_str .= 'found at ' . $this->galaxy . ':' . $this->system . ':' . $this->position . '. ';
						// error_log($log_str, 'warning');
						return $this;
					}
						
				} catch(Exception $e) {
				}
				usleep(50000);
			}
			sqlEnd($result);			
			
			sqlDelete("planet", "planetid = ".sqlVal($this->planetid)." AND userid is NULL");
			$this->planetid = 0;
		}
			

		$i = 0;
		$log_str .= 'Old Way . ';
		do {
			$is_valid_position = false;
			if(++$i <= 2)
			{
				$result = sqlSelect("galaxy_new_pos_cut2", "galaxy, system, position as free_pos",
					"", // join
					"", // where
					"", // order
					"1");
				$row = sqlFetch($result);
				$log_str .= 'Trying to get from galaxy_new_pos. ';
				if(!is_null($row["free_pos"]))
				{
					$this->galaxy = (int)$row["galaxy"];
					$this->system = (int)$row["system"];
					$this->position = (int)$row["free_pos"];
					$log_str .= 'Got ' . $this->galaxy . ':' . $this->system . ':' . $this->position . '. ';
					$is_valid_position = true;
				}
				sqlEnd($result);
			}
			if(!$is_valid_position)
			{
				$this->galaxy = mt_rand(1, NUM_GALAXYS);
				$this->system = mt_rand(1, NUM_SYSTEMS);
				$this->position = mt_rand(4, MAX_NORMAL_PLANET_POSITION);
				$log_str .= 'Making random ' . $this->galaxy . ':' . $this->system . ':' . $this->position . '. ';
			}
			$result_count = sqlSelectField("galaxy", "count(*)", "", "galaxy = ".sqlVal($this->galaxy)." AND system = ".sqlVal($this->system)." AND position = ".sqlVal($this->position));
			$is_exist = $result_count > 0;
			$log_str .= 'Random ' . ( $is_exist ? ('') : ('does not') ) . ' exists. ';
			
			if($i > 10)
			{
				break;
			}
			
		} while($is_exist);
		error_log($log_str); // PHP 8: 2-й аргумент должен быть int (0 = system log) — убрано legacy 'warning'
		return $this;
	}

	/**
	* Generates moon size.
	*
	* @return PlanetCreator
	*/
	protected function setRandMoonSize()
	{
		if($this->percent > 0)
		{
			$min = pow($this->percent, 2) * 10 + 4000;
			$this->size = mt_rand($min, $min + 1000);
		}
		else
		{
			$this->size = mt_rand(TEMP_MOON_SIZE_MIN, TEMP_MOON_SIZE_MAX);
		}
		return $this;
	}

	/**
	* Generates temperature.
	*
	* @return PlanetCreator
	*/

	public static function getTemperature($position, $percent = null)
	{
		if($percent == 99) // used for new user
		{
			$params = array( "start_pos" => 1, "end_pos" => MAX_NORMAL_PLANET_POSITION, "start_pos_temp" => array(10, 20), "end_pos_temp" => array(-30, -20) );
		}
		else if($position <= 3)
		{
			$params = array( "start_pos" => 1, "end_pos" => 3, "start_pos_temp" => array(80, 185), "end_pos_temp" => array(60, 120) );
		}
		else
		{
			$params = array( "start_pos" => 4, "end_pos" => MAX_NORMAL_PLANET_POSITION, "start_pos_temp" => array(20, 70), "end_pos_temp" => array(-90, -40) );
		}
		$position = clampVal( $position, $params["start_pos"], $params["end_pos"] );
		$len = $params["end_pos"] - $params["start_pos"];
		$t = $len > 0 ? (($position - $params["start_pos"]) / (float)$len) : 0.5;
		$min_temperature = $params["start_pos_temp"][0] + ($params["end_pos_temp"][0] - $params["start_pos_temp"][0]) * $t;
		$max_temperature = $params["start_pos_temp"][1] + ($params["end_pos_temp"][1] - $params["start_pos_temp"][1]) * $t;
		return randRoundRange( $min_temperature, $max_temperature );
	}
	
	protected function setRandTemperature()
	{
		$this->temperature = self::getTemperature($this->position, $this->percent);
		return $this;
	}

	/**
	* Generates planet size.
	*
	* @return PlanetCreator
	*/
	protected function setRandSize()
	{
		// max fields usage: top 17 players have fom 301 till 381 !
		if($this->percent == 100) // used for ARTEFACT_PLANET_CREATOR
		{
			$params = array(400, 450);
		}
		else if($this->percent == 99) // used for new user
		{
			$params = array(360, 380);
		}
		else if($this->percent == 5) // used for ARTEFACT_PLANET_TEMP_CREATOR
		{
			$params = array(50, 100);
		}
		else
		{
			if($this->position <= 3)
			{
				$params = array(150, 200);
			}
			else if($this->position >= round(MAX_NORMAL_PLANET_POSITION * 0.8))
			{
				$params = array(200, 420);
			}
			else
			{
				$params = array(250, 380);
			}
		}
		$this->size = round(sqrt(mt_rand(min(400, $params[0]), min(450, $params[1]))) * 1000);
		return $this;
	}

	/**
	* Generates picture.
	*
	* @return PlanetCreator
	*/
	protected function setPicture()
	{
		$planetenArt = new Map();
		$config = $this->getXMLConfig("PlanetPictures");
		foreach($config as $planetType)
		{
			$from = intval($planetType->getAttribute("from"));
			$to = intval($planetType->getAttribute("to"));
			if($this->position >= $from && $this->position <= $to)
			{
				$planetenArt->push(array(
					"name" => $planetType->getName(),
					"number" => $planetType->getInteger()
				));
			}
		}
		if( $planetenArt->size() == 0 )
		{
			$planetenArt->push(array(
				"name" => $planetType->getName(),
				"number" => $planetType->getInteger()
			));
		}
		$randomPlanet = $planetenArt->getRandomElement();
		$this->picture = sprintf("%s%02d", $randomPlanet["name"], mt_rand(1, $randomPlanet["number"]));
		return $this;
	}

	/**
	* Write final planet informations into database.
	*
	* @return PlanetCreator
	*/
	protected function createPlanet()
	{
		// Hook::event("PLANET_CREATOR_SAVE_PLANET", array(&$this));
		if($this->moon == 0)
		{
			$atts = array("userid", "planetname", "diameter", "picture", "temperature", "last", "metal", "silicon", "hydrogen", "solar_satellite_prod");
			$vals = array($this->userid, $this->name, $this->size, $this->picture, $this->temperature, time(), $this->metal, $this->silicon, $this->hydrogen, 100);
			
			if( !$this->planetid )
			{
				$rc = Core::getQuery()->insert("planet", $atts, $vals);
				// План 86: без проверки rc planetid указал бы на чужую
				// планету, и REPLACE INTO galaxy переписал бы координаты
				// чужой записи.
				if ($rc === false) {
					throw new Exception('Failed to create planet');
				}
				$this->planetid = Core::getDB()->insert_id();

				$atts = array("galaxy", "system", "position", "planetid");
				$vals = array($this->galaxy, $this->system, $this->position, $this->planetid);
				Core::getQuery()->replace("galaxy", $atts, $vals);
			}
			else
			{
				Core::getQuery()->update("planet", $atts, $vals, "planetid = ".sqlVal($this->planetid));
			}
			
			if($this->percent == 5) // ARTEFACT_PLANET_TEMP_CREATOR
			{
				
			}
		}
		else
		{
			$atts = array("userid", "ismoon", "planetname", "diameter", "picture", "temperature", "last", "metal", "silicon");
			$vals = array($this->userid, 1, Core::getLanguage()->getItem("MOON"), $this->size, "mond", $this->temperature - mt_rand(15, 35), time(), 0, 0);
			$rc = Core::getQuery()->insert("planet", $atts, $vals);
			// План 86: без проверки rc planetid указал бы на чужую
			// планету, и UPDATE galaxy ниже привязал бы moonid к
			// чужой записи.
			if ($rc === false) {
				throw new Exception('Failed to create moon');
			}
			$this->planetid = Core::getDB()->insert_id();

			Core::getQuery()->update("galaxy", "moonid", $this->planetid, "galaxy = ".sqlVal($this->galaxy)." AND system = ".sqlVal($this->system)." AND position = ".sqlVal($this->position));
		}
		
		AchievementsService::processAchievements($this->userid, $this->planetid);

		return $this;
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
	
	/**
	* Returns the planet name.
	*
	* @return String
	*/
	public function getPlanetName()
	{
		return $this->name;
	}
	
	/**
	* Returns the coordinates of the new planet.
	*
	* @return array
	*/
	public function getPosition()
	{
		return array(
			"galaxy" => $this->galaxy,
			"system" => $this->system,
			"position" => $this->position
			);
	}

	/**
	* Returns the XML config data.
	*
	* @param string	Config name
	*
	* @return XMLObj
	*/
	protected function getXMLConfig($name)
	{
		$config = new XML(APP_ROOT_DIR."game/xml/".$name.".xml");
		return $config->get();
	}
}
?>