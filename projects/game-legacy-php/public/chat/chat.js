/*
function update_Div(){
   $.get('chat.php', function(data){
        $("#myDiv").empty();
        $("#myDiv").append(data);
   });
}
$(document).ready(function(){
    setInterval("update_Div()", 10000);
});
*/
$(document).ready(function(){
    $(".btn-slide").click(function(){
        $("#panel").slideToggle("slow");
        $(this).toggleClass("active");
    });
});

$(document).ready(function(){
    $(".color").click(function(){
        $("#panel2").slideToggle("slow2");
        $(this).toggleClass("active2");
    });
});

var b_open = 0;
var i_open = 0;
var u_open = 0;
var s_open = 0;
var color_open = 0;

function insertbb(bb)
{
	var tagOpen = eval(bb + "_open");
	if(tagOpen == 0)
	{
		eval(bb + "_open = 1");
		document.getElementById('shoutbox_message').focus();
		document.getElementById('shoutbox_message').value += '['+bb+']';
	}else{
		eval(bb + "_open = 0");
		document.getElementById('shoutbox_message').focus();
		document.getElementById('shoutbox_message').value += '[/'+bb+']';
	}
}

function insertcolor(bb)
{
	if(color_open == 0)
	{
		eval("color_open = 1");
		document.getElementById('shoutbox_message').focus();
		document.getElementById('shoutbox_message').value += '[color='+bb+']';
	}else{
		eval("color_open = 0");
		document.getElementById('shoutbox_message').focus();
		document.getElementById('shoutbox_message').value += '[/color]';
	}
}

function insertsm(sm)
{
	document.getElementById('shoutbox_message').focus();
	document.getElementById('shoutbox_message').value += '[:'+sm+':]';
}

function chek()
{
	if(b_open == 1)
	{
		eval("b_open = 0");
		document.getElementById('shoutbox_message').focus();
		document.getElementById('shoutbox_message').value += '[/b]';
	}
	if(i_open == 1)
	{
		eval("b_open = 0");
		document.getElementById('shoutbox_message').focus();
		document.getElementById('shoutbox_message').value += '[/i]';
	}
	if(u_open == 1)
	{
		eval("b_open = 0");
		document.getElementById('shoutbox_message').focus();
		document.getElementById('shoutbox_message').value += '[/u]';
	}
	if(s_open == 1)
	{
		eval("b_open = 0");
		document.getElementById('shoutbox_message').focus();
		document.getElementById('shoutbox_message').value += '[/s]';
	}
	if(color_open == 1)
	{
		eval("color_open = 0");
		document.getElementById('shoutbox_message').focus();
		document.getElementById('shoutbox_message').value += '[/color]';
	}
}

function insertname(dat)
{
	document.getElementById('shoutbox_message').focus();
	document.getElementById('shoutbox_message').value += '[b]'+dat+'[/b], ';
}

function tag_img()
{
	var FoundErrors = '';
	var enterURL = prompt("Введите адресс картинки", "");

	if(!enterURL) FoundErrors += " " + "Не ввели адрес картинки";

	if(FoundErrors)
	{
		alert("Error!"+FoundErrors);
		return;
	}
	document.getElementById('shoutbox_message').focus();
	document.getElementById('shoutbox_message').value += '[img]'+enterURL+'[/img]';
}

function tag_url()
{
	var FoundErrors = '';
	var enterURL   = prompt("Введите адресс сайта", "");
	var enterTITLE = prompt("Введите название(не обязательно)", "");

	if (!enterURL) FoundErrors += " " + "Не ввели адрес сайта";

	if (FoundErrors)
	{
		alert("Error!"+FoundErrors);
		return;
	}

	if (!enterTITLE)
	{
		document.getElementById('shoutbox_message').focus();
		document.getElementById('shoutbox_message').value += '[url]'+enterURL+'[/url]';
	} else {
		document.getElementById('shoutbox_message').focus();
		document.getElementById('shoutbox_message').value += '[url='+enterURL+']'+enterTITLE+'[/url]';
	}
}