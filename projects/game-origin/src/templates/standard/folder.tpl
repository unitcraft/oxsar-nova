<script type="text/javascript">
//<![CDATA[
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
		var data = '<iframe src ="' + $(this).attr('href') + '" width="' + (d_width - 35) + 'px" height="' + (d_height - 60) + 'px" id="assault_IFrame"></iframe>'
		$('#assault-report-dialog').html(data);
		$('#assault_IFrame').load(function(){
			$(this).contents().find('a').attr('target', '_top');
		});
		return false;
	});
});
{/if}
function openWindow(url)
{
	win = window.open(url, "{lang}ASSAULT_REPORT{/lang}", "width=600,height=400,status=yes,scrollbars=yes,resizable=yes");
	win.focus();
}
function goPage(page){
    page_select = document.getElementById('page');
    page_select.value = page;
    document.getElementById('go').click();
}
//]]>
</script>
{if[defined('SN')]}
<div id="assault-report-dialog" title="{lang}ASSAULT_REPORT{/lang}">
</div>
{/if}
<table class="ntable">
	<thead>
        <tr>
            <th style="text-align: center" colspan="4">
                <form name="form_pages" id="form_pages" method="post" action="{@formaction}">
                <div style="float: right;">
                  <select name="page" id="page" onchange="document.getElementById('go').click();">{@pages}</select>
                </div>
                {@link_first}
                {@link_prev}
                {@page_links}
                {@link_next}
                {@link_last}
                <input type="submit" style="display: none;" name="go" id="go" value="{lang}COMMIT{/lang}" class="button" />
                </form>
            </th>
        </tr>
        <tr>
		<th>{if[$this->templateVars["mode"] != 2]}{lang}FROM{/lang}{else}{lang}RECEIVER{/lang}{/if}</th>
		<th>{lang}SUBJECT{/lang}</th>
		<th>{lang}DATE{/lang}</th>
		<th>{lang}ACTION{/lang}</th>
	</tr></thead>
  <form method="post" action="{@formaction}">
  <input type="hidden" name="msgid[]" value="0" />
	<tfoot><tr>
		<td colspan="4" class="center">
			{if[count($this->getLoop("messages")) == 0]}{lang}NO_MATCHES_FOUND{/lang}
			{else}<select name="deleteOption"><option value="1">{lang}DELETE_ALL_MARKED{/lang}</option><option value="2">{lang}DELETE_ALL_NON_MARKED{/lang}</option><option value="3">{lang}DELETE_ALL_SHOWN{/lang}</option><option value="4">{lang}EMPTY_FOLDER{/lang}</option><option value="5">{lang}REPORT_TO_MODERATOR{/lang}</option></select>
                        <input type="hidden" name="page" value="{@page}" />
			<input type="submit" name="delete" value="{lang}COMMIT{/lang}" class="button" />{/if}
		</td>
	</tr>
        <tr>
                <th style="text-align: center" colspan="4">
                    <div style="float: right;">
                      <select name="page2" id="page2" onChange="goPage(this.value)">{@pages}</select>
                    </div>
                    {@link_first}
                    {@link_prev}
                    {@page_links}
                    {@link_next}
                    {@link_last}
                </th>
        </tr>
        </tfoot>
	<tbody>
	{foreach[messages]}
	  <tr{loop}odd{/loop}>
		  <td>{loop}sender{/loop}</td>
		  <td>{loop}subject{/loop}</td>
		  <td>{loop}time{/loop}</td>
		  <td><input type="checkbox" name="msgid[]" value="{loop}msgid{/loop}" /></td>
	  </tr>
	  <tr{loop}odd{/loop}>
		  <td colspan="3">{loop}msg{/loop}</td>
		  <td></td>
	  </tr>
	{/foreach}
	</tbody>
  </form>
</table>
{if[0]}
 Oxsar http://oxsar.ru
 Copyright (c) 2009-2010 UnitPoint <support@unitpoint.ru>
  
{/if}