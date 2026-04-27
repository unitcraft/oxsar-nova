<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
<title>{lang}ASSAULT{/lang} - {config}pagetitle{/config}</title>
<!--<meta http-equiv="content-type" content="text/html; charset={@charset}" />-->

<link rel="shortcut icon" href="/favicon.ico" type="image/x-icon" />
<link rel="icon" href="/favicon.ico" type="image/x-icon" />

<link rel="stylesheet" type="text/css" href="{const}RELATIVE_URL{/const}css/layout.css?{const=CLIENT_VERSION}" />
<link rel="stylesheet" type="text/css" href="{const}RELATIVE_URL{/const}css/style.css?{const=CLIENT_VERSION}" />

{if[0]}<script type="text/javascript" src="{const=RELATIVE_URL}js/lib/jquery.js?{const=CLIENT_VERSION}"></script>{/if}
<script type="text/javascript" src="http://ajax.googleapis.com/ajax/libs/jquery/1.5.1/jquery.js"></script>
<script type="text/javascript" src="http://ajax.googleapis.com/ajax/libs/jqueryui/1.8.14/jquery-ui.js"></script>

<script type="text/javascript" src="{const=RELATIVE_URL}js/lib/jquery.bgiframe.min.js?{const=CLIENT_VERSION}"></script>

{if[$this->templateVars["display_send"] || $this->templateVars["display_settings"]]}
<!-- accordion -->
<link rel="stylesheet" href="{const}RELATIVE_URL{/const}css/ui-darkness/jquery-ui.css" type="text/css" media="all" />
<script type="text/javascript">
	$(function() {
		$("#accordion").accordion({
        collapsible: true,
        autoHeight: false
      });
	});

  var current_bg_style = '{@current_bg_style}';
  var current_table_style = '{@current_table_style}';
  var base_url = '{@report_url}'; // '{@base_url}';
</script>
<!-- accordion END-->

<script type="text/javascript" src="{const}RELATIVE_URL{/const}js/sendreport.js"></script>
{/if}

{if[0]}
<!-- player -->
{if[$this->templateVars["player"]]}
<script type="text/javascript">
var TSP_CONFIG = {
      autoPlay: true
    };
</script>
{/if}
<link rel="stylesheet" type="text/css" href="{const}RELATIVE_URL{/const}soundmanager2/360-player/360player.css" />
<script type="text/javascript" src="{const}RELATIVE_URL{/const}soundmanager2/360-player/script/berniecode-animator.js"></script>
<script type="text/javascript" src="{const}RELATIVE_URL{/const}soundmanager2/js/soundmanager2.js"></script>
<script type="text/javascript" src="{const}RELATIVE_URL{/const}soundmanager2/360-player/script/360player.js"></script>
<!-- 360 UI demo, canvas magic for IE -->
<!--[if IE]><script type="text/javascript" src="{const}RELATIVE_URL{/const}soundmanager2/360-player/script/excanvas.js"></script><![endif]-->
<!-- 360 UI demo, Apache-licensed animation library -->
<!-- <link rel="stylesheet" type="text/css" href="{const}RELATIVE_URL{/const}soundmanager2/360-player/brfix.css" /> -->
<script type="text/javascript">

	soundManager.bgColor = '#000000';
	soundManager.wmode = 'transparent';

	if (navigator.platform.match(/win32/i) && !navigator.userAgent.match(/msie/i)) {
	  // special case: Windows and wmode transparent don't get along. Sacrifice background color.
	  soundManager.bgColor = '#f9f9f9';
	}

	soundManager.url = 'soundmanager2/swf/';
	/*
	soundManager.useFlashBlock = true;
	soundManager.debugMode = true;
	soundManager.debugFlash = soundManager.debugMode;
	*/

  if (1) // !navigator.userAgent.match(/msie 6/i))
  {
	  threeSixtyPlayer.config.imageRoot = 'soundmanager2/360-player/';
  }
</script>
<link rel="stylesheet" type="text/css" href="{const}RELATIVE_URL{/const}soundmanager2/flashblock/flashblock.css" />

<style type="text/css">
  td {border: 1px solid;}
</style>
<!-- player END-->
{/if}

{@HTML_HEAD}
</head>
<body>

{if[0]}
<div id="soundmanager-debug" style="position:fixed; width:700px; height:400px; right:0; bottom:0; z-index:1000; border:solid 1px #000055; background:#000;"></div>
{/if}

<script type="text/javascript">
  $(function(){
    $(".bmat_row")
      .bind('mouseout', function(){
          $(this).find(".bmat_col").removeClass('bmat_col_sel');
        })
      .bind('mouseover', function(){
          $(this).find(".bmat_col").addClass('bmat_col_sel');
        });
  });
  
  function open_bmat(turn)
  {
    $(".bmat_panel_turn_"+turn).css("display", "inline");
    $(".bmat_open_panel_turn_"+turn).css("display", "none");
  }

  function close_bmat(turn)
  {
    $(".bmat_panel_turn_"+turn).css("display", "none");
    $(".bmat_open_panel_turn_"+turn).css("display", "inline");
  }
