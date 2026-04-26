<?php
/**
* Shows ranking for users and alliances.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class Ranking extends Page
{
	/**
	* Avarage mode enabled?
	*
	* @var boolean
	*/
	protected $average = false;

	/**
	* Handles this page.
	*
	* @return void
	*/
	public function __construct()
	{
		parent::__construct();
		$this
			->setGetAction("id", "e_points", "epointsRanking")
			->setGetAction("id", "dm_points", "dmpointsRanking")
			->setGetAction("id", "max_points", "maxpointsRanking")
			->addGetArg("typeRanking", "id")
			->setPostAction("go", "getRanking")
			->addPostArg("getRanking", "mode")
			->addPostArg("getRanking", "type")
			->addPostArg("getRanking", "avg")
			->addPostArg("getRanking", "pos")
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
        $type = SHOW_DM_POINTS ? 'dm_points' : 'points';
		return $this->getRanking("player", $type, false, null);
	}

	protected function epointsRanking()
	{
		return $this->getRanking("player", "e_points", false, null);
	}

	protected function dmpointsRanking()
	{
		return $this->getRanking("player", "dm_points", false, null);
	}

	protected function maxpointsRanking()
	{
		return $this->getRanking("player", "max_points", false, null);
	}

	/**
	* Checks for ranking type and calls the page.
	*
	* @param integer	Ranking mode (Alliance or Player)
	* @param integer	Points type and order
	* @param boolean	Avarage mode
	* @param integer	Start position
	*
	* @return Ranking
	*/
	protected function getRanking($mode, $type, $avg, $position)
	{
		Core::getLanguage()->load("Statistics,Galaxy");
		$validTypes = array(
			"dm_points" => 2,
			"max_points" => 2,
			"points" => 2,
			"b_points" => 2,
			"r_points" => 2,
			"u_points" => 2,
			"b_count" => 0,
			"r_count" => 0,
			"u_count" => 0,
			"e_points" => 0,
			"battles" => 0,
			);
        $def_type = 'points';
        if(!SHOW_DM_POINTS){
            unset($validTypes['dm_points']);
            unset($validTypes['max_points']);
        }else{
            $def_type = 'dm_points';
            Core::getTPL()->assign('dm_points_enabled', true);
            Core::getTPL()->assign('max_points_enabled', true);
        }
		if(!isset($validTypes[$type]))
		{
			$type = $def_type;
		}
		foreach($validTypes as $field => $val)
		{
			Core::getTPL()->assign($field."_sel", "");
		}
		Core::getTPL()->assign($type."_sel", " selected=\"selected\"");

		$this->average = $avg ? true : false;
		Core::getTPL()->assign("avg_on", $this->average);

		if($mode == "alliance")
		{
			Core::getTPL()->assign("player_sel", "");
			Core::getTPL()->assign("alliance_sel", " selected=\"selected\"");
			Core::getTPL()->assign("player_old_vacation_sel", "");
			Core::getTPL()->assign("player_observer_sel", "");
			$this->allianceRanking($type, $position, $validTypes[$type]);
		}
		else
		{
            Core::getTPL()->assign("player_sel", $mode == "player" ? "" : " selected='selected'");
			Core::getTPL()->assign("alliance_sel", "");
			Core::getTPL()->assign("player_old_vacation_sel", $mode == "player_old_vacation" ? " selected='selected'" : "");
			Core::getTPL()->assign("player_observer_sel", $mode == "player_observer" ? " selected='selected'" : "");
			$this->playerRanking($type, $position, $validTypes[$type], $mode);
		}
		return $this;
	}

	/**
	* Displays player ranking table.
	*
	* @param integer	Type of ranking (Fleet, Research, Points)
	* @param integer	Position to start ranking
	*
	* @return Ranking
	*/
	protected function playerRanking($type, $pos, $maxDecimals, $mode)
	{
        switch($mode){
        default:
        case "player":
            $add_where = " (u.observer = 0 AND u.umode = 0 AND u.userid NOT IN (SELECT userid FROM ".PREFIX."ban_u))";
            break;

        case "player_old_vacation":
            $add_where = " (u.observer = 0 AND u.umode = 1 AND u.last < ".sqlVal(time() - (VACATION_DISABLE_TIME-60*60*24*3))." AND u.userid NOT IN (SELECT userid FROM ".PREFIX."ban_u))";
            break;

        case "player_observer":
            $add_where = " (u.observer = 1 AND u.userid NOT IN (SELECT userid FROM ".PREFIX."ban_u))";
            break;
        }
		$oPos = sqlSelectField("user u", "count(*)", "", $type." >= ".sqlVal(NS::getUser()->get($type))
            . " AND $add_where"
        );
		$oPos = ceil(max(1, $oPos) / Core::getOptions()->get("USER_PER_PAGE"));

		$cnt = sqlSelectField("user u", "count(*)", "", "$add_where");
		$pages = ceil($cnt / Core::getOptions()->get("USER_PER_PAGE"));

		if(!is_numeric($pos)){
			$pos = $oPos;
		}
        $pos = clampVal($pos, 1, $pages);

		$ranks = "";
		for($i = 0; $i < $pages; $i++)
		{
			$n = $i * Core::getOptions()->get("USER_PER_PAGE") + 1;
			if($i + 1 == $pos) { $s = 1; } else { $s = 0; }
			if($i + 1 == $oPos) { $c = "ownPosition"; } else { $c = ""; }
			$ranks .= createOption($i + 1, fNumber($n)." - ".fNumber($n + Core::getOptions()->get("USER_PER_PAGE") - 1), $s, $c);
		}
		Core::getTPL()->assign("rankingSel", $ranks);
		$rank = abs(($pos - 1) * Core::getOptions()->get("USER_PER_PAGE"));
		$max = Core::getOptions()->get("USER_PER_PAGE");
		$select = array("u.userid", "u.username", "u.last as useractivity", "u.umode", "g.galaxy", "g.system", "g.position", "a.aid", "a.tag", "a.name", "b.to", "b.banid");
		if($this->average)
		{
			$select[] = "(u.".$type."/(('".time()."' - u.regtime)/60/60/24)) AS points";
		}
		else
		{
			$select[] = "u.".$type." AS points";
		}
		$joins	= "LEFT JOIN ".PREFIX."galaxy g ON u.hp = g.planetid ";
		$joins .= "LEFT JOIN ".PREFIX."user2ally u2a ON u2a.userid = u.userid ";
		$joins .= "LEFT JOIN ".PREFIX."alliance a ON a.aid = u2a.aid ";
		$joins .= "LEFT JOIN ".PREFIX."ban_u b ON (b.userid = u.userid)";
		$result = sqlSelect("user u", $select, $joins, "$add_where",
            "points DESC, u.username ASC", $rank.", ".$max, "u.userid");
		// Creating the user list object to handle the output
		$UserList = new UserList($result, $rank, !$this->average && $type == "points", $maxDecimals, $type);
		// Hook::event("SHOW_RANKING_PLAYER", array(&$UserList));

        Core::getTPL()->assign("player_observer_enabled", NEW_USER_OBSERVER || $mode == 'player_observer');
		Core::getTPL()->addLoop("ranking", $UserList->getArray());
		Core::getTPL()->display("playerstats");
		return $this;
	}

	/**
	* Shows alliance ranking.
	*
	* @param integer	Type of ranking (Fleet, Research, Points)
	* @param integer	Position to start ranking
	*
	* @return Ranking
	*/
	protected function allianceRanking($type, $pos, $maxDecimals)
	{
		$joins	= "LEFT JOIN ".PREFIX."user2ally u2a ON (u2a.aid = a.aid)";
		$joins .= "LEFT JOIN ".PREFIX."user u ON (u.userid = u2a.userid)";
		$result = sqlSelect("alliance a", "SUM(u.".$type.") AS points", $joins, "a.aid > '0' AND u.observer = 0", "", "", "a.aid", "HAVING points >= (points)");
		$oPos = Core::getDB()->num_rows($result);
		$oPos = ceil(max(1, $oPos) / Core::getOptions()->get("USER_PER_PAGE"));
		sqlEnd($result);

		$cnt = sqlSelectField("alliance", "count(*)", "", "aid > '0'");
		$pages = ceil($cnt / Core::getOptions()->get("USER_PER_PAGE"));
		if(!is_numeric($pos)){
			$pos = $oPos;
		}
		$pos = clampVal($pos, 1, $pages);

		$ranks = "";
		for($i = 0; $i < $pages; $i++)
		{
			$n = $i * Core::getOptions()->get("USER_PER_PAGE") + 1;
			if($i + 1 == $pos) { $s = 1; } else { $s = 0; }
			if($i + 1 == $oPos) { $c = "ownPosition"; } else { $c = ""; }
			$ranks .= createOption($i + 1, fNumber($n)." - ".fNumber($n + Core::getOptions()->get("USER_PER_PAGE") - 1), $s, $c);
		}
		Core::getTPL()->assign("rankingSel", $ranks);

		$order = $this->average ? "average" : "points";

		$rank = abs(($pos - 1) * Core::getOptions()->get("USER_PER_PAGE"));
		$max = Core::getOptions()->get("USER_PER_PAGE");
		$select = array("a.aid", "a.name", "a.tag", "COUNT(u2a.userid) AS members", "FLOOR(SUM(u.".$type.")) AS points", "AVG(u.".$type.") AS average");
		$joins	= "LEFT JOIN ".PREFIX."user2ally u2a ON u2a.aid = a.aid ";
		$joins .= "LEFT JOIN ".PREFIX."user u ON u2a.userid = u.userid ";
		$result = sqlSelect("alliance a", $select, $joins, "a.aid > '0' AND u.observer = 0", $order." DESC, members DESC, a.tag ASC", $rank.", ".$max, "a.aid");
		$AllianceList = new AllianceList($result, $rank, $maxDecimals);
		// Hook::event("SHOW_RANKING_ALLIANCE", array(&$AllianceList));
		Core::getTPL()->addLoop("ranking", $AllianceList->getArray());
		Core::getTPL()->display("allystats");
		return $this;
	}
}
?>