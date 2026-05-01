<?php
/**
* Assault simulation.
*
* Oxsar http://oxsar.ru
*
*
*/

// define("DUMMY_BEGIN",1);
define("SIM_PLANET_ID", 1);
// define("SIM_USERS",2);
define("SIM_MAX_TIME", 20);
define("SIM_MAX_SIMS_NUMBER", 5);
define("SIM_DEF_SIMS_NUMBER", 1);

if(!defined("RECIPE_ROOT_DIR")) { die("Hacking attempt detected."); }

class Simulator extends Page
{
	var $ut_prefix = array("a_", "d_");

	/**
	* Constructor. Handles requests for this page.
	*
	* @return void
	*/
	public function __construct()
	{
		parent::__construct();
		Core::getTPL()->addHTMLHeaderFile("simulator.js?".CLIENT_VERSION, "js");
		if(!Core::getUser()->get("umode"))
		{
			$this
				->setPostAction("simulate", "simulate")
				->addPostArg("simulate", null);
		}

		Core::getLanguage()->load("info,mission,AssaultReport");
		$this->proceedRequest();
		return;
	}

	/**
	* Index action.
	*
	* @return Simulator
	*/
	protected function index($assaultid = 0, $num_sim = 0, $users = null)
	{
        if(NS::getUser()->get("umode")){
            Logger::dieMessage('UMODE_ENABLED');
        }
        if(NS::getUser()->get("observer") || (DEATHMATCH && (!isDeathmatchStarted() || isDeathmatchEnd()))){
            Logger::dieMessage('OBSERVER_MODE_ENABLED');
        }
		if(defined('TW')){
			Logger::dieMessage(Core::getLang()->getItem('TW_SIMULATOR'));
		}
		Core::getTPL()->assign("max_ships", MAX_SHIPS);
		Core::getTPL()->assign("max_ships_grade", MAX_SHIPS_GRADE);
		Core::getTPL()->assign("max_ships_size", (MAX_SHIPS_GRADE - 1));
		if(!$assaultid && !mt_rand(0, 1000)) // Core::getUser()->get("userid") == 2)
		{
			sqlQuery("DELETE FROM ".PREFIX."sim_res_log");
			sqlQuery("DELETE FROM ".PREFIX."sim_message");
			sqlQuery("DELETE FROM ".PREFIX."sim_user_experience");
			sqlQuery("DELETE FROM ".PREFIX."sim_research2user");
			sqlQuery("DELETE FROM ".PREFIX."sim_fleet2assault");
			sqlQuery("DELETE FROM ".PREFIX."sim_assault WHERE time<".sqlVal(time()-60*60*24*5));
			sqlQuery("DELETE FROM ".PREFIX."sim_user WHERE userid not in (select userid from ".PREFIX."sim_assaultparticipant)");
		}

		Core::getTPL()->assign("fleet_unit_type", UNIT_TYPE_FLEET);
		Core::getTPL()->assign("def_unit_type", UNIT_TYPE_DEFENSE);
		Core::getTPL()->assign("art_unit_type", UNIT_TYPE_ARTEFACT);

		$f = array();
		$first_fleet = null;
		$first_defense = null;
		$first_artefact = null;

		$result = sqlQuery("SELECT b.buildingid,b.name,b.mode,u2s.quantity,u2s.damaged,u2s.shell_percent,d.capicity,d.speed "
			. " FROM ".PREFIX."construction b "
			. " LEFT JOIN ".PREFIX."unit2shipyard u2s ON u2s.unitid = b.buildingid AND u2s.planetid = ".sqlPlanet()
			. " LEFT JOIN ".PREFIX."ship_datasheet d ON d.unitid = b.buildingid "
			. " WHERE b.mode in (".sqlArray(UNIT_TYPE_FLEET, UNIT_TYPE_DEFENSE).") ".(!isAdmin() ? " AND test = 0" : "")
			. " ORDER BY b.mode, b.display_order");
		while($row = sqlFetch($result))
		{
			$id = $row["buildingid"];

			$f[$id]["id"] = $id;
			$f[$id]["name"] = Link::get("game.php/UnitInfo/".$id, Core::getLanguage()->getItem($row["name"]));
			$f[$id]["quantity_raw"] = intval($row["quantity"]);
			$f[$id]["quantity"] = getUnitQuantityStr($row); // fNumber($row["quantity"]);
			$f[$id]["damaged"] = intval($row["damaged"]);
			$f[$id]["shell_percent"] = round($row["shell_percent"]);
			$f[$id]["unit_type"] = $row["mode"];
			$f[$id]["can_atter"] = intval($row["speed"] > 0 || $id == UNIT_INTERPLANETARY_ROCKET);
			$f[$id]["can_defense"] = intval($id != UNIT_INTERPLANETARY_ROCKET);
			$f[$id]["capicity_raw"] = $row["capicity"];

			if(is_null($first_fleet) && $row["mode"] == UNIT_TYPE_FLEET)
			{
				$f[$id]["first_ship"] = 1;
				$first_fleet = true;
			}
			if(is_null($first_defense) && $row["mode"] == UNIT_TYPE_DEFENSE)
			{
				$f[$id]["first_defense"] = 1;
				$first_defense = true;
			}

			foreach($this->ut_prefix as $prefix)
			{
				$f[$id][$prefix."quantity_before"] = $f[$id][$prefix."quantity_after"] = 0;
				$f[$id][$prefix."damaged_before"] = $f[$id][$prefix."damaged_after"] = 0;
				$f[$id][$prefix."percent_before"] = $f[$id][$prefix."percent_after"] = 100;
			}
		}
		sqlEnd($result);

		$result = sqlQuery("SELECT b.buildingid, b.name, a2u.artid "
			. " FROM ".PREFIX."construction b "
			. " LEFT JOIN ".PREFIX."artefact_datasheet ads ON b.buildingid = ads.typeid"
			. " LEFT JOIN ".PREFIX."artefact2user a2u ON a2u.typeid = b.buildingid AND a2u.deleted=0 AND a2u.planetid=".sqlPlanet()
			. " WHERE b.mode=".sqlVal(UNIT_TYPE_ARTEFACT)." AND ads.effect_type=".sqlVal(ARTEFACT_EFFECT_TYPE_BATTLE)
			. " ORDER BY b.mode, b.display_order");
		while($row = sqlFetch($result))
		{
			$id = $row["buildingid"];

			if(isset($f[$id]))
			{
				$f[$id]["quantity"] = ++$f[$id]["quantity_raw"];
			}
			else
			{
				$f[$id]["id"] = $id;
				$f[$id]["name"] = Link::get("game.php/ArtefactInfo/".$id, Core::getLanguage()->getItem($row["name"]));
				$f[$id]["quantity"] = $f[$id]["quantity_raw"] = isset($row["artid"]) ? 1 : 0;
				$f[$id]["damaged"] = 0;
				$f[$id]["shell_percent"] = 0;
				$f[$id]["unit_type"] = UNIT_TYPE_ARTEFACT;
				$f[$id]["can_atter"] = 1;
				$f[$id]["can_defense"] = 1;

				if(is_null($first_artefact))
				{
					$f[$id]["first_artefact"] = 1;
					$first_artefact = true;
				}

				foreach($this->ut_prefix as $prefix)
				{
					$f[$id][$prefix."quantity_before"] = $f[$id][$prefix."quantity_after"] = 0;
					$f[$id][$prefix."damaged_before"] = $f[$id][$prefix."damaged_after"] = 0;
					$f[$id][$prefix."percent_before"] = $f[$id][$prefix."percent_after"] = 100;
				}
			}
		}
		sqlEnd($result);

		$buildingid = 0;
		$building_level = 0;

		$num_sim = empty($num_sim) ? SIM_DEF_SIMS_NUMBER : clampVal(intval($num_sim), 1, SIM_MAX_SIMS_NUMBER);
		Core::getTPL()->assign("num_sim", $num_sim);
		Core::getTPL()->assign("assaultid", 0);
		if($assaultid > 0)
		{
			$atter_user = $users[0];
			$defender_user = $users[1];

			$row = sqlSelectRow("sim_assault", "*", "", "assaultid=".sqlVal($assaultid));
			if($row["accomplished"] == 1)
			{
				saveAssaultReportSID();
				if($row["target_buildingid"])
				{
					$buildingid = $row["target_buildingid"];
					$building_level = $row["building_level"];

					$construction = getConstructionDesc($row["target_buildingid"]);
					Core::getTPL()->assign("building_name", Core::getLanguage()->getItem($row["name"]));
				}
				Core::getTPL()->assign("adv_assault", $row["advanced_system"]);
				Core::getTPL()->assign("assaultid", $assaultid);
				Core::getTPL()->assign("report_link", socialUrl(FULL_URL . "AssaultReport.php?id=".$assaultid."&key=".$row["key"]."&simulation=1") );
				Core::getTPL()->assign("target_buildingid", $row["target_buildingid"]);
				Core::getTPL()->assign("building_level", $row["building_level"]);
				Core::getTPL()->assign("building_metal", $row["building_metal"]);
				Core::getTPL()->assign("building_silicon", $row["building_silicon"]);
				Core::getTPL()->assign("building_hydrogen", $row["building_hydrogen"]);
				Core::getTPL()->assign("building_destroy_chance", $row["building_destroy_chance"]);
				Core::getTPL()->assign("moonchance", $row["moonchance"]);
				Core::getTPL()->assign("attacker_lost_res", fNumber($row["attacker_lost_res"]));
				Core::getTPL()->assign("attacker_lost_metal", fNumber($row["attacker_lost_metal"]));
				Core::getTPL()->assign("attacker_lost_silicon", fNumber($row["attacker_lost_silicon"]));
				Core::getTPL()->assign("attacker_lost_hydrogen", fNumber($row["attacker_lost_hydrogen"]));
				Core::getTPL()->assign("defender_lost_res", fNumber($row["defender_lost_res"]));
				Core::getTPL()->assign("defender_lost_metal", fNumber($row["defender_lost_metal"]));
				Core::getTPL()->assign("defender_lost_silicon", fNumber($row["defender_lost_silicon"]));
				Core::getTPL()->assign("defender_lost_hydrogen", fNumber($row["defender_lost_hydrogen"]));
				Core::getTPL()->assign("debris_metal", fNumber($row["debris_metal"]));
				Core::getTPL()->assign("debris_silicon", fNumber($row["debris_silicon"]));
				Core::getTPL()->assign("recyclers", fNumber( ceil(($row["debris_metal"]+$row["debris_silicon"]) / $f[UNIT_RECYCLER]["capicity_raw"]) ));
				if($f[UNIT_SHIP_TRANSPLANTATOR]["capicity_raw"]){
					Core::getTPL()->assign("transplantators", fNumber( ceil(($row["debris_metal"]+$row["debris_silicon"]) / $f[UNIT_SHIP_TRANSPLANTATOR]["capicity_raw"]) ));
				}
				Core::getTPL()->assign("attacker_exp", fNumber($row["attacker_exp"]));
				Core::getTPL()->assign("defender_exp", fNumber($row["defender_exp"]));
				Core::getTPL()->assign("turns", $row["turns"]);
				Core::getTPL()->assign("turns_min", $row["turns_min"]);
				Core::getTPL()->assign("turns_max", $row["turns_max"]);
				Core::getTPL()->assign("attacker_win_percent", $row["attacker_win_percent"]);
				Core::getTPL()->assign("defender_win_percent", $row["defender_win_percent"]);
				Core::getTPL()->assign("draw_percent", $row["draw_percent"]);
				Core::getTPL()->assign("gentime_all", fNumber($row["gentime"] * $num_sim, 2));
				Core::getTPL()->assign("gentime", fNumber($row["gentime"], 2));
			}

			$result = sqlQuery("SELECT userid,unitid,quantity,damaged,shell_percent,grasped,mode,org_quantity,org_damaged,org_shell_percent "
				. " FROM ".PREFIX."sim_fleet2assault f "
				. " WHERE assaultid=".sqlVal($assaultid));
			while($row = sqlFetch($result))
			{
				$prefix = $this->ut_prefix[$row["userid"] == $defender_user];
				$f[$row["unitid"]][$prefix."quantity_before"] = $row["org_quantity"];
				$f[$row["unitid"]][$prefix."quantity_after"] = $row["quantity"];
				$f[$row["unitid"]][$prefix."damaged_before"] = $row["org_damaged"];
				$f[$row["unitid"]][$prefix."damaged_after"] = $row["damaged"];
				$f[$row["unitid"]][$prefix."percent_before"] = $row["org_shell_percent"];
				$f[$row["unitid"]][$prefix."percent_after"] = $row["shell_percent"];
				$f[$row["unitid"]][$prefix."grasped"] = $row["grasped"];
				$f[$row["unitid"]][$prefix."quantity_result"] = getUnitQuantityStr($row);
			}
			sqlEnd($result);

			$result = sqlQuery("SELECT r.buildingid,r.userid,r.level FROM ".PREFIX."sim_research2user r "
				. " INNER JOIN ".PREFIX."sim_assaultparticipant ap ON (r.userid=ap.userid) WHERE ap.assaultid=".sqlVal($assaultid));
			while($row = sqlFetch($result))
			{
				$prefix = $this->ut_prefix[$row["userid"] == $defender_user];
				Core::getTPL()->assign($prefix."tech_".$row["buildingid"], $row["level"]);
			}
			sqlEnd($result);
		}
		$this->resetAssault($assaultid, true, null);
		Core::getTPL()->addLoop("fleet", $f);

		$constructions = array();
		$constructions[] = array(
				"id" => 0,
				"name" => "",
				"selected" => $buildingid == 0,
			);
		$result = sqlSelect("construction", "*", "", "mode in (".sqlArray(UNIT_TYPE_CONSTRUCTION, UNIT_TYPE_MOON_CONSTRUCTION).")", "mode ASC, display_order ASC, buildingid ASC");
		while($row = sqlFetch($result))
		{
			$constructions[] = array(
				"id" => $row["buildingid"],
				"name" => Core::getLanguage()->getItem($row["name"]),
				"selected" => $row["buildingid"] == $buildingid,
			);
		}
		sqlEnd($result);
		Core::getTPL()->addLoop("constructions", $constructions);

		Core::getTPL()->display("simulator");
		return $this;
	}

