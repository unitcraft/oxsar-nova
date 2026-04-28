<?php
/**
* Oxsar http://oxsar.ru
*
* 
*/

class Unit extends Structure
{
  public function __construct($id = null)
  {
    parent::__construct("construction b", "buildingid");
    $this->setTableFields(array(
      "b.buildingid", "b.mode", "b.name", "b.basic_metal", "b.basic_silicon", "b.basic_hydrogen", "b.basic_energy", "b.display_order",
      "d.unitid", "d.capicity", "d.speed", "d.consume", "d.attack", "d.shield",
      "e.engineid", "e.level", "e.base_speed"
      ))
      ->addJoin("ship_datasheet d", "b.buildingid = d.unitid")
      ->addJoin("ship2engine e", "b.buildingid = e.unitid")
      ->_setOrderBy("b.display_order ASC, b.buildingid ASC")
      ->_addWhere("(b.mode = '".self::SHIP_MODE."' OR b.mode = '".self::DEFENSE_MODE."')");
    if(!is_null($id))
    {
      $this->_setLimit(1);
      $this->load($id);
    }
    return;
  }

  public function getMetalCost($qty = 1, $format = false)
  {
    $cost = $this->getBasicMetal() * $qty;
    return ($format) ? fNumber($cost) : $cost;
  }

  public function getSiliconCost($qty = 1)
  {
    $cost = $this->getBasicSilicon() * $qty;
    return ($format) ? fNumber($cost) : $cost;
  }

  public function getHydrogenCost($qty = 1)
  {
    $cost = $this->getBasicHydrogen() * $qty;
    return ($format) ? fNumber($cost) : $cost;
  }

  public function getEnergyCost($qty = 1)
  {
    $cost = $this->getBasicEnergy() * $qty;
    return ($format) ? fNumber($cost) : $cost;
  }

  public function getHull($format = false)
  {
    if(!$this->exists("hull"))
    {
      $hull = ($this->getBasicMetal() + $this->getBasicSilicon()) / 10;
      $this->set("hull", $hull, false);
    }
    return ($format) ? fNumber($this->get("hull")) : $this->get("hull");
  }

  public function getItems()
  {
    return parent::getItems()->setGroupBy("b.buildingid");
  }
}
?>