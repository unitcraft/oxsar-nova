<?php
/**
* Oxsar http://oxsar.ru
*
*
*/

require_once(dirname(__FILE__) . "/../global.inc.php");

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

// Legacy payment: $database-массив для PS_mysqlConnect()/mysql_*-функций.
// Конфиг берём из констант (src/bd_connect_info.php). Сам платёжный код
// использует устаревшие mysql_*-функции и не работает в PHP 7+ (отключён
// до плана 37.6+); массив оставлен для совместимости сигнатур.
$database = array(
	"host"          => DB_HOST,
	"user"          => DB_USER,
	"userpw"        => DB_PWD,
	"databasename"  => DB_NAME,
	"tableprefix"   => DB_PREFIX,
	"charset"       => DB_CHAR,
);

define("PS_GAME_DOMAIN_PARAM_NAME", "shp_zgame");

function PS_gameDomain($check_request = true, $def_domain = "oxsar.ru")
{
	$game_domain = null;
    if($check_request){
        if( !empty($_POST[PS_GAME_DOMAIN_PARAM_NAME]) )
        {
            $game_domain = trim($_POST[PS_GAME_DOMAIN_PARAM_NAME]);
        }
        else if( !empty($_GET[PS_GAME_DOMAIN_PARAM_NAME]) )
        {
            $game_domain = trim($_GET[PS_GAME_DOMAIN_PARAM_NAME]);
        }
    }
    if(!$game_domain){
        if( !empty($_SERVER["HTTP_HOST"]) )
        {
            $game_domain = trim($_SERVER["HTTP_HOST"]);
        }
        else if( !empty($_SERVER["SERVER_NAME"]) )
        {
            $game_domain = trim($_SERVER["SERVER_NAME"]);
        }
    }
	if( !$game_domain || !preg_match("#^(www\.|dm\.)?(oxsar|netassault)\.(ru|com|net|org|mobi)$#is", $game_domain) )
	{
		$game_domain = $def_domain;
	}
	return $game_domain;
}

define("PS_GAME_DOMAIN", PS_gameDomain());
define("PS_GAME_RETURN_URL", "http://".PS_GAME_DOMAIN."/game.php");

define("PS_GAME_DOMAIN_PAIR", PS_GAME_DOMAIN_PARAM_NAME."=".urlencode(PS_GAME_DOMAIN));

function PS_addGameDomain($url)
{
	return $url . (strstr($url, "?") ? "&" : "?") . PS_GAME_DOMAIN_PAIR;
}

function PS_mysqlConnect()
{
	global $database;
	$conn = mysql_connect($database["host"], $database["user"], $database["userpw"]);
	if(!mysql_select_db($database["databasename"])) // or die("not connect to db");
	{
		PS_errorPage("db connect error");
		exit();
	}
	// mysql_query('SET CHARSET '.$database["charset"]);
	mysql_query('SET NAMES '.PS_sqlVal($database['charset']));
	return $conn;
}

function PS_sqlVal($data)
{
	return is_null($data) ? "NULL" : "'".mysql_real_escape_string((string)$data)."'";
}

function PS_sqlArray()
{
	$data = func_num_args() > 1 ? func_get_args() : func_get_arg(0);
	if(is_array($data))
	{
		$result = array();
		foreach($data as $value)
		{
			$result[] = PS_sqlVal($value);
		}
		return count($result) ? implode(",", $result) : 'NULL';
	}
	return PS_sqlVal($data);
}

function PS_quoteName($name)
{
	return "`".$name."`";
}

function PS_quoteNames()
{
	$data = func_num_args() > 1 ? func_get_args() : func_get_arg(0);
	return "`".implode("`, `", (array)$data)."`";
}

function PS_sqlInsert($table_name, $values)
{
	$sql = "INSERT INTO ".PREFIX."$table_name (".PS_quoteNames(array_keys($values))
		.") VALUES (".PS_sqlArray(array_values($values)).")";
	mysql_query($sql);
	return mysql_insert_id();
}

