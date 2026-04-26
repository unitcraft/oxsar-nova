<?php
/**
* Sends HTTP requests via domain socket connection.
*
* @package Recipe 1.1
* @author Sebastian Noll
* @copyright Copyright (c) 2008, Sebastian Noll
* @license <http://www.gnu.org/licenses/gpl.txt> GNU/GPL
* @version $Id: Fsock.class.php 23 2010-04-03 19:08:34Z craft $
*/

if(!defined("RECIPE_ROOT_DIR")) { die("Hacking attempt detected."); }

class Fsock extends Request_Adapter
{
  /**
  * Parsed url.
  *
  * @var array
  */
  protected $parsedURL = array();

  /**
  * The fsock handle.
  *
  * @var resource
  */
  protected $resource = null;

  /**
  * The connection timeout, in seconds.
  *
  * @var integer
  */
  protected $timeout = 10;

  /**
  * Handles the fsock session.
  *
  * @param string	URL to connect with
  *
  * @return void
  */
  public function __construct($webpage)
  {
    parent::__construct($webpage);
    if(!function_exists("fsockopen") && ini_get("allow_url_fopen"))
    {
      throw new GenericException("The server does not allow to fetch contents from a different webpage.", __FILE__, __LINE__);
    }
    $this->init();
    return;
  }

  /**
  * Initialize a fsock session.
  *
  * @return Fsock
  */
  protected function init()
  {
    // Split the url
    $this->parsedURL = @parse_url($this->webpage);
    if(!$this->parsedURL["host"])
    {
      throw new GenericException("The supplied url is not valid.", __FILE__, __LINE__);
    }
    if(!$this->parsedURL["port"])
    {
      $this->parsedURL["port"] = 80;
    }
    if(!$this->parsedURL["path"])
    {
      $this->parsedURL["path"] = "/";
    }
    if($this->parsedURL["query"])
    {
      $this->parsedURL["path"] .= "?".$this->parsedURL["query"];
    }

    $this->resource = fsockopen($this->parsedURL["host"], $this->parsedURL["port"], $this->errorNo, $this->error, $this->timeout);
    if(!$this->resource)
    {
      throw new GenericException("Connection faild via fsock request: (".$this->errorNo.") ".$this->error, __FILE__, __LINE__);
    }
    @stream_set_timeout($this->resource, $this->timeout);
    // Hook::event("HTTP_REQUEST_INIT_FIRST", array(&$this));

    // Set headers
    $headers[] = "GET ".$this->parsedURL["path"]." HTTP/1.0";
    $headers[] = "Host: ".$this->parsedURL["host"];
    $headers[] = "Connection: Close";
    $headers[] = "\r\n";
    $headers = implode("\r\n", $headers);

    // Send request
    if(!@fwrite($this->resource, $headers))
    {
      throw new GenericException("Couldn't send request.", __FILE__, __LINE__);
    }

    // Read response
    while(!feof($this->resource))
    {
      $this->response .= fgets($this->resource, 12800);
    }
    $this->response = explode("\r\n\r\n", $this->response, 2);
    $this->response = $this->response[1];
    // Hook::event("HTTP_REQUEST_INIT_LAST", array(&$this));
    return $this;
  }

  /**
  * Closes the current fsock session.
  *
  * @return void
  */
  public function close()
  {
    fclose($this->resource);
    // PHP 8: unset($this) запрещён
    return;
  }
}
?>