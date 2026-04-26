<div class="login_panel">
<fieldset>
	<legend>{lang}LOGIN{/lang}</legend>
	<form id="signin-form" method="post" action="/login.php">
	<ul>
		{if[{var=showUniSelection}]}
		<li>
			<label for="universe">{lang}UNIVERSE{/lang}</label>
			<br />
		</li>
		<li>
			<select name="universe" id="universe" tabindex="1" class="uni-selection">{@uniSelection}</select>
		</li>
		{else}
		<li>
		  <input type="hidden" name="universe" value="" id="universe" />
		</li>
		{/if}
		
		<li><label for="username">{lang}USERNAME{/lang}</label><br /></li>
		<li><input type="text" name="username" id="username" class="usernameInput" tabindex="2" accesskey="n" /><br /></li>
		
		<li><label for="password">{lang}PASSWORD{/lang}</label> ({link[FORGOTTEN]}"forgottenpw.php"{/link})<br /></li>
		<li><input type="password" name="password" id="password" class="pwInput" tabindex="3" accesskey="p" /><br /><br /></li>
		
		<li>
			<input type="submit" name="login" value="{lang=SIGN_IN}" class="SignButton" tabindex="4" accesskey="l" />
		</li>
	</ul>
	</form>
	<!-- <div>{lang}ACCEPT_AGB{/lang}</div> -->
</fieldset>
{if[{var=errorMsg} != ""]}<div class="error">{@errorMsg}</div>{/if}
<h3>{lang}WELCOME{/lang}</h3>
{lang}GAME_DESCRIPTION{/lang}
<!-- DO NOT MODIFY "POWERED BY" DIV, feel free to change "poweredby" css class -->
{if[0]}<div class="poweredby"><a href="http://netassault.ru" target="_blank">Powered by NetAssault</a></div>{/if}

<script type="text/javascript">
//<![CDATA[
$(document).ready(function() {
	var signIn = $("#signin-form");
	signIn.submit(function() {
		var action = "{const=RELATIVE_URL}";
		if($("#universe").val() != "")
		{
			action = $("#universe").val();
		}
		signIn.attr("action", action);
		return true;
	});
	$("#signup-btn").click(function() {
		location.href = "{const=RELATIVE_URL}signup.php";
	});
});
//]]>
</script>

</div>
{if[0]}
 Oxsar http://oxsar.ru
 Copyright (c) 2009-2010 UnitPoint <support@unitpoint.ru>
  
{/if}