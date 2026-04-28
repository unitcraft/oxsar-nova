<?php
/**
* Represents an universe.
*
* Oxsar http://oxsar.ru
*
* 
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class Uni
{
  /**
  * Name of the universe.
  *
  * @var string
  */
  protected $name = "";

  /**
  * Subdomain of the universe.
  *
  * @var string
  */
  protected $domain = "";

  /**
  * The universe lays on another server.
  *
  * @var boolean
  */
  protected $external = false;

  /**
  * Creates a new Universe object.
  *
  * @param string	Name
  * @param string	Subdomain
  *
  * @return void
  */
  public function __construct($name, $domain, $external = false)
  {
    $this->name = $name;
    $this->domain = $domain;
    $this->external = $external;
    return;
  }

  /**
  * Returns the name.
  *
  * @return string
  */
  public function getName()
  {
    return $this->name;
  }

  /**
  * Returns the full domain with HTTP host.
  *
  * @return string
  */
  public function getDomain()
  {
    if($this->isExternal())
    {
      return "http://".$this->domain."/";
    }
    $parsedUrl = parseUrl($_SERVER['HTTP_HOST']);
    return "http://".$this->domain.$parsedUrl["domain"].$parsedUrl["extension"]."/";
  }

  /**
  * Returns the external flag.
  *
  * @return boolean
  */
  public function isExternal()
  {
    return $this->external;
  }

  /**
  * Sets the name.
  *
  * @param string
  *
  * @return void
  */
  public function setName($name)
  {
    $this->name = $name;
    return;
  }

  /**
  * Sets the domain.
  *
  * @param string
  *
  * @return void
  */
  public function setDomain($domain)
  {
    $this->domain = $domain;
    return;
  }

  /**
  * Sets the external flag.
  *
  * @param boolean
  *
  * @return void
  */
  public function setExternal($external)
  {
    $this->external = $external;
  }
}
?>