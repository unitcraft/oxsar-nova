<?php
/**
* Shows resources transfer statistics.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class ResTransferStats extends Page
{
	/**
	* Handles this page.
	*
	* @return void
	*/
	public function __construct()
	{
		parent::__construct();

		$this->setPostAction("go", "showResTransfer")
			->addPostArg("showResTransfer", "date_first")
			->addPostArg("showResTransfer", "date_last")
			->addPostArg("showResTransfer", "group_param")
			->addPostArg("showResTransfer", "uid")
			->addPostArg("showResTransfer", "sort_field")
			->addPostArg("showResTransfer", "sort_order")
			->addPostArg("showResTransfer", "page")
			->addPostArg("showResTransfer", "search_user");
		$this->proceedRequest();
	}

	/**
	* Index action.
	*
	* @return Ranking
	*/
	protected function index()
	{
		$today = time();
		$past = $today - 60*60*24*30;
		return $this->showResTransfer($past, $today, 1, false, 'date', 'desc', 1, false);
	}

	protected function showResTransfer($date_min, $date_max, $group_param, $uid, $sort_field, $sort_order, $page, $search_user)
	{
		Core::getLanguage()->load("Statistics");

		//$hidden_days = OXSAR_RELEASED && NS::getUser()->get("userid") != 2 ? 2 : 0;
		$hidden_days = 0;

		$cur_date = mktime(0,0,0, date('m'), date('d'), date('Y'));
		$min_date = $cur_date - 60*60*24*30*3;
		$max_date = $cur_date + 60*60*24 - 1 - 60*60*24*$hidden_days;

		if(preg_match("#^(\d\d)[/\.\-](\d\d)[/\.\-](\d{4})$#", trim($date_min), $regs))
		{
			$date_min = mktime(0, 0, 0, $regs[2], $regs[1], $regs[3]);
		}
		else if(!is_numeric($date_min))
		{
			$date_min = $min_date;
		}
		$date_min = max($date_min, $min_date);

		if(preg_match("#^(\d\d)[/\.\-](\d\d)[/\.\-](\d{4})$#", trim($date_max), $regs))
		{
			$date_max = mktime(23, 59, 59, $regs[2], $regs[1], $regs[3]);
		}
		else if(!is_numeric($date_max))
		{
			$date_max = $max_date;
		}
		$date_max = min($date_max, $max_date);

		$date_max_str = date("d.m.Y", $date_max);
		$date_min_str = date("d.m.Y", $date_min);

		$date = $sender_name = $reciever_name = '';
		$where_fields = array(
			1 => array("id" => "rt.userid", "name" => "ur.username"),
			2 => array("id" => "rt.senderid", "name" => "us.username"),
			);
		if (!isset($where_fields[$group_param]))
		{
			$group_param = 1;
		}
		//		$sort_fields = array(
		//			"date" => "rt.time",
		//			"sender_name" => "rt.senderid",
		//			"reciever_name" => "rt.recieverid",
		//			);
		//
		//		if (!isset($sort_fields[$sort_field]))
		//		{
		//			$sort_field = 'date';
		//		}
		//		if($sort_order != 'asc' && $sort_order != 'desc')
		//		{
		//			$sort_order = $sort_field == 'date' ? 'desc' : 'asc';
		//		}
		//
		//		$sort = $sort_fields[$sort_field] . ' ' . $sort_order;
		$sort = "resum desc";

		Core::getTPL()->assign("date_last", $date_max_str);
		Core::getTPL()->assign("date_first", $date_min_str);
		Core::getTPL()->assign("group_param", (int)$group_param);
		Core::getTPL()->assign("sort_field", $sort_field);
		Core::getTPL()->assign("sort_order", $sort_order);
		//Core::getTPL()->assign("date_interval_comment", sprintf(Core::getLanguage()->getItem("DATE_INTERVAL_COMMENT"), $hidden_days));
		Core::getTPL()->assign($sort_field, Image::getImage($sort_order.'.png', $sort_order));

		$joins .= " inner join ".PREFIX."user us on rt.senderid = us.userid";
		$joins .= " inner join ".PREFIX."user ur on rt.userid = ur.userid";
		//$joins .= " inner join ".PREFIX."galaxy g on g.planetid = a.planetid";
		//$joins .= " inner join ".PREFIX."planet p on a.planetid = p.planetid";

		$groupby = ($group_param == 1) ? "rt.userid" : "rt.senderid";
		if ($group_param == 1)
		{
			Core::getTPL()->assign("column1_name", Core::getLanguage()->getItem("RECIEVER"));
			Core::getTPL()->assign("column2_name", Core::getLanguage()->getItem("TRANSFERS_COUNT"));
		}
		else
		{
			Core::getTPL()->assign("column1_name", Core::getLanguage()->getItem("SENDER"));
			Core::getTPL()->assign("column2_name", Core::getLanguage()->getItem("TRANSFERS_COUNT"));
		}

		if (is_numeric($uid) && $uid != 0)
		{
			$groupby = "";
			Core::getTPL()->assign("column1_name", Core::getLanguage()->getItem("RECIEVER"));
			Core::getTPL()->assign("column2_name", Core::getLanguage()->getItem("SENDER"));

			Core::getTPL()->assign("detailed", "det");
		}

		$select = array("rt.tid", "rt.time", "rt.userid", "rt.senderid",
			"us.username as sender_name",
			"ur.username as reciever_name"
			);
		if ($groupby == "")
			$select = array_merge($select, array("rt.metal", "rt.silicon", "rt.hydrogen", "rt.resum"));
		else
			$select = array_merge($select, array("SUM(rt.metal) as metal", "SUM(rt.silicon) as silicon", "SUM(rt.hydrogen) as hydrogen", "SUM(rt.resum) as resum", "count(*) as trans_cnt"));
		$where = trim("rt.time >= ".sqlVal($date_min)." and rt.time <= ".sqlVal($date_max));

		if ($groupby == "")
		{
			$where .= " AND {$where_fields[$group_param]["id"]} = $uid";
		}
		else
		{
			if (is_string($search_user))
			{
				$search_sql_user = sqlVal('%'.trim($search_user).'%');
				$where .= " AND	{$where_fields[$group_param]["name"]} LIKE ".$search_sql_user;
				Core::getTPL()->assign("search_user_name", $search_user);
			}
		}
		$result_count = sqlSelectField("res_transfer rt", "count(*)", $joins, $where, $sort, "", $groupby);
		$pages = ceil($result_count / Core::getOptions()->get("USER_PER_PAGE"));
		if(!is_numeric($page))
		{
			$page = 1;
		}
		else if($page > $pages) { $page = $pages; }
		else if($page < 1) { $page = 1; }

		$pages_to_show = 7;
		$pages_range = floor($pages_to_show / 2);
		$pages_link = "";
		$pages_sel = "";
		for($i = 0; $i < $pages; $i++)
		{
			$i1 = $i+1;
			$n = $i * Core::getOptions()->get("USER_PER_PAGE") + 1;
			if($i1 == $page) { $s = 1; } else { $s = 0; }
			$pages_sel .= createOption($i + 1, $i1, $s);
			if ((abs($i1 - $page) <= $pages_range) && ($pages_to_show-- > 0))
				$pages_link .= createPageLink($i1, $s, "[$i1]");
		}
		if ($page != $pages)
		{
			Core::getTPL()->assign("link_next", createPageLink($page +1, false, Core::getLanguage()->getItem("NEXT_PAGE")));
			Core::getTPL()->assign("link_last", createPageLink($pages, false, "&gt;&gt;"));
		}
		if ($page != 1)
		{
			Core::getTPL()->assign("link_prev", createPageLink($page -1, false, Core::getLanguage()->getItem("PREV_PAGE")));
			Core::getTPL()->assign("link_first", createPageLink(1, false, "&lt;&lt;"));
		}
		Core::getTPL()->assign("page_links", $pages_link);
		Core::getTPL()->assign("pages", $pages_sel);
		$start = abs(($page - 1) * Core::getOptions()->get("USER_PER_PAGE"));
		$max = Core::getOptions()->get("USER_PER_PAGE");
		/* PAGES.end*/

		$result = sqlSelect("res_transfer rt", $select, $joins, $where, $sort, ' '.$start . ", " . $max, $groupby);

		$resTrans = array();
		$i = 0;
		while($row = sqlFetch($result) )
		{
			$trans = array();
			//debug_var($row, 'row');
			$trans = $row;
			if ($groupby != "")
			{
				$trans["col0_val"] = ++$i;
				$trans["col2_val"] = $row["trans_cnt"];
				$trans["col1_val"] = ($group_param == 1) ? $row["reciever_name"] : $row["sender_name"];
				$trans["uid"] = ($group_param == 1) ? $row["userid"] : $row["senderid"];
			}
			else
			{
				$trans["col0_val"] = date("d.m.y H:i", $row["time"]);
				$trans["col1_val"] = $row["reciever_name"];
				$trans["col2_val"] = $row["sender_name"];
				$trans["rid"] = $row["userid"];
				$trans["sid"] = $row["senderid"];
			}
			$trans["resum"] = fNumber($row["resum"]);
			$trans["metal"] = fNumber($row["metal"]);
			$trans["silicon"] = fNumber($row["silicon"]);
			$trans["hydrogen"] = fNumber($row["hydrogen"]);


			$resTrans[$row['tid']] = $trans;
		}
		//debug_var($resTrans, "resTrans");
		Core::getTPL()->addLoop("statistics", $resTrans);

		sqlEnd($result);

		Core::getTPL()->display("restransfers");

		return $this;
	}

}
?>