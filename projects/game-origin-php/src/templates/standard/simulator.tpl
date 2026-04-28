<script type="text/javascript">

sim_units = new Array(
    {foreach[fleet]}
    {
      id:{loop=id},
      unit_type:{loop=unit_type},
      quantity:{loop=quantity_raw},
      damaged:{loop=damaged},
      shell_percent:{loop=shell_percent}
    },
    {/foreach}
    null
  );
sim_units.pop();
// alert("started" + 3);
// alert("len: "+sim_units.length);

$(document).ready(function() {
	$('input[type=text]').change(function() {
		if( $(this).val() > {@max_ships} )
		{
			$(this).val({@max_ships});
		}
	});
});

function setUnitQuantity(base, num, dmg, per)
{
  num = parseInt(num);
  dmg = parseInt(dmg);
  per = parseInt(per);
  if(num > {@max_ships}){ num = {@max_ships}; }
  else if(num < 0){ num = $('#'+base).val(); }
  if(dmg > num || dmg < 0){ dmg = num; }
	setField(base, num);
	setField(base+'_d', dmg);
	setField(base+'_p', dmg > 0 ? per : 100);
}

function disableSim()
{
	// alert("disableSim");
	$("#sim").css("display", "none");
	$("#sim_fake").css("display", "inline");
	setTimeout(enableSim, 60000);
}
function enableSim()
{
	// alert("enableSim");
	$("#sim").css("display", "inline");
	$("#sim_fake").css("display", "none");
}

</script>

{if[{var=assaultid}>0]}

{if[defined('SN') && !defined('SN_FULLSCREEN')]}
	<script type="text/javascript">
	$(function(){
		var d_height	= $('body').height() - 20;
		var d_width		= $('body').width() - 20;
		d_height = {const=MAX_HEIGHT};
		$('#sim-assault-report-dialog').dialog({
			autoOpen:		false,
			position:		'center',
			modal: 			true,
			closeOnEscape: 	true,
			draggable:		false,
			resizable:		false,
			height:			d_height,
			width:			d_width
		});
		$('#sim_report').click(function(){
			$('#sim-assault-report-dialog').dialog('open');
			var data = '<iframe src ="'
				+ $(this).attr('href')
				+ '" width="'
				+ (d_width - 50)
				+ '" height="'
				+ (d_height - 60)
				+ '" id="assault_IFrame"></iframe>';
			$('#sim-assault-report-dialog').html(data);
			$('#assault_IFrame').load(function()
			{
				$(this).contents().find('a').attr('target', '_top');
			});
			return false;
		});
	});
	function openWindow(url)
	{
		win = window.open(url, "{lang}ASSAULT_REPORT{/lang}", "width=600,height=400,status=yes,scrollbars=yes,resizable=yes");
		win.focus();
	}
	</script>
	<div id="sim-assault-report-dialog" title="{lang}ASSAULT_REPORT_TITLE{/lang}">
	</div>
{/if}
<table class="ntable">
	<thead>
		<tr>
		  <th>Результаты</th>
		</tr>
	</thead>
	<tbody>
		<tr>
			<td>
			  <table class="table_no_background" cellspacing="0" cellpadding="0" border="0">
			    <tr>
			      <td><b>Победа атакующего:</b> {@attacker_win_percent}%</td>
			    </tr>
				  <tr>
			      <td><b>Победа обороняющегося:</b> {@defender_win_percent}%</td>
			    </tr>
				  <tr>
			      <td><b>Ничья:</b> {@draw_percent}%</td>
			    </tr>
				  <tr>
			      <td><b>Раундов:</b> {@turns}</td>
			    </tr>
			    <tr>
			      <td><b>Шанс появления луны:</b> {@moonchance}%</td>
			    </tr>
				  <tr>
			      <td><b>Потери атаки:</b> {@attacker_lost_res} ({@attacker_lost_metal} мет, {@attacker_lost_silicon} крем, {@attacker_lost_hydrogen} вод)</td>
			    </tr>
				  <tr>
			      <td><b>Потери обороны:</b> {@defender_lost_res} ({@defender_lost_metal} мет, {@defender_lost_silicon} крем, {@defender_lost_hydrogen} вод)</td>
			    </tr>
			    <tr>
			      <td><b>Обломков на орбите:</b> {@debris_metal} металла и {@debris_silicon} кремния</td>
			    </tr>
			    {if[{var=recyclers} > 0]}
			    <tr>
			      <td><b>Нужно переработчиков:</b> {@recyclers}</td>
			    </tr>
			    {/if}
				  <tr>
			      <td><b>Опыт атакующего:</b> {@attacker_exp}</td>
			    </tr>
				  <tr>
			      <td><b>Опыт обороняющегося:</b> {@defender_exp}</td>
			    </tr>
				  <tr>
			      <td><b>Время:</b> {@gentime_all} с, одна симуляция: {@gentime} с</td>
			    </tr>
				{if[{var=report_link}]}
					<tr>
						<td>
							<a id='sim_report' href="{@report_link}" class="false2" target="_blank">Отчет о сражении</a>
						</td>
					</tr>
				{/if}
			  </table>
			</td>
		</tr>
	</tbody>
