<?php
/**
* Abstract class for pages.
*
* Oxsar http://oxsar.ru
*
*
*/

abstract class Page
{
	/**
	* Holds all GET-actions.
	*
	* @var array
	*/
	private $getActions = array();

	/**
	* Holds all POST-actions.
	*
	* @var Map
	*/
	private $postActions = null;

	/**
	* Holds the arguments which will be passed.
	*
	* @var array
	*/
	private $getArgs = array(), $postArgs = array(); private $args = array();

	/**
	* If an action has been called.
	*
	* @var boolean
	*/
	protected $actionCalled = false;

	/**
	* Global module constructor.
	*
	* @return void
	*/
	protected function __construct()
	{
		$this->postActions = new Map();
		Core::getTPL()->setView($this);
	}

	/**
	* Proceeds the POST and GET actions.
	*
	* @return Page
	*/
	protected function proceedRequest()
	{
		///********* *********///
		Core::getTPL()->addHTMLHeaderFile("lib/jquery.iterator.js?".CLIENT_VERSION, "js");

		$current = array();
		$storage = array();
		$production = array();
		foreach(array("metal", "silicon", "hydrogen") as $key)
		{
			$current[$key]		= NS::getPlanet()->getData($key);
			$storage[$key]		= NS::getPlanet()->getStorage($key);
			$production[$key]	= NS::getPlanet()->getProd($key);

			Core::getTPL()->assign("stor_" . $key,
				"<span class='" . ($storage[$key] > $current[$key] ? "true" : "false") . "'>" . fNumber($storage[$key] / 1000)."k</span>");

			if($storage[$key] > $current[$key] || $production[$key] < 0)
			{
				$span_str = "<script type='text/javascript'>
					//<![CDATA[
					$(function($) {
						var options = {
							startNum: {$current[$key]},
							stopNum: ".($production[$key] < 0 ? 0 : $storage[$key]).",
							step: {$production[$key]}/3600.0
						}
						$('.iter_{$key}').iterator(options);
					});
					//]]>
					</script>
					<span class='iter_{$key}".($production[$key] < 0 ? " false" : "")."'>".fNumber($current[$key])."</span>";
			}
			else
			{
				$span_str = fNumber($current[$key]);
			}
			Core::getTPL()->assign("real_" . $key, $span_str);
		}

		Core::getTPL()->assign("credit", fNumber(NS::getUser()->get("credit"), 2));

        if(NS::getUser()->get("observer") && OBSERVER_OFF_CREDIT_COST > 0 && NS::getUser()->get("credit") >= OBSERVER_OFF_CREDIT_COST){
            NS::getUser()->set("observer", 0);
            if(PROTECTION_PERIOD > 0){
                NS::getUser()->set("protection_time", time() + PROTECTION_PERIOD);
            }
        }

		if( NS::getPlanet() && !NS::getPlanet()->getPlanetId() && isHomePlanetRequiredForPage(Core::getRequest()->getGET("go")) )
		{
			doHeaderRedirection("game.php/HomePlanetRequired", false);
		}

        $isAjaxRequest = (!empty($_SERVER["HTTP_X_REQUESTED_WITH"]) && strtolower($_SERVER["HTTP_X_REQUESTED_WITH"]) === "xmlhttprequest");
		if( 0
			&& !$isAjaxRequest
			&& !mt_rand(0, 1000)
			&& isArgeementCanBeShownForPage(Core::getRequest()->getGET("go"))
			&& getArgeementTime() > NS::getUser()->get('user_agreement_read') )
		{
			doHeaderRedirection("game.php/UserAgreement", false);
		}
		$this->checkForTW();

        // if(!isAdmin(null, true))
        {
            if( !$isAjaxRequest
                    && !preg_match("#^(Prefs)$#is", Core::getRequest()->getGET("go"))
                    && NS::getUser() && !isNameCharValid(NS::getUser()->get("username")) )
            {
                doHeaderRedirection("game.php/Prefs", false);
            }

            if( !$isAjaxRequest
                    && !preg_match("#^(PlanetOptions)$#is", Core::getRequest()->getGET("go"))
                    && NS::getPlanet() && NS::getPlanet()->getPlanetId()
                    && !isNameCharValid(NS::getPlanet()->getData('planetname')) )
            {
                doHeaderRedirection("game.php/PlanetOptions", false);
            }
        }

		if( ACHIEVEMENTS_ENABLED && !$isAjaxRequest
				// && !preg_match("#Info|Prefs|Achiev|Chat|MSG|Notepad|Alliance|Friends#is", Core::getRequest()->getGET("go"))
				&& preg_match("#Main|Constructions|Research|Shipyard|Defense|Mission|Artefacts#is", Core::getRequest()->getGET("go"))
				)
		{
			AchievementsService::loadAchievementsTemplateData();
		}

		$this->proceedPostActions()->proceedGetActions();
		if(!$this->actionCalled)
		{
			$this->callMethod("index", array());
		}
		return $this;
	}

