<script type="text/javascript">

{if[defined('SN') && !defined('SN_FULLSCREEN')]}
$(function(){
	var d_height	= $('body').height() - 20;
	var d_width		= $('body').width() - 20;
	d_height = {const=MAX_HEIGHT};
	$('#assault-report-dialog').dialog({
		autoOpen:		false,
		position:		'center',
		modal: 			true,
		closeOnEscape: 	true,
		draggable:		false,
		resizable:		false,
		height:			d_height,
		width:			d_width
	});
	$('.assault-report').click(function(){
		$('#assault-report-dialog').dialog('open');
		var data = '<iframe src ="'
			+ $(this).attr('href')
			+ '" width="'
			+ (d_width - 50)
			+ '" height="'
			+ (d_height - 60)
			+ '" id="assault_IFrame"></iframe>';
		$('#assault-report-dialog').html(data);
		$('#assault_IFrame').load(function()
		{
			$(this).contents().find('a').attr('target', '_top');
		});
		return false;
	});
});
{/if}

$(function() {
	$("#date_first").datepicker({ dateFormat: 'dd.mm.yy' });
	$("#date_last").datepicker({ dateFormat: 'dd.mm.yy' });
});

function openWindow(url)
{
	win = window.open(url, "{lang}ASSAULT_REPORT{/lang}", "width=600,height=400,status=yes,scrollbars=yes,resizable=yes");
	win.focus();
}

function setOrder(field)
{
    page_select = document.getElementById('page');
    page_select.value = 1;

    so_input = document.getElementById('sort_order');
    sf_input = document.getElementById('sort_field');

    if(sf_input.value != field)
    {
      sf_input.value = field;
      so_input.value = 'desc';
    }
    else
    {
      so_input.value = so_input.value == 'asc' ? 'desc' : 'asc';
    }

    document.getElementById('go').click();
}

function goPage(page)
{
    document.getElementById('page').value = page;
    document.getElementById('go').click();
}

</script>

{if[defined('SN')]}
<div id="assault-report-dialog" title="{lang}ASSAULT_REPORT_TITLE{/lang}">
</div>
{/if}
<form name="form_battles" id="form_battles" method="post" action="{@formaction}">
<table class="ntable">
	<tr>
		<th>{lang}MENU_BATTLESTATS{/lang}</th>
	</tr>
	<tr>
		<td>
		  <table cellspacing="0" cellpadding="0" border="0" class="table_no_background">
		    <tr>
		      <td nowrap="nowrap"><label for="date_last">{lang}DATE_LAST{/lang}</label></td>
		      <td><input name="date_last" id="date_last" type="text" value="{@date_last}"/></td>
		      <td rowspan="2" valign="top">{@date_interval_comment}</td>
		    </tr>
		    <tr>
		      <td nowrap="nowrap"><label for="date_first">{lang}DATE_FIRST{/lang}</label></td>
		      <td><input name="date_first" id="date_first" type="text" value="{@date_first}"/></td>
		    </tr>
		    <tr>
		      <td nowrap="nowrap"><label for="user_filter">{lang}BS_USER_FILTER{/lang}</label></td>
		      <td><input name="user_filter" id="user_filter" type="text" value="{@user_filter}"/></td>
		      <td>{lang}BS_USER_FILTER_INFO{/lang}</td>
		    </tr>
		    <tr>
		      <td nowrap="nowrap"><label for="alliance_filter">{lang}BS_ALLIANCE_FILTER{/lang}</label></td>
		      <td><input name="alliance_filter" id="alliance_filter" type="text" value="{@alliance_filter}"/></td>
		      <td>{lang}BS_ALLIANCE_FILTER_INFO{/lang}</td>
		    </tr>
		    <tr>
		      <td colspan="3">
            <input type="checkbox" name="show_drawn" id="show_drawn" value="1"{if[$this->templateVars["show_drawn"]]} checked="checked"{/if} />
            <label for="show_drawn">{lang}BS_SHOW_DRAWN{/lang}</label>
		      </td>
		    </tr>
		    <tr>
		      <td colspan="3">
		        <input type="checkbox" name="show_no_destroyed" id="show_no_destroyed" value="1"{if[$this->templateVars["show_no_destroyed"]]} checked="checked"{/if} />
		        <label for="show_no_destroyed">{lang}BS_SHOW_NO_DESTROYED{/lang}</label>
		      </td>
		    </tr>
		    <tr>
		      <td colspan="3">
		        <input type="checkbox" name="show_aliens" id="show_aliens" value="1"{if[$this->templateVars["show_aliens"]]} checked="checked"{/if} />
		        <label for="show_aliens">{lang}BS_SHOW_UFO_BATTLES{/lang}</label>
		      </td>
		    </tr>
		    <tr>
		      <td colspan="3">
            <input type="checkbox" name="new_moon" id="new_moon" value="1"{if[$this->templateVars["new_moon"]]} checked="checked"{/if} />
            <label for="new_moon">{lang}BS_NEW_MOON{/lang}</label>
		      </td>
		    </tr>
		    <tr>
		      <td colspan="3">
            <input type="checkbox" name="moon_battle" id="moon_battle" value="1"{if[$this->templateVars["moon_battle"]]} checked="checked"{/if} />
            <label for="moon_battle">{lang}BS_MOON_BATTLE{/lang}</label>
		      </td>
		    </tr>
		  </table>
    </td>
   </tr>
   <tr>
    <td class="center">
      <input type="hidden" name="page" id="page" value="1">
      <input type="hidden" name="sort_field" id="sort_field" value="{@sort_field}">
      <input type="hidden" name="sort_order" id="sort_order" value="{@sort_order}">
			<input type="submit" name="go" id="go" value="{lang}COMMIT{/lang}" class="button" />
		</td>
	</tr>
