<?php
/**
* Compiles language phrases.
*
* @package Recipe 1.1
* @author Sebastian Noll
* @copyright Copyright (c) 2008, Sebastian Noll
* @license <http://www.gnu.org/licenses/gpl.txt> GNU/GPL
* @version $Id: LanguageCompiler.class.php 23 2010-04-03 19:08:34Z craft $
*/

if(!defined("RECIPE_ROOT_DIR")) { die("Hacking attempt detected."); }

class LanguageCompiler
{
  /**
  * Phrase content.
  *
  * @var String
  */
  protected $phrase = null;

  /**
  * Regular expression patterns.
  *
  * @var array
  */
  protected $patterns = array();

  /**
  * Replacement for the patterns.
  *
  * @var array
  */
  protected $replacement = array();

  /**
  * Modifier for regular expression.
  *
  * @var string
  */
  protected $modifier = "";

  /**
  * Constructor.
  *
  * @param string	The phrase to be compiled
  * @param boolean	Replaces wildcards dynamically or hardly
  *
  * @return void
  */
  public function __construct($phrase, $hardReplace = false)
  {
    $this->modifier = ($hardReplace) ? "siUe" : "siU";
    $this->buildPatterns($hardReplace)->setPhrase($phrase);
    return;
  }

  /**
  * Builds the search and replace pattern.
  *
  * @param boolean	Replaces the search pattern with hard code or dynamically
  *
  * @return LanguageCompiler
  */
  protected function buildPatterns($hardReplace)
  {
    $this->patterns["link"] = "/\{link\[(.*)]}(.*)\{\/link}/".$this->modifier;
    $this->patterns["config"][] = "/\{config}([^\"]+)\{\/config}/".$this->modifier;
    $this->patterns["config"][] = "/\{config=([^\"]+)\}/".$this->modifier;
    $this->patterns["user"][] = "/\{user}([^\"]+)\{\/user}/".$this->modifier;
    $this->patterns["user"][] = "/\{user=([^\"]+)\}/".$this->modifier;
    $this->patterns["request"] = "/\{request\[([^\"]+)\]\}([^\"]+)\{\/request\}/".$this->modifier;
    $this->patterns["const"][] = "/\{const}([^\"]+)\{\/const}/".$this->modifier;
    $this->patterns["const"][] = "/\{const=([^\"]+)\}/".$this->modifier;
    $this->patterns["image"] = "/\{image\[([^\"]+)]}([^\"]+)\{\/image}/".$this->modifier;
    $this->patterns["time"][] = "/\{time}(.*)\{\/time}/".$this->modifier;
    $this->patterns["time"][] = "/\{time=(.*)\}/".$this->modifier;

    if($hardReplace)
    {
      $this->replacement["link"] = 'Link::get("\\2", "\\1")';
      $this->replacement["config"] = 'Core::getOptions()->get("\\1")';
      $this->replacement["user"] = 'Core::getUser()->get("\\1")';
      $this->replacement["request"] = 'Core::getRequest()->\\1["\\2"]';
      $this->replacement["const"] = 'constant("\\1")';
      $this->replacement["image"] = 'Image::getImage("\\2", "\\1")';
      $this->replacement["time"] = 'Date::timeToString(3, -1, "\\1", false)';
    }
    else
    {
      $this->replacement["link"] = "\".Link::get(\"\\2\", \"\\1\").\"";
      $this->replacement["config"] = "\".Core::getOptions()->get(\"\\1\").\"";
      $this->replacement["user"] = "\".Core::getUser()->get(\"\\1\").\"";
      $this->replacement["request"] = "\".Core::getRequest()->\\1[\"\\2\"].\"";
      $this->replacement["const"] = "\".\\1.\"";
      $this->replacement["image"] = "\".Image::getImage(\"\\2\", \"\\1\").\"";
      $this->replacement["time"] = "\".Date::timeToString(3, -1, \"\\1\", false).\"";
    }
    return $this;
  }

  /**
  * Compiles the phrase content.
  *
  * @return LanguageCompiler
  */
  protected function compile()
  {
    $this	->phrase
      ->replace("\"", "\\\"")
      ->regEx($this->patterns["link"], $this->replacement["link"])
      ->regEx($this->patterns["config"], $this->replacement["config"])
      ->regEx($this->patterns["user"], $this->replacement["user"])
      ->regEx($this->patterns["request"], $this->replacement["request"])
      ->regEx($this->patterns["const"], $this->replacement["const"])
      ->regEx($this->patterns["image"], $this->replacement["image"])
      ->regEx($this->patterns["time"], $this->replacement["time"]);
    // Hook::event("COMPILE_PHRASE", array(&$this));
    return $this;
  }

  /**
  * Returns compiled phrase.
  *
  * @return string
  */
  public function getPhrase()
  {
    return $this->phrase->get();
  }

  /**
  * Unset this object.
  *
  * @return void
  */
  public function shutdown()
  {
    unset($this);
    return;
  }

  /**
  * Sets the phrase.
  *
  * @param String
  *
  * @return LanguageCompiler
  */
  public function setPhrase($phrase)
  {
    if($phrase instanceof String)
    {
      $this->phrase = $phrase;
    }
    else
    {
      $this->phrase = new String($phrase);
    }
    $this->compile();
    return $this;
  }
}
?>