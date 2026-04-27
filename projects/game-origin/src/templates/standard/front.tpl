<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.1//EN" "http://www.w3.org/TR/xhtml11/DTD/xhtml11.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
<title>{@page} - {config}pagetitle{/config}</title>
<meta http-equiv="Content-Type" content="text/html; charset={@charset}" />
<meta name="robots" content="index,follow" />
<meta http-equiv="expires" content="0" />
<meta http-equiv="Pragma" content="no-cache" />
<meta http-equiv="Cache-control" content="no-cache" />
<meta http-equiv="Expires" content="-1" />

<link rel="shortcut icon" href="/favicon.ico" type="image/x-icon" /> 
<link rel="icon" href="/favicon.ico" type="image/x-icon" />

<link rel="stylesheet" type="text/css" href="{const}RELATIVE_URL{/const}css/preview.css?{const=CLIENT_VERSION}" />

{if[0]}<script type="text/javascript" src="{const=RELATIVE_URL}js/lib/jquery.js?{const=CLIENT_VERSION}"></script>{/if}
<script type="text/javascript" src="http://ajax.googleapis.com/ajax/libs/jquery/1.5.1/jquery.js"></script>
<script type="text/javascript" src="http://ajax.googleapis.com/ajax/libs/jqueryui/1.8.14/jquery-ui.js"></script>

{@HTML_HEAD}
</head>
<body>
<div id="header">
<span id="title">{config}pagetitle{/config}</span>
</div>
<div id="menubar">
	{link[SIGN_IN]}"login.php"{/link}
	{link[SIGN_UP]}"signup.php"{/link}
	{foreach[headerMenu]}{loop}link{/loop}{/foreach}
</div>
<div id="content">
{include}$template{/include}
<div class="sub-links">{foreach[footerMenu]}{loop}link{/loop}{/foreach}</div>
</div>
<!-- DO NOT MODIFY "POWERED BY" DIV, feel free to change "poweredby" css class -->
{if[0]}<div class="poweredby"><a href="http://netassault.ru" target="_blank">Powered by NetAssault</a></div>{/if}
</body>
</html>
{if[0]}
 Oxsar https://oxsar-nova.ru
 Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
  
{/if}