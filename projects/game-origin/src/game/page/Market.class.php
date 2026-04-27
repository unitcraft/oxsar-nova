<?php
/**
* Market module.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class Market extends Page
{
	const UNLIMIT = 9999999999;
	
	public function __construct()
	{
		parent::__construct();
		$this->setPostActions(array(
			'a_metal'		=> 'marketMetal',
			'a_silicon'		=> 'marketSilicon',
			'a_hydrogen'	=> 'marketHydrogen',
			'a_credit'		=> 'marketCredit',
			'ex_metal'		=> 'Metal_ex',
			'ex_silicon'	=> 'Silicon_ex',
			'ex_hydrogen'	=> 'Hydrogen_ex',
			'ex_credit'		=> 'Credit_ex',
		));
		$this
			->addPostArg("Metal_ex", "silicon")
			->addPostArg("Metal_ex", "hydrogen")
			->addPostArg("Silicon_ex", "metal")
			->addPostArg("Silicon_ex", "hydrogen")
			->addPostArg("Hydrogen_ex", "metal")
			->addPostArg("Hydrogen_ex", "silicon")
			->addPostArg("Credit_ex", "metal")
			->addPostArg("Credit_ex", "silicon")
			->addPostArg("Credit_ex", "hydrogen");
		
		Core::getTPL()->assign("metal", NS::getPlanet()->getData("metal"));
		Core::getTPL()->assign("silicon", NS::getPlanet()->getData("silicon"));
		Core::getTPL()->assign("hydrogen", NS::getPlanet()->getData("hydrogen"));
		Core::getTPL()->assign("storageMetal", $this->storage_metal = NS::getPlanet()->getStorage("metal"));
		Core::getTPL()->assign("storageSilicon", $this->storage_silicon = NS::getPlanet()->getStorage("silicon"));
		Core::getTPL()->assign("sotrageHydrogen", $this->storage_hydrogen = NS::getPlanet()->getStorage("hydrogen"));

		$this->credit = NS::getUser()->get("credit");
		$this->comis = round((NS::getUser()->get("exchange_rate") - 1) * 100);

		Core::getTPL()->assign("credit_row", $this->credit);
		Core::getTPL()->assign("comis", $this->comis);

		$planet_ratio_data = getPlanetCreditRatio();
		$planet_ratio = $planet_ratio_data["planet_ratio"];
		$this->curs_metal = round(MARKET_BASE_CURS_METAL * $planet_ratio, -1);
		$this->curs_silicon = round(MARKET_BASE_CURS_SILICON * $planet_ratio, -1);
		$this->curs_hydrogen = round(MARKET_BASE_CURS_HYDROGEN * $planet_ratio, -1);
		$this->curs_credit = 1;

		Core::getTPL()->assign("curs_metal", $this->curs_metal);
		Core::getTPL()->assign("curs_silicon", $this->curs_silicon);
		Core::getTPL()->assign("curs_hydrogen", $this->curs_hydrogen);
		Core::getTPL()->assign("curs_credit", $this->curs_credit);
		Core::getTPL()->assign("planet_ratio", $planet_ratio);
		Core::getTPL()->assign("planet_ratio_procent", round($planet_ratio * 100));
		Core::getTPL()->assign("planet_ratio_base", $planet_ratio_data["real_planet_ratio"]);
		Core::getTPL()->assign("planet_ratio_base_procent", round($planet_ratio_data["real_planet_ratio"] * 100));
		Core::getTPL()->assign("is_planet_ratio_base", $planet_ratio > $planet_ratio_data["real_planet_ratio"]);

		$this->proceedRequest();
	}

	protected function index()
	{
		Core::getTPL()->display("market");
		return $this;
	}

	protected function marketMetal()
	{
		Core::getTPL()->display("market_metal");
		return $index;
	}
	
	protected function marketSilicon()
	{
		Core::getTPL()->display("market_silicon");
		return $index;
	}
	
	protected function marketHydrogen()
	{
		Core::getTPL()->display("market_hydrogen");
		return $index;
	}
	
	protected function marketCredit()
	{
		Core::getTPL()->assign("comis", $this->comis = 0);
		Core::getTPL()->assign("storageMetal", $this->storage_metal = self::UNLIMIT);
		Core::getTPL()->assign("storageSilicon", $this->storage_silicon = self::UNLIMIT);
		Core::getTPL()->assign("sotrageHydrogen", $this->storage_hydrogen = self::UNLIMIT);

		Core::getTPL()->display("market_credit");
		return $index;
	}

	protected function Metal_ex($silicon, $hydrogen)
	{
		$silicon = max(0, (int)trim($silicon));
		$hydrogen = max(0, (int)trim($hydrogen));

		$metal = round((($silicon * $this->curs_metal / $this->curs_silicon)+($hydrogen * $this->curs_metal / $this->curs_hydrogen))*($this->comis/100 + 1));

		($metal > NS::getPlanet()->getData("metal")) ? $metal="---" : "";

		if($silicon > 0) ($silicon + NS::getPlanet()->getData("silicon") > $this->storage_silicon /*NS::getPlanet()->getStorage("silicon")*/) ? $metal="---" : "";
		if($hydrogen > 0) ($hydrogen + NS::getPlanet()->getData("hydrogen") > $this->storage_hydrogen /*NS::getPlanet()->getStorage("hydrogen")*/) ? $metal="---" : "";

		if($metal != "---")
		{
			NS::updateUserRes(array(
				"type" => RES_UPDATE_EXCHANGE,
				"reload_planet" => false,
				"userid" => NS::getUser()->get("userid"),
				"planetid" => NS::getPlanet()->getPlanetId(),
				"metal" => - $metal,
				"silicon" => $silicon,
				"hydrogen" => $hydrogen,
				));
		}
		doHeaderRedirection("game.php/Market", false);
		return $this->index();
	}
	
	protected function Silicon_ex($metal, $hydrogen)
	{
		$metal = max(0, (int)trim($metal));
		$hydrogen = max(0, (int)trim($hydrogen));

		$silicon = round((($metal * $this->curs_silicon / $this->curs_metal)+($hydrogen * $this->curs_silicon / $this->curs_hydrogen))*($this->comis/100 + 1));

		($silicon > NS::getPlanet()->getData("silicon")) ? $silicon="---" : "";

		if($metal > 0) ($metal + NS::getPlanet()->getData("metal") > $this->storage_metal /*NS::getPlanet()->getStorage("metal")*/) ? $silicon="---" : "";
		if($hydrogen > 0) ($hydrogen + NS::getPlanet()->getData("hydrogen") > $this->storage_hydrogen /*NS::getPlanet()->getStorage("hydrogen")*/) ? $silicon="---" : "";

		if($silicon != "---")
		{
			NS::updateUserRes(array(
				"type" => RES_UPDATE_EXCHANGE,
				"reload_planet" => false,
				"userid" => NS::getUser()->get("userid"),
				"planetid" => NS::getPlanet()->getPlanetId(),
				"metal" => $metal,
				"silicon" => - $silicon,
				"hydrogen" => $hydrogen,
				));
		}
		doHeaderRedirection("game.php/Market", false);
		return $this->index();
	}

	protected function Hydrogen_ex($metal, $silicon)
	{
		$metal = max(0, (int)trim($metal));
		$silicon = max(0, (int)trim($silicon));

		$hydrogen = round((($metal * $this->curs_hydrogen / $this->curs_metal)+($silicon * $this->curs_hydrogen / $this->curs_silicon))*($this->comis/100 + 1));

		($hydrogen > NS::getPlanet()->getData("hydrogen")) ? $hydrogen="---" : "";

		if($metal > 0) ($metal + NS::getPlanet()->getData("metal") > $this->storage_metal /*NS::getPlanet()->getStorage("metal")*/) ? $hydrogen="---" : "";
		if($silicon > 0) ($silicon + NS::getPlanet()->getData("silicon") > $this->storage_silicon /*NS::getPlanet()->getStorage("silicon")*/) ? $hydrogen="---" : "";

		if($hydrogen != "---")
		{
			NS::updateUserRes(array(
				"type" => RES_UPDATE_EXCHANGE,
				"reload_planet" => false,
				"userid" => NS::getUser()->get("userid"),
				"planetid" => NS::getPlanet()->getPlanetId(),
				"metal" => $metal,
				"silicon" => $silicon,
				"hydrogen" => - $hydrogen,
				));
		}
		doHeaderRedirection("game.php/Market", false);
		return $this->index();
	}

	protected function Credit_ex($metal, $silicon, $hydrogen)
	{
		$this->comis = 0;
		$this->storage_metal = self::UNLIMIT;
		$this->storage_silicon = self::UNLIMIT;
		$this->storage_hydrogen = self::UNLIMIT;

		$metal = max(0, (int)trim($metal));
		$silicon = max(0, (int)trim($silicon));
		$hydrogen = max(0, (int)trim($hydrogen));

		$credit_metal_0 = ceil($metal * $this->curs_credit/$this->curs_metal);
		$credit_silicon_0 = ceil($silicon * $this->curs_credit/$this->curs_silicon);
		$credit_hydrogen_0 = ceil($hydrogen * $this->curs_credit/$this->curs_hydrogen);

		$credit_0 = $credit_metal_0 + $credit_silicon_0 + $credit_hydrogen_0;

		$credit = ceil($credit_0 * ($this->comis/100 + 1));

		($credit > $this->credit) ? $credit="---" : "";

		if($metal > 0) ($metal + NS::getPlanet()->getData("metal") > $this->storage_metal /*NS::getPlanet()->getStorage("metal")*/) ? $credit="---" : "";
		if($silicon > 0) ($silicon + NS::getPlanet()->getData("silicon") > $this->storage_silicon /*NS::getPlanet()->getStorage("silicon")*/) ? $credit="---" : "";
		if($hydrogen > 0) ($hydrogen + NS::getPlanet()->getData("hydrogen") > $this->storage_hydrogen /*NS::getPlanet()->getStorage("hydrogen")*/) ? $credit="---" : "";

		if($credit != "---")
		{
			// 37.8 REF-005: атомарная проверка credit перед списанием.
			// Без этого два параллельных запроса оба видели бы достаточный
			// credit-snapshot и оба списывали бы стоимость → credit в минус.
			$userid = NS::getUser()->get("userid");
			$stmt = sqlQuery(
				"UPDATE ".PREFIX."user SET credit = credit - ".sqlVal($credit)
				. " WHERE userid = ".sqlVal($userid)
				. " AND credit >= ".sqlVal($credit)
			);
			$reserved = $stmt ? $stmt->rowCount() : 0;
			if($reserved < 1)
			{
				Logger::dieMessage("MARKET_NOT_ENOUGH_CREDITS");
				doHeaderRedirection($this->main_page, false);
				return $this;
			}
			// Списание прошло атомарно — теперь добавляем ресурсы. Передаём
			// credit=0, чтобы updateUserRes не списал второй раз; ресурсы
			// (metal/silicon/hydrogen) добавляются как обычно.
			NS::updateUserRes(array(
				"type" => RES_UPDATE_EXCHANGE,
				"reload_planet" => false,
				"userid" => $userid,
				"planetid" => NS::getPlanet()->getPlanetId(),
				"metal" => $metal,
				"silicon" => $silicon,
				"hydrogen" => $hydrogen,
				"credit" => 0,
				));
			if( $credit > 0 )
			{
				new AutoMsg(
					MSG_CREDIT,
					NS::getUser()->get("userid"),
					time(),
					array(
						'credits' => $credit,
						'msg' => 'MSG_CREDIT_MARKET'
					)
				);
			}
		}
		doHeaderRedirection("game.php/Market", false);
		return $this->index();
	}

	public static function creditCost($metal, $silicon, $hydrogen)
	{
		$planet_ratio_data = getPlanetCreditRatio();
		$planet_ratio = $planet_ratio_data["planet_ratio"];
		$curs_metal = round(MARKET_BASE_CURS_METAL * $planet_ratio, -1);
		$curs_silicon = round(MARKET_BASE_CURS_SILICON * $planet_ratio, -1);
		$curs_hydrogen = round(MARKET_BASE_CURS_HYDROGEN * $planet_ratio, -1);
		$curs_credit = 1;
		//no ceil
		$credit_metal_0 = $metal * $curs_credit/$curs_metal;
		$credit_silicon_0 = $silicon * $curs_credit/$curs_silicon;
		$credit_hydrogen_0 = $hydrogen * $curs_credit/$curs_hydrogen;

		$credit_0 = $credit_metal_0 + $credit_silicon_0 + $credit_hydrogen_0;

		return $credit_0;
	}
}
?>