function PS_sqlUpdate($table_name, $values, $where)
{
	$setValues = array();
	foreach($values as $key => $value)
	{
		$setValues[] = PS_quoteName($key)."=".PS_sqlVal($value);
	}
	$sql = "UPDATE ".PREFIX."$table_name SET ".implode(", ", $setValues)." ".$where;
	mysql_query($sql);
}

function PS_getUser($user_id)
{
	$res = mysql_query("select * from ".PREFIX."user where userid=".PS_sqlVal((int)$user_id));
	$user = mysql_fetch_assoc($res);
	mysql_free_result($res);
	if(empty($user))
	{
		PS_errorPage("db connect error");
		exit();
	}
	return $user;
}

function PS_registerPayment($params)
{
	$user_id = (int)$params["user_id"];
	$pay_type = $params["pay_type"];

	if(isset($params["amount"]))
	{
		$pay_amount = (float)$params["amount"];
		$credit = (float)$params["credit"];

		if($pay_amount <= 0 || $credit <= 0)
		{
			PS_errorPage("register payment error: $pay_amount <= 0 || $credit <= 0");
			exit();
		}
	}
	else
	{
		$pay_amount = null;
		$credit = 0; // (float)$params["credit"];
	}

	mysql_query("INSERT into ".PREFIX."payments set pay_user_id=".PS_sqlVal($user_id)
		.", pay_type=".PS_sqlVal($pay_type)
		.", pay_from='', pay_amount=".PS_sqlVal($pay_amount)
		.", pay_credit=".PS_sqlVal($credit)
		.", pay_domain=".PS_sqlVal(PS_gameDomain(null))
		.", pay_date=NOW(), pay_status=0");
	$pay_id = mysql_insert_id();
	// mysql_close();
	if(empty($pay_id))
	{
		PS_errorPage("register payment error");
		exit();
	}
	return $pay_id;
}

define("PS_PAYMENT_OK",            0);
define("PS_PAYMENT_ALREADY_DONE",  1);
define("PS_PAYMENT_ERR_USER",      2);

$GLOBALS['PS_PAYMENT_RESULT_ID'] = 0;
function PS_finishPayment($params)
{
  $GLOBALS['PS_PAYMENT_RESULT_ID'] = 0;

  $user_id = max(0, (int)$params["user_id"]);
  $credit = max(0, (float)$params["credit"]);

  $pay_id = $params["pay_id"] === "new" ? "new" : intval(max(0, $params["pay_id"]));

  $pay_type = $params["pay_type"];
  $pay_from = $params["pay_from"];
  $amount = $params["amount"];
  $extra_info = (isset($params["extra_info"])?$params["extra_info"]:'');
  // $transaction_id = (isset($params["transaction_id"])?$params["transaction_id"]:NULL);

  $is_in_progress = true;
  if($pay_id !== "new")
  {
    $res = mysql_query("SELECT pay_id, pay_user_id FROM ".PREFIX."payments WHERE pay_id=".PS_sqlVal($pay_id)." AND pay_status=0");
	$row = mysql_fetch_array($res);
    $is_in_progress = !empty($row) ? true : false; // mysql_num_rows($res) > 0;
	$user_id = !empty($row) ? $row["pay_user_id"] : $user_id;
    mysql_free_result($res);
  }

  if($is_in_progress)
  {
    $res = mysql_query("SELECT userid FROM ".PREFIX."user WHERE userid=".PS_sqlVal($user_id));
    $is_user_found = mysql_num_rows($res) > 0;
    mysql_free_result($res);

    if($is_user_found)
    {
      if($pay_id !== "new")
      {
        mysql_query("UPDATE ".PREFIX."payments SET pay_user_id=".PS_sqlVal($user_id).", pay_credit=".PS_sqlVal($credit)
          . ", pay_type=".PS_sqlVal($pay_type).", pay_from=".PS_sqlVal($pay_from).", pay_amount_r=".PS_sqlVal($amount)
          .", pay_extra_info=".PS_sqlVal($extra_info)
          // .", pay_ext_transaction=".PS_sqlVal($transaction_id)
          . ", pay_date=NOW(), pay_status=1 "
          . " WHERE pay_id=".PS_sqlVal($pay_id));
      }
      else
      {
        mysql_query("INSERT INTO ".PREFIX."payments SET pay_user_id=".PS_sqlVal($user_id).", pay_credit=".PS_sqlVal($credit)
          . ", pay_type=".PS_sqlVal($pay_type).", pay_from=".PS_sqlVal($pay_from).", pay_amount_r=".PS_sqlVal($amount)
          .", pay_extra_info=".PS_sqlVal($extra_info)
          // .", pay_ext_transaction=".PS_sqlVal($transaction_id)
          . ", pay_date=NOW(), pay_status=1 ");
        $pay_id = mysql_insert_id();
      }

      mysql_query("UPDATE ".PREFIX."user SET credit=credit+".PS_sqlVal($credit)." WHERE userid=".PS_sqlVal($user_id));

      $res = mysql_query("SELECT credit, languageid FROM ".PREFIX."user WHERE userid = ".PS_sqlVal($user_id));
      $row = mysql_fetch_array($res);
      mysql_free_result($res);

	  PS_sqlInsert("message", array(
		"mode" => MSG_FOLDER_CREDIT,
		"time" => time(),
		"sender" => null,
		"receiver" => $user_id,
		"message" => strtr(PS_langStr("MSG_CREDIT_BUY", $row["languageid"]), array("{@credits}" => PS_fNumber($credit, 2, $row["languageid"]))),
		"subject" => PS_langStr("MSG_CREDIT", $row["languageid"]),
		"readed" => 0
		));

      PS_sqlInsert("res_log", array(
        "type" => RES_UPDATE_BUY_CREDITS,
        "userid" => $user_id,
        "credit" => $credit,
        "result_credit" => $row["credit"],
        "game_credit" => PS_gameCredit(),
        "ownerid" => $pay_id,
        ));

      PS_addCreditBonusItem($user_id, $credit, $pay_id);

      $GLOBALS['PS_PAYMENT_RESULT_ID'] = $pay_id;

      mysql_close();
      return PS_PAYMENT_OK;
    }
    mysql_close();
    return PS_PAYMENT_ERR_USER;
  }
  mysql_close();
  return PS_PAYMENT_ALREADY_DONE;
}

