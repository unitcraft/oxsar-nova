<script type="text/javascript">
//<![CDATA[
var outMetal = {@metal};
var outSilicon = {@silicon};
var outHydrogen = {@hydrogen};
var sCapacity = {@capacity};
var tMetal = {@metal};
var tSilicon = {@silicon};
var tHydrogen = {@hydrogen};
var capacity = {@capacity};
var decPoint = '{lang}DECIMAL_POINT{/lang}';
var thousandsSep = '{lang}THOUSANDS_SEPERATOR{/lang}';

var balance_levels = new Array(
    {foreach[balanceAttackLevels]}
      {
        id: {loop=tech_id}
      },
    {/foreach}
    null
    );
balance_levels.pop();

function checkAttackLevelNumberInput(name, id)
{
  var field = $('#'+name);
  var cur_active_levels = {@exp_active_levels};

  var value = isNaN(field.val()) ? 0 : parseInt(field.val());
  var is_adv_tech = id == {const=UNIT_LASER_TECH} || id == {const=UNIT_ION_TECH} || id == {const=UNIT_PLASMA_TECH};
  var min_value = is_adv_tech ? -cur_active_levels : 0;
  if(value < min_value) value = min_value;
  else if(value > cur_active_levels) value = cur_active_levels;

  field.val(value);

  cur_active_levels -= Math.abs(value);

  // alert("id: "+id+", value: "+value);
  // str = "";

  for(i = 0; i < balance_levels.length; i++)
  {
    if(balance_levels[i].id != id)
    {
      value = $('#attack_lvl_'+balance_levels[i].id).val();
	  var min_value = is_adv_tech ? -cur_active_levels : 0;
      if(value < min_value) value = min_value;
      else if(value > cur_active_levels) value = cur_active_levels;
      setField("attack_lvl_"+balance_levels[i].id, value);

      cur_active_levels -= Math.abs(value);

      // str = str + " | "+balance_levels[i].id+"="+value;
    }
  }
  // alert(str + " ^ " + cur_active_levels);

  setField('exp_active_levels', cur_active_levels);
  setField('exp_points', {@exp_points} - ({@exp_active_levels} - cur_active_levels) * {@points_per_add_level});
}

$(function(){
	$('.balanceAttackLevels').css('display', 'none');
	$('.showHoldingTime').css('display', 'none');
	$('.expeditionParams').css('display', 'none');

	function update()
	{
		var showHoldingTime = false, balanceAttackLevels = false, expeditionParams = false;
		var mode = $(this).attr('id');
		if( mode == 'mode_15' )
		{
			expeditionParams = true;
		}
		else if( mode == 'mode_13' )
		{
			showHoldingTime = balanceAttackLevels = true;
		}
		else if ( mode == 'mode_10' || mode == 'mode_23' || mode == 'mode_25' )
		{
			balanceAttackLevels = true;
		}
		$('.expeditionParams').css('display', !expeditionParams ? 'none' : 'table-row');
		$('.showHoldingTime').css('display', !showHoldingTime ? 'none' : 'table-row');
		$('.balanceAttackLevels').css('display', !balanceAttackLevels ? 'none' : 'table-row');
	};

	$('input:radio[name=mode]').click(function(){
		update.call(this);
	});

	var checked = $('input:radio[name=mode]:checked');
	if(checked.attr('id')){
		update.call(checked);
	}
});

