<?php

  require_once(dirname(__FILE__) . "/../global.inc.php");
  require_once(dirname(__FILE__) . "/../Config.inc.php");
  
  $conn = mysql_connect($database["host"], $database["user"], $database["userpw"]);
  if(mysql_select_db($database["databasename"]))
  {  
    $res = mysql_query("SHOW FULL TABLES");
    $tables = array();
    while($row = mysql_fetch_array($res))
    {
      $table_name = $row[0];
      $table_type = $row[1];

      /* echo "<pre>";
      print_r($row);
      echo "</pre>"; */

      if($table_type == "BASE TABLE")
      {
        $tables[] = $table_name;
      }
    }
    mysql_free_result($res);

    echo "<pre>";
    print_r($tables);
    echo "</pre>";

    $result = array();
    foreach($tables as $table_name)
    {
      $res = mysql_query("check table $table_name");
      $row = mysql_fetch_array($res);
      mysql_free_result($res);

      echo "<pre>";
      print_r($row);
      echo "</pre>";
      flush();
      
      $res = mysql_query("optimize table $table_name");
      $row = mysql_fetch_array($res);
      mysql_free_result($res);

      echo "<pre>";
      print_r($row);
      echo "</pre>";
      flush();
      
      $res = mysql_query("analyze table $table_name");
      $row = mysql_fetch_array($res);
      mysql_free_result($res);

      echo "<pre>";
      print_r($row);
      echo "</pre>";
      flush();
      
      $result[] = array("msg_text" => $row["Msg_text"], "table" => $table_name);
    }

    echo "<pre>";
    print_r($result);
    echo "</pre>";

    if(0)
    {
      $text = "Результат оптимизации:\n";
      for($i = 0; $i < count($result); $i++)
        $text .= "  " . $result[$i]["table"] . 
          " - " . $result[$i]["msg_text"] . "\n";
      
      if(!EMAILER_SKIP_CHECK)
        mail(get_board_var("email_check"), "cron optimize tables", $text);
    }
  }
  mysql_close($conn);
?>
