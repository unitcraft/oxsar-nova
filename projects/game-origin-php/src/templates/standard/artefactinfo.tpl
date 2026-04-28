<form method="post" action="{@formaction}">
<input type="hidden" id="artid" name="artid" value="0" />
<input type="hidden" id="typeid" name="typeid" value="{@typeid}" />
<table class="ntable">
	<tr>
		<th colspan="2">{@name}</th>
	</tr>
	<tr>
		<td colspan="2">
			{if[{var}flags{/var}]}<span style="float:right">{@flags}</span>{/if}
			{@pic}{@description}<br />
		  {if[{var}lifetime{/var}]}<i>Время существования: {@lifetime}</i><br />{/if}
			{if[{var}duration{/var}]}<i>Длительность эффекта: {@duration}</i><br />{/if}
			{if[{var}usable{/var}&&({var}use_times{/var}!=1)]}<i>Количество использований: {@use_times}</i><br />{/if}
			{if[{var}delay{/var}&&({var}use_times{/var}!=1)]}<i>Перерыв между использованиями: {@delay}</i><br />{/if}
		</td>
	</tr>
	{foreach[artefacts]}<tr>
		<td{if[!$row['use']&&!$row['activate']&&!$row['deactivate']&&!$row['counter']]} colspan="2"{/if}>
			Местоположение: {loop}location{/loop}<br />
			{if[$row['life_counter']]}Время существования: {loop}life_counter{/loop}<br />{/if}
			{if[!{var}usable{/var}||{var}duration{/var}]}
				{if[$row['active']]}Активен{if[{var}usable{/var}&&$row['eff_counter']]} (осталось {loop}eff_counter{/loop}){/if}{else}Неактивен{/if}<br />
			{/if}
			{if[{var}usable{/var}]}Осталось использовать {loop}times_left{/loop} раз из {@use_times}<br />{/if}
		</td>
    {if[$row['use']]}
			<td class="center" width=10%>
        <input type="submit" name="use" value="{lang}USE{/lang}" class="button" onClick="document.getElementById('artid').value={loop}id{/loop}" />
			</td>
    {/if}
		{if[$row['activate']]}
			<td class="center" width=10%>
				<input type="submit" name="activate" value="{lang}ACTIVATE{/lang}" class="button" onClick="document.getElementById('artid').value={loop}id{/loop}" />
			</td>
		{/if}
		{if[$row['deactivate']]}
			<td class="center" width=10%>
				<input type="submit" name="deactivate" value="{lang}DEACTIVATE{/lang}" class="button" onClick="document.getElementById('artid').value={loop}id{/loop}" />
			</td>
		{/if}
		{if[$row['counter']]}
			<td class="center" width=10%>
				{loop}counter{/loop}
			</td>
		{/if}
	</tr>{/foreach}
	{if[count($this->getLoop("entries")) > 0]}
		<tr>
			<th colspan="2">{lang}HISTORY{/lang}</th>
		</tr>
		{foreach[entries]}<tr>
			<td colspan="2">
				<b>{loop}time{/loop}</b><br />
				Отбит игроком {loop}won{/loop} у {loop}lost{/loop}.
				{if[!empty($row["assault_url"])]}<span style="float:right"><a href="{loop}assault_url{/loop}">[ Показать бой ]</a></span>{/if}
			</td>
		</tr>{/foreach}
	{/if}
</table>
</form>
{if[0]}
 Oxsar https://oxsar-nova.ru
 Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
  
{/if}