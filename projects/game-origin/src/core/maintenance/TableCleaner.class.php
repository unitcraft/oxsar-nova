<?php
/**
* Resets the auto-increment value of a table and resorts all entries.
*
* @package Recipe 1.1
* @author Sebastian Noll
* @copyright Copyright (c) 2008, Sebastian Noll
* @license <http://www.gnu.org/licenses/gpl.txt> GNU/GPL
* @version $Id: TableCleaner.class.php 23 2010-04-03 19:08:34Z craft $
*/

if(!defined("RECIPE_ROOT_DIR")) { die("Hacking attempt detected."); }

class TableCleaner
{
  /**
  * Table name that shall be reset.
  *
  * @var string
  */
  protected $tableName = "";

  /**
  * Primary key of the table.
  *
  * @var string
  */
  protected $primaryKey;

  /**
  * Field informations of the table.
  *
  * @var array
  */
  protected $fields = array();

  /**
  * Charset of the table.
  *
  * @var string
  */
  protected $defaultCharset = "binary";

  /**
  * Starts the table cleaner.
  *
  * @param string Table to clean
  * @param string Primary key
  *
  * @return void
  */
  public function __construct($tableName, $primaryKey = null)
  {
    $this->tableName = $tableName;
    $this->primaryKey = $primaryKey;
    $this->loadFields();
    return;
  }

  /**
  * Loads the fields of a table.
  *
  * @return void
  */
  protected function loadFields()
  {
    $result = Core::getQuery()->showFields($this->tableName);
    while($row = Core::getDB()->fetch($result))
    {
      array_push($this->fields, $row);
    }
    if(is_null($this->primaryKey))
    {
      $this->primaryKey = $this->fields[0]["Field"];
    }
    return;
  }

  /**
  * Creates a buffer table to save the data temporarily.
  *
  * @return void
  */
  protected function buildBufferTable()
  {
    $sql = "CREATE TABLE ".PREFIX.$this->tableName."_buffer (";
    foreach($this->fields as $field)
    {
      $sql .= "`".$field["Field"]."` ".$field["Type"]." ".(($field["Null"] == "YES") ? "NOT NULL" : "NULL")." ".(($field["Default"] != "") ? " default '".$field["Default"]."'" : "")." ".$field["Extra"].",\n";
    }
    $sql .= "PRIMARY KEY (`".$this->primaryKey."`)\n".
      ") ENGINE=MyISAM DEFAULT CHARSET=".$this->defaultCharset." AUTO_INCREMENT=1";
//    try {
    	Core::getDB()->query($sql);
//    }
//    catch(Exception $e) { $e->printError(); }
    return;
  }

  /**
  * Clean the table by using a buffer table.
  *
  * @return void
  */
  public function clean()
  {
    if(count($this->fields) <= 0)
    {
      return;
    }
    $this->buildBufferTable();
    $fieldList = array();
    foreach($this->fields as $field)
    {
      if($field["Field"] == $this->primaryKey) { continue; }
      array_push($fieldList, "`".$field["Field"]."`");
    }
    $fields = implode(", ", $fieldList);
    $sql = "INSERT INTO ".PREFIX.$this->tableName."_buffer (".$fields.") SELECT ".$fields." FROM ".PREFIX.$this->tableName." ORDER BY ".$this->primaryKey." ASC";
//    try {
    	Core::getDB()->query($sql);
//    }
//    catch(Exception $e) { $e->printError(); }

    Core::getQuery()->drop($this->tableName);
    Core::getQuery()->rename($this->tableName."_buffer", $this->tableName);
    return;
  }

  /**
  * Sets the default charset for the buffer table.
  *
  * @param string
  *
  * @return void
  */
  public function setDefaultCharset($defaultCharset)
  {
    $this->defaultCharset = $defaultCharset;
    return;
  }
}
?>