</table>
{/if}

<form method="post" id="sim_form" action="{@formaction}" onsubmit="disableSim()">
<table class="ntable">
	<thead>
	  <tr>
		  <th colspan="2">{lang}OPTIONS{/lang}</th>
	  </tr>
	</thead>
	<tbody>
	  <tr>
		  <td>
			  <table class="table_no_background" cellspacing="0" cellpadding="0" border="0">
			    <tr>
			      <td>&nbsp;</td>
			      <td>{lang}GUN_POWER{/lang}</td>
			      <td>{lang}SHIELD_POWER{/lang}</td>
			      <td>{lang}ARMORING{/lang}</td>
			      <td>{lang}BALLISTICS_POWER{/lang}</td>
			      <td>{lang}MASKING_POWER{/lang}</td>
			      <td>{lang}SHIPYARD_POWER{/lang}</td>
			      <td>{lang}DEFENSE_FACTORY_POWER{/lang}</td>
			    </tr>
			    <tr>
			      <td>{lang}ATTACKER{/lang}</td>
				  <td><input type="text" name="a_tech_15" value="{@a_tech_15}" size="2" maxlength="2" id="a_tech_15" /></td>
				  <td><input type="text" name="a_tech_16" value="{@a_tech_16}" size="2" maxlength="2" id="a_tech_16" /></td>
				  <td><input type="text" name="a_tech_17" value="{@a_tech_17}" size="2" maxlength="2" id="a_tech_17" /></td>
				  <td><input type="text" name="a_tech_103" value="{@a_tech_103}" size="2" maxlength="2" id="a_tech_103" /></td>
				  <td><input type="text" name="a_tech_104" value="{@a_tech_104}" size="2" maxlength="2" id="a_tech_104" /></td>
				  <td><input type="text" name="a_tech_8" value="{@a_tech_8}" size="2" maxlength="2" id="a_tech_8" /></td>
				  <td>&nbsp;</td>
			    </tr>
			    <tr>
			      <td>{lang}DEFENDER{/lang}</td>
				  <td><input type="text" name="d_tech_15" value="{@d_tech_15}" size="2" maxlength="2" id="d_tech_15" /></td>
				  <td><input type="text" name="d_tech_16" value="{@d_tech_16}" size="2" maxlength="2" id="d_tech_16" /></td>
				  <td><input type="text" name="d_tech_17" value="{@d_tech_17}" size="2" maxlength="2" id="d_tech_17" /></td>
				  <td><input type="text" name="d_tech_103" value="{@d_tech_103}" size="2" maxlength="2" id="d_tech_103" /></td>
				  <td><input type="text" name="d_tech_104" value="{@d_tech_104}" size="2" maxlength="2" id="d_tech_104" /></td>
				  <td>&nbsp;</td>
				  <td><input type="text" name="d_tech_101" value="{@d_tech_101}" size="2" maxlength="2" id="d_tech_101" /></td>
			    </tr>
			  </table>
		  </td>
		  <td valign="top">
			  <table class="table_no_background" cellspacing="0" cellpadding="0" border="0">
			    <tr>
			      <td>Количество симуляций</td>
			    </tr>
			    <tr>
					  <td><input type="text" name="num_sim" value="{@num_sim}" size="2" maxlength="2" id="num_sim" /></td>
			    </tr>
			  </table>
		  </td>
	  </tr>
	  {if[0]}
	  <tr>
		<th colspan="2"  style="width: 100%">
			<input type="hidden" name="new_ass" value="0">
			<input type="checkbox" id="new_ass" name="new_ass" {if[{var=adv_assault}]}checked{/if} value="1">
			<label for="new_ass">{lang}USE_NEW_ASSAULT{/lang}</label>
		</th>
	  </tr>
	  <tr>
	  	<td colspan="2">
	  		<table class="table_no_background" cellspacing="0" cellpadding="0" border="0" style="width: 100%">
	  			<tr>
			      	<td>{if[0]}{lang}SIM_TECH_LIST{/lang}{/if}</td>
					<td nowrap>{lang}LASER_TECH{/lang}</td>
					<td nowrap>{lang}ION_TECH{/lang}</td>
					<td nowrap>{lang}PLASMA_TECH{/lang}</td>
					<td width="50%"></td>
		      	</tr>
			    <tr>
			      <td>{lang}ATTACKER{/lang}</td>
					  <td><input type="text" name="a_tech_23" value="{@a_tech_23}" size="2" maxlength="2" id="a_tech_23" /></td>
					  <td><input type="text" name="a_tech_24" value="{@a_tech_24}" size="2" maxlength="2" id="a_tech_24" /></td>
					  <td><input type="text" name="a_tech_25" value="{@a_tech_25}" size="2" maxlength="2" id="a_tech_25" /></td>
			    </tr>
			    <tr>
			      <td>{lang}DEFENDER{/lang}</td>
					  <td><input type="text" name="d_tech_23" value="{@d_tech_23}" size="2" maxlength="2" id="d_tech_23" /></td>
					  <td><input type="text" name="d_tech_24" value="{@d_tech_24}" size="2" maxlength="2" id="d_tech_24" /></td>
					  <td><input type="text" name="d_tech_25" value="{@d_tech_25}" size="2" maxlength="2" id="d_tech_25" /></td>
			    </tr>
	      </table>
	      </td>
	  </tr>
	  {/if}
	</tbody>
