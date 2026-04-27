<script type="text/javascript">
//<![CDATA[
var curTime = new Date();
var delay = {const}TIME{/const} - Math.round(curTime.getTime() / 1000);
var oGalaxy = {@oGalaxy};
var oSystem = {@oSystem};
var oPos = {@oPos};
var gamespeed = {const=GAMESPEED};
var maxspeed = {@maxspeedVar};
var basicConsumption = {@basicConsumption};
var unitsGroupConsumptionPowerBase = {const=UNITS_GROUP_CONSUMTION_POWER_BASE};
var maxGroupUnitConsumptionPerHour = {const=MAX_GROUP_UNIT_CONSUMTION_PER_HOUR};
var fleetSize = {@fleetSize};
var capicity = {@capicity_raw};
var decPoint = '{lang}DECIMAL_POINT{/lang}';
var thousandsSep = '{lang}THOUSANDS_SEPERATOR{/lang}';
var maxGalaxy = {const=NUM_GALAXYS};
var maxSystem = {const=NUM_SYSTEMS};
var maxPos = {const}MAX_POSITION{/const};
var expPlanetPos = {const}EXPED_PLANET_POSITION{/const};
var expVirtPlanetPos = {const}POSITION_TO_CALC_EXP_TO{/const};
var allow_stargate_transport = '{@allow_stargate_transport}';
var stargate_transport_time_scale = {const=STARGATE_TRANSPORT_TIME_SCALE};
var stargate_transport_consumption_scale = {const=STARGATE_TRANSPORT_CONSUMPTION_SCALE};
$(document).ready(function() {
	{foreach[invitations]}
	$('#timer_{loop=eventid}').countdown({until: {loop=time_r}, compact: true, onExpiry: function() {
		$('#timer_{loop=eventid}').text('-');
	}});
	{/foreach}
	rebuild();
});
//]]>
</script>
<form method="post" action="{@formaction}">
{if[ 0 && {var=holding_eventid} ]}
<input type="hidden" name="id" value="{@holding_eventid}" />
{/if}
<table class="ntable">
	{if[!{var}can_send_fleet{/var}]}
		<tr>
			<th colspan="2">{lang}NO_FREE_FLEET_SLOTS{/lang}</th>
		</tr>
	{/if}
	{if[!{var=holding_eventid} && !{var}can_send_expo{/var} && NS::getResearch(UNIT_EXPO_TECH) > 0]}
		<tr>
			<th colspan="2">{lang}NO_FREE_EXPO_SLOTS{/lang}</th>
		</tr>
	{/if}
	<tr>
		<th colspan="2">{lang}SELECT_TARGET{/lang}</th>
	</tr>
	<tr>
		<td>{lang}TARGET{/lang}</td>
		<td>
			<input type="text" name="galaxy" id="galaxy" value="{@galaxy}" size="3" maxlength="2" onblur="javascript:rebuild();" />
			<input type="text" name="system" id="system" value="{@system}" size="3" maxlength="3" onblur="javascript:rebuild();" />
			<input type="text" name="position" id="position" value="{@position}" size="3" maxlength="2" onblur="javascript:rebuild();" />
			<select name="targetType" onchange="javascript:rebuild();"  id="targetType">
				<option value="planet">{lang}PLANET{/lang}</option>
				<option value="tf">{lang}TF{/lang}</option>
				<option value="moon">{lang}MOON{/lang}</option>
			</select>
		</td>
	</tr>
	<tr>
		<td>{lang}DISTANCE{/lang}</td>
		<td><span id="distance">{@distance}</span></td>
	</tr>
	<tr>
		<td>{lang}SPEED{/lang}</td>
		<td><input type="text" name="speed" size="3" maxlength="3" value="100" id="speed" onkeyup="javascript:rebuild();" />% <select onchange="setFromSelect('speed', this); rebuild();">{@speedFromSelectBox}</select></td>
	</tr>
	<tr>
		<td>{lang}TIME{/lang}</td>
		<td>
			<span id="time">{@time}</span>
			{if[ {var=allow_stargate_transport} ]}
				(<span id="stargate_transport_time">{@stargate_transport_time}</span> для гиперполета)
			{/if}
		</td>
	</tr>
	<tr>
		<td>{lang}FUEL{/lang}</td>
		<td>
			<span id="fuel">{@fuel}</span>
			{if[ {var=allow_stargate_transport} ]}
				(<span id="stargate_transport_fuel">{@stargate_transport_fuel}</span> для гиперполета)
			{/if}
		</td>
	</tr>
	<tr>
		<td>{lang}MAX_SPEED{/lang}</td>
		<td>{@maxspeed}{if[{var=speedbonus}]} {@speedbonus}{/if}</td>
	</tr>
	<tr>
		<td>{lang}CAPICITY{/lang}</td>
		<td><span id="capicity">{@capicity}</span></td>
	</tr>
	<tr>
		<td>{lang}FLEET{/lang}</td>
		<td>{@fleet}</td>
	</tr>
	{if[count($this->getLoop("shortlinks")) > 0]}<tr>
		<th colspan="2">{lang}SHORTLINKS{/lang}</th>
	</tr>
	{foreach[shortlinks]}{if[$key % 2 != 0]}<tr>{/if}
		<td><a class="pointer" onclick="javascript:setCoordinates({loop}galaxy{/loop}, {loop}system{/loop}, {loop}position{/loop}, {loop}type{/loop});"><strong>{loop}planetname{/loop} [{loop}galaxy{/loop}:{loop}system{/loop}:{loop}position{/loop}]</strong></a></td>
		{if[$count == $key && $key % 2 != 0]}<td></td>{/if}
	{if[$key % 2 == 0 || $count == $key]}</tr>{/if}{/foreach}{/if}
	{if[count($this->getLoop("invitations")) > 0]}<tr>
		<th colspan="2">{lang=FORMATION_INVATATIONS}</th>
	</tr>
	<tr>
		<td colspan="2">{foreach[invitations]}
			<span id="timer_{loop=eventid}">{loop=formatted_time}</span>
			&ndash; <a href="javascript:void(0);" onclick="javascript:setCoordinates({loop=galaxy}, {loop=system}, {loop=position}, {loop=type});" class="true pointer">[{loop=galaxy}:{loop=system}:{loop=position}] | {loop=name}</a><br />
		{/foreach}</td>
	</tr>{/if}
	<tr>
{if[ 0 && {var=holding_eventid} ]}
		<td colspan="2" class="center"><input type="submit" name="holding_select_mission" value="{lang}NEXT{/lang}" class="button" /></td>
{else}
		<td colspan="2" class="center"><input type="submit" name="step3" value="{lang}NEXT{/lang}" class="button" /></td>
{/if}
	</tr>
</table>
</form>
{if[0]}
 Oxsar https://oxsar-nova.ru
 Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.

{/if}