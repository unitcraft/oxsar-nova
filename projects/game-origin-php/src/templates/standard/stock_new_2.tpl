<script type="text/javascript">
//<![CDATA[
var curTime = new Date();
var delay = {const}TIME{/const} - Math.round(curTime.getTime() / 1000);
var oGalaxy = {@oGalaxy};
var oSystem = {@oSystem};
var oPos = {@oPos};
var gamespeed = {const=GAMESPEED};
//var maxspeed = {@maxspeedVar};
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
$(document).ready(function() {
	rebuildLot();
});
//]]>
</script>
<form method="post" action="{@formaction}">
<table class="ntable">
	<tr>
		<th colspan="2">{lang}EXCH_LOT_OPTIONS{/lang}</th>
	</tr>
        <tr>
		      <td nowrap="nowrap"><label for="ttl">{lang}EXCH_TTL{/lang}</label></td>
              <td><input type="text" name="ttl" id="ttl" size="4" value="{@ttl}" /></td>
        </tr>
        <tr>
		      <td nowrap="nowrap"><label for="lot_type">{lang}EXCH_LOT_TYPE{/lang}</label></td>
		      <td><select name="lot_type" id="lot_type" onchange="javascript:rebuildLot();">{@lot_types}</select></td>
        </tr>
	<tr>
		<td>{lang}EXCH_DELIVERY_HYDRO{/lang}</td>
                <td><input type="text" name="delivery_hydro" id="delivery_hydro" onchange="javascript:rebuildLot();" value="1000"/></td>
	</tr>
        <tr>
		<td>{lang}EXCH_DELIVERY_PERCENT{/lang}</td>
                <td><input type="text" name="delivery_percent" id="delivery_percent" size="3" maxlength="3" value="50"/></td>
	</tr>
	<tr>
		<td>{lang}CAPICITY{/lang}</td>
		<td><span id="capicity">{@capicity}</span></td>
	</tr>
	<tr>
		<td>{lang}FLEET{/lang}</td>
		<td>{@fleet}</td>
	</tr>
	<tr>
		<td colspan="2" class="center"><input type="submit" name="step3" value="{lang}NEXT{/lang}" class="button" /></td>
	</tr>
</table>
</form>
{if[0]}
 Oxsar https://oxsar-nova.ru
 Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.

{/if}