<?php
/**
* Extended functionality of the SimpleXMLElement.
*
* @package Recipe 1.1
* @author Sebastian Noll
* @copyright Copyright (c) 2008, Sebastian Noll
* @license <http://www.gnu.org/licenses/gpl.txt> GNU/GPL
* @version $Id: XMLObj.util.class.php 23 2010-04-03 19:08:34Z craft $
*/

if(!defined("RECIPE_ROOT_DIR")) { die("Hacking attempt detected."); }

class XMLObj extends SimpleXMLElement
{
  /**
  * Returns the children of a XML element.
  *
  * @param string	The children's name [optional]
  *
  * @return XMLObj	The children
  */
  public function getChildren($name = null)
  {
    if(is_null($name))
    {
      return $this->children();
    }
    return $this->$name;
  }

  /**
  * Returns the attributes of a XML element.
  *
  * @param string	Attribute name [optional]
  *
  * @return string	Attribute content
  */
  public function getAttribute($name = null)
  {
    if(is_null($name))
    {
      return $this->attributes();
    }
    return strval($this->attributes()->$name);
  }

  /**
  * Returns the data of a XML element.
  *
  * @param string	Element name [optional]
  * @param string	Data type [optional]
  *
  * @return mixed	Data
  */
  public function getData($name = null, $type = null)
  {
    if(!is_null($type))
    {
      $type = strtolower($type);
      switch($type)
      {
      default: case "string": case "text": case "char": case "str":
        return $this->getString($name);
        break;
      case "int": case "integer":
        return $this->getInteger($name);
        break;
      case "bool": case "boolean":
        return $this->getBoolean($name);
        break;
      case "float": case "double":
        return $this->getFloat($name);
        break;
      case "array": case "arr":
        return $this->getArray($name);
        break;
      case "map":
        return $this->getMap($name);
        break;
      }
    }
    return $this->getString($name);
  }

  /**
  * Returns the data of the XML element as a string.
  *
  * @param string	Element name [optional]
  *
  * @return string
  */
  public function getString($name = null)
  {
    if(is_null($name))
    {
      return strval($this);
    }
    return strval($this->$name);
  }

  /**
  * Returns the data of the XML element as a string.
  *
  * @param string	Element name [optional]
  *
  * @return string
  */
  public function getStringObj($name = null)
  {
    if(is_null($name))
    {
      return new OxsarString($this);
    }
    return new OxsarString($this->$name);
  }

  /**
  * Returns the data of the XML element as an integer value.
  *
  * @param string	Element name [optional]
  *
  * @return integer
  */
  public function getInteger($name = null)
  {
    if(is_null($name))
    {
      return intval($this);
    }
    return intval($this->$name);
  }

  /**
  * Returns the data of the XML element as an integer value.
  *
  * @param string	Element name [optional]
  *
  * @return integer
  */
  public function getIntegerObj($name = null)
  {
    if(is_null($name))
    {
      return new Integer($this);
    }
    return new Integer($this->$name);
  }

  /**
  * Returns the data of the XML element as a boolean value.
  *
  * @param string	Element name [optional]
  *
  * @return boolean
  */
  public function getBoolean($name = null)
  {
    if(is_null($name))
    {
      return (bool) $this;
    }
    return (bool) $this->$name;
  }

  /**
  * Returns the data of the XML element as a boolean value.
  *
  * @param string	Element name [optional]
  *
  * @return boolean
  */
  public function getBooleanObj($name = null)
  {
    if(is_null($name))
    {
      return new Boolean($this);
    }
    return new Boolean($this->$name);
  }

  /**
  * Returns the data of the XML element as a float value.
  *
  * @param string	Element name [optional]
  *
  * @return float
  */
  public function getFloat($name = null)
  {
    if(is_null($name))
    {
      return floatval($this);
    }
    return floatval($this->$name);
  }

  /**
  * Returns the data of the XML element as a float object.
  *
  * @param string	Element name [optional]
  *
  * @return Float
  */
  public function getFloatObj($name = null)
  {
    if(is_null($name))
    {
      return new Float($this);
    }
    return new Float($this->$name);
  }

  /**
  * Returns the data of the XML element as an array.
  *
  * @param string	Element name [optional]
  *
  * @return array
  */
  public function getArray($name = null)
  {
    return Arr::trim(explode(",", $this->getString($name)));
  }

  /**
  * Returns the data of the XML element as a map.
  *
  * @param string	Element name [optional]
  *
  * @return Map
  */
  public function getMap($name = null)
  {
    $map = new Map($this->getArray($name));
    return $map->trim();
  }
}
?>