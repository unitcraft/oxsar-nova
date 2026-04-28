<script type="text/javascript">
//<![CDATA[
var fleet = new Array();
var quantities = new Array();
var n = 0;
{foreach[fleet]}
{if[$row["speed"] > 0]}
fleet[n++] = {loop=id};
quantities[{loop=id}] = {loop=quantity_raw};
{/if}
{/foreach}
function resetRest(id)
{
	if( $("#" + id).val() > {@max_ships} )
	{
		$("#" + id).val({@max_ships});
	}
    var value = $("#" + id).val();
    {if[{var=newLot}]}
    $("input.center").val("0");
    $("#" + id).val(value);
    {/if}
}
{if[{var=newLot}]}
$(function() {
	$("input.center[type='checkbox']").click(function(){
		if ( $(this).attr('checked') == true )
		{
			$("input.center[type='checkbox']").attr('checked', false);
			$(this).attr('checked', true);
		}
	});
});
{/if}
//]]>
</script>

{if[ {var=stargateCountDown} ]}
<table class="ntable">
	<thead>
		<th colspan="2">{lang}STAR_GATE_JUMP{/lang}</th>
	</thead>
	<tbody>
		<td>{@stargateCountDown}</td>
		<td>Заряжаются Межгалактические врата</td>
	</tbody>
</table>
{/if}

<table class="ntable">
	<thead>
	  <!-- <tr>
		  <th colspan="6">{lang}NEW_MISSION{/lang}</th>
	  </tr> -->
	  <tr>
		  <th colspan="2">{lang}SHIP_NAME{/lang}</th>
		  {if[isFacebookSkin() || isMobiSkin()]}
		  <th>
		  	{lang}SPEED{/lang}
		  	<br/>
		  	{lang}CAPICITY{/lang}
		  </th>
		  {else}
		  <th>{lang}SPEED{/lang}</th>
		  <th>{lang}CAPICITY{/lang}</th>
		  {/if}
		  <th>{lang}WAITING{/lang}</th>
		  <th></th>
		  <th>{lang}SELECTION{/lang}</th>
	  </tr>
	</thead>
	{if[count($this->getLoop("fleet")) > 0]}
<form method="post" action="{@formaction}">
<input type="hidden" name="galaxy" value="{request[get]}g{/request}" />
<input type="hidden" name="system" value="{request[get]}s{/request}" />
<input type="hidden" name="position" value="{request[get]}p{/request}" />
	<tbody>
    {foreach[fleet]}
    <tr>
	    <td width="1">{loop}image{/loop}</td>
	    <td>{loop}name{/loop}</td>
		{if[isFacebookSkin() || isMobiSkin()]}
	    <td>
		    {loop}speed{/loop}
		    <br/>
		    {loop}capicity{/loop}
	    </td>
	    {else}
	    <td>{loop}speed{/loop}</td>
	    <td>{loop}capicity{/loop}</td>
	    {/if}
	    <td class="center">{loop}quantity{/loop}</td>
	    <td class="center">
			{if[$row["speed"] > 0 && !{var=observer}]}
				{if[!isMobiSkin()]}
					<a href="#" onclick="setField('ship_{loop}id{/loop}', {loop}quantity_raw{/loop}); resetRest('ship_{loop}id{/loop}'); return false">{lang}max{/lang}</a>
					<br /> <a href="#" onclick="setField('ship_{loop}id{/loop}', Math.ceil({loop}quantity_raw{/loop}/2)); resetRest('ship_{loop}id{/loop}'); return false">{lang}50%{/lang}</a>
					<br /> <a href="#" onclick="setField('ship_{loop}id{/loop}', 0); return false">{lang}min{/lang}</a>
				{else}
					<select onchange="setField('ship_{loop}id{/loop}', Math.ceil({loop}quantity_raw{/loop} * $(this).val() / 100))">
						<option value="100">{lang}max{/lang}</option>
						<option value="50">{lang}50%{/lang}</option>
						<option value="0">{lang}min{/lang}</option>
					</select>
				{/if}
			{/if}
		</td>
	    <td align="center">{if[$row["speed"] > 0 && !$row["blocked"] && !{var=observer}]}<input class="center" type="text" name="{loop}id{/loop}" value="0" size="{@max_ships_size}" maxlength="{@max_ships_grade}" id="ship_{loop}id{/loop}" onchange="resetRest(this.id);"/>{/if}</td>
    </tr>
    {/foreach}
	</tbody>
    {if[!{var=newLot} && !{var=observer}]}
    <tr>
		<td colspan="7" class="center">
		    <table class="table_no_background" cellspacing="0" cellpadding="0" border="0" width="100%">
		     <tr>
		      <td align="center"><a href="javascript:void(0);" onclick="selectShips();">{lang}ALL_SHIPS{/lang}</a></td>
		      <td align="center"><a href="javascript:void(0);" onclick="deselectShips();">{lang}NO_SHIPS{/lang}</a></td>
		     </tr>
		    </table>
	    </td>
	</tr>
    {/if}
	{else}
	<tbody>
	  <tr>
		  <td class="center" colspan="2">-</td>
		  <td class="center">-</td>
		  <td class="center">-</td>
		  <td class="center">-</td>
		  <td class="center">-</td>
		{if[isFacebookSkin() || isMobiSkin()]}
	    {else}
		  <td class="center">-</td>
	    {/if}
	  </tr>
	</tbody>
	{/if}

	{if[$this->getLoop("artefacts")]}
	<tr>
		<th colspan="7">{lang=ARTEFACTS}</th>
	</tr>
	<tbody>
		{foreach[artefacts]}
		<tr>
			<td width="1">{loop}image{/loop}</td>
			<td>{loop}name{/loop}</td>
			{if[!isFacebookSkin() && !isMobiSkin()]}
			<td class="center">{loop}flags{/loop}</td>
			{/if}
			<td>{loop}disappear_counter{/loop}</td>
			<td class="center">{loop}times_left{/loop} {if[$row['times_left']>=2 && $row['times_left']<=4]}раза{else}раз{/if}</td>
			{if[!isFacebookSkin() && !isMobiSkin()]}
			<td>&nbsp;</td>
			{else}
			<td class="center">{loop}flags{/loop}</td>
			{/if}
			<td class="center"><input class="center" type="checkbox" name="art[{loop}artid{/loop}]" id="art[{loop}artid{/loop}]" /></td>
		</tr>
		{/foreach}
	</tbody>
	{/if}

	{if[count($this->getLoop("fleet")) > 0 && !{var=observer}]}
	<tr>
		<td colspan="7" class="center">
			{if[!{var}can_send_fleet{/var}]}
				{lang}NO_FREE_FLEET_SLOTS{/lang}
				<br/>
			{/if}
			{if[!{var}can_send_expo{/var} && NS::getResearch(UNIT_EXPO_TECH) > 0]}
				{lang}NO_FREE_EXPO_SLOTS{/lang}
				<br/>
			{/if}
			{if[{var}can_send_expo{/var} || {var}can_send_fleet{/var}]}
				<input type="submit" name="step2" value="{lang}NEXT{/lang}" class="button" />
				{if[ {var=can_send_fleet} && !{var=stargateCountDown} && {var=stargate_level} ]}
					<input type="submit" name="stargatejump" value="{lang}STAR_GATE_JUMP{/lang}" class="button" />
				{/if}
			{/if}
		</td>
	</tr>
