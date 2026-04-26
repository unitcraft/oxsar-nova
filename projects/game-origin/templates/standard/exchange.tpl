<script type="text/javascript">
$(function() {
	$("#date_first").datepicker({ dateFormat: 'dd.mm.yy' });
	$("#date_last").datepicker({ dateFormat: 'dd.mm.yy' });
});
</script>

<form name="form_exchange_settings" id="form_exchange_settings" method="post" action="{@formaction}">
<table class="ntable">
	<thead>
		<tr>
			<th>{lang}MENU_EXCHANGE{/lang}</th>
		</tr>
	</thead>
	<tfoot>
		<tr>
			<td><input class="button" type="submit" value="{lang}COMMIT{/lang}" name="saveexchangesettings"></td>
		</tr>
	</tfoot>
	<tr><td>
		<table cellspacing="0" cellpadding="0" border="0" class="table_no_background">
			<colgroup>
				<col width="25%">
				<col width="*">
				<col width="*">
			</colgroup>
			<tr>
				<td>{lang}EXCH_TITLE{/lang}</td>
				<td colspan="2"><input id="exchangetitle" name="exchangetitle" type="text" maxlength="10" value="{@exchange_title}" /></td>
			</tr>
			<tr>
				<td>{lang}EXCH_FEE{/lang}</td>
				<td><input id="exchangefee" name="exchangefee" size="3" maxlength="3" type="text" value="{@exchange_fee}" /></td>
				<td><div class="small">{lang}EXCH_FEE_NOTE{/lang}</div></td>
			</tr>
			<tr>
				<td>{lang}EXCH_DEF_FEE{/lang}</td>
				<td><input id="exchangedeffee" name="exchangedeffee" type="text" size="3" maxlength="3" value="{@exchange_def_fee}" /></td>
				<td><div class="small">{lang}EXCH_DEF_FEE_NOTE{/lang}</div></td>
			</tr>
			<tr>
				<td>{lang}EXCH_COMISSION{/lang}</td>
				<td><input id="exchangecomission" name="exchangecomission" type="text" size="3" maxlength="3" value="{@exchange_comission}" /></td>
				<td><div class="small">{lang}EXCH_COMISSION_NOTE{/lang}</div></td>
			</tr>
			<tr>
				<td>{lang}EXCH_SLOTS{/lang}:</td>
				<td nowrap="nowrap">{@exchange_cur_slots} / {@exchange_max_slots}</td>
				<td>{lang=COMPUTER_TECH}: {@exchange_comp_tech}</td>
			</tr>
			<tr>
				<td>{lang}EXCH_DISTANCE{/lang}:</td>
				<td>{@exchange_radius}</td>
				<td>{lang=UNIT_EXCH_SUPPORT_RANGE}: {@exchange_radius_units}</td>
			</tr>
			<tr>
				<td>{lang}EXCH_LOTS{/lang}:</td>
				<td nowrap="nowrap">{@exchange_oc_lots} / {@exchange_max_lots}</td>
				<td>{lang=UNIT_EXCH_SUPPORT_SLOT}: {@exchange_max_lots_units}</td>
			</tr>
		</table>
	</td></tr>
</table>
</form>

<form name="form_exchange_stats" id="form_exchange_stats" method="post" action="{@formaction}">
<table class="ntable">
	<thead>
		<tr>
			<th colspan="4">{lang}STATISTICS{/lang}</th>
		</tr>
	</thead>
        <!--<tfoot>
        </tfoot>-->
        <tr>
            <td colspan="4">
                <table class="table_no_background">
                    <tr>
		      <td nowrap="nowrap"><label for="date_last">{lang}DATE_LAST{/lang}</label></td>
		      <td><input name="date_last" id="date_last" type="text" value="{@date_last}"/></td>
                      <td rowspan="2">{lang}EXCH_DATE_COMMENT{/lang}</td>
                      <td rowspan="2"><input class="button" name="go" id="go" type="submit" value="{lang}COMMIT{/lang}"></td>
		    </tr>
		    <tr>
		      <td nowrap="nowrap"><label for="date_first">{lang}DATE_FIRST{/lang}</label></td>
		      <td><input name="date_first" id="date_first" type="text" value="{@date_first}"/>
                          <input type="hidden" name="sort_field" id="sort_field" value="{@sort_field}">
                          <input type="hidden" name="sort_order" id="sort_order" value="{@sort_order}">
                      </td>
		    </tr>
                </table>
            </td>
        </tr>
</table>                    
<table class="ntable">
	<colgroup>
		<col width="20%" />
		<col width="20%"/>
		<col width="20%" />
		<col width="20%" />
                <col width="20%" />
	</colgroup>
    <thead>
    <tr>
        <th colspan="5">
        <table class="table_no_background center" style="width: 100%;">
        <colgroup>
            <col width="25%" />
            <col width="25%" />
            <col width="25%" />
            <col width="25%" />
        </colgroup>
        <tr>
            <td>{lang}EXCH_LOTS_TOTAL{/lang}:</td>
            <td>{lang}EXCH_LOTS_SOLD{/lang}:</td>
            <td>{lang}EXCH_TURNOVER{/lang}:</td>
            <td>{lang}EXCH_PROFIT{/lang}:</td>
        </tr>
        <tr>
            <td>{@lots_total}</td>
            <td>{@lots_sold}</td>
            <td>{@turnover}</td>
            <td>{@profit}</td>
        </tr>
        </table>
        </th>
    </tr>
    <tr>
      <th style="text-align: center" colspan="5">
        <div style="float: right">
          <select name="page" id="page" onchange="goPage(this.value)">{@pages}</select>
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
	    <th><a href="#" onclick="setOrder('lot'); return false">{lang}EXCH_LOT{/lang}</a><sup>{@lot}</sup></th>
      <th style="text-align:right"><a href="#" onclick="setOrder('lot_amount'); return false">{lang}EXCH_LOT_AMOUNT{/lang}</a><sup>{@lot_amount}</sup></th>
      <th style="text-align:right"><a href="#" onclick="setOrder('lot_price'); return false">{lang}EXCH_LOT_PRICE{/lang}</a><sup>{@lot_price}</sup></th>
      <th style="text-align:right"><a href="#" onclick="setOrder('lot_profit'); return false">{lang}EXCH_PROFIT{/lang}</a><sup>{@lot_profit}</sup></th>
    </tr>
  </thead>
	<tfoot>
    <tr>
      <th style="text-align: center" colspan="5">
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
		<td nowrap="nowrap">{loop}date{/loop}</td>
		<td><b>{loop}lot_name{/loop}</b></td>
		<td align="right" nowrap="nowrap"><b>{loop}amount{/loop}</b></td>
		<td align="right" nowrap="nowrap">{loop}price{/loop}</td>
		<td align="right" nowrap="nowrap"{loop}profit_class{/loop}>{loop}lot_profit{/loop}</td>
	  </tr>
	  {/foreach}
	</tbody>
</table>
</form>
{if[0]}
 Oxsar http://oxsar.ru
 Copyright (c) 2009-2010 UnitPoint <support@unitpoint.ru>
  
{/if}