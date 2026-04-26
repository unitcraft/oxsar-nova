<?php
/**
* Hook and event class.
*
* @package Recipe 1.1
* @author Sebastian Noll
* @copyright Copyright (c) 2008, Sebastian Noll
* @license <http://www.gnu.org/licenses/gpl.txt> GNU/GPL
* @version $Id: Hook.class.php 23 2010-04-03 19:08:34Z craft $
*/

if(!defined("RECIPE_ROOT_DIR")) { die("Hacking attempt detected."); }

class Hook
{
  /**
  * Global hook variable.
  *
  * @var array
  */
  protected static $hooks = array();

  /**
  * Arguments for event.
  *
  * @var mixed
  */
  protected static $args;

  /**
  * Last sent event.
  *
  * @var string
  */
  protected static $event;

  /**
  * Appends an hook.
  *
  * @param string	Event name
  * @param mixed		Hook data
  *
  * @return void
  */
  public static function addHook($event, $hookData)
  {
    self::$hooks[$event][] = $hookData;
    return;
  }

  /**
  * When an event is triggered.
  *
  * @param string	Event name
  *
  * @return void
  */
  public static function event($event, $args = null)
  {
    self::$event = $event;
//    try {
    	self::checkHooks();
//    }
//    catch(Exception $e) { $e->printError(); }

    if(count(self::$hooks) < 1 || !array_key_exists(self::$event, self::$hooks))
    {
      return false;
    }

    self::$args = $args;

//    try {
    	self::runHooks(self::$hooks[self::$event]);
//    }
//    catch(Exception $e) { $e->printError(); }
    return;
  }

  /**
  * Executes an hook.
  *
  * @param mixed		The hook to execute.
  *
  * @return void
  */
  protected static function runHooks($hook)
  {
    foreach($hook as $hookData)
    {
      if(is_array($hookData))
      {
        if(count($hookData) < 1)
        {
          throw new GenericException("Empty hook array received.", __CLASS__, __LINE__);
        }
        self::runHooks($hookData);
      }
      else if(is_string($hookData))
      {
//        try {
        	self::runFunction($hookData, self::$args);
//        }
//        catch(Exception $e) { $e->printError(); }
      }
      else if(is_object($hookData))
      {
        $method = "on".self::$event;
//        try {
        	self::runFunction(array($hookData, $method), self::$args);
//        }
//        catch(Exception $e) { $e->printError(); }
      }
      else
      {
        throw new GenericException("Unkown datatype for hook.", __CLASS__, __LINE__);
      }
    }
    return;
  }

  /**
  * Executes an function.
  * Note: If the called function returns false, there will be an error.
  *
  * @param string	Function to call
  * @param mixed		Arguments for function
  *
  * @return void
  */
  protected static function runFunction($callback, $args)
  {
    $return = call_user_func_array($callback, $args);
    if(is_string($return))
    {
      throw new GenericException("There is an error with plug in \"".$callback."\": ".$return);
    }
    if($return === false)
    {
      throw new GenericException("Plug in \"".$callback."\" returned an unknown error.");
    }
    return;
  }

  /**
  * Checks the hook variable for validation.
  *
  * @return void
  */
  protected static function checkHooks()
  {
    if(!is_array(self::$hooks))
    {
      throw new GenericException("Hooks has not been sent as an array.", __CLASS__, __LINE__);
    }
    return;
  }
}
?>