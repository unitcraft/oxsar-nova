<?php
	/**
	* Shows combat reports to public.
	*
	* Oxsar http://oxsar.ru
	*
	*
	*/

$config_fileName = 'main.php';
/*
if( strpos($_SERVER['HTTP_HOST'], 'odnoklassniki') !== false )
{
	$config_fileName = 'odnoklassniki.php';
}
else if( strpos($_SERVER['HTTP_HOST'], 'mailru') !== false )
{
	$config_fileName = 'mailru.php';
}
else if( strpos($_SERVER['HTTP_HOST'], 'vkontakte') !== false )
{
	$config_fileName = 'vkontakte.php';
}
*/
require_once( dirname(__FILE__).'/bd_connect_info.php' );
$GLOBALS["RUN_YII"] = 1;
$yii	= dirname(__FILE__).'/yii/framework/yii.php';
$config	= dirname(__FILE__).'/new_game/protected/config/' . $config_fileName;

define("FORCE_REWRITE", false);
define("IE_MAX_STYLES", 24); //IE does not supports more then 30 style sheets
define("LOGIN_PAGE", true);

require_once($yii);
Yii::createWebApplication($config)->run();
Yii::app()->end();