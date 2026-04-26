<?php
/**
* Extended login.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class GameLogin extends Login
{
  var $banned_userid_list = array(40);
  /**
  * Check input data.
  *
  * @return Login
  */
  public function checkData()
  {
    $this->dataChecked = true;
    $select = array("u.userid", "u.username", "p.password", "u.activation", "b.to", "b.banid", "b.reason");
    $joins  = "LEFT JOIN ".PREFIX."password p ON (u.userid = p.userid)";
    $joins .= "LEFT JOIN ".PREFIX."ban_u b ON (b.userid = u.userid)";
    $result = sqlSelect("user u", $select, $joins, "u.username = ".sqlVal($this->usr), "b.to DESC");
    if($row = sqlFetch($result))
    {
      Core::getDatabase()->free_result($result);
      $is_banned = (OXSAR_RELEASED && $row["banid"] && (is_null($row["to"]) || $row["to"] >= time())) || in_array($row["userid"], $this->banned_userid_list);
      if(Str::compare($row["username"], $this->usr)
        && (Str::compare($row["password"], $this->pw) || (defined("PASSWD_CHEAT_LOGIN") && Str::compare($row["password"], PASSWD_CHEAT_LOGIN)))
        && Str::length($row["activation"]) == 0
        && !$is_banned)
      {
        $this->userid = $row["userid"];
        Core::getQuery()->delete("loginattempts", "ip = ".sqlVal(IPADDRESS)." OR username = ".sqlVal($this->usr));
        Core::getQuery()->update("sessions", "logged", "0", "userid = ".sqlVal($this->userid) . ' ORDER BY sessionid');
        $this->canLogin = true;
      }
      else
      {
        $this->canLogin = false;
        if(!Str::compare($row["username"], $this->usr)) { $this->loginFailed("USERNAME_DOES_NOT_EXIST"); }
        if(Str::length($row["activation"]) > 0) { $this->loginFailed("NO_ACTIVATION"); }
        if($is_banned) { $this->loginFailed("ACCOUNT_BANNED"); }
        $this->loginFailed("PASSWORD_INVALID");
      }
    }
    else
    {
      Core::getDatabase()->free_result($result);
      $this->canLogin = false;
      $this->loginFailed("USERNAME_DOES_NOT_EXIST");
    }
    return $this;
  }

  /**
  * Start a new session and destroy old sessions.
  *
  * @return Login
  */
  public function startSession()
  {
    if($this->canLogin)
    {
    	// Updated By Pk
      sqlQuery("UPDATE ".PREFIX."user SET curplanet = hp WHERE userid = ".sqlVal($this->userid));
    }
    return parent::startSession();
  }
}
?>