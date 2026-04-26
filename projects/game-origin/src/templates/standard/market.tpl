<form method="post" action="{@sendAction}">
<table class="ntable">
	<thead><tr>
		<th colspan="2">Обменник</th>
	</tr></thead>
	<tr>
		<td><img src="{const=RELATIVE_URL}images/market-money.png" alt=""/></td>
		<td>Галактический Обменник - это место где Вы можете обменять ресурсы которые имеются в избытке на срочно 
		    необходимые ресурсы.
		    <p/>С каждой сделки взымается комиссия в размере 
				{if[{var=comis} > 5]}
					<span class="false"><b>{@comis}%</b></span>
				{else}
					<span class="true"><b>{@comis}%</b></span>
				{/if}
			от продаваемого ресурса (для кредитов комиссия всегда 0%).
		    {if[{var=comis} > 5]}
				<p /><span class="false2">Активируйте артефакт Знак торговца, чтобы уменьшить комиссию до 3%</span>
		    {/if}
		</td>
	</tr>
</table>
<table class="ntable">
	<tr>
		<td colspan="4" class="center"><b>Что вы хотите обменять?</b></td>
	</tr>
	<tr>
		<td class="center">{image[METAL]}met.gif{/image}</td>
		<td class="center">{image[SILICON]}silicon.gif{/image}</td>
		<td class="center">{image[HYDROGEN]}hydrogen.gif{/image}</td>
		<td class="center">{image[CREDIT]}credit.gif{/image}</td>
	</tr>
	<tr>
		<td class="center"><input type="submit" name="a_metal" value="{lang}METAL{/lang}" class="button" /></td>
		<td class="center"><input type="submit" name="a_silicon" value="{lang}SILICON{/lang}" class="button" /></td>
		<td class="center"><input type="submit" name="a_hydrogen" value="{lang}HYDROGEN{/lang}" class="button" /></td>
		<td class="center"><input type="submit" name="a_credit" value="{lang}CREDIT{/lang}" class="button" /></td>
	</tr>
	<tr>
		<td colspan="4" class="center"><b>Курс ресурсов</b></td>
	</tr>
	<tr>
		<td colspan="4" class="center">{lang=METAL} : {lang=SILICON} : {lang=HYDROGEN} : {lang=CREDIT}
		  = {@curs_metal} : {@curs_silicon} : {@curs_hydrogen} : {@curs_credit}</td>
	</tr>
  <tr>
    <td colspan="4" class="center">
      {if[{var=is_planet_ratio_base}]}
      Ваш рейтинг добычи ресурсов на планете: {@planet_ratio_base_procent}%. Стройте больше рудников 
      и очень скоро количество ресурсов по отношению к кредиту начнет увеличиваться.
      {else}
      За 100 кредитов Вы можете приобрести суточную выработку всех ресурсов на планете.
      Стройте больше рудников, чтобы увеличить объемы ресурсов, приобретаемые за единицу кредита.
      {/if}
    </td>

  </tr>
</table>
</form>
{if[0]}
 Oxsar http://oxsar.ru
 Copyright (c) 2009-2010 UnitPoint <support@unitpoint.ru>
  
{/if}