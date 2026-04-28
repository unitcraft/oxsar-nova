<script type="text/javascript">
//<![CDATA[
var amount = {@amount};
var lot_min_amount = {@lot_min_amount};
var price = {@price};
var real_price = {@real_price};
var decPoint = '{lang}DECIMAL_POINT{/lang}';
var thousandsSep = '{lang}THOUSANDS_SEPERATOR{/lang}';
var fuel_price = {@fuel_price};
var total_consumption = {@total_consumption};
var delivery_percent = {@delivery_percent};
var delivery_hydro = {@delivery_hydro};
var fuel_mult = {@fuel_mult};

lang_CONFIRM_TITLE_WARNING = "{lang=CONFIRM_TITLE_WARNING}";
lang_CONFIRM_OK = "{lang=CONFIRM_OK}";
lang_CONFIRM_CANCEL = "{lang=CONFIRM_CANCEL}";

function checkAmount()
{
    $('#warn_multiple').css("visibility", "hidden");
    var user_amount = $('#buy_amount').val();
    if( user_amount <= 0)
    {
        user_amount = lot_min_amount;
        $('#warn_multiple').css("visibility", "visible");
    }
    else if (user_amount > amount)
    {
        user_amount = amount;
        $('#warn_multiple').css("visibility", "visible");
    }
    else if (user_amount % lot_min_amount != 0 && user_amount != amount)
    {
        user_amount -= user_amount % lot_min_amount;
        $('#warn_multiple').css("visibility", "visible");
    }

    $('#buy_amount').val(user_amount);
}

var lot_cost, lot_fuel_cost;

function calcPrice()
{
    var user_amount = $('#buy_amount').val();
    var k = user_amount / amount;
    var kprice = price * k;
    var kreal_price = real_price * k;

    var consumption = total_consumption * k;
    var fuel_seller = consumption * delivery_percent / 100;
    fuel_seller = Math.min(fuel_seller, delivery_hydro);
    var fuel_buyer = consumption - fuel_seller;
    var fuel_cost = fuel_mult * fuel_buyer * fuel_price / 1000;

    $('#buy_price').text(number_format(kprice, 2, decPoint, thousandsSep));
    $('#fuel_price').text(number_format(fuel_cost, 2, decPoint, thousandsSep));
    $('#total_price').text(number_format(fuel_cost + kprice, 2, decPoint, thousandsSep));

    $('#real_price').text(number_format(kreal_price, 2, decPoint, thousandsSep));

    lot_cost = fuel_cost + kprice;
    lot_fuel_cost = fuel_cost;
    var lot_real_cost = lot_cost - lot_fuel_cost;
    if(lot_fuel_cost > lot_real_cost*0.1){
        $('#fuel_price').addClass('false');
    }else{
        $('#fuel_price').removeClass('false');
    }
}

$(document).ready(function() {
    calcPrice();

    $('#buy').click(function(){
        var lot_real_cost = lot_cost - lot_fuel_cost;
        if(lot_fuel_cost > lot_real_cost*0.1){
            setTimeout(function(){
                var message = "{lang=CONFIRM_BUY_LOT_BIG_FUEL_COST}";
                message = message.replace(/{fuel_cost}/g, $('#fuel_price').text());
                message = message.replace(/{total_cost}/g, $('#total_price').text());
                confirm_dialog(message, {
                    cache: false,
                    ok: function(){
                        $('#buy_form').submit();
                    }
                });
            }, 1);
            return false;
        }
        return true;
    });
});