</table>

<table class="ntable">
	<thead>
		<tr>
			<th>{lang}SHIP_NAME{/lang}</th>
			<th>{lang}ATTACKER{/lang}</th>
			<th>{lang}DEFENDER{/lang}</th>
		</tr>
	</thead>
	<tfoot>
	  <tr>
		  <td colspan="3" class="center"><input type="submit" id="sim" name="simulate" value="{lang}SIMULATE{/lang}" class="button" /><span id="sim_fake" style="display:none;">Подождите несколько секунд</span></td>
	  </tr>
	</tfoot>
	<tbody>
		{foreach[fleet]}
		
		{if[$row["first_ship"]]}
		<tr>
			<th class="center"><a href="#" onclick="resetAllUnits({@fleet_unit_type}); return false">Обнулить флот</a></th>
			<th class="center">
			  <a href="#" onclick="setAllUnits('a_', {@fleet_unit_type}); return false">Установить флот</a>
			  <br />
			  <a href="#" onclick="setAllUnitsShellPercent('a_', {@fleet_unit_type}, 100); return false">100%</a>&nbsp;
			  <a href="#" onclick="setAllUnitsShellPercent('a_', {@fleet_unit_type}, 90); return false">90%</a>&nbsp;
			  <a href="#" onclick="setAllUnitsShellPercent('a_', {@fleet_unit_type}, 80); return false">80%</a>&nbsp;
			  <a href="#" onclick="setAllUnitsShellPercent('a_', {@fleet_unit_type}, 70); return false">70%</a>&nbsp;
			  <a href="#" onclick="setAllUnitsShellPercent('a_', {@fleet_unit_type}, 60); return false">60%</a>
			</th>
			<th class="center">
			  <a href="#" onclick="setAllUnits('d_', {@fleet_unit_type}); return false">Установить флот</a>
			  <br />
			  <a href="#" onclick="setAllUnitsShellPercent('d_', {@fleet_unit_type}, 100); return false">100%</a>&nbsp;
			  <a href="#" onclick="setAllUnitsShellPercent('d_', {@fleet_unit_type}, 90); return false">90%</a>&nbsp;
			  <a href="#" onclick="setAllUnitsShellPercent('d_', {@fleet_unit_type}, 80); return false">80%</a>&nbsp;
			  <a href="#" onclick="setAllUnitsShellPercent('d_', {@fleet_unit_type}, 70); return false">70%</a>&nbsp;
			  <a href="#" onclick="setAllUnitsShellPercent('d_', {@fleet_unit_type}, 60); return false">60%</a>
			</th>
		</tr>
		{/if}

		{if[$row["first_defense"]]}
		<tr>
			<th class="center"><a href="#" onclick="resetAllUnits({@def_unit_type}); return false">Обнулить оборону</a></th>
			<th class="center">&nbsp;</th>
			<th class="center">
			  <a href="#" onclick="setAllUnits('d_', {@def_unit_type}); return false">Установить оборону</a>
			  <br />
			  <a href="#" onclick="setAllUnitsShellPercent('d_', {@def_unit_type}, 100); return false">100%</a>&nbsp;
			  <a href="#" onclick="setAllUnitsShellPercent('d_', {@def_unit_type}, 90); return false">90%</a>&nbsp;
			  <a href="#" onclick="setAllUnitsShellPercent('d_', {@def_unit_type}, 80); return false">80%</a>&nbsp;
			  <a href="#" onclick="setAllUnitsShellPercent('d_', {@def_unit_type}, 70); return false">70%</a>&nbsp;
			  <a href="#" onclick="setAllUnitsShellPercent('d_', {@def_unit_type}, 60); return false">60%</a>
			</th>
		</tr>
		{/if}
		
		{if[$row["first_artefact"]]}
		<tr>
			<th class="center"><a href="#" onclick="resetAllUnits({@art_unit_type}); return false">Обнулить артефакты</a></th>
			<th class="center">
			  <a href="#" onclick="setAllUnits('a_', {@art_unit_type}); return false">Установить артефакты</a>
			</th>
			<th class="center">
			  <a href="#" onclick="setAllUnits('d_', {@art_unit_type}); return false">Установить артефакты</a>
			</th>
		</tr>
		{/if}
		
		<tr>
			<td>{loop=name}{if[$row["quantity"]]}, <nobr>{loop=quantity}</nobr>{/if}</td>
			{if[$row["unit_type"] != {var=art_unit_type}]}
			<td class="center" valign="top">
			  {if[$row["can_atter"]]}
				  <table class="table_no_background" cellspacing="0" cellpadding="0" border="0">
				    <tr>
				      <td><a href="#" class="klein" onclick="autoIncUnitQuantity('a_unit_{loop=id}', '{loop=quantity_raw}', '{loop=damaged}', '{loop=shell_percent}'); return false">+</a></td>
				      <td rowspan="2" nowrap="nowrap">
					      <input type="text" name="a_{loop=id}" value="{if[{var=assaultid}>0]}{loop=a_quantity_before}{else}0{/if}" size="{@max_ships_size}" maxlength="{@max_ships_grade}" id="a_unit_{loop=id}" />
					      [<input type="text" name="a_{loop=id}_d" value="{if[{var=assaultid}>0]}{loop=a_damaged_before}{else}0{/if}" size="{@max_ships_size}" maxlength="{@max_ships_grade}" id="a_unit_{loop=id}_d" />
					      -
					      <input type="text" name="a_{loop=id}_p" value="{if[{var=assaultid}>0]}{loop=a_percent_before}{else}100{/if}" size="2" maxlength="3" id="a_unit_{loop=id}_p" />%]
				      </td>
				    </tr>
				    <tr><td><a href="#" class="klein" onclick="autoDecUnitQuantity('a_unit_{loop=id}', '0', '0', '100'); return false">-</a></td></tr>
			      {if[{var=assaultid}>0 && $row["a_quantity_result"]]}
				    <tr>
				      <td>&nbsp;</td>
				      <td>{loop=a_quantity_result}</td>
				    </tr>
			      {/if}
				  </table>
			  {else}-{/if}
			  </td>
			  <td class="center" valign="top">
			  {if[$row["can_defense"]]}
				  <table class="table_no_background" cellspacing="0" cellpadding="0" border="0">
				    <tr>
				      <td><a href="#" class="klein" onclick="autoIncUnitQuantity('d_unit_{loop=id}', '{loop=quantity_raw}', '{loop=damaged}', '{loop=shell_percent}'); return false">+</a></td>
				      <td rowspan="2" nowrap="nowrap">
					      <input type="text" name="d_{loop=id}" value="{if[{var=assaultid}>0]}{loop=d_quantity_before}{else}0{/if}" size="{@max_ships_size}" maxlength="{@max_ships_grade}" id="d_unit_{loop=id}" />
					      [<input type="text" name="d_{loop=id}_d" value="{if[{var=assaultid}>0]}{loop=d_damaged_before}{else}0{/if}" size="{@max_ships_size}" maxlength="{@max_ships_grade}" id="d_unit_{loop=id}_d" />
					      -
					      <input type="text" name="d_{loop=id}_p" value="{if[{var=assaultid}>0]}{loop=d_percent_before}{else}100{/if}" size="2" maxlength="3" id="d_unit_{loop=id}_p" />%]
				      </td>
				    </tr>
				    <tr>
				      <td><a href="#" class="klein" onclick="autoDecUnitQuantity('d_unit_{loop=id}', '0', '0', '100'); return false">-</a></td>
				    </tr>
			      {if[{var=assaultid}>0 && $row["d_quantity_result"]]}
				    <tr>
				      <td>&nbsp;</td>
				      <td>{loop=d_quantity_result}</td>
				    </tr>
			      {/if}
				  </table>
			  {else}-{/if}
			</td>
			{else}
			  <td class="center" valign="top">
				  <table class="table_no_background" cellspacing="0" cellpadding="0" border="0" align="center">
				    <tr>
				      <td nowrap="nowrap">
					      <input type="text" name="a_{loop=id}" value="{if[{var=assaultid}>0]}{loop=a_quantity_before}{else}0{/if}" size="2" maxlength="2" id="a_unit_{loop=id}" />
				      </td>
				    </tr>
				    {if[{var=assaultid}>0 && $row["a_quantity_result"]]}
				    <tr>
				      <td class="center">{loop=a_quantity_result}</td>
				    </tr>
			      {/if}
				  </table>
			  </td>
			  <td class="center" valign="top">
				  <table class="table_no_background" cellspacing="0" cellpadding="0" border="0" align="center">
				    <tr>
				      <td nowrap="nowrap">
					      <input type="text" name="d_{loop=id}" value="{if[{var=assaultid}>0]}{loop=d_quantity_before}{else}0{/if}" size="2" maxlength="2" id="d_unit_{loop=id}" />
				      </td>
				    </tr>
			      {if[{var=assaultid}>0 && $row["d_quantity_result"]]}
				    <tr>
				      <td class="center">{loop=d_quantity_result}</td>
				    </tr>
			      {/if}
				  </table>
			  </td>
			{/if}
		</tr>
		{/foreach}

		<tr>
			<th colspan="2">Уничтожить</th>
			<th>Уровень</th>
		</tr>
		<tr>
			<td colspan="2">
			  <select id="buildingid" name="buildingid">
			    {foreach[constructions]}
			    <option value="{loop=id}"{if[$row['selected']]} selected="selected"{/if}>{loop=name}</option>
			    {/foreach}
			  </select>
			</td>
			<td class="center">
			  <input type="text" id="building_level" name="building_level" value="{@building_level}" size="2" maxlength="2" />
			</td>
		</tr>
		
	</tbody>
</table>
</form>
{if[0]}
 Oxsar https://oxsar-nova.ru
 Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
  
{/if}