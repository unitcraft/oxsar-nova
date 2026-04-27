{if[{var=show_achives}]}
{if[!{var=is_achiev_other_user}]}
<script type="text/javascript">
function hide_button_clicked(obj, list, count, prefix)
{
	var id = parseInt($(obj).attr('name'));
	for(var i = 0; i < list.length; i++)
	{
		if(list[i] == id)
		{
			return;
		}
	}
	list.push(id);
	
	{if[!{var=is_achiev_page}]}
		if(list.length < count)
		{
			$('tbody.'+prefix+'_achiev_' + id).remove();
		}
		else
		{
			$('table.'+prefix+'_achiev_table').hide('blind', '', 250);
		}
	{else}
		$('.'+prefix+'_achiev_btn_' + id).remove();
	{/if}
	$.ajax({
		url: '{@achiev_ajax_hide_url}' + id,
		success: function(data)
		{
			;
		}
	});
}

</script>
{/if}
{if[{var=done_achi_paginator} || {var=aval_achi_paginator}]}
<script type="text/javascript">
//<![CDATA[
function aval_goPage(aval_page){
	aval_page_select = document.getElementById('aval_page');
	aval_page_select.value = aval_page;
    document.getElementById('aval_go').click();
}
function done_goPage(done_page){
	done_page_select = document.getElementById('done_page');
	done_page_select.value = done_page;
    document.getElementById('done_go').click();
}
//]]>
</script>
{/if}
{if[{var=aval_achi} && count($this->getLoop("new_achieves")) > 0]}
{if[!{var=is_achiev_other_user}]}
<script type="text/javascript">
var new_achiev_hidden = [];
$(document).ready(function() {
	$('.new_achiev_hide_button').live('click', function() {
		hide_button_clicked(this, new_achiev_hidden, <?= count($this->getLoop("new_achieves")) ?>, 'new');
	});
});
</script>
{/if}

{if[ NS::getUser()->get("points") < 1000 && !AchievementsService::isGrantedAchievement(Yii::app()->user->id, ACHIEVEMENT_NEWBIE_END) && Core::getRequest()->getGET("go") == "Achievements" ]}
<div class="content_width">
	<div class="ui-state-error ui-corner-all" style="padding: 5px 20px; margin: 10px auto;">
		<table>
			<tr>
				<td>
					<span class="ui-icon ui-icon-info" style="margin-right: .3em; max-width: 20px; width: 20px;"></span>
				</td>
				<td>
					<?php 
						$messages = array(
							"Чтобы выполнить достижение, необходимо выполнить все его требования. Невыполненные требования указыны красным цветом.",
							);
						echo $messages[ array_rand($messages) ];
					?>				
				</td>
			</tr>
		</table>
	</div>
</div>
{/if}

