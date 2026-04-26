<?php
/**
* Program core. Gathers required program parts.
*
* @package Recipe 1.1
* @author Sebastian Noll
* @copyright Copyright (c) 2008, Sebastian Noll
* @license <http://www.gnu.org/licenses/gpl.txt> GNU/GPL
* @version $Id: Core.class.php 23 2010-04-03 19:08:34Z craft $
*/

if(!defined("RECIPE_ROOT_DIR")) { die("Hacking attempt detected."); }

class Core
{
	/**
	* Timer object.
	*/
	protected static $TimerObj;

	/**
	* Database object.
	*
	* @var Database.
	*/
	protected static $DatabaseObj;

	/**
	* Options object.
	*
	* @var Options.
	*/
	protected static $OptionsObj;

	/**
	* Language object.
	*
	* @var Language.
	*/
	protected static $LanguageObj;
	protected static $LanguageStack = array();
	protected static $LanguageHash = array();

	/**
	* Template object.
	*
	* @var Template.
	*/
	protected static $TemplateObj;

	/**
	* Request object.
	*
	* @var Request.
	*/
	protected static $RequestObj;

	/**
	* User object.
	*
	* @var User.
	*/
	protected static $UserObj;

	/**
	* Query parser object.
	*
	* @var QueryParser
	*/
	protected static $QueryObj;

	/**
	* Cron object.
	*
	* @var Cron
	*/
	protected static $CronObj;

	/**
	* Cache object.
	*
	* @var Cache
	*/
	protected static $CacheObj;

	/**
	* Recipe version id.
	*
	* @var Array
	*/
	protected static $version;

	/**
	* The time zone that will be set, if no other can be found.
	*
	* @var string
	*/
	protected static $defaultTimeZone = "Europe/Moscow";

	/**
	* The current timezone.
	*
	* @var string
	*/
	protected static $timezone = null;

	public static $old = true;
	
	/**
	* Constructor. Initializes classes.
	*
	* @return void
	*/
	public function __construct($old = true)
	{
		self::$old = $old;
		self::$version["major"] = 1;
		self::$version["build"] = 1;
		self::$version["revesion"] = 2;
		self::$version["release"] = 112;
		define("RECIPE_VERSION", self::$version["release"]);
		// Hook::event("CORE_START");
		$this->setTimer();
		$this->setDatabase();
		if( !defined('YII_CONSOLE') && self::$old )
		{
			$this->setRequest();
		}
		$this->setQuery();
		$this->setCache();
		$this->setOptions();
		$this->setTemplate();
		if( !defined('YII_CONSOLE') && self::$old )
		{
			$this->setUser();
		}
		$this->checkBans();
		$this->setLanguage();
//		$this->initCron();
		// Hook::event("CORE_FINISHED");
		return;
	}

	/**
	* Initializes the timer class.
	*
	* @return Core
	*/
	protected function setTimer()
	{
		self::$TimerObj = new Timer(MICROTIME);
		return $this;
	}

	/**
	* Returns the timer object.
	*
	* @return Timer
	*/
	public static final function getTimer()
	{
		return self::$TimerObj;
	}

	/**
	* Initializes the request class.
	*
	* @return Core
	*/
	protected function setRequest()
	{
		/*
		try
		{
			if(false)
			{
				
			}
			else
			{
				
			}
			if( !(self::$RequestObj instanceof Request) )
			{
				throw new Exception('!(self::$RequestObj instanceof Request)');
			}
			t('ctrl Request');
		}
		catch (Exception $e)
		{
			error_log('Couldn\'t find Request object in ctrl. Creating Request object in Core class.(' . $e->getMessage() . ')');
			self::$RequestObj = Request::getInstance();
		}
		*/
		self::$RequestObj = Request::getInstance();
		return $this;
	}

	/**
	* Returns the request object.
	*
	* @return Request
	*/
	public static final function getRequest()
	{
		return self::$RequestObj;
	}

