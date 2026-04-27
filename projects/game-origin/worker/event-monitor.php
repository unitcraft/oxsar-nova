<?php
/**
 * Event-monitor воркер для game-origin (план 37.5c.3).
 *
 * Аналог legacy NewEHMonitorCommand. Работает ~125 сек, обрабатывает
 * игровые события через EventHandler::goThroughEvents(), затем выходит.
 * docker-compose сервис с restart: unless-stopped поднимает заново.
 *
 * Стратегия выбрана сознательно (см. docs/plans/37-game-origin.md §37.5c):
 * - Короткий life-cycle защищает от накопления PHP memory/state-проблем
 *   (legacy скрипт периодически падал по непонятным причинам).
 * - Docker гарантирует один экземпляр сервиса (token-файл из legacy не нужен).
 * - При краше (exit 1 или сегфолт) docker auto-restart за ~1-2 сек.
 *
 * Запуск (вне docker, для отладки):
 *   php worker/event-monitor.php
 */

// Только CLI
if(PHP_SAPI !== 'cli') {
	fwrite(STDERR, "event-monitor must run in CLI mode\n");
	exit(1);
}

// Лимит работы (секунды). Дольше — risk накопления state, короче — лишние
// рестарты. 125 — как в legacy NewEHMonitorCommand.
const WORKER_LIFETIME_SEC = 125;

// Bootstrap (как public/game.php, но с YII_CONSOLE для CLI-режима).
$src = dirname(__FILE__, 2) . '/src/';
require_once($src . 'bd_connect_info.php');

define('INGAME', true);
define('YII_CONSOLE', true); // CLI: не делаем setRequest/setUser, не пытаемся рендерить страницу
$GLOBALS['RUN_YII'] = 0;

require_once($src . 'global.inc.php');
require_once(APP_ROOT_DIR . 'game/Functions.inc.php');

new Core();

$start_time = time();
$end_time = $start_time + WORKER_LIFETIME_SEC;

fwrite(STDOUT, sprintf("[event-monitor] start pid=%d, lifetime=%ds\n", getmypid(), WORKER_LIFETIME_SEC));

$total_processed = 0;
$iterations = 0;

while(time() < $end_time)
{
	$iterations++;

	// Health-check БД: если соединение умерло — выходим, docker перезапустит
	// со свежим коннектом. Лучше быстрый exit, чем висящий процесс с мёртвым
	// connection (последний даёт fatal на каждой попытке query).
	try {
		$ping = Core::getDB()->query("SELECT 1");
		Core::getDB()->free_result($ping);
	} catch(Throwable $e) {
		fwrite(STDERR, "[event-monitor] DB ping failed: " . $e->getMessage() . "\n");
		exit(1);
	}

	// Обработка игровых событий. try/catch вокруг всей итерации — одно
	// битое событие не должно убивать воркер.
	$processed = 0;
	try {
		$eh = new EventHandler();
		$eh->externalMonitor = new stdClass(); // маркер для EventHandler что мы CLI
		$processed = $eh->goThroughEvents(100);
		$total_processed += $processed;
		unset($eh->externalMonitor);
		unset($eh);
	} catch(Throwable $e) {
		fwrite(STDERR, sprintf(
			"[event-monitor] iteration %d failed: %s in %s:%d\n",
			$iterations, $e->getMessage(), $e->getFile(), $e->getLine()
		));
		// Не exit — даём шанс следующей итерации (возможно, проблема в одном
		// конкретном event, который мы уже обработали с processed_error).
		usleep(500_000); // полсекунды паузы перед retry
		continue;
	}

	if($processed > 0) {
		fwrite(STDOUT, sprintf(
			"[event-monitor] iter=%d processed=%d total=%d elapsed=%ds\n",
			$iterations, $processed, $total_processed, time() - $start_time
		));
	}

	// Если событий не было — ждём 50ms (как в legacy NewEHMonitor).
	// Если были — продолжаем без паузы, может ещё накопились.
	if($processed === 0) {
		usleep(50_000);
	}
}

fwrite(STDOUT, sprintf(
	"[event-monitor] exit normally, iterations=%d total_processed=%d uptime=%ds\n",
	$iterations, $total_processed, time() - $start_time
));
exit(0);