	/**
	* Iterates through POST actions and call the methods.
	*
	* @return Page
	*/
	private function proceedPostActions()
	{
		if(is_null($this->postActions) || count(Core::getRequest()->getPOST()) <= 0)
		{
			return $this;
		}

		// План 37.7.2: CSRF защита (Origin/Referer check как defense-in-depth
		// поверх SameSite=Strict cookie). Любой POST должен исходить с того же
		// origin что наш сайт.
		if(!self::isPostOriginValid())
		{
			Logger::dieMessage('CSRF_INVALID_ORIGIN');
		}

		foreach(Core::getRequest()->getPOST() as $action => $value)
		{
			if($this->postActions->exists($action))
			{
				$method = $this->postActions->get($action);
				$args = $this->fetchArgs($method);
				$this->callMethod($method, $args);
			}
		}
		return $this;
	}

	/**
	 * План 37.7.2: проверка что POST пришёл с нашего origin.
	 *
	 * Логика: смотрим Origin (приоритет — современные браузеры всегда шлют
	 * для cross-origin POST), затем Referer (fallback для старых браузеров).
	 * Сравниваем с собственным host.
	 *
	 * Возвращает true если можно continue, false если CSRF подозрение.
	 */
	protected static function isPostOriginValid()
	{
		// Берём наш host из HTTP_HOST (что nginx прокинул).
		$ourHost = $_SERVER['HTTP_HOST'] ?? '';
		if($ourHost === '')
		{
			// Без HTTP_HOST не можем сравнить — допускаем (CLI тесты).
			return true;
		}

		// Origin: '<scheme>://<host>[:port]', Referer: полный URL.
		$origin = $_SERVER['HTTP_ORIGIN'] ?? '';
		if($origin !== '')
		{
			$parsed = parse_url($origin);
			$originHost = $parsed['host'] ?? '';
			if(isset($parsed['port']))
			{
				$originHost .= ':' . $parsed['port'];
			}
			return $originHost === $ourHost;
		}

		$referer = $_SERVER['HTTP_REFERER'] ?? '';
		if($referer !== '')
		{
			$parsed = parse_url($referer);
			$refererHost = $parsed['host'] ?? '';
			if(isset($parsed['port']))
			{
				$refererHost .= ':' . $parsed['port'];
			}
			return $refererHost === $ourHost;
		}

		// Нет ни Origin, ни Referer — для POST это подозрительно.
		// Современные браузеры всегда отправляют что-то одно.
		// Возможные false-positives: API-клиенты — но у нас их нет.
		return false;
	}

	/**
	* Iterates through GET actions and call the methods.
	*
	* @return Page
	*/
	private function proceedGetActions()
	{
		foreach($this->getActions as $getParam => $value)
		{
			$action = Core::getRequest()->getGET($getParam);
			if($action !== false && isset($value[$action]))
			{
				$method = $value[$action];
				$args = $this->fetchArgs($method);
				$this->callMethod($method, $args);
			}
		}
		return $this;
	}

	/**
	* Calls a methode of the child class and pass the given arguments.
	*
	* @param string	Method name
	* @param array		Arguments
	*
	* @return Page
	*/
	private function callMethod($method, array $args)
	{
		$this->actionCalled = true;
		call_user_func_array(array($this, $method), $args);
		return $this;
	}

