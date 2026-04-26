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
		$this->yii_way = false;
		// План 37.5c.1: PECL memcache не поддерживается для PHP 8.3+,
		// используем новое API `class Memcached` (с 'd' на конце).
		// Старое `class Memcache` оставлено как fallback для совместимости.
		if(class_exists("Memcached", false))
		{
			$memcache = new Memcached();
			$memcache->addServer(MC_SERVER, MC_PORT);
			// Верифицируем соединение через getStats
			$stats = @$memcache->getStats();
			if(is_array($stats) && !empty($stats))
			{
				$this->memcache = $memcache;
			}
		}
		elseif(class_exists("Memcache", false))
		{
			$memcache = new Memcache();
			for($i = 0; $i < 3; $i++)
			{
				if(@$memcache->pconnect(MC_SERVER, MC_PORT))
				{
					$this->memcache = $memcache;
					break;
				}
				usleep(100);
			}
		}
	}

	public function clear($name)
	{
		if(isset($this->memcache))
		{
			$this->memcache->delete($name, 0);
		}
	}

	public function get($name, &$value)
	{
		if(isset($this->memcache))
		{
			$value = $this->memcache->get($name);
			return $value !== false;
		}
		$value = null;
		return false;
	}

	public function set($name, &$value, $expire = 3600)
	{
		if(isset($this->memcache))
		{
			// Memcached::set($key, $value, $expire) — без флагов compression
			// (включается через ->setOption(Memcached::OPT_COMPRESSION, true)).
			// Старое API Memcache::set имело флаги вторым параметром (MEMCACHE_COMPRESSED).
			if($this->memcache instanceof Memcached)
			{
				$this->memcache->set($name, $value, $expire);
			}
			else
			{
				$this->memcache->set($name, $value, MEMCACHE_COMPRESSED, $expire);
			}
		}
	}

	public function add($name, &$value, $expire = 3600)
	{
		if(isset($this->memcache))
		{
			if($this->memcache instanceof Memcached)
			{
				return $this->memcache->add($name, $value, $expire);
			}
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