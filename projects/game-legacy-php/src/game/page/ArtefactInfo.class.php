<?php
/**
* Shows history of unique artefacts.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

require_once(APP_ROOT_DIR."game/Artefact.class.php");

class ArtefactInfo extends Page
{
	/**
	* Constructor: Shows informations about an unit.
	*
	* @return void
	*/
	public function __construct()
	{
		parent::__construct();
//		try {
			$this->setPostAction("use", "useArtefact")
				->addPostArg("useArtefact", "artid")
				->addPostArg("useArtefact", "typeid");
			$this->setPostAction("activate", "activateArtefact")
				->addPostArg("activateArtefact", "artid")
				->addPostArg("activateArtefact", "typeid");
			$this->setPostAction("deactivate", "deactivateArtefact")
				->addPostArg("deactivateArtefact", "artid")
				->addPostArg("deactivateArtefact", "typeid");
			$this->setGetAction("go", "ArtefactInfo", "showInfo")
				->addGetArg("showInfo", "id")
				->proceedRequest();
//		} catch(Exception $e) {
//			$e->printError();
//		}
		return;
	}

	/**
	* Index action.
	*
	* @return ArtefactInfo
	*/
	protected function index()
	{
		$this->showInfo(Core::getRequest()->getGET("id"));
		return $this;
	}

	/**
	* Shows the history
	*
	* @param integer	Artefact type id
	*
	* @return ArtefactInfo
	*/
	protected function showInfo($typeid)
	{
		// Common unit data
		Core::getTPL()->addHTMLHeaderFile("lib/jquery.countdown.js?".CLIENT_VERSION, "js");
		Core::getTPL()->addHTMLHeaderFile("lib/jquery.countdown-ru.js?".CLIENT_VERSION, "js");
		$result=sqlQuery("SELECT c.name,c.mode,ah.*,uw.username won,ul.username lost,asl.key,asl.key2 FROM (((".PREFIX."artefact_history ah RIGHT OUTER JOIN ".PREFIX."construction c ON ah.id=c.buildingid) LEFT OUTER JOIN ".PREFIX."user uw ON uw.userid=ah.won_id) LEFT OUTER JOIN ".PREFIX."user ul ON ul.userid=ah.lost_id) LEFT OUTER JOIN ".PREFIX."assault asl ON ah.assaultid=asl.assaultid WHERE c.buildingid=".$typeid." ORDER BY ah.time DESC");
		$row=sqlFetch($result);
		if($row["mode"]==UNIT_TYPE_ARTEFACT)
		{
			saveAssaultReportSID();

			$entries=array();
			Core::getLanguage()->load("info,ArtefactInfo");
			Core::getTPL()->assign("typeid", $typeid);
			Core::getTPL()->assign("name", Core::getLanguage()->getItem($row["name"]));
			Core::getTPL()->assign("description", Core::getLanguage()->getItem($row["name"]."_FULL_DESC"));
			Core::getTPL()->assign("pic", Image::getImage(getUnitImage($row["name"]), Core::getLanguage()->getItem($row["name"]), null, null, "leftImage"));
			if($row["id"])
			{
				do
				{
					$entry["time"]=$row["time"]?date("d.m.Y, H:i:s",$row["time"]):(Core::getLanguage()->getItem("LONG_AGO"));
					$entry["won"]=$row["won"];
					$entry["lost"]=is_null($row["lost"])?(Core::getLanguage()->getItem("ALIENS")):$row["lost"];
					$entry["assault_url"]=$row["assaultid"]==0
						? ""
						: socialUrl(RELATIVE_URL."AssaultReport.php?id=".$row["assaultid"]."&".($row["key2"]?"key2=".$row["key2"]:"key=".$row["key"]));
					$entries[]=$entry;
				}
				while($row = sqlFetch($result));
			}
			sqlEnd($result);
			Core::getTPL()->addLoop("entries", $entries);

			$expiring=array();
			$disappearing=array();
			$result=sqlQuery("SELECT mode,time,data FROM ".PREFIX."events WHERE mode=".EVENT_ARTEFACT_EXPIRE." OR mode=".EVENT_ARTEFACT_DISAPPEAR);
			while($row=sqlFetch($result))
			{
				$data=unserialize($row["data"]);
				switch($row["mode"])
				{
				case EVENT_ARTEFACT_EXPIRE:
					$expiring[$data["artid"]]=$row["time"];
					break;
				case EVENT_ARTEFACT_DISAPPEAR:
					$disappearing[$data["artid"]]=$row["time"];
					break;
				}
			}
			sqlEnd($result);

			$artefacts=array();
			$select = array("a2u.artid", "a2u.planetid", "pl.planetname", "ads.movable", "ads.unique", "ads.usable",
				"a2u.active", "a2u.next_activate", "ads.delay", "ads.use_duration", "a2u.times_left", "ads.use_times",
				"ads.lifetime", "ads.max_active", "ads.effect_type", "ads.trophy_chance");
			$joins = "INNER JOIN ".PREFIX."construction b ON (b.buildingid = ads.typeid)";
			$joins .= "LEFT JOIN ".PREFIX."artefact2user a2u ON (a2u.typeid = ads.typeid)";
			$joins .= "LEFT JOIN ".PREFIX."planet pl ON (pl.planetid = a2u.planetid)";
			$result = sqlSelect("artefact_datasheet ads", $select, $joins, "(a2u.userid = ".sqlUser()." OR a2u.userid IS NULL) AND ads.typeid = ".$typeid);
			if(Core::getDB()->num_rows($result))
			{
				$row = sqlFetch($result);
				$tmp="";
				$flags=array("unique"=>$row["unique"],"usable"=>$row["usable"],"movable"=>$row["movable"],"static"=>!$row["movable"]);
				Core::getTPL()->assign("unique",$row["unique"]);
				Core::getTPL()->assign("usable",$row["usable"]);
				Core::getTPL()->assign("movable",$row["movable"]);
				Core::getTPL()->assign("static",$row["static"]);
				Core::getTPL()->assign("use_times",$row["use_times"]);
				if($row["use_duration"]) Core::getTPL()->assign("duration",getTimeTerm($row["use_duration"]));
				if($row["delay"]) Core::getTPL()->assign("delay",getTimeTerm($row["delay"]));
				if($row["lifetime"]) Core::getTPL()->assign("lifetime",getTimeTerm($row["lifetime"]));
				foreach($flags as $name=>$flag)
					if($flag)
						$tmp.=Image::getImage(Artefact::getFlagImage($name),Core::getLanguage()->getItem("FLAG_".strtoupper($name)),16,16);
				Core::getTPL()->assign("flags",$tmp);
				$act_deact=$flags["static"]&&!$flags["usable"];
				$use=$flags["usable"];
				do
				{
					$id = $row["artid"];
					if(empty($id)) break;
					$artefacts[$id]["id"]=$id;
					$artefacts[$id]["planetname"] = empty($row["planetname"]) ? Core::getLanguage()->getItem("IN_FLIGHT") : $row["planetname"];
					$artefacts[$id]["next_activate"]=$row["next_activate"];
					$artefacts[$id]["times_left"]=$row["times_left"];
					$artefacts[$id]["active"]=$row["active"];
					$can_activate = Artefact::hasFreeSlots(NS::getUser()->get("userid"), $row["planetid"]);
					if($row["active"])
					{
						if($act_deact) $artefacts[$id]["deactivate"]=1;
						if($use&&isset($expiring[$id]))
						{
							$timeleft=$expiring[$id]-time();
							$artefacts[$id]["eff_counter"]="<script type='text/javascript'>
								$(function () {
									$('#eff_counter{$id}').countdown({until: {$timeleft}, compact: true});
							});
							</script>
								<span id='eff_counter".$id."'>".getTimeTerm($timeleft)."</span>";
						}
					}
					else
					{
						if((time()>=$row["next_activate"])&&$act_deact&&$can_activate&&Artefact::checkRequirements($typeid,NS::getUser()->get("userid"),NS::getUser()->get("curplanet"))) $artefacts[$id]["activate"]=1;
					}
					if((time()>=$row["next_activate"])&&$use&&Artefact::checkRequirements($typeid,NS::getUser()->get("userid"),NS::getUser()->get("curplanet"))) $artefacts[$id]["use"]=1;
					if(time()<$row["next_activate"])
					{
						$timeleft=$row["next_activate"]-time();
						$action=$act_deact?"activate":($use?"use":"");
						$after=$action?("<input type=\"submit\" name=\"".$action."\" value=\"".Core::getLanguage()->getItem(strtoupper($action))."\" class=\"button\" onClick=\"document.getElementById(\'artid\').value=".$id."\" />"):"";
						$artefacts[$id]["counter"]="<script type='text/javascript'>
							$(function () {
								$('#counter{$id}').countdown({until: {$timeleft}, compact: true, onExpiry: function() {
									$('#text{$id}').html('{$after}');
								}});
						});
						</script>
							<span id='text".$id."'><span id='counter".$id."'>".getTimeTerm($timeleft)."</span></span>";
					}
					if(isset($disappearing[$id]))
					{
						$timeleft=$disappearing[$id]-time();
						$artefacts[$id]["life_counter"]="<script type='text/javascript'>
							$(function () {
								$('#life_counter{$id}').countdown({until: {$timeleft}, compact: true});
						});
						</script>
							<span id='life_counter".$id."'>".getTimeTerm($timeleft)."</span>";
					}
				}
				while($row = sqlFetch($result));
				Core::getTPL()->addLoop("artefacts", $artefacts);
			}
			sqlEnd($result);

			Core::getTPL()->display("artefactinfo");
		}
		else
		{
			sqlEnd($result);
			throw new GenericException("Unknown artefact. You'd better don't mess with the URL.");
		}
		return $this;
	}

	/**
	* Uses an usable artefact.
	*
	* @param integer	artid
	* @param integer	typeid
	*
	* @return ArtefactInfo
	*/
	protected function useArtefact($artid,$typeid)
	{
		if($artid&&$typeid&&Artefact::checkRequirements($typeid,NS::getUser()->get("userid"),NS::getUser()->get("curplanet")))
		{
			Artefact::onUse($artid,$typeid,NS::getUser()->get("userid"),NS::getUser()->get("curplanet"));
		}
		return $this->index();
	}

	/**
	* Activates a given artefact.
	*
	* @param integer	artid
	* @param integer	typeid
	*
	* @return Artefacts
	*/
	protected function activateArtefact($artid,$typeid)
	{
		if($artid&&$typeid&&Artefact::checkRequirements($typeid,NS::getUser()->get("userid"),NS::getUser()->get("curplanet")))
		{
			Artefact::activate($artid,$typeid,NS::getUser()->get("userid"),NS::getUser()->get("curplanet"));
		}
		return $this->index();
	}

	/**
	* Deactivates a given artefact.
	*
	* @param integer	artid
	* @param integer	typeid
	*
	* @return Artefacts
	*/
	protected function deactivateArtefact($artid,$typeid)
	{
		if($artid&&$typeid)
		{
			Artefact::deactivate($artid,$typeid,NS::getUser()->get("userid"),NS::getUser()->get("curplanet"));
		}
		return $this->index();
	}

}
?>
