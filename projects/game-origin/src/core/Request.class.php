<?php
/**
* RequestHandler. Parses global variables.
*
* @package Recipe 1.1
* @author Sebastian Noll
* @copyright Copyright (c) 2008, Sebastian Noll
* @license <http://www.gnu.org/licenses/gpl.txt> GNU/GPL
* @version $Id: Request.class.php 23 2010-04-03 19:08:34Z craft $
*/

if(!defined("RECIPE_ROOT_DIR")) { die("Hacking attempt detected."); }

class Request
{
	/**
	* The global _GET variable.
	*
	* @var array
	*/
	protected $global_get = array();

	/**
	* The global _POST variable.
	*
	* @var array
	*/
	protected $global_post = array();

	/**
	* The global _COOKIE variable.
	*
	* @var array
	*/
	protected $global_cookie = array();

	/**
	* Parsed _POST.
	*
	* @var array
	*/
	protected $post = array();

	/**
	* Parsed _GET.
	*
	* @var array
	*/
	protected $get = array();

	/**
	* Parsed _COOKIE.
	*
	* @var array
	*/
	protected $cookie = array();

	/**
	* HTTP File Upload variables.
	*
	* @var array
	*/
	protected $files = array();

	/**
	* Arguments send via URL.
	*
	* @var array
	*/
	protected $args = array();

	/**
	* The complete requested URL
	*
	* @var string
	*/
	protected $requestedURL = "";

	/**
	* URL path level names.
	*
	* @var array
	*/
	protected $levelNames = array();
	
	protected static $instance = null;
	public static function getInstance()
	{
		if(!self::$instance)
		{
			self::$instance = new Request();
		}
		return self::$instance;
	}

	/**
	* Constructor.
	*
	* @return void
	*/
	public function __construct()
	{
		$this->global_get = $_GET;
		$this->global_post = $_POST;
		$this->global_cookie = $_COOKIE;
		$this->files = $_FILES;
	
		if( defined('SN') )
		{
			foreach([] as $name)
			{
				unset($this->global_get[$name]);
			}
			
			/*
			if(SN == ODNK_SN_ID)
			{
				foreach(SocialAPI_Odnoklassniki::getParamNames() as $name)
				{
					unset($this->global_get[$name]);
				}
				/?
				unset($this->global_get['api_server']);
				unset($this->global_get['authorized']);
				unset($this->global_get['application_key']);
				unset($this->global_get['auth_sig']);
				unset($this->global_get['apiconnection']);
				unset($this->global_get['session_key']);
				unset($this->global_get['logged_user_id']);
				unset($this->global_get['sig']);
				unset($this->global_get['session_secret_key']);
				unset($this->global_get['ip_geo_location']);
				unset($this->global_get['clientLog']);
				unset($this->global_get['refplace']);
				unset($this->global_get['web_server']);
				unset($this->global_get['custom_args']);
				unset($this->global_get['referer']);
				?/
			}
			elseif(SN == MAILRU_SN_ID)
			{
				foreach(SocialAPI_Mailru::getParamNames() as $name)
				{
					unset($this->global_get[$name]);
				}
			}
			*/
			
			unset($this->global_get['XDEBUG_SESSION_START']);
			unset($this->global_get['KEY']);

			if( !empty($this->global_get['sn_fullscreen']) )
			{
				define("SN_FULLSCREEN", 1);
			}
			unset($this->global_get['sn_fullscreen']);
		}
		$this->setLevelNames()->setRequests();

		$this->requestedURL = $_SERVER["REQUEST_URI"];
		$defined = get_defined_constants(true);
//		print_r( $defined['user'] );
//		print_r($this->global_get);
		if(count($this->get) == 13 && defined('SN'))
		{
			$this->splitURL()->putArgsIntoRequestVars();
		}
		if(count($this->get) == 0)
		{
			$this->splitURL()->putArgsIntoRequestVars();
		}
		if(
			defined("FORCE_REWRITE") &&
			FORCE_REWRITE &&
			count($this->global_get) > 0
		)
		{
			$this->normalizeURL();
		}
		$this->flushGlobals();
	}

	/**
	* Set request variables.
	*
	* @return Request
	*/
	protected function setRequests()
	{
		if(!function_exists("get_magic_quotes_gpc") || PHP_MAJOR_VERSION >= 7 || get_magic_quotes_gpc() == 0)
		{
			$this->get = $this->parseArray($this->global_get);
			$this->post = $this->parseArray($this->global_post);
			$this->cookie = $this->parseArray($this->global_cookie);
		}
		else
		{
			$this->get = $this->global_get;
			$this->post = $this->global_post;
			$this->cookie = $this->global_cookie;
		}
		return $this;
	}

