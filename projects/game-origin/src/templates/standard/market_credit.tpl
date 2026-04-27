<script type="text/javascript">
//<![CDATA[
function calculator(form)
{
metal = form.metal.value;
silicon = form.silicon.value;
hydrogen = form.hydrogen.value;

if(isNaN(metal)) { metal = 0; }
if(isNaN(silicon)) { silicon = 0; }
if(isNaN(hydrogen)) { hydrogen = 0; }

credit_metal_0 = Math.ceil(metal * ({@curs_credit}/{@curs_metal}));
credit_silicon_0 = Math.ceil(silicon * ({@curs_credit}/{@curs_silicon}));
credit_hydrogen_0 = Math.ceil(hydrogen * ({@curs_credit}/{@curs_hydrogen}));

credit_0 = credit_metal_0 + credit_silicon_0 + credit_hydrogen_0;
credit = Math.ceil(credit_0 * ({@comis}/100 + 1));
comis = credit - credit_0;

if(metal != 0) { max_metal = Math.round({@storageMetal} - {@metal} - metal); }
else { max_metal = 0; }

if(silicon != 0) { max_silicon = Math.round({@storageSilicon} - {@silicon} - silicon); }
else { max_silicon = 0; }

if(hydrogen != 0) { max_hydrogen = Math.round({@sotrageHydrogen} - {@hydrogen} - hydrogen); }
else { max_hydrogen = 0; }

if(credit > {@credit_row}) { credit = "---"; error_1 = "<font class=false2>Для совершения сделки не хватает кредитов</font>" }
else { var error_1 = ""; }

// if(credit_metal_0 <= 0 && metal > 0) { credit = "---"; error_2 = "<font class=false2>Слишком мало металла, укажите металла минимум на 1 кредит!</font>" }
if(max_metal < 0) { credit = "---"; error_2 = "<font class=false2>Не хватает места в хранилище металла</font>" }
else { var error_2 = ""; }

if(max_silicon < 0) { credit = "---"; error_3 = "<font class=false2>Не хватает места в хранилище кремния</font>" }
else { var error_3 = ""; }

if(max_hydrogen < 0) { credit = "---"; error_4 = "<font class=false2>Не хватает места в хранилище водорода</font>" }
else { var error_4 = ""; }

form.credit.value = credit;

var el_credit_0 = document.getElementById("credit_0")
el_credit_0.innerHTML = credit_0;
var el_comis = document.getElementById("comis")
el_comis.innerHTML = comis;

var el_error_1 = document.getElementById("el_error_1")
el_error_1.innerHTML = error_1;
var el_error_2 = document.getElementById("el_error_2")
el_error_2.innerHTML = error_2;
var el_error_3 = document.getElementById("el_error_3")
el_error_3.innerHTML = error_3;
var el_error_4 = document.getElementById("el_error_4")
el_error_4.innerHTML = error_4;
}
//]]>
</script>

<form method="post" action="{@sendAction}">
<table class="ntable">
	<thead><tr>
		<th colspan="4">Приобрести ресурсы за кредиты</th>
	</tr></thead>
	{if[{var=comis}]}
	<tr>
		<td class="center" colspan="4">Внимание!<br/>С каждой сделки взымается комиссия в размере {@comis}% от продаваемого ресурса.</td>
	</tr>
	{/if}
	<tr>
		<td>{lang}METAL{/lang}</td>
		<td width="1">{image[METAL]}met.gif{/image}</td>
		<td><input type="text" name="metal" size="12" maxlength="12"  onchange="this.value=this.value.replace(/([^0-9])/g,''); calculator(this.form);" onkeyup="var n=this.value.replace(/([^0-9])/g,''); if(n!=this.value) this.value=n; calculator(this.form);"></td>
		<td><span id="el_error_2"></span></td>
	</tr>
	<tr>
		<td>{lang}SILICON{/lang}</td>
		<td width="1">{image[SILICON]}silicon.gif{/image}</td>
		<td><input type="text" name="silicon" size="12" maxlength="12"  onchange="this.value=this.value.replace(/([^0-9])/g,''); calculator(this.form);" onkeyup="var n=this.value.replace(/([^0-9])/g,''); if(n!=this.value) this.value=n; calculator(this.form);"></td>
		<td><span id="el_error_3"></span></td>
	</tr>
	<tr>
		<td>{lang}HYDROGEN{/lang}</td>
		<td width="1">{image[HYDROGEN]}hydrogen.gif{/image}</td>
		<td><input type="text" name="hydrogen" size="12" maxlength="12"  onchange="this.value=this.value.replace(/([^0-9])/g,''); calculator(this.form);" onkeyup="var n=this.value.replace(/([^0-9])/g,''); if(n!=this.value) this.value=n; calculator(this.form);"></td>
		<td><span id="el_error_4"></span></td>
	</tr>
	<thead><tr>
		<th colspan="4">Необходимо кредитов</th>
	</tr></thead>
	<tr>
		<td colspan="2">По курсу</td>
		<td colspan="2"><span id="credit_0"></span></td>
	</tr>
	<tr>
		<td colspan="2">Комиссия</td>
		<td colspan="2"><span id="comis"></span></td>
	</tr>
	<tr>
		<td width="15%">Всего</td>
		<td width="1">{image[CREDIT]}credit.gif{/image}</td>
		<td width="15%"><input type="text" name="credit" size="12" maxlength="12"  readonly></td>
		<td><span id="el_error_1"></span></td>
	</tr>
	<tr>
		<td colspan="4" class="center"><input type="submit" class="button" name="ex_credit" value="Обменять" /></td>
	</tr>
</table>
</form>
{if[0]}
 Oxsar https://oxsar-nova.ru
 Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
  
{/if}