<?php
/**
* Support page.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined('APP_ROOT_DIR')) { die('Hacking attempt detected.'); }

class Support extends Page
{
	/**
	* Constructor. Handles requests for this page.
	*
	* @return void
	*/
	public function __construct()
	{
		parent::__construct();
		Core::getLanguage()->load('info');
		$this->proceedRequest();
		return;
	}
	
	protected function index()
	{
		Core::getTPL()->display('support');
		return $this;
	}
}