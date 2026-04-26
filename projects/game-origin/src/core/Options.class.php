<?php
/**
* Sets global variables.
*
* @package Recipe 1.1
* @author Sebastian Noll
* @copyright Copyright (c) 2008, Sebastian Noll
* @license <http://www.gnu.org/licenses/gpl.txt> GNU/GPL
* @version $Id: Options.class.php 23 2010-04-03 19:08:34Z craft $
*/

if(!defined("RECIPE_ROOT_DIR")) { die("Hacking attempt detected."); }

class Options extends Collection
{
	/**
	* Enables cache function.
	*
	* @var boolean
	*/
	protected $cacheActive = false;

	/**
	* Indicates if all variables were already cached.
	*
	* @var boolean
	*/
	protected $uncached = true;

	/**
	* Configuration file.
	*
	* @var string
	*/
	protected $configFile = "Config.xml";
	
	protected $yii_way = false;

	/**
	* Constructor.
	*
	* @return void
	*/
	public function __construct( $yii_way = true )
	{
		$this->yii_way = $yii_way;
		if( !$yii_way )
		{
			if($this->cacheActive)
			{
				$this->cacheActive = CACHE_ACTIVE;
			}
			$this->item = array();
			$this->setItems();
		}
		else
		{
			$this->item = [];
//			$this->loadFromMemcache();
		}
		return;
	}
	
	protected function loadFromMemcache()
	{
		return;
		$c_name = 'Options.items';
		$c_dur	= 3600*24*3;
		$result = false;
		if( $result === false )
		{
			$configs = Config_YII::model()->findAll();
			foreach($configs as $config)
			{
				$result[ $config->var ] = $this->parseItemByTypeNew($config->value, $config->type);
			}
			$from_xml = array();
			$file = RECIPE_ROOT_DIR.$this->configFile;
			if( file_exists($file) )
			{
				$xml = new XML($file);
				$config = $xml->get()->getChildren();
				$xml->kill();
				$from_xml = $this->getFromXML($config);
			}
			$result = CMap::mergeArray($from_xml, $result);
			if( file_exists($file) )
			{
				// cache disabled
			}
			else
			{
				// cache disabled
			}
		}
		$this->item = $result;
		return $this;
	}
	
	protected function generateYiiCode()
	{
		$result = '';
		foreach( $this->item as $key => $item )
		{
			if( is_array($item) )
			{
				$result .= "\r\n" . (is_numeric($key) ? "$key" : "'$key'" ) . " => array(";
				foreach ( $item as $k => $v )
				{
					if( is_array($v) )
					{
						$result .= "\r\n\t" . (is_numeric($k) ? "$k" : "'$k'" ) . " => array(";
						foreach ( $v as $k1 => $v1 )
						{
							$result .= "\r\n\t\t" . (is_numeric($k1) ? "$k1" : "'$k1'" ) . " => " . (is_numeric($v1) ? "$v1" : "'$v1'" ) . ",";
						}
						$result .= "\r\n\t),";
					}
					else
					{
						$result .= "\r\n\t" . (is_numeric($k) ? "$k" : "'$k'" ) . " => " . (is_numeric($v) ? "$v" : "'$v'" ) . ",";
					}
				}
				$result .= "\r\n),";
			}
			else
			{
				$result .= "\r\n" . (is_numeric($key) ? "$key" : "'$key'" ) . " => " . (is_numeric($item) ? "$item" : "'$item'" ) . ",";
			}
		}
		return $result;
	}

	/**
	* Sets options.
	*
	* @return Options
	*/
	protected function setItems()
	{
		if( $this->yii_way )
		{
			return $this->loadFromMemcache();
		}
		
		if($this->cacheActive)
		{
			$this->item = array_merge($this->item, Core::getCache()->getConfigCache());
		}
		else
		{
			$result = Core::getDB()->query("SELECT var, value, type FROM `" . PREFIX . "config`");
			while($row = Core::getDB()->fetch($result))
			{
				$this->item[$row["var"]] = $this->parseItemByTypeNew($row["value"], $row["type"]);
			}
		}
		return $this;
	}

