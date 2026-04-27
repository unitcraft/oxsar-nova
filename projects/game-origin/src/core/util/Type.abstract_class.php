<?php
/**
* Object-oriented typing: Abstract typing
*
* @package Recipe 1.1
* @author Sebastian Noll
* @copyright Copyright (c) 2008, Sebastian Noll
* @license <http://www.gnu.org/licenses/gpl.txt> GNU/GPL
* @version $Id: Type.abstract_class.php 23 2010-04-03 19:08:34Z craft $
*/

abstract class Type
{
  protected $validObjects = array(
    "OxsarString",
    "Integer",
    "Float",
    "Map",
    "Boolean"
    );

  /**
  * Returns the raw data of an argument.
  *
  * @param mixed Argument
  *
  * @return Object The converted data
  */
  protected function getFromArgument($arg, $type = null)
  {
    if(is_object($arg))
    {
//      try {
        $this->checkObject($arg);
//      }
//      catch(Exception $e)
//      {
//        $e->printError();
//      }
      return $arg;
    }
    if(!is_null($type))
    {
      $type = new OxsarString($type);
      $type = $type->upperWords()->get();
      if(in_array($type, $this->validObjects))
      {
        return new $type($arg);
      }
    }
//    try {
      return $this->convertToObject($arg);
//    }
//    catch(Exception $e)
//    {
//      $e->printError();
//    }
  }

  /**
  * Checks an object for validation.
  *
  * @param Object
  *
  * @return void
  */
  protected function checkObject($object)
  {
    $className = get_class($object);
    if(!in_array($className, $this->validObjects))
    {
      throw new IssueException("Supplied argument (".$className.") is no valid object.");
    }
    return;
  }

  /**
  * Converts an argument to an object.
  *
  * @param mixed
  *
  * @return Object
  */
  protected function convertToObject($arg)
  {
    if(is_string($arg))
    {
      return new OxsarString($arg);
    }
    if(is_array($arg))
    {
      return new Map($arg);
    }
    if(is_float($arg))
    {
      return new OxsarFloat($arg);
    }
    if(is_int($arg) || is_numeric($arg))
    {
      return new Integer($arg);
    }
    if(is_bool($arg))
    {
      return new Boolean($arg);
    }
    throw new IssueException("Supplied argument has an unkown type.");
  }

  /**
  * Calls a function
  *
  * @param string	Method name
  * @param array		Method arguments
  *
  * @return mixed	The callback as an object or false
  */
  protected function call($method, array $args = null)
  {
    $method = $this->getFromArgument($method, "String");
    if(function_exists($method->get()))
    {
      $args = $this->getFromArgument($args, "Map");
      $type = $this->getType()->get();
      $callback = call_user_func_array($method->get(), $args->add($this->get())->getArray());
      return new $type($callback);
    }
    return false;
  }

  /**
  * Returns the variable type of the current instance.
  *
  * @return String
  */
  private function getType()
  {
    return new OxsarString(get_class($this));
  }
}
?>
