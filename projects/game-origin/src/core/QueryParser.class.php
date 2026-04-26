<?php
/**
* Query parser. Processes queries for SQL access.
*
* @package Recipe 1.1
* @author Sebastian Noll
* @copyright Copyright (c) 2008, Sebastian Noll
* @license <http://www.gnu.org/licenses/gpl.txt> GNU/GPL
* @version $Id: QueryParser.class.php 23 2010-04-03 19:08:34Z craft $
*/

if(!defined("RECIPE_ROOT_DIR")) { die("Hacking attempt detected."); }

class QueryParser
{
  /**
  * Last SQL query.
  *
  * @var string
  */
  protected $sql = "";

  /**
  * Disables auto-sending of a query.
  *
  * @var boolean
  */
  protected $send = true;

  /**
  * @ignore
  */
  public function __construct() {}

  /**
  * Generates an insert sql query.
  *
  * @param string	Table name to insert
  * @param mixed		Attributes to insert
  * @param mixed		Values to insert
  *
  * @return QueryParser
  */
  public function insert($table, $attribute, $value, $op = "INSERT")
  {
//    try {
    	Arr::checkArrays($attribute, $value);
//    }
//    catch(Exception $e) { $e->printError(); }

    if(!is_array($attribute) && !is_array($value))
    {
      $attribute = Arr::trimArray(explode(",", $attribute));
      $value = Arr::trimArray(explode(",", $value));
    }

//    try {
    	Arr::checkArraySize($attribute, $value);
//    }
//    catch(Exception $e) { $e->printError(); }

    $attribute = implode(",", $this->setBackQuotes($attribute));
    $value = implode(",", $this->setSimpleQuotes($value));

    $this->sql = "$op INTO ".PREFIX.$table." (".$attribute.") VALUES (".$value.")";
    if($this->send)
    {
//      try {
      	Core::getDB()->query($this->sql);
//      }
//      catch(Exception $e) { $e->printError(); }
    }
    $this->send = true;
    return $this;
  }

  public function replace($table, $attribute, $value)
  {
    return $this->insert($table, $attribute, $value, "REPLACE");
  }

  /**
  * Generates an insert sql query.
  *
  * @param string	Table name to insert
  * @param mixed		Attributes to insert
  * @param mixed		Values to insert
  * @param string	Where clauses
  *
  * @return QueryParser
  */
  public function update($table, $attribute, $value, $where = null)
  {
//    try {
    	Arr::checkArrays($attribute, $value);
//    }
//    catch(Exception $e) { $e->printError(); }

    if(!is_array($attribute) && !is_array($value))
    {
      $attribute = Arr::trimArray(explode(",", $attribute));
      $value = Arr::trimArray(explode(",", $value));
    }

    $value = $this->setSimpleQuotes($value);
    $attribute = $this->setBackQuotes($attribute);

//    try {
    	Arr::checkArraySize($attribute, $value);
//    }
//    catch(Exception $e) { $e->printError(); }

    $update = $attribute[0]." = ".$value[0];
    for($i = 1; $i < count($attribute); $i++)
    {
      $update .= ", ".$attribute[$i]." = ".$value[$i];
    }

    $this->sql = "UPDATE ".PREFIX.$table." SET ".$update;
    if($where != null)
    {
      $this->sql .= " WHERE ".$where;
    }

    if($this->send)
    {
//      	echo '...trying to exec sql...';
//      	echo $this->sql;
      	$result = Core::getDB()->query($this->sql);
//      	echo 111;
//      	$e->getMessage();
    }
    $this->send = true;
    return $this;
  }

  /**
  * Generates a select query.
  *
  * @param string	Table name to select
  * @param mixed		Attributes to select
  * @param string	Tables to join
  * @param string	Where clauses
  *
  * @return resource	The SQL-Statement
  */
  public function select($table, $select, $join = "", $where = "", $order = "", $limit = "", $groupby = "", $other = "")
  {
    if(!is_array($select))
    {
      $select = Arr::trimArray(explode(",", $select));
    }
  
    if(is_array($join))
    {
		if( count($join) > 0 )
		{
			$join = implode(" ", $join);
		}
		else
		{
			$join = "";
		}
    }
  
    if(is_array($where))
    {
		if( count($where) > 0 )
		{
			$where = "(" . implode(") AND (", $where) . ")";
		}
		else
		{
			$where = "";			
		}
		
    }

    if($join) { $join = " ".$join; }
    if($where) { $where = " WHERE ".$where; }
    if($groupby) { $groupby = " GROUP BY ".$groupby; }
    if($order) { $order = " ORDER BY ".$order; }
    if($limit) { $limit = " LIMIT ".$limit; }
    $other = ($other != "") ? " ".$other : "";
    $select = implode(", ", $select);

    $this->sql = "SELECT ".$select." FROM ".PREFIX.$table.$join.$where.$groupby.$order.$limit.$other;
    if($this->send)
    {
//      try {
      	return Core::getDB()->query($this->sql);
//      }
//      catch(Exception $e) { $e->printError(); }
    }
    $this->send = true;
    return $this;
  }

  public function selectRow($table, $select, $join = "", $where = "", $order = "", $limit = "", $groupby = "", $other = "")
  {
    $result = $this->select($table, $select, $join, $where, $order, $limit, $groupby, $other);
    $row = Core::getDB()->fetch($result);
    Core::getDB()->free_result($result);

    return $row;
  }

  public function selectField($table, $select, $join = "", $where = "", $order = "", $limit = "", $groupby = "", $other = "")
  {
    $result = $this->select($table, $select, $join, $where, $order, $limit, $groupby, $other);
    // $field = Core::getDB()->fetch_field($result, is_array($select) ? reset($select) : $select);
    $row = Core::getDB()->fetch($result);
    Core::getDB()->free_result($result);

    return is_array($row) ? reset($row) : null; // $field;
  }

