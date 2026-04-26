<?php
/**
* Unsorted functions.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

/**
 * Gets planet or moon coordinates ant it's name
 * @param int $id Planet or Moon id
 */
function getCoordsAndPlanetname( $id )
{
	return sqlSelectRow(
		'planet p',
		array(
			'p.planetname',
			'IFNULL(g.galaxy, gm.galaxy) as galaxy',
			'IFNULL(g.system, gm.system) as system',
			'IFNULL(g.position, gm.position) as position',
		),
		'LEFT JOIN '.PREFIX.'galaxy g ON g.planetid = p.planetid'
			. ' LEFT JOIN '.PREFIX.'galaxy gm ON gm.moonid = p.planetid',
		'p.planetid = '.sqlVal( $id )
	);
}

/**
* Executes a simple formula.
*
* @param string	Formular to execute
* @param integer	Basic costs
* @param integer	Level
*
* @return integer	Result
*/
function parseChargeFormula($formula, $basic, $level)
{
	// Hook::event("PARSE_SIMPLE_FORMULA", array(&$formula, &$basic, &$level));
	$formula = Str::replace("{level}", $level, $formula);
	$formula = Str::replace("{basic}", $basic, $formula);
	$formula = trim($formula);
	if(!$formula)
	{
		$formula = $basic;
	}
	$result = 0;
	eval("\$result = ".$formula.";");
	return max(0, round($result));
}

/**
 * Function that updates global user states
 * @param int $user_id User id
 * @param int $state_name What state to update
 * @param boolean $recalc Whether to racalculate achievements
 */
function updateUserState($user_id, $state_name, $recalc = false)
{
	if( empty($user_id) || $user_id == 0 )
	{
		return;
	}
	if( empty($state_name) || $state_name == 0 )
	{
		return;
	}
	$state_field = 'unknown';
	switch ($state_name)
	{
		case STATE_ASSAULT_SIMULATION:
			$state_field = 'simulated_assault';break;
		case STATE_RES_UPDATE_EXCHANGE:
			$state_field = 'exchanged_ress';break;
	}
	$state = UserStates_YII::model()->findByPk($user_id);
	if( $state )
	{
		$val = $state->getAttribute( $state_field );
		if( $val != 0 )
		{
			$state->setAttribute($state_field, ($val + 1) );
		}
		else
		{
			$recalc = true;
			$state->setAttribute($state_field, 1);
		}
	}
	else
	{
		$recalc = true;
		$state = new UserStates_YII();
		$state->setAttributes(
			array(
				'user_id' 	=> $user_id,
				$state_field=> 1,
			),
			false
		);
	}
	$state->save(false);
	if( $recalc )
	{
		$user = User_YII::model()->findByPk($user_id);
		Achievements::processAchievements($user_id, $user->curplanet);
	}
}

/**
* Generates a time term of given seconds.
*
* @param integer	Seconds
*
* @return string	Time term
*/
function getTimeTerm($secs)
{
	$secs = abs($secs);
	// Hook::event("FORMAT_TIME_START", array(&$secs));
	$days = floor($secs/60/60/24);
	$hours = floor(($secs/60/60)%24);
	$mins = floor(($secs/60)%60);
	$seconds = fmod($secs, 60);
	// Hook::event("FORMAT_TIME_END", array(&$days, &$days, &$mins, &$seconds));
	$parsed = sprintf("%dd %02d:%02d:%06.3f", $days, $hours, $mins, $seconds);
	if($days == 0) { $parsed = Str::substring($parsed, 3); }
	if(fmod($seconds, 1) == 0) { $parsed = Str::substring($parsed, 0, -4); }
	return $parsed;
}

/**
* Prettify number: 1000 => 1.000
*
* @param integer	The number being formatted
* @param integer	Sets the number of decimal points
*
* @return string	Number
*/
function fNumber($number, $decimals = 0)
{
	return number_format($number, $decimals, Core::getLang()->getItem("DECIMAL_POINT"), Core::getLang()->getItem("THOUSANDS_SEPERATOR"));
}

function pNumber($number, $maxDecimals = 2)
{
	if($number < 10) $decimals = min(2, $maxDecimals);
	else if($number < 100) $decimals = min(1, $maxDecimals);
	else $decimals = 0;
	return fNumber($number, $decimals);
}

/**
* Checks a string for valid characters.
* http://ru2.php.net/manual/en/regexp.reference.unicode.php
*
* @param string
*
* @return boolean
*/
function checkCharacters($string)
{
    $res = preg_match("#^[a-z\d_\-\s\.\pL]"
        . "{"
        . max(1, (int)Core::getOptions()->get("MIN_USER_CHARS")) . ","
        . min(100, (int)Core::getOptions()->get("MAX_USER_CHARS")) . "}"
        // . "+"
        . "$#iu", $string);
	// debug_var(array($string, $res, Core::getOptions()->get("MIN_USER_CHARS"), Core::getOptions()->get("MAX_USER_CHARS")), "[check chars]");
    return $res;
}


/**
* Checks a string for valid email address.
*
* @param string
*
* @return boolean
*/
function checkEmail($mail)
{
	if(preg_match("#^[a-zA-Z0-9-]+([._a-zA-Z0-9.-]+)*@[a-zA-Z0-9.-]+\.([a-zA-Z]{2,4})$#is", $mail))
	{
		return true;
	}
	return false;
}


/**
* Generates a coordinate link
*
* @param integer	Galaxy
* @param integer	System
* @param integer	Position
* @param boolean	Replace session with wildcard
*
* @return string	Link code
*/
function getCoordLink($galaxy, $system, $position, $sidWildcard = false, $planetid = false, $addname = false)
{
	$planetname = "";
	$planet_post_str = "";
	if(is_numeric($planetid))
	{
		static $planet_data = array();

		$planetid = (int)$planetid;
		if(!isset($planet_data[$planetid]))
		{
			if(NS::getPlanet() && NS::getPlanet()->getPlanetid() == $planetid)
			{
				$planet_data[$planetid] = array(
					"ismoon" => NS::getPlanet()->getData("ismoon"),
					"galaxy" => NS::getPlanet()->getData("galaxy"),
					"system" => NS::getPlanet()->getData("system"),
					"position" => NS::getPlanet()->getData("position"),
					"planetname" => NS::getPlanet()->getData("planetname"),
					);
			}
			else
			{
				$row = sqlSelectRow("planet", array("planetname", "ismoon"), "", "planetid=".sqlVal($planetid));
				$planet_data[$planetid] = array(
					"ismoon" => $row["ismoon"],
					"galaxy" => $galaxy,
					"system" => $system,
					"position" => $position,
					"planetname" => $row["planetname"],
					);

				if(!$galaxy)
				{
					$row = sqlSelectRow("galaxy", "*", "",
						($planet_data[$planetid]["ismoon"] ? "moonid=" : "planetid=").sqlVal($planetid));
					$planet_data[$planetid]["galaxy"] = $row["galaxy"];
					$planet_data[$planetid]["system"] = $row["system"];
					$planet_data[$planetid]["position"] = $row["position"];
				}
			}
		}
		$planet_post_str = $planet_data[$planetid]["ismoon"] ? Core::getLanguage()->getItem("LOON_POST") : "";
		$galaxy = $planet_data[$planetid]["galaxy"];
		$system = $planet_data[$planetid]["system"];
		$position = $planet_data[$planetid]["position"];
		$planetname = $planet_data[$planetid]["planetname"];
	}
	if($sidWildcard === "select" && is_numeric($planetid))
	{
		return ($addname ? "$planetname " : "")."<a href='#' onclick='return false' class='goto' lang='$planetid'>[".$galaxy.":".$system.":".$position."]".$planet_post_str."</a>";
	}

	return ($addname ? "$planetname " : "").Link::get("game.php/go:Galaxy/galaxy:".$galaxy."/system:".$system, "[".$galaxy.":".$system.":".$position."]".$planet_post_str,
		"", // $title = "",
		"", // $cssClass = "",
		"", // $attachment = "",
		"", // $appendSession = false,
		true, // $rewrite = true,
		true, // $refdir = true,
		false, // $appendSNparams = false - not used at the moment
		'', // $sidWildcard ? 'ODNOKLASSNIKI' : '' // $sn_name = 'ODNOKLASSNIKI'
		!$sidWildcard
	);
}