function PS_getCreditBonusInfo()
{
	$day = date("j");
	$month = date("n");
	$credit_scale = 1.5;
	$items = array(
		array(
			"unitid" => ARTEFACT_IGLA_MORI,
			"credit" => 5900 * $credit_scale
		),
		array(
			"unitid" => ARTEFACT_ROBOT_CONTROL_SYSTEM,
			"credit" => 13900 * $credit_scale
		),
		array(
			"unitid" => ARTEFACT_ANNIHILATION_ENGINE_10,
			"credit" => 19000 * $credit_scale
		),
		array(
			"unitid" => ARTEFACT_BATTLE_ATTACK_POWER_10,
			"credit" => 8900 * $credit_scale
		),
		array(
			"unitid" => ARTEFACT_BATTLE_SHIELD_POWER_10,
			"credit" => 8900 * $credit_scale
		),
		array(
			"unitid" => ARTEFACT_BATTLE_SHELL_POWER_10,
			"credit" => 9000 * $credit_scale
		),
		array(
			"unitid" => ARTEFACT_SUPERCOMPUTER,
			"credit" => 8900 * $credit_scale
		),
		array(
			"unitid" => ARTEFACT_ALLY_IGN,
			"credit" => 59000 * $credit_scale
		),
		array(
			"unitid" => ARTEFACT_MERCHANTS_MARK,
			"credit" => 4900 * $credit_scale
		),
		array(
			"unitid" => ARTEFACT_ANNIHILATION_ENGINE,
			"credit" => 4900 * $credit_scale
		),
		array(
			"unitid" => NANOBOT_REPAIR_SYSTEM,
			"credit" => 14900 * $credit_scale
		),
 	);
	$i = $day + $month*31;
	$each_day = 10;
	if(($i % $each_day) != 0){
		return array(
			"unitid" => 0,
			"credit" => 0
		);
	}
	$i = ($i / $each_day) % count($items);
	return $items[$i];
}

