<?php
/**
* Advanced string functions.
*
* @package Recipe 1.1
* @author Sebastian Noll
* @copyright Copyright (c) 2008, Sebastian Noll
* @license <http://www.gnu.org/licenses/gpl.txt> GNU/GPL
* @version $Id: Str.util.class.php 23 2010-04-03 19:08:34Z craft $
*/

if(!defined("RECIPE_ROOT_DIR")) { die("Hacking attempt detected."); }

class Str
{
  /**
  * Holds all valid encryptions used by encode().
  *
  * @var array
  */
  protected static $validEncryptions = array("md5", "sha1", "crypt", "crc32", "base64_encode");

  protected static function _resolveCharset($charset)
  {
    if($charset === null && Core::getLanguage())
      $charset = Core::getLanguage()->getOpt("charset");
    // getOpt returns the key name when not found — treat that as no charset
    if(empty($charset) || $charset === 'charset')
      return null;
    return $charset;
  }

  /**
  * Generate XHTML valid code.
  *
  * @param string	Incoming text
  *
  * @return string	Validated text
  */
  public static function validateXHTML($text)
  {
    if(strlen($text) > 0)
    {
      if(self::compare(Core::getLanguage()->getOpt("charset"), "utf-8")) { $chars = HTML_SPECIALCHARS; }
      else { $chars = HTML_ENTITIES; }
      $XHTMLConvertEntities = get_html_translation_table($chars);
      $text = trim($text);
      $text = strtr($text, $XHTMLConvertEntities);
    }
    return $text;
  }

  /**
  * Encode string to specified encryption method.
  *
  * @param string	String to encode
  * @param string	Encryption method
  *
  * @return string	Encoded string
  */
  public static function encode($text, $encryption = "md5")
  {
    if(in_array($encryption, self::$validEncryptions))
    {
      return $encryption($text);
    }
    return $text;
  }

  /**
  * Identical to substr().
  *
  * @param string	String
  * @param integer	Start position
  * @param integer	Substring length
  *
  * @return string	The extracted part of string
  */
  public static function substring($string, $start, $length = null, $charset = null)
  {
    if($length === null)
    {
      $length = self::length($string, $charset);
    }
    $charset = self::_resolveCharset($charset);
    if($charset)
      return mb_substr($string, $start, $length, $charset);
    return substr($string, $start, $length);
  }

  /**
  * Identical to strlen().
  *
  * @param string
  *
  * @return integer	Length of string
  */
  public static function length($string, $charset = null)
  {
    $charset = self::_resolveCharset($charset);
    if($charset)
      return mb_strlen($string, $charset);
    return strlen($string);
  }

  /**
  * Compare two strings for equality.
  *
  * @param string
  * @param string
  * @param boolean	Enable or disable case sensitive for comparision
  *
  * @return boolean True, if strings are equal, false, if not.
  */
  public static function compare($string1, $string2, $caseSensitive = false)
  {
    if($caseSensitive)
    {
      if(strcmp($string1, $string2) == 0) { return true; } else { return false; }
    }
    else
    {
      if(strcasecmp($string1, $string2) == 0) { return true; } else { return false; }
    }
  }

  /**
  * Identical to str_replace().
  *
  * @param mixed		Search
  * @param mixed		Replace
  * @param mixed		Subject
  *
  * @return mixed	Returns a string or an array with the replaced values
  */
  public static function replace($search, $replace, $subject)
  {
    return str_replace($search, $replace, $subject);
  }

  public static function tr($subject, $search)
  {
    foreach($search as $key => $value)
    {
      $subject = str_replace($key, $value, $subject);
    }
    return $subject;
  }

  /**
  * Returns the portion of haystack which
  * ends at the last occurrence of needle
  * and starts with the begin of haystack.
  *
  * @param string 	The string to search in
  * @param string	Needle
  *
  * @return string	New String
  */
  public static function reverse_strrchr($haystack, $needle)
  {
    $pos = strrpos($haystack, $needle);
    if($pos == false) { return $haystack; }
    return substr($haystack, 0, $pos + 1);
  }

  /**
  * Checks if needle can be found in haystack.
  *
  * @param string	Needle
  * @param string	Haystack
  *
  * @return boolean
  */
  public static function inString($needle, $haystack)
  {
    $match = strpos($haystack, $needle);
    if($match === false)
    {
      return false;
    }
    return true;
  }
}
?>