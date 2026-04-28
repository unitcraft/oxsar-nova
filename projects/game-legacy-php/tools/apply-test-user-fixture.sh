#!/bin/bash
# План 37.5d.2: применить test-user-snapshot.sql к нашей game-origin БД.
#
# 1. Сохранить global_user_id текущего dev-юзера (userid=1) в /tmp.
# 2. Очистить userid=1 во всех таблицах snapshot.
# 3. Накатить snapshot.
# 4. Восстановить global_user_id (или поставить 'dev-user-001' если не было).
#
# Использование:
#   bash projects/game-legacy-php/tools/apply-test-user-fixture.sh
set -eu

CONT="docker-mysql-1"
DB="oxsar_db"
SNAPSHOT="$(dirname "$0")/../migrations/fixtures/test-user-snapshot.sql"

if [ ! -f "$SNAPSHOT" ]; then
  echo "ERROR: snapshot not found: $SNAPSHOT" >&2
  echo "Run first: bash projects/game-legacy-php/tools/snapshot-legacy-user.sh" >&2
  exit 1
fi

# === Step 1: backup global_user_id ===
echo "Step 1: backup global_user_id..."
GUID=$(docker exec "$CONT" mysql -uroot -proot_pass "$DB" -N -B \
  -e "SELECT IFNULL(global_user_id, '') FROM na_user WHERE userid=1" 2>/dev/null || true)
echo "  saved: [${GUID}]"

# === Step 2: cleanup ===
# Удаляем не только записи с userid=1, но и orphan planetid'ы из migration 002
# (там есть данные test-юзера с userid=NULL — после первого cleanup они
# оставляют PRIMARY KEY конфликт при apply snapshot).
# Hardcoded planetids = те, что test-юзер занимает в legacy (см. snapshot).
echo "Step 2: cleanup..."
TMP_CLEANUP=$(mktemp)
trap "rm -f $TMP_CLEANUP" EXIT
cat > "$TMP_CLEANUP" <<'SQL'
SET FOREIGN_KEY_CHECKS=0;
-- planetids test-юзера в legacy (см. snapshot)
SET @pids = '1,2,501,856,2151,2978,3350,676807,754192';

-- Очистка таблиц по planetid (включая orphan записи из migration 002)
DELETE FROM na_building2planet WHERE FIND_IN_SET(planetid, @pids);
DELETE FROM na_galaxy WHERE FIND_IN_SET(planetid, @pids) OR FIND_IN_SET(moonid, @pids);
DELETE FROM na_unit2shipyard WHERE FIND_IN_SET(planetid, @pids);
DELETE FROM na_temp_fleet WHERE FIND_IN_SET(planetid, @pids);
DELETE FROM na_stargate_jump WHERE FIND_IN_SET(planetid, @pids);
DELETE FROM na_exchange_lots WHERE FIND_IN_SET(planetid, @pids);
DELETE FROM na_planet WHERE FIND_IN_SET(planetid, @pids);

-- Stock — очистка всех активных лотов и брокеров (snapshot v2 заливает заново)
DELETE FROM na_exchange_lots WHERE status IN (1, 5);
DELETE FROM na_exchange;

-- Артефакты на бирже (lot_id > 0) — snapshot v4 заливает заново
DELETE FROM na_artefact2user WHERE lot_id > 0 AND userid != 1;

-- Alliance — расширенные таблицы для нашей alliance (aid=42 у test-юзера)
DELETE FROM na_allyrank WHERE aid IN (SELECT aid FROM na_user2ally WHERE userid=1);
DELETE FROM na_allyapplication WHERE aid IN (SELECT aid FROM na_user2ally WHERE userid=1);
DELETE FROM na_ally_relationships WHERE rel1 IN (SELECT aid FROM na_user2ally WHERE userid=1) OR rel2 IN (SELECT aid FROM na_user2ally WHERE userid=1);
DELETE FROM na_ally_relationships_application WHERE candidate_ally IN (SELECT aid FROM na_user2ally WHERE userid=1) OR request_ally IN (SELECT aid FROM na_user2ally WHERE userid=1);

-- Очистка по userid=1
DELETE FROM na_user WHERE userid=1;
DELETE FROM na_password WHERE userid=1;
DELETE FROM na_user2group WHERE userid=1;
DELETE FROM na_user2ally WHERE userid=1;
DELETE FROM na_research2user WHERE userid=1;
DELETE FROM na_artefact2user WHERE userid=1;
DELETE FROM na_officer WHERE userid=1;
DELETE FROM na_user_experience WHERE userid=1;
DELETE FROM na_credit_bonus_item WHERE userid=1;
DELETE FROM na_res_log WHERE userid=1;
DELETE FROM na_res_transfer WHERE userid=1;

-- Notes (PK user_id)
DELETE FROM na_notes WHERE user_id=1;

-- Messages (sender/receiver)
DELETE FROM na_message WHERE receiver=1 OR sender=1;

-- Также удалить event-record нашего юзера колонизации (от 37.5c)
DELETE FROM na_events WHERE user=1;

SET FOREIGN_KEY_CHECKS=1;
SQL
docker exec -i "$CONT" mysql -uroot -proot_pass "$DB" < "$TMP_CLEANUP"
echo "  cleanup done"

# Останавливаем event-monitor чтобы он не пересоздавал юзера 1 пока мы
# заливаем snapshot (он мог получить EVENT_COLONIZE_NEW_USER_PLANET).
echo "Step 2.5: stop event-monitor..."
docker compose -f projects/game-legacy-php/docker/docker-compose.yml stop event-monitor 2>&1 | tail -1

# === Step 3: apply snapshot ===
echo "Step 3: apply snapshot ($(wc -l < "$SNAPSHOT") lines)..."
# --force: продолжать даже при ошибках (на случай если есть остаточные данные).
# Реальный fail виден будет в Step 5 verification.
docker exec -i "$CONT" mysql --force -uroot -proot_pass "$DB" < "$SNAPSHOT" 2>&1 | grep -v "Warning" | tail -10 || true
echo "  applied"

# === Step 4: restore global_user_id ===
RESTORE_GUID="${GUID:-dev-user-001}"
if [ -z "$RESTORE_GUID" ]; then RESTORE_GUID="dev-user-001"; fi
echo "Step 4: restore global_user_id=[${RESTORE_GUID}]..."
docker exec "$CONT" mysql -uroot -proot_pass "$DB" 2>/dev/null \
  -e "UPDATE na_user SET global_user_id='${RESTORE_GUID}' WHERE userid=1"
echo "  restored"

# === Verification ===
echo ""
echo "Verification:"
docker exec "$CONT" mysql -uroot -proot_pass "$DB" 2>/dev/null -e "
  SELECT 'user' tbl, COUNT(*) n FROM na_user WHERE userid=1
  UNION ALL SELECT 'planets', COUNT(*) FROM na_planet WHERE userid=1
  UNION ALL SELECT 'buildings', COUNT(*) FROM na_building2planet
    WHERE planetid IN (SELECT planetid FROM na_planet WHERE userid=1)
  UNION ALL SELECT 'researches', COUNT(*) FROM na_research2user WHERE userid=1
  UNION ALL SELECT 'artefacts', COUNT(*) FROM na_artefact2user WHERE userid=1;
"

echo ""
echo "Done. dev-login → /game.php/Main теперь рендерит как test-юзер из legacy."
