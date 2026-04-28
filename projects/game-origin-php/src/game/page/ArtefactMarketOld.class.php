<?php
/**
* Allows to buy artefacts.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class ArtefactMarketOld extends Page
{
	/**
	* Constructor. Handles requests for this page.
	*
	* @return void
	*/
	public function __construct()
	{
		parent::__construct();
		if(!NS::getUser()->get("umode"))
		{
			$this->setPostAction("buy_res", "buyArtefactRes")
				->addPostArg("buyArtefactRes", "typeid");
			$this->setPostAction("buy_cred", "buyArtefactCred")
				->addPostArg("buyArtefactCred", "typeid");
		}

		Core::getLanguage()->load("info,ArtefactInfo");
//		try {
			$this->proceedRequest();
//		} catch(Exception $e) {
//			$e->printError();
//		}
		return;
	}

	/**
	* Index action.
	*
	* @return ArtefactMarket
	*/
	protected function index()
	{
		$flags = array("unique","usable","static","attach");

		$select = array("b.buildingid", "b.name", "b.basic_metal", "b.basic_silicon", "b.basic_hydrogen", "b.basic_energy",
			"b.basic_credit", "ads.can_fly", "ads.unique", "ads.usable", "ads.use_duration", "ads.delay", "ads.lifetime",
			"ads.max_active", "ads.planet_effect", "ads.can_capture");
		$joins = "INNER JOIN ".PREFIX."construction b ON (b.buildingid = ads.typeid)";
		$result = sqlSelect("artefact_datasheet ads", $select, $joins, "ads.can_buy=1", "ads.typeid");
		if(Core::getDB()->num_rows($result))
		{
			while($row = sqlFetch($result))
			{
				$id = $row["buildingid"];
				$artefacts[$id]["quantity"]++;
				
				Artefact::setViewParams($artefacts[$id], $row, 110);

				$artefacts[$id]["buy_res"] = $artefacts[$id]["buy_cred"] = 1;
				foreach(array("metal", "silicon", "hydrogen", "energy", "credit") as $res_name)
				{
					$required_res = $row["basic_".$res_name];
					switch($res_name)
					{
					case "energy":
						$available_res = NS::getPlanet()->getEnergy();
						break;

					case "credit":
						$available_res = NS::getUser()->get("credit");
						if($required_res>0) $artefacts[$id]["use_credit"]=1;
						break;

					default:
						$available_res = NS::getPlanet()->getData($res_name);
						if($required_res>0) $artefacts[$id]["use_resources"]=1;
						break;
					}
					$artefacts[$id][$res_name."_required"] = $required_res > 0 ? fNumber($required_res) : "";
					$artefacts[$id][$res_name."_notavailable"] = $required_res > 0 && $available_res < $required_res ? fNumber($available_res - $required_res) : "";
					if(!empty($artefacts[$id][$res_name."_notavailable"]))
					{
						if($res_name!="credit") $artefacts[$id]["buy_res"]=0;
						if(($res_name=="credit")||($res_name=="energy")) $artefacts[$id]["buy_cred"]=0;
					}
				}
			}
		}
		sqlEnd($result);
		Core::getTPL()->addLoop("artefacts", $artefacts);
		Core::getTPL()->display("artefactmarket");
		return $this;
	}

	/**
	* Buys an artefact for resources.
	*
	* @param integer	typeid
	*
	* @return ArtefactMarket
	*/
	protected function buyArtefactRes($typeid)
	{
		if($typeid)
		{
			$select = array("b.basic_metal", "b.basic_silicon", "b.basic_hydrogen", "b.basic_energy", "ads.can_fly", "ads.usable", "ads.use_times", "ads.lifetime");
			$joins = "INNER JOIN ".PREFIX."construction b ON (b.buildingid = ads.typeid)";
			$result = sqlSelect("artefact_datasheet ads", $select, $joins, "ads.can_buy=1 AND ads.typeid=".sqlVal($typeid));
			if($row = sqlFetch($result))
			{
				$names=array();
				$rests=array();
				$buy=true;
				foreach(array("metal", "silicon", "hydrogen", "energy") as $res_name)
				{
					$required_res = $row["basic_".$res_name];
					switch($res_name)
					{
					case "energy":
						if(NS::getPlanet()->getEnergy()<$required_res) $buy=false;
						break;
					default:
						$names[]=$res_name;
						$rests[]=$rest=NS::getPlanet()->getData($res_name)-$required_res;
						if($rest<0) $buy=false;
						break;
					}
				}
				if($buy)
				{
					Core::getQuery()->update("planet", $names, $rests, "planetid = ".sqlPlanet());
					$active=$row["can_fly"]&&!$row["usable"];
					Core::getQuery()->insert("artefact2user", array("typeid","userid","planetid","active","next_activate","times_left"), array($typeid,NS::getUser()->get("userid"),NS::getUser()->get("curplanet"),(int)$active,0,$row["use_times"]));
					if($row["lifetime"])
						NS::getEH()->addEvent(EVENT_ARTEFACT_DISAPPEAR,$row["lifetime"]+time(),NS::getUser()->get("curplanet"),NS::getUser()->get("userid"),NS::getUser()->get("curplanet"),array("artid"=>Core::getDB()->insert_id(),"typeid"=>$typeid));
					doHeaderRedirection("game.php/ArtefactMarket", false);
				}
			}
			sqlEnd($result);
		}
		return $this->index();
	}

	/**
	* Buys an artefact for credit.
	*
	* @param integer	typeid
	*
	* @return ArtefactMarket
	*/
	protected function buyArtefactCred($typeid)
	{
		if($typeid)
		{
			$select = array("b.basic_energy", "b.basic_credit", "ads.can_fly", "ads.usable", "ads.use_times", "ads.lifetime");
			$joins	= "INNER JOIN ".PREFIX."construction b ON (b.buildingid = ads.typeid)";
			$result = sqlSelect("artefact_datasheet ads", $select, $joins, "ads.can_buy=1 AND ads.typeid=".$typeid);
			if($row = sqlFetch($result))
			{
				$rest=NS::getUser()->get("credit");
				$buy=true;
				foreach(array("credit", "energy") as $res_name)
				{
					$required_res = $row["basic_".$res_name];
					switch($res_name)
					{
					case "energy":
						if(NS::getPlanet()->getEnergy()<$required_res) $buy=false;
						break;
					case "credit":
						$rest-=$required_res;
						if($rest<0) $buy=false;
						break;
					}
				}
				if($buy)
				{
					Core::getQuery()->update("officer", array("credit"), array($rest), "userid = ".sqlUser());
					$active=$row["can_fly"]&&!$row["usable"];
					Core::getQuery()->insert("artefact2user", array("typeid","userid","planetid","active","next_activate","times_left"), array($typeid,NS::getUser()->get("userid"),NS::getUser()->get("curplanet"),(int)$active,0,$row["use_times"]));
					if($row["lifetime"])
						NS::getEH()->addEvent(EVENT_ARTEFACT_DISAPPEAR,$row["lifetime"]+time(),NS::getUser()->get("curplanet"),NS::getUser()->get("userid"),NS::getUser()->get("curplanet"),array("artid"=>Core::getDB()->insert_id(),"typeid"=>$typeid));
					doHeaderRedirection("game.php/ArtefactMarket", false);
				}
			}
			sqlEnd($result);
		}
		return $this->index();
	}

}
?>
