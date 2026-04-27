<script type="text/javascript">
//<![CDATA[
function calculator(form) {
metal = form.metal.value;
hydrogen = form.hydrogen.value;

if(isNaN(metal)) { metal = 0; }
if(isNaN(hydrogen)) { hydrogen = 0; }

silicon_0 = Math.round((metal * ({@curs_silicon}/{@curs_metal})) + (hydrogen * ({@curs_silicon}/{@curs_hydrogen})));
silicon = Math.round(((metal * ({@curs_silicon}/{@curs_metal})) + (hydrogen * ({@curs_silicon}/{@curs_hydrogen}))) * ({@comis}/100 + 1));
komis = Math.round(silicon_0 * {@comis}/100);

if(metal != 0) { max_metal = Math.round({@storageMetal} - {@metal} - metal); }
else { max_metal = 0; }
if(hydrogen != 0) { max_hydrogen = Math.round({@sotrageHydrogen} - {@hydrogen} - hydrogen); }
else { max_hydrogen = 0; }

if(silicon > {@silicon}) { silicon = "---"; error_1 = "<font class=false2>Для совершения сделки не хватает ресурса</font>" }
else { var error_1 = ""; }
if(max_metal < 0) { silicon = "---"; error_2 = "<font class=false2>Не хватает места в хранилище металла</font>" }
else { var error_2 = ""; }
if(max_hydrogen < 0) { silicon = "---"; error_3 = "<font class=false2>Не хватает места в хранилище водорода</font>" }
else { var error_3 = ""; }

form.silicon.value = silicon;

var siliconl_0 = document.getElementById("siliconl_0")
siliconl_0.innerHTML = silicon_0;
var komiss = document.getElementById("komiss")
komiss.innerHTML = komis;

var errors_1 = document.getElementById("errors_1")
errors_1.innerHTML = error_1;
var errors_2 = document.getElementById("errors_2")
errors_2.innerHTML = error_2;
var errors_3 = document.getElementById("errors_3")
errors_3.innerHTML = error_3;
}
//]]>
</script>

<form method="post" action="{@sendAction}">
<table class="ntable">
	<thead><tr>
		<th colspan="4">Приобрести ресурсы:</th>
	</tr></thead>
	<tr>
		<td class="center" colspan="4">Внимание!<br/>С каждой сделки взымается комиссия в размере {@comis}% от продаваемого ресурса.</td>
	</tr>
	<tr>
		<td>{lang}METAL{/lang}</td>
		<td width="1">{image[METAL]}met.gif{/image}</td>
		<td><input type="text" name="metal" size="12" maxlength="12"  onchange="this.value=this.value.replace(/([^0-9])/g,''); calculator(this.form);" onkeyup="var n=this.value.replace(/([^0-9])/g,''); if(n!=this.value) this.value=n; calculator(this.form);"></td>
		<td><span id="errors_2"></td>
	</tr>
	<tr>
		<td>{lang}HYDROGEN{/lang}</td>
		<td width="1">{image[HYDROGEN]}hydrogen.gif{/image}</td>
		<td><input type="text" name="hydrogen" size="12" maxlength="12"  onchange="this.value=this.value.replace(/([^0-9])/g,''); calculator(this.form);" onkeyup="var n=this.value.replace(/([^0-9])/g,''); if(n!=this.value) this.value=n; calculator(this.form);"></td>
		<td><span id="errors_3"></span></td>
	</tr>
	<thead><tr>
		<th colspan="4">Необходимо кремния:</th>
	</tr></thead>
	<tr>
		<td colspan="2">По курсу</td>
		<td colspan="2"><span id="siliconl_0"></span></td>
	</tr>
	<tr>
		<td colspan="2">Комиссия</td>
		<td colspan="2"><span id="komiss"></span></td>
	</tr>
	<tr>
		<td width="15%">Всего</td>
		<td width="1">{image[SILICON]}silicon.gif{/image}</td>
		<td width="15%"><input type="text" name="silicon" size="12" maxlength="12"  readonly></td>
		<td><span id="errors_1"></span></td>
	</tr>
	<tr>
		<td colspan="4" class="center"><input type="submit" class="button" name="ex_silicon" value="Обменять" /></td>
	</tr>
</table>
</form>
{if[0]}
 Oxsar https://oxsar-nova.ru
 Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
  
{/if}