<?php
/**
* Found alliances, shows alliance page and manage it.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class Alliance extends Page
{
	/**
	* Displaying alliance id.
	*
	* @var integer
	*/
	protected $aid;

	/**
	* Validation pattern for alliance name and tag.
	*
	* @var string
	*/
	protected $namePattern;

	/**
	* Constructor: Handles post and get actions.
	*
	* @return void
	*/
	public function __construct()
	{
		parent::__construct();
		Core::getLanguage()->load("Alliance");

		$this->namePattern = "#^[a-z\d_\-\s\.\pL]+$#iu";
			// "/^[A-z0-9_\-\s\."
			// . 'ёйцукенгшщзхъфывапролджэячсмитьбюЁЙЦУКЕНГШЩЗХЪФЫВАПРОЛДЖЭЯЧСМИТЬБЮ'
			// . "]+$/iu";

		// Set alliance id for the current session.
		$this->aid = NS::getUser()->get("aid");
		if(Core::getRequest()->getGET("aid"))
		{
			$this->aid = Core::getRequest()->getGET("aid");
		}
		else if(Core::getRequest()->getPOST("aid"))
		{
			$this->aid = Core::getRequest()->getPOST("aid");
		}
		else if(Core::getRequest()->getGET("go") == "MemberList" || Core::getRequest()->getGET("go") == "AlliancePage")
		{
			$this->aid = Core::getRequest()->getGET("id");
		}

		// POST actions for handler.
		$this->setPostAction("abandonally", "abandonAlly");
		$this->setPostAction("leave", "leaveAlliance");
		$this->setPostAction("apply", "apply")
			->addPostArg("apply", "application");
		$this->setPostAction("enter", "writeApplicationPost")
			->addPostArg("writeApplicationPost", "aid");
		$this->setPostAction("SendContract", "applyRelationship")
			->addPostArg("applyRelationship", "status")
			->addPostArg("applyRelationship", "message")
			->addPostArg("applyRelationship", "tag");
		$this->setPostAction("referfounder", "referFounderStatus")
			->addPostArg("referFounderStatus", "userid");
		$this->setPostAction("cancel", "cancleApplication")
			->addPostArg("cancleApplication", "alliance");
		$this->setPostAction("receipt", "manageCadidates")
			->setPostAction("refuse", "manageCadidates")
			->addPostArg("manageCadidates", "receipt")
			->addPostArg("manageCadidates", "refuse")
			->addPostArg("manageCadidates", "action");

		// GET actions for handler.
		$this->setGetAction("id", "Diplomacy", "diplomacy");
		$this->setGetAction("id", "DetermineRel", "determineRelation")
			->addGetArg("determineRelation", "relid");
		$this->setGetAction("id", "AcceptRelation", "acceptRelation")
			->addGetArg("acceptRelation", "candidate");
		$this->setGetAction("id", "RefuseRelation", "refuseRelation")
			->addGetArg("refuseRelation", "candidate")
			->addGetArg("refuseRelation", "requested");
		$this->setGetAction("id", "RelApplications", "relApplications");
		$this->setGetAction("go", "AlliancePage", "allyPage");
		$this->setGetAction("go", "MemberList", "memberList");
		$this->setGetAction("id", "ShowCandidates", "candidates");
		$this->setGetAction("id", "Manage", "manageAlliance");
		$this->setGetAction("id", "RightManagement", "manageRanks")
			->addPostArg("manageRanks", null);
		$this->setGetAction("id", "Join", "allySearch")
			->addPostArg("allySearch", "searchitem")
			->addPostArg("allySearch", "search");
		$this->setGetAction("id", "GlobalMail", "globalMail")
			->addPostArg("globalMail", "send_global_message")
			->addPostArg("globalMail", "message")
			->addPostArg("globalMail", "subject")
			->addPostArg("globalMail", "receiver");
		$this->setGetAction("do", "Apply", "writeApplication");
		if(!NS::getUser()->get("aid"))
		{
			$this->setGetAction("id", "Found", "foundAllyForm");
			$this->setPostAction("found", "foundAlliance")
				->addPostArg("foundAlliance", "tag")
				->addPostArg("foundAlliance", "name");
		}

		$this->proceedRequest();
		return;
	}

	/**
	* Index action.
	*
	* @return Alliance
	*/
	protected function index()
	{
		if(NS::getUser()->get("aid"))
		{
			$this->allyPage($this->aid);
		}
		else
		{
			$found = Link::get("game.php/Alliance/Found", Core::getLanguage()->getItem("FOUND_ALLIANCE"));
			$join = Link::get("game.php/Alliance/Join", Core::getLanguage()->getItem("JOIN_ALLIANCE"));

			$apps = array();
			$result = sqlSelect("allyapplication aa", array("a.aid", "a.tag", "a.name", "aa.date", "aa.application"), "LEFT JOIN ".PREFIX."alliance a ON (a.aid = aa.aid)", "aa.userid = ".sqlUser());
			while($row = sqlFetch($result))
			{
				$apps[$row["aid"]]["tag"] = Link::get("game.php/AlliancePage/".$row["aid"], $row["tag"], $row["name"]);
				$apps[$row["aid"]]["date"] = Date::timeToString(1, $row["date"]);
				$apps[$row["aid"]]["apptext"] = parseBBCode($row["application"]);
				$apps[$row["aid"]]["aid"] = $row["aid"];
			}
			sqlEnd($result);
			// Hook::event("ALLIANCE_OVERVIEW", array(&$apps));
			Core::getTPL()->addLoop("applications", $apps);
			Core::getTPL()->assign("foundAlly", $found);
			Core::getTPL()->assign("joinAlly", $join);
			Core::getTPL()->display("ally");
		}
		return $this;
	}

	/**
	* Displays the form to found an alliance.
	*
	* @return Alliance
	*/
	protected function foundAllyForm()
	{
		if(NS::getUser()->get("points") < ALLIANCE_FOUND_USER_MIN_POINTS)
		{
			Core::getLanguage()->assign("min_points", fNumber(ALLIANCE_FOUND_USER_MIN_POINTS));
			Core::getLanguage()->assign("user_points", fNumber(NS::getUser()->get("points")));
			Logger::dieMessage("ALLIANCE_FOUND_BLOCKED_BY_USER_POINTS");
		}
		Core::getTPL()->display("foundally");
		return $this;
	}

	/**
	* Cancles an application.
	*
	* @param array		Applications to delete
	*
	* @return Alliance
	*/
	protected function cancleApplication($alliance)
	{
		foreach($alliance as $aid)
		{
			Core::getQuery()->delete("allyapplication", "userid = ".sqlUser()." AND aid = ".sqlVal($aid));
		}
		return $this->index();
	}

	/**
	* Create the posted alliance.
	*
	* @param array	Post
	*
	* @return Alliance
	*/
	protected function foundAlliance($tag, $name)
	{
		$minCharsTag = Core::getOptions()->get("MIN_CHARS_ALLY_TAG");
		$maxCharsTag = Core::getOptions()->get("MAX_CHARS_ALLY_TAG");
		$minCharsName = Core::getOptions()->get("MIN_CHARS_ALLY_NAME");
		$maxCharsName = Core::getOptions()->get("MAX_CHARS_ALLY_NAME");
		// Hook::event("FOUND_ALLIANCE", array(&$tag, &$name));
		if(Str::length($tag) >= $minCharsTag && Str::length($tag) <= $maxCharsTag && Str::length($name) >= $minCharsName && Str::length($name) <= $maxCharsName && preg_match($this->namePattern, $tag) && preg_match($this->namePattern, $name))
		{
			$result_count = sqlSelectField("alliance", "count(*)", "", "tag = ".sqlVal($name)." OR name = ".sqlVal($name));
			if($result_count == 0)
			{
				Core::getQuery()->insert("alliance", array("tag", "name", "founder", "open"), array($tag, $name, NS::getUser()->get("userid"), 1));
				$aid = Core::getDB()->insert_id();
				Core::getQuery()->insert("user2ally", array("userid", "aid", "joindate", "rank"), array(NS::getUser()->get("userid"), $aid, time(), Core::getLanguage()->getItem("FOUNDER")));
				NS::getUser()->rebuild();
				// Hook::event("ALLIANCE_FOUNDED", array($tag, $name, $aid));
				doHeaderRedirection("game.php/Alliance", false);
			}
			else
			{
				Logger::addMessage("ALLIANCE_ALREADY_EXISTS");
			}
		}
		else
		{
			if(Str::length($tag) < $minCharsTag || Str::length($tag) > $maxCharsTag || !preg_match($this->namePattern, $tag))
			{
				Core::getTPL()->assign("tagError", Logger::getMessageField("ALLIANCE_TAG_INVALID"));
			}

			if(Str::length($name) < $minCharsName || Str::length($name) > $maxCharsName || !preg_match($this->namePattern, $name))
			{
				Core::getTPL()->assign("nameError", Logger::getMessageField("ALLIANCE_NAME_INVALID"));
			}
		}
		return $this;
	}

	/**
	* Shows the alliance page of given id.
	*
	* @return Alliance
	*/
	protected function allyPage()
	{
		$select = array("u2a.userid", "u2a.joindate", "a.aid", "a.name", "a.tag", "a.textextern", "a.textintern", "a.founder", "a.foundername", "a.logo", "a.homepage", "a.showmember", "a.showhomepage", "count(aa.userid) as applications");
		$join	= "LEFT JOIN ".PREFIX."alliance a ON (u2a.aid = a.aid)";
		$join .= " LEFT JOIN ".PREFIX."allyapplication aa ON (u2a.aid = aa.aid)";
		$result = sqlSelect("user2ally u2a", $select, $join, "u2a.aid = ".sqlVal($this->aid), "", "", "u2a.userid");
		$totalmember = Core::getDB()->num_rows($result);
		$row = sqlFetch($result);
		sqlEnd($result);
		// Hook::event("SHOW_ALLIANCE_PAGE", array(&$row));

		if(NS::getUser()->get("aid") == $row["aid"] && $row["founder"] != NS::getUser()->get("userid"))
		{
			// Load rights
			$rights = sqlSelectRow("allyrank ar", array("ar.name", "ar.CAN_SEE_MEMBERLIST", "ar.CAN_SEE_APPLICATIONS", "ar.CAN_MANAGE", "ar.CAN_WRITE_GLOBAL_MAILS"), "LEFT JOIN ".PREFIX."user2ally u2a ON u2a.rank = ar.rankid", "u2a.userid = ".sqlUser());
			Core::getTPL()->assign("rank", ($rights["name"] != "") ? $rights["name"] : Core::getLanguage()->getItem("NEWBIE"));
			Core::getTPL()->assign("CAN_SEE_MEMBERLIST", $rights["CAN_SEE_MEMBERLIST"]);
			Core::getTPL()->assign("CAN_SEE_APPLICATIONS", $rights["CAN_SEE_APPLICATIONS"]);
			Core::getTPL()->assign("CAN_MANAGE", $rights["CAN_MANAGE"]);
			Core::getTPL()->assign("CAN_WRITE_GLOBAL_MAILS", $rights["CAN_WRITE_GLOBAL_MAILS"]);
			if($rights["CAN_MANAGE"]) { $manage = "(".Link::get("game.php/Alliance/Manage", Core::getLanguage()->getItem("MANAGEMENT")).")"; }
			else { $manage = ""; }
		}
		else if($row["founder"] == NS::getUser()->get("userid"))
		{
			Core::getTPL()->assign("rank", ($row["foundername"] != "") ? $row["foundername"] : Core::getLanguage()->getItem("FOUNDER"));
			$manage = "(".Link::get("game.php/Alliance/Manage", Core::getLanguage()->getItem("MANAGEMENT")).")";
		}
		else
		{
			$result_count = sqlSelectField("allyapplication", "count(*)", "", "userid = ".sqlUser()." AND aid = ".sqlVal($row["aid"]));
			Core::getTPL()->assign("appInProgress", $result_count);
			$manage = "";
		}

		$parser = new AllyPageParser($this->aid);
		$row["textextern"] = ($row["textextern"] != "") ? $parser->startParser($row["textextern"]) : Core::getLanguage()->getItem("WELCOME");
		$row["textintern"] = ($row["textintern"] != "") ? $parser->startParser($row["textintern"]) : "";
		$parser->kill();

		Core::getTPL()->assign("appnumber", $row["applications"]);
		Core::getTPL()->assign("applications", Link::get("game.php/Alliance/ShowCandidates", sprintf(Core::getLanguage()->getItem("CANDIDATES"), fNumber($row["applications"]))));
		Core::getTPL()->assign("founder", $row["founder"]);
		Core::getTPL()->assign("manage", $manage);
		Core::getTPL()->assign("tag", $row["tag"]);
		Core::getTPL()->assign("name", $row["name"]);
		Core::getTPL()->assign("aid", $row["aid"]);
		Core::getTPL()->assign("logo", ($row["logo"] != "") ? Image::getImage($row["logo"], "") : "");

		Core::getTPL()->assign("textextern", $row["textextern"]);
		Core::getTPL()->assign("homepage", $row["homepage"] ? bbcode_text("[url]".$row["homepage"]."[/url]") : ""); // ($row["homepage"] != "") ? Link::get($row["homepage"], $row["homepage"]) : "");
		Core::getTPL()->assign("showHomepage", $row["showhomepage"]);
		Core::getTPL()->assign("textintern", $row["textintern"]);
		Core::getTPL()->assign("memberNumber", fNumber($totalmember));
		Core::getTPL()->assign("memberList", Link::get("game.php/MemberList/".$row["aid"], Core::getLanguage()->getItem("MEMBER_LIST")));
		Core::getTPL()->assign("showMember", $row["showmember"]);
		unset($row);
		Core::getTPL()->display("allypage_own");
		exit();
		return $this;
	}

	/**
	* Loads the rights and check for access.
	*
	* @param array Rights to check
	*
	* @return boolean
	*/
	protected function getRights($rights)
	{
		$select = array("a.aid", "a.founder", "ar.CAN_MANAGE", "ar.CAN_SEE_MEMBERLIST", "ar.CAN_SEE_APPLICATIONS", "ar.CAN_BAN_MEMBER", "ar.CAN_SEE_ONLINE_STATE", "ar.CAN_WRITE_GLOBAL_MAILS");
		$joins	= "LEFT JOIN ".PREFIX."alliance a ON (a.aid = u2a.aid)";
		$joins .= "LEFT JOIN ".PREFIX."allyrank ar ON (ar.rankid = u2a.rank)";
		$result = sqlSelect("user2ally u2a", $select, $joins, "u2a.userid = ".sqlUser());
		if($_row = sqlFetch($result))
		{
			sqlEnd($result);
			// Hook::event("GET_ALLIANCE_RIGHTS", array($rights, &$_row));
			if($_row["founder"] == NS::getUser()->get("userid"))
			{
				return true;
			}
			if(!is_array($rights)) { $rights = Arr::trim(explode(",", $rights)); }
			foreach($rights as $right)
			{
				if($_row[$right] == 0) { return false; }
			}

			return true;
		}
		else { sqlEnd($result); }
		return false;
	}

	/**
	* Displays applications for relationships.
	*
	* @return Alliance
	*/
	protected function relApplications()
	{
		if($this->getRights(array("CAN_MANAGE")))
		{
			$apps = array(); $i = 0;
			$joins	= "LEFT JOIN ".PREFIX."alliance a1 ON (a1.aid = ara.candidate_ally)";
			$joins .= "LEFT JOIN ".PREFIX."alliance a2 ON (a2.aid = ara.request_ally)";
			$select = array("ara.candidate_ally", "ara.request_ally", "ara.mode", "ara.application", "ara.time", "a1.tag AS tag1", "a1.name AS name1", "a2.tag AS tag2", "a2.name AS name2");
			$result = sqlSelect("ally_relationships_application ara", $select, $joins, "ara.candidate_ally = ".sqlVal($this->aid)." OR ara.request_ally = ".sqlVal($this->aid));
			while($row = sqlFetch($result))
			{
				$apps[$i]["time"] = Date::timeToString(1, $row["time"]);
				$apps[$i]["application"] = $row["application"];
				$apps[$i]["status"] = $this->getDiploStatus($row["mode"]);
				if($row["candidate_ally"] == $this->aid)
				{
					$apps[$i]["ally"] = Link::get("game.php/AlliancePage/".$row["request_ally"], $row["tag2"], $row["name2"]);
					$apps[$i]["refuse"] = Link::get("game.php/Alliance/id:RefuseRelation/requested:".$row["request_ally"], Core::getLang()->getItem("REFUSE"));
				}
				else
				{
					$apps[$i]["ally"] = Link::get("game.php/AlliancePage/".$row["candidate_ally"], $row["tag1"], $row["name1"]);
					$apps[$i]["accept"] = Link::get("game.php/Alliance/id:AcceptRelation/candidate:".$row["candidate_ally"], Core::getLang()->getItem("ACCEPT"));
					$apps[$i]["refuse"] = Link::get("game.php/Alliance/id:RefuseRelation/candidate:".$row["candidate_ally"], Core::getLang()->getItem("REFUSE"));
				}
				$i++;
			}
			sqlEnd($result);
			// Hook::event("SHOW_REL_APPLICATIONS", array(&$apps));
			Core::getTPL()->addLoop("apps", $apps);
			Core::getTPL()->display("relation_applications");
			exit();
		}
		return $this;
	}

	/**
	* Executes a diplomacy application.
	*
	* @param integer	Relation status
	* @param string	Message
	* @param string	Receiver alliance (tag)
	*
	* @return Alliance
	*/
	protected function applyRelationship($status, $message, $tag)
	{
		if($this->getRights(array("CAN_MANAGE")))
		{
			if($status == 0 && Str::length($message) <= Core::getOptions()->get("MAX_PM_LENGTH") && Str::length($message) > 0)
			{
                NS::checkSpan($message, 'Ally.applyRelationship');

				$result = sqlSelect("user2ally u2a", array("u2a.userid", "u2a.aid"), "LEFT JOIN ".PREFIX."alliance a ON (a.aid = u2a.aid)", "a.tag = ".sqlVal($tag));
				while($row = sqlFetch($result))
				{
					if($row["aid"] == $this->aid) { return; } // You cannot send a message to your own alliance
					Core::getQuery()->insert("message", array("mode", "time", "sender", "receiver", "message", "subject", "readed"), array(6, time(), NS::getUser()->get("userid"), $row["userid"], Str::validateXHTML($message), Core::getLang()->getItem("MESSAGE_BY_ALLIANCE"), 0));
				}
				if(Core::getDB()->num_rows($result) == 0)
				{
					Logger::addMessage("NO_MATCHES_ALLIANCE");
				}
				sqlEnd($result);
				$this->diplomacy();
			}
			else
			{
				$row = sqlSelectRow("alliance a", array("a.aid"), "", "a.tag = ".sqlVal($tag));
				if($row)
				{
					if($row["aid"] == $this->aid) { return; }
					// Check for existing relations and applications
					$where	= "(ar.rel1 = ".sqlVal($row["aid"])." AND ar.rel2 = ".sqlVal($this->aid).") OR ";
					$where .= "(ar.rel1 = ".sqlVal($this->aid)." AND ar.rel2 = ".sqlVal($row["aid"]).") OR ";
					$where .= "(ara.candidate_ally = ".sqlVal($this->aid)." AND ara.request_ally = ".sqlVal($row["aid"]).") OR ";
					$where .= "(ara.candidate_ally = ".sqlVal($row["aid"])." AND ara.request_ally = ".sqlVal($this->aid).")";
					$result_count = sqlSelectField("ally_relationships_application ara, ".PREFIX."ally_relationships ar", "count(*)", "", $where);
					if($status == 3 && $result_count <= 0)
					{
						Core::getQuery()->insert("ally_relationships", array("rel1", "rel2", "time", "mode"), array($this->aid, $row["aid"], time(), $status));
						$this->relApplications();
					}
					else if($result_count <= 0 && Str::length($message) > 0 && $status <= 5 && Str::length($message) <= Core::getOptions()->get("MAX_PM_LENGTH"))
					{
						$atts = array("candidate_ally", "request_ally", "userid", "mode", "application", "time");
						$vals = array($this->aid, $row["aid"], NS::getUser()->get("userid"), $status, Str::validateXHTML($message), time());
						Core::getQuery()->insert("ally_relationships_application", $atts, $vals);
						$this->relApplications();
					}
				}
				else
				{
					Logger::addMessage("NO_MATCHES_ALLIANCE");
				}
			}
		}
		Logger::addMessage("ERROR_WITH_DOPLOMACY_APPLICATION");
		$this->diplomacy();
		return $this;
	}

	/**
	* Creates a relationship.
	*
	* @param integer	Candidate alliance
	*
	* @return Alliance
	*/
	protected function acceptRelation($candidateAlly)
	{
		if($this->getRights(array("CAN_MANAGE")))
		{
			// Check existing applications
			$result = sqlSelect("ally_relationships_application", array("mode"), "", "request_ally = ".sqlVal($this->aid)." AND candidate_ally = ".sqlVal($candidateAlly));
			$row = sqlFetch($result);
			sqlEnd($result);
			if($row)
			{
				// Hook::event("ACCEPT_ALLY_RELATION", array($candidateAlly, $row));
				if($row["mode"] != 5)
				{
					Core::getQuery()->insert("ally_relationships", array("rel1", "rel2", "time", "mode"), array($this->aid, $candidateAlly, time(), $row["mode"]));
				}
				else
				{
					Core::getQuery()->delete("ally_relationships", "(rel1 = ".sqlVal($this->aid)." AND rel2 = ".sqlVal($candidateAlly).") OR (rel1 = ".sqlVal($candidateAlly)." AND rel2 = ".sqlVal($this->aid).")");
				}
				$this->deleteAllyRelationApplication($this->aid, $candidateAlly);
			}
		}
		$this->relApplications();
		return $this;
	}

	/**
	* Deletes a relation application.
	*
	* @param integer	Candidate alliance
	* @param integer	Requeseting alliance
	*
	* @return Alliance
	*/
	protected function refuseRelation($candidateAlly, $requestAlly)
	{
		if($this->getRights(array("CAN_MANAGE")))
		{
			if($candidateAlly > 0)
			{
				$result = sqlSelect("ally_relationships_application", array("mode"), "", "request_ally = ".sqlVal($this->aid)." AND candidate_ally = ".sqlVal($candidateAlly));
				$row = sqlFetch($result);
				sqlEnd($result);
				if($row)
				{
					$this->deleteAllyRelationApplication($this->aid, $candidateAlly);
				}
			}
			else if($requestAlly > 0)
			{
				$result = sqlSelect("ally_relationships_application", array("mode"), "", "request_ally = ".sqlVal($requestAlly)." AND candidate_ally = ".sqlVal($this->aid));
				$row = sqlFetch($result);
				sqlEnd($result);
				if($row)
				{
					$this->deleteAllyRelationApplication($this->aid, $requestAlly);
				}
			}

		}
		$this->relApplications();
		return $this;
	}

	/**
	* Deltes a relation application.
	*
	* @param integer
	* @param integer
	*
	* @return Alliance
	*/
	protected function deleteAllyRelationApplication($aid1, $aid2)
	{
		Core::getQuery()->delete("ally_relationships_application", "(request_ally = ".sqlVal($aid1)." AND candidate_ally = ".sqlVal($aid2).") OR (request_ally = ".sqlVal($aid2)." AND candidate_ally = ".sqlVal($aid1).")");
		return $this;
	}

	/**
	* Shows diplomacy management.
	*
	* @return Alliance
	*/
	protected function diplomacy()
	{
		if($this->getRights(array("CAN_MANAGE")))
		{
			$rels = array(); $i = 1;
			$select = array("ar.relid", "ar.mode", "ar.time", "ar.rel1", "a1.name as name1", "a1.tag as tag1", "ar.rel2", "a2.name as name2", "a2.tag as tag2", "u1.username as user1", "u2.username as user2");
			$joins	= "LEFT JOIN ".PREFIX."alliance a1 ON (a1.aid = ar.rel1)";
			$joins .= "LEFT JOIN ".PREFIX."alliance a2 ON (a2.aid = ar.rel2)";
			$joins .= "LEFT JOIN ".PREFIX."user u1 ON (u1.userid = a1.founder)";
			$joins .= "LEFT JOIN ".PREFIX."user u2 ON (u2.userid = a2.founder)";
			$result = sqlSelect("ally_relationships ar", $select, $joins, "(ar.rel1 = ".sqlVal($this->aid)." OR ar.rel2 = ".sqlVal($this->aid).")", "ar.mode ASC");
			while($row = sqlFetch($result))
			{
				$rels[$i]["num"] = $i;
				$rels[$i]["status"] = $this->getDiploStatus($row["mode"]);
				$rels[$i]["time"] = Date::timeToString(1, $row["time"]);
				if(1/*$row["mode"] != 3*/) { $rels[$i]["determine"] = "[".Link::get("game.php/Alliance/id:DetermineRel/relid:".$row["relid"], Core::getLang()->getItem("DETERMINE"))."]"; }
				else { $rels[$i]["determine"] = ""; }
				if($this->aid == $row["rel1"])
				{
					$rels[$i]["alliance"] = Link::get("game.php/AlliancePage/".$row["rel2"], $row["tag2"], $row["name2"]);
					$rels[$i]["founder"] = Link::get("game.php/MSG/Write/Receiver:".$row["user2"], $row["user2"]);
				}
				else
				{
					$rels[$i]["alliance"] = Link::get("game.php/AlliancePage/".$row["rel1"], $row["tag1"], $row["name1"]);
					$rels[$i]["founder"] = Link::get("game.php/MSG/Write/Receiver:".$row["user1"], $row["user1"]);
				}
				$i++;
			}

			// Running applications?
			$result_count = sqlSelectField("ally_relationships_application", "count(*)", "", "candidate_ally = ".sqlVal($this->aid)." OR request_ally = ".sqlVal($this->aid));
			$applications = $result_count;
			if($applications > 0)
			{
				Core::getLang()->assign("applications", $applications);
				Core::getTPL()->assign("applications", Link::get("game.php/Alliance/RelApplications", Core::getLang()->getItem("SHOW_RELATION_APPLICATIONS")));
			}
			else { Core::getTPL()->assign("applications", false); }

			Core::getTPL()->assign("maxpmlength", fNumber(Core::getOptions()->get("MAX_PM_LENGTH")));
			Core::getTPL()->addLoop("relations", $rels);
			Core::getTPL()->display("ally_diplomacy");
			exit();
		}
		return $this;
	}

	/**
	* Determines a relation between two alliances.
	*
	* @param integer	Relation to determine
	*
	* @return Alliance
	*/
	protected function determineRelation($relid)
	{
		if($this->getRights(array("CAN_MANAGE")))
		{
			// Hook::event("DETERMINE_RELATION");
			Core::getQuery()->delete("ally_relationships", "relid = ".sqlVal($relid)
				// ." AND mode != '3'"
				." AND (rel1 = ".sqlVal($this->aid)." OR rel2 = ".sqlVal($this->aid).")");
			doHeaderRedirection("game.php/Alliance/Diplomacy", false);
		}
		Logger::addMessage("CANNOT_DELETE_RELATIONSHIP");
		$this->diplomacy();
		return $this;
	}

	/**
	* Returns the diplomacy status on a given id.
	*
	* @param integer	Status id
	*
	* @return string	Status raw text
	*/
	protected function getDiploStatus($status)
	{
		$ret = "UNKOWN";
		switch($status)
		{
		case 1:
			$ret = "PROTECTION";
			break;
		case 2:
			$ret = "CONFEDERATION";
			break;
		case 3:
			$ret = "WAR";
			break;
		case 4:
			$ret = "TRADE_AGREEMENT";
			break;
		case 5:
			$ret = "CEASEFIRE";
			break;
		}
		return Core::getLang()->getItem($ret);
	}

	/**
	* Alliance management.
	*
	* @return Alliance
	*/
	protected function manageAlliance()
	{
		$select = array("a.aid", "a.founder", "a.foundername", "a.name", "a.tag", "a.textextern", "a.textintern", "a.applicationtext", "a.homepage", "a.logo", "a.showhomepage", "a.showmember", "a.memberlistsort", "a.open", "u2a.rank", "ar.CAN_MANAGE");
		$joins	= "LEFT JOIN ".PREFIX."alliance a ON (a.aid = u2a.aid)";
		$joins .= "LEFT JOIN ".PREFIX."allyrank ar ON (ar.rankid = u2a.rank)";
		$result = sqlSelect("user2ally u2a", $select, $joins, "u2a.userid = ".sqlUser());
		if($row = sqlFetch($result))
		{
			sqlEnd($result);
			if($row["founder"] == NS::getUser()->get("userid") || $row["CAN_MANAGE"])
			{
				$this->resetActions();
				$this->setPostAction("changeprefs", "updateAllyPrefs")
					->setPostAction("changetext", "updateAllyPrefs")
					->addPostArg("updateAllyPrefs", "showmember")
					->addPostArg("updateAllyPrefs", "showhomepage")
					->addPostArg("updateAllyPrefs", "open")
					->addPostArg("updateAllyPrefs", "foundername")
					->addPostArg("updateAllyPrefs", "memberlistsort")
					->addPostArg("updateAllyPrefs", "textextern")
					->addPostArg("updateAllyPrefs", "textintern")
					->addPostArg("updateAllyPrefs", "logo")
					->addPostArg("updateAllyPrefs", "homepage")
					->addPostArg("updateAllyPrefs", "applicationtext");
				$this->setPostAction("changetag", "updateAllyTag")
					->addPostArg("updateAllyTag", "tag")
					->addArg("updateAllyTag", $row["tag"]);
				$this->setPostAction("changename", "updateAllyName")
					->addPostArg("updateAllyName", "name")
					->addArg("updateAllyName", $row["name"]);
				$this->proceedRequest();

				Core::getTPL()->assign("allytag", $row["tag"]);
				Core::getTPL()->assign("allyname", $row["name"]);
				Core::getTPL()->assign("textextern", $row["textextern"]);
				Core::getTPL()->assign("textintern", $row["textintern"]);
				Core::getTPL()->assign("applicationtext", $row["applicationtext"]);
				Core::getTPL()->assign("founder", $row["founder"]);
				Core::getTPL()->assign("foundername", $row["foundername"]);
				Core::getTPL()->assign("logo", $row["logo"]);
				Core::getTPL()->assign("homepage", $row["homepage"]);

				Core::getTPL()->assign("maxallytext", fNumber(Core::getOptions()->get("MAX_ALLIANCE_TEXT_LENGTH")));
				Core::getTPL()->assign("maxapplicationtext", fNumber(Core::getOptions()->get("MAX_APPLICATION_TEXT_LENGTH")));

				if($row["showhomepage"])
				{
					Core::getTPL()->assign("showhp", " checked=\"checked\"");
				}
				if($row["showmember"])
				{
					Core::getTPL()->assign("showmember",	" checked=\"checked\"");
				}
				if($row["open"])
				{
					Core::getTPL()->assign("open",	" checked=\"checked\"");
				}
				switch($row["memberlistsort"])
				{
				case 1: Core::getTPL()->assign("bypoinst",	" selected=\"selected\""); break;
				case 2: Core::getTPL()->assign("byname",	" selected=\"selected\""); break;
				}

				if($row["founder"] == NS::getUser()->get("userid"))
				{
					$referfounder = "";
					$result = sqlSelect("user2ally u2a", array("u2a.userid", "u.username"), "LEFT JOIN ".PREFIX."user u ON (u.userid = u2a.userid)", "u2a.aid = ".sqlVal($row["aid"])." AND u2a.userid != ".sqlUser(), "u.username ASC");
					while($row = sqlFetch($result))
					{
						$referfounder .= createOption($row["userid"], $row["username"], 0);
					}
					sqlEnd($result);
					Core::getTPL()->assign("referfounder", $referfounder);
				}

				Core::getTPL()->display("manage_ally");
				exit();
			}
		}
		else { sqlEnd($result); }
		return $this;
	}

	/**
	* Displays the memberlist.
	*
	* @return Alliance
	*/
	protected function memberList()
	{
		if($this->aid == NS::getUser()->get("aid"))
		{
			$select = array("ar.CAN_SEE_MEMBERLIST", "ar.CAN_MANAGE", "ar.CAN_BAN_MEMBER", "ar.CAN_SEE_ONLINE_STATE", "a.founder", "a.showmember");
			$result = sqlSelect("user2ally u2a", $select, "LEFT JOIN ".PREFIX."allyrank ar ON (u2a.rank = ar.rankid) LEFT JOIN ".PREFIX."alliance a ON (a.aid = u2a.aid)", "u2a.userid = ".sqlUser()." AND u2a.aid = ".sqlVal($this->aid));
			$row = sqlFetch($result);
			sqlEnd($result);
			if($row["founder"] != NS::getUser()->get("userid"))
			{
				$can_see_memberlist = ($row["showmember"]) ? 1 : $row["CAN_SEE_MEMBERLIST"];
				$can_manage = $row["CAN_MANAGE"];
				$can_ban_member = $row["CAN_BAN_MEMBER"];
				$can_see_onlie_state = $row["CAN_SEE_ONLINE_STATE"];
			}
			else
			{
				$can_see_memberlist = 1;
				$can_manage = 1;
				$can_ban_member = 1;
				$can_see_onlie_state = 1;
			}
			Core::getTPL()->assign("founder", $row["founder"]);
		}
		else
		{
			$result = sqlSelect("alliance", "showmember", "", "aid = ".sqlVal($this->aid));
			$row = sqlFetch($result);
			sqlEnd($result);
			$can_see_memberlist = $row["showmember"];
			$can_manage = 0;
			$can_ban_member = 0;
			$can_see_onlie_state = 0;
		}
		unset($row); unset($result);

		if($can_see_memberlist)
		{
			if($can_manage)
			{
				if(Core::getRequest()->getPOST("changeMembers"))
				{
					$result = sqlSelect("user2ally", "userid", "", "aid = ".sqlVal($this->aid));
					while($row = sqlFetch($result))
					{
						$rankid = Core::getRequest()->getPOST("rank_".$row["userid"]);
						Core::getQuery()->update(
							"user2ally",
							"rank",
							$rankid,
							"userid = ".sqlVal($row["userid"])." AND aid = ".sqlVal($this->aid)
								. ' ORDER BY userid'
						);
					}
					sqlEnd($result);
				}
				else if(count(Core::getRequest()->getPOST()) > 0)
				{
					foreach(Core::getRequest()->getPOST() as $key => $value)
					{
						if(preg_match("#^kick\_#i", $key))
						{
							$kickuserid = Str::replace("kick_", "", $key);
				$data = sqlSelectRow("user", array("userid", "username"), "", "userid=".sqlVal($kickuserid));
							Core::getQuery()->delete("user2ally", "userid = ".sqlVal($kickuserid)." AND aid = ".sqlVal($this->aid));
							$_result = sqlSelect("user2ally", "userid", "", "aid = ".sqlVal($this->aid));
							while($_row = sqlFetch($_result))
							{
								new AutoMsg(MSG_MEMBER_KICKED, $_row["userid"], time(), $data);
							}
							sqlEnd($_result);
							unset($_row);
							break;
						}
					}
				}
				$ranks = array();
				$result = sqlSelect("allyrank", array("rankid", "name"), "", "aid = ".sqlVal($this->aid));
				while($row = sqlFetch($result))
				{
					$ranks[$row["rankid"]]["name"] = $row["name"];
				}
				sqlEnd($result);
			}
			if($can_ban_member && $can_see_onlie_state) { $colspan = "8"; }
			else if($can_ban_member || $can_see_onlie_state) { $colspan = "7"; }
			else { $colspan = "6"; }
			Core::getTPL()->assign("colspan", $colspan);
			Core::getTPL()->assign("can_manage", $can_manage);
			Core::getTPL()->assign("can_ban_member", $can_ban_member);
			Core::getTPL()->assign("can_see_onlie_state", $can_see_onlie_state);

			$select = array("u2a.userid", "u.username", "FLOOR(u.points) AS points", "u2a.joindate", "u.last", "g.galaxy", "g.system", "g.position", "ar.rankid", "ar.name AS rankname", "a.tag", "a.founder", "a.foundername");
			$joins	= "LEFT JOIN ".PREFIX."user u ON (u.userid = u2a.userid)";
			$joins .= "LEFT JOIN ".PREFIX."galaxy g ON (g.planetid = u.hp)";
			$joins .= "LEFT JOIN ".PREFIX."allyrank ar ON (ar.rankid = u2a.rank)";
			$joins .= "LEFT JOIN ".PREFIX."alliance a ON (a.aid = u2a.aid)";
			$result = sqlSelect("user2ally u2a", $select, $joins, "u2a.aid = ".sqlVal($this->aid)." ORDER BY u.points DESC LIMIT 100");
			$sum = 0;
			$membercount = 0;
			while($row = sqlFetch($result))
			{
				$uid = $row["userid"];
				$members[$uid]["userid"] = $uid;
				$members[$uid]["username"] = $row["username"];
				$members[$uid]["points"] = fNumber($row["points"]);
				$members[$uid]["joindate"] = Date::timeToString(1, $row["joindate"]);
				$members[$uid]["last"] = $row["last"];
				if($can_manage) { $members[$uid]["rankselection"] = $this->getRankSelect($ranks, $row["rankid"]); }
				if($row["founder"] == $row["userid"])
				{
					if($row["foundername"] != "") { $members[$uid]["rank"] = $row["foundername"]; }
					else { $members[$uid]["rank"] = Core::getLanguage()->getItem("FOUNDER"); }
				}
				else if($row["rankname"])
				{
					$members[$uid]["rank"] = $row["rankname"];
				}
				else
				{
					$members[$uid]["rank"] = Core::getLanguage()->getItem("NEWBIE");
				}
				$members[$uid]["position"] = getCoordLink($row["galaxy"], $row["system"], $row["position"]);
				$members[$uid]["message"] = Link::get("game.php/MSG/Write/Receiver:".$row["username"], Image::getImage("pm.gif", Core::getLanguage()->getItem("WRITE_MESSAGE")));
				if(time() - $row["last"] < 900)
				{
					$members[$uid]["online"] = Image::getImage("on.gif", getTimeTerm(time() - $row["last"]));
				}
				else
				{
					$members[$uid]["online"] = Image::getImage("off.gif", getTimeTerm(time() - $row["last"]));
				}
				$sum += $row["points"];
				$membercount++;
			}
			sqlEnd($result);
			Core::getTPL()->assign("totalmembers", fNumber($membercount));
			Core::getTPL()->assign("totalpoints", fNumber($sum));
			Core::getTPL()->addLoop("members", $members);
			Core::getTPL()->display("memberlist");
			exit();
		}
		return $this;
	}

	/**
	* Generates the select list to chose rank.
	*
	* @param array		Rank data
	* @param integer	Current rank id
	*
	* @return string	Select box content
	*/
	protected function getRankSelect($ranks, $id)
	{
		$return = "";
		foreach($ranks as $key => $value)
		{
			if($id == $key) { $s = 1; } else { $s = 0; }
			$return .= createOption($key, $value["name"], $s);
		}
		return $return;
	}

	/**
	* Checks and saves a new alliance tag.
	*
	* @param string	New tag
	* @param string	Old tag
	*
	* @return Alliance
	*/
	protected function updateAllyTag($tag, $otag)
	{
		$minCharsTag = Core::getOptions()->get("MIN_CHARS_ALLY_TAG");
		$maxCharsTag = Core::getOptions()->get("MAX_CHARS_ALLY_TAG");
		if(!Str::compare($tag, $otag))
		{
			$result_count = sqlSelectField("alliance", "count(*)", "", "tag = ".sqlVal($tag));
			if($result_count > 0)
			{
				$tag = $otag;
				Logger::addMessage("ALLIANCE_ALREADY_EXISTS");
			}
			if(Str::length($tag) < $minCharsTag || Str::length($tag) > $maxCharsTag || !preg_match($this->namePattern, $tag))
			{
				$tag = $otag;
				Logger::addMessage("ALLIANCE_TAG_INVALID");
			}
		}
		Core::getQuery()->update(
			"alliance",
			"tag",
			$tag,
			"aid = ".sqlVal($this->aid)
				. ' ORDER BY aid '
		);
		return $this;
	}

	/**
	* Checks and saves a new alliance name.
	*
	* @param string	New name
	* @param string	Old name
	*
	* @return Alliance
	*/
	protected function updateAllyName($name, $oname)
	{
		$minCharsName = Core::getOptions()->get("MIN_CHARS_ALLY_NAME");
		$maxCharsName = Core::getOptions()->get("MAX_CHARS_ALLY_NAME");
		if(!Str::compare($name, $oname))
		{
			$result_count = sqlSelectField("alliance", "count(*)", "", "tag = ".sqlVal($name));
			if($result_count > 0)
			{
				$name = $oname;
				Logger::addMessage("ALLIANCE_ALREADY_EXISTS");
			}
			if(Str::length($name) < $minCharsName || Str::length($name) > $maxCharsName || !preg_match($this->namePattern, $name))
			{
				$name = $oname;
				Logger::addMessage("ALLIANCE_NAME_INVALID");
			}
		}
		Core::getQuery()->update("alliance", "name", $name, "aid = ".sqlVal($this->aid)
			. ' ORDER BY aid ');
		return $this;
	}

	/**
	* Saves the updated alliance preferences.
	*
	* @param boolean	Show member list to everyone
	* @param boolean	Show homepage to everyone
	* @param boolean	Open applications
	* @param string	Founder rank name
	* @param integer	Default memer list sort
	* @param string	Extern alliance text
	* @param string	Intern alliance text
	* @param string	Logo URL
	* @param string	Homepage URL
	* @param string	Application template
	*
	* @return Alliance
	*/
	protected function updateAllyPrefs(
		$showmember, $showhomepage, $open, $foundername, $memberlistsort,
		$textextern, $textintern, $logo, $homepage, $applicationtext
		)
	{
		if($showmember == 1) { $showmember = 1; } else { $showmember = 0; }
		if($showhomepage == 1) { $showhomepage = 1; } else { $showhomepage = 0; }
		if($open == 1) { $open = 1; } else { $open = 0; }
		if(Str::length($foundername) > Core::getOptions()->get("MAX_CHARS_ALLY_NAME")) { $foundername = ""; }
		$further = 1;
		if(Str::length($textextern) > Core::getOptions()->get("MAX_ALLIANCE_TEXT_LENGTH")) { $further = 0; }
		if(Str::length($textintern) > Core::getOptions()->get("MAX_ALLIANCE_TEXT_LENGTH")) { $further = 0; }
		if((!isValidImageURL($logo) || Str::length($logo) > 128) && $logo != "") { $further = 0; }
		if((!isValidURL($homepage) || Str::length($homepage) > 128) && $homepage != "") { $further = 0; }
		if(Str::length($applicationtext) > Core::getOptions()->get("MAX_APPLICATION_TEXT_LENGTH")) { $further = 0; }
		if($further == 1)
		{
			$atts = array("logo", "textextern", "textintern", "applicationtext", "homepage", "showmember", "showhomepage", "memberlistsort", "open", "foundername");

			$textintern = str_replace('{{', '[[', $textintern);
			$textintern = str_replace('}}', ']]', $textintern);

			$textextern = str_replace('{{', '[[', $textextern);
			$textextern = str_replace('}}', ']]', $textextern);

			$vals = array($logo, Str::validateXHTML($textextern), Str::validateXHTML($textintern), Str::validateXHTML($applicationtext), $homepage, $showmember, $showhomepage, $memberlistsort, $open, Str::validateXHTML($foundername));
			Core::getQuery()->update("alliance", $atts, $vals, "aid = ".sqlVal($this->aid)
				. ' ORDER BY aid ');
			doHeaderRedirection("game.php/Alliance/Manage", false);
		}
		else
		{
			if(Str::length($textextern) > Core::getOptions()->get("MAX_ALLIANCE_TEXT_LENGTH"))
			{
				Core::getTPL()->assign("externerr", Logger::getMessageField("TEXT_INVALID"));
			}
			if(Str::length($textintern) > Core::getOptions()->get("MAX_ALLIANCE_TEXT_LENGTH"))
			{
				Core::getTPL()->assign("internerr", Logger::getMessageField("TEXT_INVALID"));
			}
			if(Str::length($applicationtext) > Core::getOptions()->get("MAX_APPLICATION_TEXT_LENGTH"))
			{
				Core::getTPL()->assign("apperr", Logger::getMessageField("TEXT_INVALID"));
			}
			if((!isValidImageURL($logo) || Str::length($logo) > 128) && $logo != "")
			{
				Core::getTPL()->assign("logoerr", Logger::getMessageField("LOGO_INVALID"));
			}
			if((!isValidURL($homepage) || Str::length($homepage) > 128) && $homepage != "")
			{
				Core::getTPL()->assign("hperr", Logger::getMessageField("HOMEPAGE_INVALID"));
			}
		}
		return $this;
	}

	/**
	* Rank management for the alliance.
	*
	* @param array		POST parameter
	*
	* @return Alliance
	*/
	protected function manageRanks($post)
	{
		if($this->getRights(array("CAN_MANAGE")))
		{
			if($post["createrank"])
			{
				if(Str::length($post["name"]) <= 30)
				{
					Core::getQuery()->insert("allyrank", array("aid", "name"), array($this->aid, Str::validateXHTML($post["name"])));
				}
			}
			else if($post["changerights"])
			{
				$result = sqlSelect("allyrank", "rankid", "", "aid = ".sqlVal($this->aid));
				while($row = sqlFetch($result))
				{
					if(isset($post["CAN_SEE_MEMBERLIST_".$row["rankid"]])) { $can_see_memberlist = 1; }
					else { $can_see_memberlist = 0; }
					if(isset($post["CAN_SEE_APPLICATIONS_".$row["rankid"]])) { $can_sse_applications = 1; }
					else { $can_sse_applications = 0; }
					if(isset($post["CAN_MANAGE_".$row["rankid"]])) { $can_manage = 1; }
					else { $can_manage = 0; }
					if(isset($post["CAN_BAN_MEMBER_".$row["rankid"]])) { $can_ban_member = 1; }
					else { $can_ban_member = 0; }
					if(isset($post["CAN_SEE_ONLINE_STATE_".$row["rankid"]])) { $can_see_onlie_state = 1; }
					else { $can_see_onlie_state = 0; }
					if(isset($post["CAN_WRITE_GLOBAL_MAILS_".$row["rankid"]])) { $can_write_global_mails = 1; }
					else { $can_write_global_mails = 0; }

					$atts = array("CAN_SEE_MEMBERLIST", "CAN_SEE_APPLICATIONS", "CAN_MANAGE", "CAN_BAN_MEMBER", "CAN_SEE_ONLINE_STATE", "CAN_WRITE_GLOBAL_MAILS");
					$vals = array($can_see_memberlist, $can_sse_applications, $can_manage, $can_ban_member, $can_see_onlie_state, $can_write_global_mails);
					Core::getQuery()->update("allyrank", $atts, $vals, "rankid = ".sqlVal($row["rankid"])." AND aid = ".sqlVal($this->aid)
						. ' ORDER BY rankid'
					);
				}
				sqlEnd($result);
			}
			else if($post)
			{
				foreach($post as $key => $value)
				{
					if(preg_match("#^delete\_#i", $key)) { Core::getQuery()->delete("allyrank", "rankid = ".sqlVal(Str::replace("delete_", "", $key))." AND aid = ".sqlVal($this->aid)); break; }
				}
			}

			$select = array("rankid", "name", "CAN_SEE_MEMBERLIST", "CAN_SEE_APPLICATIONS", "CAN_MANAGE", "CAN_BAN_MEMBER", "CAN_SEE_ONLINE_STATE", "CAN_WRITE_GLOBAL_MAILS");
			$result = sqlSelect("allyrank", $select, "", "aid = ".sqlVal($this->aid));
			Core::getTPL()->assign("num", Core::getDB()->num_rows($result));
			Core::getTPL()->addLoop("ranks", $result);
			Core::getTPL()->display("manage_ranks");
			exit();
		}
		return $this;
	}

	/**
	* Refers to writeApplication().
	*
	* @param integer	Alliance id
	*
	* @return Alliance
	*/
	protected function writeApplicationPost($aid)
	{
		$this->writeApplication($aid);
		return $this;
	}

	/**
	* Displays application form.
	*
	* @return Alliance
	*/
	protected function writeApplication()
	{
		$result = sqlSelect("alliance", array("open", "tag", "applicationtext"), "", "aid = ".sqlVal($this->aid));
		if($row = sqlFetch($result))
		{
			sqlEnd($result);
			if($row["open"])
			{
				Core::getTPL()->assign("alliance", $row["tag"]);
				Core::getTPL()->assign("applicationtext", parseBBCode($row["applicationtext"]));
				Core::getTPL()->assign("aid", $this->aid);
				Core::getTPL()->assign("maxapplicationtext", fNumber(Core::getOptions()->get("MAX_APPLICATION_TEXT_LENGTH")));
				Core::getTPL()->display("apply");
				exit();
			}
		}
		else
		{
			sqlEnd($result);
		}
		return $this;
	}

	/**
	* Sends an application to the alliance.
	*
	* @param string	Application text
	*
	* @return Alliance
	*/
	protected function apply($text)
	{
		$result1_count = sqlSelectField("allyapplication", "count(*)", "", "userid = ".sqlUser()." AND aid = ".sqlVal($this->aid));
		$result2_count = sqlSelectField("user2ally", "count(*)", "", "userid = ".sqlUser());

		if($result1_count == 0 && $result2_count == 0 && Str::length($text) > 0)
		{
			$result3_count = sqlSelectField("alliance", "count(*)", "", "aid = ".sqlVal($this->aid)." AND open = '1'");
			if($result3_count > 0)
			{
				$applicationtext = sqlSelectField('alliance', 'applicationtext', '', "aid = ".sqlVal($this->aid)." AND open = '1'");
				Core::getQuery()->insert(
					"allyapplication",
					array("userid", "aid", "date", "application"),
					array(NS::getUser()->get("userid"), $this->aid, time(), Str::validateXHTML( $applicationtext . "\r\n" . $text))
				);
				doHeaderRedirection("game.php/Alliance", false);
				return;
			}
		}
		return $this;
	}

	/**
	* If user search for an alliance to apply.
	*
	* @param string	Alliance to search
	* @param boolean	Start search
	*
	* @return Alliance
	*/
	protected function allySearch($searchItem, $search)
	{
		$result = null;
		$searchItem = Str::validateXHTML($searchItem);
		$searchObj = new String($searchItem);
		$searchObj->setMinSearchLenght(1)
			->prepareForSearch();
		Core::getTPL()->assign("searchitem", $searchItem);
		if($search && $searchObj->validSearchString)
		{
			$select = array("a.aid", "a.tag", "a.name", "a.showhomepage", "a.homepage", "a.open", "COUNT(u2a.userid) AS members", "SUM(u.points) AS points");
			$joins	= "LEFT JOIN ".PREFIX."user2ally u2a ON (a.aid = u2a.aid)";
			$joins .= "LEFT JOIN ".PREFIX."user u ON (u2a.userid = u.userid)";
			$result = sqlSelect("alliance a", $select, $joins, "a.tag LIKE ".sqlVal($searchObj->get())." OR a.name LIKE ".sqlVal($searchObj->get()), "", "", "a.aid");
		}
		$Results = new AllianceList($result);
		Core::getTPL()->addLoop("results", $Results->getArray());
		Core::getTPL()->display("allysearch");
		exit();
		return $this;
	}

	/**
	* Refuses or receipts an application.
	*
	* @param boolean	Receipt
	* @param boolean	Refuse
	* @param array		Candidates
	*
	* @return Alliance
	*/
	protected function manageCadidates($receipt, $refuse, $action)
	{
		if($this->getRights(array("CAN_SEE_APPLICATIONS")))
		{
			foreach($action as $userid)
			{
				$result = sqlSelect("allyapplication ap", array("a.tag", "u.username", "u2a.aid"), "LEFT JOIN ".PREFIX."alliance a ON (a.aid = ap.aid) LEFT JOIN ".PREFIX."user u ON (u.userid = ap.userid) LEFT JOIN ".PREFIX."user2ally u2a ON (u2a.userid = u.userid)", "ap.userid = ".sqlVal($userid)." AND ap.aid = ".sqlVal($this->aid));
				if($row = sqlFetch($result))
				{
					if($receipt && !$row["aid"])
					{
						Core::getQuery()->insert("user2ally", array("userid", "aid", "joindate"), array($userid, $this->aid, time()));
						new AutoMsg(24, $userid, time(), $row);
						$_result = sqlSelect("user2ally", "userid", "", "aid = ".sqlVal($this->aid));
						while($_row = sqlFetch($_result))
						{
							new AutoMsg(100, $_row["userid"], time(), $row);
						}
						sqlEnd($_result);
						Core::getQuery()->delete("allyapplication", "userid = ".sqlVal($userid));
					}
					else
					{
						new AutoMsg(25, $userid, time(), $row);
						Core::getQuery()->delete("allyapplication", "userid = ".sqlVal($userid)." AND aid = ".sqlVal($this->aid));
					}
				}
				sqlEnd($result);
			}
		}
		return $this;
	}

	/**
	* Shows the candidates and their application.
	*
	* @return Alliance
	*/
	protected function candidates()
	{
		if($this->getRights(array("CAN_SEE_APPLICATIONS")))
		{
			$apps = array();
			$result = sqlSelect("allyapplication a", array("a.userid", "a.date", "a.application ", "u.username", "u.points", "g.galaxy", "g.system", "g.position"), "LEFT JOIN ".PREFIX."user u ON (u.userid = a.userid) LEFT JOIN ".PREFIX."galaxy g ON (g.planetid = u.hp)", "a.aid = ".sqlVal($this->aid), "u.username ASC, a.date ASC");
			while($row = sqlFetch($result))
			{
				$apps[$row["userid"]]["date"] = Date::timeToString(1, $row["date"]);
				$apps[$row["userid"]]["message"] = Link::get("game.php/MSG/Write/Receiver:".$row["username"], Image::getImage("pm.gif", Core::getLanguage()->getItem("WRITE_MESSAGE")));
				$apps[$row["userid"]]["apptext"] = parseBBCode($row["application"]);
				$apps[$row["userid"]]["userid"] = $row["userid"];
				$apps[$row["userid"]]["username"] = $row["username"];
				$apps[$row["userid"]]["points"] = fNumber($row["points"]);
				$apps[$row["userid"]]["position"] = getCoordLink($row["galaxy"], $row["system"], $row["position"]);
			}
			Core::getTPL()->assign("candidates", sprintf(Core::getLanguage()->getItem("CANDIDATES"), Core::getDB()->num_rows($result)));
			sqlEnd($result);
			Core::getTPL()->addLoop("applications", $apps);
			Core::getTPL()->display("applications");
			exit();
		}
		return $this;
	}

	/**
	* User leaves the alliance.
	*
	* @return Alliance
	*/
	protected function leaveAlliance()
	{
		$result = sqlSelect("alliance", "founder", "", "aid = ".sqlVal(NS::getUser()->get("aid")));
		if($row = sqlFetch($result))
		{
			if($row["founder"] != NS::getUser()->get("userid"))
			{
				Core::getQuery()->delete("user2ally", "userid = ".sqlUser());
				NS::getUser()->rebuild();
				$_result = sqlSelect("user2ally", "userid", "", "aid = ".sqlVal($this->aid));
				while($_row = sqlFetch($_result))
				{
					$data = array(
						"username"	=> NS::getUser()->get("username")
						);
					new AutoMsg(101, $_row["userid"], time(), $data);
				}
				sqlEnd($_result);
			}
		}
		sqlEnd($result);
		doHeaderRedirection("game.php/Main", false);
		return $this->index();
	}

	/**
	* Refers the founder status to a different alliance member.
	*
	* @param integer	Referring user id
	*
	* @return Alliance
	*/
	protected function referFounderStatus($userid)
	{
		$result = sqlSelect("alliance", "founder", "", "aid = ".sqlVal(NS::getUser()->get("aid")));
		if($row = sqlFetch($result))
		{
			sqlEnd($result);
			if($row["founder"] == NS::getUser()->get("userid"))
			{
				$result = sqlSelect("user2ally", "rank", "", "userid = ".sqlVal($userid)." AND aid = ".sqlVal(NS::getUser()->get("aid")));
				if($row = sqlFetch($result))
				{
					sqlEnd($result);
					Core::getQuery()->update("alliance", "founder", $userid, "aid = ".sqlVal(NS::getUser()->get("aid")). ' ORDER BY aid');
					Core::getQuery()->update("user2ally", "rank", $row["rank"], "aid = ".sqlVal(NS::getUser()->get("aid"))." AND userid = ".sqlUser() . ' ORDER BY userid ');
				}
			}
		}
		else
		{
			sqlEnd($result);
		}
		return $this->index();
	}

	/**
	* Abandons an alliance, if user has permissions.
	*
	* @return Alliance
	*/
	protected function abandonAlly()
	{
		$result = sqlSelect("alliance", "founder", "", "aid = ".sqlVal(NS::getUser()->get("aid")));
		if($row = sqlFetch($result))
		{
			sqlEnd($result);
			if($row["founder"] == NS::getUser()->get("userid"))
			{
				deleteAlliance(NS::getUser()->get("aid"));
				NS::getUser()->rebuild();
			}
		}
		return $this->index();
	}

	/**
	* Allows the user to write a global mail to all alliance member.
	*
	* @param boolean	Send global message
	* @param string	Message
	* @param string	Subject
	* @param integer	Send only to the specific ranks
	*
	* @return Alliance
	*/
	protected function globalMail($sendGlobalMessage, $message, $subject, $receiver)
	{
		$result = sqlSelect("user2ally u2a", array("a.founder", "ar.CAN_WRITE_GLOBAL_MAILS"), "LEFT JOIN ".PREFIX."alliance a ON (a.aid = u2a.aid) LEFT JOIN ".PREFIX."allyrank ar ON (ar.rankid = u2a.rank)", "u2a.userid = ".sqlUser());
		if($row = sqlFetch($result))
		{
			sqlEnd($result);
			if($row["CAN_WRITE_GLOBAL_MAILS"] || $row["founder"] == NS::getUser()->get("userid"))
			{
				Core::getLanguage()->load("Message");
				if($sendGlobalMessage)
				{
					$message = Str::validateXHTML($message);
					$subject = Str::validateXHTML($subject);
					if(Str::length($message) > 2 && Str::length($message) <= Core::getOptions()->get("MAX_PM_LENGTH") && Str::length($subject) > 0 && Str::length($subject) < 101)
					{
                        NS::checkSpan($message, 'Ally.globalMail.message');
                        NS::checkSpan($subject, 'Ally.globalMail.subject');

						$_result = sqlSelect("user2ally", "userid", "", (($receiver == "foo") ? "aid = ".sqlVal($this->aid) : "rank = ".sqlVal($receiver)." AND aid = ".sqlVal($this->aid)));
						while($_row = sqlFetch($_result))
						{
							Core::getQuery()->insert("message", array("mode", "time", "sender", "receiver", "message", "subject", "readed"), array(6, time(), NS::getUser()->get("userid"), $_row["userid"], $message, $subject, (($_row["userid"] == NS::getUser()->get("userid")) ? 1 : 0)));
						}
						sqlEnd($_result);
						Logger::addMessage("SENT_SUCCESSFUL", "success");
					}
					else
					{
						if(Str::length($message) < 3 || Str::length($message) > Core::getOptions()->get("MAX_PM_LENGTH")) { Core::getTPL()->assign("messageError", Logger::getMessageField("MESSAGE_TOO_SHORT")); }
						if(Str::length($subject) == 0 || Str::length($subject) > 100) { Core::getTPL()->assign("subjectError", Logger::getMessageField("SUBJECT_TOO_SHORT")); }
					}
				}
				$ranks = sqlSelect("allyrank", array("rankid", "name"), "", "aid = ".sqlVal($this->aid));
				Core::getTPL()->assign("maxpmlength", fNumber(Core::getOptions()->get("MAX_PM_LENGTH")));
				Core::getTPL()->addLoop("ranks", $ranks);
				Core::getTPL()->display("globalmail");
				exit();
			}
		}
		return $this;
	}

	protected function bbcode($sourse)
	{
		return bbcode_text($sourse);
	}
}
?>