//]]>
</script>
<form method="post" action="{@formaction}">
<table class="ntable">
	<tr>
		<th colspan="4">{@targetName}</th>
	</tr>
	<tr>
		<th>
			{lang}MISSION{/lang}
		</th>

		{if[ !{var=holding_eventid} ]}
		<th colspan="3">{lang}RESOURCES{/lang}</th>
		{/if}
	</tr>
	<tr>
		<td {if[ !{var=holding_eventid} ]}rowspan="5"{else}colspan="4"{/if}>
			{if[ !{var=holding_eventid} ]}
				{if[!{var}can_send_fleet{/var}]}
					{lang}NO_FREE_FLEET_SLOTS{/lang}
					<br/>
				{/if}
				{if[!{var}can_send_expo{/var} && NS::getResearch(UNIT_EXPO_TECH) > 0]}
					{lang}NO_FREE_EXPO_SLOTS{/lang}
					<br/>
				{/if}
			{/if}
			{if[count($this->getLoop("missions")) > 0]}
				{foreach[missions]}
					<input type="radio" name="mode" id="mode_{loop}mode{/loop}" value="{loop}mode{/loop}" /> <label for="mode_{loop}mode{/loop}">{loop}mission{/loop}</label><br />
				{/foreach}
			{else}
				{lang}NO_MISSIONS_AVAILABLE{/lang}
			{/if}
		</td>

		{if[ !{var=holding_eventid} ]}
		<td>{lang}METAL{/lang}</td>
		<td class="center"><a href="javascript:void();" onclick="javascript:setMaxRes('metal', tMetal);">{lang}100%{/lang}</a>
		  | <a href="javascript:void();" onclick="javascript:setMaxRes('metal', tMetal/2);">{lang}50%{/lang}</a>
		  | <a href="javascript:void();" onclick="javascript:setMinRes('metal');">{lang}0%{/lang}</a></td>
		<td><input type="text" name="metal" size="12" maxlength="13" id="metal" onkeyup="javascript:renewTransportRes();" /></td>
		{/if}

	</tr>

	{if[ !{var=holding_eventid} ]}
	<tr>
		<td>{lang}SILICON{/lang}</td>
		<td class="center"><a href="javascript:void();" onclick="javascript:setMaxRes('silicon', tSilicon);">{lang}100%{/lang}</a>
		  | <a href="javascript:void();" onclick="javascript:setMaxRes('silicon', tSilicon/2);">{lang}50%{/lang}</a>
		  | <a href="javascript:void();" onclick="javascript:setMinRes('silicon');">{lang}0%{/lang}</a></td>
		<td><input type="text" name="silicon" size="12" maxlength="13" id="silicon" onkeyup="javascript:renewTransportRes();" /></td>
	</tr>
	<tr>
		<td>{lang}HYDROGEN{/lang}</td>
		<td class="center"><a href="javascript:void();" onclick="javascript:setMaxRes('hydrogen', tHydrogen);">{lang}100%{/lang}</a>
		  | <a href="javascript:void();" onclick="javascript:setMaxRes('hydrogen', tHydrogen/2);">{lang}50%{/lang}</a>
		  | <a href="javascript:void();" onclick="javascript:setMinRes('hydrogen');">{lang}0%{/lang}</a></td>
		<td><input type="text" name="hydrogen" size="12" maxlength="13" id="hydrogen" onkeyup="javascript:renewTransportRes();" /></td>
	</tr>
	<tr>
		<td>{lang}CAPICITY{/lang}</td>
		<td colspan="2" class="center"><span id="rest" class="available">{@rest}</span></td>
	</tr>
	<tr>
		<td colspan="3" class="center"><a href="javascript:void();" onclick="javascript:setAllResources();">{lang}ALL_RESOURCES{/lang}</a> | <a href="javascript:void();" onclick="javascript:setNoResources();">{lang}NO_RESOURCES{/lang}</a></td>
	</tr>

	<tr class="balanceAttackLevels">
		<th colspan="4">Распределение очков боевого опыта</th>
	</tr>
	{if[count($this->getLoop("balanceAttackLevels")) > 0]}
		<tr class="balanceAttackLevels">
			<td colspan="2">Накопленного опыта</td>
			<td colspan="2"><input type="text" disabled="disabled" name="exp_points" id="exp_points"
			  value="{@exp_points}" size="4" maxlength="4" /> {@points_per_add_level} очков = 1 уровень</td>
		</tr>
		<tr class="balanceAttackLevels">
			<td colspan="2">Доступно уровней</td>
			<td colspan="2"><input type="text" disabled="disabled" name="exp_active_levels" id="exp_active_levels"
			  value="{@exp_active_levels}" size="2" maxlength="2" /> максимум {@max_add_levels}</td>
		</tr>
		{foreach[balanceAttackLevels]}
			<tr class="balanceAttackLevels">
				<td colspan="2">{loop=tech_name}</td>
				<td colspan="2">
				<input type="text" name="attack_lvl_{loop=tech_id}" id="attack_lvl_{loop=tech_id}" value="0" size="2" maxlength="2"
				  onblur="checkAttackLevelNumberInput('attack_lvl_{loop=tech_id}', {loop=tech_id});" />
				{if[$row['select_options']]}<select onchange="setFromSelect('attack_lvl_{loop=tech_id}', this); checkAttackLevelNumberInput('attack_lvl_{loop=tech_id}', {loop=tech_id});">{loop=select_options}</select>{/if}
				</td>
			</tr>
		{/foreach}
	{else}
		<tr class="balanceAttackLevels">
			<td colspan="2">Накопленного опыта</td>
			<td colspan="2"><b>0</b>, {@points_per_add_level} очков = 1 уровень</td>
		</tr>
	{/if}

	{/if}

    {if[{var=planet_under_attack}]}
	<tr>
		<td colspan="4" class="center false2">
            На планете идет бой.
		</td>
	</tr>
	{/if}

    {if[{var=target_invalid_by_ally_attack}]}
	<tr>
		<td colspan="4" class="center false2">
            Вы не можете напасть на игрока, с которым атаковали вместе в течении последних {@target_check_ally_attack_days} дней.
		</td>
	</tr>
	{/if}

	{if[{var=max_bashing_count}]}
	<tr>
		<th colspan="4">Башинг</th>
	</tr>
	<tr>
		<td colspan="4" class="center">
            Одного игрока можно атаковать максимум {@max_bashing_count} раза в течении {@bashing_hours} часов.
            {if[{var=cur_bashing_count} >= {var=max_bashing_count}]}<span class='false2'>{/if}
                Вы атаковали этого игрока {@cur_bashing_count} раз(а).
            {if[{var=cur_bashing_count} >= {var=max_bashing_count}]}</span>{/if}
		</td>
	</tr>
	{/if}

	{if[{var=showHoldingTime}]}
	<tr class="showHoldingTime">
		<th colspan="4">{lang=HOLDING_TIME}</th>
	</tr>
	<tr class="showHoldingTime">
		<td colspan="4" class="center">
			<input type="text" name="holdingtime" id="holdingtime" value="1" size="2" maxlength="2" onblur="checkNumberInput(this, 0, 99);" class="center" /> <label for="holdingtime">{lang=HOURS}</label>
		</td>
	</tr>
	{/if}

    {if[{var=expeditionMode}]}
	<tr class="expeditionParams">
		<th colspan="4">{lang=EXPEDITION_TIME}</th>
	</tr>
	<tr class="expeditionParams">
		<td colspan="4" class="center">
			<select name="holdingtime" id="holdingtime" class="center" />
                        {@expTime}
                        </select><label for="holdingtime">{lang}HOURS{/lang}</label>
			<p />
			<span class="false2">{lang=EXPEDITION_WARNING}</span>
		</td>
	</tr>
	{/if}

	<tr>
		<td colspan="4" class="center">{if[count($this->getLoop("missions")) > 0]}<input type="submit" name="{if[ !{var=holding_eventid} ]}step4{else}holding_send_fleet{/if}" value="{lang}NEXT{/lang}" class="button" />{/if}</td>
	</tr>
</table>
</form>
{if[0]}
 Oxsar https://oxsar-nova.ru
 Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.

{/if}