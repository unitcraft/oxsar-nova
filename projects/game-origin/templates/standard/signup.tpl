<script type="text/javascript">
//<![CDATA[
$(document).ready(function() {
	// Define vars
	var userInvalid = '{@userCheck}';
	var emailInvalid = '{lang}EMAIL_CHECK{/lang}';
	var passwordInvalid = '{@passwordCheck}';
	var valid = 'OK';
	var uni = $('#universe');
	var username = $('#username');
	var password = $('#password');
	var email = $('#email');
	var validation = $('#sign-live-check');

	function hideLiveCheck()
	{
		validation.hide();
		return true;
	}

	function showLiveCheck(text, error)
	{
		validation.fadeIn(1000);
		validation.html(text);
		if(error)
		{
			validation.removeClass('field_success');
			validation.addClass('field_warning');
		}
		else
		{
			validation.removeClass('field_warning');
			validation.addClass('field_success');
		}
		return;
	}

	function checkUser()
	{
		var len = username.val().length;
		if(len < {config}MIN_USER_CHARS{/config} || len > {config}MAX_USER_CHARS{/config})
		{
			showLiveCheck(userInvalid, true);
			validation.text(userInvalid);
			validation.addClass('field_warning');
			return false;
		}
		else
		{
			showLiveCheck(valid, false);
		}
		return true;
	}

	function checkEmail()
	{
		var regexp = /^(\w+(?:\.\w+)*)@((?:\w+\.)*\w[\w-]{0,66})\.([a-z]{2,6}(?:\.[a-z]{2})?)$/i;
		var result = regexp.test(email.val());
		if(!result)
		{
			showLiveCheck(emailInvalid, true);
			return false;
		}
		else
		{
			showLiveCheck(valid, false);
		}
		return true;
	}

	function checkPassword()
	{
		var len = password.val().length;
		if(len < {config}MIN_USER_CHARS{/config} || len > {config}MAX_USER_CHARS{/config})
		{
			showLiveCheck(passwordInvalid, true);
			return false;
		}
		else
		{
			showLiveCheck(valid, false);
		}
		return true;
	}

	username.keyup(checkUser);
	username.focus(checkUser);
	username.blur(hideLiveCheck);
	password.keyup(checkPassword);
	password.focus(checkPassword);
	password.blur(hideLiveCheck);
	email.keyup(checkEmail);
	email.focus(checkEmail);
	email.blur(hideLiveCheck);
	$('#signup-btn').click(function() {
		var url = "{const}RELATIVE_URL{/const}";
		if(uni.val() != "")
		{
			url = uni.val();
		}
		url += "signup.php/go:CheckUser";
		$.post(url, { username: username.val(), password: password.val(), email: email.val() }, function(data) {
			$('#Ajax_Out').html(data);
		});
		return true;
	});
});
//]]>
</script>

<form method="post" id="reg" action="">
<fieldset>
<legend>{lang}REGISTRATION{/lang}</legend>
<ul>
<li>
	{if[{var=showUniSelection}]}<select name="universe" id="universe" tabindex="1" class="uni-selection">{@uniSelection}</select>
	<label for="universe">{lang}UNIVERSE{/lang}</label>
	{else}<input type="hidden" name="universe" value="" id="universe" />{/if}
</li>
<li>
	<input type="text" name="username" id="username" maxlength="{config}MAX_USER_CHARS{/config}" class="sign-input" /> <label for="username">{lang}USERNAME{/lang}</label>
</li>
<li>
	<input type="text" name="email" id="email" maxlength="50" class="sign-input" /> <label for="email">{lang}EMAIL{/lang}</label>
</li>
<li>
	<input type="password" name="password" id="password" maxlength="{config}MAX_USER_CHARS{/config}" class="sign-input" /> <label for="password">{lang}PASSWORD{/lang}</label>
</li>
<li>
	<input type="button" name="signup" value="Sign Up" id="signup-btn" class="SignButton" />
</li>
</ul>
</fieldset>
</form>
<div id="sign-live-check" class="field_warning" style="display: none;"></div>
<div id="Ajax_Out"></div>
{if[0]}
 Oxsar http://oxsar.ru
 Copyright (c) 2009-2010 UnitPoint <support@unitpoint.ru>
  
{/if}