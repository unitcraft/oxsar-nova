<script type="text/javascript">

$(function() {
	$("#date_first").datepicker({ dateFormat: 'dd.mm.yy' });
  $("#date_last").datepicker({ dateFormat: 'dd.mm.yy' });
});

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

function goDetails(uid, gparam)
{
    document.getElementById('uid').value = uid;
    if (gparam == 1)
        document.getElementById('group_paramr').checked = true;
    else
        document.getElementById('group_params').checked = true;
    document.getElementById('go').click();
}

</script>


<form name="form_battles" id="form_battles" method="post" action="{@formaction}">
<table class="ntable">
	<tr>
		<th>{lang}MENU_RESTRANSFERS{/lang}</th>
	</tr>
	<tr>
		<td>
		  <table cellspacing="0" cellpadding="0" border="0" class="table_no_background">
		    <tr>
		      <td nowrap="nowrap"><label for="date_last">{lang}DATE_LAST{/lang}</label></td>
		      <td><input name="date_last" id="date_last" type="text" value="{@date_last}"/></td>
		    </tr>
		    <tr>
		      <td nowrap="nowrap"><label for="date_first">{lang}DATE_FIRST{/lang}</label></td>
		      <td><input name="date_first" id="date_first" type="text" value="{@date_first}"/></td>
		    </tr>
		    <tr>
		      <td colspan="2">
                        <input type="radio" name="group_param" id="group_paramr" value="1"{if[$this->templateVars["group_param"] == 1]} checked="checked"{/if} />
                        <label for="group_param">{lang}RT_GROUP_RECIEVER{/lang}</label>
		      </td>
		    </tr>
		    <tr>
		      <td colspan="2">
		         <input type="radio" name="group_param" id="group_params" value="2"{if[$this->templateVars["group_param"] == 2]} checked="checked"{/if} />
		        <label for="group_param">{lang}RT_GROUP_SENDER{/lang}</label>
		      </td>
		    </tr>
                  {if[$this->templateVars["detailed"] != 'det']}
                    <tr>
                        <td nowrap="nowrap"><label for="search_user">{lang}RT_SEARCH_USER{/lang}</label></td>
                        <td><input name="search_user" id="search_user" type="text" value="{@search_user_name}"/></td>
                    </tr>
                  {/if}
		  </table>
    </td>
   </tr>
   <tr>
    <td class="center">
      <input type="hidden" name="page" id="page" value="1">
      <input type="hidden" name="sort_field" id="sort_field" value="{@sort_field}">
      <input type="hidden" name="sort_order" id="sort_order" value="{@sort_order}">
      <input type="hidden" name="uid" id="uid" value="0">
			<input type="submit" name="go" id="go" value="{lang}COMMIT{/lang}" class="button" />
		</td>
	</tr>
</table>
<table class="ntable">
	<colgroup>
                <col width="*" />
		<col width="15%" />
		<col width="15%"/>
		<col width="15%" />
		<col width="15%" />
		<col width="15%" />
                <col width="15%" />
	</colgroup>
	<thead>
    <tr>
      <th style="text-align: center" colspan="7">
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
      <th>#</th>
      <th>{@column1_name}</th>
      <th>{@column2_name}</th>
      <th>{lang}TOTAL_RES_TRANSFER{/lang}</th>
      <th>{lang}METAL{/lang}</th>
      <th>{lang}SILICON{/lang}</th>
      <th>{lang}HYDROGEN{/lang}</th>
    </tr>
  </thead>
  <tfoot>
    <tr>
      <th style="text-align: center" colspan="7">
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
	</tfoot>
	<tbody>
	  {foreach[statistics]}
	  <tr>
            <td>{loop}col0_val{/loop}</td>
            {if[$this->templateVars["detailed"] == 'det']}<td nowrap="nowrap"><a href="#" onclick="goDetails({loop}rid{/loop}, 1); return false;">{loop}col1_val{/loop}</a></td>
            {else}
                <td nowrap="nowrap"><a href="#" onclick="document.getElementById('uid').value = {loop}uid{/loop}; document.getElementById('go').click(); return false;">{loop}col1_val{/loop}</a></td>
            {/if}
            {if[$this->templateVars["detailed"] == 'det']}<td nowrap="nowrap"><a href="#" onclick="goDetails({loop}sid{/loop}, 2); return false;">{loop}col2_val{/loop}</a></td>
            {else}
                <td nowrap="nowrap">{loop}col2_val{/loop}</td>
            {/if}
            <td nowrap="nowrap">{loop}resum{/loop}</td>
            <td nowrap="nowrap">{loop}metal{/loop}</td>
            <td nowrap="nowrap">{loop}silicon{/loop}</td>
            <td nowrap="nowrap">{loop}hydrogen{/loop}</td>
	  </tr>
	  {/foreach}
	</tbody>
</table>
</form>
{if[0]}
 Oxsar https://oxsar-nova.ru
 Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
  
{/if}