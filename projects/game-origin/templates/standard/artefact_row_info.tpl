<?php $row = $this->getLoopRow(); ?>
{if[$row['auto']]}
<tr>
  <td colspan="3" title="Активируется автоматически с учетом задержки">Не имеет фазы ожидания</td>
</tr>
{/if}
{if[$row['lifetime']]}
<tr>
  <td>Время существования</td>
  <td colspan="2">{loop}lifetime{/loop}</td>
</tr>
{/if}
{if[$row['delay']]}
<tr>
  <td>Заряжается</td>
  <td colspan="2">{loop}delay{/loop}</td>
</tr>
{/if}
{if[$row['times'] > 1]}
<tr>
  <td>Можно включить</td>
  <td colspan="2">{loop}times{/loop} {if[$row['times']>=2 && $row['times']<=4]}раза{else}раз{/if}</td>
</tr>
{/if}
{if[$row['duration']]}
<tr>
  <td>Длительность эффекта</td>
  <td colspan="2">{loop}duration{/loop}</td>
</tr>
{/if}
{if[$row['max_active']]}
<tr>
  <td{if[$row['active_count'] >= $row['max_active']]} class='false2'{/if}>Максимум активированных</td>
  <td colspan="2"{if[$row['active_count'] >= $row['max_active']]} class='false'{/if}>{loop}max_active{/loop}</td>
</tr>
{/if}
{if[$row['trophy_chance'] > 0]}
<tr>
  <td>Вероятность захвата</td>
  <td colspan="2">{loop}trophy_chance{/loop}%</td>
</tr>
{/if}
{if[$row['quota'] > 0]}
<tr>
  <td>Распространенность</td>
  <td colspan="2">{loop}quota{/loop}</td>
</tr>
{/if}