	/**
	* Simulates an assault.
	*
	* @param array		Fleet to send
	*
	* @return Simulator
	*/
	protected function simulate($ships)
	{
		if( defined('TW') )
		{
			Logger::dieMessage(Core::getLang()->getItem('TW_SIMULATOR'));
		}
		if(!NS::isFirstRun("Simulator::simulate:".md5(serialize($ships))."-" . $_SESSION["userid"] ?? 0))
		{
			Logger::dieMessage('TOO_MANY_REQUESTS');
		}
		$users = array();
		foreach(array(__toUtf8(""), __toUtf8("")) as $i => $username)
		{
			$users[] = $cur_userid = sqlInsert("sim_user", array("username" => $username, "regtime" => time()));

			$aval_tech = array(UNIT_GUN_TECH, UNIT_SHIELD_TECH, UNIT_SHELL_TECH, UNIT_BALLISTICS_TECH, UNIT_MASKING_TECH,
				 UNIT_LASER_TECH, UNIT_ION_TECH, UNIT_PLASMA_TECH,
				 UNIT_SHIPYARD, UNIT_DEFENSE_FACTORY);

			foreach( $aval_tech as $researchid)
			{
				$tech = $this->ut_prefix[$i]."tech_".$researchid;
				if(!is_null($ships[$tech]))
				{
					sqlInsert("sim_research2user", array(
						"buildingid" => $researchid,
						"userid" => $cur_userid,
						"level" => intval($ships[$tech])));
					unset($ships[$tech]);
				}
			}
		}

		$num_sim = clampVal($ships["num_sim"], 1, SIM_MAX_SIMS_NUMBER);
		unset($ships["num_sim"]);

		$players = array();

		//Load ship stats
		$fleet_after = array();
		$stats = array();
		$select = array("d.unitid", "d.capicity", "d.speed", "d.consume", "b.name");
		$joins = "LEFT OUTER JOIN ".PREFIX."ship_datasheet d ON (d.unitid = b.buildingid)";
		$result = sqlSelect("construction b", $select, $joins,
			"b.mode = ".UNIT_TYPE_FLEET." OR b.mode = ".UNIT_TYPE_DEFENSE);
		while($row = sqlFetch($result))
		{
			$stats[$row["unitid"]]["name"] = $row["name"];
			$stats[$row["unitid"]]["capicity"] = $row["capicity"];
			$stats[$row["unitid"]]["speed"] = $row["speed"];
			$stats[$row["unitid"]]["consume"] = $row["consume"];
			foreach($users as $userid)
			{
				$fleet_after[$userid][$row["unitid"]]["quantity"] = 0;
				$fleet_after[$userid][$row["unitid"]]["damaged"] = 0;
				$fleet_after[$userid][$row["unitid"]]["shell_percent"] = 0;
				$fleet_after[$userid][$row["unitid"]]["grasped"] = 0;
			}
		}
		sqlEnd($result);

		foreach($users as $i => $userid)
		{
			foreach($ships as $id => $num)
			{
				if($num > 0 && preg_match("#^".preg_quote($this->ut_prefix[$i], "#")."(\d+)$#is", $id, $regs))
				{
					$unitid = $regs[1]; // substr($id, strpos($id, $this->ut_prefix[$i]) + 2);
					$num_capped = clampVal($num, 0, MAX_SHIPS);
					$players[$userid]["ships"][$unitid]["id"] = $unitid;
					$players[$userid]["ships"][$unitid]["quantity"] = $num_capped;
					$players[$userid]["ships"][$unitid]["damaged"] = clampVal($ships[$id."_d"], 0, $num_capped);
					$players[$userid]["ships"][$unitid]["shell_percent"] = clampVal($ships[$id."_p"], 0, 100);
					$players[$userid]["ships"][$unitid]["name"] = $stats[$unitid]["name"];
				}
			}
			$players[$userid]["galaxy"] = $players[$userid]["system"] = $players[$userid]["position"] = 0;
			$players[$userid]["sgalaxy"] = $players[$userid]["ssystem"] = $players[$userid]["sposition"] = 0;
			$players[$userid]["maxspeed"] = 0;
			$players[$userid]["consumption"] = 0;
			$players[$userid]["metal"] = $players[$userid]["silicon"] = $players[$userid]["hydrogen"] = 0;
			$players[$userid]["capacity"] = 0;
			$players[$userid]["time"] = 0;
			$players[$userid]["mode"] = intval($this->ut_prefix[$i] == "a_");
		}

		$target_building = null;
		$ships["buildingid"] = max(0, (int)$ships["buildingid"]);
		$ships["building_level"] = max(0, (int)$ships["building_level"]);
		if($ships["buildingid"] > 0 && $ships["building_level"] > 0)
		{
			$target_building = Assault::getTargetBuildingRes($ships["buildingid"], $ships["building_level"]);
		}

		require_once(APP_ROOT_DIR."game/Assault.class.php");
		$assaultid = sqlInsert("sim_assault", array(
			"planetid" 			=> SIM_PLANET_ID,
			"time" 				=> time(),
			"target_buildingid" 	=> $target_building ? $target_building["id"] : null,
			"building_level" 		=> $target_building ? $target_building["level"] : null,
			"building_metal" 		=> $target_building ? $target_building["metal"] : null,
			"building_silicon" 	=> $target_building ? $target_building["silicon"] : null,
			"building_hydrogen" 	=> $target_building ? $target_building["hydrogen"] : null,
			"advanced_system"		=> (($ships['new_ass'] == 1) ? 1: 0),
			));

		// DB access for external Java sim — константы из bd_connect_info.php.
		//Simulation
		$win = array(0 => 0, 1 => 0, 2 => 0);
		$this->prepare($building_destroy_chance);
		$this->prepare($moonchance);
		$this->prepare($attacker_lost_metal);
		$this->prepare($attacker_lost_silicon);
		$this->prepare($attacker_lost_hydrogen);
		$this->prepare($defender_lost_metal);
		$this->prepare($defender_lost_silicon);
		$this->prepare($defender_lost_hydrogen);
		$this->prepare($debris_metal);
		$this->prepare($debris_silicon);
		$this->prepare($attacker_exp);
		$this->prepare($defender_exp);
		$this->prepare($turns);
		$this->prepare($gentime);

		if(count($players[$users[0]]["ships"]) > 0 && count($players[$users[1]]["ships"]) > 0)
		{
			$start = microtime(true);
			$this->addParticipant($assaultid, $users[0], $players);
			$this->addParticipant($assaultid, $users[1], $players);

			$eventid = null;
			for($i = 0; $i < $num_sim; $i++)
			{
				if($i > 0)
				{
					$this->resetAssault($assaultid, false, $target_building);
				}

				if( 0 && defined('YII_CONSOLE_IS_RUNNING') && YII_CONSOLE_IS_RUNNING )
				{
					if(empty($eventid))
					{
						$eventid = NS::getEH()->addEvent(EVENT_RUN_SIM_ASSAULT, time(), null, NS::getUser()->get("userid"), null,
									array("assaultid" => $assaultid, "quantity" => $num_sim));
					}
					else
					{
						sqlUpdate(
							"events",
							array("processed" => EVENT_PROCESSED_WAIT, "prev_rc" => null, "time" => time()),
							"eventid=".sqlVal($eventid)
								. ' ORDER BY eventid'
						);
					}

					for($wait_step = 0;; $wait_step++)
					{
						usleep($wait_step < 5 ? 300000 : ($wait_step < 8 ? 500000 : 1000000));
						$row = sqlSelectRow("events", array("processed"), "",
									"eventid=".sqlVal($eventid)
									. " AND processed in (".sqlArray(EVENT_PROCESSED_OK, EVENT_PROCESSED_ERROR).")");
						if($row)
						{
							break;
						}
					}

				}
				else
				{
					// Start assault
					try
					{
						$temp_array = array(
							DB_HOST,
							DB_USER,
							DB_PWD,
							DB_NAME,
							DB_PREFIX."sim_",
							$assaultid
						);
						// -cp вместо -jar: нужен Assault.jar + mysql-connector
						// в classpath (см. bd_connect_info.php).
						$classpath = APP_ROOT_DIR.'game/'.SIMULATOR_ASSAULT_JAR.':'.MYSQL_CONNECTOR_JAR;
						$s = '/usr/bin/java -cp '.escapeshellarg($classpath).' assault.Assault "' . implode('" "', $temp_array) . '"';
						exec( $s );
					}
					catch(Exception $e)
					{
						;
					}
				}

				$row = sqlSelectRow("sim_assault",
					array("result",
					"target_buildingid", "building_level", "building_metal", "building_silicon", "building_hydrogen", "target_destroyed",
					"moonchance", "accomplished",
					"attacker_lost_metal", "attacker_lost_silicon", "attacker_lost_hydrogen",
					"defender_lost_metal", "defender_lost_silicon", "defender_lost_hydrogen",
					"debris_metal", "debris_silicon",
					"attacker_exp", "defender_exp", "turns", "gentime"),
					"", "assaultid = ".sqlVal($assaultid));
				if(!$row["accomplished"])
				{
					Logger::addMessage("Error while simulating battle (".$assaultid.").");
				}
				updateUserState($_SESSION["userid"] ?? 0, STATE_ASSAULT_SIMULATION);

				$win[$row["result"]]++;
				$this->processAvgMinMax($building_destroy_chance, $row["target_destroyed"] ? 100 : 0);
				$this->processAvgMinMax($moonchance, $row["moonchance"]);
				$this->processAvgMinMax($attacker_lost_metal, $row["attacker_lost_metal"]);
				$this->processAvgMinMax($attacker_lost_silicon, $row["attacker_lost_silicon"]);
				$this->processAvgMinMax($attacker_lost_hydrogen, $row["attacker_lost_hydrogen"]);
				$this->processAvgMinMax($defender_lost_metal, $row["defender_lost_metal"]);
				$this->processAvgMinMax($defender_lost_silicon, $row["defender_lost_silicon"]);
				$this->processAvgMinMax($defender_lost_hydrogen, $row["defender_lost_hydrogen"]);
				$this->processAvgMinMax($debris_metal, $row["debris_metal"]);
				$this->processAvgMinMax($debris_silicon, $row["debris_silicon"]);
				$this->processAvgMinMax($attacker_exp, $row["attacker_exp"]);
				$this->processAvgMinMax($defender_exp, $row["defender_exp"]);
				$this->processAvgMinMax($turns, $row["turns"]);
				$this->processAvgMinMax($gentime, $row["gentime"]);

				$result = sqlSelect("sim_fleet2assault",
					array("userid", "unitid", "quantity", "damaged", "shell_percent", "grasped"),
					"", "assaultid=".sqlVal($assaultid));
				while($row = sqlFetch($result))
				{
					$fleet_after[$row["userid"]][$row["unitid"]]["quantity"] += $row["quantity"];
					$fleet_after[$row["userid"]][$row["unitid"]]["damaged"] += $row["damaged"];
					$fleet_after[$row["userid"]][$row["unitid"]]["shell_percent"] += $row["shell_percent"];
					$fleet_after[$row["userid"]][$row["unitid"]]["grasped"] += $row["grasped"];
				}
				sqlEnd($result);

				if(microtime(true) - $start > SIM_MAX_TIME)
				{
					$i++;
					break;
				}
			}
			$save_num_sim = $num_sim;

			$num_sim = $i;
			$win[0] = round(100 * $win[0] / $num_sim);
			$win[1] = round(100 * $win[1] / $num_sim);
			$win[2] = round(100 * $win[2] / $num_sim);

			$building_destroy_chance["avg"] /= $num_sim;
			$moonchance["avg"] /= $num_sim;
			$attacker_lost_metal["avg"] /= $num_sim;
			$attacker_lost_silicon["avg"] /= $num_sim;
			$attacker_lost_hydrogen["avg"] /= $num_sim;
			$defender_lost_metal["avg"] /= $num_sim;
			$defender_lost_silicon["avg"] /= $num_sim;
			$defender_lost_hydrogen["avg"] /= $num_sim;
			$debris_metal["avg"] /= $num_sim;
			$debris_silicon["avg"] /= $num_sim;
			$attacker_exp["avg"] /= $num_sim;
			$defender_exp["avg"] /= $num_sim;
			$turns["avg"] /= $num_sim;
			$gentime["avg"] = round($gentime["avg"] / $num_sim) / 1000;
			$gentime["min"] /= 1000;
			$gentime["max"] /= 1000;

			sqlUpdate(
				"sim_assault",
				array(
					"building_destroy_chance" => $target_building ? $building_destroy_chance["avg"] : null,
					"moonchance" => $moonchance["avg"],
					"attacker_lost_res" => $attacker_lost_metal["avg"] + $attacker_lost_silicon["avg"],
					"attacker_lost_metal" => $attacker_lost_metal["avg"],
					"attacker_lost_silicon" => $attacker_lost_silicon["avg"],
					"attacker_lost_hydrogen" => $attacker_lost_hydrogen["avg"],
					"defender_lost_res" => $defender_lost_metal["avg"] + $defender_lost_silicon["avg"],
					"defender_lost_metal" => $defender_lost_metal["avg"],
					"defender_lost_silicon" => $defender_lost_silicon["avg"],
					"defender_lost_hydrogen" => $defender_lost_hydrogen["avg"],
					"debris_metal" => $debris_metal["avg"],
					"debris_silicon" => $debris_silicon["avg"],
					"attacker_exp" => $attacker_exp["avg"],
					"defender_exp" => $defender_exp["avg"],
					"turns" => $turns["avg"],
					"turns_min" => $turns["min"],
					"turns_max" => $turns["max"],
					"gentime" => $gentime["avg"],
					"attacker_win_percent" => $win[1],
					"defender_win_percent" => $win[2],
					"draw_percent" => $win[0],
				),
				"assaultid=".sqlVal($assaultid)
					. ' ORDER BY assaultid '
			);

			foreach($fleet_after as $userid => $fleet)
			{
				foreach($fleet as $id => $ship)
				{
					$ship["quantity"] = round($ship["quantity"] / $num_sim);
					$ship["damaged"] = round($ship["damaged"] / $num_sim);
					$ship["shell_percent"] = $ship["shell_percent"] / $num_sim;
					$ship["grasped"] = round($ship["grasped"] / $num_sim);

					sqlUpdate(
						"sim_fleet2assault",
						array(
							"quantity" => $ship["quantity"],
							"damaged" => $ship["damaged"],
							"shell_percent" => $ship["shell_percent"],
							"grasped" => $ship["grasped"]
						),
						"assaultid=".sqlVal($assaultid)." AND userid=".sqlVal($userid)." AND unitid=".sqlVal($id)
							. ' ORDER BY assaultid DESC, userid DESC, unitid DESC '
					);
				}
			}
		}
		else if((count($players[$users[0]]["ships"]) == 0) && (count($players[$users[1]]["ships"]) > 0))
		{
			Logger::addMessage("SIM_NO_ATTER_UNITS");

			// $this->addParticipant($assaultid, $users[1], $players);
			sqlUpdate(
				"sim_assault",
				array(
					"accomplished" => 2,
					"attacker_win_percent" => 0,
					"defender_win_percent" => 100,
					"draw_percent" => 0,
					// "sim_report" => __toUtf8("Победа атакующего: 0%, победа обороняющегося: 100%, ничья: 0%")
				),
				"assaultid=".sqlVal($assaultid)
					. ' ORDER BY assaultid '
			);
		}
		else if((count($players[$users[0]]["ships"]) > 0) && (count($players[$users[1]]["ships"]) == 0))
		{
			Logger::addMessage("SIM_NO_DEFENDER_UNITS");

			// $this->addParticipant($assaultid, $users[0], $players);
			sqlUpdate(
				"sim_assault",
				array(
					"accomplished" => 2,
					"attacker_win_percent" => 100,
					"defender_win_percent" => 0,
					"draw_percent" => 0,
					// "sim_report" => __toUtf8("Победа атакующего: 100%, победа обороняющегося: 0%, ничья: 0%")
				),
				"assaultid=".sqlVal($assaultid)
					. ' ORDER BY assaultid '
			);
		}
		else if((count($players[$users[0]]["ships"]) == 0) && (count($players[$users[1]]["ships"]) == 0))
		{
			Logger::addMessage("SIM_NO_UNITS");

			sqlUpdate(
				"sim_assault",
				array(
					"accomplished" => 2,
					"attacker_win_percent" => 0,
					"defender_win_percent" => 0,
					"draw_percent" => 100,
					// "sim_report" => __toUtf8("Победа атакующего: 0%, победа обороняющегося: 0%, ничья: 100%")
				),
				"assaultid=".sqlVal($assaultid)
					. ' ORDER BY assaultid '
			);
		}

		return $this->index($assaultid, $save_num_sim, $users);
	}

