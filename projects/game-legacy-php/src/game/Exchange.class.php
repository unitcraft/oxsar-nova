<?php
/**
* Description of Exchange class
*
* Oxsar http://oxsar.ru
*
*
*/

class Exchange
{

	protected static $lot_types = array(
		0 => "ALL",
		1 => "RESOURCES",
		2 => "FLEET",
		3 => "ARTEFACT",
    );

	protected static $lot_names = array(
		METAL => "METAL",
		SILICON => "SILICON",
		HYDROGEN => "HYDROGEN"
	);
	protected static $lot_names_loaded = false;

	protected static $lot_ids = array(
		METAL => "metal",
		SILICON => "silicon",
		HYDROGEN => "hydrogen"
	);

	public static function isValidLotType($type)
	{
        return isset(self::$lot_types[$type]);
		// $keys = array_keys(self::$lot_types);
		// return in_array($type, $keys);
	}

	public static function getExchangeRate($user_id = null, $premium = false, $seller = false)
	{
		$merchant_used = Artefact::getMerchantMark($user_id);
        if(EXCH_NEW_PROFIT_TYPE){
            if($premium){
                $commission = $merchant_used ? EXCH_MERCHANT_PREMIUM_COMMISSION : EXCH_NO_MERCHANT_PREMIUM_COMMISSION;
            }else{
                $commission = $merchant_used ? EXCH_MERCHANT_COMMISSION : EXCH_NO_MERCHANT_COMMISSION;
            }
            return EXCH_COMMISSION_BASE_UNIT + $commission * ($seller ? -0.01 : 0.01);
        }
		if($premium){
			return 1 + ($merchant_used ? EXCH_MERCHANT_PREMIUM_COMMISSION : EXCH_NO_MERCHANT_PREMIUM_COMMISSION) / 100.0;
		}
		return 1 + ($merchant_used ? EXCH_MERCHANT_COMMISSION : EXCH_NO_MERCHANT_COMMISSION) / 100.0;
	}

	public static function getExchangeSellerRate($user_id = null, $premium = false)
	{
        if(EXCH_NEW_PROFIT_TYPE){
            return self::getExchangeRate($user_id, $premium, true);
        }
		return 2 - self::getExchangeRate($user_id, $premium);
	}

	public static function getFuelCostMult($premium = false)
	{
		$fuel_cost_mult = (defined('EXCHANGE_FUEL_MULT') ? EXCHANGE_FUEL_MULT : 1);
		return $premium ? $fuel_cost_mult * EXCH_PREMIUM_LOT_FUEL_COST_MULT : $fuel_cost_mult;
	}

	/**
	 *
	 * Creates Option tags with types of stock
	 * @param bool $all All options ?
	 * @param int $selected Key of selected option
	 */
	public static function getLotTypes_opt($all = false, $selected = 0, $number = false, $allow_types = null)
	{
		$lot_types = "";
		$i = (int)!$all;
		for ($i; $i < count(self::$lot_types); $i++)
		{
			if ( $number && $number != $i)
			{
				continue;
			}
		if(is_array($allow_types) && !in_array($i, $allow_types))
			{
				continue;
			}
				$lot_types .= createOption($i, Core::getLang()->getItem(self::$lot_types[$i]), $selected == $i);
		}

		return $lot_types;
	}

	public static function lotTypes_count()
	{
		return count(self::$lot_types);
	}

