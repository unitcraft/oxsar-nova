<script type="text/javascript" src="{const}RELATIVE_URL{/const}js/lib/wz_tooltip.js"></script>
<script type="text/javascript">
//<![CDATA[
var buildings = new Array();
{foreach[data]}{if[$row["level"] > 0]}
buildings.push({loop}id{/loop});
{/if}{/foreach}
{if[{var}sats{/var} > 0]}
buildings.push(39);
{/if}
//]]>
</script>
<form method="post" action="{@formaction}">
<table class="ntable" class="center">
	<tr>
		<th colspan="6">{lang}RESOURCE_PRODUCTION_FOR_PLANET{/lang} {@planetName}</th>
	</tr>
	<tr>
		<td></td>
		<td align="right"><b>{lang}METAL{/lang}</b></td>
		<td align="right"><b>{lang}SILICON{/lang}</b></td>
		<td align="right"><b>{lang}HYDROGEN{/lang}</b></td>
		<td align="right"><b>{lang}ENERGY{/lang}</b></td>
		<td></td>
	</tr>
	<tr>
		<td><b>{lang}BASIC_PRODUCTION{/lang}</b></td>
		<td align="right"><span class="{if[{var}basicMetal{/var} <= 0]}{else}true{/if}">{@basicMetal}</span></td>
		<td align="right"><span class="{if[{var}basicSilicon{/var} <= 0]}{else}true{/if}">{@basicSilicon}</span></td>
		<td align="right">0</td>
		<td align="right">0</td>
		<td></td>
	</tr>
	{foreach[data]}{if[$row["level"] > 0]}<tr>
		<td><b{if[$row['helptip']]} class="helptip" onmouseover="Tip('{loop=helptip}', FADEIN, 500);" onmouseout="UnTip();"{/if}>{loop}name{/loop} ({loop}level{/loop})</b></td>
		<td align="right">{if[$row["metal"] > 0]}<span class="true">{loop}metal{/loop}</span>{else if[$row["metalCons"] > 0]}<span class="false">{loop}metalCons{/loop}</span>{else}0{/if}</td>
		<td align="right">{if[$row["silicon"] > 0]}<span class="true">{loop}silicon{/loop}</span>{else if[$row["siliconCons"] > 0]}<span class="false">{loop}siliconCons{/loop}</span>{else}0{/if}</td>
		<td align="right">{if[$row["hydrogen"] > 0]}<span class="true">{loop}hydrogen{/loop}</span>{else if[$row["hydrogenCons"] > 0]}<span class="false">{loop}hydrogenCons{/loop}</span>{else}0{/if}</td>
		<td align="right">{if[$row["energy"] > 0]}<span class="true">{loop}energy{/loop}</span>{else if[$row["energyCons"] > 0]}<span class="false">{loop}energyCons{/loop}</span>{else}0{/if}</td>
		<td>{if[!isset($row['allow_factor']) || $row['allow_factor']]}<input type="text" name="{loop}id{/loop}" id="factor_{loop}id{/loop}" value="{loop}factor{/loop}" 
			maxlength="3" size="3" onblur="checkNumberInput(this, 0, 100);" />% 
			<select onchange="setFromSelect('factor_{loop}id{/loop}', this)"><option value="none" class="center">-</option>{@selectProd}</select>{else}&nbsp;{/if}</td>
	</tr>{/if}{/foreach}
	<tr>
		<td class="strongBorderTop"><b>{lang}STORAGE_CAPICITY{/lang}</b></td>
		<td align="right" class="strongBorderTop"><span class="true">{@storageMetal}</span></td>
		<td align="right" class="strongBorderTop"><span class="true">{@storageSilicon}</span></td>
		<td align="right" class="strongBorderTop"><span class="true">{@sotrageHydrogen}</span></td>
		<td align="right" class="strongBorderTop">-</td>
		<td class="strongBorderTop"><input type="button" class="button" value="{lang}SHUT_DOWN{/lang}" onclick="javascript:setProdTo0();" /></td>
	</tr>
	<tr>
		<td><b>{lang}HOURLY_PRODUCTION{/lang}</b></td>
		<td align="right"><span class="{if[{var}totalMetal{/var} <= 0]}false{else}true{/if}">{@totalMetal}</span></td>
		<td align="right"><span class="{if[{var}totalSilicon{/var} <= 0]}false{else}true{/if}">{@totalSilicon}</span></td>
		<td align="right"><span class="{if[{var}totalHydrogen{/var} <= 0]}false{else}true{/if}">{@totalHydrogen}</span></td>
		<td align="right"><span class="{if[{var}totalEnergy{/var} <= 0]}false{else}true{/if}">{@totalEnergy}</span></td>
		<td><input type="button" class="button" value="{lang}START_UP{/lang}" onclick="javascript:setProdTo100();" /></td>
	</tr>
	<tr>
		<td class="strongBorderTop"><b>{lang}DAILY_PRODUCTION{/lang}</b></td>
		<td align="right" class="strongBorderTop"><span class="{if[{var}totalMetal{/var} <= 0]}false{else}true{/if}">{@dailyMetal}</span></td>
		<td align="right" class="strongBorderTop"><span class="{if[{var}totalSilicon{/var} <= 0]}false{else}true{/if}">{@dailySilicon}</span></td>
		<td align="right" class="strongBorderTop"><span class="{if[{var}totalHydrogen{/var} <= 0]}false{else}true{/if}">{@dailyHydrogen}</span></td>
		<td align="right" class="strongBorderTop">-</td>
		<td class="strongBorderTop"><input type="submit" name="update" value="{lang}COMMIT{/lang}" class="button" /></td>
	</tr>
	<tr>
		<td><b>{lang}WEEKLY_PRODUCTION{/lang}</b></td>
		<td align="right"><span class="{if[{var}totalMetal{/var} <= 0]}false{else}true{/if}">{@weeklyMetal}</span></td>
		<td align="right"><span class="{if[{var}totalSilicon{/var} <= 0]}false{else}true{/if}">{@weeklySilicon}</span></td>
		<td align="right"><span class="{if[{var}totalHydrogen{/var} <= 0]}false{else}true{/if}">{@weeklyHydrogen}</span></td>
		<td align="right">-</td>
		<td>&nbsp;</td>
	</tr>
	{if[0 && !isFacebookSkin()]}
	<tr>
		<td><b>{lang}MONTHLY_PRODUCTION{/lang}</b></td>
		<td align="right"><span class="{if[{var}totalMetal{/var} <= 0]}false{else}true{/if}">{@monthlyMetal}</span></td>
		<td align="right"><span class="{if[{var}totalSilicon{/var} <= 0]}false{else}true{/if}">{@monthlySilicon}</span></td>
		<td align="right"><span class="{if[{var}totalHydrogen{/var} <= 0]}false{else}true{/if}">{@monthlyHydrogen}</span></td>
		<td align="right">-</td>
		<td>&nbsp;</td>
	</tr>
	{/if}
</table>
</form>
{if[0]}
 Oxsar https://oxsar-nova.ru
 Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
  
{/if}