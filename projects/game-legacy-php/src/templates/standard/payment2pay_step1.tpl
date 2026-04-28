<script>
//<![CDATA[
	function convCredit()
  {
    var el = document.getElementById("amount");
		var el_value = el.value;
		var r_max = {@r_max};
		var r_min = {@r_min};
		var r_change = {@r_change};
		var purse_t = document.getElementById('purse_t').value;
		if(isInteger(el_value) == false){
			document.getElementById("amount").value = el_value.substr(0, el_value.length-1);
			//alert("only numbers allowed");
			return null;
		} else {
			
			if(purse_t == 'RUR')
      {
        to_pay = Math.max(el_value * r_change, r_min);
        if(to_pay > r_max)
        {
          el_value = Math.round(r_max / r_change);
          to_pay = el_value * r_change;
          el.value = el_value;
        }
      }
						
		  // num = document.getElementById('to_pay').value / 2;
			document.getElementById('to_pay').innerHTML = (Math.round(to_pay * 100)/100)+" "+purse_t;
		}
	}

	function isInteger(val)
	{
	    if(val==null)
	    {
	        //alert(val);
	        return false;
	    }
	    if (val.length==0)
	    {
	        //alert(val);
	        return false;
	    }
	    for (var i = 0; i < val.length; i++) 
	    {
	        var ch = val.charAt(i)
	        if (ch < "0" || ch > "9")
	        {
	            return false
	        }
	    }
	    return true
	}
//]]>
</script>

<form method="post" action="" target="_blank">
<input type="hidden" name="user_id" value="{@user_id}">
<input type="hidden" name="sid" value="{@sid}">
<table class="ntable">
	<thead><tr>
		<th colspan="2">Пополнить счет</th>
	</tr></thead>
	<tr>
		<td><img src="{const=RELATIVE_URL}images/market-money.png" alt=""/></td>
		<td width="100%">
      {lang=PAY_MAIN_TEXT}
      <!-- <p />
      Максимальная сумма пополнения за раз: <nobr>{@r_max} RUR</nobr> -->
    </td>
	</tr>
</table>
<table class="ntable">
	<tr>
		<td class="right" width="50%">Введите сумму кредитов для покупки:</td>
    <td class="left" width="50%">
      <input type="text" id="amount" name="amount" value="" onchange="this.value=this.value.replace(/([^0-9])/g,''); convCredit();" onkeyup="var n=this.value.replace(/([^0-9])/g,''); if(n!=this.value) this.value=n; convCredit();" />
    </td>
	</tr>
	<tr>
		<td class="right" width="50%">Валюта по умолчанию</td>
    <td class="left" width="50%">
		<input type="hidden" name="purse_t" value="RUR" id="purse_t">
       RUR
    </td>
	</tr>
	<tr>
		<td class="right" width="50%">К оплате:</td>
    <td class="left" width="50%"><span id="to_pay"></span></td>
	</tr>
	<tr>
		<td colspan="2" class="center"><input type="submit" name="The2pay2" value="Перейти к оплате" class="button" /></td>
	</tr>
	<tr>
		<td colspan="2" class="center">Курс: USD : RUR : EUR = {@z_change} : {@r_change} : {@e_change}</td>
	</tr>
	<tr>
		<td colspan="2">{lang=PAY_SUPPORT_TEXT}</td>
	</tr>
</table>
</form>
{if[0]}
 Oxsar https://oxsar-nova.ru
 Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
  
{/if}