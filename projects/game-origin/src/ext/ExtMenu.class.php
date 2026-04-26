<?php
/**
* ExtMenu class: Generates the menu by XML.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class ExtMenu extends Menu
{
	/**
	* Generates the menu items.
	*
	* @return Menu
	*/
	protected function generateMenu()
	{
		$this->menu	= array();
		$temp_menu	= array();
		$output_menu= array();
		$was_div	= false;
		$i			= 0;
		
		$msgs		= sqlSelectField("message", "count(*)", "", "receiver = ".sqlUser()." AND readed = '0'");
		$friends	= sqlSelectField("buddylist", "count(*)", "", "friend2 = ".sqlUser()." AND accepted = '0'");
		
		$allyapplications = 0;
		if(NS::getUser()->get("aid")) // && sqlSelectField("alliance", "founder", "", "aid"=.sqlVal(NS::getUser()->get("aid"))) == NS::getUser()->get("userid"))
		{
			$allyapplications = sqlSelectField("allyapplication", "count(*)", "", "aid = ".sqlVal(NS::getUser()->get("aid")));
		}
		
		foreach( $this->xml as $first )
		{
			$i++;
			$inner		= $this->getClass($first);
			$outer_bot	= "";
			$title_min	= trim($first->getAttribute("title-min"));
			$class_max	= trim($first->getAttribute("class-max"));
			$class_min	= trim($first->getAttribute("class-min"));
			
			if ($class_max != "")
			{
				if ($inner != "")
				{
					$inner = preg_replace("#class=\"([^\"]+)\"#i", "class=\"$1 $class_max\"", $inner);
				}
				else
				{
					$inner = "class=\"$class_max\"";
				}
			}
			$k = 0;
			$temp_menu[$i] = array(
//				"outer_bot"	=> $outer_bot,
				"inner"		=> $inner,
//				"content"	=> $this->getLabel($first),
//				"direct"	=> "",
				'title'		=> $title_min,
			);
			foreach( $first->getChildren() as $second )
			{
				$k++;
				
				// NS::getPlanetStack();
				
				if (
					$second->getAttribute("href") === "ExchangeOpts"
						&& NS::getPlanet()->getBuilding(UNIT_EXCHANGE) < 1
				)
				{
					continue;
				}

				if ( $second->getAttribute("href") === "MSG" && $msgs > 0 )
				{
					$content = $this->getLink($second, " ($msgs)");
				}
				else if ( $second->getAttribute("href") === "Friends" && $friends > 0 )
				{
					$content = $this->getLink($second, " ($friends)");
				}
				else if ($second->getAttribute("href") === "Alliance" && $allyapplications > 0)
				{
					$content = $this->getLink($second, " ($allyapplications)");
				}
				else
				{
					$content = $this->getLink($second);
				}
				
				$temp_menu[$i]['childs'][] = array(
					"inner" => $this->getClass($second),
					"content" => $content,
					"direct" => $this->getDirectLink($second)
				);
			}
			$temp_menu[$i]['nmb_chlds']	= $k;
		}
		$i = 0;
		$k = 0;
		$temp = array();
		foreach( $temp_menu as $menu_item)
		{
			if( ($k + $menu_item['nmb_chlds']) <= MAX_MENU_ITEMS )
			{
				$k += $menu_item['nmb_chlds'];
			}
			else
			{
				$k = $menu_item['nmb_chlds'];
				if($temp)
				{
					$output_menu[$i]['items'] = $temp;
					$temp = array();
					$i++;
				}
			}
			$temp[] = $menu_item;
		}
		if($temp)
		{
			$output_menu[$i]['items'] = $temp;
		}
		$this->menu = $output_menu;
//		print_r($output_menu);
//		if ( $was_div )
//		{
//			$temp_menu[count($temp_menu) -1]["outer_bot"] = "</ul></li>";
//		}
		return $this;
	}
}
?>