<?php
/**
* Shows and manages friend list.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class Friends extends Page
{
	/**
	* Constructor: Shows buddy list.
	*
	* @return void
	*/
	public function __construct()
	{
		parent::__construct();
		$this->setPostAction("delete", "removeFromList")
			->addPostArg("removeFromList", "remove")
			->setPostAction("accept", "acceptRequest")
			->addPostArg("acceptRequest", "relid")
			->setGetAction("id", "Add", "addToBuddylist")
			->addGetArg("addToBuddylist", "User");
		$this->proceedRequest();
		return;
	}

	/**
	* Index action.
	*
	* @return Friends
	*/
	protected function index()
	{
		Core::getLanguage()->load(array("Statistics", "Buddylist"));
		$bl = array();
		$select = array(
			"b.relid", "b.friend1", "b.friend2", "b.accepted",
			"u1.username as user1", "u1.points as points1", "u1.last as lastlogin1",
			"u2.points as points2", "u2.username as user2", "u2.last as lastlogin2",
			"a1.tag as ally1", "a1.aid as allyid1",
			"a2.tag as ally2", "a2.aid as allyid2",
			"g1.galaxy as gala1", "g1.system as sys1", "g1.position as pos1",
			"g2.galaxy as gala2", "g2.system as sys2", "g2.position as pos2"
			);
		$joins	= "LEFT JOIN ".PREFIX."user u1 ON (u1.userid = b.friend1)";
		$joins .= "LEFT JOIN ".PREFIX."user u2 ON (u2.userid = b.friend2)";
		$joins .= "LEFT JOIN ".PREFIX."galaxy g1 ON (g1.planetid = u1.hp)";
		$joins .= "LEFT JOIN ".PREFIX."galaxy g2 ON (g2.planetid = u2.hp)";
		$joins .= "LEFT JOIN ".PREFIX."user2ally u2a1 ON (u2a1.userid = b.friend1)";
		$joins .= "LEFT JOIN ".PREFIX."user2ally u2a2 ON (u2a2.userid = b.friend2)";
		$joins .= "LEFT JOIN ".PREFIX."alliance a1 ON (a1.aid = u2a1.aid)";
		$joins .= "LEFT JOIN ".PREFIX."alliance a2 ON (a2.aid = u2a2.aid)";
		$result = sqlSelect("buddylist b", $select, $joins, "b.friend1 = ".sqlUser()." OR b.friend2 = ".sqlUser(), "u1.points DESC, u2.points DESC, u1.username ASC, u2.username ASC");
		while($row = sqlFetch($result))
		{
			// Hook::event("SHOW_BUDDY_FIRST", array(&$row));
			if($row["friend1"] == NS::getUser()->get("userid"))
			{
				if($row["lastlogin2"] > time() - 900)
				{
					$status = Image::getImage("on.gif", getTimeTerm(time() - $row["lastlogin2"]));
				}
				else
				{
					$status = Image::getImage("off.gif", getTimeTerm(time() - $row["lastlogin2"]));
				}
				$username = Link::get("game.php/MSG/Write/Receiver:".$row["user2"], Image::getImage("pm.gif", Core::getLanguage()->getItem("WRITE_MESSAGE")))." ".Link::get("game.php/MSG/Write/Receiver:".$row["user2"], $row["user2"]);
				$points = $row["points2"];
				$position = getCoordLink($row["gala2"], $row["sys2"], $row["pos2"]);
				$ally = Link::get("game.php/AlliancePage/".$row["allyid2"], $row["ally2"]);
			}
			else
			{
				if($row["lastlogin1"] > time() - 900)
				{
					$status = Image::getImage("on.gif", getTimeTerm(time() - $row["lastlogin1"]));
				}
				else
				{
					$status = Image::getImage("off.gif", getTimeTerm(time() - $row["lastlogin1"]));
				}
				$username = Link::get("game.php/MSG/Write/Receiver:".$row["user1"], Image::getImage("pm.gif", Core::getLanguage()->getItem("WRITE_MESSAGE")))." ".Link::get("game.php/MSG/Write/Receiver:".$row["user1"], $row["user1"]);
				$points = $row["points1"];
				$position = getCoordLink($row["gala1"], $row["sys1"], $row["pos1"]);
				$ally = Link::get("game.php/AlliancePage/".$row["allyid1"], $row["ally1"]);
			}
			$bl[$row["relid"]]["f1"] = $row["friend1"];
			$bl[$row["relid"]]["f2"] = $row["friend2"];
			$bl[$row["relid"]]["relid"] = $row["relid"];
			$bl[$row["relid"]]["username"] = $username;
			$bl[$row["relid"]]["accepted"] = $row["accepted"];
			$bl[$row["relid"]]["points"] = fNumber($points);
			$bl[$row["relid"]]["status"] = $status;
			$bl[$row["relid"]]["position"] = $position;
			$bl[$row["relid"]]["ally"] = $ally;
			// Hook::event("SHOW_BUDDY_LAST", array($row, &$bl));
		}
		sqlEnd($result);
		Core::getTPL()->addLoop("buddylist", $bl);
		Core::getTPL()->display("buddylist");
		return $this;
	}

	/**
	* Adds an user to buddylist.
	*
	* @param integer	User to add
	*
	* @return Friends
	*/
	protected function addToBuddylist($userid)
	{
		if($userid == NS::getUser()->get("userid"))
		{
			Logger::dieMessage("SELF_REQUEST");
		}
		$result_count = sqlSelectField("buddylist", "count(*)", "", "friend1 = ".sqlVal($userid)." AND friend2 = ".sqlUser()." OR friend1 = ".sqlUser()." AND friend2 = ".sqlVal($userid));
		if($result_count == 0)
		{
			// Hook::event("ADD_TO_BUDDYLIST", array($userid));
			Core::getQuery()->insert("buddylist", array("friend1", "friend2"), array(NS::getUser()->get("userid"), $userid));
			// TODO: Send pm that buddy request has been sent.
		}
		return $this->index();
	}

	/**
	* Removes an user from the buddylist.
	*
	* @param integer	User to remove
	*
	* @return Friends
	*/
	protected function removeFromList($remove)
	{
		foreach($remove as $relid)
		{
			$result_count = sqlSelectField("buddylist", "count(*)", "", "relid = ".sqlVal($relid)." AND (friend1 = ".sqlUser()." OR friend2 = ".sqlUser().")");
			if($result_count > 0)
			{
				// Hook::event("REMOVE_FROM_BUDDYLIST", array($relid));
				Core::getQuery()->delete("buddylist", "relid = ".sqlVal($relid));
				// TODO: Send pm that buddy record has been deleted
			}
		}
		return $this->index();
	}

	/**
	* Accepts a buddylist request.
	*
	* @param array		Post
	*
	* @return Friends
	*/
	protected function acceptRequest($relid)
	{
		if(!empty($relid) && is_numeric($relid))
		{
			// Hook::event("ACCEPT_BUDDYLIST_REQUEST", array($relid));
			Core::getQuery()->update(
				"buddylist",
				"accepted",
				1,
				"relid = ".sqlVal($relid)." AND friend2 = ".sqlUser()
					. ' ORDER BY relid'
			);
			// TODO: Send pm that request has been accepted.
		}
		return $this->index();
	}
}
?>