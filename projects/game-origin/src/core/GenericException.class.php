<?php
/**
* Exception class. For error handling.
*
* @package Recipe 1.1
* @author Sebastian Noll
* @copyright Copyright (c) 2008, Sebastian Noll
* @license <http://www.gnu.org/licenses/gpl.txt> GNU/GPL
*/

if(!defined("RECIPE_ROOT_DIR")) { die("Hacking attempt detected."); }

class GenericException extends Exception implements GlobalException
{
    public $_file;
    public $_line;

  /**
  * Constructor.
  *
  * @param string	Message
  * @param string	File
  * @param integer	Line
  * @param integer	Additional code
  *
  * @return void
  */
  public function __construct($message, $file = "", $line = 0, $code = 0)
  {
    parent::__construct($message, $code);
    $this->_file = $file;
    $this->_line = $line;
  }

  /**
  * Sums the exception up.
  *
  * @return string	Formatted message
  */
  public function __toString()
  {
    if($this->_file != "" && $this->_line != 0)
    {
      return "Error in class ".$this->_file." (".$this->_line.") occurred: ".$this->message;
    }
    return "Error: $this->message";
  }

  /**
  * Prints this error.
  *
  * @return void
  */
  public function printError()
  {
    if(function_exists('debug_exception_handler'))
    {
      debug_exception_handler($this);
      return;
    }

    if(!file_exists(RECIPE_ROOT_DIR."error.tpl")) { die("Crash."); }
    $template = @file_get_contents(RECIPE_ROOT_DIR."error.tpl");
    $template = str_replace("ErrorMessage", $this->__toString(), $template);
    die($template);
  }

  /**
  * Returns the error message.
  *
  * @return string	The error description
  */
  public function getErrorMessage()
  {
    return $this->message;
  }
}
