<?php
/**
* Handles new stock.
*
* Oxsar http://oxsar.ru, http://oxsar.ru
*
*
*
*
*/

class StockNew extends Page {

	protected $lot_types = array(
		0 => "ALL",
		1 => "RESOURCES",
		2 => "FLEET",
		3 => "ARTEFACT",
		);

	public function __construct()
	{
		parent::__construct();

		Core::getLanguage()->load("Statistics");
		Core::getLanguage()->load("info,mission");

		$this->setPostAction("go", "showExchanges")
			->setPostAction("step1", "selectFleet")
			->setPostAction("step2", "lotOptions")
			->setPostAction("step3", "lotOptions2")
			->setPostAction("step4", "addLot")
			->addPostArg("showExchanges", "sort_field")
			->addPostArg("showExchanges", "sort_order")
			->addPostArg("showExchanges", "page")
			->addPostArg("selectFleet", "eid")
			->addPostArg("lotOptions", null)
			->addPostArg("lotOptions2", "lot_type")
			->addPostArg("lotOptions2", "delivery_hydro")
			->addPostArg("lotOptions2", "delivery_percent")
			->addPostArg("lotOptions2", "ttl")
			->addPostArg("addLot", "quant")
			->addPostArg("addLot", "price")
			->addPostArg("addLot", "discount")
			->addPostArg("addLot", "res_type")
			->addPostArg("addLot", "resource")
			->addPostArg("addLot", "wide_price")
			->addPostArg("addLot", "premium")
			;

		$this->proceedRequest();
	}

	protected function index()
	{
        if(NS::getUser()->get("umode")){
            Logger::dieMessage('UMODE_ENABLED');
            // doHeaderRedirection("game.php/Stock", false);
        }
        /* if(!EXCH_ENABLED && !EXCH_NEW_PROFIT_TYPE){
            Logger::addMessage('SUSPENDED_FOR_TECHNICAL_REASONS');
            Core::getTPL()->display("empty");
            return;
        } */

		$this->showExchanges('title', 'asc', 1);
	}

