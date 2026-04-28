<?php
/**
* Search the universe.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class Search extends Page
{
	/**
	* The current entered search item.
	*
	* @var string
	*/
	protected $searchItem = "";

	/**
	* Handles search modes.
	*
	* @return void
	*/
	public function __construct()
	{
		parent::__construct();
		$this->setPostAction("seek", "seek")
			->addPostArg("seek", "what")
			->addPostArg("seek", "where")
			->proceedRequest();
		return;
	}

	/**
	* Index action.
	*
	* @return Search
	*/
	protected function index()
	{
		return $this->seek("", null);
	}

	/**
	* Starts the search.
	*
	* @param string	The search item
	* @param string	The search mode
	*
	* @return Search
	*/
	protected function seek($what, $where)
	{
		Core::getLanguage()->load("Statistics");
		$this->searchItem = new OxsarString($what);
		$this->searchItem->trim()->prepareForSearch();
		if(!$this->searchItem->validSearchString)
		{
			$this->searchItem->set("");
		}

		Core::getTPL()->assign("what", $what);
		switch($where)
		{
		case 1:
			Core::getTPL()->assign("players", " selected=\"selected\"");
			$this->playerSearch();
			break;
		case 2:
			Core::getTPL()->assign("planets", " selected=\"selected\"");
			$this->planetSearch();
			break;
		case 3:
			Core::getTPL()->assign("allys", " selected=\"selected\"");
			$this->allianceSearch();
			break;
		default:
			Core::getTPL()->display("searchheader");
			break;
		}
		return $this;
	}

	/**
	* Displays the player search result.
	*
	* @return Search
	*/
	protected function playerSearch()
	{
		$select = array("u.userid", "u.username", "u.points", "u.last as useractivity", "u.umode", "p.planetname", "g.galaxy", "g.system", "g.position", "a.aid", "a.tag", "a.name", "b.to", "b.banid");
		$joins	= "LEFT JOIN ".PREFIX."planet p ON (p.planetid = u.hp)";
		$joins .= "LEFT JOIN ".PREFIX."galaxy g ON (g.planetid = u.hp)";
		$joins .= "LEFT JOIN ".PREFIX."user2ally u2a ON (u2a.userid = u.userid)";
		$joins .= "LEFT JOIN ".PREFIX."alliance a ON (a.aid = u2a.aid)";
		$joins .= "LEFT JOIN ".PREFIX."ban_u b ON (b.userid = u.userid)";
		$result = sqlSelect("user u", $select, $joins,
                "u.username LIKE ".sqlVal($this->searchItem->get()),
                "length(`username`) - ".(strlen($this->searchItem->get())-2)." ASC, u.username ASC", "25");

		$UserList = new UserList($result, 0, true);

		Core::getTPL()->addLoop("result", $UserList->getArray());
		Core::getTPL()->display("player_search_result");
		return $this;
	}

	/**
	* Displays the planet search result.
	*
	* @return Search
	*/
	protected function planetSearch()
	{
		$sr = array(); $i = 0;
		$select = array(
			"u.userid", "u.username", "u.points", "u.last as useractivity", "u.umode", "u.hp", "b.to",
			"p.planetid", "p.planetname", "p.ismoon",
			"g.galaxy", "g.system", "g.position",
			"a.aid", "a.tag", "a.name",
			"gm.galaxy as moongala", "gm.system as moonsys", "gm.position as moonpos"
			);
		$joins	= "LEFT JOIN ".PREFIX."user u ON (p.userid = u.userid)";
		$joins .= "LEFT JOIN ".PREFIX."galaxy g ON (g.planetid = p.planetid)";
		$joins .= "LEFT JOIN ".PREFIX."galaxy gm ON (gm.moonid = p.planetid)";
		$joins .= "LEFT JOIN ".PREFIX."user2ally u2a ON (u2a.userid = u.userid)";
		$joins .= "LEFT JOIN ".PREFIX."alliance a ON (a.aid = u2a.aid)";
		$joins .= "LEFT JOIN ".PREFIX."ban_u b ON (b.userid = u.userid)";
		$result = sqlSelect("planet p", $select, $joins,
                "p.planetname LIKE ".sqlVal($this->searchItem->get()),
                "length(p.planetname) - ".(strlen($this->searchItem->get())-2)." ASC, p.planetname ASC, u.username ASC", "25");
		while($row = sqlFetch($result))
		{
			$sr[$i] = $row;
			if($row["planetid"] == $row["hp"])
			{
				$p_addition = " (HP)";
			}
			else if($row["ismoon"])
			{
				$p_addition = " (".Core::getLanguage()->getItem("MOON").")";
			}
			else { $p_addition = ""; }
			$sr[$i]["planetname"] = $row["planetname"].$p_addition;
			$i++;
		}
		sqlEnd($result);

		$UserList = new UserList();
		$UserList->setByArray($sr);
		// Hook::event("SEARCH_RESULT_PLANET", array(&$UserList));

		Core::getTPL()->addLoop("result", $UserList->getArray());
		Core::getTPL()->display("player_search_result");
		return $this;
	}

	/**
	* Displays the alliance search result.
	*
	* @return Search
	*/
	protected function allianceSearch()
	{
        $slen = strlen($this->searchItem->get())-2;
		$sr = array();
		$select = array("SUM(u.points) as points", "COUNT(u2a.userid) as members", "a.aid", "a.tag", "a.name", "a.homepage", "a.showhomepage", "a.showmember");
		$joins	= "LEFT JOIN ".PREFIX."user2ally u2a ON (u2a.aid = a.aid)";
		$joins .= "LEFT JOIN ".PREFIX."user u ON (u.userid = u2a.userid)";
		$result = sqlSelect("alliance a", $select, $joins,
                "a.tag LIKE ".sqlVal($this->searchItem->get())." OR a.name LIKE ".sqlVal($this->searchItem->get()),
                "case when a.tag LIKE ".sqlVal($this->searchItem->get())."
                    then length(a.tag) - $slen
                    else length(a.name) - $slen end ASC, a.tag ASC", "25", "u2a.aid");
		$AllianceList = new AllianceList($result);
		// Hook::event("SEARCH_RESULT_ALLIANCE", array(&$AllianceList));
		Core::getTPL()->addLoop("result", $AllianceList->getArray());
		Core::getTPL()->display("ally_search_result");
		return $this;
	}
}
?>