<table class="ntable">
	<colgroup>
		<col width="4%" />
		<col width="20%" />
		<col width="20%" />
		<col width="20%" />
		<col width="20%" />
		<col width="16%" />
	</colgroup>
	
	{if[count($this->getLoop("featured_exchanges")) > 0]}
		<tr>
			<th colspan="7">{lang}EXCH_NEWLOT_FEATURED{/lang}</th>
		</tr>
		
		{if[0]}
		<tr class="center">
		  <th>{lang}SEQ_NUM{/lang}</th>
			{if[isFacebookSkin() || isMobiSkin()]}
			  <th>
				{lang}EXCH_TITLE{/lang}
				<br/>
				{lang}PLAYER{/lang}
				<br/>
				{lang}PLANET{/lang}
			  </th>
			  {else}
				  <th>{lang}EXCH_TITLE{/lang}</th>
				  <th>{lang}PLAYER{/lang}</th>
				  <th>{lang}PLANET{/lang}</th>
			  {/if}
		  <th>{lang}EXCH_FEE{/lang}<br/>({lang}EXCH_COMISSION{/lang})<sup>{@fee}</sup></th>
		  <th>&nbsp;</th>
		  <th>&nbsp;</th>
		</tr>
		{/if}
			
		{foreach[featured_exchanges]}
		<tr>
		  {if[isFacebookSkin() || isMobiSkin()]}
			<td{loop}class{/loop}>{loop}i{/loop}</td>
			<td{loop}class{/loop}>
				{loop}title{/loop} {loop=level}
				<br/>
				{loop}username{/loop}
				<br/>
				{loop}planetname{/loop}&nbsp;{loop}coords{/loop}
			</td>
			<td{loop}class{/loop}>{loop}fee{/loop}% ({loop}comission{/loop} {lang}CR{/lang})</td>
		  {else}
			<td nowrap="nowrap"{loop}class{/loop}>{loop}i{/loop}</td>
			<td nowrap="nowrap"{loop}class{/loop}>{loop}title{/loop} {loop=level}</td>
			<td nowrap="nowrap"{loop}class{/loop}>{loop}username{/loop}</td>
			<td{loop}class{/loop}>{loop}planetname{/loop}&nbsp;{loop}coords{/loop}</td>
			<td nowrap="nowrap"{loop}class{/loop}>{loop}fee{/loop}% ({loop}comission{/loop} {lang}CR{/lang})</td>
		  {/if}
			<td>{loop}status{/loop}</td>
			<td>{loop}action{/loop}</td>
		</tr>
		{/foreach}
	{/if}

	<tr>
		<th colspan="7">{lang}EXCH_NEWLOT_S1{/lang}</th>
	</tr>
	<tr>
	  <th style="text-align: center" colspan="7">
		  <form method="post" action="{@formaction}" style="display: none;">
			  <input type="hidden" name="page" id="page" value="1">
			  <input type="hidden" name="sort_field" id="sort_field" value="{@sort_field}">
			  <input type="hidden" name="sort_order" id="sort_order" value="{@sort_order}">
			  <input type="submit" name="go" id="go" value="{lang}COMMIT{/lang}" class="button" style="display: none;"/>
		  </form>
		<div style="float: right">
		  <select name="page1" id="page1" onchange="goPage(this.value)">{@pages}</select>
		</div>
		{@link_first}
		{@link_prev}
		{@page_links}
		{@link_next}
		{@link_last}
	  </th>
	</tr>
	<tr class="center">
	  <th>{lang}SEQ_NUM{/lang}</th>
		{if[isFacebookSkin() || isMobiSkin()]}
		  <th>
			<sup>{@title}</sup><a href="#" onclick="setOrder('title'); return false">{lang}EXCH_TITLE{/lang}</a>
			<br/>
			<sup>{@player}</sup><a href="#" onclick="setOrder('player'); return false">{lang}PLAYER{/lang}</a>
			<br/>
			<sup>{@planet}</sup><a href="#" onclick="setOrder('planet'); return false">{lang}PLANET{/lang}</a>
		  </th>
		  {else}
			  <th><a href="#" onclick="setOrder('title'); return false">{lang}EXCH_TITLE{/lang}</a><sup>{@title}</sup></th>
			  <th><a href="#" onclick="setOrder('player'); return false">{lang}PLAYER{/lang}</a><sup>{@player}</sup></th>
			  <th><a href="#" onclick="setOrder('planet'); return false">{lang}PLANET{/lang}</a><sup>{@planet}</sup></th>
		  {/if}
	  <th><a href="#" onclick="setOrder('fee'); return false">{lang}EXCH_FEE{/lang}<br/>({lang}EXCH_COMISSION{/lang})</a><sup>{@fee}</sup></th>
	  <th>&nbsp;</th>
	  <th>&nbsp;</th>
	</tr>
	
	<tfoot>
		<tr>
		  <th style="text-align: center" colspan="7">
			<div style="float: right">
			  <select name="page2" id="page2" onchange="goPage(this.value)">{@pages}</select>
			</div>
			{@link_first}
			{@link_prev}
			{@page_links}
			{@link_next}
			{@link_last}
		  </th>
		</tr>
	</tfoot>
	
	{foreach[exchanges]}
	<tr>
	  {if[isFacebookSkin() || isMobiSkin()]}
		<td{loop}class{/loop}>{loop}i{/loop}</td>
		<td{loop}class{/loop}>
			{loop}title{/loop} {loop=level}
			<br/>
			{loop}username{/loop}
			<br/>
			{loop}planetname{/loop}&nbsp;{loop}coords{/loop}
		</td>
		<td{loop}class{/loop}>{loop}fee{/loop}% ({loop}comission{/loop} {lang}CR{/lang})</td>
	  {else}
		<td nowrap="nowrap"{loop}class{/loop}>{loop}i{/loop}</td>
		<td nowrap="nowrap"{loop}class{/loop}>{loop}title{/loop} {loop=level}</td>
		<td nowrap="nowrap"{loop}class{/loop}>{loop}username{/loop}</td>
		<td{loop}class{/loop}>{loop}planetname{/loop}&nbsp;{loop}coords{/loop}</td>
		<td nowrap="nowrap"{loop}class{/loop}>{loop}fee{/loop}% ({loop}comission{/loop} {lang}CR{/lang})</td>
	  {/if}
		<td>{loop}status{/loop}</td>
		<td>{loop}action{/loop}</td>
	</tr>
	{/foreach}
</table>
{if[0]}
 Oxsar https://oxsar-nova.ru
 Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
  
{/if}