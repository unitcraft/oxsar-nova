<?php
/**
* Exchange module.
*
* Oxsar http://oxsar.ru
*
*
*/

class Exchange extends Page {

	protected $data;
	protected $userid;
	protected $title;
	protected $fee;
	protected $defenders_fee;

	function __construct()
	{
		parent::__construct();

		Core::getLanguage()->load("Statistics");
		Core::getLanguage()->load("info");

		Core::getTPL()->addHTMLHeaderFile("stats.js?".CLIENT_VERSION);

		$this->data = sqlSelectRow("exchange", "*", "", "eid=".sqlUser());
		Core::getTPL()->assign("exchange_title", $this->data['title']);
		Core::getTPL()->assign("exchange_fee", $this->data['fee']);
		Core::getTPL()->assign("exchange_def_fee", $this->data['def_fee']);

		$this->setPostAction("saveexchangesettings", "saveExchSettings")
			->addPostArg("saveExchSettings", "exchangetitle")
			->addPostArg("saveExchSettings", "exchangefee")
			->addPostArg("saveExchSettings", "exchangedeffee")
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

	function saveExchSettings($title, $fee, $def_fee)
	{
		//TODO: ограничение 0-100
		$fee = is_numeric($fee) ? $fee : 0;
		$def_fee = is_numeric($def_fee) ? $def_fee : 0;
		sqlUpdate(
			"exchange",
			array(
				"title" => $title,
				"fee" => $fee,
				"def_fee" => $def_fee
			),
			"`eid`=".sqlUser()
			. ' ORDER BY eid '
		);

		doHeaderRedirection("game.php/".Core::getRequest()->getGET("go"), false);
	}

	function showStatistics($date_min, $date_max, $sort_field, $sort_order, $page)
	{
		$min_const = 60*60*24*30;

		$date_min = Date::validateDate($date_min, $min_const, 'max');
		$date_max = Date::validateDate($date_max);
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

		$select = array("brokerid", "l.sold_date as date", "l.lot", "l.price", "l.amount", "l.status",
			//"e.eid", "e.fee",
			//"(l.price * e.fee * 0.01) as `lot_profit`"
			);
		// TODO: % есть в классе. Сделать без объединения.
		//$joins = " inner join ".PREFIX."exchange e on e.eid = l.brokerid";
		$joins = "";

		$where = "l.sold_date >= " . sqlVal($date_min['stamp']) . " AND l.sold_date <= " . sqlVal($date_max['stamp'])
			. " AND l.brokerid = " . sqlUser();
		$where .= " AND l.status IN (". sqlArray(array(ESTATUS_SOLD, ESTATUS_REMOVED, ESTATUS_RECALL)) . ")";

		$result_count = sqlSelectField("exchange_lots l", "count(*)", $joins, $where);
		$pages = ceil($result_count / Core::getOptions()->get("USER_PER_PAGE"));

		$start = createPaginator($pages, $page, Core::getOptions()->get("USER_PER_PAGE"));
		$max = Core::getOptions()->get("USER_PER_PAGE");

		//echo "wh: $where, sort: $sort, $start, $max <br/>";
		$result = sqlSelect("exchange_lots l", $select, $joins, $where, $sort, "$start, $max", $groupby);
		$lots = array();
		$lots_total = $lots_sold = Core::getDatabase()->num_rows($result);
		$turnover = $profit = 0;
		while($row = Core::getDatabase()->fetch($result))
		{
			$row['date'] = date("d.m.y H:i", $row["date"]);
			if ($row['status'] == ESTATUS_RECALL || $row['status'] == ESTATUS_REMOVED)
			{
				$row['lot_profit'] = 0;
				$lots_sold--;
				$row['profit_class'] = " class=\"false\"";
			}
			else
			{
				$row['lot_profit'] = $row["price"] * $this->data['fee'] * 0.01;
				$turnover += $row['price'];
				$profit += $row['lot_profit'];
			}
			$row['lot'] = Core::getLanguage()->getItem($row['lot']);
			$lots[] = $row;
			//debug_var($row, 'row');
		}

		Core::getTPL()->assign("date_last", $date_max['string']);
		Core::getTPL()->assign("date_first", $date_min['string']);
		Core::getTPL()->assign("sort_field", $sort_field);
		Core::getTPL()->assign("sort_order", $sort_order);
		Core::getTPL()->assign($sort_field, Image::getImage($sort_order.'.png', $sort_order));

		Core::getTPL()->assign("lots_total", $lots_total);
		Core::getTPL()->assign("lots_sold", $lots_sold);
		Core::getTPL()->assign("turnover", $turnover);
		Core::getTPL()->assign("profit", $profit);
		Core::getTPL()->addLoop("statistics", $lots);

		Core::getTPL()->display("exchange");
	}

}
?>