  /**
  * Generates a delete query.
  *
  * @param string	Table name to delete
  * @param mixed		Where clauses
  *
  * @return QueryParser
  */
  public function delete($table, $where = null)
  {
    if($where != null && strlen($where) > 0)
    {
      $whereclause = " WHERE ".$where;
    }
    else { $whereclause = ""; }
    $this->sql = "DELETE FROM ".PREFIX.$table.$whereclause;

    if($this->send)
    {
//      try {
      	Core::getDB()->query($this->sql);
//      }
//      catch(Exception $e) { $e->printError(); }
    }
    $this->send = true;
    return $this;
  }

  /**
  * Empties a table completely.
  *
  * @param string	Table to empty
  *
  * @return QueryParser
  */
  public function truncate($table)
  {
    $table = $this->setBackQuotes(PREFIX.$table);
    $this->sql = "TRUNCATE TABLE ".$table;

    if($this->send)
    {
//      try {
      	Core::getDB()->query($this->sql);
//      }
//      catch(Exception $e) { $e->printError(); }
    }
    $this->send = true;
    return $this;
  }

  /**
  * Reclaims the unused space of a database file.
  *
  * @param mixed		Tables to optimize
  *
  * @return QueryParser
  */
  public function optimize($tables)
  {
    if(is_array($tables))
    {
      $tables = $this->setPrefix($tables);
      $tables = $this->setBackQuotes($tables);
      $table = implode(", ", $tables);
    }
    else
    {
      $table = $this->setBackQuotes(PREFIX.$tables);
    }
    $this->sql = "OPTIMIZE TABLE ".$table;

    if($this->send)
    {
//      try {
      	Core::getDB()->query($this->sql);
//      }
//      catch(Exception $e) { $e->printError(); }
    }
    $this->send = true;
    return $this;
  }

  /**
  * Removes one or more tables.
  *
  * @param mixed		Tables
  *
  * @return QueryParser
  */
  public function drop($tables)
  {
    if(is_array($tables))
    {
      $tables = $this->setPrefix($tables);
      $tables = $this->setBackQuotes($tables);
      $table = implode(", ", $tables);
    }
    else
    {
      $table = $this->setBackQuotes(PREFIX.$tables);
    }
    $this->sql = "DROP TABLE IF EXISTS ".$table;

    if($this->send)
    {
//      try {
      	Core::getDB()->query($this->sql);
//      }
//      catch(Exception $e) { $e->printError(); }
    }
    $this->send = true;
    return $this;
  }

  /**
  * Renames a table.
  *
  * @param string	Table to rename
  * @param string	New table name
  *
  * @return QueryParser
  */
  public function rename($table, $newname)
  {
    $table = $this->setBackQuotes(PREFIX.$table);
    $newname = $this->setBackQuotes(PREFIX.$newname);
    $this->sql = "RENAME TABLE ".$table." TO ".$newname;

    if($this->send)
    {
//      try {
      	Core::getDB()->query($this->sql);
//      }
//      catch(Exception $e) { $e->printError(); }
    }
    $this->send = true;
    return $this;
  }

  /**
  * Returns information about the columns in the given table.
  *
  * @param string	Table name
  *
  * @return resource	The SQL-Statement
  */
  public function showFields($table)
  {
    $table = $this->setBackQuotes(PREFIX.$table);
    $this->sql = "SHOW COLUMNS FROM ".$table;

    if($this->send)
    {
//      try {
      	return Core::getDB()->query($this->sql);
//      }
//      catch(Exception $e) { $e->printError(); }
    }
    $this->send = true;
    return $this;
  }

  /**
  * Surrounds an array with simple quotes.
  *
  * @param array
  *
  * @return array
  */
  protected function setSimpleQuotes($data)
  {
    if(is_array($data))
    {
      $size = count($data);
      for($i = 0; $i < $size; $i++)
      {
        $data[$i] = (is_null($data[$i])) ? "NULL" : Core::getDatabase()->quote_db_value($data[$i]);
      }
      return $data;
    }
    return (is_null($data)) ? "NULL" : Core::getDatabase()->quote_db_value($data);
  }

  /**
  * Surrounds an array with back quotes.
  *
  * @param array
  *
  * @return array
  */
  protected function setBackQuotes($data)
  {
    if(is_array($data))
    {
      $size = count($data);
      for($i = 0; $i < $size; $i++)
      {
        if(Str::substring($data[$i], 0, 1) == "`") { continue; }
        $data[$i] = "`".$data[$i]."`";
      }
      return $data;
    }
    if(Str::substring($data, 0, 1) == "`")
    {
      return $data;
    }
    return "`".$data."`";
  }

  /**
  * Returns the last generated SQL query.
  *
  * @return string	Last SQL query
  */
  public function getLastQuery()
  {
    return $this->sql;
  }

  /**
  * Sets the prefix to an array of table names.
  *
  * @param array
  *
  * @return array
  */
  protected function setPrefix($tables)
  {
    $size = count($tables);
    for($i = 0; $i < $size; $i++)
    {
      $tables[$i] = PREFIX.$tables[$i];
    }
    return $tables;
  }

  /**
  * Sets whether the next generated query should be executed.
  *
  * @param boolean
  *
  * @return QueryParser
  */
  public function sendNextQuery($send)
  {
    $this->send = $send;
    return $this;
  }

  /**
  * Executes the last query.
  *
  * @return QueryParser
  */
  public function executeLastQuery()
  {
//    try {
      Core::getDB()->query($this->sql);
//    }
//    catch(Exception $e)
//    {
//      $e->printError();
//    }
    return $this;
  }
}
?>
