<script type="text/javascript"><!--
//<![CDATA[
	{if[defined('SN') && !defined('SN_FULLSCREEN')]}
	$(function(){
		var d_height	= $('body').height() - 20;
		var d_width		= $('body').width() - 20;
		d_height = {const=MAX_HEIGHT};
		$('#falang-scan').dialog({
			autoOpen:		false,
			position:		'center',
			modal: 			true,
			closeOnEscape: 	true,
			draggable:		false,
			resizable:		false,
			height:			d_height,
			width:			d_width
		});
	});
	{/if}

	function sendFleet(mode, target)
	{
		url = "{const=RELATIVE_URL}game.php/go:FleetAjax/target:" + target + "/mode:" + mode+"?<?php echo Yii::app()->socialAPI->getSuffix();?>";
		$.get(url, function(data) {
			$("#ajaxresponse").html(data);
		});
	}

	function openWindow(id)
	{
		{if[defined('SN') && !defined('SN_FULLSCREEN')]}
			$('#falang-scan').dialog('open');
			var data = '<iframe src ="{const=RELATIVE_URL}game.php/MonitorPlanet/'+id+'?<?php echo Yii::app()->socialAPI->getSuffix();?>" width="100%" height="100%" id="falang_IFrame"></iframe>'
			$('#falang-scan').html(data);
			$('#falang_IFrame').load(function(){
				$(this).contents().find('a').attr('target', '_top');
			});
		{else}
			win = window.open("{const=RELATIVE_URL}game.php/MonitorPlanet/"+id+"?<?php echo Yii::app()->socialAPI->getSuffix();?>", "", "width=700,height=400,status=yes,scrollbars=yes,resizable=yes");
			win.focus();
		{/if}
	}

	function galaxySubmit(type)
	{
		theForm = document.forms['galaxy_form'];
		theForm.submittype.value = type;
		theForm.submit();
	}
//]]>
--></script>
{if[defined('SN') && !defined('SN_FULLSCREEN')]}
<div id="falang-scan" title="{lang}FALANG_SCAN{/lang}">
</div>
{/if}
<form method="post" name="galaxy_form" action="<?php echo socialUrl(RELATIVE_URL.'game.php/Galaxy'); ?>">
<input type="hidden" name="submittype" value="" />
<div id="ajaxresponse" class="idiv"></div>
<div class="idiv">
	<table class="ntable galaxy-browser">
		<tr>
			<th colspan="3">{lang}GALAXY{/lang}</th><th colspan="3">{lang}SYSTEM{/lang}</th>
		</tr>
		<tr>
			<td>
				<input type="button" name="prevgalaxy" value="&laquo;" class="button" onclick="galaxySubmit('prevgalaxy');" />
			</td>
			<td>
				<input type="text" name="galaxy" value="{@galaxy}" size="3" maxlength="2" class="center" onblur="checkNumberInput(this, 1, {const=NUM_GALAXYS});" />
			</td>
			<td>
				<input type="button" name="nextgalaxy" value="&raquo;" class="button" onclick="galaxySubmit('nextgalaxy');" />
			</td>

			<td>
				<input type="button" name="prevsystem" value="&laquo;" class="button" onclick="galaxySubmit('prevsystem');" />
			</td>
			<td>
				<input type="text" name="system" value="{@system}" size="3" maxlength="3" class="center" onblur="checkNumberInput(this, 1, {const=NUM_SYSTEMS});" />
			</td>
			<td>
				<input type="button" name="nextsystem" value="&raquo;" class="button" onclick="galaxySubmit('nextsystem');" />
			</td>
		</tr>
		<tr>
			<td colspan="6" class="center"><input type="submit" name="jump" value="{lang}COMMIT{/lang}" class="button" /></td>
		</tr>
	</table>
