<?php
/**
* Cron class. Searches for expired cron tasks, execute them and
* calculate next execution time.
*
* @package Recipe 1.1
* @author Sebastian Noll
* @copyright Copyright (c) 2008, Sebastian Noll
* @license <http://www.gnu.org/licenses/gpl.txt> GNU/GPL
* @version $Id: Cron.class.php 28 2010-04-04 14:07:47Z craft $
*/

if(!defined("RECIPE_ROOT_DIR")) { die("Hacking attempt detected."); }

class Cron
{
  /**
  * Constructor.
  *
  * @return void
  */
  public function __construct()
  {
//    try {
    	$this->exeCron();
//    }
//    catch(Exception $e) { $e->printError(); }
    return;
  }

  /**
  * Search in database for expired cron tasks and execute them.
  *
  * @return void
  */
  protected function exeCron()
  {
    if(defined("EXEC_CRON") && EXEC_CRON)
    {
      $select = array("cronid", "script", "month", "day", "weekday", "hour", "minute", "xtime", "active");
      $result = Core::getQuery()->select("cronjob", $select, "", "xtime <= ".sqlVal(time())." AND active = '1'");
      while($row = Core::getDB()->fetch($result))
      {
        $script = CRONJOB_DIR.$row["script"];
        if(file_exists($script))
        {
          // Hook::event("EXECUTE_CRONJOB", array(&$this, &$row, $script));
          Core::getQuery()->update("cronjob", array("xtime", "last"), array($this->calcNextExeTime($row), $row["xtime"]), "cronid = ".sqlVal($row["cronid"]));
          require_once($script);
        }
        else
        {
          throw new GenericException("Cannot execute cron job \"$script\".", __FILE__, __LINE__);
        }
      }
      Core::getDatabase()->free_result($result);
    }
  }

  /**
  * Calculate next execution time.
  *
  * @param array		Database row of cron job
  *
  * @return integer	Next execution time as Unix timestamp
  */
  public function calcNextExeTime($row)
  {
//    try {
    	$row =  $this->validate($row);
//    }
//    catch(Exception $e) { $e->printError(); }

    $curTime = Date::getDateTime($row["xtime"]);
    $curTime["weekday"] = $this->getWeekdayAsInt($curTime["D"]);

    //### Next Minute ###//
    if(count($row["minute"]) > 1)
    {
      $minute = $this->nextMinute($row, $curTime);

      if($curTime["G"] > 23) { $curTime["G"] = 1; $curTime["j"]++; }
      $hour = $curTime["G"];
      if($curTime["j"] > $curTime["t"]) { $curTime["j"] = 1; $curTime["n"]++; }
      $d["day"] = $curTime["j"];
      if($curTime["n"] > 12) { $curTime["n"] = 1; $curTime["Y"]++; }
      $d["month"] = $curTime["n"];

    }
    else
    {
      $minute = $curTime["i"];

      //### Next Hour ###//
      if(count($row["hour"]) > 1)
      {
        $hour = $this->nextHour($row, $curTime);

        if($curTime["j"] > $curTime["t"]) { $curTime["j"] = 1; $curTime["n"]++; }
        $d["day"] = $curTime["j"];
        if($curTime["n"] > 12) { $curTime["n"] = 1; $curTime["Y"]++; }
        $d["month"] = $curTime["n"];
      }
      else
      {
        $hour = $curTime["G"];

        //### Next Day ###//
        $d = $this->nextDate($row, $curTime);
      }
    }
    // Hook::event("CALCULATE_NEXT_CRON_TIME", array(&$this, $row, &$hour, &$minute, &$d, &$curTime));
    return mktime($hour, $minute, 0/*second*/, $d["month"], $d["day"], $curTime["Y"]);
  }

  /**
  * Check incoming cron time data and put them into separate arrays.
  *
  * @param array		Database row to validate
  *
  * @return array	Validated row
  */
  protected function validate($row)
  {
    $row["minute"] = Arr::trimArray(explode(",", $row["minute"]));
    $row["hour"] = Arr::trimArray(explode(",", $row["hour"]));
    $row["day"] = Arr::trimArray(explode(",", $row["day"]));
    $row["weekday"] = Arr::trimArray(explode(",", $row["weekday"]));
    $row["month"] = Arr::trimArray(explode(",", $row["month"]));
    return $row;
  }

