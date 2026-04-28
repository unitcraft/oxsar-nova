function repaint()
{
    return false;
    $('.sm2-360btn').each(
        function()
            {$(this).attr('src', playButtonImg);}
    );
}
function selectTrack(id)
{
    $('#track').val(id);
    $('div.ui360 a').each(
        function(){
            $(this).removeClass('true');
            if (this.id == id)
                {$(this).addClass('true');}
        }
    );
    generateFriendLink();
}

function generateFriendLink()
{
    var fl = base_url;
    var bg_style = $('#user_bg_style').val();
    var table_style = $('#user_table_style').val();
    var track = $('#track').val();
    if (bg_style != '')
    {fl += '&bg=' + encodeURI(bg_style);}
    if (table_style != '')
    {fl += '&tb=' + encodeURI(table_style);}
    if (track != '')
    {fl += '&track=' + encodeURI(track);}
    $('#friendlink').attr('href', fl);
    $('#friendlink').html(fl);
    $('#friendlink2').attr('href', fl);
    $('#friendlink2').html(fl);

}

function SoHform()
{
    var animation_speed = 250;
    if ($('#sendHeader').css('cursor') == 's-resize')
    {
        $('#sendHeader').css('cursor', 'n-resize');
        $('#sendTable').show('blind', '', animation_speed, repaint);
    }
    else
    {
        $('#sendTable').hide('blind', '', animation_speed);
        $('#sendHeader').css('cursor', 's-resize');
    }
}

function sendReport(base_url)
{
    var data = 'email=' + encodeURI($('#email').val());
    data += '&id=' + encodeURI($('#id').val());
    data += '&key=' + encodeURI($('#key').val());
    data += '&key2=' + encodeURI($('#key2').val());
    data += '&bg=' + encodeURI($('#user_bg_style').val());
    data += '&tb=' + encodeURI($('#user_table_style').val());
    data += '&text=' + encodeURI($('#usertext').val());
    data += '&track=' + encodeURI($('#track').val());
    $.ajax({
        type: 'POST',
        url: (base_url + $('#odnoklassniki').val()),
        data: data,
        success: function(html){
            $('#sendreport').css('display', '').html(html);
        }
    }
    );
    $('#go').css('visibility', 'hidden');
    SoHform();
    setTimeout(function(){$('#go').css('visibility', 'visible');}, 5000);
    document.documentElement.scrollTop = 0;
}

function setActiveStyleSheet(href, type) {
		var i, a;
		for(i=0; (a = document.getElementsByTagName('link')[i]); i++) {
			if(a.getAttribute('rel').indexOf('style') != -1 && a.getAttribute('href').indexOf('/css/us_' + type + '/') != -1) {
				a.disabled = true;
				if(href != '' && a.getAttribute('href').indexOf(href) != -1) {
                                    a.disabled = false;
				}
			}
		}
            }

$(document).ready(function(){
    setActiveStyleSheet(current_bg_style, 'bg');
    setActiveStyleSheet(current_table_style, 'table');
    generateFriendLink();
});