<?php
/**
* Advanced array functions.
*
* @package Recipe 1.1
* @author Sebastian Noll
* @copyright Copyright (c) 2008, Sebastian Noll
* @license <http://www.gnu.org/licenses/gpl.txt> GNU/GPL
* @version $Id: Arr.util.class.php 23 2010-04-03 19:08:34Z craft $
*/

if(!defined("RECIPE_ROOT_DIR")) { die("Hacking attempt detected."); }

class Arr
{
  /**
  * Trims all elements of an array.
  *
  * @param array		Array to be trimed
  *
  * @return array	Trimed array
  */
  public static function trimArray($array)
  {
    for($i = 0; $i < count($array); $i++)
    {
      $array[$i] = trim($array[$i]);
    }
    return $array;
  }

  /**
  * Check two arrays for equal size.
  *
  * @param array		First
  * @param array		Second
  *
  * @return void
  */
  public static function checkArraySize($array1, $array2)
  {
    if(count($array1) != count($array2))
    {
      throw new GenericException("Different number of attributes and values.");
    }
    return;
  }

  /**
  * Checks if both parameters are arrays.
  *
  * @param mixed		First array
  * @param mixed		Second array
  *
  * @return void
  */
  public static function checkArrays($array1, $array2)
  {
    if(!is_array($array1) && is_array($array2) || !is_array($array2) && is_array($array1))
    {
      throw new GenericException("You must send two arrays!");
    }
    return;
  }

  /**
  * Remove all elements with no content.
  *
  * @param array		Array to be cleaned
  *
  * @return array	Cleaned array
  */
  public static function clean($array)
  {
    for($i = 0; $i < count($array); $i++)
    {
      if(Str::length($array[$i]) > 0) { $rArray[$i] = $array[$i]; }
    }
    return $rArray;
  }

  /**
  * Alias to trimArray()-method.
  *
  * @param array		Array to be trimed
  *
  * @return array	Trimed array
  */
  public static function trim($array)
  {
    return self::trimArray($array);
  }
}
?>