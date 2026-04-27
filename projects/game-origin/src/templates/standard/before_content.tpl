{if[OXSAR_RELEASED && !mt_rand(0, 4) && !DEATHMATCH]}
<div style="padding:10px; text-align:center">
	<?php
		$msg = array(
			'<font color = "#00FF00">Помогай новичкам словом и делом!</font>',
			'<font color = "#00FF00">Помогая новичкам, ты помогаешь игре!</font>',
			'Игрок удаляется, если месяц отсутствует в игре.',
			'Отпуск автоматически выключается, если игрок месяц отсутствует.',
			'Выработка шахт выключается после трех дней отсутствия игрока.',
			'Не пропадай, выработка шахт выключается после трех дней отсутствия игрока.',
		);
		if(!defined('SN'))
		{
			$msg = array_merge($msg, array(
				'<a href="Stock"><font color="#00FF00">Купи новенький флот на бирже прямо сейчас!</font></a>',
				'<a href="Stock"><font color = "#00FF00">Нужен срочно флот, ресурсы или артефакты?</font> - загляни на биржу.</a>',
				'<a href="Payment"><font color = "#00FF00">Пополняя кредиты, ты помогаешь своей любимой игре!</font></a>',
				'<a href="Payment"><font color = "#00FF00">Пополнение кредитов через 2PAY:</font> СМС (<b>Россия</b>, <b>Украина</b>, Казахстан, Германия и мн. др. страны), <b>Электронные деньги</b> и др. способы</a>',
				'<a href="Payment"><font color = "#00FF00">Выгодное пополнение кредитов:</font> 2PAY (десятки всевозможных способов), WebMoney, SMS, ROBOKASSA.</a>',
				'<a href="Payment"><font color = "#00FF00">Пополни кредиты прямо сейчас:</font> 2PAY (десятки всевозможных способов), WebMoney, SMS, ROBOKASSA.</a>',
				'<a href="Payment"><font color = "#00FF00">Пополняй кредиты, помогай развитию игры!</font> 2PAY: СМС (<b>Россия</b>, <b>Украина</b>, Казахстан, Германия и мн. др. страны).</a>',
				));
		}
		echo $msg[array_rand($msg)];
	?>
</div>
{/if}

{if[!defined('SN')]}
{if[!OXSAR_RELEASED && !isFacebookSkin()]}
<table width="600" cellpadding=0 cellspacing=0 border=0 style="margin-bottom:-20px;">
  <tr>
    <td style="padding:10px; text-align:center">
      <span class="false"><b><font size="+1">OXSAR {const=OXSAR_VERSION}</font></b></span>
      <br />
      <a href="/forums/index.php?showtopic=1679" target="_blank">
        Это версия игры предназначена для тестирования новых возможностей.
        Информацию о найденных ошибках, предложения и коментарии оставляйте пожалуйста
        на форум в эту тему.
      </a>
    </td>
  </tr>
</table>
{/if}
{/if}
