<?php
/**
* TestAlientAI module.
*
* Oxsar http://oxsar.ru
*
* 
*/

class TestAlienAI extends Page
{
	public function __construct()
	{
		parent::__construct();

		// $this->setPostAction("hire", "hireOfficer");
		// $this->addPostArg("hireOfficer", "off_id");
		$this->proceedRequest();
	}
	protected function index()
	{
		echo "chat online: ".User_YII::showChatOnline()."<br />";
		CVarDumper::dump(AlienAI::checkAlientNeeds(), 10, 1);
		echo '<br /> TestAlientAI::index: done';
	}
}
?>