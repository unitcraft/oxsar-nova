<?php
/**
* MySQL database functions.
*
* @package Recipe 1.1
* @author Sebastian Noll
* @copyright Copyright (c) 2008, Sebastian Noll
* @license <http://www.gnu.org/licenses/gpl.txt> GNU/GPL
* @version $Id: DB_MySQL.class.php 23 2010-04-03 19:08:34Z craft $
*/

if(!defined("RECIPE_ROOT_DIR")) { die("Hacking attempt detected."); }

require_once(RECIPE_ROOT_DIR."database/Database.abstract_class.php");

class DB_MySQL extends Database
{
	/**
	* Pointer on current database.
	*
	* @var resource
	*/
	protected $dbpointer;

	/**
	* State of database selection.
	*
	* @var boolean;
	*/
	protected $dbselect = false;

	/**
	* Constructor: Set database access data.
	*
	* @param string	The database host
	* @param string	The database username
	* @param string	The database password
	* @param string	The database name
	* @param integer	The database port
	*
	* @return void
	*/
	public function __construct($host, $user, $pw, $db, $port = null)
	{
		parent::__construct($host, $user, $pw, $db, $port);
		try { $this->connectToDatabase(); }
		catch(Exception $e) { $e->printError(); }
		try { $this->selectDatabase(); }
		catch(Exception $e) { $e->printError(); }
		return;
	}

	/**
	* Close current database connection.
	*
	* @return void
	*/
	public function __destruct()
	{
		if($this->dbpointer === false) { return; }
		mysql_close($this->dbpointer);
		return $this;
	}

	/**
	* Establish database connection.
	*
	* @return void
	*/
	protected function connectToDatabase()
	{
		$this->dbpointer = @mysql_connect($this->host.($this->port) ? ":".$this->port : "", $this->user, $this->pw);
		if($this->dbpointer === false) { throw new GenericException("Could not connect to database: ".mysql_error());	}
		$this->query("set charset utf8;");
		return $this;
	}

	/**
	* Select the database which is to used.
	*
	* @return void
	*/
	protected function selectDatabase()
	{
		$this->dbselect = @mysql_select_db($this->database, $this->dbpointer);
		if($this->dbselect === false) { throw new GenericException("Could not select database \"".$this->database."\"."); }
		return $this;
	}

	/**
	* Purpose a query on selected database.
	*
	* @param string	The SQL query
	*
	* @return resource	Results of the query
	*/
	public function query($sql)
	{
		$queryTime = new Timer(microtime());
		$this->query = @mysql_query($sql, $this->dbpointer);
		$this->qTime[$this->queryCount+1] = $queryTime->getTime(false);
		if(mysql_errno()) { throw new GenericException("SQL Error (".mysql_errno()."): ".mysql_error()."<br /><br />Query Code: ".$sql); }
		$this->queryCount++;
		return $this->query;
	}

	/**
	* Returns the row of a query as an object.
	*
	* @param resource	The SQL query id
	* @return object	The data of a row
	*/
	public function fetch_object($query)
	{
		$this->result = mysql_fetch_object($query);
		return $this->result;
	}

	/**
	* Returns the row of a query as an array.
	*
	* @param resource	The SQL query id
	* @return array	The data of a row
	*/
	public function fetch_array($query)
	{
		$this->result = mysql_fetch_array($query);
		return $this->result;
	}

	/**
	* Fetch a result row as an associative array.
	*
	* @param resource	The SQL query id
	* @return array	The data of a row
	*/
	public function fetch($query)
	{
		$this->result = mysql_fetch_assoc($query);
		return $this->result;
	}

	/**
	* Returns the value from a result resource.
	*
	* @param resource	The SQL query id
	* @param string	The column name to fetch
	* @param integer	Row number in result to fetch
	*
	* @return string
	*/
	public function fetch_field($query, $field, $row = null)
	{
		if($row !== null)
		{
			$this->result = mysql_result($query, $row, $field);
		}
		else
		{
			$array = $this->fetch($query);
			$this->result = $array[$field];
		}
		return $this->result;
	}

	/**
	* Get a row as an enumerated array.
	*
	* @param resource	The SQL query id
	*
	* @return array
	*/
	public function fetch_row($query)
	{
		$this->result = mysql_fetch_row($query);
		return $this->result;
	}

	/**
	* Returns the total row numbers of a query.
	*
	* @param resource	The SQL query id
	* @return integer	The total row number
	*/
	public function num_rows($query)
	{
		if($query)
		{
			return mysql_num_rows($query);
		}
		return 0;
	}

	/**
	* Returns the number of affected rows by the last query.
	*
	* @return integer	Affected rows
	*/
	public function affected_rows()
	{
		$affected_rows = mysql_affected_rows($this->dbpointer);
		if($affected_rows < 0) { $affected_rows = 0; }
		return $affected_rows;
	}

	/**
	* Returns the last inserted ID of a table.
	*
	* @return integer	The last inserted id
	*/
	public function insert_id()
	{
		return mysql_insert_id();
	}

	/**
	* Escapes a string for a safe SQL query.
	*
	* @param string	The string that is to be escaped
	*
	* @return string	Returns the escaped string, or false on error
	*/
	public function real_escape_string($string)
	{
		return mysql_real_escape_string($string, $this->dbpointer);
	}

	/**
	* Returns used MySQL-Verions.
	*
	* @return string	MySQL-Version
	*/
	public function getVersion()
	{
		return @mysql_get_client_info();
	}

	/**
	* Type of database.
	*
	* @return string
	*/
	public function getDatabaseType()
	{
		return "MySQL";
	}

	/**
	* Resets a result resource to row number 0.
	*
	* @param resource	Resource to reset
	*
	* @return void
	*/
	public function reset_resource(&$resource)
	{
		return mysql_data_seek($resource, 0);
	}

	/**
	* Frees stored result memory for the given statement handle.
	*
	* @param resource	The statement to free
	*
	* @return boolean	True on success, false on failure
	*/
	public function free_result($resource)
	{
		return mysql_free_result($resource);
	}
}
?>