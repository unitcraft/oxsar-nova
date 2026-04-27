<script type="text/javascript" src="{const=RELATIVE_URL}chat/chat.js"></script>
<style type="text/css" media="all">@import url({const=RELATIVE_URL}chat/chat.css);</style>

<script type="text/javascript">

var refresh_timeout = 0;
function chatRefresh()
{
	clearTimeout(refresh_timeout);
	$.ajax({
		url: "<?php echo socialUrl('/chat.php?r=chat/index');?>",
		dataType: 'json',
		success: function(data){
			$('#chat_div').html(data.chat_html);
			$('#chat_online').html(data.online);
            updateNews(data.news);
			refresh_timeout = setTimeout(chatRefresh, {const=CHAT_REFRESH_RATE});
		},
		error: function(xhr, textStatus, errorThrown){
			refresh_timeout = setTimeout(chatRefresh, {const=CHAT_REFRESH_RATE});
		}
	});
}

var news_refresh_timeout = 0;
function chatNewsRefresh()
{
	clearTimeout(news_refresh_timeout);
	$.ajax({
		url: "<?php echo socialUrl('/chat.php?r=chat/news');?>",
		dataType: 'json',
		success: function(data){
            updateNews(data.news);
			news_refresh_timeout = setTimeout(chatNewsRefresh, {const=CHAT_NEWS_REFRESH_RATE});
		},
		error: function(xhr, textStatus, errorThrown){
			news_refresh_timeout = setTimeout(chatNewsRefresh, {const=CHAT_NEWS_REFRESH_RATE});
		}
	});
}

var org_chat_news, chat_news_padding = 0, news_change_blocked = true, news_stored = [];
function updateNews(news)
{
    if(news == "org" && org_chat_news !== undefined){
        news = org_chat_news;
    }
    if(news_change_blocked){
        if(news && news_stored.length < 5){
            news_stored.push(news);
        }
        return;
    }
    if(!news && news_stored.length > 0){
        news = news_stored.splice(0, 1);
    }
    if(news){
        news_change_blocked = true;
        var $old_chat_news = $("#chat-news");
        // $old_chat_news.parent().css("height", $old_chat_news.height());

        $new_chat_news = $('<div />')
            .html(news)
            .attr('id', 'chat-news')
            .css({position: 'absolute', padding: chat_news_padding})
            .addClass('ui-helper-hidden-accessible')
            .insertAfter($old_chat_news);

        var bigger_height = true;
        setTimeout(function(){
            if($new_chat_news.height() <= $new_chat_news.parent().height()){
                $new_chat_news
                    .css({
                        left: Math.max(0, ($old_chat_news.parent().width() - $new_chat_news.width())/2),
                        top: bigger_height ? Math.max(0, ($old_chat_news.parent().height() - $new_chat_news.height())/2) : 0,
                        display: 'none'
                    })
                    .removeClass('ui-helper-hidden-accessible')
                    .fadeIn(1000, function(){
                        if(bigger_height){
                            news_change_blocked = false;
                            return;
                        }
                        $new_chat_news.parent().animate({height: $new_chat_news.height()}, 300, null, function(){
                            news_change_blocked = false;
                        });
                    });
            }else{
                $new_chat_news.parent().animate({height: $new_chat_news.height()}, 300, null, function(){
                    $new_chat_news
                        .css({
                            left: Math.max(0, ($old_chat_news.parent().width() - $new_chat_news.width())/2),
                            top: 0,
                            display: 'none'
                        })
                        .removeClass('ui-helper-hidden-accessible')
                        .fadeIn(1000, function(){
                            news_change_blocked = false;
                        });
                });
            }
        }, 0);

        $old_chat_news
            .removeAttr('id')
            .fadeOut(1000, function(){
                $(this).remove();
            });
    }
}

$(function(){
    org_chat_news = $('#chat-news').html();
    setTimeout(function(){
        var $old_chat_news = $("#chat-news");
        $old_chat_news.parent().css({
            position: 'relative',
            width: $old_chat_news.width(),
            height: $old_chat_news.height(),
            padding: 0
        });

        $new_chat_news = $('<div />')
            .attr('id', 'chat-news')
            .css({position: 'absolute', padding: chat_news_padding})
            .insertAfter($old_chat_news);

        $old_chat_news
            .remove()
            .removeAttr('id')
            .appendTo($new_chat_news);

        news_change_blocked = false;
        chatNewsRefresh();
    }, 10000);
});

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
			url: "<?php echo socialUrl('/chat.php?r=chat/send'); ?>",
			success: function(data){
				$('#shoutbox_message').val('');
				chatRefresh();
			}
		});
		return false;
	});

});

</script>

{if[0 && isAdmin()]}
<script type="text/javascript" src="{const=RELATIVE_URL}js/tiny_mce/tiny_mce.js"></script>
<script type="text/javascript">

