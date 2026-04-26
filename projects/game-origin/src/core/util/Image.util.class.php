<?php
/**
* Mainly usage: Generate image codes for HTML.
*
* @package Recipe 1.1
* @author Sebastian Noll
* @copyright Copyright (c) 2008, Sebastian Noll
* @license <http://www.gnu.org/licenses/gpl.txt> GNU/GPL
* @version $Id: Image.util.class.php 23 2010-04-03 19:08:34Z craft $
*/

if(!defined("RECIPE_ROOT_DIR")) { die("Hacking attempt detected."); }

class Image
{
  /**
  * Default CSS class for images.
  *
  * @var string
  */
  const IMAGE_CSS_CLASS = "image";

  /**
  * Generate image tag in HTML.
  *
  * @param string	Image URL
  * @param string	Additional title
  * @param integer	Image width
  * @param integer	Image height
  * @param string	Additional CSS class designation
  *
  * @return string	Image tag
  */
  public static function getImage($url, $title, $width = null, $height = null, $cssClass = "", $use_raw_url = false)
  {
    if(Core::getUser()->exists("theme") && Core::getUser()->get("theme") && !Link::isExternal($url))
    {
      $url = Core::getUser()->get("theme")."images/".$url;
    }
    if(!Link::isExternal($url) && !$use_raw_url)
    {
      $url = RELATIVE_URL."images/".$url;
    }
    if ( $use_raw_url )
    {
      $url = RELATIVE_URL.$url;
    }

    if(Str::length($cssClass) == 0) { $cssClass = self::IMAGE_CSS_CLASS; }
    $width = !is_null($width) ? " width=\"".$width."\"" : "";
    $height = !is_null($height) ? " height=\"".$height."\"" : "";
    $img = "<img src=\"".$url."\" title=\"".$title."\" "/*."alt=\"".$title."\""*/.$width.$height." class=\"".$cssClass."\" />";
    return $img;
  }
}
?>