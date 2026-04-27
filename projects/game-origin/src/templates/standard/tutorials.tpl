{if[count($this->getLoop("avaliable_tutorials")) > 0]}
<table class="ntable">
	{foreach[avaliable_tutorials]}
		<thead>
			<tr>
				<th>
					{loop=number}
				</th>
				<th>
					{loop=title}
				</th>
			</tr>
		</thead>
			<tbody>
		{foreach2[.states]}
				<tr>
					<td>
						{loop=number}
					</td>
					<td>
						<a id='change_tut_state_{loop=id}_link' href='#'>{loop=title}</a>
						<script type="text/javascript">$(function(){$("#change_tut_state_{loop=id}_link").click(function(){$("#change_tut_state_{loop=id}").submit();return false;})});</script>
						<form action="{loop=formaction}" method="post" id='change_tut_state_{loop=id}'>
							<input type="hidden" name="tutorial_data[new]" value="{loop=id}">
							<input type="hidden" name="tutorial_data[force]" value="1">
						</form>
					</td>
				</tr>
		{/foreach2}
			</tbody>
	{/foreach}
</table>
{/if}
{if[0]}
 Oxsar https://oxsar-nova.ru
 Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
{/if}