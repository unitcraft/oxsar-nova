<?php
/**
* Exchange admin module.
*
* Oxsar http://oxsar.ru
*
*
*/

class ExchangeOpts extends Page {

	protected $data;
	protected $userid;
	protected $title;
	protected $fee;
	protected $defenders_fee;

	function __construct()
	{
		parent::__construct();

		if (NS::getPlanet()->getBuilding(UNIT_EXCHANGE) < 1)
		{
			Logger::dieMessage("NO_UNIT_EXCHANGE");
		}
		Core::getLanguage()->load("Statistics");
		Core::getLanguage()->load("info");

		Core::getTPL()->addHTMLHeaderFile("stats.js?".CLIENT_VERSION);

		$range_unit_count = $lots_unit_count = 0;
		$comp_tech = NS::getResearch(UNIT_COMPUTER_TECH);

		$result = sqlSelect("unit2shipyard", array("unitid", "quantity"), "", "planetid = ".sqlPlanet()
			. " AND unitid in (".sqlArray(UNIT_EXCH_SUPPORT_RANGE, UNIT_EXCH_SUPPORT_SLOT).")");
		while($row = Core::getDatabase()->fetch($result))
		{
			$param = $row["quantity"] * $comp_tech;
			if($row["unitid"] == UNIT_EXCH_SUPPORT_RANGE)
			{
				$range_unit_count += $row["quantity"];
			}
			else
			{
				$lots_unit_count += $row["quantity"];
			}
			Core::getTPL()->assign("unit_{$row["unitid"]}", $param);
		}
		Core::getDatabase()->free_result($result);

		Core::getTPL()->assign("exchange_comp_tech", $comp_tech);

		Core::getTPL()->assign("exchange_cur_slots", $range_unit_count + $lots_unit_count);
		Core::getTPL()->assign("exchange_max_slots", NS::getPlanet()->getBuilding(UNIT_EXCHANGE) * EXCH_LEVEL_SLOTS);

		Core::getTPL()->assign("exchange_max_lots", round($lots_unit_count * $comp_tech * EXCH_MAX_LOTS_FACTOR));
		Core::getTPL()->assign("exchange_radius", round($range_unit_count * $comp_tech * EXCH_RADIUS_FACTOR));

		Core::getTPL()->assign("exchange_max_lots_units", $lots_unit_count);
		Core::getTPL()->assign("exchange_radius_units", $range_unit_count);

		$oc_lots = sqlSelectField("exchange_lots", "count(*) as count", "", "brokerid = ".sqlPlanet()." AND status = ".sqlVal(ESTATUS_OK));
		Core::getTPL()->assign("exchange_oc_lots", $oc_lots);

		$this->data = sqlSelectRow("exchange", "*", "", "eid=".sqlPlanet());
		Core::getTPL()->assign("exchange_title", $this->data['title']);
		Core::getTPL()->assign("exchange_fee", $this->data['fee']);
		Core::getTPL()->assign("exchange_def_fee", $this->data['def_fee']);
		Core::getTPL()->assign("exchange_comission", $this->data['comission']);

		$this->setPostAction("saveexchangesettings", "saveExchSettings")
			->addPostArg("saveExchSettings", "exchangetitle")
			->addPostArg("saveExchSettings", "exchangefee")
			->addPostArg("saveExchSettings", "exchangedeffee")
			->addPostArg("saveExchSettings", "exchangecomission")
			->setPostAction("go", "showStatistics")
			->addPostArg("showStatistics", "date_first")
			->addPostArg("showStatistics", "date_last")
			->addPostArg("showStatistics", "sort_field")
			->addPostArg("showStatistics", "sort_order")
			->addPostArg("showStatistics", "page");
		$this->proceedRequest();
	}

	function index()
	{
		$this->showStatistics(false, false, 'date', 'desc', 1);
	}

	function saveExchSettings($title, $fee, $def_fee, $comission)
	{
		$fee = clampVal($fee, EXCH_FEE_MIN, EXCH_FEE_MAX);
		$def_fee = clampVal($def_fee, 0, 100);
		$comission = clampVal($comission, EXCH_COMMISSION_MIN, EXCH_COMMISSION_MAX);
		
		$charset = Core::getLanguage()->getOpt("charset");
		$title = Str::substring(trim($title), 0, 10, $charset);
		
		$titleExists = sqlSelectField("exchange", "count(*)", "", "eid != ".sqlPlanet()." AND title=".sqlVal($title));
		if($titleExists)
		{
			Core::getLanguage()->assign("exch_title", $title);
			Logger::addFlashMessage("EXCH_TITLE_EXISTS"); 
		}

		sqlUpdate("exchange", array(
			"fee" => $fee,
			"def_fee" => $def_fee,
			"comission" => $comission)
			+ ( $titleExists ? array() : array("title" => $title) )
			,
			"`eid`=".sqlPlanet()." ORDER BY eid");

		doHeaderRedirection("game.php/".Core::getRequest()->getGET("go"), false);
	}

