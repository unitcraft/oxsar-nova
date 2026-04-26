<?php
/**
* Allows the user to change the preferences.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class Profession extends Page
{
	/**
	* Shows the preferences form.
	*
	* @return void
	*/
	public function __construct()
	{
		parent::__construct();
        $this
            ->setPostAction("save", "changeProfession")
            ->addPostArg("changeProfession", "profession")
        ;
		$this->proceedRequest();
	}

    protected function changeProfession($id)
    {
        if(NS::getUser()->get("umode")){
            Logger::dieMessage('UMODE_ENABLED');
            // doHeaderRedirection("game.php/Stock", false);
        }
        $id = max(0, (int)$id);
        if(NS::getUser()->get("profession") != $id
            && isset($GLOBALS["PROFESSIONS"][$id])
            && NS::isFirstRun("changeProfession($id)".NS::getUser()->get("userid")))
        {
            $profession_change_cost = NS::getProfessionChangeCost();
            if($profession_change_cost > 0 && NS::getUser()->get("credit") < $profession_change_cost){
                Logger::dieMessage('NO_CREDIT_TO_CHANGE_PROFESSION');
            }
            if($profession_change_cost > 0){
                NS::updateUserRes(array(
                    // 'block_minus'	=> true,
                    'type'			=> RES_UPDATE_CHANGE_PROFESSION,
                    'reload_planet' => false,
                    'userid'		=> NS::getUser()->get('userid'),
                    // 'planetid'		=> NS::getUser()->get('curplanet'),
                    'credit'		=> - $profession_change_cost,
                ));

				// if(empty($res_log["minus_blocked"]))
				{
					new AutoMsg(
						MSG_CREDIT,
						NS::getUser()->get("userid"),
						time(),
						array(
							'credits'	=> $profession_change_cost,
							'msg' 		=> 'MSG_CREDIT_PROFESSION_CHANGED',
							)
					);
				}
            }
            // NS::applyProfession(NS::getUser()->get("userid"), NS::getUser()->get("profession"), false);

            sqlUpdate('user', array(
                'profession' => $id,
                'prof_time' => time(),
            ), 'userid='.sqlUser());

            // NS::applyProfession(NS::getUser()->get("userid"), $id);

            NS::updateProfession(NS::getUser()->get("userid"));
        }
        doHeaderRedirection("game.php/Profession", false);
    }

    /**
	* Index action.
	*
	* @return Preferences
	*/
	protected function index()
	{
        $tech_list = array();
        $professions = array();
        foreach($GLOBALS["PROFESSIONS"] as $id => $profession){
            $prof = array(
                'id' => $id,
                'selected' => NS::getUser()->get("profession") == $id,
                'name' => Core::getLanguage()->getItem($profession['name']),
                'desc' => Core::getLanguage()->getItem($profession['name'].'_DESC'),
            );
            if(isset($profession['tech_special'])){
                foreach($profession['tech_special'] as $tech_id => $level_diff){
                    if(!isset($tech_list[$tech_id])){
                        $tech_list[$tech_id] = sqlSelectField('construction',
                                'name', '', 'buildingid='.sqlVal($tech_id));
                    }
                    $tech_name = $tech_list[$tech_id];
                    $prof['tech_special'][] = array(
                        'name' => Core::getLanguage()->getItem($tech_name),
                        'level_diff' => $level_diff,
                    );
                }
            }
            $professions[] = $prof;
        }
        $profession_change_cost = NS::getProfessionChangeCost();
        if($profession_change_cost > 0){
            $profession_change_info = Core::getLanguage()->getItemWith('PROFESSION_CHANGE_COST_INFO', array(
                'days_remain' => NS::getProfessionChangeDaysRemain(),
                'cost' => $profession_change_cost,
            ));
        }else{
            $profession_change_info = Core::getLanguage()->getItemWith('PROFESSION_CHANGE_NO_COST_INFO', array(
                'days_remain' => PROFESSION_CHANGE_MIN_DAYS,
                'cost' => PROFESSION_CHANGE_COST,
            ));
        }
		Core::getTPL()->assign("current_profession", NS::getProfessionName());
        Core::getTPL()->assign('profession_change_info', $profession_change_info);
		Core::getTPL()->addLoop("professions", $professions);

		Core::getTPL()->display("profession");
		return $this;
	}
}
?>