	public static function resolveLotNames( & $data, $add_refs = false )
	{
		//TODO: кэшировать
		if(!self::$lot_names_loaded)
		{
			$result = sqlSelect("construction", array("buildingid", "name"), "", "mode in (".sqlArray(UNIT_TYPE_FLEET, UNIT_TYPE_ARTEFACT).")");
			while($row = Core::getDatabase()->fetch($result))
			{
				self::$lot_names[$row["buildingid"]] = $row["name"];
			}
			sqlEnd($result);
			self::$lot_names_loaded = true;
		}

		foreach($data as $id => $lot)
		{
			if (is_array($lot))
			{
				if ( $lot['lot'] == ARTEFACT_PACKED_BUILDING || $lot['lot'] == ARTEFACT_PACKED_RESEARCH )
				{
					$art_data = unserialize($lot['data']);
					$temp = array_keys($art_data['ships']);
					$art_data = $art_data['ships'][$temp[1]]['art_ids'][key($art_data['ships'][$temp[1]]['art_ids'])];
					$data[$id]["lot_name"] = Core::getLanguage()->getItem(self::$lot_names[$lot["lot"]]);
					$img_link = '&typeid=' .$lot["lot"];
					if ( !empty($art_data['con_name']) && $art_data['con_name'] !== 0 )
					{
						$data[$id]["lot_name"] .= ' "';
						$data[$id]["lot_name"] .= Core::getLanguage()->getItem($art_data['con_name']);
						$img_link = 'cid='.$art_data['con_id'] . '&typeid=' .$lot["lot"];
						if ( !empty($art_data['level']) && $art_data['level'] != 0 )
						{
							$img_link = 'cid='.$art_data['con_id'] . '&level=' . $art_data['level'] . '&typeid=' .$lot["lot"];
							$data[$id]["lot_name"] .= ': ';
							$data[$id]["lot_name"] .= $art_data['level'];

							if( NS::getPlanet() )
							{
								$data[$id]["packed"] = true;
								// $data[$id]["check_req"] = Artefact::checkRequirements($id, null, null, $art_data["artid"]);
								// Artefact::checkRequirements($type_id, $user_id = null, $planet_id = null, $art_id = null)
								if( $lot['lot'] == ARTEFACT_PACKED_BUILDING )
								{
									$data[$id]["cur_level"] = NS::getPlanet()->getBuilding($art_data['con_id']);
									$data[$id]["cur_level_added"] = NS::getPlanet()->getAddedBuilding($art_data['con_id']);
								}
								else
								{
									$data[$id]["cur_level"] = NS::getResearch($art_data['con_id']);
									$data[$id]["cur_level_added"] = NS::getAddedResearch($art_data['con_id']);
								}
								$data[$id]["new_level"] = Artefact::getUpgradedLevel(array(
									'level' => $data[$id]["cur_level"],
									'added' => $data[$id]["cur_level_added"]
									), $art_data['level']);
							}
						}
						$data[$id]["lot_name"] .= '"';
					}
				}
				else
				{
					$data[$id]["lot_name"] = Core::getLang()->getItem(self::$lot_names[$lot["lot"]]);
				}
				$data[$id]["lot_name_no_url"] = $data[$id]["lot_name"];
				if($add_refs && isset($data[$id]["lot"]))
				{
					//print_r($lot);
					if( $lot['type'] == ETYPE_ARTEFACT )
					{
						//print_r($lot);
						$art_type	= $lot['lot'];
						$art_data 	= unserialize($lot['data']);
						$temp 		= array_keys($art_data['ships']);
						$art_data 	= $art_data['ships'][$temp[1]]['art_ids'][key($art_data['ships'][$temp[1]]['art_ids'])];
						$art_level 	= 0;
						if( isset($art_data['level']) && !empty($art_data['level']) )
						{
							$art_level = $art_data['level'];
						}
						$art_con = 0;
						if( isset($art_data['con_id']) && !empty($art_data['con_id']) )
						{
							$art_con = $art_data['con_id'];
						}
						//print_r($art_data);
						if( $lot['lot'] == ARTEFACT_PACKED_BUILDING || $lot['lot'] == ARTEFACT_PACKED_RESEARCH )
						{
							$link = "game.php/ArtefactInfo/".$data[$id]["lot"].'_'.$art_level.'_'.$art_con.'_1';
						}
						else
						{
							$link = "game.php/ArtefactInfo/".$data[$id]["lot"].'_'.$art_level.'_'.$art_con.'_1';
						}
						if( $lot['lot'] == ARTEFACT_PACKED_BUILDING || $lot['lot'] == ARTEFACT_PACKED_RESEARCH )
						{
							$data[$id]["image"] = Link::get(
								$link,
								Image::getImage(
									// YII_GAME_DIR.'/index.php?r=artefact2user_YII/image_new&'.$img_link,
									artImageUrl("image_new", $img_link, false),
									$data[$id]["lot_name"],
									60,
									null,
									'',
									true
								)
							);
						}
						else
						{
							$data[$id]["image"] = Link::get(
								$link,
								Image::getImage(
									getUnitImage(self::$lot_names[$lot["lot"]]),
									$data[$id]["lot_name"],
									60
								)
							);
						}
						$data[$id]["lot_name"] = Link::get($link, $data[$id]["lot_name"]);
					}
					elseif($data[$id]["lot"] > 0)
					{
						$data[$id]["image"] = Link::get("game.php/UnitInfo/".$data[$id]["lot"], Image::getImage(getUnitImage(self::$lot_names[$lot["lot"]]), $data[$id]["lot_name"], 60));
						$data[$id]["lot_name"] = Link::get("game.php/UnitInfo/".$data[$id]["lot"], $data[$id]["lot_name"]);
					}
					else if($data[$id]["lot"] == METAL)
					{
						$data[$id]["image"] = Image::getImage("met.gif", $data[$id]["lot_name"]);
					}
					else if($data[$id]["lot"] == SILICON)
					{
						$data[$id]["image"] = Image::getImage("silicon.gif", $data[$id]["lot_name"]);
					}
					else if($data[$id]["lot"] == HYDROGEN)
					{
						$data[$id]["image"] = Image::getImage("hydrogen.gif", $data[$id]["lot_name"]);
					}
				}
			}
			else if ($id == 'lot')
			{
				$data["lot_name"] = Core::getLang()->getItem(self::$lot_names[$data["lot"]]);
			}
		}

		// План 37.5d.10: legacy "return self;" — bug, в PHP 5/7 интерпретировался
		// как строка "self", в PHP 8 fatal "Undefined constant self".
		// Метод модифицирует $data по ссылке, return value не используется
		// (см. Stock/StockNew/ExchangeOpts callers).
	}

    public static function getMinLotPrice($lot, $amount, $art_id = null)
    {
        if($art_id){
            return null;
        }
        return max(EXCH_MIN_UNIT_PRICE, sqlSelectField('exchange_lots',
                'lot_unit_price', // 'lot_price / lot_amount AS min_price_calculated',
                '',
                'lot_unit_price IS NOT NULL AND lot_unit_price > 0
                    AND status='.sqlVal(ESTATUS_OK).' AND lot='.sqlVal($lot),
                'lot_unit_price',
                1)) * $amount;
    }

    /**
	* @param int $lot			���(HYDROGEN, METALL, id unit'a etc.)
	* @param int $amount	 ���������� ������
	* @return array()			(metal, silicon, hydrogen, metal_equiv, credit)
	*/
	public static function lotCost($lot, $amount, $art_id = null)
	{
		$res = array('metal' => 0,
			'silicon' => 0,
			'hydrogen' => 0,
			'metal_equiv' => 0,
			'credit' => 0);
		switch ($lot)
		{
			case METAL:
				$res['metal'] = $amount;
				break;
			case SILICON:
				$res['silicon'] = $amount;
				break;
			case HYDROGEN:
				$res['hydrogen'] = $amount;
				break;
			default:
				if ($row = sqlSelectRow("construction", array("basic_metal", "basic_silicon", "basic_hydrogen"), "", "buildingid = ".sqlVal($lot)))
				{
					$res['metal'] = $row['basic_metal'] * $amount;
					$res['silicon'] = $row['basic_silicon'] * $amount;
					$res['hydrogen'] = $row['basic_hydrogen'] * $amount;
				}
				break;
		}

		$res['metal_equiv'] = $res['metal']
			+ $res['silicon'] * MARKET_BASE_CURS_METAL / MARKET_BASE_CURS_SILICON
			+ $res['hydrogen'] * MARKET_BASE_CURS_METAL / MARKET_BASE_CURS_HYDROGEN;

		if ( !empty($art_id) )
		{
			$res["credit"] = Artefact::getArtefactCost($art_id);
		}
		else
		{
			$res["credit"] = Market::creditCost($res['metal_equiv'], 0, 0);
		}
		return $res;
	}

