<script type="text/javascript" src="{const=RELATIVE_URL}chat/chat.js"></script>
<style type="text/css" media="all">@import url({const=RELATIVE_URL}chat/chat.css);</style>

<form method="post" action="{@sendAction}" id='chat_form'>
<table class="ntable">
	<thead><tr>
		<th width="50%"><center>{@chat_link}</center></th>
		<th width="50%"><center>{@a_chat_link}</center></th>
	</tr></thead>
</table>
<table class="ntable">
	<tr>
		<td>
			<div id='chat_div'>
				{if[0]}
					Это окно чата
				{/if}
			</div>
		</td>
	</tr>
	<tr>
		<td>
            <table cellspacing="0" cellpadding="0" border="0" class="table_no_background">
				<tr>
					<td width="100%">
						<input type="text" name="shoutbox_message" id="shoutbox_message" size="56" maxlength="500"
							{if[ !isMobileSkin() ]}style="width:100%"{/if}
							/>
					{if[ !isMobileSkin() ]}
					</td>
					<td>
					{/if}
						<input onfocus="chek()" type="submit" class="button" name="send_message" value="Отправить" />
					</td>
					<td nowrap="nowrap"><span style="float:right">Читают: <span class="false" id="chat_online">~~</span></span></td>
				</tr>
            </table>
		</td>
	</tr>
	<tr>
		<td>
		<a href="javascript:insertbb('b')"><img title="Полужирный текст" src="{const=RELATIVE_URL}chat/bbcodes/b.gif" width="23" height="25" border="0"></a>
		<a href="javascript:insertbb('i')"><img title="Наклонный текст" src="{const=RELATIVE_URL}chat/bbcodes/i.gif" width="23" height="25" border="0"></a>
		<a href="javascript:insertbb('u')"><img title="Подчеркнутый текст" src="{const=RELATIVE_URL}chat/bbcodes/u.gif" width="23" height="25" border="0"></a>
		<a href="javascript:insertbb('s')"><img title="Зачеркнутый текст" src="{const=RELATIVE_URL}chat/bbcodes/s.gif" width="23" height="25" border="0"></a>
		<a href="#" class="color"><img title="Цвет текста" src="{const=RELATIVE_URL}chat/bbcodes/color.gif" width="23" height="25" border="0"></a>
		<a href="javascript:tag_img()"><img title="Рисунок" src="{const=RELATIVE_URL}chat/bbcodes/image.gif" width="23" height="25" border="0"></a>
		<a href="javascript:tag_url()"><img title="Ссылка" src="{const=RELATIVE_URL}chat/bbcodes/link.gif" width="23" height="25" border="0"></a>
		<a href="#" class="btn-slide"><img title="Смайлы" src="{const=RELATIVE_URL}chat/bbcodes/emo.gif" width="23" height="25" border="0"></a>
		</td>
	</tr>
</table>
</form>

<table class="ntable" id="panel" style="display: none;">
<tr><td>
<?php
	for($i = 1; $i <= 350; $i++)
	{
		if(file_exists(APP_ROOT_DIR."chat/emo/".$i.".gif"))
		{
			echo "<a href=\"javascript:insertsm('".$i."')\"><img src='".RELATIVE_URL."chat/emo/".$i.".gif?".CLIENT_VERSION."' border='0' alt=''></a>";
		}
	}
?>
</td></tr>
</table>

<div id="panel2" style="margin-left: 107px">
<table cellpadding="0" cellspacing="1" border="1">
<tr>
	<td bgcolor="#FFFFFF"><a href="javascript:insertcolor('#FFFFFF')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#FFCCCC"><a href="javascript:insertcolor('#FFCCCC')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#FFCC99"><a href="javascript:insertcolor('#FFCC99')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#FFFF99"><a href="javascript:insertcolor('#FFFF99')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#FFFFCC"><a href="javascript:insertcolor('#FFFFCC')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#99FF99"><a href="javascript:insertcolor('#99FF99')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#99FFFF"><a href="javascript:insertcolor('#99FFFF')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#CCFFFF"><a href="javascript:insertcolor('#CCFFFF')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#CCCCFF"><a href="javascript:insertcolor('#CCCCFF')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#FFCCFF"><a href="javascript:insertcolor('#FFCCFF')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
</tr>
<tr>
	<td bgcolor="#CCCCCC"><a href="javascript:insertcolor('#CCCCCC')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#FF6666"><a href="javascript:insertcolor('#FF6666')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#FF9966"><a href="javascript:insertcolor('#FF9966')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#FFFF66"><a href="javascript:insertcolor('#FFFF66')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#FFFF33"><a href="javascript:insertcolor('#FFFF33')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#66FF99"><a href="javascript:insertcolor('#66FF99')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#33FFFF"><a href="javascript:insertcolor('#33FFFF')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#66FFFF"><a href="javascript:insertcolor('#66FFFF')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#9999FF"><a href="javascript:insertcolor('#9999FF')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#FF99FF"><a href="javascript:insertcolor('#FF99FF')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
