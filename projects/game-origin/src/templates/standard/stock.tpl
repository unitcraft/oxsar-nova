<script type="text/javascript">

lang_CONFIRM_TITLE_WARNING = "{lang=CONFIRM_TITLE_WARNING}";
lang_CONFIRM_OK = "{lang=CONFIRM_OK}";
lang_CONFIRM_CANCEL = "{lang=CONFIRM_CANCEL}";

$(function() {
  $("#date_first").datepicker({ dateFormat: 'dd.mm.yy' });
  $("#date_last").datepicker({ dateFormat: 'dd.mm.yy' });
});

</script>

<form name="form_exchanges" id="form_exchanges" method="post" action="{@formaction}">
<table class="ntable">
	<tr>
		<th>{lang}MENU_STOCK{/lang}</th>
	</tr>
	<tr>
		<td>
		  <table cellspacing="0" cellpadding="0" border="0" class="table_no_background">
		    <tr>
		      <td nowrap="nowrap"><label for="lot_type">{lang}EXCH_LOT_TYPE{/lang}</label></td>
		      <td><select name="lot_type" id="lot_type">{@lot_types}</select></td>
		    </tr>
            <tr>
		      <td nowrap="nowrap"><label for="exch_filter">{lang}EXCH_FILTER_COMMENT{/lang}</label></td>
		      <td><select name="exch_filter" id="exch_filter">{@exch_filters}</select></td>
		    </tr>
			<tr>
				<td colspan="2">
					<br/>
					{if[$this->templateVars["has_tp"]]}{@new_lot_link} ({@used_slots} / {@total_slots}){/if}
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
		  <input type="hidden" name="whereid" id="whereid" value="{@where}">
		  <input type="hidden" name="id" id="id" value="{@id}">
		  <input type="submit" name="go" id="go" value="{lang}COMMIT{/lang}" class="button" />
		</td>
    </tr>
</table>

