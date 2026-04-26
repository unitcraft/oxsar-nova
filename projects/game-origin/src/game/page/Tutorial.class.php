<?php
/**
* Displays Tutorials
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined('APP_ROOT_DIR')) { die('Hacking attempt detected.'); }

class Tutorial extends Page
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
		$crit = new CDbCriteria();
		$crit->order = 't.id ASC, states.display_order ASC';
		$tutorials = TutorialStatesCategory_YII::model()->with('states')->findAll($crit);
		if( empty($tutorials) )
		{
			$result = array();
		}
		else
		{
			foreach( $tutorials as $tutorial )
			{
				$id = $tutorial->id;
				$result[$id]['number']	= $id;
				$result[$id]['title']	= Core::getLang()->getItem($tutorial->title);
				if( $tutorial->states )
				{
					$i = 0;
					foreach( $tutorial->states as $state )
					{
						$i++;
						$result[$id]['states'][$i]['number']	= $id . '.' . $i;
						$result[$id]['states'][$i]['id']		= $state->id;
						$result[$id]['states'][$i]['formaction']= $state->formaction;
						$result[$id]['states'][$i]['title']		= Core::getLang()->getItem($state->name . '_QUESTION');
					}
				}
			}
		}
		$tutorials = $result;
		Core::getTPL()->addLoop('avaliable_tutorials', $tutorials);
		Core::getTPL()->display('tutorials');
		return $this;
	}
}
