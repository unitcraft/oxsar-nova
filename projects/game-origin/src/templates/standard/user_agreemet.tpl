<form method="post" action="{@formaction}">
<table class="ntable">
	<thead>
		<tr>
			<th>{lang=USER_AGREEMENT}</th>
		</tr>
		{if[{var=new_agreemet_pre_text}]}
			<tr>
				<td style="text-align:justify;">{@new_agreemet_pre_text}</td>
			</tr>
		{/if}
		{if[{var=agreemet_pre_text}]}
			<tr>
				<td style="text-align:justify;">{@agreemet_pre_text}</td>
			</tr>
		{/if}
	</thead>
	
	<tfoot>
		{if[{var=agreemet_post_text}]}
			<tr>
				<td style="text-align:justify;">{@agreemet_post_text}</td>
			</tr>
		{/if}
		<tr>
			<td class="center"><input type="submit" name="agree" value="{lang=OK}" class="button"></td>
		</tr>
	</tfoot>
	
	{if[count($this->getLoop("agreemet")) > 0]}
		{foreach[agreemet]}
			<tbody>
			<tr><td>
				<div style="text-align: justify">
					{if[$row['depth_str']]}<b>{loop=depth_str}</b>{/if}
					{if[$row['is_new']]}<span class="false2">{/if}{loop=text}{if[$row['is_new']]}</span>{/if}
				</div>
			</td></tr>
				{if[count($this->getLoop(".childs")) > 0]}
					{foreach2[.childs]}
						<tr><td>
							<div style="text-align: justify{if[$row['logic_depth'] > 0]}; margin-left: <?php echo round(30*$row['logic_depth']); ?>px{/if}">
								{if[$row['depth_str']]}<b>{loop=depth_str}</b>{/if}
								{if[$row['is_new']]}<span class="false2">{/if}{loop=text}{if[$row['is_new']]}</span>{/if}
							</div>
						</td></tr>
						{if[count($this->getLoop("..childs")) > 0]}
							{foreach3[..childs]}
								<tr><td>
									<div style="text-align: justify{if[$row['logic_depth'] > 0]}; margin-left: <?php echo round(30*$row['logic_depth']); ?>px{/if}">
										{if[$row['depth_str']]}<b>{loop=depth_str}</b>{/if}
										{if[$row['is_new']]}<span class="false2">{/if}{loop=text}{if[$row['is_new']]}</span>{/if}
									</div>
								</td></tr>
							{/foreach3}
						{/if}
					{/foreach2}
				{/if}
			</tbody>
		{/foreach}
	{/if}
</table>
</form>
{if[0]}
 Oxsar https://oxsar-nova.ru
 Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
 
{/if}