<form method="post" action="{@formaction}">
<div><input type="hidden" name="aid" value="{@aid}" /></div>
<table class="ntable">
	<tr>
		<th>{lang}APPLICATION{/lang} {@alliance}</th>
	</tr>
	<tr>
	<td>
		{@applicationtext}
	</td>
	</tr>
	<tr>
		<td>
			<textarea cols="60" rows="10" class="center" name="application" id="application" onkeyup="maxlength(this,{config}MAX_APPLICATION_TEXT_LENGTH{/config},'counter');">
			</textarea>
			<br />
			{lang}MAXIMUM{/lang} <span id="counter">0</span> / {@maxapplicationtext} {lang}CHARACTERS{/lang}
			<script type="text/javascript">
			//<![CDATA[
				document.getElementById('counter').innerHTML = document.getElementById('application').value.length;
			//]]>
			</script>
		</td>
	</tr>
	<tr>
		<td class="center"><input type="submit" name="apply" value="{lang}COMMIT{/lang}" class="button" /></td>
	</tr>
</table>
</form>
{if[0]}
 Oxsar http://oxsar.ru
 Copyright (c) 2009-2010 UnitPoint <support@unitpoint.ru>
  
{/if}