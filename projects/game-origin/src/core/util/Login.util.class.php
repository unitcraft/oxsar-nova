<?php
/**
* Handle Login.
*
* @package Recipe 1.1
* @author Sebastian Noll
* @copyright Copyright (c) 2008, Sebastian Noll
* @license <http://www.gnu.org/licenses/gpl.txt> GNU/GPL
* @version $Id: Login.util.class.php 23 2010-04-03 19:08:34Z craft $
*/

if(!defined("RECIPE_ROOT_DIR")) { die("Hacking attempt detected."); }

class Login
{
  /**
  * Username.
  *
  * @var string
  */
  protected $usr = "";

  /**
  * Password as MD5 Hash.
  *
  * @var string
  */
  protected $pw = "";

  /**
  * Unencrypted password.
  *
  * @var string
  */
  protected $rawPassword = "";

  /**
  * User Id.
  */
  protected $userid = 0;

  /**
  * Session Id.
  */
  protected $sid = "";

  /**
  * Session Id length.
  *
  * @var integer
  */
  const SESSION_LENGTH = 12;

  /**
  * Days to cookie session expire.
  *
  * @var integer
  */
  const COOKIE_SESSION_EXPIRE = 365;

  /**
  * Redirection url.
  *
  * @var string
  */
  protected $redirection = "";

  /**
  * Can login variable.
  *
  * @var boolean
  */
  protected $canLogin = false;

  /**
  * Maximum counter to login.
  *
  * @var integer
  */
  protected $maxLoginAttempts = 5;

  /**
  * Wating time for renew login (in minutes).
  *
  * @var integer
  */
  protected $bannedLoginTime = 5;

  /**
  * Encryption name.
  *
  * @var string
  */
  protected $encryption = "md5";

  /**
  * Enables cache operation.
  *
  * @var boolean
  */
  protected $cacheActive = false;

  /**
  * Stores login failures.
  *
  * @var Map
  */
  protected $errors = null;

  /**
  * Redirection after login failed.
  *
  * @var boolean
  */
  protected $redirectOnFailure = true;

  /**
  * Redirection after succeeding login.
  *
  * @var boolean
  */
  protected $redirectOnSuccess = true;

  /**
  * Url to redirect after succeeding login.
  *
  * @var string
  */
  protected $sessionUrl = "";

  /**
  * Flag for data check.
  *
  * @var boolean
  */
  protected $dataChecked = false;

  /**
  * Count login attempts.
  *
  * @var boolean
  */
  protected $countLoginAttempts = true;

  /**
  * Constructor.
  *
  * @param string	Username
  * @param string	Password
  * @param string	Redirection after login
  * @param string	Password encryption
  *
  * @return void
  */
  public function __construct($usr, $pw, $redirection = "", $encryption = "md5")
  {
//    if(!$this->cacheActive) { $this->cacheActive = CACHE_ACTIVE; }
    $this->setMaxLoginAttempts()->setBannedLoginTime();
    $this->errors = new Map();
    $this->setUsername($usr)
      ->setEncryption($encryption)
      ->setPassword($pw)
      ->setRedirection($redirection);
    return;
  }

  /**
  * Count login attempts.
  *
  * @return Login
  */
  public function checkLoginAttempts()
  {
    $select = array("COUNT(*) AS hits", "MAX(time) AS lastattempt");
    $result = Core::getQuery()->select("loginattempts", $select, "", "ip = ".sqlVal(IPADDRESS)." OR username = ".sqlVal($this->usr));
    $loginattempts = Core::getDB()->fetch($result);
    Core::getDatabase()->free_result($result);
    if($loginattempts["hits"] >= $this->maxLoginAttempts && $loginattempts["lastattempt"] > time() - ($this->bannedLoginTime * 60))
    {
      $this->loginFailed("MAXIMUM_LOGIN_ATTEMPTS_REACHED");
    }
    return $this;
  }

