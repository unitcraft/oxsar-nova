<?php
/**
* Initilizing file. Launches program and define important constants.
*
* @package Recipe 1.1
* @author Sebastian Noll
* @copyright Copyright (c) 2008, Sebastian Noll
* @license <http://www.gnu.org/licenses/gpl.txt> GNU/GPL
* @version $Id: init.php 23 2010-04-03 19:08:34Z craft $
*/

// Set error reporting level.
error_reporting(ERROR_REPORTING_TYPE);

// If an error occured, throw an exception.
function errorHandler($errNo, $message, $file, $line)
{
  throw new GenericException($message, $file, $line, $errNo);
}
//
if($GLOBALS["RUN_YII"] != 1)
{
	set_error_handler("errorHandler", ERROR_REPORTING_TYPE);
}

// Handles uncaught exceptions
function exceptionHandler($exception)
{
//  $exception->printError();
}
//set_exception_handler("exceptionHandler");
if($GLOBALS["RUN_YII"] != 1)
{
	set_exception_handler("exceptionHandler");
}
// Init debuger
require_once(RECIPE_ROOT_DIR."Debuger.php");

// Load program.
require_once(RECIPE_ROOT_DIR."AutoLoader.php");
?>
