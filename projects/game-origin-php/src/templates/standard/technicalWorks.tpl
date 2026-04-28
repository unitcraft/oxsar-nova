{if[{var=tech_works}]}
<script type='text/javascript'>
$(function () {
	$('#queueCountDown_tw').countdown({
		until: {@tech_time},
		compact: true,
		onExpiry: function() {}
	});
});
</script>
<div class="info">
	<table class="table_no_background" style="width: 100%;">
	  <tr>
	    <td width="100px"><span id='queueCountDown_tw'></span></td>
	    <td>{@tech_works_name}</td>
	  </tr>
	</table>
</div>
{/if}