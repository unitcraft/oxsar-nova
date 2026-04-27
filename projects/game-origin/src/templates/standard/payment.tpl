{if[ defined('SN') && defined('SN_FULLSCREEN') && ({var=isOdnoklassnikiOpen} || {var=isMailruOpen} || {var=isVkontakteOpen}) ]}

<div id='fullcsreen_payment_dialog'>
	<p style="text-align: center;">
	<img src="{const=RELATIVE_URL}images/market-money.png" alt="">
	<p />
	<p style="text-align: center;">
	Чтобы пополнить кредиты, закройте полноэкранный режим и/или перейдите в игру по ссылке
	<br /><br />
	<?php /* план 37.5d.9: gameUrl убран — соцсетевой URL */ ?>
	<br /><br />
	Затем кликните Пополнить. После этого, Вы можете снова вернуться в полноэкранный режим.
	<p />
</div>
<script type="text/javascript">
	$('#fullcsreen_payment_dialog').dialog({
		title: 'Внимание',
		width: 500,
		height: 320,
		resizable: false,
		dragable: false
	});
</script>

{else if[{var=tooManyBought}]}
<table class="ntable">
  <thead>
    <tr>
      <th>
        Пополнить счет
      </th>
    </tr>
  </thead>
  <tr>
    <td>
      Мы очень ценим Ваше желание помочь развитию игры, но большое пополнение кредитов может отрицательно сказаться на балансе игры в целом.
	  Пожалуйста пополните кредиты чуть позже. Приносим Вам свои извинения.
      {if[{var=is_admin}]}
      <p />
      Максимальное количество купленных кредитов: {@other_bought_credit} ({@other_username})
      {if[{var=max_credit_available} != {var=other_bought_credit}]}
      <br /> Реальный лимит для покупки: {@max_credit_available}
      {/if}
      <br /> Икрок уже приобрел: {@cur_bought_credit} ({@cur_username})
      {/if}
    </td>
  </tr>
</table>

{else}

