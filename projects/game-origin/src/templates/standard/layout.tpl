
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

<style>
body {
	background-color: black;
}
</style>

<link rel="stylesheet" type="text/css" href="{const=RELATIVE_URL}css/layout.css?{const=CLIENT_VERSION}" />
<link rel="stylesheet" type="text/css" href="{const=RELATIVE_URL}css/style.css?{const=CLIENT_VERSION}" />

<script type="text/javascript" src="http://ajax.googleapis.com/ajax/libs/jquery/1.5.1/jquery.min.js"></script>
<script type="text/javascript" src="http://ajax.googleapis.com/ajax/libs/jqueryui/1.8.14/jquery-ui.js"></script>

<script type="text/javascript" src="{const=RELATIVE_URL}js/lib/jquery.cookie.js?{const=CLIENT_VERSION}"></script>

<script type="text/javascript">
var has_left_menu = true;
var has_top_menu = true;
{if[isFacebookSkin()]}
var has_top_menu = false;
{/if}
{if[isMobiSkin()]}
var has_left_menu = false;
{/if}
var menu_done = false;
var galaxy_distance_mult = 1;
</script>
<script type="text/javascript" src="{const=RELATIVE_URL}js/main.js?{const=CLIENT_VERSION}"></script>
{if[0]}{* OAuth социальных сетей убран из oxsar-nova *}{/if}

{if[0]}<script type="text/javascript" src="http://jquery-ui.googlecode.com/svn/tags/latest/external/bgiframe/jquery.bgiframe.min.js"></script>{/if}
<link class="ui-theme" rel="Stylesheet" type="text/css" href="http://ajax.googleapis.com/ajax/libs/jqueryui/1.8.14/themes/ui-darkness/jquery-ui.css" />
{if[0]}<link rel="stylesheet" type="text/css" href="{const=RELATIVE_URL}/css/ui-darkness/jquery-ui.css">{/if}

{@HTML_HEAD}

{if[defined('SN') && !defined('SN_FULLSCREEN') && SN == VKNT_SN_ID]}

<link rel="stylesheet" type="text/css" href="{const=RELATIVE_URL}css/skin/vkontakte.css?{const=CLIENT_VERSION}">

{else}

{if[isFacebookSkin()]}
<link rel="stylesheet" type="text/css" href="{const=RELATIVE_URL}css/skin/fb.css?{const=CLIENT_VERSION}">
{/if}

{if[isMobiSkin()]}
<link rel="stylesheet" type="text/css" href="{const=RELATIVE_URL}css/skin/mobi.css?{const=CLIENT_VERSION}">
{/if}

{/if}

</head>
<body>

{if[defined('SN') && !defined('SN_FULLSCREEN')]}
<div id='SN_warper' style='overflow-y: auto;'>
{/if}
<div id="contentHtml" align="center">
<div class="main_content">

{hook}HtmlBegin{/hook}

{if[ !isFacebookSkin() && !isMobiSkin() ]}
<div id="leftMenu">
	<ul>
		{if[0 && isFacebookSkin()]}<li><b>{@currentPlanet}</b></li>{/if}
		{foreach[navigation]}<li{loop}inner{/loop}>{loop}content{/loop}</li>{if[!empty($row["outer_bot"])]}{loop}outer_bot{/loop}{/if}{/foreach}
		{hook}ListMenu{/hook}
		<li><nobr>Oxsar {const=OXSAR_VERSION}</nobr></li>
	</ul>
	{perm[CAN_EDIT_CONSTRUCTIONS]}
		<div style="text-align:center">SQL-Queries: {SQLQueries}</div>
	{/perm}
	{if[0]}
		<!-- DO NOT MODIFY "POWERED BY" DIV, feel free to change "poweredby" css class -->
		{if[0]}<div class="poweredby"><a href="http://netassault.ru" target="_blank">Powered by NetAssault</a></div>{/if}
	{/if}
	{include}"menu_counters"{/include}
	{if[ TUTORIAL_ENABLED ]}
	<div style="text-align: center;">
		<img src="/images/system-help-big.png" id='show_me_tutorial' style="margin-left: auto;margin-right: auto;">
	</div>
	{/if}
