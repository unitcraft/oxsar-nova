<?php
/**
* Main page, contains overview and planet preferences.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class Main extends Page
{
	/**
	* Holds building events for this user.
	*
	* @var array
	*/
	protected $buildingEvent = array();

	/**
	* Holds all fleet events (own events and other events heading to user's planets)
	*
	* @var array
	*/
	protected $fleetEvent = array();

	/**
	* Shows account and planet overview.
	*
	* @return void
	*/
	public function __construct()
	{
		Core::getLanguage()->load(array("Main", "info", "mission"));
		parent::__construct();

		$this
			->setPostAction("changeplanetoptions", "changePlanetOptions")
			->addPostArg("changePlanetOptions", "planetname")
			->addPostArg("changePlanetOptions", "abandon")
			->addPostArg("changePlanetOptions", "password")
			->addPostArg("changePlanetOptions", "ishomeplanet")
			->addPostArg("changePlanetOptions", "leave")
			->addPostArg("changePlanetOptions", "editplanetid")
			->setPostAction("retreat", "retreatFleet")
			->addPostArg("retreatFleet", "id")
			->setGetAction("go", "PlanetOptions", "planetOptions")
			->setGetAction("go", "HomePlanetRequired", "homePlanetRequired")
			;

		$this->proceedRequest();
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
		doHeaderRedirection("game.php/Main", false);
		return $this;
	}

	protected function homePlanetRequired()
	{
		return $this->index();
	}

	/**
	* Index action.
	*
	* @return Main
	*/
	protected function index()
	{
		// Из ExtMain: инициализация jclock + serverClock TPL var
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
		Core::getTPL()->assign("serverClock", date("d.m.Y", time())." ".$clock);

		Core::getTPL()->addHTMLHeaderFile("lib/jquery.countdown.js?".CLIENT_VERSION, "js");
		$this->buildingEvent = NS::getEH()->getFirstBuildingEvent();

		$donated = true;
		/* if(NS::getUser()->get("regtime") < time()-60*60*24*30)
		{
			$row = sqlSelectRow("payments", "sum(pay_credit) as credit", "",
				"pay_user_id=".sqlUser()." AND pay_date > ".sqlVal(date("Y-m-d", time()-60*60*24*14))." AND pay_status=1");
			$donated = $row["credit"] > 0;
		} */
		Core::getTPL()->assign("donated", $donated);
        Core::getTPL()->assign("universe_name_full", (defined('UNIVERSE_NAME_FULL') ? UNIVERSE_NAME_FULL : 'Origin'));

        // mini_games iframe (7j7.ru) удалён — внешний сервис не используется в oxsar-nova

		if( NS::getPlanet()->getPlanetId() )
		{
			Core::getTPL()->assign("homeplanet", getCoordLink(false, false, false, "select", NS::getUser()->get("hp"), true));
			
			$reseach_virt_lab = fNumber(NS::getPlanet()->getReseachVirtLab(), 2);
			$umi_planets = NS::getPlanet()->getData("umi-planets");
			if(is_array($umi_planets)){
				$umi_info = array();
				foreach($umi_planets as $row){
					$umi_info[] = $row["umi"]
						. (empty($row["galaxy"]) ? "" : " " . getCoordLink($row["galaxy"], $row["system"], $row["position"]));
				}
				if(count($umi_info)-1 < NS::getPlanet()->getData("umi-arts")){
					$str = Core::getLanguage()->getItem("ASK_ALLY_USER_ART_IGN");
					$link = Link::get("game.php/ChatAlly", "...", $str);
					$umi_info[] = "<nobr> $link </nobr>";
				}else{
					$str = Core::getLanguage()->getItem("ADD_ART_ALLY_IGN");
					$link = Link::get("game.php/ArtefactMarket", "...", $str);
					$umi_info[] = "<nobr> $link </nobr>";
				}
				$reseach_virt_lab .= " = (".join(" + ", $umi_info).")"
					// . " x (1 + ".fNumber(NS::getPlanet()->getData("umi-ign")/20, 1)
					// . " + ".fNumber(NS::getPlanet()->getData("umi-arts")/10, 1).")"
					." x ".fNumber(NS::getPlanet()->getData("umi-x"), 2)
					;
			}
			Core::getTPL()->assign("reseachVirtLab", $reseach_virt_lab);
		}
		else
		{
            if(!DEATHMATCH){
                Logger::addMessage("HOME_PLANET_REQUIRED");
            }
			Core::getTPL()->assign("homeplanet", Core::getLanguage()->getItem("PLANET_IS_NOT_FORMED_YET"));
		}

		// Messages
		$result_count = sqlSelectField("message", "count(*)", "", "receiver = ".sqlUser()." AND readed = '0'");
		$msgs = $result_count;
		Core::getTPL()->assign("unreadedmsg", $msgs);
		if($msgs > 0)
		{
			if($msgs > 1)
			{
				$new_mess_text = sprintf(Core::getLanguage()->getItem("F_NEW_MESSAGES"), $msgs);
			}
			else
			{
				$new_mess_text = Core::getLanguage()->getItem("F_NEW_MESSAGE");
			}
			Core::getTPL()->assign("newMessages", Link::get("game.php/MSG", $new_mess_text));
		}

		// Fleet events
		$fleetEvent = NS::getEH()->getFleetEvents();
		$fe = array();
		$fe_ally = array();
		if($fleetEvent)
		{
			foreach($fleetEvent as $f)
			{
				$is_ally_attack = $f["mode"] == EVENT_ALLIANCE_ATTACK_ADDITIONAL && isset($f["data"]["alliance_attack"]["eventid"]);
				if(!$is_ally_attack || !isset($fe_ally[$f["data"]["alliance_attack"]["eventid"]]))
				{
					$fe[$f["eventid"]] = $this->parseEvent($f);
					if(!is_array($fe[$f["eventid"]]))
					{
						unset($fe[$f["eventid"]]);
					}
					else if($is_ally_attack)
					{
						$fe_ally[$f["data"]["alliance_attack"]["eventid"]] = 1;
					}
				}
			}
		}
		Core::getTPL()->addLoop("fleetEvents", $fe);

		if( false ) // isMobileSkin())
		{
			$big_planet_img_width = "70px";
			$small_planet_img_width = "50px";
			$moon_planet_img_width = "30px";
		}
		else
		{
			$big_planet_img_width = "200px";
			$small_planet_img_width = "89px";
			$moon_planet_img_width = "50px";
		}

		Core::getTPL()->assign("serverTime", Date::timeToString(1, time(), "", false));
		$planetAction = Core::getLanguage()->getItem(($this->buildingEvent) ? $this->buildingEvent["data"]["buildingname"] : "PLANET_FREE");
		if($this->buildingEvent)
		{
			$timeleft = $this->buildingEvent["time"] - time();
			$timer = "<script type=\"text/javascript\">
				$(function () {
					$('#bCountDown').countdown({until: ".$timeleft.", compact: true, onExpiry: function() {
						$('#bCountDown').text('-');
					}});
				});
			</script>
			<span id=\"bCountDown\">".getTimeTerm($timeleft)."<br />".$abort."</span>";
		}
		else { $timer = ""; }
		$planetAction = Link::get("game.php/Constructions", $planetAction).$timer;
		Core::getTPL()->assign("planetAction", $planetAction); unset($planetAction);
		Core::getTPL()->assign("control_action", socialUrl(RELATIVE_URL . "game.php/Mission"));
		Core::getTPL()->assign("occupiedFields", NS::getPlanet()->getFields(true));
		Core::getTPL()->assign("planetImage", Image::getImage("planets/".NS::getPlanet()->getData("picture").Core::getConfig()->get("PLANET_IMG_EXT"), NS::getPlanet()->getData("planetname"), $big_planet_img_width, $big_planet_img_width));
		$maxFields = NS::getPlanet()->getMaxFields();
		$maxFields2 = NS::getPlanet()->getMaxFields(true);
		if($maxFields < $maxFields2)
		{
			$maxFields .= "-".$maxFields2;
		}
		Core::getTPL()->assign("maxFields", $maxFields);
		Core::getTPL()->assign("planetDiameter", fNumber(NS::getPlanet()->getData("diameter")));
		Core::getTPL()->assign("planetNameLink", Link::get("game.php/PlanetOptions", NS::getPlanet()->getData("planetname")));
		Core::getTPL()->assign("planetPosition", NS::getPlanet()->getCoords());
		Core::getTPL()->assign("planetTemp", NS::getPlanet()->getData("temperature"));
		Core::getTPL()->assign("points", Link::get("game.php/Ranking", pNumber(NS::getUser()->get("points"))));
		Core::getTPL()->assign("e_points", Link::get("game.php/Ranking/id:e_points", pNumber(NS::getUser()->get("e_points"))));
		Core::getTPL()->assign("be_points", pNumber(NS::getUser()->get("be_points")));

        if(DEATHMATCH){
            Core::getTPL()->assign("dm_points_enabled", true);
            Core::getTPL()->assign("dm_points", Link::get("game.php/Ranking/id:dm_points", pNumber(NS::getUser()->get("dm_points"))));

            Core::getTPL()->assign("max_points_enabled", true);
            Core::getTPL()->assign("max_points", Link::get("game.php/Ranking/id:max_points", pNumber(NS::getUser()->get("max_points"))));
        }

		// Points
		$totalUsers = sqlSelectField("user", "count(*)");
		$rank = sqlSelectField("user u", "count(*)", "", "points >= ".sqlVal(NS::getUser()->get("points"))
            . " AND u.observer = 0 "
            . " AND u.umode = 0 AND u.userid NOT IN (SELECT userid FROM ".PREFIX."ban_u)"
        );

		Core::getLang()->assign("totalUsers", $totalUsers);
		Core::getLang()->assign("rank", $rank);

		if(NS::getPlanet()->getData("moonid") > 0)
		{
			if(NS::getPlanet()->getData("ismoon"))
			{
				// Planet has moon
				$row = sqlSelectRow("galaxy g", array("p.planetid", "p.planetname", "p.picture"), "LEFT JOIN ".PREFIX."planet p ON (p.planetid = g.planetid)", "g.galaxy = ".sqlVal(NS::getPlanet()->getData("moongala"))." AND g.system = ".sqlVal(NS::getPlanet()->getData("moonsys"))." AND g.position = ".sqlVal(NS::getPlanet()->getData("moonpos")));
			}
			else
			{
				// Planet of current moon
				$row = sqlSelectRow("galaxy g", array("p.planetid", "p.planetname", "p.picture"), "LEFT JOIN ".PREFIX."planet p ON (p.planetid = g.moonid)", "g.galaxy = ".sqlVal(NS::getPlanet()->getData("galaxy"))." AND g.system = ".sqlVal(NS::getPlanet()->getData("system"))." AND g.position = ".sqlVal(NS::getPlanet()->getData("position")));
			}
			Core::getTPL()->assign("moon", $row["planetname"]);
			$img = Image::getImage("planets/".$row["picture"].Core::getConfig()->get("PLANET_IMG_EXT"), $row["planetname"], $moon_planet_img_width, $moon_planet_img_width);
			Core::getTPL()->assign("moonImage", "<a title=\"".$row["planetname"]."\" class=\"goto pointer\" rel=\"".$row["planetid"]."\">".$img."</a>");
		}
		else
		{
			Core::getTPL()->assign("moon", "");
			Core::getTPL()->assign("moonImage", "");
		}

		// $off_result = sqlSelectRow("user", array("of_points", "of_level"), "", "userid = ".sqlUser());
		Core::getTPL()->assign("offPoints", fNumber(NS::getUser()->get("of_points"))); // $off_row["of_points"]);
		Core::getTPL()->assign("offLevel", fNumber(NS::getUser()->get("of_level")));

		$need = 200;
		$need_points = (NS::getUser()->get("of_level") < 1) ? $need/2 : fNumber(pow(1.5, NS::getUser()->get("of_level") - 1) * $need);
		Core::getTPL()->assign("need_points", $need_points);

		// Planet list
		$order = getPlanetOrder(NS::getUser()->get("planetorder"));
		// $events = NS::getEH()->getAllPlanetsBuildingEvents();
		$result = sqlSelect("planet p", array("p.planetid", "p.planetname", "p.picture", "p.ismoon"), "LEFT JOIN ".PREFIX."galaxy g ON (g.planetid = p.planetid)", "p.userid = ".sqlUser()." AND p.planetid != ".sqlPlanet()." AND p.ismoon = '0' ORDER BY $order");
		$i = 0;
		$loop = array();
		while($row = sqlFetch($result))
		{
			$i++;
			$cur_planet_event = NS::getEH()->getFirstPlanetBuildingEvent($row["planetid"]);
			$activity = $cur_planet_event ? $cur_planet_event["data"]["buildingname"] : "";
			$loop[$row["planetid"]]["planetname"] = $row["planetname"];
			$loop[$row["planetid"]]["counter"] = $i;
			$img = Image::getImage("planets/".$row["picture"].Core::getConfig()->get("PLANET_IMG_EXT"), $row["planetname"], $small_planet_img_width, $small_planet_img_width);
			$loop[$row["planetid"]]["planetimage"] = "<a title=\"".$row["planetname"]."\" class=\"goto pointer\" rel=\"".$row["planetid"]."\">".$img."</a>";
			if(empty($activity)) { $activity = "PLANET_FREE"; }
			$loop[$row["planetid"]]["activity"] = Core::getLanguage()->getItem($activity);
			$loop[$row["planetid"]]["total"] = Core::getDB()->num_rows($result);
			$mod = $i %2;
		}
		sqlEnd($result);
		// Hook::event("MAIN_PLANET_OVERVIEW", array(&$loop));
		Core::getTPL()->addLoop("planetSelection", $loop);

		Core::getTPL()->display("main");
		return $this;
	}

	/**
	* Shows form for planet options.
	*
	* @return Main
	*/
	protected function planetOptions()
	{
		if( !NS::getPlanet()->getPlanetId() )
		{
			doHeaderRedirection("game.php/Main", false);
		}
        if(!isNameCharValid(NS::getPlanet()->getData('planetname'))){
            Logger::addMessage("INVALID_PLANET_NAME");
        }
		Core::getTPL()->assign("curplanet", getCoordLink(false, false, false, false, NS::getPlanet()->getPlanetId(), true));
		Core::getTPL()->assign("homeplanet", getCoordLink(false, false, false, "select", NS::getUser()->get("hp"), true));
		Core::getTPL()->assign("position", NS::getPlanet()->getCoords());
		Core::getTPL()->assign("planetName", NS::getPlanet()->getData("planetname"));
		Core::getTPL()->display("planetOptions");
		return $this->quit();
	}

	/**
	* Shows form for planet options.
	*
	* @param array		_POST
	*
	* @return Main
	*/
	protected function changePlanetOptions($planetname, $abandon, $password, $ishomeplanet, $leave = '', $editplanetid = 0)
	{
		$planetname = trim($planetname);
		// Hook::event("SAVE_PLANET_OPTIONS", array(&$planetname, &$abandon));
		if( $abandon == 1 )
		{
			if(NS::isPlanetUnderAttack())
			{
				Logger::dieMessage('PLANET_UNDER_ATTACK');
			}

			if( NS::getUser()->get("hp") == NS::getUser()->get("curplanet") )
			{
				Logger::addMessage("CANNOT_DELETE_HOME_PLANET");
				$this->planetOptions();
				return $this;
			}
			if( NS::getEH()->getPlanetFleetEvents()
				|| NS::getUser()->get("curplanet") != $editplanetid
				|| NS::getUser()->get("curplanet") != NS::getPlanet()->getPlanetId()
				)
			{
				Logger::addMessage("CANNOT_DELETE_PLANET");
				$this->planetOptions();
				return $this;
			}
//			// Do not delete till April;

//			if( !defined('SN') )
//			{
//				$row = sqlSelectRow("password", "password", "", "userid = ".sqlUser());
//				if(!Str::compare( $row["password"], md5($password)) )
//				{
//					Logger::addMessage("WRONG_PASSWORD");
//					$this->planetOptions();
//				}
//			}
//			else
//			{
				$leave = trim($leave);
				if( empty($leave) || !Str::compare( Core::getLang()->getItem('LEAVE').NS::getPlanet()->getPlanetId(), $leave, true ) )
				{
					Logger::addMessage("WRONG_LEAVE_PHRASE");
					$this->planetOptions();
					return $this;
				}
//			}
			// Update By Pk
			sqlUpdate( "user", array( "curplanet" => NS::getUser()->get("hp") ),
				"userid = ".sqlUser()
					// . ' ORDER BY userid'
			);
			// NS::getPlanet()->removeEvents();
			NS::deletePlanet(NS::getPlanet()->getPlanetId(), NS::getUser()->get("userid"), NS::getPlanet()->getData("ismoon"));
			// NS::getUser()->rebuild();
			doHeaderRedirection("game.php/Main", false);
		}

		if($ishomeplanet
				&& !NS::getPlanet()->getData("ismoon")
				&& NS::getPlanet()->getPlanetId() != NS::getUser()->get("hp")
				&& !NS::getPlanet()->getData("destroy_eventid"))
		{
			NS::getUser()->set("hp", NS::getPlanet()->getPlanetId());
		}

        if(isNameCharValid($planetname)){
            if(checkCharacters($planetname)){
                Core::getQuery()->update(
                    "planet",
                    "planetname",
                    $planetname,
                    "planetid = ".sqlPlanet()
                        . ' ORDER BY planetid'
                );
                doHeaderRedirection("game.php/Main", false);
            }else{
                Logger::addMessage("INVALID_PLANET_NAME");
                $this->planetOptions();
            }
        }else{
           doHeaderRedirection("game.php/Main", false);
        }
		return $this;
	}

	/**
	* Parses an event.
	*
	* @param array		Event data
	*
	* @return array	Parsed event data
	*/
	protected function parseEvent($f, $fleet_list_params = array())
	{
		if($f["mode"] == EVENT_ALLIANCE_ATTACK_ADDITIONAL && $f["user"] != NS::getUser()->get("userid"))
		{
			return false; // Hide foreign formations
		}
		$is_return_from_debris = isset($f["data"]["oldmode"]) && $f["data"]["oldmode"] == EVENT_RECYCLING;

		$event["time_r"] = $f["time"] - time();
		$event["time"] = getTimeTerm($event["time_r"]);
		$event["eventid"] = $f["eventid"];

		$event["event_start"] = $f["start"];
		$event["event_end"] = $f["time"];
		// $event["event_timeleft"] = max(0, $f["time"] - time());
		$event["event_percent_timeout"] = $f["time"] > $f["start"] ? ceil(($f["time"] - $f["start"]) * 1000 / 100.0) : 1000*10;
		$event["event_pb_value"] = $f["time"] > $f["start"] ? max(1, floor((time() - $f["start"]) * 100 / ($f["time"] - $f["start"]))) : 0;

		$ships = "";
		$is_event_owner = $f["user"] == NS::getUser()->get("userid") || $f["startuserid"] == NS::getUser()->get("userid");
		$mission_mode = $f["mode"] == EVENT_RETURN && isset($f["data"]["pathstack"]) && count($f["data"]["pathstack"]) > 0 ? EVENT_HALT : $f["mode"];
		Core::getLanguage()->assign("rockets", $f["data"]["rockets"]);
		Core::getLanguage()->assign("planet", !$is_return_from_debris ? $f["planetname"] : Core::getLanguage()->getItem("DEBRIS"));
		Core::getLanguage()->assign("coords", $f["galaxy"] ? getCoordLink($f["galaxy"], $f["system"], $f["position"], false, $f["planetid"]) : (!$is_return_from_debris ? getCoordLink($f["data"]["sgalaxy"], $f["data"]["ssystem"], $f["data"]["sposition"], false, $f["planetid"]) : getCoordLink($f["data"]["galaxy"], $f["data"]["system"], $f["data"]["position"], false, $f["planetid"])));
		Core::getLanguage()->assign("target", $f["mode"] != EVENT_RECYCLING ? $f["destplanet"] : Core::getLanguage()->getItem("DEBRIS"));
		Core::getLanguage()->assign("targetcoords", $f["galaxy2"] ? getCoordLink($f["galaxy2"], $f["system2"], $f["position2"], false, $f["destination"]) : getCoordLink($f["data"]["galaxy"], $f["data"]["system"], $f["data"]["position"], false, $f["destination"]));
		Core::getLanguage()->assign("metal", $f["data"]["metal"] ? fNumber($f["data"]["metal"]) : 0);
		Core::getLanguage()->assign("silicon", $f["data"]["silicon"] ? fNumber($f["data"]["silicon"]) : 0);
		$hydrogen = $f["data"]["hydrogen"] + (isset($f["data"]["ret_consumption"]) ? $f["data"]["ret_consumption"] : 0);
		Core::getLanguage()->assign("hydrogen", $hydrogen ? fNumber($hydrogen) : 0);
		Core::getLanguage()->assign("username", $f["username"]);
		Core::getLanguage()->assign("message", Link::get("game.php/MSG/Write/Receiver:".$f["username"], Image::getImage("pm.gif", Core::getLanguage()->getItem("WRITE_MESSAGE"))));
		Core::getLanguage()->assign("mission", $mission_mode == EVENT_RETURN ? NS::getMissionName($f["data"]["oldmode"], $is_event_owner) : NS::getMissionName($mission_mode, $is_event_owner));

		Core::getLanguage()->assign("fleet", getUnitListStr($f["data"]["ships"],
			$is_event_owner ? $fleet_list_params : array_merge(array("no_damaged" => true), $fleet_list_params)
			));
		$art_id = 0;
		if( isset($f["data"]["ships"]) )
		{
			foreach( $f["data"]["ships"] as $type => $values )
			{
				if ( isset($values['art_ids']) && !empty($values['art_ids']) )
				{
					$art_id = key( $values['art_ids'] );
				}
			}
		}
		if( !empty($art_id) )
		{
			$art_data = getArtefactNameAndPosition($art_id);
			Core::getLanguage()->assign("artefact", $art_data['name']);
		}

		$event["class"] = getFleetMessageClass($mission_mode, $is_event_owner);

		if(in_array($f["mode"], array(EVENT_ATTACK_ALLIANCE,
										EVENT_ATTACK_ALLIANCE_DESTROY_BUILDING,
										EVENT_ATTACK_ALLIANCE_DESTROY_MOON
										)))
		{
			if($is_event_owner)
			{
				$event["message"] = Core::getLanguage()->getItem("FLEET_MESSAGE_OWN");
				$event["retreat_eventid"] = $f["eventid"];
			}
			else
			{
				$event["message"] = Core::getLanguage()->getItem("FLEET_MESSAGE_OTHER");
				$fleet_list_params = array_merge(array("no_damaged" => true), $fleet_list_params);
				if(isAdmin(null, true))
				{
					$event["retreat_eventid"] = $f["eventid"];
				}
			}
			$allyFleets = NS::getEH()->getFormationAllyFleets($f["eventid"]); // $f["destination"], $f["time"]);
			foreach($allyFleets as $af)
			{
				$coords = getCoordLink($af["data"]["sgalaxy"], $af["data"]["ssystem"], $af["data"]["sposition"], false, $af["planetid"]);
				$msg = Core::getLanguage()->getItem("FLEET_MESSAGE_FORMATION");
				$msg = sprintf($msg, $af["username"], $af["planetname"], $coords, getUnitListStr($af["data"]["ships"], $fleet_list_params));
				$event["message"] .= $msg;
			}
		}
		else if($f["mode"] == EVENT_ALLIANCE_ATTACK_ADDITIONAL)
		{
			$mainFleet = NS::getEH()->getMainFormationFleet($f["data"]["alliance_attack"]["eventid"]);
			if($mainFleet["user"] == NS::getUser()->get("userid"))
			{
				return false;
			}
			$event["message"] = Core::getLanguage()->getItem("FLEET_MESSAGE_OWN");
			$allyFleets = NS::getEH()->getFormationAllyFleets($f["data"]["alliance_attack"]["eventid"]); // $f["destination"], $f["time"]);
			$allyFleets[] = $mainFleet;
			foreach($allyFleets as $af)
			{
				$coords = getCoordLink($af["data"]["sgalaxy"], $af["data"]["ssystem"], $af["data"]["sposition"], false, $af["planetid"]);
				$msg = Core::getLanguage()->getItem("FLEET_MESSAGE_FORMATION");
				$msg = sprintf($msg, $af["username"], $af["planetname"], $coords, getUnitListStr($af["data"]["ships"]));
				$event["message"] .= $msg;
			}
			if(time() > $event["event_end"] + 60*60*12){
				$parent_event_exist = sqlSelectField("events", "count(*)", null, "eventid=".sqlVal($f["parent_eventid"]));
				if(!$parent_event_exist){
					$event["retreat_eventid"] = $f["eventid"];
				}
				/* echo "<pre>"; 
				print_r(array("parent_exist" => $parent_event_exist)); 
				print_r($f); 
				echo "</pre>";
				throw new Exception("sfsds"); */
			}
		}
		else if($f["mode"] == EVENT_HOLDING)
		{
			$event["message"] = Core::getLanguage()->getItem($is_event_owner ? "FLEET_MESSAGE_HOLDING_1" : "FLEET_MESSAGE_HOLDING_2");
			$event["retreat_eventid"] = $f["eventid"];
		}
		else if($f["mode"] == EVENT_ALIEN_HOLDING)
		{
			$event["message"] = Core::getLanguage()->getItem("FLEET_MESSAGE_ALIEN_HOLDING");
		}
		else if($f["mode"] == EVENT_ALIEN_FLY_UNKNOWN || $f["mode"] == EVENT_ALIEN_GRAB_CREDIT
				|| $f["mode"] == EVENT_ALIEN_ATTACK || $f["mode"] == EVENT_ALIEN_ATTACK_CUSTOM || $f["mode"] == EVENT_ALIEN_HALT)
		{
			$event["message"] = Core::getLanguage()->getItem($f["startuserid"] != NS::getUser()->get("userid") ? "FLEET_MESSAGE_ALIEN" : "FLEET_MESSAGE_ALIEN_OTHER");
		}
		else if($f["mode"] == EVENT_RETURN && $is_event_owner)
		{
			$event["message"] = Core::getLanguage()->getItem("FLEET_MESSAGE_RETURN");
		}
		else if($f["mode"] == EVENT_ROCKET_ATTACK && $is_event_owner)
		{
			$event["message"] = Core::getLanguage()->getItem("FLEET_MESSAGE_ROCKET_ATTACK");
		}
		else if($f["mode"] == EVENT_ROCKET_ATTACK)
		{
			$event["message"] = Core::getLanguage()->getItem("FLEET_MESSAGE_ROCKET_ATTACK_FOREIGN");
		}
		else if($f["mode"] == EVENT_DELIVERY_UNITS)
		{
			$event["message"] = Core::getLanguage()->getItem(
				$is_event_owner ? "FLEET_MESSAGE_DELIVERY_UNITS" : "FLEET_MESSAGE_DELIVERY_UNITS_SELLER");
		}
		else if($f["mode"] == EVENT_DELIVERY_RESOURSES)
		{
			$event["message"] = Core::getLanguage()->getItem(
				$is_event_owner ? "FLEET_MESSAGE_OWN" : "FLEET_MESSAGE_OTHER_RES");
		}
		else if($f["mode"] == EVENT_DELIVERY_ARTEFACTS)
		{
			$event["message"] = Core::getLanguage()->getItem(
				$is_event_owner ? "FLEET_MESSAGE_DELIVERY_ARTEFACTS" : "FLEET_MESSAGE_DELIVERY_ARTEFACTS_SELLER");
		}
		else if($f["mode"] == EVENT_TEMP_PLANET_DISAPEAR)
		{
			$coordinates = Link::get(
				"game.php/go:Mission/g:".$f['data']["galaxy"]."/s:".$f['data']["system"]."/p:".$f['data']["position"],
				"[{$f['data']['galaxy']}:{$f['data']['system']}:{$f['data']['position']}]"
		);
				$event["message"] = Core::getLanguage()->getItemWith("PLANET_ANIGILATION", array('planet' => $coordinates));
		}
		else if($f["mode"] == EVENT_COLONIZE_NEW_USER_PLANET) // EVENT_COLONIZE_RANDOM_PLANET && !$f["planetid"])
		{
			$event["message"] = Core::getLanguage()->getItem("FLEET_MESSAGE_COLONIZE_NEW_USER_PLANET");
		}
		else if($is_event_owner)
		{
			$event["message"] = Core::getLanguage()->getItem("FLEET_MESSAGE_OWN");
			if($f["mode"] != EVENT_STARGATE_JUMP)
			{
				$event["retreat_eventid"] = $f["eventid"];
			}
		}
		else
		{
			$event["message"] = Core::getLanguage()->getItem("FLEET_MESSAGE_OTHER");
			if($f["mode"] == EVENT_HALT || $f["mode"] == EVENT_TRANSPORT
				|| (isAdmin(null, true) && in_array($f["mode"], array(EVENT_ATTACK_SINGLE, EVENT_ATTACK_DESTROY_BUILDING, EVENT_ATTACK_DESTROY_MOON))))
			{
				$event["retreat_eventid"] = $f["eventid"];
			}
		}
		if($f["mode"] == EVENT_HOLDING
			// && $f["destination"] == NS::getPlanet()->getPlanetId()
			// && (NS::getUser()->get("userid") == 1 || NS::getUser()->get("userid") == 3)
			)
		{
			$event["control_eventid"] = $f["eventid"];
		}
		return $event;
	}
}
?>