	/**
	 *
	 * Calculates Artefact lot cost. NOT WORKING
	 * @param int $lot Artefact_ID
	 * @param int $amount
	 */
	public static function artefactLotCost($type, $art_id)
	{
		return 1;
	}

	public static function reserveLot($data, $brokerid)
	{
		$metal = 0;
		$silicon = 0;
		$hydrogen = $data["delivery_hydro"];
		$update_fleet = false;
		switch($data["lot"])
		{
			case METAL: $metal += $data["amount"]; break;
			case SILICON: $silicon += $data["amount"]; break;
			case HYDROGEN: $hydrogen += $data["amount"]; break;
			//default:
		}
		$k = array_keys($data["ships"]);
		$unitid = $data["ships"][$k[0]]["id"];
		$art_id = 0;
		if ( isset($data["ships"][$k[1]]["art_ids"]) )
		{
			$art_id = key($data["ships"][$k[1]]["art_ids"]);
			// План 37.5d.5#7: replaced Artefact2user_YII::findByPk()
			$art_data = sqlSelectRow("artefact2user", array("planetid", "lot_id"), "", "artid=".sqlVal($art_id));
			if( !$art_data || $art_data["planetid"] == 0 || !empty($art_data["lot_id"]) )
			{
				return 'ARTEFACT';
			}
		}
		$fleet = sqlSelectRow("unit2shipyard u", array("unitid", "u.quantity - u.damaged as quantity"), "", "planetid = ".sqlPlanet() . " and unitid = ".sqlVal($unitid));
		if ($fleet["quantity"] - $data["ships"][$k[0]]["quantity"] < 0)
		{
			return 'FLEET';
			return false;
		}

		$broker = sqlSelectRow('exchange', '*', '', 'eid = '.sqlVal($brokerid));
		if ( NS::getUser()->get('credit') < $broker['comission'] )
		{
			return 'COMISSION';
			return false;
		}

		if ( NS::getUser()->get('credit') < $broker['comission'] + $data["wide_price"] + $data["premium_price"] )
		{
			return 'WIDE_PRICE';
			return false;
		}

		if($metal || $silicon || $hydrogen || $broker['comission'] || $data["wide_price"] || $data["premium_price"])
		{
			$res_log = NS::updateUserRes(array(
				'block_minus'	=> true,
				'type'			=> RES_UPDATE_EXCH_LOT_RESERVE,
				'reload_planet' => false,
				'userid'		=> NS::getUser()->get('userid'),
				'planetid'		=> NS::getUser()->get('curplanet'),
				'metal'			=> - $metal,
				'silicon'		=> - $silicon,
				'hydrogen'		=> - $hydrogen,
				'credit'		=> - $broker['comission'],
			));

			if(empty($res_log['minus_blocked']))
			{
				if($data["wide_price"])
				{
					NS::updateUserRes(array(
						// 'block_minus'	=> true,
						'type'			=> RES_UPDATE_EXCH_LOT_PRICE_EXT,
						'reload_planet' => false,
						'userid'		=> NS::getUser()->get('userid'),
						'planetid'		=> NS::getUser()->get('curplanet'),
						'credit'		=> - $data["wide_price"],
					));
				}

				if($data["premium_price"])
				{
					NS::updateUserRes(array(
						// 'block_minus'	=> true,
						'type'			=> RES_UPDATE_EXCH_LOT_PREMIUM,
						'reload_planet' => false,
						'userid'		=> NS::getUser()->get('userid'),
						'planetid'		=> NS::getUser()->get('curplanet'),
						'credit'		=> - $data["premium_price"],
					));
				}

				new AutoMsg(
					MSG_CREDIT,
					$_SESSION["userid"] ?? 0,
					time(),
					array('credits' => $broker['comission'] + $data["wide_price"] + $data["premium_price"], 'msg' => 'MSG_CREDIT_EXCHANGE_RESERVE' )
				);

				// Update By Pk
				Core::getDatabase()->query(
					"update ".PREFIX."unit2shipyard set "
						. " quantity = quantity-".sqlVal($data["ships"][$k[0]]["quantity"])
						. " where planetid=".sqlPlanet()." and unitid=".sqlVal($unitid)
				);

				if ( !empty($art_id) )
				{
					// Update By Pk
					sqlUpdate('artefact2user', array('planetid' => 0), 'artid = '. $art_id );
				}

				if(NS::getUser()->get('userid') != $broker["uid"])
				{
                    if(0 && EXCH_NEW_PROFIT_TYPE && $broker["comission"] >= 10){
                        $profit = max(0.01, round($broker["comission"] * self::getExchangeSellerRate($broker["uid"]), 2));
                    }else{
                        $profit = max(0.01, round($broker["comission"], 2));
                    }
					NS::updateUserRes(array(
						"type"			=> RES_UPDATE_EXCH_LOT_COMISSION,
						"reload_planet" => false,
						"userid" 		=> $broker["uid"],
						"credit"		=> $profit, // $broker["comission"],
					));
					new AutoMsg(
						MSG_CREDIT,
						$broker['uid'],
						time(),
						array('credits' => $profit, 'msg' => 'MSG_CREDIT_EXCHANGE_FROM_RESERVE' )
					);
				}

				return true;
			}
			return 'CANT_RESERV';
		}
		return false;
	}