	/**
	* Fetches the arguments from the given request method.
	*
	* @param string	The method name
	*
	* @return array	Method arguments
	*/
	private function fetchArgs($method)
	{
		$getArgStack = $this->getMethodArgsFromType("get", $method);
		$postArgStack = $this->getMethodArgsFromType("post", $method);
		$args = array();
		foreach($getArgStack as $arg)
		{
			$args[] = Core::getRequest()->getArgument("get", $arg);
		}
		foreach($postArgStack as $arg)
		{
			$args[] = Core::getRequest()->getArgument("post", $arg);
		}
		$coreArgs = $this->getMethodArgsFromType("inner", $method);
		if(count($coreArgs) > 0)
		{
			$args = array_merge($args, $coreArgs);
		}
		return $args;
	}

	/**
	* Returns the arguments according to the request type.
	*
	* @param string	Request type
	* @param string	Method name
	*
	* @return array	Arguments
	*/
	private function getMethodArgsFromType($type, $method)
	{
		$type = strtolower($type);
		switch($type)
		{
		case "inner":
			return (isset($this->args[$method])) ? $this->args[$method] : array();
			break;
		case "post":
			return (isset($this->postArgs[$method])) ? $this->postArgs[$method] : array();
			break;
		case "get":
			return (isset($this->getArgs[$method])) ? $this->getArgs[$method] : array();
			break;
		}
		return array();
	}

	/**
	* Adds a new get action.
	*
	* @param string	Action label
	* @param string	Action value
	* @param string	Method to call
	*
	* @return Page
	*/
	protected function setGetAction($action, $value, $method = null)
	{
		if(empty($method))
		{
			$method = $value;
		}
		$this->getActions[$action][$value] = $method;
		return $this;
	}

	/**
	* Adds a new post action.
	*
	* @param string	Action label
	* @param string	Method to call
	*
	* @return Page
	*/
	protected function setPostAction($action, $method = null)
	{
		if(is_null($this->postActions))
		{
			$this->postActions = new Map();
		}
		if(empty($method))
		{
			$method = $action;
		}
		$this->postActions->set($action, $method);
		return $this;
	}

	/**
	 * Adds post actions.
	 * @param array $actions Actions as array('action' => 'method')
	 * @return Page
	 */
	protected function setPostActions($actions)
	{
		if(is_null($this->postActions))
		{
			$this->postActions = new Map();
		}
		foreach( $actions as $action => $method )
		{
			if( is_numeric($action) )
			{
				$action = $method;
			}
			$this->postActions->set($action, $method);
		}
		return $this;
	}

	/**
	* Adds a new get argument onto stack.
	*
	* @param string	Method name
	* @param string	Argument name
	*
	* @return Page
	*/
	protected function addGetArg($method, $arg)
	{
		$this->getArgs[$method][] = $arg;
		return $this;
	}

	/**
	* Adds a new post argument onto stack.
	*
	* @param string	Method name
	* @param string	Argument name
	*
	* @return Page
	*/
	protected function addPostArg($method, $arg)
	{
		$this->postArgs[$method][] = $arg;
		return $this;
	}

	/**
	* Adds a new common argument onto the argument stack.
	*
	* @param string	Method name
	* @param string	Argument name
	*
	* @return Page
	*/
	protected function addArg($method, $arg)
	{
		$this->args[$method][] = $arg;
		return $this;
	}

	/**
	* Resets all actions.
	*
	* @return Page
	*/
	protected function resetActions()
	{
		$this->args = array();
		$this->getActions = array();
		$this->getArgs = array();
		$this->postActions = new Map();
		$this->postArgs = array();
		return $this;
	}

	/**
	* Quits the script.
	*
	* @return Page
	*/
	protected function quit()
	{
		exit();
		return $this;
	}