	/**
	* Pass requset array and perform escaping function on each value.
	*
	* @param array		The request array
	*
	* @return array	Escaped array
	*/
	protected function parseArray($array)
	{
		if(!is_array($array)) { return false; }
		/*
		foreach($array as $key => $val)
		{
			if(is_array($array[$key]))
			{
				$array[$key] = $this->parseArray($array[$key]);
			}
			else
			{
				// $array[$key] = Core::getDatabase()->quote_db_value($array[$key]);
				$array[$key] = $array[$key];
			}
		}
		*/
		return $array;
	}

	/**
	* Clear global arrays and deallocate disk space.
	*
	* @return Request
	*/
	protected function flushGlobals()
	{
		unset($this->global_get);
		unset($this->global_post);
		unset($this->global_cookie);
		if(!defined("FORCE_REQUEST_CLEAR") || FORCE_REQUEST_CLEAR)
		{
//			unset($_GET);
//			unset($_POST);
//			//unset($_COOKIE);
//			unset($_REQUEST);
		}
		return $this;
	}

	//### All functions below serve the internal rewrite engine ###//

	/**
	* Split the URL into arguments.
	*
	* @return Request
	*/
	protected function splitURL()
	{
		$url = "http://".$_SERVER["SERVER_NAME"].$this->requestedURL;
		if(0) // Core::getUser()->get("userid") == 2)
		{
			echo "requestedURL: <pre>";
			print_r($url);
			echo "</pre>";
		}
		$splitted = false;
		try
		{
			$splitted = @parse_url($url);
		}
		catch(\Throwable $e)
		{
			error_log('Couldn\'t parse_url: ' . $url);
		}

		if(!$splitted)
		{
			/* try
			{
			$splitted = @parse_url(Str::substring($this->requestedURL, 1));
			}
			catch(Exception $e)
			{
			// echo "catched2: " . $e->getMessage();
			} */

			if(!$splitted)
			{
				/* if(preg_match("#^https?://[^/]+([^\?]+)$#is", $this->requestedURL, $regs))
				{
				$splitted["path"] = $regs[1];
				}
				else */
				{
					Logger::dieMessage("Error request: " . $this->requestedURL);
				}
			}
		}

		if(0)
		{
			echo "splitted: <pre>";
			print_r($splitted);
			echo "</pre>";
		}

		// Remove real path from virtual path.
		$path = Str::substring(Str::replace($_SERVER["SCRIPT_NAME"], "", $splitted["path"]), 1);

		if(Str::length($path) > 0)
		{
			// Hook::event("SPLIT_URL_START", array(&$this, &$path));
			$path = explode("/", $path);
			for($i = 0; $i < count($path); $i++)
			{
				$pos = strpos($path[$i], ":");
				if($pos > 0)
				{
					$path[$i] = Str::replace(" ", "_", $path[$i]);
					$path[$i] = array( substr($path[$i], 0, $pos), substr($path[$i], $pos+1) );
					// $path[$i] = explode(":", Str::replace(" ", "_", $path[$i]));
					$path[$i] = Arr::trimArray($path[$i]);
				}
			}
			$this->args = $path;
			// Hook::event("SPLIT_URL_CLOSE", array(&$this, &$path));
		}
		if(0)
		{
			echo "args: <pre>";
			print_r($this->args);
			echo "</pre>";
		}
		return $this;
	}

	/**
	* This function puts the arguments into the
	* respective request variables.
	* http://www.anypage.com/GoArgument/SpecificArgument:Content
	* get["go"] => GoArgument, get["SpecificArgument"] => Content.
	* If the argument has no stated name, the function will use level
	* names (These can be changed by class variable levelNames).
	* [Note that you have to mind the right order!])
	*
	* @return Request
	*/
	protected function putArgsIntoRequestVars()
	{
		for($i = 0; $i < count($this->args); $i++)
		{
			if(is_array($this->args[$i]))
			{
				$this->get[$this->args[$i][0]] = $this->args[$i][1];
			}
			else
			{
				if(Str::length($this->levelNames[$i]) > 0)
				{
					$this->get[$this->levelNames[$i]] = $this->args[$i];
				}
				else
				{
					$this->get[$this->args[$i]] = $this->args[$i];
				}
			}
		}
		// Hook::event("ARGS_INTO_REQUEST_VARS", array(&$this));
		return $this;
	}

