<table class="itable">
{foreach[planetSelection]}
	{if[$row["counter"] % 2 != 0]}<tr>{/if}
		<td>{loop}planetname{/loop}<br />{loop}planetimage{/loop}<br />{loop}activity{/loop}</td>
	{if[$row["total"] == $row["counter"] && $row["counter"] % 2 != 0]}<td></td>{/if}
	{if[$row["counter"] % 2 == 0 || $row["total"] == $row["counter"]]}</tr>{/if}
{/foreach}
</table>
{if[0]}
 Oxsar http://oxsar.ru
 Copyright (c) 2009-2010 UnitPoint <support@unitpoint.ru>
  
{/if}