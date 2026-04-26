<?php
/**
* Timer class. Stopwatch function.
*
* @package Recipe 1.1
* @author Sebastian Noll
* @copyright Copyright (c) 2008, Sebastian Noll
* @license <http://www.gnu.org/licenses/gpl.txt> GNU/GPL
* @version $Id: Timer.class.php 23 2010-04-03 19:08:34Z craft $
*/

if(!defined("RECIPE_ROOT_DIR")) { die("Hacking attempt detected."); }

class Timer
{
  /**
  * Time When Script starts.
  *
  * @var float
  */
  protected $startTime = 0;

  /**
  * Time When Script Ends.
  *
  * @var float
  */
  protected $endTime = 0;

  /**
  * Formatted time in seconds.
  *
  * @var string
  */
  protected $timeInSec = "";

  /**
  * Number of decimal places.
  *
  * @var integer
  */
  protected $decimalPlaces = 4;

  /**
  * Time.
  *
  * @var mixed
  */
  protected $time;

  /**
  * Constructor initialize Timer.
  *
  * @param integer	Start time
  *
  * @return void
  */
  public function __construct($time)
  {
    $this->time = $time;
    $this->startTimer();
    return;
  }

  /**
  * Record time, when script is launching.
  *
  * @return void
  */
  protected function startTimer()
  {
    list($usec, $sec) = explode(" ",$this->time);
    $this->startTime = ((float)$usec + (float)$sec);
    return;
  }

  /**
  * Stop recording time.
  *
  * @return float	The latest micro time
  */
  protected function stopTimer()
  {
    list($usec, $sec) = explode(" ",microtime(0));
    $this->endTimer = ((float)$usec + (float)$sec);
    return $this->endTimer;
  }

  /**
  * Returns time of stopwatch.
  *
  * @param boolean	Turns of the auto-format
  *
  * @return string	The total generating time in seconds
  */
  public function getTime($format = true)
  {
    $this->timeInSec = $this->stopTimer() - $this->startTime;
    if($format) { return $this->format($this->timeInSec); }
    return $this->timeInSec;
  }

  /**
  * Format the decimal places and add dimension.
  *
  * @param flaot		Unformatted time
  *
  * @return sting	Formatted time
  */
  protected function format($number)
  {
    return number_format($number, $this->decimalPlaces)." seconds";
  }
}
?>
