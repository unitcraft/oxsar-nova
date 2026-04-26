<form method="post" action="{@formaction}">
<input type="hidden" id="editplanetid" name="editplanetid" value="<?= NS::getPlanet()->getPlanetId() ?>" />
<table class="ntable">
	<thead><tr>
		<th colspan="2">{lang}PLANET_OPTIONS{/lang}</th>
	</tr></thead>
	<tfoot><tr>
		<td colspan="2" class="center"><input type="submit" name="changeplanetoptions" value="{lang}COMMIT{/lang}" class="button" /></td>
	</tr></tfoot>
	<tbody>
	
	{if[!NS::getPlanet()->getData("ismoon") && NS::getPlanet()->getPlanetId() != NS::getUser()->get("hp") && !NS::getPlanet()->getData("destroy_eventid")]}
	<tr>
		<td>{lang}CUR_HOME_PLANET{/lang}</td>
		<td>
		  {@homeplanet}
		</td>
	</tr>
	<tr>
		<td>{lang}SET_NEW_HOME_PLANET{/lang}</td>
		<td>
		  {@curplanet}
		  <input type="checkbox" name="ishomeplanet" value="1" />
		</td>
	</tr>
	{else}
	<tr>
	  <td>{if[NS::getPlanet()->getPlanetId() == NS::getUser()->get("hp")]}{lang}CUR_HOME_PLANET{/lang}{else}{lang}POSITION{/lang}{/if}</td>
		<td>{@curplanet}</td>
	</tr>
	{/if}

	<tr>
		<td>{lang}NEW_PLANET_NAME{/lang}</td>
		<td><input type="text" name="planetname" value="{@planetName}" maxlength="{config}MAX_USER_CHARS{/config}" /></td>
	</tr>
	<tr>
		<th colspan="2">&nbsp;</th>
	</tr>
	<tr>
		<td><span class="false2"><b>{lang}ABANDON_PLANET{/lang}</span></b></td>
		<td>
			{if[0 && !defined('SN')]}
			<input type="checkbox" name="abandon" value="1" onclick="javascript:showHideId('password');" />
			<input type="password" name="password" id="password" maxlength="{config}MAX_USER_CHARS{/config}" style="display: none;" class="pwInput" />
			{else}
			<input type="checkbox" name="abandon" value="1" onclick="javascript:showHideId('leave_panel');"/>
			<div id="leave_panel" style="display: none;">
				<i>Введите</i> <b>{lang}LEAVE{/lang}<?= NS::getPlanet()->getPlanetId() ?></b> <br />
				<input type="text" id="leave" name="leave" class="pwInput" />
			</div>
			{/if}			
		</td>
	</tr></tbody>
</table>
</form>
{if[0]}
 Oxsar http://oxsar.ru
 Copyright (c) 2009-2010 UnitPoint <support@unitpoint.ru>
  
{/if}