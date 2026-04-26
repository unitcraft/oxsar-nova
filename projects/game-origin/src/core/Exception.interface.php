<?php
/**
* Exception interface. Define expection classes.
* 
* @package Recipe 1.1
* @author Sebastian Noll
* @copyright Copyright (c) 2008, Sebastian Noll
* @license <http://www.gnu.org/licenses/gpl.txt> GNU/GPL
* @version $Id: Exception.interface.php 23 2010-04-03 19:08:34Z craft $ 
*/

if(!defined("RECIPE_ROOT_DIR")) { die("Hacking attempt detected."); }

interface GlobalException
{
  public function printError();
}
?>
