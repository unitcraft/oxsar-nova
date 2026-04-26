<script type="text/javascript">
lang_CONFIRM_TITLE_WARNING = "{lang=CONFIRM_TITLE_WARNING}";
lang_CONFIRM_OK = "{lang=CONFIRM_OK}";
lang_CONFIRM_CANCEL = "{lang=CONFIRM_CANCEL}";
</script>

{if[count($this->getLoop("events")) > 0]}
<table class="ntable">
  <tr>
    <th colspan="4">{lang}OUTSTANDING_MISSIONS{/lang}</th>
  </tr>
  {foreach[events]}
  <tr>
    <td width="1px">
      {loop}number{/loop}.
    </td>
    <td {if[ !$row["vip_link"] && !(!isMobileSkin() && $row['event_pb_value']) ]}colspan="2"{/if}>
      {loop}name{/loop}&nbsp;{loop}level{/loop}
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
{/if}

<table class="ntable">
    {if[{var=info_id} > 0]}
    {else}
      <tr>
        <th colspan="3">{lang}CONSTRUCTIONS{/lang}</th>
      </tr>
      <tr>
        <th colspan="2">{include}"select_style_form"{/include}</th>
        <th style="text-align:center">&nbsp;</th>
      </tr>
    {/if}

  {foreach[constructions]}
    {if[{var=info_id} > 0]}
      <tr>
        <th colspan="3">
            {if[$row['level'] > 0 || $row['added_level']]}<span style="float:right">{lang=LEVEL} {loop}level{/loop}{if[$row['added_level']>0]} <span class='true'>(+{loop=added_level})</span>{else if[$row['added_level']<0]} <span class='false'>({loop=added_level})</span>{/if}</span>{/if}
            {loop}name{/loop}
        </th>
      </tr>
    {/if}

  <tr>
    <td width="1px" style="vertical-align: top;">{loop}image{/loop}</td>
    <td style="vertical-align: top;">
        {if[{var=info_id} > 0]}
            {loop}description{/loop}
        {else}
          <div style="width:100%">
            {if[$row['level'] > 0 || $row['added_level']]}<span style="float:right">{lang=LEVEL} {loop}level{/loop}{if[$row['added_level']>0]} <span class='true'>(+{loop=added_level})</span>{else if[$row['added_level']<0]} <span class='false'>({loop=added_level})</span>{/if}</span>{/if}
            {loop}name{/loop}
          </div>
          <div style="clear:both; font-size:smaller">{loop}description{/loop} {perm[CAN_EDIT_CONSTRUCTIONS]}{loop}edit{/loop}{/perm}</div>
        {/if}

      <div>
        {if[$row["can_build"]]}
          {include}"required_res_table"{/include}
        {else}
          <span class="normal">{lang}REQUIRED_CONSTRUCTIONS{/lang}</span>
          <br />{loop=required_constructions}
        {/if}
      </div>
    </td>
    {if[$row["can_build"]]}
    <td width="100px" align="center" id='build_construction_<?php echo $key;?>'>
      {loop}upgrade{/loop}
    </td>
    {else}
    <td width="100px"></td>
    {/if}
  </tr>
  {/foreach}

  {if[{var=info_id} > 0]}
    {if[{var=ext_chart_type} && {var=ext_chart_type} != "error"]}
    <tr>
        <td colspan="3" align="center">{include}{var=ext_chart_type}{/include}</td>
    </tr>
    {/if}

    {if[{var}ext_demolish{/var}]}
        <tr>
            <th colspan="3" align="center">{lang}DEMOLISH{/lang} {@ext_name} {@ext_level}</th>
        </tr>
        <tr>
            <td colspan="3" align="center">{lang}REQUIRES{/lang} {@ext_demolish_metal} {@ext_demolish_silicon} {@ext_demolish_hydrogen}<br />{lang}PRODUCTION_TIME{/lang} {@ext_demolish_time}</td>
        </tr>
        {if[{var}ext_demolish_now{/var}]}
        <tr>
            <td colspan="3" align="center">{@ext_demolish_now}</td>
        </tr>
        {/if}
    {/if}

    {if[{var}ext_pack_building{/var}]}
    <tr>
        <td colspan="3" align="center">{@ext_pack_building} <br /> {@ext_name} {@ext_level}</td>
    </tr>
    {/if}

    {if[{var}ext_pack_research{/var}]}
    <tr>
        <td colspan="3" align="center">{@ext_pack_research} <br /> {@ext_name} {@ext_level}</td>
    </tr>
    {/if}
  {/if}

</table>
{if[0]}
 Oxsar http://oxsar.ru
 Copyright (c) 2009-2010 UnitPoint <support@unitpoint.ru>

{/if}