//]]>
</script>
<table class="ntable">
    <colgroup>
        <col width="50%"/>
        <col width="*"/>

    </colgroup>
    <tr>
        <th colspan="2">{lang}EXCH_LOT_DETAILS{/lang}</th>
    </tr>
    <tr>
        <td>{lang}EXCH_TITLE{/lang}</td>
        <td>{@exchange}</td>
    </tr>
    <tr>
        <td>{lang}EXCH_SELLER{/lang}</td>
        <td>{@seller}</td>
    </tr>
    {if[{var=showall}]}
    <tr>
        <td>{lang}RAISING_DATE{/lang}</td>
        <td>{@raising_date}</td>
    </tr>
    <tr>
        <td>{lang}EXCH_EXPIRY_DATE{/lang}</td>
        <td>{@expiry_date}</td>
    </tr>
    {/if}
    <tr>
        <td>{lang}EXCH_DELIVERY_MSG{/lang}</td>
        <td>{@delivery_str}</td>
    </tr>
    <tr>
        <td>{lang}EXCH_LOT{/lang}</td>
        <td>{@lot_name}</td>
    </tr>
    <tr>
        <td>{lang}EXCH_LOT_AMOUNT{/lang}</td>
        <td>{@amount_str}</td>
    </tr>
    <tr>
        <td>{lang}EXCH_QUANT{/lang}</td>
        <td>{@lot_min_amount_str}</td>
    </tr>

	{if[ {var=showBuy} ]}
	<tr>
		<td>{lang=ARTEFACT} {lang=MERCHANTS_MARK}</td>
		<td>
			{if[{var=merchant_mark_used}]}
				<span class="true">Активирован</span>
			{else}
				<span class="false2">Не активирован</span>
			{/if}
			{if[0]}<br /> комиссия {@comis}% (включена в цену){/if}
		</td>
    </tr>
	{/if}

    <tr>
        <td>{lang}EXCH_LOT_PRICE{/lang}</td>
        <td>{@price_str}</td>
    </tr>

    {if[{var=showBuy}]}
    <tr>
        <td>{lang}EXCH_REAL_PRICE{/lang}</td>
        <td>{@real_price_str}</td>
    </tr>
    {/if}

    <tr>
        <td>{lang}EXCH_DISCOUNT{/lang}</td>
        <td{@ally_discount_class}>{@ally_discount}</td>
    </tr>
</table>

{if[{var=showBuy}]}
<br/>
<form action="{@formaction}" method="post" id="buy_form">
<table class="ntable">
    <thead>
        <tr>
            <th colspan="{if[ {var=premium_image} ]}3{else}2{/if}">{lang}EXCH_PURCHASE{/lang}</th>
        </tr>
    </thead>
    <tfoot>
        <tr>
            <td colspan="{if[ {var=premium_image} ]}3{else}2{/if}">
                <input type="submit" name="buy" id="buy" value="{lang}BUY{/lang}" class="button" />
                <input type="hidden" name="lid" id="lid" value="{@lid}" />
            </td>
        </tr>
    </tfoot>
    <tr>
		{if[ {var=premium_image} ]}<td rowspan="3" style="width:1px; padding:0 10px">{@premium_image}</td>{/if}
        <td>{lang}EXCH_LOT_AMOUNT{/lang}</td>
        <td><input type="text" id="buy_amount" name="buy_amount" value ="{@amount}" onchange="checkAmount();calcPrice();"/> <span class="false" id="warn_multiple" style="visibility: hidden;">{lang}EXCH_MUST_BE_MULTIPLE{/lang}</span></td>
    </tr
    <tr>
        <td>{lang}EXCH_LOT_PRICE{/lang}</td>
        <td>
            <table cellspacing="0" cellpadding="0" border="0" class="table_no_background">
                <tr>
                    <td>{lang}EXCH_LOT{/lang}</td>
                    <td align="right"><span id="buy_price">{@price}</span></td>
                </tr>
                <tr>
                    <td>{lang}FUEL{/lang}</td>
                    <td align="right"><span id="fuel_price"></span></td>
                </tr>
                <tr>
                    <td>{lang}TOTAL{/lang}</td>
                    <td align="right"><span id="total_price"></span></td>
                </tr>
            </table>
        </td>
    </tr>
    <tr>
        <td>{lang}EXCH_REAL_PRICE{/lang}</td>
        <td><span id="real_price"></span></td>
    </tr>
</table>
</form>
{/if}
{if[0]}
 Oxsar https://oxsar-nova.ru
 Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.

{/if}