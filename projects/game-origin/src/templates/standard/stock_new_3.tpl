<script type="text/javascript">
//<![CDATA[
var outMetal = {@metal};
var outSilicon = {@silicon};
var outHydrogen = {@hydrogen};
var sCapacity = {@capacity};
var tMetal = {@metal};
var tSilicon = {@silicon};
var tHydrogen = {@hydrogen};
var capacity = {@capacity};
var decPoint = '{lang}DECIMAL_POINT{/lang}';
var thousandsSep = '{lang}THOUSANDS_SEPERATOR{/lang}';
var MARKET_METAL = {@market_metal};
var MARKET_SILICON = {@market_silicon};
var MARKET_HYDROGEN = {@market_hydrogen};
var MARKET_CREDIT = {@market_credit};
var planetRatio = {@planet_ratio};
var fee = {@fee};
var minPrice = {@min_price};
var maxPrice = {@max_price};
var widePrice = 0;
var resLot = {@loadRes};
var sellerMinProfit = {@exch_seller_min_profit};
var sellerDefProfit = {@exch_seller_def_profit};
var met_eq = {@met_eq};
var min_ally_disc = {@min_ally_disc};
var max_ally_disc = {@max_ally_disc};
var sellerMaxProfit = {@sellerMaxProfit};
var quantity = {@quantity};

function renewRest()
{
	var loaded = parseSafeInt($('#resource').val());
	if( loaded > capacity )
	{
		$('#resource').val(capacity);
		loaded = capacity;
	}
	var rest = capacity - loaded;
	obj = $('#rest');
	if (rest >= 0) {
        obj.addClass('true');
        obj.removeClass('false');
	} else {
        obj.addClass('false');
        obj.removeClass('true');
	}
	obj.text(fNumber(rest));
	$('#quant').val($('#resource').val());
}

function calcCost()
{
	var loaded = parseSafeInt($('#resource').val());
	var type = $('#res_type').val();

	var metal = 0;
	var silicon = 0;
	var hydrogen = 0;
	var metal_equiv = 0;
	if (type == -1){ metal = loaded; }
	else if (type == -2){ silicon = loaded; }
	else if (type == -3){ hydrogen = loaded; }

	metal_equiv = metal + silicon * MARKET_METAL / MARKET_SILICON + hydrogen * MARKET_METAL / MARKET_HYDROGEN;
	var real_price = metal_equiv * MARKET_CREDIT / (MARKET_METAL * planetRatio);
	minPrice =	Math.ceil(Math.max(1, real_price + real_price*(parseSafeInt($('#discount').val()) + sellerMinProfit + fee) / 100));
	maxPrice =	Math.floor(Math.max(1, real_price + real_price*(parseSafeInt($('#discount').val()) + sellerMaxProfit + fee) / 100));

	widePrice = parseSafeInt($('#wide_price').val());
	if(widePrice > minPrice-1) widePrice = minPrice-1;
	if(widePrice < 0) widePrice = 0;
	if($('#wide_price').val() != widePrice)
	{
		$('#wide_price').val(widePrice);
	}
	minPrice = Math.max(1, minPrice - widePrice);
	maxPrice += widePrice;

	var price =	real_price + real_price*(parseSafeInt($('#discount').val()) + sellerDefProfit + fee) / 100;
	$('#price').val(Math.ceil(Math.max(1, price)));
	$('#min_price').text(fNumber(minPrice, 2));
	$('#max_price').text(fNumber(maxPrice, 2));
	$('#real_price').text(fNumber(real_price, 2));
	$('#cost_met').text(fNumber(metal));
	$('#cost_sil').text(fNumber(silicon));
	$('#cost_hyd').text(fNumber(hydrogen));
	$('#met_equiv').text(fNumber(metal_equiv));
	met_eq = metal_equiv;

	updatePriceCommission();
}

