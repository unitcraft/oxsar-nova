{include}"searchheader"{/include}
<table class="ntable">
	<tr>
		<th>{lang}TAG{/lang}</th>
		<th>{lang}ALLIANCE{/lang}</th>
		<th>{lang}MEMBERS{/lang}</th>
		<th>{lang}POINTS{/lang}</th>
	</tr>
	{foreach[result]}<tr>
		<td>{loop}tag{/loop}</td>
		<td>{loop}name{/loop}</td>
		<td>{loop}members{/loop}</td>
		<td>{loop}points{/loop}</td>
	</tr>{/foreach}
	{if[count($this->getLoop("result")) == 0]}<tr>
		<td colspan="4" class="center">{lang}NO_MATCHES_FOUND{/lang}</td>
	</tr>{/if}
</table>
{if[0]}
 Oxsar https://oxsar-nova.ru
 Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
  
{/if}