	/**
	* Initializes the database.
	*
	* @return Core
	*/
	protected function setDatabase()
	{
		$database = array();
		$database["type"]         = defined('DB_TYPE')   ? DB_TYPE   : 'DB_MYSQL_PDO';
		$database["host"]         = defined('DB_HOST')   ? DB_HOST   : '127.0.0.1';
		$database["user"]         = defined('DB_USER')   ? DB_USER   : '';
		$database["userpw"]       = defined('DB_PWD')    ? DB_PWD    : '';
		$database["databasename"] = defined('DB_NAME')   ? DB_NAME   : '';
		$database["port"]         = defined('DB_PORT')   ? DB_PORT   : null;
		self::$DatabaseObj = new $database["type"]($database["host"], $database["user"], $database["userpw"], $database["databasename"], $database["port"]);
		return $this;
	}

	/**
	* Returns the database object.
	*
	* @return Database
	*/
	public static final function getDatabase()
	{
		return self::$DatabaseObj;
	}

	/**
	* Shorter Version of getDatabase().
	*/
	public static final function getDB()
	{
		return self::getDatabase();
	}

	/**
	* Initializes the query object.
	*
	* @return Core
	*/
	protected function setQuery()
	{
		self::$QueryObj = new QueryParser();
		return $this;
	}

	/**
	* Returns the query object.
	*
	* @return QueryParser
	*/
	public static final function getQuery()
	{
		return self::$QueryObj;
	}

	/**
	* Initializes the options.
	*
	* @return Core
	*/
	protected function setOptions()
	{
		self::$OptionsObj = new Options(false); // false = читать из na_config через прямой SQL (без Yii)
		return $this;
	}

	/**
	* Returns the options object.
	*
	* @return Options
	*/
	public static final function getOptions()
	{
		return self::$OptionsObj;
	}

	/**
	* Alias to getOptions().
	*
	* @return Options
	*/
	public static final function getConfig()
	{
		return self::getOptions();
	}

	/**
	* Initializes the language system.
	*
	* @return Core
	*/
	protected function setLanguage()
	{
		if( !defined('YII_CONSOLE') && self::$old )
		{
			$langid = (self::getRequest()->getGET("lang")) ? self::getRequest()->getGET("lang") : self::getUser()->get("languageid");
			self::$LanguageObj = new Language($langid, 'info, global');
		}
		else
		{
			self::$LanguageObj = new Language(DEF_LANGUAGE_ID, 'info, global');
		}
		self::$LanguageHash[self::$LanguageObj->getOpt("languageid")] = self::$LanguageObj;
		return $this;
	}
	
	public static function pushLanguage()
	{
		array_push(self::$LanguageStack, self::$LanguageObj);
	}
	
	public static function popLanguage()
	{
		self::$LanguageObj = array_pop(self::$LanguageStack);
	}
	
	public static function selectLanguage($lang_id)
	{
		if(self::$LanguageObj->getOpt("languageid") != $lang_id)
		{
			if(isset(self::$LanguageHash[$lang_id]))
			{
				self::$LanguageObj = self::$LanguageHash[$lang_id];
			}
			else
			{
				self::$LanguageHash[$lang_id] = self::$LanguageObj = new Language($lang_id, 'info, global');
			}			
		}
		return self::$LanguageObj;
	}

	/**
	* Returns the language object.
	*
	* @return Language
	*/
	public static final function getLanguage()
	{
		return self::$LanguageObj;
	}

	/**
	* Short form of getLanguage().
	*
	* @return Language
	*/
	public static final function getLang()
	{
		return self::getLanguage();
	}

	/**
	* Initializes the template system.
	*
	* @return Core
	*/
	protected function setTemplate()
	{
		self::$TemplateObj = new Template();
		return $this;
	}

	/**
	* Returns the template object.
	*
	* @return Template
	*/
	public static final function getTemplate()
	{
		return self::$TemplateObj;
	}