function updatePriceCommission()
{
	var price = parseSafeInt($('#price').val());
    {if[ EXCH_NEW_PROFIT_TYPE ]}
        var premium_price = Math.max({const=EXCH_PREMIUM_MIN_COST}, price * {const=EXCH_PREMIUM_PERCENT}/100.0);
        $('#premium_price').text(fNumber(premium_price, 2));

        var discount = $('#discount').val();
        function showPrice(prefix, name, commission, is_trade)
        {
            var p = price * ({const=EXCH_COMMISSION_BASE_UNIT} + commission/100.0);
            if(is_trade){
                p *= 1 - discount/100.0;
            }
            if(commission < 0){
                var exch = p * fee/100.0;
                $('#'+prefix+'exch_'+name).text(fNumber(exch, 2));
                p -= exch;
            }
            $('#'+prefix+name).text(fNumber(p, 2));
        };
        for(var i = 0; i < 2; i++){
            var is_trade = i > 0, prefix = i > 0 ? 'trade_' : '';

            showPrice(prefix, 'price_no_merchant_commission', -{const=EXCH_NO_MERCHANT_COMMISSION}, is_trade);
            showPrice(prefix, 'price_merchant_commission', -{const=EXCH_MERCHANT_COMMISSION}, is_trade);

            showPrice(prefix, 'buyer_price_no_merchant_commission', +{const=EXCH_NO_MERCHANT_COMMISSION}, is_trade);
            showPrice(prefix, 'buyer_price_merchant_commission', +{const=EXCH_MERCHANT_COMMISSION}, is_trade);

            showPrice(prefix, 'price_no_merchant_premium_commission', -{const=EXCH_NO_MERCHANT_PREMIUM_COMMISSION}, is_trade);
            showPrice(prefix, 'price_merchant_premium_commission', -{const=EXCH_MERCHANT_PREMIUM_COMMISSION}, is_trade);

            showPrice(prefix, 'buyer_price_no_merchant_premium_commission', +{const=EXCH_NO_MERCHANT_PREMIUM_COMMISSION}, is_trade);
            showPrice(prefix, 'buyer_price_merchant_premium_commission', +{const=EXCH_MERCHANT_PREMIUM_COMMISSION}, is_trade);
        }
    {else}
        var p = price * (1 - fee/100.0);
        $('#price_commission').text(fNumber(p, 2));

        $('#price_no_merchant_commission').text(fNumber(p * (1 - {const=EXCH_NO_MERCHANT_COMMISSION}/100.0), 2));
        $('#price_merchant_commission').text(fNumber(p * (1 - {const=EXCH_MERCHANT_COMMISSION}/100.0), 2));

        $('#buyer_price_no_merchant_commission').text(fNumber(price * (1 + {const=EXCH_NO_MERCHANT_COMMISSION}/100.0), 2));
        $('#buyer_price_merchant_commission').text(fNumber(price * (1 + {const=EXCH_MERCHANT_COMMISSION}/100.0), 2));

        $('#price_no_merchant_premium_commission').text(fNumber(p * (1 - {const=EXCH_NO_MERCHANT_PREMIUM_COMMISSION}/100.0), 2));
        $('#price_merchant_premium_commission').text(fNumber(p * (1 - {const=EXCH_MERCHANT_PREMIUM_COMMISSION}/100.0), 2));

        $('#buyer_price_no_merchant_premium_commission').text(fNumber(price * (1 + {const=EXCH_NO_MERCHANT_PREMIUM_COMMISSION}/100.0), 2));
        $('#buyer_price_merchant_premium_commission').text(fNumber(price * (1 + {const=EXCH_MERCHANT_PREMIUM_COMMISSION}/100.0), 2));
    {/if}
}

