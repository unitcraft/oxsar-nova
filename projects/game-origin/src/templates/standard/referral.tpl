<form method="post" action="{@sendAction}">
<table class="ntable">
	<thead><tr>
		<th colspan="5">{lang=MENU_REFERAL}</th>
	</tr></thead>
	<tr>
		{if[!defined('SN')]}
		<td><b>Реф. ссылка</b><br/>
		<td colspan="4"><b>{const=BASE_FULL_URL}profile{user=userid}</b></td>
		{else}
		<td colspan="5" class='center'>
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
		{/if}
	</tr>
	{if[$this->templateVars["is_bonus_active"]]}
	<tr>
		<td valign="top"><font class="true"><b>Описание</b></font></td>
		<td colspan="4">
			{if[ !defined('SN') ]}
			Приглашайте по этой ссылке друзей и знакомых. После регистрации друга в игре по данной ссылке, он станет Вашим рефералом.
			{else}
			Приглашайте в игру друзей и знакомых. После регистрации друга в игре, он станет Вашим рефералом.
			{/if}
			<p />Реферальные бонусы начисляются, когда реферал наберёт {@ref_bonus_points} очков. 
			<p />Кредитный бонус составляет <font color = "#00FF00"><b>20%</b></font> от каждого пополнения рефералом кредитов!
			<p />Дополнительный ресурсный бонус начисляется в случае, если количество Ваших собственных очков меньше {@max_bonus_points}.
			Он составляет:
			<font color = "#00FF00"><b>{@p_metal}%</b></font> металла, 
			<font color = "#00FF00"><b>{@p_silicon}%</b></font> кремния и 
			<font color = "#00FF00"><b>{@p_hydrogen}%</b></font> водорода от количества Ваших очков в момент начисления бонуса.
			{if[ {var=w_Metal} > 0 || {var=w_Silicon} > 0 || {var=w_Hydrogen} > 0 ]}
			<p />Если бы ресурсный бонус начислился сейчас, то он составил бы:
			<br/> {@w_Metal} металла, {@w_Silicon} кремния и {@w_Hydrogen} водорода.
			{/if}
		</td>
	</tr>
	<tr>
		<td valign="top"><font class="true"><b>Рекомендации</b></font></td>
		<td colspan="4">
			Расскажите другу об игре, о ее плюсах. Сообщите, что Вы поможете
			разобраться с игрой и подбросите немного ресурсов для развития (не забудьте реально помочь после того, как друг зарегистрируется).
		</td>
	</tr>
   	{else}
   	<tr>
	   	<td colspan="5">
			{if[ !defined('SN') ]}
			Приглашайте по этой ссылке друзей и знакомых. После регистрации друга в игре по данной ссылке, он станет Вашим рефералом.
			{else}
			Приглашайте в игру друзей и знакомых. После регистрации друга в игре, он станет Вашим рефералом.
			{/if}
			<p />Реферальные бонусы начисляются, когда реферал наберёт {@ref_bonus_points} очков. 
			<p />Кредитный бонус составляет <font color = "#00FF00"><b>20%</b></font> от каждого пополнения рефералом кредитов!
			{if[0]}Т.к. у Вас уже больше {@max_bonus_points} очков, ресурсная часть бонуса не начисляется.{/if}
			<p /><b>Рекомендации</b>
			<br />Расскажите другу об игре, о ее плюсах. Сообщите, что Вы поможете
			разобраться с игрой и подбросите немного ресурсов для развития (не забудьте реально помочь после того, как друг зарегистрируется).
		</td>
	</tr>
	{/if}
	<tr></tr>
	<tr>
		<td colspan="5" class="center"><b>Список привлечённых игроков</b></td>
	</tr>
	<tr>
		<td width="25%" class="center"><b>Регистрация</b></td>
		<td width="35%" class="center"><b>Реферал</b></td>
		<td width="20%" class="center"><b>Очки</b></td>
		<td width="10%" class="center"><b>Кредиты</b></td>
		<td width="10%" class="center"><b>Бонус</b></td>
	</tr>
	{foreach[shoutReferral]}
	<tr>
		<td>{loop}ref_time{/loop}</td>
		<td>{loop}username{/loop}</td>
		<td>{loop}points{/loop}</td>
		<td>{loop}bonus_credit{/loop}</td>
		<td class="center">
			{loop=bonus_img}
		</td>
	</tr>
	{/foreach}
</table>
</form>
{if[0]}
 Oxsar http://oxsar.ru
 Copyright (c) 2009-2010 UnitPoint <support@unitpoint.ru>
  
{/if}