<table class="ntable new_achiev_table">
	<thead>
		{if[{var=aval_achi_paginator}]}
		<tr>
            <th style="text-align: center" colspan="3">
                <form name="form_pages" id="form_pages" method="post" action="{@aval_formaction}">
                <div style="float: right;">
                  <select name="page" id="aval_page" onchange="document.getElementById('aval_go').click();">{@aval_pages}</select>
                </div>
                {@aval_link_first}
                {@aval_link_prev}
                {@aval_page_links}
                {@aval_link_next}
                {@aval_link_last}
                <input type="submit" style="display: none;" name="go" id="aval_go" value="{lang}COMMIT{/lang}" class="button" />
                </form>
            </th>
		</tr>
		{/if}
		<tr>
			<th colspan=3>
				{if[{var=achievement_id}]}
					{lang=ACHIEVEMENT}
				{else}
					{lang=AVAIL_ACHIEVEMENTS}{if[{var=real_new_achieve_count} > 0]} <div style="float:right">{lang=NEW_ACHIEVEMENTS_PREFIX} <span class="true"><b>+{@real_new_achieve_count}</b></span></div>{/if}
				{/if}
			</th>
		</tr>
	</thead>
	
	{if[!{var=is_achiev_other_user}]}
	<tfoot>
		<tr>
			<td colspan=3>
				<form method="post" action="{@achiev_recalc_url}">
					<input class="button" type="submit" value="{lang=RECALC_ACHIEVEMENTS}">
				</form>
			</td>
		</tr>
	</tfoot>
	{/if}
	
	{foreach[new_achieves]}
	<tbody class='new_achiev_{loop=achievement_id}'>
	<tr>
		<td rowspan="{if[$row['reqs_exist']]}5{else}3{/if}" width="1px">
			{loop}image{/loop}
		</td>
		<td style="vertical-align: top;">
			<div style="width:100%">
				{loop}name{/loop}
			</div>
			<div style="clear:both; font-size:smaller">
				{loop}desc{/loop}
			</div>
		</td>
		<td rowspan="{if[$row['reqs_exist']]}5{else}3{/if}" width="100px" align="center">
			{if[!{var=is_achiev_other_user}]}
				{if[!{var=is_achiev_page}]}
					<input type='button' class="new_achiev_hide_button button" name="{loop=achievement_id}" value="Скрыть">
				{else if[$row['state'] != ACHIEV_STATE_HIDDEN]}
					<input type='button' class="new_achiev_hide_button button new_achiev_btn_{loop=achievement_id}" name="{loop=achievement_id}" value="ОК">
				{/if}
			{/if}
		</td>
	</tr>
	<tr>
		<td>
			<b>{lang=ACHIEVEMENT_BONUS}</b>
		</td>
	</tr>
	<tr>
		<td>
			<table class="table_no_background" cellspacing="0" cellpadding="0" border="0" style="clear:both">
				{foreach2[.bonus_items]}
					<tr>
						<td>{loop=res_name}</td>
						<td class='true'>{loop=res_bonus}</td>
					</tr>
				{/foreach2}
			</table>
			<div style="clear:both">
				{loop}bonuses{/loop}
			</div>
		</td>
	</tr>
	
	{if[$row['reqs_exist']]}
	<tr class="for_pointer_req">
		<td>
			<b>{lang=REQUIRES}</b>
		</td>
	</tr>
	<tr>
		<td>
			<table class="table_no_background" cellspacing="0" cellpadding="0" border="0" style="clear:both">
				{foreach2[.req_items]}
					{if[$row["res_required"]]}
					<tr>
						<td>{loop=res_name}</td>
						<td class='{if[$row["res_notavailable"]]}notavailable{else}true{/if}'>{loop=res_required}</td>
						<td>{if[$row["res_notavailable"]]}({loop=res_notavailable}){/if}</td>
					</tr>
					{/if}
				{/foreach2}
			</table>
			<div style="clear:both">
				{loop}reqs{/loop}
			</div>
		</td>
	</tr>
	{/if}
	</tbody>
	{/foreach}
</table>
{/if}

