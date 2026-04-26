<?php
/**
* This class handles Ajax requests.
*
* @package Recipe 1.1
* @author Sebastian Noll
* @copyright Copyright (c) 2008, Sebastian Noll
* @license <http://www.gnu.org/licenses/gpl.txt> GNU/GPL
* @version $Id: AjaxRequestHelper.abstract_class.php 23 2010-04-03 19:08:34Z craft $
*/

if(!defined("RECIPE_ROOT_DIR")) { die("Hacking attempt detected."); }

abstract class AjaxRequestHelper extends Page
{
  /**
  * If HTTP header has been sent.
  *
  * @var boolean
  */
  protected $headersSent = false;

  /**
  * If out stream is compressed.
  *
  * @var boolean
  */
  protected $compressed = false;

  /**
  * Displays the text for clear Ajax output.
  *
  * @param string	The text to output
  *
  * @return AjaxRequestHelper
  */
  protected function display($outstream)
  {
    $this->sendHeader();
    terminate($outstream);
    return $this;
  }

  /**
  * Set response header.
  *
  * @return AjaxRequestHelper
  */
  protected function sendHeader()
  {
    if(!headers_sent() || !$this->headersSent)
    {
      if(@extension_loaded('zlib') && !$this->compressed && GZIP_ACITVATED)
      {
        ob_start("ob_gzhandler");
        $this->compressed = true;
      }
      @header("Content-Type: text/html; charset=".Core::getLanguage()->getOpt("charset"));
      $this->headersSent = true;
    }
    return $this;
  }
}
?>