</form>
	{else if[!{var=observer}]}
	<tr>
		<td colspan="7" class="center">{lang}NO_SHIPS_REDEADY{/lang}</td>
	</tr>
	{/if}

	{if[$this->getLoop("defense")]}
<form method="post" action="{@formaction}">
<input type="hidden" name="galaxy" value="{request[get]}g{/request}" />
<input type="hidden" name="system" value="{request[get]}s{/request}" />
<input type="hidden" name="position" value="{request[get]}p{/request}" />
	<tr>
		<th colspan="7">Другие юниты для телепортации</th>
	</tr>
	<tbody>
    {foreach[defense]}
    <tr>
	    <td width="1">{loop}image{/loop}</td>
	    <td colspan="{if[isFacebookSkin() || isMobiSkin()]}2{else}3{/if}">{loop}name{/loop}</td>
	    <td class="center">{loop}quantity{/loop}</td>
	    <td class="center">
			{if[ {var=can_send_fleet} && !{var=stargateCountDown} && {var=stargate_level} ]}
				{if[!isMobiSkin()]}
					<a href="#" onclick="setField('ship_{loop}id{/loop}', {loop}quantity_raw{/loop}); resetRest('ship_{loop}id{/loop}'); return false">{lang}max{/lang}</a>
					<br /> <a href="#" onclick="setField('ship_{loop}id{/loop}', Math.ceil({loop}quantity_raw{/loop}/2)); resetRest('ship_{loop}id{/loop}'); return false">{lang}50%{/lang}</a>
					<br /> <a href="#" onclick="setField('ship_{loop}id{/loop}', 0); return false">{lang}min{/lang}</a>
				{else}
					<select onchange="setField('ship_{loop}id{/loop}', Math.ceil({loop}quantity_raw{/loop} * $(this).val() / 100))">
						<option value="100">{lang}max{/lang}</option>
						<option value="50">{lang}50%{/lang}</option>
						<option value="0">{lang}min{/lang}</option>
					</select>
				{/if}
			{/if}
		</td>
	    <td align="center">
			{if[ !$row["blocked"] && {var=can_send_fleet} && !{var=stargateCountDown} && {var=stargate_level} ]}
			<input class="center" type="text" name="{loop}id{/loop}" value="0" size="{@max_ships_size}" maxlength="{@max_ships_grade}" id="ship_{loop}id{/loop}" onchange="resetRest(this.id);"/>
			{/if}
		</td>
    </tr>
    {/foreach}
	</tbody>
	{if[ {var=can_send_fleet} && !{var=stargateCountDown} && {var=stargate_level} ]}
	<tr>
		<td colspan="7" class="center">
			<input type="submit" name="stargatejump_defense" value="{lang}STAR_GATE_JUMP{/lang}" class="button" />
		</td>
	</tr>
	{/if}
