<script type="text/javascript">
//<![CDATA[
$(document).ready(function() {
	$('#counter1').text($('#textextern').val().length);
	$('#counter2').text($('#textintern').val().length);
	$('#counter3').text($('#applicationtext').val().length);
});
//]]>
</script>
<form method="post" action="{@formaction}">
<table class="ntable">
	<tr>
		<th colspan="2">{lang}ALLIANCE_MANAGEMENT{/lang}</th>
	</tr>
	<tr>
		<td><input type="text" name="tag" id="tag" value="{@allytag}" maxlength="{config}MAX_CHARS_ALLY_TAG{/config}" /> <label for="tag">{lang}ALLIANCE_TAG{/lang}</label></td>
		<td><input type="submit" name="changetag" value="{lang}PROCEED{/lang}" class="button" /></td>
	</tr>
	<tr>
		<td><input type="text" name="name" id="name" value="{@allyname}" maxlength="{config}MAX_CHARS_ALLY_NAME{/config}" /> <label for="name">{lang}ALLIANCE_NAME{/lang}</label></td>
		<td><input type="submit" name="changename" value="{lang}PROCEED{/lang}" class="button" /></td>
	</tr>
	<tr>
		<td colspan="2" class="center">[ {link[DIPLOMACY]}"game.php/Alliance/Diplomacy"{/link} ] [ {link[RIGHT_MANAGEMENT]}"game.php/Alliance/RightManagement"{/link} ]</td>
	</tr>
</table>
</form>
<form method="post" action="{@formaction}">
<table class="ntable">
	<tr>
		<th><a onclick="javascript:displayAllyText('ExternAllyText');" class="tab active-tab" id="ExternAllyText_Tab">{lang}EXTERN_ALLIANCE_TEXT{/lang}</a><a onclick="javascript:displayAllyText('InternAllyText');" class="tab" id="InternAllyText_Tab">{lang}INTERN_ALLIANCE_TEXT{/lang}</a><a onclick="javascript:displayAllyText('ApplicationAllyText');" class="tab" id="ApplicationAllyText_Tab">{lang}APPLICATION_TEXT{/lang}</a></th>
	</tr>
	<tr>
		<td>
			<div id="ExternAllyText">
				<textarea name="textextern" id="textextern" cols="75" rows="15" onkeyup="maxlength(this,{config}MAX_ALLIANCE_TEXT_LENGTH{/config},'counter1')" class="center">{@textextern}</textarea><br />
				{lang}MAXIMUM{/lang} <span id="counter1">0</span> / {@maxallytext} {lang}CHARACTERS{/lang}
			</div>
			<div id="InternAllyText" style="display: none;">
				<textarea name="textintern" id="textintern" cols="75" rows="15" onkeyup="maxlength(this,{config}MAX_ALLIANCE_TEXT_LENGTH{/config},'counter2')" class="center">{@textintern}</textarea><br />
				{lang}MAXIMUM{/lang} <span id="counter2">0</span> / {@maxallytext} {lang}CHARACTERS{/lang}
			</div>
			<div id="ApplicationAllyText" style="display: none;">
				<textarea name="applicationtext" id="applicationtext" cols="75" rows="15" onkeyup="maxlength(this,{config}MAX_APPLICATION_TEXT_LENGTH{/config},'counter3')" class="center">{@applicationtext}</textarea><br />
				{lang}MAXIMUM{/lang} <span id="counter3">0</span> / {@maxapplicationtext} {lang}CHARACTERS{/lang}
			</div>
			<br />{@externerr}{@internerr}{@apperr}
		</td>
	</tr>
	<tr>
		<td class="center"><input type="reset" value="{lang}RESET_BUTTON{/lang}" class="button" /><input type="submit" name="changetext" value="{lang}PROCEED{/lang}" class="button" /></td>
	</tr>
</table>
<table class="ntable">
	<tr>
		<th colspan="2">{lang}MENU_PREFERENCES{/lang}</th>
	</tr>
	<tr>
		<td><label for="logo">{lang}LOGO{/lang}</label><br />{@logoerr}</td>
		<td><input type="text" name="logo" id="logo" value="{@logo}" maxlength="100" /></td>
	</tr>
	<tr>
		<td><label for="hp">{lang}HOMEPAGE{/lang}</label><br />{@hperr}</td>
		<td><input type="text" name="homepage" id="hp" value="{@homepage}" maxlength="100" /><input type="checkbox" name="showhomepage" value="1" id="showhp"{@showhp} /><label for="showhp">{lang}VISIBLE_TO_ALL{/lang}</label></td>
	</tr>
	<tr>
		<td><label for="memberlist">{lang}MEMBERLIST_SORT{/lang}</label></td>
		<td><select name="memberlistsort" id="memberlist"><option value="1"{@bypoinst}>{lang}BY_POINTS{/lang}</option><option value="2"{@byname}>{lang}BY_NAME{/lang}</option></select><input type="checkbox" name="showmember" value="1" id="showmember"{@showmember} /><label for="showmember">{lang}VISIBLE_TO_ALL{/lang}</label></td>
	</tr>
	<tr>
		<td><label for="apps">{lang}ENABLE_APPLICATIONS{/lang}</label></td>
		<td><input type="checkbox" name="open" id="apps" value="1"{@open} /></td>
	</tr>
	{if[$this->templateVars["founder"] == Core::getUser()->get("userid")]}<tr>
		<td><label for="founder">{lang}FOUNDER_NAME{/lang}</label></td>
		<td><input type="text" name="foundername" id="founder" value="{@foundername}" maxlength="{config}MAX_CHARS_ALLY_NAME{/config}" /></td>
	</tr>{/if}
	<tr>
		<td colspan="2" class="center"><input type="submit" name="changeprefs" value="{lang}PROCEED{/lang}" class="button" /></td>
	</tr>
</table>
{if[$this->templateVars["founder"] != Core::getUser()->get("userid")]}<input type="hidden" name="foundername" value="{@foundername}" />{/if}
</form>
{if[$this->templateVars["founder"] == Core::getUser()->get("userid")]}
{if[$this->templateVars["referfounder"] != ""]}<form method="post" action="{@formaction}">
<table class="ntable">
	<tr>
		<th>{lang}REFER_FOUNDER_STATUS{/lang}</th>
	</tr>
	<tr>
		<td class="center"><select name="userid">{@referfounder}</select><input type="submit" name="referfounder" value="{lang}COMMIT{/lang}" onclick="return confirm('{lang}CONFIRM_PROCEEDING{/lang}');" class="button" /></td>
	</tr>
</table>
</form>{/if}
<form method="post" action="{@formaction}">
<table class="ntable">
	<tr>
		<th>{lang}ABANDON_ALLIANCE{/lang}</th>
	</tr>
	<tr>
		<td class="center"><input type="submit" name="abandonally" value="{lang}ABANDON_ALLIANCE{/lang}" onclick="return confirm('{lang}CONFIRM_ABANDON_ALLIANCE{/lang}');" class="button" /></td>
	</tr>
</table>
</form>
{/if}
{if[0]}
 Oxsar https://oxsar-nova.ru
 Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
  
{/if}