<?php
/**
* Extension of PlanetCreator class.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class ExpedPlanetCreator extends PlanetCreator
{
	/**
	* Metal scrapts that are needed to be placed on orbit.
	*
	* @var integer
	*/
	protected $metal_scraps;

	/**
	* Silicon scrapts that are needed to be placed on orbit.
	*
	* @var integer
	*/
	protected $silicon_scraps;
	
	/**
	* Starts creator.
	*
	* @param integer Owner user ID
	* @param integer Data, from witch it'll create planet or moon
	*
	* @return void
	*/
	public function __construct( $user_id, $data = array())
	{
		$this->userid	= $user_id;
		$this->moon		= ( (isset($data['ismoon'])) ? $data['ismoon'] : false );
		$this->percent	= ( (isset($data['percent'])) ? $data['percent'] : 0 );

		$this->metal	= ( (isset($data['metal'])) ? $data['metal'] : 0 );
		$this->silicon	= ( (isset($data['silicon'])) ? $data['silicon'] : 0 );
		$this->hydrogen = ( (isset($data['hydrogen'])) ? $data['hydrogen'] : 0 );
		
		$this->metal_scraps		= ( (isset($data['metal_scraps'])) ? $data['metal_scraps'] : 0 );
		$this->silicon_scraps	= ( (isset($data['silicon_scraps'])) ? $data['silicon_scraps'] : 0 );
		if( !isset($data['galaxy']) )
		{
			$this->setRandPos();
		}
		else
		{
			$this->galaxy = ( (isset($data['galaxy'])) ? $data['galaxy'] : 0 );
			$this->system = ( (isset($data['system'])) ? $data['system'] : 0 );
			$this->position = ( (isset($data['position'])) ? $data['position'] : 0 );
		}
		
		$this->name	= ( (isset($data['name'])) ? $data['name'] : Core::getLanguage()->getItem("HOME_PLANET") );
		
		if( isset($data['size']) )
		{
			$this->size = $data['size'];
		}
		elseif( $this->moon )
		{
			$this->setRandMoonSize();
		}
		else
		{
			$this->setRandSize();
		}

		$this->setRandTemperature();
		
		if( !$this->moon )
		{
			$this->setPicture();
		}
		
		$this->createPlanet();
		return;
	}
	
	/**
	* Write final planet informations into database.
	*
	* @return ExpedPlanetCreator
	*/
	protected function createPlanet()
	{
		if($this->moon == 0)
		{
			$this->planetid = sqlInsert(
				'planet',
				array(
					'userid'				=> $this->userid,
					'planetname'			=> $this->name,
					'diameter'				=> $this->size,
					'picture'				=> $this->picture,
					'temperature'			=> $this->temperature,
					'last'					=> time(),
					'metal'					=> $this->metal,
					'silicon'				=> $this->silicon,
					'hydrogen'				=> $this->hydrogen,
					'solar_satellite_prod'	=> 100,
				)
			);
			sqlInsert(
				'galaxy',
				array(
					'galaxy'	=> $this->galaxy,
					'system'	=> $this->system,
					'position'	=> $this->position,
					'planetid'	=> $this->planetid,
					'metal'		=> $this->metal_scraps,
					'silicon'	=> $this->silicon_scraps,
				)
			);
		}
		else
		{
			$this->planetid = sqlInsert(
				'planet',
				array(
					'userid'		=> $this->userid,
					'ismoon'		=> 1,
					'planetname'	=> $this->name,
					'diameter'		=> $this->size,
					'picture'		=> 'mond',
					'temperature'	=> $this->temperature - mt_rand(15, 35),
					'last'			=> time(),
					'metal'			=> 0,
					'silicon'		=> 0,
				)
			);
			sqlQuery(
				'UPDATE ' . PREFIX . 'galaxy SET '.
					'(moonid = ' . $this->planetid .
					', silicon = silicon + ' . $this->silicon_scraps .
					', metal = metal + ' . $this->metal_scraps .
				') WHERE ' . 'galaxy = '.sqlVal($this->galaxy).
					' AND system = '.sqlVal($this->system).
					' AND position = '.sqlVal($this->position)
				. ' ORDER BY galaxy DESC, system DESC, position DESC '
			);
		}
		return $this;
	}
}