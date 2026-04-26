<?php
/**
* Deletes all sessions from cache folder and disable sessions in database.
* 
* Oxsar http://oxsar.ru
*
* 
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

/**
* @return void
*/
function clearSessions($deletionDays)
{
  // Hook::event("CLEAN_SESSIONS_BEGIN");

  // Delete all contents from session cache
  $sessionCache = APP_ROOT_DIR."cache/sessions/";
  File::rmDirectoryContent($sessionCache);

  // Disable sessions
  $deleteTime = time() - ($deletionDays-1) * 86400;
  sqlUpdate("sessions", array("logged" => 0), "logged = 1 AND time < ".sqlVal($deleteTime));

  // Delete old sessions
  $deleteTime = time() - $deletionDays * 86400;
  sqlDelete("sessions", "time < ".sqlVal($deleteTime));
}

clearSessions(7);

?>