	function showStatistics($date_min, $date_max, $sort_field, $sort_order, $page)
	{
		$min_const = 60*60*24*30;

		$date_min = Date::validateDate($date_min, $min_const, 'max');
		$date_max = Date::validateDate($date_max);

		$date_max['stamp'] = mktime(23,59,59, date('m', $date_max['stamp']), date('d', $date_max['stamp']), date('Y', $date_max['stamp']));
		$date_max['string'] = date("d.m.Y", $date_max['stamp']);

		Core::getTPL()->assign("date_last", $date_max['string']);
		Core::getTPL()->assign("date_first", $date_min['string']);

		//$date = $lot = $lot_price = $lot_profit = '';
		$sort_fields = array(
			"date" => "l.raising_date",
			"lot" => "l.lot",
			"lot_price" => "l.price",
			"lot_amount" => "l.amount",
			"lot_profit" => "lot_profit",
			);

		if (!isset($sort_fields[$sort_field]))
			$sort_field = 'date';

		if($sort_order != 'asc' && $sort_order != 'desc')
			$sort_order = $sort_field == 'date' ? 'desc' : 'asc';

		$sort = $sort_fields[$sort_field] . ' ' . $sort_order;

		$select = array("brokerid", "l.sold_date as date", "l.lot", "l.price", "l.amount", "l.status", "l.fee", "l.data"
			, "case when status=".sqlVal(ESTATUS_SOLD)." then l.price * l.fee * 0.01 else 0 end as lot_profit"
			);

		$select_stat = array(
			"count(*) as cnt",
			"sum(case when status=".sqlVal(ESTATUS_SOLD)." then 1 else 0 end) as sold",
			"sum(case when status=".sqlVal(ESTATUS_SOLD)." then l.price else 0 end) as price",
			"sum(case when status=".sqlVal(ESTATUS_SOLD)." then l.price * l.fee * 0.01 else 0 end) as lot_profit",
			);

		// TODO: % есть в классе. Сделать без объединения.
		//$joins = " inner join ".PREFIX."exchange e on e.eid = l.brokerid";
		$joins = "";

		$where = "l.sold_date >= " . sqlVal($date_min['stamp']) . " AND l.sold_date <= " . sqlVal($date_max['stamp'])
			. " AND l.brokerid = " . sqlPlanet();
		$where_stat = $where." AND l.status IN (". sqlArray(ESTATUS_SOLD, ESTATUS_REMOVED, ESTATUS_RECALL) . ")";
		// $where .= " AND l.status IN (". sqlArray(ESTATUS_SOLD /*, ESTATUS_REMOVED, ESTATUS_RECALL*/) . ")";
		$where .= " AND l.status=".sqlVal(ESTATUS_SOLD);
		// $where_stat = $where;

		$stat_row = sqlSelectRow("exchange_lots l", $select_stat, $joins, $where_stat);
		$pages = ceil($stat_row["sold"] / Core::getOptions()->get("USER_PER_PAGE"));

		$start = createPaginator($pages, $page, Core::getOptions()->get("USER_PER_PAGE"));
		$max = Core::getOptions()->get("USER_PER_PAGE");
		//die($where);
		//echo "wh: $where, sort: $sort, $start, $max <br/>";
		$groupby = "";
		$result = sqlSelect("exchange_lots l", $select, $joins, $where, $sort, "$start, $max", $groupby);
		$lots = array();
		$lots_total = $stat_row["cnt"];;
		$lots_sold = $stat_row["sold"];
		$turnover = $stat_row["price"];
		$profit = $stat_row["lot_profit"];
		while($row = Core::getDatabase()->fetch($result))
		{
			$row['date'] = date("d.m.y H:i", $row["date"]);
			if ($row['status'] == ESTATUS_RECALL || $row['status'] == ESTATUS_REMOVED)
			{
				// $row['lot_profit'] = 0;
				$lots_sold--;
				$row['profit_class'] = " class=\"false\"";
			}
			else
			{
				// $row['lot_profit'] = $row["price"] * $row["fee"] * 0.01;
				// $turnover += $row['price'];
				// $profit += $row['lot_profit'];

				$row['price'] = fNumber($row['price'], 2);
				$row['lot_profit'] = fNumber($row['lot_profit'], 2);
			}
			$row['lot'] = Core::getLanguage()->getItem($row['lot']);
			$lots[] = $row;
			//debug_var($row, 'row');
		}

		Exchange::resolveLotNames($lots);

		Core::getTPL()->assign("date_last", $date_max['string']);
		Core::getTPL()->assign("date_first", $date_min['string']);
		Core::getTPL()->assign("sort_field", $sort_field);
		Core::getTPL()->assign("sort_order", $sort_order);
		Core::getTPL()->assign($sort_field, Image::getImage($sort_order.'.png', $sort_order));

		Core::getTPL()->assign("lots_total", fNumber($lots_total));
		Core::getTPL()->assign("lots_sold", fNumber($lots_sold));
		Core::getTPL()->assign("turnover", fNumber($turnover, 2));
		Core::getTPL()->assign("profit", fNumber($profit, 2));
		Core::getTPL()->addLoop("statistics", $lots);

		Core::getTPL()->display("exchange");
	}

}
?>