	/**
	* Clears simulation data from tables
	*
	* @param integer	Id
	* @param bool	full deletion
	*/
	protected function resetAssault($assaultid, $full, $target_building)
	{
		if($full)
		{
			;
		}
		else
		{
			sqlQuery("UPDATE ".PREFIX."sim_fleet2assault SET quantity=org_quantity, damaged=org_damaged, shell_percent=org_shell_percent WHERE assaultid=".sqlVal($assaultid). ' ORDER BY assaultid');
			sqlUpdate(
				"sim_assault",
				array(
					"accomplished" => 0,
					"key" => "",
					"result" => "",
					"target_destroyed" => null, // 0
					"moonchance" => "",
					"moon" => "",
					"attacker_lost_res" => "",
					"attacker_lost_metal" => "",
					"attacker_lost_silicon" => "",
					"attacker_lost_hydrogen" => "",
					"defender_lost_res" => "",
					"defender_lost_metal" => "",
					"defender_lost_silicon" => "",
					"defender_lost_hydrogen" => "",
					"debris_metal" => "",
					"debris_silicon" => "",
					"attacker_exp" => "",
					"defender_exp" => "",
					"turns" => "",
					"gentime" => "",
					"report" => "",
					"message" => ""
				),
				"assaultid=".sqlVal($assaultid)
					. ' ORDER BY assaultid '
			);
		}
	}

