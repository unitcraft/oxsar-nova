<?php
/**
* HTTP request adapter.
*
* @package Recipe 1.1
* @author Sebastian Noll
* @copyright Copyright (c) 2008, Sebastian Noll
* @license <http://www.gnu.org/licenses/gpl.txt> GNU/GPL
* @version $Id: Request_Adapter.class.php 23 2010-04-03 19:08:34Z craft $
*/

if(!defined("RECIPE_ROOT_DIR")) { die("Hacking attempt detected."); }

abstract class Request_Adapter
{
  /**
  * The requesting url.
  *
  * @var string
  */
  protected $webpage = "";

  /**
  * Response body.
  *
  * @var string
  */
  protected $response = "";

  /**
  * The last error number.
  *
  * @var integer
  */
  protected $errorNo = 0;

  /**
  * The last error message.
  *
  * @var string
  */
  protected $error = "";

  /**
  * Constructor.
  *
  * @param string	URL to connect with
  *
  * @return void
  */
  public function __construct($webpage)
  {
    $this->webpage = $webpage;
    return;
  }

  /**
  * Returns the response.
  *
  * @return string
  */
  public function getResponse()
  {
    return $this->response;
  }

  /**
  * Returns the latest error message.
  *
  * @return string
  */
  public function getError()
  {
    return $this->error;
  }

  /**
  * Returns the latest error number.
  *
  * @return integer
  */
  public function getErrorNo()
  {
    return $this->errorNo;
  }

  /**
  * Sets the destination URL.
  *
  * @param string
  *
  * @return Request_Adapter
  */
  public function setWebpage($webpage)
  {
    $this->webpage = $webpage;
    return $this;
  }

  /**
  * Force adapting classes to implement init method.
  *
  * @return Request_Adapter
  */
  abstract protected function init();

  /**
  * Force adapting classes to implement the destructor.
  *
  * @return Request_Adapter
  */
  abstract public function close();
}
?>