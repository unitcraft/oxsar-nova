<table class="ntable">
	<tr><th colspan="3">
		{if[ NS::getPlanet()->getPlanetId() ]}
			<b>{@currentPlanet}</b> {@currentCoords} ({user}username{/user})
		{else}
			{user}username{/user}
		{/if}
		{if[0]}<span style="float:left">{lang}PLANET{/lang} "{@planetNameLink}" ({user}username{/user})</span>{/if}
        {if[isAdmin()]}
		<span style="float:right">Сейчас играют: <span class="false"><?php
			// План 37.5d.9: replaced User_YII::showOnline() (Yii AR + cache).
			// na_user_online — pre-aggregated VIEW/таблица из cron-job.
			$online = sqlSelectRow("user_online", "*", "", "");
			echo fNumber($online ? $online['online_15'] : 0);
		?></span> | За 24 часа: <span class="false"><?php
			echo fNumber($online ? $online['core_1'] : 0);
		?></span></span>
        {/if}
        </th>
	</tr>

	{if[!{var=donated}]}
	{/if}

	<tr>
		<td>{lang}CUR_HOME_PLANET{/lang}</td>
		<td colspan="2">{@homeplanet}<span style="float:right">{@universe_name_full}</span></td>
	</tr>
	<tr>
		<td>{lang}MENU_PROFESSION{/lang}</td>
		<td colspan="2">
            <?php
                $u = htmlspecialchars(socialUrl(RELATIVE_URL . 'game.php/Profession'));
                echo '<a href="' . $u . '">' . htmlspecialchars(NS::getProfessionName()) . '</a>';
            ?>
        </td>
	</tr>
	<tr>
		<td>{lang}SERVER_TIME{/lang}</td>
		<td colspan="2">
            <span style="float:right">
                <?php /* Universe switcher: убран — заменим в plan-37.5 на vanilla JS */ ?>
            </span>
            <span id="serverwatch">{@serverClock}</span>
        </td>
	</tr>

	{include}"main_news_rows"{/include}

    {if[$this->templateVars["unreadedmsg"]]}
    <tr>
		<td colspan="3" class="center">{@newMessages}</td>
	</tr>{/if}
    {if[{var=mini_games}]}
    <tr><th colspan="3">Мини игры, пока летит флот</th></tr>
    <tr><td colspan="3" align="center">{@mini_games}</td></tr>
    {/if}
	<tr><th colspan="3">{lang}EVENTS{/lang}</th></tr>
	{foreach[fleetEvents]}
		<tr>
			<td class="center">
				{if[ !isMobileSkin() && $row['event_pb_value'] ]}
				<div id="evpb{loop=eventid}" style="height:14px"></div>
				{/if}
			  <span id="timer_{loop}eventid{/loop}">{loop}time{/loop}</span>
			  {if[$row["control_eventid"]]}
			  <br />
			  <form method="post" action="{@control_action}">
				<input type="submit" name="control" value="{lang}CONTROL_EVENT{/lang}" class="button" />
				<input type="hidden" name="id" value="{loop=control_eventid}" />
			  </form>
			  {else if[$row["retreat_eventid"]]}
			  <br />
			  <form method="post" action="{@formaction}">
				<input type="submit" name="retreat" value="{lang}RETREAT{/lang}" class="button" />
				<input type="hidden" name="id" value="{loop=retreat_eventid}" />
			  </form>
			  {/if}
			</td>
			<td colspan="2">
			  <span class="{loop}class{/loop}">{loop}message{/loop}</span>
			</td>
		</tr>
	{/foreach}