	protected function showExchanges($sort_field, $sort_order, $page)
	{
		Core::getTPL()->addHTMLHeaderFile("stats.js?".CLIENT_VERSION);

		$i = 0;
		$featured = array();
		$result = sqlSelect("exchange_lots", "brokerid, max(lid) as lid", $join,
			"planetid IN (SELECT planetid FROM ".PREFIX."planet WHERE userid=".sqlUser().") GROUP BY brokerid",
			"lid DESC", 5);
		while($row = sqlFetch($result))
		{
			$featured[$row["brokerid"]] = $i++;
		}
		sqlEnd($result);
		if( !isset($featured[EXCH_INVIOLABLE]) )
		{
			$featured[EXCH_INVIOLABLE] = $i++;
		}

		$sort_fields = array(
			"title" => "e.title",
			"player" => "u.username",
			"planet" => "p.planetname",
			"fee" => "e.fee"
			);

		if (!isset ($sort_fields[$sort_field]))
			$sort_field = "title";

		$sort_order = strtolower($sort_order);
		if($sort_order != 'asc' && $sort_order != 'desc')
			$sort_order = 'asc';

		$order = $sort_fields[$sort_field] . " " .$sort_order;
		if($sort_field == "title")
		{
			$order = "bp.level ".($sort_order == 'asc' ? 'desc' : 'asc') . ", $order";
		}

		$select = array(
			"e.*", "bp.level",
			"p.planetname", "p.userid",
			"u.username",
			"g.galaxy", "g.system", "g.position",
			"unit.quantity",
			);

		$join = "inner join ".PREFIX."planet p on p.planetid = e.eid"
			. " inner join ".PREFIX."user u on u.userid = p.userid"
			. " inner join ".PREFIX."galaxy g on g.planetid = p.planetid"
			. " inner join ".PREFIX."building2planet bp on bp.planetid = p.planetid and bp.buildingid = ".sqlVal(UNIT_EXCHANGE)
			. " left join ".PREFIX."unit2shipyard unit on unit.planetid = p.planetid and unit.unitid = " . sqlVal(UNIT_EXCH_SUPPORT_RANGE);

		$sqlRows = array(array(), array());
		if($page == 1 && count($featured) > 0)
		{
			$where = "e.eid IN (".sqlArray(array_keys($featured)).")";
			$result = sqlSelect("exchange e", $select, $join, $where);
			while($row = sqlFetch($result))
			{
				$row["featured_sort_order"] = $featured[$row["eid"]];
				$sqlRows[0][] = $row;
			}
			sqlEnd($result);

			usort($sqlRows[0], function($a, $b){
				return $a["featured_sort_order"] - $b["featured_sort_order"];
			});
		}

		$where = count($featured) > 0 ? "e.eid NOT IN (".sqlArray(array_keys($featured)).")" : "";
		$result_count = sqlSelectField("exchange e", "count(*)", $join, $where);
		$pages = ceil($result_count / Core::getOptions()->get("USER_PER_PAGE"));

		$start = createPaginator($pages, $page, Core::getOptions()->get("USER_PER_PAGE"));
		$max = Core::getOptions()->get("USER_PER_PAGE");

		$result = sqlSelect("exchange e", $select, $join, $where, $order, "$start, $max");
		while($row = sqlFetch($result))
		{
			$sqlRows[1][] = $row;
		}
		sqlEnd($result);

		$langNext = Core::getLang()->getItem("NEXT");
		$langNotEnoughRange = Core::getLang()->getItem("NOT_ENOUGH_EXCHANGE_RANGE");
		$langFreeSlots = Core::getLang()->getItem("EXCH_FREE_SLOTS");
		$langNoSlots = Core::getLang()->getItem("EXCH_NO_SLOTS");
		$langBan = Core::getLang()->getItem("EXCH_BAN");

		$ban_exch = NS::getEH()->getExchBans();
		//print_r($ban_exch);

		//print_r();

		// while($row = Core::getDatabase()->fetch($result))
		for($j = 0; $j < 2; $j++)
		{
			if(empty($sqlRows[$j]))
			{
				continue;
			}
			$i = 0;
			$exchanges = array();
			foreach($sqlRows[$j] as $row)
			{
				$i++;
				$row["i"] = $i;
				$dist = NS::getDistanceSystems($row["galaxy"], $row["system"], $row["pos"]);
				$row["distance"] =	$dist < 0 ? -$dist . " g" : $dist . " s"; //NS::getDistance($row["galaxy"], $row["system"], $row["position"]);
				$row["coords"] = Link::get("game.php/go:Galaxy/galaxy:{$row['galaxy']}/system:{$row['system']}", "[{$row['galaxy']}:{$row['system']}:{$row['position']}]");
				$exch_range = round($row["quantity"] * NS::getResearch(UNIT_COMPUTER_TECH, $row["userid"]) * EXCH_RADIUS_FACTOR);

				$distance = $dist >= 0 ? $dist : abs(EXCH_RADIUS_SYSTEMS_PER_GALAXY * $dist);
				//$row["title_link"] = $exch_range >= $distance ? "<a href=\"{$_SERVER["PHP_SELF"]}/step:fleet/id:{$row["eid"]}\">{$row["title"]}</a>" : $row["title"];
				$row["class"] = $exch_range >= $distance || $row["eid"] == EXCH_INVIOLABLE ? "" : " class=\"false\"";
				$slots = Exchange::freeSlots($row["eid"]);
				if ($exch_range < $distance && $row["eid"] != EXCH_INVIOLABLE)
				{
					$row["status"] = $langNotEnoughRange;
					$row["action"] = "";

				}
				else if ($slots < 1 && $row["eid"] != EXCH_INVIOLABLE)
				{
					$row["status"] = $langNoSlots;
					$row["action"] = "";
				}
				elseif ( in_array($row["eid"], $ban_exch) && $row["eid"] != EXCH_INVIOLABLE )
				{
					$row["status"] = $langBan;
					$row["action"] = "";
				}
				else
				{
					if ($row["eid"] == EXCH_INVIOLABLE)
					{
						$slots = '*';
					}
					$row["status"] = "$langFreeSlots $slots";
					$row["action"] = "<form method=\"post\" action=\"\"><input type=\"hidden\" name=\"eid\" value=\"{$row["eid"]}\"/>"
						. "<input type=\"submit\" name=\"step1\" value=\"$langNext\" class=\"button\" /></form>";
				}

				$exchanges[] = $row;
			}
			Core::getTPL()->addLoop($j ? "exchanges" : "featured_exchanges", $exchanges);
		}
		//debug_var($exchanges, 'exchanges');
		// sqlEnd($result);

		Core::getTPL()->assign("sort_field", $sort_field);
		Core::getTPL()->assign("sort_order", $sort_order);
		Core::getTPL()->assign($sort_field, Image::getImage($sort_order.'.png', $sort_order));
		Core::getTPL()->display("stock_new_1");
	}