</table>
<table class="ntable">
	<colgroup>
		<col width="10%" />
		<col width="*"/>
		<col width="15%" />
		<col width="15%" />
		<col width="5%" />
                <col width="10%" />
	</colgroup>
	<thead>
    <tr>
      <th style="text-align: center" colspan="6">
        <div style="float: right">
          <select name="page1" id="page1" onchange="goPage(this.value)">{@pages}</select>
        </div>
        {@link_first}
        {@link_prev}
        {@page_links}
        {@link_next}
        {@link_last}
      </th>
    </tr>
    <tr>
      <th><a href="#" onclick="setOrder('date'); return false">{lang}DATE{/lang}</a><sup>{@date}</sup></th>
	    <th><a href="#" onclick="setOrder('planet_name'); return false">{lang}PLANET{/lang}</a> / <a href="#" onclick="setOrder('outcome'); return false">{lang}BS_OUTCOME{/lang}</a><sup>{@outcome}</sup></th>
      <th><a href="#" onclick="setOrder('attacker_lost'); return false">{lang}ATTACKER_LOST{/lang}</a><sup>{@attacker_lost}</sup></th>
	    <th><a href="#" onclick="setOrder('defender_lost'); return false">{lang}DEFENDER_LOST{/lang}</a><sup>{@defender_lost}</sup></th>
      <th><a href="#" onclick="setOrder('moon'); return false">{lang}MOON{/lang}</a><sup>{@moon}</sup></th>
	    <th><a href="#" onclick="setOrder('parts'); return false">{lang}PARTS{/lang}</a><sup>{@parts}</sup><sup>1</sup></th>
    </tr>
  </thead>
	<tfoot>
    <tr>
      <th style="text-align: center" colspan="6">
        <div style="float: right">
          <select name="page2" id="page2" onchange="goPage(this.value)">{@pages}</select>
        </div>
        {@link_first}
        {@link_prev}
        {@page_links}
        {@link_next}
        {@link_last}
      </th>
    </tr>
    <tr>
		  <td colspan="6">
			  <span class="assault-defeat">{lang}DFNDR_DEFEAT{/lang}</span>
			  &nbsp;&nbsp;&nbsp;<span class="assault-victory">{lang}DFNDR_VICTORY{/lang}</span>
        &nbsp;&nbsp;&nbsp;<span class="assault-drawn">{lang}DRAWN{/lang}</span>
		  </td>
	  </tr>
    <tr>
      <td style="text-align: center" colspan="6"><sup>1</sup> {lang}PARTS{/lang} - {lang}PARTS_NOTICE{/lang}</td>
    </tr>
	</tfoot>
	<tbody>
	  {foreach[statistics]}
	  <tr>
            <td nowrap="nowrap">{loop}date{/loop}</td>
            <td><b>{loop}planet_name{/loop}</b></td>
            <td nowrap="nowrap">{loop}attacker_lost{/loop}</td>
            <td nowrap="nowrap">{loop}defender_lost{/loop}</td>
            <td nowrap="nowrap">{loop}moonchance{/loop}</td>
            <td nowrap="nowrap">{loop}attackers{/loop} - {loop}defenders{/loop} ({loop}total_parts{/loop})</td>
	  </tr>
	  {/foreach}
	</tbody>
</table>
</form>
{if[0]}
 Oxsar http://oxsar.ru
 Copyright (c) 2009-2010 UnitPoint <support@unitpoint.ru>
  
{/if}