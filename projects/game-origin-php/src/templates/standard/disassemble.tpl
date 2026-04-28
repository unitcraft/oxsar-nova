<table class="ntable">
  <tr>
    <th colspan="4">
      <span style="float:right">{lang=LEVEL} {@construction_level}</span>
      {@construction_name}
    </th>
  </tr>
  <tr>
    <td colspan="4">
      <div style="float:left; padding-right:5px">{@construction_image}</div>
      <div style="display:table">{@construction_description}
        <div style="padding: 5px 0">
          <div class='rep_destroyed_back_div' style="clear:both"><div class='rep_alive_over_div' style='width: {@construction_free_percent}%' /></div>
        </div>
        <table class="table_no_background" cellspacing="0" cellpadding="0" border="0">
          <tr>
            <td>{lang=REPAIR_STORAGE}</td>
            <td>{@construction_storage}</td>
          </tr>
          <tr>
            <td>{lang=REPAIR_USED}</td>
            <td>{@construction_used}</td>
          </tr>
          <tr>
            <td>{lang=REPAIR_FREE}</td>
            <td>{@construction_free}</td>
          </tr>
        </table>
      </div>
    </td>
  </tr>
  {foreach[events]}
  <tr>
    <td width="1px">
      {loop}number{/loop}.
    </td>
    <td {if[ !$row["vip_link"] && !(!isMobileSkin() && $row['event_pb_value']) ]}colspan="2"{/if}>
      {loop}name{/loop}:&nbsp;{loop}quantity{/loop}
    </td>
    {if[ $row["vip_link"] ]}
    <td width="130px">
      {loop}vip_link{/loop}
    </td>
	{else if[ !isMobileSkin() && $row['event_pb_value'] ]}
    <td width="130px">
		<div id="evpb{loop=eventid}" style="height:14px"></div>
    </td>
    {/if}
    <td width="100px">
      {loop}cancel_link{/loop}
    </td>
  </tr>
  {/foreach}
</table>

{if[ !isMobileSkin() ]}
<script type="text/javascript">
$(function(){
	{foreach[events]}
		{if[ !$row["vip_link"] && $row['event_pb_value'] ]}
			$("#evpb{loop=eventid}").progressbar({
				value: {loop=event_pb_value}
			});
			setInterval(function(){
				var value = $("#evpb{loop=eventid}").progressbar("option", "value");
				if(value < 100)
				{
					$("#evpb{loop=eventid}").progressbar("option", "value", value+1);
				}
			}, {loop=event_percent_timeout});
		{/if}
	{/foreach}
});
</script>
{/if}

