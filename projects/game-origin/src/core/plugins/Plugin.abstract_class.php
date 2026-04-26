<?php
/**
* Abstract class for Plugins. Should be implemented by every plug in.
*
* @package Recipe 1.1
* @author Sebastian Noll
* @copyright Copyright (c) 2008, Sebastian Noll
* @license <http://www.gnu.org/licenses/gpl.txt> GNU/GPL
* @version $Id: Plugin.abstract_class.php 23 2010-04-03 19:08:34Z craft $
*/

if(!defined("RECIPE_ROOT_DIR")) { die("Hacking attempt detected."); }

abstract class Plugin
{
  /**
  * Plug in name.
  *
  * @var string
  */
  protected $pluginName;

  /**
  * Plugin version.
  *
  * @var string
  */
  protected $pluginVersion;

  /**
  * Return plug in name.
  *
  * @return string
  */
  public function getPluginName()
  {
    return $this->pluginName;
  }

  /**
  * Return plug in version.
  *
  * @return string
  */
  public function getPluginVersion()
  {
    return $this->pluginVersion;
  }

  public abstract function admin();
  public abstract function install();
}
?>