<?php
$dhost = "localhost";
$duser = "root";
$duserpw = "asd50.1lmk";
$dname = "xnova-uni2";

echo "<meta http-equiv=\"Content-Type\" content=\"text/html; charset=utf-8\" />";
echo "<script type=\"text/javascript\" src=\"".RELATIVE_URL."chat.js\"></script>";

mysql_connect($dhost,$duser,$duserpw);
mysql_select_db($dname);

$sql = 'SELECT c.time, c.message, u.username, g.usergroupid FROM na_chat c LEFT JOIN na_user u ON (u.userid = c.userid) LEFT JOIN na_user2group g ON (g.userid = c.userid) WHERE u.userid > \'0\' ORDER BY time DESC LIMIT 50';
$result = mysql_query($sql);
	while($row = mysql_fetch_array($result, MYSQL_BOTH))
	{
		$row["time"] = date("[H:i:s]", $row["time"]);
		$row["message"] = bbcode($row["message"]);

echo $row["time"]."<a href=\"javascript:insertname('".$row["username"]."')>";
if($row["usergroupid"] == 2)
{
echo "<font color=#FF0000><b>".$row["username"]."</b></font>";
} elseif($row["usergroupid"] == 4) {
echo "<font color=#00FF00><b>".$row["username"]."</b></font>";
} else {
echo "<font color=#82b7ff><b>".$row["username"]."</b></font>";
}
echo "</a>: ".$row["message"]."<br />";

}
mysql_close();

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
			$bb[] = "#\[:(.*?):\]#si";
			$html[] = "<img src=".RELATIVE_URL."chat/emo/\\1.gif>";
			$bb[] = "#\[img\](.*?)\[/img\]#si";
			$html[] = "<a href=\"javascript:void(0);\" onclick=\"window.open('".RELATIVE_URL."refdir.php?url=\\1');\" class=\"external\">\\1</a>";
			$bb[] = "#\[url\](.*?)\[/url\]#si";
			$html[] = "<a href=\"javascript:void(0);\" onclick=\"window.open('".RELATIVE_URL."refdir.php?url=\\1');\" class=\"external\">\\1</a>";
			$bb[] = "#\[url=(.*?)\](.*?)\[/url\]#si";
			$html[] = "<a href=\"javascript:void(0);\" onclick=\"window.open('".RELATIVE_URL."refdir.php?url=\\1');\" class=\"external\">\\2</a>";
			$sourse = preg_replace ($bb, $html, $sourse);
			return $sourse;
		}
?>