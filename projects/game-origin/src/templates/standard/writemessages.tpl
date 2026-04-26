<form method="post" action="{@sendAction}">
<table class="ntable">
	<thead><tr>
		<th colspan="2">{lang}NEW_MESSAGE{/lang}</th>
	</tr></thead>
	<tbody><tr>
		<td><label for="receiver">{lang}RECEIVER{/lang}</label></td>
        <td><?php echo CHtml::tag('input', array('name' => 'receiver', 'id' => 'receiver', 'value' => {var=receiver}, 'maxlength' => Core::getOptions()->get("MAX_USER_CHARS"))); ?><br />{@userError}</td>
		{if[0]}<td><input type="text" name="receiver" id="receiver" value=<?php echo CJavaScript::encode({var=receiver}); ?> maxlength="{config}MAX_USER_CHARS{/config}" /><br />{@userError}</td>{/if}
	</tr>
	<tr>
		<td><label for="subject">{lang}SUBJECT{/lang}</label></td>
		<td><input type="text" name="subject" id="subject" value="{@subject}" maxlength="50" /><br />{@subjectError}</td>
	</tr>
	<tr>
		<td><label for="message">{lang}MESSAGE{/lang}</label></td>
		<td>
			<textarea name="message" id="message" cols="35" rows="8" onkeyup="maxlength(this,{config}MAX_PM_LENGTH{/config},'counter')"></textarea><br />
			{lang}MAXIMUM{/lang} <span id="counter">0</span> / {@maxpmlength} {lang}CHARACTERS{/lang}<br />{@messageError}
		</td>
	</tr></tbody>
	<tfoot><tr>
		<td colspan="2"><input type="submit" name="send" value="{lang}COMMIT{/lang}" class="button" /></td>
	</tr></tfoot>
</table>
</form>
{if[0]}
 Oxsar http://oxsar.ru
 Copyright (c) 2009-2010 UnitPoint <support@unitpoint.ru>

{/if}