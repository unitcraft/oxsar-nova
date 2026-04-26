<script type="text/javascript">
//<![CDATA[
function calculator(form) {
silicon = form.silicon.value;
hydrogen = form.hydrogen.value;

if(isNaN(silicon)) { silicon = 0; }
if(isNaN(hydrogen)) { hydrogen = 0; }

metal = Math.round(((silicon * ({@curs_metal}/{@curs_silicon})) + (hydrogen * ({@curs_metal}/{@curs_hydrogen}))) * ({@comis}/100 + 1));
metal_0 = Math.round((silicon * ({@curs_metal}/{@curs_silicon})) + (hydrogen * ({@curs_metal}/{@curs_hydrogen})));
komis = Math.round(metal_0 * {@comis}/100);

if(silicon != 0) { max_silicon = Math.round({@storageSilicon} - {@silicon} - silicon); }
else { max_silicon = 0; }
if(hydrogen != 0) { max_hydrogen = Math.round({@sotrageHydrogen} - {@hydrogen} - hydrogen); }
else { max_hydrogen = 0; }

if(metal > {@metal}) { metal = "---"; error_1 = "<font class=false2>Для совершения сделки не хватает ресурса</font>" }
else { var error_1 = ""; }
if(max_silicon < 0) { metal = "---"; error_2 = "<font class=false2>Не хватает места в хранилище кремния</font>" }
else { var error_2 = ""; }
if(max_hydrogen < 0) { metal = "---"; error_3 = "<font class=false2>Не хватает места в хранилище водорода</font>" }
else { var error_3 = ""; }

form.metal.value = metal;

var metall_0 = document.getElementById("metall_0")
metall_0.innerHTML = metal_0;
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
		<td>{lang}SILICON{/lang}</td>
		<td width="1">{image[SILICON]}silicon.gif{/image}</td>
		<td><input type="text" name="silicon" size="12" maxlength="12"  onchange="this.value=this.value.replace(/([^0-9])/g,''); calculator(this.form);" onkeyup="var n=this.value.replace(/([^0-9])/g,''); if(n!=this.value) this.value=n; calculator(this.form);"></span></td>
		<td><span id="errors_2"></td>
	</tr>
	<tr>
		<td>{lang}HYDROGEN{/lang}</td>
		<td width="1">{image[HYDROGEN]}hydrogen.gif{/image}</td>
		<td><input type="text" name="hydrogen" size="12" maxlength="12"  onchange="this.value=this.value.replace(/([^0-9])/g,''); calculator(this.form);" onkeyup="var n=this.value.replace(/([^0-9])/g,''); if(n!=this.value) this.value=n; calculator(this.form);"></td>
		<td><span id="errors_3"></span></td>
	</tr>
	<thead><tr>
		<th colspan="4">Необходимо металла:</th>
	</tr></thead>
	<tr>
		<td colspan="2">По курсу</td>
		<td colspan="2"><span id="metall_0"></span></td>
	</tr>
	<tr>
		<td colspan="2">Комиссия</td>
		<td colspan="2"><span id="komiss"></span></td>
	</tr>
	<tr>
		<td width="15%">Всего</td>
		<td width="1">{image[METAL]}met.gif{/image}</td>
		<td width="15%"><input type="text" name="metal" size="12" maxlength="12"  readonly></td>
		<td><span id="errors_1"></span></td>
	</tr>
	<tr>
		<td colspan="4" class="center"><input type="submit" class="button" name="ex_metal" value="Обменять" /></td>
	</tr>
</table>
</form>
{if[0]}
 Oxsar http://oxsar.ru
 Copyright (c) 2009-2010 UnitPoint <support@unitpoint.ru>
  
{/if}