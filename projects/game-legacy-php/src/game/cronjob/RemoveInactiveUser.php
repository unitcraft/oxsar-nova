<?php
/**
* Deletes inavtive users.
*
* Oxsar http://oxsar.ru
*
* 
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

/**
* @return void
*/
function removeInactiveUsers()
{
  if(defined("OXSAR_RELEASED") && !OXSAR_RELEASED)
  {
    return;
  }

  if(defined("TEST_REMOVE_USERS"))
  {
    $num = 0;
    echo "[removeInactiveUsers]<br>";
  }
  
  sqlQuery("UPDATE ".PREFIX."user SET "
  	. "  last = last+".sqlVal(60*60*24*20)
  	. ", umode = 0 "
  	. " WHERE umode = 1 AND last < ".sqlVal(time() - 60*60*24*40));
  
  $where = "("
  	. " (u.last < ".sqlVal(time() - 60*60*24*30)." OR (u.`delete` > 0 AND u.`delete` < ".sqlVal(time() - 60*60*24*3)."))"
    . " AND u.umode = 0 AND (u2g.usergroupid is null OR (u2g.usergroupid != '2' AND u2g.usergroupid != '4'))"
    . ")";
  $result = sqlSelect("user u", "u.userid", "LEFT JOIN ".PREFIX."user2group u2g ON u2g.userid = u.userid", $where);
  while($row = sqlFetch($result))
  {
    if(defined("TEST_REMOVE_USERS"))
    {
      if(0) // ++$num > 10)
      {
        echo "BREAK <br>";
        return;
      }
      echo "<pre>";
      print_r($row);
      echo "</pre>";
      continue;
    }

    $userid = $row["userid"];
    $_result = sqlSelect("planet", array("planetid", "ismoon"), "", "userid = ".sqlVal($userid));
    while($_row = sqlFetch($_result))
    {
      if(!$_row["ismoon"])
      {
        NS::deletePlanet($_row["planetid"], $userid, false);
      }
    }
    sqlEnd($_result);
    $_result = sqlSelect("alliance", "aid", "", "founder = ".sqlVal($userid));
    if($_row = sqlFetch($_result))
    {
      deleteAlliance($_row["aid"]);
    }
    sqlEnd($_result);
    sqlDelete("user", "userid = ".sqlVal($userid));
    sqlDelete("officer", "userid = ".sqlVal($userid));
  }
  sqlEnd($result);
}

removeInactiveUsers();

?>
