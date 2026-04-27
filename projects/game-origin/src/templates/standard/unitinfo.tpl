<table class="ntable">
	<tr>
		<th colspan="3">{@name}</th>
	</tr>
	<tr>
		<td colspan="3">
			{@pic}{@description}
			{if[count($this->getLoop("rapidfire")) > 0]}
			<br class="clear" /><br />
			<div>
				<strong>{lang}RAPIDFIRE{/lang}</strong><br />
				{foreach[rapidfire]}
				{loop}rapidfire{/loop}: {loop}value{/loop}<br />
				{/foreach}
			</div>{/if}
			{perm[CAN_EDIT_CONSTRUCTIONS]}<div class="right">{@edit}</div>{/perm}
		</td>
	</tr>
	<tr>
		<th>Боевые характеристики</th>
		{if[ {var=mode} == UNIT_TYPE_FLEET && {var=speed} ]}
		<th>В атаке</th>
		<th>В обороне</th>
		{else}
		<th colspan="2">В обороне</th>
		{/if}
	</tr>
	<tr>
		<td>{lang}ATTACK{/lang}</td>
		{if[ {var=mode} == UNIT_TYPE_FLEET && {var=speed} ]}
		<td>{@attacker_attack}</td>
		<td>{@attack}</td>
		{else}
		<td colspan="2">{@attack}</td>
		{/if}
	</tr>
	<tr>
		<td>{lang}SHIELD{/lang}</td>
		{if[ {var=mode} == UNIT_TYPE_FLEET && {var=speed} ]}
		<td>{@attacker_shield}</td>
		<td>{@shield}</td>
		{else}
		<td colspan="2">{@shield}</td>
		{/if}
	</tr>
	<tr>
		<td>{lang}SHELL{/lang}</td>
		{if[ {var=mode} == UNIT_TYPE_FLEET && {var=speed} ]}
		<td>{@shell}</td>
		<td>{@shell}</td>
		{else}
		<td colspan="2">{@shell}</td>
		{/if}
	</tr>
	<tr>
		<td>{lang}FRONT{/lang}</td>
		{if[ {var=mode} == UNIT_TYPE_FLEET && {var=speed} ]}
		<td>{@attacker_front} ( {@attacker_battle_weight} )</td>
		<td>{@front} ( {@battle_weight} )</td>
		{else}
		<td colspan="2">{@front} ( {@battle_weight} )</td>
		{/if}
	</tr>
	<tr>
		<td>{lang}BALLISTICS_POWER{/lang}</td>
		{if[ {var=mode} == UNIT_TYPE_FLEET && {var=speed} ]}
		<td>{@attacker_ballistics}</td>
		<td>{@ballistics}</td>
		{else}
		<td colspan="2">{@ballistics}</td>
		{/if}
	</tr>
	<tr>
		<td>{lang}MASKING_POWER{/lang}</td>
		{if[ {var=mode} == UNIT_TYPE_FLEET && {var=speed} ]}
		<td>{@attacker_masking}</td>
		<td>{@masking}</td>
		{else}
		<td colspan="2">{@masking}</td>
		{/if}
	</tr>
	
	<tr>
		<th colspan="3">Другие характеристики</th>
	</tr>
	{if[{var=mode} == UNIT_TYPE_FLEET]}
	<tr>
		<td>{lang}BASIC_SPEED{/lang}</td>
		<td colspan="2">
		  {foreach[engines]}
		    {loop=value} <br />
		  {/foreach}
		</td>
	</tr>
	<tr>
		<td>{lang}CAPACITY{/lang}</td>
		<td colspan="2">{@capacity}</td>
	</tr>
	<tr>
		<td>{lang}CONSUMPTION{/lang}</td>
		<td colspan="2">{@consume} ({@base_consume}) - {@gravi_name_link}</td>
	</tr>
	{/if}
	<tr>
		<td width="35%">{lang}STRUCTURE{/lang}</td>
		<td colspan="2">{@structure}</td>
	</tr>
	<tr>
		<td>{lang=COST}</td>
		<td colspan="2" nowrap="nowrap">{lang=METAL}: {@basic_metal}, {lang=SILICON}: {@basic_silicon}, {lang=HYDROGEN}: {@basic_hydrogen}</td>
	</tr>
	<tr>
		<td>{lang}REQUIRE_TIME_EXT{/lang}</td>
		<td colspan="2">{@productiontime}</td>
	</tr>
</table>
{if[0]}
 Oxsar https://oxsar-nova.ru
 Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
  
{/if}