</form>
	{/if}

</table>

{if[count($this->getLoop("missions")) > 0]}
<table class="ntable">
	<thead>
		<tr>
			<th colspan="9">{lang}RUNNING_MISSIONS{/lang}</th>
		</tr>
		<tr>
			<td colspan="3">{lang}SERVER_TIME{/lang}</td>
			<td colspan="6"><span id="serverwatch">{@serverClock}</span></td>
		</tr>
		<tr>
		    <th>П</th>
		    <th>Э</th>
		    <th colspan="2">{lang}MISSION{/lang}</th>
			<th>{lang}START{/lang}</th>
			<th>{lang}ARRIVAL{/lang}</th>
			<th>{lang}TARGET{/lang}</th>
			<th>{lang}RETURN{/lang}</th>
			<th>{lang}ORDER{/lang}</th>
		</tr>
	</thead>
	<tbody>
   {foreach[missions]}
    <tr>
      <td{if[$row["fleet"]]} rowspan="2"{/if}>{if[$row["fleet_event"] == true]}{loop}fleet_slot{/loop}.{/if}</td>
      <td{if[$row["fleet"]]} rowspan="2"{/if}>{if[$row["exped_event"] == true]}{loop}exped_slot{/loop}.{/if}</td>
      <td colspan="2">
        <strong>{loop}mode{/loop}</strong>
        {if[$row["mode_o"]]}<br />{loop}mode_o{/loop}{/if}
      </td>
      <td>{loop}start{/loop}</td>
	  <td>{loop}arrival{/loop}</td>
	  <td>{loop}target{/loop}</td>
	  <td>{if[$row["return_exists"]]}{loop}return{/loop}{else}&nbsp;{/if}</td>
	  <td{if[$row["fleet"]]} rowspan="2"{/if} align="center">
		{if[ !isMobileSkin() && $row['event_pb_value'] ]}
		<div style="margin-bottom:5px"><div id="evpb{loop=eventid}" style="height:14px"></div></div>
		{/if}
        {if[$row["mode_r"] != EVENT_RETURN && $row["mode_r"] != EVENT_ROCKET_ATTACK && !$row["is_exchange"]]}
          <form method="post" action="{@formaction}">
		  {if[$row["mode_r"] == EVENT_HOLDING]}
			<input type="submit" name="control" value="{lang}CONTROL_EVENT{/lang}" class="button" />
		  {else}
			<input type="submit" name="retreat" value="{lang}RETREAT{/lang}" class="button" />
			{if[in_array($row["mode_r"], array(EVENT_ATTACK_SINGLE, EVENT_ATTACK_ALLIANCE, EVENT_ATTACK_DESTROY_BUILDING, EVENT_ATTACK_ALLIANCE_DESTROY_BUILDING, EVENT_ATTACK_DESTROY_MOON, EVENT_ATTACK_ALLIANCE_DESTROY_MOON))]}
				<br /><input type="submit" name="formation" value="{lang}FORMATION{/lang}" class="button" />
			{/if}
		  {/if}
          <input type="hidden" name="id" value="{loop}id{/loop}" />
          </form>
        {/if}
      </td>
	</tr>
    {if[$row["fleet"]]}
    <tr>
      <td colspan="6">
        <strong>{loop}quantity{/loop}</strong> ({loop}fleet{/loop})
      </td>
    </tr>
    {/if}
    {if[$row["num"] < $count]}
    <tr>
      <td colspan="9" style="height:5px"></td>
    </tr>
    {/if}
   {/foreach}
  </tbody>
</table>

{if[ !isMobileSkin() ]}
<script type="text/javascript">
$(function(){
	{foreach[missions]}
		{if[ $row['event_pb_value'] ]}
			$("#evpb{loop=eventid}").progressbar({
				value: {loop=event_pb_value}
			});
			setInterval(function(){
				var value = $("#evpb{loop=eventid}").progressbar("option", "value");
				if(value < 100)
				{
					$("#evpb{loop=eventid}").progressbar("option", "value", value+1);
				}
			}, {loop=event_percent_timeout});
		{/if}
	{/foreach}
});
</script>
{/if}

{/if}

{if[0]}
 Oxsar https://oxsar-nova.ru
 Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.

{/if}