</script>

    {if[$this->templateVars["player"]]}
    <div class="player_conteiner" style="margin-left: 10px; float: left;">
        <div class="ui360"><a href="{@player_path}">{@player_title}</a></div>
    </div>
    {/if}
    
    {if[$this->templateVars["display_send"] || $this->templateVars["display_settings"]]}
    <center>
    <form method="post" action="{@formaction}">
    <table class="ntable" id="sendHeader" style="cursor: s-resize; text-align: left;">
      <tr>
        <th colspan="2" onclick="SoHform();">
         {lang}MENU_PREFERENCES{/lang}
         {if[$this->templateVars["display_send"]]} / {lang}REPORT_SEND_TO_FRIEND{/lang}{/if}
        </th>
      </tr>
    </table>
        <table class="ntable" id="sendTable" style="text-align: left; display: none;">
        {if[$this->templateVars["display_send"]]}
        <tfoot>
          <tr>
            <td>
              <input type="button" name="go" id="go" value="{lang}COMMIT{/lang}" onclick="sendReport('{const=RELATIVE_URL}sendreport.php');" class="button" />
              <input type="hidden" name="sid" id="sid" value="{@sid}" />
              <input type="hidden" name="id" id="id" value="{@id}" />
              <input type="hidden" name="key" id="key" value="{@key}" />
              <input type="hidden" name="key2" id="key2" value="{@key2}" />
              <input type="hidden" name="odnoklassniki" id="odnoklassniki" value="?" />
              <input type="hidden" name="track" id="track"/>
            </td>
          </tr>
        </tfoot>
        {/if}
        <tbody>
          <tr>
            <td>
              <table class="table_no_background">
                <colgroup>
                    <col width="40%"/>
                    <col width="*"/>
                </colgroup>
                <tbody>
                <tr>
                    <td>
                        <label for="user_bg_style">{lang}USER_BACKGROUND_STYLE{/lang}</label>
                    </td>
                    <td><select name="user_bg_style" id="user_bg_style" onChange='setActiveStyleSheet($("#user_bg_style > option:selected").attr("value"), "bg"); generateFriendLink();'><option value="">-</option>{foreach[user_bg_styles]}<option value="{loop}path{/loop}"{if[$row["path"] == $this->templateVars["current_bg_style"]]} selected="selected"{/if}>{loop}name{/loop}</option>{/foreach}</select></td>
                </tr>
                <tr>
                    <td>
                        <label for="user_table_style">{lang}USER_TABLE_STYLE{/lang}</label>
                    </td>
                    <td><select name="user_table_style" id="user_table_style" onChange='setActiveStyleSheet($("#user_table_style > option:selected").attr("value"), "table"); generateFriendLink();'><option value="">-</option>{foreach[user_table_styles]}<option value="{loop}path{/loop}"{if[$row["path"] == $this->templateVars["current_table_style"]]} selected="selected"{/if}>{loop}name{/loop}</option>{/foreach}</select></td>
                </tr>
                {if[$this->templateVars["display_send"]]}
                <tr>
                    <td>
                        <label for="email">{lang}REPORT_EMAIL{/lang}</label>
                    </td>
                    <td><textarea id="email" name="email" rows="2" cols="50"></textarea></td>
                </tr>
                <tr>
                    <td>
                        <label for="usertext">{lang}REPORT_SEND_TEXT{/lang}</label>
                    </td>
                    <td><textarea cols="50" rows="5" id="usertext" name="usertext"></textarea></td>
                </tr>
                {/if}
                </tbody>
              </table>
            </td>
          </tr>
          {if[0]}
          <tr>
            <td colspan="2" style="text-align: left;">
              <div id="accordion">
                  {foreach[albums]}
                  <h3><a href="#">{loop}title{/loop}</a></h3>
                  <div>
                      {loop}table{/loop}
                  </div>
                  {/foreach}
              </div>
            </td>
          </tr>
          {/if}
        </tbody>
        </table>
    </form>
    <table class="ntable"><tr><td id="sendreport" style="display: none;"></td></tr></table>
    </center>
    {/if}
    {if[defined('SN')]}
    {else}
    {if[1 || $this->templateVars["display_send"]]}
      <center>
        <p>
          {lang}FRIEND_LINK{/lang}
          <br/><a href="{@report_url}" id="friendlink2">{@report_url}</a>
        </p>
      </center>
    {/if}
    {/if}

{if[ !defined('SN') ]}
<center>
  <p>
	{include}"banner_top"{/include}
  </p>
</center>
{/if}
	
{@report}

{if[ defined('SN') ]}
{else}
{if[1 || $this->templateVars["display_send"]]}
<center>
  <p>
    {lang}FRIEND_LINK{/lang}
    <br/><a href="{@report_url}" id="friendlink">{@report_url}</a>
  </p>
</center>
{/if}
{/if}

{if[ !defined('SN') ]}
<center>
  <p>
	{include}"banner_bottom"{/include}
  </p>
</center>
{/if}

<!-- DO NOT MODIFY "POWERED BY" DIV, feel free to change "poweredby" css class -->
{if[0]}<div class="poweredby"><a href="http://netassault.ru" target="_blank">Powered by NetAssault</a></div>{/if}
</body>
</html>
{if[0]}
 Oxsar http://oxsar.ru
 Copyright (c) 2009-2010 UnitPoint <support@unitpoint.ru>
  
{/if}