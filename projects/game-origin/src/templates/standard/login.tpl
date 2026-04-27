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
<fieldset>
	<legend>{lang}LOGIN{/lang}</legend>
	<form id="signin-form" method="post" action="/login.php">
		<ul>
			<li>{if[{var=showUniSelection}]}
				<select name="universe" id="universe" tabindex="1" class="uni-selection">{@uniSelection}</select>
				<label for="universe">{lang}UNIVERSE{/lang}</label>
			{else}<input type="hidden" name="universe" value="" id="universe" />{/if}
			</li>
			
			<li><label for="username">{lang}USERNAME{/lang}</label><br /></li>
			<li><input type="text" name="username" id="username" class="usernameInput" tabindex="2" accesskey="n" /><br /></li>
			
			<li><label for="password">{lang}PASSWORD{/lang}</label> ({link[FORGOTTEN]}"forgottenpw.php"{/link})<br /></li>
			<li><input type="password" name="password" id="password" class="pwInput" tabindex="3" accesskey="p" /><br /><br /></li>
			
			<li>
				<input type="submit" name="login" value="Sign In" class="SignButton" tabindex="4" accesskey="l" />
				<input type="button" name="signup" value="Sign Up" id="signup-btn" class="SignButton" tabindex="5" />
			</li>
		</ul>
	</form>
	<div>{lang}ACCEPT_AGB{/lang}</div>
</fieldset>
<br />
{if[{var=errorMsg} != ""]}<div class="error">{@errorMsg}</div>{/if}
<h3>{lang}WELCOME{/lang}</h3>
{lang}GAME_DESCRIPTION{/lang}

<!--LiveInternet counter--><script type="text/javascript"><!--
document.write("<a href='http://www.liveinternet.ru/click' "+
"target=_blank><img src='http://counter.yadro.ru/hit?t43.5;r"+
escape(document.referrer)+((typeof(screen)=="undefined")?"":
";s"+screen.width+"*"+screen.height+"*"+(screen.colorDepth?
screen.colorDepth:screen.pixelDepth))+";u"+escape(document.URL)+
";"+Math.random()+
"' alt='' title='LiveInternet' "+
"border='0' width='31' height='31'><\/a>")
//--></script><!--/LiveInternet-->
{if[0]}
 Oxsar https://oxsar-nova.ru
 Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
  
{/if}