{if[{var=done_achi} && count($this->getLoop("cur_achieves")) > 0]}
{if[!{var=is_achiev_other_user}]}
<script type="text/javascript">
var cur_achiev_hidden = [];
$(document).ready(function() {
	$('.cur_achiev_hide_button').live('click', function() {
		hide_button_clicked(this, cur_achiev_hidden, <?= count($this->getLoop("cur_achieves")) ?>, 'cur');
	});
});
</script>
{/if}
<table class="ntable cur_achiev_table">
	<thead>
		{if[{var=done_achi_paginator}]}
		<tr>
            <th style="text-align: center" colspan="3">
                <form name="form_pages" id="form_pages" method="post" action="{@done_formaction}">
                <div style="float: right;">
                  <select name="page" id="done_page" onchange="document.getElementById('done_go').click();">{@done_pages}</select>
                </div>
                {@done_link_first}
                {@done_link_prev}
                {@done_page_links}
                {@done_link_next}
                {@done_link_last}
                <input type="submit" style="display: none;" name="go" id="done_go" value="{lang}COMMIT{/lang}" class="button" />
                </form>
            </th>
		</tr>
		{/if}
		<tr>
			<th colspan=3>
				{if[{var=achievement_id} && {var=is_achiev_other_user}]}
				{lang=ACHIEVEMENT}
				{else}
				{lang=DONE_ACHIEVEMENTS}{if[{var=cur_new_achieve_count} > 0]} <div style="float:right">{lang=NEW_ACHIEVEMENTS_PREFIX} <span class="true"><b>+{@cur_new_achieve_count}</b></span></div>{/if}
				{/if}
			</th>
		</tr>
		
	</thead>
	{foreach[cur_achieves]}
	<tbody class='cur_achiev_{loop=achievement_id}'>
	<tr>
		<td rowspan="{if[$row['reqs_exist']]}5{else}3{/if}" width="1px">
			{loop}image{/loop}
		</td>
		<td>
			<div style="width:100%">
				{loop}name{/loop}
			</div>
			<div style="clear:both; font-size:smaller">
				{loop}desc{/loop}
			</div>
		</td>
		
		<td rowspan="{if[$row['reqs_exist']]}5{else}3{/if}" width="100px" align="center">
			{if[!{var=is_achiev_other_user}]}
				{if[$row['state'] == ACHIEV_STATE_PROCESSED]}
				{else if[$row['granted'] > 0]}
					{if[ !empty($row['bonus_blocked']) ]}
						{if[ $row['state'] == ACHIEV_STATE_BONUS_GIVEN ]}
							{if[!{var=is_achiev_page}]}
								<input type='button' class="cur_achiev_hide_button button" name="{loop=achievement_id}" value="Скрыть">
							{else}
								<input type='button' class="cur_achiev_hide_button button cur_achiev_btn_{loop=achievement_id}" name="{loop=achievement_id}" value="ОК">
							{/if}
						{else if[ $row['bonus_build_type'] == ACHIEV_BONUS_BUILD_TYPE_PLANET ]}
							{lang}PLANET_BONUS_ONLY{/lang} <br />
						{else if[ $row['bonus_build_type'] == ACHIEV_BONUS_BUILD_TYPE_MOON ]}
							{lang}MOON_BONUS_ONLY{/lang} <br />
						{else}
							{lang}ERROR_ACHIEV_BONUS{/lang} <br />
						{/if}
						{if[$row['state'] == ACHIEV_STATE_ALERT]}
							{if[!{var=is_achiev_page}]}
								<input type='button' class="cur_achiev_hide_button button" name="{loop=achievement_id}" value="Скрыть">
							{else}
								<input type='button' class="cur_achiev_hide_button button cur_achiev_btn_{loop=achievement_id}" name="{loop=achievement_id}" value="ОК">
							{/if}
						{/if}
					{else if[{var=is_achiev_page}]}
						{if[ $row['state'] == ACHIEV_STATE_BONUS_GIVEN ]}
							<input type='button' class="cur_achiev_hide_button button cur_achiev_btn_{loop=achievement_id}" name="{loop=achievement_id}" value="ОК">
						{else}
							<form method="post" action="{@formaction}">
								<input type="hidden" name="achievement_process_id" value="{loop=achievement_id}">
								<input type="submit" name="process" value="Получить бонус" class="button">
							</form>
						{/if}
					{else}
						{if[ $row['state'] == ACHIEV_STATE_BONUS_GIVEN ]}
							<input type='button' class="cur_achiev_hide_button button" name="{loop=achievement_id}" value="Скрыть">
						{else}
							<form method="post" action="{@formaction}">
								<input type="hidden" name="achievement_get_bonus_id" value="{loop=achievement_id}">
								<input type="submit" name="get_bonus" value="Получить бонус" class="button">
							</form>
							<form method="post" action="{@formaction}">
								<input type="hidden" name="achievement_process_id" value="{loop=achievement_id}">
								<input type="submit" name="process" value="Получить бонус и скрыть" class="button">
							</form>
						{/if}
					{/if}
				{else if[$row['state'] == ACHIEV_STATE_ALERT]}
					{if[!{var=is_achiev_page}]}
						<input type='button' class="cur_achiev_hide_button button" name="{loop=achievement_id}" value="Скрыть">
					{else}
						<input type='button' class="cur_achiev_hide_button button cur_achiev_btn_{loop=achievement_id}" name="{loop=achievement_id}" value="ОК">
					{/if}
				{/if}
			{/if}
		</td>
	</tr>
	
	{if[$row['reqs_exist']]}
	<tr>
		<td>
			<b>{lang=REQUIRES}</b>
		</td>
	</tr>
	<tr>
		<td>
			<table class="table_no_background" cellspacing="0" cellpadding="0" border="0" style="clear:both">
				{foreach2[.req_items]}
					{if[$row["res_required"]]}
					<tr>
						<td>{loop=res_name}</td>
						<td class='{if[$row["res_notavailable"]]}notavailable{else}true{/if}'>{loop=res_required}</td>
						<td>{if[$row["res_notavailable"]]}({loop=res_notavailable}){/if}</td>
					</tr>
					{/if}
				{/foreach2}
			</table>
			<div style="clear:both">
				{loop}reqs{/loop}
			</div>
		</td>
	</tr>
	{/if}
	
	<tr class="for_pointer_bonus">
		<td>
			<b>{lang=ACHIEVEMENT_BONUS}</b>
		</td>
	</tr>
	<tr>
		<td>
			<table class="table_no_background" cellspacing="0" cellpadding="0" border="0" style="clear:both">
				{foreach2[.bonus_items]}
					<tr>
						<td>{loop=res_name}</td>
						<td class='true'>{loop=res_bonus}</td>
					</tr>
				{/foreach2}
			</table>
			<div style="clear:both">
				{loop}bonuses{/loop}
			</div>
		</td>
	</tr>
	</tbody>
	{/foreach}
</table>
{/if}
<p />
{/if}
{if[0]}
 Oxsar https://oxsar-nova.ru
 Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
{/if}