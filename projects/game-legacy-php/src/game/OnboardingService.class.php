<?php
/**
 * Onboarding для нового юзера: запланировать создание стартовой планеты.
 *
 * План 37.5c: triггер EVENT_COLONIZE_NEW_USER_PLANET (как в legacy
 * BaseWebUser::checkAndCreateHomePlanet, но без Yii). Само создание
 * планеты выполняет EventHandler::colonize() через event-monitor воркер.
 */

if(!defined("APP_ROOT_DIR")) { die("Hacking attempt detected."); }

class OnboardingService
{
	/**
	 * Если у юзера ещё нет home planet (na_user.hp IS NULL) и нет уже
	 * запланированного события колонизации — вставить event
	 * EVENT_COLONIZE_NEW_USER_PLANET.
	 *
	 * Идемпотентно: безопасно вызывать на каждом HTTP-запросе.
	 *
	 * @param int $userid
	 * @return bool true если event был вставлен, false если ничего не сделано
	 */
	public static function ensureColonizationScheduled($userid)
	{
		$userid = (int)$userid;
		if($userid <= 0) {
			return false;
		}

		// Проверяем что у юзера ещё нет home planet
		$hp = sqlSelectField("user", "hp", "", "userid = ".sqlVal($userid));
		if($hp) {
			return false;
		}

		// Защита от race condition при параллельных запросах:
		// MySQL advisory lock с timeout=0 (legacy использовал Yii cache->add).
		// Если кто-то другой уже вошёл в эту секцию — выходим без действий.
		$lock_name = "oxsar_colonize:".$userid;
		$db = Core::getDB();
		$lock_result = $db->query("SELECT GET_LOCK(".sqlVal($lock_name).", 0) AS got");
		$lock_row = $db->fetch($lock_result);
		$db->free_result($lock_result);
		if(empty($lock_row["got"])) {
			return false;
		}

		try {
			// Повторная проверка под локом — может другой процесс уже вставил event
			$existing = sqlSelectField("events", "count(*)", "",
				"user = ".sqlVal($userid)
				." AND processed = ".EVENT_PROCESSED_WAIT
				." AND mode = ".EVENT_COLONIZE_NEW_USER_PLANET);
			if($existing > 0) {
				return false;
			}

			$mission_time = COLONIZE_NEW_USER_PLANET_TIME
				+ mt_rand(0, COLONIZE_NEW_USER_PLANET_TIME_MAX_DELTA);

			$data = array(
				"time" => $mission_time,
			);

			// Колониальный корабль (как в legacy)
			$ship_data = sqlSelectRow("construction", array("name", "mode"), "",
				"buildingid = ".sqlVal(UNIT_COLONY_SHIP));
			$data["ships"][UNIT_COLONY_SHIP] = array(
				"id"            => UNIT_COLONY_SHIP,
				"quantity"      => 1,
				"damaged"       => 0,
				"shell_percent" => 100,
				"name"          => $ship_data["name"],
				"mode"          => $ship_data["mode"],
			);

			sqlInsert("events", array(
				"mode"          => EVENT_COLONIZE_NEW_USER_PLANET,
				"start"         => time(),
				"time"          => time() + $mission_time,
				"planetid"      => null,
				"user"          => $userid,
				"destination"   => null,
				"data"          => serialize($data),
				"protected"     => 0,
			));

			return true;
		} finally {
			$rel = $db->query("SELECT RELEASE_LOCK(".sqlVal($lock_name).")");
			$db->free_result($rel);
		}
	}
}
