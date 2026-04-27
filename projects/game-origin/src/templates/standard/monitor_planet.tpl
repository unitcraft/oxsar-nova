<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
<title>{config}pagetitle{/config}</title>
<meta http-equiv="content-type" content="text/html; charset={@charset}" />
<meta http-equiv="Pragma" content="no-cache" />
<meta http-equiv="Cache-control" content="no-cache" />
<meta http-equiv="Expires" content="-1" />

<link rel="shortcut icon" href="/favicon.ico" type="image/x-icon" />
<link rel="icon" href="/favicon.ico" type="image/x-icon" />

{if[0]}<script type="text/javascript" src="{const=RELATIVE_URL}js/lib/jquery.js?{const=CLIENT_VERSION}"></script>{/if}
<script type="text/javascript" src="http://ajax.googleapis.com/ajax/libs/jquery/1.5.1/jquery.js"></script>
<script type="text/javascript" src="http://ajax.googleapis.com/ajax/libs/jqueryui/1.8.14/jquery-ui.js"></script>

<link rel="stylesheet" type="text/css" href="{const=RELATIVE_URL}css/layout.css?{const=CLIENT_VERSION}" />
<link rel="stylesheet" type="text/css" href="{const=RELATIVE_URL}css/style.css?{const=CLIENT_VERSION}" />

<script type="text/javascript" src="{const=RELATIVE_URL}js/lib/jquery.cookie.js?{const=CLIENT_VERSION}"></script>
<script type="text/javascript" src="{const=RELATIVE_URL}js/main.js?{const=CLIENT_VERSION}"></script>
<script type="text/javascript" src="{const=RELATIVE_URL}js/lib/jquery.countdown.js?{const=CLIENT_VERSION}"></script>

{@HTML_HEAD}
<script type="text/javascript">
//<![CDATA[
$(function () {
{foreach[events]}
	$('#timer_{loop}eventid{/loop}').countdown({until: {loop}time_r{/loop}, compact: true, onExpiry: function() {
		$('#timer_{loop}eventid{/loop}').text('-');
	}});
{/foreach}
});
//]]>
</script>
</head>
<body>

<table class="ntable">
	<tr>
		<th colspan="2">{lang}FLEET_ACTIVITIES{/lang}</th>
	</tr>
	{if[{var}num_rows{/var} <= 0]}
	  <tr>
		  <td>&nbsp;</td>
		  <td>{lang}NO_MATCHES_FOUND{/lang}</td>
	  </tr>
	{else}
	  {foreach[events]}
	  <tr>
		  <td width="100px" align="center">
			  <span id="timer_{loop}eventid{/loop}">{loop}time{/loop}</span>
		  </td>
		  <td><span class="{loop}class{/loop}">{loop}message{/loop}</span></td>
	  </tr>
	  {/foreach}
	{/if}
</table>
<!-- DO NOT MODIFY "POWERED BY" DIV, feel free to change "poweredby" css class -->
{if[0]}{if[0]}<div class="poweredby"><a href="http://netassault.ru" target="_blank">Powered by NetAssault</a></div>{/if}{/if}

</body>
</html>
{if[0]}
 Oxsar https://oxsar-nova.ru
 Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
  
{/if}