<?php
/**
* Automatic mailer.
*
* @package Recipe 1.1
* @author Sebastian Noll
* @copyright Copyright (c) 2008, Sebastian Noll
* @license <http://www.gnu.org/licenses/gpl.txt> GNU/GPL
* @version $Id: Email.util.class.php 23 2010-04-03 19:08:34Z craft $
*/

if(!defined("RECIPE_ROOT_DIR")) { die("Hacking attempt detected."); }

class Email
{
  /**
  * The receiver.
  *
  * @var string
  */
  protected $receiver;

  /**
  * The subject.
  *
  * @var string
  */
  protected $subject;

  /**
  * Mail body.
  *
  * @var String
  */
  protected $message;

  /**
  * Sender name.
  *
  * @var string
  */
  protected $sender;

  /**
  * Sender mail address.
  *
  * @var string
  */
  protected $senderMail;

  /**
  * Email header.
  *
  * @var string
  */
  protected $header = "";

  /**
  * Content type.
  *
  * @var string
  */
  protected $contentType = "text/html";

  /**
  * Separates email header fields.
  *
  * @var string
  */
  protected $headerSeparator = "\n";

  /**
  * Holds the mail transmission status.
  *
  * @var boolean
  */
  protected $success = null;

  /**
  * Mail charset.
  *
  * @var string
  */
  protected $charset = null;

  /**
  * Constructor: Builds the header and starts mailer.
  *
  * @param string	The receiver's mail address [optional]
  * @param string	Mail subject [optional]
  * @param string	The message [optional]
  *
  * @return void
  */
  public function __construct($receiver = null, $subject = null, $message = null)
  {
    if(!is_null($receiver))
    {
      $this->setReceiver($receiver);
    }
    if(!is_null($subject))
    {
      $this->setSubject($subject);
    }
    if(!is_null($message))
    {
      $this->setMessage($message);
    }
    $this->setSender(Core::getOptions()->pagetitle);
    $this->setSenderMail(Core::getOptions()->mailaddress);
    return;
  }

  /**
  * Checks if mail address is valid.
  *
  * @param string	EMail address
  * @param boolean	Disable exceptions (default: true) [optional]
  *
  * @return boolean
  */
  protected function isMail($mail, $throwException = true)
  {
    if(!preg_match("#^[a-zA-Z0-9-]+([._a-zA-Z0-9.-]+)*@[a-zA-Z0-9.-]+\.([a-zA-Z]{2,4})$#is", $mail))
    {
      $this->success = false;
      if($throwException)
      {
        throw new IssueException("NO_VALID_EMAIL_ADDRESS");
      }
      return false;
    }
    return true;
  }

  /**
  * Sends mail.
  *
  * @param boolean	Disable exceptions (default: true) [optional]
  *
  * @return Email
  */
  public function sendMail($throwException = true)
  {
    // Hook::event("SEND_MAIL", array(&$this));
    if(!@mail($this->receiver, $this->subject, $this->message->get(), $this->header))
    {
      $this->success = false;
      if($throwException)
      {
        throw new GenericException("There is an error with sending mail.<br />Receiver: ".$this->receiver.", Subject: ".$this->subject.", Header: ".$this->header."<br /><br />".$this->message);
      }
      return $this;
    }
    $this->success = true;
    return $this;
  }

  /**
  * Builds the mail header.
  *
  * @return Email
  */
  protected function buildHeader()
  {
    $this->header  = "From: ".$this->sender.(($this->senderMail != "") ? " <".$this->senderMail.">" : "").$this->headerSeparator;
    $this->header .= "Reply-To: ".$this->sender.$this->headerSeparator;
    $this->header .= "MIME-Version: 1.0".$this->headerSeparator;
    $this->header .= "Content-Transfer-Encoding: 8bit".$this->headerSeparator;
    $this->header .= "Content-Type: ".$this->contentType."; charset=".$this->getCharset().$this->headerSeparator;
    return $this;
  }

  public function getSuccess()
  {
    return (bool) $this->success;
  }

  /**
  * Setter-method for mail header saparator.
  *
  * @param string	\n\r, \n
  *
  * @return Email
  */
  public function setHeaderSeprarator($headerSaparator)
  {
    $this->headerSeparator = $headerSaparator;
    return $this->buildHeader();
  }

  /**
  * Setter-method for subject.
  *
  * @param string
  *
  * @return Email
  */
  public function setSubject($subject)
  {
    $this->subject = $subject;
    return $this;
  }

  /**
  * Setter-method for mail message.
  *
  * @param string
  *
  * @return Email
  */
  public function setMessage($message)
  {
    if(!($message instanceof OxsarString))
    {
      $message = new OxsarString($message);
    }
    $this->message = $message->trim()->regEx("#(\r\n|\r|\n)#", $this->headerSeparator);
    return $this;
  }

  /**
  * Setter-method for sender name.
  *
  * @param string
  *
  * @return Email
  */
  public function setSender($sender)
  {
    $this->sender = $sender;
    return $this->buildHeader();
  }

  /**
  * Setter-method for sender mail address
  *
  * @param string
  *
  * @return Email
  */
  public function setSenderMail($senderMail)
  {
    $this->senderMail = $senderMail;
    return $this->buildHeader();
  }

  /**
  * Setter-method for receiver mail address.
  *
  * @param string
  *
  * @return Email
  */
  public function setReceiver($receiver)
  {
    $this->isMail($receiver);
    $this->receiver = $receiver;
    return $this;
  }

  /**
  * Setter-method for the content type.
  *
  * @param string	text/plain or text/html
  *
  * @return Email
  */
  public function setContentType($contentType)
  {
    $this->contentType = $contentType;
    return $this->buildHeader();
  }

  /**
  * Sets the email charset.
  *
  * @param string
  *
  * @return Email
  */
  public function setCharset($charset)
  {
    $this->charset = $charset;
    return $this->buildHeader();
  }

  /**
  * Returns the sender name.
  *
  * @return string
  */
  public function getSender()
  {
    return $this->sender;
  }

  /**
  * Returns sender mail address.
  *
  * @return string
  */
  public function getSenderMail()
  {
    return $this->senderMail;
  }

  /**
  * Returns email charset.
  *
  * @return string
  */
  public function getCharset()
  {
    if(is_null($this->charset))
    {
      $this->charset = Core::getLang()->getOpt("charset");
    }
    return $this->charset;
  }
}
?>