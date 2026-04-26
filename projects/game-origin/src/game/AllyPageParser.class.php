<?php
/**
* Parse the alliance page.
* {{war|yes|no|yes}} - List of wars [as links|show points|show member]
* {{protection|yes|no|yes}} - List of protection contracts [as links|show points|show member]
* {{confed|yes|yes|yes}} - List of confederations [as links|show points|show member]
* {{trage|yes|no|no}} - List of trade agreements [as links|show points|show member]
* {{armistice|yes|no|no}} - List of trade agreements [as links|show points|show member]
* {{member|yes|points}} - List of current member [show points|show member|order]
* {{points}} - Total alliance points
* {{totalmember}} - Total number of member
* {{avarage}} - Point avarage
* {{no}} - Number of alliance
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class AllyPageParser
{
	/**
	* Alliance id.
	*
	* @var integer
	*/
	protected $aid = 0;

	/**
	* Alliance text to parse.
	*
	* @var string
	*/
	protected $text = "";

	/**
	* Holds alliance relationships.
	*
	* @var array
	*/
	protected $rels = array();

	/**
	* Holds list of alliance member.
	*
	* @var array
	*/
	protected $member = array();

	/**
	* Total number of member.
	*
	* @var integer
	*/
	protected $totalMember = 0;

	/**
	* Total points.
	*
	* @var integer
	*/
	protected $points = 0;

	/**
	* Alliance rank.
	*
	* @var integer
	*/
	protected $number = 0;

	/**
	* Text has background or not.
	*
	* @var boolean
	*/
	protected $hasBg = false;

	/**
	* @var boolean
	*/
	protected $relsLoaded = false;

	/**
	* @var boolean
	*/
	protected $memberLoaded = false;

	/**
	* Sets alliance id.
	*
	* @param integer
	*
	* @return void
	*/
	public function __construct($aid)
	{
		$this->aid = $aid;
		return;
	}

	/**
	* Loads a list of all relations for this alliance.
	*
	* @return AllyPageParser
	*/
	protected function loadRelations()
	{
		if($this->relsLoaded) { return; }
		$this->relsLoaded = true;

		$joins	= "LEFT JOIN ".PREFIX."alliance a ON (ar.rel1 = a.aid)";
		$joins .= "LEFT JOIN ".PREFIX."user2ally u2a ON (ar.rel1 = u2a.aid)";
		$joins .= "LEFT JOIN ".PREFIX."user u ON (u2a.userid = u.userid)";
		$select = array("ar.mode", "a.aid", "a.tag", "a.name", "COUNT(u2a.userid) AS member", "FLOOR(SUM(u.points)) AS points");
		$result = sqlSelect("ally_relationships ar", $select, $joins, "ar.rel2 = ".sqlVal($this->aid), "a.tag ASC", "", "u2a.aid");
		while($row = sqlFetch($result))
		{
			$this->rels[$row["mode"]][] = $row;
		}
		sqlEnd($result);

		$joins	= "LEFT JOIN ".PREFIX."alliance a ON (ar.rel2 = a.aid)";
		$joins .= "LEFT JOIN ".PREFIX."user2ally u2a ON (ar.rel2 = u2a.aid)";
		$joins .= "LEFT JOIN ".PREFIX."user u ON (u2a.userid = u.userid)";
		$result = sqlSelect("ally_relationships ar", $select, $joins, "ar.rel1 = ".sqlVal($this->aid), "a.tag ASC", "", "u2a.aid");
		while($row = sqlFetch($result))
		{
			$this->rels[$row["mode"]][] = $row;
		}
		sqlEnd($result);
		return $this;
	}

	/**
	* Loads list of member.
	*
	* @param string	Order by term
	*
	* @return AllyPageParser
	*/
	protected function loadMember($order = "")
	{
		if($this->memberLoaded) { return; }
		$this->memberLoaded = true;
		switch($order)
		{
		case "points":
		case "b_points":
		case "r_points":
		case "u_points":
		case "b_count":
		case "r_count":
		case "u_count":
		case "e_points":
		case "battles":
			$sort = "DESC";
			break;
		case "name":
		default:
			$order = "username";
			$sort = "ASC";
			break;
		}

		$result = sqlSelect("user2ally u2a", array("u.username", "u.points"), "LEFT JOIN ".PREFIX."user u ON (u.userid = u2a.userid)", "u2a.aid = ".sqlVal($this->aid), "u.".$order." ".$sort);
		while($row = sqlFetch($result))
		{
			$this->member[] = $row;
			$this->points += $row["points"];
		}
		$this->points = floor($this->points);
		$this->totalMember = Core::getDB()->num_rows($result);
		sqlEnd($result);
		return $this;
	}

	/**
	* Starts parsing a text.
	*
	* @param string	Text to parse
	*
	* @return string	Parsed text
	*/
	public function startParser($text)
	{
		$this->text = $text;
		$this->text = parseBBCode($this->text, false);
		$_self = $this;
		$this->text = preg_replace_callback("/\{\{(war|protection|confed|trade|armistice)\|(yes|no)\|(yes|no)\|(yes|no)\}\}/siU", function($m) use($_self){ return $_self->replaceList($m[1],$m[2],$m[3],$m[4]); }, $this->text);
			$this->text = preg_replace_callback("/\[\[(war|protection|confed|trade|armistice)\|(yes|no)\|(yes|no)\|(yes|no)\]\]/siU", function($m) use($_self){ return $_self->replaceList($m[1],$m[2],$m[3],$m[4]); }, $this->text);
		$this->text = preg_replace_callback("/\{\{member\|(yes|no)\|([^\"]+)\}\}/siU", function($m) use($_self){ return $_self->replaceMember($m[1],$m[2]); }, $this->text);
			$this->text = preg_replace_callback("/\[\[member\|(yes|no)\|([^\"]+)\]\]/siU", function($m) use($_self){ return $_self->replaceMember($m[1],$m[2]); }, $this->text);
		$this->text = preg_replace_callback("/\{\{points\}\}/siU", function($m) use($_self){ return $_self->getPoints(); }, $this->text);
			$this->text = preg_replace_callback("/\[\[points\]\]/siU", function($m) use($_self){ return $_self->getPoints(); }, $this->text);
		$this->text = preg_replace_callback("/\{\{totalmember\}\}/siU", function($m) use($_self){ return $_self->getTotalMember(); }, $this->text);
			$this->text = preg_replace_callback("/\[\[totalmember\]\]/siU", function($m) use($_self){ return $_self->getTotalMember(); }, $this->text);
		$this->text = preg_replace_callback("/\{\{avarage\}\}/siU", function($m) use($_self){ return $_self->getAvarage(); }, $this->text);
			$this->text = preg_replace_callback("/\[\[avarage\]\]/siU", function($m) use($_self){ return $_self->getAvarage(); }, $this->text);
		$this->text = preg_replace_callback("/\{\{no\}\}/siU", function($m) use($_self){ return $_self->getNumber(); }, $this->text);
			$this->text = preg_replace_callback("/\[\[no\]\]/siU", function($m) use($_self){ return $_self->getNumber(); }, $this->text);
		$this->text = preg_replace_callback("/\{\{bg\|(url|color)\|([^\"]+)\}\}/siU", function($m) use($_self){ return $_self->setBackground($m[1],$m[2]); }, $this->text);
			$this->text = preg_replace_callback("/\[\[bg\|(url|color)\|([^\"]+)\]\]/siU", function($m) use($_self){ return $_self->setBackground($m[1],$m[2]); }, $this->text);
		$this->text = preg_replace("/\{\{line\}\}/siU", '<hr />', $this->text);
			$this->text = preg_replace("/\[\[line\]\]/siU", '<hr />', $this->text);
		if($this->hasBg) { $this->text .= "</div>"; }
		// Hook::event("ALLY_PAGE_PARSER_END", array($this));
		return $this->text;
	}

	/**
	* Sets the background for a text.
	*
	* @param string	Background type (url = picture)
	* @param string	Picture url or color
	*
	* @return string	HTML string
	*/
	protected function setBackground($type, $data)
	{
		$this->hasBg = true;
		if($type == "url")
		{
			$background = "<div style=\"background-image: url(".$data.");\">";
		}
		else
		{
			$background = "<div style=\"background-color: ".$data.";\">";
		}
		// Hook::event("SET_ALLY_TEXT_BACKGROUND", array(&$background, $type, $data));
		return $background;
	}

	/**
	* Returns points.
	*
	* @return string	Points
	*/
	public function getPoints()
	{
		$this->loadMember();
		return fNumber($this->points);
	}

	/**
	* Returns total number of member.
	*
	* @return string	Member
	*/
	public function getTotalMember()
	{
		$this->loadMember();
		return fNumber($this->totalMember);
	}

	/**
	* Calculates points avarage for this alliance.
	*
	* @return string	Avarage points
	*/
	public function getAvarage()
	{
		$this->loadMember();
		return fNumber($this->points / $this->totalMember);
	}

	/**
	* Number in ranking.
	*
	* @return string
	*/
	public function getNumber()
	{
		if($this->number > 0) { return $this->number; }
		$this->loadMember();

		$joins	= "LEFT JOIN ".PREFIX."user2ally u2a ON (u2a.aid = a.aid)";
		$joins .= "LEFT JOIN ".PREFIX."user u ON (u.userid = u2a.userid)";
		$result = sqlSelect("alliance a", "a.aid", $joins, "", "", "", "u2a.aid", "HAVING SUM(u.points) >= ".sqlVal($this->points));
		$this->number = fNumber(Core::getDB()->num_rows($result));
		sqlEnd($result);
		return $this->number;
	}

	/**
	* Generates a diplomacy list of the given mode.
	*
	* @param string	List mode (confed, war, ...)
	* @param string	Make names clickable
	* @param string	Append points
	* @param string	Show number of members
	*
	* @return string	Formatted diplomacy list
	*/
	protected function replaceList($mode, $asLink, $showPoints, $showMember)
	{
		$this->loadRelations();
		$asLink = $this->term($asLink);
		$showPoints = $this->term($showPoints);
		$showMember = $this->term($showMember);
		$what = 0;

		switch($mode)
		{
		case "protection":
			$what = 1;
			break;
		case "confed":
			$what = 2;
			break;
		case "war":
			$what = 3;
			break;
		case "trade":
			$what = 4;
			break;
        case "armistice":
			$what = 5;
			break;
		}

		// Hook::event("ALLY_PAGE_PARSER_REPLACE_LIST", array($what, $this));

		if(count($this->rels[$what]) <= 0)
		{
			return Core::getLang()->getItem("NOTHING");
		}

		$out = "";
		foreach($this->rels[$what] as $rels)
		{
			if($asLink)
			{
				$out .= Link::get("game.php/AlliancePage/".$rels["aid"], $rels["tag"], $rels["name"]);
			}
			else
			{
				$out .= $rels["tag"];
			}
			if($showPoints)
			{
				$out .= " - ".fNumber($rels["points"]);
			}
			if($showMember)
			{
				$out .= " - ".fNumber($rels["member"]);
			}
			$out .= "<br />";
		}
		$out = Str::substring($out, 0, -6);
		return $out;
	}

	/**
	* Generates a list of alliance member.
	*
	* @param string	Show points of members
	* @param string	List order
	*
	* @return string	Formatted list
	*/
	protected function replaceMember($showPoints, $order)
	{
		$this->loadMember($order);
		$showPoints = $this->term($showPoints);

		$out = "";
		foreach($this->member as $member)
		{
			$out .= $member["username"];
			if($showPoints)
			{
				$out .= " - ".fNumber($member["points"]);
			}
			$out .= "<br />";
		}
		$out = Str::substring($out, 0, -6);
		return $out;
	}

	/**
	* Makes an boolean value of the given expression.
	*
	* @param string	Expression
	*
	* @return boolean
	*/
	protected function term($expr)
	{
		if($expr == "yes") { return true; }
		return false;
	}

	/**
	* Kills parser.
	*
	* @return void
	*/
	public function kill()
	{
		unset($this);
		return;
	}
}
?>