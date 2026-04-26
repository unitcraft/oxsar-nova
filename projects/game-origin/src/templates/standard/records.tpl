<table class="ntable">
	<tr>
		<th colspan="2">{lang}CONSTRUCTION{/lang}</th>
		<th>{lang}PLAYER{/lang}</th>
		<th align="center">{lang}LEVEL{/lang}</th>
	</tr>
	{foreach[construction]}
	<tr>
	  {if[Core::getUser()->get("imagepackage") != "empty"]}
		<td width="1px">{loop}image{/loop}</td>
		<td>{loop}name{/loop}</td>
		{else}
		<td colspan="2">{loop}name{/loop}</td>
		{/if}
		<td>{loop}player{/loop}</td>
		<td align="center">{loop}record{/loop}</td>
	</tr>
	{/foreach}
	<tr>
		<th colspan="2">{lang}MENU_RESEARCH{/lang}</th>
		<th>{lang}PLAYER{/lang}</th>
		<th align="center">{lang}LEVEL{/lang}</th>
	</tr>
	{foreach[research]}
	<tr>
	  {if[Core::getUser()->get("imagepackage") != "empty"]}
		<td width="1px">{loop}image{/loop}</td>
		<td>{loop}name{/loop}</td>
		{else}
		<td colspan="2">{loop}name{/loop}</td>
		{/if}
		<td>{loop}player{/loop}</td>
		<td align="center">{loop}record{/loop}</td>
	</tr>
	{/foreach}
	<tr>
		<th colspan="2">{lang}SHIPYARD{/lang}</th>
		<th>{lang}PLAYER{/lang}</th>
		<th align="center">{lang}QUANTITY{/lang}</th>
	</tr>
	{foreach[shipyard]}
	<tr>
	  {if[Core::getUser()->get("imagepackage") != "empty"]}
		<td>{loop}image{/loop}</td>
		<td>{loop}name{/loop}</td>
		{else}
		<td colspan="2">{loop}name{/loop}</td>
		{/if}
		<td>{loop}player{/loop}</td>
		<td align="center">{loop}record{/loop}</td>
	</tr>
	{/foreach}
	<tr>
		<th colspan="2">{lang}DEFENSE{/lang}</th>
		<th>{lang}PLAYER{/lang}</th>
		<th align="center">{lang}QUANTITY{/lang}</th>
	</tr>
	{foreach[defense]}
	<tr>
	  {if[Core::getUser()->get("imagepackage") != "empty"]}
		<td>{loop}image{/loop}</td>
		<td>{loop}name{/loop}</td>
		{else}
		<td colspan="2">{loop}name{/loop}</td>
		{/if}
		<td>{loop}player{/loop}</td>
		<td align="center">{loop}record{/loop}</td>
	</tr>
	{/foreach}
	<tr>
		<th colspan="2">{lang}MOON_BUILDING{/lang}</th>
		<th>{lang}PLAYER{/lang}</th>
		<th align="center">{lang}LEVEL{/lang}</th>
	</tr>
	{foreach[moon]}
	<tr>
	  {if[Core::getUser()->get("imagepackage") != "empty"]}
		<td>{loop}image{/loop}</td>
		<td>{loop}name{/loop}</td>
		{else}
		<td colspan="2">{loop}name{/loop}</td>
		{/if}
		<td>{loop}player{/loop}</td>
		<td align="center">{loop}record{/loop}</td>
	</tr>
	{/foreach}
</table>
{if[0]}
 Oxsar http://oxsar.ru
 Copyright (c) 2009-2010 UnitPoint <support@unitpoint.ru>
  
{/if}