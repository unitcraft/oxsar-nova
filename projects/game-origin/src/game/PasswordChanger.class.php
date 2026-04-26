<?php
/**
* Changes the password, if the key passed the check.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

require_once(APP_ROOT_DIR."game/Functions.inc.php");
require_once(RECIPE_ROOT_DIR."AjaxRequestHelper.abstract_class.php");

class PasswordChanger extends AjaxRequestHelper
{
	/**
	* Key check and set new password.
	*
	* @param integer User id
	* @param string Key, transmitted by email
	* @param string New password
	*
	* @return void
	*/
	public function __construct($userid, $key, $newpw)
	{
		if(
			Str::length($newpw) < Core::getOptions()->get("MIN_USER_CHARS")
			|| Str::length($newpw) > Core::getOptions()->get("MAX_USER_CHARS")
		)
		{
			$this->printIt("PASSWORD_INVALID");
			return $this;
		}

		$result = sqlSelect("user", "userid", "", "userid = ".sqlVal($userid)." AND password_activation = ".sqlVal($key));
		if(Core::getDB()->num_rows($result) > 0)
		{
			sqlEnd($result);
			// Updated By Pk
			Core::getQuery()->update("password", array("password", "time"), array(md5($newpw), time()), "userid = ".sqlVal($userid));
			// Updated By Pk
			sqlUpdate("user", array('password_activation' => '', 'activation' => ''), "userid = ".sqlVal($userid));
			$this->printIt("PASSWORD_CHANGED", false);
			return $this;
		}
		sqlEnd($result);
		$this->printIt("ERROR_PASSWORD_CHANGED");
		return;
	}

	/**
	* Prints either an error or success message.
	*
	* @param string Output stream
	* @param boolean Message is error
	*
	* @return PasswordChanger
	*/
	protected function printIt($output, $error = true)
	{
		if($error === true)
		{
			$this->display("<div class=\"error\">".Core::getLanguage()->getItem($output)."</div>");
		}
		$this->display("<div class=\"success\">".Core::getLanguage()->getItem($output)."</div><br />");
		return $this;
	}
}
?>