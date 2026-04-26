<?php
/**
* Object-oriented typing: Boolean
*
* @package Recipe 1.1
* @author Sebastian Noll
* @copyright Copyright (c) 2008, Sebastian Noll
* @license <http://www.gnu.org/licenses/gpl.txt> GNU/GPL
* @version $Id: Boolean.util.class.php 23 2010-04-03 19:08:34Z craft $
*/

if(!defined("RECIPE_ROOT_DIR")) { die("Hacking attempt detected."); }

require_once(RECIPE_ROOT_DIR."util/Type.abstract_class.php");

class Boolean extends Type
{
  /**
  * This is the actual boolean value.
  *
  * @var boolean
  */
  protected $bool = 0.0;

  /**
  * Initializes a newly created Bool object.
  *
  * @param mixed
  *
  * @return void
  */
  public function __construct($bool = null)
  {
    $this->set($bool);
    return;
  }

  /**
  * Sets a new boolean value.
  *
  * @return Boolean
  */
  public function set($bool)
  {
    if(is_null($bool))
    {
      $this->bool = false;
    }
    else if(is_bool($bool))
    {
      $this->bool = $bool;
    }
    else if(is_numeric($bool))
    {
      if($bool <= 0)
        $this->bool = false;
      else
        $this->bool = true;
    }
    else if(is_string($bool))
    {
      if($bool === "")
        $this->bool = false;
      else if($bool === "true")
        $this->bool = true;
      else if($bool === "false")
        $this->bool = false;
      else if($bool === "0")
        $this->bool = false;
      else
        $this->bool = true;
    }
    else if(is_array($bool))
    {
      if(count($bool) > 0)
        $this->bool = true;
      else
        $this->bool = false;
    }
    else if(is_object($bool))
    {
      $obj = $this->getFromArgument($bool);
      $this->set($obj->get());
    }
    else
    {
      $this->bool = false;
    }
    return $this;
  }

  /**
  * Returns the float value.
  *
  * @return boolean
  */
  public function get()
  {
    return $this->bool;
  }

  /**
  * Converts the boolean value to a string.
  *
  * @return string
  */
  public function toString()
  {
    if($this->bool)
    {
      return "true";
    }
    return "false";
  }

  /**
  * Returns the boolean as a String.
  *
  * @return String
  */
  public function getString()
  {
    return new OxsarString($this->toString());
  }

  /**
  * Converts the boolean value to an integer.
  *
  * @return integer
  */
  public function toInteger()
  {
    if($this->bool)
    {
      return 1;
    }
    return 0;
  }

  /**
  * Returns the boolean as an Integer.
  *
  * @return Integer
  */
  public function getInteger()
  {
    return new Integer($this->toInteger());
  }

  /**
  * Converts the boolean value to a float.
  *
  * @return float
  */
  public function toFloat()
  {
    if($this->bool)
    {
      return 1.0;
    }
    return 0.0;
  }

  /**
  * Returns the boolean as an Float.
  *
  * @return Float
  */
  public function getFloat()
  {
    return new Float($this->toFloat());
  }
}
?>