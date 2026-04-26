<?php
/**
* mysql_pdo database functions — Yii-free PDO implementation.
*/

if(!defined("RECIPE_ROOT_DIR")) { die("Hacking attempt detected."); }

require_once(RECIPE_ROOT_DIR."database/Database.abstract_class.php");

class DB_mysql_pdo extends Database
{
	public $mysql_pdo;

	public function __construct($host, $user, $pw, $db, $port = null)
	{
		parent::__construct($host, $user, $pw, $db, $port);
		$this->init();
	}

	public function __destruct()
	{
		$this->mysql_pdo = null;
	}

	protected function init()
	{
		$port = $this->port ? $this->port : '3306';
		$dsn  = "mysql:host={$this->host};port={$port};dbname={$this->database};charset=utf8mb4";
		$this->mysql_pdo = new PDO($dsn, $this->user, $this->pw, [
			PDO::ATTR_ERRMODE            => PDO::ERRMODE_EXCEPTION,
			PDO::ATTR_DEFAULT_FETCH_MODE => PDO::FETCH_ASSOC,
			PDO::ATTR_EMULATE_PREPARES   => true,
		]);
		return $this;
	}

	public function query($sql)
	{
		$sql = trim($sql);
		$queryTime = new Timer(microtime());
		$stmt = $this->mysql_pdo->query($sql);
		$this->qTime[++$this->queryCount] = $queryTime->getTime(false);
		return $stmt;
	}

	public function fetch($result)
	{
		return $result ? $result->fetch(PDO::FETCH_ASSOC) : false;
	}

	public function num_rows($query)
	{
		if ($query) {
			return $query->rowCount();
		}
		return 0;
	}

	public function insert_id()
	{
		return $this->mysql_pdo->lastInsertId();
	}

	public function quote_db_value($string)
	{
		return $this->mysql_pdo->quote($string);
	}

	public function free_result($resource)
	{
		if ($resource) {
			$resource->closeCursor();
		}
		return true;
	}
}
