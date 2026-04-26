<?php
/**
* Account activation function.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class AccountActivation
{
	/**
	* Activation key to verify email address.
	*
	* @var string
	*/
	protected $key;
	
	/**
	* Is it e-mail activation.
	*
	* @var bool
	*/
	protected $email;

	/**
	* Constructor.
	*
	* @param string Activation key
	*
	* @return void
	*/
	public function __construct($key, $email = FALSE)
	{
		$this->key = $key;
		$this->email = $email;
//		$this->activateAccount();
		return;
	}

	/**
	* Activates an account, if key exists.
	* Starts log in on success.
	*
	* @return AccountActivation
	*/
	public function activateAccount()
	{
		$result = sqlSelect(
			"user u",
			array("u.userid", "u.username", "p.password", "temp_email"),
			"LEFT JOIN ".PREFIX."password p ON (p.userid = u.userid)",
			"u." . ( ($this->email) ? 'email_' : '' ) . "activation = ".sqlVal($this->key)
		);
		if($row = sqlFetch($result))
		{
			sqlEnd($result);
			
			// Update By Pk
			$user = User_YII::model()->findByPK($row["userid"]);
			if( $this->email )
			{
				$user->email_activation		= '';
				$user->password_activation	= '';
				$user->email				= $user->temp_email;
			}
			$user->activation	= '';
			$user->save(false);
//			Core::getQuery()->update(
//				"user",
//				array( ( ($this->email) ? 'email_' : '' ) . "activation", "email"),
//				array("", $row["temp_email"]),
//				"userid = ".sqlVal($row["userid"])
//					// . ' ORDER BY userid'
//			);
//			if( !$this->email )
//			{
//				// Update By Pk
//				Core::getQuery()->update(
//					"user",
//					array("password_activation"),
//					array(""),
//					"userid = ".sqlVal($row["userid"])
//						// . ' ORDER BY userid'
//				);
//			}
			return true;
		}
		sqlEnd($result);
//		throw new GenericException("Activation failed.");
		return false;
	}
}
?>