	/**
	 *
	 * Restores resources, fleet and artefact, when lot is recalled or expired.
	 * @param int $id Lot ID
	 */
	public static function sendBackLot($id)
	{
		if( !NS::isFirstRun("Exchange::sendBackLot:{$id}") )
		{
			return false;
		}
		// 37.8 REF-004: атомарно «забронировать» лот переводом status=OK→RECALL.
		// Только победивший процесс делает refund. Закрывает race-окно >2 сек,
		// которое isFirstRun (TTL=2) не покрывает (бот-скрипт с sleep 2.5s).
		$stmt = sqlQuery(
			'UPDATE '.PREFIX.'exchange_lots SET status='.sqlVal(ESTATUS_RECALL).', sold_date='.sqlVal(time())
			. ' WHERE lid='.sqlVal($id).' AND status='.sqlVal(ESTATUS_OK)
		);
		if( !$stmt || $stmt->rowCount() < 1 )
		{
			return false;
		}
		// Status уже RECALL — SELECT с условием status=OK не подходит, читаем без него.
		$lot = sqlSelectRow( 'exchange_lots l', 'l.*, p.userid as sellerid',
				'INNER JOIN '.PREFIX.'planet p ON p.planetid = l.planetid',
				'l.lid = '.sqlVal($id) );
		if ( $lot )
		{
			$data 		= unserialize($lot["data"]);
			$unitKeys	= array_keys($data["ships"]);
			$unitData	= $data["ships"][$unitKeys[0]];
			$unitType	= $unitData["id"];

			$metal = $silicon = $hydrogen = 0;
			$hydrogen += $lot["delivery_hydro"];
			// If Resources
			if( $lot['type'] == EXCH_TYPE_RESOURCES )
			{
				switch ( $lot["lot"] )
				{
					case METAL:
						$metal += $lot["amount"];
						break;
					case SILICON:
						$silicon += $lot["amount"];
						break;
					case HYDROGEN:
						$hydrogen += $lot["amount"];
						break;
				}
			}
			// If Artefact
			if( $lot['type'] == EXCH_TYPE_ARTEFACT )
			{
				$artData	= reset( $data['ships'][$unitKeys[1]]['art_ids'] );
				$artKey		= $artData['artid'];
				// Update By Pk
				sqlUpdate('artefact2user', array( 'planetid' => $lot['planetid'], 'lot_id' => 0 ), "artid = ".sqlVal($artKey));
			}

			// $seller_id = sqlSelectField("planet", "userid", "", "planetid=".sqlVal($lot['planetid'])); // $lot['brokerid'],
			NS::updateUserRes(array(
				'type' 			=> RES_UPDATE_EXCH_LOT_UNLOAD,
				'reload_planet' => false,
				'ownerid'		=> $lot['lid'],
				'userid'		=> $lot['sellerid'],
				'planetid'		=> $lot['planetid'],
				'metal'			=> $metal,
				'silicon'		=> $silicon,
				'hydrogen'		=> $hydrogen,
			));
			// Fleet back
			$units		= sqlSelectRow(
				'unit2shipyard',
				'unitid, quantity, damaged, shell_percent',
				'',
				'unitid = '.sqlVal( $unitType ).' AND planetid = '.sqlVal( $lot['planetid'] )
			);
			if( $units )
			{
				// Update By Pk
				sqlQuery('UPDATE '.PREFIX.'unit2shipyard SET '
					. addQuantitySetSql( $unitData )
					. ' WHERE unitid = ' . sqlVal( $unitType ) . ' AND planetid = ' . sqlVal($lot['planetid']));
			}
			else
			{
				sqlInsert(
					'unit2shipyard',
					array(
						'unitid'		=> $unitData['id'],
						'planetid'		=> $lot['planetid'],
						'quantity'		=> $unitData['quantity'],
						'damaged'		=> $unitData['damaged'],
						'shell_percent'	=> $unitData['shell_percent'],
					)
				);
			}
			// 37.8 REF-004: status=RECALL уже установлен атомарно в начале метода.

			if( $metal > 0 || $silicon > 0 || $hydrogen > 0 )
			{
				$data = array( $id => &$lot );
				Exchange::resolveLotNames($data);

				new AutoMsg(MSG_EXCH_LOT_BACK_RESOURSES, $lot['sellerid'], time(), array(
					'metal' => $metal,
					'silicon' => $silicon,
					'hydrogen' => $hydrogen,
					'planetid' => $lot['planetid'],
					'lot' => $lot['lot_name'].' '.fNumber($lot['amount']),
					));
			}
			return true;
		}
		return false;
	}

	public static function recallLot($id, $check_time = true)
	{
		if(!NS::isFirstRun("Exchange::recallLot:{$id}"))
		{
			return false;
		}
		if ( $lot = sqlSelectRow(
			"exchange_lots l",
			"l.*",
			"",
			"lid = ".sqlVal($id)." AND status=".sqlVal(ESTATUS_OK)
		))
		{
			$now = time();
			$time_constraint = $lot["raising_date"] + EXCH_MIN_TTL * 3600 * 24;
			if ($check_time && ($now < $time_constraint) && !isAdmin(null, true))
			{
				Logger::dieMessage("CANT_RECALL_YET");
			}
			if($check_time && !isAdmin(null, true))
			{
				$cur_user_id = NS::getUser()->get("userid");
				if( !sqlSelectField("planet", "userid", "", "planetid=".sqlVal($lot["planetid"])." AND userid=".sqlVal($cur_user_id))
					&& !sqlSelectField("planet", "userid", "", "planetid=".sqlVal($lot["brokerid"])." AND userid=".sqlVal($cur_user_id)) )
				{
					Logger::dieMessage("CANT_RECALL_THIS_LOT");
				}
			}

			return Exchange::sendBackLot($id);
		}
		return false;
	}

