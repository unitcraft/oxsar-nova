<?php
/**
* This class imports XML language files into the database.
*
* @package Recipe 1.1
* @author Sebastian Noll
* @copyright Copyright (c) 2008, Sebastian Noll
* @license <http://www.gnu.org/licenses/gpl.txt> GNU/GPL
* @version $Id: XML_LanguageImporter.class.php 23 2010-04-03 19:08:34Z craft $
*/

if(!defined("RECIPE_ROOT_DIR")) { die("Hacking attempt detected."); }

class XML_LanguageImporter extends LanguageImporter
{
  /**
  * XML data.
  *
  * @var SimpleXMLElement
  */
  protected $xml = null;

  /**
  * Constructor
  *
  * @param string	XML-file or XML-data
  *
  * @return void
  */
  public function __construct($data)
  {
    $this->setData($data);
    return;
  }

  /**
  * Sets the XML data
  *
  * @param string	XML-file or XML-data
  *
  * @return XML_LanguageImporter
  */
  public function setData($data)
  {
    $this->xml = new XML($data);
    $this->xml = $this->xml->get();
    return $this;
  }

  /**
  * Proceeds the import.
  *
  * @return XML_LanguageImporter
  */
  public function proceed()
  {
    if(is_null($this->xml))
    {
      throw new GenericException("No XML data set.");
    }
    $defaultLang = false;
    foreach($this->xml->children() as $lang)
    {
      if($lang->getAttribute("action") == "setDefault")
      {
        $defaultLang = $lang->getName();
      }
      $createData = array(
        "langcode"	=> $lang->getName(),
        "title"		=> $lang->getAttribute("title"),
        "charset"	=> $lang->getAttribute("charset")
        );
      if(!$lang->getAttribute("create"))
      {
        $createData = null;
      }
      $this->getFromLangCode($lang->getName(), $createData);
      foreach($lang->getChildren() as $group)
      {
        $this->getFromGroupName($group->getName());
        $this->importData = array();
        foreach($group->getChildren() as $phrase)
        {
          $this->importData[$phrase->getName()] = $phrase->getString();
        }
        $this->import();
      }
    }
    if($defaultLang)
    {
      $this->setDefaultLanguage($defaultLang);
    }
    return $this;
  }
}
?>