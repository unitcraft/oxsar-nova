<table class="ntable center">
<thead><tr>
	<th>{lang}LEVEL{/lang}</th>
	<th>{lang}CONSUME{/lang}</th>
	<th>{lang}DIFFERENCE{/lang}</th>
</tr></thead>
<tbody>{foreach[chart]}<tr>
	<td><span class="{if[$row["level"] == {var}buildingLevel{/var}]}true{/if}">{loop}level{/loop}</span></td>
	<td>{loop}cons{/loop}</td>
	<td><span class="{if[$row["s_diffCons"] > 0]}true{else if[$row["s_diffCons"] < 0]}false{/if}">{loop}diffCons{/loop}</span></td>
</tr>{/foreach}</tbody>
</table>
{if[0]}
 Oxsar http://oxsar.ru
 Copyright (c) 2009-2010 UnitPoint <support@unitpoint.ru>
  
{/if}