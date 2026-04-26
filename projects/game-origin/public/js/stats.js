/**
 * Oxsar http://oxsar.ru
 *
 * 
 */

function setOrder(field)
{
    page_select = document.getElementById('page');
    page_select.value = 1;

    so_input = document.getElementById('sort_order');
    sf_input = document.getElementById('sort_field');

    if(sf_input.value != field)
    {
      sf_input.value = field;
      so_input.value = 'desc';
    }
    else
    {
      so_input.value = so_input.value == 'asc' ? 'desc' : 'asc';
    }

    document.getElementById('go').click();
}

function goPage(page)
{
    document.getElementById('page').value = page;
    document.getElementById('go').click();
}

function goWhere(whereid, id)
{
    document.getElementById('whereid').value = whereid;
    document.getElementById('id').value = id;
    document.getElementById('go').click();
}