</div>
{else if[ isFacebookSkin() ]}
<div id="leftMenu">
	<ul>
		{if[0 && isFacebookSkin()]}<li><b>{@currentPlanet}</b></li>{/if}
		{foreach[navigation]}<li{loop}inner{/loop}>{loop}content{/loop}</li>{if[!empty($row["outer_bot"])]}{loop}outer_bot{/loop}{/if}{/foreach}
		{hook}ListMenu{/hook}
		<li><nobr>Oxsar {const=OXSAR_VERSION}</nobr></li>
	</ul>
	{perm[CAN_EDIT_CONSTRUCTIONS]}
		<div style="text-align:center">SQL-Queries: {SQLQueries}</div>
	{/perm}
	{if[0]}
		<!-- DO NOT MODIFY "POWERED BY" DIV, feel free to change "poweredby" css class -->
		{if[0]}<div class="poweredby"><a href="http://netassault.ru" target="_blank">Powered by NetAssault</a></div>{/if}
	{/if}
	{include}"menu_counters"{/include}
</div>
<script type="text/javascript">
	$(function(){
		$('.planet_selector').change(function() {
			gotoPlanet('planetSelection', $('option:selected',$(this)).attr('rel'));
		});
	});
</script>
<div id="topMenu">
	<table class="top_menu">
		<tr>
		{if[TUTORIAL_ENABLED && isFacebookSkin()]}
			<td><img src="/images/system-help.png" id='show_me_tutorial'></td>
		{/if}
		{foreach[menu_headers]}
			{if[$row["link"]]}
				<td class='{loop=class}'>
					{loop=link}
				</td>
			{else if[$row['pselect']]}
				{if[ NS::getPlanet()->getPlanetId() ]}
				<td class='{loop=class}'>
					<select class='planet_selector' style="width: 100%;">
						{foreach2[.planets]}
							<option rel='{loop=planetid}'{if[NS::getPlanet()->getPlanetId() == $row['planetid']]} selected{/if}>
								{loop=coords} {loop=planetname}
							</option>
							{if[$row['moonid']]}
							<option rel='{loop=moonid}'{if[NS::getPlanet()->getPlanetId() == $row['moonid']]}selected{/if}>
								{loop=coords} {loop=moon}
							</option>
							{/if}
						{/foreach2}
					</select>
				</td>
				{/if}
			{else}
				<td class='{loop=class}'>
					{loop=name}
				</td>
			{/if}
		{/foreach}
		{if[defined('SN') && !defined('SN_FULLSCREEN')]}
			<td><a href="{if[1]}{else}{@formaction}&sn_fullscreen=1{/if}" target="_blank">{image[FULLSCREEN]}fullscreen-icon.png{/image}</a></td>
		{else if[defined('SN') && defined('SN_FULLSCREEN')]}
			<td><a href="#" onclick="window.close(); return false" target="_blank">{image[FULLSCREEN_CLOSE]}fullscreen-close-icon.png{/image}</a></td>
		{/if}
		</tr>
	</table>
</div>
{else}
<div id="topMenu">
	<table class="top_menu">
		<tr>
		{if[0]}<td>{@currentPlanet}</td>{/if}
		{foreach[menu_headers]}
			{if[$row["link"]]}
				<td class='{loop=class}'>
					{loop=link}
				</td>
			{else if[$row['pselect']]}
				{if[ NS::getPlanet()->getPlanetId() ]}
				<td class='{loop=class}'>
					<select class='planet_selector' style="width: 100%;">
						{foreach2[.planets]}
							<option rel='{loop=planetid}'{if[NS::getPlanet()->getPlanetId() == $row['planetid']]} selected{/if}>
								{loop=coords} {loop=planetname}
							</option>
							{if[$row['moonid']]}
							<option rel='{loop=moonid}'{if[NS::getPlanet()->getPlanetId() == $row['moonid']]}selected{/if}>
								{loop=coords} {loop=moon}
							</option>
							{/if}
						{/foreach2}
					</select>
				</td>
				{/if}
			{else}
				<td class='{loop=class}'>
					{loop=name}
				</td>
			{/if}
		{/foreach}
		{if[defined('SN') && !defined('SN_FULLSCREEN')]}
			<td><a href="{@formaction}&sn_fullscreen=1" target="_blank">{image[FULLSCREEN]}fullscreen-icon.png{/image}</a></td>
		{else if[defined('SN') && defined('SN_FULLSCREEN')]}
			<td><a href="#" onclick="window.close(); return false" target="_blank">{image[FULLSCREEN_CLOSE]}fullscreen-close-icon.png{/image}</a></td>
		{/if}
		{if[ TUTORIAL_ENABLED ]}
		<td><img src="/images/system-help.png" id='show_me_tutorial'></td>
		{/if}
		</tr>
	</table>
	<script type="text/javascript">
		$(function(){
			var m_opened = false;
			$('.navigation_header').click(
				function(){
					if(m_opened)
					{
						$('.navigation').css('display','none');
						m_opened = false;
					}
					else
					{
						$('.navigation').css('display','block');
						m_opened = true;
					}
				}
			);
			$(function(){
				$('.planet_selector').change(function() {
					gotoPlanet('planetSelection', $('option:selected',$(this)).attr('rel'));
				});
			});
		});
	</script>
	<div class='navigation'>
		<table>
		<tr>
		{foreach[navigation]}
			<td class='main_menu'>
				<ul>
					<?php $block_num = 0; ?>
					{foreach2[.items]}
						{if[ $block_num++ > 0]}
						<li {loop=inner}>
							{if[0]}{loop=title}{/if}
						</li>
						{/if}
						{foreach3[..childs]}
							<li>
								{loop=content}
							</li>
						{/foreach3}
					{/foreach2}
				</ul>
			</td>
		{/foreach}
		</tr>
		</table>
	</div>
