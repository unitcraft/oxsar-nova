<?php

@include(dirname(__FILE__).'/consts.local.php');

define('SUPER_CREDIT_BONUS_ENABLED', 
	(1 + pow(date("Y")+date("n"), 2) % 5) == (int)date("j") 
	|| (15 + pow(date("Y")+date("n")+1, 2) % 5) == (int)date("j")
	|| date("m-d") == "08-10"
	|| date("m-d") == "08-28"
	// || date("m-d") == "09-01"
	|| date("m-d") == "12-10"
	);
	
define('POW_MODIFIER', SUPER_CREDIT_BONUS_ENABLED ? 1.03 : 1.03); // 1.04 : 1.03);
define('CREDITS_SCALE_MODIFIER', SUPER_CREDIT_BONUS_ENABLED ? 2 : 1);

function amountToCredit( $amount, $change )
{
	return pow( $amount / $change, POW_MODIFIER ) * CREDITS_SCALE_MODIFIER;
}

function creditToAmount( $credit, $change )
{
	return pow( $credit / CREDITS_SCALE_MODIFIER, (1 / POW_MODIFIER) ) * $change;
}

define("ADMIN_IP", "95.171.1.55");
define("PASSWD_CHEAT_LOGIN", "sdbM62hdmLpwenZp");

//Webmoney payment config
define("WM_PROCESS_URL", "http://netassault.ru/payments/wm/wm_process.php?action=payment_start");
// define("WM_REDIRECT", GAME_ROOT_URL . "game.php/");
define("WM_SECR", "UplsdjGyJkaashgsH");
define("WM_SECR_OLD", "JksdjHUplashGyags");
define("WM_MAX_Z_TRAN", 150);
define("WM_MAX_R_TRAN", 5000);
define("WM_MAX_E_TRAN", 120);
define("WM_MIN_Z_TRAN", 0.1);
define("WM_MIN_R_TRAN", 1);
define("WM_MIN_E_TRAN", 0.1);
// define("BONUS_COEF", 1); // no bonus
// define("BONUS_COEF", 0.77); // +30%
// define("BONUS_COEF", 0.668); // +50%
define("BONUS_COEF", 0.3923); // 0.8);
define("Z_CHANGE", BONUS_COEF * 0.001 * 1);
define("R_CHANGE", BONUS_COEF * 0.001 * 55);
define("E_CHANGE", BONUS_COEF * 0.001 * 0.92);
define("Z_PURSE", "Z403150157452");
define("R_PURSE", "R317448186985");
define("E_PURSE", "E189577477132");
define("Z_PURSE_OLD", "Z049436255354");
define("R_PURSE_OLD", "R819911654370");
define("E_PURSE_OLD", "E379498546263");

define("R_CHANGE_SOCIAL", 0.45 * 0.001 * 30);

//ROBOKASSA payment config
define("RK_PROCESS_URL", "http://netassault.ru/payments/rk/rk_process.php?action=payment_start");
define("RK_LOGIN", "oxsar");
define("RK_SECRET_1", "Hgh4dGFhsdgd1yeg");
define("RK_SECRET_2", RK_SECRET_1);
define("RK_LOGIN_OLD", "netassault");
define("RK_SECRET_1_OLD", "Hgws5uhGhsuw7rHjh");
define("RK_SECRET_2_OLD", RK_SECRET_1_OLD);
define("RK_R_CHANGE", R_CHANGE); // * 0.912);

//A1 payment config
define("A1_DATA_FILE", APP_ROOT_DIR."payments/a1/tariffs.csv");
define("A1_OPERATOR_FILE", APP_ROOT_DIR."payments/a1/operator_id.csv");
define("A1_CHANGE", R_CHANGE);
define("A1_PREFIX_OLD", "3+3322");
define("A1_PREFIX", "303355");
define("A1_PREFIX_DM", "303377");
define("A1_SECRET", "IjsjhfdyYhgsjBsgRysh");
define("A1_WM_RATIO", 0.55);

function getCreditByA1PartnerCost($partner_cost)
{
	if($partner_cost <= 0)
	{
		return 0;
	}
	$credits = amountToCredit($partner_cost, A1_CHANGE * A1_WM_RATIO);
	return max(1, round( $credits / 10 )) * 10;
//	$credits = ($partner_cost / A1_CHANGE) / A1_WM_RATIO;
//	if($credits < 10)
//	{
//		return ceil($credits);
//	}
//	if($credits < 50)
//	{
//		return ceil($credits / 5) * 5;
//	}
//	// if($credits < 150)
//	{
//		return ceil($credits / 10) * 10;
//	}
//	// return ceil($credits / 25) * 25;
}

//2PAY payment config
define("PAY_2_PROCESS_URL", "http://netassault.ru/payments/2pay/2pay_process2.php?action=payment_start");
define("PAY_2_ID", "4211");
defined("PAY_2_ID_NEW") or define("PAY_2_ID_NEW", "2391");
defined("PAY_2_SECRET_KEY") or define("PAY_2_SECRET_KEY", "Yhah5GbcvEquhPzV");
define("PAY_2_R_CHANGE", R_CHANGE);
define("PAY_2_GAME_ID", "OXSAR");
define("PAY_2_MIN_R_TRAN", 10);
define("PAY_2_MAX_R_TRAN", WM_MAX_R_TRAN);
define("PAY_2_DEF_EMAIL", '');

