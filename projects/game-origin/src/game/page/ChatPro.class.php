<?php
/**
* It's not used yet.
*
* Oxsar http://oxsar.ru
*
* 
*/

class Chat extends Page
{
	public function __construct()
	{
		parent::__construct();
		$this->proceedRequest();
	}
	protected function index()
	{
		// Path to the chat directory:
		define('AJAX_CHAT_PATH', APP_ROOT_DIR.'../novax/chat/');
		define('AJAX_CHAT_URL', '/novax/chat/');

		// Include Class libraries:
		require(AJAX_CHAT_PATH.'lib/classes.php');

		// Initialize the chat:
		$ajaxChat = new OxsarAJAXChat();

		echo "TEST <br />";

		return $this->quit();
	}
}
?>