	/**
	* Similar to Assault::addParticipant
	*
	* @param integer	assault Id
	* @param integer	user Id
	* @param array	players info
	*/
	protected function addParticipant($assaultid, $userid, $players)
	{
		$rc = Core::getQuery()->insert("sim_assaultparticipant",
			array("assaultid", "userid", "planetid", "mode", "consumption", "preloaded"),
			array($assaultid, $userid, SIM_PLANET_ID, $players[$userid]["mode"], $players[$userid]["consumption"], 0));
		// План 86: без проверки rc следующий INSERT в sim_fleet2assault
		// получил бы lastInsertId() от ПРЕДЫДУЩЕГО успешного INSERT
		// (например sim_assault.assaultid) → каскадный FK violation.
		if ($rc === false) {
			return;
		}
		$participantid = Core::getDB()->insert_id();
		foreach($players[$userid]["ships"] as $ship)
		{
			fixUnitDamagedVars($ship);
			Core::getQuery()->insert("sim_fleet2assault",
				array("assaultid", "participantid", "userid", "unitid", "mode",
				"quantity", "damaged", "shell_percent",
				"org_quantity", "org_damaged", "org_shell_percent"),
				array($assaultid, $participantid, $userid, $ship["id"], $players[$userid]["mode"],
				min( $ship["quantity"], MAX_SHIPS ) , $ship["damaged"], $ship["shell_percent"],
				min( $ship["quantity"], MAX_SHIPS ) , $ship["damaged"], $ship["shell_percent"])
				);
		}
	}

	private function prepare(&$stats)
	{
		$stats["avg"] = 0;
		$stats["min"] = null;
		$stats["max"] = null;
	}

	private function processAvgMinMax(&$stats, $data)
	{
		$stats["avg"] += $data;
		if(is_null($stats["min"]) || ($stats["min"] > $data)) $stats["min"] = $data;
		if(is_null($stats["max"]) || ($stats["max"] < $data)) $stats["max"] = $data;
	}
}

function __toUtf8($s)
{
	return $s;
}

?>