	/**
	 * Calculates price WITHOUT fuel.
	 * @param int $buyer_userid User id.
	 * @param int $seller_userid User id.
	 * @param float $price Whole price for lot.
	 * @param float $multiplyer Amount Multiplyer.
	 * @param int $fee Owner fee(in %).
	 * @param int $ally_disc Trade union discount(in %).
	 *
	 * @return array( buyer, seller, owner, rest )
	 */
	protected static function calculatePriceWithoutFuel( $buyer_userid, $seller_userid, $price, $bought_price, $fee, $ally_disc = 0 )
	{
		$result = array(
			'buyer'		=> 0,
			'seller'	=> 0,
			'owner'		=> 0,
			'rest'		=> 0,
		);
		$result['buyer']	= $bought_price;
		$result['rest']		= $price - $result['buyer'];
		if( isTradeUnion( $seller_userid, $buyer_userid ) )
		{
			$result['buyer'] = $result['buyer'] * ( 1 - $ally_disc / 100 );
		}
        if(EXCH_NEW_PROFIT_TYPE){
            $result['seller'] = $result['buyer'];
        }else{
            $result['owner'] = $result['buyer'] * ( $fee / 100 );
            $result['seller'] = $result['buyer'] - $result['owner'];
        }
		foreach( $result as $k => $v )
		{
			$result[$k] = round($v, 2);
		}
		//t($result, 'Exchange.calculatePriceWithoutFuel');

		return $result;
	}

	protected static function generateFleetWithLot($type, $data, $amount, $multiplyer, $lot, $lid)
	{
		$points = array(
			// 'points'	=> 0,
			'u_points'	=> 0,
			'u_count'	=> 0,
		);
		$fleet	= array();
		$fleet	= $data;
		$keys	= array_keys( $fleet['ships'] );
		$two_way= true;
		$fleet['exchange'] = $type;
		// Resolve consumption
		$fleet['consumption'] 	= round($data['consumption'] * $multiplyer, 2);
		$data['consumption'] 	= $data['consumption'] - $fleet['consumption'];
		if( $type == ETYPE_FLEET )
		{
			$two_way = false;
			// Resolve Amount
			$fleet['ships'][$keys[0]]['quantity']	= $amount;
			$data['ships'][$keys[0]]['quantity']	= $data['ships'][$keys[0]]['quantity'] - $amount;
			$temp_points = PointRenewer::getUnitStats($fleet['ships']);
			$points['u_points'] = $temp_points['points'];
			$points['u_count']	= $temp_points['count'];
		}
		elseif( $type == ETYPE_ARTEFACT )
		{
			$fleet['artef_id'] = key( $fleet['ships'][ $keys[1] ]['art_ids'] );
		}
		elseif( $type == ETYPE_RESOURCE )
		{
			$fleet['metal'] = $fleet['silicon'] = $fleet['hydrogen'] = 0;
			switch( $lot )
			{
				case METAL : 	$fleet['metal']		+= $amount; 	$fleet['res_type'] = 'metal'; 	break;
				case SILICON : 	$fleet['silicon']	+= $amount; 	$fleet['res_type'] = 'silicon'; break;
				case HYDROGEN :	$fleet['hydrogen']	+= $amount; 	$fleet['res_type'] = 'hydrogen';break;
			}
			$fleet['sold'] = $multiplyer >= 1;
			$fleet['lid'] = $lid;
		}
		else
		{
			Logger::dieMessage('unknown lot type');
		}
		$result = array(
			'fleet' => $fleet,
			'data'	=> $data,
			'2way'	=> $two_way,
			'points'=> $points,
		);
		return $result;
	}

	/**
	 * Calculates Fuel Consumption and Price for rest of fuel. Also return resultng fuel in lot
	 * @param array $start Start Coords
	 * @param array $end End Coords
	 * @param int $consumption Fleet Consumption
	 * @param int $speed Fleet Speed
	 * @param boolean $two_way Either to calc 2 way cons and price , or one way
	 * @param float $got_fuel Lot fuel left
	 * @param int $percent Lot percent fuel
	 *
	 * @return array(rest_fuel, buyer_pay, one_way_consumption, two_way_consumption)
	 */
	protected static function getFuelConsumptionAndPrice( $start, $end, $consumption, $speed, $two_way = true, $got_fuel = 0, $percent = 0, $premium = false )
	{
		// t('2 way = ' . $two_way);
		$result	= array(
			'rest_fuel'				=> 0,
			'buyer_pay'				=> 0,
			'one_way_consumption'	=> 0,
			'two_way_consumption'	=> 0,
			'time'					=> 0,
		);

		$fuel_cost		= 0;
		$distance		= NS::getDistance($start['galaxy'], $start['system'], $start['position']);
//		$distance		= calculateDistance( $start, $end );
		//t($distance, 'Exchange.getFuelConsumptionAndPrice.distance');
		$one_way_cons	= NS::getFlyConsumption( $consumption, $distance ) / 2 * Exchange::getFuelCostMult($premium);
		// t($one_way_cons, 'one_way_cons');
		$two_way_cons	= $one_way_cons * 2;
		// t($two_way_cons, 'two_way_cons');

		$fly_time = NS::getFlyTime($distance, $speed, 100, $premium);

		if( $two_way )
		{
			$have_fuel = $two_way_cons * ( $percent / 100 );
			$have_fuel = min( $have_fuel, $got_fuel );
			$need_fuel = $two_way_cons - $have_fuel;
		}
		else
		{
			$have_fuel = $one_way_cons * ( $percent / 100 );
			$have_fuel = min( $have_fuel, $got_fuel );
			$need_fuel = $one_way_cons - $have_fuel;
		}
		$got_fuel = $got_fuel - $have_fuel;

		if( $need_fuel > 0 )
		{
			// t($need_fuel, 'need_fuel');
			$fuel_cost = self::lotCost(HYDROGEN, $need_fuel);
			$fuel_cost = $fuel_cost["credit"];
			// t($fuel_cost, 'fuel_cost');
		}
		$result['one_way_consumption']	= $one_way_cons;
		$result['two_way_consumption']	= $two_way_cons;
		$result['rest_fuel']			= $got_fuel;
		$result['buyer_pay']			= $fuel_cost;
		$result['time']					= $fly_time;
		$result['used_fuel']			= $have_fuel;

		foreach( $result as $k => $v )
		{
			$result[$k] = round($v, 2);
		}
		//t($result, 'Exchange.getFuelConsumptionAndPrice');
		return $result;
	}

