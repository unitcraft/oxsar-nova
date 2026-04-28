<?php
/**
* Allows moderators to change user data and manage bans.
*
* Oxsar http://oxsar.ru
*
* 
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class Moderator extends Page
{
	/**
	* User id of currently moderating user.
	*
	* @var integer
	*/
	protected $userid = 0;

	/**
	* Handles the moderator class.
	*
	* @return void
	*/
	public function __construct()
	{
		NS::getUser()->checkPermissions("CAN_MODERATE_USER");
		parent::__construct();
		Core::getLanguage()->load(array("Prefs" ,"Statistics"));
		$this->userid = (Core::getRequest()->getPOST("userid")) ? Core::getRequest()->getPOST("userid") : Core::getRequest()->getGET("id");

		if($this->userid)
		{
			$this->setPostAction("proceedban", "proceedBan")
				->addPostArg("proceedBan", "ban")
				->addPostArg("proceedBan", "timeend")
				->addPostArg("proceedBan", "reason")
				->addPostArg("proceedBan", "b_umode")
				->setPostAction("proceedro", "proceedRO")
				->addPostArg("proceedRO", "ro")
				->addPostArg("proceedRO", "timeendro")
				->addPostArg("proceedRO", "reasonro") 
				->setPostAction("proceed", "proceed")
				->addPostArg("proceed", "username")
				->addPostArg("proceed", "email")
				->addPostArg("proceed", "delete")
				->addPostArg("proceed", "umode")
				->addPostArg("proceed", "activation")
				->addPostArg("proceed", "ipcheck")
				->addPostArg("proceed", "usergroupid")
				// ->addPostArg("proceed", "points")
				// ->addPostArg("proceed", "fpoints")
				// ->addPostArg("proceed", "rpoints")
				->addPostArg("proceed", "password")
				->addPostArg("proceed", "languageid")
				->addPostArg("proceed", "templatepackage")
				->addPostArg("proceed", "theme")
				->setGetAction("do", "AnnulBan", "annulBan")
				->addGetArg("annulBan", "banid")
				->proceedRequest();
		}
		return;
	}

	/**
	* Displays the moderator form to edit users.
	*
	* @return Moderator
	*/
	protected function index()
	{
		$select = array(
			"u.userid", "u.username", "u.email", "u.temp_email", "u.languageid", "u.templatepackage", "u.theme", 
			"u.points", "u.b_points", "u.r_points", "u.u_points", "u.b_count", "u.r_count", "u.u_count", "u.e_points", "u.battles",
			"u.ipcheck", "u.activation", "u.last", "u.umode", "u.umode", "u.delete", "u.regtime",
			"a.tag", "a.name", "u2g.usergroupid"
			);
		$joins	= "LEFT JOIN ".PREFIX."user2ally u2a ON (u2a.userid = u.userid)";
		$joins .= "LEFT JOIN ".PREFIX."alliance a ON (a.aid = u2a.aid)";
		$joins .= "LEFT JOIN ".PREFIX."user2group u2g ON (u.userid = u2g.userid)";
		$result = sqlSelect("user u", $select, $joins, "u.userid = ".sqlVal($this->userid));
		if($row = sqlFetch($result))
		{
			sqlEnd($result);
			// Hook::event("MODERATE_USER", array(&$row));
			$row["deletion"] = $row["delete"];
			unset($row["delete"]);
			$row["vacation"] = $row["umode"];
			unset($row["umode"]);

			$row["last"] = Date::timeToString(1, $row["last"]);
			$row["regtime"] = Date::timeToString(1, $row["regtime"]);

			foreach($row as $key => $value)
				{
					Core::getTPL()->assign($key, $value);
				}

			if($row["usergroupid"] == 4)
				{
					Core::getTPL()->assign("isMod", true);
				}
			else if($row["usergroupid"] == 2)
				{
					Core::getTPL()->assign("isAdmin", true);
				}

			$result = sqlSelect("languages", array("languageid", "title"), "", "", "title ASC");
			Core::getTPL()->addLoop("langs", $result);

			$bans = array(); $i = 0;
			$result = sqlSelect("ban_u", array("`banid`", "`to`", "`reason`"), "", "userid = ".sqlVal($this->userid));
			while($row = sqlFetch($result))
				{
					$bans[$i]["reason"] = $row["reason"];
					$bans[$i]["to"] = Date::timeToString(1, $row["to"]);
					if($row["to"] > time()) { $bans[$i]["annul"] = Link::get("game.php/Moderator/".$this->userid."/do:AnnulBan/banid:".$row["banid"], Core::getLanguage()->getItem("ANNUL")); }
					else { $bans[$i]["annul"] = Core::getLanguage()->getItem("ANNUL"); }
					$i++;
				}
			sqlEnd($result);
			Core::getTPL()->addLoop("bans", $bans);
			Core::getTPL()->assign("eBans", $i);

			$ros = array(); $i = 0;
			$result = sqlSelect("chat_ro_u", array("`roid`", "`to`", "`reason`"), "", "userid = ".sqlVal($this->userid));
			while($row = sqlFetch($result))
				{
					$ros[$i]["reason"] = $row["reason"];
					$ros[$i]["to"] = Date::timeToString(1, $row["to"]);
					if($row["to"] > time()) { $ros[$i]["annul"] = Link::get("game.php/Moderator/".$this->userid."/do:AnnulRO/roid:".$row["roid"], Core::getLanguage()->getItem("ANNUL")); }
					else { $ros[$i]["annul"] = Core::getLanguage()->getItem("ANNUL"); }
					$i++;
				}
			sqlEnd($result);
			Core::getTPL()->addLoop("ros", $ros);
			Core::getTPL()->assign("eROs", $i); 

			Core::getTPL()->display("moderate_user");
		}
		return $this->quit();
	}

	/**
	* Bans an user.
	*
	* @param integer	Ban factor
	* @param integer	Ban time
	* @param string	Reason for ban
	* @param boolean	Set user into umode
	*
	* @return Moderator
	*/
	protected function proceedBan($ban, $timeEnd, $reason, $forceUmode)
	{
		$to = time() + $ban * $timeEnd;
		if($to > 9999999999) { $to = 9999999999; }
		// Hook::event("BAN_USER", array(&$to, $reason, $forceUmode));
		$atts = array("userid", "from", "to", "reason");
		$vals = array($this->userid, time(), $to, $reason);
		Core::getQuery()->insert("ban_u", $atts, $vals);
		if($forceUmode)
		{
			Core::getQuery()->update("user", array("umode"), array(1), "userid = ".sqlVal($this->userid));
			setProdOfUser($this->userid, 0);
		}
		Core::getQuery()->update("sessions", array("logged"), array(0), "userid = ".sqlVal($this->userid));
		return $this->index();
	}

	/**
	* Annuls a ban.
	*
	* @param integer	Ban to annul
	*
	* @return Moderator
	*/
	protected function annulBan($banid)
	{
		// Hook::event("UNBAN_USER", array($banid));
		Core::getQuery()->update("ban_u", array("to", "reason"), array(time(), Core::getLanguage()->getItem("ANNULED")), "banid = ".sqlVal($banid));
		$result = sqlSelect("ban_u", "userid", "", "banid = ".sqlVal($banid));
		$row = sqlFetch($result);
		sqlEnd($result);
		doHeaderRedirection("game.php/Moderator/".$row["userid"], false);
		return $this;
	}

	protected function proceedRO($ro, $timeEndro, $reasonro)
	{
		$to = time() + $ro * $timeEndro;
		if($to > 9999999999) { $to = 9999999999; }
		// Hook::event("RO_USER", array(&$to, $reasonro));
		$atts = array("userid", "from", "to", "reason");
		$vals = array($this->userid, time(), $to, $reasonro);
		Core::getQuery()->insert("chat_ro_u", $atts, $vals);
		Core::getQuery()->update("sessions", array("logged"), array(0), "userid = ".sqlVal($this->userid));
		return $this->index();
	}
	/**
	* Annuls a ban.
	*
	* @param integer	Ban to annul
	*
	* @return Moderator
	*/
	protected function annulRO($roid)
	{
		// Hook::event("UNRO_USER", array($banid));
		Core::getQuery()->update("chat_ro_u", array("to", "reason"), array(time(), Core::getLanguage()->getItem("ANNULED")), "roid = ".sqlVal($roid));
		$result = sqlSelect("chat_ro_u", "userid", "", "roid = ".sqlVal($roid));
		$row = sqlFetch($result);
		sqlEnd($result);
		doHeaderRedirection("game.php/Moderator/".$row["userid"], false);
		return $this;
	} 

	/**
	* Updates the moderator form.
	*
	* @param array _POST
	*
	* @return Moderator
	*/
	protected function proceed($username, $email, $delete, $umode, $activation, $ipcheck, $usergroupid, /*$points, $fpoints, $rpoints,*/ $password, $languageid, $templatepackage, $theme)
	{
		$select = array("userid", "username", "email");
		$result = sqlSelect("user", $select, "", "userid = ".sqlVal($this->userid));
		if($row = sqlFetch($result))
		{
			sqlEnd($result);
			// Hook::event("SAVE_USER_MODERATION", array(&$row));
			$delete = ($delete == 1) ? 1 : 0;
			$umode = ($umode == 1) ? 1 : 0;
			$activation = ($activation == 1) ? "" : "1";
			$ipcheck = ($ipcheck == 1) ? 1 : 0;

			if(NS::getUser()->ifPermissions("CAN_EDIT_USER"))
			{
				Core::getQuery()->delete("user2group", "userid = ".sqlVal($this->userid));
				Core::getQuery()->insert("user2group", array("usergroupid", "userid"), array($usergroupid, $this->userid));
				// Core::getQuery()->update("user", array("points", "fpoints", "rpoints"), array(floatval($points), intval($fpoints), intval($rpoints)), "userid = ".sqlVal($this->userid));
			}

			if($umode)
			{
				setProdOfUser($this->userid, 0);
			}

			if(!Str::compare($username, $row["username"]))
			{
				$num = sqlSelectField("user", "count(*)", "", "username = ".sqlVal($username));
				if($num > 0)
				{
					$username = $row["username"];
				}
			}

			if(!Str::compare($email, $row["email"]))
			{
				$num = sqlSelectField("user", "count(*)", "", "email = ".sqlVal($email));
				if($num > 0)
				{
					$email = $row["email"];
				}
			}

			if(Str::length($password) > 0)
			{
				Core::getQuery()->update("password", "password", md5($password), "userid = ".sqlVal($this->userid));
			}

			$atts = array("username", "email", "delete", "umode", "activation", "languageid", "ipcheck", "templatepackage", "theme");
			$vals = array($username, $email, $delete, $umode, $activation, $languageid, $ipcheck, $templatepackage, $theme);
			Core::getQuery()->update("user", $atts, $vals, "userid = ".sqlVal($this->userid));
		}
		return $this->index();
	}
}
?>