{include}"statsheader"{/include}

<table class="ntable">
	<colgroup>
		<col width="1"/>
		<col width="*"/>
		<col width="1"/>
		<col width="*"/>
		<col width="*"/>
		<col width="*"/>
	</colgroup>
	<thead><tr>
		<th>#</th>
		<th>{lang}PLAYER{/lang}</th>
		<th></th>
		<th>{lang}ALLIANCE{/lang}</th>
		<th>{lang}POINTS{/lang}</th>
		<th>{lang}POSITION{/lang}</th>
	</tr></thead>
	<tfoot><tr>
		<td colspan="6">
			<p class="legend"><cite><span>i</span> = {lang}LOWER_INACTIVE{/lang}</cite><cite><span>I</span> = {lang}UPPER_INACTIVE{/lang}</cite><cite><span class="banned">b</span> = {lang}BANNED{/lang}</cite><cite><span class="vacation-mode">v</span> = {lang}VACATION_MODE{/lang}</cite></p>
			<p class="legend"><cite><span class="ownPosition">{lang=ONESELF}</span></cite><cite><span class="alliance">{lang=ALLIANCE}</span></cite><cite><span class="friend">{lang=FRIEND}</span></cite><cite><span class="enemy">{lang=ENEMY}</span></cite><cite><span class="confederation">{lang=CONFEDERATION}</span></cite><cite><span class="trade-union">{lang=TRADE_UNION}</span></cite><cite><span class="protection">{lang=PROTECTION_ALLIANCE}</span></cite></p>
		</td>
	</tr></tfoot>
	<tbody>{foreach[ranking]}<tr>
		<td>{loop}rank{/loop}</td>
		<td>{if[ $row["premium_seller"] ]}{loop=premium_seller}&nbsp;{/if}{loop}username_link{/loop}{if[$row["user_status_long"] != ""]} ({loop}user_status_long{/loop}){/if}</td>
		<td class="center" nowrap="nowrap">{if[$row["userid"] != Core::getUser()->get("userid")]}{perm[CAN_MODERATE_USER]}{loop}moderator{/loop} {/perm}{loop}message{/loop} {loop}buddyrequest{/loop} {loop}report{/loop}{/if}</td>
		<td>{loop}alliance{/loop}</td>
		<td>{loop}points{/loop}</td>
		<td>{loop}position{/loop}</td>
	</tr>{/foreach}</tbody>
</table>
{if[0]}
 Oxsar https://oxsar-nova.ru
 Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
  
{/if}