	public static function buyLot($id, $amount = null)
	{
		if(0)
		{
			$bselect = array("g.*", "o.credit");
			$bjoin = "inner join ".PREFIX."galaxy g on g.planetid = p.planetid"
				. " inner join ".PREFIX."user o on o.userid = p.userid";
			$buyer = sqlSelectRow("planet p", $bselect, $bjoin, "p.planetid = ".sqlPlanet());
		}
		else // buying from lunes are possible
		{
			$buyer = array(
				"galaxy" => NS::getPlanet()->getData("galaxy"),
				"system" => NS::getPlanet()->getData("system"),
				"position" => NS::getPlanet()->getData("position"),
				"credit" => NS::getUser()->get("credit")
				);
		}

		$select = array("l.*",
			"g.galaxy", "g.system", "g.position",
			"p.userid as sellerid",//, "u2a.aid as seller_aid"
			"e.eid", "e.uid", "e.def_fee", "e.fee", "e.comission"
		);
		$join = "inner join ".PREFIX."galaxy g on g.planetid = l.planetid"
			. " inner join ".PREFIX."planet p on p.planetid = l.planetid"
			. " inner join ".PREFIX."exchange e on e.eid = l.brokerid";
		$where = "l.lid = ".sqlVal($id) . " and l.status=".ESTATUS_OK;

		if ( $row = sqlSelectRow("exchange_lots l", $select, $join, $where) )
		{
			$ban_exch = NS::getEH()->getExchBans();
			if ( in_array($row["brokerid"], $ban_exch) )
			{
				return false;
			}
			if ( $amount <= 0 )
			{
				// $amount = $row["amount"];
				return false;
			}

			if (
				( $amount % $row["lot_min_amount"] != 0 && $amount != $row["amount"] )
					|| $amount <= 0
			)
			{
				return false;
			}

			switch ($row['type'])
			{
				case ETYPE_FLEET:
					$ev_type = EVENT_DELIVERY_UNITS;
					if($row['lot'] <= 0)
					{
						return false;
					}
					break;

				case ETYPE_ARTEFACT:
					$ev_type = EVENT_DELIVERY_ARTEFACTS;
					if($row['lot'] <= 0)
					{
						return false;
					}
					break;

				case ETYPE_RESOURCE:
					$ev_type = EVENT_DELIVERY_RESOURSES;
					if($row['lot'] > 0)
					{
						return false;
					}
					break;

				default:
					return false;
					// Logger::dieMessage('unknown lot type');
					// break;
			}

			$min_part_price		= ( $row['price'] * $row['lot_min_amount'] ) / $row['amount'];
			$mins_bought		= $amount / $row['lot_min_amount'];
			$mins_bought_price	= $mins_bought * $min_part_price;
			$whole_mins			= $row['amount'] / $row['lot_min_amount'];
			$multiplyer			= $amount / $row['amount'];

			$data				= unserialize($row["data"]);
			$fee				= $row["fee"]; // sqlSelectField("exchange", "fee", "", "eid = ".sqlVal($row["brokerid"]));
			// Seller coords
			$data['sgalaxy']	= $row['galaxy'];
			$data['ssystem']	= $row['system'];
			$data['sposition']	= $row['position'];
			// Buyer coords
			$data['galaxy']		= $buyer['galaxy'];
			$data['system']		= $buyer['system'];
			$data['position']	= $buyer['position'];

			$org_amount = $row['amount'];
			$row['amount'] = $row['amount'] - $amount;

			$price_data = self::calculatePriceWithoutFuel(
				$_SESSION["userid"] ?? 0,
				$row['sellerid'],
				$row['price'],
				$mins_bought_price,
				$fee,
				$row['ally_discount']
			);

			$row['price']	= $price_data['rest'];
			$price 			= $price_data['buyer'];
            $exchange_profit= $price_data['owner'];
            $seller_profit	= $price_data['seller'];

			$fleet_data = self::generateFleetWithLot(
				$row['type'],
				$data,
				$amount,
				$multiplyer,
				$row['lot'],
				$row['lid']
			);

			$in_lot		= $fleet_data['fleet'];
			$fleet2send	= $fleet_data['fleet'];
			$data		= $fleet_data['data'];
			$two_way	= $fleet_data['2way'];
			$points		= $fleet_data['points'];

			$fuel_data = self::getFuelConsumptionAndPrice(
				array(
					'galaxy'	=> $data['sgalaxy'],
					'system'	=> $data['ssystem'],
					'position'	=> $data['sposition'],
				),
				array(
					'galaxy'	=> $data['galaxy'],
					'system'	=> $data['system'],
					'position'	=> $data['position'],
				),
				$in_lot['consumption'],
				$data['maxspeed'],
				$two_way,
				$row['delivery_hydro'],
				$row['delivery_percent'],
				$row["featured_date"]
			);

			$row['delivery_hydro']	= $fuel_data['rest_fuel'];
			$fuel_buyer				= $fuel_data['buyer_pay'];
			$fly_time				= $fuel_data['time'];

			$fleet2send['time'] = $fly_time;

			$buyer_exchange_rate = Exchange::getExchangeRate(null, $row["featured_date"]);
			$org_price = $price;
			$price = max( 0.01, round($price * $buyer_exchange_rate, 2) );
			$price_and_fuel = max( 0.01, round($price + $fuel_buyer, 2) );

			if ( $buyer['credit'] < $price_and_fuel )
			{
				return false;
			}

			$buyer_id = NS::getUser()->get("userid"); // $_SESSION["userid"] ?? 0;
			NS::updateUserRes(array(
				'type'			=> RES_UPDATE_EXCH_LOT_BUY,
				'reload_planet'	=> false,
				'userid'		=> $buyer_id,
				'credit'		=> - $price_and_fuel,
			));
			new AutoMsg(
				MSG_CREDIT,
				$buyer_id,
				time(),
				array('credits' => $price_and_fuel, 'msg' => 'MSG_CREDIT_EXCHANGE_BUY', 'content' => $in_lot )
			);

			$seller_exchange_rate = Exchange::getExchangeSellerRate( $row["sellerid"], $row["featured_date"] );
            $seller_profit = max( 0.01, round($seller_profit * $seller_exchange_rate, 2) );
            if(EXCH_NEW_PROFIT_TYPE){
                $exchange_profit = max( 0.01, round($seller_profit * ( $fee / 100 ), 2) );
                $seller_profit = max( 0.01, round($seller_profit - $exchange_profit, 2) );
            }

			NS::updateUserRes(array(
				'type'			=> RES_UPDATE_EXCH_LOT_SELL,
				'reload_planet' => false,
				'userid'		=> $row['sellerid'],
				'credit'		=> $seller_profit,
			));
			new AutoMsg(
				MSG_CREDIT,
				$row['sellerid'],
				time(),
				array('credits' => $seller_profit, 'msg' => 'MSG_CREDIT_EXCHANGE_SELL', 'content' => $in_lot )
			);

			NS::updateUserPoints(array(
				'user_id'	=> $buyer_id,
				'points'	=> $points,
				'inc'		=> true,
			));
			NS::updateUserPoints(array(
				'user_id'	=> $row['sellerid'],
				'points'	=> $points,
				'inc'		=> false,
			));

			Exchange::exchangeProfit( $row['uid'], $row['brokerid'], $exchange_profit, $row['def_fee'] );

			NS::getEH()->addEvent(
				$ev_type,
				$fly_time + time(),
				$row['planetid'],
				$row['sellerid'],
				NS::getUser()->get('curplanet'), $fleet2send
			);

			if ( $row['amount'] <= 0 )
			{
				// Update By Pk
				sqlUpdate(
					'exchange_lots',
					array(
						'buyerid'			=> $_SESSION["userid"] ?? 0,
						'buyerplanet'		=> $_SESSION["curplanet"] ?? 0,
						'sold_date'			=> time(),
						'data'				=> serialize($fleet2send),
						'status'			=> ESTATUS_SOLD,
						'lot_amount'		=> $row['lot_amount'],
						'lot_price'			=> $row['lot_price'],
						'payed_seller'		=> $seller_profit,
						'payed_buyer'		=> $price,
						'payed_exchange'	=> $exchange_profit,
						'payed_fuel'		=> $fuel_buyer,
						'used_fuel'			=> $fuel_data['used_fuel'],
						'rest_fuel'			=> $fuel_data['rest_fuel'],
					),
					'lid = '.sqlVal($row['lid'])
				);
//				Core::getQuery()->delete('exchange_lots', 'lid = '.sqlVal($row['lid']));
				if( $row['delivery_hydro'] > 0 )
				{
					//t('start !!! ');
					NS::updateUserRes(array(
						'type'			=> RES_UPDATE_EXCH_FUEL_REST,
						'reload_planet' => false,
						'userid'		=> $row['sellerid'],
						'hydrogen'		=> $row['delivery_hydro'],
						'planetid'		=> $row['planetid'],
					));
				}

				$owner_planet_data 	= sqlSelectRow( 'planet', '*', '', '	planetid = '.sqlVal( $row['planetid'] ) );
				$owner_id 			= $owner_planet_data['userid'];
				// Remove Lot event
				$found = false;
				$result2 = sqlSelect(
					'events',
					'eventid, data',
					'',
					'mode = ' . EVENT_EXCH_EXPIRE
						. ' AND user = ' . sqlVal( $owner_id )
						. ' AND planetid = ' . sqlVal( $row['planetid'] )
						. ' AND destination = ' . sqlVal( $row['planetid'] )
						. ' AND processed = ' . EVENT_PROCESSED_WAIT
				);
				while ( ( $row2 = sqlFetch($result2) ) && !$found )
				{
					$row2['data'] = unserialize($row2['data']);
					if(
						$row2['data']['exchid'] == $id
							|| $row2['data']['lot_id'] == $id
					)
					{
						$found = true;
						NS::getEH()->removeEvent($row2['eventid'], true);
					}
				}
				sqlEnd($result2);

				if( $row['delivery_hydro'] > 0 )
				{
					$data = array( $id => &$row );
					Exchange::resolveLotNames($data);

					new AutoMsg(MSG_EXCH_LOT_BACK_RESOURSES, $row['sellerid'], time(), array(
						'metal' => 0,
						'silicon' => 0,
						'hydrogen' => $row['delivery_hydro'],
						'planetid' => $row['planetid'],
						'lot' => $row['lot_name'].' '.fNumber($org_amount),
						));
				}
			}
			else
			{
				sqlInsert(
					'exchange_lots',
					array(
						'planetid'			=> $row['planetid'],
						'brokerid'			=> $row['brokerid'],
						'buyerid'			=> $_SESSION["userid"] ?? 0,
						'buyerplanet'		=> $_SESSION["curplanet"] ?? 0,
						'raising_date'		=> $row['raising_date'],
						'sold_date'			=> time(),
						'expiry_date'		=> $row['expiry_date'],
						'delivery_hydro'	=> $row['delivery_hydro'],
						'delivery_percent'	=> $row['delivery_percent'],
						'type'				=> $row['type'],
						'data'				=> serialize($fleet2send),
						'lot'				=> $row['lot'],
						'amount'			=> $amount,
						'lot_min_amount'	=> $row['lot_min_amount'],
						'price'				=> $org_price,
						'fee'				=> $row['fee'],
						'ally_discount'		=> $row['ally_discount'],
						'status'			=> ESTATUS_SOLD,
						'lot_amount'		=> $row['lot_amount'],
						'lot_price'			=> $row['lot_price'],
						'lot_unit_price'    => $row['lot_unit_price'],
						'lot_parent_id'		=> $row['lid'],
						'payed_seller'		=> $seller_profit,
						'payed_buyer'		=> $price,
						'payed_exchange'	=> $exchange_profit,
						'payed_fuel'		=> $fuel_buyer,
						'used_fuel'			=> $fuel_data['used_fuel'],
						'rest_fuel'			=> $fuel_data['rest_fuel'],
					)
				);
				// Update By Pk
				sqlUpdate(
					'exchange_lots',
					array(
						'price'			=> $row['price'],
						'amount'		=> $row['amount'],
						'delivery_hydro'=> $row['delivery_hydro'],
						'data'			=> serialize($data),
						'status'		=> ( ($row['type'] == ETYPE_RESOURCE) ? ESTATUS_SUSPENDED : $row['status'] ),
					),
					'lid = '.sqlVal( $row['lid'] )
				);
			}
			return true;
		}
		return false;
	}

