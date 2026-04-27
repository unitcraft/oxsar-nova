<?php $row = $this->getLoopRow(); ?>
<table class="table_no_background" cellspacing="0" cellpadding="0" border="0" title="{lang=REQUIRES}">
  {if[$row["metal_required"]]}
  <tr>
    <td>{lang=METAL}</td>
    <td class='{if[$row["metal_notavailable"]]}notavailable{else}true{/if}'>{loop=metal_required}</td>
    <td>{if[$row["metal_notavailable"]]}({loop=metal_notavailable}){/if}</td>
  </tr>
  {/if}
  {if[$row["silicon_required"]]}
  <tr>
    <td>{lang=SILICON}</td>
    <td class='{if[$row["silicon_notavailable"]]}notavailable{else}true{/if}'>{loop=silicon_required}</td>
    <td>{if[$row["silicon_notavailable"]]}({loop=silicon_notavailable}){/if}</td>
  </tr>
  {/if}
  {if[$row["hydrogen_required"]]}
  <tr>
    <td>{lang=HYDROGEN}</td>
    <td class='{if[$row["hydrogen_notavailable"]]}notavailable{else}true{/if}'>{loop=hydrogen_required}</td>
    <td>{if[$row["hydrogen_notavailable"]]}({loop=hydrogen_notavailable}){/if}</td>
  </tr>
  {/if}
  {if[$row["energy_required"]]}
  <tr>
    <td>{lang=ENERGY}</td>
    <td class='{if[$row["energy_notavailable"]]}notavailable{else}true{/if}'>{loop=energy_required}</td>
    <td>{if[$row["energy_notavailable"]]}({loop=energy_notavailable}){/if}</td>
  </tr>
  {/if}
  {if[$row["credit_required"]]}
  <tr>
    <td>{lang=CREDITS}</td>
    <td class='{if[$row["credit_notavailable"]]}notavailable{else}true{/if}'>{loop=credit_required}</td>
    <td>{if[$row["credit_notavailable"]]}({loop=credit_notavailable}){/if}</td>
  </tr>
  {/if}
  {if[$row["points_required"]]}
  <tr>
    <td>{lang=POINTS}</td>
    <td class='{if[$row["points_notavailable"]]}notavailable{else}true{/if}'>{loop=points_required}</td>
    <td>{if[$row["points_notavailable"]]}({loop=points_notavailable}){/if}</td>
  </tr>
  {/if}
  <tr>
    <td>{lang=REQUIRE_TIME}</td>
    <td colspan="2">{loop=productiontime}</td>
  </tr>
</table>
{if[0]}
 Oxsar https://oxsar-nova.ru
 Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
  
{/if}