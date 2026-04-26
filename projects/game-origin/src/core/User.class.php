<?php
/**
* Loads user data.
*
* @package Recipe 1.1
* @author Sebastian Noll
* @copyright Copyright (c) 2008, Sebastian Noll
* @license <http://www.gnu.org/licenses/gpl.txt> GNU/GPL
* @version $Id: User.class.php 23 2010-04-03 19:08:34Z craft $
*/

if(!defined("RECIPE_ROOT_DIR")) { die("Hacking attempt detected."); }

class User extends Collection
{
	/**
	* Session ID.
	*
	* @var string
	*/
	protected $sid = "";

	/**
	* Permissions.
	*
	* @var array
	*/
	protected $permissions = array();

	/**
	* User groups.
	*
	* @var array
	*/
	protected $groups = array();

	/**
	* Particular group data.
	*
	* @var array
	*/
	protected $groupdata = array();

	/**
	* If user is guest.
	*
	* @var boolean
	*/
	protected $isGuest = false;

	/**
	* Enables cache operation.
	*
	* @var boolean
	*/
	protected $cacheActive = false;

	/**
	* Default user group for guests.
	*
	* @var integer
	*/
	protected $guestGroupId = 1;

	/**
	* Allowed timezones.
	* TODO: Someday we should consider daylight saving time.
	*
	* @var array
	*/
	protected $timezones = array(
		"-11" => "Pacific/Samoa",
		"-10" => "Pacific/Honolulu",
		"-9" => "America/Anchorage",
		"-8" => "America/Los_Angeles",
		"-7" => "America/Denver",
		"-6" => "America/Chicago",
		"-5" => "America/New_York",
		"-4.5" => "America/Caracas",
		"-4" => "America/Asuncion",
		"-3.5" => "America/St_Johns",
		"-3" => "America/Argentina/Buenos_Aires",
		"-2" => "Atlantic/South_Georgia",
		"-1" => "Atlantic/Cape_Verde",
		"0" => "Europe/London",
		"1" => "Europe/Paris",
		"2" => "Europe/Istanbul",
		"3" => "Europe/Moscow",
		"4" => "Asia/Baku",
		"5" => "Asia/Bishkek",
		"5.5" => "Asia/Colombo",
		"5.75" => "Asia/Katmandu",
		"6" => "Asia/Dhaka",
		"6.5" => "Indian/Cocos",
		"7" => "Asia/Bangkok",
		"8" => "Asia/Singapore",
		"9" => "Asia/Tokyo",
		"9.5" => "Australia/Adelaide",
		"10" => "Australia/Queensland",
		"10.5" => "Australia/Lord_Howe",
		"11" => "Pacific/Noumea",
		"11.5" => "Pacific/Norfolk",
		"12" => "Pacific/Fiji"
		);

	/**
	* Constructor: Starts user check.
	*
	* @param string	The session id
	*
	* @return void
	*/
	public function __construct($sid)
	{
		if( !empty($_SESSION["userid"]) )
		{
			$this->sid = $_SESSION["sid"] ?? "";
		}
		else
		{
			$this->sid = $sid;
		}
		
		$this->setGuestGroupId();
		
		if( $this->cacheActive )
		{
			$this->cacheActive = CACHE_ACTIVE;
		}
		
		$this->getData();
		if( !$this->isGuest )
		{
			$this->setGroups();
		}
		$this->setPermissions();
		$this->setTimezone();
		return;
	}

