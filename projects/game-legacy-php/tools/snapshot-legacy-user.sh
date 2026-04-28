#!/bin/bash
# План 37.5d.1: снять снимок test-юзера (userid=1) из боевой legacy БД
# для использования как dev-fixture в game-origin (UI-сравнение).
#
# Запускать ДО старта game-origin:
#   bash projects/game-legacy-php/tools/snapshot-legacy-user.sh
#
# Результат:
#   projects/game-legacy-php/migrations/fixtures/test-user-snapshot.sql

set -euo pipefail

LEGACY_CONTAINER="oxsar2-mysql-1"
LEGACY_USER="root"
LEGACY_PASS="root"
LEGACY_DB="oxsar_db"
USERID="${1:-1}"

OUT="$(dirname "$0")/../migrations/fixtures/test-user-snapshot.sql"
mkdir -p "$(dirname "$OUT")"

echo "Snapshotting userid=${USERID} from ${LEGACY_CONTAINER}..." >&2

# Получить список planetid юзера (для taблиц связанных через planet)
PLANETIDS=$(docker exec "$LEGACY_CONTAINER" mysql -u"$LEGACY_USER" -p"$LEGACY_PASS" "$LEGACY_DB" \
  -N -B -e "SELECT GROUP_CONCAT(planetid) FROM na_planet WHERE userid=$USERID" 2>/dev/null)

if [ -z "$PLANETIDS" ]; then
  echo "ERROR: no planets for userid=$USERID" >&2
  exit 1
fi
echo "  planets: $PLANETIDS" >&2

# Header в выходной файл
cat > "$OUT" <<EOF
-- Snapshot test-юзера (userid=${USERID}) из legacy oxsar2 для UI-сравнения.
-- Сгенерировано $(date -u +%Y-%m-%dT%H:%M:%SZ) скриптом snapshot-legacy-user.sh.
-- Применять через apply-test-user-fixture.sh — он сохраняет global_user_id.
SET FOREIGN_KEY_CHECKS=0;
SET unique_checks=0;
SET autocommit=0;

EOF

# Группа 1: таблицы где у юзера есть колонка userid (фильтр WHERE userid=N).
# Исключено: na_event_src/na_event_dest — это VIEW, dump для них бессмыслен
# (содержимое генерируется из na_events).
USER_TABLES=(
  na_user na_password na_user2group na_user2ally
  na_planet na_research2user na_artefact2user na_officer
  na_user_experience na_credit_bonus_item
  na_res_log na_res_transfer
)

# Group 1b: таблицы по user_id (с подчёркиванием)
USER_ID_TABLES=(
  na_notes
)

for tbl in "${USER_TABLES[@]}"; do
  echo "  dump $tbl WHERE userid=$USERID" >&2
  echo "-- $tbl" >> "$OUT"
  docker exec "$LEGACY_CONTAINER" mysqldump -u"$LEGACY_USER" -p"$LEGACY_PASS" "$LEGACY_DB" \
    --no-create-info --skip-extended-insert --skip-comments --hex-blob --complete-insert \
    --where="userid=$USERID" "$tbl" 2>/dev/null >> "$OUT" || {
      echo "  WARN: dump failed for $tbl, skipping" >&2
    }
  echo "" >> "$OUT"
done

for tbl in "${USER_ID_TABLES[@]}"; do
  echo "  dump $tbl WHERE user_id=$USERID" >&2
  echo "-- $tbl" >> "$OUT"
  docker exec "$LEGACY_CONTAINER" mysqldump -u"$LEGACY_USER" -p"$LEGACY_PASS" "$LEGACY_DB" \
    --no-create-info --skip-extended-insert --skip-comments --hex-blob --complete-insert \
    --where="user_id=$USERID" "$tbl" 2>/dev/null >> "$OUT" || true
  echo "" >> "$OUT"
done

