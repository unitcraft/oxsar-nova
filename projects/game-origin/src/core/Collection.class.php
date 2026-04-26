<?php
/**
* Base collection class.
*
* @package Recipe 1.1
* @author Sebastian Noll
* @copyright Copyright (c) 2008, Sebastian Noll
* @license <http://www.gnu.org/licenses/gpl.txt> GNU/GPL
* @version $Id: Collection.class.php 23 2010-04-03 19:08:34Z craft $
*/

if(!defined("RECIPE_ROOT_DIR")) { die("Hacking attempt detected."); }

abstract class Collection implements Countable, IteratorAggregate
{
  // PHP 8: count() требует Countable; делегируем в size()
  public function count(): int
  {
    return $this->size();
  }

  // PHP 8: foreach по объекту требует Iterator; даём ArrayIterator над $item
  public function getIterator(): Iterator
  {
    return new ArrayIterator(is_array($this->item) ? $this->item : [$this->item]);
  }
  // public abstract function __construct();
  public abstract function get($var);
  public abstract function set($var, $value);

  /**
  * Holds collection data.
  *
  * @var unknown_type
  */
  protected $item = array();

  /**
  * Supports access to collection variable as class property.
  * ($collection->variable)
  *
  * @param string	Variable name
  *
  * @return string	Value
  */
  public function __get($var)
  {
    return $this->get($var);
  }

  /**
  * Supports access to collection variable as class property.
  * ($collection->variable = "Value")
  *
  * @param string	Variable name
  * @param mixed		Value
  *
  * @return Collection
  */
  public function __set($var, $value)
  {
    $this->set($var, $value);
    return $this;
  }

  /**
  * Checks if the given key or index exists in the collection.
  *
  * @param string	Variable to check
  *
  * @return boolean	True if the given key is set in the session
  */
  public function exists($var)
  {
    if(array_key_exists($var, $this->item)) { return true; } else { return false; }
  }

  /**
  * Returns the size of all collection items.
  *
  * @return integer
  */
  public function size()
  {
    return count($this->item);
  }
}
?>