<script type="text/javascript" src="{const}RELATIVE_URL{/const}js/AjaxRequest.js"></script>
<script type="text/javascript">
//<![CDATA[
$(document).ready(function() {
	$("#send").click(function() { 
		var url = "{const}RELATIVE_URL{/const}";
		if($("#universe").val() != "")
		{
			url = $("#universe").val();
		}
		url += "forgottenpw.php/go:ChangePW";
		$.post(url, { key: $("#key").val(), userid: $("#userid").val(), password: $("#password").val() }, function(data) {
			$('#Ajax_Out').html(data);
		});
	});
});
//]]>
</script>

<fieldset>
	<form method="post" action="">
	
	<input type="hidden" name="key" value="{@seckey}" id="key" />
	<input type="hidden" name="userid" value="{@userid}" id="userid" />
	<legend>{lang}CHANGE_PASSWORD{/lang}</legend>
	<ul>
		<li>{if[{var=showUniSelection}]}
			<select name="universe" id="universe" class="uni-selection">{@uniSelection}</select>
			<label for="universe">{lang}UNIVERSE{/lang}</label>
			{else}<input type="hidden" name="universe" id="universe" value="" />{/if}
		</li>
		<li><input type="text" name="password" id="password" /> <label for="password">{lang}PASSWORD{/lang}</label></li>
		<li>
			<input type="button" id="send" value="{lang}COMMIT{/lang}" class="SignButton" />
		</li>
	</ul>
	</form>
	<div id="Ajax_Out"></div>
</fieldset>
{if[0]}
 Oxsar http://oxsar.ru
 Copyright (c) 2009-2010 UnitPoint <support@unitpoint.ru>
  
{/if}