<table class="ntable">
	<colgroup>
		<col width="1px" />
		<col width="*" />
		<col width="10%" />
		<col width="10%" />
		<col width="10%" />
		<col width="10%" />
	</colgroup>

	{if[count($this->getLoop("featured_lots")) > 0]}

		<tr>
		  <th colspan="6">{lang=EXCH_LOTS_FEATURED}</th>
		</tr>

		{foreach[featured_lots]}
		<tr>
		  {if[Core::getUser()->get("imagepackage") != "empty"]}
		  <td align="center">{loop}image{/loop}</td>
		  {else}
		  <td>{loop}i{/loop}.</td>
		  {/if}
		  <td>
            {if[$row['featured_date']]}
                <table cellspacing="0" cellpadding="0" border="0" class="table_no_background"><tr><td>{@vip_image}</td><td>
            {/if}
            {loop}lot_name{/loop}
			{if[ $row['packed'] && $row['new_level'] > $row['cur_level'] ]}
				<nobr>({loop=cur_level} -> {loop=new_level})</nobr>
			{/if}
            {if[$row['featured_date']]}
                </td></tr></table>
			{/if}
		  </td>
		  <td align="right">{loop}amount{/loop}{if[$row["lot_min_amount"]]}<br />{loop}lot_min_amount{/loop}{/if}</td>
		  <td align="right">{loop}price{/loop}{if[$row["disc_price"]]} ({loop}disc_price{/loop}){/if}{if[$row["lot_min_amount"]]}<br />{loop}min_price{/loop}{if[$row["disc_min_price"]]} ({loop}disc_min_price{/loop}){/if}{/if}
		  </td>
		  <td align="right"><nobr>{loop}distance{/loop} ({loop}fly_time{/loop})</nobr>
			  <br /> {loop}planetname{/loop} <br /> {loop}title{/loop}</td>
		  <td align="center" valign="middle">
			  {loop}action{/loop}
		  </td>
		</tr>
		{if[0 && isAdmin() && $row["action"]]}
		<tr>
		  <td colspan="6" align="right">
			  {loop}action{/loop}
		  </td>
		</tr>
		{/if}
		{/foreach}

		<tr>
		  <td colspan="6" align="center">{@premiun_list_max_size}</td>
		</tr>

	{/if}

	<tr>
	  <th colspan="6" style="text-align: center">
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
	{if[ !{var=merchant_mark_used} ]}
        {if[1]}
        <tr>
            <th colspan="6" style="text-align: center">Комиссия для покупателя и продавца</th>
        </tr>
        <tr>
            <td colspan="6" align="center">
                <table cellspacing="0" cellpadding="0" border="0" class="table_no_background">
                    <tr>
                        <td>Знак торговца: активирован</td>
                        <td class="true">{const=EXCH_MERCHANT_PREMIUM_COMMISSION}% (для премиум)</td>
                        <td>{const=EXCH_MERCHANT_COMMISSION}% (для обычных)</td>
                    </tr>
                    <tr>
                        <td>Знак торговца: неактивирован</td>
                        <td>{const=EXCH_NO_MERCHANT_PREMIUM_COMMISSION}% (для премиум)</td>
                        <td class="false2">{const=EXCH_NO_MERCHANT_COMMISSION}% (для обычных)</td>
                    </tr>
                </table>
            </td>
        </tr>
        <tr>
            <td colspan="6" style="text-align: center" class="false2">Активируйте артефакт Знак торговца, чтобы дешевле покупать
                или увеличить прибыль от продажи своих товаров. Также вы можете обменять ресурсы
                в <a href="<?php echo htmlspecialchars(socialUrl(RELATIVE_URL . 'game.php/Market')); ?>">Обменнике</a> и/или приобрести артефакты
                в <nobr><a href="<?php echo htmlspecialchars(socialUrl(RELATIVE_URL . 'game.php/ArtefactMarket')); ?>">Магазине артефактов</a></nobr></td>
        </tr>
        {/if}
	{/if}
	<tr class="center">
	  {if[Core::getUser()->get("imagepackage") == "empty"]}
	  <th>{lang}SEQ_NUM{/lang}</th>
	  <th>
	  {else}
	  <th colspan="2">
	  {/if}
		  <a href="#" onclick="setOrder('lot'); return false">{lang}EXCH_LOT{/lang}</a></th>
	  <th><a href="#" onclick="setOrder('lot_amount'); return false">{lang}EXCH_LOT_AMOUNT{/lang}</a>
			<br />{lang}EXCH_LOT_MIN_AMOUNT{/lang}</th>

	  <th><a href="#" onclick="setOrder('lot_price'); return false">{lang}EXCH_LOT_PRICE{/lang}</a>
			<br />{lang}EXCH_LOT_MIN_PRICE{/lang}</th>
	  <th style="text-align:right"><a href="#" onclick="setOrder('distance'); return false">{lang}DISTANCE{/lang}</a>
		  <br /><a href="#" onclick="setOrder('seller'); return false">{lang}EXCH_SELLER{/lang}</a>
		  <br />{lang}MENU_STOCK{/lang}</th>
	  <th>&nbsp;</th>
	</tr>

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
	</tfoot>

	{foreach[lots]}
	<tr>
	  {if[Core::getUser()->get("imagepackage") != "empty"]}
	  <td align="center">{loop}image{/loop}</td>
	  {else}
	  <td>{loop}i{/loop}.</td>
	  {/if}
	  <td>
        {if[$row['featured_date']]}
            <table cellspacing="0" cellpadding="0" border="0" class="table_no_background"><tr><td>{@vip_image}</td><td>
        {/if}
        {loop}lot_name{/loop}
		{if[ $row['packed'] && $row['new_level'] > $row['cur_level'] ]}
			<nobr>({loop=cur_level} -> {loop=new_level})</nobr>
		{/if}
        {if[$row['featured_date']]}
            </td></tr></table>
        {/if}
	  </td>
	  <td align="right">{loop}amount{/loop}{if[$row["lot_min_amount"]]}<br />{loop}lot_min_amount{/loop}{/if}</td>
	  <td align="right">{loop}price{/loop}
		{if[$row["disc_price"]]}(<span class='trade-union'>{loop}disc_price{/loop}</span>){/if}
		{if[$row["lot_min_amount"]]}<br />{loop}min_price{/loop}
			{if[$row["disc_min_price"]]}(<span class='trade-union'>{loop}disc_min_price{/loop}</span>){/if}
		{/if}
	  </td>
	  <td align="right"><nobr>{loop}distance{/loop} ({loop}fly_time{/loop})</nobr>
		  <br /> {loop}planetname{/loop} <br /> {loop}title{/loop}</td>
	  <td align="center" valign="middle">
		  {loop}action{/loop}
		  {if[$row['ban']]}
		  <br />{loop}banlink{/loop}
		  {/if}
	  </td>
	</tr>
	{/foreach}
</table>
</form>
{if[0]}
 Oxsar http://oxsar.ru
 Copyright (c) 2009-2010 UnitPoint <support@unitpoint.ru>

{/if}