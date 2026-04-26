<?php
/**
* This class loads all relations of an user and his alliance.
* Afterwards we can check the relations to another user.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class Relation
{
	/**
	* The user's id.
	*
	* @var integer
	*/
	protected $userid = 0;

	/**
	* The user's alliance.
	*
	* @var integer
	*/
	protected $aid = 0;

	/**
	* Available alliance relations.
	*
	* @var Map
	*/
	protected $alliances = null;

	/**
	* Avaiable player relations (from buddylist).
	*
	* @var Map
	*/
	protected $players = null;

	/**
	* Ensures that we do not load the alliance relations twice.
	*
	* @var boolean
	*/
	protected $alliancesLoaded = false;

	/**
	* Ensures that we do not load the player relations twice.
	*
	* @var boolean
	*/
	protected $playersLoaded = false;

	/**
	* Holds CSS classes to format username or alliance.
	*
	* @var array
	*/
	protected $cssClasses = array(
		1			=> "protection",
		2			=> "confederation",
		3			=> "enemy",
		4 			=> "trade-union",
        5           => "armistice",
		"friend"	=> "friend",
		"ally"		=> "alliance",
		"self"		=> "ownPosition"
		);

	/**
	* Creates a new relation object.
	*
	* @param integer	User id
	* @param integer	Alliance id
	*
	* @return void
	*/
	public function __construct($userid, $aid = false)
	{
		$this->userid = $userid;
		$this->aid = $aid;
		$this->alliances = new Map();
		$this->players = new Map();
		return;
	}

	/**
	* Checks is the users has positive relation.
	*
	* @param integer	User id
	* @param integer	Alliance id
	*
	* @return boolean
	*/
	public function hasRelation($userid, $aid = false)
	{
		$alliance = $this->getAllyRelation($aid);
		$friend = $this->getPlayerRelation($userid);

		if((is_numeric($alliance) && $alliance != 3) || $friend || (!empty($aid) && $aid == $this->aid))
		{
			return true;
		}
		return false;
	}

	/**
	* Checks an relation to the indicated alliance.
	*
	* @param integer	Alliance id to check
	*
	* @return mixed	The relation mode or false
	*/
	public function getAllyRelation($aid = false)
	{
		if(!$aid || !$this->aid)
		{
			return false;
		}

		$this->loadAllianceRelations();

		if($this->alliances->size() > 0 && $this->alliances->exists($aid))
		{
			return $this->alliances->get($aid);
		}
		return false;
	}

	/**
	* Loads the alliance relations.
	*
	* @return Relation
	*/
	protected function loadAllianceRelations()
	{
		if($this->alliancesLoaded)
		{
			return $this;
		}

		$result = sqlSelect("ally_relationships", array("rel1", "rel2", "mode"), "", "rel1 = ".sqlVal($this->aid)." OR rel2 = ".sqlVal($this->aid));
		while($row = sqlFetch($result))
		{
			$rel = ($row["rel1"] == $this->aid) ? $row["rel2"] : $row["rel1"];
			$this->alliances->set($rel, $row["mode"]);
		}
		sqlEnd($result);
		$this->alliancesLoaded = true;
		// Hook::event("ALLIANCE_RELATIONS_LOADED", array(&$this->alliances));
		return $this;
	}

	/**
	* Checks an relation to the indicated user.
	*
	* @param integer	Userid
	*
	* @return boolean	True if a relation exists, false if not
	*/
	protected function getPlayerRelation($userid)
	{
		if(!$userid || !$this->userid)
		{
			return false;
		}
		$this->loadPlayerRelations();
		if($this->players->size() > 0 && $this->players->contains($userid))
		{
			return true;
		}
		return false;
	}

	/**
	* Loads the player relations (from buddylist).
	*
	* @return Relation
	*/
	protected function loadPlayerRelations()
	{
		if($this->playersLoaded)
		{
			return $this;
		}

		$result = sqlSelect("buddylist", array("friend1", "friend2"), "", "accepted = 1 AND (friend1 = ".sqlVal($this->userid)." OR friend2 = ".sqlVal($this->userid).")");
		while($row = sqlFetch($result))
		{
			$rel = ($row["friend1"] == $this->userid) ? $row["friend2"] : $row["friend1"];
			$this->players->push($rel);
		}
		sqlEnd($result);
		$this->playersLoaded = true;
		// Hook::event("PLAYER_RELATIONS_LOADED", array(&$this->players));
		return $this;
	}

	/**
	* Fetches the alliance relations and returns the CSS class.
	*
	* @param integer	Alliance id
	*
	* @return string	CSS class to format the name
	*/
	public function getAllyRelationClass($aid = false)
	{
		if($aid == $this->aid && $aid !== false)
		{
			return $this->cssClasses["ally"];
		}
		$relations = $this->getAllyRelation($aid);
		return (isset($this->cssClasses[$relations])) ? $this->cssClasses[$relations] : "";
	}

	/**
	* Fetches the player relations and returns the CSS class.
	*
	* @param integer	User id
	*
	* @return string	CSS class to format the name
	*/
	public function getPlayerRelationClass($userid, $aid = false)
	{
		if($this->userid == $userid)
		{
			return $this->cssClasses["self"];
		}

		if($this->getPlayerRelation($userid))
		{
			return $this->cssClasses["friend"];
		}
		if(!$aid || !$this->aid)
		{
			return "";
		}
		if($this->aid == $aid)
		{
			return $this->cssClasses["ally"];
		}
		$relations = $this->getAllyRelation($aid);
		return (isset($this->cssClasses[$relations])) ? $this->cssClasses[$relations] : "";
	}
}
?>