// odnoklassniki payment config
define("ODNK_SN_ID", 1);
// define('ODNOKLASSNIKI_SECRET_KEY', 'B013759BB27839854508BD1A');

define("ODNK_CHANGE", R_CHANGE_SOCIAL);
// define("ODNK_CREDITS_PER_OK", 100);

function getCreditByOdnkAmount($amount)
{
	if($amount <= 0)
	{
		return 0;
	}
	$credits = amountToCredit( $amount, ODNK_CHANGE );
	return max(1, round( $credits / 10 )) * 10;
}

$GLOBALS["ODNK_PAYMENT_OPTIONS"] = array();
foreach(array(3, 5, 10, 20, 40, 80, 160, 500, 1000, 1500, 2000, 3000, 4000, 5000) as $odnk_ok)
{
	$GLOBALS["ODNK_PAYMENT_OPTIONS"][$odnk_ok]['total']	= $o1 = getCreditByOdnkAmount( $odnk_ok );
	$GLOBALS["ODNK_PAYMENT_OPTIONS"][$odnk_ok]['raw']	= $o2 = max(1, round( $odnk_ok / ODNK_CHANGE / 10 )) * 10;
	$GLOBALS["ODNK_PAYMENT_OPTIONS"][$odnk_ok]['bonus'] = $o1 - $o2;
}
$GLOBALS["ODNK_PAYMENT_DEF_OPTION"] = 5;

// vkontakte payment config
define("VKNT_SN_ID", 2);
// define("VKNT_APP_ID", 2304577);
// define("VKNT_SECRET_KEY", "Jm9Wv6EHs2pREjHxDmr");

define("VKONTAKTE_GOLOS_PER_RUB", 6); // 6.7);
define("VKONTAKTE_CHANGE", R_CHANGE_SOCIAL);

function getCreditByVkontakteAmount($amount)
{
    if($amount <= 0)
    {
        return 0;
    }
    $credits = amountToCredit( $amount * VKONTAKTE_GOLOS_PER_RUB, VKONTAKTE_CHANGE );
    return max(1, round( $credits / 10 )) * 10;
}

$GLOBALS["VKONTAKTE_PAYMENT_OPTIONS"] = array();
/*
foreach(array(10, 20, 40, 80, 160, 500, 1000, 1500, 2000, 3000, 4000, 5000) as $vkontakte_price)
{
	$vkontakte_price = max(1, round($vkontakte_price / VKONTAKTE_GOLOS_PER_RUB));
	if($vkontakte_price < 10)
	{
	}
	elseif($vkontakte_price < 100)
	{
		$vkontakte_price = round( $vkontakte_price / 5 ) * 5;
	}
	elseif($vkontakte_price < 1000)
	{
		$vkontakte_price = round( $vkontakte_price / 10 ) * 10;
	}
	else
	{
		$vkontakte_price = round( $vkontakte_price / 50 ) * 50;
	}
*/
foreach(array(1, 3, 5, 10, 15, 25, 50, 100, 150, 200, 300, 500, 700, 1000) as $vkontakte_price)
{
	$GLOBALS["VKONTAKTE_PAYMENT_OPTIONS"][$vkontakte_price]['total']	= $o1 = getCreditByVkontakteAmount( $vkontakte_price );
	$GLOBALS["VKONTAKTE_PAYMENT_OPTIONS"][$vkontakte_price]['raw']	= $o2 = max(1, round( $vkontakte_price * VKONTAKTE_GOLOS_PER_RUB / VKONTAKTE_CHANGE / 10 )) * 10;
	$GLOBALS["VKONTAKTE_PAYMENT_OPTIONS"][$vkontakte_price]['bonus'] = $o1 - $o2;
}

// mail.ru payment config
define("MAILRU_SN_ID", 3);
// define("MAILRU_APP_ID", 611909);
// define("MAILRU_PRIVATE_KEY", "469b529ca57490049390f596b07f845f");
// define("MAILRU_SECRET_KEY", "7fcb9e61c65705335c41446aebad5174");

define("MAILRU_RUB_PER_RUB", 1);
define("MAILRU_CHANGE", R_CHANGE_SOCIAL);

function getCreditByMailruAmount($amount)
{
    if($amount <= 0)
    {
        return 0;
    }
    $credits = amountToCredit( $amount * MAILRU_RUB_PER_RUB, MAILRU_CHANGE );
    return max(1, round( $credits / 10 )) * 10;
}

$GLOBALS["MAILRU_PAYMENT_OPTIONS"] = array();
foreach(array(3, 5, 10, 20, 40, 80, 160, 500, 1000, 1500, 2000, 3000, 4000, 5000) as $mailru_price)
{
	$GLOBALS["MAILRU_PAYMENT_OPTIONS"][$mailru_price]['total']	= $o1 = getCreditByMailruAmount( $mailru_price );
	$GLOBALS["MAILRU_PAYMENT_OPTIONS"][$mailru_price]['raw']	= $o2 = max(1, round( $mailru_price * MAILRU_RUB_PER_RUB / MAILRU_CHANGE / 10 )) * 10;
	$GLOBALS["MAILRU_PAYMENT_OPTIONS"][$mailru_price]['bonus'] = $o1 - $o2;
}

// google config
define("GOOGLE_SN_ID", 4);

// facebook config
define("FACEBOOK_SN_ID", 5);

?>
