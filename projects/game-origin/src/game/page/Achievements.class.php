<?php
/**
* Displays Achievements
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined('APP_ROOT_DIR')) { die('Hacking attempt detected.'); }

class Achievements extends Page
{
	protected $pages2show = 3;
	protected $per_page = 10;
	/**
	* Constructor. Handles requests for this page.
	*
	* @return void
	*/
	public function __construct()
	{
		parent::__construct();
			$this
				->setGetAction('go', 'AchievementInfo', 'achievementInfo')
				->addGetArg('achievementInfo', 'id')
				->setGetAction('go', 'AchievementHideAjax', 'achievementHideAjax')
				->addGetArg('achievementHideAjax', 'id')
				->setGetAction('go', 'AchievementGetBonus', 'achievementGetBonus')
				->addPostArg('achievementGetBonus', 'id')
				->setGetAction('go', 'AchievementProcess', 'achievementProcess')
				->addPostArg('achievementProcess', 'id')
				->setGetAction('go', 'AchievementsRecalc', 'achievementsRecalc')
				->setGetAction('go', 'AchievementsProfile', 'achievementsProfile')
				->addGetArg('achievementsProfile', 'id')
				->addPostArg('achievementsProfile', 'page')
				->setGetAction('go', 'AchievementsDone', 'achievementsDone')
				->addPostArg('achievementsDone', 'page')
				->setGetAction('go', 'AchievementsAvaliable', 'achievementsAvaliable')
				->addPostArg('achievementsAvaliable', 'page')
				;
		Core::getLanguage()->load('info,Achievements');
		$this->proceedRequest();
		return;
	}
	
	/**
	 * Index Action. Shows all achievements, owned by user and avaliable to user
	 */
	protected function index($page = 0)
	{
		$paginator = array(
			'paginator' => array(
				'enabled'	=> true,
				'pages2show'=> $this->pages2show,
				'page'		=> $page,
				'per_page'	=> $this->per_page,
			)
		);
		AchievementsService::loadAchievementsTemplateData($paginator);
		Core::getTPL()->display('blank');
		return $this;
	}
	
	protected function achievementsDone($page = 1)
	{
		if( empty($page) )
		{
			$page = 1;
		}
		$paginator = array(
			'paginator' => array(
				'enabled'	=> true,
				'pages2show'=> $this->pages2show,
				'page'		=> $page,
				'per_page'	=> $this->per_page,
				'user_got'	=> true,
			)
		);
		AchievementsService::loadAchievementsTemplateData($paginator);
		Core::getTPL()->assign('done_achi', true);
		Core::getTPL()->assign('aval_achi', false);
		Core::getTPL()->assign('page', $page);
		Core::getTPL()->display('blank');
		return $this;
	}
	
	protected function achievementsAvaliable($page = 0)
	{
		if( empty($page) )
		{
			$page = 1;
		}
		$paginator = array(
			'skip_done' => true,
			'paginator' => array(
				'enabled'	=> true,
				'pages2show'=> $this->pages2show,
				'page'		=> $page,
				'per_page'	=> $this->per_page,
				'user_got'	=> false,
			)
		);
		AchievementsService::loadAchievementsTemplateData($paginator);
		Core::getTPL()->assign('done_achi', false);
		Core::getTPL()->assign('aval_achi', true);
		Core::getTPL()->assign('page', $page);
//		Core::getTPL()->assign('achi_paginator', ( $page == 0 ? false:true ));
		Core::getTPL()->assign('achi_paginator', true);
		Core::getTPL()->display('blank');
		return $this;
	}
	
	/**
	 * Get achievement bonus
	 * @param int $achievement_id Achievement ID.
	 */
	protected function achievementsProfile( $id , $page)
	{
		$params = array(
			'user_id' => max(0, (int)$id),
			'paginator' => array(
				'enabled'	=> true,
				'pages2show'=> $this->pages2show,
				'page'		=> $page,
				'per_page'	=> $this->per_page,
				'user_got'	=> true,
			)
		);
		AchievementsService::loadAchievementsTemplateData($params);
		Core::getTPL()->assign( 'done_formaction', RELATIVE_URL . 'game.php/AchievementsProfile/'.$id );
		Core::getTPL()->assign('done_achi', true);
		Core::getTPL()->assign('aval_achi', false);
		Core::getTPL()->display('blank');
	}
	
	/**
	 * Shows detailed info about achievement
	 * @param int $achievement_id Achievement ID.
	 */
	protected function achievementInfo( $achievement_id )
	{
		AchievementsService::loadAchievementsTemplateData(array(
			"achievement_id" => max(0, (int)$achievement_id),
		));
		Core::getTPL()->display('blank');
	}
	
	/**
	 * Hides achievement from all pages. Used in AJAX
	 * @param int $achievement_id Achievement ID.
	 */
	protected function achievementHideAjax( $achievement_id )
	{
		$achievement_id = max(0, (int)$achievement_id);
		$user_id = NS::getUser()->get('userid');
		if(!NS::isFirstRun("AchievementsService::state:{$user_id}-{$achievement_id}"))
		{
			// Logger::dieMessage("TOO_MANY_REQUESTS");
			error_log('TOO_MANY_REQUESTS in achievementHideAjax, achivement_id: ' . $achievement_id );
			exit();
			return;
		}
		$planet_id = NS::getUser()->get("curplanet");
		AchievementsService::setAchievementState( $user_id, $planet_id, $achievement_id, ACHIEV_STATE_HIDDEN );
		exit();
	}
	
	protected function redirectBack()
	{
		doHeaderRedirection("game.php/Achievements", false);
	}
	
	/**
	 * Get achievement bonus
	 * @param int $achievement_id Achievement ID.
	 */
	protected function achievementGetBonus( $achievement_id )
	{
		AchievementsService::setAchievementState( null, null, max(0, (int)$achievement_id), ACHIEV_STATE_BONUS_GIVEN );
		$this->redirectBack();
	}
	
	/**
	 * Process achievement
	 * @param int $achievement_id Achievement ID.
	 */
	protected function achievementProcess( $achievement_id )
	{
		AchievementsService::setAchievementState( null, null, max(0, (int)$achievement_id), ACHIEV_STATE_PROCESSED );
		$this->redirectBack();
	}
	
	/**
	 * Recalc achievements
	 */
	protected function achievementsRecalc()
	{
		$user_id = NS::getUser()->get('userid');
		if(!NS::isFirstRun("AchievementsService::recalc:{$user_id}"))
		{
			Logger::dieMessage("TOO_MANY_REQUESTS");
		}
		$planet_id = NS::getUser()->get("curplanet");
		AchievementsService::processAchievements( $user_id, $planet_id );
		$this->redirectBack();
	}
}