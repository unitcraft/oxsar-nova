<table class="ntable">
	{if[count($this->getLoop("achievement")) > 0]}
	{foreach[achievement]}
	<tr>
		{if[$row['image']]}
			<td rowspan="2" width="1px">
				{loop}image{/loop}
			</td>
		{/if}
		<td>
		<div style="width:100%">
				{loop}name{/loop} {loop}progress{/loop}%
				<br/>
			</div>
			<div style="clear:both; font-size:smaller">
				{loop}desc{/loop}
			</div>
			<div>
				{loop}reqs{/loop}
			</div>
		</td>
	</tr>
	<tr>
		<td>
			{loop}bonuses{/loop}
		</td>
	</tr>
	{/foreach}
	{/if}
</table>
{if[0]}
 Oxsar http://oxsar.ru
 Copyright (c) 2009-2010 UnitPoint <support@unitpoint.ru>
 
{/if}