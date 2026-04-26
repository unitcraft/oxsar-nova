<?php
/**
* Sends HTTP requests via cURL extension.
*
* @package Recipe 1.1
* @author Sebastian Noll
* @copyright Copyright (c) 2008, Sebastian Noll
* @license <http://www.gnu.org/licenses/gpl.txt> GNU/GPL
* @version $Id: Curl.class.php 23 2010-04-03 19:08:34Z craft $
*/

if(!defined("RECIPE_ROOT_DIR")) { die("Hacking attempt detected."); }

class Curl extends Request_Adapter
{
  /**
  * The cURL handle.
  *
  * @var resource
  */
  protected $resource = null;

  /**
  * Handles the cURL session.
  *
  * @param string	URL to connect with
  *
  * @return void
  */
  public function __construct($webpage)
  {
    $this->webpage = $webpage;
    if(!function_exists("curl_init"))
    {
      throw new GenericException("The cURL library is not available on the server.", __FILE__, __LINE__);
    }
    $this->init();
    return;
  }

  /**
  * Initialize a cURL session.
  *
  * @return Curl
  */
  protected function init()
  {
    $this->resource = @curl_init();
    if(!$this->resource)
    {
      throw new GenericException("Connection faild via cURL request.", __FILE__, __LINE__);
    }
    // Hook::event("HTTP_REQUEST_INIT_FIRST", array(&$this));
    @curl_setopt($this->resource, CURLOPT_RETURNTRANSFER, true);
    @curl_setopt($this->resource, CURLOPT_URL, $this->webpage);
    $this->response = curl_exec($this->resource);
    $this->errorNo = curl_errno($this->resource);
    if($this->errorNo)
    {
      $this->error = curl_error($this->resource);
      throw new GenericException("There is an error occured in cURL session (".$this->errorNo."): ".$this->error, __FILE__, __LINE__);
    }
    // Hook::event("HTTP_REQUEST_INIT_LAST", array(&$this));
    return $this;
  }

  /**
  * Closes the current cURL session.
  *
  * @return void
  */
  public function close()
  {
    curl_close($this->resource);
    // PHP 8: unset($this) запрещён
    return;
  }
}
?>