/**
* Returns the CSS class of the fleet mode.
*
* @param integer	EH-Mode
*
* @return string	CSS class
*/
function getFleetMessageClass($id, $is_owner = false)
{
	switch($id)
	{
	case EVENT_POSITION:
		$return = "STATIONATE";
		break;

	case EVENT_TRANSPORT:
	case EVENT_DELIVERY_UNITS:
	case EVENT_DELIVERY_RESOURSES:
	case EVENT_DELIVERY_ARTEFACTS:
		$return = "TRANSPORT";
		break;

	case EVENT_STARGATE_TRANSPORT:
	case EVENT_STARGATE_JUMP:
	case EVENT_TELEPORT_PLANET:
		$return = "STARGATE_TRANSPORT";
		break;

	case EVENT_COLONIZE:
	case EVENT_COLONIZE_RANDOM_PLANET:
	case EVENT_COLONIZE_NEW_USER_PLANET:
		$return = "COLONIZE";
		break;

	case EVENT_RECYCLING:
		$return = "RECYCLING";
		break;

	case EVENT_ATTACK_SINGLE:
	case EVENT_ATTACK_DESTROY_BUILDING:
	case EVENT_ATTACK_DESTROY_MOON:
	case EVENT_ALIEN_ATTACK:
	case EVENT_ALIEN_ATTACK_CUSTOM:
	case EVENT_ALIEN_FLY_UNKNOWN:
	case EVENT_ALIEN_GRAB_CREDIT:
		$return = $is_owner ? "ATTACK" : "ATTACKED";
		break;

	case EVENT_SPY:
		$return = "SPY";
		break;

	case EVENT_ATTACK_ALLIANCE_DESTROY_BUILDING:
	case EVENT_ATTACK_ALLIANCE_DESTROY_MOON:
	case EVENT_ATTACK_ALLIANCE:
	case EVENT_ALLIANCE_ATTACK_ADDITIONAL:
		$return = $is_owner ? "ALLIANCE_ATTACK" : "ALLIANCE_ATTACKED";
		break;

	case EVENT_HALT:
	case EVENT_HOLDING:
	case EVENT_ALIEN_HALT:
	case EVENT_ALIEN_HOLDING:
		$return = "HALT";
		break;

	case EVENT_MOON_DESTRUCTION:
		$return = "MOON_DESTRUCTION";
		break;

	case EVENT_EXPEDITION:
		$return = "EXPEDITION";
		break;

	case EVENT_ROCKET_ATTACK:
		$return = "ROCKET_ATTACK";
		break;

	case EVENT_RETURN:
		$return = "RETURN_FLY";
		break;

	default:
		$return = "UNKOWN";
		break;
	}
	// Hook::event("GET_CONSTRUCTION_TIME", array(&$return, $id));
	return strtolower($return);
}


/**
* Generates HTML code for a select option.
*
* @param string	Option value
* @param string	Option name
* @param boolean	Set selected tag
* @param string	Additional CSS class
*
* @return string	HTML
*/
function createOption($value, $option, $selected, $class = "")
{
	$class = ($class != "") ? " class=\"".$class."\"" : "";
	$coption = "<option value=\"".$value."\"".$class;
	if($selected == 1)
	{
		$coption .= " selected=\"selected\"";
	}
	$coption .= ">".$option."</option>";
	return $coption;
}


/**
* Deletes all data of an alliance.
*
* @param integer	Alliance id to delete
*
* @return void
*/
function deleteAlliance($aid)
{
	$result = sqlSelect("user2ally u2a", array("u2a.userid", "a.tag"), "LEFT JOIN ".PREFIX."alliance a ON (a.aid = u2a.aid)", "u2a.aid = ".sqlVal($aid));
	while($row = sqlFetch($result))
	{
		new AutoMsg(MSG_ALLY_ABANDONED, $row["userid"], time(), $row);
	}
	sqlEnd($result);
	Core::getQuery()->delete("alliance", "aid = ".sqlVal($aid));
	// Hook::event("DELETE_ALLIANCE", array($aid));
	return;
}

/**
* Returns the planet order.
*
* @param integer	Order mode
*
* @return string	ORDER BY term for SQL query
*/
function getPlanetOrder($mode)
{
	switch($mode)
	{
	case 1:
		$order = "p.planetid ASC";
		break;
	case 2:
		$order = "p.planetname ASC";
		break;
	case 3:
		$order = "g.galaxy ASC, g.system ASC, g.position ASC";
		break;
	default:
		$order = "p.planetid ASC";
		break;
	}
	// Hook::event("GET_PLANET_ORDER", array($mode, &$order));
	return $order;
}


/**
* Checks if user is newbie protected.
*
* @param integer	Points of requesting user
* @param integer	Points of target user
*
* @return integer	0: no protection, 1: weak, 2: strong
*/
function isNewbieProtected($committer, $target)
{
	// Hook::event("NEWBIE_PROTECTION", array(&$committer, &$target));

	if(!NEWBIE_PROTECTION_ENABLED)
	{
		return 0;
	}

	if($committer <= 1 || $target <= 1)
	{
		return $committer >= $target ? 1 : 2;
	}

	$active_points = min($committer, $target);
	foreach(array(
		array("points" => NEWBIE_PROTECTION_1_POINTS, "percent" => NEWBIE_PROTECTION_1_PERCENT),
		array("points" => NEWBIE_PROTECTION_2_POINTS, "percent" => NEWBIE_PROTECTION_2_PERCENT),
		array("points" => NEWBIE_PROTECTION_3_POINTS, "percent" => NEWBIE_PROTECTION_3_PERCENT),
		array("points" => 99999999999999999999, "percent" => NEWBIE_PROTECTION_MAX_POINTS_PERCENT),
		) as $row)
	{
		if($active_points < $row["points"])
		{
			$newbieScale = $row["percent"] / 100;
			if($committer * $newbieScale > $target)
			{
				return 1;
			}
			if($target * $newbieScale > $committer)
			{
				return 2;
			}
			break;
		}
	}
	return 0;
}


/**
* Time that a rocket needs to reach is destination.
*
* @param integer	Difference between galaxies
*
* @return integer	Time in seconds
*/
function getRocketFlightDuration($diff)
{
	return ceil((30 + 60 * abs($diff)) / ROCKET_SPEED_FACTOR);
}

/**
* Replaces special characters.
*
* @param string	String to convert
*
* @return string	Converted string
*/
function convertSpecialChars($string)
{
	$string = htmlentities($string, ENT_NOQUOTES, "UTF-8", false);
	$string = htmlspecialchars_decode($string, ENT_NOQUOTES);
	$string = nl2br($string);
	return $string;
}


/**
* Checks a string for valid url.
*
* @param string	String to check
*
* @return boolean
*/
function isValidURL($string)
{
	$pattern = "#^(http|https|mailto|news|irc)://([\w\d][\w\d\$\_\.\+\!\*\(\)\,\;\/\?\:\(at)\&\~\=\-]+)(:(\d +))?(/[^ ]*)?$#is";
	if(preg_match($pattern, $string)) { return true; }
	return false;
}


/**
* Checks a string for valid image url.
*
* @param string	String to check
*
* @return boolean
*/
function isValidImageURL($string)
{
	$pattern = "#^http://([\w\d][\w\d\$\_\.\+\!\*\(\)\,\;\/\?\:\(at)\&\~\=\-]+)(:(\d +))?(/[^ ]*)?\.(jpg|jpeg|gif|png)$#is";
	return preg_match($pattern, $string);
}


/**
* Sets the production factor of an user (e.g. for vacation mode).
*
* @param integer	The user id
* @param integer	The new production factor
*
* @return void
*/
function setProdOfUser($userid, $prod)
{
	$sql = "UPDATE `".PREFIX."building2planet` b2p, `".PREFIX."planet` p SET b2p.`prod_factor` = ".sqlVal($prod)." WHERE b2p.`planetid` = p.`planetid` AND p.`userid` = ".sqlVal($userid);
	sqlQuery($sql);
	Core::getQuery()->update("planet", array("solar_satellite_prod"), array($prod), "userid = ".sqlVal($userid) . ' ORDER BY planetid');
	return;
}

