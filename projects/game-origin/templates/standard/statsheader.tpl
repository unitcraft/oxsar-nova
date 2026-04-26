<form method="post" action="{@formaction}">
<table class="ntable">
	<tr>
		<th>{lang}MENU_RANKING{/lang}</th>
	</tr>
	<tr>
		<td class="center">
			<nobr><label for="mode">{lang}MODE{/lang}</label>
			<select name="mode" id="mode">
			  <option value="player"{@player_sel}>{lang}PLAYER{/lang}</option>
			  <option value="alliance"{@alliance_sel}>{lang}ALLIANCE{/lang}</option>
			  <option value="player_old_vacation"{@player_old_vacation_sel}>{lang}PLAYER_OLD_VACATION{/lang}</option>
              {if[{var=player_observer_enabled}]}
			  <option value="player_observer"{@player_observer_sel}>{lang}PLAYER_OBSERVER{/lang}</option>
              {/if}
			</select></nobr>
			<nobr><label for="type">{lang}TYPE{/lang}</label>
			<select name="type" id="type">
              {if[{var=dm_points_enabled}]}
			  <option value="dm_points"{@dm_points_sel}>{lang}STAT_DM_POINTS{/lang}</option>
              {/if}
              {if[{var=max_points_enabled}]}
			  <option value="max_points"{@max_points_sel}>{lang}STAT_MAX_POINTS{/lang}</option>
              {/if}
			  <option value="points"{@points_sel}>{lang}STAT_POINTS{/lang}</option>
			  <option value="u_points"{@u_points_sel}>{lang}STAT_UNIT_POINTS{/lang}</option>
			  <option value="r_points"{@r_points_sel}>{lang}STAT_RESEARCH_POINTS{/lang}</option>
			  <option value="b_points"{@b_points_sel}>{lang}STAT_BUILD_POINTS{/lang}</option>
			  <option value="e_points"{@e_points_sel}>{lang}BATTLE_EXPERIENCE{/lang}</option>
			  {if[1]}
			  <option value="u_count"{@u_count_sel}>{lang}STAT_UNIT_COUNT{/lang}</option>
			  <option value="r_count"{@r_count_sel}>{lang}STAT_RESEARCH_COUNT{/lang}</option>
			  <option value="b_count"{@b_count_sel}>{lang}STAT_BUILD_COUNT{/lang}</option>
			  {/if}
			</select></nobr>
			<nobr><label for="pos">{lang}RANK{/lang}</label>
			<select name="pos" id="pos">{@rankingSel}</select></nobr>
			<nobr><input type="checkbox" name="avg" id="avg" value="1"{if[$this->templateVars["avg_on"]]} checked="checked"{/if} />
			<label for="avg">{lang}AVERAGE{/lang}</label></nobr>
			<input type="submit" name="go" value="{lang}COMMIT{/lang}" class="button" />
		</td>
	</tr>
</table>
</form>
{if[0]}
 Oxsar http://oxsar.ru
 Copyright (c) 2009-2010 UnitPoint <support@unitpoint.ru>

{/if}