	protected static function exchangeProfit($uid, $eid, $profit, $def_fee)
	{
		$deffenders = array();
		if($def_fee > 0)
		{
			$result = sqlSelect('events', '*', '', 'mode = '.EVENT_HOLDING.' AND destination = '.sqlVal($eid).' AND processed = 0 AND user != '.sqlVal($uid));
			while( $row = sqlFetch($result) )
			{
				$deffenders[ $row['user'] ] = $row['user'];
			}
			sqlEnd($result);

			if (count($deffenders) > 0)
			{
				$def_pie = $profit * (1 - $def_fee / 100);
				$def_profit = $def_pie / count($deffenders);
				$profit -= $def_pie;
			}
		}
		$profit = max( 0.01, round($profit * 1 /* Exchange::getExchangeSellerRate($uid) */, 2) );
		NS::updateUserRes(array(
			'type'			=> RES_UPDATE_EXCH_OWNER_PROFIT,
			'reload_planet' => false,
			'userid'		=> $uid,
			'credit'		=> $profit,
		));
		new AutoMsg(
			MSG_CREDIT,
			$uid,
			time(),
			array('credits' => $profit, 'msg' => 'MSG_CREDIT_EXCHANGE_PROFIT' )
		);

		foreach($deffenders as $deffender_id)
		{
			$cur_def_profit = max( 0.01, round($def_profit * 1 /* Exchange::getExchangeSellerRate($deffender_id) */, 2) );
			NS::updateUserRes(array(
				'type'			=> RES_UPDATE_EXCH_DEFENDER_PROFIT,
				'reload_planet' => false,
				'userid'		=> $deffender_id,
				'credit'		=> $cur_def_profit,
			));
			new AutoMsg(
				MSG_CREDIT,
				$deffender_id,
				time(),
				array('credits' => $cur_def_profit, 'msg' => 'MSG_CREDIT_EXCHANGE_DEFFEND' )
			);
		}
	}

