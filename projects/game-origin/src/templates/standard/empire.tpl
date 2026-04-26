<table class="ntable">
  <tr>
    <th>№</th>
    <th>Планета</th>
    <th>Диаметр</th>
    <th>Поля</th>
    <th>Температура</th>
    <th>УМИ</th>
    <th colspan="2">Ресурсы</th>
  </tr>
  {foreach[planets]}
  <tr>
	<td>{loop=num}.</td>
	<td>{if[ $row["selected"] ]}* {/if}{loop=planetname} <br /> {loop=coords}</td>
	<td>{loop=diameter}</td>
	<td>{loop=fields} ({loop=maxFields}{if[ $row["ismoon"] && $row["maxFields"] < $row["maxFields2"] ]}-{loop=maxFields2}{/if})</td>
	<td>{loop=temperature}</td>
	<td>{loop=reseachVirtLab}</td>
	<td>М <br /> К <br /> В</td>
	<td align="right">{loop=metal}
		<br /> {loop=silicon}
		<br /> {loop=hydrogen}
	</td>
  </tr>
  {/foreach}
</table>

<table class="ntable">
	<tr>
		<th {if[(isFacebookSkin() || isMobiSkin()) && totalnum > MAX_FB_EMPIRE_PLANETS]}{else}colspan="2"{/if}>{lang}CONSTRUCTION{/lang}</th>
		{@num}
	</tr>
	{foreach[construction]}
	<tr>
		{if[(isFacebookSkin() || isMobiSkin()) && totalnum > MAX_FB_EMPIRE_PLANETS]}
			{if[Core::getUser()->get("imagepackage") != "empty"]}
				<td align='center'>{loop}image{/loop}</td>
			{/if}
		{else}
			{if[Core::getUser()->get("imagepackage") != "empty"]}
				<td>{loop}image{/loop}</td>
				<td width="100%">{loop}name{/loop}</td>
			{else}
				<td width="100%" colspan="2">{loop}name{/loop}</td>
			{/if}
		{/if}
		{loop}requirements{/loop}
	</tr>
	{/foreach}
	<tr>
		<th {if[(isFacebookSkin() || isMobiSkin()) && totalnum > MAX_FB_EMPIRE_PLANETS]}{else}colspan="2"{/if}>{lang}SHIPYARD{/lang}</th>
		{@num}
	</tr>
	{foreach[shipyard]}
	<tr>
		{if[(isFacebookSkin() || isMobiSkin()) && totalnum > MAX_FB_EMPIRE_PLANETS ]}
			{if[Core::getUser()->get("imagepackage") != "empty"]}
				<td align='center'>{loop}image{/loop}</td>
			{/if}
		{else}
			{if[Core::getUser()->get("imagepackage") != "empty"]}
				<td>{loop}image{/loop}</td>
				<td width="100%">{loop}name{/loop}</td>
			{else}
				<td width="100%" colspan="2">{loop}name{/loop}</td>
			{/if}
		{/if}
		{loop}requirements{/loop}
	</tr>
	{/foreach}
	<tr>
		<th {if[(isFacebookSkin() || isMobiSkin()) && totalnum > MAX_FB_EMPIRE_PLANETS]}{else}colspan="2"{/if}>{lang}DEFENSE{/lang}</th>
		{@num}
	</tr>
	{foreach[defense]}
	<tr>
		{if[(isFacebookSkin() || isMobiSkin()) && totalnum > MAX_FB_EMPIRE_PLANETS]}
			{if[Core::getUser()->get("imagepackage") != "empty"]}
				<td align='center'>{loop}image{/loop}</td>
			{/if}
		{else}
			{if[Core::getUser()->get("imagepackage") != "empty"]}
				<td>{loop}image{/loop}</td>
				<td width="100%">{loop}name{/loop}</td>
			{else}
				<td width="100%" colspan="2">{loop}name{/loop}</td>
			{/if}
		{/if}
		{loop}requirements{/loop}
	</tr>
	{/foreach}
	<tr>
		<th {if[(isFacebookSkin() || isMobiSkin()) && totalnum > MAX_FB_EMPIRE_PLANETS]}{else}colspan="2"{/if}>{lang}MOON_BUILDING{/lang}</th>
		{@num}
	</tr>
	{foreach[moon]}
	<tr>
		{if[(isFacebookSkin() || isMobiSkin()) && totalnum > MAX_FB_EMPIRE_PLANETS]}
			{if[Core::getUser()->get("imagepackage") != "empty"]}
				<td align='center'>{loop}image{/loop}</td>
			{/if}
		{else}
			{if[Core::getUser()->get("imagepackage") != "empty"]}
				<td>{loop}image{/loop}</td>
				<td width="100%">{loop}name{/loop}</td>
			{else}
				<td width="100%" colspan="2">{loop}name{/loop}</td>
			{/if}
		{/if}
		{loop}requirements{/loop}
	</tr>
	{/foreach}
	<tr>
		<th {if[(isFacebookSkin() || isMobiSkin()) && totalnum > MAX_FB_EMPIRE_PLANETS]}{else}colspan="2"{/if}>{lang}RESEARCH{/lang}</th>
		<th colspan="{@totalnum}">{lang}LEVEL{/lang}</th>
	</tr>
	{foreach[research]}
	<tr>
		{if[(isFacebookSkin() || isMobiSkin()) && totalnum > MAX_FB_EMPIRE_PLANETS]}
			{if[Core::getUser()->get("imagepackage") != "empty"]}
				<td align='center'>{loop}image{/loop}</td>
			{/if}
		{else}
			{if[Core::getUser()->get("imagepackage") != "empty"]}
				<td>{loop}image{/loop}</td>
				<td width="100%">{loop}name{/loop}</td>
			{else}
				<td width="100%" colspan="2">{loop}name{/loop}</td>
			{/if}
		{/if}
		<td colspan="{@totalnum}" {if[(isFacebookSkin() || isMobiSkin()) && totalnum > MAX_FB_EMPIRE_PLANETS]}align="left"{else}align="center"{/if}>{loop}requirements{/loop}</td>
	</tr>
	{/foreach}
</table>
{if[0]}
 Oxsar http://oxsar.ru
 Copyright (c) 2009-2010 UnitPoint <support@unitpoint.ru>
  
{/if}