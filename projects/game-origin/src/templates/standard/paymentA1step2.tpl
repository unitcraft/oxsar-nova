<table class="ntable">
	<thead>
		<tr>
			<th colspan="3">Пополнить счет</th>
		</tr>
	</thead>
	<tr>
		<td><img src="{const=RELATIVE_URL}images/market-money.png" alt=""/></td>
		<td width="100%">
			{lang=PAY_MAIN_TEXT}
			<p />
			<b>Текст СМС</b> нужно вводить без пробелов. Пожалуйста будьте внимательны при наборе текста.
			<p />
		<b>Для абонентов МТС</b><br />
		Стоимость доступа к услугам контент-провайдера устанавливается Вашим оператором. Подробную информацию можно узнать:<br />
		- в разделе «Услуги по коротким номерам» на сайте www.mts.ru<br />
		- в контактном центре по телефону 8 800 333 0890 (0890 для абонентов МТС)
		</td>
	</tr>
</table>
<form method="post" action="">
	<table class="ntable">
	{if[{var=isPaymentLocked}]}
		<tr>
			<td colspan="5" class="center">Временно отключено.</td>
		</tr>
	{else}
		<tr>
			<td colspan="5" class="center"><b>Пополнить через СМС. Страна: {@country}.</b></td>
		</tr>
		<tr>
			<td>Номер</td><td>Текст СМС</td><td colspan="2">Стоимость СМС (без НДС)</td><td>Получите кредитов</td>
		</tr>
		{foreach[sms_data]}
		<tr>
			<td>{loop}number{/loop}</td><td>{@prefix}{@user_id}</td>
			<td>от {loop}smsCostMin{/loop} {loop}smsVal{/loop} до {loop}smsCostMax{/loop} {loop}smsVal{/loop}</td>
			<td align=center><div class="ui-state-error ui-corner-all" style="height:16px; width:16px;"><a href="#" onclick="myshowdialog{loop}number{/loop}()" alt="Подробная информация о стоимости SMS"><span class="ui-icon ui-icon-help"></span></a></div></td>
			<td>{loop}credit{/loop}</td>
		</tr>
		{/foreach}
	{/if}
		<tr>
			<td colspan="5">{lang=PAY_SUPPORT_TEXT}</td>
		</tr>
	</table>
</form>


<?php
	$data = file_get_contents(APP_ROOT_DIR."cache/a1data.ser");
	$data = unserialize($data);

	foreach($data[$this->get('country')] as $value)
	{
?>

<div id="dialog<?php echo $value['number']?>" style="display:none;" title="Стоимость SMS сообщений на номер <?php echo $value['number']?>">
	<table id= align="center" width="100%" cellpadding=2 cellspacing=0 border=0>
	 <?php
		$vis_rows = 17;
		$all_rows = count($value['operatorsprice']);
		$cols = min(3, ceil($all_rows / $vis_rows));
		if($cols > 0)
		{
			$items_per_col = ceil($all_rows / $cols);
			$row_item_num = 0;
			$col_percent = ceil(100 / $cols);
			$row_num = 0;
		
			foreach ($value['operatorsprice'] as $operator_id => $cena_nomera)
			{
				$row_num++;
				if($row_item_num == 0)
				{
					if($row_num & 1)
						echo "<tr class='ui-state-default'>";
					else
						echo "<tr>";
				}
				if(++$row_item_num == $cols)
				{
					$row_item_num = 0;
				}
	?>
			<td width="<?=$col_percent?>%"> <?php echo $value['operatorsnames'][$operator_id]; ?></td>
			<td align='center' class='ui-state-error'><?php echo $cena_nomera . $value['smsVal'];?></td>
	<?php
				if($row_item_num == 0)
				{
					echo "</tr>";
				}
			}
			
			if($row_item_num < $cols)
			{
				echo str_repeat("<td>&nbsp;</td>", $cols - $row_item_num);
				echo "</tr>";
			}
	?>
	<?php
		} // if($cols > 0)
	?>
	</table>
</div>
<script type="text/javascript">
	$('#dialog<?php echo $value['number']?>').dialog({
			autoOpen: false,
			width: <?php echo ceil(400+$cols*150); ?>,
			position: "center"
		}
	);
	
	function myshowdialog<?php echo $value['number']?>()
	{
		$('#dialog<?php echo $value['number']?>').dialog('open');
	};
</script>
<?php
	} // foreach($data[$this->get('country')]
?>
{if[0]}
 Oxsar http://oxsar.ru
 Copyright (c) 2009-2010 UnitPoint <support@unitpoint.ru>
	
{/if}