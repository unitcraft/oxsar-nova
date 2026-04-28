<?php
/**
* Oxsar http://oxsar.ru
*
* 
*/

abstract class Structure extends Model
{
  const BUILDING_MODE = 1;
  const RESEARCH_MODE = 2;
  const SHIP_MODE = 3;
  const DEFENSE_MODE = 4;

  public function getName()
  {
    if(!$this->exists("translation"))
    {
      $this->setTranslation();
    }
    return $this->get("translation");
  }

  public function setTranslation()
  {
    $this->set("translation", Core::getLang()->get($this->get("name")), false);
  }

  public function getMetal($format = false)
  {
    return ($format) ? fNumber($this->get("basic_metal")) : $this->get("basic_metal");
  }

  public function getSilicon($format = false)
  {
    return ($format) ? fNumber($this->get("basic_silicon")) : $this->get("basic_silicon");
  }

  public function getHydrogen($format = false)
  {
    return ($format) ? fNumber($this->get("basic_hydrogen")) : $this->get("basic_hydrogen");
  }

  public function getEnergy($format = false)
  {
    return ($format) ? fNumber($this->get("basic_energy")) : $this->get("basic_energy");
  }
}
?>