function checkMinPrice()
{
	var uprice = parseSafeInt($('#price').val());
	var disc = parseSafeInt($('#discount').val());
	if( disc > max_ally_disc )
	{
		disc = max_ally_disc;
		$('#discount').val(max_ally_disc);
	}
	if( disc < min_ally_disc )
	{
		disc = min_ally_disc;
		$('#discount').val(min_ally_disc);
	}

	{if[!{var=is_artefact}]}
	var real_price = met_eq * MARKET_CREDIT / (MARKET_METAL * planetRatio);
	minPrice = Math.ceil(Math.max(1, real_price + real_price*(disc + sellerMinProfit + fee) / 100));
	maxPrice = Math.floor(Math.max(1, real_price + real_price*(disc + sellerMaxProfit + fee) / 100));
	{else}
	minPrice = {@min_price};
	maxPrice = {@max_price};
	{/if}

	widePrice = parseSafeInt($('#wide_price').val());
	if(widePrice > minPrice-1) widePrice = minPrice-1;
	if(widePrice < 0 || minPrice < 1) widePrice = 0;
	if($('#wide_price').val() != widePrice)
	{
		$('#wide_price').val(widePrice);
	}
	minPrice -= widePrice;
	maxPrice += widePrice;

	if(uprice == 0 && widePrice == 0)
	{
		{if[!{var=is_artefact}]}
		uprice = Math.ceil(real_price + real_price*(disc + sellerDefProfit + fee) / 100);
		{else}
		uprice = Math.round((minPrice + maxPrice) / 2);
		{/if}
		$('#price').val(uprice);
	}

	$('#min_price').text(fNumber(minPrice, 2));
	$('#max_price').text(fNumber(maxPrice, 2));

	if(uprice < minPrice)
	{
		$('#price').val(minPrice);
		$('#price_commission').text(fNumber($('#price').val() * (1 - fee/100), 2, decPoint, thousandsSep));
	}
	else if(uprice > maxPrice)
	{
		$('#price').val(maxPrice);
		$('#price_commission').text(fNumber($('#price').val() * (1 - fee/100), 2));
	}
	updatePriceCommission();
}
$(function() {
	// updatePriceCommission();
    checkMinPrice();
	// $('#price_commission').text(fNumber($('#price').val() * (1 - fee/100), 2));
	$('#price').live('change', function(){
		checkMinPrice();
		/*
		if( $('#price').val() < minPrice )
		{
			$('#price').val( minPrice );
			$('#price_commission').text(fNumber($('#price').val() * (1 - fee/100), 2));
		}
		if( $('#price').val() > maxPrice )
		{
			$('#price').val( maxPrice );
			$('#price_commission').text(fNumber($('#price').val() * (1 - fee/100), 2));
		}
		*/
	});
	$('#wide_price').live('change', function(){
		$('#price').val(0);
		checkMinPrice();
	});
	$('#quant').live('change', function(){
		if( $('#resource').val() )
		{
			if( $(this).val() > $('#resource').val() )
			{
				$(this).val($('#resource').val());
			}
		}
		else
		{
			if(  $(this).val() > quantity )
			{
				$(this).val(quantity);
			}
		}
		if( $(this).val() <= 0 )
		{
			$(this).val(1);
		}
	});
	$('#res_type').live('change', function(){
		calcCost();
	});
	$('#resource').live('change', function(){
		renewRest();
		calcCost();
	});
});
//]]>
</script>
<form method="post" action="{@formaction}">
	<table class="ntable">
		<tr>
			<th colspan="5">{lang}EXCH_LOT_OPTIONS{/lang}</th>
		</tr>
		<tr>
			<td>{lang}FLEET{/lang}</td>
			<td colspan="4">{@fleet}</td>
		</tr>
		{if[{var=loadRes}]}
			<tr>
				<td>
					<label for="res_type">{lang}RESOURCE{/lang} </label><select name="res_type" id="res_type">
						<option value="-1">{lang}METAL{/lang}</option>
						<option value="-2">{lang}SILICON{/lang}</option>
						<option value="-3">{lang}HYDROGEN{/lang}</option>
					</select>
				</td>
				<td colspan="4">
					<input type="text" name="resource" id="resource" size="10" maxlength="10" />
				</td>
			</tr>
			<tr>
				<td>{lang}CAPICITY{/lang}</td>
				<td colspan="4"><span id="rest" class="available">{@rest}</span></td>
			</tr>
		{/if}
		{if[{var=art_type}]}
			<tr>
				<td nowrap="nowrap"><label for="ttl">{lang}EXCH_QUANT{/lang}</label></td>
				<td colspan="4"><input type="text" name="quant" id="quant" value="{@quantity}" size="10" /></td>
			</tr>
			<tr>
				<td>{lang}EXCH_COST{/lang}</td>
				<td colspan="4">{lang}METAL{/lang}: <span id="cost_met">{@cost_met}</span><br />
						{lang}SILICON{/lang}: <span id="cost_sil">{@cost_sil}</span><br />
						{lang}HYDROGEN{/lang}: <span id="cost_hyd">{@cost_hyd}</span><br />
				</td>
			</tr>
			<tr>
				<td>{lang}EXCH_MET_EQUIV{/lang}</td>
				<td colspan="4"><span id="met_equiv">{@met_equiv}</span></td>
			</tr>
		{/if}
		<tr>
			<td>{lang}EXCH_REAL_PRICE{/lang}</td>
			<td colspan="4"><span id="real_price">{@real_price_str}</span></td>
		</tr>
		<tr>
			<td><span class="false2">Расширить диапазон цен (кр.)</span></td>
			<td colspan="4"><input type="text" name="wide_price" id="wide_price" value="0" />
				<br /> <span class="false2">Разовый платеж при подаче заявки, при снятии заявки не возвращается.
				С его помощью можно расширить разрешенный диапазон цен, чтобы выставить цену лота дешевле или дороже стандартной.</span>
			</td>
		</tr>
		<tr>
			<td>Разрешенный диапазон цен</td>
			<td colspan="4">от <span id="min_price" class="false">{@min_price_str}</span> до <span id="max_price" class="false">{@max_price_str}</span> кредитов</td>
		</tr>
        {if[ EXCH_NEW_PROFIT_TYPE ]}
            <tr>
                <td nowrap="nowrap"><label for="price">{lang}EXCH_LOT_PRICE{/lang}</label></td>
                <td colspan="4"><input type="text" name="price" id="price" value="{@def_price}" /></td>
            </tr>
            <tr>
                <td nowrap="nowrap"><label for="discount">{lang}EXCH_DISCOUNT{/lang}</label></td>
                <td colspan="4"><input type="text" name="discount" id="discount" value="{@def_discount}" size="2" onchange="checkMinPrice();"/></td>
            </tr>
            <tr>
                <td nowrap="nowrap">Комиссия владельца биржи</td>
                <td colspan="4">{@fee_str}%</td>
            </tr>

            <tr>
                <td>&nbsp;</td>
                <th colspan="2">{lang=MERCHANTS_MARK}</th>
                <th colspan="2">нет</th>
            </tr>
            <tr>
                <td rowspan="2" style="vertical-align:bottom">Цена для обычных покупателей</td>
                <th>премиум</th>
                <th>обычный</th>
                <th>премиум</th>
                <th>обычный</th>
            </tr>
            <tr>
                <td><span id="buyer_price_merchant_premium_commission"></span></td>
                <td><span id="buyer_price_merchant_commission"></span></td>
                <td><span id="buyer_price_no_merchant_premium_commission"></span></td>
                <td><span id="buyer_price_no_merchant_commission"></span></td>
            </tr>
            <tr>
                <td><span class="false2">Владельцу биржи</span></td>
                <td><span id="exch_price_merchant_premium_commission"></span></td>
                <td><span id="exch_price_merchant_commission"></span></td>
                <td><span id="exch_price_no_merchant_premium_commission"></span></td>
                <td><span id="exch_price_no_merchant_commission"></span></td>
            </tr>
            <tr>
                <td><span class="true">Моя прибыль</span></td>
                <td><span id="price_merchant_premium_commission"></span></td>
                <td><span id="price_merchant_commission"></span></td>
                <td><span id="price_no_merchant_premium_commission"></span></td>
                <td><span id="price_no_merchant_commission"></span></td>
            </tr>

            <tr>
                <td>&nbsp;</td>
                <th colspan="2">{lang=MERCHANTS_MARK}</th>
                <th colspan="2">нет</th>
            </tr>
            <tr>
                <td rowspan="2" style="vertical-align:bottom">Цена для торговых союзов</td>
                <th>премиум</th>
                <th>обычный</th>
                <th>премиум</th>
                <th>обычный</th>
            </tr>
            <tr>
                <td><span id="trade_buyer_price_merchant_premium_commission"></span></td>
                <td><span id="trade_buyer_price_merchant_commission"></span></td>
                <td><span id="trade_buyer_price_no_merchant_premium_commission"></span></td>
                <td><span id="trade_buyer_price_no_merchant_commission"></span></td>
            </tr>
            <tr>
                <td><span class="false2">Владельцу биржи</span></td>
                <td><span id="trade_exch_price_merchant_premium_commission"></span></td>
                <td><span id="trade_exch_price_merchant_commission"></span></td>
                <td><span id="trade_exch_price_no_merchant_premium_commission"></span></td>
                <td><span id="trade_exch_price_no_merchant_commission"></span></td>
            </tr>
            <tr>
                <td><span class="true">Моя прибыль</span></td>
                <td><span id="trade_price_merchant_premium_commission"></span></td>
                <td><span id="trade_price_merchant_commission"></span></td>
                <td><span id="trade_price_no_merchant_premium_commission"></span></td>
                <td><span id="trade_price_no_merchant_commission"></span></td>
            </tr>
        {else}
            <tr>
                <td nowrap="nowrap"><label for="price">{lang}EXCH_LOT_PRICE{/lang}</label></td>
                <td colspan="4">
                    <input type="text" name="price" id="price" value="{@def_price}" />
                    <br />Если у покупателя активирован {lang=MERCHANTS_MARK}: <span id="buyer_price_merchant_commission"></span>, премиум: <span id="buyer_price_merchant_premium_commission"></span>
                    <br />Если у покупателя не активирован {lang=MERCHANTS_MARK}: <span id="buyer_price_no_merchant_commission"></span>, премиум: <span id="buyer_price_no_merchant_premium_commission"></span>
                </td>
            </tr>
            <tr>
                <td nowrap="nowrap"><label for="discount">{lang}EXCH_DISCOUNT{/lang}</label></td>
                <td colspan="4"><input type="text" name="discount" id="discount" value="{@def_discount}" size="2" onchange="checkMinPrice();"/></td>
            </tr>
            <tr>
                <td nowrap="nowrap">Комиссия биржи</td>
                <td colspan="4">{@fee_str}%</td>
            </tr>
            <tr>
                <td nowrap="nowrap">Моя прибыль</td>
                <td colspan="4">
                    {lang}EXCH_LOT_PRICE_COMMISSION{/lang}: <span id="price_commission"></span>
                    <br />Если у меня активирован {lang=MERCHANTS_MARK}: <span id="price_merchant_commission"></span>, премиум: <span id="price_merchant_premium_commission"></span>
                    <br />Если у меня не активирован {lang=MERCHANTS_MARK}: <span id="price_no_merchant_commission"></span>, премиум: <span id="price_no_merchant_premium_commission"></span>
                </td>
            </tr>
        {/if}
        {if[ EXCH_NEW_PROFIT_TYPE || {var=premium_price} ]}
        <tr>
            <td nowrap="nowrap"><label for="premium" class="false2">Премиум</label></td>
            <td colspan="4"><input type="checkbox" name="premium" id="premium" />
                <br /> <span class="false2">Автоматически поднять лот в список премиум за
                {if[EXCH_NEW_PROFIT_TYPE]}
                    <span id='premium_price'>~</span>
                    кредитов ({const=EXCH_PREMIUM_PERCENT}% от стоимости лота, минимум {const=EXCH_PREMIUM_MIN_COST} кредитов).
                {else}
                    {@premium_price} кредитов.
                {/if}
                При снятии заявки кредиты не возвращаются.</span>
            </td>
        </tr>
        {/if}
		<tr>
			<td colspan="5" class="center"><input type="submit" name="step4" value="{lang}NEXT{/lang}" class="button" /></td>
		</tr>
	</table>
</form>
{if[0]}
 Oxsar https://oxsar-nova.ru
 Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.

{/if}