$(function(){
	tinyMCE.init({
		// General options
		mode : "exact",
		elements : "tinymce_chat_message",
		theme : "advanced",

		skin : "o2k7",
		skin_variant : "black",

		force_p_newlines : false,
		force_br_newlines : true,
		forced_root_block : "",
		paste_create_linebreaks : false,
		// language: "es",
		cleanup_on_startup : true,
		cleanup: true,
		debug : false,
		// file_browser_callback : "dame_contenido",
		auto_focus : "tinymce_chat_message",
		plugins : "emotions,inlinepopups,noneditable,visualchars,xhtmlxtras",

		// Theme options
		theme_advanced_buttons1 :",undo,redo,|,bold,italic,underline,strikethrough,{if[ NS::getUser()->get("points") > 5 ]}{if[ NS::getUser()->get("points") > 20 ]}emotions,charmap,{/if}|,forecolor,{if[ NS::getUser()->get("points") > 100 ]}backcolor,{/if}|,link,unlink,{/if}",
		theme_advanced_buttons2 : "",
		theme_advanced_buttons3 : "",
		theme_advanced_buttons4 : "",
		theme_advanced_toolbar_location : "top",
		theme_advanced_toolbar_align : "left",
		theme_advanced_statusbar_location : "none",
		theme_advanced_buttons1_add_before : "save",

		/*
		plugins : "pagebreak,style,layer,table,save,advhr,advimage,advlink,emotions,iespell,inlinepopups,insertdatetime,preview,media,searchreplace,print,contextmenu,paste,directionality,fullscreen,noneditable,visualchars,nonbreaking,xhtmlxtras,template",

		// Theme options
		theme_advanced_buttons1 : "save,newdocument,|,bold,italic,underline,strikethrough,|,justifyleft,justifycenter,justifyright,justifyfull,|,styleselect,formatselect,fontselect,fontsizeselect",
		theme_advanced_buttons2 : "cut,copy,paste,pastetext,pasteword,|,search,replace,|,bullist,numlist,|,outdent,indent,blockquote,|,undo,redo,|,link,unlink,anchor,image,cleanup,help,code,|,insertdate,inserttime,preview,|,forecolor,backcolor",
		theme_advanced_buttons3 : "tablecontrols,|,hr,removeformat,visualaid,|,sub,sup,|,charmap,emotions,iespell,media,advhr,|,print,|,ltr,rtl,|,fullscreen",
		theme_advanced_buttons4 : "insertlayer,moveforward,movebackward,absolute,|,styleprops,|,cite,abbr,acronym,del,ins,attribs,|,visualchars,nonbreaking,template,pagebreak",
		theme_advanced_toolbar_location : "top",
		theme_advanced_toolbar_align : "left",
		theme_advanced_statusbar_location : "bottom",
		theme_advanced_resizing : true,
		*/

		// Example content CSS (should be your site CSS)
		content_css : "{const=RELATIVE_URL}css/style.css?{const=CLIENT_VERSION}",

		// Drop lists for link/image/media/template dialogs
		/*
		template_external_list_url : "js/template_list.js",
		external_link_list_url : "js/link_list.js",
		external_image_list_url : "js/image_list.js",
		media_external_list_url : "js/media_list.js",
		*/

		setup: function(ed)
			{
			  ed.onKeyPress.add(function(ed, e)
			  {
				  var keyCode = e.keyCode||e.which||e.charCode;
				  if( keyCode == 13 )
				  {
					tinymce_chat_send( ed.getContent() );
					// ajaxpost('chat_insert.php','texto='+ dame_contenido());
					ed.setContent(' ');
					ed.selection.select(ed.getBody());
					ed.selection.collapse(true);
					e.preventDefault();
				  }
			  });
			},
	});
	// tinyMCE.execCommand("mceAddControl", true, "tinymce_chat_control");

	$('#tinymce_chat_form').live('submit',function(){
		tinymce_chat_send();
		return false;
	});

});

function tinymce_chat_send(msg)
{
	if(msg == undefined)
	{
		msg = tinyMCE.get('tinymce_chat_message').getContent();
	}

	msg = encodeURIComponent(msg); // $('#shoutbox_message').val() );
	var p_data = 'msg=' + msg + '&tinymce=1';
	$.ajax({
		data: p_data,
		type: "POST",
		url: "<?php echo socialUrl('/chat.php?r=chat/send'); ?>",
		success: function(data){
			$('#shoutbox_message').val('');
			chatRefresh();
		}
	});
}

</script>
{/if}

