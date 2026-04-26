<?php
/**
* Oxsar http://oxsar.ru
*
*
*/

/* if(isset($_SERVER['REMOTE_ADDR'])){
    foreach(array(
        // П�?ПЕЦЦЦ, пипец...
        // '188.114.21.', '109.165.48.', '31.23.74.', '46.61.100.', '93.178.92.',
        '178.150.156.', // 2012-03-22 21:52
        '92.242.76.', // 2012-03-23 17:02
        '128.73.18.', // 2012-03-23 23:33
        ) as $ip_start){
        if($ip_start && strpos($_SERVER['REMOTE_ADDR'], $ip_start) === 0){
            exit;
        }
    }
} */

if(defined("MICROTIME")){
	return;
}

// Required constants (DO NOT MODIFY!)
define("MICROTIME", microtime());
define("TIME", time());

function __get_request_dir__()
{
  $script_dir = !empty($_SERVER["SERVER_NAME"]) ? substr(dirname($_SERVER["SCRIPT_NAME"]), 1) : "";
  if(!defined("USE_FACEBOOK_SKIN"))
  {
    define("USE_FACEBOOK_SKIN", intval($script_dir == "oxsar-fb"));
  }
  return $script_dir ? $script_dir."/" : "";
}

// Required constants (Just for advanced)
// APP_ROOT_DIR — корень src/ (где лежит этот файл)
define("APP_ROOT_DIR",    str_replace("\\", "/", dirname(__FILE__))."/");
// GAME_ORIGIN_DIR — корень всего проекта game-origin/
define("GAME_ORIGIN_DIR", str_replace("\\", "/", dirname(__FILE__, 2))."/");
define("RECIPE_ROOT_DIR", APP_ROOT_DIR."core/");
define("CRONJOB_DIR",     APP_ROOT_DIR."ext/cronjobs/");
define("MAINTAIN_DIR",    RECIPE_ROOT_DIR."maintenance/");
define("HTTP_HOST", "http://".(!empty($_SERVER["HTTP_HOST"]) ? $_SERVER["HTTP_HOST"] : (!empty($_SERVER["SERVER_NAME"]) ? $_SERVER["SERVER_NAME"] : "oxsar.ru"))."/");
define("REQUEST_DIR", __get_request_dir__());
define("FULL_URL", HTTP_HOST.REQUEST_DIR);
define("LOCAL_URL", "/".REQUEST_DIR);
define("RELATIVE_URL", LOCAL_URL);
define("BASE_FULL_URL", HTTP_HOST);
define("BASE_DM_FULL_URL", HTTP_HOST);
define("GAME_ROOT_URL", HTTP_HOST . (defined("USE_FACEBOOK_SKIN") && USE_FACEBOOK_SKIN ? "oxsar-fb/" : ""));
define("IPADDRESS", isset($_SERVER["REMOTE_ADDR"]) ? $_SERVER["REMOTE_ADDR"] : "0.0.0.0");
define("ERROR_REPORTING_TYPE", E_ALL & ~E_NOTICE);
if(1 || $_SERVER['REMOTE_ADDR'] != '95.171.1.55')
{
	defined("FORCE_REWRITE") or define("FORCE_REWRITE", defined("FORCE_REWRITE_OVERRIDE") ? FORCE_REWRITE_OVERRIDE : true);
}
else
{
	defined("FORCE_REWRITE") or define("FORCE_REWRITE", false);
}
define("DATABASE_SUBDOMAIN", false);
define("AUTOLOAD_PATH_CORE", "/,template/,database/,http/,maintenance/");
define("AUTOLOAD_PATH_APP", "game/,game/page/,game/models/");
define("AUTOLOAD_PATH_APP_EXT", "ext/,ext/page/,ext/models/");
define("REQUEST_LEVEL_NAMES", "go,id,1,2,3,4,5");
// YII_GAME_DIR убран — Yii не используется

// Required constants (Global preferences)
// define("COOKIE_PREFIX", "oxsar_" . preg_replace("#[^\w\d]+#is", "_", OXSAR_VERSION)); // Prefix for cookies.
define("COOKIE_PREFIX", "oxsar-"); // Prefix for cookies.
define("CACHE_ACTIVE", true); 		// Global switch to enable/disable cache funcion.
define("GZIP_ACITVATED", true);		// Enables GZIP compression.
define("COOKIE_SESSION", false);	// Session will be stored in cookies.
define("URL_SESSION", true);		// Session will be committed via URL.
defined('IPCHECK') or define('IPCHECK', false);			// Enables IP check for sessions.
define("LOGIN_REQUIRED", true);		// If false, access can be handled by permissions.
define("EXEC_CRON", true); // OXSAR_RELEASED);			// Enables cron jobs.

define("MC_SERVER", 'localhost'); // memcache server
define("MC_PORT", 11211);         // memcache port

// Загружаем игровые константы из config/
// GAME_UNIVERSE определяет какой конфиг вселенной использовать (dm, dominator, niro, ...)
$_universe = getenv('GAME_UNIVERSE') ?: 'dm';
require_once(GAME_ORIGIN_DIR."config/consts.php");
@include_once(GAME_ORIGIN_DIR."config/consts." . $_universe . ".local.php");
unset($_universe);

@include_once(APP_ROOT_DIR."ext/global.inc.php");

/* if(!empty($_GET["taram"]))
{
    echo "<pre>";
    print_r($GLOBALS);
    echo "</pre>";
    exit;
} */

if(LOGIN_REQUIRED)
{
  // Логин через auth-service portal (plan-36), не собственная страница
  define("LOGIN_URL", (getenv('PORTAL_URL') ?: '') . "/login");
}
require_once(RECIPE_ROOT_DIR."init.php");

// JWT authentication (plan-36 auth-service)
// Должен подключаться после init.php, т.к. использует Core::getDB()
if (!defined("LOGIN_PAGE")) {
    session_start();
    require_once(RECIPE_ROOT_DIR."JwtAuth.php");
    define('AUTH_JWKS_URL', getenv('AUTH_JWKS_URL') ?: '');
    JwtAuth::authenticate();
}

