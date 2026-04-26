<table class="ntable">
	<tr>
		<th>{@building_name}</th>
	</tr>
	<tr>
		<td>{@building_image}{@building_desc}{perm[CAN_EDIT_CONSTRUCTIONS]}<div class="right">{@edit}</div>{/perm}</td>
	</tr>
	{if[{var}chartType{/var} != "error"]}<tr>
		<td>{include}{var}chartType{/var}{/include}</td>
	</tr>{/if}
</table>
{if[{var}demolish{/var}]}<table class="ntable center">
	<tr>
		<th>{lang}DEMOLISH{/lang} {@building_name} {lang}LEVEL{/lang} {@building_level}</th>
	</tr>
	<tr>
		<td>{lang}REQUIRES{/lang} {@metal} {@silicon} {@hydrogen}<br />{lang}PRODUCTION_TIME{/lang} {@dimolishTime}</td>
	</tr>
	{if[{var}demolish_now{/var}]}<tr>
		<td>{@demolish_now}</td>
	</tr>{/if}
</table>{/if}
	{if[{var}pack_building{/var}]}<table class="ntable center">
	<tr>
		<td>{@pack_building} <br /> {@building_name} {@building_level}</td>
	</tr>
	</table>{/if}
	{if[{var}pack_research{/var}]}<table class="ntable center">
	<tr>
		<td>{@pack_research} <br /> {@building_name} {@building_level}</td>
	</tr>
	</table>{/if}
{if[0]}
 Oxsar http://oxsar.ru
 Copyright (c) 2009-2010 UnitPoint <support@unitpoint.ru>
  
{/if}