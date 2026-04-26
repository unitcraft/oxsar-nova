<?php
/**
* Referral module.
*
* Oxsar http://oxsar.ru
*
*
*/

class Referral extends Page
{
	public function __construct()
	{
		parent::__construct();
		$this->proceedRequest();
	}
	
	protected function index()
	{
		$max_bonus_points = REFERRAL_MAX_BONUS_POINTS;
		$ref_bonus_points = REFERRAL_BONUS_POINTS;

		$userid = NS::getUser()->get("userid");

		$rt = sqlSelect("planet", array("planetid", "metal", "silicon", "hydrogen"), "", "userid = ".sqlVal($userid), "planetid ASC", "1");
		$res = sqlFetch($rt);

		$points = NS::getUser()->get("points");
		$is_bonus_active = true;
		if($points > $max_bonus_points)
		{
			$is_bonus_active = false;
			$points = 0;
		}

		$s_metal 	= REFERRAL_METAL_BONUS;
		$s_silicon 	= REFERRAL_SILICON_BONUS;
		$s_hydrogen = REFERRAL_HYDROGEN_BONUS;
		$w_metal	= $points * $s_metal;
		$w_silicon	= $points * $s_silicon;
		$w_hydrogen	= $points * $s_hydrogen;
		

		Core::getTPL()->assign("max_bonus_points", fNumber($max_bonus_points));
		Core::getTPL()->assign("ref_bonus_points", fNumber($ref_bonus_points));
		Core::getTPL()->assign("is_bonus_active", $is_bonus_active);
		Core::getTPL()->assign("p_metal", $s_metal * 100);
		Core::getTPL()->assign("p_silicon", $s_silicon * 100);
		Core::getTPL()->assign("p_hydrogen", $s_hydrogen * 100);
		Core::getTPL()->assign("w_Metal", fNumber($w_metal));
		Core::getTPL()->assign("w_Silicon", fNumber($w_silicon));
		Core::getTPL()->assign("w_Hydrogen", fNumber($w_hydrogen));

		$is_updated = false;
		
		$referral = array();
		$result = sqlSelect(
			"referral r",
			array("r.ref_time", "r.ref_id", "r.bonus", "r.bonus_credit", "u.username", "u.points"),
			"LEFT JOIN ".PREFIX."user u ON (u.userid = r.ref_id)",
			"r.userid = ".sqlVal($userid), "ref_time DESC"
		);
		while($row = sqlFetch($result))
		{
			if($row["bonus"] == 0)
			{
				$ref_points = sqlSelectField("user", "points", "", "userid = ".sqlVal($row["ref_id"]));
				if($ref_points >= $ref_bonus_points)
				{
			if( !NS::isFirstRun("Referral::CalculateBonus:{$userid}-{$row["ref_id"]}") )
			{
			doHeaderRedirection("game.php/Referral", false);
				return;
			}
					$row["bonus"] = $is_bonus_active ? 1 : 2;
					
					if($is_bonus_active)
					{
						sqlUpdate(
							"referral",
							array(
								"bonus"			 => $row["bonus"],
								"bonus_time"	 => time(),
								"bonus_metal"	 => $w_metal,
								"bonus_silicon"	 => $w_silicon,
								"bonus_hydrogen" => $w_hydrogen,
							),
							"userid = ".sqlVal($userid)." AND ref_id = ".sqlVal($row["ref_id"])
								. ' ORDER BY ref_id'
						);
		
						sqlUpdate(
					 		"planet",
					 		array(
					 			"metal" 	=> $res["metal"] + $w_metal,
					 			"silicon" 	=> $res["silicon"] + $w_silicon,
					 			"hydrogen" 	=> $res["hydrogen"] + $w_hydrogen,
	//				 			"last" 		=> time()+1, // Field last is used in recounting planet resourses
					 		),
				 			"planetid = ".sqlVal($res["planetid"])
					 			. ' ORDER BY planetid'
						);
						$is_updated = true;
					}
				}
			}
			$row["ref_time"]		= Date::timeToString(1, $row["ref_time"]);
			$row["points"]		= fNumber($row["points"]);
			$row["bonus_credit"]	= fNumber($row["bonus_credit"], 2);
			
			switch($row["bonus"])
			{
				case 0: $row["bonus_img"] = Image::getImage("ref-.jpg", ""); break;
				case 1: $row["bonus_img"] = Image::getImage("ref+.jpg", ""); break;
				case 2: $row["bonus_img"] = Image::getImage("ref+off.jpg", ""); break;
			}
			$referral[] = $row;
		}
		if($is_updated)
		{
			doHeaderRedirection("game.php/".Core::getRequest()->getGET("go"), false);
		}

		Core::getTPL()->addLoop("shoutReferral", $referral);
		Core::getTPL()->display("referral");
		return $this->quit();
	}
}
?>