	/**
	 *
	 * Adds Users Artefacts to template
	 */
	protected function addArtefacts()
	{
		$select = array('b.buildingid', 'b.name',
			'a2u.active', 'a2u.times_left', 'a2u.delay_eventid', 'a2u.expire_eventid', 'a2u.lifetime_eventid',
			'ads.movable', 'ads.unique', 'ads.usable', 'ads.use_duration', 'ads.delay', 'ads.use_times', 'ads.lifetime', 'ads.max_active', 'ads.effect_type', 'ads.trophy_chance',
			'a2u.planetid', 'a2u.artid', 'e.time as cur_lifetime',
			);
		$joins	= 'INNER JOIN '.PREFIX.'construction b ON (b.buildingid = a2u.typeid)';
		$joins .= 'INNER JOIN '.PREFIX.'artefact_datasheet ads ON (a2u.typeid = ads.typeid)';
		$joins .= 'LEFT JOIN '.PREFIX.'events e ON e.eventid = a2u.lifetime_eventid';
		$result = sqlSelect(
			'artefact2user a2u',
			$select,
			$joins,
			'a2u.deleted=0 AND a2u.delay_eventid=0'
				. ' AND a2u.expire_eventid=0 AND ads.movable=1'
				. ' AND a2u.active=0 AND a2u.userid='.sqlUser()
				. ' AND a2u.planetid='.sqlPlanet() . ' AND a2u.times_left >= ads.use_times ',
			'b.display_order ASC, ads.unique DESC, ads.typeid ASC'
		);
		$artefacts = array();
		while($row = sqlFetch($result))
		{
			$id = $row["artid"];
			$type = $row["buildingid"];
			if ( $type == ARTEFACT_PACKED_BUILDING || $type == ARTEFACT_PACKED_RESEARCH )
			{
				$temp 		= sqlSelectRow('artefact2user', '*', '', "artid = ".sqlVal($id));;
				$img_link = $temp['construction_id'] . '_' . $temp['level'] . '_' . $type;
				$row['link'] = $type . '_' . $temp['level'] . '_' . $temp['construction_id'];
				Artefact::setViewParams($artefacts[$id], $row, 60,
					// YII_GAME_DIR.'/index.php?r=artefact2user_YII/image&id='.$id.''
					artImageUrl("image", "id=".$id, false)
					);
			}
			else
			{
				Artefact::setViewParams($artefacts[$id], $row, 60);
			}
			if($row["cur_lifetime"] > 0 && $row["cur_lifetime"] > time())
			{
				$timeleft = max(1, $row["cur_lifetime"] - time());
				$artefacts[$id]["disappear_counter"] = "<script type='text/javascript'>
					$(function () {
						$('#disappear_counter{$artid}').countdown({until: {$timeleft}, compact: true, onExpiry: function() {
							$('#disappear_text{$id}').html('-');
						}});
					});
				</script>
				<span id='disappear_text{$artid}'><span id='disappear_counter{$artid}'>".getTimeTerm($timeleft)."</span></span>";
			}
		}
		sqlEnd($result);
		Core::getTPL()->addLoop("artefacts", $artefacts);
	}

	/**
	 *
	 * Shows Fleet you can select
	 * @param int $eid
	 */
	protected function selectFleet($eid)
	{
		Core::getTPL()->assign("can_send_fleet", true);
		Core::getTPL()->assign("can_send_expo", true);
        Core::getTPL()->assign("max_ships", MAX_SHIPS);
        Core::getTPL()->assign("max_ships_grade", MAX_SHIPS_GRADE);
        Core::getTPL()->assign("max_ships_size", (MAX_SHIPS_GRADE - 1));
		Core::getTPL()->addHTMLHeaderFile("fleet.js?".CLIENT_VERSION, "js");

		if (Exchange::freeOfficeSlots() < 1)
			Logger::dieMessage("EXCH_NO_FREE_OFFICE_SLOTS");
		$select = array("b.buildingid", "b.name", "u2s.quantity - u2s.damaged as quantity", "d.capicity", "d.speed");
		$joins	= "LEFT JOIN ".PREFIX."construction b ON (b.buildingid = u2s.unitid)";
		$joins .= "LEFT JOIN ".PREFIX."ship_datasheet d ON (d.unitid = u2s.unitid)";
		$result = sqlSelect("unit2shipyard u2s", $select, $joins,
			"b.mode = ".UNIT_TYPE_FLEET." AND u2s.planetid = ".sqlPlanet()." AND u2s.quantity - u2s.damaged > 0");
		$f = array();
		while($row = sqlFetch($result))
		{
			$id = $row["buildingid"];
			$speed = NS::getSpeed($id, $row["speed"]);
			if($speed > 0)
			{
				$f[$id]["name"] = Link::get("game.php/UnitInfo/".$id, Core::getLanguage()->getItem($row["name"]));
				$f[$id]["image"] = Link::get("game.php/UnitInfo/".$id, Image::getImage(getUnitImage($row["name"]), Core::getLanguage()->getItem($row["name"]), 60));
				$f[$id]["quantity_raw"] = $row["quantity"];
				$f[$id]["quantity"] = $row["quantity"];
				$f[$id]["speed"] = fNumber($speed);
				$f[$id]["capicity"] = fNumber($row["capicity"] * $row["quantity"]);
				$f[$id]["id"] = $id;
			}
		}
		sqlEnd($result);

		self::addArtefacts();

		$data["brokerid"] = $eid;

		Core::getQuery()->delete("temp_fleet", "planetid = ".sqlPlanet());
		Core::getQuery()->insert("temp_fleet", array("planetid", "data"), array(NS::getUser()->get("curplanet"), serialize($data)));

		// Hook::event("MISSION_FLEET_LIST", array(&$f));
		Core::getTPL()->addLoop("fleet", $f);

		Core::getTPL()->assign("newLot", true);
		Core::getTPL()->assign("canSendFleet", true);

		Core::getTPL()->display("missions");
		return $this;
	}