<table class="ntable">
	<tr>
		<td>
            <div>
                <div id="chat-news">
{if[!defined('SN')]}
Игровой чат предназначен для общения игроков, <b>помощи новичкам</b> и обсуждения внутриигровых событий. В чате запрещается грубить, использовать ненормативную лексику, провоцировать конфликты, а также
публиковать информацию о найденных ошибках. Если Вы нашли ошибку сообщите о ней Оператору на <a href="http://cakeuniverse.ru/index.php?showforum=42" target="_blank">форуме (раздел Ошибки в игре)</a> или на
почту <a href="mailto:support@oxsar.ru" target="_blank">support@oxsar.ru</a>. Спасибо за соблюдение этих простых правил.
{else}
Игровой чат предназначен для общения игроков, <b>помощи новичкам</b> и обсуждения внутриигровых событий. В чате запрещается грубить, использовать ненормативную лексику, провоцировать конфликты, а также
публиковать информацию о найденных ошибках. Если Вы нашли ошибку сообщите о ней Оператору на
почту <a href="mailto:support@oxsar.ru" target="_blank">support@oxsar.ru</a>. Спасибо за соблюдение этих простых правил.
{/if}
                </div>
            </div>
		</td>
	</tr>
</table>

<table class="ntable">
	<thead><tr>
		<th width="50%"><center id='chat_link_for_tutorial'>{@chat_link}</center></th>
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
	{if[0 && isAdmin()]}
	<tr>
		<td>
			<form method="post" action="{@sendAction}" id='tinymce_chat_form'>
			<table cellspacing="0" cellpadding="0" border="0" class="table_no_background">
				<tr>
					<td width="100%">
						{if[0]}<textarea name="tinymce_chat_message" id="tinymce_chat_message" cols="60" rows="2" style="width:100%">
						</textarea>{/if}
						<input type="text" name="tinymce_chat_message" id="tinymce_chat_message" style="width:100%" size="60" maxlength="500" />
					</td>
					<td width="1px" valign="top" align="center"><input type="submit" class="button" name="send_message" value="Отправить" />
						<p />
						<span>Читают<br /><span class="false" id="chat_online2">~~</span></span>
					</td>
				</tr>
				<tr>
					<td><div id="tinymce_chat_control"></div></td>
				</tr>
			</table>
			</form>
		</td>
	</tr>
	{/if}
	<tr>
		<td>
			<table cellspacing="0" cellpadding="0" border="0" class="table_no_background">
				<tr>
					<form method="post" action="{@sendAction}" id='chat_form'>
					<td width="100%">
					{if[!{var}eROreason{/var}]}
						<input type="text" name="shoutbox_message" id="shoutbox_message" size="56" maxlength="500"
							{if[ !isMobileSkin() ]}style="width:100%"{/if}
							/>
					{if[ !isMobileSkin() ]}
					</td>
					<td>
					{/if}
						<input onfocus="chek()" type="submit" class="button" name="send_message" value="Отправить" />
					{/if}
					{if[{var}eROreason{/var}]}<b>Вы не можете оставлять сообщения в чате.</b> Причина: {@eROreason}{/if}
					</td>
					</form>
					<td nowrap="nowrap"><span style="float:right">Читают: <span class="false" id="chat_online">~~</span></span></td>
				</tr>
			</table>
		</td>
	</tr>
    {if[!{var}eROreason{/var}]}
	<tr>
		<td>
		<a href="javascript:insertbb('b')"><img title="Полужирный текст" src="{const=RELATIVE_URL}chat/bbcodes/b.gif" width="23" height="25" border="0"></a>
		<a href="javascript:insertbb('i')"><img title="Наклонный текст" src="{const=RELATIVE_URL}chat/bbcodes/i.gif" width="23" height="25" border="0"></a>
		<a href="javascript:insertbb('u')"><img title="Подчеркнутый текст" src="{const=RELATIVE_URL}chat/bbcodes/u.gif" width="23" height="25" border="0"></a>
		<a href="javascript:insertbb('s')"><img title="Зачеркнутый текст" src="{const=RELATIVE_URL}chat/bbcodes/s.gif" width="23" height="25" border="0"></a>
		{if[ NS::getUser()->get("points") > 5 ]}
			<a href="#" class="color"><img title="Цвет текста" src="{const=RELATIVE_URL}chat/bbcodes/color.gif" width="23" height="25" border="0"></a>
			{if[0]}<a href="javascript:tag_img()"><img title="Рисунок" src="{const=RELATIVE_URL}chat/bbcodes/image.gif" width="23" height="25" border="0"></a>{/if}
			<a href="javascript:tag_url()"><img title="Ссылка" src="{const=RELATIVE_URL}chat/bbcodes/link.gif" width="23" height="25" border="0"></a>
			{if[ NS::getUser()->get("points") > 20 ]}
				<a href="#" class="btn-slide"><img title="Смайлы" src="{const=RELATIVE_URL}chat/bbcodes/emo.gif" width="23" height="25" border="0"></a>
			{/if}
		{/if}
		</td>
	</tr>
    {/if}
</table>

{if[!{var}eROreason{/var}]}
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
{/if}

{if[0]}
 Oxsar https://oxsar-nova.ru
 Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.

{/if}