<?php
/**
* Construction & builings page.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class ArtefactMarket extends Construction
{
  /**
  * Displays list of available buildings.
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
        ->setGetAction("go", "BuyArtefact", "buyArtefact")
        ->addGetArg("buyArtefact", "id")
        ->addGetArg("buyArtefact", "con_id")
        ->addGetArg("buyArtefact", "level")
        ;
    }
//    try {
      $this->proceedRequest();
//    } catch(Exception $e) {
//      $e->printError();
//    }
  }

  /**
  * Index action.
  *
  * @return Constructions
  */
  protected function index()
  {
    Core::getLanguage()->load("info,ArtefactInfo");

    $credit = NS::getUser()->get("credit");

	$building_levels = array();
    foreach($GLOBALS['PACKED_ARTEFACT_BUILDING_LEVELS'] as $level)
    {
    	$building_levels[]['sel_id'] = $level;
    }

	$research_levels = array();
    foreach($GLOBALS['PACKED_ARTEFACT_RESEARCH_LEVELS'] as $level)
    {
    	$research_levels[]['sel_id'] = $level;
    }

    $counter = 1;
    $items = array();

    Core::getTPL()->addHTMLHeaderFile("lib/jquery.countdown.js?".CLIENT_VERSION, "js");
    Core::getTPL()->addHTMLHeaderFile("lib/jquery.countdown-ru.js?".CLIENT_VERSION, "js");

    $result = sqlSelect(
      "construction",
      'buildingid, mode, name',
      "",
      "mode in (".sqlArray(UNIT_TYPE_CONSTRUCTION, UNIT_TYPE_MOON_CONSTRUCTION, UNIT_TYPE_RESEARCH).")"
		. " AND buildingid not in (".sqlArray( $GLOBALS["BLOCKED_MARKET_UNITS"] ).")"
		. (!isAdmin() ? " AND test = 0" : "")
    );
    while($row = sqlFetch($result))
    {
    	if ( $row['mode'] == UNIT_TYPE_RESEARCH )
    	{
    		$research_options[] = array(
    			'sel_name' => Core::getLanguage()->getItem($row['name']),
    			'sel_id' => $row['buildingid']
    		);
    	}
    	else
    	{
    		$building_options[] = array(
    			'sel_name' => Core::getLanguage()->getItem($row['name']),
    			'sel_id' => $row['buildingid']
    		);
    	}
    }
    sqlEnd($result);

    $result = sqlSelect("construction b", array(
        "b.buildingid", "b.name", "b.basic_metal", "b.basic_silicon", "b.basic_hydrogen", "b.basic_energy", "b.basic_credit",  "b.basic_points",
        "b.charge_metal", "b.charge_silicon", "b.charge_hydrogen", "b.charge_energy", "b.charge_credit", "b.charge_points",
        "ads.movable", "ads.buyable", "ads.unique", "ads.usable", "ads.use_times", "ads.use_duration",
        "ads.delay", "ads.lifetime", "ads.effect_type", "ads.max_active", "ads.trophy_chance", "ads.quota"
        ),
        "INNER JOIN ".PREFIX."artefact_datasheet ads ON b.buildingid = ads.typeid",
        "mode = ".sqlVal($this->unit_type) /* ." AND buyable=1" */
        . (isset($GLOBALS['DISABLED_ARTEFACTS']) ? " AND b.buildingid NOT IN (".sqlArray($GLOBALS['DISABLED_ARTEFACTS']).")" : "")
        , "display_order ASC, buildingid ASC");
    while($row = sqlFetch($result))
    {
      $id = $row["buildingid"];
	  $is_buyable = $row["buyable"] || isAdmin(null, true);
      $can_build = $is_buyable && NS::checkRequirements($id);

      $can_show_all = $this->canShowAllUnits();
      Core::getTPL()->assign("show_all_units", $can_show_all);

      if(1) // $can_build || $can_show_all)
      {
		Artefact::setViewParams($items[$id], $row, null, null, false, false);

		if ( $id == ARTEFACT_PACKED_BUILDING )
		{
        	$items[$id]['selects'] = $building_options;
			$items[$id]['levels'] = $building_levels;
        }
        if ( $id == ARTEFACT_PACKED_RESEARCH )
        {
        	$items[$id]['selects'] = $research_options;
			$items[$id]['levels'] = $research_levels;
        }
        $items[$id]['raw_credit'] = $row['basic_credit'];
        $items[$id]["quantity_num"] = sqlSelectField("artefact2user", "count(*)", "",
          "deleted=0 AND typeid=".sqlVal($id)." AND userid=".sqlUser()); // ." AND planetid=".sqlPlanet());
        $items[$id]["quantity"] = fNumber($items[$id]["quantity_num"]);

        $items[$id]["buyable"] = 0;
        $items[$id]["required_constructions"] = NS::requiremtentsList($id, null, null, true);

        // Required resources
        $this->setRequieredResources(0, $row);

        //Construction output
        if(!$is_buyable)
        {
          $items[$id]["buy"] = "<span class='false'>".Core::getLanguage()->getItem("ARTEFACT_NOT_BUYABLE")."</span>";
        }
        else if(NS::getUser()->get("umode")) // || isset($queue_units[$id]))
        {
          // $items[$id]["buyable"] = 0;
        }
		else if($can_build && $this->checkResources())
        {
          $items[$id]["buyable"] = 1;
          $items[$id]["buy"] = Link::get("game.php/BuyArtefact/".$id, Core::getLanguage()->getItem("BUY"), "", "true");
		  if(!$row["buyable"])
		  {
			$items[$id]["buy"] .= "<br /><br /><span class='false'>".Core::getLanguage()->getItem("ARTEFACT_NOT_BUYABLE")."</span>";
		  }
        }
        else
        {
          $items[$id]["buy"] = "<span class='false'>".Core::getLanguage()->getItem("BUY")."</span>";
        }
        // End upgrade output

        // Data output
        foreach(array("metal", "silicon", "hydrogen", "energy", "credit", "points") as $res_name)
        {
          switch($res_name)
          {
          case "energy":
            $available_res = NS::getPlanet()->getEnergy();
            break;
          case "credit":
            $available_res = NS::getUser()->get("credit");
            break;
		 case "points":
			$available_res = NS::getUser()->get("points");
			break;
          default:
            $available_res = NS::getPlanet()->getAvailableRes($res_name);
          }
          $required_field = "required" . ucfirst($res_name);
          $required_res = $this->$required_field;
          $items[$id][$res_name."_required"] = $required_res > 0 ? fNumber($required_res) : "";
          $items[$id][$res_name."_notavailable"] = $required_res > 0 && $available_res < $required_res ? fNumber($available_res - $required_res) : "";
        }
        // $time = NS::getBuildingTime($this->requiredMetal, $this->requiredSilicon, $this->unit_type);
        // $items[$id]["productiontime"] = getTimeTerm($time);
      }
    }
    sqlEnd($result);

    //print_r($building_options);

    // Hook::event("CONSTRUCTIONS_LOADED", array(&$items));
    // debug_var($items, "artefacts"); exit;

    Core::getTPL()->addLoop("imagePacks", getImgPacks());
    Core::getTPL()->addLoop("artefacts", $items);/*

    Core::getTPL()->addLoop("research_options", $research_options);

    Core::getTPL()->addLoop("building_options", $building_options);*/

    Core::getTPL()->display("artefactmarket2");
    return $this;
  }

  protected function buyArtefact($id, $con_id = 0, $level = 0)
  {
    $id = max(0, (int)$id);
  	$con_id = max(0, (int)$con_id);
  	$level = max(0, (int)$level);

  	if(!NS::isFirstRun("ArtefactMarket::buy:{$id}-" . $_SESSION["userid"] ?? 0))
  	{
		doHeaderRedirection($this->main_page, false);
  		return;
  	}
  	if ( !empty($level)
			&& ( ($id != ARTEFACT_PACKED_BUILDING && $id != ARTEFACT_PACKED_RESEARCH)
				|| !in_array($level, $GLOBALS[$id == ARTEFACT_PACKED_BUILDING ? 'PACKED_ARTEFACT_BUILDING_LEVELS' : 'PACKED_ARTEFACT_RESEARCH_LEVELS']))
		)
  	{
  	    doHeaderRedirection($this->main_page, false);
  		return;
  	}
    if($id == ARTEFACT_PACKED_BUILDING || $id == ARTEFACT_PACKED_RESEARCH){
        if(in_array($con_id, $GLOBALS["BLOCKED_MARKET_UNITS"])
                || in_array($con_id, $GLOBALS["CANT_PACK_UNITS"])
                || (isset($GLOBALS["MAX_UNIT_LEVELS"][$con_id]) && $level > $GLOBALS["MAX_UNIT_LEVELS"][$con_id])
                || ($level > ($id == ARTEFACT_PACKED_BUILDING ? MAX_BUILDING_LEVEL : MAX_RESEARCH_LEVEL))
            )
        {
            doHeaderRedirection($this->main_page, false);
            return;
        }
    }
    $select = array("b.basic_metal", "b.basic_silicon", "b.basic_hydrogen", "b.basic_energy", "b.basic_credit", "b.basic_points", "ads.buyable", "ads.movable", "ads.usable", "ads.use_times", "ads.lifetime");
    $joins  = "INNER JOIN ".PREFIX."construction b ON b.buildingid = ads.typeid";
    $row = sqlSelectRow("artefact_datasheet ads", $select, $joins,
		// (isAdmin(null, true) ? "1=1" : "ads.buyable=1")
        "1=1"
        . " AND ads.typeid=".sqlVal($id)
        . (!isAdmin() ? " AND test = 0" : "")
        . (isset($GLOBALS['DISABLED_ARTEFACTS']) ? " AND b.buildingid NOT IN (".sqlArray($GLOBALS['DISABLED_ARTEFACTS']).")" : "")
        );

    // $row = getConstructionDesc($id, $this->unit_type);
    if(!$row || (empty($row['buyable']) && !isAdmin(null, true)))
    {
        doHeaderRedirection($this->main_page, false);
        throw new GenericException("Unkown artefact $id, mode: $this->unit_type :(");
    }

    $row['basic_credit'] = $row['basic_credit'] * ( empty($level) ? 1 : $level );

    // Check for requirements
    if(!NS::checkRequirements($id))
    {
      throw new GenericException("You does not fulfil the requirements to build this.");
    }

    // Hook::event("UPGRADE_BUILDING_FIRST", array(&$row));

    $this->setRequieredResources(0, $row);

    $is_storage_unit = false;

    // Check resources
    if($this->checkResources(!$is_storage_unit))
    {
      $userid	= NS::getUser()->get("userid");
      $planetid = NS::getUser()->get("curplanet");

      $res_log = NS::updateUserRes(array(
        "block_minus" => true,
        "type" => RES_UPDATE_BUY_ARTEFACT,
        // "event_mode" => $mode,
        // "reload_planet" => false,
        "userid" => $userid,
        "planetid" => $planetid,
        "metal" => - $this->requiredMetal,
        "silicon" => - $this->requiredSilicon,
        "hydrogen" => - $this->requiredHydrogen,
        "credit" => - $this->requiredCredit,
        ));
      if(!empty($res_log["minus_blocked"]))
      {
        // throw new GenericException("Not enough resources to build this.");
		doHeaderRedirection($this->main_page, false);
        return;
      }
      $art_id = Artefact::appear(
      	$id,
      	$userid,
      	$planetid,
      	array("delay" => 0, 'level' => $level, 'construction_id' => $con_id, 'bought' => true)
      );
      if( $this->requiredCredit > 0 )
      {
      	new AutoMsg(
      		MSG_CREDIT,
      		$userid,
      		time(),
      		array(
      			'credits'	=> $this->requiredCredit,
      			'msg' 		=> 'MSG_CREDIT_ARTEFACT_BUY',
      			'content'	=> array('exchange' => ETYPE_ARTEFACT, 'artef_id' => $art_id) )
      	);
      }
      doHeaderRedirection($this->main_page, false);
    }
    else
    {
      return $this->index();
//      throw new GenericException("Not enough resources to build this.");
    }
    return $this;
  }

}

?>