	protected function lotOptions($ships)
	{
		Core::getTPL()->addHTMLHeaderFile("lib/jquery.countdown.js?".CLIENT_VERSION, "js");
		Core::getTPL()->addHTMLHeaderFile("fleet.js?".CLIENT_VERSION, "js");
		$row = sqlSelectRow("temp_fleet", "data", "", "planetid = ".sqlPlanet());
		if($row)
		{
			$ship = array();
			foreach($ships as $k => $v)
			{
				if(is_numeric($k) && $v > 0)
				{
					$ship[$k] = $v;
					break;
				}
			}
			$artefact = array();
			if ( isset($ships['art']) && !empty($ships['art']) )
			{
				$art_key = key($ships['art']);
				$artefact[$art_key] = 1 ;
				$ship['art'] = $artefact;
				$art_valid = sqlSelectField(
					'artefact2user AS a2u',
					' ( a2u.times_left >= ads.use_times ) AS valid',
					'INNER JOIN '.PREFIX.'artefact_datasheet ads ON (a2u.typeid = ads.typeid)',
					' a2u.artid = ' . sqlVal($art_key)
				);
				if( !$art_valid )
				{
					return $this->index();
				}
			}

			$select = array("u2s.unitid");
			$joins	= "LEFT JOIN ".PREFIX."construction b ON (b.buildingid = u2s.unitid)";
			$joins .= "LEFT JOIN ".PREFIX."ship_datasheet d ON (d.unitid = u2s.unitid)";
			$result_inner = sqlSelect("unit2shipyard u2s", $select, $joins, "b.mode=".UNIT_TYPE_FLEET." AND u2s.planetid = ".sqlPlanet());
			while($row_inner = sqlFetch($result_inner))
			{
				if( isset($ship[$row_inner['unitid']]) )
				{
					$ship[$row_inner['unitid']] = min( MAX_SHIPS ,$ship[$row_inner['unitid']] );
				}
			}
			sqlEnd($result_inner);

			$move_data = Mission::calcMoveToCoords($galaxy, $system, $position, $ship, true);
			$data = $move_data["ships"];
			$capicity = $move_data["capicity"];
			$consumption = $move_data["consumption"];
			$speed = $move_data["maxspeed"];
			$fleet_size = $move_data["fleet_size"];
			foreach($data as & $ship)
			{
				$ship["damaged"] = 0;
				$ship["shell_percent"] = 100;
			}
			if(count($data) > 0 && $speed > 0)
			{
				Core::getTPL()->assign("capicity_raw", $capicity);
				Core::getTPL()->assign("basicConsumption", $consumption);
				Core::getTPL()->assign("fleetSize", $fleet_size);
				Core::getTPL()->assign("capicity", fNumber($capicity));
				Core::getTPL()->assign("fleet", getUnitListStr($data));
				if ( !empty($artefact) )
				{
					$temp_lot_types = Exchange::getLotTypes_opt(false, ETYPE_ARTEFACT, ETYPE_ARTEFACT, array(ETYPE_ARTEFACT));
				}
				else
				{
					$temp_lot_types = Exchange::getLotTypes_opt(false, 0, false, array(ETYPE_RESOURCE, ETYPE_FLEET));
				}
				Core::getTPL()->assign("lot_types", $temp_lot_types);
				Core::getTPL()->assign("oGalaxy", NS::getPlanet()->getData("galaxy"));
				Core::getTPL()->assign("oSystem", NS::getPlanet()->getData("system"));
				Core::getTPL()->assign("oPos", NS::getPlanet()->getData("position"));
				Core::getTPL()->assign("quantity", $ship[$k]);
				Core::getTPL()->assign("ttl", EXCH_DEF_TTL);
				Core::getTPL()->assign("distance", 0);


				$tmp = unserialize($row["data"]);
				$data["brokerid"] = $tmp["brokerid"];
				$data["consumption"] = $consumption;
				$data["maxspeed"] = $speed;
				$data["capacity"] = $capicity;
				$data["galaxy"] = NS::getPlanet()->getData("galaxy");
				$data["system"] = NS::getPlanet()->getData("system");
				$data["position"] = NS::getPlanet()->getData("position");

				$join = "left outer join ".PREFIX."unit2shipyard u on u.planetid=e.eid and u.unitid = ".sqlVal(UNIT_EXCH_SUPPORT_RANGE)
					. " inner join ".PREFIX."galaxy g on g.planetid=e.eid";
				$row = sqlSelectRow("exchange e", array("e.*", "u.quantity", "g.galaxy", "g.system"), $join, "e.eid = ".sqlVal($data["brokerid"]));
				$data["fee"] = $row["fee"];
				if($data["brokerid"] != EXCH_INVIOLABLE)
				{
					$dist = NS::getDistanceSystems($row["galaxy"], $row["system"], 0);
					$distance = $dist >= 0 ? $dist : EXCH_RADIUS_SYSTEMS_PER_GALAXY * (-$dist);
					if ($distance > round($row["quantity"] * NS::getResearch(UNIT_COMPUTER_TECH, $row["uid"]) * EXCH_RADIUS_FACTOR))
					{
						Logger::dieMessage("NOT_ENOUGH_EXCHANGE_RANGE");
					}
				}
				$data = serialize($data);
				sqlUpdate("temp_fleet", array("data" => $data), "planetid = ".sqlPlanet(). ' ORDER BY planetid' );

				Core::getTPL()->display("stock_new_2");
				exit();
			}
		}
		return $this->index();
	}

