<?php
/**
* Oxsar http://oxsar.ru
*
* 
*/

define("INGAME", true);
//define("EXEC_CRON", true);
define("LOGIN_REQUIRED", true);
define("FORCE_REWRITE_OVERRIDE", false);
define("FORCE_REQUEST_CLEAR", false);

require_once(dirname(__FILE__) . "/global.inc.php");
require_once(APP_ROOT_DIR."game/Functions.inc.php");

if(1) // !isset($_GET['ajax']))
{
  new Core();
  // new NS();
}
else
{
  define("SID", $_GET['sid']);
}

// Path to the chat directory:
define('AJAX_CHAT_PATH', APP_ROOT_DIR . 'chat2/');
define('AJAX_CHAT_URL', '/oxsar/chat2/');
define('AJAX_CHAT_SCRIPT_URL', '/oxsar/chatPro.php?sid=' . SID);

// Include Class libraries:
require(AJAX_CHAT_PATH.'lib/classes.php');

// Initialize the chat:
$ajaxChat = new OxsarAJAXChat();

// /oxsar/chatPro.php?sid=6370e98ce615&ajax=true&lastID=0&getInfos=userID%2CuserName%2CuserRole%2CchannelID%2CchannelName&channelID=0
// /oxsar/chatPro.php?sid=6370e98ce615&ajax=true&lastID=0

?>