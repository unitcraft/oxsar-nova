<?php
/**
* Sends email, if password or username has been forgotten.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

require_once(APP_ROOT_DIR."game/Functions.inc.php");
require_once(RECIPE_ROOT_DIR."AjaxRequestHelper.abstract_class.php");

class LostPassword extends AjaxRequestHelper
{
	const LOST_USERNAME = 1;
	const LOST_PASSWORD = 2;
	/**
	* @var string
	*/
	protected $username = "";

	/**
	* @var string
	*/
	protected $email = "";

	/**
	* Email message which will be sent.
	*
	* @var string
	*/
	protected $message = "";

	/**
	* Security key transmitted as link.
	* Serves as verification.
	*
	* @var string
	*/
	protected $secKey = "";
	
	protected $id;

	/**
	* Handles lost password requests.
	*
	* @param string	Entered username
	* @param string	Entered email address
	*
	* @return void
	*/
	public function __construct($username, $email)
	{
		$this->username = $username;
		$this->email	= $email;
		$mode			= self::LOST_USERNAME;
		if($this->username != '')
		{
			$mode = self::LOST_PASSWORD;
		}
		if( !checkEmail($this->email) )
		{
			$this->printIt('EMAIL_INVALID');
		}

		// План 37.5d.5#8: replaced User_YII::find() + CDbCriteria.
		$user = sqlSelectRow("user", array("userid", "username"), "", "email=".sqlVal($this->email));
		if( !$user )
		{
			$this->printIt('EMAIL_NOT_FOUND');
			return $this;
		}
		$this->id = $user["userid"];
		Core::getLanguage()->assign('req_username', $user["username"]);
		Core::getLanguage()->assign('req_ipaddress', IPADDRESS);

		if($mode == self::LOST_USERNAME)
		{
			$this->message = Core::getLanguage()->getItem('REQUEST_USERNAME');
		}
		else if( Str::compare($this->username, $user["username"]) )
		{
			$this->secKey	= randString(8);
			$reactivate		= HTTP_HOST . REQUEST_DIR . 'signup.php/Activation:' . $this->secKey;
			$url			= HTTP_HOST . REQUEST_DIR . 'forgottenpw.php/NewPassword:' . $this->secKey . '/User:' . $user["userid"];
			
			Core::getLanguage()->assign('new_pw_url', $url);
			Core::getLanguage()->assign('reactivate_url', $reactivate);
			
			$this->message	= Core::getLanguage()->getItem('REQUEST_PASSWORD');
			$this->setNewPw();
		}
		else
		{
			$this->printIt('USERNAME_DOES_NOT_EXIST');
			return $this;
		}

		$this->sendMail($mode);
		return;
	}

	/**
	* Saves the new password.
	*
	* @return LostPassword
	*/
	protected function setNewPw()
	{
		if( !empty( $this->id ) )
		{
			// Updated By Pk
			Core::getQuery()->update("user", "password_activation", $this->secKey, "userid = ".sqlVal($this->id));
		}
		return $this;
	}

	/**
	* Sends email with lost password message.
	*
	* @param integer	Mode
	*
	* @return LostPassword
	*/
	protected function sendMail($mode)
	{
		$mail = new Email($this->email, Core::getLanguage()->getItem("LOST_PW_SUBJECT_".$mode), $this->message);
		$mail->sendMail();
		$this->printIt("PW_LOST_EMAIL_SUCCESS", false);
		return $this;
	}

	/**
	* Prints either an error or success message.
	*
	* @param string	Output stream
	* @param boolean	Message is error
	*
	* @return LostPassword
	*/
	protected function printIt($output, $error = true)
	{
		if($error)
		{
			$this->display("<div class=\"error\">".Core::getLanguage()->getItem($output)."</div>");
		}
		$this->display("<div class=\"success\">".Core::getLanguage()->getItem($output)."</div><br />");
		return $this;
	}
}
?>