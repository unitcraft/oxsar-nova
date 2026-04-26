<?php
/**
* Handles caching using memcache.
*
* Oxsar http://oxsar.ru
*
*
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class MemCacheHandler
{
	protected $memcache = null;
	protected $yii_way = true;

	public function __construct()
	{
		if( true )
		{
			$this->yii_way = false;
			if(class_exists("Memcache", false))
			{
				$memcache = new memcache();
				for($i = 0; $i < 3; $i++)
				{
					if($memcache->pconnect(MC_SERVER, MC_PORT))
					{
						$this->memcache = $memcache;
						break;
					}
					usleep(100);
				}
			}
			return ;
		}
	}

	public function clear($name)
	{
		if( $this->yii_way )
		{
			// cache disabled
		}
		elseif(isset($this->memcache))
		{
			$this->memcache->delete($name, 0);
		}
	}

	public function get($name, &$value)
	{
		if( $this->yii_way )
		{
			$value = false;
			return $value !== false;
		}
		elseif(isset($this->memcache))
		{
			$value = $this->memcache->get($name);
			return $value !== false;
		}
		$value = null;
		return false;
	}

	public function set($name, &$value, $expire = 3600)
	{
		if( $this->yii_way )
		{
			$value = // cache disabled
		}
		elseif(isset($this->memcache))
		{
			$this->memcache->set($name, $value, MEMCACHE_COMPRESSED, $expire);
		}
	}

	public function add($name, &$value, $expire = 3600)
	{
		if( $this->yii_way )
		{
			return // cache disabled
		}
		elseif(isset($this->memcache))
		{
			return $this->memcache->add($name, $value, MEMCACHE_COMPRESSED, $expire);
		}
		return false;
	}
	
	public function is_valid()
	{
		if( $this->yii_way )
		{
			return false;
		}
		else
		{
			return isset($this->memcache);
		}
	}
}
?>