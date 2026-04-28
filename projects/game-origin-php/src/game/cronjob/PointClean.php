<?php
/**
* Cleans points.
*
* Oxsar http://oxsar.ru
*
* 
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

function cleanPoints()
{
  $debug = defined("CRONJOB_DEBUG") ? CRONJOB_DEBUG : false;

  // Hook::event("CLEAN_POINTS_BEGIN");
  $result = sqlSelect("user", "userid");
  while($row = sqlFetch($result))
  {
    $build_points = $build_count = 0;
    $research_points = $research_count = 0;
    $unit_points = $unit_count = 0;

    $sub_result = sqlSelect("planet", "planetid", "", "userid = ".sqlVal($row["userid"]));
    while($sub_row = sqlFetch($sub_result))
    {
      $stat = PointRenewer::getPlanetBuildingStats($sub_row["planetid"]);
      $build_points += $stat["points"];
      $build_count += $stat["count"];

      $stat = PointRenewer::getPlanetFleetStats($sub_row["planetid"]);
      $unit_points += $stat["points"];
      $unit_count += $stat["count"];
    }
    sqlEnd($sub_result);

    $stat = PointRenewer::getUserResearchStats($row["userid"]);
    $research_points += $stat["points"];
    $research_count += $stat["count"];

    $stat = PointRenewer::getUserFleetEventStats($row["userid"]);
    $unit_points += $stat["points"];
    $unit_count += $stat["count"];

    $sum_points = $build_points + $research_points + $unit_points;

    sqlUpdate("user", array(
      "points" => $sum_points,
      "b_points" => $build_points,
      "r_points" => $research_points,
      "u_points" => $unit_points,
      "b_count" => $build_count,
      "r_count" => $research_count,
      "u_count" => $unit_count,
      ), "userid = ".sqlVal($row["userid"]));

    if($debug)
    {
      debug_var(array(
        "points" => $sum_points,
        "b_points" => $build_points,
        "r_points" => $research_points,
        "u_points" => $unit_points,
        "b_count" => $build_count,
        "r_count" => $research_count,
        "u_count" => $unit_count,
        ), "[$debug] USERID: ".$row["userid"]);

      if(--$debug <= 0)
      {
        break;
      }
    }
  }
  sqlEnd($result);
}

cleanPoints();

?>
