<script type="text/javascript">
//<![CDATA[
$(document).ready(function() {
	$("#send").click(function() {
		var url = "{const}RELATIVE_URL{/const}";
		if($("#universe").val() != "")
		{
			url = $("#universe").val();
		}
		url += "forgottenpw.php/go:RequestData";
		$.post(url, { username: $("#username").val(), email: $("#email").val() }, function(data) {
			$('#Ajax_Out').html(data);
		});
	});
});
//]]>
</script>

<fieldset>
	<legend>{lang}PASSWORD_FORGOTTEN{/lang}</legend>
	<form method="post" action="{@formaction}" id="lostpw">
	<ul>
		<li>{if[{var=showUniSelection}]}
			<select name="universe" id="universe" class="uni-selection">{@uniSelection}</select>
			<label for="universe">{lang}UNIVERSE{/lang}</label>
			{else}<input type="hidden" name="universe" id="universe" value="" />{/if}
		</li>
		<li><input type="text" name="username" id="username" class="sign-input" /> <label for="username">{lang}USERNAME{/lang}</label></li>
		<li><input type="text" name="email" id="email" class="sign-input" /> <label for="email">{lang}EMAIL{/lang}</label></li>
		<li>
			<input type="button" id="send" value="{lang}REQUEST_DATA{/lang}" class="SignButton" />
		</li>
	</ul>
	</form>
	<div id="Ajax_Out"></div>
</fieldset>
<h3>{lang}LOST_PW_HINT_1{/lang}</h3>
{lang}LOST_PW_HINT_2{/lang}
{if[0]}
 Oxsar https://oxsar-nova.ru
 Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
  
{/if}