<?php
/**
 * Achievements system.
 *
 * Oxsar http://oxsar.ru
 *
 *
 */

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class AchievementsService
{
	protected static function checkReqExchRes($user_id, $planet_id = null, $achievemnt_id = null)
	{
		if( $user_id == null )
		{
			return true;
		}
		// План 37.5d.5#2: replaced UserStates_YII::model()->findByPk()
		$val = sqlSelectField("user_states", "exchanged_ress", "", "user_id=".sqlVal($user_id));
		if($val !== null)
		{
			if($val > 0)
			{
				return true;
			}
		}
		else
		{
			return false;
		}
	}

	protected static function getReqNameExchRes()
	{
		return Core::getLang()->getItem('STATE_RES_UPDATE_EXCHANGE');
	}

	protected static function checkReqGotSim($user_id, $planet_id = null, $achievemnt_id = null)
	{
		if( $user_id == null )
		{
			return true;
		}
		// План 37.5d.5#2: replaced UserStates_YII::model()->findByPk()
		$val = sqlSelectField("user_states", "simulated_assault", "", "user_id=".sqlVal($user_id));
		if($val !== null)
		{
			if($val >= 1)
			{
				return true;
			}
		}
		else
		{
			return false;
		}
	}

	protected static function getReqNameGotSim()
	{
		return Core::getLang()->getItem('STATE_ASSAULT_SIMULATION');
	}

	/**
	 * Opens new achievements
	 * @param int $user_id User ID
	 */
	public static function openNewAchievements( $user_id )
	{
		sqlQuery('INSERT INTO '.PREFIX.'achievements2user (user_id, achievement_id, state, created)
			SELECT '.sqlVal($user_id).', ach.achievement_id, '.sqlVal(ACHIEV_STATE_ALERT).', '.sqlVal(time()).'
			FROM '.PREFIX.'achievement_datasheet ach
			WHERE ach.achievement_id NOT IN (SELECT achievement_id FROM '.PREFIX.'achievements2user WHERE user_id = '.sqlVal($user_id).')
				AND (
						SELECT CEIL(count(*) * 0.4) FROM '.PREFIX.'requirements ir
						JOIN '.PREFIX.'achievement_datasheet iach ON iach.achievement_id = ir.needs
						WHERE ir.buildingid = ach.achievement_id
					) <= (
						SELECT count(*) FROM '.PREFIX.'requirements ir
						JOIN '.PREFIX.'achievement_datasheet iach ON iach.achievement_id = ir.needs
						JOIN '.PREFIX.'achievements2user ia ON ia.achievement_id = iach.achievement_id
						WHERE ir.buildingid = ach.achievement_id
							AND ia.user_id='.sqlVal($user_id).'
							AND ia.granted >= ('.sqlVal(time()).' - iach.time)
							AND ia.bonus_planet_id IS NOT NULL
					)'
			);
	}

	/**
	 * Creates Achievement for User on Planet
	 * @param int $user_id User ID, who got this achievement
	 * @param int $planet_id Planet ID, on witch this achievement was made. For non-planet achievements use null
	 */
	public static function processAchievements($user_id, $planet_id)
	{
        if(!ACHIEVEMENTS_ENABLED){
            return;
        }

		self::openNewAchievements($user_id);

		$result = sqlSelect('achievements2user a2u',
			'a2u.*, ads.*, u.*',
			'	JOIN '.PREFIX.'achievement_datasheet ads ON a2u.achievement_id = ads.achievement_id
				JOIN '.PREFIX.'user u ON u.userid = a2u.user_id',
			'a2u.user_id = '.sqlVal($user_id).' AND a2u.granted < ('.sqlVal(time()).' - ads.time)');
		while ( $row = sqlFetch($result) )
		{
			$achievement_id = $row['achievement_id'];
			$is_req_fields_ok = true;
			foreach( $row as $key => $val )
			{
				if( substr($key, 0, 4) == 'req_' && $val > 0 )
				{
					$field = substr($key, 4);
					if( isset($row[$field]) && is_numeric($row[$field]) && ($val - $row[$field]) > 0 )
					{
						$is_req_fields_ok = false;
						break;
					}
				}
				elseif( substr($key, 0, 11) == 'custom_req_' && !empty($val) )
				{
					$check_funct_name = 'checkReq' . $val;
					if( !method_exists('Achievements', $check_funct_name) )
					{
						error_log('Can\'t find custom functon ' . $check_funct_name, 'warning');
						$is_req_fields_ok = false;
						break;
					}
					if( !self::$check_funct_name($user_id, $planet_id, $achievement_id) )
					{
						$is_req_fields_ok = false;
						break;
					}
				}
			}
			if( $is_req_fields_ok && NS::checkRequirements($achievement_id, $user_id, $planet_id) )
			{
				// Update By Uk
				sqlQuery( 'UPDATE ' . PREFIX . 'achievements2user SET
					granted = '.sqlVal(time()).'
					, granted_planet_id = '.sqlVal($planet_id).'
					, state = '.sqlVal(ACHIEV_STATE_ALERT).'
					, quantity = quantity + 1
					WHERE user_id = '.sqlVal($user_id).' AND achievement_id = '.sqlVal($achievement_id)
					// . ' ORDER BY id'
					);
			}
		}
		sqlEnd($result);
	}

	/**
	 * Checks if user has  this achievement
	 * @param int $user_id User ID
	 * @param int $achievement_id Achievement ID
	 */
	public static function isGrantedAchievement($user_id, $achievement_id)
	{
		static $cache = array();
		if(!isset($cache[$user_id][$achievement_id]))
		{
			$cnt = sqlSelectField(
				'achievements2user a2u',
				'count(*)',
				'JOIN '.PREFIX.'achievement_datasheet ads ON a2u.achievement_id = ads.achievement_id',
				'a2u.user_id = '.sqlVal( $user_id )
					. ' AND a2u.achievement_id = '.sqlVal( $achievement_id )
					. ' AND a2u.granted >= ('.sqlVal(time()).' - ads.time)'
			);
			$cache[$user_id][$achievement_id] = $cnt;
		}
		return $cnt == 1;
	}

	/**
	 * Get achievement quantity
	 * @param int $user_id User ID
	 * @param int $achievement_id Achievement ID
	 */
	public static function getAchievementQuantity($user_id, $achievement_id)
	{
		$quantity = sqlSelectField(
			'achievements2user a2u',
			'a2u.quantity',
			'',
			'a2u.user_id = '.sqlVal( $user_id ).' AND a2u.achievement_id = '.sqlVal( $achievement_id )
		);
		return $quantity;
	}

	/**
	 * Get processed achievement quantity
	 * @param int $user_id User ID
	 * @param int $achievement_id Achievement ID
	 */
	public static function getProcessedAchievementQuantity($user_id, $achievement_id)
	{
		$quantity = sqlSelectField(
			'achievements2user a2u',
			'a2u.quantity - case when a2u.bonus_planet_id > 0 then 0 else 1 end as quantity',
			'',
			'a2u.user_id = '.sqlVal( $user_id )
				.' AND a2u.achievement_id = '.sqlVal( $achievement_id )
				.' AND a2u.granted > 0'
		);
		return $quantity;
	}

	/**
	 * Marks Achievement as seen and saves current timestamp
	 * @param int $user_id User ID
	 * @param int $planet_id Planet ID
	 * @param int $achievement_id Achievement ID
	 * @param int $state New achievement state
	 */
	public static function setAchievementState( $user_id, $planet_id, $achievement_id, $state )
	{
		if(empty($user_id))
		{
			if( !NS::getUser() )
			{
				return;
			}
			$user_id = NS::getUser()->get('userid');
			$planet_id = NS::getUser()->get("curplanet");
			if( !NS::isFirstRun("AchievementsService::state:{$user_id}-{$achievement_id}") )
			{
				return;
			}
		}
		$is_planet_moon = sqlSelectField("planet", "ismoon", "", "planetid = ".sqlVal($planet_id));

		$row = sqlSelectRow(
			'achievements2user a2u',
			array('a2u.*', 'ads.*', self::getSqlBonusBuildTypeField()),
			'JOIN '.PREFIX.'achievement_datasheet ads ON a2u.achievement_id = ads.achievement_id'
			. ' LEFT JOIN '.PREFIX.'construction c1 ON c1.buildingid = ads.bonus_1_unit_id'
			. ' LEFT JOIN '.PREFIX.'construction c2 ON c1.buildingid = ads.bonus_2_unit_id'
			. ' LEFT JOIN '.PREFIX.'construction c3 ON c1.buildingid = ads.bonus_3_unit_id'
			,
			'a2u.user_id = '.sqlVal($user_id).' AND a2u.achievement_id = '.sqlVal($achievement_id)
		);
		if( !empty($row) )
		{
			if(empty($row['granted']))
			{
				$state = ACHIEV_STATE_HIDDEN;
			}
			else if($state == ACHIEV_STATE_HIDDEN)
			{
				$state = ACHIEV_STATE_PROCESSED;
			}

			if( $state <= $row['state'] )
			{
				return;
			}

			// $planet_id = $row['planet_id'];
			$is_activate = !empty($row['granted']) && empty($row['bonus_planet_id']) && in_array($state, array(ACHIEV_STATE_BONUS_GIVEN, ACHIEV_STATE_PROCESSED));
			if($is_activate)
			{
				if( isset($row['bonus_build_type'])
					&& ( $row['bonus_build_type'] != ACHIEV_BONUS_BUILD_TYPE_ANY
						&& ( $row['bonus_build_type'] == ACHIEV_BONUS_BUILD_TYPE_ERROR
								|| ($row['build_bonus_type'] == ACHIEV_BONUS_BUILD_TYPE_MOON) != $is_planet_moon ) ) )
				{
					$is_activate = false;
					$state = ACHIEV_STATE_HIDDEN;
				}
				else
				{
					NS::updateUserRes(
						array(
							'type'			=> RES_UPDATE_ACHIEVEMENT,
							'reload_planet'	=> false,
							'userid'		=> $user_id,
							'planetid'		=> $planet_id,
							'metal'			=> $row['bonus_metal'],
							'silicon'		=> $row['bonus_silicon'],
							'hydrogen'		=> $row['bonus_hydrogen'],
							'credit'		=> $row['bonus_credit'],
						)
					);

					sqlQuery( 'UPDATE ' . PREFIX . 'user SET '
						. ' a_points = a_points+'.sqlVal($row['points'])
						. ', a_count = a_count+1 '
						. ' WHERE userid = '.sqlVal( $user_id ) );

					foreach( array('bonus_1', 'bonus_2', 'bonus_3') as $sufix )
					{
						$id_name	= $sufix . '_unit_id';
						$level_name = $sufix . '_unit_level';
						if( !empty( $row[$id_name] ) && !empty( $row[$level_name] ) )
						{
							$type = sqlSelectField('construction', 'mode', '', 'buildingid = '.sqlVal( $row[$id_name] ));
							if( empty($type) )
							{
								continue;
							}
							if(
								( $type == UNIT_TYPE_CONSTRUCTION || $type == UNIT_TYPE_MOON_CONSTRUCTION )
								&& ( NS::getBuilding($row[$id_name], $planet_id) < $row[$level_name] )
							)
							{
								self::addBuildingBonus($user_id, $planet_id, $row[$id_name], $row[$level_name]);
							}
							elseif(
								( $type == UNIT_TYPE_RESEARCH )
								&& ( NS::getResearch($row[$id_name], $user_id) < $row[$level_name] )
							)
							{
								self::addResearchBonus($user_id, $row[$id_name], $row[$level_name]);
							}
							elseif(
								( $type == UNIT_TYPE_FLEET || $type == UNIT_TYPE_DEFENSE )
							)
							{
								self::addUnitBonus($user_id, $planet_id, $row[$id_name], $row[$level_name]);
							}
							elseif( ( $type == UNIT_TYPE_ARTEFACT ) )
							{
								self::addArtefactBonus($user_id, $planet_id, $row[$id_name], $row[$level_name]);
							}
						}
					}
				}
			}
			// Update By Uk
			sqlUpdate(
				'achievements2user',
				array('state' => $state) + ($is_activate ? array('bonus_planet_id' => $planet_id) : array()),
				'user_id = '.sqlVal($user_id).' AND achievement_id = '.sqlVal($achievement_id)
				// . ' ORDER BY id'
			);
			if($is_activate)
			{
				self::processAchievements($user_id, $planet_id);
			}
		}
	}

	/**
	 * Grants Building to User
	 * @param int $user_id User ID
	 * @param int $planet_id Planet ID
	 * @param int $bid Building ID
	 * @param int $level Building level
	 */
	protected static function addBuildingBonus($user_id, $planet_id, $bid, $level)
	{
		if($level < 1)
		{
			return;
		}
		$data = sqlSelectRow('construction', '*', '', ' buildingid = ' . sqlVal( $bid ));

		$add_row = sqlSelectRow(
			'building2planet',
			'level, added',
			'',
			'buildingid = '.sqlVal( $bid )
				. ' AND planetid = '.sqlVal( $planet_id )
		);
		$insert = (!empty($add_row)) ? false : true;
		if($insert)
		{
			sqlInsert(
				'building2planet',
				array(
					'planetid'		=> $planet_id,
					'buildingid'	=> $bid,
					'level'			=> $level
				)
			);

			if( $bid == UNIT_EXCHANGE )
			{
				sqlInsert('exchange', array(
					'eid'		=> $planet_id,
					'uid'		=> $user_id,
					'title'		=> Core::getLang()->get('MENU_STOCK'),
					'fee'		=> EXCH_FEE_MIN,
					'def_fee'	=> EXCH_FEE_MIN,
					'comission'	=> EXCH_COMMISSION_MIN));
			}
		}
		else
		{
			if($level <= $add_row['level'] - $add_row['added'])
			{
				return;
			}
			// Update By Uk
			sqlUpdate('building2planet', array(
					'level' => $level + $add_row['added']
				), 'buildingid = '.sqlVal( $bid ).' AND planetid = '.sqlVal( $planet_id )
				// . ' ORDER BY buildingid, planetid'
			);
		}

		$res_data = self::getNeededResources( $add_row['level'], $level + $add_row['added'], $data );
		$points = round(($res_data['metal'] + $res_data['silicon'] + $res_data['hydrogen']) * RES_TO_BUILD_POINTS, POINTS_PRECISION);

		// Update By Pk
		sqlQuery('UPDATE '.PREFIX.'user SET '
			. ' points = points + '.sqlVal($points)
			. ', b_points = b_points + '.sqlVal($points)
			. ', b_count = b_count + '.sqlVal($level - $add_row['level'])
            . ", ".updateDmPointsSetSql()
			. ' WHERE userid = '.sqlVal($user_id)
			// . ' ORDER BY userid'
		);

		if(in_array($data['buildingid'], array(UNIT_METALMINE, UNIT_SILICON_LAB, UNIT_HYDROGEN_LAB)))
		{
			updateMinerPoints($user_id, $points, true);
		}
	}

	/**
	 * Grants Reserch to User
	 * @param int $user_id User ID
	 * @param int $rid Research ID
	 * @param int $level Research level
	 */
	protected static function addResearchBonus($user_id, $rid, $level)
	{
		if($level < 1)
		{
			return;
		}
		$data = sqlSelectRow('construction', '*', '', ' buildingid = ' . sqlVal( $rid ) );

		$add_row = sqlSelectRow('research2user', 'level, added', '', 'buildingid = '.sqlVal( $rid ).' AND userid = '.sqlVal( $user_id ));
		$insert = (!empty($add_row)) ? false : true;
		if( $insert )
		{
			sqlInsert(
				'research2user',
				array(
					'buildingid' => $rid,
					'userid' => $user_id,
					'level' => $level
				)
			);
		}
		else
		{
			if($level <= $add_row['level'] - $add_row['added'])
			{
				return;
			}
			// Update By Pk
			sqlUpdate('research2user', array(
					'level' => $level + $add_row['added']
				),
				'buildingid = '.sqlVal( $rid ).' AND userid = '.sqlVal( $user_id )
					// . ' ORDER BY buildingid, userid'
			);
		}

		$res_data = self::getNeededResources( $add_row['level'], $level + $add_row['added'], $data );
		$points = round(($res_data['metal'] + $res_data['silicon'] + $res_data['hydrogen']) * RES_TO_RESEARCH_POINTS, POINTS_PRECISION);

		// Update By Pk
		sqlQuery('UPDATE '.PREFIX.'user SET '
			. ' points = points + '.sqlVal($points)
			. ', r_points = r_points + '.sqlVal($points)
			. ', r_count = r_count + '.sqlVal($level - $add_row['level'])
            . ", ".updateDmPointsSetSql()
			. ' WHERE userid = '.sqlVal($user_id)
				// . ' ORDER BY userid'
		);
	}

	protected static function addArtefactBonus($user_id, $planet_id, $unit_id, $quantity)
	{
		Artefact::appear($unit_id, $user_id, $planet_id, array());
	}

	/**
	 * Grants fleet unit to player
	 * @param int $user_id User ID
	 * @param int $planet_id Planet ID on witch this fleet will be placed
	 * @param int $unit_id Unit ID that will be placed
	 * @param int $quantity How many Units will be placed
	 */
	protected static function addUnitBonus($user_id, $planet_id, $unit_id, $quantity)
	{
		$data = sqlSelectRow('construction', '*', '', ' buildingid = ' . sqlVal( $unit_id ) );
		logShipsChanging($unit_id, $planet_id, $quantity, 1, '[achievement]');
		if( sqlSelectField('unit2shipyard', 'count(*)', '', 'unitid = '.sqlVal( $unit_id ).' AND planetid = '.sqlVal( $planet_id )) )
		{
			// Update By Pk
			sqlQuery(
				'UPDATE '
				. PREFIX
				. 'unit2shipyard SET quantity = quantity + '.sqlVal( $quantity )
				. ' WHERE planetid = '.sqlVal( $planet_id ).' AND unitid = '.sqlVal( $unit_id )
				// . ' ORDER BY planetid, unitid'
			);
		}
		else
		{
			sqlInsert('unit2shipyard', array('unitid' => $unit_id, 'planetid' => $planet_id, 'quantity' => $quantity));
		}
		$points = round(($data['basic_metal'] + $data['basic_silicon'] + $data['basic_hydrogen']) * RES_TO_UNIT_POINTS, POINTS_PRECISION) * $quantity;
		// Update By Pk
		sqlQuery(
			'UPDATE '.PREFIX.'user SET '
			. ' points = points + '.sqlVal($points)
			. ', u_points = u_points + '.sqlVal($points)
			. ', u_count = u_count + '.sqlVal($quantity)
            . ", ".updateDmPointsSetSql()
			. ' WHERE userid = '.sqlVal($user_id)
				// . ' ORDER BY userid'
		);
	}

	/**
	 * Counts Resources needed to build or research from level $from, to level $to
	 * @param int $from Level, that is already build
	 * @param int $to Level, that will be counted to
	 * @param array $data Row from na_construction of this building/reserch
	 */
	protected static function getNeededResources( $from, $to, $data )
	{
		$req_res = array(
			'metal'		=> 0,
			'silicon'	=> 0,
			'hydrogen'	=> 0,
		);
		for($from++; $from <= $to; $from++)
		{
			if($data['basic_metal'] > 0)
			{
				$req_res['metal'] += parseChargeFormula($data['charge_metal'], $data['basic_metal'], $from);
			}
			if($data['basic_silicon'] > 0)
			{
				$req_res['silicon'] += parseChargeFormula($data['charge_silicon'], $data['basic_silicon'], $from);
			}
			if($data['basic_hydrogen'] > 0)
			{
				$req_res['hydrogen'] = parseChargeFormula($data['charge_hydrogen'], $data['basic_hydrogen'], $from);
			}
		}
		return $req_res;
	}

	private static function getSqlBonusBuildTypeMode($t)
	{
		return "case $t.mode
					when ".UNIT_TYPE_MOON_CONSTRUCTION." then ".ACHIEV_BONUS_BUILD_TYPE_MOON."
					when ".UNIT_TYPE_CONSTRUCTION." then ".ACHIEV_BONUS_BUILD_TYPE_PLANET."
					when ".UNIT_TYPE_DEFENSE." then ".ACHIEV_BONUS_BUILD_TYPE_PLANET."
					else ".ACHIEV_BONUS_BUILD_TYPE_ANY." end";
	}

	private static function getSqlBonusBuildTypeField()
	{
		$mode1 = self::getSqlBonusBuildTypeMode("c1");
		$mode2 = self::getSqlBonusBuildTypeMode("c2");
		$mode3 = self::getSqlBonusBuildTypeMode("c3");
		return "CASE ( $mode1 | $mode2 | $mode3 )"
				. " WHEN ".ACHIEV_BONUS_BUILD_TYPE_ANY." THEN ".ACHIEV_BONUS_BUILD_TYPE_ANY
				. " WHEN ".ACHIEV_BONUS_BUILD_TYPE_MOON." THEN ".ACHIEV_BONUS_BUILD_TYPE_MOON
				. " WHEN ".ACHIEV_BONUS_BUILD_TYPE_PLANET." THEN ".ACHIEV_BONUS_BUILD_TYPE_PLANET
				. " ELSE ".ACHIEV_BONUS_BUILD_TYPE_ERROR." END as bonus_build_type";
	}

	/**
	 * Gets All User unseen achievements and assignes them into tpl
	 * @param int $user_id User ID
	 */
	public static function loadAchievementsTemplateData($params = null)
	{
        if(!ACHIEVEMENTS_ENABLED){
            return;
        }
		if( Core::getRequest()->getGET("ajax") || !NS::getUser() )
		{
			return;
		}
		Core::getLanguage()->load(array('Achievements', 'info'));

		$is_achiev_page	= false;
		$cur_page		= Core::getRequest()->getGET("go");
		if( stristr(Core::getRequest()->getGET("go"), "Achievement") !== false )
		{
			$is_achiev_page = true;
			Core::getTPL()->assign('is_achiev_page', true);
		}

		if( is_null($params) && $is_achiev_page )
		{
			return;
		}

		$achievement_id = isset( $params["achievement_id"])
			? max(0, $params["achievement_id"])
			: null;
		if( $achievement_id )
		{
			$user_id	= 0;
			$planet_id	= 0;
			$is_planet_moon = false;
		}
		else
		{
			$user_id = isset($params["user_id"])
				? max(0, $params["user_id"])
				: NS::getUser()->get('userid');
			if( $user_id )
			{
				$planet_id = isset($params["planet_id"])
					? max(0, $params["planet_id"])
					: NS::getUser()->get("curplanet");
				if( $user_id != NS::getUser()->get('userid') || !NS::getPlanet() || $planet_id != NS::getPlanet()->getPlanetId() ) // NS::getUser()->get("curplanet") )
				{
					$row = sqlSelectRow("planet", "planetid, ismoon", "", "userid=".sqlVal($user_id)." AND planetid=".sqlVal($planet_id));
					if( empty($row) )
					{
						$planet_id = 0;
						$is_planet_moon = false;
					}
					else
					{
						$is_planet_moon = $row["ismoon"];
					}
				}
				else
				{
					$is_planet_moon = NS::getPlanet()->getData('ismoon');
				}
			}
		}
		$is_other_user = ($user_id != NS::getUser()->get('userid') && $achievement_id == null);

		$user_achievements	= array();
		$cur_new_achieve_count = 0;

		$new_achievements = array();
		$real_new_achieve_count = 0;
		Core::getTPL()->assign('done_achi', true);
		Core::getTPL()->assign('aval_achi', true);
		foreach( array(true, false) as $branching_key => $branching_val )
		{
			if( $branching_val && isset($params['skip_done']) && $params['skip_done'] == true )
			{
				Core::getTPL()->assign('done_achi', false);
				continue;
			}
			elseif( !$branching_val && isset($params['skip_aval']) && $params['skip_aval'] == true )
			{
				Core::getTPL()->assign('aval_achi', false);
				continue;
			}

			$use_limit = false;
			$max_val	= '';
			$start_val	= '';
			$params['paginator']['user_got'] = $branching_val;
			if(
				isset($params['paginator'])
					&& !empty($params['paginator'])
					&& $params['paginator']['enabled']
					&& !empty($user_id)
					&& $user_id != 0
			)
			{
				list($use_limit, $start_val, $max_val) = self::createPaginator( $user_id, $params['paginator']);
				if( $use_limit )
				{
					Core::getTPL()->assign(($branching_val ?'done_':'aval_') . 'achi_paginator', true);
					Core::getTPL()->assign(
						($branching_val ?'done_':'aval_') . 'formaction',
						socialUrl(RELATIVE_URL . 'game.php/Achievements' . ($branching_val ? 'Done' : 'Avaliable'))
					);
				}
				else
				{
					Core::getTPL()->assign(($branching_val ?'done_':'aval_') . 'achi_paginator', false);
				}
			}

			$result = sqlSelect(
				'achievement_datasheet ach',
				'ach.*, c.name'
					. ( $user_id
						? ', a2u.*, u.*, case when granted > ('.sqlVal(time()).' - ach.time) then 1 else 0 end as is_granted, '
							. self::getSqlBonusBuildTypeField()
						: ', 0 as granted, 0 as is_granted'),
				' JOIN '.PREFIX.'construction c ON c.buildingid = ach.achievement_id '
				. ($user_id ?
					' JOIN '.PREFIX.'achievements2user a2u ON ach.achievement_id = a2u.achievement_id'
					. ' JOIN '.PREFIX.'user u ON u.userid = a2u.user_id'
					. ' LEFT JOIN '.PREFIX.'construction c1 ON c1.buildingid = ach.bonus_1_unit_id'
					. ' LEFT JOIN '.PREFIX.'construction c2 ON c1.buildingid = ach.bonus_2_unit_id'
					. ' LEFT JOIN '.PREFIX.'construction c3 ON c1.buildingid = ach.bonus_3_unit_id'
					: ''),
				'1 = 1'
					. (($achievement_id == null)?($branching_val ? ' AND a2u.granted > 0' : ' AND a2u.granted = 0' ):'')
					. ($is_other_user? ' AND a2u.granted > 0':'')
					. ($achievement_id ? ' AND ach.achievement_id='.sqlVal($achievement_id) : '')
					. ($user_id ? ' AND a2u.user_id = '.sqlVal( $user_id ) : '')
					. ($user_id && !$is_achiev_page ? ' AND a2u.state in ('.sqlArray(ACHIEV_STATE_ALERT, ACHIEV_STATE_BONUS_GIVEN).')' : '')
				,
				($user_id ? 'a2u.granted DESC, a2u.created DESC, ' : '') .'c.display_order DESC, c.buildingid DESC',
				( $use_limit
					? ' '.$start_val . ', ' . $max_val
					: '')
			);
			while ( $row = sqlFetch($result) )
			{
				$row['desc']	= Core::getLanguage()->getItem($row['name'] . '_DESC');
				$row['image']	= Image::getImage( getUnitImage( ($row['image'] ? 'achiev_'.$row['image'] : $row['name']) ), Core::getLanguage()->getItem($row['name']) );
				$row['name']	= Link::get("game.php/AchievementInfo/".$row["achievement_id"], Core::getLanguage()->getItem($row['name']));
				$row['bonus_items']	= self::getBonusResItems( $row );
				$row['bonuses']	= self::getBonusUnitList( $row );
				if( $row['granted'] == 0 || $achievement_id > 0 || $is_other_user)
				{
					$row['req_items']	= self::getRequirementResItems(
						$row,
						$row['granted'] == 0 || $achievement_id > 0,
						$row['granted'] > 0 || empty($user_id),
						$user_id,
						$planet_id
					); // || $is_other_user );
					$row['reqs'] = NS::requiremtentsList(
						$row['achievement_id'],
						$user_id,
						$planet_id,
						$row['granted'] == 0 || $achievement_id > 0,
						$row['granted'] > 0 || empty($user_id)
					); // || $is_other_user );
					// $row['reqs'][]	= self::getRequirementResItems( $row ); // , $row['granted'] == 0 || $achievement_id > 0, $row['granted'] > 0 ); // || $is_other_user );
					// $row['reqs']	= implode('<br/>', $row['reqs']);
					$row['reqs_exist'] = $row['req_items'] || $row['reqs'];
				}
				if($row['granted'] > 0)
				{
					if($row['state'] == ACHIEV_STATE_ALERT) // if($row['state'] != ACHIEV_STATE_BONUS_GIVEN && $row['state'] != ACHIEV_STATE_PROCESSED)
					{
						$cur_new_achieve_count++;
					}
					if( isset($row['bonus_build_type']) )
					{
						$row['bonus_blocked'] = false;
						if( $row['bonus_build_type'] != ACHIEV_BONUS_BUILD_TYPE_ANY
							&& ( $row['bonus_build_type'] == ACHIEV_BONUS_BUILD_TYPE_ERROR
									|| ($row['build_bonus_type'] == ACHIEV_BONUS_BUILD_TYPE_MOON) != $is_planet_moon ) )
						{
							$row['bonus_blocked'] = true;
						}
					}
					$user_achievements[ $row['achievement_id'] ] = $row;
				}
				else
				{
					if($row['state'] == ACHIEV_STATE_ALERT) // $row['created'] > time()-60*60*8)
					{
						$real_new_achieve_count++;
					}
					$new_achievements[ $row['achievement_id'] ] = $row;
				}
			}
			sqlEnd($result);
		}
		Core::getTPL()->assign(
			'achiev_ajax_hide_url',
			addUrlParams(socialUrl(FULL_URL.'game.php/AchievementHideAjax/'), 'id=')
		);
		Core::getTPL()->assign(
			'achiev_get_bonus_url',
			socialUrl(FULL_URL.'game.php/AchievementGetBonus/')
		);
		Core::getTPL()->assign(
			'achiev_process_url',
			socialUrl(FULL_URL.'game.php/AchievementProcess/')
		);
		Core::getTPL()->assign(
			'achiev_recalc_url',
			socialUrl(FULL_URL.'game.php/AchievementsRecalc/')
		);
		Core::getTPL()->addLoop('cur_achieves', $user_achievements);
		Core::getTPL()->addLoop('new_achieves', $new_achievements);
		Core::getTPL()->assign('cur_new_achieve_count', $cur_new_achieve_count);
		Core::getTPL()->assign('real_new_achieve_count', $real_new_achieve_count);
		Core::getTPL()->assign('is_achiev_other_user', $is_other_user);
		Core::getTPL()->assign('achievement_id', $achievement_id);
		Core::getTPL()->assign('show_achives', $user_achievements || $new_achievements ? true : false);
	}

	/**
	 * Creates Paginator for achivements page
	 * @param integer $user_id User ID for whom we create pagintor
	 * @param array $data Paginator related data
	 */
	protected static function createPaginator($user_id, array $data)
	{
		$result_count = sqlSelectField(
			'achievements2user a2u',
			'count(*)',
			'',
			'a2u.user_id = '.sqlVal( $user_id )
				.' AND granted ' . ($data['user_got']?' > 0':' = 0')
		);
		$count_all_user_achivs = $result_count;
		if( $count_all_user_achivs <= $data['per_page'] )
		{
			return array(false, 0, 0);
		}
		$pages = ceil($count_all_user_achivs / $data['per_page']);

		$page = $data['page'];
		if(!is_numeric($page))
		{
			$page = 1;
		}
		elseif($page > $pages)
		{
			$page = $pages;
		}
		else if($page < 1)
		{
			$page = 1;
		}
		$pages_to_show	= $data['pages2show'];
		$pages_range	= floor($pages_to_show / 2);
		$pages_link		= '';
		$pages_sel		= '';
		for($i = 0; $i < $pages; $i++)
		{
			$i1	= $i+1;
			$n	= $i * $data['per_page'] + 1;
			if($i1 == $page)
			{
				$s = 1;
			}
			else
			{
				$s = 0;
			}
			$pages_sel .= createOption($i + 1, $i1, $s);
			if (
				(abs($i1 - $page) <= $pages_range)
					&& ($pages_to_show-- > 0)
			)
			{
				$pages_link .= createPageLink($i1, $s, "[$i1]", ($data['user_got'] ? 'done_' : 'aval_' ) );
			}
		}
		if ( $page != $pages )
		{
			Core::getTPL()->assign(($data['user_got'] ?'done_':'aval_') . 'link_next', createPageLink($page +1, false, Core::getLanguage()->getItem('NEXT_PAGE'), ($data['user_got'] ?'done_':'aval_')));
			Core::getTPL()->assign(($data['user_got'] ?'done_':'aval_') . 'link_last', createPageLink($pages, false, '&gt;&gt;', ($data['user_got'] ?'done_':'aval_')));
		}

		if ($page != 1)
		{
			Core::getTPL()->assign(($data['user_got'] ?'done_':'aval_') . 'link_prev', createPageLink($page -1, false, Core::getLanguage()->getItem('PREV_PAGE'), ($data['user_got'] ?'done_':'aval_')));
			Core::getTPL()->assign(($data['user_got'] ?'done_':'aval_') . 'link_first', createPageLink(1, false, '&lt;&lt;', ($data['user_got'] ?'done_':'aval_')));
		}
		Core::getTPL()->assign(($data['user_got'] ?'done_':'aval_') . 'page',		$page);
		Core::getTPL()->assign(($data['user_got'] ?'done_':'aval_') . 'page_links',$pages_link);
		Core::getTPL()->assign(($data['user_got'] ?'done_':'aval_') . 'pages',		$pages_sel);

		$start	= abs( ($page - 1) * $data['per_page'] );
		$max	= $data['per_page'];
		return array(true, $start, $max);
	}

	/**
	 * Gets Achievement requirements
	 * @param array $row Achievement Data
	 * @param boolean $show_all Whether to show req even if it's not avaluable
	 * @param boolean $safe_mode Whether to show req even if it's not avaluable like it's avaluable.
	 * @param integer $user_id What user do we calculate for, defaults to 0
	 * @param integer $planet_id On What planet do we calculate, defaults to 0
	 */
	public static function getRequirementResItems( $row, $show_all = false, $safe_mode = false, $user_id = 0, $planet_id = 0 )
	{
		static $user_stat_fields = array(
			  'points'		=> 'STAT_POINTS',
			  'u_points'	=> 'STAT_UNIT_POINTS',
			  'r_points'	=> 'STAT_RESEARCH_POINTS',
			  'b_points'	=> 'STAT_BUILD_POINTS',
			  'e_points'	=> 'BATTLE_EXPERIENCE',
			  'u_count'		=> 'STAT_UNIT_COUNT',
			  'r_count'		=> 'STAT_RESEARCH_COUNT',
			  'b_count'		=> 'STAT_BUILD_COUNT',
		);

		$requirements = array();
		foreach( $row as $key => $required_res )
		{
			if( substr($key, 0, 4) == 'req_' && $required_res > 0 )
			{
				$res_name = substr($key, 4);
				if( isset($row[$res_name]) )
				{
					$available_res = (float)$row[$res_name];
				}
				else
				{
					$available_res = 0;
				}
				$is_err = $required_res > 0 && $available_res < $required_res;
				$req = array(
					"res_name"			=> Core::getLanguage()->getItem( isset($user_stat_fields[$res_name]) ? $user_stat_fields[$res_name] : "ACHIEV_REQ_".strtoupper($res_name) ),
					"res_required"		=> fNumber($required_res),
					"res_notavailable"	=> $is_err ? fNumber($available_res - $required_res) : "",
				);

				if( $safe_mode )
				{
					unset( $req["res_notavailable"] );
					$requirements[] = $req;
				}
				else if( $is_err || $show_all )
				{
					$requirements[] = $req;
				}
			}
			elseif( substr($key, 0, 11) == 'custom_req_' && !empty($required_res) )
			{
				$check_funct_name = 'checkReq' . $required_res;
				if( !method_exists('Achievements', $check_funct_name) )
				{
						error_log('Can\'t find custom functon ' . $check_funct_name, 'warning');
						continue;
				}
				$name_funct_name = 'getReqName' . $required_res;
				if( !method_exists('Achievements', $name_funct_name) )
				{
						error_log('Can\'t find custom name functon ' . $name_funct_name, 'warning');
						continue;
				}
				$user_id	= ($user_id == 0 ? null: $user_id);
				$planet_id	= ($planet_id == 0 ? null: $planet_id);
				$achiev_id	= $row['achievement_id'];
				$req = array(
					"res_name"			=> '',
					"res_required"		=> self::$name_funct_name(),
					"res_notavailable"	=> ( $safe_mode ? '' : self::$check_funct_name($user_id, $planet_id, $achiev_id) ),
				);
				if( $safe_mode )
				{
					$requirements[] = $req;
				}
				else if( !$req["res_notavailable"] || $show_all )
				{
					$req["res_notavailable"] = !$req["res_notavailable"];
					$requirements[] = $req;
				}
			}
		}
		return $requirements;
	}

	/**
	 * Converts Achievement Bonuses to string
	 * @param array $row Achievement Data
	 */
	public static function getBonusResItems( $row )
	{
		$bonuses = array();
		foreach ( array('METAL', 'SILICON', 'HYDROGEN', 'CREDIT') as $res )
		{
			$field_name = 'bonus_' . strtolower($res);
			if( !empty($row[$field_name]) )
			{
				$bonuses[] = array(
					"res_name" => Core::getLanguage()->getItem($res),
					"res_bonus" => fNumber($row[$field_name]),
				);
			}
		}
		return $bonuses;
	}

	public static function getBonusUnitList( $row )
	{
		$c_name = 'Achievemets.getBonusUnitList.' . $row['achievement_id'];
		$c_dur	= 3600*24*3;
		$result = false;false;
		if( $result === false )
		{
			$bonuses = array();
			foreach( array('bonus_1', 'bonus_2', 'bonus_3') as $sufix )
			{
				$id_name	= $sufix . '_unit_id';
				$level_name = $sufix . '_unit_level';
				if( !empty( $row[$id_name] ) && !empty( $row[$level_name] ) )
				{
					$item = sqlSelectRow('construction', 'name, mode', '', 'buildingid = '.sqlVal($row[$id_name]), 'display_order ASC, buildingid ASC');
					if( empty($row) )
					{
						continue;
					}
					$name = Core::getLanguage()->getItem($item['name']) . ' ' . fNumber($row[$level_name]);
					$style_class = "true";
					switch($item["mode"])
					{
						case UNIT_TYPE_FLEET:
						case UNIT_TYPE_DEFENSE:
							$page = "UnitInfo";
							break;
						case UNIT_TYPE_ACHIEVEMENT:
							$page = "AchievementInfo";
							break;
						case UNIT_TYPE_ARTEFACT:
							$page = "ArtefactInfo";
							break;
                        case UNIT_TYPE_CONSTRUCTION:
                        case UNIT_TYPE_MOON_CONSTRUCTION:
							$page = "ConstructionInfo";
							break;
                        case UNIT_TYPE_RESEARCH:
							$page = "ResearchInfo";
							break;
						default:
							$page = "UnknownInfo";
							break;
					}
					$name = "<span class='$style_class'>" . Link::get("game.php/{$page}/".$row[$id_name], $name, "", $style_class);

					$bonuses[] = $name;
				}
			}
			$result = implode('<br />', $bonuses);
//			// cache disabled
		}
		return $result;
	}
}
?>