<table width="100%" class="table_no_background" cellspacing="0" cellpadding="0" border="0">
  <tr>
  	{if[!defined('SN')]}
    <td style="padding-left:0;padding-right:0">
      <form method="post" action="{@formaction}" name="tform" id="tform" style="padding:0; margin:0">
      <b>{lang}IMAGE_PACKAGE{/lang}</b>
      <select name="image_package" id="image_package" onchange="document.forms['tform'].submit()">
        {foreach[imagePacks]}<option value="{loop}dir{/loop}"{if[$row["dir"] == Core::getUser()->get("imagepackage")]} selected="selected"{/if}>{loop}name{/loop}</option>{/foreach}
      </select>
      </form>
    </td>
    {/if}
   
    <td align="right" style="padding-left:0;padding-right:0">
      <form method="post" action="{@formaction}" name="favailable" id="favailable" style="padding:0; margin:0">
      <label for="show_all_units_checkbox"><b>{lang}SHOW_UNAVAILABLE{/lang}</b></label>
      <input type="checkbox" name="show_all_units_checkbox" id="show_all_units_checkbox" value="1" onchange="submitShowAllUnits()" {if[$this->templateVars["show_all_units"]]} checked="checked"{/if}/>
      <input type="hidden" name="show_all_units" id="show_all_units" value="{@show_all_units}">
      </form>
      <script type="text/javascript">
        function submitShowAllUnits()
        {
          cb = document.getElementById('show_all_units_checkbox');
          su = document.getElementById('show_all_units');
          su.value = cb.checked ? '1' : '0';
          document.getElementById('favailable').submit();
        }
      </script>
    </td>
  </tr>
</table>
{if[0]}
 Oxsar http://oxsar.ru
 Copyright (c) 2009-2010 UnitPoint <support@unitpoint.ru>
  
{/if}