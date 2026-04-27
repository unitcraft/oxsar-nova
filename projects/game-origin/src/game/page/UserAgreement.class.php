<?php
/**
* Displays User Agreement
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined('APP_ROOT_DIR')) { die('Hacking attempt detected.'); }

class UserAgreement extends Page
{
	/**
	* Constructor. Handles requests for this page.
	*
	* @return void
	*/
	public function __construct()
	{
		parent::__construct();
		$this->setPostAction("agree", "actionAgree");
		Core::getLanguage()->load('info,UserAgreement');
		$this->proceedRequest();
		return;
	}
	
	protected function index()
	{
		$is_new = getArgeementTime() > NS::getUser()->get('user_agreement_read');
		Core::getTPL()->addLoop('agreemet', $this->getChildAgreements());
		Core::getTPL()->assign('is_new_agreemet', $is_new);
		Core::getTPL()->assign('new_agreemet_pre_text', $is_new ? Core::getLanguage()->getItem('USER_AGREEMENT_NEW_PRE_TEXT') : '');
		Core::getTPL()->assign('agreemet_pre_text', Core::getLanguage()->getItem('USER_AGREEMENT_PRE_TEXT'));
		Core::getTPL()->assign('agreemet_post_text', Core::getLanguage()->getItem('USER_AGREEMENT_POST_TEXT'));
		Core::getTPL()->display('user_agreemet');
		return $this;
	}
	
	protected function actionAgree()
	{
		// План 37.5d.5#5: replaced User_YII::model()->updateByPk()
		sqlUpdate("user", array("user_agreement_read" => time()),
			"userid=".sqlVal($_SESSION["userid"] ?? 0));
		doHeaderRedirection("game.php/Main", false);
	}

	protected function getChildAgreements( $id = NULL, $parent_depth_str = "", $logic_depth = 0 )
	{
		// План 37.5d.5#5: replaced UserAgreement_YII + CDbCriteria.
		// Логика: ищем строки na_user_agreement с указанным parent_id (или
		// IS NULL для корня) и lang = текущий. Если для текущего языка ничего
		// нет — fallback на DEF_LANGUAGE_ID.
		$lang = $_SESSION['languageid'] ?? DEF_LANGUAGE_ID;
		$parent_clause = ($id === NULL) ? 'parent_id IS NULL' : 'parent_id=' . sqlVal($id);

		$loadAgreements = function($lang_id) use ($parent_clause) {
			$rows = array();
			$result = sqlSelect("user_agreement", "*", "",
				$parent_clause . " AND lang=" . sqlVal($lang_id),
				"display_order ASC, id ASC");
			while($row = sqlFetch($result))
			{
				$rows[] = $row;
			}
			sqlEnd($result);
			return $rows;
		};

		$agreements = $loadAgreements($lang);
		if(!$agreements && $lang != DEF_LANGUAGE_ID)
		{
			$agreements = $loadAgreements(DEF_LANGUAGE_ID);
		}

		$i = 1;
		$result = array();
		foreach( $agreements as $key => $agreement )
		{
			foreach( $agreement as $k => $v)
			{
				if( $k == 'date' )
				{
					$result[$key][$k.'_nix'] = strtotime($v);
				}
				$result[$key][$k] = $v;
			}
			$result[$key]["is_new"] = $result[$key]['date_nix'] > NS::getUser()->get('user_agreement_read');
			switch($result[$key]["type"])
			{
			case 0:
				$result[$key]["depth_str"] = "";
				$sub_depth_str = "";
				$sub_logic_depth = $logic_depth + 1;
				break;
				
			default:
			case 1:
				$result[$key]["depth_str"] = $parent_depth_str . $i++ . ".";
				$sub_depth_str = $result[$key]["depth_str"];
				$sub_logic_depth = $logic_depth + 1;
				break;
				
			case 2:
				$result[$key]["depth_str"] = "-";
				$sub_depth_str = "";
				$sub_logic_depth = $logic_depth;
				break;
			}
			$result[$key]["logic_depth"] = $logic_depth;
			$result[$key]['childs'] = $this->getChildAgreements($agreement['id'], $sub_depth_str, $sub_logic_depth);
		}
		return $result;
	}
}
