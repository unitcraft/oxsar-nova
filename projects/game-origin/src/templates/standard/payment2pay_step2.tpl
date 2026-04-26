<!--- Слои с информацией -->
<div id="dvapay_emain" style="">
</div>
<div id="dvapay_terminals" style="float: left; width: 300px;">
</div>
<div id="dvapay_emoney" style="float: left; width: 300px;">
</div>
<br style="clear: both; "/>
<div id="dvapay_ecard" style="float: left; width: 300px;">
</div>
<div id="dvapay_ebank" style="float: left; width: 300px;">
</div>
<br style="clear: both; "/>
<div id="dvapay_esendmoney" style="float: left; width: 300px;">
</div>

<script>
var id='{@id_2pay}';
var v1='{@pay_id}';
var v2='{@user_id}';
var v3='{@game_name}';
var page='3021';
var country='0';
var conf='ssl, utf8';
document.write('<script type="text/javascript" src="http://2pay.ru/view/script.php?id='+id+'&v1='+v1+'&v2='+v2+'&v3='+v3+'&country='+country+'&page='+page+'&conf='+conf+'"></' + 'script>');
</script>
<!--- Скрипт загрузки, должен быть в конце страницы var id=''; номер вашего проекта в нашей системе var v1=''; v1 v2 v3 ник игрока\номер счета или аккаунт, если известен - вся служебная информация по идентификации пользователя var v2=''; если эта информация известна - нужно подставлять в v1 v2 v3 var v3=''; var page='3021';  служебная информация var country='0';  служебная информация var conf='123';  служебная информация -->
{if[0]}
 Oxsar http://oxsar.ru
 Copyright (c) 2009-2010 UnitPoint <support@unitpoint.ru>
  
{/if}