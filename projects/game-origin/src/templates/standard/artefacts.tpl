{if[!{var=artefact_typeid}]}
<table class="ntable">
	<tr>
		<th colspan="4">
			<span style="float:right">{lang=LEVEL} {@artefact_tech_level}</span>
			{@artefact_tech_name}
		</th>
	</tr>
	<tr>
		<td colspan="4">
			<div style="float:left; padding-right:5px">{@artefact_tech_image}</div>
			<div style="display:table">{@artefact_tech_description}
				<div style="padding: 5px 0">
					<div class='rep_destroyed_back_div' style="clear:both"><div class='rep_alive_over_div' style='width: {@artefact_tech_free_percent}%' /></div>
				</div>
				<table class="table_no_background" cellspacing="0" cellpadding="0" border="0">
					<tr>
						<td>{lang=ARTEFACT_TECH_STORAGE}</td>
						<td>{@artefact_tech_storage}</td>
					</tr>
					<tr>
						<td>{lang=ARTEFACT_TECH_USED}</td>
						<td>{@artefact_tech_used}</td>
					</tr>
					<tr>
						<td{if[{var=artefact_tech_free} <= 0]} class='false'{/if}>{lang=ARTEFACT_TECH_FREE}</td>
						<td{if[{var=artefact_tech_free} <= 0]} class='false'{/if}>{@artefact_tech_free}</td>
					</tr>
				</table>
			</div>
		</td>
	</tr>
</table>
{/if}
<form method="post" action="{@formaction}">
<input type="hidden" id="artid" name="artid" value="0" />
<input type="hidden" id="typeid" name="typeid" value="0" />
<table class="ntable">
{if[count($this->getLoop("artefacts")) > 0]}
	<tr>
		{if[!{var=nonpersonal}]}
		<th colspan="2">&nbsp;</th>
		<th width=10%>{lang}QUANTITY{/lang}</th>
		{else}
		<th colspan="3">&nbsp;</th>
		{/if}
	</tr>
	{foreach[artefacts]}
	<tr>
		<td width="1" rowspan="2">{loop}image{/loop}</td>
		<td>
			<div style="width:100%">
				{if[$row['flags']]}
					<span style="float:right">{loop}flags{/loop}</span>
				{/if}
				{loop}name{/loop}
			</div>
			<div style="clear:both; font-size:smaller">
				{if[{var=artefact_typeid}]}
					{loop=description_full}
				{else}
					{loop=description}
				{/if}
			</div>
			<div>
				{if[1 || $row["can_build"]]}
					<table class="table_no_background" cellspacing="0" cellpadding="0" border="0">
						{include}"artefact_row_info"{/include}
					</table>
				{/if}
				
				{if[$row["required_constructions"]]}
					<span class="normal">{lang}REQUIRED_LIST_TITLE{/lang}</span>
					<br />{loop=required_constructions}
				{/if}
			</div>
		</td>
		{if[!{var=nonpersonal}]}
		<td class="center">
		
			{if[$row['quantity']]}
				{loop}quantity{/loop}{if[$row['active_count'] || $row['inactive_count']]} (<span class="active">{loop}active_count{/loop}</span>/<span class="inactive">{loop}inactive_count{/loop}</span>){/if}
				<br />
			{/if}
		</td>
		{else}
		{/if}
	</tr>
	<tr>
		{if[!{var=nonpersonal}]}
		<td colspan="2">
			<table class="table_no_background" cellspacing="0" cellpadding="0" border="0">
				<tr>
					<th>Расположен</th>
					<th>Состояние</th>
					<th>Время</th>
					{if[$row['fleet']]}
					<th>Эффект в полете</th>
					{else if[$row['battle']]}
					<th>Эффект в бою</th>
					{else}
					<th>Можно включить</th>
					{/if}
				</tr>
				{foreach2[.items]}
					<tr>
						<td>{if[$row['isInLot']]}на бирже{else}{loop=planetname}{/if}</td>
						<td>
							{if[$row['isInLot']]}существует
							{else if[$row['expire_counter']]}<span class='true'>активирован</span>
							{else if[$row['delay_counter']]}<span class='false'>заряжается</span>
							{else if[$row['disappear_counter']]}
								{if[$row['active']]}<span class='true'>активирован</span>
								{else}существует{/if}
							{else}существует{/if}
						</td>
						<td nowrap="nowrap" width="110">
							{if[$row['disappear_counter']]}
								{loop=disappear_counter}
							{else if[$row['expire_counter']]}
								{loop=expire_counter}
							{else if[$row['delay_counter']]}
								{loop=delay_counter}
							{else}{/if}
						</td>
						<td>{loop=times_left}
							{if[$row['times_left']>=2 && $row['times_left']<=4]}
								раза
							{else}
								раз
							{/if}
							{if[!$row['no_activation'] && $row['action']]}
								, {loop=action}
							{/if}
							{if[$row['packed'] && $row['check_req']]}
								({loop=cur_level} -> {loop=new_level})
							{/if}
						</td>
					</tr>
				{/foreach2}
			</table>
		</td>
		{/if}
	</tr>
	{/foreach}
	{else}
	<tr>
		<th colspan="4" class="center">{lang}NO_ARTEFACTS{/lang}</th>
	</tr>
	{/if}
</table>
</form>
{if[count($this->getLoop("uniques")) > 0]}
&nbsp;
<table class="ntable">
	<tr>
		<th colspan="2">{lang}UNIQUE_ARTEFACTS{/lang}</th>
		<th width="10%">{lang}OWNER{/lang}</th>
	</tr>
	{foreach[uniques]}
	<tr>
		<td width="1">{loop}image{/loop}</td>
		<td>
			<div style="width:100%">
				<span style="float:right">{loop}history{/loop}</span>
				{loop}name{/loop}
			</div>
			<div style="clear:both; font-size:smaller; margin:5px;">{loop}description{/loop}</div>
			<div><i>{lang}LOCATED{/lang} {loop}planetname{/loop}</i></div>
		</td>
		<td class="center">
			{loop}owner{/loop}
		</td>
	</tr>
	{/foreach}
</table>
{/if}
{if[0]}
 Oxsar https://oxsar-nova.ru
 Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
 
{/if}