# na_message — фильтр по receiver (нашему юзеру) + sender
echo "  dump na_message WHERE receiver=$USERID OR sender=$USERID" >&2
echo "-- na_message" >> "$OUT"
docker exec "$LEGACY_CONTAINER" mysqldump -u"$LEGACY_USER" -p"$LEGACY_PASS" "$LEGACY_DB" \
  --no-create-info --skip-extended-insert --skip-comments --hex-blob --complete-insert \
  --where="receiver=$USERID OR sender=$USERID" na_message 2>/dev/null >> "$OUT" || true
echo "" >> "$OUT"

# Группа 2: таблицы где фильтр по planetid IN (...). Только planetid —
# moonid есть только у na_galaxy (см. отдельную обработку ниже).
PLANET_TABLES=(
  na_building2planet na_unit2shipyard
  na_temp_fleet na_stargate_jump na_exchange_lots
)

for tbl in "${PLANET_TABLES[@]}"; do
  echo "  dump $tbl WHERE planetid IN (...)" >&2
  echo "-- $tbl" >> "$OUT"
  docker exec "$LEGACY_CONTAINER" mysqldump -u"$LEGACY_USER" -p"$LEGACY_PASS" "$LEGACY_DB" \
    --no-create-info --skip-extended-insert --skip-comments --hex-blob --complete-insert \
    --where="planetid IN ($PLANETIDS)" "$tbl" 2>/dev/null >> "$OUT" || true
  echo "" >> "$OUT"
done

# na_galaxy — отдельно с фильтром по planetid OR moonid
echo "  dump na_galaxy WHERE planetid OR moonid IN (...)" >&2
echo "-- na_galaxy" >> "$OUT"
docker exec "$LEGACY_CONTAINER" mysqldump -u"$LEGACY_USER" -p"$LEGACY_PASS" "$LEGACY_DB" \
  --no-create-info --skip-extended-insert --skip-comments --hex-blob --complete-insert \
  --where="planetid IN ($PLANETIDS) OR moonid IN ($PLANETIDS)" na_galaxy 2>/dev/null >> "$OUT" || true
echo "" >> "$OUT"

# Группа 3: связанные таблицы (alliance — нужны другие участники для рендера)
ALLY_AID=$(docker exec "$LEGACY_CONTAINER" mysql -u"$LEGACY_USER" -p"$LEGACY_PASS" "$LEGACY_DB" \
  -N -B -e "SELECT aid FROM na_user2ally WHERE userid=$USERID LIMIT 1" 2>/dev/null || echo "")

if [ -n "$ALLY_AID" ]; then
  echo "  dump na_alliance WHERE aid=$ALLY_AID" >&2
  echo "-- na_alliance" >> "$OUT"
  docker exec "$LEGACY_CONTAINER" mysqldump -u"$LEGACY_USER" -p"$LEGACY_PASS" "$LEGACY_DB" \
    --no-create-info --skip-extended-insert --skip-comments --hex-blob --complete-insert \
    --where="aid=$ALLY_AID" na_alliance 2>/dev/null >> "$OUT" || true

  # Другие участники той же alliance — нужны их user-записи для отрисовки списка
  ALLY_MEMBERS=$(docker exec "$LEGACY_CONTAINER" mysql -u"$LEGACY_USER" -p"$LEGACY_PASS" "$LEGACY_DB" \
    -N -B -e "SELECT GROUP_CONCAT(userid) FROM na_user2ally WHERE aid=$ALLY_AID" 2>/dev/null)

  if [ -n "$ALLY_MEMBERS" ] && [ "$ALLY_MEMBERS" != "$USERID" ]; then
    echo "  dump na_user (alliance members: $ALLY_MEMBERS)" >&2
    echo "-- alliance members (na_user)" >> "$OUT"
    docker exec "$LEGACY_CONTAINER" mysqldump -u"$LEGACY_USER" -p"$LEGACY_PASS" "$LEGACY_DB" \
      --no-create-info --skip-extended-insert --skip-comments --hex-blob --complete-insert \
      --where="userid IN ($ALLY_MEMBERS) AND userid != $USERID" na_user 2>/dev/null >> "$OUT" || true

    echo "-- alliance members (na_user2ally)" >> "$OUT"
    docker exec "$LEGACY_CONTAINER" mysqldump -u"$LEGACY_USER" -p"$LEGACY_PASS" "$LEGACY_DB" \
      --no-create-info --skip-extended-insert --skip-comments --hex-blob --complete-insert \
      --where="aid=$ALLY_AID AND userid != $USERID" na_user2ally 2>/dev/null >> "$OUT" || true
  fi
  echo "" >> "$OUT"
