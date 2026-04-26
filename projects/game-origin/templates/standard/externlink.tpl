<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
<title>{config}pagetitle{/config}</title>
<meta http-equiv="content-type" content="text/html; charset={@charset}" />
<meta http-equiv="Cache-control" content="no-cache" />
<meta http-equiv="pragma" content="no-cache" />

<link rel="shortcut icon" href="/favicon.ico" type="image/x-icon" />
<link rel="icon" href="/favicon.ico" type="image/x-icon" />

<link rel="stylesheet" type="text/css" href="{const}RELATIVE_URL{/const}css/layout.css?{const=CLIENT_VERSION}" />
<link rel="stylesheet" type="text/css" href="{const}RELATIVE_URL{/const}css/style.css?{const=CLIENT_VERSION}" />

<meta http-equiv="refresh" content="3; URL={@link}" />

</head>
<body>
<div id="forward" class="centered">
<a href="{@link}">
	{lang}EXTERN_LINK{/lang}
	<br />
	{@link}
</a>
</div>
</body>
</html>
{if[0]}
 Oxsar http://oxsar.ru
 Copyright (c) 2009-2010 UnitPoint <support@unitpoint.ru>
  
{/if}