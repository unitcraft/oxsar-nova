<?php
/**
* Sql functions.
*
* Oxsar http://oxsar.ru
*
* 
*/

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class Sql
{
  public static function v()
  {
    $data = func_num_args() > 1 ? func_get_args() : func_get_arg(0);
    if(is_array($data))
    {
      $result = array();
      foreach($data as $value)
      {
        $result[] = sqlVal($value);
      }
      return count($result) ? implode(",", $result) : 'NULL';
    }
    return sqlVal($data);
  }

  public static function select($params)
  {
    if(isset($params['from']))
    {
      $from = array();
      foreach((array)$params['from'] as $key => $value)
      {
        if(is_string($value))
        {
          if(is_string($key))
          {
            $from[] = "`$key` `$value`"
          }
          else
          {
            $from[] = $value
          }
          continue;
        }
        if(is_array($value))
        {

        }
      }
    }
  }

  protected static function formatWhereQuery($params)
  {
    $list = array();
    $data = func_num_args() > 1 ? func_get_args() : func_get_arg(0);
    if(!is_array($data))
    {
      return sqlVal($data);
    }
    foreach($data as $key => $value)
    {
      if(is_array($value))
      {
        switch($key)
        {
        case 'OR': case 'or':
        case 'AND': case 'and':
        case '+':
        case '-':
        case '*':
        case '/':
        case '%':
        case '&':
        case '|':
          $list[] = "(".implode(" $key ", self::formatWhereQuery($value).")";
          continue;

        case 'NOT':
        case 'not':
        case '!':
          $list[] = "( NOT (".implode(" AND ", self::formatWhereQuery($value)."))";
          continue;
        }
        if(preg_match("#^\s*(.+?)\s+BETWEEN\s*$#is", $key, $regs))
        {
          $list[] = "(`".$regs[1])."` $regs[2] ".sqlVal($value[0])." AND ".sqlVal($value[1]).")";
          continue;
        }
        continue;
      }
      if(preg_match("#^\s*(.+?)\s+(=|>=|<=|!=|<>|LIKE)\s*$#is", $key, $regs))
      {
        $list[] = "(`".$regs[1])."` $regs[2] ".sqlVal($value).")";
      }
      else if(preg_match("#^\s*(.+?)\s+(IS(\s+NOT)?)\s*$#is", $key, $regs))
      {
        $list[] = "(`".$regs[1])."` $regs[2] NULL)";
      }
      else
      {
        $list[] = "(`".trim($key)."` = ".sqlVal($value).")";
      }
    }
    return $list;
  }
}

?>