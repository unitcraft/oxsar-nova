<form method="post" action="{@formaction}">
<table class="ntable">
	<thead>
	  <tr>
		  <th colspan="2">{lang}STAR_GATE_JUMP{/lang}</th>
	  </tr>
	</thead>
	{if[count($this->getLoop("moons")) > 0]}
	  <tfoot>
	    <tr>
		    <td colspan="2"><input type="submit" name="execjump" value="{lang}EXECUTE_JUMP{/lang}" class="button" /></td>
	    </tr>
	  </tfoot>
	{/if}
	<tbody>
	  <tr>
	    <td>{lang=FLEET}</td>
	    <td>{@fleet}</td>
	  </tr>
	  <tr>
		  <td><label for="moonid">{lang}SELECT_TARGET_MOON{/lang}</label></td>
		  <td>
			  <select name="moonid" id="moonid">{foreach[moons]}<option value="{loop}planetid{/loop}">{loop}planetname{/loop} [{loop}galaxy{/loop}:{loop}system{/loop}:{loop}position{/loop}]</option>{/foreach}</select>
		  </td>
	  </tr>
	  {if[ {var=show_ext_message} ]}
	  <tr>
	    <td colspan="2">Телепортация флота между лунами выполняется без временных затрат. Телепортация обороны, а также юнитов с планеты или на планету, занимает 10 минут.</td>
	  </tr>
	  {/if}
	</tbody>
</table>
</form>
{if[0]}
 Oxsar http://oxsar.ru
 Copyright (c) 2009-2010 UnitPoint <support@unitpoint.ru>
  
{/if}