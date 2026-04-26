<?php
/**
* Displays Widgets
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined('APP_ROOT_DIR')) { die('Hacking attempt detected.'); }

class ExtWidgets extends Page
{
	/**
	* Constructor. Handles requests for this page.
	*
	* @return void
	*/
	public function __construct()
	{
		parent::__construct();
		$this->proceedRequest();
		return;
	}
	
	protected function index()
	{
		Core::getTPL()->display('widgets');
		return $this;
	}
}