	/**
	* Fetches the user data in term of a session id.
	*
	* @return boolean	True if user is valid, false if not
	*/
	protected function getData()
	{
		if( !empty($_SESSION["userid"]) )
		{
			$this->sid = $_SESSION["sid"] ?? "";
		}

		if( $this->cacheActive )
		{
			$this->item = Core::getCache()->getUserCache($this->sid);
		}

		if( empty($this->item) && !empty($_SESSION["userid"]) )
		{
			// Прямой SQL-запрос вместо Yii ActiveRecord
			$uid = (int)$_SESSION["userid"];
			$sql = "SELECT u.*, a.aid AS allyid, a.name AS ally_name, a.tag AS ally_tag"
			     . " FROM `" . PREFIX . "user` u"
			     . " LEFT JOIN `" . PREFIX . "user2ally` ua ON ua.userid = u.userid"
			     . " LEFT JOIN `" . PREFIX . "alliance` a ON a.aid = ua.aid"
			     . " WHERE u.userid = " . $uid . " LIMIT 1";
			$result = Core::getDB()->query($sql);
			$user_record = Core::getDB()->fetch($result);
			if( $user_record )
			{
				$this->item = $user_record;
				$this->item['banned'] = false;
			}
//			$select = array("u.*", "b.to as ban_to", "b.banid", "s.ipaddress");
//			$joins	= "LEFT JOIN ".PREFIX."user u ON s.userid = u.userid";
//			$joins .= " LEFT JOIN ".PREFIX."ban_u b ON u.userid = b.userid";
//
//			// Get custom user data from configuration
//			if(Core::getConfig()->exists("userselect"))
//			{
//				$userConfigSelect = Core::getConfig()->get("userselect");
//				$select = array_merge($select, $userConfigSelect["fieldsnames"]);
//			}
//			if(Core::getConfig()->exists("userjoins"))
//			{
//				$joins .= " ".Str::replace("PREFIX", PREFIX, Core::getConfig()->get("userjoins"));
//			}
//			$this->item = Core::getQuery()->selectRow("sessions s", $select, $joins, "s.sessionid = ".sqlVal($this->sid)." AND s.logged = 1");
//
//			$this->item["banned"] = $this->item["banid"] && (is_null($this->item["ban_to"]) || $this->item["ban_to"] >= time());
//			unset($this->item["banid"], $this->item["ban_to"]);
		}
		
		if( $this->size() > 0 )
		{
			if( $GLOBALS["RUN_YII"] == 1 && !empty($_SESSION["userid"]) )
			{
				define("SID", '');
			}
			else
			{
				define("SID", $this->sid);
			}
			if( !defined('SN') && IPCHECK && $this->get("ipcheck") )
			{
				if($this->get("ipaddress") != IPADDRESS)
				{
					if(!defined("ADMIN_IP") || IPADDRESS != ADMIN_IP)
					{
						forwardToLogin("IPADDRESS_INVALID");
					}
				}
			}
			if($this->get("templatepackage") != "") { Core::getTemplate()->setTemplatePackage($this->get("templatepackage")); }
		}
		else
		{
			if(LOGIN_REQUIRED && !defined("LOGIN_PAGE"))
			{
				forwardToLogin("NO_ACCESS");
			}
			$this->setGuest();
		}
		// Hook::event("USER_DATA_LOADED", array(&$this));
		return $this;
	}

	/**
	* Sets the user groups.
	*
	* @return User
	*/
	protected function setGroups()
	{
		$select = array("usergroupid", "data");
		$result = Core::getQuery()->select("user2group", $select, "", "userid = ".sqlVal($this->get("userid")));
		while($row = Core::getDatabase()->fetch($result))
		{
			array_push($this->groups, $row["usergroupid"]);
			$this->groupdata[$row["usergroupid"]] = $row["data"];
		}
		Core::getDatabase()->free_result($result);
		return $this;
	}

	/**
	* Fetches permissions of an usergroup. Regard: Permissions are
	* favoured in opposite of restrictions.
	*
	* @return User
	*/
	protected function setPermissions()
	{
		if(count($this->groups) > 1)
		{
			foreach($this->groups as $group)
			{
				if($this->cacheActive)
				{
					$permCache = Core::getCache()->getPermissionCache($group);
					foreach($permCache as $key => $value)
					{
						if($this->permissions[$key] == 0 || !isset($this->permissions[$key])) { $this->permissions[$key] = $value; }
					}
				}
				else
				{
					$select = array("p.permission", "p2g.value");
					$result = Core::getQuery()->select("group2permission p2g", $select, "LEFT JOIN ".PREFIX."permissions AS p ON (p2g.permissionid = p.permissionid)", "p2g.groupid = ".sqlVal($group));
					while($row = Core::getDatabase()->fetch($result))
					{
						if($this->permissions[$row["permission"]] == 0 || !isset($this->permissions[$row["permission"]])) { $this->permissions[$row["permission"]] = $row["value"]; }
					}
					Core::getDatabase()->free_result($result);
				}
			}
		}
		else if(count($this->groups) > 0)
		{
			if($this->cacheActive)
			{
				$this->permissions = Core::getCache()->getPermissionCache($this->groups[0]);
			}
			else
			{
				$select = array("p.permission", "p2g.value");
				$result = Core::getQuery()->select("group2permission p2g", $select, "LEFT JOIN ".PREFIX."permissions AS p ON (p2g.permissionid = p.permissionid)", "p2g.groupid = ".sqlVal($this->groups[0]));
				while($row = Core::getDatabase()->fetch($result))
				{
					$this->permissions[$row["permission"]] = $row["value"];
				}
				Core::getDatabase()->free_result($result);
			}
		}
		return $this;
	}

