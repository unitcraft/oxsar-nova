<script type="text/javascript">
// <![CDATA[
function setActiveStyleSheet(href, type)
{
	var i, a;
	var titleFound = false;
	for(i=0; (a = document.getElementsByTagName('link')[i]); i++)
	{
		if(a.getAttribute('rel').indexOf('style') != -1 && a.getAttribute('href').indexOf('/css/us_' + type + '/') != -1)
		{
			a.disabled = true;
			if(href != '' && a.getAttribute('href').indexOf(href) != -1)
			{
        		a.disabled = false;
		        titleFound = true;
			}
		}
	}
}
setActiveStyleSheet('{@current_bg_style}', 'bg');
setActiveStyleSheet('{@current_table_style}', 'table');
//]]>
</script>
<form method="post" action="{@formaction}">
<table class="ntable">
<colgroup>
            <col width="50%"/>
            <col width="50%"/>
</colgroup>
	{if[ false ]}{*план 37.7.1: Yii::app()->user->activation убран — email-activation legacy не используется (auth через portal+JWT)*}
		<tr>
			<th colspan="2">&nbsp;</th>
		</tr>
		<tr>
			<td><label class="true">{lang}RESEND_ACTIVATION_LABEL{/lang}</label></td>
			<td><input type="submit" name="resent_activation" value='{lang}RESEND_ACTIVATION{/lang}' title="{lang}RESEND_ACTIVATION{/lang}" class="button" /></td>
		</tr>
	{/if}
{if[Core::getUser()->get("umode")]}
	<tr>
		<th colspan="2">{lang}UMODE_PREFERENCES{/lang}</th>
	</tr>
	<tr>
		<td colspan="2" class="center">{if[$this->templateVars["can_disable_umode"]]}<input type="submit" name="disable_umode" value="{lang}DISABLE_UMODE{/lang}" class="button" />{else}{@umode_to}{/if}</td>
	</tr>
    {if[0]}
	<tr>
		<th colspan="2">{lang}ADVANCED_PREFERENCES{/lang}</th>
	</tr>
	<tr>
		<td><label for="del-acc">{lang}DELETE_ACCOUNT{/lang}</label></td>
		<td><input type="checkbox" name="delete" id="del-acc" value="1"{if[Core::getUser()->get("delete") > 0]} checked="checked"{/if} />{if[Core::getUser()->get("delete") > 0]}<span class="notavailable">{@delmessage}</span>{/if}</td>
	</tr>
    {/if}
	<tr>
		<td class="center" colspan="2"><input type="submit" name="update_deletion" value="{lang}COMMIT{/lang}" class="button" /></td>
	</tr>
{else}
	<tr>
		<th colspan="2">{lang}CUSTOM_LOOK{/lang}</th>
	</tr>
	{if[count($this->getLoop("templatePacks")) > 1]}<tr>
		<td><label for="template-pack">{lang}TEMPLATE_PACKAGE{/lang}</label></td>
		<td><select name="templatepackage" id="template-pack"><option value=""></option>{foreach[templatePacks]}<option value="{loop}package{/loop}"{if[$row["package"] == Core::getUser()->get("templatepackage")]} selected="selected"{/if}>{loop}package{/loop}</option>{/foreach}</select></td>
	</tr>
	{/if}
	{if[defined('SN') && !defined('SN_FULLSCREEN')]}
	{else}
	{if[!defined('MOBI')]}
	{if[count($this->getLoop("skins")) > 1]}<tr>
		<td><label for="skin_type">{lang}SKIN_TYPE{/lang}</label></td>
		<td><select name="skin_type" id="skin_type">{foreach[skins]}<option value="{loop}value{/loop}"{if[$row["value"] == Core::getUser()->get("skin_type")]} selected="selected"{/if}>{loop}name{/loop}</option>{/foreach}</select></td>
	</tr>
	{/if}
	{/if}
	{/if}
  {if[count($this->getLoop("user_bg_styles")) > 1]}<tr>
		<td><label for="user_bg_style">{lang}USER_BACKGROUND_STYLE{/lang}</label></td>
		<td><select name="user_bg_style" id="user_bg_style" onChange='setActiveStyleSheet($("#user_bg_style > option:selected").attr("value"), "bg");'><option value="--empty--">-</option>{foreach[user_bg_styles]}<option value="{loop}path{/loop}"{if[$row["path"] == Core::getUser()->get("user_bg_style")]} selected="selected"{/if}>{loop}name{/loop}</option>{/foreach}</select></td>
	</tr>{/if}
  {if[count($this->getLoop("user_table_styles")) > 1]}<tr>
		<td><label for="user_table_style">{lang}USER_TABLE_STYLE{/lang}</label></td>
		<td><select name="user_table_style" id="user_table_style" onChange='setActiveStyleSheet($("#user_table_style > option:selected").attr("value"), "table");'><option value="--empty--">-</option>{foreach[user_table_styles]}<option value="{loop}path{/loop}"{if[$row["path"] == Core::getUser()->get("user_table_style")]} selected="selected"{/if}>{loop}name{/loop}</option>{/foreach}</select></td>
	</tr>{/if}
	{if[count($this->getLoop("imagePacks")) > 1]}<tr>
		<td><label for="image-pack">{lang}IMAGE_PACKAGE{/lang}</label></td>
		<td><select name="imagepackage" id="image-pack">{foreach[imagePacks]}<option value="{loop}dir{/loop}"{if[$row["dir"] == Core::getUser()->get("imagepackage")]} selected="selected"{/if}>{loop}name{/loop}</option>{/foreach}</select></td>
	</tr>{/if}
	<!-- <tr>
		<td><label for="theme">{lang}THEME{/lang}</label><br /><span class="small">{lang}THEME_HINT{/lang}</span></td>
		<td><input type="text" name="theme" id="theme" maxlength="{config}MAX_INPUT_LENGTH{/config}" value="{user}theme{/user}" /></td>
	</tr> -->
	<tr>
		<td><label>{lang}SHOW_UNAVAILABLE_UNITS{/lang}</label></td>
		<td>
		  <input type="checkbox" name="show_all_constructions" id="show_all_constructions" {if[Core::getUser()->get("show_all_constructions")]}checked="checked"{/if}/><label for="show_all_constructions">{lang}MENU_CONSTRUCTIONS{/lang}</label><br/>
		  <input type="checkbox" name="show_all_research" id="show_all_research" {if[Core::getUser()->get("show_all_research")]}checked="checked"{/if}/><label for="show_all_research">{lang}MENU_RESEARCH{/lang}</label><br/>
		  <input type="checkbox" name="show_all_shipyard" id="show_all_shipyard" {if[Core::getUser()->get("show_all_shipyard")]}checked="checked"{/if}/><label for="show_all_shipyard">{lang}MENU_SHIPYARD{/lang}</label><br/>
		  <input type="checkbox" name="show_all_defense" id="show_all_defense" {if[Core::getUser()->get("show_all_defense")]}checked="checked"{/if}/><label for="show_all_defense">{lang}MENU_DEFENSE{/lang}</label><br/>
		</td>
	</tr>
	<tr>
		<td><label for="planet-order">{lang}PLANET_ORDER{/lang}</label></td>
		<td><select name="planetorder" id="planet-order"><option value="1"{if[Core::getUser()->get("planetorder") == 1]} selected="selected"{/if}>{lang}EVOLUTION{/lang}</option><option value="2"{if[Core::getUser()->get("planetorder") == 2]} selected="selected"{/if}>{lang}ALPHABETICAL{/lang}</option><option value="3"{if[Core::getUser()->get("planetorder") == 3]} selected="selected"{/if}>{lang}COORDINATES{/lang}</option></select></td>
	</tr>
	<tr>
		<th colspan="2">{lang}ADVANCED_PREFERENCES{/lang}</th>
	</tr>
	{if[Core::getDB()->num_rows($this->getLoop("langs")) > 1]}<tr>
		<td><label for="language">{lang}LANGUAGE{/lang}</label></td>
		<td><select name="language" id="language">{while[langs]}<option value="{loop}languageid{/loop}"{if[Core::getUser()->get("languageid") == $row["languageid"]]} selected="selected"{/if}>{loop}title{/loop}</option>{/while}</select></td>
	</tr>{/if}
	<tr>
		<td><label for="num-esps">{lang}ESPIONAGE_PROBES{/lang}</label></td>
		<td><input type="text" name="esps" id="num-esps" maxlength="2" value="{user}esps{/user}" /></td>
	</tr>
	{if[defined('SN') && !defined('SN_EXT')]}
	{else}
	{if[defined('IPCHECK') && IPCHECK]}
	<tr>
		<td><label for="ip-check">{lang}IP_CHECK_ACTIVATED{/lang}</label></td>
		<td><input type="checkbox" name="ipcheck" id="ip-check" value="1"{if[Core::getUser()->get("ipcheck")]} checked="checked"{/if} /></td>
	</tr>
	{/if}
	{/if}
	<tr>
		<td><label for="vacation">{lang}VACATION_MODE{/lang}</label></td>
		<td><input type="checkbox" name="umode" id="vacation" value="1" onclick="return confirm('{lang}VACATION_WARNING{/lang}');" /></td>
	</tr>
	{if[defined('SN') && !defined('SN_EXT')]}
	{else}
	<tr>
		<th colspan="2">&nbsp;</th>
	</tr>
	<tr>
		<td><label for="del-acc"><span class="false">{lang}DELETE_ACCOUNT{/lang}</span></label></td>
		<td style="text-align: right"><input type="checkbox" name="delete" id="del-acc" value="1"{if[Core::getUser()->get("delete") > 0]} checked="checked"{/if} />{if[Core::getUser()->get("delete") > 0]}<span class="notavailable">{@delmessage}</span>{/if}</td>
	</tr>
	{/if}
	<tr>
		<th colspan="2">{lang}USER_DATA{/lang}</th>
	</tr>
	<tr>
		<td><label for="username">{lang}USERNAME{/lang}</label></td>
		<td><input type="text" name="username" id="username" maxlength="{config}MAX_USER_CHARS{/config}" value="{user}username{/user}" /></td>
	</tr>
	{if[defined('SN') && !defined('SN_EXT')]}
	{else}
		<tr>
			<td><label for="email">{lang}EMAIL{/lang}</label></td>
			<td><input type="text" name="email" id="email" maxlength="50"
                value="{if[checkEmail(Core::getUser()->get("email"))]}{user}email{/user}{/if}" /></td>
		</tr>
		{if[Core::getUser()->get("email") != Core::getUser()->get("temp_email") && checkEmail(Core::getUser()->get("temp_email"))]}
		<tr>
			<td><label for="email">{lang}EMAIL_CONFIRM{/lang}</label></td>
			<td><?php echo Core::getUser()->get("temp_email"); ?></td>
		</tr>
		{/if}
        {if[checkEmail(Core::getUser()->get("email"))]}
		<tr>
			<td><label for="new-pw">{lang}NEW_PASSWORD{/lang}</label></td>
			<td><input type="password" name="password" id="new-pw" maxlength="{config}MAX_USER_CHARS{/config}" /></td>
		</tr>
        {/if}
	{/if}
	<tr>
		<td class="center" colspan="2"><input type="submit" name="saveuserdata" value="{lang}COMMIT{/lang}" class="button" /></td>
	</tr>
{/if}
</table>
</form>
{if[0]}
 Oxsar http://oxsar.ru
 Copyright (c) 2009-2010 UnitPoint <support@unitpoint.ru>

{/if}