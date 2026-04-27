<?php
/**
* Generic functions libary.
*
* @package Recipe 1.1
* @author Sebastian Noll
* @copyright Copyright (c) 2008, Sebastian Noll
* @license <http://www.gnu.org/licenses/gpl.txt> GNU/GPL
* @version $Id: Functions.php 23 2010-04-03 19:08:34Z craft $
*/

if(!defined("RECIPE_ROOT_DIR")) { die("Hacking attempt detected."); }

/**
* Displays string and shut program down.
* (Improved function of die().)
*
* @param string	The to displayed string.
*
* @return void
*/
function terminate($string)
{
  if(is_array($string)) { print_r($string); }
  else { echo $string; }
  exit;
}

/**
* Forward to login page.
*
* @param string	Error id to output
*
* @return void
*/
function forwardToLogin($errorid)
{
  if(LOGIN_REQUIRED)
  {
    if(strpos(LOGIN_URL, '?') === false)
    {
      $login = LOGIN_URL."?error=".$errorid;
    }
    else { $login = LOGIN_URL; }
    // Hook::event("FORWARD_TO_LOGIN_PAGE", array(&$login, $errorid));
    doHeaderRedirection($login, false);
  }
  Core::getLanguage()->load("account");
  Logger::addMessage($errorid);
  Core::getTPL()->display("login");
  return;
}

/**
* Perform an header redirection.
*
* @param string	URL
*
* @return void
*/

// Stub: socialUrl был для соц.сетей (ОК/VK iframe). В oxsar-nova OAuth убран.
if (!function_exists('socialUrl')) {
  function socialUrl($url) { return $url; }
}

// План 37.5d.9: artImageUrl — генерация URL для preview артефакт-картинок.
// В legacy указывала на index.php/artefact2user_YII?{suffix}.
// У нас Yii нет — endpoint реализован в public/artefact-image.php (GD-генерация
// idential-логики из legacy controller renderImage()).
// $action = "image_new" | "image" (action_view), у нас оба → один endpoint.
// $suffix = строка query-params от social API (мы её игнорируем).
if (!function_exists('artImageUrl')) {
  function artImageUrl($action, $suffix, $real_url = true) {
    return ($real_url ? RELATIVE_URL : "") . "artefact-image.php?";
  }
}

function doHeaderRedirection($url, $appendSession = true)
{
  // $test =  strpos("http://", $url);
   if(Link::isExternal($url) || strpos("http://", $url) !== false || substr($url, 0, 7) == 'http://')
  //if(Link::isExternal($url) || strpos("http://", $url) !== false)
  {
    $path = $url;
  }
  else
  {
  	// If need sid, but dont have it.
    if( $appendSession && strpos($url, "sid=") === false && URL_SESSION )
    {
      (strpos($url, "?") === false) ? $url .= "?sid=".SID : $url .= "&sid=".SID;
    }

    if( defined("FORCE_REWRITE") && FORCE_REWRITE )
    {
    	$path = Link::normalizeURL($url);
    }
    else
    {
    	$path = HTTP_HOST.REQUEST_DIR.$url;
    }
  }
  if( defined('SN') && strpos($path, '?') === false && strpos($path, 'api_server') === false )
  {
	$path = socialUrl($path);
	/*
  	if( strpos($path, '?') === false )
  	{
  		$path .= '?';
  	}
  	else
  	{
  		$path .= '&';
  	}
  	$path .= '';
	*/
  	if( !preg_match("#^(".preg_quote(FULL_URL, "#")."|".preg_quote(RELATIVE_URL, "#").")#is", $path) )
  	{
  		$path = RELATIVE_URL . $path;
  	}
  }
  header("Location: ".$path);
  exit();
}

/**
* Checks whether the incoming email address is valid.
*
* @param string	Email address to check
*
* @return boolean
*/
function isMail($mail)
{
  if(preg_match("#^[^\\x00-\\x1f@]+@[^\\x00-\\x1f@]{2,}\.[a-z]{2,}$#i", $mail) == 0)
  {
    return false;
  }
  return true;
}

/**
* Generates a random text.
*
* @param integer	The length of the random text
*
* @return string	The random text
*/
function randString($length)
{
  $pool = "qwertzupasdfghkyxcvbnm";
  $pool .= "23456789";
  $pool .= "QWERTZUPLKJHGFDSAYXCVBNM";
  srand ((double)microtime()*1000000);
  for($index = 0; $index < $length; $index++)
  {
    $randstr .= substr($pool,(mt_rand()%(strlen($pool))), 1);
  }
  return $randstr;
}

/**
* Parses an URL and return its components.
*
* @param string	The URL to parse
*
* @return array	The URL components
*/
function parseUrl($url)
{
  $out = array();
  $r  = "^(?:(?P<scheme>\w+)://)?";
  $r .= "(?:(?P<login>\w+):(?P<pass>\w+)@)?";
  $r .= "(?P<host>(?:(?P<subdomain>[\w\.]+)\.)?" . "(?P<domain>\w+\.(?P<extension>\w+)))";
  $r .= "(?::(?P<port>\d+))?";
  $r .= "(?P<path>[\w/]*/(?P<file>\w+(?:\.\w+)?)?)?";
  $r .= "(?:\?(?P<arg>[\w=&]+))?";
  $r .= "(?:#(?P<anchor>\w+))?";
  $r = "!$r!";
  preg_match($r, $url, $out);
  return $out;
}

