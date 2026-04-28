<table class="ntable">
	<thead><tr>
		<th>{lang}RELEASE{/lang}</th>
		<th>{lang}CHANGES{/lang}</th>
	</tr></thead>
	{if[{var}latestRevision{/var} > NS_REVISION]}{perm[CAN_MODERATE_USER]}<tfoot><tr>
		<td class="pointer" colspan="2" onclick="window.open('{const}RELATIVE_URL{/const}refdir.php?url=http://sourceforge.net/projects/net-assault');"><span class="external">{lang}AVAILABLE_VERSION{/lang} {@latestVersion}</span></td>
	</tr></tfoot>{/perm}{/if}
	<tbody>
	{hook}showChangeLog{/hook}
	{foreach[release]}<tr>
			<td>{loop}version{/loop}</td>
			<td><pre>{loop}changes{/loop}</pre></td>
		</tr>{/foreach}
	</tbody>
</table>
{if[0]}
 Oxsar https://oxsar-nova.ru
 Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
  
{/if}