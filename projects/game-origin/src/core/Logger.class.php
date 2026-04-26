<?php
/**
* Displays message in error box.
*
* @package Recipe 1.1
* @author Sebastian Noll
* @copyright Copyright (c) 2008, Sebastian Noll
* @license <http://www.gnu.org/licenses/gpl.txt> GNU/GPL
* @version $Id: Logger.class.php 23 2010-04-03 19:08:34Z craft $
*/

if(!defined("RECIPE_ROOT_DIR")) { die("Hacking attempt detected."); }

class Logger
{
	/**
	* Add a message to log.
	*
	* @param string	Message to log
	* @param string	Log mode
	*
	* @return void
	*/
	public static function addMessage($message, $mode = "error")
	{
		Core::getLanguage()->load("error");
		$message = Core::getLanguage()->getItem($message);
		$message = "<div class=\"".$mode."\">".$message."</div>";
		Core::getTPL()->addLogMessage($message);
		return;
	}
	
	public static function addFlashMessage($message, $mode = "ui-state-error ui-corner-all")
	{
		Core::getLanguage()->load("error");
		$message = Core::getLanguage()->getItem($message);
		if( $mode == 'ui-state-error ui-corner-all' )
		{
			$message = '<td><span class="ui-icon ui-icon-alert" style="float: left; margin-right: .3em;"></span></td><td>'
				. $message
				. '</td>';
		}
		if( $mode == 'success' )
		{
			$mode	.= ' ui-state-highlight ui-corner-all';
			$message = '<td><span class="ui-icon ui-icon-info" style="float: left; margin-right: .3em;"></span></td><td>'
				. $message
				. '</td>';
		}
		if( $mode == 'error' )
		{
			$mode	.= ' ui-state-error ui-corner-all';
			$message = '<td><span class="ui-icon ui-icon-alert" style="float: left; margin-right: .3em;"></span></td><td>'
				. $message
				. '</td>';
		}
		$i = 0;
		while ( $_SESSION["flash_" . $mode . " logger" . $i] ?? false != false )
		{
			++$i;
		}
		$_SESSION["flash_" . $mode . " logger" . $i] = Core::getLanguage()->getItem($message);
		return;
	}

	/**
	* Displays a message and shut program down.
	*
	* @param string	Message to log
	*
	* @return void
	*/
	public static function dieMessage($message, $mode = "error")
	{
		error_log($message, 'warning');
		Core::getLanguage()->load("error");
		$message = Core::getLanguage()->getItem($message);
		Core::getTPL()->addLogMessage("<div class=\"".$mode."\">".$message."</div>");
		Core::getTemplate()->display("error");
		exit;
	}

	/**
	* Formats a message.
	*
	* @param string	Raw log message
	* @param string	Log mode
	*
	* @return string	Formatted message
	*/
	public static function getMessageField($message, $mode = "error")
	{
		Core::getLanguage()->load("error");
		$message = Core::getLanguage()->getItem($message);
		$message = "<span class=\"field_".$mode."\">".$message."</span>";
		// Hook::event("MESSAGE_FIELD", array(&$message));
		return $message;
	}
}
?>