  /**
  * Check input data.
  *
  * @return Login
  */
  public function checkData()
  {
    $this->dataChecked = true;
    $select = array("u.userid", "u.username", "p.password");
    $joins  = "LEFT JOIN ".PREFIX."password p ON (u.userid = p.userid)";
    $result = Core::getQuery()->select("user u", $select, $joins, "u.username = ".sqlVal($this->usr));
    if($row = Core::getDB()->fetch($result))
    {
      Core::getDatabase()->free_result($result);
      if(Str::compare($row["username"], $this->usr) && Str::compare($row["password"], $this->pw))
      {
        $this->userid = $row["userid"];
        Core::getQuery()->delete("loginattempts", "ip = ".sqlVal(IPADDRESS)." OR username = ".sqlVal($this->usr));
        Core::getQuery()->update("sessions", "logged", "0", "userid = ".sqlVal($this->userid).  ' ORDER BY sessionid ');
        $this->canLogin = true;
      }
      else
      {
        $this->canLogin = false;
        if(!Str::compare($row["username"], $this->usr)) { $this->loginFailed("USERNAME_DOES_NOT_EXIST"); }
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
    if(!$this->dataChecked)
    {
      $this->checkData();
    }
    // Disables old sessions.
    if($this->cacheActive) { Core::getCache()->cleanUserCache($this->userid); }
    Core::getQuery()->update("sessions", "logged", 0, "userid = ".sqlVal($this->userid).  ' ORDER BY sessionid ');

    // Start new session.
    $sessionSeed = Str::encode((string) microtime(1), $this->encryption);
    $this->sid = Str::Substring($sessionSeed, 0, self::SESSION_LENGTH);
    unset($sessionSeed);
    $att = array("sessionid", "userid", "ipaddress", "useragent", "time", "logged");
    $value = array($this->sid, $this->userid, IPADDRESS, $_SERVER["HTTP_USER_AGENT"], time(), "1");
    Core::getQuery()->insert("sessions", $att, $value);
    if($this->canLogin)
    {
      if(COOKIE_SESSION)
      {
        Core::getRequest()->setCookie("sid", $this->sid, self::COOKIE_SESSION_EXPIRE);
        $this->sessionUrl = $this->redirection;
      }
      else
      {
        if(Str::inString("?", $this->redirection))
        {
          $this->sessionUrl = $this->redirection."&sid=".$this->sid;
        }
        else
        {
          $this->sessionUrl = $this->redirection."?sid=".$this->sid;
        }
      }
//      if($this->cacheActive) { Core::getCache()->buildUserCache($this->sid); }
      // Hook::event("START_SESSION", array(&$this, $this->sessionUrl));
      if($this->redirectOnSuccess) { doHeaderRedirection($this->sessionUrl); }
    }
    else
    {
      $this->loginFailed("CANNOT_LOGIN");
    }
    return $this;
  }

  /**
  * Count login attempt up and redirect to login site.
  *
  * @param string	Error message id
  *
  * @return Login
  */
  protected function loginFailed($errorid)
  {
    $this->errors->push($errorid);
    if($this->countLoginAttempts)
    {
      $att = array("time", "ip", "username");
      $value = array(time(), IPADDRESS, $this->usr);
      Core::getQuery()->insert("loginattempts", $att, $value);
    }
    if($this->redirectOnFailure) { forwardToLogin($errorid); }
    return $this;
  }

  /**
  * Setter-method for password encryption mode.
  *
  * @param string
  *
  * @return Login
  */
  public function setEncryption($encryption)
  {
    $this->encryption = $encryption;
    $this->pw = Str::encode($this->rawPassword, $this->encryption);
    return $this;
  }

  /**
  * Setter-method for username.
  *
  * @param string
  *
  * @return Login
  */
  public function setUsername($username)
  {
    $this->usr = $username;
    return $this;
  }

  /**
  * Setter-method for password.
  *
  * @param string
  *
  * @return Login
  */
  public function setPassword($password)
  {
    $this->rawPassword = $password;
    $this->pw = Str::encode($password, $this->encryption);
    return $this;
  }

  /**
  * Setter-method for login redirection.
  *
  * @param string
  *
  * @return Login
  */
  public function setRedirection($redirection)
  {
    $this->redirection = $redirection;
    return $this;
  }

  /**
  * Set redirect on failure.
  *
  * @param boolean
  *
  * @return Login
  */
  public function setRedirectOnFailure($redirectOnFailure)
  {
    $this->redirectOnFailure = $redirectOnFailure;
    return $this;
  }

  /**
  * Set redirect on success.
  *
  * @param boolean
  *
  * @return Login
  */
  public function setRedirectOnSuccess($redirectOnSuccess)
  {
    $this->redirectOnSuccess = $redirectOnSuccess;
    return $this;
  }

  /**
  * Returns the password encryption mode.
  *
  * @return string
  */
  public function getEncryption()
  {
    return $this->encryption;
  }

  /**
  * Returns the username.
  *
  * @return string
  */
  public function getUsername()
  {
    return $this->usr;
  }

  /**
  * Returns the encrypted password.
  *
  * @return string
  */
  public function getPassword()
  {
    return $this->pw;
  }

  /**
  * Returns the login redirection url.
  *
  * @return string
  */
  public function getRedirection()
  {
    return $this->redirection;
  }

  /**
  * Returns the user id.
  *
  * @return integer
  */
  public function getUserId()
  {
    return $this->userid;
  }

  /**
  * Returns login status.
  *
  * @return boolean
  */
  public function getCanLogin()
  {
    return $this->canLogin;
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
  * Returns the session url.
  *
  * @return string
  */
  public function getSessionUrl()
  {
    return $this->sessionUrl;
  }

  /**
  * Returns the errors that has been occured during login.
  *
  * @return Map
  */
  public function getErrors()
  {
    return $this->errors;
  }

  /**
  * Setter-method for max. login attempts.
  *
  * @param integer
  *
  * @return Login
  */
  public function setMaxLoginAttempts($maxLoginAttempts = null)
  {
    $this->maxLoginAttempts = $maxLoginAttempts;
    if(is_null($this->maxLoginAttempts) && Core::getOptions()->exists("maxloginattempts"))
    {
      $this->maxLoginAttempts = Core::getOptions()->get("maxloginattempts");
    }
    if(!is_numeric($this->maxLoginAttempts))
    {
      $this->setCountLoginAttempts(false);
    }
    return $this;
  }

  /**
  * Setter-method for banning login time (in minutes).
  *
  * @param integer
  *
  * @return Login
  */
  public function setBannedLoginTime($bannedLoginTime = null)
  {
    $this->bannedLoginTime = $bannedLoginTime;
    if(is_null($this->bannedLoginTime) && Core::getOptions()->exists("bannedlogintime"))
    {
      $this->bannedLoginTime = Core::getOptions()->get("bannedlogintime");
    }
    if(!is_numeric($this->bannedLoginTime))
    {
      $this->setCountLoginAttempts(false);
    }
    return $this;
  }

  /**
  * Setter-method for counting login attempts.
  *
  * @param boolean
  *
  * @return Login
  */
  public function setCountLoginAttempts($countLoginAttempts)
  {
    $this->countLoginAttempts = $countLoginAttempts;
    return $this;
  }
}
?>