<?php
/**
* Allows the user to change the preferences.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class Preferences extends Page
{
	/**
	* Shows the preferences form.
	*
	* @return void
	*/
	public function __construct()
	{
		parent::__construct();
		Core::getLanguage()->load("Prefs");
		// Hook::event("SHOW_USER_PREFERENCES");
		if(!NS::getUser()->get("umode"))
		{
			$this->setPostAction("resent_activation", "resendActivationMail");
			$this->setPostAction("saveuserdata", "updateUserData")
				->addPostArg("updateUserData", "username")
				->addPostArg("updateUserData", "email")
				->addPostArg("updateUserData", "password")
				->addPostArg("updateUserData", "theme")
				->addPostArg("updateUserData", "language")
				->addPostArg("updateUserData", "templatepackage")
				->addPostArg("updateUserData", "imagepackage")
				->addPostArg("updateUserData", "umode")
				->addPostArg("updateUserData", "delete")
				->addPostArg("updateUserData", "ipcheck")
				->addPostArg("updateUserData", "planetorder")
				->addPostArg("updateUserData", "esps")
				->addPostArg("updateUserData", "show_all_constructions")
				->addPostArg("updateUserData", "show_all_research")
				->addPostArg("updateUserData", "show_all_shipyard")
				->addPostArg("updateUserData", "show_all_defense")
				->addPostArg("updateUserData", "user_bg_style")
				->addPostArg("updateUserData", "user_table_style")
				->addPostArg("updateUserData", "skin_type")
			;
		}
		else if(time() > NS::getUser()->get("umodemin"))
		{
			$this->setPostAction("disable_umode", "disableUmode");
		}
		$this->setPostAction("update_deletion", "updateDeletion")
			->addPostArg("updateDeletion", "delete")
			->proceedRequest();
		return;
	}

	/**
	* Index action.
	*
	* @return Preferences
	*/
	protected function index()
	{
        if(!isNameCharValid(NS::getUser()->get("username"))){
            Logger::addMessage('USERNAME_WRONG_CHARS');
        }
		if(NS::getUser()->get("delete") > 0){
			$delmsg = Date::timeToString(2, NS::getUser()->get("delete"), "", false);
			$delmsg = sprintf(Core::getLanguage()->getItem("DELETE_DATE"), $delmsg);
			Core::getTPL()->assign("delmessage", $delmsg);
		}
		if(NS::getUser()->get("umode")){
			$canDisableUmode = true;
			if(NS::getUser()->get("umodemin") > time())
			{
				$canDisableUmode = false;
				$umodemsg = Date::timeToString(1, NS::getUser()->get("umodemin"), "", false);
				$umodemsg = sprintf(Core::getLanguage()->getItem("UMODE_DATE"), $umodemsg);
				Core::getTPL()->assign("umode_to", $umodemsg);
			}
			Core::getTPL()->assign("can_disable_umode", $canDisableUmode);
		}
		$packs = array();
		$handle = opendir(APP_ROOT_DIR."templates/");
		while($dir = readdir($handle))
		{
			if($dir != "." && $dir != ".." && is_dir(APP_ROOT_DIR."templates/".$dir))
			{
				$packs[]["package"] = $dir;
			}
		}
		// Hook::event("LOAD_TEMPLATE_PACKAGES", array(&$packs));
		Core::getTPL()->addLoop("templatePacks", $packs);

		//handling image packs
		if( !defined('SN') )
		{
			$imgpacks = array();
			$handle = opendir(APP_ROOT_DIR."images/buildings/");
			while($dir = readdir($handle))
			{
				if($dir != "." && $dir != ".." && is_dir(APP_ROOT_DIR."images/buildings/".$dir))
				{
					$item = array("dir" => $dir, "name" => $dir, "order" => 999999, "selected" => intval($dir == NS::getUser()->get("imagepackage")));
					$descr = @file(APP_ROOT_DIR."images/buildings/".$dir."/description.txt");
					if(is_array($descr))
					{
						$item["name"] = isset($descr[0]) ? trim($descr[0]) : $item["name"];
						$item["order"] = isset($descr[1]) ? trim($descr[1]) : $item["order"];
					}
					$imgpacks[] = $item;
				}
			}
			function sort_image_packages($a, $b)
			{
				if($a["order"] < $b["order"]) return -1;
				if($a["order"] > $b["order"]) return 1;
				return strcasecmp($a["name"], $b["name"]);
			}
			usort($imgpacks, "sort_image_packages");
		}
		else
		{
			$imgpacks = array(array(
				'dir' => 'std',
				'name' => 'Standard',
				'order' => '1',
				'selected' => '1',
			));
		}

		Core::getTPL()->addLoop("imagePacks", $imgpacks);

		//user styles
		foreach(array("bg", "table") as $type)
		{
			$styles = getUserStyles($type);
			$currentStyle = NS::getUser()->get("user_{$type}_style");
			foreach($styles as $userStyle)
			{
				if ($userStyle["path"] != $currentStyle)
					Core::getTPL()->addHTMLHeaderFile($userStyle["path"]."?".CLIENT_VERSION, "css");
			}
			Core::getTPL()->assign("current_{$type}_style", $currentStyle);
			Core::getTPL()->addLoop("user_{$type}_styles", $styles);
		}
		//
		$skins = array();
		if( !USE_FACEBOOK_SKIN )
		{
			$skins[] = array("name" => Core::getLanguage()->getItem("SKIN_TYPE_GENERIC"), "value" => SKIN_TYPE_GENERIC);
			$skins[] = array("name" => Core::getLanguage()->getItem("SKIN_TYPE_FB"), "value" => SKIN_TYPE_FB);
			$skins[] = array("name" => Core::getLanguage()->getItem("SKIN_TYPE_MOBI"), "value" => SKIN_TYPE_MOBI);
		}
		Core::getTPL()->assign("current_skin", NS::getUser()->get("skin_type"));
		Core::getTPL()->addLoop("skins", $skins);

		$result = sqlSelect("languages", array("languageid", "title"), "", "", "display_order ASC, title ASC");
		Core::getTPL()->addLoop("langs", $result);

		Core::getTPL()->display("preferences");
		return $this;
	}

	protected function resendActivationMail()
	{
		Core::getLanguage()->load('Registration');
		$activation = $_SESSION["activation"] ?? 0;
		$url = BASE_FULL_URL."signup.php/Activation:".$activation;
		Core::getLang()->assign("regUsername", $_SESSION["username"] ?? "");
		Core::getLang()->assign("activationLink", $url);
		$message = Core::getLanguage()->getItem("REGISTRATION_MAIL_AGAIN");
		$subject = Core::getLanguage()->getItem("REGISTRATION");
		$mail = new Email($_SESSION["email"] ?? "", $subject, $message);
		$mail->sendMail();
		$_SESSION["flash_success"] = "Сообщение отправленно вам на e-mail : " . $_SESSION["email"] ?? "" ;
		doHeaderRedirection("game.php/".Core::getRequest()->getGET("go"), false);
	}

	/**
	* Updates the deletion flag for this user.
	*
	* @param boolean	Enable/Disable deletion
	*
	* @return Preferences
	*/
	protected function updateDeletion($deletion)
	{
		// Deletition
		if( $deletion )
		{
			$delete = time() + 604800;
		}
		else
		{
			$delete = 0;
		}
		// Hook::event("UPDATE_USER_DELETION", array(&$delete));
		sqlUpdate("user", array("delete" => $delete), "userid = ".sqlUser() . ' ORDER BY userid');
		NS::getUser()->rebuild();

		doHeaderRedirection("game.php/Prefs", false);
		return $this;
	}

	/**
	* Saves the entered preferences.
	*
	* @return Preferences
	*/
	protected function updateUserData( $username, $temp_email, $pw, $theme, $language, $templatepackage, $imagepackage, $umode, $delete, $ipcheck, $planetorder, $esps,
		$show_all_constructions, $show_all_research, $show_all_shipyard, $show_all_defense, $user_bg_style, $user_table_style, $skin_type )
	{
		if(NS::getUser()->get("umode")) { throw new GenericException("Vacation mode is still enabled."); }
		Core::getLanguage()->load("Registration");
		$username = trim($username);
		$language = empty($language) ? DEF_LANGUAGE_ID : $language;
		if(preg_match("#[^\w\d\-_\./]#is", $theme) ||	preg_match("#\.\.#is", $theme))
		{
			$theme = "";
		}
		else if(!preg_match("/^.+\/$/i", $theme) && $theme != "")
		{
			$theme .= "/";
		}
		$templatepackage = (empty($templatepackage)) ? "standard" : $templatepackage;
		if(!is_dir(APP_ROOT_DIR."templates/".$templatepackage))
		{
			$templatepackage = "standard";
		}
		$imagepackage = empty($imagepackage) || preg_match("#[^\w\d\-_]#is", $imagepackage) ? "std" : $imagepackage;
		if(!is_dir(APP_ROOT_DIR."images/buildings/".$imagepackage))
		{
			$imagepackage = "std";
		}

		$activation = "";

		// Check language
		if( NS::getUser()->get("languageid") != $language ) // && !defined('SN') && !defined('SN_EXT') )
		{
			$result = sqlSelect("languages", "languageid", "", "languageid = ".sqlVal($language));
			if(Core::getDB()->num_rows($result) <= 0)
			{
				$language = NS::getUser()->get("languageid");
			}
			sqlEnd($result);
		}

		// Check username
        if(!isNameCharValid($username))
        {
            $username = NS::getUser()->get("username");
        }
		if( !Str::compare($username, NS::getUser()->get("username")) )
		{
			$result = sqlSelect("user", "userid", "", "username = ".sqlVal($username));
			if(Core::getDB()->num_rows($result) == 0)
			{
				sqlEnd($result);
				if(!checkCharacters($username))
				{
					$username = NS::getUser()->get("username");
					Logger::addFlashMessage("USERNAME_INVALID");
				}
				else
				{
					Logger::addFlashMessage("USERNAME_CHANGED", 'success');
				}
			}
			else
			{
				sqlEnd($result);
				$username = NS::getUser()->get("username");
				Logger::addFlashMessage("USERNAME_EXISTS");
			}
		}

		// Check email
		$email_activation = '';
		if( !Str::compare($temp_email, NS::getUser()->get("email")) && (!defined('SN') || defined('SN_EXT')) )
		{
			$result = sqlSelect("user", "userid", "", "email = ".sqlVal($temp_email));
			if(Core::getDB()->num_rows($result) == 0)
			{
				sqlEnd($result);
				if(!checkEmail($temp_email))
				{
					$temp_email = NS::getUser()->get("email");
					Logger::addFlashMessage("EMAIL_INVALID");
				}
				else
				{
					if(!Core::getConfig()->get("EMAIL_ACTIVATION_CHANGED_EMAIL"))
					{
						$email_activation = randString(8);
						$url = BASE_FULL_URL."signup.php/Activation:".$email_activation.'/email:confirm';
						$message = Core::getLanguage()->getItemWith('EMAIL_EMAIL_MESSAGE', array('req_username' => $username, 'mail_activation_link' => $url));
//						$message = sprintf(Core::getLanguage()->getItem("EMAIL_EMAIL_MESSAGE"), $username, Core::getOptions()->get("pagetitle"), $url, Core::getOptions()->get("pagetitle"));
						$mail = new Email($temp_email, Core::getLanguage()->getItem("EMAIL_ACTIVATION"), $message);
						$mail->sendMail();
					}
					Logger::addFlashMessage("EMAIL_CHANGED", 'success');
				}
			}
			else
			{
				sqlEnd($result);
				Logger::addFlashMessage("EMAIL_EXISTS");
				$temp_email = NS::getUser()->get("email");
			}
		}

		// Check password
		if ( Str::length($pw) > 0 && ( md5($pw) != sqlSelectField('password', 'password' , '', "userid = ".sqlUser()) )
				&& (!defined('SN') || defined('SN_EXT'))
                && checkEmail(NS::getUser()->get("email")) )
		{
			if(Str::length($pw) >= Core::getOptions()->get("MIN_USER_CHARS"))
			{
				Core::getQuery()->update("password", "password", md5($pw), "userid = ".sqlUser() . ' ORDER BY userid');
				if($activation == "" && !Core::getConfig()->get("EMAIL_ACTIVATION_CHANGED_PASSWORD"))
				{
					$activation = randString(8);
					$url = BASE_FULL_URL."signup.php/Activation:".$activation;
					$message = Core::getLanguage()->getItemWith('EMAIL_PASSWORD_MESSAGE', array('req_username' => $username, 'password' => $pw));
					$mail = new Email(NS::getUser()->get("email"), Core::getLanguage()->getItem("PASSWORD_ACTIVATION"), $message);
					$mail->sendMail();
				}
				Logger::addFlashMessage("PASSWORD_CHANGED", 'success');
			}
			else
			{
				Logger::addFlashMessage("PASSWORD_INVALID");
			}
		}

		// Umode
		if( $umode == 1 ) // && !defined('SN') )
		{
			// Check if umode can be activated
			$planets = NS::getPlanetstack();
			$where = "user=".sqlUser();
			// $where .= $planets ? " AND (planetid IN (".sqlArray($planets).") OR destination IN (".sqlArray($planets)."))" : "";
			$where .= " AND processed=".EVENT_PROCESSED_WAIT;
			$where .= " AND mode IN (".sqlArray($GLOBALS['VACATION_BLOCKING_EVENTS']).")";

			if( 1 ) // release version
			{
				$count = sqlSelectField("events", "count(*)", "", $where);
			}
			else // debug vesion
			{
				$events = array();
				$res = sqlSelect("events", "*", "", $where);
				while( $row = sqlFetch($res) )
				{
					$events[] = $row;
				}
				sqlEnd($res);
				debug_var(array($events, $where), "[check events]");
				$count = count($events);
			}
			if($count > 0)
			{
				Logger::dieMessage("CANNOT_ACTIVATE_UMODE");
			}
			$umode = 1;
			$umodemin = time() + 172800;
			setProdOfUser(NS::getUser()->get("userid"), 0);
		}
		else
		{
			$umode = 0;
		}

		// Deletition
		// $delete = 0;
		if( $delete && (!defined('SN') || defined('SN_EXT')) )
		{
			$delete = time() + 604800;
		}

		// Other prefs
		//ip check
		if( ( !$ipcheck || !defined('IPCHECK') || !IPCHECK ) && (!defined('SN') || defined('SN_EXT')) )
		{
			$ipcheck = 0;
		}

		// Planet order
		switch( $planetorder )
		{
			case 1:
			case 2:
			case 3:
				break;
			default: $planetorder = 1; break;
		}

		// Number of spy drones
		if( $esps > 99 )
		{
			$esps = 99;
		}
		elseif( $esps < 0 )
		{
			$esps = 1;
		}

		// Show blocked constructions
		foreach(array("show_all_constructions", "show_all_research", "show_all_shipyard", "show_all_defense") as $show_all_field)
		{
			$$show_all_field = !empty($$show_all_field) && (strcasecmp($$show_all_field, "on") == 0 || $$show_all_field == 1);
		}

		//user style
		/**
		 * Background style and tables style
		 */
		foreach(array("bg", "table") as $type)
		{
			$styles = getUserStyles($type);
			$usVar = "user_{$type}_style";
			if (!isset($styles[$$usVar]))
				$$usVar = "";
		}

		// Save it
		$values = array(
			"username" 					=> $username,
			"temp_email" 				=> $temp_email,
			"email_activation"			=> $email_activation,
			"languageid" 				=> $language,
			"templatepackage" 			=> $templatepackage,
			"theme" 					=> $theme,
			"ipcheck" 					=> $ipcheck,
			"umode" 					=> $umode,
			"umodemin" 					=> $umodemin,
			"planetorder" 				=> $planetorder,
			"delete" 					=> $delete,
			"esps" 						=> $esps,
			"show_all_constructions" 	=> $show_all_constructions,
			"show_all_research" 		=> $show_all_research,
			"show_all_shipyard" 		=> $show_all_shipyard,
			"show_all_defense" 			=> $show_all_defense,
			"user_bg_style" 			=> $user_bg_style,
			"user_table_style" 			=> $user_table_style,
		);

		if( !USE_FACEBOOK_SKIN && ((!defined('SN') || defined('SN_EXT')) || defined('SN_FULLSCREEN')) && !defined('MOBI') )
		{
			$values["skin_type"] = $skin_type;
		}

		if( !defined('SN') || defined('SN_EXT') )
		{
			$values['imagepackage']	= $imagepackage;
		}

		if( defined('SN') && !defined('SN_EXT') )
		{
//			unset($values['username']);
			unset($values['temp_email']);
			unset($values['email_activation']);
			unset($values['ipcheck']);
			unset($values['delete']);
			// unset($values['languageid']);
			unset($values['templatepackage']);
		}

		sqlUpdate("user", $values, "userid = ".sqlUser() . ' ORDER BY userid');

		NS::getUser()->rebuild();
		// return $this->index();
		doHeaderRedirection("game.php/".Core::getRequest()->getGET("go"), false);
	}

	/**
	* Disables the vacation mode and starts the resource production.
	*
	* @return Preferences
	*/
	protected function disableUmode()
	{
		setProdOfUser(NS::getUser()->get("userid"), 100);
		Core::getQuery()->update("user", array("umode"), array(0), "userid = ".sqlUser() . ' ORDER BY userid');
		NS::getUser()->rebuild();
		Logger::dieMessage("UMODE_DISABLED", "info");
		return $this;
	}
}
?>