  /**
  * Returns the integer value depending on a weekday.
  * Week begins with monday (= 1) and ends with sunday (= 7).
  *
  * @param string	The weekday (three letters!)
  *
  * @return integer	Weekday value
  */
  protected function getWeekdayAsInt($weekday)
  {
    switch($weekday)
    {
    case "Mon":
      return 1;
      break;
    case "Tue":
      return 2;
      break;
    case "Wed":
      return 3;
      break;
    case "Thu":
      return 4;
      break;
    case "Fri":
      return 5;
      break;
    case "Sat":
      return 6;
      break;
    case "Sun":
      return 7;
      break;
    default:
      return 1;
      break;
    }
  }

  /**
  * Calculate next minute.
  *
  * @param array		Cronjob designations
  * @param array		All current time designations
  *
  * @return integer	Next minute
  */
  protected function nextMinute($row, &$curTime)
  {
    for($i = 0; $i < count($row["minute"]); $i++)
    {
      if($row["minute"][$i] > intval($curTime["i"]))
      {
        $minute = $row["minute"][$i];
        break;
      }
      if($i == count($row["minute"]) - 1)
      {
        $minute = $row["minute"][0];
        $curTime["G"]++;
        break;
      }
    }
    return $minute;
  }

  /**
  * Calculate next hour.
  *
  * @param array		Cronjob designations
  * @param array		All current time designations
  *
  * @return integer	Next hour
  */
  protected function nextHour($row, &$curTime)
  {
    if($curTime["G"] > 23) { $curTime["G"] = 0; $curTime["j"]++; }
    for($i = 0; $i < count($row["hour"]); $i++)
    {
      if($row["hour"][$i] > $curTime["G"])
      {
        $hour = $row["hour"][$i];
        break;
      }
      if($i == count($row["hour"]) - 1)
      {
        $hour = $row["hour"][0];
        $curTime["j"]++;
        $curTime["weekday"]++;
        break;
      }
    }
    return $hour;
  }

  /**
  * Calculate next date.
  *
  * @param array		Cronjob designations
  * @param array		All current time designations
  *
  * @return array	Next date
  */
  protected function nextDate($row, &$curTime)
  {
    if(count($row["weekday"]) < 7)
    {
      if($curTime["weekday"] > 7) { $curTime["weekday"] = 1; }
      for($i = 0; $i < count($row["weekday"]); $i++)
      {
        if($row["weekday"][$i] > $curTime["weekday"])
        {
          $weekday = $row["weekday"][$i];
          break;
        }
        if($i == count($row["weekday"]) - 1)
        {
          $weekday = 7 - $curTime["weekday"] + $row["weekday"][0];
          break;
        }
      }
      $d["day"] = $curTime["j"] + $weekday;
      if($d["day"] > $curTime["t"])
      {
        $d["day"] -= $curTime["t"];
        $curTime["n"]++;
      }
      $d["month"] = $curTime["n"];
    }
    else
    {
      if($curTime["j"] > $curTime["t"])
      {
        $curTime["j"] = 1;
        $curTime["n"]++;
      }
      for($i = 0; $i < count($row["day"]); $i++)
      {
        if($row["day"][$i] > $curTime["j"])
        {
          $d["day"] = $row["day"][$i];
          break;
        }
        if($i == count($row["day"]) - 1)
        {
          $d["day"] = $row["day"][0];
          $curTime["n"]++;
          break;
        }
      }

      //### Next Month ###//
      if($curTime["n"] > 12) { $curTime["n"] = 1; $curTime["Y"]++; }
      if(!in_array($curTime["n"], $row["month"]))
      {
        for($i = 0; $i < count($row["month"]); $i++)
        {
          if($row["month"][$i] > $curTime["n"])
          {
            $d["month"] = $row["month"][$i];
            break;
          }
          if($i == count($row["month"]) - 1)
          {
            $d["month"] = $row["month"][0];
            $curTime["Y"]++;
            break;
          }
        }
      }
      else
      {
        $d["month"] = $curTime["n"];
      }
    }
    return $d;
  }
}
?>