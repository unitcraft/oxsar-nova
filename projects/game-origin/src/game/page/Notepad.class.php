<?php
/**
* Displays Notepad
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined('APP_ROOT_DIR')) { die('Hacking attempt detected.'); }

class Notepad extends Page
{
	/**
	* Constructor. Handles requests for this page.
	*
	* @return void
	*/
	public function __construct()
	{
		parent::__construct();
		$this
			->setGetAction('go', 'SaveNotes', 'saveNotes')
			->addPostArg('saveNotes', 'notes');
			
		Core::getLanguage()->load('info');
		$this->proceedRequest();
		return;
	}
	
	/**
	 * Index Action. Shows notes area
	 */
	protected function index()
	{
		$crit = new CDbCriteria();
		$note = Notes_YII::model()->findByPk($_SESSION["userid"] ?? 0);
		Core::getTPL()->assign('notes', $note->notes);
		Core::getTPL()->assign('formaction', socialUrl(RELATIVE_URL . 'game.php/SaveNotes'));
		Core::getTPL()->display('notes');
		return $this;
	}
	
	protected function saveNotes($notes = '')
	{
		if( trim($notes) )
		{
			$note = Notes_YII::model()->findByPk($_SESSION["userid"] ?? 0);
			if( !$note )
			{
				$note = new Notes_YII();
				$note->user_id = $_SESSION["userid"] ?? 0;
			}
			$note->notes= $notes;
			$note->save(false);
		}
		// return $this->index();
		doHeaderRedirection("game.php/Notepad", false);
	}
}