function loadDisplaySpaceEvents()
{
	Core::getLanguage()->load(array("info", "buildings"));

	$list_size = 15;

	$buildings[] = array();
	$space_events = array();
	$joins .= "LEFT JOIN ".PREFIX."galaxy g ON (g.planetid = e.planetid) ";
	$select = array("e.*", "g.galaxy", "g.system", "g.position");
	$result = sqlSelect("events e", $select, $joins,
		"processed=".EVENT_PROCESSED_WAIT
		. " AND mode NOT IN (".sqlArray(EVENT_ARTEFACT_EXPIRE, EVENT_ARTEFACT_DISAPPEAR, EVENT_ARTEFACT_DELAY).")",
		"eventid desc", $list_size * 3);
	while($row = sqlFetch($result))
	{
		$row["data"] = unserialize($row["data"]);
		$row["buildingid"] = $row["data"]["buildingid"];
		$row["coords"] = "[".$row["galaxy"].":".($row["system"] >= 10 ? (int)($row["system"]/10) : "")."Y:ZZ]";

		$timeleft = $row["time"] - time();
		$continue = Core::getLanguage()->getItem("DONE");
		$elementId = "bCountDown" . $row["eventid"];
		$js = "<script type=\"text/javascript\">
			//<![CDATA[
			$(function () {
				$('#".$elementId."').countdown({until: ".$timeleft.", compact: true, description: '', onExpiry: function() {
					$('#".$elementId."').html('".$continue."');
				}});
		});
		//]]>
		</script>
			<span id=\"".$elementId."\">".getTimeTerm($timeleft)."</span>";
		$row["time_countdown"] = $js;

		$mode_lang_id = "";
		switch($row["mode"])
		{
		case 1: // Construction
			$mode_lang_id = "EVENT_MODE_CONSTRUCTION";
			break;
		case 2: // Demolish
			$mode_lang_id = "EVENT_MODE_DEMOLISH";
			break;
		case 3: // Research
			$mode_lang_id = "EVENT_MODE_RESEARCH";
			break;
		case 4: // Fleet
			$mode_lang_id = "EVENT_MODE_FLEET";
			break;
		case 5: // Defense
			$mode_lang_id = "EVENT_MODE_DEFENSE";
			break;

			//### Fleet missions ###//
		case 6: // Position
			$mode_lang_id = "EVENT_MODE_POSITION";
			break;
		case 7: // Transport
			$mode_lang_id = "EVENT_MODE_TRANSPORT";
			break;
		case 8: // Colonize
			$mode_lang_id = "EVENT_MODE_COLONIZE";
			break;
		case 9: // Recycling
			$mode_lang_id = "EVENT_MODE_RECYCLING";
			break;
		case 10: // Attack
			$mode_lang_id = "EVENT_MODE_ATTACK";
			break;
		case 11: // Spy
			$mode_lang_id = "EVENT_MODE_SPY";
			break;
		case 12: // Alliance attack
			$mode_lang_id = "EVENT_MODE_ALLIANCE_ATTACK";
			break;
		case 13: // Halt
			$mode_lang_id = "EVENT_MODE_HALT";
			break;
		case 14: // Moon destruction
			$mode_lang_id = "EVENT_MODE_MOON_DESTRUCTION";
			break;
		case 15: // Expedition
			$mode_lang_id = "EVENT_MODE_EXPEDITION";
			break;
		case 16: // Rocket attack
			$mode_lang_id = "EVENT_MODE_ATTACK";
			break;
		case 17: // Holding
			$mode_lang_id = "EVENT_MODE_HOLDING";
			break;
		case 18: // Serves as referer to alliance attack
			$mode_lang_id = "EVENT_MODE_ALLIANCE_ATTACK";
			break;
		case 20: // Return
			$mode_lang_id = "EVENT_MODE_FLEET_RETURNED";
			break;
		}
		$row["mode_name"] = Core::getLanguage()->getItem($mode_lang_id);

		$buildings[$row["buildingid"]] = array();
		$space_events[$row["eventid"]] = $row;
	}
	sqlEnd($result);

	$result = sqlSelect("construction", array("buildingid", "name"), "",
		"buildingid in (".sqlArray(array_keys($buildings)).")");
	while($row = sqlFetch($result))
	{
		$id = $row["buildingid"];
		$buildings[$id]["name"] = Core::getLanguage()->getItem($row["name"]);
		$buildings[$id]["description"] = Core::getLanguage()->getItem($row["name"]."_DESC");
		$buildings[$id]["image"] = Image::getImage(getUnitImage($row["name"]), $buildings[$id]["name"], 90);
	}
	sqlEnd($result);

	$prev_event = false;
	$events_number = count($space_events);
	foreach($space_events as $key => & $event)
	{
		if($events_number > $list_size && $prev_event
			&& $prev_event["buildingid"] == $event["buildingid"])
		{
			unset($space_events[$key]);
			$events_number--;
		}
		else
		{
			$event["building_name"] = $buildings[$event["buildingid"]]["name"];
			$event["building_desc"] = $buildings[$event["buildingid"]]["description"];
			$event["building_image"] = $buildings[$event["buildingid"]]["image"];

			$prev_event = & $event;
		}
	}

	return $space_events;
}

/**
 * Checks if Second user is in trade union relation with first.
 * @param int $id_1 First User id
 * @param int $id_2 Second User id
 */
function isTradeUnion( $id_1, $id_2 )
{
	if( !$id_1 || !$id_2 )
	{
		return false;
	}

	$ally_1 = findAllyYii($id_1);
	$ally_2 = findAllyYii($id_2);

	if( !$ally_1 || !$ally_2 )
	{
		return false;
	}
	if( $ally_1->aid == $ally_2->aid )
	{
		return true;
	}
	require_once(APP_ROOT_DIR."game/Relation.class.php");
	$Relations = new Relation( $id_1 , $ally_1->aid );
	return 4 == $Relations->getAllyRelation($ally_2->aid);
}

function calculateDistance( $start_coords, $end_coords )
{
	if(
		!is_array($start_coords) || empty($start_coords)
			|| !is_array($end_coords) || empty($end_coords)
	)
	{
		return 5;
	}
	$data = array(
		'galaxy'	=> array('mult' => 20000, 'summ' => 0),
		'system'	=> array('mult' => 5 * 19, 'summ' => 2700),
		'position'	=> array('mult' => 5, 'summ' => 1000),
	);
	foreach( $data as $ar_key => $numbers )
	{
		if(
			!isset($start_coords[$ar_key]) || empty($start_coords[$ar_key])
				|| !isset($end_coords[$ar_key]) || empty($end_coords[$ar_key])
		)
		{
			continue;
		}

		if( $end_coords[$ar_key] - $start_coords[$ar_key] != 0 )
		{
			return abs($end_coords[$ar_key] - $start_coords[$ar_key]) * $numbers['mult'] + $numbers['summ'];
		}
	}
	return 5;
}

/**
 * Finds record about ally, using user_id
 * @param int $user_id User, whos ally we are searching.
 */
function findAllyYii( $user_id )
{
	if( !$user_id || empty($user_id) )
	{
		return false;
	}
	$crit = new CDbCriteria();
	$crit->addCondition('userid=:userid');
	$crit->params = array( ':userid' => $user_id );
	$ally = User2ally_YII::model()->find($crit);
	if( !$ally )
	{
		return false;
	}
	return $ally;
}


function getPlanetCreditRatio()
{
	return NS::getPlanet()->getCreditRatio();
}

function logShipsChanging($unitid, $planetid, $quantity, $is_adding, $message)
{
//	try
//	{
//		if(!mt_rand(0, 100))
//		{
//			Core::getQuery()->delete("unit2shipyard_log", "created < ".sqlVal(date("Y-m-d H:i:s", time() - 60*60*24*7)));
//		}
//
//		if(0)
//		{
//			$res = sqlSelect("unit2shipyard", "quantity", "", "unitid = ".sqlVal($unitid)." AND planetid = ".sqlVal($planetid));
//			$cur = sqlFetch($res);
//			sqlEnd($res);
//			$old_quantity = !empty($cur) ? $cur['quantity'] : 0;
//			$new_quantity = $is_adding ? $old_quantity + $quantity : $quantity;
//
//			Core::getQuery()->insert("unit2shipyard_log",
//				array("created", "unitid", "planetid", "quantity", "is_adding", "new_quantity", "old_quantity", "message"),
//				array(date("Y-m-d H:i:s"), $unitid, $planetid, $quantity, $is_adding, $new_quantity, $old_quantity, $message));
//		}
//	}
//	catch(Exception $e)
//	{
//		t("Can't log ship changing. unitid: $unitid, planetid: $planetid, quantity: $quantity, is_adding: $is_adding, message: $message", 'warning');
//	}
}