<table class="ntable">
  <!-- <tr>
    <th colspan="3">{lang=REPAIR_NEEDED_UNITS}</th>
  </tr> -->
  <tr>
    <th colspan="2">
      <table width="100%" class="table_no_background" cellspacing="0" cellpadding="0" border="0">
        <tr>
         <form method="post" action="{@formaction}" name="tform" id="tform">
          <td style="padding-left:0;padding-right:0">
            <b>{lang}IMAGE_PACKAGE{/lang}</b>
            <select name="image_package" id="image_package" onchange="document.forms['tform'].submit()">
              {foreach[imagePacks]}<option value="{loop}dir{/loop}"{if[$row["dir"] == Core::getUser()->get("imagepackage")]} selected="selected"{/if}>{loop}name{/loop}</option>{/foreach}
            </select>
          </td>
         </form>
        </tr>
      </table>
    </th>
    <th style="text-align:center">{lang}QUANTITY{/lang}</th>
  </tr>

  <form method="post" action="{@formaction}" style="padding:0; margin:0">
  {foreach[items]}
  {if[$row["fleet_started"]]}
  <tr>
    <th colspan="3">{lang=FLEET}</th>
  </tr>
  {/if}
  {if[$row["defense_started"]]}
  <tr>
    <th colspan="3">{lang=DEFENSE}</th>
  </tr>
  {/if}
  <tr>
    <td width="1px">{loop}image{/loop}</td>
    <td valign="top">
      <div style="width:100%">
        <span style="float:right">
          <table class="table_no_background" cellspacing="0" cellpadding="0" border="0">
            <tr>
            {if[$row['dock_capacity'] > 0]}
              <td align="right">{lang=REPAIR_UNITS_CAPACITY}:</td>
              <td>{loop=dock_capacity}</td>
            {else}
              <td colspan="2">{lang=REPAIR_UNIT_TOO_LARGE}</td>
            {/if}
            </tr>
            {if[$row['dock_capacity'] > 0]}
              <tr>
              {if[$row['max_dock_units'] > 0]}
                <td align="right">{lang=MAX_REPAIR_QUANTITIES}:</td>
                <td>{loop=max_dock_units}</td>
              {else}
                <td colspan="2">{lang=REPAIR_ZERO_QUANTITIES}</td>
              {/if}
              </tr>
            {/if}
          </table>
        </span>
        {loop}name{/loop}
      </div>
      <div>
        <table class="table_no_background" cellspacing="0" cellpadding="0" border="0">
          <tr>
            <td>{lang=UNIT_FIELDS}</td>
            <td class='{if[$row['no_free_repair_fields']]}notavailable{else}true{/if}'>{loop=unit_fields}</td>
            <td>{if[$row["no_free_repair_fields"]]}({loop=no_free_repair_fields}){/if}</td>
          </tr>
          {if[$row["can_build"]]}
            {if[$row["metal_required"]]}
            <tr>
              <td>{lang=METAL}</td>
              <td class='{if[$row["metal_notavailable"]]}notavailable{else}true{/if}' title="{lang=REQUIRES}">{loop=metal_required}</td>
              <td>{if[$row["metal_notavailable"]]}({loop=metal_notavailable}){/if}</td>
              <td>+{loop=metal_earned}</td>
            </tr>
            {/if}
            {if[$row["silicon_required"]]}
            <tr>
              <td>{lang=SILICON}</td>
              <td class='{if[$row["silicon_notavailable"]]}notavailable{else}true{/if}' title="{lang=REQUIRES}">{loop=silicon_required}</td>
              <td>{if[$row["silicon_notavailable"]]}({loop=silicon_notavailable}){/if}</td>
              <td>+{loop=silicon_earned}</td>
            </tr>
            {/if}
            {if[$row["hydrogen_required"]]}
            <tr title="{lang=REQUIRES}">
              <td>{lang=HYDROGEN}</td>
              <td class='{if[$row["hydrogen_notavailable"]]}notavailable{else}true{/if}'>{loop=hydrogen_required}</td>
              <td>{if[$row["hydrogen_notavailable"]]}({loop=hydrogen_notavailable}){/if}</td>
            </tr>
            {/if}
            {if[$row["energy_required"]]}
            <tr title="{lang=REQUIRES}">
              <td>{lang=ENERGY}</td>
              <td class='{if[$row["energy_notavailable"]]}notavailable{else}true{/if}'>{loop=energy_required}</td>
              <td>{if[$row["energy_notavailable"]]}({loop=energy_notavailable}){/if}</td>
            </tr>
            {/if}
            <tr title="{lang=REQUIRES}">
              <td>{lang=REQUIRE_TIME}</td>
              <td colspan="3">{loop=productiontime}</td>
            </tr>
          {/if}
        </table>
      </div>
    </td>
    <td width="100px" align="center" valign="top">
      {loop}quantity{/loop}
      <br /><br />
      {if[$row["can_build"]]}
      {loop}construct{/loop}</td>
      {else}
      {if[0]}{lang=DISASSEMBLE_NOT_AVAILABLE}{/if}
      {/if}
  </tr>
  {/foreach}
  <tr>
    <td colspan="3" align="center">
      <input type="submit" name="sendmission" value="{lang}DISASSEMBLE{/lang}" class="button" />
    </td>
  </tr>
  </form>
</table>
{if[0]}
 Oxsar https://oxsar-nova.ru
 Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.

{/if}