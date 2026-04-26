{if[ACHIEVEMENTS_ENABLED]}{include}"achievements"{/include}{/if}
<?php if( 0 && !defined('SN') && rand(1, 5) == 1 ) Yii::app()->controller->widget('BestAssaults'); ?>
{if[0 && {var=usertip}]}
<table class="ntable">
  <tr>
    <th>{lang}Помощь новичку{/lang}</th>
  </tr>
  <tr>
    <td>
    	{@usertip}
    </td>
  </tr>
</table>
{/if}
{if[0]}
 Oxsar http://oxsar.ru
 Copyright (c) 2009-2010 UnitPoint <support@unitpoint.ru>
{/if}