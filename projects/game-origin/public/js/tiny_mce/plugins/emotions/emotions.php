<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
	<title>{#emotions_dlg.title}</title>
	<script type="text/javascript" src="../../tiny_mce_popup.js"></script>
	<script type="text/javascript" src="js/emotions.js"></script>
</head>
<body style="display: none" role="application" aria-labelledby="app_title">
<span style="display:none;" id="app_title">{#emotions_dlg.title}</span>
<div align="center">
	<table role="presentation" border="0" cellspacing="0" cellpadding="4">
		<tr>
			<td>
				<?php
				
					$files = (array)glob(dirname(__FILE__)."/img/*.*");
					usort($files, function($a, $b){
						return strnatcasecmp($a, $b);
					});
					foreach($files as $filename)
					{
						$filename = basename($filename);
						echo "<a href=\"javascript:EmotionsDialog.insert('{$filename}','');\"><img src='img/{$filename}' border='0' alt='' title='' /></a>";
					}
					
				?>
			</td>
		</tr>
	</table>
</div>
</body>
</html>