function isPlanetOccupied($planet_id)
{
	if(empty($planet_id))
	{
		return false;
	}
	$count = sqlSelectField("planet", "count(*)", "", "planetid = ".sqlVal($planet_id)." AND userid is not null");
	return $count == 1;
}

/**
* Get the max count buildings in the list of constructions
*
* @return integer
*/
function getMaxCountListConstructions()
{
	$max_count = 5;
	return $max_count;
}

/**
* Get the max count research in the list of research
*
* @return integer
*/
function getMaxCountListResearch()
{
	$max_count = 5;
	return $max_count;
}

/**
* Get the max count constructions in the list of shipyard
*
* @return integer
*/
function getMaxCountListShipyard()
{
	$max_count = 5;
	return $max_count;
}

/**
* Get the max count constructions in the list of defense
*
* @return integer
*/
function getMaxCountListDefense()
{
	$max_count = 5;
	return $max_count;
}

/**
* Get the max count constructions in the list of repair
*
* @return integer
*/
function getMaxCountListRepair()
{
	$max_count = 5;
	return $max_count;
}

/**
* Get the credit for immediately start in the constructions
*
* @return integer
*/
function getCreditImmStartConstructions($level)
{
	$credit = max(1, round(pow($level, $level < 30 ? 1.3 : ($level < 35 ? 2 : 2.5)) * 5, -1));
	return $credit;
}

/**
* Get the credit for immediately start in the research
*
* @return integer
*/
function getCreditImmStartResearch($level)
{
	$credit = max(1, round(pow($level, $level < 30 ? 1.3 : ($level < 35 ? 2 : 2.5)) * 10, -1));
	return $credit;
}

/**
* Get the credit for immediately start in the shipyard
*
* @return integer
*/
function getCreditImmStartShipyard($quantity)
{
	$quantity = round(pow($quantity, 0.8), -1);
	$credit = clampVal($quantity, 10, 100000);
	return $credit;
}

function getCreditAbortShipyard($quantity, $save_scale)
{
	$quantity = round(pow($quantity * $save_scale, 0.8), -1);
	$credit = clampVal($quantity, 10, 100000);
	return $credit;
}

/**
* Get the credit for immediately start in the repair
*
* @return integer
*/
function getCreditImmStartRepair($quantity)
{
	return getCreditImmStartShipyard($quantity);
}

function getCreditImmStartDisassemble($quantity)
{
	return getCreditImmStartShipyard($quantity);
}

function getCreditAbortRepair($quantity, $save_scale)
{
	return getCreditAbortShipyard($quantity, $save_scale);
}

/**
* Get the max count constructions in the order of shipyard
*
* @return integer
*/
function getMaxCountOrderShipyard()
{
	$max_count = MAX_BUILDING_ORDER_UNITS;
	return $max_count;
}

/**
* Get the max count constructions in the order of defense
*
* @return integer
*/
function getMaxCountOrderDefense()
{
	$max_count = MAX_BUILDING_ORDER_UNITS;
	return $max_count;
}

function getUnitFields($row)
{
	$struct = $row["basic_metal"] + $row["basic_silicon"];
	return $struct > 0 ? max(1, floor($struct / 1000)) : 0;
}

function getDockUnitsCapacity($row, $repair_storage = null)
{
	$fields = getUnitFields($row);
	if($fields > 0)
	{
		if(is_null($repair_storage))
		{
			$repair_storage = NS::getRepairStorage();
		}
		return floor($repair_storage / $fields);
	}
	return 0;
}

function getUnitImage($name)
{
	try
	{
		$imagepackage = NS::getUser() ? NS::getUser()->get("imagepackage") : null;
	}
	catch(Exception $e){}
	if(empty($imagepackage))
	{
		$imagepackage = "novax";
	}
	if($imagepackage == "empty")
	{
		return "buildings/empty/empty.gif";
	}
	if(strpos($name,"__") !== false)
	{
		$name = substr($name, 0, strpos($name,"__"));
	}

	$name = strtolower($name);
	if( !defined('SN') )
	{
		$all_img_packs = array_unique(array($imagepackage, "std"));
	}
	else
	{
		$all_img_packs = array("std");
	}
	foreach($all_img_packs as $imagepackage)
	{
		foreach(array("gif", "png") as $ext)
		{
			$image = "buildings/{$imagepackage}/{$name}.{$ext}";
			if(file_exists(APP_ROOT_DIR."images/{$image}"))
			{
				return $image;
			}
		}
	}
	if(substr($name, 0, 7) == "achiev_")
	{
		return "buildings/std/def_achiev.png";
	}
	return "buildings/empty/empty.gif";
}

function sortByOrder($a, $b)
{
	if($a["order"] < $b["order"]) return -1;
	if($a["order"] > $b["order"]) return 1;
	return strcasecmp($a["name"], $b["name"]);
}

function getImgPacks()
{
	if( !defined('SN') )
	{
		$imgpacks = array();
		$handle = opendir(APP_ROOT_DIR."images/buildings/");
		while($dir = readdir($handle))
		{
			if($dir != "." && $dir != ".." && is_dir(APP_ROOT_DIR."images/buildings/".$dir))
			{
				$item = array("dir" => $dir, "name" => $dir, "order" => 999999, "selected" => intval($dir == NS::getUser()->get("imagepackage")));
				$descr = @file(APP_ROOT_DIR."images/buildings/".$dir."/description.txt");
				if(is_array($descr))
				{
					$item["name"] = isset($descr[0]) ? trim($descr[0]) : $item["name"];
					$item["order"] = isset($descr[1]) ? trim($descr[1]) : $item["order"];
				}
				$imgpacks[] = $item;
			}
		}
		closedir($handle);

		usort($imgpacks, "sortByOrder");
	}
	else
	{
		$imgpacks = array(array(
			'dir' => 'std',
			'name' => 'Standard',
			'order' => '1',
			'selected' => '1',
		));
	}
	return $imgpacks;
}

function getBgImagePlaneStyles($name)
{
	$path = APP_ROOT_DIR . "images/bg/";

	$styles = array();
	$handle = opendir($path);
	while(false !== ($filename = readdir($handle)))
	{
		if(!is_file($path . $filename) || !preg_match("#^(".preg_quote($name, "#")."-?(\d+))#is", $filename, $regs))
		{
			continue;
		}
		$styles[] = array(
			"path" => "us_bg/".$regs[1].".css",
			"name" => $regs[1],
			"order" => $regs[2]
		);
	}
	closedir($handle);

	return $styles;
}

function getUserStyles($type)
{
	$css_dir_web = "us_$type/";
	$css_dir_local = APP_ROOT_DIR . "css/$css_dir_web";

	$styles = array();
	if (!is_dir($css_dir_local)) {
		return $styles;
	}
	$handle = opendir($css_dir_local);
	if ($handle === false) {
		return $styles;
	}
	while(false !== ($file = readdir($handle)))
	{
		if(!is_file($css_dir_local . $file))
		{
			continue;
		}
		$chunk = explode(".", $file);
		switch($chunk[count($chunk)-1])
		{
			case "css":
				$item = array("path" => $css_dir_web.$file, "name" => $file, "order" => 999999, /*"selected" => intval($dir == NS::getUser()->get("imagepackage"))*/);
				$f = $css_dir_local . $chunk[0] . ".txt";
				$descr = @file($f);
				if(is_array($descr))
				{
					$item["name"] = isset($descr[0]) ? trim($descr[0]) : $file;
					$item["order"] = isset($descr[1]) ? trim($descr[1]) : $item["order"];
				}
				$styles[] = $item;
				break;

			case "php":
				if($type == "bg")
				{
					$styles = array_merge($styles, getBgImagePlaneStyles($chunk[0]));
				}
				break;
		}
	}
	closedir($handle);

	usort($styles, "sortByOrder");

	$rStyles = array();
	foreach($styles as $style)
	{
		$rStyles[$style["path"]] = $style;
	}
	return $rStyles;
}

