<form method="post" action="{@formaction}">
<table class="ntable">
	<tr>
		<th colspan="2">{lang}MODERATE_USER{/lang}</th>
	</tr>
	<tr>
		<th colspan="2">{lang}USER_DATA{/lang}<input type="hidden" name="userid" value="{@userid}" /></th>
	</tr>
	<tr>
		<td><label for="username">{lang}USERNAME{/lang}</label></td>
		<td><input type="text" name="username" id="username" value="{@username}" maxlength="{config}MAX_USER_CHARS{/config}" /></td>
	</tr>
	<tr>
		<td><label for="email">{lang}EMAIL{/lang}</label></td>
		<td><input type="text" name="email" id="email" value="{@email}" maxlength="{config}MAX_INPUT_LENGTH{/config}" /></td>
	</tr>
	<tr>
		<td><label for="temp-email">{lang}TEMP_EMAIL{/lang}</label></td>
		<td><a href="mailto:{@temp_email}" id="temp-email">{@temp_email}</a></td>
	</tr>
	<tr>
		<td><label for="pw">{lang}NEW_PASSWORD{/lang}</label></td>
		<td><input type="text" name="password" id="pw" maxlength="{config}MAX_USER_CHARS{/config}" /></td>
	</tr>
	<tr>
		<td><label>{lang}LAST_ACTIVITY{/lang}</label></td>
		<td>{@last}</td>
	</tr>
	<tr>
		<td><label>{lang}REGISTRATION_DATE{/lang}</label></td>
		<td>{@regtime}</td>
	</tr>
	{if[{var}tag{/var} != ""]}
	<tr>
		<td><label>{lang}ALLIANCE{/lang}</label></td>
		<td><strong>{@tag}</strong></td>
	</tr>
	{/if}
	
	<tr>
		<td>{lang}STAT_POINTS{/lang}</td>
		<td>{@points}</td>
	</tr>
	<tr>
		<td>{lang}STAT_UNIT_POINTS{/lang}</td>
		<td>{@u_points}</td>
	</tr>
	<tr>
		<td>{lang}STAT_RESEARCH_POINTS{/lang}</td>
		<td>{@r_points}</td>
	</tr>
	<tr>
		<td>{lang}STAT_BUILD_POINTS{/lang}</td>
		<td>{@b_points}</td>
	</tr>
	<tr>
		<td>{lang}BATTLE_EXPERIENCE{/lang}</td>
		<td>{@points}</td>
	</tr>
	<tr>
		<td>{lang}STAT_UNIT_COUNTS{/lang}</td>
		<td>{@u_count}</td>
	</tr>
	<tr>
		<td>{lang}STAT_RESEARCH_COUNTS{/lang}</td>
		<td>{@r_count}</td>
	</tr>
	<tr>
		<td>{lang}STAT_BUILD_COUNTS{/lang}</td>
		<td>{@b_count}</td>
	</tr>

	<tr>
		<th colspan="2">{lang}ADVANCED_PREFERENCES{/lang}</th>
	</tr>
	{if[Core::getDB()->num_rows($this->getLoop("langs")) > 1]}<tr>
		<td><label for="langauage">{lang}LANGUAGE{/lang}</label></td>
		<td><select name="langauageid" id="langauage">{while[langs]}<option value="{loop}languageid{/loop}"{if[{var}languageid{/var} == $row["languageid"]]} selected="selected"{/if}>{loop}title{/loop}</option>{/while}</select></td>
	</tr>{/if}
	{perm[CAN_EDIT_USER]}
	<tr>
		<td><label for="usergroup">{lang}USER_GROUP{/lang}</label></td>
		<td><select name="usergroupid" id="usergroup"><option value="3">User</option><option value="2"{if[{var}isAdmin{/var}]} selected="selected"{/if}>Administrator</option><option value="4"{if[{var}isMod{/var}]}selected="selected"{/if}>Moderator</option></select></td>
	</tr>
	{/perm}
	<tr>
		<td><label for="del-acc">{lang}DELETE_ACCOUNT{/lang}</label></td>
		<td><input type="checkbox" name="delete" id="del-acc" value="1"{if[{var}deletion{/var} > 0]} checked="checked"{/if} />{if[{var}deletion{/var} > 0]}<span class="notavailable">{@delmessage}</span>{/if}</td>
	</tr>
	<tr>
		<td><label for="ip-check">{lang}IP_CHECK_ACTIVATED{/lang}</label></td>
		<td><input type="checkbox" name="ipcheck" id="ip-check" value="1"{if[{var}ipcheck{/var}]} checked="checked"{/if} /></td>
	</tr>
	<tr>
		<td><label for="activation">{lang}ACTIVATED{/lang}</label></td>
		<td><input type="checkbox" name="activation" id="activation" value="1"{if[{var}activation{/var} == ""]} checked="checked"{/if} /></td>
	</tr>
	<tr>
		<td><label for="umode">{lang}VACATION_MODE{/lang}</label></td>
		<td><input type="checkbox" name="umode" id="umode" value="1"{if[{var}vacation{/var}]} checked="checked"{/if} /></td>
	</tr>
	<tr>
		<th colspan="2">{lang}CUSTOM_LOOK{/lang}</th>
	</tr>
	<tr>
		<td><label for="templatepackage">{lang}TEMPLATE_PACKAGE{/lang}</label></td>
		<td><input type="text" name="templatepackage" id="templatepackage" maxlength="{config}MAX_INPUT_LENGTH{/config}" /></td>
	</tr>
	<tr>
		<td><label for="theme">{lang}THEME{/lang}</label></td>
		<td><input type="text" name="theme" id="theme" maxlength="{config}MAX_INPUT_LENGTH{/config}" /></td>
	</tr>
	<tr>
		<td class="center" colspan="2"><input type="submit" name="proceed" value="{lang}PROCEED{/lang}" class="button" /></td>
	</tr>
	<tr>
		<th colspan="2">{lang}BAN_USER{/lang}</th>
	</tr>
	<tr>
		<td><label for="ban-for">{lang}BAN_FOR{/lang}</label></td>
		<td>
			<input type="text" name="ban" id="ban-for" value="1" maxlength="2" size="4" />
			<select name="timeend"><option value="60">{lang}MINUTES{/lang}</option><option value="3600">{lang}HOURS{/lang}</option><option value="86400">{lang}DAYS{/lang}</option><option value="604800">{lang}WEEKS{/lang}</option><option value="2419200">{lang}MONTHS{/lang}</option><option value="3153600">{lang}YEARS{/lang}</option></select>
		</td>
	</tr>
	<tr>
		<td><label for="reason">{lang}REASON{/lang}</label></td>
		<td>
			<input type="text" name="reason" id="reason" maxlength="{config}MAX_INPUT_LENGTH{/config}" /><br />
			<input type="checkbox" name="b_umode" id="b_umode" value="1" /><label for="b_umode">{lang}VACATION_MODE{/lang}</label>
		</td>
	</tr>
	<tr>
		<td class="center" colspan="2"><input type="submit" name="proceedban" value="{lang}BAN{/lang}" class="button" /></td>
	</tr>

  <tr>
    <th colspan="2">{lang}RO_USER{/lang}</th>
  </tr>
  <tr>
    <td>
      <label for="ro-for">{lang}RO_FOR{/lang}</label>
    </td>
    <td>
      <input type="text" name="ro" id="ro-for" value="1" maxlength="2" size="4" />
      <select name="timeendro">
        <option value="60">{lang}MINUTES{/lang}</option>
        <option value="3600">{lang}HOURS{/lang}</option>
        <option value="86400">{lang}DAYS{/lang}</option>
        <option value="604800">{lang}WEEKS{/lang}</option>
        <option value="2419200">{lang}MONTHS{/lang}</option>
        <option value="3153600">{lang}YEARS{/lang}</option>
      </select>
    </td>
  </tr>
  <tr>
    <td>
      <label for="ro-reason">{lang}REASON{/lang}</label>
    </td>
    <td>
      <input type="text" name="reasonro" id="reason" maxlength="{config}MAX_INPUT_LENGTH{/config}" />
      <br />
    </td>
  </tr>
  <tr>
    <td class="center" colspan="2">
      <input type="submit" name="proceedro" value="{lang}RO{/lang}" class="button" />
    </td>
  </tr>
  
  {if[{var}eBans{/var} > 0]}<tr>
		<th colspan="2">{lang}EXISTING_BANS{/lang}</th>
	</tr>
	<tr>
		<td colspan="2">
			{foreach[bans]}<div>
				{loop}reason{/loop} - {loop}to{/loop} - {loop}annul{/loop}
			</div>{/foreach}
		</td>
	</tr>{/if}
  
	{if[{var}eROs{/var} > 0]}<tr>
		<th colspan="2">{lang}EXISTING_ROS{/lang}</th>
	</tr>
	<tr>
		<td colspan="2">
			{foreach[ros]}<div>
				{loop}reason{/loop} - {loop}to{/loop} - {loop}annul{/loop}
			</div>{/foreach}
		</td>
	</tr>{/if}
  
</table>
</form>
{if[0]}
 Oxsar https://oxsar-nova.ru
 Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
  
{/if}