{if[ !isMobileSkin() ]}
<script type="text/javascript">
$(function(){
	{foreach[fleetEvents]}
		{if[ $row['event_pb_value'] ]}
			$("#evpb{loop=eventid}").progressbar({
				value: {loop=event_pb_value}
			});
			setInterval(function(){
				var value = $("#evpb{loop=eventid}").progressbar("option", "value");
				if(value < 100)
				{
					$("#evpb{loop=eventid}").progressbar("option", "value", value+1);
				}
			}, {loop=event_percent_timeout});
		{/if}
	{/foreach}
});
</script>
{/if}

	{if[ NS::getPlanet()->getPlanetId() ]}

		<tr>
			<td class="center">{@moon}<br />{@moonImage}</td>
			<td class="center">{@planetImage}{if[ NS::getPlanet()->getPlanetId() ]}<br />{@planetAction}{/if}</td>
			<td class="center">{include}"planetMainSelection"{/include}</td>
			{if[0]}
			<td class="center"{if[isFacebookSkin()]} colspan='2'{/if}>{@planetImage}<br />{@planetAction}</td>
			{if[!isFacebookSkin()]}<td class="center">{include}"planetMainSelection"{/include}</td>{/if}
			{/if}
		</tr>

		<tr>
			<td>{lang}PLANETDIAMETER{/lang}</td>
			<td colspan="2">{@planetDiameter} {lang}PLANETDIAMETER_KM{/lang} ({lang}PLANET_OCCUPIED_FIELDS{/lang})</td>
		</tr>
		<tr>
			<td>{lang}TEMPERATURE{/lang}</td>
			<td colspan="2">{@planetTemp} &deg;C</td>
		</tr>
		<tr>
			<td>{lang}POSITION{/lang}</td>
			<td colspan="2">{@planetPosition}</td>
		</tr>

		<tr>
			<td>{lang}BATTLE_EXPERIENCE{/lang}</td>
			<td colspan="2">{@e_points}</td>
		</tr>
		<tr>
			<td>{lang}BATTLE_ACTIVE_EXPERIENCE{/lang}</td>
			<td colspan="2">{@be_points}</td>
		</tr>
		<tr>
			<td>{lang}POINTS{/lang}</td>
			<td colspan="2">{@points} ({lang}RANK_OF_USERS{/lang})</td>
		</tr>
        {if[{var=max_points_enabled}]}
		<tr>
			<td>{lang}MAX_POINTS{/lang}</td>
			<td colspan="2">{@max_points}</td>
		</tr>
        {/if}
        {if[{var=dm_points_enabled}]}
		<tr>
			<td>{lang}DM_POINTS{/lang}</td>
			<td colspan="2">{@dm_points}</td>
		</tr>
        {/if}
		<tr>
			<td>Шахтёр ({@offLevel})</td>
			<td colspan="2">{@offPoints} / {@need_points}</td>
		</tr>
		{if[ {var=reseachVirtLab} ]}
		<tr>
			<td>Уровень межгалактических исследований</td>
			<td colspan="2">{@reseachVirtLab}</td>
		</tr>
		{/if}
		{if[!defined('SN')]}
		{else}
		<tr>
		<td colspan="3" class='center'>
			<button id='invite_friend_button'>
				<span class='ui-button-text'>Пригласить друга</span>
			</button>
			<script type="text/javascript">
			$(function () {
				$( "#invite_friend_button" ).button();
				$( "#invite_friend_button" ).live('click', function(){
					{if[SN == ODNK_SN_ID]}
						show_odnoklassniki_invite();
					{else if[SN == VKNT_SN_ID]}
						show_vkontakte_invite();
					{else if[SN == MAILRU_SN_ID]}
						show_mailru_invite();
					{/if}
					return false;
				});
			});
			</script>
		</td>
		</tr>
		<tr>
			<td colspan="3" align="center">
			  <p /><img src="{const=FULL_URL}userbar/{user=userid}.jpg" alt="Oxsar - новая космическая онлайн стратегия" {if[isMobileSkin()]}width="95%"{/if} />
			  <p /><b>BBCODE юзербара</b>:
			  <br /><textarea id="userbar_bbcode_text" cols="50" rows="3">[url={const=BASE_FULL_URL}][img]{const=FULL_URL}userbar/{user=userid}.jpg[/img][/url]</textarea>
			  <br />Разместите Ваш юзербар в своем блоке или в подписи на форумах.
			</td>
		</tr>
		{/if}
	{/if}
</table>
<script type="text/javascript">
//<![CDATA[
$(function () {
{foreach[fleetEvents]}
	$('#timer_{loop}eventid{/loop}').countdown({until: {loop}time_r{/loop}, compact: true, onExpiry: function() {
		$('#timer_{loop}eventid{/loop}').text('-');
	}});
{/foreach}
});
//]]>
</script>
{if[0]}
 Oxsar https://oxsar-nova.ru
 Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.

{/if}