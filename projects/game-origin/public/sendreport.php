<?php
/**
* Sends combat reports.
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
include_once( dirname(__FILE__).'/bd_connect_info.php' );
$GLOBALS["RUN_YII"] = 1;
$yii	= dirname(__FILE__).'/yii/framework/yii.php';
$config	= dirname(__FILE__).'/new_game/protected/config/' . $config_fileName;

define("FORCE_REWRITE", false);
define("LOGIN_PAGE", true);

require_once($yii);
Yii::createWebApplication($config)->run();
Yii::app()->end();