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
		// План 37.5d.5#11: stub TestAlienAI debug-страницы.
		// User_YII::showChatOnline() удалён вместе с Yii (cnt online users
		// без статичного метода). CVarDumper тоже Yii-only.
		// Если понадобится — раскомментировать и заменить на наш аналог.
		echo "chat online: N/A (stubbed in plan 37.5d.5)<br />";
		echo '<pre>';
		var_export(AlienAI::checkAlientNeeds());
		echo '</pre>';
		echo '<br /> TestAlientAI::index: done';
	}
}
?>