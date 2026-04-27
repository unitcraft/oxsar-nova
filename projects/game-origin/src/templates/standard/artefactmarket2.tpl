<script type="text/javascript">
//<![CDATA[
var decPoint = '{lang}DECIMAL_POINT{/lang}';
var thousandsSep = '{lang}THOUSANDS_SEPERATOR{/lang}';

$(document).ready(function() {
	var building_credit = $( '.raw_credit', $('#art_pack_build_level').parent() ).val();
	var research_credit = $( '.raw_credit', $('#art_pack_research_level').parent() ).val();
	var building_link = $( 'a', $('#art_pack_build').parent() ).attr('href');
	var research_link = $( 'a', $('#art_pack_research').parent() ).attr('href');
	var change_build_stuff = function(){
		var type = $('#art_pack_build').parent().attr('id');
		var construction = $(":selected", $('#art_pack_build')).val();
		var level = $(":selected", $('#art_pack_build_level')).val();
  	  	$( 'img', $('#art_building_img_td') ).attr('src', '<?php echo artImageUrl("image_new", ""); /*{const=FULL_URL}new_game/index.php?r=artefact2user_YII/image_new&*/ ?>cid=' + construction + '&level=' + level + '&typeid=' + type);
  	  	$( 'a', $('#art_pack_build').parent() ).attr('href', '{const=FULL_URL}game.php/BuyArtefact/{const=ARTEFACT_PACKED_BUILDING}' + '/con_id:' + construction + '/level:' + level + '?' );
  	  	$( '.credit_cost', $('#art_pack_build').parent().parent() ).html( fNumber(building_credit*1*level*1) );
	};
	var change_research_stuff = function(){
		var type = $('#art_pack_research').parent().attr('id');
		var construction = $(":selected", $('#art_pack_research')).val();
		var level = $(":selected", $('#art_pack_research_level')).val();
  	  	$( 'img', $('#art_research_img_td') ).attr('src', '<?php echo artImageUrl("image_new", ""); /*{const=FULL_URL}new_game/index.php?r=artefact2user_YII/image_new&*/ ?>cid=' + construction + '&level=' + level + '&typeid=' + type);
  	  	$( 'a', $('#art_pack_research').parent() ).attr('href','{const=FULL_URL}game.php/BuyArtefact/{const=ARTEFACT_PACKED_RESEARCH}' + '/con_id:' + construction + '/level:' + level + '?' );
  	  	$( '.credit_cost', $('#art_pack_research').parent().parent() ).html( fNumber(research_credit*1*level*1) );
	};
	
	var temp = $('.option0', $('#art_pack_build'));
	temp.attr('selected','selected');
	change_build_stuff();

	temp = $('.option0', $('#art_pack_research'));
	temp.attr('selected','selected');
	change_research_stuff();
	
	$('#art_pack_build').change(function () {
		change_build_stuff();
	});
	$('#art_pack_research').change(function () {
		change_research_stuff();
    });
	$('#art_pack_build_level').change(function () {
		change_build_stuff();
    });
	$('#art_pack_research_level').change(function () {
		change_research_stuff();
    });
});
//]]>
</script>

<table class="ntable">
  {if[0]}
  <tr>
    <th colspan="3">{lang}ARTEFACT_MARKET{/lang}</th>
  </tr>
  <tr>
    <th colspan="2">{include}"select_style_form"{/include}</th>
    <th style="text-align:center">{lang}QUANTITY{/lang}</th>
  </tr>
  {else}
  <tr>
    <th colspan="2">{lang}ARTEFACT_MARKET{/lang}</th>
    <th style="text-align:center">{lang}QUANTITY{/lang}</th>
  </tr>
  {/if}
  
  <form method="post" action="{@formaction}" style="padding:0; margin:0">
  <input type="hidden" id="typeid" name="typeid" value="0" />
  {foreach[artefacts]}
  <tr>
    <td width="1px"{if[$key == ARTEFACT_PACKED_BUILDING]} id="art_building_img_td"{/if}{if[$key == ARTEFACT_PACKED_RESEARCH]} id="art_research_img_td"{/if}>{loop}image{/loop}</td>
    <td valign="top">
      <div style="width:100%">
        {if[$row['flags']]}<span style="float:right">{loop}flags{/loop}</span>{/if}
        {loop}name{/loop}
      </div>
      <div style="clear:both; font-size:smaller">{loop}description{/loop}</div>
      <div>
        {if[1 || $row["can_build"]]}
          <table class="table_no_background" cellspacing="0" cellpadding="0" border="0">
            {include}"artefact_row_info"{/include}
            
            {include}"required_res_info"{/include}
            
            {if[$row["productiontime"]]}sd
            <tr title="{lang=REQUIRES}">
              <td>{lang=REQUIRE_TIME}</td>
              <td colspan="2">{loop=productiontime}</td>
            </tr>
            {/if}
          </table>
        {/if}
        {if[$row["required_constructions"]]}
          <span class="normal">{lang}REQUIRED_LIST_TITLE{/lang}</span>
          <br />{loop=required_constructions}
        {/if}
      </div>
    </td>
    <td width="100px" align="center" valign="top" id="<?php echo $key; ?>">
    	<input type='hidden' class='raw_credit' value='{loop}raw_credit{/loop}'>
      {if[$row['quantity_num'] > 0]}
        {loop}quantity{/loop}
        <br />
      {/if}
      {if[$key == ARTEFACT_PACKED_BUILDING]}
      <br />
      <select id='art_pack_build' style="width: 120px;">
  		{foreach2[.selects]}
      		<option value='{loop}sel_id{/loop}' class='option<?php echo $key; ?>'>{loop}sel_name{/loop}</option>
      	{/foreach2}
      </select>
      <select id='art_pack_build_level' style="width: 120px;">
  		{foreach2[.levels]}
      		<option value='{loop}sel_id{/loop}' class='option<?php echo $key; ?>'>{lang}LEVEL{/lang}:{loop}sel_id{/loop}</option>
      	{/foreach2}
      </select>
      {/if}
      {if[$key == ARTEFACT_PACKED_RESEARCH]}
      <br />
      <select id='art_pack_research' style="width: 120px;">
  		{foreach2[.selects]}
      		<option value='{loop}sel_id{/loop}' class='option<?php echo $key; ?>'>{loop}sel_name{/loop}</option>
      	{/foreach2}
      </select>
      <select id='art_pack_research_level' style="width: 120px;">
  		{foreach2[.levels]}
      		<option value='{loop}sel_id{/loop}' class='option<?php echo $key; ?>'>{lang}LEVEL{/lang}:{loop}sel_id{/loop}</option>
      	{/foreach2}
      </select>
      {/if}
      <br />{loop}buy{/loop}
    </td>
  </tr>
  {/foreach}
  </form>
</table>
{if[0]}
 Oxsar https://oxsar-nova.ru
 Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
  
{/if}