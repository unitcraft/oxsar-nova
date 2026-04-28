<?php
/**
* Stock module.
*
* Oxsar http://oxsar.ru
*
*
*/

class Stock extends Page {

	protected $exch_filters = array(
		0 => "---",
		1 => "EXCH_MY_LOTS",
		2 => "EXCH_MY_EXCH",
		//3 => "EXCH_GROUP_EXCH",
		);

	public function __construct()
	{
		parent::__construct();

        // Logger::dieMessage('SUSPENDED_FOR_TECHNICAL_REASONS');

		Core::getLanguage()->load("Statistics");
		Core::getLanguage()->load("info");
		Core::getTPL()->addHTMLHeaderFile("stats.js?".CLIENT_VERSION);
		Core::getTPL()->assign("pageaddress", RELATIVE_URL."game.php/");

		$this
			->setPostAction("go", "showLots")
			->setPostAction("buy_amount", "buyLot")
			->setPostAction("details", "showLotDetails")
			->setPostAction("recall", "recall")
			->addPostArg("showLots", "date_first")
			->addPostArg("showLots", "date_last")
			->addPostArg("showLots", "sort_field")
			->addPostArg("showLots", "sort_order")
			->addPostArg("showLots", "page")
			->addPostArg("showLots", "lot_type")
			->addPostArg("showLots", "exch_filter")
			->addPostArg("showLots", "whereid")
			->addPostArg("showLots", "id")
			->addPostArg("showLotDetails", "lid")
			->addPostArg("recall", "lid")
			->addPostArg("buyLot", "lid")
			->addPostArg("buyLot", "buy_amount")
			->setGetAction("go", "StockBan", "ban")
			->addGetArg("ban", "id")
			->setGetAction("go", "StockLotPremium", "premiumLot")
			->addGetArg("premiumLot", "id")
			->setGetAction("go", "StockLotRecall", "recallByGET")
			->addGetArg("recallByGET", "id")
			;

		$this->proceedRequest();
	}

