<script type="text/javascript">
//<![CDATA[
var comis = {@comis};
var fleetMetal = {@fleet_metal};
var fleetSilicon = {@fleet_silicon};
var fleetHydrogen = {@fleet_hydrogen};

var outMetal = {@metal};
var outSilicon = {@silicon};
var outHydrogen = {@hydrogen};
var sCapacity = {@capacity};
var tMetal = outMetal;
var tSilicon = outSilicon;
var tHydrogen = outHydrogen;
var capacity = sCapacity;

var decPoint = '{lang}DECIMAL_POINT{/lang}';
var thousandsSep = '{lang}THOUSANDS_SEPERATOR{/lang}';

function setUnloadRes(id, value)
{
	value = parseSafeInt(value);
	$('#'+id).val(value);
	$('#'+id+'_comis').html(Math.ceil(value * comis / 100));
}

function unloadResChanged(id)
{
	var valStr = $('#'+id).val();
	var value = parseSafeInt(valStr);
	if(value < 0) value = 0;
	else if(id == 'fleetMetal' && value > fleetMetal) value = fleetMetal;
	else if(id == 'fleetSilicon' && value > fleetSilicon) value = fleetSilicon;
	else if(id == 'fleetHydrogen' && value > fleetHydrogen) value = fleetHydrogen;
	if(value != valStr)
	{
		$('#'+id).val(value);
	}
	$('#'+id+'_comis').html(Math.ceil(value * comis / 100));
}

function setMaxResComis(id, value)
{
	setMaxRes(id, value);
	$('#'+id+'_comis').html(Math.ceil(parseSafeInt($('#'+id).val()) * comis / 100));
}

function setMinResComis(id)
{
	setMinRes(id);
	$('#'+id+'_comis').val(0);
}

function updateAllComis()
{
	$('#metal_comis').html(Math.ceil(parseSafeInt($('#metal').val()) * comis / 100));
	$('#silicon_comis').html(Math.ceil(parseSafeInt($('#silicon').val()) * comis / 100));
	$('#hydrogen_comis').html(Math.ceil(parseSafeInt($('#hydrogen').val()) * comis / 100));
}

function renewTransportResComis()
{
	renewTransportRes();
	updateAllComis();
}

function setAllResourcesComis()
{
	setAllResources();
	updateAllComis();
}

function setNoResourcesComis()
{
	setNoResources();
	updateAllComis();
}