	/**
	* Short form of getTemplate().
	*
	* @return Template
	*/
	public static final function getTPL()
	{
		return self::getTemplate();
	}

	/**
	* Initializes the user class.
	*
	* @return Core
	*/
	protected function setUser()
	{
		// JWT lazy join (требует уже инициализированной БД)
		if (class_exists('JwtAuth', false)) {
			JwtAuth::resolveUser();
		}
		if(URL_SESSION)
		{
			$sid = (self::getRequest()->getGET("sid")) ? self::getRequest()->getGET("sid") : "";
		}
		else
		{
			$sid = (self::getRequest()->getCOOKIE("sid")) ? self::getRequest()->getCOOKIE("sid") : "";
		}
		self::$UserObj = new User($sid);

		// План 37.5c: для авторизованного юзера без home planet
		// планируем колонизацию (асинхронно, обработает event-monitor).
		if(!empty($_SESSION["userid"]) && class_exists('OnboardingService', true))
		{
			OnboardingService::ensureColonizationScheduled($_SESSION["userid"]);
		}

		return $this;
	}

	/**
	* Returns the user object.
	*
	* @return User
	*/
	public static final function getUser()
	{
		return self::$UserObj;
	}

	/**
	* Check Database for banned IP address.
	*
	* @return Core
	*/
	protected function checkBans()
	{
		// debug_var(IPADDRESS, "[checkBans] ip: ".IPADDRESS); exit;

		// $where_ip = strstr(IPADDRESS, "*") ? "ipaddress like '".str_replace("*", "%", IPADDRESS)."'" : "ipaddress = '".IPADDRESS."'";
/*		$row = self::getQuery()->selectRow("ban", "reason", "", "'".IPADDRESS."' like ipaddress AND timebegin <= '".time()."' AND (timeend is null OR timeend > '".time()."')");
		if($row)
		{
			header("Content-Type: text/html; charset=utf-8");
			// terminate("IP address ".IPADDRESS." has been banned:\n".$row["reason"]);
			// terminate("Sorry, you has been banned."); // .$row["reason"]);

			echo <<<END
				<?xml version="1.0" encoding="utf-8"?>
				<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd">
				<html xmlns="http://www.w3.org/1999/xhtml" xml:lang="en-US" lang="en-US">
				<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
				<title>Error: Banned!</title>
				</head>
				<body>
				<h1>Sorry, you has been banned.</h1>
				</div>
				</body>
				</html>
END;
			exit;
		}*/
		return $this;
	}

	/**
	* Initializes the Cronjob class.
	*
	* @return Core
	*/
	protected function initCron()
	{
		self::$CronObj = new Cron();
		return $this;
	}

	/**
	* Returns the Cron object.
	*
	* @return Cron
	*/
	public static final function getCron()
	{
		return self::$CronObj;
	}

	/**
	* Initializes the Cache.
	*
	* @return Core
	*/
	protected function setCache()
	{
		self::$CacheObj = new Cache();
		return $this;
	}

	/**
	* Returns the Cache object.
	*/
	public static final function getCache()
	{
		return self::$CacheObj;
	}

	/**
	* Returns the version as a string.
	*
	* @return string Recipe version.
	*/
	public static function versionToString()
	{
		return self::$version["major"].".".self::$version["build"]."rev".self::$version["revesion"];
	}

	/**
	* Sets the default timezone.
	*
	* @param string	Timezone to set
	*
	* @return string	Timezone
	*/
	public static function setDefaultTimeZone($newTimezone = null)
	{
		date_default_timezone_set('Europe/Moscow');
		return date_default_timezone_get();
		if(!is_null(self::$timezone))
		{
			return date_default_timezone_get();
		}

		self::$timezone = $newTimezone;
		
		if(self::getOptions()->exists("timezone"))
		{
			self::$timezone = self::getOptions()->timezone;
		}
		else
		{
			self::$timezone = self::$defaultTimeZone;
		}
		date_default_timezone_set(self::$timezone);
		return date_default_timezone_get();
	}
}
?>