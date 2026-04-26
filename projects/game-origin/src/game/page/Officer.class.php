<?php
/**
* Officer module.
*
* Oxsar http://oxsar.ru
*
* 
*/

class Officer extends Page
{
	var $off_cost = array();

	public function __construct()
	{
		parent::__construct();

		$this->off_cost = array("", 100, 200, 200, 200);

		$this->setPostAction("hire", "hireOfficer");
		$this->addPostArg("hireOfficer", "off_id");
		$this->proceedRequest();
	}
	protected function index()
	{
		$off_result = sqlSelect("officer", array("credit", "of_1", "of_2", "of_3", "of_4"), "", "userid = ".sqlUser());
		$off_res = sqlFetch($off_result);

		for ($i = 1; $i <= 4; $i++)
		{
			$get{$i} = ($off_res["credit"] >= $this->off_cost[$i]) ? "<input type=\"submit\" class=\"button\" name=\"hire\" value=\"Нанять\" onclick=\"document.getElementById('off_id').value='{$i}';\" />" : "не хватает кредитов";
			Core::getTPL()->assign("get{$i}", $get{$i});

			$off_res["of_{$i}"] = ($off_res["of_{$i}"] == 0) ? "<span class=\"false2\">- не нанят -</span>" : date("d.m.Y H:i:s", $off_res["of_{$i}"]);
			Core::getTPL()->assign("off_{$i}", $off_res["of_{$i}"]);
			Core::getTPL()->assign("off_cost_{$i}", $this->off_cost[$i]);
		}

		Core::getTPL()->display("officer");
		return $this->quit();
	}

	protected function hireOfficer($off_id)
	{
		$off_result = sqlSelect("officer", array("credit", "of_{$off_id}"), "", "userid = ".sqlUser());
		$off_res = sqlFetch($off_result);

		$credit = $off_res["credit"] - $this->off_cost[$off_id];
		if($credit >= 0)
		{
			$of_time = ($off_res["of_{$off_id}"] == 0) ? time() + (7*60*60*24) : $off_res["of_{$off_id}"] + (7*60*60*24);
			Core::getQuery()->update("officer", array("credit", "of_{$off_id}"), array($credit, $of_time), "userid = ".sqlUser());
		}
		doHeaderRedirection("game.php/Officer", false);
		return $this->index();
	}
}
?>