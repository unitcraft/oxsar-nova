<?php
/**
* Function to clear old sessions.
* 
* @package Recipe 1.1
* @author Sebastian Noll
* @copyright Copyright (c) 2008, Sebastian Noll
* @license <http://www.gnu.org/licenses/gpl.txt> GNU/GPL
* 
* @param integer Delete sessions older than "days".
*/

function clearSessions($days = 0)
{
  if($days > 0)
  {
    $deldate = $days * 86400;
    $and = " AND time < ".sqlVal($deldate);
  }
  else { $and = ""; }
  Core::getQuery()->delete("sessions", "logged = 0 $and");
}
?>