{if[defined('SN')]}
<script type="text/javascript">
{if[SN == ODNK_SN_ID]}
function post_init_process()
{
	FAPI.UI.showPayment('{@ok_title}', '{@ok_desc}', '{@oc_code}1', '{@ok_ammount}1', '{@odnoklassnikiOptions}', null, 'ok', 'false');
	{if[ defined('SN_FULLSCREEN') ]}
		setTimeout(function(){ window.close(); }, 3000);
	{/if}
};
{/if}
$(function(){
	$('#payment_social_waiter').css('display', 'none');
});
</script>
{/if}
  {if[!{var=isPay2iFrameOpen}]}
  <table class="ntable">
    <thead>
      <tr>
        <th colspan="3">
          Пополнить счет</th>
      </tr>
    </thead>
    <tr>
      <td>
        <img src="{const=RELATIVE_URL}images/market-money.png" alt="" /></td>
      <td width="100%">
        {lang=PAY_MAIN_TEXT}
      </td>
    </tr>
  </table>
  {/if}

    {if[{var=isOdnoklassnikiOpen}]}
		{if[count(Core::getTPL()->getLoop("odnoklassniki_credits")) > 0]}
			<table class="ntable center">
				<thead>
					<tr>
						<th colspan="2">{lang}CREDITS{/lang}</th>
						<th></th>
					</tr>
				</thead>
				<tbody>
					{foreach[odnoklassniki_credits]}
						<tr>
							<td>{loop}title{/loop}</td>
							<td>за <b class="confederation">{loop}price2show{/loop}</b> ОК</td>
							<td><input type="button" class="button" title="{lang}BUY{/lang}" value="{lang}BUY{/lang}"
								onclick="show_odnoklassniki_payment('{loop}name{/loop}', '{loop}code{/loop}', '{loop}name{/loop}', {loop}price{/loop});">
							</td>
						</tr>
					{/foreach}
				</tbody>
			</table>
		{/if}
	{/if}

    {if[{var=isMailruOpen}]}
		{if[count(Core::getTPL()->getLoop("mailru_credits")) > 0]}
			<table class="ntable center">
				<thead>
					<tr>
						<th colspan="2">{lang}CREDITS{/lang}</th>
						<th></th>
					</tr>
				</thead>
				<tbody>
					{foreach[mailru_credits]}
						<tr>
							<td>{loop}title{/loop}</td>
							<td>за <b class="confederation">{loop}price2show{/loop}</b> руб</td>
							<td><input type="button" class="button" title="{lang}BUY{/lang}" value="{lang}BUY{/lang}"
								onclick="show_mailru_payment('{loop}name{/loop}', '{loop}code{/loop}', '{loop}name{/loop}', {loop}price{/loop});">
							</td>
						</tr>
					{/foreach}
				</tbody>
			</table>
		{/if}
	{/if}

    {if[{var=isVkontakteOpen}]}
		{if[count(Core::getTPL()->getLoop("vkontakte_credits")) > 0]}
			<table class="ntable center">
				<thead>
					<tr>
						<th colspan="2">{lang}CREDITS{/lang}</th>
						<th></th>
					</tr>
				</thead>
				<tbody>
					{foreach[vkontakte_credits]}
						<tr>
							<td>{loop}title{/loop}</td>
							<td>за <b class="confederation">{loop}price2show{/loop}</b>
							{if[ ($row['price'] < 10 || $row['price'] >= 20) && ($row['price'] % 10) == 1 ]}голос{else if[ ($row['price'] % 10) > 0 && ($row['price'] % 10) <= 4 ]}голоса{else}голосов{/if}</td>
							<td><input type="button" class="button" title="{lang}BUY{/lang}" value="{lang}BUY{/lang}"
								onclick="show_vkontakte_payment('{loop}name{/loop}', '{loop}code{/loop}', '{loop}name{/loop}', {loop}price{/loop});">
							</td>
						</tr>
					{/foreach}
				</tbody>
			</table>
			<div id='payment_progress_dialog'>
				<p style="text-align: center;">
				<img src="/images/ajax-loader_2.gif">
				<p />
				<p style="text-align: center;">
				Выполняется обработка платежа ...
				<p />
			</div>
			<script type="text/javascript">
				$('#payment_progress_dialog').dialog({
					autoOpen: false,
					resizable: false,
					title: 'Пополнить кредиты',
					width: 300,
					dragable: false
				});
			</script>
		{/if}
	{/if}

  	<table class="ntable">

    {if[{var=isOAuth2OdnoklassnikiOpen}]}
	<form method="post" action="{@payAction}" target="_blank">
    <tr>
      <td class="center">
		&nbsp;<br />
	    <input type="image" src="/images/pay_logos/odnk.png" class="button" name="Webmoney" title="Одноклассники">
		<br /><br />
        <span class="false2"><b>Комиссия 15%</b></span><br />
        После открытия игры на сайте Одноклассников, нужно нажать Пополнить еще раз.
		{if[0]}<br /><br />
		<input type="submit" class="button" name="OAuth2_Odnoklassniki" value="Продолжить" />{/if}
		<br />&nbsp;
	  </td>
    </tr>
	</form>
	{/if}

    {if[{var=isOAuth2MailruOpen}]}
	<form method="post" action="{@payAction}" target="_blank">
    <tr>
      <td class="center">
		&nbsp;<br />
	    <input type="image" src="/images/pay_logos/mailru.png" class="button" name="Webmoney" title="Мой Мир @ mail.ru">
		<br /><br />
        <span class="false2"><b>Комиссия 15%</b></span><br />
        После открытия игры в Моем Мире, нужно нажать Пополнить еще раз.
		{if[0]}<br /><br />
		<input type="submit" class="button" name="OAuth2_Mailru" value="Продолжить" />{/if}
		<br />&nbsp;
	  </td>
    </tr>
	</form>
	{/if}

    {if[{var=isOAuth2VkontakteOpen}]}
	<form method="post" action="{@payAction}" target="_blank">
    <tr>
      <td class="center">
		&nbsp;<br />
	    <input type="image" src="/images/pay_logos/vkontakte.png" class="button" name="Webmoney" title="Вконтакте">
		<br /><br />
        <span class="false2"><b>Комиссия 15%</b></span><br />
        После открытия игры на сайте Вконтакте, нужно нажать Пополнить еще раз.
		{if[0]}<br /><br />
		<input type="submit" class="button" name="OAuth2_Vkontakte" value="Продолжить" />{/if}
		<br />&nbsp;
	  </td>
    </tr>
	</form>
	{/if}

    {if[{var=isWebmoneyOpen}]}
	<form method="post" action="{@formaction}Webmoney">
    <tr>
      <td class="center">
		&nbsp;<br />
	    <input type="image" src="/images/pay_logos/webmoney.png" class="button" name="Webmoney" title="WebMoney">
		<br /><br />
        Оплата в WMR, WMZ, WME. <span class="true"><b>Комиссия 0%</b></span>
		{if[0]}<br /><br />
		<input type="submit" class="button" name="Webmoney" value="Продолжить" />{/if}
		<br />&nbsp;
	  </td>
    </tr>
	</form>
	{/if}

    {if[{var=isPay2iFrameOpen}]}
		<tr>
		  <td width="100%" height="700px">
            <iframe id="paystation" src="{@pay2_iframe_url}" width="100%" height="100%"></iframe>
          </td>
        </tr>
    {/if}
    {if[{var=isPay2Open} && !{var=isPay2iFrameOpen}]}
		<form method="post" action="{@pay_2_process_url}" target="_blank">
		<input type="hidden" name="user_id" value="{@user_id}">
		<input type="hidden" name="{const=PS_GAME_DOMAIN_PARAM_NAME}" value="{const=PS_GAME_DOMAIN}">
		<tr>
		  <td class="center">
			&nbsp;<br />
			<input type="image" src="/images/pay_logos/xsolla.png" class="button" name="The2pay1" title="xsolla">
			<br /><br />
			{if[ !SUPER_CREDIT_BONUS_ENABLED ]}
			<font color="#00FF00"><b>Оплата с использованием практически любых способов:</b></font>
			{else}
			<font class="false2">Акция не распространяется на XSOLLA!</font>
			{/if}
			<br /><b>СМС</b> (<b>Россия</b>, <b>Украина</b>, Казахстан, Германия и мн. др. страны), <b>Электронные деньги</b>, <b>VISA</b>,
			<b>телебанк ВТБ24</b>, <b>PayPal</b>, <b>Liqpay</b>, Терминалы оплаты и др. Кредиты через 2pay начисляются обычно с задержкой в 5 минут после платежа.
			Комиссия зависит от выбранной системы оплаты.
			{if[0]}<br /><br />
			<input type="submit" class="button" name="The2pay1" value="Продолжить" />{/if}
			<br />&nbsp;
		  </td>
		</tr>
		</form>
    {/if}

    {if[{var=isA1Open}]}
	<form method="post" action="{@formaction}SMS">
    <tr>
      <td class="center">
		&nbsp;<br />
        <input type="image" src="/images/pay_logos/a1a.png" class="button" name="SMS" title="A1A SMS">
		<br /><br />
        Оплата через отправку СМС. Комиссия зависит от оператора.
		{if[0]}<br /><br />
        <input type="submit" class="button" name="SMS" value="Продолжить" />{/if}
		<br />&nbsp;
      </td>
    </tr>
	</form>
	{/if}

    {if[{var=isRoboKassaOpen}]}
	<form method="post" action="{@formaction}Robokassa">
    <tr>
      <td class="center">
		&nbsp;<br />
        <input type="image" src="/images/pay_logos/robokassa.png" class="button" name="ROBOKASSA" title="ROBOKASSA">
		<br /><br />
        Оплата многими способами: Яндекс.Деньги, RUR RBK Money, EasyPay, RUR PayExpress,
        RUR MoneyMail, RUR Единый Кошелек, СМС, Терминалы оплаты и др. Комиссия зависит от выбранной системы оплаты.
		{if[0]}<br /><br />
        <input type="submit" class="button" name="ROBOKASSA" value="Продолжить" />{/if}
		<br />&nbsp;
	  </td>
    </tr>
	</form>
    {/if}

  </table>
{/if}
{if[0]}
 Oxsar https://oxsar-nova.ru
 Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.

{/if}