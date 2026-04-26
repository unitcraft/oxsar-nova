{include}"statsheader"{/include}

<table class="ntable">
	<colgroup>
		<col width="30"/>
		<col width="200"/>
		<col width="60"/>
		<col width="105"/>
		<col width="105"/>
	</colgroup>
	<thead><tr>
		<th>#</th>
		<th>{lang}ALLIANCE{/lang}</th>
		<th>{lang}MEMBERS{/lang}</th>
		<th>{lang}POINTS{/lang}</th>
		<th>{lang}AVERAGE{/lang}</th>
	</tr></thead>
	<tfoot><tr>
		<td colspan="5">
			<p class="legend"><cite><span class="alliance">{lang=ALLIANCE}</span></cite><cite><span class="enemy">{lang=ENEMY}</span></cite><cite><span class="confederation">{lang=CONFEDERATION}</span></cite><cite><span class="trade-union">{lang=TRADE_UNION}</span></cite><cite><span class="protection">{lang=PROTECTION_ALLIANCE}</span></cite></p>
		</td>
	</tr></tfoot>
	<tbody>{foreach[ranking]}<tr>
		<td>{loop}rank{/loop}</td>
		<td>{loop}tag{/loop}</td>
		<td>{loop}members{/loop}</td>
		<td>{loop}totalpoints{/loop}</td>
		<td>{loop}average{/loop}</td>
	</tr>{/foreach}</tbody>
</table>
{if[0]}
 Oxsar http://oxsar.ru
 Copyright (c) 2009-2010 UnitPoint <support@unitpoint.ru>
  
{/if}