function getTracks($byAlbums = false)
{
	$tracks = array();
	$result = sqlSelect("tracks", "*");
	while($row = sqlFetch($result))
	{
		if (!$row["album"])
		{
			$row["album"] = Core::getLanguage()->getItem("unknown_album");
		}
		if (!$byAlbums)
		{
			$tracks[] = $row;
			continue;
		}
		if (!isset($tracks[$row["album"]]))
		{
			$tracks[$row["album"]] = array();
		}
		$tracks[$row["album"]][] = $row;
	}
	sqlEnd($result);
	return $tracks;
}

function getTrackById($id)
{
	return sqlSelectRow("tracks", "*", "", "trackid = ".sqlVal($id));
}

function saveAssaultReportSID()
{
	Core::getRequest()->setCookie("_sid", SID, 24, "/AssaultReport.php");
}

function getConstructionDesc($id, $unit_type = null)
{
	$type_where = "";
	if(is_array($unit_type))
	{
		$type_where = " AND mode in (".sqlArray($unit_type).")";
	}
	else if(is_scalar($unit_type))
	{
		$type_where = " AND mode = ".sqlVal($unit_type);
	}
	$row = sqlSelectRow("construction", "*", "", "buildingid = ".sqlVal($id).$type_where);
	return $row;
}

function getShipyardQuantity($id, $return_row = false, $planetid = null)
{
	$row = sqlSelectRow("unit2shipyard", "*", "",
		"unitid = ".sqlVal($id)." AND planetid = ".(is_null($planetid) ? sqlPlanet() : sqlVal($planetid)));

	if($return_row)
	{
		return $row;
	}

	return empty($row["quantity"]) ? 0 : $row["quantity"];
}

function fixUnitDamagedVars(&$unit)
{
	if(!isset($unit["damaged"]))
	{
		$unit["damaged"] = 0;
		$unit["shell_percent"] = 100;
	}
}

function getUnitQuantityStr($unit, $params = array())
{
	$str = fNumber($unit["quantity"]);
	if(isset($unit["damaged"]) && $unit["damaged"] > 0 && empty($params["no_damaged"]))
	{
		$params = array_merge(array(
			"nobr" => true,
			"bracket" => "["
			), (array)$params);

		$nobr = !empty($params["nobr"]);
		$bracket = isset($params["bracket"]) ? $params["bracket"] : false;
		if($bracket === true || $bracket == "(")
		{
			$bracket_begin = "(";
			$bracket_end = ")";
		}
		else if($bracket == "[")
		{
			$bracket_begin = "[";
			$bracket_end = "]";
		}
		else
		{
			$bracket_begin = $bracket_end = "";
		}

		if(!empty($params["no_quantity"]))
		{
			$unit["quantity"] = -1;
			$str = "";
		}

		$span_class = $unit["shell_percent"] <= 70 ? "rep_quantity_damage" : "rep_quantity_damage_low";

		if(isset($unit["grasped"]) && $unit["grasped"] > 0)
		{
			if($unit["grasped"] == $unit["quantity"])
			{
				$str = "<span class='rep_quantity_grasped'>+{$str}</span>";
			}
			else
			{
				$str .= " <span class='rep_quantity_grasped'>";
				$str .= $bracket ? $bracket_begin : "";
				$str .= "+".fNumber($unit["grasped"]);
				$str .= $bracket ? $bracket_end : "";
				$str .= "</span>";
			}
		}

		$str = $nobr ? "<nobr>{$str}" : $str;
		$str .= (isset($params["splitter"]) ? $params["splitter"] : "") . " ";
		$str .= "<span class='{$span_class}'>";
		$str .= $bracket ? $bracket_begin : "";
		$str .= ($unit["damaged"] != $unit["quantity"] ? fNumber($unit["damaged"])." - " : "").fNumber($unit["shell_percent"])."%";
		$str .= $bracket ? $bracket_end : "";
		$str .= "</span>";
		$str .= $nobr ? "</nobr>" : "";
	}
	return $str;
}

/**
* Generates readable list of a ship array.
*
* @param array		Ship data
*
* @return string	Ship list
*/
function getUnitListStr($units, $params = array())
{
	$items = array();
	if(is_array($units))
	{
		foreach($units as $key => $unit)
		{
			$temp_str = Core::getLanguage()->getItem($unit["name"]);
			if ( ( $key == ARTEFACT_PACKED_BUILDING || $key == ARTEFACT_PACKED_RESEARCH ) && isset($unit['art_ids']) )
			{
			$art_strs = array();
				foreach ( $unit['art_ids'] as $a_key => $value )
				{
					$new_key = $value["con_id"] . '_' . $value["level"];
					$art_strs[ $new_key ]['text'] = ' "' . Core::getLanguage()->getItem($value["con_name"]) . ' ' . $value["level"] . '"';
					if ( !isset($art_strs[ $new_key ]['quantity']) || empty($art_strs[ $new_key ]['quantity']) )
					{
						$art_strs[ $new_key ]['quantity'] = 0;
					}
					$art_strs[ $new_key ]['quantity'] += 1;
				}
				foreach ( $art_strs as $a_key => $value )
				{
					$items[] = $temp_str . $value['text'] . ": ".getUnitQuantityStr($value, array());
				}
			}
			else
			{
				$temp_str .= ": ".getUnitQuantityStr($unit, $params);
				$items[] = $temp_str;
			}
		}
	}
	return implode(", ", $items);
}

/**
 * Returns array
 * 	[user] => user_id of this art,
 * 	[name] => full name of this art
 * 	[pos]	=> location of this art. IN_FLIGHT if is in flight
 * 	[link] => last part of the link to artefact_info
 */
function getArtefactNameAndPosition( $art_id )
{
		$art_data = sqlSelectRow(
			'artefact2user AS a2u',
			'a2u.userid, a2u.typeid AS art_type, c.name AS art_Name, g.galaxy, g.system, g.position, co.name AS pack_Name, a2u.level, a2u.planetid AS currPlanet, a2u.construction_id AS pack_type',
			'LEFT JOIN '. PREFIX .'construction AS c ON c.buildingid = a2u.typeid '.
				'LEFT JOIN '. PREFIX .'galaxy AS g ON g.planetid = a2u.planetid '.
				'LEFT JOIN '. PREFIX .'construction AS co ON co.buildingid = a2u.construction_id ',
			'a2u.artid = '.$art_id
		);
		$result = array();
		$result['user'] = $art_data['userid'];
		if ( empty($art_data['currPlanet']) || $art_data['currPlanet'] == 0 )
		{
			$result['pos'] = Core::getLanguage()->getItem('IN_FLIGHT');
		}
		else
		{
			$result['pos'] = getCoordLink(0, 0, 0, false, $art_data['currPlanet']);
		}
		$result['link'] = $art_data['art_type'];
		$result['name'] = Core::getLanguage()->getItem($art_data['art_Name']);
		if ( !empty($art_data['pack_Name']) && $art_data['pack_Name'] !== 0 )
		{
			$result['name'] .= ' "';
			$result['name'] .= Core::getLanguage()->getItem($art_data['pack_Name']);
			if ( !empty($art_data['level']) && $art_data['level'] != 0 )
			{
				$result['name'] .= ': ';
				$result['name'] .= $art_data['level'];
				$result['link'] .= '_' . $art_data['level'];
			}
			$result['link'] .= '_' . $art_data['pack_type'];
			$result['name'] .= '"';
		}
		return $result;
}

function accumulateUnits(&$dest, $add)
{
	fixUnitDamagedVars($dest);

	$dest["quantity"] += $add["quantity"];
	if(isset($add["damaged"]) && $add["damaged"] > 0)
	{
		$new_shell_percent = ($dest["shell_percent"] * $dest["damaged"]
				+ $add["shell_percent"] * $add["damaged"]) / max(1, $dest["damaged"] + $add["damaged"]);
		$dest["damaged"] += $add["damaged"];
		$dest["shell_percent"] = $new_shell_percent;
	}
}