	protected function checkForTW()
	{
		$curr_time = time();
		$TW = []; // technical works отключены (legacy: Yii::app()->params['technicalWorks'])
		$start = 99999999999;
		$f_key = null;
		foreach( $TW as $key => $TWdata )
		{
			if( $TWdata['start'] < $start
				&& $curr_time <= strtotime($TWdata['end'])
				&& (
					(($curr_time + 60 * 60) >= strtotime($TWdata['start']) && $curr_time <= strtotime($TWdata['start']))
					|| ($curr_time >= strtotime($TWdata['start']))
				)
			)
			{
				$f_key = $key;
				$start = $TWdata['start'];
			}
		}
        if($f_key !== null){
            $TWdata = $TW[$f_key];
            $u_start= strtotime($TWdata['start']);
            $u_end 	= strtotime($TWdata['end']);
            if(
                ($curr_time + 60 * 60) >= $u_start
                    && $curr_time <= $u_end
                    && $curr_time <= $u_start
            )
            {
                Core::getTPL()->addHTMLHeaderFile("lib/jquery.countdown.js?".CLIENT_VERSION, "js");
                Core::getTPL()->addHTMLHeaderFile("lib/jquery.countdown-ru.js?".CLIENT_VERSION, "js");
                $timeleft = $u_start - $curr_time;
                $u_end = $u_end - $curr_time;
                Core::getTPL()->assign('tech_time', $timeleft);
                Core::getTPL()->assign('tech_works', true);
                Core::getLanguage()->assign('tw_end_time', $TWdata['end']);
                if( $TWdata['block'] )
                {
                    Core::getTPL()->assign('tech_works_name', Core::getLang()->getItem('TW_UPCOMING_BLOCK'));
                }
                else
                {
                    Core::getTPL()->assign('tech_works_name', Core::getLang()->getItem('TW_UPCOMING'));
                }
            }elseif($curr_time >= $u_start && $curr_time <= $u_end){
                Core::getTPL()->addHTMLHeaderFile("lib/jquery.countdown.js?".CLIENT_VERSION, "js");
                Core::getTPL()->addHTMLHeaderFile("lib/jquery.countdown-ru.js?".CLIENT_VERSION, "js");
                $timeleft = $u_end - $curr_time;
                Core::getTPL()->assign('tech_time', $timeleft);
                Core::getTPL()->assign('tech_works', true);
                Core::getTPL()->assign('tech_works_name', Core::getLang()->getItem('TW_IN_PROGRESS'));
                define('TW', true);
            }
        }elseif(DEATHMATCH){
            if(DEATHMATCH_END_TIME && $curr_time >= DEATHMATCH_END_TIME){
                Core::getTPL()->addHTMLHeaderFile("lib/jquery.countdown.js?".CLIENT_VERSION, "js");
                Core::getTPL()->addHTMLHeaderFile("lib/jquery.countdown-ru.js?".CLIENT_VERSION, "js");
                $timeleft = 0;
                Core::getTPL()->assign('tech_time', $timeleft);
                Core::getTPL()->assign('tech_works', true);
                Core::getTPL()->assign('tech_works_name', Core::getLang()->getItem('DEATHMATCH_END_MESSAGE'));
                define('TW', true);
            }elseif(DEATHMATCH_END_TIME && $curr_time >= DEATHMATCH_END_TIME-60*60*24*1){
                Core::getTPL()->addHTMLHeaderFile("lib/jquery.countdown.js?".CLIENT_VERSION, "js");
                Core::getTPL()->addHTMLHeaderFile("lib/jquery.countdown-ru.js?".CLIENT_VERSION, "js");
                $timeleft = DEATHMATCH_END_TIME - $curr_time;
                Core::getTPL()->assign('tech_time', $timeleft);
                Core::getTPL()->assign('tech_works', true);
                Core::getTPL()->assign('tech_works_name', Core::getLang()->getItem('DEATHMATCH_END_COUNTDOWN_MESSAGE'));
                define('TW', true);
            }elseif(DEATHMATCH_START_TIME && $curr_time < DEATHMATCH_START_TIME){
                Core::getTPL()->addHTMLHeaderFile("lib/jquery.countdown.js?".CLIENT_VERSION, "js");
                Core::getTPL()->addHTMLHeaderFile("lib/jquery.countdown-ru.js?".CLIENT_VERSION, "js");
                $timeleft = DEATHMATCH_START_TIME - $curr_time;
                Core::getTPL()->assign('tech_time', $timeleft);
                Core::getTPL()->assign('tech_works', true);
                Core::getTPL()->assign('tech_works_name', Core::getLang()->getItem('DEATHMATCH_START_COUNTDOWN_MESSAGE'));
                define('TW', true);
            }
        }
	}
}
?>