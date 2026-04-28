<form method="post" action="{@formaction}">
<table class="ntable">
	<tr>
		<th colspan="2">{lang}GLOBAL_MAIL{/lang}</th>
	</tr>
	<tr>
		<td><label for="receiver">{lang}RECEIVER{/lang}</label></td>
		<td><select name="receiver" id="receiver"><option value="foo">{lang}ALL_MEMBERS{/lang}</option>{while[ranks]}<option value="{loop}rankid{/loop}">{lang}RANK{/lang} {loop}name{/loop}</option>{/while}</select></td>
	</tr>
	<tr>
		<td><label for="subject">{lang}SUBJECT{/lang}</label></td>
		<td><input type="text" name="subject" id="subject" value="{lang}NO_SUBJECT{/lang}" maxlength="50" /><br />{@subjectError}</td>
	</tr>
	<tr>
		<td><label for="message">{lang}MESSAGE{/lang}</label></td>
		<td>
			<textarea name="message" id="message" cols="35" rows="8" onkeyup="maxlength(this,{config}MAX_PM_LENGTH{/config},'counter')"></textarea><br />
			{lang}MAXIMUM{/lang} <span id="counter">0</span> / {@maxpmlength} {lang}CHARACTERS{/lang}<br />{@messageError}
		</td>
	</tr>
	<tr>
		<td class="center" colspan="2"><input type="submit" name="send_global_message" value="{lang}COMMIT{/lang}" class="button" /></td>
	</tr>
</table>
</form>
{if[0]}
 Oxsar https://oxsar-nova.ru
 Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
  
{/if}