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
		// План 37.5d.5#6: replaced Sessions_YII (find + save + deleteAll) → raw SQL.
		// PK na_sessions.sessionid, фильтр по userid.
		sqlUpdate("sessions", array("logged" => 0), "userid=".sqlVal($user_id));

		if(Core::getConfig()->exists("SESSION_SAVING_DAYS"))
		{
			$days = intval(Core::getConfig()->get("SESSION_SAVING_DAYS"));
		}
		else
		{
			$days = $this->savingDays;
		}
		$deleteTime = time() - 60*60*24 * max(1, $days);
		sqlDelete("sessions", "time < ".sqlVal($deleteTime));
		
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