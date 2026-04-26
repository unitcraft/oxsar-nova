<script type="text/javascript">
//<![CDATA[
function calculator(form) {
metal = form.metal.value;
silicon = form.silicon.value;

if(isNaN(metal)) { metal = 0; }
if(isNaN(silicon)) { silicon = 0; }

hydrogen_0 = Math.round((metal * ({@curs_hydrogen}/{@curs_metal})) + (silicon * ({@curs_hydrogen}/{@curs_silicon})));
hydrogen = Math.round(((metal * ({@curs_hydrogen}/{@curs_metal})) + (silicon * ({@curs_hydrogen}/{@curs_silicon}))) * ({@comis}/100 + 1));
komis = Math.round(hydrogen_0 * {@comis}/100);

if(metal != 0) { max_metal = Math.round({@storageMetal} - {@metal} - metal); }
else { max_metal = 0; }
if(silicon != 0) { max_silicon = Math.round({@storageSilicon} - {@silicon} - silicon); }
else { max_silicon = 0; }

if(hydrogen > {@hydrogen}) { hydrogen = "---"; error_1 = "<font class=false2>Для совершения сделки не хватает ресурса</font>" }
else { var error_1 = ""; }
if(max_metal < 0) { hydrogen = "---"; error_2 = "<font class=false2>Не хватает места в хранилище металла</font>" }
else { var error_2 = ""; }
if(max_silicon < 0) { hydrogen = "---"; error_3 = "<font class=false2>Не хватает места в хранилище кремния</font>" }
else { var error_3 = ""; }

form.hydrogen.value = hydrogen;

var hydrogenl_0 = document.getElementById("hydrogenl_0")
hydrogenl_0.innerHTML = hydrogen_0;
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
		<td><span id="errors_2"></span></td>
	</tr>
	<tr>
		<td>{lang}SILICON{/lang}</td>
		<td width="1">{image[SILICON]}silicon.gif{/image}</td>
		<td><input type="text" name="silicon" size="12" maxlength="12"  onchange="this.value=this.value.replace(/([^0-9])/g,''); calculator(this.form);" onkeyup="var n=this.value.replace(/([^0-9])/g,''); if(n!=this.value) this.value=n; calculator(this.form);"></span></td>
		<td><span id="errors_3"></td>
	</tr>
	<thead><tr>
		<th colspan="4">Необходимо водорода:</th>
	</tr></thead>
	<tr>
		<td colspan="2">По курсу</td>
		<td colspan="2"><span id="hydrogenl_0"></span></td>
	</tr>
	<tr>
		<td colspan="2">Комиссия</td>
		<td colspan="2"><span id="komiss"></span></td>
	</tr>
	<tr>
		<td width="15%">Всего</td>
		<td width="1">{image[HYDROGEN]}hydrogen.gif{/image}</td>
		<td width="15%"><input type="text" name="hydrogen" size="12" maxlength="12"  readonly></td>
		<td><span id="errors_1"></span></td>
	</tr>
	<tr>
		<td colspan="4" class="center"><input type="submit" class="button" name="ex_hydrogen" value="Обменять" /></td>
	</tr>
</table>
</form>
{if[0]}
 Oxsar http://oxsar.ru
 Copyright (c) 2009-2010 UnitPoint <support@unitpoint.ru>
  
{/if}