function extractUnits(&$units, $quantity, &$dest)
{
	fixUnitDamagedVars($units);

	$quantity = max(0, $quantity);
	$quantity = min($quantity, $units["quantity"]);

	$units["quantity"] -= $quantity;
	$dest["quantity"] = $quantity;
	if($units["damaged"] > $units["quantity"])
	{
		$dest["damaged"] = $units["damaged"] - $units["quantity"];
		$dest["shell_percent"] = $units["shell_percent"];
		$units["damaged"] -= $dest["damaged"];
	}
	else
	{
		$dest["damaged"] = 0;
		$dest["shell_percent"] = 100;
	}
}

function isDeathmatchStarted()
{
    return DEATHMATCH && defined('DEATHMATCH_START_TIME')
        && (DEATHMATCH_START_TIME == 0 || DEATHMATCH_START_TIME < time());
}

function isDeathmatchEnd()
{
    return DEATHMATCH && defined('DEATHMATCH_END_TIME')
        && (DEATHMATCH_END_TIME == 0 || DEATHMATCH_END_TIME < time());
}

function updateDmPointsSetSql()
{
    // return " dm_points = points * GREATEST(1, e_points) * ".sqlVal(DM_POINTS_BATTLE_EXP_SCALE);
    /*
    $scale = sqlVal(DM_POINTS_BATTLE_EXP_SCALE);
    $power = sqlVal(DM_MAX_POINTS_POWER);
    return "
        max_points = GREATEST(max_points, points),
        dm_points = e_points * $scale / POW(GREATEST(100, max_points), $power)";
     *
    return "
        max_points = GREATEST(max_points, points),
        dm_points = e_points * points / POW(GREATEST(1, max_points), 0.93)";
     *
    return "
        max_points = GREATEST(max_points, points),
        dm_points =  LEAST(e_points, POW(points, 0.4)) * points / POW(GREATEST(1, max_points), 0.9)";
     */
	// POW( GREATEST( LEAST( e_points, 100 ), LEAST( e_points, POW( points /3000, 1.1 ) + e_points /100 ) ) * points / POW( GREATEST( 1, max_points ), 0.9 ), 0.7 ) * 10
    // POW( GREATEST( LEAST( e_points, 100 ), LEAST( e_points, POW( points /4000, 1.1 ) + e_points /100 ) ) * points / POW( GREATEST( 1, max_points ), 0.9 ), 0.5 ) * 100
    return "
        max_points = GREATEST(max_points, points),
        dm_points =  POW( GREATEST( LEAST( e_points, 100 ), LEAST( e_points, POW( points /4000, 1.1 ) + e_points /100 ) ) * points / POW( GREATEST( 1, max_points ), 0.9 ), 0.5 ) * 100";
}

function addQuantitySetSql($add)
{
	fixUnitDamagedVars($add);
	$add["quantity"] = isset($add["quantity"]) ? max($add["quantity"], $add["damaged"]) : $add["damaged"];
	return " shell_percent = (damaged * shell_percent + ".sqlVal($add['damaged'] * $add['shell_percent']).") / GREATEST(1, damaged + ".sqlVal($add['damaged']).")"
		. ", quantity = quantity + ".sqlVal($add["quantity"])
		. ", damaged = damaged + ".sqlVal($add["damaged"]);
}

function subQuantitySetSql($sub)
{
	if(is_array($sub))
	{
		if(isset($sub["damaged"]))
		{
			$sub["quantity"] = isset($sub["quantity"]) ? max($sub["quantity"], $sub["damaged"]) : $sub["damaged"];
			return " quantity = quantity - ".sqlVal($sub["quantity"]).", damaged = GREATEST(0, damaged - ".sqlVal($sub["damaged"]).")";
		}
		$sub = $sub["quantity"];
	}
	return " quantity = quantity - ".sqlVal($sub).", damaged = GREATEST(0, damaged - ".sqlVal($sub).")";
}

function sqlUser()
{
	return sqlVal(NS::getUser()->get("userid"));
}

function sqlPlanet()
{
	return sqlVal(NS::getUser()->get("curplanet"));
}

function createPageLink($page, $current = false, $name = false, $funct_prefix = '')
{
	if(!$name)
	{
		$name = $page;
	}
	$class = ' class="page_link"';
	$nbsp = '';
	if($current)
	{
		$class = ' class="page_link, true"';
		$nbsp = '&nbsp;';
	}
	return "$nbsp<a href='#' onclick='{$funct_prefix}goPage($page); return false'$class>$name</a>";
}

function clampVal($i, $min, $max)
{
	if($i < $min) return $min;
	if($i > $max && $max > $min) return $max;
	return $i;
}

function randFloat()
{
	return abs((float)mt_rand() / (float)mt_getrandmax());
}

function randSignFloat()
{
	return randFloat() * 2 - 1;
}

function randFloatRange($a, $b)
{
	return $a + ($b - $a) * randFloat();
}

function randRoundRange($a, $b)
{
	return round(randFloatRange($a, $b));
}

function toUtf8($s)
{
	return iconv('cp1251', 'utf-8//IGNORE', $s);
}

function getTech($id, $userid)
{
	return NS::getResearch($id, $userid);
}

function getBuilding($id, $planetid)
{
	return NS::getBuilding($id, $planetid);
}

function addShips($id, $num, $planetid)
{
	$result_count = sqlSelectField("unit2shipyard", "count(*)", "", "unitid = ".sqlVal($id)." AND planetid = ".sqlVal($planetid));
	if($result_count > 0)
	{
		// Update By Pk
		sqlQuery("UPDATE ".PREFIX."unit2shipyard SET quantity = quantity + ".sqlVal($num)." WHERE planetid = ".sqlVal($planetid)." AND unitid = ".sqlVal($id));
	}
	else
	{
		Core::getQuery()->insert("unit2shipyard", array("unitid", "planetid", "quantity"), array($id, $planetid, $num));
	}
}

/**
* Задаёт переменные шаблона: page_links, pages, link_next, link_last, link_prev, link_first.
* @param int $pages		Всего страниц
* @param int $page		Текущая страница
* @param int $per_page	Элементов на страницу
*
* @return int индекс строки, с кот. начинать выборку.
*/
function createPaginator($pages, $page, $per_page)
{
	if(!is_numeric($page))
	{
		$page = 1;
	}
	else if($page > $pages) { $page = $pages; }
	else if($page < 1) { $page = 1; }

	$pages_to_show = 7;
	$pages_range = floor($pages_to_show / 2);
	$pages_link = "";
	$pages_sel = "";
	for($i = 0; $i < $pages; $i++)
	{
		$i1 = $i+1;
		$n = $i * $per_page + 1;
		if($i1 == $page) { $s = 1; } else { $s = 0; }
		$pages_sel .= createOption($i + 1, $i1, $s);
		if ((abs($i1 - $page) <= $pages_range) && ($pages_to_show-- > 0))
			$pages_link .= createPageLink($i1, $s, "[$i1]");
	}
	if ($page != $pages)
	{
		Core::getTPL()->assign("link_next", createPageLink($page +1, false, Core::getLanguage()->getItem("NEXT_PAGE")));
		Core::getTPL()->assign("link_last", createPageLink($pages, false, "&gt;&gt;"));
	}
	if ($page != 1)
	{
		Core::getTPL()->assign("link_prev", createPageLink($page -1, false, Core::getLanguage()->getItem("PREV_PAGE")));
		Core::getTPL()->assign("link_first", createPageLink(1, false, "&lt;&lt;"));
	}
	Core::getTPL()->assign("page_links", $pages_link);
	Core::getTPL()->assign("pages", $pages_sel);
	//TODO: проверить $start
	$start = abs(($page - 1) * $per_page);
	$max = $per_page;

	return $start;
}

