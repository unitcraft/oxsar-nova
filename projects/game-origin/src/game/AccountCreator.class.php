<?php
/**
* Create accounts.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

require_once(APP_ROOT_DIR."game/Functions.inc.php");
require_once(RECIPE_ROOT_DIR."AjaxRequestHelper.abstract_class.php");

class AccountCreator extends AjaxRequestHelper
{
	/**
	* Requested username.
	*
	* @var string
	*/
	protected $username = "";

	/**
	* Entered password.
	*
	* @var string
	*/
	protected $password = "";

	/**
	* Entered email address.
	*
	* @var string
	*/
	protected $email = "";

	/**
	* Language id.
	*
	* @var integer
	*/
	protected $lang = 0;

	/**
	* Activation key.
	*
	* @var string
	*/
	protected $activation = "";

	protected $agreed = false;

	/**
	* Length of the generated activation key.
	*
	* @var integer
	*/
	const ACTIVATION_KEY_LENGTH = 8;

	protected $errors = array();

	/**
	* Starts account factory.
	*
	* @param string	Username
	* @param string	Password
	* @param string	Email address
	* @param integer	Language id
	*
	* @return void
	*/
	public function __construct($username, $password, $email, $lang, $agreed = true)
	{
		$this->agreed = ($agreed != false);
		$this->username = trim($username);
		$this->password = $password;
		$this->email = $email;
		$this->lang = (!is_numeric($lang)) ? Core::getOptions()->defaultlanguage : $lang;
		Core::getLang()->load('Registration');
		return;
	}

	public function registerUser()
	{
		if( !$this->checkIt() )
		{
			return false;
		}
		if ( !$this->sendMail() )
		{
			return false;
		}
		if( !$this->create() )
		{
			return false;
		};
		return true;
	}

	/**
	* Checks the entered data for validation.
	*
	* @return AccountCreator
	*/
	protected function checkIt()
	{
		$error = array();
		$checkTime = time() - Core::getOptions()->get("WATING_TIME_REGISTRATION") * 60;
		$result = sqlSelect("registration", array("time"), "", "ipaddress = ".sqlVal(IPADDRESS)." AND time >= ".sqlVal($checkTime));
		if( $row = sqlFetch($result) )
		{
			$minutes = ceil(($row["time"] - $checkTime) / 60);
			Core::getLang()->assign("minutes", $minutes);
			$error[] = "REGISTRATION_BANNED_FOR_IP";
		}
		sqlEnd($result);
		if( !checkCharacters($this->username) )
		{
			$error[] = "USERNAME_INVALID";
		}
		if(!checkEmail($this->email))
		{
			$error[] = "EMAIL_INVALID";
		}
		if(Str::length($this->password) < Core::getOptions()->get("MIN_USER_CHARS") || Str::length($this->password) > Core::getOptions()->get("MAX_USER_CHARS"))
		{
			$error[] = "PASSWORD_INVALID";
		}
		$row = sqlSelectRow(
			"user",
			array("username", "email"),
			"",
			"username = ".sqlVal($this->username)." OR email = ".sqlVal($this->email)
		);
		if($row)
		{
			if(Str::compare($this->username, $row["username"]))
			{
				$error[] = "USERNAME_EXISTS";
			}
			if(Str::compare($this->email, $row["email"]))
			{
				$error[] = "EMAIL_EXISTS";
			}
		}

		foreach( array("ukr.net") as $email_domain )
		{
			if( preg_match("#".preg_quote("@".$email_domain, "#")."#is", $this->email) )
			{
				Core::getLang()->assign("email_domain_failed", "@".$email_domain);
				$error[] = "EMAIL_DOMAIN_FAILED";
				break;
			}
		}

		$result_count = sqlSelectField("languages", "count(*)", "", "languageid = ".sqlVal($this->lang));
		if($result_count <= 0)
		{
			$error[] = "UNKOWN_LANGUAGE";
		}
		if(!($this->agreed))
		{
			$error[] = "REG_NO_AGREEMENT";
		}
		if(count($error) > 0)
		{
			$this->errors = $error;
			return false;
//			$this->printIt($error);
		}
		return true;
	}

	/**
	* Sends email with activation key.
	*
	* @return AccountCreator
	*/
	protected function sendMail()
	{
		if( !empty($this->errors) )
		{
			return false;
//			return $this;
		}
		if(!Core::getConfig()->get("EMAIL_ACTIVATION_DISABLED"))
		{
			$this->activation = randString(self::ACTIVATION_KEY_LENGTH);
			$url = HTTP_HOST.REQUEST_DIR."signup.php/Activation:".$this->activation;
			Core::getLang()->assign("regUsername", $this->username);
			Core::getLang()->assign("regPassword", $this->password);
			Core::getLang()->assign("activationLink", $url);
			$message = Core::getLanguage()->getItem("REGISTRATION_MAIL");
			$mail = new Email($this->email, Core::getLanguage()->getItem("REGISTRATION"), $message);
			$mail->sendMail();
		}
		return true;
//		return $this;
	}

	/**
	* Creates the user account.
	*
	* @return AccountCreator
	*/
	protected function create()
	{
		if( !empty($this->errors) )
		{
			return false;
//			return $this;
		}
		$userid = sqlInsert(
			"user",
			array(
				"username"	 => $this->username,
				"email"		 => $this->email,
				"temp_email" => $this->email,
				"languageid" => $this->lang,
				"activation" => $this->activation,
				"regtime"	 => time(),
				"last"		 => time() - 60*60*24*25,
                "observer"   => NEW_USER_OBSERVER,
                "protection_time" => !NEW_USER_OBSERVER && PROTECTION_PERIOD > 0 ? time() + PROTECTION_PERIOD : 0,
			)
		);
		sqlInsert("password", array("userid" => $userid, "password" => md5($this->password), "time" => time()));

		if( 1 ) // IPADDRESS == "95.221.100.214")
		{
            /* if(!NS::getUser()->get("observer") && PROTECTION_PERIOD > 0){
                NS::getUser()->set("protection_time", time() + PROTECTION_PERIOD);
            } */
		}
		else
		{
			$planet = new PlanetCreator($userid, null, null, null, 0, 99);
			$planetid = $planet->getPlanetId();

			// Update By Pk
			sqlUpdate("user", array("curplanet" => $planetid, "hp" => $planetid),
				"userid = ".sqlVal($userid)
				// . ' ORDER BY userid'
				);
		}
		if( $cookie = Core::getRequest()->getCOOKIE("ref") )
		{
			$last_time = time() - 86400;
			$ip_ref = IPADDRESS;
			$ip_adress = sqlSelectRow("registration", array("ipaddress"), "", "ipaddress = ".sqlVal($ip_ref)." AND time > ".sqlVal($last_time));
			if( !empty($ip_adress) && false )
			{
				sqlInsert(
					"message",
					array(
						"mode"		=> 1,
						"time"		=> time(),
						"sender"	=> null,
						"receiver" 	=> $cookie,
						"message"	=> "По Вашей ссылке зарегистрировался пользователь - <b>".
							$this->username.
							"</b>!<p/>Но он не защитан Вашим рефералом, так как система посчитала это накруткой регистраций.",
						"subject"	=> "Регистрация по Вашей ссылке",
						"readed"	=> 0,
					)
				);
			}
			else
			{
				sqlInsert("referral", array("userid" => $cookie, "ref_id" => $userid, "ref_time" => time(), "ref_ip" => IPADDRESS));
				sqlInsert(
					"message",
					array(
						"mode"		=> 1,
						"time"		=> time(),
						"sender"	=> null,
						"receiver"	=> $cookie,
						"message"	=> "У Вас появился новый реферал - <b>".$this->username."</b>!"
										. "<p/>Подробная информация о бонусах за рефералов находится на странице Рефералы.",
						"subject"	=> "Новый реферал!",
						"readed"	=> 0,
					)
				);
			}
		}
		sqlInsert("registration", array("time" => time(), "ipaddress" => IPADDRESS, "useragent" => $_SERVER['HTTP_USER_AGENT']));

        if(SEND_NEW_USER_MESSAGE){
            // Send start-up message
            sqlInsert(
                "message",
                array(
                    "mode"		=> 1,
                    "time"		=> time(),
                    "sender"	=> null,
                    "receiver"	=> $userid,
                    "message"	=> Core::getLang()->getItem("START_UP_MESSAGE"),
                    "subject"	=> Core::getLang()->getItem("START_UP_MESSAGE_SUBJECT"),
                    "readed"	=> 0,
                )
            );
        }
		// Delete Registrations older than 7 days
		sqlDelete("registration", "time < ".sqlVal(time() - 604800));
		$this->errors = "SUCCESS_REGISTRATION";
		return true;
//		$this->printIt("SUCCESS_REGISTRATION");
//		return $this;
	}

	/**
	* Displays error or success message.
	*
	* @param string	Message to display
	*
	* @return AccountCreator
	*/
	protected function printIt($output)
	{
//		if(is_string($output))
//		{
//			$this->display("<div class=\"success\">".Core::getLanguage()->getItem($output)."</div><br />");
//		}
//		$outstream = "";
//		foreach($output as $output)
//		{
//			$outstream .= "<div class=\"error\">".Core::getLanguage()->getItem($output)."</div><br />";
//		}
//		$this->display($outstream);
		return $this;
	}

	public function getErrors()
	{
		if(is_string($this->errors))
		{
			return ("<div class=\"success\">".Core::getLanguage()->getItem($this->errors)."</div><br />");
		}
		$outstream = "";
		foreach($this->errors as $output)
		{
			$outstream .= "<div class=\"error\">".Core::getLanguage()->getItem($output)."</div><br />";
		}
		return ($outstream);
	}
}
?>