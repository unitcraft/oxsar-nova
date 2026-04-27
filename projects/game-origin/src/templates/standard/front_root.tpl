<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
<title>{@page} - {config}pagetitle{/config}</title>
<meta http-equiv="Content-Type" content="text/html; charset={@charset}" />
<meta name="robots" content="index,follow" />
<meta http-equiv="expires" content="0" />
<meta http-equiv="Pragma" content="no-cache" />
<meta http-equiv="Cache-control" content="no-cache" />
<meta http-equiv="Expires" content="-1" />
<META NAME="webmoney.attestation.label" CONTENT="webmoney attestation label#C61D9149-3BBF-43E4-B1ED-85104EED3D51" />

<link rel="shortcut icon" href="/favicon.ico" type="image/x-icon" />
<link rel="icon" href="/favicon.ico" type="image/x-icon" />

<link rel="stylesheet" type="text/css" href="{const}RELATIVE_URL{/const}css/preview.css?{const=CLIENT_VERSION}" />

{if[0]}<script type="text/javascript" src="{const=RELATIVE_URL}js/lib/jquery.js?{const=CLIENT_VERSION}"></script>{/if}
<script type="text/javascript" src="http://ajax.googleapis.com/ajax/libs/jquery/1.5.1/jquery.js"></script>
<script type="text/javascript" src="http://ajax.googleapis.com/ajax/libs/jqueryui/1.8.14/jquery-ui.js"></script>

<link rel="stylesheet" type="text/css" href="{const}RELATIVE_URL{/const}css/layout.css?{const=CLIENT_VERSION}" />
<link rel="stylesheet" type="text/css" href="{const}RELATIVE_URL{/const}css/style.css?{const=CLIENT_VERSION}" />
<link rel="stylesheet" type="text/css" href="{const}RELATIVE_URL{/const}css/front.css?{const=CLIENT_VERSION}" />
<link rel="stylesheet" type="text/css" href="{const}RELATIVE_URL{/const}css/us_table/table_std_bg_70.css?{const=CLIENT_VERSION}" />
{@HTML_HEAD}
{if[isFacebookSkin() || isMobiSkin()]}
<link rel="stylesheet" type="text/css" href="{const=RELATIVE_URL}css/skin/fb.css?{const=CLIENT_VERSION}">
{/if}
</head>
<body>
	<div align="center">
		<div class="main_content">
			<div id="header">
				<span id="title">
					<a href="http://www.oxsar.ru/">{config}pagetitle{/config}</a>
				</span>
			</div>
			<div id="menubar">
				{link[SIGN_IN]}"login.php"{/link}
				{link[SIGN_UP]}"signup.php"{/link}
				{foreach[headerMenu]}{loop}link{/loop}{/foreach}
			</div>

			<div class="clear"></div>

			<div class="front_left_small">
				{include}$template{/include}
				{if[isMobiSkin()]}
					{include}"front_user_areement"{/include}
				{/if}
			</div>
			{if[!isMobiSkin() && count($this->getLoop("agreemet")) > 0]}
				<div class="front_right">
					{include}"front_user_areement"{/include}
				</div>
			{else}
				<table class="ntable front_right">
					{include}"signup_ext"{/include}
					<tr>
						<th colspan="6">{lang=SPACE_EVENTS_LIST}</th>
					</tr>
					{foreach[spaceEvents]}
						<tr>
							<td>{loop=coords}</td>
							{if[$row["building_name"]]}
								<td>{loop=mode_name}</td>
								<td>{loop=building_image}</td>
								<td>{loop=building_name}</td>
							{else}
								<td colspan="3">{loop=mode_name}</td>
							{/if}
							<td>{loop=time_countdown}</td>
						</tr>
					{/foreach}
				</table>
			{/if}

			<div class="clear"></div>

			<div class="sub-links">
				{foreach[footerMenu]}
					{loop}link{/loop}
				{/foreach}
			</div>

			{include}"game_counters"{/include}
		</div>
	</div>
</body>
</html>
{if[0]}
 Oxsar https://oxsar-nova.ru
 Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
	
{/if}