<?php
/**
* Chat module.
*
* Oxsar http://oxsar.ru
*
*
*/

class Chat extends Page
{
	public function __construct()
	{
		parent::__construct();
        Core::getLanguage()->load("Prefs");
		$this->setPostAction("send_message", "sendMessage");
		$this->addPostArg("sendMessage", "shoutbox_message");
		$this->proceedRequest();
	}

	protected function index()
	{
		function bbcode($sourse)
		{
			$bb[] = "#\[b\](.*?)\[/b\]#si";
			$html[] = "<b>\\1</b>";
			$bb[] = "#\[i\](.*?)\[/i\]#si";
			$html[] = "<i>\\1</i>";
			$bb[] = "#\[u\](.*?)\[/u\]#si";
			$html[] = "<u>\\1</u>";
			$bb[] = "#\[s\](.*?)\[/s\]#si";
			$html[] = "<s>\\1</s>";
			$bb[] = "#\[color=(.*?)\](.*?)\[/color\]#si";
			$html[] = "<font color=\\1>\\2</font>";
			$bb[] = "#\[:(.*?):\]#si";
			$html[] = "<img src=".RELATIVE_URL."chat/emo/\\1.gif?".CLIENT_VERSION.">";
			$bb[] = "#\[img\](.*?)\[/img\]#si";
			$html[] = "<a href=\"javascript:void(0);\" onclick=\"window.open('".RELATIVE_URL."refdir.php?url=\\1');\" class=\"external\">\\1</a>";
			$bb[] = "#\[url\](.*?)\[/url\]#si";
			$html[] = "<a href=\"javascript:void(0);\" onclick=\"window.open('".RELATIVE_URL."refdir.php?url=\\1');\" class=\"external\">\\1</a>";
			$bb[] = "#\[url=(.*?)\](.*?)\[/url\]#si";
			$html[] = "<a href=\"javascript:void(0);\" onclick=\"window.open('".RELATIVE_URL."refdir.php?url=\\1');\" class=\"external\">\\2</a>";
			$sourse = preg_replace ($bb, $html, $sourse);
			return $sourse;
		}
		$messages = array();
		$result = sqlSelect("chat c", array("c.time", "c.message", "u.username", "g.usergroupid"), "LEFT JOIN ".PREFIX."user u ON (u.userid = c.userid) LEFT JOIN ".PREFIX."user2group g ON (g.userid = u.userid)", "u.userid > '0'", "messageid DESC", "75");
		while($row = sqlFetch($result))
		{
			$row["time"] = date("[H:i:s]", $row["time"]);
			$row["username"] = htmlspecialchars((string)$row["username"], ENT_QUOTES, 'UTF-8');
			$row["message"] = bbcode(stripslashes($row["message"]));
			$messages[] = $row;
		}
		Core::getTPL()->addLoop("shoutBoxMessages", $messages);
		Core::getTPL()->assign("eROreason", $this->checkRO());
		Core::getTPL()->assign('chat_link', Link::get('game.php/Chat/', '&#171;&#171;&#171;&#171;&#171; Общий чат &#187;&#187;&#187;&#187;&#187;'));
		Core::getTPL()->assign('a_chat_link', Link::get('game.php/ChatAlly/', 'Чат альянса'));
		Core::getTPL()->display("chat");
		return $this->quit();
	}

	protected function checkRO()
	{
		if(NS::getUser()->get("observer")){
            return Core::getLanguage()->getItem('OBSERVER_MODE_ENABLED');
        }
		if(NS::getUser()->get("umode")){
            return Core::getLanguage()->getItem('VACATION_MODE');
        }
        return false;
	}

	protected function sendMessage($message)
	{
		$message = preg_replace("#\<#","&lt;",$message);
		$message = preg_replace("#\>#","&gt;",$message);

		function Teg($one,$teg,$too,$path)
		{
			$on = preg_quote($one,"~");
			$to = preg_quote($too,"~");
			$saerch1 = preg_quote($one,"~") . $teg . preg_quote ($too,"~");
			$saerch2 = preg_quote($one,"~") . "/" . $teg . preg_quote ($too,"~");
			$path = preg_replace("~[ ]+~"," ",$path);
			$path = trim($path);
			$path = preg_replace("~(".$saerch1."[ ]?".$saerch1.")+~", $one.$teg.$too, $path);
			$path = preg_replace("~(".$saerch2."[ ]?".$saerch2.")+~", $one."/".$teg.$too, $path);
			$path = preg_replace("~ˆ([ ]*".$saerch2.")*~","",$path);
			$path = preg_replace("~(".$saerch1."[ ]*)*$~","",$path);
			$_search = "~".$saerch1."(.+)".$saerch2."~U";
			$search = array();
			if ( preg_match_all($_search, $path, $array, PREG_PATTERN_ORDER)){
				while ( list(, $val) = each($array[0])){
					$content = "~" . preg_quote($val,"~") . "~U";
					if(@array_search($content,$search) or $content==$search[0]){ continue; }
					$search[] = $content;
					$pp = preg_replace("~".$on."[/]?".$teg.$to."~","",$val);
					if(!preg_match ("~[a-zA-Z0-9а-яА-Я_]~",$pp)){ $pp = " "; }
					else{ $pp = $one.$teg.$too.$pp.$one."/".$teg.$too; }
					$replace[] = $pp;
				}
				$search[] = "~0~e";
				$replace[] = "0";
				$path = preg_replace($search, $replace, $path);
				$path = preg_replace("~[ ]+~"," ",$path);
				$path = preg_replace("~".$on."([/]?)".$teg.$to."[ ]?".$on."\\1".$teg.$to."~s", $one."\\1".$teg.$too, $path);
			}
			if(!preg_match ($_search, $path)){ $path = preg_replace ("~".$on."[/]?".$teg.$to."~", '', $path); }
			return $path;
		}

		$message = Teg("[","b","]",$message);
		$message = Teg("[","i","]",$message);
		$message = Teg("[","u","]",$message);
		$message = Teg("[","s","]",$message);

		// План 50 Ф.4 (149-ФЗ): UGC-маскирование сообщений чата.
		$message = Moderation::mask($message);

		if( $message != "" && !$this->checkRO() )
		{
			$userid = NS::getUser()->get("userid");
			$mes = sqlSelectRow("chat", array("messageid", "message"), "", "userid = ".sqlVal($userid), "messageid DESC", "1");
			if( $message != $mes["message"] )
			{
				sqlInsert("chat", array("time" => time(), "userid" => $userid, "message" => $message));
			}
		}
		return $this->index();
	}

}
?>