	/**
	* Checks if user has permissions and if so issue an error.
	*
	* @param string	The permissions; Splitted with commas
	*
	* @return User
	*/
	public function checkPermissions($perms)
	{
		// Hook::event("CHECK_USER_PERMISSIONS", array($perms));
		if(!$this->ifPermissions($perms))
		{
			// echo "NO_ACCESS !permissions"; exit;
			forwardToLogin("NO_ACCESS");
		}
		return $this;
	}

	/**
	* Checks if user has permissions.
	*
	* @param mixed		The permissions; Splitted with commas
	*
	* @return boolean	True, if all permissions are valid, false, if not
	*/
	public function ifPermissions($perms)
	{
		if(!is_array($perms)) { $perms = explode(",", Arr::trimArray($perms)); }
		// Hook::event("USER_HAS_PERMISSIONS", array($perms));
		foreach($perms as $p)
		{
			if($this->permissions[$p] == 0 || !$this->permissions[$p])
			{
				return false;
			}
		}
		return true;
	}

	/**
	* Assign user to group "guest".
	*
	* @return User
	*/
	protected function setGuest()
	{
		$this->isGuest = true;
		array_push($this->groups, $this->guestGroupId);
		return $this;
	}

	/**
	* Rebuild user cache.
	*
	* @return User
	*/
	public function rebuild()
	{
		if( $this->cacheActive )
		{
//			Core::getCache()->buildUserCache($this->sid);
//			$this->item = Core::getCache()->getUserCache($this->sid);
		}
		$this->getData();
		return $this;
	}

	/**
	* Checks the timezone.
	*
	* @return User
	*/
	public function setTimezone()
	{
		$timezone = null;
		if($this->exists("timezone") && is_numeric($this->get("timezone")) && array_key_exists($this->get("timezone"), $this->timezones))
		{
			$timezone = $this->timezones[$this->get("timezone")];
		}
		Core::setDefaultTimeZone($timezone);
		return $this;
	}

	/**
	* Returns a session value.
	*
	* @param string	Session variable
	*
	* @return mixed	Value
	*/
	public function get($var)
	{
		if(is_null($var))
		{
			return $this->item;
		}
		if($this->exists($var))
		{
			return $this->item[$var];
		}
		return false;
	}

	/**
	* Sets a session value.
	*
	* @param string	Session variable
	* @param mixed		Value
	*
	* @return User
	*/
	public function set($var, $value)
	{
		if(Str::compare($var, "userid"))
		{
			throw new GenericException("The primary key of a data record cannot be changed.");
		}
		$this->item[$var] = $value;
		if($this->exists($var))
		{
			Core::getQuery()->update("user", array($var), array($value), "userid = ".sqlVal($this->get("userid")));
			$this->rebuild();
		}
		return $this;
	}

	/**
	* Checks group membership.
	*
	* @param integer	Group id
	*
	* @return boolean
	*/
	public function inGroup($groupid)
	{
		return (in_array($groupid, $this->groups)) ? true : false;
	}

	/**
	* Returns specific group membership data.
	*
	* @param integer	Group id
	*
	* @return mixed	Group membership data
	*/
	public function getGroupData($groupid)
	{
		return (isset($this->groupdata[$groupid])) ? $this->groupdata[$groupid] : false;
	}

	/**
	* Returns the session id.
	*
	* @return string
	*/
	public function getSid()
	{
		return $this->sid;
	}

	/**
	* Returns if the user is in guest mode.
	*
	* @return boolean
	*/
	public function isGuest()
	{
		return $this->isGuest;
	}

	/**
	* Sets the guest group id.
	*
	* @param integer	Guest group id [optional]
	*
	* @return User
	*/
	public function setGuestGroupId($guestGroupId = null)
	{
		if(is_null($guestGroupId))
		{
			$guestGroupId = Core::getConfig()->guestgroupid;
		}
		$this->guestGroupId = $guestGroupId;
		return $this;
	}
}
?>