<form method="post" action="{@formaction}">
<table class="ntable">
	<tr>
		<th>{lang}BROWSE_UNIVERSE{/lang}</th>
	</tr>
	<tr>
		<td>
			<select name="where"><option value="1"{@players}>{lang}PLAYERS{/lang}</option><option value="2"{@planets}>{lang}PLANETS{/lang}</option><option value="3"{@allys}>{lang}ALLIANCES{/lang}</option></select>
			<input type="text" name="what" maxlength="128" value="{@what}" class="searchInput" />
			<input type="submit" name="seek" value="{lang}COMMIT{/lang}" class="button" />
		</td>
	</tr>
</table>
</form>
{if[0]}
 Oxsar https://oxsar-nova.ru
 Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
  
{/if}