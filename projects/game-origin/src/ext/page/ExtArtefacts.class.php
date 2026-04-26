<?php
/**
* Displays artefacts owned by user.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class ExtArtefacts extends Page
{
	/**
	* Constructor. Handles requests for this page.
	*
	* @return void
	*/
	public function __construct()
	{
		parent::__construct();
		NS::processCreditBonusItems();
		if(!NS::getUser()->get("umode"))
		{
			$this
				->setGetAction("go", "ArtefactOn", "activateArtefact")
				->addGetArg("activateArtefact", "id")
				->setGetAction("go", "ArtefactOff", "deactivateArtefact")
				->addGetArg("deactivateArtefact", "id")
				->setGetAction("go", "ArtefactInfo", "showArtefact")
				->addGetArg("showArtefact", "id")
				;
			/*
			$this->setPostAction("use", "useArtefact")
				->addPostArg("useArtefact", "artid")
				->addPostArg("useArtefact", "typeid");
			$this->setPostAction("activate", "activateArtefact")
				->addPostArg("activateArtefact", "typeid");
			$this->setPostAction("deactivate", "deactivateArtefact")
				->addPostArg("deactivateArtefact", "typeid");
			*/
		}

		Core::getLanguage()->load("info,ArtefactInfo");
//		try {
			$this->proceedRequest();
//		} catch(Exception $e) {
//			$e->printError();
//		}
		return;
	}

	protected function showArtefact($id)
	{
		return $this->index($id);
	}

	/**
	* Index action.
	*
	* @return Artefacts
	*/
	protected function index($info_typeid = null)
	{
		$sel_contstraction_id = 0;
		$sel_level = 0;
		$non_personal = 0;
		if ( !empty($info_typeid) && !is_numeric($info_typeid) )
		{
			$temp = explode('_', $info_typeid);
			$info_typeid = (int)$temp[0];
			$sel_contstraction_id = (int)$temp[2];
			$sel_level = (int)$temp[1];
			if (isset($temp[3]) && !empty($temp[3]))
			{
				$non_personal = 1;
			}
		}

		$info_typeid = max(0, (int)$info_typeid);

		Core::getTPL()->addHTMLHeaderFile("lib/jquery.countdown.js?".CLIENT_VERSION, "js");
		Core::getTPL()->addHTMLHeaderFile("lib/jquery.countdown-ru.js?".CLIENT_VERSION, "js");

		$umode = NS::getUser()->get("umode");

		Core::getTPL()->assign("artefact_typeid", $info_typeid);

		$artefact_tech_storage = Artefact::getStorageSlots();
		$artefact_tech_usage = Artefact::getUsedSlots();
		$artefact_tech_free = $artefact_tech_storage - $artefact_tech_usage;
		$has_free_slots = $artefact_tech_free > 0;

		$row = sqlSelectRow("construction", "*", "", "buildingid=".UNIT_ARTEFACTS_TECH);
		Core::getTPL()->assign("artefact_tech_level", Artefact::getTechLevel());
		Core::getTPL()->assign("artefact_tech_name", Core::getLanguage()->getItem($row["name"]));
		Core::getTPL()->assign("artefact_tech_image", Link::get("game.php/ResearchInfo/".UNIT_ARTEFACTS_TECH, Image::getImage(getUnitImage($row["name"]), Core::getLanguage()->getItem($row["name"]))));
		Core::getTPL()->assign("artefact_tech_description", Core::getLanguage()->getItem($row["name"]."_DESC"));
		Core::getTPL()->assign("artefact_tech_storage", fNumber($artefact_tech_storage));
		Core::getTPL()->assign("artefact_tech_used", fNumber($artefact_tech_usage));
		Core::getTPL()->assign("artefact_tech_free", fNumber($artefact_tech_free));
		Core::getTPL()->assign("artefact_tech_free_percent", $artefact_tech_storage > 0 ? min(100, round($artefact_tech_free * 100 / $artefact_tech_storage)) : 100);

		$select = array('b.buildingid', 'b.name', 'a2u.artid AS art_id', 'a2u.userid AS owner',
			'a2u.active', 'a2u.times_left', 'a2u.delay_eventid', 'a2u.expire_eventid', 'a2u.lifetime_eventid',
			'a2u.construction_id AS add_const_id', 'a2u.level AS add_level', 'a2u.lot_id AS isInLot',
			'ads.movable', 'ads.unique', 'ads.usable', 'ads.use_duration', 'ads.delay', 'ads.use_times',
			'ads.lifetime', 'ads.max_active', 'ads.effect_type', 'ads.trophy_chance', 'ads.quota',
			'IFNULL(g.galaxy, gm.galaxy) as galaxy',
			'IFNULL(g.system, gm.system) as system',
			'IFNULL(g.position, gm.position) as position',
			'a2u.planetid', 'a2u.artid', 'p.planetname',
		);

		$joins	= 'INNER JOIN '.PREFIX.'construction b ON (b.buildingid = a2u.typeid)';
		$joins .= 'INNER JOIN '.PREFIX.'artefact_datasheet ads ON (a2u.typeid = ads.typeid)';
		$joins .= 'LEFT JOIN '.PREFIX.'planet p ON (p.planetid = a2u.planetid)';
		$joins .= 'LEFT JOIN '.PREFIX.'galaxy g ON (a2u.planetid = g.planetid)';
		$joins .= 'LEFT JOIN '.PREFIX.'galaxy gm ON (a2u.planetid = gm.moonid)';
		$result = sqlSelect(
			'artefact2user a2u',
			$select,
			$joins,
			'a2u.deleted=0'
				. (!$non_personal ? ' AND a2u.userid='.sqlUser() : '')
				. ($info_typeid ? ' AND a2u.typeid='.sqlVal($info_typeid) : '')
				. ($sel_contstraction_id ? ' AND a2u.construction_id='.sqlVal($sel_contstraction_id) : '')
				. ($sel_level ? ' AND a2u.level='.sqlVal($sel_level) : ''),
			'b.display_order ASC, a2u.artid ASC',
			($non_personal ? ' 1' : '')
		);
		while($row = sqlFetch($result))
		{
			$id = $bid = $row["buildingid"];
			$art_id = $artid = $row["art_id"];

			/* $art_pack_row = false;
			if ( $id == ARTEFACT_PACKED_BUILDING || $id == ARTEFACT_PACKED_RESEARCH )
			{
				$art_pack_row = sqlSelectRow('artefact2user', '*', '', "artid = ".sqlVal($art_id));;
				$id	 = $id . '_' . $art_pack_row['level'] . '_' . $art_pack_row['construction_id'];
			} */
			$art_data = sqlSelectRow(
				'artefact2user AS a',
				'*', '
				LEFT JOIN '.PREFIX.'construction AS c ON c.buildingid = a.construction_id
				LEFT JOIN '.PREFIX.'phrases AS p ON p.title = c.name ',
				"a.artid = ".sqlVal($art_id)." AND p.languageid=".sqlVal(NS::getLanguageId())
			);

			$cur_level = 0;
			$is_packed_art = false;
			if( $bid == ARTEFACT_PACKED_BUILDING || $bid == ARTEFACT_PACKED_RESEARCH )
			{
				$is_packed_art = true;
				$id	 = $id . '_' . $art_data['level'] . '_' . $art_data['construction_id'];
				if( $bid == ARTEFACT_PACKED_BUILDING )
				{
					$cur_level = NS::getPlanet()->getBuilding($art_data['construction_id']);
				}
				else
				{
					$cur_level = NS::getResearch( $art_data['construction_id'] );
				}
			}

			$check_req = null;
			if(!isset($artefacts[$id]))
			{
				$artefacts[$id]["required_constructions"] = NS::requiremtentsList($bid);
				$artefacts[$id]["check_req"] = $check_req = Artefact::checkRequirements($id, null, null, $art_id);
				$artefacts[$id]["new_level"] = $artefacts[$id]["check_req"];
				$artefacts[$id]["active_count"] = 0;
				$artefacts[$id]["inactive_count"] = 0;
				if ( !is_numeric($id) )// $id == ARTEFACT_PACKED_BUILDING || $id == ARTEFACT_PACKED_RESEARCH )
				{
					$row['link'] = $id;
					Artefact::setViewParams($artefacts[$id], $row, null,
						// YII_GAME_DIR.'/index.php?r=artefact2user_YII/image&id='.$art_id.'',
						artImageUrl("image", "id=".$art_id, false),
						$art_id);
				}
				else
				{
					Artefact::setViewParams($artefacts[$id], $row);
				}
			}
			$artefacts[$id]["quantity"]++;

			$artefacts[$id]["items"][$art_id] = array(
				"artid"				=> $art_id,
				"active"		 	=> $row["active"],
				"times_left" 		=> $row["times_left"],
				"planetname" 		=> $row["planetid"] ? $row["planetname"]." ".getCoordLink($row["galaxy"], $row["system"], $row["position"], "select", $row["planetid"]) : Core::getLanguage()->getItem("IN_FLIGHT"),
				"check_req"			=> is_null($check_req) ? Artefact::checkRequirements($id, null, null, $art_id) : $check_req,
				"c_level"			=> empty($art_data['level'])? 0 : $art_data['level'],
				"construction_id"	=> empty($art_data['construction_id'])? 0 : $art_data['construction_id'],
				"c_name"			=> $art_data['content'],
				"cur_level"			=> $cur_level,
				'packed'			=> $is_packed_art,
				'isInLot'			=> $row["isInLot"],
			);
			$artefacts[$id]["items"][$art_id]["new_level"] = $artefacts[$id]["items"][$art_id]["check_req"];

			$artefacts[$id]['more_levels']	= !empty($art_data['level']);

			$artefacts[$id]["items"][$art_id]['no_activation'] = false;
			if( $id == ARTEFACT_PACKING_BUILDING || $id == ARTEFACT_PACKING_RESEARCH )
			{
				$artefacts[$id]["items"][$art_id]['no_activation'] = true;
			}

			foreach(array("delay_eventid" => "cur_delay", "expire_eventid" => "cur_expire", "lifetime_eventid" => "cur_lifetime") as $eventid => $time)
			{
				if($row[$eventid])
				{
					$row[$time] = sqlSelectField("events", "time", "", "eventid=".sqlVal($row[$eventid]));
				}
			}

			if($row["cur_delay"] > time())
			{
				$timeleft = max(1, $row["cur_delay"] - time());
				$artefacts[$id]["items"][$artid]["delay_counter"] = "<script type='text/javascript'>
					$(function () {
						$('#delay_counter{$artid}').countdown({until: {$timeleft}, compact: true, onExpiry: function() {
							$('#delay_text{$id}').html('-');
						}});
					});
				</script>
				<span id='delay_text{$artid}'><span id='delay_counter{$artid}'>".getTimeTerm($timeleft)."</span></span>";
			}
			else if($row["cur_expire"] > 0 && $row["cur_expire"] > time())
			{
				$timeleft = max(1, $row["cur_expire"] - time());
				if($row["cur_lifetime"] > 0)
				{
					$timeleft = min($timeleft, max(1, $row["cur_lifetime"] - time()));
				}
				$artefacts[$id]["items"][$artid]["expire_counter"] = "<script type='text/javascript'>
					$(function () {
						$('#expire_counter{$artid}').countdown({until: {$timeleft}, compact: true, onExpiry: function() {
							$('#expire_text{$id}').html('-');
						}});
					});
				</script>
				<span id='expire_text{$artid}'><span id='expire_counter{$artid}'>".getTimeTerm($timeleft)."</span></span>";
			}
			else if($row["cur_lifetime"] > 0 && $row["cur_lifetime"] > time())
			{
				$timeleft = max(1, $row["cur_lifetime"] - time());
				$artefacts[$id]["items"][$artid]["disappear_counter"] = "<script type='text/javascript'>
					$(function () {
						$('#disappear_counter{$artid}').countdown({until: {$timeleft}, compact: true, onExpiry: function() {
							$('#disappear_text{$id}').html('-');
						}});
					});
				</script>
				<span id='disappear_text{$artid}'><span id='disappear_counter{$artid}'>".getTimeTerm($timeleft)."</span></span>";
			}

			$check_req = $artefacts[$id]["items"][$artid]["check_req"] ;
			$usable = $artefacts[$id]["usable"];
			if($row["active"])
			{
				$artefacts[$id]["active_count"]++;
			}
			else
			{
				$artefacts[$id]["inactive_count"]++;
			}
			if(!$umode)
			{
				if($row["cur_delay"] > time()
					|| !$row["planetid"]
					|| !in_array($row["effect_type"], array(ARTEFACT_EFFECT_TYPE_PLANET, ARTEFACT_EFFECT_TYPE_EMPIRE))
					|| $row["isInLot"]
					// || !$has_free_slots
					)
				{
					;
				}
				elseif($row["active"])
				{
					$artefacts[$id]["items"][$artid]["action"] = Link::get("game.php/ArtefactOff/".$artid, Core::getLanguage()->getItem("ARTEFACT_OFF"), "", "false");
				}
				elseif( /*$row["effect_type"] == ARTEFACT_EFFECT_TYPE_PLANET && */ $row["planetid"] != NS::getUser()->get("curplanet"))
				{
					$artefacts[$id]["items"][$artid]["action"] = Core::getLanguage()->getItem("WRONG_ARTEFACT_PLANET");
					$artefacts[$id]["items"][$art_id]["packed"] = false;
				}
				else if(!$row["active"] && !$check_req)
				{
					$artefacts[$id]["items"][$artid]["action"] = Core::getLanguage()->getItem("OUTSTANDING_ARTEFACT_REQUIREMENTS");
				}
				else
				{
					$valid = true;
					if($id == ARTEFACT_MOON_CREATOR)
					{
						if(
							NS::getPlanet()->getData("ismoon")
							|| NS::getPlanet()->getData("destroy_eventid")
							|| sqlSelectField("galaxy", "moonid", "", "planetid=".sqlPlanet())
						)
						{
							$valid = false;
						}
					}

					if($valid && $check_req)
					{
						$artefacts[$id]["items"][$artid]["action"] = Link::get("game.php/ArtefactOn/".$artid, Core::getLanguage()->getItem("ARTEFACT_ON"), "", "true");
					}
				}
			}
		}
		sqlEnd($result);
		Core::getTPL()->addLoop("artefacts", $artefacts);
		Core::getTPL()->assign("nonpersonal", $non_personal);

		$uniques=array();
		$select = array("b.buildingid", "b.name", "a2u.planetid", "pl.planetname", "u.username");
		$joins	= "INNER JOIN ".PREFIX."construction b ON (b.buildingid = a2u.typeid)";
		$joins .= "INNER JOIN ".PREFIX."artefact_datasheet ads ON (a2u.typeid = ads.typeid)";
		$joins .= "LEFT JOIN ".PREFIX."planet pl ON (pl.planetid = a2u.planetid)";
		$joins .= "INNER JOIN ".PREFIX."user u ON (a2u.userid = u.userid)";
		$result = sqlSelect("artefact2user a2u", $select, $joins, "ads.unique = 1".($info_typeid ? " AND a2u.typeid=".sqlVal($info_typeid) : ""), "ads.typeid");
		while($row = sqlFetch($result))
		{
			$id = $row["buildingid"];
			$uniques[$id]["quantity"]++;
			$uniques[$id]["planetname"] = $row["planetid"] ? Core::getLanguage()->getItem("ON_PLANET")." ".$row["planetname"] : Core::getLanguage()->getItem("IN_FLIGHT");
			$uniques[$id]["owner"] = $row["username"];
			Artefact::setViewParams($uniques[$id], $row);
		}
		sqlEnd($result);
		Core::getTPL()->addLoop("uniques", $uniques);

		Core::getTPL()->display("artefacts");
		return $this;
	}

	protected function activateArtefact($artid, $active = true)
	{
		$artid = max(0, (int)$artid);
		$active = intval((int)$active != 0);

		if(!NS::isFirstRun("Artefacts::activate:{$artid}-{$active}"))
		{
			doHeaderRedirection("game.php/Artefacts", false);
			return;
		}

		if(NS::getUser()->get("umode"))
		{
			throw new GenericException("Your account is still in vacation mode.");
		}

		$select = "b.buildingid, a2u.times_left, ad.use_duration";
		$joins	= "INNER JOIN ".PREFIX."construction b ON (b.buildingid = a2u.typeid)".
		"LEFT JOIN ".PREFIX."artefact_datasheet ad ON (ad.typeid = a2u.typeid)";
		$row = sqlSelectRow("artefact2user a2u", $select, $joins,
			"a2u.deleted=0 AND a2u.active=".sqlVal(1-$active)." AND a2u.artid=".sqlVal($artid)." AND a2u.userid=".sqlUser()); // ." AND a2u.planetid=".sqlPlanet());
		if(!$row)
		{
			throw new GenericException("Error artefact: $artid");
		}

		$id = $row["buildingid"];
		if($active && !Artefact::checkRequirements($id, null, null, $artid))
		{
			throw new GenericException("Requirements are failed for typeid: $id");
		}
		if($active)
		{
			Artefact::activate($artid, NS::getUser()->get("userid"), NS::getUser()->get("curplanet"));
		}
		else
		{
			Artefact::deactivate($artid, NS::getUser()->get("userid"), NS::getUser()->get("curplanet"));
		}

		doHeaderRedirection("game.php/Artefacts", false);
	}

	protected function deactivateArtefact($artid)
	{
		$this->activateArtefact($artid, false);
	}
}
?>
