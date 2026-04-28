<?php
/**
* Oxsar http://oxsar.ru
*
* 
*/

class Items implements Countable, IteratorAggregate
{
  protected $items = array();
  protected $iterator = 0;
  protected $loaded = false;
  protected $name = "Model";
  protected $model = null;
  protected $groupBy = "";

  public function __construct(Model $model = null)
  {
    if(!is_null($model))
    {
      $this->setModel($model);
    }
    return;
  }

  public function setModel(Model $model)
  {
    $this->model = $model;
    $this->setName($model->__toString());
    return $this;
  }

  public function getModel()
  {
    return $this->model;
  }

  public function load()
  {
    if(!$this->isLoaded())
    {
      $this->loaded = true;
      $table = $this->getModel()->_getTable();
      $attributes = $this->getModel()->_getTableFields();
      $joins = $this->getModel()->_getJoins();
      $where = $this->getModel()->_getWhere();
      $orderBy = $this->getModel()->_getOrderBy();
      $limit = $this->getModel()->_getLimit();
      $result = sqlSelect($table, $attributes, $joins, $where, $orderBy, $limit, $this->groupBy);
      while(($row = sqlFetch($result)))
      {
        $item = new $this->name($table, $this->getModel()->_getPrimaryKey());
        $item->_setData($row);
        $this->items[] = $item;
      }
    }
    return $this;
  }

  public function size()
  {
    return $this->count();
  }

  public function get()
  {
    $this->load();
    return $this->items;
  }

  public function isLoaded()
  {
    return $this->loaded;
  }

  public function clear()
  {
    $this->rewind();
    $this->items = array();
    $this->loaded = false;
    return $this;
  }

  public function getLastItem()
  {
    $this->load();
    if($this->count() > 0)
    {
      return end($this->items);
    }
    return new $this->name();
  }

  public function getFirstItem()
  {
    $this->load();
    if($this->count() > 0 && isset($this->items[0]))
    {
      return $this->items[0];
    }
    return new $this->name();
  }

  public function getRandomItem()
  {
    $this->load();
    if($this->count() > 0)
    {
      $rand = mt_rand(0, $this->count() - 1);
      if(isset($this->items[$rand]))
      {
        return $this->items[$rand];
      }
    }
    return new $this->name();
  }

  public function setName($name)
  {
    $this->name = $name;
    return $this;
  }

  public function getName()
  {
    return $this->name;
  }

  public function setGroupBy($groupBy)
  {
    $this->groupBy = $groupBy;
    return $this;
  }

  public function getIterator()
  {
    $this->load();
    return new ArrayIterator($this->items);
  }

  public function count()
  {
    $this->load();
    return count($this->items);
  }
}
?>