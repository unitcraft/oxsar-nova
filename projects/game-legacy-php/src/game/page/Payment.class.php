<?php
/**
* Payment module.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("RECIPE_ROOT_DIR")) { die("Hacking attempt detected."); }

require_once(APP_ROOT_DIR . "payments/a1/Functions.inc.php");
require_once(APP_ROOT_DIR . "ext/payment.inc.php");

class Payment extends Page
{
	public function __construct()
	{
	// it's hask for a while, will be removed soon
	/* if(NS::getUser()->get("userid") != 2)
	{
		doHeaderRedirection("game.php");
	} */

		parent::__construct();

        if(isDeathmatchEnd()){
            Logger::dieMessage('DEATHMATCH_END_MESSAGE');
        }

		$check_credit_available = false;
		$check_min_credit = SUPER_CREDIT_BONUS_ENABLED ? 500000 : 450000;
		$check_max_credit = SUPER_CREDIT_BONUS_ENABLED ? 700000 : 500000;
		if( $check_credit_available )
		{
			$other_user = sqlSelectRow("res_log l", array("u.username", "l.userid", "sum(l.credit) as credit"),
				"JOIN ".PREFIX."user u on u.userid=l.userid",
				"`type`=".sqlVal(RES_UPDATE_BUY_CREDITS)." AND `time` >= SUBDATE(now(), INTERVAL 5 DAY) AND l.userid != ".sqlUser()
				. " GROUP BY l.userid ORDER BY sum(l.credit) DESC LIMIT 1");

			$cur_user = sqlSelectRow("res_log l", array("l.userid", "sum(l.credit) as credit"),
				"",
				"`type`=".sqlVal(RES_UPDATE_BUY_CREDITS)." AND `time` >= SUBDATE(now(), INTERVAL 5 DAY) AND l.userid = ".sqlUser()
				. " GROUP BY l.userid"
				);

			$max_credit_available = clampVal($other_user["credit"] * 2, $check_min_credit, $check_max_credit);
		}
		if(!$check_credit_available || empty($other_user) || $cur_user["credit"] < $max_credit_available)
		{
			$this->setPostAction("Webmoney", "paymentWebmoney")
				->setPostAction("ROBOKASSA", "paymentRobokassa")
				->setPostAction("SMS", "paymentA1")
				->setPostAction("SMS2", "paymentA1step2")
				->addPostArg("paymentA1step2", "country")
				->setPostAction("The2pay1", "payment2pay_step1")
				->setPostAction("The2pay2", "payment2pay_step2")
				->addPostArg("payment2pay_step2", "amount")
				->setGetAction('go', 'PaymentVkontakte', 'paymentVkontakte')
				->addGetArg('paymentVkontakte', 'amount')
				->setGetAction('go', 'PaymentWebmoney', 'paymentWebmoney')
				->setGetAction('go', 'PaymentRobokassa', 'paymentRobokassa')
				->setGetAction('go', 'PaymentSMS', 'paymentA1')
				;
		}
		else
		{
			Core::getTPL()->assign("tooManyBought", 1);
		}
		if(!empty($other_user) && defined("ADMIN_IP") && IPADDRESS == ADMIN_IP)
		{
			Core::getTPL()->assign("is_admin", 1);
			Core::getTPL()->assign("max_credit_available", fNumber($max_credit_available));
			Core::getTPL()->assign("other_bought_credit", fNumber($other_user["credit"]));
			Core::getTPL()->assign("other_username", $other_user["username"]);
			Core::getTPL()->assign("cur_bought_credit", fNumber($cur_user["credit"]));
			Core::getTPL()->assign("cur_username", Core::getUser()->get("username"));
		}

		Core::getTPL()->assign("user_id", Core::getUser()->get("userid"));
		Core::getTPL()->assign("z_max", WM_MAX_Z_TRAN);
		Core::getTPL()->assign("r_max", WM_MAX_R_TRAN);
		Core::getTPL()->assign("e_max", WM_MAX_E_TRAN);
		Core::getTPL()->assign("z_min", WM_MIN_Z_TRAN);
		Core::getTPL()->assign("r_min", WM_MIN_R_TRAN);
		Core::getTPL()->assign("e_min", WM_MIN_E_TRAN);
		Core::getTPL()->assign("z_change", Z_CHANGE);
		Core::getTPL()->assign("r_change", R_CHANGE);
		Core::getTPL()->assign("e_change", E_CHANGE);
		Core::getTPL()->assign("wm_process_url", WM_PROCESS_URL);
		Core::getTPL()->assign("rk_process_url", RK_PROCESS_URL);
		Core::getTPL()->assign("pay_2_process_url", PAY_2_PROCESS_URL);
		Core::getTPL()->assign("sid", SID);
		Core::getTPL()->assign("rootModifier", 1 / POW_MODIFIER);
		Core::getTPL()->assign("powModifier", POW_MODIFIER);
		$this->proceedRequest();
	}

	protected function index()
	{
//		if( defined('SN') && $_SERVER['REMOTE_ADDR'] != '95.171.1.55' )
//		{
//			Core::getTPL()->assign("text", 'Сервис временно недоступен. Приносим свои извенения.');
//			Core::getTPL()->display("blank");
//			return $this;
//
//			doHeaderRedirection( socialUrl(RELATIVE_URL . "game.php") );
//		}

		Core::getLanguage()->load('info,Payment');
		if( !defined('SN') )
		{
            $email = NS::getUser()->get('email');
            $is_email = strpos($email, '@') !== false;
            Core::getTPL()->assign("pay2_iframe_url", "https://secure.xsolla.com/paystation/?".http_build_query(array(
                'projectid' => PAY_2_ID_NEW,
                'id_theme' => 34, // 16,
                'local' => 'ru',
                'v1' => NS::getUser()->get('userid'),
                'v2' => NS::getUser()->get('username').' ('.($is_email ? $email : 'email').') '.PS_GAME_DOMAIN,
                'v3' => 0, // pay id
                'email' => $is_email ? $email : PAY_2_DEF_EMAIL,
                // PS_GAME_DOMAIN_PARAM_NAME => PS_GAME_DOMAIN,
            )));

			if(isAdmin())
			{
				Core::getTPL()->assign("isWebmoneyOpen", true);
				Core::getTPL()->assign("isA1Open", false); // true);
				Core::getTPL()->assign("isRoboKassaOpen", true);
                Core::getTPL()->assign("isPay2iFrameOpen", true);
				Core::getTPL()->assign("isPay2Open", true);
				Core::getTPL()->assign("isOdnoklassnikiOpen", false);
				Core::getTPL()->assign("isVkontakteOpen", false);
				Core::getTPL()->assign("isMailruOpen", false);
			}
			else
			{
				Core::getTPL()->assign("isWebmoneyOpen", true);
				Core::getTPL()->assign("isA1Open", false); // true);
				Core::getTPL()->assign("isRoboKassaOpen", true);
                Core::getTPL()->assign("isPay2iFrameOpen", true);
				Core::getTPL()->assign("isPay2Open", true);
				Core::getTPL()->assign("isOdnoklassnikiOpen", false);
				Core::getTPL()->assign("isVkontakteOpen", false);
				Core::getTPL()->assign("isMailruOpen", false);
			}
			
			if($socialAPI->isExternal)
			{
				Core::getTPL()->assign("payAction", $socialAPI->internalUrl);
				if($socialAPI instanceof SocialAPI_OAuth2_Odnoklassniki)
				{
					Core::getTPL()->assign("isOAuth2OdnoklassnikiOpen", true);
				}
				else if($socialAPI instanceof SocialAPI_OAuth2_Mailru)
				{
					Core::getTPL()->assign("isOAuth2MailruOpen", true);
				}
				else if($socialAPI instanceof SocialAPI_OAuth2_Vkontakte)
				{
					Core::getTPL()->assign("isOAuth2VkontakteOpen", true);
				}
			}
		}
		else if( 0 && defined('SN_FULLSCREEN') )
		{
			header("Location: " . BASE_FULL_URL); exit();
			// doHeaderRedirection(BASE_FULL_URL, false);
		}
		else if(SN == MAILRU_SN_ID)
		{
			Core::getTPL()->assign("isWebmoneyOpen", false);
			Core::getTPL()->assign("isA1Open", false);
			Core::getTPL()->assign("isRoboKassaOpen", false);
			Core::getTPL()->assign("isPay2Open", false);
			Core::getTPL()->assign("isOdnoklassnikiOpen", false);
			Core::getTPL()->assign("isVkontakteOpen", false);
			Core::getTPL()->assign("isMailruOpen", false); // true);

			Core::getTPL()->assign("ok_title", Core::getLanguage()->getItem('SN_OK_TITLE'));
			Core::getTPL()->assign("ok_desc", Core::getLanguage()->getItem('SN_OK_DESC'));

			Core::getTPL()->assign("ok_code", 0);
			Core::getTPL()->assign("ok_amount", 1);

			$options = array();
			$options_new = array();
			foreach( $GLOBALS['MAILRU_PAYMENT_OPTIONS'] as $price => $credits )
			{
				$options[$price] = array(
					'name'			=> Core::getLanguage()->getItemWith('SN_OK_CREDITS', array("credit" => $credits['total'], "ok" => $price)),
					'price' 		=> $price,
					'code'			=> $price,
				);
				$options_new[$price] = array_merge($options[$price], array(
					'description'	=> Core::getLanguage()->getItem('SN_OK_DESC'),
					'total' 		=> $credits['total'],
					'raw'			=> $credits['raw'],
					'bonus' 		=> $credits['bonus'],
					'price2show' 	=> fNumber($price),
					'title'			=> Core::getLanguage()->getItemWith(
						'SN_OK_CREDITS_STRING',
						array("total" => fNumber($credits['total']), "raw" => fNumber($credits['raw']), 'bonus' => fNumber($credits['bonus']))
					),));
			}
			// Core::getTPL()->assign("odnoklassnikiOptions", json_encode($options));
			Core::getTPL()->addLoop("mailru_credits", $options_new);
		}
		else if(SN == VKNT_SN_ID)
		{
			Core::getTPL()->assign("isWebmoneyOpen", false);
			Core::getTPL()->assign("isA1Open", false);
			Core::getTPL()->assign("isRoboKassaOpen", false);
			Core::getTPL()->assign("isPay2Open", false);
			Core::getTPL()->assign("isOdnoklassnikiOpen", false);
			Core::getTPL()->assign("isVkontakteOpen", false); // true);
			Core::getTPL()->assign("isMailruOpen", false);

			Core::getTPL()->assign("ok_title", Core::getLanguage()->getItem('SN_OK_TITLE'));
			Core::getTPL()->assign("ok_desc", Core::getLanguage()->getItem('SN_OK_DESC'));

			Core::getTPL()->assign("ok_code", 0);
			Core::getTPL()->assign("ok_amount", 1);

			$options = array();
			$options_new = array();
			foreach( $GLOBALS['VKONTAKTE_PAYMENT_OPTIONS'] as $price => $credits )
			{
				$options[$price] = array(
					'name'			=> Core::getLanguage()->getItemWith('SN_OK_CREDITS', array("credit" => $credits['total'], "ok" => $price)),
					'price' 		=> $price,
					'code'			=> $price,
				);
				$options_new[$price] = array_merge($options[$price], array(
					'description'	=> Core::getLanguage()->getItem('SN_OK_DESC'),
					'total' 		=> $credits['total'],
					'raw'			=> $credits['raw'],
					'bonus' 		=> $credits['bonus'],
					'price2show' 	=> fNumber($price),
					'title'			=> Core::getLanguage()->getItemWith(
						'SN_OK_CREDITS_STRING',
						array("total" => fNumber($credits['total']), "raw" => fNumber($credits['raw']), 'bonus' => fNumber($credits['bonus']))
					),));
			}
			// Core::getTPL()->assign("odnoklassnikiOptions", json_encode($options));
			Core::getTPL()->addLoop("vkontakte_credits", $options_new);
		}
		else if(SN == ODNK_SN_ID)
		{
			Core::getTPL()->assign("isWebmoneyOpen", false);
			Core::getTPL()->assign("isA1Open", false);
			Core::getTPL()->assign("isRoboKassaOpen", false);
			Core::getTPL()->assign("isPay2Open", false);
			Core::getTPL()->assign("isOdnoklassnikiOpen", false); // true);

			Core::getTPL()->assign("ok_title", Core::getLanguage()->getItem('SN_OK_TITLE'));
			Core::getTPL()->assign("ok_desc", Core::getLanguage()->getItem('SN_OK_DESC'));

			Core::getTPL()->assign("ok_code", 0);
			Core::getTPL()->assign("ok_amount", 1);

			Core::getTPL()->assign("fapi_url", $_GET['api_server'] . 'js/fapi.js');
			Core::getTPL()->assign("api_server", $_GET["api_server"]);
			Core::getTPL()->assign("apiconnection", $_GET["apiconnection"]);

			$options = array();
			$options_new = array();
			foreach( $GLOBALS['ODNK_PAYMENT_OPTIONS'] as $oks => $credits )
			{
				$options[$oks] = array(
					'name'			=> Core::getLanguage()->getItemWith('SN_OK_CREDITS', array("credit" => $credits['total'], "ok" => $oks)),
					'price' 		=> $oks,
					'code'			=> $oks,
				);
				$options_new[$oks] = array_merge($options[$oks], array(
					'description'	=> Core::getLanguage()->getItem('SN_OK_DESC'),
					'total' 		=> $credits['total'],
					'raw'			=> $credits['raw'],
					'bonus' 		=> $credits['bonus'],
					'price2show' 	=> fNumber($oks),
					'title'			=> Core::getLanguage()->getItemWith(
						'SN_OK_CREDITS_STRING',
						array("total" => fNumber($credits['total']), "raw" => fNumber($credits['raw']), 'bonus' => fNumber($credits['bonus']))
					),));
			}
			Core::getTPL()->assign("odnoklassnikiOptions", json_encode($options));
			Core::getTPL()->addLoop("odnoklassniki_credits", $options_new);
			// t($options_new);
		}
		Core::getTPL()->display("payment");
		return $this;
	}

	protected function paymentVkontakte($amount)
	{
		// Платежи через VK отключены (план 37 — соцсети не используются).
		doHeaderRedirection("game.php/Payment", false);
		return $this;
	}

	protected function generateCreditStringForOdnoklassniki($data)
	{
		return ;
	}

	/*
	protected function calcSignature( $data )
	{
		$sig = array();
		foreach ($data as $key => $val)
		{
			if ($key != 'sig')
			{
				$sig[] = "$key=$val";
			}
		}
		sort($sig);
		return md5( join('', $sig) . ODNOKLASSNIKI_SECRET_KEY );
	}
	*/

	protected function paymentWebmoney()
	{
		Core::getTPL()->display("paymentWebmoney");
		return $this;
	}

	protected function paymentRobokassa()
	{
		Core::getTPL()->assign("r_change", RK_R_CHANGE);
		Core::getTPL()->display("paymentRobokassa");
		return $this;
	}

	protected function checkPaymentA1Lock()
	{
		/*
		$user_id = Core::getUser()->get("userid");
		if(!in_array($user_id, array(-2, 6)))
		{
		Core::getTPL()->assign("isPaymentLocked", true);
		}
		*/
	}

	protected function paymentA1()
	{
		$this->checkPaymentA1Lock();

		$this->checkA1csv();

		$data = file_get_contents(APP_ROOT_DIR."cache/a1data.ser");
		$data = unserialize($data);
		foreach($data as $key=>$value){
			$countries[]['name'] = $key;
		}
		Core::getTPL()->addLoop("countries", $countries);
		Core::getTPL()->display("paymentA1");
		return $this;
	}

	protected function paymentA1step2($country)
	{
		$this->checkPaymentA1Lock();

		$this->checkA1csv();

		$data = file_get_contents(APP_ROOT_DIR."cache/a1data.ser");
		$data = unserialize($data);
		if(count($data[$country])>0){
			$i = 0;
			foreach($data[$country] as $value)
			{
				/* if(in_array($value["number"], array("4124", "3649")))
				{
					continue;
				} */

				$sms_data[$i] = $value;
				$sms_data[$i]['credit'] = getCreditByA1PartnerCost($value['partnerCost']);
				/*foreach ($value['operatorsprice'] as $operator_id=>$cena_nomera)
				{
				$nomer_data[$value]['operatorsprice'][$operator_id] = $cena_nomera;
				$nomer_data[$value]['operatorsnames'][$operator_id] = $data[$country][$value]['operatorsnames'][$operator_id];
				} */

				$i++;
			}
			// Core::getTPL()->addHTMLHeaderFile("lib/jquery.js?".CLIENT_VERSION, "js");
//			Core::getTPL()->addHTMLHeaderFile("lib/jquery-ui.js?".CLIENT_VERSION, "js");
//			Core::getTPL()->addHTMLHeaderFile("ui-darkness/jquery-ui.css?".CLIENT_VERSION, "css");
			Core::getTPL()->assign("country", $country);
			Core::getTPL()->assign("prefix", UNIVERSE_NAME == 'Dominator' ? A1_PREFIX_DM : A1_PREFIX);
			Core::getTPL()->addLoop("sms_data", $sms_data);

			Core::getTPL()->display("paymentA1step2");
		} else {
			foreach($data as $key=>$value){
				$countries[]['name'] = $key;
			}
			Core::getTPL()->addLoop("countries", $countries);
			Core::getTPL()->display("paymentA1");
		}
		return $this;
	}

	protected function checkA1csv()
	{
		if(!file_exists(APP_ROOT_DIR."cache/a1data.ser"))
		{
			$this->createA1ser();
		}
		else if(date("mdY", filectime(APP_ROOT_DIR."cache/a1data.ser")) != date("mdY"))
		{
			$this->createA1ser();
		}
		//Core::getTPL()->display("paymentWebmoney");
		return null;
	}

	protected function createA1ser()
	{
		$data = file(A1_DATA_FILE);
		//echo(count($data));
		$tmp = array();
		foreach($data as $dat){
			$dat = trim($dat);
			// $dat = iconv("Windows-1251", "UTF-8", $dat);
			$dat = iconv('cp1251', 'utf-8//IGNORE', $dat);
			$dat = explode(";", $dat);
			if(!empty($dat[0]) && empty($dat[2])){
				$tmp[$dat[0]] = '';
				$curr_country = $dat[0];
				$i = 0;
			}
			if(!empty($dat[0]) && !empty($dat[2]) && is_numeric($dat[0])){

				$tmp[$curr_country][$dat[0]]['number'] = $dat[0];
				//$tmp[$curr_country][$dat[0]]['operatorID'] = $dat[1];
				//$tmp[$curr_country][$dat[0]]['operator'] = $dat[2];
				$tmp[$curr_country][$dat[0]]["operatorsprice"][$dat[1]] = $dat[5];
				$tmp[$curr_country][$dat[0]]["operatorsnames"][$dat[1]] = $dat[2];
				$dat[3] = str_replace(",", ".", $dat[3]);
				if(empty($tmp[$curr_country][$dat[0]]['partnerCost']))
				{
					$tmp[$curr_country][$dat[0]]['partnerCostMin'] = $dat[3];
					$tmp[$curr_country][$dat[0]]['partnerCostMax'] = $dat[3];
				}
				else
				{
					$tmp[$curr_country][$dat[0]]['partnerCostMin'] = min($tmp[$curr_country][$dat[0]]['partnerCostMin'], $dat[3]);
					$tmp[$curr_country][$dat[0]]['partnerCostMax'] = max($tmp[$curr_country][$dat[0]]['partnerCostMax'], $dat[3]);
				}
				// $tmp[$curr_country][$dat[0]]['partnerCost'] = $tmp[$curr_country][$dat[0]]['partnerCostMin'];
				// $tmp[$curr_country][$dat[0]]['partnerCost'] = $tmp[$curr_country][$dat[0]]['partnerCostMax'];
				$tmp[$curr_country][$dat[0]]['partnerCost'] = ceil(($tmp[$curr_country][$dat[0]]['partnerCostMin'] + $tmp[$curr_country][$dat[0]]['partnerCostMax']) / 2);
				$tmp[$curr_country][$dat[0]]['partnerVal'] = $dat[4];
				$dat[5] = str_replace(",", ".", $dat[5]);
				if(empty($tmp[$curr_country][$dat[0]]['smsCostMin'])) $tmp[$curr_country][$dat[0]]['smsCostMin'] = $dat[5];
				if(!empty($tmp[$curr_country][$dat[0]]['smsCostMin']) && $tmp[$curr_country][$dat[0]]['smsCostMin']>$dat[5]) $tmp[$curr_country][$dat[0]]['smsCostMin'] = $dat[5];
				if(empty($tmp[$curr_country][$dat[0]]['smsCostMax'])) $tmp[$curr_country][$dat[0]]['smsCostMax'] = $dat[5];
				if(!empty($tmp[$curr_country][$dat[0]]['smsCostMax']) && $tmp[$curr_country][$dat[0]]['smsCostMax']<$dat[5]) $tmp[$curr_country][$dat[0]]['smsCostMax'] = $dat[5];
				//$tmp[$curr_country][$dat[0]]['smsCost'] = $dat[5];
				$tmp[$curr_country][$dat[0]]['smsVal'] = $dat[6];
				//$i++;
			}
		}
		file_put_contents(APP_ROOT_DIR."cache/a1data.ser", serialize($tmp)); // , "w+");
	}

	protected function payment2pay_step1()
	{
		Core::getTPL()->assign("r_max", PAY_2_MAX_R_TRAN);
		Core::getTPL()->assign("r_min", PAY_2_MIN_R_TRAN);
		Core::getTPL()->display("payment2pay_step1");
		return $this;
	}

	protected function payment2pay_step2($amount)
	{
		$to_pay = max(WM_MIN_R_TRAN, (int)$amount * PAY_2_R_CHANGE);
		$credit = round($to_pay / PAY_2_R_CHANGE);
		$to_pay = round($to_pay, 2);

		$pay_type = "2pay RUR";

		if(isset($to_pay))
		{
			$user_id = Core::getUser()->get("userid");
			$pay_id = sqlInsert("payments", array(
				"pay_user_id" => $user_id,
				"pay_type" => $pay_type,
				"pay_from" => '',
				"pay_amount" => $to_pay,
				"pay_credit" => $credit,
				"pay_date" => date("Y-m-d H:i:s"),
				"pay_status" => 0));
			// План 86 audit (billing-critical): $pay_id уходит в URL
			// 2pay-сервиса, и при коллбэке кредиты начисляются по
			// этому id. false → 0 → 2pay вернётся с pay_id=0, и
			// callback handler начислит кредиты на чужую запись или
			// проигнорирует. Не маскируем — рвём явно.
			if ($pay_id === false) {
				Logger::dieMessage('DB_ERROR_PAYMENT');
			}

			Header("Location: https://2pay.ru/oplata/?id=".PAY_2_ID."&v1=".PAY_2_GAME_ID.$pay_id."&v2=".Core::getUser()->get("userid")."&v3=".PAY_2_GAME_ID."&amount=".$to_pay);

			//Core::getTPL()->display("payment2pay_step2");
			return $this;
		}
		else
		{
			Core::getTPL()->assign("r_max", PAY_2_MAX_R_TRAN);
			Core::getTPL()->assign("r_min", PAY_2_MIN_R_TRAN);
			Core::getTPL()->display("payment2pay_step1");
			return $this;
		}
	}
}
?>