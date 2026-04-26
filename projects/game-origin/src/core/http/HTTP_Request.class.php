<?php
/**
* Sends HTTP requests to webpages and returns their response.
*
* @package Recipe 1.1
* @author Sebastian Noll
* @copyright Copyright (c) 2008, Sebastian Noll
* @license <http://www.gnu.org/licenses/gpl.txt> GNU/GPL
* @version $Id: HTTP_Request.class.php 23 2010-04-03 19:08:34Z craft $
*/

if(!defined("RECIPE_ROOT_DIR")) { die("Hacking attempt detected."); }

class HTTP_Request
{
  /**
  * The request session.
  *
  * @var object
  */
  protected $session = null;

  /**
  * Function check for the allowed sessions.
  *
  * @var array
  */
  protected $functions = array(
    "Curl" => "curl_init",
    "Fsock" => "fsockopen"
    );

  /**
  * Initializes a new HTTP request with the given session.
  *
  * @param string	Webpage url
  * @param string	Session name
  *
  * @return void
  */
  public function __construct($webpage, $sessionName = null)
  {
    if(!is_null($sessionName) && array_key_exists($sessionName, $this->functions))
    {
      if(!function_exists($this->functions[$sessionName]))
      {
        throw new GenericException("Couldn't start HTTP request via ".$sessionName.". The function is not available on the server.", __FILE__, __LINE__);
      }
    }
    else
    {
      foreach($this->functions as $s => $f)
      {
        if(function_exists($f) && class_exists($s))
        {
          $sessionName = $s;
          break;
        }
      }
    }

    if($sessionName == "")
    {
      throw new GenericException("Unable to establish an HTTP request. Maybe the server does not support this feature.", __FILE__, __LINE__);
    }
    // Hook::event("START_HTTP_REQUEST", array(&$this, &$webpage, &$sessionName));

//    try {
    	$this->session = new $sessionName($webpage);
//    }
//    catch(Exception $e)
//    {
//      $e->printError();
//    }
    return;
  }

  /**
  * Returns the response.
  *
  * @return string
  */
  public function getResponse()
  {
    return $this->session->getResponse();
  }

  /**
  * Destructor for request.
  *
  * @return void
  */
  public function kill()
  {
    $this->session->close();
    // PHP 8: unset($this) запрещён
    return;
  }
}
?>