fi

# Группа 3.4: Артефакты на бирже (lot_id > 0). Stock рендерит type=3 лоты
# (ETYPE_ARTEFACT) и для них дополнительно проверяет na_artefact2user
# через artid из data blob. Без этих записей — silent fail в Stock loop.
# Также нужны na_events для lifetime/expire/delay этих артефактов.
echo "  dump na_artefact2user (artefacts on auction, lot_id > 0)" >&2
echo "-- na_artefact2user (auction artefacts)" >> "$OUT"
docker exec "$LEGACY_CONTAINER" mysqldump -u"$LEGACY_USER" -p"$LEGACY_PASS" "$LEGACY_DB" \
  --no-create-info --skip-extended-insert --skip-comments --hex-blob --complete-insert \
  --where="lot_id > 0 AND userid != $USERID" na_artefact2user 2>/dev/null >> "$OUT" || true

# events для lifetime/expire/delay этих артефактов
echo "  dump na_events for auction artefacts (lifetime/expire/delay)" >&2
AUCTION_EVENT_IDS=$(docker exec "$LEGACY_CONTAINER" mysql -u"$LEGACY_USER" -p"$LEGACY_PASS" "$LEGACY_DB" \
  -N -B -e "
    SELECT GROUP_CONCAT(DISTINCT eid) FROM (
      SELECT lifetime_eventid AS eid FROM na_artefact2user WHERE lot_id > 0 AND lifetime_eventid > 0
      UNION SELECT expire_eventid FROM na_artefact2user WHERE lot_id > 0 AND expire_eventid > 0
      UNION SELECT delay_eventid FROM na_artefact2user WHERE lot_id > 0 AND delay_eventid > 0
    ) t
  " 2>/dev/null)
if [ -n "$AUCTION_EVENT_IDS" ] && [ "$AUCTION_EVENT_IDS" != "NULL" ]; then
  echo "-- na_events (auction artefact events)" >> "$OUT"
  docker exec "$LEGACY_CONTAINER" mysqldump -u"$LEGACY_USER" -p"$LEGACY_PASS" "$LEGACY_DB" \
    --no-create-info --skip-extended-insert --skip-comments --hex-blob --complete-insert \
    --where="eventid IN ($AUCTION_EVENT_IDS)" na_events 2>/dev/null >> "$OUT" || true
fi
echo "" >> "$OUT"

# Группа 3.5: Stock — все брокеры биржи + активные лоты + связанные планеты/юзеры/galaxy

# Все брокеры биржи (na_exchange)
echo "  dump na_exchange (all brokers, ~63 rows)" >&2
echo "-- na_exchange (all brokers)" >> "$OUT"
docker exec "$LEGACY_CONTAINER" mysqldump -u"$LEGACY_USER" -p"$LEGACY_PASS" "$LEGACY_DB" \
  --no-create-info --skip-extended-insert --skip-comments --hex-blob --complete-insert \
  na_exchange 2>/dev/null >> "$OUT" || true

echo "  dump na_exchange_lots WHERE status IN (1,5)" >&2
echo "-- na_exchange_lots (all active for Stock page)" >> "$OUT"
docker exec "$LEGACY_CONTAINER" mysqldump -u"$LEGACY_USER" -p"$LEGACY_PASS" "$LEGACY_DB" \
  --no-create-info --skip-extended-insert --skip-comments --hex-blob --complete-insert \
  --where="status IN (1,5)" na_exchange_lots 2>/dev/null >> "$OUT" || true
echo "" >> "$OUT"

# Юзеры/планеты/galaxy владельцев активных лотов
LOT_PLANETIDS=$(docker exec "$LEGACY_CONTAINER" mysql -u"$LEGACY_USER" -p"$LEGACY_PASS" "$LEGACY_DB" \
  -N -B -e "SELECT GROUP_CONCAT(DISTINCT planetid) FROM na_exchange_lots WHERE status IN (1,5)" 2>/dev/null)
if [ -n "$LOT_PLANETIDS" ]; then
  echo "  dump na_planet+na_galaxy for lot-owners" >&2
  echo "-- na_planet (Stock lot owners)" >> "$OUT"
  docker exec "$LEGACY_CONTAINER" mysqldump -u"$LEGACY_USER" -p"$LEGACY_PASS" "$LEGACY_DB" \
    --no-create-info --skip-extended-insert --skip-comments --hex-blob --complete-insert \
    --where="planetid IN ($LOT_PLANETIDS) AND userid != $USERID" na_planet 2>/dev/null >> "$OUT" || true

  echo "-- na_galaxy (Stock lot owner planets)" >> "$OUT"
  docker exec "$LEGACY_CONTAINER" mysqldump -u"$LEGACY_USER" -p"$LEGACY_PASS" "$LEGACY_DB" \
    --no-create-info --skip-extended-insert --skip-comments --hex-blob --complete-insert \
    --where="planetid IN ($LOT_PLANETIDS)" na_galaxy 2>/dev/null >> "$OUT" || true

  # И их юзеров (если ещё не вытащены)
  LOT_USERIDS=$(docker exec "$LEGACY_CONTAINER" mysql -u"$LEGACY_USER" -p"$LEGACY_PASS" "$LEGACY_DB" \
    -N -B -e "SELECT GROUP_CONCAT(DISTINCT userid) FROM na_planet WHERE planetid IN ($LOT_PLANETIDS) AND userid IS NOT NULL" 2>/dev/null)
  if [ -n "$LOT_USERIDS" ]; then
    echo "-- na_user (Stock lot owners)" >> "$OUT"
    docker exec "$LEGACY_CONTAINER" mysqldump -u"$LEGACY_USER" -p"$LEGACY_PASS" "$LEGACY_DB" \
      --no-create-info --skip-extended-insert --skip-comments --hex-blob --complete-insert \
      --where="userid IN ($LOT_USERIDS) AND userid != $USERID" na_user 2>/dev/null >> "$OUT" || true
  fi
fi
echo "" >> "$OUT"

# Группа 3.6: Alliance — расширенные таблицы. У каждой свой WHERE.
if [ -n "$ALLY_AID" ]; then
  echo "  dump alliance extended tables (allyrank, applications, relationships)" >&2

  echo "-- na_allyrank WHERE aid=$ALLY_AID" >> "$OUT"
  docker exec "$LEGACY_CONTAINER" mysqldump -u"$LEGACY_USER" -p"$LEGACY_PASS" "$LEGACY_DB" \
    --no-create-info --skip-extended-insert --skip-comments --hex-blob --complete-insert \
    --where="aid=$ALLY_AID" na_allyrank 2>/dev/null >> "$OUT" || true

  echo "-- na_allyapplication WHERE aid=$ALLY_AID" >> "$OUT"
  docker exec "$LEGACY_CONTAINER" mysqldump -u"$LEGACY_USER" -p"$LEGACY_PASS" "$LEGACY_DB" \
    --no-create-info --skip-extended-insert --skip-comments --hex-blob --complete-insert \
    --where="aid=$ALLY_AID" na_allyapplication 2>/dev/null >> "$OUT" || true

  echo "-- na_ally_relationships WHERE rel1/rel2=$ALLY_AID" >> "$OUT"
  docker exec "$LEGACY_CONTAINER" mysqldump -u"$LEGACY_USER" -p"$LEGACY_PASS" "$LEGACY_DB" \
    --no-create-info --skip-extended-insert --skip-comments --hex-blob --complete-insert \
    --where="rel1=$ALLY_AID OR rel2=$ALLY_AID" na_ally_relationships 2>/dev/null >> "$OUT" || true

  echo "-- na_ally_relationships_application WHERE candidate_ally/request_ally=$ALLY_AID" >> "$OUT"
  docker exec "$LEGACY_CONTAINER" mysqldump -u"$LEGACY_USER" -p"$LEGACY_PASS" "$LEGACY_DB" \
    --no-create-info --skip-extended-insert --skip-comments --hex-blob --complete-insert \
    --where="candidate_ally=$ALLY_AID OR request_ally=$ALLY_AID" na_ally_relationships_application 2>/dev/null >> "$OUT" || true

  echo "" >> "$OUT"
fi

# Группа 4: чат — последние N сообщений + их авторы (для рендера Chat страницы)
echo "  dump last 100 chat messages + their authors" >&2
LAST_CHAT_AUTHORS=$(docker exec "$LEGACY_CONTAINER" mysql -u"$LEGACY_USER" -p"$LEGACY_PASS" "$LEGACY_DB" \
  -N -B -e "SELECT GROUP_CONCAT(DISTINCT userid) FROM (SELECT userid FROM na_chat ORDER BY messageid DESC LIMIT 100) t" 2>/dev/null)

if [ -n "$LAST_CHAT_AUTHORS" ]; then
  echo "-- na_chat (last 100 messages)" >> "$OUT"
  docker exec "$LEGACY_CONTAINER" mysqldump -u"$LEGACY_USER" -p"$LEGACY_PASS" "$LEGACY_DB" \
    --no-create-info --skip-extended-insert --skip-comments --hex-blob --complete-insert \
    --where="messageid IN (SELECT messageid FROM (SELECT messageid FROM na_chat ORDER BY messageid DESC LIMIT 100) t)" \
    na_chat 2>/dev/null >> "$OUT" || true

  echo "-- na_user (chat authors)" >> "$OUT"
  docker exec "$LEGACY_CONTAINER" mysqldump -u"$LEGACY_USER" -p"$LEGACY_PASS" "$LEGACY_DB" \
    --no-create-info --skip-extended-insert --skip-comments --hex-blob --complete-insert \
    --where="userid IN ($LAST_CHAT_AUTHORS) AND userid != $USERID" na_user 2>/dev/null >> "$OUT" || true
fi

# Группа 5: рейтинг — топ-50 юзеров по points для рендера Ranking
echo "  dump top-50 users for ranking" >&2
TOP_USERS=$(docker exec "$LEGACY_CONTAINER" mysql -u"$LEGACY_USER" -p"$LEGACY_PASS" "$LEGACY_DB" \
  -N -B -e "SELECT GROUP_CONCAT(userid) FROM (SELECT userid FROM na_user ORDER BY points DESC LIMIT 50) t" 2>/dev/null)
if [ -n "$TOP_USERS" ]; then
  echo "-- na_user (top-50 by points, кроме уже импортированных)" >> "$OUT"
  docker exec "$LEGACY_CONTAINER" mysqldump -u"$LEGACY_USER" -p"$LEGACY_PASS" "$LEGACY_DB" \
    --no-create-info --skip-extended-insert --skip-comments --hex-blob --complete-insert \
    --where="userid IN ($TOP_USERS) AND userid != $USERID" na_user 2>/dev/null >> "$OUT" || true
fi

cat >> "$OUT" <<EOF

COMMIT;
SET FOREIGN_KEY_CHECKS=1;
SET unique_checks=1;
EOF

SIZE=$(wc -c < "$OUT")
LINES=$(wc -l < "$OUT")
echo "" >&2
echo "Done: $OUT (${SIZE} bytes, ${LINES} lines)" >&2
