<script type="text/javascript">

adv_tech_matrix = [
	[{@cfg_tech_1_1}, {@cfg_tech_1_2}, {@cfg_tech_1_3}],
	[{@cfg_tech_2_1}, {@cfg_tech_2_2}, {@cfg_tech_2_3}],
	[{@cfg_tech_3_1}, {@cfg_tech_3_2}, {@cfg_tech_3_3}],
	];

adv_tech_scale = [ {@cfg_tech_scale_1}, {@cfg_tech_scale_2}, {@cfg_tech_scale_3} ];

$(function(){ advTechCalculate(); });

function advTechCalculate()
{
	advTechSideCalculate(0);
	advTechSideCalculate(1);

	for(var i = 0; i < 2; i++)
	{
		var a = i ? "d_" : "a_";
		var d = i ? "a_" : "d_";
		var shell_after = 0;
		var shots = $("#"+a+"base_shots").val();
		for(var j = 1; j <= 3; j++)
		{
			var shield = $("#"+d+"shield_"+j).val() - $("#"+a+"attack_"+j).val() * shots;
			if(shield < 0) shell_after += shield;
			$("#"+d+"shield_after_"+j).val( shield );
		}
		$("#"+d+"shell_after").val( shell_after );
	}
}

function advTechSideCalculate(side)
{
	var prefix = side ? "d_" : "a_";
	
	var base_list = ["base_attack", "base_shield"];
	for(var i = 0; i < 2; i++)
	{
	  var value = $("#"+prefix+base_list[i]).val();
	  var fixed_value = Math.abs(value.replace(/([^0-9])/g, ''));
	  if(fixed_value != value)
	  {
		  $("#"+prefix+base_list[i]).val( fixed_value );
	  }
	}	

	var adv_tech_def_attack = $("#"+prefix+"base_attack").val();
	var adv_tech_def_shield = $("#"+prefix+"base_shield").val();
	
	var tech_list = [];
	for(var i = 1; i <= 3; i++)
	{
	  var value = $("#"+prefix+"tech_"+i).val();
	  var fixed_value = Math.abs(value.replace(/([^0-9])/g, ''));
	  if(fixed_value != value)
	  {
		  $("#"+prefix+"tech_"+i).val( fixed_value );
	  }
	  tech_list[i-1] = Math.round(fixed_value * adv_tech_scale[i-1]);
	}
	if(1)
	{
		var tech_max = Math.max(tech_list[0], Math.max(tech_list[1], tech_list[2]));
		var tech_effect = [ 0, 0, 0 ];
		for(var i = 3; i >= 0; i--)
		{
			if(tech_list[i] == tech_max)
			{
				tech_effect[i] = tech_list[i];
				break;
			}
		}
	}
	else
	{
		var tech_min = Math.min(tech_list[0], Math.min(tech_list[1], tech_list[2]));
		var tech_effect = [
				tech_list[0] - tech_min,
				tech_list[1] - tech_min,
				tech_list[2] - tech_min,
			];
	}
	var tech_sum = tech_effect[0] + tech_effect[1] + tech_effect[2];
	if(tech_sum > 0)
	{
		tech_list[0] = tech_effect[0] / tech_sum + tech_list[0]/100;
		tech_list[1] = tech_effect[1] / tech_sum + tech_list[1]/100;
		tech_list[2] = tech_effect[2] / tech_sum + tech_list[2]/100;
	}
	else
	{
		tech_list[0] = 1.0 / 3 + tech_list[0]/100;
		tech_list[1] = 1.0 / 3 + tech_list[1]/100;
		tech_list[2] = 1.0 / 3 + tech_list[2]/100;
	}
	$("#"+prefix+"proc_1").val( Math.round(tech_list[0]*100)+"%" );
	$("#"+prefix+"proc_2").val( Math.round(tech_list[1]*100)+"%" );
	$("#"+prefix+"proc_3").val( Math.round(tech_list[2]*100)+"%" );

	var tech_attack = [ 
			Math.ceil(tech_list[0]*adv_tech_def_attack), 
			Math.ceil(tech_list[1]*adv_tech_def_attack), 
			Math.ceil(tech_list[2]*adv_tech_def_attack), 
		];

	$("#"+prefix+"attack_1").val( tech_attack[0] );
	$("#"+prefix+"attack_2").val( tech_attack[1] );
	$("#"+prefix+"attack_3").val( tech_attack[2] );
	
	var tech_shield = [ 0, 0, 0 ];
	for(var i = 0; i < 3; i++)
	{
		tech_shield[i] = (tech_list[0] * adv_tech_matrix[i][0]
			+ tech_list[1] * adv_tech_matrix[i][1]
			+ tech_list[2] * adv_tech_matrix[i][2]) / 1;
	}

	$("#"+prefix+"shield_proc_1").val( Math.round(tech_shield[0]*100)+"%" );
	$("#"+prefix+"shield_proc_2").val( Math.round(tech_shield[1]*100)+"%" );
	$("#"+prefix+"shield_proc_3").val( Math.round(tech_shield[2]*100)+"%" );
	
	var tech_shield = [ 
  			Math.ceil(tech_shield[0]*adv_tech_def_shield), 
  			Math.ceil(tech_shield[1]*adv_tech_def_shield), 
  			Math.ceil(tech_shield[2]*adv_tech_def_shield), 
  		];
	       	
	$("#"+prefix+"shield_1").val( tech_shield[0] );
	$("#"+prefix+"shield_2").val( tech_shield[1] );
	$("#"+prefix+"shield_3").val( tech_shield[2] );
}

