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
		// План 37.5d.5#5: replaced TutorialStatesCategory_YII->with('states')->findAll().
		// FK: tutorial_states.category → tutorial_states_category.id.
		// Yii relation 'states' = категория HAS_MANY tutorial_states по category.
		$result = array();
		$rows = Core::getDB()->query(
			"SELECT c.id AS cat_id, c.title AS cat_title,"
			. " s.id AS state_id, s.formaction AS state_formaction, s.name AS state_name"
			. " FROM ".PREFIX."tutorial_states_category c"
			. " LEFT JOIN ".PREFIX."tutorial_states s ON s.category = c.id"
			. " ORDER BY c.id ASC, s.display_order ASC"
		);
		$counters = array(); // state index per category
		while( $row = Core::getDB()->fetch($rows) )
		{
			$cat_id = $row["cat_id"];
			if( !isset($result[$cat_id]) )
			{
				$result[$cat_id]['number'] = $cat_id;
				$result[$cat_id]['title']  = Core::getLang()->getItem($row["cat_title"]);
				$counters[$cat_id] = 0;
			}
			if( !empty($row["state_id"]) )
			{
				$counters[$cat_id]++;
				$i = $counters[$cat_id];
				$result[$cat_id]['states'][$i] = array(
					'number'     => $cat_id . '.' . $i,
					'id'         => $row["state_id"],
					'formaction' => $row["state_formaction"],
					'title'      => Core::getLang()->getItem($row["state_name"] . '_QUESTION'),
				);
			}
		}
		Core::getDB()->free_result($rows);

		Core::getTPL()->addLoop('avaliable_tutorials', $result);
		Core::getTPL()->display('tutorials');
		return $this;
	}
}
