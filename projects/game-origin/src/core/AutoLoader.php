<?php
/**
* Include required files.
*
* @package Recipe 1.1
* @author Sebastian Noll
* @copyright Copyright (c) 2008, Sebastian Noll
* @license <http://www.gnu.org/licenses/gpl.txt> GNU/GPL
* @version $Id: AutoLoader.php 23 2010-04-03 19:08:34Z craft $
*/

if(!defined("RECIPE_ROOT_DIR")) { die("Hacking attempt detected."); }

//### All files to include ###//
$includingFiles = array(
                        "Functions.php",
                        "Exception.interface.php",
                        "plugins/Plugin.abstract_class.php"
                        );

//### Automatic plug in loader ###//
$pluginpath = RECIPE_ROOT_DIR."plugins/";
$handle = opendir($pluginpath);
$plugins = array();
while($file = readdir($handle))
{
  if($file != "." && $file != ".." && $file != ".htaccess" && $file != ".svn")
  {
    if(is_dir($pluginpath.$file))
    {
      // Open first directory layer and try to find another plug in
      $_handle = opendir($pluginpath.$file);
      while($_file = readdir($_handle))
      {
        if(!file_exists($pluginpath.$file."/".$_file.".dis") && preg_match("/^.+\.php$/i", $_file))
        {
          array_push($plugins, "plugins/".$file."/".$_file);
        }
      }
      unset($_file); unset($_handle);
    }
    // Create .dis-file to disable any plug in
    else if(!file_exists($pluginpath.$file.".dis") && preg_match("/^.+\.php$/i", $file) && !preg_match("/^.+\.abstract_class\.php$/i", $file))
    {
      array_push($plugins, "plugins/".$file);
    }
  }
}
closedir($handle);
if(count($plugins) > 0) { $includingFiles = array_merge($includingFiles, $plugins); }

//### Include files ###//
foreach($includingFiles as $inc)
{
  if(!file_exists(RECIPE_ROOT_DIR.$inc)) { die("Crash: ".$inc); }
  require_once(RECIPE_ROOT_DIR.$inc);
}

function __autoload($class)
{
  $include = false;
  $class = getClassPath($class);
  $coreDirs = explode(",", AUTOLOAD_PATH_CORE);
  foreach($coreDirs as $dir)
  {
    $classFile = RECIPE_ROOT_DIR.$dir.$class.".class.php";
    if(file_exists($classFile))
    {
      $include = $classFile;
      break;
    }
  }
  if(!$include)
  {
    $appDirs = explode(",", AUTOLOAD_PATH_APP);
    foreach($appDirs as $dir)
    {
      $classFile = APP_ROOT_DIR.$dir.$class.".class.php";
      if(file_exists($classFile))
      {
        $include = $classFile;
        break;
      }
    }
  }
  if(!$include)
  {
    $include = RECIPE_ROOT_DIR."util/".$class.".util.class.php";
  }
  if($include !== false && file_exists($include))
  {
    require_once($include);
  }
}
?>