function recycleDebris(&$data)
{
	$capacity = max(0, round($data["capacity"]));

	$rest_metal = $data["debrismetal"] = max(0, round($data["debrismetal"]));
	$rest_silicon = $data["debrissilicon"] = max(0, round($data["debrissilicon"]));
	$rest_hydrogen = $data["debrishydrogen"] = max(0, round($data["debrishydrogen"]));

	// $log = beginLog("[recycleDebris] capacity: $capacity, debris: $rest_metal, $rest_silicon, $rest_hydrogen");

	while($capacity > 0 && ($rest_metal > 0 || $rest_silicon > 0 || $rest_hydrogen > 0))
	{
		if($rest_metal > 0)
		{
			$cnt = 1 + (int)($rest_silicon > 0) + (int)($rest_hydrogen > 0);
			$res = min($rest_metal, ceil($capacity / $cnt));
			$capacity -= $res;
			$rest_metal -= $res;
		}
		if($rest_silicon > 0)
		{
			$cnt = 1 + (int)($rest_hydrogen > 0);
			$res = min($rest_silicon, ceil($capacity / $cnt));
			$capacity -= $res;
			$rest_silicon -= $res;
		}
		if($rest_hydrogen > 0)
		{
			$res = min($rest_hydrogen, $capacity);
			$capacity -= $res;
			$rest_hydrogen -= $res;
		}
	}
	foreach(array("metal", "silicon", "hydrogen") as $res_name)
	{
		$data["recycled".$res_name] = $data["debris".$res_name] - ${"rest_".$res_name};
		$data[$res_name] = (isset($data[$res_name]) ? $data[$res_name] : 0) + $data["recycled".$res_name];
	}

	// endLog($log, $log["message"] . "; recycled: {$data['recycledmetal']}, {$data['recycledsilicon']}, {$data['recycledhydrogen']}");

	return $capacity;
}

function updateMinerPoints($userid, $points, $add_credit)
{
	$row = sqlSelectRow("user", array("userid", "of_points", "of_level"), "", "userid = ".sqlVal($userid));

	$need = 200;
	if(empty($row["of_level"])) $row["of_level"] = 0;

	$need_points = ($row["of_level"] < 1) ? $need / 2 : round(pow(1.5, $row["of_level"] - 1) * $need);

	$now_points = $row["of_points"] + $points;

	$credit = 0;
	// Loop for adding level and calculating credits to update
	while($now_points >= $need_points)
	{
		$now_points -= $need_points;
		$row["of_level"]++;
		// INFO: no CREDIT here
		if( $add_credit )
		{
			if ( !isset($GLOBALS['POINTS_PER_MINNING_LEVEL'][$row["of_level"]]) )
			{
				$credit += end($GLOBALS['POINTS_PER_MINNING_LEVEL']);
			}
			else
			{
				$credit += $GLOBALS['POINTS_PER_MINNING_LEVEL'][$row["of_level"]];
			}
		}
		$need_points = round(pow(1.5, $row["of_level"] - 1) * $need);
	}

	// Update By Pk
	sqlQuery("UPDATE ".PREFIX."user SET ".
		" credit = credit + ".sqlVal($credit).
		", of_points = ".sqlVal($now_points).
		", of_level = ".sqlVal($row["of_level"]).
		" WHERE userid = ".sqlVal($userid)
	);

	if( $credit > 0 )
	{
		new AutoMsg(MSG_CREDIT, $userid, time(), array('credits' => $credit, 'msg' => 'MSG_CREDIT_FOR_MINER' ));
	}
}

function unitGroupConsumptionPerHour($unitid, $count, $in_fling = false)
{
	if(!$in_fling && $count < 1000)
	{
		return 0;
	}
	// 1,000003^500000 = 4,481678986569
	// 1,000003219^500000 = 5,000297494256
	// 1,000003^500000 = 4,481678986569
	$scale = $unitid == UNIT_VIRT_DEFENSE ? 0.5 : 1.0;
	$cons = min(MAX_GROUP_UNIT_CONSUMTION_PER_HOUR, pow( UNITS_GROUP_CONSUMTION_POWER_BASE, $count ) * $scale / 10 / 24.0) * $count;
	return $cons >= 0.01 ? $cons : 0;
}

function starGateRecycleSecs($star_gate_level)
{
	// sync to table construction, id = 56, special
	return 60*60 * pow(0.7, max(0, $star_gate_level-1));
}

function logMessage($message)
{
	return sqlInsert("log", array(
		"message" => $message,
	));
}

function updateLogDeltaTime($logid, $dt, $message = null)
{
	$data = array("dt" => $dt);
	if(!empty($message))
	{
		$data["message"] = $message;
	}
	// Update By Pk
	sqlUpdate("log", $data, "logid=".sqlVal($logid));
}

function beginLog($message)
{
	return array(
		"start_time_sec" => microtime(true),
		"logid" => logMessage($message),
		"message" => $message,
		);
}

function endLog($log_data, $message = null)
{
	if(isset($log_data["logid"]) && isset($log_data["start_time_sec"]))
	{
		updateLogDeltaTime($log_data["logid"], microtime(true) - $log_data["start_time_sec"], $message);
	}
}

function isFacebookSkin()
{
	if( !defined('SN_FULLSCREEN') )
	{
		if( defined('MOBI') )
		{
			return false;
		}
		if( defined('SN') )
		{
			return true;
		}
	}
	return !empty($_SESSION["userid"]) && $_SESSION["skin_type"] ?? "standard" == SKIN_TYPE_FB;
}

function isMobiSkin()
{
	return isMobileSkin();
}

function isMobileSkin()
{
	if( !defined('SN_FULLSCREEN') )
	{
		if( defined('SN') )
		{
			return false;
		}
		if( defined('MOBI') )
		{
			return true;
		}
	}
	return !empty($_SESSION["userid"]) && $_SESSION["skin_type"] ?? "standard" == SKIN_TYPE_MOBI;
}

function isHomePlanetRequiredForPage($page_name)
{
	static $openPages = null;
	if(empty($openPages))
	{
		foreach(array('Main', 'HomePlanetRequired', /*'Resource', 'Constructions', 'Research', 'Shipyard', 'Defense', 'Mission',
			'Artefacts', 'Repair', 'Disassemble', 'Stock', 'ExchangeOpts',*/ 'Chat', /*'ChatAlly',*/ 'MSG',
			/*'Alliance', 'MemberList',*/ 'Friends', /*'Referral', 'Galaxy', 'Empire',*/ 'Techtree', 'Ranking',
			'Records', 'Battlestats', 'ResTransferStats', 'Simulator', 'AdvTechCalculator', /*'Market',
			'ArtefactMarket',*/ 'Payment', 'Search', /*'PlanetOptions',*/ 'Prefs', 'UserAgreement', 'Support'
			) as $name)
		{
			$openPages[strtolower($name)] = 1;
		}
	}
	return !isset($openPages[strtolower($page_name)]);
}

function isArgeementCanBeShownForPage($page_name)
{
	static $list = null;
	if(empty($list))
	{
		foreach(array('Main', /*'HomePlanetRequired',*/ 'Resource', 'Constructions', 'Research', 'Shipyard', 'Defense', 'Mission',
			'Artefacts', 'Repair', 'Disassemble', 'Stock', 'ExchangeOpts', 'Chat', 'ChatAlly', 'MSG',
			'Alliance', 'MemberList', 'Friends', 'Referral', 'Galaxy', 'Empire', 'Techtree', 'Ranking',
			'Records', 'Battlestats', 'ResTransferStats', 'Simulator', 'AdvTechCalculator', 'Market',
			'ArtefactMarket', /*'Payment'*/ 'Search', 'PlanetOptions', 'Prefs', /*'UserAgreement', 'Support'*/
			) as $name)
		{
			$list[strtolower($name)] = 1;
		}
	}
	return isset($list[strtolower($page_name)]);
}

function getArgeementTime()
{
	if(!NS::getMCH()->get("ArgeementTime", $time))
	{
		$time = (int)sqlSelectField("user_agreement", "UNIX_TIMESTAMP(MAX(date))");
		NS::getMCH()->set("ArgeementTime", $time, 60*10);
	}
	return $time;
}

function isAdmin($userid = null, $allow_admin_login = false)
{
	if($allow_admin_login && (!empty($_SESSION["is_admin"])))
	{
		return true;
	}
	if(empty($GLOBALS["ADMIN_USERS"])){
		return false;
	}
	$admin_users = $GLOBALS["ADMIN_USERS"];
	if( !is_null($userid) )
	{
		return in_array($userid, $admin_users);
	}
	return NS::getUser() && in_array(NS::getUser()->get("userid"), $admin_users);
}

