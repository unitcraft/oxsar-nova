
<script type="text/javascript">
//<![CDATA[
$(document).ready(function() {
	// Define vars
	var userInvalid = '{@userCheck}';
	var emailInvalid = '{lang}EMAIL_CHECK{/lang}';
	var passwordInvalid = '{@passwordCheck}';
	var no_agreement = '{lang}REG_NO_AGREEMENT{/lang}';
	
	var valid = 'OK';
	
	var uni = $('#universe');
	var username = $('#username');
	var password = $('#password');
	var email = $('#email');
	var agreement = $('#agreement');
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

	function checkAgreement()
	{
		hideLiveCheck();
		if( agreement.attr('checked') )
		{
			showLiveCheck(valid, false);
			return true;
		}
		showLiveCheck(no_agreement, true);
		return false;
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
	$('#agreement-li').change(checkAgreement);
	$('#agreement-li').focus(checkAgreement);
	$('#agreement-li').blur(checkAgreement);
	
	$('#signup-btn').click(function() {
		var url = "{const}RELATIVE_URL{/const}";
		if(uni.val() != "")
		{
			url = uni.val();
		}
		url += "signup.php/go:CheckUser";
		$.post(
			url,
			{ username: username.val(), password: password.val(), email: email.val(), agreement: agreement.attr('checked') },
			function(data) {
				$('#Ajax_Out').html(data);
				if( $('div', $('#Ajax_Out')).hasClass('success') )
				{
					var hid_form = $('#hidden_login_form');
					$('input[name=username]', hid_form).val(username.val());
					$('input[name=password]', hid_form).val(password.val());
					hid_form.submit();
				}
			}
		);
		return true;
	});
});
//]]>
</script>
<div class="login_panel">
	<form method="post" id="hidden_login_form" action="/game.php" style="display: none;">
		<input type="text" name="login" value="1">
		<input type="text" name="username">
		<input type="text" name="password">
	</form>
	<form method="post" id="reg" action="">
		<fieldset>
		<legend>{lang}REGISTRATION{/lang}</legend>
		<ul>
			{if[{var=showUniSelection}]}
				<li>
					<label for="universe">{lang}UNIVERSE{/lang}</label><br />
				</li>
				<li>
					<select name="universe" id="universe" tabindex="1" class="uni-selection">{@uniSelection}</select>
				</li>
			{else}
				<li>
					<input type="hidden" name="universe" value="" id="universe" />
				</li>
			{/if}

			<li>
				<label for="username">{lang}USERNAME{/lang}</label><br />
			</li>
			<li>
				<input type="text" name="username" id="username" maxlength="{config}MAX_USER_CHARS{/config}" class="sign-input" />
			</li>

			<li>
				<label for="email">{lang}EMAIL{/lang}</label><br />
			</li>
			<li>
				<input type="text" name="email" id="email" maxlength="50" class="sign-input" />
			</li>

			<li>
				<label for="password">{lang}PASSWORD{/lang}</label><br />
			</li>
			<li>
				<input type="password" name="password" id="password" maxlength="{config}MAX_USER_CHARS{/config}" class="sign-input" />
				<br /><br />
			</li>
			

			<li>
				<label for="agreement">{lang}REG_AGREEMENT{/lang}</label><br />
			</li>
			<li id="agreement-li">
				<input type="checkbox" name="agreement" value="1" id="agreement"/>
				<span>{lang=REG_AGREE}</span>
			</li>
			
			<li>
				<input type="button" name="signup" value="{lang=SIGN_UP}" id="signup-btn" class="SignButton" />
			</li>
		</ul>
		</fieldset>
	</form>

	<div id="sign-live-check" class="field_warning" style="display: none;">
	</div>

	<div class="field_warning">
		{@error}
	</div>

	<div id="Ajax_Out">
	</div>

	<h3>{lang}WELCOME{/lang}</h3>
	{lang}GAME_DESCRIPTION{/lang}
	{include}"signup_ext"{/include}
</div>

{if[0]}
 Oxsar https://oxsar-nova.ru
 Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
  
{/if}