function PS_addCreditBonusItem($user_id, $credit, $pay_id)
{
	$item = PS_getCreditBonusInfo();
	if(!empty($item["credit"]) && !empty($item["unitid"]) && $credit >= $item["credit"]){
		mysql_query("INSERT into ".PREFIX."credit_bonus_item set unitid=".PS_sqlVal($item["unitid"])
			.", userid=".PS_sqlVal($user_id)
			.", credit=".PS_sqlVal($credit)
			.", date=NOW(), done=0");
	}
}

function PS_gameCredit()
{
    $res = mysql_query("SELECT sum(credit) as game_credit FROM ".PREFIX."user");
    $row = mysql_fetch_array($res);
    mysql_free_result($res);
    return $row["game_credit"];
}

function PS_langStr($str_name, $lang_id = DEF_LANGUAGE_ID)
{
	static $lang_cache = array();
	if(!isset($lang_cache[$str_name]))
	{
		$res = mysql_query("SELECT content FROM ".PREFIX."phrases WHERE languageid = ".PS_sqlVal($lang_id)." AND title = ".PS_sqlVal($str_name)." LIMIT 1");
		$row = mysql_fetch_array($res);
		mysql_free_result($res);
		$lang_cache[$str_name] = $row["content"];
	}
	return $lang_cache[$str_name];
}

/**
* Prettify number: 1000 => 1.000
*
* @param integer	The number being formatted
* @param integer	Sets the number of decimal points
*
* @return string	Number
*/
function PS_fNumber($number, $decimals = 2, $lang_id = DEF_LANGUAGE_ID)
{
  return number_format($number, $decimals, PS_langStr("DECIMAL_POINT", $lang_id), PS_langStr("THOUSANDS_SEPERATOR", $lang_id));
}

function PS_errorPage($message = "")
{
	?>
<html>
<head>
<META HTTP-EQUIV=REFRESH CONTENT="5;URL='<?php echo PS_GAME_RETURN_URL; ?>'/">
<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
</head>
<body>
<center><font color="red">Ошибка платежа</font>
<p />
Вы будуте перемещены на главную страницу через 5 секунд.
<p />
<a href='<?php echo PS_GAME_RETURN_URL; ?>'>Вы можете перейти по этой ссылке, если не хотите ждать.</a>
<?php if(!empty($message)){ echo "<p />".$message; } ?>
</center>
<?php if(0) { ?>
<pre>
<?php var_dump(PS_gameDomain(null)); echo "\n"; print_r($GLOBALS); ?>
</pre>
<?php } ?>
</body>
</html>
	<?php
    exit();
}

function PS_okPage($message = "")
{
	?>
<html>
<head>
<META HTTP-EQUIV=REFRESH CONTENT="5;URL='<?php echo PS_GAME_RETURN_URL; ?>'/">
<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
</head>
<body>
<center><font color="green">Платеж успешно завершен</font>
<p />
Вы будуте перемещены на главную страницу через 5 секунд.
<p />
<a href='<?php echo PS_GAME_RETURN_URL; ?>'>Вы можете перейти по этой ссылке, если не хотите ждать.</a>
<?php if(!empty($message)){ echo "<p />".$message; } ?>
</center>
</body>
</html>
	<?php
    exit();
}

function PS_proxyToUrl( $url , $params = array() )
{
	$options = array(
		CURLOPT_RETURNTRANSFER => true,
	);

	if(empty($params)){
		$params = array(
			'get' => true,
			'post' => true,
		);
	}
	if(!empty($params['get']) && !empty($_GET)){
		$options[CURLOPT_HTTPGET] = true;
		$url .= (strstr($url, "?") ? '&' : '?').http_build_query($_GET, '', '&');
	}
	if(!empty($params['post']) && !empty($_POST)){
		unset($options[CURLOPT_HTTPGET]);
		$options[CURLOPT_POST] = true;
		$options[CURLOPT_POSTFIELDS] = http_build_query($_POST, '', '&');
	}
	$options[CURLOPT_URL] = $url;

	/*
	echo "<pre>";
	print_r($_GET);
	print_r($_POST);
	print_r($options);
	echo "</pre>";
	*/

	$ch = curl_init();
	curl_setopt_array($ch, $options);
	$res = curl_exec($ch);
	curl_close($ch);

	return $res;
}

?>