function bbcode_text($text, $extended_syntax = true)
{
	$bb = array();
	$bb["#\[:(\d+):\]#si"] = "<img src=".RELATIVE_URL."chat/emo/\\1.gif?".CLIENT_VERSION.">";
	$bb["#\[img\]([^\"\[\]]*?)\[/img\]#si"] = $extended_syntax ? "<img src=\"\\1\" alt=''>" : "<a href=\"\\1\" target='_blank' rel='nofollow'>\\1</a>";
	$bb["#\[url\]([^\"]*?)\[/url\]#si"] = "<a href=\"\\1\" target='_blank' rel='nofollow'>\\1</a>";
	$bb["#\[url=([^\"\[\]]*?)\](.*?)\[/url\]#si"] = "<a href=\"\\1\" target='_blank' rel='nofollow'>\\2</a>";
	if(!$extended_syntax)
	{
		$bb["#\[b\](.*?)\[/b\]#si"] = "<b>\\1</b>";
		$bb["#\[i\](.*?)\[/i\]#si"] = "<i>\\1</i>";
		$bb["#\[u\](.*?)\[/u\]#si"] = "<u>\\1</u>";
		$bb["#\[s\](.*?)\[/s\]#si"] = "<s>\\1</s>";
		$bb["#\[d\](.*?)\[/d\]#si"] = "<span style='text-decoration: line-through'>\\1</span>";
		$bb["#\[blink\](.*?)\[/blink\]#si"] = "<span style='text-decoration: blink'>\\1</span>";
		$bb["#\[color=(\#[a-f0-9]+|\w+)\](.*?)\[/color\]#si"] = "<span style='color:\\1'>\\2</span>";
		$bb["#\[size=([\d\w]+)\](.*?)\[/size\]#si"] = "<span style='font-size:\\1'>\\2</span>";
	}
	else
	{
		$bb["#\[b\]#si"] = "<span style='font-weight:bold'>";
		$bb["#\[i\]#si"] = "<span style='font-style:italic'>";
		$bb["#\[u\]#si"] = "<span style='text-decoration:underline'>";
		$bb["#\[s\]#si"] = "<span style='text-decoration:line-through'>";
		$bb["#\[d\]#si"] = "<span style='text-decoration:line-through'>";
		$bb["#\[blink\]#si"] = "<span style='text-decoration: blink'>";
		$bb["#\[color=(\#[a-f0-9]+|\w+)\]#si"] = "<span style='color:\\1'>";
		$bb["#\[size=([\d\w]+)\]#si"] = "<span style='font-size:\\1'>";

		// $html[] = "<a href=\"javascript:void(0);\" onclick=\"window.open( '\\1' );\" class=\"external\">\\2</a>";
		// $bb["#\[align=(center|right|left|justify)\](.*?)\[/align\]#si"] = "<div style='text-align:\\1'>\\2</div>";
		$bb["#\[align=(center|right|left|justify)\]#si"] = "<div style='text-align:\\1'>";
		// $bb["#{{line}}#si"] = "<hr />";
		$bb["#\[/(color|size|d|blink|b|i|u|s)\]#si"] = "</span>";
		$bb["#\[/align\]#si"] = "</div>";
		$bb["#\[list\]([^\"]*?)\[/list\]#sie"] = 'bbcode_list_items("\\1", \$extended_syntax)';
	}
	foreach($bb as $pat => $rep)
	{
		$text = preg_replace ($pat, $rep, $text);
	}
	return nl2br($text);
}

/**
* Plain BB-Code parser.
*
* @param string	Text to parse
*
* @return string
*/
function parseBBCode($text)
{
	return bbcode_text($text);
	/*
	$patterns = array();
	$patterns[] = "/\[url=(&quot;|['\"]?)([^\"']+)\\1](.+)\[\/url\]/esiU";
	$patterns[] = "/\[url]([^\"\[]+)\[\/url\]/eiU";
	$patterns[] = "/\[img]([^\"]+)\[\/img\]/siU";
	$patterns[] = "/\[b]([^\"]+)\[\/b\]/siU";
	$patterns[] = "/\[u]([^\"]+)\[\/u\]/siU";
	$patterns[] = "/\[i]([^\"]+)\[\/i\]/siU";
	$patterns[] = "/\[blink]([^\"]+)\[\/blink\]/siU";
	$patterns[] = "/\[d]([^\"]+)\[\/d\]/siU";
	$patterns[] = "/\[color=([^\"]+)\]/siU";
	$patterns[] = "/\[size=([^\"]+)\]/siU";
	$patterns[] = "/\[align=(center|right|left|justify)\]/siU";
	$patterns[] = "/\[list](.+)\[\/list\]/esiU";
	$replace = array();
	$replace[] = 'Link::get("\\2", "\\3", "\\2")';
	$replace[] = 'Link::get("\\1", "\\1")';
	$replace[] = "<img src=\"\\1\" alt=\"\" title=\"\" />";
	$replace[] = "<span style=\"font-weight: bold;\">\\1</span>";
	$replace[] = "<span style=\"text-decoration: underline;\">\\1</span>";
	$replace[] = "<i>\\1</i>";
	$replace[] = "<span style=\"text-decoration: blink;\">\\1</span>";
	$replace[] = "<span style=\"text-decoration: line-through;\">\\1</span>";
	$replace[] = "<span style=\"color: \\1;\">";
	$replace[] = "<span style=\"font-size: \\1;\">";
	$replace[] = "<div style=\"text-align: \\1;\">";
	$replace[] = 'parseList("\\1")';
	// Hook::event("PARSE_BB_CODE_START", array(&$text, &$patterns, &$replace));
	$text = preg_replace($patterns, $replace, $text);

	$search = array("[/color]", "[/size]", "[/align]", "{{line}}", "\t");
	$replace = array("</span>", "</span>", "</div>", "<hr />", " ");
	$text = str_replace($search, $replace, $text);
	$text = nl2br($text);
	// Hook::event("PARSE_BB_CODE_END", array(&$text));
	return $text;
	*/
}

/**
* Extended BB-Code parser to generate HTML list.
*
* @param string	Raw code
*
* @return string	HTML
*/
function parseList($list)
{
	return bbcode_list_items($list);
	/*
	if($list == "") { return $list; }
	$list = str_replace("*", "</li>\n<li>", $list);
	if(strstr($list, "<li>")) { $list .= "</li>"; }
	$list = preg_replace("/^.*(<li>)/sU", "\\1", $list);
	return "<ul class=\"ally-list\">".$list."</ul>";
	*/
}

function bbcode_list_items($text, $extended_syntax = true)
{
	$list = "";
	foreach(explode("[*]", $text) as $str)
	{
		$str = trim($str);
		if($str)
		{
			$list .= "<li>".$str."</li>";
		}
	}
	return $list ? "<ul class='ally-list'>".$list."</ul>" : "";
}

function showTopBannerBlock()
{
    $is_admin = isAdmin();
    if(0 && mt_rand(0, 100)){
        return false;
    }
    if(Core::getRequest()->getGET("go") == 'Payment'){
        return false;
    }
    if(NS::getUser() && NS::getUser()->get("regtime") < time()-60*60*24*10)
    {
        $row = sqlSelectRow("payments", "sum(pay_credit) as credit", "",
            "pay_user_id=".sqlUser()." AND pay_date > ".sqlVal(date("Y-m-d", time()-60*60*24*21))." AND pay_status=1");
        return $row["credit"] < 1;
    }
    return $is_admin;
}

function showBottomBannerBlock()
{
	return true;
}

function isNameCharValid($username)
{
    // $charset = Core::getLanguage()->getOpt("charset");
    $chars = preg_quote('~`!@#$%^&*()_+-={}[]\|:;"\'<,>?/№', '#');
    return // !preg_match("#[\s\.{$chars}]v+[\s\.{$chars}]{0,}$#i", $username)
        !preg_match("#против|отпуск|vacation|суки$#iu", $username)
        // && !preg_match("#[aа][dд][mм]in$#iu", $username)
        && !preg_match("#[ёйцукенгшщзхъфывапролджэячсмитьбюqwertyuiopasdfghjklzxcvbnm]V$#u", $username)
        ;
}

?>