//]]>
</script>
<table class="ntable">
	<thead>
	<tr>
		<th colspan="5">Управление флотом</th>
	</tr>
	</thead>
	<tbody>
	<tr>
		<td colspan="4">
			{@fleet}
		</td>
		<td align="center">
			{if[ {var=back_possible} ]}
				<form method="post" action="{@formaction}">
					<input type="submit" name="retreat" value="{lang}RETREAT{/lang}" class="button" />
					<input type="hidden" name="id" value="{@eventid}" />
				</form>
			{else}
				<span class="false2">Не хватает {@back_consumption_needed} водорода</span>
			{/if}
		</td>
	</tr>
	</tbody>

	{if[ {var=holding_planet_selected} ]}
	<form method="post" action="{@holding_select_coords_action}">
	<input type="hidden" name="id" value="{@eventid}" />
	<thead>
		<tr>
			<th colspan="5">Новое задание</th>
		</tr>
	</thead>
	<tbody>
		<tr>
			{if[ !{var=remain_controls} ]}
				<td colspan="5" align="center">
					<span class="false2">Все задания израсходованы. Можно либо отозвать флот на предыдущий пункт дислокации, либо отправить флот на одну из планет его владельца.</span>
				</td>
			{else}
				<td colspan="4">
					Количество неиспользованных заданий.
				</td>
				<td align="center">
					<span class="true"><b>{@remain_controls}</b></span>
				</td>
			{/if}
		</tr>
		<tr>
			<td colspan="4">
				Чтобы переправить флот c {@holding_planet_select} на другие координаты, необходимо выбрать пункт назначения и задание. 
				При выполнении задания "Оставить" флот вернется своему владельцу, при "Удержании" 
				флот останется на новых координатах. В остальных случаях флот вернется на {@holding_planet_select}
				после выполнения задания. 
			</td>
			<td align="center"><input type="submit" name="holding_select_coords" value="{lang}NEXT{/lang}" class="button" /></td>
		</tr>
	</tbody>
	</form>

	{if[ {var=load_possible} ]}	
	<form method="post" action="{@formaction}">
	<input type="hidden" name="id" value="{@eventid}" />
	<thead>
		<tr>
			<th colspan="3">Загрузить ресурсы во флот</th>
			<th>Комиссия {@comis}%</th>
			<th>&nbsp;</th>
		</tr>
	</thead>
	<tbody>
		<tr>
			<td>{lang=METAL}</td>
			<td class="center"><a href="javascript:void();" onclick="javascript:setMaxResComis('metal', tMetal);">{lang}100%{/lang}</a>
			  | <a href="javascript:void();" onclick="javascript:setMaxResComis('metal', tMetal/2);">{lang}50%{/lang}</a>
			  | <a href="javascript:void();" onclick="javascript:setMinResComis('metal');">{lang}0%{/lang}</a></td>
			<td><input type="text" name="metal" size="12" maxlength="13" id="metal" onkeyup="javascript:renewTransportResComis();" /></td>
			<td><span id="metal_comis">0</span></td>
			<td rowspan="5" align="center"><input type="submit" name="load_resources" value="{lang}LOAD_RESOURCES{/lang}" class="button" /></td>
		</tr>
		<tr>
			<td>{lang=SILICON}</td>
			<td class="center"><a href="javascript:void();" onclick="javascript:setMaxResComis('silicon', tSilicon);">{lang}100%{/lang}</a>
			  | <a href="javascript:void();" onclick="javascript:setMaxResComis('silicon', tSilicon/2);">{lang}50%{/lang}</a>
			  | <a href="javascript:void();" onclick="javascript:setMinResComis('silicon');">{lang}0%{/lang}</a></td>
			<td><input type="text" name="silicon" size="12" maxlength="13" id="silicon" onkeyup="javascript:renewTransportResComis();" /></td>
			<td><span id="silicon_comis">0</span></td>
		</tr>
		<tr>
			<td>{lang=HYDROGEN}</td>
			<td class="center"><a href="javascript:void();" onclick="javascript:setMaxResComis('hydrogen', tHydrogen);">{lang}100%{/lang}</a>
			  | <a href="javascript:void();" onclick="javascript:setMaxResComis('hydrogen', tHydrogen/2);">{lang}50%{/lang}</a>
			  | <a href="javascript:void();" onclick="javascript:setMinResComis('hydrogen');">{lang}0%{/lang}</a></td>
			<td><input type="text" name="hydrogen" size="12" maxlength="13" id="hydrogen" onkeyup="javascript:renewTransportResComis();" /></td>
			<td><span id="hydrogen_comis">0</span></td>
		</tr>
		<tr>
			<td>{lang}CAPICITY{/lang}</td>
			<td colspan="3" class="center"><span id="rest" class="available">{@rest}</span></td>
		</tr>
		<tr>
			<td colspan="4" class="center"><a href="javascript:void();" onclick="javascript:setAllResourcesComis();">{lang}ALL_RESOURCES{/lang}</a> | <a href="javascript:void();" onclick="javascript:setNoResourcesComis();">{lang}NO_RESOURCES{/lang}</a></td>
		</tr>
	</tbody>
	</form>
	{/if}
	
	{/if}
	
	{if[ {var=fleet_metal} || {var=fleet_silicon} || {var=fleet_hydrogen} ]}

	<form method="post" action="{@formaction}">
	<input type="hidden" name="id" value="{@eventid}" />
	<thead>
		<tr>
			<th colspan="3">Выгрузить ресурсы на {@holding_planet_select}</th>
			<th>Комиссия {@comis}%</th>
			<th>&nbsp;</th>
		</tr>
	</thead>
	<tbody>
		<tr>
			<td>{lang=METAL}</td>
			<td class="center"><a href="javascript:void();" onclick="javascript:setUnloadRes('fleetMetal', fleetMetal);">{lang}100%{/lang}</a>
			  | <a href="javascript:void();" onclick="javascript:setUnloadRes('fleetMetal', fleetMetal/2);">{lang}50%{/lang}</a>
			  | <a href="javascript:void();" onclick="javascript:setUnloadRes('fleetMetal', 0);">{lang}0%{/lang}</a>
			</td>
			<td><input type="text" name="fleetMetal" size="8" maxlength="11" id="fleetMetal" onkeyup="javascript:unloadResChanged('fleetMetal');" /></td>
			<td><span id="fleetMetal_comis">0</span></td>
			<td rowspan="3" align="center"><input type="submit" name="unload_resources" value="{lang}UNLOAD_RESOURCES{/lang}" class="button" /></td>
		</tr>
		<tr>
			<td>{lang=SILICON}</td>
			<td class="center"><a href="javascript:void();" onclick="javascript:setUnloadRes('fleetSilicon', fleetSilicon);">{lang}100%{/lang}</a>
			  | <a href="javascript:void();" onclick="javascript:setUnloadRes('fleetSilicon', fleetSilicon/2);">{lang}50%{/lang}</a>
			  | <a href="javascript:void();" onclick="javascript:setUnloadRes('fleetSilicon', 0);">{lang}0%{/lang}</a>
			</td>
			<td><input type="text" name="fleetSilicon" size="8" maxlength="11" id="fleetSilicon" onkeyup="javascript:unloadResChanged('fleetSilicon');" /></td>
			<td><span id="fleetSilicon_comis">0</span></td>
		</tr>
		<tr>
			<td>{lang=HYDROGEN}</td>
			<td class="center"><a href="javascript:void();" onclick="javascript:setUnloadRes('fleetHydrogen', fleetHydrogen);">{lang}100%{/lang}</a>
			  | <a href="javascript:void();" onclick="javascript:setUnloadRes('fleetHydrogen', fleetHydrogen/2);">{lang}50%{/lang}</a>
			  | <a href="javascript:void();" onclick="javascript:setUnloadRes('fleetHydrogen', 0);">{lang}0%{/lang}</a>
			</td>
			<td><input type="text" name="fleetHydrogen" size="8" maxlength="11" id="fleetHydrogen" onkeyup="javascript:unloadResChanged('fleetHydrogen');" /></td>
			<td><span id="fleetHydrogen_comis">0</span></td>
		</tr>
	</tbody>
	</form>
	
	{/if}
	
	{if[ !{var=merchant_mark_used} && ({var=load_possible} || {var=fleet_metal} || {var=fleet_silicon} || {var=fleet_hydrogen}) ]}
	<tr>
		<td colspan="8" align="center"><span class="false2">Комиссия {const=EXCH_NO_MERCHANT_COMMISSION}%, 
			активируйте артефакт Знак торговца, чтобы уменьшить комиссию до {const=EXCH_MERCHANT_COMMISSION}%</span></td>
    </tr>
	{/if}
	
	{if[ !{var=holding_planet_selected} ]}
	<tbody>
		<tr>
			<td colspan="5" align="center">
				{if[ {var=holding_planet_owner} ]}
				Командир! Чтобы управлять флотом, перейди на {@holding_planet_select}
				{else}
				Полностью управлять флотом может владелец {@holding_planet_select} {@holding_planet_username} {@holding_planet_message}
				{/if}
			</td>
		</tr>
	</tbody>
	{/if}
	
</table>
