<?php
/**
* Shows messages.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class MSG extends Page
{
	protected $msg_per_page = 20;

	/**
	* Constructor: Displays message folders.
	*
	* @return void
	*/
	public function __construct()
	{
		parent::__construct();
		Core::getLanguage()->load("Message");
		$this->setPostAction("send", "sendMessage")
			->addPostArg("sendMessage", "receiver")
			->addPostArg("sendMessage", "subject")
			->addPostArg("sendMessage", "message")
			->setPostAction("delete", "deleteMessages")
			->addGetArg("deleteMessages", "folder")
			->addPostArg("deleteMessages", "deleteOption")
			->addPostArg("deleteMessages", "msgid")
			->addPostArg("deleteMessages", "page")
			->setGetAction("id", "DeleteAll", "deleteAllMessages")
			->setGetAction("id", "Write", "createNewMessage")
			->addGetArg("createNewMessage", "Receiver")
			->addGetArg("createNewMessage", "Reply")
			->setGetAction("id", "ReadFolder", "readFolder")
			->addGetArg("readFolder", "folder")
			->addPostArg("readFolder", "page")
			->addGetArg("readFolder", "ruser")
			->proceedRequest();
		return;
	}

	/**
	* Index action.
	*
	* @return MSG
	*/
	protected function index()
	{
		$select = array("f.folder_id", "f.label", "f.is_standard", "COUNT(m.msgid) AS messages", "SUM(m.readed) AS `read`", "SUM(LENGTH(m.message)) AS `storage`");
		$joins = "LEFT JOIN ".PREFIX."message m ON m.mode = f.folder_id AND m.receiver = ".sqlUser();
		$where = "f.userid = ".sqlUser()." OR f.is_standard = '1'";
		$result = sqlSelect("folder f", $select, $joins, $where, "", "", "f.display_order, f.folder_id");
		$folders = array();
		while($row = sqlFetch($result))
		{
			$unreadMessages = $row["messages"] - intval($row["read"]);
			if($unreadMessages > 1)
			{
				$read = "UNREADED";
				$newMessages = sprintf(Core::getLanguage()->getItem("F_NEW_MESSAGES"), fNumber($unreadMessages));
			}
			else if($unreadMessages > 0)
			{
				$read = "UNREADED";
				$newMessages = Core::getLanguage()->getItem("F_NEW_MESSAGE");
			}
			else
			{
				$read = "READED";
				$newMessages = "";
			}
			$label = ($row["is_standard"]) ? Core::getLang()->get($row["label"]) : $row["label"];
			$link = "game.php/go:MSG/id:ReadFolder/folder:".$row["folder_id"];
			$folders[] = array(
				"image"		=> Image::getImage(strtolower($read).".gif", Core::getLang()->get($read)),
				"label"		=> Link::get($link, $label, Core::getLang()->get($read)),
				"messages"	=> fNumber($row["messages"]),
				"newMessages" => $newMessages,
				"size"		=> File::bytesToString($row["storage"]),
				);
		}
		sqlEnd($result);
		Core::getTPL()->addLoop("folders", $folders);

		Core::getTPL()->display("messages");
		return $this->quit();
	}

	protected function deleteAllMessages()
	{
		Core::getQuery()->delete("message", "receiver = ".sqlUser());
		doHeaderRedirection("game.php/MSG", false);
		// return $this->index();
	}

    protected function checkSendMessage()
    {
        if(NS::getUser()->get("observer")){
            Logger::dieMessage('CANT_SEND_MESSAGE_DUE_OBSERVER');
        }
        if(NS::getUser()->get("umode")){
            Logger::dieMessage('CANT_SEND_MESSAGE_DUE_VACATION');
        }
        if(NS::getUser()->get("points") < ALLOW_SEND_MESSAGE_POINTS){
            Core::getLanguage()->assign('points', ALLOW_SEND_MESSAGE_POINTS);
            Logger::dieMessage('CANT_SEND_MESSAGE_DUE_POINTS');
        }
    }

    /**
	* Form to send a new message.
	*
	* @param string	Receiver
	* @param string	Reply subject
	*
	* @return MSG
	*/
	protected function createNewMessage($receiver, $reply)
	{
        $this->checkSendMessage();

		// echo "[createNewMessage] $receiver, $reply"; exit;
		$receiver = rawurldecode($receiver);
		$reply = rawurldecode($reply);
		Core::getLanguage()->load("Message");
		Core::getTPL()->assign("receiver", $receiver);
		Core::getTPL()->assign("subject", ($reply != "") ? $reply : Core::getLanguage()->getItem("NO_SUBJECT"));
		Core::getTPL()->assign("maxpmlength", fNumber(Core::getOptions()->get("MAX_PM_LENGTH")));
		$sendAction = RELATIVE_URL."game.php/MSG/Write/Receiver:".rawurlencode($receiver);
			// . ( (defined('SN')) ? ('?' . ''): ('') );
		if($reply != "")
		{
			$sendAction .= "/Reply:".rawurlencode($reply);
		}
		Core::getTPL()->assign("sendAction", socialUrl($sendAction));
		Core::getTPL()->display("writemessages");
		return $this;
	}

	/**
	* Sends a message.
	*
	* @param string	Receiver name
	* @param string	Subject
	* @param string	Message
	*
	* @return MSG
	*/
	protected function sendMessage($receiver, $subject, $message)
	{
        $this->checkSendMessage();

		$charset = Core::getLanguage()->getOpt("charset");
		if(Str::length($message, $charset) > 2
			&& Str::length($message, $charset) <= Core::getOptions()->get("MAX_PM_LENGTH")
			&& Str::length($subject, $charset) > 0
			&& Str::length($subject, $charset) < 101)
		{
            NS::checkSpan($message, 'MSG.PM.message');
            NS::checkSpan($subject, 'MSG.PM.subject');

			$result = sqlSelect("user", "userid", "", "username = ".sqlVal($receiver));
			if($row = sqlFetch($result))
			{
				sqlEnd($result);
				if($row["userid"] != NS::getUser()->get("userid"))
				{
					$subject = preg_replace("#((RE|FW):\s)+#is", "\\1", $subject); // Remove excessive reply or forward notes
					// Hook::event("SEND_PRIVATE_MESSAGE", array(&$row, $receiver, &$subject, &$message));
					Core::getQuery()->insert("message", array("mode", "time", "sender", "receiver", "subject", "message", "readed", "related_user"), array(1, time(), NS::getUser()->get("userid"), $row["userid"], Str::validateXHTML($subject), Str::validateXHTML($message), 0, $row["userid"]));
					Core::getQuery()->insert("message", array("mode", "time", "sender", "receiver", "subject", "message", "readed", "related_user"), array(2, time(), $row["userid"], NS::getUser()->get("userid"), Str::validateXHTML($subject), Str::validateXHTML($message), 1, NS::getUser()->get("userid")));
					$eventid = Core::getDB()->insert_id();
					/*
					if($eventid % 500 == 0)
					{
					Core::getQuery()->optimize("message");
					}
					if($eventid > 10000000)
					{
					require_once(MAINTAIN_DIR."TableCleaner.class.php");
					$TableCleaner = new TableCleaner("message", "msgid");
					$TableCleaner->clean();
					$TableCleaner->kill();
					}
					*/
					Logger::addMessage("SENT_SUCCESSFUL", "success");
				}
				else
				{
					Core::getTPL()->assign("userError", Logger::getMessageField("SELF_MESSAGE"));
				}
			}
			else
			{
				sqlEnd($result);
				Core::getTPL()->assign("userError", Logger::getMessageField("USER_NOT_FOUND"));
			}
		}
		else
		{
			if(Str::length($message) < 3 || Str::length($message) > Core::getOptions()->get("MAX_PM_LENGTH")) { Core::getTPL()->assign("messageError", Logger::getMessageField("MESSAGE_TOO_SHORT")); }
			if(Str::length($subject) == 0 || Str::length($subject) > 100) { Core::getTPL()->assign("subjectError", Logger::getMessageField("SUBJECT_TOO_SHORT")); }
		}
		return $this;
	}

	/**
	* Deletes messages.
	*
	* @param integer Mode to delete content
	* @param integer Folder id
	*
	* @return MSG
	*/
	protected function deleteMessages($folder, $option, $msgs, $page)
	{
		if( !mt_rand(0, 1000) )
		{
			$deltime = 604800;
			if(is_numeric(Core::getOptions()->get("DEL_MESSAGE_DAYS")) && Core::getOptions()->get("DEL_MESSAGE_DAYS") > 0)
			{
				$deltime = intval(Core::getOptions()->get("DEL_MESSAGE_DAYS")) * 86400;
			}
			$deltime = time() - $deltime;
			Core::getQuery()->delete("message", "time <= ".sqlVal($deltime));
		}

		$joins	= "LEFT JOIN ".PREFIX."user u ON (u.userid = m.sender)";
		$joins .= "LEFT JOIN ".PREFIX."galaxy g ON (g.planetid = u.hp)";
		$result_count = sqlSelectField("message m", "count(*)", $joins, "receiver = ".sqlUser()." AND mode = ".sqlVal($folder), "time DESC, msgid DESC");
		$pages = ceil($result_count / $this->msg_per_page);
		if(!is_numeric($page))
		{
			$page = 1;
		}
		else if($page > $pages) { $page = $pages; }
		else if($page < 1) { $page = 1; }
		$start = abs(($page - 1) * $this->msg_per_page);
		$max = $this->msg_per_page;

		switch($option)
		{
		case 1:
			foreach($msgs as $msgid)
			{
				Core::getQuery()->delete("message", "msgid = ".sqlVal($msgid)." AND receiver = ".sqlUser());
			}
			break;
		case 2:
			$result = sqlSelect("message m", array("m.*", "u.username", "g.galaxy", "g.system", "g.position"), $joins, "receiver = ".sqlUser()." AND mode = ".sqlVal($folder), "time DESC, msgid DESC", "$start, $max");
			while($row = sqlFetch($result))
			{
				if(!in_array($row["msgid"], $msgs))
				{
					Core::getQuery()->delete("message", "msgid = ".sqlVal($row["msgid"]));
				}
			}
			sqlEnd($result);
			break;
		case 3:
			$result = sqlSelect("message m", array("m.*", "u.username", "g.galaxy", "g.system", "g.position"), $joins, "receiver = ".sqlUser()." AND mode = ".sqlVal($folder), "time DESC, msgid DESC", "$start, $max");
			while($row = sqlFetch($result))
			{
				Core::getQuery()->delete("message", "msgid = ".sqlVal($row["msgid"]));
			}
			sqlEnd($result);
			break;
		case 4:
			Core::getQuery()->delete("message", "receiver = ".sqlUser()." AND mode = ".sqlVal($folder));
			break;
		case 5:
			$reports = array();
			// $modId = NS::getRandomModerator();
		$moderators = NS::getModerators();
			foreach($msgs as $msgid)
			{
				$result = sqlSelect("message m", array("m.sender", "m.message", "m.time", "u.username"), "LEFT JOIN ".PREFIX."user u ON (u.userid = m.sender)", "m.msgid = ".sqlVal($msgid)." AND m.receiver = ".sqlUser());
				if($row = sqlFetch($result))
				{
					if($row["sender"] > 0
			// && $row["sender"] != $modId)
			&& !in_array($row["sender"], $moderators))
					{
						$reports[] = $row;
					}
				}
			}
			if(count($reports) > 0)
			{
				Logger::addMessage("MESSAGES_REPORTED", "success");
				// 37.7.3: NS::getUser()->get('username') уже эскейпится в User::get (план 37.7.1)
				Core::getLang()->assign("reportSender", NS::getUser()->get("username"));
				foreach($reports as $report)
				{
					Core::getLang()->assign("reportMessage", htmlspecialchars((string)$report["message"], ENT_QUOTES, 'UTF-8'));
					Core::getLang()->assign("reportUser", htmlspecialchars((string)$report["username"], ENT_QUOTES, 'UTF-8'));
					Core::getLang()->assign("reportSendTime", Date::timeToString(1, $report["time"], "", false));
					$message = Core::getLang()->get("MODERATOR_REPORT_MESSAGE");
					$subject = Core::getLang()->get("MODERATOR_REPORT_SUBJECT");
					$attr = array("sender", "mode", "subject", "message", "receiver", "time", "readed");
			$time = time();
			foreach($moderators as $modId)
			{
				$vals = array(null, 1, $subject, $message, $modId, $time, 0);
					Core::getQuery()->insert("message", $attr, $vals);
				}
			}
			}
			break;
		}
		return $this;
	}

	/**
	* Shows content of a message folder.
	*
	* @param integer	Folder id
	*
	* @return MSG
	*/
	protected function readFolder($id, $ruser, $page)
	{
        $m = array();
        $joins	= "LEFT JOIN ".PREFIX."user u ON (u.userid = m.sender)";
        $joins .= "LEFT JOIN ".PREFIX."galaxy g ON (g.planetid = u.hp)";
        $mode = " AND m.mode = ".sqlVal($id);
        if (is_numeric($ruser))
            $mode = " AND m.related_user=$ruser";
        $result_count = sqlSelectField("message m", "count(*)", $joins, "m.receiver = ".sqlUser().$mode);
        $pages = ceil($result_count / $this->msg_per_page);
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
            $n = $i * $this->msg_per_page + 1;
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
        Core::getTPL()->assign("page", $page);
        Core::getTPL()->assign("page_links", $pages_link);
        Core::getTPL()->assign("pages", $pages_sel);
        $start = abs(($page - 1) * $this->msg_per_page);
        $max = $this->msg_per_page;
        /* Pages.END */

        $result = sqlSelect("message m", array("m.*", "u.username", "g.galaxy", "g.system", "g.position"), $joins, "m.receiver = ".sqlUser().$mode, "m.time DESC, msgid DESC", ' '.$start . ", " . $max);
        $odd = false;
        $read_ids = array();
        while($row = sqlFetch($result))
        {
            // 37.7.3: XSS — username из user-controlled, эскейпать перед использованием в HTML
            $row["username"] = htmlspecialchars((string)$row["username"], ENT_QUOTES, 'UTF-8');
            $read_ids[] = $row["msgid"];
            $mid = $row["msgid"];
            $m[$mid]["odd"] = $odd ? ' class="odd"' : "";
            $odd = !$odd;
            $m[$mid]["msgid"] = $mid;
            $m[$mid]["sender"] = ($row["sender"] > 0) ? $row["sender"] : Core::getLanguage()->getItem("FLEET_COMMAND");
            $m[$mid]["msg"] = parseBBCode($row["message"]);
            $m[$mid]["subject"] = $row["subject"];
            $m[$mid]["time"] = Date::timeToString(1, $row["time"]);
            if(($row["mode"] == 1 || $row["mode"] == 2 || $row["mode"] == 6) && !isAdmin())
            {
                if(NS::checkSpan($m[$mid]["subject"], '', false)){
                    // unset($m[$mid]);
                    // continue;
                    $m[$mid]["subject"] = "Unknown";
                }
                if(NS::checkSpan($m[$mid]["msg"], '', false)){
                    // unset($m[$mid]);
                    // continue;
                    $m[$mid]["msg"] = "Violation of User Agreement 5.13";
                }
            }
            if($row["mode"] == 1 || $row["mode"] == 2)
            {
                $reply = Link::get("game.php/MSG/Write/Receiver:".$row["username"]."/Reply:".rawurlencode("RE: ".$row["subject"]), Image::getImage("pm.gif", Core::getLanguage()->getItem("REPLY")));
                $m[$mid]["sender"] = ($row["sender"] > 0) ? $reply." ".$row["username"]." ".getCoordLink($row["galaxy"], $row["system"], $row["position"]) : "System";
            }
            else if($row["mode"] == 3 || $row["mode"] == 4 || $row["mode"] == 7)
            {
                $m[$mid]["subject"] = Str::replace("%SID%", SID, $m[$mid]["subject"]);
                $m[$mid]["msg"] = Str::replace("%SID%", SID, $m[$mid]["msg"]);
                if( defined('SN') )
                {
                    // $m[$mid]["subject"] = Str::replace("%ODNOKLASSNIKI%", '?' . '', $m[$mid]["subject"]);
                    // $m[$mid]["msg"] = Str::replace("%ODNOKLASSNIKI%", '?' . ''/* . ''*/, $m[$mid]["msg"]);
                    $m[$mid]["msg"] = preg_replace("#\"((?:".preg_quote(RELATIVE_URL."game.php", "#")."|".preg_quote(FULL_URL."game.php", "#").").*?)\"#ise", '$this->fixMessageLink("\1")', $m[$mid]["msg"]);
                }
            }
            else if($row["mode"] == 6)
            {
                $m[$mid]["sender"] = ($row["username"]) ? Core::getLanguage()->getItem("ALLIANCE_GLOBAL_MAIL")."(".$row["username"].")" : Core::getLanguage()->getItem("ALLIANCE");
            }
            else
            {
                if($row["mode"] == 5 && preg_match("#^(\d+)(.*)$#is", $row["message"], $regs))
                {
                    $assault_id = $regs[1];
                    $assault_type = substr($regs[2], 1, 100);
                    saveAssaultReportSID();
                    $select = array("a.`key`", "a.key2", "a.turns", "a.result",
                        "a.attacker_lost_res", "a.defender_lost_res", "a.gentime", "a.planetid"
                    );
                    $_row = sqlSelectRow("assault a", $select, "", "a.assaultid = ".sqlVal($assault_id));
                    if($_row)
                    {
                        $m[$mid]["subject"] = getCoordLink(0, 0, 0, false, $_row["planetid"], true);
                        $url = HTTP_HOST.REQUEST_DIR."AssaultReport.php?id=".$assault_id;
                        if($assault_type == "atter" && $_row["turns"] < 2 && $_row["result"] == 2)
                        {
                            $url .= "&key2=".$_row["key2"];
                        }
                        else
                        {
                            $url .= "&key=".$_row["key"];
                        }
                        $url = socialUrl($url);
                        /* if( defined('SN') )
                        {
                            $url .= '&' . '';
                        } */
                        $gentime = $_row["gentime"] / 1000;
                        $label = Core::getLanguage()->getItem($assault_type != "rocket" ? "ASSAULT_REPORT" : "ROCKET_ATTACK_REPORT")
                            . " (A:".fNumber($_row["attacker_lost_res"])
                            . ", D: ".fNumber($_row["defender_lost_res"])
                            . ($assault_type != "rocket" ? ", ".Core::getLanguage()->getItem("ASSAULT_REPORT_TURNS").": ".$_row["turns"] : "")
                            . ") ".$gentime."s";
                        $m[$mid]["msg"] = "<center><a class='assault-report' href='{$url}' target='_blank'>{$label}</a></center>";
                    }
                }
                $m[$mid]["sender"] = Core::getLanguage()->getItem("FLEET_COMMAND");
            }
        }
        sqlEnd($result);
        if (count($read_ids) > 0)
            Core::getQuery()->update("message", "readed", 1, "msgid in (". implode(",", $read_ids).")" . ' ORDER BY msgid');
        Core::getTPL()->addLoop("messages", $m);
        Core::getTPL()->assign("mode", $id);
        Core::getTPL()->display("folder");
		return $this;
	}

	protected function fixMessageLink($url)
	{
		return socialUrl($url);
		/*
		if( defined('SN') )
		{
			$sn_params = '';
			if(!strstr($url, $sn_params))
			{
				$url .= strpos($url, '?') === false ? '?' : '&';
				$url .= $sn_params;
			}
		}
		return $url;
		*/
	}
}
?>