<?php
/**
* Assault simulation.
*
* Oxsar http://oxsar.ru
*
* 
*/

class ExtMain extends Main
{
  protected function index()
  {
    Core::getTPL()->addHTMLHeaderFile("lib/jquery.jclock.js?".CLIENT_VERSION, "js");
    $clock = "<script type=\"text/javascript\">
      //<![CDATA[
      $(function($) {
        var options = {
seedTime: ".time()." * 1000
        }
        $('.jclock').jclock(options);
    });
    //]]>
    </script>
      <span class=\"jclock\"></span>";
    $time_now = date("d.m.Y", time())." ".$clock;
    Core::getTPL()->assign("serverClock", $time_now);
    parent::index();
  }
}
?>