	/**
	* Put the global get arguments into rewrite url.
	*
	* @return Request
	*/
	protected function normalizeURL()
	{
		$url = HTTP_HOST.Str::substring($_SERVER["PHP_SELF"], 1);
		$i = 0;
		$exeptions = array(
			'debug_host',
			'start_debug',
			'debug_port',
			'original_url',
			'send_sess_end',
			'debug_stop',
			'debug_start_session',
			'debug_no_cache',
			'debug_session_id',
			'api_server',
			'authorized',
			'application_key',
			'auth_sig',
			'apiconnection',
			'session_key',
			'logged_user_id',
			'sig',
			'session_secret_key',
			'ip_geo_location',
			'clientLog',
			'refplace',
			'web_server',
			'sn_fullscreen',
			'custom_args',
			'referer',
			'KEY',
			'XDEBUG_SESSION_START',
		);
		$get_parts = array();
		foreach($this->get as $key => $val)
		{
			if( Str::compare($key, $this->levelNames[$i], true) && !in_array($key, $exeptions))
			{
				$url .= "/".$val;
			}
			elseif ( !in_array($key, $exeptions) )
			{
				$url .= "/".$key.":".$val; // urlencode($val);
			}
			else
			{
				$get_parts[] = $key.'='.$val;
			}
			$i++;
		}
		$url .= ( !empty($get_parts) ) ? ( '?' . implode('&', $get_parts)) : '';
		error_log($url);
		doHeaderRedirection($url, true);
		return $this;
	}

	/**
	* Returns the value of the given request type parameter.
	*
	* @param string	Request type (get, post or cookie)
	* @param string	Parameter
	*
	* @return mixed	The value or false
	*/
	public function getArgument($requestType, $param)
	{
		$requestType = strtolower($requestType);
		switch($requestType)
		{
		case "get":
			return $this->getGET($param);
			break;
		case "post":
			return $this->getPOST($param);
			break;
		case "cookie":
			return $this->getCOOKIE($param);
			break;
		case "files":
			return $this->getFILES($param);
			break;
		}
		return false;
	}

	/**
	* Returns the parameter of the GET query.
	*
	* @param string	Parameter
	*
	* @return mixed	The value or false
	*/
	public function getGET($param = null, $default = false)
	{
		if(is_null($param))
		{
			return $this->get;
		}
		return (isset($this->get[$param])) ? $this->get[$param] : $default;
	}

	/**
	* Returns the parameter of the POST query.
	*
	* @param string	Parameter
	*
	* @return mixed	The value or false
	*/
	public function getPOST($param = null, $default = false)
	{
		if(is_null($param))
		{
			return $this->post;
		}
		return (isset($this->post[$param])) ? $this->post[$param] : $default;
	}

	/**
	* Returns the parameter of the COOKIE query.
	*
	* @param string	Parameter
	*
	* @return mixed	The value or false
	*/
	public function getCOOKIE($param = null, $default = false)
	{
		if(is_null($param))
		{
			return $this->cookie;
		}
		return (isset($this->cookie[COOKIE_PREFIX.$param])) ? $this->cookie[COOKIE_PREFIX.$param] : $default;
	}

	/**
	* Returns uploaded items of the FILES query.
	*
	* @param string	Upload id
	*
	* @return mixed	The file upload data, false otherwise
	*/
	public function getFILES($fileid = null, $default = false)
	{
		if(is_null($param))
		{
			return $this->files;
		}
		return (isset($this->files[$fileid])) ? $this->files[$fileid] : $default;
	}

	/**
	* Returns an HTTP-parameter.
	*
	* @param string	HTTP-type
	* @param string	Parameter
	*
	* @return mixed	The value or false
	*/
	public function get($http, $param)
	{
		return (isset($this->{$http}[$param])) ? $this->{$http}[$param] : false;
	}

	/**
	* Returns the requested URL.
	*
	* @return string
	*/
	public function getRequestedUrl()
	{
		return $this->requestedURL;
	}

	/**
	* Sets the request parameter sequence.
	*
	* @param array		Parameter sequence [optional]
	*
	* @return Request
	*/
	public function setLevelNames(array $levelNames = null)
	{
		if(is_null($levelNames) && defined("REQUEST_LEVEL_NAMES"))
		{
			$levelNames = explode(",", REQUEST_LEVEL_NAMES);
		}
		$this->levelNames = $levelNames;
		return $this;
	}

	/**
	* Sets a new cookie.
	*
	* @param string	Cookie name
	* @param string	Cookie value
	* @param integer	Cookie expires in $ days
	*
	* @return Request
	*/
	public function setCookie($cookie, $value, $expires = 365)
	{
		$domain = ($_SERVER["SERVER_NAME"] != "localhost") ? Str::replace("www", "", $_SERVER["SERVER_NAME"]) : null;
		setcookie(COOKIE_PREFIX.$cookie, $value, time() + 86400 * $expires, "/", $domain);
		return $this;
	}
}
?>