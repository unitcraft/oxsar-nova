<?php
// План 50 Ф.5 (149-ФЗ): кнопка «Пожаловаться» в legacy-вселенной.
// Контракт совпадает с game-nova/src/components/ReportButton.tsx.
// API живёт на portal-backend (план 56), endpoint POST /api/reports.
// PORTAL_BASE_URL — определена в config/consts.php.
//
// Подключение: один раз в layout.tpl перед </body>. Затем в местах
// рендера никнеймов/чата/альянса вызывается helper getReportButton(),
// который возвращает HTML <button onclick="oxReport.open(...)">.
?>
<style>
.ox-report-btn {
	background: transparent;
	border: 0;
	color: #f48;
	cursor: pointer;
	font-size: 13px;
	padding: 0 4px;
	vertical-align: middle;
	line-height: 1;
}
.ox-report-btn:hover { color: #f00; }
.ox-report-overlay {
	display: none;
	position: fixed;
	inset: 0;
	background: rgba(0,0,0,0.6);
	z-index: 10000;
	align-items: center;
	justify-content: center;
}
.ox-report-overlay.open { display: flex; }
.ox-report-modal {
	background: #1a1a1a;
	border: 1px solid #444;
	border-radius: 6px;
	padding: 20px;
	max-width: 420px;
	width: 90%;
	color: #ddd;
	font-family: inherit;
	box-shadow: 0 4px 24px rgba(0,0,0,0.6);
}
.ox-report-modal h3 {
	margin: 0 0 12px;
	color: #f80;
	font-size: 16px;
}
.ox-report-modal label {
	display: block;
	margin: 8px 0 4px;
	font-size: 12px;
	color: #aaa;
}
.ox-report-modal select,
.ox-report-modal textarea {
	width: 100%;
	box-sizing: border-box;
	background: #2a2a2a;
	color: #ddd;
	border: 1px solid #555;
	padding: 6px 8px;
	font-family: inherit;
	font-size: 13px;
	border-radius: 3px;
}
.ox-report-modal textarea { resize: vertical; min-height: 70px; }
.ox-report-actions {
	display: flex;
	justify-content: flex-end;
	gap: 8px;
	margin-top: 14px;
}
.ox-report-actions button {
	background: #333;
	color: #ddd;
	border: 1px solid #555;
	padding: 6px 14px;
	cursor: pointer;
	border-radius: 3px;
	font-family: inherit;
}
.ox-report-actions button.primary {
	background: #c73;
	border-color: #c73;
	color: #fff;
}
.ox-report-actions button:disabled { opacity: 0.5; cursor: default; }
.ox-report-msg { margin-top: 10px; font-size: 12px; min-height: 14px; }
.ox-report-msg.err { color: #f55; }
.ox-report-msg.ok  { color: #6c6; }
</style>

<div class="ox-report-overlay" id="oxReportOverlay" onclick="oxReport.maybeClose(event)">
	<div class="ox-report-modal" onclick="event.stopPropagation()">
		<h3>Пожаловаться</h3>
		<form id="oxReportForm" onsubmit="return oxReport.submit(event)">
			<input type="hidden" id="oxReportTargetType" />
			<input type="hidden" id="oxReportTargetId" />
			<label for="oxReportReason">Причина</label>
			<select id="oxReportReason" required>
				<option value="profanity">Мат / оскорбления</option>
				<option value="extremism">Экстремизм / разжигание</option>
				<option value="drugs">Наркотики</option>
				<option value="spam">Спам / реклама</option>
				<option value="impersonation">Выдача за другое лицо</option>
				<option value="cheat">Чит / эксплойт</option>
				<option value="other">Другое</option>
			</select>
			<label for="oxReportComment">Комментарий (необязательно, до 1000 символов)</label>
			<textarea id="oxReportComment" maxlength="1000" placeholder="Опишите подробнее, что произошло"></textarea>
			<div class="ox-report-msg" id="oxReportMsg"></div>
			<div class="ox-report-actions">
				<button type="button" onclick="oxReport.close()">Отмена</button>
				<button type="submit" class="primary" id="oxReportSubmit">Отправить</button>
			</div>
		</form>
	</div>
</div>

<script type="text/javascript">
window.oxReport = (function(){
	var ENDPOINT = "<?php echo defined('PORTAL_BASE_URL') ? rtrim(PORTAL_BASE_URL, '/') : ''; ?>" + "/api/reports";
	var JWT = "<?php
		// JWT для cross-origin POST. Cookie oxsar-jwt — HttpOnly +
		// SameSite=Strict (см. public/dev-login.php), браузер не
		// отправит её на portal-backend. Поэтому read'им cookie на
		// сервере и embed'им в JS (ровно для текущего юзера, XSS-риск
		// уже принят — токен уже у юзера в его браузере).
		// json_encode(JSON_HEX_*) экранирует кавычки/слэши/теги;
		// затем strip первый и последний кавычки от json-строки.
		$_t = $_COOKIE['oxsar-jwt'] ?? '';
		echo trim(json_encode($_t, JSON_HEX_QUOT | JSON_HEX_TAG | JSON_HEX_AMP | JSON_HEX_APOS), '"');
	?>";

	function $(id) { return document.getElementById(id); }

	function open(targetType, targetId) {
		$('oxReportTargetType').value = targetType;
		$('oxReportTargetId').value = String(targetId);
		$('oxReportReason').value = 'profanity';
		$('oxReportComment').value = '';
		$('oxReportMsg').textContent = '';
		$('oxReportMsg').className = 'ox-report-msg';
		$('oxReportSubmit').disabled = false;
		$('oxReportOverlay').classList.add('open');
	}

	function close() {
		$('oxReportOverlay').classList.remove('open');
	}

	function maybeClose(ev) {
		if (ev.target && ev.target.id === 'oxReportOverlay') close();
	}

	function submit(ev) {
		ev.preventDefault();
		var msg = $('oxReportMsg');
		var btn = $('oxReportSubmit');
		btn.disabled = true;
		msg.className = 'ox-report-msg';
		msg.textContent = 'Отправка…';

		var headers = { 'Content-Type': 'application/json' };
		if (JWT) headers['Authorization'] = 'Bearer ' + JWT;

		fetch(ENDPOINT, {
			method: 'POST',
			headers: headers,
			body: JSON.stringify({
				target_type: $('oxReportTargetType').value,
				target_id:   $('oxReportTargetId').value,
				reason:      $('oxReportReason').value,
				comment:     $('oxReportComment').value
			})
		}).then(function(res){
			if (res.ok) {
				msg.className = 'ox-report-msg ok';
				msg.textContent = 'Спасибо. Жалоба принята и будет рассмотрена в течение 24 часов.';
				setTimeout(close, 2000);
				return;
			}
			return res.json().catch(function(){ return null; }).then(function(body){
				var detail = (body && body.error && body.error.message) ? body.error.message : ('HTTP ' + res.status);
				msg.className = 'ox-report-msg err';
				msg.textContent = 'Не удалось отправить: ' + detail;
				btn.disabled = false;
			});
		}).catch(function(err){
			msg.className = 'ox-report-msg err';
			msg.textContent = 'Не удалось отправить: ' + (err && err.message ? err.message : 'сетевая ошибка');
			btn.disabled = false;
		});
		return false;
	}

	return { open: open, close: close, maybeClose: maybeClose, submit: submit };
})();
</script>
