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
 Oxsar https://oxsar-nova.ru
 Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
{/if}