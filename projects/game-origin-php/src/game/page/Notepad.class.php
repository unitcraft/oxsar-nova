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
		// План 37.5d.5#4: replaced Notes_YII::model()->findByPk() + CDbCriteria.
		$user_id = $_SESSION["userid"] ?? 0;
		$notes = sqlSelectField("notes", "notes", "", "user_id=".sqlVal($user_id));
		// 37.7.3: notes идёт в <textarea> через {@notes} (не эскейпается шаблоном) — экранируем чтобы </textarea> не сломал разметку
		Core::getTPL()->assign('notes', htmlspecialchars((string)($notes ?? ''), ENT_QUOTES, 'UTF-8'));
		Core::getTPL()->assign('formaction', socialUrl(RELATIVE_URL . 'game.php/SaveNotes'));
		Core::getTPL()->display('notes');
		return $this;
	}

	protected function saveNotes($notes = '')
	{
		if( trim($notes) )
		{
			// План 37.5d.5#4: replaced Notes_YII active record (findByPk + new + save).
			$user_id = $_SESSION["userid"] ?? 0;
			$existing = sqlSelectField("notes", "user_id", "", "user_id=".sqlVal($user_id));
			if( $existing !== null )
			{
				sqlUpdate("notes", array("notes" => $notes), "user_id=".sqlVal($user_id));
			}
			else
			{
				sqlInsert("notes", array("user_id" => $user_id, "notes" => $notes));
			}
		}
		// return $this->index();
		doHeaderRedirection("game.php/Notepad", false);
	}
}