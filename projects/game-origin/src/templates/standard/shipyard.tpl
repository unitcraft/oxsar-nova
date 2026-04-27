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
{/if}

<table class="ntable">
  <tr>
    <th colspan="3">{@shipyard}</th>
  </tr>
  <tr>
    <th colspan="2">{include}"select_style_form"{/include}</th>
    <th style="text-align:center">{lang}QUANTITY{/lang}</th>
  </tr>
  
  <form method="post" action="{@formaction}" style="padding:0; margin:0">
  {foreach[shipyard]}
  <tr>
    <td width="1px">{loop}image{/loop}</td>
    <td valign="top">
      <div style="width:100%">
        <!-- {if[$row['quantity_num'] > 0]}<span style="float:right">{lang=QUANTITIES_EXIST}: {loop}quantity{/loop}</span>{/if} -->
        {loop}name{/loop}
      </div>
      <div style="clear:both; font-size:smaller">{loop}description{/loop} {perm[CAN_EDIT_CONSTRUCTIONS]}{loop}edit{/loop}{/perm}</div>
      <div>
        {if[$row["can_build"]]}
          {include}"required_res_table"{/include}
        {else}
          <span class="normal">{lang}REQUIRED_CONSTRUCTIONS{/lang}</span>
          <br />{loop=required_constructions}
        {/if}
      </div>
    </td>
    <td width="100px" align="center" valign="top">
      {if[$row['quantity_num'] > 0]}
        {loop}quantity{/loop}
        <br />
      {/if}
      <br />{loop}construct{/loop}
    </td>
  </tr>
  {/foreach}
  {if[{var=can_add_to_queue}]}
	  <tr>
	    <td colspan="3" align="center">
	      <input type="submit" name="sendmission" value="{lang}BUILD{/lang}" class="button" />
	    </td>
	  </tr>
  {/if}
  </form>
</table>
{if[0]}
 Oxsar https://oxsar-nova.ru
 Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
  
{/if}