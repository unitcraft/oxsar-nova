<?php
/**
* Removes destroyed planets.
*
* Oxsar http://oxsar.ru
*
* 
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

/**
* @return void
*/
function removeGalaxyGarbage()
{
  $result = sqlSelect("galaxy g", array("g.planetid", "e.eventid as eventid", "e2.eventid as eventid2"), 
    "LEFT JOIN ".PREFIX."events e ON (e.destination = g.planetid)"
    . " LEFT JOIN ".PREFIX."events e2 ON (e2.planetid = g.planetid)", 
    "g.destroyed = '1' AND e.destination is null AND e2.planetid is null");
  while($row = sqlFetch($result))
  {
    if(empty($row["eventid"]))
    {
      $id = $row["planetid"];
      if(defined("TEST_REMOVE"))
      {
        echo "[removeGalaxyGarbage()] planet id: $id <br>";
        echo "<pre>";
        print_r($row);
        echo "</pre>";
        continue;
      }
      sqlDelete("planet", "planetid = ".sqlVal($id));
    }
  }
  sqlEnd($result);
}

removeGalaxyGarbage();

?>
