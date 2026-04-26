<form method="post" action="{@formaction}">
<input type="hidden" id="typeid" name="typeid" value="0" />
<table class="ntable">
<tbody>
{if[count($this->getLoop("artefacts")) > 0]}
	<tr>
		<th colspan="3">{lang}ARTEFACT_MARKET{/lang}</th>
	</tr>
	{foreach[artefacts]}<tr>
		<td width="1">{loop}image{/loop}</td>
		<td>
			<div style="width:100%">
        {if[$row['flags']]}<span style="float:right">{loop}flags{/loop}</span>{/if}
				{loop}name{/loop}
			</div>
			<div style="clear:both; font-size:smaller; margin:5px;">
			  {loop}description{/loop}<br />
			  {if[$row['lifetime']]}<i>Время существования: {loop}lifetime{/loop}</i><br />{/if}
			  {if[$row['duration']]}<i>Длительность эффекта: {loop}duration{/loop}</i><br />{/if}
			  {if[$row['delay']]}<i>Перерыв между использованиями: {loop}delay{/loop}</i><br />{/if}
			</div>
      <div>
				<table class="table_no_background" cellspacing="0" cellpadding="0" border="0" title="{lang=REQUIRES}">
				<tr>
					{if[$row["use_resources"]]}
					<td>
						<table class="table_no_background" cellspacing="0" cellpadding="0" border="0">
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
						</table>
					</td>
					{/if}
					{if[$row["use_resources"]&&$row["use_credit"]]}
					<td valign="middle">
						{lang}OR{/lang}
					</td>
					{/if}
					{if[$row["use_credit"]]}
					<td>
						<table class="table_no_background" cellspacing="0" cellpadding="0" border="0">
							{if[$row["credit_required"]]}
							<tr>
								<td>{lang=CREDITS}</td>
								<td class='{if[$row["credit_notavailable"]]}notavailable{else}true{/if}'>{loop=credit_required}</td>
								<td>{if[$row["credit_notavailable"]]}({loop=credit_notavailable}){/if}</td>
							</tr>
							{/if}
							{if[$row["energy_required"]]}
							<tr>
								<td>{lang=ENERGY}</td>
								<td class='{if[$row["energy_notavailable"]]}notavailable{else}true{/if}'>{loop=energy_required}</td>
								<td>{if[$row["energy_notavailable"]]}({loop=energy_notavailable}){/if}</td>
							</tr>
							{/if}
						</table>
					</td>
					{/if}
				</tr>
				</table>
      </div>
		</td>
    <td class="center" width="10%">
      {if[$row['use_resources']&&$row['buy_res']]}
				<input type="submit" name="buy_res" value="{lang}BUY_FOR_RESOURCES{/lang}" class="button" onClick="document.getElementById('typeid').value={loop}id{/loop};" />
      {/if}
      {if[$row['use_credit']&&$row['buy_cred']]}
				<input type="submit" name="buy_cred" value="{lang}BUY_FOR_CREDITS{/lang}" class="button" onClick="document.getElementById('typeid').value={loop}id{/loop};" />
      {/if}
    </td>
	</tr>{/foreach}
	{else}<thead><tr>
		<th colspan="4" class="center">{lang}NO_ARTEFACTS{/lang}</th>
	</tr></thead>
{/if}
</tbody>
</table>
</form>
{if[0]}
 Oxsar http://oxsar.ru
 Copyright (c) 2009-2010 UnitPoint <support@unitpoint.ru>
  
{/if}