	protected function lotOptions2($lot_type, $delivery_hydro, $delivery_percent, $ttl)
	{
		Core::getTPL()->addHTMLHeaderFile("fleet.js?".CLIENT_VERSION);

		$result = sqlSelect("temp_fleet", "data", "", "planetid = ".sqlPlanet());
		if($row = sqlFetch($result))
		{
			$data = unserialize($row["data"]);
			// Set spaceships
			if ($data["capacity"] < $delivery_hydro)
			{
				Logger::dieMessage(Core::getLanguage()->getItem("NOT_ENOUGH_CAPACITY"));
			}

			foreach($data as $key => $value)
			{
				if(is_numeric($key))
				{
					fixUnitDamagedVars($value);

					$data["ships"][$key]["id"] = $key;
					$data["ships"][$key]["quantity"] = $value["quantity"];
					$data["ships"][$key]["damaged"] = $value["damaged"];
					$data["ships"][$key]["shell_percent"] = $value["shell_percent"];
					$data["ships"][$key]["name"] = $value["name"];
					if ( isset($value['art_ids']) )
					{
						$data["ships"][$key]["art_ids"] = $value['art_ids'];
					}
				}
			}
			$data["speed"] = 100;

			if ( Exchange::isValidLotType($lot_type) )
			{
				$data["lot_type"] = $lot_type;
			}
			else
			{
				Logger::dieMessage(Core::getLang()->getItem("ERROROUS_PARAM"));
			}
			$ttl = $ttl > 0 ? $ttl : EXCH_MAX_TTL;
			$data["ttl"] = $ttl;

			$delivery_hydro = $delivery_hydro > 0 ? $delivery_hydro : 0;
			$data["delivery_hydro"] = $delivery_hydro;

			$delivery_percent = clampVal($delivery_percent, 0, 100);
			$data["delivery_percent"] = $delivery_percent;

			if ($delivery_hydro > 0)
			{
				$consumption = $delivery_hydro;
			}
			else
			{
				$consumption = 0;
			}

			$planet_ratio = NS::getPlanet()->getCreditRatio();
			$default_price = 0;
			if ($lot_type == ETYPE_FLEET)
			{
				$fleet_keys = array_keys($data["ships"]);
				$data["lot"] = $data["ships"][$fleet_keys[0]]["id"];
				$data["amount"] = $data["ships"][$fleet_keys[0]]["quantity"];

				$cost = Exchange::lotCost($data["lot"], $data["amount"]);
				$min_price = $cost["credit"];
				$min_price += $min_price * ($data["fee"] + EXCH_SELLER_MIN_PROFIT + EXCH_DEF_DISCOUNT) / 100.0;
				$min_price = ceil( min($min_price, Exchange::getMinLotPrice($data["lot"], $data["amount"])) );

				$default_price = $cost["credit"];
                if(EXCH_NEW_PROFIT_TYPE){
                    $default_price *= 1 + EXCH_SELLER_DEF_PROFIT/100.0;
                    $default_price /= (100 - $data["fee"]) / 100.0;
                    $default_price /= (100 - EXCH_MERCHANT_PREMIUM_COMMISSION*2) / 100.0;
                    $default_price /= (100 - EXCH_DEF_DISCOUNT) / 100.0;
                }else{
                    $default_price += $default_price * ($data["fee"] + EXCH_SELLER_DEF_PROFIT + EXCH_DEF_DISCOUNT) / 100.0;
                }
				$default_price = ceil($default_price);

				$max_price = $cost["credit"];
				$max_price += $max_price * ($data["fee"] + EXCH_SELLER_MAX_PROFIT + EXCH_DEF_DISCOUNT) / 100.0;
				$max_price = ceil($max_price);
			}
			if ($lot_type == ETYPE_ARTEFACT)
			{
				$fleet_keys = array_keys($data["ships"]);
				$data["lot"] = $data["ships"][$fleet_keys[1]]["id"];
				$data["amount"] = $data["ships"][$fleet_keys[1]]["quantity"];
				$cost = Exchange::lotCost($data["lot"], $data["amount"], key($data["ships"][$fleet_keys[1]]["art_ids"]));

				$min_price = $cost["credit"] * ( EXCH_ART_MIN_PERCENT / 100.0 );
				$min_price += $min_price * ($data["fee"] + EXCH_SELLER_MIN_PROFIT + EXCH_DEF_DISCOUNT) / 100.0;
				$min_price = ceil($min_price);

				$max_price = $cost["credit"]; // * ( EXCH_ART_MIN_PERCENT / 100.0 ));
				$max_price += $max_price * ($data["fee"] + EXCH_SELLER_ART_MAX_PROFIT + EXCH_DEF_DISCOUNT) / 100.0;
				$max_price = ceil($max_price);

				$default_price = ceil($cost["credit"] * ( EXCH_ART_DEF_PERCENT / 100.0 ));
			}

			$min_price = ceil(max(1, $min_price));
			$max_price = ceil(max(1, $max_price));

			Core::getQuery()->update("temp_fleet", "data", serialize($data), "planetid = ".sqlPlanet() . ' ORDER BY planetid');

			Core::getTPL()->assign("fee", $data["fee"]);
			Core::getTPL()->assign("fee_str", fNumber($data["fee"]));
			Core::getTPL()->assign("exch_seller_min_profit", EXCH_SELLER_MIN_PROFIT);
			Core::getTPL()->assign("exch_seller_def_profit", EXCH_SELLER_DEF_PROFIT);
			Core::getTPL()->assign("quantity", ( (isset($data["amount"]) && !empty($data["amount"])) ? $data["amount"] :0 ));
			Core::getTPL()->assign("loadRes", ($lot_type == ETYPE_RESOURCE)? 1 : 0 );
			Core::getTPL()->assign("art_type", ($lot_type != ETYPE_ARTEFACT)? 1 : 0 );
			Core::getTPL()->assign("is_artefact", ($lot_type == ETYPE_ARTEFACT)? 1 : 0 );
			Core::getTPL()->assign("fleet", getUnitListStr($data["ships"]));
			Core::getTPL()->assign("metal", NS::getPlanet()->getData("metal"));
			Core::getTPL()->assign("silicon", NS::getPlanet()->getData("silicon"));
			Core::getTPL()->assign("hydrogen", NS::getPlanet()->getData("hydrogen") - $consumption);
			Core::getTPL()->assign("capacity", $data["capacity"] - $consumption);
			Core::getTPL()->assign("rest", fNumber($data["capacity"] - $consumption));
			Core::getTPL()->assign("cost_met", fNumber($cost["metal"]));
			Core::getTPL()->assign("cost_sil", fNumber($cost["silicon"]));
			Core::getTPL()->assign("cost_hyd", fNumber($cost["hydrogen"]));
			Core::getTPL()->assign("met_equiv", fNumber($cost["metal_equiv"]));
			Core::getTPL()->assign("market_metal", MARKET_BASE_CURS_METAL);
			Core::getTPL()->assign("market_silicon", MARKET_BASE_CURS_SILICON);
			Core::getTPL()->assign("market_hydrogen", MARKET_BASE_CURS_HYDROGEN);
			Core::getTPL()->assign("market_credit", MARKET_BASE_CURS_CREDIT);
			Core::getTPL()->assign("planet_ratio", $planet_ratio["planet_ratio"]);
			Core::getTPL()->assign("real_price", $cost['credit']);
			Core::getTPL()->assign("real_price_str", fNumber($cost['credit'], 2));
			Core::getTPL()->assign("min_price", $min_price);
			Core::getTPL()->assign("max_price", $max_price);
			Core::getTPL()->assign("min_price_str", fNumber($min_price, 2));
			Core::getTPL()->assign("max_price_str", fNumber($max_price, 2));
			Core::getTPL()->assign("met_eq", ( (isset($cost["metal_equiv"]) && !empty($cost["metal_equiv"])) ? $cost["metal_equiv"] : 0 ) );
			Core::getTPL()->assign("sellerMaxProfit", EXCH_SELLER_MAX_PROFIT);
			Core::getTPL()->assign("min_ally_disc", EXCH_MIN_DISCOUNT);
			Core::getTPL()->assign("max_ally_disc", EXCH_MAX_DISCOUNT);
            if(EXCH_NEW_PROFIT_TYPE){
                // Core::getTPL()->assign("premium_price_p", EXCH_PREMIUM_PERCENT);
                // Core::getTPL()->assign("min_premium_price", EXCH_PREMIUM_MIN_COST);
            }else{
                Core::getTPL()->assign("premium_price", EXCH_PREMIUM_LOT_COST);
            }
			if( $default_price != 0 )
			{
				Core::getTPL()->assign("def_price", ceil(max(1, $default_price)));
			}
			else
			{
				Core::getTPL()->assign("def_price", $min_price);
			}
			Core::getTPL()->assign("def_discount", EXCH_DEF_DISCOUNT);
			Core::getTPL()->display("stock_new_3");

			exit();
		}
		sqlEnd($result);
		return $this;
	}

