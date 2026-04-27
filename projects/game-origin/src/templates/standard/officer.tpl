<form method="post" action="{@sendAction}">
<input type="hidden" id="off_id" name="off_id">
<table class="ntable">
	<thead><tr>
		<th colspan = "3">Офицеры</th>
	</tr></thead>
	<tr>
		<td width="120"><img src="{const=RELATIVE_URL}images/officer/01.jpg" alt=""/></td>
		<td><b>Торговый представитель</b><br/>Cамый опытный представитель своей профессии, обладает навыками торговли, что позволяет уменьшить процент комиссии при торговле ресурсами на рынке<p/>Уменьшает комиссию на рынке с 20% до 3%.<p/>Стоимость найма на 7 дней: {@off_cost_1} кредитов<p/>Нанят до: {@off_1}</td>
		<td width="50" class = "center">{@get1}</td>
	</tr>
	<tr>
		<td width="120"><img src="{const=RELATIVE_URL}images/officer/02.jpg" alt=""/></td>
		<td><b>Шахтёр</b><br/>Эксперт в астроминералогии и кристаллографии. Со своей командой металлургов и химиков он развивает новые источники ресурсов и оптимизацию их очистки.<p/>Увеличивает добычу ресурсов шахтами Металла, Кремния и Водорода на 20%.<p/>Стоимость найма на 7 дней: {@off_cost_2} кредитов<p/>Нанят до: {@off_2}</td>
		<td width="50" class = "center">{@get2}</td>
	</tr>
	<tr>
		<td width="120"><img src="{const=RELATIVE_URL}images/officer/03.jpg" alt=""/></td>
		<td><b>Энергетик</b><br/>Благодаря знаниям в сфере новых технологий производства энергии позволяет получать больше энергии на планете при тех же затратах ресурсов.<p/>Увеличивает добычу энергии на планете на 20%.<p/>Стоимость найма на 7 дней: {@off_cost_3} кредитов<p/>Нанят до: {@off_3}</td>
		<td width="50" class = "center">{@get3}</td>
	</tr>
	<tr>
		<td width="120"><img src="{const=RELATIVE_URL}images/officer/04.jpg" alt=""/></td>
		<td><b>Кладовщик</b><br/>Эксперт в управлении хранилищами ресурсов. Благодаря своей команде профессионалов тщательно сортирует ресурсы по хранилищам, тем самым позволяя увеличить вместительность складов не увеличивая их.<p/>Увеличивает ёмкость хранилищ на 20%.<p/>Стоимость найма на 7 дней: {@off_cost_4} кредитов<p/>Нанят до: {@off_4}</td>
		<td width="50" class = "center">{@get4}</td>
	</tr>
</table>
</form>
{if[0]}
 Oxsar https://oxsar-nova.ru
 Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
  
{/if}