</div>
</form>
<table class="ntable">
	<thead><tr>
		<th colspan="{if[!isMobileSkin()]}8{else}7{/if}">{lang}SUNSYSTEM{/lang} {@galaxy}:{@system}</th>
	</tr>
	<tr>
		<th>#</th>
		<th colspan="{if[!isMobileSkin()]}3{else}2{/if}">{lang}PLANET{/lang}</th>
		<!-- <th>{lang}NAME{/lang}</th> -->
		<!-- <th>{lang}TF{/lang}</th> -->
		<th>{lang}MOON{/lang}</th>
		<th>{lang}USER{/lang}</th>
		<th>{lang}ALLIANCE{/lang}</th>
		<th>{lang}ACTIONS{/lang}</th>
	</tr></thead>
	<tfoot><tr>
		<td colspan="{if[!isMobileSkin()]}8{else}7{/if}">
			<p class="legend"><cite><span>i</span> = {lang}LOWER_INACTIVE{/lang}</cite><cite><span>I</span> = {lang}UPPER_INACTIVE{/lang}</cite><cite><span class="banned">b</span> = {lang}BANNED{/lang}</cite><cite><span class="strong-player">s</span> = {lang}STRONG_PLAYER{/lang}</cite><cite><span class="weak-player">n</span> = {lang}NEWBIE{/lang}</cite><cite><span class="vacation-mode">v</span> = {lang}VACATION_MODE{/lang}</cite></p>
			<p class="legend"><cite><span class="ownPosition">{lang=ONESELF}</span></cite><cite><span class="alliance">{lang=ALLIANCE}</span></cite><cite><span class="friend">{lang=FRIEND}</span></cite><cite><span class="enemy">{lang=ENEMY}</span></cite><cite><span class="confederation">{lang=CONFEDERATION}</span></cite><cite><span class="trade-union">{lang=TRADE_UNION}</span></cite><cite><span class="protection">{lang=PROTECTION_ALLIANCE}</span></cite></p>
		</td>
	</tr></tfoot>
	<tbody>{foreach[sunsystem]}<tr>
		<td class="center">{loop}systempos{/loop}</td>
		{if[!isMobileSkin()]}
		<td class="center">{loop}picture{/loop}</td>
		{/if}
		{if[$row["metal"] || $row["silicon"]]}
		<td class="left">{if[$row["destroyed"]]}<i>{/if}{loop}planetname{/loop}{if[$row["destroyed"]]}</i>{/if} {if[$row["activity"]]}<nobr>{loop}activity{/loop}</nobr>{/if}</td>
		<td class="center">
			<script type="text/javascript">//<![CDATA[
				var debris_{loop}systempos{/loop} = "<table class=ttable><tr><td rowspan=&quot;3&quot;>{@debris}</td><th colspan=&quot;2&quot;>{lang}RESOURCES{/lang}</th></tr><tr><td>{lang}METAL{/lang}:</td><td>{loop}metal{/loop}</td></tr><tr><td>{lang}SILICON{/lang}:</td><td>{loop}silicon{/loop}</td></tr></table>";
			//]]></script>
			<span style="cursor: pointer;" onmouseover="Tip(debris_{loop}systempos{/loop}, TITLE, '{lang}DEBRIS{/lang}', FADEIN, 300, FADEOUT, 300, STICKY, 1, CLOSEBTN, true);" onmouseout="UnTip();">{loop}debris{/loop}</span>
		</td>
		{else}
		<td class="left" colspan="2">{if[$row["destroyed"]]}<i>{/if}{loop}planetname{/loop}{if[$row["destroyed"]]}</i>{/if} {if[$row["activity"]]}<nobr>{loop}activity{/loop}</nobr>{/if}</td>
		{/if}
		<td class="center">
			{if[$row["moonpicture"] != ""]}
			<script type="text/javascript">//<![CDATA[
				var moon_{loop}systempos{/loop} = "<table class=ttable><tr><td rowspan=&quot;3&quot;>{@moon}</td><th colspan=&quot;2&quot;>{lang}FEATURES{/lang}</th></tr><tr><td>{lang}SIZE{/lang}</td>"
					+ "<td>{loop}moonsize{/loop}km</td></tr><tr><td>{lang}TEMPERATURE{/lang}</td><td>{loop}moontemp{/loop} &deg;C</td></tr>"
					+ "{if[$row['moonid']]}<tr><td colspan='3'><a href='#' onclick='sendFleet({const=EVENT_SPY}, {loop}moonid{/loop}); return false'>{lang=SEND_ESPIONAGE_PROBE}</a></td><td>{/if}"
					+ "{if[$row['moonrocket']]}<tr><td colspan='3'>{loop=moonrocket}</td><td>{/if}"
					+ "</table>";
			//]]></script>
			<span style="cursor: pointer;" onmouseover="Tip(moon_{loop}systempos{/loop}, TITLE, '{loop}moon{/loop}', FADEIN, 300, FADEOUT, 300, STICKY, 1, CLOSEBTN, true);" onmouseout="UnTip();">{loop}moonpicture{/loop}</span>
			{/if}
		</td>
		<td class="center normal">{loop}username{/loop} {if[$row["user_status_long"] != ""]}({loop}user_status_long{/loop}){/if}{if[!empty($row["userid"])]}<br /><span class="galaxysub">#{loop}rank{/loop} / {loop}cur_points{/loop}{if[$row["e_points_num"]]} / {loop}e_points{/loop}{/if}</span>{/if}</td>
		<td class="center">
			{if[$row["alliance"] != ""]}
			<script type="text/javascript">//<![CDATA[
				var ally_{loop}systempos{/loop} = "<table class=ttable><tr><th>{loop}allydesc{/loop}</th></tr><tr><td>{loop}allypage{/loop}</td></tr>{loop}homepage{/loop}{loop}memberlist{/loop}</table>";
			//]]></script>
			<a href="javascript:void(0);" onmouseover="Tip(ally_{loop}systempos{/loop}, TITLE, '{lang}ALLIANCE{/lang}', FADEIN, 200, FADEOUT, 200, STICKY, 1, CLOSEBTN, true);" onmouseout="UnTip();">{loop}alliance{/loop}</a>{if[!empty($row["userid"])]}<br /><span class="galaxysub">#{loop}alliance_rank{/loop}</span>{/if}
			{/if}
		</td>
		<td class="center">
			{if[$row["userid"] != Core::getUser()->get("userid") && $row["userid"]]}
				<span class="pointer" onclick="sendFleet({const=EVENT_SPY}, {loop}planetid{/loop});">{@sendesp}</span>
				{loop}message{/loop}
				{loop}buddyrequest{/loop}
				{loop}rocketattack{/loop}
				{if[{var}canMonitorActivity{/var}]}<span onclick="openWindow({loop}planetid{/loop});" class="pointer">{@monitorfleet}</span>{/if}
			{/if}
		</td>
	</tr>{/foreach}
    {if[EXPEDITION_ENABLED]}
	<tr><td colspan="8" align="center"><a href="{@goMissionLink}g:{@galaxy}/s:{@system}/p:{const=EXPED_PLANET_POSITION}?<?php echo Yii::app()->socialAPI->getSuffix();?>">{lang=SUNSYSTEM_OUTSIDE}</a></td></tr>
    {/if}
	</tbody>
</table>
<script type="text/javascript" src="{const=RELATIVE_URL}js/lib/wz_tooltip.js"></script>
{if[0]}
 Oxsar http://oxsar.ru
 Copyright (c) 2009-2010 UnitPoint <support@unitpoint.ru>

{/if}