	protected function addLot($quant, $price, $discount, $res_type, $resource, $wide_price, $premium)
	{
		if( !NS::isFirstRun("StockNew::addLot:" . $_SESSION["userid"] ?? 0) )
		{
			error_log('Exchange Multiclick: ' . __CLASS__);
			Logger::dieMessage('ERROROUS_PARAM');
		}

		if(NS::isPlanetUnderAttack())
		{
			Logger::dieMessage('PLANET_UNDER_ATTACK');
		}

		$quant = max(1, (int)$quant);
		$price = max(0, (float)$price);
		$discount = max(0, (int)$discount);
		$resource = max(0, (int)$resource);
		$wide_price = max(0, (int)$wide_price);
		$res_type = (int)$res_type;

		$result = sqlSelect("temp_fleet", "data", "", "planetid = ".sqlPlanet());
		if($row = sqlFetch($result))
		{
			$data = unserialize($row["data"]);

			if ($data["brokerid"] != EXCH_INVIOLABLE && Exchange::freeSlots($data["brokerid"]) < 1)
			{
				Logger::dieMessage(Core::getLang()->getItem("EXCH_NO_SLOTS"));
			}

			if($price <= 0 || ($res_type < 0 && $resource <= 0) )
			{
				Logger::dieMessage("ERROROUS_PARAM");
			}

			$data["price"] = $price;
			$data["discount"] = clampVal($discount, EXCH_MIN_DISCOUNT, EXCH_MAX_DISCOUNT); // ? $discount : Logger::dieMessage("ERROROUS_PARAM");
			$data["res_type"] = $res_type;
			$data["resource"] = $resource;

			//$res_types = array("METAL", "SILICON", "HYDROGEN");
			if ($res_type >= -3 && $res_type <= -1)
			{
				if($data['lot_type'] != ETYPE_RESOURCE)
				{
					Logger::dieMessage("ERROROUS_PARAM");
				}
				//die($res_type);
				$data["lot"] = $res_type;
				$data["amount"] = max(0, $resource);
				if($data["capacity"] < $data["delivery_hydro"] + $data["amount"])
				{
					Logger::dieMessage(Core::getLanguage()->getItem("NOT_ENOUGH_CAPACITY"));
				}
			}
			else if( $data['lot_type'] == ETYPE_RESOURCE )
			{
				Logger::dieMessage("ERROROUS_PARAM");
			}
			elseif( $data['lot_type'] != ETYPE_ARTEFACT )
			{
				//die($res_type);
				$fleet_keys = array_keys($data["ships"]);
				$data["lot"] = $data["ships"][$fleet_keys[0]]["id"];
				$data["amount"] = $data["ships"][$fleet_keys[0]]["quantity"];
			}
			else
			{
				$fleet_keys = array_keys($data["ships"]);
				$data["lot"] = $data["ships"][$fleet_keys[1]]["id"];
				$data["amount"] = $data["ships"][$fleet_keys[1]]["quantity"];
			}

			if ( $data['lot_type'] == ETYPE_ARTEFACT )
			{
				$fleet_keys = array_keys($data["ships"]);
				$cost = Exchange::lotCost($data["lot"], $data["amount"], key($data["ships"][$fleet_keys[1]]["art_ids"]));

				$min_price = $cost["credit"] * ( EXCH_ART_MIN_PERCENT / 100.0 );
				$min_price += $min_price * ($data["fee"] + EXCH_SELLER_MIN_PROFIT + $data["discount"]) / 100.0;

				$max_price = $cost["credit"]; // * ( EXCH_ART_MIN_PERCENT / 100.0 ));
				$max_price += $max_price * ($data["fee"] + EXCH_SELLER_ART_MAX_PROFIT + $data["discount"]) / 100.0;
			}
			else
			{
				$cost = Exchange::lotCost($data["lot"], $data["amount"]);

				$min_price = $cost["credit"];
				$min_price += $min_price * ($data["fee"] + EXCH_SELLER_MIN_PROFIT + $data["discount"])/100.0;
                $min_price = min($min_price, Exchange::getMinLotPrice($data["lot"], $data["amount"]));

				$max_price = $cost["credit"];
				$max_price += $max_price * ($data["fee"] + EXCH_SELLER_MAX_PROFIT + $data["discount"])/100.0;
			}
			$min_price = max(1, ceil($min_price));
			$max_price = max(1, floor($max_price));

			$wide_price = clampVal($wide_price, 0, $min_price-1);
			$data["wide_price"] = $wide_price;
			$min_price -= $wide_price;
			$max_price += $wide_price;

			$min_price = max(1, $min_price - 1);
			$max_price = max(1, $max_price + 1);

			if($data["price"] < $min_price)
			{
				Logger::dieMessage(Core::getLang()->getItem("EXCH_MIN_PRICE_ERR"));
			}
			else if($data["price"] > $max_price)
			{
				Logger::dieMessage(Core::getLang()->getItem("EXCH_MAX_PRICE_ERR"));
			}

			$quant = clampVal($quant, 0, $data["amount"]);
			$data["quant"] = $quant;

			$data["raising_date"] = time();

			$data["expiry_date"] = $data["raising_date"] + min(EXCH_MAX_TTL, $data["ttl"]) * 3600 * 24;

			// debug_var($data, "[data]"); exit;

			$row_data["ships"] = $data["ships"];
			$row_data["sgalaxy"] = NS::getPlanet()->getData("galaxy");
			$row_data["ssystem"] = NS::getPlanet()->getData("system");
			$row_data["sposition"] = NS::getPlanet()->getData("position");
			$row_data["maxspeed"] = $data["maxspeed"];
			$row_data["consumption"] = $data["consumption"];
			$row_data["metal"] = $row_data["silicon"] = $row_data["hydrogen"] = 0;
			$row_data["capacity"] = $data["capacity"];

            if(EXCH_NEW_PROFIT_TYPE){
                $data["premium_price"] = $premium ? max(EXCH_PREMIUM_MIN_COST, $data["price"] * EXCH_PREMIUM_PERCENT/100.0) : 0;
            }else{
                $data["premium_price"] = $premium ? EXCH_PREMIUM_LOT_COST : 0;
            }
			$reserv_result = Exchange::reserveLot($data, $data["brokerid"]);
			if ( $reserv_result !== true )
			{
				Logger::dieMessage("BIG_LOT_" . $reserv_result);
			}

			// Created exch expire event

			$exchid = sqlInsert(
				"exchange_lots",
				array(
					"planetid" => NS::getUser()->get("curplanet"),
					"brokerid" => $data["brokerid"],
					"raising_date" => $data["raising_date"],
					"expiry_date" => $data["expiry_date"],
					"delivery_hydro" => $data["delivery_hydro"],
					"delivery_percent" => $data["delivery_percent"],
					"type" => $data["lot_type"],
					"data" => serialize($row_data),
					"lot" => $data["lot"],
					"amount" => $data["amount"],
					"lot_min_amount" => $data["quant"],
					"price" => $data["price"],
					"fee" => $data["fee"],
					"ally_discount" => $data["discount"],
					"status" => ESTATUS_OK,
					"lot_amount" => $data["amount"],
					"lot_price" => $data["price"],
                    "lot_unit_price" => max(EXCH_MIN_UNIT_PRICE, ($data["price"] + $data["wide_price"]) / $data["amount"]),
					"featured_date" => $premium ? time() : null,
				)
			);
			// План 86 audit: $exchid идёт в artefact2user.lot_id и в
			// EXCH_EXPIRE event payload. Без guard'а артефакт связали
			// бы с чужим лотом, событие истечения сработало бы на
			// чужом лоте.
			if ($exchid === false) {
				Logger::dieMessage('DB_ERROR_EXCHANGE');
			}

			if( $data['lot_type'] == ETYPE_ARTEFACT )
			{
				$k = array_keys($data["ships"]);
				$unitid = $data["ships"][$k[0]]["id"];
				$art_id = 0;
				if ( isset($data["ships"][$k[1]]["art_ids"]) )
				{
					$art_id = key($data["ships"][$k[1]]["art_ids"]);
					// Update By Pk
					sqlUpdate('artefact2user', array('lot_id' => $exchid), 'artid = '. $art_id); // . ' ORDER BY artid' );
				}
			}

			$temp = NS::getEH()->addEvent(
				EVENT_EXCH_EXPIRE,
				$data["expiry_date"],
				NS::getUser()->get("curplanet"),
				NS::getUser()->get("userid"),
				NS::getUser()->get("curplanet"),
				array('lot_id' => $exchid),
				null,
				null,
				null,
				null
			);

			$seller = sqlSelectRow("user", "username", "", "userid = ".sqlUser());
			$exchange = sqlSelectRow("exchange", "title", "", "eid = ".sqlVal($data["brokerid"]));

			Exchange::resolveLotNames($data);
			Core::getTPL()->assign("exchange", $exchange["title"]);
			Core::getTPL()->assign("seller", $seller["username"]);
			Core::getTPL()->assign("raising_date", Date::timeToString(1, $data["raising_date"]));
			Core::getTPL()->assign("expiry_date", Date::timeToString(1, $data["expiry_date"]));

			if ($data["delivery_hydro"] == 0 || $data["delivery_percent"] == 0)
				$delivery_str = Core::getLang()->getItem("EXCH_DELIVERY_BUYER");
			else
				$delivery_str = Core::getLang()->getItemWith("EXCH_DELIVERY_STR", array("hydro" => fNumber($data["delivery_hydro"]), "percent" => fNumber($data["delivery_percent"])));
			Core::getTPL()->assign("delivery_str", $delivery_str);
			Core::getTPL()->assign("lot_name", $data["lot_name"]);
			Core::getTPL()->assign("amount", $data["amount"]);
			Core::getTPL()->assign("amount_str", fNumber($data["amount"]));
			Core::getTPL()->assign("lot_min_amount", $data["quant"]);
			Core::getTPL()->assign("lot_min_amount_str", fNumber($data["quant"]));
			Core::getTPL()->assign("price", $data["price"]);
			Core::getTPL()->assign("price_str", fNumber($data["price"], 2));
			Core::getTPL()->assign("ally_discount", fNumber($data["discount"]));

			Core::getTPL()->assign("showall", true);
			Core::getTPL()->display("lot_details");
			exit();
		}
		sqlEnd($result);
		return $this;
	}
}
?>
