<?php
/**
* Recalcultes points.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class PointRenewer
{
	/**
	* Caclulates points of buildings.
	*
	* @param integer	Planet id
	*
	* @return float	Points
	*/
	public static function getPlanetBuildingStats($planetid) // getBuildingPoints
	{
		$result = sqlSelect(
			"building2planet b2p",
			array(
				"b2p.level",
				"c.buildingid",
				"c.mode",
				"c.basic_metal",
				"c.basic_silicon",
				"c.basic_hydrogen",
				"c.charge_metal",
				"c.charge_silicon",
				"c.charge_hydrogen"
			),
			"LEFT JOIN ".PREFIX."construction c ON (c.buildingid = b2p.buildingid)",
			"b2p.planetid = ".sqlVal($planetid)." AND (c.mode = ".UNIT_TYPE_CONSTRUCTION." OR c.mode = ".UNIT_TYPE_MOON_CONSTRUCTION.")"
		);
		return self::getChargeStats($result, RES_TO_BUILD_POINTS);
	}

	private static function getChargeStats($query_handle, $points_scale)
	{
		$count = 0;
		$points = 0;
		while($row = sqlFetch($query_handle))
		{
			if( isset($row["mode"]) )
			{
				switch($row["mode"])
				{
				case UNIT_TYPE_CONSTRUCTION:
				case UNIT_TYPE_MOON_CONSTRUCTION:
					$row["level"] = min($row["level"], MAX_BUILDING_LEVEL);
					break;

				case UNIT_TYPE_RESEARCH:
					$row["level"] = min($row["level"], MAX_RESEARCH_LEVEL);
					break;
				}
			}
			if( isset($row["buildingid"]) && isset($GLOBALS["MAX_UNIT_LEVELS"][$row["buildingid"]]) )
			{
				$row["level"] = min($row["level"], $GLOBALS["MAX_UNIT_LEVELS"][$row["buildingid"]]);
			}
			for($i = 1; $i <= $row["level"]; $i++)
			{
				if($row["basic_metal"] > 0)
				{
					$points += parseChargeFormula($row["charge_metal"], $row["basic_metal"], $i);
				}
				if($row["basic_silicon"] > 0)
				{
					$points += parseChargeFormula($row["charge_silicon"], $row["basic_silicon"], $i);
				}
				if($row["basic_hydrogen"] > 0)
				{
					$points += parseChargeFormula($row["charge_hydrogen"], $row["basic_hydrogen"], $i);
				}
				$count++;
			}
		}
		sqlEnd($query_handle);
		return array("points" => round($points * $points_scale, POINTS_PRECISION), "count" => $count);
	}

	/**
	* Caclulates points of research.
	*
	* @param integer	User id
	*
	* @return float	Points
	*/
	public static function getUserResearchStats($userid) // getResearchPoints
	{
		$result = sqlSelect("research2user r2u",
			array(
				"r2u.level",
				"c.buildingid", "c.mode",
				"c.basic_metal", "c.basic_silicon", "c.basic_hydrogen",
				"c.charge_metal", "c.charge_silicon", "c.charge_hydrogen"
			),
			"LEFT JOIN ".PREFIX."construction c ON (c.buildingid = r2u.buildingid)",
			"r2u.userid = ".sqlVal($userid)." AND c.mode = ".UNIT_TYPE_RESEARCH);
		return self::getChargeStats($result, RES_TO_RESEARCH_POINTS);
	}

	/**
	* Caclulates points of fleet.
	*
	* @param integer	Planet id
	*
	* @return float	Points
	*/
	public static function getPlanetFleetStats($planetid) // getFleetPoints
	{
		$count = 0;
		$points = 0;
		$result = sqlSelect("unit2shipyard u2s", array("u2s.quantity", "u2s.damaged", "u2s.shell_percent", "c.basic_metal", "c.basic_silicon", "c.basic_hydrogen"), "LEFT JOIN ".PREFIX."construction c ON (c.buildingid = u2s.unitid)", "u2s.planetid = ".sqlVal($planetid)." AND (c.mode = ".UNIT_TYPE_FLEET." OR c.mode = ".UNIT_TYPE_DEFENSE.")");
		while($row = sqlFetch($result))
		{
			$basic_res = $row["basic_metal"] + $row["basic_silicon"] + $row["basic_hydrogen"];
			$points += $basic_res * $row["quantity"]; // ($row["quantity"] - $row["damaged"]) + $basic_res * $row["damaged"] * $row["shell_percent"] / 100;
			$count += $row["quantity"];
		}
		sqlEnd($result);
		return array("points" => round($points * RES_TO_UNIT_POINTS, POINTS_PRECISION), "count" => $count);
	}

	/**
	* Caclulates points of fleet from events.
	*
	* @param integer	User id
	*
	* @return float	Points
	*/
	public static function getUserFleetEventStats($userid) // getFleetEventPoints
	{
		$count = 0;
		$points = 0;
		$result = sqlSelect("events", "data", "",
			"mode >= ".EVENT_MARK_FIRST_FLEET." AND mode <= ".EVENT_MARK_LAST_FLEET
			. " AND user = ".sqlVal($userid)
			. " AND processed=".EVENT_PROCESSED_WAIT);
		while($row = sqlFetch($result))
		{
			$row["data"] = unserialize($row["data"]);
			if(!is_array($row["data"]["ships"]))
			{
				continue;
			}
			foreach($row["data"]["ships"] as $ship)
			{
				$shipData = sqlSelectRow("construction", array("basic_metal", "basic_silicon", "basic_hydrogen"), "", "buildingid = ".sqlVal($ship["id"]));
				$basic_res = $shipData["basic_metal"] + $shipData["basic_silicon"] + $shipData["basic_hydrogen"];
				$points += $basic_res * $ship["quantity"];
				$count += $ship["quantity"];
			}
		}
		sqlEnd($result);
		return array("points" => round($points * RES_TO_UNIT_POINTS, POINTS_PRECISION), "count" => $count);
	}

	/**
	* Caclulates points of fleet from lots.
	*
	* @param integer	User id
	*
	* @return float	Points
	*/
	public static function getUserLotStats($userid) // getFleetEventPoints
	{
		$count = 0;
		$points = 0;
		$result = sqlSelect("exchange_lots l", "data",
			"JOIN ".PREFIX."planet p on l.planetid = p.planetid",
			"p.userid=".sqlVal($userid)." AND l.status=".ESTATUS_OK
		);
		while($row = sqlFetch($result))
		{
			$row["data"] = unserialize($row["data"]);
			if(!is_array($row["data"]["ships"]))
			{
				continue;
			}
			foreach($row["data"]["ships"] as $ship)
			{
				$shipData = sqlSelectRow("construction", array("basic_metal", "basic_silicon", "basic_hydrogen"), "", "buildingid = ".sqlVal($ship["id"]));
				$basic_res = $shipData["basic_metal"] + $shipData["basic_silicon"] + $shipData["basic_hydrogen"];
				$points += $basic_res * $ship["quantity"];
				$count += $ship["quantity"];
			}
		}
		sqlEnd($result);
		return array("points" => round($points * RES_TO_UNIT_POINTS, POINTS_PRECISION), "count" => $count);
	}

	/**
	* Caclulates points of units.
	*
	* @param integer	User id
	*
	* @return float	Points
	*/
	public static function getUnitStats($ships)
	{
		$count = 0;
		$points = 0;
		if(is_array($ships))
		{
			foreach($ships as $ship)
			{
				$shipData = sqlSelectRow("construction", array("basic_metal", "basic_silicon", "basic_hydrogen"), "", "buildingid = ".sqlVal($ship["id"]));
				$basic_res = $shipData["basic_metal"] + $shipData["basic_silicon"] + $shipData["basic_hydrogen"];
				$points += $basic_res * $ship["quantity"];
				$count += $ship["quantity"];
			}
		}
		return array("points" => round($points * RES_TO_UNIT_POINTS, POINTS_PRECISION), "count" => $count);
	}

	/**
	 * Calculates points of achievements
	 * @param int $user_id User ID
	 *
	 * @return array Points and Count
	 */
	public static function getUserAchievementsStats($user_id)
	{
		$count	= 0;
		$points = 0;
		$result = sqlSelectRow(
			'achievements2user AS a',
			// 'sum(quantity) AS a_quantity, SUM(d.points * a.quantity) AS a_points',
			'SUM(case when a.state > '.ACHIEV_STATE_ALERT.' or quantity = 0 then quantity else quantity-1 end) AS a_quantity,
			 SUM(d.points * (case when a.state > '.ACHIEV_STATE_ALERT.' or quantity = 0 then quantity else quantity-1 end)) AS a_points',
			' LEFT JOIN ' . PREFIX . 'achievement_datasheet AS d ON d.achievement_id = a.achievement_id',
			'user_id = ' . sqlVal($user_id)
		);
		return array("points" => $result['a_points'], "count" =>  $result['a_quantity']);
	}

    public static function updateUserPoints($user_id, $save = true)
    {
        $build_points = $build_count = 0;
        $research_points = $research_count = 0;
        $unit_points = $unit_count = 0;
        $achievement_points = $achievement_count = 0;

        $sub_result = sqlSelect("planet", "planetid", "", "userid = ".sqlVal($user_id));
        while($sub_row = sqlFetch($sub_result))
        {
// echo "before getPlanetBuildingStats\n";

            $stat = PointRenewer::getPlanetBuildingStats($sub_row["planetid"]);
            $build_points += $stat["points"];
            $build_count += $stat["count"];
// echo "after getPlanetBuildingStats, points: ${stat['points']}, count: ${stat['count']}\n";

            $stat = PointRenewer::getPlanetFleetStats($sub_row["planetid"]);
            $unit_points += $stat["points"];
            $unit_count += $stat["count"];
// echo "after getPlanetFleetStats , points: ${stat['points']}, count: ${stat['count']}\n";
        }
        sqlEnd($sub_result);

        $stat = PointRenewer::getUserResearchStats($user_id);
        $research_points += $stat["points"];
        $research_count += $stat["count"];
// echo "after getUserResearchStats, points: ${stat['points']}, count: ${stat['count']}\n";

        $stat = PointRenewer::getUserFleetEventStats($user_id);
        $unit_points += $stat["points"];
        $unit_count += $stat["count"];
// echo "after getUserFleetEventStats, points: ${stat['points']}, count: ${stat['count']}\n";

        $stat = PointRenewer::getUserLotStats($user_id);
        $unit_points += $stat["points"];
        $unit_count += $stat["count"];
// echo "after getUserLotStats, points: ${stat['points']}, count: ${stat['count']}\n";

        $stat = PointRenewer::getUserAchievementsStats($user_id);
        $achievement_points += $stat["points"];
        $achievement_count  += $stat["count"];

        $sum_points = $build_points + $research_points + $unit_points;

        $array_to_add = array(
            "points"	=> $sum_points,
            "b_points"	=> $build_points,
            "r_points"	=> $research_points,
            "u_points"	=> $unit_points,
            "b_count"	=> $build_count,
            "r_count"	=> $research_count,
            "u_count"	=> $unit_count,
            "a_points"	=> $achievement_points,
            "a_count"	=> $achievement_count,
        );
        if($save){
            // sqlUpdate("user", $array_to_add, "userid = ".sqlVal($row["userid"]));
            $update = $array_to_add;
            foreach($update as $field => $value)
            {
                $update[$field] = "$field = GREATEST(0, ".sqlVal($value).")";
            }
            $update[] = updateDmPointsSetSql();
            // Updated By Pk
            Core::getDB()->query('UPDATE ' . PREFIX . 'user SET ' . implode(', ', $update) . ' WHERE userid = ' . sqlVal($user_id));
        }
        return $array_to_add;
    }
}
?>