</script>


<table class="ntable">
	<thead>
		<tr>
			<th colspan="2">Калькулятор специализации</th>
		</tr>
	</thead>
	<tfoot>
		<tr>
			<td colspan="2"><input class="button" type="button" onclick="advTechCalculate()"
				value="{lang}Пересчитать{/lang}">
			</td>
		</tr>
		<tr>
			<td width="50%">
				<table cellspacing="0" cellpadding="0" border="0" class="table_no_background">
					<tr>
						<td colspan="4">Влияние технологий</td>
					</tr>
					<tr>
						<td>&nbsp;</td>
						<td>ЛА</td>
						<td>ИО</td>
						<td>ПЛ</td>
					</tr>
					<tr>
						<td>Коеф.</td>
						<td>{@cfg_tech_scale_1}</td>
						<td>{@cfg_tech_scale_2}</td>
						<td>{@cfg_tech_scale_3}</td>
					</tr>
				</table>
			</td>
			<td width="50%">
				<table cellspacing="0" cellpadding="0" border="0" class="table_no_background">
					<tr>
						<td colspan="4">&nbsp;</td>
					</tr>
					<tr>
						<td colspan="4">Эффективность щитов</td>
					</tr>
					<tr>
						<td>&nbsp;</td>
						<td>ЛА&nbsp;<sub>щиты</sub></td>
						<td>ИО&nbsp;<sub>щиты</sub></td>
						<td>ПЛ&nbsp;<sub>щиты</sub></td>
					</tr>
					<tr>
						<td align="right">ЛА&nbsp;<sub>атака</sub></td>
						<td>{@cfg_tech_1_1_procent}</td>
						<td>{@cfg_tech_1_2_procent}</td>
						<td>{@cfg_tech_1_3_procent}</td>
					</tr>
					<tr>
						<td align="right">ИО&nbsp;<sub>атака</sub></td>
						<td>{@cfg_tech_2_1_procent}</td>
						<td>{@cfg_tech_2_2_procent}</td>
						<td>{@cfg_tech_2_3_procent}</td>
					</tr>
					<tr>
						<td align="right">ПЛ&nbsp;<sub>атака</sub></td>
						<td>{@cfg_tech_3_1_procent}</td>
						<td>{@cfg_tech_3_2_procent}</td>
						<td>{@cfg_tech_3_3_procent}</td>
					</tr>
				</table>
			</td>
		</tr>
		<tr>
			<td colspan="2">
				<table cellspacing="0" cellpadding="0" border="0" class="table_no_background">
					<tr>
						<td style="text-align:left">ЛА - лазерная технология</td>
					</tr>
					<tr>
						<td style="text-align:left">ИО - ионная технология</td>
					</tr>
					<tr>
						<td style="text-align:left">ПЛ - плазменная технология</td>
					</tr>
				</table>
			</td>
		</tr>
	</tfoot>
	<tr>
		<td valign="top">
			<table cellspacing="0" cellpadding="0" border="0" class="table_no_background">
				<tr>
					<td align="right">&nbsp;</td>
					<td colspan="3">Атакующий</td>
				</tr>
				<tr>
					<td align="right">Базовая атака</td>
					<td colspan="3"><input type="text" name="a_base_attack" value="{@cfg_tech_def_attack}" size="4" maxlength="5" id="a_base_attack" onchange="advTechCalculate()" onkeyup="advTechCalculate()" /></td>
				</tr>
				<tr>
					<td align="right">Базовый щит</td>
					<td colspan="3"><input type="text" name="a_base_shield" value="{@cfg_tech_def_shield}" size="4" maxlength="5" id="a_base_shield" onchange="advTechCalculate()" onkeyup="advTechCalculate()" /></td>
				</tr>
				<tr>
					<td align="right">Выстрелов</td>
					<td colspan="3"><input type="text" name="a_base_shield" value="1" size="2" maxlength="2" id="a_base_shots" onchange="advTechCalculate()" onkeyup="advTechCalculate()" /></td>
				</tr>
				<tr>
					<td align="right">&nbsp;</td>
					<td>ЛА</td>
					<td>ИО</td>
					<td>ПЛ</td>
				</tr>
				<tr>
					<td align="right">Технология</td>
					<td><input type="text" name="a_tech_1" value="{@a_tech_1}" size="2" maxlength="2" id="a_tech_1" onchange="advTechCalculate()" onkeyup="advTechCalculate()" /></td>
					<td><input type="text" name="a_tech_2" value="{@a_tech_2}" size="2" maxlength="2" id="a_tech_2" onchange="advTechCalculate()" onkeyup="advTechCalculate()" /></td>
					<td><input type="text" name="a_tech_3" value="{@a_tech_3}" size="2" maxlength="2" id="a_tech_3" onchange="advTechCalculate()" onkeyup="advTechCalculate()" /></td>
				</tr>
				<tr>
					<td align="right">Атака %</td>
					<td><input readonly="readonly" type="text" name="a_proc_1" value="" size="4" maxlength="4" id="a_proc_1" /></td>
					<td><input readonly="readonly" type="text" name="a_proc_2" value="" size="4" maxlength="4" id="a_proc_2" /></td>
					<td><input readonly="readonly" type="text" name="a_proc_3" value="" size="4" maxlength="4" id="a_proc_3" /></td>
				</tr>
				<tr>
					<td align="right">Атака</td>
					<td><input readonly="readonly" type="text" name="a_attack_1" value="" size="4" maxlength="4" id="a_attack_1" /></td>
					<td><input readonly="readonly" type="text" name="a_attack_2" value="" size="4" maxlength="4" id="a_attack_2" /></td>
					<td><input readonly="readonly" type="text" name="a_attack_3" value="" size="4" maxlength="4" id="a_attack_3" /></td>
				</tr>
				<tr>
					<td align="right">Щиты %</td>
					<td><input readonly="readonly" type="text" name="a_shield_proc_1" value="" size="4" maxlength="4" id="a_shield_proc_1" /></td>
					<td><input readonly="readonly" type="text" name="a_shield_proc_2" value="" size="4" maxlength="4" id="a_shield_proc_2" /></td>
					<td><input readonly="readonly" type="text" name="a_shield_proc_3" value="" size="4" maxlength="4" id="a_shield_proc_3" /></td>
				</tr>
				<tr>
					<td align="right">Щиты</td>
					<td><input readonly="readonly" type="text" name="a_shield_1" value="" size="4" maxlength="4" id="a_shield_1" /></td>
					<td><input readonly="readonly" type="text" name="a_shield_2" value="" size="4" maxlength="4" id="a_shield_2" /></td>
					<td><input readonly="readonly" type="text" name="a_shield_3" value="" size="4" maxlength="4" id="a_shield_3" /></td>
				</tr>
				<tr>
					<td align="right">Щитов после боя</td>
					<td><input readonly="readonly" type="text" name="a_shield_after_1" value="" size="4" maxlength="4" id="a_shield_after_1" /></td>
					<td><input readonly="readonly" type="text" name="a_shield_after_2" value="" size="4" maxlength="4" id="a_shield_after_2" /></td>
					<td><input readonly="readonly" type="text" name="a_shield_after_3" value="" size="4" maxlength="4" id="a_shield_after_3" /></td>
				</tr>
				<tr>
					<td align="right">Повреждение брони</td>
					<td colspan="3"><input readonly="readonly" type="text" name="a_shell_after" value="" size="4" maxlength="4" id="a_shell_after" /></td>
				</tr>
			</table>
		</td>
		<td valign="top">
			<table cellspacing="0" cellpadding="0" border="0" class="table_no_background">
				<tr>
					<td colspan="3">Обороняющийся</td>
				</tr>
				<tr>
					<td colspan="3"><input type="text" name="d_base_attack" value="{@cfg_tech_def_attack}" size="4" maxlength="5" id="d_base_attack" onchange="advTechCalculate()" onkeyup="advTechCalculate()" /></td>
				</tr>
				<tr>
					<td colspan="3"><input type="text" name="d_base_shield" value="{@cfg_tech_def_shield}" size="4" maxlength="5" id="d_base_shield" onchange="advTechCalculate()" onkeyup="advTechCalculate()" /></td>
				</tr>
				<tr>
					<td colspan="3"><input type="text" name="d_base_shield" value="1" size="2" maxlength="2" id="d_base_shots" onchange="advTechCalculate()" onkeyup="advTechCalculate()" /></td>
				</tr>
				<tr>
					<td>ЛА</td>
					<td>ИО</td>
					<td>ПЛ</td>
				</tr>
				<tr>
					<td><input type="text" name="d_tech_1" value="{@d_tech_1}" size="2" maxlength="2" id="d_tech_1" onchange="advTechCalculate()" onkeyup="advTechCalculate()" /></td>
					<td><input type="text" name="d_tech_2" value="{@d_tech_2}" size="2" maxlength="2" id="d_tech_2" onchange="advTechCalculate()" onkeyup="advTechCalculate()" /></td>
					<td><input type="text" name="d_tech_3" value="{@d_tech_3}" size="2" maxlength="2" id="d_tech_3" onchange="advTechCalculate()" onkeyup="advTechCalculate()" /></td>
				</tr>
				<tr>
					<td><input readonly="readonly" type="text" name="d_proc_1" value="" size="4" maxlength="4" id="d_proc_1" /></td>
					<td><input readonly="readonly" type="text" name="d_proc_2" value="" size="4" maxlength="4" id="d_proc_2" /></td>
					<td><input readonly="readonly" type="text" name="d_proc_3" value="" size="4" maxlength="4" id="d_proc_3" /></td>
				</tr>
				<tr>
					<td><input readonly="readonly" readonly="readonly" type="text" name="d_attack_1" value="" size="4" maxlength="4" id="d_attack_1" /></td>
					<td><input type="text" name="d_attack_2" value="" size="4" maxlength="4" id="d_attack_2" /></td>
					<td><input readonly="readonly" type="text" name="d_attack_3" value="" size="4" maxlength="4" id="d_attack_3" /></td>
				</tr>
				<tr>
					<td><input readonly="readonly" type="text" name="d_shield_proc_1" value="" size="4" maxlength="4" id="d_shield_proc_1" /></td>
					<td><input readonly="readonly" type="text" name="d_shield_proc_2" value="" size="4" maxlength="4" id="d_shield_proc_2" /></td>
					<td><input readonly="readonly" type="text" name="d_shield_proc_3" value="" size="4" maxlength="4" id="d_shield_proc_3" /></td>
				</tr>
				<tr>
					<td><input readonly="readonly" type="text" name="d_shield_1" value="" size="4" maxlength="4" id="d_shield_1" /></td>
					<td><input readonly="readonly" type="text" name="d_shield_2" value="" size="4" maxlength="4" id="d_shield_2" /></td>
					<td><input readonly="readonly" type="text" name="d_shield_3" value="" size="4" maxlength="4" id="d_shield_3" /></td>
				</tr>
				<tr>
					<td><input readonly="readonly" type="text" name="d_shield_after_1" value="" size="4" maxlength="4" id="d_shield_after_1" /></td>
					<td><input readonly="readonly" readonly="readonly" type="text" name="d_shield_after_2" value="" size="4" maxlength="4" id="d_shield_after_2" /></td>
					<td><input type="text" name="d_shield_after_3" value="" size="4" maxlength="4" id="d_shield_after_3" /></td>
				</tr>
				<tr>
					<td colspan="3"><input readonly="readonly" type="text" name="d_shell_after" value="" size="4" maxlength="4" id="d_shell_after" /></td>
				</tr>
			</table>
		</td>
	</tr>
</table>

{if[0]}
 Oxsar http://oxsar.ru
 Copyright (c) 2009-2010 UnitPoint <support@unitpoint.ru>
  
{/if}