<?php
/**
* This class helps to create a user list.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

require_once(APP_ROOT_DIR."game/Relation.class.php");
require_once(APP_ROOT_DIR."game/AllianceList.class.php");

class UserList extends AllianceList
{
  /**
  * Images which can be used in a list.
  *
  * @var string
  */
  protected $pmPic = "", $buddyPic = "", $modPic = "";

  /**
  * Points of the current user.
  *
  * @var integer
  */
  protected $points = 0;

  /**
  * Consider newbie protection when formatting the usernames?
  *
  * @var boolean
  */
  protected $newbieProtection = false;

  protected $type = false;

  /**
  * Creates a new user list object.
  *
  * @param resource	Query result for a list
  *
  * @return void
  */
  public function __construct($list = null, $start = 0, $newbieProtection = false, $maxDecimals = 2, $type = null)
  {
    $this->newbieProtection = $newbieProtection;
    $this->pmPic = Image::getImage("pm.gif", Core::getLanguage()->getItem("WRITE_MESSAGE"));
    $this->buddyPic = Image::getImage("b.gif", Core::getLanguage()->getItem("ADD_TO_BUDDYLIST"));
    $this->modPic = Image::getImage("moderator.gif", Core::getLanguage()->getItem("MODERATE"));
	$this->premiumSellerPic = Image::getImage("exch_lot_vip_icon.png", Core::getLanguage()->getItem("PREMIUN_SELLER_INFO"));
    $this->points = NS::getUser()->get("points");
	$this->type = $type;
    parent::__construct($list, $start, $maxDecimals);
  }

  private static function getUserPremiumCredit($userid)
  {
		$cache_name = "UserPremiumCredit($userid)";
		if(!NS::getMCH()->get($cache_name, $credit))
		{
			$credit = sqlQueryField("SELECT abs(sum(rl.credit)) FROM ".PREFIX."res_log rl WHERE rl.userid=".sqlVal($userid)
				." AND type=".sqlVal(RES_UPDATE_EXCH_LOT_PREMIUM)
				." AND time >= (now() - interval 1 day)");
			NS::getMCH()->set($cache_name, $credit, randRoundRange(60*60*1, 60*60*2));
		}
		return $credit;
  }

  /**
  * Formats an user record.
  *
  * @param array		User data
  *
  * @return array	Formatted user data
  */
  protected function formatRow($row)
  {
    // Hook::event("FORMAT_USERLIST_ROW_FIRST", array(&$row, &$this));
    if(empty($row["userid"]))
    {
      return $row;
    }

    // Quick buttons
    $row["message"] = Link::get("game.php/MSG/Write/Receiver:".rawurlencode($row["username"]), $this->pmPic);
    $row["buddyrequest"] = Link::get("game.php/go:Friends/id:Add/User:".$row["userid"], $this->buddyPic);
    $row["moderator"] = Link::get("game.php/Moderator/".$row["userid"], $this->modPic);

	if($this->newbieProtection) // $this->type == "points") // && isAdmin())
	{
		$premium_credit = self::getUserPremiumCredit($row["userid"]);
		if($premium_credit > 100)
		{
			$row["premium_seller"] = $this->premiumSellerPic;
		}
	}

    $userClass = ""; $inactive = ""; $ban = ""; $umode = ""; $status = "";
    // Newbie protection
    if($this->newbieProtection && NS::getUser()->userid != $row["userid"])
    {
      $ignoreNP = false;
      if($row["useractivity"] <= time() - 604800)
      {
        $ignoreNP = true;
      }
      else if($row["banid"] && (is_null($row["to"]) || $row["to"] >= time()))
      {
        $ignoreNP = true;
      }

      if($ignoreNP === false)
      {
        $isProtected = isNewbieProtected($this->points, $row["points"]);
        //var_dump($isProtected);
        if($isProtected == 1)
        {
          $status = $ignoreNP == false ? " <span class='weak-player'>n</span> " : "";
          $userClass = "weak-player";
        }
        else if($isProtected == 2)
        {
          $status = $ignoreNP == false ? " <span class='strong-player'>s</span> " : "";
          $userClass = "strong-player";
        }
      }
    }

    // User activity
    if($row["useractivity"] <= time() - 604800)
    {
      $inactive = " i ";
    }
    if($row["useractivity"] <= time() - 1814400)
    {
      $inactive = " I "; // " i I ";
    }

    // Vacation mode
    if($row["umode"])
    {
      $umode = " <span class=\"vacation-mode\">v</span> ";
      $userClass = "vacation-mode";
    }elseif($row["observer"])
    {
      $umode = " <span class=\"vacation-mode\">o</span> ";
      $userClass = "vacation-mode";
    }

    // User banned?
    if($row["banid"] && (is_null($row["to"]) || $row["to"] >= time()))
    {
      $ban = " b ";
      $umode = "";
      $userClass = "banned";
    }

    $row["user_status"] = sprintf("%s%s%s", $inactive, $ban, $umode);
    $row["user_status_long"] = sprintf("%s%s%s%s", $inactive, $ban, $umode, $status);
    $row["username"] = $this->formatUsername($row["username"], $row["userid"], $row["aid"], $userClass);
    $sid = Core::getRequest()->getGET("sid");
    $row["username_link"] = Link::get("game.php/go:MSG/id:ReadFolder/folder:4/ruser:{$row["userid"]}", $row["username"], Core::getLanguage()->getItem("RELATED_MSG"));
    if(!empty($row["aid"]) && !empty($row["tag"]))
    {
      $row["alliance"] = $this->formatAllyTag($row["tag"], $row["name"], $row["aid"]);
      if($this->fetchRankByQuery)
      {
        $row["alliance_rank"] = $this->getAllianceRank($row["aid"], $this->pointType);
      }
    }

    if($row["ismoon"] && !empty($row["moongala"]) && !empty($row["moonsys"]) && !empty($row["moonpos"]))
    {
      $row["position"] = getCoordLink($row["moongala"], $row["moonsys"], $row["moonpos"]);
      $row["position"] .= Core::getLanguage()->getItem("LOON_POST");
    }
    else if(!empty($row["galaxy"]) && !empty($row["system"]) && !empty($row["position"]))
    {
      $row["position"] = getCoordLink($row["galaxy"], $row["system"], $row["position"]);
    }

    // Points
    if($this->fetchRankByQuery)
    {
      $row["rank"] = $this->getUserRank($row[$this->pointType], $this->pointType);
    }
    else
    {
      $row["rank"] = fNumber($row["rank"]);
    }

    foreach(array(
      "dm_points",
      "points",
      "b_points",
      "r_points",
      "u_points",
      "b_count",
      "r_count",
      "u_count",
      "e_points",
      "battles",
      ) as $field)
    {
      $row[$field . "_num"] = $row[$field];
      $row[$field] = pNumber($row[$field], $this->maxDecimals);
    }
    $row["cur_points_num"] = $row[$this->pointType . "_num"];
    $row["cur_points"] = $row[$this->pointType];

    // Hook::event("FORMAT_USERLIST_ROW_LAST", array(&$row, &$this));
    return $row;
  }

  /**
  * Fetches the user rank from database.
  *
  * @param integer	Points
  *
  * @return integer	Rank
  */
  protected function getUserRank($points, $pointType)
  {
    return sqlSelectField("user u", "count(*)", "", $pointType." >= ".sqlVal($points)
            . " AND u.observer = 0 AND u.umode = 0 AND u.userid NOT IN (SELECT userid FROM ".PREFIX."ban_u)");
  }

  /**
  * Sets the CSS class for the username.
  *
  * @param string	Username
  * @param integer	User id
  * @param integer	Alliance id
  * @param string	Default CSS class
  *
  * @return string	Formatted username
  */
  protected function formatUsername($username, $userid, $aid, $defaultClass = "")
  {
    $class = $this->relation->getPlayerRelationClass($userid, $aid);
    $class = (empty($class) && !empty($defaultClass)) ? $defaultClass : $class;
    return (!empty($class)) ? "<span class=\"".$class."\">".$username."</span>" : $username;
  }

  /**
  * Setter-method for newbie protection.
  *
  * @param boolean
  *
  * @return UserList
  */
  public function setNewbieProtection($newbieProtection)
  {
    $this->newbieProtection = $newbieProtection;
    return $this;
  }
}
?>