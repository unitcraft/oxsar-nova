<?php
/**
* Common functions for frontend.
*
* Oxsar http://oxsar.ru
*
*
*/

require_once("global.inc.php");
define("LOGIN_PAGE", true);

// Create core
new Core();

// Create CMS object
require_once(APP_ROOT_DIR."game/CMS.class.php");
$cms = new CMS();

Core::getLanguage()->load("Registration");

// Set menus
Core::getTPL()->addLoop("headerMenu", $cms->getMenu("h"));
Core::getTPL()->addLoop("footerMenu", $cms->getMenu("f"));

// Multiple universes
require_once(APP_ROOT_DIR."game/Uni.class.php");
$unis[] = new Uni("Uni 1", "subdomain", false);
// Add new Unis here:
// 1. Parameter: Universe name
// 2. Parameter: Domain or subdomain
// 3. Parameter: Internal or external server

/**
* Generates a select list with all universes.
*
* @param array		The unis
*
* @return string	Options
*/
function uniAsOptionList($unis)
{
  $options = "";
  foreach($unis as $uni)
  {
    if(Core::getRequest()->getCOOKIE("uni") == $uni->getName())
    {
      $selected = " selected=\"selected\"";
    }
    else
    {
      $selected = "";
    }
    $options .= "<option value=\"".$uni->getDomain()."\"".$selected.">".$uni->getName()."</option>";
  }
  return $options;
}

$showUniSelection = false;
if(count($unis) > 1)
{
  $showUniSelection = true;
  Core::getTPL()->assign("uniSelection", uniAsOptionList($unis));
}
Core::getTPL()->assign("showUniSelection", $showUniSelection);

// Assign char set
Core::getTPL()->assign("charset", Core::getLang()->getOpt("charset"));
?>