</div>
{/if}

<div id="contentTopAndBody">

	<div id="topHeader" align=center>
	 <table width="auto" cellpadding=0 cellspacing=0 border=0 class="top_header_res">
	  <tr>
			{if[ !isFacebookSkin() && !isMobiSkin() ]}<td class="header-planet-name">
				{if[defined('SN') && defined('SN_FULLSCREEN')]}
					<a href="#" onclick="window.close(); return false" target="_blank">{image[FULLSCREEN_CLOSE]}fullscreen-close-icon.png{/image}</a>
					<br />
				{/if}
				<b>{@currentPlanet}</b> {@currentCoords}
                <br/>{if[DEATHMATCH]}<span class='false'>{const=UNIVERSE_NAME}</span>{/if}
                <br/>Склад:</td>
			{/if}
			<td class="header-resource">
				{if[!isMobileSkin()]}{image[METAL]}met.gif{/image}<br />{/if}
				<span class="ressource">{lang}METAL{/lang}</span>
				<br />
				<span id='header_layout_metal' class="{@metalClass}">{@real_metal}</span>
				<br/>
				{@stor_metal}
			</td>
			<td class="header-resource">
				{if[!isMobileSkin()]}{image[SILICON]}silicon.gif{/image}<br />{/if}
				<span class="ressource">{lang}SILICON{/lang}</span>
				<br />
				<span id='header_layout_silicon' class="{@siliconClass}">{@real_silicon}</span>
				<br/>
				{@stor_silicon}
			</td>
			<td class="header-resource">
				{if[!isMobileSkin()]}{image[HYDROGEN]}hydrogen.gif{/image}<br />{/if}
				<span class="ressource">{lang}HYDROGEN{/lang}</span>
				<br />
				<span id='header_layout_hydrogen' class="{@hydrogenClass}">{@real_hydrogen}</span>
				<br/>
				{@stor_hydrogen}
			</td>
			<td class="header-resource">
				{if[!isMobileSkin()]}{image[ENERGY]}energy.gif{/image}<br />{/if}
				<span class="ressource">{lang}ENERGY{/lang}</span>
				<br />
				<span id='header_layout_energy' class="{@energyClass}">{@planetAEnergy} ({@remainingEnergy})</span>
				<br/>
				{@planetEnergy}
			</td>
			<td class="header-resource">
				{if[!isMobileSkin()]}<img src="{const=RELATIVE_URL}images/credit.gif" alt="Игровая валюта"/><br />{/if}
				<span class="ressource">Кредиты</span>
				<br />
				<span id='header_layout_credit'>{@credit}</span>
				{if[OXSAR_RELEASED]}
					<br />
					{link[CREDIT_PAY]}"game.php/Payment"{/link}
				{/if}
			</td>
	  </tr>
	 </table>
	 {include}"before_content"{/include}
	</div>

	<div class="clear"></div>

	<div id="content">
	<table width="100%" cellpadding=0 cellspacing=0 border=0>
		{if[ defined('SN') && defined('SN_FULLSCREEN') && SN == MAILRU_SN_ID ]}
		<tr>
			<td width="100%" align="center" style="margin-bottom:20px">
				{include}"appgrade-mailru-code"{/include}
			</td>
		</tr>
		{/if}
		{if[ defined('SN') && defined('SN_FULLSCREEN') && SN == VKNT_SN_ID ]}
		<tr>
			<td width="100%" align="center" style="margin-bottom:20px">
				{include}"appgrade-vkontakte-code"{/include}
			</td>
		</tr>
		{/if}
		{if[ !defined('SN') && showTopBannerBlock() ]}
			<?php if(0): ?>
			<tr>
				<td width="100%" align="center" style="padding-bottom:10px">
					<?php /* план 37.5d.12: CHtml::link → raw HTML (Yii убран) */ ?>
					<a href="<?php echo htmlspecialchars(socialUrl(RELATIVE_URL . 'game.php/Payment')); ?>" class="false2">Эта рекламу можно убрать пополнив запасы кредитов</a>
				</td>
			</tr>
			<?php endif; ?>
		<tr>
            <td width="100%" align="center" style="margin-bottom:20px">
			{include}"banner_top"{/include}
			</td>
		</tr>
		{/if}
		<tr>
			<td width="100%" align="center">
				<table cellpadding=0 cellspacing=0 border=0>
					<tr>
						<td align="left">
                            <?php /* Yii-виджеты (Prize/Newbie/News/Notify) убраны — заменим в plan-37 на нативные блоки */ ?>
							{include}'technicalWorks'{/include}
							{hook}ContentStarts{/hook}
							{if[{var}delete{/var}]}<div class="warning">{@delete}</div>{/if}
							{if[{var}umode{/var}]}<div class="info">{@umode}</div>{/if}
							{if[{var}LOG{/var}]}{@LOG}{/if}
							{include}"tutorial"{/include}
							{include}$template{/include}
							{hook}ContentEnds{/hook}
						</td>
					</tr>
				</table>
			</td>
		</tr>
		{if[ !defined('SN') && showBottomBannerBlock() ]}
		<tr>
			<td style="padding-top:10px; padding-bottom:5px; text-align:center">
			{include}"banner_bottom"{/include}
			</td>
		</tr>
		{/if}
	</table>
	</div>

