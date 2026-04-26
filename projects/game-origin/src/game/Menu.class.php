<?php
/**
* Menu class: Generates the menu by XML.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class Menu implements IteratorAggregate, Countable
{
  /**
  * The main XML menu element.
  *
  * @var SimpleXMLElement
  */
  protected $xml = "";

  /**
  * The menu config file.
  *
  * @var string
  */
  protected $menuFile = "";

  /**
  * Holds the menu items.
  *
  * @var array
  */
  protected $menu = null;

  /**
  * Constructor.
  *
  * @param string	Menu config file
  *
  * @return void
  */
  public function __construct($file)
  {
    $this->setMenuFile($file);
    return;
  }

  /**
  * Loads the XML file.
  *
  * @return Menu
  */
  protected function loadXML()
  {
    $this->xml = new XML($this->menuFile);
    $this->xml = $this->xml->get();
    return $this;
  }

  /**
  * Generates the menu items.
  *
  * @return Menu
  */
  protected function generateMenu()
  {
    $this->menu = array();
    $was_div = false;
    $i = 0;

    $msgs = sqlSelectField("message", "count(*)", "", "receiver = ".sqlUser()." AND readed = '0'");
    $friends = sqlSelectField("buddylist", "count(*)", "", "friend2 = ".sqlUser()." AND accepted = '0'");

	$allyapplications = 0;
	if(NS::getUser()->get("aid")) // && sqlSelectField("alliance", "founder", "", "aid"=.sqlVal(NS::getUser()->get("aid"))) == NS::getUser()->get("userid"))
	{
		$allyapplications = sqlSelectField("allyapplication", "count(*)", "", "aid = ".sqlVal(NS::getUser()->get("aid")));
	}

    foreach($this->xml as $first)
    {
    	if (
    		defined('SN')
    		&& ( trim($first->getAttribute("class-min")) == 'menu8-min' || trim($first->getAttribute("class-min")) == 'menu6-min')
    	)
    	{
    		continue;
    	}
      $i++;
      $inner = $this->getClass($first);

      $outer_bot = "";
      //$this->menu[$i]["outer_bot"] = "";
      $title_min = trim($first->getAttribute("title-min"));
      $class_max = trim($first->getAttribute("class-max"));
      $class_min = trim($first->getAttribute("class-min"));
      if ($class_max != "")
      {
        if ($inner != "")
        {
          $inner = preg_replace("#class=\"([^\"]+)\"#i", "class=\"$1 $class_max\"", $inner);
        }
        else
        {
          $inner = "class=\"$class_max\"";
        }
      }

      if (trim($this->getClass($first)) === 'class="menu-info"')
      {
        $inner .= " id=\"menuli$i\" onClick='menuSoH(\"$i\", \"$class_max\", \"$class_min\", \"$title_min\");'";
        $outer_bot = "<li id='menudiv$i'><ul>";
        if ($was_div)
          $this->menu[count($this->menu) -1]["outer_bot"] = "</ul></li>";
        $was_div = true;
      }
      array_push($this->menu, array(
        "outer_bot" => $outer_bot,
        "inner" => $inner,
        "content" => $this->getLabel($first),
        "direct" => ""
        ));
      foreach($first->getChildren() as $second)
      {
        /* $is_chat_test = $second->getAttribute("chat-test");
        if($is_chat_test)
        {
          // debug_var($second, "[generateMenu] group: $i, child");
          if(!(USE_CHAT_TEST && (!defined('IS_MOBILE_REQUEST') || !IS_MOBILE_REQUEST)))
          {
            continue;
          }
        } */
        if( $second->getAttribute("href") === "UserAgreement" && !SHOW_USER_AGREEMENT ){
            continue;
        }
        if( $second->getAttribute("href") === "Achievements" && !ACHIEVEMENTS_ENABLED ){
            continue;
        }
        if( $second->getAttribute("href") === "Referral" && !REFERRALS_ENABLED ){
            continue;
        }
    	if( $second->getAttribute("href") === "Widgets" && $_SESSION["username"] ?? "" != 'admin' )
    	{
    		continue;
    	}
        /* if( $second->getAttribute("href") === "Profession" && !isAdmin() ){
            continue;
        } */

        if( defined('SN') && $second->getAttribute("href") === "Logout" )
        {
            continue;
        }

//        if( defined('SN') && $second->getAttribute("href") === "Referral" && $_SERVER['REMOTE_ADDR'] != '95.171.1.55' && !defined('LOCAL') )
//        {
//            continue;
//        }

        // NS::getPlanetStack();
        if ($second->getAttribute("href") === "ExchangeOpts" && NS::getPlanet()->getBuilding(UNIT_EXCHANGE) < 1)
          continue;

        if ($second->getAttribute("href") === "MSG" && $msgs > 0)
          $content = $this->getLink($second, " ($msgs)");
        else if ($second->getAttribute("href") === "Friends" && $friends > 0)
          $content = $this->getLink($second, " ($friends)");
        else if ($second->getAttribute("href") === "Alliance" && $allyapplications > 0)
          $content = $this->getLink($second, " ($allyapplications)");
        else
          $content = $this->getLink($second);


        array_push($this->menu, array(
          "inner" => $this->getClass($second),
          "content" => $content,
          "direct" => $this->getDirectLink($second)
          ));
      }
    }

    if ($was_div)
      $this->menu[count($this->menu) -1]["outer_bot"] = "</ul></li>";

    return $this;
  }

  /**
  * Generates the label of a XML element.
  *
  * @param SimpleXMLElement
  *
  * @return string
  */
  protected function getLabel(XMLObj $xml)
  {
    $label = strval($xml->getAttribute("label"));
    $noLangVar = (bool) $xml->getAttribute("nolangvar");
    if($noLangVar)
    {
      return $label;
    }
    return ($label != "") ? Core::getLang()->get($label) : "";
  }

  /**
  * Returns the direct link of a XML element
  *
  * @param SimpleXMLElement
  *
  * @return string
  */
  protected function getDirectLink(XMLObj $xml)
  {
    $link = strval($xml->getAttribute("href"));
    $external = (bool) $xml->getAttribute("external");
    if(!$external)
    {
		$url = socialUrl(RELATIVE_URL."game.php/".$link);
		/*
	    if( defined('SN') )
	    {
		    if( strpos($url, '?') === false )
		    {
		    	$url .= '?' . '';
		    }
		    else
		    {
		    	$url .= '&' . '';
		    }
	    }
		*/
	    return $url;
    }
	return socialUrl($link);
	/*
    if( defined('SN') )
    {
	    if( strpos($link, '?') != false )
	    {
	    	$link .= '&' . '';
	    }
	    else
	    {
	    	$link .= '?' . '';
	    }
    }
    return $link;
	*/
  }

  /**
  * Returns the CSS class generated from XML element.
  *
  * @param SimpleXMLElement
  *
  * @return string
  */
  protected function getClass(XMLObj $xml)
  {
    $class = strval($xml->getAttribute("class"));
    return ($class != "") ? " class=\"$class\"" : "";
  }

  /**
  * Generates the from XML element.
  *
  * @param SimpleXMLElement
  *
  * @return string
  */
  protected function getName(XMLObj $xml)
  {
    $name = $xml->getString();
    $func = strval($xml->getAttribute("func"));
    if($func != "")
    {
      return $this->func($func, $name);
    }
    $noLangVar = (bool) $xml->getAttribute("nolangvar");
    if($noLangVar)
    {
      return $name;
    }
    return ($name != "") ? Core::getLang()->get($name) : "";
  }

  /**
  * Generates a link from XML element.
  *
  * @param SimpleXMLElement
  *
  * @return string
  */
  protected function getLink(XMLObj $xml, $str = '')
  {
    $name = $this->getName($xml) . $str;
    $title = $this->getLabel($xml);
    $attachment = strval($xml->getAttribute("target"));
    $attachment = ($attachment != "") ? "target='{$attachment}'" : "";
    $onclick = strval($xml->getAttribute("onclick"));
    if($onclick)
    {
      $attachment .= " onclick='{$onclick}'";
    }
    $id = strval($xml->getAttribute("id"));
    if($id)
    {
      $attachment .= " id='{$id}'";
    }
    $refdir = (bool) $xml->getAttribute("refdir");
    $external = (bool) $xml->getAttribute("external");
    $link = strval($xml->getAttribute("href"));
    if(!$external)
    {
      $link = "game.php/".$link;
    }
    return Link::get($link, $name, $title, "", $attachment, false, false, $refdir);
  }

  /**
  * Returns the menu items.
  *
  * @return array
  */
  public function getMenu()
  {
    if(is_null($this->menu))
    {
      $this->loadXML()->generateMenu();
    }
    return $this->menu;
  }

  /**
  * Sets the menu XML file.
  *
  * @param string	File name
  *
  * @return Menu
  */
  public function setMenuFile($file)
  {
    $this->menuFile = APP_ROOT_DIR."game/xml/".$file.".xml";
    return $this;
  }

  /**
  * Parses the function attrbiute.
  *
  * @param string	Function name
  * @param string	Parameter
  *
  * @return string
  */
  protected function func($function, $param = null)
  {
    $function = strtolower($function);
    switch($function)
    {
    case "version":
      return "v".NS_VERSION;
      break;
    case "const":
      return constant($param);
      break;
    case "config":
      return Core::getConfig()->get($param);
      break;
    case "user":
      return NS::getUser()->get($param);
      break;
    case "cookie":
      return Core::getRequest()->getCOOKIE($param);
      break;
    case "image":
      return Image::getImage($param, "");
      break;
    }
    return "";
  }

  /**
  * Retrieves an external iterator.
  *
  * @return ArrayIterator
  */
  public function getIterator(): Iterator
  {
    return new ArrayIterator($this->getMenu());
  }

  public function count(): int
  {
    $menu = $this->getMenu();
    return is_array($menu) ? count($menu) : 0;
  }

  /**
  * Destructor.
  *
  * @return void
  */
  public function kill()
  {
    // PHP 8: unset($this) запрещён
    return;
  }

	public function getMenuTitles($planets = array())
	{
		Core::getLanguage()->load('info,Menu');
		$result = array(
			array(
				'name'	=> 'TOP_MENU',
				'class'	=> 'navigation_header',
				'pselect' => false,
			),
			array(
				'name' => 'TOP_MENU_PLANETS',
				'class'	=> 'planets_header',
				'pselect' => true,
			),
			array(
				'name'	=> 'TOP_MENU_BUY_CREDIT',
				'url'	=> 'game.php/Payment',
				'class'	=> 'link',
				'pselect' => false,
			),
		);
		foreach ( $result as $key => $val )
		{
			if( $val['name'] == 'TOP_MENU' && isFacebookSkin() )
			{
				unset($result[$key]);
				continue;
			}
			if( $val['pselect'] == true )
			{
				$result[$key]['planets'] = $planets;
			}
			$result[$key]['name'] = $val['name'] = Core::getLanguage()->getItem($val['name']);
			if( isset($val['url']) )
			{
				$result[$key]['link'] = Link::get($val['url'], $val['name'], '', $val['class']);
			}
		}
		return $result;
	}
}
?>