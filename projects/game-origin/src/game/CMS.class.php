<?php
/**
* A really light CMS method.
*
* Oxsar http://oxsar.ru
*
* 
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class CMS
{
  /**
  * Language id.
  *
  * @var integer
  */
  protected $langid = 0;

  /**
  * Holds all menu items.
  *
  * @var unknown_type
  */
  protected $menuItems = array();

  /**
  * Creates a new CMS object.
  *
  * @return void
  */
  public function __construct()
  {
    $this->langid = Core::getLang()->getOpt("languageid");
    $this->loadMenuItems();
    return;
  }

  /**
  * Displays a page.
  *
  * @param string	Page label
  *
  * @return CMS
  */
  public function showPage($page)
  {
    if(empty($page)) { return; }
    $result = sqlSelect("page", array("title", "content"), "", "label = ".sqlVal($page)." AND languageid = ".sqlVal($this->langid)." AND label != ''");
    if($row = sqlFetch($result))
    {
      Core::getTPL()->assign("page", $row["title"]);
      Core::getTPL()->assign("content", $row["content"]);
      Core::getTPL()->display("cms_page", false, "front");
      exit;
    }
    return $this;
  }

  /**
  * Loads all menu items.
  *
  * @return CMS
  */
  protected function loadMenuItems()
  {
    $result = sqlSelect("page", array("position", "title", "label", "link"), "", "languageid = ".sqlVal($this->langid), "displayorder ASC");
    while($row = sqlFetch($result))
    {
      $position = $row["position"];
      if(!empty($row["link"]))
      {
        $this->menuItems[$position][]["link"] = Link::get($row["link"], $row["title"], $row["title"]);
      }
      else
      {
        $this->menuItems[$position][]["link"] = Link::get("login.php/go:".$row["label"], $row["title"], $row["title"]);
      }
    }
    return $this;
  }

  /**
  * Returns an array with all menu items of the given location.
  * (e.g. f = footer, h = header)
  *
  * @param char		Menu location
  *
  * @return array	Menu items
  */
  public function getMenu($position)
  {
    if(isset($this->menuItems[$position]) && count($this->menuItems[$position]) > 0)
    {
      return $this->menuItems[$position];
    }
    return array();
  }
}
?>