<?php
$config_fileName = 'main.php';
/*
if( strpos($_SERVER['HTTP_HOST'], 'odnoklassniki') !== false )
{
	$config_fileName = 'odnoklassniki.php';
}
*/
require_once('bd_connect_info.php');
$yii	= dirname(__FILE__).'/yii/framework/yii.php';
$config_fileName = 'main.php';
/*
$config_fileName = 'chat.php';
if( strpos($_SERVER['HTTP_HOST'], 'odnoklassniki') !== false )
{
	$config_fileName = 'odnoklassniki_' . $config_fileName;
}
else if( strpos($_SERVER['HTTP_HOST'], 'mailru') !== false )
{
	$config_fileName = 'mailru_' . $config_fileName;
}
else if( strpos($_SERVER['HTTP_HOST'], 'vkontakte') !== false )
{
	$config_fileName = 'vkontakte_' . $config_fileName;
}
*/
$config	= dirname(__FILE__).'/new_game/protected/config/' . $config_fileName;
require_once($yii);
Yii::createWebApplication($config)->run();
Yii::app()->end();