<table class="ntable">
	<colgroup>
		<col width="60%"/>
		<col width="40%"/>
	</colgroup>
	<thead>
		<tr>
			<th>{lang}FOLDER{/lang}</th>
			<th>{lang}MESSAGES{/lang}</th>
		</tr>
	</thead>
	<tfoot>
		<tr>
			<td colspan="3" class="center"><a href="{const=RELATIVE_URL}game.php/MSG/DeleteAll" onclick="return confirm('{lang=CONFIRM_DELETE_ALL}')">{lang=DELETE_ALL}</a> {link[CREATE_NEW_MESSAGE]}"game.php/MSG/Write"{/link}</td>
		</tr>
	</tfoot>
	<tbody>{foreach[folders]}
		<tr>
			<td>{loop=image} {loop=label}</td>
			<td>{loop=messages} {loop=newMessages}</td>
		</tr>
	{/foreach}</tbody>
</table>
{if[0]}
 Oxsar http://oxsar.ru
 Copyright (c) 2009-2010 UnitPoint <support@unitpoint.ru>

{/if}