</tr>
<tr>
	<td bgcolor="#C0C0C0"><a href="javascript:insertcolor('#C0C0C0')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#FF0000"><a href="javascript:insertcolor('#FF0000')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#FF9900"><a href="javascript:insertcolor('#FF9900')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#FFCC66"><a href="javascript:insertcolor('#FFCC66')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#FFFF00"><a href="javascript:insertcolor('#FFFF00')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#33FF33"><a href="javascript:insertcolor('#33FF33')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#66CCCC"><a href="javascript:insertcolor('#66CCCC')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#33CCFF"><a href="javascript:insertcolor('#33CCFF')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#6666CC"><a href="javascript:insertcolor('#6666CC')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#CC66CC"><a href="javascript:insertcolor('#CC66CC')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
</tr>
<tr>
	<td bgcolor="#999999"><a href="javascript:insertcolor('#999999')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#CC0000"><a href="javascript:insertcolor('#CC0000')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#FF6600"><a href="javascript:insertcolor('#FF6600')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#FFCC33"><a href="javascript:insertcolor('#FFCC33')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#FFCC00"><a href="javascript:insertcolor('#FFCC00')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#33CC00"><a href="javascript:insertcolor('#33CC00')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#00CCCC"><a href="javascript:insertcolor('#00CCCC')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#3366FF"><a href="javascript:insertcolor('#3366FF')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#6633FF"><a href="javascript:insertcolor('#6633FF')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#CC33CC"><a href="javascript:insertcolor('#CC33CC')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
</tr>
<tr>
	<td bgcolor="#666666"><a href="javascript:insertcolor('#666666')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#990000"><a href="javascript:insertcolor('#990000')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#CC6600"><a href="javascript:insertcolor('#CC6600')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#CC9933"><a href="javascript:insertcolor('#CC9933')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#999900"><a href="javascript:insertcolor('#999900')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#009900"><a href="javascript:insertcolor('#009900')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#339999"><a href="javascript:insertcolor('#339999')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#3333FF"><a href="javascript:insertcolor('#3333FF')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#6600CC"><a href="javascript:insertcolor('#6600CC')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#993399"><a href="javascript:insertcolor('#993399')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
</tr>
<tr>
	<td bgcolor="#333333"><a href="javascript:insertcolor('#333333')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#660000"><a href="javascript:insertcolor('#660000')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#993300"><a href="javascript:insertcolor('#993300')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#996633"><a href="javascript:insertcolor('#996633')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#666600"><a href="javascript:insertcolor('#666600')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#006600"><a href="javascript:insertcolor('#006600')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#336666"><a href="javascript:insertcolor('#336666')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#000099"><a href="javascript:insertcolor('#000099')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#333399"><a href="javascript:insertcolor('#333399')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#663366"><a href="javascript:insertcolor('#663366')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
</tr>
<tr>
	<td bgcolor="#000000"><a href="javascript:insertcolor('#000000')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#330000"><a href="javascript:insertcolor('#330000')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#663300"><a href="javascript:insertcolor('#663300')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#663333"><a href="javascript:insertcolor('#663333')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#333300"><a href="javascript:insertcolor('#333300')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#003300"><a href="javascript:insertcolor('#003300')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#003333"><a href="javascript:insertcolor('#003333')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#000066"><a href="javascript:insertcolor('#000066')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#330099"><a href="javascript:insertcolor('#330099')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
	<td bgcolor="#330033"><a href="javascript:insertcolor('#330033')"><img src="{const=RELATIVE_URL}chat/bbcodes/empty.png" border="0"></a></td>
</tr>
</table>
		</div>
<script>

var refresh_timeout = 0;
function chatRefresh()
{
	clearTimeout(refresh_timeout);
	$.ajax({
		url: "<?php echo socialUrl('/chat.php?r=chat/ally');?>",
		dataType: 'json',
		success: function(data){
			$('#chat_div').html(data.chat_html);
			$('#chat_online').html(data.online);
            // updateNews(data.news);
			refresh_timeout = setTimeout(chatRefresh, {const=CHAT_REFRESH_RATE});
		},
		error: function(xhr, textStatus, errorThrown){
			refresh_timeout = setTimeout(chatRefresh, {const=CHAT_REFRESH_RATE});
		}
	});
}

$(function(){
	$('#shoutbox_message').focus();

	chatRefresh();

	$('#chat_form').live('submit',function(){
		var msg = encodeURIComponent( $('#shoutbox_message').val() );
		{if[0]}msg = msg.replace(new RegExp("&",'g'),"amp;");{/if}
		var p_data = 'msg=' + msg;
		$.ajax({
			data: p_data,
			type: "POST",
			url: "<?php echo socialUrl('/chat.php?r=chat/sendAlly'); ?>",
			success: function(data){
				$('#shoutbox_message').val('');
				chatRefresh();
			}
		});
		return false;
	});

});

</script>
{if[0]}
 Oxsar https://oxsar-nova.ru
 Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.

{/if}