	protected function premiumLot($id)
	{
		$id = max(0, (int)$id);
        if(!NS::isFirstRun("Stock::premiumLot($id)")){
            doHeaderRedirection("game.php/Stock", false);
        }
        if(EXCH_NEW_PROFIT_TYPE){
            $row = sqlSelectRow("exchange_lots", "*", "",
                    'lid='.sqlVal($id).'
                        AND status IN ('.sqlArray(ESTATUS_OK, ESTATUS_SUSPENDED).')
                        AND expiry_date > '.sqlVal($time - EXCH_PREMIUM_LOT_EXPIRY_TIME-10));
            $premium_price = $row ? max(EXCH_PREMIUM_MIN_COST, $row["price"] * EXCH_PREMIUM_PERCENT/100.0) : 0;
        }else{
            $premium_price = EXCH_PREMIUM_LOT_COST;
        }
        if($row && NS::getUser()->get("credit") >= $premium_price)
		{
            $time = time();
			sqlUpdate("exchange_lots", array(
				"featured_date" => $time
				), 'lid='.sqlVal($id).' AND status IN ('.sqlArray(ESTATUS_OK, ESTATUS_SUSPENDED).') AND expiry_date > '.sqlVal($time - EXCH_PREMIUM_LOT_EXPIRY_TIME-10));
			if( $row = sqlSelectRow("exchange_lots", "*", "", "lid=".sqlVal($id)." AND featured_date=".sqlVal($time)) )
			{
				$data = array( $id => &$row );
				Exchange::resolveLotNames($data);

				$res_log = NS::updateUserRes(array(
					"block_minus"	=> true,
					"type" 			=> RES_UPDATE_EXCH_LOT_PREMIUM,
					"ownerid"		=> $id,
					"userid"		=> NS::getUser()->get("userid"),
					"credit"		=> - $premium_price,
				));

				// if(empty($res_log["minus_blocked"]))
				{
					new AutoMsg(
						MSG_CREDIT,
						NS::getUser()->get("userid"),
						time(),
						array(
							'credits'	=> $premium_price,
							'msg' 		=> 'MSG_CREDIT_EXCHANGE_PREMIUM_LOT',
							'content'	=> array( 'name' => $row["lot_name"] )
							)
					);
				}
			}
		}
		doHeaderRedirection("game.php/Stock", false);
	}

	protected function ban($id)
	{
		$id = max(0, (int)$id);
		$userid = NS::getUser()->get("userid");
		if(NS::isFirstRun("Stock::ban($id, $userid)"))
		{
			$lot = sqlSelectRow(
				'exchange_lots l',
				'l.*',
				'',
				'lid = ' . sqlVal( $id ) . ' AND status=' . sqlVal( ESTATUS_OK )
			);
			if ( !empty($lot) )
			{
				$time 				= EXCH_BAN_TIME + time();
				$exchang_id 		= $lot['brokerid'];
				$owner_planet_data 	= sqlSelectRow( 'planet', '*', '', 'planetid = '.sqlVal( $lot['planetid'] ) );
				$owner_id 			= $owner_planet_data['userid'];
				$data 				= array( 'exch_id' => $exchang_id );
				// Remove lot
				Exchange::sendBackLot($id);
				// Ban User
				$ban_event_id = NS::getEH()->addEvent(
					EVENT_EXCH_BAN,
					$time,
					$lot['planetid'],
					$owner_id,
					null,
					$data
				);
				// Remove lot
				sqlUpdate( 'exchange_lots', array('status' => ESTATUS_BANNED), 'lid = ' . sqlVal( $id ) . ' ORDER BY lid');
				// Remove Lot event
				$result = sqlSelect(
					'events',
					'eventid, data',
					'',
					'mode = ' . EVENT_EXCH_EXPIRE
						. ' AND user = ' . sqlVal( $owner_id )
						. ' AND planetid = ' . sqlVal( $lot['planetid'] )
						. ' AND destination = ' . sqlVal( $lot['planetid'] )
						. ' AND processed = ' . EVENT_PROCESSED_WAIT
				);
				while( $row = sqlFetch($result) )
				{
					$row['data'] = unserialize($row['data']);
					if($row['data']['exchid'] == $id || $row['data']['lot_id'] == $id)
					{
						NS::getEH()->removeEvent($row['eventid'], "banned");
						break;
					}
				}
				sqlEnd($result);
			}
		}
		doHeaderRedirection("game.php/Stock", false);
		return $this->index();
	}

	protected function deleteExpireEvent($id)
	{
		$lot = sqlSelectRow(
			'exchange_lots l',
			'l.*',
			'',
			'lid = ' . sqlVal( $id )
		);
		if ( !empty($lot) )
		{
			$time 				= EXCH_BAN_TIME + time();
			$exchang_id 		= $lot['brokerid'];
			$owner_planet_data 	= sqlSelectRow( 'planet', '*', '', '	planetid = '.sqlVal( $lot['planetid'] ) );
			$owner_id 			= $owner_planet_data['userid'];
			$data 				= array( 'exch_id' => $exchang_id );
			// Remove Lot event
			$result = sqlSelect(
				'events',
				'eventid, data',
				'',
				'mode = ' . EVENT_EXCH_EXPIRE
					. ' AND user = ' . sqlVal( $owner_id )
					. ' AND planetid = ' . sqlVal( $lot['planetid'] )
					. ' AND destination = ' . sqlVal( $lot['planetid'] )
					. ' AND processed = ' . EVENT_PROCESSED_WAIT
			);
			while( $row = sqlFetch($result) )
			{
				$row['data'] = unserialize($row['data']);
				if( $row['data']['exchid'] == $id || $row['data']['lot_id'] == $id )
				{
					NS::getEH()->removeEvent($row['eventid']);
					break;
				}
			}
			sqlEnd($result);
			return true;
		}
		return false;
	}

	protected function index()
	{
		$this->showLots(false, false, false, false, false, false, false, false, false);
		return $this;
	}

	protected function showLots($date_first, $date_last, $sort_field, $sort_order, $page, $lot_type, $exch_filter, $whereid, $id)
	{
        Core::getTPL()->addHTMLHeaderFile("lib/jquery.countdown.js?".CLIENT_VERSION, "js");

		$exch_filters = "";
		for ($i = 0; $i < count($this->exch_filters); $i++)
		{
			$exch_filters .= createOption($i, Core::getLang()->getItem($this->exch_filters[$i]), $exch_filter == $i);
		}

		$min_const = 60*60*24*30;
		$date_first = Date::validateDate($date_first, $min_const, 'max');
		$date_last = Date::validateDate($date_last);

		$sort_fields = array(
			"date" => "l.raising_date",
			"seller" => "p.planetname",
			"lot" => "l.lot",
			"lot_price" => "l.price",
			"lot_amount" => "l.amount",
			"distance" => "",
			);


		if (!isset ($sort_fields[$sort_field]))
			$sort_field = "date";

		if($sort_order != 'asc' && $sort_order != 'desc')
			$sort_order = $sort_field == 'date' ? 'desc' : 'asc';
		$sort = $sort_field == "distance" ? "" : $sort_fields[$sort_field] ." $sort_order";

		$where = "l.status IN (".sqlArray(ESTATUS_OK, ESTATUS_SUSPENDED).")";

		if (!is_numeric($lot_type) || $lot_type < 1 || $lot_type > Exchange::lotTypes_count() -1)
			$lot_type = 0;
		else
			$where .= " AND l.type = " . sqlVal($lot_type);
		$groupby = "";
		switch ($exch_filter)
		{
		case 1:
			$where .= " AND p.userid = " . sqlUser();
			break;
		case 2:
			$where .= " AND e.uid = " . sqlUser();
			break;
		case 3:
			$groupby = "l.brokerid";
			break;

		default:
			break;
		}

		$valid_where = array("broker" => "l.brokerid", "planet" => "l.planetid");
		if (isset($valid_where[$whereid]) && is_numeric($id))
			$where .= " AND {$valid_where[$whereid]}= ".sqlVal($id);

        if($exch_filter != 1){ // my lots, used to view lots by admin under banned user
            $where .= " AND u.umode = 0 AND u.userid NOT IN (SELECT userid FROM ".PREFIX."ban_u)";
        }

		$joins = "INNER JOIN ".PREFIX."exchange e on l.brokerid = e.eid"
			. " INNER JOIN ".PREFIX."planet p on p.planetid = l.planetid"
			. " INNER JOIN ".PREFIX."galaxy g on g.planetid = l.planetid"
			. " INNER JOIN ".PREFIX."user u on u.userid = p.userid"
			;

		$select = array(
			"l.*",
			"e.eid", "e.title", "e.uid AS Owner_ID",
			"p.planetname", "p.userid", "u.username",
			"g.galaxy", "g.system", "g.position"
			);

		$sqlRows = array();
		// if($page == 1)
		// if( isAdmin() )
		{
			$featured_lots = array();
			$featured_where = "l.status in (".sqlArray(ESTATUS_OK, ESTATUS_SUSPENDED).") "
                    . ($lot_type ? " AND l.type = ".sqlVal($lot_type) : "")
                    . " AND l.featured_date IS NOT NULL"
                    ;
            if($exch_filter != 1){ // my lots, used to view lots by admin under banned user
                $featured_where .= " AND u.userid NOT IN (SELECT userid FROM ".PREFIX."ban_u)";
            }
			$result = sqlSelect("exchange_lots l", $select, $joins, $featured_where, "l.featured_date DESC", EXCH_PREMIUM_LIST_MAX_SIZE);
			while($row = sqlFetch($result))
			{
				$sqlRows[0][] = $row;
				$featured_lots[$row["lid"]] = 1;
			}
			sqlEnd($result);

			if(count($featured_lots) > 0)
			{
				$where .= " AND l.lid NOT IN (".sqlArray(array_keys($featured_lots)).")";
			}
		}

		$result_count = sqlSelectField("exchange_lots l", "count(*)", $joins, $where);
		$pages = ceil($result_count / Core::getOptions()->get("USER_PER_PAGE"));

		$start = createPaginator($pages, $page, Core::getOptions()->get("USER_PER_PAGE"));
		$max = Core::getOptions()->get("USER_PER_PAGE");

		$result = sqlSelect("exchange_lots l", $select, $joins, $where, $sort, "$start, $max", $groupby);
		while($row = sqlFetch($result))
		{
			$sqlRows[1][] = $row;
		}
		sqlEnd($result);

		$planet_stack = NS::getPlanetStack();
		// $exchange_rate = Exchange::getExchangeRate();
		$formaction = socialUrl($_SERVER["PHP_SELF"]);
		$mobile_skin = (int)isMobileSkin();
		$vip_image = Image::getImage("exch_lot_vip.png", Core::getLang()->getItem("PREMIUN_LOT_SLOGAN"));

		for($j = 0; $j < 2; $j++)
		{
			if(empty($sqlRows[$j]))
			{
				continue;
			}
			$i = 0;
			$lots = array();
			$i = $start;
			$lots_id = array();
			foreach($sqlRows[$j] as $row)
			{
				$i++;
				//echo $row['Owner_ID'];die;
				$data = unserialize($row["data"]);

                $lot_valid = true;
                $lot_error = '';
                $lot_expire_time = $row["expiry_date"];
                $lot_timer = '';
                if($row['type'] == ETYPE_ARTEFACT){
					// $artid = key($data['ships'][key($data['ships'])]['art_ids']);
                    $art_data = $data;
					$temp = array_keys($art_data['ships']);
					$art_data = $art_data['ships'][$temp[1]]['art_ids'][key($art_data['ships'][$temp[1]]['art_ids'])];
                    $artid = $art_data['artid'];
                    $art_row = sqlSelectRow('artefact2user a', '*', '
                        LEFT JOIN '.PREFIX.'artefact_datasheet d ON d.typeid=a.typeid
                        LEFT JOIN '.PREFIX.'events e ON e.eventid = a.lifetime_eventid',
                        'a.artid='.  sqlVal($artid));
                    if(!$art_row){
                        $lot_valid = false;
                        $lot_error = Core::getLanguage()->getItem('STOCK_ARTEFACT_NOT_FOUND');
                    }elseif($art_row['times_left'] < $art_row['use_times']){
                        $lot_valid = false;
                        $lot_error = Core::getLanguage()->getItem('STOCK_ARTEFACT_USED');
                    }elseif(!$art_row['eventid']){
                        $lot_valid = false;
                        $lot_error = Core::getLanguage()->getItem('STOCK_ARTEFACT_LIFETIME_NOT_FOUND');
                    }elseif($art_row['deleted'] || $art_row['time'] - time() <= 0){
                        $lot_valid = false;
                        $lot_error = Core::getLanguage()->getItem('STOCK_ARTEFACT_EXPIRED');
                    }elseif($art_row['time'] - time() < 60){
                        // $lot_valid = false;
                        $lot_error = Core::getLanguage()->getItem('STOCK_ARTEFACT_LIFETIME_TOO_SHORT');
                    }
                    $lot_expire_time = min($lot_expire_time, $art_row['time']);
                }
                if($lot_valid){
                    $lot_timeleft = $lot_expire_time - time();
                    if($lot_timeleft <= 0){
                        // $lot_timer = '00:00:00';
                        $lot_valid = false;
                        if(!$lot_error){
                            $lot_error = Core::getLanguage()->getItem('STOCK_LOT_EXPIRED');
                        }
                    }elseif($lot_timeleft < 60*60*48){
                        $lot_timer_id = $row['lid'].'timer';
                        $lot_timer = "
<script type='text/javascript'>
    $(function (){
        $('#{$lot_timer_id}').countdown({until: {$lot_timeleft}, compact: true, onExpiry: function(){
            $('#{$lot_timer_id}').html('<span class=false>00:00:00</span>');
        }});
    });
</script>
<span id='{$lot_timer_id}'>".getTimeTerm($lot_timeleft)."</span>";
                    }
                }
                $row['valid'] = $lot_valid;
                $row['error'] = $lot_error;
                $row['timer'] = $lot_timer;

				$row["i"] = $i;
				$row["date"] = date("d.m.y H:i", $row["date"]);
				$dist = NS::getDistanceSystems($row["galaxy"], $row["system"], $row["position"]);
				$row["real_dist"] = NS::getDistance($row["galaxy"], $row["system"], $row["position"]);
				$row["fly_time"] = NS::getFlyTime($row["real_dist"], $data["maxspeed"], 100, $row["featured_date"]);
				$row["fly_time"] = getTimeTerm($row["fly_time"]);
				$row["distance"] =	$dist < 0 ? -$dist . " g" : $dist . " s";
                $row['org_price'] = $row["price"];

				$exchange_rate = Exchange::getExchangeRate(null, $row["featured_date"]);
				$price = $row["price"] * $exchange_rate;
				$amount = $row["amount"];
				$lot_min_amount = $row["lot_min_amount"];
				$is_min_active = $row["lot_min_amount"] > 0 && $row["lot_min_amount"] < $row["amount"];

				$row["amount"] = fNumber($row["amount"]);
				$row["lot_min_amount"] = $is_min_active ? fNumber($row["lot_min_amount"]) : "";
				$row["price"] = fNumber($price, 2);
				$row["min_price"] = $is_min_active ? fNumber($lot_min_amount / $amount * $price, 2) : "";
				if($row["ally_discount"] > 0 && NS::getUser()->get("aid"))
				{
					$seller = sqlSelectRow("planet p", array("u2a.userid", "u2a.aid"),
						"inner join ".PREFIX."user2ally u2a on p.userid = u2a.userid",
						"p.planetid = ".sqlVal($row["planetid"]));
					if($seller["aid"] == NS::getUser()->get("aid"))
					{
						$trade_union = true;
					}
					else
					{
						require_once(APP_ROOT_DIR."game/Relation.class.php");
						$Relations = new Relation($seller["userid"], $seller["aid"]);
						$trade_union = 4 == $Relations->getAllyRelation(NS::getUser()->get("aid")); //4 -- trade-union
					}

					$discount_class = "";
					if ($trade_union)
					{
						$discount_price = $price * (1 - $row["ally_discount"] / 100);
						$row["disc_price"] = fNumber($discount_price, 2);
						$row["disc_min_price"] = $is_min_active ? fNumber($lot_min_amount / $amount * $discount_price, 2) : "";
					}
				}
				$action = "";
				$planetprefix = "";
				// if(isAdmin())
				{
					$row["planetname"] = "[".$row["galaxy"].":".$row["system"].":".$row["position"]."]";
					$planetprefix = $row["username"] . " ";
				}
				if ($whereid != "planet")
					$row["planetname"] = "<a href=\"#\" onclick=\"goWhere('planet', {$row["planetid"]}); return false;\">{$row["planetname"]}</a>";
				if ($whereid != "broker")
					$row["title"] = "<a href=\"#\" onclick=\"goWhere('broker', {$row["brokerid"]}); return false;\">{$row["title"]}</a>";

				$row["planetname"] = $planetprefix . $row["planetname"];

				$row["action"] = "";
				if ($row["status"] == ESTATUS_SUSPENDED)
				{
					$row["action"] = Core::getLang()->getItem("EXCH_EXECUTING");
				}
				else if(in_array($row["planetid"], $planet_stack))
				{
					$action_string = Core::getLang()->getItem("RECALL");
					$url = socialUrl(RELATIVE_URL."game.php/StockLotRecall/".$row["lid"]);
					$confirm_message = Core::getLang()->getItemWith("CONFIRM_LOT_RECALL", array(
						"name" => "{##lot_name_ext##}",
						));
					$row["action"] = "<div><input type='button' value='$action_string' class='button' onclick=\"confirm_dialog('$confirm_message', {id: 'recall{$row["lid"]}', url: '$url', mobile: $mobile_skin}); return false\" /></div>";
					/*
					$row["action"] = "<form method=\"post\" action=\"$formaction\">"
						."\r\n<input type=\"hidden\" name=\"lid\" value=\"{$row["lid"]}\" />"
						."\r\n<input type=\"submit\" name=\"recall\" value=\"$action_string\" />"
						."\r\n</form>";
					*/
				}
				else
				{
                    if($lot_valid){
                        $action_string = Core::getLang()->getItem("BUY");
                        $row["action"] = "<form method=\"post\" action=\"$formaction\">"
                            ."\r\n<input type=\"hidden\" name=\"lid\" value=\"{$row["lid"]}\" />"
                            ."\r\n<input type=\"submit\" name=\"details\" value=\"$action_string\" class=\"button\" />"
                            ."\r\n</form>";
                    }
					if(isAdmin())
					{
						$action_string = Core::getLang()->getItem("RECALL");
						$url = socialUrl(RELATIVE_URL."game.php/StockLotRecall/".$row["lid"]);
						$confirm_message = Core::getLang()->getItemWith("CONFIRM_LOT_RECALL", array(
							"name" => "{##lot_name_ext##}",
							));
						$row["action"] .= "<div><input type='button' value='$action_string' class='button' onclick=\"confirm_dialog('$confirm_message', {id: 'recall{$row["lid"]}', url: '$url', mobile: $mobile_skin}); return false\" /></div>";
						/*
						$row["action"] .= " <form method=\"post\" action=\"$formaction\">"
							."\r\n<input type=\"hidden\" name=\"lid\" value=\"{$row["lid"]}\" />"
							."\r\n<input type=\"submit\" name=\"recall\" value=\"$action_string\" class=\"button\" />"
							."\r\n</form>";
						*/
					}
				}
				if($j)
				{
					$row['ban'] = false;
					if ( $row['Owner_ID'] == NS::getUser()->get("userid") && !in_array($row["planetid"], $planet_stack) )
					{
						$action_string = Core::getLang()->getItem("EXCH_RECALL_AND_BAN");
						$url = socialUrl(RELATIVE_URL."game.php/StockBan/".$row["lid"]);
						$confirm_message = Core::getLang()->getItemWith("CONFIRM_EXCH_RECALL_AND_BAN", array(
							"name" => "{##lot_name_ext##}",
							));
						$row["action"] .= "<div><input type='button' value='$action_string' class='button' onclick=\"confirm_dialog('$confirm_message', {id: 'ban{$row["lid"]}', url: '$url', mobile: $mobile_skin}); return false\" /></div>";
						/*
						$row['ban'] = true;
						$row['banlink'] = Link::get(
							"game.php/StockBan/".$row['lid'],
							Core::getLanguage()->getItem("BAN_USER")
						);
						*/
					}
				}
				$lots[] = $row;
			}
			Exchange::resolveLotNames($lots, true);
			if($j)
			{
                $action_string = Core::getLang()->getItem("PREMIUM");
                $time = time() - EXCH_PREMIUM_LOT_EXPIRY_TIME;
                foreach($lots as $key => $row)
                {
                    if($row["expiry_date"] > $time && $row['valid'])
                    {
                        if(EXCH_NEW_PROFIT_TYPE){
                            $premium_price = max(EXCH_PREMIUM_MIN_COST, $row['org_price'] * EXCH_PREMIUM_PERCENT/100.0);
                            $premium_price = round($premium_price, 2);
                        }else{
                            $premium_price = EXCH_PREMIUM_LOT_COST;
                        }
                        $premium_url = socialUrl(RELATIVE_URL."game.php/StockLotPremium/".$row["lid"]);
                        $confirm_message = Core::getLang()->getItemWith("CONFIRM_PREMIUN_LOT_USE_CREDIT", array(
                            "name" => "{##lot_name_ext##}",
                            "credit" => $premium_price,
                            ));
                        $lots[$key]["action"] = "<div><input type='button' value='$action_string' class='button' onclick=\"confirm_dialog('$confirm_message', {id: 'premium{$row["lid"]}', url: '$premium_url', mobile: $mobile_skin}); return false\" /></div>" . $row["action"];
                    }
                }

				if ($sort_field == "distance")
				{
					usort($lots, "Stock::compare_by_distance");
					if ($sort_order == "desc")
						$lots = array_reverse($lots);
				}
			}
			foreach($lots as $key => $row)
			{
				$lots[$key]["action"] = str_replace("{##lot_name_ext##}", strtr($row["lot_name_no_url"], array("\"" => "", "'" => ""))." | ".$row["amount"]." | ".$row["price"], $lots[$key]["action"]);
				if($lots[$key]["featured_date"])
				{
					// $lots[$key]["lot_name"] =  $vip_image . "&nbsp;" . $lots[$key]["lot_name"];
				}
                if($row['error']){
                    $lots[$key]["lot_name"] .= "<br /><span class='false'>{$row['error']}</span>";
                }
                if($row['timer']){
                    $lots[$key]["lot_name"] .= "<br />{$row['timer']}";
                }
			}

			Core::getTPL()->addLoop($j ? "lots" : "featured_lots", $lots);
		}

        Core::getTPL()->assign("vip_image", $vip_image);
		Core::getTPL()->assign("date_last", $date_max['string']);
		Core::getTPL()->assign("date_first", $date_min['string']);
		Core::getTPL()->assign("sort_field", $sort_field);
		Core::getTPL()->assign("sort_order", $sort_order);
		Core::getTPL()->assign($sort_field, Image::getImage($sort_order.'.png', $sort_order));

		$free = Exchange::freeOfficeSlots();
		$total = Exchange::freeOfficeSlots(true);
		$link = Core::getLang()->getItem("EXCH_NEW_LOT");
		if ($free > 0)
		{
			$link = Link::get('game.php/StockNew', $link, "", "", "", false, true, true, true);
//			$link = '<a href="'. RELATIVE_URL ."game.php/StockNew\">$link</a>";
		}
		// debug_var(array($link, $free, $total), "[link, free, total]");
		Core::getTPL()->assign("new_lot_link", $link);
		Core::getTPL()->assign("used_slots", $total - $free);
		Core::getTPL()->assign("total_slots", $total);
		Core::getTPL()->assign("has_tp", NS::getPlanet()->getBuilding(UNIT_EXCH_OFFICE));
		Core::getTPL()->assign("lot_types", Exchange::getLotTypes_opt(true, $lot_type));
		Core::getTPL()->assign("exch_filters", $exch_filters);
		Core::getTPL()->assign("merchant_mark_used", Artefact::getMerchantMark());
		Core::getTPL()->assign("premiun_list_max_size", Core::getLang()->getItemWith("PREMIUN_LIST_MAX_SIZE", array("count" => EXCH_PREMIUM_LIST_MAX_SIZE)));
		Core::getTPL()->display("stock");
	}

	protected function showLotDetails($id)
	{
		$id = max(0, (int)$id);

		Core::getTPL()->addHTMLHeaderFile("fleet.js?".CLIENT_VERSION);
		$select = array("l.*", "u.username", "e.title", "g.galaxy", "g.system", "g.position");
		$join = ""
			. " inner join ".PREFIX."exchange e on e.eid = l.brokerid"
			. " inner join ".PREFIX."galaxy g on g.planetid = l.planetid"
			. " inner join ".PREFIX."planet p on p.planetid = l.planetid"
			. " inner join ".PREFIX."user u on u.userid = p.userid"
			;
		if ($row = sqlSelectRow("exchange_lots l", $select, $join, "lid = ".sqlVal($id)))
		{
			$exchange_rate = Exchange::getExchangeRate(null, $row["featured_date"]);
			$row["price"] *= $exchange_rate;

			Exchange::resolveLotNames($row);

			$real_price = Exchange::lotCost($row["lot"], $row["amount"]);
			$real_price = $real_price["credit"];
			//t($real_price,'real_price');

			$data = unserialize($row["data"]);
			$distance = NS::getDistance($row["galaxy"], $row["system"], $row["position"]);
			//t($distance,'$distance');
			$mltp = 0.5;
			if ( $row['type'] != ETYPE_FLEET )
			{
				$mltp = 1;
			}
			$total_consumption = NS::getFlyConsumption($data["consumption"], $distance) * $mltp;
			//t($total_consumption,'total_consumption');

			if ( $row["ally_discount"] > 0 && NS::getUser()->get("aid") )
			{
				$seller = sqlSelectRow("planet p", array("u2a.userid", "u2a.aid"),
					"inner join ".PREFIX."user2ally u2a on p.userid = u2a.userid",
					"p.planetid = ".sqlVal($row["planetid"]));
				if($seller["aid"] == NS::getUser()->get("aid"))
				{
					$trade_union = true;
				}
				else
				{
					require_once(APP_ROOT_DIR."game/Relation.class.php");
					$Relations = new Relation($seller["userid"], $seller["aid"]);
					$trade_union = 4 == $Relations->getAllyRelation(NS::getUser()->get("aid")); //4 -- trade-union
				}

				$discount_class = "";
				if ($trade_union)
				{
					$discount_class = ' class="true"';
					$row["price"] *= 1 - $row["ally_discount"] / 100;
				}
			}

			Core::getTPL()->assign("ally_discount", $row["ally_discount"]);
			Core::getTPL()->assign("ally_discount_class", $discount_class);

			Core::getTPL()->assign("lid", $row["lid"]);
			Core::getTPL()->assign("exchange", $row["title"]);
			Core::getTPL()->assign("seller", $row["username"]);
			if ($row["delivery_hydro"] == 0 || $row["delivery_percent"] == 0)
			{
				$delivery_str = Core::getLang()->getItem("EXCH_DELIVERY_BUYER");
			}
			else
			{
				$delivery_str = Core::getLang()->getItemWith("EXCH_DELIVERY_STR", array("hydro" => fNumber($row["delivery_hydro"]), "percent" => fNumber($row["delivery_percent"])));
				//t($delivery_str,'delivery_str');
			}
			Core::getTPL()->assign("delivery_str", $delivery_str);
			Core::getTPL()->assign("lot_name", $row["lot_name"]);
			Core::getTPL()->assign("amount", $row["amount"]);
			Core::getTPL()->assign("amount_str", fNumber($row["amount"]));
			Core::getTPL()->assign("lot_min_amount", $row["lot_min_amount"]);
			Core::getTPL()->assign("lot_min_amount_str", fNumber($row["lot_min_amount"]));
			Core::getTPL()->assign("comis", round(($exchange_rate - EXCH_COMMISSION_BASE_UNIT) * 100));
			Core::getTPL()->assign("merchant_mark_used", Artefact::getMerchantMark());
			Core::getTPL()->assign("price", $row["price"]);
			Core::getTPL()->assign("price_str", fNumber($row["price"], 2));
			Core::getTPL()->assign("real_price", $real_price);
			Core::getTPL()->assign("real_price_str", fNumber($real_price, 2));
			Core::getTPL()->assign("premium_image", $row["featured_date"] ? Image::getImage("exch_lot_vip_big.png", Core::getLang()->getItem("PREMIUN_LOT_SLOGAN")) : "");

			$fuel_price = Exchange::lotCost(HYDROGEN, 1000);
			//t($fuel_price,'fuel_price');

			Core::getTPL()->assign("total_consumption", $total_consumption);
			Core::getTPL()->assign("fuel_price", $fuel_price["credit"] * Exchange::getFuelCostMult($row["featured_date"]));
			Core::getTPL()->assign("fuel_mult", Exchange::getFuelCostMult($row["featured_date"]));
			Core::getTPL()->assign("delivery_percent", $row["delivery_percent"]);
			Core::getTPL()->assign("delivery_hydro", $row["delivery_hydro"]);

			Core::getTPL()->assign("showBuy", true);
		}
		else
		{
			Logger::dieMessage(Core::getLanguage()->getItem("ERROROUS_PARAM"));
		}
		Core::getTPL()->display("lot_details");
	}

	protected function buyLot($lid, $buy_amount)
	{
		if(!NS::isFirstRun( "Stock::buyLot:{$lid}-{$buy_amount}-" . $_SESSION["userid"] ?? 0 ))
		{
			error_log('TOO_MANY_REQUESTS in buyLot, $lid: ' . $lid );
			Logger::dieMessage("TOO_MANY_REQUESTS");
		}
		if ( !Exchange::buyLot($lid, $buy_amount) )
		{
			Logger::dieMessage(Core::getLang()->getItem("BUY_FAILED"));
		}
		else
		{
			Logger::addFlashMessage(Core::getLang()->getItem("BUY_SECCEED"), "info");
		}
		doHeaderRedirection("game.php/Stock", false);
		return $this->index();
	}

	protected function recallByGET($id)
	{
		return $this->recall($id);
	}

	protected function recall($id)
	{
		if(NS::isFirstRun( "Stock::recall:{$lid}"))
		{
			Exchange::recallLot($id);
			$this->deleteExpireEvent($id);
		}
		doHeaderRedirection("game.php/Stock", false);
		return $this->index();
	}

	static protected function compare_by_distance($a, $b)
	{
		if ($a["real_dist"] == $b["real_dist"])
			return 0;
		return ($a["real_dist"] < $b["real_dist"]) ? -1 : 1;
	}
}
?>
