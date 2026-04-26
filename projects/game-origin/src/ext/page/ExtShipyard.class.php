<?php
/**
* Assault simulation.
*
* Oxsar http://oxsar.ru
*
* 
*/

class ExtShipyard extends Shipyard
{
  protected function index()
  {
    Core::getTPL()->addHTMLHeaderFile("lib/jquery.countdown.js?".CLIENT_VERSION, "js");
    $result = sqlSelect("shipyard", array("finished"), "", "planetid = '".Core::getUser()->get("curplanet")."'", "finished DESC", "1");
    $row = sqlFetch($result);
    $timeleft = $row["finished"] - time();

    $tim_all = "<script type=\"text/javascript\">
      //<![CDATA[
      $(function () {
        $('#bCountDown').countdown({until: ".$timeleft.", compact: true, onExpiry: function() {
          $('#bCountDown').text('-');
        }});
    });
    //]]>
    </script>
      <span id=\"bCountDown\">".getTimeTerm($tim_all)."</span>";

    Core::getTPL()->assign("all_time", $tim_all);
    parent::index();
  }
}
?>