	/**
	* Changes a configuration parameter.
	*
	* @param string	Variable name
	* @param string	New value of the variable
	* @param boolean	Turn off auto-caching
	*
	* @return Options
	*/
	public function setValue($var, $value, $renewcache = false)
	{
		if($this->hasVariable($var, false))
		{
			Core::getQuery()->update("config", "value", $value, "var = ".sqlVal($var));
		}
		else
		{
			$att = array("var", "value", "type", "groupid", "islisted");
			$val = array($var, $value, "char", "1", "1");
			Core::getQuery()->insert("config", $att, $val);
		}
		if($this->cacheActive && $renewcache)
		{
			Core::getCache()->buildConfigCache();
		}
		return $this;
	}

	public function setValueNew($var, $value, $renewcache = false)
	{
		return $this;
	}

	/**
	* Checks if variable exists.
	*
	* @param string	Variable name
	* @param boolean	Turn off auto-caching
	*
	* @return boolean	True, if variable exists, flase, if not
	*/
	protected function hasVariable($var, $renewcache = true)
	{
		if($this->exists($var)) { return true; }
		if($this->cacheActive && $renewcache && $this->uncached)
		{
			Core::getCache()->buildConfigCache();
			$this->item = Core::getCache()->getConfigCache();
			$this->uncached = false;
			return $this->hasVariable($var, false);
		}
		return false;
	}

	/**
	* Returns a configuration parameter.
	*
	* @param string	Variable name
	*
	* @return mixed	Parameter content
	*/
	public function get($var)
	{
		if( $this->yii_way )
		{
			if( isset($this->item[$var]) )
			{
				return $this->item[$var];
			}
			else
			{
				return $var;
			}
		}
		if($this->hasVariable($var)) { return $this->item[$var]; }
		return $var;
	}

	/**
	* Changes a configuration parameter.
	*
	* @param string	Variable name
	* @param string	New value of the variable
	*
	* @return Options
	*/
	public function set($var, $value)
	{
		if( $this->yii_way )
		{
			return $this->setValueNew($var, $value);
		}
		return $this->setValue($var, $value);
	}

	/**
	* Gets the data from the config file.
	* The loaded configuration can be overwritten by the database
	* entries.
	*
	* @return Options
	*/
	protected function loadConfigFile()
	{
		$file = RECIPE_ROOT_DIR.$this->configFile;
		if( file_exists($file) )
		{
			$xml = new XML($file);
			$config = $xml->get()->getChildren();
			$xml->kill();
			$this->item = $this->getFromXML($config);
		}
		return $this;
	}

	/**
	* Parses XML data.
	*
	* @param SimpleXMLElement	XML to parse
	*
	* @return array			parsed XML
	*/
	protected function getFromXML(SimpleXMLElement $xml)
	{
		$item = array();
		foreach($xml as $index => $child)
		{
			$item[$index] = $this->parseItemByType($child, $child->getAttribute("type"));
		}
		return $item;
	}

	/**
	* Parses an item by the given type.
	*
	* @param XMLObj	Item
	* @param string	Type
	*
	* @return mixed	Parsed item
	*/
	protected function parseItemByType(XMLObj $item, $type = null)
	{
		switch($type)
		{
			case "level":
				return $this->getFromXML($item);
				break;
			case "array":
				return $item->getArray();
				break;
			case "integer":
			case "int":
				return $item->getInteger();
				break;
			case "string":
			case "char":
			case "text":
			default:
				return $item->getString();
				break;
			case "bool":
			case "boolean":
				return $item->getBooleanObj();
				$item = (bool) $item;
				break;
			case "dbquery":
				return Str::replace("PREFIX", PREFIX, $item->getString());
				break;
			case "map":
				return $item->getMap();
				break;
		}
		return $item;
	}
	
	protected function parseItemByTypeNew( $item, $type = null )
	{
		if( is_numeric($item) )
		{
			return floatval($item);
		}
		
		switch($type)
		{
			case "integer":
			case "int":
			case "float":
			case "double":
				return floatval($item);
				break;
			case "string":
			case "char":
			case "text":
			default:
				return strval($item);
				break;
			case "bool":
			case "boolean":
				return ((bool) $item);
				break;
		}
		return $item;
	}
}
?>