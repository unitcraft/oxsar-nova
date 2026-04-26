<?php
/**
* This class helps to create a alliance list.
*
* Oxsar http://oxsar.ru
*
* 
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class AllianceList implements IteratorAggregate
{
  /**
  * Map which contains all alliances of a list.
  *
  * @var Map
  */
  protected $list = null;

  /**
  * Which type should be shown as points.
  *
  * @var string
  */
  protected $pointType = "points";

  /**
  * How many decimals can be printed using pNumber function.
  *
  * @var int
  */
  protected $maxDecimals = 2;

  /**
  * The rank will be fetched by an extra query.
  *
  * @var boolean
  */
  protected $fetchRankByQuery = false;

  /**
  * Relation object.
  *
  * @var Relation
  */
  protected $relation = null;

  /**
  * The tag will be formatted as a link.
  *
  * @var boolean
  */
  protected $tagAsLink = true;

  /**
  * Specific key when loading from database.
  *
  * @var string
  */
  protected $key = null;

  /**
  * Creates a new alliance list object.
  *
  * @param resource	Query result for a list
  *
  * @return void
  */
  public function __construct($list = null, $start = 0, $maxDecimals = 2)
  {
    $this->maxDecimals = $maxDecimals;
    $this->list = new Map();
    $this->relation = new Relation(NS::getUser()->get("userid"), NS::getUser()->get("aid"));
    if(is_array($list))
    {
      $this->setByArray($list, $start);
    }
    else if(!is_null($list))
    {
      $this->load($list, $start);
    }
    return;
  }

  /**
  * Loads the list from a sql query.
  *
  * @param resource	Result set
  * @param integer	Rank/Counter start
  *
  * @return AllianceList
  */
  public function load($result, $start = 0)
  {
    while($row = sqlFetch($result))
    {
      $row["counter"] = $start;
      $start++;
      $row["rank"] = $start;
      $row = $this->formatRow($row);
      $key = !is_null($this->key) ? intval($row[$this->key]) : $start;
      $this->list->set($key, $row);
    }
    sqlEnd($result);
    return $this;
  }

  /**
  * Sets the list by an array.
  *
  * @param array		Contains a list of all elements
  * @param integer	Rank/Counter start
  *
  * @return AllianceList
  */
  public function setByArray(array $list, $start = 0)
  {
    foreach($list as $key => $value)
    {
      $list[$key]["counter"] = $start;
      $start++;
      $row["rank"] = $start;
      $list[$key] = $this->formatRow($list[$key]);
    }
    $this->list = new Map($list);
    return $this;
  }

  /**
  * Formats an alliance record.
  *
  * @param array		Alliance data
  *
  * @return array	Formatted alliance data
  */
  protected function formatRow($row)
  {
    $row["tag"] = $this->formatAllyTag($row["tag"], $row["name"], $row["aid"]);
    $row["name"] = ($row["showhomepage"] && $row["homepage"] != "") ? Link::get($row["homepage"], $row["name"]) : $row["name"];
    if(!NS::getUser()->get("aid") && $row["open"] > 0)
    {
      $row["join"] = Link::get("game.php/go:Alliance/do:Apply/aid:".$row["aid"], Image::getImage("apply.gif", Core::getLanguage()->getItem("JOIN")));
    }
    $row["average"] = $row["members"] > 0 ? fNumber(floor($row["points"] / $row["members"])) : 0;
    $row["members"] = fNumber($row["members"]);
    if($this->fetchRankByQuery)
    {
      $row["rank"] = $this->getAllianceRank($row["aid"], $this->pointType);
    }
    else
    {
      $row["rank"] = fNumber($row["rank"]);
    }
    foreach(array(
      "points",
      "b_points",
      "r_points",
      "u_points",
      "b_count",
      "r_count",
      "u_count",
      "e_points",
      "battles",
      ) as $field)
    {
      $row[$field . "_num"] = $row[$field];
      $row[$field] = pNumber($row[$field], $this->maxDecimals);
    }
    $row["totalpoints_num"] = $row[$this->pointType . "_num"];
    $row["totalpoints"] = $row[$this->pointType];

    return $row;
  }

  /**
  * Fetches the alliance rank from database.
  *
  * @param integer	Alliance id
  *
  * @return integer	Rank
  */
  protected function getAllianceRank($aid, $pointType)
  {
    $joins  = "LEFT JOIN ".PREFIX."user2ally u2a ON (u2a.aid = a.aid)";
    $joins .= "LEFT JOIN ".PREFIX."user u ON (u.userid = u2a.userid)";
    $subselect = "(SELECT SUM(u.".$pointType.") FROM ".PREFIX."user2ally u2a LEFT JOIN ".PREFIX."user u ON (u.userid = u2a.userid) WHERE u2a.aid = ".sqlVal($aid).")";
    $result = sqlSelect("alliance a", "a.aid", $joins, "", "", "", "u2a.aid", "HAVING SUM(u.".$pointType.") >= ".$subselect, "");
    $rank = fNumber(Core::getDB()->num_rows($result));
    sqlEnd($result);
    return $rank;
  }

  /**
  * Formats an alliance tag as a link.
  *
  * @param string	Alliance tag
  * @param string	Alliance name
  * @param integer	Alliance id
  *
  * @return string	Formatted alliance tag
  */
  protected function formatAllyTag($tag, $name, $aid)
  {
    $class = $this->relation->getAllyRelationClass($aid);
    if($this->tagAsLink)
    {
      return Link::get("game.php/AlliancePage/".$aid, $tag, $name, $class);
    }
    return "<span class=\"".$class."\">".$tag."</span>";
  }

  /**
  * Returns the user list as an array.
  *
  * @return array
  */
  public function getArray()
  {
    return $this->list->getArray();
  }

  /**
  * Returns the user list as a map.
  *
  * @return Map
  */
  public function getMap()
  {
    return $this->list;
  }

  /**
  * Setter-function for point type.
  *
  * @param string	Point type
  *
  * @return AllianceList
  */
  public function setPointType($pointType = "points")
  {
    $this->pointType = $pointType;
    return $this;
  }

  /**
  * Setter-method for ranks.
  *
  * @param boolean
  *
  * @return AllianceList
  */
  public function setFetchRank($fetchRankByQuery)
  {
    $this->fetchRankByQuery = $fetchRankByQuery;
    return $this;
  }

  /**
  * Setter-method for tags.
  *
  * @param boolean
  *
  * @return AllianceList
  */
  public function setTagAsLink($tagAsLink)
  {
    $this->tagAsLink = $tagAsLink;
    return $this;
  }

  /**
  * Setter-method for map key.
  *
  * @param string	Map key label
  *
  * @return AllianceList
  */
  public function setKey($key)
  {
    $this->key = $key;
    return $this;
  }

  /**
  * Retrieves an external iterator.
  *
  * @return ArrayIterator
  */
  public function getIterator()
  {
    return $this->list->getIterator();
  }
}
?>