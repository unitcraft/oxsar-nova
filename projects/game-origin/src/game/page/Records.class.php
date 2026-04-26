<?php
/**
* Shows all available constructions and their requirements.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class Records extends Page
{
	/**
	* Main method to display records.
	*
	* @return void
	*/
	public function __construct()
	{
		Core::getLanguage()->load("info,buildings,Statistics");
		$this->proceedRequest();
		return;
	}

	/**
	* Index action.
	*
	* @return Records
	*/
	protected function index()
	{
		$c_name = 'Records.index'; // .( (defined('SN')) ? ('SN') : ('Original') );
		$c_dur	= 60*15;
		$stats	= false;
		if( $stats === false )
		{
			$query=array(
				"SELECT con.name,con.buildingid,con.mode,usr.username,stats.max_level FROM ((".
				PREFIX."view_max_building_stats stats INNER JOIN (SELECT buildingid,max(max_level) ml FROM ".
				PREFIX."view_max_building_stats stats GROUP BY buildingid) m ON stats.buildingid=m.buildingid AND stats.max_level=m.ml) INNER JOIN ".
				PREFIX."construction con ON stats.buildingid=con.buildingid) INNER JOIN ".
				PREFIX."user usr ON stats.userid=usr.userid GROUP BY con.display_order, con.buildingid",
				"SELECT con.name,con.buildingid,con.mode,usr.username,stats.sum_quantity FROM ((".
				PREFIX."view_sum_unit_stats stats INNER JOIN (SELECT unitid,max(sum_quantity) mq FROM ".
				PREFIX."view_sum_unit_stats stats GROUP BY unitid) m ON stats.unitid=m.unitid AND stats.sum_quantity=m.mq) INNER JOIN ".
				PREFIX."construction con ON stats.unitid=con.buildingid) INNER JOIN ".
				PREFIX."user usr ON stats.userid=usr.userid GROUP BY con.display_order, con.buildingid",
			"SELECT con.name,con.buildingid,con.mode,usr.username,stats.level FROM ((".
				PREFIX."research2user stats INNER JOIN (SELECT buildingid,max(level) ml FROM ".
				PREFIX."research2user stats GROUP BY buildingid) m ON stats.buildingid=m.buildingid AND stats.level=m.ml) INNER JOIN ".
				PREFIX."construction con ON stats.buildingid=con.buildingid) INNER JOIN ".
				PREFIX."user usr ON stats.userid=usr.userid LEFT JOIN na_ban_u on na_ban_u.userid = usr.`userid` where na_ban_u.userid is null GROUP BY con.display_order, con.buildingid"
				);
			$stats = array(); //0 - constructions, 1 - moon, 2 - ships, 3 - defense, 4 - research
			for( $i = 0; $i < 3; $i++ )
			{
                $stats[$i] = array();
				$result = sqlQuery($query[$i]);
				while($row = sqlFetch($result))
				{
					$stats[$i][] = $row;
				}
				sqlEnd($result);
			}
			// cache disabled
		}

		$rows_data = (array)$stats;

		$mode = array(
			array(1,5),
			array(3,4),
			array(2,0)
			);
		$record=array(
			"max_level",
			"sum_quantity",
			"level"
			);

		$stats = array(); //0 - constructions, 1 - moon, 2 - ships, 3 - defense, 4 - research
		for( $i = 0; $i < 3; $i++ )
		{
			foreach($rows_data[$i] as $row)
			{
				$name = Core::getLanguage()->getItem($row["name"]);
				$image = Image::getImage(getUnitImage($row["name"]), $name, 60);

				$item = array();
				$item["player"] = $row["username"];
				$item["record"] = $row[$record[$i]];

				switch($row["mode"])
				{
					case UNIT_TYPE_CONSTRUCTION:
					case UNIT_TYPE_MOON_CONSTRUCTION:
						$item["name"] = Link::get("game.php/ConstructionInfo/".$row["buildingid"], $name);
						$item["image"] = Link::get("game.php/ConstructionInfo/".$row["buildingid"], $image);
						break;

					case UNIT_TYPE_RESEARCH:
						$item["name"] = Link::get("game.php/ResearchInfo/".$row["buildingid"], $name);
						$item["image"] = Link::get("game.php/ResearchInfo/".$row["buildingid"], $image);
						break;

					case UNIT_TYPE_FLEET:
					case UNIT_TYPE_DEFENSE:
						$item["name"] = Link::get("game.php/UnitInfo/".$row["buildingid"], $name);
						$item["image"] = Link::get("game.php/UnitInfo/".$row["buildingid"], $image);
						break;
				}

				switch($row["mode"])
				{
					case $mode[$i][0]:
						$stats[$i*2][] = $item;
						break;

					case $mode[$i][1]:
						$stats[$i*2+1][] = $item;
						break;
				}
			}
		}

		Core::getTPL()->addLoop("construction", $stats[0]);
		Core::getTPL()->addLoop("moon", $stats[1]);
		Core::getTPL()->addLoop("shipyard", $stats[2]);
		Core::getTPL()->addLoop("defense", $stats[3]);
		Core::getTPL()->addLoop("research", $stats[4]);
		Core::getTPL()->display("records");
		return $this;
	}
}
?>