<p />
<table class="ntable">
	<tr>
		<th style="text-align: center;">
			{lang=NOTEPAD}
		</th>
	</tr>
	<tr>
		<td style="text-align: center;">
			В блокноте можно сохранять любую текстовую информацию, например, заметки, планы и т.п.
			<br /> Эта информация доступна только Вам.
		</td>
	</tr>
	<tr>
		<td style="text-align: center;">
			<form method="post" action="{@formaction}">
				<textarea rows="{if[!defined('SN')]}20{else}17{/if}" name="notes" style="width: 90%;">{@notes}</textarea>
				<br/>
				<input type="submit" value="{lang=OK}">
			</form>
		</td>
	</tr>
</table>
{if[0]}
 Oxsar http://oxsar.ru
 Copyright (c) 2009-2010 UnitPoint <support@unitpoint.ru>
 
{/if}