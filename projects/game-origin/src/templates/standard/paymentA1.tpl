<form method="post" action="">
<table class="ntable">
	<thead><tr>
		<th colspan="3">Пополнить счет</th>
	</tr></thead>
	<tr>
		<td><img src="{const=RELATIVE_URL}images/market-money.png" alt=""/></td>
    <td width="100%">
      {lang=PAY_MAIN_TEXT}
    </td>
  </tr>
</table>
<table class="ntable">
{if[{var=isPaymentLocked}]}
	<tr>
		<td colspan="3" class="center">Временно отключено.</td>
	</tr>
{else}
	<tr>
		<td colspan="3" class="center">Выберите вашу страну:
			<select name="country" id="country">
			{foreach[countries]}
				<option value="{loop}name{/loop}">{loop}name{/loop}</option>
			{/foreach}
			</select>
		</td>
	</tr>
	<tr>
		<td colspan="3" class="center"><input type="submit" name="SMS2" value="Продолжить" class="button" /></td>
	</tr>
{/if}
	<tr>
		<td colspan="3">{lang=PAY_SUPPORT_TEXT}</td>
	</tr>
</table>
</form>
{if[0]}
 Oxsar https://oxsar-nova.ru
 Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
  
{/if}