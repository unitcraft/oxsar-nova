<?php
/**
* Shows battle statistics for users.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class Battlestats extends Page
{
	/**
	* Handles this page.
	*
	* @return void
	*/
	public function __construct()
	{
		parent::__construct();

		$this->setPostAction("go", "showBattles")
			->addPostArg("showBattles", "date_first")
			->addPostArg("showBattles", "date_last")
			->addPostArg("showBattles", "user_filter")
			->addPostArg("showBattles", "alliance_filter")
			->addPostArg("showBattles", "show_no_destroyed")
			->addPostArg("showBattles", "show_aliens")
			->addPostArg("showBattles", "new_moon")
			->addPostArg("showBattles", "show_drawn")
			->addPostArg("showBattles", "sort_field")
			->addPostArg("showBattles", "sort_order")
			->addPostArg("showBattles", "page")
			->addPostArg("showBattles", "moon_battle")
			;
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
		return $this->showBattles($past, $today, '', '', true, false, false, true, 'date', 'desc', 1, false);
	}

	protected function showBattles($date_min, $date_max, $user_filter, $alliance_filter, $show_no_destroyed, $show_aliens, $new_moon, $show_drawn, $sort_field, $sort_order, $page, $moon_battle)
	{
		self::getBattles($date_min, $date_max, $user_filter, $alliance_filter, $show_no_destroyed, $show_aliens, $new_moon, $show_drawn, $sort_field, $sort_order, $page, $moon_battle);

		Core::getTPL()->display("battlestats");

		return $this;
	}

	public static function getBattles($date_min, $date_max, $user_filter, $alliance_filter, $show_no_destroyed, $show_aliens, $new_moon, $show_drawn, $sort_field, $sort_order, $page, $moon_battle)
	{
		Core::getLanguage()->load("Statistics");

		$hidden_days = !DEATHMATCH && OXSAR_RELEASED && !isAdmin() ? 3 : 0;

		$max_date = time() - 60*60*24*$hidden_days;
		$min_date = mktime(0,0,0, date('m'), date('d'), date('Y')) - 60*60*24*15 - 60*60*24*$hidden_days;

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

		$date = $planet_name = $defender_lost = $attacker_lost = '';
		$sort_fields = array(
			"date"			=> "a.time",
			"planet_name"	=> "p.planetname",
			"outcome"		=> "a.result",
			"defender_lost" => "a.defender_lost_res",
			"attacker_lost" => "a.attacker_lost_res",
			"moon"			=> "a.moonchance",
			"parts"			=> ""
			);

		if (!isset($sort_fields[$sort_field]))
		{
			$sort_field = 'date';
		}
		if($sort_order != 'asc' && $sort_order != 'desc')
		{
			$sort_order = $sort_field == 'date' ? 'desc' : 'asc';
		}

		$sort = $sort_field != 'parts' ? $sort_fields[$sort_field] . ' ' . $sort_order : "";
		$user_filter = trim($user_filter);
		$alliance_filter = trim($alliance_filter);

		Core::getTPL()->assign("date_last", $date_max_str);
		Core::getTPL()->assign("date_first", $date_min_str);
		Core::getTPL()->assign("user_filter", $user_filter);
		Core::getTPL()->assign("alliance_filter", $alliance_filter);
		Core::getTPL()->assign("show_no_destroyed", (int)$show_no_destroyed);
		Core::getTPL()->assign("show_aliens", (int)$show_aliens);
		Core::getTPL()->assign("new_moon", (int)$new_moon);
		Core::getTPL()->assign("moon_battle", (int)$moon_battle);
		Core::getTPL()->assign("show_drawn", (int)$show_drawn);
		Core::getTPL()->assign("sort_field", $sort_field);
		Core::getTPL()->assign("sort_order", $sort_order);
		Core::getTPL()->assign("date_interval_comment", sprintf(Core::getLanguage()->getItem("DATE_INTERVAL_COMMENT"), $hidden_days));
		Core::getTPL()->assign($sort_field, Image::getImage($sort_order.'.png', $sort_order));

		$select = array("a.assaultid", "a.key", "a.result", "a.planetid", "a.time",
			"a.moonchance", "a.moon", "a.attacker_lost_res", "a.defender_lost_res",
			"ap.userid", "ap.mode",
			"u.username",
			"p.planetname", "p.ismoon",
			"g.galaxy", "g.system", "g.position",
			"IFNULL(g.galaxy, gm.galaxy) as galaxy",
			"IFNULL(g.system, gm.system) as system",
			"IFNULL(g.position, gm.position) as position",
		);
		$joins = " JOIN ".PREFIX."assaultparticipant ap on a.assaultid = ap.assaultid";
		$joins .= " JOIN ".PREFIX."user u on ap.userid = u.userid";
		$joins .= " LEFT JOIN ".PREFIX."planet p ON a.planetid = p.planetid";
		$joins .= " LEFT JOIN ".PREFIX."galaxy g ON g.planetid = a.planetid";
		$joins .= " LEFT JOIN ".PREFIX."galaxy gm ON gm.moonid = a.planetid";

		$where_show_no_destroyed = $show_no_destroyed ? "" : "and (a.attacker_lost_res != 0 || a.defender_lost_res != 0)" ;
		$where_show_aliens = $show_aliens ? "" : "and (a.planetid is not null)" ;
		$where_show_drawn = $show_drawn ? "" : "and a.result != 0" ;
		$where_new_moon = $new_moon ? "and a.moon = 1" : "";
		$where_moon_battle = $moon_battle ? "and p.ismoon = 1" : "";
		$where = trim("a.time >= ".sqlVal($date_min)." and a.time <= ".sqlVal($date_max)." AND a.accomplished=1 $where_show_no_destroyed $where_show_aliens $where_new_moon $where_moon_battle $where_show_drawn");
		if($user_filter)
		{
			$user_filter_str = new String($user_filter);
			$user_filter_str->trim()->prepareForSearch();
			if($user_filter_str->validSearchString)
			{
				$where .= " AND a.assaultid IN (SELECT distinct assaultid FROM ".PREFIX."assaultparticipant ap2 "
					. "join ".PREFIX."user u2 on u2.userid = ap2.userid where u2.username like ".sqlVal($user_filter_str->get()).")";
			}
		}
		if($alliance_filter)
		{
			$alliance_filter_str = new String($alliance_filter);
			$alliance_filter_str->trim()->prepareForSearch();
			if($alliance_filter_str->validSearchString)
			{
				$where .= " AND a.assaultid IN (SELECT distinct assaultid FROM ".PREFIX."assaultparticipant ap2 "
					. " where ap2.userid in ( "
					. " select distinct userid from ".PREFIX."user2ally al join ".PREFIX."alliance a on al.aid=a.aid "
					. " where a.tag like ".sqlVal($alliance_filter_str->get())." ))";
			}
		}

		$c_name = 'Battlestats.getBattles.NumRows.md5.' . md5($select . $joins . $where);
		$c_dur	= 60*5;
		$pages	= false;
		if( $pages === false )
		{
			$result_count = sqlSelectField("assault a", "count(*)", $joins, $where);
			$pages = ceil($result_count / Core::getOptions()->get("USER_PER_PAGE"));
			// cache disabled
		}
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
		/* ��������.end*/

		saveAssaultReportSID();
		$moon_post = Core::getLanguage()->getItem("BS_MOON_POST");
		$in_space = Core::getLanguage()->getItem("BS_IN_SPACE");

		$c_name = 'Battlestats.getBattles.md5.' . md5($select . $joins . $where . $sort . $start . $max);
		$c_dur	= 60*5;
		$pre_res= false;
		if( $pre_res === false )
		{
			$result		= sqlSelect("assault a", $select, $joins, $where, $sort, ' '.$start . ", " . $max);
			while($row	= sqlFetch($result))
			{
				$pre_res[] = $row;
			}
			sqlEnd($result);
			if( empty($pre_res) )
			{
				$pre_res = array();
			}
			// cache disabled
		}
		$battles	= array();
		foreach( $pre_res as $row )
		{
			$battle = array();
			if (isset($battles[$row['assaultid']]))
			{
				$row['mode'] == 1 ? $battles[$row['assaultid']]["attackers"]++ : $battles[$row['assaultid']]["defenders"]++;
				$battles[$row['assaultid']]["total_parts"]++;
				continue;
			}

			if ($row['result'] == 2)
				$outcome = "assault-victory";
			else if ($row['result'] == 1)
				$outcome = "assault-defeat";
			else
				$outcome = "assault-drawn";
			$battle["total_parts"] = 1;
			$battle["date"] = date("d.m.y H:i", $row["time"]);
			// Setting link params
			$link_name	= '';
			$link_url	= '';
			$link_class	= '';
			$link_title	= '';
			$link_attach= '';
			$after_link	= '';
			// URL
			$link_url = socialUrl(/*RELATIVE_URL.*/"AssaultReport.php?id={$row['assaultid']}&key={$row['key']}");
			/* if( defined('SN') )
			{
				$link_url .= '&' . '';
			} */
			// Name
			if($row['galaxy'])
			{
				$link_name .= $row['planetname'];
				if ( !NS::getUser()->isGuest() )
				{
					$after_link .= " ".
						Link::get(
							"game.php/".
								SID.
								"/go:Galaxy/galaxy:{$row['galaxy']}/system:{$row['system']}"
								// .( (defined('SN')) ? ('?' . ''): ('') )
								,
							"[{$row['galaxy']}:{$row['system']}:{$row['position']}]".($row["ismoon"] ? $moon_post : "")
						);
				}
				else
				{
					$after_link .= " [{$row['galaxy']}:{$row['system']}:{$row['position']}]";
				}
			}
			else
			{
				$link_name .= $in_space;
			}
			// Title
			$link_title	.= '';
			// Class
			$link_class	.= $outcome . ' assault-report';
			// Attach
			$link_attach .= "target='_blank'";
			// Generate link
			$battle["planet_name"] = Link::get(
				$link_url,
				$link_name,
				$link_title,
				$link_class,
				$link_attach
			) . $after_link;
			$battle["defender_lost"] = fNumber($row["defender_lost_res"]);
			$battle["attacker_lost"] = fNumber($row["attacker_lost_res"]);
			if ($row['moonchance'] == '0')
				$battle["moonchance"] = '-';
			else if ($row['moon'] == '1')
				$battle["moonchance"] = "<span class=\"true\">".$row['moonchance']."%</span>";
			else
				$battle["moonchance"] = $row['moonchance'].'%';
			$battle["attackers"] = $battle["defenders"] = 0;
			$row['mode'] == 1 ? $battle["attackers"]++ : $battle["defenders"]++;

			$battles[$row['assaultid']] = $battle;
		}
		if ($sort_field == 'parts')
		{
			usort($battles, 'Battlestats::compare_by_participants');
			if ($sort_order == 'desc')
				$battles = array_reverse($battles);
		}
		Core::getTPL()->addLoop("statistics", $battles);

		return $battles;
	}

	static protected function compare_by_participants($a, $b)
	{
		if ($a["total_parts"] == $b["total_parts"])
			return 0;
		return ($a["total_parts"] < $b["total_parts"]) ? -1 : 1;
	}
}
?>