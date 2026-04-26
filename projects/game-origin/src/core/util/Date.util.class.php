<?php
/**
* Date related functions.
*
* @package Recipe 1.1
* @author Sebastian Noll
* @copyright Copyright (c) 2008, Sebastian Noll
* @license <http://www.gnu.org/licenses/gpl.txt> GNU/GPL
* @version $Id: Date.util.class.php 23 2010-04-03 19:08:34Z craft $
*/

if(!defined("RECIPE_ROOT_DIR")) { die("Hacking attempt detected."); }

class Date
{
	/**
	* Date and time formats.
	*
	* @var string
	*/
	protected static $dateFormat = "", $timeFormat = "";

	/**
	* Default date and time formats.
	*
	* @var string
	*/
	protected static $defaultDateFormat = "Y-m-d", $defaultTimeFormat = "H:i:s";

	/**
	* Returns current timestamp.
	*
	* @return integer	Unix timestamp
	*/
	public static function getTimestamp()
	{
		return time();
	}

	/**
	* Generates the timestamp of a date.
	*
	* @return integer	The converted timestamp or now.
	*/
	public static function strToTime($date)
	{
		if(($timestamp = strtotime($date)) !== false)
		{
			return $timestamp;
		}

		if(list($day, $month, $year, $time) = split('[/. ]', $date))
		{
			return strtotime($day."-".$month."-".$year." ".$time);
		}
		return self::getTimestamp();
	}

	public static function timeToString($type = 1, $timestamp = -1, $format = "", $replaceShortDate = true)
	{
		if($timestamp < 0)
		{
			$timestamp = self::getTimestamp();
		}
		if($format != "")
		{
			$type = 3;
		}
		$date = "";
		$isShortDate = false;
		// Replace date designation, if timestamp is today or yesterday.
		// Note: Does not work with specific date format.
		if($replaceShortDate && $format == "")
		{
			$currentDay = intval(date("j", time()));
			$statedDay = intval(date("j", $timestamp));
			if(self::getTimestamp() - $timestamp < 172800 && self::getTimestamp() > $timestamp && $currentDay - 1 == $statedDay)
			{
				$date = Core::getLanguage()->getItem("YESTERDAY");
				$isShortDate = true;
			}
			if(self::getTimestamp() - $timestamp < 86400 && self::getTimestamp() > $timestamp && $currentDay == $statedDay)
			{
				$date = Core::getLanguage()->getItem("TODAY");
				$isShortDate = true;
			}

			if($type == 1 && $isShortDate)
			{
				return "<span class=\"cur-day\">".$date."</span> ".date(self::getTimeFormat(), $timestamp);
			}
			else if($type == 2 && $isShortDate)
			{
				return "<span class=\"cur-day\">".$date."</span>";
			}
		}

		switch($type)
		{
				// Date + Time Format
			default:
			case 1:
				$date = date(self::getDateFormat()." ".self::getTimeFormat(), $timestamp);
				break;
	
				// Only Date Format
			case 2:
				$date = date(self::getDateFormat(), $timestamp);
				break;
	
				// Specific Format
			case 3:
				$date = date($format, $timestamp);
				break;
		}
		// Hook::event("CONVERT_TIMESTAMP_CLOSE", array($type, $timestamp, &$format, &$date));
		return $date;
	}

	public static function getDateTime($timestamp = -1)
	{
		if($timestamp < 0)
		{
			$timestamp = self::getTimestamp();
		}
		return array(
			'a' => date('a', $timestamp), // am/pm
			'A' => date('A', $timestamp), // AM/PM
			'd' => date('d', $timestamp), // Day of month, 2 digits
			'D' => date('D', $timestamp), // Day of week, 3 letters
			'F' => date('F', $timestamp), // Full name of month
			'g' => intval(date('g', $timestamp)), // 12-hour format without leading zeros
			'G' => intval(date('G', $timestamp)), // 24-hour format without leading zeros
			'h' => date('h', $timestamp), // 12-hour format with leading zeros
			'H' => date('H', $timestamp), // 24-hour format with leading zeros
			'i' => date('i', $timestamp), // Minutes with leading zeros
			'I' => date('I', $timestamp), // Daylight Saving Time
			'j' => intval(date('j', $timestamp)), // Day of month without leading zeros
			'l' => date('l', $timestamp), // Full name of week
			'L' => date('L', $timestamp), // Leap year
			'm' => date('m', $timestamp), // Month as digit with leading zeros
			'M' => date('M', $timestamp), // Day of month, 3 letters
			'n' => intval(date('n', $timestamp)), // Month as digit without leading zeros
			's' => date('s', $timestamp), // Seconds with leading zeros
			't' => intval(date('t', $timestamp)), // Day number of month
			'T' => TIMEZONE,//date('T', $timestamp), // Timezone abbreviation
			'w' => intval(date('w', $timestamp)), // Numeric day of week (0 = Sunday, 6 = Saturday)
			'Y' => intval(date('Y', $timestamp)), // Full number of year, 4 digits
			'y' => date('y', $timestamp), // Year, 2 digits
			'z' => intval(date('z', $timestamp)), // Day of given year
			);
	}

	/**
	* Returns the used date format.
	*
	* @return string
	*/
	public static function getDateFormat()
	{
		if(!empty(self::$dateFormat))
		{
			return self::$dateFormat;
		}
		if(Core::getLang()->exists("DATE_FORMAT"))
		{
			self::$dateFormat = Core::getLang()->get("DATE_FORMAT");
		}
		else
		{
			self::$dateFormat = self::$defaultDateFormat;
		}
		return self::$dateFormat;
	}

	/**
	* Returns the used time format.
	*
	* @return string
	*/
	public static function getTimeFormat()
	{
		if(!empty(self::$timeFormat))
		{
			return self::$timeFormat;
		}
		if(Core::getLang()->exists("TIME_FORMAT"))
		{
			self::$timeFormat = Core::getLang()->get("TIME_FORMAT");
		}
		else
		{
			self::$timeFormat = self::$defaultTimeFormat;
		}
		return self::$timeFormat;
	}


	/**
	* Проверяет дату(дд.мм.гггг) на соответствие ограничению, в случае необходимости корректирует.
	*
	* @param string $usr_date	Строковое значение формата dd.mm.yyyy
	* @param int $constraint	Ограничение в секундах, макс\мин разница между текущей и проверяемой датой.
	* @param string $ctype		Функция сравнения -- max. или min.
	*
	* @return array		('stamp' => значение в Unix формате, 'string' => строковое значение)
	*/
	public static function validateDate($usr_date, $constraint = 0, $ctype='max')
	{
		$cur_date = mktime(23,59,59, date('m'), date('d'), date('Y'));
		$const_date = $cur_date - $constraint;

		if(preg_match("#^(\d\d)[/\.\-](\d\d)[/\.\-](\d{4})$#", trim($usr_date), $regs))
		{
			$usr_date = mktime(0, 0, 0, $regs[2], $regs[1], $regs[3]);
		}
		else if(!is_numeric($usr_date))
		{
			$usr_date = $const_date;
		}

		if ($constraint != 0)
			$usr_date = $ctype === 'max' ? max($usr_date, $const_date) : min($usr_date, $const_date);

		$res['stamp'] = $usr_date;
		$res['string'] = date("d.m.Y", $usr_date);

		return $res;
	}
}
?>