	public static function freeSlots($eid, $total = false)
	{
		$select = array("e.uid",	"u2s.quantity");
		$join = "inner join ".PREFIX."unit2shipyard u2s on u2s.planetid = e.eid";

		$where = "e.eid = " . sqlVal($eid) . " AND u2s.unitid = ".sqlVal(UNIT_EXCH_SUPPORT_SLOT);

		if ($row = sqlSelectRow("exchange e", $select, $join, $where))
		{
			$max_lots = round( $row["quantity"] * NS::getResearch(UNIT_COMPUTER_TECH, $row["uid"]) * EXCH_MAX_LOTS_FACTOR );
			if ($total)
			{
				return $max_lots;
			}
			$count = sqlSelectRow("exchange_lots", "count(lid) as cnt", "", "brokerid = ".sqlVal($eid)." AND status = ".sqlVal(ESTATUS_OK));
			return $max_lots - $count["cnt"];
		}
		else
			return 0;
	}

	public static function freeUnitSlots()
	{
		return NS::getPlanet()->getBuilding(UNIT_EXCHANGE) * EXCH_LEVEL_SLOTS
			- getShipyardQuantity(UNIT_EXCH_SUPPORT_RANGE)
			- getShipyardQuantity(UNIT_EXCH_SUPPORT_SLOT)
			- NS::getEH()->getWorkingExchangeUnits();

	}

	public static function freeOfficeSlots($total = false)
	{
		$t = NS::getPlanet()->getBuilding(UNIT_EXCH_OFFICE) * 2;
		if (!$total)
		{
			$used = sqlSelectRow("exchange_lots", "count(*) as count", "", "planetid = ".sqlPlanet()." AND status = ".sqlVal(ESTATUS_OK));
			$t -= $used["count"];
		}
		return $t;
	}
}
?>