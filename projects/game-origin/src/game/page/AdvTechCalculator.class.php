<?php
/**
* Advanced tech calculator module.
*
* Oxsar http://oxsar.ru
*
* 
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class AdvTechCalculator extends Page
{
	public function __construct()
	{
		// Core::getLanguage()->load("info,buildings");
		$this->proceedRequest();
	}

	protected function index()
	{
		$tech_table = array(
			array(0.75, 2.00, 0.00),
			array(0.00, 1.00, 2.00),
			array(2.00, 0.00, 1.00)
		);
	
	Core::getTPL()->assign("cfg_tech_scale_1", 1);
	Core::getTPL()->assign("cfg_tech_scale_2", 1.1);
	Core::getTPL()->assign("cfg_tech_scale_3", 1.2);

	Core::getTPL()->assign("cfg_tech_def_attack", 1000);
	Core::getTPL()->assign("cfg_tech_def_shield", 1000);
	
		$i = 1;
		foreach($tech_table as $tech_line)
		{
			$j = 1;
			foreach($tech_line as $tech_scale)
			{
				Core::getTPL()->assign("cfg_tech_{$i}_{$j}", $tech_scale);
				Core::getTPL()->assign("cfg_tech_{$i}_{$j}_procent", round($tech_scale*100)."%");
				$j++;
			}
			$i++;
		}
		$i = 1;
		foreach(array(UNIT_LASER_TECH, UNIT_ION_TECH, UNIT_PLASMA_TECH) as $tech)
		{
			Core::getTPL()->assign("a_tech_$i", NS::getResearch($tech));
			Core::getTPL()->assign("d_tech_$i", NS::getResearch($tech));
			$i++;
		}
		// Core::getTPL()->addLoop("moon", $moon);
		Core::getTPL()->display("adv_tech_calc");
		return $this;
	}
}
?>