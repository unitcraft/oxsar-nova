<?php
/**
 * Database access data.
 *
 * Oxsar http://oxsar.ru
 *
 */

if(!defined("RECIPE_ROOT_DIR")) { die("Hacking attempt detected."); }

//### Database access ###//

$database["host"] = "localhost";
$database["port"] = null;
$database["tableprefix"] = "na_";
$database["type"] = "DB_MySQLi";
$database["user"] = "user";
$database["userpw"] = "passwd";
$database["databasename"] = "db";

//### Do not change beyond here ###//

define("PREFIX", $database['tableprefix']);
if(DATABASE_SUBDOMAIN)
{
	$parsedUrl = parseUrl($_SERVER['HTTP_HOST']);
	if(isset($database2subdomain))
	{
		$database['databasename'] = $database2subdomain[$parsedUrl['subdomain']];
	}
	else
	{
		$database['databasename'] = $parsedUrl['subdomain'];
	}
}
?>
