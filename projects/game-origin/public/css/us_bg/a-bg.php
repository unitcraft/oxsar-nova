<?php
/**
* Oxsar http://oxsar.ru
*
* 
*/

header('Content-type: text/css');

?>
body
{
  background-image: url( "../../images/bg/a-bg-<?php echo intval($_GET['id']); ?>.jpg" );
  background-repeat:no-repeat;
  background-attachment: fixed;
  background-position: center center;
  background-color: #000;
}

