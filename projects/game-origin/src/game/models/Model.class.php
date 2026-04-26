<?php
/**
* Oxsar http://oxsar.ru
*
* 
*/

class Model implements IteratorAggregate, ArrayAccess
{
  protected $_fields = array();
  protected $_data = array();
  protected $_items = null;
  protected $_changedData = array();
  protected $_primaryKey = "";
  protected $_table = "";
  protected $_joins = null;
  protected $_orderBy = "";
  protected $_limit = "";
  protected $_where = null;

  public function __construct($table = "", $primaryKey = "")
  {
    $this->_joins = new Map();
    $this->_where = new Map();
    $this->_setTable($table)->_setPrimaryKey($primaryKey);
    return;
  }

  public function __toString()
  {
    return get_class($this);
  }

  public function __call($func, $args = array())
  {
    $name = $this->funcToParam(Str::substring($func, 3));
    switch(Str::substring($func, 0, 3))
    {
    case "get":
      return $this->get($name, (isset($args[0])) ? $args[0] : null);
      break;
    case "set":
      return $this->set($name, (isset($args[0])) ? $args[0] : null);
      break;
    case "has":
      return $this->exists($name);
      break;
    case "del":
      return $this->delete($name);
      break;
    default:
      throw new GenericException("Call to undefined method ".$this->__toString()."::".$func.".");
      break;
    }
    return $this;
  }

  public function __get($var)
  {
    return $this->get($this->funcToParam($var));
  }

  public function get($var = null, $default = null)
  {
    if(is_null($var))
    {
      return $this->_data;
    }
    return (isset($this->_data[$var])) ? $this->_data[$var] : $default;
  }

  public function __set($var, $value)
  {
    return $this->set($this->funcToParam($var), $value);
  }

  public function _setData($data)
  {
    $this->_data = $data;
    return $this;
  }

  public function set($var, $value = null, $setChanged = true)
  {
    if($setChanged)
    {
      $this->_changedData[$var] = 1;
    }
    $this->_data[$var] = $value;
    return $this;
  }

  public function _setTable($table)
  {
    $this->_table = $table;
    return $this;
  }

  public function setTableFields($fields)
  {
    if($fields instanceof Map)
    {
      $fields = $fields->get();
    }
    $this->_fields = $fields;
    return $this;
  }

  public function getFields()
  {
    return array_keys($this->_data);
  }

  public function _setPrimaryKey($primaryKey)
  {
    $this->_primaryKey = $primaryKey;
    return $this;
  }

  public function _getPrimaryKey()
  {
    return $this->_primaryKey;
  }

  public function _getTableFields()
  {
    if(!is_array($this->_fields) || count($this->_fields) <= 0)
    {
      return array("*");
    }
    return $this->_fields;
  }

  protected function funcToParam($name)
  {
    return strtolower(preg_replace('/(.)([A-Z])/', "$1_$2", $name));
  }

  public function __unset($var)
  {
    return $this->delete($this->funcToParam($var));
  }

  public function delete($var)
  {
    if($this->exists($var))
    {
      $this->_changedData[$var] = -1;
      unset($this->_data[$var]);
    }
    return $this;
  }

  public function __isset($var)
  {
    return $this->exists($this->funcToParam($var));
  }

  public function exists($var)
  {
    if(array_key_exists($var, $this->_data))
    {
      return true;
    }
    return false;
  }

  public function getId($default = null)
  {
    return $this->get($this->_primaryKey, $default);
  }

  public function load($resource)
  {
    if($resource instanceof Map)
    {
      $resource = $resource->get();
    }
    if(is_array($resource))
    {
      foreach($resource as $key => $value)
      {
        $this->set($key, $value, false);
      }
      return $this;
    }

    $result = sqlSelect($this->_getTable(), $this->_getTableFields(), $this->_getJoins(), $this->_getWhere($resource), $this->_orderBy, $this->_limit);
    $this->_data = sqlFetch($result);
    sqlEnd($result);
    return $this;
  }

