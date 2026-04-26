<?php
/**
* Clears user cache and disables session.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class Logout
{
	/**
	* Deletes all sessions older than $days.
	*
	* @var integer
	*/
	protected $savingDays = 30;

	/**
	* Perfom log out proccess.
	*
	* @return void
	*/
	public function __construct()
	{
		$user_id = $_SESSION["userid"] ?? 0;
//		Core::getCache()->cleanUserCache($user_id);
		$session = Sessions_YII::model()->find('userid = ?', $user_id );
		if($session)
		{
			$session->attributes = array('logged' => '0');
			$session->save();
		}
		if(Core::getConfig()->exists("SESSION_SAVING_DAYS"))
		{
			$days = intval(Core::getConfig()->get("SESSION_SAVING_DAYS"));
		}
		else
		{
			$days = $this->savingDays;
		}
		$deleteTime = time() - 60*60*24 * max(1, $days);
		Sessions_YII::model()->deleteAll('time < ?', array($deleteTime));
		
		header("Location: " . BASE_FULL_URL); exit();
		/*
		error_log(RELATIVE_URL);
		header('Location: ' . RELATIVE_URL );
		exit();
		return;
		*/
	}
}
?>