</div>

{if[ !isFacebookSkin() && !isMobiSkin() ]}
<div id="planets">
<ul>
	{foreach[planetHeaderList]}
		<li lang="{loop}planetid{/loop}" class="goto{if[$row["planetid"] == NS::getPlanet()->getPlanetId()]} cur-planet{/if}">
			{loop}picture{/loop}<br />{loop}planetname{/loop}
		</li>
		{if[$row["moonid"] > 0]}
			<li lang="{loop}moonid{/loop}" class="goto {if[$row["moonid"] == NS::getPlanet()->getPlanetId()]}cur-moon{else}moon-select{/if}">
				{loop}mpicture{/loop}<br />{loop}moon{/loop}
			</li>
		{/if}
	{/foreach}
</ul>
</div>
{/if}
{hook}HtmlEnd{/hook}

</div>
</div>

{if[0 && USE_CHAT_TEST && (!defined('IS_MOBILE_REQUEST') || !IS_MOBILE_REQUEST)]}

<script type="text/javascript">
  $(function()
  {
  $('#chat_window').dialog({
      bgiframe: true,
      autoOpen: false,
      draggable: true,
		  width: 700,
		  height: 500,
		  position: ['left','bottom'],
		  modal: false,
		  // show: 'blind'
		}
  )
  });

	function open_chat_window()
	{
		$('#chat_container').load("{const=RELATIVE_URL}chatPro.php?sid={const=SID}");
		$('#chat_window').dialog('open');
	}

</script>

<div id="chat_window" title="Chat" style="display:none">
  <div style="position:relative; width:100%; height:98%">
    <div id="chat_container" style="position:absolute; width:100%; height:98%"
      class="ui-helper-reset ui-helper-clearfix">
    </div>
  </div>
</div>
{/if}

<form method="post" id="planetSelection" action="{@formaction}" style="display: none;"><input type="hidden" name="planetid" value="0" id="planetid" /></form>

<div class="oxsar-footer">
	<div class="age-rating" title="Возрастная категория 12+">12+</div>
	<div class="legal-links">
		<a href="https://oxsar-nova.ru/offer" target="_blank" rel="noopener">Оферта</a>
		<span class="sep">|</span>
		<a href="https://oxsar-nova.ru/game-rules" target="_blank" rel="noopener">Правила</a>
		<span class="sep">|</span>
		<a href="https://oxsar-nova.ru/refund" target="_blank" rel="noopener">Возврат</a>
		<span class="sep">|</span>
		<a href="https://oxsar-nova.ru/privacy" target="_blank" rel="noopener">Конфиденциальность</a>
	</div>
</div>

{if[defined('SN') && !defined('SN_FULLSCREEN')]}
</div>
{/if}
<?php /* TutorialDialog widget убран — заменим в plan-37 на нативный блок */ ?>
</body>
</html>
{if[0]}
 Oxsar https://oxsar-nova.ru
 Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.

{/if}