  public function _getJoins()
  {
    return $this->_joins->toString(" ");
  }

  public function addJoin($table, $on, $type = "LEFT")
  {
    $join = $type." JOIN ".PREFIX.$table." ON (".$on.")";
    $this->_joins->set($table, $join);
    return $this;
  }

  public function removeJoin($key)
  {
    $this->_joins->set($key, null);
    return $this;
  }

  public function _getWhere($id = null)
  {
    if(!is_null($id))
    {
      return $this->_getPrimaryKey()." = '".$id."'";
    }
    if($this->_where->size() <= 0)
    {
      return "";
    }
    return $this->_where->toString(" AND ");
  }

  public function _addWhere($expression)
  {
    $this->_where->push($expression);
    return $this;
  }

  public function _getTable()
  {
    return $this->_table;
  }

  public function getIterator()
  {
    return new ArrayIterator($this->_data);
  }

  public function offsetSet($var, $value)
  {
    return $this->set($var, $value);
  }

  public function offsetExists($var)
  {
    return $this->exists($var);
  }

  public function offsetUnset($var)
  {
    return $this->delete($var);
  }

  public function offsetGet($var)
  {
    return $this->get($var);
  }

  public function _setOrderBy($orderBy)
  {
    $this->_orderBy = $orderBy;
    return $this;
  }

  public function _getOrderBy()
  {
    return $this->_orderBy;
  }

  public function _setLimit($limit)
  {
    $this->_limit = $limit;
    return $this;
  }

  public function _getLimit($default = null)
  {
    if(!empty($default) && empty($this->_limit))
    {
      return $default;
    }
    return $this->_limit;
  }

  public function getItems()
  {
    if(is_null($this->_items))
    {
      $this->_items = new Items($this);
    }
    return $this->_items;
  }

  public function save()
  {
    if($this->getId(false))
    {
      // Update data
      if($this->_joins->size() > 0)
      {
        while($this->_joins->next())
        {
          $this->update($this->_joins->key());
        }
        $this->update($this->_getTable());
      }
      else
      {
        Core::getQuery()->update($this->_getTable(), $this->_getTableFields(), $this->get(), $this->_getWhere($this->getId()));
      }
    }
    else
    {
      if($this->_joins->size() > 0)
      {
        while($this->_joins->next())
        {
          $this->insert($this->_joins->key());
        }
        $this->insert($this->_getTable());
      }
      else
      {
        Core::getQuery()->insert($this->_getTable(), $this->_getTableFields(), $this->get());
      }
    }
    return $this;
  }

  protected function update($tableKey)
  {
    $table = $this->_getRealTable($tableKey);
    $atts = $this->_getValuesFromTable($table);
    $vals = array();
    foreach($this->getFields() as $field)
    {
      if(in_array($field, $vals))
      {
        $vals[] = $this->get($field);
      }
    }
    Core::getQuery()->update($table, $atts, $vals, $this->_getWhere($this->getId()));
    return $this;
  }

  protected function insert($tableKey)
  {
    $table = $this->_getRealTable($tableKey);
    $atts = $this->_getValuesFromTable($table);
    $vals = array();
    foreach($this->getFields() as $field)
    {
      if(in_array($field, $vals))
      {
        $vals[] = $this->get($field);
      }
    }
    Core::getQuery()->insert($table, $atts, $vals);
    return $this;
  }

  public function _getRealTable($table)
  {
    $buffer = explode(" ", $table);
    return $buffer[0];
  }

  protected function _getValuesFromTable($table)
  {
    $key = explode(" ", $table);
    $key = $key[1].".";
    $values = array();
    foreach($this->_getTableFields() as $tableField)
    {
      if(Str::inString($key, $tableField))
      {
        $values[] = Str::replace($key, "", $tableField);
      }
    }
    return $values;
  }

  public function drop()
  {
    if($this->getId(false))
    {
      Core::getQuery()->delete($this->_getTable(), $this->_getWhere($this->getId()));
    }
    return $this;
  }
}
?>