/**
* Capitalizes the first letter of each directory part.
*
* @param string	Path
* @param char		Path separator [optional]
*
* @return string
*/
function getClassPath($path, $s = '/')
{
  if(preg_match("#".$s."$#i", $path))
  {
    $path = substr($path, 0, -1);
  }
  if(preg_match("#^".$s."#i", $path))
  {
    $path = substr($path, 1);
  }
  return str_replace(' ', $s, ucwords(str_replace($s, ' ', $path)));
}

function sqlVal($data)
{
  return is_null($data) ? "NULL" : Core::getDatabase()->quote_db_value((string)$data);
}

function sqlArray()
{
  $data = func_num_args() > 1 ? func_get_args() : func_get_arg(0);
  if(is_array($data))
  {
    $result = array();
    foreach($data as $value)
    {
      $result[] = sqlVal($value);
    }
    return count($result) ? implode(",", $result) : 'NULL';
  }
  return sqlVal($data);
}

function sqlSelect($table, $select, $join = "", $where = "", $order = "", $limit = "", $groupby = "", $other = "")
{
  return Core::getQuery()->select($table, $select, $join, $where, $order, $limit, $groupby, $other);
}

function sqlSelectRow($table, $select, $join = "", $where = "", $order = "", $limit = "", $groupby = "", $other = "")
{
  return Core::getQuery()->selectRow($table, $select, $join, $where, $order, $limit, $groupby, $other);
}

function sqlSelectField($table, $select, $join = "", $where = "", $order = "", $limit = "", $groupby = "", $other = "")
{
  return Core::getQuery()->selectField($table, $select, $join, $where, $order, $limit, $groupby, $other);
}

function sqlInsert($table_name, $values)
{
  Core::getQuery()->insert($table_name,
    array_keys($values),
    array_values($values));
  return Core::getDB()->insert_id();
}

function sqlUpdate($table_name, $values, $where)
{
  return Core::getQuery()->update($table_name,
    array_keys($values),
    array_values($values),
    $where);
}

function sqlDelete($table_name, $where)
{
  Core::getQuery()->delete($table_name, $where);
}

function sqlQuery($sql)
{
  return Core::getDB()->query($sql);
}

function sqlQueryRow($sql)
{
  return Core::getDB()->queryRow($sql);
}

function sqlQueryField($sql)
{
  return Core::getDB()->queryField($sql);
}

function sqlFetch($result)
{
  return Core::getDB()->fetch($result);
}

function sqlEnd($result)
{
  return Core::getDB()->free_result($result);
}

/**
* Gets an environment variable from available sources, and provides emulation
* for unsupported or inconsistent environment variables (i.e. DOCUMENT_ROOT on
* IIS, or SCRIPT_NAME in CGI mode).  Also exposes some additional custom
* environment information.
*
* @param  string $key Environment variable name.
* @return string Environment variable setting.
* @link http://book.cakephp.org/view/701/env
*/
function env($key) {
  if ($key == 'HTTPS') {
    if (isset($_SERVER['HTTPS'])) {
      return (!empty($_SERVER['HTTPS']) && $_SERVER['HTTPS'] !== 'off');
    }
    return (strpos(env('SCRIPT_URI'), 'https://') === 0);
  }

  if ($key == 'SCRIPT_NAME') {
    if (env('CGI_MODE') && isset($_ENV['SCRIPT_URL'])) {
      $key = 'SCRIPT_URL';
    }
  }

  $val = null;
  if (isset($_SERVER[$key])) {
    $val = $_SERVER[$key];
  } elseif (isset($_ENV[$key])) {
    $val = $_ENV[$key];
  } elseif (getenv($key) !== false) {
    $val = getenv($key);
  }

  if ($key === 'REMOTE_ADDR' && $val === env('SERVER_ADDR')) {
    $addr = env('HTTP_PC_REMOTE_ADDR');
    if ($addr !== null) {
      $val = $addr;
    }
  }

  if ($val !== null) {
    return $val;
  }

  switch ($key) {
      case 'SCRIPT_FILENAME':
        if (defined('SERVER_IIS') && SERVER_IIS === true) {
          return str_replace('\\\\', '\\', env('PATH_TRANSLATED'));
        }
        break;
      case 'DOCUMENT_ROOT':
        $name = env('SCRIPT_NAME');
        $filename = env('SCRIPT_FILENAME');
        $offset = 0;
        if (!strpos($name, '.php')) {
          $offset = 4;
        }
        return substr($filename, 0, strlen($filename) - (strlen($name) + $offset));
        break;
      case 'PHP_SELF':
        return str_replace(env('DOCUMENT_ROOT'), '', env('SCRIPT_FILENAME'));
        break;
      case 'CGI_MODE':
        return (PHP_SAPI === 'cgi');
        break;
      case 'HTTP_BASE':
        $host = env('HTTP_HOST');
        if (substr_count($host, '.') !== 1) {
          return preg_replace('/^([^.])*/i', null, env('HTTP_HOST'));
        }
        return '.' . $host;
        break;
  }
  return null;
}

function is_mobile_request()
{
  return preg_match('/' . REQUEST_MOBILE_UA . '/i', env('HTTP_USER_AGENT'));
}

define('REQUEST_MOBILE_UA', '(iPhone|MIDP|AvantGo|BlackBerry|J2ME|Opera Mini|DoCoMo|NetFront|Nokia|PalmOS|PalmSource|portalmmm|Plucker|ReqwirelessWeb|SonyEricsson|Symbian|UP\.Browser|Windows CE|Xiino)');
define('IS_MOBILE_REQUEST', is_mobile_request());

?>
