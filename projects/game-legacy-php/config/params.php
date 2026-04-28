<?php

if(!empty($_SERVER['DOCUMENT_URI']))
{
	$_SERVER['PHP_SELF'] = $_SERVER['DOCUMENT_URI'];
}

require_once(dirname(__FILE__).'/../../../global.inc.php');

function gs_a()
{
	$sufix = '';
	if( strpos($_SERVER['HTTP_HOST'], 'odnoklassniki') !== false )
	{
		$sufix .= '&' . http_build_query($GLOBALS['ODNOKLASSNIKI'], '', '&');
	}
	return $sufix;
}

function gs_n()
{
	$sufix = '';
	if( strpos($_SERVER['HTTP_HOST'], 'odnoklassniki') !== false )
	{
		$sufix .= http_build_query($GLOBALS['ODNOKLASSNIKI'], '', '&');
	}
	return $sufix;
}

function gs_q()
{
	$sufix = '';
	if( strpos($_SERVER['HTTP_HOST'], 'odnoklassniki') !== false )
	{
		$sufix .= '?' . http_build_query($GLOBALS['ODNOKLASSNIKI'], '', '&');
	}
	return $sufix;
}

function t($msg, $level = 'myTrace')
{
//	return;
	ob_start();
	print_r($msg);
	$content_grabbed=ob_get_contents();
	ob_end_clean();
	if( $level != 'myTrace' )
	{
		$level = 'myTrace.' . $level;
	}
	Yii::trace($content_grabbed, $level);
}

function artImageUrl($action, $suffix, $real_url = true)
{
	// return YII_GAME_DIR.'/index.php?r=artefact2user_YII/image&id='.$art_id;
	return ($real_url ? RELATIVE_URL : "")."index.php/artefact2user_YII/{$action}?{$suffix}";
}

function socialUrl($url)
{
	if( defined('SN') )
	{
		$suf = Yii::app()->socialAPI->getSuffix();
		if( $suf && strpos($url, $suf) === false )
		{
			return addUrlParams($url, $suf);
		}
	}
	return $url;
}

function addUrlParams($url, $params)
{
	if(!$params)
	{
		return $url;
	}
	return $url . (strpos($url, '?') === false ? '?' : '&') . $params;
}

return array(
	'adminEmail'	=> 'support@oxsar.ru',
	'galery_link'	=> 'http://cakeuniverse.ru/?app=gallery&module=gallery&controller=browse&album=28',
	'forum_link'	=> 'http://cakeuniverse.ru/index.php',
	// 'game_link'		=> '/game.php',
	'login_link'	=> '/login.php',
	'mobi_link'		=> 'http://oxsar.mobi',
	'dm_mobi_link'		=> 'http://dm.oxsar.mobi',
	'dm_game_link'		=> 'http://dm.oxsar.ru',
	'game_link'		=> 'http://oxsar.ru',
	'withManualConnect'			=> true,
	'time_assault_public'		=> 60*60*24*3,
	'time_assault_public_end'	=> 60*60*24*6,
	'technicalWorks'	=> require(dirname(dirname(dirname(dirname(__FILE__)))) . DIRECTORY_SEPARATOR . 'technicalWorks.php'),
	'exchange_fuel_mult' => 1, // 0.01,
	'galaxy_distance_mult' => 20000,

	// game config
	'pagetitle' => 'Мир Oxsar©'.UNIVERSE_NAME,
	'universe_name_full' => 'Мир Oxsar©'.UNIVERSE_NAME,
	'maxloginattempts' => 5,
	'bannedlogintime' => 5,
	'mailaddress' => 'support@oxsar.ru',
	'standardlanggroups' => 'global',
	'timezone' => 'Europe/Moscow',
	'defaultlanguage' => DEF_LANGUAGE_ID,
	'templatepackage' => 'standard/',
	'templateextension' => '.tpl',
	'maintemplate' => 'layout',
	'guestgroupid' => 1,
	'userselect' => array(
		'fieldsnames' => array(
			0 => 'u2a.aid',
		),
		'indexnames' => array(
			0 => 'aid',
		),
	),
	'userjoins' => 'LEFT JOIN na_user2ally u2a ON (u2a.userid = u.userid)',
	'ATTACKING_STOPPAGE' => 0,
	'DEFAULT_HYDROGEN' => 0,
	'DEFAULT_METAL' => 500,
	'DEFAULT_SILICON' => 500,
	'DEFAULT_START_HYDROGEN' => 0,
	'DEFAULT_START_METAL' => 1000,
	'DEFAULT_START_SILICON' => 500,
	'DEL_MESSAGE_DAYS' => 30,
	'EMAIL_ACTIVATION_CHANGED_EMAIL' => 0,
	'EMAIL_ACTIVATION_CHANGED_PASSWORD' => 0,
	'EMAIL_ACTIVATION_DISABLED' => 0,
	'FORUM_URL' => 'http://dm.oxsar.org/forums/', // index.php?showforum=33',
	// 'GALAXYS' => 8, => NUM_GALAXYS
	// 'SYSTEMS' => 600, => NUM_SYSTEMS
	// 'GAMESPEED' => 0.75, => GAMESPEED
	'HELP_PAGE_URL' => 'http://dm.oxsar.org/forums/index.php?showforum=34',
	'HOME_PLANET_SIZE' => 18800,
	'HYDROGEN_BASIC_PROD' => 0,
	'MAX_ALLIANCE_TEXT_LENGTH' => 20000,
	'MAX_APPLICATION_TEXT_LENGTH' => 3000,
	'MAX_ASTEROID_SIZE' => 100,
	'MAX_CHARS_ALLY_NAME' => 30,
	'MAX_CHARS_ALLY_TAG' => 8,
	'MAX_INPUT_LENGTH' => 255,
	// 'MAX_PLANETS' => 10,
	'MAX_PMS' => 10,
	'MAX_PM_LENGTH' => 1000,
	'MAX_USER_CHARS' => 30,
	'METAL_BASIC_PROD' => 20,
	'MIN_CHARS_ALLY_NAME' => 3,
	'MIN_CHARS_ALLY_TAG' => 2,
	'MIN_USER_CHARS' => 3,
	'NEWBIE_PROTECTION_MAX_POINTS' => 100000,
	'NEWBIE_PROTECTION_MID_PERCENT' => 20,
	'NEWBIE_PROTECTION_MID_POINTS' => 10000,
	'NEWBIE_PROTECTION_PERCENT' => 10,
	'NEWBIE_PROTECTION_START_PERCENT' => 30,
	'NEWBIE_PROTECTION_START_POINTS' => 3000,
	'PLANET_FIELD_ADDITION' => 10,
	'PLANET_IMG_EXT' => '.jpg',
	'PRODUCTION_FACTOR' => 1.5,
	'SHIPYARD_ORDER_ABORT_PERCENT' => 70,
	'SILICON_BASIC_PROD' => 10,
	'STAR_SURVEILLANCE_CONSUMPTION' => 5000,
	'USER_PER_PAGE' => 100,
	'WATING_TIME_REGISTRATION' => 5,
);